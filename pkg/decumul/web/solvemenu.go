package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// SolverOption is one controllable way to reach the target ruin: the lever, a
// human-readable instruction, and whether the target is reachable through it
// alone.
type SolverOption struct {
	Lever string `json:"lever"`
	Text  string `json:"text"`
	OK    bool   `json:"ok"`
}

// SolverMenu answers "what do I need to keep ruin at my target?" per controllable
// lever, under the calibrated central (Student-t) model: the menu of equivalent
// ways to get there, rather than a single number. It evaluates at the user's
// planned spend, so the flex and buffer options keep that spend and change
// something else instead.
type SolverMenu struct {
	TargetRuin  float64        `json:"targetRuin"`
	CurrentRuin float64        `json:"currentRuin"`
	Options     []SolverOption `json:"options"`
}

// bufferCandidates are the buffer-years tried when solving the buffer lever.
var bufferCandidates = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10}

// SolveMenu computes the per-lever menu for the central model. Capital, spend
// and allocation stay as the user set them except for the one lever each option
// varies.
func SolveMenu(pr Params, panel *scenario.Panel) SolverMenu {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	const seed = uint64(7)

	base := pr.plan()
	base.Monthly = false
	base.Source = scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years}

	menu := SolverMenu{TargetRuin: target}
	menu.CurrentRuin = base.Simulate(pr.NPaths, simWorkers, seed).RuinProb()

	// Withdrawal: the safe spend at the target (always reachable by spending less).
	safe := base.Solve(target, decumul.WithdrawalAxis(0, pr.Capital*0.15), pr.NPaths, simWorkers, seed)
	menu.Options = append(menu.Options, SolverOption{
		Lever: "Spend less", OK: true,
		Text: fmt.Sprintf("Spend up to %.0f k€/yr (%.1f%%) instead of %.0f k€ (%.1f%%)",
			safe/1000, safe/pr.Capital*100, pr.NeedAnnual/1000, pr.NeedAnnual/pr.Capital*100),
	})

	// Temporary downturn cut (flex): keep the spend, accept a reversible cut.
	flexBase := base
	flexBase.Flex.Threshold = 0.20
	cut := flexBase.Solve(target, decumul.FlexCutAxis(0, 0.60), pr.NPaths, simWorkers, seed)
	menu.Options = append(menu.Options, flexOption(flexBase, cut, target, pr.NPaths, seed))

	// Buffer: keep the spend, hold N years of cash (scan; ruin is non-monotonic).
	menu.Options = append(menu.Options, bufferOption(base, target, pr.NPaths, seed))

	return menu
}

// flexOption describes the smallest downturn cut reaching the target, checking
// reachability at the solved depth.
func flexOption(p decumul.Plan, cut, target float64, nPaths int, seed uint64) SolverOption {
	q := p
	q.Flex.Cut = cut
	if q.Simulate(nPaths, simWorkers, seed).RuinProb() > target+0.01 {
		return SolverOption{Lever: "Cut in downturns", OK: false,
			Text: "Even a 60% downturn spending cut does not reach the target alone"}
	}
	return SolverOption{Lever: "Cut in downturns", OK: true,
		Text: fmt.Sprintf("Keep the spend but accept up to a %.0f%% cut in downturns (drawdowns over 20%%)", cut*100)}
}

// bufferOption finds the smallest cash buffer (in years) that reaches the target
// at the current spend, or reports it is not reachable by buffer alone.
func bufferOption(p decumul.Plan, target float64, nPaths int, seed uint64) SolverOption {
	for _, years := range bufferCandidates {
		q := p
		q.Buffer.Years = years
		if q.Simulate(nPaths, simWorkers, seed).RuinProb() <= target {
			return SolverOption{Lever: "Cash buffer", OK: true,
				Text: fmt.Sprintf("Keep the spend but hold a %.0f-year cash buffer", years)}
		}
	}
	return SolverOption{Lever: "Cash buffer", OK: false,
		Text: "A cash buffer up to 10 years does not reach the target alone"}
}
