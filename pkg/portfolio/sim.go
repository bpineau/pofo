package portfolio

import (
	"fmt"
	"time"

	"github.com/bpineau/portfodor/pkg/marketdata"
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

	// Leverage keeps the weights as written: the residual (1 − Σ weights)
	// is a cash position, negative when the portfolio is levered.
	Leverage bool

	// BorrowSpread is the yearly spread (percent) added to the cash rate
	// on borrowed money (negative cash). Ignored without Leverage.
	BorrowSpread float64

	// Cash is the financing/deposit rate series, in annualized percent
	// levels (e.g. ^IRX). Nil means a flat 0 % rate.
	Cash *marketdata.Series
}

// SimResult is the simulated value of a portfolio over time, starting at 100.
type SimResult struct {
	Dates  []time.Time
	Values []float64

	// Ruined is true when a levered portfolio's net value hit zero: the
	// series is truncated at that point.
	Ruined bool
}

// Simulate replays the portfolio from the first date where every asset has a
// quote until the last date where every asset still has one. The portfolio
// starts at an index value of 100 and is rebalanced back to its target
// weights every rebalanceDays calendar days (0 disables rebalancing).
// Prices are forward-filled across each asset's non-trading days.
func Simulate(p *Portfolio, rebalanceDays int) (*SimResult, error) {
	if len(p.Assets) == 0 {
		return nil, fmt.Errorf("empty portfolio")
	}
	for _, a := range p.Assets {
		if len(a.Series.Points) == 0 {
			return nil, fmt.Errorf("no quotes for %s", a.Symbol)
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
		return nil, fmt.Errorf("no common period between the portfolio's assets")
	}

	// Union of trading dates inside the window, prices forward-filled.
	// The cash-rate series, when present, is aligned too but never
	// constrains the window (start/end come from the assets alone).
	seriesList := make([]*marketdata.Series, len(p.Assets))
	for i, a := range p.Assets {
		seriesList[i] = a.Series
	}
	rateIdx := -1
	if p.Leverage && p.Cash != nil {
		rateIdx = len(seriesList)
		seriesList = append(seriesList, p.Cash)
	}
	dates, prices := marketdata.Align(seriesList, start, end)

	// Without explicit leverage, weights are normalized defensively (the
	// parser already does); with it, they are exposures of the capital
	// and the residual lives in a cash position.
	sumW := 0.0
	for _, a := range p.Assets {
		sumW += a.Weight
	}
	if sumW <= 0 {
		return nil, fmt.Errorf("weights sum to zero")
	}
	norm := sumW
	if p.Leverage {
		norm = 1
	}

	shares := make([]float64, len(p.Assets))
	cash := 0.0
	setShares := func(k int, total float64) {
		invested := 0.0
		for i, a := range p.Assets {
			shares[i] = total * (a.Weight / norm) / prices[i][k]
			invested += shares[i] * prices[i][k]
		}
		if p.Leverage {
			cash = total - invested
		}
	}
	// dailyCashRate accrues the cash position: deposits earn the cash
	// rate, borrowed money pays it plus the spread.
	dailyCashRate := func(k int) float64 {
		if !p.Leverage || cash == 0 {
			return 0
		}
		r := 0.0
		if rateIdx >= 0 {
			r = prices[rateIdx][k-1] / 100 / 252
		}
		if cash < 0 && p.BorrowSpread > 0 {
			r += p.BorrowSpread / 100 / 252
		}
		return r
	}

	values := make([]float64, len(dates))
	values[0] = 100
	setShares(0, 100)
	dailyFee := 0.0
	if p.EnvelopeFees > 0 {
		dailyFee = p.EnvelopeFees / 100 / 252
	}
	nextRebalance := dates[0].AddDate(0, 0, rebalanceDays)
	ruined := false
	for k := 1; k < len(dates); k++ {
		cash *= (1 - dailyFee) * (1 + dailyCashRate(k))
		v := cash
		for i := range shares {
			shares[i] *= 1 - dailyFee
			v += shares[i] * prices[i][k]
		}
		if p.Leverage && v <= 0 {
			// Capital wiped out: the series stops here.
			dates, values, ruined = dates[:k], values[:k], true
			break
		}
		values[k] = v
		if rebalanceDays > 0 && !dates[k].Before(nextRebalance) {
			setShares(k, v)
			nextRebalance = dates[k].AddDate(0, 0, rebalanceDays)
		}
	}
	return &SimResult{Dates: dates, Values: values, Ruined: ruined}, nil
}
