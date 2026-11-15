package decumul

import (
	"encoding/json"
	"math"
	"testing"
)

func TestSpendUndefined(t *testing.T) {
	empty := Ensemble{}
	s := empty.SpendStats()
	for _, v := range []float64{s.EverCutShare, s.FirstCutMedian, s.CutYearsMedian, s.CutYearsP90, empty.SpendCV()} {
		if !math.IsNaN(v) {
			t.Errorf("empty estimate %v, want NaN", v)
		}
	}
	e := Ensemble{Years: 1, Paths: []PathResult{{Wealth: []float64{100, 100}, Spend: []float64{0}, FirstCut: -1}}}
	s = e.SpendStats()
	if s.EverCutShare != 0 || !math.IsNaN(s.FirstCutMedian) || !math.IsNaN(s.CutYearsMedian) || !math.IsNaN(s.CutYearsP90) || !math.IsNaN(e.SpendCV()) {
		t.Fatalf("no-cut/zero-spend estimates %+v CV %v", s, e.SpendCV())
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(b, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["FirstCutMedian"] != nil {
		t.Fatalf("no-cut JSON %s", b)
	}
}
