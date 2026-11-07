package portfolio

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestNewSpecWithLeverageBeforeNormalization(t *testing.T) {
	lines := []Line{{ID: "VTI", Weight: .9}, {ID: "BND", Weight: .6}}
	ordinary, err := NewSpec("p", lines...)
	if err != nil {
		t.Fatal(err)
	}
	levered, err := NewSpecWith("p", SpecOptions{Leverage: true}, lines...)
	if err != nil {
		t.Fatal(err)
	}
	if ordinary.Holdings[0].Weight != .6 || ordinary.Holdings[1].Weight != .4 {
		t.Fatal("ordinary normalization changed")
	}
	if levered.Holdings[0].Weight != .9 || levered.Holdings[1].Weight != .6 || !levered.Leverage {
		t.Fatal("leverage applied after normalization")
	}
	parsed, err := Parse("p", strings.NewReader("#meta leverage:on\n90 VTI\n60 BND\n"))
	if err != nil {
		t.Fatal(err)
	}
	parsed.Meta = nil // typed options do not invent textual directives
	if !reflect.DeepEqual(levered, parsed) {
		t.Fatalf("typed/file specs differ: %+v, %+v", levered, parsed)
	}
	days, fees, spread := 0, .5, 1.0
	configured, err := NewSpecWith("p", SpecOptions{Leverage: true, RebalanceDays: &days, EnvelopeFees: &fees, BorrowSpread: &spread}, lines...)
	if err != nil {
		t.Fatal(err)
	}
	if configured.RebalanceDays != 0 || configured.EnvelopeFees != fees || configured.BorrowSpread != spread {
		t.Fatal("typed options lost")
	}
	if _, err := NewSpecWith("p", SpecOptions{Leverage: true}, Line{ID: "VTI", Weight: 2}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), -.1, 21} {
		v := bad
		if _, err := NewSpecWith("p", SpecOptions{EnvelopeFees: &v}, lines...); err == nil {
			t.Errorf("accepted fees %g", bad)
		}
	}
}

func optionPortfolio(dates []time.Time, first, second []float64) *Portfolio {
	a, _ := marketdata.NewSeries("A", dates, first)
	b, _ := marketdata.NewSeries("B", dates, second)
	return &Portfolio{Capital: 1000, Assets: []Asset{{Symbol: "A", Weight: .5, Series: a}, {Symbol: "B", Weight: .5, Series: b}}}
}

func TestSimulateWithCostFunding(t *testing.T) {
	dates := []time.Time{day(0), day(1), day(2)}
	p := optionPortfolio(dates, []float64{100, 200, 200}, []float64{100, 100, 100})
	got, err := SimulateWith(p, SimOptions{RebalanceDays: 1, TransactionCost: .01})
	if err != nil {
		t.Fatal(err)
	}
	// Before rebalancing holdings are 1000 and 500. Post-cost 50/50
	// targets are 747.5 each: gross trades 252.5+247.5=500, cost 5.
	if math.Abs(got.Values[1]-1495) > 1e-9 || math.Abs(got.Costs[1]-5) > 1e-9 || math.Abs(got.Turnover[1]-1.0/3) > 1e-12 {
		t.Fatalf("wrong funded rebalance: %+v", got)
	}
	if math.Abs(got.Index[1]-149.5) > 1e-10 || got.Costs[0] != 0 || got.Turnover[0] != 0 {
		t.Fatal("entry cost convention or index wrong")
	}
	if math.Abs(got.Contributions[0][1]-.5) > 1e-12 || got.Contributions[1][1] != 0 {
		t.Fatal("cost attributed to an asset")
	}
	if math.Abs(got.Values[2]-1495) > 1e-9 || math.Abs(got.Costs[2]) > 1e-9 {
		t.Fatal("flat holdings traded again")
	}
}

func TestSimulateWithCostsAndFlows(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)}
	for _, withdraw := range []bool{false, true} {
		p := optionPortfolio(dates, []float64{100, 100, 100}, []float64{100, 100, 100})
		if withdraw {
			p.Withdraw = Flow{Amount: 100, Period: Monthly}
		} else {
			p.Contribute = Flow{Amount: 100, Period: Monthly}
		}
		got, err := SimulateWith(p, SimOptions{TransactionCost: .01})
		if err != nil {
			t.Fatal(err)
		}
		// A contribution buys 100/1.01 notional, costing 100/101.
		// A withdrawal sells 100/.99 notional, costing 100/99.
		notional := 100 / 1.01
		netFlow := 100.0
		denominator := 1100.0
		if withdraw {
			notional = 100 / .99
			netFlow = -100
			denominator = 1000
		}
		cost := .01 * notional
		if math.Abs(got.Costs[1]-cost) > 1e-9 || math.Abs(got.Values[1]-(1000+netFlow-cost)) > 1e-9 {
			t.Fatalf("flow conservation: %+v", got)
		}
		if math.Abs(got.Turnover[1]-notional/1000) > 1e-12 || got.FlowAmounts[0] != netFlow {
			t.Fatal("flow or turnover wrong")
		}
		if math.Abs(got.Index[1]-100*(1-cost/denominator)) > 1e-10 {
			t.Fatal("cost did not reduce time-weighted index")
		}
		if got.Contributions[0][1] != 0 || got.Contributions[1][1] != 0 {
			t.Fatal("flow cost attributed to assets")
		}
	}
}

func TestSimulateWithCalendarAndCompatibility(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 30, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)}
	p := optionPortfolio(dates, []float64{100, 200, 200, 200}, []float64{100, 100, 100, 100})
	got, err := SimulateWith(p, SimOptions{RebalancePeriod: Monthly, TransactionCost: .01})
	if err != nil {
		t.Fatal(err)
	}
	if got.Turnover[1] != 0 || got.Turnover[2] <= 0 || math.Abs(got.Turnover[3]) > 1e-12 {
		t.Fatalf("calendar foresight: %v", got.Turnover)
	}
	for _, days := range []int{0, 1, 30} {
		legacy, err := Simulate(p, days)
		if err != nil {
			t.Fatal(err)
		}
		modern, err := SimulateWith(p, SimOptions{RebalanceDays: days})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(legacy, modern) {
			t.Fatal("zero-cost compatibility lost")
		}
	}
}

func TestSimulateWithInvalid(t *testing.T) {
	p := optionPortfolio([]time.Time{day(0), day(1)}, []float64{100, 101}, []float64{100, 100})
	for _, bad := range []SimOptions{{TransactionCost: -.01}, {TransactionCost: math.NaN()}, {TransactionCost: math.Inf(1)}, {TransactionCost: 1}, {RebalanceDays: -1}, {RebalancePeriod: Weekly}, {RebalancePeriod: Monthly, RebalanceDays: 30}} {
		if _, err := SimulateWith(p, bad); err == nil {
			t.Fatalf("accepted %+v", bad)
		}
	}
	for _, bad := range []*Portfolio{nil, {Assets: []Asset{{Weight: 1}}}, {Assets: []Asset{{Weight: math.NaN(), Series: p.Assets[0].Series}}}} {
		if _, err := SimulateWith(bad, SimOptions{}); err == nil {
			t.Fatal("invalid direct portfolio accepted")
		}
	}
}

func TestSimulateWithLeveragedCostAndDebt(t *testing.T) {
	dates := []time.Time{day(0), day(30), day(31)}
	p := optionPortfolio(dates, []float64{100, 200, 200}, []float64{100, 100, 100})
	p.Leverage = true
	p.Assets[0].Weight = .9
	p.Assets[1].Weight = .6
	p.Cash, _ = marketdata.NewSeries("RATE", dates, []float64{10, 10, 10})
	got, err := SimulateWith(p, SimOptions{RebalanceDays: 30, TransactionCost: .01})
	if err != nil {
		t.Fatal(err)
	}
	// Debt starts at 500 and accrues 30 days before the first rebalance.
	before := 1900 - 500*.1*30/365.25
	// Asset trades are 1200-.3*net; net=before-.01*(1200-.3*net).
	net := (before - 12) / .997
	cost := before - net
	if math.Abs(got.Values[1]-net) > 1e-8 || math.Abs(got.Costs[1]-cost) > 1e-8 {
		t.Fatalf("levered funding: %+v, want net %g", got, net)
	}
	// The post-cost target's cash is -.5*net; it incurs one further day
	// of financing, without another scheduled rebalance.
	wantFinal := net - .5*net*.1/365.25
	if math.Abs(got.Values[2]-wantFinal) > 1e-8 {
		t.Fatalf("financing charged on wrong debt: %g, want %g", got.Values[2], wantFinal)
	}
}

func TestSimulateWithCostWithdrawalRuin(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)}
	p := optionPortfolio(dates, []float64{100, 100, 100}, []float64{100, 100, 100})
	p.Withdraw = Flow{Amount: 1000, Period: Monthly}
	got, err := SimulateWith(p, SimOptions{TransactionCost: .01})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ruined || len(got.Dates) != 2 || got.Values[1] != 0 || math.Abs(got.Withdrawn-990) > 1e-10 || math.Abs(got.FlowAmounts[0]+990) > 1e-10 {
		t.Fatalf("unfunded withdrawal: %+v", got)
	}
	if math.Abs(got.Costs[1]-10) > 1e-10 || math.Abs(got.Index[1]-99) > 1e-10 || got.Turnover[1] != 1 {
		t.Fatal("liquidation cost or index wrong")
	}
}

func TestSimulateWithNetFlowAndRebalance(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)}
	p := optionPortfolio(dates, []float64{100, 200, 200}, []float64{100, 100, 100})
	p.Contribute = Flow{Amount: 100, Period: Monthly}
	got, err := SimulateWith(p, SimOptions{RebalancePeriod: Monthly, TransactionCost: .01})
	if err != nil {
		t.Fatal(err)
	}
	// Combined targets are 797.5 each: sell 202.5 and buy 297.5.
	if math.Abs(got.Costs[1]-5) > 1e-10 || math.Abs(got.Values[1]-1595) > 1e-10 {
		t.Fatal("flow and rebalance charged separately")
	}
	if math.Abs(got.Index[1]-150*(1-5.0/1600)) > 1e-10 {
		t.Fatal("new contribution did not bear execution cost")
	}
}

func TestSimulateWithCalendarPeriods(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 30, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 5, 0, 0, 0, 0, time.UTC)}
	p := optionPortfolio(dates, []float64{100, 200, 200, 200, 200, 200}, []float64{100, 100, 100, 100, 100, 100})
	for _, tc := range []struct {
		period Period
		first  int
	}{{Monthly, 2}, {Quarterly, 3}, {Yearly, 4}} {
		got, err := SimulateWith(p, SimOptions{RebalancePeriod: tc.period})
		if err != nil {
			t.Fatal(err)
		}
		first := -1
		for i, v := range got.Turnover {
			if v > 1e-12 {
				first = i
				break
			}
		}
		if first != tc.first {
			t.Fatalf("period %s first trade %d, want %d", tc.period, first, tc.first)
		}
	}
}
