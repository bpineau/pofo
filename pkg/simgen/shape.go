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

// alignMonthEnd re-dates each monthly anchor point onto the last shape date in
// the same calendar month. The anchor levels are month-END values (see the
// SP500-USD / MSCIWORLD-USD headers); anchorShape pins each to the first shape
// date ON OR AFTER the anchor's date, so an anchor dated on a calendar month-end
// that falls on a weekend or holiday would pin to the next month's first trading
// day and slip the whole reconstruction by a few days (up to ~1.5 %/yr on a
// volatile December, the trap that first showed up rebuilding a EUR MSCI World).
// Snapping the date to the shape's own last trading day of the month removes
// that slip and is a no-op for an anchor already dated on a trading day (e.g.
// SP500-USD, dated from ^GSPC). Anchor months the shape does not reach keep
// their date; month order is preserved, so the result stays ascending.
func alignMonthEnd(anchor, shape *marketdata.Series) *marketdata.Series {
	if shape == nil || len(shape.Points) == 0 {
		return anchor
	}
	last := make(map[string]time.Time, len(shape.Points)/20)
	for _, p := range shape.Points {
		last[p.Date.Format("2006-01")] = p.Date // shape ascending: keeps the month's last date
	}
	out := *anchor
	out.Points = make([]marketdata.Point, len(anchor.Points))
	for i, p := range anchor.Points {
		d := p.Date
		if sd, ok := last[p.Date.Format("2006-01")]; ok {
			d = sd
		}
		out.Points[i] = marketdata.Point{Date: d, Close: p.Close}
	}
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

// despike drops single-day bad prints from a daily shape series. The
// signature is strict, because real crashes must pass untouched: a point
// whose arrival and departure returns are BOTH enormous against the local
// volatility (>6 sigma over the surrounding ~5 weeks, the spike itself
// excluded), OPPOSITE in sign, and mostly cancel (the two-day round trip
// nets to less than a third of either leg) is a provider glitch, not a
// market day. 1987-10-19 keeps its -23% (the legs do not cancel) and
// 2020-03-12/13 its -10%/+9% (against a ~5% local sigma neither leg
// clears the bar); what goes is the lone fabricated print, like Yahoo's
// MSCI World price index quoting 1180.3 on 2001-04-03 between 1055 and
// 1025 (+11.9% then -13.2% inside a 1.4%-sigma neighbourhood), which a
// month-anchored reconstruction would otherwise bake into every world
// equity backcast. Dropping the point keeps the two-day net move via the
// direct ratio of its neighbours.
func despike(points []marketdata.Point) []marketdata.Point {
	const (
		zBar   = 6.0 // each leg must exceed zBar times the local sigma
		window = 25  // returns on each side defining "local"
	)
	if len(points) < 10 {
		return points
	}
	ret := func(i int) float64 { return points[i].Close/points[i-1].Close - 1 }
	out := make([]marketdata.Point, 0, len(points))
	out = append(out, points[0])
	for i := 1; i < len(points); i++ {
		if i+1 >= len(points) {
			out = append(out, points[i])
			continue
		}
		r1, r2 := ret(i), ret(i+1)
		if r1*r2 >= 0 {
			out = append(out, points[i])
			continue
		}
		net := (1+r1)*(1+r2) - 1
		if math.Abs(net) >= math.Min(math.Abs(r1), math.Abs(r2))/3 {
			out = append(out, points[i])
			continue
		}
		// Local sigma, the suspect pair excluded.
		var sum, sumsq float64
		n := 0
		for j := max(1, i-window); j <= min(len(points)-1, i+window); j++ {
			if j == i || j == i+1 {
				continue
			}
			r := ret(j)
			sum += r
			sumsq += r * r
			n++
		}
		if n < 10 {
			out = append(out, points[i])
			continue
		}
		mean := sum / float64(n)
		sigma := math.Sqrt(sumsq/float64(n) - mean*mean)
		if sigma <= 0 || math.Abs(r1) <= zBar*sigma || math.Abs(r2) <= zBar*sigma {
			out = append(out, points[i])
			continue
		}
		// The point is a glitch: skip it (its neighbours' direct ratio
		// carries the true two-day move).
	}
	return out
}
