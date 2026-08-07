package firebook

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// trendCorrWindow is the rolling window the plate uses: 252 common trading days,
// i.e. the twelve months its title claims.
const trendCorrWindow = 252

// trendCorrRecord recomputes, from pkg/datasets alone, the whole series the
// conditional-correlation plate freezes: the Pearson correlation of the daily
// returns of the bundled trend programme (CTA) and of the S&P 500 (SP500) over
// the trailing 252 days both series quote, read at every month end from January
// 2001 and excluding the final incomplete month. It returns the month key of
// each reading (year*12 + month - 1) and the correlation itself.
func trendCorrRecord(t *testing.T) (months []int, corr []float64) {
	t.Helper()
	read := func(id string) *marketdata.Series {
		s, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), id)
		if err != nil || !ok {
			t.Fatalf("bundled %s simdata: ok=%v err=%v", id, ok, err)
		}
		return s
	}
	equity := map[time.Time]float64{}
	for _, p := range read("SP500").Points {
		equity[p.Date] = p.Close
	}
	var dates []time.Time
	var trendLvl, eqLvl []float64
	for _, p := range read("CTA").Points {
		c, ok := equity[p.Date]
		if !ok || c <= 0 || p.Close <= 0 {
			continue
		}
		dates = append(dates, p.Date)
		trendLvl = append(trendLvl, p.Close)
		eqLvl = append(eqLvl, c)
	}
	if len(dates) < trendCorrWindow {
		t.Fatalf("only %d days quoted by both CTA and SP500", len(dates))
	}
	rt := make([]float64, len(dates))
	re := make([]float64, len(dates))
	for i := 1; i < len(dates); i++ {
		rt[i] = trendLvl[i]/trendLvl[i-1] - 1
		re[i] = eqLvl[i]/eqLvl[i-1] - 1
	}
	for i := trendCorrWindow; i < len(dates)-1; i++ {
		if dates[i+1].Month() == dates[i].Month() {
			continue // not a month end
		}
		k := dates[i].Year()*12 + int(dates[i].Month()) - 1
		if k < trendCorrStart {
			continue
		}
		months = append(months, k)
		corr = append(corr, pearson(rt[i-trendCorrWindow+1:i+1], re[i-trendCorrWindow+1:i+1]))
	}
	return months, corr
}

// pearson is the correlation of two equally long return samples.
func pearson(a, b []float64) float64 {
	n := float64(len(a))
	var ma, mb float64
	for i := range a {
		ma += a[i]
		mb += b[i]
	}
	ma, mb = ma/n, mb/n
	var sab, sa, sb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		sab += da * db
		sa += da * da
		sb += db * db
	}
	return sab / math.Sqrt(sa*sb)
}

// The plate's 306 monthly readings are frozen literals; recompute every one of
// them from pkg/datasets and fail on any drift.
func TestTrendCorrMatchesTheRecord(t *testing.T) {
	months, corr := trendCorrRecord(t)
	if got, want := len(corr), len(trendCorrPoints); got != want {
		t.Fatalf("the record closes %d twelve-month windows, the plate freezes %d", got, want)
	}
	if months[0] != trendCorrStart {
		t.Fatalf("the first plotted window closes at month key %d, the plate assumes %d (January 2001)", months[0], trendCorrStart)
	}
	for i := 1; i < len(months); i++ {
		if months[i] != months[i-1]+1 {
			t.Fatalf("the record skips a month at index %d (key %d after %d)", i, months[i], months[i-1])
		}
	}
	for i, want := range trendCorrPoints {
		if math.Abs(corr[i]-want) > 0.005 {
			t.Errorf("window %d (month key %d): the record reads %.4f, the plate freezes %.2f", i, months[i], corr[i], want)
		}
	}
}

// The two annotated records must really be the period's extremes, and their
// dates must be the ones the plate writes out.
func TestTrendCorrExtremesAreTheRealOnes(t *testing.T) {
	months, corr := trendCorrRecord(t)
	lo, hi := 0, 0
	for i, v := range corr {
		if v < corr[lo] {
			lo = i
		}
		if v > corr[hi] {
			hi = i
		}
	}
	if lo != trendCorrMinIdx {
		t.Errorf("the record bottoms at index %d (month key %d, %.4f), the plate annotates index %d",
			lo, months[lo], corr[lo], trendCorrMinIdx)
	}
	if hi != trendCorrMaxIdx {
		t.Errorf("the record peaks at index %d (month key %d, %.4f), the plate annotates index %d",
			hi, months[hi], corr[hi], trendCorrMaxIdx)
	}
	// The labels name April 2003 and January 2018.
	for _, c := range []struct {
		idx        int
		year       int
		month      time.Month
		label, why string
	}{
		{trendCorrMinIdx, 2003, time.April, "avril 2003", "trough"},
		{trendCorrMaxIdx, 2018, time.January, "janvier 2018", "peak"},
	} {
		k := months[c.idx]
		if got, want := k, c.year*12+int(c.month)-1; got != want {
			t.Errorf("the %s sits at month key %d, the plate labels it %s", c.why, got, c.label)
		}
		if !strings.Contains(figTrendCorrelation(), c.label) {
			t.Errorf("the plate no longer names %s", c.label)
		}
	}
}

// The plate's two headline numbers, the period mean and the share of months
// outside the flat band, must be what the record says once rounded the way the
// plate prints them.
func TestTrendCorrHeadlineNumbers(t *testing.T) {
	_, corr := trendCorrRecord(t)
	var sum, outside float64
	for _, v := range corr {
		sum += v
		if math.Abs(v) > trendCorrBand {
			outside++
		}
	}
	mean := sum / float64(len(corr))
	if got, want := math.Round(mean*100), math.Round(trendCorrMean()*100); got != want {
		t.Errorf("the record's mean rounds to %+.2f, the plate prints %+.2f", got/100, want/100)
	}
	// The plate freezes hundredths, so a month sitting on the band edge can
	// fall either side of it: one point of slack on the share, no more.
	share := outside / float64(len(corr)) * 100
	if got, want := math.Round(share), math.Round(trendCorrOutside()); math.Abs(got-want) > 1 {
		t.Errorf("the record leaves %.0f %% of the months outside ±%.1f, the plate prints %.0f %%", got, trendCorrBand, want)
	}
	// The article's own two examples must hold in the record it is drawn from:
	// short equities at the end of 2008, long equities at the end of 2021.
	if v := trendCorrPoints[(2008-2001)*12+11]; v > -0.5 {
		t.Errorf("end of 2008 reads %+.2f: the article claims a clearly negative correlation", v)
	}
	if v := trendCorrPoints[(2021-2001)*12+11]; v < 0.5 {
		t.Errorf("end of 2021 reads %+.2f: the article claims a clearly positive correlation", v)
	}
}

// House rules on the rendered plate: no rgba, no opacity, no em-dash, no rotated
// label, and nothing running past the frozen viewBox.
func TestTrendCorrPlateRespectsTheHouseRules(t *testing.T) {
	svg := figTrendCorrelation()
	if !strings.Contains(svg, `viewBox="0 0 640 384"`) {
		t.Error("the plate's viewBox moved: recheck the rendered PNG before freezing it")
	}
	for _, banned := range []string{"rgba(", "opacity", "—", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, banned in the book's figures", banned)
		}
	}
	for _, want := range []string{"MANAGED FUTURES", "moyenne", "acheteur d'actions", "vendeur d'actions", "SG Trend Index"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate no longer carries %q", want)
		}
	}
}
