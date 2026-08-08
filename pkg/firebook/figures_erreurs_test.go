package firebook

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The cout-des-erreurs plate ranks six construction errors by the ruin they add
// to one shared plan. Its numbers are frozen literals so the figure stays a pure
// function; this test recomputes every one of them from pkg/decumul and the
// bundled broad-sample record, and fails the moment the plate and the engine
// disagree. A refreshed JST dataset or a change in the withdrawal kernel
// therefore breaks the build here rather than leaving a wrong ranking in the
// book.

// errTestPaths/errTestSeed/errTestWorkers are the plate's reproduction recipe.
// The worker count matters: DrawPaths splits the sampling per worker, so the
// same seed with a different worker count is a different Monte-Carlo.
const (
	errTestPaths   = 200000
	errTestSeed    = 1
	errTestWorkers = 8
	errTestYears   = 50
	// tol absorbs the plate's rounding to two decimals plus any last-bit
	// floating-point difference between architectures; it is still two orders
	// of magnitude tighter than the gaps the plate ranks.
	errTestTol = 0.05
)

// errRefPlan is the reference plan the whole plate is measured against.
func errRefPlan() decumul.Plan {
	return decumul.Plan{
		Capital:    1e6,
		NeedAnnual: 35000,
		Years:      errTestYears,
		Cashflows:  []decumul.Cashflow{{FromYear: 20, Annual: 14000}},
		Flex:       decumul.FlexRule{Threshold: 0.20, Cut: 0.10},
		Source:     errBroadSource(0.60),
	}
}

func TestErreursPlateMatchesTheEngine(t *testing.T) {
	frozenAgainstData(t)
	ruin := func(p decumul.Plan) float64 {
		return p.Simulate(errTestPaths, errTestWorkers, errTestSeed).RuinProb() * 100
	}
	if got := ruin(errRefPlan()); math.Abs(got-errCostRef) > errTestTol {
		t.Errorf("reference plan ruin = %.3f, plate says %.2f", got, errCostRef)
	}

	// Each variant changes exactly one field of the reference plan.
	for _, tc := range []struct {
		label  string
		mutate func(decumul.Plan) decumul.Plan
	}{
		{"Portefeuille mono-régime, tout obligataire",
			func(p decumul.Plan) decumul.Plan { p.Source = errBroadSource(0); return p }},
		{"Flexibilité surestimée",
			func(p decumul.Plan) decumul.Plan { p.NeedAnnual = 45000; return p }},
		{"Dépenses sous-estimées de 20 %",
			func(p decumul.Plan) decumul.Plan { p.NeedAnnual = 42000; return p }},
		{"Pension oubliée",
			func(p decumul.Plan) decumul.Plan { p.Cashflows = nil; return p }},
		{"Taux de 4 % au lieu de 3,5 %",
			func(p decumul.Plan) decumul.Plan { p.NeedAnnual = 40000; return p }},
		{"Fiscalité et PUMa ignorées",
			func(p decumul.Plan) decumul.Plan { p.NeedAnnual = 39200; return p }},
	} {
		var want float64
		found := false
		for _, e := range errCosts {
			if e.label == tc.label {
				want, found = e.ruin, true
			}
		}
		if !found {
			t.Errorf("the plate no longer carries a bar labelled %q", tc.label)
			continue
		}
		if got := ruin(tc.mutate(errRefPlan())); math.Abs(got-want) > errTestTol {
			t.Errorf("%s: engine %.3f, plate %.2f", tc.label, got, want)
		}
	}

	// The plate is sorted, most expensive first, and every bar costs something.
	for i, e := range errCosts {
		if e.ruin <= errCostRef {
			t.Errorf("%s: ruin %.2f is not above the reference %.2f", e.label, e.ruin, errCostRef)
		}
		if i > 0 && e.ruin > errCosts[i-1].ruin {
			t.Errorf("bar %d (%s) breaks the sort order", i, e.label)
		}
	}

	// The all-equity mirror quoted in the footnote: lower ruin, more cut years.
	eq := errRefPlan()
	eq.Source = errBroadSource(1)
	for _, tc := range []struct {
		name       string
		plan       decumul.Plan
		ruin, cuts float64
	}{
		{"60/40", errRefPlan(), errCostRef, errCostCutRef},
		{"100 % actions", eq, errCostEquity, errCostCutEquity},
	} {
		e := tc.plan.Simulate(errTestPaths, errTestWorkers, errTestSeed)
		if got := e.RuinProb() * 100; math.Abs(got-tc.ruin) > errTestTol {
			t.Errorf("%s ruin: engine %.3f, plate %.2f", tc.name, got, tc.ruin)
		}
		cuts := 0.0
		for _, p := range e.Paths {
			cuts += float64(p.CutYears)
		}
		if got := cuts / float64(len(e.Paths)); math.Abs(got-tc.cuts) > errTestTol {
			t.Errorf("%s cut years: engine %.3f, plate %.2f", tc.name, got, tc.cuts)
		}
	}
	if errCostEquity >= errCostRef || errCostCutEquity <= errCostCutRef {
		t.Error("the footnote claims the all-equity plan ruins less but cuts more; the constants say otherwise")
	}
}

// The tax bar is the only variant whose size comes from the article's own prose
// rather than from a model: the household spends the 12 % friction it did not
// budget. This checks the arithmetic and that the article still states the band
// that 12 % sits in.
func TestErreursFrictionMatchesTheArticle(t *testing.T) {
	frozenAgainstData(t)
	if got := 35000 * 1.12; math.Abs(got-39200) > 1e-9 {
		t.Errorf("35 k€ raised by 12 %% = %.0f, the plate says 39 200", got)
	}
	raw, err := assets.ReadFile("assets/book/fr/erreurs-classiques-fire.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"8 à 20 %", "10-15 %"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("the article no longer states the friction band %q the plate's 12 %% comes from", want)
		}
	}
}

// --- the broad-sample pool, mirroring pkg/decumul/web/broadsample.go ---

// errBroadSource pool-bootstraps annual real-return paths from the bundled
// per-country JST record, each country blended at equity weight w and cut at
// the gaps in its bond record (war breaks), exactly as the FIRE explorer's
// empirical model does at w = 0.60.
func errBroadSource(w float64) scenario.Source {
	var pool [][]float64
	for _, c := range errBroadCountries() {
		var run []float64
		flush := func() {
			if len(run) >= 10 { // shorter stubs hold no useful block
				pool = append(pool, run)
			}
			run = nil
		}
		for i, e := range c.equity {
			b := c.bond[i]
			if math.IsNaN(b) {
				flush()
				continue
			}
			run = append(run, w*e+(1-w)*b)
		}
		flush()
	}
	return scenario.PooledBootstrap{Series: pool, MeanBlock: 10, Periods: errTestYears}
}

type errCountry struct {
	equity, bond []float64
}

// errBroadCountries parses the embedded iso,year,equity,bond,bill table into one
// series per country, in file order (ascending year within a country).
func errBroadCountries() []errCountry {
	byISO := map[string]*errCountry{}
	var order []string
	for _, line := range strings.Split(string(datasets.BroadSample()), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 4 || f[2] == "" {
			continue
		}
		c, seen := byISO[f[0]]
		if !seen {
			c = &errCountry{}
			byISO[f[0]], order = c, append(order, f[0])
		}
		e, err := strconv.ParseFloat(f[2], 64)
		if err != nil {
			continue
		}
		bond := math.NaN()
		if f[3] != "" {
			if v, err := strconv.ParseFloat(f[3], 64); err == nil {
				bond = v
			}
		}
		c.equity, c.bond = append(c.equity, e), append(c.bond, bond)
	}
	out := make([]errCountry, 0, len(order))
	for _, iso := range order {
		out = append(out, *byISO[iso])
	}
	return out
}
