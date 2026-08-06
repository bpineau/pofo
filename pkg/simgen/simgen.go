package simgen

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"time"

	"github.com/bpineau/portfodor/pkg/marketdata"
	"github.com/bpineau/portfodor/pkg/metrics"
)

// ErrUnfaithful marks reconstructions whose fit against reality is too poor
// to be worth storing; callers should treat it as a documented skip.
var ErrUnfaithful = errors.New("replication too unfaithful")

// Fetcher provides price histories; *marketdata.Client satisfies it.
type Fetcher interface {
	Fetch(id string, from time.Time) (*marketdata.Series, error)
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
}

// Frame holds daily returns of several components aligned on the dates where
// every component trades (forward-filled in between).
type Frame struct {
	Dates   []time.Time
	Returns map[string][]float64 // same length as Dates; Returns[id][0] is always 0
}

// BuildFrame fetches every id and aligns daily returns on the union of
// trading dates from the latest first-quote on. The rate ids ^IRX, ^FVX, ^TNX and
// ^TYX are treated as annualized percent levels and converted to daily
// accruals instead of price returns.
func BuildFrame(f Fetcher, ids []string, from time.Time) (*Frame, error) {
	series := make(map[string]*marketdata.Series, len(ids))
	start := from
	for _, id := range ids {
		s, err := f.Fetch(id, from)
		if err != nil {
			return nil, fmt.Errorf("component %s: %w", id, err)
		}
		if len(s.Points) < 2 {
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
// annualized percent (Yahoo's ^IRX, ^TNX, …) rather than a price.
func isRate(id string) bool {
	switch id {
	case "^IRX", "^FVX", "^TNX", "^TYX":
		return true
	}
	return false
}

// Leg is one exposure of a linear composite.
type Leg struct {
	ID     string
	Weight float64
	Excess bool // futures-like: earns Weight×(return − cash)
}

// Composite builds an index (base 100) from constant daily-rebalanced legs.
// cashID (e.g. "^IRX") backs both the Excess financing and an optional
// collateral leg; annualFee is deducted pro rata temporis.
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
// their overlap.
type Validation struct {
	Overlap     int // number of common daily returns
	Start, End  time.Time
	Corr        float64 // correlation of daily returns
	WeeklyCorr  float64 // correlation of 5-day returns (kinder to stale quotes)
	Beta        float64 // slope sim→real
	TrackingErr float64 // annualized stdev of (real − sim) daily returns
	CAGRSim     float64
	CAGRReal    float64
}

func (v Validation) String() string {
	return fmt.Sprintf("corr=%.3f (weekly %.3f) beta=%.2f TE=%.1f%%/yr CAGR sim %.2f%% vs real %.2f%% (overlap %d d from %s to %s)",
		v.Corr, v.WeeklyCorr, v.Beta, v.TrackingErr*100, v.CAGRSim*100, v.CAGRReal*100,
		v.Overlap, v.Start.Format("2006-01-02"), v.End.Format("2006-01-02"))
}

// Validate compares a simulated series with the real one on common dates.
func Validate(sim, real *marketdata.Series) (Validation, error) {
	realByDate := make(map[time.Time]float64, len(real.Points))
	for _, p := range real.Points {
		realByDate[p.Date] = p.Close
	}
	var dates []time.Time
	var sv, rv []float64
	for _, p := range sim.Points {
		if r, ok := realByDate[p.Date]; ok {
			dates = append(dates, p.Date)
			sv = append(sv, p.Close)
			rv = append(rv, r)
		}
	}
	if len(sv) < 60 {
		return Validation{}, fmt.Errorf("insufficient overlap (%d common points)", len(sv))
	}
	var v Validation
	v.Overlap = len(sv) - 1
	v.Start, v.End = dates[0], dates[len(dates)-1]
	years := v.End.Sub(v.Start).Hours() / 24 / 365.25
	v.CAGRSim = math.Pow(sv[len(sv)-1]/sv[0], 1/years) - 1
	v.CAGRReal = math.Pow(rv[len(rv)-1]/rv[0], 1/years) - 1

	srets := metrics.Returns(sv)
	rrets := metrics.Returns(rv)
	ms, mr := metrics.Mean(srets), metrics.Mean(rrets)
	var covSR, varS, varR, varDiff float64
	for i := range srets {
		ds, dr := srets[i]-ms, rrets[i]-mr
		covSR += ds * dr
		varS += ds * ds
		varR += dr * dr
		diff := rrets[i] - srets[i]
		varDiff += diff * diff
	}
	if varS > 0 && varR > 0 {
		v.Corr = covSR / math.Sqrt(varS*varR)
		v.Beta = covSR / varS
	}
	n := float64(len(srets))
	meanDiff := (mr - ms)
	v.TrackingErr = math.Sqrt(math.Max(0, varDiff/n-meanDiff*meanDiff)) * math.Sqrt(252)

	// Weekly (5 trading days) correlation.
	var sw, rw []float64
	for i := 5; i < len(sv); i += 5 {
		sw = append(sw, sv[i]/sv[i-5]-1)
		rw = append(rw, rv[i]/rv[i-5]-1)
	}
	if len(sw) >= 12 {
		msw, mrw := metrics.Mean(sw), metrics.Mean(rw)
		var cov, vs, vr float64
		for i := range sw {
			cov += (sw[i] - msw) * (rw[i] - mrw)
			vs += (sw[i] - msw) * (sw[i] - msw)
			vr += (rw[i] - mrw) * (rw[i] - mrw)
		}
		if vs > 0 && vr > 0 {
			v.WeeklyCorr = cov / math.Sqrt(vs*vr)
		}
	}
	return v, nil
}

// SeriesFromFrame packages composite values as a marketdata series.
func SeriesFromFrame(name string, fr *Frame, values []float64) *marketdata.Series {
	s := &marketdata.Series{Name: name, Source: "simdata"}
	for i := range fr.Dates {
		s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[i], Close: values[i]})
	}
	return s
}

// WithRefData returns a Fetcher that serves series found in fsys (CSV files
// in the simdata format — typically the embedded datasets.Refdata, or an
// os.DirFS for development) before falling back to the wrapped fetcher.
func WithRefData(fsys fs.FS, fallback Fetcher) Fetcher {
	return refFetcher{fsys: fsys, fallback: fallback}
}

type refFetcher struct {
	fsys     fs.FS
	fallback Fetcher
}

func (r refFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	if s, ok, err := marketdata.ReadSimdataFS(r.fsys, id); err == nil && ok {
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
