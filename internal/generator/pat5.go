package generator

import (
	"context"
	"fmt"
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

// Template represents a rectangular pattern template used in MARPAT generation
type Template struct {
	Width  int
	Height int
	Matrix [][]bool // true = filled, false = empty
}

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

// PixelBlock represents a block with actual pixel dimensions for rendering
type PixelBlock struct {
	Width       int     // Actual pixel width
	Height      int     // Actual pixel height
	BlockType   string  // "full", "half_h", "quarter", "half_v"
	Probability float64 // Weighted probability for selection
}

// BlockRegion represents a region in the image with a specific block size and color
type BlockRegion struct {
	X         int // Pixel X position
	Y         int // Pixel Y position
	Width     int // Block width in pixels
	Height    int // Block height in pixels
	ColorIdx  int // Color index to use
}

// ClusterSeed represents a starting point for organic cluster growth
type ClusterSeed struct {
	X         int     // Pixel X position
	Y         int     // Pixel Y position
	ColorIdx  int     // Color index for this cluster
	MaxSize   int     // Maximum cluster size
	GrowthDir []int   // Preferred growth directions (0-7)
	Intensity float64 // Growth probability/strength
}

// OrganicCluster represents a grown organic cluster of connected pixels
type OrganicCluster struct {
	Pixels    []Pixel // All pixels in the cluster
	ColorIdx  int     // Color index for the cluster
	BlockType string  // "large", "medium", "small", "detail"
}

// Pixel represents a single pixel position
type Pixel struct {
	X int // Pixel X position
	Y int // Pixel Y position
}

// Generate creates a pat5-style MARPAT-inspired camouflage pattern using fractal-based
// grid generation for authentic digital camouflage with 100% coverage.
func (pg *Pat5Generator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Use the centralized pixel size adjustment for perfect fit
	adjustedBasePixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	// Calculate grid dimensions for complete coverage
	gridWidth := cfg.Width / adjustedBasePixelSize
	gridHeight := cfg.Height / adjustedBasePixelSize

	// Create complete pixel grid (like other successful patterns)
	grid := make([][]int, gridHeight)
	for y := range grid {
		grid[y] = make([]int, gridWidth)
	}

	// Layer 1: Initialize with MARPAT digital pixel foundation
	pg.initializeDigitalPixelBase(grid, gridWidth, gridHeight, colors, cfg.ColorRatios)

	// Layer 2: Apply multi-scale digital pixel clustering (like authentic MARPAT)
	pg.applyMARPATPixelClustering(grid, gridWidth, gridHeight, len(colors))

	// Layer 3: Add small-scale digital texture and noise
	pg.addDigitalTextureNoise(grid, gridWidth, gridHeight, len(colors))

	// Render the complete grid with 100% coverage
	pg.renderGrid(img, grid, colors, adjustedBasePixelSize)

	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, adjustedBasePixelSize)
	}

	return img, nil
}

// initializeDigitalPixelBase creates authentic MARPAT digital pixel foundation
func (pg *Pat5Generator) initializeDigitalPixelBase(grid [][]int, width, height int, colors []color.RGBA, colorRatios []float64) {
	// Use MARPAT ratios if not provided
	ratios := colorRatios
	if len(ratios) == 0 || len(ratios) != len(colors) {
		ratios = pg.getMARPATColorRatios(len(colors))
	}

	// Initialize with weighted random color assignment (high frequency digital noise)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Use high-frequency noise for initial pixel assignment
			noiseX := float64(x) * 0.3
			noiseY := float64(y) * 0.3
			noise := (pg.digitalNoise(noiseX, noiseY) + 1.0) / 2.0

			// Convert noise to color using weighted distribution
			colorIdx := pg.noiseToColorIndex(noise, ratios, len(colors))
			grid[y][x] = colorIdx
		}
	}
}

// digitalNoise generates high-frequency digital noise for MARPAT pixels
func (pg *Pat5Generator) digitalNoise(x, y float64) float64 {
	// Use integer coordinates for digital/pixelated effect
	xi := int(x) & 255
	yi := int(y) & 255

	// Simple hash-based noise for sharp digital transitions
	n := xi + yi*57
	n = (n << 13) ^ n
	return float64(((n*(n*n*15731+789221)+1376312589)&0x7fffffff))/1073741824.0 - 1.0
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

// generateFractalNoise creates multi-octave fractal noise for organic patterns
func (pg *Pat5Generator) generateFractalNoise(x, y, width, height int) float64 {
	// Multiple octaves for fractal complexity
	noise := 0.0
	amplitude := 1.0
	frequency := 0.02 // Base frequency
	maxValue := 0.0   // For normalization

	// Add 4 octaves of noise
	for i := 0; i < 4; i++ {
		noise += pg.perlinNoise(float64(x)*frequency, float64(y)*frequency) * amplitude
		maxValue += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	// Normalize to [0, 1]
	return (noise/maxValue + 1.0) / 2.0
}

// perlinNoise generates Perlin noise value at given coordinates
func (pg *Pat5Generator) perlinNoise(x, y float64) float64 {
	// Simplified Perlin noise implementation
	xi := int(x) & 255
	yi := int(y) & 255
	xf := x - float64(int(x))
	yf := y - float64(int(y))

	// Fade function for smooth interpolation
	u := xf * xf * (3.0 - 2.0*xf)
	v := yf * yf * (3.0 - 2.0*yf)

	// Hash function for gradient vectors
	aa := pg.hash(xi, yi)
	ab := pg.hash(xi, yi+1)
	ba := pg.hash(xi+1, yi)
	bb := pg.hash(xi+1, yi+1)

	// Interpolate
	x1 := pg.lerp(aa, ba, u)
	x2 := pg.lerp(ab, bb, u)
	return pg.lerp(x1, x2, v)
}

// hash generates pseudo-random value for noise
func (pg *Pat5Generator) hash(x, y int) float64 {
	n := x + y*57
	n = (n << 13) ^ n
	return (1.0 - float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff)/1073741824.0)
}

// lerp performs linear interpolation
func (pg *Pat5Generator) lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

// noiseToColorIndex converts fractal noise to color index using weighted ratios
func (pg *Pat5Generator) noiseToColorIndex(noise float64, colorRatios []float64, numColors int) int {
	// Use MARPAT ratios if not provided
	ratios := colorRatios
	if len(ratios) == 0 || len(ratios) != numColors {
		ratios = pg.getMARPATColorRatios(numColors)
	}

	// Convert noise value to color index using cumulative distribution
	cumulative := 0.0
	for i, ratio := range ratios {
		cumulative += ratio
		if noise <= cumulative {
			return i
		}
	}
	return numColors - 1 // Fallback
}

// initializeMARPATGrid creates a base grid with MARPAT-style color distribution
// 40-50% base colors, 25-35% secondary, 10-20% accent colors
// applyCellularAutomataClustering applies cellular automata rules for digital clustering
func (pg *Pat5Generator) applyCellularAutomataClustering(grid [][]int, width, height, numColors int) {
	// Apply cellular automata iterations to create organic clustering
	iterations := 3 // Number of CA iterations

	for iter := 0; iter < iterations; iter++ {
		newGrid := make([][]int, height)
		for y := range newGrid {
			newGrid[y] = make([]int, width)
		}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				// Count neighbors for each color
				colorCounts := make(map[int]int)
				totalNeighbors := 0

				// Check 3x3 neighborhood
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx := (x + dx + width) % width   // Wrap around
						ny := (y + dy + height) % height // Wrap around
						colorCounts[grid[ny][nx]]++
						totalNeighbors++
					}
				}

				// Find dominant color in neighborhood
				dominantColor := grid[y][x] // Default to current
				maxCount := 0
				for color, count := range colorCounts {
					if count > maxCount {
						maxCount = count
						dominantColor = color
					}
				}

				// Apply clustering rules (bias toward dominant neighbor color)
				if maxCount >= 5 { // If 5 or more neighbors are same color
					newGrid[y][x] = dominantColor
				} else {
					// Keep original with some probability of changing
					if rand.Float64() < 0.3 { // 30% chance to change
						newGrid[y][x] = dominantColor
					} else {
						newGrid[y][x] = grid[y][x]
					}
				}
			}
		}

		// Copy new grid back
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				grid[y][x] = newGrid[y][x]
			}
		}
	}
}

// addRectangularPixelStructure adds MARPAT-style rectangular pixel blocks
func (pg *Pat5Generator) addRectangularPixelStructure(grid [][]int, width, height, numColors int) {
	// Create rectangular pixel groups (2x1, 1x2, 2x2 blocks)
	blockSize := 2

	for y := 0; y < height-blockSize; y += blockSize {
		for x := 0; x < width-blockSize; x += blockSize {
			// Randomly choose block type
			blockType := rand.Intn(4)

			switch blockType {
			case 0: // 2x1 horizontal rectangle
				dominantColor := pg.getDominantColorInArea(grid, x, y, 2, 1)
				for bx := 0; bx < 2 && x+bx < width; bx++ {
					if y < height {
						grid[y][x+bx] = dominantColor
					}
				}
			case 1: // 1x2 vertical rectangle
				dominantColor := pg.getDominantColorInArea(grid, x, y, 1, 2)
				for by := 0; by < 2 && y+by < height; by++ {
					if x < width {
						grid[y+by][x] = dominantColor
					}
				}
			case 2: // 2x2 square block
				dominantColor := pg.getDominantColorInArea(grid, x, y, 2, 2)
				for by := 0; by < 2 && y+by < height; by++ {
					for bx := 0; bx < 2 && x+bx < width; bx++ {
						grid[y+by][x+bx] = dominantColor
					}
				}
			case 3: // Leave as-is (preserve fractal detail)
				// Do nothing, keep original fractal pattern
			}
		}
	}
}

// addMultiScaleRefinement adds detail at multiple scales
func (pg *Pat5Generator) addMultiScaleRefinement(grid [][]int, width, height, numColors int) {
	// Add fine-scale details
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Add some randomness for texture (10% chance)
			if rand.Float64() < 0.1 {
				// Choose a different color from neighbors
				neighbors := []int{}
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx := (x + dx + width) % width
						ny := (y + dy + height) % height
						neighbors = append(neighbors, grid[ny][nx])
					}
				}
				if len(neighbors) > 0 {
					grid[y][x] = neighbors[rand.Intn(len(neighbors))]
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

// getDominantColorInArea finds the most common color in a rectangular area
func (pg *Pat5Generator) getDominantColorInArea(grid [][]int, x, y, w, h int) int {
	colorCounts := make(map[int]int)
	maxCount := 0
	dominantColor := 0

	// Count colors in the specified area
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			if y+dy < len(grid) && x+dx < len(grid[y+dy]) {
				color := grid[y+dy][x+dx]
				colorCounts[color]++
				if colorCounts[color] > maxCount {
					maxCount = colorCounts[color]
					dominantColor = color
				}
			}
		}
	}

	return dominantColor
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

// createRectangularTemplates generates rectangular templates for MARPAT-style patterns with varied proportions
func (pg *Pat5Generator) createRectangularTemplates() []Template {
	templates := []Template{
		// Square templates (traditional)
		pg.createTemplate(3, 3, [][]bool{
			{true, true, true},
			{true, false, true},
			{true, true, true},
		}),
		pg.createTemplate(4, 4, [][]bool{
			{true, true, false, false},
			{true, true, false, false},
			{false, false, true, true},
			{false, false, true, true},
		}),

		// Horizontal rectangular templates (MARPAT characteristic)
		pg.createTemplate(6, 3, [][]bool{
			{true, true, true, false, false, false},
			{true, true, true, false, false, false},
			{false, false, false, true, true, true},
		}),
		pg.createTemplate(8, 2, [][]bool{
			{true, true, true, true, false, false, false, false},
			{false, false, false, false, true, true, true, true},
		}),
		pg.createTemplate(4, 2, [][]bool{
			{true, false, true, false},
			{false, true, false, true},
		}),

		// Vertical rectangular templates
		pg.createTemplate(2, 6, [][]bool{
			{true, false},
			{true, false},
			{false, true},
			{false, true},
			{true, false},
			{true, false},
		}),
		pg.createTemplate(3, 4, [][]bool{
			{true, false, true},
			{false, true, false},
			{true, false, true},
			{false, true, false},
		}),

		// Quarter-sized templates
		pg.createTemplate(2, 2, [][]bool{
			{true, false},
			{false, true},
		}),
		pg.createTemplate(3, 3, [][]bool{
			{false, true, false},
			{true, true, true},
			{false, true, false},
		}),

		// Mixed proportion templates (MARPAT-style)
		pg.createTemplate(5, 3, [][]bool{
			{true, false, true, false, true},
			{false, true, true, true, false},
			{true, false, false, false, true},
		}),
		pg.createTemplate(3, 5, [][]bool{
			{true, false, true},
			{false, true, false},
			{true, true, true},
			{false, true, false},
			{true, false, true},
		}),

		// Thin strip templates
		pg.createTemplate(6, 1, [][]bool{
			{true, false, true, false, true, false},
		}),
		pg.createTemplate(1, 6, [][]bool{
			{true},
			{false},
			{true},
			{false},
			{true},
			{false},
		}),
	}

	return templates
}

// createTemplate creates a template with given dimensions and pattern
func (pg *Pat5Generator) createTemplate(width, height int, pattern [][]bool) Template {
	matrix := make([][]bool, height)
	for i := range matrix {
		matrix[i] = make([]bool, width)
		copy(matrix[i], pattern[i])
	}
	return Template{Width: width, Height: height, Matrix: matrix}
}

// applyTemplatesGreedy applies rectangular templates using an improved placement algorithm for non-square rectangles
func (pg *Pat5Generator) applyTemplatesGreedy(grid [][]int, templates []Template, width, height, numColors int) {
	used := make([][]bool, height)
	for i := range used {
		used[i] = make([]bool, width)
	}

	attempts := (width * height) / 15 // More attempts for better coverage with varied templates

	for attempt := 0; attempt < attempts; attempt++ {
		template := templates[rand.Intn(len(templates))]

		// Skip if template is too large for the grid
		if template.Width > width || template.Height > height {
			continue
		}

		startY := rand.Intn(height - template.Height + 1)
		startX := rand.Intn(width - template.Width + 1)

		// Check if template can be placed with minimal overlap tolerance
		canPlace := pg.canPlaceTemplate(used, template, startX, startY, 0.3) // Allow 30% overlap

		if canPlace {
			// Choose color based on template characteristics
			var colorIdx int
			if template.Width > template.Height {
				// Horizontal rectangles - use weighted selection favoring base colors
				colorIdx = selectWeightedColor([]float64{0.5, 0.3, 0.15, 0.05}, numColors)
			} else if template.Height > template.Width {
				// Vertical rectangles - use accent colors more often
				colorIdx = selectWeightedColor([]float64{0.2, 0.3, 0.3, 0.2}, numColors)
			} else {
				// Square templates - balanced distribution
				colorIdx = selectWeightedColor([]float64{0.4, 0.3, 0.2, 0.1}, numColors)
			}

			// Place template with edge-aware placement
			pg.placeTemplateWithEdgeAwareness(grid, used, template, startX, startY, colorIdx)
		}
	}
}

// canPlaceTemplate checks if a template can be placed with overlap tolerance
func (pg *Pat5Generator) canPlaceTemplate(used [][]bool, template Template, startX, startY int, overlapTolerance float64) bool {
	overlapCount := 0
	totalCells := 0

	for ty := 0; ty < template.Height; ty++ {
		for tx := 0; tx < template.Width; tx++ {
			if template.Matrix[ty][tx] {
				totalCells++
				if used[startY+ty][startX+tx] {
					overlapCount++
				}
			}
		}
	}

	if totalCells == 0 {
		return false
	}

	overlapRatio := float64(overlapCount) / float64(totalCells)
	return overlapRatio <= overlapTolerance
}

// placeTemplateWithEdgeAwareness places a template with consideration for edges and existing patterns
func (pg *Pat5Generator) placeTemplateWithEdgeAwareness(grid [][]int, used [][]bool, template Template, startX, startY, colorIdx int) {
	height, width := len(grid), len(grid[0])

	for ty := 0; ty < template.Height; ty++ {
		for tx := 0; tx < template.Width; tx++ {
			if template.Matrix[ty][tx] {
				y, x := startY+ty, startX+tx

				// Edge-aware placement probability
				placementProb := 0.85 // Base probability

				// Increase probability near edges for better pattern continuity
				if y < 2 || y >= height-2 || x < 2 || x >= width-2 {
					placementProb = 0.95
				}

				// Reduce probability if too much clustering
				neighborCount := pg.countSimilarNeighbors(grid, x, y, colorIdx)
				if neighborCount > 6 {
					placementProb = 0.6 // Reduce clustering
				}

				if rand.Float64() < placementProb {
					grid[y][x] = colorIdx
					used[y][x] = true
				}
			}
		}
	}
}

// countSimilarNeighbors counts neighboring cells with the same color in a 3x3 area
func (pg *Pat5Generator) countSimilarNeighbors(grid [][]int, x, y, colorIdx int) int {
	count := 0
	height, width := len(grid), len(grid[0])

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			ny, nx := y+dy, x+dx
			if ny >= 0 && ny < height && nx >= 0 && nx < width {
				if grid[ny][nx] == colorIdx {
					count++
				}
			}
		}
	}

	return count
}

// createRectangleClusters defines MARPAT-style rectangle variations with weighted probabilities
func (pg *Pat5Generator) createRectangleClusters() []RectangleCluster {
	return []RectangleCluster{
		// Full blocks (40% - main pattern base)
		{Width: 4, Height: 4, ClusterType: "full", Probability: 0.25},
		{Width: 6, Height: 6, ClusterType: "full", Probability: 0.15},

		// Horizontal rectangles (30% - MARPAT characteristic)
		{Width: 6, Height: 3, ClusterType: "horizontal", Probability: 0.12},
		{Width: 8, Height: 2, ClusterType: "horizontal", Probability: 0.10},
		{Width: 4, Height: 2, ClusterType: "horizontal", Probability: 0.08},

		// Quarter blocks (20% - detail areas)
		{Width: 2, Height: 2, ClusterType: "quarter", Probability: 0.12},
		{Width: 3, Height: 3, ClusterType: "quarter", Probability: 0.08},

		// Vertical rectangles (10% - variation)
		{Width: 2, Height: 6, ClusterType: "vertical", Probability: 0.05},
		{Width: 3, Height: 4, ClusterType: "vertical", Probability: 0.05},
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

// blendFractalWithTemplates combines fractal noise with template-based patterns
func (pg *Pat5Generator) blendFractalWithTemplates(grid, fractalLayer [][]int, width, height int, fractalWeight float64) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rand.Float64() < fractalWeight {
				grid[y][x] = fractalLayer[y][x]
			}
		}
	}
}

// createPixelBlocks defines MARPAT-style pixel blocks with actual pixel dimensions
func (pg *Pat5Generator) createPixelBlocks(basePixelSize int) []PixelBlock {
	return []PixelBlock{
		// Full blocks (40% - main pattern base)
		{Width: basePixelSize, Height: basePixelSize, BlockType: "full", Probability: 0.40},

		// Half horizontal blocks (30% - MARPAT characteristic)
		{Width: basePixelSize, Height: basePixelSize / 2, BlockType: "half_h", Probability: 0.20},
		{Width: basePixelSize * 2, Height: basePixelSize / 2, BlockType: "half_h", Probability: 0.10},

		// Quarter blocks (20% - detail areas)
		{Width: basePixelSize / 2, Height: basePixelSize / 2, BlockType: "quarter", Probability: 0.15},
		{Width: basePixelSize / 2, Height: basePixelSize, BlockType: "quarter", Probability: 0.05},

		// Half vertical blocks (10% - variation)
		{Width: basePixelSize / 2, Height: basePixelSize, BlockType: "half_v", Probability: 0.05},
		{Width: basePixelSize / 2, Height: basePixelSize * 2, BlockType: "half_v", Probability: 0.05},
	}
}

// selectWeightedPixelBlock selects a pixel block based on weighted probabilities
func (pg *Pat5Generator) selectWeightedPixelBlock(blocks []PixelBlock) PixelBlock {
	r := rand.Float64()
	cumulative := 0.0

	for _, block := range blocks {
		cumulative += block.Probability
		if r <= cumulative {
			return block
		}
	}

	// Fallback to first block
	return blocks[0]
}

// generateBlockRegions creates block regions with varied pixel sizes across the image
func (pg *Pat5Generator) generateBlockRegions(imgWidth, imgHeight int, pixelBlocks []PixelBlock, baseGrid [][]int, basePixelSize, numColors int) []BlockRegion {
	var regions []BlockRegion
	used := make(map[string]bool) // Track used pixel positions

	baseGridWidth := len(baseGrid[0])
	baseGridHeight := len(baseGrid)

	// Generate block regions
	attempts := (imgWidth * imgHeight) / (basePixelSize * 2) // More attempts for smaller blocks

	for attempt := 0; attempt < attempts; attempt++ {
		block := pg.selectWeightedPixelBlock(pixelBlocks)

		// Ensure block dimensions are at least 1 pixel
		blockWidth := block.Width
		blockHeight := block.Height
		if blockWidth < 1 {
			blockWidth = 1
		}
		if blockHeight < 1 {
			blockHeight = 1
		}

		// Random position within image bounds
		x := rand.Intn(imgWidth - blockWidth + 1)
		y := rand.Intn(imgHeight - blockHeight + 1)

		// Align to grid boundaries for better pattern continuity
		x = (x / basePixelSize) * basePixelSize
		y = (y / basePixelSize) * basePixelSize

		// Check if region overlaps significantly with existing blocks
		if pg.hasSignificantOverlap(x, y, blockWidth, blockHeight, used, 0.3) {
			continue
		}

		// Get color from base grid
		baseGridX := x / basePixelSize
		baseGridY := y / basePixelSize
		var colorIdx int
		if baseGridX < baseGridWidth && baseGridY < baseGridHeight {
			colorIdx = baseGrid[baseGridY][baseGridX]
		} else {
			colorIdx = rand.Intn(numColors)
		}

		// Adjust color probability based on block type
		if block.BlockType == "half_h" {
			// Horizontal blocks favor base colors
			colorIdx = selectWeightedColor([]float64{0.6, 0.25, 0.10, 0.05}, numColors)
		} else if block.BlockType == "quarter" {
			// Quarter blocks use more accent colors
			colorIdx = selectWeightedColor([]float64{0.2, 0.3, 0.3, 0.2}, numColors)
		}

		regions = append(regions, BlockRegion{
			X:        x,
			Y:        y,
			Width:    blockWidth,
			Height:   blockHeight,
			ColorIdx: colorIdx,
		})

		// Mark region as used
		pg.markRegionUsed(x, y, blockWidth, blockHeight, used)
	}

	return regions
}

// hasSignificantOverlap checks if a region has significant overlap with existing blocks
func (pg *Pat5Generator) hasSignificantOverlap(x, y, width, height int, used map[string]bool, threshold float64) bool {
	totalPixels := width * height
	overlapPixels := 0

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			key := fmt.Sprintf("%d,%d", x+dx, y+dy)
			if used[key] {
				overlapPixels++
			}
		}
	}

	overlapRatio := float64(overlapPixels) / float64(totalPixels)
	return overlapRatio > threshold
}

// markRegionUsed marks all pixels in a region as used
func (pg *Pat5Generator) markRegionUsed(x, y, width, height int, used map[string]bool) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			key := fmt.Sprintf("%d,%d", x+dx, y+dy)
			used[key] = true
		}
	}
}

// renderPixelBlocks renders block regions at their actual pixel dimensions
func (pg *Pat5Generator) renderPixelBlocks(img *image.NRGBA, regions []BlockRegion, colors []color.RGBA) {
	bounds := img.Bounds()

	for _, region := range regions {
		color := colors[region.ColorIdx]

		// Render the block at its exact pixel dimensions
		for dy := 0; dy < region.Height; dy++ {
			for dx := 0; dx < region.Width; dx++ {
				x := region.X + dx
				y := region.Y + dy

				// Ensure we stay within image bounds
				if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
					img.Set(x, y, color)
				}
			}
		}
	}
}

// initializeImageWithBaseColor fills image with base color using MARPAT distribution
func (pg *Pat5Generator) initializeImageWithBaseColor(img *image.NRGBA, colors []color.RGBA, colorRatios []float64, numColors int) {
	// Use MARPAT ratios if not provided
	ratios := colorRatios
	if len(ratios) == 0 || len(ratios) != numColors {
		ratios = pg.getMARPATColorRatios(numColors)
	}

	// Fill with base color (typically the most prominent color)
	baseColorIdx := selectWeightedColor([]float64{0.8, 0.15, 0.04, 0.01}, numColors)
	baseColor := colors[baseColorIdx]

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, baseColor)
		}
	}
}

// generateClusterSeeds creates seed points for organic cluster growth using blue noise distribution
func (pg *Pat5Generator) generateClusterSeeds(imgWidth, imgHeight, basePixelSize, numColors int) []ClusterSeed {
	var seeds []ClusterSeed

	// Use Poisson disk sampling for natural spatial distribution
	minDistance := float64(basePixelSize) * 0.5 // Allow closer seeds for denser coverage
	maxAttempts := imgWidth * imgHeight / (basePixelSize * basePixelSize * 4) // Generate more seeds

	activePoints := []Pixel{}
	occupiedGrid := make(map[string]bool)
	cellSize := minDistance / math.Sqrt(2)

	// Add initial seed
	// Start with many seeds distributed across the canvas for dense coverage
	initialSeedCount := 20 + (imgWidth * imgHeight) / (basePixelSize * basePixelSize * 100) // Much denser initial seeding
	for i := 0; i < initialSeedCount; i++ {
		// Distribute seeds across entire canvas, not just center
		x := rand.Intn(imgWidth)
		y := rand.Intn(imgHeight)

		initialSeed := ClusterSeed{
			X:         x,
			Y:         y,
			ColorIdx:  rand.Intn(numColors),
			MaxSize:   basePixelSize * (10 + rand.Intn(20)), // Much larger clusters for dense coverage
			GrowthDir: pg.getRandomGrowthDirections(),
			Intensity: 0.5 + rand.Float64()*0.5,
		}
		seeds = append(seeds, initialSeed)
		activePoints = append(activePoints, Pixel{X: x, Y: y})
		pg.markGridOccupied(occupiedGrid, x, y, cellSize)
	}

	// Generate additional seeds using Poisson disk sampling
	attempts := 0
	for len(activePoints) > 0 && attempts < maxAttempts {
		attempts++

		// Pick random active point
		pointIdx := rand.Intn(len(activePoints))
		point := activePoints[pointIdx]

		// Try to place new point around it
		found := false
		for attempt := 0; attempt < 30; attempt++ {
			// Generate candidate point in annulus around active point
			angle := rand.Float64() * 2 * math.Pi
			distance := minDistance + rand.Float64()*minDistance
			newX := int(float64(point.X) + distance*math.Cos(angle))
			newY := int(float64(point.Y) + distance*math.Sin(angle))

			// Check bounds
			if newX < 0 || newX >= imgWidth || newY < 0 || newY >= imgHeight {
				continue
			}

			// Check minimum distance from other points
			if pg.isValidSeedPosition(occupiedGrid, newX, newY, cellSize, minDistance) {
				// Create new seed
				newSeed := ClusterSeed{
					X:         newX,
					Y:         newY,
					ColorIdx:  rand.Intn(numColors),
					MaxSize:   basePixelSize * (8 + rand.Intn(18)), // Much larger secondary clusters
					GrowthDir: pg.getRandomGrowthDirections(),
					Intensity: 0.5 + rand.Float64()*0.5,
				}
				seeds = append(seeds, newSeed)
				activePoints = append(activePoints, Pixel{X: newX, Y: newY})
				pg.markGridOccupied(occupiedGrid, newX, newY, cellSize)
				found = true
				break
			}
		}

		// Remove active point if no valid position found
		if !found {
			activePoints = append(activePoints[:pointIdx], activePoints[pointIdx+1:]...)
		}
	}

	return seeds
}

// getRandomGrowthDirections returns preferred growth directions for organic clustering
func (pg *Pat5Generator) getRandomGrowthDirections() []int {
	numDirections := 2 + rand.Intn(3) // 2-4 preferred directions
	directions := []int{}
	used := make(map[int]bool)

	for len(directions) < numDirections {
		dir := rand.Intn(8) // 8 directions (N, NE, E, SE, S, SW, W, NW)
		if !used[dir] {
			directions = append(directions, dir)
			used[dir] = true
		}
	}

	return directions
}

// isValidSeedPosition checks if a seed position maintains minimum distance from others
func (pg *Pat5Generator) isValidSeedPosition(occupiedGrid map[string]bool, x, y int, cellSize, minDistance float64) bool {
	// Check surrounding grid cells
	gridX := int(float64(x) / cellSize)
	gridY := int(float64(y) / cellSize)

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			key := fmt.Sprintf("%d,%d", gridX+dx, gridY+dy)
			if occupiedGrid[key] {
				// Would be too close to existing point
				return false
			}
		}
	}

	return true
}

// markGridOccupied marks grid cell as occupied for spatial distribution
func (pg *Pat5Generator) markGridOccupied(occupiedGrid map[string]bool, x, y int, cellSize float64) {
	gridX := int(float64(x) / cellSize)
	gridY := int(float64(y) / cellSize)
	key := fmt.Sprintf("%d,%d", gridX, gridY)
	occupiedGrid[key] = true
}

// growOrganicClusters grows organic clusters from seeds using cellular automata rules
func (pg *Pat5Generator) growOrganicClusters(imgWidth, imgHeight int, seeds []ClusterSeed, basePixelSize, numColors int) []OrganicCluster {
	clusters := []OrganicCluster{}

	// Direction vectors for 8-connected growth
	directions := []Pixel{
		{-1, -1}, {0, -1}, {1, -1}, // NW, N, NE
		{-1, 0}, {1, 0},           // W, E
		{-1, 1}, {0, 1}, {1, 1},   // SW, S, SE
	}

	occupiedPixels := make(map[string]bool)

	for _, seed := range seeds {
		cluster := OrganicCluster{
			Pixels:   []Pixel{},
			ColorIdx: seed.ColorIdx,
			BlockType: pg.getClusterBlockType(seed.MaxSize, basePixelSize),
		}

		// Start with seed pixel
		activePixels := []Pixel{{X: seed.X, Y: seed.Y}}
		cluster.Pixels = append(cluster.Pixels, Pixel{X: seed.X, Y: seed.Y})
		occupiedPixels[fmt.Sprintf("%d,%d", seed.X, seed.Y)] = true

		// Grow cluster using cellular automata-like rules
		for len(activePixels) > 0 && len(cluster.Pixels) < seed.MaxSize {
			newActivePixels := []Pixel{}

			for _, activePixel := range activePixels {
				// Try to grow in preferred directions with higher probability
				growthAttempts := 2 + rand.Intn(4) // 2-5 growth attempts per active pixel

				for attempt := 0; attempt < growthAttempts && len(cluster.Pixels) < seed.MaxSize; attempt++ {
					// Choose direction - prefer seed's growth directions
					var direction Pixel
					if len(seed.GrowthDir) > 0 && rand.Float64() < 0.7 {
						// Prefer seed's growth directions 70% of the time
						dirIdx := seed.GrowthDir[rand.Intn(len(seed.GrowthDir))]
						direction = directions[dirIdx]
					} else {
						// Random direction 30% of the time for organic variation
						direction = directions[rand.Intn(len(directions))]
					}

					newX := activePixel.X + direction.X
					newY := activePixel.Y + direction.Y

					// Check bounds
					if newX < 0 || newX >= imgWidth || newY < 0 || newY >= imgHeight {
						continue
					}

					pixelKey := fmt.Sprintf("%d,%d", newX, newY)
					if occupiedPixels[pixelKey] {
						continue
					}

					// Growth probability based on seed intensity and cluster density
					localDensity := pg.calculateLocalDensity(newX, newY, cluster.Pixels, basePixelSize)
					growthProb := seed.Intensity * (1.0 - localDensity*0.3) // Reduce growth in dense areas

					if rand.Float64() < growthProb {
						newPixel := Pixel{X: newX, Y: newY}
						cluster.Pixels = append(cluster.Pixels, newPixel)
						newActivePixels = append(newActivePixels, newPixel)
						occupiedPixels[pixelKey] = true
					}
				}
			}

			// Update active pixels (only newer pixels remain active for next iteration)
			activePixels = newActivePixels

			// Gradually reduce number of active pixels for natural tapering
			if len(activePixels) > 1 && rand.Float64() < 0.3 {
				// Randomly deactivate some pixels
				activePixels = activePixels[:len(activePixels)/2]
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// calculateLocalDensity calculates pixel density around a position for growth control
func (pg *Pat5Generator) calculateLocalDensity(x, y int, clusterPixels []Pixel, radius int) float64 {
	nearby := 0
	total := 0

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				total++
				checkX, checkY := x+dx, y+dy

				for _, pixel := range clusterPixels {
					if pixel.X == checkX && pixel.Y == checkY {
						nearby++
						break
					}
				}
			}
		}
	}

	if total == 0 {
		return 0
	}
	return float64(nearby) / float64(total)
}

// getClusterBlockType determines cluster type based on size
func (pg *Pat5Generator) getClusterBlockType(maxSize, basePixelSize int) string {
	sizeRatio := float64(maxSize) / float64(basePixelSize)

	if sizeRatio > 20 {
		return "large"
	} else if sizeRatio > 8 {
		return "medium"
	} else if sizeRatio > 3 {
		return "small"
	} else {
		return "detail"
	}
}

// addMultiScaleDetails adds small detail clusters for realistic MARPAT appearance
func (pg *Pat5Generator) addMultiScaleDetails(clusters []OrganicCluster, imgWidth, imgHeight, basePixelSize, numColors int) {
	// Add small detail clusters in empty areas
	detailAttempts := (imgWidth * imgHeight) / (basePixelSize * basePixelSize * 4)

	occupiedPixels := make(map[string]bool)
	for _, cluster := range clusters {
		for _, pixel := range cluster.Pixels {
			occupiedPixels[fmt.Sprintf("%d,%d", pixel.X, pixel.Y)] = true
		}
	}

	for attempt := 0; attempt < detailAttempts; attempt++ {
		x := rand.Intn(imgWidth)
		y := rand.Intn(imgHeight)

		pixelKey := fmt.Sprintf("%d,%d", x, y)
		if occupiedPixels[pixelKey] {
			continue
		}

		// Create small detail cluster
		maxDetailSize := basePixelSize / 2
		if maxDetailSize < 1 {
			maxDetailSize = 1
		}
		detailSize := 1 + rand.Intn(maxDetailSize)

		detailCluster := OrganicCluster{
			Pixels:    []Pixel{{X: x, Y: y}},
			ColorIdx:  rand.Intn(numColors),
			BlockType: "detail",
		}

		// Add a few more pixels to detail cluster
		for i := 0; i < detailSize; i++ {
			if len(detailCluster.Pixels) > 0 {
				basePixel := detailCluster.Pixels[rand.Intn(len(detailCluster.Pixels))]
				newX := basePixel.X + rand.Intn(3) - 1
				newY := basePixel.Y + rand.Intn(3) - 1

				if newX >= 0 && newX < imgWidth && newY >= 0 && newY < imgHeight {
					newKey := fmt.Sprintf("%d,%d", newX, newY)
					if !occupiedPixels[newKey] {
						detailCluster.Pixels = append(detailCluster.Pixels, Pixel{X: newX, Y: newY})
						occupiedPixels[newKey] = true
					}
				}
			}
		}

		clusters = append(clusters, detailCluster)
	}
}

// renderOrganicClusters renders organic clusters with rectangular pixel blocks
func (pg *Pat5Generator) renderOrganicClusters(img *image.NRGBA, clusters []OrganicCluster, colors []color.RGBA, basePixelSize int) {
	bounds := img.Bounds()

	for _, cluster := range clusters {
		color := colors[cluster.ColorIdx]

		// Determine rectangular block size based on cluster type
		var blockWidth, blockHeight int
		switch cluster.BlockType {
		case "large":
			blockWidth = basePixelSize
			blockHeight = basePixelSize
		case "medium":
			// Mix of full and half blocks
			if rand.Float64() < 0.6 {
				blockWidth = basePixelSize
				blockHeight = basePixelSize
			} else {
				blockWidth = basePixelSize
				blockHeight = basePixelSize / 2
				if blockHeight < 1 {
					blockHeight = 1
				}
			}
		case "small":
			// More half and quarter blocks
			switch rand.Intn(3) {
			case 0:
				blockWidth = basePixelSize / 2
				blockHeight = basePixelSize / 2
			case 1:
				blockWidth = basePixelSize
				blockHeight = basePixelSize / 2
			case 2:
				blockWidth = basePixelSize / 2
				blockHeight = basePixelSize
			}
			if blockWidth < 1 {
				blockWidth = 1
			}
			if blockHeight < 1 {
				blockHeight = 1
			}
		case "detail":
			blockWidth = 1
			blockHeight = 1
		}

		// Render each pixel in cluster with rectangular blocks
		for _, pixel := range cluster.Pixels {
			// Align to rectangular block boundaries
			alignedX := (pixel.X / blockWidth) * blockWidth
			alignedY := (pixel.Y / blockHeight) * blockHeight

			// Draw rectangular block
			for dy := 0; dy < blockHeight; dy++ {
				for dx := 0; dx < blockWidth; dx++ {
					x := alignedX + dx
					y := alignedY + dy

					if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
						img.Set(x, y, color)
					}
				}
			}
		}
	}
}