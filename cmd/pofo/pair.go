// The -pair mode: one series against its reference, as text or JSON.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// pairOptions are the -pair settings beyond those it shares with -dump.
type pairOptions struct {
	dumpOptions
	json    bool // -json: the study as JSON rather than text
	leadLag bool // -lead-lag: analyze.PairOptions.LeadLag
}

// runPair compares two series, "A,B" in arg, with analyze.Pair and writes
// the study to w: plain text by default, JSON under -json. A is the
// candidate (a reconstruction, a refreshed file), B the reference (the real
// fund, the index, the file before the refresh).
//
// Each side is an identifier, fetched as -dump fetches it (SIM convention,
// -simulate, -currency when set, the quote cache, -offline, a bundled
// reference series read from the bundle), or the path of a "date,value" CSV
// file (pairSeries tells them apart), which is how an older version of a
// bundled file joins in: git show HEAD~1:pkg/datasets/refdata/X.csv >
// /tmp/old.csv, then -pair X,/tmp/old.csv. Both are cut to -start and -end.
func runPair(ctx context.Context, c *marketdata.Client, w io.Writer, arg string, opt *options, p pairOptions) error {
	ids := splitIDs(arg)
	if len(ids) != 2 {
		return fmt.Errorf("-pair takes two series, A,B (an identifier or a CSV path each): got %q", arg)
	}
	var list [2]*marketdata.Series
	for i, id := range ids {
		s, err := pairSeries(ctx, c, id, opt, p.dumpOptions)
		if err != nil {
			return err
		}
		if list[i] = marketdata.Trim(s, opt.start, opt.end); list[i].Len() == 0 {
			return fmt.Errorf("%s: no point between -start and -end", id)
		}
	}
	st, err := analyze.Pair(list[0], list[1], analyze.PairOptions{LeadLag: p.leadLag})
	if err != nil {
		return err
	}
	if p.json {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(st)
	}
	return st.WriteText(w)
}

// pairSeries reads one side of a pair. An argument holding a path separator
// or ending in ".csv" is a file, read with marketdata.ReadCSV and labelled
// with the path as written; converting it under -currency needs the
// currency its header states. Anything else is an identifier (dumpSeries).
func pairSeries(ctx context.Context, c *marketdata.Client, arg string, opt *options, d dumpOptions) (*marketdata.Series, error) {
	if !strings.ContainsRune(arg, '/') && !strings.ContainsRune(arg, os.PathSeparator) && !strings.HasSuffix(strings.ToLower(arg), ".csv") {
		return dumpSeries(ctx, c, arg, opt, d)
	}
	f, err := os.Open(arg)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := marketdata.ReadCSV(f, arg)
	if err != nil {
		return nil, err
	}
	if d.currency == "" {
		return s, nil
	}
	if s.Currency == "" {
		return nil, fmt.Errorf("%s: the file states no currency to convert from; drop -currency", arg)
	}
	out, _, err := c.ConvertCurrency(ctx, s, d.currency, opt.start)
	return out, err
}
