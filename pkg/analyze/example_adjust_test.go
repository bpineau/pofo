package analyze_test

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// Series accepts valuations owned by the caller; the price-level index is a
// separately chosen consumption basket, not inferred from a currency code.
func ExampleSeries() {
	dates := []time.Time{
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	asset, err := marketdata.NewSeries("valuation", dates, []float64{100, 110, 121})
	if err != nil {
		panic(err)
	}
	cpi, err := marketdata.NewSeries("basket", dates, []float64{100, 105, 110.25})
	if err != nil {
		panic(err)
	}
	cash, err := marketdata.NewSeries("cash", dates, []float64{1, 1.02, 1.0404})
	if err != nil {
		panic(err)
	}
	asset.Currency, cash.Currency = "EUR", "EUR"
	study, err := analyze.Series(asset, nil, analyze.Options{Inflation: cpi, RiskFreeIndex: cash})
	if err != nil {
		panic(err)
	}
	fmt.Printf("nominal %.2f, purchasing power %.2f, excess growth %.4f\n",
		study.Series.Last().Close, study.Real.Series.Last().Close, study.Excess.Series.Last().Close)
	// Output: nominal 121.00, purchasing power 109.75, excess growth 1.1664
}
