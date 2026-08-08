package firebook

import (
	"os"
	"testing"
)

// driftEnv is the switch that turns the figure-drift checks on. "make
// figure-drift" sets it; nothing else does.
const driftEnv = "POFO_FIGURE_DRIFT"

// frozenAgainstData marks a test whose expectations are numbers read off the
// bundled datasets and frozen into a plate.
//
// Those numbers are a snapshot of the data, and the data moves: Shiller revises
// a CAPE, a CPI month lands and opens one more full decade, a reference series
// is rebuilt. None of that is a defect, and none of it should stand between a
// data refresh and a green tree: "make refresh" must never require a code
// change, because the book is allowed to lag the data it was drawn from.
//
// So these checks live outside "make check". "make figure-drift" runs them, and
// what it prints is the worklist for bringing a plate back in step with its
// data, deliberately, when someone decides to. The plates' own rendering,
// wording and house-rule tests are unaffected and keep running everywhere: only
// the recomputation from bundled data is gated here.
func frozenAgainstData(t *testing.T) {
	t.Helper()
	if os.Getenv(driftEnv) == "" {
		t.Skip("frozen against the bundled data; `make figure-drift` reports what the plates owe it")
	}
}
