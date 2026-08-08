package firebook

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// triFirstAnchor is the month key (year*12 + month - 1) of the first month end
// of the window, December 2000: the first monthly return of the plate is
// January 2001, where the trend leg's reconstruction is calibrated.
const triFirstAnchor = 2000*12 + 11

// triReturns recomputes, from pkg/datasets alone, the monthly returns of every
// plotted series over the plate's single common window: the days all seven
// series quote, reduced to the last such day of each month, from December 2000
// on. It returns the month keys of the anchors and one return slice per row.
func triReturns(t *testing.T) (anchors []int, rets [][]float64) {
	t.Helper()
	levels := make([]map[time.Time]float64, len(triRows))
	for i, s := range triRows {
		fsys := datasets.Simdata()
		if s.ref == "refdata" {
			fsys = datasets.Refdata()
		}
		series, ok, err := marketdata.ReadSimdataFS(fsys, s.id)
		if err != nil || !ok {
			t.Fatalf("bundled %s: ok=%v err=%v", s.id, ok, err)
		}
		m := make(map[time.Time]float64, len(series.Points))
		for _, p := range series.Points {
			if p.Close > 0 {
				m[p.Date] = p.Close
			}
		}
		levels[i] = m
	}

	// The days every series quotes, in order.
	var days []time.Time
	for d := range levels[0] {
		all := true
		for _, m := range levels[1:] {
			if _, ok := m[d]; !ok {
				all = false
				break
			}
		}
		if all {
			days = append(days, d)
		}
	}
	if len(days) < 2 {
		t.Fatalf("only %d days quoted by all seven series", len(days))
	}
	for i := 1; i < len(days); i++ { // insertion order of a map range is random
		for j := i; j > 0 && days[j].Before(days[j-1]); j-- {
			days[j], days[j-1] = days[j-1], days[j]
		}
	}

	// The last quoted day of each month, from December 2000 on.
	var ends []time.Time
	for i, d := range days {
		if i+1 < len(days) && days[i+1].Month() == d.Month() {
			continue
		}
		if k := d.Year()*12 + int(d.Month()) - 1; k < triFirstAnchor {
			continue
		}
		ends = append(ends, d)
	}
	for _, d := range ends {
		anchors = append(anchors, d.Year()*12+int(d.Month())-1)
	}
	rets = make([][]float64, len(triRows))
	for i, m := range levels {
		r := make([]float64, 0, len(ends)-1)
		for k := 1; k < len(ends); k++ {
			r = append(r, m[ends[k]]/m[ends[k-1]]-1)
		}
		rets[i] = r
	}
	return anchors, rets
}

// The window the plate claims: 306 month ends, December 2000 to May 2026, that
// is 305 monthly returns with no month missing.
func TestTriangleWindowIsTheOneTheCaptionStates(t *testing.T) {
	anchors, rets := triReturns(t)
	if got, want := anchors[0], triFirstAnchor; got != want {
		t.Fatalf("the window opens at month key %d, the plate assumes %d (December 2000)", got, want)
	}
	if got, want := anchors[len(anchors)-1], 2026*12+4; got != want {
		t.Fatalf("the window closes at month key %d, the caption says May 2026 (%d)", got, want)
	}
	for i := 1; i < len(anchors); i++ {
		if anchors[i] != anchors[i-1]+1 {
			t.Fatalf("the window skips a month at index %d (key %d after %d)", i, anchors[i], anchors[i-1])
		}
	}
	if got, want := len(rets[0]), 305; got != want {
		t.Errorf("the window holds %d monthly returns, the caption says %d", got, want)
	}
	if !strings.Contains(figTriangleCorrelations(), "305 mois") {
		t.Error("the plate no longer states the length of its window")
	}
}

// Every one of the 21 frozen cells must be what pkg/datasets says.
func TestTriangleMatchesTheData(t *testing.T) {
	_, rets := triReturns(t)
	for i := 1; i < len(triRows); i++ {
		if got, want := len(triCorr[i]), i; got != want {
			t.Fatalf("row %d freezes %d cells, the lower triangle has %d", i, got, want)
		}
		for j := 0; j < i; j++ {
			got := pearson(rets[i], rets[j])
			if math.Abs(got-triCorr[i][j]) > 0.005 {
				t.Errorf("%s against %s: the data reads %+.4f, the plate freezes %+.2f",
					triRows[i].id, triRows[j].id, got, triCorr[i][j])
			}
		}
	}
	if len(triCorr[0]) != 0 {
		t.Error("the first row of a lower triangle holds no cell")
	}
}

// The two ranges the plate prints, and the two claims the article makes about
// them. The equity block runs from 0,85 to 1,00 (0,9955 for the S&P 500
// against the whole US market, which the plate rounds to 1,00), so the
// article's own sentence must quote that upper bound and not 0,95. The bricks
// must stay inside the band the article claims, −0,2 to +0,3.
func TestTriangleRangesAreTheOnesThePlatePrints(t *testing.T) {
	eqLo, eqHi := triRange(triEquityPairs())
	if got := len(triEquityPairs()); got != 6 {
		t.Fatalf("the equity block holds %d pairs, the plate says six", got)
	}
	if eqLo != 0.85 || eqHi != 1.00 {
		t.Errorf("the equity block runs from %+.2f to %+.2f, the plate says 0,85 to 1,00", eqLo, eqHi)
	}
	brLo, brHi := triRange(triBrickPairs())
	if got := len(triBrickPairs()); got != 6 {
		t.Fatalf("the four bricks hold %d pairs, the plate says six", got)
	}
	if brLo != -0.12 || brHi != 0.23 {
		t.Errorf("the bricks run from %+.2f to %+.2f, the plate says −0,12 to +0,23", brLo, brHi)
	}
	if brLo < -0.3 || brHi > 0.3 {
		t.Errorf("the bricks leave the −0,3 / +0,3 band the article claims (%+.2f to %+.2f)", brLo, brHi)
	}
	// Not one equity pair may be as loose as the loosest brick pair: that gap
	// is the whole plate.
	if eqLo <= brHi {
		t.Errorf("the blocks overlap: equity floor %+.2f, brick ceiling %+.2f", eqLo, brHi)
	}
	svg := figTriangleCorrelations()
	for _, want := range []string{"+0,85 à +1,00", "−0,12 à +0,23"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate no longer prints %q", want)
		}
	}
}

// The two footnote numbers: the small-value fund the plate does not draw, and
// the ten-year slices that show these correlations are period averages.
//
// The small-value leg stays GROSS here, unlike the plates that quote a return
// (figScvEcart10Ans, figTousTempsEchange): a fee is a constant multiplicative
// drag, it moves no correlation, and the footnote reads a correlation.
func TestTriangleFootnoteNumbers(t *testing.T) {
	anchors, rets := triReturns(t)
	scv, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "USSCV-USD")
	if err != nil || !ok {
		t.Fatalf("bundled USSCV-USD: ok=%v err=%v", ok, err)
	}
	byMonth := map[int]float64{}
	for _, p := range scv.Points {
		if p.Close > 0 {
			byMonth[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close // last quote of the month wins
		}
	}
	small := make([]float64, 0, len(rets[0]))
	for i := 1; i < len(anchors); i++ {
		prev, cur := byMonth[anchors[i-1]], byMonth[anchors[i]]
		if prev == 0 || cur == 0 {
			t.Fatalf("USSCV-USD does not quote month key %d or %d", anchors[i-1], anchors[i])
		}
		small = append(small, cur/prev-1)
	}
	if got := pearson(rets[0], small); math.Abs(got-0.79) > 0.01 {
		t.Errorf("small value against the world fund reads %+.4f, the plate prints 0,79", got)
	}

	// World equities against long Treasuries, decade by decade.
	lo, hi := 1.0, -1.0
	for _, d := range [][2]int{{2001, 2010}, {2011, 2020}, {2021, 2026}} {
		var a, b []float64
		for i := 1; i < len(anchors); i++ {
			if y := anchors[i] / 12; y >= d[0] && y <= d[1] {
				a = append(a, rets[0][i-1])
				b = append(b, rets[4][i-1])
			}
		}
		v := pearson(a, b)
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	if math.Abs(lo-(-0.42)) > 0.005 || math.Abs(hi-0.58) > 0.005 {
		t.Errorf("the ten-year slices run from %+.4f to %+.4f, the plate prints −0,42 to +0,58", lo, hi)
	}
}

// House rules on the rendered plate: no rgba, no opacity, no em-dash, no
// rotated label, and nothing running past the frozen viewBox.
func TestTrianglePlateRespectsTheHouseRules(t *testing.T) {
	svg := figTriangleCorrelations()
	if !strings.Contains(svg, `viewBox="0 0 640 452"`) {
		t.Error("the plate's viewBox moved: recheck the rendered PNG before freezing it")
	}
	for _, banned := range []string{"rgba(", "opacity", "—", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, banned in the book's figures", banned)
		}
	}
	for _, want := range []string{
		"DIVERSIFICATION", "fonds monde", "obligations longues", "trend",
		"valeurs liquidatives réelles", "Ken French", "corrélation des rendements mensuels",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate no longer carries %q", want)
		}
	}
	// Every drawn cell carries its number, and the diagonal carries none.
	for i := 1; i < len(triRows); i++ {
		for j := 0; j < i; j++ {
			if !strings.Contains(svg, ">"+triNum(triCorr[i][j])+"<") {
				t.Errorf("cell (%d,%d) is missing from the plate", i, j)
			}
		}
	}
	if strings.Contains(svg, ">1,00<") != true {
		t.Error("the plate should print the 1,00 cell it is built around")
	}
	if strings.Count(svg, ">1,00<") != 1 {
		t.Error("only one cell may read 1,00: the plate draws no diagonal")
	}
}
