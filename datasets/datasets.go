// Package datasets embeds the repository's versioned data into the binary:
// the permanent simulated histories (simdata/), the imported reference
// series (refdata/) and the catalog asset metadata (assetmeta/). The binary
// can therefore run from any directory; after a regeneration (-gen-simdata),
// a recompilation is needed to re-embed the files.
package datasets

import (
	"embed"
	"io/fs"
)

//go:embed simdata refdata assetmeta/assets.json
var bundle embed.FS

// Simdata returns the embedded simulated-history files.
func Simdata() fs.FS {
	sub, err := fs.Sub(bundle, "simdata")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
	}
	return sub
}

// Refdata returns the embedded imported reference series.
func Refdata() fs.FS {
	sub, err := fs.Sub(bundle, "refdata")
	if err != nil {
		panic(err)
	}
	return sub
}

// AssetMeta returns the embedded asset-metadata JSON (the factual tags for
// the bundled catalog used by the -suggest analysis).
func AssetMeta() []byte {
	b, err := bundle.ReadFile("assetmeta/assets.json")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
	}
	return b
}
