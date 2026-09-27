package marketdata

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/bpineau/pofo/pkg/datasets"
)

// bundles are the two families of series the binary embeds, in the order
// Bundled looks through them. No identifier names a series in both (a test
// holds it), so the order decides nothing and a lookup is never ambiguous.
var bundles = []struct {
	source string // Series.Source of what it serves
	dir    string // where the files live in the repository, for errors
	fsys   func() fs.FS
}{
	// Backcasts: the reconstructed history of a catalog asset, which the SIM
	// convention splices in front of its real quotes (FetchExtended).
	{"simdata", "pkg/datasets/simdata", datasets.Simdata},
	// References: the indices, yields, cash rates and NAV snapshots the
	// recipes are built and validated on, none of them a catalog asset.
	{"refdata", "pkg/datasets/refdata", datasets.Refdata},
}

// Bundled returns the series the binary embeds under id, in one call: no
// Client, no cache, no network. id goes through CanonicalID, so an alias
// reaches its asset's file. BundledIDs lists every identifier it answers.
//
// Two families answer, told apart by Series.Source:
//
//   - "simdata", the BACKCAST of a catalog asset (pkg/datasets/simdata): its
//     history as the recipe wrote it, in the quote currency of the asset's
//     catalog record, which Series.Currency carries. Fetching "<id>SIM"
//     through Client.FetchExtended is what splices it in front of the live
//     quotes.
//   - "refdata", a REFERENCE series (pkg/datasets/refdata): an index, a
//     yield, a cash rate, a fund's NAV snapshot, the series the recipes are
//     built on and the goldens validate against. None is a catalog asset,
//     and none states its currency, which the identifier's suffix usually
//     names (TREASURY-LONG-USD, EUROGOV-EUR); a few are rate LEVELS in
//     annualized percent (TREASURY-LONG-YIELD, TBILL-3M), which the file's
//     Name says and which never belong in a return computation.
//
// The file's "# name:" header becomes Series.Name and its "# junctions:"
// header Series.Junctions, the dates across which no return may be read. An
// identifier nothing is bundled under is an error that wraps fs.ErrNotExist
// and says where it looked.
//
// The series is the one the binary was built with, so a data refresh moves
// its numbers.
func Bundled(id string) (*Series, error) {
	canonical := CanonicalID(id)
	for _, b := range bundles {
		s, ok, err := ReadSimdataFS(b.fsys(), canonical)
		if err != nil {
			return nil, fmt.Errorf("marketdata: bundled %s: %w", canonical, err)
		}
		if !ok {
			continue
		}
		s.Symbol, s.Source = canonical, b.source
		if e, isAsset := catalogByID()[canonical]; isAsset && b.source == "simdata" {
			s.Currency = e.Currency
		}
		return s, nil
	}
	dirs := make([]string, len(bundles))
	for i, b := range bundles {
		dirs[i] = b.dir
	}
	return nil, fmt.Errorf("marketdata: nothing bundled under %s (looked for %s.csv in %s; BundledIDs lists what is): %w",
		canonical, sanitizeFilename(canonical), strings.Join(dirs, " and "), fs.ErrNotExist)
}

// BundledIDs lists, sorted, every identifier Bundled answers: the catalog
// assets that carry a backcast and the reference series. It reads directory
// listings only, so it is cheap; Bundled on an entry says what it is.
func BundledIDs() []string {
	var ids []string
	for _, b := range bundles {
		entries, err := fs.ReadDir(b.fsys(), ".")
		if err != nil {
			panic(err) // broken embedded layout: impossible at runtime
		}
		for _, e := range entries {
			if id, ok := strings.CutSuffix(e.Name(), ".csv"); ok && !e.IsDir() {
				ids = append(ids, id)
			}
		}
	}
	slices.Sort(ids)
	return slices.Compact(ids)
}
