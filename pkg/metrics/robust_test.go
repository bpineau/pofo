package metrics

import (
	"math"
	"reflect"
	"testing"
)

func TestShrinkageCovariance(t *testing.T) {
	rows := [][]float64{{1, 2, 3}, {2, 4, 6}}
	for _, tc := range []struct {
		lambda float64
		off    float64
	}{{0, 2}, {.25, 1.5}, {1, 0}} {
		got, err := ShrinkageCovariance(rows, tc.lambda)
		if err != nil {
			t.Fatal(err)
		}
		want := [][]float64{{1, tc.off}, {tc.off, 4}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("shrinkage %g: got %v, want %v", tc.lambda, got, want)
		}
	}
	for _, bad := range [][][]float64{nil, {{1}}, {{1, 2}, {1}}, {{1, math.NaN()}}, {{1, math.Inf(1)}}} {
		if _, err := ShrinkageCovariance(bad, .5); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	for _, bad := range []float64{-.1, 1.1, math.NaN(), math.Inf(1)} {
		if _, err := ShrinkageCovariance(rows, bad); err == nil {
			t.Errorf("accepted shrinkage %v", bad)
		}
	}
}

func TestRegressHACHandCalculated(t *testing.T) {
	// x is centered, sum(x*x)=28, and e=(1,2,-3,-3,2,1) is
	// orthogonal to both x and the intercept. OLS is alpha=2, beta=.5.
	// HC0 variances are sum(e²)/36=7/9 and sum(x²e²)/28²=17/196.
	// At lag 1 Bartlett weight is 1/2: sum(e_t e_t-1)=1 and
	// sum(x_t x_t-1 e_t e_t-1)=-9. Variances become 29/36 and 59/784.
	// No degrees-of-freedom correction is applied.
	x := []float64{-3, -2, -1, 1, 2, 3}
	y := []float64{1.5, 3, -1.5, -.5, 5, 4.5}
	ordinary, err := Regress(y, x)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		lags   int
		va, vb float64
	}{{0, 7.0 / 9, 17.0 / 196}, {1, 29.0 / 36, 59.0 / 784}} {
		got, err := RegressHAC(y, tc.lags, x)
		if err != nil {
			t.Fatal(err)
		}
		if got.Alpha.Value != ordinary.Alpha.Value || got.Betas[0].Value != ordinary.Betas[0].Value || got.R2 != ordinary.R2 || got.ResidualSD != ordinary.ResidualSD {
			t.Fatal("HAC changed the OLS fit")
		}
		if math.Abs(got.Alpha.SE*got.Alpha.SE-tc.va) > 1e-12 || math.Abs(got.Betas[0].SE*got.Betas[0].SE-tc.vb) > 1e-12 {
			t.Fatalf("lags %d: SE alpha %g, beta %g", tc.lags, got.Alpha.SE, got.Betas[0].SE)
		}
		if math.Abs(got.Alpha.T-2/math.Sqrt(tc.va)) > 1e-12 || math.Abs(got.Betas[0].T-.5/math.Sqrt(tc.vb)) > 1e-12 {
			t.Fatal("incorrect HAC T statistic")
		}
	}
	// Shifting x leaves its slope unchanged but changes the intercept's
	// variance by shift² times slope variance (cross covariance here is 0).
	shifted := make([]float64, len(x))
	for i, v := range x {
		shifted[i] = v + 10
	}
	got, err := RegressHAC(y, 1, shifted)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Alpha.SE*got.Alpha.SE-(29.0/36+100*59.0/784)) > 1e-10 {
		t.Fatal("intercept did not account for centering")
	}
}

func TestRegressHACInvalid(t *testing.T) {
	y, x := []float64{1, 2, 3, 4}, []float64{1, 2, 4, 8}
	for _, lags := range []int{-1, 4} {
		if _, err := RegressHAC(y, lags, x); err == nil {
			t.Fatal("accepted invalid lags")
		}
	}
	for _, bad := range [][]float64{{1, 2}, {1, 1, 1, 1}, {1, 2, math.NaN(), 4}} {
		if _, err := RegressHAC(y, 0, bad); err == nil {
			t.Fatal("accepted invalid regressor")
		}
	}
	if _, err := RegressHAC(y, 0, x, x); err == nil {
		t.Fatal("accepted rank deficiency")
	}
	if _, err := RegressHAC(y, 0); err == nil {
		t.Fatal("accepted no regressors")
	}
}

func TestRegressHACMultipleRegressors(t *testing.T) {
	// Centered orthogonal x,z have sums of squares 28 and 4.
	// Residuals (1,1,-2,-2,1,1) are orthogonal to the intercept and both.
	// At Bartlett lag 1, the residual-score sums of squares plus lag
	// products are 14 for the intercept, 34 for x and 2 for z.
	x := []float64{-3, -2, -1, 1, 2, 3}
	z := []float64{1, -1, 0, 0, -1, 1}
	e := []float64{1, 1, -2, -2, 1, 1}
	y := make([]float64, 6)
	for i := range y {
		y[i] = 2 + .5*x[i] - .7*z[i] + e[i]
	}
	got, err := RegressHAC(y, 1, x, z)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ se, variance float64 }{{got.Alpha.SE, 14.0 / 36}, {got.Betas[0].SE, 34.0 / 784}, {got.Betas[1].SE, 2.0 / 16}} {
		if math.Abs(tc.se*tc.se-tc.variance) > 1e-12 {
			t.Fatalf("SE %g, want variance %g", tc.se, tc.variance)
		}
	}
}
