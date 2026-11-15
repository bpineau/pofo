// Package remote exposes five compact, read-only numerical operations through
// one service and optional HTTP and MCP adapters, using only the Go standard
// library. No user data is persisted and no local-file or URL input is accepted.
//
// # Start here
//
// New accepts an analyze.Source, usually a marketdata.Client. Execute accepts
// strict JSON and returns Result or *Error. Handler mounts independently enabled
// HTTP API and MCP routes. Tools and OpenAPI share input schemas. Consult
// docs/usage/library/remote.md for examples, limits and deployment assumptions.
//
// # Units and defaults
//
// Fields ending in _pct use percent (5 means five percent); FIRE money is real
// EUR at plan start. Analysis is nominal in the requested currency. Defaults
// evolve with the engine and every applied default is returned in Meta.Defaulted.
// Number preserves undefined statistics as JSON null. Identity's missing TER is
// null, not a zero charge. PlanInput pointers distinguish zero from omission.
//
// # Data and computation
//
// The service uses analyze for asset/portfolio studies and fire/decumul for
// annual retirement models. Portfolio studies share their holding window;
// benchmark overlap is separate. FIRE portfolio mode uses fixed monthly weights,
// EUR prices and French HICP, not the analysis rebalancing strategy. Exact-only
// identifiers, bounded payloads and path-step budgets are enforced before work.
// A caller-owned Source must honor cancellation and cache-only contexts. The
// server supplies shared request admission and foreign-ID quotas via HTTPOptions
// and Service.BeforeFetch; the mathematical packages do not own client quotas.
//
// # Protocol
//
// MCP implements stateless 2025-11-25 Streamable HTTP with ordinary JSON responses,
// a single JSON text block per tool result, and no outputSchema. Anonymous
// cancellation notifications cannot identify another caller and are no-ops;
// original request cancellation propagates through data fetches and simulation.
package remote
