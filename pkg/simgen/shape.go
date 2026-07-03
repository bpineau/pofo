package simgen

import (
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// anchorShape blends a coarse total-return index (anchors, e.g. monthly
// net-TR levels) with a finer price series of the same market (shape, e.g.
// the daily price index): the result carries one point per shape date,
// follows the shape's day-to-day moves, and passes exactly through every
// anchor level. The residual between consecutive anchors (dividends,
// methodology drift) is compounded evenly across the shape steps in
// between, a negligible per-day adjustment next to daily volatility, so
// the output keeps the anchors' levels AND the shape's realized daily
// variance. Anchors the shape does not reach are dropped (the caller
// typically splices them back behind with marketdata.ExtendBack); with no
// usable overlap the anchors are returned unchanged. Both inputs must be
// ascending with positive closes.
// shapedSeries returns the anchors series with the daily shape blended in
// where the shape covers it (anchorShape) and the untouched anchors kept on
// both sides: spliced back in front (ExtendBack) and appended after the
// shape's last date. A shape may legitimately stop decades before the
// anchors' end (e.g. the daily Treasury yields only matter before the real
// fund's inception, which replaces the proxy from there anyway); the
// remaining anchors then keep their own cadence rather than being dropped
// or, worse, vetoing the whole blend. A missing or non-overlapping shape
// leaves the anchors unchanged.
func shapedSeries(anchors, shape *marketdata.Series) *marketdata.Series {
	if shape == nil || len(shape.Points) == 0 || len(anchors.Points) == 0 {
		return anchors
	}
	out := *anchors
	out.Points = anchorShape(anchors.Points, shape.Points)
	// Anchors past the shape's coverage: anchorShape ends exactly on an
	// anchor level, so later anchors continue the same index seamlessly.
	last := out.Points[len(out.Points)-1].Date
	for _, a := range anchors.Points {
		if a.Date.After(last) {
			out.Points = append(out.Points, a)
		}
	}
	out.SimulatedBefore = time.Time{} // allow the pre-shape months back in front
	marketdata.ExtendBack(&out, anchors)
	return &out
}

func anchorShape(anchors, shape []marketdata.Point) []marketdata.Point {
	// Boundary of anchor j: the first shape point on or after its date.
	// Several anchors falling before one shape point collapse to the
	// latest of them (the earlier ones predate the shape's coverage).
	type bound struct{ anchor, shape int }
	var bounds []bound
	i := 0
	for j, a := range anchors {
		for i < len(shape) && shape[i].Date.Before(a.Date) {
			i++
		}
		if i >= len(shape) {
			break
		}
		if n := len(bounds); n > 0 && bounds[n-1].shape == i {
			bounds[n-1].anchor = j
			continue
		}
		bounds = append(bounds, bound{j, i})
	}
	if len(bounds) < 2 {
		return anchors
	}
	out := make([]marketdata.Point, 0, bounds[len(bounds)-1].shape-bounds[0].shape+1)
	for k := 0; k+1 < len(bounds); k++ {
		a, b := bounds[k], bounds[k+1]
		la, lb := anchors[a.anchor].Close, anchors[b.anchor].Close
		sa, sb := shape[a.shape].Close, shape[b.shape].Close
		if la <= 0 || lb <= 0 || sa <= 0 || sb <= 0 {
			return anchors
		}
		// Per-step residual factor: hitting lb at b.shape exactly.
		r := math.Pow((lb/la)/(sb/sa), 1/float64(b.shape-a.shape))
		v := la
		out = append(out, marketdata.Point{Date: shape[a.shape].Date, Close: v})
		for i := a.shape + 1; i < b.shape; i++ {
			v *= shape[i].Close / shape[i-1].Close * r
			out = append(out, marketdata.Point{Date: shape[i].Date, Close: v})
		}
	}
	last := bounds[len(bounds)-1]
	out = append(out, marketdata.Point{Date: shape[last.shape].Date, Close: anchors[last.anchor].Close})
	return out
}
