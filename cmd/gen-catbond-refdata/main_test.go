package main

import (
	"math"
	"strings"
	"testing"
	"time"
)

// doc is the published document in miniature: three years of monthly returns
// in PERCENT, keyed by year, plus the summary statistics the generator checks
// itself against. The numbers are chosen so every published figure is exactly
// reproducible from the table, which is the property the real download must
// also have.
func doc() source {
	return source{
		MonthlyReturns: map[string][]float64{
			"2006": {1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			"2007": {1, 1, 1, 1, 1, 1, 1, 1, -5, 1, 1, 1},
			"2008": {1, 1, 1},
		},
		RiskReturnMetrics: map[string]string{
			"Maximum Drawdown":              "-5.00%",
			"Best Monthly Return":           "1.00%",
			"Worst Monthly Return":          "-5.00%",
			"Annualised Standard Deviation": "4.00%",
			"Percentage of Positive Months": "96.30%",
		},
		KeyStatistics: map[string]string{
			"Last 3 Months Return":    "3.03%",
			"Last 3 Years Annualised": "9.66%",
			"Last 5 Years Annualised": "9.66%",
			"Annualised Return":       "9.66%",
		},
	}
}

func TestParse(t *testing.T) {
	months, err := parse(doc())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(months) != 27 {
		t.Fatalf("got %d months, want 27", len(months))
	}
	if got, want := months[0].date, monthEndOf(2006, time.January); !got.Equal(want) {
		t.Errorf("first month %s, want %s", got, want)
	}
	if got, want := months[len(months)-1].date, monthEndOf(2008, time.March); !got.Equal(want) {
		t.Errorf("last month %s, want %s", got, want)
	}
	// The source quotes percent; the generator must hand over fractions.
	if math.Abs(months[0].ret-0.01) > 1e-12 {
		t.Errorf("first return %.6f, want 0.01", months[0].ret)
	}
	if math.Abs(months[20].ret+0.05) > 1e-12 {
		t.Errorf("2007-09 return %.6f, want -0.05", months[20].ret)
	}
}

// A table quoted in fractions rather than percent (0.01 for +1 %) would ship an
// index that never moves; only the reverse mistake is detectable, and it is.
func TestParseRejectsImplausibleReturn(t *testing.T) {
	d := doc()
	d.MonthlyReturns["2006"][0] = 80 // +80 % in a month
	if _, err := parse(d); err == nil {
		t.Fatal("parsed a monthly return of 80 %")
	}
}

func TestParseRejectsThirteenMonths(t *testing.T) {
	d := doc()
	d.MonthlyReturns["2006"] = append(d.MonthlyReturns["2006"], 1)
	if _, err := parse(d); err == nil {
		t.Fatal("parsed a year with thirteen months")
	}
}

func TestCheckAgainstSource(t *testing.T) {
	months, err := parse(doc())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := checkAgainstSource(months, doc()); err != nil {
		t.Fatalf("cross-check: %v", err)
	}
}

// The whole point of the cross-check: a table that no longer reproduces the
// statistics printed beside it is a table that changed meaning.
func TestCheckAgainstSourceRejectsDisagreement(t *testing.T) {
	months, err := parse(doc())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	d := doc()
	d.RiskReturnMetrics["Worst Monthly Return"] = "-2.00%"
	err = checkAgainstSource(months, d)
	if err == nil || !strings.Contains(err.Error(), "Worst Monthly Return") {
		t.Fatalf("err = %v, want a Worst Monthly Return mismatch", err)
	}
}

func TestCheckAgainstSourceRejectsMissingStatistic(t *testing.T) {
	months, err := parse(doc())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	d := doc()
	delete(d.RiskReturnMetrics, "Maximum Drawdown")
	if err := checkAgainstSource(months, d); err == nil {
		t.Fatal("accepted a document with no maximum drawdown to check against")
	}
}

// check guards the shipped file: a truncated download, a series at the wrong
// volatility, or one with no event loss in it at all.
func TestCheckBands(t *testing.T) {
	long := make([]point, 0, 240)
	for i := range 240 {
		r := 0.004
		if i%12 == 5 {
			r = -0.004 // enough seasonal swing to reach an ILS volatility
		}
		if i == 100 {
			r = -0.06
		}
		long = append(long, point{monthEndOf(2006+i/12, time.Month(i%12+1)), r})
	}
	if err := check(long); err != nil {
		t.Fatalf("rejected a plausible series: %v", err)
	}
	if err := check(long[:100]); err == nil {
		t.Fatal("accepted a truncated series")
	}
	flat := make([]point, len(long))
	copy(flat, long)
	flat[100].ret = 0.004
	if err := check(flat); err == nil {
		t.Fatal("accepted an ILS series with no event loss in it")
	}
	gapped := make([]point, len(long))
	copy(gapped, long)
	gapped[50].date = monthEndOf(2050, time.January)
	if err := check(gapped); err == nil {
		t.Fatal("accepted a series with a hole in it")
	}
}
