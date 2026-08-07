package web

import (
	"strconv"
	"strings"
	"sync"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/scenario"
)

// countrySeries is one JST country's annual real equity returns, year-indexed
// so a specific historical start date (a "vintage") can be replayed.
type countrySeries struct {
	ISO       string
	FirstYear int
	Returns   []float64
}

// broadSampleCountries is the bundled Jorda-Schularick-Taylor per-country real
// equity history (annual, 1870-2020), parsed once into one series per country.
// It backs the empirical "Broad-sample" model: a pool of actual single-market
// records rather than a synthetic prior or a pre-diversified world index, so
// the real bear decades that cause ruin (France/Portugal early-century, Japan
// post-1990) can land inside a retirement at full force.
var broadSampleCountries = sync.OnceValue(func() []countrySeries {
	return parseBroadSample(datasets.BroadSample())
})

// broadSampleEquity is the same history as bare return series, the shape the
// pooled bootstrap consumes.
func broadSampleEquity() [][]float64 {
	cs := broadSampleCountries()
	out := make([][]float64, len(cs))
	for i, c := range cs {
		out[i] = c.Returns
	}
	return out
}

// broadSampleSource pool-bootstraps annual real-return paths of the given length
// from the per-country equity histories, preserving each market's internal
// ordering (so a bad decade stays a bad decade) while mixing markets between
// blocks. The horizon is annual, matching the annual data, so no compounding
// wrapper is needed.
func broadSampleSource(years int) scenario.Source {
	return scenario.PooledBootstrap{
		Series:    broadSampleEquity(),
		MeanBlock: 10,
		Periods:   years,
	}
}

// parseBroadSample reads the iso,year,equity,bond,bill CSV (comment lines start
// with '#') into one year-indexed real-equity-return series per country,
// ordered as the file is (ascending year within each country).
func parseBroadSample(csv []byte) []countrySeries {
	byISO := map[string]*countrySeries{}
	var order []string
	for line := range strings.SplitSeq(string(csv), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 3 || f[2] == "" {
			continue
		}
		iso := f[0]
		c, seen := byISO[iso]
		if !seen {
			c = &countrySeries{ISO: iso, FirstYear: int(atof(f[1]))}
			byISO[iso] = c
			order = append(order, iso)
		}
		c.Returns = append(c.Returns, atof(f[2]))
	}
	out := make([]countrySeries, 0, len(order))
	for _, iso := range order {
		out = append(out, *byISO[iso])
	}
	return out
}

func atof(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
