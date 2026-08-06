package marketdata

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// hicpPrefix marks the synthetic identifiers served from Eurostat's Harmonised
// Index of Consumer Prices, e.g. "^HICP-FR" (France), "^HICP-EA" (euro area).
const hicpPrefix = "^HICP-"

// hicpGeo reports the Eurostat geo code of a HICP identifier ("^HICP-FR" ->
// "FR"). ok is false for any other symbol.
func hicpGeo(symbol string) (geo string, ok bool) {
	geo = strings.TrimPrefix(symbol, hicpPrefix)
	if geo == symbol || geo == "" {
		return "", false
	}
	return geo, true
}

// hicpName gives a few well-known geographies a readable name; others fall
// back to the bare code.
var hicpName = map[string]string{
	"FR": "France",
	"EA": "euro area",
	"EU": "European Union",
	"DE": "Germany",
}

// fetchHICP returns the daily-interpolated HICP index for a geography, cached
// like the other non-Yahoo sources.
func (c *Client) fetchHICP(symbol, geo string, from time.Time) (*Series, error) {
	return c.cachedHistory("eurostat", symbol, from, func() (*Series, error) {
		return c.downloadHICP(symbol, geo)
	})
}

// downloadHICP fetches the monthly all-items HICP (2015=100) for geo from the
// Eurostat dissemination API and interpolates it to a daily index.
func (c *Client) downloadHICP(symbol, geo string) (*Series, error) {
	u := fmt.Sprintf("%s/eurostat/api/dissemination/statistics/1.0/data/prc_hicp_midx"+
		"?format=JSON&lang=EN&freq=M&unit=I15&coicop=CP00&geo=%s", c.EurostatBase, url.QueryEscape(geo))
	body, err := c.get(u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Value     map[string]float64 `json:"value"`
		Dimension struct {
			Time struct {
				Category struct {
					Index map[string]int `json:"index"`
				} `json:"category"`
			} `json:"time"`
		} `json:"dimension"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("eurostat HICP %s: %w", geo, err)
	}

	// With every dimension but time filtered to a single category, the JSON-stat
	// linear index equals the time index, so value[pos] is that month's level.
	monthly := make([]Point, 0, len(resp.Dimension.Time.Category.Index))
	for label, pos := range resp.Dimension.Time.Category.Index {
		v, ok := resp.Value[strconv.Itoa(pos)]
		if !ok || v <= 0 {
			continue
		}
		t, err := time.ParseInLocation("2006-01", label, time.UTC)
		if err != nil {
			continue
		}
		monthly = append(monthly, Point{Date: t, Close: v})
	}
	if len(monthly) < 2 {
		return nil, fmt.Errorf("eurostat HICP %s: only %d monthly points", geo, len(monthly))
	}
	sort.Slice(monthly, func(i, j int) bool { return monthly[i].Date.Before(monthly[j].Date) })

	name := geo
	if n, ok := hicpName[geo]; ok {
		name = n
	}
	return &Series{
		Symbol: symbol,
		Name:   fmt.Sprintf("HICP %s (all-items, 2015=100)", name),
		Source: "eurostat",
		Points: monthlyToDaily(monthly),
	}, nil
}

// monthlyToDaily expands monthly index anchors into a daily series, spreading
// each month's change geometrically across its calendar days so the curve
// compounds smoothly (no month-boundary steps) while every anchor is hit
// exactly. anchors must be sorted ascending.
func monthlyToDaily(anchors []Point) []Point {
	if len(anchors) < 2 {
		return anchors
	}
	out := make([]Point, 0, len(anchors)*31)
	for i := 0; i+1 < len(anchors); i++ {
		a, b := anchors[i], anchors[i+1]
		gap := int(math.Round(b.Date.Sub(a.Date).Hours() / 24))
		ratio := b.Close / a.Close
		for k := range gap {
			out = append(out, Point{
				Date:  a.Date.AddDate(0, 0, k),
				Close: a.Close * math.Pow(ratio, float64(k)/float64(gap)),
			})
		}
	}
	return append(out, anchors[len(anchors)-1])
}
