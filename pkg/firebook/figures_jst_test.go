package firebook

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The Jorda-Schularick-Taylor extraction the Trinity plate and the
// broad-sample plate share: annual REAL total returns, per country, from
// pkg/datasets/broadsample. Both plates recompute their own numbers on it
// rather than quoting anyone's published cells, so this file is where the
// convention lives, once.
//
// The mixed pool below mirrors pkg/decumul/web's broadSampleMixed (60/40
// domestic, contiguous runs only, war breaks splitting a country rather than
// splicing across it): that package imports firebook to serve the book, so
// importing it back would cycle, and the comments name the counterpart.

// jstSeries is one country's contiguous annual real record.
type jstSeries struct {
	iso    string
	first  int
	equity []float64
	bond   []float64 // NaN where the record is missing
}

// jstParse reads the bundled iso,year,equity,bond,bill table.
func jstParse(t *testing.T) []jstSeries {
	t.Helper()
	frozenAgainstData(t)
	byISO := map[string]*jstSeries{}
	var order []string
	for _, line := range strings.Split(string(datasets.BroadSample()), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 4 || f[2] == "" {
			continue
		}
		c, seen := byISO[f[0]]
		if !seen {
			year, _ := strconv.Atoi(strings.TrimSpace(f[1]))
			c = &jstSeries{iso: f[0], first: year}
			byISO[f[0]] = c
			order = append(order, f[0])
		}
		eq, _ := strconv.ParseFloat(strings.TrimSpace(f[2]), 64)
		bond := math.NaN()
		if f[3] != "" {
			bond, _ = strconv.ParseFloat(strings.TrimSpace(f[3]), 64)
		}
		c.equity = append(c.equity, eq)
		c.bond = append(c.bond, bond)
	}
	out := make([]jstSeries, 0, len(order))
	for _, iso := range order {
		out = append(out, *byISO[iso])
	}
	return out
}

// jstUSA is the United States record, trimmed to the contiguous years where
// both the equity and the bond series exist: the sample both plates measure
// the American case on.
func jstUSA(t *testing.T) jstSeries {
	t.Helper()
	for _, c := range jstParse(t) {
		if c.iso != "USA" {
			continue
		}
		var out jstSeries
		out.iso = c.iso
		for i, b := range c.bond {
			if math.IsNaN(b) {
				out.equity, out.bond = nil, nil
				continue
			}
			if len(out.equity) == 0 {
				out.first = c.first + i
			}
			out.equity = append(out.equity, c.equity[i])
			out.bond = append(out.bond, b)
		}
		return out
	}
	t.Fatal("the bundled panel carries no USA record")
	return jstSeries{}
}

// jstSuccess replays every rolling window of the record and reports the share
// that finished solvent, plus how many windows there were.
//
// The withdrawal convention is the one the retirement literature runs on and
// the one this book states everywhere: a fixed real amount, a share of the
// STARTING capital, taken at the beginning of each year and held constant in
// purchasing power (the returns are already real), with the two legs
// rebalanced to their weights every year. A window fails the moment the
// capital cannot pay a withdrawal in full.
func jstSuccess(s jstSeries, equity, rate float64, years int) (float64, int) {
	windows, ok := 0, 0
	for start := 0; start+years <= len(s.equity); start++ {
		windows++
		w := 1.0
		alive := true
		for k := 0; k < years; k++ {
			w -= rate
			if w <= 0 {
				alive = false
				break
			}
			w *= 1 + equity*s.equity[start+k] + (1-equity)*s.bond[start+k]
		}
		if alive && w > 0 {
			ok++
		}
	}
	if windows == 0 {
		return 0, 0
	}
	return float64(ok) / float64(windows), windows
}

// broadPool builds the book's broad-sample source: the pooled bootstrap over
// every country's contiguous 60/40 runs, exactly as the FIRE page's
// Broad-sample column does (web.broadSampleMixed + web.broadSampleSource, ten
// year mean blocks), so the plate reads the same model the book quotes.
func broadPool(t *testing.T, years int) scenario.Source {
	t.Helper()
	const equityW = 0.60
	var pool [][]float64
	for _, c := range jstParse(t) {
		var run []float64
		flush := func() {
			if len(run) >= 10 {
				pool = append(pool, run)
			}
			run = nil
		}
		for i, e := range c.equity {
			if math.IsNaN(c.bond[i]) {
				flush()
				continue
			}
			run = append(run, equityW*e+(1-equityW)*c.bond[i])
		}
		flush()
	}
	return scenario.PooledBootstrap{Series: pool, MeanBlock: 10, Periods: years}
}
