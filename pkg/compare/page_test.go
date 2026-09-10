package compare

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
)

// statColumn builds a minimal but complete column for the statistics table:
// a straight-line index over n month-spaced points, its computed stats, and
// the portfolio the money rows read.
func statColumn(t *testing.T, name string, n int, p *portfolio.Portfolio, sim *portfolio.SimResult) *column {
	t.Helper()
	dates := months(n)
	values := make([]float64, n)
	for i := range values {
		values[i] = 100 * math.Pow(1.005, float64(i))
	}
	if sim == nil {
		sim = &portfolio.SimResult{}
	}
	sim.Dates, sim.Index = dates, values
	if sim.Values == nil {
		sim.Values = values
	}
	stats, err := metrics.Compute(dates, values)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		p = &portfolio.Portfolio{}
	}
	p.Name = name
	return &column{
		p: p, sim: sim, color: "#1f6f78", rebalanceDays: 90, currency: "EUR",
		specName: name, winDates: dates, winValues: values, stats: stats,
	}
}

// row finds a statistics row by label prefix; the label carries the benchmark
// name for the relative rows, hence the prefix match.
func row(t *testing.T, rows []report.StatRow, label string) report.StatRow {
	t.Helper()
	for _, r := range rows {
		if strings.HasPrefix(r.Label, label) {
			return r
		}
	}
	t.Fatalf("no row %q in %d rows", label, len(rows))
	return report.StatRow{}
}

func hasRow(rows []report.StatRow, label string) bool {
	for _, r := range rows {
		if strings.HasPrefix(r.Label, label) {
			return true
		}
	}
	return false
}

// The money rows describe the whole simulated span and follow the money, so
// they only make sense for a portfolio that declares a starting capital. They
// appear as soon as ONE column declares one, and the columns that do not show
// a dash rather than a fabricated zero.
func TestStatRowsMoneyBlockAppearsWithCapital(t *testing.T) {
	const n = 40
	dates := months(n)
	funded := statColumn(t, "Funded", n, &portfolio.Portfolio{Capital: 100000},
		&portfolio.SimResult{
			Values:      []float64{},
			Contributed: 12000,
			Withdrawn:   3000,
			FlowDates:   []time.Time{dates[10], dates[20]},
			FlowAmounts: []float64{12000, -3000},
		})
	// Values must follow the money, not the index: the final value is a money
	// row.
	funded.sim.Values = make([]float64, n)
	for i := range funded.sim.Values {
		funded.sim.Values[i] = 100000 * math.Pow(1.005, float64(i))
	}
	plain := statColumn(t, "Plain", n, nil, nil)

	rows := buildStatRows([]*column{funded, plain}, "")
	for _, label := range []string{"Starting capital", "Total contributed", "Total withdrawn", "Final value", "IRR"} {
		r := row(t, rows, label)
		if r.Cells[1].Text != "-" {
			t.Errorf("row %q for the capital-less column = %q, want \"-\"", label, r.Cells[1].Text)
		}
	}
	if got := row(t, rows, "Starting capital").Cells[0].Text; got != "100 000" {
		t.Errorf("starting capital = %q, want a thin-space grouped 100 000", got)
	}
	if got := row(t, rows, "Total withdrawn").Cells[0].Text; got != "3 000" {
		t.Errorf("total withdrawn = %q, want 3 000", got)
	}
	irr := row(t, rows, "IRR").Cells[0].Text
	if irr == "-" || !strings.HasSuffix(irr, "%") {
		t.Errorf("IRR = %q, want a percentage", irr)
	}
	// Without any capital anywhere, the block is absent entirely.
	if rows := buildStatRows([]*column{plain}, ""); hasRow(rows, "Starting capital") {
		t.Error("money rows rendered for a portfolio with no capital")
	}
}

// The benchmark rows are always laid out (the table's shape must not depend on
// the data) but read "-" until the column actually carries a relative
// measurement, and the benchmark's name is in the label so the reader knows
// what the alpha is against.
func TestStatRowsBenchmarkColumns(t *testing.T) {
	const n = 40
	bare := statColumn(t, "Bare", n, nil, nil)
	rich := statColumn(t, "Rich", n, nil, nil)
	rich.hasRel, rich.rel = true, metrics.Relative{Alpha: 0.02, InfoRatio: 0.5, UpCapture: 1.1, DownCapture: 0.8}
	rich.stats.HasBeta, rich.stats.Beta = true, 0.85
	rich.stats.HasCWARP, rich.stats.CWARP = true, 12.5
	rich.hasVTS, rich.vts = true, metrics.VolTermStructure{MonthlyVol: 0.08, Ratio: 0.9, MonthlySharpe: 0.7, MonthlySortino: 1.1}
	rich.hasReal, rich.realStats = true, metrics.Stats{MaxDrawdown: -0.31, TTRDays: 800}

	rows := buildStatRows([]*column{bare, rich}, "^GSPC")
	if got := row(t, rows, "Alpha").Label; !strings.Contains(got, "^GSPC") {
		t.Errorf("alpha label = %q, want the benchmark named", got)
	}
	for _, label := range []string{"Alpha", "Information ratio", "Up capture", "Down capture", "Beta", "CWARP"} {
		r := row(t, rows, label)
		if r.Cells[0].Text != "-" {
			t.Errorf("row %q without a benchmark measurement = %q, want \"-\"", label, r.Cells[0].Text)
		}
		if r.Cells[1].Text == "-" {
			t.Errorf("row %q for the measured column = \"-\", want a value", label)
		}
	}
	if got := row(t, rows, "CWARP").Cells[1].Text; got != "+12.5" {
		t.Errorf("CWARP = %q, want a signed +12.5", got)
	}
	// The real and monthly rows follow the same rule.
	for _, label := range []string{"Max Drawdown (real)", "TTR real", "Volatility (monthly", "Variance ratio", "Sharpe (monthly)", "Sortino (monthly)"} {
		if got := row(t, rows, label).Cells[0].Text; got != "-" {
			t.Errorf("row %q without the measurement = %q, want \"-\"", label, got)
		}
	}
	if got := row(t, rows, "TTR real").Cells[1].Text; !strings.Contains(got, "2.2 y") {
		t.Errorf("real TTR = %q, want it expressed in years past one year", got)
	}
}

// The best cell of a row is highlighted, on the right side of the comparison:
// higher CAGR wins, LOWER volatility wins, and a row nobody can win (every
// value missing) highlights nothing.
func TestStatRowsMarkBestBySide(t *testing.T) {
	const n = 40
	slow := statColumn(t, "Slow", n, nil, nil)
	fast := statColumn(t, "Fast", n, nil, nil)
	fast.stats.CAGR = slow.stats.CAGR * 2
	slow.stats.Volatility, fast.stats.Volatility = 0.08, 0.16

	rows := buildStatRows([]*column{slow, fast}, "")
	cagr := row(t, rows, "CAGR")
	if !cagr.Cells[1].Best || cagr.Cells[0].Best {
		t.Errorf("CAGR best = %v/%v, want the higher one", cagr.Cells[0].Best, cagr.Cells[1].Best)
	}
	vol := row(t, rows, "Volatility (annualized)")
	if !vol.Cells[0].Best || vol.Cells[1].Best {
		t.Errorf("volatility best = %v/%v, want the lower one", vol.Cells[0].Best, vol.Cells[1].Best)
	}
	beta := row(t, rows, "Beta")
	if beta.Cells[0].Best || beta.Cells[1].Best {
		t.Error("beta highlighted a winner: it has no better side")
	}
	// A single column is not a comparison: nothing is highlighted.
	one := buildStatRows([]*column{fast}, "")
	if row(t, one, "CAGR").Cells[0].Best {
		t.Error("a lone column was highlighted as the best of its row")
	}
}

// The fee row sums weight x TER over the lines that publish one, adds the
// envelope fee, and says "(i)" when a component is missing rather than
// pretending the total is complete.
func TestStatRowsWeightedFees(t *testing.T) {
	const n = 40
	full := statColumn(t, "Full", n, &portfolio.Portfolio{
		Assets:       []portfolio.Asset{{Weight: 0.5, Fees: 0.20}, {Weight: 0.5, Fees: 0.40}},
		EnvelopeFees: 0.50,
	}, nil)
	partial := statColumn(t, "Partial", n, &portfolio.Portfolio{
		Assets: []portfolio.Asset{{Weight: 0.5, Fees: 0.20}, {Weight: 0.5, Fees: -1}},
	}, nil)
	unknown := statColumn(t, "Unknown", n, &portfolio.Portfolio{
		Assets: []portfolio.Asset{{Weight: 1, Fees: -1}},
	}, nil)

	rows := buildStatRows([]*column{full, partial, unknown}, "")
	cells := row(t, rows, "Weighted ongoing charges").Cells
	if cells[0].Text != "0.80 %" {
		t.Errorf("full fees = %q, want 0.80 %% (0.5x0.20 + 0.5x0.40 + 0.50 envelope)", cells[0].Text)
	}
	if cells[1].Text != "0.10 % (i)" {
		t.Errorf("partial fees = %q, want 0.10 %% flagged incomplete", cells[1].Text)
	}
	if cells[2].Text != "-" {
		t.Errorf("all-unknown fees = %q, want \"-\"", cells[2].Text)
	}
}

// The small formatters, including the money one: a report's amounts are
// grouped with thin spaces (never a comma, which a French reader would read
// as a decimal point), and a missing number is a dash, never NaN.
func TestFormatters(t *testing.T) {
	amounts := map[float64]string{
		0: "0", 999: "999", 1000: "1 000", 1234567: "1 234 567",
		-45000: "-45 000", 1234.6: "1 235",
	}
	for v, want := range amounts {
		if got := fmtAmount(v); got != want {
			t.Errorf("fmtAmount(%v) = %q, want %q", v, got, want)
		}
	}
	if got := fmtPct(0.1234); got != "12.34 %" {
		t.Errorf("fmtPct = %q, want 12.34 %%", got)
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got := fmtPct(bad); got != "-" {
			t.Errorf("fmtPct(%v) = %q, want \"-\"", bad, got)
		}
		if got := fmtNum(bad); got != "-" {
			t.Errorf("fmtNum(%v) = %q, want \"-\"", bad, got)
		}
	}
	ttr := map[metrics.Stats]string{
		// Never underwater: a measured zero, not a missing value (see fmtTTR).
		{TTRDays: 0}:                     "0 d",
		{TTRDays: 90}:                    "90 d",
		{TTRDays: 731}:                   "2.0 y (731 d)",
		{TTRDays: 90, TTROngoing: true}:  "90 d (ongoing)",
		{TTRDays: 731, TTROngoing: true}: "2.0 y (731 d) (ongoing)",
	}
	for s, want := range ttr {
		if got := fmtTTR(s); got != want {
			t.Errorf("fmtTTR(%+v) = %q, want %q", s, got, want)
		}
	}
}

// The one starred cell that must never read "not measured". A portfolio that
// never closed below a previous peak has the best recovery time of its row,
// and the table has to SAY so with a number: a dash marked best would tell the
// reader the winner is the column with no data.
func TestStatRowsZeroTTRShowsAValueAndWins(t *testing.T) {
	const n = 40
	calm := statColumn(t, "Calm", n, nil, nil)
	calm.stats.TTRDays, calm.stats.TTROngoing = 0, false
	rough := statColumn(t, "Rough", n, nil, nil)
	rough.stats.TTRDays, rough.stats.TTROngoing = 500, false

	r := row(t, buildStatRows([]*column{calm, rough}, ""), "TTR (longest recovery)")
	if got := r.Cells[0].Text; got != "0 d" {
		t.Errorf("never-underwater TTR = %q, want \"0 d\"", got)
	}
	if !r.Cells[0].Best {
		t.Error("the never-underwater column is not marked best on TTR")
	}
	if r.Cells[1].Best {
		t.Error("the 500-day column is marked best on TTR")
	}
	// And the row that really has nothing to show keeps its dash, unstarred.
	real := row(t, buildStatRows([]*column{calm, rough}, ""), "TTR real")
	for i, c := range real.Cells {
		if c.Text != "-" || c.Best {
			t.Errorf("TTR real cell %d = %q best=%v, want an unstarred dash", i, c.Text, c.Best)
		}
	}
}

// foldInto and relabel are how the composition pies collapse the catalog's
// vocabulary into the report's: merging must accumulate rather than
// overwrite, and folding a key into itself must not double it.
func TestFoldIntoAndRelabel(t *testing.T) {
	agg := map[string]float64{"a": 1, "b": 2, "c": 3, "dst": 4}
	foldInto(agg, "dst", "a", "b", "missing", "dst")
	if agg["dst"] != 7 || len(agg) != 2 {
		t.Errorf("after foldInto: %v, want dst 7 and c untouched", agg)
	}
	relabel(agg, "c", "dst")
	if agg["dst"] != 10 || len(agg) != 1 {
		t.Errorf("after relabel: %v, want dst 10 alone", agg)
	}
	relabel(agg, "absent", "dst") // a no-op, never a zero entry
	if len(agg) != 1 {
		t.Errorf("relabel of an absent key created one: %v", agg)
	}
}

// The whole per-portfolio block assembly, on a fixture whose holdings the
// catalog knows and whose contributions actually move: the pies, the coverage
// bars, the risk budget and the contribution charts must all land in the
// section, each with the footnote that explains it.
func TestHTMLPageAssemblesEveryBlock(t *testing.T) {
	cols, bench, start, end, meta := fabricatedColumns(t)
	// The golden's contributions are constant, so their covariance is
	// degenerate and the risk budget declines to guess. Give them a real
	// path instead, out of phase per holding so the classes differ.
	for _, c := range cols {
		for j, series := range c.sim.Contributions {
			for k := 1; k < len(series); k++ {
				series[k] = 0.01 * math.Sin(float64(k)/3+float64(j))
			}
		}
	}
	opt := Options{Rebalance: 90, Benchmark: "^GSPC", Framework: suggest.RegimeFramework(), Currency: "EUR"}
	c := newTestComparison(cols, bench, start, end, meta, opt)

	page := c.HTMLPage(Decoration{FireHref: map[string]string{"Alpha": "/fire?p=alpha"}})
	if len(page.Portfolios) != 2 {
		t.Fatalf("sections = %d, want 2", len(page.Portfolios))
	}
	s := page.Portfolios[0]
	if len(s.Breakdowns) == 0 || len(s.Coverage) == 0 || len(s.RiskBudget) == 0 {
		t.Fatalf("blocks missing: %d pies, %d coverage bars, %d risk rows",
			len(s.Breakdowns), len(s.Coverage), len(s.RiskBudget))
	}
	if s.CoverageLabel != "Macro-regime coverage (by weight)" {
		t.Errorf("coverage label = %q under the regime framework", s.CoverageLabel)
	}
	if s.ContribSVG == "" || s.ContribMonthlySVG == "" || s.RegimeSVG == "" {
		t.Error("contribution charts missing from the section")
	}
	if s.FireHref != "/fire?p=alpha" {
		t.Errorf("FireHref = %q, want the decoration's deep link", s.FireHref)
	}
	// The risk-budget caption is what keeps the three columns from being read
	// as an efficiency ratio, so it must travel with the block.
	note := strings.Join(s.Notes, "\n")
	if !strings.Contains(note, "Risk budget over "+start.Format("2006-01-02")) {
		t.Errorf("risk-budget note missing or unwindowed: %q", note)
	}
	// One footnote per block that is present.
	foot := strings.Join(page.Footnotes, "\n")
	for _, want := range []string{"Realized contribution charts", "Composition pies", "Macro-regime coverage"} {
		if !strings.Contains(foot, want) {
			t.Errorf("footnote %q missing", want)
		}
	}
	// The factor framework renames the coverage block rather than adding one.
	optF := opt
	optF.Framework = suggest.FactorFramework()
	if got := newTestComparison(cols, bench, start, end, meta, optF).HTMLPage(Decoration{}).Portfolios[0].CoverageLabel; got != "Risk-factor coverage (by weight)" {
		t.Errorf("factor coverage label = %q", got)
	}
	// The public accessors read the same model.
	if len(c.CoverageBars(cols[0].p.Assets)) != len(s.Coverage) {
		t.Error("CoverageBars disagrees with the rendered section")
	}
	if len(c.StatRows()) == 0 {
		t.Error("StatRows returned nothing")
	}
}

// Without catalog metadata there is nothing to look through, so the
// composition blocks omit themselves (and their footnotes) instead of
// rendering a single grey "Unknown" wedge.
func TestHTMLPageOmitsBlocksWithoutMetadata(t *testing.T) {
	cols, _, start, end, _ := fabricatedColumns(t)
	c := newTestComparison(cols[:1], nil, start, end, nil,
		Options{Rebalance: 90, Framework: suggest.RegimeFramework()})
	page := c.HTMLPage(Decoration{})
	s := page.Portfolios[0]
	if len(s.Breakdowns) != 0 || len(s.Coverage) != 0 || len(s.RiskBudget) != 0 {
		t.Errorf("blocks rendered without metadata: %d/%d/%d",
			len(s.Breakdowns), len(s.Coverage), len(s.RiskBudget))
	}
	for _, f := range page.Footnotes {
		if strings.HasPrefix(f, "Composition pies") || strings.HasPrefix(f, "Macro-regime coverage") {
			t.Errorf("footnote for an absent block: %q", f)
		}
	}
	if len(page.PortfolioNames) != 1 || page.CommonStart != start.Format("2006-01-02") {
		t.Errorf("page header = %v / %q", page.PortfolioNames, page.CommonStart)
	}
}
