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
)

// Meta is the factual metadata of one catalog asset, a row of
// datasets/assetmeta/assets.json.
type Meta struct {
	ID           string             `json:"id"`
	ISIN         string             `json:"isin"`
	AssetClass   string             `json:"asset_class"`
	Underlying   string             `json:"underlying"`
	Benchmark    string             `json:"benchmark_index"`
	Strategy     string             `json:"strategy"`
	Geography    map[string]float64 `json:"geography"`
	Sectors      map[string]float64 `json:"sectors"`
	Currency     string             `json:"currency"`
	Distribution string             `json:"distribution"`
	Leverage     float64            `json:"leverage"`
	Notes        string             `json:"notes"`
	Confidence   string             `json:"confidence"`
	Sources      []string           `json:"sources"`
}

// LoadMeta parses the asset-metadata JSON (an array of Meta) into a map
// keyed by both the catalog id and the ISIN, so either resolves.
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
