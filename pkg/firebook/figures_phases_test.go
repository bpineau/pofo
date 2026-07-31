package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// The three-phases plate is a projection, not a measurement, so the guard is
// not "does the engine still say this" but "is the arithmetic on the plate the
// arithmetic it claims". Every test below recomputes the projection its own
// way, from the closed forms of the two recurrences, and checks the four
// statements the plate makes on its face.

// phaseSpend is the yearly spending the plan is sized on, in the same unit as
// the projection: one year of gross income, minus what is saved.
const phaseSpend = 1 - phaseSavingsRate

// The paying-in branch is a plain yearly annuity, so it has a closed form:
// after k contributions the portfolio is rate*((1+r)^k-1)/r. The drawing branch
// is an affine recurrence, so it has one too: it moves away from its fixed
// point spend*(1+r)/r at (1+r) per year. Both must reproduce the recurrence the
// figure runs, year by year.
func TestPhaseProjectionMatchesItsClosedForms(t *testing.T) {
	accum, retire := phaseProjection()
	if len(accum) != 25 || len(retire) != 41 {
		t.Fatalf("projection has %d accumulation years and %d retirement years, expected 25 and 41",
			len(accum), len(retire))
	}
	for i, p := range accum {
		k := float64(i + 1)
		if got, want := p.age, phaseStartAge+k; got != want {
			t.Fatalf("accumulation year %d is age %.0f, expected %.0f", i, got, want)
		}
		w := phaseSavingsRate * (math.Pow(1+phaseAccumReturn, k) - 1) / phaseAccumReturn
		if math.Abs(p.value-w) > 1e-9 {
			t.Errorf("age %.0f: the plate carries %.6f years of income, the annuity says %.6f", p.age, p.value, w)
		}
		if r := phaseSavingsRate / w * 100; math.Abs(p.ratio-r) > 1e-9 {
			t.Errorf("age %.0f: the plate carries %.6f %%, the annuity says %.6f %%", p.age, p.ratio, r)
		}
		// The figure draws the curve from phaseAccumRatio, so that shortcut has
		// to agree with the recurrence at every integer age.
		if r := phaseAccumRatio(p.age); math.Abs(p.ratio-r) > 1e-9 {
			t.Errorf("age %.0f: the drawn curve says %.6f %%, the recurrence says %.6f %%", p.age, r, p.ratio)
		}
	}
	fixed := phaseSpend * (1 + phaseRetireReturn) / phaseRetireReturn
	w0 := retire[0].value
	for i, p := range retire {
		if got, want := p.age, phaseDepartAge+float64(i); got != want {
			t.Fatalf("retirement year %d is age %.0f, expected %.0f", i, got, want)
		}
		w := fixed + (w0-fixed)*math.Pow(1+phaseRetireReturn, float64(i))
		if math.Abs(p.value-w) > 1e-9 {
			t.Errorf("age %.0f: the plate carries %.6f years of income, the closed form says %.6f", p.age, p.value, w)
		}
		if r := -phaseSpend / w * 100; math.Abs(p.ratio-r) > 1e-9 {
			t.Errorf("age %.0f: the plate carries %.6f %%, the closed form says %.6f %%", p.age, p.ratio, r)
		}
	}
	// The two branches share the departure year's portfolio: same capital, one
	// last contribution and a first withdrawal. That identity is the whole
	// point of the plate, so it is worth an assertion of its own.
	if a := accum[len(accum)-1]; a.age != retire[0].age || math.Abs(a.value-retire[0].value) > 1e-12 {
		t.Errorf("the branches part at %.0f/%.0f with %.6f and %.6f: they must share the departure portfolio",
			a.age, retire[0].age, a.value, retire[0].value)
	}
}

// The curve crosses zero at the departure age, and nowhere else.
func TestPhaseCrossesZeroAtTheDepartureAge(t *testing.T) {
	accum, retire := phaseProjection()
	for _, p := range accum {
		if p.ratio <= 0 {
			t.Fatalf("the paying-in branch turns negative at %.0f (%.3f %%)", p.age, p.ratio)
		}
	}
	for _, p := range retire {
		if p.ratio >= 0 {
			t.Fatalf("the drawing branch turns positive at %.0f (%.3f %%)", p.age, p.ratio)
		}
	}
	if last := accum[len(accum)-1]; last.age != phaseDepartAge {
		t.Errorf("the last contribution lands at %.0f, the article departs at %.0f", last.age, phaseDepartAge)
	}
	if first := retire[0]; first.age != phaseDepartAge {
		t.Errorf("the first withdrawal lands at %.0f, the article departs at %.0f", first.age, phaseDepartAge)
	}
}

// The paying-in branch falls every single year: the portfolio grows faster than
// the contribution, from the first year to the last. That monotonicity is the
// plate's spine, and it is what makes the transition the low point.
func TestPhaseAccumulationFallsEveryYear(t *testing.T) {
	accum, retire := phaseProjection()
	for i := 1; i < len(accum); i++ {
		if accum[i].ratio >= accum[i-1].ratio {
			t.Errorf("age %.0f: %.3f %% does not fall below the %.3f %% of the year before",
				accum[i].age, accum[i].ratio, accum[i-1].ratio)
		}
	}
	// The smallest flow of the whole plan, in either direction, is the one at
	// the transition: this is the article's "le portefeuille est le plus gros
	// par rapport aux flux", checked rather than asserted.
	best, at := math.Inf(1), 0.0
	for _, p := range append(append([]phaseYear{}, accum...), retire...) {
		if a := math.Abs(p.ratio); a < best {
			best, at = a, p.age
		}
	}
	if at != phaseDepartAge {
		t.Errorf("the smallest flow-to-portfolio ratio (%.3f %%) falls at %.0f, not at the departure age %.0f",
			best, at, phaseDepartAge)
	}
}

// After the departure the ratio settles on the withdrawal rate: it starts a
// hair above 4 % and drifts slowly, and it never comes back anywhere near the
// values of the saving years.
func TestPhaseRetirementSettlesOnTheWithdrawalRate(t *testing.T) {
	_, retire := phaseProjection()
	if first := -retire[0].ratio; math.Abs(first-4) > 0.2 {
		t.Errorf("the first withdrawal takes %.2f %% of the portfolio, the plan is sized for 4 %%", first)
	}
	for _, p := range retire {
		if r := -p.ratio; r < 4 || r > 5.5 {
			t.Errorf("age %.0f draws %.2f %% of the portfolio, outside the 4 to 5,5 %% the plate shows", p.age, r)
		}
	}
	for i := 1; i < len(retire); i++ {
		if retire[i].ratio >= retire[i-1].ratio {
			t.Errorf("age %.0f: %.3f %% does not fall below the %.3f %% of the year before",
				retire[i].age, retire[i].ratio, retire[i-1].ratio)
		}
	}
	// The departure portfolio is worth 25 years of spending, give or take: the
	// plan is sized on the 4 % rule and reaches it the year the article departs.
	if mult := retire[0].value / phaseSpend; math.Abs(mult-phaseTargetYears) > 0.5 {
		t.Errorf("the departure portfolio is worth %.2f years of spending, the plan targets %.0f",
			mult, phaseTargetYears)
	}
}

// The projection is not told the article's mid-course checkpoint, it has to
// reproduce it: "Inès, 44 ans, à 90 % de sa cible".
func TestPhaseReproducesTheArticleCheckpoint(t *testing.T) {
	accum, _ := phaseProjection()
	target := phaseSpend * phaseTargetYears
	for _, p := range accum {
		if p.age != 44 {
			continue
		}
		if share := p.value / target * 100; share < 88 || share > 94 {
			t.Errorf("at 44 the projection reaches %.1f %% of the target, the article writes 90 %%", share)
		}
		return
	}
	t.Fatal("the projection has no year 44")
}

// The left end of the curve is an artefact of the starting balance, not a fact
// about saving: the plate must cut the axis before the ratio becomes absurd,
// and the value it quotes at the cut must be the real one.
func TestPhaseAxisCutIsHonest(t *testing.T) {
	if r := phaseAccumRatio(phaseStartAge + 1); math.Abs(r-100) > 1e-9 {
		t.Errorf("one year in, the ratio is %.3f %%: the whole contribution is the portfolio, so it must be 100 %%", r)
	}
	// The claim the plate prints at the cut, and the backlog's "a saver in
	// their early years still pays in 30 to 50 % of their capital".
	if r := phaseAccumRatio(phaseFirstFigAge); r < 30 || r > 50 {
		t.Errorf("at %.0f the ratio is %.1f %%, outside the 30 to 50 %% the plate claims", phaseFirstFigAge, r)
	}
	entry := phaseAgeAtRatio(phaseAxisTop)
	if entry <= 25 || entry >= 26 {
		t.Errorf("the curve enters the plot at %.2f, outside the plotted range's first year", entry)
	}
	if r := phaseAccumRatio(entry); math.Abs(r-phaseAxisTop) > 1e-9 {
		t.Errorf("the entry age reads %.6f %%, the axis is cut at %.1f %%", r, phaseAxisTop)
	}
}

// What the plate says on its face has to be what the projection computes.
func TestPhasePlateReads(t *testing.T) {
	svg := figFluxRelatifPhases()
	accum, retire := phaseProjection()
	for _, want := range []string{
		"Projection déterministe sur les hypothèses de l'article, pas une simulation",
		"axe coupé à +15 % : le ratio vaut 49 % à 22 ans, et diverge au premier euro versé",
		"à 30 ans, le versement de l'année vaut encore 8 % du capital :",
		"45 ans : le flux change de signe,",
		fmt.Sprintf("de +%s à %s du portefeuille, du jour au lendemain.",
			frPct(accum[len(accum)-1].ratio, 1), frPct(retire[0].ratio, 1)),
		"ACCUMULATION", "TRANSITION", "RETRAIT",
		"vous versez", "vous prélevez",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// crengine paints rgba fills solid black, and the book bans the em-dash.
	if strings.Contains(svg, "rgba(") || strings.Contains(svg, "opacity") {
		t.Error("the plate uses a translucent fill: pre-blend it onto #fffdf9 instead")
	}
	if strings.Contains(svg, "—") {
		t.Error("the plate contains an em-dash")
	}
	if !strings.Contains(svg, `viewBox="0 0 640 392"`) {
		t.Error("the plate's viewBox moved: recheck the rendered PNG before freezing it")
	}
}
