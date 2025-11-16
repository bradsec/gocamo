package generator

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bradsec/gocamo/internal/utils"
	"github.com/bradsec/gocamo/pkg/config"
)

// Generator defines the interface for all camouflage pattern generators.
// Implementations create different types of camouflage patterns (blob, box, image-based).
type Generator interface {
	Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error)
}

// GeneratePattern creates a camouflage pattern image using the specified color palette and pattern type.
// It generates a pattern based on the configuration settings including pattern type (blob or box),
// dimensions, and visual effects. The generated image is saved to the output directory with a
// descriptive filename that includes the pattern index, color palette name, color codes, pattern type,
// and dimensions.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - cfg: Configuration containing pattern type, dimensions, and visual effects settings
//   - camo: Color palette containing the name and hex color codes to use
//   - index: Sequential number for the generated pattern (used in filename)
//   - outputPath: Directory path where the generated image will be saved
//
// Returns an error if the color palette is empty, color conversion fails, pattern type is unknown,
// pattern generation fails, or file saving fails.
func GeneratePattern(ctx context.Context, cfg *config.Config, camo config.CamoColors, index int, outputPath string) error {
	if len(camo.Colors) == 0 {
		return fmt.Errorf("no colors provided in color palette")
	}

	colors, err := utils.HexToRGBA(camo.Colors)
	if err != nil {
		return fmt.Errorf("error converting hex to RGBA: %w", err)
	}

	// Set color ratios for this pattern generation
	if err := cfg.SetColorRatios(len(colors)); err != nil {
		return fmt.Errorf("error setting color ratios: %w", err)
	}

	var gen Generator
	switch cfg.PatternType {
	case "pat5":
		gen = &Pat5Generator{}
	case "pat4":
		gen = &Pat4Generator{}
	case "pat3":
		gen = &Pat3Generator{}
	case "pat2":
		gen = &Pat2Generator{}
	case "pat1":
		gen = &Pat1Generator{}
	default:
		return fmt.Errorf("unknown pattern type: %s", cfg.PatternType)
	}

	img, err := gen.Generate(ctx, cfg, colors)
	if err != nil {
		return fmt.Errorf("error generating pattern: %w", err)
	}

	colorCodes := make([]string, len(camo.Colors))
	for i, hex := range camo.Colors {
		colorCodes[i] = strings.TrimPrefix(hex, "#")
	}
	colorCodesStr := strings.Join(colorCodes, "_")

	patternName := cfg.PatternType

	fileName := fmt.Sprintf("gocamo_%03d_%s_%s_%s_w%dx%d.png",
		index, camo.Name, colorCodesStr, patternName, cfg.Width, cfg.Height)
	filePath := filepath.Join(outputPath, fileName)

	return saveImageToFile(img, filePath)
}

// GenerateFromImage creates a camouflage pattern by analyzing an input image and extracting its dominant colors.
// It uses k-means clustering to identify the main colors from the input image, then generates a
// camouflage pattern using those colors. The process includes image preprocessing with max pooling
// and Laplacian filtering to enhance edge detection and color extraction.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - cfg: Configuration containing dimensions, k-value for clustering, and visual effects settings
//   - imagePath: Path to the input image file to analyze for color extraction
//   - index: Sequential number for the generated pattern (used in filename)
//   - outputPath: Directory path where the generated image will be saved
//
// The generated filename includes the source image name, index, extracted color codes,
// k-value used for clustering, and final dimensions.
//
// Returns an error if the image cannot be loaded, pattern generation fails, or file saving fails.
func GenerateFromImage(ctx context.Context, cfg *config.Config, imagePath string, index int, outputPath string) error {
	gen := &ImageGenerator{InputFile: imagePath}

	img, mainColors, err := gen.Generate(ctx, cfg, nil)

	// Sort the main colors
	sortColors(mainColors)

	// Convert main colors to hex for filename
	hexColors := make([]string, len(mainColors))
	for i, c := range mainColors {
		hexColors[i] = fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
	}
	colorCodesStr := strings.Join(hexColors, "_")

	if err != nil {
		return fmt.Errorf("error generating pattern from image %s: %w", imagePath, err)
	}

	baseName := filepath.Base(imagePath)
	fileName := fmt.Sprintf("gocamo_from_image_%s_%03d_%s_k%d_w%dx%d.png",
		strings.TrimSuffix(baseName, filepath.Ext(baseName)),
		index, colorCodesStr, cfg.KValue, cfg.Width, cfg.Height)
	filePath := filepath.Join(outputPath, fileName)

	if err := saveImageToFile(img, filePath); err != nil {
		return fmt.Errorf("error saving image %s: %w", filePath, err)
	}

	return nil
}

func saveImageToFile(img image.Image, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	if err := utils.SaveImage(img, f); err != nil {
		return fmt.Errorf("error saving image: %w", err)
	}

	return nil
}

func sortColors(colors []color.RGBA) {
	sort.Slice(colors, func(i, j int) bool {
		iSum := int(colors[i].R) + int(colors[i].G) + int(colors[i].B)
		jSum := int(colors[j].R) + int(colors[j].G) + int(colors[j].B)
		return iSum < jSum
	})
}

func shuffleColors(colors []color.RGBA) []color.RGBA {
	shuffled := make([]color.RGBA, len(colors))
	copy(shuffled, colors)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

// selectRandomColor picks a random color from the palette instead of using brightness-based selection
func selectRandomColor(colors []color.RGBA) int {
	return rand.Intn(len(colors))
}
