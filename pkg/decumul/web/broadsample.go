package web

import (
	"strconv"
	"strings"
	"sync"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/scenario"
)

// broadSampleEquity is the bundled Jorda-Schularick-Taylor per-country real
// equity history (annual, 1870-2020), parsed once into one series per country.
// It backs the empirical "Broad-sample" model: a pool of actual single-market
// records rather than a synthetic prior or a pre-diversified world index, so
// the real bear decades that cause ruin (France/Portugal early-century, Japan
// post-1990) can land inside a retirement at full force.
var broadSampleEquity = sync.OnceValue(func() [][]float64 {
	return parseBroadSample(datasets.BroadSample())
})

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
// with '#') into one real-equity-return series per country, ordered as the file
// is (ascending year within each country).
func parseBroadSample(csv []byte) [][]float64 {
	byISO := map[string][]float64{}
	var order []string
	for _, line := range strings.Split(string(csv), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 3 || f[2] == "" {
			continue
		}
		iso := f[0]
		if _, seen := byISO[iso]; !seen {
			order = append(order, iso)
		}
		byISO[iso] = append(byISO[iso], atof(f[2]))
	}
	out := make([][]float64, 0, len(order))
	for _, iso := range order {
		out = append(out, byISO[iso])
	}
	return out
}

func atof(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
