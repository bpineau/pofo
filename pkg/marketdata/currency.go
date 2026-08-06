package marketdata

import (
	"fmt"
	"strings"
	"time"
)

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
	fx, err := c.History(src+target+"=X", fxFrom)
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
