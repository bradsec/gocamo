package config

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/bradsec/gocamo/internal/utils"
)

// Config holds all configuration parameters for camouflage pattern generation.
// It includes image dimensions, pattern settings, file paths, and generation options.
type Config struct {
	Width         int
	Height        int
	BasePixelSize int
	JSONFile      string
	OutputDir     string
	ColorsString  string
	Cores         int
	AddEdge       bool
	AddNoise      bool
	PatternType   string
	ImageDir      string
	KValue        int
	RatiosString  string
	ColorRatios   []float64
}

// CamoColors represents a named color palette for camouflage pattern generation.
// It contains a descriptive name and a list of hex color codes to be used in the pattern.
type CamoColors struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

// SetColorRatios sets the color ratios for this configuration based on the number of colors.
// This should be called after colors are determined but before pattern generation.
func (cfg *Config) SetColorRatios(numColors int) error {
	ratios, err := parseColorRatios(cfg.RatiosString, numColors)
	if err != nil {
		return fmt.Errorf("error parsing color ratios: %w", err)
	}
	cfg.ColorRatios = ratios
	return nil
}

// AdjustBasePixelSize calculates the optimal base pixel size that divides evenly into both width and height.
// It finds the divisor closest to the requested BasePixelSize to maintain the user's intended scale while
// ensuring perfect fit without transparent borders.
//
// The function prioritizes larger divisors (closer to the original) when there are ties, and ensures
// the result is always at least 1.
func (cfg *Config) AdjustBasePixelSize() int {
	if cfg.BasePixelSize <= 0 {
		cfg.BasePixelSize = 4 // Set default if invalid
	}
	
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return cfg.BasePixelSize // Return original if dimensions are invalid
	}
	
	// If the current base pixel size already fits perfectly, use it
	if cfg.Width%cfg.BasePixelSize == 0 && cfg.Height%cfg.BasePixelSize == 0 {
		return cfg.BasePixelSize
	}
	
	// Find all common divisors of width and height
	commonDivisors := findCommonDivisors(cfg.Width, cfg.Height)
	
	// Find the divisor closest to the requested BasePixelSize
	bestDivisor := 1
	minDifference := abs(cfg.BasePixelSize - 1)
	
	for _, divisor := range commonDivisors {
		difference := abs(cfg.BasePixelSize - divisor)
		if difference < minDifference || (difference == minDifference && divisor > bestDivisor) {
			bestDivisor = divisor
			minDifference = difference
		}
	}
	
	return bestDivisor
}

// findCommonDivisors returns all common divisors of two positive integers, sorted in ascending order.
func findCommonDivisors(a, b int) []int {
	if a <= 0 || b <= 0 {
		return []int{1}
	}
	
	// Find GCD using Euclidean algorithm
	gcd := func(x, y int) int {
		for y != 0 {
			x, y = y, x%y
		}
		return x
	}
	
	commonGCD := gcd(a, b)
	
	// Find all divisors of the GCD
	var divisors []int
	for i := 1; i*i <= commonGCD; i++ {
		if commonGCD%i == 0 {
			divisors = append(divisors, i)
			if i != commonGCD/i {
				divisors = append(divisors, commonGCD/i)
			}
		}
	}
	
	// Sort divisors in ascending order
	for i := 0; i < len(divisors)-1; i++ {
		for j := i + 1; j < len(divisors); j++ {
			if divisors[i] > divisors[j] {
				divisors[i], divisors[j] = divisors[j], divisors[i]
			}
		}
	}
	
	return divisors
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func validateHexColor(hex string) error {
	original := hex
	hex = utils.StripHash(strings.TrimSpace(hex))

	// Check if original had spaces (after stripping hash)
	stripped := utils.StripHash(original)
	if stripped != strings.TrimSpace(stripped) {
		return fmt.Errorf("invalid hex color with spaces: %s", original)
	}

	if len(hex) != 3 && len(hex) != 6 {
		return fmt.Errorf("invalid hex color length: %s (should be 6 characters or 3 for short form)", hex)
	}
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("invalid hex color character in %s", hex)
		}
	}
	return nil
}

func cleanColorString(colors string) (string, error) {
	// Remove all whitespace and split
	parts := strings.Split(strings.ReplaceAll(colors, " ", ""), ",")
	// Filter out empty strings and validate format
	var cleaned []string
	for _, p := range parts {
		if p != "" {
			// Validate hex color format
			if err := validateHexColor(p); err != nil {
				return "", fmt.Errorf("invalid color %s: %v", p, err)
			}
			cleaned = append(cleaned, p)
		}
	}

	// Check minimum number of colors
	if len(cleaned) < 2 {
		return "", fmt.Errorf("at least 2 colors are required, got %d", len(cleaned))
	}

	return strings.Join(cleaned, ","), nil
}

func parseColorRatios(ratiosString string, numColors int) ([]float64, error) {
	if ratiosString == "" {
		// Default: Equal ratios for all colors
		ratios := make([]float64, numColors)
		for i := 0; i < numColors; i++ {
			ratios[i] = 1.0 / float64(numColors)
		}
		return ratios, nil
	}

	if ratiosString == "equal" {
		// Equal ratios for all colors (each gets value 1)
		ratios := make([]float64, numColors)
		for i := 0; i < numColors; i++ {
			ratios[i] = 1.0 / float64(numColors)
		}
		return ratios, nil
	}

	if ratiosString == "random" {
		// Generate random ratios (simple integers between 1-5)
		ratios := make([]float64, numColors)
		sum := 0.0
		for i := 0; i < numColors; i++ {
			// Random integer between 1 and 5 for simplicity
			randomInt := 1 + rand.Intn(5)
			ratios[i] = float64(randomInt)
			sum += ratios[i]
		}
		// Normalize to sum to 1.0
		for i := 0; i < numColors; i++ {
			ratios[i] /= sum
		}
		return ratios, nil
	}

	// Parse custom ratios (expect simple integers like "2,1,3")
	parts := strings.Split(strings.ReplaceAll(ratiosString, " ", ""), ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("no ratio values provided")
	}

	// Parse the provided ratios first
	parsedRatios := make([]float64, len(parts))
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty ratio value at position %d", i+1)
		}

		// Parse as integer for simplicity
		intRatio, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid ratio value '%s' at position %d: must be a positive integer", part, i+1)
		}

		if intRatio < 0 {
			return nil, fmt.Errorf("ratio value cannot be negative: %d at position %d", intRatio, i+1)
		}

		if intRatio == 0 {
			return nil, fmt.Errorf("ratio value cannot be zero: use a positive integer at position %d", i+1)
		}

		parsedRatios[i] = float64(intRatio)
	}

	// Create ratios array for all colors, cycling through provided ratios
	ratios := make([]float64, numColors)
	sum := 0.0

	for i := 0; i < numColors; i++ {
		// Cycle through the provided ratios
		ratioIndex := i % len(parsedRatios)
		ratios[i] = parsedRatios[ratioIndex]
		sum += ratios[i]
	}

	// Normalize ratios to sum to 1.0
	for i := 0; i < numColors; i++ {
		ratios[i] /= sum
	}

	return ratios, nil
}

// ParseFlags parses command-line arguments and returns a validated configuration.
// It defines and processes all command-line flags for the gocamo application,
// providing default values and validation for user inputs. The function handles
// configuration for image dimensions, pattern generation settings, file paths,
// and processing options.
//
// The function performs validation on several parameters:
//   - CPU core count (clamped to available cores)
//   - Image dimensions (minimum value enforcement)
//   - Base pixel size validation
//   - Color string format validation and cleanup
//   - Automatic pattern type detection based on input flags
//
// Special behaviors:
//   - Using the -i flag automatically sets pattern type to "image"
//   - Color strings are cleaned and validated for proper hex format
//   - Invalid configurations cause the program to exit with an error message
//
// Returns a pointer to a Config struct containing all parsed and validated settings.
// The program will exit with status 1 if critical validation errors occur.
func ParseFlags() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Width, "w", 1500, "Set the image width")
	flag.IntVar(&cfg.Height, "h", 1500, "Set the image height")
	flag.IntVar(&cfg.BasePixelSize, "b", 4, "Set the base pixel size (will be adjusted if necessary)")
	flag.StringVar(&cfg.JSONFile, "j", "", "Process a JSON file containing a list of color palettes")
	flag.StringVar(&cfg.OutputDir, "o", "output", "The output directory for generated images")
	flag.StringVar(&cfg.ColorsString, "c", "", "Generate a single pattern using a comma-separated list of hex colors")
	flag.IntVar(&cfg.Cores, "cores", runtime.NumCPU(), fmt.Sprintf("Number of CPU cores to use (1-%d available)", runtime.NumCPU()))
	flag.BoolVar(&cfg.AddEdge, "edge", false, "Add edge details to the pattern")
	flag.BoolVar(&cfg.AddNoise, "noise", false, "Add noise to the pattern")
	flag.StringVar(&cfg.PatternType, "t", "pat1", "Set the pattern type (pat1, pat2, pat3, pat4, pat5, all, or image)")
	flag.StringVar(&cfg.ImageDir, "i", "input", "Input directory containing images for image-based camouflage")
	flag.IntVar(&cfg.KValue, "k", 4, "Number of main colors for image-based camouflage")
	flag.StringVar(&cfg.RatiosString, "r", "", "Color ratios: 'random' for random ratios, integers like '2,1,3' (cycles if fewer than colors) (default: equal)")

	flag.Parse()

	// Validate cores
	if cfg.Cores < 1 {
		cfg.Cores = 1
	} else if cfg.Cores > runtime.NumCPU() {
		cfg.Cores = runtime.NumCPU()
	}

	// Validate dimensions
	if cfg.Width < 1 {
		cfg.Width = 1500 // default
	}
	if cfg.Height < 1 {
		cfg.Height = 1500 // default
	}
	if cfg.BasePixelSize < 1 {
		cfg.BasePixelSize = 4 // default
	}

	// If -i flag is used, set pattern type to "image"
	if isFlagPassed("i") {
		cfg.PatternType = "image"
	}

	// Clean and validate the colors string if provided
	if cfg.ColorsString != "" {
		cleaned, err := cleanColorString(cfg.ColorsString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		cfg.ColorsString = cleaned
	}

	return cfg
}

// Helper function to check if a flag was explicitly passed
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
