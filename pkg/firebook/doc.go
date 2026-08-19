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
// Every heading carries a hover-revealed "§" anchor that copies a direct link
// to that section, on article pages and on the index alike (there the part
// titles get it, so "Les stratégies de retrait" can be pointed at directly,
// and an article's top bar links back to its own part). The ids come from one
// rule, bookmd.HeadingID, so the index and the articles cannot drift apart.
// It is web chrome only: the EPUB export never carries it.
//
// The pofo web surfaces mount the book under /firebook/<lang>/; any other
// server (for example finador) can mount it by importing this package.
//
// # Editions
//
// One Edition value carries everything a language needs: its manifest, its
// chrome strings, its EPUB identity and its figure renderer. French is the
// source edition and the package-level API is a thin wrapper over it, so
// firebook.Handler() and firebook.EPUB(t) keep meaning "the French book":
//
//	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr", firebook.French.Handler()))
//	blob, err := firebook.French.EPUB(time.Now())
//
// All writing, correction and upkeep happens in French first. A translated
// edition pairs each of its articles with its French original (Article.Source
// plus an in-file source stamp, dropped at render time) and Drift reports what
// the translation owes: the articles whose original moved since, then the ones
// nothing covers yet, then the ones the French edition keeps for itself, which
// say so with an in-file "<!-- edition: fr-only -->" marker (dropped at render
// time too). "pofo -book-drift" prints that worklist, naming for each
// untranslated article the English slug planned for it.
//
// English is the second edition, "The Quiet FIRE": its own translated slugs,
// its own EPUB identity, English chrome and callout labels, and the French tax
// part replaced by a US-framework one. Its figures are not duplicated: the
// plate generators stay French and single-source, and FigureSVGEnglish
// translates the rendered SVG text through a dictionary (see figures_i18n.go;
// scripts/figure-audit.sh checks that no translated label runs off its plate).
// It is mounted at /firebook/en/ and is complete: every French article that is
// not marked fr-only has a counterpart there, and a guard test says so.
//
// Two mounted editions cross-link with WithAlternate, which each mount hands
// the OTHER one's base path (the handler emits relative URLs, so it cannot
// know where its sibling sits):
//
//	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr",
//		firebook.Handler(firebook.WithAlternate("/firebook/en/", firebook.English))))
//
// Every paired page then declares both languages with rel="alternate"
// hreflang links and offers a switch to its counterpart; a page the sibling
// does not carry declares nothing. The design is
// docs/fire-book-en-edition-design.md.
//
// EPUB(modified) assembles the whole book as a standard EPUB 3 file (via
// pkg/epub, styled with the theme-neutral assets/book/epub.css): a title page,
// one page per category with its articles nested beneath it in the table of
// contents, and every article rendered through bookmd with wiki-links pointing
// at "<slug>.xhtml" and figures kept as inline SVG. The bytes are deterministic
// for a given modified time, so a serving route can hash them for an ETag.
package firebook
