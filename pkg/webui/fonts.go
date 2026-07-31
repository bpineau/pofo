package webui

import (
	"embed"
	"encoding/base64"
	"fmt"
	"strings"
)

// The identity typefaces ship inside the binary as latin-subset variable
// WOFF2 files (SIL Open Font License 1.1; see fonts/OFL.txt), so both HTML
// surfaces render identically everywhere with no build step and no network
// fetch at runtime. FontsCSS exposes them the same way theme.css is exposed:
// a stylesheet string the report inlines and the FIRE server serves.
//
//go:embed fonts/*.woff2
var fontFiles embed.FS

// FontsCSS holds the @font-face rules for the embedded identity typefaces
// (Instrument Sans for text, Spline Sans Mono for figures), with the font
// data inlined as base64 data: URIs. Serve it as text/css or inline it in a
// <style> element before theme.css.
var FontsCSS = buildFontsCSS()

func buildFontsCSS() string {
	faces := []struct {
		file, family string
		lo, hi       int // wght variation axis range
	}{
		{"fonts/instrumentsans.woff2", "Instrument Sans", 400, 700},
		{"fonts/splinesansmono.woff2", "Spline Sans Mono", 400, 700},
	}
	var sb strings.Builder
	for _, f := range faces {
		data, err := fontFiles.ReadFile(f.file)
		if err != nil {
			panic("webui: embedded font missing: " + f.file)
		}
		fmt.Fprintf(&sb,
			"@font-face{font-family:'%s';font-style:normal;font-weight:%d %d;"+
				"font-display:swap;src:url(data:font/woff2;base64,%s) format('woff2')}\n",
			f.family, f.lo, f.hi, base64.StdEncoding.EncodeToString(data))
	}
	return sb.String()
}
