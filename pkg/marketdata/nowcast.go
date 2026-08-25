package marketdata

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// A fund priced once a day and published with a lag (a French employee-savings
// fund, whose NAV of day D appears on D+2 or so) leaves a gap between its last
// NAV and today. The catalog can name a NOWCAST PROXY for it: a listed asset
// whose moves, converted into the fund's currency, stand in for the fund from
// the last published NAV onward. Two consumers read it:
//
//   - Fetch extends the daily series to the proxy's last close (nowcastForward),
//     each estimated day carrying the proxy's return, and stamps EstimatedFrom
//     on the series so a reader can tell the tail apart. The estimate is never
//     cached and never enters a shipped dataset (Series.WithoutEstimates).
//   - Intraday builds today's path (nowcastIntraday): the last daily value
//     scaled by the proxy's intraday move since the close that value stands on,
//     the proxy's currency converted tick by tick.
//
// What the estimate ignores is stated: the fund's own charge (0.4 %/yr over a
// few days is below the cent) and the proxy's tracking of the fund, which the
// catalog note quantifies; for the FCPE the proxy is chosen for TIMING (a
// US-listed tracker closes when the fund's NAV is struck) rather than for a
// perfect fee match.

// nowcastForward appends to s the business days after its last point that the
// proxy has closed on, each carrying the proxy's return in s's currency. It
// returns s untouched when there is nothing to add or the proxy cannot be
// read, and a copy otherwise: the memoized series must stay real.
func (c *Client) nowcastForward(ctx context.Context, s *Series, proxyID string) *Series {
	if s == nil || len(s.Points) == 0 || proxyID == "" {
		return s
	}
	last := s.Last()
	p, err := c.FetchExtended(ctx, proxyID, FetchOptions{
		From: last.Date.AddDate(0, 0, -14), Currency: s.Currency, NoSim: true,
	})
	if err != nil {
		c.Logf("warning: %s: nowcast proxy %s unavailable (%v), series ends at the last NAV", s.Symbol, proxyID, err)
		return s
	}
	base, _, ok := p.At(last.Date)
	if !ok || base <= 0 {
		return s
	}
	i := sort.Search(len(p.Points), func(k int) bool { return p.Points[k].Date.After(last.Date) })
	if i >= len(p.Points) {
		return s
	}
	out := *s
	out.Points = make([]Point, len(s.Points), len(s.Points)+len(p.Points)-i)
	copy(out.Points, s.Points)
	for _, pt := range p.Points[i:] {
		out.Points = append(out.Points, Point{Date: pt.Date, Close: last.Close * pt.Close / base})
	}
	out.EstimatedFrom = p.Points[i].Date
	out.EstimateProxy = proxyID
	c.Logf("%s: %d day(s) after the last NAV (%s) estimated from %s",
		s.Symbol, len(p.Points)-i, last.Date.Format("2006-01-02"), proxyID)
	return &out
}

// WithoutEstimates returns the series without its nowcast tail: s itself when
// it carries none, otherwise a copy ending at the last published point. Every
// consumer that stores or validates data reads through it.
func (s *Series) WithoutEstimates() *Series {
	if s == nil || s.EstimatedFrom.IsZero() {
		return s
	}
	out := *s
	out.EstimatedFrom, out.EstimateProxy = time.Time{}, ""
	n := sort.Search(len(s.Points), func(k int) bool { return !s.Points[k].Date.Before(s.EstimatedFrom) })
	out.Points = append([]Point(nil), s.Points[:n]...)
	return &out
}

// nowcastIntraday estimates today's path of a fund from its proxy's intraday
// path: the fund's last daily value before the session (a NAV, or the forward
// nowcast standing on the proxy's previous close) scaled by the proxy's move
// since that close, each tick converted into the fund's currency at the
// intraday cross rate.
func (c *Client) nowcastIntraday(ctx context.Context, id string, e datasets.Asset) (*IntradaySeries, error) {
	proxy, err := c.Intraday(ctx, e.NowcastProxy)
	if err != nil {
		return nil, fmt.Errorf("%s: nowcast proxy %s: %w", id, e.NowcastProxy, err)
	}
	if len(proxy.Points) == 0 {
		return nil, fmt.Errorf("%s: nowcast proxy %s: %w", id, e.NowcastProxy, ErrNotCovered)
	}
	// The session's calendar day, in the exchange's own time zone.
	first := proxy.First().Time
	session := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)
	daily, err := c.Fetch(ctx, id, session.AddDate(0, 0, -30))
	if err != nil {
		return nil, err
	}
	base, on, ok := daily.At(session.AddDate(0, 0, -1))
	if !ok || base <= 0 {
		return nil, fmt.Errorf("%s: no daily value before the %s session", id, session.Format("2006-01-02"))
	}
	// The proxy's close that value stands on, in the fund's currency through
	// the daily conversion, so the anchor and the value agree.
	ref, err := c.FetchExtended(ctx, e.NowcastProxy, FetchOptions{
		From: on.AddDate(0, 0, -14), Currency: daily.Currency, NoSim: true,
	})
	if err != nil {
		return nil, err
	}
	anchor, _, ok := ref.At(on)
	if !ok || anchor <= 0 {
		return nil, fmt.Errorf("%s: no %s close on %s to anchor the estimate", id, e.NowcastProxy, on.Format("2006-01-02"))
	}
	var fx *IntradaySeries
	if proxy.Currency != "" && daily.Currency != "" && proxy.Currency != daily.Currency {
		fx, err = c.Intraday(ctx, proxy.Currency+daily.Currency+"=X")
		if err != nil || len(fx.Points) == 0 {
			return nil, fmt.Errorf("%s: intraday %s→%s rate: %w", id, proxy.Currency, daily.Currency, err)
		}
	}
	out := &IntradaySeries{
		Symbol: id, Name: e.Name + " (estimated from " + e.NowcastProxy + ")",
		Currency: daily.Currency, Source: "nowcast", Estimate: true, Proxy: e.NowcastProxy,
	}
	for _, pt := range proxy.Points {
		v := pt.Close
		if fx != nil {
			rate, ok := fx.rateAt(pt.Time)
			if !ok {
				continue
			}
			v *= rate
		}
		out.Points = append(out.Points, IntradayPoint{Time: pt.Time, Close: base * v / anchor})
	}
	if len(out.Points) == 0 {
		return nil, fmt.Errorf("%s: no %s tick with a %s rate yet: %w", id, e.NowcastProxy, proxy.Currency+daily.Currency, ErrNotCovered)
	}
	return out, nil
}

// rateAt returns the last rate printed at or before t (forward fill), false
// before the first tick.
func (s *IntradaySeries) rateAt(t time.Time) (float64, bool) {
	i := sort.Search(len(s.Points), func(k int) bool { return s.Points[k].Time.After(t) })
	if i == 0 {
		return 0, false
	}
	return s.Points[i-1].Close, true
}
