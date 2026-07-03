package marketdata

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"
)

// vixSymbol is the CBOE Volatility Index. Like the yield symbols (^IRX...)
// it is an annualized percent LEVEL, not a price: it charts fine but never
// belongs in a return computation directly.
const vixSymbol = "^VIX"

// vixSnapshot is a bundled offline fallback for ^VIX: the full official
// daily close history (1990→), used when every live source fails and
// nothing is cached. The live series is always preferred; this only needs
// refreshing occasionally to keep the recent tail current. Regenerate from
// the CBOE endpoint fetchCBOEVIX reads, keeping "YYYY-MM-DD,close" rows
// under the comment header.
//
//go:embed data/vix.csv
var vixSnapshot string

// fetchCBOEVIX downloads the official VIX daily history from the CBOE CSV
// endpoint (1990→, refreshed daily), the fallback source when Yahoo cannot
// serve ^VIX.
func (c *Client) fetchCBOEVIX(ctx context.Context, from time.Time) (*Series, error) {
	body, err := c.get(ctx, c.CBOEBase+"/api/global/us_indices/daily_prices/VIX_History.csv")
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(body, []byte("DATE,")) {
		return nil, fmt.Errorf("cboe: unexpected response")
	}
	rows, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unreadable cboe CSV: %w", err)
	}
	var pts []Point
	for _, row := range rows[1:] {
		if len(row) < 5 {
			continue
		}
		t, terr := time.ParseInLocation("01/02/2006", row[0], time.UTC)
		if terr != nil || (!from.IsZero() && t.Before(from)) {
			continue
		}
		cl, cerr := strconv.ParseFloat(row[4], 64)
		if cerr != nil || cl <= 0 {
			continue
		}
		pts = append(pts, Point{Date: t, Close: cl})
	}
	if len(pts) < 2 {
		return nil, fmt.Errorf("cboe: only %d usable points", len(pts))
	}
	return vixSeries(pts), nil
}

// vixSeries packages daily VIX closes as a series. A volatility level
// carries no currency.
func vixSeries(pts []Point) *Series {
	return &Series{
		Symbol: vixSymbol,
		Name:   "CBOE Volatility Index (VIX)",
		Source: "cboe",
		Points: pts,
	}
}

// embeddedHistory returns the bundled last-resort snapshot of a symbol, if
// any: it answers only when every live source failed and nothing is cached,
// so a chart still renders offline at the cost of missing the latest days.
// ^VIX serves its full daily snapshot; the euro crosses serve the long
// long ECU/DM/EUR proxy (daily, 1971→) that normally only extends a live cross,
// so USD↔EUR conversion keeps working offline, at monthly granularity.
func embeddedHistory(symbol string) (*Series, bool) {
	switch symbol {
	case vixSymbol:
		return vixSeries(parseAnchors(vixSnapshot, "2006-01-02")), true
	}
	if proxy, ok := eurusdLongCross(symbol); ok {
		base, quote, _ := fxCross(symbol)
		return &Series{
			Symbol:   symbol,
			Name:     base + "/" + quote + " (bundled monthly proxy)",
			Currency: quote,
			Source:   "fred",
			Points:   proxy,
		}, true
	}
	return nil, false
}
