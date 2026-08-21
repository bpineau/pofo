package firebook

import (
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// The plate reads the same bundled series as the crisis-report plate, through
// the same helpers (defensesAsOf, defensesCPI, defensesRates,
// defensesCashGrowth in figures_defenses_test.go): one convention for real
// dollar returns across the campaign, written once.

// douleurYearly returns, for one bundled series, its real total return in each
// calendar year of the window, in percent. The bounds are December 31 to
// December 31 and the deflator is the December CPI of the same two years.
func douleurYearly(t *testing.T, id string, ref bool, cpi map[string]float64) map[int]float64 {
	t.Helper()
	fsys := datasets.Simdata()
	if ref {
		fsys = datasets.Refdata()
	}
	s, ok, err := marketdata.ReadSimdataFS(fsys, id)
	if err != nil || !ok {
		t.Fatalf("read %s: ok=%v err=%v", id, ok, err)
	}
	out := map[int]float64{}
	for y := douleurFirst; y <= douleurLast; y++ {
		from := defensesAsOf(t, s, defensesDay(t, endOfYear(y-1)))
		to := defensesAsOf(t, s, defensesDay(t, endOfYear(y)))
		out[y] = (to/from/douleurInflation(t, cpi, y) - 1) * 100
	}
	return out
}

// endOfYear is the December 31 bound of a calendar year.
func endOfYear(y int) string { return time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC).Format("2006-01-02") }

// douleurInflation is the CPI growth of one calendar year, December on
// December.
func douleurInflation(t *testing.T, cpi map[string]float64, y int) float64 {
	t.Helper()
	from, to := cpi[endOfYearMonth(y-1)], cpi[endOfYearMonth(y)]
	if from == 0 || to == 0 {
		t.Fatalf("no CPI for December %d or %d", y-1, y)
	}
	return to / from
}

func endOfYearMonth(y int) string { return time.Date(y, 12, 1, 0, 0, 0, 0, time.UTC).Format("2006-01") }

// douleurCashYearly is the same for cash, which is a RATE series rather than an
// index: each calendar year earns its twelve monthly rates, day by day.
func douleurCashYearly(t *testing.T, cpi map[string]float64) map[int]float64 {
	t.Helper()
	rates := defensesRates(t)
	out := map[int]float64{}
	for y := douleurFirst; y <= douleurLast; y++ {
		grow := defensesCashGrowth(t, rates, defensesDay(t, endOfYear(y-1)), defensesDay(t, endOfYear(y)))
		out[y] = (grow/douleurInflation(t, cpi, y) - 1) * 100
	}
	return out
}

// douleurBlend composes the 60/40 the book keeps as its reference mix: sixty
// percent equities, forty percent intermediate Treasuries, weights reset every
// December 31, which is exactly what a yearly return series of each leg allows.
func douleurBlend(equities, bonds map[int]float64) map[int]float64 {
	out := map[int]float64{}
	for y, e := range equities {
		out[y] = (0.6*(1+e/100) + 0.4*(1+bonds[y]/100) - 1) * 100
	}
	return out
}

// douleurCAGR is the annualized real return of a yearly series over the window,
// in percent.
func douleurCAGR(r map[int]float64) float64 {
	grow, n := 1.0, 0
	for y := douleurFirst; y <= douleurLast; y++ {
		grow *= 1 + r[y]/100
		n++
	}
	return (math.Pow(grow, 1/float64(n)) - 1) * 100
}

// douleurMean is the arithmetic mean of a yearly series over the given years.
func douleurMean(r map[int]float64, years []int) float64 {
	sum := 0.0
	for _, y := range years {
		sum += r[y]
	}
	return sum / float64(len(years))
}

// douleurMeasure rebuilds everything the plate freezes: the ten worst equity
// years, both coordinates of every point, and the least-squares line through
// the wage-earning ones.
func douleurMeasure(t *testing.T) (worst []int, pain, premium map[string]float64, a, b float64) {
	t.Helper()
	frozenAgainstData(t)
	cpi := defensesCPI(t)

	yearly := map[string]map[int]float64{
		"Actions US":       douleurYearly(t, "SP500", false, cpi),
		"Treasuries 7-10":  douleurYearly(t, "IEF", false, cpi),
		"Treasuries 20+":   douleurYearly(t, "TLT", false, cpi),
		"Or":               douleurYearly(t, "XAUUSD", false, cpi),
		"Cash (bons 3 m.)": douleurCashYearly(t, cpi),
	}
	yearly["60/40"] = douleurBlend(yearly["Actions US"], yearly["Treasuries 7-10"])

	// the ten worst calendar years of equities, inside the window
	years := make([]int, 0, douleurLast-douleurFirst+1)
	for y := douleurFirst; y <= douleurLast; y++ {
		years = append(years, y)
	}
	sort.SliceStable(years, func(i, j int) bool {
		return yearly["Actions US"][years[i]] < yearly["Actions US"][years[j]]
	})
	worst = append([]int{}, years[:douleurWorstN]...)

	cash := douleurCAGR(yearly["Cash (bons 3 m.)"])
	pain, premium = map[string]float64{}, map[string]float64{}
	for name, r := range yearly {
		pain[name] = douleurMean(r, worst)
		premium[name] = douleurCAGR(r) - cash
	}

	// the least-squares line through the wage-earning assets only
	var sx, sy, sxx, sxy, n float64
	for _, name := range []string{"Actions US", "Treasuries 7-10", "Treasuries 20+", "60/40", "Cash (bons 3 m.)"} {
		x, y := pain[name], premium[name]
		sx, sy, sxx, sxy, n = sx+x, sy+y, sxx+x*x, sxy+x*y, n+1
	}
	b = (n*sxy - sx*sy) / (n*sxx - sx*sx)
	a = (sy - b*sx) / n
	return worst, pain, premium, a, b
}

// The plate's frozen coordinates, the ten worst years and the fitted line,
// rebuilt from the bundled series. This is what "make figure-drift" reports
// when a refresh moves them.
func TestDouleurCloudMatchesTheData(t *testing.T) {
	worst, pain, premium, a, b := douleurMeasure(t)
	for i, y := range worst {
		if y != douleurWorstYears[i] {
			t.Errorf("worst year %d is %d, the plate freezes %d", i+1, y, douleurWorstYears[i])
		}
	}
	// the plate's names, in the order douleurAssets holds them
	measured := []string{"Actions US", "60/40", "Treasuries 7-10", "Treasuries 20+",
		"Cash (bons 3 m.)", "Or"}
	for i, key := range measured {
		got := douleurAssets[i]
		if math.Abs(pain[key]-got.pain) > 0.05 {
			t.Errorf("%s: the data puts the pain at %.2f %%, the plate draws %.2f",
				got.name, pain[key], got.pain)
		}
		if math.Abs(premium[key]-got.premium) > 0.05 {
			t.Errorf("%s: the data puts the premium at %.2f points, the plate draws %.2f",
				got.name, premium[key], got.premium)
		}
	}
	if math.Abs(a-douleurFitA) > 0.01 || math.Abs(b-douleurFitB) > 0.005 {
		t.Errorf("the wage line is premium = %.4f %+.4f * pain, the plate draws %.4f %+.4f",
			a, b, douleurFitA, douleurFitB)
	}
	if got := douleurR2(pain, premium, a, b); math.Abs(got-douleurFitR2) > 0.005 {
		t.Errorf("the fit accounts for %.4f of the wages, the plate says %.4f", got, douleurFitR2)
	}
}

// douleurR2 is the share of the wage assets' premiums the fitted line accounts
// for, which the plate prints in its legend.
func douleurR2(pain, premium map[string]float64, a, b float64) float64 {
	wage := []string{"Actions US", "60/40", "Treasuries 7-10", "Treasuries 20+", "Cash (bons 3 m.)"}
	mean := 0.0
	for _, w := range wage {
		mean += premium[w]
	}
	mean /= float64(len(wage))
	var total, resid float64
	for _, w := range wage {
		total += (premium[w] - mean) * (premium[w] - mean)
		r := premium[w] - (a + b*pain[w])
		resid += r * r
	}
	return 1 - resid/total
}

// The two properties the plate exists to show, checked on the frozen numbers
// alone: the assets that claim a wage sit on one line, and gold does not. If
// either failed, the plate would be arguing with its own article.
func TestDouleurWagesAlignAndGoldDoesNot(t *testing.T) {
	if douleurFitB >= 0 {
		t.Errorf("the wage line slopes %+.3f: more pain would mean less pay", douleurFitB)
	}
	worstWage := 0.0
	for _, a := range douleurAssets {
		if !a.wage {
			continue
		}
		if r := math.Abs(a.premium - douleurPredict(a.pain)); r > worstWage {
			worstWage = r
		}
	}
	if worstWage > 1.5 {
		t.Errorf("a wage asset sits %.2f points off the line: they no longer align", worstWage)
	}
	gold := douleurGoldResidual()
	if gold < 3 {
		t.Errorf("gold sits only %.2f points off the line: the plate's exception is gone", gold)
	}
	if gold < 2*worstWage {
		t.Errorf("gold's %.2f points are not clearly apart from the wage assets' %.2f",
			gold, worstWage)
	}
	// Gold is paid without hurting: that is the whole reason it is not a wage.
	if douleurGold().pain <= 0 {
		t.Errorf("gold loses %.2f %% in the ten worst equity years: it would be earning its pay",
			douleurGold().pain)
	}
	// Cash is the origin by construction.
	for _, a := range douleurAssets {
		if a.labelSuppressed && a.premium != 0 {
			t.Errorf("cash carries a premium of %.2f points over itself", a.premium)
		}
	}
}

// The plate against the article that carries it.
func TestDouleurAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "primes-de-risque")
	for _, want := range []string{
		"::: figure douleur-prime",
		"il rapporte parce qu'il fait mal au mauvais moment",
		"entre 4 et 6 points par an en moyenne géométrique",
		"il n'y a pas de prime de risque de l'or",
		"ce qui compte n'est pas la volatilité en soi, mais la **covariance avec les mauvais états du monde**",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The equity premium the plate measures has to land in the band the
	// article quotes, or one of the two is wrong.
	if eq := douleurAssets[0].premium; eq < 4 || eq > 6 {
		t.Errorf("the plate measures an equity premium of %.2f points, the article says 4 to 6", eq)
	}
	// And equities have to be the asset that hurts most, since they define the
	// years the pain is measured in.
	for _, a := range douleurAssets[1:] {
		if a.pain < douleurAssets[0].pain {
			t.Errorf("%q loses more than equities in equities' own worst years", a.name)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestDouleurPlateRenders(t *testing.T) {
	svg := FigureSVG("douleur-prime")
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
	for _, a := range douleurAssets {
		if a.labelSuppressed {
			continue
		}
		if !strings.Contains(svg, ">"+a.name+"<") {
			t.Errorf("the plate does not name %q", a.name)
		}
		if want := ">" + douleurPts(a.premium) + "<"; !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q for %q", want, a.name)
		}
	}
	for _, want := range []string{
		">assurance, pas salaire<",
		">la droite des salaires<",
		">+4,2 pts<",
		"Cash (bons du Trésor 3 mois) : l'étalon, prime nulle par construction",
		"R² = 0,86",
		"L'or part du prix fixe de 35 dollars abandonné en 1971",
		"2008, 1974, 2002, 2022, 1973, 1969, 1977, 2001, 1981, 2000.",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// One dot per asset, and no other circle.
	if n := strings.Count(svg, "<circle"); n != len(douleurAssets) {
		t.Errorf("%d dots drawn, expected one per asset (%d)", n, len(douleurAssets))
	}
	// The wage line in two segments plus gold's drop: three polylines.
	if n := strings.Count(svg, "<polyline"); n != 3 {
		t.Errorf("%d lines drawn, expected the wage line's two segments and gold's drop", n)
	}
}
