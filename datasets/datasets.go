// Package datasets embarque dans le binaire les données versionnées du
// dépôt : les historiques simulés permanents (simdata/) et les séries de
// référence importées (refdata/). Le binaire peut ainsi tourner depuis
// n'importe quel répertoire ; après une régénération (-gen-simdata), une
// recompilation est nécessaire pour ré-embarquer les fichiers.
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
		panic(err) // structure du dépôt cassée: impossible au runtime
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
