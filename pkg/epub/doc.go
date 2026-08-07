// Package epub is a generic, dependency-free EPUB 3 container writer: books in,
// standard .epub bytes out. It knows nothing about any specific book; a caller
// (for example pofo's firebook) assembles a Book from its own content and hands
// it to Write.
//
// # OCF layout
//
// Write emits the regulatory Open Container Format layout, in a fixed order:
//
//	mimetype                 stored (uncompressed), first, no extra field
//	META-INF/container.xml   points at OEBPS/content.opf
//	OEBPS/content.opf        the EPUB 3 package: metadata, manifest, spine
//	OEBPS/nav.xhtml          the EPUB 3 navigation document (epub:type="toc")
//	OEBPS/toc.ncx            the legacy NCX, mirroring the same tree
//	OEBPS/style.css          Book.CSS, applied to every chapter
//	OEBPS/cover.png          Book.Cover, when present
//	OEBPS/cover.xhtml        cover page, when a cover is present
//	OEBPS/<FileName>         one content document per chapter
//
// Chapters nest exactly one level (a category page and its articles). The nav
// document and NCX mirror that nesting; the spine lists the cover page first
// (when present), then the chapters depth first (a parent page before its
// children). The navigation document is in the manifest but never in the spine.
//
// # Determinism
//
// Output is byte-for-byte deterministic: entry order is fixed and every zip
// timestamp is Book.Modified. Two Writes of the same Book produce equal bytes,
// so a caller can hash the result for an HTTP ETag or golden a structure.
//
// # XHTML normalization
//
// EPUB content documents are parsed as XML, but book renderers (pofo's
// pkg/bookmd) emit HTML5. Normalize rewrites that finite tag inventory into
// well-formed XHTML (self-closing <hr>, task-list checkboxes into text glyphs);
// callers run it over each chapter body before building the Book.
package epub
