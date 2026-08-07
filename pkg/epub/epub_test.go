package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"
)

var testModified = time.Date(2026, 7, 23, 10, 30, 0, 0, time.UTC)

// sampleBook is a nested two-category fixture used across the structural tests.
func sampleBook() *Book {
	return &Book{
		Title:       "Un titre <avec> & entités",
		Author:      "Anne Auteur",
		Language:    "fr",
		Identifier:  "urn:uuid:11111111-2222-3333-4444-555555555555",
		Description: "Une description brève.",
		Modified:    testModified,
		CSS:         "body{font-family:serif}",
		Chapters: []Chapter{
			{FileName: "intro.xhtml", Title: "Introduction", Body: "<p>Bienvenue.</p>"},
			{
				FileName: "cat-a.xhtml", Title: "Catégorie A", Body: "<p>Aperçu A.</p>",
				Children: []Chapter{
					{FileName: "a1.xhtml", Title: "Article A1", Body: "<p>A1.</p>"},
					{FileName: "a2.xhtml", Title: "Article A2", Body: "<p>A2.</p>"},
				},
			},
			{
				FileName: "cat-b.xhtml", Title: "Catégorie B", Body: "<p>Aperçu B.</p>",
				Children: []Chapter{
					{FileName: "b1.xhtml", Title: "Article B1", Body: "<p>B1.</p>"},
				},
			},
		},
	}
}

func writeToBytes(t *testing.T, b *Book) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return buf.Bytes()
}

func openZip(t *testing.T, raw []byte) *zip.Reader {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	return zr
}

func readEntry(t *testing.T, zr *zip.Reader, name string) []byte {
	t.Helper()
	f, err := zr.Open(name)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

// TestMimetypeFirst enforces the OCF offset rule: the mimetype entry is first,
// stored (uncompressed), carries no extra field, and its content is exact.
func TestMimetypeFirst(t *testing.T) {
	raw := writeToBytes(t, sampleBook())

	// Raw local-file-header inspection.
	if got := string(raw[30:38]); got != "mimetype" {
		t.Errorf("bytes 30..38 = %q, want %q", got, "mimetype")
	}
	if method := uint16(raw[8]) | uint16(raw[9])<<8; method != zip.Store {
		t.Errorf("mimetype method = %d, want Store (%d)", method, zip.Store)
	}
	if extraLen := uint16(raw[28]) | uint16(raw[29])<<8; extraLen != 0 {
		t.Errorf("mimetype extra field length = %d, want 0", extraLen)
	}

	zr := openZip(t, raw)
	if zr.File[0].Name != "mimetype" {
		t.Fatalf("first entry = %q, want mimetype", zr.File[0].Name)
	}
	if zr.File[0].Method != zip.Store {
		t.Errorf("mimetype not stored")
	}
	if got := string(readEntry(t, zr, "mimetype")); got != "application/epub+zip" {
		t.Errorf("mimetype content = %q", got)
	}
}

// container.xml structs.
type xmlContainer struct {
	Rootfiles struct {
		Rootfile []struct {
			FullPath  string `xml:"full-path,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"rootfile"`
	} `xml:"rootfiles"`
}

func TestContainer(t *testing.T) {
	zr := openZip(t, writeToBytes(t, sampleBook()))
	var c xmlContainer
	if err := xml.Unmarshal(readEntry(t, zr, "META-INF/container.xml"), &c); err != nil {
		t.Fatalf("parse container.xml: %v", err)
	}
	if len(c.Rootfiles.Rootfile) != 1 {
		t.Fatalf("want 1 rootfile, got %d", len(c.Rootfiles.Rootfile))
	}
	if got := c.Rootfiles.Rootfile[0].FullPath; got != "OEBPS/content.opf" {
		t.Errorf("rootfile full-path = %q", got)
	}
	if got := c.Rootfiles.Rootfile[0].MediaType; got != "application/oebps-package+xml" {
		t.Errorf("rootfile media-type = %q", got)
	}
}

// content.opf structs (matched by local name; encoding/xml ignores namespace
// when the tag omits one).
type opfPackage struct {
	Version  string `xml:"version,attr"`
	UniqueID string `xml:"unique-identifier,attr"`
	Metadata struct {
		Identifiers []struct {
			ID    string `xml:"id,attr"`
			Value string `xml:",chardata"`
		} `xml:"identifier"`
		Title       string `xml:"title"`
		Language    string `xml:"language"`
		Creator     string `xml:"creator"`
		Description string `xml:"description"`
		Metas       []struct {
			Property string `xml:"property,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID         string `xml:"id,attr"`
			Href       string `xml:"href,attr"`
			MediaType  string `xml:"media-type,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		Toc  string `xml:"toc,attr"`
		Refs []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

func parseOPF(t *testing.T, zr *zip.Reader) opfPackage {
	t.Helper()
	var p opfPackage
	if err := xml.Unmarshal(readEntry(t, zr, "OEBPS/content.opf"), &p); err != nil {
		t.Fatalf("parse content.opf: %v", err)
	}
	return p
}

func TestOPFMetadata(t *testing.T) {
	b := sampleBook()
	zr := openZip(t, writeToBytes(t, b))
	p := parseOPF(t, zr)

	if p.Version != "3.0" {
		t.Errorf("package version = %q, want 3.0", p.Version)
	}
	if len(p.Metadata.Identifiers) != 1 {
		t.Fatalf("want 1 dc:identifier, got %d", len(p.Metadata.Identifiers))
	}
	id := p.Metadata.Identifiers[0]
	if id.ID != p.UniqueID {
		t.Errorf("unique-identifier %q does not resolve to a dc:identifier id %q", p.UniqueID, id.ID)
	}
	if id.Value != b.Identifier {
		t.Errorf("dc:identifier = %q, want %q", id.Value, b.Identifier)
	}
	if p.Metadata.Title != b.Title {
		t.Errorf("dc:title = %q, want %q", p.Metadata.Title, b.Title)
	}
	if p.Metadata.Language != b.Language {
		t.Errorf("dc:language = %q, want %q", p.Metadata.Language, b.Language)
	}
	if p.Metadata.Creator != b.Author {
		t.Errorf("dc:creator = %q, want %q", p.Metadata.Creator, b.Author)
	}
	if p.Metadata.Description != b.Description {
		t.Errorf("dc:description = %q, want %q", p.Metadata.Description, b.Description)
	}
	var modified string
	for _, m := range p.Metadata.Metas {
		if m.Property == "dcterms:modified" {
			modified = m.Value
		}
	}
	if modified != "2026-07-23T10:30:00Z" {
		t.Errorf("dcterms:modified = %q, want 2026-07-23T10:30:00Z", modified)
	}
}

func TestOPFManifest(t *testing.T) {
	zr := openZip(t, writeToBytes(t, sampleBook()))
	p := parseOPF(t, zr)

	byHref := map[string]string{} // href -> properties
	ids := map[string]bool{}
	hasNav, hasCSS, hasNCX := false, false, false
	for _, it := range p.Manifest.Items {
		if ids[it.ID] {
			t.Errorf("duplicate manifest id %q", it.ID)
		}
		ids[it.ID] = true
		byHref[it.Href] = it.Properties
		switch it.Href {
		case "nav.xhtml":
			hasNav = true
			if it.Properties != "nav" {
				t.Errorf("nav.xhtml properties = %q, want nav", it.Properties)
			}
			if it.MediaType != "application/xhtml+xml" {
				t.Errorf("nav media-type = %q", it.MediaType)
			}
		case "style.css":
			hasCSS = true
			if it.MediaType != "text/css" {
				t.Errorf("style.css media-type = %q", it.MediaType)
			}
		case "toc.ncx":
			hasNCX = true
			if it.MediaType != "application/x-dtbncx+xml" {
				t.Errorf("toc.ncx media-type = %q", it.MediaType)
			}
		}
	}
	if !hasNav || !hasCSS || !hasNCX {
		t.Errorf("manifest missing entries: nav=%v css=%v ncx=%v", hasNav, hasCSS, hasNCX)
	}
	for _, want := range []string{"intro.xhtml", "cat-a.xhtml", "a1.xhtml", "a2.xhtml", "cat-b.xhtml", "b1.xhtml"} {
		if _, ok := byHref[want]; !ok {
			t.Errorf("manifest missing chapter %q", want)
		}
	}
}

func TestSpineOrder(t *testing.T) {
	zr := openZip(t, writeToBytes(t, sampleBook()))
	p := parseOPF(t, zr)

	// idref -> href from the manifest.
	href := map[string]string{}
	for _, it := range p.Manifest.Items {
		href[it.ID] = it.Href
	}
	if p.Spine.Toc == "" {
		t.Errorf("spine has no toc attribute (needed for the ncx)")
	}
	var order []string
	for _, r := range p.Spine.Refs {
		order = append(order, href[r.IDRef])
		if href[r.IDRef] == "nav.xhtml" {
			t.Errorf("nav.xhtml must not appear in the spine")
		}
	}
	want := []string{"intro.xhtml", "cat-a.xhtml", "a1.xhtml", "a2.xhtml", "cat-b.xhtml", "b1.xhtml"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("spine order = %v, want %v", order, want)
	}
}

func TestSpineCoverFirst(t *testing.T) {
	b := sampleBook()
	b.Cover = []byte("\x89PNG\r\n\x1a\nfake-png-body")
	zr := openZip(t, writeToBytes(t, b))
	p := parseOPF(t, zr)

	href := map[string]string{}
	for _, it := range p.Manifest.Items {
		href[it.ID] = it.Href
	}
	if len(p.Spine.Refs) == 0 || href[p.Spine.Refs[0].IDRef] != "cover.xhtml" {
		t.Fatalf("first spine item is not cover.xhtml")
	}
	// cover.png in manifest with the cover-image property.
	found := false
	for _, it := range p.Manifest.Items {
		if it.Href == "cover.png" {
			found = true
			if it.Properties != "cover-image" {
				t.Errorf("cover.png properties = %q, want cover-image", it.Properties)
			}
			if it.MediaType != "image/png" {
				t.Errorf("cover.png media-type = %q", it.MediaType)
			}
		}
	}
	if !found {
		t.Errorf("cover.png missing from manifest")
	}
	// The bytes are present.
	if got := readEntry(t, zr, "OEBPS/cover.png"); !bytes.HasPrefix(got, []byte("\x89PNG")) {
		t.Errorf("cover.png bytes not stored")
	}
	readEntry(t, zr, "OEBPS/cover.xhtml")
}

// nav.xhtml nesting.
type navDoc struct {
	Body struct {
		Nav struct {
			Type string `xml:"type,attr"`
			OL   navOL  `xml:"ol"`
		} `xml:"nav"`
	} `xml:"body"`
}
type navOL struct {
	LI []struct {
		A struct {
			Href string `xml:"href,attr"`
			Text string `xml:",chardata"`
		} `xml:"a"`
		OL *navOL `xml:"ol"`
	} `xml:"li"`
}

func TestNavNesting(t *testing.T) {
	zr := openZip(t, writeToBytes(t, sampleBook()))
	var nd navDoc
	if err := xml.Unmarshal(readEntry(t, zr, "OEBPS/nav.xhtml"), &nd); err != nil {
		t.Fatalf("parse nav.xhtml: %v", err)
	}
	if nd.Body.Nav.Type != "toc" {
		t.Errorf("nav epub:type = %q, want toc", nd.Body.Nav.Type)
	}
	top := nd.Body.Nav.OL.LI
	if len(top) != 3 {
		t.Fatalf("top-level nav entries = %d, want 3", len(top))
	}
	if top[0].A.Href != "intro.xhtml" || top[0].A.Text != "Introduction" {
		t.Errorf("first nav entry = %+v", top[0].A)
	}
	if top[0].OL != nil {
		t.Errorf("intro should have no nested ol")
	}
	if top[1].OL == nil || len(top[1].OL.LI) != 2 {
		t.Fatalf("category A should nest 2 children")
	}
	if top[1].OL.LI[0].A.Href != "a1.xhtml" {
		t.Errorf("first child of A = %q", top[1].OL.LI[0].A.Href)
	}
}

// toc.ncx tree.
type ncxDoc struct {
	Head struct {
		Metas []struct {
			Name    string `xml:"name,attr"`
			Content string `xml:"content,attr"`
		} `xml:"meta"`
	} `xml:"head"`
	NavMap struct {
		Points []ncxPoint `xml:"navPoint"`
	} `xml:"navMap"`
}
type ncxPoint struct {
	ID        string `xml:"id,attr"`
	PlayOrder int    `xml:"playOrder,attr"`
	NavLabel  struct {
		Text string `xml:"text"`
	} `xml:"navLabel"`
	Content struct {
		Src string `xml:"src,attr"`
	} `xml:"content"`
	Points []ncxPoint `xml:"navPoint"`
}

func TestNCX(t *testing.T) {
	b := sampleBook()
	zr := openZip(t, writeToBytes(t, b))
	var nc ncxDoc
	if err := xml.Unmarshal(readEntry(t, zr, "OEBPS/toc.ncx"), &nc); err != nil {
		t.Fatalf("parse toc.ncx: %v", err)
	}
	var uid string
	for _, m := range nc.Head.Metas {
		if m.Name == "dtb:uid" {
			uid = m.Content
		}
	}
	if uid != b.Identifier {
		t.Errorf("ncx dtb:uid = %q, want %q", uid, b.Identifier)
	}
	if len(nc.NavMap.Points) != 3 {
		t.Fatalf("ncx top navPoints = %d, want 3", len(nc.NavMap.Points))
	}
	if len(nc.NavMap.Points[1].Points) != 2 {
		t.Errorf("category A navPoint children = %d, want 2", len(nc.NavMap.Points[1].Points))
	}
}

// flattenNCX returns navPoints in document (depth-first pre-order) order.
func flattenNCX(pts []ncxPoint) []ncxPoint {
	var out []ncxPoint
	for _, p := range pts {
		out = append(out, p)
		out = append(out, flattenNCX(p.Points)...)
	}
	return out
}

// TestNCXSequential enforces unique navPoint ids and a gapless playOrder
// running 1..N in document order over the whole tree (epubcheck rejects
// duplicate ids and playOrder gaps).
func TestNCXSequential(t *testing.T) {
	zr := openZip(t, writeToBytes(t, sampleBook()))
	var nc ncxDoc
	if err := xml.Unmarshal(readEntry(t, zr, "OEBPS/toc.ncx"), &nc); err != nil {
		t.Fatalf("parse toc.ncx: %v", err)
	}
	flat := flattenNCX(nc.NavMap.Points)
	if len(flat) != 6 { // intro + (cat-a, a1, a2) + (cat-b, b1)
		t.Fatalf("total navPoints = %d, want 6", len(flat))
	}
	ids := map[string]bool{}
	for i, p := range flat {
		if p.PlayOrder != i+1 {
			t.Errorf("navPoint %d (%q) playOrder = %d, want %d", i, p.Content.Src, p.PlayOrder, i+1)
		}
		if p.ID == "" || ids[p.ID] {
			t.Errorf("navPoint id %q is empty or duplicated", p.ID)
		}
		ids[p.ID] = true
	}
}

// TestSVGProperty enforces EPUB 3 OPF-014: a content document containing SVG
// must carry properties="svg" on its manifest item, and one without must not.
func TestSVGProperty(t *testing.T) {
	b := &Book{
		Title: "T", Author: "A", Language: "fr",
		Identifier: "urn:uuid:svg", Modified: testModified, CSS: "x",
		Chapters: []Chapter{
			{FileName: "plain.xhtml", Title: "Plain", Body: "<p>rien</p>"},
			{FileName: "withsvg.xhtml", Title: "Fig", Body: `<figure><svg viewBox="0 0 4 4"><rect x="0" y="0" width="4" height="4"/></svg></figure>`},
		},
	}
	zr := openZip(t, writeToBytes(t, b))
	p := parseOPF(t, zr)
	props := map[string]string{}
	for _, it := range p.Manifest.Items {
		props[it.Href] = it.Properties
	}
	if props["withsvg.xhtml"] != "svg" {
		t.Errorf("withsvg.xhtml properties = %q, want svg", props["withsvg.xhtml"])
	}
	if props["plain.xhtml"] != "" {
		t.Errorf("plain.xhtml properties = %q, want empty", props["plain.xhtml"])
	}
}

func TestChapterShell(t *testing.T) {
	zr := openZip(t, writeToBytes(t, sampleBook()))
	doc := string(readEntry(t, zr, "OEBPS/intro.xhtml"))
	for _, want := range []string{
		`<?xml version="1.0" encoding="utf-8"?>`,
		`<!DOCTYPE html>`,
		`<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">`,
		`<title>Introduction</title>`,
		`href="style.css"`,
		`<p>Bienvenue.</p>`,
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("chapter shell missing %q in:\n%s", want, doc)
		}
	}
	// Well-formed XML.
	if err := xml.Unmarshal([]byte(doc), new(struct {
		XMLName xml.Name
	})); err != nil {
		t.Errorf("chapter not well-formed XML: %v", err)
	}
}

func TestDeterminism(t *testing.T) {
	b := sampleBook()
	b.Cover = []byte("\x89PNG\r\n\x1a\nbody")
	a := writeToBytes(t, b)
	c := writeToBytes(t, b)
	if !bytes.Equal(a, c) {
		t.Errorf("two Writes produced different bytes (%d vs %d)", len(a), len(c))
	}
}

func TestValidationErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Book)
	}{
		{"empty title", func(b *Book) { b.Title = "" }},
		{"empty identifier", func(b *Book) { b.Identifier = "" }},
		{"empty language", func(b *Book) { b.Language = "" }},
		{"no chapters", func(b *Book) { b.Chapters = nil }},
		{"filename not xhtml", func(b *Book) { b.Chapters[0].FileName = "intro.html" }},
		{"duplicate filename", func(b *Book) { b.Chapters[1].Children[0].FileName = "intro.xhtml" }},
		{"nesting too deep", func(b *Book) {
			b.Chapters[1].Children[0].Children = []Chapter{{FileName: "deep.xhtml", Title: "Deep"}}
		}},
		{"cover not png", func(b *Book) { b.Cover = []byte("GIF89a") }},
		{"reserved filename", func(b *Book) { b.Chapters[0].FileName = "nav.xhtml" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := sampleBook()
			c.mutate(b)
			var buf bytes.Buffer
			if err := b.Write(&buf); err == nil {
				t.Errorf("expected error, got nil")
			} else if buf.Len() != 0 {
				t.Errorf("bytes written despite validation error (%d bytes)", buf.Len())
			}
		})
	}
}
