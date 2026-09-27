package simgen

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// Fetcher provides price histories. A batch generator has no per-request
// cancellation to gain, so the interface stays context-free: wrap a
// *marketdata.Client with WithContext to satisfy it.
type Fetcher interface {
	Fetch(id string, from time.Time) (*marketdata.Series, error)
}

// WithContext adapts a marketdata.Client to the Fetcher interface, binding
// every Fetch to ctx.
func WithContext(ctx context.Context, c *marketdata.Client) Fetcher {
	return ctxFetcher{ctx: ctx, c: c}
}

type ctxFetcher struct {
	ctx context.Context
	c   *marketdata.Client
}

// Fetch serves real quotes only: a nowcast tail (Series.EstimatedFrom) is an
// estimate, and nothing a recipe splices, validates against or ships may
// carry one.
func (f ctxFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	s, err := f.c.Fetch(f.ctx, id, from)
	if err != nil {
		return nil, err
	}
	return s.WithoutEstimates(), nil
}

// Recipe describes how to rebuild one asset's past.
type Recipe struct {
	ID     string // canonical identifier the simdata file extends
	Name   string // display name for the simdata header
	Method string // one-line description of the construction

	// Build assembles the simulated series from component histories.
	Build func(f Fetcher, from time.Time) (*marketdata.Series, error)

	// ValidateAgainst is the identifier of the real series used for the
	// overlap check (often the asset itself, or its US-listed twin).
	ValidateAgainst string

	// SpliceReal, when non-empty, grafts this real series on top of the
	// composite so the simdata file carries real data wherever available.
	SpliceReal string

	// Donors lists the records this recipe splices behind the asset, nearest
	// trade first, for the recipes built as a donor chain. It is declared for
	// the audit report (Audit), which grades every junction of the chain on
	// its own overlap; a recipe passes the same slice to its Build, so the two
	// cannot drift apart. Empty for the recipes that are not chains.
	Donors []string
}

// Frame holds daily returns of several components aligned on the dates where
// every component trades (forward-filled in between).
type Frame struct {
	Dates   []time.Time
	Returns map[string][]float64 // same length as Dates; Returns[id][0] is always 0
}

// BuildFrame fetches every id and aligns daily returns on the union of
// trading dates from the latest first-quote on. Rate ids (the Yahoo yield
// symbols ^IRX, ^FVX, ^TNX and ^TYX, the policy and money-market family of
// marketdata.RateSymbols, and the spliced financing rate usdOvernight) are
// treated as annualized percent levels and converted to daily accruals instead
// of price returns.
func BuildFrame(f Fetcher, ids []string, from time.Time) (*Frame, error) {
	series := make(map[string]*marketdata.Series, len(ids))
	start := from
	for _, id := range ids {
		s, err := f.Fetch(id, from)
		if err != nil {
			return nil, fmt.Errorf("component %s: %w", id, err)
		}
		if s == nil || len(s.Points) < 2 {
			return nil, fmt.Errorf("component %s: empty history", id)
		}
		series[id] = s
		if fd := s.Points[0].Date; fd.After(start) {
			start = fd
		}
	}

	uniqueIDs := make([]string, 0, len(series))
	ordered := make([]*marketdata.Series, 0, len(series))
	seen := map[string]bool{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			uniqueIDs = append(uniqueIDs, id)
			ordered = append(ordered, series[id])
		}
	}
	dates, levels := marketdata.Align(ordered, start, time.Time{})
	if len(dates) < 2 {
		return nil, fmt.Errorf("not enough common dates")
	}

	fr := &Frame{Dates: dates, Returns: make(map[string][]float64, len(uniqueIDs))}
	for i, id := range uniqueIDs {
		lv := levels[i]
		ret := make([]float64, len(dates))
		if isRate(id) {
			// Annualized percent level → daily accrual.
			for k := 1; k < len(dates); k++ {
				ret[k] = lv[k-1] / 100 / 252
			}
		} else {
			for k := 1; k < len(dates); k++ {
				ret[k] = lv[k]/lv[k-1] - 1
			}
		}
		fr.Returns[id] = ret
	}
	return fr, nil
}

// isRate reports whether an identifier is a yield series quoted in
// annualized percent (Yahoo's ^IRX, ^TNX, …) rather than a price. The policy
// and money-market symbols answer through marketdata's own registry rather
// than a second list here, so a rate added there is never read as a price.
func isRate(id string) bool {
	switch id {
	case "^IRX", "^FVX", "^TNX", "^TYX", usdOvernight:
		return true
	}
	return marketdata.RateName(id) != ""
}

// Leg is one exposure of a linear composite.
type Leg struct {
	ID     string
	Weight float64
	Excess bool // futures-like: earns Weight×(return − cash)
}

// Composite builds an index (base 100) from constant daily-rebalanced legs.
// cashID (e.g. "^IRX") backs both the Excess financing and an optional
// collateral leg.
//
// annualFee is charged PER FRAME STEP at a 252-step year, the same convention
// the excess legs finance at, so it is only a yearly charge on a frame whose
// steps are trading days. Hand it a weekly or monthly frame and the fee shrinks
// with the step count: twelve monthly steps carry a twentieth of it. Every
// caller here passes a daily frame, and the fee schedules that have to survive a
// coarse era go through afterFeeSteps instead, which compounds on calendar days
// and is exact at any cadence.
func Composite(fr *Frame, legs []Leg, cashID string, annualFee float64) ([]float64, error) {
	cash := fr.Returns[cashID]
	for _, l := range legs {
		if _, ok := fr.Returns[l.ID]; !ok {
			return nil, fmt.Errorf("component %s missing from frame", l.ID)
		}
		if l.Excess && cash == nil {
			return nil, fmt.Errorf("cashID required for excess leg %s", l.ID)
		}
	}
	values := make([]float64, len(fr.Dates))
	values[0] = 100
	feeDaily := annualFee / 252
	for k := 1; k < len(fr.Dates); k++ {
		r := -feeDaily
		for _, l := range legs {
			lr := fr.Returns[l.ID][k]
			if l.Excess {
				lr -= cash[k]
			}
			r += l.Weight * lr
		}
		values[k] = values[k-1] * (1 + r)
	}
	return values, nil
}

// Validation summarizes how well a simulated series tracks the real one over
// their overlap: the dates both quote, a marketdata.Panel at Daily, so every
// return of either side spans the same sessions.
type Validation struct {
	Overlap     int // number of common returns (daily for two daily series)
	Start, End  time.Time
	Corr        float64 // correlation of those returns
	WeeklyCorr  float64 // correlation of five-session returns (kinder to stale quotes); zero under twelve of them
	Beta        float64 // slope of sim on real
	TrackingErr float64 // sample stdev of (sim - real) returns, annualized at the common calendar's cadence
	VolSim      float64 // sample volatility of each side on that calendar, annualized
	VolReal     float64
	CAGRSim     float64 // over the common window, 365.25-day years
	CAGRReal    float64
}

// String renders the validation as a one-line summary (daily/weekly
// correlation, beta, tracking error, sim vs real CAGR, and overlap window).
func (v Validation) String() string {
	return fmt.Sprintf("corr=%.3f (weekly %.3f) beta=%.2f TE=%.1f%%/yr CAGR sim %.2f%% vs real %.2f%% (overlap %d d from %s to %s)",
		v.Corr, v.WeeklyCorr, v.Beta, v.TrackingErr*100, v.CAGRSim*100, v.CAGRReal*100,
		v.Overlap, v.Start.Format("2006-01-02"), v.End.Format("2006-01-02"))
}

// minValidationPoints is the fewest dates two series must share for
// Validate to measure anything.
const minValidationPoints = 60

// Validate compares a simulated series with the real one on the dates both
// quote. It is an error when they share fewer than sixty dates, and when
// either holds a close that is not a positive price on one of them.
func Validate(sim, real *marketdata.Series) (Validation, error) {
	p, err := pairPanel(marketdata.Daily, sim, real)
	if err != nil {
		return Validation{}, fmt.Errorf("simgen: Validate: %w", err)
	}
	if p.Len()+1 < minValidationPoints {
		return Validation{}, fmt.Errorf("simgen: Validate: insufficient overlap (%d common points)", p.Len()+1)
	}
	t, err := p.Track(columnA, columnB)
	if err != nil {
		return Validation{}, fmt.Errorf("simgen: Validate: %w", err)
	}
	v := Validation{
		Overlap: t.Periods, Start: p.Starts[0], End: p.Ends[p.Len()-1],
		Corr: t.Corr, Beta: t.Beta, TrackingErr: t.TrackingError, VolSim: t.VolA, VolReal: t.VolB,
	}
	days := int(v.End.Sub(v.Start).Hours() / 24)
	v.CAGRSim = metrics.Annualize(growth(p.R[0])-1, days)
	v.CAGRReal = metrics.Annualize(growth(p.R[1])-1, days)
	if w, err := p.Compound(5); err == nil && w.Len() >= 12 {
		v.WeeklyCorr = metrics.Corr(w.R[0], w.R[1])
	}
	return v, nil
}

// The column names pairPanel gives its two series: the Symbols of an engine
// and of the quotes it rebuilds may well be equal.
const (
	columnA = "A"
	columnB = "B"
)

// pairPanel puts two series, a reconstruction and its reference or two
// links of a donor chain, on one calendar of f, as columnA and columnB.
func pairPanel(f marketdata.Frequency, a, b *marketdata.Series) (*marketdata.Panel, error) {
	if a.Len() == 0 || b.Len() == 0 {
		return nil, errors.New("empty series")
	}
	ca, cb := *a, *b
	ca.Symbol, cb.Symbol = columnA, columnB
	return marketdata.NewPanel(f, &ca, &cb)
}

// growth is the growth factor returns compound to.
func growth(returns []float64) float64 {
	g := 1.0
	for _, r := range returns {
		g *= 1 + r
	}
	return g
}

// seriesFromFrame packages composite values as a marketdata series.
func seriesFromFrame(name string, fr *Frame, values []float64) *marketdata.Series {
	s := &marketdata.Series{Name: name, Source: "simdata"}
	for i := range fr.Dates {
		s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[i], Close: values[i]})
	}
	return s
}

// WithRefData returns a Fetcher that serves series found in fsys (CSV files in
// the simdata format: the bundled datasets.Refdata, or an os.DirFS of local
// reference series used in development via -refdata) before falling back to the
// wrapped fetcher.
//
// A reference id that fsys does not hold falls through, which is how a recipe
// reaches ordinary quotes. A reference id it DOES hold but cannot read does
// not: the read error is returned. The distinction matters because these ids
// (EM-USD, SP500-USD, TREND-NET-USD, …) look nothing like a ticker, so the
// fallback resolves them by fuzzy search, and a fuzzy search always finds
// something. EM-USD matches a crypto token named "Eminer USD" and SP500-USD a
// derivatives index, both of which a build would then splice in silence where
// its reference belonged. A malformed bundled file is a defect to report, never
// a reason to go looking on the network.
func WithRefData(fsys fs.FS, fallback Fetcher) Fetcher {
	return refFetcher{fsys: fsys, fallback: fallback}
}

type refFetcher struct {
	fsys     fs.FS
	fallback Fetcher
}

func (r refFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	s, ok, err := marketdata.ReadSimdataFS(r.fsys, id)
	if err != nil {
		return nil, fmt.Errorf("reference %s: %w", id, err)
	}
	if ok {
		return s, nil
	}
	return r.fallback.Fetch(id, from)
}

// Splice returns the real series extended backwards by the simulated
// composite: real quotes wherever they exist, rescaled simulation before.
func Splice(real, sim *marketdata.Series) *marketdata.Series {
	out := &marketdata.Series{Symbol: real.Symbol, Name: real.Name, Currency: real.Currency, Source: "simdata"}
	out.Points = append(out.Points, real.Points...)
	marketdata.ExtendBack(out, sim)
	return out
}
