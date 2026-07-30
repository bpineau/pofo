package suggest_test

import (
	"fmt"

	"github.com/bpineau/portfodor/pkg/suggest"
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
	cov, _ := suggest.Coverage(holdings)
	fmt.Printf("growth %.0f%%  inflation %.0f%%  crisis %.0f%%  deflation %.0f%%\n",
		cov[suggest.Growth]*100, cov[suggest.Inflation]*100, cov[suggest.Crisis]*100, cov[suggest.Deflation]*100)

	gaps := suggest.Gaps(cov, 0.10)
	fmt.Println("gaps:", gaps)
	// Output:
	// growth 60%  inflation 40%  crisis 40%  deflation 0%
	// gaps: [deflation]
}
