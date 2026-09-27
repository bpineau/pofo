# The web app

`pofo -serve` runs the whole tool as one local web app: a portfolio
visualizer, the FIRE simulator and the book. A public instance runs at
[pofo.zouh.org](https://pofo.zouh.org).

```sh
./pofo -serve                               # http://127.0.0.1:8787/
./pofo -serve -listen 127.0.0.1:9000        # another port
./pofo -serve examples/portfolios/fire-decumulation-core.txt   # seed the FIRE models from a file
```

It binds to loopback by default. To reach it from a phone or another machine,
put it behind a private network rather than opening a port, for instance
`tailscale serve 8787`.

## What it serves

| URL | Surface |
|---|---|
| `/` | the landing page |
| `/visualizer` | the portfolio visualizer: compose portfolios or tick bundled examples |
| `/view?...` | a comparison report, addressed by a shareable URL |
| `/firesimulator/` | the FIRE simulator |
| `/firebook/fr/`, `/firebook/en/` | the book, "Le FIRE tranquille" and "The Quiet FIRE" |
| `/healthz` | a liveness probe (`200 ok`), kept out of the access log |

## The visualizer and share URLs

The visualizer renders the same HTML report as the command line, but every
comparison is a URL, so you can bookmark or share it:

```text
/view?ex=dragon-decumulation-household&ex=golden-butterfly
/view?p=NTSG:60,IGLN:20,IBCI:20!sim:on&currency=EUR
```

| Parameter | Meaning |
|---|---|
| `ex=NAME` | a bundled example portfolio; repeat it to add more |
| `p=ID:W,ID:W!key:value` | an ad hoc portfolio; `!` appends a `#meta` directive |
| `start`, `end` | the window, `YYYY-MM-DD` |
| `rebalance` | days between rebalances |
| `currency` | an ISO code, or `native` to keep each series' own currency |
| `bench` | a catalog identifier, or empty to drop beta |
| `sim` | `on` backcasts every holding |

A page holds up to six portfolios of twenty holdings each.

Every report carries a **composer**: edit holdings and weights in place, fork
an example into an editable copy, and the URL follows every change. Each
portfolio also has a **Simulate** link that opens the FIRE simulator on it.
The visualizer's home remembers your default currency, rebalance and sim
settings in a cookie; a `/view` link stays self-contained.

### Which identifiers a visitor may use

Anything the bundled catalog resolves (ids, ISINs, aliases, fund tickers, with
or without `SIM`) is free and unlimited. Anything else is fetched from the
quote sources, under two guards:

- it must look like an instrument: a valid ISIN, or a plausible exchange
  ticker such as `DGRO` or `IWDA.AS`;
- it counts against two rolling hourly budgets, one per visitor and one for
  the whole server.

```sh
./pofo -serve -serve-foreign-per-hour 10 -serve-foreign-global-per-hour 60   # the defaults
./pofo -serve -serve-foreign-per-hour 0                                      # catalog only
```

An identifier already in the quote cache costs nothing. Such identifiers are
resolved exactly, with no fuzzy search, so a typo fails instead of quoting an
unrelated fund. An identifier no source knows answers `404`; a spent budget
answers `429`.

## The FIRE simulator

`/firesimulator/` is the page `pofo -fire` opens: sliders for capital,
spending, horizon, pension, spending rules and taxes, with ruin probabilities
under several return models. Started with portfolio files, the server feeds
their history to the simulator's historical models. The
[FIRE guide](fire.md) explains how to read it.

## The book

The book is a handbook of living off one's capital, written in French and
translated in full into English. Every page links to its counterpart in the
other edition.

| Format | Where |
|---|---|
| web pages | `/firebook/fr/`, `/firebook/en/` |
| EPUB 3 | `le-fire-tranquille.epub`, `the-quiet-fire.epub` on each mount, or `pofo -export-epub FILE [-book-lang en]` |
| OPDS catalog | `opds.xml` on each mount: add it once to an e-reader such as KOReader, then refresh the book in place |
| Atom feed | `feed.xml` on each mount, one entry per article |
| Markdown | any article's URL plus `.md`, the source as written |

The server also publishes `/sitemap.xml`, a permissive `/robots.txt` and an
[`/llms.txt`](https://llmstxt.org) index of both editions.

## Running a public instance

A container image builds from the repository:

```sh
DOCKER_BUILDKIT=1 docker build -f deploy/docker/Dockerfile -t pofo:latest .
docker run --rm -p 127.0.0.1:8787:8787 -v pofo-cache:/var/lib/pofo pofo:latest
```

It serves `-serve` on port 8787 with the quote cache on a volume, and probes
`/healthz`. Publish it behind whatever edge you run.

Two optional features exist for a public deploy, both off by default:

| Feature | How |
|---|---|
| push new URLs to search engines ([IndexNow](https://www.indexnow.org/)) | serve a key with `-serve -indexnow-key KEY` (it publishes `/KEY.txt`), then run `pofo -indexnow https://your.origin -indexnow-key KEY` after each deploy |
| count readers with [Cloudflare Web Analytics](https://developers.cloudflare.com/web-analytics/), cookieless | `-serve -cf-beacon-token TOKEN`, or `POFO_CF_BEACON_TOKEN` in the environment |

A running server never contacts a search engine on its own, and without a
beacon token the pages carry no analytics of any kind.
