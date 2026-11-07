package marketdata_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func ExampleCashIndex() {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 360)
	rates, err := marketdata.NewSeries("deposit", []time.Time{start, end}, []float64{3.6, 3.6})
	if err != nil {
		panic(err)
	}
	rates.Currency = "EUR"
	cash, err := marketdata.CashIndex(rates, []time.Time{start, end}, marketdata.Actual360)
	if err != nil {
		panic(err)
	}
	fmt.Printf("360 calendar days: %.4f %%\n", 100*(cash.Last().Close-1))
	// Output: 360 calendar days: 3.6654 %
}

func ExampleReadSeriesCSV() {
	input := "# FUND currency: EUR\nid,date,value\nFUND,2024-01-31,100\nFUND,2024-02-29,102\n"
	series, err := marketdata.ReadSeriesCSV(strings.NewReader(input), "fallback")
	if err != nil {
		panic(err)
	}
	fmt.Println(series[0].Symbol, series[0].Currency, series[0].Last().Close)
	// Output: FUND EUR 102
}

func ExampleSeries_Deflate() {
	dates := []time.Time{
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
	}
	asset, err := marketdata.NewSeries("valuation", dates, []float64{100, 110})
	if err != nil {
		panic(err)
	}
	basket, err := marketdata.NewSeries("basket", dates, []float64{100, 105})
	if err != nil {
		panic(err)
	}
	real, err := asset.Deflate(basket, marketdata.AdjustmentOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("purchasing power: %.2f\n", real.Last().Close)
	// Output: purchasing power: 104.76
}
