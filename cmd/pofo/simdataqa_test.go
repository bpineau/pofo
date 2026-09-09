package main

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/simgen"
)

// The three number formatters of the QA page all say "-" for a statistic that
// could not be measured, and the signed one always carries its sign: a report
// that printed 0.00 % for "unknown" would read as a perfect match.
func TestQANumberFormats(t *testing.T) {
	if pctOf(math.NaN()) != "-" || signedPct(math.NaN()) != "-" || num2(math.NaN()) != "-" {
		t.Error("an unmeasured statistic did not print as a dash")
	}
	if got := pctOf(0.0812); got != "8.12 %" {
		t.Errorf("pctOf(0.0812) = %q", got)
	}
	if got := signedPct(0.0812); got != "+8.12 %" {
		t.Errorf("signedPct(0.0812) = %q", got)
	}
	if got := signedPct(-0.0812); got != "-8.12 %" {
		t.Errorf("signedPct(-0.0812) = %q", got)
	}
	if got := num2(0.876); got != "0.88" {
		t.Errorf("num2(0.876) = %q", got)
	}
}

// The page names an ISIN the way a reader knows it, and says nothing extra
// when the identifier is already the name.
func TestQALabel(t *testing.T) {
	if got := qaLabel("IE00B3XXRP09"); got == "" {
		t.Error("a catalog ISIN got no reader-facing label")
	}
	if got := qaLabel("NOSUCHASSET"); got != "" {
		t.Errorf("an unknown identifier got the label %q", got)
	}
	// An identifier that is its own label adds nothing to the heading.
	for id, label := range qaLabels {
		if strings.EqualFold(id, label) {
			if got := qaLabel(id); got != "" {
				t.Errorf("%s is its own label yet printed %q", id, got)
			}
			break
		}
	}
}

// A recipe the audit could not measure keeps its identity and its reason, and
// no statistic is invented for it.
func TestQACardOfUnmeasured(t *testing.T) {
	c := qaCardOf(simgen.AuditResult{
		ID: "DBMF", Name: "DBi Managed Futures", Method: "donor chain",
		Err: "no independent reference", Rejected: []string{"KMLM: too correlated", "CTA: too short"},
		Level: simgen.VerdictUnknown, Path: simgen.VerdictUnknown,
	})
	if c.Have {
		t.Error("a card with no measurement claims one")
	}
	if c.Anchor != "dbmf" {
		t.Errorf("anchor = %q, want a lower-case slug", c.Anchor)
	}
	if c.Rejected != "KMLM: too correlated, CTA: too short" {
		t.Errorf("the refused references were not listed: %q", c.Rejected)
	}
	if c.CAGRSim != "" || c.Chart != "" {
		t.Error("an unmeasured card carries statistics or a chart")
	}
}

// A measured audit fills every column of the card: window, levels, texture,
// the two charts and the donor-chain junctions.
func TestQACardOfMeasured(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2020, 1, d, 0, 0, 0, 0, time.UTC) }
	mk := func(closes ...float64) *marketdata.Series {
		s := &marketdata.Series{}
		for i, c := range closes {
			s.Points = append(s.Points, marketdata.Point{Date: day(i + 1), Close: c})
		}
		return s
	}
	a := simgen.AuditResult{
		ID: "IE00BSPLC413", Name: "SPDR MSCI USA Small Value", Method: "index leg",
		Reference: "ZPRV", RealFrom: day(1),
		Start: day(1), End: day(5), Years: 4.2,
		DailyCorr: 0.91, WeeklyCorr: 0.94, MonthlyCorr: 0.97, Beta: 1.02,
		TrackingErr: 0.03, CAGRSim: 0.08, CAGRReal: 0.075, Delta: 0.005, TotalDrift: 0.02,
		VolSim: 0.15, VolReal: 0.14, WorstSim: -0.07, WorstReal: -0.06,
		Level: simgen.VerdictOK, Path: simgen.VerdictWarn,
		Engine: mk(100, 101, 102, 103, 104),
		Real:   mk(100, 100.5, 101.5, 102.5, 104),
		Others: []*marketdata.Series{mk(100, 100.2, 100.4, 100.6, 100.8)},
		Chain: []simgen.Junction{
			{Span: "1996-2005", Pair: "donor vs fund", Months: 118, Corr: 0.88, GapYear: -0.012, Measured: true},
			{Span: "before 1996", Pair: "engine vs donor", Note: "no common month", Measured: false},
		},
		Caveat: "the donor is weekly-dealing",
	}
	c := qaCardOf(a)
	if !c.Have || c.Label == "" {
		t.Fatalf("a measured card came back empty: %+v", c)
	}
	if c.Window != "4.2 y (2020-01 to 2020-01)" {
		t.Errorf("window = %q", c.Window)
	}
	if c.RealFrom != "2020-01-01" {
		t.Errorf("RealFrom = %q", c.RealFrom)
	}
	if c.Delta != "+0.50 %" || c.Drift != "+2.00 %" {
		t.Errorf("delta/drift = %q / %q", c.Delta, c.Drift)
	}
	if c.Vols != "15.00 % / 14.00 %" || c.Worst != "-7.00 % / -6.00 %" {
		t.Errorf("vols/worst = %q / %q", c.Vols, c.Worst)
	}
	if c.TEVol != "0.21" {
		t.Errorf("tracking error over vol = %q, want 0.21", c.TEVol)
	}
	if c.Daily != "0.91" || c.Monthly != "0.97" || c.Beta != "1.02" {
		t.Errorf("texture columns = %q %q %q", c.Daily, c.Monthly, c.Beta)
	}
	if !strings.Contains(string(c.Chart), "<svg") || !strings.Contains(string(c.DriftChart), "<svg") {
		t.Error("a measured card is missing one of its two charts")
	}
	if len(c.Chain) != 2 {
		t.Fatalf("%d junctions, want 2", len(c.Chain))
	}
	if c.Chain[0].Corr != "0.88" || c.Chain[0].Gap != "-1.2 pt/yr" {
		t.Errorf("measured junction = %+v", c.Chain[0])
	}
	if c.Chain[1].Corr != "-" || c.Chain[1].Gap != "-" {
		t.Errorf("an unmeasurable junction shows numbers: %+v", c.Chain[1])
	}
	if c.Caveat != a.Caveat {
		t.Errorf("the hand-written caveat was dropped")
	}
}

// A real series flat over the window leaves the tracking-error ratio
// undefined rather than dividing by its zero volatility. A measured audit
// always carries both series (see simgen.Audit), so the card may rely on it.
func TestQACardOfZeroVol(t *testing.T) {
	flat := &marketdata.Series{Points: []marketdata.Point{
		{Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Close: 100},
		{Date: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), Close: 100},
	}}
	c := qaCardOf(simgen.AuditResult{ID: "X", TrackingErr: 0.02, Engine: flat, Real: flat})
	if c.TEVol != "-" {
		t.Errorf("TEVol = %q with no real volatility, want a dash", c.TEVol)
	}
}
