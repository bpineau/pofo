package scenario

import (
	"math"
	"math/rand/v2"
)

// BlockBootstrap resamples contiguous blocks of BlockLen periods from the
// Panel's history (sampling on the time axis, so cross-asset correlations
// and regimes survive), applies Weights (nil uses Panel.Weights) and
// concatenates until Periods returns are produced.
type BlockBootstrap struct {
	Panel    Panel
	Weights  []float64
	BlockLen int
	Periods  int
}

// Len reports the path length.
func (b BlockBootstrap) Len() int { return b.Periods }

// Draw returns one resampled path.
func (b BlockBootstrap) Draw(rng *rand.Rand) Sequence {
	hist := b.Panel.Combine(b.Weights)
	return blocks(rng, hist, b.BlockLen, b.Periods, func() bool { return false })
}

// StationaryBootstrap is a block bootstrap with random block lengths drawn
// from a geometric distribution of mean MeanBlock (Politis-Romano): each
// period continues the current block with probability 1-1/MeanBlock, else
// starts a new random block. It avoids the fixed-length artefacts of
// BlockBootstrap.
type StationaryBootstrap struct {
	Panel     Panel
	Weights   []float64
	MeanBlock float64
	Periods   int
}

// Len reports the path length.
func (s StationaryBootstrap) Len() int { return s.Periods }

// Draw returns one resampled path.
func (s StationaryBootstrap) Draw(rng *rand.Rand) Sequence {
	hist := s.Panel.Combine(s.Weights)
	pNew := 1.0
	if s.MeanBlock > 1 {
		pNew = 1 / s.MeanBlock
	}
	return blocks(rng, hist, math.MaxInt, s.Periods, func() bool { return rng.Float64() < pNew })
}

// blocks builds a path of n periods by copying from hist starting at random
// indices. A new block starts when the running block reaches blockLen or
// restart() returns true. hist is treated circularly.
func blocks(rng *rand.Rand, hist Sequence, blockLen, n int, restart func() bool) Sequence {
	out := make(Sequence, 0, n)
	h := len(hist)
	if h == 0 {
		return make(Sequence, n)
	}
	pos, left := rng.IntN(h), blockLen
	for len(out) < n {
		if left == 0 || restart() {
			pos, left = rng.IntN(h), blockLen
		}
		out = append(out, hist[pos%h])
		pos++
		if left != math.MaxInt {
			left--
		}
	}
	return out
}
