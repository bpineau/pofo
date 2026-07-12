// Package firebook embeds and serves the FIRE book: a French-language
// handbook of retirement decumulation (safe withdrawal rates, sequence risk,
// withdrawal strategies, resilient portfolios, buffers, French taxation, the
// human side), written as standalone cross-linked articles.
//
// The articles live under assets/book/fr/<slug>.md in a small Markdown dialect
// shared with locador's embedded documentation: ## / ### headings, pipe
// tables, lists, blockquotes, [[slug]] wiki-links and ::: callout blocks
// (cle, astuce, attention, exemple, encart, science, terrain). ToHTML renders
// that dialect; Handler serves the whole book as self-contained HTML pages
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
package firebook
