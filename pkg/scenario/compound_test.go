package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestAnnualize(t *testing.T) {
	if got := Annualize(make(Sequence, 12), 12); len(got) != 1 || math.Abs(got[0]) > 1e-12 {
		t.Errorf("12 zeros -> %v, want [0]", got)
	}
	// two years of constant +1%/month: (1.01^12 - 1) each.
	monthly := make(Sequence, 24)
	for i := range monthly {
		monthly[i] = 0.01
	}
	got := Annualize(monthly, 12)
	want := math.Pow(1.01, 12) - 1
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for i, v := range got {
		if math.Abs(v-want) > 1e-12 {
			t.Errorf("year %d = %.6f, want %.6f", i, v, want)
		}
	}
}

// stubMonthly is a Source that always returns a fixed monthly sequence.
type stubMonthly struct{ seq Sequence }

func (s stubMonthly) Draw(*rand.Rand) Sequence { return s.seq }
func (s stubMonthly) Len() int                 { return len(s.seq) }

func TestCompounded(t *testing.T) {
	inner := stubMonthly{seq: Sequence{0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01}}
	c := Compounded{Inner: inner, Group: 12}
	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
	got := c.Draw(rand.New(rand.NewPCG(1, 1)))
	want := math.Pow(1.01, 12) - 1
	if len(got) != 1 || math.Abs(got[0]-want) > 1e-12 {
		t.Errorf("Draw = %v, want [%.6f]", got, want)
	}
}
