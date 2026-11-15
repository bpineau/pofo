//go:build ignore

// Fireholdout confronts training-only FIRE distributions with later realized
// US 60/40 paths. It writes JSON to stdout and never uses the network.
//
// Usage:
//
//	go run examples/code/fireholdout.go -train 20 -years 20 -paths 6000 > /tmp/holdout.json
//
// January origins have complete training and test years. Models are today's
// conventions on today's revised reference, not vintage real-time forecasts.
// CAPE and broad_sample are excluded because their inputs are not vintage-cut.
// No parameters or withdrawal rates are tuned on the held-out observations.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/replay"
	"github.com/bpineau/pofo/pkg/scenario"
)

type prediction struct {
	Model, Policy                  string
	Paths                          int
	P5, P50, P95, ObservedTerminal float64 // real wealth, initial capital = 1
	PredictedRuin                  float64
	ObservedRuin                   bool
}

type fold struct {
	TrainStart, TrainEnd, TestStart, TestEnd string // inclusive calendar dates
	Seed                                     uint64
	Fit                                      fire.Fit
	CentralMu, CentralSigma, CentralNu       float64
	HistoricalCohorts                        int
	Predictions                              []prediction
}

type summary struct {
	Model, Policy               string
	Origins                     int
	BelowP5, BelowP50, AboveP95 int // strict comparisons; ties are inside
	ObservedRuins               int
	MeanPredictedRuin, Brier    float64
}

type report struct {
	Reference, Caveat                            string
	TrainYears, HorizonYears, StrideYears, Paths int
	Seed                                         uint64
	SpendingRate                                 float64
	HistoricalAvailable                          bool
	Folds                                        []fold
	Summary                                      []summary
}

func main() {
	log.SetFlags(0)
	train := flag.Int("train", 20, "complete years preceding each origin (at least 2)")
	years := flag.Int("years", 20, "complete held-out years following each origin")
	stride := flag.Int("stride", 1, "years between January origins; set to horizon for disjoint test windows")
	paths := flag.Int("paths", 6000, "Monte Carlo paths per origin/model; historical cohorts are enumerated exactly")
	seed := flag.Uint64("seed", 61001, "base seed; the origin's calendar year is added")
	spend := flag.Float64("spend", .04, "fixed annual spending as a fraction of initial capital; growth also evaluated")
	workers := flag.Int("workers", 4, "simulation workers (does not change results)")
	flag.Parse()
	if *train < 2 || *train > 200 || *years < 1 || *years > 200 || *stride < 1 || *stride > 200 || *paths < 1 || *workers < 1 || math.IsNaN(*spend) || math.IsInf(*spend, 0) || *spend < 0 {
		log.Fatal("require train in [2,200], years/stride in [1,200], positive paths/workers and finite nonnegative spending")
	}
	panel, err := referencePanel()
	if err != nil {
		log.Fatal(err)
	}
	out := report{
		Reference:  replay.Label + ", real US CPI, current bundled vintage",
		Caveat:     "Retrospective conditional diagnostic. Current model conventions and revised data; no vintage real-time claim. Annual origins overlap unless stride >= horizon. Disjoint windows are not necessarily independent. Historical unavailable when training is shorter than horizon. No calibration or spending optimization on test data.",
		TrainYears: *train, HorizonYears: *years, StrideYears: *stride, Paths: *paths,
		Seed: *seed, SpendingRate: *spend, HistoricalAvailable: *train >= *years,
	}
	firstYear := 0
	for cut := *train * 12; cut+*years*12 <= len(panel.Ends); cut++ {
		if panel.Ends[cut].Month() != time.January {
			continue
		}
		y := panel.Ends[cut].Year()
		if firstYear == 0 {
			firstYear = y
		}
		if (y-firstYear)%*stride != 0 {
			continue
		}
		f, err := evaluateFold(panel, cut, *train, *years, *paths, *workers, *spend, *seed+uint64(y))
		if err != nil {
			log.Fatal(err)
		}
		out.Folds = append(out.Folds, f)
	}
	if len(out.Folds) == 0 {
		log.Fatal("reference has no complete training/test split for these horizons")
	}
	out.Summary = summarize(out.Folds)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}

func referencePanel() (*marketdata.Panel, error) {
	ref, err := replay.Reference()
	if err != nil {
		return nil, err
	}
	s := &marketdata.Series{Symbol: "REAL-US6040"}
	for i, d := range ref.Dates {
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: ref.Index[i]})
	}
	p, err := marketdata.NewPanel(marketdata.Monthly, s)
	if err != nil {
		return nil, err
	}
	// Scenario models need an unbroken monthly sequence. Never silently
	// chain across a dropped period, including a future reference junction.
	for i := 1; i < len(p.Ends); i++ {
		if !p.Starts[i].Equal(p.Ends[i-1]) {
			return nil, fmt.Errorf("reference gap before %s", p.Ends[i].Format("2006-01-02"))
		}
	}
	return p, nil
}

func evaluateFold(p *marketdata.Panel, cut, train, years, paths, workers int, spending float64, seed uint64) (fold, error) {
	start, end := cut-train*12, cut+years*12
	training := scenario.Panel{Returns: [][]float64{p.R[0][start:cut]}, Weights: []float64{1}}
	fit := fire.FitParametric(training, nil)
	market := fire.MarketSpec{Mu: fit.Mu, Sigma: fit.Sigma, Nu: fit.Df, Panel: &training}
	f := fold{TrainStart: p.Starts[start].AddDate(0, 0, 1).Format("2006-01-02"), TrainEnd: p.Ends[cut-1].Format("2006-01-02"), TestStart: p.Starts[cut].AddDate(0, 0, 1).Format("2006-01-02"), TestEnd: p.Ends[end-1].Format("2006-01-02"), Seed: seed, Fit: fit}
	// These returns never enter FitParametric or PrepareModel.
	observed := scenario.Annualize(p.R[0][cut:end], 12)
	models := []string{fire.Central, fire.Bootstrap}
	if train >= years {
		models = append(models, fire.Historical)
	}
	for _, id := range models {
		model, err := fire.PrepareModel(id, market, years)
		if err != nil {
			return f, err
		}
		if id == fire.Central {
			f.CentralMu, f.CentralSigma, f.CentralNu = model.EffectiveMu, model.EffectiveSigma, model.EffectiveNu
		}
		if id == fire.Historical {
			f.HistoricalCohorts = model.HistoricalCohorts
		}
		plan, err := fire.BuildPlan(fire.PlanSpec{Capital: 1, Years: years})
		if err != nil {
			return f, err
		}
		plan.Source = model.Source
		var draws decumul.Draws
		if id == fire.Historical {
			// Enumerate, do not add Monte Carlo error to a finite cohort set.
			cohorts := scenario.HistoricalCohorts{Panel: training, Periods: years * 12}
			for i := 0; i < cohorts.Count(); i++ {
				draws.Returns = append(draws.Returns, scenario.Annualize(cohorts.Cohort(i), 12))
			}
		} else {
			draws = plan.Draw(paths, workers, seed)
		}
		for _, policy := range []string{"growth", "fixed"} {
			plan.NeedAnnual = 0
			if policy == "fixed" {
				plan.NeedAnnual = spending
			}
			actual := plan.RunPath(observed, decumul.Lives{})
			ensemble := plan.SimulateOn(draws, workers)
			terminals := make([]float64, len(ensemble.Paths))
			ruined := 0
			for i, path := range ensemble.Paths {
				terminals[i] = path.Wealth[len(path.Wealth)-1]
				if path.Ruined {
					ruined++
				}
			}
			q := metrics.Quantiles(terminals, .05, .5, .95)
			f.Predictions = append(f.Predictions, prediction{Model: id, Policy: policy, Paths: len(terminals), P5: q[0], P50: q[1], P95: q[2], ObservedTerminal: actual.Wealth[len(actual.Wealth)-1], PredictedRuin: float64(ruined) / float64(len(terminals)), ObservedRuin: actual.Ruined})
		}
	}
	return f, nil
}

func summarize(folds []fold) []summary {
	var out []summary
	for _, f := range folds {
		for _, p := range f.Predictions {
			i := 0
			for i < len(out) && (out[i].Model != p.Model || out[i].Policy != p.Policy) {
				i++
			}
			if i == len(out) {
				out = append(out, summary{Model: p.Model, Policy: p.Policy})
			}
			s := &out[i]
			s.Origins++
			if p.ObservedTerminal < p.P5 {
				s.BelowP5++
			}
			if p.ObservedTerminal < p.P50 {
				s.BelowP50++
			}
			if p.ObservedTerminal > p.P95 {
				s.AboveP95++
			}
			y := 0.0
			if p.ObservedRuin {
				s.ObservedRuins++
				y = 1
			}
			s.MeanPredictedRuin += p.PredictedRuin
			s.Brier += (p.PredictedRuin - y) * (p.PredictedRuin - y)
		}
	}
	for i := range out {
		out[i].MeanPredictedRuin /= float64(out[i].Origins)
		out[i].Brier /= float64(out[i].Origins)
	}
	return out
}
