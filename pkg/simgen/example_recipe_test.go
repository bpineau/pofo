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
// tenth (its beta on the real series, 0.90, says so), adds a wobble of its
// own and drifts a little lower: the path is right, the level is not, and
// only the growth rates say so.
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
	// 749 common returns, correlation 0.98, beta 0.90
	// CAGR 8.1 % rebuilt vs 12.1 % real
}

// held is a Fetcher over series already in memory.
type held map[string]*marketdata.Series

func (h held) Fetch(id string, _ time.Time) (*marketdata.Series, error) {
	if s, ok := h[id]; ok {
		return s, nil
	}
	return nil, errors.New("not held: " + id)
}

// VolManaged holds an asset at a multiplier of its capital, the target
// volatility over the asset's trailing realized one, the rest in cash, and
// refreshes it at every month end. Here a synthetic fund runs at about 9 %/yr
// of volatility through a rising year, then at three times that through a
// falling one: against a 12 % target, capped at 1 and floored at 0.5, the
// engine is fully invested in the calm year and at its floor in the storm,
// which then costs it less.
func ExampleVolManaged() {
	var dates []time.Time
	var fund, cash []float64
	f, c := 100.0, 100.0
	for d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC); d.Year() < 2022; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		drift, amp := 0.0003, 0.008
		if d.Year() == 2021 {
			drift, amp = -0.0008, 0.024
		}
		f *= 1 + drift + amp*math.Sin(1.7*float64(len(dates)))
		c *= 1 + 0.0001
		dates, fund, cash = append(dates, d), append(fund, f), append(cash, c)
	}
	fundS, _ := marketdata.NewSeries("FUND", dates, fund)
	cashS, _ := marketdata.NewSeries("CASH", dates, cash)

	fr, err := simgen.BuildFrame(held{"FUND": fundS, "CASH": cashS}, []string{"CASH", "FUND"}, time.Time{})
	if err != nil {
		panic(err)
	}
	p, err := simgen.VolManaged(fr, simgen.VolManagedConfig{
		AssetID: "FUND", CashID: "CASH",
		TargetVol: 0.12, Lookback: 63, Floor: 0.5, Cap: 1,
	})
	if err != nil {
		panic(err)
	}
	for _, month := range []string{"2020-09", "2021-09"} {
		for i, m := range p.Multiplier {
			if fr.Dates[p.Start+i].Format("2006-01") == month {
				fmt.Printf("%s: multiplier %.2f\n", month, m)
				break
			}
		}
	}
	fmt.Printf("from %s: managed %.1f, fund %.1f\n", fr.Dates[p.Start].Format(time.DateOnly),
		p.Values[len(p.Values)-1], 100*fund[len(fund)-1]/fund[p.Start])
	// Output:
	// 2020-09: multiplier 1.00
	// 2021-09: multiplier 0.50
	// from 2020-03-31: managed 93.6, fund 83.1
}
