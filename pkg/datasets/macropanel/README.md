# macropanel

`oecd-monthly.csv` is a multi-country **monthly** macro panel, columns
`iso,date,ip,cpi,shortrate,longrate,shareprice` for 30 advanced and large
emerging economies. `date` is `YYYY-MM`; `ip`, `cpi` and `shareprice` are index
levels; `shortrate` and `longrate` are per-cent yields. Cells are left empty
where a series does not cover that month, so the panel is deliberately sparse.

| column | OECD MEI series | meaning |
|---|---|---|
| `ip` | `PRINTO01.IXOBSA` | industrial production, index, seasonally adjusted (growth proxy) |
| `cpi` | `CPALTT01.IXOB` | consumer prices, all items, index (inflation) |
| `shortrate` | `IR3TIB01.ST` | 3-month interbank rate (immediate rate `IRSTCI01.ST` where absent) |
| `longrate` | `IRLTLT01.ST` | long-term government bond yield |
| `shareprice` | `SPASTT01.IXOB` | share-price index (capital only, no dividends) |

The panel carries the drivers of macro-regime work: **growth x inflation
breadth** (the share of countries whose industrial-production or CPI year-on-year
is accelerating is a smoothed "world point"), and the **monetary quadrant** (the
long vs short rate). The pofo binary embeds this committed CSV via
`pkg/datasets`; it never fetches OECD at runtime.

## Source & citation

OECD, *Main Economic Indicators*, served through the free, key-less DBnomics
mirror, <https://db.nomics.world/OECD/MEI>. Cite the OECD MEI when reusing.

## Regenerate

```sh
make macropanel        # fetches OECD MEI via DBnomics and rewrites the CSV
```

The generator (`cmd/gen-macropanel`) pulls each series per country from the
DBnomics JSON API with stdlib `net/http`+`encoding/json`, and writes the long
per-country-month table. Note the OECD MEI frozen mirror ends around late 2023.
