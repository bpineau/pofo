package firebook

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

// rattrapReal rebuilds the real S&P 500 the plate freezes: the bundled
// month-end total-return index from December 1972, deflated by the ^CPI-US
// levels marketdata serves offline, compounded from 100. It returns the whole
// record (not just the plate's window), so the durability of the recovery can
// be checked against every month since.
func rattrapReal(t *testing.T) ([]time.Time, []float64) {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "SP500-USD")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("bundled SP500-USD refdata not found")
	}
	base := time.Date(1972, 12, 1, 0, 0, 0, 0, time.UTC)
	var pts []marketdata.Point
	for _, p := range s.Points {
		if p.Close > 0 && !p.Date.Before(base) {
			pts = append(pts, p)
		}
	}
	cs, err := marketdata.NewClient("").Fetch(context.Background(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	real := scenario.Deflate(pts, cs.Points)
	dates := make([]time.Time, len(pts))
	index := make([]float64, len(pts))
	dates[0], index[0] = pts[0].Date, 100
	for i, r := range real {
		dates[i+1], index[i+1] = pts[i+1].Date, index[i]*(1+r)
	}
	return dates, index
}

// Every frozen month of the plate comes from the bundled series, and the base
// month is the one the caption names.
func TestRattrapIndexMatchesTheData(t *testing.T) {
	dates, index := rattrapReal(t)
	if len(index) < len(rattrapIndex) {
		t.Fatalf("the record holds %d months, the plate freezes %d", len(index), len(rattrapIndex))
	}
	if got := dates[0].Format("2006-01"); got != "1972-12" {
		t.Fatalf("the plate is based on December 1972, the record starts in %s", got)
	}
	for i, want := range rattrapIndex {
		if math.Abs(index[i]-want) > 0.02 {
			t.Errorf("%s: data %.2f, plate %.2f", dates[i].Format("2006-01"), index[i], want)
		}
	}
	// Fifteen years, ending on the December the footnote talks about.
	if got := dates[len(rattrapIndex)-1].Format("2006-01"); got != "1987-12" {
		t.Errorf("the window must end in December 1987, it ends in %s", got)
	}
}

// The two annotated months are read off the series, not off the prose: the
// minimum of the window, and the month after which the index never falls back
// below 100 (the plate's stated definition of a durable recovery).
func TestRattrapTroughAndRecovery(t *testing.T) {
	dates, index := rattrapReal(t)
	win := rattrapIndex

	low := 0
	for i, v := range win {
		if v < win[low] {
			low = i
		}
	}
	if low != rattrapTrough {
		t.Errorf("the trough is month %d (%s), the plate marks %d", low, dates[low].Format("2006-01"), rattrapTrough)
	}
	if got := dates[rattrapTrough].Format("2006-01"); got != "1974-09" {
		t.Errorf("the plate says September 1974, the data says %s", got)
	}
	// "47,7, soit −52 % de pouvoir d'achat" and "la moitié" in the title.
	if v := win[rattrapTrough]; math.Abs(v-47.74) > 0.02 || v > 50 {
		t.Errorf("trough at %.2f: the plate says 47,7 and more than half the purchasing power gone", v)
	}

	// The durable recovery: the last month strictly below 100 over the WHOLE
	// record, so the claim "never below since" is checked against today's tail
	// and not only against the plotted window.
	lastBelow := -1
	for i, v := range index {
		if v < 100 {
			lastBelow = i
		}
	}
	if lastBelow+1 != rattrapBack {
		t.Errorf("the index is last below 100 in %s, so the durable recovery is month %d, the plate marks %d",
			dates[lastBelow].Format("2006-01"), lastBelow+1, rattrapBack)
	}
	if got := dates[rattrapBack].Format("2006-01"); got != "1985-01" {
		t.Errorf("the plate says January 1985, the data says %s", got)
	}
	// Here the first crossing and the durable one are the same month, which is
	// what lets the plate mark a single date.
	for i := 1; i < rattrapBack; i++ {
		if index[i] >= 100 {
			t.Errorf("%s already reached %.2f: the plate claims January 1985 is the first crossing too",
				dates[i].Format("2006-01"), index[i])
			break
		}
	}
	// "douze ans et un mois", and the recovery sits inside the plotted window.
	if y, m := rattrapBack/12, rattrapBack%12; y != 12 || m != 1 {
		t.Errorf("the gap is %d years and %d months, the plate says twelve years and one month", y, m)
	}
	if rattrapBack >= len(rattrapIndex) {
		t.Error("the recovery month falls outside the plotted window")
	}
}

// The footnote's two ratios, over the same window as the curve.
func TestRattrapFootnoteRatios(t *testing.T) {
	s, _, err := marketdata.ReadSimdataFS(datasets.Refdata(), "SP500-USD")
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(1972, 12, 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC)
	var lo, hi float64
	for _, p := range s.Points {
		if p.Close <= 0 || p.Date.Before(first) || !p.Date.Before(last) {
			continue
		}
		if lo == 0 {
			lo = p.Close
		}
		hi = p.Close
	}
	if got := hi / lo; math.Abs(got-rattrapNominalx) > 0.01 {
		t.Errorf("the nominal index is multiplied by %.2f, the plate says %.2f", got, rattrapNominalx)
	}

	cs, err := marketdata.NewClient("").Fetch(context.Background(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	level := func(t time.Time) float64 {
		v := cs.Points[0].Close
		for _, p := range cs.Points {
			if p.Date.After(t) {
				break
			}
			v = p.Close
		}
		return v
	}
	dates, _ := rattrapReal(t)
	if got := level(dates[len(rattrapIndex)-1]) / level(dates[0]); math.Abs(got-rattrapCPIx) > 0.01 {
		t.Errorf("US prices are multiplied by %.2f, the plate says %.2f", got, rattrapCPIx)
	}
	// The plate ends its curve at 148, "48 % au-dessus du départ".
	if v := rattrapIndex[len(rattrapIndex)-1]; math.Abs(v-148) > 0.6 {
		t.Errorf("the last month reads %.2f, the plate labels it 148", v)
	}
}

// rattrapRuns must cut the curve exactly where it meets the rule, otherwise the
// two washes bleed across the frontier the plate is about.
func TestRattrapRunsSplitAtTheRule(t *testing.T) {
	below, above := rattrapRuns([][2]float64{{0, 10}, {10, 30}, {20, 10}}, 20)
	if len(below) != 1 || len(above) != 2 {
		t.Fatalf("got %d runs below and %d above, want 1 and 2", len(below), len(above))
	}
	for _, run := range append(append([][][2]float64{}, below...), above...) {
		if run[0][1] != 20 || run[len(run)-1][1] != 20 {
			t.Errorf("run %v is not closed onto the rule", run)
		}
	}
	if x := below[0][0][0]; math.Abs(x-5) > 1e-9 {
		t.Errorf("the first crossing is at x=%.3f, want 5", x)
	}
}

// The plate must render, register under its id and stay inside the house rules.
func TestRattrapPlateRenders(t *testing.T) {
	svg := FigureSVG("actions-rattrapent")
	if svg == "" {
		t.Fatal("actions-rattrapent is not registered")
	}
	if strings.Contains(svg, "rgba(") || strings.Contains(svg, "opacity") {
		t.Error("no rgba and no opacity in a plate: crengine paints them black")
	}
	if strings.Contains(svg, "rotate") {
		t.Error("every label stays horizontal")
	}
	if strings.Contains(svg, "—") {
		t.Error("no em-dash")
	}
	for _, want := range []string{
		"creux de septembre 1974", "−52 %", "janvier 1985", "12 ans et 1 mois",
		"pouvoir d'achat de départ", "CPI américain", "1973", "1987",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate is missing %q", want)
		}
	}
}
