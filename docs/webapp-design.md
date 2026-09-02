# The web app (`-serve`): the constellation

Status: shipped (2026-07-19).

`pofo -serve` puts the whole tool behind one local port. Everything the CLI
already does, plus the FIRE explorer and the FIRE book, becomes a set of web
surfaces a person can browse, bookmark and share on a private network, without
running a command per comparison.

## Goals

- **One binary, one port, every surface.** No separate services, no build
  step, no external process. The same embedded datasets and the same in-memory
  quote cache back the report, the simulator and the book.
- **Shareable state in the URL.** A comparison is a link (`/view?...`), so it
  can be bookmarked or sent to a peer and reproduces exactly.
- **Safe to expose to yourself.** The default bind is loopback; the intended
  way to reach it from another device is a private tailnet, not an open port.
  Anonymous visitors can never make the server fetch arbitrary identifiers.
- **Coherent, not uniform.** The three surfaces (report, simulator, book) keep
  their own layers but share one visual identity, so the app reads as one tool.

## Route map

| Route | Handler | What it serves |
|---|---|---|
| `/` | `landing` (`landing.go`) | the front door: a minimal English landing page, the two-tone pofo mark, one sentence, and four cards linking the sections (the two book editions, the visualizer, the simulator) |
| `/visualizer` | `hub` (`hub.go`) | the portfolio visualizer's home ("Put portfolios side by side."): the live composer plus the bundled example portfolios as a pure-GET checkbox form that submits ticked names to `/view`; the trailing-slash form 301-redirects to the canonical path |
| `/view` | `view` (`serve.go`, grammar in `view.go`) | the HTML comparison report the CLI writes, addressed by a query string (`ex=` / `p=` + global overrides) |
| `/examples/<name>.txt` | `exampleFile` | one embedded portfolio file, raw text (the hub's "Source" link) |
| `/firesimulator/` | `fire` (`serve.go`) -> `pkg/decumul/web.Handler`, prefix-stripped | the FIRE simulator on the startup panel, identical to `-fire`; the old `/firesimulator/` path 301-redirects here (sub-path and query preserved) |
| `/firesimulator/e/<name>/` | `fire` -> a per-example `web.Handler` | the simulator pre-loaded with one example's historical panel (the hub's "Simulate" link), built and cached lazily on first use |
| `/firesimulator/p/<spec>/` | `fire` -> a per-spec `web.Handler` | the simulator bound to an ad-hoc composed portfolio, `<spec>` being exactly the `p=` grammar in one path segment; catalog-gated, bounded lazily-built cache |
| `/firebook/fr/` | `pkg/firebook.Handler`, prefix-stripped | the French FIRE book ("Le FIRE tranquille"), with a chrome nav bar back to the other surfaces and a home-linked kicker (`firebook.WithHome`); the old `/book/fr/` path 301-redirects here |
| `/firebook/en/` | `firebook.English.Handler`, prefix-stripped | the English edition ("The Quiet FIRE"), complete since 2026-08-19: every French article that is not French-only has a counterpart here |
| `/theme.css`, `/fonts.css` | inline | the shared `pkg/webui` identity tokens and embedded fonts; content-fingerprinted (see below) |
| `/favicon.svg`, `/favicon.ico` | inline (`serve.go`) | the shared tab icon (`webui.FaviconSVG`, a petrol "p"); every head links `/favicon.svg`, the report inlines it as a data URI to stay self-contained |
| `/sitemap.xml`, `/robots.txt`, `/llms.txt` | `firebook.Site.Handle` (`pkg/firebook/site.go`) | the machine-readable face of the constellation: every article, index and EPUB of both book editions plus the app surfaces; every page allowed, `/view` disallowed as a compute endpoint (`Site.Disallow`, repeated in the named-crawler record) and the AI crawlers named one by one; the llmstxt.org index. Absolute URLs are built per request from the `Host` header (`firebook.RequestOrigin`), since the public origin is not known at build time. Also mounted by `pkg/decumul/web` |
| `/firebook/<lang>/<slug>.md` | `pkg/firebook.Handler` | one article's Markdown SOURCE, untransformed, behind a one-line HTML-comment provenance header naming the page to cite; strong ETag, `text/markdown`. The HTML page declares it (`<link rel="alternate" type="text/markdown">`) and `llms.txt` lists it |
| `/firebook/<lang>/card.png` | `pkg/firebook.Handler` (`card.go`) | the edition's social card, 1200x630, embedded via `go:embed` and served with a strong ETag; what `og:image` points at |
| `/firebook/<lang>/feed.xml` | `pkg/firebook.Handler` (`feed.go`) | the edition's Atom 1.0 feed: one entry per article in reading order, the manifest blurb as its summary, absolute URLs from `RequestOrigin`. Every page of the edition declares it in its head and `llms.txt` lists it; the sitemap does not (a feed is not a page to index). Mounted by both servers, since it rides the edition handler itself |
| `/<key>.txt` | `firebook.Site.Handle` | the IndexNow ownership key file, mounted only when `-serve -indexnow-key <key>` names one (`pkg/decumul/web`: `WithIndexNowKey`). Its name is the key and its body is the key again |
| `/catalog.json` | inline (`serve.go`) | the local catalog as JSON (`marketdata.LocalCatalog`: `{ID,Name,Class,Alt}` sorted, byte-stable), marshaled once at startup; GET-only, `Cache-Control: public, max-age=3600`; feeds the composer's autocomplete and inline validation |
| `/composer.js`, `/composer.css` | inline (`composer.go`) | the live composer's embedded front end (the in-page editor over the `/view` grammar); content-fingerprinted (see below) |
| `/healthz` | `health` (`serve.go`) | the liveness probe: `200 text/plain` `ok`, `Cache-Control: no-store`, GET/HEAD only. It checks that the process serves and nothing else, on purpose: a health endpoint that depended on a quote source would report the server dead every time that source hiccups. The one route excluded from the access log (see below); absent from the sitemap and from `robots.txt`, since it is not a page |

Every `text/html` route above passes through one response filter,
`webui.Beacon`, which splices the optional Cloudflare Web Analytics tag in front
of the closing `</body>`. It wraps the finished mux, so no renderer knows about
it, and without a token it returns the mux itself. See "Audience: the optional
analytics beacon" below.

The two `/firebook` mounts are cross-linked by
`firebook.WithAlternate(base, sibling)`, which each mount hands the OTHER one's
base path (the handler emits relative URLs and
cannot know where its sibling sits). A page with a counterpart then carries a
`<link rel="alternate" hreflang="fr">` / `hreflang="en">` pair in its head, an
`hreflang="x-default"` naming the French (source) edition, and a discreet
language switch at the end of the top bar, labelled in the
target's language ("English version" on a French page). The two indexes always
pair; an article pairs through `Article.Source`, so the French-only tax part
declares nothing and points nowhere. Without the option the book renders
exactly as before, which the offline and `-fire` mounts rely on.

A handful of head URLs are NOT relative, on purpose: `rel="canonical"`,
`og:url`, the three `hreflang` links and `og:image`. They are **fully
qualified**: the request's own origin (`firebook.RequestOrigin`, the same
helper the sitemap uses) in front of a path built from `Edition.HomePath`.

Two separate reasons stack here. `HomePath` answers *which copy counts*: the
constellation republishes the whole book under the simulator's own prefix
(`/firesimulator/firebook/fr/...`, a consequence of mounting `pkg/decumul/web`
under a path), and a relative canonical would declare every copy canonical.
`RequestOrigin` answers *on which host*: an `hreflang` annotation is ignored
unless both ends are complete URLs, a canonical link should agree with the
`hreflang` beside it, and a social-card crawler fetching `og:image` holds no
document base to resolve a path against. The visible language switch in the top
bar stays a path (a visitor should stay on the host they reached), and
everything else the handler emits stays relative and mount-agnostic.

The four static assets above are **content-fingerprinted**: the HTML surfaces
link them as `…?v=<hash>` (`assetURL`/`versionedAssets` in `serve.go`, applied to
the hub, composer and error templates at parse time), and a versioned request is
served `Cache-Control: immutable`. A deploy that changes an asset changes its
URL, so an edge cache (Cloudflare, which keys by full URL, query string
included) cannot serve stale CSS/JS: the HTML that carries the new hash is itself
dynamic (`cf-cache-status: DYNAMIC`, never edge-cached), so the fresh URL reaches
the browser on the next request with no manual purge. `pkg/decumul/web` does the
same for the FIRE page's own `app.js`/`app.css`/`theme.css`/`fonts.css`.

The mux (`server.handler` in `serve.go`) is a plain `http.ServeMux`; the
lifecycle (`runServe`) mirrors `runFire`: bind, serve, shut down on context
cancel. Portfolio file arguments are turned into a FIRE panel once at startup
(`firePanel`) and handed to both the simulator and, later, per-request `/view`
work; they seed the historical models exactly as `-fire <file>` does.

The mux is wrapped by `logAccess`, which prints one NCSA combined-log-format
line per request to stdout (client IP, timestamp, request line, status,
response bytes, referer, user agent, latency); the startup banner and
application errors keep going to stderr, so `pofo -serve >access.log` cleanly
separates the two streams. The client IP honors a left-most `X-Forwarded-For`
entry when present, so the log stays truthful behind a reverse proxy.

### Keeping the two logs readable in production

Two kinds of noise drown a long-lived deployment's logs, and each has exactly
one countermeasure.

**The liveness probe.** The container image's `HEALTHCHECK` polls every 30
seconds, which is one access-log line every 30 seconds saying nothing. `/healthz`
exists so that probe has a route of its own, and `logAccess` skips that ONE path.
The exclusion is by path, never by user agent or source address: a filter on
either could silently hide a request a real visitor made, and crawler 404s (which
are signal) keep being logged like everything else.

**The fetch narration.** The fetch path tells its story as it goes: identifier
resolutions, `history extended via simdata starting …`, cache decisions. That is
the point interactively, where a run prints each line once; a server replays the
same pipeline on every request and the same handful of sentences then scroll
forever. `runServe` and `runFire` therefore install `dedupServerLog`
(`logdedup.go`), a filter in front of the standard logger's destination that
prints each distinct informational line ONCE per process. A line starting with
`warning:` is always printed, however often it repeats: warnings report degraded
data (a held-flat FX rate, a stale cache, a missing benchmark) and their
repetition is itself the signal. Nothing about the lines changes, and the CLI
modes never install the filter, so an interactive run stays as verbose as it was.

The filter sits at the logger rather than at `marketdata.Client.Logf` because the
narration reaches the log through two doors: that field (which `cmd/pofo` wires
to `log.Printf`) and a few direct `log.Printf` calls in `pkg/compare`, the
`resolved X -> name` line among them. Filtering at the client seam would have
caught only half of it. The memory is bounded (`dedupLimit` distinct lines, then
forget everything and start over), since anonymous visitors can mint unbounded
identifiers through the `/view` grammar.

## The `/view` URL grammar

The visualizer is driven entirely by its query string. This is the
authoritative shape, kept in sync with `view.go`'s godoc (the code is the
source of truth; edit both together):

- **`ex=<name>`** selects a bundled example by file base name. Repeat the
  parameter to stack several (`ex=a&ex=b`). Unknown names are rejected.
- **`p=<spec>`** is one ad-hoc portfolio: `ID:WEIGHT` pairs comma-separated,
  optionally followed by `!key:value` meta directives
  (`p=NTSG:60,IGLN:20,IBCI:20!sim:on!name:my dragon`). The `!` delimiter
  replaces the file format's `;` because a raw `;` is not valid in a Go query
  string. `!name:...` sets the portfolio's display name; every other
  `!key:value` becomes a `#meta key:value` line. Repeat `p=` for several
  ad-hoc portfolios. The value is capped at 2000 bytes and control characters
  (a URL-decoded newline in particular) are rejected, since the parser rebuilds
  a line-based portfolio file and a smuggled newline would inject a holding
  line past the catalog gate and the holdings-count limit.
- **Global overrides**, each mirroring the matching CLI flag, layered on the
  server's default options: `start` / `end` (`YYYY-MM-DD`), `rebalance` (day
  count, `0` = never), `currency`, `bench`, `sim` (`on` / `off`). Two of these
  carry attacker-shaped identifiers and are gated before they can reach an
  outbound fetch: `currency` accepts a three-letter ISO code or the sentinel
  `native` (keep each series in its own currency; a present-but-empty
  `currency=` reads the same way); `bench` accepts an empty
  value (disable Beta), any locally resolvable identifier (`KnownLocal`), or
  the exact server-default benchmark symbol (`^GSPC`, which is not "local");
  anything else is a 400, so no arbitrary bytes mint an FX cache file or
  poison the shared quote cache.

Both `ex=` and `p=` build a `portfolio.Spec` by rebuilding the file text form
and feeding `portfolio.Parse`, so the URL grammar can never drift from the file
grammar: `/view` accepts exactly what a portfolio file accepts.

### Guardrails

The composer is meant for a human on a small server, so it is bounded on every
axis: at most **6 portfolios** per page (`ex` + `p` combined), **20 holdings**
each, `p=` value **<= 2000 bytes**, a **60 s** compute timeout per request, and
**2 concurrent** renders (a semaphore; each render is CPU- and fetch-heavy).
The concurrency bound is safe because `marketdata.Client` guards its caches and
its on-disk writes (temp file then rename, each write carrying complete JSON).

The FIRE simulator's `POST /api/*` endpoints carry their own bounds
(`pkg/decumul/web/bounds.go`): body capped at 64 KB, `nPaths` clamped to the
slider's own maximum (10 000) and every year-like field to 100, and the
computations queue behind a small semaphore (half the cores, at least 2), a
request whose client has gone away being refused with 503. Before those
bounds, one posted `nPaths` of a million took 2.4 GB and 22 s of every core,
which on a shared host is a denial of service of everything else on it.

### Catalog-only identifiers for `p=`

An `ex=` file is a vetted build shipped in the binary, so it carries no
identifier restriction. A `p=` spec, by contrast, comes from an anonymous
visitor, and the server must never fetch an arbitrary symbol on their behalf
(an SSRF and abuse vector, and a way to poison the shared quote cache). So
every `p=` identifier is gated by `marketdata.KnownLocal`: catalog ids,
catalog ISINs, aliases and embedded fund tickers resolve (the `SIM` suffix is
allowed); a raw quote symbol or an unknown identifier is rejected before any
network call. The bundled catalog is wide enough to compose real portfolios;
anything outside it is a CLI or portfolio-file job, not an anonymous web one.

## The live composer

Every `/view` report carries a small in-page editor over the `p=` grammar, so
the shareable link never falls out of step with what is on screen. It is
injected under the site nav through the optional `report.Page.Composer` field
(empty for the standalone CLI report, which stays byte-for-byte unchanged) and
served entirely from the binary: `composerMount` (`composer.go`) renders the
panel through `html/template` (portfolio names and identifiers are user input,
so every value, the `encoding/json` data attributes included, rides the
template's contextual escaping), and `/composer.js` / `/composer.css` carry the
front end.

The design rests on a few decisions:

- **The URL is the live state.** Every edit rewrites the query string in place
  via `history.replaceState`, so the address bar is always a faithful,
  copyable link to the current composition, with no server round trip. On a
  result page the charts are the content, so the panel renders collapsed to its
  chip bar (which tracks the live portfolio count) and expanding it is a user
  gesture; the hub's compose surface, being the primary content there, opens.
- **Run, not live compute.** Editing never triggers a fetch or a render.
  Run (or Enter) navigates to the rewritten URL and the server renders the
  report, so the composer adds no compute or fetch surface beyond the existing
  `/view` path. The **server stays authoritative** on every gate; the client
  only mirrors the rules to warn early.
- **Catalog autocomplete and inline validation.** The editor fetches
  `/catalog.json` once and drives id autocomplete and per-row validation from
  it (naming each holding, flagging an unknown id). If that fetch fails it
  degrades to no autocomplete and no client validation; the server still
  rejects anything outside the catalog gate, so correctness never depends on
  the front end.
- **Caps mirrored client-side.** The `/view` guardrails (6 portfolios, 20
  holdings, 2000-byte `p=`) are handed to the front end in a `data-caps`
  attribute (`composerCaps`), which gates add/remove of holdings and
  portfolios and drives a live byte-budget meter, plus a weight-sum badge with
  a one-click normalize. These are affordances; the server enforces the same
  bounds regardless.
- **Fork from an example.** Each `ex=` card is read-only. If at least one of
  its holdings resolves locally, the card offers a Fork affordance carrying a
  `data-fork-<i>` payload (`specToP`, the inverse of `adhocSpec` for what the
  grammar can express): forking swaps that `ex=` for an editable `p=` and
  surfaces a dismissible note listing what could not ride the grammar (holdings
  that do not resolve locally, the comparison-shaping metas `optimize` and
  `currencies`, and any name or meta whose text would break the `!` segment
  grammar). An example whose holdings all drop is not composable and shows no
  Fork button.
- **Presets and Clone, the two ways not to retype a portfolio.** The bundled
  examples the grammar can express ride every mount (hub and `/view` alike) as
  `data-preset-<i>` payloads, so "add preset" drops a whole bundled build in as
  an editable card without a detour through the front door; `viewPresets`
  memoizes the translation, and the payloads add about 9 kB to a page that
  already carries its charts. **Clone** duplicates one editable card in place,
  right below the original, holdings and metas included: comparing a build
  against itself with a single holding swapped (three candidate small-value
  ETFs, say) is the common editing move, and it should not cost a retype. The
  copy's name takes a " copy" suffix when the original has one, so the report
  labels the pair rather than leaning on the server's `X` / `X (2)` dedup.
- **Opaque rows pass through verbatim.** A `p=` value the front end cannot
  parse is not discarded or rewritten: it renders as a locked, read-only
  "manual" row and is passed through unchanged, so a hand-authored or
  future-grammar link always survives a round trip through the editor.

The currency control includes the `native` sentinel, with the empty value
meaning "server default". A `#composer-selftest` URL hash runs the parse and
serialize round-trip self-test in the browser console (the repo is stdlib-only
and carries no JS test harness, so the self-test plus the live smoke stand in
for a unit suite).

One benign edge remains: a half-filled holding row inside an otherwise-filled
portfolio (an id typed with no weight yet, `p=IWDA:`, or a blank row appended
to a filled card, `IWDA:60,:`) rides into the live URL, and Run at that exact
instant would 400. It resolves the moment the row is completed, which is the
next action the row invites. A portfolio with no filled holding at all never
reaches the URL: serialization prunes it (the card stays on screen), so the
hub's freshly-booted empty card cannot poison a Run.

## The composed simulator and the prefs cookie

Two features close the loop between the report and the simulator.

`/firesimulator/p/<spec>/` mounts the FIRE simulator on an ad-hoc composed portfolio.
`<spec>` is exactly the `/view` `p=` grammar carried in a single path segment,
so a composed comparison and its simulator share one vocabulary. The spec is
validated before anything is built: the same catalog gate as `p=`, the 2000-byte
cap, the control-character rejection and the 20-holdings limit all apply up
front, so an anonymous visitor can never make the server fetch an arbitrary
symbol here either. A `!sim:on` directive is honored (the panel splices
simulated history); the panel is built with the server's default currency.
Built handlers live in a small bounded cache (arbitrary eviction past its cap),
and the builds share the `/view` render semaphore, so the composed simulator
adds no new fetch surface or concurrency beyond the visualizer's. The naked
`/firesimulator/e/<name>` and `/firesimulator/p/<spec>` forms 301 to their trailing-slash
canonical.

Every `-serve` FIRE mount is `web.Embedded()` (2026-08): an inner app serves
only itself (the page, its assets, `/api`), never the book mounts or the site
files, which belong to the root that owns them. `/firesimulator/e/<name>/firebook/...`
is therefore a 404 where it used to republish the whole book under a second
URL (a crawler was seen wandering there). The standalone `pofo -fire` server
keeps the full site: it owns its root.

Every FIRE mount also tells the page what it is running on, through two
`pkg/decumul/web` options the front end reads back from `/api/meta`.
`WithSourceLabel` names the market (the example name for `/e/<name>/`,
"custom portfolio" for `/p/<spec>/`, the startup portfolio's name for the
plain mount, the file's base name under `-fire`), which the top bar shows as a
provenance pill: amber "generic market · load portfolio" when no panel is
bound, green "market: <name>" when one is. `WithPicker` hands the front end
the loader payload (`{base, catalogURL, viewURL,
examples:[{name,title,blurb}]}`, sent as `meta.picker`) that fills the drawer's
empty state with the bundled builds and a search-and-weigh composer, and that a
bound mount folds away under its allocation bar ("change portfolio", its draft
seeded from the live holdings). Choosing one is pure navigation to
`<base>/e/<name>/` or `<base>/p/<spec>!sim:on/`, carrying the page's current
hash so the visitor's scenario survives the move; `viewURL` adds the other
reading of a draft, a link to `/view?p=<spec>`, the same grammar this
visualizer's own composer writes. The examples arrive as plain data, so
`pkg/decumul/web` never learns about `examples.List`; only `-serve` sets a
picker, since only it owns those mounts, `/catalog.json` and `/view`, and the
standalone `-fire` page states the command-line route in that slot instead.

Each `/view` report section then carries a **Simulate** link to the matching
mount: an `ex=` section links `/firesimulator/e/<name>/`, a `p=` section links
`/firesimulator/p/<escaped spec>/`. The link is optional in the report template (empty
means omitted, so the standalone CLI report is byte-for-byte unchanged) and only
appears under `-serve`. An optimized portfolio's "as written" twin and its
multi-currency columns share the base spec's link, which is the intended
portfolio.

A small `pofo_prefs` cookie remembers a visitor's non-sensitive preferences
(base currency, default rebalance, sim on/off), each validated field by field,
`HttpOnly`, `SameSite=Lax`, one year. It **pre-fills the hub only**: the hub's
defaults row starts where the visitor left it, and a row's Open link carries the
stored options when the cookie exists. `/view` **writes** the cookie from its
explicit, valid `currency` / `rebalance` / `sim` parameters (merge semantics)
but **never reads** it: a `/view` URL is state entirely on its own, so a shared
link reproduces the same report for everyone regardless of their cookie. The
URL-as-state invariant is preserved.

The "keep native currencies" choice travels end to end as the sentinel
`currency=native`: the hub's native `<option>` submits it, `/view` maps it to
an empty (non-nil) currency override, and the cookie stores it as the empty ISO
code (the codec's internal form). A stored preference that falls outside the
hub's option lists (an ISO code or rebalance cadence the row does not hardcode)
is appended as its own selected option, so the select never silently rewrites
it on submit.

## Style layering

`pkg/webui` owns the shared "instrument" identity (tokens, embedded fonts,
chart chrome; see `docs/webui-instrument-redesign.md`). `-serve` serves those
tokens once from `/theme.css` and `/fonts.css`, and every surface links them.
The reading surfaces then remap the tokens to the book's warm paper-and-ink
palette with `webui.WarmSkin` (one CSS file, variable overrides only), so the
constellation reads as the book's kin while the simulator keeps the cool
instrument look:

- The **hub** links the tokens, then inlines `webui.WarmSkin` and sets its
  headings in the book's serif: the book's calm reading rhythm in the book's
  own palette, no JavaScript.
- The **report** gets the same warm skin under `-serve` only, through the
  optional `report.Page.SkinCSS`/`SiteNav` fields (empty for the CLI, so the
  standalone report is byte-for-byte unchanged), plus a slim cross-nav bar.
- The **simulator** keeps its instrument-dark layer. It darkens each chart
  through `pkg/decumul/web`'s own wrappers (`theme.go`), not the `chart`
  process-global, so it stays dark even sharing a process with the light
  report under `-serve`.
- The **book** keeps its own reading layer. Its default `Handler()` stays
  chrome-free for offline and `-fire` use; under `-serve` it is mounted with
  `firebook.WithNav`, which adds a **print-hidden** nav bar (chrome, not
  content) linking back to the hub and the simulator. Each mount gets the
  labels of its own edition ("Portefeuilles"/"Simulateur" on the French book,
  "Portfolios"/"Simulator" on the English one), and the language switch
  `WithAlternate` appends closes the bar.

## Being indexed, and being quoted

The book is the one surface here written for strangers, so it states what it is
twice: once for search engines, once for the agents that answer questions by
quoting sources. The machinery is `pkg/firebook` (per-page metadata, the three
root files) over `pkg/seo` (the file formats themselves, book-agnostic).

For search engines, every book page carries a title and a meta description
(the manifest blurb, one per article, guarded non-empty and distinct by a
test), Open Graph and Twitter-card metas, a canonical link, the hreflang pair
and its x-default, and schema.org JSON-LD: `WebSite` + `Book` (with the EPUB as
a `workExample` and the whole table of contents as `Chapter` parts) on an
index, `Article` + `BreadcrumbList` on a page. `/sitemap.xml` lists every
article, index and EPUB of both editions plus the app surfaces, and
`/robots.txt` points at it. No `lastmod` in the sitemap: the articles are
embedded in the binary with no honest modification date.

Shared rather than searched, a link to the book now shows a **social card**:
one image per edition at `/firebook/<lang>/card.png`, 1200x630, emitted as
`og:image` (absolute, with `og:image:width`/`height`/`alt`) and paired with a
`summary_large_image` Twitter card. The drawing is Go code like every other
plate, `firebook.(*Edition).CardSVG` in `card.go`: the book's own hero block at
card size in the v2 plate identity (paper ground, letterspaced kicker, serif
title, the accent rule an article heading carries, sans deck). Every word on it
comes from the `Edition` value already, none of it written for the card: the
publisher mark, `SiteName`, the head clause of `SiteLede` as the promise, its
tail clause as the contents line, and `UI.SwitchLabel` as the edition marker.

Rasterizing SVG needs a browser, so the PNG is generated once by
`scripts/card-shot.sh` (headless Chrome, the book's embedded fonts, exactly
1200x630) and committed under `pkg/firebook/assets/cards/`, which the binary
embeds. Run the script after any change to `CardSVG` or to an edition's title
or lede, and look at the two PNGs. A guard test checks the embedded images are
real PNGs at the declared size, that the SVG carries the edition's own words,
and that the pages' `og:image` points at the served route.

For readers who would rather follow than search, each edition publishes an
**Atom feed** at `/firebook/<lang>/feed.xml`, one entry per article in reading
order with its manifest blurb as the summary, declared from every page's head
(`rel="alternate" type="application/atom+xml"`) and listed in `llms.txt`. Atom
makes `<updated>` mandatory, and the same problem that keeps `lastmod` out of
the sitemap applies here: no article has an honest date. So one stamp answers
for the feed and for every entry of it, the mount's single publication time
(`built` in `Handler`, taken once when the handler is assembled), which is also
what the EPUB writes into `dcterms:modified` and what the OPDS catalog carries.
The entry ids are derived from `Edition.EPUBIdentifier`, not from the URLs, so
they do not move when the same book is reached under another host name. A
guard test holds the one-stamp invariant.

For AI agents the decisive move is the **Markdown mirror**: `<article>.md`
serves the article's source exactly as written, callouts, tables and
`[[wiki-links]]` included, behind a single HTML-comment line naming the page to
cite. An agent quoting clean Markdown gets the text right; an agent scraping
rendered HTML gets the furniture too. The mirror is declared from the HTML head
(`rel="alternate" type="text/markdown"`), inside the article's structured data
(`encoding`), and in `/llms.txt`, which follows the llmstxt.org convention: one
H1, one summary paragraph, then a section per edition (index, EPUB, OPDS) and
one per part of that edition listing every article by its Markdown URL. Both
editions live in ONE file, sections labelled by language: an agent fetches one
`llms.txt` per host, and splitting the languages would hide each edition from
whoever found the other. The mirrors are deliberately absent from the sitemap,
where they would read as duplicate content.

### IndexNow: pushing after a deploy

Everything above waits to be crawled. **IndexNow** (indexnow.org) is the other
direction: one POST tells the participating engines (Bing, Yandex, Seznam,
Naver, and whoever else shares the endpoint) which URLs exist, and they fetch
what they do not have.

Ownership is proved by a **key file**: a text file at the root of the host whose
name is the key and whose body is that same key. The key is therefore **per
host**, not per binary and not per build: a key minted for one domain proves
nothing about another, and whoever holds it can submit URLs for that host. Mint
one unguessable value (8 to 128 letters, digits and dashes, e.g. `uuidgen | tr
-d - | tr 'A-Z' 'a-z'`), keep it with the deployment secrets, and hand it to the
server:

```sh
./pofo -serve -listen 127.0.0.1:8787 -indexnow-key 1a2b3c4d5e6f7890...
```

That mounts `/1a2b3c4d5e6f7890....txt` and nothing else changes. Without the
flag the feature is entirely off, which is what every local run and every
`-fire` mount wants. `pkg/decumul/web` takes the same key through
`web.WithIndexNowKey`.

Then, **after a deploy**, from any machine that can reach the endpoint:

```sh
./pofo -indexnow https://pofo.example.org -indexnow-key 1a2b3c4d5e6f7890...
```

The key may also come from the environment as POFO_INDEXNOW_KEY, read
whenever the flag is empty: a container image with a fixed command line
turns the feature on by setting that variable.

### Audience: the optional analytics beacon

Being indexed says nothing about being read. **Cloudflare Web Analytics** is the
one measurement the deployment may switch on, and it follows the IndexNow
pattern exactly: a value that belongs to the public host, not to the program.

```sh
./pofo -serve -listen 127.0.0.1:8787 -cf-beacon-token 0123456789abcdef...
```

An empty flag defers to `POFO_CF_BEACON_TOKEN`, for the same fixed-command-line
reason as the key above. **Both empty, the feature is entirely off and every
page is byte-identical to what it was before it existed**: no markup, no request,
no name. `-fire` takes the same flag, and `pkg/decumul/web` the same value
through `web.WithBeaconToken`, the twin of `WithIndexNowKey`.

The tag is the one Cloudflare documents, spliced in front of the closing
`</body>`:

```html
<script defer src="https://static.cloudflareinsights.com/beacon.min.js"
        data-cf-beacon='{"token":"..."}'></script>
```

**Coverage is the point**, and it is why the injection is a response filter
(`webui.Beacon`) rather than a line in a template. The constellation's HTML
comes out of four independent renderers that share design tokens but no
head-or-foot helper: the landing and hub templates in `cmd/pofo`, the comparison
report in `pkg/report`, the book in `pkg/firebook`, the simulator page in
`pkg/decumul/web`. Wrapping the finished mux is the only seam that sees all of
them, so the tag lands on the landing page, the visualizer, every `/view`
report, the shared error page, every page of both book editions and every
simulator mount, with no renderer aware of it. `-serve` wraps from the outside
and therefore passes no `WithBeaconToken` to the mounts it makes; the injector
is idempotent anyway, so a nested pair still leaves exactly one tag.

Only a `text/html` response is rewritten, and only at a status a visitor lands
on (200, and the 4xx/5xx error pages, where a broken link is exactly what
deserves counting). The Markdown mirrors, the Atom feeds, the sitemap,
`robots.txt`, `llms.txt`, `catalog.json`, the stylesheets, the SVG favicon, the
social cards, the EPUB, the raw portfolio files, the redirect stubs and every
304 stream through untouched, which leaves the strong ETags those routes compute
matching the bytes they send. A rewritten response drops its `Content-Length`,
since the body is longer than the inner handler announced.

The token is operator-supplied and escaped anyway: `encoding/json` handles `<`,
`>`, `&`, the double quote and the control characters, and the single quote that
would close the attribute is replaced.

**No consent banner, deliberately.** The beacon sets no cookie and stores nothing
on the visitor's device, which is the condition under which the CNIL exempts
audience measurement from prior consent; there is no identifier to refuse.
Nothing about this appears in the book, which is about retirement withdrawal and
not about this server. No other analytics of any kind exists in the program, and
with no token there is none at all.

**No Content-Security-Policy** is set anywhere in pofo: no handler emits the
header and no page carries the equivalent `<meta>`. Nothing therefore had to be
widened for the beacon. Should one ever be added, it needs
`https://static.cloudflareinsights.com` in `script-src` (the loader) and
`https://cloudflareinsights.com` in `connect-src` (where the beacon POSTs its
page-view record).

It builds the same `firebook.Site` the server serves, takes its sitemap URL
list (`Site.URLs`, which `SitemapXML` renders: one list, so what is pushed and
what is crawlable cannot drift), and POSTs it as one batch per 10 000 URLs. A
200 or a 202 is success; 202 simply means the endpoint has not read the key file
yet. Any other status is reported with the endpoint's own words, which say more
than the code (403 is a key it could not verify, 422 URLs that do not match the
host).

This is the **only** place in the program that talks to a search engine, and it
does so only when asked by name: a running server never submits anything, on a
schedule or otherwise. The protocol lives in `pkg/seo` (`IndexNow`, endpoint an
argument), so every test drives it against an `httptest` server and the suite
never leaves the machine.

`robots.txt` allows every page and then names GPTBot, OAI-SearchBot,
ChatGPT-User, ClaudeBot, Claude-Web, PerplexityBot, Google-Extended and CCBot
one by one, with a comment saying why: the book wants to be indexed, quoted and
cited, there is no paywall, and the thing to cite is the page. The one
exception (2026-08) is `Disallow: /view`, carried by `firebook.Site.Disallow`
and repeated in the named-crawler record (a named record replaces the
catch-all for its agents, so a rule left only on `*` would never reach them):
`/view` is a compute endpoint whose every hit fetches quotes and runs
simulations over an unbounded parameter space, not a page to index. People
keep sharing `/view` links; crawlers are told not to walk them.

The three files are dynamic because the public origin is not knowable at build
time (the same binary answers on localhost, on a tailnet name and behind a
proxy): `firebook.RequestOrigin` reads the request's own `Host`, and only the
scheme is taken from `X-Forwarded-Proto`, never the host, so an untrusted
client cannot mint a sitemap pointing somewhere else.

## Perimeter

The server binds `127.0.0.1:8787` by default (`-listen` to change it). The hub
footer states the contract plainly: everything runs on this machine, no
portfolio leaves it. To reach the app from another device, the intended path is
`tailscale serve 8787`, which publishes it over the tailnet under HTTPS without
opening a public port. Binding a non-loopback address is possible but is the
user's explicit choice, not the default.

## Milestone ladder

The web app shipped as a read-mostly constellation. The planned follow-ups,
smallest lever first:

- **M2: per-request FIRE panel + a user-settings cookie.** Shipped
  (2026-07-19). `/firesimulator/e/<name>/` builds a panel per bundled example on demand,
  `/firesimulator/p/<spec>/` generalizes it to an arbitrary composed portfolio (a panel
  from a `p=` spec, catalog-gated, lazily and boundedly cached), every `/view`
  section carries a Simulate link to its mount, and the `pofo_prefs` cookie
  remembers a visitor's non-sensitive preferences (base currency, default
  rebalance, sim on/off) to pre-fill the hub, so the composer opens where they
  left it. See "The composed simulator and the prefs cookie" above.
- **M3: a live composer that writes the URL.** Shipped (2026-07-19). A small
  in-page editor for the `p=` spec (add/remove holdings and portfolios, edit
  weights with a sum badge and normalize, catalog autocomplete, Fork from an
  example) that rewrites the query string as you go via `history.replaceState`,
  so the shareable link is always the current state, and Run navigates to let
  the server render. No new server capability beyond the `/catalog.json`
  read-only endpoint, just a front end over the existing grammar. See "The live
  composer" above.
- **M4: extract the report assembly into `pkg/`.** Shipped (2026-07-20),
  completing the ladder. The reusable pipeline moved into `pkg/compare` (the way
  `FetchExtended` and `portfolio.Build` were extracted): `compare.Compute`
  fetches, builds, simulates, aligns the common window and computes the
  nominal/real statistics, and `Comparison.HTMLPage` assembles the report
  `Page`. `/view` renders through it, so any server, not just this CLI, can
  produce the comparison report. `renderComparison` in `cmd/pofo` is now a thin
  mapper (`opt.compareOptions()` -> `compare.Compute` -> `HTMLPage(opt.decoration())`
  -> `report.Render`); the web chrome (skin CSS, site nav, composer, FIRE href)
  arrives through `compare.Decoration`, keeping the standalone CLI report
  byte-identical. `cmd/pofo` is back to wiring.
