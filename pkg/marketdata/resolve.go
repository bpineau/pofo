package marketdata

import "time"

// Resolution is how pofo maps a user identifier to a quotable instrument:
// which source serves it, under which symbol, plus the resolved name and quote
// currency. Currency may be empty when a source does not report it.
type Resolution struct {
	Source   string // "yahoo", "stooq", "ft" or "morningstar"
	Symbol   string // Yahoo or Stooq symbol, or Morningstar id; empty for ft
	Xid      string // FT internal id; empty otherwise
	Name     string
	Currency string
}

func toResolution(r resolution) Resolution {
	return Resolution{Source: r.Source, Symbol: r.Symbol, Xid: r.Xid, Name: r.Name, Currency: r.Currency}
}

// resolveFrom is the history depth Resolve fetches over when it must run a full
// resolution: deep enough that the multi-source search settles on the same
// instrument it would for a real long-horizon request.
func resolveFrom() time.Time { return time.Now().AddDate(-15, 0, 0) }

// Resolve returns the instrument pofo would quote for a user identifier
// (ticker, ISIN or alias). It uses the bundled catalog and the on-disk
// resolution cache first, then the same multi-source search Fetch uses. It may
// perform network I/O and caches the result, so a later Fetch of the same id
// reuses this work.
func (c *Client) Resolve(id string) (Resolution, error) {
	canonical := CanonicalID(id)
	if res, ok := c.loadResolution(canonical); ok {
		return toResolution(res), nil
	}
	s, err := c.Fetch(id, resolveFrom())
	if err != nil {
		return Resolution{}, err
	}
	// An ISIN or fund path adopts a resolution Fetch can now load back.
	if res, ok := c.loadResolution(canonical); ok {
		r := toResolution(res)
		if r.Currency == "" {
			r.Currency = s.Currency
		}
		if r.Name == "" {
			r.Name = s.Name
		}
		return r, nil
	}
	// A direct ticker keeps no resolution file: its identity is the series.
	return Resolution{Source: s.Source, Symbol: s.Symbol, Name: s.Name, Currency: s.Currency}, nil
}
