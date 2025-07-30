package generator

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/bradsec/gocamo/internal/utils"
	"github.com/bradsec/gocamo/pkg/config"
	"golang.org/x/image/draw"
)

// ImageGenerator creates camouflage patterns by analyzing input images to extract dominant colors.
// It uses k-means clustering, image preprocessing, and edge detection to generate patterns based on real images.
type ImageGenerator struct {
	InputFile string
}

// Generate creates a camouflage pattern by analyzing an input image to extract dominant colors using k-means clustering.
// It applies image preprocessing, max pooling, and Laplacian filtering before generating the final pattern.
func (ig *ImageGenerator) Generate(ctx context.Context, cfg *config.Config, _ []color.RGBA) (image.Image, []color.RGBA, error) {
	// Use the centralized pixel size adjustment for perfect fit
	adjustedBasePixelSize := cfg.AdjustBasePixelSize()

	inputImg, err := utils.LoadImage(ig.InputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading image: %w", err)
	}
	resized := resizeAndCropImage(inputImg, cfg.Width, cfg.Height)
	pooled := maxPooling(resized, adjustedBasePixelSize)
	enhanced := laplacianFilter(pooled)
	bounds := enhanced.Bounds()
	pixels := make([]color.Color, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixels = append(pixels, enhanced.At(x, y))
		}
	}
	mainColors := kMeansClustering(pixels, cfg.KValue, 100)
	result := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			enhancedX := x * bounds.Dx() / cfg.Width
			enhancedY := y * bounds.Dy() / cfg.Height
			pixel := enhanced.At(enhancedX, enhancedY)

			// Find closest color with soft blending for edge transitions
			closestColor := mainColors[0]
			minDistance := colorDistance(pixel, closestColor)
			secondClosest := mainColors[0]
			secondDistance := minDistance

			for _, candidateColor := range mainColors[1:] {
				d := colorDistance(pixel, candidateColor)
				if d < minDistance {
					secondClosest = closestColor
					secondDistance = minDistance
					closestColor = candidateColor
					minDistance = d
				} else if d < secondDistance {
					secondClosest = candidateColor
					secondDistance = d
				}
			}

			// Soft blending when distances are close to reduce harsh edges
			if secondDistance-minDistance < 30 && minDistance > 10 {
				r1, g1, b1, _ := closestColor.RGBA()
				r2, g2, b2, _ := secondClosest.RGBA()

				// Blend based on distance ratio
				blend := minDistance / (minDistance + secondDistance)
				finalR := uint8((float64(r1>>8)*(1-blend) + float64(r2>>8)*blend))
				finalG := uint8((float64(g1>>8)*(1-blend) + float64(g2>>8)*blend))
				finalB := uint8((float64(b1>>8)*(1-blend) + float64(b2>>8)*blend))

				result.Set(x, y, color.RGBA{R: finalR, G: finalG, B: finalB, A: 255})
			} else {
				result.Set(x, y, closestColor)
			}
		}
	}

	if cfg.AddNoise {
		addNoiseRGBA(result, mainColors)
	}

	if cfg.AddEdge {
		addEdgeDetailsRGBA(result, adjustedBasePixelSize)
	}
	return result, mainColors, nil
}

func maxPooling(img image.Image, poolSize int) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	newWidth, newHeight := (width+poolSize-1)/poolSize, (height+poolSize-1)/poolSize

	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			var maxR, maxG, maxB uint32
			maxR, maxG, maxB = 0, 0, 0

			for py := 0; py < poolSize && y*poolSize+py < height; py++ {
				for px := 0; px < poolSize && x*poolSize+px < width; px++ {
					r, g, b, _ := img.At(x*poolSize+px, y*poolSize+py).RGBA()
					maxR = maxU(maxR, r)
					maxG = maxU(maxG, g)
					maxB = maxU(maxB, b)
				}
			}

			result.Set(x, y, color.RGBA{
				R: uint8(maxR >> 8),
				G: uint8(maxG >> 8),
				B: uint8(maxB >> 8),
				A: 255,
			})
		}
	}

	return result
}

func laplacianFilter(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	result := image.NewRGBA(bounds)

	// Gentler edge detection kernel to reduce harsh borders
	kernel := [][]float64{
		{0, -0.25, 0},
		{-0.25, 2, -0.25},
		{0, -0.25, 0},
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var sumR, sumG, sumB float64

			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					nx, ny := x+kx, y+ky
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						r, g, b, _ := img.At(nx, ny).RGBA()
						k := kernel[ky+1][kx+1]
						sumR += float64(r>>8) * k
						sumG += float64(g>>8) * k
						sumB += float64(b>>8) * k
					}
				}
			}

			// Reduce the edge enhancement effect and blend with original
			r, g, b, _ := img.At(x, y).RGBA()
			origR, origG, origB := float64(r>>8), float64(g>>8), float64(b>>8)

			// Mix 70% original + 30% edge-enhanced for subtle effect
			mixR := origR*0.7 + (origR+sumR*0.3)*0.3
			mixG := origG*0.7 + (origG+sumG*0.3)*0.3
			mixB := origB*0.7 + (origB+sumB*0.3)*0.3

			result.Set(x, y, color.RGBA{
				R: uint8(clampLap(int32(mixR))),
				G: uint8(clampLap(int32(mixG))),
				B: uint8(clampLap(int32(mixB))),
				A: 255,
			})
		}
	}

	return result
}

func maxU(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func clampLap(v int32) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func kMeansClustering(pixels []color.Color, k int, maxIterations int) []color.RGBA {
	// Convert pixels to a slice of [3]float64 for easier computation
	points := make([][3]float64, len(pixels))
	for i, p := range pixels {
		r, g, b, _ := p.RGBA()
		points[i] = [3]float64{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
	}

	// Initialize centroids randomly
	centroids := make([][3]float64, k)
	for i := range centroids {
		centroids[i] = points[rand.Intn(len(points))]
	}

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Assign points to clusters
		clusters := make([][][3]float64, k)
		for _, point := range points {
			closestCentroid := 0
			minDistance := distance(point, centroids[0])
			for j := 1; j < k; j++ {
				d := distance(point, centroids[j])
				if d < minDistance {
					minDistance = d
					closestCentroid = j
				}
			}
			clusters[closestCentroid] = append(clusters[closestCentroid], point)
		}

		// Update centroids
		for i, cluster := range clusters {
			if len(cluster) == 0 {
				continue
			}
			var sumR, sumG, sumB float64
			for _, point := range cluster {
				sumR += point[0]
				sumG += point[1]
				sumB += point[2]
			}
			centroids[i] = [3]float64{
				sumR / float64(len(cluster)),
				sumG / float64(len(cluster)),
				sumB / float64(len(cluster)),
			}
		}
	}

	// Convert centroids to color.RGBA
	result := make([]color.RGBA, k)
	for i, centroid := range centroids {
		result[i] = color.RGBA{
			R: uint8(centroid[0]),
			G: uint8(centroid[1]),
			B: uint8(centroid[2]),
			A: 255,
		}
	}
	return result
}

func distance(a, b [3]float64) float64 {
	return math.Sqrt(math.Pow(a[0]-b[0], 2) + math.Pow(a[1]-b[1], 2) + math.Pow(a[2]-b[2], 2))
}

func colorDistance(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return math.Sqrt(math.Pow(float64(r1>>8)-float64(r2>>8), 2) +
		math.Pow(float64(g1>>8)-float64(g2>>8), 2) +
		math.Pow(float64(b1>>8)-float64(b2>>8), 2))
}

// BilinearScale performs bilinear interpolation to resize an image to the specified dimensions.
// It uses bilinear interpolation to calculate pixel values by considering the four nearest
// pixels in the source image and computing a weighted average based on the fractional
// position. This provides smooth scaling with better quality than nearest-neighbor interpolation.
//
// The function maps each pixel in the destination image to a corresponding position in the
// source image, then interpolates the color values from the surrounding pixels to create
// a smooth transition.
//
// Parameters:
//   - src: The source image to be resized
//   - dstWidth: Target width for the resized image
//   - dstHeight: Target height for the resized image
//
// Returns a new RGBA image with the specified dimensions containing the interpolated pixel data.
func BilinearScale(src image.Image, dstWidth, dstHeight int) *image.RGBA {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))

	for dy := 0; dy < dstHeight; dy++ {
		for dx := 0; dx < dstWidth; dx++ {
			// Calculate the corresponding position in the source image
			sx := (float64(dx) / float64(dstWidth)) * float64(srcWidth)
			sy := (float64(dy) / float64(dstHeight)) * float64(srcHeight)

			// Determine the four nearest pixels
			x0 := int(math.Floor(sx))
			y0 := int(math.Floor(sy))
			x1 := int(math.Ceil(sx))
			y1 := int(math.Ceil(sy))

			// Clamp coordinates to ensure they are within bounds
			if x1 >= srcWidth {
				x1 = srcWidth - 1
			}
			if y1 >= srcHeight {
				y1 = srcHeight - 1
			}

			// Calculate interpolation weights
			tx := sx - float64(x0)
			ty := sy - float64(y0)

			// Get the colors of the four surrounding pixels
			c00 := src.At(x0, y0)
			c01 := src.At(x0, y1)
			c10 := src.At(x1, y0)
			c11 := src.At(x1, y1)

			// Compute the interpolated color
			r00, g00, b00, a00 := c00.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			r := uint8((float64(r00)*(1-tx)*(1-ty) +
				float64(r01)*(1-tx)*ty +
				float64(r10)*tx*(1-ty) +
				float64(r11)*tx*ty) / 256)
			g := uint8((float64(g00)*(1-tx)*(1-ty) +
				float64(g01)*(1-tx)*ty +
				float64(g10)*tx*(1-ty) +
				float64(g11)*tx*ty) / 256)
			b := uint8((float64(b00)*(1-tx)*(1-ty) +
				float64(b01)*(1-tx)*ty +
				float64(b10)*tx*(1-ty) +
				float64(b11)*tx*ty) / 256)
			a := uint8((float64(a00)*(1-tx)*(1-ty) +
				float64(a01)*(1-tx)*ty +
				float64(a10)*tx*(1-ty) +
				float64(a11)*tx*ty) / 256)

			dst.Set(dx, dy, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return dst
}

// resizeAndCropImage uses BilinearScale to resize the image and then crops it
func resizeAndCropImage(img image.Image, targetWidth, targetHeight int) image.Image {
	const smallestSide = 256

	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	var scaleFactor float64
	if srcWidth < srcHeight {
		scaleFactor = float64(smallestSide) / float64(srcWidth)
	} else {
		scaleFactor = float64(smallestSide) / float64(srcHeight)
	}

	newWidth := int(float64(srcWidth) * scaleFactor)
	newHeight := int(float64(srcHeight) * scaleFactor)

	// Scale down the image
	scaledDown := BilinearScale(img, newWidth, newHeight)

	widthRatio := float64(targetWidth) / float64(newWidth)
	heightRatio := float64(targetHeight) / float64(newHeight)

	// Use the larger ratio to ensure the image fills the target dimensions
	ratio := math.Max(widthRatio, heightRatio)

	// Calculate dimensions for the resized image
	resizeWidth := int(float64(newWidth) * ratio)
	resizeHeight := int(float64(newHeight) * ratio)

	// Ensure the resized dimensions are at least as large as the target dimensions
	if resizeWidth < targetWidth {
		resizeWidth = targetWidth
	}
	if resizeHeight < targetHeight {
		resizeHeight = targetHeight
	}

	// Resize the scaled-down image to fill the target dimensions
	resized := BilinearScale(scaledDown, resizeWidth, resizeHeight)

	// Calculate cropping bounds
	cropX := (resizeWidth - targetWidth) / 2
	cropY := (resizeHeight - targetHeight) / 2

	// Ensure cropping coordinates are within bounds
	if cropX < 0 {
		cropX = 0
	}
	if cropY < 0 {
		cropY = 0
	}

	cropRect := image.Rect(cropX, cropY, cropX+targetWidth, cropY+targetHeight)

	result := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.Draw(result, result.Bounds(), resized, cropRect.Min, draw.Src)

	return result
}
