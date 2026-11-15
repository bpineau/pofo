package fire

import (
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestMonthlyDefinitionChangesSplitHistory(t *testing.T) {
	var points []marketdata.Point
	for i := range 9 {
		value := float64(100 + i)
		if i >= 4 {
			value *= 10
		}
		points = append(points, marketdata.Point{Date: mo(2020, time.Month(i+1)), Close: value})
	}
	for _, junction := range []time.Time{points[4].Date, points[3].Date.AddDate(0, 0, 1)} {
		assets := []AssetSeries{{Weight: .5, Points: points, Junctions: []time.Time{junction}}, {Weight: .5, Points: points, Junctions: []time.Time{junction}}}
		panel, coverage, err := BuildMonthlyPanel(assets, []marketdata.Point{{Date: mo(2019, 12), Close: 100}})
		if err != nil {
			t.Fatal(err)
		}
		if panel.Periods() != 4 || coverage.JunctionMonths != 1 || coverage.GapMonths != 3 || !coverage.First.Equal(time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("wrong exclusion: %d periods, %+v", panel.Periods(), coverage)
		}
		for _, row := range panel.Returns {
			for _, r := range row {
				if r > .02 {
					t.Fatalf("definition jump survived: %g", r)
				}
			}
		}
	}
}
