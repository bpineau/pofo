package optimize

import (
	"math"
	"strings"
	"testing"
)

// blCov is a well-conditioned three-asset covariance: a hot equity line, a
// mid-volatility diversifier and a calm bond line, correlated the way such a
// book usually is.
func blCov() [][]float64 {
	sd := []float64{0.18, 0.10, 0.05}
	corr := [][]float64{
		{1.00, 0.20, -0.10},
		{0.20, 1.00, 0.30},
		{-0.10, 0.30, 1.00},
	}
	cov := make([][]float64, 3)
	for i := range cov {
		cov[i] = make([]float64, 3)
		for j := range cov[i] {
			cov[i][j] = corr[i][j] * sd[i] * sd[j]
		}
	}
	return cov
}

// blReturns builds three deterministic daily return series whose sample
// covariance is well conditioned, so Solve has a real problem to chew on.
func blReturns(t int) [][]float64 {
	out := make([][]float64, 3)
	for i := range out {
		out[i] = make([]float64, t)
	}
	// Three independent-looking deterministic waves plus a common factor.
	for k := 0; k < t; k++ {
		f := math.Sin(float64(k) * 0.11)
		out[0][k] = 0.0004 + 0.011*f + 0.006*math.Sin(float64(k)*0.37)
		out[1][k] = 0.0002 + 0.004*f + 0.005*math.Cos(float64(k)*0.23)
		out[2][k] = 0.0001 - 0.001*f + 0.003*math.Sin(float64(k)*0.71)
	}
	return out
}

// blSpec is a resolved Black-Litterman spec over three assets named A, B, C.
func blSpec(prior []float64, views ...View) Spec {
	s := Spec{Objective: BlackLitterman, Prior: prior, Views: views}
	if err := s.Resolve([][]string{{"A"}, {"B"}, {"C"}}); err != nil {
		panic(err)
	}
	return s
}

// The identity that makes the objective legible: with no view the posterior
// IS the implied vector, and the long-only optimum IS the prior. Reverse
// optimization and optimization are exact inverses, so a file with no view
// gets its own weights back and a reading of what they expect.
func TestBlackLittermanNoViewReturnsThePrior(t *testing.T) {
	returns := blReturns(1500)
	prior := []float64{0.5, 0.3, 0.2}
	res, err := Solve(returns, blSpec(prior))
	if err != nil {
		t.Fatal(err)
	}
	for i := range prior {
		approx(t, res.Weights[i], prior[i], 1e-6, "no-view weight")
	}
	for i := range prior {
		approx(t, res.Posterior[i], res.Implied[i], 1e-12, "no-view posterior")
	}
	// The same, inside bounds the prior already satisfies: the constrained
	// path must land on the prior too, not merely near it.
	spec := blSpec(prior)
	spec.Bounds = map[string][2]float64{"A": {0.40, 0.60}}
	if err := spec.Resolve([][]string{{"A"}, {"B"}, {"C"}}); err != nil {
		t.Fatal(err)
	}
	bounded, err := Solve(returns, spec)
	if err != nil {
		t.Fatal(err)
	}
	for i := range prior {
		approx(t, bounded.Weights[i], prior[i], 5e-3, "no-view bounded weight")
	}
}

// An absolute view above the implied return raises that asset's posterior and
// its weight; below, it lowers both. Nothing else in the problem changes.
func TestBlackLittermanAbsoluteViewDirection(t *testing.T) {
	returns := blReturns(1500)
	prior := []float64{0.5, 0.3, 0.2}
	base, err := Solve(returns, blSpec(prior))
	if err != nil {
		t.Fatal(err)
	}
	implied := base.Implied[1]

	up, err := Solve(returns, blSpec(prior, View{Asset: "B", Return: implied + 0.03, Confidence: 0.5}))
	if err != nil {
		t.Fatal(err)
	}
	if up.Posterior[1] <= implied {
		t.Fatalf("a view above the implied %.4f left the posterior at %.4f", implied, up.Posterior[1])
	}
	if up.Weights[1] <= prior[1] {
		t.Fatalf("a bullish view did not raise the weight: %.4f vs prior %.4f", up.Weights[1], prior[1])
	}

	down, err := Solve(returns, blSpec(prior, View{Asset: "B", Return: implied - 0.03, Confidence: 0.5}))
	if err != nil {
		t.Fatal(err)
	}
	if down.Posterior[1] >= implied {
		t.Fatalf("a view below the implied %.4f left the posterior at %.4f", implied, down.Posterior[1])
	}
	if down.Weights[1] >= prior[1] {
		t.Fatalf("a bearish view did not cut the weight: %.4f vs prior %.4f", down.Weights[1], prior[1])
	}
}

// A relative view moves the pair in opposite directions: the leg it favours
// gains what the other loses, and the untouched line keeps its weight.
func TestBlackLittermanRelativeViewMovesThePair(t *testing.T) {
	returns := blReturns(1500)
	prior := []float64{0.5, 0.3, 0.2}
	base, err := Solve(returns, blSpec(prior))
	if err != nil {
		t.Fatal(err)
	}
	gap := base.Implied[0] - base.Implied[1]

	// State a gap far wider than the equilibrium one: A must gain, B lose.
	res, err := Solve(returns, blSpec(prior, View{Asset: "A", Versus: "B", Return: gap + 0.04, Confidence: 0.6}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Weights[0] <= prior[0] || res.Weights[1] >= prior[1] {
		t.Fatalf("the pair did not move apart: %v (prior %v)", res.Weights, prior)
	}
	// The view says nothing about the level of the pair, only its spread:
	// the posterior spread must have widened past the equilibrium one.
	if got := res.Posterior[0] - res.Posterior[1]; got <= gap {
		t.Fatalf("posterior spread %.4f did not widen past the implied %.4f", got, gap)
	}
	if math.Abs(res.Weights[2]-prior[2]) > 0.10 {
		t.Fatalf("the unviewed line moved too far: %.4f vs prior %.4f", res.Weights[2], prior[2])
	}
}

// Confidence is the dial: at the floor the posterior is the equilibrium, at
// the ceiling the view holds exactly (P mu = Q), and it is monotone between.
func TestBlackLittermanConfidenceIsMonotone(t *testing.T) {
	cov := blCov()
	prior := []float64{0.5, 0.3, 0.2}
	implied := ImpliedReturns(prior, cov, 3.0)
	target := implied[1] + 0.05

	at := func(c float64) []float64 {
		s := blSpec(prior, View{Asset: "B", Return: target, Confidence: c})
		mu, err := Posterior(implied, cov, s.Views)
		if err != nil {
			t.Fatal(err)
		}
		return mu
	}
	almostNone := at(1e-9)
	approx(t, almostNone[1], implied[1], 1e-6, "posterior at zero confidence")
	almostSure := at(1 - 1e-9)
	approx(t, almostSure[1], target, 1e-6, "posterior at full confidence")

	prev := implied[1]
	for _, c := range []float64{0.1, 0.25, 0.5, 0.75, 0.9} {
		got := at(c)[1]
		if got <= prev {
			t.Fatalf("posterior at confidence %.2f = %.5f, not above %.5f", c, got, prev)
		}
		prev = got
	}
	if prev >= target {
		t.Fatalf("posterior %.5f overshot the view %.5f", prev, target)
	}
}

// tau cancels from the posterior because Omega is proportional to it, which
// is why it is fixed and exposed nowhere.
func TestBlackLittermanTauCancels(t *testing.T) {
	cov := blCov()
	prior := []float64{0.5, 0.3, 0.2}
	implied := ImpliedReturns(prior, cov, 3.0)
	s := blSpec(prior, View{Asset: "A", Versus: "C", Return: 0.06, Confidence: 0.4})
	want, err := Posterior(implied, cov, s.Views)
	if err != nil {
		t.Fatal(err)
	}
	got, err := posteriorAt(implied, cov, s.Views, 1.0)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		approx(t, got[i], want[i], 1e-12, "posterior with tau = 1")
	}
}

// Reverse optimization scales linearly with the risk aversion, and the two
// ways of fixing it are the documented ones: prior-return over the prior's
// variance, and an assumed Sharpe of 0.4 over its volatility.
func TestBlackLittermanRiskAversion(t *testing.T) {
	cov := blCov()
	prior := []float64{0.5, 0.3, 0.2}
	one := ImpliedReturns(prior, cov, 1)
	ten := ImpliedReturns(prior, cov, 10)
	for i := range one {
		approx(t, ten[i], 10*one[i], 1e-12, "implied returns scale with lambda")
	}

	variance := quad(cov, prior)
	lam, err := riskAversion(prior, cov, 0.046)
	if err != nil {
		t.Fatal(err)
	}
	approx(t, lam, 0.046/variance, 1e-12, "lambda from prior-return")
	// prior-return is exactly what the prior then expects.
	approx(t, dot(ImpliedReturns(prior, cov, lam), prior), 0.046, 1e-12, "the prior's expected return")

	def, err := riskAversion(prior, cov, 0)
	if err != nil {
		t.Fatal(err)
	}
	approx(t, def, 0.4/math.Sqrt(variance), 1e-12, "default lambda")
	// He and Litterman's 2.5 is NOT the default: on a calm book it would
	// imply a far smaller expected return than the Sharpe-0.4 assumption.
	if dot(ImpliedReturns(prior, cov, 2.5), prior) >= dot(ImpliedReturns(prior, cov, def), prior) {
		t.Fatal("the default lambda should expect more than delta = 2.5 on a low-volatility book")
	}
}

// The prior is the file's own weights, so Solve refuses anything else: a
// wrong length or a set of weights that is not an allocation.
func TestBlackLittermanPriorValidation(t *testing.T) {
	returns := blReturns(600)
	for _, tc := range []struct {
		name  string
		prior []float64
		want  string
	}{
		{"missing", nil, "3 assets"},
		{"short", []float64{0.5, 0.5}, "3 assets"},
		{"not an allocation", []float64{0.5, 0.3, 0.4}, "sum to"},
	} {
		_, err := Solve(returns, Spec{Objective: BlackLitterman, Prior: tc.prior})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s prior: error = %v, want one mentioning %q", tc.name, err, tc.want)
		}
	}
}

// A view must name a holding of the file, and its confidence must leave room
// for both the equilibrium and the view.
func TestBlackLittermanViewValidation(t *testing.T) {
	cov := blCov()
	implied := ImpliedReturns([]float64{0.5, 0.3, 0.2}, cov, 3)
	for _, tc := range []struct {
		name string
		view View
		want string
	}{
		{"confidence at the floor", View{Asset: "A", Return: 0.05}, "confidence"},
		{"confidence at the ceiling", View{Asset: "A", Return: 0.05, Confidence: 1}, "confidence"},
	} {
		v := tc.view
		v.asset, v.versus = 0, -1
		if _, err := Posterior(implied, cov, []View{v}); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: error = %v, want one mentioning %q", tc.name, err, tc.want)
		}
	}
	// A line compared with itself says nothing.
	self := View{Asset: "A", Versus: "A", Return: 0.01, Confidence: 0.5}
	self.asset, self.versus = 0, 0
	if _, err := Posterior(implied, cov, []View{self}); err == nil {
		t.Fatal("a view of a line against itself must be refused")
	}
}

// Resolve names the identifier it could not place, so a typo in a view reads
// as an error rather than as a view nobody applied.
func TestResolveViews(t *testing.T) {
	s := Spec{Objective: BlackLitterman, Views: []View{
		{Asset: "GOLD", Return: 0.02, Confidence: 0.5},
		{Asset: "TREND", Versus: "BONDS", Return: 0.03, Confidence: 0.7},
	}}
	if err := s.Resolve([][]string{{"BONDS", "DTLA"}, {"GOLD", "IGLN"}, {"TREND", "DBMFE"}}); err != nil {
		t.Fatal(err)
	}
	if s.Views[0].asset != 1 || s.Views[0].versus != -1 {
		t.Fatalf("absolute view resolved to %d/%d, want 1/-1", s.Views[0].asset, s.Views[0].versus)
	}
	if s.Views[1].asset != 2 || s.Views[1].versus != 0 {
		t.Fatalf("relative view resolved to %d/%d, want 2/0", s.Views[1].asset, s.Views[1].versus)
	}
	// The resolved symbol is accepted as well as the written id.
	sym := Spec{Objective: BlackLitterman, Views: []View{{Asset: "IGLN", Return: 0.02, Confidence: 0.5}}}
	if err := sym.Resolve([][]string{{"BONDS", "DTLA"}, {"GOLD", "IGLN"}}); err != nil {
		t.Fatal(err)
	}
	if sym.Views[0].asset != 1 {
		t.Fatalf("view on a resolved symbol landed on %d, want 1", sym.Views[0].asset)
	}

	bad := Spec{Objective: BlackLitterman, Views: []View{{Asset: "GLD", Return: 0.02, Confidence: 0.5}}}
	err := bad.Resolve([][]string{{"GOLD", "IGLN"}})
	if err == nil || !strings.Contains(err.Error(), "GLD") {
		t.Fatalf("error = %v, want one naming GLD", err)
	}
	other := Spec{Objective: BlackLitterman, Views: []View{{Asset: "GOLD", Versus: "TRND", Return: 0.02, Confidence: 0.5}}}
	err = other.Resolve([][]string{{"GOLD", "IGLN"}})
	if err == nil || !strings.Contains(err.Error(), "TRND") {
		t.Fatalf("error = %v, want one naming TRND", err)
	}
	// Resolve must not write through to the caller's views: pkg/compare
	// copies the Spec by value and the slice header travels with it.
	shared := []View{{Asset: "GOLD", Return: 0.02, Confidence: 0.5}}
	sp := Spec{Objective: BlackLitterman, Views: shared}
	if err := sp.Resolve([][]string{{"BONDS"}, {"GOLD"}}); err != nil {
		t.Fatal(err)
	}
	if shared[0].asset != 0 {
		t.Fatal("Resolve mutated the caller's view slice")
	}
}

// Black-Litterman accepts every constraint the other objectives do: the box
// simplex and the feasibility limits both bind on the posterior utility.
func TestBlackLittermanUnderBounds(t *testing.T) {
	returns := blReturns(1500)
	prior := []float64{0.5, 0.3, 0.2}
	spec := blSpec(prior, View{Asset: "A", Return: 0.30, Confidence: 0.9})
	spec.Bounds = map[string][2]float64{"A": {0, 0.35}}
	if err := spec.Resolve([][]string{{"A"}, {"B"}, {"C"}}); err != nil {
		t.Fatal(err)
	}
	res, err := Solve(returns, spec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Weights[0] > 0.35+1e-6 {
		t.Fatalf("the cap did not bind: %v", res.Weights)
	}
	approx(t, res.Weights[0], 0.35, 5e-3, "capped weight")
	sum := 0.0
	for _, w := range res.Weights {
		sum += w
	}
	approx(t, sum, 1, 1e-9, "weights sum")
	if res.Lambda <= 0 || len(res.Implied) != 3 || len(res.Posterior) != 3 {
		t.Fatalf("the report inputs are missing: lambda %v, implied %v", res.Lambda, res.Implied)
	}
	// Return/Volatility/Sharpe describe the POSTERIOR means, not the sample.
	approx(t, res.Return, dot(res.Posterior, res.Weights), 1e-9, "Return at the posterior")
}
