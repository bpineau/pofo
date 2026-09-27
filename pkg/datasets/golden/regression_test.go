package golden

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/metrics"
)

// The regression golden: metrics.Regress replayed on the Longley data set of
// the NIST Statistical Reference Datasets, the standard accuracy test of a
// least squares program. Data and certified values were FETCHED on
// 2026-09-27 from
//
//	https://www.itl.nist.gov/div898/strd/lls/data/LINKS/DATA/Longley.dat
//
// (Longley, J. W. (1967), "An Appraisal of Least Squares Programs for the
// Electronic Computer from the Viewpoint of the User", JASA 62, 819-841),
// every figure below transcribed from that file. NIST grades it "higher
// level of difficulty": six regressors so nearly collinear that a solver
// forming the normal equations loses most of its digits, which is what the
// QR factorization of the centered data exists to avoid.

// longley is the 16 observations: y, then x1 to x6.
var longley = [16][7]float64{
	{60323, 83.0, 234289, 2356, 1590, 107608, 1947},
	{61122, 88.5, 259426, 2325, 1456, 108632, 1948},
	{60171, 88.2, 258054, 3682, 1616, 109773, 1949},
	{61187, 89.5, 284599, 3351, 1650, 110929, 1950},
	{63221, 96.2, 328975, 2099, 3099, 112075, 1951},
	{63639, 98.1, 346999, 1932, 3594, 113270, 1952},
	{64989, 99.0, 365385, 1870, 3547, 115094, 1953},
	{63761, 100.0, 363112, 3578, 3350, 116219, 1954},
	{66019, 101.2, 397469, 2904, 3048, 117388, 1955},
	{67857, 104.6, 419180, 2822, 2857, 118734, 1956},
	{68169, 108.4, 442769, 2936, 2798, 120445, 1957},
	{66513, 110.8, 444546, 4681, 2637, 121950, 1958},
	{68655, 112.6, 482704, 3813, 2552, 123366, 1959},
	{69564, 114.2, 502601, 3931, 2514, 125368, 1960},
	{69331, 115.7, 518173, 4806, 2572, 127852, 1961},
	{70551, 116.9, 554894, 4007, 2827, 130081, 1962},
}

// longleyCertified is NIST's certified estimate and standard deviation of
// B0 (the intercept) to B6.
var longleyCertified = [7][2]float64{
	{-3482258.63459582, 890420.383607373},
	{15.0618722713733, 84.9149257747669},
	{-0.358191792925910e-01, 0.334910077722432e-01},
	{-2.02022980381683, 0.488399681651699},
	{-1.03322686717359, 0.214274163161675},
	{-0.511041056535807e-01, 0.226073200069370},
	{1829.15146461355, 455.478499142212},
}

func TestGoldenRegressionLongley(t *testing.T) {
	y := make([]float64, len(longley))
	xs := make([][]float64, 6)
	for j := range xs {
		xs[j] = make([]float64, len(longley))
	}
	for i, row := range longley {
		y[i] = row[0]
		for j := range xs {
			xs[j][i] = row[j+1]
		}
	}
	r, err := metrics.Regress(y, xs...)
	if err != nil {
		t.Fatal(err)
	}

	// Eleven significant digits on every figure (thirteen measured on
	// arm64, two kept as the margin a different floating-point unit may
	// take); the StRD grades a solver by the digits it gets right, and a
	// normal-equations solver in float64 keeps barely a handful on this set.
	const digits = 1e-11
	check := func(name string, got, want float64) {
		t.Helper()
		if rel := math.Abs(got-want) / math.Abs(want); !(rel <= digits) {
			t.Errorf("%s = %.15g, certified %.15g (relative error %.2g)", name, got, want, rel)
		}
	}
	check("B0", r.Alpha.Value, longleyCertified[0][0])
	check("SD(B0)", r.Alpha.SE, longleyCertified[0][1])
	for j, b := range r.Betas {
		check("B"+string(rune('1'+j)), b.Value, longleyCertified[j+1][0])
		check("SD(B"+string(rune('1'+j))+")", b.SE, longleyCertified[j+1][1])
	}
	check("residual standard deviation", r.ResidualSD, 304.854073561965)
	check("R-squared", r.R2, 0.995479004577296)
}
