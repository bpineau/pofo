# Daily-granularity backcasts: status and follow-ups

Date: 2026-07-02

Status: partially done (commits 5497122, d77065b); this file tracks what
remains monthly and the identified real-data sources to fix each piece.

## Why this matters

The daily-frequency statistics (volatility, Sharpe, Sortino) assume one
observation per trading day. A monthly backcast point fed to them counts a
whole month's move as one day: the annualized arithmetic mean inflates ~21x
and the volatility ~4.6x on that portion, so a mostly-monthly series shows
an absurdly high Sharpe next to a daily one. A variance ratio far below 1
(monthly/daily annualized variance) combined with `SimulatedBefore` inside
the window is the tell. CAGR and drawdowns are date-aware and unaffected.

## Done

- `simgen.anchorShape` / `shapedSeries`: a coarse total-return series keeps
  setting the levels (every anchor is hit exactly) while a daily series of
  the same market supplies the day-to-day shape, the residual (dividends,
  universe drift) compounding evenly across each month's trading days.
- MSCI World (IWDA/URTH simdata): monthly Curvo net-TR anchors shaped by
  the Yahoo daily MSCI World PRICE index `^990100-USD-STRD` (daily from
  1972-01-03; its levels lag TR by the dividend yield, never use them).
- Developed-ex-US (the `VTMGX` leg behind NTSG, VT, RSSB, Winton):
  `DEVEXUS-USD` monthly anchors shaped by `DEVEXUS-DAILY` refdata (Ken
  French "Developed ex US 3 Factors [Daily]", Mkt-RF+RF cumulated,
  1990-07→), wired through `simgen.dailyShape` in the extending fetcher.
- Gold: `XAUUSD-LBMA` refdata is now the real daily London PM fix
  (https://prices.lbma.org.uk/json/gold_pm.json, 1968-04→). Real daily
  LEVELS, no shape blending needed. `silver.json` exists at the same
  endpoint if a long XAGUSD is ever wanted.

## Not a bug: the EUR variance ratio

After the fixes, a diversified EUR portfolio still shows VR ~0.6 on fully
daily data (NTSGSIM: VR 0.76 in USD since 1995, 0.46 in EUR since 2005).
That is genuine economics, not an artifact: unhedged daily FX noise partly
nets out at the monthly horizon (dollar/equity correlation), plus the
90/60 overlay's mean reversion. Do not try to "fix" it; compare portfolios
on the monthly rows.

## Done in the 2026-07-03 pass

- **Treasury legs daily (1962→)**: `TREASURY-INT-DAILY` / `TREASURY-LONG-DAILY`
  refdata shapes from the FRED daily CMT yields (DGS5 1962→1992, DGS20
  1962→1986; real VFITX/VUSTX take over from there) through `TreasuryTR`,
  wired via `simgen.dailyShape`. `shapedSeries` now keeps anchors past the
  shape's end instead of vetoing the blend. IEF/TLT/ZROZ pre-1991 and the
  NTSX/RSSB/TSMOM rates legs are daily back to 1962.
- **S&P 500 daily shape (1927→)**: `dailyShape[SP500-USD] = ^GSPC` (Yahoo
  daily price index), same anchors+shape blend as MSCI World; the VFINX leg
  is daily before 1976 now.
- **WTI daily (1986→2000)**: `WTI-DAILY` refdata (FRED DCOILWTICO spot)
  shapes the WTI-USD monthly averages; improved every TSMOM validation.
- **Daily FX from 1971**: `eurusd-long.csv` is now daily throughout: FRED
  DEXUSEU daily 1999→, monthly ECU anchors carrying the real daily
  Bundesbank Frankfurt DM/USD fixing 1979-1998 (levels unchanged at every
  month boundary), rescaled daily DM alone 1971-1978. GBPUSD gets the same
  treatment through `longBack[GBPUSD=X] = GBPUSD-DAILY` (FRED DEXUSUK,
  1971→), which moved the DPGT backcast start from 2004 to 1994.
- **FX out of the frames**: `dbmfeBuild`/`dpgtBuild` no longer join the FX
  cross into the strategy frame (Yahoo FX prints Sundays: DBMFE carried ~31
  Sunday rows a year and a ~290-day calendar that under-annualized vol and
  over-accrued the cash leg). The cross is forward-filled onto the
  strategy's own calendar (`fxOnDates`/`convertDaily`).
- **FX bad prints**: Yahoo printed EURUSD=X at 1.4918 on 2008-12-08 between
  1.2717 and 1.2926, a ±15% one-day artefact in every EUR-converted series;
  `dropFXSpikes` (marketdata clean layer, "=X" symbols only) now strips
  isolated self-cancelling spikes ≥8%.

## Follow-ups, by expected impact

1. **SG CTA index daily (2000→)**. `docs/SG-CTA-Index-Daily-Returns-since-1999-12-31.csv`
   already sits in this repo (daily index returns since 1999-12-31, not yet
   referenced by any code). Candidate uses: a refdata series to validate or
   anchor the DBMF-family TSMOM replications after 2000, or a splice source
   for a generic managed-futures asset. Before 2000 no free daily managed
   futures series is known (Barclay CTA is monthly). Not wired in the
   2026-07-03 pass because it is a modeling decision (anchoring the TSMOM
   replicas to a different strategy's index changes their returns), not a
   granularity fix.

2. **Emerging markets daily shape (VEIEX leg, monthly before 1994)**.
   Checked 2026-07-03: no free source. Ken French publishes no EM daily
   factors (monthly only), and Yahoo's `^891800-USD-STRD` (MSCI EM) serves
   no history. Revisit if a daily EM index (gross or price) ever surfaces.

3. **Pre-1972 / pre-1990 equity shapes (low priority)**. No free daily
   MSCI data before 1972; a composite of national daily indices (^N225
   1965→, ^FTSE 1984→, ^GDAXI 1987→) could shape the dev-ex-US anchors
   further back if the pre-1972 era ever matters at daily granularity.
   (The US side is covered: the ^GSPC shape carries SP500-USD daily from
   1927-12.)

Also worth considering instead of (or alongside) more data work: make the
daily statistics time-aware (annualize each return by its actual calendar
gap) or have the report flag daily stats when `SimulatedBefore` falls
inside the window with monthly spacing; either would make the remaining
monthly segments honest instead of silently wrong.

## New backcast candidates (2026-07-03, from the examples/fire-* work)

Bare assets that keep capping the common window of the curated
decumulation portfolios, in the order they would pay:

1. **EUR overnight cash (XEON, LU0290358497; real quotes 2007→).** The
   cash sleeve of every decumulation build. simgen already carries an
   internal `EURCASH-EUR` leg (EONIA/€STR chain), so a recipe would be
   nearly free and would stop pinning portfolios at 2007.
2. **Euro government bonds (IEGA, IE00B4WXJJ64; real quotes 2014→).** The
   natural EUR duration brick; today `fire-simple-no-leverage` has to use
   IEFSIM (US 7-10y) as a stand-in. The Bundesbank daily series already
   fetched for the DM work could shape a Bund-based backcast.
3. **Broad commodities (ICOM, IE00BDFL4P12; real quotes 2017→).** Used by
   modern-all-weather, ntsx-all-weather and stagflation-bunker; a BCOM
   total-return chain would take it decades back.
4. **Direct-trend UCITS (Winton IE00BG382Q20 2018→, AQR LU1103257975
   2014→).** The `fire-trend-sleeve-lab` legs. The SG CTA daily CSV above
   (or the SG Trend index) is the obvious splice source; same modeling
   caveat as follow-up (1).
5. **Euro linkers before 2009 (LU1645380442 sims via IBCI, 2009→).** Caps
   the whole fire-* family at 2009. Hard: no free daily euro-linker series
   before IBCI's inception; French OATi index data would be needed. Low
   priority next to (1)-(3).
