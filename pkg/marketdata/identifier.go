package marketdata

import (
	"regexp"
	"strings"
)

// maxTickerLen bounds an exchange ticker: the longest real ones (a base plus
// a share-class part plus an exchange suffix) stay well under fifteen
// characters, and anything longer is a sentence, not an identifier.
const maxTickerLen = 15

// tickerPattern is the exchange-ticker shape plausibleTicker accepts: a base
// of letters and digits (VOO, 7203), an optional "-" share-class part (BRK-B)
// and at most one "." exchange suffix (IWDA.AS, 7203.T). Nothing else: no
// space, no "^" quote symbol, no punctuation a URL or a file name would have
// to escape.
var tickerPattern = regexp.MustCompile(`^[A-Z0-9]{1,12}(-[A-Z0-9]{1,4})?(\.[A-Z]{1,4})?$`)

// plausibleTicker reports whether id has the shape of an exchange ticker.
func plausibleTicker(id string) bool {
	u := strings.ToUpper(strings.TrimSpace(id))
	return len(u) <= maxTickerLen && tickerPattern.MatchString(u)
}

// PlausibleID reports whether an identifier is well-formed enough to be worth
// resolving against the external sources, judged on its SHAPE alone: a valid
// ISIN (check digit included, see IsISIN) or a plausible exchange ticker (a
// base of letters and digits, an optional "-" share-class part, at most one
// "." exchange suffix, at most fifteen characters).
//
// It says nothing about whether such an instrument exists; it exists to keep
// junk out of the resolution path. An ISIN-shaped string whose check digit is
// wrong is refused rather than retried as a ticker: it is a typo, and the
// name-based fallbacks would gladly adopt an unrelated fund for it (pair this
// with FetchOptions.ExactOnly for the same reason).
//
// Callers that accept identifiers from untrusted hands use it as their first
// gate: the web app admits an identifier outside the bundled catalog
// (KnownLocal) only when PlausibleID accepts it.
func PlausibleID(id string) bool {
	u := strings.ToUpper(strings.TrimSpace(id))
	if isinPattern.MatchString(u) {
		return IsISIN(u)
	}
	return plausibleTicker(u)
}
