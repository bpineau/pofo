package marketdata

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCatalogPinsResolutionWithoutSearch(t *testing.T) {
	// Un actif du catalogue ne déclenche aucune recherche réseau: seule la
	// source épinglée est interrogée.
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
		t.Error("le catalogue doit court-circuiter la recherche")
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch("VOO", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "VOO" || len(s.Points) != 70 {
		t.Fatalf("résolution catalogue: %+v", s)
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
			t.Errorf("FundISIN(%q) = %q, %v — attendu %q", ticker, got, ok, want)
		}
	}
	if _, ok := FundISIN("AAPL"); ok {
		t.Error("AAPL ne doit pas figurer dans la liste des fonds")
	}
	// Les ETF PEA iShares sont épinglés durablement (ticker et ISIN).
	for _, q := range []string{"SPEA", "IE000DQLYVB9", "WPEA", "IE0002XZSHO1"} {
		if _, ok := catalogResolution(q); !ok {
			t.Errorf("%s devrait être épinglé au catalogue", q)
		}
	}
	if res, _ := catalogResolution("SPEA"); res.Source != "ft" || res.Xid != "990931048" {
		t.Errorf("résolution SPEA: %+v", res)
	}
}

// TestFundTickersCSVIntegrity garantit que chaque ligne de la liste
// embarquée est réellement chargée: un ISIN à clé de contrôle invalide est
// silencieusement ignoré par le parseur, ce test le transformerait en échec.
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
			t.Errorf("ligne mal formée: %q", line)
			continue
		}
		if !IsISIN(parts[0]) {
			t.Errorf("ISIN invalide (clé de contrôle ?): %s (%s)", parts[0], parts[1])
		}
		tickers := strings.Fields(parts[2])
		if len(tickers) == 0 {
			t.Errorf("aucun ticker pour %s", parts[0])
		}
		for _, tk := range tickers {
			tk = strings.ToUpper(tk)
			if prev, dup := m[tk]; dup && prev != parts[0] {
				t.Errorf("ticker %s mappé sur deux ISIN: %s et %s", tk, prev, parts[0])
			}
			m[tk] = parts[0]
			if got, ok := FundISIN(tk); !ok || got != parts[0] {
				t.Errorf("FundISIN(%s) = %q, %v — attendu %s", tk, got, ok, parts[0])
			}
		}
	}
	if lineCount < 75 {
		t.Errorf("la liste embarquée semble tronquée: %d lignes", lineCount)
	}
}

func TestUCITSFlag(t *testing.T) {
	cases := map[string]bool{
		"IE00B4L5Y983": true,  // iShares Core MSCI World UCITS
		"NTSX":         true,  // préférence: l'UCITS, pas le jumeau US
		"LU0171310443": true,  // BGF (SICAV UCITS)
		"FR0011147594": true,  // Omnibond (OPCVM)
		"VOO":          false, // ETF US
		"NTSX-US":      false,
		"IE00B579F325": false, // ETC or (titre de dette, pas UCITS)
		"XAUUSD":       false,
		"GG00BQBFY362": false, // closed-end Guernsey
	}
	for id, want := range cases {
		got, known := UCITSFlag(id)
		if !known || got != want {
			t.Errorf("UCITSFlag(%s) = %v/%v, attendu %v", id, got, known, want)
		}
	}
	if _, known := UCITSFlag("TICKER-INCONNU"); known {
		t.Error("un actif hors catalogue doit être indéterminé")
	}
}

func TestWarmupIDsAreCanonicalAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, id := range WarmupIDs() {
		if seen[id] {
			t.Errorf("doublon dans WarmupIDs: %s", id)
		}
		seen[id] = true
		if id != CanonicalID(id) {
			t.Errorf("%s n'est pas canonique", id)
		}
	}
	for _, want := range []string{"VOO", "XAUUSD", "CL=F", "IE000KF370H3", "LU0319687124"} {
		if !seen[want] {
			t.Errorf("WarmupIDs devrait contenir %s", want)
		}
	}
}

// TestWarmupIDsAllPinned garantit que chaque actif du bundle se résout
// statiquement par le catalogue, sans dépendre des moteurs de recherche.
func TestWarmupIDsAllPinned(t *testing.T) {
	withFees := 0
	for _, id := range WarmupIDs() {
		res, ok := catalogResolution(id)
		if !ok {
			t.Errorf("%s n'est pas épinglé au catalogue", id)
			continue
		}
		switch res.Source {
		case "yahoo", "stooq", "morningstar":
			if res.Symbol == "" {
				t.Errorf("%s: symbole manquant (source %s)", id, res.Source)
			}
		case "ft":
			if res.Xid == "" {
				t.Errorf("%s: xid FT manquant", id)
			}
		default:
			t.Errorf("%s: source inattendue %q", id, res.Source)
		}
		if e, found := catalogByID()[id]; found && e.Fees > 0 {
			withFees++
		}
	}
	if withFees < 80 {
		t.Errorf("trop peu d'actifs avec TER épinglé: %d", withFees)
	}
}
