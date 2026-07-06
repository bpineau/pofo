// Command gen-macropanel builds the bundled multi-country monthly macro panel
// (pkg/datasets/macropanel/oecd-monthly.csv) from the OECD Main Economic
// Indicators, served through the free, key-less DBnomics mirror. It runs at
// data-generation time only (network); the pofo binary never fetches OECD, it
// embeds the committed CSV.
//
// The panel carries, per country and month, the macro drivers behind regime and
// factor analysis: industrial production (a growth proxy), consumer prices
// (inflation), the long and short interest rates (the monetary quadrant), and a
// share-price index. Growth/inflation breadth across these countries (the share
// accelerating) is the "world point" of the growth x inflation model; the rates
// drive the defensive rotation.
//
// Series (OECD MEI codes):
//
//	ip         PRINTO01.IXOBSA  industrial production, index, seasonally adj.
//	cpi        CPALTT01.IXOB    consumer prices, all items, index
//	shortrate  IR3TIB01.ST      3-month interbank rate, per cent (IRSTCI01 fallback)
//	longrate   IRLTLT01.ST      long-term government bond yield, per cent
//	shareprice SPASTT01.IXOB    share prices, index
//
// Usage:
//
//	gen-macropanel [-base URL] [-o path] [-dry]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultBase = "https://api.db.nomics.world/v22"

// countries covered by the panel: a broad OECD + large-emerging set, wide enough
// that growth/inflation breadth is a meaningful proxy for the world.
var countries = []string{
	"USA", "JPN", "DEU", "FRA", "GBR", "ITA", "CAN", "AUS", "ESP", "NLD",
	"BEL", "SWE", "CHE", "AUT", "DNK", "FIN", "NOR", "PRT", "GRC", "IRL",
	"POL", "CZE", "HUN", "KOR", "MEX", "TUR", "NZL", "ZAF", "BRA", "IND",
}

// the OECD MEI series pulled per country, in output-column order.
var series = []struct{ col, code string }{
	{"ip", "PRINTO01.IXOBSA"},
	{"cpi", "CPALTT01.IXOB"},
	{"shortrate", "IR3TIB01.ST"},
	{"longrate", "IRLTLT01.ST"},
	{"shareprice", "SPASTT01.IXOB"},
}

// shortFallback is spliced under shortrate where the 3-month rate is missing.
const shortFallback = "IRSTCI01.ST"

func main() {
	base := flag.String("base", defaultBase, "DBnomics API base URL")
	out := flag.String("o", "pkg/datasets/macropanel/oecd-monthly.csv", "output CSV path")
	dry := flag.Bool("dry", false, "print coverage without writing")
	flag.Parse()

	// panel[iso][month][col] = value.
	panel := map[string]map[string]map[string]float64{}
	var mu sync.Mutex
	put := func(iso, col string, s map[string]float64) {
		mu.Lock()
		defer mu.Unlock()
		if panel[iso] == nil {
			panel[iso] = map[string]map[string]float64{}
		}
		for m, v := range s {
			if panel[iso][m] == nil {
				panel[iso][m] = map[string]float64{}
			}
			if _, seen := panel[iso][m][col]; !seen { // primary wins over fallback
				panel[iso][m][col] = v
			}
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	fetchInto := func(iso, col, code string) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()
		s, err := fetch(*base, iso, code)
		if err != nil || s == nil {
			return
		}
		put(iso, col, s)
	}
	for _, iso := range countries {
		for _, s := range series {
			wg.Add(1)
			go fetchInto(iso, s.col, s.code)
		}
		wg.Add(1)
		go fetchInto(iso, "shortrate", shortFallback) // fallback, primary wins
	}
	wg.Wait()

	recs := flatten(panel)
	logCoverage(recs)
	if *dry {
		return
	}
	if err := writeCSV(*out, recs); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d rows, %d countries)", *out, len(recs), len(panel))
}

// fetch downloads one OECD MEI monthly series from DBnomics, month ("YYYY-MM")
// keyed to its float value.
func fetch(base, iso, code string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/series/OECD/MEI/%s.%s.M?observations=1", base, iso, code)
	cl := &http.Client{Timeout: 60 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "pofo-gen-macropanel/1.0")
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s.%s: HTTP %d", iso, code, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	var doc struct {
		Series struct {
			Docs []struct {
				Period []string          `json:"period"`
				Value  []json.RawMessage `json:"value"`
			} `json:"docs"`
		} `json:"series"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	if len(doc.Series.Docs) == 0 {
		return nil, nil
	}
	d := doc.Series.Docs[0]
	out := make(map[string]float64, len(d.Period))
	for i, p := range d.Period {
		if i >= len(d.Value) {
			break
		}
		var v float64
		if json.Unmarshal(d.Value[i], &v) != nil { // "NA"
			continue
		}
		if len(p) != 7 || p[4] != '-' { // keep only "YYYY-MM"
			continue
		}
		out[p] = v
	}
	return out, nil
}

// record is one country-month row of the panel.
type record struct {
	ISO, Month string
	Val        map[string]float64
}

func flatten(panel map[string]map[string]map[string]float64) []record {
	var out []record
	for iso, byMonth := range panel {
		for m, cols := range byMonth {
			out = append(out, record{iso, m, cols})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ISO != out[j].ISO {
			return out[i].ISO < out[j].ISO
		}
		return out[i].Month < out[j].Month
	})
	return out
}

func writeCSV(path string, recs []record) error {
	var b strings.Builder
	b.WriteString("# Multi-country monthly macro panel: OECD Main Economic Indicators.\n")
	b.WriteString("# Columns: iso,date(YYYY-MM),ip,cpi,shortrate,longrate,shareprice\n")
	b.WriteString("# ip/cpi/shareprice are index levels; shortrate/longrate are per-cent yields.\n")
	b.WriteString("# Source: OECD MEI via DBnomics (https://db.nomics.world, provider OECD/MEI).\n")
	b.WriteString("# shortrate = 3-month interbank (IR3TIB01), immediate rate (IRSTCI01) where absent.\n")
	b.WriteString("# Regenerate: make macropanel\n")
	b.WriteString("iso,date,ip,cpi,shortrate,longrate,shareprice\n")
	num := func(cols map[string]float64, k string) string {
		v, ok := cols[k]
		if !ok {
			return ""
		}
		return strconv.FormatFloat(v, 'f', 4, 64)
	}
	for _, r := range recs {
		fmt.Fprintf(&b, "%s,%s,%s,%s,%s,%s,%s\n", r.ISO, r.Month,
			num(r.Val, "ip"), num(r.Val, "cpi"), num(r.Val, "shortrate"),
			num(r.Val, "longrate"), num(r.Val, "shareprice"))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func logCoverage(recs []record) {
	byISO := map[string]int{}
	colCount := map[string]int{}
	for _, r := range recs {
		byISO[r.ISO]++
		for k := range r.Val {
			colCount[k]++
		}
	}
	log.Printf("panel: %d country-month rows, %d countries", len(recs), len(byISO))
	for _, s := range series {
		log.Printf("  %-10s %d rows", s.col, colCount[s.col])
	}
}
