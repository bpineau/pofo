package decumul

import "github.com/bpineau/pofo/pkg/metrics"

// SpendStats summarises the lived cost of an adaptive spending policy across
// an ensemble: how often the household actually had to live below its uncut
// standard, when that first happened, and for how long. It quantifies the
// price of the flex/guardrails insurance in lifestyle terms rather than in
// ruin probability.
type SpendStats struct {
	EverCutShare   float64 // share of paths with at least one cut year
	FirstCutMedian float64 // median first cut year, among paths that cut
	CutYearsMedian float64 // median number of cut years, among paths that cut
	CutYearsP90    float64 // 90th percentile of cut years, among paths that cut
}

// SpendStats computes the bundle; the medians are conditional on cutting at
// least once (they answer "if I have to cut, when and for how long?").
func (e Ensemble) SpendStats() SpendStats {
	var s SpendStats
	if len(e.Paths) == 0 {
		return s
	}
	var firsts, lengths []float64
	for _, p := range e.Paths {
		if p.FirstCut >= 0 {
			firsts = append(firsts, float64(p.FirstCut))
			lengths = append(lengths, float64(p.CutYears))
		}
	}
	s.EverCutShare = float64(len(firsts)) / float64(len(e.Paths))
	if len(firsts) > 0 {
		s.FirstCutMedian = metrics.Quantiles(firsts, 0.50)[0]
		q := metrics.Quantiles(lengths, 0.50, 0.90)
		s.CutYearsMedian, s.CutYearsP90 = q[0], q[1]
	}
	return s
}

// SpendBands returns per-year quantiles of the delivered spending across
// paths: Bands[p][year] is the pcts[p] quantile of Spend that year. It is the
// spending counterpart of the wealth fan, showing how deep and how long the
// dips in living standard get.
func (e Ensemble) SpendBands(pcts []float64) [][]float64 {
	bands := make([][]float64, len(pcts))
	if len(e.Paths) == 0 || e.Years == 0 {
		return bands
	}
	for p := range bands {
		bands[p] = make([]float64, e.Years)
	}
	col := make([]float64, len(e.Paths))
	for y := 0; y < e.Years; y++ {
		for i, path := range e.Paths {
			col[i] = path.Spend[y]
		}
		q := metrics.Quantiles(col, pcts...)
		for p := range pcts {
			bands[p][y] = q[p]
		}
	}
	return bands
}
