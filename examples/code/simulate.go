//go:build ignore

// Simulate answers "how does this portfolio file behave?": the file built,
// rebalanced and replayed over the window all its holdings quote, with its
// statistics, its calendar years, its deepest drawdowns, and each holding's
// own statistics next to its share of the risk and of the return.
//
// Usage:
//
//	go run examples/code/simulate.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-rebalance 90] [-sim] [-benchmark ID] [-years 10] FILE
//
// Example:
//
//	go run examples/code/simulate.go examples/portfolios/golden-butterfly.txt
//
// The file is the format pofo reads (weights in percent, "#meta"
// directives; see examples/portfolios). The work is analyze.Portfolio's,
// which the pofo report's columns are built with too; the holdings are read
// through marketdata.Client.Load, so a catalog fund with a bundled backcast
// runs on its whole reconstructed history, offline, and a holding written
// with the SIM suffix (or the whole file under "#meta sim:on" or -sim) reads
// its live quotes with that backcast in front.
package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// loader is the analyze.Source that reads through Client.Load: the bundle
// before the network. Fees comes with the embedded client.
type loader struct{ *marketdata.Client }

func (l loader) FetchExtended(ctx context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error) {
	return l.Load(ctx, id, opt)
}

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "EUR", "the portfolio's currency, every holding converted into it")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	rebalance := flag.Int("rebalance", 0, "rebalancing period in days when the file sets none (0: the default, 90)")
	sim := flag.Bool("sim", false, "read every holding with its SIM suffix, as #meta sim:on does")
	bench := flag.String("benchmark", "", "an identifier to measure beta and alpha against")
	years := flag.Int("years", 10, "calendar years listed, most recent last (0: all)")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: simulate [flags] FILE")
	}

	spec, err := portfolio.ParseFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	st, err := analyze.Portfolio(context.Background(), loader{client}, spec, analyze.Options{
		Currency: *currency, Sim: *sim, Benchmark: *bench,
		From: date(*from), To: date(*to), Rebalance: *rebalance,
	})
	if err != nil {
		log.Fatal(err)
	}

	s := st.Stats
	fmt.Printf("%s, %s, %s to %s (%.1f years)\n", spec.Name, *currency,
		s.Start.Format(time.DateOnly), s.End.Format(time.DateOnly), s.Years)
	fmt.Printf("CAGR %.2f %%/yr   volatility %.1f %%/yr   Sharpe %.2f   Sortino %.2f   max drawdown %.1f %%   longest underwater %.1f years\n",
		s.CAGR*100, s.Volatility*100, s.Sharpe, s.Sortino, s.MaxDrawdown*100, float64(s.TTRDays)/365.25)
	if r := st.Relative; r != nil {
		fmt.Printf("against %s: beta %.2f, alpha %+.2f %%/yr, up capture %.0f %%, down capture %.0f %%\n",
			*bench, r.Beta, r.Alpha*100, r.UpCapture*100, r.DownCapture*100)
	}

	// Each holding on the portfolio's window, with its share of the risk
	// (Euler decomposition of the variance) and of the realized return,
	// both read on monthly contributions.
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "\nholding\tweight\tCAGR\tvol\tmax DD\trisk share\treturn share\t")
	for i, h := range st.Holdings {
		a := st.Portfolio.Assets[i]
		fmt.Fprintf(tw, "%s\t%.1f %%\t%.2f %%\t%.1f %%\t%.1f %%\t%.0f %%\t%.0f %%\t\n", a.ID, a.Weight*100,
			h.Stats.CAGR*100, h.Stats.Volatility*100, h.Stats.MaxDrawdown*100,
			st.Attribution.Risk[i]*100, st.Attribution.Return[i]*100)
	}
	tw.Flush()

	cal := st.Years
	if *years > 0 && len(cal) > *years {
		cal = cal[len(cal)-*years:]
	}
	fmt.Fprintln(tw, "\nyear\treturn\t")
	for _, y := range cal {
		fmt.Fprintf(tw, "%d\t%+.1f %%\t\n", y.End.Year(), y.Return*100)
	}
	tw.Flush()

	deepest := slices.Clone(st.Drawdowns) // chronological in the study
	slices.SortFunc(deepest, func(a, b metrics.Episode) int { return cmp.Compare(a.Depth, b.Depth) })
	fmt.Fprintln(tw, "\ndrawdown\tpeak\ttrough\trecovered\t")
	for _, e := range deepest[:min(3, len(deepest))] {
		back := "not yet"
		if !e.Ongoing {
			back = e.RecoverDate.Format("2006-01")
		}
		fmt.Fprintf(tw, "%.1f %%\t%s\t%s\t%s\t\n", e.Depth*100, e.PeakDate.Format("2006-01"), e.TroughDate.Format("2006-01"), back)
	}
	tw.Flush()

	for _, w := range st.Warnings {
		fmt.Println("warning:", w)
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
