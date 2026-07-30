package marketdata

import (
	_ "embed"
	"fmt"
	"strings"
	"time"
)

// eurusdLongCSV is the bundled long monthly EUR/USD (USD per EUR) history:
// ECU/USD before 1999 chained 1:1 to the euro after, used to extend the
// fetchable EURUSD=X cross (Yahoo, ~2003→) back to 1978 so a EUR-investor
// backcast and the EUR share-class reconstructions cover a long retirement.
//
//go:embed data/eurusd-long.csv
var eurusdLongCSV string

// eurusdLongCross returns the bundled long EUR/USD proxy expressed as the given
// currency cross: "EURUSD=X" (USD per EUR) directly, "USDEUR=X" (EUR per USD)
// as the reciprocal. ok is false for any other symbol, so the splice only ever
// touches the euro cross.
func eurusdLongCross(symbol string) (proxy []Point, ok bool) {
	anchors := parseMonthlyAnchors(eurusdLongCSV)
	switch symbol {
	case "EURUSD=X":
		return anchors, true
	case "USDEUR=X":
		out := make([]Point, 0, len(anchors))
		for _, p := range anchors {
			if p.Close > 0 {
				out = append(out, Point{Date: p.Date, Close: 1 / p.Close})
			}
		}
		return out, true
	}
	return nil, false
}

// extendFXBack splices the bundled long monthly EUR/USD history behind a
// freshly fetched euro cross so USD↔EUR conversion (and the EUR reconstructions
// that read the cross) reach back to 1978. Any other symbol is left untouched.
func extendFXBack(symbol string, s *Series) {
	if s == nil {
		return
	}
	if proxy, ok := eurusdLongCross(symbol); ok {
		ExtendBack(s, &Series{Symbol: symbol + " (ECU/EUR long)", Points: proxy})
	}
}

// ConvertCurrency returns a copy of s converted into the target currency
// using daily FX rates (Yahoo "<FROM><TO>=X" crosses, forward-filled on the
// asset's trading days). Pence-quoted series (GBp) are first scaled to GBP.
// When the FX history starts after the series does, earlier points use the
// first known rate; extrapolatedBefore reports that date (zero when exact).
// A series whose currency is empty or already the target is returned as-is.
func (c *Client) ConvertCurrency(s *Series, target string, from time.Time) (out *Series, extrapolatedBefore time.Time, err error) {
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
		return &out, time.Time{}, nil
	}

	fxFrom := from
	if first := s.First().Date; first.Before(fxFrom) {
		fxFrom = first
	}
	fx, err := c.fxHistory(src, target, fxFrom)
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
	return &cp, extrapolatedBefore, nil
}

// fxHistory returns the src→target daily FX cross from Yahoo. It always fetches
// under a FIXED (zero) start so the cache key is constant across assets: the
// caller passes each asset's own first date, which would otherwise miss the
// cache and refetch the FX series once per converted asset. The full cross is
// small and covers every asset's range; the euro cross is additionally
// extended back to 1978 by the bundled ECU/EUR proxy (see extendFXBack), and
// dates before any cross starts are held constant by the caller.
func (c *Client) fxHistory(src, target string, _ time.Time) (*Series, error) {
	return c.History(src+target+"=X", time.Time{})
}
