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
	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	// Start with an initial square size
	squareSize := cfg.BasePixelSize * 4

	// Find the largest square size that fits perfectly
	adjustedSquareSize := findLargestDivisor(cfg.Width, cfg.Height, squareSize)

	// Calculate number of squares
	numCols := cfg.Width / adjustedSquareSize
	numRows := cfg.Height / adjustedSquareSize

	// Iterate over grid positions to fill squares
	for row := 0; row < numRows; row++ {
		for col := 0; col < numCols; col++ {
			// Calculate the top-left corner of the square
			startX := col * adjustedSquareSize
			startY := row * adjustedSquareSize

			// Pick a random color
			color := colors[rand.Intn(len(colors))]

			// Draw the square
			drawRectangle(img, startX, startY, adjustedSquareSize, adjustedSquareSize, color)
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

func findLargestDivisor(width, height, maxSize int) int {
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	commonDivisor := gcd(width, height)
	for size := maxSize; size >= 1; size-- {
		if commonDivisor%size == 0 && width%size == 0 && height%size == 0 {
			return size
		}
	}
	return 1
}

func drawRectangle(img *image.NRGBA, x, y, width, height int, c color.RGBA) {
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			img.Set(i, j, c)
		}
	}
}
