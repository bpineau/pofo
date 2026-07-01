package golden

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// These golden tests validate the bundled long backcast series
// (pkg/datasets/refdata/) themselves, complementing the SPY/URTH fixtures that
// validate the daily computations. The references were cross-checked against
// public sources on 2026-07-01; each series is asserted over several
// year-to-year windows.
//
// The refdata series are MONTHLY and dated first-of-month, where a point dated
// YYYY-MM-01 is the close AT THE END of month MM (the base 1969-12-01 of MSCI
// World is 10000, the following 1970-01-01 already carries January's return).
// So a published calendar-decade return (e.g. "the 2010s") is a December-to-
// December compounding: refWindow selects Dec(y0)..Dec(y1) inclusive.
//
// Only CAGR and MaxDrawdown are asserted here: metrics.Compute annualizes
// volatility with sqrt(252) (daily), which is meaningless on monthly points, so
// the volatility/Sharpe/Sortino conventions stay validated by the daily
// fixtures. Where a vol sanity band is useful it is computed as sigma_m*sqrt(12)
// directly.

// refFinding documents a discrepancy this validation pass surfaced.
//
// MSCIWORLD-USD is the MSCI World NET total-return index, not gross as its CSV
// header and the simgen recipe Method strings claim: its Dec2012->Dec2024 CAGR
// is 10.82%/yr and Dec2014->Dec2024 is 9.95%/yr, which match MSCI World NET USD
// exactly, whereas the official GROSS figures are 11.41%/yr and 10.52%/yr (the
// gross-vs-net withholding drag). This is the correct proxy for an Irish-
// domiciled UCITS World ETF (IWDA/URTH are benchmarked against the net index),
// and the recipe only deducts the TER on top, so the reconstruction is right;
// the "gross" wording in the labels was the only error and has been corrected.

func loadRefdata(t *testing.T, id string) *marketdata.Series {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil || !ok {
		t.Fatalf("refdata %s: ok=%v err=%v", id, ok, err)
	}
	return s
}

// refWindow returns the monthly points from the December of y0 (inclusive) to
// the December of y1 (inclusive), i.e. a (y1-y0)-year December-to-December span.
func refWindow(t *testing.T, s *marketdata.Series, y0, y1 int) ([]time.Time, []float64) {
	t.Helper()
	var dates []time.Time
	var values []float64
	for _, p := range s.Points {
		afterStart := p.Date.Year() > y0 || (p.Date.Year() == y0 && p.Date.Month() == 12)
		beforeEnd := p.Date.Year() <= y1
		if afterStart && beforeEnd {
			dates = append(dates, p.Date)
			values = append(values, p.Close)
		}
	}
	if len(values) < 12 {
		t.Fatalf("refWindow %d..%d: only %d points", y0, y1, len(values))
	}
	return dates, values
}

// monthlyVol is the annualized volatility of monthly returns (sigma_m*sqrt(12)),
// the meaningful figure for these monthly series.
func monthlyVol(values []float64) float64 {
	r := metrics.Returns(values)
	if len(r) < 2 {
		return math.NaN()
	}
	m := metrics.Mean(r)
	var v float64
	for _, x := range r {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v/float64(len(r)-1)) * math.Sqrt(12)
}

type refCase struct {
	name       string
	y0, y1     int
	cagr, ctol float64 // reference CAGR (%) and tolerance
	minDD      float64 // MaxDrawdown must be at least this deep (more negative), 0 to skip
	volLo      float64 // monthly-annualized vol lower/upper sanity band (%), 0 to skip
	volHi      float64
}

func runRefCases(t *testing.T, s *marketdata.Series, cases []refCase) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dates, values := refWindow(t, s, c.y0, c.y1)
			st, err := metrics.Compute(dates, values)
			if err != nil {
				t.Fatal(err)
			}
			within(t, "CAGR", st.CAGR*100, c.cagr, c.ctol)
			if c.minDD != 0 && st.MaxDrawdown*100 > c.minDD {
				t.Errorf("MaxDrawdown = %.1f %%, expected at least %.1f %% deep", st.MaxDrawdown*100, c.minDD)
			}
			if c.volHi != 0 {
				if v := monthlyVol(values) * 100; v < c.volLo || v > c.volHi {
					t.Errorf("monthly vol = %.1f %%, expected within [%.1f, %.1f]", v, c.volLo, c.volHi)
				}
			}
		})
	}
}

// TestGoldenSP500 validates SP500-USD (S&P 500 total return, Shiller/Cowles
// reconstruction reinvested monthly) against the widely published S&P 500 TR
// decade returns (nominal, dividends reinvested), e.g. S&P Dow Jones / dqydj:
// 1970s 5.9 %, 1980s 17.6 %, 1990s 18.2 %, 2000s -0.9 %, 2010s 13.6 %; the
// reconstruction sits within ~0.35 point of these.
func TestGoldenSP500(t *testing.T) {
	s := loadRefdata(t, "SP500-USD")
	runRefCases(t, s, []refCase{
		{name: "1970s", y0: 1969, y1: 1979, cagr: 5.9, ctol: 0.5, volLo: 11, volHi: 16},
		{name: "1980s", y0: 1979, y1: 1989, cagr: 17.6, ctol: 0.6},
		{name: "1990s", y0: 1989, y1: 1999, cagr: 18.2, ctol: 0.5},
		{name: "2000s (lost decade)", y0: 1999, y1: 2009, cagr: -0.9, ctol: 0.5, minDD: -46},
		{name: "2010s", y0: 2009, y1: 2019, cagr: 13.6, ctol: 0.5},
		// Long run: since 1928 the S&P TR is ~10 %/yr nominal; this shorter,
		// stronger 1971-2024 span runs a touch above at ~11 %.
		{name: "1971-2024", y0: 1971, y1: 2024, cagr: 11.0, ctol: 0.5, volLo: 11, volHi: 15},
	})
}

// TestGoldenMSCIWorld validates MSCIWORLD-USD as the MSCI World NET total-return
// index (see refFinding): 2013-2024 net 10.82 %/yr and 2015-2024 net 9.95 %/yr
// (gross was 11.41 % and 10.52 %); the 2000s net was ~-0.2 %/yr.
func TestGoldenMSCIWorld(t *testing.T) {
	s := loadRefdata(t, "MSCIWORLD-USD")
	runRefCases(t, s, []refCase{
		{name: "2013-2024 (net)", y0: 2012, y1: 2024, cagr: 10.82, ctol: 0.4, volLo: 12, volHi: 17},
		{name: "2015-2024 (net)", y0: 2014, y1: 2024, cagr: 9.95, ctol: 0.4},
		{name: "2000s (net)", y0: 1999, y1: 2009, cagr: -0.2, ctol: 0.6, minDD: -50},
		{name: "1971-2024 (net)", y0: 1971, y1: 2024, cagr: 9.1, ctol: 0.5, volLo: 13, volHi: 17},
	})
}

// TestGoldenGold validates XAUUSD-LBMA (London/LBMA gold fix, monthly) against
// published nominal gold returns. The datahub series is a monthly average, so
// the volatile 1979/1980 boundary reads a point or two below the year-end fixes;
// the modern and long windows are validated where that averaging washes out.
// Gold nominal: 2000s ~14-15 %/yr, since 1971 ~8 %/yr.
func TestGoldenGold(t *testing.T) {
	s := loadRefdata(t, "XAUUSD-LBMA")
	runRefCases(t, s, []refCase{
		{name: "2000s", y0: 1999, y1: 2009, cagr: 14.9, ctol: 1.0},
		{name: "2000-2020", y0: 1999, y1: 2020, cagr: 9.4, ctol: 1.0},
		{name: "1971-2024", y0: 1971, y1: 2024, cagr: 8.0, ctol: 1.0, minDD: -55},
	})
}
