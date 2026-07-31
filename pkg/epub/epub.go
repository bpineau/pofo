package epub

import (
	"archive/zip"
	"fmt"
	"hash/crc32"
	"html"
	"io"
	"strings"
	"text/template"
	"time"
)

// Book is a complete EPUB 3 publication: metadata, a single stylesheet applied
// to every chapter, an optional PNG cover and a tree of chapters (one level of
// nesting drives the table of contents).
type Book struct {
	Title       string
	Author      string
	Language    string // BCP 47, e.g. "fr"
	Identifier  string // stable urn:uuid, one per book
	Description string
	Modified    time.Time // dcterms:modified and every zip timestamp; UTC
	CSS         string    // single stylesheet applied to every chapter
	Cover       []byte    // optional PNG; nil means no cover page
	Chapters    []Chapter // reading order; nesting drives the TOC
}

// Chapter is one content document. Children nest exactly one level deep
// (a category page and its articles); a Child may not itself have Children.
type Chapter struct {
	FileName string // e.g. "combien-il-vous-faut.xhtml"; must be unique across the tree
	Title    string // TOC label and <title>
	Body     string // XHTML fragment placed inside <body>
	Children []Chapter
}

// Fixed OCF locations. Content documents live under OEBPS/.
const (
	mimetypeContent = "application/epub+zip"
	opfPath         = "OEBPS/content.opf"
	pubID           = "pub-id" // the unique-identifier reference in the OPF
	ncxID           = "ncx"
)

// Write emits the complete EPUB 3 container to w. It validates the book first
// and writes nothing on error, so a caller never gets a half-formed file. The
// output is deterministic: entry order is fixed and every zip timestamp is
// Book.Modified, so two Writes of the same Book produce equal bytes.
func (b *Book) Write(w io.Writer) error {
	if err := b.validate(); err != nil {
		return err
	}

	zw := zip.NewWriter(w)

	// The mimetype entry must come first, stored (uncompressed) and with no
	// extra field, so its name sits at byte offset 30. CreateRaw writes the
	// header verbatim (no data descriptor, no injected extra field).
	if err := b.writeMimetype(zw); err != nil {
		return err
	}

	flat := b.flatten() // depth-first chapter list, one id each

	// Entry order is fixed for determinism: OCF machinery, then the cover
	// (image before its page), then the chapters in spine order.
	entries := []entry{
		{"META-INF/container.xml", []byte(containerXML)},
		{opfPath, []byte(b.renderOPF(flat))},
		{"OEBPS/nav.xhtml", []byte(b.renderNav())},
		{"OEBPS/toc.ncx", []byte(b.renderNCX())},
		{"OEBPS/style.css", []byte(b.CSS)},
	}
	if b.Cover != nil {
		entries = append(entries,
			entry{"OEBPS/cover.png", b.Cover},
			entry{"OEBPS/cover.xhtml", []byte(b.renderCover())},
		)
	}
	for _, c := range flat {
		entries = append(entries, entry{"OEBPS/" + c.FileName, []byte(b.renderChapter(c.Chapter))})
	}

	for _, e := range entries {
		if err := b.writeDeflated(zw, e.name, e.data); err != nil {
			return err
		}
	}

	return zw.Close()
}

// entry is one zip member: its OCF path and its raw bytes.
type entry struct {
	name string
	data []byte
}

func (b *Book) writeMimetype(zw *zip.Writer) error {
	data := []byte(mimetypeContent)
	fh := &zip.FileHeader{Name: "mimetype", Method: zip.Store, Modified: b.Modified}
	fh.CRC32 = crc32.ChecksumIEEE(data)
	fh.CompressedSize64 = uint64(len(data))
	fh.UncompressedSize64 = uint64(len(data))
	fw, err := zw.CreateRaw(fh)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

func (b *Book) writeDeflated(zw *zip.Writer, name string, data []byte) error {
	fw, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: b.Modified})
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

// flatChapter pairs a chapter with the manifest/spine id assigned to it.
type flatChapter struct {
	Chapter
	ID string
}

// flatten returns the chapters in spine order (parent before its children)
// with a stable id per document.
func (b *Book) flatten() []flatChapter {
	var out []flatChapter
	n := 0
	var walk func(cs []Chapter)
	walk = func(cs []Chapter) {
		for _, c := range cs {
			out = append(out, flatChapter{Chapter: c, ID: fmt.Sprintf("c%d", n)})
			n++
			walk(c.Children)
		}
	}
	walk(b.Chapters)
	return out
}

func (b *Book) validate() error {
	switch {
	case b.Title == "":
		return fmt.Errorf("epub: empty Title")
	case b.Identifier == "":
		return fmt.Errorf("epub: empty Identifier")
	case b.Language == "":
		return fmt.Errorf("epub: empty Language")
	case len(b.Chapters) == 0:
		return fmt.Errorf("epub: no chapters")
	}
	if b.Cover != nil && !strings.HasPrefix(string(b.Cover), "\x89PNG") {
		return fmt.Errorf("epub: Cover is not a PNG (missing \\x89PNG magic)")
	}
	seen := map[string]bool{}
	var check func(cs []Chapter, depth int) error
	check = func(cs []Chapter, depth int) error {
		for _, c := range cs {
			if depth > 1 {
				return fmt.Errorf("epub: chapter %q nests deeper than one level", c.FileName)
			}
			if !strings.HasSuffix(c.FileName, ".xhtml") {
				return fmt.Errorf("epub: chapter FileName %q does not end in .xhtml", c.FileName)
			}
			if seen[c.FileName] {
				return fmt.Errorf("epub: duplicate chapter FileName %q", c.FileName)
			}
			seen[c.FileName] = true
			if err := check(c.Children, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	return check(b.Chapters, 0)
}

// esc escapes text for XML character data and double-quoted attributes.
// html.EscapeString covers & < > " ' with valid XML entities.
func esc(s string) string { return html.EscapeString(s) }

var tmplFuncs = template.FuncMap{"esc": esc}

const containerXML = `<?xml version="1.0" encoding="utf-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
`

// opfData feeds the content.opf template.
type opfData struct {
	Book     *Book
	Modified string
	Items    []manifestItem
	Spine    []string // idrefs, in reading order
}

type manifestItem struct {
	ID, Href, MediaType, Properties string
}

func (b *Book) renderOPF(flat []flatChapter) string {
	items := []manifestItem{
		{ID: "nav", Href: "nav.xhtml", MediaType: "application/xhtml+xml", Properties: "nav"},
		{ID: ncxID, Href: "toc.ncx", MediaType: "application/x-dtbncx+xml"},
		{ID: "css", Href: "style.css", MediaType: "text/css"},
	}
	var spine []string
	if b.Cover != nil {
		items = append(items,
			manifestItem{ID: "cover-image", Href: "cover.png", MediaType: "image/png", Properties: "cover-image"},
			manifestItem{ID: "cover", Href: "cover.xhtml", MediaType: "application/xhtml+xml"},
		)
		spine = append(spine, "cover")
	}
	for _, c := range flat {
		items = append(items, manifestItem{ID: c.ID, Href: c.FileName, MediaType: "application/xhtml+xml"})
		spine = append(spine, c.ID)
	}

	return render(opfTmpl, opfData{
		Book:     b,
		Modified: b.Modified.UTC().Format("2006-01-02T15:04:05Z"),
		Items:    items,
		Spine:    spine,
	})
}

var opfTmpl = template.Must(template.New("opf").Funcs(tmplFuncs).Parse(
	`<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="` + pubID + `">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="` + pubID + `">{{esc .Book.Identifier}}</dc:identifier>
    <dc:title>{{esc .Book.Title}}</dc:title>
    <dc:language>{{esc .Book.Language}}</dc:language>
{{- if .Book.Author}}
    <dc:creator>{{esc .Book.Author}}</dc:creator>
{{- end}}
{{- if .Book.Description}}
    <dc:description>{{esc .Book.Description}}</dc:description>
{{- end}}
    <meta property="dcterms:modified">{{.Modified}}</meta>
  </metadata>
  <manifest>
{{- range .Items}}
    <item id="{{esc .ID}}" href="{{esc .Href}}" media-type="{{.MediaType}}"{{if .Properties}} properties="{{.Properties}}"{{end}}/>
{{- end}}
  </manifest>
  <spine toc="` + ncxID + `">
{{- range .Spine}}
    <itemref idref="{{esc .}}"/>
{{- end}}
  </spine>
</package>
`))

// navData / navEntry feed both nav.xhtml and toc.ncx.
type navEntry struct {
	Href, Title string
	PlayOrder   int
	Children    []navEntry
}

func (b *Book) navTree() []navEntry {
	order := 0
	var build func(cs []Chapter) []navEntry
	build = func(cs []Chapter) []navEntry {
		var out []navEntry
		for _, c := range cs {
			order++
			out = append(out, navEntry{
				Href:      c.FileName,
				Title:     c.Title,
				PlayOrder: order,
				Children:  build(c.Children),
			})
		}
		return out
	}
	return build(b.Chapters)
}

func (b *Book) renderNav() string {
	return render(navTmpl, struct {
		Title   string
		Entries []navEntry
	}{Title: b.Title, Entries: b.navTree()})
}

var navTmpl = template.Must(template.New("nav").Funcs(tmplFuncs).Parse(
	`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/2000/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<meta charset="utf-8"/>
<title>{{esc .Title}}</title>
<link rel="stylesheet" type="text/css" href="style.css"/>
</head>
<body>
<nav epub:type="toc" id="toc">
{{- template "navol" .Entries}}
</nav>
</body>
</html>
{{define "navol"}}<ol>
{{- range .}}
<li><a href="{{esc .Href}}">{{esc .Title}}</a>{{if .Children}}{{template "navol" .Children}}{{end}}</li>
{{- end}}
</ol>{{end}}`))

func (b *Book) renderNCX() string {
	return render(ncxTmpl, struct {
		UID, Title string
		Entries    []navEntry
	}{UID: b.Identifier, Title: b.Title, Entries: b.navTree()})
}

var ncxTmpl = template.Must(template.New("ncx").Funcs(tmplFuncs).Parse(
	`<?xml version="1.0" encoding="utf-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
<head>
<meta name="dtb:uid" content="{{esc .UID}}"/>
<meta name="dtb:depth" content="2"/>
<meta name="dtb:totalPageCount" content="0"/>
<meta name="dtb:maxPageNumber" content="0"/>
</head>
<docTitle><text>{{esc .Title}}</text></docTitle>
<navMap>
{{- template "navpoints" .Entries}}
</navMap>
</ncx>
{{define "navpoints"}}{{range .}}
<navPoint id="np{{.PlayOrder}}" playOrder="{{.PlayOrder}}">
<navLabel><text>{{esc .Title}}</text></navLabel>
<content src="{{esc .Href}}"/>
{{- if .Children}}{{template "navpoints" .Children}}{{end}}
</navPoint>
{{- end}}{{end}}`))

func (b *Book) renderChapter(c Chapter) string {
	return render(chapterTmpl, struct {
		Title, Body string
	}{Title: c.Title, Body: c.Body})
}

var chapterTmpl = template.Must(template.New("chapter").Funcs(tmplFuncs).Parse(
	`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/2000/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<meta charset="utf-8"/>
<title>{{esc .Title}}</title>
<link rel="stylesheet" type="text/css" href="style.css"/>
</head>
<body>
{{.Body}}
</body>
</html>
`))

func (b *Book) renderCover() string {
	return render(coverTmpl, struct {
		Title string
	}{Title: b.Title})
}

var coverTmpl = template.Must(template.New("cover").Funcs(tmplFuncs).Parse(
	`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/2000/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<meta charset="utf-8"/>
<title>{{esc .Title}}</title>
<link rel="stylesheet" type="text/css" href="style.css"/>
</head>
<body>
<section epub:type="cover" class="cover">
<img src="cover.png" alt="{{esc .Title}}"/>
</section>
</body>
</html>
`))

// render executes a template into a string; templates are static and validated
// at init, so execution cannot fail for these value types.
func render(t *template.Template, data any) string {
	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		panic(err) // impossible: the templates are fixed and the data is plain
	}
	return sb.String()
}
