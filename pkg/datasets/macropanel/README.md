# macropanel

`oecd-monthly.csv` is a multi-country **monthly** macro panel, columns
`iso,date,ip,cpi,shortrate,longrate,shareprice` for 30 advanced and large
emerging economies. `date` is `YYYY-MM`; `ip`, `cpi` and `shareprice` are index
levels; `shortrate` and `longrate` are per-cent yields. Cells are left empty
where a series does not cover that month, so the panel is deliberately sparse.

| column | OECD series (dataflow, key) | meaning |
|---|---|---|
| `ip` | `DSD_STES@DF_INDSERV`, `{ISO}.M.PRVM.IX.BTE.Y._Z._Z.N` | production, industry B-to-E, index, seasonally adjusted (growth proxy) |
| `cpi` | `DSD_PRICES@DF_PRICES_ALL`, `{ISO}.M.N.CPI.IX._T.N._Z` | consumer prices, all items, index (inflation) |
| `shortrate` | `DSD_STES@DF_FINMARK`, `{ISO}.M.IR3TIB.PA._Z._Z._Z._Z.N` | 3-month interbank rate |
| `longrate` | `DSD_STES@DF_FINMARK`, `{ISO}.M.IRLT.PA._Z._Z._Z._Z.N` | long-term government bond yield |
| `shareprice` | `DSD_STES@DF_FINMARK`, `{ISO}.M.SHARE.IX._Z._Z._Z._Z.N` | share-price index (capital only, no dividends) |

Two columns have a documented fallback, applied per country-month in priority
order: the primary series owns every month it quotes, the fallback fills only
the rest, so the result does not depend on which download finished first.

- `ip` falls back to manufacturing alone (`ACTIVITY` `C`) where a country has no
  whole-industry aggregate (South Africa) or where its aggregate starts later
  than its manufacturing index (France, Sweden, Greece, Turkey). Being a level,
  the fallback is **rebased** onto the aggregate at the first month both quote,
  so the two never meet at a jump.
- `shortrate` falls back to the immediate (call money) rate `IRSTCI`, which is
  what carries the early decades of half the panel and all of Turkey and Brazil.

Countries the provider has no monthly series for at all: `ip` for Australia and
New Zealand, `cpi` for New Zealand, `longrate` for Turkey. Japan's CPI stops at
2021-06 at the source.

The panel carries the drivers of macro-regime work: **growth x inflation
breadth** (the share of countries whose industrial-production or CPI year-on-year
is accelerating is a smoothed "world point"), and the **monetary quadrant** (the
long vs short rate). The pofo binary embeds this committed CSV via
`pkg/datasets`; it never fetches OECD at runtime. Only ratios of the index
columns are ever read, so their base years do not matter.

## Source & citation

OECD short-term statistics (`DSD_STES`) and prices (`DSD_PRICES`), served
through the free, key-less DBnomics mirror, <https://db.nomics.world/OECD>.
Cite the OECD when reusing. The panel was read from the legacy `OECD/MEI`
dataset until 2026-08; that dataflow stopped being updated in 2024-01 while
still answering HTTP 200, which is why the generator now leads its validation
pass with a freshness check.

## Regenerate

```sh
make macropanel        # fetches the OECD dataflows via DBnomics and rewrites the CSV
```

The generator (`cmd/gen-macropanel`) pulls each series per country from the
DBnomics JSON API with stdlib `net/http`+`encoding/json`, merges each column's
sources deterministically, and writes the long per-country-month table. Two runs
over one vintage of the provider's data produce byte-identical files. Before
writing, it grades the result (`-check`, on by default): freshness per column,
country coverage, a rate series that ends on a run of repeated levels, and one
public anchor per column (the 2020 collapse in US production, the ~9 % US
inflation peak of 2022, the ~5.3 % US 3-month rate of 2023, the 1981 and 2020
extremes of the US long yield, the 2007-2009 fall in US share prices). It
refuses to write if any of them fails.
