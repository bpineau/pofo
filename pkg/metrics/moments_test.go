package metrics

import (
	"math"
	"testing"
)

func TestSkewnessSymmetric(t *testing.T) {
	// A symmetric set has zero skewness.
	near(t, "skew", Skewness([]float64{1, 2, 3, 4, 5}), 0, 1e-12)
}

func TestSkewnessRightTail(t *testing.T) {
	// Nine values at -1 and one at +9 around mean 1: m2=9, m3=72, skew=72/27.
	got := Skewness([]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 10})
	near(t, "skew", got, 72.0/27.0, 1e-9)
}

func TestExcessKurtosisUniform(t *testing.T) {
	// {1,2,3,4,5}: m2=2, m4=6.8, excess kurtosis = 6.8/4 - 3 = -1.3.
	near(t, "kurt", ExcessKurtosis([]float64{1, 2, 3, 4, 5}), -1.3, 1e-9)
}

func TestMomentsDegenerate(t *testing.T) {
	if !math.IsNaN(Skewness([]float64{42})) {
		t.Error("skewness of a single point should be NaN")
	}
	if !math.IsNaN(ExcessKurtosis([]float64{1, 1, 1})) {
		t.Error("excess kurtosis of a zero-variance set should be NaN")
	}
}

func TestComputeReportsMoments(t *testing.T) {
	// A long zig-zag has near-zero skew; just check the fields are populated
	// and consistent with the standalone functions over the same returns.
	values := []float64{100, 110, 99, 108.9, 98.01, 107.8}
	s, err := Compute(days(len(values)), values)
	if err != nil {
		t.Fatal(err)
	}
	r := Returns(values)
	near(t, "Skew", s.Skew, Skewness(r), 1e-12)
	near(t, "Kurtosis", s.Kurtosis, ExcessKurtosis(r), 1e-12)
}

func TestAutocorr(t *testing.T) {
	ac := Autocorr([]float64{1, -1, 1, -1, 1, -1, 1, -1, 1, -1}, 3)
	if len(ac) != 4 {
		t.Fatalf("len = %d, want 4", len(ac))
	}
	near(t, "lag0", ac[0], 1, 1e-12)
	// Perfect alternation: lag-1 autocorrelation is close to -1.
	if ac[1] > -0.8 {
		t.Errorf("lag1 = %v, want strongly negative", ac[1])
	}
	near(t, "lag2", ac[2], 0.8, 0.2) // even lag: positive
}

func TestHistogram(t *testing.T) {
	edges, counts := Histogram([]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 5)
	if len(edges) != 6 || len(counts) != 5 {
		t.Fatalf("edges=%d counts=%d", len(edges), len(counts))
	}
	near(t, "edge0", edges[0], 0, 1e-12)
	near(t, "edge5", edges[5], 10, 1e-12)
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 11 {
		t.Errorf("counts sum = %d, want 11", total)
	}
}
