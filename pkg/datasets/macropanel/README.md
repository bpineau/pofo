# macropanel

`oecd-monthly.csv` is a multi-country **monthly** macro panel, columns
`iso,date,ip,cpi,shortrate,longrate,shareprice` for 30 advanced and large
emerging economies. `date` is `YYYY-MM`; `ip`, `cpi` and `shareprice` are index
levels; `shortrate` and `longrate` are per-cent yields. Cells are left empty
where a series does not cover that month, so the panel is deliberately sparse.

| column | OECD series (dataflow, key) | meaning |
|---|---|---|
| `ip` | `DSD_STES@DF_INDSERV`, `{ISO}.M.PRVM.IX.BTE.Y._Z._Z.N` | production, industry B-to-E, index, seasonally adjusted (growth proxy) |
| `cpi` | `DSD_PRICES_COICOP2018@DF_PRICES_C2018_ALL`, `{ISO}.M.N.CPI.IX._T.N._Z` | consumer prices, all items, index (inflation) |
| `shortrate` | `DSD_STES@DF_FINMARK`, `{ISO}.M.IR3TIB.PA._Z._Z._Z._Z.N` | 3-month interbank rate |
| `longrate` | `DSD_STES@DF_FINMARK`, `{ISO}.M.IRLT.PA._Z._Z._Z._Z.N` | long-term government bond yield |
| `shareprice` | `DSD_STES@DF_FINMARK`, `{ISO}.M.SHARE.IX._Z._Z._Z._Z.N` | share-price index (capital only, no dividends) |

Three columns have a documented fallback, applied per country-month in priority
order: the primary series owns every month it quotes, the fallback fills only
the rest, so the result does not depend on which download finished first.

- `ip` falls back to manufacturing alone (`ACTIVITY` `C`) where a country has no
  whole-industry aggregate (South Africa) or where its aggregate starts later
  than its manufacturing index (France, Sweden, Greece, Turkey). Being a level,
  the fallback is **rebased** onto the aggregate at the first month both quote,
  so the two never meet at a jump.
- `shortrate` falls back to the immediate (call money) rate `IRSTCI`, which is
  what carries the early decades of half the panel and all of Turkey and Brazil.
- `cpi` falls back to the COICOP 1999 dataflow `DSD_PRICES@DF_PRICES_ALL`, same
  key. The OECD is migrating its price statistics to COICOP 2018 country by
  country, and a country that has moved simply STOPS in the old dataflow while
  the old dataflow keeps answering for the years it already holds: Japan stopped
  there at 2021-06, Mexico at 2024-07, South Africa at 2025-01 and eighteen
  European members at 2025-12. Nine countries (USA, DEU, GBR, AUS, KOR, POL,
  BRA, IND, NZL) are the other way round and have no COICOP 2018 series at all,
  which is why both are read and the newer classification wins. Where the two
  overlap they carry the same index on the same base, so the rebasing is a
  formality here (every measured factor was 1.0000).

Countries the provider has no monthly series for at all: `ip` for Australia and
New Zealand, `cpi` for New Zealand, `longrate` for Turkey.

The panel carries the drivers of macro-regime work: **growth x inflation
breadth** (the share of countries whose industrial-production or CPI year-on-year
is accelerating is a smoothed "world point"), and the **monetary quadrant** (the
long vs short rate). The report's regime strip reads the growth x inflation
breadth (`pkg/compare/regime.go`). The pofo binary embeds this committed CSV via
`pkg/datasets`; it never fetches OECD at runtime. Only ratios of the index
columns are ever read, so their base years do not matter.

## Source & citation

OECD short-term statistics (`DSD_STES`) and prices (`DSD_PRICES_COICOP2018`,
falling back to `DSD_PRICES`), read from the OECD's own key-less SDMX API,
<https://sdmx.oecd.org/public/rest>. Cite the OECD when reusing. Until 2026-09
it came through the DBnomics mirror, whose OECD copy was last indexed
2026-06-16 and left the panel stopped at 2026-05 while the OECD served 2026-08.
The panel was read from the legacy `OECD/MEI`
dataset until 2026-08; that dataflow stopped being updated in 2024-01 while
still answering HTTP 200, which is why the generator now leads its validation
pass with a freshness check. That check is run twice: once per column, and once
per COUNTRY against the newest month its column reaches, because a single
country freezing (as Japan's CPI did, at 2021-06, when it moved to COICOP 2018)
is invisible to the first.

## Regenerate

```sh
make macropanel        # fetches the OECD dataflows from the OECD API and rewrites the CSV
```

The generator (`cmd/gen-macropanel`) reads the panel through
`cmd/internal/refgen.OECD`, ONE download per dataflow (every country and every
fallback key of a dataflow in a single SDMX request, four downloads in all),
because the API admits 60 downloads an hour and answers the next with HTTP 429;
it merges each column's sources deterministically and writes the long
per-country-month table. Two runs over one vintage of the OECD's data produce
byte-identical files. Before writing, it grades the result (`-check`, on by
default): freshness per column (three months, four for production) and per
country, country coverage, a rate series that ends on a run of repeated levels,
and one public anchor per column (the 2020 collapse in US production, the ~9 %
US inflation peak of 2022, the ~5.3 % US 3-month rate of 2023, the 1981 and
2020 extremes of the US long yield, the 2007-2009 fall in US share prices). It
then compares the rebuilt panel with the committed one, column by column, and
refuses to write if any of the checks fails or if a rate more than a year old
moved by more than 0.10 pt; revisions of the index columns (seasonal
adjustment, national rebasing) are reported, not refused.
