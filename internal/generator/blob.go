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

	// Adjust the scale factor to create smaller blobs
	scaleFactor := 1

	// Create the pattern grid with smaller cells
	patternWidth, patternHeight := cfg.Width/(adjustedBasePixelSize*scaleFactor), cfg.Height/(adjustedBasePixelSize*scaleFactor)
	pattern := make([][]int, patternHeight)
	for y := range pattern {
		pattern[y] = make([]int, patternWidth)
		for x := range pattern[y] {
			pattern[y][x] = rand.Intn(len(shuffledColors))
		}
	}

	// Apply cellular automata to create clustered blob regions
	iterations := 3
	for i := 0; i < iterations; i++ {
		newPattern := make([][]int, patternHeight)
		for y := range newPattern {
			newPattern[y] = make([]int, patternWidth)
			for x := range newPattern[y] {
				colorCounts := make(map[int]int)
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
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
