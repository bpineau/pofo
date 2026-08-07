package simgen

import (
	"math"
	"testing"
	"time"
)

// TestFeeAligned checks the three rules the uplift table rests on: a donor
// costlier than its target is lifted by the difference, a cheaper one is not
// pushed down, and the target's own performance fee is subtracted (which is
// what keeps the AQR chain, whose donor is cheaper on base fees than the
// target is all in, at zero).
func TestFeeAligned(t *testing.T) {
	got := feeAligned("DBMF", []string{"ASFYX", "RYMFX", ahlDiversified})
	want := []Donor{
		{ID: "ASFYX", Uplift: 0.0145 - 0.0085},
		{ID: "RYMFX", Uplift: 0.0199 - 0.0085},
		{ID: ahlDiversified, Uplift: 0.0274 - 0.0085},
	}
	for i, w := range want {
		if got[i].ID != w.ID || math.Abs(got[i].Uplift-w.Uplift) > 1e-12 {
			t.Errorf("donor %d = %+v, want %+v", i, got[i], w)
		}
	}

	// The UCITS class is 0.10 points cheaper than the US ETF that donates to
	// it: same manager, same book, a different price list and nothing else.
	if u := feeAligned("LU2951555585", []string{"DBMF"})[0].Uplift; math.Abs(u-0.0010) > 1e-12 {
		t.Errorf("DBMF into the UCITS class = %v, want 0.0010", u)
	}

	// AQMIX is 0.50 points dearer than AQR UCITS class A on base fees, but
	// class A also pays a performance fee AQMIX does not, worth more than the
	// difference: the aligned uplift is zero, never negative.
	if u := feeAligned("LU1103257975", []string{"AQMIX"})[0].Uplift; u != 0 {
		t.Errorf("AQMIX into AQR UCITS class A = %v, want 0", u)
	}
}

// TestFeeAlignedUnknownVehicle: a donor nobody looked up a price for must be
// loud, not silently aligned against a zero fee load.
func TestFeeAlignedUnknownVehicle(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("feeAligned accepted an undocumented donor")
		}
	}()
	feeAligned("DBMF", []string{"NOSUCHFUND"})
}

// TestDonorChainUplift: the uplift is a constant annual addition to that
// donor's segment and to nothing else. The chain here is one donor as
// volatile as the fund it stands in for, so the volatility match is a no-op
// and the two chains differ by exactly one compounding of the uplift.
func TestDonorChainUplift(t *testing.T) {
	const n = 800
	f := fakeFetcher{
		"FUND":  mkWave("FUND", n, 3e-4, 0.010, 1.0, 0.2),
		"DONOR": mkWave("DONOR", n, 1e-4, 0.010, 1.0, 0.2),
		"^IRX":  mkLevels("^IRX", n, 2.0),
	}
	cagr := func(uplift float64) float64 {
		s, err := DonorChain(f, "^IRX", "FUND", []Donor{{ID: "DONOR", Uplift: uplift}}, time.Time{}, nil)
		if err != nil {
			t.Fatalf("DonorChain: %v", err)
		}
		years := s.Last().Date.Sub(s.First().Date).Hours() / 24 / 365.25
		return math.Pow(s.Last().Close/s.First().Close, 1/years) - 1
	}
	near(t, "uplift", (1+cagr(0.01))/(1+cagr(0))-1, 0.01, 1e-6)
}
