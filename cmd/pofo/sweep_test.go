package main

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// A statistic that does not exist over the window prints as "-", never as a
// zero the reader would take for a measurement.
func TestSweepPct(t *testing.T) {
	if got := pct(math.NaN()); got != "-" {
		t.Errorf("pct(NaN) = %q, want %q", got, "-")
	}
	if got := pct(0.0725); got != "7.25 %" {
		t.Errorf("pct(0.0725) = %q", got)
	}
	if got := pct(-0.1); got != "-10.00 %" {
		t.Errorf("pct(-0.1) = %q", got)
	}
}

// The recovery column: years, a trailing "+" while the drawdown is still
// open at the end of the window, "-" when there was nothing to recover from.
func TestSweepTTR(t *testing.T) {
	for _, tc := range []struct {
		stats metrics.Stats
		want  string
	}{
		{metrics.Stats{}, "-"},
		{metrics.Stats{TTRDays: 731}, "2.0 y"},
		{metrics.Stats{TTRDays: 731, TTROngoing: true}, "2.0 y+"},
	} {
		if got := ttr(tc.stats); got != tc.want {
			t.Errorf("ttr(%+v) = %q, want %q", tc.stats, got, tc.want)
		}
	}
}

// -sweep checks its own arguments before fetching a single quote: a nil
// client would panic if it did not.
func TestRunSweepGuards(t *testing.T) {
	opt := &options{currency: "EUR", rebalance: 90}
	if err := runSweep(t.Context(), nil, nil, opt, 5); err == nil {
		t.Error("-sweep accepted an empty run")
	}
	specs := []*portfolio.Spec{portfolio.Single("VTI")}
	for _, step := range []float64{0, -1, 50.1, 100} {
		if err := runSweep(t.Context(), nil, specs, opt, step); err == nil {
			t.Errorf("-sweep-step %g was accepted", step)
		}
	}
}
