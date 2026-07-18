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
	"html/template"
	iofs "io/fs"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/decumul/web"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/optimize"
	"github.com/bpineau/pofo/pkg/permanent"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/scenario"
	"github.com/bpineau/pofo/pkg/simgen"
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
	if len(files) == 0 && *assetsList == "" && !*warmup && !*genSimdata && !*verifyData && !*suggestFlag && !*coverageFlag && !*fireFlag && !*permanentFlag {
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
	if len(specs) == 0 && !*warmup && !*verifyData && !*suggestFlag && !*coverageFlag && !*fireFlag && !*permanentFlag {
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
	if *fireFlag {
		return runFire(ctx, &opt, client, specs)
	}
	if *permanentFlag {
		return runPermanent(ctx, &opt, client)
	}

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
				s, err := fetchAssetIn(ctx, client, fetchID, &opt, cur)
				if err != nil {
					return fmt.Errorf("portfolio %s, asset %q (%s): %w", spec.Name, h.ID, cur, err)
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
		results = append(results, &result{p: p, sim: sim, color: chart.PaletteColor(len(results)), rebalanceDays: days, currency: currency})
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
				return err
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
					return fmt.Errorf("portfolio %s: %w", spec.Name, err)
				}
				p.Name = spec.Name + " (as written)"
				if err := simulateInto(p, spec, cur); err != nil {
					return err
				}
				if err := simulateInto(pOpt, spec, cur); err != nil {
					return err
				}
				results[len(results)-1].note = note
				continue
			}
			if err := simulateInto(p, spec, cur); err != nil {
				return err
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
		return errors.New("no common period across the portfolios")
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
			return fmt.Errorf("portfolio %s: too few points in the common window", r.p.Name)
		}
		r.winDates = r.sim.Dates[i:j]
		r.winValues = rebase(r.sim.Index[i:j])
		st, err := metrics.Compute(r.winDates, r.winValues)
		if err != nil {
			return fmt.Errorf("portfolio %s: %w", r.p.Name, err)
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

	if opt.cli {
		return renderCLI(results, &opt, commonStart, commonEnd, assetMeta)
	}

	page := buildPage(results, &opt, bench, commonStart, commonEnd, assetMeta)
	var buf bytes.Buffer
	if err := report.Render(&buf, page); err != nil {
		return err
	}
	outPath := opt.out
	if outPath == "" {
		outPath = fmt.Sprintf("/tmp/pofo-%s.html", time.Now().Format("20060102-150405"))
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return err
	}
	log.Printf("report written to %s", outPath)
	if !opt.noOpen {
		openBrowser(outPath)
	}
	return nil
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

// runGenSimdata (re)builds the simulated histories, the former standalone
// simgen command, kept as a sub-mode. Files are written to pkg/datasets/simdata
// (or -simdata when set): regeneration is a repository activity, and a
// rebuild re-embeds the result into the binary.
func runGenSimdata(ctx context.Context, client *marketdata.Client, opt *options, refdataDir string, ids []string, dry bool) error {
	recipes := simgen.All()
	if len(ids) > 0 {
		recipes = recipes[:0]
		for _, id := range ids {
			r, ok := simgen.Find(id)
			if !ok {
				return fmt.Errorf("no recipe for %q", id)
			}
			recipes = append(recipes, r)
		}
	}
	outDir := opt.simdataDir
	if outDir == "" {
		outDir = "pkg/datasets/simdata"
	}
	// Recipes read long reference series (e.g. MSCI World) from the embedded
	// refdata first, so regeneration is self-contained; -refdata adds or
	// overrides with extra local CSVs on top.
	var fetcher simgen.Fetcher = simgen.WithRefData(datasets.Refdata(), simgen.WithContext(ctx, client))
	if refdataDir != "" {
		fetcher = simgen.WithRefData(os.DirFS(refdataDir), fetcher)
	}
	failures := 0
	for _, r := range recipes {
		err := genOne(ctx, client, fetcher, outDir, r, dry)
		switch {
		case errors.Is(err, simgen.ErrUnfaithful):
			log.Printf("⚠ %-14s skipped: %v", r.ID, err)
		case err != nil:
			log.Printf("✗ %-14s %v", r.ID, err)
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d recipe(s) failed", failures)
	}
	if !dry {
		log.Printf("rebuild (make build) to re-embed pkg/datasets/simdata into the binary")
	}
	return nil
}

func genOne(ctx context.Context, client *marketdata.Client, fetcher simgen.Fetcher, dir string, r simgen.Recipe, dry bool) error {
	sim, err := r.Build(fetcher, simgen.ComponentsFrom)
	if err != nil {
		return err
	}
	validation := "not validated (no real series)"
	if r.ValidateAgainst != "" {
		real, err := client.Fetch(ctx, r.ValidateAgainst, simgen.ComponentsFrom)
		if err != nil {
			return fmt.Errorf("real series %s: %w", r.ValidateAgainst, err)
		}
		v, err := simgen.Validate(sim, real)
		if err != nil {
			return fmt.Errorf("validation vs %s: %w", r.ValidateAgainst, err)
		}
		validation = fmt.Sprintf("%s vs %s", v, r.ValidateAgainst)
	}
	if r.SpliceReal != "" {
		real, err := client.Fetch(ctx, r.SpliceReal, simgen.ComponentsFrom)
		if err != nil {
			return fmt.Errorf("series to splice %s: %w", r.SpliceReal, err)
		}
		sim = simgen.Splice(real, sim)
	}
	log.Printf("✓ %-14s %s → %s (%d points)", r.ID,
		sim.First().Date.Format("2006-01-02"), sim.Last().Date.Format("2006-01-02"), len(sim.Points))
	log.Printf("  %s", validation)
	if dry {
		return nil
	}
	return marketdata.WriteSimdata(dir, &marketdata.SimdataFile{
		ID:         r.ID,
		Name:       r.Name,
		Method:     r.Method,
		Validation: validation,
		Generated:  time.Now().Format("2006-01-02"),
		Points:     sim.Points,
	})
}

// runVerifyData is the data doctor: it fetches every asset referenced by
// the given portfolios (or the whole bundled catalog when none is given)
// and reports data-quality findings from marketdata.Verify. It returns an
// error when any series has error-grade problems.
func runVerifyData(ctx context.Context, c *marketdata.Client, specs []*portfolio.Spec, opt *options) error {
	var ids []string
	seen := map[string]bool{}
	for _, spec := range specs {
		for _, h := range spec.Holdings {
			if !seen[h.ID] {
				seen[h.ID] = true
				ids = append(ids, h.ID)
			}
		}
	}
	if len(ids) == 0 {
		ids = marketdata.WarmupIDs()
		fmt.Printf("no portfolio given: checking the %d bundled catalog assets\n", len(ids))
	}
	now := time.Now()
	broken, suspicious := 0, 0
	for _, id := range ids {
		s, err := fetchAsset(ctx, c, id, opt)
		if err != nil {
			fmt.Printf("%-22s FETCH FAILED: %v\n", id, err)
			broken++
			continue
		}
		issues := marketdata.Verify(s, now)
		span := fmt.Sprintf("%s → %s, %d points, %s",
			s.First().Date.Format("2006-01-02"), s.Last().Date.Format("2006-01-02"),
			len(s.Points), s.Source)
		if len(issues) == 0 {
			fmt.Printf("%-22s ok (%s)\n", id, span)
			continue
		}
		hasError := false
		for _, is := range issues {
			if is.Severity == "error" {
				hasError = true
			}
		}
		if hasError {
			broken++
		} else {
			suspicious++
		}
		fmt.Printf("%-22s %d finding(s) (%s)\n", id, len(issues), span)
		for _, is := range issues {
			fmt.Printf("    %s\n", is)
		}
	}
	fmt.Printf("\n%d asset(s) checked: %d clean, %d with warnings, %d broken\n",
		len(ids), len(ids)-suspicious-broken, suspicious, broken)
	if broken > 0 {
		return fmt.Errorf("%d asset(s) with error-grade data problems", broken)
	}
	return nil
}

// runSuggest analyses each portfolio's macro-regime coverage, flags
// redundant holdings, and recommends catalog assets to add that fill the
// gaps, validated out-of-sample. See pkg/suggest and docs/suggest-design.md.
func runSuggest(ctx context.Context, c *marketdata.Client, specs []*portfolio.Spec, opt *options) error {
	meta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		return fmt.Errorf("asset metadata: %w", err)
	}
	if len(specs) == 0 {
		return errors.New("-suggest needs a portfolio (a file or -assets)")
	}
	for _, spec := range specs {
		if err := suggestForPortfolio(ctx, c, spec, opt, meta); err != nil {
			return fmt.Errorf("portfolio %s: %w", spec.Name, err)
		}
	}
	return nil
}

// metaFor resolves a portfolio identifier (alias or SIM suffix tolerated) to
// its catalog metadata.
func metaFor(meta map[string]suggest.Meta, id string) (suggest.Meta, string, bool) {
	base, _ := marketdata.SplitSim(id)
	canon := marketdata.CanonicalID(base)
	if m, ok := meta[canon]; ok {
		return m, canon, true
	}
	if m, ok := meta[base]; ok {
		return m, canon, true
	}
	return suggest.Meta{}, canon, false
}

func suggestForPortfolio(ctx context.Context, c *marketdata.Client, spec *portfolio.Spec, opt *options, meta map[string]suggest.Meta) error {
	if len(spec.Holdings) == 0 {
		return errors.New("empty portfolio")
	}
	// Fetch the held assets (series come back already in the base currency).
	held := make([]*marketdata.Series, len(spec.Holdings))
	holdings := make([]suggest.Holding, len(spec.Holdings))
	weights := make([]float64, len(spec.Holdings))
	heldCanon := map[string]bool{}
	heldEquiv := map[string]bool{}
	for i, h := range spec.Holdings {
		s, err := fetchAsset(ctx, c, h.ID, opt)
		if err != nil {
			return fmt.Errorf("asset %q: %w", h.ID, err)
		}
		m, canon, ok := metaFor(meta, h.ID)
		held[i] = s
		weights[i] = h.Weight
		holdings[i] = suggest.Holding{ID: h.ID, Weight: h.Weight, Meta: m, HasMeta: ok}
		heldCanon[canon] = true
		if ok {
			heldEquiv[m.AssetClass+"|"+m.Benchmark] = true
		}
	}

	start, end := commonWindow(held)
	if !start.Before(end) {
		return errors.New("the held assets have no period in common")
	}
	_, prices := marketdata.Align(held, start, end)
	heldRet := make([][]float64, len(prices))
	for i, px := range prices {
		heldRet[i] = metrics.Returns(px)
	}

	opts := suggest.DefaultOptions()
	cov, _ := suggest.Coverage(holdings, opt.fw)
	gaps := suggest.Gaps(cov, opt.fw, opts.GapThreshold)

	candidates := buildCandidates(ctx, c, opt, meta, gaps, heldCanon, heldEquiv, held, weights)
	res := suggest.Analyze(holdings, heldRet, candidates, opts, opt.fw)
	renderSuggest(spec.Name, start, end, res, opt.fw)
	return nil
}

// buildCandidates picks a small, diverse set of catalog assets that fill a
// gap regime (deduped by class+strategy, capped per regime, never a
// near-duplicate of a holding), fetches their histories (simulated extension
// included for the longest fair comparison) and aligns each with the held
// portfolio over their overlap.
func buildCandidates(ctx context.Context, c *marketdata.Client, opt *options, meta map[string]suggest.Meta, gaps []suggest.Category,
	heldCanon, heldEquiv map[string]bool, held []*marketdata.Series, weights []float64) []suggest.Candidate {
	if len(gaps) == 0 {
		return nil
	}
	gapSet := map[suggest.Category]bool{}
	for _, g := range gaps {
		gapSet[g] = true
	}
	const (
		maxPerGap = 4
		minYears  = 3.0
	)
	// One representative per (class, strategy), highest confidence wins.
	type rep struct {
		id string
		m  suggest.Meta
	}
	repByKey := map[string]rep{}
	for _, id := range marketdata.WarmupIDs() {
		if heldCanon[id] {
			continue
		}
		m, ok := meta[id]
		if !ok || !intersectsGap(opt.fw.Classify(m), gapSet) {
			continue
		}
		if heldEquiv[m.AssetClass+"|"+m.Benchmark] {
			continue // you already hold an equivalent fund
		}
		key := m.AssetClass + "|" + m.Strategy
		if ex, seen := repByKey[key]; !seen || confRank(m.Confidence) > confRank(ex.m.Confidence) {
			repByKey[key] = rep{id, m}
		}
	}
	// Deterministic order (map iteration is randomized).
	reps := make([]rep, 0, len(repByKey))
	for _, r := range repByKey {
		reps = append(reps, r)
	}
	sort.Slice(reps, func(i, j int) bool { return reps[i].id < reps[j].id })

	// Cap per gap category, most-under-covered first.
	perGap := map[suggest.Category]int{}
	picked := map[string]bool{}
	var order []rep
	for _, g := range gaps {
		for _, r := range reps {
			if picked[r.id] || perGap[g] >= maxPerGap {
				continue
			}
			if !intersectsGap(opt.fw.Classify(r.m), map[suggest.Category]bool{g: true}) {
				continue
			}
			perGap[g]++
			picked[r.id] = true
			order = append(order, r)
		}
	}
	if dropped := len(reps) - len(order); dropped > 0 {
		log.Printf("suggest: %d gap-filling candidate(s) beyond %d per category were not evaluated", dropped, maxPerGap)
	}

	var out []suggest.Candidate
	for _, r := range order {
		cs, err := fetchAsset(ctx, c, r.id+"SIM", opt)
		if err != nil {
			log.Printf("suggest: candidate %s unavailable: %v", r.id, err)
			continue
		}
		list := append(append([]*marketdata.Series{}, held...), cs)
		cstart, cend := commonWindow(list)
		if !cstart.Before(cend) {
			continue
		}
		years := cend.Sub(cstart).Hours() / 24 / 365.25
		if years < minYears {
			log.Printf("suggest: candidate %s skipped (only %.1f years overlap)", r.id, years)
			continue
		}
		_, p := marketdata.Align(list, cstart, cend)
		heldRet := make([][]float64, len(held))
		for i := range held {
			heldRet[i] = metrics.Returns(p[i])
		}
		out = append(out, suggest.Candidate{
			Meta:        r.m,
			PortReturns: suggest.PortfolioReturns(weights, heldRet),
			Returns:     metrics.Returns(p[len(held)]),
			Years:       years,
			Simulated:   !cs.SimulatedBefore.IsZero(),
		})
	}
	return out
}

func intersectsGap(rs []suggest.Category, gapSet map[suggest.Category]bool) bool {
	for _, r := range rs {
		if gapSet[r] {
			return true
		}
	}
	return false
}

func confRank(c string) int {
	switch c {
	case "high":
		return 2
	case "medium":
		return 1
	default:
		return 0
	}
}

// commonWindow returns the latest first date and earliest last date across
// the series (their overlapping period).
func commonWindow(list []*marketdata.Series) (start, end time.Time) {
	start, end = list[0].First().Date, list[0].Last().Date
	for _, s := range list[1:] {
		if f := s.First().Date; f.After(start) {
			start = f
		}
		if l := s.Last().Date; l.Before(end) {
			end = l
		}
	}
	return start, end
}

// printCoverageBars prints one bar per framework category, marking gaps.
func printCoverageBars(cov map[suggest.Category]float64, gaps []suggest.Category, unclassified float64, fw suggest.Framework) {
	gapSet := map[suggest.Category]bool{}
	for _, g := range gaps {
		gapSet[g] = true
	}
	for _, c := range fw.Categories {
		pct := cov[c] * 100
		bars := min(int(pct/5+0.5), 20)
		mark := ""
		if gapSet[c] {
			mark = "   ← gap"
		}
		fmt.Printf("  %-11s %-20s %3.0f %%%s\n", c, strings.Repeat("█", bars), pct, mark)
	}
	if unclassified > 0 {
		fmt.Printf("  (%.0f %% of the portfolio is unclassified, no catalog metadata)\n", unclassified*100)
	}
}

// renderSuggest prints the analysis for one portfolio.
func renderSuggest(name string, start, end time.Time, res suggest.Result, fw suggest.Framework) {
	fmt.Printf("\n=== Suggestions for %s (%s) ===\n", name, fw.Name)
	fmt.Printf("Coverage over %s → %s (by weight):\n",
		start.Format("2006-01-02"), end.Format("2006-01-02"))
	printCoverageBars(res.Coverage, res.Gaps, res.Unclassified, fw)

	if len(res.Redundancies) > 0 {
		fmt.Println("\nRedundancies (effectively one bet held several times):")
		for _, g := range res.Redundancies {
			fmt.Printf("  • %s: %.0f %% of the portfolio, correlation ≥ %.2f\n",
				strings.Join(g.IDs, " + "), g.Weight*100, g.MinCorr)
		}
	}

	if len(res.Gaps) == 0 {
		fmt.Printf("\nEvery %s category is covered, no gap to fill.\n", fw.Name)
		return
	}
	if len(res.Suggestions) == 0 {
		fmt.Println("\nNo gap-filling asset showed a robust out-of-sample benefit over the available history.")
		return
	}
	fmt.Println("\nSuggestions to add (fill the gaps, validated out-of-sample):")
	for i, s := range res.Suggestions {
		fmt.Printf("  %d. %s (%s), fills the %s gap\n", i+1, s.Meta.ID, s.Meta.AssetClass, s.Fills)
		fmt.Printf("     suggested weight %.0f %%  ·  correlation to portfolio %.2f  ·  daily vol %.2f %% → %.2f %%\n",
			s.Weight*100, s.Corr, s.VolBefore*100, s.VolAfter*100)
		fmt.Printf("     out-of-sample: Sharpe improved in %d/%d windows, max-drawdown in %d/%d (median Sharpe gain %+.2f)\n",
			s.SharpeWins, s.Windows, s.DDWins, s.Windows, s.MedSharpeGain)
		line := "     "
		if s.Meta.Notes != "" {
			line += s.Meta.Notes
		}
		if s.Simulated {
			line += "  [history partly simulated]"
		}
		if strings.TrimSpace(line) != "" {
			fmt.Println(line)
		}
	}
}

// runCoverage is the offline coverage advisor: it shows which framework
// categories a portfolio under-covers and lists the catalog assets that
// would fill each gap. It needs no price data, only the embedded metadata.
func runCoverage(specs []*portfolio.Spec, opt *options) error {
	meta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		return fmt.Errorf("asset metadata: %w", err)
	}
	if len(specs) == 0 {
		return errors.New("-coverage needs a portfolio (a file or -assets)")
	}
	for _, spec := range specs {
		coverageAdvice(spec, opt, meta)
	}
	return nil
}

func coverageAdvice(spec *portfolio.Spec, opt *options, meta map[string]suggest.Meta) {
	holdings := make([]suggest.Holding, len(spec.Holdings))
	held := map[string]bool{}
	for i, h := range spec.Holdings {
		m, canon, ok := metaFor(meta, h.ID)
		holdings[i] = suggest.Holding{ID: h.ID, Weight: h.Weight, Meta: m, HasMeta: ok}
		held[canon] = true
	}
	cov, uncl := suggest.Coverage(holdings, opt.fw)
	gaps := suggest.Gaps(cov, opt.fw, suggest.DefaultOptions().GapThreshold)

	fmt.Printf("\n=== Coverage advisor for %s (%s) ===\n", spec.Name, opt.fw.Name)
	fmt.Println("Coverage (by weight):")
	printCoverageBars(cov, gaps, uncl, opt.fw)
	if len(gaps) == 0 {
		fmt.Printf("\nEvery %s category is covered.\n", opt.fw.Name)
		return
	}

	fmt.Println("\nTo fill the gaps, the catalog offers (run -suggest to rank them):")
	for _, g := range gaps {
		byClass := map[string][]string{}
		var classes []string
		for _, id := range marketdata.WarmupIDs() {
			if held[id] {
				continue
			}
			m, ok := meta[id]
			if !ok || !intersectsGap(opt.fw.Classify(m), map[suggest.Category]bool{g: true}) {
				continue
			}
			if _, seen := byClass[m.AssetClass]; !seen {
				classes = append(classes, m.AssetClass)
			}
			byClass[m.AssetClass] = append(byClass[m.AssetClass], id)
		}
		fmt.Printf("  %s:\n", g)
		if len(classes) == 0 {
			fmt.Println("    (no catalog asset available)")
			continue
		}
		sort.Strings(classes)
		for _, cl := range classes {
			ids := byClass[cl]
			extra := ""
			if len(ids) > 3 {
				extra = fmt.Sprintf(" … (+%d)", len(ids)-3)
				ids = ids[:3]
			}
			fmt.Printf("    %-16s %s%s\n", cl, strings.Join(ids, ", "), extra)
		}
	}
}

// runWarmup pre-fetches the whole bundled asset catalog into the cache so
// that later runs work fast and (mostly) offline.
func runWarmup(ctx context.Context, c *marketdata.Client, opt *options) error {
	// Refresh the bundled CPI/HICP deflators from their live source (a normal
	// run serves them offline-first from the embedded snapshot).
	c.RefreshInflation = true
	for _, sym := range []string{"^CPI-US", "^HICP-FR"} {
		if s, err := c.Fetch(ctx, sym, opt.start); err != nil {
			log.Printf("FAIL  %s: %v", sym, err)
		} else {
			log.Printf("OK    %-24s %s → %s  (%d points)", sym,
				s.First().Date.Format("2006-01-02"), s.Last().Date.Format("2006-01-02"), len(s.Points))
		}
	}

	ids := marketdata.WarmupIDs()
	var failed []string
	for i, id := range ids {
		if i > 0 {
			time.Sleep(300 * time.Millisecond) // go easy on the sources
		}
		s, err := fetchAsset(ctx, c, id, opt)
		if err != nil {
			log.Printf("FAIL  %s: %v", id, err)
			failed = append(failed, id)
			continue
		}
		feesText := ""
		if !opt.noFees {
			if ter, ok := c.Fees(ctx, id); ok {
				feesText = fmt.Sprintf("  TER %.2f %%", ter)
			}
		}
		log.Printf("OK    %-24s %s → %s  (%d quotes)%s %s", id,
			s.First().Date.Format("2006-01-02"), s.Last().Date.Format("2006-01-02"),
			len(s.Points), feesText, s.Name)
	}
	log.Printf("warmup done: %d/%d assets cached", len(ids)-len(failed), len(ids))
	if len(failed) > 0 {
		log.Printf("failed: %s", strings.Join(failed, ", "))
	}
	return nil
}

// runFire starts the embedded decumulation explorer on a local port and
// opens it in the browser. With a portfolio file it builds a historical
// real-return panel from the holdings (deflated by ^HICP-FR) so the UI can
// switch to the bootstrap/cohort models and re-weight allocations live. It
// blocks, serving until interrupted.
func runFire(ctx context.Context, opt *options, c *marketdata.Client, specs []*portfolio.Spec) error {
	chart.SetDark(true) // the FIRE explorer renders in the terminal dark theme
	var panel *scenario.Panel
	var labels []string
	if len(specs) > 0 {
		var assets []web.AssetSeries
		for _, h := range specs[0].Holdings {
			// Honour "#meta sim:on" exactly like portfolio.Build: fetch the
			// SIM (backcast-extended) variant, falling back to the real
			// quotes when no backcast exists. The FIRE panel needs the deep
			// history; real-only overlaps of recent funds are often too
			// short to fit or resample a retirement-length horizon.
			fetchID := portfolio.SimFetchID(h.ID, specs[0].Sim)
			s, err := fetchAsset(ctx, c, fetchID, opt)
			if err != nil && fetchID != h.ID {
				log.Printf("fire: %s: no simulated history, using real quotes", h.ID)
				s, err = fetchAsset(ctx, c, h.ID, opt)
			}
			if err != nil {
				log.Printf("fire: skipping %s: %v", h.ID, err)
				continue
			}
			labels = append(labels, h.ID)
			assets = append(assets, web.AssetSeries{Weight: h.Weight, Points: s.Points})
		}
		if hicp, err := fetchAsset(ctx, c, "^HICP-FR", opt); err == nil {
			if pnl, err := web.BuildMonthlyPanel(assets, hicp.Points); err == nil {
				panel = &pnl
			} else {
				log.Printf("fire: no historical panel: %v", err)
			}
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Fprintf(os.Stderr, "FIRE explorer on %s (Ctrl-C to stop)\n", url)
	if !opt.noOpen {
		openBrowser(url)
	}
	// main() routes SIGINT/SIGTERM into ctx (signal.NotifyContext), which
	// replaces the default die-on-Ctrl-C behavior, so the server must watch
	// the context and shut down when it fires.
	srv := &http.Server{Handler: web.Handler(panel, labels)}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// fetchAsset downloads the history of an identifier (ticker or ISIN). A
// bare identifier sticks to the asset's real quotes, from its actual
// inception. A "SIM"-suffixed identifier (DBMFSIM, VOOSIM…) additionally
// extends the uncovered period backwards: first with the permanent simulated
// series (embedded datasets, or -simdata), then a known proxy; real
// quotes always win wherever they exist.
// fetchAsset runs the full library pipeline (SIM extension, currency
// conversion, window) for one asset, in the CLI's base currency.
func fetchAsset(ctx context.Context, c *marketdata.Client, id string, opt *options) (*marketdata.Series, error) {
	return fetchAssetIn(ctx, c, id, opt, opt.currency)
}

// fetchAssetIn is fetchAsset with an explicit target currency, used when a
// portfolio is evaluated in several currencies ("#meta currencies").
func fetchAssetIn(ctx context.Context, c *marketdata.Client, id string, opt *options, currency string) (*marketdata.Series, error) {
	return c.FetchExtended(ctx, id, marketdata.FetchOptions{
		From:     opt.start,
		To:       opt.end,
		NoSim:    opt.noSim,
		Simdata:  opt.simdata,
		Currency: currency,
	})
}

// effectiveCurrencies is the list of base currencies a spec expands into:
// its "#meta currencies" list when set, otherwise the single CLI default.
func effectiveCurrencies(spec *portfolio.Spec, def string) []string {
	if len(spec.Currencies) > 0 {
		return spec.Currencies
	}
	return []string{def}
}

// inflationSeries returns the consumer-price index used to deflate nominal
// returns into real (purchasing-power) ones for the base currency, and whether
// one is available. The euro is deflated by French HICP (^HICP-FR, the long
// bundled series, ~1955→), the dollar by the US CPI (^CPI-US, bundled from
// 1913); other currencies have no wired CPI yet, so their real drawdown/TTR
// columns are simply omitted.
func inflationSeries(ctx context.Context, c *marketdata.Client, currency string, from time.Time) (*marketdata.Series, bool) {
	sym := ""
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "EUR":
		sym = "^HICP-FR"
	case "USD":
		sym = "^CPI-US"
	}
	if sym == "" {
		return nil, false
	}
	s, err := c.Fetch(ctx, sym, from)
	if err != nil || s == nil || len(s.Points) < 2 {
		if err != nil {
			log.Printf("warning: inflation index %s unavailable (%v); real drawdowns omitted", sym, err)
		}
		return nil, false
	}
	return s, true
}

// deflate converts a nominal value series into real terms (base-date purchasing
// power): real_t = nominal_t × CPI_base / CPI_t, with CPI forward-filled on the
// value dates. Dates before the CPI history hold its first level (no deflation),
// so early points degrade gracefully rather than break.
func deflate(dates []time.Time, values []float64, cpi *marketdata.Series) []float64 {
	out := make([]float64, len(values))
	j, rate := 0, cpi.Points[0].Close
	var base float64
	for k, d := range dates {
		for j < len(cpi.Points) && !cpi.Points[j].Date.After(d) {
			rate = cpi.Points[j].Close
			j++
		}
		if k == 0 {
			base = rate
		}
		if rate > 0 {
			out[k] = values[k] * base / rate
		} else {
			out[k] = values[k]
		}
	}
	return out
}

// optimizedPortfolio returns a copy of base whose weights are replaced by
// the optimizer's, computed over the period where every asset has a quote.
// The original (base) keeps the weights written in the file.
func optimizedPortfolio(base *portfolio.Portfolio, spec *portfolio.Spec, bench *marketdata.Series) (*portfolio.Portfolio, string, error) {
	list := make([]*marketdata.Series, len(base.Assets))
	start := base.Assets[0].Series.First().Date
	end := base.Assets[0].Series.Last().Date
	for i, a := range base.Assets {
		list[i] = a.Series
		if f := a.Series.First().Date; f.After(start) {
			start = f
		}
		if l := a.Series.Last().Date; l.Before(end) {
			end = l
		}
	}
	if !start.Before(end) {
		return nil, "", errors.New("optimize: the assets have no period in common")
	}
	// CWARP scores the blend against a replacement portfolio, so its solver
	// also needs the benchmark's returns on the very same dates: align it
	// alongside the assets and split it back off.
	cwarpObj := spec.Optimize.Objective == optimize.CWARP
	if cwarpObj && bench == nil {
		return nil, "", errors.New("optimize: cwarp needs a benchmark (see -benchmark)")
	}
	alignList := list
	if cwarpObj {
		alignList = append(append([]*marketdata.Series{}, list...), bench)
	}
	_, prices := marketdata.Align(alignList, start, end)
	var benchReturns []float64
	if cwarpObj {
		benchReturns = metrics.Returns(prices[len(prices)-1])
		prices = prices[:len(prices)-1]
	}
	returns := make([][]float64, len(prices))
	for i, px := range prices {
		returns[i] = metrics.Returns(px)
	}
	var res optimize.Result
	var err error
	if cwarpObj {
		res, err = optimize.SolveCWARP(returns, benchReturns, *spec.Optimize)
	} else {
		res, err = optimize.Solve(returns, *spec.Optimize)
	}
	if err != nil {
		return nil, "", fmt.Errorf("optimize: %w", err)
	}

	cp := *base
	cp.Name = spec.Name + " (" + string(spec.Optimize.Objective) + ")"
	cp.Assets = make([]portfolio.Asset, len(base.Assets))
	copy(cp.Assets, base.Assets)
	parts := make([]string, len(cp.Assets))
	for i := range cp.Assets {
		cp.Assets[i].Weight = res.Weights[i]
		parts[i] = fmt.Sprintf("%s %.1f %%", cp.Assets[i].Symbol, res.Weights[i]*100)
	}
	note := fmt.Sprintf(
		"weights computed by the optimizer (%s) over %s→%s: %s, in-sample expected return %.1f %%/yr, volatility %.1f %%, Sharpe %.2f",
		spec.Optimize.Objective, start.Format("2006-01-02"), end.Format("2006-01-02"),
		strings.Join(parts, ", "), res.Return*100, res.Volatility*100, res.Sharpe)
	if cwarpObj {
		note += fmt.Sprintf(", achieved CWARP %+.1f vs %s. Note: these are the best diversifier "+
			"of %s to overlay on top of equity beta, not a standalone portfolio; its own CAGR / "+
			"volatility / drawdown below will look weak by design (that is the point): the value is "+
			"the +CWARP it adds when layered on the benchmark",
			res.CWARP, bench.Symbol, bench.Symbol)
	}
	switch spec.Optimize.Objective {
	case optimize.MaxSortino:
		note += fmt.Sprintf(", achieved Sortino %.2f", res.Sortino)
	case optimize.ReturnToDrawdown:
		note += fmt.Sprintf(", achieved return/max-drawdown %.2f", res.ReturnToMaxDD)
	case optimize.MinUlcer:
		note += fmt.Sprintf(", achieved Ulcer Index %.1f", res.Ulcer)
	case optimize.MaxWorst5y:
		note += fmt.Sprintf(", achieved worst rolling 5y return %.1f %%/yr", res.Worst5y*100)
	}
	if spec.Optimize.Objective == optimize.RiskParity && spec.Optimize.MaxWeight > 0 {
		note += " (max-weight does not apply to risk-parity)"
	}
	return &cp, note, nil
}

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

// neutralSliceColor fills the catch-all "Other" wedge of the composition
// pies; specialSliceColor fills the informative non-category wedges ("No
// country", "None (real assets)", …), a darker neutral so the two read as
// different kinds of remainder. Both stay visually distinct from the
// palette-colored slices.
const (
	neutralSliceColor = "#C6CEDA"
	specialSliceColor = "#9AA2B1"
)

// holdingsFor adapts a portfolio's assets to suggest holdings, resolving each
// identifier to its catalog metadata (aliases and SIM suffix tolerated) and
// keeping the base identifier for display.
func holdingsFor(assets []portfolio.Asset, meta map[string]suggest.Meta) []suggest.Holding {
	holdings := make([]suggest.Holding, len(assets))
	for i, a := range assets {
		base, _ := marketdata.SplitSim(a.ID)
		m, _, ok := metaFor(meta, a.ID)
		holdings[i] = suggest.Holding{ID: base, Weight: a.Weight, Meta: m, HasMeta: ok}
	}
	return holdings
}

// breakdownPies builds the look-through composition pies (geography, currency
// exposure, equity sectors, asset type) for a portfolio's detail section from
// the suggest composition splits. Returns the non-empty pie SVGs (nil when no
// metadata is available at all).
func breakdownPies(assets []portfolio.Asset, meta map[string]suggest.Meta) []template.HTML {
	if len(meta) == 0 {
		return nil
	}
	holdings := holdingsFor(assets, meta)

	geo := suggest.GeographySplit(holdings)
	foldInto(geo, "Other", suggest.BucketUnknown)

	cur := suggest.CurrencySplit(holdings)
	foldInto(cur, "Other", suggest.CurrencyOther, suggest.BucketUnknown)
	relabel(cur, suggest.CurrencyNone, "None (real assets)")
	relabel(cur, suggest.CurrencyDynamic, "Dynamic (futures)")

	sec, equity := suggest.EquitySectorSplit(holdings)
	secTitle := fmt.Sprintf("Equity sectors (%.0f%% of capital)", equity*100)

	cls := map[string]float64{}
	for class, w := range suggest.AssetClassSplit(holdings) {
		cls[prettyClass(class)] += w
	}

	svgs := []string{
		chart.Pie(chart.PieOptions{Title: "Geography"},
			breakdownSlices(geo, 8, suggest.BucketNoCountry)),
		chart.Pie(chart.PieOptions{Title: "Currency exposure"},
			breakdownSlices(cur, 8, "None (real assets)", "Dynamic (futures)")),
		chart.Pie(chart.PieOptions{Title: secTitle},
			breakdownSlices(sec, 9, suggest.BucketUnknown)),
		chart.Pie(chart.PieOptions{Title: "Asset type (look-through)"},
			breakdownSlices(cls, 8, prettyClass(suggest.BucketUnknown))),
	}
	var pies []template.HTML
	for _, s := range svgs {
		if s != "" {
			pies = append(pies, template.HTML(s))
		}
	}
	return pies
}

// foldInto merges the listed keys of a split into the dst key.
func foldInto(agg map[string]float64, dst string, keys ...string) {
	for _, k := range keys {
		if v, ok := agg[k]; ok && k != dst {
			agg[dst] += v
			delete(agg, k)
		}
	}
}

// relabel renames a split key, merging with any existing value.
func relabel(agg map[string]float64, from, to string) {
	if v, ok := agg[from]; ok {
		agg[to] += v
		delete(agg, from)
	}
}

// breakdownSlices turns an aggregation map into pie slices: largest first,
// wedges below 3 % and the literal "Other" key merged into a trailing neutral
// "Other" slice, capped at maxSlices colored entries. The special labels
// (informative non-categories like "No country") are pinned after it in the
// given order, in a darker neutral. A pie carrying no colored slice at all
// (no real composition) returns nil so it is omitted.
func breakdownSlices(agg map[string]float64, maxSlices int, special ...string) []chart.Slice {
	specialSet := map[string]bool{}
	for _, s := range special {
		specialSet[s] = true
	}
	type kv struct {
		k string
		v float64
	}
	items := make([]kv, 0, len(agg))
	total, other := 0.0, 0.0
	for k, v := range agg {
		total += v
		if k == "Other" {
			other += v
			continue
		}
		if !specialSet[k] {
			items = append(items, kv{k, v})
		}
	}
	if total <= 0 {
		return nil
	}
	sort.Slice(items, func(i, j int) bool { return items[i].v > items[j].v })
	slices := make([]chart.Slice, 0, maxSlices)
	for _, it := range items {
		if it.v/total < 0.03 || len(slices) >= maxSlices-1 {
			other += it.v
			continue
		}
		slices = append(slices, chart.Slice{Label: it.k, Value: it.v})
	}
	if len(slices) == 0 {
		return nil // only remainders: nothing to show
	}
	if other > 0 {
		slices = append(slices, chart.Slice{Label: "Other", Value: other, Color: neutralSliceColor})
	}
	for _, s := range special {
		if v := agg[s]; v > 0 {
			slices = append(slices, chart.Slice{Label: s, Value: v, Color: specialSliceColor})
		}
	}
	return slices
}

// prettyClass turns a catalog asset_class slug ("aggregate-bond") into a
// display label ("Aggregate bond").
func prettyClass(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// coverageBars computes a portfolio's macro-regime coverage for the report,
// each bar split into per-holding segments (stable color per holding across
// the rows) with a compact contributor line beneath. It returns nil when no
// asset carries metadata (nothing meaningful to show).
func coverageBars(assets []portfolio.Asset, meta map[string]suggest.Meta, fw suggest.Framework) []report.CoverageBar {
	holdings := holdingsFor(assets, meta)
	anyMeta := false
	for _, h := range holdings {
		anyMeta = anyMeta || h.HasMeta
	}
	if !anyMeta {
		return nil
	}
	cov, _ := suggest.Coverage(holdings, fw)
	contrib := suggest.Contributors(holdings, fw)
	gapSet := map[suggest.Category]bool{}
	for _, g := range suggest.Gaps(cov, fw, suggest.DefaultOptions().GapThreshold) {
		gapSet[g] = true
	}
	bars := make([]report.CoverageBar, 0, len(fw.Categories))
	for _, rg := range fw.Categories {
		// The track represents max(coverage, 100 %): segments stay
		// proportional even when notional coverage exceeds the portfolio.
		scale := math.Max(cov[rg], 1)
		var segs []report.CoverageSeg
		var parts []string
		for _, c := range contrib[rg] {
			segs = append(segs, report.CoverageSeg{
				Width: math.Round(c.Weight/scale*1000) / 10,
				Color: chart.PaletteColor(c.Index),
				Title: fmt.Sprintf("%s %.0f%%", c.ID, c.Weight*100),
			})
			parts = append(parts, fmt.Sprintf("%s %.0f", c.ID, c.Weight*100))
		}
		bars = append(bars, report.CoverageBar{
			Regime:   string(rg),
			Pct:      int(cov[rg]*100 + 0.5),
			Gap:      gapSet[rg],
			Segments: segs,
			Detail:   strings.Join(parts, " · "),
		})
	}
	return bars
}

// compositionNotes renders the look-through duration and currency summary
// lines shown under a portfolio's composition (empty without metadata).
func compositionNotes(assets []portfolio.Asset, meta map[string]suggest.Meta, base string) []string {
	if len(meta) == 0 {
		return nil
	}
	holdings := holdingsFor(assets, meta)
	var notes []string

	led := suggest.DurationSplit(holdings)
	switch {
	case led.Nominal > 0:
		line := fmt.Sprintf("Rate duration (look-through): %.1f y nominal per unit of capital (≈ %.0f pts of 7y-bond equivalent)",
			led.Nominal, led.Nominal/7*100)
		if led.Real > 0 {
			line += fmt.Sprintf(", plus %.1f y real-rate from inflation-linked bonds", led.Real)
		}
		if led.Missing > 0.02 {
			line += fmt.Sprintf("; no duration figure for %.0f%% of the bond notional", led.Missing*100)
		}
		notes = append(notes, line+".")
	case led.Real > 0:
		notes = append(notes, fmt.Sprintf("Rate duration (look-through): %.1f y real-rate from inflation-linked bonds.", led.Real))
	}

	p := suggest.CurrencyProfile(suggest.CurrencySplit(holdings), base)
	if p.Base+p.Foreign+p.NonFiat > 0 {
		line := fmt.Sprintf("Currency (look-through): %.0f%% %s-native or hedged · %.0f%% unhedged foreign", p.Base*100, base, p.Foreign*100)
		if p.Top != "" {
			line += fmt.Sprintf(" (mostly %s, %.0f%%)", p.Top, p.TopShare*100)
		}
		if p.NonFiat > 0 {
			line += fmt.Sprintf(" · %.0f%% non-fiat or futures-driven", p.NonFiat*100)
		}
		notes = append(notes, line+".")
	}
	return notes
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

// negate returns a sign-flipped copy: portfolio flows (contributions
// positive) become investor flows (money out of pocket negative).
func negate(xs []float64) []float64 {
	out := make([]float64, len(xs))
	for i, x := range xs {
		out[i] = -x
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

// window returns the bounds [i, j) of dates within [from, to].
func window(dates []time.Time, from, to time.Time) (int, int) {
	i := sort.Search(len(dates), func(k int) bool { return !dates[k].Before(from) })
	j := sort.Search(len(dates), func(k int) bool { return dates[k].After(to) })
	return i, j
}

// rebase rescales a value slice so that it starts at 100.
func rebase(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = v / values[0] * 100
	}
	return out
}

func seriesSlices(s *marketdata.Series) ([]time.Time, []float64) {
	dates := make([]time.Time, len(s.Points))
	values := make([]float64, len(s.Points))
	for i, p := range s.Points {
		dates[i] = p.Date
		values[i] = p.Close
	}
	return dates, values
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

// runPermanent backtests the tactical Permanent Portfolio 2.0 (pkg/permanent)
// against the static Browne PP and MSCI World, all in real terms, then reports
// the decumulation ruin probabilities that matter for FIRE. It fetches the four
// sleeves, deflates them to real monthly returns, drives the allocation from the
// embedded macro panel, and block-bootstraps the realized tactical and static
// return streams through the decumul engine.
func runPermanent(ctx context.Context, opt *options, c *marketdata.Client) error {
	monthEnd := func(s *marketdata.Series) map[time.Time]float64 {
		out := map[time.Time]float64{}
		for _, p := range s.Points {
			out[time.Date(p.Date.Year(), p.Date.Month(), 1, 0, 0, 0, 0, time.UTC)] = p.Close
		}
		return out
	}
	fetchX := func(id string) (map[time.Time]float64, error) {
		s, err := c.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "USD"})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		return monthEnd(s), nil
	}
	fetchLevel := func(id string) (map[time.Time]float64, error) {
		s, err := c.Fetch(ctx, id, time.Time{})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		return monthEnd(s), nil
	}
	cpi, err := fetchLevel("^CPI-US")
	if err != nil {
		return err
	}
	eqM, err := fetchX("URTHSIM")
	if err != nil {
		return err
	}
	boM, err := fetchX("TLTSIM")
	if err != nil {
		return err
	}
	goM, err := fetchX("XAUUSDSIM")
	if err != nil {
		return err
	}
	irx, err := fetchLevel("^IRX")
	if err != nil {
		return err
	}

	var months []time.Time
	for m := range cpi {
		months = append(months, m)
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Before(months[j]) })

	// nominal cash index accrued from the short rate, then deflated like the rest.
	cashIdx := map[time.Time]float64{}
	cash := 1.0
	var prev time.Time
	for _, m := range months {
		if !prev.IsZero() {
			if y, ok := irx[prev]; ok {
				cash *= 1 + y/100/12
			}
		}
		cashIdx[m] = cash
		prev = m
	}
	realRet := func(nom map[time.Time]float64, pr, m time.Time) (float64, bool) {
		n0, o1 := nom[pr]
		n1, o2 := nom[m]
		p0, o3 := cpi[pr]
		p1, o4 := cpi[m]
		if !(o1 && o2 && o3 && o4) || n0 == 0 || p0 == 0 || p1 == 0 {
			return 0, false
		}
		return (n1/p1)/(n0/p0) - 1, true
	}

	var ar permanent.AssetReturns
	for i := 1; i < len(months); i++ {
		m, pr := months[i], months[i-1]
		e, o1 := realRet(eqM, pr, m)
		b, o2 := realRet(boM, pr, m)
		ca, o3 := realRet(cashIdx, pr, m)
		g, o4 := realRet(goM, pr, m)
		if !(o1 && o2 && o3 && o4) {
			continue
		}
		ar.Dates = append(ar.Dates, m)
		ar.Equity = append(ar.Equity, e)
		ar.Bonds = append(ar.Bonds, b)
		ar.Cash = append(ar.Cash, ca)
		ar.Gold = append(ar.Gold, g)
	}
	if len(ar.Dates) < 120 {
		return fmt.Errorf("permanent: too few aligned months (%d)", len(ar.Dates))
	}

	panel, err := permanent.LoadPanel()
	if err != nil {
		return err
	}
	regimes := panel.Regimes(ar.Dates[0].AddDate(-1, 0, 0), ar.Dates[len(ar.Dates)-1], permanent.DefaultSignalConfig())
	res, err := permanent.Simulate(regimes, ar, permanent.DefaultParams())
	if err != nil {
		return err
	}

	// equity-only real returns aligned to the backtest dates, for a benchmark row.
	eqByDate := make(map[time.Time]float64, len(ar.Dates))
	for i, d := range ar.Dates {
		eqByDate[d] = ar.Equity[i]
	}
	eqReal := make([]float64, len(res.Dates))
	for i, d := range res.Dates {
		eqReal[i] = eqByDate[d]
	}

	fmt.Printf("Tactical Permanent Portfolio 2.0 (Darcet), REAL, monthly, %s..%s\n",
		res.Dates[0].Format("2006-01"), res.Dates[len(res.Dates)-1].Format("2006-01"))
	fmt.Printf("Reconstruction of an undisclosed method; see docs/darcet-permanent-portfolio-design.md\n\n")
	fmt.Printf("%-24s %7s %6s %7s %5s %7s\n", "portfolio", "CAGR", "vol", "maxDD", "%UW", "longUW")
	statRow := func(name string, series []float64) {
		s := permanent.Compute(series)
		fmt.Printf("%-24s %6.2f%% %5.1f%% %6.1f%% %4.0f%% %5.1fy\n",
			name, s.CAGR*100, s.Vol*100, s.MaxDrawdown*100, s.UnderwaterFraction*100, float64(s.LongestUnderwater)/12)
	}
	statRow("tactical PP 2.0", res.Tactical)
	statRow("static Browne PP", res.Static)
	statRow("MSCI World (equity)", eqReal)

	const years = 40
	rates := []float64{0.030, 0.035, 0.040, 0.045}
	ruin := func(series []float64, wr float64) float64 {
		src := scenario.StationaryBootstrap{
			Panel:     scenario.Panel{Returns: [][]float64{series}, Weights: []float64{1}},
			MeanBlock: 24,
			Periods:   years * 12,
		}
		plan := decumul.Plan{
			Capital: 1, NeedAnnual: wr, Years: years,
			Source: src, Monthly: true, Tax: decumul.CTOFlatTax{Rate: 0},
		}
		return plan.Simulate(3000, runtime.NumCPU(), 1).Outcome().RuinProb
	}
	fmt.Printf("\n%d-year ruin probability at a fixed real withdrawal (stationary bootstrap):\n", years)
	fmt.Printf("%-24s", "withdrawal rate")
	for _, wr := range rates {
		fmt.Printf(" %6.1f%%", wr*100)
	}
	fmt.Println()
	ruinRow := func(name string, series []float64) {
		fmt.Printf("%-24s", name)
		for _, wr := range rates {
			fmt.Printf(" %6.1f%%", ruin(series, wr)*100)
		}
		fmt.Println()
	}
	ruinRow("tactical PP 2.0", res.Tactical)
	ruinRow("static Browne PP", res.Static)
	fmt.Println("\nRuin = share of 40-year retirements exhausted. The realized real series is")
	fmt.Println("block-bootstrapped (mean block 24 months): one historical path, no tax or fees.")
	_ = opt
	return nil
}
