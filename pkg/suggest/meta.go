// Package suggest recommends catalog assets to add to a portfolio so it
// covers the market regimes it is missing, and flags redundant holdings.
//
// It is the structure-first half of the optimizer: candidates are screened
// by what they ARE (asset class, strategy → macro-regime coverage and
// statistical diversification) before any return-based ranking, and the
// return ranking is validated out-of-sample (walk-forward) so a suggestion
// reflects a consistent benefit rather than one lucky period. Conventions
// match pkg/metrics: simple daily returns, 252 trading days per year.
package suggest

import (
	"encoding/json"
	"io"

	"github.com/bpineau/portfodor/datasets"
)

// Meta is the factual metadata of one catalog asset. It is an alias for the
// canonical datasets.Asset (a full row of datasets/assetmeta/assets.json):
// load the bundled catalog directly with datasets.Catalog, or decode any
// reader of the same JSON with LoadMeta.
type Meta = datasets.Asset

// LoadMeta decodes a JSON array of assets (e.g. datasets.AssetMeta) into a
// map keyed by both the canonical id and the ISIN, so either resolves.
// Resolve a ticker/alias to its id with marketdata.CanonicalID before
// indexing the map.
func LoadMeta(r io.Reader) (map[string]Meta, error) {
	var list []Meta
	if err := json.NewDecoder(r).Decode(&list); err != nil {
		return nil, err
	}
	m := make(map[string]Meta, 2*len(list))
	for _, e := range list {
		m[e.ID] = e
		if e.ISIN != "" {
			m[e.ISIN] = e
		}
	}
	return m, nil
}
