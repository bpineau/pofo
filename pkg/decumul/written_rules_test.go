package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// steady repeats r for n periods.
func steady(r float64, n int) scenario.Sequence {
	s := make(scenario.Sequence, n)
	for i := range s {
		s[i] = r
	}
	return s
}

// Cap stops a raise at the cap, and a level already above it is not raised.
func TestGuardrailsCap(t *testing.T) {
	g := Guardrails{Upper: 0.06, Lower: 0.03, Cut: 0.1, Raise: 0.1, Cap: 52_000}
	if got := g.adjust(50_000, 2_000_000); got != 52_000 {
		t.Errorf("raise from 50 000 = %.0f, want the cap 52 000", got)
	}
	if got := g.adjust(53_000, 2_000_000); got != 53_000 {
		t.Errorf("raise from 53 000 = %.0f, want unchanged above the cap", got)
	}
	g.Cap = 0
	if got := g.adjust(50_000, 2_000_000); math.Abs(got-55_000) > 1e-9 {
		t.Errorf("uncapped raise = %.0f, want 55 000", got)
	}
}

// recentred moves the band onto the new rate at its relative width, and keeps
// it when the rate is not positive.
func TestGuardrailsRecentred(t *testing.T) {
	g := Guardrails{Upper: 0.048, Lower: 0.032}.recentred(0.01)
	if math.Abs(g.Upper-0.012) > 1e-12 || math.Abs(g.Lower-0.008) > 1e-12 {
		t.Errorf("band = [%.4f, %.4f], want [0.0080, 0.0120]", g.Lower, g.Upper)
	}
	if h := g.recentred(0); h != g {
		t.Errorf("a zero rate moved the band to %+v", h)
	}
}

// Without any income, the net sensor reads the same rate as the plain one, so
// NetOfIncome changes nothing, in both kernels.
func TestNetOfIncomeWithoutIncomeIsPlain(t *testing.T) {
	base := Plan{Capital: 1_000_000, NeedAnnual: 40_000, Years: 30,
		Guard: Guardrails{Upper: 0.048, Lower: 0.032, Cut: 0.1, Raise: 0.1, Floor: 32_000}}
	net := base
	net.Guard.NetOfIncome = true
	rets := scenario.Sequence{-0.2, -0.15, 0.3, 0.25, 0.1, -0.05, 0.2, 0.15, 0.12, -0.3}
	a, b := base.RunPath(rets, Lives{}), net.RunPath(rets, Lives{})
	for k := range a.Spend {
		if a.Spend[k] != b.Spend[k] {
			t.Fatalf("annual year %d: %.2f with the net sensor, %.2f without", k, b.Spend[k], a.Spend[k])
		}
	}
	monthly := steady(0.004, 30*12)
	monthly[5], monthly[40] = -0.15, 0.12
	a, b = base.RunPathMonthly(monthly, Lives{}), net.RunPathMonthly(monthly, Lives{})
	for k := range a.Spend {
		if a.Spend[k] != b.Spend[k] {
			t.Fatalf("monthly year %d: %.2f with the net sensor, %.2f without", k, b.Spend[k], a.Spend[k])
		}
	}
}

// A pension arriving in year 5 cuts the portfolio's net withdrawal from 4 %
// to 1 %: the re-centred band reads that as the new normal and does not raise.
func TestNetOfIncomeRecentresOnPension(t *testing.T) {
	p := Plan{Capital: 1_000_000, NeedAnnual: 40_000, Years: 8,
		Cashflows: []Cashflow{{FromYear: 5, Annual: 30_000}},
		Guard:     Guardrails{Upper: 0.048, Lower: 0.032, Cut: 0.1, Raise: 0.1, NetOfIncome: true}}
	res := p.RunPath(steady(0.04, 8), Lives{})
	for k, want := range []float64{40_000, 40_000, 40_000, 40_000, 40_000, 10_000, 10_000} {
		if math.Abs(res.Spend[k]-want) > 1e-6 {
			t.Errorf("year %d: the portfolio paid %.0f, want %.0f", k, res.Spend[k], want)
		}
	}
	// The same band left where it was reads 1 % against a 3.2 % floor and
	// raises the budget at once.
	stale := p
	stale.Cashflows = nil
	stale.Guard.NetOfIncome = true
	if r := stale.Guard.netMove(40_000, 40_000, 30_000, 1_000_000); r <= 40_000 {
		t.Errorf("unrecentred band: budget %.0f, want a raise", r)
	}
}

// RefillThreshold leaves a dead band: a -15 % year empties the buffer, an
// +18 % one brings the book back to ~8 % under its peak, where the default
// refills the buffer and a 5 % threshold does not.
func TestBufferRefillThreshold(t *testing.T) {
	five := 0.05
	p := Plan{Capital: 1_000_000, NeedAnnual: 40_000, Years: 3,
		Buffer: BufferSleeve{Years: 1}}
	rets := scenario.Sequence{-0.15, 0.18, 0}
	_, def := p.RunLedger(rets, Lives{})
	p.Buffer.RefillThreshold = &five
	_, dead := p.RunLedger(rets, Lives{})
	if def[1].FromBuffer == 0 || def[1].Buffer != 0 {
		t.Fatalf("year 1 should empty the buffer: drew %.0f, left %.0f", def[1].FromBuffer, def[1].Buffer)
	}
	if def[2].FromBuffer != 0 || dead[2].FromBuffer != 0 {
		t.Fatalf("buffer drawn at ~8 %% of drawdown (threshold 10 %%)")
	}
	if def[2].Buffer <= 0 || dead[2].Buffer != 0 {
		t.Errorf("year 2 buffer: %.0f by default, %.0f with the 5 %% threshold; want a refill only by default", def[2].Buffer, dead[2].Buffer)
	}
}

// MarketDrawdown reads the drawdown on returns: a flat market spent down by
// withdrawals never taps the buffer, while the wealth-based drawdown does.
func TestMarketDrawdownIgnoresSpendingDown(t *testing.T) {
	p := Plan{Capital: 1_000_000, NeedAnnual: 60_000, Years: 6,
		Buffer: BufferSleeve{Years: 2}}
	rets := steady(0, 6)
	_, wealth := p.RunLedger(rets, Lives{})
	p.MarketDrawdown = true
	_, market := p.RunLedger(rets, Lives{})
	tapped := func(rows []LedgerYear) bool {
		for _, r := range rows {
			if r.FromBuffer > 0 {
				return true
			}
		}
		return false
	}
	if !tapped(wealth) {
		t.Fatalf("wealth-based drawdown never tapped the buffer on a spent-down path")
	}
	if tapped(market) {
		t.Errorf("market drawdown tapped the buffer on a flat market")
	}
}
