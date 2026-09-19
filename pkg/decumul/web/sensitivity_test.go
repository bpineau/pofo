package web

import (
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
)

func TestSensitivityRendersSignedBars(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 50000, Years: 40,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500}

	res := Sensitivity(pr, nil)

	if !strings.HasPrefix(res.SVG, "<svg") {
		t.Fatalf("expected an SVG, got %.30q", res.SVG)
	}
	// Spending less and adding capital both reduce ruin (green); they must appear.
	for _, label := range []string{"Spend -5 k€/yr", "Capital +100 k€"} {
		if !strings.Contains(res.SVG, label) {
			t.Errorf("missing lever %q", label)
		}
	}
	// The FIRE UI renders dark (see theme.go), so the ruin-reducing bars use
	// the dark-theme green (#0C8A47 darkened), not the light one.
	if !strings.Contains(res.SVG, "#34A46E") {
		t.Errorf("expected at least one ruin-reducing (green) bar")
	}
}

// The risk guardrail's safe-rate table is indexed by plan year and holds the
// rate still safe for the horizon REMAINING, so a nudge that shortens the plan
// must shift it. Left stale, every year of the shorter plan quotes the band of
// a retirement five years longer, the rule cuts early and hard, and the
// "Horizon -5 y" bar reads tens of points of ruin it does not buy.
func TestSensitivityHorizonNudgeShiftsTheSafeRateTable(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 40_000, Years: 42, NPaths: 3000,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.328, GainFrac: 0.5,
		RiskGuard: true, TargetRuin: 0.05}.bounded()
	base := pr.plan()
	base.Source = pr.detailSource(nil, pr.Years)
	draws := base.Draw(3000, simWorkers, 7)
	baseRuin := base.SimulateOn(draws, simWorkers).RuinProb()

	short := base
	short.Years = pr.Years - 5
	stale := short.SimulateOn(draws, simWorkers).RuinProb()
	short.RiskGuard = shortenedTable(short.RiskGuard, 5)
	shifted := short.SimulateOn(draws, simWorkers).RuinProb()

	// Five fewer years of withdrawals under an adaptive rule is a modest move;
	// the stale table manufactured a far larger one.
	if d := baseRuin - shifted; d > 0.10 {
		t.Errorf("shifted table: %.1fpp of ruin from five fewer years, implausibly large", d*100)
	}
	if baseRuin-stale <= 2*(baseRuin-shifted) {
		t.Errorf("the stale table no longer distorts the bar: base %.4f stale %.4f shifted %.4f",
			baseRuin, stale, shifted)
	}
	t.Logf("base %.4f | stale table %.4f (%+.1fpp) | shifted table %.4f (%+.1fpp)",
		baseRuin, stale, (stale-baseRuin)*100, shifted, (shifted-baseRuin)*100)
}

// A table shorter than the nudge is left alone rather than resliced out of
// range (a one-year plan carries a one-entry table).
func TestShortenedTableToleratesAShortTable(t *testing.T) {
	g := decumul.RiskGuardrails{SafeWR: []float64{0.04, 0.05}, Band: 0.2, Cut: 0.1}
	if got := shortenedTable(g, 5); len(got.SafeWR) != 2 {
		t.Errorf("table of 2 shortened by 5: got %d entries", len(got.SafeWR))
	}
	if got := shortenedTable(g, 1); len(got.SafeWR) != 1 || got.SafeWR[0] != 0.05 {
		t.Errorf("shift by 1: %v", got.SafeWR)
	}
}
