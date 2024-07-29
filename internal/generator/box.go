package generator

import (
	"context"
	"image"
	"image/color"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

type BoxGenerator struct{}

func (bg *BoxGenerator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Shuffle the colors
	shuffledColors := shuffleColors(colors)

	// Adjust base pixel size to fit perfectly within the dimensions
	adjustedBasePixelSize := cfg.BasePixelSize
	for cfg.Width%adjustedBasePixelSize != 0 || cfg.Height%adjustedBasePixelSize != 0 {
		adjustedBasePixelSize--
	}

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
			grid[y][x] = rand.Intn(len(shuffledColors))
		}
	}

	// Apply cellular automaton rules to create larger clusters
	for i := 0; i < 3; i++ {
		newGrid := make([][]int, cellHeight)
		for y := range newGrid {
			newGrid[y] = make([]int, cellWidth)
			copy(newGrid[y], grid[y])
		}

		for y := 0; y < cellHeight; y++ {
			for x := 0; x < cellWidth; x++ {
				// Count neighboring colors with variable neighborhood size
				neighborhoodSize := rand.Intn(3) + 1 // 1, 2, or 3
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
				color := shuffledColors[grid[cellY][cellX]]
				img.Set(x, y, color)
			}
		}
	}

	if cfg.AddNoise {
		addNoiseNRGBA(img, shuffledColors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, adjustedBasePixelSize)
	}

	return img, nil
}
