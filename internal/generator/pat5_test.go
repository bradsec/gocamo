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
		{R: 70, G: 72, B: 47, A: 255},    // Dark green
		{R: 109, G: 104, B: 81, A: 255},  // Medium green
		{R: 155, G: 150, B: 127, A: 255}, // Light green
		{R: 30, G: 36, B: 21, A: 255},    // Very dark green
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

func TestPat5Generator_DefaultColorRatios(t *testing.T) {
	gen := &Pat5Generator{}

	tests := []struct {
		name      string
		numColors int
		expected  []float64
	}{
		{
			name:      "4-color default",
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
			ratios := gen.getPat5DefaultColorRatios(tt.numColors)

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

func TestPat5Generator_InitializePat5Grid(t *testing.T) {
	gen := &Pat5Generator{}
	width, height := 10, 10
	numColors := 4

	grid := gen.initializePat5Grid(width, height, nil, numColors)

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

func TestPat5Generator_RectangularCA(t *testing.T) {
	gen := &Pat5Generator{}
	width, height := 20, 20
	numColors := 4

	grid := gen.initializePat5Grid(width, height, nil, numColors)
	gen.applyRectangularCA(grid, width, height, 5, 3, 2)

	for y, row := range grid {
		for x, colorIdx := range row {
			if colorIdx < 0 || colorIdx >= numColors {
				t.Errorf("Invalid color index after CA at [%d][%d]: %d", y, x, colorIdx)
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
		ColorRatios:   []float64{0.6, 0.25, 0.10, 0.05}, // Custom pat5-style ratios
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

func TestPat5Generator_DefaultRatios(t *testing.T) {
	// When no explicit ratios are set (RatiosString == ""), pat5 must use its
	// default ratios internally, so colour 0 (the base) dominates.
	ctx := context.Background()
	cfg := &config.Config{
		Width:         200,
		Height:        200,
		BasePixelSize: 4,
		PatternType:   "pat5",
		RatiosString:  "",
		ColorRatios:   []float64{0.25, 0.25, 0.25, 0.25}, // equal — set by SetColorRatios default
	}

	colors := []color.RGBA{
		{R: 90, G: 107, B: 60, A: 255},   // index 0 — base green (should dominate ~45%)
		{R: 212, G: 197, B: 167, A: 255}, // index 1 — tan
		{R: 74, G: 63, B: 42, A: 255},    // index 2 — brown
		{R: 45, G: 54, B: 42, A: 255},    // index 3 — dark green (accent ~10%)
	}

	gen := &Pat5Generator{}
	img, err := gen.Generate(ctx, cfg, colors)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if img == nil {
		t.Fatal("Generate returned nil image")
	}

	// Count colour pixels in the output
	bounds := img.Bounds()
	counts := make([]int, len(colors))
	total := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := img.At(x, y)
			r, g, b, _ := px.RGBA()
			rr, gg, bb := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			for i, c := range colors {
				if c.R == rr && c.G == gg && c.B == bb {
					counts[i]++
					total++
					break
				}
			}
		}
	}

	if total == 0 {
		t.Skip("No exact colour matches found (blending may be active)")
	}

	targets := []float64{0.45, 0.30, 0.15, 0.10}
	names := []string{"base green", "tan", "brown", "dark green"}
	tolerance := 0.10 // ±10% absolute

	for i := 0; i < len(colors); i++ {
		actual := float64(counts[i]) / float64(total)
		if actual < targets[i]-tolerance || actual > targets[i]+tolerance {
			t.Errorf("Colour %d (%s): actual %.3f outside target %.2f±%.2f",
				i, names[i], actual, targets[i], tolerance)
		}
	}
}
