package simgen

import (
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// longBack maps a short-history component to a longer real proxy whose
// (rescaled) history is spliced in before the component's own inception, so a
// reconstruction covers each period with the most reliable real series
// available for it. The proxies are authoritative long index series:
//
//   - VTMGX (Vanguard Developed Markets, 1999) is extended with the MSCI EAFE
//     gross total-return index (Yahoo ^990300-USD-STRD, ~1970), the standard
//     long developed-ex-US series.
//   - GC=F (COMEX gold futures, 2000) is extended with XAU/USD spot (~1968).
//
// Both proxies are total-return / spot levels in USD, so the splice is
// homogeneous with the component it extends.
var longBack = map[string]string{
	"VTMGX": "^990300-USD-STRD",
	"GC=F":  "XAUUSD",
}

// extendingFetcher wraps a Fetcher so that a configured component is spliced
// with a longer proxy (marketdata.ExtendBack) at fetch time. A missing or empty
// proxy is skipped silently, leaving the component unchanged, so the wrapper is
// safe to apply unconditionally.
type extendingFetcher struct {
	inner Fetcher
	back  map[string]string
}

// Fetch fetches id and, when a longer proxy is configured for it, prepends the
// proxy's rescaled history before the component's first quote.
func (e extendingFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	s, err := e.inner.Fetch(id, from)
	if err != nil {
		return nil, err
	}
	pid, ok := e.back[id]
	if !ok || s == nil {
		return s, err
	}
	if p, perr := e.inner.Fetch(pid, from); perr == nil && p != nil && len(p.Points) > 1 {
		marketdata.ExtendBack(s, p)
	}
	return s, nil
}

// extend wraps a fetcher with the standard long-history proxy splices. The
// shared composite and tsmom builders route their component fetches through it,
// so every reconstruction using an extendable component reaches further back.
func extend(f Fetcher) Fetcher { return extendingFetcher{inner: f, back: longBack} }
