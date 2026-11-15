package web

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

func TestRiskSafeRateOneYearNotClipped(t *testing.T) {
	pr := Params{Mu: 0.04, Sigma: 0.15, Df: 5}
	if got := pr.safeRateAt(1, 0.05); math.Abs(got-1) > 1e-5 {
		t.Fatalf("one-year safe rate %v, want 1", got)
	}
}

func TestModelBoundIsVisible(t *testing.T) {
	p := decumul.Plan{Capital: 1e6, NeedAnnual: 40000, Years: 1}
	m := evalModel(p, namedSource{"Student-t", "test", scenario.ParametricSource{Periods: 1}}, 1e6, 0.05, 100)
	if !strings.Contains(m.Help, "bound") {
		t.Fatalf("saturated safe spend lacks limitation: %+v", m)
	}
	v := verdict([]ModelStat{m}, Params{Capital: 1e6, NeedAnnual: 40000}, 0.05)
	if !strings.Contains(v, "bound") {
		t.Fatalf("unqualified verdict: %s", v)
	}
}
