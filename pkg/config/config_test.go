package config

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/bradsec/gocamo/internal/utils"
)

func TestStripHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with hash", "#ff0000", "ff0000"},
		{"without hash", "ff0000", "ff0000"},
		{"empty string", "", ""},
		{"only hash", "#", ""},
		{"multiple hashes", "##ff0000", "#ff0000"},
		{"hash in middle", "ff#0000", "ff#0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.StripHash(tt.input)
			if result != tt.expected {
				t.Errorf("utils.StripHash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateHexColor(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		errMsg    string
	}{
		{"valid 6-char hex", "ff0000", false, ""},
		{"valid 6-char hex with hash", "#ff0000", false, ""},
		{"valid 3-char hex", "f00", false, ""},
		{"valid 3-char hex with hash", "#f00", false, ""},
		{"uppercase valid", "FF0000", false, ""},
		{"mixed case valid", "Ff0000", false, ""},
		{"invalid length 5", "ff000", true, "invalid hex color length"},
		{"invalid length 7", "ff00000", true, "invalid hex color length"},
		{"invalid length 2", "ff", true, "invalid hex color length"},
		{"invalid character", "gg0000", true, "invalid hex color character"},
		{"invalid character z", "ff000z", true, "invalid hex color character"},
		{"empty string", "", true, "invalid hex color length"},
		{"only hash", "#", true, "invalid hex color length"},
		{"space in hex", "ff 000", true, "invalid hex color character"},
		{"special characters", "ff@000", true, "invalid hex color character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHexColor(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("validateHexColor(%q) expected error, got nil", tt.input)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateHexColor(%q) error = %v, want error containing %q", tt.input, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateHexColor(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestCleanColorString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
		errMsg    string
	}{
		{"valid colors", "ff0000,00ff00,0000ff", "ff0000,00ff00,0000ff", false, ""},
		{"colors with spaces", "ff0000, 00ff00 , 0000ff", "ff0000,00ff00,0000ff", false, ""},
		{"colors with hashes", "#ff0000,#00ff00,#0000ff", "#ff0000,#00ff00,#0000ff", false, ""},
		{"mixed format", "#ff0000, 00ff00 ,#0000ff", "#ff0000,00ff00,#0000ff", false, ""},
		{"short form", "f00,0f0,00f", "f00,0f0,00f", false, ""},
		{"single color", "ff0000", "", true, "at least 2 colors are required"},
		{"empty string", "", "", true, "at least 2 colors are required"},
		{"only spaces", "   ", "", true, "at least 2 colors are required"},
		{"empty parts", "ff0000,,00ff00", "ff0000,00ff00", false, ""},
		{"invalid color", "ff0000,invalid,0000ff", "", true, "invalid color"},
		{"too short color", "ff0000,ff00,0000ff", "", true, "invalid color"},
		{"invalid character", "ff0000,gg0000,0000ff", "", true, "invalid color"},
		{"trailing comma", "ff0000,00ff00,", "ff0000,00ff00", false, ""},
		{"leading comma", ",ff0000,00ff00", "ff0000,00ff00", false, ""},
		{"multiple spaces", "ff0000  ,  00ff00  ,  0000ff", "ff0000,00ff00,0000ff", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cleanColorString(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("cleanColorString(%q) expected error, got nil", tt.input)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("cleanColorString(%q) error = %v, want error containing %q", tt.input, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("cleanColorString(%q) unexpected error: %v", tt.input, err)
				} else if result != tt.expected {
					t.Errorf("cleanColorString(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestIsFlagPassed(t *testing.T) {
	// Save original CommandLine
	originalCommandLine := flag.CommandLine
	defer func() {
		flag.CommandLine = originalCommandLine
	}()

	// Create a new FlagSet for testing
	testFlags := flag.NewFlagSet("test", flag.ContinueOnError)
	flag.CommandLine = testFlags

	var testFlag string
	testFlags.StringVar(&testFlag, "testflag", "", "test flag")

	// Test flag not passed
	testFlags.Parse([]string{})
	if isFlagPassed("testflag") {
		t.Error("isFlagPassed should return false for flag not passed")
	}

	// Test flag passed
	testFlags = flag.NewFlagSet("test", flag.ContinueOnError)
	flag.CommandLine = testFlags
	testFlags.StringVar(&testFlag, "testflag", "", "test flag")
	testFlags.Parse([]string{"-testflag", "value"})
	if !isFlagPassed("testflag") {
		t.Error("isFlagPassed should return true for flag passed")
	}

	// Test non-existent flag
	if isFlagPassed("nonexistent") {
		t.Error("isFlagPassed should return false for non-existent flag")
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	// Save original CommandLine and args
	originalCommandLine := flag.CommandLine
	originalArgs := os.Args
	defer func() {
		flag.CommandLine = originalCommandLine
		os.Args = originalArgs
	}()

	// Set up for testing with no arguments
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	cfg := ParseFlags()

	// Test default values
	if cfg.Width != 1500 {
		t.Errorf("default Width = %d, want 1500", cfg.Width)
	}
	if cfg.Height != 1500 {
		t.Errorf("default Height = %d, want 1500", cfg.Height)
	}
	if cfg.BasePixelSize != 4 {
		t.Errorf("default BasePixelSize = %d, want 4", cfg.BasePixelSize)
	}
	if cfg.OutputDir != "output" {
		t.Errorf("default OutputDir = %q, want %q", cfg.OutputDir, "output")
	}
	if cfg.Cores != runtime.NumCPU() {
		t.Errorf("default Cores = %d, want %d", cfg.Cores, runtime.NumCPU())
	}
	if cfg.PatternType != "box" {
		t.Errorf("default PatternType = %q, want %q", cfg.PatternType, "box")
	}
	if cfg.ImageDir != "input" {
		t.Errorf("default ImageDir = %q, want %q", cfg.ImageDir, "input")
	}
	if cfg.KValue != 4 {
		t.Errorf("default KValue = %d, want 4", cfg.KValue)
	}
	if cfg.AddEdge != false {
		t.Errorf("default AddEdge = %v, want false", cfg.AddEdge)
	}
	if cfg.AddNoise != false {
		t.Errorf("default AddNoise = %v, want false", cfg.AddNoise)
	}
}

func TestConfig_ValidationCores(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"negative cores", -1, 1},
		{"zero cores", 0, 1},
		{"valid cores", 2, 2},
		{"max cores", runtime.NumCPU(), runtime.NumCPU()},
		{"too many cores", runtime.NumCPU() + 5, runtime.NumCPU()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original CommandLine and args
			originalCommandLine := flag.CommandLine
			originalArgs := os.Args
			defer func() {
				flag.CommandLine = originalCommandLine
				os.Args = originalArgs
			}()

			// Set up test arguments
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

			// For proper testing, we need to pass the cores value as a string
			args := []string{"test", "-cores", fmt.Sprintf("%d", tt.input)}
			os.Args = args

			cfg := ParseFlags()
			if cfg.Cores != tt.expected {
				t.Errorf("Cores validation: input %d, got %d, want %d", tt.input, cfg.Cores, tt.expected)
			}
		})
	}
}

func TestConfig_ValidationDimensions(t *testing.T) {
	// Save original CommandLine and args
	originalCommandLine := flag.CommandLine
	originalArgs := os.Args
	defer func() {
		flag.CommandLine = originalCommandLine
		os.Args = originalArgs
	}()

	tests := []struct {
		name           string
		width, height  string
		expectedWidth  int
		expectedHeight int
	}{
		{"negative dimensions", "-1", "-1", 1500, 1500}, // Should use defaults
		{"zero dimensions", "0", "0", 1500, 1500},       // Should use defaults
		{"valid dimensions", "800", "600", 800, 600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{"test", "-w", tt.width, "-h", tt.height}
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

			cfg := ParseFlags()
			if cfg.Width != tt.expectedWidth {
				t.Errorf("Width validation: got %d, want %d", cfg.Width, tt.expectedWidth)
			}
			if cfg.Height != tt.expectedHeight {
				t.Errorf("Height validation: got %d, want %d", cfg.Height, tt.expectedHeight)
			}
		})
	}
}

func TestConfig_ImageFlagSetsPatternType(t *testing.T) {
	// Save original CommandLine and args
	originalCommandLine := flag.CommandLine
	originalArgs := os.Args
	defer func() {
		flag.CommandLine = originalCommandLine
		os.Args = originalArgs
	}()

	os.Args = []string{"test", "-i", "test_input"}
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	cfg := ParseFlags()
	if cfg.PatternType != "image" {
		t.Errorf("PatternType should be 'image' when -i flag is used, got %q", cfg.PatternType)
	}
	if cfg.ImageDir != "test_input" {
		t.Errorf("ImageDir should be 'test_input', got %q", cfg.ImageDir)
	}
}

// Test that we can create a valid Config struct
func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		Width:         1000,
		Height:        800,
		BasePixelSize: 8,
		JSONFile:      "test.json",
		OutputDir:     "test_output",
		ColorsString:  "ff0000,00ff00",
		Cores:         4,
		AddEdge:       true,
		AddNoise:      true,
		PatternType:   "blob",
		ImageDir:      "test_images",
		KValue:        6,
	}

	// Test that all fields are accessible
	if cfg.Width != 1000 {
		t.Errorf("Width = %d, want 1000", cfg.Width)
	}
	if cfg.Height != 800 {
		t.Errorf("Height = %d, want 800", cfg.Height)
	}
	if cfg.BasePixelSize != 8 {
		t.Errorf("BasePixelSize = %d, want 8", cfg.BasePixelSize)
	}
	if cfg.JSONFile != "test.json" {
		t.Errorf("JSONFile = %q, want %q", cfg.JSONFile, "test.json")
	}
	if cfg.OutputDir != "test_output" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "test_output")
	}
	if cfg.ColorsString != "ff0000,00ff00" {
		t.Errorf("ColorsString = %q, want %q", cfg.ColorsString, "ff0000,00ff00")
	}
	if cfg.Cores != 4 {
		t.Errorf("Cores = %d, want 4", cfg.Cores)
	}
	if !cfg.AddEdge {
		t.Error("AddEdge should be true")
	}
	if !cfg.AddNoise {
		t.Error("AddNoise should be true")
	}
	if cfg.PatternType != "blob" {
		t.Errorf("PatternType = %q, want %q", cfg.PatternType, "blob")
	}
	if cfg.ImageDir != "test_images" {
		t.Errorf("ImageDir = %q, want %q", cfg.ImageDir, "test_images")
	}
	if cfg.KValue != 6 {
		t.Errorf("KValue = %d, want 6", cfg.KValue)
	}
}

func TestCamoColors_Struct(t *testing.T) {
	camo := CamoColors{
		Name:   "test_pattern",
		Colors: []string{"#ff0000", "#00ff00", "#0000ff"},
	}

	if camo.Name != "test_pattern" {
		t.Errorf("Name = %q, want %q", camo.Name, "test_pattern")
	}
	if len(camo.Colors) != 3 {
		t.Errorf("Colors length = %d, want 3", len(camo.Colors))
	}
	expectedColors := []string{"#ff0000", "#00ff00", "#0000ff"}
	for i, color := range camo.Colors {
		if color != expectedColors[i] {
			t.Errorf("Colors[%d] = %q, want %q", i, color, expectedColors[i])
		}
	}
}

// Benchmark tests
func BenchmarkStripHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.StripHash("#ff0000")
	}
}

func BenchmarkValidateHexColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validateHexColor("ff0000")
	}
}

func BenchmarkCleanColorString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cleanColorString("ff0000,00ff00,0000ff")
	}
}

// Edge case testing
func TestValidateHexColor_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"whitespace only", "   ", false},
		{"newline", "\n", false},
		{"tab", "\t", false},
		{"with spaces", " ff0000 ", false}, // Spaces should be trimmed but still invalid chars
		{"unicode", "ff000ü", false},
		{"very long", "ff0000ff0000ff0000", false},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateHexColor(tc.input)
			if tc.valid && err != nil {
				t.Errorf("validateHexColor(%q) should be valid, got error: %v", tc.input, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("validateHexColor(%q) should be invalid, got nil error", tc.input)
			}
		})
	}
}
