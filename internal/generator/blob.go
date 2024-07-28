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
	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	// Create the pattern grid
	patternWidth, patternHeight := cfg.Width/cfg.BasePixelSize, cfg.Height/cfg.BasePixelSize
	pattern := make([][]int, patternHeight)
	for y := range pattern {
		pattern[y] = make([]int, patternWidth)
		for x := range pattern[y] {
			pattern[y][x] = rand.Intn(len(colors))
		}
	}

	// Apply cellular automata to create clustered blob regions
	for i := 0; i < 5; i++ {
		newPattern := make([][]int, patternHeight)
		for y := range newPattern {
			newPattern[y] = make([]int, patternWidth)
			for x := range newPattern[y] {
				colorCounts := make(map[int]int)
				for dy := -2; dy <= 2; dy++ {
					for dx := -3; dx <= 3; dx++ {
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
			patternY, patternX := y/cfg.BasePixelSize, x/cfg.BasePixelSize
			colorIndex := pattern[patternY][patternX]
			c := colors[colorIndex]
			img.Set(x, y, c)
		}
	}

	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, cfg.BasePixelSize)
	}

	return img, nil
}
