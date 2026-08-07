package bookmd

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Callout describes one ::: block type: the label heads the box when the
// author gives no title, the glyph prefixes it.
type Callout struct{ Glyph, Label string }

// Callouts is the built-in superset of ::: block types used by all books.
// Unknown types degrade to "encart".
var Callouts = map[string]Callout{
	"encart":    {"❖", "En passant"},
	"cle":       {"🔑", "L'idée clé"},
	"astuce":    {"💡", "Astuce"},
	"attention": {"⚠", "Point de vigilance"},
	"exemple":   {"🧮", "Exemple chiffré"},
	"science":   {"🔬", "Ce que dit la recherche"},
	"terrain":   {"🗣", "Retour de terrain"},
	"admin":     {"📋", "Côté administratif"},
}

// Options tunes rendering. The zero value renders wiki-links as plain text
// (Titles nil), keeps href="<slug>" (Href nil) and drops figure blocks
// entirely (Figure nil).
type Options struct {
	Titles map[string]string        // written slugs -> display titles
	Href   func(slug string) string // wiki-link target; nil -> the slug itself
	Figure func(id string) string   // ::: figure payload; nil -> figure block dropped entirely
}

var (
	reCallout  = regexp.MustCompile(`^:::\s*(\w+)?\s*(.*)$`)
	reHeading  = regexp.MustCompile(`^(#{1,4})\s+(.*)$`)
	reRule     = regexp.MustCompile(`^---+\s*$`)
	reTableSep = regexp.MustCompile(`^\s*\|?[\s:|-]+\|[\s:|-]*$`)
	reUL       = regexp.MustCompile(`^\s*[-*]\s+(.*)$`)
	reOL       = regexp.MustCompile(`^\s*\d+\.\s+(.*)$`)
	reTask     = regexp.MustCompile(`^\[( |x|X)\]\s+(.*)$`)
	reQuote    = regexp.MustCompile(`^>\s?`)
	reBlank    = regexp.MustCompile(`^\s*$`)
	reBlockAny = regexp.MustCompile(`^(#{1,4}\s|:::|>\s?|---+\s*$|\s*[-*]\s+|\s*\d+\.\s+)`)
	reAnchor   = regexp.MustCompile(`[^\p{L}\p{N}]+`)

	reCode     = regexp.MustCompile("`([^`]+)`")
	reWikiLbl  = regexp.MustCompile(`\[\[([^\]|]+)\|([^\]]+)\]\]`)
	reWiki     = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	reLink     = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalic   = regexp.MustCompile(`(^|[^*])\*([^*]+)\*`)
	reCodeSpan = regexp.MustCompile(`<code>.*?</code>`)
)

// wikiLink renders one [[slug]] reference. Written articles (present in
// opt.Titles) become links; planned-but-unwritten targets degrade to their
// label as plain text so readers never hit a dead link. The link target is
// opt.Href(slug) when a hook is set, otherwise the bare slug.
func wikiLink(slug, label string, opt Options) string {
	slug, label = strings.TrimSpace(slug), strings.TrimSpace(label)
	if title, ok := opt.Titles[slug]; ok {
		if label == "" {
			label = title
		}
		href := slug
		if opt.Href != nil {
			href = opt.Href(slug)
		}
		return fmt.Sprintf(`<a href="%s" class="doc-link">%s</a>`, href, label)
	}
	if label == "" {
		label = slug
	}
	return label
}

// mdInline applies inline formatting to already-HTML-escaped text.
func mdInline(s string, opt Options) string {
	s = reCode.ReplaceAllString(s, "<code>$1</code>")
	s = reWikiLbl.ReplaceAllStringFunc(s, func(m string) string {
		g := reWikiLbl.FindStringSubmatch(m)
		return wikiLink(g[1], g[2], opt)
	})
	s = reWiki.ReplaceAllStringFunc(s, func(m string) string {
		g := reWiki.FindStringSubmatch(m)
		return wikiLink(g[1], "", opt)
	})
	s = reLink.ReplaceAllString(s, `<a href="$2" target="_blank" rel="noopener">$1</a>`)
	// Emphasis must not rewrite the * or ` that may appear inside code
	// spans: shield them, emphasize, restore.
	var spans []string
	s = reCodeSpan.ReplaceAllStringFunc(s, func(m string) string {
		spans = append(spans, m)
		return fmt.Sprintf("\x00%d\x00", len(spans)-1)
	})
	// Italic first, then bold: a bold span may wrap an italic (book titles,
	// "**Auteur, *Titre* (an)**"), and reBold's [^*]+ cannot span the inner
	// stars, so the italics must be resolved to <em> before bold runs.
	s = reItalic.ReplaceAllString(s, "$1<em>$2</em>")
	s = reBold.ReplaceAllString(s, "<strong>$1</strong>")
	for i, span := range spans {
		s = strings.Replace(s, fmt.Sprintf("\x00%d\x00", i), span, 1)
	}
	return s
}

// uniqueID returns base the first time it is seen, then base-2, base-3, ...
// for later collisions, bumping past any candidate already taken (including
// natural slugs), so every heading id in one render is distinct. used is the
// per-document set, shared across nested renders (callout bodies, quotes).
func uniqueID(base string, used map[string]bool) string {
	id := base
	for n := 2; used[id]; n++ {
		id = fmt.Sprintf("%s-%d", base, n)
	}
	used[id] = true
	return id
}

// ToHTML renders one article body (the book's Markdown dialect) to HTML.
// opt.Titles maps the slugs of WRITTEN articles to their display titles; it
// drives [[slug]] links (see wikiLink). The zero Options renders wiki-links
// as plain text, keeps href="<slug>" and drops ::: figure blocks.
func ToHTML(src string, opt Options) string {
	return render(src, opt, map[string]bool{})
}

// render is the recursive worker. used carries the heading-id set of the whole
// document so ids stay unique across callout and blockquote sub-renders.
func render(src string, opt Options, used map[string]bool) string {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var b strings.Builder
	inline := func(s string) string { return mdInline(html.EscapeString(s), opt) }

	i := 0
	for i < len(lines) {
		line := lines[i]

		if g := reCallout.FindStringSubmatch(line); g != nil && strings.HasPrefix(line, ":::") {
			typ := strings.ToLower(g[1])
			// Figure block: "::: figure <id>" + caption lines + ":::".
			// opt.Figure supplies the payload (in firebook an inline SVG
			// generated in Go); a nil hook drops the block entirely.
			if typ == "figure" {
				id := strings.TrimSpace(g[2])
				var cap []string
				i++
				for i < len(lines) && strings.TrimSpace(lines[i]) != ":::" {
					cap = append(cap, lines[i])
					i++
				}
				i++ // closing :::
				if opt.Figure == nil {
					continue
				}
				fmt.Fprintf(&b, `<figure class="book-fig">%s<figcaption>%s</figcaption></figure>`,
					opt.Figure(id), inline(strings.TrimSpace(strings.Join(cap, " "))))
				continue
			}
			meta, ok := Callouts[typ]
			if !ok {
				typ, meta = "encart", Callouts["encart"]
			}
			title := strings.TrimSpace(g[2])
			if title == "" {
				title = meta.Label
			}
			var body []string
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) != ":::" {
				body = append(body, lines[i])
				i++
			}
			i++ // closing :::
			fmt.Fprintf(&b, `<aside class="doc-box doc-box--%s"><div class="doc-box-h"><span class="doc-box-glyph">%s</span> %s</div>%s</aside>`,
				typ, meta.Glyph, inline(title), render(strings.Join(body, "\n"), opt, used))
			continue
		}

		if g := reHeading.FindStringSubmatch(line); g != nil {
			lvl := min(len(g[1])+1, 4) // # is demoted: the shell owns the h1
			text := strings.TrimSpace(g[2])
			base := strings.Trim(reAnchor.ReplaceAllString(strings.ToLower(text), "-"), "-")
			id := uniqueID(base, used)
			fmt.Fprintf(&b, `<h%d id="%s">%s</h%d>`, lvl, id, inline(text), lvl)
			i++
			continue
		}

		if reRule.MatchString(line) {
			b.WriteString("<hr>")
			i++
			continue
		}

		if strings.Contains(line, "|") && i+1 < len(lines) && reTableSep.MatchString(lines[i+1]) {
			var rows []string
			for i < len(lines) && strings.Contains(lines[i], "|") {
				rows = append(rows, lines[i])
				i++
			}
			cells := func(r string) []string {
				r = strings.TrimSpace(r)
				r = strings.TrimPrefix(r, "|")
				r = strings.TrimSuffix(r, "|")
				parts := strings.Split(r, "|")
				for j := range parts {
					parts[j] = strings.TrimSpace(parts[j])
				}
				return parts
			}
			b.WriteString(`<div class="table-wrap"><table><thead><tr>`)
			for _, c := range cells(rows[0]) {
				fmt.Fprintf(&b, "<th>%s</th>", inline(c))
			}
			b.WriteString("</tr></thead><tbody>")
			for _, r := range rows[2:] {
				b.WriteString("<tr>")
				for _, c := range cells(r) {
					fmt.Fprintf(&b, "<td>%s</td>", inline(c))
				}
				b.WriteString("</tr>")
			}
			b.WriteString("</tbody></table></div>")
			continue
		}

		if reUL.MatchString(line) || reOL.MatchString(line) {
			re, tag := reUL, "ul"
			if reOL.MatchString(line) {
				re, tag = reOL, "ol"
			}
			fmt.Fprintf(&b, "<%s>", tag)
			for i < len(lines) && re.MatchString(lines[i]) {
				item := re.FindStringSubmatch(lines[i])[1]
				if t := reTask.FindStringSubmatch(item); t != nil {
					checked := ""
					if t[1] != " " {
						checked = " checked"
					}
					fmt.Fprintf(&b, `<li class="task"><input type="checkbox" disabled%s> %s</li>`, checked, inline(t[2]))
				} else {
					fmt.Fprintf(&b, "<li>%s</li>", inline(item))
				}
				i++
			}
			fmt.Fprintf(&b, "</%s>", tag)
			continue
		}

		if reQuote.MatchString(line) {
			var body []string
			for i < len(lines) && reQuote.MatchString(lines[i]) {
				body = append(body, reQuote.ReplaceAllString(lines[i], ""))
				i++
			}
			fmt.Fprintf(&b, "<blockquote>%s</blockquote>", render(strings.Join(body, "\n"), opt, used))
			continue
		}

		if reBlank.MatchString(line) {
			i++
			continue
		}

		var para []string
		for i < len(lines) && !reBlank.MatchString(lines[i]) && !reBlockAny.MatchString(lines[i]) {
			para = append(para, lines[i])
			i++
		}
		if len(para) > 0 {
			fmt.Fprintf(&b, "<p>%s</p>", inline(strings.Join(para, " ")))
		}
	}
	return b.String()
}
