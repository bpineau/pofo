package firebook

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// The plates of the inflation part. They answer the questions a French rentier
// asks of the last century, on French data, in real terms.

// frSignedNum formats v French style with dec decimals: U+2212 as the minus
// sign, a comma decimal separator, a plain space every three integer digits,
// and an explicit plus sign on gains.
func frSignedNum(v float64, dec int) string {
	s := strconv.FormatFloat(math.Abs(v), 'f', dec, 64)
	whole, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		whole, frac = s[:i], s[i+1:]
	}
	var b strings.Builder
	if v < 0 {
		b.WriteString("−")
	} else {
		b.WriteString("+")
	}
	for i := 0; i < len(whole); i++ {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteByte(whole[i])
	}
	if frac != "" {
		b.WriteString("," + frac)
	}
	return b.String()
}

// A victimeRow is one brick of a rentier's portfolio over one regime: Mult is
// the REAL capital left at the end of the window for 1 EUR invested at its
// start (1 = purchasing power exactly preserved), Years the number of annual
// returns actually compounded to get there.
type victimeRow struct {
	Key   string // the panel column: equity, bond or bill
	Label string
	Color string
	Mult  float64
	Years int
}

// A victimeRegime is one of the article's French inflation regimes, from year
// From to year To inclusive.
type victimeRegime struct {
	Name     string
	From, To int
	Rows     [3]victimeRow
}

// victimesRegimes freezes the twelve multiples the plate draws.
//
// Source: pkg/datasets.BroadSample(), the bundled Jorda-Schularick-Taylor
// Macrohistory R6 panel, rows iso = FRA, columns equity / bond / bill. Those
// are annual REAL total returns, each year already deflated by the French CPI,
// so the multiple is the plain product of (1 + r) over the window, both
// endpoints included, and no currency or inflation adjustment is left to do.
// Compounding convention: yearly, arithmetic order irrelevant, missing years
// skipped rather than interpolated (Years then falls below the window length).
//
// The four windows are the article's own regimes, which is why the first two
// overlap over 1950-1958: the destructions run to the new franc of 1958 while
// the organized inflation of the Trente Glorieuses starts in 1950.
//
// One gap: the panel carries no French short rate from 1915 to 1921, so the
// monetary leg of 1914-1958 compounds 38 years instead of 45 and misses the
// worst of the Great War. The plate marks that row and says so.
//
// figures_inflation_test.go recomputes every multiple and every year count
// from pkg/datasets and fails the moment the plate and the panel disagree.
var victimesRegimes = []victimeRegime{
	{"Les grandes destructions", 1914, 1958, [3]victimeRow{
		{"equity", "Actions", figAccent, 0.109241, 45},
		{"bond", "Obligations", figBad, 0.024588, 45},
		{"bill", "Monétaire", figBlue, 0.048094, 38},
	}},
	{"L'inflation organisée", 1950, 1970, [3]victimeRow{
		{"equity", "Actions", figAccent, 1.132278, 21},
		{"bond", "Obligations", figBad, 0.706855, 21},
		{"bill", "Monétaire", figBlue, 0.885772, 21},
	}},
	{"La grande inflation", 1973, 1985, [3]victimeRow{
		{"equity", "Actions", figAccent, 0.907719, 13},
		{"bond", "Obligations", figBad, 0.884655, 13},
		{"bill", "Monétaire", figBlue, 1.065897, 13},
	}},
	{"La désinflation", 1985, 2020, [3]victimeRow{
		{"equity", "Actions", figAccent, 14.066339, 36},
		{"bond", "Obligations", figBad, 10.061203, 36},
		{"bill", "Monétaire", figBlue, 1.913970, 36},
	}},
}

// victimeBandHead sets a regime band header: the regime's name in the sans, its
// window and length trailing in the mono.
func victimeBandHead(y float64, name, span string) string {
	return sTxt(24, y, 11.5, figSoft, "start", "600", name+
		` <tspan font-family="`+figMono+`" font-size="10" font-weight="400" fill="`+figMuted+`">`+span+`</tspan>`)
}

// figVictimesRegimes draws the hierarchy of the victims, regime by regime, on
// French data. The hard part is the scale: +1 307 % has to sit next to −97,5 %
// without flattening the two regimes where nothing much happened. The answer
// is to plot what a cumulative return really is, the final capital for 1 EUR
// invested, on a LOGARITHMIC axis, where a loss of 97,5 % (x 1/41) and a gain
// of 1 307 % (x 14) are of comparable size and where the near-flat regimes stay
// legible as dots hugging the x 1 line. Position carries the shape, the two
// mono columns carry the exact numbers.
func figVictimesRegimes() string {
	const (
		plotL, plotR = 112.0, 492.0
		lo, hi       = -1.75, 1.30 // decades, x 1/56 to x 20
		top, bottom  = 78.0, 368.0
		colCum       = 558.0
		colYear      = 620.0
	)
	x := func(mult float64) float64 {
		return plotL + (math.Log10(mult)-lo)/(hi-lo)*(plotR-plotL)
	}
	var b strings.Builder
	b.WriteString(plateHead("un siècle français", "La hiérarchie des victimes, régime par régime"))
	b.WriteString(sTxt(24, 64, 10.5, figMuted, "start", "400",
		"Ce que devient 1 € réel placé pendant tout le régime (échelle log : un cran = ×10)"))
	b.WriteString(sTxt(colCum, 64, 9.5, figMuted, "end", "600", "cumulé"))
	b.WriteString(sTxt(colYear, 64, 9.5, figMuted, "end", "600", "par an"))

	// decade gridlines, dashed, plus the x 1 line as a solid rule since it is
	// the frontier between a capital preserved and a capital eaten
	for _, g := range []float64{0.1, 10} {
		b.WriteString(dashLine(x(g), top, x(g), bottom, figRule, 1, "2 4"))
	}
	b.WriteString(line(x(1), top, x(1), bottom, figRule, 1.2))
	b.WriteString(mTxt(x(0.1), 384, 10, figMuted, "middle", "400", "×1/10"))
	b.WriteString(mTxt(x(1), 384, 10, figMuted, "middle", "400", "×1"))
	b.WriteString(sTxt(x(1)+9, 384, 9.5, figMuted, "start", "400", "(capital intact)"))
	b.WriteString(mTxt(x(10), 384, 10, figMuted, "middle", "400", "×10"))

	for i, reg := range victimesRegimes {
		bandTop := 84 + 72*float64(i)
		span := fmt.Sprintf("%d-%d · %d ans", reg.From, reg.To, reg.To-reg.From+1)
		b.WriteString(victimeBandHead(bandTop+11, reg.Name, span))
		for j, r := range reg.Rows {
			y := bandTop + 26 + 16*float64(j)
			partial := r.Years < reg.To-reg.From+1
			star := ""
			if partial {
				star = "*"
			}
			b.WriteString(sTxt(104, y+3.5, 11, figSoft, "end", "400", r.Label))
			// the stem measures the decades won or lost, the dot marks the value;
			// the row the panel cannot fully compute gets a dashed stem and a
			// hollow dot
			if partial {
				b.WriteString(dashLine(x(1), y, x(r.Mult), y, r.Color, 1.8, "2 3"))
			} else {
				b.WriteString(line(x(1), y, x(r.Mult), y, r.Color, 1.8))
			}
			if partial {
				fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="#fffdf9" stroke="%s" stroke-width="1.8"/>`,
					x(r.Mult), y, r.Color)
			} else {
				fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, x(r.Mult), y, r.Color)
			}
			cum := (r.Mult - 1) * 100
			dec := 1
			if math.Abs(cum) >= 100 {
				dec = 0
			}
			b.WriteString(mTxt(colCum, y+3.5, 10.5, figInk, "end", "600", frSignedNum(cum, dec)+" %"+star))
			perYear := (math.Pow(r.Mult, 1/float64(r.Years)) - 1) * 100
			b.WriteString(mTxt(colYear, y+3.5, 10.5, figMuted, "end", "400", frSignedNum(perYear, 1)+" %"))
		}
	}

	b.WriteString(sTxt(24, 406, 11, figSoft, "start", "600",
		"Dans les quatre régimes, l'action finit devant l'obligation nominale, sans être épargnée pour autant."))
	b.WriteString(sTxt(24, 424, 9.5, figMuted, "start", "400",
		"Rendements réels annuels français du panel Jorda-Schularick-Taylor (R6), capitalisés année par année."))
	b.WriteString(sTxt(24, 438, 9.5, figMuted, "start", "400",
		"* Monétaire 1914-1958 : sept années manquent au panel (1915-1921, les pires). La vraie destruction est plus forte."))
	return svg(640, 452, b.String())
}
