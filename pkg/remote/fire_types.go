package remote

// Pension is real household income in EUR, beginning at zero-based StartYear.
type Pension struct {
	AnnualAmount *float64 `json:"annual_amount"`            // Annual pension income in real EUR, required.
	StartYear    *int     `json:"start_year"`               // Zero-based first pension payment year, through the horizon.
	RealDriftPct *float64 `json:"real_drift_pct,omitempty"` // Annual real pension drift in percent, default zero.
}

// PlanInput describes a fixed horizon in real EUR. The solve axis amount must
// be omitted for solve_fire. All other optional numeric omissions use shared defaults.
type PlanInput struct {
	Capital             *float64 `json:"capital,omitempty"`                // Starting real EUR wealth, including the buffer; omit when solving capital.
	AnnualSpending      *float64 `json:"annual_spending,omitempty"`        // Annual real EUR household spending, including pension-funded needs; omit when solving spending.
	HorizonYears        *int     `json:"horizon_years"`                    // Required fixed horizon, 1 through 100 years.
	Policy              *string  `json:"policy,omitempty"`                 // fixed_real by default; guyton_klinger is available for simulation.
	BufferYears         *float64 `json:"buffer_years,omitempty"`           // Initial buffer in years of spending, default 3; capped at capital.
	BufferRealReturnPct *float64 `json:"buffer_real_return_pct,omitempty"` // Annual real buffer return in percent, shared default 0.5.
	TaxRatePct          *float64 `json:"tax_rate_pct,omitempty"`           // Planning tax rate in percent, below 100; shared labelled default 32.8.
	InitialGainPct      *float64 `json:"initial_gain_pct,omitempty"`       // Unrealized gains as percent of starting invested wealth, default 50.
	InflationPct        *float64 `json:"inflation_pct,omitempty"`          // Annual inflation assumption for nominal gain taxation, default 2 percent.
	Pension             *Pension `json:"pension,omitempty"`                // Optional single real annual household pension.
}

// MarketInput selects exactly one of parametric assumptions or portfolio calibration.
// Arithmetic real means, volatilities and drift are percentages, not fractions.
type MarketInput struct {
	Mode                string    `json:"mode"`                            // Exactly parametric or portfolio.
	RealMeanReturnPct   *float64  `json:"real_mean_return_pct,omitempty"`  // Arithmetic expected annual real return in percent.
	AnnualVolatilityPct *float64  `json:"annual_volatility_pct,omitempty"` // Annual real-return standard deviation in percent.
	DegreesOfFreedom    *float64  `json:"degrees_of_freedom,omitempty"`    // Student-t tail parameter; zero selects the source default.
	Holdings            []Holding `json:"holdings,omitempty"`              // One through twenty fully invested holdings, or requested holding studies.
	History             *string   `json:"history,omitempty"`               // Observed prices by default, or explicitly extended reconstructed history.
	From                string    `json:"from,omitempty"`                  // Inclusive UTC start date, YYYY-MM-DD; empty is unbounded.
	To                  string    `json:"to,omitempty"`                    // Inclusive UTC end date, YYYY-MM-DD; empty is unbounded.
	DataPolicy          *string   `json:"data_policy,omitempty"`           // cached_or_fetch by default; cache_only forbids all network access.
	CAPEAdjust          *bool     `json:"cape_adjust,omitempty"`           // Whether to anchor applicable synthetic means to bundled CAPE, default false.
}

// Bounds is a finite, ordered search interval in EUR.
type Bounds struct {
	Min *float64 `json:"min"` // Inclusive lower search bound in real EUR.
	Max *float64 `json:"max"` // Inclusive upper search bound in real EUR.
}

// SafeSpending selects a fixed-real spending reference for each simulated model.
type SafeSpending struct {
	TargetRuinPct *float64 `json:"target_ruin_pct,omitempty"` // Maximum tolerated sampled ruin percentage, default 5; explicit zero is preserved.
	Bounds        *Bounds  `json:"bounds,omitempty"`          // Explicit search interval; defaults are reported when omitted.
}

// SimulationInput runs one to six models with bounded seeded paths and optional details.
type SimulationInput struct {
	Plan         *PlanInput    `json:"plan"`                    // Required real-EUR withdrawal plan or its resolved assumptions.
	Market       *MarketInput  `json:"market"`                  // Required parametric assumptions or portfolio calibration input.
	Models       []string      `json:"models,omitempty"`        // Distinct requested models, default central; at most six.
	Paths        *int          `json:"paths,omitempty"`         // Seeded sample count, 100 through 10000, default 2000.
	Seed         *int64        `json:"seed,omitempty"`          // Nonnegative integer through 2^53-1, default 7.
	Sections     []string      `json:"sections,omitempty"`      // Distinct requested result sections; defaults depend on the operation.
	SafeSpending *SafeSpending `json:"safe_spending,omitempty"` // Optional per-model fixed-real spending reference; requires its selected section.
}

// SolveInput finds fixed-real spending or capital; adaptive-policy bisection is
// unsupported because its capital-risk curve need not be monotone.
type SolveInput struct {
	Plan          *PlanInput   `json:"plan"`                      // Required real-EUR withdrawal plan or its resolved assumptions.
	Market        *MarketInput `json:"market"`                    // Required parametric assumptions or portfolio calibration input.
	Model         *string      `json:"model,omitempty"`           // One return model, default central, or its effective model identity.
	SolveFor      string       `json:"solve_for"`                 // Exactly annual_spending or capital; only fixed_real policy is supported.
	TargetRuinPct *float64     `json:"target_ruin_pct,omitempty"` // Maximum tolerated sampled ruin percentage, default 5; explicit zero is preserved.
	Bounds        *Bounds      `json:"bounds,omitempty"`          // Explicit search interval; defaults are reported when omitted.
	Paths         *int         `json:"paths,omitempty"`           // Seeded sample count, 100 through 10000, default 2000.
	Seed          *int64       `json:"seed,omitempty"`            // Nonnegative integer through 2^53-1, default 7.
}

// EffectivePlan contains resolved real-EUR assumptions and kernel conventions.
type EffectivePlan struct {
	PlanInput
	BufferAmount             float64  `json:"buffer_amount"`                  // Effective initial real EUR buffer, capped at capital.
	BufferDrawdownTriggerPct float64  `json:"buffer_drawdown_trigger_pct"`    // Wealth-drawdown threshold for drawing the buffer, in percent.
	BufferRefillGrowthCapPct float64  `json:"buffer_refill_growth_cap_pct"`   // Annual growth-sleeve fraction available for refill, in percent.
	GuardUpperRatePct        *float64 `json:"guard_upper_rate_pct,omitempty"` // Candidate-specific gross withdrawal rate triggering a cut, in percent.
	GuardLowerRatePct        *float64 `json:"guard_lower_rate_pct,omitempty"` // Candidate-specific gross withdrawal rate triggering a raise, in percent.
	GuardCutPct              *float64 `json:"guard_cut_pct,omitempty"`        // Spending reduction at an upper-band crossing, in percent.
	GuardRaisePct            *float64 `json:"guard_raise_pct,omitempty"`      // Spending increase at a lower-band crossing, in percent.
}

// CAPEData dates the bundled valuation anchor and distinguishes it from live data.
type CAPEData struct {
	AsOf                 string `json:"as_of"`                   // Source observation date or metadata month.
	Value                Number `json:"value"`                   // Resolved value of this field.
	ImpliedRealReturnPct Number `json:"implied_real_return_pct"` // CAPE-implied geometric real annual return in percent.
	Stale                bool   `json:"stale"`                   // Whether the observation is more than 366 days old.
}

// FireAssumptions describes the request's effective numerical conventions.
type FireAssumptions struct {
	PlanRole         string        `json:"plan_role"`                   // PlanRole distinguishes requested_plan, verified_solution and favorable_bound_evaluation.
	Plan             EffectivePlan `json:"plan"`                        // Required real-EUR withdrawal plan or its resolved assumptions.
	Market           MarketInput   `json:"market"`                      // Required parametric assumptions or portfolio calibration input.
	Paths            int           `json:"paths"`                       // Seeded sample count, 100 through 10000, default 2000.
	Seed             int64         `json:"seed"`                        // Nonnegative integer through 2^53-1, default 7.
	StepFrequency    string        `json:"step_frequency"`              // Withdrawal time step, annual in this contract.
	RuinDefinition   string        `json:"ruin_definition"`             // Fixed-horizon inability to fund a required net withdrawal.
	WeightConvention string        `json:"weight_convention,omitempty"` // Fixed monthly weights for portfolio FIRE; no analysis-strategy replay.
	HistoryMonths    int           `json:"history_months,omitempty"`    // Usable consecutive real-return history months.
	Provenance       []Provenance  `json:"provenance,omitempty"`        // Required source, observation and reconstruction qualifications.
	WorkPathSteps    int64         `json:"work_path_steps"`             // Conservative admitted generation and simulation work estimate.
}

// ModelParameters reports arithmetic real mean, annual volatility and Student-t tail shape.
type ModelParameters struct {
	RealMeanReturnPct   Number `json:"real_mean_return_pct"`  // Arithmetic expected annual real return in percent.
	AnnualVolatilityPct Number `json:"annual_volatility_pct"` // Annual real-return standard deviation in percent.
	DegreesOfFreedom    Number `json:"degrees_of_freedom"`    // Student-t tail parameter; zero selects the source default.
}

// ModelInfo identifies allocation relevance, calibration and historical sampling.
type ModelInfo struct {
	ID                  string           `json:"id"`                             // Canonical or explicitly requested asset/model identifier.
	Name                string           `json:"name"`                           // Human-readable name.
	Description         string           `json:"description"`                    // Human-readable operation or model semantics.
	Basis               string           `json:"basis"`                          // portfolio_history, portfolio_calibration, parametric_assumptions or external_reference.
	CAPEApplied         bool             `json:"cape_applied"`                   // Whether this particular model used the CAPE anchor.
	RawParameters       *ModelParameters `json:"raw_parameters,omitempty"`       // Supplied or fitted inputs before central calibration.
	EffectiveParameters *ModelParameters `json:"effective_parameters,omitempty"` // Effective synthetic inputs after calibration.
	Sampling            string           `json:"sampling,omitempty"`             // sampled_cohorts for uniformly sampled historical windows with replacement.
	AvailableCohorts    int              `json:"available_cohorts,omitempty"`    // Number of available cohort starts, not independent historical evidence.
}

// SpendingSummary conditions cut timing and duration on paths with at least one
// household-spending cut. Household spending includes pension income.
type SpendingSummary struct {
	EverCutPct             Number `json:"ever_cut_pct"`             // Percentage of paths with at least one household-spending cut.
	FirstCutMedianYear     Number `json:"first_cut_median_year"`    // Median first cut year conditional on cutting, null without cuts.
	CutYearsMedian         Number `json:"cut_years_median"`         // Median cut-year count conditional on cutting, null without cuts.
	CutYearsP90            Number `json:"cut_years_p90"`            // 90th percentile cut-year count conditional on cutting, null without cuts.
	CoefficientOfVariation Number `json:"coefficient_of_variation"` // Delivered household-spending standard deviation divided by its mean.
}

// AnnualBands contains p5/p50/p95 real wealth at start/end boundaries and
// household spending during plan years; the final boundary has no spending.
type AnnualBands struct {
	Year        int     `json:"year"`                   // Zero-based plan year or final wealth boundary.
	WealthP5    Number  `json:"wealth_p5"`              // 5th percentile real EUR wealth.
	WealthP50   Number  `json:"wealth_p50"`             // Median real EUR wealth.
	WealthP95   Number  `json:"wealth_p95"`             // 95th percentile real EUR wealth.
	SpendingP5  *Number `json:"spending_p5,omitempty"`  // 5th percentile real EUR household spending during this year.
	SpendingP50 *Number `json:"spending_p50,omitempty"` // Median real EUR household spending during this year.
	SpendingP95 *Number `json:"spending_p95,omitempty"` // 95th percentile real EUR household spending during this year.
}

// OutcomeMetrics preserves undefined populations as null. Effective tax is the
// median tax/gross withdrawal ratio among paths with positive gross withdrawals;
// decade p5 includes only complete decades; CDaR averages worst 5% path drawdowns.
type OutcomeMetrics struct {
	EffectiveTaxRatePct     Number `json:"effective_tax_rate_pct"`     // Median tax/gross-withdrawal ratio over positive-gross paths, percent.
	MedianCumulativeRealTax Number `json:"median_cumulative_real_tax"` // Median cumulative real EUR tax across all sampled paths.
	CDaRPct                 Number `json:"cdar_pct"`                   // Mean of the worst five percent of path drawdowns, percent loss magnitude.
	WorstDecadeP5Pct        Number `json:"worst_decade_p5_pct"`        // 5th percentile of defined paths' worst complete ten-year real CAGR, percent.
}

// SolveData reports a verified feasible candidate or an explicit bounded outcome.
// Sampling uncertainty is descriptive and conditional on the model and selected draws.
type SolveData struct {
	Model                        ModelInfo      `json:"model"`                            // One return model, default central, or its effective model identity.
	SolveFor                     string         `json:"solve_for"`                        // Exactly annual_spending or capital; only fixed_real policy is supported.
	Status                       string         `json:"status"`                           // solved, boundary_satisfies or target_unreachable_in_bounds.
	Solution                     Number         `json:"solution"`                         // Verified feasible EUR candidate, null when unreachable.
	RuinPct                      Number         `json:"ruin_pct"`                         // Sampled ruin probability in percent; null without a solution.
	TargetRuinPct                float64        `json:"target_ruin_pct"`                  // Maximum tolerated sampled ruin percentage, default 5; explicit zero is preserved.
	Bounds                       Bounds         `json:"bounds"`                           // Explicit search interval; defaults are reported when omitted.
	LowerRuinPct                 Number         `json:"lower_ruin_pct"`                   // Ruin percentage evaluated at the original lower bound.
	UpperRuinPct                 Number         `json:"upper_ruin_pct"`                   // Ruin percentage evaluated at the original upper bound.
	FinalBounds                  Bounds         `json:"final_bounds"`                     // Final bisection bracket, or original interval for boundary outcomes.
	BracketWidth                 float64        `json:"bracket_width"`                    // Attained search interval width in EUR.
	Tolerance                    float64        `json:"tolerance"`                        // Effective stopping tolerance in EUR.
	Iterations                   int            `json:"iterations"`                       // Number of bisection halvings performed.
	ProbabilityResolutionPct     float64        `json:"probability_resolution_pct"`       // One sampled failure expressed in percentage points.
	RuinMCStandardErrorPP        Number         `json:"ruin_mc_standard_error_pp"`        // Descriptive plug-in Monte Carlo standard error in percentage points.
	RequestedPolicy              string         `json:"requested_policy"`                 // Policy explicitly requested by the caller or its resolved default.
	SolvedPolicy                 string         `json:"solved_policy"`                    // Policy actually used by the fixed-rule solve.
	PolicyChangedForSafeSpending bool           `json:"policy_changed_for_safe_spending"` // True only when the fixed-real reference differs from an adaptive request.
	EffectivePlan                *EffectivePlan `json:"effective_plan"`                   // Candidate plan that reproduces the solution; null when unreachable.
	WithdrawalRatePct            Number         `json:"withdrawal_rate_pct"`              // Candidate annual spending divided by capital, in percent.
}

// FireModelData is one requested model evaluated on a single shared ensemble.
type FireModelData struct {
	Model                 ModelInfo       `json:"model"`                     // One return model, default central, or its effective model identity.
	RuinPct               Number          `json:"ruin_pct"`                  // Sampled ruin probability in percent; null without a solution.
	TerminalRealWealthP5  Number          `json:"terminal_real_wealth_p5"`   // 5th percentile real EUR wealth at horizon.
	TerminalRealWealthP50 Number          `json:"terminal_real_wealth_p50"`  // Median real EUR wealth at horizon.
	Spending              SpendingSummary `json:"spending"`                  // Household spending and conditional policy-cut statistics.
	AnnualBands           []AnnualBands   `json:"annual_bands,omitempty"`    // Optional annual wealth/spending quantiles, at most 101 rows.
	OutcomeMetrics        *OutcomeMetrics `json:"outcome_metrics,omitempty"` // Optional tax, drawdown and complete-decade measurements.
	SafeSpending          *SolveData      `json:"safe_spending,omitempty"`   // Optional per-model fixed-real spending reference; requires its selected section.
}

// SimulationData contains all requested models, in request order.
type SimulationData struct {
	Models []FireModelData `json:"models"` // Distinct requested models, default central; at most six.
}
