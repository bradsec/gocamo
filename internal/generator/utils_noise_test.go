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
