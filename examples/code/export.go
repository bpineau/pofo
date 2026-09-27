//go:build ignore

// Export answers "how do I get this data into another tool?": series written
// as one CSV any tool reads, in the long "id,date,value" layout with each
// series' metadata as "#" comment lines, daily or at month-ends.
//
// Usage:
//
//	go run examples/code/export.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-monthly] [-o FILE] ID...
//
// Example:
//
//	go run examples/code/export.go -monthly -o /tmp/series.csv IWDA XAUUSD-LBMA SP500-USD
//
// marketdata.ReadLongCSV reads the file back, values exact; "pofo -dump" is
// the same on the command line (its "list" names every bundled series). Each
// ID goes through marketdata.Client.Load (a CSV path, else the bundle, else
// the network) and is written under the identifier as given, so the id
// column joins back to the request.
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert every series into this currency (default: each one's own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	monthly := flag.Bool("monthly", false, "keep the last close of every calendar month")
	out := flag.String("o", "", "file to write (default: standard output)")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: export [flags] ID...")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf // on stderr: nothing but CSV reaches standard output
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}
	var list []*marketdata.Series
	for _, id := range flag.Args() {
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		// Resample keeps the last trading close of every calendar month,
		// the cadence of the bundled monthly references.
		if *monthly {
			if s, err = s.Resample(marketdata.Monthly); err != nil {
				log.Fatal(err)
			}
		}
		cp := *s // relabel a copy: the client memoizes what it serves
		cp.Symbol = id
		list = append(list, &cp)
	}

	var w io.Writer = os.Stdout
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w = f
	}
	if err := marketdata.WriteCSV(w, list...); err != nil {
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
