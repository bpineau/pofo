package datasets

import (
	"embed"
	"io/fs"
)

//go:embed simdata assetmeta/assets.json refdata broadsample/country-real.csv cape/shiller-cape.csv macropanel/oecd-monthly.csv
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

// CAPE returns the embedded Shiller CAPE (PE10) monthly series (date,cape),
// used by the FIRE explorer to anchor the central case to today's valuation.
// Regenerate with "make cape".
func CAPE() []byte {
	b, err := bundle.ReadFile("cape/shiller-cape.csv")
	if err != nil {
		panic(err) // broken repository layout: impossible at runtime
	}
	return b
}

// MacroPanel returns the embedded multi-country monthly macro panel
// (date,iso,ip,cpi,shortrate,longrate,shareprice from OECD MEI, 30 economies):
// the growth/inflation and short/long-rate drivers behind macro-regime analysis
// and the growth x inflation breadth model. Regenerate with "make macropanel".
func MacroPanel() []byte {
	b, err := bundle.ReadFile("macropanel/oecd-monthly.csv")
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
