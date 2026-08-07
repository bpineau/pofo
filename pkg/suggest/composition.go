package suggest

import (
	"sort"
	"strings"
)

// composition.go computes look-through composition splits of a portfolio:
// what the holdings actually are once stacked funds are opened up into their
// legs (asset classes), where the money sits (geography), which fiat
// currencies it is exposed to, which equity sectors the equity sleeve holds,
// and how much interest-rate duration the book carries. All functions take
// the same []Holding as Coverage and return label → weight maps (fractions
// of portfolio capital unless noted); presentation (sorting, small-slice
// merging, colors) is left to the caller.

// Reserved bucket labels used by the composition splits.
const (
	// BucketUnknown collects holdings without catalog metadata.
	BucketUnknown = "Unknown"
	// BucketNoCountry collects asset classes for which a country split is
	// meaningless (gold, commodities, managed futures, volatility, cash),
	// as opposed to "Other", which aggregates real but small countries.
	BucketNoCountry = "No country"
	// CurrencyNone marks capital without fiat-currency exposure: real
	// assets (gold, commodities), whose price in the investor's currency
	// does not depend on the quote currency of the wrapper.
	CurrencyNone = "None"
	// CurrencyDynamic marks capital whose currency exposure is set by a
	// futures book (managed futures, volatility strategies): long or short
	// any currency at any time, with no static figure to report.
	CurrencyDynamic = "Dynamic"
	// CurrencyOther collects region labels that map to no single currency
	// ("Other developed", "Emerging markets", …).
	CurrencyOther = "Other"
)

// AssetClassSplit returns the portfolio's look-through notional weight per
// asset class: a plain holding contributes its weight to its own class, a
// stacked / efficient-core fund (one with an Exposures breakdown) contributes
// each leg's notional, so a 28 % position in a 90/60 fund counts 25.2 points
// of equity plus 16.8 of government bonds. Values sum to the portfolio's
// total economic exposure and can exceed 1; normalize for a share-of-exposure
// view. Holdings without metadata land in BucketUnknown.
func AssetClassSplit(holdings []Holding) map[string]float64 {
	out := map[string]float64{}
	for _, h := range holdings {
		if h.Weight <= 0 {
			continue
		}
		switch {
		case !h.HasMeta:
			out[BucketUnknown] += h.Weight
		case len(h.Meta.Exposures) > 0:
			for class, notional := range h.Meta.Exposures {
				out[class] += h.Weight * notional
			}
		case h.Meta.AssetClass != "":
			out[h.Meta.AssetClass] += h.Weight
		default:
			out[BucketUnknown] += h.Weight
		}
	}
	return out
}

// noCountryClass reports whether a country split is meaningless for the
// class, rather than merely missing from the catalog.
func noCountryClass(class string) bool {
	switch class {
	case "gold", "broad-commodity", "managed-futures", "long-volatility",
		"tail-risk", "money-market", "other":
		return true
	}
	return false
}

// GeographySplit returns the capital-weighted country/region split of the
// portfolio, region labels canonicalized (CanonRegion). Holdings of a class
// with no meaningful geography (gold, managed futures, …) land in
// BucketNoCountry; holdings whose catalog record merely lacks the split, or
// that have no metadata at all, land in BucketUnknown. Values sum to ~1.
func GeographySplit(holdings []Holding) map[string]float64 {
	out := map[string]float64{}
	for _, h := range holdings {
		if h.Weight <= 0 {
			continue
		}
		switch {
		case !h.HasMeta:
			out[BucketUnknown] += h.Weight
		case len(h.Meta.Geography) > 0:
			for region, pct := range h.Meta.Geography {
				out[CanonRegion(region)] += h.Weight * pct / 100
			}
		case noCountryClass(h.Meta.AssetClass):
			out[BucketNoCountry] += h.Weight
		default:
			out[BucketUnknown] += h.Weight
		}
	}
	return out
}

// equityNotional is the holding's look-through equity exposure per unit of
// weight: 1 for a plain equity fund, the equity leg's notional for a stacked
// fund, 0 otherwise.
func equityNotional(m Meta) float64 {
	if len(m.Exposures) > 0 {
		return m.Exposures["equity"]
	}
	if m.AssetClass == "equity" {
		return 1
	}
	return 0
}

// EquitySectorSplit returns the sector split of the portfolio's look-through
// equity sleeve (plain equity holdings plus the equity legs of stacked
// funds), sector labels canonicalized (CanonSector). The split's values are
// fractions of the sleeve, not of the portfolio, so they sum to ~1; equity
// capital whose record carries no sector breakdown lands in BucketUnknown.
// The second result is the sleeve's notional size as a fraction of portfolio
// capital (0 when the portfolio holds no classified equity).
func EquitySectorSplit(holdings []Holding) (split map[string]float64, equity float64) {
	split = map[string]float64{}
	for _, h := range holdings {
		if h.Weight <= 0 || !h.HasMeta {
			continue
		}
		e := h.Weight * equityNotional(h.Meta)
		if e <= 0 {
			continue
		}
		equity += e
		if len(h.Meta.Sectors) == 0 {
			split[BucketUnknown] += e
			continue
		}
		for sector, pct := range h.Meta.Sectors {
			split[CanonSector(sector)] += e * pct / 100
		}
	}
	if equity <= 0 {
		return nil, 0
	}
	for k, v := range split {
		split[k] = v / equity
	}
	return split, equity
}

// CurrencySplit returns the portfolio's look-through fiat-currency exposure:
// which currencies the capital is actually at risk in, regardless of quote
// or listing currency (a EUR-quoted MSCI World wrapper is ~70 % USD; a
// USD-quoted gold ETC has no fiat exposure at all). Values are fractions of
// capital summing to ~1, keyed by ISO code or by the CurrencyNone /
// CurrencyDynamic / CurrencyOther / BucketUnknown buckets.
//
// Per holding, the first applicable rule wins:
//  1. an explicit catalog CurrencyExposure map (percent of capital; any
//     shortfall below 100 counts as CurrencyNone);
//  2. gold and broad commodities → CurrencyNone (real assets);
//  3. a currency-hedged share class → 100 % HedgedTo;
//  4. managed futures, long volatility, tail risk and "other" → CurrencyDynamic;
//  5. the geography split, each country mapped to its currency (eurozone
//     members → EUR; region labels with no single currency → CurrencyOther);
//  6. the quote currency, as a last resort (money markets, bonds without a
//     geography split).
//
// Rule 5 reads a stacked fund's geography, which describes its equity leg:
// a fair approximation, since bond futures legs carry no FX exposure. Rule 6
// is wrong for funds whose denomination differs from their listing (some
// corporate-bond and EM funds); give those an explicit CurrencyExposure.
func CurrencySplit(holdings []Holding) map[string]float64 {
	out := map[string]float64{}
	for _, h := range holdings {
		if h.Weight <= 0 {
			continue
		}
		if !h.HasMeta {
			out[BucketUnknown] += h.Weight
			continue
		}
		m := h.Meta
		switch {
		case len(m.CurrencyExposure) > 0:
			covered := 0.0
			for cur, pct := range m.CurrencyExposure {
				out[cur] += h.Weight * pct / 100
				covered += pct
			}
			if covered < 100 {
				out[CurrencyNone] += h.Weight * (100 - covered) / 100
			}
		case m.AssetClass == "gold" || m.AssetClass == "broad-commodity":
			out[CurrencyNone] += h.Weight
		case m.CurrencyHedged && m.HedgedTo != "":
			out[m.HedgedTo] += h.Weight
		case m.AssetClass == "managed-futures" || m.AssetClass == "long-volatility" ||
			m.AssetClass == "tail-risk" || m.AssetClass == "other":
			out[CurrencyDynamic] += h.Weight
		case len(m.Geography) > 0:
			for region, pct := range m.Geography {
				out[regionCurrency(region)] += h.Weight * pct / 100
			}
		case m.Currency != "":
			out[m.Currency] += h.Weight
		default:
			out[BucketUnknown] += h.Weight
		}
	}
	return out
}

// FXProfile summarizes a CurrencySplit against the investor's base currency.
// Base + Foreign + NonFiat + Unknown ≈ 1.
type FXProfile struct {
	Base     float64 // base-currency share: native, or hedged back to it
	Foreign  float64 // unhedged foreign fiat share (CurrencyOther included)
	Top      string  // largest single unhedged foreign currency ("" when none)
	TopShare float64 // its share of portfolio capital
	NonFiat  float64 // real assets and futures books (CurrencyNone + CurrencyDynamic)
	Unknown  float64 // holdings without metadata
}

// CurrencyProfile condenses a CurrencySplit into the number that matters to
// an investor billed in base: how much of the book moves with foreign
// exchange rates, and which currency dominates that part.
func CurrencyProfile(split map[string]float64, base string) FXProfile {
	p := FXProfile{}
	for cur, w := range split {
		switch cur {
		case base:
			p.Base += w
		case CurrencyNone, CurrencyDynamic:
			p.NonFiat += w
		case BucketUnknown:
			p.Unknown += w
		default:
			p.Foreign += w
			if cur != CurrencyOther && w > p.TopShare {
				p.Top, p.TopShare = cur, w
			}
		}
	}
	return p
}

// DurationLedger aggregates the portfolio's look-through interest-rate
// duration, in duration-years per unit of capital: a 10 % position in a 20-year
// bond fund contributes 2.0. Stacked funds count each bond leg's notional at
// the fund's Duration figure (the duration of its bond exposure).
type DurationLedger struct {
	Nominal float64 // from nominal bonds (government, aggregate, corporate)
	Real    float64 // from inflation-linked bonds (real-rate duration)
	Missing float64 // bond notional (fraction of capital) with no duration figure
}

// DurationSplit computes the portfolio's duration ledger. Money-market legs
// are ignored (duration ≈ 0); bond legs without a catalog Duration are
// tallied in Missing rather than silently dropped.
func DurationSplit(holdings []Holding) DurationLedger {
	var led DurationLedger
	for _, h := range holdings {
		if h.Weight <= 0 || !h.HasMeta {
			continue
		}
		legs := h.Meta.Exposures
		if len(legs) == 0 {
			legs = map[string]float64{h.Meta.AssetClass: 1}
		}
		for class, notional := range legs {
			var bucket *float64
			switch class {
			case "government-bond", "aggregate-bond", "corporate-bond":
				bucket = &led.Nominal
			case "inflation-linked-bond":
				bucket = &led.Real
			default:
				continue
			}
			if h.Meta.Duration <= 0 {
				led.Missing += h.Weight * notional
				continue
			}
			*bucket += h.Weight * notional * h.Meta.Duration
		}
	}
	return led
}

// Contributor is one holding's notional contribution to a coverage category.
type Contributor struct {
	Index  int     // index into the holdings slice (stable, e.g. for colors)
	ID     string  // the holding's identifier
	Weight float64 // notional contribution, as a fraction of portfolio capital
}

// Contributors decomposes Coverage: for each category of the framework, the
// holdings that cover it and how much each contributes (largest first), using
// the same notional accounting (a stacked fund contributes each leg's
// notional to that leg's categories). Holdings without metadata contribute
// nowhere.
func Contributors(holdings []Holding, fw Framework) map[Category][]Contributor {
	out := map[Category][]Contributor{}
	for i, h := range holdings {
		if !h.HasMeta || h.Weight <= 0 {
			continue
		}
		for c, frac := range fw.Contribution(h.Meta) {
			if w := h.Weight * frac; w > 0 {
				out[c] = append(out[c], Contributor{Index: i, ID: h.ID, Weight: w})
			}
		}
	}
	for _, list := range out {
		sort.SliceStable(list, func(a, b int) bool { return list[a].Weight > list[b].Weight })
	}
	return out
}

// CanonRegion canonicalizes a catalog geography label so synonymous spellings
// aggregate ("United States" and "US", "Other Developed" and "Other
// developed"). Unrecognized labels pass through unchanged.
func CanonRegion(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "us", "usa", "u.s.", "u.s.a.", "united states", "united states of america":
		return "US"
	case "uk", "u.k.", "united kingdom", "great britain":
		return "UK"
	case "other developed":
		return "Other developed"
	case "other emerging", "other emerging markets", "other em", "emerging markets", "emerging":
		return "Other emerging"
	case "other eurozone":
		return "Other eurozone"
	case "other europe", "other europe ex-uk":
		return "Other Europe"
	case "south korea", "korea", "korea (south)":
		return "South Korea"
	}
	return strings.TrimSpace(s)
}

// CanonSector canonicalizes a catalog sector label to the GICS sector it
// names ("Technology" and "Information Technology", "Basic Materials" and
// "Materials"). Unrecognized labels pass through unchanged.
func CanonSector(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "technology", "information technology", "info tech":
		return "Information Technology"
	case "basic materials", "materials":
		return "Materials"
	case "telecommunications", "telecom", "communication services", "communications":
		return "Communication Services"
	case "consumer cyclical", "consumer discretionary":
		return "Consumer Discretionary"
	case "consumer defensive", "consumer staples":
		return "Consumer Staples"
	case "financial services", "financials":
		return "Financials"
	case "healthcare", "health care":
		return "Health Care"
	}
	return strings.TrimSpace(s)
}

// regionCurrency maps a canonicalized geography label to the fiat currency a
// holding there exposes the investor to. Eurozone members map to EUR; labels
// spanning several currencies map to CurrencyOther.
func regionCurrency(region string) string {
	switch CanonRegion(region) {
	case "US", "North America", "Foreign (USD-denominated)":
		return "USD" // North America ≈ 90 % US in cap-weighted indices
	case "UK":
		return "GBP"
	case "Austria", "Belgium", "Croatia", "Cyprus", "Estonia", "Finland",
		"France", "Germany", "Greece", "Ireland", "Italy", "Latvia",
		"Lithuania", "Luxembourg", "Malta", "Netherlands", "Portugal",
		"Slovakia", "Slovenia", "Spain", "Eurozone", "Other eurozone":
		return "EUR"
	case "Japan":
		return "JPY"
	case "Switzerland":
		return "CHF"
	case "Sweden":
		return "SEK"
	case "Denmark":
		return "DKK"
	case "Norway":
		return "NOK"
	case "Canada":
		return "CAD"
	case "Australia":
		return "AUD"
	case "New Zealand":
		return "NZD"
	case "China":
		return "CNY"
	case "Hong Kong":
		return "HKD"
	case "Taiwan":
		return "TWD"
	case "South Korea":
		return "KRW"
	case "India":
		return "INR"
	case "Singapore":
		return "SGD"
	case "Brazil":
		return "BRL"
	case "Mexico":
		return "MXN"
	case "South Africa":
		return "ZAR"
	case "Saudi Arabia":
		return "SAR"
	case "Indonesia":
		return "IDR"
	case "Thailand":
		return "THB"
	case "Poland":
		return "PLN"
	}
	return CurrencyOther
}
