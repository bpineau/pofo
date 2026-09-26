package analyze_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// fakeSource is an offline analyze.Source over synthetic series, keyed by
// bare identifier. It honors the parts of marketdata.Client's contract a
// study relies on: the SIM suffix (and NoSim), the From/To window, the
// currency tag, and ErrUnknownIdentifier for an identifier it does not serve.
// "Converting" only relabels the currency: the levels stay as generated.
type fakeSource struct {
	series   map[string]*marketdata.Series // real quotes
	backcast map[string]*marketdata.Series // what a SIM request returns, when present
	fees     map[string]float64            // TERs in percent per year
	failSim  map[string]bool               // a SIM request fails (an unreadable backcast)
	outage   map[string]bool               // every request fails without blaming the identifier
	fetched  []string                      // every identifier asked for, in order
}

func (f *fakeSource) FetchExtended(_ context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error) {
	f.fetched = append(f.fetched, id)
	base, sim := marketdata.SplitSim(id)
	if opt.NoSim {
		sim = false
	}
	if f.outage[base] {
		return nil, fmt.Errorf("fake: %s: source did not answer", base)
	}
	s, ok := f.series[base]
	if !ok {
		return nil, &marketdata.UnknownIdentifierError{ID: base, Failures: errors.New("fake: " + base + " is not served")}
	}
	if sim && f.failSim[base] {
		return nil, fmt.Errorf("fake: %s: backcast unreadable", base)
	}
	if b, ok := f.backcast[base]; sim && ok {
		s = b
	}
	out := marketdata.Trim(s, opt.From, opt.To)
	if opt.Currency != "" && out.Currency != "" && out.Currency != opt.Currency {
		cp := *out
		cp.Currency = opt.Currency
		out = &cp
	}
	return out, nil
}

func (f *fakeSource) Fees(_ context.Context, id string) (float64, bool) {
	ter, ok := f.fees[id]
	return ter, ok
}

// rng is a xorshift64* generator: integer arithmetic only, so the synthetic
// series, and the examples' printed output, are the same on every platform.
type rng struct{ s uint64 }

func (r *rng) next() uint64 {
	r.s ^= r.s >> 12
	r.s ^= r.s << 25
	r.s ^= r.s >> 27
	return r.s * 2685821657736338717
}

// normal is an approximately standard normal draw (Irwin-Hall, twelve
// uniforms), plenty for a synthetic market.
func (r *rng) normal() float64 {
	sum := 0.0
	for range 12 {
		sum += float64(r.next()>>11) / (1 << 53)
	}
	return sum - 6
}

var (
	calStart = time.Date(2010, 1, 4, 0, 0, 0, 0, time.UTC)
	calEnd   = time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)
)

// calendar is every weekday of the synthetic market.
func calendar() []time.Time {
	var out []time.Time
	for d := calStart; !d.After(calEnd); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			out = append(out, d)
		}
	}
	return out
}

// market is the daily common factor every synthetic asset loads on.
func market(n int) []float64 {
	r := &rng{s: 0x2545F4914F6CDD1D}
	out := make([]float64, n)
	for i := range out {
		out[i] = r.normal()
	}
	return out
}

// asset describes one synthetic series: yearly drift and volatility as
// fractions, beta its loading on the market factor (the rest idiosyncratic).
type asset struct {
	symbol, name, source, currency string
	from                           time.Time
	drift, vol, beta               float64
	seed                           uint64
}

// synth generates the asset's daily closes from its first date to calEnd.
func synth(a asset) *marketdata.Series {
	dates := calendar()
	z := market(len(dates))
	r := &rng{s: a.seed}
	s := &marketdata.Series{Symbol: a.symbol, Name: a.name, Source: a.source, Currency: a.currency}
	level := 100.0
	idio := math.Sqrt(1 - a.beta*a.beta)
	for i, d := range dates {
		e := r.normal()
		if d.Before(a.from) {
			continue
		}
		if len(s.Points) > 0 {
			level *= 1 + a.drift/252 + a.vol/math.Sqrt(252)*(a.beta*z[i]+idio*e)
		}
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: level})
	}
	return s
}

// igRealStart is when the late holding's real quotes begin.
var igRealStart = time.Date(2013, 1, 2, 0, 0, 0, 0, time.UTC)

// newFake serves the test universe: a world-equity line and an aggregate
// bond line from 2010, a distributing world-equity line quoted as a NAV with
// its dividends, a gold line quoting from 2013 whose SIM request returns a
// backcast from 2010, and a benchmark index. Real public identifiers, fake
// numbers.
func newFake() *fakeSource {
	world := asset{"IWDA", "iShares Core MSCI World UCITS ETF USD (Acc)", "yahoo", "USD", calStart, 0.08, 0.15, 0.9, 11}
	bonds := asset{"AGGH", "iShares Core Global Aggregate Bond UCITS ETF EUR Hedged (Acc)", "yahoo", "EUR", calStart, 0.02, 0.04, -0.2, 12}
	dist := asset{"VWRL", "Vanguard FTSE All-World UCITS ETF (USD) Distributing", "ft", "USD", calStart, 0.075, 0.16, 0.85, 13}
	gold := asset{"IGLN", "iShares Physical Gold ETC", "yahoo", "USD", calStart, 0.04, 0.15, 0.05, 14}
	index := asset{"MSCIWORLD", "MSCI World Net Total Return (index, fee-free)", "index", "USD", calStart, 0.082, 0.15, 1, 15}

	vwrl := synth(dist)
	for _, p := range vwrl.Points {
		if p.Date.Weekday() == time.Wednesday && p.Date.Day() <= 7 && p.Date.Month()%3 == 0 {
			vwrl.Dividends = append(vwrl.Dividends, marketdata.Dividend{Date: p.Date, Amount: 0.4})
		}
	}
	igSim := synth(gold)
	igSim.SimulatedBefore, igSim.ProxySymbol = igRealStart, "simdata"
	igReal := marketdata.Trim(igSim, igRealStart, time.Time{})
	igReal = &marketdata.Series{Symbol: igReal.Symbol, Name: igReal.Name, Source: igReal.Source, Currency: igReal.Currency, Points: igReal.Points}

	return &fakeSource{
		series: map[string]*marketdata.Series{
			"IWDA":      synth(world),
			"AGGH":      synth(bonds),
			"VWRL":      vwrl,
			"IGLN":      igReal,
			"MSCIWORLD": synth(index),
		},
		backcast: map[string]*marketdata.Series{"IGLN": igSim},
		fees:     map[string]float64{"IWDA": 0.2, "VWRL": 0.19, "IGLN": 0.12},
		failSim:  map[string]bool{},
		outage:   map[string]bool{},
	}
}
