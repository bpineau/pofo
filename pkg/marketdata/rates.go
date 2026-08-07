package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Policy and money-market rate symbols.
//
// The yield symbols ("^IRX", "^TNX", …) are market rates quoted by an exchange
// feed. This family adds the rates that SET them: what a central bank pays or
// charges, and what banks lend to each other at. They are the reference leg of
// every floating-rate instrument a portfolio can hold, so drawing one next to a
// fund answers the question that always follows a floating-rate chart, "how
// much of this is the coupon and how much is the spread?".
//
// Like the yield symbols, they are annualized percent LEVELS, not prices: they
// carry no currency, they may legitimately be zero or NEGATIVE (the euro area
// lived below zero from 2014 to 2022), and they never belong in a return
// computation directly. Chart them and read them.
//
// Each symbol is taken from the institution that sets it, with a second source
// behind it when one exists: the ECB's own publication for the euro rates
// (through the DBnomics mirror of the ECB Data Portal, which also carries
// Euribor, licensed by EMMI and absent from most free feeds), and the New York
// Fed plus FRED for the American ones. The sources are tried in order, so a
// throttled provider costs a retry, not a hole.
//
// Two depth caveats worth knowing. ^FEDFUNDS reaches 1954 from FRED but only
// 2000 from the New York Fed, so a series fetched while FRED was unreachable is
// shorter (delete its cache entry to deepen it later). And ^ECB-MRO is the
// FIXED-rate tender level, which the ECB did not publish between June 2000 and
// October 2008, when the operations ran as variable-rate tenders: that
// eight-year hole is in the data, not in the fetch.
const (
	rateFRED     = "fred"
	rateDBnomics = "dbnomics"
	rateNYFed    = "nyfed"
)

// rateSource is one place a rate can be read: a provider and the identifier it
// knows the series by. field selects the column for a provider that publishes
// several per row (the New York Fed serves the effective rate and the FOMC
// target range in the same payload).
type rateSource struct {
	provider string
	id       string
	field    string // nyfed only: "" for the rate itself, or "targetRateTo"
}

// rateSpec describes one symbol: where to read it, and at what cadence.
type rateSpec struct {
	sources []rateSource
	name    string
	monthly bool // periods are months, dated at month end
}

var policyRates = map[string]rateSpec{
	"^ESTR": {
		sources: []rateSource{{rateDBnomics, "ECB/EST/B.EU000A2X2A25.WT", ""}},
		name:    "Euro short-term rate (ESTR, overnight)",
	},
	"^ECB-DFR": {
		sources: []rateSource{{rateDBnomics, "ECB/FM/D.U2.EUR.4F.KR.DFR.LEV", ""}, {rateFRED, "ECBDFR", ""}},
		name:    "ECB deposit facility rate",
	},
	"^ECB-MRO": {
		sources: []rateSource{{rateDBnomics, "ECB/FM/D.U2.EUR.4F.KR.MRR_FR.LEV", ""}, {rateFRED, "ECBMRRFR", ""}},
		name:    "ECB main refinancing rate (fixed-rate tenders)",
	},
	"^EURIBOR3M": {
		sources: []rateSource{{rateDBnomics, "ECB/FM/M.U2.EUR.RT.MM.EURIBOR3MD_.HSTA", ""}},
		name:    "Euribor 3-month (monthly average)",
		monthly: true,
	},
	"^SOFR": {
		sources: []rateSource{{rateNYFed, "secured/sofr", ""}, {rateFRED, "SOFR", ""}},
		name:    "Secured overnight financing rate (SOFR)",
	},
	"^FEDFUNDS": {
		sources: []rateSource{{rateFRED, "DFF", ""}, {rateNYFed, "unsecured/effr", ""}},
		name:    "Federal funds effective rate",
	},
	"^FED-TARGET": {
		sources: []rateSource{{rateNYFed, "unsecured/effr", "targetRateTo"}, {rateFRED, "DFEDTARU", ""}},
		name:    "Federal funds target rate (upper bound)",
	},
}

// isPolicyRate reports whether a symbol belongs to the rate registry.
func isPolicyRate(symbol string) bool {
	_, ok := policyRates[symbol]
	return ok
}

// RateSymbols lists the policy and money-market rate symbols, sorted, so a
// caller can offer them without duplicating the registry.
func RateSymbols() []string {
	out := make([]string, 0, len(policyRates))
	for s := range policyRates {
		out = append(out, s)
	}
	sortStrings(out)
	return out
}

// RateName returns the display name of a rate symbol, or "" if unknown.
func RateName(symbol string) string { return policyRates[symbol].name }

// fetchPolicyRate serves a rate symbol through the normal disk cache, trying
// its sources in order. No cleaning pass runs on it: these are published
// official levels, and the repairs that protect prices (dropouts, denomination
// breaks) would read a legitimate zero or negative rate as a bad print.
func (c *Client) fetchPolicyRate(ctx context.Context, symbol string, from time.Time) (*Series, error) {
	spec := policyRates[symbol]
	return c.cachedHistory(ctx, spec.sources[0].provider, symbol, from, false, func() (*Series, error) {
		var last error
		for _, src := range spec.sources {
			pts, err := c.fetchRateFrom(ctx, src, spec.monthly)
			if err == nil {
				return Trim(&Series{Symbol: symbol, Name: spec.name, Source: src.provider, Points: pts}, from, time.Time{}), nil
			}
			if ctx.Err() != nil {
				return nil, err
			}
			c.Logf("%s: %s unavailable (%v)", symbol, src.provider, err)
			last = err
		}
		return nil, fmt.Errorf("%s: no source answered: %w", symbol, last)
	})
}

func (c *Client) fetchRateFrom(ctx context.Context, src rateSource, monthly bool) ([]Point, error) {
	switch src.provider {
	case rateFRED:
		return c.fetchFREDRate(ctx, src.id)
	case rateDBnomics:
		return c.fetchDBnomics(ctx, src.id, monthly)
	case rateNYFed:
		return c.fetchNYFed(ctx, src.id, src.field)
	}
	return nil, fmt.Errorf("unknown rate provider %q", src.provider)
}

// fetchDBnomics reads one series from the DBnomics API ("provider/dataset/code")
// and returns its observations. Monthly periods ("2026-01") are dated at the
// END of their month, so a monthly rate lines up with the month-end anchors the
// rest of the toolkit uses instead of sliding a month early.
func (c *Client) fetchDBnomics(ctx context.Context, path string, monthly bool) ([]Point, error) {
	u := fmt.Sprintf("%s/v22/series/%s?observations=1", c.DBnomicsBase, path)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("dbnomics %s: %w", path, err)
	}
	var payload struct {
		Series struct {
			Docs []struct {
				Period []string    `json:"period"`
				Value  []jsonValue `json:"value"`
			} `json:"docs"`
		} `json:"series"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("dbnomics %s: unreadable response: %w", path, err)
	}
	if len(payload.Series.Docs) == 0 {
		return nil, fmt.Errorf("dbnomics %s: no such series", path)
	}
	doc := payload.Series.Docs[0]
	pts := make([]Point, 0, len(doc.Period))
	for i, p := range doc.Period {
		if i >= len(doc.Value) || !doc.Value[i].ok {
			continue // "NA" observations are published as a string
		}
		d, err := parsePeriod(p, monthly)
		if err != nil {
			continue
		}
		pts = append(pts, Point{Date: d, Close: doc.Value[i].v})
	}
	if len(pts) == 0 {
		return nil, fmt.Errorf("dbnomics %s: no usable observation", path)
	}
	return pts, nil
}

// nyFedWindow is the date range asked of the New York Fed reference-rate API,
// wide enough to hold any of its series (the effective funds rate starts in
// 2000, SOFR in 2018) so one request returns the whole history. Its "last/N"
// endpoint refuses large N, hence the search form.
const nyFedWindow = "startDate=1990-01-01&endDate=2099-12-31"

// fetchNYFed reads a reference rate from the New York Fed, the institution that
// computes SOFR and the effective federal funds rate. field selects the column:
// empty for the rate itself, "targetRateTo" for the upper bound of the FOMC
// target range, which the same payload carries.
func (c *Client) fetchNYFed(ctx context.Context, path, field string) ([]Point, error) {
	u := fmt.Sprintf("%s/api/rates/%s/search.json?%s", c.NYFedBase, path, nyFedWindow)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("nyfed %s: %w", path, err)
	}
	var payload struct {
		RefRates []struct {
			Date   string   `json:"effectiveDate"`
			Rate   *float64 `json:"percentRate"`
			Target *float64 `json:"targetRateTo"`
		} `json:"refRates"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("nyfed %s: unreadable response: %w", path, err)
	}
	pts := make([]Point, 0, len(payload.RefRates))
	for _, r := range payload.RefRates {
		v := r.Rate
		if field == "targetRateTo" {
			v = r.Target
		}
		if v == nil {
			continue
		}
		d, err := time.ParseInLocation("2006-01-02", r.Date, time.UTC)
		if err != nil {
			continue
		}
		pts = append(pts, Point{Date: d, Close: *v})
	}
	if len(pts) < 2 {
		return nil, fmt.Errorf("nyfed %s: only %d usable rows", path, len(pts))
	}
	// The API answers newest first; every series in the toolkit runs forward.
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
	return pts, nil
}

// jsonValue decodes a DBnomics observation, which is a number or the string
// "NA" when the period has no value.
type jsonValue struct {
	v  float64
	ok bool
}

func (j *jsonValue) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' { // "NA"
		return nil
	}
	if err := json.Unmarshal(b, &j.v); err != nil {
		return err
	}
	j.ok = true
	return nil
}

// parsePeriod turns a DBnomics period into a date: "2026-01-31" as is, and
// "2026-01" at the last day of the month when the series is monthly.
func parsePeriod(p string, monthly bool) (time.Time, error) {
	if !monthly {
		return time.Parse("2006-01-02", p)
	}
	t, err := time.Parse("2006-01", p)
	if err != nil {
		return time.Time{}, err
	}
	return t.AddDate(0, 1, -1), nil
}

// sortStrings is an insertion sort, adequate for a handful of symbols and
// cheaper than importing sort for one call site.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
