package generator

import (
	"context"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

// Pat5Generator creates pat5 (MARPAT-inspired) camouflage patterns using fractal algorithms
// and rectangular template systems. It generates digital patterns with variable rectangle sizes
// and unequal color distribution matching authentic MARPAT characteristics.
type Pat5Generator struct{}

// FractalParams holds parameters for Iterative Function System (IFS) generation
type FractalParams struct {
	ScaleX    float64
	ScaleY    float64
	RotAngle  float64
	TransX    float64
	TransY    float64
	ColorIdx  int
	Prob      float64
}

// RectangleCluster represents a rectangular cluster with separate width and height
type RectangleCluster struct {
	Width       int
	Height      int
	ClusterType string  // "full", "horizontal", "quarter", "vertical", "strip"
	Probability float64 // Weighted probability for selection
}

// Generate creates a pat5-style MARPAT-inspired camouflage pattern using fractal-based
// grid generation for authentic digital camouflage with 100% coverage.
// Generate creates a MARPAT-inspired digital camouflage pattern using a 5-layer pipeline:
// IFS fractal macro structure → weighted colour grid → rectangle clustering → digital pixels → texture noise.
func (pg *Pat5Generator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	adjustedBasePixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	gridWidth := cfg.Width / adjustedBasePixelSize
	gridHeight := cfg.Height / adjustedBasePixelSize

	// Use MARPAT ratios when the user has not specified explicit ratios.
	ratios := cfg.ColorRatios
	if cfg.RatiosString == "" {
		ratios = pg.getMARPATColorRatios(len(colors))
	}

	// Layer 1: IFS fractal — large-scale macro colour structure (sparse hint layer).
	fractalLayer := pg.generateFractalLayer(gridWidth, gridHeight, colors)

	// Layer 2: MARPAT-weighted colour grid — base distribution.
	grid := pg.initializeMARPATGrid(gridWidth, gridHeight, ratios, len(colors))

	// Layer 3: Rectangle clustering — mid-scale rectangular pixel groups guided by fractal.
	pg.applyRectangleClustering(grid, fractalLayer, gridWidth, gridHeight, adjustedBasePixelSize)

	// Layer 4: Digital pixel clustering — small 1×2/2×1/2×2 digital blocks.
	pg.applyMARPATPixelClustering(grid, gridWidth, gridHeight, len(colors))

	// Layer 5: Fine texture noise — 5% single-pixel variation.
	pg.addDigitalTextureNoise(grid, gridWidth, gridHeight, len(colors))

	pg.renderGrid(img, grid, colors, adjustedBasePixelSize)

	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}
	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, adjustedBasePixelSize)
	}

	return img, nil
}

// applyMARPATPixelClustering creates MARPAT-style digital pixel clusters
func (pg *Pat5Generator) applyMARPATPixelClustering(grid [][]int, width, height, numColors int) {
	// Apply multiple passes of small-scale digital clustering
	for pass := 0; pass < 2; pass++ {
		// Create small rectangular pixel blocks (1x2, 2x1, 2x2)
		for y := 0; y < height-1; y += 1 {
			for x := 0; x < width-1; x += 1 {
				// Randomly choose to create a digital pixel block
				if rand.Float64() < 0.4 { // 40% chance to create block
					blockType := rand.Intn(3)
					baseColor := grid[y][x]

					switch blockType {
					case 0: // 2x1 horizontal rectangle
						if x+1 < width {
							grid[y][x+1] = baseColor
						}
					case 1: // 1x2 vertical rectangle
						if y+1 < height {
							grid[y+1][x] = baseColor
						}
					case 2: // 2x2 square block
						if x+1 < width && y+1 < height {
							grid[y][x+1] = baseColor
							grid[y+1][x] = baseColor
							grid[y+1][x+1] = baseColor
						}
					}
				}
			}
		}

		// Apply light cellular automata for clustering (preserve digital appearance)
		pg.applyLightCellularAutomata(grid, width, height, numColors)
	}
}

// applyLightCellularAutomata applies minimal cellular automata to maintain digital pixels
func (pg *Pat5Generator) applyLightCellularAutomata(grid [][]int, width, height, numColors int) {
	newGrid := make([][]int, height)
	for y := range newGrid {
		newGrid[y] = make([]int, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Count immediate neighbors (cross pattern only for sharper digital edges)
			colorCounts := make(map[int]int)
			neighbors := 0

			// Check only 4-connected neighbors (not 8) for sharper digital appearance
			directions := [][]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}
			for _, dir := range directions {
				nx := x + dir[0]
				ny := y + dir[1]
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					colorCounts[grid[ny][nx]]++
					neighbors++
				}
			}
			colorCounts[grid[y][x]]++ // Include self
			neighbors++

			// Find dominant color
			dominantColor := grid[y][x]
			maxCount := 0
			for color, count := range colorCounts {
				if count > maxCount {
					maxCount = count
					dominantColor = color
				}
			}

			// Apply very light clustering (preserve digital sharpness)
			if maxCount >= 3 && rand.Float64() < 0.3 { // Light clustering
				newGrid[y][x] = dominantColor
			} else {
				newGrid[y][x] = grid[y][x] // Keep original
			}
		}
	}

	// Copy back
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grid[y][x] = newGrid[y][x]
		}
	}
}

// addDigitalTextureNoise adds fine-scale digital texture
func (pg *Pat5Generator) addDigitalTextureNoise(grid [][]int, width, height, numColors int) {
	// Add random single-pixel noise for authentic digital texture
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 5% chance to randomly change pixel (creates digital texture)
			if rand.Float64() < 0.05 {
				// Choose a different color from available colors
				grid[y][x] = rand.Intn(numColors)
			}
		}
	}
}

// renderGrid renders the complete grid to the image
func (pg *Pat5Generator) renderGrid(img *image.NRGBA, grid [][]int, colors []color.RGBA, pixelSize int) {
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Map pixel coordinates to grid coordinates
			gridX := x / pixelSize
			gridY := y / pixelSize

			// Ensure we don't go out of bounds
			if gridY < len(grid) && gridX < len(grid[gridY]) {
				colorIdx := grid[gridY][gridX]
				if colorIdx >= 0 && colorIdx < len(colors) {
					img.Set(x, y, colors[colorIdx])
				}
			}
		}
	}
}

func (pg *Pat5Generator) initializeMARPATGrid(width, height int, colorRatios []float64, numColors int) [][]int {
	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, width)
	}

	// If no custom ratios, use MARPAT-style distribution
	ratios := colorRatios
	if len(ratios) == 0 || len(ratios) != numColors {
		ratios = pg.getMARPATColorRatios(numColors)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grid[y][x] = selectWeightedColor(ratios, numColors)
		}
	}

	return grid
}

// getMARPATColorRatios returns authentic MARPAT color distribution ratios
func (pg *Pat5Generator) getMARPATColorRatios(numColors int) []float64 {
	if numColors == 4 {
		// Authentic 4-color MARPAT distribution
		return []float64{0.45, 0.30, 0.15, 0.10} // Base, secondary, accent, highlight
	} else if numColors == 3 {
		// 3-color adapted distribution
		return []float64{0.50, 0.35, 0.15}
	} else {
		// Equal distribution fallback
		equal := 1.0 / float64(numColors)
		ratios := make([]float64, numColors)
		for i := range ratios {
			ratios[i] = equal
		}
		return ratios
	}
}

// generateFractalLayer creates a fractal noise layer using IFS (Iterative Function System)
func (pg *Pat5Generator) generateFractalLayer(width, height int, colors []color.RGBA) [][]int {
	layer := make([][]int, height)
	for i := range layer {
		layer[i] = make([]int, width)
	}

	// Define IFS parameters for fractal generation
	ifsParams := []FractalParams{
		{ScaleX: 0.5, ScaleY: 0.5, RotAngle: 0, TransX: 0, TransY: 0, ColorIdx: 0, Prob: 0.25},
		{ScaleX: 0.5, ScaleY: 0.5, RotAngle: 0, TransX: 0.5, TransY: 0, ColorIdx: 1, Prob: 0.25},
		{ScaleX: 0.5, ScaleY: 0.5, RotAngle: 0, TransX: 0, TransY: 0.5, ColorIdx: 2, Prob: 0.25},
		{ScaleX: 0.5, ScaleY: 0.5, RotAngle: math.Pi/4, TransX: 0.5, TransY: 0.5, ColorIdx: 3, Prob: 0.25},
	}

	// Generate fractal points
	numPoints := width * height / 4
	x, y := 0.5, 0.5

	for i := 0; i < numPoints; i++ {
		// Select random IFS transformation
		r := rand.Float64()
		var param FractalParams
		cumProb := 0.0
		for _, p := range ifsParams {
			cumProb += p.Prob
			if r <= cumProb {
				param = p
				break
			}
		}

		// Apply fractal transformation
		newX := param.ScaleX*math.Cos(param.RotAngle)*x - param.ScaleY*math.Sin(param.RotAngle)*y + param.TransX
		newY := param.ScaleX*math.Sin(param.RotAngle)*x + param.ScaleY*math.Cos(param.RotAngle)*y + param.TransY

		// Map to grid coordinates
		gridX := int(newX * float64(width))
		gridY := int(newY * float64(height))

		if gridX >= 0 && gridX < width && gridY >= 0 && gridY < height {
			layer[gridY][gridX] = param.ColorIdx % len(colors)
		}

		x, y = newX, newY
	}

	return layer
}

// createRectangleClusters defines MARPAT-style rectangle variations with weighted probabilities
func (pg *Pat5Generator) createRectangleClusters() []RectangleCluster {
	return []RectangleCluster{
		// Full blocks — main pattern base
		{Width: 10, Height: 10, ClusterType: "full", Probability: 0.25},
		{Width: 14, Height: 14, ClusterType: "full", Probability: 0.15},

		// Horizontal rectangles — MARPAT characteristic
		{Width: 14, Height: 6, ClusterType: "horizontal", Probability: 0.12},
		{Width: 18, Height: 4, ClusterType: "horizontal", Probability: 0.10},
		{Width: 10, Height: 4, ClusterType: "horizontal", Probability: 0.08},

		// Quarter blocks — detail areas
		{Width: 5, Height: 5, ClusterType: "quarter", Probability: 0.12},
		{Width: 7, Height: 7, ClusterType: "quarter", Probability: 0.08},

		// Vertical rectangles — variation
		{Width: 4, Height: 14, ClusterType: "vertical", Probability: 0.05},
		{Width: 6, Height: 10, ClusterType: "vertical", Probability: 0.05},
	}
}

// selectWeightedRectangle selects a rectangle cluster based on weighted probabilities
func (pg *Pat5Generator) selectWeightedRectangle(clusters []RectangleCluster) RectangleCluster {
	r := rand.Float64()
	cumulative := 0.0

	for _, cluster := range clusters {
		cumulative += cluster.Probability
		if r <= cumulative {
			return cluster
		}
	}

	// Fallback to first cluster
	return clusters[0]
}

// applyRectangleClustering creates variable-sized rectangular clusters with MARPAT-style variation
func (pg *Pat5Generator) applyRectangleClustering(grid, fractalLayer [][]int, width, height, basePixelSize int) {
	rectangleClusters := pg.createRectangleClusters()
	attempts := (width * height) / 30 // Slightly more attempts for better coverage

	for attempt := 0; attempt < attempts; attempt++ {
		cluster := pg.selectWeightedRectangle(rectangleClusters)

		// Ensure cluster fits within grid bounds
		if cluster.Width >= width || cluster.Height >= height {
			continue
		}

		startY := rand.Intn(height - cluster.Height + 1)
		startX := rand.Intn(width - cluster.Width + 1)

		// Choose color from fractal layer or dominant surrounding color
		var colorIdx int
		if rand.Float64() < 0.4 {
			centerY := startY + cluster.Height/2
			centerX := startX + cluster.Width/2
			colorIdx = fractalLayer[centerY][centerX]
		} else {
			colorIdx = pg.getDominantColor(grid, startX, startY, cluster.Width, cluster.Height)
		}

		// Apply rectangular cluster with varied fill probability based on cluster type
		fillProb := 0.8 // Default fill probability
		switch cluster.ClusterType {
		case "full":
			fillProb = 0.9 // Full blocks more solid
		case "horizontal", "vertical":
			fillProb = 0.85 // Rectangles moderately solid
		case "quarter":
			fillProb = 0.75 // Quarter blocks more fragmented
		}

		for dy := 0; dy < cluster.Height; dy++ {
			for dx := 0; dx < cluster.Width; dx++ {
				if rand.Float64() < fillProb {
					grid[startY+dy][startX+dx] = colorIdx
				}
			}
		}
	}
}

// getDominantColor finds the most common color in a rectangular region
func (pg *Pat5Generator) getDominantColor(grid [][]int, startX, startY, width, height int) int {
	colorCount := make(map[int]int)
	gridHeight, gridWidth := len(grid), len(grid[0])

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			y, x := startY+dy, startX+dx
			if y < gridHeight && x < gridWidth {
				colorCount[grid[y][x]]++
			}
		}
	}

	maxCount := 0
	dominantColor := 0
	for color, count := range colorCount {
		if count > maxCount {
			maxCount = count
			dominantColor = color
		}
	}

	return dominantColor
}

