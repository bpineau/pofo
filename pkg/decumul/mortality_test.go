package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestGompertzSurvivalShape(t *testing.T) {
	g := FrenchMortality
	if s := g.Survival(50, 0); math.Abs(s-1) > 1e-12 {
		t.Errorf("Survival(50, 0) = %.4f, want 1", s)
	}
	prev := 1.0
	for years := 1.0; years <= 50; years++ {
		s := g.Survival(50, years)
		if s > prev {
			t.Fatalf("survival must decrease: S(%v) = %.4f > %.4f", years, s, prev)
		}
		prev = s
	}
	// A 50-year-old reaching 100 should be rare but possible.
	if s := g.Survival(50, 50); s < 0.001 || s > 0.10 {
		t.Errorf("Survival(50, 50) = %.4f, want a small but nonzero share", s)
	}
	// Median remaining life for a 50-year-old is in the mid-30s of years.
	if s := g.Survival(50, 35); s < 0.35 || s > 0.65 {
		t.Errorf("Survival(50, 35) = %.4f, want around one half", s)
	}
}

// inseeRemaining is the remaining life expectancy (complete expectation, in
// years) of INSEE's triennial France table for 2023-2025, provisional results
// published 2026-01-13 in "Bilan démographique 2025", file
// 6_FR_Table_mortalite_triennale.xlsx of https://www.insee.fr/fr/statistiques/8721189:
// age, both sexes ("Ensemble"), women, men.
var inseeRemaining = []struct {
	age, both, women, men float64
}{
	{50, 34.68, 37.05, 32.37},
	{55, 30.19, 32.42, 27.99},
	{60, 25.86, 27.91, 23.81},
	{65, 21.73, 23.55, 19.87},
	{70, 17.81, 19.34, 16.19},
	{75, 14.08, 15.29, 12.73},
	{80, 10.59, 11.47, 9.52},
	{85, 7.50, 8.10, 6.70},
	{90, 5.03, 5.39, 4.47},
}

// tgf05Gen1978 is the regulatory TGF05 table (arrêté of 1 August 2006), for
// the generation born in 1978: the remaining life expectancy at age x
// (curtate expectation from the published survivor counts, plus half a year)
// and the joint-life annuity-due factor at a 1 % real rate over two lives of
// that age, both read off the table itself.
var tgf05Gen1978 = []struct {
	age, remaining, joint float64
}{
	{60, 36.26, 34.43},
	{62, 34.40, 33.12},
	{65, 31.62, 31.12},
	{70, 27.00, 27.65},
	{75, 22.45, 24.05},
	{80, 18.01, 0},
	{90, 10.08, 0},
}

// FrenchAnnuitants must price what the regulatory table prices: within half a
// year of its remaining life expectancy from 60 to 90, and within half a
// percent of its joint-life annuity factor at the ages an annuity is bought.
func TestFrenchAnnuitantsMatchesTGF05(t *testing.T) {
	for _, row := range tgf05Gen1978 {
		if got := FrenchAnnuitants.LifeExpectancy(row.age); math.Abs(got-row.remaining) > 0.5 {
			t.Errorf("at %v: %.2f years left, TGF05 %.2f", row.age, got, row.remaining)
		}
		if row.joint == 0 {
			continue
		}
		if got := AnnuityFactor(FrenchAnnuitants, row.age, 0.01); math.Abs(got/row.joint-1) > 0.005 {
			t.Errorf("at %v: joint annuity factor %.2f, TGF05 %.2f", row.age, got, row.joint)
		}
	}
}

// The bundled laws must reproduce INSEE's remaining life expectancy within
// half a year at every age from 50 to 90, for both sexes and each sex: a law
// that runs short flatters every plan it drives.
func TestBundledLawsMatchINSEE(t *testing.T) {
	const tol = 0.5
	for _, row := range inseeRemaining {
		for _, c := range []struct {
			name string
			law  Gompertz
			want float64
		}{
			{"FrenchMortality", FrenchMortality, row.both},
			{"FrenchWomen", FrenchWomen, row.women},
			{"FrenchMen", FrenchMen, row.men},
		} {
			got := c.law.LifeExpectancy(row.age)
			if math.Abs(got-c.want) > tol {
				t.Errorf("%s at %v: %.2f years left, INSEE %.2f (tolerance %.1f)", c.name, row.age, got, c.want, tol)
			}
		}
	}
}

// LifeExpectancy is the integral of the survival curve: check it against a
// brute-force trapezoid, and against the kernel's own count of yearly
// payments (AnnuityFactor at a zero rate on a single life), which the
// Euler-Maclaurin formula puts half a year above the integral.
func TestGompertzLifeExpectancy(t *testing.T) {
	g := Gompertz{Mode: 88, Dispersion: 10}
	brute, h := 0.0, 0.001
	for x := 0.0; x < 80; x += h {
		brute += h * (g.Survival(65, x) + g.Survival(65, x+h)) / 2
	}
	if got := g.LifeExpectancy(65); math.Abs(got-brute) > 1e-4 {
		t.Errorf("LifeExpectancy(65) = %.5f, trapezoid %.5f", got, brute)
	}
	payments := AnnuityFactorFor(func(years float64) float64 { return g.Survival(65, years) }, 0)
	if d := payments - g.LifeExpectancy(65); math.Abs(d-0.5) > 0.02 {
		t.Errorf("expected payments %.3f minus life expectancy = %.3f, want about 0.5", payments, d)
	}
	if g.LifeExpectancy(90) >= g.LifeExpectancy(60) {
		t.Error("remaining life expectancy must fall with age")
	}
}

// A couple (first death does not end the household) survives longer than a
// single person at every horizon.
func TestCoupleSurvivalDominates(t *testing.T) {
	g := FrenchMortality
	for years := 5.0; years <= 45; years += 10 {
		single, couple := g.Survival(50, years), g.CoupleSurvival(50, years)
		if couple < single {
			t.Errorf("couple survival %.4f < single %.4f at %v years", couple, single, years)
		}
		if couple > 1 {
			t.Errorf("couple survival %.4f > 1", couple)
		}
	}
}

// LifeCurve splits each year-end into dead / broke-but-alive / funded-alive,
// summing to 1, with the broke share tracking cumulative ruin among survivors.
func TestLifeCurve(t *testing.T) {
	e := Ensemble{Years: 3, Paths: []PathResult{
		{Wealth: []float64{1, 1, 1, 1}, RuinYear: -1},
		{Wealth: []float64{1, 0, 0, 0}, Ruined: true, RuinYear: 1},
	}}
	// A toy survival: everyone alive to year 2, half alive at year 3.
	surv := func(years float64) float64 {
		if years < 3 {
			return 1
		}
		return 0.5
	}
	pts := e.LifeCurve(surv)
	if len(pts) != 4 {
		t.Fatalf("len = %d, want Years+1 = 4", len(pts))
	}
	for i, pt := range pts {
		if s := pt.Dead + pt.Broke + pt.Funded; math.Abs(s-1) > 1e-9 {
			t.Errorf("point %d shares sum to %.4f, want 1", i, s)
		}
	}
	// Year 0: nobody dead, nobody broke yet.
	if pts[0].Broke != 0 || pts[0].Dead != 0 {
		t.Errorf("year 0 = %+v, want all funded", pts[0])
	}
	// Year 2: path 2 is ruined (ruin year 1 <= 2), everyone alive: broke 50%.
	if math.Abs(pts[2].Broke-0.5) > 1e-9 {
		t.Errorf("year 2 broke = %.3f, want 0.5", pts[2].Broke)
	}
	// Year 3: half died; the broke share is halved with them.
	if math.Abs(pts[3].Dead-0.5) > 1e-9 || math.Abs(pts[3].Broke-0.25) > 1e-9 {
		t.Errorf("year 3 = %+v, want dead 0.5, broke 0.25", pts[3])
	}
}

// RuinYearHistogram counts the share of all paths ruined in each year.
func TestRuinYearHistogram(t *testing.T) {
	e := Ensemble{Years: 5, Paths: []PathResult{
		{RuinYear: -1}, {RuinYear: 2, Ruined: true},
		{RuinYear: 2, Ruined: true}, {RuinYear: 4, Ruined: true},
	}}
	h := e.RuinYearHistogram()
	if math.Abs(h[2]-0.5) > 1e-9 {
		t.Errorf("h[2] = %.3f, want 0.5", h[2])
	}
	if math.Abs(h[4]-0.25) > 1e-9 {
		t.Errorf("h[4] = %.3f, want 0.25", h[4])
	}
	if h[0] != 0 || h[1] != 0 || h[3] != 0 {
		t.Errorf("unexpected nonzero buckets: %v", h)
	}
}

// The kernel records the first cut year and the number of cut years when the
// flex rule bites.
func TestRunPathCutAccounting(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 4,
		Flex: FlexRule{Threshold: 0.20, Cut: 0.25}, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{-0.5, 0, 1.2, 0}, Lives{})
	// Year 0 full spend; years 1 and 2 in deep drawdown (cut); year 3 back
	// above the -20% drawdown line after the +120% year.
	if res.FirstCut != 1 {
		t.Errorf("FirstCut = %d, want 1", res.FirstCut)
	}
	if res.CutYears != 2 {
		t.Errorf("CutYears = %d, want 2", res.CutYears)
	}

	nocut := Plan{Capital: 100000, NeedAnnual: 4000, Years: 2, Tax: CTOFlatTax{Rate: 0}}
	if r := nocut.RunPath(scenario.Sequence{0, 0}, Lives{}); r.FirstCut != -1 || r.CutYears != 0 {
		t.Errorf("FirstCut = %d CutYears = %d, want -1 and 0 without a flex rule", r.FirstCut, r.CutYears)
	}
}

// SpendStats aggregates the cut accounting across paths.
func TestSpendStats(t *testing.T) {
	e := Ensemble{Years: 10, Paths: []PathResult{
		{FirstCut: -1}, {FirstCut: -1},
		{FirstCut: 2, CutYears: 3}, {FirstCut: 6, CutYears: 5},
	}}
	s := e.SpendStats()
	if math.Abs(s.EverCutShare-0.5) > 1e-9 {
		t.Errorf("EverCutShare = %.3f, want 0.5", s.EverCutShare)
	}
	if s.FirstCutMedian < 2 || s.FirstCutMedian > 6 {
		t.Errorf("FirstCutMedian = %.1f, want between 2 and 6", s.FirstCutMedian)
	}
	if s.CutYearsMedian < 3 || s.CutYearsMedian > 5 {
		t.Errorf("CutYearsMedian = %.1f, want between 3 and 5", s.CutYearsMedian)
	}
}

// SpendBands gives per-year quantiles of the delivered spending, the spending
// counterpart of the wealth fan.
func TestSpendBands(t *testing.T) {
	e := Ensemble{Years: 2, Paths: []PathResult{
		{Spend: []float64{100, 100}},
		{Spend: []float64{100, 50}},
		{Spend: []float64{100, 0}},
	}}
	bands := e.SpendBands([]float64{0.5})
	if len(bands) != 1 || len(bands[0]) != 2 {
		t.Fatalf("bands shape = %dx%d, want 1x2", len(bands), len(bands[0]))
	}
	if math.Abs(bands[0][1]-50) > 1e-9 {
		t.Errorf("median spend year 1 = %.1f, want 50", bands[0][1])
	}
}
