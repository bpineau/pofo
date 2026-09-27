package marketdata

import (
	"math"
	"time"
)

// This file holds the two data guards that read a series as DATA rather than
// as returns, because what they hunt is invisible to a statistic: a stretch
// of calendar the series does not cover (FindGaps), and a print no instrument
// could have made (FindSpikes). Each is ONE rule, shared by every caller: the
// data doctor (Verify) reports both, the fetch-time cleaner (dropRoundTrips)
// and pkg/simgen's shape despiker drop what their own SpikeRule finds, and
// the bundle's golden guards run both over every embedded file.

// Gap is a step between two consecutive quotes longer than the series' own
// pace allows: a stretch of calendar it does not cover. A consumer reading
// one row, then the next, computes one enormous "period" across it followed
// by none at all, so statistics computed over a gap are not noisy, they are
// wrong, and they look plausible.
type Gap struct {
	From, To time.Time // the quotes on either side
	Days     float64   // calendar days between them
	Limit    float64   // the longest step the series' local pace allows, in days
}

// Gap-rule constants.
const (
	// gapSlack is how many of its own local steps a series may take at once
	// before the step is a gap: half a step of slack absorbs every calendar
	// irregularity (a 31-day month after a 28-day one, a month-end on the last
	// trading day), while a step holding one whole missing observation is at
	// least two.
	gapSlack = 1.5
	// minGapDays is the floor of that limit, in calendar days: a daily series
	// is not holed by an exchange closed for a week (New Year, the week after
	// 2001-09-11), which a limit of a day and a half would convict.
	minGapDays = 14
)

// FindGaps returns the steps of s longer than its pace allows, in date order.
// The pace is LOCAL (the median step among the twenty neighbouring ones, as
// Verify reads it), so a series that reported monthly for eighty years
// before turning daily is judged by the pace it kept at the time, and a gap
// is a step beyond max(14 days, 1.5 local steps): a monthly series that
// skips a month (a 59-day step against a 45-day limit), a weekly NAV that
// skips two weeks, a daily series silent for three weeks. A single hole
// cannot move the median it is judged against.
//
// The rule is the one the data doctor reports ("no quotes for N days") and
// the one the bundle's golden guard refuses in any embedded file. It reads
// dates only, so it applies to rate levels as well as prices. It returns nil
// for a nil series or one of fewer than two points.
func FindGaps(s *Series) []Gap {
	if s.Len() < 2 {
		return nil
	}
	pace := localSpacingDays(s.Points)
	var out []Gap
	for k := 1; k < len(s.Points); k++ {
		days := s.Points[k].Date.Sub(s.Points[k-1].Date).Hours() / 24
		if limit := math.Max(minGapDays, gapSlack*pace[k]); days > limit {
			out = append(out, Gap{From: s.Points[k-1].Date, To: s.Points[k].Date, Days: days, Limit: limit})
		}
	}
	return out
}

// Spike is a one-session round trip: a print that leaves the level and comes
// straight back, between two neighbours that agree with each other.
type Spike struct {
	Date    time.Time // the suspect print
	In, Out float64   // the returns into and out of it, fractions of opposite signs
	Sigma   float64   // standard deviation of the returns around it, the suspect pair left out
}

// SpikeRule is the test a round trip must pass to be called a print no
// instrument made. Every rule shares the two clauses that make a real market
// day safe: the legs point opposite ways and EACH spans more than six local
// standard deviations (over the 25 returns on either side, the suspect pair
// excluded, so a bad print cannot inflate the yardstick meant to convict
// it; a crash raises the local sigma with it). The fields add the clauses
// that differ with what the caller does about a finding.
type SpikeRule struct {
	// MinLeg is the smallest |return| each leg must reach, a fraction: below
	// it the candidates are cash-like series whose local sigma is a rounding
	// error. Zero means no floor.
	MinLeg float64
	// MaxNet, when positive, is how far the round trip may leave the level
	// once both legs are compounded: |(1 + In)(1 + Out) - 1| at most MaxNet.
	MaxNet float64
	// MaxNetShare, when positive, bounds the same net move RELATIVE to the
	// smaller leg: strictly below MaxNetShare times min(|In|, |Out|).
	MaxNetShare float64
}

// The shared clauses of every SpikeRule.
const (
	// spikeZ is how many local standard deviations each leg must span.
	spikeZ = 6.0
	// spikeWindow is how many returns on each side define "local": five weeks
	// of trading either way, long enough for a stable sigma and short enough
	// to follow a regime change.
	spikeWindow = 25
)

// FindSpikes returns the one-session round trips of s that no instrument
// could have made, in date order: legs of opposite signs, each at least 2 %
// and beyond six local standard deviations, cancelling to within a third of
// the smaller leg. That is the rule the bundle's golden guard applies to
// every embedded file and the data doctor reports; the equity sessions of
// 1987-10-19, October 2008 and March 2020 clear none of its last two
// clauses, though a credit fund's March 2020 round trip can (the doctor
// names it, a sibling listing settles it). A rate
// level (^IRX, ^ESTR...) has no spikes to find, since a ratio of two rates
// is not a return, and a series shorter than 50 points too few to judge.
//
// Finding is not dropping: the fetch-time cleaner drops only the round trips
// its own, stricter rule proves (a full reversal, within 2 %, of legs beyond
// the asset class's floor), so a fetched series can still carry a finding
// here, for a reader to check against a sibling listing.
func FindSpikes(s *Series) []Spike {
	return artefactSpikes.Find(s)
}

// artefactSpikes is FindSpikes' rule: what shipped data must never carry.
var artefactSpikes = SpikeRule{MinLeg: 0.02, MaxNetShare: 1.0 / 3}

// Find returns the round trips of s that r convicts, in date order. It
// returns nil for a nil series, one shorter than 50 points and a rate level;
// a point next to a non-positive close is never judged.
func (r SpikeRule) Find(s *Series) []Spike {
	if s == nil || isRateSymbol(s.Symbol) || isPolicyRate(s.Symbol) {
		return nil
	}
	spikes, _ := r.findAt(s.Points)
	return spikes
}

// Drop returns s as a new Series without the prints r convicts (its points
// share s's array when none is dropped): each dropped
// point's neighbours then meet directly, which carries the true two-session
// move, and every point is judged on s as given, so two adjacent convictions
// both go. It is what a pass that REPAIRS rather than reports runs (the
// fetch-time cleaner, pkg/simgen's shape despiker). A nil series, a rate
// level and a series shorter than 50 points come back unchanged.
func (r SpikeRule) Drop(s *Series) *Series {
	if s == nil {
		return nil
	}
	out := *s
	if !isRateSymbol(s.Symbol) && !isPolicyRate(s.Symbol) {
		out.Points = r.drop(s.Points)
	}
	return &out
}

// findAt returns the round trips r convicts in pts, with the index of each.
func (r SpikeRule) findAt(pts []Point) (out []Spike, at []int) {
	n := len(pts)
	if n < 2*spikeWindow {
		return nil, nil
	}
	for i := 1; i+1 < n; i++ {
		if pts[i-1].Close <= 0 || pts[i].Close <= 0 || pts[i+1].Close <= 0 {
			continue
		}
		in, out2 := pts[i].Close/pts[i-1].Close-1, pts[i+1].Close/pts[i].Close-1
		smaller := math.Min(math.Abs(in), math.Abs(out2))
		net := math.Abs((1+in)*(1+out2) - 1)
		switch {
		case in*out2 >= 0, smaller < r.MinLeg:
			continue
		case r.MaxNet > 0 && net > r.MaxNet:
			continue
		case r.MaxNetShare > 0 && net >= r.MaxNetShare*smaller:
			continue
		}
		sigma, ok := localSigma(pts, i)
		if !ok || smaller <= spikeZ*sigma {
			continue
		}
		out = append(out, Spike{Date: pts[i].Date, In: in, Out: out2, Sigma: sigma})
		at = append(at, i)
	}
	return out, at
}

// drop returns pts minus the round trips r convicts. Each dropped point's
// neighbours then meet directly, which carries the true two-session move.
// Every point is judged on the original series, so two adjacent convictions
// are both dropped.
func (r SpikeRule) drop(pts []Point) []Point {
	_, at := r.findAt(pts)
	if len(at) == 0 {
		return pts
	}
	out := make([]Point, 0, len(pts)-len(at))
	k := 0
	for i, p := range pts {
		if k < len(at) && i == at[k] {
			k++
			continue
		}
		out = append(out, p)
	}
	return out
}

// localSigma is the standard deviation of the returns around index i, with
// the suspect pair (the returns into and out of i) excluded so a bad print
// cannot inflate the very yardstick meant to convict it. ok is false when too
// few returns surround i for the estimate to mean anything.
func localSigma(pts []Point, i int) (float64, bool) {
	var sum, sumsq float64
	count := 0
	for j := max(1, i-spikeWindow); j <= min(len(pts)-1, i+spikeWindow); j++ {
		if j == i || j == i+1 || pts[j-1].Close <= 0 || pts[j].Close <= 0 {
			continue
		}
		r := pts[j].Close/pts[j-1].Close - 1
		sum += r
		sumsq += r * r
		count++
	}
	if count < 10 {
		return 0, false
	}
	mean := sum / float64(count)
	variance := sumsq/float64(count) - mean*mean
	if variance <= 0 {
		return 0, false
	}
	return math.Sqrt(variance), true
}
