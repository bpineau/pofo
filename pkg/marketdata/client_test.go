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
		// 14:30 UTC, like a US close.
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
	c.EurostatBase = base
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
		t.Fatalf("series misread: %+v", s)
	}
	if !s.Points[0].Date.Equal(days[0]) {
		t.Errorf("date not normalized to midnight UTC: %v", s.Points[0].Date)
	}
	if s.Points[2].Close != 99 {
		t.Errorf("close: %v", s.Points[2].Close)
	}

	// A second client (without the memo) pointing at a dead server must
	// serve the same series from the disk cache, with no network request.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	s2, err := c2.History("VOO", from)
	if err != nil {
		t.Fatalf("the cache should have answered: %v", err)
	}
	if len(s2.Points) != 3 || s2.Points[1].Close != 101.5 {
		t.Fatalf("corrupted cache: %+v", s2.Points)
	}
	if requests != 1 {
		t.Errorf("expected exactly 1 request, counted: %d", requests)
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
	// With a negative MaxAge the cache is always stale: a new request.
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	c2.MaxAge = -time.Second
	if _, err := c2.History("SPY", from); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Errorf("expected 2 requests, counted: %d", requests)
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
	// Stale cache + failed refresh: the stale data must be served with a
	// warning, never lost.
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	c2.MaxAge = -time.Second
	warned := false
	c2.Logf = func(format string, args ...any) {
		if strings.Contains(fmt.Sprintf(format, args...), "refreshing SPY failed") {
			warned = true
		}
	}
	s, err := c2.History("SPY", from)
	if err != nil {
		t.Fatalf("the stale cache should have been served: %v", err)
	}
	if len(s.Points) != 3 || s.Points[2].Close != 12 {
		t.Fatalf("stale data altered: %+v", s.Points)
	}
	if !warned {
		t.Error("a stderr warning was expected")
	}
}

func TestFetchISINViaYahoo(t *testing.T) {
	// 100 points: deep enough for the resolution to be deemed reliable.
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
			t.Errorf("unexpected search query: %q", q)
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
		t.Fatalf("series resolved incorrectly: %+v", s)
	}
	// Resolution and history are cached on disk: a new client pointing at
	// a dead server must work without the network.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("FR0000120271", from); err != nil || s2.Symbol != "IWDA.AS" {
		t.Errorf("resolution from the cache: %+v, %v", s2, err)
	}
	if searches != 1 {
		t.Errorf("expected exactly 1 search, counted: %d", searches)
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
	// Yahoo does not know this fund.
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[]}`)
	})
	mux.HandleFunc("/data/searchapi/searchsecurities", func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query().Get("query"); q != "DE0007164600" {
			t.Errorf("unexpected FT search: %q", q)
		}
		fmt.Fprint(w, `{"data":{"security":[{"name":"BGF World Technology A2","symbol":"DE0007164600:EUR","xid":"28295854","isPrimary":true}]}}`)
	})
	mux.HandleFunc("/data/chartapi/series", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected FT method: %s", r.Method)
		}
		var req struct {
			Elements []struct {
				Symbol string `json:"Symbol"`
			} `json:"elements"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Elements) != 1 || req.Elements[0].Symbol != "28295854" {
			t.Errorf("unexpected FT request body: %+v (%v)", req, err)
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
		t.Fatalf("FT series misread: %+v", s)
	}
	if len(s.Points) != 80 || s.Points[0].Close != 10.5 || !s.Points[0].Date.Equal(days[0]) {
		t.Fatalf("FT points: %d points, first %+v", len(s.Points), s.Points[0])
	}
	// The FT resolution is cached: no further requests needed.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("DE0007164600", from); err != nil || s2.Source != "ft" || len(s2.Points) != 80 {
		t.Errorf("FT reload from the cache: %+v, %v", s2, err)
	}
}

func TestFetchISINPicksDeepestCandidate(t *testing.T) {
	// The first candidate (a moribund exchange listing) has only one point;
	// the second (a Morningstar "fund" entry) has a deep history and must
	// be picked even though it comes later in the initial ranking.
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
		t.Fatalf("the deepest candidate should have won: %s (%d points)", s.Symbol, len(s.Points))
	}
}

func TestFetchISINViaBoursoramaMorningstar(t *testing.T) {
	// Yahoo and FT find nothing by ISIN; Boursorama supplies the Morningstar
	// identifier, and the Morningstar timeseries API carries the history.
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
			t.Error("expected XMLHttpRequest header")
		}
		fmt.Fprint(w, `<a href="/bourse/opcvm/cours/0P0000VHO6/" class="search__list-link"><span class="search__item-title">BGF World Healthscience A2 </span></a>`)
	})
	mux.HandleFunc("/api/rest.svc/timeseries_price/"+morningstarToken, func(w http.ResponseWriter, r *http.Request) {
		if id := r.URL.Query().Get("id"); id != "0P0000VHO6" {
			t.Errorf("unexpected morningstar identifier: %q", id)
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
		t.Fatalf("Boursorama→Morningstar bridge: %+v (%d points)", s, len(s.Points))
	}
	if s.Name != "BGF World Healthscience A2" {
		t.Errorf("name extracted from the HTML: %q", s.Name)
	}
	if !s.Points[0].Date.Equal(deep[0]) || s.Points[0].Close != 8 {
		t.Errorf("first point: %+v", s.Points[0])
	}
	// Resolution and history cached: replayable without the network.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("US0378331005", from); err != nil || s2.Source != "morningstar" || len(s2.Points) != 70 {
		t.Errorf("Morningstar reload from the cache: %v", err)
	}
}

func TestFetchTickerFallsBackToSearch(t *testing.T) {
	// NTSG does not exist as such on Yahoo (404): the search-based resolution
	// must find the European listing of the same ticker (NTSG.MI), preferring
	// it over a deeper namesake fund under a different ticker.
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
			{"symbol":"OTHER.F","longname":"Unrelated fund","quoteType":"MUTUALFUND"},
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
		t.Fatalf("ticker resolution: %s (%d points)", s.Symbol, len(s.Points))
	}
	// The resolution is cached: replayable without the network.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch("QQZZ", from); err != nil || s2.Symbol != "QQZZ.MI" {
		t.Errorf("ticker resolution from the cache: %+v, %v", s2, err)
	}
	if searches != 1 {
		t.Errorf("expected exactly 1 search, counted: %d", searches)
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
		t.Errorf("lowercase ticker: %+v, %v", s, err)
	}
}

func TestHistoryFallsBackToStooq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		if s := r.URL.Query().Get("s"); s != "xyz.us" {
			t.Errorf("unexpected stooq symbol: %q", s)
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
		t.Fatalf("stooq fallback: %+v", s)
	}
}

func TestFetchTickerPrefersFundOverNamesakeStock(t *testing.T) {
	// A namesake stock with a deep history (Saipem under SPEA.MU) must not
	// steal the resolution of a young ETF with the same ticker (SPEA.PA).
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
		t.Fatalf("the same-ticker ETF should have won, got %s (%s)", s.Symbol, s.Name)
	}
}
