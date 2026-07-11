package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// TestFirstDecadeReturn checks the annualization on both kernel conventions.
func TestFirstDecadeReturn(t *testing.T) {
	// Annual kernel: 10 years at +5% -> 5%.
	ann := make(scenario.Sequence, 40)
	for i := range ann {
		ann[i] = 0.05
	}
	if got := firstDecadeReturn(ann, 10, 1); math.Abs(got-0.05) > 1e-12 {
		t.Errorf("annual: got %v, want 0.05", got)
	}
	// Monthly kernel: 120 months at the 12th root of 1.05 -> 5%/yr.
	m := math.Pow(1.05, 1.0/12) - 1
	mon := make(scenario.Sequence, 240)
	for i := range mon {
		mon[i] = m
	}
	if got := firstDecadeReturn(mon, 120, 12); math.Abs(got-0.05) > 1e-9 {
		t.Errorf("monthly: got %v, want 0.05", got)
	}
	// A window shorter than a decade still annualizes over what exists.
	if got := firstDecadeReturn(scenario.Sequence{0.10}, 10, 1); math.Abs(got-0.10) > 1e-12 {
		t.Errorf("short: got %v, want 0.10", got)
	}
	if got := firstDecadeReturn(nil, 10, 1); got != 0 {
		t.Errorf("empty: got %v, want 0", got)
	}
	// A -100% year floors at -1 rather than NaN.
	if got := firstDecadeReturn(scenario.Sequence{-1, 0.05}, 2, 1); got != -1 {
		t.Errorf("wipeout: got %v, want -1", got)
	}
}

// TestDecadeBuckets locks the sequence-risk decomposition: paths whose first
// decade is bad must concentrate the ruin.
func TestDecadeBuckets(t *testing.T) {
	// Build a deterministic ensemble by hand: Ret10 spread over [-5%, +7%],
	// with ruin exactly on the paths whose first decade lost money.
	var e Ensemble
	e.Years = 30
	for i := range 100 {
		ret10 := -0.05 + 0.12*float64(i)/99
		p := PathResult{Ret10: ret10, Wealth: []float64{1e6, 2e6}, RuinYear: -1}
		if ret10 < 0 {
			p.Ruined = true
			p.Wealth[1] = 0
		}
		e.Paths = append(e.Paths, p)
	}
	bk := e.DecadeBuckets(5)
	if len(bk) != 5 {
		t.Fatalf("want 5 buckets, got %d", len(bk))
	}
	if bk[0].RuinProb != 1 {
		t.Errorf("worst bucket ruin = %v, want 1 (all negative decades ruined)", bk[0].RuinProb)
	}
	if bk[4].RuinProb != 0 {
		t.Errorf("best bucket ruin = %v, want 0", bk[4].RuinProb)
	}
	if bk[0].LoRet >= bk[4].HiRet {
		t.Errorf("buckets not ordered: %v vs %v", bk[0].LoRet, bk[4].HiRet)
	}
	if got := bk[0].Paths + bk[1].Paths + bk[2].Paths + bk[3].Paths + bk[4].Paths; got != 100 {
		t.Errorf("bucket sizes sum to %d, want 100", got)
	}
	if bk[4].TerminalP50 != 2e6 {
		t.Errorf("best bucket median terminal = %v, want 2e6", bk[4].TerminalP50)
	}
	// Too few paths: nil.
	if DecadeBuckets := (Ensemble{Paths: e.Paths[:3]}).DecadeBuckets(5); DecadeBuckets != nil {
		t.Errorf("want nil for tiny ensembles")
	}
}

// TestRunPathRecordsRet10 checks both kernels populate the field.
func TestRunPathRecordsRet10(t *testing.T) {
	p := Plan{Capital: 1e6, NeedAnnual: 30000, Years: 20}
	seq := make(scenario.Sequence, 20)
	for i := range seq {
		seq[i] = 0.04
	}
	if got := p.RunPath(seq).Ret10; math.Abs(got-0.04) > 1e-12 {
		t.Errorf("annual kernel Ret10 = %v, want 0.04", got)
	}
	mseq := make(scenario.Sequence, 240)
	rm := math.Pow(1.04, 1.0/12) - 1
	for i := range mseq {
		mseq[i] = rm
	}
	if got := p.RunPathMonthly(mseq).Ret10; math.Abs(got-0.04) > 1e-9 {
		t.Errorf("monthly kernel Ret10 = %v, want 0.04", got)
	}
}
