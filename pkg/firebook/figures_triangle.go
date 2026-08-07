package firebook

import (
	"fmt"
	"strings"
)

// The correlation triangle of the diversification article. Its subject is the
// article's blunt claim, "trente fonds actions ne font qu'un seul actif": a
// half matrix over seven series, four equity funds first, then the three
// non-equity bricks, so the saturated block of equity pairs and the pale zone
// of everything else are read in one look, without a single number being
// needed to make the point.
//
// Form: a lower triangle, no decorative diagonal (the diagonal cell carries
// the column's index, the head of its own column), horizontal labels on both
// axes, one number per cell in the UI mono, and a diverging ramp of SOLID hex
// steps, never rgba.

// triSeries is one plotted series: the bundled identifier it is measured on,
// the French label of its row, and whether it is one of the four bricks the
// article opposes to the pile of equity funds (the label of a redundant equity
// fund is set in muted grey, the bricks in ink: the grouping needs no legend).
type triSeries struct {
	id    string // bundled series identifier
	ref   string // "" for pkg/datasets/simdata, "refdata" for pkg/datasets/refdata
	label string
	brick bool
}

// The seven rows, equity funds first so their block lands top left.
//
//	MSCIWORLD     MSCI World net total return (index, fee-free)
//	SP500         S&P 500 total return
//	VTI           the whole US market (Vanguard Total Stock Market)
//	DEVEXUS-DAILY developed markets ex-US, Ken French's daily market return
//	TLT           Treasuries 20 years and over
//	XAUUSD        gold spot, USD
//	CTA           a diversified trend programme, TSMOM reconstruction
//
// Every series is measured in USD, so no currency effect enters a correlation.
var triRows = []triSeries{
	{id: "MSCIWORLD", label: "fonds monde", brick: true},
	{id: "SP500", label: "fonds S&amp;P 500"},
	{id: "VTI", label: "fonds US toutes tailles"},
	{id: "DEVEXUS-DAILY", ref: "refdata", label: "fonds hors US"},
	{id: "TLT", label: "obligations longues", brick: true},
	{id: "XAUUSD", label: "or", brick: true},
	{id: "CTA", label: "trend", brick: true},
}

// triCorr is the lower triangle of the correlation matrix of the MONTHLY
// returns of those seven series, over ONE common window: the month ends every
// series quotes, from December 2000 to May 2026, that is 305 monthly returns
// from January 2001 to May 2026. Row i holds the correlations of series i with
// series 0 to i-1.
//
// Reproduce: read each identifier with marketdata.ReadSimdataFS on
// datasets.Simdata() (datasets.Refdata() for DEVEXUS-DAILY); keep the days all
// seven quote; keep the last such day of each month, from December 2000 on;
// take simple returns between consecutive month ends; then the Pearson
// correlation of every pair. figures_triangle_test.go recomputes the whole
// triangle from pkg/datasets and fails if the plate and the data disagree.
//
// The window starts in January 2001 because that is where the trend leg's
// reconstruction is calibrated (the same start as the rolling-correlation
// plate of the managed-futures article), and THE TREND LEG IS A
// RECONSTRUCTION before 2022, not a live track record: the plate says so.
//
// Two cells are worth reading exactly, since the plate rounds to two decimals:
// S&P 500 against the whole US market is 0,9955, and MSCI World against the
// S&P 500 is 0,9736.
var triCorr = [][]float64{
	{},
	{0.97},
	{0.97, 1.00},
	{0.94, 0.85, 0.86},
	{-0.12, -0.13, -0.13, -0.09},
	{0.13, 0.06, 0.07, 0.23, 0.23},
	{-0.03, -0.06, -0.06, 0.01, 0.15, 0.22},
}

// triEquityRows is the number of leading rows that are equity funds: the block
// the plate is about. triBrickRows lists the four bricks the article opposes
// to it, by row index (the world fund is both, which is the whole point).
const triEquityRows = 4

var triBrickRows = []int{0, 4, 5, 6}

// The two ranges the plate prints, recomputed from the triangle itself so the
// captions can never drift from the cells.
func triRange(pairs [][2]int) (lo, hi float64) {
	lo, hi = 1, -1
	for _, p := range pairs {
		v := triCorr[p[0]][p[1]]
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return lo, hi
}

// triEquityPairs are the six pairs inside the equity block, triBrickPairs the
// six pairs among the four bricks.
func triEquityPairs() [][2]int {
	var out [][2]int
	for i := 1; i < triEquityRows; i++ {
		for j := 0; j < i; j++ {
			out = append(out, [2]int{i, j})
		}
	}
	return out
}

func triBrickPairs() [][2]int {
	var out [][2]int
	for a := 1; a < len(triBrickRows); a++ {
		for b := 0; b < a; b++ {
			out = append(out, [2]int{triBrickRows[a], triBrickRows[b]})
		}
	}
	return out
}

// triSteps is the diverging ramp, eight SOLID hex steps and never an rgba fill
// or an opacity (crengine, KOReader's EPUB renderer, paints those solid
// black). Each step is figAccent (180,120,60) on the positive side and figBlue
// (58,109,180) on the negative side, composited onto the figure card
// background #fffdf9 (255,253,249) at alpha a, channel by channel:
// c = bg + a×(colour − bg), rounded.
//
//	positive  0.10 -> #F7F0E6  0.25 -> #ECDCCA  0.45 -> #DDC1A4
//	          0.70 -> #CBA074  1.00 -> #B4783C (= figAccent)
//	negative  0.15 -> #E2E7EF  0.35 -> #BACBE1  0.60 -> #89A7D0
//
// upTo is the step's exclusive upper bound in correlation.
var triSteps = []struct {
	upTo float64
	fill string
}{
	{-0.4, "#89A7D0"},
	{-0.2, "#BACBE1"},
	{0.0, "#E2E7EF"},
	{0.2, "#F7F0E6"},
	{0.4, "#ECDCCA"},
	{0.6, "#DDC1A4"},
	{0.8, "#CBA074"},
	{2.0, figAccent},
}

func triFill(v float64) string {
	for _, s := range triSteps {
		if v < s.upTo {
			return s.fill
		}
	}
	return figAccent
}

// triCellInk keeps the printed number readable on the two darkest steps.
func triCellInk(fill string) string {
	if fill == figAccent || fill == "#CBA074" {
		return "#fffdf9"
	}
	return figInk
}

// triNum prints a correlation the French way, with a comma and a typographic
// minus sign (U+2212).
func triNum(v float64) string {
	if v < 0 {
		return "−" + frNum(-v, 2)
	}
	return frNum(v, 2)
}

// triSigned prints a correlation with its sign always shown, for the two
// ranges the plate states in words.
func triSigned(v float64) string {
	if v < 0 {
		return triNum(v)
	}
	return "+" + triNum(v)
}

// Plate geometry: seven rows of the same height, seven columns of the same
// width, labels to the left of the grid and the column index on the diagonal.
const (
	triGridL  = 208.0
	triGridT  = 96.0
	triCellW  = 52.0
	triCellH  = 32.0
	triLabelR = 196.0
)

// figTriangleCorrelations draws the half matrix. The four equity funds come
// first, so their pairs fill the top-left staircase in the darkest step of the
// ramp, while every pair that mixes families stays nearly blank.
func figTriangleCorrelations() string {
	n := len(triRows)
	cellX := func(j int) float64 { return triGridL + float64(j)*triCellW }
	cellY := func(i int) float64 { return triGridT + float64(i)*triCellH }

	var b strings.Builder
	b.WriteString(plateHead("diversification", "Quatre fonds actions ne font qu'un seul actif"))
	b.WriteString(sTxt(24, 64, 10.2, figMuted, "start", "400",
		"corrélation des rendements mensuels, janvier 2001 à mai 2026, une seule fenêtre commune de 305 mois"))

	// Rows: the name, then the index right against the grid, where it heads the
	// column of the same number. The three redundant equity funds are set in
	// muted grey, the four bricks in ink: the grouping needs no legend.
	for i, s := range triRows {
		y := cellY(i) + 21
		col, weight := figMuted, "400"
		if s.brick {
			col, weight = figSoft, "600"
		}
		b.WriteString(sTxt(triLabelR-18, y, 10.5, col, "end", weight, s.label))
		b.WriteString(mTxt(triLabelR, y, 10, figMuted, "end", "400", fmt.Sprintf("%d", i+1)))
	}

	// The cells of the lower triangle, each inset by a hairline and outlined in
	// the grid colour, so the columns stay followable across the pale zone.
	for i := 1; i < n; i++ {
		for j := 0; j < i; j++ {
			v := triCorr[i][j]
			fill := triFill(v)
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1"/>`,
				cellX(j)+0.8, cellY(i)+0.8, triCellW-1.6, triCellH-1.6, fill, figGrid)
			b.WriteString(mTxt(cellX(j)+triCellW/2, cellY(i)+triCellH/2+3.6, 10.5,
				triCellInk(fill), "middle", "600", triNum(v)))
		}
		// The diagonal cell carries no correlation, only the index of the
		// column whose cells start just below it.
		b.WriteString(mTxt(cellX(i)+triCellW/2, cellY(i)+triCellH/2+3.6, 10, figMuted, "middle", "400",
			fmt.Sprintf("%d", i+1)))
	}
	b.WriteString(mTxt(cellX(0)+triCellW/2, cellY(0)+triCellH/2+3.6, 10, figMuted, "middle", "400", "1"))

	// The equity block, outlined as the staircase it is, two pixels clear of
	// the cells so the frame reads as a frame.
	const pad = 2.5
	var stair [][2]float64
	for i := 1; i < triEquityRows; i++ {
		stair = append(stair,
			[2]float64{cellX(i-1) - pad, cellY(i) - pad},
			[2]float64{cellX(i) + pad, cellY(i) - pad})
	}
	stair = append(stair,
		[2]float64{cellX(triEquityRows-1) + pad, cellY(triEquityRows) + pad},
		[2]float64{cellX(0) - pad, cellY(triEquityRows) + pad},
		[2]float64{cellX(0) - pad, cellY(1) - pad})
	b.WriteString(poly(stair, figDeep, 1.3, ""))

	// The two readings, in the empty half of the matrix.
	eqLo, eqHi := triRange(triEquityPairs())
	brLo, brHi := triRange(triBrickPairs())
	for _, l := range []struct {
		y     float64
		mono  bool
		color string
		s     string
	}{
		{146, false, figSoft, "les six paires de fonds actions"},
		{163, true, figDeep, triSigned(eqLo) + " à " + triSigned(eqHi)},
		{194, false, figSoft, "les six paires des quatre briques"},
		{211, true, figDeep, triSigned(brLo) + " à " + triSigned(brHi)},
	} {
		if l.mono {
			b.WriteString(mTxt(420, l.y, 11.5, l.color, "start", "600", l.s))
			continue
		}
		b.WriteString(sTxt(420, l.y, 10.5, l.color, "start", "600", l.s))
	}

	// The key: the eight steps of the ramp, with the boundaries under them.
	const keyY, keyW, keyH = 346.0, 36.0, 13.0
	b.WriteString(sTxt(triLabelR, keyY+10, 10, figMuted, "end", "400", "corrélation"))
	for i, s := range triSteps {
		x := triGridL + float64(i)*keyW
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, keyY, keyW, keyH, s.fill)
		if i == len(triSteps)-1 {
			continue
		}
		lab := frNum(s.upTo, 1)
		switch {
		case s.upTo == 0:
			lab = "0"
		case s.upTo < 0:
			lab = "−" + frNum(-s.upTo, 1)
		}
		b.WriteString(mTxt(x+keyW, keyY+keyH+12, 9.5, figMuted, "middle", "400", lab))
	}

	for i, s := range []string{
		"Sept séries en dollars : MSCI World, S&amp;P 500, marché américain toutes tailles, développés hors US (Ken French),",
		"Treasuries 20 ans et plus, or au comptant, programme de suivi de tendance (reconstruction avant 2022).",
		"Prendre des actions vraiment différentes n'y change rien : le small value américain reste à 0,79 du fonds monde.",
		"Ces corrélations sont des moyennes de période : par tranches de dix ans, la paire actions monde / obligations longues va de −0,42 à +0,58.",
	} {
		b.WriteString(sTxt(24, 392+float64(i)*14, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 452, b.String())
}
