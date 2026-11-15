package remote

import "encoding/json"

// Warning qualifies a result without silently changing its assumptions.
type Warning struct {
	Code    string `json:"code"`           // Stable machine-readable outcome or warning code.
	Message string `json:"message"`        // English explanation safe to return to remote callers.
	Path    string `json:"path,omitempty"` // Input or output location qualified by this entry.
}

// Default records an omitted input and the value used by this build.
type Default struct {
	Path  string `json:"path"`  // Input or output location qualified by this entry.
	Value any    `json:"value"` // Resolved value of this field.
}

// Window is an inclusive UTC observation interval. Empty dates are unbounded.
type Window struct {
	From string `json:"from,omitempty"` // Inclusive UTC start date, YYYY-MM-DD; empty is unbounded.
	To   string `json:"to,omitempty"`   // Inclusive UTC end date, YYYY-MM-DD; empty is unbounded.
}

// Meta identifies the engine, data, units, selected sections and applied defaults.
type Meta struct {
	CAPE            *CAPEData `json:"cape,omitempty"`             // Bundled valuation observation when CAPE adjustment was requested.
	SchemaVersion   string    `json:"schema_version"`             // Public contract revision, independent of numerical corrections.
	EngineVersion   string    `json:"engine_version"`             // Build or module revision used for this result.
	DatasetRevision string    `json:"dataset_revision"`           // Fingerprint of embedded reference data.
	ComputedAt      string    `json:"computed_at"`                // UTC RFC3339 computation start time.
	DurationMS      int64     `json:"duration_ms"`                // Execution wall time in milliseconds, excluding transport admission.
	Currency        string    `json:"currency,omitempty"`         // Evaluation currency, distinct from economic currency exposure.
	RequestedWindow *Window   `json:"requested_window,omitempty"` // Requested inclusive observation interval.
	EffectiveWindow *Window   `json:"effective_window,omitempty"` // Actual common observation interval.
	Sections        []string  `json:"sections"`                   // Distinct requested result sections; defaults depend on the operation.
	Defaulted       []Default `json:"defaulted"`                  // Every omitted optional input resolved to a concrete value.
	Assumptions     any       `json:"assumptions,omitempty"`      // Effective analysis or FIRE inputs and conventions.
}

// Result is the shared HTTP/MCP domain envelope. Data is one operation's typed output.
type Result struct {
	Meta     Meta      `json:"meta"`     // Versions, units, observation windows and applied defaults.
	Data     any       `json:"data"`     // The complete typed operation output.
	Warnings []Warning `json:"warnings"` // All substantive qualifications, including source warnings.
}

// Error is a public, credential-free operation failure; bounds without a solution are results.
type Error struct {
	Code              string `json:"code"`                          // Stable machine-readable outcome or warning code.
	Message           string `json:"message"`                       // English explanation safe to return to remote callers.
	Path              string `json:"path,omitempty"`                // Input or output location qualified by this entry.
	Retryable         bool   `json:"retryable"`                     // Whether the same input may succeed after a temporary condition clears.
	RetryAfterSeconds int    `json:"retry_after_seconds,omitempty"` // Suggested retry delay in seconds, zero when unspecified.
}

// Error returns the public message.
func (e *Error) Error() string { return e.Message }

// ErrorResult is the shared failure envelope.
type ErrorResult struct {
	Error *Error `json:"error"` // Public failure details; no internal credentials or paths.
}

// Number serializes undefined floating-point statistics as JSON null.
type Number float64

// MarshalJSON represents NaN and infinities as null.
func (n Number) MarshalJSON() ([]byte, error) {
	v := float64(n)
	if v != v || v > 1.7976931348623157e308 || v < -1.7976931348623157e308 {
		return []byte("null"), nil
	}
	return json.Marshal(v)
}

// SearchInput browses the embedded catalog without network access.
type SearchInput struct {
	Query      string `json:"query,omitempty"`       // Case-insensitive catalog ID, alias or name query; at most 200 characters.
	AssetClass string `json:"asset_class,omitempty"` // Catalog classification filter or resolved classification.
	Currency   string `json:"currency,omitempty"`    // Evaluation currency, distinct from economic currency exposure.
	Limit      *int   `json:"limit,omitempty"`       // Search page size, 1 through 50; omitted means 10.
	Cursor     string `json:"cursor,omitempty"`      // Opaque continuation tied to the current build and filters.
}

// AnalysisInput contains shared analysis options. Percentages use 5 for five percent.
type AnalysisInput struct {
	Currency    *string  `json:"currency,omitempty"`     // Evaluation currency, distinct from economic currency exposure.
	From        string   `json:"from,omitempty"`         // Inclusive UTC start date, YYYY-MM-DD; empty is unbounded.
	To          string   `json:"to,omitempty"`           // Inclusive UTC end date, YYYY-MM-DD; empty is unbounded.
	History     *string  `json:"history,omitempty"`      // Observed prices by default, or explicitly extended reconstructed history.
	DataPolicy  *string  `json:"data_policy,omitempty"`  // cached_or_fetch by default; cache_only forbids all network access.
	BenchmarkID string   `json:"benchmark_id,omitempty"` // Optional exact benchmark ID, evaluated on its own overlap.
	Sections    []string `json:"sections,omitempty"`     // Distinct requested result sections; defaults depend on the operation.
}

// AssetsInput analyzes one to ten exact identifiers; it never accepts paths or URLs.
type AssetsInput struct {
	AnalysisInput
	IDs []string `json:"ids"` // One through ten exact asset identifiers.
}

// Holding is a fully invested allocation weight in percent.
type Holding struct {
	ID        string   `json:"id"`         // Canonical or explicitly requested asset/model identifier.
	WeightPct *float64 `json:"weight_pct"` // Nonnegative allocation percentage; all holdings sum to 100.
}

// Allocation defines holdings and optional additional annual charges in percent.
type Allocation struct {
	Label         string    `json:"label,omitempty"`          // Optional caller label, at most 80 characters.
	Holdings      []Holding `json:"holdings"`                 // One through twenty fully invested holdings, or requested holding studies.
	RebalanceDays *int      `json:"rebalance_days,omitempty"` // Calendar-day interval; zero is buy-and-hold, default 90.
	AnnualFeePct  *float64  `json:"annual_fee_pct,omitempty"` // Additional annual portfolio charge in percent, default zero.
}

// PortfoliosInput compares one to four allocations over one common holding window.
type PortfoliosInput struct {
	AnalysisInput
	Portfolios []Allocation `json:"portfolios"` // One through four allocations or their complete studies.
}

// Identity contains public catalog fields, without provider-internal identifiers.
type Identity struct {
	ID                       string   `json:"id"`                         // Canonical or explicitly requested asset/model identifier.
	Name                     string   `json:"name"`                       // Human-readable name.
	ISIN                     string   `json:"isin,omitempty"`             // International security identifier when known.
	AssetClass               string   `json:"asset_class,omitempty"`      // Catalog classification filter or resolved classification.
	QuoteCurrency            string   `json:"quote_currency,omitempty"`   // Native quote currency, before evaluation conversion.
	TERPct                   *float64 `json:"ter_pct"`                    // Published annual asset charge in percent; null when unknown.
	ExtendedHistoryAvailable bool     `json:"extended_history_available"` // Whether the build contains a reconstructed history for this asset.
}

// SearchData is a deterministic catalog page. An empty cursor ends pagination.
type SearchData struct {
	Items      []Identity `json:"items"`                 // Catalog identities in deterministic relevance/ID order.
	NextCursor string     `json:"next_cursor,omitempty"` // Continuation cursor; empty means the last page.
}

// Statistics uses nominal returns, zero risk-free rate, and the series' measured cadence.
type Statistics struct {
	CAGR           Number  `json:"cagr_pct"`         // Geometric annual growth in percent.
	Volatility     Number  `json:"volatility_pct"`   // Annualized periodic-return standard deviation in percent.
	MaxDrawdown    Number  `json:"max_drawdown_pct"` // Deepest peak-to-trough loss in percent, negative or zero.
	Sharpe         Number  `json:"sharpe"`           // Annualized excess mean divided by volatility, risk-free zero.
	Sortino        Number  `json:"sortino"`          // Annualized excess mean divided by downside deviation, target zero.
	Years          float64 `json:"span_years"`       // Observation duration in 365.25-day years.
	Observations   int     `json:"observations"`     // Count of evaluated level observations.
	PeriodsPerYear float64 `json:"periods_per_year"` // Measured cadence used to annualize statistics.
}

// AnnualReturn is a calendar-year return; dates disclose partial final years.
type AnnualReturn struct {
	Window
	ReturnPct Number `json:"return_pct"` // Simple calendar-period return in percent.
	Partial   bool   `json:"partial"`    // First period begins at the series start rather than a prior year end.
}

// Drawdown is one episode, sorted deepest first. Recovery is absent when ongoing.
type Drawdown struct {
	Peak     string `json:"peak"`               // Peak observation date, YYYY-MM-DD.
	Trough   string `json:"trough"`             // Trough observation date, YYYY-MM-DD.
	Recovery string `json:"recovery,omitempty"` // Recovery observation date; absent while ongoing.
	DepthPct Number `json:"depth_pct"`          // Peak-to-trough loss in percent, negative or zero.
	Ongoing  bool   `json:"ongoing"`            // Whether the episode remains unrecovered at the last observation.
}

// AssetMetadata describes exposure percentages and their observation date.
type AssetMetadata struct {
	GeographyPct        map[string]float64 `json:"geography_pct,omitempty"`         // Geographic exposure percentages.
	SectorsPct          map[string]float64 `json:"sectors_pct,omitempty"`           // Sector percentages of the equity sleeve.
	CurrencyExposurePct map[string]float64 `json:"currency_exposure_pct,omitempty"` // Economic currency exposure percentages, distinct from quote currency.
	AsOf                string             `json:"as_of,omitempty"`                 // Source observation date or metadata month.
	Notes               string             `json:"notes,omitempty"`                 // Source qualifications and descriptive notes.
	Sources             []string           `json:"sources,omitempty"`               // Public source attribution links.
}

// Provenance distinguishes requested identity, resolution and reconstructed history.
type Provenance struct {
	ResolvedID          string `json:"resolved_id"`                    // Actual resolved series identifier.
	Source              string `json:"source"`                         // Series provider or bundled reconstruction source.
	Currency            string `json:"currency,omitempty"`             // Evaluation currency, distinct from economic currency exposure.
	Window              Window `json:"window"`                         // Actual inclusive observation interval.
	FullyReconstructed  bool   `json:"fully_reconstructed"`            // True when no observed quote segment exists.
	ReconstructedBefore string `json:"reconstructed_before,omitempty"` // First observed quote date after reconstructed history, when known.
	ProxyID             string `json:"proxy_id,omitempty"`             // Reconstruction proxy identifier when known.
}

// RelativeStatistics is measured only on the reported benchmark overlap.
type RelativeStatistics struct {
	BenchmarkID      string `json:"benchmark_id"`      // Optional exact benchmark ID, evaluated on its own overlap.
	Window           Window `json:"window"`            // Actual inclusive observation interval.
	Beta             Number `json:"beta"`              // Sensitivity to benchmark returns, unitless.
	AlphaPct         Number `json:"alpha_pct"`         // Annualized Jensen alpha in percent.
	InformationRatio Number `json:"information_ratio"` // Annualized mean active return divided by tracking error.
	UpCapturePct     Number `json:"up_capture_pct"`    // Geometric capture on benchmark up periods, in percent.
	DownCapturePct   Number `json:"down_capture_pct"`  // Geometric capture on benchmark down periods, in percent.
}

// AssetData is a compact projection; requested details never hide provenance or warnings.
type AssetData struct {
	RequestedID   string              `json:"requested_id"`             // Identifier exactly as supplied by the caller.
	Identity      *Identity           `json:"identity,omitempty"`       // Selected public asset identity.
	Provenance    *Provenance         `json:"provenance,omitempty"`     // Required source, observation and reconstruction qualifications.
	Statistics    *Statistics         `json:"statistics,omitempty"`     // Selected compact nominal risk/return measurements.
	Metadata      *AssetMetadata      `json:"metadata,omitempty"`       // Selected catalog exposures and source dates.
	AnnualReturns []AnnualReturn      `json:"annual_returns,omitempty"` // At most 150 calendar-year rows.
	Drawdowns     *[]Drawdown         `json:"drawdowns,omitempty"`      // At most ten episodes, deepest first.
	Relative      *RelativeStatistics `json:"relative,omitempty"`       // Benchmark overlap and relative measurements; undefined values are null.
}

// AssetsData contains every requested asset in input order.
type AssetsData struct {
	Assets []AssetData `json:"assets"` // Requested asset studies in input order.
}

// Composition reports notional exposures; sectors use the equity sleeve denominator.
type Composition struct {
	AssetClassPct   map[string]float64 `json:"asset_class_pct"`       // Notional capital exposure percentages by asset class.
	GeographyPct    map[string]float64 `json:"geography_pct"`         // Geographic exposure percentages.
	CurrencyPct     map[string]float64 `json:"currency_exposure_pct"` // Economic currency exposure percentages, distinct from quote currency.
	SectorsPct      map[string]float64 `json:"equity_sectors_pct"`    // Sector percentages with the equity sleeve as denominator.
	EquityPct       float64            `json:"equity_sleeve_pct"`     // Equity notional as a percentage of total capital.
	CoveragePct     map[string]float64 `json:"coverage_pct"`          // Classified percentage for each composition dimension.
	UnclassifiedPct float64            `json:"unclassified_pct"`      // Capital percentage without catalog classification.
}

// Attribution reports monthly covariance risk shares and simple contribution return shares.
type Attribution struct {
	Frequency string   `json:"frequency"`        // Measurement frequency, monthly for attribution.
	RiskPct   []Number `json:"risk_share_pct"`   // Euler shares of monthly contribution variance, in percent.
	ReturnPct []Number `json:"return_share_pct"` // Shares of summed monthly return contributions, in percent.
}

// PortfolioData reports one allocation on the shared effective window.
type PortfolioData struct {
	Label         string              `json:"label,omitempty"`          // Optional caller label, at most 80 characters.
	Holdings      []Holding           `json:"effective_holdings"`       // Normalized effective holdings in caller order.
	RebalanceDays int                 `json:"rebalance_days"`           // Calendar-day interval; zero is buy-and-hold, default 90.
	AnnualFeePct  float64             `json:"annual_fee_pct"`           // Additional annual portfolio charge in percent, default zero.
	Statistics    *Statistics         `json:"statistics,omitempty"`     // Selected compact nominal risk/return measurements.
	Composition   *Composition        `json:"composition,omitempty"`    // Selected look-through exposure and coverage data.
	Assets        []AssetData         `json:"holdings,omitempty"`       // One through twenty fully invested holdings, or requested holding studies.
	Correlations  [][]Number          `json:"correlations,omitempty"`   // Pearson matrix on shared monthly intervals, in holding order.
	Attribution   *Attribution        `json:"attribution,omitempty"`    // Selected monthly risk and return contribution shares.
	AnnualReturns []AnnualReturn      `json:"annual_returns,omitempty"` // At most 150 calendar-year rows.
	Drawdowns     *[]Drawdown         `json:"drawdowns,omitempty"`      // At most ten episodes, deepest first.
	Relative      *RelativeStatistics `json:"relative,omitempty"`       // Benchmark overlap and relative measurements; undefined values are null.
	Provenance    []Provenance        `json:"provenance"`               // Required source, observation and reconstruction qualifications.
}

// PortfoliosData contains every requested allocation, with no partial comparisons.
type PortfoliosData struct {
	Portfolios []PortfolioData `json:"portfolios"` // One through four allocations or their complete studies.
}

// AnalysisAssumptions names nominal return and measurement conventions.
type AnalysisAssumptions struct {
	ReturnBasis          string  `json:"return_basis"`                    // ReturnBasis is nominal for all analysis operations.
	RiskFreePct          float64 `json:"risk_free_pct"`                   // RiskFreePct is zero for unadjusted Sharpe and Sortino.
	CorrelationFrequency string  `json:"correlation_frequency,omitempty"` // CorrelationFrequency is monthly for portfolios.
	BenchmarkID          string  `json:"benchmark_id,omitempty"`          // BenchmarkID is optional and uses its own overlap.
}
