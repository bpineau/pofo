package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// TrendAnchorID is the bundled monthly reference the managed-futures
// reconstructions take their month-to-month path from (see
// cmd/gen-trend-refdata for what it is and where it comes from).
const TrendAnchorID = "TREND-TSMOM-USD"

// AnchorTrend rewrites a trend reconstruction so that its return over every
// anchored month equals the reference's, rescaled to the book's volatility
// target, while the reconstruction keeps supplying the day-to-day texture
// inside each month.
//
// The engine is a seven-market book: it moves like a trend programme, but its
// month-to-month path only agrees with the funds it replicates at a
// correlation near 0.55, because seven markets cannot reproduce what a
// programme trading fifty does. The reference agrees at 0.60 to 0.71 on the
// same funds. Anchoring therefore buys the one thing the reconstruction could
// not manufacture, the real sequence of good and bad trend months, and gives
// up nothing: the daily variance, the drawdown shape inside a month and the
// crisis timing all still come from the engine.
//
// The reference is an EXCESS-of-cash index. earnCash says whether the output
// is a funded total return (a fund: the collateral earns cash) or a pure
// overlay excess (the trend sleeve of a stacked fund), matching
// TSMOMConfig.EarnCash. cash is the daily cash return aligned to dates.
//
// Months the reference does not cover keep the reconstruction's own returns,
// so a series that starts before the reference does is not truncated.
func AnchorTrend(f Fetcher, dates []time.Time, values, cash []float64, targetVol float64, earnCash bool) ([]float64, error) {
	if len(dates) != len(values) || len(dates) != len(cash) {
		return nil, fmt.Errorf("anchor: %d dates, %d values, %d cash", len(dates), len(values), len(cash))
	}
	ref, err := f.Fetch(TrendAnchorID, dates[0].AddDate(-2, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("anchor %s: %w", TrendAnchorID, err)
	}
	if ref == nil || len(ref.Points) < 60 {
		return nil, fmt.Errorf("anchor %s: too short", TrendAnchorID)
	}
	scale := targetVol / monthlyVol(ref)
	if !(scale > 0) || math.IsInf(scale, 0) {
		return nil, fmt.Errorf("anchor %s: unusable volatility", TrendAnchorID)
	}

	// Anchor month m onto the last series date of that month, and turn the
	// reference's excess return into the return the output must realize.
	lastOfMonth := make(map[string]int, len(dates)/20)
	for i, d := range dates {
		lastOfMonth[d.Format("2006-01")] = i // dates ascending: keeps the month's last
	}
	type anchor struct {
		idx int
		ret float64
	}
	var anchors []anchor
	for i := 1; i < len(ref.Points); i++ {
		prev, cur := ref.Points[i-1], ref.Points[i]
		if prev.Close <= 0 || cur.Close <= 0 {
			continue
		}
		idx, ok := lastOfMonth[cur.Date.Format("2006-01")]
		if !ok {
			continue
		}
		if n := len(anchors); n > 0 && anchors[n-1].idx >= idx {
			continue // the series has no room for that month
		}
		anchors = append(anchors, anchor{idx, (cur.Close/prev.Close - 1) * scale})
	}
	if len(anchors) < 24 {
		return nil, fmt.Errorf("anchor %s: only %d usable months", TrendAnchorID, len(anchors))
	}

	// Daily returns of the reconstruction, kept outside the anchored span.
	rets := make([]float64, len(values))
	for i := 1; i < len(values); i++ {
		if values[i-1] > 0 {
			rets[i] = values[i]/values[i-1] - 1
		}
	}
	for k := 0; k+1 < len(anchors); k++ {
		a, b := anchors[k], anchors[k+1]
		// What the month must return, cash included when the book is funded.
		want := b.ret
		if earnCash {
			c := 1.0
			for i := a.idx + 1; i <= b.idx; i++ {
				c *= 1 + cash[i]
			}
			want = (1+want)*c - 1
		}
		got := 1.0
		for i := a.idx + 1; i <= b.idx; i++ {
			got *= 1 + rets[i]
		}
		if got <= 0 || 1+want <= 0 {
			continue
		}
		// Spread the correction evenly over the month's steps: the texture
		// (which day moved, and by how much relative to its neighbours) is
		// untouched, only the month's total is pinned.
		adj := math.Pow((1+want)/got, 1/float64(b.idx-a.idx))
		for i := a.idx + 1; i <= b.idx; i++ {
			rets[i] = (1+rets[i])*adj - 1
		}
	}

	out := make([]float64, len(values))
	out[0] = values[0]
	for i := 1; i < len(out); i++ {
		out[i] = out[i-1] * (1 + rets[i])
	}
	return out, nil
}

// monthlyVol is the annualized volatility of a monthly index's returns.
func monthlyVol(s *marketdata.Series) float64 {
	var rets []float64
	for i := 1; i < len(s.Points); i++ {
		if s.Points[i-1].Close > 0 {
			rets = append(rets, s.Points[i].Close/s.Points[i-1].Close-1)
		}
	}
	if len(rets) < 24 {
		return 0
	}
	var m float64
	for _, r := range rets {
		m += r
	}
	m /= float64(len(rets))
	var v float64
	for _, r := range rets {
		v += (r - m) * (r - m)
	}
	return math.Sqrt(v/float64(len(rets)-1)) * math.Sqrt(12)
}
