package marketdata

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// hicpFRSnapshot is a bundled offline fallback for ^HICP-FR: the monthly index
// anchors (long history, ~1955→), used when the live Eurostat API is
// unreachable. It carries the same chain the live path builds (Eurostat HICP
// from 1996 with the OECD French CPI (FRED FRACPIALLMINMEI) spliced before it)
// so an offline run still deflates the high-inflation 1955-1990s a long
// retirement backcast needs. The live series is always preferred; this only
// needs refreshing occasionally to keep the recent tail current. Regenerate
// (requires curl + jq) with:
//
//	url='https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data/prc_hicp_midx?format=JSON&lang=EN&freq=M&unit=I15&coicop=CP00&geo=FR'
//	curl -s "$url" | jq -r '.dimension.time.category.index as $i | .value as $v |
//	  ($i|to_entries|sort_by(.value)[]) | select($v[(.value|tostring)]!=null) |
//	  "\(.key),\($v[(.value|tostring)])"'
//
// then prepend the rescaled FRED FRACPIALLMINMEI months before 1996 (chain at
// the overlap, as extendMonthlyBack does), keeping the comment header.
//
//go:embed data/hicp-fr.csv
var hicpFRSnapshot string

// embeddedHICP returns the bundled monthly anchors for a geography, if any.
func embeddedHICP(geo string) ([]Point, bool) {
	switch geo {
	case "FR":
		return parseMonthlyAnchors(hicpFRSnapshot), true
	}
	return nil, false
}

// parseMonthlyAnchors reads "YYYY-MM,value" lines; it backs every bundled
// monthly series (HICP and CPI indices).
func parseMonthlyAnchors(csv string) []Point { return parseAnchors(csv, "2006-01") }

// parseAnchors reads "date,value" lines whose date matches one of the given
// layouts (ignoring blanks and # comments), skipping any malformed row,
// sorted ascending by date. It backs every bundled snapshot series; a series
// may mix cadences (e.g. monthly anchors then daily rates) by listing both
// layouts.
func parseAnchors(csv string, layouts ...string) []Point {
	var pts []Point
	for _, line := range strings.Split(csv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		label, val, ok := strings.Cut(line, ",")
		if !ok {
			continue
		}
		var t time.Time
		err := errNoLayout
		for _, layout := range layouts {
			if t, err = time.ParseInLocation(layout, strings.TrimSpace(label), time.UTC); err == nil {
				break
			}
		}
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil || v <= 0 {
			continue
		}
		pts = append(pts, Point{Date: t, Close: v})
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].Date.Before(pts[j].Date) })
	return pts
}

// errNoLayout is parseAnchors' sentinel for a row matching none of the
// requested date layouts.
var errNoLayout = fmt.Errorf("date matches no layout")

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

// fetchHICP returns the daily-interpolated HICP index for a geography. The
// live Eurostat API is preferred (and disk-cached like the other non-Yahoo
// sources); if it is unreachable and no cached copy exists, a bundled snapshot
// keeps the series available offline, at the cost of missing the latest months.
func (c *Client) fetchHICP(ctx context.Context, symbol, geo string, from time.Time) (*Series, error) {
	s, err := c.cachedHistory(ctx, "eurostat", symbol, from, false, func() (*Series, error) {
		return c.downloadHICP(ctx, symbol, geo)
	})
	if err != nil {
		if anchors, ok := embeddedHICP(geo); ok {
			c.Logf("warning: Eurostat unavailable (%v), using the embedded %s snapshot (ends %s)",
				err, symbol, anchors[len(anchors)-1].Date.Format("2006-01"))
			return hicpSeries(symbol, geo, anchors), nil
		}
		return nil, err
	}
	return s, nil
}

// hicpSeries packages monthly anchors as a daily-interpolated index series.
func hicpSeries(symbol, geo string, monthly []Point) *Series {
	name := geo
	if n, ok := hicpName[geo]; ok {
		name = n
	}
	return &Series{
		Symbol: symbol,
		Name:   fmt.Sprintf("HICP %s (all-items, 2015=100)", name),
		Source: "eurostat",
		Points: monthlyToDaily(monthly),
	}
}

// downloadHICP fetches the monthly all-items HICP (2015=100) for geo from the
// Eurostat dissemination API and interpolates it to a daily index.
func (c *Client) downloadHICP(ctx context.Context, symbol, geo string) (*Series, error) {
	u := fmt.Sprintf("%s/eurostat/api/dissemination/statistics/1.0/data/prc_hicp_midx"+
		"?format=JSON&lang=EN&freq=M&unit=I15&coicop=CP00&geo=%s", c.EurostatBase, url.QueryEscape(geo))
	body, err := c.get(ctx, u)
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

	// Eurostat's harmonised index only starts in the mid-1990s. For France,
	// extend it back with the OECD national CPI from FRED (monthly, 1955→),
	// chained at the overlap, so real-return deflation covers the high-inflation
	// 1970s-80s a long retirement backcast needs. Best-effort: on failure the
	// series simply keeps its Eurostat start.
	if geo == "FR" {
		if older, ferr := c.fetchFRED(ctx, "FRACPIALLMINMEI"); ferr == nil {
			monthly = extendMonthlyBack(monthly, older)
		} else {
			c.Logf("warning: FRED French CPI unavailable (%v); %s starts %s", ferr, symbol, monthly[0].Date.Format("2006-01"))
		}
	}
	return hicpSeries(symbol, geo, monthly), nil
}

// extendMonthlyBack prepends older's history, rescaled to base's level at the
// splice month, before base's first date, chaining the two indices so the level
// stays continuous. Both slices must be ascending. It is the monthly-index
// analogue of ExtendBack.
func extendMonthlyBack(base, older []Point) []Point {
	if len(base) == 0 || len(older) == 0 {
		return base
	}
	anchor := base[0]
	i := sort.Search(len(older), func(i int) bool { return older[i].Date.After(anchor.Date) }) - 1
	if i < 0 || older[i].Close <= 0 {
		return base
	}
	scale := anchor.Close / older[i].Close
	var pre []Point
	for _, p := range older {
		if !p.Date.Before(anchor.Date) {
			break
		}
		pre = append(pre, Point{Date: p.Date, Close: p.Close * scale})
	}
	if len(pre) == 0 {
		return base
	}
	return append(pre, base...)
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
