package suggest

import (
	"math"
	"testing"
)

// householdHoldings mirrors the dragon-household build in miniature: a 90/60
// stacked core, a long euro sovereign, trend (one hedged, one not), gold,
// euro linkers and two plain equity funds.
func householdHoldings() []Holding {
	return []Holding{
		{ID: "NTSG", Weight: 0.28, HasMeta: true, Meta: Meta{
			AssetClass: "multi-asset",
			Exposures:  map[string]float64{"equity": 0.9, "government-bond": 0.6},
			Geography:  map[string]float64{"United States": 70, "Japan": 6, "UK": 4, "Other developed": 20},
			Duration:   7,
			Currency:   "USD",
		}},
		{ID: "MTH", Weight: 0.04, HasMeta: true, Meta: Meta{
			AssetClass: "government-bond",
			Geography:  map[string]float64{"France": 40, "Italy": 22, "Germany": 15, "Spain": 13, "Other eurozone": 10},
			Duration:   20,
			Currency:   "EUR",
		}},
		{ID: "DBMFE", Weight: 0.13, HasMeta: true, Meta: Meta{
			AssetClass: "managed-futures", Currency: "EUR",
		}},
		{ID: "AQR", Weight: 0.13, HasMeta: true, Meta: Meta{
			AssetClass: "managed-futures", Currency: "EUR",
			CurrencyHedged: true, HedgedTo: "EUR",
		}},
		{ID: "IGLN", Weight: 0.15, HasMeta: true, Meta: Meta{
			AssetClass: "gold", Currency: "USD",
		}},
		{ID: "IBCI", Weight: 0.13, HasMeta: true, Meta: Meta{
			AssetClass: "inflation-linked-bond",
			Geography:  map[string]float64{"France": 35, "Italy": 30, "Germany": 15, "Spain": 12, "Other": 8},
			Duration:   8,
			Currency:   "EUR",
		}},
		{ID: "WPEA", Weight: 0.05, HasMeta: true, Meta: Meta{
			AssetClass: "equity",
			Geography:  map[string]float64{"US": 72, "Japan": 6, "United Kingdom": 4, "Other": 18},
			Sectors:    map[string]float64{"Information Technology": 28, "Financials": 16, "Health Care": 9, "Other": 47},
			Currency:   "EUR",
		}},
		{ID: "SMALL", Weight: 0.09, HasMeta: true, Meta: Meta{
			AssetClass: "equity",
			Geography:  map[string]float64{"France": 22, "Germany": 9, "Italy": 17, "Other Europe ex-UK": 52},
			Sectors:    map[string]float64{"Industrials": 57, "Technology": 6, "Other": 37},
			Currency:   "EUR",
		}},
	}
}

func TestAssetClassSplitLookThrough(t *testing.T) {
	split := AssetClassSplit(householdHoldings())
	// The 90/60 core opens into legs: equity 0.28*0.9 + 0.05 + 0.09 = 0.392,
	// government bonds 0.28*0.6 + 0.04 = 0.208. No multi-asset bucket remains.
	if _, ok := split["multi-asset"]; ok {
		t.Fatalf("multi-asset should be split into legs, got %v", split)
	}
	if math.Abs(split["equity"]-0.392) > 1e-9 || math.Abs(split["government-bond"]-0.208) > 1e-9 {
		t.Fatalf("split = %v, want equity 0.392, government-bond 0.208", split)
	}
	total := 0.0
	for _, v := range split {
		total += v
	}
	if math.Abs(total-1.14) > 1e-9 { // 1 + the 0.14 embedded leverage of NTSG
		t.Fatalf("total notional = %v, want 1.14", total)
	}
}

func TestAssetClassSplitUnknown(t *testing.T) {
	split := AssetClassSplit([]Holding{{ID: "X", Weight: 1}})
	if math.Abs(split[BucketUnknown]-1) > 1e-9 {
		t.Fatalf("split = %v, want all in %q", split, BucketUnknown)
	}
}

func TestGeographySplit(t *testing.T) {
	split := GeographySplit(householdHoldings())
	// Gold and both trend sleeves have no meaningful country: 0.15+0.13+0.13.
	if math.Abs(split[BucketNoCountry]-0.41) > 1e-9 {
		t.Fatalf("no-country = %v, want 0.41 (split %v)", split[BucketNoCountry], split)
	}
	// US aggregates the "United States" and "US" spellings:
	// 0.28*0.70 + 0.05*0.72 = 0.232.
	if math.Abs(split["US"]-0.232) > 1e-9 {
		t.Fatalf("US = %v, want 0.232", split["US"])
	}
	// UK aggregates "UK" and "United Kingdom": 0.28*0.04 + 0.05*0.04.
	if math.Abs(split["UK"]-0.0132) > 1e-9 {
		t.Fatalf("UK = %v, want 0.0132", split["UK"])
	}
	total := 0.0
	for _, v := range split {
		total += v
	}
	if math.Abs(total-1) > 1e-9 {
		t.Fatalf("total = %v, want 1", total)
	}
}

func TestGeographySplitMissingVsMeaningless(t *testing.T) {
	split := GeographySplit([]Holding{
		{ID: "GOLD", Weight: 0.5, HasMeta: true, Meta: Meta{AssetClass: "gold"}},
		{ID: "EQ", Weight: 0.3, HasMeta: true, Meta: Meta{AssetClass: "equity"}}, // split missing
		{ID: "X", Weight: 0.2}, // no metadata
	})
	if math.Abs(split[BucketNoCountry]-0.5) > 1e-9 || math.Abs(split[BucketUnknown]-0.5) > 1e-9 {
		t.Fatalf("split = %v, want 0.5 no-country + 0.5 unknown", split)
	}
}

func TestEquitySectorSplit(t *testing.T) {
	split, equity := EquitySectorSplit(householdHoldings())
	// Equity sleeve: 0.28*0.9 + 0.05 + 0.09 = 0.392 of capital.
	if math.Abs(equity-0.392) > 1e-9 {
		t.Fatalf("equity sleeve = %v, want 0.392", equity)
	}
	// The NTSG equity leg (0.252) has no sector split: Unknown share.
	if math.Abs(split[BucketUnknown]-0.252/0.392) > 1e-9 {
		t.Fatalf("unknown share = %v, want %v", split[BucketUnknown], 0.252/0.392)
	}
	// "Technology" (SMALL) and "Information Technology" (WPEA) aggregate:
	// (0.05*0.28 + 0.09*0.06) / 0.392.
	want := (0.05*0.28 + 0.09*0.06) / 0.392
	if math.Abs(split["Information Technology"]-want) > 1e-9 {
		t.Fatalf("tech share = %v, want %v", split["Information Technology"], want)
	}
	total := 0.0
	for _, v := range split {
		total += v
	}
	if math.Abs(total-1) > 1e-9 {
		t.Fatalf("total = %v, want 1", total)
	}
}

func TestEquitySectorSplitNoEquity(t *testing.T) {
	split, equity := EquitySectorSplit([]Holding{
		{ID: "GOLD", Weight: 1, HasMeta: true, Meta: Meta{AssetClass: "gold"}},
	})
	if split != nil || equity != 0 {
		t.Fatalf("want no sector split for an equity-free book, got %v (%v)", split, equity)
	}
}

func TestCurrencySplitRules(t *testing.T) {
	split := CurrencySplit(householdHoldings())
	// Gold → None; unhedged trend → Dynamic; hedged trend → EUR.
	if math.Abs(split[CurrencyNone]-0.15) > 1e-9 {
		t.Fatalf("none = %v, want 0.15", split[CurrencyNone])
	}
	if math.Abs(split[CurrencyDynamic]-0.13) > 1e-9 {
		t.Fatalf("dynamic = %v, want 0.13", split[CurrencyDynamic])
	}
	// EUR: hedged AQR 0.13 + MTH 0.04 (all-eurozone geography) + IBCI
	// 0.13*0.92 + WPEA 0 + SMALL 0.48*0.09.
	wantEUR := 0.13 + 0.04 + 0.13*0.92 + 0.09*0.48
	if math.Abs(split["EUR"]-wantEUR) > 1e-9 {
		t.Fatalf("EUR = %v, want %v", split["EUR"], wantEUR)
	}
	// USD via geography look-through, not quote currency: NTSG 0.28*0.70 +
	// WPEA 0.05*0.72 (EUR-quoted, USD-exposed).
	if math.Abs(split["USD"]-(0.28*0.70+0.05*0.72)) > 1e-9 {
		t.Fatalf("USD = %v, want %v", split["USD"], 0.28*0.70+0.05*0.72)
	}
	total := 0.0
	for _, v := range split {
		total += v
	}
	if math.Abs(total-1) > 1e-9 {
		t.Fatalf("total = %v, want 1", total)
	}
}

func TestCurrencySplitOverride(t *testing.T) {
	// An explicit currency_exposure wins over every derivation rule, and a
	// shortfall below 100 lands in None.
	split := CurrencySplit([]Holding{
		{ID: "NTSG", Weight: 0.5, HasMeta: true, Meta: Meta{
			AssetClass:       "multi-asset",
			Geography:        map[string]float64{"US": 100},
			CurrencyExposure: map[string]float64{"USD": 60, "JPY": 10},
			Currency:         "USD",
		}},
		{ID: "MM", Weight: 0.5, HasMeta: true, Meta: Meta{
			AssetClass: "money-market", Currency: "EUR", // rule 6 fallback
		}},
	})
	if math.Abs(split["USD"]-0.30) > 1e-9 || math.Abs(split["JPY"]-0.05) > 1e-9 ||
		math.Abs(split[CurrencyNone]-0.15) > 1e-9 || math.Abs(split["EUR"]-0.5) > 1e-9 {
		t.Fatalf("split = %v", split)
	}
}

func TestCurrencyProfile(t *testing.T) {
	p := CurrencyProfile(CurrencySplit(householdHoldings()), "EUR")
	if p.Top != "USD" {
		t.Fatalf("top foreign = %q, want USD", p.Top)
	}
	if math.Abs(p.NonFiat-0.28) > 1e-9 { // gold 0.15 + unhedged trend 0.13
		t.Fatalf("non-fiat = %v, want 0.28", p.NonFiat)
	}
	if p.Unknown != 0 {
		t.Fatalf("unknown = %v, want 0", p.Unknown)
	}
	if total := p.Base + p.Foreign + p.NonFiat + p.Unknown; math.Abs(total-1) > 1e-9 {
		t.Fatalf("profile total = %v, want 1", total)
	}
	if p.TopShare >= p.Foreign || p.TopShare <= 0 {
		t.Fatalf("top share %v out of range (foreign %v)", p.TopShare, p.Foreign)
	}
}

func TestDurationSplit(t *testing.T) {
	led := DurationSplit(householdHoldings())
	// Nominal: NTSG bond leg 0.28*0.6*7 + MTH 0.04*20 = 1.976.
	if math.Abs(led.Nominal-1.976) > 1e-9 {
		t.Fatalf("nominal = %v, want 1.976", led.Nominal)
	}
	// Real: IBCI 0.13*8 = 1.04.
	if math.Abs(led.Real-1.04) > 1e-9 {
		t.Fatalf("real = %v, want 1.04", led.Real)
	}
	if led.Missing != 0 {
		t.Fatalf("missing = %v, want 0", led.Missing)
	}
}

func TestDurationSplitMissing(t *testing.T) {
	led := DurationSplit([]Holding{
		{ID: "AGG", Weight: 0.5, HasMeta: true, Meta: Meta{AssetClass: "aggregate-bond"}}, // no Duration
	})
	if led.Nominal != 0 || math.Abs(led.Missing-0.5) > 1e-9 {
		t.Fatalf("ledger = %+v, want 0.5 missing", led)
	}
}

func TestContributors(t *testing.T) {
	holdings := householdHoldings()
	fw := RegimeFramework()
	contrib := Contributors(holdings, fw)

	// Every category's contributor sum must equal the aggregate Coverage.
	cov, _ := Coverage(holdings, fw)
	for cat, want := range cov {
		got := 0.0
		for _, c := range contrib[cat] {
			got += c.Weight
		}
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("%s: contributor sum %v != coverage %v", cat, got, want)
		}
	}

	// Growth is led by the NTSG equity leg (0.252), sorted descending.
	g := contrib[Growth]
	if len(g) == 0 || g[0].ID != "NTSG" || math.Abs(g[0].Weight-0.252) > 1e-9 {
		t.Fatalf("growth contributors = %v", g)
	}
	for i := 1; i < len(g); i++ {
		if g[i].Weight > g[i-1].Weight {
			t.Fatalf("growth contributors not sorted: %v", g)
		}
	}

	// Index points back into the holdings slice.
	for _, list := range contrib {
		for _, c := range list {
			if holdings[c.Index].ID != c.ID {
				t.Fatalf("index mismatch: %+v", c)
			}
		}
	}
}

// The catalog is written by hand from a dozen fund factsheets, which spell the
// same region and the same sector a dozen ways. Canonicalization is what makes
// the pies aggregate instead of showing "Technology" next to "Information
// Technology"; an unrecognized label passes through, trimmed, rather than
// being silently dropped into "Other".
func TestCanonRegionAndSector(t *testing.T) {
	regions := map[string]string{
		"united states": "US", "U.S.A.": "US", "USA": "US", "  US  ": "US",
		"United Kingdom": "UK", "Great Britain": "UK", "u.k.": "UK",
		"other developed":  "Other developed",
		"Emerging Markets": "Other emerging", "other em": "Other emerging",
		"Other Eurozone":     "Other eurozone",
		"Other Europe ex-UK": "Other Europe",
		"korea":              "South Korea",
		"Japan":              "Japan", // unrecognized labels pass through
		"Ruritania":          "Ruritania",
	}
	for in, want := range regions {
		if got := CanonRegion(in); got != want {
			t.Errorf("CanonRegion(%q) = %q, want %q", in, got, want)
		}
	}
	sectors := map[string]string{
		"Technology": "Information Technology", "info tech": "Information Technology",
		"Basic Materials": "Materials",
		"Telecom":         "Communication Services", "communications": "Communication Services",
		"Consumer Cyclical": "Consumer Discretionary", "consumer defensive": "Consumer Staples",
		"Financial Services": "Financials", "healthcare": "Health Care", "Health Care": "Health Care",
		"Industrials": "Industrials", "Widgets": "Widgets",
	}
	for in, want := range sectors {
		if got := CanonSector(in); got != want {
			t.Errorf("CanonSector(%q) = %q, want %q", in, got, want)
		}
	}
}

// The currency of a holding is where its earnings are, not where it is listed,
// so every geography label must resolve to a currency. Eurozone members share
// EUR; a label spanning several currencies (a regional aggregate) lands in
// CurrencyOther rather than pretending to a single one.
func TestRegionCurrency(t *testing.T) {
	cases := map[string]string{
		"United States": "USD", "North America": "USD", "Foreign (USD-denominated)": "USD",
		"United Kingdom": "GBP", "Japan": "JPY", "Switzerland": "CHF",
		"Sweden": "SEK", "Denmark": "DKK", "Norway": "NOK",
		"Canada": "CAD", "Australia": "AUD", "New Zealand": "NZD",
		"China": "CNY", "Hong Kong": "HKD", "Taiwan": "TWD", "korea": "KRW",
		"India": "INR", "Singapore": "SGD", "Brazil": "BRL", "Mexico": "MXN",
		"South Africa": "ZAR", "Saudi Arabia": "SAR", "Indonesia": "IDR",
		"Thailand": "THB", "Poland": "PLN",
		"France": "EUR", "Germany": "EUR", "Italy": "EUR", "Spain": "EUR",
		"Ireland": "EUR", "Other eurozone": "EUR", "Eurozone": "EUR",
		// No single currency behind these, nor behind an unknown label.
		"Other developed": CurrencyOther, "Other emerging": CurrencyOther,
		"Other Europe": CurrencyOther, "Other": CurrencyOther, "Ruritania": CurrencyOther,
	}
	for region, want := range cases {
		if got := regionCurrency(region); got != want {
			t.Errorf("regionCurrency(%q) = %q, want %q", region, got, want)
		}
	}
}

// The rule ladder of CurrencySplit, rule by rule, on one holding at a time:
// each rule must win over the ones below it. Rule 1's shortfall is the subtle
// one (a currency exposure summing to less than 100 % is real-asset capital,
// not missing data), and rule 6 is the last resort that makes a bond fund
// without a geography split still count somewhere.
func TestCurrencySplitRuleLadder(t *testing.T) {
	cases := []struct {
		name string
		m    Meta
		want map[string]float64
	}{
		{"1. explicit exposure, with a real-asset shortfall",
			Meta{AssetClass: "multi-asset", Currency: "EUR",
				CurrencyExposure: map[string]float64{"USD": 60, "EUR": 20}},
			map[string]float64{"USD": 0.6, "EUR": 0.2, CurrencyNone: 0.2}},
		{"2. gold is not a currency, whatever it is quoted in",
			Meta{AssetClass: "gold", Currency: "USD", Geography: map[string]float64{"US": 100}},
			map[string]float64{CurrencyNone: 1}},
		{"2. broad commodities likewise",
			Meta{AssetClass: "broad-commodity", Currency: "USD"},
			map[string]float64{CurrencyNone: 1}},
		{"3. a hedged share class is its hedge target",
			Meta{AssetClass: "equity", Currency: "USD", CurrencyHedged: true, HedgedTo: "EUR",
				Geography: map[string]float64{"US": 100}},
			map[string]float64{"EUR": 1}},
		{"4. a futures book sets its own exposure",
			Meta{AssetClass: "managed-futures", Currency: "USD"},
			map[string]float64{CurrencyDynamic: 1}},
		{"5. geography drives it, eurozone members merging",
			Meta{AssetClass: "equity", Currency: "EUR",
				Geography: map[string]float64{"United States": 50, "France": 30, "Germany": 10, "Other developed": 10}},
			map[string]float64{"USD": 0.5, "EUR": 0.4, CurrencyOther: 0.1}},
		{"6. last resort: the quote currency",
			Meta{AssetClass: "money-market", Currency: "EUR"},
			map[string]float64{"EUR": 1}},
		{"nothing at all is unknown, never a guess",
			Meta{AssetClass: "corporate-bond"},
			map[string]float64{BucketUnknown: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CurrencySplit([]Holding{{ID: "X", Weight: 1, HasMeta: true, Meta: c.m}})
			if len(got) != len(c.want) {
				t.Fatalf("split = %v, want %v", got, c.want)
			}
			for cur, w := range c.want {
				if math.Abs(got[cur]-w) > 1e-9 {
					t.Errorf("split[%q] = %v, want %v (full split %v)", cur, got[cur], w, got)
				}
			}
		})
	}
}

// A zero-weight line (a holding parked at 0 % in a file being edited) must
// leave every split untouched: it owns no capital, so it owns no share of
// anything, and a holding with no metadata is never guessed at.
func TestSplitsIgnoreZeroWeightAndUnknownMeta(t *testing.T) {
	holdings := []Holding{
		{ID: "PARKED", Weight: 0, HasMeta: true, Meta: Meta{AssetClass: "equity", Currency: "USD"}},
		{ID: "REAL", Weight: 0.6, HasMeta: true, Meta: Meta{AssetClass: "equity", Currency: "EUR",
			Geography: map[string]float64{"France": 100}, Sectors: map[string]float64{"Technology": 100}}},
		{ID: "MYSTERY", Weight: 0.4},
	}
	if got := CurrencySplit(holdings); math.Abs(got["EUR"]-0.6) > 1e-9 || math.Abs(got[BucketUnknown]-0.4) > 1e-9 || len(got) != 2 {
		t.Errorf("CurrencySplit = %v, want EUR 0.6 / Unknown 0.4", got)
	}
	if got := AssetClassSplit(holdings); math.Abs(got["equity"]-0.6) > 1e-9 || math.Abs(got[BucketUnknown]-0.4) > 1e-9 {
		t.Errorf("AssetClassSplit = %v, want equity 0.6 / Unknown 0.4", got)
	}
	if got := GeographySplit(holdings); math.Abs(got["France"]-0.6) > 1e-9 || math.Abs(got[BucketUnknown]-0.4) > 1e-9 {
		t.Errorf("GeographySplit = %v, want France 0.6 / Unknown 0.4", got)
	}
	// The sector pie covers the equity sleeve only, so it renormalizes over
	// the equity notional it could actually read.
	sectors, sleeve := EquitySectorSplit(holdings)
	if math.Abs(sectors["Information Technology"]-1) > 1e-9 || math.Abs(sleeve-0.6) > 1e-9 {
		t.Errorf("EquitySectorSplit = %v over a %v sleeve, want the whole readable sleeve in tech at 0.6", sectors, sleeve)
	}
	if got := Contributors(holdings, RegimeFramework()); len(got[Growth]) != 1 || got[Growth][0].ID != "REAL" {
		t.Errorf("Contributors[growth] = %+v, want the one funded equity line", got[Growth])
	}
}
