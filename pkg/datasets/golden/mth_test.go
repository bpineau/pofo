package golden

import (
	"math"
	"testing"
)

// Amundi's September 2026 factsheet publishes MTH's calendar returns before
// its selected FT line starts in September 2018. These independent years
// guard the donor era: a curve-only regression misses by several points.
// Reference: https://www.amundietf.lu/pdfDocuments/monthly-factsheet/LU1686832194/ENG/LUX/INSTITUTIONNEL/ETF
// The two indices have different country weights and costs, so this is a
// proxy sanity band, not the much narrower tolerance for the same NAV.
func TestGoldenMTHPreQuotePublishedReturns(t *testing.T) {
	s := loadSimdata(t, "MTH")
	for _, v := range []struct {
		year                 int
		published, tolerance float64
	}{{2016, 9.46, 1.5}, {2017, -1.72, 1.0}} {
		got := calendarYear(t, s, v.year)
		if math.Abs(got-v.published) > v.tolerance {
			t.Errorf("%d: MTH proxy %.2f %%, Amundi published %.2f %%, tolerance %.2f points", v.year, got, v.published, v.tolerance)
		}
	}
}
