package metrics

import (
	"cmp"
	"math"
	"slices"
)

// LowestK returns the POSITIONS of the k smallest values of xs, smallest
// first: the worst k months of a return column, which the caller maps to
// their dates (the Ends of a marketdata.Panel) and to the other columns on
// the same periods (Panel.Pick). Equal values keep their order in xs, the
// earlier position first, so the answer never depends on the sort. NaN
// values are never selected; k beyond the number of the others returns them
// all, and k <= 0 returns nil. xs is not modified.
//
// TopK is the values-only sibling for large samples: it places just the
// order statistic that cuts the tail, where LowestK sorts every position.
func LowestK(xs []float64, k int) []int {
	return extremeK(xs, k, cmp.Compare[float64])
}

// HighestK returns the positions of the k largest values of xs, largest
// first, with LowestK's rules: equal values in their order in xs, NaN never
// selected, xs not modified.
func HighestK(xs []float64, k int) []int {
	return extremeK(xs, k, func(a, b float64) int { return cmp.Compare(b, a) })
}

// extremeK is the k first positions of xs's non-NaN values under order, a
// stable sort.
func extremeK(xs []float64, k int, order func(a, b float64) int) []int {
	if k <= 0 {
		return nil
	}
	idx := make([]int, 0, len(xs))
	for i, x := range xs {
		if !math.IsNaN(x) {
			idx = append(idx, i)
		}
	}
	slices.SortStableFunc(idx, func(a, b int) int { return order(xs[a], xs[b]) })
	if k < len(idx) {
		idx = idx[:k:k]
	}
	if len(idx) == 0 {
		return nil
	}
	return idx
}
