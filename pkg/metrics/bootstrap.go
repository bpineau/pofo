package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// MaxBootstrapRepetitions bounds BlockBootstrap's retained sampling results
// to three slices of at most 100000 values, independent of sample length.
const MaxBootstrapRepetitions = 100000

// BootstrapOptions supplies the explicit assumptions for BlockBootstrap.
// There are no inferred defaults for cadence, block length or confidence.
type BootstrapOptions struct {
	// PeriodsPerYear annualizes returns and volatility; it must be positive and finite.
	PeriodsPerYear float64
	// RiskFree is a finite annual arithmetic fraction used only in Sharpe.
	RiskFree float64
	// BlockLength is the number of consecutive circular observations per draw, in [1,N].
	BlockLength int
	// Repetitions is the number of resamples, in [2,MaxBootstrapRepetitions].
	Repetitions int
	// Confidence is the central percentile interval's probability, in (0,1).
	Confidence float64
	// Seed fixes the local PRNG for reproducibility; zero is a valid fixed seed.
	Seed int64
}

// Interval is a point estimate and a central bootstrap percentile interval.
// Nonfinite replicate statistics are omitted; Defined reports how many
// remained. An undefined point or absent bounds retain NaN in Go and null in
// JSON. Intervals with omitted replicates are conditional on being defined.
type Interval struct {
	// Estimate is the statistic measured on the original observations.
	Estimate float64
	// Lower is the lower percentile of defined replicate statistics.
	Lower float64
	// Upper is the upper percentile of defined replicate statistics.
	Upper float64
	// Defined counts finite replicate statistics used for this interval.
	Defined int
}

// MarshalJSON preserves undefined estimates and interval endpoints as null.
func (i Interval) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Estimate, Lower, Upper jsonNumber
		Defined                int
	}{jsonNumber(i.Estimate), jsonNumber(i.Lower), jsonNumber(i.Upper), i.Defined})
}

// BootstrapResult contains estimation uncertainty on one historical sample.
// It does not describe a future investment outcome or a retirement forecast.
type BootstrapResult struct {
	// CAGR is annualized geometric return and its percentile interval.
	CAGR Interval
	// Volatility is annualized sample standard deviation and its interval.
	Volatility Interval
	// Sharpe is annualized arithmetic excess return over Volatility and its interval.
	Sharpe Interval
	// Observations is the original sample size and every replicate's length.
	Observations int
	// Options records the explicit resampling and measurement assumptions.
	Options BootstrapOptions
}

// BlockBootstrap estimates uncertainty of CAGR, volatility and Sharpe from
// one aligned, regularly spaced simple-return sample. Each replicate draws
// uniform starting positions with replacement, copies BlockLength consecutive
// observations with circular wraparound, and truncates the final block to N.
// BlockLength 1 is the IID bootstrap. Overlapping circular blocks preserve
// within-block dependence; the caller chooses a defensible block length.
//
// These central percentile intervals describe estimation uncertainty under
// resampling of the observed past, not future returns or predictive coverage.
// No future data are used or fitted. A fixed local seed makes repeated calls
// reproducible. Finite returns greater than -1 and at least two observations
// are required. Constant returns have zero volatility and undefined Sharpe.
//
// Work is O(N*Repetitions); storage is O(N+Repetitions), never a matrix of all
// replicate paths. The caller explicitly budgets repetitions, bounded by
// MaxBootstrapRepetitions. Finite statistics only enter their own interval;
// always inspect Interval.Defined, especially on short or degenerate samples.
func BlockBootstrap(returns []float64, opt BootstrapOptions) (BootstrapResult, error) {
	n := len(returns)
	if n < 2 {
		return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: need at least two returns")
	}
	if !(opt.PeriodsPerYear > 0) || isNaNOrInf(opt.PeriodsPerYear) || isNaNOrInf(opt.RiskFree) {
		return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: cadence must be finite and positive, risk-free rate finite")
	}
	if opt.BlockLength < 1 || opt.BlockLength > n {
		return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: block length must be in [1,N]")
	}
	if opt.Repetitions < 2 || opt.Repetitions > MaxBootstrapRepetitions {
		return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: repetitions must be in [2,%d]", MaxBootstrapRepetitions)
	}
	if !(opt.Confidence > 0 && opt.Confidence < 1) {
		return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: confidence must be in (0,1)")
	}
	for i, r := range returns {
		if isNaNOrInf(r) || r <= -1 {
			return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: return %d must be finite and greater than -1", i)
		}
	}
	point := bootstrapStats(returns, opt)
	if isNaNOrInf(point[0]) || isNaNOrInf(point[1]) {
		return BootstrapResult{}, fmt.Errorf("metrics: BlockBootstrap: point statistics overflow")
	}
	distributions := [3][]float64{}
	for i := range distributions {
		distributions[i] = make([]float64, 0, opt.Repetitions)
	}
	rng := rand.New(rand.NewSource(opt.Seed))
	sample := make([]float64, n)
	for rep := 0; rep < opt.Repetitions; rep++ {
		for out := 0; out < n; {
			start := rng.Intn(n)
			for j := 0; j < opt.BlockLength && out < n; j++ {
				sample[out] = returns[(start+j)%n]
				out++
			}
		}
		stats := bootstrapStats(sample, opt)
		for i, v := range stats {
			if !isNaNOrInf(v) {
				distributions[i] = append(distributions[i], v)
			}
		}
	}
	intervals := [3]Interval{}
	tail := (1 - opt.Confidence) / 2
	for i, dist := range distributions {
		interval := Interval{Estimate: point[i], Lower: math.NaN(), Upper: math.NaN(), Defined: len(dist)}
		if len(dist) > 0 {
			sort.Float64s(dist)
			interval.Lower = sortedPercentile(dist, tail)
			interval.Upper = sortedPercentile(dist, 1-tail)
		}
		intervals[i] = interval
	}
	return BootstrapResult{CAGR: intervals[0], Volatility: intervals[1], Sharpe: intervals[2], Observations: n, Options: opt}, nil
}

// bootstrapStats computes the same bare-return estimators as the metrics
// helpers, with an explicitly undefined Sharpe for a constant sample.
func bootstrapStats(returns []float64, opt BootstrapOptions) [3]float64 {
	logSum := 0.0
	constant := true
	for _, r := range returns {
		logSum += math.Log1p(r)
		constant = constant && r == returns[0]
	}
	cagr := math.Expm1(logSum / float64(len(returns)) * opt.PeriodsPerYear)
	vol := Volatility(returns, opt.PeriodsPerYear)
	sharpe := math.NaN()
	if constant {
		vol = 0
	} else if vol > 0 {
		sharpe = (Mean(returns)*opt.PeriodsPerYear - opt.RiskFree) / vol
	}
	return [3]float64{cagr, vol, sharpe}
}

// sortedPercentile linearly interpolates a nonempty sorted sample.
func sortedPercentile(values []float64, p float64) float64 {
	at := p * float64(len(values)-1)
	lo := int(math.Floor(at))
	hi := int(math.Ceil(at))
	return values[lo] + (values[hi]-values[lo])*(at-float64(lo))
}
