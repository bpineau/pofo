package firebook

import (
	"bytes"
	"encoding/csv"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
)

// millesimesPanelUSA reads the bundled Jorda-Schularick-Taylor panel and
// returns the American 50/50 real annual return of every year where both legs
// are present.
func millesimesPanelUSA(t *testing.T) map[int]float64 {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(datasets.BroadSample()))
	r.Comment = '#'
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	out := map[int]float64{}
	for _, rec := range recs[1:] {
		if rec[0] != "USA" || rec[2] == "" || rec[3] == "" {
			continue
		}
		year, err := strconv.Atoi(rec[1])
		if err != nil {
			t.Fatal(err)
		}
		eq, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			t.Fatal(err)
		}
		bd, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			t.Fatal(err)
		}
		out[year] = 0.5*eq + 0.5*bd
	}
	if len(out) == 0 {
		t.Fatal("no American rows in the broad-sample panel")
	}
	return out
}

// millesimesSolve is the plate's solver: the highest initial withdrawal rate
// (percent of the starting capital, held constant in real terms, taken at the
// start of each year) that the thirty years from start would have carried. It
// is the exact annuity rate, so no search is needed.
func millesimesSolve(t *testing.T, usa map[int]float64, start int) float64 {
	t.Helper()
	pv, disc := 0.0, 1.0
	for i := 0; i < 30; i++ {
		r, ok := usa[start+i]
		if !ok {
			t.Fatalf("the panel has no American return for %d", start+i)
		}
		pv += disc
		disc /= 1 + r
	}
	return 100 / pv
}

// millesimesReplay withdraws rate percent of the starting capital at the start
// of each of the thirty years, holding the amount constant in real terms, and
// returns the balance left after the last withdrawal has compounded. Negative
// means the money ran out.
func millesimesReplay(usa map[int]float64, start int, rate float64) float64 {
	bal, draw := 100.0, rate
	for i := 0; i < 30; i++ {
		bal -= draw
		bal *= 1 + usa[start+i]
	}
	return bal
}

// Every one of the plate's sixty-six numbers is recomputed from the bundled
// panel. This is what fails the day the panel is regenerated.
func TestMillesimesSoutenablesMatchesTheJSTPanel(t *testing.T) {
	usa := millesimesPanelUSA(t)
	for i, want := range millesimesRates {
		start := millesimesFirst + i
		if got := millesimesSolve(t, usa, start); math.Abs(got-want) > 1e-3 {
			t.Errorf("%d: the panel gives %.4f %%, the plate says %.3f %%", start, got, want)
		}
	}
	// the plate stops at the last vintage whose thirty years fit in the panel
	last := millesimesFirst + len(millesimesRates) - 1
	if _, ok := usa[last+30]; ok {
		t.Errorf("the panel reaches %d, so the plate could draw the %d vintage too", last+30, last+1)
	}
	if _, ok := usa[last+29]; !ok {
		t.Errorf("the panel stops before %d, so the %d vintage is incomplete", last+29, last)
	}
}

// The closed-form rate must be the rate a plain year-by-year replay confirms:
// at it the capital lands on zero, a tenth of a point above it the money runs
// out. Checked on every vintage, not just the famous one.
func TestMillesimesSolverAgreesWithAReplay(t *testing.T) {
	usa := millesimesPanelUSA(t)
	for i, rate := range millesimesRates {
		start := millesimesFirst + i
		// the frozen rate is rounded to a thousandth of a point, and thirty
		// years of compounding magnify that by up to ~130x, hence the tolerance
		if end := millesimesReplay(usa, start, rate); math.Abs(end) > 0.15 {
			t.Errorf("%d: at %.3f %% the replay ends on %.3f, not on zero", start, rate, end)
		}
		if end := millesimesReplay(usa, start, rate+0.1); end >= 0 {
			t.Errorf("%d: %.3f %% is not the maximum, %.3f %% survives too", start, rate, rate+0.1)
		}
		if end := millesimesReplay(usa, start, rate-0.1); end <= 0 {
			t.Errorf("%d: %.3f %% already fails, the maximum is lower", start, rate-0.1)
		}
	}
}

// The three numbers the plate and its caption state, and the claim they carry.
func TestMillesimesSoutenablesStats(t *testing.T) {
	if n := len(millesimesRates); n != 66 {
		t.Fatalf("the plate draws %d vintages, its subtitle says 66", n)
	}
	worst, worstYear := millesimesWorst()
	if worstYear != 1966 {
		t.Errorf("the floor is set by %d, the plate and the article name 1966", worstYear)
	}
	if math.Abs(worst-3.672) > 1e-3 {
		t.Errorf("the floor is %.3f %%, the plate says 3,67 %%", worst)
	}
	// Bengen publishes 4,15 % on his own reconstruction; ours must stay in the
	// same neighbourhood, or the plate's footnote is dishonest.
	if math.Abs(worst-4.15) > 0.6 {
		t.Errorf("the floor is %.2f %% against Bengen's 4,15 %%: too far to call it the same result", worst)
	}
	med := millesimesMedian()
	if math.Abs(med-5.891) > 1e-3 {
		t.Errorf("the median is %.3f %%, the plate says 5,9 %%", med)
	}
	if med <= worst+2 {
		t.Errorf("the median (%.2f %%) is only %.2f points above the floor; the plate is built on that gap",
			med, med-worst)
	}
	// the plate's closing sentence, rounded the way it prints it
	if got := math.Round((med/worst - 1) * 100); got != 60 {
		t.Errorf("the median vintage carried %.0f %% more than 1966, the plate says 60 %%", got)
	}
	best, bestYear := millesimesBest()
	if bestYear != 1982 || math.Abs(best-10.191) > 1e-3 {
		t.Errorf("the best vintage is %d at %.3f %%, the plate labels 1982 at 10,2 %%", bestYear, best)
	}
	// the band under the rule: six consecutive mid-sixties departures
	if n := millesimesUnder(4); n != 6 {
		t.Errorf("%d vintages fall under 4 %%, the plate says six", n)
	}
	for y := 1926; y <= 1991; y++ {
		under := millesimesRates[y-millesimesFirst] < 4
		if want := y >= 1964 && y <= 1969; under != want {
			t.Errorf("%d under 4 %% = %v, the plate says the run is exactly 1964-1969", y, under)
		}
	}
}

// House rules the renderer cannot catch on its own.
func TestMillesimesSoutenablesPlateHygiene(t *testing.T) {
	s := figMillesimesSoutenables()
	if strings.Contains(s, "—") {
		t.Error("the plate contains an em-dash")
	}
	if strings.Contains(s, "rgba(") || strings.Contains(s, "opacity") {
		t.Error("the plate uses rgba or opacity, which crengine paints solid black")
	}
	if strings.Contains(s, "rotate") {
		t.Error("the plate rotates a label, the house rules forbid it")
	}
	for _, want := range []string{"la règle", "4,0 %", "médiane", "5,9 %", "1966", "3,67 %", "Bengen"} {
		if !strings.Contains(s, want) {
			t.Errorf("the plate no longer says %q", want)
		}
	}
}

// Figure and prose must agree: the article's own claim about the middle of the
// record is what the plate draws.
func TestMillesimesSoutenablesAgreesWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/la-regle-des-4-pourcents.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	if !strings.Contains(article, "::: figure millesimes-soutenables") {
		t.Error("the article must carry the plate")
	}
	// the article says the median vintage supports "près de 6 %"; the plate
	// measures 5,89 %, so that wording must stay true
	if med := millesimesMedian(); med < 5.5 || med >= 6 {
		t.Errorf("the median is %.2f %%, the article says the median vintage supports near 6 %%", med)
	}
	if !strings.Contains(article, "près de 6 %") {
		t.Error("the article no longer states what the plate's median line shows")
	}
	// Bengen's own figure stays attributed to Bengen, never to this plate
	if !strings.Contains(article, "4,15 %") {
		t.Error("the article no longer quotes Bengen's published 4,15 %")
	}
}
