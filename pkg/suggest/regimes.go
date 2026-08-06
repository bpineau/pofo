package suggest

import "strings"

// Regime is one of the four macro environments a diversified portfolio
// should cover (the growth x inflation quadrants behind All-Weather- and
// Dragon-style portfolios).
type Regime string

// The four macro regimes.
const (
	Growth    Regime = "growth"    // rising growth, benign inflation
	Deflation Regime = "deflation" // falling growth, low/falling inflation (recession)
	Inflation Regime = "inflation" // rising inflation
	Crisis    Regime = "crisis"    // protracted dislocations / stagflation / divergent trends
)

// AllRegimes lists the four regimes in display order.
var AllRegimes = []Regime{Growth, Deflation, Inflation, Crisis}

// Regimes maps an asset's metadata to the regimes it helps in. The mapping
// is driven by asset_class and strategy, with a few keyword refinements for
// equity sub-cases (gold miners and energy lean inflationary, value and
// dividend tilts add an inflation leg). An asset can help in several.
func Regimes(m Meta) []Regime {
	hint := strings.ToLower(m.Underlying + " " + m.Benchmark + " " + m.Notes)
	has := func(words ...string) bool {
		for _, w := range words {
			if strings.Contains(hint, w) {
				return true
			}
		}
		return false
	}

	switch m.AssetClass {
	case "equity":
		switch {
		case has("gold", "mining", "miner", "precious metal"):
			return []Regime{Inflation}
		case has("energy", "oil", "commodit"):
			return []Regime{Growth, Inflation}
		case has("value", "dividend", "high yield equity"):
			return []Regime{Growth, Inflation}
		default:
			return []Regime{Growth}
		}
	case "multi-asset":
		return []Regime{Growth, Deflation}
	case "real-estate":
		return []Regime{Growth, Inflation}
	case "corporate-bond":
		return []Regime{Growth, Deflation}
	case "aggregate-bond", "government-bond":
		return []Regime{Deflation}
	case "inflation-linked-bond":
		return []Regime{Inflation, Deflation}
	case "money-market":
		return []Regime{Deflation}
	case "gold":
		return []Regime{Inflation, Crisis}
	case "broad-commodity":
		return []Regime{Inflation, Crisis}
	case "managed-futures":
		return []Regime{Crisis, Inflation}
	case "long-volatility", "tail-risk":
		return []Regime{Deflation, Crisis}
	default: // "other" (e.g. global macro hedge funds)
		return []Regime{Crisis}
	}
}

// Coverage returns the total weight of the holdings that help in each
// regime (an asset helping in several contributes its full weight to each),
// plus the weight of holdings with no metadata (unclassified). Weights are
// fractions; coverage values are not normalized and can exceed 1 when
// assets span regimes.
func Coverage(holdings []Holding) (cov map[Regime]float64, unclassified float64) {
	cov = map[Regime]float64{Growth: 0, Deflation: 0, Inflation: 0, Crisis: 0}
	for _, h := range holdings {
		if !h.HasMeta {
			unclassified += h.Weight
			continue
		}
		for _, r := range Regimes(h.Meta) {
			cov[r] += h.Weight
		}
	}
	return cov, unclassified
}

// Gaps returns the regimes whose coverage is at or below threshold (a
// fraction of portfolio weight), ordered from least to most covered.
func Gaps(cov map[Regime]float64, threshold float64) []Regime {
	var gaps []Regime
	for _, r := range AllRegimes {
		if cov[r] <= threshold {
			gaps = append(gaps, r)
		}
	}
	// Least-covered first.
	for i := 1; i < len(gaps); i++ {
		for j := i; j > 0 && cov[gaps[j]] < cov[gaps[j-1]]; j-- {
			gaps[j], gaps[j-1] = gaps[j-1], gaps[j]
		}
	}
	return gaps
}
