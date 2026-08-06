package scenario

import "math/rand/v2"

// Annualize compounds consecutive groups of `group` periodic returns into one
// coarser return each: out[k] = Π(1+s[k*group+j]) − 1 for j in [0,group). A
// trailing partial group (when len(s) is not a multiple of group) is dropped.
// With group 12 it turns a path of monthly returns into annual returns.
func Annualize(s Sequence, group int) Sequence {
	if group < 1 {
		return s
	}
	n := len(s) / group
	out := make(Sequence, n)
	for k := 0; k < n; k++ {
		prod := 1.0
		for j := 0; j < group; j++ {
			prod *= 1 + s[k*group+j]
		}
		out[k] = prod - 1
	}
	return out
}

// Compounded wraps a higher-frequency Source and compounds each block of
// Group returns into one, so a monthly bootstrap or cohort source (Group 12)
// yields the annual paths the decumulation kernel expects. It reuses any
// inner Source verbatim; Inner is exported so callers can inspect it (e.g. to
// detect a cohorts source with too little history).
type Compounded struct {
	Inner Source
	Group int
}

// Len is the inner length divided by Group.
func (c Compounded) Len() int { return c.Inner.Len() / c.Group }

// Draw compounds one inner path into the coarser frequency.
func (c Compounded) Draw(rng *rand.Rand) Sequence {
	return Annualize(c.Inner.Draw(rng), c.Group)
}
