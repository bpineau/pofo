package marketdata

import (
	"context"
	"errors"
	"fmt"
)

// FetchAny fetches the first identifier that answers with a usable series,
// tried in order (most authoritative first) - the natural call when one
// instrument is known under several identifiers (ISIN, ticker, name).
//
// With opt.Currency set, the ids are first scanned for a series natively
// quoted in that currency: an ISIN resolution may favor a deeper twin
// listing on another exchange, and the caller's own ticker often is the
// native line. Only when no id answers natively does FetchAny fall back to
// the single-id semantics: ErrWrongCurrency under NoConvert, conversion of
// the most authoritative answer otherwise. The client's per-run
// memoization makes the second pass cheap.
//
// When every id fails, the errors are joined so no cause is masked.
func (c *Client) FetchAny(ctx context.Context, ids []string, opt FetchOptions) (*Series, error) {
	if len(ids) == 0 {
		return nil, errors.New("FetchAny: no identifier")
	}
	if opt.Currency == "" {
		return c.fetchFirst(ctx, ids, opt)
	}
	native := opt
	native.NoConvert = true
	s, nativeErr := c.fetchFirst(ctx, ids, native)
	if nativeErr == nil {
		return s, nil
	}
	if opt.NoConvert {
		return nil, nativeErr
	}
	return c.fetchFirst(ctx, ids, opt)
}

// fetchFirst returns the first id FetchExtended serves with at least one
// point, joining every failure otherwise.
func (c *Client) fetchFirst(ctx context.Context, ids []string, opt FetchOptions) (*Series, error) {
	var errs []error
	for _, id := range ids {
		s, err := c.FetchExtended(ctx, id, opt)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if len(s.Points) == 0 {
			errs = append(errs, fmt.Errorf("%s: empty series", id))
			continue
		}
		return s, nil
	}
	return nil, errors.Join(errs...)
}
