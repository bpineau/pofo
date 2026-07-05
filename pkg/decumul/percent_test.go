package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// A percentage-of-portfolio (VPW) rule can never run out: spending is always a
// fraction of what remains. Even a brutal return path leaves the capital
// positive and every year funded.
func TestPercentRuleNeverRuins(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 40, Percent: 0.05}
	// A long run of deep losses then flat: a fixed rule would be wiped out.
	seq := make(scenario.Sequence, 40)
	for i := range seq {
		seq[i] = -0.30
	}
	res := p.RunPath(seq)
	if res.Ruined {
		t.Errorf("VPW must not ruin, ruined at year %d", res.RuinYear)
	}
	for k, w := range res.Wealth {
		if w < 0 {
			t.Errorf("wealth went negative at year %d: %.2f", k, w)
		}
	}
}

// Spending tracks the portfolio: after a crash the VPW household spends
// strictly less than after a boom, the defining trade-off of the rule.
func TestPercentRuleSpendingTracksWealth(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3, Percent: 0.05}
	boom := p.RunPath(scenario.Sequence{0.3, 0.3, 0.3})
	bust := p.RunPath(scenario.Sequence{-0.3, -0.3, -0.3})
	if !(bust.Spend[2] < boom.Spend[2]) {
		t.Errorf("VPW spending should fall after losses: bust=%.0f boom=%.0f", bust.Spend[2], boom.Spend[2])
	}
}

// Against a fixed rule, VPW converts ruin risk into spending volatility: over a
// Monte-Carlo it should ruin far less but deliver a more variable standard of
// living.
func TestPercentRuleTradesRuinForVolatility(t *testing.T) {
	src := scenario.ParametricSource{Mu: 0.05, Sigma: 0.18, Df: 5, Periods: 40}
	base := Plan{Capital: 1_000_000, NeedAnnual: 50_000, Years: 40, Source: src}
	fixed := base.Simulate(4000, 8, 1)
	vpw := base
	vpw.Percent = 0.05
	vpwEns := vpw.Simulate(4000, 8, 1)

	if !(vpwEns.RuinProb() < fixed.RuinProb()) {
		t.Errorf("VPW ruin %.3f should be below fixed %.3f", vpwEns.RuinProb(), fixed.RuinProb())
	}
	if cv := vpwEns.SpendCV(); cv <= fixed.SpendCV() {
		t.Errorf("VPW spending CV %.3f should exceed fixed %.3f", cv, fixed.SpendCV())
	}
}
