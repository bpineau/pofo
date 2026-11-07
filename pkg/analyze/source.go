package analyze

import (
	"context"
	"fmt"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// SeriesSource is an in-memory Source keyed by the identifiers callers use.
// It performs no network access or currency conversion. FetchExtended returns
// a windowed copy; a requested currency differing from a known currency is an
// error, not a relabel. Populate Currency before requesting converted studies.
// Fee lookup is unknown; specify fees on portfolio.Line when needed. Do not
// mutate the map or its series while a study runs.
type SeriesSource map[string]*marketdata.Series

// FetchExtended reads an exact key and honors the window. SIM histories are
// supplied under their explicit SIM keys. NoSim removes reconstructed points.
func (src SeriesSource) FetchExtended(ctx context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s, ok := src[id]
	if !ok || s == nil {
		return nil, fmt.Errorf("%s: %w", id, marketdata.ErrUnknownIdentifier)
	}
	if opt.Currency != "" && s.Currency != "" && opt.Currency != s.Currency {
		return nil, fmt.Errorf("analyze: %s is in %s, convert to %s before using SeriesSource", id, s.Currency, opt.Currency)
	}
	start := opt.From
	if opt.NoSim && s.SimulatedBefore.After(start) {
		start = s.SimulatedBefore
	}
	out, err := marketdata.Trim(s, start, opt.To).Resample(marketdata.Daily)
	return out, err
}

// Fees returns unknown; SeriesSource holds prices, not a fee catalog.
func (src SeriesSource) Fees(context.Context, string) (float64, bool) { return 0, false }

// Series studies caller-owned data without a client or identifier lookup.
// benchmark may be nil; when present it overrides Options.Benchmark. If it is
// absent Options.Benchmark must be empty. Window, currency checks, inflation
// and risk-free options have exactly the same meaning as in Asset.
func Series(s, benchmark *marketdata.Series, opt Options) (*AssetStudy, error) {
	if s == nil {
		return nil, fmt.Errorf("analyze: nil series")
	}
	if benchmark == nil && opt.Benchmark != "" {
		return nil, fmt.Errorf("analyze: supply the benchmark series")
	}
	id := s.Symbol
	if id == "" {
		id = "series"
	}
	src := SeriesSource{id: s}
	if benchmark != nil {
		opt.Benchmark = id + "/benchmark"
		src[opt.Benchmark] = benchmark
	}
	return Asset(context.Background(), src, id, opt)
}

// fetch uses the explorer's loading contract when the source supports it,
// preserving compatibility with consumers implementing only Source.
func fetch(ctx context.Context, src Source, id string, opt marketdata.FetchOptions) (*marketdata.Series, error) {
	if src == nil {
		return nil, fmt.Errorf("analyze: nil source")
	}
	if loader, ok := src.(interface {
		Load(context.Context, string, marketdata.FetchOptions) (*marketdata.Series, error)
	}); ok {
		return loader.Load(ctx, id, opt)
	}
	return src.FetchExtended(ctx, id, opt)
}
