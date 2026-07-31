package permanent_test

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/permanent"
)

// A regime halfway to paradise (60% of countries growing, 40% inflating) gets a
// moderate equity tilt under the default quadratic damping.
func ExampleRegime_EquityWeight() {
	r := permanent.Regime{GrowthBreadth: 0.6, InflationBreadth: 0.4}
	fmt.Printf("%.0f%%", r.EquityWeight(permanent.DefaultParams())*100)
	// Output: 47%
}

// Allocate turns a regime into a full four-sleeve target that sums to 1.
func ExampleRegime_Allocate() {
	paradise := permanent.Regime{GrowthBreadth: 1, InflationBreadth: 0, Slope: 2, RealShort: 1}
	a := paradise.Allocate(permanent.DefaultParams())
	fmt.Printf("equity=%.0f%% (defensive sleeve empty in full paradise)", a.Equity*100)
	// Output: equity=100% (defensive sleeve empty in full paradise)
}

// LoadPanel reads the embedded OECD macro panel; the regime is a smoothed
// breadth reading over its countries.
func ExampleLoadPanel() {
	p, err := permanent.LoadPanel()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d countries", len(p.Countries()))
	// Output: 30 countries
}
