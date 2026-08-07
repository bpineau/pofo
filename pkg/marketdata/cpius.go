package marketdata

import (
	"context"
	_ "embed"
	"time"
)

// cpiUSSymbol is the synthetic identifier of the United States consumer price
// index, the dollar sibling of the "^HICP-<geo>" family: an index LEVEL
// (1982-84=100), carrying no currency, that charts fine and deflates nominal
// series but never belongs in a return computation directly.
const cpiUSSymbol = "^CPI-US"

// cpiUSSnapshot is a bundled offline fallback for ^CPI-US: the monthly CPI-U
// all-items NSA index (BLS, 1913→), used when the live FRED endpoint is
// unreachable and nothing is cached. The live series is always preferred;
// this only needs refreshing occasionally to keep the recent tail current.
// Regenerate from FRED fredgraph.csv?id=CPIAUCNS (or the datahub mirror,
// raw.githubusercontent.com/datasets/cpi-us/main/data/cpiai.csv, the same
// BLS series), rewriting rows as "YYYY-MM,value" under the comment header.
//
//go:embed data/cpi-us.csv
var cpiUSSnapshot string

// fetchCPIUS returns the daily-interpolated US CPI index, served offline-first
// from the bundled snapshot (CPIAUCNS, monthly since 1913): a normal run never
// downloads it. The live FRED series is fetched only under RefreshInflation
// (set during warmup), which then updates the disk cache a later run prefers.
func (c *Client) fetchCPIUS(ctx context.Context, from time.Time) (*Series, error) {
	return c.embeddedHistory(ctx, "fred", cpiUSSymbol, from,
		func() (*Series, bool) {
			return cpiUSSeries(parseMonthlyAnchors(cpiUSSnapshot)), true
		},
		func() (*Series, error) {
			monthly, err := c.fetchFRED(ctx, "CPIAUCNS")
			if err != nil {
				return nil, err
			}
			return cpiUSSeries(monthly), nil
		})
}

// cpiUSSeries packages monthly CPI anchors as a daily-interpolated index
// series, the same shape fetchHICP serves.
func cpiUSSeries(monthly []Point) *Series {
	return &Series{
		Symbol: cpiUSSymbol,
		Name:   "US CPI (all items, 1982-84=100)",
		Source: "fred",
		Points: monthlyToDaily(monthly),
	}
}
