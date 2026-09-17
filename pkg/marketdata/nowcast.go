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
//
// THE ANCHOR is that timing made explicit. The estimate is the last published
// NAV times the proxy's move SINCE THE PRINT THAT NAV WAS STRUCK ON, which the
// record names: the proxy's close of the NAV's day by default, its OPEN when
// the catalog says nowcast_anchor "open" (a fund whose valuation rules price
// its holding at the opening of the valuation day, as a single-stock FCPE
// does). Anchoring such a fund on the close would carry that session's
// open-to-close move as an offset for as long as the NAV is the last one:
// measured on the 264 NAV spans of ERES_DATADOG since 2022-04-11, the
// one-span-ahead estimate lands at 3.8 % rmse anchored on the close against
// 2.8 % anchored on the open. The open is a best effort: a day the proxy did
// not trade, a source with no opening print or a failed fetch falls back on
// the close silently, never on an error.

// nowcastForward appends to s the business days after its last point that the
// proxy has closed on, each carrying the proxy's return in s's currency since
// the print the last published value was struck on (e.NowcastAnchor). It
// returns s untouched when there is nothing to add or the proxy cannot be
// read, and a copy otherwise: the memoized series must stay real.
func (c *Client) nowcastForward(ctx context.Context, s *Series, e datasets.Asset) *Series {
	proxyID := e.NowcastProxy
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
	base, on, ok := p.At(last.Date)
	if !ok || base <= 0 {
		return s
	}
	// The published value was struck on the proxy's open of its own day, so
	// the anchor moves back there; a forward-filled close (the proxy did not
	// trade that day) has no open of that day to move to.
	if e.NowcastAnchor == datasets.NowcastAnchorOpen && on.Equal(last.Date) {
		if f, ok := c.openAnchorFactor(ctx, proxyID, last.Date); ok {
			base *= f
		}
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
	anchor, refOn, ok := ref.At(on)
	if !ok || anchor <= 0 {
		return nil, fmt.Errorf("%s: no %s close on %s to anchor the estimate", id, e.NowcastProxy, on.Format("2006-01-02"))
	}
	// A PUBLISHED value of a fund struck at the opening print stands on that
	// open, not on its day's close; a forward nowcast day already stands on
	// the proxy's close of the same day, so it anchors there whatever the
	// record says.
	published := daily.EstimatedFrom.IsZero() || on.Before(daily.EstimatedFrom)
	if e.NowcastAnchor == datasets.NowcastAnchorOpen && published && refOn.Equal(on) {
		if f, ok := c.openAnchorFactor(ctx, e.NowcastProxy, on); ok {
			anchor *= f
		}
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

// openAnchorFactorView is the cache and memoization identity of a proxy's
// open-to-close factor series, alongside the "~raw" view of viewKey: the
// factors are a series of their own, so no consumer of the ordinary close
// series and no cache file written before they existed is touched.
func openAnchorFactorView(symbol string) string { return symbol + "~open" }

// openAnchorFactor returns the factor that moves a value standing on the
// proxy's CLOSE of day to the same value standing on its OPEN of that session:
// the proxy's open divided by its close, currency-independent (one session
// carries one exchange rate, which the ratio cancels).
//
// ok is false whenever the factor cannot be established for that exact day (a
// proxy with no Yahoo symbol, a day it did not trade, a fetch failure), so the
// caller keeps its close-anchored estimate rather than failing.
func (c *Client) openAnchorFactor(ctx context.Context, proxyID string, day time.Time) (float64, bool) {
	symbol, ok := c.yahooSymbol(ctx, proxyID)
	if !ok {
		c.Logf("warning: nowcast proxy %s has no Yahoo symbol, the estimate stays anchored on the close", proxyID)
		return 1, false
	}
	from := day.AddDate(0, 0, -14)
	s, err := c.cachedHistory(ctx, "Yahoo opens", openAnchorFactorView(symbol), from, false, func() (*Series, error) {
		return c.fetchYahooOpenFactors(ctx, symbol, from)
	})
	if err != nil {
		c.Logf("warning: %s: no opening prices (%v), the estimate stays anchored on the close", proxyID, err)
		return 1, false
	}
	i := sort.Search(len(s.Points), func(k int) bool { return !s.Points[k].Date.Before(day) })
	if i >= len(s.Points) || !s.Points[i].Date.Equal(day) || s.Points[i].Close <= 0 {
		c.Logf("warning: %s: no opening price on %s, the estimate stays anchored on the close",
			proxyID, day.Format("2006-01-02"))
		return 1, false
	}
	return s.Points[i].Close, true
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
