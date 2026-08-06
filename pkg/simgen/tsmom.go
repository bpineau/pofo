package simgen

import (
	"fmt"
	"math"
)

// TSMOMConfig drives the classic time-series-momentum reconstruction used
// for managed-futures funds (DBMF, KMLM, CTA…): each market is held long or
// short according to the sign of its trailing 12-month excess return and
// weighted inverse to its own volatility (risk parity), then the whole book is
// scaled so its covariance-implied volatility meets TargetVol.
//
// The scaling uses the full rolling covariance rather than summing per-market
// risks in isolation: a basket whose legs are positively correlated (the equity
// legs among themselves, the bond legs among themselves) realizes more
// volatility than the sum of standalone risks would suggest, so ignoring the
// cross-terms overshoots the target by the correlation factor.
//
// Trend signals turn slowly, so they are refreshed every Rebalance days, but
// risk is rescaled EVERY day against an exponentially weighted covariance
// (CovHalfLife). The two cadences are deliberately separate. Sizing the book
// once a month off a flat trailing window leaves it frozen at its pre-storm
// size through a volatility explosion: in February 2020 the engine sized on
// the 14th against three calm months, then held that book through the crash
// to the 17th of March, losing 12 % on the 12th against a 10 % volatility
// target while the funds it replicates lost 3 to 4 %. A daily rescale on a
// decaying estimator cuts as the storm builds, which is what the real
// programmes do.
//
// This is the standard academic replication (Moskowitz–Ooi–Pedersen); real
// funds add carry, faster signals and execution details, so expect daily
// correlations around 0.5–0.7, not 0.9.
type TSMOMConfig struct {
	Markets     []string // component ids, traded as excess returns vs cash
	CashID      string   // e.g. "^IRX"
	Lookback    int      // signal window in trading days (e.g. 252)
	VolWindow   int      // window of the staleness check and the covariance seed (e.g. 63)
	CovHalfLife int      // EWMA half-life of the risk model in days (0 = defaultCovHalfLife)
	Rebalance   int      // refresh the trend signal every N trading days (e.g. 21)
	TargetVol   float64  // annualized portfolio volatility target (e.g. 0.10)
	MaxLeverage float64  // cap on a single market's position (e.g. 2)
	AnnualFee   float64  // e.g. 0.0085
	EarnCash    bool     // collateral earns the cash rate
}

// defaultCovHalfLife is the decay of the risk model when a config leaves it
// unset: about one trading month, the usual compromise between reacting to a
// volatility regime change and trading on noise.
const defaultCovHalfLife = 23

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
	// Excess returns per market, plus the raw price returns kept alongside to
	// detect stale (forward-filled) legs in the deep backcast.
	excess := make([][]float64, len(cfg.Markets))
	raw := make([][]float64, len(cfg.Markets))
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
		raw[i] = r
	}

	start := cfg.Lookback + 1
	positions := make([]float64, len(cfg.Markets))
	signs := make([]float64, len(cfg.Markets))
	values := make([]float64, n-start)
	values[0] = 100
	feeDaily := cfg.AnnualFee / 252
	sinceSignal := cfg.Rebalance // force a signal refresh on the first step

	// The risk model is seeded with the flat covariance of the warm-up window,
	// then decayed forward one day at a time.
	halfLife := cfg.CovHalfLife
	if halfLife <= 0 {
		halfLife = defaultCovHalfLife
	}
	lambda := math.Pow(0.5, 1/float64(halfLife))
	cov := rollingCov(excess, start, cfg.VolWindow)

	for k := start + 1; k < n; k++ {
		decayCov(cov, excess, k-1, lambda)
		if sinceSignal >= cfg.Rebalance {
			sinceSignal = 0
			refreshSigns(signs, excess, raw, k, cfg)
		}
		sinceSignal++
		sizePositions(positions, signs, cov, raw, k, cfg)

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

// refreshSigns sets each market long or short by the sign of its trailing
// Lookback excess return, at date index k. A market whose proxy is stale
// (mostly forward-filled in the deep backcast, which would corrupt the
// covariance) is left flat until it trades again.
func refreshSigns(signs []float64, excess, raw [][]float64, k int, cfg TSMOMConfig) {
	for i := range excess {
		if activeFraction(raw[i], k, cfg.VolWindow) < 0.5 {
			signs[i] = 0
			continue
		}
		cum := 1.0
		for j := k - cfg.Lookback; j < k; j++ {
			cum *= 1 + excess[i][j]
		}
		signs[i] = 1
		if cum < 1 {
			signs[i] = -1
		}
	}
}

// sizePositions turns the current signs into positions at date index k: each
// market weighted inverse to its own volatility (risk parity), then the book
// scaled as a whole so its covariance-implied volatility meets cfg.TargetVol,
// and finally capped at ±MaxLeverage per market. It runs every day, so the
// book shrinks as volatility rises even between two signal refreshes. A market
// with no measurable volatility, or one that has gone stale since the last
// refresh, stays flat.
func sizePositions(positions, signs []float64, cov [][]float64, raw [][]float64, k int, cfg TSMOMConfig) {
	w := make([]float64, len(signs))
	for i := range signs {
		if signs[i] == 0 || activeFraction(raw[i], k, cfg.VolWindow) < 0.5 {
			continue
		}
		if vol := math.Sqrt(cov[i][i]); vol > 0 {
			w[i] = signs[i] / vol
		}
	}
	// Portfolio variance of the risk-parity weights, cross-terms included.
	portVar := 0.0
	for i := range w {
		for j := range w {
			portVar += w[i] * w[j] * cov[i][j]
		}
	}
	scale := 0.0
	if portVar > 0 {
		scale = cfg.TargetVol / math.Sqrt(portVar)
	}
	for i := range positions {
		positions[i] = math.Max(-cfg.MaxLeverage, math.Min(cfg.MaxLeverage, scale*w[i]))
	}
}

// activeFraction is the share of the window [k-window, k) on which the raw
// price return actually moved. A live daily leg sits near 1; a monthly proxy
// forward-filled onto the daily frame sits near 1/21, flagging it as stale.
func activeFraction(xs []float64, k, window int) float64 {
	lo := max(k-window, 1)
	n := k - lo
	if n < 1 {
		return 0
	}
	nz := 0
	for j := lo; j < k; j++ {
		if xs[j] != 0 {
			nz++
		}
	}
	return float64(nz) / float64(n)
}

// decayCov folds the returns of date index k into the exponentially weighted
// covariance in place: cov <- lambda·cov + (1-lambda)·r_k·r_kᵀ, annualized,
// the RiskMetrics recursion. Zero mean, the usual convention for daily risk:
// over a day the drift is negligible next to the noise, and it keeps the
// estimator to one pass per day instead of a full window rescan.
func decayCov(cov [][]float64, excess [][]float64, k int, lambda float64) {
	for i := range cov {
		for j := i; j < len(cov); j++ {
			c := lambda*cov[i][j] + (1-lambda)*excess[i][k]*excess[j][k]*252
			cov[i][j], cov[j][i] = c, c
		}
	}
}

// rollingCov is the annualized covariance matrix of the excess-return series
// over the window [k-window, k). Element [i][j] is Cov(excess_i, excess_j)·252.
func rollingCov(excess [][]float64, k, window int) [][]float64 {
	m := len(excess)
	cov := make([][]float64, m)
	for i := range cov {
		cov[i] = make([]float64, m)
	}
	lo := max(k-window, 1)
	n := k - lo
	if n < 2 {
		return cov
	}
	means := make([]float64, m)
	for i := range excess {
		s := 0.0
		for j := lo; j < k; j++ {
			s += excess[i][j]
		}
		means[i] = s / float64(n)
	}
	for i := range m {
		for j := i; j < m; j++ {
			c := 0.0
			for t := lo; t < k; t++ {
				c += (excess[i][t] - means[i]) * (excess[j][t] - means[j])
			}
			c = c / float64(n-1) * 252
			cov[i][j], cov[j][i] = c, c
		}
	}
	return cov
}
