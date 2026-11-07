# Repository maintenance

Read the sections relevant to the changed subsystem. The entry rules and completion gates are in `AGENTS.md`. These are recurring procedures, source incidents and validation practices.

## Commands

```sh
make build     # ./pofo binary, pkg/datasets/ embedded via go:embed
make test      # go test ./...  (unit tests, runnable examples, the godoc guards; NO network)
make lint      # go vet + staticcheck
make examples  # build + vet + staticcheck every examples/code/*.go script, one file at a time (no tests: a compile gate)
make check     # fmt-check + lint + test + examples: run this before any commit
make golden    # computation goldens vs frozen external references
make refresh   # refresh EVERY bundled series from its live source, in order (network)
make simdata   # regenerate pkg/datasets/simdata/ (network) then rebuild
make broadsample # regenerate the JST broad-sample panel (network) then rebuild
make cape      # regenerate the Shiller CAPE series (network) then rebuild
make macropanel # regenerate the OECD monthly macro panel (network) then rebuild
make euro-refdata # regenerate the euro-area reference series (network) then rebuild
make gbond-refdata # regenerate the German/Japanese/British govt bond reference series (network); run make simdata after
make tyield-refdata # regenerate the SIX US Treasury references: TREASURY-LONG-YIELD (the long par yield the STRIPS reconstruction is priced off), the month-end total returns TREASURY-LONG-USD / TREASURY-INT-USD, the daily shapes TREASURY-LONG-DAILY / TREASURY-INT-DAILY, and TBILL-3M (FRED TB3MS as published, the dollar cash leg behind ^IRX; the refresh must reproduce its history point for point) (network); run make simdata after
make dbi-refdata # regenerate the DBi family's nearest donor (the all-styles composite half-projected on the fund's ten futures); run after make sgtrend-refdata, then make simdata
make sp500-refdata # regenerate the month-end SP500-USD reference (network); run make simdata after
make usmkt-refdata # regenerate USMKT-USD, the CRSP total-market factor behind the VTI backcast (network); run after make sp500-refdata, then make simdata
make msci-refdata # extend MSCIWORLD-USD/DEVEXUS-USD/EM-USD past their manual Curvo export with a validated ETF proxy tail (network); run make simdata after
make french-refdata # regenerate USSCV-USD (small-cap value, behind DFSVX/ZPRV) and DEVEXUS-DAILY (the DEVEXUS-USD daily shape) from the Ken French Data Library (network); run after make msci-refdata, then make simdata
make wti-refdata # regenerate WTI-ER-USD, the rolled-futures EXCESS return of WTI crude, and refresh WTI-USD, the monthly spot (FRED WTISPLC as published) (network); run after make tyield-refdata
make gold-refdata # regenerate XAUUSD-LBMA, the daily LBMA gold fix since 1968 (network, NOT in make refresh: the LBMA gates plain clients); run make simdata after
make bcom-refdata # regenerate BCOM-ER-USD, the daily Bloomberg Commodity Index (excess return) since 1991, from the FT, checked against BCOMTR less TBILL-3M (network); run after make tyield-refdata, then make simdata
make trend-refdata # regenerate the monthly trend reference (network); run make simdata after
make trendnet-refdata # regenerate the monthly NET managed-futures reference (network); run make simdata after
make sgtrend-refdata # regenerate the daily NET pure-trend reference (network); run make simdata after
make catbond-refdata # regenerate the monthly NET insurance-linked reference (network); run make simdata after
make eres-refdata # regenerate the bundled NAV snapshot of the catalog's FCPE funds (ERESMONDEM-NAV, network); run make simdata after
make snapshots # regenerate pkg/marketdata/data/'s offline fallback snapshots (network)
make simdata-qa # reconstruction quality: every engine vs the real quotes, HTML (network)
make verify-catalog # data doctor over the whole catalog: hygiene, plausibility bands, identity (network)
make book-drift # what the FIRE book's translations owe their French source
make figure-drift # what the FIRE book's frozen figures owe the bundled data
```

**`make check` is the single completion gate** (fmt-check + vet + staticcheck +
the whole test suite + the example scripts compiled and linked one by one).
Green looks like `ok` or `no test files` on every package and nothing else: any
`FAIL`, any staticcheck line, any gofmt diff, any `examples: FILE` line is a
failure. It takes a couple of minutes cold on this tree and about twenty seconds
warm.
`make golden` is the second gate for anything that touches a computation, and it
is NOT part of `make check` (it is slower and pins external references).

**Nothing runs on a push.** The only workflow (`.github/workflows/release-image.yml`)
is tag-gated, so the local gate is all there is: do not push red. Commit and
push directly to `master` once `make check` passes; there are no branches and no
PR flow.

**A `v*` tag is a PRODUCTION DEPLOY, not a bookmark.** It builds the container
image, pushes it to the registry, and the public instance picks `latest` up by
itself. It is also what the sibling repos pin to. Never tag unless publishing is
exactly what was asked for.

`make refresh` runs every generator below it in dependency order (references
first, then the simdata built on them, then the offline snapshots); the
individual targets are for touching one series. Follow it with `make check` and
`make golden`, both of which must stay green (refreshing data must never require
a code change), then `make verify-catalog`, which says what the new quotes look
like. Neither gate compares a rebuilt file with the one it replaced: the goldens
pin a reconstruction to its external reference, inside bands wide enough for a
donor to move a third of a point a year and nine points of level unseen (the DBi
donor did, on 2026-10-24, when the FX fill of 2026-10-23 changed its currency
legs). So when a code change touched an input since the last refresh (a fetch
path, a cleaner, an alignment), run `go run examples/code/oldnew.go -rev <last
refresh> ID` on the reconstructions that read it, and name every move and its
cause in the refresh commit. The generator modes fetch fresh (`-gen-simdata` and `-verify-simdata`
re-download any quote older than a day, since what they write ships in the
binary), but the doctor does not: `-cache-age` defaults to 30 days, so
`make verify-catalog` on a cache left warm for weeks reports the cache's age as
stale market data; pass `-cache-age 6h` after a pause.

EVERY bundled series that tracks a live source has a generator in that chain,
with one declared exception (`XAUUSD-LBMA`, below): a file refreshed by hand
freezes unseen (`TBILL-3M`, `WTI-USD`, `XAUUSD-LBMA`, `USSCV-USD` and
`DEVEXUS-DAILY` all did, at 2026-05, until 2026-09-27). A
MIRROR can freeze too, behind a generator that runs fine: DBnomics stopped
indexing the Bundesbank curve at 2026-07-03 (`BUND-DAILY` now reads the
Bundesbank's own SDMX service, with a freshness check), and its OECD mirror
was last indexed 2026-06-16, which froze every OECD-fed file (`EMU-EUR`,
`GILT-GBP`, `GBCASH-GBP`, `JPCASH-JPY`, the macro panel, the OECD tails) at
2026-05 until the generators moved to the OECD's own SDMX API in 2026-09
(`refgen.OECD`). That API admits 60 downloads an HOUR and answers the next
with HTTP 429: never loop over series against it, pass a dataflow's keys to
ONE `refgen.OECD` call (it folds them into one SDMX key), and do not burn the
budget on exploratory probes. A DATASET can freeze as well, when its
publisher rebases: Eurostat moved the HICP to a new dataset (`prc_hicp_minr`,
ECOICOP 2) in 2026 and left the old `prc_hicp_midx` answering at 2025-12, so
`^HICP-FR` deflated all of 2026 at zero inflation until 2026-09-27; the
offline snapshots (`pkg/marketdata/data/`) are now held to their own
`# generated:` stamp by `TestSnapshotsCurrentWhenGenerated` and by
`cmd/gen-snapshots`' `stale` check (`docs/specs/inflation-deflators.md`).
A SOURCE can lock plain clients out, too: since 2026-10 prices.lbma.org.uk
answers curl and Go alike with a Cloudflare challenge (HTTP 403), so
`XAUUSD-LBMA` holds at 2026-09-25 and `gold-refdata` left the `make refresh`
chain, the one generator outside it: the file only backs gold before the live
`XAUUSD`/`GC=F` quotes begin (~2000), where it never changes, so its frozen
tail moves no backcast; it carries an `# ends:` header at its last fix so the
staleness sweep stays quiet, and `xauusdBuild` fails rather than serve the fix
alone when the live quote cannot be read. No key-less mirror of the
daily PM fix since 1968 was found; the godoc of `cmd/gen-gold-refdata` lists
what was tried, and its `-src` takes a browser download as a local path.
The shared plumbing (FRED, OECD and Ken French
readers, the history-reproduction (`SameHistory`, `CompareSteps`), cadence and
flat-run checks, the header writer) is `cmd/internal/refgen`. A
file that STOPS by design (its source was discontinued, or it is trimmed where
real quotes take over) says so in an `# ends: <date> (<reason>)` header, read
into `marketdata.Series.Ends`; `Verify` and `go run examples/code/scanbundle.go`
then never call it stale. `go run examples/code/scanbundle.go -stale 45` is the
sweep that finds a frozen file.

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

- Data questions (a series, a pair, a refreshed file, holes and bad prints):
  the table of "Exploring data" above. `-pair` prints the CAGR gap with its
  standard error, the level gap, the ratio A/B and the first date it moves,
  daily and monthly correlation, tracking error, beta, alpha, dated
  divergences and the calendar years side by side.
- Report/chart changes: `scripts/report-shot.sh [file] [out-prefix]` builds,
  renders the report with every section unfolded and screenshots it full-page
  (needs Chrome and a warm quote cache; `./pofo -warmup` once). Crop a region
  with `sips -c <height> 1500 --cropOffset <y> 0 out.png --out crop.png` and
  read the PNG. `examples/portfolios/dragon-decumulation-household.txt` is the reference
  fixture: 9 holdings across a stacked 90/60, long bonds, two trend funds,
  gold, linkers and small-value equity, so it exercises every report block.
- Chart hover/tooltip data: grep the rendered HTML for
  `<metadata class="hover">` and replay the front-end math on the JSON
  payload in node or Go instead of driving a browser.
- Is a backcast any good? `./pofo -verify-simdata [ID...]` replays each engine
  without its real graft, against the real quotes, and opens an HTML report:
  level and path verdicts, drift panel, donor-chain junctions. This is the
  cheapest way to see what a recipe change did. It also prints one text row
  per recipe on stdout, and `-json` prints the whole `[]simgen.AuditGroup`
  instead of the page (no NaN, no curves): read those, never scrape the HTML.
  `-offline` replays it from the quote cache, which makes a before/after diff
  of a change exact.
- One golden at a time: `go test -v -run TestGoldenGold ./pkg/datasets/golden/`.
  Chart snapshots moved on purpose? Regenerate with
  `UPDATE_SNAPSHOTS=1 go test ./pkg/chart -run TestChartSnapshots` and justify
  the diff in the commit message.
- Catalog edits: `make test` revalidates `assets.json`;
  `./pofo -verify-data -assets <id>` checks a single asset end to end, and
  `make verify-catalog` runs the doctor over all of it (plausibility bands per
  `asset_class`, identity vs the record); run it after any catalog edit or
  `make refresh`.
- Documentation: `go test -run 'TestGodoc|TestDocSnippets|TestUnitsTable' .`
  is the cheap check (all run in `make test`); `go doc ./pkg/<name>` renders
  what pkg.go.dev will show.
