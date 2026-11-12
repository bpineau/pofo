package web

import (
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
)

// The written-rules switches reach the plan: the Guyton-Klinger band frames
// the year-0 NET rate when gkNet is set, gkCap caps its raises, bufferRefill
// sets the refill threshold and marketDrawdown the drawdown sensor; left at
// zero, none of them touches the plan.
func TestWrittenRulesParams(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 60_000, Years: 40, Guardrails: true,
		SideAnnual: 12_000, SideUntilYear: 14}
	plain := pr.plan()
	if plain.Guard.NetOfIncome || plain.Guard.Cap != 0 || plain.Buffer.RefillThreshold != nil || plain.MarketDrawdown {
		t.Fatalf("defaults moved the plan: %+v, %+v, %v", plain.Guard, plain.Buffer, plain.MarketDrawdown)
	}
	pr.GKNet, pr.GKCap, pr.BufferRefill, pr.MarketDrawdown = true, 1.2, 0.05, true
	p := pr.plan()
	want := decumul.Guardrails{Upper: 0.048 * 1.2, Lower: 0.048 * 0.8, Cut: 0.10, Raise: 0.10,
		Cap: 72_000, NetOfIncome: true}
	if g := p.Guard; abs(g.Upper-want.Upper) > 1e-12 || abs(g.Lower-want.Lower) > 1e-12 || g.Cap != want.Cap || !g.NetOfIncome {
		t.Errorf("guard = %+v, want %+v (band on the 48 000 net of the side income)", g, want)
	}
	if p.Buffer.RefillThreshold == nil || *p.Buffer.RefillThreshold != 0.05 || !p.MarketDrawdown {
		t.Errorf("buffer %+v, market drawdown %v", p.Buffer, p.MarketDrawdown)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
