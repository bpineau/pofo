package marketdata

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"
)

// Bundled long DAILY "<CCY>USD" histories, each in the "USD per one foreign
// unit" convention. They extend the fetchable Yahoo crosses (~2003→) back to
// 1971 so a foreign-currency backcast and the share-class reconstructions cover
// a long retirement at daily granularity throughout. Every file carries real
// daily FRED noon rates; the euro additionally chains monthly ECU/USD anchors
// (1:1 to the euro) over the Bundesbank Frankfurt DM/USD fixing before 1999.
// See each file's header for regeneration.
//
//go:embed data/eurusd-long.csv
var eurusdLongCSV string

//go:embed data/gbpusd-long.csv
var gbpusdLongCSV string

//go:embed data/jpyusd-long.csv
var jpyusdLongCSV string

//go:embed data/chfusd-long.csv
var chfusdLongCSV string

// longFXProxies lists every bundled cross, keyed by the foreign currency quoted
// against the US dollar. csv holds the "USD per one unit" daily history; tag is
// the provenance suffix stamped on the spliced series name.
var longFXProxies = []struct{ ccy, csv, tag string }{
	{"EUR", eurusdLongCSV, "ECU/EUR long"},
	{"GBP", gbpusdLongCSV, "GBP long"},
	{"JPY", jpyusdLongCSV, "JPY long"},
	{"CHF", chfusdLongCSV, "CHF long"},
}

// longFXCross returns a bundled long proxy for a currency cross, expressed in
// the requested direction. It triangulates through the US dollar from the
// bundled "USD per unit" histories: a cross "XY=X" quotes Y per X, which is
// (USD per X) / (USD per Y). This covers any pair among the dollar and the
// bundled currencies (EUR, GBP, JPY, CHF) with at least one non-dollar leg, so
// "<CCY>USD=X" resolves directly, "USD<CCY>=X" as the reciprocal, and a
// cross-euro pair such as "GBPEUR=X" as GBP/USD ÷ EUR/USD. tag is the proxy's
// provenance suffix. ok is false when a leg carries no bundled history, so the
// splice only ever touches crosses it can extend.
func longFXCross(symbol string) (proxy []Point, tag string, ok bool) {
	base, quote, valid := fxCross(symbol)
	if !valid {
		return nil, "", false
	}
	num, numTag, numOK := usdPerUnit(base)  // USD per base (nil for USD itself)
	den, denTag, denOK := usdPerUnit(quote) // USD per quote
	switch {
	case !numOK || !denOK: // a leg is neither the dollar nor a bundled currency
		return nil, "", false
	case num == nil && den == nil: // USD/USD: nothing to extend
		return nil, "", false
	case den == nil: // <CCY>USD: USD per base, directly
		return num, numTag, true
	case num == nil: // USD<CCY>: the reciprocal
		return reciprocalPoints(den), denTag, true
	default: // <CCY1><CCY2>: base/quote, triangulated through the dollar
		return divideSeries(num, den), numTag + "÷" + denTag, true
	}
}

// usdPerUnit returns the bundled long "USD per one unit of ccy" daily history.
// The dollar itself is the constant 1, signalled by a nil series with ok true;
// ok is false for a currency without a bundled proxy.
func usdPerUnit(ccy string) (series []Point, tag string, ok bool) {
	if ccy == "USD" {
		return nil, "", true
	}
	for _, x := range longFXProxies {
		if x.ccy == ccy {
			return parseAnchors(x.csv, "2006-01-02", "2006-01"), x.tag, true
		}
	}
	return nil, "", false
}

// reciprocalPoints returns 1/close for each strictly positive point, order
// preserved; it flips a "<CCY>USD" proxy into "USD<CCY>" and back.
func reciprocalPoints(in []Point) []Point {
	out := make([]Point, 0, len(in))
	for _, p := range in {
		if p.Close > 0 {
			out = append(out, Point{Date: p.Date, Close: 1 / p.Close})
		}
	}
	return out
}

// divideSeries returns num/den on num's date grid, den forward-filled to each
// date (points before den's first quote, or with a non-positive rate, are
// dropped). Both must be sorted ascending. With num = USD-per-X and
// den = USD-per-Y it yields the "Y per X" cross.
func divideSeries(num, den []Point) []Point {
	out := make([]Point, 0, len(num))
	j, d := 0, 0.0
	for _, p := range num {
		for j < len(den) && !den[j].Date.After(p.Date) {
			d = den[j].Close
			j++
		}
		if d > 0 && p.Close > 0 {
			out = append(out, Point{Date: p.Date, Close: p.Close / d})
		}
	}
	return out
}

// extendFXBack splices the bundled long daily history behind a freshly fetched
// cross so conversion between the dollar and the bundled currencies, and the
// cross-euro pairs among them, reach back to 1971 (with the reconstructions
// that read the cross). Any cross without a bundled proxy is left untouched.
func extendFXBack(symbol string, s *Series) {
	if s == nil {
		return
	}
	if proxy, tag, ok := longFXCross(symbol); ok {
		ExtendBack(s, &Series{Symbol: symbol + " (" + tag + ")", Points: proxy})
	}
}

// ConvertCurrency returns a copy of s converted into the target currency
// using daily FX rates (Yahoo "<FROM><TO>=X" crosses, forward-filled on the
// asset's trading days). Pence-quoted series (GBp) are first scaled to GBP.
// When the FX history starts after the series does, earlier points use the
// first known rate; extrapolatedBefore reports that date (zero when exact).
// A series whose currency is empty or already the target is returned as-is.
func (c *Client) ConvertCurrency(ctx context.Context, s *Series, target string, from time.Time) (out *Series, extrapolatedBefore time.Time, err error) {
	target = strings.ToUpper(strings.TrimSpace(target))
	if target == "" || s.Currency == "" || len(s.Points) == 0 {
		return s, time.Time{}, nil
	}
	src := s.Currency
	scale := 1.0
	if src == "GBp" || src == "GBX" {
		src, scale = "GBP", 0.01
	}
	if src == target {
		if scale == 1.0 {
			return s, time.Time{}, nil
		}
		out := *s
		out.Currency = target
		out.Points = make([]Point, len(s.Points))
		for i, p := range s.Points {
			out.Points[i] = Point{Date: p.Date, Close: p.Close * scale}
		}
		out.Dividends = make([]Dividend, len(s.Dividends))
		for i, d := range s.Dividends {
			out.Dividends[i] = Dividend{Date: d.Date, Amount: d.Amount * scale}
		}
		return &out, time.Time{}, nil
	}

	fxFrom := from
	if first := s.First().Date; first.Before(fxFrom) {
		fxFrom = first
	}
	fx, err := c.fxHistory(ctx, src, target, fxFrom)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("FX rate %s→%s: %w", src, target, err)
	}
	if len(fx.Points) == 0 {
		return nil, time.Time{}, fmt.Errorf("FX rate %s→%s: empty series", src, target)
	}

	cp := *s
	cp.Currency = target
	cp.Points = make([]Point, len(s.Points))
	j := 0
	rate := fx.Points[0].Close // backward extrapolation before FX history
	for i, p := range s.Points {
		for j < len(fx.Points) && !fx.Points[j].Date.After(p.Date) {
			rate = fx.Points[j].Close
			j++
		}
		if p.Date.Before(fx.Points[0].Date) && extrapolatedBefore.IsZero() {
			extrapolatedBefore = fx.Points[0].Date
		}
		cp.Points[i] = Point{Date: p.Date, Close: p.Close * scale * rate}
	}
	// Dividends convert at their ex-date with the same forward-filled
	// crosses (the first rate extrapolated backward, like points).
	cp.Dividends = make([]Dividend, len(s.Dividends))
	for i, d := range s.Dividends {
		rate, _, ok := fx.At(d.Date)
		if !ok {
			rate = fx.Points[0].Close
		}
		cp.Dividends[i] = Dividend{Date: d.Date, Amount: d.Amount * scale * rate}
	}
	return &cp, extrapolatedBefore, nil
}

// FXRate returns the multiplier turning an amount quoted in `from` into
// `to` at the given time, using the same daily "<FROM><TO>=X" cross as
// ConvertCurrency, forward-filled to the requested date (a weekend or
// holiday uses the last quoted cross). It errors when the date predates
// the available FX history; ConvertCurrency instead holds the earliest
// rate flat there, which suits series conversion but would silently skew
// a point-in-time quote.
func (c *Client) FXRate(ctx context.Context, from, to string, at time.Time) (float64, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))
	if from == to {
		return 1, nil
	}
	fx, err := c.fxHistory(ctx, from, to, time.Time{})
	if err != nil {
		return 0, fmt.Errorf("FX rate %s→%s: %w", from, to, err)
	}
	rate, _, ok := fx.At(dayUTC(at))
	if !ok {
		return 0, fmt.Errorf("no %s→%s rate on or before %s", from, to, at.Format("2006-01-02"))
	}
	return rate, nil
}

// fxHistory returns the src→target daily FX cross from Yahoo, with Stooq
// then the ECB reference rates as fallbacks (see historyFallback). It
// always fetches
// under a FIXED (zero) start so the cache key is constant across assets: the
// caller passes each asset's own first date, which would otherwise miss the
// cache and refetch the FX series once per converted asset. The full cross is
// small and covers every asset's range; the bundled crosses (EUR, GBP, JPY,
// CHF) are additionally extended back to 1971 by their long proxy (see
// extendFXBack), and dates before any cross starts are held constant by the
// caller.
func (c *Client) fxHistory(ctx context.Context, src, target string, _ time.Time) (*Series, error) {
	return c.History(ctx, src+target+"=X", time.Time{})
}
