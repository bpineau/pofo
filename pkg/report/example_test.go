package report_test

import (
	"fmt"
	"strings"

	"github.com/bpineau/pofo/pkg/report"
)

// Render assembles a self-contained HTML document from a Page model. The
// caller computes everything (cells, best flags, SVG charts); the package
// only renders. Mark the winning cell of each row with Best.
func ExampleRender() {
	page := &report.Page{
		Title:          "Portfolios: A, B",
		GeneratedAt:    "2026-01-01 12:00",
		RebalanceDays:  90,
		CommonStart:    "2010-01-04",
		CommonEnd:      "2026-01-01",
		PortfolioNames: []string{"A", "B"},
		Portfolios: []report.PortfolioSection{
			{Name: "A", ChartSVG: "<svg></svg>", Assets: []report.AssetRow{{
				Weight: "100 %", ID: "VOO", Symbol: "VOO",
				Name: "Vanguard S&P 500 ETF", UCITS: "no", Fees: "0.03 %",
				Currency: "USD", History: "2010-09-09 → 2026-01-01",
			}}},
		},
		StatRows: []report.StatRow{
			{Label: "CAGR (annualized return)", Cells: []report.StatCell{
				{Text: "10.0 %", Best: true}, {Text: "8.0 %"},
			}},
		},
		Footnotes: []string{"Adjusted closes, dividends reinvested."},
	}
	var b strings.Builder
	if err := report.Render(&b, page); err != nil {
		panic(err)
	}
	html := b.String()
	fmt.Println(strings.Contains(html, `<details class="pf">`)) // folded sections
	fmt.Println(strings.Contains(html, `class="n best"`))       // highlighted best cell
	// Output:
	// true
	// true
}

// RenderText writes the same summary for a terminal; color=false uses a *
// marker for best cells instead of ANSI green.
func ExampleRenderText() {
	page := &report.Page{
		Title:          "Portfolios: A, B",
		CommonStart:    "2010-01-04",
		CommonEnd:      "2026-01-01",
		PortfolioNames: []string{"A", "B"},
		StatRows: []report.StatRow{
			{Label: "Sharpe", Cells: []report.StatCell{{Text: "0.91", Best: true}, {Text: "0.74"}}},
		},
	}
	var b strings.Builder
	if err := report.RenderText(&b, page, false); err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(b.String(), "*0.91"))
	// Output:
	// true
}
