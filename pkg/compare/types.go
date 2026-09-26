package compare

import (
	"html/template"
	"io/fs"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// Options carries everything Compute needs beyond the specs themselves: the
// base currency and benchmark to evaluate against, the analysis window, the
// rebalancing cadence, the SIM/fee toggles, an optional embedded simdata
// filesystem, and the suggestion framework used for coverage/gap analysis.
type Options struct {
	Currency  string // base currency every column is evaluated in
	Benchmark string // identifier of the benchmark series, if any
	Start     time.Time
	End       time.Time
	Rebalance int   // rebalancing period in days
	NoSim     bool  // ignore SIM-suffixed simulated history extension
	NoFees    bool  // do not fetch the assets' ongoing charges (TER); envelope fees still apply
	Simdata   fs.FS // optional filesystem of simulated-history CSVs
	Framework suggest.Framework
	// ExactForeign resolves identifiers OUTSIDE the bundled catalog
	// exactly (marketdata.FetchOptions.ExactOnly): no instrument matched by
	// name, so a typo fails instead of quoting an unrelated fund. Catalog
	// identifiers are pinned already and unaffected. Callers that compare
	// portfolios composed by untrusted hands (the web app's p= grammar) set
	// it; the CLI leaves it off.
	ExactForeign bool
}

// Decoration carries optional presentation chrome injected into the rendered
// page: a skin stylesheet, site navigation and composer markup, and per-spec
// deep links into the FIRE UI. It is inert data; the renderer decides what to
// do with it.
type Decoration struct {
	SkinCSS  template.CSS      // extra stylesheet appended to the report skin
	SiteNav  template.HTML     // site navigation markup, if any
	Composer template.HTML     // composer widget markup, if any
	FireHref map[string]string // spec name -> FIRE deep link
}

// Column is the narrow public view of one compared portfolio: its identity,
// its full and common-window value series, its holdings, and its statistics.
// It is built from the internal column record by Comparison.Columns.
type Column struct {
	Name      string
	Color     string
	SimDates  []time.Time
	WinDates  []time.Time
	SimValues []float64
	WinValues []float64
	Assets    []portfolio.Asset
	Stats     metrics.Stats
	// Note is the informational line a computed column carries, currently
	// the optimizer's account of the weights it chose, the window it fitted
	// them on and how they did outside it. Empty for a plain portfolio.
	Note string
}

// column is the full per-portfolio compute record produced by Compute. Its
// numbers come from study, the column's analyze.PortfolioStudy: p and sim are
// that study's Portfolio and Sim (a test may set them without a study), and
// the composition and risk attribution are read off it. What the record adds
// is the comparison's own: the common-window view and its statistics, the
// real (deflated) ones, the identity color, the report's warnings.
// Comparison exposes only the narrow Column view publicly.
type column struct {
	study         *analyze.PortfolioStudy
	p             *portfolio.Portfolio
	sim           *portfolio.SimResult
	warnings      []string // the report's: the build's, the simulation's, a ruin
	color         string
	rebalanceDays int
	currency      string // base currency this column was evaluated in
	specName      string // the spec this column came from (p.Name may be decorated: currency tag, "as written")
	note          string // informational line (e.g. optimizer choice)
	// Common-window view, renormalized to 100: the comparison chart, the
	// drawdowns, the rolling and the real rows. stats, rel and vts are
	// measured on the same window of sim.Index itself (see column.measure).
	winDates  []time.Time
	winValues []float64
	stats     metrics.Stats
	realStats metrics.Stats // stats on the inflation-adjusted (deflated) window
	hasReal   bool
	rel       metrics.Relative
	hasRel    bool
	vts       metrics.VolTermStructure // daily/monthly volatility term structure
	hasVTS    bool
}

// Comparison is the computed comparison model: the aligned columns, the
// optional benchmark series, the common analysis window, the resolved asset
// metadata, and the options it was computed with. Its fields are private;
// callers read it through the accessors below.
type Comparison struct {
	columns     []*column
	bench       *marketdata.Series
	commonStart time.Time
	commonEnd   time.Time
	meta        map[string]suggest.Meta
	opt         Options
}

// CommonStart is the latest inception across the compared columns: the start of
// the window every column shares.
func (c *Comparison) CommonStart() time.Time { return c.commonStart }

// CommonEnd is the earliest last quote across the compared columns: the end of
// the window every column shares.
func (c *Comparison) CommonEnd() time.Time { return c.commonEnd }

// composition is the column's look-through, zero without a study.
func (r *column) composition() analyze.Composition {
	if r.study == nil {
		return analyze.Composition{}
	}
	return r.study.Composition
}

// attribution is the column's risk and return attribution, zero without a
// study.
func (r *column) attribution() metrics.Attribution {
	if r.study == nil {
		return metrics.Attribution{}
	}
	return r.study.Attribution
}

// Studies returns the numbers behind every column, in column order: one
// analyze.PortfolioStudy per compared portfolio, each on that column's OWN
// simulation window (Columns and StatRows read the common window instead).
// Every column Compute builds has one, the optimizer's column and each
// currency of a "#meta currencies" spec included: the study's Spec is then
// the column's (named as the column, the currency expanded, the optimizer's
// weights written in), so a study says exactly what its column shows. The
// benchmark is not a column and has no study. The studies are shared, not
// copied: treat them as read-only.
func (c *Comparison) Studies() []*analyze.PortfolioStudy {
	out := make([]*analyze.PortfolioStudy, len(c.columns))
	for i, col := range c.columns {
		out[i] = col.study
	}
	return out
}

// Columns returns the narrow public view of every compared portfolio, in the
// order they were computed.
func (c *Comparison) Columns() []Column {
	out := make([]Column, len(c.columns))
	for i, col := range c.columns {
		out[i] = Column{
			Name: col.p.Name, Color: col.color,
			SimDates: col.sim.Dates, SimValues: col.sim.Values,
			WinDates: col.winDates, WinValues: col.winValues,
			Assets: col.p.Assets, Stats: col.stats, Note: col.note,
		}
	}
	return out
}
