//go:build ignore

// Oldnew answers "did a data refresh move the past?": a bundled series
// against an older version of its own file, read out of git, with the date
// from which the two part ways (a rescaled copy forgiven) and the largest
// divergences, dated.
//
// Usage:
//
//	go run examples/code/oldnew.go [-rev HEAD~10] [-k 5] ID
//
// Example:
//
//	go run examples/code/oldnew.go -rev HEAD~20 TREASURY-LONG-USD
//
// It runs "git show", so it needs a checkout of the repository and runs
// from its root. ID must be bundled (pofo -dump list names them).
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

func main() {
	log.SetFlags(0)
	rev := flag.String("rev", "HEAD~10", "the git revision to compare with")
	k := flag.Int("k", analyze.DefaultDivergences, "largest divergences listed per calendar")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: oldnew [flags] ID")
	}

	// The version the module embeds today.
	cur, err := marketdata.Bundled(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	// The same file at rev: its Source names the directory it lives in and
	// its Symbol is the canonical identifier the file is named after.
	path := fmt.Sprintf("pkg/datasets/%s/%s.csv", cur.Source, cur.Symbol)
	out, err := exec.Command("git", "show", *rev+":"+path).Output()
	if err != nil {
		log.Fatalf("git show %s:%s: %v", *rev, path, err)
	}
	// ReadCSV reads any "date,value" file, the bundled headers included.
	old, err := marketdata.ReadCSV(bytes.NewReader(out), cur.Symbol+"@"+*rev)
	if err != nil {
		log.Fatal(err)
	}
	old.Currency = cur.Currency // a file does not state it; Bundled reads it off the catalog

	// Pair's identity line says whether the two agree, and from which date
	// the new file departs from the old one.
	st, err := analyze.Pair(cur, old, analyze.PairOptions{Divergences: *k})
	if err != nil {
		log.Fatal(err)
	}
	if err := st.WriteText(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
