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

// fetchECBFX downloads the daily history of a currency cross ("USDEUR=X")
// from the European Central Bank reference rates, the FX fallback when both
// Yahoo and Stooq fail. The ECB publishes one rate per business day around
// 16:00 CET, EUR against some forty currencies, daily since 1999-01-04, as a
// zipped CSV explicitly meant for programmatic use. A cross off the euro is
// derived through it: rate(AAA/BBB) = rate(EUR/BBB) / rate(EUR/AAA).
//
// Reference rates are a daily fixing, not market closes: fine as a fallback,
// which is the only place this source is consulted.
func (c *Client) fetchECBFX(ctx context.Context, symbol string, from time.Time) (*Series, error) {
	base, quote, ok := fxCross(symbol)
	if !ok || base == quote {
		return nil, fmt.Errorf("%s is not a currency cross", symbol)
	}
	body, err := c.get(ctx, c.ECBBase+"/stats/eurofxref/eurofxref-hist.zip")
	if err != nil {
		return nil, err
	}
	rows, err := ecbRows(body)
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
// per-source error when every source failed.
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
		return nil, fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v; ecb: %v)", symbol, yahooErr, stooqErr, ecbErr)
	}
	if symbol == vixSymbol {
		s, cboeErr := c.fetchCBOEVIX(ctx, from)
		if cboeErr == nil {
			c.Logf("%s fetched via CBOE", symbol)
			return s, nil
		}
		return nil, fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v; cboe: %v)", symbol, yahooErr, stooqErr, cboeErr)
	}
	return nil, fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v)", symbol, yahooErr, stooqErr)
}
