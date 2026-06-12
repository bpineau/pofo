package marketdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// cacheFile is the JSON document stored in the cache directory, one per symbol.
type cacheFile struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Currency      string    `json:"currency"`
	Source        string    `json:"source"`
	RequestedFrom string    `json:"requested_from"`
	FetchedAt     time.Time `json:"fetched_at"`
	Dates         []string  `json:"dates"`
	Closes        []float64 `json:"closes"`
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
	return s, cf.FetchedAt, true
}

// saveCache persists a downloaded series; failures are logged, never fatal.
func (c *Client) saveCache(s *Series, from time.Time) {
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
	data, err := json.Marshal(cf)
	if err != nil {
		return
	}
	c.writeCacheFile(c.cachePath(s.Symbol), data)
}

// writeCacheFile writes data atomically, creating the cache directory on demand.
func (c *Client) writeCacheFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		c.Logf("avertissement: répertoire de cache inutilisable: %v", err)
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		c.Logf("avertissement: écriture du cache: %v", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		c.Logf("avertissement: écriture du cache: %v", err)
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
