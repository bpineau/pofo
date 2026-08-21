package firebook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/bookmd"

	"github.com/bpineau/pofo/pkg/opds"
	"github.com/bpineau/pofo/pkg/seo"
	"github.com/bpineau/pofo/pkg/webui"
)

// NavLink is one entry of the optional site navigation bar.
type NavLink struct{ Label, Href string }

// handlerConfig collects Handler options.
type handlerConfig struct {
	nav     []NavLink
	home    string
	altBase string   // sibling edition's mount path, trailing slash included
	alt     *Edition // the sibling edition itself
}

// Option configures Handler.
type Option func(*handlerConfig)

// WithNav prepends a slim, constant navigation bar to every page: a
// "Sommaire" link back to the book index, then the given links. The bar
// is chrome, not content: articles are untouched, and the bar disappears
// in print. Without this option the book renders exactly as before, which
// the offline and -fire mounts rely on.
func WithNav(links []NavLink) Option { return func(c *handlerConfig) { c.nav = links } }

// WithAlternate cross-links this edition with its sibling ed, published at
// base (an absolute path, e.g. "/firebook/en/"): every page that has a
// counterpart there gains a <link rel="alternate" hreflang="..."> pair in its
// head and a discreet language-switch link at the end of the top bar, named in
// the SIBLING's language ("English version" on a French page). The index
// always pairs with the sibling index; an article pairs through
// Article.Source, so an article the sibling does not carry (the French tax
// part) is left alone rather than pointed at a 404.
//
// The base path is the caller's to give: the handler emits relative URLs so it
// can be mounted anywhere, and only the caller knows where the sibling sits.
// The hreflang links are the one exception, and deliberately: they are built
// from the edition's declared home (Edition.HomePath) plus the request's own
// origin, because an hreflang pair must be made of canonical, fully-qualified
// URLs. The visible switch link in the top bar stays a path. An x-default
// completes the pair and points at the source edition; see alternateHead.
func WithAlternate(base string, ed *Edition) Option {
	return func(c *handlerConfig) {
		if base == "" || ed == nil {
			return
		}
		c.altBase, c.alt = strings.TrimSuffix(base, "/")+"/", ed
	}
}

// WithHome turns the "pofo" kicker above the index title into a link to href,
// rendered as the site's two-tone wordmark (po in ink, fo in the accent).
// Only the -serve constellation passes it: on a standalone mount the kicker
// stays the plain text it always was.
func WithHome(href string) Option { return func(c *handlerConfig) { c.home = href } }

// Handler serves the edition: the index at "/", one HTML page per article at
// "/<slug>" (and its Markdown source at "/<slug>.md"), the EPUB with its OPDS
// catalog, the Atom feed at "/feed.xml", and the shared identity stylesheets at
// "/theme.css" and "/fonts.css". Every URL it emits is relative, bar the head's
// canonical, hreflang and social-card URLs (see absolute), so the handler can
// be mounted under any prefix, e.g.
// http.StripPrefix("/firebook/fr", firebook.French.Handler()).
func (e *Edition) Handler(opts ...Option) http.Handler {
	var cfg handlerConfig
	for _, o := range opts {
		o(&cfg)
	}
	mux := http.NewServeMux()
	css := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.HandleFunc("/theme.css", css(webui.CSS))
	mux.HandleFunc("/fonts.css", css(webui.FontsCSS))

	// built is this mount's single publication stamp, taken once when the
	// handler is assembled. The book is embedded, so it cannot change while
	// the process runs, and no article carries an honest date of its own: one
	// stamp for the whole mount is the only truthful answer, and every format
	// that must name a time (the EPUB's dcterms:modified, the OPDS catalog,
	// the Atom feed and all of its entries) says the same thing.
	built := time.Now()

	// The whole book as a single EPUB, built once on first demand and cached
	// with a strong ETag (deterministic for a given built, hence stable for the
	// life of the process). build() is shared by the download route, the OPDS
	// catalog and the index page, whose download link shows the file size.
	var (
		once sync.Once
		body []byte
		etag string
		bErr error
	)
	build := func() ([]byte, string, error) {
		once.Do(func() {
			body, bErr = e.EPUB(built)
			if bErr != nil {
				log.Printf("firebook: EPUB build failed: %v", bErr)
				return
			}
			sum := sha256.Sum256(body)
			etag = `"` + hex.EncodeToString(sum[:16]) + `"`
		})
		return body, etag, bErr
	}
	mux.HandleFunc("/"+e.EPUBFileName, func(w http.ResponseWriter, r *http.Request) {
		data, tag, err := build()
		if err != nil {
			http.Error(w, e.UI.EPUBUnavailable, http.StatusInternalServerError)
			return
		}
		w.Header().Set("ETag", tag)
		if r.Header.Get("If-None-Match") == tag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", "application/epub+zip")
		w.Header().Set("Content-Disposition", `attachment; filename="`+e.EPUBFileName+`"`)
		_, _ = w.Write(data)
	})

	// An OPDS 1.2 acquisition catalog with the single book as its only entry.
	// The acquisition link is the relative epub path, so a reader (KOReader)
	// that added this feed downloads and later refreshes the same file, which
	// preserves its annotation sidecar. Feed and entry share the book build
	// time, so the feed changes only when the embedded book does.
	mux.HandleFunc("/opds.xml", func(w http.ResponseWriter, r *http.Request) {
		data, _, err := build()
		if err != nil {
			http.Error(w, e.UI.CatalogUnavailable, http.StatusInternalServerError)
			return
		}
		feed := &opds.Feed{
			Title:   e.SiteName,
			ID:      e.EPUBIdentifier + ":catalog",
			Updated: built,
			Self:    "opds.xml",
			Entries: []opds.Entry{{
				Title:   e.SiteName,
				Author:  "pofo",
				Summary: e.SiteLede,
				ID:      e.EPUBIdentifier,
				Updated: built,
				Href:    e.EPUBFileName,
				Type:    "application/epub+zip",
				Size:    int64(len(data)),
			}},
		}
		w.Header().Set("Content-Type", opds.FeedType)
		_, _ = w.Write(feed.XML())
	})

	// The social card (og:image), embedded in the binary and immutable: one
	// strong ETag computed at mount time, and a long freshness window, since
	// the URL only changes bytes when the binary does.
	card := e.CardPNG()
	cardTag := `"` + hex.EncodeToString(sha256Prefix(card)) + `"`
	mux.HandleFunc("/"+CardFileName, func(w http.ResponseWriter, r *http.Request) {
		if len(card) == 0 {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("ETag", cardTag)
		if r.Header.Get("If-None-Match") == cardTag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(card)
	})

	// The Atom syndication feed: one entry per article, so a reader can follow
	// the book from a feed reader rather than by revisiting the index. Its
	// URLs are absolute (an aggregator stores the document and resolves
	// nothing against where it fetched it from), hence built per request.
	mux.HandleFunc("/"+FeedFileName, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", seo.AtomType)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(e.FeedXML(RequestOrigin(r), built))
	})

	// The sibling edition's pairing is a static property of the two
	// manifests, so it is resolved once, at mount time.
	var pairs map[string]string
	if cfg.alt != nil {
		pairs = e.alternates(cfg.alt)
	}
	// alternate describes the sibling page of the one being served: slug is
	// the article, empty for the index (which always pairs with the sibling
	// index). The zero value means "no counterpart, no cross-link". Both ends
	// of an hreflang pair must be fully-qualified URLs, so the request's own
	// origin prefixes them; the sibling sits on the same host by construction
	// (one server mounts both editions).
	alternate := func(origin, slug string) altLink {
		if cfg.alt == nil {
			return altLink{}
		}
		link := altLink{Lang: cfg.alt.Lang, Label: cfg.alt.UI.SwitchLabel,
			Self: e.absolute(origin, "."), Original: cfg.alt.Original}
		if slug == "" {
			link.Href, link.Nav = origin+cfg.altBase, cfg.altBase
			return link
		}
		other, ok := pairs[slug]
		if !ok {
			return altLink{}
		}
		link.Href, link.Nav = origin+cfg.altBase+other, cfg.altBase+other
		link.Self = e.absolute(origin, slug)
		return link
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		origin := RequestOrigin(r)
		slug := strings.Trim(r.URL.Path, "/")
		if slug == "" {
			data, _, err := build()
			var size int
			if err == nil {
				size = len(data)
			}
			e.writePage(w, origin, cfg.nav, page{
				Title: e.SiteName, Description: e.SiteDescription, Self: ".",
				Alt: alternate(origin, ""), Body: e.indexHTML(size, cfg.home),
			})
			return
		}
		// The markdown mirror of an article, for readers that would rather
		// parse the source than the page. It is a suffix on the SAME URL, so
		// an agent holding the HTML link needs no second lookup.
		if base, ok := strings.CutSuffix(slug, markdownSuffix); ok {
			e.writeMarkdown(w, r, base)
			return
		}
		art, cat, ok := e.find(slug)
		if !ok {
			http.NotFound(w, r)
			return
		}
		e.writePage(w, origin, cfg.nav, page{
			Title: art.Title, Description: art.Blurb, Self: art.Slug,
			Alt: alternate(origin, slug), Article: &art, Category: &cat,
			Body: e.articleHTML(art, cat),
		})
	})
	return mux
}

// markdownSuffix turns an article URL into its source mirror.
const markdownSuffix = ".md"

// sha256Prefix is the first 16 bytes of a body's SHA-256, the strong-ETag
// material every immutable route here hashes with.
func sha256Prefix(body []byte) []byte {
	sum := sha256.Sum256(body)
	return sum[:16]
}

// writeMarkdown serves one article's markdown source as it is written, with a
// single HTML-comment line prepended naming the page it backs. Nothing else is
// transformed: the callout blocks, the [[wiki-links]] and the in-file stamps
// stay exactly as the author wrote them, which is the whole point of the
// route. The source is embedded and immutable for the life of the process, so
// the response carries a strong ETag and honors If-None-Match.
func (e *Edition) writeMarkdown(w http.ResponseWriter, r *http.Request, slug string) {
	art, _, ok := e.find(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	raw, err := assets.ReadFile(e.AssetDir + "/" + art.Slug + ".md")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// The provenance line points at the HTML page, which is the URL to cite:
	// the article's canonical location on the host the reader reached, so a
	// quote carries the address the book wants to be cited by rather than
	// whichever mount served the bytes.
	body := []byte("<!-- source: " + RequestOrigin(r) + e.canonical(art.Slug) +
		" · " + e.SiteName + " · pofo -->\n\n")
	body = append(body, raw...)
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write(body)
}

// canonical turns a page's own relative URL ("." for the index, the slug for
// an article) into the one URL that page should be indexed under: the
// edition's declared home (HomePath) plus that slug, root-relative.
//
// This is the ONE place the handler leaves its mount-anywhere discipline, and
// deliberately. The same book is reachable under more than one prefix on a
// pofo server (the simulator mount republishes the whole constellation under
// its own path), and a relative canonical would happily declare each copy
// canonical. HomePath is exactly the statement needed here: where this edition
// lives online, the same value the EPUB's title page prints. An edition that
// declares no HomePath keeps the relative form.
func (e *Edition) canonical(self string) string {
	if e.HomePath == "" {
		return self
	}
	if self == "." || self == "" {
		return e.HomePath
	}
	return e.HomePath + self
}

// absolute is canonical fully qualified: the request's own origin (scheme and
// host, see RequestOrigin) in front of the canonical path.
//
// The head's canonical link, its hreflang pair and the social-card URLs must
// all be complete URLs, not paths: hreflang annotations are ignored outright
// unless both ends are fully qualified, and a card image is fetched by a
// crawler that holds no document base. The origin cannot be known at build
// time (the same binary answers on localhost, on a tailnet name and behind a
// proxy), so it arrives per request, exactly as the sitemap's does. An edition
// with no HomePath declares no online home and keeps the relative form.
func (e *Edition) absolute(origin, self string) string {
	if e.HomePath == "" {
		return e.canonical(self)
	}
	return origin + e.canonical(self)
}

// page is one rendered page's head material: what it is, where it sits, and
// what it is part of. Article and Category are nil on the index.
type page struct {
	Title       string
	Description string
	Self        string // this page's own relative URL: "." or the slug
	Alt         altLink
	Article     *Article
	Category    *Category
	Body        string
}

func (e *Edition) writePage(w http.ResponseWriter, origin string, nav []NavLink, p page) {
	title, description, alt, body := p.Title, p.Description, p.Alt, p.Body
	var bar strings.Builder
	if len(nav) > 0 || alt.Href != "" {
		bar.WriteString(`<nav class="book-sitenav"><a href=".">` + e.UI.IndexLink + `</a>`)
		for _, l := range nav {
			fmt.Fprintf(&bar, `<a href="%s">%s</a>`, html.EscapeString(l.Href), html.EscapeString(l.Label))
		}
		// The language switch closes the bar: it is the way out of this
		// edition, and the way back in (the index link) already holds the
		// other edge. It uses the sibling's PATH, not the fully-qualified URL
		// the hreflang link needs: a visitor should stay on the host they
		// reached, whichever one that is.
		if alt.Href != "" {
			fmt.Fprintf(&bar, `<a class="book-lang" lang="%s" hreflang="%s" href="%s">%s</a>`,
				alt.Lang, alt.Lang, html.EscapeString(alt.Nav), html.EscapeString(alt.Label))
		}
		bar.WriteString("</nav>\n")
	}

	// The index is the one page whose own title equals the site title; every
	// other page is an article. This drives the <title> shape and og:type.
	index := p.Article == nil
	pageTitle := title + " · " + e.SiteName
	ogType := "article"
	if index {
		pageTitle = e.SiteName
		ogType = "website"
	}
	// The markdown mirror, declared for whoever prefers the source: a machine
	// reading this page can follow the link instead of unpicking the HTML.
	var mdLink string
	if p.Article != nil {
		mdLink = "\n" + `<link rel="alternate" type="text/markdown" href="` +
			html.EscapeString(p.Article.Slug+markdownSuffix) + `">`
	}

	self := e.absolute(origin, p.Self)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<meta name="description" content="%s">
<meta name="robots" content="index, follow">
<meta property="og:type" content="%s">
<meta property="og:locale" content="%s">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<meta property="og:image:type" content="image/png">
<meta property="og:image:width" content="%d">
<meta property="og:image:height" content="%d">
<meta property="og:image:alt" content="%s">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<script type="application/ld+json">%s</script>
<link rel="canonical" href="%s">%s%s
<link rel="alternate" type="application/atom+xml" title="%s" href="%s">
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<link rel="stylesheet" href="fonts.css">
<link rel="stylesheet" href="theme.css">
<style>%s</style>
</head>
<body class="book">
%s%s
<script>%s</script>
</body>
</html>`, e.Lang, html.EscapeString(pageTitle), html.EscapeString(description), ogType, e.OGLocale,
		html.EscapeString(e.SiteName), html.EscapeString(pageTitle), html.EscapeString(description),
		html.EscapeString(self),
		html.EscapeString(e.absolute(origin, CardFileName)), CardWidth, CardHeight,
		html.EscapeString(e.SiteName),
		html.EscapeString(pageTitle), html.EscapeString(description),
		e.jsonLD(p), html.EscapeString(self), e.alternateHead(alt), mdLink,
		html.EscapeString(e.SiteName), FeedFileName,
		e.css(), bar.String(), body, bookJS)
}

// altLink is the sibling edition's counterpart of the page being rendered.
// The zero value means the page has none and gets no cross-link at all.
type altLink struct {
	Href     string // the counterpart's fully-qualified URL, for hreflang
	Nav      string // the same counterpart as a path, for the on-page switch link
	Lang     string // the sibling's language code, for hreflang and the anchor's lang
	Label    string // switch-link text, written in the SIBLING's language
	Self     string // this page's own canonical URL, for the self-referencing hreflang
	Original bool   // the SIBLING is the source edition (it takes the x-default)
}

// alternateHead renders the hreflang links of a cross-linked page: this page
// in this edition's language, its counterpart in the sibling's, and an
// x-default. A search engine needs both halves of the pair to treat them as
// one document in two languages, so the self-referencing link is not optional,
// and it needs FULLY-QUALIFIED URLs on both ends: a path-relative hreflang is
// dropped without a word. altLink carries them already absolute.
//
// The x-default names the SOURCE edition (Edition.Original). There is no
// language-selector page to point it at, and no edition is a translation
// hub, but a reader whose language is neither French nor English is served
// something rather than left to the engine's own tie-break; naming the
// edition every other one is translated from makes that choice deterministic
// and stable. Without a known original, none is emitted.
func (e *Edition) alternateHead(alt altLink) string {
	if alt.Href == "" {
		return ""
	}
	head := fmt.Sprintf("\n"+`<link rel="alternate" hreflang="%s" href="%s">`+
		"\n"+`<link rel="alternate" hreflang="%s" href="%s">`,
		e.Lang, html.EscapeString(alt.Self), alt.Lang, html.EscapeString(alt.Href))
	switch {
	case e.Original:
		head += "\n" + `<link rel="alternate" hreflang="x-default" href="` + html.EscapeString(alt.Self) + `">`
	case alt.Original:
		head += "\n" + `<link rel="alternate" hreflang="x-default" href="` + html.EscapeString(alt.Href) + `">`
	}
	return head
}

// alternates maps this edition's slugs to the sibling edition's counterparts.
// The pairing is Article.Source, which a translation carries and its original
// does not: it reads forward for a translation (its Source names the sibling
// slug) and backward for a source edition (the sibling article naming it).
// Both directions are computed, so the function does not care which edition it
// is called on, and an article with no counterpart is simply absent.
func (e *Edition) alternates(sib *Edition) map[string]string {
	slugs := func(ed *Edition) map[string]bool {
		set := make(map[string]bool)
		for _, cat := range ed.Categories {
			for _, a := range cat.Articles {
				set[a.Slug] = true
			}
		}
		return set
	}
	own, other := slugs(e), slugs(sib)
	pairs := make(map[string]string)
	for _, cat := range e.Categories {
		for _, a := range cat.Articles {
			if a.Source != "" && other[a.Source] {
				pairs[a.Slug] = a.Source
			}
		}
	}
	for _, cat := range sib.Categories {
		for _, a := range cat.Articles {
			if a.Source != "" && own[a.Source] {
				pairs[a.Source] = a.Slug
			}
		}
	}
	return pairs
}

// css is the page stylesheet of one edition: the shared book CSS around the
// single edition-dependent literal it carries, the "link copied" toast of the
// heading anchors (CSS cannot read it from anywhere else).
func (e *Edition) css() string { return bookCSSHead + e.UI.LinkCopied + bookCSSTail }

// publisher is the only identity the structured data declares. The book is
// published by the project, not signed by a person: no author node anywhere.
var publisher = map[string]any{"@type": "Organization", "name": "pofo"}

// jsonLD builds the schema.org structured-data blob for a page: a WebSite and
// the Book itself (its parts and its EPUB edition) on the index, an Article
// plus its BreadcrumbList on a chapter. Every URL is relative, like every
// other URL the handler emits: JSON-LD resolves relative IRIs against the
// document, so the mount-anywhere invariant holds here too.
//
// encoding/json escapes "<", ">" and "&", so the result is safe to inline in
// the <script>.
func (e *Edition) jsonLD(p page) string {
	var nodes []any
	if p.Article == nil {
		// The Book stands on its own here, so it carries the context; nested
		// under an Article below it inherits the enclosing node's.
		book := e.bookNode(p.Description, true)
		book["@context"] = "https://schema.org"
		nodes = []any{
			map[string]any{
				"@context":    "https://schema.org",
				"@type":       "WebSite",
				"name":        e.SiteName,
				"description": p.Description,
				"inLanguage":  e.Lang,
				"url":         e.canonical(p.Self),
				"publisher":   publisher,
			},
			book,
		}
	} else {
		article := map[string]any{
			"@context":    "https://schema.org",
			"@type":       "Article",
			"headline":    p.Title,
			"description": p.Description,
			"inLanguage":  e.Lang,
			"url":         e.canonical(p.Self),
			"isPartOf":    e.bookNode("", false),
			"publisher":   publisher,
			// The markdown mirror, declared as an encoding of this very
			// article: an agent reading the structured data finds the clean
			// source without guessing the URL rule.
			"encoding": map[string]any{
				"@type":           "MediaObject",
				"encodingFormat":  "text/markdown",
				"contentUrl":      e.canonical(p.Article.Slug + markdownSuffix),
				"inLanguage":      e.Lang,
				"isAccessibleFor": "free",
			},
		}
		if p.Category != nil {
			article["articleSection"] = p.Category.Title
		}
		nodes = []any{article, e.breadcrumb(p)}
	}
	b, err := json.Marshal(nodes)
	if err != nil { // a tree of strings never fails to marshal
		return "[]"
	}
	return string(b)
}

// bookNode describes the edition as a Book. full adds what only the index
// page is the right place for: the EPUB as a workExample, and the whole table
// of contents as Chapter parts, which is the machine-readable twin of the
// sommaire the page draws.
func (e *Edition) bookNode(description string, full bool) map[string]any {
	book := map[string]any{
		"@type":      "Book",
		"name":       e.SiteName,
		"inLanguage": e.Lang,
		"url":        e.canonical("."),
		"bookFormat": "https://schema.org/EBook",
		"publisher":  publisher,
	}
	if description != "" {
		book["description"] = description
	}
	if !full {
		return book
	}
	book["workExample"] = map[string]any{
		"@type":          "Book",
		"name":           e.SiteName,
		"bookFormat":     "https://schema.org/EBook",
		"encodingFormat": "application/epub+zip",
		"inLanguage":     e.Lang,
		"url":            e.canonical(e.EPUBFileName),
	}
	var parts []any
	for _, cat := range e.Categories {
		for _, a := range cat.Articles {
			parts = append(parts, map[string]any{
				"@type": "Chapter", "name": a.Title, "description": a.Blurb, "url": e.canonical(a.Slug),
			})
		}
	}
	if len(parts) > 0 {
		book["hasPart"] = parts
	}
	return book
}

// breadcrumb is the article's place in the book: the index, its part (linked
// to the part's own anchor on the index, the same one the top bar uses), then
// the article itself, which carries no item because it is the current page.
func (e *Edition) breadcrumb(p page) map[string]any {
	items := []any{map[string]any{
		"@type": "ListItem", "position": 1, "name": e.SiteName, "item": e.canonical("."),
	}}
	if p.Category != nil {
		items = append(items, map[string]any{
			"@type": "ListItem", "position": len(items) + 1, "name": p.Category.Title,
			"item": e.canonical(".") + "#" + bookmd.HeadingID(p.Category.Title),
		})
	}
	items = append(items, map[string]any{
		"@type": "ListItem", "position": len(items) + 1, "name": p.Title,
	})
	return map[string]any{
		"@context": "https://schema.org", "@type": "BreadcrumbList", "itemListElement": items,
	}
}

// indexHTML renders the sommaire from the manifest. epubSize is the byte
// length of the generated EPUB (0 when it could not be built): a non-zero
// value adds a discreet "Version EPUB" download link with the file size.
// home, when non-empty (WithHome), makes the "pofo" kicker a two-tone link
// there. A category without a written article yet is skipped entirely: the
// English edition fills in as the translation campaign progresses, and an
// empty part would read as a broken page, not a promise.
func (e *Edition) indexHTML(epubSize int, home string) string {
	var b strings.Builder
	b.WriteString(`<header class="book-hero">`)
	if home != "" {
		b.WriteString(`<p class="book-kicker"><a href="` + html.EscapeString(home) + `">po<b>fo</b></a></p>`)
	} else {
		b.WriteString(`<p class="book-kicker">pofo</p>`)
	}
	b.WriteString(`<h1>` + e.SiteName + `</h1>`)
	b.WriteString(`<p class="book-lede">` + e.SiteLede)
	if epubSize > 0 {
		fmt.Fprintf(&b, ` <span class="book-epub"><a href="%s">%s</a> `+
			`<span class="book-epub-size">(%s)</span></span>`,
			e.EPUBFileName, e.UI.EPUBLink, e.UI.HumanSize(epubSize))
	}
	b.WriteString(`</p>`)
	b.WriteString(`</header><main>`)
	for _, cat := range e.Categories {
		if len(cat.Articles) == 0 {
			continue
		}
		// The part title carries the same anchor affordance as an article
		// heading, so a part of the sommaire can be linked to directly.
		id := bookmd.HeadingID(cat.Title)
		fmt.Fprintf(&b, `<section class="book-cat"><h2 id="%s">%s`+
			`<a class="book-hanchor" href="#%s" aria-label="%s">§</a></h2>`+
			`<p class="book-cat-blurb">%s</p><ul class="book-toc">`,
			id, html.EscapeString(cat.Title), id, html.EscapeString(e.UI.PartAnchorLabel),
			html.EscapeString(cat.Blurb))
		for _, a := range cat.Articles {
			fmt.Fprintf(&b, `<li><a href="%s">%s</a><span class="book-toc-blurb">%s</span></li>`,
				a.Slug, html.EscapeString(a.Title), html.EscapeString(a.Blurb))
		}
		b.WriteString(`</ul></section>`)
	}
	b.WriteString(`</main>`)
	return b.String()
}

// articleHTML renders one article page: top bar, title, rendered body, and a
// "same category" footer for lateral navigation.
func (e *Edition) articleHTML(art Article, cat Category) string {
	raw, err := assets.ReadFile(e.AssetDir + "/" + art.Slug + ".md")
	if err != nil {
		return "<p>" + html.EscapeString(e.UI.NotFound) + "</p>"
	}
	body := articleBody(raw)
	var b strings.Builder
	fmt.Fprintf(&b, `<nav class="book-top"><a href=".">← %s</a>`+
		`<a class="book-cat-tag" href="./#%s">%s</a></nav>`,
		html.EscapeString(e.UI.IndexLink), bookmd.HeadingID(cat.Title), html.EscapeString(cat.Title))
	fmt.Fprintf(&b, `<article><h1>%s</h1>%s</article>`,
		html.EscapeString(art.Title), e.addHeadingAnchors(e.ToHTML(body, e.Titles())))
	var others strings.Builder
	for _, a := range cat.Articles {
		if a.Slug == art.Slug {
			continue
		}
		fmt.Fprintf(&others, `<li><a href="%s">%s</a></li>`, a.Slug, html.EscapeString(a.Title))
	}
	if others.Len() > 0 {
		fmt.Fprintf(&b, `<footer class="book-more"><h2>%s</h2><ul>%s</ul></footer>`,
			html.EscapeString(e.UI.SameCategory), others.String())
	}
	return b.String()
}

// reHeading matches one rendered article heading (bookmd emits h2-h4 with a
// slug id and no nesting, so the lazy body match always stops at the right
// closing tag).
var reHeading = regexp.MustCompile(`(?s)<h([2-4]) id="([^"]+)">(.*?)</h[2-4]>`)

// addHeadingAnchors appends a hover-revealed anchor link ("§") to every
// article heading, so a section can be bookmarked or shared. This is a
// web-chrome affordance: it decorates the HTML served by the handler only,
// never the EPUB export (which renders its own body via articleEPUBBody).
// The plain <a href="#id"> works without JavaScript; bookJS layers
// copy-to-clipboard on top.
func (e *Edition) addHeadingAnchors(rendered string) string {
	return reHeading.ReplaceAllString(rendered,
		`<h$1 id="$2">$3<a class="book-hanchor" href="#$2" aria-label="`+
			html.EscapeString(e.UI.SectionAnchorLabel)+`">§</a></h$1>`)
}

// bookJS upgrades the heading anchors: a click copies the section's full URL
// to the clipboard (with a small "lien copié" confirmation) and updates the
// address bar without scrolling. Without JavaScript, or when the clipboard
// API is unavailable (non-secure origins), the anchor still navigates and
// puts the shareable URL in the address bar.
const bookJS = `document.querySelectorAll(".book-hanchor").forEach(function(a){
a.addEventListener("click",function(e){
if(!navigator.clipboard||!navigator.clipboard.writeText)return;
e.preventDefault();
var h=a.getAttribute("href");
history.replaceState(null,"",h);
navigator.clipboard.writeText(location.origin+location.pathname+h).then(function(){
a.classList.add("copied");
setTimeout(function(){a.classList.remove("copied")},1400);
});
});
});`

// bookCSS layers a reading-oriented layout over the shared webui identity: a
// single comfortable measure, generous leading, and the callout boxes. It is
// split in two around its one edition-dependent literal, the "link copied"
// toast of the heading anchors; (*Edition).css joins the halves.
const bookCSSHead = `
body.book{
  --paper:#faf6ef; --paper-2:#f2ebdd; --card:#fffdf9;
  --ink:#211c16; --ink-soft:#4c4438; --muted:#877c6d;
  --rule:rgba(60,48,34,.14); --rule-soft:rgba(60,48,34,.07);
  --accent:#b4783c; --accent-deep:#8a5526; --accent-wash:rgba(180,120,60,.10);
  --good:#3f8f6f; --good-wash:rgba(63,143,111,.08);
  --bad:#c0655b; --bad-wash:rgba(192,101,91,.08);
  --admin:#4a6da0; --admin-wash:rgba(111,147,196,.08);
  --gold-wash:rgba(180,140,50,.11);
  --serif:Georgia,"Iowan Old Style","Palatino Linotype",Palatino,"Times New Roman",serif;
  max-width:44rem;margin:0 auto;padding:2.6rem 1.3rem 5rem;
  font-family:var(--sans);color:var(--ink-soft);
  background:radial-gradient(1100px 560px at 82% -8%,rgba(180,120,60,.07),transparent 60%),var(--paper);
  font-size:1.03rem;line-height:1.72;-webkit-font-smoothing:antialiased}
body.book ::selection{background:var(--accent-wash)}
.book-hero{border-bottom:1px solid var(--rule);padding-bottom:1.3rem;margin-bottom:1.9rem}
.book-kicker{font-family:var(--mono);font-size:.7rem;letter-spacing:.16em;text-transform:uppercase;
  color:var(--accent-deep);opacity:.85;margin:0 0 .55rem}
.book-kicker a{color:var(--ink);text-decoration:none}
.book-kicker a b{color:var(--accent);font-weight:inherit}
.book-kicker a:hover b{color:var(--accent-deep)}
.book h1{font-family:var(--serif);font-weight:600;color:var(--ink);font-size:2.1rem;line-height:1.13;
  margin:0 0 .7rem;letter-spacing:.005em}
.book-lede{color:var(--ink-soft);font-size:1.05rem;line-height:1.62;margin:0;max-width:62ch}
.book-epub{font-family:var(--mono);font-size:.72rem;letter-spacing:.06em;text-transform:uppercase;white-space:nowrap}
.book-epub a{color:var(--accent-deep);text-decoration:none;border-bottom:1px solid var(--rule)}
.book-epub a:hover{border-bottom-color:var(--accent)}
.book-epub-size{color:var(--muted)}
.book-cat{margin:2.4rem 0}
.book-cat h2{font-family:var(--serif);font-weight:600;color:var(--ink);font-size:1.35rem;margin:0 0 .2rem}
.book-cat-blurb{color:var(--muted);font-size:.9rem;margin:0 0 .9rem}
.book-toc{list-style:none;margin:0;padding:0;display:grid;gap:.6rem}
.book-toc li{padding:.75rem .95rem;border:1px solid var(--rule);border-radius:11px;background:var(--card);
  transition:border-color .15s,transform .15s}
.book-toc li:hover{border-color:var(--accent);transform:translateY(-1px)}
.book-toc a{font-family:var(--serif);font-weight:600;font-size:1.05rem;color:var(--ink);
  text-decoration:none;display:block}
.book-toc-blurb{display:block;color:var(--muted);font-size:.86rem;line-height:1.45;margin-top:.2rem}
.book-top{display:flex;justify-content:space-between;align-items:baseline;margin-bottom:1.9rem;font-size:.9rem}
.book-top a{color:var(--accent-deep);text-decoration:none}
.book-top a:hover{text-decoration:underline}
.book-cat-tag{font-family:var(--sans);font-size:.68rem;letter-spacing:.1em;text-transform:uppercase;
  color:var(--muted);text-decoration:none;border-bottom:0}
.book-cat-tag:hover{color:var(--accent-deep)}
article h1{padding-bottom:.55rem;border-bottom:2px solid var(--accent);margin-bottom:1.15rem}
article h2{font-family:var(--serif);font-weight:600;color:var(--ink);font-size:1.5rem;line-height:1.2;margin:2.1rem 0 .7rem}
article h3{font-family:var(--sans);font-weight:600;color:var(--ink);font-size:1.12rem;margin:1.7rem 0 .45rem}
article h4{font-family:var(--sans);font-weight:600;color:var(--accent-deep);font-size:.82rem;
  text-transform:uppercase;letter-spacing:.05em;margin:1.4rem 0 .4rem}
article p{margin:0 0 1rem}
article strong{color:var(--ink);font-weight:600}
article ul,article ol{margin:0 0 1rem;padding-left:1.35rem}
article li{margin:.32rem 0}
article li.task{list-style:none;margin-left:-1.1rem}
article code{font-family:var(--mono);font-size:.85em;background:var(--paper-2);padding:.08em .35em;border-radius:4px;color:var(--ink)}
article a{color:var(--accent-deep);text-decoration:none;border-bottom:1px solid var(--rule)}
article a:hover{border-bottom-color:var(--accent)}
article h2[id],article h3[id],article h4[id],.book-cat h2[id]{scroll-margin-top:1rem}
.book-hanchor{margin-left:.45em;font-family:var(--sans);font-weight:400;font-size:.72em;
  color:var(--accent-deep);border-bottom:0;text-decoration:none;opacity:0;
  transition:opacity .15s ease;position:relative}
h2:hover .book-hanchor,h3:hover .book-hanchor,h4:hover .book-hanchor,
.book-hanchor:focus-visible{opacity:.55}
.book-hanchor:hover{opacity:1}
.book-hanchor.copied::after{content:"`

const bookCSSTail = `";position:absolute;bottom:1.35em;left:50%;
  transform:translateX(-50%);font-family:var(--mono);font-size:.6rem;letter-spacing:.06em;
  text-transform:none;color:var(--accent-deep);background:var(--card);border:1px solid var(--rule);
  border-radius:6px;padding:.14em .55em;white-space:nowrap}
@media (hover:none){.book-hanchor{opacity:.3}}
article blockquote{margin:1.1rem 0;padding:.2rem 0 .2rem 1rem;border-left:3px solid var(--accent);color:var(--muted);font-style:italic}
article hr{border:0;border-top:1px solid var(--rule);margin:2rem 0}
.book-fig{margin:1.7rem 0}
.book-fig svg{border:1px solid var(--rule);border-radius:11px;background:var(--card);padding:.4rem}
.book-fig figcaption{font-family:var(--sans);font-size:.82rem;color:var(--muted);margin-top:.5rem;text-align:center;line-height:1.4}
.table-wrap{overflow-x:auto;margin:1.15rem 0}
article table{width:100%;border-collapse:collapse;font-family:var(--sans);font-size:.86rem;margin:0}
article th,article td{text-align:left;padding:.45rem .6rem;border-bottom:1px solid var(--rule-soft);vertical-align:top}
article thead th{font-weight:600;color:var(--accent-deep);border-bottom:1px solid var(--rule)}
article tr:last-child td{border-bottom:0}
.doc-box{margin:1.5rem 0;padding:.9rem 1.05rem;border-radius:11px;border:1px solid var(--rule);
  background:var(--paper-2);font-size:.95em;line-height:1.6}
.doc-box p:first-of-type{margin-top:.35rem}
.doc-box p:last-child{margin-bottom:0}
.doc-box-h{font-family:var(--sans);font-weight:600;font-size:.83rem;margin-bottom:.35rem;color:var(--ink)}
.doc-box-glyph{margin-right:.3rem}
.doc-box--cle{background:var(--card);border-color:var(--accent);border-left:3px solid var(--accent)}
.doc-box--cle .doc-box-h{color:var(--accent-deep)}
.doc-box--astuce{background:var(--good-wash);border-color:rgba(63,143,111,.3)}
.doc-box--astuce .doc-box-h{color:var(--good)}
.doc-box--attention{background:var(--bad-wash);border-color:rgba(192,101,91,.3)}
.doc-box--attention .doc-box-h{color:var(--bad)}
.doc-box--exemple{background:var(--accent-wash);border-color:rgba(180,120,60,.3)}
.doc-box--exemple .doc-box-h{color:var(--accent-deep)}
.doc-box--science{background:var(--admin-wash);border-color:rgba(111,147,196,.3)}
.doc-box--science .doc-box-h{color:var(--admin)}
.doc-box--terrain{background:var(--gold-wash);border-color:rgba(180,140,50,.32)}
.doc-box--terrain .doc-box-h{color:#8a6a1c}
.doc-box--admin{background:var(--card);border-color:var(--rule);border-left:3px solid var(--admin)}
.doc-box--admin .doc-box-h{color:var(--admin)}
.book-more{margin-top:2.6rem;padding-top:1.2rem;border-top:1px solid var(--rule)}
.book-more h2{font-family:var(--sans);font-size:.72rem;letter-spacing:.09em;text-transform:uppercase;color:var(--muted);margin:0}
.book-more ul{list-style:none;padding:0;margin:.5rem 0 0}
.book-more li{margin:.3rem 0}
.book-more a{font-family:var(--serif);color:var(--accent-deep);text-decoration:none}
.book-more a:hover{text-decoration:underline}
@media (max-width:640px){body.book{font-size:1rem;padding:1.6rem 1.1rem 4rem}
  .book h1{font-size:1.8rem}article h2{font-size:1.32rem}article table{font-size:.8rem}}
/* The index link anchors the left, the cross-navigation sits right: the way
   back into the book and the way out of it never compete for the same edge. */
.book-sitenav{display:flex;flex-wrap:wrap;gap:.5rem 1.15rem;margin:0 0 1.7rem;padding-bottom:.65rem;
  border-bottom:1px solid var(--rule-soft);font-family:var(--mono);font-size:.72rem;
  letter-spacing:.07em;text-transform:uppercase}
.book-sitenav a{color:var(--muted);text-decoration:none}
.book-sitenav a:hover{color:var(--accent-deep)}
.book-sitenav a:first-child{color:var(--accent-deep);margin-right:auto}
.book-sitenav a.book-lang{color:var(--accent-deep);opacity:.8}
.book-sitenav a.book-lang:hover{opacity:1}
@media print{.book-sitenav{display:none}}
`
