package decumul

import "sort"

// RecoveryBucket is the share of underwater episodes whose recovery took
// Years years.
type RecoveryBucket struct {
	Years int
	Share float64
}

// RecoveryTimeDistribution is the full histogram of years-to-regain a prior
// real high across every underwater episode of every path. Unlike a mean, it
// exposes the psychologically costly tail ("14 years below my initial
// wealth"). An episode still underwater at the horizon is counted at its
// current length. Years always at a fresh high produce no episode.
func (e Ensemble) RecoveryTimeDistribution() []RecoveryBucket {
	counts := map[int]int{}
	total := 0
	for _, p := range e.Paths {
		for _, spell := range underwaterSpells(p.Wealth) {
			counts[spell]++
			total++
		}
	}
	if total == 0 {
		return nil
	}
	out := make([]RecoveryBucket, 0, len(counts))
	for y, c := range counts {
		out = append(out, RecoveryBucket{Years: y, Share: float64(c) / float64(total)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Years < out[j].Years })
	return out
}

// underwaterSpells returns the length in years of each peak-to-recovery
// episode in a wealth path. An episode opens when wealth first drops below
// the running peak and closes (recording its length) when it regains that
// peak; an episode still open at the horizon is recorded at its length.
func underwaterSpells(w []float64) []int {
	var spells []int
	peak := w[0]
	under := 0
	for _, v := range w[1:] {
		if v >= peak {
			if under > 0 {
				spells = append(spells, under)
			}
			peak = v
			under = 0
		} else {
			under++
		}
	}
	if under > 0 { // still underwater at the horizon
		spells = append(spells, under)
	}
	return spells
}
