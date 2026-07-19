// Package webui holds the shared visual identity for pofo's two HTML surfaces:
// the static portfolio tearsheet (pkg/report) and the interactive FIRE
// decumulation explorer (pkg/decumul/web). Both embed the same design tokens
// and base component styles from theme.css, and the same identity typefaces
// from FontsCSS (Instrument Sans, Spline Sans Mono; OFL-licensed WOFF2 shipped
// in the binary), so they read as one product. Everything is embedded
// (go:embed): no build step, no network fetch at runtime, plain `go build`.
//
// The identity ("instrument") and the reasoning behind it are documented in
// docs/webui-instrument-redesign.md.
package webui

import _ "embed"

// CSS is the shared design-token and base-component stylesheet. The report
// inlines it into its self-contained document; the FIRE server serves it at
// /theme.css. View-specific rules live alongside each view.
//
//go:embed theme.css
var CSS string

// WarmSkin remaps the instrument tokens to the FIRE book's warm paper-and-ink
// palette. Inlined or linked after CSS, it gives the constellation's reading
// surfaces (the hub and the /view report) the book's identity while the FIRE
// simulator keeps the instrument look. It only redefines variables and the
// heading face, so any component built on the tokens reskins with no markup
// change.
//
//go:embed warmskin.css
var WarmSkin string
