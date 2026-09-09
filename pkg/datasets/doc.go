// Package datasets embeds the repository's versioned data into the binary, so
// it runs from any directory: the permanent simulated histories (simdata/),
// the long reference series they are built on (refdata/), the catalog asset
// metadata (assetmeta/), and the three research panels behind the FIRE and
// macro-regime work (broadsample/, cape/, macropanel/). After a regeneration
// (-gen-simdata, make refresh), a recompilation re-embeds the files.
//
// Simdata and Refdata expose their directory as an fs.FS of "<canonical
// id>.csv" files (comment stamps, then date,close rows: read one with
// marketdata.ReadSimdataFS); the panels are returned as raw CSV bytes.
//
// Catalog returns the typed asset records (with their geography, sectors,
// factors and exposures), and AssetMeta the same data as raw JSON. For a
// resolution-aware, by-identifier lookup that also accepts aliases and fund
// tickers, use marketdata.Lookup.
package datasets
