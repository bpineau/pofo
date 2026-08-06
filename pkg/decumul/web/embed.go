package web

import (
	"embed"
	"io/fs"
)

//go:embed assets/index.html assets/app.js assets/app.css
var assets embed.FS

// mustSub exposes the assets directory at the URL root (so / serves
// assets/index.html).
func mustSub() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}
