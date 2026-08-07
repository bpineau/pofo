# Offline-first CPI/HICP deflators

## Motivation

A normal `pofo <portfolio>` run now shows real (inflation-adjusted) columns and
therefore needs a CPI deflator: `^CPI-US` for USD, `^HICP-FR` for EUR. Both are
fetched **live-first** (`fetchCPIUS` -> FRED, `fetchHICP` -> Eurostat) through
`cachedHistory`, whose order is memo -> fresh disk cache (< `MaxAge`, 30 days)
-> download -> embedded snapshot only on error. So on any cold or expired cache
the run downloads them inline (the "downloading ^CPI-US via fred..." line the
user saw), and results depend on network state.

Every other bundled long-history series (S&P 500, treasuries, gold, FX) is
served offline-first from an embedded snapshot. CPI/HICP are the exception, even
though their embeds already carry the full history: `cpi-us.csv` 1913->2026-05,
`hicp-fr.csv` ->2025-12. Nothing needs adding; the embed just is not the primary
source.

## Goal (one sentence)

Serve the `^CPI-US` and `^HICP-FR` deflators from their embedded snapshots by
default so a normal run never downloads them, while `-warmup` still refreshes
them from the live source.

## Non-goals

- Changing any other series' fetch policy (Yahoo, refdata, simdata unchanged).
- Windowing/altering the embed contents or the deflation math (deflation aligns
  by date and picks its base from the value window, so extra earlier CPI points
  are inert; results are unchanged).
- Wiring embeds for HICP geographies that have none (`^HICP-EA`, `^HICP-DE`,
  ...): those keep the current live path.

## Behaviour (confirmed)

- `pofo <file>`: memo -> fresh disk cache (if a prior warmup wrote one) ->
  embedded snapshot. No network for `^CPI-US`/`^HICP-FR`.
- `pofo -warmup`: fetches `^CPI-US` and `^HICP-FR` live and writes the disk
  cache; a later normal run then prefers that fresher cache over the embed
  (until it ages past `MaxAge`, after which it falls back to the embed, still
  no inline download).
- A HICP geography without an embed: unchanged (live via `cachedHistory`).

## Architecture

A single policy helper on `Client`, used by both fetch functions:

```go
// embeddedHistory serves a bundled long-history index offline-first:
// memoized, then a fresh disk cache, then the embedded snapshot. The live
// source is consulted only when RefreshInflation is set (warmup) or when no
// embed is available for this id.
func (c *Client) embeddedHistory(
	ctx context.Context, source, id string, from time.Time,
	embed func() (*Series, bool), live func() (*Series, error),
) (*Series, error)
```

Logic:
1. `c.RefreshInflation` true -> delegate to `cachedHistory(...live...)` (live
   with its existing embed-on-error fallback), refreshing the disk cache.
2. memo hit -> return.
3. fresh disk cache (`loadCache`) -> memoize, return.
4. `embed()` available -> memoize, return (full embedded series, as the current
   error-fallback already does; not windowed).
5. no embed -> delegate to `cachedHistory(...live...)`.

`fetchCPIUS` supplies `embed = func() (*Series, bool) { return cpiUSSeries(parseMonthlyAnchors(cpiUSSnapshot)), true }`
and `live = func() (*Series, error) { m, err := c.fetchFRED(ctx, "CPIAUCNS"); ...; return cpiUSSeries(m), nil }`.
`fetchHICP` supplies `embed` from `embeddedHICP(geo)` (returns `false` for geos
without a snapshot) and `live = func() { return c.downloadHICP(ctx, symbol, geo) }`.

New field `Client.RefreshInflation bool` (default false, zero value keeps the
offline-first behaviour). `runWarmup` sets it true and, after the catalog loop,
explicitly fetches `^CPI-US` and `^HICP-FR` so warmup refreshes them.

The memo/cache key must match `cachedHistory`'s convention exactly
(`source + ":" + viewKey(id, false) + "|" + from`) so the two paths share
memoization.

## Error handling

- Live fetch failing during warmup still falls back to the embed (unchanged
  `cachedHistory` behaviour), so warmup never hard-fails on these.
- `CacheDir == ""` (no disk cache): step 3 is skipped, embed serves directly.

## Testing

- `pkg/marketdata/cpius_test.go`: `TestFetchCPIUSLive` must set
  `c.RefreshInflation = true` before `Fetch`, since the live path is now
  refresh-only. Add `TestFetchCPIUSEmbedFirst`: a default client whose FRED
  endpoint would error still returns the embed (1913-> series) without touching
  the network (assert a long series and `Source`/name from the embed).
- `pkg/marketdata/eurostat_test.go`: the equivalent HICP live test set to
  refresh mode; an embed-first test for `^HICP-FR`; confirm a geo without an
  embed (e.g. `^HICP-EA`) still hits the live path in default mode.
- `make check` green; `make golden` green (deflation results unchanged).

## Docs

- Godoc on `fetchCPIUS`/`fetchHICP` and `cpiUSSnapshot`/`hicpFRSnapshot` updated
  from "live preferred, embed on error" to "embed served by default, live only
  on -warmup".
- `pkg/marketdata/doc.go` conventions note for the `^CPI-US`/`^HICP-<geo>`
  family updated to state they are offline-first.
- CLAUDE.md "Long-history data sources" memory context already lists the embed
  provenance; the regen recipe stays in the file-level comments (unchanged).
