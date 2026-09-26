package marketdata

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"
)

// stooqFXCSV is a Stooq daily-history fixture for the conventional eurusd
// listing (USD per EUR). The 1998 row predates the euro: it is a synthetic
// backcast on the provider side and must never survive into a series.
const stooqFXCSV = "Date,Open,High,Low,Close,Volume\n" +
	"1998-12-30,9.99,9.99,9.99,9.99,0\n" +
	"2020-01-06,1.11,1.12,1.10,1.1194,0\n" +
	"2020-01-07,1.11,1.12,1.10,1.1025,0\n"

func TestStooqSymbolFX(t *testing.T) {
	cases := []struct {
		symbol string
		ss     string
		invert bool
	}{
		// Conventional direction: served as-is.
		{"EURUSD=X", "eurusd", false},
		{"GBPUSD=X", "gbpusd", false},
		{"USDJPY=X", "usdjpy", false},
		{"EURGBP=X", "eurgbp", false},
		{"AUDUSD=X", "audusd", false},
		// Reciprocal direction: Stooq lists the conventional pair only.
		{"USDEUR=X", "eurusd", true},
		{"JPYUSD=X", "usdjpy", true},
		{"GBPEUR=X", "eurgbp", true},
		{"CHFEUR=X", "eurchf", true},
		// A minor currency ranks below every major.
		{"USDSEK=X", "usdsek", false},
		{"SEKUSD=X", "usdsek", true},
		// Not a mappable cross.
		{"SEKNOK=X", "", false}, // no major leg: unvetted on stooq
		{"EUREUR=X", "", false},
		{"FOOBA=X", "", false},
		// Non-FX symbols keep their existing mapping.
		{"VOO", "voo.us", false},
		{"CL=F", "cl.f", false},
		{"IWDA.AS", "", false},
	}
	for _, tc := range cases {
		ss, invert := stooqSymbol(tc.symbol)
		if ss != tc.ss || invert != tc.invert {
			t.Errorf("stooqSymbol(%q) = %q, %v; want %q, %v", tc.symbol, ss, invert, tc.ss, tc.invert)
		}
	}
}

func TestHistoryFXFallsBackToStooq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		if s := r.URL.Query().Get("s"); s != "eurusd" {
			t.Errorf("stooq symbol = %q, want eurusd (the conventional listing)", s)
		}
		fmt.Fprint(w, stooqFXCSV)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "USDEUR=X", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "stooq" || s.Currency != "EUR" {
		t.Fatalf("source/currency misread: %+v", s)
	}
	// USDEUR=X is the reciprocal of the eurusd listing.
	if got, want := s.Last().Close, 1/1.1025; math.Abs(got-want) > 1e-9 {
		t.Errorf("last close = %v, want %v (inverted eurusd)", got, want)
	}
	// The synthetic pre-euro row is dropped; the vetted bundled ECU/EUR
	// anchors take its place, so the series still reaches back to 1971.
	for _, p := range s.Points {
		if math.Abs(p.Close-1/9.99) < 1e-3 {
			t.Errorf("synthetic pre-euro stooq row survived: %+v", p)
		}
	}
	if first := s.First().Date; first.Year() != 1971 {
		t.Errorf("series starts %s, want 1971 (bundled ECU/DM/EUR splice)", first.Format("2006-01"))
	}
}

// TestStooqSymbolNonFX pins the non-FX half of the mapping table: the handful
// of indices and commodities Stooq lists under its own names, the ".us" suffix
// every US ticker takes, and the shapes that have no equivalent at all.
func TestStooqSymbolNonFX(t *testing.T) {
	cases := map[string]string{
		"^GSPC":   "^spx",
		"^NDX":    "^ndx",
		"^DJI":    "^dji",
		"^IXIC":   "^ndq",
		"XAUUSD":  "xauusd",
		"XAGUSD":  "xagusd",
		"CL=F":    "cl.f",
		"CL.F":    "cl.f",
		"SPY":     "spy.us",
		"^VIX":    "", // an index Stooq is not vetted for
		"SXR8.DE": "", // an exchange suffix
		"GC=F":    "", // another futures chain
	}
	for in, want := range cases {
		if got, invert := stooqSymbol(in); got != want || invert {
			t.Errorf("stooqSymbol(%q) = %q, %v; want %q, false", in, got, invert, want)
		}
	}
}

// TestFetchStooqPayloadEdges walks what the CSV endpoint answers when it has
// nothing: an unmapped symbol, the anti-bot HTML page, a header-only file. Each
// must read as an ABSENCE (evidence about the identifier), never as an outage,
// and rows it cannot parse must be skipped rather than poison the series.
func TestFetchStooqPayloadEdges(t *testing.T) {
	cases := []struct {
		name    string
		symbol  string
		body    string
		err     string
		absent  bool
		points  int
		wantCcy string
	}{
		{name: "no stooq equivalent", symbol: "IWDA.AS", err: "no stooq equivalent", absent: true},
		{name: "the anti-bot page", symbol: "SPY", body: stooqChallengeHTML, err: "no data for", absent: true},
		{name: "header only", symbol: "SPY", body: "Date,Open,High,Low,Close,Volume\n", err: "no data for", absent: true},
		{name: "a header too narrow", symbol: "SPY", body: "Date,Close\n2020-01-06,1\n", err: "no data for", absent: true},
		{
			name:   "unreadable rows are skipped",
			symbol: "SPY",
			body: "Date,Open,High,Low,Close,Volume\n" +
				"06/01/2020,1,1,1,10,0\n" + // unparseable date
				"2020-01-07,1,1,1,abc,0\n" + // unparseable close
				"2020-01-08,1,1,1,0,0\n" + // non-positive close
				"2020-01-09,1,1,1,11,0\n",
			points: 1, wantCcy: "USD",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tc.body)
			})
			c, srv := newTestClient(t, "", mux)
			defer srv.Close()
			s, err := c.fetchStooq(context.Background(), tc.symbol, d(2020, 1, 1))
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
			if len(s.Points) != tc.points || s.Currency != tc.wantCcy {
				t.Fatalf("series = %+v (%s)", s.Points, s.Currency)
			}
		})
	}
}

// TestFetchStooqNamesTheCommodities: the three symbols Stooq serves as spot
// commodities are labelled and denominated here, nowhere else.
func TestFetchStooqNamesTheCommodities(t *testing.T) {
	body := "Date,Open,High,Low,Close,Volume\n2020-01-06,1,1,1,1500,0\n2020-01-07,1,1,1,1510,0\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
	c, srv := newTestClient(t, "", mux)
	defer srv.Close()
	cases := map[string]string{
		"XAUUSD": "Gold (XAU/USD spot)",
		"XAGUSD": "Silver (XAG/USD spot)",
		"CL=F":   "WTI crude oil (continuous futures)",
	}
	for symbol, want := range cases {
		s, err := c.fetchStooq(context.Background(), symbol, d(2020, 1, 1))
		if err != nil {
			t.Fatal(err)
		}
		if s.Name != want || s.Currency != "USD" || len(s.Points) != 2 {
			t.Errorf("%s: name=%q currency=%q points=%d", symbol, s.Name, s.Currency, len(s.Points))
		}
	}
}
