package firebook

import (
	"fmt"
	"strings"
)

// The matched-floor plate. The article's closing example states a balance
// sheet and an income in the same paragraph, and the reader has to hold both
// in their head at once to see the point. The point is a correspondence: a
// rung bought today IS a year of floor paid later, and once the floor is
// bought the rest of the money owes almost nothing.
//
// So the plate is two panels. On the left, what the couple buys: the indexed
// ladder, drawn as its fourteen rungs, next to the comfort portfolio. On the
// right, what they get, year by year from 52 to 75: the rungs pay the floor
// until the pensions take over at 66, with the comfort layer on top all the
// way. Each rung wears its own tint on the left and wears it again on the year
// it pays on the right, which is the stock-to-flow link drawn without a single
// arrow.
//
// No market data is involved: everything is the discounting arithmetic of the
// article's own example, held in pure functions the guard test confronts with
// the numbers the prose quotes.

// The household of the article's example, in k EUR of today's money. The
// couple retires at 52 on a floor they will not go under and a comfort level
// they aim at; the pensions cover the floor at 66, so the uncovered phase is
// fourteen years long and takes fourteen rungs.
const (
	floorLadderAge     = 52   // the year the plan starts
	floorLadderPension = 66   // the pensions cover the floor from here on
	floorLadderLast    = 75   // the right panel stops here, ten years into the pensions
	floorLadderRungs   = 14   // one rung per uncovered year, 52 to 65
	floorLadderFloor   = 40.0 // the floor, matched by a rung or paid by the pensions
	floorLadderComfort = 55.0 // the level the couple actually lives at
	floorLadderReal    = 0.01 // the real yield the rungs are bought at
	floorLadderWealth  = 1600.0
)

// The rung tints: one blue family, light at the near rungs, dark at the far
// ones, so the left panel reads as a ladder and the right panel as a sweep
// through it. Both ends are opaque hex, never rgba (crengine, KOReader's EPUB
// SVG renderer, paints rgba solid black), and every step between them is
// mixed to another opaque hex.
const (
	floorLadderLight = "#A9C0DE"
	floorLadderDark  = "#22406E"
	floorLadderPaper = "#fffdf9"
)

// floorLadderRungCost is what the rung of rank k (1-based) costs today, in
// k EUR: it pays the floor at the end of year k, discounted at the market real
// yield. The far rungs are the cheap ones, which is the whole reason the
// fourteen years of floor cost less than the fourteen payments add up to.
func floorLadderRungCost(k int) float64 {
	d := 1.0
	for i := 0; i < k; i++ {
		d *= 1 + floorLadderReal
	}
	return floorLadderFloor / d
}

// floorLadderCost is the price of the whole ladder, in k EUR: the
// annuity-certain of fourteen yearly payments of the floor at the market real
// yield. The article quotes it as ~520 000 EUR.
func floorLadderCost() float64 {
	sum := 0.0
	for k := 1; k <= floorLadderRungs; k++ {
		sum += floorLadderRungCost(k)
	}
	return sum
}

// floorLadderRest is what stays outside the ladder, in k EUR, and
// floorLadderRate the withdrawal rate the comfort layer asks of it: the
// article's 1,4 %, the number the whole example exists to produce.
func floorLadderRest() float64 { return floorLadderWealth - floorLadderCost() }

func floorLadderRate() float64 {
	return (floorLadderComfort - floorLadderFloor) / floorLadderRest() * 100
}

// floorLadderTint is the colour of the rung of rank k (1-based), mixed along
// the blue family from the near rungs to the far ones.
func floorLadderTint(k int) string {
	t := float64(k-1) / float64(floorLadderRungs-1)
	return mixHex(floorLadderLight, floorLadderDark, t)
}

// figLinkersPlancherAdosse draws the two panels: the balance sheet that buys
// the floor, and the income that floor buys.
func figLinkersPlancherAdosse() string {
	const (
		// left panel: the stock, in k EUR
		lBar1, lBar2 = 68.0, 150.0
		lBarW        = 50.0
		lBase        = 340.0
		lTop         = 132.0 // the comfort portfolio's height
		// right panel: the flow, in k EUR per year
		rx0, rx1 = 262.0, 616.0
		rBase    = 340.0
		rTop     = 140.0 // the comfort level, and the flat top of every year
	)
	rest := floorLadderRest()
	yL := func(v float64) float64 { return lBase - v*(lBase-lTop)/rest }
	yR := func(v float64) float64 { return rBase - v*(rBase-rTop)/floorLadderComfort }
	years := floorLadderLast - floorLadderAge + 1
	slot := (rx1 - rx0) / float64(years)
	barW := slot - 3.4
	rLeft := func(i int) float64 { return rx0 + slot*float64(i) + 1.7 }
	rMid := func(i int) float64 { return rLeft(i) + barW/2 }

	var b strings.Builder
	b.WriteString(plateHead("les obligations indexées",
		"Le plancher adossé : ce qu'on achète, et ce qu'on touche"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Couple, 52 ans : plancher 40 k€, confort 55 k€, pensions à 66 ans, patrimoine 1,6 M€"))

	// The two panels, named and kept apart by their own baselines: one is a
	// stock, the other a flow, and their heights are not comparable.
	b.WriteString(sTxt(62, 86, 11, figInk, "start", "600", "Ce qu'on achète"))
	b.WriteString(sTxt(62, 99, 9.5, figMuted, "start", "400", "le patrimoine, 1,6 M€"))
	b.WriteString(sTxt(262, 86, 11, figInk, "start", "600", "Ce qu'on touche"))
	b.WriteString(sTxt(262, 99, 9.5, figMuted, "start", "400", "le revenu annuel, en k€ d'aujourd'hui"))

	// Left panel. The ladder, rung by rung from the nearest at the base, and
	// next to it the money that stays invested. A hairline of paper between
	// two rungs keeps the fourteen countable without shortening the bar.
	yCur := lBase
	for k := 1; k <= floorLadderRungs; k++ {
		h := floorLadderRungCost(k) * (lBase - lTop) / rest
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			lBar1, yCur-h, lBarW, h, floorLadderTint(k))
		if k > 1 {
			b.WriteString(line(lBar1, yCur, lBar1+lBarW, yCur, floorLadderPaper, 1))
		}
		yCur -= h
	}
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		lBar2, yL(rest), lBarW, lBase-yL(rest), figAccent)
	b.WriteString(line(58, lBase, 208, lBase, figRule, 1))

	b.WriteString(mTxt(lBar1+lBarW/2, yCur-9, 11, figInk, "middle", "600",
		fmt.Sprintf("%.0f k€", floorLadderCost())))
	b.WriteString(mTxt(lBar2+lBarW/2, yL(rest)-9, 11, figInk, "middle", "600", "1,08 M€"))
	b.WriteString(sTxt(lBar1+lBarW/2, 356, 10.5, figSoft, "middle", "600", "échelle indexée"))
	b.WriteString(sTxt(lBar1+lBarW/2, 369, 9.5, figMuted, "middle", "400", "détenue à terme"))
	b.WriteString(sTxt(lBar2+lBarW/2, 356, 10.5, figAccent, "middle", "600", "portefeuille"))
	b.WriteString(sTxt(lBar2+lBarW/2, 369, 10.5, figAccent, "middle", "600", "de confort"))
	b.WriteString(sTxt(lBar1+lBarW/2, 386, 9, figMuted, "middle", "400", "14 barreaux de 40 k€ réels"))

	// Right panel. One bar per age: the floor at the base, worn in the tint of
	// the rung that pays it until the pensions take over, and the comfort
	// layer on top, which the portfolio serves every single year.
	for i := 0; i < years; i++ {
		fill := figGreen
		if i < floorLadderRungs {
			fill = floorLadderTint(i + 1)
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			rLeft(i), yR(floorLadderFloor), barW, rBase-yR(floorLadderFloor), fill)
		b.WriteString(barV(rLeft(i), barW, yR(floorLadderFloor)-1.5,
			yR(floorLadderComfort), figAccent))
	}
	b.WriteString(line(rx0-8, rBase, rx1, rBase, figRule, 1))
	for _, t := range []struct {
		v     float64
		label string
	}{{floorLadderComfort, "confort"}, {floorLadderFloor, "plancher"}, {0, ""}} {
		b.WriteString(mTxt(rx0-14, yR(t.v)+3.5, 10, figMuted, "end", "400",
			fmt.Sprintf("%.0f", t.v)))
		if t.label != "" {
			b.WriteString(sTxt(rx0-14, yR(t.v)+16, 9.5, figMuted, "end", "400", t.label))
		}
	}
	for _, age := range []int{52, 56, 60, 64, 68, 72, 75} {
		b.WriteString(mTxt(rMid(age-floorLadderAge), 356, 9.5, figMuted, "middle", "400",
			fmt.Sprintf("%d", age)))
	}
	b.WriteString(sTxt((rx0+rx1)/2, 380, 10, figMuted, "middle", "400", "âge  →"))

	// The hand-over, and the two readings the plate exists for.
	xr := rx0 + slot*float64(floorLadderRungs)
	b.WriteString(dashLine(xr, 136, xr, 344, figGreen, 1.4, "4 4"))
	b.WriteString(sTxt(xr+7, 118, 10.5, figGreen, "start", "600", "les pensions"))
	b.WriteString(sTxt(xr+7, 131, 10.5, figGreen, "start", "600", "prennent le relais"))
	b.WriteString(sTxt(rx0+4, 118, 10.5, figAccent, "start", "600", "le confort, 15 k€ par an"))
	b.WriteString(sTxt(rx0+4, 131, 10, figSoft, "start", "400",
		fmt.Sprintf("15/1080 = %s %% du portefeuille", frNum(floorLadderRate(), 1))))

	// The legend: the tint ramp itself, which is the only key the plate needs.
	for k := 1; k <= floorLadderRungs; k++ {
		fmt.Fprintf(&b, `<rect x="%.1f" y="402" width="7" height="9" fill="%s"/>`,
			24+7*float64(k-1), floorLadderTint(k))
	}
	b.WriteString(sTxt(132, 410, 10, figSoft, "start", "400",
		"une teinte par barreau : à gauche ce qu'il coûte, à droite l'année qu'il paie"))

	b.WriteString(sTxt(24, 432, 9.5, figMuted, "start", "400", fmt.Sprintf(
		"Montants en euros d'aujourd'hui. Les 14 barreaux versent 40 k€ chacun et coûtent %.0f k€ au taux réel de 1 %%.",
		floorLadderCost())))
	b.WriteString(sTxt(24, 447, 9.5, figMuted, "start", "400",
		"Le portefeuille de confort n'est pas garanti ; le plancher adossé l'est, jusqu'aux pensions."))
	return svg(640, 462, b.String())
}
