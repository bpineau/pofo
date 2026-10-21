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

// TestCurrencyCrossDatedOnTheSessionItCloses: Yahoo stamps every daily bar of
// a cross at MIDNIGHT LONDON, and since 2010-08-04 the bar's close is the rate
// at that instant, the end of the PREVIOUS session (open and close equal, the
// level within 0.1 % of the previous day's FRED noon fixing). The bar stamped
// Monday 00:00 is therefore Friday's close, and so on through the week, on
// either side of a clock change. Before the switch a bar carried its own
// day's rate. A bar stamped at any other instant is the live quote and keeps
// its day. The old UTC dating put the whole of winter a day late and, in
// summer, dated Monday's bar on a Sunday and left Friday empty.
func TestCurrencyCrossDatedOnTheSessionItCloses(t *testing.T) {
	lon, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	at := func(y int, m time.Month, d, hh, mm int) time.Time { return time.Date(y, m, d, hh, mm, 0, 0, lon) }
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
	bars := []struct {
		stamp time.Time
		close float64
		want  time.Time
	}{
		// Before the switch: the bar is its own day, summer time or not.
		{at(2009, 6, 1, 0, 0), 1.41, day(2009, 6, 1)}, // Monday, BST: used to land on Sunday
		{at(2009, 6, 5, 0, 0), 1.42, day(2009, 6, 5)},
		// After it, across the March clock change (GMT -> BST on 2024-03-31).
		{at(2024, 3, 28, 0, 0), 1.0830, day(2024, 3, 27)},
		{at(2024, 3, 29, 0, 0), 1.0790, day(2024, 3, 28)},
		{at(2024, 4, 1, 0, 0), 1.0786, day(2024, 3, 29)}, // Monday's bar is Friday's close
		{at(2024, 4, 2, 0, 0), 1.0740, day(2024, 4, 1)},
		// And across the October one (BST -> GMT on 2024-10-27).
		{at(2024, 10, 24, 0, 0), 1.0781, day(2024, 10, 23)},
		{at(2024, 10, 25, 0, 0), 1.0824, day(2024, 10, 24)},
		{at(2024, 10, 28, 0, 0), 1.0797, day(2024, 10, 25)},
		{at(2024, 10, 29, 0, 0), 1.0817, day(2024, 10, 28)},
		// The live quote, stamped when it was taken: its own day.
		{at(2024, 10, 30, 2, 11), 1.0821, day(2024, 10, 30)},
	}
	ts, cl := "", ""
	for i, b := range bars {
		if i > 0 {
			ts, cl = ts+",", cl+","
		}
		ts += fmt.Sprint(b.stamp.Unix())
		cl += fmt.Sprint(b.close)
	}
	// A cross with no bundled long proxy, so the series is exactly what the
	// provider sent (see extendFXBack).
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SEKNOK=X", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"chart":{"result":[{"meta":{"currency":"NOK","symbol":"SEKNOK=X","exchangeTimezoneName":"Europe/London"},"timestamp":[%s],"indicators":{"quote":[{"close":[%s]}],"adjclose":[{"adjclose":[%s]}]}}],"error":null}}`, ts, cl, cl)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "SEKNOK=X", day(2009, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != len(bars) {
		t.Fatalf("a cross must keep every bar, got %d points for %d bars", len(s.Points), len(bars))
	}
	for i, p := range s.Points {
		if !p.Date.Equal(bars[i].want) || p.Close != bars[i].close {
			t.Errorf("bar stamped %s: %.4f dated %s, want %.4f on %s", bars[i].stamp.Format("Mon 2006-01-02 15:04 MST"),
				p.Close, p.Date.Format("Mon 2006-01-02"), bars[i].close, bars[i].want.Format("Mon 2006-01-02"))
		}
		if wd := p.Date.Weekday(); wd == time.Saturday || wd == time.Sunday {
			t.Errorf("a cross point dated on a %s: %s", wd, p.Date.Format("2006-01-02"))
		}
	}
}

// TestCrossCacheFromBeforeTheFixIsRefetched: a cross file written under an
// older format carries the old dating with no fingerprint to recognise it by
// (winter is simply a day late), so every such cross is refetched once; an
// exchange-traded file of the same vintage keeps answering.
func TestCrossCacheFromBeforeTheFixIsRefetched(t *testing.T) {
	dir := t.TempDir()
	write := func(symbol string, version int) {
		cf := cacheFile{Version: version, Symbol: symbol, Currency: "USD", Source: "yahoo",
			RequestedFrom: "2020-01-01", FetchedAt: time.Now()}
		for i, d := range testDays(3) {
			cf.Dates = append(cf.Dates, d.Format("2006-01-02"))
			cf.Closes = append(cf.Closes, 1.1+float64(i)/100)
		}
		data, err := json.Marshal(cf)
		if err != nil {
			t.Fatal(err)
		}
		c := NewClient(dir)
		c.writeCacheFile(c.cachePath(symbol), data)
	}
	write("SEKNOK=X", 1)
	write("VOO", 1)
	write("NOKSEK=X", cacheFormat)

	c := NewClient(dir)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, ok := c.loadCache("SEKNOK=X", from); ok {
		t.Error("a cross cached before the cross dating fix must be refetched")
	}
	if _, ok := c.loadCache("NOKSEK=X", from); !ok {
		t.Error("a cross written under the current format must answer")
	}
	if _, ok := c.loadCache("VOO", from); !ok {
		t.Error("an exchange-traded file of the same vintage must keep answering")
	}
}
