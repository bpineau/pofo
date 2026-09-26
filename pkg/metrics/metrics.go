package metrics

import (
	"fmt"
	"math"
	"time"
)

const (
	tradingDaysPerYear = 252
	daysPerYear        = 365.25
	minBetaOverlap     = 30
)

// Stats summarizes the behaviour of a value series.
type Stats struct {
	Start, End  time.Time
	Years       float64
	CAGR        float64 // annualized growth rate (0.07 = +7 %/year)
	Volatility  float64 // annualized standard deviation of daily returns (0.16 = 16 %/year)
	Sharpe      float64 // annualized mean return / volatility, risk-free rate 0
	Sortino     float64 // annualized mean return / downside deviation
	Ulcer       float64 // Ulcer Index, in PERCENT POINTS (e.g. 12.8), not a fraction like the fields above
	MaxDrawdown float64 // deepest peak-to-trough loss (-0.55 = −55 %)
	TTRDays     int     // longest underwater stretch (peak to recovery), calendar days
	TTROngoing  bool    // the longest stretch had not recovered by End
	Beta        float64
	HasBeta     bool
	CWARP       float64 // Cole Wins Above Replacement Portfolio vs the benchmark, in percent (+ improves, - hurts)
	HasCWARP    bool
	Skew        float64 // skewness of daily returns (negative = longer left tail)
	Kurtosis    float64 // excess kurtosis of daily returns (>0 = fatter tails than normal)
}

// Compute derives Stats from a value series. dates must be ascending and
// values strictly positive and finite, both of equal length >= 2.
func Compute(dates []time.Time, values []float64) (Stats, error) {
	if len(dates) != len(values) || len(values) < 2 {
		return Stats{}, fmt.Errorf("series too short (%d points)", len(values))
	}
	for _, v := range values {
		// An infinity passes "> 0" and then poisons half the statistics with
		// NaN while leaving the other half (the CAGR, the drawdown) looking
		// like numbers. Refuse it where the series is still readable.
		if !(v > 0) || math.IsInf(v, 1) {
			return Stats{}, fmt.Errorf("non-positive or infinite value in series")
		}
	}
	var s Stats
	s.Start, s.End = dates[0], dates[len(dates)-1]
	s.Years = s.End.Sub(s.Start).Hours() / 24 / daysPerYear
	if s.Years <= 0 {
		return Stats{}, fmt.Errorf("empty period")
	}
	s.CAGR = math.Pow(values[len(values)-1]/values[0], 1/s.Years) - 1

	r := Returns(values)
	mean := Mean(r)

	variance, downSq := 0.0, 0.0
	for _, x := range r {
		variance += (x - mean) * (x - mean)
		if x < 0 {
			downSq += x * x
		}
	}
	s.Volatility, s.Sharpe, s.Sortino = math.NaN(), math.NaN(), math.NaN()
	if len(r) >= 2 {
		std := math.Sqrt(variance / float64(len(r)-1))
		s.Volatility = std * math.Sqrt(tradingDaysPerYear)
		if s.Volatility > 0 {
			s.Sharpe = mean * tradingDaysPerYear / s.Volatility
		}
	}
	if downDev := math.Sqrt(downSq/float64(len(r))) * math.Sqrt(tradingDaysPerYear); downDev > 0 {
		s.Sortino = mean * tradingDaysPerYear / downDev
	}
	s.Skew = Skewness(r)
	s.Kurtosis = ExcessKurtosis(r)

	// Drawdown-derived statistics.
	peak, peakDate := values[0], dates[0]
	sumSqDD, maxDD := 0.0, 0.0
	maxTTR := time.Duration(0)
	ongoing := false
	for i, v := range values {
		if v >= peak {
			if spell := dates[i].Sub(peakDate); spell > maxTTR {
				maxTTR = spell
				ongoing = false
			}
			peak, peakDate = v, dates[i]
		}
		dd := v/peak - 1
		if dd < maxDD {
			maxDD = dd
		}
		sumSqDD += dd * dd * 10000 // drawdown in percent, squared
	}
	if spell := s.End.Sub(peakDate); spell > maxTTR {
		maxTTR = spell
		ongoing = true
	}
	s.MaxDrawdown = maxDD
	s.Ulcer = math.Sqrt(sumSqDD / float64(len(values)))
	s.TTRDays = int(math.Round(maxTTR.Hours() / 24))
	s.TTROngoing = ongoing
	return s, nil
}

// Returns computes simple daily returns between consecutive values. It
// returns nil for fewer than two values.
func Returns(values []float64) []float64 {
	if len(values) < 2 {
		return nil
	}
	r := make([]float64, 0, len(values)-1)
	for i := 1; i < len(values); i++ {
		r = append(r, values[i]/values[i-1]-1)
	}
	return r
}

// Beta regresses the series' daily returns on the benchmark's, matching
// observations by date. ok is false when fewer than 30 dates overlap.
func Beta(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64) (float64, bool) {
	p := pairReturns(dates, values, benchDates, benchValues)
	if len(p.bench) < minBetaOverlap {
		return 0, false
	}
	b := slope(p.bench, p.own)
	if math.IsNaN(b) {
		return 0, false
	}
	return b, true
}

// paired is two series' returns matched by date: own[k] and bench[k] are
// each series' return from its own previous point to end[k], a date both
// quote, and start[k] is the date own[k] is measured from.
type paired struct {
	start, end []time.Time
	own, bench []float64
}

// pairReturns matches a series' simple returns with a benchmark's on the
// dates both quote, each return taken from that series' own previous point:
// the pairing Beta, VsBenchmark, RollingBeta and RollingCorr share. It is
// empty when the slices are mismatched or shorter than two points.
func pairReturns(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64) paired {
	var p paired
	if len(dates) != len(values) || len(dates) < 2 || len(benchDates) != len(benchValues) || len(benchDates) < 2 {
		return p
	}
	bench := make(map[time.Time]float64, len(benchDates)-1)
	for i := 1; i < len(benchDates); i++ {
		bench[benchDates[i]] = benchValues[i]/benchValues[i-1] - 1
	}
	for i := 1; i < len(dates); i++ {
		if br, ok := bench[dates[i]]; ok {
			p.start = append(p.start, dates[i-1])
			p.end = append(p.end, dates[i])
			p.own = append(p.own, values[i]/values[i-1]-1)
			p.bench = append(p.bench, br)
		}
	}
	return p
}

// slope is the least-squares slope of ys on xs, cov(x, y) / var(x): NaN when
// xs is constant or the samples hold fewer than two points.
func slope(xs, ys []float64) float64 {
	if len(xs) < 2 || len(xs) != len(ys) {
		return math.NaN()
	}
	mx, my := Mean(xs), Mean(ys)
	var cov, varx float64
	for i := range xs {
		cov += (xs[i] - mx) * (ys[i] - my)
		varx += (xs[i] - mx) * (xs[i] - mx)
	}
	if varx == 0 {
		return math.NaN()
	}
	return cov / varx
}

// Mean returns the arithmetic mean of xs.
func Mean(xs []float64) float64 {
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}
