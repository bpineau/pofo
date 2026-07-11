package datasets

import "encoding/json"

// Asset is one row of the bundled asset catalog
// (pkg/datasets/assetmeta/assets.json): a fund, ETF, index or commodity with its
// resolution metadata (how to fetch its quotes) and its descriptive metadata
// (class, geography, sectors…). It is the single canonical catalog type,
// consumed across the toolkit (marketdata for resolution, suggest for the
// regime/factor analysis).
//
// Units: Fees is a published TER in percent per year (0 = unknown); Geography
// and Sectors map a label to a percent of holdings (each map sums to ~100);
// Leverage is a notional exposure multiple (1 = unlevered); Duration is in
// years; Exposures maps asset_class → notional weight (may sum past 1 for
// stacked funds).
type Asset struct {
	ID       string   `json:"id"`                  // canonical identifier (European ticker or ISIN)
	ISIN     string   `json:"isin"`                // ISIN; may be empty for indices/commodities
	Aliases  []string `json:"aliases"`             // alternative identifiers accepted on input
	Name     string   `json:"name"`                // full asset / share-class name
	UCITS    bool     `json:"ucits"`               // UCITS-regulated fund or ETF
	EURetail bool     `json:"eu_retail,omitempty"` // buyable by an EU/French retail investor: UCITS, or EU-listed with a PRIIPs KID (ETCs, listed closed-end funds); false for US-listed funds without a KID; omitted for non-tradable series (indices, spot, futures)
	Source   string   `json:"source"`              // quote source: "yahoo", "ft", "morningstar", "stooq", or "index" (served from the embedded reconstruction, no live symbol)
	Symbol   string   `json:"symbol"`              // Yahoo/Stooq symbol or Morningstar id; empty for ft and index
	Xid      string   `json:"xid"`                 // FT internal id; empty otherwise
	Currency string   `json:"currency"`            // quote currency (ISO 4217)
	Fees     float64  `json:"fees"`                // published TER, percent per year; 0 = unknown
	Since    string   `json:"since,omitempty"`     // inception date (YYYY-MM-DD); bounds real-data coverage before SIM extension

	AssetClass     string             `json:"asset_class"`               // e.g. "equity", "government-bond", "gold"
	Underlying     string             `json:"underlying"`                // free-text description of the holdings
	Benchmark      string             `json:"benchmark_index"`           // tracked index, when applicable
	Strategy       string             `json:"strategy"`                  // e.g. "physical-replication", "synthetic-swap"
	Geography      map[string]float64 `json:"geography"`                 // country/region → percent of holdings
	Sectors        map[string]float64 `json:"sectors"`                   // GICS sector → percent of holdings
	Distribution   string             `json:"distribution"`              // "accumulating" or "distributing"
	Leverage       float64            `json:"leverage"`                  // notional exposure multiple (1 = unlevered)
	CurrencyHedged bool               `json:"currency_hedged,omitempty"` // share class hedges FX exposure back to HedgedTo
	HedgedTo       string             `json:"hedged_to,omitempty"`       // target currency of the hedge (ISO 4217); empty = unhedged
	Duration       float64            `json:"duration,omitempty"`        // effective duration in years (fixed income); 0 = n/a
	Exposures      map[string]float64 `json:"exposures,omitempty"`       // asset_class → notional weight, for stacked / efficient-core funds
	Factors        []string           `json:"factors,omitempty"`         // explicit factor tilts: value, size, momentum, quality, low-vol
	Notes          string             `json:"notes"`                     // human-readable notes
	Confidence     string             `json:"confidence"`                // metadata confidence: "high", "medium", "low"
	Sources        []string           `json:"sources"`                   // provenance URLs
}

// Catalog parses the embedded asset metadata into the full list of catalog
// assets, the structured, typed view of assets.json for library consumers.
// For the raw bytes (to decode into your own type), use AssetMeta instead.
//
// It panics only if the bundled JSON is corrupt, which is impossible at
// runtime: the file is validated at build time.
func Catalog() []Asset {
	var assets []Asset
	if err := json.Unmarshal(AssetMeta(), &assets); err != nil {
		panic("datasets: cannot parse the embedded asset catalog: " + err.Error())
	}
	return assets
}
