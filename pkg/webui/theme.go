// Package webui holds the shared visual identity for pofo's two HTML surfaces:
// the static portfolio tearsheet (pkg/report) and the interactive FIRE
// decumulation explorer (pkg/decumul/web). Both embed the same design tokens and
// base component styles from theme.css so they read as one product. Everything is
// embedded in the binary (go:embed): no build step, no external fonts, plain
// `go build`.
package webui

import _ "embed"

// CSS is the shared design-token and base-component stylesheet. The report
// inlines it into its self-contained document; the FIRE server serves it at
// /theme.css. View-specific rules live alongside each view.
//
//go:embed theme.css
var CSS string
