package marketdata

import "testing"

func TestRestoreFTCatBondQuote(t *testing.T) {
	for _, tc := range []struct {
		id                 string
		day, end           int
		before, nav, after float64
	}{
		{"IE00B3Q8M574", 9, 16, 12.5132, 12.5657, 12.6723},
		{"LI0049587301", 13, 31, 1181.13, 1186.49, 1185.29},
	} {
		s := &Series{Symbol: tc.id, Currency: "EUR", Points: []Point{
			{Date: date(2017, 9, 29), Close: tc.before},
			{Date: date(2017, 10, tc.end), Close: tc.after},
		}}
		restoreFTCatBondQuote(s)
		if len(s.Points) != 3 || !s.Points[1].Date.Equal(date(2017, 10, tc.day)) || s.Points[1].Close != tc.nav {
			t.Fatalf("%s lost its observed NAV: %v", tc.id, s.Points)
		}
		s.Points[1].Close += 0.01
		restoreFTCatBondQuote(s)
		if len(s.Points) != 3 || s.Points[1].Close != tc.nav+0.01 {
			t.Fatalf("%s overwrote a provider revision", tc.id)
		}
		s.Points = append(s.Points[:1], s.Points[2])
		s.Points[0].Close *= 2
		restoreFTCatBondQuote(s)
		if len(s.Points) != 2 {
			t.Fatalf("%s filled a gap on a different scale", tc.id)
		}
	}
}
