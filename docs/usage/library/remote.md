# Numerical API and MCP

`pkg/remote` serves five operations through one numerical service. It uses
`analyze`, `fire` and `decumul`; both transports return the same domain object.
No SDK or third-party Go dependency is required.

| Operation | HTTP route | Result |
|---|---|---|
| `search_assets` | `GET /api/v1/assets/search` | Catalog identities, paginated without network access |
| `analyze_assets` | `POST /api/v1/assets/analyze` | Identity, provenance and selected asset statistics |
| `analyze_portfolio` | `POST /api/v1/portfolios/analyze` | One to four allocations on one common holding window |
| `simulate_fire` | `POST /api/v1/fire/simulate` | Annual real-EUR withdrawal outcomes by model |
| `solve_fire` | `POST /api/v1/fire/solve` | Verified fixed-rule spending or capital within stated bounds |

## Start the server

Both endpoints are disabled by default and enabled independently:

```sh
./pofo -serve -serve-api -serve-mcp -listen 127.0.0.1:8787
```

`GET /api/v1/openapi.json` describes input and output shapes; `/api/v1/` is a
small documentation page. MCP clients connect to `/mcp`. The server implements
revision `2025-11-25`, stateless Streamable HTTP, ordinary JSON responses and
five read-only tools. It deliberately returns one compact JSON text block,
without `structuredContent` or `outputSchema`. GET `/mcp` returns 405; there
are no SSE streams, sessions, resources, tasks or stored portfolios.

FIRE schemas keep all input properties on their containing object. Conditional
requirements use JSON Schema `if`/`then`/`else`: portfolio versus parametric
market inputs, and the known amount versus the solved amount. Partial `oneOf`
branches can hide the shared fields from clients that generate tool argument
types from those branches. Runtime validation remains authoritative.

```sh
curl 'http://127.0.0.1:8787/api/v1/assets/search?query=IWDA&limit=5'
curl -H 'Content-Type: application/json' \
  -d '{"ids":["IWDA"],"sections":["identity","statistics"]}' \
  http://127.0.0.1:8787/api/v1/assets/analyze
curl -H 'Content-Type: application/json' \
  -d '{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":30},"market":{"mode":"parametric","real_mean_return_pct":4,"annual_volatility_pct":15,"degrees_of_freedom":5},"models":["central","stress"]}' \
  http://127.0.0.1:8787/api/v1/fire/simulate
```

All money and allocations belong in POST bodies. Responses are `no-store`.
New-route access logs omit query strings, referrers, user-agent strings and
bodies. Market-data diagnostics remain enabled, including asset identifiers and
fallback warnings; they do not log submitted allocations, labels, capital or
spending. Aggregate operation costs and response warnings remain available.
No share URL is generated and no submitted portfolio is persisted.

## Direct Go use

```go
// from remote.ExampleService_Execute
// Catalog search needs neither a data client nor network access.
service := remote.New(nil, "example")
result, err := service.Execute(context.Background(), "search_assets", []byte(`{"query":"IWDA","limit":1}`))
if err != nil {
	panic(err)
}
page := result.Data.(remote.SearchData)
fmt.Println(len(page.Items), page.Items[0].ISIN)
```

Use `marketdata.NewClient` for fetching or `analyze.SeriesSource` for prepared
in-memory fixtures. `Service.Execute` performs validation, execution and compact
projection. `Service.BeforeFetch` lets a host apply the same foreign-ID quota
as its other routes. Custom sources must honor context cancellation and the
cache-only policy; the provided client does. HTTP hosting separately supplies
shared admission through `HTTPOptions.Acquire` and trusted client identity
through `HTTPOptions.ClientID`.

## Contract and interpretation

Percentages use `5` for five percent, dates use `YYYY-MM-DD`, and unknown or
undefined requested measurements are `null`. Unknown fields, duplicate keys,
trailing JSON and unsupported configurations are errors. Optional pointers
preserve explicit zero. `meta.defaulted` names every filled default; engine
and dataset revisions identify the computation. Defaults may change with a
new build, so supply explicit assumptions for stable inputs.

`pkg/remote/testdata/consumer_cases.json` holds 24 consumer prompts with explicit
calls and expected interpretations. `TestConsumerCorpus` checks them through
the Go service, HTTP and MCP against deterministic fixtures: identity, observed
versus reconstructed history, percentage units, explicit zeros, undefined
statistics, shared windows, model basis, bounded solves and rejected inputs.
This is a transport and numerical contract corpus. It does not measure an
LLM's ability to choose a call or explain its result; that requires running
the prompts with a named model and grading its actual responses separately.

Asset/portfolio returns are nominal. CAGR is geometric annual growth;
volatility annualizes periodic standard deviation at the reported cadence;
Sharpe and Sortino use zero risk-free return. Maximum drawdown is negative.
Annual rows expose their actual first/last observation dates. Drawdowns are
ranked deepest first, at most ten. Relative statistics use the benchmark's
own overlap; they do not shorten the comparison window.

Holdings are nonnegative percentages summing to 100 within `1e-6` percentage
point. Duplicate canonical holdings are rejected. Default rebalancing is 90
calendar days; explicit zero means buy-and-hold. An annual portfolio fee is
additional to asset charges already included in prices. Monthly shared
intervals underlie correlations and attribution. Composition reports notional
exposures; sector percentages use the equity sleeve as denominator. Stacked
exposures need not sum to 100. Attribution excludes unallocated envelope fees
and is the monthly Euler variance share or simple return-contribution share.

A bare quoted fund ID means observed prices. `SIM` or `history=extended`
explicitly requests reconstructed history, with provenance and existing data
warnings preserved. IDs always resolve exactly. Paths, URLs and imported
visualizer links are not accepted. Rate and inflation levels are unsuitable
for investable-return statistics.

FIRE uses the [shared numerical builder](fire.md), annual withdrawal steps,
real EUR, one pension and fixed horizons. Its portfolio mode fits EUR holdings
against French HICP with fixed monthly weights. At least 24 consecutive return
months are required. This differs from portfolio analysis rebalancing.
The six models are central, bootstrap, historical, stress, broad_sample and
lost_decade. Bootstrap and historical require portfolio mode; historical needs
a full horizon and samples overlapping cohorts with replacement. Broad-sample
is an external domestic 60/40 reference, not the submitted allocation.

The shared tax/gain/inflation defaults are planning assumptions. The current
tax default is explicitly labelled a French taxable-account assumption to
verify or override, not a statutory rate. CAPE adjustment uses the dated
bundled observation, reports staleness and affects only central/stress/lost_decade.
Zero failures is not proof of zero risk. Household spending includes pension;
cut timing and duration are conditional on paths that cut. Tax-rate medians
exclude paths without gross withdrawals; decade metrics exclude paths lacking
a complete positive-wealth decade. Undefined populations remain null.

`solve_fire` supports `fixed_real` only. A measured counterexample invalidates
monotone capital bisection for Guyton-Klinger; adaptive simulation remains
available. Default spending bounds are `[0, 0.15*capital]`, capital bounds
`[200000, 6000000]`. Bounds are visible, never universal economic limits.
For an unreachable solve, `meta.assumptions.plan_role=favorable_bound_evaluation`
labels the tested favorable endpoint; `data.effective_plan` remains null.
Statuses distinguish `solved`, `boundary_satisfies`, and
`target_unreachable_in_bounds` (null solution). The selected candidate is
feasible on the shared draws. Precision is `max(10 EUR, 0.001*candidate)`;
Monte Carlo standard error is separately reported in percentage points,
conditional on the model and draws, without a post-selection guarantee.

Optional `safe_spending` solves the fixed-real spending reference for each
model using the simulation's existing draws, even for an adaptive request.
The response explicitly identifies that policy change. Optional `annual_bands`
and `outcome_metrics` reuse the same ensemble. Full paths are never returned.

## Limits and deployment

| Resource | Limit |
|---|---|
| JSON request / encoded response | 64 KiB / 1 MiB |
| Search page / asset batch | 50 / 10 |
| Compared portfolios / holdings each | 4 / 20 |
| Distinct loaded dependencies | 40 |
| FIRE paths / horizon | 100-10000 / 1-100 years |
| Seed | integer 0 through `2^53-1` |
| FIRE work | 60 million path-steps including generation and solves |
| Per-client operation admission | 60/minute, burst 10 |
| Asset/portfolio computations / waiting queue (shared with `/view`) | 2 / 16 |
| FIRE computations (shared with the web simulator) | `max(2, GOMAXPROCS/2)`, cancellable waiting |
| API/MCP deadline | 60 seconds, including queue wait |
| Request-triggered upstream HTTP | 60/minute, burst 10, including retries/redirects |
| New foreign IDs | default 10/client/hour, 60/process/hour, shared with web |

`cache_only` forbids all network access, including FX, inflation and fee lookup,
and refuses stale cache fallbacks. Bundled snapshots remain available and are
identified by the dataset revision. `cached_or_fetch` permits bounded retrieval;
existing upstream fallback warnings remain visible. Simultaneous identical
fetches coalesce; one cancelled waiter does not cancel another, and shared work
stops when the last waiter leaves rather than at a separate flight deadline. Unknown IDs
are negative-cached for five minutes with a bounded table; outages are not.

`-serve-origins` accepts comma-separated exact browser origins. Cross-origin
access is disabled by default, there is no credentialed wildcard, and MCP
validates Origin independently of CORS. Requests without Origin are accepted;
CORS does not authenticate remote clients. Browser extensions need appropriate
host permissions.

Web, API and MCP share `clientIP`: validated `CF-Connecting-IP`, then the final
`X-Forwarded-For` hop, then the socket peer. Forwarded headers require a trusted
proxy and an origin restricted to that proxy. They are fairness hints, not
identity or authentication; the global budget remains authoritative. HTTPS
and perimeter access restrictions belong to the deployment.

Cancellation reaches API/MCP data loads, generation and simulation loops.
Anonymous MCP cancellation notifications are no-ops because an ID or shared
IP does not authorize cancelling another request. The original HTTP context
is authoritative. Existing rendered web views retain their historical compute
interfaces; their shared admission bounds concurrency but does not promise
mid-computation cancellation for every legacy view.
