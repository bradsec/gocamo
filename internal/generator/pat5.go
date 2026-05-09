package generator

import (
	"context"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

// Pat5Generator creates pat5 digital camouflage patterns.
// Uses 45/30/15/10 colour ratios and a Voronoi-seed approach
// to produce consistent rectangular digital-pixel clusters at the correct scale.
type Pat5Generator struct{}

// seed is a Voronoi seed point with a colour index.
type seed struct {
	x, y  int
	color int
}

// Generate creates a pat5 digital camouflage pattern.
// Pipeline: Voronoi seeding → directional CA for rectangular shaping → pixel blocks → ratio enforcement.
func (pg *Pat5Generator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	adjustedBasePixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	gridWidth := cfg.Width / adjustedBasePixelSize
	gridHeight := cfg.Height / adjustedBasePixelSize

	// Use pat5 default ratios when the user has not specified explicit ratios.
	ratios := cfg.ColorRatios
	if cfg.RatiosString == "" {
		ratios = pg.getPat5DefaultColorRatios(len(colors))
	}

	// Layer 1: Voronoi seeding — scatter weighted seeds and grow Voronoi regions.
	grid := pg.generateVoronoiGrid(gridWidth, gridHeight, ratios, len(colors))

	// Layer 2: Directional CA — reshape diagonal Voronoi boundaries into axis-aligned
	// rectangles characteristic of digital camouflage pixels.
	pg.applyRectangularCA(grid, gridWidth, gridHeight, 3, 1, 2) // horizontal
	pg.applyRectangularCA(grid, gridWidth, gridHeight, 1, 3, 2) // vertical

	// Layer 3: Small digital pixel blocks for fine texture.
	pg.applyDigitalPixelBlocks(grid, gridWidth, gridHeight)

	// Layer 4: Restore target colour ratios that CA majority-voting may have eroded.
	// Places proper-sized rectangular patches so corrections match the pat5 style.
	pg.enforceColorRatios(grid, gridWidth, gridHeight, ratios, len(colors))

	pg.renderGrid(img, grid, colors, adjustedBasePixelSize)

	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}
	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, adjustedBasePixelSize)
	}

	return img, nil
}

// generateVoronoiGrid places weighted seed points using grid-jitter sampling
// (one seed per regular jitter cell, with random offset within the cell). This guarantees
// a minimum seed spacing of ~jitterSize cells, preventing tiny isolated Voronoi regions
// that would appear as single-pixel dots in the output.
func (pg *Pat5Generator) generateVoronoiGrid(width, height int, ratios []float64, numColors int) [][]int {
	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, width)
	}

	// Keep seed spacing in grid cells so BasePixelSize scales pat5 the same way
	// it scales the other pattern generators.
	jitterSize := 4
	if width < jitterSize || height < jitterSize {
		jitterSize = width
		if height < jitterSize {
			jitterSize = height
		}
	}

	var seeds []seed
	cols := (width + jitterSize - 1) / jitterSize
	rows := (height + jitterSize - 1) / jitterSize

	for gy := 0; gy < rows; gy++ {
		for gx := 0; gx < cols; gx++ {
			x0 := gx * jitterSize
			y0 := gy * jitterSize
			x1 := x0 + jitterSize
			y1 := y0 + jitterSize
			if x1 > width {
				x1 = width
			}
			if y1 > height {
				y1 = height
			}

			sx := x0 + rand.Intn(x1-x0)
			sy := y0 + rand.Intn(y1-y0)
			seeds = append(seeds, seed{
				x:     sx,
				y:     sy,
				color: selectWeightedColor(ratios, numColors),
			})
		}
	}

	// Assign each cell the nearest seed's colour (Euclidean distance).
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			minDist := math.MaxFloat64
			best := 0
			for _, s := range seeds {
				dx := float64(x - s.x)
				dy := float64(y - s.y)
				d := dx*dx + dy*dy
				if d < minDist {
					minDist = d
					best = s.color
				}
			}
			grid[y][x] = best
		}
	}

	return grid
}

// applyRectangularCA performs cellular automata with a rectangular neighbourhood.
// Directional passes (3×1 horizontal, 1×3 vertical) snap Voronoi boundaries to
// axis-aligned edges, giving the pattern its characteristic rectangular pixel look.
func (pg *Pat5Generator) applyRectangularCA(grid [][]int, width, height, winW, winH, iterations int) {
	halfW := winW / 2
	halfH := winH / 2

	for iter := 0; iter < iterations; iter++ {
		newGrid := make([][]int, height)
		for y := range newGrid {
			newGrid[y] = make([]int, width)
			copy(newGrid[y], grid[y])
		}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				colorCounts := make(map[int]int)

				for dy := -halfH; dy <= halfH; dy++ {
					ny := y + dy
					if ny < 0 || ny >= height {
						continue
					}
					for dx := -halfW; dx <= halfW; dx++ {
						nx := x + dx
						if nx < 0 || nx >= width {
							continue
						}
						colorCounts[grid[ny][nx]]++
					}
				}

				dominant := grid[y][x]
				maxCount := 0
				for c, count := range colorCounts {
					if count > maxCount {
						maxCount = count
						dominant = c
					}
				}

				newGrid[y][x] = dominant
			}
		}

		for y := 0; y < height; y++ {
			copy(grid[y], newGrid[y])
		}
	}
}

// applyDigitalPixelBlocks adds fine-scale 2×1, 1×2, and 2×2 pixel blocks
// that give the pattern its characteristic digital/pixelated texture.
func (pg *Pat5Generator) applyDigitalPixelBlocks(grid [][]int, width, height int) {
	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {
			if rand.Float64() < 0.25 {
				base := grid[y][x]
				switch rand.Intn(3) {
				case 0:
					grid[y][x+1] = base
				case 1:
					grid[y+1][x] = base
				case 2:
					grid[y][x+1] = base
					grid[y+1][x] = base
					grid[y+1][x+1] = base
				}
			}
		}
	}
}

// renderGrid renders the complete grid to the image
func (pg *Pat5Generator) renderGrid(img *image.NRGBA, grid [][]int, colors []color.RGBA, pixelSize int) {
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gridX := x / pixelSize
			gridY := y / pixelSize

			if gridY < len(grid) && gridX < len(grid[gridY]) {
				colorIdx := grid[gridY][gridX]
				if colorIdx >= 0 && colorIdx < len(colors) {
					img.Set(x, y, colors[colorIdx])
				}
			}
		}
	}
}

func (pg *Pat5Generator) initializePat5Grid(width, height int, colorRatios []float64, numColors int) [][]int {
	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, width)
	}

	ratios := colorRatios
	if len(ratios) == 0 || len(ratios) != numColors {
		ratios = pg.getPat5DefaultColorRatios(numColors)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grid[y][x] = selectWeightedColor(ratios, numColors)
		}
	}

	return grid
}

// enforceColorRatios corrects colour distribution after CA and pixel-block passes by placing
// rectangular clusters of under-represented colours into regions of over-represented colours.
// Runs up to 10 iterations until every colour is within 2% of its target ratio.
func (pg *Pat5Generator) enforceColorRatios(grid [][]int, width, height int, ratios []float64, numColors int) {
	total := width * height

	for iter := 0; iter < 10; iter++ {
		counts := pg.countColors(grid, width, height, numColors)

		// Check whether every colour is within tolerance.
		allOK := true
		for i := 0; i < numColors; i++ {
			if math.Abs(float64(counts[i])/float64(total)-ratios[i]) > 0.02 {
				allOK = false
				break
			}
		}
		if allOK {
			break
		}

		// Identify the most under-represented and most over-represented colours.
		underColor, overColor := -1, -1
		maxDeficit, maxExcess := 0.0, 0.0
		for i := 0; i < numColors; i++ {
			actual := float64(counts[i]) / float64(total)
			if d := ratios[i] - actual; d > maxDeficit {
				maxDeficit = d
				underColor = i
			}
			if e := actual - ratios[i]; e > maxExcess {
				maxExcess = e
				overColor = i
			}
		}
		if underColor < 0 || overColor < 0 {
			break
		}

		// Place rectangular patches of underColor inside overColor territory.
		// Patch size (3–7 wide, 2–5 tall grid cells) mirrors the pat5 cluster scale.
		targetPixels := int(maxDeficit * float64(total))
		placed := 0
		maxAttempts := total * 4

		for attempt := 0; attempt < maxAttempts && placed < targetPixels; attempt++ {
			x := rand.Intn(width)
			y := rand.Intn(height)
			if grid[y][x] != overColor {
				continue
			}
			w := 3 + rand.Intn(5) // 3–7 cells wide
			h := 2 + rand.Intn(4) // 2–5 cells tall
			for dy := 0; dy < h && y+dy < height; dy++ {
				for dx := 0; dx < w && x+dx < width; dx++ {
					if grid[y+dy][x+dx] == overColor {
						grid[y+dy][x+dx] = underColor
						placed++
					}
				}
			}
		}
	}
}

// countColors returns per-colour pixel counts for the grid.
func (pg *Pat5Generator) countColors(grid [][]int, width, height, numColors int) []int {
	counts := make([]int, numColors)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if c := grid[y][x]; c >= 0 && c < numColors {
				counts[c]++
			}
		}
	}
	return counts
}

// getPat5DefaultColorRatios returns pat5's default colour distribution ratios.
func (pg *Pat5Generator) getPat5DefaultColorRatios(numColors int) []float64 {
	if numColors == 4 {
		return []float64{0.45, 0.30, 0.15, 0.10}
	} else if numColors == 3 {
		return []float64{0.50, 0.35, 0.15}
	}
	equal := 1.0 / float64(numColors)
	ratios := make([]float64, numColors)
	for i := range ratios {
		ratios[i] = equal
	}
	return ratios
}
