// Package firebook embeds and serves the FIRE book: a French-language
// handbook of retirement decumulation (safe withdrawal rates, sequence risk,
// withdrawal strategies, resilient portfolios, buffers, French taxation, the
// human side), written as standalone cross-linked articles.
//
// The articles live under assets/book/fr/<slug>.md in a small Markdown dialect
// shared with locador's embedded documentation: ## / ### headings, pipe
// tables, lists, blockquotes, [[slug]] wiki-links and ::: callout blocks
// (cle, astuce, attention, exemple, encart, science, terrain). That dialect
// is rendered by the neutral pkg/bookmd package; firebook.ToHTML is a thin
// wrapper over it wired with the book's figure generator (FigureSVG). Handler
// serves the whole book as self-contained HTML pages
// (index plus one page per article) styled with the shared pkg/webui identity.
//
// The table of contents is data (Categories); the index page and the
// navigation are generated from it, so adding an article means adding its
// .md file and one manifest line. Wiki-links may point at planned but not yet
// written articles (the full plan is docs/fire-book-design.md); those render
// as plain text until the target exists, and a guard test keeps files,
// manifest and links consistent.
//
// The pofo -fire web UI mounts Handler under /livre/; any other server (for
// example finador) can mount the same book by importing this package.
//
// EPUB(modified) assembles the whole book as a standard EPUB 3 file (via
// pkg/epub, styled with the theme-neutral assets/book/epub.css): a title page,
// one page per category with its articles nested beneath it in the table of
// contents, and every article rendered through bookmd with wiki-links pointing
// at "<slug>.xhtml" and figures kept as inline SVG. The bytes are deterministic
// for a given modified time, so a serving route can hash them for an ETag.
package firebook
