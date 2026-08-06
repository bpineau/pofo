package simgen

import (
	"fmt"
	"math"
)

// TSMOMConfig drives the classic time-series-momentum reconstruction used
// for managed-futures funds (DBMF, KMLM, CTA…): each market is held long or
// short according to the sign of its trailing 12-month excess return, sized
// to an equal share of a portfolio volatility target.
//
// This is the standard academic replication (Moskowitz–Ooi–Pedersen); real
// funds add carry, faster signals and execution details, so expect daily
// correlations around 0.5–0.7, not 0.9.
type TSMOMConfig struct {
	Markets     []string // component ids, traded as excess returns vs cash
	CashID      string   // e.g. "^IRX"
	Lookback    int      // signal window in trading days (e.g. 252)
	VolWindow   int      // rolling volatility window (e.g. 63)
	Rebalance   int      // recompute positions every N trading days (e.g. 21)
	TargetVol   float64  // annualized portfolio volatility target (e.g. 0.10)
	MaxLeverage float64  // cap on a single market's position (e.g. 2)
	AnnualFee   float64  // e.g. 0.0085
	EarnCash    bool     // collateral earns the cash rate
}

// TSMOM builds the strategy index (base 100) on the frame. The first
// Lookback+1 dates are consumed by the signal warm-up.
func TSMOM(fr *Frame, cfg TSMOMConfig) ([]float64, int, error) {
	cash, ok := fr.Returns[cfg.CashID]
	if !ok {
		return nil, 0, fmt.Errorf("cash %s missing from frame", cfg.CashID)
	}
	n := len(fr.Dates)
	if n <= cfg.Lookback+2 {
		return nil, 0, fmt.Errorf("history too short for the lookback (%d dates)", n)
	}
	// Excess returns per market.
	excess := make([][]float64, len(cfg.Markets))
	for i, id := range cfg.Markets {
		r, ok := fr.Returns[id]
		if !ok {
			return nil, 0, fmt.Errorf("market %s missing from frame", id)
		}
		ex := make([]float64, n)
		for k := 1; k < n; k++ {
			ex[k] = r[k] - cash[k]
		}
		excess[i] = ex
	}

	start := cfg.Lookback + 1
	positions := make([]float64, len(cfg.Markets))
	values := make([]float64, n-start)
	values[0] = 100
	feeDaily := cfg.AnnualFee / 252
	sinceRebalance := cfg.Rebalance // force sizing on the first step

	for k := start + 1; k < n; k++ {
		if sinceRebalance >= cfg.Rebalance {
			sinceRebalance = 0
			for i := range cfg.Markets {
				// Sign of the trailing total excess return.
				cum := 1.0
				for j := k - cfg.Lookback; j < k; j++ {
					cum *= 1 + excess[i][j]
				}
				sign := 1.0
				if cum < 1 {
					sign = -1
				}
				// Inverse-volatility sizing toward the per-market share of
				// the portfolio volatility target (naive risk parity).
				vol := rollingVol(excess[i], k, cfg.VolWindow)
				perMarket := cfg.TargetVol / math.Sqrt(float64(len(cfg.Markets)))
				lev := 0.0
				if vol > 0 {
					lev = math.Min(perMarket/vol, cfg.MaxLeverage)
				}
				positions[i] = sign * lev
			}
		}
		sinceRebalance++

		r := -feeDaily
		if cfg.EarnCash {
			r += cash[k]
		}
		for i := range positions {
			r += positions[i] * excess[i][k]
		}
		idx := k - start
		values[idx] = values[idx-1] * (1 + r)
	}
	return values, start, nil
}

// rollingVol is the annualized standard deviation of xs[(k-window):k].
func rollingVol(xs []float64, k, window int) float64 {
	lo := max(k-window, 1)
	n := k - lo
	if n < 2 {
		return 0
	}
	m := 0.0
	for j := lo; j < k; j++ {
		m += xs[j]
	}
	m /= float64(n)
	v := 0.0
	for j := lo; j < k; j++ {
		v += (xs[j] - m) * (xs[j] - m)
	}
	return math.Sqrt(v/float64(n-1)) * math.Sqrt(252)
}
