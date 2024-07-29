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
			grid[y][x] = rand.Intn(len(colors))
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
				// Count neighboring colors
				colorCount := make(map[int]int)
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						ny, nx := (y+dy+cellHeight)%cellHeight, (x+dx+cellWidth)%cellWidth
						colorCount[grid[ny][nx]]++
					}
				}

				// Find the most common color
				maxCount, maxColor := 0, grid[y][x]
				for color, count := range colorCount {
					if count > maxCount {
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

	// Create larger squares
	maxSquareSize := 4 // Maximum size of larger squares
	for y := 0; y < cellHeight; y += maxSquareSize {
		for x := 0; x < cellWidth; x += maxSquareSize {
			if rand.Float32() < 0.3 { // 30% chance to create a larger square
				color := grid[y][x]
				size := rand.Intn(maxSquareSize) + 1 // Random size between 1 and maxSquareSize
				for dy := 0; dy < size && y+dy < cellHeight; dy++ {
					for dx := 0; dx < size && x+dx < cellWidth; dx++ {
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
				color := colors[grid[cellY][cellX]]
				img.Set(x, y, color)
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
