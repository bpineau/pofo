package golden

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// These golden tests pin the bundled AQR Managed Futures series to an AUDITED
// external reference: the net asset value per share published, per share class
// and per financial year end (31 March), in the AQR UCITS Funds annual report
// audited by PricewaterhouseCoopers. Unlike an index reconstruction, a fund NAV
// has one true value per day, so the anchors are asserted to the cent.
//
// Why this matters beyond arithmetic: a managed-futures share class is exactly
// the kind of identifier a fuzzy provider search silently mismatches (the fund
// has more than forty share classes, several of them EUR-hedged, differing only
// by a fee schedule). Matching three consecutive audited year ends to the cent
// proves the resolution AND the fee schedule of the class we actually fetch.
//
// Anchors (audited annual report for the year ended 31 March 2026, "Statistical
// Information", cross-checked against the prior year's figures in the same
// table). 31 March 2024 fell on Easter Sunday, so the NAV struck for that year
// end is the Thursday 2024-03-28 close; the report dates it 31 March 2024.
//
// The two bundled series carry real quotes over this window (the simdata
// recipes splice the real class NAVs in front of their reconstruction), so the
// assertion also guards `make simdata` regenerations.
type navAnchor struct {
	date string  // the trading day the year-end NAV was struck on
	nav  float64 // audited net asset value per share, in the class currency
}

// TestGoldenAQRManagedFutures checks the EUR-hedged flat-fee class (RAEF) and
// the USD reference class (A) against their audited year-end NAVs.
func TestGoldenAQRManagedFutures(t *testing.T) {
	for _, c := range []struct {
		id, label string
		anchors   []navAnchor
	}{
		{
			id:    "LU1662501532",
			label: "RAEF EUR (hedged, flat fee)",
			anchors: []navAnchor{
				{"2024-03-28", 135.54},
				{"2025-03-31", 138.79},
				{"2026-03-31", 160.34},
			},
		},
		{
			id:    "LU1103257975",
			label: "A USD (reference class, 10% performance fee)",
			anchors: []navAnchor{
				{"2024-03-28", 137.52},
				{"2025-03-31", 142.24},
				{"2026-03-31", 165.34},
			},
		},
	} {
		t.Run(c.id, func(t *testing.T) {
			s, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), c.id)
			if err != nil || !ok {
				t.Fatalf("simdata %s: ok=%v err=%v", c.id, ok, err)
			}
			for _, a := range c.anchors {
				got, found := closeOn(s, a.date)
				if !found {
					t.Errorf("%s (%s): no quote on %s", c.id, c.label, a.date)
					continue
				}
				// One cent: the report and the provider publish the same
				// two-decimal NAV, so any drift is a real mismatch.
				if math.Abs(got-a.nav) > 0.005 {
					t.Errorf("%s (%s) %s: NAV %.2f, audited %.2f",
						c.id, c.label, a.date, got, a.nav)
				}
			}
		})
	}
}

// TestGoldenAQRManagedFuturesFeeOrder checks the one conclusion the audited
// accounts let us draw about the EUR-hedged classes, and that the catalog
// records: the flat-fee RAEF has out-returned the institutional IAE2F
// (LU1662499612, a LOWER headline management fee, 0.75% against 1.00%) in
// every financial year they both ran. The audited year-end NAVs make the
// comparison exact and currency-free (both classes are fully EUR-hedged, so
// the gap is cost, not FX).
//
// The gap has been remarkably steady, 0.6 to 0.7 points a year, which is why
// the catalog steers to RAEF: nothing in the published fee schedules predicts
// it, so it must be measured rather than derived. The test guards the sign and
// a generous band, not the exact figure.
func TestGoldenAQRManagedFuturesFeeOrder(t *testing.T) {
	raef, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), "LU1662501532")
	if err != nil || !ok {
		t.Fatalf("simdata RAEF: ok=%v err=%v", ok, err)
	}
	// IAE2F is not bundled as a simdata series (it has a three-year NAV hole),
	// so its audited year ends stand in for it directly.
	iae2f := map[string]float64{"2024-03-28": 115.65, "2025-03-31": 117.65, "2026-03-31": 135.05}

	for _, fy := range []struct{ from, to string }{
		{"2024-03-28", "2025-03-31"},
		{"2025-03-31", "2026-03-31"},
	} {
		a, okA := closeOn(raef, fy.from)
		b, okB := closeOn(raef, fy.to)
		if !okA || !okB {
			t.Fatalf("RAEF: missing quote on %s or %s", fy.from, fy.to)
		}
		gap := 100 * ((b/a - 1) - (iae2f[fy.to]/iae2f[fy.from] - 1))
		if gap < 0.3 || gap > 1.2 {
			t.Errorf("FY %s..%s: RAEF beat IAE2F by %+.2f pts, expected +0.3 to +1.2",
				fy.from, fy.to, gap)
		}
	}
}

// closeOn returns the close of the point dated exactly day ("2006-01-02").
func closeOn(s *marketdata.Series, day string) (float64, bool) {
	d, err := time.Parse("2006-01-02", day)
	if err != nil {
		return 0, false
	}
	for _, p := range s.Points {
		if p.Date.Equal(d) {
			return p.Close, true
		}
	}
	return 0, false
}
