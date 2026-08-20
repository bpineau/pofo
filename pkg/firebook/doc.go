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
// serves the whole book as self-contained HTML pages (index plus one page per
// article, each also available as its Markdown source at "<slug>.md") styled
// with the shared pkg/webui identity.
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
// hreflang links (plus an x-default naming the source edition) and offers a
// switch to its counterpart; a page the sibling does not carry declares
// nothing. The design is docs/fire-book-en-edition-design.md.
//
// # Being found, and being quoted
//
// The book is published to be read, indexed and cited, so every page states
// what it is in the two vocabularies that matter.
//
// For search engines: a per-page title and meta description (the manifest
// blurb), Open Graph and Twitter-card metas, a self-referencing canonical link
// and og:url built from Edition.HomePath (the same book is reachable under more
// than one prefix on a pofo server, and only HomePath says which copy counts),
// the hreflang pair and its x-default, and schema.org JSON-LD: a WebSite plus
// the Book (its EPUB as a workExample, its whole table of contents as Chapter
// parts) on the index, an Article and its BreadcrumbList on every page.
//
// Shared rather than searched, a link to the book shows its social card: one
// image per edition at "<mount>/card.png", 1200x630, declared as og:image with
// a summary_large_image Twitter card. The drawing is Go code like every other
// plate ((*Edition).CardSVG, the book's own hero block at card size in the v2
// plate identity), and every word on it comes from the Edition value already:
// nothing is written for the card. The committed PNG next to it is that SVG
// rasterized once by scripts/card-shot.sh, since a crawler wants a bitmap.
//
// Those head URLs are the handler's one departure from the relative,
// mount-anywhere URLs it emits everywhere else: they are FULLY QUALIFIED, the
// request's own origin (RequestOrigin) in front of the HomePath-derived path.
// An hreflang annotation whose ends are not complete URLs is ignored outright,
// and a canonical link should agree with the hreflang next to it. The visible
// language switch in the top bar stays a path, so a visitor never leaves the
// host they reached.
//
// For AI agents: every article is also served as its own Markdown source at
// "<slug>.md", untransformed (callouts, tables and [[wiki-links]] included),
// with one HTML-comment provenance line naming the page to cite; the HTML page
// declares that mirror with rel="alternate" type="text/markdown" and in its
// structured data. Site (BookSite) renders the three files that index all of
// it, for a server to mount at its root:
//
//	firebook.BookSite(firebook.Page{Path: "/", Title: "pofo"}).Handle(mux)
//
// which serves /sitemap.xml (every article, index and EPUB of every mounted
// edition, no invented lastmod), /robots.txt (everything allowed, the AI
// crawlers named one by one, the sitemap declared) and /llms.txt (the
// llmstxt.org index of the whole site, both editions in one file, every
// article linked as its Markdown mirror). The formats themselves live in
// pkg/seo.
//
// For readers who follow rather than search, each edition also publishes an
// Atom feed at "<mount>/feed.xml" (FeedXML), one entry per article in reading
// order, declared from every page's head. Its <updated> is the mount's single
// publication stamp, the same time the EPUB writes into dcterms:modified, on
// the feed and on every entry alike: the articles are embedded in the binary
// and none of them carries an honest date of its own.
//
// # Adding a plate
//
// A figure is a Go function returning inline SVG, so adding one touches a
// handful of files in a fixed order.
//
//  1. Write the plate in figures_<theme>.go: the closed model as pure
//     functions (or the numbers read off pkg/datasets, frozen into arrays),
//     then the renderer over them. The palette, the type primitives and the
//     shared utilities are in figures_kit.go; the FORM itself is invented per
//     plate and never factored out.
//  2. Register the slug in the figures map of figures.go.
//  3. Write the guard test in figures_<theme>_test.go: numbers taken from a
//     dataset are recomputed under frozenAgainstData (they are allowed to lag
//     a refresh, "make figure-drift" reports what they owe); a closed model is
//     confronted with the numbers its article's prose quotes; and the rendered
//     SVG is checked for the labels and readings the plate claims to draw.
//  4. Insert "::: figure <slug>" plus a caption in the French article AND in
//     its English counterpart, add a figureDict entry for every text node of
//     the plate (in figures_i18n_data.go), and refresh the English article's
//     source stamp so "make book-drift" stays clean.
//  5. Check it: go test ./pkg/firebook, scripts/figure-audit.sh fr and en for
//     labels running off the plate, and scripts/figure-shot.sh <slug> to look
//     at the thing with your own eyes, which no test replaces.
//
// The hard rules, all of them enforced by tests somewhere: labels stay
// horizontal, fills are opaque hex and never rgba (crengine, KOReader's EPUB
// renderer, paints rgba solid black), no em- or en-dash anywhere, decimal
// commas in French and points in English, and English money amounts convert
// one for one (500 k EUR becomes $500k).
//
// EPUB(modified) assembles the whole book as a standard EPUB 3 file (via
// pkg/epub, styled with the theme-neutral assets/book/epub.css): a title page,
// one page per category with its articles nested beneath it in the table of
// contents, and every article rendered through bookmd with wiki-links pointing
// at "<slug>.xhtml" and figures kept as inline SVG. The bytes are deterministic
// for a given modified time, so a serving route can hash them for an ETag.
package firebook
