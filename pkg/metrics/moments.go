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
// interpolation on the sorted sample, convenient for QQ comparisons. xs is not
// modified.
//
// A quantile reads at most two order statistics and callers ask for a handful
// of them, so only the ranks needed are placed (selectRanks, linear on
// average) rather than the whole sample sorted: a decumulation wealth fan is
// five percentiles of thousands of paths, once per year of the horizon. The
// values are the order statistics either way, so the result is identical.
func Quantiles(xs []float64, qs ...float64) []float64 {
	if len(xs) == 0 {
		return nil
	}
	buf := append([]float64(nil), xs...)
	ranks := make([]int, 0, 2*len(qs))
	for _, q := range qs {
		lo, hi := quantileRanks(len(buf), q)
		ranks = append(ranks, lo, hi)
	}
	sort.Ints(ranks)
	selectRanks(buf, ranks)
	out := make([]float64, len(qs))
	for i, q := range qs {
		lo, hi := quantileRanks(len(buf), q)
		pos := q * float64(len(buf)-1)
		out[i] = buf[lo] + (pos-float64(lo))*(buf[hi]-buf[lo])
	}
	return out
}

// quantileRanks is the pair of order statistics quantile q interpolates
// between in a sample of n values, clamped to the sample.
func quantileRanks(n int, q float64) (lo, hi int) {
	pos := q * float64(n-1)
	lo, hi = int(math.Floor(pos)), int(math.Ceil(pos))
	if lo < 0 {
		lo = 0
	}
	if hi >= n {
		hi = n - 1
	}
	return lo, hi
}

// TopK returns the n largest values of xs, in descending order, without
// modifying xs. It is the raw material of a tail statistic (a CVaR, a
// conditional drawdown): only the rank that cuts the tail off is placed, and
// only the tail itself is then sorted, so reading the worst 5 % of a large
// sample costs a fraction of sorting all of it. n at or above len(xs) returns
// the whole sample sorted descending, and n <= 0 returns nil.
func TopK(xs []float64, n int) []float64 {
	if n <= 0 || len(xs) == 0 {
		return nil
	}
	if n > len(xs) {
		n = len(xs)
	}
	buf := append([]float64(nil), xs...)
	cut := len(buf) - n
	selectRanks(buf, []int{cut})
	tail := buf[cut:]
	sort.Float64s(tail)
	for i, j := 0, len(tail)-1; i < j; i, j = i+1, j-1 {
		tail[i], tail[j] = tail[j], tail[i]
	}
	return tail
}

// selectRanksCutoff is the sample size below which placing a few ranks is not
// worth the partitioning machinery and a plain sort wins.
const selectRanksCutoff = 48

// selectRanks reorders xs in place so that every listed rank holds the order
// statistic of that rank, with every smaller value before it and every larger
// one after. ranks must be ascending; duplicates are allowed.
//
// Small samples, and samples holding a NaN (whose comparisons are all false,
// which no partition can order), fall back to the full sort, which places
// every rank at once and keeps NaN exactly where sort.Float64s puts it.
func selectRanks(xs []float64, ranks []int) {
	if len(xs) < selectRanksCutoff || hasNaN(xs) {
		sort.Float64s(xs)
		return
	}
	multiselect(xs, 0, len(xs)-1, ranks)
}

// multiselect places every rank of ranks (ascending, all inside [lo, hi]) by
// splitting the span once and then descending only into the halves that still
// hold a wanted rank. Each level of the split costs one pass of the span, so k
// ranks cost about n log k where k independent quickselects would cost k*n:
// the difference shows on a wealth fan, five percentiles of thousands of paths
// once per year of the horizon.
func multiselect(xs []float64, lo, hi int, ranks []int) {
	for len(ranks) > 0 && lo < hi {
		if hi-lo < selectRanksCutoff {
			sort.Float64s(xs[lo : hi+1]) // small span: sorting places every rank at once
			return
		}
		p := partition(xs, lo, hi)
		// The ranks up to p are settled inside the left half; the loop carries
		// on with the right one, so only the left descent recurses.
		i := sort.SearchInts(ranks, p+1)
		multiselect(xs, lo, p, ranks[:i])
		ranks, lo = ranks[i:], p+1
	}
}

// hasNaN reports whether the sample holds a NaN.
func hasNaN(xs []float64) bool {
	for _, x := range xs {
		if math.IsNaN(x) {
			return true
		}
	}
	return false
}

// partition splits xs[lo:hi+1] around the median of its first, middle and last
// element (Hoare's scheme, which splits runs of equal values evenly instead of
// degenerating on them, as a wealth column full of zeroed ruined paths is).
// It returns p with lo <= p < hi, every element up to p at most every element
// after it.
func partition(xs []float64, lo, hi int) int {
	pivot := medianOf3(xs[lo], xs[(lo+hi)/2], xs[hi])
	i, j := lo-1, hi+1
	for {
		for i++; xs[i] < pivot; i++ {
		}
		for j--; xs[j] > pivot; j-- {
		}
		if i >= j {
			return j
		}
		xs[i], xs[j] = xs[j], xs[i]
	}
}

// medianOf3 returns the middle of three values.
func medianOf3(a, b, c float64) float64 {
	if a > b {
		a, b = b, a
	}
	if b > c {
		b = c
	}
	if a > b {
		return a
	}
	return b
}
