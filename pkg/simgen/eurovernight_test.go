package simgen

import (
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// The euro overnight index accrues ACT/360 over the CALENDAR days between two
// publications, as the published compounded indices do: a Friday rate pays the
// weekend. Reading it as one accrual per quote would lose three days a week.
func TestEurOvernightAccruesActual360(t *testing.T) {
	f := fakeFetcher{
		"^EONIA":      levelsOn("^EONIA", []int{0, 3, 4}, 3.60),
		"EURCASH-EUR": nil, // no deep tail: the index stands on its own
	}
	s, err := eurOvernightDaily(f, day(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 3 {
		t.Fatalf("%d points, want 3", len(s.Points))
	}
	near(t, "three-day accrual", s.Points[1].Close, 100*(1+0.036*3/360), 1e-9)
	near(t, "one-day accrual", s.Points[2].Close, s.Points[1].Close*(1+0.036*1/360), 1e-9)
}

// ESTR takes over from its own first date and EONIA covers what is before it.
// The order matters: over their overlap EONIA is DEFINED as ESTR plus 8.5 basis
// points, so preferring EONIA would ship a rate that is knowingly too high.
func TestEurOvernightPrefersESTROverEONIA(t *testing.T) {
	f := fakeFetcher{
		"^EONIA": levelsOn("^EONIA", []int{0, 1, 2, 3, 4}, 2.085),
		"^ESTR":  levelsOn("^ESTR", []int{2, 3, 4}, 2.00),
	}
	rate := overnightEUR(f, day(0))
	want := []float64{2.085, 2.085, 2.00, 2.00, 2.00}
	if len(rate.Points) != len(want) {
		t.Fatalf("%d points, want %d", len(rate.Points), len(want))
	}
	for i, w := range want {
		if rate.Points[i].Close != w {
			t.Errorf("day %d: rate %v, want %v", i, rate.Points[i].Close, w)
		}
	}
}

// With no overnight rate reachable, the 3-month money-market index stands: a
// cash sleeve reaching 1994 must not fail because a rate feed starting in 1999
// is down.
func TestEurOvernightFallsBackToTheMoneyMarketIndex(t *testing.T) {
	monthly := &marketdata.Series{Symbol: "EURCASH-EUR"}
	for i := range 5 {
		monthly.Points = append(monthly.Points,
			marketdata.Point{Date: day(i * 30), Close: 100 * (1 + 0.002*float64(i))})
	}
	s, err := eurOvernightDaily(fakeFetcher{"EURCASH-EUR": monthly}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) < 80 {
		t.Fatalf("%d points, want the interpolated money-market index", len(s.Points))
	}
	if got := s.Last().Close; got != monthly.Last().Close {
		t.Errorf("last level %v, want the index's own %v", got, monthly.Last().Close)
	}
}
