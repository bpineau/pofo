package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The hold-to-maturity plate of the bond article. Its central sentence is that
// "tenir à échéance" protects nothing: after a rate rise the holder of a single
// bond takes exactly the same loss as the fund, as an opportunity cost instead
// of a price fall, and the final wealth is identical. The article states it in
// its own numbers, a 2 % coupon against a 4 % market, on the seven-year
// duration it calls the standard aggregate.
//
// Prose cannot show two accountings of the same position at once. Two curves
// on one timeline can: they run together for a year, split at the shock, and
// land on the same point at the horizon. The band between them is the loss,
// tinted toward the fund's colour on its side and toward the bond's on the
// other, and its two names are the plate.
//
// Everything here is closed-form bond arithmetic, no market data of any kind.

// The article's own example, and the conventions the two curves are computed
// under:
//
//   - a FLAT yield curve, at 2 % until the shock and 4 % after it, static from
//     then on. On a flat static curve every bond earns exactly its yield, which
//     is what makes the convergence below exact rather than approximate;
//   - the shock is a single instantaneous parallel shift of +2 points, at the
//     twelfth month;
//   - the plan's HORIZON equals the position's remaining maturity, seven years
//     from the purchase, which is the immunization condition the article states
//     as "la duration de la poche doit rester inférieure ou égale à l'horizon";
//   - the capital is 100 and everything is nominal, since the article's claim is
//     about nominal accounting.
const (
	tenirCapital = 100.0
	tenirYield0  = 0.02
	tenirYield1  = 0.04
	tenirHorizon = 7.0 // years, and the remaining maturity at the purchase
	tenirShock   = 1.0 // years: the twelfth month
)

// tenirFace is what the position pays at the horizon: the capital compounded at
// the yield it was bought on. It is fixed at the purchase and no later rate
// changes it, which is the whole reason the two accountings converge.
func tenirFace() float64 { return tenirCapital * math.Pow(1+tenirYield0, tenirHorizon) }

// tenirYield is the market yield at t, in years since the purchase.
func tenirYield(t float64) float64 {
	if t < tenirShock {
		return tenirYield0
	}
	return tenirYield1
}

// tenirBook is the wealth the holder of the single bond SEES at t: amortized
// cost, the capital accreting at the yield he bought on, whatever the market
// does. It never falls, and it reaches the face exactly at the horizon.
func tenirBook(t float64) float64 {
	return tenirCapital * math.Pow(1+tenirYield0, t)
}

// tenirMarket is the wealth the SAME position is worth at t on the market, which
// is what a fund holding it prints: the face discounted at the current yield
// over the maturity left. Before the shock the two are the same number; after
// it, the discount is harsher and the value falls, then accretes at the new,
// higher yield. At t = horizon there is no maturity left to discount, so it is
// the face again: the convergence is an identity, not an approximation.
func tenirMarket(t float64) float64 {
	return tenirFace() / math.Pow(1+tenirYield(t), tenirHorizon-t)
}

// tenirGap is the distance between the two accountings at t, in points of the
// capital: what the fund's holder has already suffered on screen, and what the
// bond's holder is silently owed by his frozen coupon.
func tenirGap(t float64) float64 { return tenirBook(t) - tenirMarket(t) }

// tenirDrop is the price fall the shock prints, as a fraction: the market value
// just after the shock against the book value it left.
func tenirDrop() float64 { return tenirMarket(tenirShock)/tenirBook(tenirShock) - 1 }

// tenirRecoveryMonth is the first month at which the printed price is back to
// its pre-shock level. Nothing was lost by then and nothing had been lost
// before; the four years are the ones the fund's holder spends looking at a
// loss he will not take.
func tenirRecoveryMonth() int {
	before := tenirBook(tenirShock)
	for m := int(tenirShock * 12); m <= int(tenirHorizon*12); m++ {
		if tenirMarket(float64(m)/12) >= before {
			return m
		}
	}
	return 0
}

// The plate's geometry.
const (
	tenX0, tenX1   = 70.0, 600.0
	tenTop, tenBot = 110.0, 310.0
	tenValLo       = 88.0
	tenValHi       = 118.0
)

func tenirScales() (x, y figScale) {
	return figScale{Min: 0, Max: tenirHorizon, Px0: tenX0, Px1: tenX1},
		figScale{Min: tenValLo, Max: tenValHi, Px0: tenBot, Px1: tenTop}
}

// tenirCurve samples one accounting month by month, in pixels. The market curve
// carries the shock's discontinuity: the month of the shock appears twice, once
// on the old yield and once on the new, which draws the fall as the vertical
// drop it is.
func tenirCurve(f func(float64) float64, withDrop bool) [][2]float64 {
	x, y := tenirScales()
	var pts [][2]float64
	for m := 0; m <= int(tenirHorizon*12); m++ {
		t := float64(m) / 12
		if withDrop && t == tenirShock {
			pts = append(pts, [2]float64{x.Map(t), y.Map(tenirBook(t))})
		}
		pts = append(pts, [2]float64{x.Map(t), y.Map(f(t))})
	}
	return pts
}

func figTenirEcheance() string {
	x, y := tenirScales()
	var b strings.Builder
	b.WriteString(plateHead("tenir à échéance", "Deux comptabilités, une seule richesse"))
	b.WriteString(plateDeck(
		"Le même portefeuille obligataire à sept ans, après une hausse de deux points au douzième mois"))
	legendChips(&b, 74, [][2]string{
		{figDeep, "le titre détenu à échéance"},
		{figBlue, "le fonds, valorisé au marché"},
	})

	axisTicks(&b, y, []float64{90, 95, 100, 105, 110, 115}, 0, "", tenX0, tenX1, false)
	for yr := 0.0; yr <= tenirHorizon; yr++ {
		b.WriteString(mTxt(x.Map(yr), tenBot+18, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", yr)))
	}
	b.WriteString(sTxt(tenX0, tenBot+37, 9.5, figMuted, "start", "400", "années depuis l'achat  →"))

	// The band between the two accountings, from the shock to the horizon, in
	// two halves: the fund's side tinted toward the fund's colour, the bond's
	// side toward the bond's. It is one loss, and the two halves are the two
	// names it goes by.
	book, market := tenirBandPoints(tenirBook), tenirBandPoints(tenirMarket)
	mid := make([][2]float64, len(book))
	for i := range book {
		mid[i] = [2]float64{book[i][0], (book[i][1] + market[i][1]) / 2}
	}
	b.WriteString(tenirFill(book, mid, mixHex("#fffdf9", figDeep, 0.13)))
	b.WriteString(tenirFill(mid, market, mixHex("#fffdf9", figBlue, 0.13)))

	// The shock, drawn as the thing it is: a rule down to the price it wipes
	// out, then the cote of what it wiped out.
	sx := x.Map(tenirShock)
	b.WriteString(line(sx, tenTop, sx, y.Map(tenirBook(tenirShock)), figRule, 1))
	b.WriteString(sTxt(sx+8, 124, 10, figSoft, "start", "600", "+2 points de taux, au mois 12"))
	b.WriteString(line(sx, y.Map(tenirBook(tenirShock)), sx, y.Map(tenirMarket(tenirShock)), figMuted, 1.2))
	b.WriteString(line(sx, y.Map(tenirMarket(tenirShock)), sx+5, y.Map(tenirMarket(tenirShock)), figMuted, 1.2))
	b.WriteString(mTxt(sx+8, (y.Map(tenirBook(tenirShock))+y.Map(tenirMarket(tenirShock)))/2+4,
		10.5, figSoft, "start", "600", frNum(tenirGap(tenirShock), 1)+" points"))

	b.WriteString(poly(tenirCurve(tenirMarket, true), figBlue, 2.4, ""))
	b.WriteString(poly(tenirCurve(tenirBook, false), figDeep, 2.4, ""))

	// The two names of the same band. They sit just OUTSIDE the band, each
	// against its own curve: the band tilts upward too fast for a horizontal
	// label to stay inside a half over its own width, and a sloped label is not
	// something this book draws. The tints keep saying which half is whose.
	// The bond's label is hung from the RIGHT end of its run and the fund's from
	// the LEFT end of its own, so that each curve rises away from its label
	// instead of through it.
	const bookLabel, marketLabel = 3.9, 2.0 // years
	b.WriteString(sTxt(x.Map(bookLabel), y.Map(tenirBook(bookLabel))-12, 10.5, figDeep, "end", "600",
		"coût d'opportunité invisible"))
	b.WriteString(sTxt(x.Map(marketLabel), y.Map(tenirMarket(marketLabel))+20, 10.5, figBlue, "start", "600",
		"baisse de prix affichée"))

	// The convergence, marked where it happens.
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.5" fill="%s"/>`,
		tenX1, y.Map(tenirFace()), figInk)
	b.WriteString(sTxt(tenX1-8, 118, 10, figSoft, "end", "600",
		"même richesse finale : "+frNum(tenirFace(), 2)+" pour 100 investis"))

	b.WriteString(plateConclusion(372,
		"La perte est la même des deux côtés : affichée en prix chez l'un, invisible dans le coupon chez l'autre."))
	b.WriteString(plateFoot(394, []string{
		"Arithmétique obligataire fermée, aucune donnée de marché. Capital 100, courbe plate à 2 % portée à 4 % " +
			"au mois 12 puis stable,",
		"horizon 7 ans, égal à la maturité restante. Le fonds valorise au marché : le prix tombe de " +
			frNum(-tenirDrop()*100, 1) + " % au choc, puis",
		"s'accrète au nouveau taux. Le titre reste au coût amorti : aucune baisse à l'écran, mais un coupon figé " +
			"à 2 % quand le marché sert 4 %.",
		"Les deux finissent sur le même nombre, au centime près : c'est le même actif, lu sur deux comptabilités. " +
			"L'écart culmine à",
		frNum(tenirGap(tenirShock), 1) + " points juste après le choc, et il faut " +
			fmt.Sprint(tenirRecoveryMonth()) + " mois au prix affiché pour retrouver son niveau d'avant.",
		"L'égalité tient parce que l'horizon égale la maturité restante. Un fonds qui roule pour garder 7 ans de " +
			"duration finit, lui, sous le titre.",
	}))
	return svg(640, 490, b.String())
}

// tenirBandPoints samples one accounting from the shock to the horizon, which
// is the stretch the band covers.
func tenirBandPoints(f func(float64) float64) [][2]float64 {
	x, y := tenirScales()
	var pts [][2]float64
	for m := int(tenirShock * 12); m <= int(tenirHorizon*12); m++ {
		t := float64(m) / 12
		pts = append(pts, [2]float64{x.Map(t), y.Map(f(t))})
	}
	return pts
}

// tenirFill closes the ribbon between two sampled curves into one opaque
// filled path (the book's figures never ship a translucent fill).
func tenirFill(top, bottom [][2]float64, fill string) string {
	var d strings.Builder
	for i, p := range top {
		verb := "L"
		if i == 0 {
			verb = "M"
		}
		fmt.Fprintf(&d, "%s %.1f,%.1f ", verb, p[0], p[1])
	}
	for i := len(bottom) - 1; i >= 0; i-- {
		fmt.Fprintf(&d, "L %.1f,%.1f ", bottom[i][0], bottom[i][1])
	}
	return fmt.Sprintf(`<path d="%sZ" fill="%s"/>`, d.String(), fill)
}
