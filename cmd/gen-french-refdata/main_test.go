package main

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/cmd/internal/refgen"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// walk is n business-ish days of a deterministic wobbling path from 100.
func walk(n int, scale float64) []marketdata.Point {
	out := make([]marketdata.Point, n)
	level := 100.0
	d := time.Date(1990, 7, 2, 0, 0, 0, 0, time.UTC)
	for i := range out {
		level *= 1 + scale*math.Sin(float64(i)*1.7)/100
		out[i] = marketdata.Point{Date: d, Close: level}
		d = d.AddDate(0, 0, 1)
	}
	return out
}

func TestCheckRevision(t *testing.T) {
	dir := t.TempDir()
	old := walk(400, 1)
	if err := refgen.Write(dir, refgen.Header{ID: "X", Name: "x", Source: "s"}, old); err != nil {
		t.Fatal(err)
	}
	fresh := walk(430, 1)
	if _, err := checkRevision(dir, "X", fresh); err != nil {
		t.Errorf("an extension of the same series was refused: %v", err)
	}
	if _, err := checkRevision(dir, "X", walk(430, 3)); err == nil {
		t.Error("a series three times as volatile passed as a revision")
	}
	if _, err := checkRevision(dir, "X", fresh[10:]); err == nil {
		t.Error("a refresh that lost bundled days was accepted")
	}
	if msg, err := checkRevision(dir, "NONE", fresh); err != nil || msg == "" {
		t.Errorf("no bundled file: %q, %v", msg, err)
	}
}

func TestCheckAgainstAnchors(t *testing.T) {
	daily := walk(12000, 1)
	var anchors []marketdata.Point
	for i, p := range daily {
		if i+1 == len(daily) || daily[i+1].Date.Month() != p.Date.Month() {
			anchors = append(anchors, p)
		}
	}
	if _, err := checkAgainstAnchors(daily, anchors); err != nil {
		t.Errorf("a series against its own month-ends was refused: %v", err)
	}
	if _, err := checkAgainstAnchors(walk(12000, 2), anchors); err == nil {
		t.Error("a path twice as volatile as its anchors was accepted")
	}
}
