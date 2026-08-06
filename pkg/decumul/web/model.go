package web

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

const simWorkers = 8

// Params is the slider state posted by the browser. Weights is nil in
// parametric mode and holds per-holding fractions in portfolio mode.
type Params struct {
	Capital        float64   `json:"capital"`
	NeedAnnual     float64   `json:"needAnnual"`
	BufferYears    float64   `json:"bufferYears"`
	Mu             float64   `json:"mu"`
	Sigma          float64   `json:"sigma"`
	Df             float64   `json:"df"`
	BufferReturn   float64   `json:"bufferReturn"`
	Years          int       `json:"years"`
	PensionYear    int       `json:"pensionYear"`
	PensionAnnual  float64   `json:"pensionAnnual"`
	FlexCut        float64   `json:"flexCut"`
	TaxRate        float64   `json:"taxRate"`
	NPaths         int       `json:"nPaths"`
	Weights        []float64 `json:"weights"`
	Model          string    `json:"model"`          // "parametric" (default), "bootstrap", "cohorts"
	TargetRuin     float64   `json:"targetRuin"`     // solve target (fraction), used by /api/solve
	Monthly        bool      `json:"monthly"`        // step the kernel monthly (salary-like withdrawals)
	Regime         bool      `json:"regime"`         // stress: cluster bad years (Markov regime source, annual)
	BufferStopYear int       `json:"bufferStopYear"` // glidepath: stop refilling the buffer from this year (0 = never)
	SideAnnual     float64   `json:"sideAnnual"`     // temporary side income /yr (rental/activity)
	SideUntilYear  int       `json:"sideUntilYear"`  // side income runs until this year, exclusive
	Guardrails     bool      `json:"guardrails"`     // Guyton-Klinger guardrails (replaces the flex cut)
}

// Card is one labelled summary figure shown above the charts.
type Card struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Result is the JSON returned for one parameter set. Note carries a
// user-facing caveat (e.g. a horizon longer than the available history for
// the cohorts model), empty when the result is fully usable. Cards is an
// ordered list so the UI shows the figures in a stable, readable order.
type Result struct {
	Note         string `json:"note"`
	Cards        []Card `json:"cards"`
	ArbitrageSVG string `json:"arbitrageSvg"` // ruin % and terminal wealth vs buffer years (dual axis)
	RecoverySVG  string `json:"recoverySvg"`
}

// plan builds a decumul.Plan from the params, with a parametric source by
// default (source() may override it for the portfolio models).
func (pr Params) plan() decumul.Plan {
	p := decumul.Plan{
		Capital:    pr.Capital,
		NeedAnnual: pr.NeedAnnual,
		Years:      pr.Years,
		Buffer:     decumul.BufferSleeve{Years: pr.BufferYears, RealReturn: pr.BufferReturn, RefillStopYear: pr.BufferStopYear},
		Flex:       decumul.FlexRule{Threshold: 0.20, Cut: pr.FlexCut},
		Tax:        decumul.CTOFlatTax{Rate: pr.TaxRate},
		Source:     scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years},
		Monthly:    pr.Monthly && !pr.Regime, // the regime source is annual
	}
	if pr.PensionAnnual > 0 {
		p.Cashflows = append(p.Cashflows, decumul.Cashflow{FromYear: pr.PensionYear, Annual: pr.PensionAnnual})
	}
	if pr.SideAnnual > 0 {
		p.Cashflows = append(p.Cashflows, decumul.Cashflow{FromYear: 0, ToYear: pr.SideUntilYear, Annual: pr.SideAnnual})
	}
	// Guardrails band centred on the initial withdrawal rate (±20%).
	if pr.Guardrails && pr.Capital > 0 {
		wr0 := pr.NeedAnnual / pr.Capital
		p.Guard = decumul.Guardrails{Upper: wr0 * 1.2, Lower: wr0 * 0.8, Cut: 0.10, Raise: 0.10}
	}
	return p
}

// source picks the return model. With a non-nil (monthly) panel and a
// non-parametric Model, it resamples that panel at monthly frequency under
// the live weights and compounds to annual; otherwise it falls back to the
// annual parametric source.
func (pr Params) source(panel *scenario.Panel) scenario.Source {
	months := pr.Years * 12
	if panel != nil && pr.Weights != nil {
		var inner scenario.Source
		switch pr.Model {
		case "bootstrap":
			inner = scenario.StationaryBootstrap{Panel: *panel, Weights: pr.Weights, MeanBlock: 24, Periods: months}
		case "cohorts":
			inner = scenario.HistoricalCohorts{Panel: *panel, Weights: pr.Weights, Periods: months}
		}
		if inner != nil {
			if pr.Monthly {
				return inner // the monthly kernel consumes the monthly source directly
			}
			return scenario.Compounded{Inner: inner, Group: 12}
		}
	}
	if pr.Regime {
		// Stress regimes: a mean-preserving two-state Markov source (annual)
		// where bad years cluster, preserving the target long-run mean so the
		// stress measures sequence risk only, not a secretly lower expected return.
		return scenario.NewMarkovRegime(pr.Mu, pr.Sigma, pr.Df, pr.Years)
	}
	if pr.Monthly {
		// Monthly i.i.d. parametric draws that compound to the annual mu/sigma.
		return scenario.ParametricSource{
			Mu: math.Pow(1+pr.Mu, 1.0/12) - 1, Sigma: pr.Sigma / math.Sqrt(12), Df: pr.Df, Periods: months}
	}
	return scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years}
}

// Compute runs the parametric model (no panel).
func Compute(pr Params) Result { return ComputeWithPanel(pr, nil) }

// ComputeWithPanel is Compute with an optional historical panel for the
// bootstrap/cohort models and live re-weighting.
func ComputeWithPanel(pr Params, panel *scenario.Panel) Result {
	p := pr.plan()
	p.Source = pr.source(panel)
	res := computeFrom(pr, p)
	if res.Note == "" {
		res.Note = reliabilityCaveat(pr, panel)
	}
	return res
}

// reliabilityCaveat warns when the historical sample is too short to speak to a
// retirement-length horizon. A 27-year window (e.g. MSCI World since 1999) holds
// no independent 40-year retirement, so any precise ruin figure from it
// understates the long-horizon, sequence-of-returns risk that broad, century-long
// samples reveal (Anarkulova, Cederburg & O'Doherty 2023 find materially higher
// failure rates for a fixed 4% rule). Returns an empty string when the sample is
// adequate or the model is purely parametric.
func reliabilityCaveat(pr Params, panel *scenario.Panel) string {
	if panel == nil || pr.Model == "parametric" {
		return ""
	}
	histYears := panel.Periods() / 12
	if histYears >= pr.Years {
		return ""
	}
	return fmt.Sprintf(
		"Caution: the historical sample is %d years but the horizon is %d, so it contains no independent full-length retirement. The ruin figure is optimistic about long-horizon sequence risk; broad, century-long studies find a fixed 4%% rule fails far more often.",
		histYears, pr.Years)
}

// cohortsNote returns a user-facing caveat when the plan's source is a cohorts
// model with too little history for the horizon, otherwise an empty string.
func cohortsNote(pr Params, p decumul.Plan) string {
	src := p.Source
	if c, ok := src.(scenario.Compounded); ok {
		src = c.Inner // the historical source is wrapped in a Compounded
	}
	if hc, ok := src.(scenario.HistoricalCohorts); ok && hc.Count() == 0 {
		return fmt.Sprintf(
			"Not enough history for a %d-year horizon under the cohorts model (only %d years of aligned data). Use the bootstrap or parametric model, or shorten the horizon.",
			pr.Years, hc.Panel.Periods()/12)
	}
	return ""
}

// computeFrom runs the simulation and renders the charts for a built plan.
func computeFrom(pr Params, p decumul.Plan) Result {
	if pr.NPaths == 0 {
		pr.NPaths = 5000
	}
	if note := cohortsNote(pr, p); note != "" {
		return Result{Note: note}
	}
	seed := uint64(7)

	// buffer arbitrage curve (ruin and terminal vs buffer years). BufferYears
	// applies to every Source, so this sweep cannot fail; surface any error
	// rather than hide it.
	bufVals := []float64{0, 1, 2, 3, 4, 5, 6, 8, 10}
	sweep, err := p.Sweep1D(decumul.BufferYears, bufVals, pr.NPaths, simWorkers, seed)
	if err != nil {
		return Result{Note: err.Error()}
	}

	// headline outcome and recovery distribution at the selected buffer.
	e := p.Simulate(pr.NPaths, simWorkers, seed)
	o := e.Outcome()
	var bars []chart.Bar
	for _, b := range e.RecoveryTimeDistribution() {
		bars = append(bars, chart.Bar{
			Label: fmt.Sprintf("%dy", b.Years),
			Value: b.Share * 100,
			Text:  fmt.Sprintf("%.0f%%", b.Share*100),
		})
	}

	return Result{
		// The hero strip (/api/models) carries the multi-model ruin and safe
		// withdrawal shown in the UI; these detail metrics are computed for the
		// API response and the tests, not rendered on the page.
		Cards: []Card{
			{"Ruin", fmt.Sprintf("%.1f%%", o.RuinProb*100)},
			{"Withdrawal rate", fmt.Sprintf("%.2f%%", pr.NeedAnnual/pr.Capital*100)},
			{"Terminal wealth (p50)", fmt.Sprintf("%.0f k€", o.TerminalP50/1000)},
			{"Terminal wealth (p5)", fmt.Sprintf("%.0f k€", o.TerminalP5/1000)},
			{"Median years underwater", fmt.Sprintf("%.0f y", o.MedianYearsUnderwater)},
			{"Worst 10y real CAGR (p5)", fmt.Sprintf("%.1f%%/yr", o.Worst10yP5*100)},
			{"Worst 10y real CAGR (min)", fmt.Sprintf("%.1f%%/yr", o.Worst10yCAGR*100)},
			{"Conditional drawdown (worst 5%)", fmt.Sprintf("%.1f%%", o.CDaR*100)},
			{"Median cumulative tax", fmt.Sprintf("%.0f k€", o.MedianCumTax/1000)},
			{"Effective tax rate", fmt.Sprintf("%.1f%%", o.EffectiveTaxRate*100)},
		},
		ArbitrageSVG: chart.LineDual(chart.Options{Title: "Buffer arbitrage: ruin vs terminal wealth"},
			"Buffer years", ruinSeries(sweep), terminalSeries(sweep)),
		RecoverySVG: chart.Bars(chart.Options{Title: "Recovery-time distribution (share %)"}, bars),
	}
}

// SolveResult answers the two "solve" questions for a scenario: the capital
// needed to hit a target ruin, and the ruin-minimising buffer at the current
// capital. Note carries a caveat when the model cannot answer (e.g. cohorts).
type SolveResult struct {
	Note            string  `json:"note"`
	TargetRuin      float64 `json:"targetRuin"`      // requested ruin target (fraction)
	RequiredCapital float64 `json:"requiredCapital"` // smallest capital meeting the target
	BestBufferYears float64 `json:"bestBufferYears"` // ruin-minimising buffer at current capital
	BestBufferRuin  float64 `json:"bestBufferRuin"`  // ruin at that buffer
}

// capital search bounds for the solver, generous around the slider range.
const solveLo, solveHi = 200000.0, 6000000.0

// Solve answers the capital and buffer solve questions for the params. A target
// of 0 defaults to 5% ruin.
func Solve(pr Params, panel *scenario.Panel) SolveResult {
	if pr.NPaths == 0 {
		pr.NPaths = 5000
	}
	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	p := pr.plan()
	p.Source = pr.source(panel)
	if note := cohortsNote(pr, p); note != "" {
		return SolveResult{Note: note}
	}
	seed := uint64(7)
	years, ruin, err := p.BestBuffer([]float64{0, 1, 2, 3, 4, 5, 6, 8, 10}, pr.NPaths, simWorkers, seed)
	if err != nil {
		return SolveResult{Note: err.Error()}
	}
	return SolveResult{
		TargetRuin:      target,
		RequiredCapital: p.CapitalForRuin(target, solveLo, solveHi, pr.NPaths, simWorkers, seed),
		BestBufferYears: years,
		BestBufferRuin:  ruin,
	}
}

// ruinSeries is the ruin-probability curve (%) against buffer years.
func ruinSeries(s []decumul.SweepPoint) chart.XYSeries {
	xs, ys := make([]float64, len(s)), make([]float64, len(s))
	for i, p := range s {
		xs[i], ys[i] = p.Value, p.RuinProb*100
	}
	return chart.XYSeries{Name: "Ruin %", Xs: xs, Ys: ys, Color: chart.PaletteColor(3)}
}

// terminalSeries is the median terminal-wealth curve (k€) against buffer years.
func terminalSeries(s []decumul.SweepPoint) chart.XYSeries {
	xs, ys := make([]float64, len(s)), make([]float64, len(s))
	for i, p := range s {
		xs[i], ys[i] = p.Value, p.TerminalP50/1000
	}
	return chart.XYSeries{Name: "Terminal wealth p50 (k€)", Xs: xs, Ys: ys, Color: chart.PaletteColor(2)}
}
