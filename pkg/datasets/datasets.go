package datasets

import (
	"embed"
	"io/fs"
)

//go:embed simdata assetmeta/assets.json refdata broadsample/country-real.csv
var bundle embed.FS

// Simdata returns the embedded simulated-history files.
func Simdata() fs.FS {
	sub, err := fs.Sub(bundle, "simdata")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
	}
	return sub
}

// Refdata returns the embedded long reference series used by the simgen recipes
// (e.g. the MSCI World total-return history), so regeneration is self-contained
// and needs no external -refdata directory.
func Refdata() fs.FS {
	sub, err := fs.Sub(bundle, "refdata")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
	}
	return sub
}

// BroadSample returns the embedded broad-sample per-country real-return table
// (Jorda-Schularick-Taylor, 18 economies, 1870-2020; iso,year,equity,bond,bill
// as real annual fractions), pool-bootstrapped by the FIRE explorer's empirical
// model. Regenerate with "make broadsample".
func BroadSample() []byte {
	b, err := bundle.ReadFile("broadsample/country-real.csv")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
	}
	return b
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
