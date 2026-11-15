package web

import (
	"strconv"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/fire"
)

// capeHistory renders the Shiller CAPE series since 1881 as a single-axis line
// (no dual axis: the implied return is a readout, not a second scale), with the
// historical median marked and today's value emphasised at the right end. The
// accent-coloured line darkens to amber in the terminal theme. Static, so it is
// served with the page metadata.
func capeHistory() string {
	s := fire.CAPEHistory()
	if len(s) == 0 {
		return ""
	}
	xs := make([]float64, len(s))
	ys := make([]float64, len(s))
	for i, p := range s {
		xs[i] = capeYear(p.Date)
		ys[i] = p.Value
	}
	snap := capeSnapshot()
	return darkMultiLine(
		chart.Options{Width: 760, Height: 300},
		"", "CAPE (PE10)",
		[]chart.XYSeries{{Name: "CAPE", Xs: xs, Ys: ys, Color: "#0B7285"}},
		chart.Marker{Axis: 'y', Value: snap.Median, Label: "median"},
		chart.Marker{Axis: 'x', Value: xs[len(xs)-1], Label: "today"},
	)
}

// capeYear turns a "YYYY-MM-DD" date into a fractional year for the x-axis.
func capeYear(date string) float64 {
	y, m := 0, 1
	if len(date) >= 7 {
		y, _ = strconv.Atoi(date[:4])
		m, _ = strconv.Atoi(date[5:7])
	}
	return float64(y) + (float64(m)-0.5)/12
}

// CapeSnapshot is the bundled valuation reading; see fire.CAPESnapshot.
type CapeSnapshot = fire.CAPESnapshot

func capeSnapshot() CapeSnapshot { return fire.CurrentCAPE() }
func capeAdjustedMu(sigma float64) float64 {
	mu, _ := fire.CurrentCAPE().ArithmeticMean(sigma)
	return mu
}
