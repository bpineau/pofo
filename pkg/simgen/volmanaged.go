package simgen

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// VolManagedConfig drives the volatility-managed engine (VolManaged): one
// asset held at a multiplier m of its capital, the rest in cash, m set to
// TargetVol over the asset's trailing realized volatility and bounded by Floor
// and Cap. This is the construction of Moreira and Muir, "Volatility-Managed
// Portfolios" (Journal of Finance, 2017), on a single asset: exposure falls
// when volatility rises, because volatility clusters and returns do not rise
// with it in the short run.
//
// Every rate and target is a FRACTION per year (0.12 = 12 %/yr), the package
// convention.
type VolManagedConfig struct {
	// AssetID is the frame column of the managed asset, a price series.
	AssetID string
	// CashID is the frame column the unallocated capital earns (1 − m when m
	// < 1) and the exposure above 1 is financed at (m − 1 when m > 1): an
	// accrual index (XEON, a money-market fund) or a rate symbol (^IRX),
	// which BuildFrame turns into the same accrual.
	CashID string
	// TargetVol is the annualized volatility the multiplier aims at, a
	// FRACTION (0.12 = 12 %/yr). It must be positive.
	TargetVol float64
	// Lookback is how many daily returns of the asset (frame steps) the
	// realized volatility reads, e.g. 63 for about three months. At least 2.
	Lookback int
	// Floor and Cap bound the multiplier, 0 <= Floor <= Cap, Cap > 0. A Cap
	// of 1 never levers; above 1 the excess is financed at the cash rate and
	// pays Costs.
	Floor, Cap float64
	// Rebalance refreshes the multiplier, and trades the position back to
	// it, every Rebalance frame steps; 0 (the default) does so on the last
	// frame date of every calendar month, monthly as the published studies
	// do. In between, the position is held and its weights drift.
	Rebalance int
	// Costs is what the exposure above 1 pays on top of the cash rate; the
	// zero value pays nothing more. Never charged while m <= 1.
	Costs LeverageCosts
}

// LeverageCosts prices the part of a volatility-managed position above 100 %
// of capital, X = m − 1 when m > 1. Every figure is a FRACTION per year,
// charged over the calendar days each frame step spans (the convention of
// every fee in this package).
type LeverageCosts struct {
	// AnnualFee is charged on X (a leveraged ETF's ongoing charge, say);
	// 0.0035 = 0.35 %/yr.
	AnnualFee float64
	// Spread is the financing spread over the cash rate, charged on X; the
	// cash rate itself is already paid through the negative cash weight.
	Spread float64
	// ETFLeverage, when above 1, carries the excess through a daily-reset
	// ETF of that leverage L, which then pays its reset drag: X/(L − 1) of
	// capital sits in the ETF, and an ETF of leverage L loses about
	// (L² − L)/2 × σ² a year against L times its index, σ being the asset's
	// realized volatility of the moment (the trailing Lookback estimate as of
	// the previous close). For L = 2 that is X × σ². It is an APPROXIMATION:
	// the exact drag depends on the path (it is the second-order term of the
	// ETF's daily compounding), and it is charged on top of an index that
	// holds X as a plain financed position. 0 (or 1) charges no drag.
	ETFLeverage float64
}

// VolManagedPath is what VolManaged computes, aligned on the frame's dates
// from Start on.
type VolManagedPath struct {
	// Start is the frame index of the first value: the first refresh date
	// with Lookback returns behind it.
	Start int
	// Values is the index, base 100 at Dates[Start], one value per frame
	// date from Start.
	Values []float64
	// Multiplier[i] is the target m in force at the close of
	// Dates[Start+i], decided with the returns through the last refresh and
	// traded to there; the exposure then drifts from it until the next one.
	// Same length as Values.
	Multiplier []float64
}

// VolManaged runs the volatility-managed engine on a frame holding
// cfg.AssetID and cfg.CashID. The return of each step k is
//
//	w × r_asset[k] + (1 − w) × r_cash[k] − costs
//
// with w the exposure in force. At every refresh w is reset to the
// multiplier m: TargetVol over the annualized sample standard deviation of
// the asset's last Lookback daily returns (simple returns, annualized at
// metrics.TradingDaysPerYear), bounded by Floor and Cap, decided at a close
// for the steps that follow it, never for the step that produced it. A window
// with no measurable volatility sets m to Cap.
//
// Between two refreshes the position is HELD, not traded: w drifts with the
// two legs' own returns, as an investor's does between two trades and as the
// monthly arithmetic of Moreira and Muir (m × the month's return) implies.
// Rebalancing to m every day instead is a different strategy, and not a free
// one to assume: on MSCI World in euros over 2005-2026, at a 12 % target and a
// monthly refresh, it reads 0.08 pt a year higher. With Cap <= 1 the cash
// position stays positive, so w never exceeds 1 and cfg.Costs, charged on the
// excess w − 1 only, never bites.
func VolManaged(fr *Frame, cfg VolManagedConfig) (*VolManagedPath, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	asset, ok := fr.Returns[cfg.AssetID]
	if !ok {
		return nil, fmt.Errorf("simgen: VolManaged: asset %s missing from frame", cfg.AssetID)
	}
	cash, ok := fr.Returns[cfg.CashID]
	if !ok {
		return nil, fmt.Errorf("simgen: VolManaged: cash %s missing from frame", cfg.CashID)
	}
	n := len(fr.Dates)
	refresh := cfg.refresher(fr.Dates)
	start := -1
	for k := cfg.Lookback; k < n-1; k++ {
		if refresh(k, cfg.Rebalance) {
			start = k
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("simgen: VolManaged: history too short for a %d-day lookback (%d dates)", cfg.Lookback, n)
	}

	m := cfg.multiplier(realizedVol(asset, start, cfg.Lookback))
	out := &VolManagedPath{Start: start, Values: []float64{100}, Multiplier: []float64{m}}
	held, cashHeld := 100*m, 100*(1-m) // the two positions, in index points
	since := 0
	for k := start + 1; k < n; k++ {
		v0 := held + cashHeld
		x := held/v0 - 1 // the exposure above 100 % carried into the step
		held *= 1 + asset[k]
		cashHeld *= 1 + cash[k]
		if x > 0 {
			cashHeld -= v0 * feeAccrual(cfg.Costs.annual(x, realizedVol(asset, k-1, cfg.Lookback)), fr.Dates[k-1], fr.Dates[k])
		}
		v := held + cashHeld
		since++
		if refresh(k, since) {
			m, since = cfg.multiplier(realizedVol(asset, k, cfg.Lookback)), 0
			held, cashHeld = m*v, (1-m)*v
		}
		out.Values = append(out.Values, v)
		out.Multiplier = append(out.Multiplier, m)
	}
	return out, nil
}

// validate rejects a configuration the engine cannot run, naming the field.
func (cfg VolManagedConfig) validate() error {
	c := cfg.Costs
	switch {
	case cfg.AssetID == "" || cfg.CashID == "":
		return errors.New("simgen: VolManaged: AssetID and CashID are required")
	case !(cfg.TargetVol > 0):
		return fmt.Errorf("simgen: VolManaged: TargetVol %v must be positive (a FRACTION, 0.12 = 12 %%/yr)", cfg.TargetVol)
	case cfg.Lookback < 2:
		return fmt.Errorf("simgen: VolManaged: Lookback %d must be at least 2 returns", cfg.Lookback)
	case !(cfg.Cap > 0) || cfg.Floor < 0 || cfg.Floor > cfg.Cap:
		return fmt.Errorf("simgen: VolManaged: bounds Floor %v, Cap %v must satisfy 0 <= Floor <= Cap, Cap > 0", cfg.Floor, cfg.Cap)
	case cfg.Rebalance < 0:
		return fmt.Errorf("simgen: VolManaged: Rebalance %d must be positive, or 0 for month ends", cfg.Rebalance)
	case c.AnnualFee < 0 || c.Spread < 0 || c.ETFLeverage < 0:
		return fmt.Errorf("simgen: VolManaged: leverage costs must not be negative: %+v", c)
	}
	return nil
}

// refresher returns the rule that says whether the multiplier is refreshed at
// the close of frame index k, given the steps since the last refresh: every
// Rebalance steps, or on the last date of each calendar month when Rebalance
// is 0 (knowing that tomorrow opens a new month is the calendar, not data;
// the frame's last date, which no step follows, is not one).
func (cfg VolManagedConfig) refresher(dates []time.Time) func(k, since int) bool {
	if cfg.Rebalance > 0 {
		return func(_, since int) bool { return since >= cfg.Rebalance }
	}
	return func(k, _ int) bool {
		return k+1 < len(dates) && (dates[k+1].Month() != dates[k].Month() || dates[k+1].Year() != dates[k].Year())
	}
}

// multiplier is the exposure a realized volatility calls for, within bounds.
func (cfg VolManagedConfig) multiplier(vol float64) float64 {
	if !(vol > 0) {
		return cfg.Cap
	}
	return math.Max(cfg.Floor, math.Min(cfg.Cap, cfg.TargetVol/vol))
}

// annual is the yearly charge, a FRACTION of capital, of an excess exposure x
// at realized volatility vol (see LeverageCosts).
func (c LeverageCosts) annual(x, vol float64) float64 {
	cost := x * (c.AnnualFee + c.Spread)
	if l := c.ETFLeverage; l > 1 {
		cost += x / (l - 1) * (l*l - l) / 2 * vol * vol
	}
	return cost
}

// realizedVol is the annualized sample standard deviation of the returns
// xs[k-window+1 .. k], the window ending at index k included.
func realizedVol(xs []float64, k, window int) float64 {
	lo := k - window + 1
	if lo < 1 || window < 2 {
		return 0
	}
	mean := 0.0
	for _, x := range xs[lo : k+1] {
		mean += x
	}
	mean /= float64(window)
	ss := 0.0
	for _, x := range xs[lo : k+1] {
		ss += (x - mean) * (x - mean)
	}
	return math.Sqrt(ss / float64(window-1) * metrics.TradingDaysPerYear)
}

// VolManagedBuilder runs the volatility-managed engine on series fetched by
// identifier, the form a Recipe.Build takes. NewVolManagedBuilder makes one.
type VolManagedBuilder struct {
	name string
	cfg  VolManagedConfig
}

// NewVolManagedBuilder checks cfg and returns a builder whose Build fetches
// cfg.AssetID and cfg.CashID and names the resulting series name. Both must
// be in the same currency: the engine mixes their returns as they come.
func NewVolManagedBuilder(name string, cfg VolManagedConfig) (*VolManagedBuilder, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &VolManagedBuilder{name: name, cfg: cfg}, nil
}

// Build fetches the asset and the cash from f, aligns them (BuildFrame, from
// the later of the two first quotes) and returns the volatility-managed index,
// base 100 on its first refresh date. Its signature is Recipe.Build's.
func (b *VolManagedBuilder) Build(f Fetcher, from time.Time) (*marketdata.Series, error) {
	fr, err := BuildFrame(f, []string{b.cfg.CashID, b.cfg.AssetID}, from)
	if err != nil {
		return nil, err
	}
	p, err := VolManaged(fr, b.cfg)
	if err != nil {
		return nil, err
	}
	s := &marketdata.Series{Name: b.name, Source: "simdata"}
	for i, v := range p.Values {
		s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[p.Start+i], Close: v})
	}
	return s, nil
}

// iwdaVM12 is the volatility-managed world equity line's rule: a 12 %/yr
// target on the trailing 63 sessions (about three months), refreshed at
// every month end, exposure between 50 and 100 % (no leverage, so no cost
// model), the rest in euro overnight cash.
var iwdaVM12 = VolManagedConfig{
	AssetID: "world-eur", CashID: "cash-eur",
	TargetVol: 0.12, Lookback: 63, Floor: 0.5, Cap: 1,
}

// iwdaVM12Recipe ships a STRATEGY rather than a backcast: the iShares Core
// MSCI World held in EUR at a volatility-managed exposure (VolManaged with
// iwdaVM12), served as the catalog index IWDA-VM12 so a portfolio file can
// hold it like any other line. Its two legs are the series a euro investor
// in that pair holds:
//
//   - the world equity leg is iwdaInEUR, real IWDA from 2009 and its MSCI
//     World reconstruction before, converted at the EUR/USD spot, from the
//     first day it quotes DAILY (dailyStretch): before 1972 the reconstruction
//     is month-end only, and a daily volatility read on month-end returns is
//     not one;
//   - the cash leg is XEON, the euro overnight swap fund: its real NAV from
//     2007, the euro overnight chain less its 0.10 %/yr charge before
//     (xeonHistory).
//
// The legs' charges are in their prices, so nothing is deducted on top. The
// file starts a lookback after the equity leg's daily start and is rebuilt by
// "pofo -gen-simdata IWDA-VM12" (make simdata). No real series exists to
// grade it against, so the audit leaves it out (Ungraded).
func iwdaVM12Recipe() Recipe {
	return Recipe{
		ID:   "IWDA-VM12",
		Name: "iShares Core MSCI World in EUR, volatility-managed, 12 % target (strategy)",
		Method: "IWDA in EUR (real from 2009, MSCI World reconstruction before, EURUSD spot, daily from 1972) held at " +
			"m = 12% / realized vol of the last 63 sessions, bounded to [0.5, 1], refreshed at month ends and held in between, " +
			"the rest in XEON (real from 2007, euro overnight chain less 0.10%/yr before); Moreira-Muir (JF 2017)",
		Build:    iwdaVM12Build,
		Ungraded: "a volatility-managed strategy on IWDA in EUR, not a fund: no real series quotes it",
	}
}

// iwdaVM12Build assembles the two legs and runs the engine on them.
func iwdaVM12Build(f Fetcher, from time.Time) (*marketdata.Series, error) {
	world, err := iwdaInEUR("MSCI World in EUR (IWDA)", f, from)
	if err != nil {
		return nil, err
	}
	cash, err := xeonHistory(f, from)
	if err != nil {
		return nil, err
	}
	b, err := NewVolManagedBuilder("IWDA in EUR, volatility-managed 12 %", iwdaVM12)
	if err != nil {
		return nil, err
	}
	return b.Build(seriesFetcher{own: map[string]*marketdata.Series{
		iwdaVM12.AssetID: dailyStretch(world),
		iwdaVM12.CashID:  cash,
	}, fallback: f}, from)
}

// xeonHistory is XEON as a SIM consumer reads it: the fund's real NAV where it
// quotes, its recipe's reconstruction before. The reconstruction alone stands
// when the real NAV cannot be fetched.
func xeonHistory(f Fetcher, from time.Time) (*marketdata.Series, error) {
	r := xeonRecipe()
	recon, err := r.Build(f, from)
	if err != nil {
		return nil, err
	}
	if real, err := f.Fetch(r.SpliceReal, from); err == nil && real != nil && len(real.Points) > 300 {
		return Splice(real.WithoutEstimates(), recon), nil
	}
	return recon, nil
}

// dailyStretch drops the front of s up to its last gap of more than ten calendar
// days, which leaves the stretch that quotes daily (a reconstruction that is
// month-end before its daily shape begins). A series with no such gap is
// returned as it is.
func dailyStretch(s *marketdata.Series) *marketdata.Series {
	cut := 0
	for i := 1; i < len(s.Points); i++ {
		if s.Points[i].Date.Sub(s.Points[i-1].Date) > 10*24*time.Hour {
			cut = i
		}
	}
	if cut == 0 {
		return s
	}
	out := *s
	out.Points = s.Points[cut:]
	return &out
}
