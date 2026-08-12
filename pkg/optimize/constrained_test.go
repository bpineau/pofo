package optimize

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// threeSleeves builds a hot/volatile asset, a mild one and a quiet one that
// dampens the other two: enough structure for a volatility cap or a return
// floor to actually bind.
func threeSleeves() [][]float64 {
	const n = 800
	hot := make([]float64, n)
	mild := make([]float64, n)
	quiet := make([]float64, n)
	for i := 0; i < n; i++ {
		x := float64(i)
		hot[i] = 0.0012 + 0.020*math.Sin(x*0.21)
		mild[i] = 0.0006 + 0.008*math.Sin(x*0.37+1)
		quiet[i] = 0.0002 - 0.004*math.Sin(x*0.21) + 0.002*math.Cos(x*0.9)
	}
	return [][]float64{hot, mild, quiet}
}

// blended returns the portfolio's daily returns for weights w.
func blended(returns [][]float64, w []float64) []float64 {
	out := make([]float64, len(returns[0]))
	blend(returns, w, out)
	return out
}

// TestBoundsAreRespected: per-asset bounds hold, and the free weight still
// sums to one.
func TestBoundsAreRespected(t *testing.T) {
	r := threeSleeves()
	spec := Spec{
		Objective: MaxReturn,
		Lower:     []float64{0.10, math.NaN(), 0.20},
		Upper:     []float64{0.30, math.NaN(), 0.40},
	}
	res, err := Solve(r, spec)
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.0
	for _, w := range res.Weights {
		sum += w
	}
	approx(t, sum, 1, 1e-6, "Σ weights")
	if res.Weights[0] < 0.10-1e-6 || res.Weights[0] > 0.30+1e-6 {
		t.Fatalf("asset 0 weight %.4f outside [0.10, 0.30]", res.Weights[0])
	}
	if res.Weights[2] < 0.20-1e-6 || res.Weights[2] > 0.40+1e-6 {
		t.Fatalf("asset 2 weight %.4f outside [0.20, 0.40]", res.Weights[2])
	}
	// Unbounded max-return would put everything in the hot asset; the cap
	// must actually bite.
	approx(t, res.Weights[0], 0.30, 1e-3, "capped max-return w0")
}

// TestMinWeightKeepsEveryLine: a weight floor stops the search from dropping
// assets it dislikes in sample.
func TestMinWeightKeepsEveryLine(t *testing.T) {
	r := threeSleeves()
	res, err := Solve(r, Spec{Objective: MaxReturn, MinWeight: 0.15})
	if err != nil {
		t.Fatal(err)
	}
	for i, w := range res.Weights {
		if w < 0.15-1e-6 {
			t.Fatalf("asset %d weight %.4f below the 15 %% floor", i, w)
		}
	}
}

// TestVolatilityCapBinds: max-return under a volatility cap lands on the cap
// (the search wants more return, the cap is what stops it) and beats the
// minimum-volatility portfolio's return.
func TestVolatilityCapBinds(t *testing.T) {
	r := threeSleeves()
	floor, err := Solve(r, Spec{Objective: MinVolatility})
	if err != nil {
		t.Fatal(err)
	}
	cap := metrics.Volatility(blended(r, floor.Weights)) * 1.5
	res, err := Solve(r, Spec{Objective: MaxReturn, Limits: Limits{MaxVolatility: cap}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Feasible {
		t.Fatal("solve reported the volatility cap as unreachable")
	}
	got := metrics.Volatility(blended(r, res.Weights))
	if got > cap+1e-4 {
		t.Fatalf("volatility %.4f above the %.4f cap", got, cap)
	}
	if got < cap*0.9 {
		t.Fatalf("volatility %.4f well under the %.4f cap: the search left return on the table", got, cap)
	}
	if res.CAGR <= compound(blended(r, floor.Weights)) {
		t.Fatalf("capped max-return CAGR %.4f no better than min-volatility's %.4f",
			res.CAGR, compound(blended(r, floor.Weights)))
	}
}

// TestReturnFloorBinds: min-volatility under a return floor reaches the floor
// and is calmer than the unconstrained max-return portfolio.
func TestReturnFloorBinds(t *testing.T) {
	r := threeSleeves()
	top, err := Solve(r, Spec{Objective: MaxReturn})
	if err != nil {
		t.Fatal(err)
	}
	floor := compound(blended(r, top.Weights)) * 0.7
	res, err := Solve(r, Spec{Objective: MinVolatility, Limits: Limits{MinReturn: floor}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Feasible {
		t.Fatal("solve reported the return floor as unreachable")
	}
	if got := compound(blended(r, res.Weights)); got < floor-1e-4 {
		t.Fatalf("CAGR %.4f below the %.4f floor", got, floor)
	}
	if metrics.Volatility(blended(r, res.Weights)) >= metrics.Volatility(blended(r, top.Weights)) {
		t.Fatal("constrained min-volatility is not calmer than plain max-return")
	}
}

// TestDrawdownBudget: a drawdown budget is respected when reachable.
func TestDrawdownBudget(t *testing.T) {
	r := threeSleeves()
	res, err := Solve(r, Spec{Objective: MaxReturn, Limits: Limits{MaxDrawdown: 0.05}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Feasible {
		t.Skip("no allocation meets a 5 % drawdown budget on this fixture")
	}
	if dd := math.Abs(pathMaxDrawdown(blended(r, res.Weights))); dd > 0.05+1e-4 {
		t.Fatalf("drawdown %.4f above the 5 %% budget", dd)
	}
}

// TestUnreachableLimitReportsInfeasible: an impossible limit returns the
// least-violating point, flagged, rather than a plausible-looking answer.
func TestUnreachableLimitReportsInfeasible(t *testing.T) {
	r := threeSleeves()
	res, err := Solve(r, Spec{Objective: MaxReturn, Limits: Limits{MinReturn: 10}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Feasible {
		t.Fatal("a 1000 %/yr return floor was reported as feasible")
	}
}

// TestInfeasibleBoxErrors: bounds that cannot hold a portfolio fail loudly.
func TestInfeasibleBoxErrors(t *testing.T) {
	r := threeSleeves()
	if _, err := Solve(r, Spec{Objective: MaxReturn, MinWeight: 0.4}); err == nil {
		t.Fatal("three 40 % floors sum to 120 % and must be rejected")
	}
	_, err := Solve(r, Spec{Objective: MaxReturn, Upper: []float64{0.2, 0.2, 0.2}})
	if err == nil || !strings.Contains(err.Error(), "below 100") {
		t.Fatalf("caps summing to 60 %% must be rejected, got %v", err)
	}
}

// TestConstrainedPathMatchesClosedForm: with a cap but no bounds or limits the
// closed form still runs; asked for the same problem, the penalized path
// search finds the same optimum. Any drift between the two is a bug in the new
// path, since the old one is the reference.
func TestConstrainedPathMatchesClosedForm(t *testing.T) {
	r := threeSleeves()
	closed, err := Solve(r, Spec{Objective: MaxSharpe, MaxWeight: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	// MinWeight 0 changes nothing but forces the constrained route.
	spec := Spec{Objective: MaxSharpe, MaxWeight: 0.5, Lower: []float64{0, math.NaN(), math.NaN()}}
	search, err := Solve(r, spec)
	if err != nil {
		t.Fatal(err)
	}
	for i := range closed.Weights {
		approx(t, search.Weights[i], closed.Weights[i], 5e-3, "weight")
	}
}

// TestResolveBounds: bounds are matched by any of an asset's spellings, and an
// identifier that matches nothing is an error rather than a silent no-op.
func TestResolveBounds(t *testing.T) {
	spec := Spec{Bounds: map[string][2]float64{"NTSG": {0.15, 0.30}, "gde": {0.10, 0.25}}}
	ids := [][]string{{"NTSG", "IE00077IIPQ8"}, {"ZPRV", "IE00BSPLC413"}, {"GDE", "US97717Y5684"}}
	if err := spec.Resolve(ids); err != nil {
		t.Fatal(err)
	}
	approx(t, spec.Lower[0], 0.15, 1e-9, "NTSG floor")
	approx(t, spec.Upper[2], 0.25, 1e-9, "GDE cap (matched case-insensitively)")
	if !math.IsNaN(spec.Lower[1]) {
		t.Fatal("an unbounded asset must keep NaN bounds, not 0")
	}
	typo := Spec{Bounds: map[string][2]float64{"NTGS": {0.15, 0.30}}}
	if err := typo.Resolve(ids); err == nil {
		t.Fatal("a bound naming no holding must fail")
	}
}

// TestParseSpecConstraints covers the new "#meta optimize:" grammar.
func TestParseSpecConstraints(t *testing.T) {
	spec, err := ParseSpec("max-return,max-vol:9.5,min-weight:5,bounds:NTSG:15-30,bounds:GDE:-25,train:1996..2015")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Objective != MaxReturn {
		t.Fatalf("objective = %q", spec.Objective)
	}
	approx(t, spec.Limits.MaxVolatility, 0.095, 1e-9, "max-vol")
	approx(t, spec.MinWeight, 0.05, 1e-9, "min-weight")
	approx(t, spec.Bounds["NTSG"][0], 0.15, 1e-9, "NTSG floor")
	approx(t, spec.Bounds["NTSG"][1], 0.30, 1e-9, "NTSG cap")
	if !math.IsNaN(spec.Bounds["GDE"][0]) {
		t.Fatal("an omitted low end must stay open")
	}
	if got := spec.Train.String(); got != "1996-01-01→2015-12-31" {
		t.Fatalf("train window = %s", got)
	}

	if spec, err = ParseSpec("min-volatility,min-return:10.5,max-drawdown:-20,train:..2015-06-30"); err != nil {
		t.Fatal(err)
	}
	approx(t, spec.Limits.MinReturn, 0.105, 1e-9, "min-return")
	approx(t, spec.Limits.MaxDrawdown, 0.20, 1e-9, "max-drawdown (sign-insensitive)")
	if !spec.Train.Start.IsZero() || spec.Train.End.Format("2006-01-02") != "2015-06-30" {
		t.Fatalf("open-start train window = %s", spec.Train)
	}

	for _, bad := range []string{
		"max-return,max-vol:0",
		"max-return,bounds:NTSG",
		"max-return,bounds:NTSG:40-10",
		"max-return,train:2015..1996",
		"max-return,train:..",
		"max-return,train:yesterday..2015",
		"max-return,min-weight:60,max-weight:40",
	} {
		if _, err := ParseSpec(bad); err == nil {
			t.Fatalf("ParseSpec(%q) must fail", bad)
		}
	}
}

// TestParseSpecRejectsUnenforceableConstraints: risk-parity and cwarp are
// solved by code that never reads Limits (and, for cwarp, never reads the
// per-asset box either), so those combinations must fail at parse time
// instead of producing weights that ignore what was asked.
func TestParseSpecRejectsUnenforceableConstraints(t *testing.T) {
	for _, bad := range []string{
		"risk-parity,max-vol:9.5",
		"risk-parity,min-return:5",
		"risk-parity,max-drawdown:20",
		"cwarp,max-vol:9.5",
		"cwarp,min-return:5",
		"cwarp,max-dd:20",
		"cwarp,min-weight:5",
		"cwarp,bounds:NTSG:15-30",
	} {
		if _, err := ParseSpec(bad); err == nil {
			t.Fatalf("ParseSpec(%q) must fail: the solver would ignore the constraint", bad)
		}
	}
	// The lenient cases stay lenient: risk-parity accepts (and the report
	// notes) weight bounds, and cwarp accepts its scalar cap.
	for _, ok := range []string{
		"risk-parity,max-weight:40",
		"risk-parity,min-weight:5",
		"risk-parity,bounds:NTSG:15-30",
		"cwarp,max-weight:50",
		"risk-parity,train:1996..2015",
	} {
		if _, err := ParseSpec(ok); err != nil {
			t.Fatalf("ParseSpec(%q) must parse: %v", ok, err)
		}
	}
}

// TestParseSpecRejectsInertReturnFloor: Limits uses zero as its "no limit"
// sentinel, so a zero or negative CAGR floor cannot be expressed; it used to
// parse and then constrain nothing.
func TestParseSpecRejectsInertReturnFloor(t *testing.T) {
	for _, bad := range []string{"max-sharpe,min-return:0", "max-sharpe,min-return:-5"} {
		if _, err := ParseSpec(bad); err == nil {
			t.Fatalf("ParseSpec(%q) must fail: the floor would be silently inert", bad)
		}
	}
	spec, err := ParseSpec("max-sharpe,min-return:0.5")
	if err != nil {
		t.Fatal(err)
	}
	approx(t, spec.Limits.MinReturn, 0.005, 1e-12, "a small but real floor")
}

// TestWindowContains: the parsed window is inclusive on both ends.
func TestWindowContains(t *testing.T) {
	w, err := ParseSpec("max-sharpe,train:2000..2005")
	if err != nil {
		t.Fatal(err)
	}
	in := func(s string) bool {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return w.Train.Contains(d)
	}
	for _, d := range []string{"2000-01-01", "2003-06-15", "2005-12-31"} {
		if !in(d) {
			t.Fatalf("%s should be inside %s", d, w.Train)
		}
	}
	for _, d := range []string{"1999-12-31", "2006-01-01"} {
		if in(d) {
			t.Fatalf("%s should be outside %s", d, w.Train)
		}
	}
}
