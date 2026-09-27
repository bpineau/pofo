package metrics

import (
	"math"
	"strings"
	"testing"
)

// TestRegressHandComputed is the textbook five-point case, every figure
// derivable on paper: slope Sxy/Sxx = 6/10, intercept 4 - 0.6*3, residuals
// -0.8 0.6 1.0 -0.6 -0.2 (SSR 2.4, s² = 2.4/3), SST 6.
func TestRegressHandComputed(t *testing.T) {
	r, err := Regress([]float64{2, 4, 5, 4, 5}, []float64{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatal(err)
	}
	if r.N != 5 || len(r.Betas) != 1 {
		t.Fatalf("N=%d, %d betas", r.N, len(r.Betas))
	}
	near(t, "beta", r.Betas[0].Value, 0.6, 1e-12)
	near(t, "alpha", r.Alpha.Value, 2.2, 1e-12)
	near(t, "R2", r.R2, 0.6, 1e-12)
	near(t, "ResidualSD", r.ResidualSD, math.Sqrt(0.8), 1e-12)
	near(t, "SE(beta)", r.Betas[0].SE, math.Sqrt(0.8/10), 1e-12)
	near(t, "SE(alpha)", r.Alpha.SE, math.Sqrt(0.8*(1.0/5+9.0/10)), 1e-12)
	near(t, "T(beta)", r.Betas[0].T, 0.6/math.Sqrt(0.08), 1e-9)
}

// TestRegressMatchesSlope pins the one-regressor fit to slope, the estimator
// Beta has always used.
func TestRegressMatchesSlope(t *testing.T) {
	x, y := make([]float64, 300), make([]float64, 300)
	for i := range x {
		x[i] = 0.01 * math.Sin(float64(i)*0.7)
		y[i] = 0.0003 + 1.3*x[i] + 0.004*math.Cos(float64(i)*1.9)
	}
	r, err := Regress(y, x)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "beta", r.Betas[0].Value, slope(x, y), 1e-12)
	near(t, "R2 = corr²", r.R2, math.Pow(Corr(x, y), 2), 1e-12)
}

// TestRegressExactRecovery fits noise-free data built from known
// coefficients: they must come back, with a perfect fit.
func TestRegressExactRecovery(t *testing.T) {
	n := 120
	y, x1, x2, x3 := make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)
	for i := range n {
		f := float64(i)
		x1[i] = 0.04 * math.Sin(f*0.37)
		x2[i] = 0.02*math.Cos(f*1.11) + 0.5*x1[i]
		x3[i] = 0.01 * math.Sin(f*2.3+1)
		y[i] = 0.002 + 0.8*x1[i] - 0.3*x2[i] + 1.7*x3[i]
	}
	r, err := Regress(y, x1, x2, x3)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "alpha", r.Alpha.Value, 0.002, 1e-14)
	for j, want := range []float64{0.8, -0.3, 1.7} {
		near(t, "beta", r.Betas[j].Value, want, 1e-12)
	}
	near(t, "R2", r.R2, 1, 1e-12)
	near(t, "ResidualSD", r.ResidualSD, 0, 1e-14)
}

// TestRegressConstantY: nothing to explain, yet a well-posed fit.
func TestRegressConstantY(t *testing.T) {
	r, err := Regress([]float64{3, 3, 3, 3}, []float64{1, 2, 4, 8})
	if err != nil {
		t.Fatal(err)
	}
	near(t, "beta", r.Betas[0].Value, 0, 1e-15)
	near(t, "alpha", r.Alpha.Value, 3, 1e-15)
	if !math.IsNaN(r.R2) {
		t.Errorf("R2 = %v, want NaN", r.R2)
	}
}

func TestRegressAnnualize(t *testing.T) {
	r := Regression{Alpha: Estimate{Value: 0.001}, ResidualSD: 0.02}
	near(t, "AnnualAlpha", r.AnnualAlpha(12), 0.012, 1e-15)
	near(t, "AnnualResidualVol", r.AnnualResidualVol(12), 0.02*math.Sqrt(12), 1e-15)
}

func TestRegressErrors(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	cases := []struct {
		name string
		y    []float64
		xs   [][]float64
		want string
	}{
		{"no regressor", x, nil, "no regressor"},
		{"length", x, [][]float64{{1, 2, 3}}, "regressor 0 has 3 values, y has 5"},
		{"too few", []float64{1, 2, 3}, [][]float64{{1, 2, 4}, {0, 1, 0}}, "3 observations for 2 regressors, at least 4"},
		{"NaN y", []float64{1, math.NaN(), 3, 4, 5}, [][]float64{x}, "y[1] is not finite"},
		{"Inf x", x, [][]float64{{1, 2, math.Inf(1), 4, 5}}, "regressor 0, value 2 is not finite"},
		{"constant", x, [][]float64{{2, 2, 2, 2, 2}}, "regressor 0 is constant"},
		{"collinear", x, [][]float64{x, {3, 5, 7, 9, 11}}, "regressor 1 is collinear"},
	}
	for _, c := range cases {
		_, err := Regress(c.y, c.xs...)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want it to contain %q", c.name, err, c.want)
		}
	}
}

func TestExtremeK(t *testing.T) {
	xs := []float64{0.02, -0.05, math.NaN(), 0.01, -0.05, 0.03, -0.01}
	eq := func(name string, got, want []int) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%s = %v, want %v", name, got, want)
			}
		}
	}
	eq("LowestK 3", LowestK(xs, 3), []int{1, 4, 6}) // the tie keeps its order
	eq("HighestK 2", HighestK(xs, 2), []int{5, 0})
	eq("LowestK all", LowestK(xs, 99), []int{1, 4, 6, 3, 0, 5}) // NaN never selected
	eq("LowestK 0", LowestK(xs, 0), nil)
	eq("HighestK NaN only", HighestK([]float64{math.NaN()}, 1), nil)
	if xs[1] != -0.05 || !math.IsNaN(xs[2]) {
		t.Errorf("xs was modified: %v", xs)
	}
}
