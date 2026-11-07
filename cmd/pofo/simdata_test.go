package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/simgen"
)

func TestGenOnePreservesReconstructionMetadata(t *testing.T) {
	day := func(n int) time.Time { return time.Date(2020, 1, n, 0, 0, 0, 0, time.UTC) }
	sim := &marketdata.Series{
		SimulatedBefore: day(3),
		Junctions:       []time.Time{day(2)},
		Warnings:        []string{"reference coverage is limited"},
		Ends:            day(4),
		Points: []marketdata.Point{
			{Date: day(1), Close: 50}, {Date: day(2), Close: 51},
			{Date: day(3), Close: 52}, {Date: day(4), Close: 53},
		},
	}
	dir := t.TempDir()
	r := simgen.Recipe{ID: "TEST", Build: func(simgen.Fetcher, time.Time) (*marketdata.Series, error) { return sim, nil }}
	if err := genOne(context.Background(), marketdata.NewClient(""), nil, dir, r, false); err != nil {
		t.Fatal(err)
	}
	got, ok, err := marketdata.ReadSimdata(dir, "TEST")
	if err != nil || !ok {
		t.Fatalf("read generated file: ok=%v err=%v", ok, err)
	}
	if !got.SimulatedBefore.Equal(day(3)) || !slices.Equal(got.Junctions, []time.Time{day(2)}) ||
		!slices.Contains(got.Warnings, "reference coverage is limited") || !got.Ends.Equal(day(4)) {
		t.Errorf("generation lost provenance: frontier=%v junctions=%v warnings=%v ends=%v", got.SimulatedBefore, got.Junctions, got.Warnings, got.Ends)
	}
}

func TestGenOneDoesNotGuessReconstructionState(t *testing.T) {
	for _, fully := range []bool{false, true} {
		t.Run(fmt.Sprint(fully), func(t *testing.T) {
			dir := t.TempDir()
			r := simgen.Recipe{ID: "TEST", Build: func(simgen.Fetcher, time.Time) (*marketdata.Series, error) {
				return &marketdata.Series{FullySimulated: fully, Source: "simdata", Points: []marketdata.Point{
					{Date: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), Close: 100},
					{Date: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Close: 101},
				}}, nil
			}}
			if err := genOne(context.Background(), marketdata.NewClient(""), nil, dir, r, false); err != nil {
				t.Fatal(err)
			}
			s, ok, err := marketdata.ReadSimdata(dir, "TEST")
			if err != nil || !ok {
				t.Fatalf("read generated file: ok=%v err=%v", ok, err)
			}
			if s.FullySimulated != fully || !s.SimulatedBefore.IsZero() {
				t.Errorf("generation changed explicit provenance: fully=%v frontier=%v", s.FullySimulated, s.SimulatedBefore)
			}
		})
	}
}

func TestGenOnePersistsRealSpliceBoundary(t *testing.T) {
	day := func(n int) time.Time { return time.Date(2020, 1, n, 0, 0, 0, 0, time.UTC) }
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v8/finance/chart/TST" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `{"chart":{"result":[{"meta":{"currency":"USD","symbol":"TST","longName":"Test fund"},"timestamp":[%d,%d],"indicators":{"quote":[{"close":[104,106]}],"adjclose":[{"adjclose":[104,106]}]}}],"error":null}}`, day(3).Unix(), day(4).Unix())
	}))
	defer srv.Close()
	c := marketdata.NewClient("")
	c.ChartBase, c.SearchBase, c.StooqBase = srv.URL, srv.URL, srv.URL
	c.FTBase, c.MorningstarBase, c.BoursoramaBase = srv.URL, srv.URL, srv.URL
	c.CookieBase = srv.URL
	dir := t.TempDir()
	r := simgen.Recipe{ID: "TEST", SpliceReal: "TST", Build: func(simgen.Fetcher, time.Time) (*marketdata.Series, error) {
		return &marketdata.Series{
			Symbol:    "PROXY",
			Warnings:  []string{"proxy coverage caveat"},
			Junctions: []time.Time{day(2)},
			Points: []marketdata.Point{
				{Date: day(1), Close: 50}, {Date: day(2), Close: 51},
				{Date: day(3), Close: 52}, {Date: day(4), Close: 53},
			},
		}, nil
	}}
	if err := genOne(context.Background(), c, nil, dir, r, false); err != nil {
		t.Fatal(err)
	}
	got, ok, err := marketdata.ReadSimdata(dir, "TEST")
	if err != nil || !ok {
		t.Fatalf("read generated file: ok=%v err=%v", ok, err)
	}
	if !got.SimulatedBefore.Equal(day(3)) {
		t.Errorf("generated splice frontier = %v, want 2020-01-03", got.SimulatedBefore)
	}
	if !slices.Equal(got.Junctions, []time.Time{day(2)}) || !strings.Contains(strings.Join(got.Warnings, " "), "proxy coverage caveat") {
		t.Errorf("generated splice lost metadata: junctions=%v warnings=%v", got.Junctions, got.Warnings)
	}
	if len(got.Points) != 4 || got.Points[0].Close != 100 || got.Points[2].Close != 104 || got.Points[3].Close != 106 {
		t.Errorf("generated prices changed: %v", got.Points)
	}
}
