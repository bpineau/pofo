package metrics

import (
	"math"
	"testing"
)

// threeAssets is a hand-built sample on one calendar: b leans with a, c
// against it.
var threeAssets = [][]float64{
	{0.010, -0.020, 0.015, 0.003, -0.007, 0.012},
	{0.020, -0.010, 0.010, 0.000, -0.012, 0.008},
	{-0.005, 0.010, 0.002, -0.004, 0.006, -0.009},
}

func TestCorrKnownValues(t *testing.T) {
	near(t, "perfect", Corr([]float64{1, 2, 3}, []float64{2, 4, 6}), 1, 1e-12)
	near(t, "inverse", Corr([]float64{1, 2, 3}, []float64{3, 2, 1}), -1, 1e-12)
	// x and x^2 on a symmetric support are uncorrelated.
	near(t, "orthogonal", Corr([]float64{-1, 0, 1}, []float64{1, 0, 1}), 0, 1e-12)
	for name, c := range map[string]float64{
		"constant":   Corr([]float64{1, 1, 1}, []float64{1, 2, 3}),
		"mismatched": Corr([]float64{1, 2, 3}, []float64{1, 2}),
		"one point":  Corr([]float64{1}, []float64{2}),
		"empty":      Corr(nil, nil),
	} {
		if c != 0 {
			t.Errorf("%s: Corr = %v, want 0", name, c)
		}
	}
}

func TestCorrelationMatrix(t *testing.T) {
	m := CorrelationMatrix(threeAssets)
	if len(m) != 3 {
		t.Fatalf("got %d rows, want 3", len(m))
	}
	for i := range m {
		if m[i][i] != 1 {
			t.Errorf("diagonal [%d][%d] = %v, want exactly 1", i, i, m[i][i])
		}
		for j := range m {
			if m[i][j] != m[j][i] {
				t.Errorf("not symmetric at [%d][%d]: %v vs %v", i, j, m[i][j], m[j][i])
			}
			if i != j && m[i][j] != Corr(threeAssets[i], threeAssets[j]) {
				t.Errorf("[%d][%d] = %v, Corr says %v", i, j, m[i][j], Corr(threeAssets[i], threeAssets[j]))
			}
		}
	}
	if !(m[0][1] > 0.5 && m[0][2] < -0.5) {
		t.Errorf("signs: corr(a,b) = %v, corr(a,c) = %v", m[0][1], m[0][2])
	}
}

func TestCorrelationMatrixConstantAsset(t *testing.T) {
	m := CorrelationMatrix([][]float64{{0.01, -0.01, 0.02}, {0, 0, 0}})
	if m[0][0] != 1 || m[1][1] != 0 || m[0][1] != 0 || m[1][0] != 0 {
		t.Errorf("constant asset: %v", m)
	}
}

func TestCovariance(t *testing.T) {
	cov := Covariance(threeAssets)
	corr := CorrelationMatrix(threeAssets)
	for i := range cov {
		sd := sampleStdev(threeAssets[i])
		near(t, "variance", cov[i][i], sd*sd, 1e-15)
		for j := range cov {
			if cov[i][j] != cov[j][i] {
				t.Errorf("not symmetric at [%d][%d]", i, j)
			}
			want := corr[i][j] * sd * sampleStdev(threeAssets[j])
			near(t, "cov = corr sd sd", cov[i][j], want, 1e-15)
		}
	}
}

func TestMatricesEdgeCases(t *testing.T) {
	if m := CorrelationMatrix(nil); len(m) != 0 {
		t.Errorf("CorrelationMatrix(nil) = %v", m)
	}
	if m := Covariance(nil); len(m) != 0 {
		t.Errorf("Covariance(nil) = %v", m)
	}
	one := Covariance([][]float64{{0.01}, {0.02}})
	for _, row := range one {
		for _, c := range row {
			if !math.IsNaN(c) {
				t.Errorf("one period: covariance %v, want NaN", c)
			}
		}
	}
	if m := CorrelationMatrix([][]float64{{0.01}, {0.02}}); m[0][0] != 0 || m[0][1] != 0 {
		t.Errorf("one period: correlation %v, want zeros", m)
	}
	for name, fn := range map[string]func([][]float64) [][]float64{
		"CorrelationMatrix": CorrelationMatrix, "Covariance": Covariance,
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: ragged input did not panic", name)
				}
			}()
			fn([][]float64{{0.01, 0.02}, {0.01}})
		}()
	}
}
