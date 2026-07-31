package portfolio

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func day(i int) time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
}

// constSeries builds a daily series of n points starting at startDay with a
// constant price.
func constSeries(symbol string, startDay, n int, price float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(startDay + i), Close: price})
	}
	return s
}

func TestSimulateStaggeredStartAndConstantPrices(t *testing.T) {
	p := &Portfolio{
		Name: "t",
		Assets: []Asset{
			{Symbol: "A", Weight: 0.5, Series: constSeries("A", 0, 10, 50)},
			{Symbol: "B", Weight: 0.5, Series: constSeries("B", 2, 6, 200)},
		},
	}
	sim, err := Simulate(p, 90)
	if err != nil {
		t.Fatal(err)
	}
	if !sim.Dates[0].Equal(day(2)) {
		t.Errorf("expected start on day 2, got %v", sim.Dates[0])
	}
	if last := sim.Dates[len(sim.Dates)-1]; !last.Equal(day(7)) {
		t.Errorf("expected end on day 7, got %v", last)
	}
	for i, v := range sim.Values {
		if math.Abs(v-100) > 1e-9 {
			t.Errorf("constant prices: value[%d] = %v, want 100", i, v)
		}
	}
}

func TestSimulateRebalancingTrimsWinner(t *testing.T) {
	// A is flat; B compounds 1 % per day. Rebalancing sells part of B back
	// into A every 90 days, so the rebalanced portfolio must end lower than
	// the buy-and-hold one (and match its closed form).
	n := 200
	a := constSeries("A", 0, n, 100)
	b := &marketdata.Series{Symbol: "B"}
	price := 1.0
	for i := range n {
		b.Points = append(b.Points, marketdata.Point{Date: day(i), Close: price})
		price *= 1.01
	}
	mk := func() *Portfolio {
		return &Portfolio{Name: "t", Assets: []Asset{
			{Symbol: "A", Weight: 0.5, Series: a},
			{Symbol: "B", Weight: 0.5, Series: b},
		}}
	}
	hold, err := Simulate(mk(), 0)
	if err != nil {
		t.Fatal(err)
	}
	reb, err := Simulate(mk(), 90)
	if err != nil {
		t.Fatal(err)
	}
	endHold := hold.Values[len(hold.Values)-1]
	endReb := reb.Values[len(reb.Values)-1]
	wantHold := 50 + 50*math.Pow(1.01, float64(n-1))
	if math.Abs(endHold-wantHold) > 1e-6 {
		t.Errorf("buy-and-hold: %v, want %v", endHold, wantHold)
	}
	if endReb >= endHold {
		t.Errorf("rebalancing should hold B back: rebalanced %v >= buy-and-hold %v", endReb, endHold)
	}
	if reb.Values[50] != hold.Values[50] {
		t.Errorf("before the first rebalancing both must coincide")
	}
}

func TestSimulateNoOverlap(t *testing.T) {
	p := &Portfolio{Name: "t", Assets: []Asset{
		{Symbol: "A", Weight: 0.5, Series: constSeries("A", 0, 5, 10)},
		{Symbol: "B", Weight: 0.5, Series: constSeries("B", 10, 5, 10)},
	}}
	if _, err := Simulate(p, 90); err == nil {
		t.Fatal("expected error without a common period")
	}
}

// rampSeries builds a daily series whose price grows by step each day.
func rampSeries(symbol string, startDay, n int, base, step float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(startDay + i), Close: base + step*float64(i)})
	}
	return s
}

func TestSimulateContributions(t *testing.T) {
	p := &Portfolio{Name: "t", Assets: []Asset{
		{Symbol: "A", Weight: 0.6, Series: rampSeries("A", 0, 30, 100, 1)},
		{Symbol: "B", Weight: 0.4, Series: constSeries("B", 0, 30, 200)},
	}}
	sim, err := Simulate(p, 7) // several rebalances inside the window
	if err != nil {
		t.Fatal(err)
	}
	if len(sim.Contributions) != 2 {
		t.Fatalf("want one contribution series per asset, got %d", len(sim.Contributions))
	}
	for i, c := range sim.Contributions {
		if len(c) != len(sim.Dates) {
			t.Fatalf("asset %d: %d contributions for %d dates", i, len(c), len(sim.Dates))
		}
		if c[0] != 0 {
			t.Fatalf("asset %d: day-0 contribution = %v, want 0", i, c[0])
		}
	}
	// A flat asset contributes nothing; the sum reproduces the index return.
	for k := 1; k < len(sim.Dates); k++ {
		if sim.Contributions[1][k] != 0 {
			t.Fatalf("day %d: flat asset contributed %v", k, sim.Contributions[1][k])
		}
		sum := sim.Contributions[0][k] + sim.Contributions[1][k]
		r := sim.Index[k]/sim.Index[k-1] - 1
		if math.Abs(sum-r) > 1e-12 {
			t.Fatalf("day %d: contributions sum %v != index return %v", k, sum, r)
		}
	}
}

func TestSimulateContributionsEnvelopeFees(t *testing.T) {
	p := &Portfolio{Name: "t", EnvelopeFees: 2.52, Assets: []Asset{
		{Symbol: "A", Weight: 1, Series: rampSeries("A", 0, 20, 100, 0.5)},
	}}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	// The fee drag stays out of the attribution: the index return is the
	// contribution sum shaved by the daily fee, (1+r) = (1-f)(1+sum).
	f := 2.52 / 100 / 252
	for k := 1; k < len(sim.Dates); k++ {
		r := sim.Index[k]/sim.Index[k-1] - 1
		want := (1+r)/(1-f) - 1
		if math.Abs(sim.Contributions[0][k]-want) > 1e-12 {
			t.Fatalf("day %d: contribution %v, want %v", k, sim.Contributions[0][k], want)
		}
	}
}
