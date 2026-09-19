package portfolio

import (
	"math"
	"math/rand"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// The properties a simulation owes its reader, stated once and checked on
// synthetic paths rather than on any particular fixture. They are the net
// under the arithmetic: each one fails loudly for a whole family of mistakes
// (a stale price index, a weight read in the wrong units, an attribution that
// forgets the drift between rebalancings) that no single example would catch.
// Seeds are fixed, so a failure is always reproducible.

// walk builds a geometric random walk quoting every calendar day.
func walk(rng *rand.Rand, symbol string, startDay, n int, base, vol, drift float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	px := base
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(startDay + i), Close: px})
		px *= 1 + drift + vol*rng.NormFloat64()
	}
	return s
}

// rescaled returns the same path quoted in other units (a denomination
// change, a share split the provider back-adjusted).
func rescaled(s *marketdata.Series, c float64) *marketdata.Series {
	out := &marketdata.Series{Symbol: s.Symbol}
	for _, p := range s.Points {
		out.Points = append(out.Points, marketdata.Point{Date: p.Date, Close: p.Close * c})
	}
	return out
}

// worstDiff is the largest absolute difference between two equally long
// series, relative for values above 1.
func worstDiff(t *testing.T, a, b []float64) float64 {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("series of different lengths: %d and %d", len(a), len(b))
	}
	m := 0.0
	for i := range a {
		d := math.Abs(a[i] - b[i])
		if s := math.Abs(a[i]); s > 1 {
			d /= s
		}
		m = math.Max(m, d)
	}
	return m
}

// A price is a unit of account, not a quantity: quoting one holding a
// thousand times higher must leave the portfolio, and every holding's
// contribution to it, exactly where they were.
func TestPropertyPriceScaleInvariance(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := range 20 {
		a := walk(rng, "A", 0, 400, 100, 0.01, 0.0003)
		b := walk(rng, "B", 0, 400, 37, 0.007, 0.0002)
		mk := func(c float64) *Portfolio {
			return &Portfolio{Name: "t", EnvelopeFees: 0.5, Assets: []Asset{
				{Symbol: "A", Weight: 0.6, Series: a},
				{Symbol: "B", Weight: 0.4, Series: rescaled(b, c)},
			}}
		}
		base, err := Simulate(mk(1), 30)
		if err != nil {
			t.Fatal(err)
		}
		up, err := Simulate(mk(1234.5), 30)
		if err != nil {
			t.Fatal(err)
		}
		if d := worstDiff(t, base.Index, up.Index); d > 1e-10 {
			t.Fatalf("trial %d: the index moved by %g under a price rescale", trial, d)
		}
		for i := range base.Contributions {
			if d := worstDiff(t, base.Contributions[i], up.Contributions[i]); d > 1e-10 {
				t.Fatalf("trial %d, asset %d: contributions moved by %g", trial, i, d)
			}
		}
	}
}

// Writing one holding twice at half its weight is the same book.
func TestPropertySplitHoldingChangesNothing(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	a := walk(rng, "A", 0, 300, 100, 0.01, 0.0003)
	b := walk(rng, "B", 0, 300, 50, 0.008, 0.0001)
	one, err := Simulate(&Portfolio{Assets: []Asset{
		{Symbol: "A", Weight: 0.6, Series: a},
		{Symbol: "B", Weight: 0.4, Series: b},
	}}, 20)
	if err != nil {
		t.Fatal(err)
	}
	two, err := Simulate(&Portfolio{Assets: []Asset{
		{Symbol: "A", Weight: 0.3, Series: a},
		{Symbol: "A-again", Weight: 0.3, Series: a},
		{Symbol: "B", Weight: 0.4, Series: b},
	}}, 20)
	if err != nil {
		t.Fatal(err)
	}
	if d := worstDiff(t, one.Index, two.Index); d > 1e-10 {
		t.Fatalf("splitting a holding moved the index by %g", d)
	}
}

// A single holding at 100 % with no fees IS the asset, at any rebalancing
// period: the simulation must add nothing of its own.
func TestPropertySingleHoldingReproducesTheAsset(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	a := walk(rng, "A", 0, 500, 100, 0.012, 0.0002)
	for _, rebalance := range []int{0, 1, 7, 90} {
		sim, err := Simulate(&Portfolio{Assets: []Asset{{Symbol: "A", Weight: 1, Series: a}}}, rebalance)
		if err != nil {
			t.Fatal(err)
		}
		for k := range sim.Dates {
			want := 100 * a.Points[k].Close / a.Points[0].Close
			if math.Abs(sim.Index[k]/want-1) > 1e-12 {
				t.Fatalf("rebalance %d, day %d: index %v, want %v", rebalance, k, sim.Index[k], want)
			}
		}
	}
}

// Rebalancing between two holdings that never diverge trades nothing, so it
// can cost nothing either.
func TestPropertyRebalancingIdenticalHoldings(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	a := walk(rng, "A", 0, 300, 100, 0.01, 0.0002)
	mk := func() *Portfolio {
		return &Portfolio{Assets: []Asset{
			{Symbol: "A", Weight: 0.5, Series: a},
			{Symbol: "B", Weight: 0.5, Series: rescaled(a, 3)},
		}}
	}
	never, err := Simulate(mk(), 0)
	if err != nil {
		t.Fatal(err)
	}
	daily, err := Simulate(mk(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if d := worstDiff(t, never.Index, daily.Index); d > 1e-10 {
		t.Fatalf("daily rebalancing between identical holdings moved the index by %g", d)
	}
}

// The holdings are a set, not a list: their order in the file cannot change
// what the book did.
func TestPropertyHoldingOrderChangesNothing(t *testing.T) {
	rng := rand.New(rand.NewSource(6))
	a := walk(rng, "A", 0, 300, 100, 0.01, 0.0003)
	b := walk(rng, "B", 0, 300, 50, 0.006, 0.0001)
	forward, err := Simulate(&Portfolio{Assets: []Asset{
		{Symbol: "A", Weight: 0.6, Series: a}, {Symbol: "B", Weight: 0.4, Series: b}}}, 30)
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := Simulate(&Portfolio{Assets: []Asset{
		{Symbol: "B", Weight: 0.4, Series: b}, {Symbol: "A", Weight: 0.6, Series: a}}}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if d := worstDiff(t, forward.Index, reverse.Index); d > 1e-12 {
		t.Fatalf("reversing the holdings moved the index by %g", d)
	}
}

// Index is time-weighted: it measures the strategy, so the SIZE of the saver's
// contributions must not touch it.
func TestPropertyIndexIgnoresTheFlowSize(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	a := walk(rng, "A", 0, 500, 100, 0.01, 0.0003)
	b := walk(rng, "B", 0, 500, 50, 0.006, 0.0001)
	mk := func(amount float64) *Portfolio {
		return &Portfolio{Capital: 100000, Contribute: Flow{Amount: amount, Period: Monthly},
			Assets: []Asset{
				{Symbol: "A", Weight: 0.6, Series: a}, {Symbol: "B", Weight: 0.4, Series: b}}}
	}
	small, err := Simulate(mk(1), 30)
	if err != nil {
		t.Fatal(err)
	}
	big, err := Simulate(mk(50000), 30)
	if err != nil {
		t.Fatal(err)
	}
	if d := worstDiff(t, small.Index, big.Index); d > 1e-9 {
		t.Fatalf("the index moved by %g when the contribution size changed", d)
	}
}

// Without flows the two series only differ by scale (SimResult's own words).
func TestPropertyValuesTrackIndexWithoutFlows(t *testing.T) {
	rng := rand.New(rand.NewSource(8))
	a := walk(rng, "A", 0, 300, 100, 0.01, 0.0003)
	sim, err := Simulate(&Portfolio{Capital: 12345, EnvelopeFees: 0.8,
		Assets: []Asset{{Symbol: "A", Weight: 1, Series: a}}}, 30)
	if err != nil {
		t.Fatal(err)
	}
	for k := range sim.Dates {
		if math.Abs(sim.Values[k]/12345*100/sim.Index[k]-1) > 1e-12 {
			t.Fatalf("day %d: values %v against index %v", k, sim.Values[k], sim.Index[k])
		}
	}
}

// The attribution must close: the holdings' contributions rebuild the index
// return exactly, except for the two things SimResult says they leave out,
// each of which is bounded by what it costs over the step.
func TestPropertyContributionsRebuildTheReturn(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	a := walk(rng, "A", 0, 400, 100, 0.01, 0.0003)
	b := walk(rng, "B", 0, 400, 50, 0.006, 0.0001)
	cash := &marketdata.Series{Symbol: "^IRX"}
	for i := range 400 {
		cash.Points = append(cash.Points, marketdata.Point{Date: day(i), Close: 4})
	}
	const oneDay = 1 / 365.25
	for _, tc := range []struct {
		name string
		p    *Portfolio
		tol  float64
	}{
		{"plain", &Portfolio{Assets: []Asset{
			{Symbol: "A", Weight: 0.6, Series: a}, {Symbol: "B", Weight: 0.4, Series: b}}}, 1e-14},
		{"flows", &Portfolio{Capital: 100000, Contribute: Flow{Amount: 500, Period: Monthly},
			Assets: []Asset{
				{Symbol: "A", Weight: 0.6, Series: a}, {Symbol: "B", Weight: 0.4, Series: b}}}, 1e-14},
		// Envelope fees are unattributed by design, so the gap is one step of
		// the fee, levied on the day's grown value: (1+r) = (1-f)(1+sum).
		{"envelope fees", &Portfolio{EnvelopeFees: 1.2, Assets: []Asset{
			{Symbol: "A", Weight: 0.6, Series: a}, {Symbol: "B", Weight: 0.4, Series: b}}},
			1.2 / 100 * oneDay},
		// So is the leverage cash leg: financing at 4 % plus a 1 % spread on
		// half the capital borrowed.
		{"leverage", &Portfolio{Leverage: true, BorrowSpread: 1, Cash: cash, Assets: []Asset{
			{Symbol: "A", Weight: 0.9, Series: a}, {Symbol: "B", Weight: 0.6, Series: b}}},
			5.0 / 100 * oneDay},
	} {
		sim, err := Simulate(tc.p, 30)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		for k := 1; k < len(sim.Dates); k++ {
			sum := 0.0
			for i := range sim.Contributions {
				sum += sim.Contributions[i][k]
			}
			r := sim.Index[k]/sim.Index[k-1] - 1
			want := tc.tol*(1+math.Abs(sum)) + 1e-14
			if d := math.Abs(sum - r); d > want {
				t.Fatalf("%s, day %d: contributions sum %v against index return %v (gap %g, tolerance %g)",
					tc.name, k, sum, r, d, want)
			}
		}
	}
}
