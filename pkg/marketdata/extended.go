package marketdata

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// DefaultCacheDir returns the standard per-user cache directory for pofo
// downloads ("~/Library/Caches/pofo" on macOS, "~/.cache/pofo" on Linux),
// the same directory the pofo CLI uses, so a library consumer and the CLI
// share their cache. It falls back to "data" in the current directory when
// the user cache directory cannot be determined.
func DefaultCacheDir() string {
	if c, err := os.UserCacheDir(); err == nil {
		return filepath.Join(c, "pofo")
	}
	return "data"
}

// FetchOptions parameterizes Client.FetchExtended. The zero value keeps
// every default: full available history, SIM suffixes honored against the
// simulated histories embedded in the binary, no currency conversion.
type FetchOptions struct {
	// From is the history depth requested from the sources (the zero time
	// requests everything). The returned series may still start later, when
	// quotes only begin later, or earlier, when a deeper download is
	// already cached.
	From time.Time

	// To clamps the end of the returned series; the zero time keeps every
	// point through today.
	To time.Time

	// NoSim disables the SIM-suffix convention: "VOOSIM" is then fetched
	// as "VOO", real quotes only, with no history extension.
	NoSim bool

	// Simdata is the source of the simulated histories used to extend a
	// SIM asset. Nil uses the series embedded in the binary
	// (datasets.Simdata); use os.DirFS for a development directory.
	Simdata fs.FS

	// Currency converts the series into this quote currency using daily FX
	// crosses (see Client.ConvertCurrency); empty keeps the native
	// currency. A series whose currency is unknown passes through
	// unchanged.
	Currency string

	// Raw asks for unadjusted closes: split-adjusted but with dividends
	// NOT reinvested (the price actually traded), instead of the default
	// adjusted (total-return) closes. Pair it with Series.Dividends to
	// account for income separately. Raw combined with a SIM suffix is an
	// error: simulated histories are total-return by construction.
	Raw bool
}

// FetchExtended fetches an asset the way the pofo CLI does: Fetch, then for
// "…SIM" identifiers (see SplitSim) the history extension, splicing the
// bundled simulated series or else a known long-history proxy (ProxySymbol)
// in front of the real quotes, then the optional conversion into a target
// currency. Real quotes always take precedence wherever they exist;
// Series.SimulatedBefore marks the frontier. Progress and degradations
// (unreadable simdata, unavailable proxy, FX rates held flat before their
// history) are reported through Client.Logf.
//
// Plain identifiers skip the extension: FetchExtended("VOO", …) behaves
// like Fetch plus the window and currency handling, so it is safe to route
// every asset of a portfolio through this single entry point.
func (c *Client) FetchExtended(ctx context.Context, id string, opt FetchOptions) (*Series, error) {
	base, wantSim := SplitSim(id)
	if opt.NoSim {
		wantSim = false
	}
	if opt.Raw && wantSim {
		return nil, fmt.Errorf("%s: raw closes cannot be SIM-extended (simulated histories are total-return); set NoSim or drop Raw", id)
	}
	if !wantSim {
		s, err := c.fetch(ctx, base, opt.From, opt.Raw)
		if err != nil {
			return nil, err
		}
		if s, err = c.convertTo(ctx, s, opt.Currency, opt.From); err != nil {
			return nil, err
		}
		return Trim(s, time.Time{}, opt.To), nil
	}

	simdata := opt.Simdata
	if simdata == nil {
		simdata = datasets.Simdata()
	}
	canonical := CanonicalID(base)
	sim, simOK, simErr := ReadSimdataFS(simdata, canonical)
	if simErr != nil {
		c.Logf("warning: simdata %s unreadable: %v", canonical, simErr)
	}
	if simOK {
		sim = Trim(sim, opt.From, time.Time{})
		simOK = len(sim.Points) >= 2
	}
	s, err := c.Fetch(ctx, base, opt.From)
	if err != nil {
		if simOK {
			c.Logf("warning: %s unavailable (%v), using simulated data only", base, err)
			sim.SimulatedBefore = sim.Last().Date
			sim.ProxySymbol = "simdata"
			return Trim(sim, time.Time{}, opt.To), nil
		}
		return nil, err
	}
	// The client memoizes series by symbol: work on a copy so that the
	// extension never leaks into the bare (real-data-only) variant of the
	// same asset. ExtendBack allocates a fresh Points slice.
	cp := *s
	s = &cp
	if simOK && ExtendBack(s, sim) {
		s.ProxySymbol = "simdata"
		c.Logf("%s: history extended via simdata starting %s",
			canonical, s.First().Date.Format("2006-01-02"))
	}
	// Only bother with a proxy when at least six months are missing.
	if s.SimulatedBefore.IsZero() && s.First().Date.After(opt.From.AddDate(0, 6, 0)) {
		proxySym, ok := ProxySymbol(canonical)
		if !ok {
			proxySym, ok = ProxySymbol(s.Symbol)
		}
		if ok {
			ps, perr := c.History(ctx, proxySym, opt.From)
			if perr != nil {
				c.Logf("warning: proxy %s for %s unavailable: %v", proxySym, s.Symbol, perr)
			} else if ExtendBack(s, ps) {
				c.Logf("%s: history extended via %s starting %s",
					s.Symbol, proxySym, s.First().Date.Format("2006-01-02"))
			}
		}
	}
	if s, err = c.convertTo(ctx, s, opt.Currency, opt.From); err != nil {
		return nil, err
	}
	return Trim(s, time.Time{}, opt.To), nil
}

// convertTo reprices s into currency via ConvertCurrency. It is a no-op
// when currency is empty, already the series' own, or when the series does
// not report one (the caller may warn about the mix); FX rates missing at
// the start of the window are held flat with a Logf warning.
func (c *Client) convertTo(ctx context.Context, s *Series, currency string, from time.Time) (*Series, error) {
	if currency == "" || s.Currency == "" || s.Currency == currency {
		return s, nil
	}
	out, extrapolated, err := c.ConvertCurrency(ctx, s, currency, from)
	if err != nil {
		return nil, err
	}
	if !extrapolated.IsZero() {
		c.Logf("warning: %s: FX rate %s→%s unavailable before %s, held constant earlier",
			s.Symbol, s.Currency, currency, extrapolated.Format("2006-01-02"))
	}
	return out, nil
}

// Trim returns s restricted to [from, to]; a zero bound is open on that
// side. Dividends are clipped to the same window. When the series already
// fits the window it is returned as is, otherwise the result is a copy
// with fresh slices and the same metadata.
func Trim(s *Series, from, to time.Time) *Series {
	if len(s.Points) == 0 ||
		((from.IsZero() || !s.First().Date.Before(from)) &&
			(to.IsZero() || !s.Last().Date.After(to))) {
		return s
	}
	inWindow := func(d time.Time) bool {
		return (from.IsZero() || !d.Before(from)) && (to.IsZero() || !d.After(to))
	}
	out := *s
	out.Points = nil
	for _, p := range s.Points {
		if inWindow(p.Date) {
			out.Points = append(out.Points, p)
		}
	}
	out.Dividends = nil
	for _, d := range s.Dividends {
		if inWindow(d.Date) {
			out.Dividends = append(out.Dividends, d)
		}
	}
	return &out
}
