package firebook

import (
	"fmt"
	"io/fs"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/simgen"
)

// The reconstruction window of the all-weather plate: the value of
// December 1971 is the base, December 2024 the last reading, so the plate
// covers the 53 calendar years 1972-2024, the whole free-gold record.
const (
	ttFirst = 1971
	ttLast  = 2024
	ttYears = ttLast - ttFirst
)

// ttMonths lists the month keys ("1971-12" ... "2024-12") of the window.
func ttMonths() []string {
	out := make([]string, 0, ttYears*12+1)
	for y := ttFirst; y <= ttLast; y++ {
		for mth := 1; mth <= 12; mth++ {
			if (y == ttFirst && mth < 12) || (y == ttLast && mth > 12) {
				continue
			}
			out = append(out, fmt.Sprintf("%04d-%02d", y, mth))
		}
	}
	return out
}

// ttMonthEnds reduces a bundled series to its last value of each month.
func ttMonthEnds(t *testing.T, fsys fs.FS, id string) map[string]float64 {
	t.Helper()
	frozenAgainstData(t)
	s, ok, err := marketdata.ReadSimdataFS(fsys, id)
	if err != nil || !ok {
		t.Fatalf("read %s: ok=%v err=%v", id, ok, err)
	}
	out := make(map[string]float64, len(s.Points))
	for _, p := range s.Points {
		out[p.Date.Format("2006-01")] = p.Close
	}
	return out
}

// ttStats is what one portfolio reads over the window: every field is in
// percent except worstYear, which names the calendar year worst belongs to.
type ttStats struct {
	cagr      float64 // annualized real return
	vol       float64 // standard deviation of the calendar-year real returns
	worst     float64 // worst calendar year, December to December
	worstYear int
	drawdown  float64 // deepest drawdown of the monthly real index
}

// ttCalendarYears reduces a monthly real index to one return per calendar year
// of the window, read December to December, as fractions.
func ttCalendarYears(real []float64) []float64 {
	out := make([]float64, 0, len(real)/12)
	for y := 1; y*12 < len(real); y++ {
		out = append(out, real[y*12]/real[(y-1)*12]-1)
	}
	return out
}

// ttWorstCalendarYear reads the deepest of those years, in percent, and names
// it. Both sides of the comparison are percentages on purpose: a fraction
// never falls below −1, so comparing one against a minimum already stored in
// percent freezes that minimum on the first losing year of the window and
// reports it whatever comes after.
func ttWorstCalendarYear(years []float64) (worst float64, year int) {
	worst = math.Inf(1)
	for i, r := range years {
		if pct := r * 100; pct < worst {
			worst, year = pct, ttFirst+1+i
		}
	}
	return worst, year
}

// ttVol is the volatility of a portfolio as the family's publisher states it:
// the year-to-year variation of the real returns, i.e. their sample standard
// deviation, in percent. See the control table in figures_curseur_test.go for
// why the book measures it on calendar years rather than on months.
func ttVol(years []float64) float64 {
	mean := 0.0
	for _, r := range years {
		mean += r
	}
	mean /= float64(len(years))
	sum := 0.0
	for _, r := range years {
		sum += (r - mean) * (r - mean)
	}
	return math.Sqrt(sum/float64(len(years)-1)) * 100
}

// ttRealIndex runs one portfolio the way the plate's comment says: monthly
// total returns of the legs, weights reset every December, deflated by the US
// CPI, over 1972-2024. It returns the monthly real index, based at 1 in
// December 1971.
func ttRealIndex(legs map[string]func(prev, cur string) float64,
	weights map[string]float64, cpi map[string]float64) []float64 {
	var ids []string
	for id := range weights {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	mons := ttMonths()
	hold := make(map[string]float64, len(ids))
	for _, id := range ids {
		hold[id] = weights[id]
	}
	real := make([]float64, 1, len(mons))
	real[0] = 1
	base := cpi[mons[0]]
	for i := 1; i < len(mons); i++ {
		prev, cur := mons[i-1], mons[i]
		total := 0.0
		for _, id := range ids {
			hold[id] *= 1 + legs[id](prev, cur)
			total += hold[id]
		}
		real = append(real, total/(cpi[cur]/base))
		if strings.HasSuffix(cur, "-12") {
			for _, id := range ids {
				hold[id] = total * weights[id]
			}
		}
	}

	return real
}

// ttRecompute measures one portfolio over that index.
func ttRecompute(t *testing.T, legs map[string]func(prev, cur string) float64,
	weights map[string]float64, cpi map[string]float64) ttStats {
	t.Helper()
	real := ttRealIndex(legs, weights, cpi)
	years := ttCalendarYears(real)
	out := ttStats{cagr: (math.Pow(real[len(real)-1], 1/float64(ttYears)) - 1) * 100, vol: ttVol(years)}
	out.worst, out.worstYear = ttWorstCalendarYear(years)
	// the deepest drawdown of the real index
	peak := real[0]
	for _, v := range real {
		if v > peak {
			peak = v
		}
		if d := (v/peak - 1) * 100; d < out.drawdown {
			out.drawdown = d
		}
	}
	return out
}

// ttSmallValueNet charges the small-value leg what the Ken French factor does
// not pay. The plate quotes a return a reader could aim for, so its 20 % sleeve
// cannot be an academic portfolio free of fees, commissions and spreads: the
// leg takes the same measured 1.0 %/yr the investable reconstruction takes
// (simgen.USSCVGrossCost), accrued month by month.
func ttSmallValueNet(s map[string]float64) func(prev, cur string) float64 {
	monthly := math.Pow(1-simgen.USSCVGrossCost, 1.0/12)
	return func(prev, cur string) float64 { return s[cur]/s[prev]*monthly - 1 }
}

// The plate's fourteen points are frozen literals, so the book's figures stay
// pure functions with no data dependency at render time. This rebuilds every
// one of them from the bundled series and fails the moment the plate and the
// data disagree, which is also what happens when those series are regenerated.
func TestTousTempsFigureMatchesTheData(t *testing.T) {
	frozenAgainstData(t)
	sim, ref := datasets.Simdata(), datasets.Refdata()
	price := map[string]map[string]float64{
		"equities":     ttMonthEnds(t, sim, "SP500"),
		"long":         ttMonthEnds(t, sim, "TLT"),
		"intermediate": ttMonthEnds(t, sim, "IEF"),
		"short":        ttMonthEnds(t, sim, "SHY"),
		"gold":         ttMonthEnds(t, sim, "XAUUSD"),
		"smallvalue":   ttMonthEnds(t, ref, "USSCV-USD"),
		"commodities":  ttMonthEnds(t, ref, "WTI-USD"),
	}
	bills := ttMonthEnds(t, ref, "TBILL-3M")

	// the CPI deflator, from the snapshot pkg/marketdata embeds: monthly
	// anchors dated the first of the month, served without any network.
	cpiSeries, err := marketdata.NewClient("").Fetch(t.Context(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	cpi := map[string]float64{}
	for _, p := range cpiSeries.Points {
		if p.Date.Day() == 1 {
			cpi[p.Date.Format("2006-01")] = p.Close
		}
	}

	legs := map[string]func(prev, cur string) float64{}
	for name, s := range price {
		s := s
		legs[name] = func(prev, cur string) float64 { return s[cur]/s[prev] - 1 }
	}
	// cash earns the previous month's annualized bill rate, one twelfth at a time
	legs["cash"] = func(prev, cur string) float64 { return bills[prev] / 1200 }
	legs["smallvalue"] = ttSmallValueNet(price["smallvalue"])
	// SHY only starts in 1991-10; T-bills stand in for the short sleeve before
	shy := price["short"]
	legs["short"] = func(prev, cur string) float64 {
		if p, ok := shy[prev]; ok && p != 0 {
			if c, ok := shy[cur]; ok {
				return c/p - 1
			}
		}
		return bills[prev] / 1200
	}

	cases := []struct {
		point   tousTempsPoint
		weights map[string]float64
	}{
		{tousTempsFamily[0], map[string]float64{"equities": .25, "long": .25, "gold": .25, "cash": .25}},
		{tousTempsFamily[1], map[string]float64{"equities": .30, "long": .40, "intermediate": .15, "gold": .075, "commodities": .075}},
		{tousTempsFamily[2], map[string]float64{"equities": .20, "smallvalue": .20, "long": .20, "short": .20, "gold": .20}},
	}
	for i, p := range tousTempsLadder {
		eq := float64(i) / 10
		cases = append(cases, struct {
			point   tousTempsPoint
			weights map[string]float64
		}{p, map[string]float64{"equities": eq, "intermediate": 1 - eq}})
	}
	for _, c := range cases {
		got := ttRecompute(t, legs, c.weights, cpi)
		if math.Abs(got.cagr-c.point.cagr) > 0.06 {
			t.Errorf("%s: real return %.2f %%, plate says %.2f %%", c.point.name, got.cagr, c.point.cagr)
		}
		if math.Abs(got.drawdown-c.point.drawdn) > 0.15 {
			t.Errorf("%s: worst drawdown %.1f %%, plate says %.1f %%", c.point.name, got.drawdown, c.point.drawdn)
		}
		if got.vol <= 0 || got.worst >= 0 {
			t.Errorf("%s: vol %.2f %% and worst year %.1f %% are not plausible",
				c.point.name, got.vol, got.worst)
		}
	}
}

// The worst calendar year is a running minimum, the shape that silently breaks
// when its two sides are not in the same unit. On a series whose worst year is
// the LAST one, an accumulator comparing a fraction against a percentage would
// stop at the first loser and never see it.
func TestTTWorstCalendarYearFindsALateLoser(t *testing.T) {
	yearly := []float64{0.10, -0.05, 0.20, -0.40} // 1972 to 1975
	real := make([]float64, len(yearly)*12+1)
	real[0] = 1
	for y, r := range yearly {
		step := math.Pow(1+r, 1.0/12)
		for m := 1; m <= 12; m++ {
			real[y*12+m] = real[y*12+m-1] * step
		}
	}
	years := ttCalendarYears(real)
	if len(years) != len(yearly) {
		t.Fatalf("%d calendar years read, the series holds %d", len(years), len(yearly))
	}
	worst, year := ttWorstCalendarYear(years)
	if math.Abs(worst-(-40)) > 1e-6 || year != 1975 {
		t.Errorf("worst year read %.2f %% in %d, expected −40.00 %% in 1975", worst, year)
	}
}

// The plate draws two claims, and both must survive an edit of the data: the
// ladder never reaches the family's corner, and at the Golden Butterfly's own
// return the ladder plunges twice as deep.
func TestTousTempsClaimsHold(t *testing.T) {
	shallowest := math.Inf(-1) // the ladder's least deep drawdown
	for i, p := range tousTempsLadder {
		if p.drawdn > shallowest {
			shallowest = p.drawdn
		}
		if i > 0 && p.cagr <= tousTempsLadder[i-1].cagr {
			t.Errorf("the ladder must grow in return: %s at %.2f %% after %.2f %%",
				p.name, p.cagr, tousTempsLadder[i-1].cagr)
		}
	}
	for _, p := range tousTempsFamily {
		if p.drawdn <= shallowest {
			t.Errorf("%s plunges %.1f %%, no shallower than the whole ladder (%.1f %%)",
				p.name, p.drawdn, shallowest)
		}
	}
	// the title's claim, read at the Golden Butterfly's return
	gb := tousTempsFamily[2]
	if ratio := tousTempsLadderAt(gb.cagr) / gb.drawdn; ratio < 1.7 || ratio > 2.05 {
		t.Errorf("at %.2f %% the ladder plunges %.1f times deeper, the plate title says nearly twice",
			gb.cagr, ratio)
	}
	if got := tousTempsLadderAt(6.0); math.Abs(got-(-42.1)) > 0.2 {
		t.Errorf("the ladder reads %.1f %% at 6 %%, expected about −42.1 %%", got)
	}
}

// Figure and prose must state the same numbers: the plate replaces the article's
// six-number sentence, it must not contradict it.
func TestTousTempsFigureAgreesWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/portefeuilles-tous-temps.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	if !strings.Contains(article, "::: figure tous-temps-echange") {
		t.Error("the article must carry the plate")
	}
	for _, want := range []string{
		"5,6 %", "−38 %", // 60/40
		"4,4 %", "−22 %", // Browne
		"5,8 %",          // Golden Butterfly
		"5,1 %", "−29 %", // All-Weather
		"−35 %", // the ladder's own floor
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer states %q, which the plate draws", want)
		}
	}
}
