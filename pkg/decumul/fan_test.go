package decumul

import "testing"

func TestFanBandsOrderedAndAnchored(t *testing.T) {
	e := Ensemble{Years: 2, Paths: []PathResult{
		{Wealth: []float64{100, 110, 120}},
		{Wealth: []float64{100, 90, 80}, Ruined: true},
		{Wealth: []float64{100, 100, 100}},
		{Wealth: []float64{100, 130, 150}},
	}}

	fan := e.Fan([]float64{0.05, 0.50, 0.95}, 2)

	if fan.Years != 2 {
		t.Fatalf("Years = %d, want 2", fan.Years)
	}
	if len(fan.Bands) != 3 {
		t.Fatalf("Bands rows = %d, want 3 (one per percentile)", len(fan.Bands))
	}
	for p, row := range fan.Bands {
		if len(row) != 3 {
			t.Fatalf("band %d length = %d, want 3 (Years+1)", p, len(row))
		}
		if row[0] != 100 {
			t.Errorf("band %d at year 0 = %.0f, want 100 (all paths start at capital)", p, row[0])
		}
	}
	// At each year the bands must be ordered low to high.
	for y := 0; y <= 2; y++ {
		if !(fan.Bands[0][y] <= fan.Bands[1][y] && fan.Bands[1][y] <= fan.Bands[2][y]) {
			t.Errorf("year %d bands not ordered: %.0f, %.0f, %.0f", y, fan.Bands[0][y], fan.Bands[1][y], fan.Bands[2][y])
		}
	}
}

func TestFanSamplesSpanTheOutcomes(t *testing.T) {
	var paths []PathResult
	for i := range 50 {
		end := float64(i) * 10 // terminal wealth 0..490, the low ones ruined
		paths = append(paths, PathResult{Wealth: []float64{100, end / 2, end}, Ruined: end == 0})
	}
	e := Ensemble{Years: 2, Paths: paths}

	fan := e.Fan([]float64{0.5}, 4)

	if len(fan.Samples) != 4 {
		t.Fatalf("samples = %d, want 4", len(fan.Samples))
	}
	for _, s := range fan.Samples {
		if len(s.Wealth) != 3 {
			t.Errorf("sample path length = %d, want 3", len(s.Wealth))
		}
	}
	// Samples are drawn across the terminal distribution, so the lowest must be
	// well below the highest.
	lo, hi := fan.Samples[0].Wealth[2], fan.Samples[len(fan.Samples)-1].Wealth[2]
	if !(hi > lo) {
		t.Errorf("samples do not span outcomes: lo=%.0f hi=%.0f", lo, hi)
	}
}
