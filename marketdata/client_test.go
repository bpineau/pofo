package marketdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func chartJSON(symbol string, days []time.Time, closes []float64) string {
	ts, cl := "", ""
	for i := range days {
		if i > 0 {
			ts += ","
			cl += ","
		}
		// 14:30 UTC, comme une clôture américaine.
		ts += fmt.Sprint(days[i].Add(14*time.Hour + 30*time.Minute).Unix())
		cl += fmt.Sprint(closes[i])
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":"USD","symbol":%q,"longName":"Test Fund %s"},"timestamp":[%s],"indicators":{"quote":[{"close":[%s]}],"adjclose":[{"adjclose":[%s]}]}}],"error":null}}`,
		symbol, symbol, ts, cl, cl)
}

func testDays(n int) []time.Time {
	out := make([]time.Time, n)
	for i := range out {
		out[i] = time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
	}
	return out
}

func newTestClient(t *testing.T, dir string, mux *http.ServeMux) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(mux)
	c := NewClient(dir)
	stubAllBases(c, ts.URL)
	c.retryDelay = time.Millisecond
	return c, ts
}

// stubAllBases points every data source at the test server so that no test
// can ever reach the real APIs.
func stubAllBases(c *Client, base string) {
	c.ChartBase, c.SearchBase, c.StooqBase = base, base, base
	c.FTBase, c.BoursoramaBase, c.MorningstarBase = base, base, base
	c.JustETFBase = base
	c.retryDelay = time.Millisecond
}

func TestHistoryFetchParseAndCache(t *testing.T) {
	days := testDays(3)
	closes := []float64{100, 101.5, 99}
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		requests++
		fmt.Fprint(w, chartJSON("VOO", days, closes))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.History("VOO", from)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "Test Fund VOO" || s.Currency != "USD" || len(s.Points) != 3 {
		t.Fatalf("série mal lue: %+v", s)
	}
	if !s.Points[0].Date.Equal(days[0]) {
		t.Errorf("date non normalisée à minuit UTC: %v", s.Points[0].Date)
	}
	if s.Points[2].Close != 99 {
		t.Errorf("clôture: %v", s.Points[2].Close)
	}

	// Un second client (sans mémo) pointant vers un serveur mort doit servir
	// la même série depuis le cache disque, sans aucune requête réseau.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	s2, err := c2.History("VOO", from)
	if err != nil {
		t.Fatalf("le cache aurait dû répondre: %v", err)
	}
	if len(s2.Points) != 3 || s2.Points[1].Close != 101.5 {
		t.Fatalf("cache corrompu: %+v", s2.Points)
	}
	if requests != 1 {
		t.Errorf("1 seule requête attendue, comptées: %d", requests)
	}
}

func TestHistoryCacheExpiry(t *testing.T) {
	days := testDays(2)
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SPY", func(w http.ResponseWriter, r *http.Request) {
		requests++
		fmt.Fprint(w, chartJSON("SPY", days, []float64{10, 11}))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	if _, err := c.History("SPY", from); err != nil {
		t.Fatal(err)
	}
	// Avec MaxAge négatif, le cache est toujours périmé: nouvelle requête.
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	c2.MaxAge = -time.Second
	if _, err := c2.History("SPY", from); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Errorf("2 requêtes attendues, comptées: %d", requests)
	}
}

func TestHistoryStaleCacheFallback(t *testing.T) {
	days := testDays(3)
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SPY", func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests > 1 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, chartJSON("SPY", days, []float64{10, 11, 12}))
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	if _, err := c.History("SPY", from); err != nil {
		t.Fatal(err)
	}
	// Cache périmé + refresh en échec: la donnée périmée doit être servie
	// avec un avertissement, jamais perdue.
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	c2.MaxAge = -time.Second
	warned := false
	c2.Logf = func(format string, args ...any) {
		if strings.Contains(fmt.Sprintf(format, args...), "rafraîchissement de SPY impossible") {
			warned = true
		}
	}
	s, err := c2.History("SPY", from)
	if err != nil {
		t.Fatalf("le cache périmé aurait dû être servi: %v", err)
	}
	if len(s.Points) != 3 || s.Points[2].Close != 12 {
		t.Fatalf("données périmées altérées: %+v", s.Points)
	}
	if !warned {
		t.Error("un avertissement stderr était attendu")
	}
}

func TestFetchISINViaYahoo(t *testing.T) {
	// 100 points: assez profond pour que la résolution soit jugée fiable.
	days := testDays(100)
	closes := make([]float64, len(days))
	for i := range closes {
		closes[i] = 70 + float64(i)
	}
	searches := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		searches++
		if q := r.URL.Query().Get("q"); q != "FR0000120271" {
			t.Errorf("requête de recherche inattendue: %q", q)
		}
		fmt.Fprint(w, `{"quotes":[{"symbol":"IWDA.AS","longname":"iShares Core MSCI World UCITS ETF","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/IWDA.AS", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("IWDA.AS", days, closes))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.Fetch("FR0000120271", from)
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "IWDA.AS" || s.Source != "yahoo" || len(s.Points) != 100 {
		t.Fatalf("série mal résolue: %+v", s)
	}
	// Résolution et historique sont en cache disque: un nouveau client
	// pointant vers un serveur mort doit fonctionner sans réseau.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("FR0000120271", from); err != nil || s2.Symbol != "IWDA.AS" {
		t.Errorf("résolution depuis le cache: %+v, %v", s2, err)
	}
	if searches != 1 {
		t.Errorf("1 seule recherche attendue, comptées: %d", searches)
	}
}

func TestFetchISINFallsBackToFT(t *testing.T) {
	days := testDays(80)
	var ftDates, ftCloses []string
	for i, d := range days {
		ftDates = append(ftDates, fmt.Sprintf("%q", d.Format("2006-01-02T15:04:05")))
		ftCloses = append(ftCloses, fmt.Sprintf("%g", 10.5+float64(i)))
	}
	mux := http.NewServeMux()
	// Yahoo ne connaît pas ce fonds.
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[]}`)
	})
	mux.HandleFunc("/data/searchapi/searchsecurities", func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query().Get("query"); q != "DE0007164600" {
			t.Errorf("recherche FT inattendue: %q", q)
		}
		fmt.Fprint(w, `{"data":{"security":[{"name":"BGF World Technology A2","symbol":"DE0007164600:EUR","xid":"28295854","isPrimary":true}]}}`)
	})
	mux.HandleFunc("/data/chartapi/series", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("méthode FT inattendue: %s", r.Method)
		}
		var req struct {
			Elements []struct {
				Symbol string `json:"Symbol"`
			} `json:"elements"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Elements) != 1 || req.Elements[0].Symbol != "28295854" {
			t.Errorf("corps de requête FT inattendu: %+v (%v)", req, err)
		}
		fmt.Fprintf(w, `{"Dates":[%s],"Elements":[{"Currency":"EUR","ComponentSeries":[{"Type":"Open","Values":[%s]},{"Type":"Close","Values":[%s]}]}]}`,
			strings.Join(ftDates, ","), strings.Join(ftCloses, ","), strings.Join(ftCloses, ","))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.Fetch("DE0007164600", from)
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "DE0007164600" || s.Source != "ft" || s.Currency != "EUR" || s.Name != "BGF World Technology A2" {
		t.Fatalf("série FT mal lue: %+v", s)
	}
	if len(s.Points) != 80 || s.Points[0].Close != 10.5 || !s.Points[0].Date.Equal(days[0]) {
		t.Fatalf("points FT: %d points, premier %+v", len(s.Points), s.Points[0])
	}
	// La résolution FT est en cache: plus aucune requête nécessaire.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("DE0007164600", from); err != nil || s2.Source != "ft" || len(s2.Points) != 80 {
		t.Errorf("rechargement FT depuis le cache: %+v, %v", s2, err)
	}
}

func TestFetchISINPicksDeepestCandidate(t *testing.T) {
	// Le premier candidat (cotation de bourse moribonde) n'a qu'un point;
	// le second (entrée « fonds » Morningstar) a un historique profond et
	// doit être retenu même s'il arrive après dans le classement initial.
	deep := testDays(90)
	deepCloses := make([]float64, len(deep))
	for i := range deepCloses {
		deepCloses[i] = 100 + float64(i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[
			{"symbol":"ERDV.F","longname":"BGF World Healthscience (Frankfurt)","quoteType":"EQUITY"},
			{"symbol":"0P0000VHO6.F","longname":"BGF World Healthscience A2","quoteType":"MUTUALFUND"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/ERDV.F", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("ERDV.F", testDays(1), []float64{42}))
	})
	mux.HandleFunc("/v8/finance/chart/0P0000VHO6.F", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("0P0000VHO6.F", deep, deepCloses))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch("LU0171307068", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "0P0000VHO6.F" || len(s.Points) != 90 {
		t.Fatalf("le candidat le plus profond devait gagner: %s (%d points)", s.Symbol, len(s.Points))
	}
}

func TestFetchISINViaBoursoramaMorningstar(t *testing.T) {
	// Yahoo et FT ne trouvent rien par ISIN; Boursorama fournit l'identifiant
	// Morningstar, et l'API timeseries Morningstar porte l'historique.
	deep := testDays(70)
	var rows []string
	for i, d := range deep {
		rows = append(rows, fmt.Sprintf("[%d, %g]", d.UnixMilli(), 8+float64(i)*0.1))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[]}`)
	})
	mux.HandleFunc("/data/searchapi/searchsecurities", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":{"security":[]}}`)
	})
	mux.HandleFunc("/recherche/ajax", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Error("en-tête XMLHttpRequest attendu")
		}
		fmt.Fprint(w, `<a href="/bourse/opcvm/cours/0P0000VHO6/" class="search__list-link"><span class="search__item-title">BGF World Healthscience A2 </span></a>`)
	})
	mux.HandleFunc("/api/rest.svc/timeseries_price/"+morningstarToken, func(w http.ResponseWriter, r *http.Request) {
		if id := r.URL.Query().Get("id"); id != "0P0000VHO6" {
			t.Errorf("identifiant morningstar inattendu: %q", id)
		}
		fmt.Fprintf(w, "[%s]", strings.Join(rows, ","))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.Fetch("US0378331005", from)
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "US0378331005" || s.Source != "morningstar" || len(s.Points) != 70 {
		t.Fatalf("pont Boursorama→Morningstar: %+v (%d points)", s, len(s.Points))
	}
	if s.Name != "BGF World Healthscience A2" {
		t.Errorf("nom extrait du HTML: %q", s.Name)
	}
	if !s.Points[0].Date.Equal(deep[0]) || s.Points[0].Close != 8 {
		t.Errorf("premier point: %+v", s.Points[0])
	}
	// Résolution et historique en cache: rejouable sans réseau.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("US0378331005", from); err != nil || s2.Source != "morningstar" || len(s2.Points) != 70 {
		t.Errorf("rechargement Morningstar depuis le cache: %v", err)
	}
}

func TestFetchTickerFallsBackToSearch(t *testing.T) {
	// NTSG n'existe pas tel quel sur Yahoo (404): la résolution par recherche
	// doit trouver la cotation européenne du même ticker (NTSG.MI), en la
	// préférant à un fonds homonyme plus profond mais d'un autre ticker.
	days := testDays(100)
	closes := make([]float64, len(days))
	for i := range closes {
		closes[i] = 30 + float64(i)*0.1
	}
	searches := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/QQZZ", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		searches++
		fmt.Fprint(w, `{"quotes":[
			{"symbol":"OTHER.F","longname":"Fonds sans rapport","quoteType":"MUTUALFUND"},
			{"symbol":"QQZZ.MI","longname":"WisdomTree Global Efficient Core UCITS ETF","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/QQZZ.MI", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("QQZZ.MI", days, closes))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.Fetch("qqzz", from)
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "QQZZ.MI" || len(s.Points) != 100 {
		t.Fatalf("résolution du ticker: %s (%d points)", s.Symbol, len(s.Points))
	}
	// La résolution est en cache: rejouable sans réseau.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("QQZZ", from); err != nil || s2.Symbol != "QQZZ.MI" {
		t.Errorf("résolution du ticker depuis le cache: %+v, %v", s2, err)
	}
	if searches != 1 {
		t.Errorf("1 seule recherche attendue, comptées: %d", searches)
	}
}

func TestFetchTickerUppercases(t *testing.T) {
	days := testDays(2)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, []float64{10, 11}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	s, err := c.Fetch("voo", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || s.Symbol != "VOO" {
		t.Errorf("ticker en minuscules: %+v, %v", s, err)
	}
}

func TestHistoryFallsBackToStooq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		if s := r.URL.Query().Get("s"); s != "xyz.us" {
			t.Errorf("symbole stooq inattendu: %q", s)
		}
		fmt.Fprint(w, "Date,Open,High,Low,Close,Volume\n2020-01-06,1,1,1,42.5,100\n2020-01-07,1,1,1,43,100\n")
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()

	s, err := c.History("XYZ", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "stooq" || len(s.Points) != 2 || s.Points[0].Close != 42.5 {
		t.Fatalf("fallback stooq: %+v", s)
	}
}

func TestFetchTickerPrefersFundOverNamesakeStock(t *testing.T) {
	// Une action homonyme à l'historique profond (Saipem sous SPEA.MU) ne
	// doit pas voler la résolution d'un jeune ETF du même ticker (SPEA.PA).
	deepStock := testDays(500)
	stockCloses := make([]float64, len(deepStock))
	for i := range stockCloses {
		stockCloses[i] = 5 + float64(i)*0.01
	}
	youngETF := testDays(100)
	etfCloses := make([]float64, len(youngETF))
	for i := range etfCloses {
		etfCloses[i] = 10 + float64(i)*0.01
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SPEA", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[
			{"symbol":"SPEA.MU","longname":"Saipem SpA","quoteType":"EQUITY"},
			{"symbol":"SPEA.PA","longname":"iShares S&P 500 Swap PEA UCITS ETF","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/SPEA.MU", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("SPEA.MU", deepStock, stockCloses))
	})
	mux.HandleFunc("/v8/finance/chart/SPEA.PA", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("SPEA.PA", youngETF, etfCloses))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch("SPEA", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "SPEA.PA" {
		t.Fatalf("l'ETF du même ticker devait gagner, obtenu %s (%s)", s.Symbol, s.Name)
	}
}
