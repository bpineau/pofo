# Domain conventions and maintenance workflows

Read the relevant subsection before modifying its feature. General contribution rules are in `AGENTS.md`, public units in `go doc .`, and runnable consumer recipes in `docs/usage/library/`.

## Conventions and traps (do not guess, check here)

- UNITS, the number one trap. The per-package table lives in the root
  `doc.go` (authoritative) and in `docs/usage/library/README.md` ("Units");
  `TestUnitsTable` keeps the two in step. Fees and rates mix two conventions:
  - PERCENT per year: `portfolio.Holding.Fees`, `Portfolio.EnvelopeFees`,
    `Portfolio.BorrowSpread`, `marketdata.Client.Fees` (0.85 = 0.85 %/yr).
  - FRACTION per year: everything in `pkg/simgen` (fees, vol targets:
    0.0085 = 0.85 %/yr), all of `pkg/metrics` outputs except `Stats.Ulcer`
    (percent points), returns everywhere (0.04 = +4 %).
  - Weights: FRACTIONS in `portfolio.Asset.Weight`/`Holding.Weight` (sum
    to 1), PERCENT in portfolio files and `Holding.RawWeight`.
- Dates: every `marketdata.Point.Date` is normalized to 00:00 UTC. Metrics
  match series by exact `time.Time` equality; keep the invariant.
- CURRENCY of a bundled reference series: the file states none, so
  `Bundled` takes the one its identifier's last segment names (`EM-USD` is
  USD, `EURCASH-EUR` EUR) and `Client.Load` converts it; an identifier that
  names none (`TREASURY-LONG-DAILY`, `TBILL-3M`) is refused with
  `ErrUnknownCurrency` when a `Currency` is asked, never passed through
  unconverted. A FETCHED series whose source states no currency still passes
  `FetchExtended` unconverted, with a warning.
- `marketdata.Align` requires `start` at or after every series' first
  quote, otherwise it forward-fills zeros. `portfolio.Simulate` computes
  that window for you; direct callers must too.
- SIM convention: a bare id (`VOO`) = real quotes only; the `SIM` suffix
  (`VOOSIM`) also splices simulated/proxy history in front.
  `Client.FetchExtended` handles it; plain `Fetch` does not. `Bundled` and
  `Client.Load` obey it too: a quoted fund's simdata file answers only to its
  SIM id (`Bundled("IWDASIM")`), never to the bare one, and a catalog
  `source: "index"` asset (`SP500`), which has no quotes, is long under both
  (`TestBundledSimConvention`). Three ways to
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
- Annualization follows the data's CADENCE: `metrics.PeriodsPerYear(dates)`
  (the prevailing spacing's observations per year, snapped to 252 on any
  trading-day calendar, 52 weekly, 12 monthly, 365 when weekends quote), so a
  daily figure is the classic 252-day one and a monthly refdata file no longer
  reads sqrt(21) too volatile. Functions over bare returns (`Volatility`,
  `Sharpe`, `Sortino`, `CWARP`, `optimize.Solve`...) take the cadence as an
  argument: `metrics.TradingDaysPerYear` for daily returns, never a literal
  252. A line that CHANGES cadence (weekly then daily) is read at its
  prevailing one and misreads the rest: resample it first. Zero risk-free
  rate, CAGR over 365.25-day years. Comparisons with PortfolioVisualizer et
  al. differ for documented reasons (see `pkg/metrics/doc.go`).
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

## Common tasks

- Add a catalog asset: edit `pkg/datasets/assetmeta/assets.json` (see its
  README for the schema and vetting rules), then `make test` revalidates.
- Add a ticker alias: `pkg/marketdata/aliases.go`.
- New simulated history: add a recipe in `pkg/simgen/recipes.go`, validate
  with `./pofo -gen-simdata -dry <ID>`, generate with `make simdata`.
- AQR Managed Futures share classes (`LU1103...`, `LU1662...`, `LU2622...`):
  read `docs/specs/aqr-mf.txt` first. Every EUR class holds the same portfolio, so
  NAV differences are fee differences; `RAEF`'s management fee is WAIVED at the
  manager's discretion, which is why it looks like it beats its siblings and
  why its pin is the least durable in the catalog.
- Managed-futures / trend reconstructions: read
  `docs/specs/trend-reconstruction-design.md` first. RELIABILITY BOUNDS LENGTH is the
  standing decision: every file stops where its evidence stops, the donor chains
  at their deepest real NAV (1996-03) and the overlays at their reference's
  first day (2000-01), and no engine tail is shipped in front of either. Real
  NAVs of the closest programmes come first (`DonorChain`; a weekly-dealing
  donor is projected onto the engine's daily calendar, and every donor segment
  is lifted to the target's published fee load by `feeAligned`, never to close a
  measured return gap); the engine supplies only the daily texture, and a
  bundled reference the monthly path AND the level (`anchorTrend`, no pin
  anywhere any more): the diversified funds read the NET all-styles composite
  (`TREND-NET-USD`, monthly), the overlays (RSST, RSBT, Winton) the NET
  pure-trend one (`TREND-PURE-NET-USD`, daily, `cmd/gen-sgtrend-refdata`), both
  funded total returns. `TREND-TSMOM-USD` is the gross academic factor, kept as
  a shape yardstick and anchoring nothing. The two REPLICATION funds also take a
  published index as their nearest DONOR rather than another manager's fund:
  DBMF and its UCITS classes the daily all-styles composite
  (`TREND-ALLSTYLES-NET-USD`, same generator), Simplify CTA the daily pure-trend
  one. Campbell (`IE00BKYBHJ61`) takes that same all-styles composite as its
  SOLE donor for the opposite reason (nothing public tracks it above 0.53, its
  0.50 correlation to the DBi book being what it is held for) and stops at the
  composite's own 2000-01 floor, taking no deeper donor so the deep Man AHL leg
  behind the two other trend legs is not shared a third time. Since 2026-08 the DBi family reads that composite through the TEN futures
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
- Long / zero-coupon Treasury work (`ZROZ`, STRIPS, `EDV`-shaped funds): read
  `docs/specs/long-treasury-zero-coupon-design.md` first. TWO DATA TRAPS live there
  besides the engine: H.15's 20-year point changed DEFINITION on 1973-01-04
  (6.04 to 6.78 overnight with the 10-year point still), which the bundled
  `TREASURY-LONG-YIELD` declares in a `# junctions:` header so no engine prices
  a return across it; and `VFITX`'s adjusted close carries an unhandled
  year-end capital gain on 1993-12-31 (-3.16 %, permanent), which `simgen`'s
  `tracked` map refuses against `TREASURY-INT-DAILY`. `VFINX` carries the same
  signature on ten sessions of 1980-1986 (about 30 % of level, ~4 pts/yr
  cold on every recipe whose equity leg it carried then), refused since
  2026-09-26 against `^GSPC`, the index it tracks (nine by tolerance, one
  on the record, all pinned by `TestGoldenVanguard500Published` to the fund's
  published 1985-1994 returns); `VUSTX` (1992-12) has no
  separating band and stays measured, reported, NOT repaired. A STRIPS fund is NOT a
  geared coupon fund: the gearing that matches a 27-year strip to a long coupon
  fund is 1.66 at 3 % yields and 3.31 at 12 %, so a constant fitted on the
  low-yield era halves the risk of the high-yield one (the incumbent recipe read
  14.7 %/yr of volatility and a 46 % worst drawdown over 1953-2009 where the
  strip's own arithmetic gives 25.0 % and 80 %). A zero is priced off the yield
  itself, by `TreasuryZeroTR` (`pkg/simgen/treasury.go`, the semiannual sibling
  of `TreasuryTR`, sharing its loop) over the bundled `TREASURY-LONG-YIELD`
  (`cmd/gen-tyield-refdata`, `make tyield-refdata`). Its maturity is the
  maturity of the PAPER, a-priori, never tuned to a fund's realized volatility,
  and the 0.2 to 0.4 pt/yr by which it runs cold is the par-yield and no-roll-down
  approximation, documented and deliberately NOT corrected. The coupon-to-coupon
  gearing of the 20+ year ETFs (`longTreasuryGearing`) drifts only 7 % across
  the same rate range and stands.
- Rolled commodity / crude oil work: read `docs/specs/wti-rolled-reference-design.md`
  first. `WTI-USD` and `WTI-DAILY` are SPOT and are not investable: the roll
  yield was +9.5 points a year over 1986-2000 and -12.8 over 2005-2016, so a
  position priced off spot is wrong by a double-digit rate whose SIGN FLIPS by
  era. `WTI-ER-USD` (`cmd/gen-wti-refdata`, `make wti-refdata`) is the rolled
  EXCESS return (no collateral; fund it with `TBILL-3M` for a total return),
  rebuilt from EIA's first/second nearby NYMEX settlement prices with the S&P
  GSCI roll schedule and validated per calendar year against the published S&P
  GSCI Crude Oil total return. It ENDS 2024-04-05 because EIA discontinued those
  series there, and no engine tail is shipped in front of that bound.
- World-equity blends (`VT`, the FTSE All-World UCITS class, `RSSB`'s equity
  sleeve): read `docs/specs/world-equity-capweight-design.md` first. Their weights are
  CAP-WEIGHTED, not fixed: `CapWeighted` holds the published 2009-10-31 FTSE
  All-World split (40.4 US / 46.1 developed-ex-US / 13.5 emerging, `worldLegs`)
  and never rebalances, so the weights drift with the legs' own returns. A fixed
  blend at today's split is a look-ahead bias worth +1.8 pts/yr over 1988-2008;
  the drift ignores net issuance, measured at 5.2 points of US weight over the
  fourteen years after the anchor. `Composite` stays for the funds that really
  do rebalance to fixed weights.
- Total-US-market / size-completion work: read
  `docs/specs/us-total-market-reference-design.md` first. `VTI` is backcast on the
  target's OWN share class (`VTSMX`, 1992-04) extended by the CRSP total-market
  factor (`USMKT-USD`, `make usmkt-refdata`), and the S&P 500 is NOT a stand-in
  for the whole market: the mid/small completion premium is worth about a point
  a year and CHANGES SIGN by era (-1.05 pts/yr over the two funds' own
  1992-2001 overlap, +0.61 over 1962-2001). The factor is gross of costs, so it
  owes a measured 0.30 %/yr charged once, in `longBackFee`. `VFINX` stays the
  donor of the recipes that MEAN the S&P 500 (`SP500`, `GDE`, `NTSX`, `RSST`,
  the world blends' US leg).
- Employee-savings fund (FCPE: `ERESMONDEM`, the world-equity fund, and
  `ERES_DATADOG`, the single-stock DDOG fund kept apart from the listed `DDOG`)
  work: read `docs/specs/eres-fcpe-design.md` first. Such a fund has no ISIN and no
  listing; its official NAV comes live from the airfund.io API (`source:
  "airfund"`, share code in `symbol`, the Eres SITE's widget id in `xid`, the
  same for every fund of the site) with `refdata/<ID>-NAV.csv` as the offline
  fallback (`make eres-refdata`). `ERES_DATADOG` was WEEKLY until 2026-07-13, daily since
  (a line that changes cadence: its statistics annualize at the prevailing
  weekly count and misread the daily tail, so read the monthly columns) and is valued
  at the NASDAQ OPENING price since 2022-04-11 and at the CLOSE before (the FY2025 annual report's valuation rules; measured two ways: 2.0 % rmse vs 3.6 % for the
  close, and 09:30 New York the best-fitting instant on the 5-minute history),
  so its record carries `nowcast_anchor: "open"` (catalog field, default
  `close`): the nowcast anchors on the proxy's OPENING print of the last NAV's
  day, an anchor worth 3.8 % rmse against 2.8 % on the one-span-ahead estimate,
  and falls back on the close whenever that open cannot be read (the opens
  travel as an open-to-close FACTOR series cached under `SYMBOL~open`, never as
  a field on `Point`). THE CLOCK TRAP: the NAV of day D is struck on the two
  ETFs' official NAVs, i.e. after New York closes, so a US-listed tracker
  (URTH) follows its daily moves at 0.87 correlation and a Xetra line at 0.62;
  that is why `nowcast_proxy` is URTH and why the recipe's audit reads
  `path=warn` on daily, 0.98 on monthly. Estimated points (`EstimatedFrom`,
  `IntradaySeries.Estimate`, `Quote.Source == "nowcast"`) are never cached and
  never shipped: every storing consumer reads through `WithoutEstimates`.
  The wrapper charge in the recipe (0.41 %/yr) is the a-priori figure from the
  FY2025 report, never the measured residual. The Xtrackers donor lines carry
  prints that are not the fund's, so every donor is held to the MSCI World EUR
  path it tracks (`simgen.trackIndex`, tolerance stated per pair, the reference
  allowed to lead or lag one session so a Xetra close is not convicted of a US
  move) and `DBXW.DE` starts 2009-01-02, its earlier quotes being the fund's
  USD line (excess return correlated 0.998 with EUR/USD).
- Catastrophe bond / insurance-linked (ILS) work: read
  `docs/specs/catbond-sleeve-design.md` first. The reference is `ILS-NET-USD`
  (`cmd/gen-catbond-refdata`, monthly from 2006-01, already net of the
  constituent funds' fees), served as `ILSFUND` / `ILSFUNDE` (EUR-hedged, via
  the shared `hedgeToEUR`) and spliced behind three retail UCITS classes by
  `pkg/simgen/catbond.go`, each rescaled to its OWN monthly volatility with
  `monthlyVolMatch` (a monthly donor and a weekly fund share no observation
  dates, so `volMatch` would skip the donor in silence). Nothing reaches before
  2006-01. These lines quote weekly or semi-monthly; their statistics
  annualize at that measured cadence (52 or 24 a year), no longer at 252, but
  a line that switched cadence along the way is read at its prevailing one, so
  the monthly columns stay the safe read.
- Black-Litterman (`optimize:black-litterman`, `view:`, `prior-return:`): read
  `docs/specs/black-litterman-design.md` first. The prior is the FILE's weights, not
  a market-capitalization portfolio, and with no view the objective returns
  those weights EXACTLY (the identity every test hangs on). The golden lives
  in `pkg/datasets/golden/blacklitterman_test.go` and pins the model on the
  printed tables of He and Litterman (1999) and Idzorek (2005); a confidence
  of 0.5 is those papers' own `Ω = diag(P τ Σ Pᵀ)`.
- Weight search (bounded/constrained optimize, `train:`, `-sweep`): read
  `docs/specs/weight-search-design.md` first; it also records what was deliberately
  left for later (the Pareto `improve` mode, the frontier chart) and the traps
  (explicit `Spec.RiskFree` for Sharpe/Sortino; apply and clear `Spec.Train` before solving).
- New statistic: `pkg/metrics` + tests + a golden anchor if externally
  checkable; expose it in `report.StatRow` via `pkg/compare/page.go`
  (`buildStatRows`) if the CLI should show it.
- Ad hoc analysis across series (blends, regressions, conditional stats on
  the worst months, fee drags, named episodes): in Go, over
  `marketdata.NewPanel` and `pkg/metrics`, never in a Python notebook with
  its own return math. `docs/usage/library/panels.md` is the recipe;
  throwaway programs go under the gitignored `/scratch/`.
- Report per-portfolio blocks (composition pies, coverage bars, risk budget,
  realized contribution charts): assembled in `pkg/compare` (`breakdownPies`,
  `coverageBars` in `composition.go`; `riskBudgetRows` in `riskbudget.go`, over
  the study's `Attribution`; `contributionCharts` in `contrib.go`) from
  the column study's `Composition` (pkg/suggest's splits; the coverage bars
  still call suggest, under the comparison's framework),
  `SimResult.(Monthly)Contributions` and
  the descriptive macro-regime reading of `regime.go` (growth x inflation
  breadth over `datasets.MacroPanel`); rendering primitives live in `pkg/chart`
  (`DivergingStack`, `BarMatrix`, `Pie`), the template and its instant-tooltip
  layer in `pkg/report/html.go`.
- New CLI mode: a `run*` function in its own `cmd/pofo/<mode>.go` file, but push any
  reusable logic down into a `pkg/` package first (see `FetchExtended`
  and `portfolio.Build`, which were extracted exactly that way).
- Web app (`-serve`) work: the route map is `cmd/pofo/serve.go`'s godoc and
  the `/view` URL grammar is authoritatively `view.go`'s godoc; there is no
  separate design doc to keep in sync. A `p=` identifier outside the
  bundled catalog is accepted only if it passes `marketdata.PlausibleID` and
  the hourly budgets in `foreign.go`, and is then resolved EXACTLY
  (`FetchOptions.ExactOnly`) so a visitor's typo fails instead of caching an
  unrelated fund.
- FIRE/decumulation work: `go doc ./pkg/decumul` is the design (its "Realism"
  section records the optimism that was measured out of the defaults and must
  not creep back). The lifecycle view (§05) runs the exact kernel: its band
  is `Ensemble.LifeBands` (gone from the top, broke on the baseline, the
  funded split by wealth against the starting capital), drawn with
  `chart.StackedAreaWith`. One item is known open and deliberately so: a
  CAPE-conditioned spending rule (WR = a + b/CAPE), refused until a
  scenario source simulates a valuation path. The per-envelope tax model was measured and refused
  (`docs/specs/fire-envelopes-tax-model-design.md`). Anything
  touching mortality, estates, couples or annuities goes through
  `docs/specs/stochastic-lifetime-kernel-design.md` as well: `Plan.Lifetime` draws the
  lifespan INSIDE each path, and the standing rule is that the household never
  sees its own drawn death (spending rules plan over `PlanHorizon`, never over
  the lifespan the path was dealt).
- FIRE book work: French is the SOURCE OF TRUTH and every edit lands there
  first. After a French edit, `make book-drift` lists the translations it made
  stale; that report is the English worklist. TRANSLATING an article: follow
  `docs/specs/fire-book-en-translation-brief.md` step by step (it names the FR -> EN
  slug map, the `<!-- edition: fr-only -->` marker that flags never-translated
  French articles, the stamp, the manifest entry, the ledger) with
  `docs/specs/fire-book-en-glossary.md` open. Read
  `docs/specs/fire-book-en-edition-design.md` before touching anything English, and
  never duplicate a plate: figures stay single-source French, and the English
  edition translates the rendered SVG through `figureDict`
  (`scripts/figure-audit.sh en` after any dictionary change). NEW PLATE: the
  recipe is `go doc ./pkg/firebook` § "Adding a plate", the toolbox is
  `pkg/firebook/figures_kit.go`, and `scripts/figure-shot.sh <slug> [fr|en]`
  renders one to PNG for the eye.
- Global Efficient Core (NTSG) / multi-currency government bond baskets: read
  `docs/specs/ntsg-global-efficient-core-design.md` first. The bond overlay is FOUR
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
  answering HTTP 200; no generator is left on it), read at the OECD's own SDMX
  API since 2026-09 rather than through the DBnomics mirror, which froze at
  2026-05.
- Eurozone Efficient Core (NTSZ) / euro-native backcasts, incl. the long euro
  govt sleeve (DBXG, `dbxgRecipe`): read
  `docs/specs/ntsz-eurozone-efficient-core-design.md` first. The deep euro reference
  series (`EMU-EUR`, `EUROGOV-EUR{,-DAILY}`, `EUROGOV-LONG-EUR{,-DAILY}` for the
  25+ segment, `DECASH-EUR`, and since 2026-08 the euro cash leg `EURCASH-EUR`,
  which had no generator and froze when its FRED source died) come from the
  OECD's own SDMX API (the OECD inputs) and DBnomics (the ECB ones) via
  `cmd/gen-euro-refdata`
  (`make euro-refdata`), which validates every series before writing it (its
  freshness and flat-run checks exist because the MEI freeze and a degraded
  fetch both shipped unnoticed); a short rate is a number PLUS a convention, so
  each cash leg is accrued at the one its publisher states, and BOTH were
  re-quoted in 2026-08 after being read wrong (`EURCASH-EUR` simple act/360, per
  EMMI's published Euribor convention, +0.033 %/yr on the euro-era path;
  `DECASH-EUR` simple on a 360-day year, German 360/360 to 1990-06 then act/360,
  +0.23 %/yr on the pre-1994 path); a per-year check restating the publisher's
  own convention gates each of them; note the equity-leg daily-vol/FX caveat there,
  and that the long sleeve is a real `TreasuryTR` long-bond reconstruction (never a
  levered short bond, which overstates a bond bull). Its LEVEL follows the real
  ECB 25y curve point from 2004-09 (the monthly `EUROGOV-LONG-EUR` is that
  curve's month-ends there, spliced onto the synthesized deep tail); the affine
  25y-on-10y map carries the pre-2004 years only. SINCE 2026-09-19 the same
  doctrine covers `EUROGOV-EUR` (real ECB 10y curve month-ends from 2004-09) and
  `BUND-EUR` (real Bundesbank curve month-ends from 1997-08): an OECD monthly
  yield is the month's AVERAGE, which smooths a reconstruction by a fifth and
  puts its months half a step out of phase (monthly correlation with the fund
  behind `EUROGOV-EUR` went 0.59 -> 0.88, and with the one behind `BUND-EUR`
  0.60 -> 0.95). Every OECD-driven monthly file now carries a month-END label,
  the cash accruals included, which emptied the `gaps_test.go` allow-list;
  `GILT-GBP`, `EMU-EUR`, `WTI-USD` and `TBILL-3M` are measured, named and left
  (see the sweep section of `docs/specs/ntsz-eurozone-efficient-core-design.md`).
