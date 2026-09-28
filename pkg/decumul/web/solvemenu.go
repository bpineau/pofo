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
	Lever string `json:"lever"` // what is moved (spend, capital, buffer, flex...)
	Text  string `json:"text"`  // the instruction, for a human
	OK    bool   `json:"ok"`    // the target is reachable through this lever alone
}

// SolverMenu answers "what do I need to keep ruin at my target?" per controllable
// lever, under the calibrated central (Student-t) model: the menu of equivalent
// ways to get there, rather than a single number. It evaluates at the user's
// planned spend, so the flex and buffer options keep that spend and change
// something else instead.
//
// When the plan already meets the target (Met), the levers to reach it would all
// collapse to "no change" (a 0% cut, a 0-year buffer), which reads as nonsense;
// the menu then reports the headroom instead: how much more could be spent while
// still meeting the target.
type SolverMenu struct {
	TargetRuin  float64        `json:"targetRuin"`  // the ruin target (fraction)
	CurrentRuin float64        `json:"currentRuin"` // ruin at the planned spend (fraction)
	Met         bool           `json:"met"`         // the plan already meets the target
	Options     []SolverOption `json:"options"`     // one per lever
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
	base.Source = pr.detailSource(panel, pr.Years)

	// Every figure below varies a spending rule, the buffer or a cut, never the
	// return model, so they all read one set of drawn paths.
	draws := base.Draw(pr.NPaths, simWorkers, seed)
	menu := SolverMenu{TargetRuin: target}
	menu.CurrentRuin = base.RuinProbOn(draws, simWorkers)
	menu.Met = menu.CurrentRuin <= target

	// The safe spend at the target on the fixed rule (no flex/guardrails), the
	// monotonic and conventional safe withdrawal. Above the plan when the target
	// is already met (headroom), below it when the plan is too aggressive. When
	// the target is missed, the two other levers are solved alongside it.
	var safe float64
	var flex, buffer SolverOption
	flexBase := base
	flexBase.Flex.Threshold = 0.20
	levers := []func(){
		func() {
			safe = fixedRule(base).SolveOn(target, decumul.WithdrawalAxis(0, pr.Capital*0.15), draws, simWorkers)
		},
		// Temporary downturn cut (flex): keep the spend, accept a reversible cut.
		func() {
			cut := flexBase.SolveOn(target, decumul.FlexCutAxis(0, 0.60), draws, simWorkers)
			flex = flexOption(flexBase, cut, target, draws)
		},
		// Buffer: keep the spend, hold N years of cash (scan; ruin is non-monotonic).
		func() { buffer = bufferOption(base, target, draws) },
	}
	if menu.Met {
		levers = levers[:1]
	}
	concurrently(len(levers), func(i int) { levers[i]() })

	// Met: the reach-the-target levers would all read "no change" (a 0% cut, a
	// 0-year buffer), so report the spending headroom instead of a nonsense menu.
	if menu.Met {
		menu.Options = append(menu.Options, SolverOption{
			Lever: "Room to spare", OK: true,
			Text: fmt.Sprintf("You could spend up to %.0f k€/yr (%.1f%%) and still meet the target (you plan %.0f k€, %.1f%%)",
				safe/1000, safe/pr.Capital*100, pr.NeedAnnual/1000, pr.NeedAnnual/pr.Capital*100),
		})
		return menu
	}

	// Missed: the equivalent ways to bring ruin down to the target.
	menu.Options = append(menu.Options, SolverOption{
		Lever: "Spend less", OK: true,
		Text: fmt.Sprintf("Spend down to %.0f k€/yr (%.1f%%) instead of %.0f k€ (%.1f%%)",
			safe/1000, safe/pr.Capital*100, pr.NeedAnnual/1000, pr.NeedAnnual/pr.Capital*100),
	})

	menu.Options = append(menu.Options, flex, buffer)
	return menu
}

// flexOption describes the smallest downturn cut reaching the target, checking
// reachability at the solved depth.
func flexOption(p decumul.Plan, cut, target float64, draws decumul.Draws) SolverOption {
	q := p
	q.Flex.Cut = cut
	if q.RuinProbOn(draws, simWorkers) > target+0.01 {
		return SolverOption{Lever: "Cut in downturns", OK: false,
			Text: "Even a 60% downturn spending cut does not reach the target alone"}
	}
	return SolverOption{Lever: "Cut in downturns", OK: true,
		Text: fmt.Sprintf("Keep the spend but accept up to a %.0f%% cut in downturns (drawdowns over 20%%)", cut*100)}
}

// bufferOption finds the smallest cash buffer (in years) that reaches the target
// at the current spend, or reports it is not reachable by buffer alone.
func bufferOption(p decumul.Plan, target float64, draws decumul.Draws) SolverOption {
	for _, years := range bufferCandidates {
		q := p
		q.Buffer.Years = years
		if q.RuinProbOn(draws, simWorkers) <= target {
			return SolverOption{Lever: "Cash buffer", OK: true,
				Text: fmt.Sprintf("Keep the spend but hold a %.0f-year cash buffer", years)}
		}
	}
	return SolverOption{Lever: "Cash buffer", OK: false,
		Text: "A cash buffer up to 10 years does not reach the target alone"}
}
