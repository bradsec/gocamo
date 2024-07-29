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

type Generator interface {
	Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error)
}

func GeneratePattern(ctx context.Context, cfg *config.Config, camo config.CamoColors, index int, outputPath string) error {
	colors, err := utils.HexToRGBA(camo.Colors)
	if err != nil {
		return fmt.Errorf("error converting hex to RGBA: %w", err)
	}

	var gen Generator
	switch cfg.PatternType {
	case "blob":
		gen = &BlobGenerator{}
	case "box":
		gen = &BoxGenerator{}
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

	fileName := fmt.Sprintf("gocamo_%03d_%s_%s_%s_w%dx%d.png",
		index, camo.Name, colorCodesStr, cfg.PatternType, cfg.Width, cfg.Height)
	filePath := filepath.Join(outputPath, fileName)

	return saveImageToFile(img, filePath)
}

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
