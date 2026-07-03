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
	"github.com/bpineau/pofo/pkg/decumul/web"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/optimize"
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
	var endStr string
	fs.StringVar(&endStr, "end", "", "end date (YYYY-MM-DD, default: last available quote)")
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
	genSimdata := fs.Bool("gen-simdata", false, "(re)generate the simulated histories (recipes as arguments, default: all) then stop; rebuild afterwards to re-embed them")
	dry := fs.Bool("dry", false, "with -gen-simdata: validate without writing")
	refdataDir := fs.String("refdata", "", "dev override: directory of extra local reference CSVs for -gen-simdata")
	assetsList := fs.String("assets", "", "comma-separated list of tickers/ISINs, each compared as a portfolio 100 % invested in it")
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
                             min-volatility or risk-parity, with an
                             optional ",max-weight:40" cap. The report
                             shows the optimized weights next to the
                             written ones.

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
	if len(files) == 0 && *assetsList == "" && !*warmup && !*genSimdata && !*verifyData && !*suggestFlag && !*coverageFlag && !*fireFlag {
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
	if len(specs) == 0 && !*warmup && !*verifyData && !*suggestFlag && !*coverageFlag && !*fireFlag {
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

	// Download every distinct asset once.
	seriesByID := map[string]*marketdata.Series{}
	for _, spec := range specs {
		for _, h := range spec.Holdings {
			if _, ok := seriesByID[h.ID]; ok {
				continue
			}
			s, err := fetchAsset(ctx, client, h.ID, &opt)
			if err != nil {
				return fmt.Errorf("portfolio %s, asset %q: %w", spec.Name, h.ID, err)
			}
			seriesByID[h.ID] = s
		}
	}

	// Benchmark for Beta, best effort.
	var bench *marketdata.Series
	if opt.benchmark != "" {
		b, err := client.FetchExtended(ctx, opt.benchmark, marketdata.FetchOptions{
			From: opt.start, NoSim: true, Currency: opt.currency,
		})
		if err != nil {
			log.Printf("warning: benchmark %s unavailable (no Beta): %v", opt.benchmark, err)
		} else {
			bench = b
		}
	}

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
	simulateInto := func(p *portfolio.Portfolio, spec *portfolio.Spec) error {
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
		results = append(results, &result{p: p, sim: sim, color: chart.PaletteColor(len(results)), rebalanceDays: days})
		return nil
	}
	for _, spec := range specs {
		p, err := portfolio.Build(spec, portfolio.BuildOptions{
			Fetch:        func(id string) (*marketdata.Series, error) { return seriesByID[id], nil },
			Fees:         feesFor,
			Cash:         cashRate,
			BorrowSpread: 1.0, // default: cash + 1 %/yr
			BaseCurrency: opt.currency,
		})
		if err != nil {
			return err
		}
		// An optimized portfolio is shown next to its written weights, so
		// the optimizer's choice can be compared with the baseline.
		if spec.Optimize != nil {
			pOpt, note, err := optimizedPortfolio(p, spec)
			if err != nil {
				return fmt.Errorf("portfolio %s: %w", spec.Name, err)
			}
			p.Name = spec.Name + " (as written)"
			if err := simulateInto(p, spec); err != nil {
				return err
			}
			if err := simulateInto(pOpt, spec); err != nil {
				return err
			}
			results[len(results)-1].note = note
			continue
		}
		if err := simulateInto(p, spec); err != nil {
			return err
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
	var benchDates []time.Time
	var benchValues []float64
	if bench != nil {
		benchDates, benchValues = seriesSlices(bench)
	}
	// Consumer-price index of the base currency, to report drawdowns/TTR in real
	// (purchasing-power) terms alongside the nominal ones. Best-effort: absent it,
	// the real columns are simply omitted.
	deflator, hasDeflator := inflationSeries(ctx, client, opt.currency, commonStart)
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
		if hasDeflator {
			if rs, err := metrics.Compute(r.winDates, deflate(r.winDates, r.winValues, deflator)); err == nil {
				r.realStats, r.hasReal = rs, true
			}
		}
		if bench != nil {
			if rel, ok := metrics.VsBenchmark(r.winDates, r.winValues, benchDates, benchValues); ok {
				st.Beta, st.HasBeta = rel.Beta, true
				r.rel, r.hasRel = rel, true
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
	var panel *scenario.Panel
	var labels []string
	if len(specs) > 0 {
		var assets []web.AssetSeries
		for _, h := range specs[0].Holdings {
			s, err := fetchAsset(ctx, c, h.ID, opt)
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
// conversion, window) for one asset, from the CLI options.
func fetchAsset(ctx context.Context, c *marketdata.Client, id string, opt *options) (*marketdata.Series, error) {
	return c.FetchExtended(ctx, id, marketdata.FetchOptions{
		From:     opt.start,
		To:       opt.end,
		NoSim:    opt.noSim,
		Simdata:  opt.simdata,
		Currency: opt.currency,
	})
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
func optimizedPortfolio(base *portfolio.Portfolio, spec *portfolio.Spec) (*portfolio.Portfolio, string, error) {
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
	_, prices := marketdata.Align(list, start, end)
	returns := make([][]float64, len(prices))
	for i, px := range prices {
		returns[i] = metrics.Returns(px)
	}
	res, err := optimize.Solve(returns, *spec.Optimize)
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
	if spec.Optimize.Objective == optimize.RiskParity && spec.Optimize.MaxWeight > 0 {
		note += " (max-weight does not apply to risk-parity)"
	}
	return &cp, note, nil
}

func buildPage(results []*result, opt *options, bench *marketdata.Series, commonStart, commonEnd time.Time, meta map[string]suggest.Meta) *report.Page {
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.p.Name
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
		section.Breakdowns = breakdownPies(r.p.Assets, meta)
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
			if u, known := marketdata.GuessUCITS(base, a.Name); known {
				ucitsText = map[bool]string{true: "yes", false: "no"}[u]
			}
			assetClass := ""
			if m, _, ok := metaFor(meta, a.ID); ok {
				assetClass = m.AssetClass
			}
			section.Assets = append(section.Assets, report.AssetRow{
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
				Note: strings.Join(notes, "; "),
			})
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

	if opt.currency != "" {
		page.Footnotes = append(page.Footnotes, fmt.Sprintf(
			"All series converted to %s (daily Yahoo FX crosses; the earliest known rate is held constant before the FX history starts).", opt.currency))
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
			"Up / Down capture: the portfolio's average return on the benchmark's up (resp. down) days, as a % of the benchmark's own average on those days. Up capture above 100 % amplifies rallies; Down capture below 100 % cushions losses. The ideal profile is high up / low down (e.g. 95 % / 70 %).")
	}
	var hasBreakdowns, hasCoverage bool
	for _, s := range page.Portfolios {
		hasBreakdowns = hasBreakdowns || len(s.Breakdowns) > 0
		hasCoverage = hasCoverage || len(s.Coverage) > 0
	}
	if hasBreakdowns {
		page.Footnotes = append(page.Footnotes,
			"Composition pies (per portfolio): geography, sector and asset type, each holding's published breakdown weighted by its portfolio weight. Holdings without a geography or sector split (bonds, gold, managed futures…) are pooled into a neutral \"Other\" wedge, so each pie still totals the whole portfolio.")
	}
	if hasCoverage {
		page.Footnotes = append(page.Footnotes,
			"Macro-regime coverage: notional exposure to each growth/inflation environment (an asset can span several; leveraged stacked funds count each leg's notional, so bars can exceed 100%); a low bar is a gap. Run \"-suggest\" for assets to fill it.")
	}
	return page
}

// neutralSliceColor fills the catch-all "Other" wedge of the composition pies,
// keeping it visually distinct from the palette-colored slices.
const neutralSliceColor = "#C6CEDA"

// breakdownPies builds the composition pies (geography, sector, asset type)
// for a portfolio's detail section, each holding's split weighted by its
// portfolio weight. Holdings whose catalog metadata lacks a geography or
// sector split (bonds, gold, managed futures…) fall into a single "Other /
// N/A" wedge so every pie still sums to the whole portfolio. Returns the
// non-empty pie SVGs (nil when no metadata is available at all).
func breakdownPies(assets []portfolio.Asset, meta map[string]suggest.Meta) []template.HTML {
	if len(meta) == 0 {
		return nil
	}
	const naLabel = "Other / N/A"
	geo := map[string]float64{}
	sec := map[string]float64{}
	cls := map[string]float64{}
	for _, a := range assets {
		if a.Weight <= 0 {
			continue
		}
		m, _, ok := metaFor(meta, a.ID)
		if !ok {
			geo[naLabel] += a.Weight
			sec[naLabel] += a.Weight
			cls["Unknown"] += a.Weight
			continue
		}
		addBreakdown(geo, m.Geography, a.Weight, naLabel, canonRegion)
		addBreakdown(sec, m.Sectors, a.Weight, naLabel, nil)
		if m.AssetClass != "" {
			cls[prettyClass(m.AssetClass)] += a.Weight
		} else {
			cls["Unknown"] += a.Weight
		}
	}
	svgs := []string{
		chart.Pie(chart.PieOptions{Title: "Geography"}, breakdownSlices(geo, naLabel, 8)),
		chart.Pie(chart.PieOptions{Title: "Sector"}, breakdownSlices(sec, naLabel, 9)),
		chart.Pie(chart.PieOptions{Title: "Asset type"}, breakdownSlices(cls, "Unknown", 8)),
	}
	var pies []template.HTML
	for _, s := range svgs {
		if s != "" {
			pies = append(pies, template.HTML(s))
		}
	}
	return pies
}

// addBreakdown adds weight w split over an asset's percentage map (values
// summing to ~100); an empty map sends the whole weight to the N/A wedge.
// norm, when set, canonicalizes keys so synonymous labels merge.
func addBreakdown(agg, split map[string]float64, w float64, naLabel string, norm func(string) string) {
	if len(split) == 0 {
		agg[naLabel] += w
		return
	}
	for k, v := range split {
		if norm != nil {
			k = norm(k)
		}
		agg[k] += w * v / 100
	}
}

// canonRegion merges the obvious country-label synonyms used inconsistently
// across the catalog (the granular regions like "North America" are kept).
func canonRegion(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "united states", "usa", "u.s.", "u.s.a.", "united states of america":
		return "US"
	case "united kingdom", "u.k.", "great britain":
		return "UK"
	}
	return s
}

// breakdownSlices turns an aggregation map into pie slices: largest first,
// wedges below 3 % (and the explicit N/A bucket) merged into a trailing
// neutral "Other" slice, capped at maxSlices entries. A pie that would carry
// only the neutral slice (no real composition) returns nil so it is omitted.
func breakdownSlices(agg map[string]float64, naLabel string, maxSlices int) []chart.Slice {
	total := 0.0
	for _, v := range agg {
		total += v
	}
	if total <= 0 {
		return nil
	}
	type kv struct {
		k string
		v float64
	}
	items := make([]kv, 0, len(agg))
	other := 0.0
	for k, v := range agg {
		if k == naLabel {
			other += v
			continue
		}
		items = append(items, kv{k, v})
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
	if other > 0 {
		slices = append(slices, chart.Slice{Label: "Other", Value: other, Color: neutralSliceColor})
	}
	if len(slices) == 1 && slices[0].Color == neutralSliceColor {
		return nil // only "Other": nothing to show
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

// coverageBars computes a portfolio's macro-regime coverage for the report.
// It returns nil when no asset carries metadata (nothing meaningful to show).
func coverageBars(assets []portfolio.Asset, meta map[string]suggest.Meta, fw suggest.Framework) []report.CoverageBar {
	holdings := make([]suggest.Holding, len(assets))
	anyMeta := false
	for i, a := range assets {
		m, _, ok := metaFor(meta, a.ID)
		holdings[i] = suggest.Holding{ID: a.ID, Weight: a.Weight, Meta: m, HasMeta: ok}
		anyMeta = anyMeta || ok
	}
	if !anyMeta {
		return nil
	}
	cov, _ := suggest.Coverage(holdings, fw)
	gapSet := map[suggest.Category]bool{}
	for _, g := range suggest.Gaps(cov, fw, suggest.DefaultOptions().GapThreshold) {
		gapSet[g] = true
	}
	bars := make([]report.CoverageBar, 0, len(fw.Categories))
	for _, rg := range fw.Categories {
		pct := int(cov[rg]*100 + 0.5)
		width := min(pct, 100)
		bars = append(bars, report.CoverageBar{Regime: string(rg), Pct: pct, Width: width, Gap: gapSet[rg]})
	}
	return bars
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
