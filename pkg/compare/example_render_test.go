package compare_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/compare"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
)

// Example_render goes from specs to pictures: Compute runs the CLI's whole
// comparison pipeline, Studies hands the numbers behind every chart, chart.Line
// draws one of them as a standalone SVG, and report.Render writes the full
// HTML report the CLI opens. The two indices are bundled, so it runs offline;
// it prints shapes, not values, which move with every data refresh.
func Example_render() {
	client := marketdata.NewClient("") // "" = no disk cache
	spec, _ := portfolio.NewSpec("world and US",
		portfolio.Line{ID: "MSCIWORLD", Weight: 0.6},
		portfolio.Line{ID: "SP500", Weight: 0.4})
	cmp, err := compare.Compute(context.Background(), client, []*portfolio.Spec{spec}, compare.Options{
		Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		panic(err)
	}

	st := cmp.Studies()[0]
	svg := chart.Line(chart.Options{Title: st.Spec.Name, Width: 800, Height: 400}, []chart.Series{
		{Name: st.Spec.Name, Dates: st.Sim.Dates, Values: st.Sim.Index},
	})

	var page strings.Builder
	if err := report.Render(&page, cmp.HTMLPage(compare.Decoration{})); err != nil {
		panic(err)
	}
	fmt.Println(len(st.Holdings), "holdings,", len(st.Correlation), "x", len(st.Correlation[0]), "correlation")
	fmt.Println(strings.HasPrefix(svg, "<svg"), strings.Contains(page.String(), "</html>"))
	// Output:
	// 2 holdings, 2 x 2 correlation
	// true true
}
