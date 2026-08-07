package marketdata

import "sort"

// LocalAsset is one composable asset of the local catalog: the canonical
// entry plus every alternate identifier that resolves to it without a
// network lookup. It is the client-facing view of the KnownLocal set,
// served by the web app as /catalog.json for autocomplete and inline
// validation; the server-side gates remain the authority.
type LocalAsset struct {
	ID    string   `json:"id"`            // canonical id (the CanonicalID target)
	Name  string   `json:"name"`          // display name; "" for bare fund ISINs outside the catalog
	Class string   `json:"class"`         // asset class; "" when unknown
	Alt   []string `json:"alt,omitempty"` // ISIN, aliases and embedded fund tickers mapping to ID
}

// LocalCatalog lists every asset KnownLocal accepts (the SIM suffix
// aside), one entry per canonical id, sorted by ID with sorted alternates,
// so its serialized form is byte-stable.
func LocalCatalog() []LocalAsset {
	byID := map[string]*LocalAsset{}
	for key, target := range canonicalIndex().byKey {
		e, ok := byID[target]
		if !ok {
			e = &LocalAsset{ID: target}
			if a, inCatalog := catalogByID()[target]; inCatalog {
				e.Name, e.Class = a.Name, a.AssetClass
			}
			byID[target] = e
		}
		if key != target {
			e.Alt = append(e.Alt, key)
		}
	}
	out := make([]LocalAsset, 0, len(byID))
	for _, e := range byID {
		sort.Strings(e.Alt)
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
