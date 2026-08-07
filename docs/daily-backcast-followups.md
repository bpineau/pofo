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

## Follow-ups, by expected impact

1. **Treasury legs daily (1962→)**. `TREASURY-INT-USD` / `TREASURY-LONG-USD`
   are monthly (FRED GS5/GS20 through `simgen.TreasuryTR`). FRED also
   publishes the DAILY constant-maturity yields: `DGS5` (1962→, continuous)
   and `DGS20`/`DGS30` (1962→/1977→, with gaps 1987-1993 and 2002-2006).
   Run the daily series through `TreasuryTR` (generalize its monthly
   repricing step to the actual day gap) to regenerate the refdata daily;
   for the LONG leg, bridge the gaps with the monthly anchors (anchorShape
   needs a small extension to keep anchors inside shape gaps, today it
   collapses them). Affects NTSG/NTSX before VFITX (1991) and the rates
   leg of every TSMOM frame. FRED was unreachable from the sandbox on
   2026-07-02 (timeouts); retry, or find a mirror.

2. **SG CTA index daily (2000→)**. `docs/SG-CTA-Index-Daily-Returns-since-1999-12-31.csv`
   already sits in this repo (daily index returns since 1999-12-31, not yet
   referenced by any code). Candidate uses: a refdata series to validate or
   anchor the DBMF-family TSMOM replications after 2000, or a splice source
   for a generic managed-futures asset. Before 2000 no free daily managed
   futures series is known (Barclay CTA is monthly).

3. **Emerging markets daily shape (VEIEX leg, monthly before 1994)**.
   `EM-USD` (Curvo MSCI EM net, monthly ~1988→) could be shaped by the Ken
   French "Emerging Markets 5 Factors [Daily]" file the same way as
   dev-ex-US; check the daily file's actual start (the monthly one starts
   1989-07). One more `dailyShape` entry once the refdata is generated.

4. **Daily FX before 2003 (all EUR-converted statistics)**. The euro cross
   is daily from ~2003 (Yahoo) and monthly anchors before (FRED ECU/EUR),
   so every USD asset converted to EUR carries monthly FX steps before
   2003. FRED publishes DAILY `DEXUSEU` (1999→) and `DEXGEUS` (DEM/USD,
   1971-1998, chain at 1.95583 DEM/EUR): a daily extension of the bundled
   eurusd-long proxy would make pre-2003 EUR statistics fully daily.

5. **WTI daily (CL=F leg, monthly before 2000)**. `WTI-USD` is FRED
   `WTISPLC` monthly averages. The EIA publishes the daily WTI spot from
   1986 (API key required, free) and FRED mirrors it as `DCOILWTICO`
   (daily, 1986→). Affects the energy leg of the TSMOM frames.

6. **Pre-1972 / pre-1990 equity shapes (low priority)**. No free daily
   MSCI data before 1972; a composite of national daily indices (^GSPC
   1927→, ^N225 1965→, ^FTSE 1984→, ^GDAXI 1987→) could shape the world
   and dev-ex-US anchors further back if the pre-1972 era ever matters at
   daily granularity.

Also worth considering instead of (or alongside) more data work: make the
daily statistics time-aware (annualize each return by its actual calendar
gap) or have the report flag daily stats when `SimulatedBefore` falls
inside the window with monthly spacing; either would make the remaining
monthly segments honest instead of silently wrong.
