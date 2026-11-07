# Web security validation, 2026-11-07

Focused review of public HTTP entry points, upstream quote fetching and client
budgets. This is not a complete penetration test or a container OS package scan.

## Finding and correction

The per-client foreign-identifier budget used the first X-Forwarded-For hop.
A visitor could rotate that prefix through an appending reverse proxy and
receive a fresh allowance. The independent global budget still bounded total
fetches. Impact: quota bypass and inaccurate access-log client addresses, not
arbitrary URL fetching or code execution.

Use a valid CF-Connecting-IP first, otherwise the final XFF hop across all
header fields, otherwise the socket peer. Normalize IP spelling. The origin
must remain behind a trusted proxy; direct clients can forge forwarding
headers. With several proxy hops, the fallback may conservatively group clients
behind the final proxy. See [Cloudflare's header semantics](https://developers.cloudflare.com/fundamentals/reference/http-headers/).

## Evidence

- `TestForeignBudgetIgnoresForgedForwardedPrefix` failed before the correction
  with and without CF-Connecting-IP. It now verifies that rotating a prefix
  cannot reset the per-client budget.
- Address tests cover edge precedence, repeated headers, invalid/empty final
  hops, IPv6 normalization and socket fallback.
- `make check` passes: format, vet, staticcheck, package tests and example builds.
- `govulncheck` on Linux/amd64 with Go 1.26.10, the Docker build toolchain,
  reports no known vulnerabilities for the root, cmd and pkg packages.
  Gitignored scratch programs are excluded. The module uses only the stdlib.

The reviewed example-file reads require membership in the embedded example
catalog. Foreign identifiers pass the existing syntax gate and exact resolution;
HTTP requests cannot provide arbitrary filesystem paths or quote-source URLs.
Existing body, calculation and concurrency bounds remain in place.
