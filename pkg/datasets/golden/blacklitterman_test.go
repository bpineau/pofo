package golden

import (
	"testing"

	"github.com/bpineau/pofo/pkg/optimize"
)

// The Black-Litterman golden: pkg/optimize's reverse optimization
// (ImpliedReturns) and Bayesian blend (Posterior) replayed on the two worked
// examples of the model's own literature. Both papers were FETCHED on
// 2026-08-27 and every figure below is transcribed from one of their printed
// tables, none from memory.
//
// Reference A, the seven-country equity example:
//
//	He, G. and Litterman, R. (1999), "The Intuition Behind Black-Litterman
//	Model Portfolios", Goldman Sachs Investment Management Research.
//	https://people.duke.edu/~charvey/Teaching/BA453_2005/GS_The_intuition_behind.pdf
//	Table 1 (annualized volatilities, market-capitalization weights and
//	equilibrium expected returns for the seven countries), Table 2 (the
//	correlations among the equity index returns) and footnote 3, which fixes
//	the risk aversion: "Throughout our examples, we use delta = 2.5 as the
//	risk aversion parameter representing the world average risk tolerance".
//
// Reference B, the eight-asset example carrying views:
//
//	Idzorek, T. (2005), "A Step-by-Step Guide to the Black-Litterman Model".
//	https://people.duke.edu/~charvey/Teaching/BA453_2006/Idzorek_onBL.pdf
//	Table 5 (the covariance matrix of excess returns), Table 2 final column
//	(the market-capitalization weights), Table 1 final column (the Implied
//	Equilibrium Return Vector) and its footnote, which fixes the risk
//	aversion at "approximately 3.07"; the three views of section 2.2 with
//	the market-capitalization view matrix P of formula 7; Tables 3a and 3b
//	(the two mini-portfolios of view 3 and their weighted equilibrium
//	returns); and Table 6 (the New Combined Return Vector and the New Weight
//	Vector).
//
// The confidence mapping is what ties this implementation to the papers.
// Omega here is diag((1/c - 1) * (P tau Sigma P')_kk), so a confidence of 0.5
// is exactly the Omega = diag(P tau Sigma P') of Idzorek's formula 8, which
// is what Table 6 is computed under. (The "Confidence of View" figures of
// 25 %, 50 % and 65 % printed beside the views belong to Idzorek's own
// ITERATIVE mapping, developed in his section 3.2 and deliberately not
// implemented here; they do not produce Table 6.) The scalar tau differs
// (0.025 in the paper, 0.05 here) and must not matter, since Omega is
// proportional to it: that the transcribed outputs come back anyway is the
// external half of the cancellation proof.
//
// Tolerances are the published precision. He and Litterman print one decimal
// of a percent and the inputs their table is rebuilt from are rounded too, so
// 0.05 points is as tight as that paper allows; Idzorek prints two decimals,
// where 0.03 points is.

// hlNames are the seven countries of He and Litterman (1999), in the order
// used by every vector below.
var hlNames = []string{"AUL", "CAN", "FRA", "GER", "JAP", "UKG", "USA"}

// TestGoldenBlackLittermanHeLitterman1999 replays the reverse optimization of
// the reference paper: the equilibrium expected returns of its Table 1 must
// come back out of ImpliedReturns fed with that same table's volatilities and
// market-capitalization weights, the correlations of Table 2, and delta = 2.5.
func TestGoldenBlackLittermanHeLitterman1999(t *testing.T) {
	// Table 2, correlations among the equity index returns (the paper prints
	// one triangle of this symmetric matrix).
	corr := [][]float64{
		{1.000, 0.488, 0.478, 0.515, 0.439, 0.512, 0.491},
		{0.488, 1.000, 0.664, 0.655, 0.310, 0.608, 0.779},
		{0.478, 0.664, 1.000, 0.861, 0.355, 0.783, 0.668},
		{0.515, 0.655, 0.861, 1.000, 0.354, 0.777, 0.653},
		{0.439, 0.310, 0.355, 0.354, 1.000, 0.405, 0.306},
		{0.512, 0.608, 0.783, 0.777, 0.405, 1.000, 0.652},
		{0.491, 0.779, 0.668, 0.653, 0.306, 0.652, 1.000},
	}
	// Table 1, columns "Equity Index Volatility (%)", "Equilibrium Portfolio
	// Weight (%)" and "Equilibrium Expected Returns (%)".
	vol := []float64{0.160, 0.203, 0.248, 0.271, 0.210, 0.200, 0.187}
	weight := []float64{0.016, 0.022, 0.052, 0.055, 0.116, 0.124, 0.615}
	want := []float64{3.9, 6.9, 8.4, 9.0, 4.3, 6.8, 7.6}

	cov := make([][]float64, len(vol))
	for i := range cov {
		cov[i] = make([]float64, len(vol))
		for j := range cov[i] {
			cov[i][j] = corr[i][j] * vol[i] * vol[j]
		}
	}
	got := optimize.ImpliedReturns(weight, cov, 2.5)
	if len(got) != len(want) {
		t.Fatalf("ImpliedReturns returned %d values for %d countries", len(got), len(want))
	}
	for i, name := range hlNames {
		within(t, "equilibrium return "+name, got[i]*100, want[i], 0.05)
	}
}

// idzorekNames are the eight asset classes of Idzorek (2005), in the order of
// his tables and of his view matrix P, followed by the two mini-portfolios
// his view 3 is stated between (see the test below).
var idzorekNames = []string{
	"US Bonds", "Int'l Bonds", "US Large Growth", "US Large Value",
	"US Small Growth", "US Small Value", "Int'l Dev. Equity", "Int'l Emerg. Equity",
	"US Growth 90/10", "US Value 90/10",
}

// idzorekCov is Table 5, the covariance matrix of excess returns.
var idzorekCov = [][]float64{
	{0.001005, 0.001328, -0.000579, -0.000675, 0.000121, 0.000128, -0.000445, -0.000437},
	{0.001328, 0.007277, -0.001307, -0.000610, -0.002237, -0.000989, 0.001442, -0.001535},
	{-0.000579, -0.001307, 0.059852, 0.027588, 0.063497, 0.023036, 0.032967, 0.048039},
	{-0.000675, -0.000610, 0.027588, 0.029609, 0.026572, 0.021465, 0.020697, 0.029854},
	{0.000121, -0.002237, 0.063497, 0.026572, 0.102488, 0.042744, 0.039943, 0.065994},
	{0.000128, -0.000989, 0.023036, 0.021465, 0.042744, 0.032056, 0.019881, 0.032235},
	{-0.000445, 0.001442, 0.032967, 0.020697, 0.039943, 0.019881, 0.028355, 0.035064},
	{-0.000437, -0.001535, 0.048039, 0.029854, 0.065994, 0.032235, 0.035064, 0.079958},
}

// idzorekMarketWeights is the final column of Table 2, and idzorekImplied the
// final column of Table 1 (the Implied Equilibrium Return Vector, in percent).
var (
	idzorekMarketWeights = []float64{0.1934, 0.2613, 0.1209, 0.1209, 0.0134, 0.0134, 0.2418, 0.0349}
	idzorekImplied       = []float64{0.08, 0.67, 6.41, 4.08, 7.43, 3.70, 4.80, 6.60}
)

// idzorekLambda is the risk aversion of the paper, "approximately 3.07" (the
// risk premium of 3 over the variance of the benchmark's excess returns).
const idzorekLambda = 3.07

// TestGoldenBlackLittermanIdzorek2005 replays the whole model on the
// eight-asset example: reverse optimization into Table 1's Implied
// Equilibrium Return Vector, the three views of section 2.2 blended into
// Table 6's New Combined Return Vector, and Table 6's New Weight Vector
// checked through the reverse-optimization identity that produced it.
//
// View 3 of the paper is stated over four lines with market-capitalization
// legs (0.9 large, 0.1 small on each side, formula 7), which the
// one-line-against-one-line "view:" grammar cannot write. It is fed here the
// way the paper itself describes it, as a view between the two MINI-
// PORTFOLIOS of its Tables 3a and 3b: two extra lines are appended to the
// problem, each a fixed blend of two of the eight, and the view is stated
// between them. A blend's covariance with everything (itself included) is the
// same blend of its legs' covariances, so the resulting view row is exactly
// the paper's, and so is the posterior over the original eight lines. Their
// equilibrium returns come out as Tables 3a and 3b print them, which is what
// says the two extra lines were built right.
func TestGoldenBlackLittermanIdzorek2005(t *testing.T) {
	// The two mini-portfolios of Tables 3a and 3b, over the eight lines.
	growth := []float64{0, 0, 0.9, 0, 0.1, 0, 0, 0}
	value := []float64{0, 0, 0, 0.9, 0, 0.1, 0, 0}
	cov := withBlends(idzorekCov, growth, value)
	prior := append(append([]float64(nil), idzorekMarketWeights...), 0, 0)

	implied := optimize.ImpliedReturns(prior, cov, idzorekLambda)
	for i := range idzorekImplied {
		within(t, "equilibrium return "+idzorekNames[i], implied[i]*100, idzorekImplied[i], 0.03)
	}
	// Tables 3a and 3b, "Total" of the Weighted Excess Return column.
	within(t, "equilibrium return "+idzorekNames[8], implied[8]*100, 6.52, 0.02)
	within(t, "equilibrium return "+idzorekNames[9], implied[9]*100, 4.04, 0.02)

	// The three views of section 2.2, at the confidence that reproduces the
	// paper's Omega = diag(P tau Sigma P'), which is 0.5 here.
	spec := optimize.Spec{Objective: optimize.BlackLitterman, Prior: prior, Views: []optimize.View{
		{Asset: "Int'l Dev. Equity", Return: 0.0525, Confidence: 0.5},
		{Asset: "Int'l Bonds", Versus: "US Bonds", Return: 0.0025, Confidence: 0.5},
		{Asset: "US Growth 90/10", Versus: "US Value 90/10", Return: 0.02, Confidence: 0.5},
	}}
	ids := make([][]string, len(idzorekNames))
	for i, n := range idzorekNames {
		ids[i] = []string{n}
	}
	if err := spec.Resolve(ids); err != nil {
		t.Fatal(err)
	}
	mu, err := optimize.Posterior(implied, cov, spec.Views)
	if err != nil {
		t.Fatal(err)
	}
	// Table 6, "New Combined Return Vector E[R]".
	wantMu := []float64{0.07, 0.50, 6.50, 4.32, 7.59, 3.94, 4.93, 6.84}
	for i := range wantMu {
		within(t, "posterior return "+idzorekNames[i], mu[i]*100, wantMu[i], 0.03)
	}

	// Table 6, "New Weight Vector w". It was produced by the unconstrained
	// maximization w = (lambda*Sigma)^-1 E[R] (the paper's formula 2), i.e.
	// E[R] = lambda*Sigma*w: reverse-optimizing the published weights must
	// give back the published returns. That is the same ImpliedReturns the
	// equilibrium step runs, pointed at the answer instead of the prior.
	// (Those weights sum to 103.63 %, so they are not an allocation and this
	// package's long-only Solve, which holds the budget at 100 %, does not
	// and should not reproduce them.)
	wantW := []float64{0.2988, 0.1559, 0.0935, 0.1482, 0.0104, 0.0165, 0.2781, 0.0349}
	back := optimize.ImpliedReturns(wantW, idzorekCov, idzorekLambda)
	for i := range wantW {
		within(t, "reverse-optimized weight "+idzorekNames[i], back[i]*100, wantMu[i], 0.03)
	}
}

// withBlends extends a covariance matrix with one extra line per blend, each
// blend a fixed portfolio of the original lines. The covariance of a blend
// with any line is that same blend of its legs' covariances, which is all
// this needs: cov(b, x) = sum_i b_i cov(i, x).
func withBlends(cov [][]float64, blends ...[]float64) [][]float64 {
	n, k := len(cov), len(blends)
	out := make([][]float64, n+k)
	for i := range out {
		out[i] = make([]float64, n+k)
	}
	for i := 0; i < n; i++ {
		copy(out[i], cov[i])
	}
	// Each blend's covariance with the original lines, then with the blends.
	legs := make([][]float64, k) // legs[b][i] = cov(blend b, line i)
	for b, blend := range blends {
		legs[b] = make([]float64, n)
		for i := 0; i < n; i++ {
			s := 0.0
			for j := 0; j < n; j++ {
				s += blend[j] * cov[i][j]
			}
			legs[b][i] = s
			out[i][n+b], out[n+b][i] = s, s
		}
	}
	for a := 0; a < k; a++ {
		for b := 0; b < k; b++ {
			s := 0.0
			for j := 0; j < n; j++ {
				s += blends[a][j] * legs[b][j]
			}
			out[n+a][n+b] = s
		}
	}
	return out
}
