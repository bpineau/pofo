// Command pofo reads portfolio description files, downloads the price
// history of each asset, simulates the portfolios with periodic rebalancing
// and produces a self-contained HTML report comparing them.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	iofs "io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
)

func main() {
	log.SetFlags(0)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		log.Fatal("pofo: ", err)
	}
}

type options struct {
	out        string
	dataDir    string
	simdataDir string
	simdata    iofs.FS // source of the simulated histories (embedded or -simdata)
	rebalance  int
	start      time.Time
	end        time.Time // zero = up to today
	benchmark  string
	noOpen     bool
	noSim      bool
	noFees     bool
	currency   string
	cli        bool
	web        bool              // rendered inside the -serve web app (warm skin + site nav)
	fireHref   map[string]string // per-spec-name simulator links for the web report (opt.web only)
	width      int
	cacheAge   time.Duration
	fw         suggest.Framework // classification used by coverage and -suggest
}

// frameworkFor resolves the -framework flag to a classification.
func frameworkFor(name string) (suggest.Framework, error) {
	switch name {
	case "", "regimes":
		return suggest.RegimeFramework(), nil
	case "factors":
		return suggest.FactorFramework(), nil
	default:
		return suggest.Framework{}, fmt.Errorf("unknown -framework %q (regimes or factors)", name)
	}
}

// result holds everything computed for one portfolio.
type result struct {
	p             *portfolio.Portfolio
	sim           *portfolio.SimResult
	color         string
	rebalanceDays int
	currency      string // base currency this column was evaluated in
	specName      string // the spec this column came from (p.Name may be decorated: currency tag, "as written")
	note          string // informational line (e.g. optimizer choice)
	// Common-window view, renormalized to 100, used for stats and comparison.
	winDates  []time.Time
	winValues []float64
	stats     metrics.Stats
	realStats metrics.Stats // stats on the inflation-adjusted (deflated) window
	hasReal   bool
	rel       metrics.Relative
	hasRel    bool
	vts       metrics.VolTermStructure // daily/monthly volatility term structure
	hasVTS    bool
}

func run(ctx context.Context, argv []string) error {
	fs := flag.NewFlagSet("pofo", flag.ContinueOnError)
	var opt options
	var startStr string
	fs.StringVar(&opt.out, "out", "", "output HTML file (default: /tmp/pofo-<timestamp>.html)")
	fs.StringVar(&opt.dataDir, "data", defaultDataDir(), "quote cache directory")
	fs.StringVar(&opt.simdataDir, "simdata", "", "directory of simulated histories (default: embedded in the binary)")
	fs.IntVar(&opt.rebalance, "rebalance", 90, "rebalance every N calendar days (0 = never)")
	fs.StringVar(&startStr, "start", "", "desired start date (YYYY-MM-DD, default: earliest available)")
	fs.StringVar(&startStr, "s", "", "shorthand for -start")
	var endStr string
	fs.StringVar(&endStr, "end", "", "end date (YYYY-MM-DD, default: last available quote)")
	fs.StringVar(&endStr, "e", "", "shorthand for -end")
	fs.StringVar(&opt.benchmark, "benchmark", "^GSPC", "reference symbol for Beta (empty = no Beta)")
	fs.BoolVar(&opt.noOpen, "no-open", false, "do not open the report in the browser")
	fs.BoolVar(&opt.noSim, "no-simulate", false, "ignore SIM suffixes: real data only")
	fs.BoolVar(&opt.noFees, "no-fees", false, "do not fetch the assets' ongoing charges (TER)")
	fs.StringVar(&opt.currency, "currency", "EUR", "convert every series to this currency (empty: keep native currencies)")
	fs.BoolVar(&opt.cli, "cli", false, "render in the terminal (curves + summary table), no HTML")
	fs.IntVar(&opt.width, "width", 0, "chart width in -cli mode, in columns (default: $COLUMNS, else 100)")
	fs.DurationVar(&opt.cacheAge, "cache-age", 30*24*time.Hour, "re-download quotes older than this duration")
	warmup := fs.Bool("warmup", false, "pre-fetch the cache for the bundled asset catalog, then stop")
	verifyData := fs.Bool("verify-data", false, "data doctor: check the quotes of the referenced assets (or the whole catalog) for anomalies, then exit")
	suggestFlag := fs.Bool("suggest", false, "suggest catalog assets to add for better regime coverage/diversification, then exit")
	frameworkName := fs.String("framework", "regimes", "classification for coverage and -suggest: \"regimes\" (macro quadrants) or \"factors\" (risk factors)")
	coverageFlag := fs.Bool("coverage", false, "offline coverage advisor: show which regimes/factors a portfolio misses and the catalog assets that fill them, then exit")
	fireFlag := fs.Bool("fire", false, "open the decumulation/FIRE explorer (local web UI; optionally for a portfolio file), then serve until stopped")
	serveFlag := fs.Bool("serve", false, "serve the web app (hub, visualizer, FIRE simulator, book) until stopped; portfolio file args feed the FIRE historical models")
	listenAddr := fs.String("listen", "127.0.0.1:8787", "listen address for -serve")
	permanentFlag := fs.Bool("permanent", false, "backtest the tactical Permanent Portfolio 2.0 (Darcet) and its ruin probabilities vs the static PP, then exit")
	genSimdata := fs.Bool("gen-simdata", false, "(re)generate the simulated histories (recipes as arguments, default: all) then stop; rebuild afterwards to re-embed them")
	dry := fs.Bool("dry", false, "with -gen-simdata: validate without writing")
	refdataDir := fs.String("refdata", "", "dev override: directory of extra local reference CSVs for -gen-simdata")
	assetsList := fs.String("assets", "", "comma-separated list of tickers/ISINs, each compared as a portfolio 100 % invested in it")
	fs.StringVar(assetsList, "a", "", "shorthand for -assets")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Usage: pofo [options] portfolio.txt [portfolio2.txt …]
       pofo [options] -assets VOO,IWDA,NTSG

Without files, -assets A,B,C compares each asset as a portfolio
100 %% invested in it (can be combined with files).

File format: one line per asset:

    <weight in %%> <identifier> [fees in %%/yr]

  - Everything after a # is a comment, and nothing else may follow the
    optional fee column; blank lines are ignored.
  - Identifier: US ticker (VOO), European ticker from the bundled list
    (IWDA, CW8, CSPX…), ISIN, or catalog alias (GOLD, NTSG, BHMG…).
  - SIM suffix (VOOSIM, DBMFSIM…): extends the history before the first
    real quote via pkg/datasets/simdata/ or a proxy; bare = real data only.
  - Optional 3rd column: the asset's TER in %%/yr (overrides the automatic
    lookup); use a # comment for any other free text.
  - Per-portfolio directives:
        #meta rebalance:N    rebalance every N days (0 = never)
        #meta extra-fees:X   envelope fees in %%/yr, deducted from the
                             performance (synonym: envelope-fees)
        #meta sim:on         backcast every holding (as if each id carried
                             the SIM suffix); falls back to real quotes when
                             a holding has no simulated history
        #meta leverage:on    weights kept as written: sum > 100 %%
                             financed at the cash rate (^IRX) + spread
        #meta borrow-spread:X  borrowing spread in %%/yr (default 1.0)
        #meta capital:X      starting amount (required for flows; money
                             rows and IRR appear in the statistics)
        #meta contribute:A/P add A every period P in {week, month,
                             quarter, year}, e.g. contribute:500/month
        #meta withdraw:A/P   take out A, or A%% of the current value
                             (withdraw:4%%/year), every period P
        #meta optimize:OBJ   compute the weights: OBJ is max-sharpe,
                             min-volatility, risk-parity, max-sortino,
                             return-to-drawdown, min-ulcer, max-worst-5y or
                             cwarp (maximize CWARP vs the benchmark), with an
                             optional ",max-weight:40" cap. The report shows
                             the optimized weights next to the written ones.

Example:
    #meta rebalance:30
    #meta extra-fees:0.5
    60   VTI           US equities
    25,5 IE00B4L5Y983  # ISIN; decimal comma accepted
    14.5 GOLDSIM       gold, history extended before the first quote

Options:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(argv); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	files := fs.Args()
	if len(files) == 0 && *assetsList == "" && !*warmup && !*genSimdata && !*verifyData && !*suggestFlag && !*coverageFlag && !*fireFlag && !*serveFlag && !*permanentFlag {
		fs.Usage()
		return errors.New("no portfolio file and no -assets option")
	}
	// An empty -start means "earliest available": leave opt.start at the zero
	// time so fetches and the simdata trim keep every point, and the common
	// window then aligns on the youngest holding's inception. This surfaces the
	// full backcast by default instead of clipping it at a fixed recent date.
	if startStr != "" {
		start, err := time.ParseInLocation("2006-01-02", startStr, time.UTC)
		if err != nil {
			return fmt.Errorf("invalid -start option: %w", err)
		}
		opt.start = start
	}
	var err error
	if opt.fw, err = frameworkFor(*frameworkName); err != nil {
		return err
	}
	if endStr != "" {
		end, err := time.ParseInLocation("2006-01-02", endStr, time.UTC)
		if err != nil {
			return fmt.Errorf("invalid -end option: %w", err)
		}
		if !end.After(opt.start) {
			return errors.New("-end must be after -start")
		}
		opt.end = end
	}

	if opt.simdataDir != "" {
		opt.simdata = os.DirFS(opt.simdataDir)
	} else {
		opt.simdata = datasets.Simdata()
	}

	if *serveFlag {
		for name, on := range map[string]bool{
			"-fire": *fireFlag, "-cli": opt.cli, "-warmup": *warmup,
			"-verify-data": *verifyData, "-suggest": *suggestFlag,
			"-coverage": *coverageFlag, "-permanent": *permanentFlag,
			"-gen-simdata": *genSimdata,
		} {
			if on {
				return fmt.Errorf("-serve cannot be combined with %s", name)
			}
		}
	}

	// Generation mode consumes positional args as recipe ids, not files;
	// dispatch before any portfolio parsing.
	if *genSimdata {
		genClient := marketdata.NewClient(opt.dataDir)
		genClient.MaxAge = opt.cacheAge
		genClient.Logf = log.Printf
		return runGenSimdata(ctx, genClient, &opt, *refdataDir, fs.Args(), *dry)
	}

	// Parse every portfolio file, disambiguating duplicate names, then add
	// one synthetic 100 % portfolio per -assets entry.
	specs := make([]*portfolio.Spec, 0, len(files))
	nameCount := map[string]int{}
	addSpec := func(spec *portfolio.Spec) {
		nameCount[spec.Name]++
		if n := nameCount[spec.Name]; n > 1 {
			spec.Name = fmt.Sprintf("%s (%d)", spec.Name, n)
		}
		specs = append(specs, spec)
	}
	for _, f := range files {
		spec, err := portfolio.ParseFile(f)
		if err != nil {
			return err
		}
		addSpec(spec)
	}
	for id := range strings.SplitSeq(*assetsList, ",") {
		if id = strings.TrimSpace(id); id != "" {
			addSpec(portfolio.Single(id))
		}
	}
	if len(specs) == 0 && !*warmup && !*verifyData && !*suggestFlag && !*coverageFlag && !*fireFlag && !*serveFlag && !*permanentFlag {
		return errors.New("the -assets option contains no identifier")
	}

	client := marketdata.NewClient(opt.dataDir)
	client.MaxAge = opt.cacheAge
	client.Logf = log.Printf

	if *warmup {
		return runWarmup(ctx, client, &opt)
	}
	if *verifyData {
		return runVerifyData(ctx, client, specs, &opt)
	}
	if *suggestFlag {
		return runSuggest(ctx, client, specs, &opt)
	}
	if *coverageFlag {
		return runCoverage(specs, &opt)
	}
	if *serveFlag {
		return runServe(ctx, &opt, client, specs, *listenAddr)
	}
	if *fireFlag {
		return runFire(ctx, &opt, client, specs)
	}
	if *permanentFlag {
		return runPermanent(ctx, &opt, client)
	}

	cmp, err := computeComparison(ctx, client, &opt, specs)
	if err != nil {
		return err
	}
	if opt.cli {
		return renderCLI(cmp.results, &opt, cmp.commonStart, cmp.commonEnd, cmp.meta)
	}
	buf, err := renderPage(cmp, &opt)
	if err != nil {
		return err
	}
	outPath := opt.out
	if outPath == "" {
		outPath = fmt.Sprintf("/tmp/pofo-%s.html", time.Now().Format("20060102-150405"))
	}
	if err := os.WriteFile(outPath, buf, 0o644); err != nil {
		return err
	}
	log.Printf("report written to %s", outPath)
	if !opt.noOpen {
		openBrowser(outPath)
	}
	return nil
}

// comparison is the outcome of the fetch -> build -> simulate -> stats
// pipeline, ready for either renderer (HTML page or terminal).
type comparison struct {
	results     []*result
	bench       *marketdata.Series
	commonStart time.Time
	commonEnd   time.Time
	meta        map[string]suggest.Meta
}

// computeComparison runs the whole comparison pipeline for already-parsed
// specs: quotes and benchmark fetches, portfolio builds, simulations, the
// common window, and nominal/real statistics.
func computeComparison(ctx context.Context, client *marketdata.Client, opt *options, specs []*portfolio.Spec) (*comparison, error) {
	// Download every distinct (currency, asset) once. A "#meta currencies"
	// directive evaluates the same portfolio in several currencies.
	seriesByCur := map[string]map[string]*marketdata.Series{}
	resolved := map[string]bool{} // report each id's resolved instrument once
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.currency) {
			m := seriesByCur[cur]
			if m == nil {
				m = map[string]*marketdata.Series{}
				seriesByCur[cur] = m
			}
			for _, h := range spec.Holdings {
				// "#meta sim:on" backcasts every holding: fetch (and cache)
				// its SIM variant, keyed by the same id Build will request.
				// -no-simulate is honored downstream in FetchExtended (NoSim),
				// which fetches real quotes for a SIM id, so the flag still
				// wins over the meta with no extra handling here.
				fetchID := portfolio.SimFetchID(h.ID, spec.Sim)
				if _, ok := m[fetchID]; ok {
					continue
				}
				s, err := fetchAssetIn(ctx, client, fetchID, opt, cur)
				if err != nil {
					return nil, fmt.Errorf("portfolio %s, asset %q (%s): %w", spec.Name, h.ID, cur, err)
				}
				m[fetchID] = s
				// Surface what each identifier resolved to: a fuzzy source match
				// can return a wrong instrument (e.g. "SP500" -> an S&P sector
				// sub-index), and a silent mismatch is how delirious numbers slip
				// through. Show it once so the user can catch it.
				if !resolved[h.ID] {
					resolved[h.ID] = true
					log.Printf("resolved %s -> %q [%s, %s]", h.ID, s.Name, s.Source, s.Currency)
				}
			}
		}
	}

	// Benchmark for Beta/CWARP, best effort, memoized per currency. The chart's
	// reference curve uses the default currency (benchIn(opt.currency)).
	benchCache := map[string]*marketdata.Series{}
	benchIn := func(cur string) *marketdata.Series {
		if opt.benchmark == "" {
			return nil
		}
		if b, ok := benchCache[cur]; ok {
			return b
		}
		b, err := client.FetchExtended(ctx, opt.benchmark, marketdata.FetchOptions{
			From: opt.start, NoSim: true, Currency: cur,
		})
		if err != nil {
			log.Printf("warning: benchmark %s unavailable in %s (no Beta): %v", opt.benchmark, cur, err)
			b = nil
		}
		benchCache[cur] = b
		return b
	}
	bench := benchIn(opt.currency)

	// Simulate each portfolio; a "#meta rebalance:N" directive overrides
	// the CLI default for that portfolio only.
	var feesFor func(string) (float64, bool)
	if !opt.noFees {
		feesFor = func(id string) (float64, bool) {
			base, _ := marketdata.SplitSim(id)
			return client.Fees(ctx, base)
		}
	}
	// The financing rate (leverage) is only fetched when needed.
	var cashRate *marketdata.Series
	for _, spec := range specs {
		if spec.Leverage {
			cr, err := client.Fetch(ctx, "^IRX", opt.start)
			if err != nil {
				log.Printf("warning: financing rate ^IRX unavailable (%v), leverage financed at 0 %%", err)
			} else {
				cashRate = cr
			}
			break
		}
	}

	results := make([]*result, 0, len(specs))
	simulateInto := func(p *portfolio.Portfolio, spec *portfolio.Spec, currency string) error {
		days := opt.rebalance
		if spec.RebalanceDays >= 0 {
			days = spec.RebalanceDays
		}
		sim, err := portfolio.Simulate(p, days)
		if err != nil {
			return fmt.Errorf("portfolio %s: %w", p.Name, err)
		}
		if sim.Ruined {
			cause := "the leveraged exposure exhausted the net value"
			if p.Withdraw.Active() && !p.Leverage {
				cause = "withdrawals exhausted the capital"
			}
			when := sim.Dates[len(sim.Dates)-1].Format("2006-01-02")
			log.Printf("warning: portfolio %s wiped out on %s, series truncated", p.Name, when)
			p.Warnings = append(p.Warnings, fmt.Sprintf(
				"capital wiped out on %s: %s; the series stops there", when, cause))
		}
		results = append(results, &result{p: p, sim: sim, color: chart.PaletteColor(len(results)), rebalanceDays: days, currency: currency, specName: spec.Name})
		return nil
	}
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.currency) {
			p, err := portfolio.Build(spec, portfolio.BuildOptions{
				Fetch:        func(id string) (*marketdata.Series, error) { return seriesByCur[cur][id], nil },
				Fees:         feesFor,
				Cash:         cashRate,
				BorrowSpread: 1.0, // default: cash + 1 %/yr
				BaseCurrency: cur,
			})
			if err != nil {
				return nil, err
			}
			// Multi-currency: tag each column with its currency.
			if len(spec.Currencies) > 0 {
				p.Name = fmt.Sprintf("%s (%s)", p.Name, cur)
			}
			// An optimized portfolio is shown next to its written weights, so
			// the optimizer's choice can be compared with the baseline.
			// (Optimize and currencies cannot be combined, so cur is unique here.)
			if spec.Optimize != nil {
				pOpt, note, err := optimizedPortfolio(p, spec, benchIn(cur))
				if err != nil {
					return nil, fmt.Errorf("portfolio %s: %w", spec.Name, err)
				}
				p.Name = spec.Name + " (as written)"
				if err := simulateInto(p, spec, cur); err != nil {
					return nil, err
				}
				if err := simulateInto(pOpt, spec, cur); err != nil {
					return nil, err
				}
				results[len(results)-1].note = note
				continue
			}
			if err := simulateInto(p, spec, cur); err != nil {
				return nil, err
			}
		}
	}

	// Common window across portfolios: statistics and the comparison chart
	// must cover the same period to be meaningful.
	commonStart := results[0].sim.Dates[0]
	commonEnd := results[0].sim.Dates[len(results[0].sim.Dates)-1]
	for _, r := range results[1:] {
		if f := r.sim.Dates[0]; f.After(commonStart) {
			commonStart = f
		}
		if l := r.sim.Dates[len(r.sim.Dates)-1]; l.Before(commonEnd) {
			commonEnd = l
		}
	}
	if !commonStart.Before(commonEnd) {
		return nil, errors.New("no common period across the portfolios")
	}
	// Consumer-price index per currency, memoized, to report drawdowns/TTR and
	// real stats in purchasing-power terms alongside the nominal ones.
	// Best-effort: a currency without a wired CPI simply has no real columns.
	deflatorCache := map[string]*marketdata.Series{}
	deflatorIn := func(cur string) (*marketdata.Series, bool) {
		if s, ok := deflatorCache[cur]; ok {
			return s, s != nil
		}
		s, ok := inflationSeries(ctx, client, cur, commonStart)
		if !ok {
			s = nil
		}
		deflatorCache[cur] = s
		return s, s != nil
	}
	for _, r := range results {
		i, j := window(r.sim.Dates, commonStart, commonEnd)
		if j-i < 2 {
			return nil, fmt.Errorf("portfolio %s: too few points in the common window", r.p.Name)
		}
		r.winDates = r.sim.Dates[i:j]
		r.winValues = rebase(r.sim.Index[i:j])
		st, err := metrics.Compute(r.winDates, r.winValues)
		if err != nil {
			return nil, fmt.Errorf("portfolio %s: %w", r.p.Name, err)
		}
		if d, ok := deflatorIn(r.currency); ok {
			if rs, err := metrics.Compute(r.winDates, deflate(r.winDates, r.winValues, d)); err == nil {
				r.realStats, r.hasReal = rs, true
			}
		}
		if b := benchIn(r.currency); b != nil {
			bd, bv := seriesSlices(b)
			if rel, ok := metrics.VsBenchmark(r.winDates, r.winValues, bd, bv); ok {
				st.Beta, st.HasBeta = rel.Beta, true
				r.rel, r.hasRel = rel, true
			}
			if c, ok := metrics.CWARPvs(r.winDates, r.winValues, bd, bv, metrics.CWARPParams{}); ok {
				st.CWARP, st.HasCWARP = c, true
			}
		}
		r.vts, r.hasVTS = metrics.VarianceRatio(r.winDates, r.winValues)
		r.stats = st
	}

	assetMeta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		log.Printf("warning: asset metadata unavailable (%v), regime coverage omitted", err)
	}

	return &comparison{results: results, bench: bench, commonStart: commonStart, commonEnd: commonEnd, meta: assetMeta}, nil
}

// renderPage renders a computed comparison to the HTML report bytes.
func renderPage(cmp *comparison, opt *options) ([]byte, error) {
	page := buildPage(cmp.results, opt, cmp.bench, cmp.commonStart, cmp.commonEnd, cmp.meta)
	var buf bytes.Buffer
	if err := report.Render(&buf, page); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderComparison runs the whole pipeline and renders the HTML report:
// the single entry point the web server needs.
func renderComparison(ctx context.Context, client *marketdata.Client, opt *options, specs []*portfolio.Spec) ([]byte, error) {
	cmp, err := computeComparison(ctx, client, opt, specs)
	if err != nil {
		return nil, err
	}
	return renderPage(cmp, opt)
}

// renderCLI prints the comparison curves and the summary table straight to
// the terminal, for quick checks without opening a browser. Per-portfolio
// details are intentionally omitted.
func renderCLI(results []*result, opt *options, commonStart, commonEnd time.Time, meta map[string]suggest.Meta) error {
	color := os.Getenv("NO_COLOR") == "" && isTerminal(os.Stdout)
	names := make([]string, len(results))
	cmp := make([]chart.Series, len(results))
	for i, r := range results {
		names[i] = r.p.Name
		if len(results) == 1 {
			cmp[i] = chart.Series{Name: r.p.Name, Dates: r.sim.Dates, Values: r.sim.Values}
		} else {
			cmp[i] = chart.Series{Name: r.p.Name, Dates: r.winDates, Values: r.winValues}
		}
	}
	title := "Comparison (base 100"
	if len(results) == 1 {
		title = results[0].p.Name + " (base 100"
	}
	title += " at " + cmp[0].Dates[0].Format("2006-01-02") + ")"
	fmt.Print(chart.Term(chart.TermOptions{Title: title, Width: termWidth(opt.width), Color: color}, cmp))
	fmt.Println()

	page := &report.Page{
		Title:          "Portfolios: " + strings.Join(names, ", "),
		CommonStart:    commonStart.Format("2006-01-02"),
		CommonEnd:      commonEnd.Format("2006-01-02"),
		PortfolioNames: names,
		StatRows:       buildStatRows(results, opt.benchmark),
	}
	if err := report.RenderText(os.Stdout, page, color); err != nil {
		return err
	}
	printCoverageCLI(results, meta, opt.fw)
	return nil
}

// printCoverageCLI prints each portfolio's macro-regime coverage under the
// CLI summary table (same data as the HTML report and -suggest).
func printCoverageCLI(results []*result, meta map[string]suggest.Meta, fw suggest.Framework) {
	var lines []string
	for _, r := range results {
		bars := coverageBars(r.p.Assets, meta, fw)
		if bars == nil {
			continue
		}
		parts := make([]string, len(bars))
		for i, b := range bars {
			parts[i] = fmt.Sprintf("%s %d %%", b.Regime, b.Pct)
			if b.Gap {
				parts[i] += " (gap)"
			}
		}
		lines = append(lines, "  "+r.p.Name+": "+strings.Join(parts, "   "))
	}
	if len(lines) == 0 {
		return
	}
	fmt.Println("\nRegime coverage (share of weight; gap = under-covered; run -suggest):")
	for _, l := range lines {
		fmt.Println(l)
	}
}

// isTerminal reports whether f is attached to a character device.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// termWidth picks the chart width: the explicit -width flag wins, then
// $COLUMNS (capped), then 100.
func termWidth(flag int) int {
	if flag >= 40 {
		return min(flag, 500)
	}
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n >= 40 {
			return min(n, 160)
		}
	}
	return 100
}

// defaultDataDir picks the standard per-user cache location
// (~/Library/Caches/pofo on macOS, ~/.cache/pofo on Linux),
// falling back to a local directory when the home is unknown.
func defaultDataDir() string {
	if c, err := os.UserCacheDir(); err == nil {
		return filepath.Join(c, "pofo")
	}
	return "data"
}

func openBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		log.Printf("open %s manually", path)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("could not open the browser: %v", err)
	}
}
