package utils

import (
	"image/color"
	"reflect"
	"strings"
	"testing"
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
		{"uppercase", "#FF0000", "FF0000"},
		{"short form", "#f00", "f00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripHash(tt.input)
			if result != tt.expected {
				t.Errorf("StripHash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectR   uint8
		expectG   uint8
		expectB   uint8
		expectErr bool
		errMsg    string
	}{
		{"valid 6-char hex", "ff0000", 255, 0, 0, false, ""},
		{"valid 6-char hex with hash", "#ff0000", 255, 0, 0, false, ""},
		{"valid 3-char hex", "f00", 255, 0, 0, false, ""},
		{"valid 3-char hex with hash", "#f00", 255, 0, 0, false, ""},
		{"green color", "00ff00", 0, 255, 0, false, ""},
		{"blue color", "0000ff", 0, 0, 255, false, ""},
		{"white color", "ffffff", 255, 255, 255, false, ""},
		{"black color", "000000", 0, 0, 0, false, ""},
		{"gray color", "808080", 128, 128, 128, false, ""},
		{"uppercase", "FF0000", 255, 0, 0, false, ""},
		{"mixed case", "Ff0000", 255, 0, 0, false, ""},
		{"short green", "0f0", 0, 255, 0, false, ""},
		{"short blue", "00f", 0, 0, 255, false, ""},
		{"short white", "fff", 255, 255, 255, false, ""},
		{"short black", "000", 0, 0, 0, false, ""},
		{"invalid length 5", "ff000", 0, 0, 0, true, "invalid hex color length"},
		{"invalid length 7", "ff00000", 0, 0, 0, true, "invalid hex color length"},
		{"invalid length 2", "ff", 0, 0, 0, true, "invalid hex color length"},
		{"invalid length 1", "f", 0, 0, 0, true, "invalid hex color length"},
		{"invalid character", "gg0000", 0, 0, 0, true, "invalid hex color format"},
		{"invalid character z", "ff000z", 0, 0, 0, true, "invalid hex color format"},
		{"empty string", "", 0, 0, 0, true, "invalid hex color length"},
		{"only hash", "#", 0, 0, 0, true, "invalid hex color length"},
		{"space in hex", "ff 000", 0, 0, 0, true, "invalid hex color length"},
		{"special characters", "ff@000", 0, 0, 0, true, "invalid hex color format"},
		{"invalid short form", "ggg", 0, 0, 0, true, "invalid hex color format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, err := hexToRGB(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("hexToRGB(%q) expected error, got nil", tt.input)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("hexToRGB(%q) error = %v, want error containing %q", tt.input, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("hexToRGB(%q) unexpected error: %v", tt.input, err)
				} else {
					if r != tt.expectR || g != tt.expectG || b != tt.expectB {
						t.Errorf("hexToRGB(%q) = (%d, %d, %d), want (%d, %d, %d)",
							tt.input, r, g, b, tt.expectR, tt.expectG, tt.expectB)
					}
				}
			}
		})
	}
}

func TestHexToRGBA(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		expected  []color.RGBA
		expectErr bool
		errMsg    string
	}{
		{
			name:  "valid colors",
			input: []string{"ff0000", "00ff00", "0000ff"},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},
				{R: 0, G: 255, B: 0, A: 255},
				{R: 0, G: 0, B: 255, A: 255},
			},
			expectErr: false,
		},
		{
			name:  "valid colors with hash",
			input: []string{"#ff0000", "#00ff00", "#0000ff"},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},
				{R: 0, G: 255, B: 0, A: 255},
				{R: 0, G: 0, B: 255, A: 255},
			},
			expectErr: false,
		},
		{
			name:  "short form colors",
			input: []string{"f00", "0f0", "00f"},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},
				{R: 0, G: 255, B: 0, A: 255},
				{R: 0, G: 0, B: 255, A: 255},
			},
			expectErr: false,
		},
		{
			name:  "mixed format",
			input: []string{"#ff0000", "00ff00", "#f00"},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},
				{R: 0, G: 255, B: 0, A: 255},
				{R: 255, G: 0, B: 0, A: 255},
			},
			expectErr: false,
		},
		{
			name:  "grayscale colors",
			input: []string{"000000", "808080", "ffffff"},
			expected: []color.RGBA{
				{R: 0, G: 0, B: 0, A: 255},
				{R: 128, G: 128, B: 128, A: 255},
				{R: 255, G: 255, B: 255, A: 255},
			},
			expectErr: false,
		},
		{
			name:      "only one color",
			input:     []string{"ff0000"},
			expected:  nil,
			expectErr: true,
			errMsg:    "at least 2 colors are required",
		},
		{
			name:      "empty slice",
			input:     []string{},
			expected:  nil,
			expectErr: true,
			errMsg:    "at least 2 colors are required",
		},
		{
			name:      "invalid color in slice",
			input:     []string{"ff0000", "invalid", "0000ff"},
			expected:  nil,
			expectErr: true,
			errMsg:    "invalid hex color",
		},
		{
			name:      "color with spaces",
			input:     []string{"ff0000", " 00ff00 ", "0000ff"},
			expected:  nil,
			expectErr: true,
			errMsg:    "invalid hex color",
		},
		{
			name:  "two valid colors minimum",
			input: []string{"ff0000", "00ff00"},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},
				{R: 0, G: 255, B: 0, A: 255},
			},
			expectErr: false,
		},
		{
			name:  "uppercase colors",
			input: []string{"FF0000", "00FF00", "0000FF"},
			expected: []color.RGBA{
				{R: 255, G: 0, B: 0, A: 255},
				{R: 0, G: 255, B: 0, A: 255},
				{R: 0, G: 0, B: 255, A: 255},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HexToRGBA(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("HexToRGBA(%v) expected error, got nil", tt.input)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("HexToRGBA(%v) error = %v, want error containing %q", tt.input, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("HexToRGBA(%v) unexpected error: %v", tt.input, err)
				} else if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("HexToRGBA(%v) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// Test edge cases and special scenarios
func TestHexToRGB_EdgeCases(t *testing.T) {
	// Test with various whitespace
	r, g, b, err := hexToRGB("  ff0000  ")
	if err != nil {
		t.Errorf("hexToRGB with whitespace should work after trimming, got error: %v", err)
	}
	if r != 255 || g != 0 || b != 0 {
		t.Errorf("hexToRGB('  ff0000  ') = (%d, %d, %d), want (255, 0, 0)", r, g, b)
	}

	// Test short form expansion
	r, g, b, err = hexToRGB("a5f")
	if err != nil {
		t.Errorf("hexToRGB('a5f') unexpected error: %v", err)
	}
	expectedR, expectedG, expectedB := uint8(0xaa), uint8(0x55), uint8(0xff)
	if r != expectedR || g != expectedG || b != expectedB {
		t.Errorf("hexToRGB('a5f') = (%d, %d, %d), want (%d, %d, %d)",
			r, g, b, expectedR, expectedG, expectedB)
	}
}

func TestHexToRGBA_EmptyStrings(t *testing.T) {
	// Test with empty strings in slice
	input := []string{"ff0000", "", "0000ff"}
	_, err := HexToRGBA(input)
	if err == nil {
		t.Error("HexToRGBA with empty string should return error")
	}
}

// Test specific color values
func TestHexToRGB_SpecificColors(t *testing.T) {
	colors := map[string][3]uint8{
		"ff0000": {255, 0, 0},     // Red
		"00ff00": {0, 255, 0},     // Green
		"0000ff": {0, 0, 255},     // Blue
		"ffff00": {255, 255, 0},   // Yellow
		"ff00ff": {255, 0, 255},   // Magenta
		"00ffff": {0, 255, 255},   // Cyan
		"800000": {128, 0, 0},     // Maroon
		"008000": {0, 128, 0},     // Green (dark)
		"000080": {0, 0, 128},     // Navy
		"808000": {128, 128, 0},   // Olive
		"800080": {128, 0, 128},   // Purple
		"008080": {0, 128, 128},   // Teal
		"c0c0c0": {192, 192, 192}, // Silver
		"404040": {64, 64, 64},    // Dark gray
	}

	for hex, expected := range colors {
		t.Run(hex, func(t *testing.T) {
			r, g, b, err := hexToRGB(hex)
			if err != nil {
				t.Errorf("hexToRGB(%q) unexpected error: %v", hex, err)
			}
			if r != expected[0] || g != expected[1] || b != expected[2] {
				t.Errorf("hexToRGB(%q) = (%d, %d, %d), want (%d, %d, %d)",
					hex, r, g, b, expected[0], expected[1], expected[2])
			}
		})
	}
}

// Benchmark tests
func BenchmarkStripHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StripHash("#ff0000")
	}
}

func BenchmarkHexToRGB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hexToRGB("ff0000")
	}
}

func BenchmarkHexToRGBA(b *testing.B) {
	colors := []string{"ff0000", "00ff00", "0000ff", "ffff00", "ff00ff"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HexToRGBA(colors)
	}
}

func BenchmarkHexToRGB_ShortForm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hexToRGB("f00")
	}
}

// Test concurrent access (go routines)
func TestHexToRGBA_Concurrent(t *testing.T) {
	colors := []string{"ff0000", "00ff00", "0000ff"}

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := HexToRGBA(colors)
			if err != nil {
				t.Errorf("Concurrent HexToRGBA failed: %v", err)
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Test alpha channel is always 255
func TestHexToRGBA_AlphaChannel(t *testing.T) {
	colors := []string{"ff0000", "00ff00", "0000ff", "000000", "ffffff"}
	rgba, err := HexToRGBA(colors)
	if err != nil {
		t.Fatalf("HexToRGBA failed: %v", err)
	}

	for i, c := range rgba {
		if c.A != 255 {
			t.Errorf("Color %d alpha = %d, want 255", i, c.A)
		}
	}
}

// Test very large slice
func TestHexToRGBA_LargeSlice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large slice test in short mode")
	}

	// Create a large slice of colors
	colors := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		colors[i] = "ff0000" // All red for simplicity
	}

	rgba, err := HexToRGBA(colors)
	if err != nil {
		t.Fatalf("HexToRGBA with large slice failed: %v", err)
	}

	if len(rgba) != 1000 {
		t.Errorf("Expected 1000 colors, got %d", len(rgba))
	}

	// Check first and last colors
	if rgba[0] != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Errorf("First color incorrect: %v", rgba[0])
	}
	if rgba[999] != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Errorf("Last color incorrect: %v", rgba[999])
	}
}
