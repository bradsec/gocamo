package generator

import (
	"image"
	"image/color"
	"math/rand"
)

func addNoiseRGBA(img *image.RGBA, colors []color.RGBA) {
	if img == nil || len(colors) == 0 {
		return
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if rand.Float32() < 0.05 { // 5% chance to add noise
				noiseColor := colors[rand.Intn(len(colors))]
				currentColor := img.RGBAAt(x, y)

				// Blend the current color with the noise color
				r := uint8((int(currentColor.R) + int(noiseColor.R)) / 2)
				g := uint8((int(currentColor.G) + int(noiseColor.G)) / 2)
				b := uint8((int(currentColor.B) + int(noiseColor.B)) / 2)

				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}
}

func addNoiseNRGBA(img *image.NRGBA, colors []color.RGBA) {
	if img == nil || len(colors) == 0 {
		return
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if rand.Float32() < 0.05 { // 5% chance to add noise
				noiseColor := colors[rand.Intn(len(colors))]
				currentColor := img.NRGBAAt(x, y)

				// Blend the current color with the noise color
				r := uint8((int(currentColor.R) + int(noiseColor.R)) / 2)
				g := uint8((int(currentColor.G) + int(noiseColor.G)) / 2)
				b := uint8((int(currentColor.B) + int(noiseColor.B)) / 2)

				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}
}

func addEdgeDetailsRGBA(img *image.RGBA, basePixelSize int) {
	if img == nil || basePixelSize <= 0 {
		return
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Check if we're near edge boundaries (within 2 pixels)
			nearEdgeX := (x%basePixelSize) <= 1 || (x%basePixelSize) >= (basePixelSize-2)
			nearEdgeY := (y%basePixelSize) <= 1 || (y%basePixelSize) >= (basePixelSize-2)

			if nearEdgeX || nearEdgeY {
				if rand.Float32() < 0.2 { // Reduced from 40% to 20% for subtlety
					currentColor := img.RGBAAt(x, y)

					// Much gentler variation - only ±8 instead of ±20
					variation := rand.Intn(17) - 8 // -8 to +8

					// Apply variation more subtly based on distance from edge
					edgeDistX := min(x%basePixelSize, basePixelSize-(x%basePixelSize))
					edgeDistY := min(y%basePixelSize, basePixelSize-(y%basePixelSize))
					edgeDist := min(edgeDistX, edgeDistY)

					// Fade the effect as we move away from edges
					fadeStrength := 1.0 - float64(edgeDist)/2.0
					if fadeStrength < 0 {
						fadeStrength = 0
					}

					adjustedVariation := int(float64(variation) * fadeStrength)

					r := uint8(clamp(int(currentColor.R)+adjustedVariation, 0, 255))
					g := uint8(clamp(int(currentColor.G)+adjustedVariation, 0, 255))
					b := uint8(clamp(int(currentColor.B)+adjustedVariation, 0, 255))
					img.Set(x, y, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func addEdgeDetailsNRGBA(img *image.NRGBA, basePixelSize int) {
	if img == nil || basePixelSize <= 0 {
		return
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x%basePixelSize == 0 || y%basePixelSize == 0 {
				if rand.Float32() < 0.4 { // 40% chance for edge details
					currentColor := img.NRGBAAt(x, y)
					r := uint8(clamp(int(currentColor.R)+rand.Intn(41)-20, 0, 255))
					g := uint8(clamp(int(currentColor.G)+rand.Intn(41)-20, 0, 255))
					b := uint8(clamp(int(currentColor.B)+rand.Intn(41)-20, 0, 255))
					img.Set(x, y, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// selectWeightedColor selects a color index based on provided ratios/weights.
// If ratios is nil or empty, falls back to random selection.
// Updated to include maxColors parameter for bounds checking
func selectWeightedColor(ratios []float64, maxColors int) int {
	if len(ratios) == 0 || maxColors == 0 {
		return 0
	}

	if len(ratios) == 1 {
		return 0
	}

	// Generate random number between 0 and 1
	r := rand.Float64()

	// Select color based on cumulative probability
	cumulative := 0.0
	for i, ratio := range ratios {
		cumulative += ratio
		if r <= cumulative {
			// Ensure the returned index is within bounds of available colors
			if i >= maxColors {
				return maxColors - 1
			}
			return i
		}
	}

	// Fallback to last color (should rarely happen due to floating point precision)
	result := len(ratios) - 1
	if result >= maxColors {
		return maxColors - 1
	}
	return result
}

// selectWeightedColorExcluding selects a weighted color but excludes specific indices.
// Useful when you want to avoid background colors or specific color combinations.
// Updated to include maxColors parameter for bounds checking
func selectWeightedColorExcluding(ratios []float64, excludeIndices []int, maxColors int) int {
	if len(ratios) == 0 || maxColors == 0 {
		return 0
	}

	// Create a map for quick lookup of excluded indices
	excludeMap := make(map[int]bool)
	for _, idx := range excludeIndices {
		excludeMap[idx] = true
	}

	// Calculate adjusted ratios excluding unwanted colors and ensuring bounds
	adjustedRatios := make([]float64, 0, len(ratios))
	indexMap := make([]int, 0, len(ratios))
	totalRatio := 0.0

	for i, ratio := range ratios {
		// Only consider indices that exist in the colors array and aren't excluded
		if i < maxColors && !excludeMap[i] {
			adjustedRatios = append(adjustedRatios, ratio)
			indexMap = append(indexMap, i)
			totalRatio += ratio
		}
	}

	if len(adjustedRatios) == 0 {
		// If all colors are excluded, return 0 as fallback
		return 0
	}

	// Normalize adjusted ratios
	for i := range adjustedRatios {
		adjustedRatios[i] /= totalRatio
	}

	// Select from adjusted ratios
	selectedIndex := selectWeightedColor(adjustedRatios, len(adjustedRatios))
	return indexMap[selectedIndex]
}
