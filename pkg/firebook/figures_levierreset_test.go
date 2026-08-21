package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// leverResetRecord is one window of the reset-frequency plate as the bundled
// data actually says it, recomputed from scratch.
type leverResetRecord struct {
	index, vol, bill float64
	mult             [3]float64
	share            [3]float64
	trough, gear     float64
}

// leverResetReplay rebuilds a whole window of the plate from pkg/datasets: the
// bundled S&P 500 total return (SP500, daily) and the 3-month bill (TBILL-3M,
// monthly, an annualized percent LEVEL, not a price).
//
// Every arm carries the same leverResetL of exposure and borrows (L-1) of it at
// the bill rate, accrued on a 252-day year. Only the reset differs: the daily
// arm is put back on target every close, the monthly one on the first quote of
// each month, the yearly one on the first quote of each January. Between two
// resets the position is left alone, so its assets and its debt drift apart and
// the effective leverage moves, which is what the term arm is there to show.
//
// trough and gear are the term arm's worst equity multiple and its highest
// effective leverage over the window, the pair the plate quotes on the 2008
// point.
func leverResetReplay(t *testing.T, y0, y1 int) leverResetRecord {
	t.Helper()
	frozenAgainstData(t)
	sp, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), "SP500")
	if err != nil || !ok {
		t.Fatalf("bundled SP500 simdata: ok=%v err=%v", ok, err)
	}
	tb, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "TBILL-3M")
	if err != nil || !ok {
		t.Fatalf("bundled TBILL-3M refdata: ok=%v err=%v", ok, err)
	}
	bill := map[int]float64{}
	for _, p := range tb.Points {
		bill[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close / 100
	}

	var days []marketdata.Point
	var base float64
	for _, p := range sp.Points {
		switch {
		case p.Date.Year() < y0:
			base = p.Close
		case p.Date.Year() <= y1:
			days = append(days, p)
		}
	}
	if base <= 0 || len(days) < 200*(y1-y0+1) {
		t.Fatalf("window %d-%d: base %.2f, %d quotes", y0, y1, base, len(days))
	}
	rateAt := func(i int) float64 {
		k := days[i].Date.Year()*12 + int(days[i].Date.Month()) - 1
		c, ok := bill[k]
		if !ok {
			t.Fatalf("no bill quote for month key %d", k)
		}
		return c
	}

	years := float64(y1 - y0 + 1)
	rec := leverResetRecord{index: days[len(days)-1].Close / base}

	// The index's own annualized volatility and the average financing rate,
	// both in percent: the two numbers the legend and the footnotes quote.
	var s, s2, rs float64
	prev := base
	for i, p := range days {
		r := p.Close/prev - 1
		prev = p.Close
		s += r
		s2 += r * r
		rs += rateAt(i)
	}
	n := float64(len(days))
	mean := s / n
	rec.vol = math.Sqrt((s2/n-mean*mean)*252) * 100
	rec.bill = rs / n * 100

	// The daily arm: exposure put back on target every close.
	e, prev := 1.0, base
	for i, p := range days {
		r := p.Close/prev - 1
		prev = p.Close
		e *= 1 + leverResetL*r - (leverResetL-1)*rateAt(i)/252
	}
	rec.mult[0] = e

	// The monthly and yearly arms: the position drifts between resets.
	periodic := func(months int) (end, trough, gear float64) {
		eq, prev := 1.0, base
		assets, debt := leverResetL*eq, (leverResetL-1)*eq
		trough, gear = math.Inf(1), 0
		start := days[0].Date.Year()*12 + int(days[0].Date.Month()) - 1
		for i, p := range days {
			r := p.Close/prev - 1
			prev = p.Close
			assets *= 1 + r
			debt *= 1 + rateAt(i)/252
			eq = assets - debt
			if eq < trough {
				trough = eq
			}
			if eq > 0 && assets/eq > gear {
				gear = assets / eq
			}
			if eq <= 0 {
				return 0, trough, math.Inf(1) // wiped out, which never happens here
			}
			k := p.Date.Year()*12 + int(p.Date.Month()) - 1
			reset := i == len(days)-1
			if !reset {
				nk := days[i+1].Date.Year()*12 + int(days[i+1].Date.Month()) - 1
				reset = nk != k && (nk-start)%months == 0
			}
			if reset {
				assets, debt = leverResetL*eq, (leverResetL-1)*eq
			}
		}
		return assets - debt, trough, gear
	}
	rec.mult[1], _, _ = periodic(1)
	rec.mult[2], rec.trough, rec.gear = periodic(12)

	// The frictionless benchmark every point is read against: twice the index's
	// own compounded growth, financing deducted.
	bench := math.Pow(rec.index, leverResetL) *
		math.Exp(-(leverResetL-1)*rec.bill/100*years)
	for i := range rec.mult {
		rec.share[i] = rec.mult[i] / bench * 100
	}
	return rec
}

// Every literal the plate freezes must be what the bundled data says.
func TestLevierResetMatchesTheRecord(t *testing.T) {
	for i, w := range leverResetWindows {
		y0, y1 := 2000+i*10, 2009+i*10
		if got := w.years; got != []string{"2000-2009", "2010-2019"}[i] {
			t.Fatalf("window %d is labelled %q, the replay covers %d-%d", i, got, y0, y1)
		}
		rec := leverResetReplay(t, y0, y1)
		near := func(name string, got, want, tol float64) {
			t.Helper()
			if math.Abs(got-want) > tol {
				t.Errorf("%s, %s: the record reads %.4f, the plate freezes %.4f", w.years, name, got, want)
			}
		}
		near("index multiple", rec.index, w.index, 0.005)
		near("volatility", rec.vol, w.vol, 0.02)
		near("bill rate", rec.bill, w.bill, 0.02)
		for a, arm := range leverResetArms {
			near(arm+" multiple", rec.mult[a], w.mult[a], math.Max(0.003, w.mult[a]*0.005))
			near(arm+" share", rec.share[a], w.share[a], 0.15)
		}
	}
}

// The plate's whole vertical scale rests on one identity: against twice the
// index's COMPOUNDED growth, the daily arm falls behind by exactly the realized
// variance, which is the beta slippage of rendements-arithmetiques-geometriques.
// If that stops holding, the plate is drawing something else.
func TestLevierResetDailyGapIsTheSlippage(t *testing.T) {
	for _, w := range leverResetWindows {
		want := 100 * math.Exp(-w.slippage()/100*10)
		if math.Abs(want-w.share[0]) > 0.6 {
			t.Errorf("%s: a slippage of %.2f points a year implies %.1f %% of the benchmark, the plate poses %.1f %%",
				w.years, w.slippage(), want, w.share[0])
		}
		if w.slippage() <= 0 {
			t.Errorf("%s: slippage %.2f, it can only cost", w.years, w.slippage())
		}
	}
	// The choppy decade must show the daily reset losing materially: that is
	// the plate's premise, and it is what makes the slippage visible at all.
	choppy := leverResetWindows[0]
	if choppy.share[0] > 70 {
		t.Errorf("the daily arm keeps %.1f %% of the benchmark in the choppy decade: no slippage worth a plate", choppy.share[0])
	}
	// And the shape the plate is really about: both ends below the middle, in
	// both decades, for opposite reasons.
	for _, w := range leverResetWindows {
		if w.share[0] >= w.share[1] || w.share[2] >= w.share[1] {
			t.Errorf("%s: the shares %.1f / %.1f / %.1f no longer peak on the monthly reset",
				w.years, w.share[0], w.share[1], w.share[2])
		}
	}
	// The term arm's gearing, the honest cost of never resetting.
	if leverResetTermGear < 10 || leverResetTermTrough > 0.1 {
		t.Errorf("the term arm's 2008 (equity %.3f, leverage %.1f) no longer says what the plate claims",
			leverResetTermTrough, leverResetTermGear)
	}
}

// The term arm's 2008, quoted on the plate, recomputed.
func TestLevierResetTermArm2008(t *testing.T) {
	rec := leverResetReplay(t, 2000, 2009)
	if math.Abs(rec.trough-leverResetTermTrough) > 0.002 {
		t.Errorf("the term arm bottoms at %.4f, the plate quotes %.3f", rec.trough, leverResetTermTrough)
	}
	if math.Abs(rec.gear-leverResetTermGear) > 0.3 {
		t.Errorf("the term arm gears to %.2f, the plate quotes %.2f", rec.gear, leverResetTermGear)
	}
}

// The two years rendements-arithmetiques-geometriques quotes to show that the
// reset cuts both ways: they come out of the same replay, so they are guarded
// here rather than left as loose prose numbers.
func TestLevierResetBothWaysYearsMatchTheProse(t *testing.T) {
	art := bookArticle(t, "rendements-arithmetiques-geometriques")
	for _, want := range []string{
		"En 2013, le S&P 500 a gagné 32 % et un ×2 quotidien 73 %, soit bien plus que le double de l'indice",
		"En 2011, le même indice a gagné 2 % et le même produit a perdu 1 %",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	for _, c := range []struct {
		year        int
		index, dayl float64
	}{
		{2013, 32, 73},
		{2011, 2, -1},
	} {
		rec := leverResetReplay(t, c.year, c.year)
		if got := math.Round((rec.index - 1) * 100); got != c.index {
			t.Errorf("%d: the index made %+.0f %%, the article says %+.0f %%", c.year, got, c.index)
		}
		if got := math.Round((rec.mult[0] - 1) * 100); got != c.dayl {
			t.Errorf("%d: the daily x2 made %+.0f %%, the article says %+.0f %%", c.year, got, c.dayl)
		}
	}
}

// The rendered plate obeys the house rules and carries every reading it claims.
func TestLevierResetPlateRenders(t *testing.T) {
	svg := FigureSVG("levier-reset")
	if !strings.Contains(svg, `viewBox="0 0 640 470"`) {
		t.Error("the plate's viewBox moved: recheck the rendered PNG before freezing it")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"—", "–", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, arm := range leverResetArms {
		if !strings.Contains(svg, ">"+arm+"<") {
			t.Errorf("the plate does not name the %q column", arm)
		}
	}
	for _, w := range leverResetWindows {
		for i := range leverResetArms {
			if want := ">" + leverResetMultiple(w.mult[i]) + "<"; !strings.Contains(svg, want) {
				t.Errorf("the plate does not draw the reading %q", want)
			}
		}
		if !strings.Contains(svg, w.tag) {
			t.Errorf("the plate no longer names the %q decade", w.tag)
		}
	}
	for _, want := range []string{
		"repère 100 % : 2 × la croissance composée de l'indice, financement déduit",
		"creux de 2008 : ×0,03",
		"levier effectif ×33",
		"le beta slippage : 4,9 points en 2000-2009, 2,2 en 2010-2019.",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate no longer carries %q", want)
		}
	}
	// One dot per posed point, one line per decade, no other shape.
	if n := strings.Count(svg, "<circle"); n != len(leverResetArms)*len(leverResetWindows) {
		t.Errorf("%d dots drawn, expected one per posed point", n)
	}
	if n := strings.Count(svg, "<polyline"); n != len(leverResetWindows) {
		t.Errorf("%d curves drawn, expected one per decade", n)
	}
}

// The article the plate belongs to must carry it, and the cross-link to the
// article that defines the slippage it measures.
func TestLevierResetArticleCarriesThePlate(t *testing.T) {
	art := bookArticle(t, "levier-et-marges")
	for _, want := range []string{
		"::: figure levier-reset",
		"[[rendements-arithmetiques-geometriques]]",
		"beta slippage",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("levier-et-marges no longer carries %q", want)
		}
	}
}
