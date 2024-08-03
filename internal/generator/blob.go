package generator

import (
	"context"
	"image"
	"image/color"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

type BlobGenerator struct{}

func (bg *BlobGenerator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Shuffle the colors
	shuffledColors := shuffleColors(colors)

	// Adjust base pixel size to fit perfectly within the dimensions
	adjustedBasePixelSize := cfg.BasePixelSize
	for cfg.Width%adjustedBasePixelSize != 0 || cfg.Height%adjustedBasePixelSize != 0 {
		adjustedBasePixelSize--
	}

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	// Adjust the scale factor to create more varied blob sizes
	scaleFactor := rand.Intn(3) + 1 // 1, 2, or 3

	// Create the pattern grid with variable cell sizes
	patternWidth, patternHeight := cfg.Width/(adjustedBasePixelSize*scaleFactor), cfg.Height/(adjustedBasePixelSize*scaleFactor)
	pattern := make([][]int, patternHeight)
	for y := range pattern {
		pattern[y] = make([]int, patternWidth)
		for x := range pattern[y] {
			pattern[y][x] = rand.Intn(len(shuffledColors))
		}
	}

	// Apply cellular automata to create clustered blob regions
	iterations := rand.Intn(3) + 2 // 2 to 4 iterations
	for i := 0; i < iterations; i++ {
		newPattern := make([][]int, patternHeight)
		for y := range newPattern {
			newPattern[y] = make([]int, patternWidth)
			for x := range newPattern[y] {
				colorCounts := make(map[int]int)
				neighborhoodSize := rand.Intn(2) + 1 // 1 or 2
				for dy := -neighborhoodSize; dy <= neighborhoodSize; dy++ {
					for dx := -neighborhoodSize; dx <= neighborhoodSize; dx++ {
						ny, nx := (y+dy+patternHeight)%patternHeight, (x+dx+patternWidth)%patternWidth
						colorCounts[pattern[ny][nx]]++
					}
				}
				maxCount, dominantColor := 0, pattern[y][x]
				for color, count := range colorCounts {
					if count > maxCount || (count == maxCount && rand.Float32() < 0.3) {
						maxCount, dominantColor = count, color
					}
				}
				newPattern[y][x] = dominantColor
			}
		}
		pattern = newPattern
	}

	// Introduce larger blob areas
	for i := 0; i < patternHeight*patternWidth/20; i++ { // Create several larger blobs
		y, x := rand.Intn(patternHeight), rand.Intn(patternWidth)
		color := rand.Intn(len(shuffledColors))
		blobSize := rand.Intn(patternHeight/4) + patternHeight/8 // Varied blob sizes
		for dy := -blobSize; dy <= blobSize; dy++ {
			for dx := -blobSize; dx <= blobSize; dx++ {
				if dx*dx+dy*dy <= blobSize*blobSize { // Circular blob shape
					ny, nx := (y+dy+patternHeight)%patternHeight, (x+dx+patternWidth)%patternWidth
					if rand.Float32() < 0.7 { // 70% chance to set the color, creating more organic edges
						pattern[ny][nx] = color
					}
				}
			}
		}
	}

	// Draw the pattern
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			patternY := (y / (adjustedBasePixelSize * scaleFactor)) % patternHeight
			patternX := (x / (adjustedBasePixelSize * scaleFactor)) % patternWidth
			colorIndex := pattern[patternY][patternX]
			c := shuffledColors[colorIndex]
			img.Set(x, y, c)
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
