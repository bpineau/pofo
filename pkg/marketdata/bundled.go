package marketdata

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Bundled returns the series the binary embeds under id, in one call: no
// Client, no cache, no network. BundledIDs lists every identifier it answers.
//
// It obeys the SIM convention every other door obeys: a bare identifier means
// the thing itself, and a reconstruction answers only to a name that says so.
// Two families answer, told apart by Series.Source:
//
//   - "simdata", the SIM history of a catalog asset (pkg/datasets/simdata):
//     the recipe's reconstruction with the asset's real quotes grafted on
//     from its inception, as of the last data refresh, in the quote currency
//     of the asset's catalog record, which Series.Currency carries. It
//     answers to the SIM form of the identifier ("DBMFSIM", "IWDASIM"),
//     whose base goes through CanonicalID, so an alias or an ISIN reaches
//     the asset's file. The bare form of a quoted asset ("IWDA") is an
//     error: its real quotes are not bundled, and serving a reconstruction
//     under its plain name is exactly the confusion the convention exists
//     to prevent; Client.FetchExtended fetches them. The one exception is a
//     catalog asset whose source is "index" (MSCIWORLD, SP500), which has no
//     quotes of its own: it answers to both forms, as FetchExtended serves
//     it long under both.
//   - "refdata", a REFERENCE series (pkg/datasets/refdata): an index, a
//     yield, a cash rate, a fund's NAV snapshot, the series the recipes are
//     built on and the goldens validate against. None is a catalog asset,
//     none carries a SIM form, and none states its currency, which the
//     identifier's suffix usually names (TREASURY-LONG-USD, EUROGOV-EUR); a
//     few are rate LEVELS in annualized percent (TREASURY-LONG-YIELD,
//     TBILL-3M), which the file's Name says and which never belong in a
//     return computation.
//
// Series.Symbol is the canonical identifier that answered ("IE00B4L5Y983SIM"
// for "IWDASIM"), the one BundledIDs lists. The file's "# name:" header
// becomes Series.Name and its "# junctions:" header Series.Junctions, the
// dates across which no return may be read. An identifier nothing is bundled
// under, the bare form of a quoted asset included, is an error that wraps
// fs.ErrNotExist and says what to ask for instead.
//
// The series is the one the binary was built with, so a data refresh moves
// its numbers.
func Bundled(id string) (*Series, error) {
	base, sim := SplitSim(id)
	asset := CanonicalID(base)
	if e, ok := catalogByID()[asset]; ok {
		s, found, err := ReadSimdataFS(datasets.Simdata(), asset)
		if err != nil {
			return nil, fmt.Errorf("marketdata: bundled %s: %w", asset, err)
		}
		if found {
			index := e.Source == "index"
			if !sim && !index {
				return nil, fmt.Errorf("marketdata: %s is a quoted catalog asset, whose real quotes are not bundled: "+
					"Bundled(%q) is its SIM history (reconstruction plus real quotes as of the last refresh), "+
					"Client.FetchExtended(%q) its real quotes: %w", asset, asset+"SIM", asset, fs.ErrNotExist)
			}
			s.Symbol, s.Source, s.Currency = bundledID(asset, index), "simdata", e.Currency
			return s, nil
		}
	}
	if !sim {
		s, found, err := ReadSimdataFS(datasets.Refdata(), id)
		if err != nil {
			return nil, fmt.Errorf("marketdata: bundled %s: %w", CanonicalID(id), err)
		}
		if found {
			s.Symbol, s.Source = CanonicalID(id), "refdata"
			return s, nil
		}
	}
	return nil, fmt.Errorf("marketdata: nothing bundled under %s (a catalog asset's SIM history lives in "+
		"pkg/datasets/simdata, a reference series in pkg/datasets/refdata; BundledIDs lists what is): %w",
		strings.ToUpper(strings.TrimSpace(id)), fs.ErrNotExist)
}

// bundledID is the identifier Bundled answers a catalog asset's simdata file
// under: its SIM form, or the bare id of an index, which has no other.
func bundledID(asset string, index bool) string {
	if index {
		return asset
	}
	return asset + "SIM"
}

// BundledIDs lists, sorted, the canonical identifier of every series Bundled
// answers: the SIM form of each quoted catalog asset that carries a
// reconstruction ("IE00B4L5Y983SIM"), the bare id of each catalog index
// ("SP500"), and the reference series ("TBILL-3M"). It reads directory
// listings only, so it is cheap; Bundled on an entry says what it is.
func BundledIDs() []string {
	var ids []string
	for _, fsys := range []fs.FS{datasets.Simdata(), datasets.Refdata()} {
		entries, err := fs.ReadDir(fsys, ".")
		if err != nil {
			panic(err) // broken embedded layout: impossible at runtime
		}
		for _, e := range entries {
			id, ok := strings.CutSuffix(e.Name(), ".csv")
			if !ok || e.IsDir() {
				continue
			}
			if a, isAsset := catalogByID()[id]; isAsset {
				id = bundledID(id, a.Source == "index")
			}
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return slices.Compact(ids)
}
