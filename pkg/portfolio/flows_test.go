package portfolio

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// flatSeries builds a daily series at a constant price over n days.
func flatSeries(symbol string, start time.Time, n int, price float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol, Name: symbol}
	for i := 0; i < n; i++ {
		s.Points = append(s.Points, marketdata.Point{
			Date:  start.AddDate(0, 0, i),
			Close: price,
		})
	}
	return s
}

func TestSimulateMonthlyContributions(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &Portfolio{
		Name:       "flows",
		Assets:     []Asset{{ID: "X", Symbol: "X", Weight: 1, Fees: -1, Series: flatSeries("X", start, 95, 50)}},
		Capital:    1000,
		Contribute: Flow{Amount: 100, Period: Monthly},
	}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	// 95 days from Jan 1st: flows fire on the first day of Feb, Mar, Apr.
	if sim.Contributed != 300 {
		t.Fatalf("Contributed = %v, want 300", sim.Contributed)
	}
	if got := sim.Values[len(sim.Values)-1]; math.Abs(got-1300) > 1e-9 {
		t.Fatalf("final value = %v, want 1300 (flat prices)", got)
	}
	// Flat prices: the time-weighted index must not move on contributions.
	for i, v := range sim.Index {
		if math.Abs(v-100) > 1e-9 {
			t.Fatalf("Index[%d] = %v, want 100 (flat prices)", i, v)
		}
	}
	if len(sim.FlowDates) != 3 || sim.FlowAmounts[0] != 100 {
		t.Fatalf("FlowDates/FlowAmounts = %v %v, want 3 contributions of 100", sim.FlowDates, sim.FlowAmounts)
	}
}

func TestSimulatePercentYearlyWithdrawal(t *testing.T) {
	start := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)
	p := &Portfolio{
		Name:     "swr",
		Assets:   []Asset{{ID: "X", Symbol: "X", Weight: 1, Fees: -1, Series: flatSeries("X", start, 400, 20)}},
		Capital:  10000,
		Withdraw: Flow{Amount: 4, Percent: true, Period: Yearly},
	}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	// One year boundary (Jan 1st 2021): 4 % of 10 000.
	if math.Abs(sim.Withdrawn-400) > 1e-9 {
		t.Fatalf("Withdrawn = %v, want 400", sim.Withdrawn)
	}
	if got := sim.Values[len(sim.Values)-1]; math.Abs(got-9600) > 1e-9 {
		t.Fatalf("final value = %v, want 9600", got)
	}
}

func TestSimulateWithdrawalRuin(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &Portfolio{
		Name:     "ruin",
		Assets:   []Asset{{ID: "X", Symbol: "X", Weight: 1, Fees: -1, Series: flatSeries("X", start, 200, 10)}},
		Capital:  1000,
		Withdraw: Flow{Amount: 600, Period: Monthly},
	}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !sim.Ruined {
		t.Fatal("expected ruin: 600/month out of 1000 flat")
	}
	if got := sim.Values[len(sim.Values)-1]; got != 0 {
		t.Fatalf("final value = %v, want 0 after ruin", got)
	}
}

func TestParseFlowsMeta(t *testing.T) {
	pf, err := Parse("test", strings.NewReader(`
#meta capital:10000
#meta contribute:500/month
#meta withdraw:4%/year
100 X
`))
	if err != nil {
		t.Fatal(err)
	}
	if pf.Capital != 10000 {
		t.Fatalf("Capital = %v", pf.Capital)
	}
	if pf.Contribute != (Flow{Amount: 500, Period: Monthly}) {
		t.Fatalf("Contribute = %+v", pf.Contribute)
	}
	if pf.Withdraw != (Flow{Amount: 4, Percent: true, Period: Yearly}) {
		t.Fatalf("Withdraw = %+v", pf.Withdraw)
	}

	if _, err := Parse("test", strings.NewReader("#meta contribute:500/month\n100 X\n")); err == nil {
		t.Fatal("contribute without capital must be rejected")
	}
	if _, err := Parse("test", strings.NewReader("#meta contribute:5%/month\n#meta capital:1000\n100 X\n")); err == nil {
		t.Fatal("percentage contribution must be rejected")
	}
	if _, err := Parse("test", strings.NewReader("#meta withdraw:4%/decade\n#meta capital:1000\n100 X\n")); err == nil {
		t.Fatal("unknown period must be rejected")
	}
}
