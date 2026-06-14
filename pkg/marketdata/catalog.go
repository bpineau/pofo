package marketdata

import (
	"bufio"
	_ "embed"
	"sort"
	"strings"
	"sync"

	"github.com/bpineau/portfodor/datasets"
)

// catalog lists the assets bundled with portfodor, the typed view of the
// embedded datasets/assetmeta/assets.json. The catalog IS the bundle:
// everything here resolves deterministically (no search APIs), has its TER
// pinned when published, and is fully cached by a single --warmup. Entries
// are addressable by ID, ISIN, Aliases and the tickers of the embedded fund
// list — but never by quote Symbol, which may collide (e.g. US-listed NTSX
// vs the NTSX UCITS).
var catalog = datasets.Catalog()

var catalogByID = sync.OnceValue(func() map[string]datasets.Asset {
	m := make(map[string]datasets.Asset, 2*len(catalog))
	for _, e := range catalog {
		m[e.ID] = e
		if e.ISIN != "" {
			m[e.ISIN] = e
		}
	}
	return m
})

// Lookup returns the full catalog metadata for an identifier (ticker, alias,
// ISIN or canonical id), which it resolves to its canonical id first. The
// second result is false when the identifier is not in the bundled catalog.
func Lookup(id string) (datasets.Asset, bool) {
	a, ok := catalogByID()[CanonicalID(id)]
	return a, ok
}

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
