// HTML report assembly: the Page model (buildPage), the statistics table
// (buildStatRows) and the shared value formatting.
package main

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
	"github.com/bpineau/pofo/pkg/webui"
)

// assetCWARP formats a single holding's CWARP as a 25 % overlay on the
// benchmark, measured over the common window, or "-" when there is no
// benchmark or too little overlap.
func assetCWARP(s *marketdata.Series, benchDates []time.Time, benchValues []float64, start, end time.Time) string {
	if s == nil || len(benchDates) == 0 {
		return "-"
	}
	dates, values := seriesSlices(s)
	i, j := window(dates, start, end)
	if j-i < 2 {
		return "-"
	}
	if c, ok := metrics.CWARPvs(dates[i:j], values[i:j], benchDates, benchValues, metrics.CWARPParams{}); ok {
		return fmt.Sprintf("%+.1f", c)
	}
	return "-"
}

// siteNavCSS and siteNavHTML style and render the web app's cross-navigation
// bar at the top of the /view report; both are used only when opt.web is set.
const siteNavCSS = `.site-nav{display:flex;gap:1.2rem;align-items:baseline;max-width:min(1180px,94vw);` +
	`margin:0 auto;padding:.8rem clamp(1rem,4vw,2rem) 0;font-family:var(--mono);font-size:.72rem;` +
	`letter-spacing:.06em;text-transform:uppercase}` +
	`.site-nav a{color:var(--muted);text-decoration:none}` +
	`.site-nav a:hover{color:var(--accent-ink)}` +
	`.site-nav a:first-child{color:var(--accent-ink)}`

var siteNavHTML = template.HTML(`<nav class="site-nav">` +
	`<a href="/">Portfolios</a><a href="/fire/">FIRE simulator</a><a href="/book/fr/">FIRE book</a>` +
	`</nav>`)

func buildPage(results []*result, opt *options, bench *marketdata.Series, commonStart, commonEnd time.Time, meta map[string]suggest.Meta) *report.Page {
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.p.Name
	}
	var benchDates []time.Time
	var benchValues []float64
	if bench != nil {
		benchDates, benchValues = seriesSlices(bench)
	}
	page := &report.Page{
		Title:          "Portfolios: " + strings.Join(names, ", "),
		GeneratedAt:    time.Now().Format("2006-01-02 15:04"),
		RebalanceDays:  opt.rebalance,
		CommonStart:    commonStart.Format("2006-01-02"),
		CommonEnd:      commonEnd.Format("2006-01-02"),
		PortfolioNames: names,
	}
	// Served inside the web app: warm the report to the book identity and add
	// the site nav. The standalone CLI report leaves both empty (opt.web is
	// false), so its output is unchanged.
	if opt.web {
		page.SkinCSS = template.CSS(webui.WarmSkin + siteNavCSS)
		page.SiteNav = siteNavHTML
	}

	anySimulated := false
	for _, r := range results {
		first := r.sim.Dates[0].Format("2006-01-02")
		last := r.sim.Dates[len(r.sim.Dates)-1].Format("2006-01-02")
		// Rendered wider than the default so the full-width report shows the
		// chart at a moderate, print-like scale rather than blown up.
		svg := chart.Line(chart.Options{
			Title:  fmt.Sprintf("%s: base 100 from %s to %s", r.p.Name, first, last),
			Width:  1200,
			Height: 400,
		}, []chart.Series{{Name: r.p.Name, Dates: r.sim.Dates, Values: r.sim.Values, Color: r.color}})

		subtitle := fmt.Sprintf("%s → %s", first, last)
		if r.rebalanceDays != opt.rebalance {
			if r.rebalanceDays == 0 {
				subtitle += ", never rebalanced (#meta)"
			} else {
				subtitle += fmt.Sprintf(", rebalanced every %d d (#meta)", r.rebalanceDays)
			}
		}
		if r.p.EnvelopeFees > 0 {
			subtitle += fmt.Sprintf(", %.2f %%/yr envelope fees deducted", r.p.EnvelopeFees)
		}
		if r.p.Leverage {
			expo := 0.0
			for _, a := range r.p.Assets {
				expo += a.Weight
			}
			subtitle += fmt.Sprintf(", exposure %.4g %%, financed at cash + %.2g %%/yr (#meta leverage)", expo*100, r.p.BorrowSpread)
		}
		section := report.PortfolioSection{
			Name:     r.p.Name,
			Subtitle: subtitle,
			ChartSVG: template.HTML(svg),
			Warnings: r.p.Warnings,
		}
		if r.note != "" {
			section.Notes = []string{r.note}
		}
		section.ContribSVG, section.ContribMonthlySVG, section.RegimeSVG = contributionCharts(r)
		section.Breakdowns = breakdownPies(r.p.Assets, meta)
		if len(section.Breakdowns) > 0 {
			section.Notes = append(section.Notes, compositionNotes(r.p.Assets, meta, r.currency)...)
		}
		section.Coverage = coverageBars(r.p.Assets, meta, opt.fw)
		if len(section.Coverage) > 0 {
			section.CoverageLabel = "Macro-regime coverage (by weight)"
			if opt.fw.Name == "factors" {
				section.CoverageLabel = "Risk-factor coverage (by weight)"
			}
		}
		for _, a := range r.p.Assets {
			var notes []string
			if !a.Series.SimulatedBefore.IsZero() {
				anySimulated = true
				notes = append(notes, fmt.Sprintf("simulated before %s via %s",
					a.Series.SimulatedBefore.Format("2006-01-02"), a.Series.ProxySymbol))
			}
			switch a.Series.Source {
			case "ft", "morningstar":
				src := "Financial Times"
				if a.Series.Source == "morningstar" {
					src = "Morningstar"
				}
				note := "source: " + src + " (NAV)"
				if marketdata.LooksDistributing(a.Series.Name) {
					note += ", distributing share class: dividends not reinvested in this series"
				}
				notes = append(notes, note)
			case "stooq":
				notes = append(notes, "source: Stooq (not dividend-adjusted)")
			}
			feesText := "-"
			if a.Fees >= 0 {
				feesText = fmt.Sprintf("%.2f %%", a.Fees)
			}
			base, _ := marketdata.SplitSim(a.ID)
			ucitsText := "?"
			ucits, ucitsKnown := marketdata.GuessUCITS(base, a.Name)
			if ucitsKnown {
				ucitsText = map[bool]string{true: "yes", false: "no"}[ucits]
			}
			assetClass := ""
			if m, _, ok := metaFor(meta, a.ID); ok {
				assetClass = m.AssetClass
				// A gold ETC or a listed closed-end fund cannot be a
				// UCITS fund, yet is freely buyable by an EU retail
				// investor (PRIIPs KID). A bare "no" reads as "not
				// buyable": name the wrapper instead.
				if ucitsKnown && !ucits && m.EURetail {
					ucitsText = "no (KID)"
					notes = append(notes, "not a UCITS fund but EU-retail-buyable: an EU-listed wrapper (ETC, closed-end fund) with a PRIIPs KID")
				}
			}
			row := report.AssetRow{
				Weight:   fmt.Sprintf("%.4g %%", a.Weight*100),
				ID:       a.ID,
				Symbol:   a.Symbol,
				Name:     a.Name,
				Class:    assetClass,
				UCITS:    ucitsText,
				Fees:     feesText,
				Currency: a.Series.Currency,
				History: fmt.Sprintf("%s → %s",
					a.Series.First().Date.Format("2006-01-02"),
					a.Series.Last().Date.Format("2006-01-02")),
				CWARP: assetCWARP(a.Series, benchDates, benchValues, commonStart, commonEnd),
				Note:  strings.Join(notes, "; "),
			}
			section.Assets = append(section.Assets, row)
		}
		page.Portfolios = append(page.Portfolios, section)
	}

	// Always show a curve up top: the comparison for several portfolios, or
	// the single portfolio's own curve, so the report opens on a chart
	// whatever the number of portfolios.
	cmp := make([]chart.Series, len(results))
	for i, r := range results {
		cmp[i] = chart.Series{Name: r.p.Name, Dates: r.winDates, Values: r.winValues, Color: r.color}
	}
	title, heading := "Portfolio comparison", "Comparison"
	if len(results) == 1 {
		title, heading = results[0].p.Name, "Performance"
	}
	page.OverviewHeading = heading + ": base 100 at " + page.CommonStart
	page.CompareSVG = template.HTML(chart.Line(chart.Options{
		Title:  title + ": base 100 at " + page.CommonStart,
		Width:  1200,
		Height: 460,
	}, cmp))

	page.StatRows = buildStatRows(results, opt.benchmark)

	// Underwater plot: every portfolio's drawdown over the common period.
	uw := make([]chart.Series, len(results))
	for i, r := range results {
		dd := metrics.Drawdowns(r.winValues)
		for k := range dd {
			dd[k] *= 100
		}
		uw[i] = chart.Series{Name: r.p.Name, Dates: r.winDates, Values: dd, Color: r.color}
	}
	page.UnderwaterSVG = template.HTML(chart.Line(chart.Options{
		Title:  "Drawdowns (%), common period",
		Width:  1200,
		Height: 300,
	}, uw))

	curSet := map[string]bool{}
	var curs []string
	for _, r := range results {
		if r.currency != "" && !curSet[r.currency] {
			curSet[r.currency] = true
			curs = append(curs, r.currency)
		}
	}
	if len(curs) > 0 {
		page.Footnotes = append(page.Footnotes, fmt.Sprintf(
			"Series converted to %s (daily Yahoo FX crosses; the earliest known rate is held constant before the FX history starts). Columns tagged with a currency show the same portfolio through that currency's numeraire and CPI.", strings.Join(curs, ", ")))
	}
	page.Footnotes = append(page.Footnotes, []string{
		"Sources: Yahoo Finance (adjusted closes, dividends and splits reinvested), Financial Times and Morningstar (fund NAVs).",
		fmt.Sprintf("Simulation: base 100, rebalanced to the target weights every %d calendar days by default (overridable per portfolio via \"#meta rebalance:N\"), with no fees or taxes.", opt.rebalance),
		"Statistics computed over the period common to all portfolios; volatility and ratios annualized over 252 trading days, zero risk-free rate for Sharpe and Sortino (Curvo convention; PortfolioVisualizer/LazyPortfolio use T-bills and monthly data; their volatilities and drawdowns therefore come out lower).",
		"Fees: published TERs (FT/justETF sources), already included in prices and NAVs, informational column; only the additional portfolio fees \"#meta extra-fees:X\" (envelope, mandate…) are deducted from the simulated performance.",
		"Monthly volatility and variance ratio (Lo-MacKinlay): the monthly figure annualizes the standard deviation of month-end returns, and the ratio divides the monthly annualized variance by the daily one. It exposes the autocorrelation the single-frequency stats hide: ≈1 means returns are serially uncorrelated (daily vol is faithful), below 1 means they mean-revert (daily vol overstates the risk realized over months), above 1 means they trend (daily vol understates it). Read it as complementary to the rolling-CAGR and drawdown columns, and note the small-sample caveat: a month-end series holds only ~12 points per year, so over short common periods the monthly figures are noisier point estimates than the daily ones.",
		"Max Drawdown, Ulcer and TTR on daily closes, harsher than monthly-step references (e.g. COVID 2020: −33.7 % daily, −20 % on monthly closes).",
		"TTR: duration of the longest stretch spent below a previous peak (peak to recovery).",
		"Real Max Drawdown / TTR real: the same measured on the inflation-adjusted series (nominal deflated by French HICP ^HICP-FR for EUR reports, by the US CPI ^CPI-US for USD ones), i.e. in purchasing power. Inflation deepens drawdowns and lengthens recoveries; the nominal figures understate the pain a spender actually feels.",
	}...)
	if anySimulated {
		page.Footnotes = append(page.Footnotes,
			"Histories extended before some funds' inception: via a proxy (older indices or funds; price indices do not include dividends) or via permanent simulated data (pkg/datasets/simdata/<id>.csv files generated by -gen-simdata, methodology and replication quality at the top of each file).")
	}
	if bench != nil {
		page.Footnotes = append(page.Footnotes,
			"Beta: regression of daily returns against "+bench.Symbol+" over the common window.",
			"Information ratio: average active return (portfolio − benchmark) divided by its tracking error (the volatility of that active return), showing how much benchmark-beating return is earned per unit of benchmark-relative risk. Higher is better; above ~0.5 is good, negative means the active bets cost return.",
			"Up / Down capture: the portfolio's average return on the benchmark's up (resp. down) days, as a % of the benchmark's own average on those days. Up capture above 100 % amplifies rallies; Down capture below 100 % cushions losses. The ideal profile is high up / low down (e.g. 95 % / 70 %).",
			"CWARP (Cole Wins Above Replacement Portfolio, Artemis Capital): the geometric average of the improvements a 25 %-of-notional overlay makes to the benchmark's Sortino ratio and return-to-max-drawdown, in percent (positive helps, negative hurts). Unlike Sharpe it rewards non-correlation and skew, since both denominators are measured on the combined series. The statistics row scores the whole portfolio as the overlay; the per-holding CWARP column scores each sleeve on its own, revealing which ones actually diversify "+bench.Symbol+" (typically gold, long duration and trend, not more equity).")
	}
	var hasBreakdowns, hasCoverage, hasContrib bool
	for _, s := range page.Portfolios {
		hasBreakdowns = hasBreakdowns || len(s.Breakdowns) > 0
		hasCoverage = hasCoverage || len(s.Coverage) > 0
		hasContrib = hasContrib || s.ContribSVG != "" || s.RegimeSVG != ""
	}
	if hasContrib {
		page.Footnotes = append(page.Footnotes,
			"Realized contribution charts (per portfolio): each day's portfolio return is decomposed as held weight × asset return. The timeline stacks each holding's contribution around zero (bands above zero carried the period, bands below cost it; the black line is the portfolio's own return, the net of the bands); hover for exact figures. The 12m-rolling window reads regimes and trends but nets a crash against the year before it; switch to the monthly window for the anatomy of a single month (e.g. who drove and who cushioned March 2020). The per-regime matrix groups the same monthly contributions by macro quadrant, annualized: it is the empirical mirror of the coverage bars (who actually delivered, vs who was supposed to). Regimes come from the embedded OECD panel (share of countries with accelerating industrial production × accelerating inflation, thresholded at one half), forward-filled at the panel's edges; contributions before a fund's listing read its backcast, and envelope fees (when any) are not attributed to holdings.")
	}
	if hasBreakdowns {
		page.Footnotes = append(page.Footnotes,
			"Composition pies (per portfolio), look-through: stacked funds are opened into their legs for the asset-type pie (shares of total economic exposure, so a 90/60 fund counts as equity plus bonds); the sector pie covers the equity sleeve only; currency exposure is derived from geography, denomination and share-class hedging, never the quote currency (a EUR-quoted world tracker is mostly USD), with gold and commodities counted as non-fiat (\"None\") and futures books as \"Dynamic\". \"No country\" collects assets for which a country split is meaningless (gold, trend…), unlike \"Other\", which aggregates small real positions.")
	}
	if hasCoverage {
		page.Footnotes = append(page.Footnotes,
			"Macro-regime coverage: notional exposure to each growth/inflation environment (an asset can span several; leveraged stacked funds count each leg's notional, so bars can exceed 100%); a low bar is a gap. Run \"-suggest\" for assets to fill it. Each bar is split by contributing holding, one stable color per holding across the rows (hover a segment for its share); the line beneath lists the contributions in points of notional weight.")
	}
	return page
}

func buildStatRows(results []*result, benchmark string) []report.StatRow {
	// cell computes a row's value (for the best-of-row comparison) and its
	// display text for one portfolio.
	type def struct {
		label  string
		hint   string
		cell   func(r *result) (float64, string)
		better int // +1 higher wins, -1 lower wins, 0 no highlight
	}
	pct := func(get func(metrics.Stats) float64) func(*result) (float64, string) {
		return func(r *result) (float64, string) { v := get(r.stats); return v, fmtPct(v) }
	}
	num := func(get func(metrics.Stats) float64) func(*result) (float64, string) {
		return func(r *result) (float64, string) { v := get(r.stats); return v, fmtNum(v) }
	}
	// Money rows only appear when a portfolio declares a starting capital
	// ("#meta capital:"). They describe the whole simulated span (not the
	// common window) and follow the money: contributions and withdrawals
	// included, unlike the time-weighted rows above them.
	anyCapital := false
	for _, r := range results {
		if r.p.Capital > 0 {
			anyCapital = true
			break
		}
	}
	money := func(get func(r *result) (float64, bool)) func(*result) (float64, string) {
		return func(r *result) (float64, string) {
			if r.p.Capital <= 0 {
				return math.NaN(), "-"
			}
			v, ok := get(r)
			if !ok {
				return math.NaN(), "-"
			}
			return v, fmtAmount(v)
		}
	}
	moneyDefs := []def{
		{"Starting capital", "from \"#meta capital:\"",
			money(func(r *result) (float64, bool) { return r.p.Capital, true }), 0},
		{"Total contributed", "external money added over the whole simulated span",
			money(func(r *result) (float64, bool) { return r.sim.Contributed, true }), 0},
		{"Total withdrawn", "money taken out over the whole simulated span",
			money(func(r *result) (float64, bool) { return r.sim.Withdrawn, true }), 0},
		{"Final value", "worth at the end of the simulated span, flows included",
			money(func(r *result) (float64, bool) { return r.sim.Values[len(r.sim.Values)-1], true }), 0},
		{"IRR (money-weighted)", "annual rate weighting each contribution and withdrawal by its date",
			func(r *result) (float64, string) {
				if r.p.Capital <= 0 {
					return math.NaN(), "-"
				}
				dates := append([]time.Time{r.sim.Dates[0]}, r.sim.FlowDates...)
				flows := append([]float64{-r.p.Capital}, negate(r.sim.FlowAmounts)...)
				irr, ok := metrics.IRR(dates, flows,
					r.sim.Dates[len(r.sim.Dates)-1], r.sim.Values[len(r.sim.Values)-1])
				if !ok {
					return math.NaN(), "-"
				}
				return irr, fmtPct(irr)
			}, +1},
	}
	defs := []def{
		{"CAGR (annualized return)", "compound annual growth rate",
			pct(func(s metrics.Stats) float64 { return s.CAGR }), +1},
		{"Volatility (annualized)", "standard deviation of daily returns, annualized",
			pct(func(s metrics.Stats) float64 { return s.Volatility }), -1},
		{"Volatility (monthly, annualized)", "standard deviation of monthly returns, annualized; lower than the daily figure means daily noise that mean-reverts within the month",
			func(r *result) (float64, string) {
				if !r.hasVTS {
					return math.NaN(), "-"
				}
				return r.vts.MonthlyVol, fmtPct(r.vts.MonthlyVol)
			}, -1},
		{"Variance ratio (monthly/daily)", "monthly vs daily annualized variance; ≈1 i.i.d., <1 mean-reverting (daily vol overstates risk), >1 trending (it understates)",
			func(r *result) (float64, string) {
				if !r.hasVTS {
					return math.NaN(), "-"
				}
				return r.vts.Ratio, fmtNum(r.vts.Ratio)
			}, 0},
		{"Sharpe", "annualized return / volatility (risk-free rate 0)",
			num(func(s metrics.Stats) float64 { return s.Sharpe }), +1},
		{"Sharpe (monthly)", "same ratio on monthly returns; above the daily Sharpe when daily noise mean-reverts (variance ratio <1), below it when the series trends",
			func(r *result) (float64, string) {
				if !r.hasVTS {
					return math.NaN(), "-"
				}
				return r.vts.MonthlySharpe, fmtNum(r.vts.MonthlySharpe)
			}, +1},
		{"Sortino", "annualized return / volatility of down days only",
			num(func(s metrics.Stats) float64 { return s.Sortino }), +1},
		{"Sortino (monthly)", "annualized return / downside deviation of monthly returns; the monthly twin of Sortino",
			func(r *result) (float64, string) {
				if !r.hasVTS {
					return math.NaN(), "-"
				}
				return r.vts.MonthlySortino, fmtNum(r.vts.MonthlySortino)
			}, +1},
		{"Ulcer Index", "average depth and duration of drawdowns (lower is better)",
			num(func(s metrics.Stats) float64 { return s.Ulcer }), -1},
		{"Max Drawdown", "worst decline from a peak",
			pct(func(s metrics.Stats) float64 { return s.MaxDrawdown }), +1},
		{"Max Drawdown (real)", "worst decline from a peak in real terms (deflated by the base-currency CPI): the loss of purchasing power",
			func(r *result) (float64, string) {
				if !r.hasReal {
					return math.NaN(), "-"
				}
				return r.realStats.MaxDrawdown, fmtPct(r.realStats.MaxDrawdown)
			}, +1},
		{"TTR (longest recovery)", "duration of the longest stretch below a peak",
			func(r *result) (float64, string) { return float64(r.stats.TTRDays), fmtTTR(r.stats) }, -1},
		{"TTR real (longest recovery)", "longest stretch below a peak in real terms; inflation lengthens it (e.g. S&P 500 dot-com: ~6y nominal vs ~13y real)",
			func(r *result) (float64, string) {
				if !r.hasReal {
					return math.NaN(), "-"
				}
				return float64(r.realStats.TTRDays), fmtTTR(r.realStats)
			}, -1},
		{"Weighted ongoing charges", "Σ weight × published TER, plus the extra-fees applied to the whole portfolio (only the latter are deducted from the simulation); \"(i)\" means some component TER is unknown, so the figure is incomplete",
			func(r *result) (float64, string) {
				w, incomplete := weightedFees(r.p)
				text := fmtPct(w / 100)
				if incomplete && !math.IsNaN(w) {
					text += " (i)"
				}
				return w, text
			}, -1},
		{"Worst rolling 5y CAGR", "lowest annualized return over any 5-year window of the common period",
			func(r *result) (float64, string) {
				worst, _, _, _, ok := metrics.RollingCAGR(r.winDates, r.winValues, 5)
				if !ok {
					return math.NaN(), "-"
				}
				return worst, fmtPct(worst)
			}, +1},
		{"Median rolling 5y CAGR", "median annualized return over all 5-year windows",
			func(r *result) (float64, string) {
				_, med, _, _, ok := metrics.RollingCAGR(r.winDates, r.winValues, 5)
				if !ok {
					return math.NaN(), "-"
				}
				return med, fmtPct(med)
			}, +1},
		{"Alpha (vs " + benchmark + ")", "annualized Jensen's alpha against the benchmark",
			func(r *result) (float64, string) {
				if !r.hasRel {
					return math.NaN(), "-"
				}
				return r.rel.Alpha, fmtPct(r.rel.Alpha)
			}, +1},
		{"Information ratio", "mean active return / tracking error vs the benchmark",
			func(r *result) (float64, string) {
				if !r.hasRel {
					return math.NaN(), "-"
				}
				return r.rel.InfoRatio, fmtNum(r.rel.InfoRatio)
			}, +1},
		{"Up capture", "participation in benchmark up days (>100 % = amplifies gains)",
			func(r *result) (float64, string) {
				if !r.hasRel || math.IsNaN(r.rel.UpCapture) {
					return math.NaN(), "-"
				}
				return r.rel.UpCapture, fmtPct(r.rel.UpCapture)
			}, +1},
		{"Down capture", "participation in benchmark down days (<100 % = cushions losses)",
			func(r *result) (float64, string) {
				if !r.hasRel || math.IsNaN(r.rel.DownCapture) {
					return math.NaN(), "-"
				}
				return r.rel.DownCapture, fmtPct(r.rel.DownCapture)
			}, -1},
		{"Beta (vs " + benchmark + ")", "sensitivity to benchmark moves",
			func(r *result) (float64, string) {
				if !r.stats.HasBeta {
					return math.NaN(), "-"
				}
				return r.stats.Beta, fmtNum(r.stats.Beta)
			}, 0},
		{"CWARP (vs " + benchmark + ")", "Cole Wins Above Replacement: does layering 25 % of this portfolio on top of the benchmark improve its risk-adjusted returns (Sortino and return-to-drawdown)? >0 helps, <0 hurts. Unlike Sharpe it rewards non-correlation and skew, since both are measured on the combined series.",
			func(r *result) (float64, string) {
				if !r.stats.HasCWARP {
					return math.NaN(), "-"
				}
				return r.stats.CWARP, fmt.Sprintf("%+.1f", r.stats.CWARP)
			}, +1},
	}

	if anyCapital {
		defs = append(defs, moneyDefs...)
	}
	rows := make([]report.StatRow, 0, len(defs))
	for _, d := range defs {
		row := report.StatRow{Label: d.label, Hint: d.hint}
		vals := make([]float64, len(results))
		for i, r := range results {
			v, text := d.cell(r)
			vals[i] = v
			row.Cells = append(row.Cells, report.StatCell{Text: text})
		}
		markBest(row.Cells, vals, d.better)
		rows = append(rows, row)
	}
	return rows
}

// weightedFees sums weight×TER over the holdings whose TER is known, plus
// the envelope fee; incomplete reports whether some TER was unknown.
func weightedFees(p *portfolio.Portfolio) (fees float64, incomplete bool) {
	known := false
	for _, a := range p.Assets {
		if a.Fees >= 0 {
			fees += a.Weight * a.Fees
			known = true
		} else {
			incomplete = true
		}
	}
	if p.EnvelopeFees > 0 {
		fees += p.EnvelopeFees
		known = true
	}
	if !known {
		return math.NaN(), incomplete
	}
	return fees, incomplete
}

// markBest highlights the cell(s) holding the best value of a row.
func markBest(cells []report.StatCell, vals []float64, better int) {
	if better == 0 || len(vals) < 2 {
		return
	}
	best := math.NaN()
	for _, v := range vals {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if math.IsNaN(best) || (better > 0 && v > best) || (better < 0 && v < best) {
			best = v
		}
	}
	if math.IsNaN(best) {
		return
	}
	for i, v := range vals {
		if !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v-best) <= 1e-12*math.Max(1, math.Abs(best)) {
			cells[i].Best = true
		}
	}
}

// fmtAmount renders a money amount with thin-space thousands separators.
func fmtAmount(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	out := strings.Join(parts, "\u202f")
	if neg {
		out = "-" + out
	}
	return out
}

func fmtPct(x float64) string {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return "-"
	}
	return fmt.Sprintf("%.2f %%", x*100)
}

func fmtNum(x float64) string {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return "-"
	}
	return fmt.Sprintf("%.2f", x)
}

func fmtTTR(s metrics.Stats) string {
	if s.TTRDays <= 0 {
		return "-"
	}
	out := fmt.Sprintf("%d d", s.TTRDays)
	if s.TTRDays >= 365 {
		out = fmt.Sprintf("%.1f y (%d d)", float64(s.TTRDays)/365.25, s.TTRDays)
	}
	if s.TTROngoing {
		out += " (ongoing)"
	}
	return out
}
