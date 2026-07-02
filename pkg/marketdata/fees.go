package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// feesMaxAge is how long a fetched (or missed) TER stays cached: published
// ongoing charges rarely change.
const feesMaxAge = 180 * 24 * time.Hour

// feesEntry is one record of the on-disk fees cache (data/fees.json).
// TER is in percent per year; a negative TER records a known miss, so that
// unresolvable assets are not re-queried on every run.
type feesEntry struct {
	TER    float64   `json:"ter"`
	Source string    `json:"source"`
	AsOf   time.Time `json:"as_of"`
}

// Fees returns the published annual ongoing charge (TER, in percent per
// year) of an asset, with its provenance. Cascade: pinned catalog value,
// then the on-disk cache, then the FT tearsheets and justETF. ok is false
// when no source knows the asset. Prices and NAVs are already net of these
// fees: the figure is informational.
func (c *Client) Fees(ctx context.Context, id string) (ter float64, ok bool) {
	canonical := CanonicalID(id)
	if e, found := catalogByID()[canonical]; found && e.Fees > 0 {
		return e.Fees, true
	}
	if e, found := c.feesLookup(canonical); found && time.Since(e.AsOf) < feesMaxAge {
		return e.TER, e.TER >= 0
	}
	ter, src, err := c.fetchFees(ctx, canonical)
	if err != nil {
		c.Logf("fees unknown for %s (%v)", canonical, err)
		c.saveFeesEntry(canonical, feesEntry{TER: -1, Source: "none", AsOf: time.Now()})
		return 0, false
	}
	c.Logf("fees for %s: %.2f %%/yr (%s)", canonical, ter, src)
	c.saveFeesEntry(canonical, feesEntry{TER: ter, Source: src, AsOf: time.Now()})
	return ter, true
}

// fetchFees tries every known fee source for a canonical identifier.
func (c *Client) fetchFees(ctx context.Context, canonical string) (float64, string, error) {
	res, _ := c.loadResolution(canonical)
	isin := ""
	if IsISIN(canonical) {
		isin = canonical
	} else if e, found := catalogByID()[canonical]; found {
		isin = e.ISIN
	}
	currency := res.Currency
	var errs []string

	if isin != "" {
		// The FT funds tearsheet covers European mutual funds and many ETFs by ISIN.
		for _, cur := range candidateCurrencies(currency, isin) {
			if ter, err := c.ftTearsheetFees(ctx, "funds", isin+":"+cur); err == nil {
				return ter, "FT", nil
			}
		}
		errs = append(errs, "FT funds: not found")
		if ter, err := c.justETFFees(ctx, isin); err == nil {
			return ter, "justETF", nil
		} else {
			errs = append(errs, fmt.Sprintf("justETF: %v", err))
		}
	}
	// US-listed ETFs and mutual funds, by plain Yahoo symbol.
	if sym := res.Symbol; sym != "" && !strings.ContainsAny(sym, ".^=") && res.Source == "yahoo" {
		for _, mic := range []string{"PCQ", "NMQ", "NSQ"} {
			if ter, err := c.ftTearsheetFees(ctx, "etfs", sym+":"+mic+":USD"); err == nil {
				return ter, "FT", nil
			}
		}
		if ter, err := c.ftTearsheetFees(ctx, "funds", sym); err == nil {
			return ter, "FT", nil
		}
		errs = append(errs, "FT US etfs/funds: not found")
	}
	if len(errs) == 0 {
		errs = append(errs, "no applicable source")
	}
	return 0, "", fmt.Errorf("%s", strings.Join(errs, "; "))
}

// candidateCurrencies orders the share-class currencies to try on FT.
func candidateCurrencies(known, isin string) []string {
	if known != "" && known != "GBp" {
		return []string{known}
	}
	if strings.HasPrefix(isin, "FR") || strings.HasPrefix(isin, "LU") || strings.HasPrefix(isin, "DE") {
		return []string{"EUR", "USD"}
	}
	return []string{"USD", "EUR"}
}

var (
	ftFeesRe      = regexp.MustCompile(`(?i)(?:Ongoing charge|Net expense ratio)</th><td[^>]*>([0-9]+[.,][0-9]+)%`)
	justETFFeesRe = regexp.MustCompile(`(?i)ter-value">\s*([0-9]+[.,][0-9]+)%`)
)

// ftTearsheetFees scrapes the ongoing charge from an FT tearsheet
// (kind: "funds" or "etfs").
func (c *Client) ftTearsheetFees(ctx context.Context, kind, symbol string) (float64, error) {
	u := fmt.Sprintf("%s/data/%s/tearsheet/summary?s=%s", c.FTBase, kind, url.QueryEscape(symbol))
	body, err := c.get(ctx, u)
	if err != nil {
		return 0, err
	}
	return parseFeesMatch(ftFeesRe, body)
}

// justETFFees scrapes the TER from a justETF profile page (UCITS ETFs only).
func (c *Client) justETFFees(ctx context.Context, isin string) (float64, error) {
	u := fmt.Sprintf("%s/en/etf-profile.html?isin=%s", c.JustETFBase, url.QueryEscape(isin))
	body, err := c.get(ctx, u)
	if err != nil {
		return 0, err
	}
	return parseFeesMatch(justETFFeesRe, body)
}

func parseFeesMatch(re *regexp.Regexp, body []byte) (float64, error) {
	m := re.FindSubmatch(body)
	if m == nil {
		return 0, fmt.Errorf("no fees on the page")
	}
	ter, err := strconv.ParseFloat(strings.ReplaceAll(string(m[1]), ",", "."), 64)
	if err != nil || ter < 0 || ter > 20 {
		return 0, fmt.Errorf("unreadable fees: %q", m[1])
	}
	return ter, nil
}

func (c *Client) feesCachePath() string {
	return filepath.Join(c.CacheDir, "fees.json")
}

// feesLookup reads the TER cache, loading fees.json into memory on first
// use; subsequent calls never touch the disk again.
func (c *Client) feesLookup(id string) (feesEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fees == nil {
		c.fees = map[string]feesEntry{}
		if c.CacheDir != "" {
			if data, err := os.ReadFile(c.feesCachePath()); err == nil {
				_ = json.Unmarshal(data, &c.fees)
			}
		}
	}
	e, ok := c.fees[id]
	return e, ok
}

// saveFeesEntry updates the in-memory TER cache and persists it, without
// re-reading the file.
func (c *Client) saveFeesEntry(id string, e feesEntry) {
	c.feesLookup(id) // guarantees the initial load
	c.mu.Lock()
	c.fees[id] = e
	data, err := json.MarshalIndent(c.fees, "", " ")
	c.mu.Unlock()
	if err == nil {
		c.writeCacheFile(c.feesCachePath(), data)
	}
}
