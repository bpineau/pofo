package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/firebook"
)

// -book-lang picks the edition, and only the edition: the file name is always
// the one the caller asked for, and an unknown language is an error, not a
// silent fallback to French.
func TestRunExportEpubBookLang(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct {
		lang string
		want *firebook.Edition
	}{{"", firebook.French}, {"fr", firebook.French}, {"en", firebook.English}} {
		path := filepath.Join(dir, "book-"+c.lang+".epub")
		if err := runExportEpub(path, c.lang); err != nil {
			t.Fatalf("-book-lang %q: %v", c.lang, err)
		}
		opf, size := epubPackage(t, path)
		// The package document carries the edition's identity: its title and
		// its own urn:uuid, which the two editions never share.
		for _, want := range []string{c.want.SiteName, c.want.EPUBIdentifier, "<dc:language>" + c.want.Lang + "</dc:language>"} {
			if !strings.Contains(opf, want) {
				t.Errorf("-book-lang %q: package document misses %q", c.lang, want)
			}
		}
		if size < 100_000 {
			t.Errorf("-book-lang %q: %d bytes, too small to be the whole book", c.lang, size)
		}
	}

	if err := runExportEpub(filepath.Join(dir, "nope.epub"), "de"); err == nil {
		t.Error("-book-lang de: want an error, got none")
	}
}

// epubPackage returns the OPF package document of the EPUB at path, and the
// file's size.
func epubPackage(t *testing.T, path string) (string, int64) {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".opf") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening %s in %s: %v", f.Name, path, err)
		}
		defer rc.Close()
		body, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("reading %s in %s: %v", f.Name, path, err)
		}
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(body), fi.Size()
	}
	t.Fatalf("%s carries no package document", path)
	return "", 0
}
