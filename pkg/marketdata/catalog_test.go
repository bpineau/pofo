package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCatalogPinsResolutionWithoutSearch(t *testing.T) {
	// A catalog asset triggers no network search: only the pinned source
	// is queried.
	days := testDays(70)
	closes := make([]float64, len(days))
	for i := range closes {
		closes[i] = 100 + float64(i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, closes))
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		t.Error("the catalog must bypass the search")
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(context.Background(), "VOO", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "VOO" || len(s.Points) != 70 {
		t.Fatalf("catalog resolution: %+v", s)
	}
}

func TestFundISINEmbeddedList(t *testing.T) {
	cases := map[string]string{
		"IWDA": "IE00B4L5Y983",
		"iwda": "IE00B4L5Y983",
		"CSPX": "IE00B5BMR087",
		"VWCE": "IE00BK5BQT80",
		"EXSA": "DE0002635307",
	}
	for ticker, want := range cases {
		got, ok := FundISIN(ticker)
		if !ok || got != want {
			t.Errorf("FundISIN(%q) = %q, %v; want %q", ticker, got, ok, want)
		}
	}
	if _, ok := FundISIN("AAPL"); ok {
		t.Error("AAPL must not appear in the fund list")
	}
	// The iShares PEA ETFs are durably pinned (ticker and ISIN).
	for _, q := range []string{"SPEA", "IE000DQLYVB9", "WPEA", "IE0002XZSHO1"} {
		if _, ok := catalogResolution(q); !ok {
			t.Errorf("%s should be pinned in the catalog", q)
		}
	}
	if res, _ := catalogResolution("SPEA"); res.Source != "ft" || res.Xid != "990931048" {
		t.Errorf("SPEA resolution: %+v", res)
	}
}

// TestFundTickersCSVIntegrity guarantees that every line of the embedded
// list is actually loaded: an ISIN with an invalid check digit is silently
// skipped by the parser, and this test would turn that into a failure.
func TestFundTickersCSVIntegrity(t *testing.T) {
	m := map[string]string{}
	lineCount := 0
	for _, line := range strings.Split(fundTickersCSV, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lineCount++
		parts := strings.Split(line, ";")
		if len(parts) != 3 {
			t.Errorf("malformed line: %q", line)
			continue
		}
		if !IsISIN(parts[0]) {
			t.Errorf("invalid ISIN (check digit?): %s (%s)", parts[0], parts[1])
		}
		tickers := strings.Fields(parts[2])
		if len(tickers) == 0 {
			t.Errorf("no tickers for %s", parts[0])
		}
		for _, tk := range tickers {
			tk = strings.ToUpper(tk)
			if prev, dup := m[tk]; dup && prev != parts[0] {
				t.Errorf("ticker %s mapped to two ISINs: %s and %s", tk, prev, parts[0])
			}
			m[tk] = parts[0]
			if got, ok := FundISIN(tk); !ok || got != parts[0] {
				t.Errorf("FundISIN(%s) = %q, %v; want %s", tk, got, ok, parts[0])
			}
		}
	}
	if lineCount < 75 {
		t.Errorf("the embedded list looks truncated: %d lines", lineCount)
	}
}

func TestUCITSFlag(t *testing.T) {
	cases := map[string]bool{
		"IE00B4L5Y983": true,  // iShares Core MSCI World UCITS
		"NTSX":         true,  // preference: the UCITS, not the US twin
		"LU0171310443": true,  // BGF (UCITS SICAV)
		"FR0011147594": true,  // Omnibond (mutual fund)
		"VOO":          false, // US ETF
		"NTSX-US":      false,
		"IE00B579F325": false, // gold ETC (a debt security, not UCITS)
		"XAUUSD":       false,
		"GG00BQBFY362": false, // closed-end Guernsey
	}
	for id, want := range cases {
		got, known := UCITSFlag(id)
		if !known || got != want {
			t.Errorf("UCITSFlag(%s) = %v/%v, want %v", id, got, known, want)
		}
	}
	if _, known := UCITSFlag("UNKNOWN-TICKER"); known {
		t.Error("an asset outside the catalog must be undetermined")
	}
}

func TestWarmupIDsAreCanonicalAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, id := range WarmupIDs() {
		if seen[id] {
			t.Errorf("duplicate in WarmupIDs: %s", id)
		}
		seen[id] = true
		if id != CanonicalID(id) {
			t.Errorf("%s is not canonical", id)
		}
	}
	for _, want := range []string{"VOO", "XAUUSD", "CL=F", "IE000KF370H3", "LU0319687124"} {
		if !seen[want] {
			t.Errorf("WarmupIDs should contain %s", want)
		}
	}
}

// TestWarmupIDsAllPinned guarantees that every bundled asset resolves
// statically through the catalog, without depending on search engines.
func TestWarmupIDsAllPinned(t *testing.T) {
	withFees := 0
	for _, id := range WarmupIDs() {
		res, ok := catalogResolution(id)
		if !ok {
			t.Errorf("%s is not pinned in the catalog", id)
			continue
		}
		switch res.Source {
		case "yahoo", "stooq", "morningstar":
			if res.Symbol == "" {
				t.Errorf("%s: missing symbol (source %s)", id, res.Source)
			}
		case "ft":
			if res.Xid == "" {
				t.Errorf("%s: missing FT xid", id)
			}
		default:
			t.Errorf("%s: unexpected source %q", id, res.Source)
		}
		if e, found := catalogByID()[id]; found && e.Fees > 0 {
			withFees++
		}
	}
	if withFees < 80 {
		t.Errorf("too few assets with a pinned TER: %d", withFees)
	}
}
