package firebook

import (
	"context"
	"encoding/csv"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// The guard test of the tente-transfert plate. It rebuilds the two legs the
// plate needs (equity and bonds SEPARATELY, since the allocation moves year by
// year, which the blended 60/40 of pkg/replay cannot give), re-solves the
// maximum sustainable withdrawal rate of every vintage under both allocations,
// and fails when any frozen number drifts.

// tenteMonth maps a date to a dense month index, the key that aligns the three
// bundled series (the equity leg is dated month-end, the bond leg
// first-of-month, the CPI first-of-month).
func tenteMonth(t time.Time) int { return t.Year()*12 + int(t.Month()) - 1 }

// tenteLeg reads one bundled reference series into month-keyed levels.
func tenteLeg(t *testing.T, id string) map[int]float64 {
	t.Helper()
	frozenAgainstData(t)
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil {
		t.Fatalf("reading %s: %v", id, err)
	}
	if !ok {
		t.Fatalf("bundled reference series %s not found", id)
	}
	out := make(map[int]float64, len(s.Points))
	for _, p := range s.Points {
		if p.Close > 0 {
			out[tenteMonth(p.Date)] = p.Close // the last observation of the month wins
		}
	}
	return out
}

// tenteCPI reads the bundled US CPI-U levels; marketdata serves ^CPI-US
// offline-first from its embedded snapshot, so this never reaches the network.
func tenteCPI(t *testing.T) map[int]float64 {
	t.Helper()
	frozenAgainstData(t)
	s, err := marketdata.NewClient("").Fetch(context.Background(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatalf("reading ^CPI-US: %v", err)
	}
	out := make(map[int]float64, len(s.Points))
	for _, p := range s.Points {
		if p.Close > 0 {
			out[tenteMonth(p.Date)] = p.Close
		}
	}
	return out
}

// tenteYearly folds one nominal leg into December-to-December real growth
// factors, keyed by calendar year (1.07 = +7 % real over the year).
func tenteYearly(level, cpi map[int]float64) map[int]float64 {
	out := make(map[int]float64, len(level)/12)
	for k, v := range level {
		if time.Month(k%12+1) != time.December {
			continue
		}
		prev, ok := level[k-12]
		c0, ok0 := cpi[k-12]
		c1, ok1 := cpi[k]
		if !ok || !ok0 || !ok1 {
			continue
		}
		out[k/12] = (v / prev) * (c0 / c1)
	}
	return out
}

// tenteSWR bisects the maximum constant real withdrawal rate a vintage can
// sustain: the withdrawal is taken at the start of the year, what is left is
// rebalanced to that year's target equity share, and the year's real returns
// apply. Survival is monotone in the rate, which is what makes the bisection
// legitimate.
func tenteSWR(eq, bd []float64, alloc func(int) float64) float64 {
	survives := func(w float64) bool {
		p := 1.0
		for t := range eq {
			p -= w
			if p < 0 {
				return false
			}
			a := alloc(t)
			p = p*a*eq[t] + p*(1-a)*bd[t]
		}
		return p >= 0
	}
	lo, hi := 0.0, 0.30
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if survives(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

// tenteSolved is one re-solved vintage: the gain in points and the static rate
// that measures how hard the vintage was.
type tenteSolved struct {
	year   int
	gain   float64
	static float64
}

// tenteSolve replays the plate's whole recipe from the bundled datasets.
func tenteSolve(t *testing.T) []tenteSolved {
	t.Helper()
	cpi := tenteCPI(t)
	eq := tenteYearly(tenteLeg(t, "SP500-USD"), cpi)
	bd := tenteYearly(tenteLeg(t, "TREASURY-INT-USD"), cpi)

	years := make([]int, 0, len(eq))
	for y := range eq {
		if _, ok := bd[y]; ok {
			years = append(years, y)
		}
	}
	sort.Ints(years)

	static := func(int) float64 { return tenteStatic }
	var out []tenteSolved
	for _, y := range years {
		re := make([]float64, 0, tenteYears)
		rb := make([]float64, 0, tenteYears)
		for k := 0; k < tenteYears; k++ {
			a, oka := eq[y+k]
			b, okb := bd[y+k]
			if !oka || !okb {
				break
			}
			re, rb = append(re, a), append(rb, b)
		}
		if len(re) < tenteYears {
			continue
		}
		s := tenteSWR(re, rb, static)
		out = append(out, tenteSolved{y, (tenteSWR(re, rb, tenteAlloc) - s) * 100, s * 100})
	}
	return out
}

// The static twin must carry exactly the tent's average equity exposure: that
// equality is the whole reason the difference of the two rates measures a
// transfer and not a change of risk budget.
func TestTenteStaticIsTheTentAverage(t *testing.T) {
	sum := 0.0
	for k := 0; k < tenteYears; k++ {
		sum += tenteAlloc(k)
	}
	if got := sum / tenteYears; math.Abs(got-tenteStatic) > 1e-9 {
		t.Errorf("static twin at %.4f, the tent averages %.4f", tenteStatic, got)
	}
	if tenteAlloc(0) != tenteStart || tenteAlloc(tenteClimb) != tenteEnd || tenteAlloc(tenteYears-1) != tenteEnd {
		t.Errorf("the tent does not run from %.2f to %.2f", tenteStart, tenteEnd)
	}
}

// Every point of the curve is re-solved from the bundled legs.
func TestTenteTransfertMatchesTheSolver(t *testing.T) {
	solved := tenteSolve(t)
	if len(solved) != len(tenteCohorts) {
		t.Fatalf("the datasets give %d vintages, the plate draws %d", len(solved), len(tenteCohorts))
	}
	if solved[0].year != 1954 || solved[len(solved)-1].year != 1996 {
		t.Fatalf("vintages run %d-%d, the plate footnote says 1954-1996", solved[0].year, solved[len(solved)-1].year)
	}
	sort.SliceStable(solved, func(i, j int) bool { return solved[i].gain > solved[j].gain })

	// The ten hardest vintages, the plate's second colour.
	byHardness := append([]tenteSolved(nil), solved...)
	sort.SliceStable(byHardness, func(i, j int) bool { return byHardness[i].static < byHardness[j].static })
	hard := make(map[int]bool, 10)
	for _, s := range byHardness[:10] {
		hard[s.year] = true
	}
	if byHardness[9].static >= byHardness[10].static {
		t.Error("the hardness ranking is not strict at the cut")
	}

	for i, c := range tenteCohorts {
		got := solved[i]
		if got.year != c.year {
			t.Errorf("rank %d: solver says %d, plate says %d", i+1, got.year, c.year)
			continue
		}
		if math.Abs(got.gain-c.gain) > 0.005 {
			t.Errorf("%d: solver gains %+.3f pt, plate draws %+.3f pt", c.year, got.gain, c.gain)
		}
		if hard[c.year] != c.hard {
			t.Errorf("%d: solver hardness %v, plate marks %v", c.year, hard[c.year], c.hard)
		}
	}
}

// The four readings printed on the plate are recomputed from the same solve, so
// the sentences cannot drift away from the curve above them.
func TestTenteReadingsMatchTheCurve(t *testing.T) {
	solved := tenteSolve(t)
	sort.SliceStable(solved, func(i, j int) bool { return solved[i].gain > solved[j].gain })

	winners := 0
	gains := make([]float64, len(solved))
	for i, s := range solved {
		if s.gain > 0 {
			winners++
		}
		gains[i] = s.gain
	}
	if winners != tenteWinners {
		t.Errorf("%d vintages gain, the plate says %d", winners, tenteWinners)
	}
	sort.Float64s(gains)
	if median := gains[len(gains)/2]; math.Abs(median-tenteMedian) > 0.005 {
		t.Errorf("median gain %+.3f pt, the plate says %+.3f pt", median, tenteMedian)
	}

	byHardness := append([]tenteSolved(nil), solved...)
	sort.SliceStable(byHardness, func(i, j int) bool { return byHardness[i].static < byHardness[j].static })
	mean := func(v []tenteSolved) float64 {
		sum := 0.0
		for _, s := range v {
			sum += s.gain
		}
		return sum / float64(len(v))
	}
	if got := mean(byHardness[:10]); math.Abs(got-tenteHardMean) > 0.005 {
		t.Errorf("the ten hardest vintages gain %+.3f pt, the plate says %+.3f pt", got, tenteHardMean)
	}
	if got := mean(byHardness[len(byHardness)-10:]); math.Abs(got-tenteEasyMean) > 0.005 {
		t.Errorf("the ten easiest vintages gain %+.3f pt, the plate says %+.3f pt", got, tenteEasyMean)
	}

	// The article's caption reads the amber markers off the curve: the ten
	// hardest vintages all sit in its favourable half.
	rank := make(map[int]int, len(solved))
	for i, s := range solved {
		rank[s.year] = i + 1
	}
	for _, s := range byHardness[:10] {
		if r := rank[s.year]; r > len(solved)/2 {
			t.Errorf("the %d vintage is one of the ten hardest but ranks %d of %d, in the unfavourable half",
				s.year, r, len(solved))
		}
	}
}

// The plate drops the article's CAPE conditioning on purpose. This test is the
// record of why: over this window no vintage ever left above a CAPE of 25, so
// the highlight the backlog asked for would have marked an empty set. It fails
// the day a longer sample makes the highlight possible, which is exactly when
// the plate should be revisited.
func TestTenteSampleCannotTestTheCAPEThesis(t *testing.T) {
	frozenAgainstData(t)
	cape := map[int]float64{}
	r := csv.NewReader(strings.NewReader(string(datasets.CAPE())))
	r.Comment = '#'
	r.FieldsPerRecord = -1
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatalf("reading the bundled CAPE: %v", err)
	}
	for _, rec := range recs {
		if len(rec) < 2 {
			continue
		}
		d, err := time.Parse("2006-01-02", rec[0])
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(rec[1], 64)
		if err != nil || v <= 0 || d.Month() != time.January {
			continue
		}
		cape[d.Year()] = v
	}
	high, top := 0, 0.0
	for _, c := range tenteCohorts {
		v, ok := cape[c.year]
		if !ok {
			t.Fatalf("no January CAPE for the %d vintage", c.year)
		}
		if v > top {
			top = v
		}
		if v > 25 {
			high++
		}
	}
	if high != 0 {
		t.Errorf("%d vintages left above a CAPE of 25 (top %.1f): the plate can now show the valuation thesis", high, top)
	}
	if top < 20 || top > 25 {
		t.Errorf("the dearest departure of the sample sits at a CAPE of %.1f, outside the documented 24,8", top)
	}
}

// House rules of the plate system, checked on the rendered surface.
func TestTenteTransfertRenders(t *testing.T) {
	if figures["tente-transfert"] == nil {
		t.Fatal("tente-transfert is not registered in the figures map")
	}
	s := figTenteTransfert()
	for _, want := range []string{"viewBox", "S&amp;P 500", "1954", "1996", "43 départs"} {
		if !strings.Contains(s, want) {
			t.Errorf("the plate never mentions %q", want)
		}
	}
	// The banned list holds the em-dash as an escape, so this file carries none.
	for _, banned := range []string{"rgba(", "opacity", "\u2014", "rotate("} {
		if strings.Contains(s, banned) {
			t.Errorf("the plate uses %q, which the plate system forbids", banned)
		}
	}
}
