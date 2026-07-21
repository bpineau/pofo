# broadsample

`country-real.csv`: per-country **real** annual total returns (fractions) for
18 advanced economies, 1870-2020, columns `iso,year,equity,bond,bill`. Each
country's nominal returns are deflated by its own CPI.

The FIRE explorer (`pkg/decumul/web`) pool-bootstraps single-market runs from
this table (see `scenario.PooledBootstrap`), so national disasters survive at
full force. It is deliberately **not** a diversified world index: diversifying
first would erase the very sequence risk this dataset exists to capture. The
pooled real equity geometric mean is ~4.4%/yr with a fat left tail, reproducing
the broad-sample safe-withdrawal evidence (Anarkulova, Cederburg & O'Doherty).

## Source & citation

Jorda, O., Schularick, M., Taylor, A. M., *Macrofinancial History and the New
Business Cycle Facts*, NBER Macroeconomics Annual 2016, vol. 31. Macrohistory
Database R6, <https://www.macrohistory.net/database/>. Free to use with
citation; cite JST when reusing the derived series.

## Regenerate

```sh
make broadsample        # downloads the JST xlsx and rewrites country-real.csv
```

The generator (`cmd/gen-broadsample`) parses the xlsx with stdlib
`archive/zip`+`encoding/xml`, deflates each country-year by its CPI, and writes
the long table. The pofo binary never fetches JST at runtime; it embeds this CSV
via `pkg/datasets`.
