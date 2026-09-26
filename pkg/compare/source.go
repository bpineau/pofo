package compare

import (
	"context"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// source is the analyze.Source every column of one Compute call is studied
// through. It applies the comparison's own fetch policy on top of the client
// (Options.NoSim, Options.NoFees, and Options.ExactForeign per identifier,
// which analyze's all-or-nothing ExactOnly cannot express) and memoizes each
// request, so an asset held by several columns, a benchmark read in every
// currency and a financing rate shared by the levered specs are fetched once
// and are the very same series in every column.
//
// holding and benchmark state the requests analyze makes (its Options turned
// into FetchOptions), so the fetches Compute makes up front, to name a
// resolution or a failure, are the ones the studies then find here.
type source struct {
	client *marketdata.Client
	opt    Options
	memo   map[request]fetched
}

// request is a fetch as the memo tells it apart: the identifier and every
// FetchOptions field that changes the series returned. Simdata is left out:
// it is one filesystem for the whole Compute call.
type request struct {
	id                           string
	from, to                     time.Time
	currency                     string
	noSim, exact, raw, noConvert bool
}

// fetched is a memoized answer, failures included: a request that failed
// once fails the same way for the next column instead of being retried.
type fetched struct {
	s   *marketdata.Series
	err error
}

var _ analyze.Source = (*source)(nil)

func newSource(client *marketdata.Client, opt Options) *source {
	return &source{client: client, opt: opt, memo: map[request]fetched{}}
}

// FetchExtended is marketdata.Client.FetchExtended under the comparison's
// policy: -no-simulate wins over any SIM request, and an identifier outside
// the bundled catalog resolves exactly when Options.ExactForeign asks for it
// (a catalog identifier is pinned already).
func (s *source) FetchExtended(ctx context.Context, id string, fo marketdata.FetchOptions) (*marketdata.Series, error) {
	fo.NoSim = fo.NoSim || s.opt.NoSim
	fo.ExactOnly = fo.ExactOnly || (s.opt.ExactForeign && !marketdata.KnownLocal(id))
	key := request{
		id: id, from: fo.From, to: fo.To, currency: fo.Currency,
		noSim: fo.NoSim, exact: fo.ExactOnly, raw: fo.Raw, noConvert: fo.NoConvert,
	}
	if f, ok := s.memo[key]; ok {
		return f.s, f.err
	}
	series, err := s.client.FetchExtended(ctx, id, fo)
	s.memo[key] = fetched{series, err}
	return series, err
}

// Fees is the client's TER lookup, or nothing under Options.NoFees, which
// leaves every TER a file does not declare unknown.
func (s *source) Fees(ctx context.Context, id string) (float64, bool) {
	if s.opt.NoFees {
		return -1, false
	}
	return s.client.Fees(ctx, id)
}

// holding is the request analyze makes for a holding of a column evaluated
// in currency cur: the comparison window, the simulated histories, the
// currency. Compute prefetches with it so the studies find every holding in
// the memo.
func (s *source) holding(cur string) marketdata.FetchOptions {
	return marketdata.FetchOptions{From: s.opt.Start, To: s.opt.End, Simdata: s.opt.Simdata, Currency: cur}
}

// benchmark is the request analyze makes for the benchmark in currency cur:
// the holding request, real quotes only.
func (s *source) benchmark(cur string) marketdata.FetchOptions {
	fo := s.holding(cur)
	fo.NoSim = true
	return fo
}
