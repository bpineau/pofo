// Export writes bundled series to one CSV file any tool reads: month-end
// closes of a daily fund backcast, gold and the S&P 500, in the long
// "id,date,value" layout, each series' metadata as "#" comment lines.
//
//	go run ./examples/lib/export [-o out.csv] [ID...]
//
// marketdata.ReadLongCSV reads the file back, values exact; "pofo -dump"
// is the same on the command line.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func main() {
	log.SetFlags(0)
	out := flag.String("o", "", "file to write (default: standard output)")
	flag.Parse()
	ids := flag.Args()
	if len(ids) == 0 {
		ids = []string{"IWDA", "XAUUSD-LBMA", "SP500-USD"}
	}

	var series []*marketdata.Series
	for _, id := range ids {
		s, err := marketdata.Bundled(id)
		if err != nil {
			log.Fatal(err)
		}
		// Resample keeps the last trading close of every calendar month,
		// the cadence of the bundled monthly references.
		if s, err = s.Resample(marketdata.Monthly); err != nil {
			log.Fatal(err)
		}
		series = append(series, s)
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
	if err := marketdata.WriteCSV(w, series...); err != nil {
		log.Fatal(err)
	}
}
