package analyze_test

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func TestUndefinedCorrelationSurvivesStudyJSON(t *testing.T) {
	a := walk(t, "A", 100)
	b := scaled(t, a, "B", 1)
	for i := range b.Points {
		b.Points[i].Close = 100
	}
	sp, err := portfolio.NewSpec("flat reference", portfolio.Line{ID: "A", Weight: .5}, portfolio.Line{ID: "B", Weight: .5})
	if err != nil {
		t.Fatal(err)
	}
	study, err := analyze.Portfolio(context.Background(), analyze.SeriesSource{"A": a, "B": b}, sp, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(study.Correlation[0][1]) {
		t.Fatal("undefined correlation became zero")
	}
	if !strings.Contains(strings.Join(study.Warnings, " "), "undefined correlation") {
		t.Fatal("missing constant-history warning")
	}
	data, err := json.Marshal(study)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct{ Correlation [][]*float64 }
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Correlation[0][1] != nil {
		t.Fatal("undefined correlation did not serialize as null")
	}
}

func TestPairTrackingJSONKeepsCalendarAndDivergences(t *testing.T) {
	a := walk(t, "A", 100)
	b := &marketdata.Series{Symbol: "B", Points: append([]marketdata.Point(nil), a.Points...)}
	for i := range b.Points {
		b.Points[i].Close = 100
	}
	study := pair(t, a, b, analyze.PairOptions{})
	data, err := json.Marshal(study.Daily)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"Start", "End", "Divergences", "Periods", "Corr"} {
		if _, ok := obj[key]; !ok {
			t.Errorf("lost %s in %s", key, data)
		}
	}
	if string(obj["Corr"]) != "null" {
		t.Fatalf("undefined correlation: %s", data)
	}
}
