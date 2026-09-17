package main

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// body mimics the source file: prose, a blank line, the column header, the
// daily rows in percent, then a copyright line.
const body = `This file was created by using the 202607 CRSP database.
The Tbill return is the simple daily rate.

,Mkt-RF,SMB,HML,RF
19260701,    0.09,   -0.23,   -0.28,    0.01
19260702,    0.45,   -0.34,   -0.03,    0.01
19260706,   -0.71,    0.44,    0.58,    0.01

Copyright 2026 Eugene F. Fama and Kenneth R. French
`

func TestCumulate(t *testing.T) {
	got, err := cumulate(body)
	if err != nil {
		t.Fatalf("cumulate: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("%d points, want 3 (the prose, header and copyright lines are not days)", len(got))
	}
	// The market's total return is the excess return plus the bill's, so the
	// level compounds (1+0.0010)(1+0.0046)(1-0.0070) from 100.
	want := 100 * 1.0010 * 1.0046 * 0.9930
	if math.Abs(got[2].Close-want) > 1e-9 {
		t.Errorf("last level %.9f, want %.9f", got[2].Close, want)
	}
	if d := got[0].Date; !d.Equal(time.Date(1926, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("first date %v, want 1926-07-01 UTC", d)
	}
}

func TestCumulateRefusesBadSource(t *testing.T) {
	if _, err := cumulate(strings.ReplaceAll(body, "0.45", "-99.99")); err == nil {
		t.Error("a missing-value sentinel was accepted")
	}
	if _, err := cumulate(strings.ReplaceAll(body, ",Mkt-RF,SMB,HML,RF", ",A,B,C,D")); err == nil {
		t.Error("a source with no named market column was accepted")
	}
}

// ramp is n consecutive daily points growing at a constant rate and ending
// today, the shape checkShape accepts.
func ramp(n int, annual float64) []marketdata.Point {
	daily := math.Pow(1+annual, 1.0/252)
	out := make([]marketdata.Point, n)
	start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(n - 1))
	v := 100.0
	for i := range out {
		out[i] = marketdata.Point{Date: start.AddDate(0, 0, i), Close: v}
		v *= daily
	}
	return out
}

// deepEnough is a calendar-day count reaching from before 1926-08 to today.
var deepEnough = int(time.Since(time.Date(1926, 7, 1, 0, 0, 0, 0, time.UTC)).Hours()/24) + 1

func TestCheckShape(t *testing.T) {
	// A calendar-day ramp long enough to reach the library's own first month.
	ok := ramp(deepEnough, 0.10)
	if err := checkShape(ok); err != nil {
		t.Fatalf("a clean ramp was refused: %v", err)
	}
	cases := map[string]func([]marketdata.Point) []marketdata.Point{
		"too short": func(p []marketdata.Point) []marketdata.Point { return p[:100] },
		"out of order": func(p []marketdata.Point) []marketdata.Point {
			p[10].Date = p[9].Date
			return p
		},
		"impossible day": func(p []marketdata.Point) []marketdata.Point {
			p[10].Close = p[9].Close * 1.5
			return p
		},
		"stale": func(p []marketdata.Point) []marketdata.Point {
			for i := range p {
				p[i].Date = p[i].Date.AddDate(-3, 0, 0)
			}
			return p
		},
		"non-positive": func(p []marketdata.Point) []marketdata.Point {
			p[10].Close = 0
			return p
		},
	}
	for name, breakIt := range cases {
		if err := checkShape(breakIt(ramp(deepEnough, 0.10))); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestCheckAgainstFund(t *testing.T) {
	fund := ramp(11000, 0.10)
	glued := ramp(11000, 0.102) // +0.2 pts/yr: a plausible grossness
	out, err := checkAgainstFund(glued, fund)
	if err != nil {
		t.Fatalf("a factor 0.2 pts/yr above the fund was refused: %v (%s)", err, out)
	}
	if !strings.Contains(out, "monthly correlation") {
		t.Errorf("summary %q says nothing about the path", out)
	}
	if _, err := checkAgainstFund(ramp(11000, 0.09), fund); err == nil {
		t.Error("a gross factor LAGGING the fund it extends was accepted")
	}
	if _, err := checkAgainstFund(ramp(11000, 0.13), fund); err == nil {
		t.Error("a factor 3 pts/yr above the fund was accepted")
	}
	if _, err := checkAgainstFund(ramp(200, 0.102), fund); err == nil {
		t.Error("an overlap of a few months was accepted")
	}
}

func TestCheckAgainstIndexRefusesAShortOverlap(t *testing.T) {
	if _, err := checkAgainstIndex(ramp(11000, 0.10), ramp(11000, 0.10)); err == nil {
		t.Error("30 calendar years were accepted as the whole span")
	}
}
