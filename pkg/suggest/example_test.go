package suggest_test

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/suggest"
)

// Regimes maps an asset's metadata to the macro regimes it helps in.
func ExampleRegimes() {
	fmt.Println(suggest.Regimes(suggest.Meta{AssetClass: "gold"}))
	fmt.Println(suggest.Regimes(suggest.Meta{AssetClass: "equity", Benchmark: "MSCI World"}))
	fmt.Println(suggest.Regimes(suggest.Meta{AssetClass: "managed-futures"}))
	// Output:
	// [inflation crisis]
	// [growth]
	// [crisis inflation]
}

// Coverage sums the weight of the holdings that help in each regime; an asset
// spanning several (gold helps in inflation and crisis) counts in each.
func ExampleCoverage() {
	holdings := []suggest.Holding{
		{Weight: 0.6, HasMeta: true, Meta: suggest.Meta{AssetClass: "equity"}},
		{Weight: 0.4, HasMeta: true, Meta: suggest.Meta{AssetClass: "gold"}},
	}
	cov, _ := suggest.Coverage(holdings, suggest.RegimeFramework())
	fmt.Printf("growth %.0f%%  inflation %.0f%%  crisis %.0f%%  deflation %.0f%%\n",
		cov[suggest.Growth]*100, cov[suggest.Inflation]*100, cov[suggest.Crisis]*100, cov[suggest.Deflation]*100)

	gaps := suggest.Gaps(cov, suggest.RegimeFramework(), 0.10)
	fmt.Println("gaps:", gaps)
	// Output:
	// growth 60%  inflation 40%  crisis 40%  deflation 0%
	// gaps: [deflation]
}

// AssetClassSplit opens stacked funds into their legs: a 50 % position in a
// 90/60 efficient-core fund counts as 45 points of equity plus 30 of bonds.
func ExampleAssetClassSplit() {
	holdings := []suggest.Holding{
		{Weight: 0.5, HasMeta: true, Meta: suggest.Meta{
			AssetClass: "multi-asset",
			Exposures:  map[string]float64{"equity": 0.9, "government-bond": 0.6},
		}},
		{Weight: 0.5, HasMeta: true, Meta: suggest.Meta{AssetClass: "gold"}},
	}
	split := suggest.AssetClassSplit(holdings)
	fmt.Printf("equity %.2f  government-bond %.2f  gold %.2f\n",
		split["equity"], split["government-bond"], split["gold"])
	// Output:
	// equity 0.45  government-bond 0.30  gold 0.50
}

// CurrencySplit reports look-through fiat exposure: quote currencies are
// ignored in favor of what the capital actually moves with, so a EUR-quoted
// world-equity fund shows up mostly as USD and a gold ETC as no fiat at all.
func ExampleCurrencySplit() {
	holdings := []suggest.Holding{
		{ID: "WORLD", Weight: 0.6, HasMeta: true, Meta: suggest.Meta{
			AssetClass: "equity", Currency: "EUR",
			Geography: map[string]float64{"US": 70, "Japan": 10, "Other": 20},
		}},
		{ID: "GOLD", Weight: 0.4, HasMeta: true, Meta: suggest.Meta{AssetClass: "gold", Currency: "USD"}},
	}
	split := suggest.CurrencySplit(holdings)
	fmt.Printf("USD %.2f  JPY %.2f  Other %.2f  None %.2f\n",
		split["USD"], split["JPY"], split[suggest.CurrencyOther], split[suggest.CurrencyNone])

	p := suggest.CurrencyProfile(split, "EUR")
	fmt.Printf("unhedged foreign %.0f%% (top %s %.0f%%), non-fiat %.0f%%\n",
		p.Foreign*100, p.Top, p.TopShare*100, p.NonFiat*100)
	// Output:
	// USD 0.42  JPY 0.06  Other 0.12  None 0.40
	// unhedged foreign 60% (top USD 42%), non-fiat 40%
}

// Contributors decomposes Coverage per holding: for each regime, who covers
// it and with how much notional weight (stacked funds count each leg's).
func ExampleContributors() {
	holdings := []suggest.Holding{
		{ID: "NTSX", Weight: 0.5, HasMeta: true, Meta: suggest.Meta{
			AssetClass: "multi-asset",
			Exposures:  map[string]float64{"equity": 0.9, "government-bond": 0.6},
		}},
		{ID: "GOLD", Weight: 0.5, HasMeta: true, Meta: suggest.Meta{AssetClass: "gold"}},
	}
	contrib := suggest.Contributors(holdings, suggest.RegimeFramework())
	for _, c := range contrib[suggest.Growth] {
		fmt.Printf("growth: %s %.0f%%\n", c.ID, c.Weight*100)
	}
	for _, c := range contrib[suggest.Crisis] {
		fmt.Printf("crisis: %s %.0f%%\n", c.ID, c.Weight*100)
	}
	// Output:
	// growth: NTSX 45%
	// crisis: GOLD 50%
}

// DurationSplit tallies look-through interest-rate duration: a stacked
// fund's bond leg counts its notional times the fund's Duration figure.
func ExampleDurationSplit() {
	holdings := []suggest.Holding{
		{ID: "NTSX", Weight: 0.4, HasMeta: true, Meta: suggest.Meta{
			AssetClass: "multi-asset",
			Exposures:  map[string]float64{"equity": 0.9, "government-bond": 0.6},
			Duration:   7,
		}},
		{ID: "LINKER", Weight: 0.2, HasMeta: true, Meta: suggest.Meta{
			AssetClass: "inflation-linked-bond", Duration: 8,
		}},
	}
	led := suggest.DurationSplit(holdings)
	fmt.Printf("nominal %.2f y, real %.2f y\n", led.Nominal, led.Real)
	// Output:
	// nominal 1.68 y, real 1.60 y
}
