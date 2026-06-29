package web

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

func mo(y int, m time.Month) time.Time { return time.Date(y, m, 28, 0, 0, 0, 0, time.UTC) }

func TestBuildMonthlyPanelAndFit(t *testing.T) {
	// 25 monthly points (24 returns) growing +1%/month, flat inflation.
	var pts, hicp []marketdata.Point
	v := 100.0
	d := mo(2000, time.January)
	for i := 0; i < 25; i++ {
		pts = append(pts, marketdata.Point{Date: d, Close: v})
		hicp = append(hicp, marketdata.Point{Date: d, Close: 100})
		v *= 1.01
		d = d.AddDate(0, 1, 0)
	}
	a := AssetSeries{Weight: 1, Points: pts}
	panel, err := BuildMonthlyPanel([]AssetSeries{a}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 24 {
		t.Fatalf("monthly periods = %d, want 24", panel.Periods())
	}
	f := FitParametric(panel, []float64{1})
	// +1%/month compounds to (1.01^12 - 1) ≈ 12.68% real per year.
	want := math.Pow(1.01, 12) - 1
	if math.Abs(f.Mu-want) > 0.005 {
		t.Errorf("annualised mu = %.4f, want ~%.4f", f.Mu, want)
	}
	if f.Sigma < 0 {
		t.Errorf("sigma negative")
	}
}

// Sigma must come from the monthly dispersion scaled by √12, far more stable
// than the std of the ~20 annual points, and df must be seeded from the
// monthly excess kurtosis.
func TestFitParametricSigmaFromMonthly(t *testing.T) {
	// 24 monthly returns alternating ±5% around zero: a known monthly std.
	monthly := make([]float64, 24)
	for i := range monthly {
		if i%2 == 0 {
			monthly[i] = 0.05
		} else {
			monthly[i] = -0.05
		}
	}
	panel := scenario.Panel{Returns: [][]float64{monthly}, Weights: []float64{1}}
	f := FitParametric(panel, []float64{1})

	wantSigma := stdev(monthly) * math.Sqrt(12)
	if math.Abs(f.Sigma-wantSigma) > 1e-9 {
		t.Errorf("sigma = %.6f, want %.6f (monthly std × √12)", f.Sigma, wantSigma)
	}
}

func TestDofFromKurtosis(t *testing.T) {
	cases := []struct{ excess, want float64 }{
		{2, 7},    // 4 + 6/2
		{6, 5},    // 4 + 6/6
		{0.1, 30}, // very fat -> clamped to the slider max
		{0, 30},   // undefined / thin -> near-normal end
		{-1, 30},  // thin tails -> near-normal end
	}
	for _, c := range cases {
		if got := dofFromKurtosis(c.excess); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("dofFromKurtosis(%.2f) = %.2f, want %.2f", c.excess, got, c.want)
		}
	}
	if got := dofFromKurtosis(math.NaN()); got != 30 {
		t.Errorf("dofFromKurtosis(NaN) = %.2f, want 30", got)
	}
}

// Assets whose monthly grids do not line up (different start/end months) must
// be aligned on shared calendar months, not by trailing position. Asset A
// covers Jan–Apr (returns Feb,Mar,Apr); asset B covers Feb–May (returns
// Mar,Apr,May). The common months are Mar and Apr, so each row must hold those
// two returns in order, not three position-truncated ones.
func TestBuildMonthlyPanelDateKeyed(t *testing.T) {
	hicp := []marketdata.Point{{Date: mo(1999, time.January), Close: 100}}
	a := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.January), Close: 100},
		{Date: mo(2000, time.February), Close: 110}, // Feb +0.10
		{Date: mo(2000, time.March), Close: 132},    // Mar +0.20
		{Date: mo(2000, time.April), Close: 145.2},  // Apr +0.10
	}}
	b := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.February), Close: 200},
		{Date: mo(2000, time.March), Close: 210}, // Mar +0.05
		{Date: mo(2000, time.April), Close: 252}, // Apr +0.20
		{Date: mo(2000, time.May), Close: 277.2}, // May +0.10
	}}
	panel, err := BuildMonthlyPanel([]AssetSeries{a, b}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 2 {
		t.Fatalf("periods = %d, want 2 (Mar, Apr common)", panel.Periods())
	}
	wantA := []float64{0.20, 0.10} // Mar, Apr
	wantB := []float64{0.05, 0.20} // Mar, Apr
	for k := range wantA {
		if math.Abs(panel.Returns[0][k]-wantA[k]) > 1e-9 {
			t.Errorf("asset A month %d = %.4f, want %.4f", k, panel.Returns[0][k], wantA[k])
		}
		if math.Abs(panel.Returns[1][k]-wantB[k]) > 1e-9 {
			t.Errorf("asset B month %d = %.4f, want %.4f", k, panel.Returns[1][k], wantB[k])
		}
	}
}

// The richer-policy params map onto the plan: glidepath stop year, bounded side
// income, and guardrails banded around the initial withdrawal rate.
func TestPlanRicherPolicies(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 40000, Years: 30,
		PensionYear: 12, PensionAnnual: 18000,
		SideAnnual: 6000, SideUntilYear: 5, BufferStopYear: 7, Guardrails: true,
		BufferYears: 2}
	p := pr.plan()
	if p.Buffer.RefillStopYear != 7 {
		t.Errorf("RefillStopYear = %d, want 7", p.Buffer.RefillStopYear)
	}
	if len(p.Cashflows) != 2 {
		t.Fatalf("cashflows = %d, want 2 (pension + side income)", len(p.Cashflows))
	}
	// Side income is the bounded one.
	var side decumul.Cashflow
	for _, c := range p.Cashflows {
		if c.ToYear != 0 {
			side = c
		}
	}
	if side.Annual != 6000 || side.ToYear != 5 {
		t.Errorf("side income = %+v, want {0 5 6000}", side)
	}
	// Guardrails band centred on the 4% initial withdrawal rate.
	if p.Guard.Upper != 0.04*1.2 || p.Guard.Lower != 0.04*0.8 {
		t.Errorf("guard band = [%.4f, %.4f], want [%.4f, %.4f]", p.Guard.Lower, p.Guard.Upper, 0.032, 0.048)
	}
}

// A historical sample shorter than the horizon must raise the reliability
// caveat; an adequate sample or the parametric model must not.
func TestReliabilityCaveat(t *testing.T) {
	short := &scenario.Panel{Returns: [][]float64{make([]float64, 27*12)}, Weights: []float64{1}}
	long := &scenario.Panel{Returns: [][]float64{make([]float64, 60*12)}, Weights: []float64{1}}

	if c := reliabilityCaveat(Params{Model: "bootstrap", Years: 40}, short); c == "" {
		t.Error("27y sample, 40y horizon: expected a caveat")
	}
	if c := reliabilityCaveat(Params{Model: "bootstrap", Years: 40}, long); c != "" {
		t.Errorf("60y sample, 40y horizon: unexpected caveat %q", c)
	}
	if c := reliabilityCaveat(Params{Model: "parametric", Years: 40}, short); c != "" {
		t.Errorf("parametric model: unexpected caveat %q", c)
	}
}

// Monthly mode must yield a monthly source (Years*12 periods) and a plan that
// steps monthly; the default stays annual.
func TestMonthlySourceAndPlan(t *testing.T) {
	pr := Params{Capital: 1e6, NeedAnnual: 40000, Years: 30, Mu: 0.04, Sigma: 0.12, Df: 6, Monthly: true}
	if got := pr.source(nil).Len(); got != 30*12 {
		t.Errorf("monthly parametric source Len = %d, want %d", got, 30*12)
	}
	if !pr.plan().Monthly {
		t.Error("plan should be marked Monthly")
	}
	pr.Monthly = false
	if got := pr.source(nil).Len(); got != 30 {
		t.Errorf("annual parametric source Len = %d, want 30", got)
	}
}

// Compare must evaluate the two allocations independently, re-fitting each from
// the panel, so clearly different allocations yield clearly different outcomes.
func TestCompareAllocationsDiffer(t *testing.T) {
	// Asset 0 grows +1%/month (no ruin); asset 1 shrinks -0.5%/month (ruin).
	good, bad := make([]float64, 36), make([]float64, 36)
	for i := range good {
		good[i], bad[i] = 0.01, -0.005
	}
	panel := scenario.Panel{Returns: [][]float64{good, bad}, Weights: []float64{0.5, 0.5}}

	pr := Params{Capital: 1_000_000, NeedAnnual: 40000, BufferYears: 2, Years: 30,
		NPaths: 2000, TaxRate: 0.30, Model: "parametric", Weights: []float64{1, 0}}
	cmp := Compare(pr, []float64{0, 1}, &panel) // baseline all-bad, variant all-good

	ruin := func(res Result) string {
		for _, c := range res.Cards {
			if c.Label == "Ruin" {
				return c.Value
			}
		}
		return ""
	}
	if ruin(cmp.Baseline) == ruin(cmp.Variant) {
		t.Errorf("baseline and variant ruin both %q; allocations were not evaluated independently", ruin(cmp.Variant))
	}
	if len(cmp.Variant.Cards) == 0 || len(cmp.Baseline.Cards) == 0 {
		t.Error("both sides should produce cards")
	}
}

// Internal gaps must not produce a spanning return masquerading as a one-month
// return: a month missing from one asset drops only that month from the common
// grid, and the multi-month return across the gap is excluded.
func TestBuildMonthlyPanelSkipsGaps(t *testing.T) {
	hicp := []marketdata.Point{{Date: mo(1999, time.January), Close: 100}}
	full := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.January), Close: 100},
		{Date: mo(2000, time.February), Close: 110},
		{Date: mo(2000, time.March), Close: 121},
		{Date: mo(2000, time.April), Close: 133.1},
	}}
	gapped := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.January), Close: 100},
		{Date: mo(2000, time.February), Close: 110}, // Feb +0.10 (Jan->Feb)
		// March missing -> the Feb->Apr return spans two months, excluded.
		{Date: mo(2000, time.April), Close: 132}, // Apr (Feb->Apr, spanning)
	}}
	panel, err := BuildMonthlyPanel([]AssetSeries{full, gapped}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	// Only February is a true one-month return common to both.
	if panel.Periods() != 1 {
		t.Fatalf("periods = %d, want 1 (only Feb is consecutive in both)", panel.Periods())
	}
	if math.Abs(panel.Returns[1][0]-0.10) > 1e-9 {
		t.Errorf("gapped asset Feb return = %.4f, want 0.10", panel.Returns[1][0])
	}
}
