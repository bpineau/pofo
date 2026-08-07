package firebook

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"
)

var epubModified = time.Date(2026, 7, 23, 10, 30, 0, 0, time.UTC)

func buildEPUB(t *testing.T) *zip.Reader {
	t.Helper()
	raw, err := EPUB(epubModified)
	if err != nil {
		t.Fatalf("EPUB: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("EPUB returned no bytes")
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	return zr
}

// TestEPUBInventory pins the exact set of files in the archive: the OCF
// machinery, one title page, one page per category, one page per article, and
// the shared nav/ncx/css.
func TestEPUBInventory(t *testing.T) {
	zr := buildEPUB(t)

	got := map[string]bool{}
	for _, f := range zr.File {
		got[f.Name] = true
	}

	want := []string{
		"mimetype",
		"META-INF/container.xml",
		"OEBPS/content.opf",
		"OEBPS/nav.xhtml",
		"OEBPS/toc.ncx",
		"OEBPS/style.css",
		"OEBPS/titlepage.xhtml",
	}
	articles := 0
	for i, cat := range Categories {
		want = append(want, "OEBPS/cat-"+itoa(i)+".xhtml")
		for _, a := range cat.Articles {
			want = append(want, "OEBPS/"+a.Slug+".xhtml")
			articles++
		}
	}
	if len(zr.File) != len(want) {
		t.Errorf("archive has %d entries, want %d", len(zr.File), len(want))
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing archive entry %q", name)
		}
	}
	// Sanity: the title page, every category and every article are present.
	if articles == 0 {
		t.Fatal("no articles counted from Categories")
	}
}

// TestEPUBCorpusGuard is the whole-corpus wellformedness gate: every content
// document produced (title page, category pages and all articles) must parse
// as XML, since EPUB content documents are served as XHTML. Any future article
// or renderer change that would break XML fails here.
func TestEPUBCorpusGuard(t *testing.T) {
	zr := buildEPUB(t)

	checked := 0
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "OEBPS/") || !strings.HasSuffix(f.Name, ".xhtml") {
			continue
		}
		if f.Name == "OEBPS/nav.xhtml" { // the TOC, not a book content document
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		dec := xml.NewDecoder(bytes.NewReader(data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("%s is not well-formed XML: %v", f.Name, err)
				break
			}
		}
		checked++
	}
	// Title page + every category + every article.
	want := 1 + len(Categories)
	for _, cat := range Categories {
		want += len(cat.Articles)
	}
	if checked != want {
		t.Errorf("guarded %d XHTML documents, want %d", checked, want)
	}
}

// TestEPUBDeterministic proves the bytes are reproducible for a fixed modified
// time, so the HTTP route can hash them for an ETag.
func TestEPUBDeterministic(t *testing.T) {
	a, err := EPUB(epubModified)
	if err != nil {
		t.Fatalf("EPUB: %v", err)
	}
	b, err := EPUB(epubModified)
	if err != nil {
		t.Fatalf("EPUB: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("EPUB output is not deterministic for a fixed modified time")
	}
}

// itoa is strconv.Itoa without the import churn in this small test file.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
