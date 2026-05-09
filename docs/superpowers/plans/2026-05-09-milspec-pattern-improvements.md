# Milspec Pattern Algorithm Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix bugs and improve all five camouflage pattern generators for higher-fidelity milspec output.

**Architecture:** Fix the `GenerateFromImage` error-ordering bug; move shared Perlin noise helpers to `utils.go`; upgrade pat2 from incoherent hash noise to spatially-coherent Perlin noise; fix pat1 to use the lightest colour as background; add a `milspec` colour-ratio preset; fix pat5 so it uses MARPAT ratios by default when no ratios are specified; wire pat5's dormant IFS fractal → MARPAT grid → rectangle clustering pipeline into `Generate()`; remove all resulting dead code from pat5.

**Tech Stack:** Go 1.21+, `math/rand`, `golang.org/x/image`

---

## File Map

| File | Action | What changes |
|---|---|---|
| `internal/generator/generator.go` | Modify | Move `err` check before `sortColors` call |
| `internal/generator/utils.go` | Modify | Add package-level `perlinNoise`, `noiseHash`, `lerp`, `fractalNoise` |
| `internal/generator/pat2.go` | Modify | Replace `simpleNoise` body; remove dead methods |
| `internal/generator/pat1.go` | Modify | Use `findBackgroundColor`; remove `selectDistributedColor` |
| `pkg/config/config.go` | Modify | Add `"milspec"` branch to `parseColorRatios` |
| `internal/generator/pat5.go` | Modify | Fix `Generate()` to use MARPAT ratios + 5-layer pipeline; remove dead code; remove dead methods from `pg.hash`, `pg.lerp`, `pg.perlinNoise` |
| `internal/generator/generator_test.go` | Modify | Add `TestGenerateFromImage_ErrorBeforeSort` |
| `internal/generator/utils_test.go` | Create | Tests for new noise helpers |
| `internal/generator/pat5_test.go` | Modify | Remove tests for dead template methods; add `TestPat5Generator_DefaultMARPATRatios` |
| `pkg/config/config_test.go` | Modify | Add `"milspec"` preset test cases |

---

## Task 1 — Fix `GenerateFromImage` error-ordering bug

**Files:**
- Modify: `internal/generator/generator.go:106-136`
- Modify: `internal/generator/generator_test.go`

### Background

`GenerateFromImage` currently calls `sortColors(mainColors)` and hex-conversion loops **before** the `if err != nil` guard. If `gen.Generate()` returns an error the unnecessary work runs on a nil/empty slice. More critically, this is a logic bug: the error must be checked immediately after the call that produces it.

- [ ] **Step 1: Run existing tests to confirm baseline**

```bash
cd /home/mark/Code/gocamo && go test ./... -v -run TestGenerateFromImage 2>&1 | tail -20
```

Expected: all TestGenerateFromImage tests pass.

- [ ] **Step 2: Read the current function (lines 106-136 of generator.go) to confirm the ordering**

Confirm that the call order is:
1. `gen.Generate()` → assigns `img, mainColors, err`
2. `sortColors(mainColors)` ← runs unconditionally
3. hex-conversion loop ← runs unconditionally
4. `if err != nil { return ... }` ← error check is late

- [ ] **Step 3: Fix the error-check ordering in `generator.go`**

Replace the body of `GenerateFromImage` (from the `gen.Generate` call through the first `saveImageToFile` call) with:

```go
func GenerateFromImage(ctx context.Context, cfg *config.Config, imagePath string, index int, outputPath string) error {
	gen := &ImageGenerator{InputFile: imagePath}

	img, mainColors, err := gen.Generate(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("error generating pattern from image %s: %w", imagePath, err)
	}

	// Sort the main colors
	sortColors(mainColors)

	// Convert main colors to hex for filename
	hexColors := make([]string, len(mainColors))
	for i, c := range mainColors {
		hexColors[i] = fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
	}
	colorCodesStr := strings.Join(hexColors, "_")

	baseName := filepath.Base(imagePath)
	fileName := fmt.Sprintf("gocamo_from_image_%s_%03d_%s_k%d_w%dx%d.png",
		strings.TrimSuffix(baseName, filepath.Ext(baseName)),
		index, colorCodesStr, cfg.KValue, cfg.Width, cfg.Height)
	filePath := filepath.Join(outputPath, fileName)

	if err := saveImageToFile(img, filePath); err != nil {
		return fmt.Errorf("error saving image %s: %w", filePath, err)
	}

	return nil
}
```

- [ ] **Step 4: Run the tests**

```bash
go test ./internal/generator/... -v -run TestGenerateFromImage 2>&1
```

Expected: all TestGenerateFromImage tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generator/generator.go
git commit -m "fix: check error from Generate() before processing mainColors in GenerateFromImage"
```

---

## Task 2 — Add package-level Perlin noise helpers to `utils.go`

**Files:**
- Modify: `internal/generator/utils.go`
- Create: `internal/generator/utils_noise_test.go` (new file; keeps noise tests separate from existing utils tests)

### Background

`pat5.go` already contains working `perlinNoise`, `hash`, and `lerp` methods on `Pat5Generator`. They need to become package-level functions so pat2 can use them too, and so they are testable independently. A `fractalNoise` helper (multi-octave fBm) is also added for use in the pat5 pipeline.

- [ ] **Step 1: Add the four helpers at the bottom of `utils.go`**

Append to `/home/mark/Code/gocamo/internal/generator/utils.go`:

```go
// noiseHash returns a pseudo-random float in [-1, 1] for integer grid coordinates.
func noiseHash(x, y int) float64 {
	n := x + y*57
	n = (n << 13) ^ n
	return 1.0 - float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff)/1073741824.0
}

// lerp linearly interpolates between a and b by t ∈ [0,1].
func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

// perlinNoise returns a smoothed noise value in approximately [-1, 1] for (x, y).
func perlinNoise(x, y float64) float64 {
	xi := int(x) & 255
	yi := int(y) & 255
	xf := x - float64(int(x))
	yf := y - float64(int(y))

	// Smoothstep fade
	u := xf * xf * (3.0 - 2.0*xf)
	v := yf * yf * (3.0 - 2.0*yf)

	aa := noiseHash(xi, yi)
	ab := noiseHash(xi, yi+1)
	ba := noiseHash(xi+1, yi)
	bb := noiseHash(xi+1, yi+1)

	x1 := lerp(aa, ba, u)
	x2 := lerp(ab, bb, u)
	return lerp(x1, x2, v)
}

// fractalNoise combines octaves of Perlin noise (fBm) and returns a value in [0, 1].
func fractalNoise(x, y float64, octaves int) float64 {
	noise := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxValue := 0.0

	for i := 0; i < octaves; i++ {
		noise += perlinNoise(x*frequency, y*frequency) * amplitude
		maxValue += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	return (noise/maxValue + 1.0) / 2.0
}
```

- [ ] **Step 2: Write the test file**

Create `/home/mark/Code/gocamo/internal/generator/utils_noise_test.go`:

```go
package generator

import (
	"math"
	"testing"
)

func TestPerlinNoise_Range(t *testing.T) {
	// Values should be in roughly [-1, 1]
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			v := perlinNoise(float64(x)*0.1, float64(y)*0.1)
			if v < -1.5 || v > 1.5 {
				t.Errorf("perlinNoise(%d*0.1, %d*0.1) = %f, outside expected range [-1.5, 1.5]", x, y, v)
			}
		}
	}
}

func TestPerlinNoise_SpatialCoherence(t *testing.T) {
	// Nearby inputs should produce similar outputs (spatial coherence).
	v1 := perlinNoise(1.0, 1.0)
	v2 := perlinNoise(1.01, 1.0)
	if math.Abs(v1-v2) > 0.1 {
		t.Errorf("perlinNoise not spatially coherent: v(1.0,1.0)=%f, v(1.01,1.0)=%f, diff=%f",
			v1, v2, math.Abs(v1-v2))
	}
}

func TestFractalNoise_Range(t *testing.T) {
	// fractalNoise must produce values in [0, 1]
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			v := fractalNoise(float64(x)*0.05, float64(y)*0.05, 4)
			if v < 0 || v > 1 {
				t.Errorf("fractalNoise at (%d, %d) = %f, outside [0, 1]", x, y, v)
			}
		}
	}
}

func TestLerp(t *testing.T) {
	if got := lerp(0, 10, 0.5); got != 5 {
		t.Errorf("lerp(0, 10, 0.5) = %f, want 5", got)
	}
	if got := lerp(0, 10, 0); got != 0 {
		t.Errorf("lerp(0, 10, 0) = %f, want 0", got)
	}
	if got := lerp(0, 10, 1); got != 10 {
		t.Errorf("lerp(0, 10, 1) = %f, want 10", got)
	}
}
```

- [ ] **Step 3: Run the new tests**

```bash
go test ./internal/generator/... -v -run "TestPerlinNoise|TestFractalNoise|TestLerp" 2>&1
```

Expected: 4 tests pass.

- [ ] **Step 4: Ensure all existing tests still pass**

```bash
go test ./internal/generator/... 2>&1
```

Expected: all pass, no compilation errors.

- [ ] **Step 5: Commit**

```bash
git add internal/generator/utils.go internal/generator/utils_noise_test.go
git commit -m "feat: add package-level perlinNoise, fractalNoise, lerp helpers to generator utils"
```

---

## Task 3 — Fix pat2: replace incoherent hash noise, remove dead methods

**Files:**
- Modify: `internal/generator/pat2.go`

### Background

`Pat2Generator.simpleNoise` uses `math.Sin(x*12.9898+y*78.233)` — a hash with **no spatial coherence**. Adjacent inputs produce completely unrelated outputs, giving blob edges a "salt-and-pepper" fraying rather than smooth organic variation. Replacing it with the package-level `perlinNoise` (from Task 2) fixes this.

Dead methods to remove: `findLightestColor`, `selectDarkerColor`, `drawAngularElement`. None are called anywhere in the file.

- [ ] **Step 1: Confirm the dead methods are unreferenced**

```bash
grep -n "findLightestColor\|selectDarkerColor\|drawAngularElement" /home/mark/Code/gocamo/internal/generator/pat2.go
```

Expected: lines only inside the function definitions themselves — no call sites.

- [ ] **Step 2: Replace `simpleNoise` in `pat2.go`**

Find this method (around line 271):

```go
// simpleNoise generates simple noise for organic shapes
func (pg *Pat2Generator) simpleNoise(x, y float64) float64 {
	// Simple pseudo-Perlin noise
	n := math.Sin(x*12.9898+y*78.233) * 43758.5453
	return n - math.Floor(n) - 0.5
}
```

Replace with:

```go
// simpleNoise returns spatially-coherent noise in approximately [-0.5, 0.5].
func (pg *Pat2Generator) simpleNoise(x, y float64) float64 {
	return perlinNoise(x, y) * 0.5
}
```

- [ ] **Step 3: Remove dead methods from `pat2.go`**

Delete the following three complete method bodies (they have no callers):

1. `findLightestColor` (~lines 86-99)
2. `selectDarkerColor` (~lines 254-268)
3. `drawAngularElement` (~lines 213-231)

After deletion, verify `math` is still needed (it is: `math.Sqrt`, `math.Pow`, `math.Max` remain in active methods). Also verify `math.Floor` is no longer needed — remove it from the import if `math.Floor` was the only usage. If `math` is still used elsewhere, leave the import as-is.

- [ ] **Step 4: Build and test**

```bash
go build ./internal/generator/... 2>&1 && go test ./internal/generator/... 2>&1
```

Expected: compiles and all tests pass. (No pat2-specific unit tests exist; compilation success is the gate.)

- [ ] **Step 5: Generate a pat2 sample and view it**

```bash
mkdir -p /tmp/gocamo_pat2 && \
./gocamo -c "#5a6b3c,#d4c5a7,#4a3f2a,#2d362a" -t pat2 -w 900 -h 900 -o /tmp/gocamo_pat2
```

View `/tmp/gocamo_pat2/*.png`. Blob edges should be smoother and more organic compared to before — less frayed/noisy appearance.

- [ ] **Step 6: Commit**

```bash
git add internal/generator/pat2.go
git commit -m "fix: replace incoherent hash noise with Perlin noise in pat2, remove dead methods"
```

---

## Task 4 — Fix pat1: use lightest colour as background, remove dead method

**Files:**
- Modify: `internal/generator/pat1.go`

### Background

`Pat1Generator` randomly picks any colour as background (`rand.Intn(len(colors))`). This can make the darkest colour the background, which inverts the visual layering. `findBackgroundColor` already exists on the struct — it selects the brightest colour, which is correct for woodland patterns where the lighter base dominates. `selectDistributedColor` is also defined but never called anywhere.

- [ ] **Step 1: Confirm `selectDistributedColor` is unreferenced**

```bash
grep -rn "selectDistributedColor" /home/mark/Code/gocamo/internal/generator/
```

Expected: only the function definition itself (in pat1.go).

- [ ] **Step 2: Fix background colour selection in `pat1.go`**

In `Generate()` (around line 35), find:

```go
	// Initialize with random background color (any color can be background)
	backgroundColorIndex := rand.Intn(len(colors))
```

Replace with:

```go
	// Use the lightest color as background — typical in woodland military patterns
	backgroundColorIndex := pg.findBackgroundColor(colors)
```

- [ ] **Step 3: Remove the dead `selectDistributedColor` method**

Delete this entire method (around lines 204-219):

```go
// selectDistributedColor ensures all colors get used prominently like real SPLAT
func (pg *Pat1Generator) selectDistributedColor(colors []color.RGBA, excludeIndex int, seed int) int {
	...
}
```

- [ ] **Step 4: Build and test**

```bash
go build ./internal/generator/... 2>&1 && go test ./internal/generator/... 2>&1
```

Expected: compiles and all tests pass.

- [ ] **Step 5: Generate pat1 samples and compare**

```bash
mkdir -p /tmp/gocamo_pat1 && \
./gocamo -c "#5a6b3c,#d4c5a7,#4a3f2a,#2d362a" -t pat1 -w 900 -h 900 -o /tmp/gocamo_pat1
```

View the output. The tan/khaki (`#d4c5a7`) should now consistently be the background base colour, with darker shapes layered on top — matching real woodland patterns.

- [ ] **Step 6: Commit**

```bash
git add internal/generator/pat1.go
git commit -m "fix: use lightest colour as background in pat1, remove unused selectDistributedColor"
```

---

## Task 5 — Add `milspec` colour-ratio preset to config

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/config_test.go`

### Background

The default ratio is equal (1/N per colour). Real milspec patterns use asymmetric ratios: ~45% base / 30% secondary / 15% tertiary / 10% accent. Adding a named `"milspec"` preset lets users opt into authentic distribution with `-r milspec`.

- [ ] **Step 1: Add the failing test first**

In `pkg/config/config_test.go`, add the following test function after the existing `TestParseColorRatios` tests:

```go
func TestParseColorRatios_Milspec(t *testing.T) {
	tests := []struct {
		numColors int
		want      []float64
	}{
		{4, []float64{0.45, 0.30, 0.15, 0.10}},
		{3, []float64{0.45 / 0.90, 0.30 / 0.90, 0.15 / 0.90}}, // normalised 3-colour
		{2, []float64{0.45 / 0.75, 0.30 / 0.75}},               // normalised 2-colour
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d-color", tt.numColors), func(t *testing.T) {
			ratios, err := parseColorRatios("milspec", tt.numColors)
			if err != nil {
				t.Fatalf("parseColorRatios(milspec, %d) error: %v", tt.numColors, err)
			}
			if len(ratios) != tt.numColors {
				t.Fatalf("got %d ratios, want %d", len(ratios), tt.numColors)
			}

			sum := 0.0
			for _, r := range ratios {
				sum += r
			}
			if math.Abs(sum-1.0) > 0.001 {
				t.Errorf("ratios sum = %f, want 1.0", sum)
			}

			for i, want := range tt.want {
				if math.Abs(ratios[i]-want) > 0.001 {
					t.Errorf("ratio[%d] = %f, want %f", i, ratios[i], want)
				}
			}
		})
	}
}
```

Add `"math"` to the imports block at the top of `config_test.go` if not already present.

- [ ] **Step 2: Run the failing test**

```bash
go test ./pkg/config/... -v -run TestParseColorRatios_Milspec 2>&1
```

Expected: FAIL — `"milspec"` is not yet handled.

- [ ] **Step 3: Add `"milspec"` branch to `parseColorRatios` in `config.go`**

In `parseColorRatios` (around line 181), after the `if ratiosString == "random"` block and before the custom-ratio parsing section, insert:

```go
	if ratiosString == "milspec" {
		milspecBase := []float64{0.45, 0.30, 0.15, 0.10}
		ratios := make([]float64, numColors)
		sum := 0.0
		for i := 0; i < numColors; i++ {
			ratios[i] = milspecBase[i%len(milspecBase)]
			sum += ratios[i]
		}
		for i := range ratios {
			ratios[i] /= sum
		}
		return ratios, nil
	}
```

Also update the `-r` flag description in `ParseFlags` (around line 301) to document the new preset:

```go
	flag.StringVar(&cfg.RatiosString, "r", "", "Color ratios: 'random', 'milspec' (45/30/15/10%), or integers like '2,1,3' (default: equal)")
```

- [ ] **Step 4: Run the tests**

```bash
go test ./pkg/config/... -v -run TestParseColorRatios_Milspec 2>&1
```

Expected: all milspec sub-tests pass.

- [ ] **Step 5: Run the full test suite**

```bash
go test ./... 2>&1
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add pkg/config/config.go pkg/config/config_test.go
git commit -m "feat: add milspec colour-ratio preset (45/30/15/10%) to config"
```

---

## Task 6 — Fix pat5: use MARPAT ratios by default

**Files:**
- Modify: `internal/generator/pat5.go`
- Modify: `internal/generator/pat5_test.go`

### Background

`Pat5Generator.getMARPATColorRatios()` exists and returns the correct 45/30/15/10 distribution. But `Generate()` uses `cfg.ColorRatios` which defaults to equal ratios (set by `SetColorRatios`). The MARPAT ratios only activate when `len(ratios) != len(colors)` — a condition that never occurs in practice. This means pat5 always produces equal colour distribution despite being explicitly MARPAT-inspired.

Fix: when `cfg.RatiosString == ""` (user did not specify ratios), `Generate()` overrides with `getMARPATColorRatios()`.

- [ ] **Step 1: Add the failing test**

Add to `pat5_test.go`:

```go
func TestPat5Generator_DefaultMARPATRatios(t *testing.T) {
	// When no explicit ratios are set (RatiosString == ""), pat5 must use MARPAT
	// ratios internally, so colour 0 (the base) dominates.
	ctx := context.Background()
	cfg := &config.Config{
		Width:         200,
		Height:        200,
		BasePixelSize: 4,
		PatternType:   "pat5",
		RatiosString:  "",
		ColorRatios:   []float64{0.25, 0.25, 0.25, 0.25}, // equal — set by SetColorRatios default
	}

	colors := []color.RGBA{
		{R: 90, G: 107, B: 60, A: 255},  // index 0 — base green (should dominate ~45%)
		{R: 212, G: 197, B: 167, A: 255}, // index 1 — tan
		{R: 74, G: 63, B: 42, A: 255},   // index 2 — brown
		{R: 45, G: 54, B: 42, A: 255},   // index 3 — dark green (accent ~10%)
	}

	gen := &Pat5Generator{}
	img, err := gen.Generate(ctx, cfg, colors)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if img == nil {
		t.Fatal("Generate returned nil image")
	}

	// Count colour pixels in the output
	bounds := img.Bounds()
	counts := make([]int, len(colors))
	total := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := img.At(x, y)
			r, g, b, _ := px.RGBA()
			rr, gg, bb := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			for i, c := range colors {
				if c.R == rr && c.G == gg && c.B == bb {
					counts[i]++
					total++
					break
				}
			}
		}
	}

	if total == 0 {
		t.Skip("No exact colour matches found (blending may be active)")
	}

	baseRatio := float64(counts[0]) / float64(total)
	accentRatio := float64(counts[3]) / float64(total)

	// Base colour should be significantly more prominent than accent colour.
	// With equal ratios both would be ~25%; with MARPAT ratios base ~45%, accent ~10%.
	if baseRatio <= accentRatio*1.5 {
		t.Errorf("Base colour ratio %f is not sufficiently dominant over accent %f — MARPAT ratios may not be applied",
			baseRatio, accentRatio)
	}
}
```

- [ ] **Step 2: Run the failing test**

```bash
go test ./internal/generator/... -v -run TestPat5Generator_DefaultMARPATRatios 2>&1
```

Expected: FAIL — ratios are currently equal, so base colour is not dominant.

- [ ] **Step 3: Add the MARPAT ratio override to `pat5.go` `Generate()`**

At the top of `Generate()`, after `adjustedBasePixelSize` and before any grid creation, insert:

```go
	// When the user has not specified explicit ratios, use authentic MARPAT distribution
	// rather than the equal default from SetColorRatios.
	ratios := cfg.ColorRatios
	if cfg.RatiosString == "" {
		ratios = pg.getMARPATColorRatios(len(colors))
	}
```

Then update the call to `initializeDigitalPixelBase` (currently in `Generate()`) to pass `ratios` instead of `cfg.ColorRatios`. The next task will replace this call entirely; for now this step only adds the ratio override so the test can pass.

In `Generate()`, find:

```go
	pg.initializeDigitalPixelBase(grid, gridWidth, gridHeight, colors, cfg.ColorRatios)
```

Change to:

```go
	pg.initializeDigitalPixelBase(grid, gridWidth, gridHeight, colors, ratios)
```

- [ ] **Step 4: Run the test**

```bash
go test ./internal/generator/... -v -run TestPat5Generator_DefaultMARPATRatios 2>&1
```

Expected: PASS.

- [ ] **Step 5: Run the full test suite**

```bash
go test ./... 2>&1
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add internal/generator/pat5.go internal/generator/pat5_test.go
git commit -m "fix: pat5 now uses MARPAT colour ratios by default when no explicit ratios are set"
```

---

## Task 7 — Wire up pat5's 5-layer pipeline

**Files:**
- Modify: `internal/generator/pat5.go`
- Modify: `internal/generator/pat5_test.go`

### Background

`pat5.go` contains a complete, well-written IFS fractal system and rectangle clustering system that were never called from `Generate()`. The current `Generate()` only runs three lightweight steps. The full pipeline is:

1. **`generateFractalLayer`** — IFS chaos game for large-scale macro structure (sparse, used as colour hint)
2. **`initializeMARPATGrid`** — seed the grid with MARPAT-weighted colour distribution
3. **`applyRectangleClustering`** — overlay rectangle-shaped colour blocks guided by the fractal layer
4. **`applyMARPATPixelClustering`** — add small-scale 1×2/2×1/2×2 digital pixel blocks
5. **`addDigitalTextureNoise`** — scatter 5% single-pixel noise for authentic MARPAT texture

- [ ] **Step 1: Verify the signatures of the methods to wire up**

Run this grep to confirm the function signatures match the planned calls:

```bash
grep -n "^func (pg \*Pat5Generator)" /home/mark/Code/gocamo/internal/generator/pat5.go | \
  grep -E "generateFractalLayer|initializeMARPATGrid|applyRectangleClustering|applyMARPATPixelClustering|addDigitalTextureNoise|renderGrid|getMARPATColorRatios"
```

Expected signatures:
- `generateFractalLayer(width, height int, colors []color.RGBA) [][]int`
- `initializeMARPATGrid(width, height int, colorRatios []float64, numColors int) [][]int`
- `applyRectangleClustering(grid, fractalLayer [][]int, width, height, basePixelSize int)`
- `applyMARPATPixelClustering(grid [][]int, width, height, numColors int)`
- `addDigitalTextureNoise(grid [][]int, width, height, numColors int)`
- `renderGrid(img *image.NRGBA, grid [][]int, colors []color.RGBA, pixelSize int)`
- `getMARPATColorRatios(numColors int) []float64`

- [ ] **Step 2: Replace `Generate()` in `pat5.go` with the 5-layer pipeline**

Find the entire existing `Generate()` function (from `func (pg *Pat5Generator) Generate(` through its closing `}`) and replace it entirely with:

```go
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
```

- [ ] **Step 3: Build**

```bash
go build ./internal/generator/... 2>&1
```

Expected: compiles with no errors. (Some methods are now unreachable — that is expected and will be fixed in Task 8.)

- [ ] **Step 4: Run all generator tests**

```bash
go test ./internal/generator/... -v 2>&1
```

Expected: all pass.

> Note: `TestPat5Generator_CreateRectangularTemplates` and `TestPat5Generator_CreateTemplate` test methods that will be removed in Task 8. They should still pass at this stage since those methods still exist.

- [ ] **Step 5: Generate a pat5 sample and view**

```bash
mkdir -p /tmp/gocamo_pat5_new && \
./gocamo -c "#5a6b3c,#d4c5a7,#4a3f2a,#2d362a" -t pat5 -w 900 -h 900 -o /tmp/gocamo_pat5_new
```

View the output. Compared to the old pat5: should show stronger large-scale structure from the IFS fractal, more rectangular pixel groupings, and the base green should dominate (~45%) with the tan and dark tones at lower proportions.

- [ ] **Step 6: Commit**

```bash
git add internal/generator/pat5.go
git commit -m "feat: wire up pat5 5-layer pipeline (IFS fractal + MARPAT grid + rectangle clustering)"
```

---

## Task 8 — Remove pat5 dead code

**Files:**
- Modify: `internal/generator/pat5.go`
- Modify: `internal/generator/pat5_test.go`

### Background

After Task 7, the following methods are no longer reachable from `Generate()` or any other active code path. Keeping them would inflate the file from ~1,500 lines to ~1,500 lines of confusion. The Poisson disk / organic cluster growth system (`generateClusterSeeds`, `growOrganicClusters`, `calculateLocalDensity`, etc.) has an O(n²) performance bug in `calculateLocalDensity` (linear scan of all cluster pixels per candidate pixel) that makes it unsuitable for production without a spatial hash rewrite — so it is removed rather than wired up.

Methods to delete from `pat5.go`:
- `initializeDigitalPixelBase` (replaced by `initializeMARPATGrid` in pipeline)
- `digitalNoise` (only used by `initializeDigitalPixelBase`)
- `noiseToColorIndex` (only used by `initializeDigitalPixelBase`)
- `applyCellularAutomataClustering` (more aggressive version, not in pipeline)
- `addRectangularPixelStructure` (replaced by `applyRectangleClustering`)
- `addMultiScaleRefinement` (replaced by `addDigitalTextureNoise`)
- `blendFractalWithTemplates`
- `createPixelBlocks`
- `selectWeightedPixelBlock`
- `generateBlockRegions`
- `hasSignificantOverlap`
- `markRegionUsed`
- `renderPixelBlocks`
- `initializeImageWithBaseColor`
- `generateClusterSeeds` (Poisson disk — O(n²) bug in `calculateLocalDensity`)
- `getRandomGrowthDirections`
- `isValidSeedPosition`
- `markGridOccupied`
- `growOrganicClusters`
- `calculateLocalDensity` (O(n²): linear scan of all cluster pixels per growth candidate)
- `getClusterBlockType`
- `addMultiScaleDetails`
- `renderOrganicClusters`
- `createRectangularTemplates` (template matrix system, different from rectangle clustering)
- `createTemplate`
- `applyTemplatesGreedy`
- `canPlaceTemplate`
- `placeTemplateWithEdgeAwareness`
- `countSimilarNeighbors`
- `getDominantColorInArea` (distinct from `getDominantColor` which stays)
- `generateFractalNoise` (superseded by package-level `fractalNoise` in utils.go)
- `perlinNoise` (method on Pat5Generator — superseded by package-level `perlinNoise`)
- `hash` (method on Pat5Generator — superseded by package-level `noiseHash`)
- `lerp` (method on Pat5Generator — superseded by package-level `lerp`)

Struct types that become orphaned (no active methods use them) — delete:
- `ClusterSeed`
- `OrganicCluster`
- `PixelBlock`
- `BlockRegion`

Struct types to **keep** (still used by active methods):
- `Template` — used by `createRectangularTemplates`... wait, that's being deleted. Check: `Template` is used by `createRectangularTemplates` and `createTemplate` and `applyTemplatesGreedy` — all deleted. Delete `Template` too.
- `FractalParams` — used by `generateFractalLayer` which is ACTIVE. **Keep.**
- `RectangleCluster` — used by `createRectangleClusters` and `applyRectangleClustering` which are ACTIVE. **Keep.**
- `Pixel` — used by `growOrganicClusters` which is being deleted... but also potentially in directions slice inside `growOrganicClusters`. Since `growOrganicClusters` is deleted, check if `Pixel` is used anywhere else. After deletion, `Pixel` is unused — **delete.**

- [ ] **Step 1: Confirm which methods are unreachable**

```bash
grep -c "^func (pg \*Pat5Generator)" /home/mark/Code/gocamo/internal/generator/pat5.go
```

Note the total count. After cleanup it should be approximately 14 (the active pipeline methods plus helpers).

- [ ] **Step 2: Remove the dead tests in `pat5_test.go`**

Delete these two test functions from `pat5_test.go` (they test methods that are about to be deleted):

1. `TestPat5Generator_CreateRectangularTemplates`
2. `TestPat5Generator_CreateTemplate`

- [ ] **Step 3: Run the remaining tests to confirm they pass before the deletions**

```bash
go test ./internal/generator/... -v -run TestPat5 2>&1
```

Expected: all remaining pat5 tests pass (the two deleted tests are gone).

- [ ] **Step 4: Delete dead struct types from `pat5.go`**

Remove the following struct type declarations (they will no longer be referenced):
- `Template`
- `PixelBlock`
- `BlockRegion`
- `ClusterSeed`
- `OrganicCluster`
- `Pixel`

Keep: `FractalParams`, `RectangleCluster`.

- [ ] **Step 5: Delete the dead methods from `pat5.go`**

Remove all methods listed in the Background section of this task. Work methodically — delete one block at a time, running `go build ./internal/generator/...` after each deletion to confirm nothing broke.

After all deletions, run:

```bash
go build ./internal/generator/... 2>&1
```

Expected: compiles with no errors and no "declared but not used" or "undefined" errors.

- [ ] **Step 6: Verify `math` and `fmt` imports are still needed in `pat5.go`**

```bash
grep -n "math\." /home/mark/Code/gocamo/internal/generator/pat5.go | head -10
grep -n "fmt\." /home/mark/Code/gocamo/internal/generator/pat5.go | head -10
```

`math` is still needed for `generateFractalLayer` (`math.Cos`, `math.Sin`, `math.Pi`). `fmt` is used by `hasSignificantOverlap` and `markRegionUsed` — both deleted. If `fmt` no longer appears, remove it from the import block.

- [ ] **Step 7: Run the full test suite**

```bash
go test ./... 2>&1
```

Expected: all tests pass.

- [ ] **Step 8: Check final line count**

```bash
wc -l /home/mark/Code/gocamo/internal/generator/pat5.go
```

Expected: roughly 300-400 lines (down from ~1,500). The file should contain only active structs and methods.

- [ ] **Step 9: Generate all patterns for a final visual check**

```bash
mkdir -p /tmp/gocamo_final && \
./gocamo -c "#5a6b3c,#d4c5a7,#4a3f2a,#2d362a" -t all -w 900 -h 900 -o /tmp/gocamo_final && \
./gocamo -c "#5a6b3c,#d4c5a7,#4a3f2a,#2d362a" -t all -w 900 -h 900 -r milspec -o /tmp/gocamo_final
```

View both sets. The `-r milspec` set should show more dominant base-colour coverage (~45%) across all patterns.

- [ ] **Step 10: Commit**

```bash
git add internal/generator/pat5.go internal/generator/pat5_test.go
git commit -m "refactor: remove ~1100 lines of dead code from pat5; keep 5-layer pipeline only"
```

---

## Task 9 — Integration smoke test and JSON palette run

**Files:** None modified — validation only.

- [ ] **Step 1: Run the complete test suite**

```bash
go test ./... -count=1 2>&1
```

Expected: all pass, zero failures.

- [ ] **Step 2: Generate all pattern types with the MARPAT woodland palette**

```bash
mkdir -p /tmp/gocamo_smoke && \
./gocamo -j colors.json -t all -w 900 -h 900 -o /tmp/gocamo_smoke
```

Expected: generates 5 × (number of palettes in colors.json) images with no errors. Runtime should remain well under 10 seconds.

- [ ] **Step 3: Test the milspec ratio preset end-to-end**

```bash
./gocamo -c "#5a6b3c,#d4c5a7,#4a3f2a,#2d362a" -t all -r milspec -w 900 -h 900 -o /tmp/gocamo_smoke
```

Expected: 5 images with distinctly asymmetric colour coverage — base green dominant.

- [ ] **Step 4: Test the image-based generator (exercises the bug fix from Task 1)**

```bash
./gocamo -i input -w 900 -h 900 -o /tmp/gocamo_smoke
```

Expected: generates from images in `input/` directory with no panics or "generating pattern" errors.

- [ ] **Step 5: Final commit (update README if applicable)**

If the new `-r milspec` flag and improved pattern descriptions warrant a README update, do so now:

```bash
git add README.md   # only if changed
git commit -m "docs: document milspec ratio preset and updated pattern descriptions"
```

---

## Summary of Changes

| Component | Change |
|---|---|
| `generator.go` | Error check moved before post-processing |
| `utils.go` | +`perlinNoise`, `fractalNoise`, `lerp`, `noiseHash` |
| `pat1.go` | Background uses lightest colour; dead method removed |
| `pat2.go` | `simpleNoise` → Perlin; 3 dead methods removed |
| `config.go` | `"milspec"` ratio preset (45/30/15/10%) |
| `pat5.go` | MARPAT ratios by default; 5-layer pipeline wired; ~1,100 lines of dead code removed |
