package web

import (
	"fmt"

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
// the mean-preserving Regime, and the Conservative broad-sample prior) is always
// present; a panel adds the data-driven Historical and Block-bootstrap columns.
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
			out = append(out, namedSource{"Historical",
				"Replays this portfolio's actual return sequences, no resampling. Honest but limited: a short history holds few independent retirements, so it tends to look optimistic for long horizons.",
				scenario.Compounded{Inner: cohorts, Group: 12}})
		}
		out = append(out, namedSource{"Block bootstrap",
			"Resamples multi-year blocks of this portfolio's real returns, preserving clustered bear markets and cross-asset correlation. Manufactures many full-length retirements from a short history, but stays anchored to that one favourable window.",
			scenario.Compounded{Inner: scenario.StationaryBootstrap{Panel: *panel, Weights: w, MeanBlock: 24, Periods: months}, Group: 12}})
	}
	out = append(out,
		namedSource{"Student-t",
			"The calibrated central case to plan on: i.i.d. annual real returns at your mean, long-horizon volatility and tails. It assumes no mean reversion across years, so long (45-50y) horizons read a little tougher than history.",
			scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years}},
		namedSource{"Regime",
			"Sequence-risk stress: clustered, persistent bull/bear regimes at the same long-run mean as Student-t, so a run of bad years can land early in retirement. Read it as the downside if the sequence is unlucky.",
			scenario.NewMarkovRegime(pr.Mu, pr.Sigma, pr.Df, pr.Years)},
		namedSource{"Conservative",
			"Forward-looking pessimism, not this fund's history: a lower real return (~3.5% geometric), higher volatility, fat left tail and clustered drawdowns, in line with broad century-long developed-market evidence (Anarkulova et al.).",
			scenario.NewMarkovRegime(consMu, consSigma, consDf, pr.Years)},
	)
	return out
}

// evalModel runs one model: ruin and median wealth at the planned spend, and the
// safe withdrawal that meets the target ruin. The shared seed keeps the figures
// comparable across models.
func evalModel(base decumul.Plan, ns namedSource, capital, target float64, nPaths int) ModelStat {
	const seed = uint64(7)
	p := base
	p.Source = ns.source

	o := p.Simulate(nPaths, simWorkers, seed).Outcome()
	safe := p.Solve(target, decumul.WithdrawalAxis(0, capital*0.15), nPaths, simWorkers, seed)
	return ModelStat{
		Name: ns.name, Help: ns.help,
		Ruin: o.RuinProb, MedianWealth: o.TerminalP50,
		SafeSpend: safe, SafeWR: safe / capital,
	}
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
	case histYears*2 >= pr.Years:
		return "MEDIUM", fmt.Sprintf("Fund history %dy vs %dy horizon: the historical models extrapolate beyond the data.", histYears, pr.Years)
	default:
		return "LOW", fmt.Sprintf("Fund history %dy vs %dy horizon: far too short for the historical models; lean on the conservative column.", histYears, pr.Years)
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
