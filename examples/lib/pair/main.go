// Pair measures a reconstruction against its reference: the bundled
// backcast of a world-equity fund against the MSCI World index it tracks.
//
//	go run ./examples/lib/pair [-json] [CANDIDATE REFERENCE]
//
// The default pair is IWDA (iShares Core MSCI World, IE00B4L5Y983), whose
// bundled history is the index less the fund's charge before its launch and
// the fund's real quotes after, against MSCIWORLD-USD. Any two bundled
// identifiers work; "pofo -pair A B" is the same on the command line.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

func main() {
	log.SetFlags(0)
	asJSON := flag.Bool("json", false, "print the study as JSON instead of text")
	flag.Parse()
	ids := []string{"IWDA", "MSCIWORLD-USD"}
	if flag.NArg() == 2 {
		ids = flag.Args()
	}

	a, err := marketdata.Bundled(ids[0])
	if err != nil {
		log.Fatal(err)
	}
	b, err := marketdata.Bundled(ids[1])
	if err != nil {
		log.Fatal(err)
	}
	// A reference file states no currency; this one is in dollars, like
	// the fund, and saying so spares a currency warning.
	if b.Currency == "" {
		b.Currency = a.Currency
	}

	// Pair reads both on the window they share: the level (CAGR gap and its
	// standard error), a Monthly block (correlation, tracking error, beta,
	// the largest divergences, dated), a Daily one when both are daily, the
	// calendar years side by side, and Warnings.
	st, err := analyze.Pair(a, b, analyze.PairOptions{Divergences: 3})
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
