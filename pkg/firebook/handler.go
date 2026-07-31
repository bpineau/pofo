package firebook

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/bpineau/pofo/pkg/webui"
)

// Handler serves the book: the sommaire at "/", one HTML page per article at
// "/<slug>", and the shared identity stylesheets at "/theme.css" and
// "/fonts.css" (relative URLs, so the handler can be mounted under any
// prefix, e.g. http.StripPrefix("/livre", firebook.Handler())).
func Handler() http.Handler {
	mux := http.NewServeMux()
	css := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.HandleFunc("/theme.css", css(webui.CSS))
	mux.HandleFunc("/fonts.css", css(webui.FontsCSS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.Trim(r.URL.Path, "/")
		if slug == "" {
			writePage(w, "Le livre FIRE", indexHTML())
			return
		}
		art, cat, ok := find(slug)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writePage(w, art.Title, articleHTML(art, cat))
	})
	return mux
}

func writePage(w http.ResponseWriter, title, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s · Le livre FIRE</title>
<link rel="stylesheet" href="fonts.css">
<link rel="stylesheet" href="theme.css">
<style>%s</style>
</head>
<body class="book">
%s
</body>
</html>`, html.EscapeString(title), bookCSS, body)
}

// indexHTML renders the sommaire from the manifest.
func indexHTML() string {
	var b strings.Builder
	b.WriteString(`<header class="book-hero">`)
	b.WriteString(`<p class="book-kicker">pofo · référence</p>`)
	b.WriteString(`<h1>Le livre FIRE</h1>`)
	b.WriteString(`<p class="book-lede">Vivre de son capital sans le survivre : la science du retrait, ` +
		`les stratégies, les portefeuilles qui résistent, la fiscalité française et le facteur humain. ` +
		`Chaque article se lit seul ; commencez par ce qui vous préoccupe. Le livre est en cours ` +
		`d'écriture : le sommaire s'allonge régulièrement.</p>`)
	b.WriteString(`</header><main>`)
	for _, cat := range Categories {
		fmt.Fprintf(&b, `<section class="book-cat"><h2>%s</h2><p class="book-cat-blurb">%s</p><ul class="book-toc">`,
			html.EscapeString(cat.Title), html.EscapeString(cat.Blurb))
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
func articleHTML(art Article, cat Category) string {
	raw, err := assets.ReadFile("assets/book/" + art.Slug + ".md")
	if err != nil {
		return "<p>Article introuvable.</p>"
	}
	body := strings.TrimSpace(string(raw))
	// Drop the in-file "# Title" front line: the shell renders the h1.
	if strings.HasPrefix(body, "# ") {
		if _, rest, found := strings.Cut(body, "\n"); found {
			body = rest
		} else {
			body = ""
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<nav class="book-top"><a href=".">← Sommaire</a><span class="book-cat-tag">%s</span></nav>`,
		html.EscapeString(cat.Title))
	fmt.Fprintf(&b, `<article><h1>%s</h1>%s</article>`, html.EscapeString(art.Title), ToHTML(body, Titles()))
	var others strings.Builder
	for _, a := range cat.Articles {
		if a.Slug == art.Slug {
			continue
		}
		fmt.Fprintf(&others, `<li><a href="%s">%s</a></li>`, a.Slug, html.EscapeString(a.Title))
	}
	if others.Len() > 0 {
		fmt.Fprintf(&b, `<footer class="book-more"><h2>Dans la même partie</h2><ul>%s</ul></footer>`, others.String())
	}
	return b.String()
}

// bookCSS layers a reading-oriented layout over the shared webui identity:
// a single comfortable measure, generous leading, and the callout boxes.
const bookCSS = `
body.book{max-width:46rem;margin:0 auto;padding:2.5rem 1.25rem 5rem;
  font-family:var(--sans);color:var(--ink);background:var(--bg);
  font-size:1.0625rem;line-height:1.65}
.book-hero{margin-bottom:2.5rem}
.book-kicker{font-family:var(--mono);font-size:.75rem;letter-spacing:.14em;
  text-transform:uppercase;color:var(--accent-ink);margin:0 0 .5rem}
.book h1{font-size:2rem;line-height:1.15;margin:0 0 .75rem;letter-spacing:-.01em}
.book-lede{color:var(--ink-soft);font-size:1.125rem;margin:0}
.book-cat{margin:2.75rem 0}
.book-cat h2{font-size:1.35rem;margin:0 0 .25rem}
.book-cat-blurb{color:var(--muted);margin:0 0 1rem}
.book-toc{list-style:none;margin:0;padding:0}
.book-toc li{padding:.7rem .9rem;border:1px solid var(--line);border-radius:var(--r-sm);
  margin-bottom:.5rem;background:var(--surface)}
.book-toc a{font-weight:600;color:var(--accent-ink);text-decoration:none;display:block}
.book-toc a:hover{text-decoration:underline}
.book-toc-blurb{display:block;color:var(--ink-soft);font-size:.95rem;margin-top:.15rem}
.book-top{display:flex;justify-content:space-between;align-items:baseline;
  margin-bottom:2rem;font-size:.95rem}
.book-top a{color:var(--accent-ink);text-decoration:none}
.book-top a:hover{text-decoration:underline}
.book-cat-tag{font-family:var(--mono);font-size:.75rem;letter-spacing:.1em;
  text-transform:uppercase;color:var(--muted)}
article h2{font-size:1.45rem;margin:2.25rem 0 .75rem;line-height:1.25}
article h3{font-size:1.15rem;margin:1.75rem 0 .5rem}
article h4{font-size:1rem;margin:1.5rem 0 .5rem}
article p{margin:.9rem 0}
article ul,article ol{padding-left:1.4rem;margin:.9rem 0}
article li{margin:.35rem 0}
article li.task{list-style:none;margin-left:-1.4rem}
article code{font-family:var(--mono);font-size:.9em;background:var(--surface-2);
  padding:.1em .35em;border-radius:4px}
article a{color:var(--accent-ink)}
article blockquote{margin:1.25rem 0;padding:.25rem 1.1rem;border-left:3px solid var(--line-strong);
  color:var(--ink-soft)}
article hr{border:0;border-top:1px solid var(--line);margin:2rem 0}
article table{border-collapse:collapse;width:100%;margin:1.25rem 0;font-size:.95rem;
  background:var(--surface);border:1px solid var(--line);border-radius:var(--r-sm)}
article th{text-align:left;padding:.55rem .7rem;border-bottom:1px solid var(--line-strong);
  font-size:.85rem;letter-spacing:.03em;color:var(--ink-soft)}
article td{padding:.5rem .7rem;border-bottom:1px solid var(--line);vertical-align:top}
article tr:last-child td{border-bottom:0}
.doc-box{margin:1.5rem 0;padding:1rem 1.15rem;border-radius:var(--r-sm);
  border:1px solid var(--line);background:var(--surface);font-size:.97em}
.doc-box p:first-of-type{margin-top:.4rem}
.doc-box p:last-child{margin-bottom:0}
.doc-box-h{font-weight:650;font-size:.95rem}
.doc-box-glyph{font-family:var(--mono);margin-right:.15rem}
.doc-box--cle{background:var(--accent-wash);border-color:var(--accent)}
.doc-box--cle .doc-box-h{color:var(--accent-ink)}
.doc-box--astuce{background:var(--good-wash);border-color:var(--good)}
.doc-box--astuce .doc-box-h{color:var(--good-ink)}
.doc-box--attention{background:var(--bad-wash);border-color:var(--bad)}
.doc-box--attention .doc-box-h{color:var(--bad-ink)}
.doc-box--exemple{background:var(--surface-2)}
.doc-box--science{border-left:3px solid var(--accent)}
.doc-box--science .doc-box-h{color:var(--accent-ink)}
.doc-box--terrain{background:var(--warn-wash);border-color:var(--warn)}
.doc-box--terrain .doc-box-h{color:var(--warn-ink)}
.book-more{margin-top:3rem;padding-top:1.25rem;border-top:1px solid var(--line)}
.book-more h2{font-size:.85rem;letter-spacing:.08em;text-transform:uppercase;color:var(--muted)}
.book-more ul{list-style:none;padding:0;margin:.5rem 0 0}
.book-more li{margin:.3rem 0}
.book-more a{color:var(--accent-ink);text-decoration:none}
.book-more a:hover{text-decoration:underline}
@media (max-width:640px){body.book{font-size:1rem;padding-top:1.5rem}}
`
