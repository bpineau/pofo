# Asset metadata: the bundled catalog

`assets.json` is the **single source of truth** for the assets pofo
bundles. It carries both how to fetch each asset (the resolution fields) and
what it is (the descriptive metadata). `pkg/marketdata` loads its catalog
from this file (embedded via `datasets.AssetMeta()`); `pkg/suggest` reads the
descriptive fields from the same file. To add or change a bundled asset, edit
this file only.

The descriptive data was collected from issuer factsheets/KIIDs, justETF,
Morningstar and index providers. Breakdowns are approximate (whole percents)
and dated (`as_of`, on the records refreshed since 2026-09; an undated record's
breakdowns are of unknown age); they describe the instrument, not a precise
point-in-time holding. Treat `confidence: medium|low` records, and undated
ones, with extra care and refresh as needed.

## Schema (`assets.json`: array of objects)

Resolution fields (consumed by `pkg/marketdata`):

| field | meaning |
|---|---|
| `id` | canonical identifier (ticker or ISIN); the key |
| `isin` | ISIN, or `null`/empty for indices, spot, futures and the `airfund` FCPEs, which have none. Every tradable record carries one, US-listed funds included: for a US security the ISIN is `US` + the CUSIP + a Luhn check digit, so the way to fill one is to read the CUSIP or ISIN off the issuer's own page and then confirm the result by looking the ISIN itself up on FT's security search, which resolves a US ISIN to its primary listing. Never enter one that a fetched, ISIN-keyed page has not named |
| `aliases` | extra identifiers accepted in portfolio files (e.g. `GOLD`, `NTSX`). An alias shadows every provider ticker of the same spelling, so check the collision before minting one: `GOLD` deliberately resolves to spot gold (`XAUUSD`) and NOT to Barrick Mining, whose NYSE ticker it also is |
| `name` | display name |
| `ucits` | `true` for UCITS funds/ETFs (ETCs, US funds, indices are not) |
| `eu_retail` | `true` when an EU/French retail investor can actually buy it: every UCITS fund, plus EU-listed products with a PRIIPs KID (gold/commodity ETCs, listed closed-end funds like BH Macro). `false` for US-listed funds without a KID. Omitted for non-tradable series (`index` benchmarks, spot, futures). This is the buyability flag; `ucits` alone understates it (no gold product can be UCITS, yet ETCs are freely buyable) |
| `source` | quote provider: `yahoo`, `ft`, `morningstar`, `stooq`, `airfund` (the official daily NAV feed behind a French employee-savings fund's page, see `ERESMONDEM`; no ISIN, no listing, the fund page is the only source) or `index` (served from its embedded reconstruction, no live symbol: a non-investable benchmark like `MSCIWORLD`, or the total-return view of one whose only public series is unusable, like `DTLETR` for the distributing `DTLE`) |
| `symbol` | provider symbol (Yahoo/Stooq ticker, Morningstar id, or the share-class code an `airfund` fund is addressed by); empty for FT and `index` |
| `xid` | FT internal id, or the fund-page widget id an `airfund` request must carry; empty otherwise |
| `nowcast_proxy` | optional: a catalog id whose daily and intraday moves, converted into `currency`, stand in for the fund after its last published NAV (an FCPE prices once a day and publishes with a lag); the estimated tail is flagged `EstimatedFrom` on the series and never cached |
| `nowcast_anchor` | optional, requires `nowcast_proxy`: which print of the proxy the NAV of a day is struck on, `close` (the default, also when the field is absent) or `open`. Set it to `open` only when the fund's own valuation rules say so, as `ERES_DATADOG`'s annual report does: the estimate is then anchored on the proxy's opening price of the last NAV's day, leaving that session's open-to-close move out of it (measured worth on that fund: 3.8 % rmse against 2.8 % on the one-span-ahead estimate). An unavailable open (a day the proxy did not trade, a source with no opening print) falls back on the close silently |
| `fees` | pinned ongoing charge (TER), percent per year; `0` = unknown, or genuinely fee-free for an `index` benchmark |
| `since` | the share class's own official launch date (`YYYY-MM-DD`), from the issuer or justETF, not the date the provider's history happens to start. Nothing is ever trimmed on it; the doctor compares it to the first quote and speaks when they are more than 400 days apart in either direction (see "What `since` means") |

Descriptive fields (consumed by `pkg/suggest`):

| field | meaning |
|---|---|
| `asset_class` | `equity`, `government-bond`, `corporate-bond`, `aggregate-bond`, `inflation-linked-bond`, `money-market`, `gold`, `broad-commodity`, `managed-futures`, `insurance-linked`, `long-volatility`, `tail-risk`, `multi-asset`, `real-estate`, `other` |
| `underlying` | one-line plain description of what it holds |
| `benchmark_index` | the index it tracks, or `active (...)` / `null` |
| `strategy` | open vocabulary; common values: `physical-replication`, `synthetic-swap`, `active`, `futures-overlay`, `leveraged-2x`, `leveraged-3x`, `trend-following`, `long-volatility`, `multi-factor`, `systematic factor tilt`, `covered-call overlay`, `fundamentally-weighted`, `single-stock` (one issuer: it widens the plausibility band, see below, and no advisor ever proposes the record, `suggest.Advisable`), `other` |
| `geography` | approximate region weights (percent), `{ "Global developed": 100 }`, or `null` when not meaningful (gold, broad managed futures, money market) |
| `sectors` | approximate equity sector weights (percent), or `null` for non-equity; for a stacked fund, describes the equity leg |
| `currency` | the quote line's currency: what `symbol`/`xid` actually serves (ISO code, plus `GBp` for a London pence line). Not the denomination, not the exposure; see "What `currency` means" |
| `currency_exposure` | optional look-through fiat exposure: currency (ISO code, plus `None` for real assets and `Dynamic` for futures books) → percent of capital; any shortfall below 100 counts as `None`. Set it only where the automatic derivation (`suggest.CurrencySplit`: hedging, asset class, geography, then quote currency) is wrong: funds denominated differently than their holdings' countries (corporate/aggregate bonds, EM hard-currency debt), mixed-region equity residuals worth resolving |
| `distribution` | `accumulating`, `distributing`, `n/a` |
| `leverage` | `1.0` normal; `2.0` for 2x daily; embedded notional for capital-efficient funds, which is the SUM of `exposures` when that field is set (`1.5` for a 90/60 Efficient Core structure, `2.0` for a 100/100 return-stacked one) |
| `duration` | effective duration in years (fixed income); for a stacked fund, the duration of its bond exposure per unit of notional (e.g. `7.0` for a 90/60 fund's intermediate futures ladder) |
| `notes` | one line on the asset's portfolio role / the market regime it serves |
| `confidence` | `high`, `medium`, `low`: confidence in the breakdowns |
| `sources` | reference URLs: the pages the current values were actually read from |
| `as_of` | the month (`YYYY-MM`) of the factsheet or holdings file the breakdowns (and, for a bond fund, the `duration`) were read from; set only when every breakdown the record carries was re-read, so a record refreshed for its fee or duration alone stays undated, i.e. the date the page itself shows, not the day it was fetched; when geography and sectors come from pages of different months, the older of the two. Absent on a record not refreshed since the field was introduced, whose breakdowns are of unknown age. A dated record's `geography` and `sectors` each sum to exactly 100 (an undated one is tolerated one point either way); `TestCatalogBreakdownsDated` holds both |

## Controlled vocabularies

The composition consumer (`pkg/suggest/composition.go`) aggregates these fields
by canonical label, so `geography`, `sectors` and `currency_exposure` use fixed
vocabularies. `suggest.CanonRegion`, `suggest.CanonSector` and
`suggest.regionCurrency` must understand every label used here; extend those Go
maps only if you genuinely add a new country label, and keep them consistent
with this list.

### Geography

Countries use English standard names, with two abbreviations (`US`, `UK`):
`US`, `UK`, `Japan`, `Germany`, `France`, `Italy`, `Spain`, `Netherlands`,
`Ireland`, `Luxembourg`, `Greece`, `Austria`, `Belgium`, `Finland`, `Portugal`,
`Switzerland`, `Sweden`, `Denmark`, `Norway`, `Canada`, `Australia`,
`New Zealand`, `China`, `Hong Kong`, `Taiwan`, `South Korea`, `India`,
`Singapore`, `Indonesia`, `Thailand`, `Brazil`, `Mexico`, `South Africa`,
`Saudi Arabia`, `Poland` (other real countries allowed, English names).

Region buckets, only for residuals a fund does not break down finer:
`Other eurozone` (euro members not listed), `Other Europe` (European mix
spanning EUR and non-EUR: CHF, SEK, DKK, NOK, GBP), `Other developed`,
`Other emerging`, `Other`. `regionCurrency` maps eurozone labels and
`Other eurozone` to EUR; `Other Europe`, `Other developed`, `Other emerging`
and `Other` map to no single currency (the `Other` currency bucket), so an
equity fund carrying a large such residual may deserve a `currency_exposure`
override. Do not reintroduce eliminated spellings (`United States`,
`United Kingdom`, `Europe`, `Europe ex-UK`, `North America`, `Global`, `Asia`,
`Emerging Markets`, `Latin America`, `Middle East & Africa`, ...); resolve them
to the list above, splitting `North America` into `US` and `Canada`.

### Equity sectors

The 11 GICS sectors exactly, plus `Other` for residuals: `Information
Technology`, `Financials`, `Health Care`, `Consumer Discretionary`,
`Consumer Staples`, `Industrials`, `Energy`, `Materials`, `Utilities`,
`Real Estate`, `Communication Services`. For a stacked fund, `sectors`
describes the equity leg (the equity index it stacks). Bond funds keep their
own bond-sleeve labels (`US Government`, `Investment-Grade Corporate`, ...);
those are ignored by the equity pie and must not be GICS-normalized.

### Plausibility bands per asset class

`asset_class` does more than group the pies: it tells the data doctor what the
record's quotes are allowed to look like. The table lives in
`marketdata.ClassBand` (`pkg/marketdata/plausible.go`) and is reproduced here,
next to the vocabulary it is keyed by, because a new class needs a row in both.

Volatility and CAGR are annualized fractions; "move" is the largest plausible
single observation, "fall" the deepest plausible drawdown. Every bound is
multiplied by the record's `leverage`, so a 3x fund is allowed three times the
move and a 90/60 stacked one 1.5 times the volatility.

| asset_class | volatility | CAGR | move | fall |
|---|---|---|---|---|
| `equity` | 6 to 42 % | -15 to +40 % | 23 % | 97 % |
| `government-bond` | 0.3 to 28 % | -10 to +15 % | 14 % | 80 % |
| `corporate-bond` | 0.5 to 14 % | -8 to +12 % | 10 % | 45 % |
| `aggregate-bond` | 1 to 12 % | -6 to +10 % | 8 % | 30 % |
| `inflation-linked-bond` | 1 to 14 % | -6 to +12 % | 8 % | 30 % |
| `money-market` | 0 to 3 % | -2 to +7 % | 3.5 % | 10 % |
| `gold` | 8 to 30 % | -10 to +25 % | 15 % | 60 % |
| `broad-commodity` | 10 to 45 % | -20 to +25 % | 22 % | 85 % |
| `managed-futures` | 4 to 28 % | -15 to +30 % | 15 % | 45 % |
| `long-volatility` | 4 to 60 % | -30 to +30 % | 25 % | 60 % |
| `tail-risk` | 4 to 60 % | -30 to +30 % | 25 % | 60 % |
| `multi-asset` | 5 to 32 % | -12 to +28 % | 15 % | 45 % |
| `real-estate` | 12 to 38 % | -15 to +25 % | 25 % | 80 % |
| `other` | 0 to 60 % | -35 to +40 % | 25 % | 60 % |

Every row describes a DIVERSIFIED holding. A record whose `strategy` is
`single-stock` names one issuer, whose idiosyncratic risk no class row covers,
so the same row is stretched by 1.75 (`marketdata.singleNameStretch`) on top of
the leverage multiple: the two single-name records here measure 55 and 60 %/yr
against the equity ceiling of 42, and gap 31 % on an earnings print against the
class's 23. The verdict says what widened the band ("at a single name").

They were calibrated on the measured statistics of every bundled record, each
bound set clear of the widest real value in its class, so a full-catalog run
flags accidents and not surprises. Consequences worth knowing before widening
one: the money-market ceiling accepts XEON's 1.0 %/yr, which its 2007 launch
weeks alone produce (the class targets ~0.7 %/yr and every other member sits
below), and its 3.5 % move accepts March 2020's ultrashort credit dislocation
(-3.2 % for ERNA) while rejecting the annual income drops of a distributing
money-market NAV; the corporate-bond floor accepts the AAA CLO funds at 0.8 to
1.1 %/yr; the equity move accepts 1987-10-19 and the equity fall a single
country through its own crisis (Greece, -96 %). A CAGR is only judged over
three years or more, and a single observation only against the pace the series
kept at the time, so a month-end series is not asked to move like a daily one.

Series whose numbers are not prices are exempt from all of it: `source: index`
reconstructions, rate levels, and the continuous-futures symbols behind spot
quotes (`XAUUSD` is served by `GC=F`).

The same table gates the fetch pipeline's round-trip cleaner
(`Band.SpikeLeg`): a leg below four class-level daily sigmas is ordinary
volatility and is never a candidate for removal, whatever else it looks like.

### What `since` means

The share class's own launch date, as the issuer or justETF publishes it. Not
the first date the provider serves, and not a date to be bent until the two
agree: the disagreement is the information. The doctor compares them and speaks
past 400 days in either direction.

- **Quotes start well BEFORE it.** The provider is serving a predecessor under
  this class's name: a merged fund, a converted share class, an older vehicle.
  The statistics are then somebody else's, and the fee load almost certainly
  differs. PFOCX (quotes from 1992-12, class launched 2008-04) is the standing
  example, and it also carries the flat runs and the -12.9 % step of that older
  series.
- **Quotes start well AFTER it.** The depth the record implies does not exist
  at this provider; a deeper listing of the same class may.

Neither is repaired automatically and neither trims anything. Nineteen
records currently say one or the other (2026-09, after the equity refresh
corrected `since` to the issuers' launch dates: six false warnings cleared, the
LifeStrategy pair, IUIT, the S&P 500 equal-weight and semiconductor lines and
VEVE, and one true one added, INFR, whose Yahoo line starts 439 days after its
launch), and all nineteen are true.

### Currency exposure

Keys are ISO 4217 codes, plus `None` (no fiat: real assets) and `Dynamic`
(futures book). No `Other`: resolve every residual to real currencies, or let
the below-100 shortfall count as `None`. Sums above 100 are intended for
levered funds (a 2x USD fund is `{"USD": 200}`); a stacked fund with a
managed-futures leg carries `Dynamic` for that leg's notional.

## Adding an asset

This file is curated. Add a record with the resolution fields (a working
`source`+`symbol`/`xid` and `currency`) and the descriptive fields, then
verify it fetches cleanly with `pofo -verify-data -assets <id>`. The
resolution fields make it part of the bundle (one `-warmup` away); the
descriptive fields feed `-coverage` and `-suggest`.

Read the name the doctor echoes back, not just its verdict: it is the only
place a pinned `symbol` admits it serves a different instrument than the
record claims.

**Run `make verify-catalog` after any catalog edit, and after `make refresh`.**
It is the same doctor over all of it, in each asset's native currency, and it
is what turns `asset_class`, `leverage`, `currency`, `distribution` and `since`
from decoration into checks: a wrong quote line shows up as a volatility
outside the class band, a sibling share class as a name that contradicts
`distribution`, a predecessor's history as a first quote years before `since`.
About a minute on a warm quote cache.

The justETF TER sweep is deliberately NOT in the binary: pinned `fees` change a
few times a year at most, the sweep is one HTTP call per ISIN with a second of
politeness between them, and it needs a human to decide whether a moved TER is
a fee cut or a scraped table. It stays a campaign script, run by hand when the
fees are due a review.

### Picking the quote line: two traps a green fetch hides

- **The symbol must be the share class the ISIN names.** Sibling classes of one
  fund share a family ticker, and a provider's ISIN search happily returns the
  other one. `IE00B3VTMJ91` (iShares € Govt Bond 1-3yr, accumulating) was pinned
  to `IBGS.L`, which is the *distributing* class `IE00B14X4Q57`; the fetch was
  clean and only the name in the log, `(Dist)`, said so. The accumulating class
  trades as `CSBGE3` in Milan and `CBE3` in London.
- **Prefer the listing quoted in the fund's own currency.** A cross-currency
  listing (a euro fund's GBP line in London, a dollar fund's EUR line on Xetra)
  is priced correctly but arrives as one more FX layer, and a provider that
  spliced two such lines into one symbol hides a currency change inside a single
  series. That is the second half of the same bug: Yahoo's `IBGS.L` is the
  fund's EUR NAV in cents until 2008-12-31 and its GBP line from 2009-01-02, a
  104x junction the scale-break repair used to weld shut, baking a fictitious
  -22 % into 2008 as soon as the series was converted back to EUR
  (`mendScaleBreak` now refuses any junction that is not a plain change of
  units, so the cliff reaches the doctor instead). `ITPS.L` has the same shape.
  When a fund's home listing exists, pin that one. The resolver enforces the
  same doctrine when a pinned line fails and a search takes over: candidates
  quoted in the record's `currency` outrank deeper cross-currency listings
  (`fetchSpec.preferCurrency`; born of FOLOW, whose young Paris line lost a
  depth contest to its Swiss CHF listing by four quotes).

One family concentrates both traps. Several iShares funds have, on Yahoo, an
LSE line served in pence (`SEAG.L`, `SEMB.L`, `SEML.L`, `IWDP.L`, `IPRP.L`,
`INFR.L`) alongside a home listing in the fund's own currency. The pence lines
carry, on top of the FX layer, isolated one-day round trips of 30 to 70 %
around 2011 that no other listing of the same fund shows. Symptom: an
aggregate-bond record with 20 % annualized volatility. Prefer `IEAG.AS`,
`IEMB.L`, `IEML.L`, `IQQ6.DE`, `IPRP.AS`, `INFR.MI`.

### What `currency` means (and what `currency_exposure` means)

`currency` is a resolution field: the currency the pinned `symbol`/`xid`
actually serves, verbatim, `GBp` included. It is not the fund's denomination
and not what its holdings are exposed to. Two consumers read it, and the first
is the one that matters:

- `pkg/marketdata` uses it as `fetchSpec.preferCurrency`, the tie-break that
  keeps a fallback search on the record's own listing instead of a deeper
  cross-currency twin. A value disagreeing with the served line aims that
  tie-break at the wrong listing, which is the opposite of its purpose.
- `suggest.CurrencySplit` uses it only as rule 6, the last resort reached by a
  holding that has neither `currency_exposure` nor `geography`: money markets,
  and bonds with no country split.

So when the quote line and the exposure disagree, do not bend `currency`.
Record what the provider serves, then state the exposure where it belongs:
`currency_exposure`, or `geography` (rule 5), or `currency_hedged`+`hedged_to`
(rule 3). And when the served currency is not the fund's own, look for a
listing that is: the two traps above are why one usually exists.

`pofo -verify-data -assets <id>` echoes the served currency next to the name,
and the doctor now says so itself when the two disagree ("catalog says the line
quotes in EUR, the source serves CHF"). `GBp` and `GBP` are deliberately read as
the same line spelled two ways: a hundredfold is the scale-break pass's
business, not the identity check's.

## Provenance and refresh recipes

Field guide for refreshing the descriptive data: per family, where the numbers
live and the dead ends. Each asset's `sources` array holds the deepest stable
URL actually used, and its `notes` line the interpretation hint.

### Where the numbers are, family by family (checked 2026-09)
- **iShares UCITS** (the richest source there is): the Swiss site
  `ishares.com/ch/professionals/en/products/<id>/<slug>` still serves the old
  page, with TER, launch dates, effective duration and weighted maturity in
  plain HTML, and a link to the full holdings CSV (`...<id>/<slug>/<n>.ajax?
  fileType=csv&fileName=<TICKER>_holdings&dataType=fund`, one line per holding
  with sector, location and market currency, "Fund Holdings as of" on its first
  line). Aggregate it by MARKET VALUE, not by the `Weight (%)` column: that
  column is rounded to two decimals, which loses a quarter of a 20,000-line bond
  fund. The product id and slug of every iShares fund come from one JSON file,
  the site's product screener
  (`ishares.com/ch/professionals/en/product-screener/product-screener-v3.1.jsn?dcrPath=/templatedata/config/product-screener-v3/data/en/ch/product-screener/ishares-product-screener-backend-config&siteEntryPassthrough=true`),
  which also carries every fund's TER, effective duration and share-class
  inception date. The UK site was rebuilt in 2026 and no longer carries any of
  it; a US page (`ishares.com/us/products/<id>/...`) embeds its exposure arrays
  as escaped JSON (`"name":"fund"` next to `"name":"type"`), nine countries
  plus "Other", and serves the full holdings file at
  `ishares.com/us/products/<id>/<slug>/latest-holdings.csv` (same columns as the
  Swiss CSV); the US product ids come from the US screener
  (`ishares.com/us/product-screener/product-screener-v3.1.jsn?dcrPath=/templatedata/config/product-screener-v3/data/en/us-ishares/ishares-product-screener-backend-config&siteEntryPassthrough=true`).
- **Broad indices**: read the index's split off a full-replication tracker's
  holdings file (IWDA for MSCI World, IEMG/EEM for MSCI EM/EM IMI, ACWI, CSEMU
  for MSCI EMU): the MSCI factsheet (`msci.com/documents/10199/<uuid>`) names
  only five countries and is a good cross-check; S&P DJI answers 403, and State
  Street publishes the S&P 500's own sector split beside SPY
  (`ssga.com/us/en/intermediary/etfs/spdr-sp-500-etf-trust-spy`). FTSE Russell's
  factsheet (`research.ftserussell.com/Analytics/FactSheets/Home/DownloadSingleIssue?issueName=AWORLDS`)
  has the full country table but classifies by ICB, not GICS: put each ICB
  industry under its nearest GICS label and say so in `notes`. Other FTSE
  codes in use: `AWD` (Developed), `AWALLE` (Emerging), `AWDEURS` (Developed
  Europe), `AWNAMERS` (North America), `WIJPN` (Japan), `AWDPACXJ` (Developed
  Asia Pacific ex Japan), `GEISLMS` (Global All Cap), `GXUSS` (All Cap ex US),
  `GDXUS` (Developed All Cap ex US), `FQEACR` (EM All Cap China A Inclusion);
  some print an ICB supersector table instead of the industry one, whose
  four-digit codes fold to industries by their first two digits.
- **Amundi**: the monthly factsheet PDF
  (`amundietf.lu/pdfDocuments/monthly-factsheet/<ISIN>/ENG/LUX/INSTITUTIONNEL/ETF`;
  French lines answer on `amundietf.fr/.../<ISIN>/FRA/FRA/INSTITUTIONNEL/ETF`)
  prints the benchmark's country and GICS sector bars with their values, the
  running costs, and both a "share-class inception" and a "first NAV" date:
  after the Lyxor mergers the former can postdate the ISIN's own first NAV,
  in which case the first NAV is the launch date.
- **Xtrackers (DWS)**: the pages are JavaScript-only, but the constituent
  export answers a script
  (`etf.dws.com/etfdata/export/GBR/ENG/excel/product/constituent/<ISIN>/`,
  an xlsx with country, currency and industry per line, no as-of date); its
  industries are GICS on some funds and ICB on others.
- **JPMorgan UCITS**: the daily holdings xlsx
  (`am.jpmorgan.com/FundsMarketingHandler/excel?type=dailyETFHoldings&cusip=<ISIN>&country=gb&role=adv&locale=en-GB`,
  country per line, no sector) and the product-data JSON
  (`.../product-data?cusip=<ISIN>&dataType=fund&documentType=FS&country=gb&role=adv&locale=en-GB`:
  `emeaSectorBreakdown`, `emeaRegionalBreakdown`, `shareClass.fees`,
  `shareClassInceptionDate`); the research-enhanced funds use JPM's own
  sector names, which do not map one to one onto GICS.
- **VanEck UCITS**: the holdings xlsx
  (`vaneck.com/<ie|nl>/en/investments/<slug>/downloads/holdings/`) answers
  once the site's cookies are kept, but carries no country or sector; the
  product pages redirect a script to the home page.
- **Vanguard LifeStrategy**: the factsheet gives only the allocation to the
  underlying Vanguard funds; the breakdowns are a look-through of those
  funds' own factsheets (bond sleeves: "Market allocation", ten countries)
  and of their FTSE indices' factsheets (equity sleeves).
- **State Street UCITS**: the daily holdings xlsx
  (`ssga.com/library-content/products/fund-data/etfs/emea/holdings-daily-emea-en-<ticker>-gy.xlsx`,
  trade country and GICS sector per line); the fund page's country table is
  JavaScript-only.
- **Vanguard US**: `investor.vanguard.com/vmf/api/<TICKER>/profile` (expense
  ratio and its date), `/characteristic` (duration, holding count, foreign
  share, benchmark), `/diversification` (sectors by ICB for domestic funds,
  the full country list for international ones, whose sector block is empty:
  read those off the benchmark's FTSE factsheet). Vanguard EMEA: the factsheet
  PDFs on `fund-docs.vanguard.com` (ten countries, ICB sectors).
- **UBS ETFs**: the monthly factsheets on swissfunddata
  (`swissfunddata.ch/sfdpub/docs/fsm-<fund>-<YYYYMMDD>-en.pdf`, month-end
  dated), with modified duration and the country table in text.
- **BlackRock BGF funds**: the factsheet PDF; its breakdown bars carry printed
  values, readable by rendering page 3 (the Read tool on the PDF).
- **Active funds without a text breakdown** (Pictet, Amundi feeders), and ETFs
  whose issuer refuses a script (WisdomTree, Dimensional, Schwab): FT's
  holdings tab (`markets.ft.com/data/funds/tearsheet/holdings?s=<ISIN>:<CCY>`
  for funds, `markets.ft.com/data/etfs/tearsheet/holdings?s=<SYMBOL>` for ETFs,
  the exchange symbol from `markets.ft.com/data/searchapi/searchsecurities?query=<ISIN>`)
  carries Morningstar's sector and REGIONAL split, dated; keep its regions as
  regional buckets, put Morningstar sectors under their GICS names (Defensive
  = Staples, Cyclical = Discretionary), keep what its ten-row tables leave out
  as "Other", and its confidence at medium. stockanalysis.com's holdings page
  shows the top 25 lines only, enough for a fund of a few ETFs (RSSB).
- **Independance AM**: the monthly reporting PDF
  (`independance-am.com/wp-content/uploads/YYYY/MM/YYMMDD-reporting-<fund>-x-eur-c2-en.pdf`)
  is image-only and its geography pie labels slices by colour only, which is
  interpretation rather than reading: refresh it only when the percentages can
  be tied to countries unambiguously.
- Keep region residuals as `Other eurozone/Europe/developed/emerging`; do not
  invent country detail an issuer does not publish. A fund denominated in one
  currency whose issuers sit in many countries (a USD corporate or ultrashort
  fund) takes a `currency_exposure` override, or its new country detail will
  read as that many currencies.

### Equity-leg sectors of stacked funds
- Use the tracked index's sector weights, not a blended fund sheet. NTSG /
  Winton = MSCI World; NTSX / RSST / GDE = S&P 500; NTSZ = MSCI EMU (read off a
  tracker's holdings, see above). RSSB holds its equity leg through ETFs
  (SPTM, VXUS, S&P 500 futures), so its breakdowns are a look-through of those
  lines' own splits, weighted by its holdings. A LifeStrategy-type
  multi-asset sheet lumps the bond sleeve into "Other"; once `exposures` is
  set, replace `sectors` with the equity sleeves' breakdown or the equity pie
  mis-reads it.

### Bond-leg durations
- Efficient Core family (NTSX/NTSI/NTSE/NTSG/NTSZ): a 2/5/10/30y government
  futures ladder, effective ~7y. WisdomTree publishes no single number; the
  ladder and the ~7-7.5y figure come from third-party reviews
  (`optimizedportfolio.com/ntsx`) and match aggregate-bond duration by design.
- Return Stacked (RSBT/RSSB): the bond stack targets the Bloomberg US Aggregate,
  ~6y (returnstackedetfs.com product pages / ReSolve commentary).
- LifeStrategy: global aggregate, EUR-hedged, ~6.5y.

### currency_exposure overrides
- Denomination beats geography for bonds: a EUR- or USD-denominated corporate or
  aggregate fund is 100% its denomination currency regardless of issuer country
  (the name/KID states it, even when the share class quotes in a third currency,
  e.g. IE00B3DKXQ41 quotes GBP but is 100% EUR). Hedged share classes need no
  override (rule 3 handles them). USD hard-currency EM sovereign = 100% USD; EM
  local debt = the index's local-currency basket (JPM GBI-EM Global Diversified
  weights, medium confidence, drifts with the index).
- Stacked funds: equity-leg currencies times leg notional, plus the fund's
  collateral currency for the margin pocket; bond FUTURES legs add no FX; a gold
  leg is `None` (shortfall); a managed-futures leg is `Dynamic` for its notional.

### Fee schedules of SICAV share classes (and their traps)
A Luxembourg SICAV publishes two documents that settle, for free, what a data
aggregator only approximates. Read both before pinning `fees` on any share
class of a multi-class fund.

- The **prospectus supplement** for the sub-fund carries a "Summary of Shares,
  Fees and Expenses" table, one column per share class: investment management
  fee, performance fee rate and hurdle, expense cap, `taxe d'abonnement` and
  minimum initial subscription. `fees` pins the TER, so it is the management
  fee plus the expense cap plus the subscription tax (the cap is a maximum, so
  a published OCF slightly below it is normal, and slightly above it is also
  normal because trading-related expenses sit outside the cap).
- The **audited annual report** carries, in "Statistical Information", the net
  asset value per share for each class at each of the last three financial year
  ends, and in the fees note the performance fee actually charged per class,
  expressed as a percent of average class NAV.

Three traps this combination catches, all of them live in the AQR Managed
Futures classes (LU1662501532 and siblings):

1. **A performance fee is invisible in the TER.** The 0.60% classes of that
   fund pay 10% of the excess over a cash hurdle: 1.5 to 2.1 points of average
   class NAV in the year to 31 March 2026, i.e. three times the pinned `fees`.
   Say so in `notes`; the pin stays the base TER.
2. **A low headline fee can hide a huge minimum.** The 0.75% institutional
   class of that fund requires EUR 80 million, not the six figures an
   aggregator's "institutional" label suggests.
3. **Cheaper on paper is not cheaper in the NAV.** Compare classes on the
   audited year-end NAVs, not on the schedules: over 2023-2026 the 1.00% flat
   class out-returned the 0.75% one by 0.6 to 0.7 points a year, every quarter.

The audited year-end NAVs are also the cheapest identifier check there is: see
`pkg/datasets/golden/aqrmf_test.go`, which pins three consecutive year ends of
two classes to the cent.

### Time sinks and dead ends
- Independance factsheets are image-only PDFs: read as PDF, not HTML.
- justETF's profile page answers a plain HTTP fetch and is the quickest TER,
  distribution and launch-date check for any UCITS ISIN, but its breakdowns are
  four countries plus "Other" and its sector labels are its own (Technology,
  Finance, Consumer Non-Cyclicals...), not GICS: never copy them into `sectors`.
- Answer 403 or render only in JavaScript, so a script gets no number: S&P DJI
  (factsheets and dashboards), WisdomTree (both the .com and the .eu site),
  PIMCO, the DWS/Xtrackers pages (the constituent export works), Boursorama,
  Morningstar's own pages, the UBS fund pages (the swissfunddata PDFs work),
  Schwab, Invesco (the QQQ microsite states its expense ratio in plain HTML),
  Dimensional, and the Eres fund page (its KID PDF works). Never fill a gap with a web-search snippet: a
  number a search engine summarised is not a page read.
- Ossiam CAPE US Sector Value (LU1079841513): the four held sectors rotate
  monthly and are not published as a stable table; any snapshot is low
  confidence (four cheapest-CAPE S&P 500 sectors at 25% each).
