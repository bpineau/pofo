package simgen

import (
	"fmt"
	"os"
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
	// Diagnostics to stderr (gen-simdata only): a silent skip hides why a
	// backcast did not lengthen, so report the proxy fetch and splice outcome.
	p, perr := e.inner.Fetch(pid, from)
	switch {
	case perr != nil:
		fmt.Fprintf(os.Stderr, "extend: %s: proxy %s fetch failed: %v\n", id, pid, perr)
	case p == nil || len(p.Points) < 2:
		fmt.Fprintf(os.Stderr, "extend: %s: proxy %s returned no usable history\n", id, pid)
	case marketdata.ExtendBack(s, p):
		fmt.Fprintf(os.Stderr, "extend: %s extended with %s back to %s\n", id, pid, s.Points[0].Date.Format("2006-01-02"))
	default:
		fmt.Fprintf(os.Stderr, "extend: %s: proxy %s (from %s) added no earlier data than %s\n",
			id, pid, p.Points[0].Date.Format("2006-01-02"), s.Points[0].Date.Format("2006-01-02"))
	}
	return s, nil
}

// extend wraps a fetcher with the standard long-history proxy splices. The
// shared composite and tsmom builders route their component fetches through it,
// so every reconstruction using an extendable component reaches further back.
func extend(f Fetcher) Fetcher { return extendingFetcher{inner: f, back: longBack} }
