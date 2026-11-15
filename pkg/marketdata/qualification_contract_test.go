package marketdata

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMixedReconstructionHasNoObservedTail(t *testing.T) {
	for _, fullWeight := range []float64{0, .25, -.25, 1} {
		a := &Series{Symbol: "OBS", Currency: "EUR", SimulatedBefore: d(2020, 1, 7), Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 102}, {d(2020, 1, 8), 101}}}
		b := a.clone()
		b.Symbol, b.FullySimulated, b.SimulatedBefore = "SYNTH", true, time.Time{}
		p, err := NewPanel(Daily, a, b)
		if err != nil {
			t.Fatal(err)
		}
		p, err = p.Mix("MIX", map[string]float64{"OBS": 1 - fullWeight, "SYNTH": fullWeight})
		if err != nil {
			t.Fatal(err)
		}
		mix, err := p.Series("MIX")
		if err != nil {
			t.Fatal(err)
		}
		if mix.FullySimulated != (fullWeight != 0) || fullWeight != 0 && !mix.SimulatedBefore.IsZero() {
			t.Fatalf("weight %g: fully=%v frontier=%v; every nonzero synthetic leg prevents an observed tail", fullWeight, mix.FullySimulated, mix.SimulatedBefore)
		}
		var csv bytes.Buffer
		if err := WriteCSV(&csv, mix); err != nil {
			t.Fatal(err)
		}
		round, err := ReadLongCSV(&csv)
		if err != nil {
			t.Fatal(err)
		}
		if round[0].FullySimulated != mix.FullySimulated || !round[0].SimulatedBefore.Equal(mix.SimulatedBefore) {
			t.Fatal("CSV lost provenance")
		}
		if a.FullySimulated || !b.SimulatedBefore.IsZero() {
			t.Fatal("mutated an input")
		}
	}
}

func TestAdjustmentsRetainIndexQualifications(t *testing.T) {
	a := &Series{Symbol: "ASSET", Currency: "EUR", Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 102}, {d(2020, 1, 8), 101}}}
	index := a.clone()
	index.Symbol, index.FullySimulated = "INDEX", true
	for i := range index.Points {
		index.Points[i].Close = 100
	}
	for name, transform := range map[string]func(*Series, AdjustmentOptions) (*Series, error){"deflate": a.Deflate, "excess": a.Excess} {
		out, err := transform(index, AdjustmentOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !out.FullySimulated || !strings.Contains(strings.Join(out.Warnings, " "), "reconstructed") {
			t.Errorf("%s lost the reconstructed index: fully=%v warnings=%v", name, out.FullySimulated, out.Warnings)
		}
	}
	if a.FullySimulated || len(index.Warnings) != 0 {
		t.Fatal("mutated source metadata")
	}
}

func TestCashIndexRetainsRateQualificationsThroughExcess(t *testing.T) {
	for _, fully := range []bool{false, true} {
		rates := &Series{Symbol: "RATE", Currency: "EUR", FullySimulated: fully,
			EstimatedFrom: d(2020, 1, 8), FXExtrapolatedBefore: d(2020, 1, 7),
			Points: []Point{{d(2020, 1, 6), 2}, {d(2020, 1, 8), 2}},
		}
		if !fully {
			rates.SimulatedBefore = d(2020, 1, 7)
		}
		asset := &Series{Symbol: "ASSET", Currency: "EUR", Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 102}, {d(2020, 1, 8), 101}}}
		cash, err := CashIndex(rates, []time.Time{d(2020, 1, 6), d(2020, 1, 7), d(2020, 1, 8)}, Actual365)
		if err != nil {
			t.Fatal(err)
		}
		excess, err := asset.Excess(cash, AdjustmentOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for _, out := range []*Series{cash, excess} {
			if out.FullySimulated != fully || !out.SimulatedBefore.Equal(rates.SimulatedBefore) || !out.EstimatedFrom.Equal(rates.EstimatedFrom) || !out.FXExtrapolatedBefore.Equal(rates.FXExtrapolatedBefore) {
				t.Errorf("%s dropped rate qualifications: fully=%v before=%v estimate=%v FX=%v", out.Symbol, out.FullySimulated, out.SimulatedBefore, out.EstimatedFrom, out.FXExtrapolatedBefore)
			}
			if len(out.Warnings) < len(rates.DataWarnings()) {
				t.Errorf("%s lost implicit rate warnings: %v", out.Symbol, out.Warnings)
			}
		}
		if len(rates.Warnings) != 0 || asset.FullySimulated || !asset.SimulatedBefore.IsZero() {
			t.Fatal("mutated an input")
		}
	}
}
