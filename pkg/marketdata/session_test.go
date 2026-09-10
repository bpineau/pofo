package marketdata

import (
	"fmt"
	"net/http"
	"testing"
)

// Fixtures below are the live Yahoo v7 answers captured on 2026-09-09 at 21:10
// New York (marketState POSTPOST for the US lines, PREPRE for the European
// one), stripped to the fields this package reads. The PRE and CLOSED states
// are the same shape with the pre/post fields moved, Yahoo serving one enum
// value per session and the same field names throughout.

// authEndpoint stubs the Yahoo cookie/crumb dance the v7 quote API needs, and
// nothing else: the caller mounts the endpoints its own case is about.
func authEndpoint(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "ck"})
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("crumb1"))
	})
	return mux
}

// quoteEndpoint stubs that dance and the v7 quote endpoint,
// answering with the given result objects.
func quoteEndpoint(t *testing.T, results ...string) *http.ServeMux {
	t.Helper()
	mux := authEndpoint(t)
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		body := ""
		for i, res := range results {
			if i > 0 {
				body += ","
			}
			body += res
		}
		fmt.Fprintf(w, `{"quoteResponse":{"result":[%s],"error":null}}`, body)
	})
	return mux
}

const (
	// DDOG, after hours: the 19:59:25 print is 43 cents above the 16:00 close.
	ddogPost = `{"symbol":"DDOG","currency":"USD","exchangeTimezoneName":"America/New_York",
	 "marketState":"POSTPOST","regularMarketPrice":225.27,"regularMarketTime":1788984001,
	 "postMarketPrice":225.7,"postMarketTime":1788998365}`
	// VT, same session.
	vtPost = `{"symbol":"VT","currency":"USD","exchangeTimezoneName":"America/New_York",
	 "marketState":"POSTPOST","regularMarketPrice":159.89,"regularMarketTime":1788984000,
	 "postMarketPrice":160.05,"postMarketTime":1788997214}`
	// IWDA.AS: an Amsterdam line runs no extended session, so Yahoo serves no
	// pre/post field at all (hasPrePostMarketData false).
	iwdaNone = `{"symbol":"IWDA.AS","currency":"EUR","exchangeTimezoneName":"Europe/Amsterdam",
	 "marketState":"PREPRE","regularMarketPrice":126.05,"regularMarketTime":1788968108}`
	// DDOG the next morning at 08:00 New York: a pre-market print, with last
	// night's post-market one still served alongside it.
	ddogPre = `{"symbol":"DDOG","currency":"USD","exchangeTimezoneName":"America/New_York",
	 "marketState":"PRE","regularMarketPrice":225.27,"regularMarketTime":1788984001,
	 "postMarketPrice":225.7,"postMarketTime":1788998365,
	 "preMarketPrice":228.4,"preMarketTime":1789041600}`
	// DDOG mid-session: the morning's pre-market print is still served, and is
	// now older than the regular one.
	ddogRegular = `{"symbol":"DDOG","currency":"USD","exchangeTimezoneName":"America/New_York",
	 "marketState":"REGULAR","regularMarketPrice":229.1,"regularMarketTime":1789056000,
	 "preMarketPrice":228.4,"preMarketTime":1789041600}`
	// A weekend, every session shut and no off-hours field left.
	ddogClosed = `{"symbol":"DDOG","currency":"USD","exchangeTimezoneName":"America/New_York",
	 "marketState":"CLOSED","regularMarketPrice":225.27,"regularMarketTime":1788984001}`
)

// TestLatestBatchExtendedPost: with the opt-in, the after-hours print wins and
// says so; without it, the very same answer yields the regular close.
func TestLatestBatchExtendedPost(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), quoteEndpoint(t, ddogPost, vtPost))
	defer srv.Close()

	got := c.LatestBatchExtended(t.Context(), []string{"DDOG", "VT"})
	q := got["DDOG"]
	if q.Price != 225.7 || q.Session != "post" || !q.Live {
		t.Fatalf("DDOG extended quote: %+v", q)
	}
	if h, m := q.Time.Hour(), q.Time.Minute(); h != 19 || m != 59 {
		t.Errorf("DDOG post-market time = %02d:%02d New York, want 19:59", h, m)
	}
	if q := got["VT"]; q.Price != 160.05 || q.Session != "post" {
		t.Errorf("VT extended quote: %+v", q)
	}

	plain := c.LatestBatch(t.Context(), []string{"DDOG", "VT"})
	if q := plain["DDOG"]; q.Price != 225.27 || q.Session != "regular" {
		t.Fatalf("the default path must not move: %+v", q)
	}
	if q := plain["VT"]; q.Price != 159.89 || q.Session != "regular" {
		t.Fatalf("the default path must not move: %+v", q)
	}
}

// TestLatestBatchExtendedStates walks the sessions a US line goes through.
func TestLatestBatchExtendedStates(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fixture string
		price   float64
		session string
	}{
		{"pre-market beats last night's post", ddogPre, 228.4, "pre"},
		{"regular session ignores the stale pre print", ddogRegular, 229.1, "regular"},
		{"closed falls back to the regular close", ddogClosed, 225.27, "regular"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, srv := newTestClient(t, t.TempDir(), quoteEndpoint(t, tc.fixture))
			defer srv.Close()
			q := c.LatestBatchExtended(t.Context(), []string{"DDOG"})["DDOG"]
			if q.Price != tc.price || q.Session != tc.session {
				t.Fatalf("got %v/%q, want %v/%q", q.Price, q.Session, tc.price, tc.session)
			}
		})
	}
}

// TestLatestBatchExtendedNoSession: a venue without extended hours serves no
// pre/post field, and its quote is the regular one, opt-in or not.
func TestLatestBatchExtendedNoSession(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), quoteEndpoint(t, iwdaNone))
	defer srv.Close()

	q := c.LatestBatchExtended(t.Context(), []string{"IWDA.AS"})["IWDA.AS"]
	if q.Price != 126.05 || q.Session != "regular" || q.Currency != "EUR" {
		t.Fatalf("IWDA.AS quote: %+v", q)
	}
}

// TestLatestAnyExtendedHours: the single-quote opt-in reaches the same print.
func TestLatestAnyExtendedHours(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), quoteEndpoint(t, ddogPost))
	defer srv.Close()

	q, err := c.LatestAny(t.Context(), []string{"DDOG"}, QuoteOptions{ExtendedHours: true})
	if err != nil {
		t.Fatal(err)
	}
	if q.Price != 225.7 || q.Session != "post" {
		t.Fatalf("extended quote: %+v", q)
	}
}

// TestLatestExtendedFallsBackToSpot: the extended leg needs a crumb and can
// fail on its own; the caller still gets the regular quote.
func TestLatestExtendedFallsBackToSpot(t *testing.T) {
	mux := authEndpoint(t)
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v8/finance/chart/DDOG", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[{"meta":{"currency":"USD","exchangeTimezoneName":"America/New_York","regularMarketPrice":225.27,"regularMarketTime":1788984001}}],"error":null}}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.LatestAny(t.Context(), []string{"DDOG"}, QuoteOptions{ExtendedHours: true})
	if err != nil {
		t.Fatal(err)
	}
	if q.Price != 225.27 || q.Session != "regular" || !q.Live {
		t.Fatalf("fallback quote: %+v", q)
	}
}

// TestQuoteSessionOfLastClose: a price that is not a market field names no
// session, so a caller can never mistake a NAV or a close for a live print.
func TestQuoteSessionOfLastClose(t *testing.T) {
	days := testDays(3)
	mux := authEndpoint(t)
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, []float64{100, 101.5, 99}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest(t.Context(), "VOO")
	if err != nil {
		t.Fatal(err)
	}
	if q.Live || q.Session != "" {
		t.Fatalf("a daily close claims no session: %+v", q)
	}
}

func TestFreshestSession(t *testing.T) {
	price := func(v float64) *float64 { return &v }
	at := func(v int64) *int64 { return &v }
	for _, tc := range []struct {
		name                string
		prePrice, postPrice *float64
		preTime, postTime   *int64
		wantPrice, wantTime float64
		wantSession         string
	}{
		{name: "nothing but the regular print", wantPrice: 100, wantTime: 1000, wantSession: "regular"},
		{name: "a newer pre print wins", prePrice: price(101), preTime: at(1100), wantPrice: 101, wantTime: 1100, wantSession: "pre"},
		{name: "an older pre print loses", prePrice: price(101), preTime: at(900), wantPrice: 100, wantTime: 1000, wantSession: "regular"},
		{name: "the freshest of the two wins", prePrice: price(101), preTime: at(1100), postPrice: price(102), postTime: at(1200), wantPrice: 102, wantTime: 1200, wantSession: "post"},
		{name: "an older post print loses to a newer pre one", prePrice: price(101), preTime: at(1300), postPrice: price(102), postTime: at(1200), wantPrice: 101, wantTime: 1300, wantSession: "pre"},
		{name: "a price without a timestamp is not a print", prePrice: price(101), wantPrice: 100, wantTime: 1000, wantSession: "regular"},
		{name: "a timestamp without a price is not a print", preTime: at(1100), wantPrice: 100, wantTime: 1000, wantSession: "regular"},
		{name: "a zero price is not a print", prePrice: price(0), preTime: at(1100), wantPrice: 100, wantTime: 1000, wantSession: "regular"},
		{name: "a tie keeps the regular print", prePrice: price(101), preTime: at(1000), wantPrice: 100, wantTime: 1000, wantSession: "regular"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, u, s := freshestSession(100, 1000, tc.prePrice, tc.preTime, tc.postPrice, tc.postTime)
			if p != tc.wantPrice || float64(u) != tc.wantTime || s != tc.wantSession {
				t.Fatalf("got %v/%d/%q, want %v/%v/%q", p, u, s, tc.wantPrice, tc.wantTime, tc.wantSession)
			}
		})
	}
}
