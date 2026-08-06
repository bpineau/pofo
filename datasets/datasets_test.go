package datasets

import (
	"io/fs"
	"testing"
)

func TestEmbeddedDatasetsPresent(t *testing.T) {
	sim, err := fs.ReadDir(Simdata(), ".")
	if err != nil || len(sim) < 8 {
		t.Fatalf("embedded simdata incomplete: %d files, %v", len(sim), err)
	}
	ref, err := fs.ReadDir(Refdata(), ".")
	if err != nil || len(ref) < 6 {
		t.Fatalf("embedded refdata incomplete: %d files, %v", len(ref), err)
	}
}
