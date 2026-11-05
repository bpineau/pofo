package web

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

// mo is the last day of month m of year y (time.Date normalizes m past
// December), the month-end a monthly close is dated on.
func mo(y int, m time.Month) time.Time { return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC) }

func TestBuildMonthlyPanelAndFit(t *testing.T) {
	// 25 monthly points (24 returns) growing +1%/month, flat inflation.
	var pts, hicp []marketdata.Point
	v := 100.0
	for i := range 25 {
		d := mo(2000, time.January+time.Month(i))
		pts = append(pts, marketdata.Point{Date: d, Close: v})
		hicp = append(hicp, marketdata.Point{Date: d, Close: 100})
		v *= 1.01
	}
	a := AssetSeries{Weight: 1, Points: pts}
	panel, _, err := BuildMonthlyPanel([]AssetSeries{a}, hicp)
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
// covers Jan-Apr (returns Feb,Mar,Apr); asset B covers Feb-May (returns
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
	panel, _, err := BuildMonthlyPanel([]AssetSeries{a, b}, hicp)
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

// The stress-regime toggle must build an annual Markov source and, at equal
// inputs, raise ruin above the i.i.d. parametric model (sequence risk).
func TestRegimeRaisesRuin(t *testing.T) {
	base := Params{Capital: 1_000_000, NeedAnnual: 40000, Years: 40,
		Mu: 0.045, Sigma: 0.12, Df: 6, NPaths: 8000}

	if _, ok := base.source(nil).(scenario.ParametricSource); !ok {
		t.Fatal("default source should be parametric")
	}
	reg := base
	reg.Regime = true
	if _, ok := reg.source(nil).(scenario.MarkovRegime); !ok {
		t.Fatal("regime source should be a MarkovRegime")
	}

	iidRuin := Compute(base).Cards
	regRuin := Compute(reg).Cards
	ruinOf := func(cards []Card) string {
		for _, c := range cards {
			if c.Label == "Ruin" {
				return c.Value
			}
		}
		return ""
	}
	if ruinOf(iidRuin) == ruinOf(regRuin) {
		t.Errorf("regime ruin %q should differ from (exceed) i.i.d. %q", ruinOf(regRuin), ruinOf(iidRuin))
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
	panel, _, err := BuildMonthlyPanel([]AssetSeries{full, gapped}, hicp)
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

// Nominal history before the deflator starts is not real history: holding
// the first price level flat backwards would read +10 % nominal as +10 %
// real. Those months are left out, and the coverage says so.
func TestBuildMonthlyPanelStopsAtTheDeflator(t *testing.T) {
	var pts []marketdata.Point
	v := 100.0
	for i := range 6 {
		pts = append(pts, marketdata.Point{Date: mo(1954, time.October+time.Month(i)), Close: v})
		v *= 1.10
	}
	hicp := []marketdata.Point{{Date: time.Date(1955, time.January, 1, 0, 0, 0, 0, time.UTC), Close: 10}}
	panel, cov, err := BuildMonthlyPanel([]AssetSeries{{Weight: 1, Points: pts}}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	// Oct..Mar quotes: only the Feb and Mar returns start on or after 1955-01.
	if panel.Periods() != 2 {
		t.Errorf("periods = %d, want 2 (Feb, Mar 1955)", panel.Periods())
	}
	if !cov.DeflatorStart.Equal(hicp[0].Date) || len(cov.Notes()) != 1 {
		t.Errorf("coverage %+v notes %q, want the deflator start named", cov, cov.Notes())
	}
	if want := time.Date(1955, time.February, 1, 0, 0, 0, 0, time.UTC); !cov.First.Equal(want) {
		t.Errorf("first month %v, want %v", cov.First, want)
	}
}

// A shared hole splits the history into two runs; the panel keeps ONE (the
// longest), since every scenario model chains its columns as consecutive
// months, and counts what it left out.
func TestBuildMonthlyPanelKeepsOneRun(t *testing.T) {
	hicp := []marketdata.Point{{Date: mo(1999, time.January), Close: 100}}
	closes := map[time.Month]float64{
		time.January: 100, time.February: 110, // Feb: a one-month run
		time.May: 121, time.June: 133.1, time.July: 146.41, time.August: 161.051, // Jun..Aug
	}
	var pts []marketdata.Point
	for m := time.January; m <= time.August; m++ {
		if c, ok := closes[m]; ok {
			pts = append(pts, marketdata.Point{Date: mo(2025, m), Close: c})
		}
	}
	panel, cov, err := BuildMonthlyPanel([]AssetSeries{{Weight: 1, Points: pts}}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 3 || cov.GapMonths != 1 {
		t.Errorf("periods %d gap months %d, want the 3-month Jun..Aug run and 1 left out", panel.Periods(), cov.GapMonths)
	}
	if want := time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC); !cov.First.Equal(want) {
		t.Errorf("first month %v, want %v", cov.First, want)
	}
}

// A last quote a few sessions into a month is not that month's return.
func TestBuildMonthlyPanelDropsTheUnfinishedMonth(t *testing.T) {
	hicp := []marketdata.Point{{Date: mo(1999, time.January), Close: 100}}
	pts := []marketdata.Point{
		{Date: mo(2025, time.January), Close: 100},
		{Date: mo(2025, time.February), Close: 105},
		{Date: time.Date(2025, time.March, 4, 0, 0, 0, 0, time.UTC), Close: 115},
	}
	panel, cov, err := BuildMonthlyPanel([]AssetSeries{{Weight: 1, Points: pts}}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 1 || math.Abs(panel.Returns[0][0]-0.05) > 1e-12 {
		t.Errorf("panel %v, want February's +5 %% alone", panel.Returns)
	}
	if want := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC); !cov.Partial.Equal(want) {
		t.Errorf("partial %v, want %v", cov.Partial, want)
	}
}
