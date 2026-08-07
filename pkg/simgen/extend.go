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
//     (refdata DEVEXUS-USD: MSCI World ex USA net TR via Curvo, monthly
//     ~1969, carried at daily granularity from 1990-07 by the Ken French
//     daily shape, see dailyShape).
//   - VEIEX (Vanguard Emerging Markets, 1994) → emerging-market equity TR
//     (refdata EM-USD: MSCI Emerging Markets net TR via Curvo, ~1988).
//   - GC=F (COMEX gold futures, 2000) → daily London/LBMA PM gold fix
//     (refdata XAUUSD-LBMA, ~1968).
//   - CL=F (NYMEX WTI futures, 2000) → monthly WTI spot (refdata WTI-USD, ~1946).
//   - VFITX (Intermediate-Term Treasury, 1991) and VUSTX (Long-Term, 1986) →
//     constant-maturity Treasury total-return reconstructions (refdata
//     TREASURY-INT-USD / TREASURY-LONG-USD, from FRED CMT yields, ~1953),
//     carried at daily granularity from 1962 by the TREASURY-*-DAILY
//     daily-yield shapes (see dailyShape).
//   - VFINX (Vanguard 500, 1976) → S&P 500 total return (refdata SP500-USD:
//     month-end levels from the ^SP500TR index 1988->, ^GSPC + Shiller dividend
//     1928-1988, Shiller 1871-1928; see cmd/gen-sp500-refdata), the index VFINX
//     tracks, carried at daily granularity from 1927-12 by the ^GSPC price shape.
//   - ^IRX (13-week T-bill rate) → the 3-month T-bill rate (refdata TBILL-3M:
//     FRED TB3MS, ~1934). A rate, not a price: rescaled by a ≈1 factor at the
//     splice (^IRX ≈ TB3MS there), then read as an isRate series.
//   - GBPUSD=X (Yahoo, ~2003) → the real daily FRED noon GBP/USD rate
//     (refdata GBPUSD-DAILY: DEXUSUK, 1971→), for the GBP-quoted recipes.
//   - DFSVX (DFA US Small Cap Value, 1993) → US small-cap value TR (refdata
//     USSCV-USD: Ken French value-weighted SMALL HiBM daily, cumulated, from
//     1963-07), the size×value factor behind ZPRV/USSC. Real daily total-return
//     levels, so no daily shape is needed.
//   - EUNH.DE (iShares Core Euro Govt Bond, 2009) → euro-area government bond TR
//     (refdata EUROGOV-EUR: OECD euro-area 10y yield through TreasuryTR, ~1970),
//     carried at daily granularity from 2004 by the ECB daily yield-curve shape
//     (EUROGOV-DAILY). The bond leg of the eurozone NTSZ recipe.
var longBack = map[string]string{
	"VTMGX":    "DEVEXUS-USD",
	"VEIEX":    "EM-USD",
	"GC=F":     "XAUUSD-LBMA",
	"CL=F":     "WTI-USD",
	"VFITX":    "TREASURY-INT-USD",
	"VUSTX":    "TREASURY-LONG-USD",
	"VFINX":    "SP500-USD",
	"^IRX":     "TBILL-3M",
	"GBPUSD=X": "GBPUSD-DAILY",
	"DFSVX":    "USSCV-USD",
	"EUNH.DE":  "EUROGOV-EUR",
}

// dailyShape maps a monthly longBack proxy to a daily series of the same
// market whose LEVELS are not authoritative (a close but not identical
// universe, gross of withholding, or a price index without income) but
// whose day-to-day moves are real. The proxy's monthly anchors keep setting
// the levels and the shape supplies the daily granularity in between (see
// shapedSeries), so reconstructions stop feeding month-sized moves to
// daily-frequency statistics. A shape may stop before the anchors' end:
// real fund quotes take over from their inception anyway, and shapedSeries
// keeps the later anchors at their own cadence.
var dailyShape = map[string]string{
	"DEVEXUS-USD":       "DEVEXUS-DAILY",       // Ken French developed-ex-US market TR, daily 1990-07→
	"SP500-USD":         "^GSPC",               // S&P 500 daily price index (Yahoo, 1927-12→)
	"TREASURY-INT-USD":  "TREASURY-INT-DAILY",  // FRED DGS5 daily 5y CMT through TreasuryTR, 1962→1992
	"TREASURY-LONG-USD": "TREASURY-LONG-DAILY", // FRED DGS20 daily 20y CMT through TreasuryTR, 1962→1986
	"WTI-USD":           "WTI-DAILY",           // FRED DCOILWTICO daily WTI spot, 1986→2000
	"EUROGOV-EUR":       "EUROGOV-DAILY",       // ECB daily euro-area 10y yield through TreasuryTR, 2004→
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
	if sid, shaped := dailyShape[pid]; shaped && perr == nil && p != nil {
		if sh, serr := e.inner.Fetch(sid, from); serr == nil {
			p = shapedSeries(p, sh)
		}
	}
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
