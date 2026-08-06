// Command simgen (re)builds the permanent simulated histories stored under
// simdata/. Each bundled recipe reconstructs a complex asset (90/60 funds,
// managed-futures ETFs, …) from long-history components, validates the
// result against the asset's real quotes and writes a self-describing CSV.
//
// Usage:
//
//	simgen [-data data] [-simdata simdata] [id …]
//
// Without arguments every bundled recipe runs; otherwise only those whose
// identifier matches.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"portfodor/pkg/marketdata"
	"portfodor/pkg/simgen"
)

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:]); err != nil {
		log.Fatal("simgen: ", err)
	}
}

func run(argv []string) error {
	fs := flag.NewFlagSet("simgen", flag.ContinueOnError)
	dataDir := fs.String("data", "data", "répertoire de cache des cotations")
	simdataDir := fs.String("simdata", "simdata", "répertoire des données simulées permanentes")
	refdataDir := fs.String("refdata", "refdata", "répertoire des séries de référence importées")
	dry := fs.Bool("dry", false, "valider sans écrire les fichiers")
	if err := fs.Parse(argv); err != nil {
		return err
	}

	client := marketdata.NewClient(*dataDir)
	client.Logf = log.Printf

	recipes := simgen.All()
	if args := fs.Args(); len(args) > 0 {
		recipes = recipes[:0]
		for _, id := range args {
			r, ok := simgen.Find(id)
			if !ok {
				return fmt.Errorf("aucune recette pour %q", id)
			}
			recipes = append(recipes, r)
		}
	}

	from := simgen.ComponentsFrom
	fetcher := simgen.WithRefData(*refdataDir, client)
	failures := 0
	for _, r := range recipes {
		err := generate(client, fetcher, *simdataDir, r, from, *dry)
		switch {
		case errors.Is(err, simgen.ErrUnfaithful):
			log.Printf("⚠ %-14s ignoré: %v", r.ID, err)
		case err != nil:
			log.Printf("✗ %-14s %v", r.ID, err)
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d recette(s) en échec", failures)
	}
	return nil
}

func generate(client *marketdata.Client, fetcher simgen.Fetcher, dir string, r simgen.Recipe, from time.Time, dry bool) error {
	sim, err := r.Build(fetcher, from)
	if err != nil {
		return err
	}
	validation := "non validé (pas de série réelle)"
	if r.ValidateAgainst != "" {
		real, err := client.Fetch(r.ValidateAgainst, from)
		if err != nil {
			return fmt.Errorf("série réelle %s: %w", r.ValidateAgainst, err)
		}
		v, err := simgen.Validate(sim, real)
		if err != nil {
			return fmt.Errorf("validation vs %s: %w", r.ValidateAgainst, err)
		}
		validation = fmt.Sprintf("%s vs %s", v, r.ValidateAgainst)
	}
	if r.SpliceReal != "" {
		real, err := client.Fetch(r.SpliceReal, from)
		if err != nil {
			return fmt.Errorf("série à greffer %s: %w", r.SpliceReal, err)
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
