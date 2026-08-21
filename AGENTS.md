# pofo, for coding agents

pofo is a dependency-free (stdlib-only) Go toolkit for tracking and designing
investment portfolios, plus one CLI (`cmd/pofo`) built on it. Everything the
CLI does is reachable as a library under `pkg/`.

Read this file first; it is the cheapest way in. Details live in each
package's `doc.go` (`go doc ./pkg/<name>` renders it) and in `README.md`
(user-facing, CLI-oriented).

## Commands

```sh
make build     # ./pofo binary, pkg/datasets/ embedded via go:embed
make test      # go test ./...  (unit tests + runnable examples, NO network)
make lint      # go vet + staticcheck
make check     # fmt-check + lint + test: run this before any commit
make golden    # computation goldens vs frozen external references
make refresh   # refresh EVERY bundled series from its live source, in order (network)
make simdata   # regenerate pkg/datasets/simdata/ (network) then rebuild
make broadsample # regenerate the JST broad-sample panel (network) then rebuild
make cape      # regenerate the Shiller CAPE series (network) then rebuild
make macropanel # regenerate the OECD monthly macro panel (network) then rebuild
make euro-refdata # regenerate the euro-area reference series (network) then rebuild
make gbond-refdata # regenerate the German/Japanese/British govt bond reference series (network); run make simdata after
make dbi-refdata # regenerate the DBi family's nearest donor (the all-styles composite half-projected on the fund's ten futures); run after make sgtrend-refdata, then make simdata
make sp500-refdata # regenerate the month-end SP500-USD reference (network); run make simdata after
make wti-refdata # regenerate WTI-ER-USD, the rolled-futures EXCESS return of WTI crude (network)
make trend-refdata # regenerate the monthly trend reference (network); run make simdata after
make trendnet-refdata # regenerate the monthly NET managed-futures reference (network); run make simdata after
make sgtrend-refdata # regenerate the daily NET pure-trend reference (network); run make simdata after
make catbond-refdata # regenerate the monthly NET insurance-linked reference (network); run make simdata after
make snapshots # regenerate pkg/marketdata/data/'s offline fallback snapshots (network)
make simdata-qa # reconstruction quality: every engine vs the real quotes, HTML (network)
make verify-catalog # data doctor over the whole catalog: hygiene, plausibility bands, identity (network)
make book-drift # what the FIRE book's translations owe their French source
make figure-drift # what the FIRE book's frozen figures owe the bundled data
```

`make refresh` runs every generator below it in dependency order (references
first, then the simdata built on them, then the offline snapshots); the
individual targets are for touching one series. Follow it with `make check` and
`make golden`, both of which must stay green (refreshing data must never require
a code change), then `make verify-catalog`, which says what the new quotes look
like.

The FIRE book's plates freeze numbers read off the bundled datasets, so a
refresh does move some of them. Those recomputation checks are therefore kept
out of `make check` by `frozenAgainstData` (`pkg/firebook/figures_drift_test.go`)
and run only under `make figure-drift`, which names the plate and the literal to
update. The book is allowed to lag its data; catching it up is a separate,
deliberate job. The plates' rendering, wording and house-rule tests are not
gated and keep running everywhere.

Tests never touch the network: HTTP sources are faked with `httptest`
(`stubAllBases` in `pkg/marketdata/client_test.go`), file sources with
`fstest.MapFS`. Keep it that way.

## Verifying changes cheaply

- Report/chart changes: `scripts/report-shot.sh [file] [out-prefix]` builds,
  renders the report with every section unfolded and screenshots it full-page
  (needs Chrome and a warm quote cache; `./pofo -warmup` once). Crop a region
  with `sips -c <height> 1500 --cropOffset <y> 0 out.png --out crop.png` and
  read the PNG. `examples/dragon-decumulation-household.txt` is the reference
  fixture: 9 holdings across a stacked 90/60, long bonds, two trend funds,
  gold, linkers and small-value equity, so it exercises every report block.
- Chart hover/tooltip data: grep the rendered HTML for
  `<metadata class="hover">` and replay the front-end math on the JSON
  payload in node/python instead of driving a browser.
- Is a backcast any good? `./pofo -verify-simdata [ID...]` replays each engine
  without its real graft, against the real quotes, and opens an HTML report:
  level and path verdicts, drift panel, donor-chain junctions. This is the
  cheapest way to see what a recipe change did.
- One golden at a time: `go test -v -run TestGoldenGold ./pkg/datasets/golden/`.
  Chart snapshots moved on purpose? Regenerate with
  `UPDATE_SNAPSHOTS=1 go test ./pkg/chart -run TestChartSnapshots` and justify
  the diff in the commit message.
- Catalog edits: `make test` revalidates `assets.json`;
  `./pofo -verify-data -assets <id>` checks a single asset end to end, and
  `make verify-catalog` runs the doctor over all of it (plausibility bands per
  `asset_class`, identity vs the record); run it after any catalog edit or
  `make refresh`.

## Map

| Path | What lives there |
|---|---|
| `pkg/marketdata` | fetch/cache daily + intraday prices; identifier resolution (alias, ticker, ISIN); FX conversion; SIM history extension; data doctor (`Verify`) |
| `pkg/metrics` | risk/return statistics on dated value series (CAGR, Sharpe, drawdowns, IRR, variance ratio, rolling, CWARP) plus per-holding attribution (`Attribute`: Euler risk shares + realized return shares from a simulation's contributions) |
| `pkg/portfolio` | portfolio file format (`Parse`), `Build` (spec + fetch callback -> Portfolio), `Simulate` (rebalancing, fees, flows, leverage, per-holding return attribution incl. monthly folding) |
| `pkg/optimize` | long-only weights: max-sharpe, min-volatility, max-return, risk-parity, max-sortino, return-to-drawdown, min-ulcer, max-worst-5y, cwarp; per-line bounds (`min-weight`, `bounds:ID:LO-HI`) and feasibility limits (`max-vol`, `min-return`, `max-drawdown`) route every objective through one penalized box-simplex search; `train:` is parsed here and applied by the caller (see `docs/weight-search-design.md`) |
| `pkg/permanent` | tactical Permanent Portfolio 2.0 (Darcet): reads `datasets.MacroPanel` into a growth×inflation + monetary regime, quadratically-damped four-sleeve allocation, monthly-real backtest, coarse `Regime.Quadrant` view (used by the report's regime strip); see `docs/darcet-permanent-portfolio-design.md` |
| `pkg/suggest` | macro-regime/factor coverage, look-through composition splits (asset classes, geography, currency exposure, equity sectors, duration), redundancy, gap-filling suggestions |
| `pkg/scenario` | synthetic real-return paths: parametric Student-t, block/stationary bootstrap, historical cohorts, behind one `Source` interface |
| `pkg/decumul` | withdrawal/FIRE engine over a `scenario.Source`: ruin probability, outcome metrics, solvers, sweeps; optional STOCHASTIC LIFETIME (`Plan.Lifetime` draws the household's lifespan per path: alive-ruin, estate at death, couple reversion, `Plan.Annuity` realising mortality credits; `docs/stochastic-lifetime-kernel-design.md`); `web/` = embedded live UI |
| `pkg/replay` | the seven canonical withdrawal rules (shared names/tags/colours/plan mutations) run deterministically over the years as they happened, on a bundled real US 60/40 (S&P 500 + 5y Treasuries, CPI-deflated, from 1954); portraits of a rule (income mean/CV/leanest year/lean years/estate), not failure probabilities; feeds the FIRE book's `sept-facons-de-vivre` article and the simulator's policy frontier |
| `pkg/bookmd` | the shared book-Markdown-dialect renderer (`ToHTML`, callouts, wiki-links, figures), extracted from firebook so other repos can reuse it; see `docs/epub-export-design.md` |
| `pkg/epub` | generic stdlib-only EPUB 3 writer (`Book`/`Chapter` in, `.epub` bytes out) + `Normalize` (HTML5 -> XHTML); deterministic output for a given `Modified`; see `docs/epub-export-design.md` |
| `pkg/opds` | generic stdlib-only OPDS 1.2 acquisition-feed builder (`Feed`/`Entry` in, Atom `.xml` bytes out); book-agnostic, relative acquisition links, deterministic output for fixed `Updated`; consumed by firebook's `opds.xml` route so KOReader can add the catalog once and refresh the book in place |
| `pkg/seo` | generic stdlib-only builders for the machine-readable files a site publishes: `Sitemap` (sitemaps.org urlset), `Robots` (robots.txt records), `LLMs` (llms.txt, the llmstxt.org convention), `Feed` (Atom 1.0 syndication) and `IndexNow` (the push protocol: `Validate`/`Bodies`/`Submit`, endpoint an argument so tests use `httptest`); content-agnostic and deterministic, assembled by `firebook.Site` |
| `pkg/firebook` | the FIRE book "Le FIRE tranquille": embedded French decumulation handbook (markdown articles under `assets/book/fr/` + manifest + renderer + handler with per-page SEO metadata), served by the fire UI at `/firebook/fr/` (old `/book/fr/` 301-redirects); `epub.go` assembles the whole book as an EPUB 3 (`EPUB`, served at `le-fire-tranquille.epub` and by `pofo -export-epub`), and the handler serves an OPDS 1.2 catalog at `opds.xml` (built via `pkg/opds`, relative acquisition link, shares the lazy EPUB build time) for KOReader; everything renders through an `Edition` value (`French`, plus `English` = "The Quiet FIRE", COMPLETE since 2026-08-19: 82 translated articles under `assets/book/en/`, mounted at `/firebook/en/`, its own `the-quiet-fire.epub`, a hard completeness guard in `manifest_en_test.go`, and `WithAlternate` hreflang cross-links between the two mounts; `pofo -export-epub -book-lang fr|en` writes either edition) and the package-level API is a thin French wrapper; `Drift` / `pofo -book-drift` / `make book-drift` report what a translation owes, via per-article source stamps; SEO/GEO lives in `site.go` (`BookSite`/`Site` renders `/sitemap.xml`, `/robots.txt`, `/llms.txt` and the optional IndexNow key file `/<key>.txt` over `pkg/seo`, mounted at the root of BOTH servers; `Site.URLs` is the one list the sitemap renders and `pofo -indexnow` pushes) and in the handler's head (canonical, `og:url`, hreflang + x-default and `og:image`, all FULLY QUALIFIED via `Edition.absolute` = `RequestOrigin` + `HomePath`; Open Graph, JSON-LD `Book`/`Article`/`BreadcrumbList`), plus the Markdown mirror route `<slug>.md` that serves an article's source untouched for AI agents, an Atom feed per edition at `feed.xml` (`feed.go`, one entry per article, `<updated>` = the mount's single publication stamp, the same one the EPUB and OPDS carry) and the social card at `card.png` (`card.go`: `CardSVG` draws the book's hero block at 1200x630 in the v2 plate identity from the edition's own words, `scripts/card-shot.sh` rasterizes it into the committed `assets/cards/<lang>.png`); plan, depth conventions and progress ledger in `docs/fire-book-design.md`, English edition in `docs/fire-book-en-edition-design.md` |
| `pkg/simgen` | rebuilds the missing past of complex assets (composites, TSMOM engine, donor chains) into simdata files; `audit.go` grades every engine against the real quotes (`Audit`/`AuditAll`, behind `pofo -verify-simdata`) |
| `pkg/chart` | stdlib-only SVG + terminal charts |
| `pkg/report` | HTML/text rendering of the comparison model |
| `pkg/compare` | `Sweep` (per-holding weight grid, the evidence behind a file's sane ranges, behind `pofo -sweep`); compute the comparison model (fetch, build, simulate, common window, nominal/real stats) and assemble the HTML report `Page`; presentation-neutral, web chrome arrives via `Decoration`, terminal output via `Columns`/`StatRows`; shared by the CLI and `-serve` |
| `pkg/datasets` | embedded data: `assetmeta/assets.json` catalog, `simdata/` CSVs, `refdata/` (incl. `ILS-NET-USD`, the monthly net insurance-linked composite, and `WTI-ER-USD`, the daily EXCESS return of a rolled long WTI futures position, 1985-2024, which prices the roll the spot series `WTI-USD`/`WTI-DAILY` cannot), `broadsample/` (JST per-country real returns for the FIRE empirical model), `cape/` (Shiller CAPE, FIRE valuation anchor), `macropanel/` (OECD monthly multi-country macro drivers: IP/CPI/rates/share prices, for regime & growth-inflation-breadth work), `golden/` (frozen-fixture tests) |
| `cmd/pofo` | wiring over `pkg/compare`, one file per concern: `main.go` (flags + mode dispatch + terminal output + `renderComparison`), `fetch.go`, `adapt.go` (maps `options` onto `compare.Options`/`Decoration`), `suggest.go`, `simdata.go`, `sweep.go` (`-sweep`), `fire.go`, `permanent.go`, `epubexport.go` (`-export-epub`: writes the FIRE book EPUB) (the report-assembly files `page.go`/`composition.go`/`contrib.go` moved into `pkg/compare`); the `-serve` web constellation is `serve.go` (mux + lifecycle), `landing.go` (the front-door landing page at `/`), `hub.go` (the portfolio visualizer's home at `/visualizer`), `view.go` (the shareable `/view` URL grammar), `prefs.go` (the settings cookie), `composer.go` (+ `composer.js`/`composer.css`: the live in-page editor over the `/view` grammar, fed by the `/catalog.json` endpoint `serve.go` exposes) and `logdedup.go` (log hygiene for the long-lived servers: each informational fetch line once per process, every `warning:` always; `/healthz` and the access log live in `serve.go`) |
| `docs/` | design docs and plans, one per feature; read before reworking a feature (`docs/README.md` is the one-line index) |
| `examples/` | portfolio files for the CLI (also exercised by `make demo`); `embed.go` embeds them (`go:embed *.txt`) and lists them (`List`) so `-serve` can build the hub catalog and serve each file raw at `/examples/<name>.txt` |

Root `doc.go` describes the layering and the typical pipeline.

## The core pipeline (library)

```go
ctx := context.Background()
client := marketdata.NewClient(marketdata.DefaultCacheDir()) // "" = no disk cache
spec, _ := portfolio.ParseFile("p.txt")
p, _ := portfolio.Build(spec, portfolio.BuildOptions{
    Fetch: func(id string) (*marketdata.Series, error) {
        return client.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "EUR"})
    },
    Fees: func(id string) (float64, bool) { base, _ := marketdata.SplitSim(id); return client.Fees(ctx, base) },
})
sim, _ := portfolio.Simulate(p, 90)          // rebalance every 90 days
stats, _ := metrics.Compute(sim.Dates, sim.Index)
```

Every step is also reachable individually (`Fetch`, `ReadSimdataFS`,
`ExtendBack`, `ConvertCurrency`, `Trim`, ...) when a caller needs to deviate.

## Conventions and traps (do not guess, check here)

- UNITS, the number one trap. Fees and rates mix two conventions:
  - PERCENT per year: `portfolio.Holding.Fees`, `Portfolio.EnvelopeFees`,
    `Portfolio.BorrowSpread`, `marketdata.Client.Fees` (0.85 = 0.85 %/yr).
  - FRACTION per year: everything in `pkg/simgen` (fees, vol targets:
    0.0085 = 0.85 %/yr), all of `pkg/metrics` outputs except `Stats.Ulcer`
    (percent points), returns everywhere (0.04 = +4 %).
  - Weights: FRACTIONS in `portfolio.Asset.Weight`/`Holding.Weight` (sum
    to 1), PERCENT in portfolio files and `Holding.RawWeight`.
- Dates: every `marketdata.Point.Date` is normalized to 00:00 UTC. Metrics
  match series by exact `time.Time` equality; keep the invariant.
- `marketdata.Align` requires `start` at or after every series' first
  quote, otherwise it forward-fills zeros. `portfolio.Simulate` computes
  that window for you; direct callers must too.
- SIM convention: a bare id (`VOO`) = real quotes only; the `SIM` suffix
  (`VOOSIM`) also splices simulated/proxy history in front.
  `Client.FetchExtended` handles it; plain `Fetch` does not. Three ways to
  ask for it wholesale, all setting `Spec.Sim` and resolved by
  `portfolio.SimFetchID` at fetch time (the written id never changes):
  `#meta sim:on` in a file, `-simulate`/`-b` on the command line (every file
  and every `-assets` id of the run), `sim=on` in the `/view` grammar.
  They only turn it ON; `-no-simulate` overrides them all.
- Asset TERs are ALREADY reflected in prices: `Simulate` never deducts
  them (informational). Envelope fees (`extra-fees`) are NOT in prices and
  are deducted daily.
- Closes are ADJUSTED (total-return) by default; `Series.Dividends` +
  adjusted closes double-counts income. Valuation consumers use
  `FetchOptions.Raw` (unadjusted closes + dividends as cash); Raw + SIM
  suffix is an error.
- BUT the FT/Morningstar NAV of a DISTRIBUTING share class is a PRICE
  return: the income it pays out is missing, silently, from every
  statistic (~3.1 %/yr for the long-bond class DTLE, measured). The
  report warns (`LooksDistributing`), nothing corrects it. Never splice
  such a series onto a total-return reconstruction; give the
  reconstruction its own `source: "index"` id instead (`DTLETR`).
- A SECOND WAY a fee escapes the price: a share class whose management
  charge is not levied inside its NAV. AQR's `RAEF` class (LU1662501532)
  lists 1.00 %/yr of management yet only 0.23 % ever reaches the NAV (FT
  ongoing charge, audited Swiss TNER), so it appears to beat its own
  siblings by exactly their fee difference. Pin `fees` to the ONGOING
  charge (what is in the price, per the convention above) and say in
  `notes` what else the holder pays; compare share classes on published
  ongoing charges, never on NAV performance.
- With external flows, `SimResult.Values` follows the money while
  `SimResult.Index` is the time-weighted series: compute statistics and
  comparisons on `Index`, money outcomes (IRR) on `Values` + flows.
- `pkg/scenario` and `pkg/decumul` work in REAL terms (inflation removed)
  and periodic returns; deflate nominal series first (`scenario.Deflate`).
- Annualization: 252 trading days, zero risk-free rate, CAGR over
  365.25-day years. Comparisons with PortfolioVisualizer et al. differ for
  documented reasons (see `pkg/metrics/doc.go`).
- Rate symbols (`^IRX`, `^FVX`, `^TNX`, `^TYX`), the policy/money-market
  family (`^ESTR`, `^EONIA`, `^EURIBOR3M`, `^ECB-DFR`, `^ECB-MRO`, `^SOFR`, `^FEDFUNDS`,
  `^FED-TARGET`, registry in `pkg/marketdata/rates.go`) and `^VIX` are
  annualized percent LEVELS, not prices; `^HICP-<geo>` and `^CPI-US` are index
  levels; all chart fine but never belong in a return computation directly.
  The policy family can be ZERO or NEGATIVE, so `-assets` (which builds a
  portfolio and computes returns) is meaningless on it: read it with
  `pofo -rates <symbols>` instead.
- A surprising CAGR/vol is usually a RESOLUTION bug, not a math bug: read the
  "resolved X -> name" log line first (a fuzzy search may have matched and
  cached an unrelated fund; delete that cache entry).
- `-gen-simdata <ID>` writes `simdata/<CanonicalID(ID)>.csv`: make the id
  canonical BEFORE generating (an alias collision silently overwrites another
  asset's file) and check `git diff --stat pkg/datasets/simdata` after.

## House rules

- Stdlib only. Do not add a third-party dependency.
- English for all code, godoc and docs. Never write an em-dash.
- Every package keeps a `doc.go` (conventions included) and runnable
  `example_test.go` examples; extend them with any new API.
- `make check` must pass; new logic comes with tests (the bar is high:
  most packages are at 75-97 % coverage).
- Calculation changes must keep `make golden` green; if a golden moves,
  justify it against the external reference, never retune the tolerance
  casually.
- Commit and push directly to `master` once `make check` passes.

## Common tasks

- Add a catalog asset: edit `pkg/datasets/assetmeta/assets.json` (see its
  README for the schema and vetting rules), then `make test` revalidates.
- Add a ticker alias: `pkg/marketdata/aliases.go`.
- New simulated history: add a recipe in `pkg/simgen/recipes.go`, validate
  with `./pofo -gen-simdata -dry <ID>`, generate with `make simdata`.
- AQR Managed Futures share classes (`LU1103...`, `LU1662...`, `LU2622...`):
  read `docs/aqr-mf.txt` first. Every EUR class holds the same portfolio, so
  NAV differences are fee differences; `RAEF`'s management fee is WAIVED at the
  manager's discretion, which is why it looks like it beats its siblings and
  why its pin is the least durable in the catalog.
- Managed-futures / trend reconstructions: read
  `docs/trend-reconstruction-design.md` first. RELIABILITY BOUNDS LENGTH is the
  standing decision: every file stops where its evidence stops, the donor chains
  at their deepest real NAV (1996-03) and the overlays at their reference's
  first day (2000-01), and no engine tail is shipped in front of either. Real
  NAVs of the closest programmes come first (`DonorChain`; a weekly-dealing
  donor is projected onto the engine's daily calendar, and every donor segment
  is lifted to the target's published fee load by `feeAligned`, never to close a
  measured return gap); the engine supplies only the daily texture, and a
  bundled reference the monthly path AND the level (`AnchorTrend`, no pin
  anywhere any more): the diversified funds read the NET all-styles composite
  (`TREND-NET-USD`, monthly), the overlays (RSST, RSBT, Winton) the NET
  pure-trend one (`TREND-PURE-NET-USD`, daily, `cmd/gen-sgtrend-refdata`), both
  funded total returns. `TREND-TSMOM-USD` is the gross academic factor, kept as
  a shape yardstick and anchoring nothing. The two REPLICATION funds also take a
  published index as their nearest DONOR rather than another manager's fund:
  DBMF and its UCITS classes the daily all-styles composite
  (`TREND-ALLSTYLES-NET-USD`, same generator), Simplify CTA the daily pure-trend
  one. Since 2026-08 the DBi family reads that composite through the TEN futures
  contracts the fund actually holds: the shipped donor `TREND-ALLSTYLES-DBI-USD`
  (`pkg/simgen/dbireplica.go`, `cmd/gen-dbi-refdata`) is half the composite and
  half a rolling 60-day regression of it on those contracts with the intercept
  discarded, which lifts the monthly agreement with the fund from 0.85 to 0.89
  and cuts the split-half swing of the level gap from 5.2 points to 1.2. The
  SAME treatment was measured for Simplify CTA in 2026-08 and REFUSED: that
  fund's own sixteen quarterly schedules show a fifty-market commodity-and-rates
  book (no equity, no currency) rather than a ten-contract one, so there is no
  instrument-set restriction to harvest, and the design doc holds the numbers.
  Touching the CTA or overlay donor era or the texture breaks two FIRE-book
  plates (their tests recompute from `pkg/datasets` and say so), because the
  weekly donor is projected onto that texture; those plates read `CTA` and
  `SP500`, so a DBi-only change leaves them alone.
- Rolled commodity / crude oil work: read `docs/wti-rolled-reference-design.md`
  first. `WTI-USD` and `WTI-DAILY` are SPOT and are not investable: the roll
  yield was +9.5 points a year over 1986-2000 and -12.8 over 2005-2016, so a
  position priced off spot is wrong by a double-digit rate whose SIGN FLIPS by
  era. `WTI-ER-USD` (`cmd/gen-wti-refdata`, `make wti-refdata`) is the rolled
  EXCESS return (no collateral; fund it with `TBILL-3M` for a total return),
  rebuilt from EIA's first/second nearby NYMEX settlement prices with the S&P
  GSCI roll schedule and validated per calendar year against the published S&P
  GSCI Crude Oil total return. It ENDS 2024-04-05 because EIA discontinued those
  series there, and no engine tail is shipped in front of that bound.
- Catastrophe bond / insurance-linked (ILS) work: read
  `docs/catbond-sleeve-design.md` first. The reference is `ILS-NET-USD`
  (`cmd/gen-catbond-refdata`, monthly from 2006-01, already net of the
  constituent funds' fees), served as `ILSFUND` / `ILSFUNDE` (EUR-hedged, via
  the shared `hedgeToEUR`) and spliced behind three retail UCITS classes by
  `pkg/simgen/catbond.go`, each rescaled to its OWN monthly volatility with
  `monthlyVolMatch` (a monthly donor and a weekly fund share no observation
  dates, so `volMatch` would skip the donor in silence). Nothing reaches before
  2006-01. THE CADENCE TRAP: these lines quote weekly or semi-monthly, so every
  per-observation statistic on them is wrong by ~sqrt(5); read the monthly
  columns.
- Weight search (bounded/constrained optimize, `train:`, `-sweep`): read
  `docs/weight-search-design.md` first; it also records what was deliberately
  left for later (the Pareto `improve` mode, the frontier chart) and the traps
  (zero risk-free Sharpe, `Spec.Train` inert inside `pkg/optimize`).
- New statistic: `pkg/metrics` + tests + a golden anchor if externally
  checkable; expose it in `report.StatRow` via `pkg/compare/page.go`
  (`buildStatRows`) if the CLI should show it.
- Report per-portfolio blocks (composition pies, coverage bars, risk budget,
  realized contribution charts): assembled in `pkg/compare` (`breakdownPies`,
  `coverageBars` in `composition.go`; `riskBudgetRows` in `riskbudget.go`, over
  `metrics.Attribute`; `contributionCharts` in `contrib.go`) from
  `pkg/suggest` composition splits, `SimResult.(Monthly)Contributions` and
  `permanent.Regime.Quadrant`; rendering primitives live in `pkg/chart`
  (`DivergingStack`, `BarMatrix`, `Pie`), the template and its instant-tooltip
  layer in `pkg/report/html.go`.
- New CLI mode: a `run*` function in its own `cmd/pofo/<mode>.go` file, but push any
  reusable logic down into a `pkg/` package first (see `FetchExtended`
  and `portfolio.Build`, which were extracted exactly that way).
- Web app (`-serve`) work: read `docs/webapp-design.md` first (route map, the
  `/view` URL grammar and its guardrails, the design decisions, the M2-M4
  ladder). The `/view` grammar is authoritatively `view.go`'s godoc; keep the
  doc and the godoc in sync when either changes.
- FIRE/decumulation work: read `docs/decumulation-fire-design.md` first;
  the follow-up backlog is `docs/decumulation-fire-program-2026-07.md`. Anything
  touching mortality, estates, couples or annuities goes through
  `docs/stochastic-lifetime-kernel-design.md` as well: `Plan.Lifetime` draws the
  lifespan INSIDE each path, and the standing rule is that the household never
  sees its own drawn death (spending rules plan over `PlanHorizon`, never over
  the lifespan the path was dealt).
- FIRE book work: French is the SOURCE OF TRUTH and every edit lands there
  first. After a French edit, `make book-drift` lists the translations it made
  stale; that report is the English worklist. TRANSLATING an article: follow
  `docs/fire-book-en-translation-brief.md` step by step (it names the FR -> EN
  slug map, the `<!-- edition: fr-only -->` marker that flags never-translated
  French articles, the stamp, the manifest entry, the ledger) with
  `docs/fire-book-en-glossary.md` open. Read
  `docs/fire-book-en-edition-design.md` before touching anything English, and
  never duplicate a plate: figures stay single-source French, and the English
  edition translates the rendered SVG through `figureDict`
  (`scripts/figure-audit.sh en` after any dictionary change). NEW PLATE: the
  recipe is `go doc ./pkg/firebook` § "Adding a plate", the toolbox is
  `pkg/firebook/figures_kit.go`, and `scripts/figure-shot.sh <slug> [fr|en]`
  renders one to PNG for the eye.
- Tactical Permanent Portfolio / Darcet / growth-inflation-regime work: read
  `docs/darcet-permanent-portfolio-design.md` first (complete findings,
  algorithms, data sources, and the empirical-vs-a-priori epistemic ledger);
  the macro drivers live in `pkg/datasets/macropanel`.
- Global Efficient Core (NTSG) / multi-currency government bond baskets: read
  `docs/ntsg-global-efficient-core-design.md` first. The bond overlay is FOUR
  local-currency sleeves (`pkg/simgen/globalbond.go`, 80 % US / 11 % German /
  6 % Japanese / 3 % British), each an excess return over its OWN money-market
  rate with NO FX and NO carry (the fund rolls the currency away with forwards
  whose points are that same differential), weights renormalized over the
  sleeves that quote. The non-US references (`BUND-EUR{,-DAILY}`, `JGB-JPY`,
  `GILT-GBP`, `JPCASH-JPY`, `GBCASH-GBP`) come from `cmd/gen-gbond-refdata`
  (`make gbond-refdata`), which validates every series before writing it; its
  OECD source is the CURRENT `OECD/DSD_STES@DF_FINMARK` dataflow, which
  `gen-euro-refdata` and `gen-macropanel` also read since 2026-08 (the legacy
  `OECD/MEI` dataflow all three used to read froze at 2024-01 while still
  answering HTTP 200; no generator is left on it).
- Eurozone Efficient Core (NTSZ) / euro-native backcasts, incl. the long euro
  govt sleeve (DBXG, `dbxgRecipe`): read
  `docs/ntsz-eurozone-efficient-core-design.md` first. The deep euro reference
  series (`EMU-EUR`, `EUROGOV-EUR{,-DAILY}`, `EUROGOV-LONG-EUR{,-DAILY}` for the
  25+ segment, `DECASH-EUR`, and since 2026-08 the euro cash leg `EURCASH-EUR`,
  which had no generator and froze when its FRED source died) come from DBnomics
  via `cmd/gen-euro-refdata`
  (`make euro-refdata`), which validates every series before writing it (its
  freshness and flat-run checks exist because the MEI freeze and a degraded
  fetch both shipped unnoticed); a short rate is a number PLUS a convention, so
  each cash leg is accrued at the one its publisher states (`EURCASH-EUR` simple
  over a calendar year, `DECASH-EUR` simple on a 360-day year, German 360/360 to
  1990-06 then act/360, re-quoted in 2026-08 for +0.23 %/yr on the pre-1994
  path) and a per-year check gates it; note the equity-leg daily-vol/FX caveat there,
  and that the long sleeve is a real `TreasuryTR` long-bond reconstruction (never a
  levered short bond, which overstates a bond bull). Its LEVEL follows the real
  ECB 25y curve point from 2004-09 (the monthly `EUROGOV-LONG-EUR` is that
  curve's month-ends there, spliced onto the synthesized deep tail); the affine
  25y-on-10y map carries the pre-2004 years only.
