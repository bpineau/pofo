package marketdata

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/bpineau/portfodor/datasets"
)

// CatalogEntry pins the resolution of a well-known asset: how to fetch its
// quotes without any network search, plus display metadata. Entries are
// loaded from the embedded datasets/assetmeta/assets.json (the single source
// of truth, which also carries each asset's descriptive metadata).
type CatalogEntry struct {
	ID       string   `json:"id"`      // canonical identifier (ticker or ISIN)
	ISIN     string   `json:"isin"`    // informational; may be empty for indices/commodities
	Aliases  []string `json:"aliases"` // alternative identifiers accepted in portfolio files
	UCITS    bool     `json:"ucits"`   // UCITS funds/ETFs (ETCs, US funds, indices… are not)
	Name     string   `json:"name"`
	Source   string   `json:"source"`   // "yahoo", "ft", "morningstar" or "stooq"
	Symbol   string   `json:"symbol"`   // yahoo/stooq symbol or Morningstar id; unused for ft
	Xid      string   `json:"xid"`      // FT internal id; unused otherwise
	Currency string   `json:"currency"` //
	Fees     float64  `json:"fees"`     // published TER, percent per year; 0 = unknown
}

// catalog lists the assets bundled with portfodor, loaded once from the
// embedded datasets/assetmeta/assets.json. The catalog IS the bundle:
// everything here resolves deterministically (no search APIs), has its TER
// pinned when published, and is fully cached by a single --warmup. Entries
// are addressable by ID, ISIN, Aliases and the tickers of the embedded fund
// list — but never by quote Symbol, which may collide (e.g. US-listed NTSX
// vs the NTSX UCITS).
var catalog = loadCatalog()

// loadCatalog parses the embedded asset metadata into catalog entries. The
// descriptive fields (asset_class, geography…) in the JSON are ignored here;
// they are consumed by pkg/suggest.
func loadCatalog() []CatalogEntry {
	var entries []CatalogEntry
	if err := json.Unmarshal(datasets.AssetMeta(), &entries); err != nil {
		panic("marketdata: cannot parse the embedded asset catalog: " + err.Error())
	}
	return entries
}

var catalogByID = sync.OnceValue(func() map[string]CatalogEntry {
	m := make(map[string]CatalogEntry, 2*len(catalog))
	for _, e := range catalog {
		m[e.ID] = e
		if e.ISIN != "" {
			m[e.ISIN] = e
		}
	}
	return m
})

// catalogResolution returns the pinned resolution for a canonical id.
func catalogResolution(id string) (resolution, bool) {
	e, ok := catalogByID()[id]
	if !ok {
		return resolution{}, false
	}
	return resolution{Source: e.Source, Symbol: e.Symbol, Xid: e.Xid, Name: e.Name, Currency: e.Currency}, true
}

// UCITSFlag reports whether a catalogued asset is a UCITS fund; known is
// false when the identifier is not in the catalog.
func UCITSFlag(id string) (ucits, known bool) {
	e, found := catalogByID()[CanonicalID(id)]
	if !found {
		return false, false
	}
	return e.UCITS, true
}

// GuessUCITS extends UCITSFlag to uncatalogued assets with a heuristic on
// the resolved name (UCITS funds advertise it in their share-class names).
func GuessUCITS(id, name string) (ucits, known bool) {
	if u, k := UCITSFlag(id); k {
		return u, true
	}
	if strings.Contains(strings.ToUpper(name), "UCITS") {
		return true, true
	}
	return false, false
}

// LooksDistributing reports whether a fund share-class name suggests a
// distributing class, whose NAV series (FT, Morningstar) excludes the
// dividends it pays out.
func LooksDistributing(name string) bool {
	n := strings.ToLower(name)
	for _, marker := range []string{"dist", "(dis)", " dis ", "(d)", " inc", "dy)"} {
		if strings.Contains(n, marker) {
			return true
		}
	}
	return false
}

// WarmupIDs lists the canonical ID of every catalog entry: the catalog IS
// the bundle of assets whose data is precomputed or one --warmup away.
func WarmupIDs() []string {
	ids := make([]string, 0, len(catalog))
	for _, e := range catalog {
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	return ids
}

//go:embed data/fund_tickers.csv
var fundTickersCSV string

// fundISINByTicker maps exchange tickers of common European funds and ETFs
// to their ISIN, from the embedded list. US tickers are deliberately absent:
// they resolve directly on Yahoo.
var fundISINByTicker = sync.OnceValue(func() map[string]string {
	m := make(map[string]string)
	sc := bufio.NewScanner(strings.NewReader(fundTickersCSV))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 3 || !IsISIN(parts[0]) {
			continue
		}
		for _, t := range strings.Fields(strings.ToUpper(parts[2])) {
			m[t] = parts[0]
		}
	}
	return m
})

// FundISIN maps a European fund/ETF exchange ticker to its ISIN using the
// embedded correspondence list.
func FundISIN(ticker string) (string, bool) {
	isin, ok := fundISINByTicker()[strings.ToUpper(strings.TrimSpace(ticker))]
	return isin, ok
}
