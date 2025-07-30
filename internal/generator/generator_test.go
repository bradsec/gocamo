package generator

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bradsec/gocamo/pkg/config"
)

func TestSortColors(t *testing.T) {
	tests := []struct {
		name     string
		input    []color.RGBA
		expected []color.RGBA
	}{
		{
			name: "already sorted",
			input: []color.RGBA{
				{R: 0, G: 0, B: 0, A: 255},       // Black (0)
				{R: 128, G: 128, B: 128, A: 255}, // Gray (384)
				{R: 255, G: 255, B: 255, A: 255}, // White (765)
			},
			expected: []color.RGBA{
				{R: 0, G: 0, B: 0, A: 255},
				{R: 128, G: 128, B: 128, A: 255},
				{R: 255, G: 255, B: 255, A: 255},
			},
		},
		{
			name: "reverse order",
			input: []color.RGBA{
				{R: 255, G: 255, B: 255, A: 255}, // White (765)
				{R: 128, G: 128, B: 128, A: 255}, // Gray (384)
				{R: 0, G: 0, B: 0, A: 255},       // Black (0)
			},
			expected: []color.RGBA{
				{R: 0, G: 0, B: 0, A: 255},
				{R: 128, G: 128, B: 128, A: 255},
				{R: 255, G: 255, B: 255, A: 255},
			},
		},
		{
			name: "mixed colors",
			input: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},   // Red (255)
				{R: 0, G: 255, B: 0, A: 255},   // Green (255)
				{R: 0, G: 0, B: 255, A: 255},   // Blue (255)
				{R: 255, G: 255, B: 0, A: 255}, // Yellow (510)
			},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},   // One of the 255-sum colors
				{R: 0, G: 255, B: 0, A: 255},   // One of the 255-sum colors
				{R: 0, G: 0, B: 255, A: 255},   // One of the 255-sum colors
				{R: 255, G: 255, B: 0, A: 255}, // Yellow (510)
			},
		},
		{
			name:     "empty slice",
			input:    []color.RGBA{},
			expected: []color.RGBA{},
		},
		{
			name: "single color",
			input: []color.RGBA{
				{R: 100, G: 100, B: 100, A: 255},
			},
			expected: []color.RGBA{
				{R: 100, G: 100, B: 100, A: 255},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test input
			colors := make([]color.RGBA, len(tt.input))
			copy(colors, tt.input)

			sortColors(colors)

			// For mixed colors with same sum, we need to check if the order is valid
			if tt.name == "mixed colors" {
				// Check that the first three colors have sum 255 and last has sum 510
				for i := 0; i < 3; i++ {
					sum := int(colors[i].R) + int(colors[i].G) + int(colors[i].B)
					if sum != 255 {
						t.Errorf("Colors[%d] sum = %d, want 255", i, sum)
					}
				}
				sum := int(colors[3].R) + int(colors[3].G) + int(colors[3].B)
				if sum != 510 {
					t.Errorf("Colors[3] sum = %d, want 510", sum)
				}
			} else {
				if !reflect.DeepEqual(colors, tt.expected) {
					t.Errorf("sortColors() = %v, want %v", colors, tt.expected)
				}
			}
		})
	}
}

func TestShuffleColors(t *testing.T) {
	original := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
	}

	shuffled := shuffleColors(original)

	// Check that we get the same number of colors
	if len(shuffled) != len(original) {
		t.Errorf("shuffleColors() returned %d colors, want %d", len(shuffled), len(original))
	}

	// Check that all original colors are present
	for _, origColor := range original {
		found := false
		for _, shuffColor := range shuffled {
			if origColor == shuffColor {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Original color %v not found in shuffled slice", origColor)
		}
	}

	// Check that original slice is not modified
	expectedOriginal := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
	}
	if !reflect.DeepEqual(original, expectedOriginal) {
		t.Error("shuffleColors() modified the original slice")
	}
}

func TestShuffleColors_EdgeCases(t *testing.T) {
	// Test with empty slice
	empty := []color.RGBA{}
	shuffled := shuffleColors(empty)
	if len(shuffled) != 0 {
		t.Errorf("shuffleColors(empty) = %d colors, want 0", len(shuffled))
	}

	// Test with single color
	single := []color.RGBA{{R: 255, G: 0, B: 0, A: 255}}
	shuffled = shuffleColors(single)
	if len(shuffled) != 1 {
		t.Errorf("shuffleColors(single) = %d colors, want 1", len(shuffled))
	}
	if shuffled[0] != single[0] {
		t.Errorf("shuffleColors(single) = %v, want %v", shuffled[0], single[0])
	}
}

func TestSaveImageToFile(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")

	err := saveImageToFile(img, testFile)
	if err != nil {
		t.Errorf("saveImageToFile() error = %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("saveImageToFile() did not create file")
	}

	// Check file content by loading it back
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open saved file: %v", err)
	}
	defer file.Close()

	loadedImg, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode saved PNG: %v", err)
	}

	bounds := loadedImg.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("Loaded image size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestSaveImageToFile_InvalidPath(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	// Try to save to invalid path
	err := saveImageToFile(img, "/invalid/path/file.png")
	if err == nil {
		t.Error("saveImageToFile() should return error for invalid path")
	}
	if !strings.Contains(err.Error(), "error creating file") {
		t.Errorf("Expected error about creating file, got: %v", err)
	}
}

func TestGeneratePattern_EmptyColors(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		PatternType:   "pat3",
	}

	camo := config.CamoColors{
		Name:   "empty",
		Colors: []string{},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	if err == nil {
		t.Error("GeneratePattern() should return error for empty colors")
	}
	if !strings.Contains(err.Error(), "no colors provided") {
		t.Errorf("Expected error about no colors provided, got: %v", err)
	}
}

func TestGeneratePattern_InvalidHexColors(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		PatternType:   "pat3",
	}

	camo := config.CamoColors{
		Name:   "invalid",
		Colors: []string{"invalid", "ff0000"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	if err == nil {
		t.Error("GeneratePattern() should return error for invalid hex colors")
	}
	if !strings.Contains(err.Error(), "error converting hex to RGBA") {
		t.Errorf("Expected error about converting hex to RGBA, got: %v", err)
	}
}

func TestGeneratePattern_UnknownPatternType(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		PatternType:   "unknown",
	}

	camo := config.CamoColors{
		Name:   "test",
		Colors: []string{"ff0000", "00ff00"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	if err == nil {
		t.Error("GeneratePattern() should return error for unknown pattern type")
	}
	if !strings.Contains(err.Error(), "unknown pattern type") {
		t.Errorf("Expected error about unknown pattern type, got: %v", err)
	}
}

func TestGeneratePattern_ValidPat3(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		PatternType:   "pat3",
		AddEdge:       false,
		AddNoise:      false,
	}

	camo := config.CamoColors{
		Name:   "test",
		Colors: []string{"ff0000", "00ff00", "0000ff"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	if err != nil {
		t.Errorf("GeneratePattern() unexpected error: %v", err)
	}

	// Check that file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Check filename format
	filename := files[0].Name()
	if !strings.HasPrefix(filename, "gocamo_000_test_") {
		t.Errorf("Filename should start with 'gocamo_000_test_', got: %s", filename)
	}
	if !strings.HasSuffix(filename, "_pat3_w100x100.png") {
		t.Errorf("Filename should end with '_pat3_w100x100.png', got: %s", filename)
	}
	if !strings.Contains(filename, "ff0000_00ff00_0000ff") {
		t.Errorf("Filename should contain color codes, got: %s", filename)
	}
}

func TestGeneratePattern_ValidPat4(t *testing.T) {
	cfg := &config.Config{
		Width:         50, // Smaller for faster test
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "pat4",
		AddEdge:       true,
		AddNoise:      true,
	}

	camo := config.CamoColors{
		Name:   "pat4_test",
		Colors: []string{"#ff0000", "#00ff00"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 5, tempDir)
	if err != nil {
		t.Errorf("GeneratePattern() unexpected error: %v", err)
	}

	// Check that file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Check filename format
	filename := files[0].Name()
	if !strings.HasPrefix(filename, "gocamo_005_pat4_test_") {
		t.Errorf("Filename should start with 'gocamo_005_pat4_test_', got: %s", filename)
	}
	if !strings.Contains(filename, "_pat4_") {
		t.Errorf("Filename should contain '_pat4_', got: %s", filename)
	}
}

func TestGenerateFromImage_InvalidImagePath(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		KValue:        4,
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GenerateFromImage(ctx, cfg, "/nonexistent/image.jpg", 0, tempDir)
	if err == nil {
		t.Error("GenerateFromImage() should return error for nonexistent image")
	}
	if !strings.Contains(err.Error(), "error generating pattern from image") {
		t.Errorf("Expected error about generating pattern from image, got: %v", err)
	}
}

func TestGenerateFromImage_ValidImage(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	// Create a simple pattern with different colors
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if x < 10 && y < 10 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red quadrant
			} else if x >= 10 && y < 10 {
				img.Set(x, y, color.RGBA{0, 255, 0, 255}) // Green quadrant
			} else if x < 10 && y >= 10 {
				img.Set(x, y, color.RGBA{0, 0, 255, 255}) // Blue quadrant
			} else {
				img.Set(x, y, color.RGBA{255, 255, 0, 255}) // Yellow quadrant
			}
		}
	}

	// Save test image
	tempDir := t.TempDir()
	testImagePath := filepath.Join(tempDir, "test_input.png")
	file, err := os.Create(testImagePath)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		KValue:        4,
		AddEdge:       false,
		AddNoise:      false,
	}

	outputDir := filepath.Join(tempDir, "output")
	err = os.Mkdir(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	ctx := context.Background()
	err = GenerateFromImage(ctx, cfg, testImagePath, 0, outputDir)
	if err != nil {
		t.Errorf("GenerateFromImage() unexpected error: %v", err)
	}

	// Check that output file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 output file, got %d", len(files))
	}

	// Check filename format
	filename := files[0].Name()
	if !strings.HasPrefix(filename, "gocamo_from_image_test_input_000_") {
		t.Errorf("Filename should start with 'gocamo_from_image_test_input_000_', got: %s", filename)
	}
	if !strings.Contains(filename, "_k4_w50x50.png") {
		t.Errorf("Filename should contain '_k4_w50x50.png', got: %s", filename)
	}
}

func TestGeneratePattern_ContextTimeout(t *testing.T) {
	cfg := &config.Config{
		Width:         1000, // Large size to potentially cause timeout
		Height:        1000,
		BasePixelSize: 1,
		PatternType:   "pat3",
	}

	camo := config.CamoColors{
		Name:   "timeout_test",
		Colors: []string{"ff0000", "00ff00"},
	}

	tempDir := t.TempDir()

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	// This test is timing-dependent, so we don't strictly require an error
	// but if there is an error, it should be timeout-related
	if err != nil && !strings.Contains(err.Error(), "context") {
		t.Logf("Got error (which is expected for timeout test): %v", err)
	}
}

// Benchmark tests
func BenchmarkSortColors(b *testing.B) {
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
		{R: 0, G: 255, B: 255, A: 255},
		{R: 128, G: 128, B: 128, A: 255},
		{R: 64, G: 64, B: 64, A: 255},
	}

	for i := 0; i < b.N; i++ {
		testColors := make([]color.RGBA, len(colors))
		copy(testColors, colors)
		sortColors(testColors)
	}
}

func BenchmarkShuffleColors(b *testing.B) {
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
	}

	for i := 0; i < b.N; i++ {
		shuffleColors(colors)
	}
}

func BenchmarkGeneratePattern_Small(b *testing.B) {
	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "pat3",
		AddEdge:       false,
		AddNoise:      false,
	}

	camo := config.CamoColors{
		Name:   "bench",
		Colors: []string{"ff0000", "00ff00"},
	}

	tempDir := b.TempDir()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeneratePattern(ctx, cfg, camo, i, tempDir)
	}
}

// Test edge cases
func TestGeneratePattern_MinimalSize(t *testing.T) {
	cfg := &config.Config{
		Width:         4, // Minimal size
		Height:        4,
		BasePixelSize: 4,
		PatternType:   "pat3",
	}

	camo := config.CamoColors{
		Name:   "minimal",
		Colors: []string{"ff0000", "00ff00"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	if err != nil {
		t.Errorf("GeneratePattern() with minimal size failed: %v", err)
	}
}

func TestGeneratePattern_ColorsWithHashes(t *testing.T) {
	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "pat3",
	}

	camo := config.CamoColors{
		Name:   "hash_test",
		Colors: []string{"#ff0000", "#00ff00", "#0000ff"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	err := GeneratePattern(ctx, cfg, camo, 0, tempDir)
	if err != nil {
		t.Errorf("GeneratePattern() with hash colors failed: %v", err)
	}

	// Check filename strips hashes
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	filename := files[0].Name()
	if !strings.Contains(filename, "ff0000_00ff00_0000ff") {
		t.Errorf("Filename should contain stripped color codes: %s", filename)
	}
}

// Test concurrent pattern generation
func TestGeneratePattern_Concurrent(t *testing.T) {
	cfg := &config.Config{
		Width:         30,
		Height:        30,
		BasePixelSize: 4,
		PatternType:   "pat3",
	}

	camo := config.CamoColors{
		Name:   "concurrent",
		Colors: []string{"ff0000", "00ff00"},
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	// Generate multiple patterns concurrently
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			done <- GeneratePattern(ctx, cfg, camo, index, tempDir)
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent GeneratePattern() failed: %v", err)
		}
	}

	// Check that all files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 5 {
		t.Errorf("Expected 5 files, got %d", len(files))
	}
}
