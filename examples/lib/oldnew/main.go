// Oldnew compares a bundled series with an older version of its own file,
// the check a data refresh calls for: did the new file move the past, and
// where?
//
//	go run ./examples/lib/oldnew [-rev HEAD~10] [ID]
//
// It runs "git show", so it needs a checkout of the repository; the default
// ID is TREASURY-LONG-USD.
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
	flag.Parse()
	id := "TREASURY-LONG-USD"
	if flag.NArg() > 0 {
		id = flag.Arg(0)
	}

	// The version the module embeds today.
	cur, err := marketdata.Bundled(id)
	if err != nil {
		log.Fatal(err)
	}

	// The same file at rev: its Source says which directory it lives in and
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
	// the new file departs from the old one (a rescaled copy forgiven).
	st, err := analyze.Pair(cur, old, analyze.PairOptions{Divergences: 3})
	if err != nil {
		log.Fatal(err)
	}
	if err := st.WriteText(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
