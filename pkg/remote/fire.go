package remote

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"slices"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/marketdata"
)

const maxWork int64 = 60_000_000

func planInput(in *PlanInput, axis string, r *Result) (fire.PlanSpec, error) {
	var p fire.PlanSpec
	if in == nil || in.HorizonYears == nil {
		return p, invalid("plan", "Plan and horizon_years are required.")
	}
	p = fire.Defaults()
	p.Years = *in.HorizonYears
	if p.Years < 1 || p.Years > 100 {
		return p, invalid("plan.horizon_years", "Horizon must be 1 through 100 years.")
	}
	for _, v := range []struct {
		name     string
		input    *float64
		dst      *float64
		positive bool
	}{{"capital", in.Capital, &p.Capital, true}, {"annual_spending", in.AnnualSpending, &p.AnnualSpending, false}} {
		if v.name == axis {
			if v.input != nil {
				return p, invalid("plan."+v.name, "Omit the amount being solved.")
			}
			continue
		}
		if v.input == nil || !finiteRange(*v.input, 0, 1e10) || v.positive && *v.input == 0 {
			return p, invalid("plan."+v.name, "Required finite EUR amount between 0 and 1e10; capital must be positive.")
		}
		*v.dst = *v.input
	}
	p.Policy = def(in.Policy, p.Policy, "plan.policy", r)
	if p.Policy != fire.FixedReal && p.Policy != fire.GuytonKlinger {
		return p, invalid("plan.policy", "Expected fixed_real or guyton_klinger.")
	}
	if axis != "" && p.Policy != fire.FixedReal {
		return p, &Error{Code: "unsupported_configuration", Message: "solve_fire supports fixed_real only: adaptive capital risk is not monotone. Use simulate_fire for adaptive plans or its fixed-rule safe_spending reference.", Path: "plan.policy"}
	}
	p.BufferYears = def(in.BufferYears, p.BufferYears, "plan.buffer_years", r)
	p.BufferRealReturn = def(in.BufferRealReturnPct, assumptionPct(p.BufferRealReturn), "plan.buffer_real_return_pct", r) / 100
	p.TaxRate = def(in.TaxRatePct, assumptionPct(p.TaxRate), "plan.tax_rate_pct", r) / 100
	p.InitialGain = def(in.InitialGainPct, assumptionPct(p.InitialGain), "plan.initial_gain_pct", r) / 100
	p.Inflation = def(in.InflationPct, assumptionPct(p.Inflation), "plan.inflation_pct", r) / 100
	if !finiteRange(p.BufferYears, 0, 100) || !finiteRange(p.BufferRealReturn, -.99, 1) || (!finiteRange(p.TaxRate, 0, 1) || p.TaxRate == 1) || !finiteRange(p.InitialGain, 0, 1) || !finiteRange(p.Inflation, 0, .2) {
		return p, invalid("plan", "Invalid buffer, tax, gain or inflation assumptions; see the contract ranges.")
	}
	if in.TaxRatePct == nil || in.InitialGainPct == nil || in.InflationPct == nil {
		msg := "Omitted tax/gain/inflation assumptions use this engine's defaults; inspect meta.defaulted."
		if in.TaxRatePct == nil {
			msg += fmt.Sprintf(" %.3g%%: %s.", p.TaxRate*100, fire.TaxAssumptionLabel)
		}
		warn(r, "defaulted_tax_assumptions", msg, "plan")
	}
	if in.Pension != nil {
		v := in.Pension
		if v.AnnualAmount == nil || v.StartYear == nil || !finiteRange(*v.AnnualAmount, 0, 1e10) || *v.StartYear < 0 || *v.StartYear > p.Years {
			return p, invalid("plan.pension", "Pension requires an amount 0..1e10 and start_year 0..horizon_years.")
		}
		drift := def(v.RealDriftPct, 0., "plan.pension.real_drift_pct", r) / 100
		if !finiteRange(drift, -.05, .05) {
			return p, invalid("plan.pension.real_drift_pct", "Pension real drift must be -5 through 5 percent.")
		}
		p.Pension = &fire.Pension{AnnualAmount: *v.AnnualAmount, StartYear: *v.StartYear, RealDrift: drift}
		if *v.StartYear == p.Years {
			warn(r, "pension_after_horizon", "Pension begins at the horizon and makes no payments in this plan.", "plan.pension.start_year")
		}
	}
	return p, nil
}
func finiteRange(v, lo, hi float64) bool { return !math.IsNaN(v) && v >= lo && v <= hi }

// assumptionPct removes binary conversion noise at the displayed assumptions
// boundary (12 decimal places in percent). Computed outcomes remain untouched.
func assumptionPct(fraction float64) float64 { return math.Round(fraction*1e14) / 1e12 }

func effectivePlan(p fire.PlanSpec) EffectivePlan {
	in := PlanInput{Capital: ptr(p.Capital), AnnualSpending: ptr(p.AnnualSpending), HorizonYears: ptr(p.Years), Policy: ptr(p.Policy), BufferYears: ptr(p.BufferYears), BufferRealReturnPct: ptr(assumptionPct(p.BufferRealReturn)), TaxRatePct: ptr(assumptionPct(p.TaxRate)), InitialGainPct: ptr(assumptionPct(p.InitialGain)), InflationPct: ptr(assumptionPct(p.Inflation))}
	if p.Pension != nil {
		in.Pension = &Pension{ptr(p.Pension.AnnualAmount), ptr(p.Pension.StartYear), ptr(p.Pension.RealDrift * 100)}
	}
	e := EffectivePlan{PlanInput: in, BufferAmount: min(p.Capital, p.BufferYears*p.AnnualSpending), BufferDrawdownTriggerPct: 100 * fire.BufferDrawThreshold, BufferRefillGrowthCapPct: 100 * fire.BufferRefillCap}
	if p.Policy == fire.GuytonKlinger && p.Capital > 0 {
		e.GuardUpperRatePct = ptr(p.AnnualSpending / p.Capital * 100 * fire.GuardUpperMultiplier)
		e.GuardLowerRatePct = ptr(p.AnnualSpending / p.Capital * 100 * fire.GuardLowerMultiplier)
		e.GuardCutPct = ptr(100 * fire.GuardCut)
		e.GuardRaisePct = ptr(100 * fire.GuardRaise)
	}
	return e
}
func sampleInput(paths *int, seed *int64, r *Result) (int, int64, error) {
	n := def(paths, 2000, "paths", r)
	k := def(seed, int64(7), "seed", r)
	if n < 100 || n > 10000 || k < 0 || k > 1<<53-1 {
		return 0, 0, invalid("paths/seed", "Paths must be 100..10000; seed an integer 0..2^53-1.")
	}
	return n, k, nil
}
func validateMarket(in *MarketInput, r *Result) error {
	if in == nil {
		return invalid("market", "Market assumptions are required.")
	}
	cape := def(in.CAPEAdjust, false, "market.cape_adjust", r)
	in.CAPEAdjust = ptr(cape)
	switch in.Mode {
	case "parametric":
		if in.Holdings != nil || in.History != nil || in.DataPolicy != nil || in.From != "" || in.To != "" {
			return invalid("market", "Parametric mode cannot include portfolio/history fields.")
		}
		if in.RealMeanReturnPct == nil || in.AnnualVolatilityPct == nil || in.DegreesOfFreedom == nil || !finiteRange(*in.RealMeanReturnPct, -100, 100) || !finiteRange(*in.AnnualVolatilityPct, 0, 200) || !finiteRange(*in.DegreesOfFreedom, 0, 1e6) {
			return invalid("market", "Parametric mode requires real_mean_return_pct (-100..100), annual_volatility_pct (0..200) and degrees_of_freedom (0..1e6).")
		}
	case "portfolio":
		if in.RealMeanReturnPct != nil || in.AnnualVolatilityPct != nil || in.DegreesOfFreedom != nil {
			return invalid("market", "Portfolio mode fits parameters; do not also supply parametric assumptions.")
		}
		if err := validateHoldings(in.Holdings, "market.holdings", in.History); err != nil {
			return err
		}
		// Resolve portfolio data defaults without adding irrelevant analysis sections.
		for _, v := range []struct {
			p           **string
			name, value string
		}{{&in.History, "history", "observed"}, {&in.DataPolicy, "data_policy", "cached_or_fetch"}} {
			x := def(*v.p, v.value, "market."+v.name, r)
			*v.p = ptr(x)
		}
		dummy := &Result{Meta: Meta{Defaulted: []Default{}}}
		_, err := analysisOptions(AnalysisInput{Currency: ptr("EUR"), History: in.History, DataPolicy: in.DataPolicy, From: in.From, To: in.To}, nil, nil, dummy)
		if err != nil {
			return err
		}
	default:
		return invalid("market.mode", "Expected parametric or portfolio.")
	}
	return nil
}
func (s *Service) prepareMarket(ctx context.Context, in MarketInput, years int, r *Result) (fire.MarketSpec, []Provenance, error) {
	m := fire.MarketSpec{CAPE: *in.CAPEAdjust}
	if m.CAPE {
		c := fire.CurrentCAPE()
		if !c.Valid() {
			return m, nil, &Error{Code: "unsupported_configuration", Message: "Requested CAPE adjustment is unavailable or invalid."}
		}
		r.Meta.CAPE = &CAPEData{c.AsOf, Number(c.Value), Number(c.ImpliedReal * 100), c.Stale}
		if c.Stale {
			warn(r, "stale_cape", "Bundled CAPE observation is older than 366 days.", "market.cape_adjust")
		}
	}
	if in.Mode == "parametric" {
		m.Mu = *in.RealMeanReturnPct / 100
		m.Sigma = *in.AnnualVolatilityPct / 100
		m.Nu = *in.DegreesOfFreedom
		return m, nil, nil
	}
	ai := AnalysisInput{Currency: ptr("EUR"), History: in.History, DataPolicy: in.DataPolicy, From: in.From, To: in.To}
	dummy := &Result{Meta: Meta{Defaulted: []Default{}}}
	opt, err := analysisOptions(ai, nil, nil, dummy)
	if err != nil {
		return m, nil, err
	}
	ctx = analysisContext(ctx, ai)
	ids := []string{}
	for _, h := range in.Holdings {
		ids = append(ids, h.ID)
		m.Weights = append(m.Weights, *h.WeightPct/100)
	}
	g := s.source()
	src, err := loaded(ctx, g, ids, opt, r)
	if err != nil {
		return m, nil, err
	}
	var assets []fire.AssetSeries
	var prov []Provenance
	for _, h := range in.Holdings {
		v := src[h.ID]
		assets = append(assets, fire.AssetSeries{Weight: *h.WeightPct / 100, Points: v.Points})
		prov = append(prov, provenance(v))
		warnings(r, v.Warnings, h.ID)
	}
	inflation, err := g.FetchExtended(ctx, "^HICP-FR", marketdata.FetchOptions{ExactOnly: true})
	if err != nil {
		return m, nil, err
	}
	prov = append(prov, provenance(inflation))
	for i := range assets {
		points := assets[i].Points
		if len(points) > 0 && points[len(points)-1].Date.After(inflation.Last().Date) {
			kept := points[:0:0]
			for _, pt := range points {
				if !pt.Date.After(inflation.Last().Date) {
					kept = append(kept, pt)
				}
			}
			assets[i].Points = kept
			warn(r, "inflation_coverage_trimmed", "Portfolio history ends at the last known inflation observation; later nominal prices were excluded.", "market")
		}
	}
	panel, coverage, err := fire.BuildMonthlyPanel(assets, inflation.Points)
	if err != nil {
		return m, nil, &Error{Code: "no_common_period", Message: "No usable consecutive monthly portfolio/inflation window."}
	}
	if panel.Periods() < fire.MinPanelMonths {
		return m, nil, &Error{Code: "unsupported_configuration", Message: "Portfolio FIRE requires at least 24 consecutive real-return months."}
	}
	sum := 0.
	for _, v := range m.Weights {
		sum += v
	}
	for i := range m.Weights {
		m.Weights[i] /= sum
	}
	panel.Weights = m.Weights
	m.Panel = &panel
	fit := fire.FitParametric(panel, m.Weights)
	if !fit.Valid() {
		return m, nil, &Error{Code: "unsupported_configuration", Message: "The supplied portfolio history cannot produce a valid parametric fit."}
	}
	m.Mu, m.Sigma, m.Nu = fit.Mu, fit.Sigma, fit.Df
	warnings(r, coverage.Notes(), "market")
	if panel.Periods() < 12*years {
		warn(r, "history_shorter_than_horizon", fmt.Sprintf("Only %d usable months for a %d-year horizon; this limited sample cannot establish full-horizon reliability.", panel.Periods(), years), "market")
	}
	r.Meta.RequestedWindow = &Window{in.From, in.To}
	r.Meta.EffectiveWindow = &Window{date(coverage.First), date(coverage.Last)}
	return m, prov, nil
}
func modelInfo(p fire.PreparedModel) ModelInfo {
	m := ModelInfo{ID: p.ID, Name: p.Name, Description: p.Description, Basis: p.Basis, CAPEApplied: p.CAPEApplied}
	if p.Basis == "parametric_assumptions" || p.Basis == "portfolio_calibration" {
		m.RawParameters = &ModelParameters{Number(p.RawMu * 100), Number(p.RawSigma * 100), Number(p.RawNu)}
		m.EffectiveParameters = &ModelParameters{Number(p.EffectiveMu * 100), Number(p.EffectiveSigma * 100), Number(p.EffectiveNu)}
	}
	if p.ID == fire.Historical {
		m.Sampling = "sampled_cohorts"
		m.AvailableCohorts = p.HistoricalCohorts
	}
	return m
}
func solveBounds(b *Bounds, axis string, p fire.PlanSpec, path string, r *Result) (float64, float64, error) {
	lo, hi := 0., .15*p.Capital
	if axis == "capital" {
		lo, hi = 200000, 6000000
	}
	if b == nil {
		r.Meta.Defaulted = append(r.Meta.Defaulted, Default{path, Bounds{ptr(lo), ptr(hi)}})
	} else {
		if b.Min == nil || b.Max == nil {
			return 0, 0, invalid(path, "Both min and max are required.")
		}
		lo, hi = *b.Min, *b.Max
	}
	if !finiteRange(lo, 0, 1e10) || !finiteRange(hi, 0, 1e10) || lo >= hi || axis == "capital" && lo == 0 {
		return 0, 0, invalid(path, "Use ordered finite bounds 0..1e10; capital bounds must be positive.")
	}
	return lo, hi, nil
}
func solveEvals(lo, hi float64) int64 { return int64(max(0., math.Ceil(math.Log2((hi-lo)/10)))) + 3 }
func workCost(models []string, n, years int, evaluations int64) int64 {
	var cost int64
	for _, m := range models {
		generation := int64(1)
		if m == fire.Bootstrap || m == fire.Historical {
			generation = 12
		}
		cost += int64(n) * int64(years) * (generation + evaluations)
	}
	return cost
}
func checkWork(cost int64) error {
	if cost > maxWork {
		return &Error{Code: "work_budget_exceeded", Message: "Request exceeds 60 million path-steps; reduce paths, models, horizon or remove safe_spending."}
	}
	return nil
}
func (s *Service) simulate(ctx context.Context, in SimulationInput, r *Result) (SimulationData, error) {
	out := SimulationData{Models: []FireModelData{}}
	p, err := planInput(in.Plan, "", r)
	if err != nil {
		return out, err
	}
	n, seed, err := sampleInput(in.Paths, in.Seed, r)
	if err != nil {
		return out, err
	}
	if err = validateMarket(in.Market, r); err != nil {
		return out, err
	}
	models := in.Models
	if models == nil {
		models = []string{fire.Central}
		r.Meta.Defaulted = append(r.Meta.Defaulted, Default{"models", models})
	}
	if len(models) < 1 || len(models) > 6 {
		return out, invalid("models", "Select one through six distinct models.")
	}
	seen := map[string]bool{}
	for _, m := range models {
		if !slices.Contains(modelNames, m) || seen[m] {
			return out, invalid("models", "Select distinct supported models.")
		}
		seen[m] = true
		if (m == fire.Bootstrap || m == fire.Historical) && in.Market.Mode != "portfolio" {
			return out, &Error{Code: "unsupported_configuration", Message: "Bootstrap and historical require portfolio mode."}
		}
	}
	r.Meta.Sections = in.Sections
	if r.Meta.Sections == nil {
		r.Meta.Sections = []string{}
		r.Meta.Defaulted = append(r.Meta.Defaulted, Default{"sections", r.Meta.Sections})
	}
	seen = map[string]bool{}
	for _, v := range r.Meta.Sections {
		if !slices.Contains([]string{"annual_bands", "safe_spending", "outcome_metrics"}, v) || seen[v] {
			return out, invalid("sections", "Invalid or duplicate simulation section.")
		}
		seen[v] = true
	}
	if in.SafeSpending != nil && !section(r, "safe_spending") {
		return out, invalid("safe_spending", "Select the safe_spending section to use its options.")
	}
	var lo, hi, target float64
	evals := int64(1)
	if section(r, "safe_spending") {
		v := in.SafeSpending
		if v == nil {
			v = &SafeSpending{}
		}
		target = def(v.TargetRuinPct, 5., "safe_spending.target_ruin_pct", r)
		if !finiteRange(target, 0, 100) {
			return out, invalid("safe_spending.target_ruin_pct", "Target must be 0..100 percent.")
		}
		lo, hi, err = solveBounds(v.Bounds, "annual_spending", p, "safe_spending.bounds", r)
		if err != nil {
			return out, err
		}
		evals += solveEvals(lo, hi)
	}
	cost := workCost(models, n, p.Years, evals)
	if err = checkWork(cost); err != nil {
		return out, err
	}
	market, prov, err := s.prepareMarket(ctx, *in.Market, p.Years, r)
	if err != nil {
		return out, err
	}
	prepared := []fire.PreparedModel{}
	for _, m := range models {
		pm, e := fire.PrepareModel(m, market, p.Years)
		if e != nil {
			return out, &Error{Code: "unsupported_configuration", Message: e.Error(), Path: "models"}
		}
		prepared = append(prepared, pm)
	}
	r.Meta.Currency = "EUR"
	assumptions := FireAssumptions{PlanRole: "requested_plan", Plan: effectivePlan(p), Market: *in.Market, Paths: n, Seed: seed, StepFrequency: "annual", RuinDefinition: "fixed_horizon_unfunded_withdrawal", Provenance: prov, WorkPathSteps: cost}
	if market.Panel != nil {
		assumptions.HistoryMonths = market.Panel.Periods()
		assumptions.WeightConvention = "fixed_monthly_weights"
	}
	r.Meta.Assumptions = assumptions
	cappedWarning(p, r)
	usedCAPE := false
	for _, pm := range prepared {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		usedCAPE = usedCAPE || pm.CAPEApplied
		if pm.Caveat != "" {
			warn(r, "model_caveat", pm.Caveat, pm.ID)
		}
		plan, e := fire.BuildPlan(p)
		if e != nil {
			return out, invalid("plan", e.Error())
		}
		plan.Source = pm.Source
		draws, e := plan.DrawContext(ctx, n, workers(), uint64(seed))
		if e != nil {
			return out, e
		}
		ensemble, e := plan.SimulateOnContext(ctx, draws, workers())
		if e != nil {
			return out, e
		}
		outcome := ensemble.Outcome()
		spend := ensemble.SpendStats()
		householdCV, e := plan.HouseholdSpendCV(ensemble)
		if e != nil {
			return out, &Error{Code: "internal_error", Message: "Household spending aggregation failed."}
		}
		d := FireModelData{Model: modelInfo(pm), RuinPct: Number(outcome.RuinProb * 100), TerminalRealWealthP5: Number(outcome.TerminalP5), TerminalRealWealthP50: Number(outcome.TerminalP50), Spending: SpendingSummary{Number(spend.EverCutShare * 100), Number(spend.FirstCutMedian), Number(spend.CutYearsMedian), Number(spend.CutYearsP90), Number(householdCV)}}
		boundaryWarning(outcome.RuinProb, r, pm.ID)
		if section(r, "annual_bands") {
			wealth := ensemble.Fan([]float64{.05, .5, .95}, 0)
			spending, e := plan.HouseholdSpendBands(ensemble, []float64{.05, .5, .95})
			if e != nil {
				return out, &Error{Code: "internal_error", Message: "Household spending aggregation failed."}
			}
			for y := 0; y <= p.Years; y++ {
				row := AnnualBands{Year: y, WealthP5: Number(wealth.Bands[0][y]), WealthP50: Number(wealth.Bands[1][y]), WealthP95: Number(wealth.Bands[2][y])}
				if y < p.Years {
					row.SpendingP5 = ptr(Number(spending[0][y]))
					row.SpendingP50 = ptr(Number(spending[1][y]))
					row.SpendingP95 = ptr(Number(spending[2][y]))
				}
				d.AnnualBands = append(d.AnnualBands, row)
			}
		}
		if section(r, "outcome_metrics") {
			d.OutcomeMetrics = &OutcomeMetrics{Number(outcome.EffectiveTaxRate * 100), Number(outcome.MedianCumTax), Number(outcome.CDaR * 100), Number(outcome.Worst10yP5 * 100)}
			if math.IsNaN(outcome.Worst10yP5) {
				warn(r, "undefined_statistic", "No complete positive-wealth decade is available for worst_decade_p5_pct.", pm.ID)
			}
			if math.IsNaN(outcome.EffectiveTaxRate) {
				warn(r, "undefined_statistic", "Effective tax rate requires positive gross withdrawals.", pm.ID)
			}
		}
		if section(r, "safe_spending") {
			fixed := p
			fixed.Policy = fire.FixedReal
			sol, e := runSolve(ctx, fixed, pm, draws, "annual_spending", lo, hi, target, r)
			if e != nil {
				return out, e
			}
			sol.RequestedPolicy = p.Policy
			sol.PolicyChangedForSafeSpending = p.Policy != fire.FixedReal
			d.SafeSpending = &sol
		}
		out.Models = append(out.Models, d)
	}
	if *in.Market.CAPEAdjust && !usedCAPE {
		warn(r, "cape_not_applied", "None of the requested models uses CAPE adjustment.", "market.cape_adjust")
	}
	if err = ctx.Err(); err != nil {
		return out, err
	}
	return out, nil
}
func workers() int { return min(4, max(1, runtime.GOMAXPROCS(0))) }
func cappedWarning(p fire.PlanSpec, r *Result) {
	if p.BufferYears*p.AnnualSpending > p.Capital {
		warn(r, "buffer_capped_to_capital", "Initial buffer is capped at total capital.", "plan.buffer_years")
	}
}
func boundaryWarning(prob float64, r *Result, path string) {
	if prob == 0 || prob == 1 {
		warn(r, "boundary_probability_estimate", "Zero or all observed failures is not certainty; a zero plug-in Monte Carlo standard error does not measure model uncertainty.", path)
	}
}
func runSolve(ctx context.Context, p fire.PlanSpec, pm fire.PreparedModel, draws decumul.Draws, axis string, lo, hi, target float64, r *Result) (SolveData, error) {
	plan, err := fire.BuildPlan(p)
	if err != nil {
		return SolveData{}, invalid("plan", err.Error())
	}
	plan.Source = pm.Source
	apply := func(_ decumul.Plan, value float64) decumul.Plan {
		q := p
		if axis == "capital" {
			q.Capital = value
		} else {
			q.AnnualSpending = value
		}
		candidate, _ := fire.BuildPlan(q)
		candidate.Source = pm.Source
		return candidate
	}
	result, err := plan.SolveOnContext(ctx, target/100, decumul.SolveAxis{Apply: apply, Increasing: axis == "annual_spending", Lo: lo, Hi: hi}, draws, workers(), decumul.SolveOptions{})
	if err != nil {
		if ctx.Err() != nil {
			return SolveData{}, ctx.Err()
		}
		return SolveData{}, &Error{Code: "internal_error", Message: "Solver invariant failed."}
	}
	out := SolveData{Model: modelInfo(pm), SolveFor: axis, Status: string(result.Status), Solution: Number(result.Solution), RuinPct: Number(result.RuinProb * 100), TargetRuinPct: target, Bounds: Bounds{ptr(lo), ptr(hi)}, LowerRuinPct: Number(result.LowerRuin * 100), UpperRuinPct: Number(result.UpperRuin * 100), FinalBounds: Bounds{ptr(result.FinalLower), ptr(result.FinalUpper)}, BracketWidth: result.Width, Tolerance: result.Tolerance, Iterations: result.Iterations, ProbabilityResolutionPct: result.ProbabilityResolution * 100, RuinMCStandardErrorPP: Number(result.MCStandardErrorPP), RequestedPolicy: p.Policy, SolvedPolicy: p.Policy, WithdrawalRatePct: Number(math.NaN())}
	if !math.IsNaN(result.Solution) {
		if axis == "capital" {
			p.Capital = result.Solution
		} else {
			p.AnnualSpending = result.Solution
		}
		out.EffectivePlan = ptr(effectivePlan(p))
		out.WithdrawalRatePct = Number(100 * p.AnnualSpending / p.Capital)
		cappedWarning(p, r)
		boundaryWarning(result.RuinProb, r, pm.ID)
	}
	return out, nil
}
func (s *Service) solve(ctx context.Context, in SolveInput, r *Result) (SolveData, error) {
	var out SolveData
	if in.SolveFor != "capital" && in.SolveFor != "annual_spending" {
		return out, invalid("solve_for", "Choose capital or annual_spending.")
	}
	p, err := planInput(in.Plan, in.SolveFor, r)
	if err != nil {
		return out, err
	}
	n, seed, err := sampleInput(in.Paths, in.Seed, r)
	if err != nil {
		return out, err
	}
	if err = validateMarket(in.Market, r); err != nil {
		return out, err
	}
	model := def(in.Model, fire.Central, "model", r)
	if !slices.Contains(modelNames, model) {
		return out, invalid("model", "Unknown return model.")
	}
	if (model == fire.Bootstrap || model == fire.Historical) && in.Market.Mode != "portfolio" {
		return out, &Error{Code: "unsupported_configuration", Message: "Bootstrap and historical require portfolio mode."}
	}
	target := def(in.TargetRuinPct, 5., "target_ruin_pct", r)
	if !finiteRange(target, 0, 100) {
		return out, invalid("target_ruin_pct", "Target must be 0..100 percent.")
	}
	lo, hi, err := solveBounds(in.Bounds, in.SolveFor, p, "bounds", r)
	if err != nil {
		return out, err
	}
	cost := workCost([]string{model}, n, p.Years, solveEvals(lo, hi))
	if err = checkWork(cost); err != nil {
		return out, err
	}
	if in.SolveFor == "capital" {
		p.Capital = hi
	} else {
		p.AnnualSpending = lo
	}
	market, prov, err := s.prepareMarket(ctx, *in.Market, p.Years, r)
	if err != nil {
		return out, err
	}
	pm, err := fire.PrepareModel(model, market, p.Years)
	if err != nil {
		return out, &Error{Code: "unsupported_configuration", Message: err.Error(), Path: "model"}
	}
	if pm.Caveat != "" {
		warn(r, "model_caveat", pm.Caveat, model)
	}
	if *in.Market.CAPEAdjust && !pm.CAPEApplied {
		warn(r, "cape_not_applied", "This model does not use CAPE adjustment.", "market.cape_adjust")
	}
	plan, err := fire.BuildPlan(p)
	if err != nil {
		return out, invalid("plan", err.Error())
	}
	plan.Source = pm.Source
	draws, err := plan.DrawContext(ctx, n, workers(), uint64(seed))
	if err != nil {
		return out, err
	}
	r.Meta.Currency = "EUR"
	a := FireAssumptions{PlanRole: "favorable_bound_evaluation", Plan: effectivePlan(p), Market: *in.Market, Paths: n, Seed: seed, StepFrequency: "annual", RuinDefinition: "fixed_horizon_unfunded_withdrawal", WorkPathSteps: cost, Provenance: prov}
	if market.Panel != nil {
		a.HistoryMonths = market.Panel.Periods()
		a.WeightConvention = "fixed_monthly_weights"
	}
	r.Meta.Assumptions = a
	out, err = runSolve(ctx, p, pm, draws, in.SolveFor, lo, hi, target, r)
	if err == nil && out.EffectivePlan != nil {
		a.Plan = *out.EffectivePlan
		a.PlanRole = "verified_solution"
		r.Meta.Assumptions = a
	}
	return out, err
}
