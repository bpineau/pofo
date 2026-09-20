package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// asxChartJSON is the real shape of a Yahoo daily chart for an ASX line: the
// meta names the venue's time zone, and each timestamp is that session's
// OPENING instant, 10:00 in Sydney. On summer time (AEDT, UTC+11) that instant
// is 23:00 UTC of the day BEFORE, which is what used to date Monday's session
// on a Sunday.
func asxChartJSON(symbol string, sessions []time.Time, closes []float64) string {
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		panic(err)
	}
	ts, cl := "", ""
	for i, d := range sessions {
		if i > 0 {
			ts, cl = ts+",", cl+","
		}
		open := time.Date(d.Year(), d.Month(), d.Day(), 10, 0, 0, 0, loc)
		ts += fmt.Sprint(open.Unix())
		cl += fmt.Sprint(closes[i])
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":"AUD","symbol":%q,"longName":"Northern Star Resources Ltd","exchangeTimezoneName":"Australia/Sydney"},"timestamp":[%s],"indicators":{"quote":[{"close":[%s]}],"adjclose":[{"adjclose":[%s]}]}}],"error":null}}`,
		symbol, ts, cl, cl)
}

// TestASXSessionsKeepTheirOwnDay: an Australian session must be dated on the
// day it ran, not on the previous UTC one. Sydney is on AEDT in February, so
// every one of these sessions used to arrive a day early, Monday's on a Sunday.
func TestASXSessionsKeepTheirOwnDay(t *testing.T) {
	// Monday 2026-02-02 to Friday 2026-02-06, all on AEDT.
	var sessions []time.Time
	for i := range 5 {
		sessions = append(sessions, time.Date(2026, 2, 2+i, 0, 0, 0, 0, time.UTC))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/NST.AX", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, asxChartJSON("NST.AX", sessions, []float64{22.9, 23.1, 22.7, 23.4, 23.0}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "NST.AX", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != len(sessions) {
		t.Fatalf("got %d points, want %d", len(s.Points), len(sessions))
	}
	for i, p := range s.Points {
		if !p.Date.Equal(sessions[i]) {
			t.Fatalf("session %s dated %s", sessions[i].Format("Mon 2006-01-02"), p.Date.Format("Mon 2006-01-02"))
		}
		if wd := p.Date.Weekday(); wd == time.Saturday || wd == time.Sunday {
			t.Fatalf("the ASX does not trade on %s: %s", wd, p.Date.Format("2006-01-02"))
		}
	}
}

// TestWesternVenuesKeepTheUTCDay: the correction is one-directional. A US
// close stamped at 14:30 UTC reads 09:30 in New York, the same day, and must
// stay there; nothing may push a session backwards onto a Sunday.
func TestWesternVenuesKeepTheUTCDay(t *testing.T) {
	days := testDays(4)
	ts := ""
	for i, d := range days {
		if i > 0 {
			ts += ","
		}
		// 00:30 UTC: the previous evening in New York, the shape that would
		// move a bar backwards if the venue's zone were trusted blindly.
		ts += fmt.Sprint(d.Add(30 * time.Minute).Unix())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/CL=F", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"chart":{"result":[{"meta":{"currency":"USD","symbol":"CL=F","exchangeTimezoneName":"America/New_York"},"timestamp":[%s],"indicators":{"quote":[{"close":[60,61,62,63]}],"adjclose":[{"adjclose":[60,61,62,63]}]}}],"error":null}}`, ts)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "CL=F", days[0])
	if err != nil {
		t.Fatal(err)
	}
	for i, p := range s.Points {
		if !p.Date.Equal(days[i]) {
			t.Fatalf("a western venue must keep its UTC day: %s became %s",
				days[i].Format("2006-01-02"), p.Date.Format("2006-01-02"))
		}
	}
}

// TestMisdatedCacheFileIsRefetched: a file written before the session-day fix
// carries the fault in plain sight, a weekend close on an exchange-traded
// line. Such a file must not answer a fresh request, and every other file of
// the same vintage must keep answering: a blanket invalidation would cost a
// full refetch of every instrument for a fault touching a handful of them.
func TestMisdatedCacheFileIsRefetched(t *testing.T) {
	dir := t.TempDir()
	old := func(symbol string, dates []time.Time) {
		cf := cacheFile{ // no Version: the shape an older release wrote
			Symbol: symbol, Currency: "AUD", Source: "yahoo",
			RequestedFrom: "2020-01-01", FetchedAt: time.Now(),
		}
		for i, d := range dates {
			cf.Dates = append(cf.Dates, d.Format("2006-01-02"))
			cf.Closes = append(cf.Closes, 10+float64(i))
		}
		data, err := json.Marshal(cf)
		if err != nil {
			t.Fatal(err)
		}
		c := NewClient(dir)
		c.writeCacheFile(c.cachePath(symbol), data)
	}
	// 2026-02-01 is a Sunday: NST.AX's Monday session, dated a day early.
	old("NST.AX", []time.Time{
		time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	old("VOO", testDays(3))

	c := NewClient(dir)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, ok := c.loadCache("NST.AX", from); ok {
		t.Error("a weekend-dated exchange series from the old format must be refetched")
	}
	if _, _, ok := c.loadCacheAnyAge("NST.AX", from); !ok {
		t.Error("the stale-cache fallback must still read it: old dates beat no data")
	}
	if _, ok := c.loadCache("VOO", from); !ok {
		t.Error("an unaffected file of the same vintage must keep answering")
	}
}

// TestCurrencyCrossKeepsTheUTCDay: a cross has no exchange, and its dating is
// left exactly as it was (see venueTimezone).
func TestCurrencyCrossKeepsTheUTCDay(t *testing.T) {
	days := testDays(3)
	ts := ""
	for i, d := range days {
		if i > 0 {
			ts += ","
		}
		ts += fmt.Sprint(d.Add(23 * time.Hour).Unix()) // 23:00 UTC, midnight in London on BST
	}
	// A cross with no bundled long proxy, so the series is exactly what the
	// provider sent (see extendFXBack).
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SEKNOK=X", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"chart":{"result":[{"meta":{"currency":"NOK","symbol":"SEKNOK=X","exchangeTimezoneName":"Europe/London"},"timestamp":[%s],"indicators":{"quote":[{"close":[1.15,1.16,1.17]}],"adjclose":[{"adjclose":[1.15,1.16,1.17]}]}}],"error":null}}`, ts)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "SEKNOK=X", days[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 3 {
		t.Fatalf("a cross must keep every point, got %d", len(s.Points))
	}
	for i, p := range s.Points {
		if !p.Date.Equal(days[i]) {
			t.Fatalf("cross point %d dated %s, want %s", i, p.Date.Format("2006-01-02"), days[i].Format("2006-01-02"))
		}
	}
}
