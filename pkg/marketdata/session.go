package marketdata

// The trading sessions a Quote can name. A quote whose source knows nothing
// of sessions (a daily close, a fund NAV, a nowcast) leaves Quote.Session
// empty rather than guessing.
const (
	sessionRegular = "regular"
	sessionPre     = "pre"
	sessionPost    = "post"
)

// freshestSession picks the most recent print among the regular session's and
// the venue's extended-hours ones, and names the session it came from.
//
// The rule is pure recency: an extended-hours candidate wins only when it has
// both a price and a timestamp STRICTLY after the regular one. That single
// guard is what makes the choice safe, and it is deliberately preferred to
// Yahoo's marketState enum, which is both wider than documented (PRE, PREPRE,
// REGULAR, POST, POSTPOST, CLOSED) and free to change. It also disposes of the
// stale-field trap on its own: during the regular session Yahoo still serves
// the morning's preMarketPrice, and before the pre-market opens it still serves
// last night's postMarketPrice, each dated at the instant it was struck, so
// each is accepted only while it really is the freshest thing known. That last
// case is a feature: at 03:00 in New York the last trade IS yesterday evening's
// after-hours print, and it comes back labelled "post" with its own Time, which
// is more than the 16:00 close says.
//
// Prices at or below zero are ignored, and a price with no timestamp is not a
// print: dated at the Unix epoch it would poison every caller.
func freshestSession(price float64, unix int64, prePrice *float64, preTime *int64, postPrice *float64, postTime *int64) (float64, int64, string) {
	session := sessionRegular
	for _, cand := range []struct {
		price *float64
		time  *int64
		name  string
	}{{prePrice, preTime, sessionPre}, {postPrice, postTime, sessionPost}} {
		if cand.price == nil || cand.time == nil || *cand.price <= 0 || *cand.time <= unix {
			continue
		}
		price, unix, session = *cand.price, *cand.time, cand.name
	}
	return price, unix, session
}
