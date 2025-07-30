package generator

import (
	"context"
	"image"
	"image/color"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

// Pat1Generator creates pat1 (military-inspired) camouflage patterns.
// Mimics proven military camouflage design principles rather than novel algorithms.
type Pat1Generator struct{}

// Generate creates woodland camouflage inspired by successful military patterns like MARPAT and MultiCam.
// Uses the same technical architecture as existing box/blob patterns.
func (pg *Pat1Generator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Use the centralized pixel size adjustment for perfect fit
	adjustedBasePixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	// Calculate the number of cells based on the image dimensions and adjusted base pixel size
	cellWidth := cfg.Width / adjustedBasePixelSize
	cellHeight := cfg.Height / adjustedBasePixelSize

	// Create a grid to store color indices (same structure as box/blob)
	grid := make([][]int, cellHeight)
	for i := range grid {
		grid[i] = make([]int, cellWidth)
	}

	// Initialize with random background color (any color can be background)
	backgroundColorIndex := rand.Intn(len(colors))
	for y := 0; y < cellHeight; y++ {
		for x := 0; x < cellWidth; x++ {
			grid[y][x] = backgroundColorIndex
		}
	}

	// Apply pat1-style layered approach with weighted color distribution
	// Layer 1: Large background shapes (like MultiCam base pattern)
	pg.addLargeShapes(grid, cellWidth, cellHeight, colors, backgroundColorIndex, cfg)

	// Layer 2: Medium woodland elements (tree-like shapes)
	pg.addWoodlandElements(grid, cellWidth, cellHeight, colors, cfg)

	// Layer 3: Small digital elements (pat1-style pixels)
	pg.addDigitalDetails(grid, cellWidth, cellHeight, colors, cfg)

	// Apply military-style smoothing (less aggressive than blob, more than box)
	pg.applyMilitarySmoothing(grid, cellWidth, cellHeight, len(colors))

	// Draw the pattern (same rendering approach as box/blob)
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			cellY := y / adjustedBasePixelSize
			cellX := x / adjustedBasePixelSize
			if cellY < cellHeight && cellX < cellWidth {
				img.Set(x, y, colors[grid[cellY][cellX]])
			}
		}
	}

	// Apply noise and edge effects if requested (same as existing patterns)
	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, adjustedBasePixelSize)
	}

	return img, nil
}

// findBackgroundColor selects the lightest color as background (typical in military camo)
func (pg *Pat1Generator) findBackgroundColor(colors []color.RGBA) int {
	maxBrightness := 0
	backgroundIndex := 0

	for i, c := range colors {
		brightness := int(c.R) + int(c.G) + int(c.B)
		if brightness > maxBrightness {
			maxBrightness = brightness
			backgroundIndex = i
		}
	}

	return backgroundIndex
}

// addLargeShapes creates MultiCam-style large background shapes with weighted color distribution
func (pg *Pat1Generator) addLargeShapes(grid [][]int, cellWidth, cellHeight int, colors []color.RGBA, backgroundIndex int, cfg *config.Config) {
	numShapes := (cellWidth * cellHeight) / 200 // Fewer, larger shapes like MultiCam

	for i := 0; i < numShapes; i++ {
		centerX := rand.Intn(cellWidth)
		centerY := rand.Intn(cellHeight)

		// Use random color selection for large shapes (excluding background)
		colorIndex := selectWeightedColorExcluding(cfg.ColorRatios, []int{backgroundIndex}, len(colors))

		// Create irregular organic shapes (not perfect circles)
		sizeX := 8 + rand.Intn(12) // 8-20 cells wide
		sizeY := 6 + rand.Intn(10) // 6-16 cells high

		pg.drawOrganicShape(grid, centerX, centerY, sizeX, sizeY, colorIndex, cellWidth, cellHeight)
	}
}

// addWoodlandElements creates tree and foliage-like patterns with weighted colors
func (pg *Pat1Generator) addWoodlandElements(grid [][]int, cellWidth, cellHeight int, colors []color.RGBA, cfg *config.Config) {
	numElements := (cellWidth * cellHeight) / 100

	for i := 0; i < numElements; i++ {
		x := rand.Intn(cellWidth)
		y := rand.Intn(cellHeight)

		// Use random color selection for vegetation elements
		colorIndex := selectWeightedColor(cfg.ColorRatios, len(colors))

		elementType := rand.Intn(3)
		switch elementType {
		case 0: // Vertical tree-like elements
			pg.drawVerticalElement(grid, x, y, colorIndex, cellWidth, cellHeight)
		case 1: // Branch-like horizontal elements
			pg.drawHorizontalElement(grid, x, y, colorIndex, cellWidth, cellHeight)
		case 2: // Leaf cluster-like elements
			pg.drawClusterElement(grid, x, y, colorIndex, cellWidth, cellHeight)
		}
	}
}

// addDigitalDetails creates pat1-style small digital elements with weighted colors
func (pg *Pat1Generator) addDigitalDetails(grid [][]int, cellWidth, cellHeight int, colors []color.RGBA, cfg *config.Config) {
	numPixels := (cellWidth * cellHeight) / 50 // Dense digital texture

	for i := 0; i < numPixels; i++ {
		x := rand.Intn(cellWidth - 1)
		y := rand.Intn(cellHeight - 1)

		// Use random color selection for digital details
		colorIndex := selectWeightedColor(cfg.ColorRatios, len(colors))

		// pat1-style rectangular digital elements
		width := 1 + rand.Intn(2)  // 1-2 cells wide
		height := 1 + rand.Intn(2) // 1-2 cells high

		for dy := 0; dy < height && y+dy < cellHeight; dy++ {
			for dx := 0; dx < width && x+dx < cellWidth; dx++ {
				grid[y+dy][x+dx] = colorIndex
			}
		}
	}
}

// applyMilitarySmoothing applies moderate smoothing typical of military patterns
func (pg *Pat1Generator) applyMilitarySmoothing(grid [][]int, cellWidth, cellHeight, numColors int) {
	// Two passes of light smoothing (between box and blob intensity)
	for pass := 0; pass < 2; pass++ {
		newGrid := make([][]int, cellHeight)
		for y := range newGrid {
			newGrid[y] = make([]int, cellWidth)
			copy(newGrid[y], grid[y])
		}

		for y := 1; y < cellHeight-1; y++ {
			for x := 1; x < cellWidth-1; x++ {
				// Count neighboring colors
				colorCount := make(map[int]int)
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						colorCount[grid[y+dy][x+dx]]++
					}
				}

				// Find the most common color
				maxCount, maxColor := 0, grid[y][x]
				for color, count := range colorCount {
					if count > maxCount || (count == maxCount && rand.Float32() < 0.3) {
						maxCount, maxColor = count, color
					}
				}

				// Apply moderate smoothing (50% probability)
				if rand.Float32() < 0.5 {
					newGrid[y][x] = maxColor
				}
			}
		}

		// Copy smoothed result back
		for y := range grid {
			copy(grid[y], newGrid[y])
		}
	}
}

// Helper functions for distributed military color selection

// selectDistributedColor ensures all colors get used prominently like real SPLAT
func (pg *Pat1Generator) selectDistributedColor(colors []color.RGBA, excludeIndex int, seed int) int {
	// Create list of available colors (excluding background if specified)
	availableColors := make([]int, 0, len(colors))
	for i := range colors {
		if i != excludeIndex {
			availableColors = append(availableColors, i)
		}
	}

	if len(availableColors) == 0 {
		return 0 // Fallback
	}

	// Distribute colors evenly using seed to ensure good distribution
	return availableColors[seed%len(availableColors)]
}

// Drawing functions for woodland elements

func (pg *Pat1Generator) drawOrganicShape(grid [][]int, centerX, centerY, sizeX, sizeY, colorIndex, cellWidth, cellHeight int) {
	for dy := -sizeY; dy <= sizeY; dy++ {
		for dx := -sizeX; dx <= sizeX; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < cellWidth && y >= 0 && y < cellHeight {
				// Create organic boundary using distance with noise
				distX := float32(dx) / float32(sizeX)
				distY := float32(dy) / float32(sizeY)
				distance := distX*distX + distY*distY

				// Add organic variation
				noise := rand.Float32()*0.3 - 0.15
				threshold := 1.0 + noise

				if distance <= threshold {
					probability := 1.0 - distance/threshold
					if rand.Float32() < probability {
						grid[y][x] = colorIndex
					}
				}
			}
		}
	}
}

func (pg *Pat1Generator) drawVerticalElement(grid [][]int, startX, startY, colorIndex, cellWidth, cellHeight int) {
	height := 4 + rand.Intn(8) // 4-12 cells high
	width := 1 + rand.Intn(2)  // 1-2 cells wide

	for y := startY; y < startY+height && y < cellHeight; y++ {
		for x := startX; x < startX+width && x < cellWidth; x++ {
			if x >= 0 && y >= 0 {
				grid[y][x] = colorIndex
			}
		}
	}
}

func (pg *Pat1Generator) drawHorizontalElement(grid [][]int, startX, startY, colorIndex, cellWidth, cellHeight int) {
	width := 4 + rand.Intn(8)  // 4-12 cells wide
	height := 1 + rand.Intn(2) // 1-2 cells high

	for y := startY; y < startY+height && y < cellHeight; y++ {
		for x := startX; x < startX+width && x < cellWidth; x++ {
			if x >= 0 && y >= 0 {
				grid[y][x] = colorIndex
			}
		}
	}
}

func (pg *Pat1Generator) drawClusterElement(grid [][]int, centerX, centerY, colorIndex, cellWidth, cellHeight int) {
	size := 2 + rand.Intn(3) // 2-4 cell radius

	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < cellWidth && y >= 0 && y < cellHeight {
				distance := dx*dx + dy*dy
				if distance <= size*size && rand.Float32() < 0.7 {
					grid[y][x] = colorIndex
				}
			}
		}
	}
}
