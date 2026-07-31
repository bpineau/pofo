package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The vivant-ruine-parti plate freezes a whole decumulation ensemble crossed
// with a couple survival curve. This test re-runs that ensemble from
// pkg/decumul and recomputes the survival curve from the closed-form Gompertz
// law, so the plate cannot drift away from the engines: a refreshed JST
// dataset, a change in the withdrawal kernel or a re-calibrated mortality law
// breaks the build here rather than leaving a wrong picture in the book.

const (
	vrpPaths   = 200000
	vrpSeed    = 1
	vrpWorkers = 8
	vrpYears   = 53
	vrpAge     = 47.0
	// vrpTol absorbs the plate's rounding to two decimals plus any last-bit
	// floating-point difference between architectures. Seed-to-seed noise at
	// this path count is about the same size (see the file's godoc), and both
	// are an order of magnitude smaller than the effect the plate shows.
	vrpTol = 0.06
)

// vrpPlan is the reference plan the plate draws: the article's couple, at the
// withdrawal rate the same article's horizon table gives for a 50-year plan.
// It reuses errBroadCountries, the parser for the bundled JST record that the
// cout-des-erreurs plate already relies on, so both plates read one dataset the
// same way; only the horizon differs (53 years here, 50 there).
func vrpPlan() decumul.Plan {
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
			run = append(run, 0.60*e+0.40*b)
		}
		flush()
	}
	return decumul.Plan{
		Capital:    1e6,
		NeedAnnual: 33000,
		Years:      vrpYears,
		Cashflows:  []decumul.Cashflow{{FromYear: 20, Annual: 14000}},
		Flex:       decumul.FlexRule{Threshold: 0.20, Cut: 0.10},
		Source:     scenario.PooledBootstrap{Series: pool, MeanBlock: 10, Periods: vrpYears},
	}
}

func TestVivantRuinePartiMatchesTheEngine(t *testing.T) {
	surv := func(years float64) float64 {
		return decumul.FrenchMortality.CoupleSurvival(vrpAge, years)
	}

	// The survival curve is a closed form, so it must match to the rounding.
	if len(vrpAlive) != vrpYears+1 || len(vrpBroke) != vrpYears+1 {
		t.Fatalf("the plate carries %d survival and %d broke points, want %d of each",
			len(vrpAlive), len(vrpBroke), vrpYears+1)
	}
	for y, want := range vrpAlive {
		if got := surv(float64(y)) * 100; math.Abs(got-want) > 0.006 {
			t.Errorf("couple survival at year %d: Gompertz %.4f, plate %.2f", y, got, want)
		}
	}
	if got := surv(40) * 100; math.Abs(got-vrpAliveY40) > 0.006 {
		t.Errorf("survival 40 years in: Gompertz %.4f, plate %.2f", got, vrpAliveY40)
	}

	e := vrpPlan().Simulate(vrpPaths, vrpWorkers, vrpSeed)

	// Headline 1: ruin as a simulator reports it, mortality ignored.
	if got := e.RuinProb() * 100; math.Abs(got-vrpGross) > vrpTol {
		t.Errorf("gross ruin: engine %.3f, plate %.2f", got, vrpGross)
	}

	// Headline 2: the probability of EVER being alive and broke, i.e. each
	// ruined path weighted by the couple's survival to its year of failure.
	// This is the same definition the FIRE explorer reports as "ever alive and
	// broke" (pkg/decumul/web/lifecycle.go).
	lived := 0.0
	for _, p := range e.Paths {
		if p.Ruined && p.RuinYear >= 0 {
			lived += surv(float64(p.RuinYear))
		}
	}
	lived = lived / float64(len(e.Paths)) * 100
	if math.Abs(lived-vrpLived) > vrpTol {
		t.Errorf("lived ruin: engine %.3f, plate %.2f", lived, vrpLived)
	}

	// The three stacked bands, year by year.
	curve := e.LifeCurve(surv)
	if len(curve) != len(vrpBroke) {
		t.Fatalf("LifeCurve returned %d points, the plate carries %d", len(curve), len(vrpBroke))
	}
	peak, peakY := 0.0, 0
	for y, pt := range curve {
		if got := pt.Broke * 100; math.Abs(got-vrpBroke[y]) > vrpTol {
			t.Errorf("alive-and-broke share at year %d: engine %.3f, plate %.2f", y, got, vrpBroke[y])
		}
		if pt.Broke > peak {
			peak, peakY = pt.Broke, y
		}
		if sum := pt.Dead + pt.Broke + pt.Funded; math.Abs(sum-1) > 1e-9 {
			t.Errorf("the three states at year %d sum to %.6f, not 1", y, sum)
		}
	}
	if math.Abs(peak*100-vrpPeak) > vrpTol || peakY != vrpPeakY {
		t.Errorf("peak alive-and-broke: engine %.3f %% at year %d, plate %.2f %% at year %d",
			peak*100, peakY, vrpPeak, vrpPeakY)
	}

	// The counterfactual the article quotes: same failures, older readers.
	lived65 := 0.0
	for _, p := range e.Paths {
		if p.Ruined && p.RuinYear >= 0 {
			lived65 += decumul.FrenchMortality.CoupleSurvival(65, float64(p.RuinYear))
		}
	}
	lived65 = lived65 / float64(len(e.Paths)) * 100
	if math.Abs(lived65-vrpLived65) > vrpTol {
		t.Errorf("lived ruin read at 65: engine %.3f, article %.2f", lived65, vrpLived65)
	}
	if 1-vrpLived65/vrpGross < 0.5 {
		t.Error("the article claims the relief passes half at 65; the constants say otherwise")
	}

	// The plate's whole point: the weighting takes about a FIFTH off, not a
	// half and not two thirds. If this band ever moves, the article's prose
	// (and the plate's footnote) must move with it.
	relief := 1 - vrpLived/vrpGross
	if relief < 0.15 || relief > 0.25 {
		t.Errorf("the plate and its footnote claim a relief of about one fifth, the numbers say %.3f", relief)
	}
}

// The plate is only honest if the article states the same figures and the same
// caveats; the numbers no longer live in the prose alone, so guard the pair.
func TestVivantRuinePartiAgreesWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/horizon-et-esperance-de-vie.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	for _, want := range []string{
		"::: figure vivant-ruine-parti",
		"17,7 %", // gross ruin
		"14,1 %", // lived ruin
		"un cinquième",
		"deux couples sur trois",
		"7,7 %", // the same failures read for a couple aged 65
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer carries %q, which the plate asserts", want)
		}
	}
	// The old claim the measurement disproved must not creep back in.
	for _, gone := range []string{"la moitié ou au tiers", "tombe à 2,5 %"} {
		if strings.Contains(article, gone) {
			t.Errorf("the article is back to claiming %q, which the engine contradicts", gone)
		}
	}
	if svg := FigureSVG("vivant-ruine-parti"); !strings.Contains(svg, "vivant et ruiné") {
		t.Error("vivant-ruine-parti is not wired into the figures registry")
	}
}
