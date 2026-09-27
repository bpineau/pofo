// Simulate reads a portfolio file, builds it on bundled histories and
// replays it over time: portfolio.Parse, Build and Simulate, then
// metrics.Compute on the result.
//
//	go run ./examples/lib/simulate [-file portfolio.txt]
//
// Without -file it reads the file below. Every holding is fetched with
// marketdata.Bundled, so only bundled identifiers resolve; a live run
// passes client.FetchExtended to Build instead (see Build's godoc).
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// defaultFile is a portfolio in the file format pofo reads: one line per
// holding, the weight in PERCENT, then the identifier; "#meta" lines are
// directives.
const defaultFile = `# World equities, Treasuries and gold, rebalanced every year
60 IWDA   # iShares Core MSCI World, IE00B4L5Y983
30 IEF    # iShares 7-10 Year Treasury Bond ETF, US4642874402
10 IGLN   # iShares Physical Gold, IE00B4ND3602
#meta rebalance:365
`

func main() {
	log.SetFlags(0)
	file := flag.String("file", "", "portfolio file to read (default: the one in the source)")
	flag.Parse()

	var r io.Reader = strings.NewReader(defaultFile)
	name := "world-ust-gold"
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r, name = f, *file
	}
	spec, err := portfolio.Parse(name, r)
	if err != nil {
		log.Fatal(err)
	}

	// Build resolves every holding through the fetch callback. The bundled
	// histories are all in dollars here; mixing currencies calls for a
	// client that converts (FetchOptions.Currency).
	p, err := portfolio.Build(spec, portfolio.BuildOptions{
		Fetch: func(id string) (*marketdata.Series, error) { return marketdata.Bundled(id) },
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, w := range p.Warnings {
		fmt.Println("warning:", w)
	}

	rebalance := spec.RebalanceDays
	if rebalance < 0 { // the file did not say
		rebalance = 90
	}
	sim, err := portfolio.Simulate(p, rebalance)
	if err != nil {
		log.Fatal(err)
	}

	// Index is the time-weighted series from 100: the one statistics read.
	st, err := metrics.Compute(sim.Dates, sim.Index)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s, %s to %s, rebalanced every %d days\n", spec.Name,
		st.Start.Format(time.DateOnly), st.End.Format(time.DateOnly), rebalance)
	for _, a := range p.Assets {
		fmt.Printf("  %4.0f %%  %-6s %s\n", a.Weight*100, a.ID, a.Name)
	}
	fmt.Printf("CAGR %.2f %%/yr, volatility %.1f %%/yr, Sharpe %.2f, max drawdown %.1f %%\n",
		st.CAGR*100, st.Volatility*100, st.Sharpe, st.MaxDrawdown*100)

	// Each holding's share of the risk (Euler decomposition of variance)
	// and of the realized return, from the simulation's per-day
	// contributions, folded into months first.
	_, monthly := sim.MonthlyContributions()
	att, err := metrics.Attribute(monthly)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n%-6s %8s %8s %8s\n", "", "capital", "risk", "return")
	for i, a := range p.Assets {
		fmt.Printf("%-6s %7.0f%% %7.0f%% %7.0f%%\n", a.ID, a.Weight*100, att.Risk[i]*100, att.Return[i]*100)
	}
}
