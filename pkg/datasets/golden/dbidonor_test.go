package golden

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// The DBi family's nearest donor is the published net all-styles composite,
// half of it projected onto the ten futures contracts the fund holds
// (TREND-ALLSTYLES-DBI-USD, built by cmd/gen-dbi-refdata, engine in
// pkg/simgen/dbireplica.go). Two things have to stay true of it and neither is
// visible in the shipped fund files, whose real quotes hide the donor from
// every statistic a user reads.
//
// The first is that it is still the composite: an average of the composite with
// a portfolio of the composite's own exposures cannot wander off, and if it
// does, a leg has gone stale or the regression has stopped fitting.
//
// The second is the reason it exists at all: it must track the fund BETTER than
// the raw composite does, over the fund's own live window, which is the only
// window where either can be judged. That is the measurement that put it in the
// chain (docs/trend-reconstruction-design.md), and a refresh that quietly
// reversed it would leave a worse donor in place with no other symptom.
const (
	dbiDonorID = "TREND-ALLSTYLES-DBI-USD"
	allStyles  = "TREND-ALLSTYLES-NET-USD"
	dbmfLive   = "2019-05-08" // the fund's first real quote, grafted into its simdata
	dbmfLiveTo = "2026-07-31"
)

func TestGoldenDBiDonorIsStillTheComposite(t *testing.T) {
	donor := mustRef(t, dbiDonorID)
	index := mustRef(t, allStyles)

	if n := len(donor.Points); n < 6000 {
		t.Fatalf("%s: %d days, expected a full daily calendar from 2000", dbiDonorID, n)
	}
	// The projection costs sixty trading days of warm-up and nothing else.
	start := donor.Points[0].Date
	if want := index.Points[0].Date; start.Before(want) || start.After(want.AddDate(0, 5, 0)) {
		t.Fatalf("%s starts %s, the composite starts %s", dbiDonorID, day(start), day(want))
	}
	if end, want := donor.Points[len(donor.Points)-1].Date, index.Points[len(index.Points)-1].Date; end.Before(want.AddDate(0, 0, -7)) {
		t.Fatalf("%s ends %s, the composite ends %s: the projection stopped early", dbiDonorID, day(end), day(want))
	}

	from, to := start, donor.Points[len(donor.Points)-1].Date
	a := between(t, donor, day(from), day(to))
	b := between(t, index, day(from), day(to))
	gap := 100 * (cagrOfPoints(a) - cagrOfPoints(b))
	ratio := annualized(logReturns(a), 252) / annualized(logReturns(b), 252)
	corr := monthlyCorrelation(a, b)
	dd := 100 * (maxDrawdown(a) - maxDrawdown(b))
	t.Logf("donor vs composite over %s..%s: %+.2f pts/yr, %.3f x volatility, %.3f monthly, drawdown %+.1f pts",
		day(from), day(to), gap, ratio, corr, dd)

	// An average of two versions of the same trade: close on the level, quieter
	// by the correlation between them, and moving with it month by month.
	if math.Abs(gap) > 1.0 {
		t.Errorf("donor CAGR is %+.2f pts/yr off the composite's, want within 1.0", gap)
	}
	if ratio < 0.75 || ratio > 1.00 {
		t.Errorf("donor volatility is %.3f times the composite's, want 0.75..1.00", ratio)
	}
	if corr < 0.90 {
		t.Errorf("donor tracks the composite at %.3f monthly, want at least 0.90", corr)
	}
	if math.Abs(dd) > 8 {
		t.Errorf("donor's worst drawdown is %+.1f points off the composite's", dd)
	}
}

// The donor earns its place only by tracking the fund better than the record it
// is made of. Measured on month-end returns, which is the honest frequency for
// a sleeve held for years, and on tracking error scaled to equal volatility, so
// neither candidate wins by simply carrying less risk.
func TestGoldenDBiDonorBeatsTheRawComposite(t *testing.T) {
	fund, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), "DBMF")
	if err != nil || !ok {
		t.Fatalf("simdata DBMF: ok=%v err=%v", ok, err)
	}
	real := between(t, fund, dbmfLive, dbmfLiveTo) // real quotes from inception
	donor := between(t, mustRef(t, dbiDonorID), dbmfLive, dbmfLiveTo)
	index := between(t, mustRef(t, allStyles), dbmfLive, dbmfLiveTo)
	if len(real) < 1500 {
		t.Fatalf("only %d days of live quotes", len(real))
	}

	cd, ci := monthlyCorrelation(donor, real), monthlyCorrelation(index, real)
	td, ti := scaledTE(donor, real), scaledTE(index, real)
	t.Logf("against the fund's own quotes: monthly correlation donor %.3f vs composite %.3f, scaled tracking error %.2f %% vs %.2f %%",
		cd, ci, 100*td, 100*ti)

	if cd <= ci {
		t.Errorf("the donor tracks the fund at %.3f monthly and the raw composite at %.3f: the projection has stopped paying", cd, ci)
	}
	if cd < 0.85 {
		t.Errorf("donor monthly correlation with the fund is %.3f, want at least 0.85", cd)
	}
	if td >= ti {
		t.Errorf("donor scaled tracking error %.2f %% is no better than the composite's %.2f %%", 100*td, 100*ti)
	}
}

func mustRef(t *testing.T, id string) *marketdata.Series {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil || !ok {
		t.Fatalf("refdata %s: ok=%v err=%v", id, ok, err)
	}
	return s
}

func day(t time.Time) string { return t.Format("2006-01-02") }

func cagrOfPoints(p []marketdata.Point) float64 {
	if len(p) < 2 {
		return 0
	}
	years := p[len(p)-1].Date.Sub(p[0].Date).Hours() / 24 / 365.25
	if years <= 0 || p[0].Close <= 0 {
		return 0
	}
	return math.Pow(p[len(p)-1].Close/p[0].Close, 1/years) - 1
}

func maxDrawdown(p []marketdata.Point) float64 {
	peak, worst := 0.0, 0.0
	for _, q := range p {
		peak = math.Max(peak, q.Close)
		if peak > 0 {
			worst = math.Min(worst, q.Close/peak-1)
		}
	}
	return worst
}

// monthEndReturns is the return of every calendar month, keyed by month.
func monthEndReturns(p []marketdata.Point) map[int]float64 {
	ends := map[int]marketdata.Point{}
	for _, q := range p {
		k := q.Date.Year()*12 + int(q.Date.Month())
		if cur, ok := ends[k]; !ok || q.Date.After(cur.Date) {
			ends[k] = q
		}
	}
	out := map[int]float64{}
	for k, q := range ends {
		if prev, ok := ends[k-1]; ok && prev.Close > 0 {
			out[k] = q.Close/prev.Close - 1
		}
	}
	return out
}

func pairMonths(a, b []marketdata.Point) (xa, xb []float64) {
	ma, mb := monthEndReturns(a), monthEndReturns(b)
	for k, v := range ma {
		if w, ok := mb[k]; ok {
			xa = append(xa, v)
			xb = append(xb, w)
		}
	}
	return xa, xb
}

func monthlyCorrelation(a, b []marketdata.Point) float64 {
	xa, xb := pairMonths(a, b)
	return correlation(xa, xb)
}

// scaledTE is the annualized tracking error of the candidate against the fund
// after the candidate is scaled to the fund's own monthly volatility. Without
// the scaling the quieter series always wins, which is the trap the design doc
// names: with correlation under one it pays to under-risk.
func scaledTE(cand, real []marketdata.Point) float64 {
	xc, xr := pairMonths(cand, real)
	if len(xc) < 12 {
		return math.NaN()
	}
	sc, sr := sampleStdev(xc), sampleStdev(xr)
	if sc <= 0 {
		return math.NaN()
	}
	k := sr / sc
	var v float64
	for i := range xc {
		d := xr[i] - k*xc[i]
		v += d * d
	}
	return math.Sqrt(v/float64(len(xc)-1)) * math.Sqrt(12)
}

func correlation(a, b []float64) float64 {
	if len(a) < 12 {
		return 0
	}
	var ma, mb float64
	for i := range a {
		ma += a[i]
		mb += b[i]
	}
	ma /= float64(len(a))
	mb /= float64(len(b))
	var cov, va, vb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		cov += da * db
		va += da * da
		vb += db * db
	}
	if va <= 0 || vb <= 0 {
		return 0
	}
	return cov / math.Sqrt(va*vb)
}

func sampleStdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	var m float64
	for _, x := range xs {
		m += x
	}
	m /= float64(len(xs))
	var v float64
	for _, x := range xs {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v / float64(len(xs)-1))
}
