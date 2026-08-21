package firebook

import (
	"fmt"
	"strings"
)

// The plates of the sequence-of-returns article. The opening figure of that
// article is a schematic (two invented paths, same average, opposite order);
// this one is the historical counterpart: one rule, one capital, two start
// dates, both replotted from the day they retired so the two lives lie on the
// same axis.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.

// The 1966 and 1982 vintages of the book's reference household, replayed on the
// bundled real US 60/40 (S&P 500 total return + 5-year Treasuries, rebalanced
// every January, deflated by CPI-U). One rule on both sides, Bengen's fixed
// real withdrawal: 1 000 000 EUR of capital, 40 000 EUR a year indexed to
// inflation, a thirty-year plan, never adjusted.
//
// Reproduce with pkg/replay, for start = 1966 then start = 1982:
//
//	replay.Run(replay.Setup{Start: start, Capital: 1e6, Spend: 40000, Years: 30,
//	    Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5})
//
// then read Rules[0] (the fixed rule): the capital below is 1000 followed by
// Rule.Wealth/1000, in k EUR, one point per elapsed year of retirement;
// millCagr* is Result.CAGR and millDecade* is Result.Decade, in percent.
// figures_sequence_test.go recomputes all of it and fails on any drift.
//
// The 1966 series stops at the year the money ran out (elapsed year 29, the
// calendar year 1994, whose 40 k EUR could only be part paid); the 1982 series
// runs the full thirty years. Neither vintage is extrapolated: the record
// covers 1954 to 2025, so both windows are complete.
var (
	millRuinYear   = 1994
	millRuinAt     = 29 // elapsed years from the 1966 retirement to the zero
	millCagr1966   = 4.22
	millCagr1982   = 7.03
	millDecade1966 = -1.24
	millDecade1982 = 11.43
	millWealth1966 = []float64{1000, 885.8, 939.2, 932.2, 795.0, 782.1, 801.0, 829.3, 669.5, 486.1, 521.4, 549.7, 458.0, 400.1, 356.3, 338.3, 281.7, 289.5, 277.4, 251.2, 258.8, 254.1, 212.6, 185.4, 172.6, 127.5, 105.8, 68.4, 30.4, 0.0}
	millWealth1982 = []float64{1000, 1149.7, 1233.8, 1263.5, 1498.8, 1694.4, 1642.5, 1722.0, 1996.5, 1881.2, 2226.3, 2271.4, 2389.6, 2264.2, 2800.1, 3087.4, 3692.3, 4370.6, 4712.9, 4455.8, 4196.0, 3680.3, 4211.2, 4344.5, 4268.9, 4602.7, 4683.1, 3843.6, 4268.1, 4599.3, 4602.5}
)

// millPct renders an annualised real return the French way, with a true minus
// sign rather than a hyphen.
func millPct(v float64) string {
	if v < 0 {
		return "−" + frNum(-v, 1) + " %/an"
	}
	return "+" + frNum(v, 1) + " %/an"
}

// figMillesimes1966Vs1982 superposes the two vintages on one axis whose
// abscissa is the year of retirement rather than the calendar year, which is
// the only way to see that the difference is the start date and nothing else.
// The tinted band is everything below the starting stake: after its first year
// the 1966 household never climbs back out of it, and the 1982 one never falls
// back into it.
func figMillesimes1966Vs1982() string {
	const (
		x0, x1     = 78.0, 560.0
		yTop, yBot = 96.0, 296.0
		vMax       = 5.2 // M EUR
		span       = 30.0
	)
	x := func(i float64) float64 { return x0 + i/span*(x1-x0) }
	y := func(v float64) float64 { return yBot - v/vMax*(yBot-yTop) }
	path := func(ks []float64) [][2]float64 {
		out := make([][2]float64, len(ks))
		for i, k := range ks {
			out[i] = [2]float64{x(float64(i)), y(k / 1000)}
		}
		return out
	}

	var b strings.Builder
	b.WriteString(plateHead("millésime 1966 contre millésime 1982",
		"Le même plan, deux dates de départ"))
	b.WriteString(plateDeck(
		"capital réel restant, en millions d'euros"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"même plan des deux côtés : 1 M€, 40 k€ retirés chaque année, indexés sur l'inflation, jamais ajustés"))

	// Everything under the starting stake, tinted once.
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, y(1), x1-x0, y(0)-y(1), figWash)
	for _, g := range []float64{0, 1, 2, 3, 4} {
		col := figGrid
		if g == 0 || g == 1 {
			col = figRule
		}
		b.WriteString(line(x0, y(g), x1, y(g), col, 1))
		b.WriteString(mTxt(x0-8, y(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(x0+6, y(0)-5, 10, figMuted, "start", "400", "sous la mise de départ"))

	// The years of retirement, not the calendar years: the whole point.
	for i := 0.0; i <= span; i += 5 {
		b.WriteString(line(x(i), yBot, x(i), yBot+4, figRule, 1))
		b.WriteString(mTxt(x(i), yBot+20, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", i)))
	}
	b.WriteString(sTxt(x0, yBot+38, 10.5, figMuted, "start", "400",
		"années écoulées depuis le départ à la retraite"))

	// The two lives.
	b.WriteString(poly(path(millWealth1982), figBlue, 2.4, ""))
	b.WriteString(poly(path(millWealth1966), figBad, 2.4, ""))

	last := millWealth1982[len(millWealth1982)-1] / 1000
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, x(span), y(last), figBlue)
	b.WriteString(mTxt(x(span)+8, y(last)+3.8, 10.5, figBlue, "start", "600", frNum(last, 1)+" M€"))
	b.WriteString(sTxt(x(18), 110, 11, figBlue, "middle", "600", "millésime 1982"))

	rx := x(float64(millRuinAt))
	b.WriteString(sTxt(538, 216, 11, figBad, "end", "600", "millésime 1966"))
	b.WriteString(sTxt(538, 232, 10.5, figBad, "end", "400",
		fmt.Sprintf("zéro en %d, l'année %d", millRuinYear, millRuinAt)))
	b.WriteString(dashLine(rx, 240, rx, yBot-7, figBad, 1, "2 3"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, rx, y(0), figBad)

	// The reading below the chart: the returns each household actually got, and
	// what it had left. The 1966 column is the one that stings.
	b.WriteString(line(24, 342, 616, 342, figGrid, 1))
	b.WriteString(sTxt(300, 358, 9.5, figMuted, "middle", "400", "réel, dix premières années"))
	b.WriteString(sTxt(430, 358, 9.5, figMuted, "middle", "400", "réel, trente ans"))
	b.WriteString(sTxt(556, 358, 9.5, figMuted, "middle", "400", "à l'arrivée"))
	for _, r := range []struct {
		y              float64
		col, name, end string
		decade, thirty float64
	}{
		{378, figBad, "départ en 1966", "zéro", millDecade1966, millCagr1966},
		{398, figBlue, "départ en 1982", frNum(last, 1) + " M€", millDecade1982, millCagr1982},
	} {
		b.WriteString(line(24, r.y-4, 40, r.y-4, r.col, 2.6))
		b.WriteString(sTxt(48, r.y, 10.5, r.col, "start", "600", r.name))
		b.WriteString(mTxt(300, r.y, 10.5, figSoft, "middle", "600", millPct(r.decade)))
		b.WriteString(mTxt(430, r.y, 10.5, figSoft, "middle", "600", millPct(r.thirty)))
		b.WriteString(sTxt(556, r.y, 10.5, r.col, "middle", "600", r.end))
	}

	b.WriteString(plateConclusion(420,
		"Le portefeuille de 1966 a pourtant rapporté plus que le retrait. C'est l'ordre des années qui a tué le plan."))
	b.WriteString(sTxt(24, 436, 10, figMuted, "start", "400",
		"60/40 américain réel (S&amp;P 500, Treasuries 5 ans, déflatés CPI-U), reconstruction du livre ; retrait fixe, sans fiscalité."))
	return svg(640, 450, b.String())
}
