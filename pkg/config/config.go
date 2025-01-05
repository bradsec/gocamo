package config

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

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
}

type CamoColors struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

func stripHash(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}

func validateHexColor(hex string) error {
	hex = stripHash(strings.TrimSpace(hex))
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
	flag.StringVar(&cfg.PatternType, "t", "box", "Set the pattern type (blob, box, or image)")
	flag.StringVar(&cfg.ImageDir, "i", "input", "Input directory containing images for image-based camouflage")
	flag.IntVar(&cfg.KValue, "k", 4, "Number of main colors for image-based camouflage")

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
