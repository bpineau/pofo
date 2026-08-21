package firebook

import (
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The plate's engine, run exactly as the simulator's FIRE page runs its central
// model: one monthly panel of real returns built from the bundled series, the
// Student-t parameters FITTED on that panel (arithmetic mean of the realised
// annual real returns, monthly dispersion times root twelve, degrees of freedom
// seeded from the monthly excess kurtosis), and the ruin kernel of pkg/decumul
// over that source.
//
// plancherPanel and plancherFit mirror web.BuildMonthlyPanel and web.FitParametric
// line for line; they are re-derived here rather than called because
// pkg/decumul/web imports this package to serve the book, so importing it back
// would cycle. Any change to the page's fit has to be mirrored here, and the
// comments below name the counterpart it copies.
//
// The page's blend toward the conservative prior is deliberately inert here and
// this is not a deviation: it only fires when the history is SHORTER than the
// horizon, and the common window is thirty-nine years against a thirty-year
// plan. Both portfolios read the same panel, so the only thing that changes
// between the two runs is the weight vector.

// plancherMeasured is one portfolio's whole reading.
type plancherMeasured struct {
	name           string
	mu, sigma, df  float64
	median, swr    float64
	weights        []float64
	panelFirstLast string
}

// plancherLegs are the four bricks the article names, in the order the plate
// weights them. Every one is a bundled series.
var plancherLegs = []string{"MSCIWORLD", "TLT", "XAUUSD", "BTOP50"}

// plancherSeries reads one bundled leg.
func plancherSeries(t *testing.T, id string) []marketdata.Point {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), id)
	if err != nil || !ok {
		t.Fatalf("read %s: ok=%v err=%v", id, ok, err)
	}
	return s.Points
}

// plancherPanel builds the one monthly real-return panel both portfolios read:
// the four legs deflated by the US CPI and aligned on the months all four
// quote, which the trend index bounds at the start of 1987.
func plancherPanel(t *testing.T) (scenario.Panel, string) {
	t.Helper()
	cpi, err := marketdata.NewClient("").Fetch(t.Context(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	// One map of real monthly returns per leg, keyed by a dense month index, and
	// a count of how many legs cover each month (web.BuildMonthlyPanel).
	perLeg := make([]map[int]float64, len(plancherLegs))
	counts := map[int]int{}
	for i, id := range plancherLegs {
		pts := plancherMonthEnds(plancherSeries(t, id))
		rets := scenario.Deflate(pts, cpi.Points)
		m := make(map[int]float64, len(rets))
		for j, r := range rets {
			key := plancherMonthKey(pts[j+1].Date)
			if key != plancherMonthKey(pts[j].Date)+1 {
				continue // not calendar-consecutive: a spanning return, drop it
			}
			m[key] = r
			counts[key]++
		}
		perLeg[i] = m
	}
	var common []int
	for key, c := range counts {
		if c == len(plancherLegs) {
			common = append(common, key)
		}
	}
	sort.Ints(common)
	if len(common) == 0 {
		t.Fatal("the four legs share no month")
	}
	rows := make([][]float64, len(plancherLegs))
	for i := range plancherLegs {
		row := make([]float64, len(common))
		for k, key := range common {
			row[k] = perLeg[i][key]
		}
		rows[i] = row
	}
	span := func(key int) string {
		return time.Date(key/12, time.Month(key%12+1), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
	}
	return scenario.Panel{Returns: rows}, span(common[0]) + " -> " + span(common[len(common)-1])
}

// plancherMonthEnds keeps the last quote of each calendar month (lastPerMonth).
func plancherMonthEnds(points []marketdata.Point) []marketdata.Point {
	var out []marketdata.Point
	for _, p := range points {
		if n := len(out); n > 0 && out[n-1].Date.Year() == p.Date.Year() &&
			out[n-1].Date.Month() == p.Date.Month() {
			out[n-1] = p
			continue
		}
		out = append(out, p)
	}
	return out
}

// plancherMonthKey maps a date to a dense month index, so consecutive calendar
// months differ by exactly one (monthKey).
func plancherMonthKey(t time.Time) int { return t.Year()*12 + int(t.Month()) - 1 }

// plancherFit is web.FitParametric: the arithmetic mean of the realised annual
// real returns, the monthly dispersion annualized by root twelve, and a
// Student-t dof seeded from the monthly excess kurtosis (inverting the t excess
// kurtosis 6/(df-4), clamped to the page's slider range).
func plancherFit(panel scenario.Panel, weights []float64) (mu, sigma, df float64) {
	monthly := panel.Combine(weights)
	annual := scenario.Annualize(monthly, 12)
	mean := metrics.Mean(monthly)
	var ss float64
	for _, x := range monthly {
		ss += (x - mean) * (x - mean)
	}
	sigma = math.Sqrt(ss/float64(len(monthly)-1)) * math.Sqrt(12)
	excess := metrics.ExcessKurtosis(monthly)
	switch {
	case math.IsNaN(excess) || excess <= 0:
		df = 30
	default:
		df = math.Max(3, math.Min(4+6/excess, 30))
	}
	return metrics.Mean(annual), sigma, df
}

// plancherMeasure runs both portfolios end to end and returns the four numbers
// the plate freezes, plus the fitted parameters behind them.
func plancherMeasure(t *testing.T) []plancherMeasured {
	t.Helper()
	frozenAgainstData(t)
	panel, window := plancherPanel(t)

	out := make([]plancherMeasured, 0, len(plancherPortfolios))
	for _, p := range plancherPortfolios {
		mu, sigma, df := plancherFit(panel, p.weights)
		if sigma <= 0 {
			t.Fatalf("%s: the panel yields no usable fit", p.name)
		}
		plan := decumul.Plan{
			Capital:    plancherCapital,
			NeedAnnual: plancherCapital * plancherRule,
			Years:      plancherYears,
			Source: scenario.ParametricSource{
				Mu: mu, Sigma: sigma, Df: df, Periods: plancherYears},
		}
		median := plan.Simulate(plancherPaths, plancherWorkers, plancherSeed).Outcome().TerminalP50
		safe := plan.Solve(1-plancherSuccess,
			decumul.WithdrawalAxis(0, plancherCapital*0.15), plancherPaths, plancherWorkers, plancherSeed)
		out = append(out, plancherMeasured{
			name: p.name, mu: mu, sigma: sigma, df: df,
			median: median / plancherCapital, swr: safe / plancherCapital,
			weights: p.weights, panelFirstLast: window,
		})
	}
	return out
}

// The plate's four frozen numbers, rebuilt from the bundled series and the
// simulator's own engine. This is what "make figure-drift" reports when a data
// refresh moves them.
func TestMedianePlancherNumbersMatchTheModel(t *testing.T) {
	got := plancherMeasure(t)
	for i, m := range got {
		p := plancherPortfolios[i]
		if math.Abs(m.median-p.median) > 0.02 {
			t.Errorf("%s: the model's median wealth is %.3f times the capital, the plate draws %.3f",
				p.name, m.median, p.median)
		}
		if math.Abs(m.swr-p.swr) > 0.0005 {
			t.Errorf("%s: the model's safe withdrawal is %.4f, the plate draws %.4f",
				p.name, m.swr, p.swr)
		}
	}
	// The fit itself has to keep saying what the plate assumes of it: the
	// basket is the calmer and the less rewarded of the two, which is the
	// premise of the whole trade.
	eq, basket := got[0], got[1]
	if basket.sigma >= eq.sigma {
		t.Errorf("the basket's volatility (%.3f) no longer falls below all equities' (%.3f)",
			basket.sigma, eq.sigma)
	}
	if basket.mu >= eq.mu {
		t.Errorf("the basket's mean return (%.3f) no longer falls below all equities' (%.3f)",
			basket.mu, eq.mu)
	}
	if got[0].panelFirstLast[:4] != plancherWindow[:4] {
		t.Errorf("the common window now starts in %s, the plate says %s",
			got[0].panelFirstLast[:4], plancherWindow[:4])
	}
}

// The two claims the article makes, checked on the frozen numbers: going from
// all equities to the four-brick basket LOWERS the median and RAISES the safe
// withdrawal. If either flipped, the plate would be arguing with its article.
func TestMedianePlancherProvesTheArticlesTwoClaims(t *testing.T) {
	eq, basket := plancherPortfolios[0], plancherPortfolios[1]
	if basket.median >= eq.median {
		t.Errorf("the basket's median wealth (%.2f) no longer falls below all equities' (%.2f)",
			basket.median, eq.median)
	}
	if basket.swr <= eq.swr {
		t.Errorf("the basket's safe withdrawal (%.3f %%) no longer rises above all equities' (%.3f %%)",
			basket.swr*100, eq.swr*100)
	}
	// The article's "un peu": a median that collapsed would make the trade a
	// bad one rather than the bargain the article describes.
	if drop := 1 - basket.median/eq.median; drop > 0.35 {
		t.Errorf("the basket costs %.0f %% of the median wealth: no longer 'un peu'", drop*100)
	}
	if len(basket.weights) != 4 {
		t.Errorf("the basket has %d bricks, the article names four", len(basket.weights))
	}
}

// The plate against the article that carries it.
func TestMedianePlancherAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "pourquoi-la-diversification-marche")
	for _, want := range []string{
		"::: figure panier-mediane-plancher",
		"quatre briques bien choisies (actions mondiales, obligations longues, or, trend au sens du suivi de tendance)",
		"la richesse médiane à trente ans baisse, d'un quart pour le panier mesuré ci-dessous, tandis que le SWR à 95 % de succès monte",
		"Il diversifie pour le percentile 5",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestMedianePlancherPlateRenders(t *testing.T) {
	svg := FigureSVG("panier-mediane-plancher")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"\u2014", "\u2013", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, p := range plancherPortfolios {
		if !strings.Contains(svg, ">"+p.name+"<") {
			t.Errorf("the plate does not name %q", p.name)
		}
		if want := ">" + plancherWealth(p.median) + "<"; !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw the median %q for %q", want, p.name)
		}
		if want := ">" + plancherRate(p.swr) + "<"; !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw the safe rate %q for %q", want, p.name)
		}
	}
	for _, want := range []string{
		"richesse médiane à 30 ans",
		"retrait sûr à 95 % de succès",
		"Le panier échange 25 % de médiane contre 0,78 point de retrait sûr",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// Two segments, one per portfolio, and four dots.
	if n := strings.Count(svg, "<polyline"); n != len(plancherPortfolios) {
		t.Errorf("%d segments drawn, expected one per portfolio", n)
	}
	if n := strings.Count(svg, "<circle"); n != 2*len(plancherPortfolios) {
		t.Errorf("%d dots drawn, expected two per portfolio", n)
	}
}
