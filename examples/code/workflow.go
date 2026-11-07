//go:build ignore

// Workflow is a complete offline consumer: load two bundled references,
// convert to EUR, align monthly returns, analyze nominal/real/excess results,
// fit weights on 2000-2009 and evaluate 2010-2019, then optionally export CSV.
// Run: go run examples/code/workflow.go [-out study.csv]
// Copy this single file into another Go module importing pofo to adapt it.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/optimize"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func main() {
	out := flag.String("out", "", "optional long CSV export (nominal, real and fitted curves)")
	flag.Parse()
	if err := run(*out); err != nil {
		log.Fatal(err)
	}
}

func run(output string) error {
	ctx := context.Background()
	client := marketdata.NewClient("") // no disk cache required
	client.Offline = true
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)
	ids := []string{"SP500-USD", "TREASURY-INT-USD"}
	source := analyze.SeriesSource{}
	var assets []*marketdata.Series
	for _, id := range ids {
		s, err := client.Load(ctx, id, marketdata.FetchOptions{From: start, To: end, Currency: "EUR"})
		if err != nil {
			return err
		}
		source[id] = s
		assets = append(assets, s)
	}
	// Choose the consumption basket independently of the FX denomination.
	cpi, err := client.Load(ctx, "^HICP-FR", marketdata.FetchOptions{})
	if err != nil {
		return err
	}
	cash, err := client.Load(ctx, "EURCASH-EUR", marketdata.FetchOptions{})
	if err != nil {
		return err
	}
	spec, err := portfolio.NewSpec("example 60/40",
		portfolio.Line{ID: ids[0], Weight: .6}, portfolio.Line{ID: ids[1], Weight: .4})
	if err != nil {
		return err
	}
	study, err := analyze.Portfolio(ctx, source, spec, analyze.Options{
		Currency: "EUR", Inflation: cpi, RiskFreeIndex: cash,
		CorrelationFrequency: marketdata.Monthly,
		Simulation:           &portfolio.SimOptions{RebalancePeriod: portfolio.Monthly, TransactionCost: .001},
	})
	if err != nil {
		return err
	}
	fmt.Printf("nominal CAGR %.2f %%, real CAGR %.2f %%, cash-excess Sharpe %.2f\n",
		study.Stats.CAGR*100, study.Real.Stats.CAGR*100, study.Stats.Sharpe)
	for _, warning := range study.Warnings {
		fmt.Println("warning:", warning)
	}
	panel, err := marketdata.NewPanel(marketdata.Monthly, assets...)
	if err != nil {
		return err
	}
	// Optimization measures frictionless, every-period rebalancing. A trading
	// simulation of the fitted allocation can use the same Simulation options.
	split := time.Date(2009, 12, 31, 0, 0, 0, 0, time.UTC)
	training := panel.Between(time.Time{}, split)
	heldOut := panel.Between(split.AddDate(0, 0, 1), time.Time{})
	fitted, err := optimize.FitEvaluate(training.R, heldOut.R, panel.PeriodsPerYear(), optimize.Spec{Objective: optimize.MinVolatility})
	if err != nil {
		return err
	}
	if !fitted.Fit.Feasible {
		return fmt.Errorf("infeasible fit: %s", fitted.Fit.FeasibilityReason)
	}
	fmt.Printf("fitted weights %.1f %% / %.1f %%, held-out volatility %.2f %%\n",
		fitted.Fit.Weights[0]*100, fitted.Fit.Weights[1]*100, fitted.OutOfSample.Volatility*100)
	mixed, err := heldOut.Mix("fitted-held-out", map[string]float64{ids[0]: fitted.Fit.Weights[0], ids[1]: fitted.Fit.Weights[1]})
	if err != nil {
		return err
	}
	fittedSeries, err := mixed.Series("fitted-held-out")
	if err != nil {
		return err
	}
	if output == "" {
		return nil
	}
	nominal, err := marketdata.NewSeries("nominal", study.Sim.Dates, study.Sim.Index)
	if err != nil {
		return err
	}
	nominal.Currency = "EUR"
	nominal.Warnings = append([]string(nil), study.Warnings...)
	real := *study.Real.Series
	real.Symbol = "real"
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	err = marketdata.WriteCSV(file, nominal, &real, fittedSeries)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
