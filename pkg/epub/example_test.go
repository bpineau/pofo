package epub_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/epub"
)

// Write assembles a tiny two-chapter book and emits a standard EPUB container.
// Here the zip's file list is printed to show the OCF layout.
func ExampleBook_Write() {
	book := &epub.Book{
		Title:      "Petit guide",
		Author:     "Anne Auteur",
		Language:   "fr",
		Identifier: "urn:uuid:00000000-0000-0000-0000-000000000000",
		Modified:   time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC),
		CSS:        "body{font-family:serif}",
		Chapters: []epub.Chapter{
			{FileName: "intro.xhtml", Title: "Introduction", Body: "<p>Bienvenue.</p>"},
			{FileName: "fin.xhtml", Title: "Conclusion", Body: epub.Normalize("<p>La fin.</p><hr>")},
		},
	}

	var buf bytes.Buffer
	if err := book.Write(&buf); err != nil {
		fmt.Println("error:", err)
		return
	}

	zr, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	for _, f := range zr.File {
		fmt.Println(f.Name)
	}

	// Output:
	// mimetype
	// META-INF/container.xml
	// OEBPS/content.opf
	// OEBPS/nav.xhtml
	// OEBPS/toc.ncx
	// OEBPS/style.css
	// OEBPS/intro.xhtml
	// OEBPS/fin.xhtml
}

// Normalize turns the HTML a book renderer emits into well-formed XHTML.
func ExampleNormalize() {
	html := `<ul><li class="task"><input type="checkbox" disabled checked> fait</li>` +
		`<li class="task"><input type="checkbox" disabled> à faire</li></ul><hr>`
	fmt.Println(epub.Normalize(html))
	// Output: <ul><li class="task">☑ fait</li><li class="task">☐ à faire</li></ul><hr/>
}
