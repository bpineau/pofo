//go:build ignore

// Calendar answers "what did each year look like?": the calendar-year
// returns of several series side by side (or quarters, or months), and
// optionally the first series against a published table of yearly returns,
// the check a backcast or a reference series owes its source.
//
// Usage:
//
//	go run examples/code/calendar.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-months 12] [-published FILE] ID...
//
// Example:
//
//	go run examples/code/calendar.go -from 2000-01-01 IWDA SP500-USD XAUUSD-LBMA
//
// Each period runs from the previous period's last close to its own, so a
// year is measured from the last session of the year before. A period marked
// "*" is partial: the series starts or stops inside it. -published reads a
// "year,return" CSV, the return in PERCENT (a "#" line or a header is
// skipped), and sets it beside the first ID. Each ID goes through
// marketdata.Client.Load (a CSV path, else the bundle, else the network).
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert every series into this currency (default: each one's own)")
	from := flag.String("from", "", "list the periods from the one holding this date, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	months := flag.Int("months", 12, "period length in months: 12 yearly, 3 quarterly, 1 monthly")
	published := flag.String("published", "", `a "year,return" CSV (return in percent) set beside the first ID`)
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: calendar [flags] ID...")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	// The whole history is loaded and the periods filtered afterwards, so
	// the first period listed is measured from the close before it.
	start := date(*from)
	opt := marketdata.FetchOptions{To: date(*to), Currency: *currency}

	// cells[label][i] is series i's return over the period named label.
	cells := map[string][]string{}
	var ids []string
	var first map[string]float64 // the first series' returns, by label, for -published
	for i, id := range flag.Args() {
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		ids = append(ids, s.Symbol)
		cal, err := metrics.CalendarReturns(s.Dates(), s.Values(), *months)
		if err != nil {
			log.Fatal(err)
		}
		if i == 0 {
			first = map[string]float64{}
		}
		for k, c := range cal {
			if c.End.Before(start) {
				continue
			}
			l := label(c.End, *months)
			if cells[l] == nil {
				cells[l] = make([]string, flag.NArg())
			}
			// CalendarReturns flags a partial first period; a last one that
			// stops before its period's last weekday is partial too.
			partial := c.Partial || (k == len(cal)-1 && !complete(c.End, *months))
			mark := ""
			if partial {
				mark = "*"
			}
			cells[l][i] = fmt.Sprintf("%+.1f %%%s", c.Return*100, mark)
			if i == 0 && !partial {
				first[l] = c.Return
			}
		}
	}

	var pub map[string]float64
	if *published != "" {
		if *months != 12 {
			log.Fatal("-published compares calendar years: drop -months")
		}
		var err error
		if pub, err = readPublished(*published); err != nil {
			log.Fatal(err)
		}
	}

	labels := make([]string, 0, len(cells))
	for l := range cells {
		labels = append(labels, l)
	}
	slices.Sort(labels)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprint(tw, "period\t")
	for _, id := range ids {
		fmt.Fprintf(tw, "%s\t", id)
	}
	if pub != nil {
		fmt.Fprint(tw, "published\tgap\t")
	}
	fmt.Fprintln(tw)
	var gaps []float64
	for _, l := range labels {
		fmt.Fprintf(tw, "%s\t", l)
		for _, c := range cells[l] {
			fmt.Fprintf(tw, "%s\t", c)
		}
		if pub != nil {
			p, okP := pub[l]
			f, okF := first[l]
			switch {
			case okP && okF:
				gaps = append(gaps, f-p)
				fmt.Fprintf(tw, "%+.1f %%\t%+.2f pt\t", p*100, (f-p)*100)
			case okP:
				fmt.Fprintf(tw, "%+.1f %%\t\t", p*100)
			default:
				fmt.Fprint(tw, "\t\t")
			}
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
	if len(gaps) > 0 {
		abs := make([]float64, len(gaps))
		for i, g := range gaps {
			abs[i] = max(g, -g)
		}
		fmt.Printf("\n%s against the published table: %d years, mean gap %+.2f pt, largest %.2f pt\n",
			ids[0], len(gaps), metrics.Mean(gaps)*100, slices.Max(abs)*100)
	}
}

// label names the calendar period of months months that ends at end.
func label(end time.Time, months int) string {
	switch months {
	case 12:
		return strconv.Itoa(end.Year())
	case 1:
		return end.Format("2006-01")
	case 3:
		return fmt.Sprintf("%d-Q%d", end.Year(), (int(end.Month())-1)/3+1)
	case 6:
		return fmt.Sprintf("%d-H%d", end.Year(), (int(end.Month())-1)/6+1)
	default:
		return fmt.Sprintf("%d-P%d", end.Year(), (int(end.Month())-1)/months+1)
	}
}

// complete reports whether a period of months months whose last quote is
// end was quoted to its close: on or after the period's last weekday.
func complete(end time.Time, months int) bool {
	m := (int(end.Month())-1)/months*months + months // the period's last month
	lastDay := time.Date(end.Year(), time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC)
	for lastDay.Weekday() == time.Saturday || lastDay.Weekday() == time.Sunday {
		lastDay = lastDay.AddDate(0, 0, -1)
	}
	return !end.Before(lastDay)
}

// readPublished reads a "year,return" file, the return in percent, into
// fractions by year label.
func readPublished(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]float64{}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		y, r, ok := strings.Cut(line, ",")
		year, errY := strconv.Atoi(strings.TrimSpace(y))
		pct, errR := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(r), "%"), 64)
		if !ok || errY != nil || errR != nil {
			if n == 1 {
				continue // a header
			}
			return nil, fmt.Errorf("%s:%d: want year,return (percent), got %q", path, n, line)
		}
		out[strconv.Itoa(year)] = pct / 100
	}
	return out, sc.Err()
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
