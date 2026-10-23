package marketdata

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// fetchECBFX returns the daily history of a currency cross ("USDEUR=X")
// from the European Central Bank reference rates: the FX fallback when both
// Yahoo and Stooq fail, and the source of the sessions those two skipped
// (fillFXHoles). The ECB publishes one rate per business day around
// 16:00 CET, EUR against some forty currencies, daily since 1999-01-04, as a
// zipped CSV explicitly meant for programmatic use. A cross off the euro is
// derived through it: rate(AAA/BBB) = rate(EUR/BBB) / rate(EUR/AAA).
//
// Reference rates are a daily fixing, not market closes: fine as a fallback
// and as the patch of a missing day, the only two places this source is
// consulted.
func (c *Client) fetchECBFX(ctx context.Context, symbol string, from time.Time) (*Series, error) {
	base, quote, ok := fxCross(symbol)
	if !ok || base == quote {
		return nil, fmt.Errorf("%s is not a currency cross", symbol)
	}
	rows, err := c.ecbTable(ctx)
	if err != nil {
		return nil, err
	}
	baseRate := ecbColumn(rows[0], base)
	quoteRate := ecbColumn(rows[0], quote)
	if baseRate == nil || quoteRate == nil {
		return nil, fmt.Errorf("ecb does not publish %s/%s", base, quote)
	}
	s := &Series{
		Symbol:   symbol,
		Name:     base + "/" + quote + " ECB reference rate",
		Currency: quote,
		Source:   "ecb",
	}
	for _, row := range rows[1:] {
		t, err := time.ParseInLocation("2006-01-02", row[0], time.UTC)
		if err != nil || (!from.IsZero() && t.Before(from)) {
			continue
		}
		rb, okb := baseRate(row)
		rq, okq := quoteRate(row)
		if !okb || !okq {
			continue // N/A: the currency was not quoted that day
		}
		s.Points = append(s.Points, Point{Date: t, Close: rq / rb})
	}
	// The ECB file lists the newest day first; a Series is ascending.
	sort.Slice(s.Points, func(i, j int) bool { return s.Points[i].Date.Before(s.Points[j].Date) })
	return s, nil
}

// ecbTableTTL bounds how long a process reuses the ECB history it downloaded.
// The file is one 1 MB archive serving every cross, and a refresh of several
// crosses (every converted currency of a portfolio) must not fetch it once
// per cross; the ECB adds one row a day, so a few hours of reuse lose at most
// the fixing that was published meanwhile.
const ecbTableTTL = 6 * time.Hour

// ecbTable returns the rows of the eurofxref-hist file, header first, newest
// day first below it, downloading them at most once per ecbTableTTL.
func (c *Client) ecbTable(ctx context.Context) ([][]string, error) {
	c.mu.Lock()
	rows, at := c.ecbRows, c.ecbRowsAt
	c.mu.Unlock()
	if rows != nil && time.Since(at) < ecbTableTTL {
		return rows, nil
	}
	body, err := c.get(ctx, c.ECBBase+"/stats/eurofxref/eurofxref-hist.zip")
	if err != nil {
		return nil, err
	}
	rows, err = ecbRows(body)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.ecbRows, c.ecbRowsAt = rows, time.Now()
	c.mu.Unlock()
	return rows, nil
}

// fillFXHoles adds to a currency cross fetched from a market source (Yahoo,
// Stooq) the business days it did not quote, taking each from the ECB
// reference rates: every ECB day after the series' first point that the
// series lacks, the days after its last point included. It changes nothing
// before the first point (extendFXBack owns the deep past) and nothing when
// the ECB cannot be read, which it logs as a warning.
//
// Yahoo serves a currency cross with a null bar now and then, and such a
// hole is not harmless: ConvertCurrency forward-fills the previous day's
// rate onto it, so every asset quoted in that currency is read 0.3 % off on
// that day and its daily return the day after is wrong by as much. The ECB
// fixing of 14:15 CET is not that day's close, but it is that day's rate,
// where the forward-fill is the day before's. The same patch catches a print
// the FX de-spiker dropped (cleanQuotes runs first) and a source whose last
// days are missing.
func (c *Client) fillFXHoles(ctx context.Context, symbol string, s *Series) {
	if s == nil || len(s.Points) == 0 {
		return
	}
	ecb, err := c.fetchECBFX(ctx, symbol, s.First().Date)
	if err != nil {
		c.Logf("warning: %s: the ECB reference rates could not be read to check for missing sessions: %v", symbol, err)
		return
	}
	have := make(map[time.Time]bool, len(s.Points))
	for _, p := range s.Points {
		have[p.Date] = true
	}
	var added []string
	for _, p := range ecb.Points {
		if have[p.Date] || !p.Date.After(s.First().Date) {
			continue
		}
		s.Points = append(s.Points, p)
		added = append(added, p.Date.Format("2006-01-02"))
	}
	if len(added) == 0 {
		return
	}
	sort.Slice(s.Points, func(i, j int) bool { return s.Points[i].Date.Before(s.Points[j].Date) })
	shown := added
	if len(shown) > 3 {
		shown = append(shown[:3:3], "…")
	}
	c.Logf("%s: %d session(s) the source did not quote filled from the ECB reference rates (%s)",
		symbol, len(added), strings.Join(shown, ", "))
}

// ecbRows extracts the CSV records from the zipped eurofxref-hist archive.
func ecbRows(zipBody []byte) ([][]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipBody), int64(len(zipBody)))
	if err != nil {
		return nil, fmt.Errorf("unreadable ecb archive: %w", err)
	}
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".csv") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		r := csv.NewReader(rc)
		r.FieldsPerRecord = -1 // every row carries a trailing comma
		rows, err := r.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("unreadable ecb CSV: %w", err)
		}
		if len(rows) < 2 {
			return nil, fmt.Errorf("empty ecb CSV")
		}
		return rows, nil
	}
	return nil, fmt.Errorf("no CSV in the ecb archive")
}

// ecbColumn returns a reader of the given currency's EUR rate in a data row,
// or nil when the file does not carry that currency. The euro itself is the
// implicit 1. ok is false on an N/A hole (a currency not quoted that day).
func ecbColumn(header []string, currency string) func(row []string) (rate float64, ok bool) {
	if currency == "EUR" {
		return func([]string) (float64, bool) { return 1, true }
	}
	for i, name := range header {
		if strings.TrimSpace(name) != currency {
			continue
		}
		return func(row []string) (float64, bool) {
			if i >= len(row) {
				return 0, false
			}
			rate, err := strconv.ParseFloat(strings.TrimSpace(row[i]), 64)
			return rate, err == nil && rate > 0
		}
	}
	return nil
}

// historyFallback consults the non-Yahoo daily sources after a Yahoo
// failure: Stooq for everything it maps, then the ECB reference rates for a
// currency cross and the CBOE endpoint for ^VIX. It returns the combined
// per-source error when every source failed, marked as an absence only when
// every one of them reported one (see markAbsentIfAll).
func (c *Client) historyFallback(ctx context.Context, symbol string, from time.Time, yahooErr error) (*Series, error) {
	s, stooqErr := c.fetchStooq(ctx, symbol, from)
	if stooqErr == nil {
		c.Logf("%s fetched via stooq (prices not dividend-adjusted)", symbol)
		return s, nil
	}
	if _, _, ok := fxCross(symbol); ok {
		s, ecbErr := c.fetchECBFX(ctx, symbol, from)
		if ecbErr == nil {
			c.Logf("%s fetched via the ECB reference rates", symbol)
			return s, nil
		}
		return nil, markAbsentIfAll(
			fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v; ecb: %v)", symbol, yahooErr, stooqErr, ecbErr),
			yahooErr, stooqErr, ecbErr)
	}
	if symbol == vixSymbol {
		s, cboeErr := c.fetchCBOEVIX(ctx, from)
		if cboeErr == nil {
			c.Logf("%s fetched via CBOE", symbol)
			return s, nil
		}
		return nil, markAbsentIfAll(
			fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v; cboe: %v)", symbol, yahooErr, stooqErr, cboeErr),
			yahooErr, stooqErr, cboeErr)
	}
	return nil, markAbsentIfAll(
		fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v)", symbol, yahooErr, stooqErr),
		yahooErr, stooqErr)
}
