//go:build ignore

// Optimize answers "what weights would the optimizer pick for these
// holdings?": the long-only weights of an objective over a portfolio file's
// lines, fitted on their monthly returns, next to the weights the file
// writes, with both allocations' in-sample return, volatility and deepest
// drawdown.
//
// Usage:
//
//	go run examples/code/optimize.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-objective SPEC] [-daily] FILE
//
// Example:
//
//	go run examples/code/optimize.go -objective min-volatility,max-weight:40 examples/portfolios/golden-butterfly.txt
//
// SPEC is what "#meta optimize:" takes (optimize.ParseSpec): an objective
// (max-sharpe, min-volatility, max-return, risk-parity, max-sortino,
// return-to-drawdown, min-ulcer, max-worst-5y, black-litterman) and its
// constraints (max-weight:40, bounds:ID:LO-HI, max-vol:9, view:ID:Q@C...).
// The default is the file's own directive, else max-sharpe; cwarp needs a
// benchmark and "train:" a hold-out, which pofo's report runs. Every figure
// is IN SAMPLE: the weights were fitted on the very months that score them.
// The holdings are read through marketdata.Client.Load: real quotes, with
// the bundled reconstruction in front under "#meta sim:on".
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/optimize"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "EUR", "the portfolio's currency, every holding converted into it")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	objective := flag.String("objective", "", "an optimize spec, as #meta optimize: takes (default: the file's, else max-sharpe)")
	daily := flag.Bool("daily", false, "fit on daily returns (default: monthly, robust to closes struck on different exchanges)")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: optimize [flags] FILE")
	}

	spec, err := portfolio.ParseFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	var sp optimize.Spec
	switch {
	case *objective != "":
		if sp, err = optimize.ParseSpec(*objective); err != nil {
			log.Fatal(err)
		}
	case spec.Optimize != nil:
		sp = *spec.Optimize
	default:
		sp = optimize.Spec{Objective: optimize.MaxSharpe}
	}
	if sp.Objective == optimize.CWARP || !sp.Train.IsZero() {
		log.Fatal("cwarp and train: need a benchmark or a hold-out: run pofo on the file instead")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}
	var list []*marketdata.Series
	written := make([]float64, len(spec.Holdings))
	ids := make([][]string, len(spec.Holdings)) // each line's spellings, for bounds and views
	for i, h := range spec.Holdings {
		s, err := client.Load(context.Background(), portfolio.SimFetchID(h.ID, spec.Sim), opt)
		if err != nil {
			log.Fatal(err)
		}
		cp := *s // label the column as written, on a copy (the client memoizes)
		cp.Symbol = h.ID
		list = append(list, &cp)
		written[i] = h.Weight // a FRACTION, normalized to sum to 1 (RawWeight is the file's percent)
		ids[i] = []string{h.ID, s.Symbol}
	}

	freq := marketdata.Monthly
	if *daily {
		freq = marketdata.Daily
	}
	p, err := marketdata.NewPanel(freq, list...)
	if err != nil {
		log.Fatal(err)
	}
	// Bounds and views arrive keyed by identifier; Resolve turns them into
	// positions. Black-Litterman's prior is the file's own allocation.
	if err := sp.Resolve(ids); err != nil {
		log.Fatal(err)
	}
	if sp.Objective == optimize.BlackLitterman {
		sp.Prior = written
	}
	ppy := p.PeriodsPerYear()
	res, err := optimize.Solve(p.R, ppy, sp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s: %s over %d periods (%.0f a year), %s to %s, in %s\n\n", spec.Name, sp.Objective, p.Len(), ppy,
		p.Starts[0].Format(time.DateOnly), p.Ends[p.Len()-1].Format(time.DateOnly), *currency)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "holding\twritten\toptimized\tCAGR\tvol\t")
	for i, id := range p.IDs {
		cagr, vol, _ := optimize.PathStats(p.R[i:i+1], []float64{1}, ppy)
		fmt.Fprintf(tw, "%s\t%.1f %%\t%.1f %%\t%.2f %%\t%.1f %%\t\n", id, written[i]*100, res.Weights[i]*100, cagr*100, vol*100)
	}
	tw.Flush()

	fmt.Fprintln(tw, "\nin sample\tCAGR\tvol\tmax DD\tSharpe\t")
	for _, row := range []struct {
		name string
		w    []float64
	}{{"written", written}, {"optimized", res.Weights}} {
		cagr, vol, dd := optimize.PathStats(p.R, row.w, ppy)
		mix := make([]float64, p.Len()) // the rebalanced blend's returns, for its Sharpe
		for t := range mix {
			for i, w := range row.w {
				mix[t] += w * p.R[i][t]
			}
		}
		fmt.Fprintf(tw, "%s\t%.2f %%\t%.1f %%\t%.1f %%\t%.2f\t\n", row.name, cagr*100, vol*100, dd*100, metrics.Sharpe(mix, 0, ppy))
	}
	tw.Flush()
	if sp.Limits.Any() && !res.Feasible {
		fmt.Println("\nno allocation meets the limits: the weights are the least-violating point found, not an answer")
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
