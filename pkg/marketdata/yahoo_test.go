package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// serveBody points every chart/search path of a stubbed client at one payload.
func serveBody(t *testing.T, body string) *Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
	c, srv := newTestClient(t, "", mux)
	t.Cleanup(srv.Close)
	return c
}

// TestFetchYahooPayloadEdges walks the answers the chart API really gives when
// something is off: a truncated body, a symbol it does not carry, a result
// without the close column, holes and repeats inside the columns.
func TestFetchYahooPayloadEdges(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		err    string // expected error fragment ("" = success)
		absent bool   // the failure must read as evidence about the identifier
		check  func(*testing.T, *Series)
	}{
		{
			name: "truncated body",
			body: `{"chart":{"result":[{`,
			err:  "unreadable yahoo response",
		},
		{
			name:   "the API reports an unknown symbol",
			body:   `{"chart":{"result":null,"error":{"code":"Not Found","description":"No data found, symbol may be delisted"}}}`,
			err:    "No data found",
			absent: true,
		},
		{
			name:   "no result at all",
			body:   `{"chart":{"result":[],"error":null}}`,
			err:    "empty response",
			absent: true,
		},
		{
			name: "timestamps without a close column",
			body: `{"chart":{"result":[{"meta":{"currency":"USD"},"timestamp":[1578315600],"indicators":{}}],"error":null}}`,
			err:  "no close series",
		},
		{
			name: "short name stands in for the long one",
			body: `{"chart":{"result":[{"meta":{"currency":"USD","shortName":"Short Co"},"timestamp":[1578315600],` +
				`"indicators":{"quote":[{"close":[10]}]}}],"error":null}}`,
			check: func(t *testing.T, s *Series) {
				if s.Name != "Short Co" {
					t.Errorf("name = %q, want the shortName", s.Name)
				}
			},
		},
		{
			name: "no name at all falls back to the symbol",
			body: `{"chart":{"result":[{"meta":{"currency":"USD"},"timestamp":[1578315600],` +
				`"indicators":{"quote":[{"close":[10]}]}}],"error":null}}`,
			check: func(t *testing.T, s *Series) {
				if s.Name != "TST" {
					t.Errorf("name = %q, want the requested symbol", s.Name)
				}
			},
		},
		{
			name: "holes and a repeated day",
			// Three timestamps on two civil days: a null, then the same day
			// quoted twice (Yahoo repeats the current one), the later value
			// winning. The adjclose column is short, so the raw one is read.
			body: `{"chart":{"result":[{"meta":{"currency":"USD","longName":"Test"},` +
				`"timestamp":[1578315600,1578402000,1578412000],` +
				`"indicators":{"quote":[{"close":[null,10,11]}],"adjclose":[{"adjclose":[10]}]}}],"error":null}}`,
			check: func(t *testing.T, s *Series) {
				if len(s.Points) != 1 || s.Points[0].Close != 11 {
					t.Errorf("points = %+v, want one point holding the later value", s.Points)
				}
			},
		},
		{
			name: "a zero dividend is not a dividend",
			body: `{"chart":{"result":[{"meta":{"currency":"USD","longName":"Test"},"timestamp":[1578315600],` +
				`"events":{"dividends":{"1":{"amount":0,"date":1578315600},"2":{"amount":1.5,"date":1578315600}}},` +
				`"indicators":{"quote":[{"close":[10]}]}}],"error":null}}`,
			check: func(t *testing.T, s *Series) {
				if len(s.Dividends) != 1 || s.Dividends[0].Amount != 1.5 {
					t.Errorf("dividends = %+v, want only the positive one", s.Dividends)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := serveBody(t, tc.body)
			s, err := c.fetchYahoo(context.Background(), "TST", d(2020, 1, 1), false)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want one about %q", err, tc.err)
				}
				if got := absent(err); got != tc.absent {
					t.Errorf("absent(err) = %v, want %v", got, tc.absent)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, s)
		})
	}
}

func TestFetchYahooIntradayPayloadEdges(t *testing.T) {
	cases := []struct {
		name string
		body string
		err  string
		want int // points expected on success
	}{
		{"truncated body", `{"chart":{"result":[{`, "unreadable yahoo intraday response", 0},
		{"the API reports an error", `{"chart":{"error":{"description":"boom"}}}`, "yahoo intraday: boom", 0},
		{"no result", `{"chart":{"result":[]}}`, ErrNotCovered.Error(), 0},
		{
			name: "holes and a short close column are skipped",
			body: `{"chart":{"result":[{"meta":{"currency":"USD","shortName":"Short"},` +
				`"timestamp":[1578315600,1578315900,1578316200],` +
				`"indicators":{"quote":[{"close":[null,10]}]}}]}}`,
			want: 1,
		},
		{
			name: "an unknown exchange time zone falls back to UTC",
			body: `{"chart":{"result":[{"meta":{"currency":"USD","exchangeTimezoneName":"Mars/Olympus"},` +
				`"timestamp":[1578315600],"indicators":{"quote":[{"close":[10]}]}}]}}`,
			want: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := serveBody(t, tc.body)
			s, err := c.fetchYahooIntraday(context.Background(), "TST")
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want one about %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(s.Points) != tc.want {
				t.Errorf("points = %+v, want %d", s.Points, tc.want)
			}
			if s.Last().Close != 10 || s.First().Close != 10 {
				t.Errorf("first/last = %v/%v", s.First(), s.Last())
			}
		})
	}
	// The zero value of both accessors, on an empty series.
	empty := &IntradaySeries{}
	if empty.First() != (IntradayPoint{}) || empty.Last() != (IntradayPoint{}) {
		t.Error("an empty intraday series must yield the zero point")
	}
}

func TestFetchYahooSpotPayloadEdges(t *testing.T) {
	cases := []struct {
		name string
		body string
		err  string
	}{
		{"truncated body", `{"chart":{"result":[{`, "unreadable yahoo spot response"},
		{"the API reports an error", `{"chart":{"error":{"description":"boom"}}}`, "yahoo spot: boom"},
		{"no result", `{"chart":{"result":[]}}`, ErrNotCovered.Error()},
		{
			name: "a price with no timestamp is not a quote",
			body: `{"chart":{"result":[{"meta":{"currency":"USD","regularMarketPrice":10}}]}}`,
			err:  ErrNotCovered.Error(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := serveBody(t, tc.body)
			if _, err := c.fetchYahooSpot(context.Background(), "TST"); err == nil ||
				!strings.Contains(err.Error(), tc.err) {
				t.Fatalf("error = %v, want one about %q", err, tc.err)
			}
		})
	}
	// An unknown time zone still yields a quote, timed in UTC.
	c := serveBody(t, `{"chart":{"result":[{"meta":{"currency":"USD","exchangeTimezoneName":"Mars/Olympus",`+
		`"regularMarketPrice":10,"regularMarketTime":1578315600}}]}}`)
	q, err := c.fetchYahooSpot(context.Background(), "TST")
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Price != 10 || q.Time.Location() != time.UTC {
		t.Errorf("quote = %+v, want a live UTC-timed one", q)
	}
}

func TestSearchPayloadEdges(t *testing.T) {
	c := serveBody(t, `{"quotes":[`)
	if _, err := c.search(context.Background(), "x"); err == nil ||
		!strings.Contains(err.Error(), "unreadable search response") {
		t.Fatalf("error = %v, want one about an unreadable payload", err)
	}
	// Symbol-less rows are dropped; nothing left reads as an absence, which
	// is what tells Fetch the identifier itself is the problem.
	c = serveBody(t, `{"quotes":[{"symbol":"","longname":"Nameless"}]}`)
	_, err := c.search(context.Background(), "x")
	if err == nil || !errors.Is(err, errAbsent) {
		t.Fatalf("error = %v, want an absence", err)
	}
	// shortname stands in for a missing longname.
	c = serveBody(t, `{"quotes":[{"symbol":"TST","shortname":"Short Co","quoteType":"ETF"}]}`)
	got, err := c.search(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Short Co" || got[0].QuoteType != "ETF" {
		t.Errorf("candidates = %+v", got)
	}
}
