package firebook

import (
	"fmt"
	"strings"
)

// errCost is one classic FIRE construction error, measured: the single entry it
// changes in a shared reference plan, and the ruin probability (in percent) of
// that plan once the error is committed.
//
// Inverse marks the one error whose bar does not describe a lived loss. A
// household that forgets its pension does not become poorer, it works longer:
// the bar is the ruin of the plan it believes it holds, and the bill is paid in
// working years. It is drawn apart for that reason.
type errCost struct {
	label, change string
	ruin          float64
	inverse       bool
}

// The reference plan and the six variants below were produced by this
// repository's own engine, decumul.Plan over the empirical broad-sample source
// (scenario.PooledBootstrap of the bundled Jorda-Schularick-Taylor per-country
// real returns blended 60/40 within each country, mean block 10 years),
// 200 000 paths, seed 1, 8 workers (the draw is split per worker, so the worker
// count is part of the reproduction recipe), annual kernel:
//
//	plan: Capital 1e6, NeedAnnual 35000 (3.5 %), Years 50,
//	      Cashflows {FromYear: 20, Annual: 14000},
//	      Flex {Threshold: 0.20, Cut: 0.10}, no tax, no buffer.
//
// Each variant changes exactly ONE field: NeedAnnual (42000 for spending
// underestimated by 20 %, 40000 for the 4 % rate, 39200 for the 12 % tax and
// PUMa friction the article puts at 8-20 %, 45000 for the rate a 40 % cut would
// justify at the reference ruin, solved with decumul.WithdrawalAxis), Cashflows
// (nil for the forgotten pension) or Source (equity weight 0 for the all-bond
// portfolio). TestErreursPlateMatchesTheEngine re-runs all of them.
//
// Sampling noise at these path counts is +-0.17 point at 95 %, and the ranking
// is unchanged across seeds; the two smallest gaps (+10.9 vs +9.5 and +7.4 vs
// +6.1) are an order of magnitude wider than that.
const errCostRef = 19.34 // ruin (%) of the clean reference plan

// errCostEquity/errCostCutEquity/errCostCutRef are the all-equity mirror of the
// mono-regime error: on the ruin axis it does not show (it is slightly BELOW
// the reference), it shows in the number of years lived under a cut budget.
const (
	errCostEquity    = 18.31 // ruin (%) of the same plan run 100 % equities
	errCostCutRef    = 16.75 // mean years with a cut budget, reference 60/40
	errCostCutEquity = 20.04 // mean years with a cut budget, 100 % equities
)

// errCosts is the measured ranking, most expensive first.
var errCosts = []errCost{
	{"Portefeuille mono-régime, tout obligataire", "0 % d'actions au lieu de 60 %", 39.57, false},
	{"Flexibilité surestimée", "4,5 % au lieu de 3,5 %, le taux qu'une coupe de 40 % justifierait", 35.95, false},
	{"Dépenses sous-estimées de 20 %", "42 k€/an au lieu de 35 k€", 30.25, false},
	{"Pension oubliée", "aucune pension comptée dans le plan", 28.82, true},
	{"Taux de 4 % au lieu de 3,5 %", "40 k€/an au lieu de 35 k€", 26.74, false},
	{"Fiscalité et PUMa ignorées", "12 % de friction non budgétée, soit 39,2 k€ à sortir", 25.44, false},
}

// figCoutDesErreurs ranks the six errors of the article that can be measured
// without inventing data, on one shared reference plan and one shared model.
// The axis is the plan's ruin probability, the dashed rule is the clean plan,
// and each bar's amber (or blue) part is what the error adds to it. The order
// the page announces is not the order the engine finds, which is the point of
// putting the plate at the top of the article.
func figCoutDesErreurs() string {
	x := func(v float64) float64 { return 24 + v/45*(596-24) }
	const (
		top   = 90.0 // first row's label baseline
		pitch = 52.0
		barH0 = 15.0
		axisY = 395.0
	)
	barTop := func(i int) float64 { return top + pitch*float64(i) + 20 }

	var b strings.Builder
	b.WriteString(plateHead("erreurs classiques", "Six erreurs, un même plan, un coût mesuré"))

	// vertical grid, behind everything
	for _, g := range []float64{0, 10, 20, 30, 40} {
		b.WriteString(line(x(g), 72, x(g), axisY, figGrid, 1))
		b.WriteString(mTxt(x(g), 409, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(line(24, axisY, 596, axisY, figRule, 1))
	b.WriteString(sTxt(24, 64, 10.5, figMuted, "start", "400", "probabilité de ruine du plan (%)"))

	// the clean plan: the spine every bar is read against
	b.WriteString(dashLine(x(errCostRef), 72, x(errCostRef), axisY, figDeep, 1.2, "3 3"))
	b.WriteString(sTxt(x(errCostRef)+7, 64, 10.5, figDeep, "start", "600",
		"le plan propre : "+frNum(errCostRef, 1)+" % de ruine"))

	for i, e := range errCosts {
		y := barTop(i)
		fill, ink := figAccent, figDeep
		if e.inverse {
			fill, ink = figBlue, figBlue
		}
		b.WriteString(sTxt(24, top+pitch*float64(i), 11, figInk, "start", "600", e.label))
		b.WriteString(sTxt(24, top+pitch*float64(i)+13, 10, figMuted, "start", "400", e.change))
		// the reference share of the bar, then the cost the error adds
		fmt.Fprintf(&b, `<rect x="24" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			y, x(errCostRef)-24, barH0, figWash)
		b.WriteString(barH(x(errCostRef), x(e.ruin), y, barH0, fill))
		b.WriteString(mTxt(x(e.ruin)+8, y+11.5, 10.5, ink, "start", "600",
			"+"+frNum(e.ruin-errCostRef, 1)+" pts"))
		if e.inverse {
			b.WriteString(sTxt(x(e.ruin)+66, y+11.5, 9.5, figBlue, "start", "400", "payée en années de travail"))
		}
	}

	for i, s := range []string{
		"Plan de référence : 1 M€, 35 k€/an réels (3,5 %), 50 ans, 60/40, pension de 14 k€/an à l'année 20, coupe tenue de 10 %.",
		"Modèle empirique : 16 pays, 1870-2020, blocs de 10 ans, 200 000 tirages. Une seule entrée change par barre.",
		"Le tout-actions, lui, ne se voit pas sur cet axe (18,3 %) : il coûte 20 années de budget coupé sur 50, contre 16,8.",
	} {
		b.WriteString(sTxt(24, 431+float64(i)*14, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 470, b.String())
}
