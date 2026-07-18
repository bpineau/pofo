// Data-maintenance modes: -gen-simdata (rebuild simulated histories),
// -verify-data (the data doctor) and -warmup (pre-fill the quote cache).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/simgen"
)

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
