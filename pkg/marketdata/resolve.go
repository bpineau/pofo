package marketdata

import (
	"context"
	"fmt"
	"time"
)

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

func toResolution(r resolution) Resolution { return Resolution(r) }

// resolveFrom is the history depth Resolve fetches over when it must run a full
// resolution: deep enough that the multi-source search settles on the same
// instrument it would for a real long-horizon request.
func resolveFrom() time.Time { return time.Now().AddDate(-15, 0, 0) }

// Search returns candidate instruments for a free-text query: a name
// ("MSCI World"), a ticker or an ISIN. When the query canonicalizes to a
// catalog-pinned asset, that resolution comes first (deterministic and
// vetted); the Yahoo search candidates follow, deduplicated by symbol. No
// price series is downloaded: pass a candidate's Symbol to Fetch (or the
// original query to Resolve) to quote it. It errors when nothing at all
// matches.
func (c *Client) Search(ctx context.Context, query string) ([]Resolution, error) {
	var out []Resolution
	seen := map[string]bool{}
	add := func(r Resolution) {
		key := r.Source + "|" + r.Symbol + "|" + r.Xid
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, r)
	}
	if res, ok := catalogResolution(CanonicalID(query)); ok {
		add(toResolution(res))
	}
	quotes, err := c.search(ctx, query)
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("search %q: %w", query, err)
	}
	for _, q := range quotes {
		add(Resolution{Source: "yahoo", Symbol: q.Symbol, Name: q.Name})
	}
	return out, nil
}

// Resolve returns the instrument pofo would quote for a user identifier
// (ticker, ISIN or alias). It uses the bundled catalog and the on-disk
// resolution cache first, then the same multi-source search Fetch uses. It may
// perform network I/O and caches the result, so a later Fetch of the same id
// reuses this work.
func (c *Client) Resolve(ctx context.Context, id string) (Resolution, error) {
	canonical := CanonicalID(id)
	if res, ok := c.loadResolution(canonical); ok {
		return toResolution(res), nil
	}
	s, err := c.Fetch(ctx, id, resolveFrom())
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
