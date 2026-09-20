package marketdata

import "strings"

// Some venues quote a price in a SUB-UNIT of their currency, and the provider
// reports that sub-unit where a currency code is expected: Yahoo answers "GBp"
// for a London listing (pence), "ZAc" for Johannesburg (cents), "USX" for the
// Chicago grain contracts (US cents); the Financial Times spells the London one
// "GBX". The number and the code then disagree by a hundredfold, and nothing
// downstream can see it: "GBp" and "GBP" are the same string to every
// case-insensitive comparison this package makes, so a pence price walks
// straight through the native-currency gates (fetchSpec.currencyOK,
// LatestAny) and is booked as pounds. That is a 100x valuation error on every
// London line, which is why the sub-unit is removed at the source instead of
// being carried and remembered.
//
// normalizeUnits (and its Quote / IntradaySeries siblings) is therefore applied
// wherever a provider's numbers enter the package, and once more when a cache
// file written before this existed is read back, so no consumer ever sees a
// sub-unit code. The rescaling is exact (a power of ten) and idempotent: a
// series already in major units carries a code the table does not hold.
//
// The table is deliberately keyed on the EXACT spellings observed, never
// folded to one case: "GBp" (pence) and "GBP" (pounds) differ by case alone,
// so a case-insensitive lookup would divide every pound price by a hundred.
var minorUnits = map[string]struct {
	currency string  // the major ISO code the sub-unit belongs to
	factor   float64 // major units per sub-unit
}{
	"GBp": {"GBP", 0.01}, // Yahoo, London pence
	"GBX": {"GBP", 0.01}, // Financial Times, same pence
	"GBx": {"GBP", 0.01},
	"ZAc": {"ZAR", 0.01}, // Yahoo, Johannesburg cents
	"ZAC": {"ZAR", 0.01},
	"ILA": {"ILS", 0.01}, // Tel Aviv agorot
	"USX": {"USD", 0.01}, // US cents (CBOT grains, ICE softs)
}

// majorUnit maps a quote currency code to the ISO currency it really measures
// and the factor that turns one quoted unit into one of those. minor is true
// only for a sub-unit code: every ordinary currency returns itself and 1, so
// callers can apply the result unconditionally.
func majorUnit(code string) (currency string, factor float64, minor bool) {
	if u, ok := minorUnits[code]; ok {
		return u.currency, u.factor, true
	}
	return code, 1, false
}

// IsMinorUnit reports whether a currency code is a venue sub-unit (pence,
// cents, agorot) rather than an ISO currency. Nothing this package serves ever
// carries one: it exists so a caller validating its own records, or a doctor
// reading a bundled file, can name the mistake rather than divide by a hundred
// twice.
func IsMinorUnit(code string) bool {
	_, ok := minorUnits[code]
	return ok
}

// mustMajor is majorUnit's currency alone, for a message that has already
// established the code is a sub-unit.
func mustMajor(code string) string {
	currency, _, _ := majorUnit(code)
	return currency
}

// sameCurrency reports whether two quote-currency codes name the same money,
// ignoring case and the venue sub-unit: "GBp", "GBX" and "gbp" all answer for
// "GBP". Every currency comparison in this package goes through it rather than
// through a bare EqualFold, which folds pence into pounds by the accident that
// the two codes differ by case alone. It is a question about the LABEL: the
// numbers are made to agree with it by normalizeUnits, at the provider
// boundary, never here.
func sameCurrency(a, b string) bool {
	a, _, _ = majorUnit(strings.TrimSpace(a))
	b, _, _ = majorUnit(strings.TrimSpace(b))
	return strings.EqualFold(a, b)
}

// normalizeUnits rescales a series quoted in a sub-unit into the major
// currency, dividends included, and relabels it. A series in an ordinary
// currency, or an empty one, is left untouched. It mutates in place: every
// caller owns the series it has just parsed or loaded.
func normalizeUnits(s *Series) {
	if s == nil {
		return
	}
	currency, factor, minor := majorUnit(s.Currency)
	if !minor {
		return
	}
	s.Currency = currency
	for i := range s.Points {
		s.Points[i].Close *= factor
	}
	for i := range s.Dividends {
		s.Dividends[i].Amount *= factor
	}
}

// normalizeQuoteUnits is normalizeUnits for a single quote.
func normalizeQuoteUnits(q *Quote) {
	if q == nil {
		return
	}
	if currency, factor, minor := majorUnit(q.Currency); minor {
		q.Currency, q.Price = currency, q.Price*factor
	}
}

// normalizeIntradayUnits is normalizeUnits for an intraday path.
func normalizeIntradayUnits(s *IntradaySeries) {
	if s == nil {
		return
	}
	currency, factor, minor := majorUnit(s.Currency)
	if !minor {
		return
	}
	s.Currency = currency
	for i := range s.Points {
		s.Points[i].Close *= factor
	}
}
