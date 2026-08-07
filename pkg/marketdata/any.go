package marketdata

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// QuoteOptions constrains LatestAny. The zero value keeps every default:
// the first identifier that answers wins, in its native currency.
type QuoteOptions struct {
	// Currency demands the quote in this ISO 4217 currency: identifiers
	// answering natively in it win; otherwise the most authoritative
	// answer is converted through FXRate at the quote's own timestamp.
	Currency string
	// NoConvert, with Currency set, fails with ErrWrongCurrency instead
	// of converting an off-currency quote.
	NoConvert bool
}

// LatestAny returns the freshest price for the first identifier that
// answers, tried in order (most authoritative first). See FetchAny for
// the native-first currency contract; unlike a series, an off-currency
// quote converts at its own timestamp, which suits spot valuations (the
// next real close overwrites the point). When every id fails, the errors
// are joined so no cause is masked.
func (c *Client) LatestAny(ctx context.Context, ids []string, opt QuoteOptions) (*Quote, error) {
	if len(ids) == 0 {
		return nil, errors.New("LatestAny: no identifier")
	}
	var errs []error
	var offCurrency *Quote
	for _, id := range ids {
		q, err := c.Latest(ctx, id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if opt.Currency == "" || q.Currency == "" || strings.EqualFold(q.Currency, opt.Currency) {
			return q, nil
		}
		if offCurrency == nil {
			offCurrency = q
		}
		errs = append(errs, fmt.Errorf("%s: %w: got %s, want %s",
			id, ErrWrongCurrency, q.Currency, opt.Currency))
	}
	if offCurrency != nil && !opt.NoConvert {
		rate, err := c.FXRate(ctx, offCurrency.Currency, opt.Currency, offCurrency.Time)
		if err != nil {
			return nil, errors.Join(append(errs, err)...)
		}
		q := *offCurrency
		q.Price *= rate
		q.Currency = opt.Currency
		return &q, nil
	}
	return nil, errors.Join(errs...)
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
