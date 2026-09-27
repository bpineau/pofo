package marketdata

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

func TestBundled(t *testing.T) {
	ref, err := Bundled("treasury-long-yield")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Symbol != "TREASURY-LONG-YIELD" || ref.Source != "refdata" || ref.Name == "" || ref.Currency != "" {
		t.Errorf("reference series = %s %q %q %q", ref.Symbol, ref.Source, ref.Name, ref.Currency)
	}
	if !slices.ContainsFunc(ref.Junctions, func(j time.Time) bool { return j.Equal(d(1973, 1, 4)) }) {
		t.Errorf("junctions %v lack the declared 1973-01-04", ref.Junctions)
	}

	sim, err := Bundled("dbmf")
	if err != nil {
		t.Fatal(err)
	}
	e, _ := Lookup("DBMF")
	if sim.Symbol != "DBMF" || sim.Source != "simdata" || sim.Currency != e.Currency || sim.Len() < 1000 {
		t.Errorf("backcast = %s %q %q, %d points", sim.Symbol, sim.Source, sim.Currency, sim.Len())
	}

	_, err = Bundled("NO-SUCH-SERIES")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("error = %v, want one wrapping fs.ErrNotExist", err)
	}
	for _, where := range []string{"NO-SUCH-SERIES.csv", "pkg/datasets/simdata", "pkg/datasets/refdata", "BundledIDs"} {
		if !strings.Contains(err.Error(), where) {
			t.Errorf("error %q does not say %q", err, where)
		}
	}
}

// Every identifier BundledIDs lists is one Bundled answers, and the two
// families never share one: the lookup order is then irrelevant, and "pofo
// -dump" may serve a non-catalog identifier from the bundle knowing it is a
// reference series.
func TestBundledIDs(t *testing.T) {
	ids := BundledIDs()
	if len(ids) < 60 || !slices.IsSorted(ids) {
		t.Fatalf("BundledIDs: %d ids, sorted=%v", len(ids), slices.IsSorted(ids))
	}
	names := func(fsys fs.FS) map[string]bool {
		entries, err := fs.ReadDir(fsys, ".")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]bool{}
		for _, e := range entries {
			m[strings.TrimSuffix(e.Name(), ".csv")] = true
		}
		return m
	}
	sim, ref := names(datasets.Simdata()), names(datasets.Refdata())
	if len(sim)+len(ref) != len(ids) {
		t.Errorf("%d backcasts + %d references != %d ids", len(sim), len(ref), len(ids))
	}
	for _, id := range ids {
		s, err := Bundled(id)
		if err != nil {
			t.Errorf("%s: %v", id, err)
			continue
		}
		switch {
		case sim[id] && ref[id]:
			t.Errorf("%s is both a backcast and a reference series", id)
		case s.Symbol != id:
			t.Errorf("Bundled(%s).Symbol = %s: the file name is not canonical", id, s.Symbol)
		case ref[id] && KnownLocal(id):
			t.Errorf("reference series %s shadows a catalog identifier", id)
		case sim[id]:
			if _, ok := Lookup(id); !ok {
				t.Errorf("backcast %s belongs to no catalog asset", id)
			}
		}
	}
}
