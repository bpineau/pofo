package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// TrendAnchor names a bundled monthly reference a trend reconstruction can
// take its month-to-month path from, and says how to read its returns. Funded
// says the reference already earns cash on its collateral (a fund's total
// return), so its excess over cash is what gets rescaled to a volatility
// target: leverage multiplies a trading book, never a cash yield.
type TrendAnchor struct {
	ID     string
	Funded bool
}

// The two references, which answer different questions and are not
// interchangeable.
//
//   - GrossTrendAnchor is the academic time-series-momentum factor (see
//     cmd/gen-trend-refdata): a pure-trend rule with no fee, no slippage and no
//     capacity limit, published as an excess over cash. It is the right path
//     for an overlay that replicates a pure-trend INDEX, which is measured the
//     same way, and its level has to be pinned down afterwards (pinTrendIR)
//     because no investable programme keeps what it earns.
//   - NetTrendAnchor is a monthly composite of REAL managed-futures
//     programmes, each already net of its own manager's fees (see
//     cmd/gen-trendnet-refdata). It is the right path for a reconstruction of a
//     diversified managed-futures FUND, and it needs no pin at all: an
//     investable level is what it already is. It is a funded total return, and
//     mistaking it for an excess index hands the deep tail a second cash leg,
//     worth six points a year in the 1990s.
var (
	GrossTrendAnchor = TrendAnchor{ID: "TREND-TSMOM-USD"}
	NetTrendAnchor   = TrendAnchor{ID: "TREND-NET-USD", Funded: true}
)

// AnchorTrend rewrites a trend reconstruction so that its return over every
// anchored month equals the reference's, rescaled to the book's volatility
// target, while the reconstruction keeps supplying the day-to-day texture
// inside each month. See the two anchors above for which one a caller wants
// and why the choice is not free.
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
// Everything here is arithmetic on EXCESS returns: a funded reference is
// stripped of its cash leg before it is rescaled, and earnCash then says
// whether the OUTPUT is a funded total return (a fund: the collateral earns
// cash) or a pure overlay excess (the trend sleeve of a stacked fund),
// matching TSMOMConfig.EarnCash. cash is the daily cash return aligned to
// dates, and is required whenever either side is funded.
//
// Months the reference does not cover keep the reconstruction's own returns,
// so a series that starts before the reference does is not truncated.
func AnchorTrend(f Fetcher, ref TrendAnchor, dates []time.Time, values, cash []float64, targetVol float64, earnCash bool) ([]float64, error) {
	if len(dates) != len(values) || len(dates) != len(cash) {
		return nil, fmt.Errorf("anchor: %d dates, %d values, %d cash", len(dates), len(values), len(cash))
	}
	anchorID := ref.ID
	index, err := f.Fetch(anchorID, dates[0].AddDate(-2, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("anchor %s: %w", anchorID, err)
	}
	if index == nil || len(index.Points) < 60 {
		return nil, fmt.Errorf("anchor %s: too short", anchorID)
	}
	// A cash leg carries no volatility worth the name (a T-bill moves by
	// hundredths of a point a month where these indices move by whole ones),
	// so the funded reference's own volatility is its excess volatility.
	scale := targetVol / monthlyVol(index)
	if !(scale > 0) || math.IsInf(scale, 0) {
		return nil, fmt.Errorf("anchor %s: unusable volatility", anchorID)
	}

	// Anchor month m onto the last series date of that month, keeping the
	// reference's raw monthly return: what it has to become depends on the
	// month's own cash accrual, which is read off the output's calendar below.
	lastOfMonth := make(map[string]int, len(dates)/20)
	for i, d := range dates {
		lastOfMonth[d.Format("2006-01")] = i // dates ascending: keeps the month's last
	}
	type anchor struct {
		idx int
		ret float64
	}
	var anchors []anchor
	for i := 1; i < len(index.Points); i++ {
		prev, cur := index.Points[i-1], index.Points[i]
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
		anchors = append(anchors, anchor{idx, cur.Close/prev.Close - 1})
	}
	if len(anchors) < 24 {
		return nil, fmt.Errorf("anchor %s: only %d usable months", anchorID, len(anchors))
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
		// What the month must return: the reference's EXCESS, rescaled to the
		// volatility target, then funded again if the output is a fund. The
		// cash factor is the same one on both sides, this month's own.
		c := 1.0
		if ref.Funded || earnCash {
			for i := a.idx + 1; i <= b.idx; i++ {
				c *= 1 + cash[i]
			}
		}
		want := b.ret
		if ref.Funded {
			want = (1+want)/c - 1
		}
		want *= scale
		if earnCash {
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
