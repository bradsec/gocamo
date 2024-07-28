package utils

import (
	"fmt"
	"image/color"
)

func HexToRGBA(hexColors []string) ([]color.RGBA, error) {
	rgbaColors := make([]color.RGBA, len(hexColors))
	for i, hex := range hexColors {
		r, g, b, err := hexToRGB(hex)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color %s: %w", hex, err)
		}
		rgbaColors[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}
	return rgbaColors, nil
}

func hexToRGB(hex string) (uint8, uint8, uint8, error) {
	hex = stripHash(hex)
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color length")
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color format")
	}
	return r, g, b, nil
}

func stripHash(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}
