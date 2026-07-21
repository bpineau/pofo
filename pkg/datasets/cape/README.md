# cape

`shiller-cape.csv`: Robert Shiller's **CAPE** (PE10) monthly series,
`date,cape`, 1881 onward. CAPE is the real S&P 500 price divided by the 10-year
average of real earnings; its inverse (`1/CAPE`) is a first-order estimate of
the next decade's real return.

The FIRE explorer uses it to anchor the central case to today's valuation
(`capeAdjust`): at a rich CAPE the next decade is compressed, which is exactly
the sequence that makes or breaks an early retirement. The bundled series also
feeds the valuation percentile shown in the explorer.

## Source

Datahub mirror of Robert Shiller's dataset (the `PE10` column of the
`s-and-p-500` dataset), the same source as the bundled SP500 history:
<https://raw.githubusercontent.com/datasets/s-and-p-500/main/data/data.csv>.
The derived PE10 column lags the raw price by a couple of years; the snapshot is
dated (`asOf`) so a stale tail is visible rather than silent.

## Regenerate

```sh
make cape
```
