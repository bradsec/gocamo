package utils

import (
	"fmt"
	"image/color"
	"strings"
)

func HexToRGBA(hexColors []string) ([]color.RGBA, error) {
	if len(hexColors) < 2 {
		return nil, fmt.Errorf("at least 2 colors are required, got %d", len(hexColors))
	}

	rgbaColors := make([]color.RGBA, len(hexColors))
	for i, hex := range hexColors {
		hex = strings.TrimSpace(hex)
		r, g, b, err := hexToRGB(hex)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color %s: %w", hex, err)
		}
		rgbaColors[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}
	return rgbaColors, nil
}

func hexToRGB(hex string) (uint8, uint8, uint8, error) {
	hex = stripHash(strings.TrimSpace(hex))

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

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color format: %s", hex)
	}
	return r, g, b, nil
}

func stripHash(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}
