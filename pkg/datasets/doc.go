// Package datasets embeds the repository's versioned data into the binary, so
// it runs from any directory: the permanent simulated histories (simdata/),
// the long reference series they are built on (refdata/), the catalog asset
// metadata (assetmeta/), and the three research panels behind the FIRE and
// macro-regime work (broadsample/, cape/, macropanel/). After a regeneration
// (-gen-simdata, make refresh), a recompilation re-embeds the files.
//
// Most consumers never import it: marketdata.Bundled reads any of these
// series by identifier into a Series, marketdata.BundledIDs lists them, and
// marketdata.Lookup answers for a catalog record by any accepted identifier.
// This package is the raw layer underneath:
//
//   - [Simdata] and [Refdata] expose their directory as an [fs.FS] of
//     "<canonical id>.csv" files (comment stamps, then date,close rows);
//   - [Catalog] returns the typed asset records ([Asset]: class, geography,
//     sectors, factors, exposures, fees in PERCENT per year), and
//     [AssetMeta] the same data as raw JSON;
//   - [BroadSample], [CAPE] and [MacroPanel] return the research panels as
//     raw CSV bytes.
package datasets
