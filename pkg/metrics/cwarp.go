package metrics

import (
	"math"
	"time"
)

// defaultCWARPWeight is Artemis's standard overlay size: the new asset is
// layered on the replacement at 25 % of its notional, financed by borrowing.
const defaultCWARPWeight = 0.25

// CWARPParams configures the CWARP overlay. The zero value uses Artemis's
// published standard: Weight 0.25, RiskFree 0, Financing 0.
type CWARPParams struct {
	Weight    float64 // overlay weight w as a fraction of the replacement (0 -> 0.25)
	RiskFree  float64 // annual risk-free rate, subtracted in every ratio's numerator
	Financing float64 // annual borrow cost charged on the overlaid notional
}

// CWARP is the Cole Wins Above Replacement Portfolio score of overlaying
// `asset` on `replacement`, from Artemis Capital Management's "Moneyball for
// Modern Portfolio Theory" (2020). It answers whether adding the asset (or
// whole portfolio) on top of a pre-existing replacement portfolio, at weight w
// and financed by borrowing, improves the replacement's risk-adjusted returns.
//
// The new portfolio's per-period return is r_new = r_repl + w*(r_asset - fin),
// where fin is the daily financing charge. CWARP is the geometric average of
// the improvements in two ratios, minus one, in percent:
//
//	CWARP = ( sqrt( (Sortino_new/Sortino_repl) * (RtMDD_new/RtMDD_repl) ) - 1 ) * 100
//
// with Sortino the annualized excess return over downside deviation and RtMDD
// the annualized excess return over the maximum drawdown magnitude. A positive
// CWARP means the overlay lifts the portfolio (typically by adding return that
// is uncorrelated with, or a hedge against, the replacement's drawdowns); a
// negative one means it hurts. Unlike the Sharpe ratio, CWARP rewards
// non-correlation and skew because both denominators are measured on the
// combined series.
//
// asset and replacement are aligned daily simple-return series of equal length.
// ok is false when the inputs are too short, or when a ratio is undefined (the
// replacement never draws down, has no downside deviation, or a denominator is
// non-positive), in which case the score would be meaningless.
func CWARP(asset, replacement []float64, p CWARPParams) (float64, bool) {
	if len(asset) != len(replacement) || len(asset) < 2 {
		return 0, false
	}
	w := p.Weight
	if w == 0 {
		w = defaultCWARPWeight
	}
	fin := p.Financing / tradingDaysPerYear
	newRet := make([]float64, len(asset))
	for i := range asset {
		newRet[i] = replacement[i] + w*(asset[i]-fin)
	}

	sRepl := Sortino(replacement, p.RiskFree)
	sNew := Sortino(newRet, p.RiskFree)
	mRepl, okR := returnToMaxDrawdown(replacement, p.RiskFree)
	mNew, okN := returnToMaxDrawdown(newRet, p.RiskFree)
	if math.IsNaN(sRepl) || math.IsNaN(sNew) || !okR || !okN {
		return 0, false
	}
	return cwarpScore(sRepl, sNew, mRepl, mNew)
}

// CWARPvs computes the CWARP of `values` overlaid on a benchmark (the
// replacement portfolio), matching their daily returns by exact date. It is
// the convenient entry point for callers holding dated value series (a
// portfolio and its benchmark). ok is false when fewer than minBetaOverlap
// dates overlap or the score is undefined (see CWARP).
func CWARPvs(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64, p CWARPParams) (float64, bool) {
	if len(dates) != len(values) || len(benchDates) != len(benchValues) || len(dates) < 2 || len(benchDates) < 2 {
		return 0, false
	}
	bench := make(map[time.Time]float64, len(benchDates)-1)
	for i := 1; i < len(benchDates); i++ {
		bench[benchDates[i]] = benchValues[i]/benchValues[i-1] - 1
	}
	var asset, repl []float64
	for i := 1; i < len(dates); i++ {
		if br, found := bench[dates[i]]; found {
			asset = append(asset, values[i]/values[i-1]-1)
			repl = append(repl, br)
		}
	}
	if len(asset) < minBetaOverlap {
		return 0, false
	}
	return CWARP(asset, repl, p)
}

// cwarpScore combines the replacement and new-portfolio Sortino and
// return-to-max-drawdown ratios into the CWARP percent. ok is false when a
// replacement ratio is non-positive (the improvement factor is then
// meaningless) or the product under the root is negative (the overlay flipped
// the portfolio to a negative risk-adjusted return).
func cwarpScore(sortinoRepl, sortinoNew, rtmddRepl, rtmddNew float64) (float64, bool) {
	if !(sortinoRepl > 0) || !(rtmddRepl > 0) {
		return 0, false
	}
	prod := (sortinoNew / sortinoRepl) * (rtmddNew / rtmddRepl)
	if prod < 0 {
		return 0, false
	}
	return (math.Sqrt(prod) - 1) * 100, true
}

// returnToMaxDrawdown returns (CAGR - rfAnnual) / |maxDrawdown| for a daily
// return series, annualizing the compound growth by the number of periods
// (252 per year, matching the rest of the package). ok is false when the path
// wipes out or never draws down (the ratio is then undefined).
func returnToMaxDrawdown(returns []float64, rfAnnual float64) (float64, bool) {
	if len(returns) < 1 {
		return 0, false
	}
	v, peak, maxDD := 1.0, 1.0, 0.0
	for _, r := range returns {
		v *= 1 + r
		if !(v > 0) {
			return 0, false // capital wiped out
		}
		if v > peak {
			peak = v
		}
		if dd := v/peak - 1; dd < maxDD {
			maxDD = dd
		}
	}
	if maxDD == 0 {
		return 0, false // no drawdown: return-to-drawdown undefined
	}
	years := float64(len(returns)) / tradingDaysPerYear
	cagr := math.Pow(v, 1/years) - 1
	return (cagr - rfAnnual) / math.Abs(maxDD), true
}
