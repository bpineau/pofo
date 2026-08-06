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
// available for it. Every proxy is a bundled refdata series (go:embed via
// datasets.Refdata), so the splices work offline and reproducibly; Yahoo's
// long MSCI index symbols (^990300-USD-STRD etc.) return nothing to the client
// and are deliberately not used. All proxies are total-return / spot levels in
// USD, homogeneous with the component they extend:
//
//   - VTMGX (Vanguard Developed Markets, 1999) → developed-ex-US equity TR
//     (refdata DEVEXUS-USD: Ken French developed-ex-US from 1990, MSCI World
//     before, ~1969).
//   - VEIEX (Vanguard Emerging Markets, 1994) → emerging-market equity TR
//     (refdata EM-USD: Ken French emerging, ~1989).
//   - GC=F (COMEX gold futures, 2000) → monthly London/LBMA gold fix
//     (refdata XAUUSD-LBMA, ~1968).
//   - CL=F (NYMEX WTI futures, 2000) → monthly WTI spot (refdata WTI-USD, ~1946).
//   - VFITX (Intermediate-Term Treasury, 1991) and VUSTX (Long-Term, 1986) →
//     constant-maturity Treasury total-return reconstructions (refdata
//     TREASURY-INT-USD / TREASURY-LONG-USD, from FRED CMT yields, ~1953).
//   - VFINX (Vanguard 500, 1976) → US equity TR (refdata USEQ-USD: Ken French
//     total US market, ~1926).
//   - ^IRX (13-week T-bill rate) → the 3-month T-bill rate (refdata TBILL-3M:
//     FRED TB3MS, ~1934). A rate, not a price: rescaled by a ≈1 factor at the
//     splice (^IRX ≈ TB3MS there), then read as an isRate series.
var longBack = map[string]string{
	"VTMGX": "DEVEXUS-USD",
	"VEIEX": "EM-USD",
	"GC=F":  "XAUUSD-LBMA",
	"CL=F":  "WTI-USD",
	"VFITX": "TREASURY-INT-USD",
	"VUSTX": "TREASURY-LONG-USD",
	"VFINX": "USEQ-USD",
	"^IRX":  "TBILL-3M",
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
