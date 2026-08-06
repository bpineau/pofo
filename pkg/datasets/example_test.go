package datasets_test

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Catalog returns the bundled assets as typed datasets.Asset records (use
// AssetMeta for the same data as raw JSON). Resolve a ticker/alias/ISIN to a
// single record with marketdata.Lookup.
func ExampleCatalog() {
	for _, a := range datasets.Catalog() {
		if a.ID == "IE00B4L5Y983" { // iShares Core MSCI World (IWDA)
			fmt.Printf("%s: TER %.2f%%, UCITS=%v, US=%g%%\n",
				a.Name, a.Fees, a.UCITS, a.Geography["US"])
		}
	}
	// Output: iShares Core MSCI World UCITS ETF USD (Acc): TER 0.20%, UCITS=true, US=68%
}
