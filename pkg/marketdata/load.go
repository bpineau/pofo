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
//  2. The BUNDLE (Bundled), for a bare identifier: a reference series (an
//     index, a yield, a cash rate) or a catalog index (SP500, MSCIWORLD),
//     offline and the same on every machine.
//  3. The CLIENT, Client.FetchExtended with opt: live quotes cached on disk,
//     the bundled reconstruction spliced in front of a SIM identifier. With
//     c.Offline set, only the quote cache and the bundle answer (a SIM
//     identifier never fetched falls back on its bundled SIM history alone),
//     and anything else fails with an error wrapping ErrOffline.
//
// So Load keeps the SIM convention: "IWDA" reads the fund's real quotes
// only, "IWDASIM" those quotes with the reconstruction in front, and
// "TREASURY-LONG-USD", a reference series no source quotes, the bundle. A
// quoted fund's reconstruction never answers to its bare name, here or in
// Bundled.
//
// opt applies to every step: From and To trim the series, and Currency
// converts it through Client.ConvertCurrency (the euro crosses are bundled,
// so converting to or from the euro works offline). A series whose currency
// is unknown passes through unchanged, as FetchExtended passes one: a bundled
// reference series states none (its identifier usually names it, as
// EURCASH-EUR does), nor may a file's header, and Client.Logf says which
// series was taken as it is. NoConvert with a Currency that is not the
// series' own is ErrWrongCurrency, as for a fetch.
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
			return c.localTo(ctx, s, "the bundled file", opt)
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
// origin names where a missing currency would have been stated, for the log.
func (c *Client) localTo(ctx context.Context, s *Series, origin string, opt FetchOptions) (*Series, error) {
	s = Trim(s, opt.From, opt.To)
	if opt.Currency == "" || s.Currency == opt.Currency || s.Len() == 0 {
		return s, nil
	}
	if s.Currency == "" {
		c.Logf("%s: %s states no currency, taken as it is rather than converted into %s", s.Symbol, origin, opt.Currency)
		return s, nil
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
