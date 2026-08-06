package portfolio

import (
	"fmt"
	"time"

	"portfodor/marketdata"
)

// Asset is a resolved portfolio constituent, ready for simulation.
type Asset struct {
	ID     string // identifier as written in the portfolio file
	Symbol string
	Name   string
	Weight float64 // fraction of the portfolio; weights sum to 1
	Fees   float64 // TER in percent per year; negative when unknown
	Series *marketdata.Series
}

// Portfolio is a fully resolved portfolio.
type Portfolio struct {
	Name     string
	Assets   []Asset
	Warnings []string

	// EnvelopeFees is the yearly fee of the hosting envelope in percent
	// per year (0 or negative: none). Asset TERs are already net in
	// prices; envelope fees are not, so Simulate deducts them daily.
	EnvelopeFees float64
}

// SimResult is the simulated value of a portfolio over time, starting at 100.
type SimResult struct {
	Dates  []time.Time
	Values []float64
}

// Simulate replays the portfolio from the first date where every asset has a
// quote until the last date where every asset still has one. The portfolio
// starts at an index value of 100 and is rebalanced back to its target
// weights every rebalanceDays calendar days (0 disables rebalancing).
// Prices are forward-filled across each asset's non-trading days.
func Simulate(p *Portfolio, rebalanceDays int) (*SimResult, error) {
	if len(p.Assets) == 0 {
		return nil, fmt.Errorf("portefeuille vide")
	}
	for _, a := range p.Assets {
		if len(a.Series.Points) == 0 {
			return nil, fmt.Errorf("aucune cotation pour %s", a.Symbol)
		}
	}

	// Common window: every asset must already trade, and still trade.
	start := p.Assets[0].Series.First().Date
	end := p.Assets[0].Series.Last().Date
	for _, a := range p.Assets[1:] {
		if f := a.Series.First().Date; f.After(start) {
			start = f
		}
		if l := a.Series.Last().Date; l.Before(end) {
			end = l
		}
	}
	if !start.Before(end) {
		return nil, fmt.Errorf("pas de période commune entre les actifs du portefeuille")
	}

	// Union of trading dates inside the window, prices forward-filled.
	seriesList := make([]*marketdata.Series, len(p.Assets))
	for i, a := range p.Assets {
		seriesList[i] = a.Series
	}
	dates, prices := marketdata.Align(seriesList, start, end)

	// Normalize weights defensively (the parser already does).
	sumW := 0.0
	for _, a := range p.Assets {
		sumW += a.Weight
	}
	if sumW <= 0 {
		return nil, fmt.Errorf("somme des poids nulle")
	}

	shares := make([]float64, len(p.Assets))
	setShares := func(k int, total float64) {
		for i, a := range p.Assets {
			shares[i] = total * (a.Weight / sumW) / prices[i][k]
		}
	}

	values := make([]float64, len(dates))
	values[0] = 100
	setShares(0, 100)
	dailyFee := 0.0
	if p.EnvelopeFees > 0 {
		dailyFee = p.EnvelopeFees / 100 / 252
	}
	nextRebalance := dates[0].AddDate(0, 0, rebalanceDays)
	for k := 1; k < len(dates); k++ {
		v := 0.0
		for i := range shares {
			shares[i] *= 1 - dailyFee
			v += shares[i] * prices[i][k]
		}
		values[k] = v
		if rebalanceDays > 0 && !dates[k].Before(nextRebalance) {
			setShares(k, v)
			nextRebalance = dates[k].AddDate(0, 0, rebalanceDays)
		}
	}
	return &SimResult{Dates: dates, Values: values}, nil
}
