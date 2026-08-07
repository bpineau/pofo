package firebook

import (
	"fmt"
	"strings"
)

// The plates that rank countries. Their common raw material is the bundled
// Jorda-Schularick-Taylor panel of per-country REAL annual total returns.

// safemaxCountry is one country's SAFEMAX: the highest initial withdrawal rate
// (percent of the starting capital, then held constant in real terms) that a
// domestic 60/40 would have carried through EVERY complete 30-year window the
// panel offers for that country.
type safemaxCountry struct {
	ISO     string
	Label   string
	Rate    float64 // percent of the starting capital
	Worst   int     // first year of the window that sets the rate
	Windows int     // complete 30-year windows the panel offers
	Partial bool    // the panel misses the country's worst years (see safemaxGaps)
}

// safemaxCountries freezes the ranking the plate draws.
//
// Source: pkg/datasets.BroadSample(), the bundled Jorda-Schularick-Taylor
// Macrohistory R6 panel (16 economies, 1871-2020, columns equity / bond / bill
// as REAL annual total returns, each country already deflated by its own CPI).
//
// Recipe, per country:
//
//  1. build the domestic 60/40 real annual series r(y) = 0.6*equity + 0.4*bond,
//     keeping only the years where BOTH legs are present (yearly rebalancing,
//     no fee, no tax);
//  2. take every 30-year window whose thirty years are all present (a window
//     with a hole is dropped, never interpolated), Windows counts them;
//  3. for one window, the highest initial rate that survives it is the exact
//     annuity rate 1 / sum(t = 0..29) prod(i = 1..t) 1/(1+r_i): the withdrawal
//     happens at the START of each year, is held constant in real terms (the
//     returns are already real, so no indexation is left to apply), and the
//     balance is allowed to reach exactly zero after the thirtieth one;
//  4. Rate is the MINIMUM of those window rates, Worst the window that sets it.
//
// safemaxWorld applies steps 2 to 4 to the equal-weight basket: the plain
// average, year by year, of the domestic 60/40 returns available that year.
//
// figures_pays_test.go recomputes every field from pkg/datasets and fails the
// moment the plate and the panel disagree.
var safemaxCountries = []safemaxCountry{
	{"DNK", "Danemark", 4.0029, 1917, 89, false},
	{"USA", "États-Unis", 3.7541, 1966, 120, false},
	{"AUS", "Australie", 3.3325, 1970, 92, false},
	{"JPN", "Japon", 3.2730, 1990, 75, true},
	{"NLD", "Pays-Bas", 3.2415, 1910, 92, false},
	{"CHE", "Suisse", 3.1887, 1962, 76, false},
	{"GBR", "Royaume-Uni", 2.9270, 1937, 121, false},
	{"NOR", "Norvège", 2.6074, 1912, 111, false},
	{"BEL", "Belgique", 2.4370, 1937, 86, false},
	{"SWE", "Suède", 2.3198, 1913, 121, false},
	{"ESP", "Espagne", 1.7372, 1974, 59, false},
	{"FIN", "Finlande", 1.4369, 1912, 96, false},
	{"ITA", "Italie", 1.0613, 1942, 121, false},
	{"DEU", "Allemagne", 0.4929, 1914, 87, true},
	{"PRT", "Portugal", 0.4277, 1974, 121, false},
	{"FRA", "France", 0.3225, 1943, 121, false},
}

// safemaxGaps are the two holes the plate marks with a star. They matter more
// than the others because they cover exactly the catastrophe: the Tokyo market
// was closed in 1946-1947 and Germany's collapse and 1948 currency reform sit
// in 1944-1948. Every 30-year window containing those years is dropped, so the
// two bars rest on the country's calmer history and are drawn hollow.
var safemaxGaps = map[string][]int{
	"JPN": {1946, 1947},
	"DEU": {1944, 1945, 1946, 1947, 1948},
}

// safemaxWorld is the same computation on an equal-weight basket of the panel's
// countries, and safemaxWorldWorst / safemaxWorldWindows its worst window and
// window count. It is a crude world: an average of national experiences, each
// measured in its own real terms, with no currency effect and no market-cap
// weighting, and the years a country is missing are missing from the basket too.
const (
	safemaxWorld        = 3.2783
	safemaxWorldWorst   = 1911
	safemaxWorldWindows = 121
)

// safemaxFranceEquity{Full,Post50} are the French real equity geometric means
// the last footnote quotes, over 1900-2020 and 1950-2020 (percent per year,
// same panel). They are there because the panel is much harsher on pre-1950
// France than the Dimson-Marsh-Staunton numbers the book quotes elsewhere, and
// a plate that puts France last must say where that comes from.
const (
	safemaxFranceEquityFull   = -0.152
	safemaxFranceEquityPost50 = 3.437
)

// figSafemaxPays ranks the sixteen countries of the panel by the withdrawal
// rate their own history would have carried, with the 4 % rule as a vertical
// reference and the equal-weight world basket as a second, different one. The
// message is the spread: the 4 % of the American tradition is one country's
// number (and even that country lands below it on this panel), the median
// developed country sits near 2,5 %, and a third of the sample never carried
// 1,5 %. France and the United States are coloured so a French reader places
// themselves at once.
func figSafemaxPays() string {
	const (
		plotL, plotR = 118.0, 500.0
		axisMax      = 4.5
		rowTop       = 88.0
		pitch        = 17.5
		barH0        = 11.0
		colWorst     = 560.0
		colWindows   = 616.0
	)
	x := func(v float64) float64 { return plotL + v/axisMax*(plotR-plotL) }
	axisY := rowTop + pitch*float64(len(safemaxCountries)) + 4

	var b strings.Builder
	b.WriteString(plateHead("seize pays, 1870-2020", "Le taux de retrait sûr est d'abord une affaire de pays"))
	b.WriteString(sTxt(24, 60, 10.5, figMuted, "start", "400",
		"Taux initial d'un 60/40 domestique, tenu 30 ans dans le pire millésime du pays"))
	b.WriteString(sTxt(colWorst, 80, 9, figMuted, "end", "600", "pire millésime"))
	b.WriteString(sTxt(colWindows, 80, 9, figMuted, "end", "600", "fenêtres"))

	// vertical grid behind the bars, the zero as a solid rule
	for _, g := range []float64{0, 1, 2, 3, 4} {
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(x(g), rowTop-4, x(g), axisY, col, 1))
		b.WriteString(mTxt(x(g), axisY+15, 10, figMuted, "middle", "400", frNum(g, 0)+" %"))
	}
	b.WriteString(line(plotL, axisY, plotR, axisY, figRule, 1))

	// the two references, drawn as rules rather than bars: the 4 % of the
	// tradition, dashed and red, labelled under the axis; the world basket,
	// solid and green, labelled above the plot with a marker at its head
	b.WriteString(dashLine(x(4), rowTop-4, x(4), axisY, figBad, 1.4, "4 4"))
	b.WriteString(sTxt(x(4), axisY+31, 10.5, figBad, "middle", "600", "la règle des 4 %"))
	b.WriteString(line(x(safemaxWorld), rowTop-9, x(safemaxWorld), axisY, figGreen, 1.4))
	fmt.Fprintf(&b, `<path d="M %.1f,%.1f l -4.5,-7 l 9,0 Z" fill="%s"/>`,
		x(safemaxWorld), rowTop-8, figGreen)
	b.WriteString(sTxt(x(safemaxWorld)-9, 77, 10.5, figGreen, "end", "600",
		"panier mondial équipondéré · "+frNum(safemaxWorld, 1)+" %"))

	for i, c := range safemaxCountries {
		y := rowTop + pitch*float64(i)
		base := y + barH0 - 2
		fill, ink, weight := figMuted, figSoft, "400"
		switch c.ISO {
		case "FRA":
			fill, ink, weight = figAccent, figDeep, "700"
		case "USA":
			fill, ink, weight = figBlue, figBlue, "700"
		}
		star := ""
		if c.Partial {
			star = "*"
		}
		b.WriteString(sTxt(plotL-8, base, 11, ink, "end", weight, c.Label+star))
		if c.Partial {
			// hollow: the panel cannot see this country's worst window, so the
			// bar is longer than the truth
			b.WriteString(barHOutline(x(0), x(c.Rate), y, barH0, fill))
		} else {
			b.WriteString(barH(x(0), x(c.Rate), y, barH0, fill))
		}
		b.WriteString(mTxt(x(c.Rate)+8, base, 10.5, figInk, "start", "600", frNum(c.Rate, 2)+" %"))
		b.WriteString(mTxt(colWorst, base, 9.5, figMuted, "end", "400", fmt.Sprint(c.Worst)))
		b.WriteString(mTxt(colWindows, base, 9.5, figMuted, "end", "400", fmt.Sprint(c.Windows)))
	}

	notes := []string{
		"Panel Jorda-Schularick-Taylor : 60 % actions / 40 % obligations du pays, réels, rééquilibrés, prélèvement en début d'année.",
		"Seules comptent les fenêtres de 30 ans complètes, d'où des comptes très inégaux d'un pays à l'autre.",
		"* Japon et Allemagne : 1946-1947 et 1944-1948 manquent au panel, et leurs pires fenêtres avec. Ces deux barres sont trop longues.",
		"Pfau et Anarkulova-Cederburg calculent sur d'autres bases : leur SAFEMAX américain ressort vers 4 %, ici 3,75 %.",
		"Le panel est très dur avec la France d'avant 1950 : " + frSignedNum(safemaxFranceEquityFull, 1) +
			" %/an réel en actions sur 1900-2020, mais " + frSignedNum(safemaxFranceEquityPost50, 1) + " %/an sur 1950-2020.",
	}
	for i, s := range notes {
		b.WriteString(sTxt(24, axisY+61+float64(i)*14, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, int(axisY+61+float64(len(notes)-1)*14+16), b.String())
}

// barHOutline draws barH's shape hollow: the card background inside, the colour
// as a hairline outline. Used for a bar whose value the data cannot vouch for.
func barHOutline(x0, x1, y, h float64, stroke string) string {
	r := 3.5
	if x1-x0 < r*2 {
		r = (x1 - x0) / 2
	}
	return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Z" `+
		`fill="#fffdf9" stroke="%s" stroke-width="1.3"/>`,
		x0, y, x1-r, y, x1, y, x1, y+r, x1, y+h-r, x1, y+h, x1-r, y+h, x0, y+h, stroke)
}
