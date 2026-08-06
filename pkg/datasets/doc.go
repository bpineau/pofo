// Package datasets embeds the repository's versioned data into the binary:
// the permanent simulated histories (simdata/) and the catalog asset
// metadata (assetmeta/). The binary can therefore run from any directory;
// after a regeneration (-gen-simdata), a recompilation is needed to re-embed
// the files.
//
// Catalog returns the typed asset records (with their geography, sectors,
// factors and exposures), and AssetMeta the same data as raw JSON. For a
// resolution-aware, by-identifier lookup that also accepts aliases and fund
// tickers, use marketdata.Lookup.
package datasets
