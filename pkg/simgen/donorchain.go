package simgen

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Donor is one real fund NAV offered to a chain, with the annualized return
// that must be added back to it before it may stand in for the target.
//
// A donor is spliced for its returns, and those returns are net of ITS price
// list, not the target's. Where the two differ, the difference is a wrapper
// cost, not a strategy result, and Uplift removes it: a constant fraction per
// year added to that donor's segment and to no other. Zero leaves the donor
// exactly as its manager published it. The uplift belongs to the caller,
// which is the only layer that knows what either vehicle charges; see
// feeAligned in the recipes for how this family derives it, and why it is
// derived from published fee schedules rather than from observed return gaps.
type Donor struct {
	ID     string
	Uplift float64 // fraction per year, e.g. 0.006 = +0.60 %/yr
}

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
// leg is not stretched with it, lifted by its own Uplift so it carries the
// target's fee load rather than its own, and spliced with
// marketdata.ExtendBack, which pins its level at the junction. A donor that
// starts no earlier than the chain, or overlaps it too briefly to calibrate,
// is skipped.
//
// The fund's own quotes are deliberately NOT part of the chain: the recipe
// grafts them on top afterwards (Recipe.SpliceReal), which leaves the chain
// itself measurable against them (Validate). calibrate names the fund the
// donors are volatility-matched to, normally that same fund.
//
// A donor need not quote daily, and the oldest ones do not: a fund that dealt
// WEEKLY until 2016 would otherwise contribute a decade of week-sized steps to
// a daily file, and statistics that annualize per observation read that as
// roughly sqrt(5) times the fund's real volatility. texture, when given, fixes
// that without touching a single NAV: a sparse donor segment is projected onto
// the texture's daily calendar with anchorShape, the donor's own NAVs as the
// anchors and the texture as the shape. The result passes exactly through
// every real NAV and takes its day-to-day moves from the texture in between.
// Sparse donors therefore carry the level, and the engine carries the texture.
// A nil texture leaves every donor as it stands.
func DonorChain(f Fetcher, cashID, calibrate string, donors []Donor, from time.Time, texture *marketdata.Series) (*marketdata.Series, error) {
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
	for _, donor := range donors {
		d, err := f.Fetch(donor.ID, from)
		if err != nil || d == nil || len(d.Points) < 30 {
			continue
		}
		scaled, ok := volMatch(ref, d, cash)
		if !ok {
			continue
		}
		if donor.Uplift != 0 {
			scaled = afterAnnualFee(scaled.Name+" (fee-aligned)", scaled, -donor.Uplift)
		}
		if out == nil {
			out = densify(scaled, texture, time.Time{})
			continue
		}
		if !d.First().Date.Before(out.First().Date) {
			continue // adds no history
		}
		start := out.First().Date
		out.SimulatedBefore = time.Time{} // let ExtendBack splice one more segment
		marketdata.ExtendBack(out, densify(scaled, texture, start))
	}
	if out == nil {
		return nil, fmt.Errorf("donor chain %s: no usable donor", calibrate)
	}
	out.SimulatedBefore = time.Time{}
	return out, nil
}

// sparseSpacing is the median spacing, in calendar days, above which a donor
// segment is too coarse to be spliced into a daily file as it stands. It sits
// just above a daily calendar's own worst case (a long weekend) and well under
// a weekly dealing cycle, which is the cadence actually met in the wild.
const sparseSpacing = 3

// densify projects a sparse donor segment onto texture's calendar, leaving a
// daily one alone. before bounds the segment the chain will actually keep of
// this donor (its points strictly before that date); the zero time means all
// of them. See DonorChain for why, and shapedSeries/anchorShape for how.
func densify(donor, texture *marketdata.Series, before time.Time) *marketdata.Series {
	if texture == nil || !sparse(donor, before) {
		return donor
	}
	return shapedSeries(donor, texture)
}

// sparse reports whether the donor's points before the given date (all of them
// when it is zero) are spaced by more than sparseSpacing calendar days at the
// median. A handful of points cannot tell a cadence from a gap, so a segment
// shorter than half a year of weekly dealing is left alone.
func sparse(donor *marketdata.Series, before time.Time) bool {
	var gaps []float64
	for i := 1; i < len(donor.Points); i++ {
		if !before.IsZero() && !donor.Points[i].Date.Before(before) {
			break
		}
		gaps = append(gaps, donor.Points[i].Date.Sub(donor.Points[i-1].Date).Hours()/24)
	}
	if len(gaps) < 26 {
		return false
	}
	slices.Sort(gaps)
	return gaps[len(gaps)/2] > sparseSpacing
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

// eurCashReturn is the return of a money-market INDEX between two dates (the
// euro cash leg is an index level, not a rate: cashAccrual would misread it).
func eurCashReturn(idx *marketdata.Series, from, to time.Time) float64 {
	if idx == nil {
		return 0
	}
	a, _, okA := idx.At(from)
	b, _, okB := idx.At(to)
	if !okA || !okB || a <= 0 {
		return 0
	}
	return b/a - 1
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
