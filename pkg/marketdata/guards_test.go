package marketdata

import (
	"math"
	"strings"
	"testing"
	"time"
)

// FindGaps judges each step against the pace the series kept around it: a
// month skipped by a monthly series, three silent weeks in a daily one, and
// nothing in a series that turned from monthly to daily.
func TestFindGaps(t *testing.T) {
	monthEnd := func(y int, m time.Month) time.Time { return date(y, m+1, 0) }
	var monthly []Point
	for m := time.January; m <= time.December; m++ {
		if m == time.June {
			continue // a skipped month: May 31 to July 31 is 61 days
		}
		monthly = append(monthly, Point{Date: monthEnd(2023, m), Close: 100 + float64(m)})
	}
	gaps := FindGaps(&Series{Points: monthly})
	if len(gaps) != 1 || !gaps[0].From.Equal(date(2023, 5, 31)) || !gaps[0].To.Equal(date(2023, 7, 31)) || gaps[0].Days != 61 {
		t.Fatalf("monthly gaps = %+v, want the one skipped month", gaps)
	}
	if gaps[0].Limit < 45 || gaps[0].Limit > 47 {
		t.Errorf("monthly limit = %.1f days, want one and a half months", gaps[0].Limit)
	}

	// Monthly for two years, then daily: neither pace is a gap of the other.
	var mixed []Point
	for m := 0; m < 24; m++ {
		mixed = append(mixed, Point{Date: monthEnd(2020, time.January+time.Month(m)), Close: 100})
	}
	for d := 1; d <= 60; d++ {
		if day := date(2022, 1, d); day.Weekday() != time.Saturday && day.Weekday() != time.Sunday {
			mixed = append(mixed, Point{Date: day, Close: 100})
		}
	}
	if gaps := FindGaps(&Series{Points: mixed}); len(gaps) != 0 {
		t.Errorf("a change of pace read as gaps: %+v", gaps)
	}

	// A daily series: a closed week passes, three silent weeks do not.
	var daily []Point
	for d := 0; d < 120; d++ {
		day := date(2024, 1, 1).AddDate(0, 0, d)
		if (d >= 30 && d < 37) || (d >= 70 && d < 91) {
			continue
		}
		daily = append(daily, Point{Date: day, Close: 100})
	}
	gaps = FindGaps(&Series{Points: daily})
	if len(gaps) != 1 || gaps[0].Days != 22 || gaps[0].Limit != minGapDays {
		t.Errorf("daily gaps = %+v, want the three weeks only, against the 14-day floor", gaps)
	}

	if FindGaps(nil) != nil || FindGaps(&Series{Points: monthly[:1]}) != nil {
		t.Error("a nil or one-point series has gaps")
	}
}

// FindSpikes convicts the lone fabricated print and leaves a crash, a bounce
// and a rate level alone.
func TestFindSpikes(t *testing.T) {
	// iemlLocalGovt2015: +21.9 % / -17.6 % against a quiet neighbourhood.
	s := &Series{Symbol: "IE00B5M4WH52", Points: iemlLocalGovt2015.series()}
	got := FindSpikes(s)
	if len(got) != 1 || !got[0].Date.Equal(s.Points[iemlLocalGovt2015.at].Date) {
		t.Fatalf("spikes = %+v, want the one at index %d", got, iemlLocalGovt2015.at)
	}
	in, out := iemlLocalGovt2015.legs()
	if math.Abs(got[0].In*100-in) > 1e-9 || math.Abs(got[0].Out*100-out) > 1e-9 || !(got[0].Sigma > 0) {
		t.Errorf("spike = %+v, want legs %+.2f / %+.2f and a sigma", got[0], in, out)
	}

	// A crash that does not come back is history, not a print.
	crash := iemlLocalGovt2015
	crash.closes = append([]float64(nil), crash.closes...)
	for i := crash.at; i < len(crash.closes); i++ {
		crash.closes[i] *= 0.7
	}
	if got := FindSpikes(&Series{Points: crash.series()}); len(got) != 0 {
		t.Errorf("a crash read as a spike: %+v", got)
	}
	// A rate level has no returns to judge.
	if got := FindSpikes(&Series{Symbol: "^IRX", Points: iemlLocalGovt2015.series()}); got != nil {
		t.Errorf("a rate level has spikes: %+v", got)
	}
	if FindSpikes(nil) != nil || FindSpikes(&Series{Points: iemlLocalGovt2015.series()[:40]}) != nil {
		t.Error("a nil or short series has spikes")
	}
}

// One mechanism, several bars: the fetch-time rule (a full reversal, within
// 2 %) declines a round trip that leaves 3 % standing, which the artefact
// rule (within a third of the smaller leg) convicts; Drop removes exactly
// what Find reports and keeps the metadata.
func TestSpikeRuleBars(t *testing.T) {
	c := append([]float64(nil), iemlLocalGovt2015.closes...)
	at := iemlLocalGovt2015.at
	c[at] = c[at-1] * 1.20
	c[at+1] = c[at] * (1 / 1.20 * 1.03) // back to 3 % above where it left
	for i := at + 2; i < len(c); i++ {
		c[i] *= 1.03
	}
	s := &Series{Symbol: "X", Name: "kept", Points: pts(c...)}
	strict := SpikeRule{MinLeg: 0.02, MaxNet: roundTripNet}
	if got := strict.Find(s); len(got) != 0 {
		t.Errorf("the 2 %% bar convicted a 3 %% net: %+v", got)
	}
	if got := FindSpikes(s); len(got) != 1 {
		t.Errorf("the third-of-a-leg bar missed it: %+v", got)
	}
	d := artefactSpikes.Drop(s)
	if d.Len() != s.Len()-1 || d.Name != "kept" || s.Len() != len(c) {
		t.Errorf("Drop = %d points (%q), source %d, want one dropped and the metadata kept", d.Len(), d.Name, s.Len())
	}
	if (SpikeRule{MinLeg: 0.5}).Find(s) != nil {
		t.Error("a floor above both legs still convicted")
	}
	if artefactSpikes.Drop(nil) != nil {
		t.Error("Drop(nil) is not nil")
	}
}

// The doctor reports both guards.
func TestVerifyReportsGapsAndSpikes(t *testing.T) {
	s := &Series{Symbol: "IE00B5M4WH52", Points: iemlLocalGovt2015.series()}
	s.Points = append(s.Points[:10:10], s.Points[40:]...) // 30 days missing
	var gaps, spikes int
	for _, is := range Verify(s, s.Last().Date) {
		switch {
		case strings.Contains(is.Message, "no quotes for 31 days"):
			gaps++
		case strings.Contains(is.Message, "a print no instrument made"):
			spikes++
		}
	}
	if gaps != 1 {
		t.Errorf("%d gap findings, want 1", gaps)
	}
	s = &Series{Symbol: "IE00B5M4WH52", Points: iemlLocalGovt2015.series()}
	for _, is := range Verify(s, s.Last().Date) {
		if strings.Contains(is.Message, "a print no instrument made") {
			spikes++
		}
	}
	if spikes != 1 {
		t.Errorf("%d spike findings, want 1", spikes)
	}
}
