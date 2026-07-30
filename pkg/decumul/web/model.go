package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

const simWorkers = 8

// Params is the slider state posted by the browser. Weights is nil in
// parametric mode and holds per-holding fractions in portfolio mode.
type Params struct {
	Capital       float64   `json:"capital"`
	NeedAnnual    float64   `json:"needAnnual"`
	BufferYears   float64   `json:"bufferYears"`
	Mu            float64   `json:"mu"`
	Sigma         float64   `json:"sigma"`
	Df            float64   `json:"df"`
	BufferReturn  float64   `json:"bufferReturn"`
	Years         int       `json:"years"`
	PensionYear   int       `json:"pensionYear"`
	PensionAnnual float64   `json:"pensionAnnual"`
	FlexCut       float64   `json:"flexCut"`
	TaxRate       float64   `json:"taxRate"`
	NPaths        int       `json:"nPaths"`
	Weights       []float64 `json:"weights"`
	Model         string    `json:"model"` // "parametric" (default), "bootstrap", "cohorts"
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
	BufferSVG    string `json:"bufferSvg"`
	RuinCurveSVG string `json:"ruinCurveSvg"`
	RecoverySVG  string `json:"recoverySvg"`
}

// plan builds a decumul.Plan from the params, with a parametric source by
// default (source() may override it for the portfolio models).
func (pr Params) plan() decumul.Plan {
	p := decumul.Plan{
		Capital:    pr.Capital,
		NeedAnnual: pr.NeedAnnual,
		Years:      pr.Years,
		Buffer:     decumul.BufferSleeve{Years: pr.BufferYears, RealReturn: pr.BufferReturn},
		Flex:       decumul.FlexRule{Threshold: 0.20, Cut: pr.FlexCut},
		Tax:        decumul.CTOFlatTax{Rate: pr.TaxRate},
		Source:     scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years},
	}
	if pr.PensionAnnual > 0 {
		p.Cashflows = []decumul.Cashflow{{FromYear: pr.PensionYear, Annual: pr.PensionAnnual}}
	}
	return p
}

// source picks the return model. With a non-nil (monthly) panel and a
// non-parametric Model, it resamples that panel at monthly frequency under
// the live weights and compounds to annual; otherwise it falls back to the
// annual parametric source.
func (pr Params) source(panel *scenario.Panel) scenario.Source {
	if panel != nil && pr.Weights != nil {
		months := pr.Years * 12
		switch pr.Model {
		case "bootstrap":
			inner := scenario.StationaryBootstrap{Panel: *panel, Weights: pr.Weights, MeanBlock: 24, Periods: months}
			return scenario.Compounded{Inner: inner, Group: 12}
		case "cohorts":
			inner := scenario.HistoricalCohorts{Panel: *panel, Weights: pr.Weights, Periods: months}
			return scenario.Compounded{Inner: inner, Group: 12}
		}
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
	return computeFrom(pr, p)
}

// computeFrom runs the simulation and renders the charts for a built plan.
func computeFrom(pr Params, p decumul.Plan) Result {
	if pr.NPaths == 0 {
		pr.NPaths = 5000
	}
	// The cohorts model cannot extrapolate beyond the available history:
	// report the limit honestly instead of producing all-zero (certain-ruin)
	// paths. The historical source is wrapped in a Compounded, so unwrap it.
	src := p.Source
	if c, ok := src.(scenario.Compounded); ok {
		src = c.Inner
	}
	if hc, ok := src.(scenario.HistoricalCohorts); ok && hc.Count() == 0 {
		return Result{Note: fmt.Sprintf(
			"Not enough history for a %d-year horizon under the cohorts model (only %d years of aligned data). Use the bootstrap or parametric model, or shorten the horizon.",
			pr.Years, hc.Panel.Periods()/12)}
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
		bars = append(bars, chart.Bar{Label: fmt.Sprintf("%dy", b.Years), Value: b.Share})
	}

	return Result{
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
		BufferSVG:    chart.Bars(chart.Options{Title: "Ruin % by buffer years"}, barsFromSweep(sweep)),
		RuinCurveSVG: chart.Bars(chart.Options{Title: "Terminal wealth p50 (k€) by buffer"}, terminalBars(sweep)),
		RecoverySVG:  chart.Bars(chart.Options{Title: "Recovery-time distribution"}, bars),
	}
}

func barsFromSweep(s []decumul.SweepPoint) []chart.Bar {
	out := make([]chart.Bar, len(s))
	for i, p := range s {
		out[i] = chart.Bar{Label: fmt.Sprintf("%.0fy", p.Value), Value: p.RuinProb * 100}
	}
	return out
}

func terminalBars(s []decumul.SweepPoint) []chart.Bar {
	out := make([]chart.Bar, len(s))
	for i, p := range s {
		out[i] = chart.Bar{Label: fmt.Sprintf("%.0fy", p.Value), Value: p.TerminalP50 / 1000}
	}
	return out
}

