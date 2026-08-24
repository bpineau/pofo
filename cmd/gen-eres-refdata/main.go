// Command gen-eres-refdata refreshes the bundled NAV snapshot of the catalog's
// "airfund" funds, the French employee-savings funds (FCPE) whose only public
// price is the NAV history behind their management company's fund page.
//
// For each such fund it downloads the whole official NAV series from the
// airfund.io delivery API (the same call the pofo binary makes live), checks
// it, and writes pkg/datasets/refdata/<ID>-NAV.csv. That file serves two
// purposes: it is the offline fallback of the live source (marketdata serves
// it when the API cannot be reached and nothing is cached), and it is the
// real series the fund's reconstruction recipe splices its backcast behind
// at simdata-generation time, so the two never disagree.
//
// It runs at data-generation time only (network); refresh it whenever the
// bundled tail should catch up with the fund (the live path does not depend
// on it).
//
// Usage: gen-eres-refdata [-dir path] [-dry] [ID...]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

func main() {
	dir := flag.String("dir", "pkg/datasets/refdata", "directory the NAV snapshots are written to")
	dry := flag.Bool("dry", false, "download and validate, write nothing")
	flag.Parse()

	ids := flag.Args()
	if len(ids) == 0 {
		for _, a := range datasets.Catalog() {
			if a.Source == "airfund" {
				ids = append(ids, a.ID)
			}
		}
	}
	if len(ids) == 0 {
		log.Fatal("no airfund fund in the catalog")
	}

	ctx := context.Background()
	client := marketdata.NewClient("") // no cache: the point is the live series
	client.Logf = func(format string, args ...any) { log.Printf(format, args...) }
	today := time.Now().UTC().Format("2006-01-02")
	failed := false
	for _, id := range ids {
		if err := refresh(ctx, client, *dir, id, today, *dry); err != nil {
			log.Printf("✗ %s: %v", id, err)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

// refresh downloads, validates and writes one fund's NAV snapshot.
func refresh(ctx context.Context, client *marketdata.Client, dir, id, today string, dry bool) error {
	a, ok := marketdata.Lookup(id)
	if !ok || a.Source != "airfund" {
		return fmt.Errorf("not a catalog airfund fund")
	}
	s, err := client.Fetch(ctx, id, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return err
	}
	if s.Source != "airfund" {
		return fmt.Errorf("served from %s, not the live API: nothing to snapshot", s.Source)
	}
	if err := validate(a, s); err != nil {
		return err
	}
	log.Printf("✓ %-12s %d NAVs %s → %s (last %.2f %s)", id, len(s.Points),
		s.First().Date.Format("2006-01-02"), s.Last().Date.Format("2006-01-02"), s.Last().Close, s.Currency)
	if dry {
		return nil
	}
	return marketdata.WriteSimdata(dir, &marketdata.SimdataFile{
		ID:   id + "-NAV",
		Name: a.Name + " (official daily NAV, " + a.Currency + ")",
		Method: fmt.Sprintf("official NAV history from the airfund.io delivery API (share code %s, widget %s), "+
			"as published by the management company; offline fallback of the live source and real segment of the %s recipe",
			a.Symbol, a.Xid, id),
		Validation: fmt.Sprintf("%d NAVs, %s → %s, checked against the catalog inception and a 15%% daily-move bound",
			len(s.Points), s.First().Date.Format("2006-01-02"), s.Last().Date.Format("2006-01-02")),
		Generated: today,
		Points:    s.Points,
	})
}

// validate refuses a series that cannot be the fund's own NAV history: too
// short, not starting at the catalog inception, out of order, non-positive,
// carrying an implausible daily move for a long-only fund, or stale.
func validate(a datasets.Asset, s *marketdata.Series) error {
	if len(s.Points) < 100 {
		return fmt.Errorf("only %d NAVs", len(s.Points))
	}
	if a.Since != "" {
		since, err := time.Parse("2006-01-02", a.Since)
		if err == nil && (s.First().Date.Before(since) || s.First().Date.Sub(since) > 30*24*time.Hour) {
			return fmt.Errorf("first NAV %s does not match the catalog inception %s",
				s.First().Date.Format("2006-01-02"), a.Since)
		}
	}
	for i, p := range s.Points {
		if p.Close <= 0 {
			return fmt.Errorf("%s: non-positive NAV %v", p.Date.Format("2006-01-02"), p.Close)
		}
		if i == 0 {
			continue
		}
		prev := s.Points[i-1]
		if !p.Date.After(prev.Date) {
			return fmt.Errorf("%s: dates out of order", p.Date.Format("2006-01-02"))
		}
		if move := math.Abs(math.Log(p.Close / prev.Close)); move > 0.15 {
			return fmt.Errorf("%s: %.1f%% daily move", p.Date.Format("2006-01-02"), (math.Exp(move)-1)*100)
		}
	}
	if age := time.Since(s.Last().Date); age > 45*24*time.Hour {
		return fmt.Errorf("last NAV %s is %.0f days old: degraded feed?", s.Last().Date.Format("2006-01-02"), age.Hours()/24)
	}
	return nil
}
