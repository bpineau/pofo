package simgen

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// vmFrame builds a frame of n daily dates whose asset returns are ret(k) and
// whose cash earns a constant daily accrual.
func vmFrame(n int, ret func(k int) float64, cash float64) *Frame {
	fr := &Frame{Returns: map[string][]float64{"A": make([]float64, n), "C": make([]float64, n)}}
	for k := range n {
		fr.Dates = append(fr.Dates, day(k))
		if k > 0 {
			fr.Returns["A"][k] = ret(k)
			fr.Returns["C"][k] = cash
		}
	}
	return fr
}

// alternating returns +a, -a, +a, ...: over any window of an even number of
// returns the sample volatility is the same.
func alternating(a float64) func(int) float64 {
	return func(k int) float64 {
		if k%2 == 0 {
			return -a
		}
		return a
	}
}

// alternatingVol is the annualized sample volatility of an even window of
// alternating(a).
func alternatingVol(a float64, window int) float64 {
	return a * math.Sqrt(float64(window)/float64(window-1)*metrics.TradingDaysPerYear)
}

func vmConfig() VolManagedConfig {
	return VolManagedConfig{AssetID: "A", CashID: "C", TargetVol: 0.10, Lookback: 20, Floor: 0, Cap: 10, Rebalance: 5}
}

func TestVolManagedConstantVolHoldsAConstantMultiplier(t *testing.T) {
	const a = 0.01
	fr := vmFrame(400, alternating(a), 1e-4)
	cfg := vmConfig()
	p, err := VolManaged(fr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := cfg.TargetVol / alternatingVol(a, cfg.Lookback)
	if p.Start != cfg.Lookback || len(p.Values) != 400-cfg.Lookback || len(p.Multiplier) != len(p.Values) {
		t.Fatalf("start %d, %d values, %d multipliers", p.Start, len(p.Values), len(p.Multiplier))
	}
	for i, m := range p.Multiplier {
		if math.Abs(m-want) > 1e-12 {
			t.Fatalf("multiplier %d = %v, want %v throughout", i, m, want)
		}
	}
}

// A volatility shock tripling the daily moves cuts the multiplier to a third,
// and only at the first refresh whose window holds nothing but the shock: no
// earlier (the engine must not see the future), no later.
func TestVolManagedCutsOnTheShockAtTheRefreshDate(t *testing.T) {
	const a, shock = 0.005, 200
	cfg := vmConfig()
	cfg.Rebalance = cfg.Lookback // refreshes at 20, 40, ..., 200, 220
	fr := vmFrame(300, func(k int) float64 {
		if k > shock {
			return 3 * alternating(a)(k)
		}
		return alternating(a)(k)
	}, 0)
	p, err := VolManaged(fr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	calm := cfg.TargetVol / alternatingVol(a, cfg.Lookback)
	at := func(k int) float64 { return p.Multiplier[k-p.Start] }
	// The refresh at 200 reads returns 181..200, all calm; the one at 220
	// reads 201..220, all shocked. Between them m holds.
	for k := p.Start; k < shock+cfg.Lookback; k++ {
		near(t, "calm multiplier", at(k), calm, 1e-12)
	}
	near(t, "multiplier after the shock", at(shock+cfg.Lookback), calm/3, 1e-12)
	// The step from 221 on is the first one the cut applies to.
	k := shock + cfg.Lookback + 1
	got := p.Values[k-p.Start]/p.Values[k-1-p.Start] - 1
	near(t, "first return at the cut", got, calm/3*fr.Returns["A"][k], 1e-15)
}

// By default the multiplier is refreshed at month ends: a shock in the middle
// of a month moves nothing before that month's last date.
func TestVolManagedRefreshesAtMonthEnds(t *testing.T) {
	const a = 0.005
	shockDay := time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC)
	fr := vmFrame(250, alternating(a), 0)
	for k, d := range fr.Dates {
		if !d.Before(shockDay) {
			fr.Returns["A"][k] *= 4
		}
	}
	cfg := vmConfig()
	cfg.Rebalance = 0
	p, err := VolManaged(fr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if d := fr.Dates[p.Start]; d.Day() != 31 || d.Month() != time.January {
		t.Errorf("first refresh %s, want the first month end past the lookback, 2020-01-31", d.Format(time.DateOnly))
	}
	prev := p.Multiplier[0]
	for i, m := range p.Multiplier {
		d := fr.Dates[p.Start+i]
		monthEnd := d.AddDate(0, 0, 1).Month() != d.Month()
		if m != prev && !monthEnd {
			t.Errorf("multiplier moved on %s, not a month end", d.Format(time.DateOnly))
		}
		if d.Equal(time.Date(2020, 5, 31, 0, 0, 0, 0, time.UTC)) && !(m < prev) {
			t.Errorf("multiplier on %s = %v, want a cut below %v", d.Format(time.DateOnly), m, prev)
		}
		prev = m
	}
}

func TestVolManagedCapOfOneNeverLevers(t *testing.T) {
	// A calm asset (about 1.6 %/yr of volatility) against a 10 % target calls
	// for six times the capital: the cap holds it at one.
	fr := vmFrame(500, func(k int) float64 { return 1e-3 * math.Sin(float64(k)) }, 1e-4)
	cfg := vmConfig()
	cfg.Cap, cfg.Floor = 1, 0.5
	p, err := VolManaged(fr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for i, m := range p.Multiplier {
		if m > 1 {
			t.Fatalf("multiplier %d = %v above a cap of 1", i, m)
		}
	}
}

func TestVolManagedAtOneIsTheAsset(t *testing.T) {
	fr := vmFrame(300, func(k int) float64 { return 0.01*math.Sin(0.7*float64(k)) + 3e-4 }, 2e-4)
	cfg := vmConfig()
	cfg.Floor, cfg.Cap = 1, 1
	cfg.Costs = LeverageCosts{AnnualFee: 0.01, Spread: 0.01, ETFLeverage: 2}
	p, err := VolManaged(fr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	level := 100.0
	for i := 1; i < len(p.Values); i++ {
		level *= 1 + fr.Returns["A"][p.Start+i]
		if p.Values[i] != level {
			t.Fatalf("value %d = %v, want the asset's own %v", i, p.Values[i], level)
		}
	}
}

func TestVolManagedCostsOnlyAboveOne(t *testing.T) {
	ret := func(k int) float64 { return 0.01*math.Sin(0.7*float64(k)) + 3e-4 }
	fr := vmFrame(300, ret, 2e-4)
	costs := LeverageCosts{AnnualFee: 0.0035, Spread: 0.005, ETFLeverage: 2}

	// Capped at one, the cost model never bites: bit for bit the same index.
	free, priced := vmConfig(), vmConfig()
	free.Cap, priced.Cap = 1, 1
	priced.Costs = costs
	pf, _ := VolManaged(fr, free)
	pp, err := VolManaged(fr, priced)
	if err != nil {
		t.Fatal(err)
	}
	for i := range pf.Values {
		if pf.Values[i] != pp.Values[i] {
			t.Fatalf("value %d: %v with costs, %v without, at m <= 1", i, pp.Values[i], pf.Values[i])
		}
	}

	// Held at 1.5, every step pays on the excess 0.5: fee and spread, plus the
	// 2x ETF's reset drag, X × σ² for L = 2.
	// Traded back to it every step, so each step starts at exactly 1.5.
	free.Floor, free.Cap, free.Rebalance = 1.5, 1.5, 1
	priced.Floor, priced.Cap, priced.Rebalance = 1.5, 1.5, 1
	pf, _ = VolManaged(fr, free)
	pp, _ = VolManaged(fr, priced)
	k := pp.Start + 10
	vol := realizedVol(fr.Returns["A"], k-1, priced.Lookback)
	want := 0.5 * (costs.AnnualFee + costs.Spread + vol*vol) / 365.25
	gotFree := pf.Values[10]/pf.Values[9] - 1
	gotPriced := pp.Values[10]/pp.Values[9] - 1
	near(t, "cost of one step at m = 1.5", gotFree-gotPriced, want, 1e-15)
}

func TestVolManagedRejectsABadConfig(t *testing.T) {
	fr := vmFrame(100, alternating(0.01), 0)
	for _, tc := range []struct {
		name string
		edit func(*VolManagedConfig)
		want string
	}{
		{"no target", func(c *VolManagedConfig) { c.TargetVol = 0 }, "TargetVol"},
		{"NaN target", func(c *VolManagedConfig) { c.TargetVol = math.NaN() }, "TargetVol"},
		{"short lookback", func(c *VolManagedConfig) { c.Lookback = 1 }, "Lookback"},
		{"floor above cap", func(c *VolManagedConfig) { c.Floor, c.Cap = 2, 1 }, "Floor"},
		{"negative cost", func(c *VolManagedConfig) { c.Costs.Spread = -0.01 }, "costs"},
		{"missing asset", func(c *VolManagedConfig) { c.AssetID = "B" }, "missing"},
		{"history too short", func(c *VolManagedConfig) { c.Lookback = 99 }, "too short"},
	} {
		cfg := vmConfig()
		tc.edit(&cfg)
		if _, err := VolManaged(fr, cfg); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want an error naming %q", tc.name, err, tc.want)
		}
	}
	if _, err := NewVolManagedBuilder("x", VolManagedConfig{}); err == nil {
		t.Error("NewVolManagedBuilder accepted an empty config")
	}
}

func TestVolManagedBuilderBuildsOnAFetcher(t *testing.T) {
	f := fakeFetcher{"A": mkWobbly("A", 300, 3e-4, 0.01), "^IRX": mkLevels("^IRX", 300, 3)}
	cfg := vmConfig()
	cfg.CashID = "^IRX"
	b, err := NewVolManagedBuilder("A managed", cfg)
	if err != nil {
		t.Fatal(err)
	}
	s, err := b.Build(f, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "A managed" || s.First().Close != 100 || !s.First().Date.Equal(day(cfg.Lookback)) || s.Len() != 300-cfg.Lookback {
		t.Errorf("built %q from %s at %v, %d points", s.Name, s.First().Date.Format(time.DateOnly), s.First().Close, s.Len())
	}
}
