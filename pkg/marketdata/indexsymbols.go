package marketdata

import (
	"fmt"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Index symbols ("^GSPC", "^BCOM", "^IRX") are Yahoo's names for an index or a
// rate, not instruments anyone can buy, and two rules follow from that.
//
// First, an index symbol is resolved by its SYMBOL only. When its direct quote
// fails, the search fallback may keep a listing of that very symbol and
// nothing else: the full-text searches behind it return a best match for any
// string, and a fund whose name carries the index's short name ("... BCOM
// ...") is exactly the unrelated instrument the fuzzy gate exists to refuse. A
// cached resolution to anything but that symbol is ignored for the same
// reason. See fetchTicker.
//
// Second, an index symbol whose feed was WITHDRAWN is served from the bundled
// reference that replaces it, below, so the name a user types keeps answering
// with the index itself rather than with a search. The reference is the
// bundle's own validated series (pkg/datasets/refdata), read without touching
// the network, in the currency the index is published in.
var withdrawnIndexSymbols = map[string]withdrawnIndex{
	// Yahoo withdrew the whole Bloomberg Commodity family in 2026-09 (HTTP
	// 404 on every symbol form); BCOM-ER-USD is the same excess-return index
	// from the Financial Times' copy, cmd/gen-bcom-refdata.
	"^BCOM": {ref: "BCOM-ER-USD", currency: "USD", name: "Bloomberg Commodity Index (excess return)"},
}

// withdrawnIndex is the bundled stand-in for an index symbol whose live feed
// is gone: the refdata file that carries it, its currency and its name.
type withdrawnIndex struct {
	ref, currency, name string
}

// isIndexSymbol reports whether an identifier is an index symbol ("^...").
func isIndexSymbol(id string) bool { return strings.HasPrefix(id, "^") }

// isWithdrawnIndex reports whether symbol is served from a bundled reference.
func isWithdrawnIndex(symbol string) bool {
	_, ok := withdrawnIndexSymbols[symbol]
	return ok
}

// fetchWithdrawnIndex serves an index symbol from its bundled reference, under
// the symbol that was asked for, trimmed to from. No network is touched and no
// cache written: the series is the one the binary was built with.
func (c *Client) fetchWithdrawnIndex(symbol string, from time.Time, spec fetchSpec) (*Series, error) {
	w := withdrawnIndexSymbols[symbol]
	s, ok, err := ReadSimdataFS(datasets.Refdata(), w.ref)
	if err != nil {
		return nil, fmt.Errorf("%s: bundled %s: %w", symbol, w.ref, err)
	}
	if !ok || len(s.Points) == 0 {
		return nil, fmt.Errorf("%s: bundled %s is missing", symbol, w.ref)
	}
	if !spec.currencyOK(w.currency) {
		return nil, fmt.Errorf("%s: quotes in %s, want %s: %w", symbol, w.currency, spec.wantCurrency, ErrWrongCurrency)
	}
	c.Logf("%s: Yahoo no longer serves it, reading the bundled %s", symbol, w.ref)
	s.Symbol, s.Name, s.Source, s.Currency = symbol, w.name, "refdata", w.currency
	return Trim(s, from, time.Time{}), nil
}
