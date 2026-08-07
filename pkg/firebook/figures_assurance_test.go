package firebook

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// The month indices the plate annotates, and that the article's prose names.
const (
	idxIrmaMaria = 46  // 2017-09
	idxCovid     = 76  // 2020-03
	idxIan       = 106 // 2022-09
)

// The two rows must describe the same months, otherwise the plate compares
// dates that are not aligned and its whole argument collapses.
func TestSinistresRowsShareTheSameMonths(t *testing.T) {
	if len(ilsMonthlyBp) != len(eqMonthlyBp) {
		t.Fatalf("rows of different lengths: %d cat bond months against %d equity months",
			len(ilsMonthlyBp), len(eqMonthlyBp))
	}
	// 2013-11 to 2025-12 inclusive.
	if want := 146; len(ilsMonthlyBp) != want {
		t.Errorf("%d months, expected %d (2013-11 to 2025-12)", len(ilsMonthlyBp), want)
	}
}

// The plate annotates the two catastrophe months as the sleeve's worst, and
// the article says so in prose. If the data ever moves, the labels lie.
func TestSinistresWorstMonthsAreTheAnnotatedOnes(t *testing.T) {
	order := rankAscending(ilsMonthlyBp)
	if order[0] != idxIan || order[1] != idxIrmaMaria {
		t.Errorf("worst cat bond months are %v, expected Ian (%d) then Irma/Maria (%d)",
			order[:2], idxIan, idxIrmaMaria)
	}
	// Irma and Maria cost the sleeve nearly 6 %, in a month equities gained.
	if ilsMonthlyBp[idxIrmaMaria] > -500 || eqMonthlyBp[idxIrmaMaria] <= 0 {
		t.Errorf("2017-09: cat bonds %d bp, equities %d bp; the plate claims a hurricane loss in a good market month",
			ilsMonthlyBp[idxIrmaMaria], eqMonthlyBp[idxIrmaMaria])
	}
}

// The caption claims exactly two months out of 146 see both rows lose
// heavily: March 2020 (forced sellers) and September 2022 (Ian on one side,
// the rate shock on the other). That count is the plate's honest exception,
// so it is guarded as such.
func TestSinistresSharedAccidentsAreTheTwoNamedMonths(t *testing.T) {
	var shared []int
	for i := range ilsMonthlyBp {
		if ilsMonthlyBp[i] <= -150 && eqMonthlyBp[i] <= -500 {
			shared = append(shared, i)
		}
	}
	if len(shared) != 2 || shared[0] != idxCovid || shared[1] != idxIan {
		t.Errorf("months where both rows fall hard: %v, expected March 2020 (%d) and September 2022 (%d)",
			shared, idxCovid, idxIan)
	}
}

// The cash ladder plate is an ordering claim: every extra basis point of
// yield is paid for with a deeper worst drawdown. Both series must therefore
// rise together, and the printed numbers must reach the SVG.
func TestEchelleDuCashIsAMonotonicLadder(t *testing.T) {
	rungs := []cashRung{
		{"monétaire (ESTR)", "XEON, l'étalon", 0, -8},
		{"obligataire ultra-court", "ERNX, crédit investment grade court", 27, -20},
		{"CLO AAA en euro", "JCL0, depuis 2024-12", 127, -70},
		{"CLO AAA en dollar", "JAAA, depuis 2020-10", 133, -260},
	}
	for i := 1; i < len(rungs); i++ {
		if rungs[i].gainBp <= rungs[i-1].gainBp {
			t.Errorf("rung %q does not pay more than %q", rungs[i].name, rungs[i-1].name)
		}
		if rungs[i].lossBp >= rungs[i-1].lossBp {
			t.Errorf("rung %q does not risk more than %q", rungs[i].name, rungs[i-1].name)
		}
	}
	svg := figEchelleDuCash()
	for _, r := range rungs {
		if !strings.Contains(svg, r.name) {
			t.Errorf("the plate does not name the %q rung", r.name)
		}
		if r.gainBp > 0 && !strings.Contains(svg, fmt.Sprintf("+%d bp", r.gainBp)) {
			t.Errorf("the plate does not print +%d bp", r.gainBp)
		}
		if !strings.Contains(svg, fmt.Sprintf("%d bp", r.lossBp)) {
			t.Errorf("the plate does not print %d bp", r.lossBp)
		}
	}
}

// rankAscending returns the month indices ordered from the worst return up.
func rankAscending(v []int) []int {
	idx := make([]int, len(v))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return v[idx[a]] < v[idx[b]] })
	return idx
}
