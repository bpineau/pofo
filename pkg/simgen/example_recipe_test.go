package simgen_test

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/simgen"
)

// offline is a Fetcher that has nothing: a recipe served by it reads the
// bundled reference series and nothing else.
type offline struct{}

func (offline) Fetch(id string, _ time.Time) (*marketdata.Series, error) {
	return nil, errors.New("offline: " + id)
}

// Find returns the bundled recipe that rebuilds an asset's past, and its Build
// runs on any Fetcher. The 25+ year STRIPS fund (ZROZ, US72201R8824) is priced
// off the long Treasury yield the repository bundles, so WithRefData serves
// all it needs; against live data the fallback is
// simgen.WithContext(ctx, marketdata.NewClient(dir)), which also fetches the
// real quotes to graft and validate against.
func ExampleFind() {
	r, ok := simgen.Find("ZROZ")
	if !ok {
		panic("no recipe")
	}
	s, err := r.Build(simgen.WithRefData(datasets.Refdata(), offline{}), time.Time{})
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Name)
	fmt.Printf("rebuilt from %s, graded against %s\n", s.First().Date.Format(time.DateOnly), r.ValidateAgainst)
	// Output:
	// PIMCO 25+Y zero-coupon: 27y Treasury STRIP
	// rebuilt from 1953-04-30, graded against ZROZ
}

// Validate grades a reconstruction on the dates it shares with the real
// series: correlation of daily and weekly returns, beta, tracking error and
// the two growth rates. Here the reconstruction under-reads every move by a
// tenth, adds a wobble of its own and drifts a little lower: the path is
// right, the level is not, and only the growth rates say so.
func ExampleValidate() {
	var dates []time.Time
	var real, sim []float64
	r, s := 100.0, 100.0
	for i := range 750 {
		move := 0.01 * math.Sin(float64(i)*0.7)
		r *= 1 + 0.0003 + move
		s *= 1 + 0.0002 + 0.9*move + 0.002*math.Cos(float64(i)*1.3)
		dates = append(dates, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
		real, sim = append(real, r), append(sim, s)
	}
	realS, _ := marketdata.NewSeries("FUND", dates, real)
	simS, _ := marketdata.NewSeries("FUND (rebuilt)", dates, sim)

	v, err := simgen.Validate(simS, realS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d common returns, correlation %.2f, beta %.2f\n", v.Overlap, v.Corr, v.Beta)
	fmt.Printf("CAGR %.1f %% rebuilt vs %.1f %% real\n", v.CAGRSim*100, v.CAGRReal*100)
	// Output:
	// 749 common returns, correlation 0.98, beta 1.06
	// CAGR 8.1 % rebuilt vs 12.1 % real
}
