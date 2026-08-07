// The -suggest and -coverage modes: macro-regime/factor coverage analysis
// and out-of-sample-validated gap-filling suggestions (pkg/suggest wiring).
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

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
