package refgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Overlap is how a rebuilt series compares with the bundled file it replaces,
// judged step by step: every pair of consecutive dates both carry is one step,
// and a step is REPRODUCED when its return is the old one to the precision the
// files are written at.
//
// Returns rather than levels, because a revised input moves the level of every
// later point of a compounded reconstruction while it moves the RETURN of one
// or two steps only: comparing returns names the months a publisher actually
// revised, and a rebased index compares as identical, which it is.
type Overlap struct {
	Steps      int       // steps both series carry
	Reproduced int       // of which carry the same return
	Worst      float64   // the largest return difference, as a fraction
	WorstAt    time.Time // the date that step ends on
	First      time.Time // the first step that moved, zero when none did
	Last       time.Time // the last step that moved
}

// MinReproduced is the share of its common steps a rebuilt series must
// reproduce before it may replace the bundled file. A publisher's revisions
// touch a few recent months, or a few scattered ones; a change of definition,
// unit or source moves nearly every step, and that is what this bar refuses.
const MinReproduced = 0.95

// CompareSteps measures how far fresh reproduces old, step by step.
func CompareSteps(old, fresh []marketdata.Point) Overlap {
	byDate := make(map[time.Time]float64, len(fresh))
	for _, p := range fresh {
		byDate[p.Date] = p.Close
	}
	var o Overlap
	for i := 1; i < len(old); i++ {
		a, b := old[i-1], old[i]
		fa, okA := byDate[a.Date]
		fb, okB := byDate[b.Date]
		if !okA || !okB || a.Close <= 0 || fa <= 0 {
			continue
		}
		o.Steps++
		was, now := b.Close/a.Close, fb/fa
		// Both files round their levels to sameValue, so a return is only
		// known to that rounding on each of its two ends, on both sides.
		tol := 2 * sameValue * (1 + math.Abs(was)) / a.Close
		d := math.Abs(now - was)
		if d <= tol {
			o.Reproduced++
			continue
		}
		if d > o.Worst {
			o.Worst, o.WorstAt = d, b.Date
		}
		if o.First.IsZero() {
			o.First = b.Date
		}
		o.Last = b.Date
	}
	return o
}

// Share is the fraction of the common steps reproduced (1 when there is none).
func (o Overlap) Share() float64 {
	if o.Steps == 0 {
		return 1
	}
	return float64(o.Reproduced) / float64(o.Steps)
}

// String is the one-line account a generator logs.
func (o Overlap) String() string {
	s := fmt.Sprintf("%d of %d common steps reproduced (%.2f%%)", o.Reproduced, o.Steps, 100*o.Share())
	if o.Reproduced < o.Steps {
		s += fmt.Sprintf(", %d moved between %s and %s, worst %.2e on %s", o.Steps-o.Reproduced,
			o.First.Format(time.DateOnly), o.Last.Format(time.DateOnly), o.Worst, o.WorstAt.Format(time.DateOnly))
	}
	return s
}
