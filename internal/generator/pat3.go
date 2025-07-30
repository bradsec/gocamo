package generator

import (
	"context"
	"image"
	"image/color"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

// Pat3Generator creates pat3 (geometric box-style) camouflage patterns with squares and rectangles.
// It generates angular patterns with clustered regions using cellular automata and geometric shapes.
type Pat3Generator struct{}

// Generate creates a pat3-style camouflage pattern with geometric shapes, squares, and rectangles.
// It uses cellular automata to create clustered regions and adds larger geometric shapes for variation.
func (pg *Pat3Generator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Use the centralized pixel size adjustment for perfect fit
	adjustedBasePixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	// Calculate the number of cells based on the image dimensions and adjusted base pixel size
	cellWidth := cfg.Width / adjustedBasePixelSize
	cellHeight := cfg.Height / adjustedBasePixelSize

	// Create a grid to store color indices
	grid := make([][]int, cellHeight)
	for i := range grid {
		grid[i] = make([]int, cellWidth)
	}

	// Generate initial random color assignment
	for y := 0; y < cellHeight; y++ {
		for x := 0; x < cellWidth; x++ {
			grid[y][x] = selectWeightedColor(cfg.ColorRatios, len(colors))
		}
	}

	// Apply cellular automaton rules to create clusters
	for i := 0; i < 3; i++ {
		newGrid := make([][]int, cellHeight)
		for y := range newGrid {
			newGrid[y] = make([]int, cellWidth)
			copy(newGrid[y], grid[y])
		}

		for y := 0; y < cellHeight; y++ {
			for x := 0; x < cellWidth; x++ {
				// Count neighboring colors with variable neighborhood size
				neighborhoodSize := rand.Intn(2) + 1 // 1 or 2
				colorCount := make(map[int]int)
				for dy := -neighborhoodSize; dy <= neighborhoodSize; dy++ {
					for dx := -neighborhoodSize; dx <= neighborhoodSize; dx++ {
						ny, nx := (y+dy+cellHeight)%cellHeight, (x+dx+cellWidth)%cellWidth
						colorCount[grid[ny][nx]]++
					}
				}

				// Find the most common color
				maxCount, maxColor := 0, grid[y][x]
				for color, count := range colorCount {
					if count > maxCount || (count == maxCount && rand.Float32() < 0.3) {
						maxCount, maxColor = count, color
					}
				}

				// Apply the most common color with a probability
				if rand.Float32() < 0.7 {
					newGrid[y][x] = maxColor
				}
			}
		}

		grid = newGrid
	}

	// Create larger squares and rectangles
	maxSize := 8 // Maximum size of larger shapes
	for y := 0; y < cellHeight; y += maxSize / 2 {
		for x := 0; x < cellWidth; x += maxSize / 2 {
			if rand.Float32() < 0.3 { // 30% chance to create a larger shape
				shapeType := rand.Intn(3) // 0: square, 1: horizontal rectangle, 2: vertical rectangle
				width := rand.Intn(maxSize) + 1
				height := rand.Intn(maxSize) + 1

				if shapeType == 1 {
					width = rand.Intn(maxSize) + maxSize/2 // Wider
					height = rand.Intn(maxSize/2) + 1      // Shorter
				} else if shapeType == 2 {
					width = rand.Intn(maxSize/2) + 1        // Narrower
					height = rand.Intn(maxSize) + maxSize/2 // Taller
				}

				color := grid[y][x]
				for dy := 0; dy < height && y+dy < cellHeight; dy++ {
					for dx := 0; dx < width && x+dx < cellWidth; dx++ {
						grid[y+dy][x+dx] = color
					}
				}
			}
		}
	}

	// Draw the pattern
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			cellY := y / adjustedBasePixelSize
			cellX := x / adjustedBasePixelSize
			if cellY < cellHeight && cellX < cellWidth {
				img.Set(x, y, colors[grid[cellY][cellX]])
			}
		}
	}

	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, adjustedBasePixelSize)
	}

	return img, nil
}
