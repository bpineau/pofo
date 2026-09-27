package marketdata

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
)

// Load returns the series of id from the first place that holds it, the
// lookup an exploration program wants: the data the module ships before the
// network, and a file of one's own named like any identifier.
//
//  1. A FILE, when id is a path (it holds a path separator or ends in
//     ".csv"): ReadCSV, labelled with the path as written. This is how an
//     older version of a bundled file joins in, piped out of "git show".
//  2. The BUNDLE (Bundled), unless id carries the SIM suffix: a catalog
//     asset's backcast, its whole reconstructed history as of the last data
//     refresh, or a reference series (an index, a yield, a cash rate).
//  3. The CLIENT, Client.FetchExtended with opt: live quotes cached on disk,
//     the backcast spliced in front of a SIM identifier. With c.Offline set,
//     only the quote cache and the bundle answer, and anything else fails
//     with an error wrapping ErrOffline.
//
// So "IWDA" reads the fund's bundled backcast, offline and the same on every
// machine, while "IWDASIM" reads its live quotes with that backcast in front,
// and "VOO", bundled under no name, is fetched.
//
// opt applies to every step: From and To trim the series, and Currency
// converts it through Client.ConvertCurrency (the euro crosses are bundled,
// so converting to or from the euro works offline). A bundled reference
// series or a file whose header states no currency cannot be converted, and
// asking is an error rather than a silent pass-through; NoConvert with a
// Currency that is not the series' own is ErrWrongCurrency, as for a fetch.
// Raw asks for unadjusted closes, which no bundled series holds (they are
// total returns), so Raw skips the bundle.
//
// A fetched series comes back without its nowcast (WithoutEstimates): an
// estimate of a fund's last published value is not data to explore. An empty
// window is an error that names id.
func (c *Client) Load(ctx context.Context, id string, opt FetchOptions) (*Series, error) {
	s, err := c.load(ctx, id, opt)
	if err != nil {
		return nil, err
	}
	if s.Len() == 0 {
		return nil, fmt.Errorf("%s: no point between %s and %s", id, dateOrOpen(opt.From), dateOrOpen(opt.To))
	}
	return s, nil
}

// load is Load before the empty-window check: the three doors in order.
func (c *Client) load(ctx context.Context, id string, opt FetchOptions) (*Series, error) {
	if isPath(id) {
		f, err := os.Open(id)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		s, err := ReadCSV(f, id)
		if err != nil {
			return nil, err
		}
		return c.localTo(ctx, s, "the file's header", opt)
	}
	if _, sim := SplitSim(id); !sim && !opt.Raw {
		s, err := Bundled(id)
		switch {
		case err == nil:
			return c.localTo(ctx, s, "a bundled reference series", opt)
		case !errors.Is(err, fs.ErrNotExist):
			return nil, err
		}
	}
	s, err := c.FetchExtended(ctx, id, opt)
	if err != nil {
		return nil, err
	}
	return Trim(s.WithoutEstimates(), opt.From, opt.To), nil
}

// localTo applies opt's window and currency to a series read from the
// bundle or a file, which no source fetched and FetchExtended never shaped.
// origin names where a missing currency should have been stated, for the
// error.
func (c *Client) localTo(ctx context.Context, s *Series, origin string, opt FetchOptions) (*Series, error) {
	s = Trim(s, opt.From, opt.To)
	if opt.Currency == "" || s.Currency == opt.Currency || s.Len() == 0 {
		return s, nil
	}
	if s.Currency == "" {
		return nil, fmt.Errorf("%s: %s states no currency to convert into %s from; drop the conversion",
			s.Symbol, origin, opt.Currency)
	}
	if opt.NoConvert {
		return nil, fmt.Errorf("%s is quoted in %s, not %s: %w", s.Symbol, s.Currency, opt.Currency, ErrWrongCurrency)
	}
	return c.convertTo(ctx, s, opt.Currency, s.First().Date)
}

// isPath reports whether Load reads id as a file rather than an identifier:
// no identifier holds a path separator or ends in ".csv".
func isPath(id string) bool {
	return strings.ContainsRune(id, '/') || strings.ContainsRune(id, os.PathSeparator) ||
		strings.HasSuffix(strings.ToLower(id), ".csv")
}

// dateOrOpen formats a window bound for an error, the zero time as an open
// bound.
func dateOrOpen(t time.Time) string {
	if t.IsZero() {
		return "(open)"
	}
	return t.Format(time.DateOnly)
}
