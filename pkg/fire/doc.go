// Package fire prepares shared fixed-horizon retirement models for numerical
// callers and the web simulator. It does not fetch quotes or simulate inside
// transport adapters: nominal histories enter BuildMonthlyPanel, fitted annual
// parameters enter PrepareModel, and BuildPlan produces a decumul.Plan.
//
// # Start here
//
// Start from Defaults, set PlanSpec.Capital, AnnualSpending and Years, then call
// BuildPlan. Defaults are current assumptions, including a labeled French
// taxable-account planning rate, not universal legal rates. Explicit zeros are
// preserved. Set the returned plan's Source from PrepareModel and use decumul's
// context-aware drawing, simulation and structured solving methods.
// Server handlers use Acquire to share process-wide computation capacity;
// it queues concurrent page sections until admitted or cancelled.
//
// All money is in starting purchasing power. Rates and weights are fractions;
// Mu is an arithmetic annual real mean, Sigma annual volatility, and Nu the
// Student-t degrees of freedom. BuildMonthlyPanel expects nominal prices already
// converted to one currency and an inflation index. It keeps one uninterrupted
// common monthly window; Coverage describes exclusions. FitParametric estimates
// annual parameters, while CentralParameters applies short-window blending or
// the optional CAPE anchor.
//
// PrepareModel exposes Central, Bootstrap, Historical, Stress, BroadSample and
// LostDecade without silently substituting an unavailable model. Annual kernels
// use compounded monthly scenarios for the two portfolio-history models.
// Historical samples start windows uniformly with replacement; path count is
// not a count of independent historical retirements. BroadSample is an external
// domestic 60/40 reference and does not represent the supplied portfolio.
//
// The supported plan policies are FixedReal and GuytonKlinger. Adaptive capital
// solving is not monotonic even when initial bands are rebuilt correctly; use
// simulation to compare adaptive capital choices. A fixed-rule safe-spending
// result is a policy reference, not an adaptive policy's guaranteed budget.
package fire
