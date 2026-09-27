//go:build ignore

// Pair answers "does this series match its reference?": a reconstruction
// against the fund it rebuilds, a fund against the index it tracks, a share
// class against its sibling. It prints analyze.Pair's study: the level gap
// and its standard error, the monthly (and, when both are daily, the daily)
// correlation, tracking error and beta, the largest divergences, dated, and
// the calendar years side by side.
//
// Usage:
//
//	go run examples/code/pair.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-k 5] [-lead-lag] [-json] CANDIDATE REFERENCE
//
// Example:
//
//	go run examples/code/pair.go IWDA MSCIWORLD-USD
//
// Each side goes through marketdata.Client.Load, so it may be a CSV path as
// well as an identifier. "pofo -pair A,B" is the same on the command line.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert both sides into this currency (default: each one's own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	k := flag.Int("k", analyze.DefaultDivergences, "largest divergences listed per calendar")
	leadLag := flag.Bool("lead-lag", false, "forgive a one-session difference in closing times (a European close against a US index)")
	asJSON := flag.Bool("json", false, "print the study as JSON")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("usage: pair [flags] CANDIDATE REFERENCE")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}
	var sides [2]*marketdata.Series
	for i, id := range flag.Args() {
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		sides[i] = s
	}
	// A reference series states no currency; when the candidate does, say
	// they agree rather than read a currency warning (convert first when
	// they do not).
	if sides[1].Currency == "" {
		sides[1].Currency = sides[0].Currency
	}

	st, err := analyze.Pair(sides[0], sides[1], analyze.PairOptions{Divergences: *k, LeadLag: *leadLag})
	if err != nil {
		log.Fatal(err)
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(st); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := st.WriteText(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// date parses a YYYY-MM-DD flag; empty is the zero time, an open bound.
func date(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		log.Fatalf("bad date %q: want YYYY-MM-DD", s)
	}
	return t
}
