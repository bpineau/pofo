package decumul

// RuinTiming decomposes the paths that ran out of money by WHEN they failed,
// across thirds of the horizon. The timing is a legible proxy for the cause: a
// path depleted in the first third can only have hit a bad early return sequence
// (a crash right after retiring, the classic sequence-of-returns risk); the
// middle third points to a prolonged low-return grind (a "lost decade"); the
// last third means the money simply outlasted most of retirement before running
// dry (a longevity / duration failure). Shares are of the ruined paths, so they
// sum to ~1; Ruined is the count they decompose.
type RuinTiming struct {
	Early  float64 // ruined in the first third: an early crash / bad sequence
	Mid    float64 // ruined in the middle third: a lost decade
	Late   float64 // ruined in the last third: outlived the money
	Ruined int     // number of ruined paths decomposed
}

// RuinTiming classifies every ruined path by the third of the horizon in which
// it failed. Returns the zero value when nothing ruined.
func (e Ensemble) RuinTiming() RuinTiming {
	if e.Years <= 0 {
		return RuinTiming{}
	}
	third := float64(e.Years) / 3
	var early, mid, late, total int
	for _, p := range e.Paths {
		if !p.Ruined || p.RuinYear < 0 {
			continue
		}
		total++
		switch y := float64(p.RuinYear); {
		case y < third:
			early++
		case y >= 2*third:
			late++
		default:
			mid++
		}
	}
	if total == 0 {
		return RuinTiming{}
	}
	n := float64(total)
	return RuinTiming{
		Early:  float64(early) / n,
		Mid:    float64(mid) / n,
		Late:   float64(late) / n,
		Ruined: total,
	}
}
