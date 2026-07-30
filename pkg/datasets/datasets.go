package datasets

import (
	"embed"
	"io/fs"
)

//go:embed simdata assetmeta/assets.json
var bundle embed.FS

// Simdata returns the embedded simulated-history files.
func Simdata() fs.FS {
	sub, err := fs.Sub(bundle, "simdata")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
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
