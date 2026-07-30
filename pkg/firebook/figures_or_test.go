package firebook

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// orRealDecember returns the real (CPI-deflated) gold level of every December
// the bundled record covers, keyed by year: the last XAU/USD close of the month
// over the ^CPI-US level of the same month. This is the computation the gold
// plate freezes, rebuilt from pkg/datasets and marketdata's offline snapshot.
func orRealDecember(t *testing.T) map[int]float64 {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), "XAUUSD")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("bundled XAUUSD simdata not found")
	}
	gold := map[int]float64{}
	for _, p := range s.Points {
		if p.Close > 0 && p.Date.Month() == time.December {
			gold[p.Date.Year()] = p.Close // last December close wins
		}
	}
	cs, err := marketdata.NewClient("").Fetch(context.Background(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	cpi := map[int]float64{}
	for _, p := range cs.Points {
		if p.Close > 0 && p.Date.Month() == time.December {
			cpi[p.Date.Year()] = p.Close
		}
	}
	out := map[int]float64{}
	for y, g := range gold {
		if c, ok := cpi[y]; ok {
			out[y] = g / c
		}
	}
	return out
}

// The gold plate freezes one annualized and one cumulative real return per
// ten-year window; recompute every one of them from the bundled series.
func TestOrDecadesMatchTheData(t *testing.T) {
	real := orRealDecember(t)
	for i := range orDecadeAnn {
		start := orDecadeFirst + orDecadeStep*i
		base, end := real[start-1], real[start+9]
		if base == 0 || end == 0 {
			t.Fatalf("window %s: the record misses December %d or %d", orDecadeLabel(start), start-1, start+9)
		}
		r := end / base
		if got := (math.Pow(r, 0.1) - 1) * 100; math.Abs(got-orDecadeAnn[i]) > 0.02 {
			t.Errorf("window %s: data %+.2f %%/yr, plate %+.2f %%/yr", orDecadeLabel(start), got, orDecadeAnn[i])
		}
		if got := (r - 1) * 100; math.Abs(got-orDecadeTotal[i]) > 0.15 {
			t.Errorf("window %s: data %+.1f %% cumulative, plate %+.1f %%", orDecadeLabel(start), got, orDecadeTotal[i])
		}
	}

	// The running window, six years long, and the reason it stops in 2025: no
	// later December is complete in both series.
	years := float64(orPartialEnd - (orPartialStart - 1))
	r := real[orPartialEnd] / real[orPartialStart-1]
	if r == 0 || math.IsInf(r, 0) {
		t.Fatalf("the record misses December %d or %d", orPartialStart-1, orPartialEnd)
	}
	if got := (math.Pow(r, 1/years) - 1) * 100; math.Abs(got-orPartialAnn) > 0.02 {
		t.Errorf("window %d-%d: data %+.2f %%/yr, plate %+.2f %%/yr", orPartialStart, orPartialEnd, got, orPartialAnn)
	}
	if got := (r - 1) * 100; math.Abs(got-orPartialTotal) > 0.15 {
		t.Errorf("window %d-%d: data %+.1f %% cumulative, plate %+.1f %%", orPartialStart, orPartialEnd, got, orPartialTotal)
	}
	if _, ok := real[orPartialEnd+1]; ok {
		t.Errorf("December %d is now covered: the running window must be extended", orPartialEnd+1)
	}
	// The running window opens the year after the last complete one started, so
	// the eleven bars tile the record without a hole.
	if last := orDecadeFirst + orDecadeStep*(len(orDecadeAnn)-1); last+orDecadeStep != orPartialStart {
		t.Errorf("the last complete window starts in %d, the running one in %d", last, orPartialStart)
	}

	// The two averages the plate draws and quotes.
	var sum, exc float64
	for i, v := range orDecadeAnn {
		sum += v
		if i > 0 {
			exc += v
		}
	}
	if got := sum / float64(len(orDecadeAnn)); math.Abs(got-orDecadeMean) > 0.02 {
		t.Errorf("mean of the windows %+.2f, plate %+.2f", got, orDecadeMean)
	}
	if got := exc / float64(len(orDecadeAnn)-1); math.Abs(got-orDecadeMeanEx70) > 0.02 {
		t.Errorf("mean without 1970-79 %+.2f, plate %+.2f", got, orDecadeMeanEx70)
	}

	// The counts the title makes: five windows in the red, three at double
	// digits (the eleven drawn bars, running window included).
	red, big := 0, 0
	for _, v := range append(append([]float64{}, orDecadeAnn...), orPartialAnn) {
		switch {
		case v < 0:
			red++
		case v >= 10:
			big++
		}
	}
	if red != 5 || big != 3 {
		t.Errorf("%d windows in the red and %d at double digits, the title says 5 and 3", red, big)
	}
	// And the plate's own claim of a continuous purgatory: the five negative
	// windows are the five consecutive ones the wash covers.
	for i := 1; i <= 5; i++ {
		if orDecadeAnn[i] >= 0 {
			t.Errorf("window %s is not negative, the wash claims 1975 to 2004 is", orDecadeLabel(orDecadeFirst+orDecadeStep*i))
		}
	}
}

// The article enumerates five gold episodes with numbers. They are read off the
// same bundled series as the plate, so they must stay in step with it: this
// pins each one to the reading it comes from, with room for the rounding the
// prose uses.
func TestOrEpisodesMatchTheData(t *testing.T) {
	s, _, err := marketdata.ReadSimdataFS(datasets.Simdata(), "XAUUSD")
	if err != nil {
		t.Fatal(err)
	}
	month := map[int]float64{} // last close of each month, keyed year*12+month-1
	var lo11to15, hi11 float64
	for _, p := range s.Points {
		if p.Close <= 0 {
			continue
		}
		month[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close
		switch y := p.Date.Year(); {
		case y == 2011 && p.Close > hi11:
			hi11 = p.Close
		case y == 2015 && (lo11to15 == 0 || p.Close < lo11to15):
			lo11to15 = p.Close
		}
	}
	dec := func(y int) float64 { return month[y*12+11] }
	real := orRealDecember(t)
	within := func(name string, got, want, tol float64) {
		if math.Abs(got-want) > tol {
			t.Errorf("%s: data %+.1f, the article says %+.1f (tolerance %.1f)", name, got, want, tol)
		}
	}
	// "1971-1980 -> +1 200 % nominal": end-1971 to end-1980.
	within("1971-1980 nominal", (dec(1980)/dec(1971)-1)*100, 1200, 60)
	// "1980-2000 -> −70 % réel": end-1979 to end-1999, real.
	within("1980-2000 real", (real[1999]/real[1979]-1)*100, -70, 5)
	// "2000-2011 -> x5": end-1999 to end-2011, nominal (the 2011 intra-year peak
	// would read x6.5, so the article's round figure is the conservative one).
	within("2000-2011 multiple", dec(2011)/dec(1999), 5, 0.6)
	// "2011-2015 -> −40 %": the 2011 peak to the 2015 trough, nominal.
	within("2011-2015 nominal", (lo11to15/hi11-1)*100, -40, 5)
	// "+5 % en 2008" and "+25 % en 2020", calendar years, nominal.
	within("2008 nominal", (dec(2008)/dec(2007)-1)*100, 5, 1.5)
	within("2020 nominal", (dec(2020)/dec(2019)-1)*100, 25, 1.5)
	// "2022 : match nul en dollars" while bonds had their worst year.
	within("2022 nominal", (dec(2022)/dec(2021)-1)*100, 0, 1.5)
	// The 2022 high stayed under the 2020 record, which is why the article no
	// longer credits gold with new dollar highs that year.
	var hi20, hi22 float64
	for _, p := range s.Points {
		if p.Date.Year() == 2020 && p.Close > hi20 {
			hi20 = p.Close
		}
		if p.Date.Year() == 2022 && p.Close > hi22 {
			hi22 = p.Close
		}
	}
	if hi22 >= hi20 {
		t.Errorf("2022 peaked at %.0f against 2020's %.0f: the article's wording assumes it did not", hi22, hi20)
	}
	// "volatilité 15-20 %": annualized from monthly returns, whole record and
	// modern era, the two ends of the range the article gives.
	vol := func(from int) float64 {
		var r []float64
		for k := from*12 + 1; k <= 2026*12; k++ {
			if month[k] > 0 && month[k-1] > 0 {
				r = append(r, month[k]/month[k-1]-1)
			}
		}
		var mu float64
		for _, x := range r {
			mu += x
		}
		mu /= float64(len(r))
		var v float64
		for _, x := range r {
			v += (x - mu) * (x - mu)
		}
		return math.Sqrt(v/float64(len(r)-1)) * math.Sqrt(12) * 100
	}
	for _, from := range []int{1968, 1990} {
		if v := vol(from); v < 15 || v > 20 {
			t.Errorf("gold vol since %d is %.1f %%, outside the article's 15-20 %%", from, v)
		}
	}
}

// The plate must render, register under its id and stay inside the house rules.
func TestOrPlateRenders(t *testing.T) {
	svg := FigureSVG("or-decennies")
	if svg == "" {
		t.Fatal("or-decennies is not registered")
	}
	if strings.Contains(svg, "rgba(") || strings.Contains(svg, "opacity") {
		t.Error("no rgba and no opacity in a plate: crengine paints them black")
	}
	if strings.Contains(svg, "rotate") {
		t.Error("every label stays horizontal")
	}
	if strings.Contains(svg, "—") {
		t.Error("no em-dash")
	}
	for _, want := range []string{"1970-79", "2015-24", "2020-25", "+21,6", "−7,2", "+14,5", "CPI américain"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate is missing %q", want)
		}
	}
}
