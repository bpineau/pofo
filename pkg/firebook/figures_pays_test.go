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

const safemaxHorizon = 30

// broadSample6040 reads the bundled Jorda-Schularick-Taylor panel and returns,
// per country, the domestic 60/40 real annual return of every year where BOTH
// the equity and the bond leg are present, plus the panel's year span.
func broadSample6040(t *testing.T) (map[string]map[int]float64, int, int) {
	t.Helper()
	frozenAgainstData(t)
	r := csv.NewReader(bytes.NewReader(datasets.BroadSample()))
	r.Comment = '#'
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) == 0 || recs[0][0] != "iso" {
		t.Fatalf("unexpected broad-sample header %v", recs[0])
	}
	out := map[string]map[int]float64{}
	first, last := math.MaxInt32, math.MinInt32
	for _, rec := range recs[1:] {
		year, err := strconv.Atoi(rec[1])
		if err != nil {
			t.Fatal(err)
		}
		if year < first {
			first = year
		}
		if year > last {
			last = year
		}
		if rec[2] == "" || rec[3] == "" {
			continue
		}
		eq, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			t.Fatalf("%s %d equity: %v", rec[0], year, err)
		}
		bd, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			t.Fatalf("%s %d bond: %v", rec[0], year, err)
		}
		if out[rec[0]] == nil {
			out[rec[0]] = map[int]float64{}
		}
		out[rec[0]][year] = 0.6*eq + 0.4*bd
	}
	if len(out) == 0 {
		t.Fatal("no usable rows in the broad-sample panel")
	}
	return out, first, last
}

// safemaxOf replays the plate's recipe over one real annual series: the highest
// initial withdrawal rate (percent, held constant in real terms, taken at the
// start of each year) that survives every complete 30-year window, the window
// that sets it, and the number of windows that could be computed at all.
func safemaxOf(series map[int]float64, first, last int) (rate float64, worst, windows int) {
	rate = math.Inf(1)
	for s := first; s+safemaxHorizon-1 <= last; s++ {
		pv, disc, ok := 0.0, 1.0, true
		for t := 0; t < safemaxHorizon; t++ {
			r, has := series[s+t]
			if !has {
				ok = false
				break
			}
			pv += disc
			disc /= 1 + r
		}
		if !ok {
			continue
		}
		windows++
		if w := 100 / pv; w < rate {
			rate, worst = w, s
		}
	}
	return rate, worst, windows
}

// The plate freezes sixteen SAFEMAX values so the book's figures stay pure
// functions. This recomputes each one from pkg/datasets and fails on drift.
func TestSafemaxPaysMatchesTheJSTPanel(t *testing.T) {
	panel, first, last := broadSample6040(t)
	if len(panel) != len(safemaxCountries) {
		t.Fatalf("the panel carries %d countries, the plate draws %d", len(panel), len(safemaxCountries))
	}
	for _, c := range safemaxCountries {
		series, ok := panel[c.ISO]
		if !ok {
			t.Fatalf("%s: no such country in the panel", c.ISO)
		}
		rate, worst, windows := safemaxOf(series, first, last)
		if math.Abs(rate-c.Rate) > 5e-4 {
			t.Errorf("%s: the panel gives %.4f %%, the plate says %.4f %%", c.ISO, rate, c.Rate)
		}
		if worst != c.Worst {
			t.Errorf("%s: the worst window starts in %d, the plate says %d", c.ISO, worst, c.Worst)
		}
		if windows != c.Windows {
			t.Errorf("%s: the panel offers %d windows, the plate says %d", c.ISO, windows, c.Windows)
		}
	}
}

// The world reference line uses the same recipe on the equal-weight basket.
func TestSafemaxWorldBasket(t *testing.T) {
	panel, first, last := broadSample6040(t)
	world := map[int]float64{}
	for y := first; y <= last; y++ {
		sum, n := 0.0, 0
		for _, series := range panel {
			if r, ok := series[y]; ok {
				sum += r
				n++
			}
		}
		if n > 0 {
			world[y] = sum / float64(n)
		}
	}
	rate, worst, windows := safemaxOf(world, first, last)
	if math.Abs(rate-safemaxWorld) > 5e-4 {
		t.Errorf("the basket gives %.4f %%, the plate says %.4f %%", rate, safemaxWorld)
	}
	if worst != safemaxWorldWorst || windows != safemaxWorldWindows {
		t.Errorf("the basket's worst window starts in %d over %d windows, the plate says %d over %d",
			worst, windows, safemaxWorldWorst, safemaxWorldWindows)
	}
	// The plate's whole point: the basket beats most single countries, without
	// reaching the 4 % of the American tradition.
	beaten := 0
	for _, c := range safemaxCountries {
		if c.Rate < rate {
			beaten++
		}
	}
	if beaten < len(safemaxCountries)/2 {
		t.Errorf("the basket only beats %d of the %d countries, the plate presents it as a shelter",
			beaten, len(safemaxCountries))
	}
	if rate >= 4 {
		t.Errorf("the basket carries %.2f %%, the plate draws it below the 4 %% rule", rate)
	}
}

// The two hollow bars are hollow because the panel misses exactly the years
// that would have set their SAFEMAX. Assert both halves of that claim: the
// years are indeed absent, and no surviving window covers them.
func TestSafemaxPartialCountries(t *testing.T) {
	panel, _, _ := broadSample6040(t)
	marked := map[string]bool{}
	for _, c := range safemaxCountries {
		if c.Partial {
			marked[c.ISO] = true
		}
	}
	for iso := range safemaxGaps {
		if !marked[iso] {
			t.Errorf("%s has a declared gap but is not marked partial on the plate", iso)
		}
	}
	for iso := range marked {
		gap, ok := safemaxGaps[iso]
		if !ok {
			t.Fatalf("%s is marked partial with no declared gap", iso)
		}
		for _, y := range gap {
			if _, has := panel[iso][y]; has {
				t.Errorf("%s: the panel does have %d, the plate says the year is missing", iso, y)
			}
		}
		// the gap is interior, so it costs the country a whole run of windows
		if c := countryOf(t, iso).Windows; c >= 120 {
			t.Errorf("%s keeps %d windows, an interior gap should have cost it many", iso, c)
		}
	}
}

func countryOf(t *testing.T, iso string) safemaxCountry {
	t.Helper()
	for _, c := range safemaxCountries {
		if c.ISO == iso {
			return c
		}
	}
	t.Fatalf("%s is not on the plate", iso)
	return safemaxCountry{}
}

// The ranking is the figure: it must be sorted, and the article's own claims
// about it must hold on the frozen numbers.
func TestSafemaxPaysRanking(t *testing.T) {
	for i := 1; i < len(safemaxCountries); i++ {
		if safemaxCountries[i].Rate > safemaxCountries[i-1].Rate {
			t.Errorf("%s (%.2f) sits above %s (%.2f): the plate ranks descending",
				safemaxCountries[i].ISO, safemaxCountries[i].Rate,
				safemaxCountries[i-1].ISO, safemaxCountries[i-1].Rate)
		}
	}
	if last := safemaxCountries[len(safemaxCountries)-1]; last.ISO != "FRA" {
		t.Errorf("the plate's last country is %s; the caption says France closes the ranking", last.ISO)
	}
	// The article: "sous 3 % dans la majorité des pays et sous 1,5 % dans les
	// pires (France comprise)".
	under3, under15 := 0, 0
	for _, c := range safemaxCountries {
		if c.Rate < 3 {
			under3++
		}
		if c.Rate < 1.5 {
			under15++
		}
	}
	if under3*2 <= len(safemaxCountries) {
		t.Errorf("only %d of %d countries are under 3 %%, the article says a majority",
			under3, len(safemaxCountries))
	}
	// [[diversification-internationale]] now says "huit pays sur seize tiennent
	// moins de 2,5 %, France comprise".
	under25 := 0
	for _, c := range safemaxCountries {
		if c.Rate < 2.5 {
			under25++
		}
	}
	if under25 != 8 {
		t.Errorf("%d countries are under 2,5 %%, the cross-reference in the diversification article says eight", under25)
	}
	if under15 < 4 {
		t.Errorf("only %d countries are under 1,5 %%, the article calls that group the worst ones", under15)
	}
	if countryOf(t, "FRA").Rate >= 1.5 {
		t.Error("France is not under 1,5 %, the article puts it in the worst group")
	}
	// The 4 % rule is a United States artefact, and even there this panel does
	// not reach it: that gap is stated in the plate's footnote.
	if r := countryOf(t, "USA").Rate; r >= 4 || math.Round(r*100)/100 != 3.75 {
		t.Errorf("the American bar is %.2f %%, the footnote quotes 3,75 %% against Pfau's ~4 %%", r)
	}
}

// The French footnote quotes two geometric means; recompute them, since they
// are the plate's answer to the Dimson-Marsh-Staunton numbers quoted elsewhere
// in the book.
func TestSafemaxFranceEquityFootnote(t *testing.T) {
	frozenAgainstData(t)
	r := csv.NewReader(bytes.NewReader(datasets.BroadSample()))
	r.Comment = '#'
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	eq := map[int]float64{}
	for _, rec := range recs[1:] {
		if rec[0] != "FRA" || rec[2] == "" {
			continue
		}
		year, _ := strconv.Atoi(rec[1])
		v, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			t.Fatal(err)
		}
		eq[year] = v
	}
	cagr := func(from, to int) float64 {
		mult, n := 1.0, 0
		for y := from; y <= to; y++ {
			v, ok := eq[y]
			if !ok {
				t.Fatalf("the panel has no French equity return for %d", y)
			}
			mult *= 1 + v
			n++
		}
		return (math.Pow(mult, 1/float64(n)) - 1) * 100
	}
	for _, c := range []struct {
		from, to int
		want     float64
	}{
		{1900, 2020, safemaxFranceEquityFull},
		{1950, 2020, safemaxFranceEquityPost50},
	} {
		if got := cagr(c.from, c.to); math.Abs(got-c.want) > 5e-3 {
			t.Errorf("French real equities %d-%d: panel %.3f %%/yr, plate %.3f %%/yr",
				c.from, c.to, got, c.want)
		}
	}
}

// House rules the renderer cannot catch on its own.
func TestSafemaxPaysPlateHygiene(t *testing.T) {
	s := figSafemaxPays()
	if strings.Contains(s, "—") {
		t.Error("the plate contains an em-dash")
	}
	if strings.Contains(s, "rgba(") || strings.Contains(s, "opacity") {
		t.Error("the plate uses rgba or opacity, which crengine paints solid black")
	}
	if strings.Contains(s, "rotate") {
		t.Error("the plate rotates a label, the house rules forbid it")
	}
	for _, want := range []string{"France", "États-Unis", "la règle des 4 %", "panier mondial équipondéré"} {
		if !strings.Contains(s, want) {
			t.Errorf("the plate no longer says %q", want)
		}
	}
}
