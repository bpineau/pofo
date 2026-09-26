package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// trackSeries builds a daily series from consecutive returns (fractions),
// starting at 100 on 2020-01-01. Weekends are ignored: what the tests need is a
// shared calendar, not a trading one.
func trackSeries(rets ...float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: "T"}
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	v := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: d, Close: v})
	for _, r := range rets {
		d = d.AddDate(0, 0, 1)
		v *= 1 + r
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: v})
	}
	return s
}

// flat is n sessions of nothing happening, the background every test needs on
// both sides of the event it is about.
func flat(n int) []float64 { return make([]float64, n) }

func withEvent(head []float64, event []float64, tail []float64) []float64 {
	out := append([]float64{}, head...)
	out = append(out, event...)
	return append(out, tail...)
}

// A crash both series lived through is history, not a defect, however violent.
func TestTrackIndexKeepsASharedCrash(t *testing.T) {
	crash := []float64{-0.12, -0.09, +0.11}
	donor := trackSeries(withEvent(flat(30), crash, flat(30))...)
	ref := trackSeries(withEvent(flat(30), crash, flat(30))...)

	got, rejects := trackIndex(donor, ref, 0.10)
	if len(rejects) != 0 {
		t.Fatalf("rejected %v on a crash both series made", rejects)
	}
	if got != donor {
		t.Error("a clean donor was rebuilt instead of returned as it came")
	}
}

// A one-sided step the reference never made is refused, and the level after it
// lands where the reference says it should.
func TestTrackIndexRefusesAOneSidedStep(t *testing.T) {
	donor := trackSeries(withEvent(flat(30), []float64{+0.25}, flat(30))...)
	ref := trackSeries(flat(61)...)

	got, rejects := trackIndex(donor, ref, 0.10)
	if len(rejects) != 1 {
		t.Fatalf("got %d rejections, want the one fabricated step: %v", len(rejects), rejects)
	}
	if rejects[0].Dropped {
		t.Error("an interior step was treated as a bad first print")
	}
	if last := got.Last().Close; math.Abs(last-100) > 1e-9 {
		t.Errorf("the repaired series ends at %.4f, want the reference's 100", last)
	}
}

// A patch of foreign prints (a step up, a month at the wrong level, a step back
// down) is stitched at both ends, and the days inside it keep their own moves.
func TestTrackIndexStitchesAPatch(t *testing.T) {
	inside := []float64{+0.25, +0.01, -0.02, +0.03, -0.20}
	donor := trackSeries(withEvent(flat(30), inside, flat(30))...)
	ref := trackSeries(withEvent(flat(30), []float64{0, +0.01, -0.02, +0.03, 0}, flat(30))...)

	got, rejects := trackIndex(donor, ref, 0.10)
	if len(rejects) != 2 {
		t.Fatalf("got %d rejections, want the patch's two ends: %v", len(rejects), rejects)
	}
	// Inside the patch the donor's own moves survive: the third of the five
	// sessions (points[33]) is -2 %.
	if r := got.Points[33].Close/got.Points[32].Close - 1; math.Abs(r+0.02) > 1e-9 {
		t.Errorf("a session inside the patch returns %+.4f, want the donor's -0.0200", r)
	}
	if last, want := got.Last().Close, ref.Last().Close; math.Abs(last-want) > 1e-9 {
		t.Errorf("the repaired series ends at %.4f, want %.4f", last, want)
	}
}

// A bad FIRST print is dropped rather than repaired: there is nothing before it
// to correct against, and replacing its step would keep the print and drag
// every later level with it.
func TestTrackIndexDropsABadFirstPrint(t *testing.T) {
	donor := trackSeries(withEvent([]float64{+0.20}, nil, flat(40))...)
	ref := trackSeries(flat(41)...)

	got, rejects := trackIndex(donor, ref, 0.10)
	if len(rejects) != 1 || !rejects[0].Dropped {
		t.Fatalf("got %v, want the first print dropped", rejects)
	}
	if len(got.Points) != len(donor.Points)-1 {
		t.Fatalf("got %d points, want one fewer than the donor's %d", len(got.Points), len(donor.Points))
	}
	if first := got.First().Close; math.Abs(first-120) > 1e-9 {
		t.Errorf("the repaired series starts at %.4f, want the donor's second print 120", first)
	}
	if first := got.First().Date; !first.Equal(donor.Points[1].Date) {
		t.Errorf("the repaired series starts on %s, want the donor's second date", first.Format("2006-01-02"))
	}
}

// A move the donor makes one session after the reference is a closing-time
// difference, not a defect, and survives however close to the tolerance it
// runs.
func TestTrackIndexAbsorbsAOneSessionClockLag(t *testing.T) {
	donor := trackSeries(withEvent(flat(30), []float64{0, +0.09}, flat(29))...)
	ref := trackSeries(withEvent(flat(30), []float64{+0.09, 0}, flat(29))...)

	if _, rejects := trackIndex(donor, ref, 0.08); len(rejects) != 0 {
		t.Errorf("rejected %v on a move the donor simply posted a session late", rejects)
	}
}

// A repeated close followed by a catch-up carrying two of the reference's
// sessions leaves no level error, so neither session is refused; a one-sided
// step of the same size next to it still is.
func TestTrackIndexAbsorbsAStalePrint(t *testing.T) {
	donor := trackSeries(withEvent(flat(30), []float64{0, -0.058, 0, 0, -0.03}, flat(29))...)
	ref := trackSeries(withEvent(flat(30), []float64{-0.015, -0.043, 0, 0, 0}, flat(29))...)

	_, rejects := trackIndex(donor, ref, 0.0115)
	if len(rejects) != 1 {
		t.Fatalf("got %v, want only the one-sided step refused", rejects)
	}
	if want := donor.Points[35].Date; !rejects[0].Date.Equal(want) {
		t.Errorf("refused %s, want the one-sided step on %s", rejects[0].Date.Format("2006-01-02"), want.Format("2006-01-02"))
	}
}

// Pooling two reference sessions needs a repeated close beside the donor
// session: next to an ordinary move, a step that happens to equal a
// two-session sum is still refused.
func TestTrackIndexPoolsOnlyAfterARepeatedClose(t *testing.T) {
	donor := trackSeries(withEvent(flat(30), []float64{-0.004, -0.058}, flat(29))...)
	ref := trackSeries(withEvent(flat(30), []float64{-0.015, -0.043}, flat(29))...)

	_, rejects := trackIndex(donor, ref, 0.0115)
	if len(rejects) != 1 || !rejects[0].Date.Equal(donor.Points[32].Date) {
		t.Fatalf("got %v, want the step after an ordinary session refused", rejects)
	}
}

// A session listed on the record is refused even under the tolerance, and says
// so; a listed date the donor does not quote changes nothing.
func TestTrackIndexRefusesOnTheRecord(t *testing.T) {
	donor := trackSeries(withEvent(flat(30), []float64{-0.008}, flat(30))...)
	ref := trackSeries(flat(61)...)
	listed := donor.Points[31].Date

	if _, rejects := trackIndex(donor, ref, 0.0115); len(rejects) != 0 {
		t.Fatalf("rejected %v under the tolerance with nothing listed", rejects)
	}
	got, rejects := trackIndex(donor, ref, 0.0115, listed, listed.AddDate(1, 0, 0))
	if len(rejects) != 1 || !rejects[0].Listed || !rejects[0].Date.Equal(listed) {
		t.Fatalf("got %v, want the listed session refused on the record", rejects)
	}
	if last := got.Last().Close; math.Abs(last-100) > 1e-9 {
		t.Errorf("the repaired series ends at %.4f, want the reference's 100", last)
	}
}

// Sessions the reference does not cover cannot convict anyone: the donor's own
// moves stand there, whatever they are.
func TestTrackIndexKeepsUngradedSessions(t *testing.T) {
	donor := trackSeries(withEvent(flat(30), []float64{+0.25}, flat(30))...)
	ref := trackSeries(flat(61)...)
	ref.Points = append(ref.Points[:29], ref.Points[33:]...) // no reference around the step

	got, rejects := trackIndex(donor, ref, 0.10)
	if len(rejects) != 0 {
		t.Fatalf("rejected %v without a reference covering the session", rejects)
	}
	if got != donor {
		t.Error("an ungradeable donor was rebuilt instead of returned as it came")
	}
}
