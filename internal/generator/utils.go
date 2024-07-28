package generator

import (
	"image"
	"image/color"
	"math/rand"
)

func addNoiseRGBA(img *image.RGBA, colors []color.RGBA) {
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
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x%basePixelSize == 0 || y%basePixelSize == 0 {
				if rand.Float32() < 0.4 { // 40% chance for edge details
					currentColor := img.RGBAAt(x, y)
					r := uint8(clamp(int(currentColor.R)+rand.Intn(41)-20, 0, 255))
					g := uint8(clamp(int(currentColor.G)+rand.Intn(41)-20, 0, 255))
					b := uint8(clamp(int(currentColor.B)+rand.Intn(41)-20, 0, 255))
					img.Set(x, y, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

func addEdgeDetailsNRGBA(img *image.NRGBA, basePixelSize int) {
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
