// The -permanent mode: the tactical Permanent Portfolio 2.0 backtest
// (pkg/permanent wiring).
package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/permanent"
	"github.com/bpineau/pofo/pkg/scenario"
)

// runPermanent backtests the tactical Permanent Portfolio 2.0 (pkg/permanent)
// against the static Browne PP and MSCI World, all in real terms, then reports
// the decumulation ruin probabilities that matter for FIRE. It fetches the four
// sleeves, deflates them to real monthly returns, drives the allocation from the
// embedded macro panel, and block-bootstraps the realized tactical and static
// return streams through the decumul engine.
func runPermanent(ctx context.Context, opt *options, c *marketdata.Client) error {
	monthEnd := func(s *marketdata.Series) map[time.Time]float64 {
		out := map[time.Time]float64{}
		for _, p := range s.Points {
			out[time.Date(p.Date.Year(), p.Date.Month(), 1, 0, 0, 0, 0, time.UTC)] = p.Close
		}
		return out
	}
	fetchX := func(id string) (map[time.Time]float64, error) {
		s, err := c.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "USD"})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		return monthEnd(s), nil
	}
	fetchLevel := func(id string) (map[time.Time]float64, error) {
		s, err := c.Fetch(ctx, id, time.Time{})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		return monthEnd(s), nil
	}
	cpi, err := fetchLevel("^CPI-US")
	if err != nil {
		return err
	}
	eqM, err := fetchX("URTHSIM")
	if err != nil {
		return err
	}
	boM, err := fetchX("TLTSIM")
	if err != nil {
		return err
	}
	goM, err := fetchX("XAUUSDSIM")
	if err != nil {
		return err
	}
	irx, err := fetchLevel("^IRX")
	if err != nil {
		return err
	}

	var months []time.Time
	for m := range cpi {
		months = append(months, m)
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Before(months[j]) })

	// nominal cash index accrued from the short rate, then deflated like the rest.
	cashIdx := map[time.Time]float64{}
	cash := 1.0
	var prev time.Time
	for _, m := range months {
		if !prev.IsZero() {
			if y, ok := irx[prev]; ok {
				cash *= 1 + y/100/12
			}
		}
		cashIdx[m] = cash
		prev = m
	}
	realRet := func(nom map[time.Time]float64, pr, m time.Time) (float64, bool) {
		n0, o1 := nom[pr]
		n1, o2 := nom[m]
		p0, o3 := cpi[pr]
		p1, o4 := cpi[m]
		if !(o1 && o2 && o3 && o4) || n0 == 0 || p0 == 0 || p1 == 0 {
			return 0, false
		}
		return (n1/p1)/(n0/p0) - 1, true
	}

	var ar permanent.AssetReturns
	for i := 1; i < len(months); i++ {
		m, pr := months[i], months[i-1]
		e, o1 := realRet(eqM, pr, m)
		b, o2 := realRet(boM, pr, m)
		ca, o3 := realRet(cashIdx, pr, m)
		g, o4 := realRet(goM, pr, m)
		if !(o1 && o2 && o3 && o4) {
			continue
		}
		ar.Dates = append(ar.Dates, m)
		ar.Equity = append(ar.Equity, e)
		ar.Bonds = append(ar.Bonds, b)
		ar.Cash = append(ar.Cash, ca)
		ar.Gold = append(ar.Gold, g)
	}
	if len(ar.Dates) < 120 {
		return fmt.Errorf("permanent: too few aligned months (%d)", len(ar.Dates))
	}

	panel, err := permanent.LoadPanel()
	if err != nil {
		return err
	}
	regimes := panel.Regimes(ar.Dates[0].AddDate(-1, 0, 0), ar.Dates[len(ar.Dates)-1], permanent.DefaultSignalConfig())
	res, err := permanent.Simulate(regimes, ar, permanent.DefaultParams())
	if err != nil {
		return err
	}

	// equity-only real returns aligned to the backtest dates, for a benchmark row.
	eqByDate := make(map[time.Time]float64, len(ar.Dates))
	for i, d := range ar.Dates {
		eqByDate[d] = ar.Equity[i]
	}
	eqReal := make([]float64, len(res.Dates))
	for i, d := range res.Dates {
		eqReal[i] = eqByDate[d]
	}

	fmt.Printf("Tactical Permanent Portfolio 2.0 (Darcet), REAL, monthly, %s..%s\n",
		res.Dates[0].Format("2006-01"), res.Dates[len(res.Dates)-1].Format("2006-01"))
	fmt.Printf("Reconstruction of an undisclosed method; see docs/darcet-permanent-portfolio-design.md\n\n")
	fmt.Printf("%-24s %7s %6s %7s %5s %7s\n", "portfolio", "CAGR", "vol", "maxDD", "%UW", "longUW")
	statRow := func(name string, series []float64) {
		s := permanent.Compute(series)
		fmt.Printf("%-24s %6.2f%% %5.1f%% %6.1f%% %4.0f%% %5.1fy\n",
			name, s.CAGR*100, s.Vol*100, s.MaxDrawdown*100, s.UnderwaterFraction*100, float64(s.LongestUnderwater)/12)
	}
	statRow("tactical PP 2.0", res.Tactical)
	statRow("static Browne PP", res.Static)
	statRow("MSCI World (equity)", eqReal)

	const years = 40
	rates := []float64{0.030, 0.035, 0.040, 0.045}
	ruin := func(series []float64, wr float64) float64 {
		src := scenario.StationaryBootstrap{
			Panel:     scenario.Panel{Returns: [][]float64{series}, Weights: []float64{1}},
			MeanBlock: 24,
			Periods:   years * 12,
		}
		plan := decumul.Plan{
			Capital: 1, NeedAnnual: wr, Years: years,
			Source: src, Monthly: true, Tax: decumul.CTOFlatTax{Rate: 0},
		}
		return plan.Simulate(3000, runtime.NumCPU(), 1).Outcome().RuinProb
	}
	fmt.Printf("\n%d-year ruin probability at a fixed real withdrawal (stationary bootstrap):\n", years)
	fmt.Printf("%-24s", "withdrawal rate")
	for _, wr := range rates {
		fmt.Printf(" %6.1f%%", wr*100)
	}
	fmt.Println()
	ruinRow := func(name string, series []float64) {
		fmt.Printf("%-24s", name)
		for _, wr := range rates {
			fmt.Printf(" %6.1f%%", ruin(series, wr)*100)
		}
		fmt.Println()
	}
	ruinRow("tactical PP 2.0", res.Tactical)
	ruinRow("static Browne PP", res.Static)
	fmt.Println("\nRuin = share of 40-year retirements exhausted. The realized real series is")
	fmt.Println("block-bootstrapped (mean block 24 months): one historical path, no tax or fees.")
	_ = opt
	return nil
}
