package generator

import (
	"context"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/bradsec/gocamo/pkg/config"
)

// FleckGenerator creates fleck (Flecktarn-inspired) camouflage patterns built
// from many small overlapping rounded spots. Spot density is modulated by a
// low-frequency noise field per color so flecks clump into larger irregular
// regions instead of spreading uniformly, mimicking the German Flecktarn
// disruptive dot pattern.
type FleckGenerator struct{}

// Generate creates a fleck-style camouflage pattern. It fills the grid with a
// weighted background color, layers clumped fleck spots for every other color
// in palette order (later colors sit on top), and finishes with a sparse
// single-cell speckle pass for fine texture.
func (fg *FleckGenerator) Generate(ctx context.Context, cfg *config.Config, colors []color.RGBA) (image.Image, error) {
	// Use the centralized pixel size adjustment for perfect fit
	pixelSize := cfg.AdjustBasePixelSize()

	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))

	gridWidth := cfg.Width / pixelSize
	gridHeight := cfg.Height / pixelSize

	// Background color chosen with the configured weighting
	backgroundIndex := selectWeightedColor(cfg.ColorRatios, len(colors))

	grid := make([][]int, gridHeight)
	for y := range grid {
		grid[y] = make([]int, gridWidth)
		for x := range grid[y] {
			grid[y][x] = backgroundIndex
		}
	}

	// Layer fleck spots for each non-background color. Each color gets its
	// own density field offset so the clumped regions do not coincide.
	for colorIdx := range colors {
		if colorIdx == backgroundIndex {
			continue
		}

		if err := ctxErr(ctx); err != nil {
			return nil, err
		}

		fieldOffset := rand.Float64() * 1000
		fg.scatterFlecks(grid, gridWidth, gridHeight, colorIdx, fieldOffset)
	}

	fg.addSpeckles(grid, gridWidth, gridHeight, len(colors))

	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	// Render the grid
	for y := 0; y < cfg.Height; y++ {
		if y%256 == 0 {
			if err := ctxErr(ctx); err != nil {
				return nil, err
			}
		}

		for x := 0; x < cfg.Width; x++ {
			gridY := y / pixelSize
			gridX := x / pixelSize
			if gridY < gridHeight && gridX < gridWidth {
				img.Set(x, y, colors[grid[gridY][gridX]])
			}
		}
	}

	if cfg.AddNoise {
		addNoiseNRGBA(img, colors)
	}

	if cfg.AddEdge {
		addEdgeDetailsNRGBA(img, pixelSize)
	}

	return img, nil
}

// scatterFlecks places clumped rounded spots of one color. Spots are accepted
// only where the low-frequency density field is strong, which produces the
// characteristic Flecktarn clusters of dots separated by calmer areas.
func (fg *FleckGenerator) scatterFlecks(grid [][]int, width, height, colorIdx int, fieldOffset float64) {
	attempts := (width * height) / 8

	for i := 0; i < attempts; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)

		density := fg.valueNoise(float64(x)*0.045+fieldOffset, float64(y)*0.045+fieldOffset)
		// Squaring the field sharpens the contrast between clump centers
		// and empty areas.
		if rand.Float64() > density*density*1.6 {
			continue
		}

		radius := 1 + rand.Intn(3)
		fg.drawFleck(grid, width, height, x, y, radius, colorIdx)
	}
}

// drawFleck stamps one irregular rounded spot onto the grid.
func (fg *FleckGenerator) drawFleck(grid [][]int, width, height, centerX, centerY, radius, colorIdx int) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			x := centerX + dx
			y := centerY + dy
			if x < 0 || x >= width || y < 0 || y >= height {
				continue
			}

			dist := math.Sqrt(float64(dx*dx + dy*dy))
			// Jittered edge keeps spots irregular rather than circular
			if dist <= float64(radius)+(rand.Float64()-0.5) {
				grid[y][x] = colorIdx
			}
		}
	}
}

// addSpeckles sprinkles sparse single-cell dots for fine dotted texture.
func (fg *FleckGenerator) addSpeckles(grid [][]int, width, height, numColors int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rand.Float64() < 0.02 {
				grid[y][x] = rand.Intn(numColors)
			}
		}
	}
}

// valueNoise returns smooth 2D value noise in [0,1] using smoothstep
// interpolation of hashed lattice points.
func (fg *FleckGenerator) valueNoise(x, y float64) float64 {
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	tx := x - x0
	ty := y - y0

	sx := tx * tx * (3 - 2*tx)
	sy := ty * ty * (3 - 2*ty)

	n00 := fg.hashLattice(int(x0), int(y0))
	n10 := fg.hashLattice(int(x0)+1, int(y0))
	n01 := fg.hashLattice(int(x0), int(y0)+1)
	n11 := fg.hashLattice(int(x0)+1, int(y0)+1)

	top := n00 + (n10-n00)*sx
	bottom := n01 + (n11-n01)*sx

	return top + (bottom-top)*sy
}

// hashLattice maps integer lattice coordinates to a deterministic value in [0,1].
func (fg *FleckGenerator) hashLattice(x, y int) float64 {
	n := x + y*57
	n = (n << 13) ^ n

	return float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff) / 2147483647.0
}
