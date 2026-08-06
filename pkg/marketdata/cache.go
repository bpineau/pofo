package marketdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// cacheFile is the JSON document stored in the cache directory, one per
// series view (a raw view lives under its own "SYMBOL~raw" identity).
// Dividend dates and amounts are parallel arrays; files written before the
// dividend columns existed simply load with none.
type cacheFile struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Currency      string    `json:"currency"`
	Source        string    `json:"source"`
	RequestedFrom string    `json:"requested_from"`
	FetchedAt     time.Time `json:"fetched_at"`
	Dates         []string  `json:"dates"`
	Closes        []float64 `json:"closes"`
	DivDates      []string  `json:"div_dates,omitempty"`
	DivAmounts    []float64 `json:"div_amounts,omitempty"`
}

func (c *Client) cachePath(symbol string) string {
	return filepath.Join(c.CacheDir, sanitizeFilename(symbol)+".json")
}

// loadCache returns the cached series for symbol if it is fresh enough and
// was downloaded with a start date covering the requested one.
func (c *Client) loadCache(symbol string, from time.Time) (*Series, bool) {
	s, fetchedAt, ok := c.loadCacheAnyAge(symbol, from)
	if !ok || time.Since(fetchedAt) > c.MaxAge {
		return nil, false
	}
	return s, true
}

// loadCacheAnyAge returns the cached series for symbol regardless of its
// age, along with its download time. It backs the stale-cache fallback: a
// failed refresh must never lose previously downloaded data.
func (c *Client) loadCacheAnyAge(symbol string, from time.Time) (*Series, time.Time, bool) {
	if c.CacheDir == "" {
		return nil, time.Time{}, false
	}
	data, err := os.ReadFile(c.cachePath(symbol))
	if err != nil {
		return nil, time.Time{}, false
	}
	var cf cacheFile
	if err := json.Unmarshal(data, &cf); err != nil || len(cf.Dates) == 0 || len(cf.Dates) != len(cf.Closes) {
		return nil, time.Time{}, false
	}
	reqFrom, err := time.ParseInLocation("2006-01-02", cf.RequestedFrom, time.UTC)
	if err != nil || reqFrom.After(from) {
		return nil, time.Time{}, false
	}
	s := &Series{Symbol: cf.Symbol, Name: cf.Name, Currency: cf.Currency, Source: cf.Source}
	for i, d := range cf.Dates {
		t, err := time.ParseInLocation("2006-01-02", d, time.UTC)
		if err != nil {
			return nil, time.Time{}, false
		}
		if t.Before(from) {
			continue
		}
		s.Points = append(s.Points, Point{Date: t, Close: cf.Closes[i]})
	}
	if len(s.Points) == 0 {
		return nil, time.Time{}, false
	}
	if len(cf.DivDates) == len(cf.DivAmounts) {
		for i, d := range cf.DivDates {
			t, err := time.ParseInLocation("2006-01-02", d, time.UTC)
			if err != nil || t.Before(from) {
				continue
			}
			s.Dividends = append(s.Dividends, Dividend{Date: t, Amount: cf.DivAmounts[i]})
		}
	}
	return s, cf.FetchedAt, true
}

// saveCache persists a downloaded series under its own symbol; failures are
// logged, never fatal.
func (c *Client) saveCache(s *Series, from time.Time) {
	c.saveCacheAs(s.Symbol, s, from)
}

// saveCacheAs persists a downloaded series under an explicit cache identity
// (the view key: "VOO", "VOO~raw", or the original ISIN for fund sources).
func (c *Client) saveCacheAs(cacheID string, s *Series, from time.Time) {
	cf := cacheFile{
		Symbol:        s.Symbol,
		Name:          s.Name,
		Currency:      s.Currency,
		Source:        s.Source,
		RequestedFrom: from.Format("2006-01-02"),
		FetchedAt:     time.Now(),
		Dates:         make([]string, 0, len(s.Points)),
		Closes:        make([]float64, 0, len(s.Points)),
	}
	for _, p := range s.Points {
		cf.Dates = append(cf.Dates, p.Date.Format("2006-01-02"))
		cf.Closes = append(cf.Closes, p.Close)
	}
	for _, d := range s.Dividends {
		cf.DivDates = append(cf.DivDates, d.Date.Format("2006-01-02"))
		cf.DivAmounts = append(cf.DivAmounts, d.Amount)
	}
	data, err := json.Marshal(cf)
	if err != nil {
		return
	}
	c.writeCacheFile(c.cachePath(cacheID), data)
}

// writeCacheFile writes data atomically, creating the cache directory on
// demand. A cache-less client (empty CacheDir) never touches the disk.
func (c *Client) writeCacheFile(path string, data []byte) {
	if c.CacheDir == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		c.Logf("warning: cache directory unusable: %v", err)
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		c.Logf("warning: cache write failed: %v", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		c.Logf("warning: cache write failed: %v", err)
	}
}

// sanitizeFilename keeps cache file names portable for symbols like ^GSPC or GC=F.
func sanitizeFilename(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '-', r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}
