// Package datasets embeds the repository's versioned data into the binary:
// the permanent simulated histories (simdata/) and the imported reference
// series (refdata/). The binary can therefore run from any directory; after
// a regeneration (-gen-simdata), a recompilation is needed to re-embed the
// files.
package datasets

import (
	"embed"
	"io/fs"
)

//go:embed simdata refdata
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
