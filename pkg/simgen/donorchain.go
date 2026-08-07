package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// DonorChain assembles a fund's history out of real quotes for as long as real
// quotes of something close enough exist: the fund itself, then the nearest
// sibling behind it, then the next one, each spliced in front of the last.
//
// It is the honest answer to a young fund. A reconstruction of a managed-
// futures programme from futures prices agrees with the fund it replicates at
// a monthly correlation near 0.6; the NAV of ANOTHER REAL FUND running the
// same trade agrees at 0.7 to 0.97. Where such a NAV exists, inventing one is
// the worse choice, and the cost of the graft is only that the donor's own
// manager skill and fee load come with it.
//
// Donors are given nearest first (same strategy and manager before same
// strategy and another manager). Each is rescaled to the volatility the chain
// realizes over their common window, on excess-over-cash returns so the cash
// leg is not stretched with it, and spliced with marketdata.ExtendBack, which
// pins its level at the junction. A donor that starts no earlier than the
// chain, or overlaps it too briefly to calibrate, is skipped.
//
// The fund's own quotes are deliberately NOT part of the chain: the recipe
// grafts them on top afterwards (Recipe.SpliceReal), which leaves the chain
// itself measurable against them (Validate). calibrate names the fund the
// donors are volatility-matched to, normally that same fund.
func DonorChain(f Fetcher, cashID, calibrate string, donors []string, from time.Time) (*marketdata.Series, error) {
	ref, err := f.Fetch(calibrate, from)
	if err != nil {
		return nil, fmt.Errorf("donor chain %s: %w", calibrate, err)
	}
	if ref == nil || len(ref.Points) < 30 {
		return nil, fmt.Errorf("donor chain %s: no usable quotes to calibrate on", calibrate)
	}
	cash, err := f.Fetch(cashID, from)
	if err != nil {
		return nil, fmt.Errorf("donor chain cash %s: %w", cashID, err)
	}

	var out *marketdata.Series
	for _, id := range donors {
		d, err := f.Fetch(id, from)
		if err != nil || d == nil || len(d.Points) < 30 {
			continue
		}
		scaled, ok := volMatch(ref, d, cash)
		if !ok {
			continue
		}
		if out == nil {
			out = scaled
			continue
		}
		if !d.First().Date.Before(out.First().Date) {
			continue // adds no history
		}
		out.SimulatedBefore = time.Time{} // let ExtendBack splice one more segment
		marketdata.ExtendBack(out, scaled)
	}
	if out == nil {
		return nil, fmt.Errorf("donor chain %s: no usable donor", calibrate)
	}
	out.SimulatedBefore = time.Time{}
	return out, nil
}

// volMatch rebuilds donor with its excess-over-cash returns scaled so that,
// over the window it shares with ref, it realizes the same volatility. The
// level is left alone: ExtendBack pins it at the junction.
func volMatch(ref, donor, cash *marketdata.Series) (*marketdata.Series, bool) {
	byDate := make(map[time.Time]float64, len(ref.Points))
	for i := 1; i < len(ref.Points); i++ {
		if ref.Points[i-1].Close > 0 {
			byDate[ref.Points[i].Date] = ref.Points[i].Close/ref.Points[i-1].Close - 1
		}
	}
	var a, b []float64
	for i := 1; i < len(donor.Points); i++ {
		if donor.Points[i-1].Close <= 0 {
			continue
		}
		r, ok := byDate[donor.Points[i].Date]
		if !ok {
			continue
		}
		a = append(a, r)
		b = append(b, donor.Points[i].Close/donor.Points[i-1].Close-1)
	}
	if len(a) < 120 { // under six months of common days, the ratio is noise
		return nil, false
	}
	sa, sb := stdev(a), stdev(b)
	if sa <= 0 || sb <= 0 {
		return nil, false
	}
	k := sa / sb
	if k < 0.5 || k > 2 { // too far apart to be the same trade
		return nil, false
	}

	out := &marketdata.Series{Symbol: donor.Symbol, Name: donor.Name, Source: donor.Source}
	v := 100.0
	out.Points = append(out.Points, marketdata.Point{Date: donor.First().Date, Close: v})
	for i := 1; i < len(donor.Points); i++ {
		if donor.Points[i-1].Close <= 0 {
			continue
		}
		r := donor.Points[i].Close/donor.Points[i-1].Close - 1
		c := cashAccrual(cash, donor.Points[i-1].Date, donor.Points[i].Date)
		v *= 1 + k*(r-c) + c
		out.Points = append(out.Points, marketdata.Point{Date: donor.Points[i].Date, Close: v})
	}
	return out, len(out.Points) > 30
}

// cashAccrual is the cash return between two dates, from an annualized
// percent level series (^IRX and friends).
func cashAccrual(cash *marketdata.Series, from, to time.Time) float64 {
	if cash == nil {
		return 0
	}
	lvl, _, ok := cash.At(from)
	if !ok {
		return 0
	}
	return lvl / 100 * to.Sub(from).Hours() / 24 / 365.25
}

func stdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	var m float64
	for _, x := range xs {
		m += x
	}
	m /= float64(len(xs))
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)-1))
}
