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
// prefix, e.g. http.StripPrefix("/book/fr", firebook.Handler())).
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
		`les modèles et leurs pièges, les stratégies, les portefeuilles qui résistent, les buffers, ` +
		`l'inflation, la fiscalité française et le facteur humain. Chaque article se lit seul et renvoie ` +
		`aux autres ; commencez par ce qui vous préoccupe, ou par le début.</p>`)
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
	raw, err := assets.ReadFile("assets/book/fr/" + art.Slug + ".md")
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
.book h1{font-family:var(--serif);font-weight:600;color:var(--ink);font-size:2.1rem;line-height:1.13;
  margin:0 0 .7rem;letter-spacing:.005em}
.book-lede{color:var(--ink-soft);font-size:1.05rem;line-height:1.62;margin:0;max-width:62ch}
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
.book-cat-tag{font-family:var(--sans);font-size:.68rem;letter-spacing:.1em;text-transform:uppercase;color:var(--muted)}
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
.book-more{margin-top:2.6rem;padding-top:1.2rem;border-top:1px solid var(--rule)}
.book-more h2{font-family:var(--sans);font-size:.72rem;letter-spacing:.09em;text-transform:uppercase;color:var(--muted);margin:0}
.book-more ul{list-style:none;padding:0;margin:.5rem 0 0}
.book-more li{margin:.3rem 0}
.book-more a{font-family:var(--serif);color:var(--accent-deep);text-decoration:none}
.book-more a:hover{text-decoration:underline}
@media (max-width:640px){body.book{font-size:1rem;padding:1.6rem 1.1rem 4rem}
  .book h1{font-size:1.8rem}article h2{font-size:1.32rem}article table{font-size:.8rem}}
`
