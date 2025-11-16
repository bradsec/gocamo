package generator

import (
	"context"
	"image/color"
	"testing"

	"github.com/bradsec/gocamo/pkg/config"
)

func TestPat5Generator_Generate(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		PatternType:   "pat5",
	}

	colors := []color.RGBA{
		{R: 70, G: 72, B: 47, A: 255},   // Dark green
		{R: 109, G: 104, B: 81, A: 255}, // Medium green
		{R: 155, G: 150, B: 127, A: 255}, // Light green
		{R: 30, G: 36, B: 21, A: 255},   // Very dark green
	}

	gen := &Pat5Generator{}
	img, err := gen.Generate(ctx, cfg, colors)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if img == nil {
		t.Fatal("Generate returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != cfg.Width || bounds.Dy() != cfg.Height {
		t.Errorf("Image dimensions incorrect: got %dx%d, want %dx%d",
			bounds.Dx(), bounds.Dy(), cfg.Width, cfg.Height)
	}
}

func TestPat5Generator_MARPATColorRatios(t *testing.T) {
	gen := &Pat5Generator{}

	tests := []struct {
		name      string
		numColors int
		expected  []float64
	}{
		{
			name:      "4-color MARPAT",
			numColors: 4,
			expected:  []float64{0.45, 0.30, 0.15, 0.10},
		},
		{
			name:      "3-color adapted",
			numColors: 3,
			expected:  []float64{0.50, 0.35, 0.15},
		},
		{
			name:      "2-color fallback",
			numColors: 2,
			expected:  []float64{0.5, 0.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratios := gen.getMARPATColorRatios(tt.numColors)

			if len(ratios) != tt.numColors {
				t.Errorf("Expected %d ratios, got %d", tt.numColors, len(ratios))
			}

			for i, expected := range tt.expected {
				if i < len(ratios) {
					if ratios[i] != expected {
						t.Errorf("Ratio %d: expected %f, got %f", i, expected, ratios[i])
					}
				}
			}

			// Test that ratios sum to 1.0
			sum := 0.0
			for _, ratio := range ratios {
				sum += ratio
			}
			if sum < 0.99 || sum > 1.01 { // Allow small floating point errors
				t.Errorf("Ratios don't sum to 1.0: got %f", sum)
			}
		})
	}
}

func TestPat5Generator_CreateRectangularTemplates(t *testing.T) {
	gen := &Pat5Generator{}
	templates := gen.createRectangularTemplates()

	if len(templates) == 0 {
		t.Fatal("No templates created")
	}

	// Test that templates have valid dimensions
	for i, template := range templates {
		if template.Width <= 0 || template.Height <= 0 {
			t.Errorf("Template %d has invalid dimensions: %dx%d", i, template.Width, template.Height)
		}

		if len(template.Matrix) != template.Height {
			t.Errorf("Template %d matrix height mismatch: expected %d, got %d",
				i, template.Height, len(template.Matrix))
		}

		for j, row := range template.Matrix {
			if len(row) != template.Width {
				t.Errorf("Template %d row %d width mismatch: expected %d, got %d",
					i, j, template.Width, len(row))
			}
		}
	}
}

func TestPat5Generator_InitializeMARPATGrid(t *testing.T) {
	gen := &Pat5Generator{}
	width, height := 10, 10
	numColors := 4

	grid := gen.initializeMARPATGrid(width, height, nil, numColors)

	if len(grid) != height {
		t.Errorf("Grid height incorrect: expected %d, got %d", height, len(grid))
	}

	for i, row := range grid {
		if len(row) != width {
			t.Errorf("Grid row %d width incorrect: expected %d, got %d", i, width, len(row))
		}

		for j, colorIdx := range row {
			if colorIdx < 0 || colorIdx >= numColors {
				t.Errorf("Invalid color index at [%d][%d]: %d (should be 0-%d)", i, j, colorIdx, numColors-1)
			}
		}
	}
}

func TestPat5Generator_GenerateFractalLayer(t *testing.T) {
	gen := &Pat5Generator{}
	width, height := 20, 20
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
	}

	layer := gen.generateFractalLayer(width, height, colors)

	if len(layer) != height {
		t.Errorf("Fractal layer height incorrect: expected %d, got %d", height, len(layer))
	}

	for i, row := range layer {
		if len(row) != width {
			t.Errorf("Fractal layer row %d width incorrect: expected %d, got %d", i, width, len(row))
		}

		for j, colorIdx := range row {
			if colorIdx < 0 || colorIdx >= len(colors) {
				t.Errorf("Invalid color index in fractal layer at [%d][%d]: %d (should be 0-%d)",
					i, j, colorIdx, len(colors)-1)
			}
		}
	}
}

func TestPat5Generator_GetDominantColor(t *testing.T) {
	gen := &Pat5Generator{}

	// Create a test grid with known dominant color
	grid := [][]int{
		{0, 0, 1, 1},
		{0, 0, 1, 2},
		{0, 2, 2, 2},
		{3, 3, 2, 2},
	}

	// Test area where color 0 is dominant (top-left 2x2)
	dominant := gen.getDominantColor(grid, 0, 0, 2, 2)
	if dominant != 0 {
		t.Errorf("Expected dominant color 0 in top-left, got %d", dominant)
	}

	// Test area where color 2 is dominant (bottom-right 3x3)
	dominant = gen.getDominantColor(grid, 1, 1, 3, 3)
	if dominant != 2 {
		t.Errorf("Expected dominant color 2 in bottom-right, got %d", dominant)
	}
}

func TestPat5Generator_CreateTemplate(t *testing.T) {
	gen := &Pat5Generator{}

	pattern := [][]bool{
		{true, false, true},
		{false, true, false},
		{true, false, true},
	}

	template := gen.createTemplate(3, 3, pattern)

	if template.Width != 3 || template.Height != 3 {
		t.Errorf("Template dimensions incorrect: got %dx%d, want 3x3", template.Width, template.Height)
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if template.Matrix[i][j] != pattern[i][j] {
				t.Errorf("Template matrix mismatch at [%d][%d]: got %t, want %t",
					i, j, template.Matrix[i][j], pattern[i][j])
			}
		}
	}
}

func TestPat5Generator_WithCustomColorRatios(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 2,
		PatternType:   "pat5",
		ColorRatios:   []float64{0.6, 0.25, 0.10, 0.05}, // Custom MARPAT-style ratios
	}

	colors := []color.RGBA{
		{R: 70, G: 72, B: 47, A: 255},
		{R: 109, G: 104, B: 81, A: 255},
		{R: 155, G: 150, B: 127, A: 255},
		{R: 30, G: 36, B: 21, A: 255},
	}

	gen := &Pat5Generator{}
	img, err := gen.Generate(ctx, cfg, colors)

	if err != nil {
		t.Fatalf("Generate with custom ratios failed: %v", err)
	}

	if img == nil {
		t.Fatal("Generate returned nil image with custom ratios")
	}
}

func TestPat5Generator_EdgeCases(t *testing.T) {
	gen := &Pat5Generator{}
	ctx := context.Background()

	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "Minimum dimensions",
			cfg: &config.Config{
				Width:         4,
				Height:        4,
				BasePixelSize: 1,
				PatternType:   "pat5",
			},
		},
		{
			name: "Large base pixel size",
			cfg: &config.Config{
				Width:         100,
				Height:        100,
				BasePixelSize: 50,
				PatternType:   "pat5",
			},
		},
	}

	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.Generate(ctx, tt.cfg, colors)
			if err != nil {
				t.Errorf("Generate failed for %s: %v", tt.name, err)
			}
			if img == nil {
				t.Errorf("Generate returned nil image for %s", tt.name)
			}
		})
	}
}