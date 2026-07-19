package marketdata

import (
	"sort"
	"strings"
	"sync"
)

// canonicalIndex maps every accepted identifier (catalog IDs, ISINs,
// per-entry aliases and the embedded fund tickers) to the catalog entry ID
// (or the ISIN itself for funds outside the catalog). Quote symbols are
// deliberately NOT indexed: they may collide with user-facing identifiers
// (e.g. the US-listed NTSX vs the NTSX UCITS share class).
type indexResult struct {
	byKey     map[string]string
	conflicts []string
}

var canonicalIndex = sync.OnceValue(func() indexResult {
	r := indexResult{byKey: map[string]string{}}
	set := func(key, target string) {
		key = strings.ToUpper(strings.TrimSpace(key))
		if key == "" {
			return
		}
		if prev, ok := r.byKey[key]; ok {
			if prev != target {
				r.conflicts = append(r.conflicts, key+": "+prev+" vs "+target)
			}
			return // first one in (priority order below) wins
		}
		r.byKey[key] = target
	}
	// Decreasing priority: entry ID, ISIN, aliases, embedded tickers.
	for _, e := range catalog {
		set(e.ID, e.ID)
	}
	for _, e := range catalog {
		if e.ISIN != "" {
			set(e.ISIN, e.ID)
		}
	}
	for _, e := range catalog {
		for _, a := range e.Aliases {
			set(a, e.ID)
		}
	}
	tickers := fundISINByTicker()
	keys := make([]string, 0, len(tickers))
	for t := range tickers {
		keys = append(keys, t)
	}
	sort.Strings(keys) // deterministic order for conflict detection
	for _, t := range keys {
		isin := tickers[t]
		target := isin
		if canon, ok := r.byKey[isin]; ok {
			target = canon
		}
		set(t, target)
	}
	return r
})

// SplitSim implements the "SIM" suffix convention: a bare identifier means
// real quotes only (starting at the asset's actual inception), while
// "<id>SIM" (DBMFSIM, VOOSIM, WINTON-TREND-EQUITYSIM…) additionally opts
// into simulated-history extension for the uncovered period.
func SplitSim(id string) (base string, sim bool) {
	u := strings.ToUpper(strings.TrimSpace(id))
	if len(u) > 3 && strings.HasSuffix(u, "SIM") {
		return u[:len(u)-3], true
	}
	return u, false
}

// CanonicalID normalizes an identifier (trim, uppercase) and maps catalog
// aliases, ISINs and embedded fund tickers to the canonical catalog ID.
func CanonicalID(id string) string {
	u := strings.ToUpper(strings.TrimSpace(id))
	if c, ok := canonicalIndex().byKey[u]; ok {
		return c
	}
	return u
}

// KnownLocal reports whether an identifier resolves locally, without any
// network lookup: a catalog ID, a catalog ISIN, a catalog alias or an
// embedded European fund ticker (exactly the set CanonicalID maps), with
// an optional SIM suffix. Quote symbols (^GSPC, ^IRX, ...) are not local.
// The web composer uses it to reject identifiers that would otherwise
// trigger a network search on behalf of an anonymous visitor.
func KnownLocal(id string) bool {
	base, _ := SplitSim(id)
	_, ok := canonicalIndex().byKey[base]
	return ok
}
