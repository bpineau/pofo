package scenario

import "math/rand/v2"

// PooledBootstrap resamples contiguous blocks from a POOL of separate history
// series, never crossing a series boundary within a block. Each block draws a
// random series then a random start inside it, and continues one period at a
// time with probability 1-1/MeanBlock (a stationary bootstrap, Politis-Romano),
// otherwise starts a fresh block from another randomly chosen series.
//
// Unlike a single-history bootstrap, it keeps each series' internal ordering
// intact while mixing across series between blocks: fed the per-country real
// equity records of a broad developed-market sample, it models a "random
// developed-market retiree" whose run can land inside any single market's
// disaster (France or Portugal early-century, Japan post-1990). Pre-averaging
// the series into a diversified world index would erase exactly that sequence
// risk, so the pooling happens here, at draw time, not in the data.
type PooledBootstrap struct {
	Series    [][]float64 // separate histories; a block stays within one series
	MeanBlock float64     // mean block length in periods (>=1)
	Periods   int         // path length to produce
}

// Len reports the path length.
func (p PooledBootstrap) Len() int { return p.Periods }

// Draw returns one pooled resampled path. Empty series are skipped; if the pool
// holds no usable series it returns a zero path.
func (p PooledBootstrap) Draw(rng *rand.Rand) Sequence {
	out := make(Sequence, p.Periods)
	cont := 1.0
	if p.MeanBlock > 1 {
		cont = 1 - 1/p.MeanBlock
	}
	var cur []float64
	var idx int
	pick := func() {
		for range 8 { // a few tries to skip empty series
			s := p.Series[rng.IntN(len(p.Series))]
			if len(s) > 0 {
				cur, idx = s, rng.IntN(len(s))
				return
			}
		}
		cur, idx = nil, 0
	}
	if len(p.Series) == 0 {
		return out
	}
	for t := 0; t < p.Periods; t++ {
		if cur == nil || idx >= len(cur) || (t > 0 && rng.Float64() >= cont) {
			pick()
		}
		if cur == nil {
			continue // pool exhausted of usable series; leave zero
		}
		out[t] = cur[idx]
		idx++
	}
	return out
}
