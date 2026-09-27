# Inflation deflators: sources, the 2026 HICP rebase, the freshness guards

The real (inflation-adjusted) side of pofo rests on two index levels:
`^HICP-<geo>` (Eurostat's Harmonised Index of Consumer Prices, `^HICP-FR`
for euro reports, bundled from 1955 through the OECD French CPI) and
`^CPI-US` (BLS CPI-U all items, not seasonally adjusted, via FRED
`CPIAUCNS`, bundled from 1913). Every consumer holds the last level flat past
a deflator's end (`scenario.Deflate`'s `hicpAt`, `pkg/compare`'s `deflate`),
so a deflator that stops updating does not fail anything: it reads every
later month as zero inflation, in the report's real statistics, in the FIRE
engine's real returns, everywhere.

## The trap: a statistics office rebases and the old dataset stops

In 2026 Eurostat moved the HICP to the ECOICOP 2 classification and a
2025=100 base, published under a NEW dataset, `prc_hicp_minr`, and stopped
updating the dataset pofo read, `prc_hicp_midx` (last month 2025-12, last
update 2026-02-06). The old dataset kept answering HTTP 200 with a full,
plausible history, and `make snapshots` regenerated `hicp-fr.csv` on
2026-09-27 still ending at 2025-12: every check it ran (rows lost, history
shortened, values revised) passed, because nothing had moved. Found by hand
nine months after the last month, with the 2026 euro-area annual rate at 3.2 %.

`prc_hicp_minr` carries five units: `I25` (2025=100), `I15` (2015=100),
`RCH_M`, `RCH_A`, `RCH_MV12MAVR`; one `coicop18` category for all items,
`TOTAL`; history from 1996-01 for the main geographies. Its months before
2016 carry Eurostat's `e` (estimated) flag: they are the back-calculation of
the new classification. pofo reads `unit=I15&coicop18=TOTAL`, which keeps
the 2015=100 base, so every level stays where it was and the bundled
snapshot's OECD head (chained at 1996-01) needs no rebuild.

Geographies: the new dataset drops `EU28` and adds `EA21` and `GE`; `FR`,
`EA`, `EA20`, `DE` and the other member states are in both.

## Validation record (2026-09-27)

Old `prc_hicp_midx` (`unit=I15`, `coicop=CP00`) against new `prc_hicp_minr`
(`unit=I15`, `coicop18=TOTAL`), common span 1996-01 to 2025-12 (359 monthly
changes):

| Geo | Worst monthly-change gap | Level ratio new/old | Since 2016 |
|---|---|---|---|
| FR | 0.000 pt | 1.00000 throughout | identical |
| EA | 0.013 pt (2001-10) | 0.99988 to 1.00013 | identical |

Annual rates computed from the new `I15` levels against the published
`RCH_A` of the same dataset, 2026-01 to 2026-08: worst gap 0.04 pt for FR,
0.05 pt for EA, i.e. the rounding of a rate published to one decimal. August
2026: FR 2.63 % computed against 2.6 % published, EA 3.24 % against 3.2 %.

`^CPI-US` was checked at the same time and is current (FRED `CPIAUCNS`,
2026-08).

## The guards

- `cmd/gen-snapshots` refuses to write a snapshot whose last row trails the
  run by more than its cadence allows (`stale`: three weeks after a daily row,
  ten weeks after the end of a monthly row's month).
- `TestSnapshotsCurrentWhenGenerated` (`pkg/marketdata`) holds each committed
  snapshot's last row to its own `# generated:` stamp with the same
  allowances, so a file refreshed from a frozen source fails `make check`;
  it compares two dates in the file and never the clock.
- The live Eurostat path (`downloadHICP`, the only path for a geography
  without a snapshot, such as `^HICP-EA`) logs a `warning:` when the series
  trails the current month by more than three months (`hicpLagging`).

- The disk cache files a HICP under `eurostat.<dataset>` (`hicpCacheSource`),
  so a move to another dataset refetches rather than replaying the old one's
  copy for up to `MaxAge`, or at any age to an `Offline` client. Copies filed
  under the former `eurostat` identity are simply never read again.

When one of them fires, look for where the series went (a new dataset, a new
unit, a new base) before touching any tolerance.
