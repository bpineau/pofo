package permanent

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// AssetReturns holds aligned monthly REAL returns for the four sleeves. All
// slices share Dates (ascending, month starts); 0.01 = +1% real that month.
type AssetReturns struct {
	Dates                     []time.Time
	Equity, Bonds, Cash, Gold []float64
}

func (ar AssetReturns) len() int { return len(ar.Dates) }

// Result is the outcome of a backtest: the tactical and the static (25/25/25/25)
// portfolio real return series, and the tactical target weights, all aligned to
// Dates.
type Result struct {
	Dates    []time.Time
	Tactical []float64
	Static   []float64
	Weights  []Allocation
}

// Simulate backtests the tactical allocator against real asset returns. Each
// month's return is earned with the target weights of the most recent regime
// dated strictly before that month (no lookahead); months before any regime, or
// with misaligned inputs, are skipped. regimes need not be contiguous with the
// returns; they are matched by date.
func Simulate(regimes []Regime, ar AssetReturns, pm Params) (Result, error) {
	if ar.len() == 0 {
		return Result{}, fmt.Errorf("permanent: empty asset returns")
	}
	for name, s := range map[string][]float64{"equity": ar.Equity, "bonds": ar.Bonds, "cash": ar.Cash, "gold": ar.Gold} {
		if len(s) != ar.len() {
			return Result{}, fmt.Errorf("permanent: %s has %d returns, want %d", name, len(s), ar.len())
		}
	}
	// regimes sorted ascending by date for the "most recent before" lookup.
	rs := append([]Regime(nil), regimes...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].Date.Before(rs[j].Date) })

	var res Result
	for i, d := range ar.Dates {
		r, ok := latestBefore(rs, d)
		if !ok {
			continue
		}
		w := r.Allocate(pm)
		tac := w.Equity*ar.Equity[i] + w.Bonds*ar.Bonds[i] + w.Cash*ar.Cash[i] + w.Gold*ar.Gold[i]
		stat := 0.25 * (ar.Equity[i] + ar.Bonds[i] + ar.Cash[i] + ar.Gold[i])
		res.Dates = append(res.Dates, d)
		res.Tactical = append(res.Tactical, tac)
		res.Static = append(res.Static, stat)
		res.Weights = append(res.Weights, w)
	}
	if len(res.Dates) == 0 {
		return Result{}, fmt.Errorf("permanent: no month had a prior regime")
	}
	return res, nil
}

// latestBefore returns the last regime dated strictly before d.
func latestBefore(sorted []Regime, d time.Time) (Regime, bool) {
	// sorted ascending; find rightmost with Date < d.
	i := sort.Search(len(sorted), func(k int) bool { return !sorted[k].Date.Before(d) })
	if i == 0 {
		return Regime{}, false
	}
	return sorted[i-1], true
}

// Stats summarizes a monthly real-return series: annualized geometric return and
// volatility, worst peak-to-trough drawdown, the fraction of months spent below
// the prior peak, and the longest underwater stretch in months. Annualization
// uses 12 periods, matching the monthly-real convention of this package.
type Stats struct {
	CAGR, Vol, MaxDrawdown float64
	UnderwaterFraction     float64
	LongestUnderwater      int
	Months                 int
}

// Compute returns the Stats of a monthly real-return series.
func Compute(returns []float64) Stats {
	var s Stats
	s.Months = len(returns)
	if s.Months == 0 {
		return s
	}
	var logSum, mean float64
	for _, r := range returns {
		logSum += math.Log(1 + r)
		mean += r
	}
	n := float64(s.Months)
	mean /= n
	s.CAGR = math.Exp(logSum*12/n) - 1
	var variance float64
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	s.Vol = math.Sqrt(variance/n) * math.Sqrt(12)

	idx, peak := 1.0, 1.0
	var underwater, run int
	for _, r := range returns {
		idx *= 1 + r
		if idx >= peak {
			peak = idx
			run = 0
		} else {
			underwater++
			run++
			if run > s.LongestUnderwater {
				s.LongestUnderwater = run
			}
		}
		if dd := idx/peak - 1; dd < s.MaxDrawdown {
			s.MaxDrawdown = dd
		}
	}
	s.UnderwaterFraction = float64(underwater) / n
	return s
}
