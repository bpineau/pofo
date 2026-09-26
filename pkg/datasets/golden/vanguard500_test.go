package golden

import (
	"math"
	"testing"
)

// This golden test pins the bundled S&P 500 tracker (IE00BFMXXD54) over the
// decade its history IS the Vanguard 500 fund (VFINX, the donor the recipe
// reads from 1980, with nothing added: the tracker's 0.07 %/yr is under the
// donor's own charge) to that fund's PUBLISHED annual total returns.
//
// Reference: Vanguard Index Trust prospectus of 1995, Financial Highlights of
// the 500 Portfolio, ten fiscal years ended December 31 (SEC EDGAR, CIK 36405,
// accession 0000893220-95-000289). The figures exclude the $10 annual account
// fee, as a NAV series does.
//
// What it guards: the provider's adjusted close for this fund lost year-end
// capital-gain distributions over 1980-1986 (up to -6.97 % in one session with
// no index move), and pkg/simgen holds the donor to the S&P 500 to repair them
// (the `tracked` map). Unrepaired, 1985 reads 22.60 % and 1986 9.30 % here;
// a regression in that grading, or a provider line that changes under it,
// fails this test. 1980-1984 are not in the prospectus and are not pinned.
//
// 1992 and 1993 are asserted as a PAIR: the provider repeats the 1992-12-31
// close and books the move on 1993-01-04, which shifts about 0.8 point across
// the year boundary and leaves the level right by the end of 1993.
func TestGoldenVanguard500Published(t *testing.T) {
	s := loadSimdata(t, "IE00BFMXXD54")
	published := map[int]float64{
		1985: 31.23, 1986: 18.06, 1987: 4.71, 1988: 16.22, 1989: 31.36,
		1990: -3.32, 1991: 30.22, 1992: 7.42, 1993: 9.89, 1994: 1.18,
	}
	// A tenth of a point a year: the provider's NAV and the fund's report
	// agree to a few hundredths wherever the line is sound (measured: 0.05 at
	// worst), and the smallest defect this guards against is 1.2 points.
	const tol = 0.10
	for y := 1985; y <= 1994; y++ {
		if y == 1992 || y == 1993 {
			continue
		}
		if got := calendarYear(t, s, y); math.Abs(got-published[y]) > tol {
			t.Errorf("%d: %.2f %%, published %.2f %%", y, got, published[y])
		}
	}
	pair := func(a, b float64) float64 { return ((1+a/100)*(1+b/100) - 1) * 100 }
	got := pair(calendarYear(t, s, 1992), calendarYear(t, s, 1993))
	if want := pair(published[1992], published[1993]); math.Abs(got-want) > tol {
		t.Errorf("1992-1993: %.2f %%, published %.2f %%", got, want)
	}
}
