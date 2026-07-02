package web

import (
	"strings"
	"testing"
)

func testParams() Params {
	return Params{
		Capital: 1_800_000, NeedAnnual: 60000, BufferYears: 2,
		Mu: 0.05, Sigma: 0.11, Df: 5, Years: 40,
		PensionYear: 15, PensionAnnual: 20000, TaxRate: 0.314,
		FlexCut: 0.15, NPaths: 400, Age: 52,
	}
}

// The spending endpoint returns the delivered-spending fan and the lived cost
// of the adaptive policy (how often, when and for how long the household cut).
func TestSpending(t *testing.T) {
	r := Spending(testParams(), nil)
	if !strings.HasPrefix(r.SVG, "<svg") {
		t.Fatalf("SVG missing: %.30q", r.SVG)
	}
	if len(r.Cards) < 3 {
		t.Errorf("cards = %d, want the cut statistics", len(r.Cards))
	}
	var found bool
	for _, c := range r.Cards {
		if strings.Contains(c.Label, "cut") || strings.Contains(c.Label, "Cut") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a cut statistic card, got %+v", r.Cards)
	}
}

// The lifecycle endpoint returns the alive-broke-dead stacked area and the
// ruin-year histogram.
func TestLifecycle(t *testing.T) {
	r := Lifecycle(testParams(), nil)
	if !strings.HasPrefix(r.LifeSVG, "<svg") {
		t.Fatalf("LifeSVG missing: %.30q", r.LifeSVG)
	}
	if !strings.HasPrefix(r.RuinYearSVG, "<svg") {
		t.Fatalf("RuinYearSVG missing: %.30q", r.RuinYearSVG)
	}
	// The stacked chart should name the three states.
	for _, s := range []string{"Funded", "Broke", "Gone"} {
		if !strings.Contains(r.LifeSVG, s) {
			t.Errorf("LifeSVG misses the %q layer", s)
		}
	}
}

// The curves endpoint returns safe-WR-vs-horizon and capital-vs-spending.
func TestCurves(t *testing.T) {
	pr := testParams()
	pr.NPaths = 300 // two solves per point: keep the test quick
	r := Curves(pr, nil)
	if !strings.HasPrefix(r.HorizonSVG, "<svg") {
		t.Fatalf("HorizonSVG missing: %.30q", r.HorizonSVG)
	}
	if !strings.HasPrefix(r.CapitalSVG, "<svg") {
		t.Fatalf("CapitalSVG missing: %.30q", r.CapitalSVG)
	}
}

// plan() must translate the envelope, ratchet, WR-trigger and schedule params
// into the decumul plan.
func TestPlanBuildsEnrichedPolicy(t *testing.T) {
	pr := testParams()
	pr.PEACapital = 175000
	pr.AVCapital = 50000
	pr.GainFrac = 0.4
	pr.Ratchet = true
	pr.WRTrigger = 0.036
	pr.SpendDrift = 0.003

	p := pr.plan()
	if len(p.Envelopes) != 3 {
		t.Fatalf("envelopes = %d, want 3 (CTO, PEA, AV)", len(p.Envelopes))
	}
	if p.Envelopes[0].Name != "CTO" || p.Envelopes[1].Name != "PEA" || p.Envelopes[2].Name != "AV" {
		t.Errorf("drain order = %v, want CTO, PEA, AV", []string{p.Envelopes[0].Name, p.Envelopes[1].Name, p.Envelopes[2].Name})
	}
	if p.Envelopes[1].Amount != 175000 {
		t.Errorf("PEA amount = %.0f, want 175000", p.Envelopes[1].Amount)
	}
	if p.Ratchet.Trigger == 0 || p.Ratchet.Step == 0 {
		t.Errorf("ratchet not configured: %+v", p.Ratchet)
	}
	if p.Flex.WRThreshold != 0.036 {
		t.Errorf("WRThreshold = %v, want 0.036", p.Flex.WRThreshold)
	}
	if len(p.SpendSchedule) != pr.Years {
		t.Errorf("schedule length = %d, want %d", len(p.SpendSchedule), pr.Years)
	}
}

// Without the new params, the plan keeps the legacy single sleeve and constant
// spending (no envelopes, no schedule, no ratchet).
func TestPlanLegacyDefaults(t *testing.T) {
	p := testParams().plan()
	if p.Envelopes != nil {
		t.Errorf("expected nil envelopes, got %+v", p.Envelopes)
	}
	if p.SpendSchedule != nil {
		t.Errorf("expected nil schedule")
	}
	if p.Ratchet.Trigger != 0 {
		t.Errorf("expected inactive ratchet")
	}
}
