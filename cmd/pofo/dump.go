// The -dump mode: series out as CSV, for another program to read.
package main

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// dumpOptions are the -dump settings that differ from the other modes'.
type dumpOptions struct {
	// currency converts every series into it; empty keeps each one native.
	// It is -currency only when the command line set it: a dump is data for
	// another program, and a silent conversion into the report default (EUR)
	// would hand it numbers in a currency it never asked for.
	currency string
	simAll   bool // -simulate: fetch every identifier with the SIM suffix
}

// runDump writes the series of ids to w in the long id,date,value layout
// (marketdata.WriteCSV, metadata as leading "#" comments) and returns. Each
// series is fetched the way every other mode fetches it (SIM convention,
// -no-simulate, -simulate, -currency, the quote cache, -offline), cut to
// -start and -end, stripped of any nowcast tail (an estimate is never
// stored, and a file is storage), resampled to month-ends under -monthly,
// and labelled with the identifier as the caller wrote it, so the id column
// joins back to the request. The single id "list" prints the bundled series
// instead (listBundled).
//
// Nothing but CSV reaches w: progress and warnings go through the client's
// log, to stderr.
func runDump(ctx context.Context, c *marketdata.Client, w io.Writer, ids []string, opt *options, d dumpOptions) error {
	if len(ids) == 1 && ids[0] == "list" {
		return listBundled(w)
	}
	list := make([]*marketdata.Series, 0, len(ids))
	for _, id := range ids {
		s, err := dumpSeries(ctx, c, id, opt, d)
		if err != nil {
			return err
		}
		s = marketdata.Trim(s, opt.start, opt.end)
		if s.Len() == 0 {
			return fmt.Errorf("%s: no point between -start and -end", id)
		}
		if opt.monthly {
			if s, err = s.Resample(marketdata.Monthly); err != nil {
				return err
			}
		}
		list = append(list, s)
	}
	return marketdata.WriteCSV(w, list...)
}

// dumpSeries returns the whole series of one identifier, relabelled with it.
//
// A reference series bundled in the binary (TREASURY-LONG-USD, TBILL-3M, the
// MSCI anchors) is no catalog asset and no quote source knows it, so it is
// read from the bundle rather than searched for online. It states no currency
// (see marketdata.Bundled), so asking to convert it is an error rather than a
// silent pass-through. Everything else goes through Client.FetchExtended.
func dumpSeries(ctx context.Context, c *marketdata.Client, id string, opt *options, d dumpOptions) (*marketdata.Series, error) {
	var s *marketdata.Series
	if ref, err := marketdata.Bundled(id); err == nil && ref.Source == "refdata" && !marketdata.KnownLocal(id) {
		if d.currency != "" {
			return nil, fmt.Errorf("%s: a bundled reference series states no currency to convert from; drop -currency", id)
		}
		s = ref
	} else {
		fetched, err := c.FetchExtended(ctx, portfolio.SimFetchID(id, d.simAll), marketdata.FetchOptions{
			From:     opt.start,
			To:       opt.end,
			NoSim:    opt.noSim,
			Simdata:  opt.simdata,
			Currency: d.currency,
		})
		if err != nil {
			return nil, err
		}
		s = fetched.WithoutEstimates()
	}
	out := *s // the client memoizes what it serves: relabel a copy
	out.Symbol = id
	return &out, nil
}

// listBundled prints every series the binary bundles, one per line: its
// identifier, its family (simdata, a catalog asset's backcast; refdata, a
// reference series), its first and last dates, its point count and its name.
func listBundled(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tKIND\tFIRST\tLAST\tPOINTS\tNAME")
	for _, id := range marketdata.BundledIDs() {
		s, err := marketdata.Bundled(id)
		if err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n", id, s.Source,
			s.First().Date.Format(time.DateOnly), s.Last().Date.Format(time.DateOnly), s.Len(), s.Name)
	}
	return tw.Flush()
}
