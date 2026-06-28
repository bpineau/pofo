package metrics

import (
	"math"
	"sort"
)

// Skewness returns the (population) skewness of xs, the standardized third
// moment. Positive values mean a longer right tail, negative a longer left
// tail; a symmetric distribution scores 0. NaN when xs has fewer than two
// points or zero variance.
func Skewness(xs []float64) float64 {
	n := float64(len(xs))
	if n < 2 {
		return math.NaN()
	}
	m := Mean(xs)
	var m2, m3 float64
	for _, x := range xs {
		d := x - m
		m2 += d * d
		m3 += d * d * d
	}
	m2 /= n
	m3 /= n
	if m2 == 0 {
		return math.NaN()
	}
	return m3 / math.Pow(m2, 1.5)
}

// ExcessKurtosis returns the (population) excess kurtosis of xs, the
// standardized fourth moment minus 3. It is 0 for a normal distribution;
// positive values signal fatter tails (more extreme returns) than normal.
// NaN when xs has fewer than two points or zero variance.
func ExcessKurtosis(xs []float64) float64 {
	n := float64(len(xs))
	if n < 2 {
		return math.NaN()
	}
	m := Mean(xs)
	var m2, m4 float64
	for _, x := range xs {
		d := x - m
		d2 := d * d
		m2 += d2
		m4 += d2 * d2
	}
	m2 /= n
	m4 /= n
	if m2 == 0 {
		return math.NaN()
	}
	return m4/(m2*m2) - 3
}

// Autocorr returns the sample autocorrelation of xs at lags 0..lags
// inclusive (so the result has lags+1 entries, with index 0 always 1).
// Returns nil when xs has fewer than two points or zero variance.
func Autocorr(xs []float64, lags int) []float64 {
	n := len(xs)
	if n < 2 || lags < 0 {
		return nil
	}
	m := Mean(xs)
	var c0 float64
	for _, x := range xs {
		d := x - m
		c0 += d * d
	}
	if c0 == 0 {
		return nil
	}
	out := make([]float64, lags+1)
	for k := 0; k <= lags && k < n; k++ {
		var ck float64
		for i := k; i < n; i++ {
			ck += (xs[i] - m) * (xs[i-k] - m)
		}
		out[k] = ck / c0
	}
	return out
}

// Histogram buckets xs into the given number of equal-width bins spanning
// [min, max]. It returns the bins+1 bin edges and the per-bin counts;
// values equal to the maximum fall into the last bin. Returns nil, nil when
// bins < 1 or xs is empty.
func Histogram(xs []float64, bins int) (edges []float64, counts []int) {
	if bins < 1 || len(xs) == 0 {
		return nil, nil
	}
	lo, hi := xs[0], xs[0]
	for _, x := range xs {
		lo, hi = math.Min(lo, x), math.Max(hi, x)
	}
	edges = make([]float64, bins+1)
	width := (hi - lo) / float64(bins)
	for i := range edges {
		edges[i] = lo + width*float64(i)
	}
	edges[bins] = hi
	counts = make([]int, bins)
	if width == 0 { // degenerate: every value identical
		counts[0] = len(xs)
		return edges, counts
	}
	for _, x := range xs {
		b := int((x - lo) / width)
		if b >= bins {
			b = bins - 1
		}
		counts[b]++
	}
	return edges, counts
}

// Quantiles returns the q-quantiles of xs (q in [0,1]) by linear
// interpolation on the sorted sample, convenient for QQ comparisons.
func Quantiles(xs []float64, qs ...float64) []float64 {
	if len(xs) == 0 {
		return nil
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	out := make([]float64, len(qs))
	for i, q := range qs {
		pos := q * float64(len(sorted)-1)
		lo := int(math.Floor(pos))
		hi := int(math.Ceil(pos))
		if lo < 0 {
			lo = 0
		}
		if hi >= len(sorted) {
			hi = len(sorted) - 1
		}
		out[i] = sorted[lo] + (pos-float64(lo))*(sorted[hi]-sorted[lo])
	}
	return out
}
