// Fetching plumbing shared by the CLI-only modes: per-asset fetch in a base
// currency and the overlapping-window helper.
package main

import (
	"context"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// commonWindow returns the latest first date and earliest last date across
// the series (their overlapping period).
func commonWindow(list []*marketdata.Series) (start, end time.Time) {
	start, end = list[0].First().Date, list[0].Last().Date
	for _, s := range list[1:] {
		if f := s.First().Date; f.After(start) {
			start = f
		}
		if l := s.Last().Date; l.Before(end) {
			end = l
		}
	}
	return start, end
}

// fetchAsset downloads the history of an identifier (ticker or ISIN). A
// bare identifier sticks to the asset's real quotes, from its actual
// inception. A "SIM"-suffixed identifier (DBMFSIM, VOOSIM…) additionally
// extends the uncovered period backwards: first with the permanent simulated
// series (embedded datasets, or -simdata), then a known proxy; real
// quotes always win wherever they exist.
// fetchAsset runs the full library pipeline (SIM extension, currency
// conversion, window) for one asset, in the CLI's base currency.
func fetchAsset(ctx context.Context, c *marketdata.Client, id string, opt *options) (*marketdata.Series, error) {
	return fetchAssetIn(ctx, c, id, opt, opt.currency)
}

// fetchAssetIn is fetchAsset with an explicit target currency, used when a
// portfolio is evaluated in several currencies ("#meta currencies").
func fetchAssetIn(ctx context.Context, c *marketdata.Client, id string, opt *options, currency string) (*marketdata.Series, error) {
	return c.FetchExtended(ctx, id, marketdata.FetchOptions{
		From:     opt.start,
		To:       opt.end,
		NoSim:    opt.noSim,
		Simdata:  opt.simdata,
		Currency: currency,
	})
}
