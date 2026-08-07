package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The all-weather article's second plate: the gap to the index, year by year.
// Where figTousTempsEchange draws what the family buys (a shallow worst path),
// this one draws what it costs, and the cost is measured in consecutive years
// of lagging, not in points of return. See figTousTempsEcart.

// ecartYear holds one calendar year of the two legs, in percent of real return.
type ecartYear struct {
	year       int
	allWeather float64 // Golden Butterfly
	equities   float64 // 100 % S&P 500
}

// The Golden Butterfly (20 % equities, 20 % small-cap value, 20 % long
// Treasuries, 20 % short, 20 % gold) against 100 % equities, calendar year by
// calendar year, in real terms, over 1972-2024.
//
// Same reconstruction as figTousTempsEchange, on the same bundled series:
// equities SP500, long Treasuries TLT, short Treasuries SHY (T-bills before
// SHY starts in 1991-10), gold XAUUSD (pkg/datasets/simdata); small-cap value
// USSCV-USD, T-bills TBILL-3M (pkg/datasets/refdata); deflator ^CPI-US (the
// snapshot pkg/marketdata embeds). Monthly total returns, weights reset every
// December, each year read December to December.
//
// TestEcartFigureMatchesTheData recomputes both legs from those series and
// fails the moment the plate and the data disagree.
var ecartYears = []ecartYear{
	{1972, 11.43, 15.01}, {1973, -1.19, -21.63}, {1974, -5.24, -34.62},
	{1975, 9.33, 28.34}, {1976, 14.33, 17.90}, {1977, 1.87, -13.25},
	{1978, 5.12, -2.40}, {1979, 21.71, 4.42}, {1980, 3.05, 17.51},
	{1981, -9.34, -12.86}, {1982, 22.18, 16.98}, {1983, 9.30, 17.99},
	{1984, 0.14, 2.10}, {1985, 17.13, 26.81}, {1986, 16.15, 17.30},
	{1987, 0.73, 0.69}, {1988, 4.93, 11.82}, {1989, 9.32, 25.84},
	{1990, -8.92, -8.68}, {1991, 14.06, 26.59}, {1992, 7.28, 4.59},
	{1993, 12.19, 7.13}, {1994, -3.98, -1.32}, {1995, 19.86, 34.17},
	{1996, 5.84, 19.01}, {1997, 12.52, 31.13}, {1998, 7.15, 26.54},
	{1999, 2.04, 17.88}, {2000, 3.92, -12.08}, {2001, 3.65, -13.23},
	{2002, 1.20, -23.91}, {2003, 20.98, 26.31}, {2004, 5.40, 7.39},
	{2005, 4.90, 1.45}, {2006, 10.19, 12.93}, {2007, 3.94, 1.36},
	{2008, -4.97, -37.06}, {2009, 9.03, 23.11}, {2010, 14.77, 13.37},
	{2011, 4.92, -0.83}, {2012, 7.55, 14.02}, {2013, 5.19, 30.43},
	{2014, 7.91, 12.83}, {2015, -4.72, 0.65}, {2016, 9.58, 9.68},
	{2017, 8.52, 19.32}, {2018, -5.66, -6.18}, {2019, 13.92, 28.55},
	{2020, 12.11, 16.81}, {2021, 5.05, 20.24}, {2022, -17.32, -23.08},
	{2023, 8.16, 22.19}, {2024, 8.37, 21.51},
}

// gap is the year's shortfall against the index, in points of real return.
func (e ecartYear) gap() float64 { return e.allWeather - e.equities }

// ecartStreak is a run of consecutive years spent behind the index.
type ecartStreak struct{ from, to, length int }

// ecartStreaks lists the runs of at least min consecutive lagging years, in
// chronological order. The plate washes them into background bands, because
// the ordeal the article warns about is their length, not their total.
func ecartStreaks(min int) []ecartStreak {
	var out []ecartStreak
	for i := 0; i < len(ecartYears); {
		if ecartYears[i].gap() >= 0 {
			i++
			continue
		}
		j := i
		for j < len(ecartYears) && ecartYears[j].gap() < 0 {
			j++
		}
		if j-i >= min {
			out = append(out, ecartStreak{ecartYears[i].year, ecartYears[j-1].year, j - i})
		}
		i = j
	}
	return out
}

// ecartBandMin is the length from which a lagging run gets its own band. Four
// years is where the run stops being a bad patch and becomes a doubt.
const ecartBandMin = 4

// figTousTempsEcart draws the criticism the article calls the most serious
// one: the yearly gap between an all-weather portfolio and the index. The
// signed bars say how much, the washed bands say how long, and nothing
// cumulates the two, because the reader gives up on the fourth consecutive red
// bar and not on the total.
func figTousTempsEcart() string {
	m := mapper(1972, 2025, -32, 36, 56, 616, 338, 92)
	y := func(v float64) float64 { return m(1972, v)[1] }
	x := func(yr float64) float64 { return m(yr, 0)[0] }
	var b strings.Builder
	b.WriteString(plateHead("portefeuilles tous-temps", "Le retard sur l'indice se compte en années, pas en points"))
	legendChips(&b, 54, [][2]string{
		{figBad, "années de retard sur l'indice"},
		{figBlue, "années d'avance (les crises)"},
	})
	b.WriteString(sTxt(56, 82, 10.5, figMuted, "start", "400", "écart annuel, en points de rendement réel"))

	// the lagging runs, washed in behind everything
	for _, s := range ecartStreaks(ecartBandMin) {
		x0, x1 := x(float64(s.from)), x(float64(s.to+1))
		fmt.Fprintf(&b, `<rect x="%.1f" y="92" width="%.1f" height="246" fill="%s"/>`, x0, x1-x0, figWash)
		b.WriteString(mTxt((x0+x1)/2, 104, 11, figSoft, "middle", "600", fmt.Sprintf("%d ans", s.length)))
	}

	// grid and the zero rule the bars hang from
	for _, g := range []float64{-20, -10, 0, 10, 20, 30} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(56, gy, 616, gy, col, 1))
		b.WriteString(mTxt(48, gy+3.5, 10, figMuted, "end", "400", ttNum(g, 0)))
	}

	// the bars, one per calendar year
	const bw = 7.0
	for _, e := range ecartYears {
		cx, g := x(float64(e.year)+0.5), e.gap()
		fill := figBlue
		if g < 0 {
			fill = figBad
		}
		if math.Abs(y(g)-y(0)) < 2 { // too small to draw honestly: a signed tick
			ty := y(0) - 1.5
			if g < 0 {
				ty = y(0) + 1.5
			}
			b.WriteString(line(cx-bw/2, ty, cx+bw/2, ty, fill, 2))
			continue
		}
		b.WriteString(barV(cx-bw/2, bw, y(0), y(g), fill))
	}

	// what the whole plate adds up to, in the one clear space it leaves
	b.WriteString(sTxt(150, 136, 11, figSoft, "start", "600", "33 années de retard sur 53,"))
	b.WriteString(sTxt(150, 150, 11, figSoft, "start", "600", "pour 0,8 point de moins par an."))

	// the single year that splits the two four-year runs, named because the
	// eye reads the notch as a respite and it was not one
	b.WriteString(mTxt(x(1987.5), y(0)-9, 10, figMuted, "middle", "400", "1987 : +0,04"))

	// the two extremes: the year the insurance paid, and the year the article's
	// sentence ("+6 % quand le monde fait +25 %") really happened
	b.WriteString(mTxt(x(2008.5), y(32.08)-7, 10.5, figBlue, "middle", "600", "2008 : +32"))
	b.WriteString(sTxt(x(2013.5), y(-25.24)+14, 10.5, figBad, "middle", "600", "2013 : +5 % contre +30 %"))

	// x axis
	b.WriteString(line(56, 338, 616, 338, figRule, 1))
	for _, t := range []int{1975, 1985, 1995, 2005, 2015, 2024} {
		b.WriteString(mTxt(x(float64(t)+0.5), 354, 10, figMuted, "middle", "400", fmt.Sprintf("%d", t)))
	}

	b.WriteString(sTxt(24, 376, 9.5, figMuted, "start", "400",
		"Golden Butterfly moins 100 % actions, écart de rendement réel annuel. US, 1972-2024, rééquilibrage annuel en décembre."))
	b.WriteString(sTxt(24, 390, 9.5, figMuted, "start", "400",
		"Bandes grises : 1983-1986, 1988-1991, 1994-1999 et 2012-2017, les séries d'au moins quatre années de retard consécutives."))
	return svg(640, 402, b.String())
}
