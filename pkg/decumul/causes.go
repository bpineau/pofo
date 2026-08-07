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

// RuinShapes decomposes the ruined paths by the SHAPE of their wealth
// trajectory, a sharper cause attribution than RuinTiming's when-it-failed
// proxy (a crash at year 3 and a grind can both deplete at year 25; the
// trajectory tells them apart):
//
//   - Crash: wealth halved within the first ten years, the classic
//     sequence-of-returns disaster (a bad opening decade the plan never
//     recovers from).
//   - Grind: the portfolio never meaningfully grew (peak below 1.25x the
//     start) and eroded to zero, the lost-decade / flat-real-returns
//     failure.
//   - Longevity: the plan prospered (peak at or above 1.25x the start) and
//     still ran out: the money worked, the retirement outlasted it.
//
// Shares are of the ruined paths, summing to ~1; Ruined counts them.
type RuinShapes struct {
	Crash     float64
	Grind     float64
	Longevity float64
	Ruined    int
}

// RuinShapes classifies every ruined path by trajectory shape. Returns the
// zero value when nothing ruined.
func (e Ensemble) RuinShapes() RuinShapes {
	var crash, grind, longevity, total int
	for _, p := range e.Paths {
		if !p.Ruined || len(p.Wealth) == 0 || p.Wealth[0] <= 0 {
			continue
		}
		total++
		start := p.Wealth[0]
		peak := 0.0
		halvedAt := -1
		for k, w := range p.Wealth {
			if w > peak {
				peak = w
			}
			if halvedAt < 0 && w < 0.5*start {
				halvedAt = k
			}
		}
		switch {
		case halvedAt >= 0 && halvedAt <= 10:
			crash++
		case peak >= 1.25*start:
			longevity++
		default:
			grind++
		}
	}
	if total == 0 {
		return RuinShapes{}
	}
	n := float64(total)
	return RuinShapes{
		Crash:     float64(crash) / n,
		Grind:     float64(grind) / n,
		Longevity: float64(longevity) / n,
		Ruined:    total,
	}
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
