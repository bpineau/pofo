package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
)

// The PPUT plate's window: the value of June 1986 is the base and December 2018
// the last reading, so the ladder covers the 390 months Cboe and Wilshire quote
// the protection index over.
const (
	pdFirstYear, pdFirstMonth = 1986, 6
	pdLastYear, pdLastMonth   = 2018, 12
	pdMonths                  = (pdLastYear-pdFirstYear)*12 + (pdLastMonth - pdFirstMonth)
)

// pdMonthKeys lists the month keys ("1986-06" ... "2018-12") of the window.
func pdMonthKeys() []string {
	out := make([]string, 0, pdMonths+1)
	y, m := pdFirstYear, pdFirstMonth
	for y < pdLastYear || (y == pdLastYear && m <= pdLastMonth) {
		out = append(out, fmt.Sprintf("%04d-%02d", y, m))
		if m++; m == 13 {
			y, m = y+1, 1
		}
	}
	return out
}

// pdRecompute measures one equity/bond mix the way the plate's comment says:
// monthly total returns of the two legs, weights reset every December, no
// deflator (the protection index is quoted in nominal total return). It returns
// the annualized return and the deepest month-end drawdown, both in percent.
func pdRecompute(equity float64, sp, bonds map[string]float64) (cagr, drawdown float64) {
	mons := pdMonthKeys()
	e, b := equity, 1-equity
	index := make([]float64, 1, len(mons))
	index[0] = 1
	for i := 1; i < len(mons); i++ {
		prev, cur := mons[i-1], mons[i]
		e *= sp[cur] / sp[prev]
		b *= bonds[cur] / bonds[prev]
		total := e + b
		index = append(index, total)
		if strings.HasSuffix(cur, "-12") {
			e, b = total*equity, total*(1-equity)
		}
	}
	cagr = (math.Pow(index[len(index)-1], 12/float64(pdMonths)) - 1) * 100
	peak := index[0]
	for _, v := range index {
		if v > peak {
			peak = v
		}
		if d := (v/peak - 1) * 100; d < drawdown {
			drawdown = d
		}
	}
	return cagr, drawdown
}

// The eleven rungs are frozen literals so the book's figures stay pure
// functions with no data dependency at render time. This rebuilds each of them
// from the bundled series and fails the moment the plate and the data disagree,
// which is also what happens when those series are regenerated.
func TestPutsLadderMatchesTheData(t *testing.T) {
	sim := datasets.Simdata()
	sp := ttMonthEnds(t, sim, "SP500")
	bonds := ttMonthEnds(t, sim, "IEF")
	if len(pdMonthKeys()) != pdMonths+1 {
		t.Fatalf("the window holds %d month ends, expected %d", len(pdMonthKeys()), pdMonths+1)
	}
	for i, r := range putsLadder {
		equity := 0.5 + 0.05*float64(i)
		cagr, dd := pdRecompute(equity, sp, bonds)
		if math.Abs(cagr-r.cagr) > 0.03 {
			t.Errorf("%s: return %.2f %%, plate says %.2f %%", r.name, cagr, r.cagr)
		}
		if math.Abs(dd-r.drawdn) > 0.1 {
			t.Errorf("%s: worst drawdown %.2f %%, plate says %.2f %%", r.name, dd, r.drawdn)
		}
	}
}

// What makes the plate legitimate is that its recomputed ladder and the
// published protection index sit on the same footing. The proof is the top
// rung: 100 % equities must reproduce the S&P 500 numbers Cboe and Wilshire
// publish over the very same 390 months, 9.8 % and −50.95 %.
func TestPutsLadderAnchorsOnThePublishedIndex(t *testing.T) {
	top := putsLadder[len(putsLadder)-1]
	if top.name != "100/0" {
		t.Fatalf("the ladder must end on 100 %% equities, it ends on %s", top.name)
	}
	if math.Abs(top.cagr-9.8) > 0.05 {
		t.Errorf("the 100/0 rung returns %.2f %%, the published S&P 500 returned 9.8 %%", top.cagr)
	}
	if math.Abs(top.drawdn-(-50.95)) > 0.05 {
		t.Errorf("the 100/0 rung plunges %.2f %%, the published S&P 500 plunged −50.95 %%", top.drawdn)
	}
}

// The plate draws three claims about the published point, and all three must
// survive an edit of the data or of the frozen arrays.
func TestPutsDominationHolds(t *testing.T) {
	for i, r := range putsLadder {
		if i > 0 && (r.cagr <= putsLadder[i-1].cagr || r.drawdn >= putsLadder[i-1].drawdn) {
			t.Errorf("the ladder must buy return with drawdown: %s (%.2f %%, %.2f %%) after %s",
				r.name, r.cagr, r.drawdn, putsLadder[i-1].name)
		}
	}
	// 1. the shaded quadrant: every rung from 50/50 to 75/25 beats the protected
	// point on both axes at once.
	for _, name := range []string{"50/50", "55/45", "60/40", "65/35", "70/30", "75/25"} {
		r := putsRungAt(name)
		if r.cagr <= putsProtected.cagr || r.drawdn <= putsProtected.drawdn {
			t.Errorf("%s (%.2f %%, %.2f %%) no longer dominates the protected point (%.2f %%, %.2f %%)",
				name, r.cagr, r.drawdn, putsProtected.cagr, putsProtected.drawdn)
		}
	}
	// and the rung just outside it must indeed be outside, otherwise the plate
	// names the wrong segment.
	if putsRungAt("80/20").drawdn > putsProtected.drawdn {
		t.Error("the 80/20 rung is shallower than the protected point: the quadrant must be redrawn")
	}
	// 2. the read at equal drawdown: 80/20 sinks as deep, to within half a point.
	eq := putsRungAt("80/20")
	if math.Abs(eq.drawdn-putsProtected.drawdn) > 0.6 {
		t.Errorf("the 80/20 rung plunges %.2f %% against %.2f %%, no longer an equal-drawdown read",
			eq.drawdn, putsProtected.drawdn)
	}
	// 3. and it pays 2.8 points a year more, the number the plate prints.
	if gap := eq.cagr - putsProtected.cagr; math.Abs(gap-2.8) > 0.05 {
		t.Errorf("the equal-drawdown gap is %.2f points, the plate prints 2.8", gap)
	}
}

// The published numbers the plate places by hand are the ones the article
// states in prose; figure and text must not drift apart.
func TestPutsFigureAgreesWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/long-volatility.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	if !strings.Contains(article, "::: figure puts-domines") {
		t.Error("the article must carry the plate")
	}
	for _, want := range []string{
		"6,6 %",   // PPUT, published
		"9,8 %",   // the S&P 500 over the same window, published and recomputed
		"−38,9 %", // the two drawdowns
		"−51 %",
		"3,2 points", // the yearly cost, which is the difference of the two returns
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer states %q, which the plate draws", want)
		}
	}
	// the cost the article quotes is exactly the gap between the two published
	// returns, so the prose cannot drift from the plate's own arithmetic.
	if gap := putsLadder[len(putsLadder)-1].cagr - putsProtected.cagr; math.Abs(gap-3.2) > 0.05 {
		t.Errorf("the published cost is %.2f points a year, the article says 3.2", gap)
	}
}
