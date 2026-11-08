package simgen

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// A second charge on the observed donor, or a curve kept over its quote era,
// changes these literal donor levels and the pre-donor fee-only growth.
func TestMTHUsesObservedDonorWithoutChargingItAgain(t *testing.T) {
	curve := &marketdata.Series{Symbol: "EUROGOV-LONG-EUR", Points: []marketdata.Point{{Date: day(0), Close: 100}, {Date: day(365), Close: 100}, {Date: day(730), Close: 100}, {Date: day(1095), Close: 100}}}
	donor := &marketdata.Series{Symbol: "DBXG", Source: "ft", Currency: "EUR", Points: []marketdata.Point{{Date: day(365), Close: 20}, {Date: day(730), Close: 25}, {Date: day(1095), Close: 22}}, Warnings: []string{"Invented donor qualification"}}
	f := fakeFetcher{"EUROGOV-LONG-EUR": curve, "DBXG": donor}
	got, err := mthRecipe().Build(f, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Points) != 4 {
		t.Fatalf("got %d points, want curve prefix plus donor", len(got.Points))
	}
	if !slices.Equal(got.Points[1:], donor.Points) {
		t.Errorf("donor levels changed: %v", got.Points[1:])
	}
	// The flat curve loses precisely 0.07% in a 365.25-day year. The first
	// interval is 365 days, so its loss differs by less than 0.000001.
	if r := got.Points[1].Close/got.Points[0].Close - 1; math.Abs(r-(-0.0007)) > 0.000001 {
		t.Errorf("deep curve return %.9f, want -0.0007", r)
	}
	if !got.FullySimulated || !got.SimulatedBefore.IsZero() {
		t.Errorf("donor quotes claimed to be observed MTH: fully=%v frontier=%v", got.FullySimulated, got.SimulatedBefore)
	}
	if got.Currency != "EUR" || got.Source != "simdata" {
		t.Errorf("unexpected result identity: %s %s", got.Currency, got.Source)
	}
	real := &marketdata.Series{Symbol: "MTH", Currency: "EUR", Source: "ft", Points: []marketdata.Point{{Date: day(730), Close: 100}, {Date: day(1095), Close: 90}}}
	shipped := Splice(real, got)
	if !slices.Equal(shipped.Points[len(shipped.Points)-2:], real.Points) {
		t.Errorf("observed target changed: %v", shipped.Points)
	}
	if shipped.FullySimulated || !shipped.SimulatedBefore.Equal(day(730)) {
		t.Errorf("target frontier lost: fully=%v frontier=%v", shipped.FullySimulated, shipped.SimulatedBefore)
	}
	if !slices.ContainsFunc(shipped.DataWarnings(), func(w string) bool { return strings.Contains(w, "Invented donor qualification") }) {
		t.Errorf("donor qualification lost: %v", shipped.DataWarnings())
	}
	if len(donor.Points) != 3 || donor.FullySimulated {
		t.Fatal("input donor mutated")
	}
}

// Missing or reconstructed donors must fail rather than silently shipping
// the old curve while the method claims an observed DBXG donor.
func TestMTHRejectsUnavailableOrUnobservedDonor(t *testing.T) {
	curve := mkLevels("EUROGOV-LONG-EUR", 10, 100)
	for _, tc := range []struct {
		name  string
		donor *marketdata.Series
	}{
		{"missing", nil},
		{"empty", &marketdata.Series{Source: "ft", Currency: "EUR"}},
		{"single", &marketdata.Series{Source: "ft", Currency: "EUR", Points: []marketdata.Point{{Date: day(0), Close: 20}}}},
		{"simdata", &marketdata.Series{Source: "simdata", Currency: "EUR", Points: curve.Points}},
		{"index", &marketdata.Series{Source: "index", Currency: "EUR", Points: curve.Points}},
		{"refdata", &marketdata.Series{Source: "refdata", Currency: "EUR", Points: curve.Points}},
		{"known-prefix", &marketdata.Series{Source: "ft", Currency: "EUR", SimulatedBefore: day(2), Points: curve.Points}},
		{"fully-simulated", &marketdata.Series{Source: "ft", Currency: "EUR", FullySimulated: true, Points: curve.Points}},
		{"wrong-currency", &marketdata.Series{Source: "ft", Currency: "USD", Points: curve.Points}},
		{"unknown-currency", &marketdata.Series{Source: "ft", Points: curve.Points}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := fakeFetcher{"EUROGOV-LONG-EUR": curve}
			if tc.donor != nil {
				f["DBXG"] = tc.donor
			}
			if _, err := mthBuild(f, time.Time{}); err == nil || !strings.Contains(err.Error(), "DBXG") {
				t.Errorf("want explicit donor failure, got %v", err)
			}
		})
	}
}

func TestMTHOmitsEstimatedDonorTail(t *testing.T) {
	donor := mkLevels("DBXG", 10, 20)
	donor.Currency = "EUR"
	donor.Source = "ft"
	donor.EstimatedFrom = day(5)
	f := fakeFetcher{"EUROGOV-LONG-EUR": mkLevels("EUROGOV-LONG-EUR", 10, 100), "DBXG": donor}
	got, err := mthBuild(f, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Last().Date != day(4) {
		t.Errorf("estimated donor points shipped through %v", got.Last().Date)
	}
}

func TestMTHAuditGradesTheDonorAndDeepCurveSeparately(t *testing.T) {
	curve := mkWave("EUROGOV-LONG-EUR", 1600, 0.0001, 0.01, 0.7, 0.2)
	donor := from(curve, 300)
	donor.Symbol = "DBXG"
	donor.Source = "ft"
	donor.Currency = "EUR"
	real := from(mkWave("MTH", 1600, 0.00011, 0.0097, 0.71, 0.3), 700)
	real.Symbol = "MTH"
	real.Source = "ft"
	real.Currency = "EUR"
	got := Audit(fakeFetcher{"EUROGOV-LONG-EUR": curve, "DBXG": donor, "MTH": real}, mthRecipe())
	if got.Err != "" {
		t.Fatal(got.Err)
	}
	if len(got.Chain) != 2 {
		t.Fatalf("got %d junctions, want both donor links", len(got.Chain))
	}
	for i, pair := range []string{"DBXG vs MTH", "EUROGOV-LONG-EUR vs DBXG"} {
		if got.Chain[i].Pair != pair || !got.Chain[i].Measured {
			t.Errorf("junction %d: %+v", i, got.Chain[i])
		}
	}
}
