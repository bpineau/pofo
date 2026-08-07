package web

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Conservative broad-sample prior: a cautious, forward-looking world-equity real
// model used regardless of the portfolio's own rosy history. The pessimism is in
// the tails, the volatility and the sequence clustering, NOT in an implausibly
// low average return: consMu is the arithmetic mean of a regime whose blended
// real geometric return lands near ~3.5% (a forward haircut on the ~4.5-5% DMS
// world-equity history), with fat tails and persistent drawdowns. Sources: DMS
// world real equity, broad-sample SWR evidence (Anarkulova, Cederburg &
// O'Doherty). It deliberately does not assume equities barely grow.
const (
	consMu    = 0.045 // arithmetic; blended geometric ~3.5% real under the regime
	consSigma = 0.13
	consDf    = 4
)

// ModelStat is one return model's outcome in the comparison strip. It separates
// epistemic uncertainty (which model) from the aleatory Monte-Carlo noise inside
// each model. SafeWR/SafeSpend are the withdrawal that meets the target ruin
// under this model; Ruin/MedianWealth are evaluated at the user's planned spend.
type ModelStat struct {
	Name         string  `json:"name"`
	Ruin         float64 `json:"ruin"`         // fraction, at the planned spend
	SafeWR       float64 `json:"safeWR"`       // fraction, safe spend / capital
	SafeSpend    float64 `json:"safeSpend"`    // euros/yr meeting the target ruin
	MedianWealth float64 `json:"medianWealth"` // median terminal real wealth
	Help         string  `json:"help"`         // plain-language hover explanation
}

// ModelsResult is the multi-model comparison: the per-model stats, the target
// ruin they were solved against, a single confidence badge about the data
// backing the historical models, and the central-case verdict sentence.
type ModelsResult struct {
	Models     []ModelStat `json:"models"`
	TargetRuin float64     `json:"targetRuin"`
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

	res := ModelsResult{TargetRuin: target}
	for _, ns := range modelSources(pr, panel) {
		res.Models = append(res.Models, evalModel(base, ns, pr.Capital, target, pr.NPaths))
	}
	res.Confidence, res.ConfNote = confidence(pr, panel)
	res.Verdict = verdict(res.Models, pr, target)
	return res
}

// modelSources builds the ordered model set: the data-driven columns first (when
// a panel and weights are available and the cohorts model has enough windows),
// then the synthetic family from optimistic to conservative.
func modelSources(pr Params, panel *scenario.Panel) []namedSource {
	var out []namedSource
	if panel != nil {
		w := pr.Weights
		if w == nil {
			w = panel.Weights
		}
		months := pr.Years * 12
		cohorts := scenario.HistoricalCohorts{Panel: *panel, Weights: w, Periods: months}
		if cohorts.Count() > 0 {
			out = append(out, namedSource{"Historical windows",
				"Replays this portfolio's actual return sequences (every historical start date, no resampling). Honest but limited: a short history holds few independent retirements, so it tends to look optimistic for long horizons.",
				scenario.Compounded{Inner: cohorts, Group: 12}})
		}
		out = append(out, namedSource{"Block bootstrap",
			"Resamples multi-year blocks of this portfolio's real returns, preserving clustered bear markets and cross-asset correlation. Manufactures many full-length retirements from a short history, but stays anchored to that one favourable window.",
			scenario.Compounded{Inner: scenario.StationaryBootstrap{Panel: *panel, Weights: w, MeanBlock: 24, Periods: months}, Group: 12}})
	}
	cMu, cSigma, cDf := centralParams(pr, panel)
	out = append(out,
		namedSource{"Student-t",
			"The central case to plan on: i.i.d. annual real returns at your mean, long-horizon volatility and tails. No mean reversion across years, so long horizons read a touch tougher than history; and when your history is shorter than the horizon it leans toward the broad-sample prior (a short window cannot show long-horizon tail and sequence risk).",
			centralSource(pr, cMu, cSigma, cDf, pr.Years)},
		namedSource{"Sequence stress",
			"Sequence-of-returns stress: clustered, persistent bull/bear regimes at the SAME long-run mean as Student-t, so a run of bad years can land early in retirement. The expected return is unchanged; only the ordering is stressed. Read it as the downside if the sequence is unlucky.",
			scenario.NewMarkovRegime(cMu, cSigma, cDf, pr.Years)},
		namedSource{"Broad-sample",
			"The empirical century: real returns block-bootstrapped from the actual 1870-2020 developed-market record (18 economies, Jorda-Schularick-Taylor), GDP-weighted world equity. Not this fund's short history and not synthetic; it carries the real bear decades (1929-32, the 1970s, Japan post-1990) that cause ruin. The broad-sample counterpoint to your favourable window.",
			broadSampleSource(pr.Years)},
		namedSource{"Lost decade",
			"Japan-style tail: a very sticky, deep bear averaging a whole decade, layered on your mean. Unlike Sequence stress this lowers the realised return too, modelling a prolonged real drawdown (Japanese equities 1990-2010). The grimmest planning model, the scenario where a retirement begins inside a lost decade.",
			scenario.NewLostDecadeRegime(cMu, cSigma, cDf, pr.Years)},
	)
	return out
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
	mu, sigma, df = pr.Mu, pr.Sigma, pr.Df
	// Anchor the central return to today's valuation: at a rich CAPE the next
	// decade's real return is compressed, so the whole-horizon central mean is
	// set to the CAPE-implied return rather than the fund's rosy history. This
	// overrides the slider mean and the panel blend below.
	if pr.CapeAdjust {
		return capeAdjustedMu(sigma), sigma, df
	}
	if panel == nil || pr.Years <= 0 {
		return
	}
	histYears := float64(panel.Periods() / 12)
	s := (float64(pr.Years) - histYears) / float64(pr.Years)
	s = math.Max(0, math.Min(s, 0.5))
	blend := func(fit, prior float64) float64 { return (1-s)*fit + s*prior }
	return blend(mu, consMu), blend(sigma, consSigma), blend(df, consDf)
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

	o := p.Simulate(nPaths, simWorkers, seed).Outcome()
	safe := fixedRule(p).Solve(target, decumul.WithdrawalAxis(0, capital*0.15), nPaths, simWorkers, seed)
	return ModelStat{
		Name: ns.name, Help: ns.help,
		Ruin: o.RuinProb, MedianWealth: o.TerminalP50,
		SafeSpend: safe, SafeWR: safe / capital,
	}
}

// fixedRule strips the adaptive spending rules (flex cut and guardrails) so the
// plan withdraws a fixed real amount. A safe-withdrawal solve must run on the
// fixed rule: guardrails rebase spending on the initial withdrawal rate, which
// makes ruin non-monotonic in the withdrawal and the bisection ill-defined (it
// can jump between very different "safe" spends for a tiny target change).
func fixedRule(p decumul.Plan) decumul.Plan {
	p.Flex = decumul.FlexRule{}
	p.Guard = decumul.Guardrails{}
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
	return fmt.Sprintf("Safe spend ≈ %.0f k€/yr (%.1f%%) at %.0f%% success · you plan %.0f k€ (%.1f%%)",
		central.SafeSpend/1000, central.SafeWR*100, (1-target)*100,
		pr.NeedAnnual/1000, pr.NeedAnnual/pr.Capital*100)
}
