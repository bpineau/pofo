# EPUB export for the embedded books

Status: design approved 2026-07-23; phase 1 (pofo) pending implementation,
phase 2 (locador) follows once phase 1 is validated on-device.

## Goal

Export the embedded books as standard EPUB 3 files so they can be read
offline in any ebook reader (KOReader on Android and desktop is the primary
target), with highlights and annotations made there exportable back as
Markdown for feedback loops. Three books are in scope, in order:

1. `pkg/firebook`: "Le FIRE tranquille" (86 French articles, 23 Go-generated
   SVG figures, 8 callout types, wiki-links).
2. locador's embedded documentation (92 French articles, same Markdown
   dialect minus figures, rendered client-side by `docs.js` today).
3. A future book (same dialect) in another repo; the exporter must be
   reusable as a library.

Non-goals: EPUB 2 output, KF8/MOBI, fidelity to the web "instrument" theme
(reader themes win by design), altering the existing HTML serving paths in
any way.

## Background: how the two books are built today

| | firebook (pofo) | locador docs |
|---|---|---|
| Source | `pkg/firebook/assets/book/fr/*.md` | `server/assets/docs/*.md` |
| Renderer | Go, `pkg/firebook/render.go` (`ToHTML`) | JS, `server/assets/docs.js` (`mdToHtml`), client-side |
| Manifest | Go, `firebook.Categories` | JS, `DOC_CATEGORIES` in `docs.js` |
| Figures | `::: figure <id>` blocks, inline SVG from `figures.go`/`figures_v2.go`, literal hex colors | none |
| Callouts | `encart cle astuce attention exemple science terrain` | `encart cle astuce attention exemple admin` |

The two Markdown dialects are otherwise identical (checked against both
implementations line by line): `##`..`####` headings (`#` demoted to h2),
pipe tables, `-`/`1.` lists with GitHub task items, `>` blockquotes, `---`
rules, `::: type Titre` callout blocks, inline bold/italic/`code`,
`[label](url)` links, `[[slug]]` / `[[slug|label]]` wiki-links. No images,
no code fences, in either corpus. locador's dialect is a strict subset of
firebook's; the `admin` callout is {📋, "Côté administratif"}.

locador already carries third-party Go dependencies; pofo is stdlib-only.
Importing pofo from locador therefore adds zero transitive dependencies
(finador already uses this pattern with `replace => ../pofo`).

## Architecture

Two new pofo packages plus one assembly file per book. Phase 1 is entirely
in pofo; phase 2 wires locador to the same packages.

```
pkg/bookmd    the shared Markdown-dialect renderer (extracted from firebook)
pkg/epub      generic EPUB 3 writer (books in, .epub bytes out)
pkg/firebook  epub.go: assembles "Le FIRE tranquille" from Categories
cmd/pofo      -export-epub flag; firebook.Handler serves the .epub route
locador       (phase 2) Go manifest + guard, epub route, CLI
```

### pkg/bookmd: the extracted renderer

A pure move of `pkg/firebook/render.go` into a neutral package so locador
can import the renderer without embedding the FIRE book's assets. The
public API:

```go
// Callout describes one ::: block type (glyph + default label).
type Callout struct{ Glyph, Label string }

// Callouts is the built-in superset used by all books:
// encart, cle, astuce, attention, exemple, science, terrain, admin.
var Callouts map[string]Callout

// Options tunes rendering. The zero value renders wiki-links as plain
// text and figure blocks as nothing.
type Options struct {
    Titles map[string]string        // written slugs -> display titles
    Href   func(slug string) string // wiki-link target; nil -> slug itself
    Figure func(id string) string   // ::: figure payload; nil -> block dropped
}

// ToHTML renders one article body to HTML.
func ToHTML(src string, opt Options) string
```

Behavior is byte-identical to today's `firebook.ToHTML(src, titles)` when
called with `Options{Titles: titles, Figure: firebook.FigureSVG}`:
`firebook.ToHTML` remains, as that thin wrapper, so all existing callers,
tests and rendered pages are unchanged. The `Href` hook exists because the
EPUB assembly needs `slug.xhtml` targets while the web needs bare `slug`
(nil keeps today's behavior). Unknown callout types still degrade to
`encart`. `render_test.go` moves with the code (adapted to the new
signature); firebook keeps a wrapper-level test proving the delegation
renders one article identically.

### pkg/epub: the generic writer

Stdlib only (`archive/zip`, `encoding/xml`, `fmt`, `strings`, `time`).

```go
type Book struct {
    Title       string
    Author      string
    Language    string    // BCP 47, e.g. "fr"
    Identifier  string    // stable urn:uuid, one per book, hardcoded by the caller
    Modified    time.Time // dcterms:modified; injected for determinism
    Description string
    CSS         string    // single stylesheet applied to every chapter
    Cover       []byte    // optional PNG; nil -> no cover page
    Chapters    []Chapter // reading order; nesting drives the TOC
}

type Chapter struct {
    FileName string    // "combien-il-vous-faut.xhtml"
    Title    string    // TOC + <title>
    Body     string    // XHTML fragment (the <body> content)
    Children []Chapter // one level of nesting (category -> articles)
}

// Write emits the complete EPUB 3 container to w.
func (b *Book) Write(w io.Writer) error
```

`Write` produces the regulatory OCF layout, in this order:

- `mimetype`: `application/epub+zip`, STORED (uncompressed), first entry,
  no extra field (the 38-byte offset rule).
- `META-INF/container.xml` pointing at `OEBPS/content.opf`.
- `OEBPS/content.opf`: `dc:identifier` (unique-identifier), `dc:title`,
  `dc:language`, `dc:creator`, `dc:description`, `dcterms:modified`
  (UTC, `2006-01-02T15:04:05Z`), manifest (chapters, nav, css, cover with
  `properties="cover-image"`), spine in reading order (cover page first
  when present, then nav is NOT in the spine, then chapters
  depth-first).
- `OEBPS/nav.xhtml`: `epub:type="toc"` nav, `<ol>` nested one level to
  mirror `Chapter.Children`.
- `OEBPS/toc.ncx` mirroring the same tree (harmless, helps conversions
  and older software).
- `OEBPS/style.css`, `OEBPS/cover.png` + `OEBPS/cover.xhtml` when present.
- One `OEBPS/<FileName>` per chapter: fixed XHTML5 shell
  (`<?xml version="1.0" encoding="utf-8"?><!DOCTYPE html><html
  xmlns="http://www.w3.org/2000/xhtml" xmlns:epub="...">`) wrapping
  `Body`.

Output is deterministic: fixed entry order, zip headers carry `Modified`,
no other timestamps. Two calls with equal input produce equal bytes (unit
tested), so the HTTP route can hash for ETag and tests can golden the
structure.

Validation errors (empty title/identifier/language, duplicate or
non-`.xhtml` FileName, more than one nesting level) fail `Write` rather
than producing a broken container.

### XHTML normalization

EPUB content documents are XML; the renderer emits HTML5. Rather than
forking the renderer, `pkg/epub` exports one targeted pass:

```go
// Normalize rewrites the finite HTML inventory emitted by pkg/bookmd
// into well-formed XHTML.
func Normalize(html string) string
```

The renderer's output tag inventory is closed (aside/div/span, h2-h4, hr,
table set, ul/ol/li, input, blockquote, p, a, strong/em/code, figure/
figcaption, svg subset), so the pass is small and exact:

- `<hr>` -> `<hr/>`.
- Task-list items: `<input type="checkbox" disabled>` (+ `checked`)
  becomes the text glyph `☐` / `☑` (readers do not render form controls;
  the glyph survives everywhere).
- Everything else is already well-formed (attributes are always quoted,
  SVG elements are self-closed at the source).

A guard test renders every embedded article, normalizes, wraps in the
chapter shell and parses it with `encoding/xml`: any future renderer or
article change that would break XML well-formedness fails `make test`.

### pkg/firebook/epub.go: assembling "Le FIRE tranquille"

`firebook.EPUB(modified time.Time) ([]byte, error)`:

- Chapters: one title page (title, subtitle, edition note), then one
  `Chapter` per `Categories` entry (category page: title, blurb, article
  list) with its articles as `Children`. Article body = `bookmd.ToHTML`
  with `Titles()`, `Href: slug + ".xhtml"`, `Figure: FigureSVG`, then
  `epub.Normalize`. The article's h1 comes from the manifest title (as on
  the web), the in-file `#` line stays dropped.
- Figures: inline SVG, kept verbatim (EPUB 3 supports inline SVG; colors
  are literal hex, fonts are generic families, so the markup is
  self-contained). This is the on-device validation risk; see below.
- CSS: a book stylesheet (`assets/book/epub.css`) that styles callout
  boxes (border + glyph head, `doc-box--<type>` accents), tables,
  blockquotes, figures and task glyphs WITHOUT setting page-wide colors
  or fonts, so reader themes (night mode, user fonts) keep working.
  Backgrounds, when any, are borders or very light tints only.
- Cover: `assets/book/cover.png`, a one-off committed asset (readers
  thumbnail raster covers only; generating text-bearing PNGs at runtime
  is not reasonable stdlib-only). Produced offline from an SVG design in
  the book's plate style, modest size (target <= 200 KB, ~1200x1920).
- Identifier: a fixed `urn:uuid:` constant in `epub.go`; `dc:source`
  points at the public book URL.

### Delivery

- Route: `firebook.Handler` itself serves `GET /le-fire-tranquille.epub`
  (`application/epub+zip`, `Content-Disposition: attachment`), so every
  mount (pofo `-serve` under `/firebook/fr/`, the `-fire` server, finador)
  gets it with no wiring. Built lazily on first request with
  `Modified` = server start, cached in memory (sync.Once; the book is
  embedded, it cannot change while running), ETag from a hash of the
  bytes. The book index page gains a discreet download link ("Version
  EPUB" with file size).
- CLI: `pofo -export-epub [path]` (default `le-fire-tranquille.epub`)
  writes the same bytes with `Modified` = now; own mode file
  `cmd/pofo/epubexport.go` per house convention.

Nothing else in the serving paths changes.

## Phase 2: locador

- `go.mod`: `require github.com/bpineau/pofo` + `replace => ../pofo`
  (the finador pattern). Zero transitive additions since pofo is
  dependency-free.
- Manifest: `DOC_CATEGORIES` lives in JS and must stay there (the web
  renderer is untouched). A Go mirror (`server/docsbook.go`,
  `[]struct{Title, Blurb string; Docs [][3]string}` or equivalent) is
  added for the exporter, plus a guard test that extracts
  `DOC_CATEGORIES` from `docs.js` with a regexp and compares slugs,
  titles and order against the Go mirror, so the two cannot drift
  (locador already guards docs consistency this way, see
  `docs_manifest_guard_test.go`).
- Rendering: `bookmd.ToHTML` with the locador titles map, `Href: slug +
  ".xhtml"`, no `Figure` hook. The `admin` callout is already in
  `bookmd.Callouts`.
- Corpus guard: render all 92 articles through normalize + `encoding/xml`
  (same guard as pofo) to prove the Go renderer covers the whole locador
  corpus before shipping.
- Delivery: same shape as pofo, one route on the locador server (path
  and final book title decided in phase 2; working title "locador : le
  guide de l'investissement locatif") and a CLI/make entry point.
- The client-side JS rendering path is not modified at all.

## Testing and validation

- `pkg/epub` unit tests: mimetype first + STORED, container/opf/nav/ncx
  parse with `encoding/xml`, spine matches chapter order, determinism
  (two writes, equal bytes), validation errors.
- Well-formedness guard over all 86 firebook articles (and later the 92
  locador articles), as described above.
- `firebook` tests: EPUB builds, contains one file per article plus
  title/category/nav pages; handler test for the route (status, MIME,
  ETag, non-empty, cached identity across requests).
- Runnable examples in each new package's `example_test.go`.
- `make check` green; no network anywhere.
- Optional local step (not in `make test`): `epubcheck` run against the
  generated file when the tool is installed.
- On-device validation gate between phases: load the generated EPUB in
  KOReader (Android) and check figures (crengine's SVG support is the
  known risk), callouts, tables, task glyphs, TOC nesting, wiki-links,
  night mode. If SVG rendering is unacceptable, the fallback is
  pre-rasterized PNG figures committed as assets (generated offline via
  the existing headless-Chrome tooling) and selected by the assembly
  step; the design isolates this in `Figure`/assembly so nothing else
  moves.

## Documentation

`pkg/bookmd` and `pkg/epub` get full `doc.go` + examples; `firebook`'s
`doc.go` gains the export paragraph; CLAUDE.md's map and README's CLI
section mention `-export-epub` and the route; this file is indexed in
`docs/README.md`.
