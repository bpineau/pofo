# World-equity reconstructions: cap-weighted, not fixed-weight

## The bias

The three world-equity reconstructions (`VT`, the FTSE All-World UCITS class
`IE00BK5BQT80`, and the equity sleeve of `RSSB`) were built as a FIXED
60 % US / 30 % developed-ex-US / 10 % emerging blend, rebalanced daily, over
their whole past.

Those weights are roughly today's global split. Carrying them back to 1988 is
a look-ahead bias, not an approximation: the US share of the world market was
near 30 % at the 1989 Japan peak and above 50 % in 2000-2002, and a
reconstruction that holds 60 % US throughout earns the US market's
outperformance in exactly the decades when the world did not own it. Measured
on the legs themselves, the fixed blend ran about +1.8 pts/yr above the
cap-weighted one over 1988-2008.

A daily rebalance makes it worse rather than better: it also enforces the
constant weights, which no index does.

## The fix: hold the anchor's shares

A cap-weighted index does not rebalance. Between reconstitutions it is a
buy-and-hold portfolio, and its country weights move with the countries' own
returns:

```
w_i,t-1  proportional to  w_i,t / (1 + r_i,t)      (backward)
w_i,t+1  proportional to  w_i,t x (1 + r_i,t+1)    (forward)
```

which is what holding one dated basket of shares produces in both directions.
`CapWeighted` (`pkg/simgen/capweight.go`) does exactly that: it takes the legs'
published split on an anchor date and never rebalances, so no weight rule is
imposed at all. `capWeights` reads the implied weights back out, which is what
the validation below checks. `Composite` (constant weights, daily rebalanced)
stays for the recipes that replicate a fund which really does rebalance to
fixed weights.

What the drift IGNORES is net issuance: weights also move when companies issue
or buy back shares, when free float opens (China A shares entering the global
indices in 2018-2019) and when an index changes its own coverage. None of that
is a return, and its size is measured below rather than waved away.

## The anchor

The FTSE All-World index column of the market-diversification table in
Vanguard Total World Stock Index Fund's annual report for the fiscal year ended
2009-10-31 (SEC Form N-CSR, accession 0000932471-09-002118, filed 2009-12-29):

| region | index |
|---|---|
| United States | 40.4 % |
| Europe | 28.5 % |
| Pacific | 14.3 % |
| Canada | 3.3 % |
| Emerging markets | 13.5 % |

Canada joins the developed-ex-US leg, the only leg that can hold it: 40.4 %
US / 46.1 % developed-ex-US / 13.5 % emerging (`worldLegs`, `worldAnchor`).

Why that date rather than the newest split: every reconstruction here serves
the years BEFORE its fund existed, so the drift runs backward and the error
accumulates INTO the past. 2009-10-31 is the earliest published split of the
world fund's own index, which puts the anchor as close as the evidence allows
to the window that is actually consumed.

## Validation

The weight path is the evidence, because the real windows are short and recent.
Neither check below is fitted: the anchor is a single dated observation.

| date | reconstruction | published / cited |
|---|---|---|
| 1989-12 | US 28.5 %, dev-ex-US 66.2 %, EM 5.3 % | Japan ALONE was 40 % of the global market at its 1989 zenith (Dimson, Marsh and Staunton data, as reported in CFA Institute's February 2025 report on the 60/40 portfolio), which a 66.2 % developed-ex-US leg can hold and a 30 % one cannot. The often-quoted "about 30 % US in 1989" agrees with the 28.5 % here but is not verified against a primary source in this repository |
| 2000-12 | US 52.1 % | above half at the dot-com peak, as expected |
| 2001-12 | US 54.5 % | " |
| 2002-12 | US 52.2 % | " |
| 2009-10 | US 40.4 % | the anchor itself |
| 2023-10 | US 66.1 % | 60.9 %, published in the same report series (fiscal year ended 2023-10-31, fund allocation). The 5.2-point overshoot over fourteen years is the net-issuance effect the mechanism ignores |

Against the real funds, without the graft, on identical windows (monthly, the
cadence these comparisons are readable at):

| fund | window | real | fixed blend | cap-weighted |
|---|---|---|---|---|
| VT | 2008-07 to 2026-08 | 9.15 %/yr | +0.76 pts/yr, monthly corr 0.9959 | +0.24 pts/yr, monthly corr 0.9971 |
| RSSB | 2024-01 to 2026-08 | 18.11 %/yr | +3.14 pts/yr | +2.98 pts/yr |
| VWRA class | 2019-08 to 2026-05 | 12.64 %/yr | +2.15 pts/yr, monthly corr 0.8098 | +2.24 pts/yr, monthly corr 0.8183 |

VT, the one fund with a long real window, improves on the level and on the path
at once. RSSB improves on the level. The UCITS class is 0.09 pts/yr worse on
the level and better on the path, which is the forward drift overshooting the
US weight between 2009 and 2019 (its own remaining gap is dominated by
something else: a London-close NAV against US-close donors, hence its 0.81
monthly correlation).

## Bounds

- The three files start at 1987-12, where the emerging leg's reference begins.
  Nothing reaches further back, and the blend is not extrapolated past it.
- The anchor is a split of the FTSE All-World (large+mid). `VT` tracks the
  FTSE Global All Cap, which adds a small-cap tail whose regional split differs
  slightly. That difference is smaller than the rounding of the published
  table and is not modelled.
- The developed-ex-US leg (`VTMGX`, extended by `DEVEXUS-USD`) did not always
  include Canada: its index was Europe-plus-Pacific for much of the deep era.
  Canada's 3.3 % therefore rides on a leg that historically excluded it, which
  is the best of the three available homes for it, not a faithful one.
