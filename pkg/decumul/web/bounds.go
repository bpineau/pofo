package web

import (
	"fmt"
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
// The horizon is the exception: it is floored at one year rather than zero,
// because a zero-year plan is not a short retirement but an empty one, and
// several views index the last plan year unconditionally (a POST of "{}" used
// to take the whole process down on that index).
func (pr Params) bounded() Params {
	pr.NPaths = clamp(pr.NPaths, 0, maxPaths)
	pr.Years = clamp(pr.Years, 1, maxYears)
	pr.PensionYear = clamp(pr.PensionYear, 0, maxYears)
	pr.SideUntilYear = clamp(pr.SideUntilYear, 0, maxYears)
	pr.BufferStopYear = clamp(pr.BufferStopYear, 0, maxYears)
	pr.Age = clamp(pr.Age, 0, 110)
	// The tax book: a fraction is a fraction, and each envelope is a pocket
	// carved out of the growth sleeve, so neither can exceed it. A GainFrac
	// above 1 would price a cost basis below zero, i.e. a tax on capital.
	pr.GainFrac = clamp(pr.GainFrac, 0, 1)
	g := pr.growthSleeve()
	pr.PEACapital = clamp(pr.PEACapital, 0, g)
	pr.AVCapital = clamp(pr.AVCapital, 0, g)
	return pr
}

// validate rejects a request the page itself cannot express: an envelope book
// whose pockets add up to more than the sleeve they are carved from. Clamping
// that one silently would simulate a different household (the pockets would
// be pro-rated and the taxable one would vanish), so the caller is told
// instead. Called on the bounded params, before any slot is taken.
func (pr Params) validate() error {
	if g := pr.growthSleeve(); pr.PEACapital+pr.AVCapital > g+0.5 {
		return fmt.Errorf("envelope amounts add up to more than the invested capital: "+
			"PEA %.0f + assurance-vie %.0f > %.0f (capital %.0f minus %.0f of cash buffer)",
			pr.PEACapital, pr.AVCapital, g, pr.Capital, pr.Capital-g)
	}
	return nil
}

func clamp[T int | float64](v, lo, hi T) T {
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
