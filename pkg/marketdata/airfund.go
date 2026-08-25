package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// The "airfund" source serves the daily NAV of a French employee-savings fund
// (FCPE) from the airfund.io delivery API, the service behind the NAV chart
// and the "Exporter les VLs" button of the management company's fund page.
// An FCPE has no ISIN, no exchange listing and no coverage on any quote site,
// so this is the only machine-readable NAV history there is.
//
// A catalog entry pins the source with its two identifiers: `symbol` is the
// share class code the company uses in place of an ISIN (Eres: 990000135629
// for "ERES XTRACKERS ACTIONS MONDE Part M"), `xid` is the id of the chart
// widget embedded on the fund page (the API refuses a request without it).
// The endpoint answers without any login, cookie or key.
//
// Offline, or when the API fails and nothing is cached, the bundled refdata
// NAV snapshot of the fund (refdata/<ID>-NAV.csv, refreshed by
// cmd/gen-eres-refdata) answers instead, exactly as a CPI deflator falls back
// on its embedded snapshot; the live series is always preferred.

// airfundChartPath is the NAV-history endpoint of the delivery API (POST,
// JSON in, JSON out).
const airfundChartPath = "/api/v1/navs-evolution-chart/data"

// fetchAirfund downloads the whole NAV history of one share class from the
// airfund delivery API. The API knows no start date: "inception" is the only
// period worth asking for, and the result is trimmed to from here.
func (c *Client) fetchAirfund(ctx context.Context, id string, res resolution, from time.Time) (*Series, error) {
	payload, err := json.Marshal(map[string]any{
		"locale":           "fr",
		"sId":              res.Xid,
		"isinCode":         res.Symbol,
		"maxPeriodCode":    "inception",
		"debug":            nil,
		"displayBenchmark": false,
	})
	if err != nil {
		return nil, err
	}
	body, err := c.post(ctx, c.AirfundBase+airfundChartPath, "application/json", payload)
	if err != nil {
		return nil, err
	}
	s, err := parseAirfundNAVs(body)
	if err != nil {
		return nil, fmt.Errorf("airfund %s: %w", id, err)
	}
	s.Symbol = id
	if res.Name != "" {
		s.Name = res.Name
	}
	if res.Currency != "" {
		s.Currency = res.Currency
	}
	return Trim(s, from, time.Time{}), nil
}

// parseAirfundNAVs decodes the chart payload: the fund name and one
// {date, value} per NAV day, in the fund's own currency (the API states none;
// the catalog record does). Rows without a usable value are skipped, and the
// result is sorted by date since nothing guarantees the API's order.
func parseAirfundNAVs(body []byte) (*Series, error) {
	var resp struct {
		FundName string `json:"fundName"`
		NAVs     []struct {
			Date  string   `json:"date"`
			Value *float64 `json:"value"`
		} `json:"navs"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable airfund response: %w", err)
	}
	s := &Series{Name: resp.FundName, Source: "airfund"}
	for _, n := range resp.NAVs {
		if n.Value == nil || *n.Value <= 0 {
			continue
		}
		d, err := time.ParseInLocation("2006-01-02", n.Date, time.UTC)
		if err != nil {
			continue
		}
		s.Points = append(s.Points, Point{Date: d, Close: *n.Value})
	}
	if len(s.Points) == 0 {
		return nil, fmt.Errorf("no NAV in response")
	}
	sort.Slice(s.Points, func(i, j int) bool { return s.Points[i].Date.Before(s.Points[j].Date) })
	return s, nil
}

// embeddedNAV serves the bundled NAV snapshot of a catalog "airfund" fund
// (refdata/<ID>-NAV.csv) when the live API cannot be reached: the same series
// the recipe of the fund is validated against, so an offline run stays
// consistent with the shipped reconstruction. ok is false for any other id.
//
// The id may carry the "~raw" view suffix (a raw fetch keys its cache that
// way, and a valuation consumer such as a portfolio tracker fetches raw), so
// it is stripped before the catalog lookup: without this the fallback misses
// and the caller wrongly falls through to a live ticker search.
func embeddedNAV(id string) (*Series, bool) {
	canonical := CanonicalID(strings.TrimSuffix(id, "~raw"))
	e, found := catalogByID()[canonical]
	if !found || e.Source != "airfund" {
		return nil, false
	}
	s, ok, err := ReadSimdataFS(datasets.Refdata(), canonical+"-NAV")
	if err != nil || !ok || len(s.Points) == 0 {
		return nil, false
	}
	s.Symbol, s.Name, s.Source = canonical, e.Name, "airfund"
	if s.Currency == "" {
		s.Currency = e.Currency
	}
	return s, true
}
