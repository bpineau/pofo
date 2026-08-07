package simgen

import (
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// levelsOn builds a rate series on the given day indices, so a test can give
// the overnight rate a calendar the bill rate does not have.
func levelsOn(symbol string, days []int, level float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for _, i := range days {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: level})
	}
	return s
}

// The financing series must carry the overnight level where an overnight rate
// exists and the bill level before it, splicing SOFR over the funds rate.
func TestFinancingSplicesOvernightOverBills(t *testing.T) {
	f := financed(fakeFetcher{
		"^IRX":      mkLevels("^IRX", 10, 4.00),
		"^FEDFUNDS": levelsOn("^FEDFUNDS", []int{3, 4, 5, 6, 7, 8, 9}, 4.50),
		"^SOFR":     levelsOn("^SOFR", []int{7, 8, 9}, 4.75),
	})
	s, err := f.Fetch(usdOvernight, day(0))
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{4.00, 4.00, 4.00, 4.50, 4.50, 4.50, 4.50, 4.75, 4.75, 4.75}
	if len(s.Points) != len(want) {
		t.Fatalf("got %d points, want %d", len(s.Points), len(want))
	}
	for i, w := range want {
		if s.Points[i].Close != w {
			t.Errorf("day %d: financing = %v, want %v", i, s.Points[i].Close, w)
		}
		if !s.Points[i].Date.Equal(day(i)) {
			t.Errorf("day %d: date = %s", i, s.Points[i].Date)
		}
	}
}

// The federal funds rate is published for every calendar day and Align takes
// the UNION of its inputs' dates, so a financing series that kept the policy
// feed's own calendar would inject weekends into every frame it joins. It must
// stay on the bill rate's trading calendar.
func TestFinancingKeepsTheBillCalendar(t *testing.T) {
	weekdays := []int{0, 1, 4, 5}
	f := financed(fakeFetcher{
		"^IRX":      levelsOn("^IRX", weekdays, 4.00),
		"^FEDFUNDS": mkLevels("^FEDFUNDS", 6, 4.50), // every calendar day
	})
	s, err := f.Fetch(usdOvernight, day(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != len(weekdays) {
		t.Fatalf("got %d points, want the bill calendar's %d", len(s.Points), len(weekdays))
	}
	for i, d := range weekdays {
		if !s.Points[i].Date.Equal(day(d)) {
			t.Errorf("point %d dated %s, want %s", i, s.Points[i].Date, day(d))
		}
	}
}

// With no policy series reachable, the bill rate stands: a reconstruction that
// reaches 1871 must not fail because a rate feed starting in 1954 is down.
func TestFinancingFallsBackToBills(t *testing.T) {
	f := financed(fakeFetcher{"^IRX": mkLevels("^IRX", 5, 4.00)})
	s, err := f.Fetch(usdOvernight, day(0))
	if err != nil {
		t.Fatal(err)
	}
	for i, p := range s.Points {
		if p.Close != 4.00 {
			t.Fatalf("day %d: financing = %v, want the bill rate 4.00", i, p.Close)
		}
	}
}

// An Excess leg finances at the overnight rate while a cash leg keeps earning
// the bill rate: the whole point of the convention is that the two differ.
func TestCompositeFinancesExcessAtTheOvernightRate(t *testing.T) {
	const n = 30
	f := financed(fakeFetcher{
		"^IRX":      mkLevels("^IRX", n, 2.52), // 0.0001/day
		"^FEDFUNDS": mkLevels("^FEDFUNDS", n, 5.04),
		"BD":        mkSeries("BD", n, 0.0),
	})
	fr, err := BuildFrame(f, []string{usdOvernight, "^IRX", "BD"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	near(t, "financing accrual", fr.Returns[usdOvernight][1], 0.0002, 1e-12)
	values, err := Composite(fr, []Leg{
		{ID: "BD", Weight: 1, Excess: true},
		{ID: "^IRX", Weight: 1},
	}, usdOvernight, 0)
	if err != nil {
		t.Fatal(err)
	}
	// A flat bond financed at 0.0002/day against collateral earning 0.0001/day
	// loses the difference every day.
	near(t, "daily return", values[1]/values[0]-1, -0.0001, 1e-12)
}

// The policy and money-market symbols are rates, not prices: reading one as a
// price would turn a 4.5 % level into a 0 % return.
func TestIsRateKnowsThePolicyFamily(t *testing.T) {
	for _, id := range append(marketdata.RateSymbols(), "^IRX", usdOvernight) {
		if !isRate(id) {
			t.Errorf("isRate(%q) = false, want true", id)
		}
	}
	for _, id := range []string{"VFINX", "GC=F", "EURCASH-EUR"} {
		if isRate(id) {
			t.Errorf("isRate(%q) = true, want false", id)
		}
	}
}
