package utils

import (
	"fmt"
	"image/color"
	"strings"
)

// HexToRGBA converts a slice of hexadecimal color strings to RGBA color values.
// It accepts hex colors in both 6-character (#RRGGBB) and 3-character (#RGB) formats,
// with or without the leading hash symbol. The 3-character format is automatically
// expanded to 6 characters (e.g., #RGB becomes #RRGGBB).
//
// The function validates that at least 2 colors are provided and that each color
// string represents a valid hexadecimal color. All returned RGBA colors have an
// alpha value of 255 (fully opaque).
//
// Parameters:
//   - hexColors: A slice of hex color strings (e.g., ["#FF0000", "00FF00", "#00F"])
//
// Returns a slice of color.RGBA values and an error if fewer than 2 colors are provided
// or if any color string is invalid.
//
// Example:
//
//	colors, err := HexToRGBA([]string{"#FF0000", "#00FF00", "#0000FF"})
//	// Returns red, green, and blue RGBA colors
func HexToRGBA(hexColors []string) ([]color.RGBA, error) {
	if len(hexColors) < 2 {
		return nil, fmt.Errorf("at least 2 colors are required, got %d", len(hexColors))
	}

	rgbaColors := make([]color.RGBA, len(hexColors))
	for i, hex := range hexColors {
		// Check if hex string has leading or trailing spaces - this should be rejected for arrays
		trimmed := strings.TrimSpace(hex)
		if hex != trimmed {
			return nil, fmt.Errorf("invalid hex color %s: contains leading or trailing spaces", hex)
		}
		r, g, b, err := hexToRGB(hex)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color %s: %w", hex, err)
		}
		rgbaColors[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}
	return rgbaColors, nil
}

func hexToRGB(hex string) (uint8, uint8, uint8, error) {
	originalHex := hex
	hex = StripHash(strings.TrimSpace(hex))

	// Check for internal spaces in the middle of hex string
	// "ff 000" should fail with length error, "  ff0000  " should work
	strippedOriginal := StripHash(originalHex)
	trimmedOriginal := strings.TrimSpace(strippedOriginal)
	if strings.Contains(trimmedOriginal, " ") {
		// Has spaces in the middle - this should be a length error for "ff 000"
		return 0, 0, 0, fmt.Errorf("invalid hex color length: %s (should be 6 characters or 3 for short form)", trimmedOriginal)
	}

	// Handle short form (#RGB)
	if len(hex) == 3 {
		r := hex[0]
		g := hex[1]
		b := hex[2]
		// Expand short form to full form (#RGB -> #RRGGBB)
		hex = string([]byte{r, r, g, g, b, b})
	}

	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color length: %s (should be 6 characters or 3 for short form)", hex)
	}

	// Validate that all characters are valid hex digits
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return 0, 0, 0, fmt.Errorf("invalid hex color format: %s", hex)
		}
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color format: %s", hex)
	}
	return r, g, b, nil
}

// StripHash removes the leading '#' character from a hex color string if present.
// It returns the hex string without the hash prefix, or the original string if no hash is found.
func StripHash(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}
