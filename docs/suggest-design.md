# `-suggest` design

A portfolio analysis that recommends **catalog assets to add** to an
existing portfolio so it covers the market regimes it is missing, validated
out-of-sample. Phase 2 of the optimizer (after `pkg/optimize`).

Decisions (2026-06-13):

1. **Output**: assets to *add* to the existing portfolio (1–3), each with a
   suggested weight. Not a from-scratch portfolio.
2. **Criterion**: *structure first, then return*: only consider assets that
   fill a coverage/diversification gap, then rank them by an
   out-of-sample-validated risk/return improvement.
3. **Framework**: *macro quadrants* (growth ↑/↓ × inflation ↑/↓), mapped
   from the `asset_class` + `strategy` tags, complemented by a statistical
   diversity measure.
4. **Redundancy**: also flag near-equivalent holdings (e.g. three S&P 500
   trackers = one bet).

## Regimes

Four macro environments; each catalog asset is mapped (deterministically,
from its metadata tags) to the regimes it *helps in*:

| regime | environment | assets that help |
|---|---|---|
| `growth` | rising growth, benign inflation | equities, credit, equity-heavy multi-asset, real estate |
| `deflation` | falling growth, low/falling inflation (recession) | long-duration government bonds, long-volatility, tail-risk, cash/money-market, aggregate bonds |
| `inflation` | rising inflation | gold, broad commodities, energy/gold-miner equities, inflation-linked bonds, real estate |
| `crisis` | protracted dislocations / stagflation / divergent trends | managed futures (trend), gold, long-volatility, broad commodities, global macro |

The mapping is `regimes(asset_class, strategy, hints) -> set`. A few equity
sub-cases are refined by keywords in the underlying/benchmark (gold-miners
and energy → `inflation`). Coverage of a regime = total portfolio weight of
the assets that help in it (an asset can help in several). A regime is a
**gap** when its coverage is below a threshold (default 10 %).

## Diversity (statistical complement)

- **Correlation** of a candidate's daily returns to the current portfolio's
  (lower is better).
- **Diversification ratio** DR = (Σ wᵢσᵢ) / σ_portfolio; 1 when everything
  is perfectly correlated, up to √N when independent. Effective number of
  bets ≈ DR². A good candidate raises DR.

## Redundancy

Within the held assets, group pairs whose daily-return correlation exceeds
0.95 **and** that share an asset class, effectively one bet.
Report each group with its combined weight. The same equivalence is used to
**dedupe candidates** so the tool never suggests a 4th S&P 500 tracker.

## Selection pipeline

1. Build the user's portfolio returns (common window), already fetched.
2. Compute regime coverage from metadata → find the gap regimes.
3. Candidate pool = catalog assets tagged for a gap regime, **not already
   held and not equivalent to a holding**, deduped to one representative per
   (asset_class, benchmark), keeping the cheapest / highest-confidence.
   *This metadata filter happens before any return download*, so only a
   handful of candidates are fetched.
4. Fetch those candidates' returns.
5. For each candidate, over a small weight grid (5/10/15/20 %), build the
   augmented portfolio (existing weights rescaled) and run **walk-forward
   validation**: split the common history into K contiguous windows; in each
   window compare the augmented Sharpe and max-drawdown to the baseline.
   Because the suggestion is a *structural* choice (add asset X at weight w),
   nothing is fitted to returns; the walk-forward purely measures whether
   the benefit is **consistent** across periods, not a one-period fluke.
6. Keep candidates that improve in a majority of windows; rank by median
   out-of-sample Sharpe gain; pick the weight with the best median gain
   (capped). Output the top 1–3.

## Output (terminal, exit-after, mirrors `-verify-data`)

```
Suggestions for <portfolio>

Regime coverage (by weight):
  growth      ████████████ 85 %
  deflation   ██ 10 %
  inflation   · 0 %   ← gap
  crisis      · 5 %   ← gap

Redundancies:
  • CSPX + VUAA + SPYL: 3 S&P 500 trackers (corr > 0.99), 62 % of the portfolio: effectively one bet

Suggestions (fill the gaps, validated out-of-sample):
  1. XAUUSD (gold): fills the inflation gap
     weight 10 %  ·  corr to portfolio 0.08  ·  diversification ratio 1.18 → 1.46
     out-of-sample: Sharpe improved in 9/11 windows, max-drawdown in 8/11
  2. KMLM (managed futures): strengthens crisis coverage
     ...
```

## Packages

- `pkg/suggest` (library brick, stdlib only):
  - `regimes.go`: `Regime`, the tag→regime map, `Coverage`.
  - `diversity.go`: correlation, diversification ratio.
  - `redundancy.go`: `Redundancies`.
  - `meta.go`: `Meta` struct + `LoadMeta(io.Reader)` (parses assets.json).
  - `suggest.go`: `Suggest(portfolio, candidates, opts) -> []Suggestion`,
    walk-forward robustness.
  - tests with closed forms / synthetic correlated series.
- `datasets.AssetMeta() []byte` embeds `assets.json` (done).
- `cmd/pofo`: `-suggest` flag; metadata-filter candidates, fetch only
  those, render the terminal block, then exit. Reuses `fetchAsset`,
  `metrics.Compute`, `marketdata` catalog.

The **coverage** chart (and the asset-class column) is also rendered in the
normal HTML report and the `-cli` summary, computed from the metadata at no
extra fetch cost. The full gap-filling **suggestions** (which require
fetching candidate histories) stay in the dedicated `-suggest` terminal mode.

Two follow-ups shipped after v1:

- **`-coverage`**: an offline advisor: the coverage chart plus, for each
  gap, the catalog assets that fill it grouped by asset class. No price
  downloads, no ranking; `-suggest` ranks them afterwards.
- **`-framework`**: the classification is pluggable (`suggest.Framework`):
  `regimes` (default, the four macro quadrants) or `factors` (market, size,
  value, momentum, quality, term, credit, alternative, cash). The factor
  mapping is intentionally coarse: diversifiers that are not Fama-French
  factors land in *alternative*, so regimes stay the default.

Still out of scope: HTML rendering of the ranked suggestions; suggesting
uncatalogued assets.
