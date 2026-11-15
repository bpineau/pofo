package web

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/scenario"
)

// ModelStat is one return model's outcome in the comparison strip. It separates
// epistemic uncertainty (which model) from the aleatory Monte-Carlo noise inside
// each model. SafeWR/SafeSpend are the withdrawal that meets the target ruin
// under this model; Ruin/MedianWealth are evaluated at the user's planned spend.
type ModelStat struct {
	Name         string              `json:"name"`         // the return model
	Ruin         float64             `json:"ruin"`         // fraction, at the planned spend
	SafeWR       float64             `json:"safeWR"`       // fraction, safe spend / capital
	SafeSpend    float64             `json:"safeSpend"`    // euros/yr meeting the target ruin
	MedianWealth float64             `json:"medianWealth"` // median terminal real wealth
	Help         string              `json:"help"`         // plain-language hover explanation
	SolveStatus  decumul.SolveStatus `json:"solveStatus"`  // feasibility or search boundary
	SolveNote    string              `json:"solveNote"`    // limitation when no crossing was bracketed
}

// MarshalJSON represents unavailable safe-spending figures as null.
func (m ModelStat) MarshalJSON() ([]byte, error) {
	type plain ModelStat
	number := func(v float64) *float64 {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil
		}
		return &v
	}
	return json.Marshal(struct {
		plain
		SafeWR    *float64 `json:"safeWR"`
		SafeSpend *float64 `json:"safeSpend"`
	}{plain(m), number(m.SafeWR), number(m.SafeSpend)})
}

// ModelsResult is the multi-model comparison: the per-model stats, the target
// ruin they were solved against, a single confidence badge about the data
// backing the historical models, and the central-case verdict sentence.
type ModelsResult struct {
	Models     []ModelStat `json:"models"`     // one per return model, in strip order
	TargetRuin float64     `json:"targetRuin"` // the ruin target solved against (fraction)
	Confidence string      `json:"confidence"` // HIGH | MEDIUM | LOW
	ConfNote   string      `json:"confNote"`   // one-line reason
	Verdict    string      `json:"verdict"`    // central-case headline
}

// namedSource pairs a return model with its label and hover help.
type namedSource struct {
	name, help string
	source     scenario.Source
}

// Models evaluates the return models side by side for one parameter set, the
// core of the rewrite: instead of a single hypersensitive ruin figure it shows
// the plausible range across calibrated models. The synthetic family (Student-t,
// the mean-preserving Sequence stress, the Broad-sample prior and the Lost-decade
// tail) is always present; a panel adds the data-driven Historical windows and
// Block-bootstrap columns.
func Models(pr Params, panel *scenario.Panel) ModelsResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	base := pr.plan()
	base.Monthly = false // the strip compares annual kernels for speed and parity

	sources := modelSources(pr, panel)
	res := ModelsResult{TargetRuin: target, Models: make([]ModelStat, len(sources))}
	concurrently(len(sources), func(i int) {
		res.Models[i] = evalModel(base, sources[i], pr.Capital, target, pr.NPaths)
	})
	res.Confidence, res.ConfNote = confidence(pr, panel)
	res.Verdict = verdict(res.Models, pr, target)
	return res
}

// minPanelMonths is the least common history the data-driven columns accept:
// below two years, resampling a handful of months manufactures absurd
// full-length retirements (a few lucky months compounded 40 years), so the
// Historical/Block-bootstrap columns are dropped rather than shown wrong.
const minPanelMonths = fire.MinPanelMonths

// modelSources builds the ordered model set: the data-driven columns first (when
// a panel and weights are available and the cohorts model has enough windows),
// then the synthetic family from optimistic to conservative.
func modelSources(pr Params, panel *scenario.Panel) []namedSource {
	if panel != nil && panel.Periods() < minPanelMonths {
		panel = nil
	}
	market := fire.MarketSpec{Mu: pr.Mu, Sigma: pr.Sigma, Nu: pr.Df, Panel: panel, Weights: normalize(pr.Weights), CAPE: pr.CapeAdjust}
	ids := []string{fire.Central, fire.Stress, fire.BroadSample, fire.LostDecade}
	if panel != nil {
		ids = append([]string{fire.Historical, fire.Bootstrap}, ids...)
	}
	var out []namedSource
	for _, id := range ids {
		model, err := fire.PrepareModel(id, market, pr.Years)
		if err != nil {
			continue
		} // the web intentionally omits unavailable historical columns
		if id == fire.Central && pr.Glidepath {
			model.Source = centralSource(pr, model.EffectiveMu, model.EffectiveSigma, model.EffectiveNu, pr.Years)
		}
		out = append(out, namedSource{model.Name, model.Description, model.Source})
	}
	return out
}

// detailSource is the return model behind every detail section (fans aside,
// which always show the four lenses): the strip column the user clicked
// (Params.Central), or the calibrated central case. Data-driven selections
// ("hist", "boot") fall back to the central case when no adequate panel
// exists; the central default honours the glidepath toggle via centralSource.
func (pr Params) detailSource(panel *scenario.Panel, years int) scenario.Source {
	if panel != nil && panel.Periods() < minPanelMonths {
		panel = nil
	}
	cMu, cSigma, cDf := centralParams(pr, panel)
	months := years * 12
	w := pr.Weights
	if panel != nil && w == nil {
		w = panel.Weights
	}
	switch pr.central() {
	case "stress":
		return scenario.NewMarkovRegime(cMu, cSigma, cDf, years)
	case "broad":
		return broadSampleSource(years)
	case "lost":
		return scenario.NewLostDecadeRegime(cMu, cSigma, cDf, years)
	case "hist":
		if panel != nil {
			c := scenario.HistoricalCohorts{Panel: *panel, Weights: w, Periods: months}
			if c.Count() > 0 {
				return scenario.Compounded{Inner: c, Group: 12}
			}
		}
	case "boot":
		if panel != nil {
			return scenario.Compounded{
				Inner: scenario.StationaryBootstrap{Panel: *panel, Weights: w, MeanBlock: 24, Periods: months},
				Group: 12}
		}
	}
	return centralSource(pr, cMu, cSigma, cDf, years)
}

// centralSource is the central-case return model shared by the verdict and the
// detail views: the calibrated Student-t, or a rising-equity glidepath (bond
// tent) blending the central equity assumptions with a fixed bond sleeve when
// the glidepath option is on. The glidepath keeps the sequence-risk danger zone
// (the first years) bond-heavy, climbing to mostly equity later. Note it trades
// return for sequence protection: with a wide equity-bond gap the drag can raise
// ruin, which is itself a useful thing for the user to see (Cederburg et al.).
func centralSource(pr Params, mu, sigma, df float64, periods int) scenario.Source {
	if pr.Glidepath {
		return scenario.Glidepath{
			EquityMu: mu, EquitySigma: sigma, Df: df,
			BondMu: 0.015, BondSigma: 0.06, Corr: 0.15,
			StartEquity: 0.30, EndEquity: 0.75, Periods: periods,
		}
	}
	return scenario.ParametricSource{Mu: mu, Sigma: sigma, Df: df, Periods: periods}
}

// centralParams blends the fitted (slider) parameters toward the broad-sample
// prior by the history shortfall: with a panel shorter than the horizon, the
// central planning models lean toward the prior, because a short favourable
// window cannot reveal the long-horizon tail and sequence risk. The blend is
// capped at half so the data is never fully discarded; with no panel or ample
// history it returns the fitted values unchanged. This pulls the rosy short-run
// fit toward a believable middle rather than leaving the central case optimistic.
func centralParams(pr Params, panel *scenario.Panel) (mu, sigma, df float64) {
	mu, sigma, df, _ = fire.CentralParameters(fire.MarketSpec{Mu: pr.Mu, Sigma: pr.Sigma, Nu: pr.Df, Panel: panel, CAPE: pr.CapeAdjust}, pr.Years)
	return
}

// evalModel runs one model: ruin and median wealth at the planned spend (under
// the user's actual policy), and the safe withdrawal that meets the target ruin.
// The safe-withdrawal solve uses the fixed rule (fixedRule strips flex and
// guardrails): the conventional definition of a safe withdrawal rate, and the
// only one that is monotonic in the withdrawal, so the bisection is well
// defined. The shared seed keeps the figures comparable across models.
func evalModel(base decumul.Plan, ns namedSource, capital, target float64, nPaths int) ModelStat {
	const seed = uint64(7)
	p := base
	p.Source = ns.source

	draws := p.Draw(nPaths, simWorkers, seed)
	o := p.SimulateOn(draws, simWorkers).Outcome()
	solve, note := solveOn(fixedRule(p), target, decumul.WithdrawalAxis(0, capital*0.15), draws)
	return ModelStat{
		Name: ns.name, Help: ns.help + " " + note,
		Ruin: o.RuinProb, MedianWealth: o.TerminalP50,
		SafeSpend: solve.Solution, SafeWR: solve.Solution / capital,
		SolveStatus: solve.Status, SolveNote: note,
	}
}

// fixedRule strips every adaptive spending rule (flex cut, guardrails both
// flavours, ratchet, percent of portfolio, ABW, bounded percent) so the plan withdraws a fixed real
// amount. A safe-withdrawal solve must run on the fixed rule: the adaptive
// rules rebase spending on wealth or on the initial withdrawal rate, which
// makes ruin non-monotonic (or identically zero) in the withdrawal and the
// bisection ill-defined.
func fixedRule(p decumul.Plan) decumul.Plan {
	p.Flex = decumul.FlexRule{}
	p.Guard = decumul.Guardrails{}
	p.Ratchet = decumul.Ratchet{}
	p.RiskGuard = decumul.RiskGuardrails{}
	p.Percent = 0
	p.Amortize = false
	p.Bounded = decumul.BoundedPct{}
	return p
}

// confidence rates how much the data-backed models can be trusted at the chosen
// horizon: a fund history shorter than the horizon holds no independent
// full-length retirement, so the historical figures are optimistic about
// long-horizon sequence risk.
func confidence(pr Params, panel *scenario.Panel) (level, note string) {
	if panel == nil {
		return "MEDIUM", "Parametric models only (no portfolio loaded); the figures are assumption-driven, not data-backed."
	}
	histYears := panel.Periods() / 12
	switch {
	case histYears >= pr.Years:
		return "HIGH", fmt.Sprintf("Fund history %dy covers the %dy horizon.", histYears, pr.Years)
	case histYears*3 >= pr.Years:
		return "MEDIUM", fmt.Sprintf("History %dy vs %dy horizon: the history-based columns (Historical, Block bootstrap) reflect one favourable window and read optimistic; Conservative is a deliberate broad-sample worst case. Plan near the central columns, between the two.", histYears, pr.Years)
	default:
		return "LOW", fmt.Sprintf("History %dy is short vs the %dy horizon: the history-based columns are unreliable here. Weight the central (Student-t/Regime) and Conservative columns.", histYears, pr.Years)
	}
}

// verdict is the central-case headline: the safe spend (in euros and as a rate)
// from the calibrated Student-t model, against the user's plan. The Regime and
// Conservative columns give the sequence-stress and pessimistic downside.
func verdict(models []ModelStat, pr Params, target float64) string {
	var central ModelStat
	for _, m := range models {
		if m.Name == "Student-t" {
			central = m
		}
	}
	if central.Name == "" || pr.Capital == 0 {
		return ""
	}
	if central.SolveNote != "" {
		return central.SolveNote
	}
	return fmt.Sprintf("Safe spend ≈ %.0f k€/yr (%.1f%%) at %.0f%% success · you plan %.0f k€ (%.1f%%)",
		central.SafeSpend/1000, central.SafeWR*100, (1-target)*100,
		pr.NeedAnnual/1000, pr.NeedAnnual/pr.Capital*100)
}
