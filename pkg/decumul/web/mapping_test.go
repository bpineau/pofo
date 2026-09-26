package web

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Each spending rule is a distinct field of the kernel's plan, and the page
// lets only one of them be ticked at a time. This pins which field each
// checkbox reaches, and that an unticked box reaches none: a rule wired to
// the wrong field silently simulates a different household.
func TestPlanMapsEachSpendingRule(t *testing.T) {
	base := Params{Capital: 1_000_000, NeedAnnual: 40_000, Years: 30,
		Mu: 0.05, Sigma: 0.12, Df: 5}
	cases := []struct {
		name  string
		tune  func(*Params)
		check func(*testing.T, decumul.Plan)
	}{
		{"percent", func(p *Params) { p.Percent = 0.045 }, func(t *testing.T, p decumul.Plan) {
			if p.Percent != 0.045 {
				t.Errorf("Percent = %v, want 0.045", p.Percent)
			}
		}},
		{"bounded", func(p *Params) { p.Bounded = true }, func(t *testing.T, p decumul.Plan) {
			// The target share is the initial withdrawal rate, with the
			// Vanguard +5%/-2.5% yearly bounds on the spending level.
			if p.Bounded.Pct != 0.04 || p.Bounded.Up != 0.05 || p.Bounded.Down != 0.025 {
				t.Errorf("Bounded = %+v", p.Bounded)
			}
		}},
		{"abw", func(p *Params) { p.ABW = true }, func(t *testing.T, p decumul.Plan) {
			// The amortization rate is the GEOMETRIC central return, the
			// arithmetic mean less the volatility drag.
			want := 0.05 - 0.12*0.12/2
			if !p.Amortize || math.Abs(p.AmortReturn-want) > 1e-12 {
				t.Errorf("Amortize = %v, AmortReturn = %v, want %v", p.Amortize, p.AmortReturn, want)
			}
		}},
		{"guardrails", func(p *Params) { p.Guardrails = true; p.GKFloor = 0.8 },
			func(t *testing.T, p decumul.Plan) {
				if p.Guard.Floor != 0.8*40_000 || p.Guard.Cut != 0.10 {
					t.Errorf("Guard = %+v", p.Guard)
				}
			}},
		{"riskguard", func(p *Params) { p.RiskGuard = true }, func(t *testing.T, p decumul.Plan) {
			if len(p.RiskGuard.SafeWR) != 30 || p.RiskGuard.Band != 0.20 {
				t.Errorf("RiskGuard band = %v, table = %d rates", p.RiskGuard.Band, len(p.RiskGuard.SafeWR))
			}
		}},
		{"ratchet", func(p *Params) { p.Ratchet = true }, func(t *testing.T, p decumul.Plan) {
			if p.Ratchet.Step != 0.10*40_000 || p.Ratchet.Cap != 1.2*40_000 ||
				p.Ratchet.Cooldown != 2 || p.Ratchet.MaxWR != 0.022 {
				t.Errorf("Ratchet = %+v", p.Ratchet)
			}
		}},
		{"flex", func(p *Params) { p.FlexCut = 0.15; p.WRTrigger = 0.036 },
			func(t *testing.T, p decumul.Plan) {
				if p.Flex.Cut != 0.15 || p.Flex.Threshold != 0.20 || p.Flex.WRThreshold != 0.036 {
					t.Errorf("Flex = %+v", p.Flex)
				}
			}},
	}
	// Nothing ticked: every rule field stays at its zero value.
	bare := base.plan()
	if bare.Percent != 0 || bare.Amortize || bare.Bounded.Pct != 0 ||
		bare.Guard.Cut != 0 || bare.RiskGuard.SafeWR != nil || bare.Ratchet.Step != 0 {
		t.Errorf("an untouched plan carries a spending rule: %+v", bare)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pr := base
			c.tune(&pr)
			c.check(t, pr.plan())
		})
	}
}

// The band rules read the withdrawal rate off the capital, so they cannot be
// built without one: a zero capital must leave them off rather than divide by
// it.
func TestPlanRulesNeedACapital(t *testing.T) {
	pr := Params{NeedAnnual: 40_000, Years: 30, Guardrails: true, RiskGuard: true, Bounded: true}
	p := pr.plan()
	if p.Guard.Upper != 0 || p.RiskGuard.SafeWR != nil || p.Bounded.Pct != 0 {
		t.Errorf("a capital-free plan built a rate-based rule: %+v", p)
	}
}

// abwReturn is the rate the amortization rule spreads wealth at: the
// geometric central return, the CAPE-implied one under the valuation anchor,
// and never negative (a grim assumption degrades to straight-line spreading,
// it does not invert the annuity).
func TestABWReturn(t *testing.T) {
	if got, want := (Params{Mu: 0.06, Sigma: 0.10}).abwReturn(), 0.06-0.005; math.Abs(got-want) > 1e-12 {
		t.Errorf("geometric return = %v, want %v", got, want)
	}
	if got := (Params{Mu: -0.02, Sigma: 0.20}).abwReturn(); got != 0 {
		t.Errorf("grim assumption = %v, want the 0 floor", got)
	}
	if got, want := (Params{Mu: 0.06, Sigma: 0.10, CapeAdjust: true}).abwReturn(),
		math.Max(0, capeSnapshot().ImpliedReal); got != want {
		t.Errorf("CAPE-anchored return = %v, want the implied %v", got, want)
	}
}

// The spending schedule is nil while spending is constant, and otherwise
// carries one real multiplier per plan year: the health-cost drift compounds,
// and the Blanchett smile falls through the go-go years, plateaus and climbs
// back.
func TestSpendScheduleDriftAndSmile(t *testing.T) {
	if s := (Params{Years: 10}).spendSchedule(); s != nil {
		t.Errorf("constant spending built a schedule: %v", s)
	}
	drift := Params{Years: 4, SpendDrift: 0.01}.spendSchedule()
	if len(drift) != 4 || drift[0] != 1 || math.Abs(drift[3]-math.Pow(1.01, 3)) > 1e-12 {
		t.Errorf("drift schedule = %v", drift)
	}
	smile := Params{Years: 40, Smile: true}.spendSchedule()
	if len(smile) != 40 {
		t.Fatalf("smile schedule = %d years, want 40", len(smile))
	}
	// -1%/yr to year 15, a 0.85 plateau to year 25, then back up, capped at
	// 1.05 and never above it.
	for _, c := range []struct {
		k    int
		want float64
	}{
		{0, 1.0}, {10, 0.90}, {15, 0.85}, {20, 0.85}, {30, 0.91}, {39, 1.018},
	} {
		if math.Abs(smile[c.k]-c.want) > 1e-9 {
			t.Errorf("smile year %d = %.4f, want %.4f", c.k, smile[c.k], c.want)
		}
	}
	if got := smile[len(smile)-1]; got > 1.05 {
		t.Errorf("smile exceeds its cap: %v", got)
	}
	if got := smileAt(100); got != 1.05 {
		t.Errorf("smileAt(100) = %v, want the 1.05 cap", got)
	}
	// Drift and smile multiply rather than replace each other.
	both := Params{Years: 4, SpendDrift: 0.02, Smile: true}.spendSchedule()
	if want := math.Pow(1.02, 3) * smileAt(3); math.Abs(both[3]-want) > 1e-12 {
		t.Errorf("drift x smile = %v, want %v", both[3], want)
	}
}

// central() is the one place the strip selection is resolved, and it has to
// keep the legacy params (the old regime checkbox, the bootstrap/cohorts
// selector) meaning what they meant, or every shared URL changes model.
func TestCentralResolvesLegacySelectors(t *testing.T) {
	for _, c := range []struct {
		pr      Params
		want    string
		monthly bool
	}{
		{Params{}, "", true},
		{Params{Central: "broad"}, "broad", false},
		{Params{Central: "lost"}, "lost", false},
		{Params{Regime: true}, "stress", false},
		{Params{Model: "bootstrap"}, "boot", true},
		{Params{Model: "cohorts"}, "hist", true},
		{Params{Model: "parametric"}, "", true},
		// An explicit selection wins over both legacy fields.
		{Params{Central: "hist", Regime: true, Model: "bootstrap"}, "hist", true},
	} {
		if got := c.pr.central(); got != c.want {
			t.Errorf("%+v: central = %q, want %q", c.pr, got, c.want)
		}
		if got := c.pr.monthlyCapable(); got != c.monthly {
			t.Errorf("%+v: monthlyCapable = %v, want %v", c.pr, got, c.monthly)
		}
	}
	// A model with no monthly form cannot step the monthly kernel, whatever
	// the checkbox says.
	if p := (Params{Monthly: true, Regime: true, Years: 10}).plan(); p.Monthly {
		t.Error("the annual regime source must not step a monthly plan")
	}
}

// In monthly mode the panel models are resampled at monthly frequency rather
// than run annually: the source is the monthly one and its length is the
// horizon in months.
func TestSourceMonthlyPanelModels(t *testing.T) {
	panel := monthlyPanel(600)
	pr := Params{Capital: 1e6, NeedAnnual: 4e4, Years: 20, Mu: 0.04, Sigma: 0.12, Df: 5,
		Monthly: true, Weights: []float64{1}}
	for _, c := range []struct {
		central string
		want    scenario.Source
	}{
		{"boot", scenario.StationaryBootstrap{}},
		{"hist", scenario.HistoricalCohorts{}},
	} {
		pr.Central = c.central
		src := pr.source(&panel)
		if got, want := fmt.Sprintf("%T", src), fmt.Sprintf("%T", c.want); got != want {
			t.Errorf("%s monthly source = %s, want %s", c.central, got, want)
		}
		if got := src.Len(); got != 20*12 {
			t.Errorf("%s monthly source Len = %d, want %d", c.central, got, 20*12)
		}
	}
	// Without weights there is nothing to resample the panel through, so the
	// selection falls back to the annual detail source.
	pr.Central, pr.Weights = "boot", nil
	if got := pr.source(&panel).Len(); got != 20 {
		t.Errorf("weightless monthly source Len = %d, want the annual 20", got)
	}
}

// The card readouts each have a "nothing happened" form, and it must be a
// dash or a zero rather than a formatted zero pretending to be a measurement.
func TestReadoutFormatters(t *testing.T) {
	if got := fmtPctShare(0.04); got != "0%" {
		t.Errorf("fmtPctShare(0.04) = %q, want %q (rounds to nothing)", got, "0%")
	}
	if got := fmtPctShare(12.34); got != "12.3%" {
		t.Errorf("fmtPctShare(12.34) = %q", got)
	}
	if got := brokeYearsIfRuined(decumul.LifeOutcome{}); got != 0 {
		t.Errorf("no failure: broke years = %v, want 0", got)
	}
	if got := brokeYearsIfRuined(decumul.LifeOutcome{RuinAlive: 0.25, BrokeYearsMean: 1.5}); got != 6 {
		t.Errorf("conditional broke years = %v, want 6", got)
	}
	if got := firstCutText(0, 12); got != "-" {
		t.Errorf("no path cut: first cut = %q, want a dash", got)
	}
	if got := firstCutText(0.3, 12); got != "year 12" {
		t.Errorf("first cut = %q", got)
	}
	if got := cutYearsText(0, 4); got != "-" {
		t.Errorf("no path cut: cut years = %q, want a dash", got)
	}
	if got := cutYearsText(0.3, 4.2); got != "4 y" {
		t.Errorf("cut years = %q", got)
	}
	if got := floorText(nil, 0); got != "-" {
		t.Errorf("no bands: floor = %q, want a dash", got)
	}
	if got := floorText([][]float64{{32000, 31000}}, 5); got != "-" {
		t.Errorf("year past the bands: floor = %q, want a dash", got)
	}
	if got := floorText([][]float64{{32000, 31000}}, 1); got != "31.0 k€" {
		t.Errorf("floor = %q", got)
	}
}

// A vintage replay ends one of three ways, and each verdict has to say which:
// ruined (with the year, counted from one), out of record, or survived.
func TestVintageVerdict(t *testing.T) {
	for _, c := range []struct {
		want string
		got  string
	}{
		{"ruined in year 8", vintageVerdict(true, 7, 30, 30, 0)},
		{"solvent when the record ends (year 22)", vintageVerdict(false, 0, 22, 30, 1.2e6)},
		{"survived all 30 years", vintageVerdict(false, 0, 30, 30, 1.2e6)},
	} {
		if !strings.HasPrefix(c.got, c.want) {
			t.Errorf("verdict = %q, want it to start with %q", c.got, c.want)
		}
	}
}

// cashflowAt mirrors the plan's cashflow construction for the spending fan:
// the pension starts at its year and never stops, the side income runs to its
// exclusive end, and both can overlap.
func TestCashflowAt(t *testing.T) {
	pr := Params{PensionAnnual: 12_000, PensionYear: 5, SideAnnual: 6_000, SideUntilYear: 6}
	for _, c := range []struct {
		year int
		want float64
	}{
		{0, 6_000}, {4, 6_000}, {5, 18_000}, {6, 12_000}, {40, 12_000},
	} {
		if got := pr.cashflowAt(c.year); got != c.want {
			t.Errorf("year %d: cashflow = %.0f, want %.0f", c.year, got, c.want)
		}
	}
	if got := (Params{}).cashflowAt(3); got != 0 {
		t.Errorf("no income: cashflow = %.0f, want 0", got)
	}
}

// The mortality age defaults to an early retiree rather than to zero, which
// would price an annuity for a newborn.
func TestAgeDefault(t *testing.T) {
	if got := (Params{}).age(); got != 52 {
		t.Errorf("default age = %v, want 52", got)
	}
	if got := (Params{Age: 64}).age(); got != 64 {
		t.Errorf("age = %v, want 64", got)
	}
}
