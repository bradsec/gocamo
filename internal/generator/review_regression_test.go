package generator

import (
	"context"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradsec/gocamo/pkg/config"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "woodland", "woodland"},
		{"spaces and symbols", "jungle green (v2)", "jungle_green_v2"},
		{"path traversal", "../../etc/evil", "etc_evil"},
		{"separators", "a/b\\c", "a_b_c"},
		{"empty", "", "palette"},
		{"only unsafe", "../..", "palette"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeName(tt.in); got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGeneratePattern_PathTraversalName(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Width:         20,
		Height:        20,
		BasePixelSize: 4,
		PatternType:   "blocks",
	}
	camo := config.CamoColors{
		Name:   "../escape",
		Colors: []string{"ff0000", "00ff00"},
	}

	if err := GeneratePattern(context.Background(), cfg, camo, 0, outputDir); err != nil {
		t.Fatalf("GeneratePattern() error: %v", err)
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file inside output dir, got %d", len(files))
	}
	if strings.Contains(files[0].Name(), "..") {
		t.Errorf("filename still contains path traversal sequence: %s", files[0].Name())
	}

	// Nothing may be written outside the output directory.
	outside, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(outside) != 1 || !outside[0].IsDir() {
		t.Errorf("unexpected entries outside output dir: %v", outside)
	}
}

func TestGenerators_CancelledContext(t *testing.T) {
	cfg := &config.Config{
		Width:         64,
		Height:        64,
		BasePixelSize: 4,
		PatternType:   "woodland",
	}
	if err := cfg.SetColorRatios(2); err != nil {
		t.Fatal(err)
	}
	colors := []color.RGBA{
		{R: 0x10, G: 0x20, B: 0x30, A: 255},
		{R: 0x40, G: 0x50, B: 0x60, A: 255},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	generators := map[string]Generator{
		"woodland": &WoodlandGenerator{},
		"multicam": &MulticamGenerator{},
		"blocks":   &BlocksGenerator{},
		"blob":     &BlobGenerator{},
		"marpat":   &MarpatGenerator{},
		"fleck":    &FleckGenerator{},
	}
	for name, gen := range generators {
		t.Run(name, func(t *testing.T) {
			img, err := gen.Generate(ctx, cfg, colors)
			if err == nil {
				t.Fatal("Generate() with cancelled context returned nil error")
			}
			if img != nil {
				t.Error("Generate() with cancelled context returned non-nil image")
			}
		})
	}
}

func TestKMeansClustering_Guards(t *testing.T) {
	ctx := context.Background()
	pixels := []color.Color{
		color.RGBA{R: 1, G: 2, B: 3, A: 255},
		color.RGBA{R: 200, G: 100, B: 50, A: 255},
	}

	if _, err := kMeansClustering(ctx, nil, 3, 10); err == nil {
		t.Error("kMeansClustering() with no pixels returned nil error")
	}
	if _, err := kMeansClustering(ctx, pixels, 0, 10); err == nil {
		t.Error("kMeansClustering() with k=0 returned nil error")
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := kMeansClustering(cancelled, pixels, 2, 10); err == nil {
		t.Error("kMeansClustering() with cancelled context returned nil error")
	}

	got, err := kMeansClustering(ctx, pixels, 2, 10)
	if err != nil {
		t.Fatalf("kMeansClustering() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("kMeansClustering() returned %d colors, want 2", len(got))
	}
}

// TestGeneratePattern_AllTypes covers happy-path generation for every
// canonical pattern type.
func TestGeneratePattern_AllTypes(t *testing.T) {
	for _, patType := range []string{"woodland", "multicam", "blocks", "blob", "marpat", "fleck"} {
		t.Run(patType, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := &config.Config{
				Width:         40,
				Height:        40,
				BasePixelSize: 4,
				PatternType:   patType,
			}
			camo := config.CamoColors{
				Name:   "test",
				Colors: []string{"46482f", "6d6851", "9b967f"},
			}

			if err := GeneratePattern(context.Background(), cfg, camo, 0, tempDir); err != nil {
				t.Fatalf("GeneratePattern(%s) error: %v", patType, err)
			}

			files, err := os.ReadDir(tempDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(files) != 1 {
				t.Fatalf("expected 1 output file, got %d", len(files))
			}
			if !strings.Contains(files[0].Name(), patType) {
				t.Errorf("filename missing pattern type %s: %s", patType, files[0].Name())
			}
		})
	}
}

// TestGeneratePattern_LegacyAliasesRemoved verifies the retired pat1-pat5
// aliases are rejected as unknown pattern types.
func TestGeneratePattern_LegacyAliasesRemoved(t *testing.T) {
	for _, alias := range []string{"pat1", "pat2", "pat3", "pat4", "pat5"} {
		t.Run(alias, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := &config.Config{
				Width:         40,
				Height:        40,
				BasePixelSize: 4,
				PatternType:   alias,
			}
			camo := config.CamoColors{
				Name:   "alias",
				Colors: []string{"46482f", "6d6851"},
			}

			err := GeneratePattern(context.Background(), cfg, camo, 0, tempDir)
			if err == nil {
				t.Fatalf("GeneratePattern(%s) should fail, aliases were removed", alias)
			}
			if !strings.Contains(err.Error(), "unknown pattern type") {
				t.Errorf("GeneratePattern(%s) error = %v, want unknown pattern type", alias, err)
			}
		})
	}
}

// TestFleckGenerator_UsesAllColors verifies the fleck pattern actually places
// flecks: the output must contain more than just the background color.
func TestFleckGenerator_UsesAllColors(t *testing.T) {
	cfg := &config.Config{
		Width:         200,
		Height:        200,
		BasePixelSize: 4,
		PatternType:   "fleck",
	}
	if err := cfg.SetColorRatios(3); err != nil {
		t.Fatal(err)
	}
	colors := []color.RGBA{
		{R: 0x46, G: 0x48, B: 0x2f, A: 255},
		{R: 0x6d, G: 0x68, B: 0x51, A: 255},
		{R: 0x1e, G: 0x24, B: 0x15, A: 255},
	}

	gen := &FleckGenerator{}
	img, err := gen.Generate(context.Background(), cfg, colors)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	seen := make(map[color.RGBA]bool)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			seen[color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}] = true
		}
	}
	if len(seen) < 2 {
		t.Errorf("fleck pattern used %d distinct colors, want at least 2", len(seen))
	}
}
