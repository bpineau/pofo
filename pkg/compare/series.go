package compare

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// effectiveCurrencies is the list of base currencies a spec expands into:
// its "#meta currencies" list when set, otherwise the single CLI default.
func effectiveCurrencies(spec *portfolio.Spec, def string) []string {
	if len(spec.Currencies) > 0 {
		return spec.Currencies
	}
	return []string{def}
}

// inflationSeries returns the consumer-price index used to deflate nominal
// returns into real (purchasing-power) ones for the base currency, and whether
// one is available. The euro is deflated by French HICP (^HICP-FR, the long
// bundled series, ~1955→), the dollar by the US CPI (^CPI-US, bundled from
// 1913); other currencies have no wired CPI yet, so their real drawdown/TTR
// columns are simply omitted.
func inflationSeries(ctx context.Context, c *marketdata.Client, currency string, from time.Time) (*marketdata.Series, bool) {
	sym := ""
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "EUR":
		sym = "^HICP-FR"
	case "USD":
		sym = "^CPI-US"
	}
	if sym == "" {
		return nil, false
	}
	s, err := c.Fetch(ctx, sym, from)
	if err != nil || s == nil || len(s.Points) < 2 {
		if err != nil {
			log.Printf("warning: inflation index %s unavailable (%v); real drawdowns omitted", sym, err)
		}
		return nil, false
	}
	return s, true
}

// negate returns a sign-flipped copy: portfolio flows (contributions
// positive) become investor flows (money out of pocket negative).
func negate(xs []float64) []float64 {
	out := make([]float64, len(xs))
	for i, x := range xs {
		out[i] = -x
	}
	return out
}

// window returns the bounds [i, j) of dates within [from, to].
func window(dates []time.Time, from, to time.Time) (int, int) {
	i := sort.Search(len(dates), func(k int) bool { return !dates[k].Before(from) })
	j := sort.Search(len(dates), func(k int) bool { return dates[k].After(to) })
	return i, j
}

// rebase rescales a value slice so that it starts at 100.
func rebase(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = v / values[0] * 100
	}
	return out
}
