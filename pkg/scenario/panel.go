package scenario

// Panel is an aligned matrix of per-asset real periodic returns plus the
// default weights of a portfolio over them. Returns is indexed
// [asset][period]; every asset row has the same length (use Deflate and an
// aligner to build it). Because resampling happens on the time axis
// (Periods), cross-asset correlations and historical regimes are preserved.
type Panel struct {
	Returns [][]float64 // [asset][period]
	Weights []float64   // default portfolio weights, summing to 1
}

// Periods is the number of historical periods, 0 for an empty panel.
func (p Panel) Periods() int {
	if len(p.Returns) == 0 {
		return 0
	}
	return len(p.Returns[0])
}

// Combine collapses the panel into one portfolio return path using weights
// (nil uses p.Weights). Reweighting is cheap, so live allocation changes do
// not need the underlying series refetched.
func (p Panel) Combine(weights []float64) Sequence {
	if weights == nil {
		weights = p.Weights
	}
	out := make(Sequence, p.Periods())
	for a, row := range p.Returns {
		w := weights[a]
		for t, r := range row {
			out[t] += w * r
		}
	}
	return out
}
