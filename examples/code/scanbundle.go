//go:build ignore

// Scanbundle answers "is the bundled data sound?": every series the module
// embeds, with its family, cadence, span and age, and what the bundle's own
// data guards find in it (marketdata.FindGaps: stretches of calendar it does
// not cover; marketdata.FindSpikes: one-session round trips no instrument
// could have made). The golden guards refuse both in any embedded file,
// except the measured and dated defects their allow-lists name (knownGaps and
// knownSpikes in pkg/datasets/golden); a finding outside those lists on a
// fresh checkout means a guard is not running.
//
// Usage:
//
//	go run examples/code/scanbundle.go [-v] [-stale 60] [-only PATTERN]
//
// Example:
//
//	go run examples/code/scanbundle.go -stale 45
//
// It reads the bundle only: no client, no network.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	verbose := flag.Bool("v", false, "list every gap and spike found, dated")
	stale := flag.Int("stale", 60, "flag a series whose last point is older than this many days")
	only := flag.String("only", "", "scan only the identifiers containing this text")
	flag.Parse()

	now := time.Now().UTC()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "id\tkind\tper yr\tfirst\tlast\tpoints\tgaps\tspikes\tflags")
	var details []string
	scanned, flagged := 0, 0
	for _, id := range marketdata.BundledIDs() {
		if *only != "" && !strings.Contains(id, strings.ToUpper(*only)) {
			continue
		}
		s, err := marketdata.Bundled(id)
		if err != nil {
			log.Fatal(err)
		}
		scanned++
		gaps, spikes := marketdata.FindGaps(s), marketdata.FindSpikes(s)
		var flags []string
		if len(gaps) > 0 {
			flags = append(flags, "gaps")
		}
		if len(spikes) > 0 {
			flags = append(flags, "spikes")
		}
		if age := now.Sub(s.Last().Date).Hours() / 24; age > float64(*stale) {
			flags = append(flags, fmt.Sprintf("stale %.0f d", age))
		}
		if len(s.Junctions) > 0 {
			flags = append(flags, fmt.Sprintf("%d junction(s)", len(s.Junctions)))
		}
		if len(flags) > 0 {
			flagged++
		}
		// PeriodsPerYear measures the prevailing cadence, snapped to 252,
		// 52, 12... A backcast that turns daily after a monthly deep past
		// reads as daily.
		fmt.Fprintf(tw, "%s\t%s\t%.0f\t%s\t%s\t%d\t%d\t%d\t%s\n", id, s.Source,
			metrics.PeriodsPerYear(s.Dates()), s.First().Date.Format(time.DateOnly),
			s.Last().Date.Format(time.DateOnly), s.Len(), len(gaps), len(spikes), strings.Join(flags, ", "))
		for _, g := range gaps {
			details = append(details, fmt.Sprintf("%s: no quote from %s to %s (%.0f days, the pace allows %.0f)",
				id, g.From.Format(time.DateOnly), g.To.Format(time.DateOnly), g.Days, g.Limit))
		}
		for _, sp := range spikes {
			details = append(details, fmt.Sprintf("%s: round trip on %s, %+.1f %% then %+.1f %% (local sigma %.2f %%)",
				id, sp.Date.Format(time.DateOnly), sp.In*100, sp.Out*100, sp.Sigma*100))
		}
	}
	tw.Flush()
	fmt.Printf("\n%d series scanned, %d flagged\n", scanned, flagged)
	if *verbose {
		for _, d := range details {
			fmt.Println(d)
		}
	}
}
