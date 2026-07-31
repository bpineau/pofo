package compare

import (
	"bytes"
	"math"
	"os"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
)

// goldenMonths returns n month-spaced UTC midnights from 2020-01-01. Each
// falls in its own calendar month, so SimResult.MonthlyContributions folds
// one day per month.
func goldenMonths(n int) []time.Time {
	out := make([]time.Time, n)
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range out {
		out[i] = d
		d = d.AddDate(0, 1, 0)
	}
	return out
}

// round4 rounds to four decimals, matching the fixture formulas exactly so
// the frozen bytes are reproducible across the code move.
func round4(x float64) float64 { return math.Round(x*10000) / 10000 }

// goldenSeries builds a deterministic price series over the given dates.
func goldenSeries(symbol, name, currency string, dates []time.Time, value func(i int) float64) *marketdata.Series {
	pts := make([]marketdata.Point, len(dates))
	for i, d := range dates {
		pts[i] = marketdata.Point{Date: d, Close: value(i)}
	}
	return &marketdata.Series{
		Symbol:   symbol,
		Name:     name,
		Currency: currency,
		Source:   "yahoo",
		Points:   pts,
	}
}

// fabricatedColumns builds the deterministic two-portfolio comparison the
// golden renders, translated field-for-field from cmd/pofo's Task 1 fixture
// (result -> column). NO fetching: prices, index and per-asset contributions
// are closed-form, so the frozen bytes match byte-for-byte across the move.
//
// Fixture:
//   - 30 month-spaced points from 2020-01-01 (29 folded months, past the
//     24-month floor the contribution charts need).
//   - Portfolio "Alpha": world equity 60% / gold 40%, index[i] = 100 + i.
//   - Portfolio "Beta":  US equity 70% / cash 30%, index[i] = 100 + round(sin(i/3)*8, 4).
//   - Per-asset monthly contribution for asset j: round(0.1*(j+1), 4),
//     constant across months (0.1 for the first holding, 0.2 for the second).
func fabricatedColumns(t *testing.T) (columns []*column, bench *marketdata.Series, start, end time.Time, meta map[string]suggest.Meta) {
	t.Helper()

	meta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		t.Fatalf("LoadMeta: %v", err)
	}

	const n = 30
	dates := goldenMonths(n)
	start, end = dates[0], dates[n-1]

	// index[i] value functions for the two portfolios.
	alphaIndex := func(i int) float64 { return 100 + float64(i) }
	betaIndex := func(i int) float64 { return 100 + round4(math.Sin(float64(i)/3)*8) }

	// contributions builds the [asset][day] contribution matrix: day 0 is
	// always zero, every later day carries the constant per-asset value, so
	// each folded month reads round(0.1*(j+1), 4).
	contributions := func(nAssets int) [][]float64 {
		c := make([][]float64, nAssets)
		for j := range c {
			c[j] = make([]float64, n)
			for k := 1; k < n; k++ {
				c[j][k] = round4(0.1 * float64(j+1))
			}
		}
		return c
	}

	build := func(name, currency string, color string, assets []portfolio.Asset, index func(i int) float64) *column {
		values := make([]float64, n)
		for i := range values {
			values[i] = index(i)
		}
		sim := &portfolio.SimResult{
			Dates:         dates,
			Values:        values,
			Index:         values,
			Contributions: contributions(len(assets)),
		}
		stats, err := metrics.Compute(dates, values)
		if err != nil {
			t.Fatalf("Compute(%s): %v", name, err)
		}
		return &column{
			p:             &portfolio.Portfolio{Name: name, Assets: assets},
			sim:           sim,
			color:         color,
			rebalanceDays: 90,
			currency:      currency,
			specName:      name,
			winDates:      dates,
			winValues:     values,
			stats:         stats,
		}
	}

	asset := func(id, name, currency string, weight, fees float64) portfolio.Asset {
		return portfolio.Asset{
			ID:     id,
			Symbol: id,
			Name:   name,
			Weight: weight,
			Fees:   fees,
			Series: goldenSeries(id, name, currency, dates, func(i int) float64 { return 100 + float64(i) }),
		}
	}

	alpha := build("Alpha", "EUR", chart.PaletteColor(0), []portfolio.Asset{
		asset("IE000EGGFVG6", "Dimensional Global Core Equity", "USD", 0.6, 0.30),
		asset("IGLN", "iShares Physical Gold", "USD", 0.4, 0.12),
	}, alphaIndex)
	beta := build("Beta", "EUR", chart.PaletteColor(1), []portfolio.Asset{
		asset("IE00B8GKDB10", "Vanguard FTSE All-World High Dividend", "USD", 0.7, 0.29),
		asset("IB01", "iShares 0-3 Month Treasury Bond", "USD", 0.3, 0.07),
	}, betaIndex)

	bench = goldenSeries("^GSPC", "S&P 500", "USD", dates, func(i int) float64 { return 100 + float64(i)*0.5 })

	return []*column{alpha, beta}, bench, start, end, meta
}

// TestReportGolden carries Task 1's characterization golden forward: the same
// fabricated comparison, now rendered through Comparison.HTMLPage, must
// reproduce the frozen bytes exactly. If it fails, the HTMLPage port diverged
// from the original buildPage; fix the port, do NOT regenerate the golden.
func TestReportGolden(t *testing.T) {
	cols, bench, start, end, meta := fabricatedColumns(t)
	opt := Options{Rebalance: 90, Benchmark: "^GSPC", Framework: suggest.RegimeFramework(), Currency: "EUR"}
	c := newTestComparison(cols, bench, start, end, meta, opt)
	page := c.HTMLPage(Decoration{})
	page.GeneratedAt = "GOLDEN"
	var buf bytes.Buffer
	if err := report.Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	const path = "testdata/report-golden.html"
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("report output changed; if intentional re-run with UPDATE_GOLDEN=1 (it must NOT change during the pkg/compare move)")
	}
}
