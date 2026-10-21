package web

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
)

// Flow is one row of the rail's dated-flows list, the long form of the two
// income sliders: a yearly income or expense over a span of years (a
// decumul.Cashflow, with its own indexation drift), or a one-off capital event
// (a decumul.Lump: an inheritance or a house sale in, a purchase out).
// Amounts are real euros, positive in and negative out.
type Flow struct {
	Label  string  `json:"label"`  // the reader's own name for it; display only
	Once   bool    `json:"once"`   // a one-off capital event rather than a yearly flow
	Amount float64 `json:"amount"` // euros per year, or once; positive in, negative out
	From   int     `json:"from"`   // plan year it starts, or happens when Once
	To     int     `json:"to"`     // exclusive end of a yearly flow; 0 = to the horizon
	Drift  float64 `json:"drift"`  // real growth per year of a yearly flow, a FRACTION
}

// The list's bounds. The public page may be served to anonymous visitors, and
// each row is scanned every simulated year of every path, so the list is
// short by construction: a dozen rows covers any household's dated events,
// and the amounts stop where no plan this page models could reach (a yearly
// flow of a million, a one-off of ten). maxFlows is refused rather than
// clamped (dropping rows would simulate another household); the amounts, the
// years and the drift are clamped like every other field.
const (
	maxFlows      = 12
	maxFlowYearly = 1_000_000
	maxFlowOnce   = 10_000_000
	maxFlowLabel  = 40
)

// boundedFlows clamps every row into its bounds, leaving the row count to
// validate.
func boundedFlows(fs []Flow) []Flow {
	if len(fs) == 0 {
		return nil
	}
	out := make([]Flow, len(fs))
	for i, f := range fs {
		lim := float64(maxFlowYearly)
		if f.Once {
			lim = maxFlowOnce
		}
		if math.IsNaN(f.Amount) {
			f.Amount = 0
		}
		f.Amount = clamp(f.Amount, -lim, lim)
		f.From = clamp(f.From, 0, maxYears)
		f.To = clamp(f.To, 0, maxYears)
		if math.IsNaN(f.Drift) {
			f.Drift = 0
		}
		f.Drift = clamp(f.Drift, -0.05, 0.05)
		if r := []rune(f.Label); len(r) > maxFlowLabel {
			f.Label = string(r[:maxFlowLabel])
		}
		out[i] = f
	}
	return out
}

// validateFlows refuses a list longer than the page can post.
func validateFlows(fs []Flow) error {
	if len(fs) > maxFlows {
		return fmt.Errorf("at most %d dated flows, got %d", maxFlows, len(fs))
	}
	return nil
}

// flowCashflows is the yearly rows as the kernel's cashflows.
func (pr Params) flowCashflows() []decumul.Cashflow {
	var out []decumul.Cashflow
	for _, f := range pr.Flows {
		if !f.Once && f.Amount != 0 {
			out = append(out, decumul.Cashflow{FromYear: f.From, ToYear: f.To, Annual: f.Amount, Drift: f.Drift})
		}
	}
	return out
}

// lumps is the one-off rows as the kernel's lumps.
func (pr Params) lumps() []decumul.Lump {
	var out []decumul.Lump
	for _, f := range pr.Flows {
		if f.Once && f.Amount != 0 {
			out = append(out, decumul.Lump{Year: f.From, Amount: f.Amount})
		}
	}
	return out
}

// lumpMarkers marks each one-off event on a year axis with its signed amount,
// so a fan's kink has a cause written above it.
func (pr Params) lumpMarkers(years int) []chart.Marker {
	var out []chart.Marker
	for _, f := range pr.Flows {
		if f.Once && f.Amount != 0 && f.From > 0 && f.From < years {
			out = append(out, chart.Marker{Axis: 'x', Value: float64(f.From), Label: fmtSigned(f.Amount)})
		}
	}
	return out
}

// fmtSigned renders a signed euro amount compactly: +250k, -1.2M.
func fmtSigned(v float64) string {
	sign := "+"
	if v < 0 {
		sign, v = "-", -v
	}
	if v >= 1e6 {
		return fmt.Sprintf("%s%.1fM", sign, v/1e6)
	}
	return fmt.Sprintf("%s%.0fk", sign, v/1e3)
}
