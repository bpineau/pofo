package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Holding a donor to the index it tracks.
//
// A reconstruction leans on real quotes of real funds, and a provider's quote
// line is not always the fund's. The line can carry a print nobody traded, or a
// patch of prints from the same fund's OTHER listing, and a backcast does not
// merely inherit such a print: it multiplies it by the leg's weight and ships
// it as the fund's own history. The hygiene passes in pkg/marketdata catch what
// can be caught from the series alone (a level that leaves and comes straight
// back, a leading placeholder, a plain change of units); what they cannot do is
// know what the series was SUPPOSED to do.
//
// trackIndex is for the case where the repository does know, because the donor
// and a bundled reference track the SAME INDEX in the SAME currency. Then the
// donor's return, session by session, has to be the reference's return, and any
// session where it is not by a wide margin is a defect of the quote line rather
// than a fact about the fund.
//
// The comparison allows the reference to LEAD OR LAG the donor by one session,
// which is what absorbs the difference in closing times. A Xetra line closes at
// 17:30 CET and a world index is struck after New York; a big US move therefore
// lands in the donor a session late, and a rule reading one session at a time
// would call that a defect. Measured over the four Xtrackers MSCI World lines
// the FCPE recipe reads (11 920 sessions against the bundled MSCI World EUR
// path), that allowance is what separates the two populations: the largest
// disagreement with no defect behind it is 3.50 %, the smallest defect is
// 14.51 %, and the band between them is empty. A one-session rule instead lets
// a stale Xetra print reach 7.4 %, half the distance to the smallest real
// defect.
//
// The tolerance is therefore NOT a package constant: it belongs to the pair.
// A Xetra ETF against a US-close index and a Treasury mutual fund against a
// constant-maturity par bond do not have the same noise floor, and each call
// site states its own measurement.
//
// Two actions, and only two:
//
//   - An interior session's return is replaced by the reference's return over
//     the same two dates. The level after the session is corrected against the
//     level before it, which is right whether the defect is one fabricated
//     print (the next session's rejection puts the level back) or the start of
//     a contaminated patch (the session closing the patch puts it back).
//   - A rejected FIRST session drops the donor's first point instead, because
//     there is nothing before it to correct against. Replacing its return would
//     keep the bad print and drag every later level with it.
//
// What it deliberately does NOT do is decide that a donor's whole early segment
// is the wrong series. That judgement needs evidence the session-by-session
// test does not have (see eresMondeLegs, where a donor's pre-2009 segment is
// refused because its excess return over the reference IS the EUR/USD move, at
// a correlation of 0.998), and it is made explicitly, dated and measured, in
// the recipe.

// trackReject is one session a donor was refused on, kept so the generation log
// can say what was rejected and why rather than silently repairing it.
type trackReject struct {
	Date    time.Time // the session's closing date
	Donor   float64   // the donor's return that session
	Ref     float64   // the reference's return over the same two dates
	Excess  float64   // the clock-tolerant excess that convicted it (log, absolute)
	Dropped bool      // the first print was dropped rather than the return replaced
}

func (r trackReject) String() string {
	what := fmt.Sprintf("return replaced by the reference's (%+.2f %% -> %+.2f %%)", r.Donor*100, r.Ref*100)
	if r.Dropped {
		what = fmt.Sprintf("first print dropped (its step was %+.2f %% against the reference's %+.2f %%)", r.Donor*100, r.Ref*100)
	}
	return fmt.Sprintf("%s: %s, excess %.2f %%", r.Date.Format("2006-01-02"), what, r.Excess*100)
}

// trackIndex grades a donor series against a reference tracking the same index
// in the same currency, and repairs the sessions where the two disagree by more
// than tol (a fraction, on log returns, after allowing the reference to lead or
// lag by one session). It returns the repaired series and what it rejected; the
// donor is left untouched, and when nothing is rejected the original series is
// returned unchanged rather than rebuilt, so a clean donor never picks up
// rounding noise.
//
// Sessions the reference does not cover are kept as they are: a donor cannot be
// convicted on evidence that does not exist.
func trackIndex(donor, reference *marketdata.Series, tol float64) (*marketdata.Series, []trackReject) {
	if donor == nil || reference == nil || len(donor.Points) < 3 || len(reference.Points) < 3 {
		return donor, nil
	}
	byDate := make(map[time.Time]float64, len(reference.Points))
	for _, p := range reference.Points {
		byDate[p.Date] = p.Close
	}

	// Donor and reference returns on the DONOR's calendar, NaN where the
	// reference does not quote both ends of the session.
	n := len(donor.Points)
	dr := make([]float64, n)
	rr := make([]float64, n)
	for i := 1; i < n; i++ {
		dr[i] = math.Log(donor.Points[i].Close / donor.Points[i-1].Close)
		a, okA := byDate[donor.Points[i-1].Date]
		b, okB := byDate[donor.Points[i].Date]
		if !okA || !okB || a <= 0 || b <= 0 {
			rr[i] = math.NaN()
			continue
		}
		rr[i] = math.Log(b / a)
	}

	// The clock allowance: the reference may run one session ahead of the donor
	// or one behind it, so a session is convicted only when it disagrees with
	// all three.
	excess := func(i int) float64 {
		if math.IsNaN(rr[i]) {
			return 0
		}
		e := math.Abs(dr[i] - rr[i])
		for _, k := range []int{-1, 1} {
			if j := i + k; j >= 1 && j < n && !math.IsNaN(rr[j]) {
				e = math.Min(e, math.Abs(dr[i]-rr[j]))
			}
		}
		return e
	}

	// The first session is judged first and alone: nothing precedes the first
	// print, so a rejected first session convicts that print rather than the
	// one after it, and the only sound repair is to drop it and start again on
	// what is left.
	if e := excess(1); e > tol {
		head := trackReject{
			Date: donor.Points[0].Date, Donor: math.Expm1(dr[1]), Ref: math.Expm1(rr[1]),
			Excess: e, Dropped: true,
		}
		trimmed := *donor
		trimmed.Points = donor.Points[1:]
		out, rest := trackIndex(&trimmed, reference, tol)
		return out, append([]trackReject{head}, rest...)
	}

	var rejects []trackReject
	repaired := make([]float64, n)
	for i := 1; i < n; i++ {
		repaired[i] = dr[i]
		if e := excess(i); e > tol {
			rejects = append(rejects, trackReject{
				Date: donor.Points[i].Date, Donor: math.Expm1(dr[i]), Ref: math.Expm1(rr[i]), Excess: e,
			})
			repaired[i] = rr[i]
		}
	}
	if len(rejects) == 0 {
		return donor, nil
	}
	out := *donor
	out.Points = make([]marketdata.Point, n)
	out.Points[0] = donor.Points[0]
	for i := 1; i < n; i++ {
		out.Points[i] = marketdata.Point{
			Date:  donor.Points[i].Date,
			Close: out.Points[i-1].Close * math.Exp(repaired[i]),
		}
	}
	return &out, rejects
}
