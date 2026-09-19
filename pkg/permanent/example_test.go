package permanent_test

import (
	"fmt"
	"time"

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

// PublicationLags reads every driver as of the last month its publisher had
// actually released, which is what a real-time allocator had. October 2008 is
// the difference in one month: the reference-date reading of the panel already
// shows inflation decelerating across the world, while the numbers on the wire
// that month still said stagflation.
func ExamplePublicationLags() {
	p, err := permanent.LoadPanel()
	if err != nil {
		panic(err)
	}
	m := time.Date(2008, 10, 1, 0, 0, 0, 0, time.UTC)
	cfg := permanent.DefaultSignalConfig()
	ref, _ := p.RegimeAt(m, cfg)
	cfg.ReleaseLags = permanent.PublicationLags()
	honest, _ := p.RegimeAt(m, cfg)
	fmt.Printf("as of the month: growth %.2f, inflation %.2f (%v)\n",
		ref.GrowthBreadth, ref.InflationBreadth, ref.Quadrant())
	fmt.Printf("as published:    growth %.2f, inflation %.2f (%v)\n",
		honest.GrowthBreadth, honest.InflationBreadth, honest.Quadrant())
	// Output:
	// as of the month: growth 0.15, inflation 0.49 (deflation)
	// as published:    growth 0.21, inflation 0.76 (crisis)
}

// Quadrant is the coarse four-season reading of a regime, for labels and
// per-regime aggregations; the allocator itself stays continuous.
func ExampleRegime_Quadrant() {
	fmt.Println(permanent.Regime{GrowthBreadth: 0.7, InflationBreadth: 0.2}.Quadrant())
	fmt.Println(permanent.Regime{GrowthBreadth: 0.3, InflationBreadth: 0.8}.Quadrant())
	// Output:
	// growth
	// crisis
}
