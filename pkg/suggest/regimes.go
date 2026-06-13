package suggest

import "strings"

// Category is one bucket of a classification framework (a macro regime, or a
// risk factor). A portfolio "covers" a category when it holds assets that
// help in it.
type Category string

// The macro regimes — the growth x inflation quadrants behind All-Weather-
// and Dragon-style portfolios (the default framework).
const (
	Growth    Category = "growth"    // rising growth, benign inflation
	Deflation Category = "deflation" // falling growth, low/falling inflation (recession)
	Inflation Category = "inflation" // rising inflation
	Crisis    Category = "crisis"    // protracted dislocations / stagflation / divergent trends
)

// The risk factors (the alternative framework). Coverage by factor is
// coarser than by regime: this catalog holds many diversifiers (gold,
// commodities, managed futures, volatility) that are not Fama-French
// factors, so they all land in "alternative".
const (
	Market      Category = "market"      // broad equity beta
	Size        Category = "size"        // small-cap premium
	Value       Category = "value"       // cheap vs expensive
	Momentum    Category = "momentum"    // recent winners
	Quality     Category = "quality"     // profitable, low-leverage
	Term        Category = "term"        // duration / interest-rate exposure
	Credit      Category = "credit"      // corporate credit spread
	Alternative Category = "alternative" // gold, commodities, CTAs, volatility, macro
	Cash        Category = "cash"        // money-market / very short rates
)

// AllRegimes lists the regime categories in display order.
var AllRegimes = []Category{Growth, Deflation, Inflation, Crisis}

// allFactors lists the factor categories in display order.
var allFactors = []Category{Market, Size, Value, Momentum, Quality, Term, Credit, Alternative, Cash}

// Framework is a way to classify assets into categories a portfolio should
// cover: the macro regimes (RegimeFramework) or the risk factors
// (FactorFramework).
type Framework struct {
	Name       string
	Categories []Category
	Classify   func(Meta) []Category
}

// RegimeFramework is the default: the four macro regimes.
func RegimeFramework() Framework {
	return Framework{Name: "regimes", Categories: AllRegimes, Classify: regimeClassify}
}

// FactorFramework is the alternative: the risk factors.
func FactorFramework() Framework {
	return Framework{Name: "factors", Categories: allFactors, Classify: factorClassify}
}

// Regimes maps an asset to the macro regimes it helps in (the default
// framework's classifier).
func Regimes(m Meta) []Category { return regimeClassify(m) }

// hinter matches keywords against an asset's factual descriptors (underlying
// and benchmark index). It deliberately ignores the free-text notes, whose
// boilerplate ("distributes annual dividends") would trigger false factor
// matches.
func hinter(m Meta) func(...string) bool {
	hint := strings.ToLower(m.Underlying + " " + m.Benchmark)
	return func(words ...string) bool {
		for _, w := range words {
			if strings.Contains(hint, w) {
				return true
			}
		}
		return false
	}
}

// regimeClassify drives the mapping from asset_class and strategy, with a few
// keyword refinements for equity sub-cases (gold miners and energy lean
// inflationary, value and dividend tilts add an inflation leg).
func regimeClassify(m Meta) []Category {
	has := hinter(m)
	switch m.AssetClass {
	case "equity":
		switch {
		case has("gold", "mining", "miner", "precious metal"):
			return []Category{Inflation}
		case has("energy", "oil", "commodit"):
			return []Category{Growth, Inflation}
		case has("value", "high dividend", "dividend yield", "dividend leader"):
			return []Category{Growth, Inflation}
		default:
			return []Category{Growth}
		}
	case "multi-asset":
		return []Category{Growth, Deflation}
	case "real-estate":
		return []Category{Growth, Inflation}
	case "corporate-bond":
		return []Category{Growth, Deflation}
	case "aggregate-bond", "government-bond":
		return []Category{Deflation}
	case "inflation-linked-bond":
		return []Category{Inflation, Deflation}
	case "money-market":
		return []Category{Deflation}
	case "gold":
		return []Category{Inflation, Crisis}
	case "broad-commodity":
		return []Category{Inflation, Crisis}
	case "managed-futures":
		return []Category{Crisis, Inflation}
	case "long-volatility", "tail-risk":
		return []Category{Deflation, Crisis}
	default: // "other" (e.g. global macro hedge funds)
		return []Category{Crisis}
	}
}

// factorClassify is a best-effort mapping to risk factors. Equity factor
// tilts are read from the benchmark/name; bonds split into term and credit;
// every non-factor diversifier lands in "alternative".
func factorClassify(m Meta) []Category {
	has := hinter(m)
	switch m.AssetClass {
	case "equity":
		switch {
		case has("gold", "mining", "miner", "precious metal"):
			return []Category{Alternative}
		case has("multi-factor", "multifactor", "diversified factor"):
			return []Category{Market, Value, Momentum, Quality}
		}
		out := []Category{Market}
		if has("energy", "oil", "commodit") {
			out = append(out, Alternative)
		}
		if has("small cap", "small-cap", "smallcap") {
			out = append(out, Size)
		}
		if has("value", "high dividend", "dividend yield", "dividend leader") {
			out = append(out, Value)
		}
		if has("momentum") {
			out = append(out, Momentum)
		}
		if has("quality") {
			out = append(out, Quality)
		}
		return out
	case "multi-asset":
		return []Category{Market, Term}
	case "real-estate":
		return []Category{Market, Alternative}
	case "government-bond", "inflation-linked-bond":
		return []Category{Term}
	case "aggregate-bond":
		return []Category{Term, Credit}
	case "corporate-bond":
		return []Category{Credit, Term}
	case "money-market":
		return []Category{Cash}
	default: // gold, broad-commodity, managed-futures, long-volatility, tail-risk, other
		return []Category{Alternative}
	}
}

// Coverage returns the total weight of the holdings that help in each
// category of the framework (an asset helping in several contributes its
// full weight to each), plus the weight of holdings with no metadata.
// Coverage values are not normalized and can exceed 1.
func Coverage(holdings []Holding, fw Framework) (cov map[Category]float64, unclassified float64) {
	cov = map[Category]float64{}
	for _, c := range fw.Categories {
		cov[c] = 0
	}
	for _, h := range holdings {
		if !h.HasMeta {
			unclassified += h.Weight
			continue
		}
		for _, c := range fw.Classify(h.Meta) {
			cov[c] += h.Weight
		}
	}
	return cov, unclassified
}

// Gaps returns the framework categories whose coverage is at or below
// threshold (a fraction of portfolio weight), least-covered first.
func Gaps(cov map[Category]float64, fw Framework, threshold float64) []Category {
	var gaps []Category
	for _, c := range fw.Categories {
		if cov[c] <= threshold {
			gaps = append(gaps, c)
		}
	}
	for i := 1; i < len(gaps); i++ {
		for j := i; j > 0 && cov[gaps[j]] < cov[gaps[j-1]]; j-- {
			gaps[j], gaps[j-1] = gaps[j-1], gaps[j]
		}
	}
	return gaps
}
