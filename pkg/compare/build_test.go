package compare

import (
	"reflect"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// newTestComparison assembles a Comparison from pre-built columns without
// fetching, so tests (and the characterization golden) can render a known
// report. Test-only: never widen the production API for this.
func newTestComparison(cols []*column, bench *marketdata.Series, start, end time.Time, meta map[string]suggest.Meta, opt Options) *Comparison {
	return &Comparison{columns: cols, bench: bench, commonStart: start, commonEnd: end, meta: meta, opt: opt}
}

// attachStudy gives a fabricated column the study analyze.Portfolio would
// carry for it, since the report reads a column's composition and risk
// attribution off its study: the look-through of its assets' weights and the
// attribution of its simulation's MONTHLY contributions, with the calls
// analyze makes. A column whose contributions change afterwards must be
// attached again.
func attachStudy(col *column, meta map[string]suggest.Meta) {
	h := holdingsFor(col.p.Assets, meta)
	comp := analyze.Composition{
		AssetClass: suggest.AssetClassSplit(h),
		Geography:  suggest.GeographySplit(h),
		Currency:   suggest.CurrencySplit(h),
		Duration:   suggest.DurationSplit(h),
	}
	comp.Sectors, comp.Equity = suggest.EquitySectorSplit(h)
	comp.Coverage, comp.Unclassified = suggest.Coverage(h, suggest.RegimeFramework())
	ps := &analyze.PortfolioStudy{Portfolio: col.p, Sim: col.sim, Composition: comp}
	if col.sim != nil {
		_, monthly := col.sim.MonthlyContributions()
		if att, err := metrics.Attribute(monthly); err == nil {
			ps.Attribution = att
		}
	}
	col.study = ps
}

// TestNewTestComparison round-trips a fully populated column through the
// builder and the public accessors, pinning the mapping from the internal
// record to the narrow Column view. It also exercises the moved series helpers
// so the foundation package stands on its own before Compute lands.
func TestNewTestComparison(t *testing.T) {
	dates := months(3)
	values := []float64{100, 110, 121}
	sim := &portfolio.SimResult{Dates: dates, Values: values}
	p := &portfolio.Portfolio{Name: "P", Assets: []portfolio.Asset{{Weight: 1}}}
	stats := metrics.Stats{CAGR: 0.1}

	col := &column{
		p: p, sim: sim, color: "#123456", rebalanceDays: 90,
		currency: "EUR", specName: "spec", note: "note",
		winDates: dates, winValues: values,
		stats: stats, realStats: metrics.Stats{CAGR: 0.08}, hasReal: true,
		rel: metrics.Relative{}, hasRel: true,
		vts: metrics.VolTermStructure{}, hasVTS: true,
	}
	bench := &marketdata.Series{Symbol: "BENCH"}
	meta := map[string]suggest.Meta{"P": {}}
	opt := Options{Currency: "EUR", Rebalance: 90}
	start, end := dates[0], dates[2]

	c := newTestComparison([]*column{col}, bench, start, end, meta, opt)

	if !c.CommonStart().Equal(start) || !c.CommonEnd().Equal(end) {
		t.Errorf("window = [%v, %v], want [%v, %v]", c.CommonStart(), c.CommonEnd(), start, end)
	}
	if c.bench != bench || c.opt.Rebalance != 90 || len(c.meta) != 1 {
		t.Errorf("bench/opt/meta not carried through: %v / %v / %d", c.bench, c.opt, len(c.meta))
	}
	cols := c.Columns()
	if len(cols) != 1 {
		t.Fatalf("Columns len = %d, want 1", len(cols))
	}
	got := cols[0]
	if got.Name != "P" || got.Color != "#123456" || got.Stats.CAGR != 0.1 {
		t.Errorf("Column view = %+v, want name P / color #123456 / CAGR 0.1", got)
	}
	if !reflect.DeepEqual(got.SimValues, values) || !reflect.DeepEqual(got.WinDates, dates) {
		t.Errorf("Column series not mapped through")
	}

	// The internal record keeps fields the narrow view drops; read them so the
	// foundation is self-contained (they feed the renderer in later tasks).
	if col.rebalanceDays != 90 || col.currency != "EUR" || col.specName != "spec" ||
		col.note != "note" || !col.hasReal || !col.hasRel || !col.hasVTS ||
		col.realStats.CAGR != 0.08 || col.rel.Alpha != 0 || col.vts.Ratio != 0 {
		t.Errorf("internal column record not preserved: %+v", col)
	}
}

// TestSeriesHelpers pins the small helpers (window, rebase, negate) Compute
// and the statistics table lean on.
func TestSeriesHelpers(t *testing.T) {
	dates := months(4)

	i, j := window(dates, dates[1], dates[2])
	if i != 1 || j != 3 {
		t.Errorf("window = [%d, %d), want [1, 3)", i, j)
	}

	if got := rebase([]float64{50, 75, 100}); got[0] != 100 || got[2] != 200 {
		t.Errorf("rebase = %v, want start 100 end 200", got)
	}

	if got := negate([]float64{1, -2, 3}); !reflect.DeepEqual(got, []float64{-1, 2, -3}) {
		t.Errorf("negate = %v, want [-1 2 -3]", got)
	}
}
