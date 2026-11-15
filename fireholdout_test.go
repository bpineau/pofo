package pofo

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os/exec"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/replay"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Exercise the runnable diagnostic's split and reproducibility contract.
// This deliberately checks dates and accounting, not frozen market outcomes.
func TestFIREHoldoutExample(t *testing.T) {
	run := func(workers string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "run", "examples/code/fireholdout.go", "-train", "35", "-years", "35", "-paths", "16", "-workers", workers)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("holdout example: %v\n%s", err, out)
		}
		return out
	}
	first, second := run("1"), run("3")
	if !bytes.Equal(first, second) {
		t.Fatal("worker count changed the diagnostic")
	}
	var report struct {
		Folds []struct {
			TrainStart, TrainEnd, TestStart, TestEnd string
			Fit                                      fire.Fit
			Predictions                              []struct {
				Model, Policy                  string
				Paths                          int
				P5, P50, P95, ObservedTerminal float64
				PredictedRuin                  float64
				ObservedRuin                   bool
			}
		}
	}
	if err := json.Unmarshal(first, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Folds) == 0 {
		t.Fatal("no held-out folds")
	}
	ref, err := replay.Reference()
	if err != nil {
		t.Fatal(err)
	}
	parse := func(s string) time.Time {
		t.Helper()
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	for _, fold := range report.Folds {
		start, cut, testStart, end := parse(fold.TrainStart), parse(fold.TrainEnd), parse(fold.TestStart), parse(fold.TestEnd)
		if !cut.Before(testStart) || !start.AddDate(35, 0, 0).Equal(testStart) || !testStart.AddDate(35, 0, -1).Equal(end) {
			t.Fatalf("overlap, look-ahead or incomplete horizon: %+v", fold)
		}
		if len(fold.Predictions) != 6 {
			t.Fatalf("predictions: %d, want 3 models x 2 plans", len(fold.Predictions))
		}
		// Select by calendar dates independently of the example's positional
		// cuts. A fit accidentally using any test months must disagree.
		s := &marketdata.Series{Symbol: "TRAIN"}
		for i, d := range ref.Dates {
			if !d.Before(start.AddDate(0, 0, -1)) && !d.After(cut) {
				s.Points = append(s.Points, marketdata.Point{Date: d, Close: ref.Index[i]})
			}
		}
		monthly, err := marketdata.NewPanel(marketdata.Monthly, s)
		if err != nil {
			t.Fatal(err)
		}
		wantFit := fire.FitParametric(scenario.Panel{Returns: monthly.R, Weights: []float64{1}}, nil)
		if fold.Fit != wantFit {
			t.Fatalf("fit used the wrong history: %+v, want %+v", fold.Fit, wantFit)
		}
		var training, heldout scenario.Sequence
		for i, y := range ref.Years {
			if y >= start.Year() && y <= cut.Year() {
				training = append(training, ref.Annual[i])
			}
			if y >= testStart.Year() && y <= end.Year() {
				heldout = append(heldout, ref.Annual[i])
			}
		}
		for _, p := range fold.Predictions {
			if p.P5 > p.P50 || p.P50 > p.P95 || p.PredictedRuin < 0 || p.PredictedRuin > 1 {
				t.Fatalf("invalid prediction: %+v", p)
			}
			if p.Model == "historical" && (p.Paths != 1 || p.P5 != p.P95) {
				t.Fatalf("35 years must supply exactly one 35-year training cohort: %+v", p)
			}
			if p.Policy == "growth" && p.ObservedRuin {
				t.Fatal("positive reference prices cannot ruin a zero-spending plan")
			}
			plan := decumul.Plan{Capital: 1, Years: 35}
			if p.Policy == "fixed" {
				plan.NeedAnnual = .04
			}
			actual := plan.RunPath(heldout, decumul.Lives{})
			if math.Abs(p.ObservedTerminal-actual.Wealth[35]) > 1e-12*math.Max(1, actual.Wealth[35]) || p.ObservedRuin != actual.Ruined {
				t.Fatalf("wrong realized path: %+v", p)
			}
			if p.Model == "historical" {
				prior := plan.RunPath(training, decumul.Lives{})
				if math.Abs(p.P50-prior.Wealth[35]) > 1e-12*math.Max(1, prior.Wealth[35]) {
					t.Fatalf("historical prediction used the wrong cohort: %+v", p)
				}
			}
		}
	}
}
