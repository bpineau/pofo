package main

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func day(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestLookup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("query") {
		case "BCOM:IOM":
			fmt.Fprint(w, `{"data":{"security":[{"name":"Some ETF","symbol":"BCOM:LSE:USD","xid":"1"},
				{"name":"Bloomberg Commodity Index","symbol":"BCOM:IOM","xid":"570179"}]}}`)
		case "RENAMED:IOM":
			fmt.Fprint(w, `{"data":{"security":[{"name":"Another index","symbol":"RENAMED:IOM","xid":"2"}]}}`)
		default:
			fmt.Fprint(w, `{"data":{"security":[]}}`)
		}
	}))
	defer srv.Close()
	if xid, err := lookup(srv.URL, "BCOM:IOM", erName); err != nil || xid != "570179" {
		t.Errorf("lookup = %q, %v; want the exact symbol's id", xid, err)
	}
	if _, err := lookup(srv.URL, "RENAMED:IOM", erName); err == nil {
		t.Error("a listing under another name was accepted")
	}
	if _, err := lookup(srv.URL, "ABSENT:IOM", erName); err == nil {
		t.Error("a symbol the search does not carry was accepted")
	}
}

func TestParse(t *testing.T) {
	raw := `{"Dates":["1991-01-02T00:00:00","1991-01-03T00:00:00","1991-01-04T00:00:00"],
		"Elements":[{"Currency":"USD","ComponentSeries":[{"Type":"Open","Values":[1,2,3]},
		{"Type":"Close","Values":[100.0,null,98.5088]}]}]}`
	pts, err := parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 2 || pts[1].Close != 98.5088 || !pts[1].Date.Equal(day("1991-01-04")) {
		t.Errorf("got %v, want the two closes with the null one skipped", pts)
	}
	if _, err := parse([]byte(strings.Replace(raw, `"USD"`, `"GBP"`, 1))); err == nil {
		t.Error("a series quoted in another currency was accepted")
	}
	if _, err := parse([]byte(strings.Replace(raw, "1991-01-04", "1991-01-01", 1))); err == nil {
		t.Error("an out-of-order date was accepted")
	}
}

// record is a synthetic index shaped like the real one: 100 on 1991-01-02, a
// smooth wave through the weekdays up to the day before now, never flat.
func record(now time.Time) []marketdata.Point {
	var out []marketdata.Point
	i := 0
	for d := day(firstDay); d.Before(now); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		out = append(out, marketdata.Point{Date: d, Close: 100 * (1 + 0.3*math.Sin(float64(i)/400)) * (1 + 1e-4*float64(i%7))})
		i++
	}
	return out
}

func TestDropReprints(t *testing.T) {
	pts := []marketdata.Point{
		{Date: day("2002-12-24"), Close: 112.9333},
		{Date: day("2002-12-25"), Close: 112.93}, // the holiday reprint, rounded to the cent
		{Date: day("2002-12-26"), Close: 113.40},
	}
	got, n := dropReprints(pts)
	if n != 1 || len(got) != 2 || !got[1].Date.Equal(day("2002-12-26")) {
		t.Errorf("got %v (%d dropped), want the reprint alone removed", got, n)
	}
}

func TestCheck(t *testing.T) {
	now := day("2026-09-27")
	pts := record(now)
	if err := check(pts, 2, now); err != nil {
		t.Fatalf("a sound record was refused: %v", err)
	}
	cases := map[string]func([]marketdata.Point) ([]marketdata.Point, int){
		"another base": func(p []marketdata.Point) ([]marketdata.Point, int) {
			p[0].Close = 101
			return p, 0
		},
		"a session beyond maxMove": func(p []marketdata.Point) ([]marketdata.Point, int) {
			for i := 5000; i < len(p); i++ {
				p[i].Close *= 1.2
			}
			return p, 0
		},
		"too many round trips": func(p []marketdata.Point) ([]marketdata.Point, int) { return p, maxSpikes + 1 },
		"a stale source": func(p []marketdata.Point) ([]marketdata.Point, int) {
			return p[:len(p)-20], 0
		},
	}
	for name, spoil := range cases {
		p, dropped := spoil(append([]marketdata.Point(nil), pts...))
		if err := check(p, dropped, now); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	flat := append([]marketdata.Point(nil), pts...)
	for i := 3000; i < 3010; i++ {
		flat[i].Close = flat[2999].Close
	}
	if err := frozen(flat); err == nil {
		t.Error("a frozen feed was accepted")
	}
}

func TestCarry(t *testing.T) {
	// The excess return is flat; the total return earns the bill, 2 % a year,
	// and 15.77 % in 2025, so 2025's bill is set to match.
	var er, tr, bill []marketdata.Point
	level := 100.0
	for y := 2009; y <= 2026; y++ {
		rate := 2.0
		if y == 2025 {
			rate = 15.77
		}
		for m := time.January; m <= time.December; m++ {
			bill = append(bill, marketdata.Point{Date: time.Date(y, m, 1, 0, 0, 0, 0, time.UTC), Close: rate})
		}
		if y > 2009 {
			level *= 1 + rate/100
		}
		end := time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC)
		er = append(er, marketdata.Point{Date: end, Close: 100})
		tr = append(tr, marketdata.Point{Date: end, Close: level})
	}
	if err := carry(er, tr, bill); err != nil {
		t.Fatalf("a sound pair was refused: %v", err)
	}
	for i := range tr {
		if tr[i].Date.Year() >= 2015 {
			tr[i].Close *= 1.02
		}
	}
	if err := carry(er, tr, bill); err == nil {
		t.Error("a year two points off the bill was accepted")
	}
}
