package golden

import (
	"math"
	"testing"
	"time"
)

// The insurance-linked reference, ILS-NET-USD, is the monthly ILS Advisers Fund
// Index (cmd/gen-catbond-refdata). These goldens hold it to three independent
// external records, so a regeneration that quietly changed units, scale or
// meaning fails here rather than in a portfolio five layers up:
//
//  1. the publisher's own summary statistics, printed next to the table it
//     serves (read 2026-08-17: maximum drawdown -12.50 %, best month +2.26 %,
//     worst month -8.61 %, annualized standard deviation 3.47 %, 87.04 % of
//     months positive);
//  2. the catastrophe record itself, which is what an ILS series IS. Every
//     large loss month must be the month of a named event, in the right order:
//     Harvey/Irma/Maria (2017-09) deepest, hurricane Ian (2022-09) second, the
//     California wildfires and Michael (2018-11) third, and the Lehman
//     collateral shock (2008-10) fourth;
//  3. the Swiss Re Global Cat Bond Total Return Index, the market's own
//     reference, whose calendar-year returns 2006-2013 are published in its
//     methodology paper (2002 to 2013: 8.77, 7.11, 6.56, 1.62, 12.02, 15.43,
//     2.45, 13.39, 11.13, 3.73, 10.28, 10.84 %). A FUND index must sit below
//     a MARKET index in every ordinary year, by the constituents' fees plus
//     the cash they hold, and must fall FURTHER in the years the market itself
//     lost ground. The test asserts that relationship rather than equality:
//     the two are different objects and equality would be the bug.
//
// One published figure of the source does NOT reconcile with its own table,
// the since-inception "Annualised Return" (5.09 %/yr published, 4.26 %/yr
// compounded from the months). Every other published figure agrees to the last
// decimal, including the trailing 3-month, 3-year and 5-year returns computed
// from the same months, so the table is what ships and the headline is treated
// as stale. See cmd/gen-catbond-refdata, which logs the divergence on every run.

// swissReCatBondTR is the Swiss Re Global Cat Bond Total Return Index calendar
// year, in percent, as published in the index methodology paper (Swiss Re
// Capital Markets, 2014-08, appendix "Index Annual Total Return"). It is the
// market reference the bundled FUND index is judged against, never spliced.
var swissReCatBondTR = map[int]float64{
	2006: 12.02, 2007: 15.43, 2008: 2.45, 2009: 13.39,
	2010: 11.13, 2011: 3.73, 2012: 10.28, 2013: 10.84,
}

func TestGoldenILSPublishedStatistics(t *testing.T) {
	rets := monthlyRets(t)
	if n := len(rets); n < 240 {
		t.Fatalf("%d monthly returns, want at least 240 (2006-01 onward)", n)
	}
	var worst, best, level, peak, drawdown float64
	var positive int
	level, peak = 1, 1
	for _, r := range rets {
		worst, best = math.Min(worst, r.ret), math.Max(best, r.ret)
		level *= 1 + r.ret
		peak = math.Max(peak, level)
		drawdown = math.Min(drawdown, level/peak-1)
		if r.ret > 0 {
			positive++
		}
	}
	years := float64(len(rets)) / 12
	for _, c := range []struct {
		name      string
		got, want float64
		tol       float64
	}{
		{"maximum drawdown", drawdown * 100, -12.50, 0.05},
		{"worst month", worst * 100, -8.61, 0.02},
		{"best month", best * 100, 2.26, 0.02},
		{"annualized volatility", stdevOf(rets) * math.Sqrt(12) * 100, 3.47, 0.20},
		{"share of positive months", float64(positive) / float64(len(rets)) * 100, 87.04, 1.0},
		{"CAGR", (math.Pow(level, 1/years) - 1) * 100, 4.3, 0.4},
	} {
		if math.Abs(c.got-c.want) > c.tol {
			t.Errorf("%s = %.2f, published %.2f (tolerance %.2f)", c.name, c.got, c.want, c.tol)
		}
	}
}

func TestGoldenILSCatastropheRecord(t *testing.T) {
	rets := monthlyRets(t)
	worst := append([]monthReturn(nil), rets...)
	for i := range worst { // simple selection of the four deepest months
		for j := i + 1; j < len(worst); j++ {
			if worst[j].ret < worst[i].ret {
				worst[i], worst[j] = worst[j], worst[i]
			}
		}
	}
	want := []struct {
		month string
		event string
	}{
		{"2017-09", "hurricanes Harvey, Irma and Maria"},
		{"2022-09", "hurricane Ian"},
		{"2018-11", "the California wildfires and hurricane Michael"},
		{"2008-10", "the Lehman collateral shock"},
	}
	for i, w := range want {
		if got := worst[i].date.Format("2006-01"); got != w.month {
			t.Errorf("loss %d of the record is %s (%.2f %%), want %s, %s",
				i+1, got, worst[i].ret*100, w.month, w.event)
		}
	}
	if worst[0].ret > -0.05 {
		t.Errorf("deepest month %.2f %%: an ILS index whose worst month is shallower than 5 %% is not this asset class", worst[0].ret*100)
	}
}

func TestGoldenILSAgainstSwissRe(t *testing.T) {
	rets := monthlyRets(t)
	year := map[int]float64{}
	for _, r := range rets {
		y := r.date.Year()
		if _, seen := year[y]; !seen {
			year[y] = 1
		}
		year[y] *= 1 + r.ret
	}
	for y, market := range swissReCatBondTR {
		fund := (year[y] - 1) * 100
		gap := market - fund
		if gap < 0.5 || gap > 8 {
			t.Errorf("%d: fund index %.2f %% against the Swiss Re market index %.2f %%, gap %.2f points outside [0.5, 8]",
				y, fund, market, gap)
		}
	}
	// 2008 is the year the relationship must invert in size: the market index
	// still gained 2.45 % while the funds lost 5.28 %, the collateral shock
	// falling on the vehicles rather than on the bonds' own risk.
	if fund := (year[2008] - 1) * 100; fund > -3 {
		t.Errorf("2008 = %.2f %%, want a loss of at least 3 points (the collateral shock)", fund)
	}
}

// monthReturn is one month of the reference: its month-end date and its return.
type monthReturn struct {
	date time.Time
	ret  float64
}

func monthlyRets(t *testing.T) []monthReturn {
	t.Helper()
	s := loadRefdata(t, "ILS-NET-USD")
	out := make([]monthReturn, 0, len(s.Points))
	for i := 1; i < len(s.Points); i++ {
		if s.Points[i-1].Close <= 0 {
			t.Fatalf("non-positive level at %s", s.Points[i-1].Date)
		}
		out = append(out, monthReturn{s.Points[i].Date, s.Points[i].Close/s.Points[i-1].Close - 1})
	}
	return out
}

func stdevOf(rets []monthReturn) float64 {
	var m float64
	for _, r := range rets {
		m += r.ret
	}
	m /= float64(len(rets))
	var s float64
	for _, r := range rets {
		s += (r.ret - m) * (r.ret - m)
	}
	return math.Sqrt(s / float64(len(rets)-1))
}
