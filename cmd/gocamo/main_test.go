package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bradsec/gocamo/pkg/config"
)

func TestRun_InvalidPatternType(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "invalid",
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for invalid pattern type, got nil")
	}
	if !strings.Contains(err.Error(), "invalid pattern type") {
		t.Errorf("expected error about invalid pattern type, got: %v", err)
	}
}

func TestRun_EmptyColorsString(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "pat3",
		ColorsString:  "   ",
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for empty colors string, got nil")
	}
	if !strings.Contains(err.Error(), "no valid colors provided") {
		t.Errorf("expected error about no valid colors, got: %v", err)
	}
}

func TestRun_NoInputSpecified(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "pat3",
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for no input specified, got nil")
	}
	if !strings.Contains(err.Error(), "no input specified") {
		t.Errorf("expected error about no input specified, got: %v", err)
	}
}

func TestRun_InvalidOutputDirectory(t *testing.T) {
	// Create a test that tries to write to an invalid directory
	invalidPath := "/invalid/path/that/does/not/exist"
	if os.Getuid() == 0 { // Skip on root user as they might have broader permissions
		t.Skip("Skipping test as root user")
	}

	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     invalidPath,
		PatternType:   "pat3",
		ColorsString:  "ff0000,00ff00",
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for invalid output directory, got nil")
	}
}

func TestRun_ImageDirectoryNotFound(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "image",
		ImageDir:      "/nonexistent/directory",
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for nonexistent image directory, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get image files") {
		t.Errorf("expected error about failed to get image files, got: %v", err)
	}
}

func TestRun_EmptyImageDirectory(t *testing.T) {
	// Create empty temp directory
	emptyDir := t.TempDir()

	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "image",
		ImageDir:      emptyDir,
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for empty image directory, got nil")
	}
	if !strings.Contains(err.Error(), "no image files found") {
		t.Errorf("expected error about no image files found, got: %v", err)
	}
}

func TestRun_ValidPat3PatternWithColors(t *testing.T) {
	outputDir := t.TempDir()
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     outputDir,
		PatternType:   "pat3",
		ColorsString:  "ff0000,00ff00,0000ff",
		Cores:         1,
		AddEdge:       false,
		AddNoise:      false,
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("unexpected error for valid pat3 pattern: %v", err)
	}

	// Check if output file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}
	if len(files) == 0 {
		t.Error("no output files were created")
	}

	// Check if the filename contains expected elements
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "gocamo_") && strings.HasSuffix(file.Name(), ".png") {
			found = true
			break
		}
	}
	if !found {
		t.Error("no PNG file with expected naming pattern was found")
	}
}

func TestRun_ValidPat4PatternWithJSON(t *testing.T) {
	outputDir := t.TempDir()

	// Create a test JSON file
	jsonContent := `[{"name": "test", "colors": ["#ff0000", "#00ff00", "#0000ff"]}]`
	jsonFile := filepath.Join(t.TempDir(), "test.json")
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test JSON file: %v", err)
	}

	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     outputDir,
		PatternType:   "pat4",
		JSONFile:      jsonFile,
		Cores:         1,
		AddEdge:       false,
		AddNoise:      false,
	}

	err = run(cfg)
	if err != nil {
		t.Errorf("unexpected error for valid pat4 pattern: %v", err)
	}

	// Check if output file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}
	if len(files) == 0 {
		t.Error("no output files were created")
	}
}

func TestRun_InvalidJSONFile(t *testing.T) {
	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "pat3",
		JSONFile:      "/nonexistent/file.json",
		Cores:         1,
	}

	err := run(cfg)
	if err == nil {
		t.Error("expected error for nonexistent JSON file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open JSON file") {
		t.Errorf("expected error about failed to open JSON file, got: %v", err)
	}
}

func TestRun_InvalidJSONContent(t *testing.T) {
	// Create a test JSON file with invalid content
	jsonFile := filepath.Join(t.TempDir(), "invalid.json")
	err := os.WriteFile(jsonFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test JSON file: %v", err)
	}

	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "pat3",
		JSONFile:      jsonFile,
		Cores:         1,
	}

	err = run(cfg)
	if err == nil {
		t.Error("expected error for invalid JSON content, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode JSON") {
		t.Errorf("expected error about failed to decode JSON, got: %v", err)
	}
}

func TestRun_EmptyJSONPalettes(t *testing.T) {
	// Create a test JSON file with empty array
	jsonFile := filepath.Join(t.TempDir(), "empty.json")
	err := os.WriteFile(jsonFile, []byte("[]"), 0644)
	if err != nil {
		t.Fatalf("failed to create test JSON file: %v", err)
	}

	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		OutputDir:     t.TempDir(),
		PatternType:   "pat3",
		JSONFile:      jsonFile,
		Cores:         1,
	}

	err = run(cfg)
	if err == nil {
		t.Error("expected error for empty JSON palettes, got nil")
	}
	if !strings.Contains(err.Error(), "no color palettes found") {
		t.Errorf("expected error about no color palettes found, got: %v", err)
	}
}

func TestRun_WithEdgeAndNoise(t *testing.T) {
	outputDir := t.TempDir()
	cfg := &config.Config{
		Width:         50, // Small size for faster test
		Height:        50,
		BasePixelSize: 4,
		OutputDir:     outputDir,
		PatternType:   "pat3",
		ColorsString:  "ff0000,00ff00",
		Cores:         1,
		AddEdge:       true,
		AddNoise:      true,
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("unexpected error for pattern with edge and noise: %v", err)
	}

	// Check if output file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}
	if len(files) == 0 {
		t.Error("no output files were created")
	}
}

// Test to ensure output directory is created successfully
func TestRun_CreatesOutputDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "new_output_dir")

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		OutputDir:     outputDir,
		PatternType:   "pat3",
		ColorsString:  "ff0000,00ff00",
		Cores:         1,
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}
}

// Test timing to ensure reasonable performance
func TestRun_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	outputDir := t.TempDir()
	cfg := &config.Config{
		Width:         200,
		Height:        200,
		BasePixelSize: 4,
		OutputDir:     outputDir,
		PatternType:   "pat3",
		ColorsString:  "ff0000,00ff00,0000ff",
		Cores:         1,
		AddEdge:       false,
		AddNoise:      false,
	}

	start := time.Now()
	err := run(cfg)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Reasonable performance expectation (adjust as needed)
	if duration > 10*time.Second {
		t.Errorf("run took too long: %v", duration)
	}
}
