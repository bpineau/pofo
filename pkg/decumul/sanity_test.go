package decumul

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestGeoMeanMatchesParametric(t *testing.T) {
	// Student-t at Mu arithmetic, Sigma vol: geometric ~ Mu - Sigma^2/2.
	src := scenario.ParametricSource{Mu: 0.05, Sigma: 0.11, Df: 5, Periods: 40}
	geo := GeoMean(src, 4000, rand.New(rand.NewPCG(1, 2)))
	want := 0.05 - 0.11*0.11/2
	if math.Abs(geo-want) > 0.005 {
		t.Errorf("geo mean = %.4f, want ~%.4f", geo, want)
	}
}

func TestPlausibility(t *testing.T) {
	if w := Plausibility(0.045, 0.033); len(w) != 0 {
		t.Errorf("sane calibration should not warn, got %v", w)
	}
	if w := Plausibility(-0.02, 0); len(w) == 0 {
		t.Error("a negative geometric mean must warn (the doom-loop guard)")
	}
	if w := Plausibility(0.12, 0); len(w) == 0 {
		t.Error("an implausibly rosy geometric mean must warn")
	}
	if w := Plausibility(0.045, 0.008); len(w) == 0 {
		t.Error("a 0.8% safe WR is below the literature range and must warn")
	}
}
