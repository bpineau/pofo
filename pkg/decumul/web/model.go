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
	Age            int       `json:"age"`            // age at year 0, for the mortality view (0 = 52)
	PEACapital     float64   `json:"peaCapital"`     // euros held in the PEA envelope (17.2% on gains)
	AVCapital      float64   `json:"avCapital"`      // euros held in assurance-vie (9 200 €/yr allowance)
	GainFrac       float64   `json:"gainFrac"`       // embedded unrealised gain fraction at start
	Ratchet        bool      `json:"ratchet"`        // only-up spending rule (the written-rules cliquet)
	WRTrigger      float64   `json:"wrTrigger"`      // flex also cuts above this current WR (0 = off)
	SpendDrift     float64   `json:"spendDrift"`     // real spending drift per year (health costs)
	Smile          bool      `json:"smile"`          // Blanchett retirement-smile spending shape
	CapeAdjust     bool      `json:"capeAdjust"`     // anchor the central return to today's CAPE valuation
	Percent        float64   `json:"percent"`        // percentage-of-portfolio (VPW) rule; 0 = fixed real spending
	Glidepath      bool      `json:"glidepath"`      // rising-equity glidepath (bond tent) on the central model
	AnnuityShare   float64   `json:"annuityShare"`   // share of capital annuitized (joint-life real income); 0 = none
}

// age resolves the mortality age, defaulting to 52 (an early retiree).
func (pr Params) age() float64 {
	if pr.Age <= 0 {
		return 52
	}
	return float64(pr.Age)
}

// Card is one labelled summary figure shown above the charts. Help, when
// set, becomes the card's plain-language hover explanation.
type Card struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Help  string `json:"help,omitempty"`
}

// Result is the JSON returned for one parameter set. Note carries a
// user-facing caveat (e.g. a horizon longer than the available history for
// the cohorts model), empty when the result is fully usable. Cards is an
// ordered list so the UI shows the figures in a stable, readable order.
type Result struct {
	Note          string `json:"note"`
	Cards         []Card `json:"cards"`
	ArbitrageSVG  string `json:"arbitrageSvg"`  // ruin % vs buffer years
	Arbitrage2SVG string `json:"arbitrage2Svg"` // median terminal wealth vs buffer years
	RecoverySVG   string `json:"recoverySvg"`
}

// plan builds a decumul.Plan from the params, with a parametric source by
// default (source() may override it for the portfolio models).
func (pr Params) plan() decumul.Plan {
	p := decumul.Plan{
		Capital:    pr.Capital,
		NeedAnnual: pr.NeedAnnual,
		Years:      pr.Years,
		Buffer:     decumul.BufferSleeve{Years: pr.BufferYears, RealReturn: pr.BufferReturn, RefillStopYear: pr.BufferStopYear},
		Flex:       decumul.FlexRule{Threshold: 0.20, Cut: pr.FlexCut, WRThreshold: pr.WRTrigger},
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
	// The written-rules cliquet (Kitces ratchet, Ben's §10 skeleton): +10% of
	// the base spend per raise, only above 120% of the initial real capital,
	// at most every two years, capped at 120% of the base spend, and only
	// while the current rate stays comfortable (< 2.2%).
	if pr.Ratchet && pr.NeedAnnual > 0 {
		p.Ratchet = decumul.Ratchet{
			Trigger: 1.2, Step: 0.10 * pr.NeedAnnual, Cap: 1.2 * pr.NeedAnnual,
			Cooldown: 2, MaxWR: 0.022,
		}
	}
	// Percentage-of-portfolio (VPW): overrides the fixed-need policy with a
	// never-ruin variable-spending rule (annual). Set last so it takes priority.
	if pr.Percent > 0 {
		p.Percent = pr.Percent
	}
	// Partial annuitisation: spend a share of capital on a joint-life, real
	// immediate annuity (1% real rate, 10% insurer load), hedging longevity. The
	// premium leaves the portfolio; its lifelong income lowers the net need.
	if income := pr.annuityIncome(); income > 0 {
		p.Capital -= pr.AnnuityShare * pr.Capital
		p.Cashflows = append(p.Cashflows, decumul.Cashflow{FromYear: 0, Annual: income})
	}
	p.SpendSchedule = pr.spendSchedule()
	p.Envelopes = pr.envelopes()
	return p
}

// annuityIncome is the lifelong yearly real income bought by the annuitised
// share of capital (joint-life, 1% real rate, 10% insurer load); 0 when the
// annuity option is off.
func (pr Params) annuityIncome() float64 {
	if pr.AnnuityShare <= 0 || pr.Capital <= 0 {
		return 0
	}
	return decumul.AnnuityIncome(decumul.FrenchMortality, pr.age(), pr.AnnuityShare*pr.Capital, 0.01, 0.90)
}

// spendSchedule builds the per-year real spending multipliers from the drift
// and smile options; nil when spending is constant.
func (pr Params) spendSchedule() []float64 {
	if pr.SpendDrift == 0 && !pr.Smile {
		return nil
	}
	s := make([]float64, pr.Years)
	for k := range s {
		m := math.Pow(1+pr.SpendDrift, float64(k))
		if pr.Smile {
			m *= smileAt(k)
		}
		s[k] = m
	}
	return s
}

// smileAt approximates the Blanchett retirement-spending smile: real spending
// drifts down through the go-go and slow-go years (about -1%/yr to year 15),
// plateaus, then climbs back with late-life health costs.
func smileAt(k int) float64 {
	switch {
	case k <= 15:
		return 1 - 0.010*float64(k)
	case k <= 25:
		return 0.85
	default:
		return math.Min(0.85+0.012*float64(k-25), 1.05)
	}
}

// envelopes translates the PEA/AV sliders into the ordered tax pockets, CTO
// first (the classic French drain order), with the shared embedded-gain
// fraction. It returns nil when the plan is the legacy single CTO sleeve.
func (pr Params) envelopes() []decumul.Envelope {
	if pr.PEACapital <= 0 && pr.AVCapital <= 0 && pr.GainFrac <= 0 {
		return nil
	}
	growth := pr.Capital - math.Min(pr.BufferYears*pr.NeedAnnual, pr.Capital)
	out := []decumul.Envelope{{
		Name:     "CTO",
		Amount:   math.Max(0, growth-pr.PEACapital-pr.AVCapital),
		GainFrac: pr.GainFrac,
		Tax:      decumul.CTOFlatTax{Rate: pr.TaxRate},
	}}
	if pr.PEACapital > 0 {
		out = append(out, decumul.Envelope{
			Name: "PEA", Amount: pr.PEACapital, GainFrac: pr.GainFrac,
			// Past 5 years, PEA withdrawals only pay social levies on gains.
			Tax: decumul.CTOFlatTax{Rate: 0.172},
		})
	}
	if pr.AVCapital > 0 {
		out = append(out, decumul.Envelope{
			Name: "AV", Amount: pr.AVCapital, GainFrac: pr.GainFrac,
			// Past 8 years: 9 200 €/yr of gains tax-free (couple), then
			// 7.5% + 17.2% social levies on the excess.
			Tax: decumul.AVTax{Rate: 0.247, Allowance: 9200},
		})
	}
	return out
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
	// The drawdown-shape detail stats are computed on the SURVIVING paths:
	// with any ruin at all, the all-paths minima saturate at -100%/yr and a
	// 100% drawdown, which restates "some paths ruin" (already the headline)
	// and hides the useful figure, how rough the ride gets even when the plan
	// works. Ruin itself stays computed on every path.
	so := survivors(e).Outcome()
	// Recovery-time distribution: keep the first years legible and fold the long
	// tail into a single "12y+" bucket, rather than 45 unreadable slivers.
	const recoveryCap = 12
	var bars []chart.Bar
	var tail float64
	for _, b := range e.RecoveryTimeDistribution() {
		if b.Years <= recoveryCap {
			bars = append(bars, chart.Bar{
				Label: fmt.Sprintf("%dy", b.Years),
				Value: b.Share * 100,
				Text:  fmt.Sprintf("%.0f%%", b.Share*100),
			})
		} else {
			tail += b.Share
		}
	}
	if tail > 0 {
		bars = append(bars, chart.Bar{
			Label: fmt.Sprintf("%dy+", recoveryCap),
			Value: tail * 100,
			Text:  fmt.Sprintf("%.0f%%", tail*100),
		})
	}

	return Result{
		// The hero strip (/api/models) carries the multi-model ruin and safe
		// withdrawal shown in the UI; these detail metrics are computed for the
		// API response and the tests, not rendered on the page.
		Cards: []Card{
			{Label: "Ruin", Value: fmt.Sprintf("%.1f%%", o.RuinProb*100)},
			{Label: "Withdrawal rate", Value: fmt.Sprintf("%.2f%%", pr.NeedAnnual/pr.Capital*100)},
			{Label: "Terminal wealth (p50)", Value: fmtWealth(o.TerminalP50)},
			{Label: "Terminal wealth (p5)", Value: fmtWealth(o.TerminalP5)},
			{Label: "Median years underwater", Value: fmt.Sprintf("%.0f y", o.MedianYearsUnderwater)},
			{Label: "Worst 10y real CAGR (surviving p5)", Value: fmt.Sprintf("%.1f%%/yr", so.Worst10yP5*100)},
			{Label: "Worst drawdown (surviving 5%)", Value: fmt.Sprintf("%.1f%%", so.CDaR*100)},
			{Label: "Median cumulative tax", Value: fmtWealth(o.MedianCumTax)},
			{Label: "Effective tax rate", Value: fmt.Sprintf("%.1f%%", o.EffectiveTaxRate*100)},
		},
		// Two single-axis panels sharing the x axis instead of a dual-axis
		// chart: each curve gets its own honest scale, so the interior
		// optimum in ruin and the growth drag on terminal wealth both show.
		ArbitrageSVG: chart.MultiLine(chart.Options{Title: "Ruin % vs buffer years", Width: 720, Height: 300},
			"Buffer years", "Ruin %", []chart.XYSeries{ruinSeries(sweep)},
			chart.Marker{Axis: 'x', Value: pr.BufferYears, Label: "your buffer"}),
		Arbitrage2SVG: chart.MultiLine(chart.Options{Title: "Median terminal wealth vs buffer years", Width: 720, Height: 300},
			"Buffer years", "Terminal p50 M€", []chart.XYSeries{terminalSeries(sweep)},
			chart.Marker{Axis: 'x', Value: pr.BufferYears, Label: "your buffer"}),
		RecoverySVG: chart.Bars(chart.Options{Title: "Recovery-time distribution (share %)", Width: 600, Height: 360}, bars),
	}
}

// survivors filters an ensemble down to its non-ruined paths, for the detail
// statistics that saturate meaninglessly once any path hits zero.
func survivors(e decumul.Ensemble) decumul.Ensemble {
	out := decumul.Ensemble{Years: e.Years}
	for _, p := range e.Paths {
		if !p.Ruined {
			out.Paths = append(out.Paths, p)
		}
	}
	return out
}

// fmtWealth renders a real-euro amount with a readable unit: M€ with two
// decimals from a million up, k€ below.
func fmtWealth(v float64) string {
	if v >= 1e6 || v <= -1e6 {
		return fmt.Sprintf("%.2f M€", v/1e6)
	}
	return fmt.Sprintf("%.0f k€", v/1000)
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

// ruinSeries is the ruin-probability curve (%) against buffer years. The name
// stays empty: the panel title carries it, so no one-entry legend is drawn.
func ruinSeries(s []decumul.SweepPoint) chart.XYSeries {
	xs, ys := make([]float64, len(s)), make([]float64, len(s))
	for i, p := range s {
		xs[i], ys[i] = p.Value, p.RuinProb*100
	}
	return chart.XYSeries{Xs: xs, Ys: ys, Color: chart.PaletteColor(3)}
}

// terminalSeries is the median terminal-wealth curve (M€) against buffer
// years; unnamed for the same one-series-no-legend reason as ruinSeries.
func terminalSeries(s []decumul.SweepPoint) chart.XYSeries {
	xs, ys := make([]float64, len(s)), make([]float64, len(s))
	for i, p := range s {
		xs[i], ys[i] = p.Value, p.TerminalP50/1e6
	}
	return chart.XYSeries{Xs: xs, Ys: ys, Color: chart.PaletteColor(2)}
}
