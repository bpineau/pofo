package compare

import (
	"html/template"
	"io/fs"
	"time"

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
	NoFees    bool  // ignore envelope fees
	Simdata   fs.FS // optional filesystem of simulated-history CSVs
	Framework suggest.Framework
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
}

// column is the full per-portfolio compute record produced by Compute (the
// former cmd/pofo result struct). It keeps every intermediate a renderer might
// need; Comparison exposes only the narrow Column view publicly.
type column struct {
	p             *portfolio.Portfolio
	sim           *portfolio.SimResult
	color         string
	rebalanceDays int
	currency      string // base currency this column was evaluated in
	specName      string // the spec this column came from (p.Name may be decorated: currency tag, "as written")
	note          string // informational line (e.g. optimizer choice)
	// Common-window view, renormalized to 100, used for stats and comparison.
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

// Columns returns the narrow public view of every compared portfolio, in the
// order they were computed.
func (c *Comparison) Columns() []Column {
	out := make([]Column, len(c.columns))
	for i, col := range c.columns {
		out[i] = Column{
			Name: col.p.Name, Color: col.color,
			SimDates: col.sim.Dates, SimValues: col.sim.Values,
			WinDates: col.winDates, WinValues: col.winValues,
			Assets: col.p.Assets, Stats: col.stats,
		}
	}
	return out
}
