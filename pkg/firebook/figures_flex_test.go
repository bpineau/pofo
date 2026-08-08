package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/replay"
	"github.com/bpineau/pofo/pkg/scenario"
)

// coupeSchedule builds the real spending multipliers of one candidate answer:
// the full standard of living everywhere, except a cut of depth d (0.15 = 15 %)
// held for dur years from the year the household reacts.
func coupeSchedule(dur int, d float64) []float64 {
	s := make([]float64, coupeYears)
	for i := range s {
		s[i] = 1
	}
	k := coupeTrigger - coupeVintage
	for i := k; i < k+dur && i < coupeYears; i++ {
		s[i] = 1 - d
	}
	return s
}

// coupeSurvives replays one candidate answer through the vintage and reports
// whether the plan paid every year it promised.
func coupeSurvives(seq scenario.Sequence, spend float64, dur int, d float64) bool {
	p := decumul.Plan{
		Capital: coupeCapital, NeedAnnual: spend, Years: coupeYears,
		SpendSchedule: coupeSchedule(dur, d),
	}
	return !p.RunPath(seq).Ruined
}

// coupeMinDepth is the plate's solver: the shallowest cut that survives the
// horizon when held for dur years, found by bisection. Survival is monotone in
// the depth (a deeper cut leaves strictly more capital), which is what makes
// the bisection legitimate. ok is false when no depth works at all, total
// abstinence included.
func coupeMinDepth(seq scenario.Sequence, spend float64, dur int) (depth float64, ok bool) {
	if coupeSurvives(seq, spend, dur, 0) {
		return 0, true
	}
	if !coupeSurvives(seq, spend, dur, 1) {
		return 0, false
	}
	lo, hi := 0.0, 1.0
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if coupeSurvives(seq, spend, dur, mid) {
			hi = mid
		} else {
			lo = mid
		}
	}
	return hi, true
}

func coupeSequence(t *testing.T) scenario.Sequence {
	t.Helper()
	frozenAgainstData(t)
	ref, err := replay.Reference()
	if err != nil {
		t.Fatal(err)
	}
	_, _, seq, ok := ref.Window(coupeVintage, coupeYears)
	if !ok {
		t.Fatalf("the bundled record does not cover a %d-year retirement starting in %d", coupeYears, coupeVintage)
	}
	if len(seq) != coupeYears {
		t.Fatalf("the %d vintage only covers %d years, the plate draws %d and never extrapolates",
			coupeVintage, len(seq), coupeYears)
	}
	return seq
}

// Every depth the plate freezes is re-solved from pkg/replay and pkg/decumul.
func TestCoupeExigeeMatchesTheSolver(t *testing.T) {
	seq := coupeSequence(t)
	for _, tc := range []struct {
		name   string
		spend  float64
		first  int
		frozen []float64
	}{
		{"4,5 %", coupeSpend45, coupeFirst45, coupeDepth45},
		{"4,0 %", coupeSpend40, coupeFirst40, coupeDepth40},
	} {
		for i, want := range tc.frozen {
			dur := tc.first + i
			got, ok := coupeMinDepth(seq, tc.spend, dur)
			if !ok {
				t.Errorf("%s, %d ans: the solver finds no depth at all, the plate freezes %.1f %%", tc.name, dur, want)
				continue
			}
			if math.Abs(got*100-want) > 0.05 {
				t.Errorf("%s, %d ans: solver %.2f %%, plate %.1f %%", tc.name, dur, got*100, want)
			}
		}
		// The curve runs to the end of the retirement and nowhere further: the
		// cut cannot outlast the plan it is saving.
		if last := tc.first + len(tc.frozen) - 1; last != coupeYears-(coupeTrigger-coupeVintage) {
			t.Errorf("%s: the plate stops at %d years, the plan leaves %d years after the trigger",
				tc.name, last, coupeYears-(coupeTrigger-coupeVintage))
		}
		// A requirement that eases with duration is the whole geometry of the
		// plate; a bump would mean the solver, not the market, moved.
		for i := 1; i < len(tc.frozen); i++ {
			if tc.frozen[i] > tc.frozen[i-1] {
				t.Errorf("%s: the frozen curve rises at %d years (%.1f after %.1f)",
					tc.name, tc.first+i, tc.frozen[i], tc.frozen[i-1])
			}
		}
	}
}

// The 4,5 % plan's opening claim: below four years there is no answer at all,
// and total abstinence is not one either.
func TestCoupeImpossibleDurations(t *testing.T) {
	seq := coupeSequence(t)
	for dur := 1; dur <= coupeImpossible; dur++ {
		if _, ok := coupeMinDepth(seq, coupeSpend45, dur); ok {
			t.Errorf("%d ans: the solver saves the 4,5 %% plan, the plate draws it off the scale", dur)
		}
	}
	if coupeFirst45 != coupeImpossible+1 {
		t.Errorf("the curve starts at %d years, the plate calls %d years impossible", coupeFirst45, coupeImpossible)
	}
	if _, ok := coupeMinDepth(seq, coupeSpend45, coupeFirst45); !ok {
		t.Errorf("%d ans has no answer either, the plate opens its curve there", coupeFirst45)
	}
	// The plate is about a plan that needs saving: both would die untouched.
	for _, spend := range []float64{coupeSpend45, coupeSpend40} {
		if coupeSurvives(seq, spend, 0, 0) {
			t.Errorf("%.0f EUR a year survives %d untouched, the plate asks what it takes to save it", spend, coupeVintage)
		}
	}
}

// The trigger is the article's own written rule ("drawdown > 20 %"), read on
// the wealth the household watches: the cut starts the year after the first
// year-end more than 20 % below the running peak.
func TestCoupeTriggerFollowsTheWrittenRule(t *testing.T) {
	seq := coupeSequence(t)
	p := decumul.Plan{Capital: coupeCapital, NeedAnnual: coupeSpend45, Years: coupeYears}
	w := p.RunPath(seq).Wealth
	peak, fired := w[0], -1
	for i := 1; i < len(w) && fired < 0; i++ {
		peak = math.Max(peak, w[i])
		if w[i] < 0.8*peak {
			fired = i
		}
	}
	if fired < 0 {
		t.Fatal("the untouched 4,5 % plan never sits 20 % below its peak, the plate's trigger cannot fire")
	}
	if got := coupeVintage + fired; got != coupeTrigger {
		t.Errorf("the 20 %% rule fires at the end of %d, so the cut starts in %d, the plate says %d",
			coupeVintage+fired-1, got, coupeTrigger)
	}
	// The plate's godoc claims the answer barely depends on the trigger; check
	// it, since that claim is what licenses a single curve.
	for _, k := range []int{0, coupeTrigger - coupeVintage + 5} {
		alt := func(dur int, d float64) bool {
			s := make([]float64, coupeYears)
			for i := range s {
				s[i] = 1
			}
			for i := k; i < k+dur && i < coupeYears; i++ {
				s[i] = 1 - d
			}
			q := decumul.Plan{Capital: coupeCapital, NeedAnnual: coupeSpend45, Years: coupeYears, SpendSchedule: s}
			return !q.RunPath(seq).Ruined
		}
		lo, hi := 0.0, 1.0
		if !alt(10, 1) {
			t.Fatalf("trigger year %d: a total stop over ten years cannot save the plan", coupeVintage+k)
		}
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if alt(10, mid) {
				hi = mid
			} else {
				lo = mid
			}
		}
		if d := math.Abs(hi*100 - coupeDepth45[10-coupeFirst45]); d > 3 {
			t.Errorf("moving the trigger to %d moves the ten-year depth by %.1f points, the plate claims at most three",
				coupeVintage+k, d)
		}
	}
}

// The two verdicts printed under the panels, checked against the frozen curves
// and the chapter's own tolerance: the tense plan never gets its requirement
// inside the tenable zone, the reasonable one does from the third year.
func TestCoupeVerdicts(t *testing.T) {
	if coupeTenableEasy >= coupeTenableWall {
		t.Fatalf("the tenable zone is empty: %d years with effort, %d years never", coupeTenableEasy, coupeTenableWall)
	}
	for i, d := range coupeDepth45 {
		if dur := coupeFirst45 + i; dur <= coupeTenableWall && d <= coupeTenableDepth {
			t.Errorf("the 4,5 %% plan asks only %.1f %% for %d years, which is inside the tenable zone", d, dur)
		}
	}
	if last := coupeDepth45[len(coupeDepth45)-1]; last > coupeTenableDepth {
		t.Errorf("even at twenty-six years the 4,5 %% plan asks %.1f %%, the plate says it comes back down to 15 %%", last)
	}
	cross := -1
	for i, d := range coupeDepth40 {
		if d <= coupeTenableDepth {
			cross = coupeFirst40 + i
			break
		}
	}
	if cross != 3 {
		t.Errorf("the 4,0 %% requirement enters the tenable zone at %d years, the plate's verdict says three", cross)
	}
	if cross > coupeTenableWall {
		t.Errorf("the crossing at %d years is past the chapter's %d-year wall", cross, coupeTenableWall)
	}
}

// The tenable zone is the CHAPTER'S claim, not a measurement, so it must keep
// matching the sentence it is drawn from.
func TestCoupeTenableZoneMatchesTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/flexibilite-realite.md")
	if err != nil {
		t.Fatal(err)
	}
	md := string(raw)
	for _, want := range []string{
		"Une coupe de 15 % se tient dix-huit mois avec du moral, cinq ans avec effort, jamais douze ans.",
		"::: figure coupe-exigee-tenable",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("the article no longer contains %q, the plate is drawn from it", want)
		}
	}
	if coupeTenableDepth != 15 || coupeTenableEasy != 5 || coupeTenableWall != 12 {
		t.Errorf("the zone reads %.0f %% / %d ans / %d ans, the sentence says 15 %% / 5 ans / 12 ans",
			coupeTenableDepth, coupeTenableEasy, coupeTenableWall)
	}
}

// House rules on the rendered plate: no rgba, no opacity, no em-dash, and the
// two objects named on its surface so no reader takes the claim for a measure.
func TestCoupeFigureSurface(t *testing.T) {
	s := FigureSVG("coupe-exigee-tenable")
	if s == "" {
		t.Fatal("coupe-exigee-tenable is not registered in the figures map")
	}
	for _, bad := range []string{"rgba(", "opacity", "\u2014"} { // the banned em-dash
		if strings.Contains(s, bad) {
			t.Errorf("the plate contains %q", bad)
		}
	}
	for _, want := range []string{
		"exigé par le marché", "affirmation de ce chapitre, pas une mesure",
		"Plan à 4,5 %", "Plan à 4,0 %",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("the plate no longer says %q", want)
		}
	}
	if coupeEUR(45000) != "45 000 €" {
		t.Errorf("coupeEUR(45000) = %q", coupeEUR(45000))
	}
}
