package replay

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/metrics"
)

// Setup is the household whose retirement is replayed: one capital, one real
// spending target, one plan horizon and one start year.
//
// Years is the household's PLAN, not the length of the record. It is what the
// horizon-aware rules quote a payment against (the amortisation rule spreads
// wealth over the forty years it was told about, not over the years the data
// happens to have), so it stays whole even where the replay stops early. A
// withdrawal path is causal, so running the full plan and reading only the
// covered prefix is identical to running the prefix, except the rules now face
// the horizon the household actually has.
//
// Mu, Sigma, Df and TargetRuin are the forward assumptions the two horizon-aware
// rules need; they should be generic, not fitted to the era being replayed,
// since no retiree knew what was coming. RaiseCap is the ceiling on the
// risk-based guardrail's raises as a multiple of Spend (0 leaves the protective
// default of 1, which lets the rule cut but never share a good decade).
type Setup struct {
	Start      int
	Capital    float64
	Spend      float64
	Years      int
	Mu         float64
	Sigma      float64
	Df         float64
	TargetRuin float64
	RaiseCap   float64
}

// Rule is one withdrawal rule's replayed retirement: not a score, a portrait.
// Spend and Wealth carry one point per covered year (Wealth at year end), Mean
// and CV describe the standard of living delivered, Low and LeanYears the depth
// and the duration of the lean spells, Final what was still on the table.
type Rule struct {
	Name, NameFR, Tag, Help, Color string
	Spend                          []float64
	Wealth                         []float64
	Mean                           float64
	CV                             float64 // coefficient of variation of the yearly income
	Low, High                      float64
	Total                          float64
	Final                          float64
	LeanYears                      int
	Ruined                         bool
	RuinYear                       int // calendar year the money ran out (only when Ruined)
}

// Result is one replayed retirement: the untouched portfolio it happened in,
// and every rule's life inside it. Everything is REAL (inflation removed) and
// gross of tax.
type Result struct {
	Setup   Setup
	Start   int
	End     int
	Years   int  // calendar years the record actually covers
	Partial bool // the record ran out before the plan did

	Dates  []time.Time // month ends of the untouched index, from the opening December
	Index  []float64   // the untouched real portfolio, base 100
	Market []float64   // calendar-year real returns, one per covered year

	CAGR        float64 // annualised real return of the untouched portfolio
	Decade      float64 // annualised real return of its first ten years
	Volatility  float64 // standard deviation of the calendar-year real returns
	WorstYear   float64
	WorstAt     int
	MaxDrawdown float64 // deepest month-end peak-to-trough real loss

	Rules []Rule
}

// Run replays every withdrawal rule of Policies through the real returns of the
// bundled reference 60/40 under one Setup. Each rule gets the same capital, the
// same target, the same horizon and the same sequence of years; only the
// spending rule differs, which is the whole point.
func Run(s Setup) (Result, error) {
	if s.Capital <= 0 || s.Spend <= 0 || s.Years <= 0 {
		return Result{}, fmt.Errorf("replay: need a positive capital, spending target and horizon")
	}
	ref, err := Reference()
	if err != nil {
		return Result{}, err
	}
	dates, index, seq, ok := ref.Window(s.Start, s.Years)
	if !ok {
		return Result{}, fmt.Errorf("replay: the bundled record does not cover a retirement starting in %d", s.Start)
	}
	years := len(seq)

	raiseCap := s.RaiseCap
	if raiseCap <= 0 {
		raiseCap = 1
	}
	cfg := PolicyConfig{
		WR:          s.Spend / s.Capital,
		SafeWR:      SafeRateTable(s.Years, s.Mu, s.Sigma, s.Df, s.TargetRuin),
		RaiseCap:    raiseCap * s.Spend,
		PVRate:      math.Min(0.06, math.Max(0.005, GeometricReturn(s.Mu, s.Sigma))),
		AmortReturn: GeometricReturn(s.Mu, s.Sigma),
	}
	base := decumul.Plan{Capital: s.Capital, NeedAnnual: s.Spend, Years: s.Years}

	out := Result{
		Setup: s, Start: s.Start, End: s.Start + years - 1, Years: years,
		Partial: years < s.Years,
		Dates:   dates, Index: index, Market: seq,
	}
	for _, pol := range Policies(cfg) {
		p := base
		pol.Apply(&p)
		res := p.RunPath(seq)

		r := Rule{Name: pol.Name, NameFR: pol.NameFR, Tag: pol.Tag, Help: pol.Help, Color: pol.Color,
			Spend: res.Spend[:years], Wealth: res.Wealth[1 : years+1],
			Low: math.Inf(1), Final: res.Wealth[years]}
		for _, v := range r.Spend {
			r.Total += v
			r.Low, r.High = math.Min(r.Low, v), math.Max(r.High, v)
			if v < s.Spend-1 {
				r.LeanYears++
			}
		}
		r.Mean = r.Total / float64(years)
		r.CV = coefVar(r.Spend, r.Mean)
		// Ruin only counts inside the window the reader is shown.
		r.Ruined = res.Ruined && res.RuinYear < years
		r.RuinYear = s.Start + res.RuinYear
		out.Rules = append(out.Rules, r)
	}
	out.fillMarketStats()
	return out, nil
}

// fillMarketStats describes the weather the retirement happened in: what the
// untouched portfolio returned, how much it moved, how deep it went, and which
// single year hurt most. All real.
func (r *Result) fillMarketStats() {
	n := float64(r.Years)
	r.CAGR = math.Pow(r.Index[len(r.Index)-1]/r.Index[0], 1/n) - 1
	// The first decade is the one that decides a retirement: it is where a bad
	// sequence does its damage, on a portfolio still carrying its full weight.
	d := min(120, len(r.Index)-1)
	r.Decade = math.Pow(r.Index[d]/r.Index[0], 12/float64(d)) - 1

	mean := metrics.Mean(r.Market)
	var v float64
	for _, x := range r.Market {
		v += (x - mean) * (x - mean)
	}
	if r.Years > 1 {
		r.Volatility = math.Sqrt(v / (n - 1))
	}
	r.WorstYear, r.WorstAt = math.Inf(1), r.Start
	for i, x := range r.Market {
		if x < r.WorstYear {
			r.WorstYear, r.WorstAt = x, r.Start+i
		}
	}
	r.MaxDrawdown = metrics.MaxDrawdown(r.Dates, r.Index).Depth
}

// coefVar is the coefficient of variation of a yearly income: its standard
// deviation over its mean. It is the one number for "how much does this rule
// move the household's life about": 0 is an income that never changed.
func coefVar(xs []float64, mean float64) float64 {
	if mean <= 0 || len(xs) < 2 {
		return 0
	}
	var v float64
	for _, x := range xs {
		d := x - mean
		v += d * d
	}
	return math.Sqrt(v/float64(len(xs)-1)) / mean
}
