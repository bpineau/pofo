package web

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/scenario"
)

// countrySeries is one JST country's annual real returns, year-indexed so a
// specific historical start date (a "vintage") can be replayed. Returns
// (equity) is contiguous from FirstYear; Bonds carries the same years with
// math.NaN where the bond record is missing (war breaks, defaults).
type countrySeries struct {
	ISO       string
	FirstYear int
	Returns   []float64
	Bonds     []float64
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

// broadSampleEquity is the bare per-country equity series. The vintage
// replays (section 02) consume it: the named worst cases are equity
// disasters, explicitly labelled as a growth-sleeve proxy. The Broad-sample
// model itself resamples broadSampleMixed, not this.
func broadSampleEquity() [][]float64 {
	cs := broadSampleCountries()
	out := make([][]float64, len(cs))
	for i, c := range cs {
		out[i] = c.Returns
	}
	return out
}

// broadSampleMixed is the pool the Broad-sample model resamples: per-country
// annual real returns of a 60/40 domestic stock/bond portfolio (rebalanced
// yearly), the baseline allocation of the broad-sample SWR literature
// (Anarkulova, Cederburg & O'Doherty), whose ~2.26% safe rate this model is
// anchored against. Pure single-market equity would stress the ALLOCATION on
// top of the data (22% vol vs a diversified retiree's 10-15%), overstating
// ruin for reasons the column does not claim. Each country contributes its
// contiguous runs of years where both records exist; the war breaks in the
// bond record (Germany 1944-48, Japan 1946-47, Spain 1937-40, WWI) split a
// country into separate runs rather than splicing across the gap. The bond
// gaps do drop a few catastrophic equity years, a known survivorship caveat
// of the JST bond record; 1929, the inflationary 1970s and Japan 1990 all
// remain at full force.
var broadSampleMixed = sync.OnceValue(func() [][]float64 {
	const equityW = 0.60
	var pool [][]float64
	for _, c := range broadSampleCountries() {
		var run []float64
		flush := func() {
			if len(run) >= 10 { // shorter stubs hold no useful block
				pool = append(pool, run)
			}
			run = nil
		}
		for i, e := range c.Returns {
			b := c.Bonds[i]
			if math.IsNaN(b) {
				flush()
				continue
			}
			run = append(run, equityW*e+(1-equityW)*b)
		}
		flush()
	}
	return pool
})

// broadSampleSource pool-bootstraps annual real-return paths of the given
// length from the per-country 60/40 histories, preserving each market's
// internal ordering (so a bad decade stays a bad decade) while mixing markets
// between blocks. The horizon is annual, matching the annual data, so no
// compounding wrapper is needed.
func broadSampleSource(years int) scenario.Source {
	return scenario.PooledBootstrap{
		Series:    broadSampleMixed(),
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
		if len(f) < 4 || f[2] == "" {
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
		bond := math.NaN()
		if f[3] != "" {
			bond = atof(f[3])
		}
		c.Bonds = append(c.Bonds, bond)
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
