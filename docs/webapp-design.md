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
| `/` | `hub` (`hub.go`) | the front door: the bundled example portfolios as a pure-GET checkbox form that submits ticked names to `/view`, plus links onward to the simulator and the book |
| `/view` | `view` (`serve.go`, grammar in `view.go`) | the HTML comparison report the CLI writes, addressed by a query string (`ex=` / `p=` + global overrides) |
| `/examples/<name>.txt` | `exampleFile` | one embedded portfolio file, raw text (the hub's "Source" link) |
| `/fire/` | `fire` (`serve.go`) -> `pkg/decumul/web.Handler`, prefix-stripped | the FIRE simulator on the startup panel, identical to `-fire` |
| `/fire/e/<name>/` | `fire` -> a per-example `web.Handler` | the simulator pre-loaded with one example's historical panel (the hub's "Simulate" link), built and cached lazily on first use |
| `/fire/p/<spec>/` | `fire` -> a per-spec `web.Handler` | the simulator bound to an ad-hoc composed portfolio, `<spec>` being exactly the `p=` grammar in one path segment; catalog-gated, bounded lazily-built cache |
| `/book/fr/` | `pkg/firebook.Handler`, prefix-stripped | the French FIRE book, with a chrome nav bar back to the other surfaces |
| `/theme.css`, `/fonts.css` | inline | the shared `pkg/webui` identity tokens and embedded fonts |

The mux (`server.handler` in `serve.go`) is a plain `http.ServeMux`; the
lifecycle (`runServe`) mirrors `runFire`: bind, serve, shut down on context
cancel. Portfolio file arguments are turned into a FIRE panel once at startup
(`firePanel`) and handed to both the simulator and, later, per-request `/view`
work; they seed the historical models exactly as `-fire <file>` does.

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

## The composed simulator and the prefs cookie

Two features close the loop between the report and the simulator.

`/fire/p/<spec>/` mounts the FIRE simulator on an ad-hoc composed portfolio.
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
`/fire/e/<name>` and `/fire/p/<spec>` forms 301 to their trailing-slash
canonical.

Each `/view` report section then carries a **Simulate** link to the matching
mount: an `ex=` section links `/fire/e/<name>/`, a `p=` section links
`/fire/p/<escaped spec>/`. The link is optional in the report template (empty
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
  content) linking back to the hub and the simulator. The labels are French
  ("Portefeuilles", "Simulateur") because the navbar sits on the French book;
  when an English book joins, its mount gets English labels.

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
  (2026-07-19). `/fire/e/<name>/` builds a panel per bundled example on demand,
  `/fire/p/<spec>/` generalizes it to an arbitrary composed portfolio (a panel
  from a `p=` spec, catalog-gated, lazily and boundedly cached), every `/view`
  section carries a Simulate link to its mount, and the `pofo_prefs` cookie
  remembers a visitor's non-sensitive preferences (base currency, default
  rebalance, sim on/off) to pre-fill the hub, so the composer opens where they
  left it. See "The composed simulator and the prefs cookie" above.
- **M3: a live composer that writes the URL.** A small in-page editor for the
  `p=` spec (add/remove holdings, drag weights, toggle meta) that updates the
  query string as you go, so the shareable link is always the current state.
  No new server capability, just a front end over the existing grammar.
- **M4: extract the report assembly into `pkg/`.** `/view` currently reaches
  into `cmd/pofo`'s report-assembly path (`renderComparison` and friends).
  Pushing that reusable pipeline down into a package (the way `FetchExtended`
  and `portfolio.Build` were extracted) lets any server, not just this CLI,
  render the comparison report, and shrinks `cmd/pofo` back to wiring.
