package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The inflation part's pivot plate: the same cumulative loss of purchasing
// power, once concentrated into a short episode and once spread over the whole
// retirement, and what each of the two does to a withdrawal plan.
//
// THE PLAN is the book's reference household, the one bengen-falaise reads:
// 600 k EUR, 24 k EUR a year indexed to prices (4.0 % to start), 30 years.
//
// THE MARKET is identical in both worlds: one constant NOMINAL return, the
// plan's 4 % real compounded with the 2 % inflation it expected. Assets
// therefore do not react to inflation at all, which is exactly the assumption
// the article's first mechanism makes ("le portefeuille nominal ne suit pas").
//
// THE TWO PATHS end at the same price level by construction: the drift rate is
// SOLVED, not chosen, from (1+d)^30 = 1.08^5 * 1.02^25. It comes out at 2.976 %
// a year, that is two hundredths of a point away from the "un point de plus que
// prévu" the article names in its opening paragraph.
//
// Because both paths carry the same 30-year price product AND the market pays
// the same nominal return, the two worlds earn EXACTLY the same 30-year real
// market return. The only thing that differs is the ORDER of the real returns,
// which is sequence risk in its purest form, and it is the whole point of the
// plate: the two price curves meet, the two withdrawal-rate curves do not.
//
// Everything is computed at render time from the constants below; nothing is
// frozen. figures_infl_wr_test.go re-derives the unroll independently and
// checks every number the plate prints.
const (
	inflCapital  = 600.0 // k EUR at departure
	inflSpend    = 24.0  // k EUR a year, indexed to prices, never adjusted
	inflYears    = 30
	inflRealPlan = 0.04 // the 60/40's real return the 4 % plan is built on
	inflExpected = 0.02 // the inflation the plan expected
	inflShock    = 0.08 // the episode's inflation
	inflShockLen = 5    // for five years
)

// inflNominal is the constant nominal return both worlds earn, year in, year
// out: the planned real return grossed up by the planned inflation.
func inflNominal() float64 { return (1+inflRealPlan)*(1+inflExpected) - 1 }

// inflDrift solves for the steady inflation rate that reaches, after thirty
// years, the very price level the episode path reaches.
func inflDrift() float64 {
	total := inflShockLen*math.Log(1+inflShock) + (inflYears-inflShockLen)*math.Log(1+inflExpected)
	return math.Exp(total/inflYears) - 1
}

// inflPaths returns the three yearly inflation paths of the plate: the episode
// (five years of shock, then the planned rate again), the drift (the solved
// steady rate) and the plan as it was written (the expected rate throughout).
func inflPaths() (episode, drift, planned []float64) {
	d := inflDrift()
	for y := 0; y < inflYears; y++ {
		rate := inflExpected
		if y < inflShockLen {
			rate = inflShock
		}
		episode = append(episode, rate)
		drift = append(drift, d)
		planned = append(planned, inflExpected)
	}
	return episode, drift, planned
}

// inflUnroll walks the plan through one inflation path and returns, for each of
// the thirty-one 1 January from departure to the end of the horizon, the price
// level (base 100) and the current withdrawal rate, in percent: the year's
// spend over the portfolio standing that morning, in real terms. That is the
// same reading bengen-falaise and wr-signal show.
func inflUnroll(path []float64) (price, rate []float64) {
	value, level, nominal := inflCapital, 100.0, inflNominal()
	for _, pi := range path {
		price = append(price, level)
		rate = append(rate, inflSpend/value*100)
		// The market pays its nominal return, prices eat their share of it.
		value = (value - inflSpend) * (1 + nominal) / (1 + pi)
		level *= 1 + pi
	}
	return append(price, level), append(rate, inflSpend/value*100)
}

func figInflationEpisodeDerive() string {
	const (
		x0, x1 = 72.0, 470.0
		gut    = x1 + 8 // the direct-label gutter, right of the plot
		// top panel: the price level, 95 to 250 (base 100)
		aTop, aBot = 110.0, 258.0
		aLo, aHi   = 95.0, 250.0
		// bottom panel: the current withdrawal rate, 3.5 to 22 %
		bTop, bBot = 302.0, 458.0
		bLo, bHi   = 3.5, 22.0
	)
	x := func(year float64) float64 { return x0 + year/inflYears*(x1-x0) }
	ya := func(v float64) float64 { return aBot - (v-aLo)/(aHi-aLo)*(aBot-aTop) }
	yb := func(v float64) float64 { return bBot - (v-bLo)/(bHi-bLo)*(bBot-bTop) }
	curve := func(vals []float64, y func(float64) float64) [][2]float64 {
		pts := make([][2]float64, len(vals))
		for i, v := range vals {
			pts[i] = [2]float64{x(float64(i)), y(v)}
		}
		return pts
	}

	epPath, drPath, plPath := inflPaths()
	epPrice, epRate := inflUnroll(epPath)
	drPrice, drRate := inflUnroll(drPath)
	plPrice, plRate := inflUnroll(plPath)
	last := inflYears

	var b strings.Builder
	b.WriteString(plateHead("l'inflation et le taux de retrait",
		"Même perte cumulée, deux formes : l'épisode contre la dérive"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"600 k€, 24 k€ par an indexés sur les prix (4,0 % au départ), 30 ans."))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"Les mêmes actifs dans les deux mondes : 6,1 % nominal par an, soit 4 % réel si l'inflation tient les 2 % prévus."))

	// --- top panel: two price paths that end at the same level --------------
	b.WriteString(sTxt(24, 98, 10.5, figSoft, "start", "600",
		"Le niveau des prix, base 100 au départ"))
	for _, g := range []float64{100, 150, 200, 250} {
		col := figGrid
		if g == 100 {
			col = figRule
		}
		b.WriteString(line(x0, ya(g), x1, ya(g), col, 1))
		b.WriteString(mTxt(x0-8, ya(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	// the guide that ties the two panels: the year the episode ends
	b.WriteString(dashLine(x(inflShockLen), aTop, x(inflShockLen), bBot, figMuted, 1, "2 4"))

	b.WriteString(poly(curve(plPrice, ya), figMuted, 1.4, "4 3"))
	b.WriteString(poly(curve(drPrice, ya), figBlue, 2.2, ""))
	b.WriteString(poly(curve(epPrice, ya), figBad, 2.4, ""))

	b.WriteString(sTxt(x(inflShockLen)+6, aTop+12, 10, figMuted, "start", "400", "fin de l'épisode"))
	b.WriteString(sTxt(x(7), ya(212), 10.5, figBad, "start", "600",
		"l'épisode : cinq ans à 8 %, puis retour à 2 %"))
	// the drift, named under its own curve, with a hairline back up to it
	b.WriteString(line(x(16), ya(123)-7, x(16), ya(157), figBlue, 1))
	b.WriteString(sTxt(x(16)+5, ya(123), 10.5, figBlue, "start", "600",
		fmt.Sprintf("la dérive : %s %% par an", frNum(inflDrift()*100, 2))))

	// the meeting point, the whole reason the plate exists
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.6" fill="none" stroke="%s" stroke-width="1.8"/>`,
		x1, ya(epPrice[last]), figInk)
	b.WriteString(sTxt(gut, ya(epPrice[last])-6, 11, figInk, "start", "600", "les deux se rejoignent"))
	b.WriteString(sTxt(gut, ya(epPrice[last])+10, 10, figMuted, "start", "400", "même perte cumulée :"))
	b.WriteString(mTxt(gut, ya(epPrice[last])+25, 11, figDeep, "start", "600",
		"×"+frNum(epPrice[last]/100, 2)))
	b.WriteString(sTxt(gut, ya(plPrice[last])-5, 10.5, figMuted, "start", "600", "le plan prévu"))
	b.WriteString(mTxt(gut, ya(plPrice[last])+11, 10.5, figMuted, "start", "400",
		"×"+frNum(plPrice[last]/100, 2)))

	// --- bottom panel: the same plan's warning light under each -------------
	b.WriteString(sTxt(24, 288, 10.5, figSoft, "start", "600",
		"Le taux de retrait courant, lu chaque 1er janvier : retrait de l'année / portefeuille"))
	// the article's severity bands, the ones bengen-falaise draws, under everything
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, bTop, x1-x0, yb(10)-bTop, figBadWash)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, yb(10), x1-x0, yb(8)-yb(10), figWash)
	b.WriteString(sTxt(x0+8, yb(8.9), 10, figDeep, "start", "600", "la zone d'alerte : 8 à 10 %"))
	b.WriteString(sTxt(x0+8, bTop+16, 10, figBad, "start", "600", "au-delà, aucun marché ne rattrape"))
	for _, g := range []float64{4, 8, 12, 16, 20} {
		col := figGrid
		if g == 4 {
			col = figRule
		}
		b.WriteString(line(x0, yb(g), x1, yb(g), col, 1))
		b.WriteString(mTxt(x0-8, yb(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}

	b.WriteString(poly(curve(plRate, yb), figMuted, 1.4, "4 3"))
	b.WriteString(poly(curve(drRate, yb), figBlue, 2.2, ""))
	b.WriteString(poly(curve(epRate, yb), figBad, 2.4, ""))

	// the year the episode's plan enters the alert zone, with the horizon left
	alert := inflAlertYear(epRate)
	ax, ay := x(float64(alert)), yb(epRate[alert])
	b.WriteString(dashLine(ax, ay-7, ax, yb(13.2), figBad, 1, "2 3"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, ax, ay, figBad)
	b.WriteString(sTxt(x(8.5), yb(15.6), 10.5, figBad, "start", "600",
		"l'épisode entre dans la zone d'alerte"))
	b.WriteString(sTxt(x(8.5), yb(14.1), 10.5, figBad, "start", "400",
		fmt.Sprintf("à l'année %d, avec %d ans encore à financer", alert, inflYears-alert)))

	// direct end labels, one per path
	type endLbl struct {
		v     float64
		name  string
		color string
	}
	for _, e := range []endLbl{
		{epRate[last], "l'épisode", figBad},
		{drRate[last], "la dérive", figBlue},
		{plRate[last], "le plan prévu", figMuted},
	} {
		b.WriteString(sTxt(gut, yb(e.v)-5, 11, e.color, "start", "600", e.name))
		b.WriteString(mTxt(gut, yb(e.v)+11, 11, e.color, "start", "600", frNum(e.v, 1)+" %"))
	}

	// --- the shared time axis ----------------------------------------------
	b.WriteString(line(x0, bBot, x1, bBot, figRule, 1))
	for _, t := range []float64{0, 5, 10, 15, 20, 25, 30} {
		b.WriteString(mTxt(x(t), 474, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt((x0+x1)/2, 494, 11, figMuted, "middle", "400", "années de retraite"))
	return svg(640, 508, b.String())
}

// inflAlertYear is the first January the reading enters the 8 to 10 % band.
func inflAlertYear(rate []float64) int {
	for i, v := range rate {
		if v >= 8 {
			return i
		}
	}
	return -1
}
