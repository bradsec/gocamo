package generator

import (
	"context"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

// Pat2Generator creates pat2 (Scattered Camouflage) patterns.
// Inspired by MARPAT digital camouflage with multi-scale fractal-like structures.
type Pat2Generator struct{}

// Generate creates pat2 style camouflage inspired by real military patterns like MARPAT.
// Uses multi-scale approach with organic blob shapes and digital pixel structure.
func (pg *Pat2Generator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Use the centralized pixel size adjustment for perfect fit
	pixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	gridWidth := cfg.Width / pixelSize
	gridHeight := cfg.Height / pixelSize

	// Initialize with random background color (any color can be background)
	grid := make([][]int, gridHeight)
	baseColorIndex := rand.Intn(len(colors))
	for y := range grid {
		grid[y] = make([]int, gridWidth)
		for x := range grid[y] {
			grid[y][x] = baseColorIndex
		}
	}

	// Multi-scale approach like MARPAT with weighted color distribution
	// Large scale: Major blobs with directional bias (like MultiCam/OCP)
	pg.generateLargeScaleBlobs(grid, gridWidth, gridHeight, colors, cfg)

	// Add color transitions and gradients (like real MultiCam)
	pg.addColorTransitions(grid, gridWidth, gridHeight, colors, cfg)

	// Medium scale: Secondary patterns with clustering
	pg.generateMediumScaleElements(grid, gridWidth, gridHeight, colors, cfg)

	// Add directional flow patterns (horizontal bias like OCP)
	pg.addDirectionalFlow(grid, gridWidth, gridHeight, colors, cfg)

	// Small scale: Digital noise and micro-patterns (like MARPAT pixels)
	pg.generateSmallScaleNoise(grid, gridWidth, gridHeight, colors, cfg)

	// Add fractal-like self-similarity
	pg.addFractalDetails(grid, gridWidth, gridHeight, colors, cfg)

	// Final blending pass for smoother transitions
	pg.blendTransitions(grid, gridWidth, gridHeight, colors, cfg)

	// Render grid to image
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			gridX := x / pixelSize
			gridY := y / pixelSize

			if gridX < gridWidth && gridY < gridHeight {
				colorIndex := grid[gridY][gridX]
				img.Set(x, y, colors[colorIndex])
			}
		}
	}

	// Apply noise and edge effects if requested (same as other patterns)
	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, pixelSize)
	}

	return img, nil
}

// findLightestColor finds the lightest color to use as background
func (pg *Pat2Generator) findLightestColor(colors []color.RGBA) int {
	maxBrightness := 0
	lightestIndex := 0

	for i, c := range colors {
		brightness := int(c.R) + int(c.G) + int(c.B)
		if brightness > maxBrightness {
			maxBrightness = brightness
			lightestIndex = i
		}
	}

	return lightestIndex
}

// generateLargeScaleBlobs creates major organic blobs with directional bias like MultiCam/OCP
func (pg *Pat2Generator) generateLargeScaleBlobs(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	numBlobs := (gridWidth * gridHeight) / 120 // Slightly more blobs for better coverage

	for i := 0; i < numBlobs; i++ {
		// Bias placement toward certain areas (like real camo clustering)
		var centerX, centerY int
		if rand.Float64() < 0.3 { // 30% chance for edge placement
			centerX = rand.Intn(gridWidth/4) + gridWidth*3/4 // Right edge bias
			centerY = rand.Intn(gridHeight)
		} else {
			centerX = rand.Intn(gridWidth)
			centerY = rand.Intn(gridHeight)
		}

		// Use random color selection
		colorIndex := selectRandomColor(colors)

		// Create organic blob with horizontal stretch (like OCP)
		blobSizeX := 6 + rand.Intn(14) // 6-20 pixels wide
		blobSizeY := 4 + rand.Intn(10) // 4-14 pixels high (horizontally biased)
		pg.drawEllipticalBlob(grid, centerX, centerY, blobSizeX, blobSizeY, colorIndex, gridWidth, gridHeight)
	}
}

// generateMediumScaleElements adds secondary patterns
func (pg *Pat2Generator) generateMediumScaleElements(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	numElements := (gridWidth * gridHeight) / 80

	for i := 0; i < numElements; i++ {
		x := rand.Intn(gridWidth)
		y := rand.Intn(gridHeight)

		colorIndex := selectWeightedColor(cfg.ColorRatios, len(colors))
		size := 3 + rand.Intn(6) // Medium elements (3-8 pixels)

		// Heavily favor organic shapes over angular ones for multicam look
		if rand.Float64() < 0.9 {
			pg.drawOrganicBlob(grid, x, y, size, colorIndex, gridWidth, gridHeight)
		} else {
			// Use more organic angular elements instead of geometric ones
			pg.drawSoftAngularElement(grid, x, y, size, colorIndex, gridWidth, gridHeight)
		}
	}
}

// generateSmallScaleNoise adds organic micro-patterns instead of rectangular pixels
func (pg *Pat2Generator) generateSmallScaleNoise(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	numPixels := (gridWidth * gridHeight) / 30 // Reduced density for more organic look

	for i := 0; i < numPixels; i++ {
		x := rand.Intn(gridWidth)
		y := rand.Intn(gridHeight)

		colorIndex := selectWeightedColor(cfg.ColorRatios, len(colors))

		// Create small organic clusters instead of rectangular pixels
		size := 1 + rand.Intn(2) // 1-2 pixel radius
		pg.drawTinyOrganicCluster(grid, x, y, size, colorIndex, gridWidth, gridHeight)
	}
}

// addFractalDetails adds self-similar patterns at different scales
func (pg *Pat2Generator) addFractalDetails(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	// Add smaller versions of larger patterns
	numFractals := (gridWidth * gridHeight) / 200

	for i := 0; i < numFractals; i++ {
		x := rand.Intn(gridWidth)
		y := rand.Intn(gridHeight)

		colorIndex := selectWeightedColor(cfg.ColorRatios, len(colors))
		size := 2 + rand.Intn(4) // Small fractal elements

		pg.drawFractalElement(grid, x, y, size, colorIndex, gridWidth, gridHeight)
	}
}

// drawOrganicBlob creates organic, natural-looking shapes with enhanced variation
func (pg *Pat2Generator) drawOrganicBlob(grid [][]int, centerX, centerY, size, colorIndex, gridWidth, gridHeight int) {
	// Use enhanced noise for more organic boundaries
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight {
				// Distance from center with organic variation
				distance := math.Sqrt(float64(dx*dx + dy*dy))

				// Multi-scale organic noise for more natural boundaries
				noiseValue1 := pg.simpleNoise(float64(x)*0.08, float64(y)*0.08)
				noiseValue2 := pg.simpleNoise(float64(x)*0.2, float64(y)*0.2) * 0.3
				combinedNoise := noiseValue1 + noiseValue2

				organicRadius := float64(size) + combinedNoise*2.5

				if distance <= organicRadius {
					// Enhanced probability curve for softer, more natural edges
					normalizedDistance := distance / organicRadius
					probability := math.Pow(1.0-normalizedDistance, 1.5)

					// Add some randomness even within the shape for texture
					if rand.Float64() < probability*(0.8+rand.Float64()*0.2) {
						grid[y][x] = colorIndex
					}
				}
			}
		}
	}
}

// drawAngularElement creates slightly angular shapes (inspired by digital camo)
func (pg *Pat2Generator) drawAngularElement(grid [][]int, centerX, centerY, size, colorIndex, gridWidth, gridHeight int) {
	// Create angular but not perfectly geometric shapes
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight {
				// Angular distance calculation
				distance := math.Max(math.Abs(float64(dx)), math.Abs(float64(dy)))

				if distance <= float64(size) && rand.Float64() < 0.8 {
					grid[y][x] = colorIndex
				}
			}
		}
	}
}

// drawFractalElement creates small self-similar patterns
func (pg *Pat2Generator) drawFractalElement(grid [][]int, centerX, centerY, size, colorIndex, gridWidth, gridHeight int) {
	// Create L-shaped or T-shaped elements like in MARPAT
	patterns := [][]struct{ dx, dy int }{
		{{0, 0}, {1, 0}, {0, 1}},                   // L-shape
		{{0, 0}, {-1, 0}, {1, 0}, {0, 1}},          // T-shape
		{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}, // Cross
	}

	pattern := patterns[rand.Intn(len(patterns))]

	for _, offset := range pattern {
		x := centerX + offset.dx
		y := centerY + offset.dy

		if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight {
			grid[y][x] = colorIndex
		}
	}
}

// selectDarkerColor selects a darker color from the palette
func (pg *Pat2Generator) selectDarkerColor(colors []color.RGBA) int {
	darkestIndex := 0
	minBrightness := 999999

	for i, c := range colors {
		brightness := int(c.R) + int(c.G) + int(c.B)
		if brightness < minBrightness {
			minBrightness = brightness
			darkestIndex = i
		}
	}

	return darkestIndex
}

// simpleNoise generates simple noise for organic shapes
func (pg *Pat2Generator) simpleNoise(x, y float64) float64 {
	// Simple pseudo-Perlin noise
	n := math.Sin(x*12.9898+y*78.233) * 43758.5453
	return n - math.Floor(n) - 0.5
}

// drawEllipticalBlob creates horizontally-biased organic shapes like OCP with enhanced natural variation
func (pg *Pat2Generator) drawEllipticalBlob(grid [][]int, centerX, centerY, sizeX, sizeY, colorIndex, gridWidth, gridHeight int) {
	for dy := -sizeY; dy <= sizeY; dy++ {
		for dx := -sizeX; dx <= sizeX; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight {
				// Elliptical distance calculation
				normalizedX := float64(dx) / float64(sizeX)
				normalizedY := float64(dy) / float64(sizeY)
				distance := normalizedX*normalizedX + normalizedY*normalizedY

				// Multi-scale organic noise for more natural, irregular boundaries
				noiseValue1 := pg.simpleNoise(float64(x)*0.12, float64(y)*0.12)
				noiseValue2 := pg.simpleNoise(float64(x)*0.25, float64(y)*0.25) * 0.4
				noiseValue3 := pg.simpleNoise(float64(x)*0.5, float64(y)*0.5) * 0.15
				combinedNoise := noiseValue1 + noiseValue2 + noiseValue3

				organicRadius := 1.0 + combinedNoise*0.6

				if distance <= organicRadius {
					// More natural probability curve for organic appearance
					normalizedDistance := math.Sqrt(distance / organicRadius)
					probability := math.Pow(1.0-normalizedDistance, 1.8)

					// Add internal texture variation
					textureNoise := pg.simpleNoise(float64(x)*0.4, float64(y)*0.4) * 0.1
					finalProbability := probability * (0.85 + textureNoise + rand.Float64()*0.15)

					if rand.Float64() < finalProbability {
						grid[y][x] = colorIndex
					}
				}
			}
		}
	}
}

// addColorTransitions creates gradual color transitions like MultiCam
func (pg *Pat2Generator) addColorTransitions(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	// Find color boundaries and add transition zones
	for y := 1; y < gridHeight-1; y++ {
		for x := 1; x < gridWidth-1; x++ {
			currentColor := grid[y][x]

			// Check for color boundaries
			neighbors := []int{
				grid[y-1][x], grid[y+1][x], grid[y][x-1], grid[y][x+1],
			}

			for _, neighborColor := range neighbors {
				if neighborColor != currentColor && rand.Float64() < 0.15 {
					// Create transition with intermediate color
					transitionColor := pg.getIntermediateColor(currentColor, neighborColor, len(colors))
					if transitionColor != -1 {
						grid[y][x] = transitionColor
					}
				}
			}
		}
	}
}

// addDirectionalFlow creates horizontal flow patterns like OCP
func (pg *Pat2Generator) addDirectionalFlow(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	numFlows := (gridWidth * gridHeight) / 300

	for i := 0; i < numFlows; i++ {
		startX := rand.Intn(gridWidth / 3) // Start from left side
		startY := rand.Intn(gridHeight)
		colorIndex := selectWeightedColor(cfg.ColorRatios, len(colors))

		// Create horizontal flow with slight vertical variation
		flowLength := 8 + rand.Intn(15)
		currentY := float64(startY)

		for step := 0; step < flowLength; step++ {
			x := startX + step
			y := int(currentY)

			if x >= gridWidth || y < 0 || y >= gridHeight {
				break
			}

			if rand.Float64() < 0.7 { // 70% chance to place
				grid[y][x] = colorIndex
			}

			// Add slight vertical variation
			currentY += (rand.Float64() - 0.5) * 0.4
		}
	}
}

// blendTransitions creates smoother color transitions
func (pg *Pat2Generator) blendTransitions(grid [][]int, gridWidth, gridHeight int, colors []color.RGBA, cfg *config.Config) {
	// Single pass smoothing for more natural appearance
	newGrid := make([][]int, gridHeight)
	for y := range newGrid {
		newGrid[y] = make([]int, gridWidth)
		copy(newGrid[y], grid[y])
	}

	for y := 1; y < gridHeight-1; y++ {
		for x := 1; x < gridWidth-1; x++ {
			// Count neighbor colors
			colorCount := make(map[int]int)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					colorCount[grid[y+dy][x+dx]]++
				}
			}

			// Find most common color
			maxCount := 0
			dominantColor := grid[y][x]
			for color, count := range colorCount {
				if count > maxCount {
					maxCount = count
					dominantColor = color
				}
			}

			// Apply gentle smoothing only for strong consensus
			if maxCount >= 6 && rand.Float64() < 0.3 {
				newGrid[y][x] = dominantColor
			}
		}
	}

	// Copy smoothed result back
	for y := range grid {
		copy(grid[y], newGrid[y])
	}
}

// getIntermediateColor finds a color between two colors in the palette
func (pg *Pat2Generator) getIntermediateColor(color1, color2, numColors int) int {
	// Simple intermediate color selection - could be enhanced
	if rand.Float64() < 0.5 {
		// Return a random color that's not the two input colors
		for attempts := 0; attempts < 5; attempts++ {
			candidate := rand.Intn(numColors)
			if candidate != color1 && candidate != color2 {
				return candidate
			}
		}
	}
	return -1 // No intermediate color found
}

// drawSoftAngularElement creates angular shapes with organic edges for more natural look
func (pg *Pat2Generator) drawSoftAngularElement(grid [][]int, centerX, centerY, size, colorIndex, gridWidth, gridHeight int) {
	// Create angular shapes but with organic edge variation
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight {
				// Angular distance calculation with organic modification
				distance := math.Max(math.Abs(float64(dx)), math.Abs(float64(dy)))

				// Add organic noise to soften angular edges
				noiseValue := pg.simpleNoise(float64(x)*0.3, float64(y)*0.3)
				organicSize := float64(size) + noiseValue*1.5

				if distance <= organicSize {
					// Soft probability transition instead of hard edges
					probability := 1.0 - (distance / organicSize)
					probability = math.Pow(probability, 1.2) // Softer falloff

					if rand.Float64() < probability*0.85 {
						grid[y][x] = colorIndex
					}
				}
			}
		}
	}
}

// drawTinyOrganicCluster creates small organic pixel clusters instead of rectangular noise
func (pg *Pat2Generator) drawTinyOrganicCluster(grid [][]int, centerX, centerY, size, colorIndex, gridWidth, gridHeight int) {
	// Create small organic clusters for natural micro-texture
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < gridWidth && y >= 0 && y < gridHeight {
				distance := math.Sqrt(float64(dx*dx + dy*dy))

				// High-frequency noise for tiny organic variations
				noiseValue := pg.simpleNoise(float64(x)*0.8, float64(y)*0.8)
				organicRadius := float64(size) + noiseValue*0.8

				if distance <= organicRadius {
					// High probability for small clusters
					probability := 1.0 - (distance / organicRadius)
					if rand.Float64() < probability*0.9 {
						grid[y][x] = colorIndex
					}
				}
			}
		}
	}
}
