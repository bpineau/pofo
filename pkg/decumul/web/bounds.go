package web

import (
	"net/http"
	"runtime"
)

// Request bounds. The simulation endpoints are driven by whatever JSON a
// client posts, and the public deployment shares a small machine with other
// services, so every dimension a request can inflate is capped here: a
// single "nPaths": 1e7 would otherwise allocate tens of gigabytes and pin
// every core for minutes, which is a denial of service of the whole host,
// not just of this page.
//
// The caps are generous for a real user: maxPaths is the top of the page's
// own slider, maxYears exceeds any retirement horizon (a plan runs to death
// at most, and the mortality tables stop at 110), and maxBodyBytes is a
// thousand times a real Params payload (a few hundred bytes).
const (
	maxPaths     = 10_000
	maxYears     = 100
	maxBodyBytes = 64 << 10
)

// simParallel bounds how many simulation requests compute at once. Each of
// them already fans out over simWorkers goroutines (one per core), so beyond
// a couple of concurrent requests the machine is saturated and extra ones
// only add memory. The page fires its dozen endpoints per render, so this is
// a queue, not a rejection: a request waits for a slot until its client
// gives up (the request context), and a caller that has gone away is turned
// down with 503 rather than computed for nobody.
var simParallel = max(2, runtime.GOMAXPROCS(0)/2)

// bounded returns pr with every size-like field clamped into its bound.
// Negative counts fall back to zero, which the endpoints read as "default".
func (pr Params) bounded() Params {
	pr.NPaths = clampInt(pr.NPaths, 0, maxPaths)
	pr.Years = clampInt(pr.Years, 0, maxYears)
	pr.PensionYear = clampInt(pr.PensionYear, 0, maxYears)
	pr.SideUntilYear = clampInt(pr.SideUntilYear, 0, maxYears)
	pr.BufferStopYear = clampInt(pr.BufferStopYear, 0, maxYears)
	pr.Age = clampInt(pr.Age, 0, 110)
	return pr
}

func clampInt(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// simGate serialises the heavy endpoints behind simParallel slots. It is a
// small type rather than a bare channel so the handler reads as intent.
type simGate chan struct{}

func newSimGate(n int) simGate { return make(simGate, n) }

// acquire takes a slot, or reports false when the request was abandoned
// while waiting. The caller must release on success.
func (g simGate) acquire(r *http.Request) bool {
	select {
	case g <- struct{}{}:
		return true
	case <-r.Context().Done():
		return false
	}
}

func (g simGate) release() { <-g }
