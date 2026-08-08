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

// ttRecompute measures one portfolio the way the plate's comment says: monthly
// total returns of the legs, weights reset every December, deflated by the US
// CPI, over 1972-2024. It returns the annualized real return, the annualized
// volatility of monthly real returns, the worst calendar year and the deepest
// drawdown of the monthly real index, all in percent.
func ttRecompute(t *testing.T, legs map[string]func(prev, cur string) float64,
	weights map[string]float64, cpi map[string]float64) (cagr, vol, worst, drawdown float64) {
	t.Helper()
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

	cagr = (math.Pow(real[len(real)-1], 1/float64(ttYears)) - 1) * 100
	// the worst calendar year, read December to December
	worst = math.Inf(1)
	for y := 1; y <= ttYears; y++ {
		if r := real[y*12]/real[(y-1)*12] - 1; r < worst {
			worst = r * 100
		}
	}
	// the annualized volatility of monthly real returns
	monthly := make([]float64, 0, len(real)-1)
	for i := 1; i < len(real); i++ {
		monthly = append(monthly, real[i]/real[i-1]-1)
	}
	mean := 0.0
	for _, r := range monthly {
		mean += r
	}
	mean /= float64(len(monthly))
	sum := 0.0
	for _, r := range monthly {
		sum += (r - mean) * (r - mean)
	}
	vol = math.Sqrt(sum/float64(len(monthly)-1)) * math.Sqrt(12) * 100
	// the deepest drawdown of the real index
	peak := real[0]
	for _, v := range real {
		if v > peak {
			peak = v
		}
		if d := (v/peak - 1) * 100; d < drawdown {
			drawdown = d
		}
	}
	return cagr, vol, worst, drawdown
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
		cagr, vol, worst, dd := ttRecompute(t, legs, c.weights, cpi)
		if math.Abs(cagr-c.point.cagr) > 0.06 {
			t.Errorf("%s: real return %.2f %%, plate says %.2f %%", c.point.name, cagr, c.point.cagr)
		}
		if math.Abs(dd-c.point.drawdn) > 0.15 {
			t.Errorf("%s: worst drawdown %.1f %%, plate says %.1f %%", c.point.name, dd, c.point.drawdn)
		}
		if vol <= 0 || worst >= 0 {
			t.Errorf("%s: vol %.2f %% and worst year %.1f %% are not plausible", c.point.name, vol, worst)
		}
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
	if got := tousTempsLadderAt(6.0); math.Abs(got-(-42.0)) > 0.2 {
		t.Errorf("the ladder reads %.1f %% at 6 %%, expected about −42.0 %%", got)
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
		"5,0 %", "−29 %", // All-Weather
		"−34 %", // the ladder's own floor
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer states %q, which the plate draws", want)
		}
	}
}
