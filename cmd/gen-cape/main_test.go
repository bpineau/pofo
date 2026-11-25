package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/bpineau/pofo/cmd/internal/xls"
)

func TestShillerColumns(t *testing.T) {
	cells := []xls.Cell{
		{Row: 1, Col: 5, Text: "CAPE", IsText: true}, // part of Excess CAPE Yield
		{Row: 2, Col: 0, Text: "Date", IsText: true},
		{Row: 2, Col: 3, Text: "TR CAPE", IsText: true},
		{Row: 2, Col: 4, Text: "CAPE", IsText: true},
		{Row: 3, Col: 0, Num: 2025.10},
		{Row: 3, Col: 3, Num: 40},
		{Row: 3, Col: 4, Num: 35},
		{Row: 1, Col: 4, Text: "P/E10 or", IsText: true},
		{Row: 1, Col: 2, Text: "Date   ", IsText: true},
	}
	got, err := shillerPoints(cells)
	if err != nil || !slices.Equal(got, []point{{"2025-10-01", 35}}) {
		t.Fatalf("CAPE = %v, %v", got, err)
	}
	cells[3].Text = "Total Return CAPE"
	if _, err := shillerPoints(cells); err == nil {
		t.Fatal("accepted a sheet without ordinary CAPE")
	}
}

func TestWorkbookLink(t *testing.T) {
	raw := []byte(`<a href="/other.xls"></a><a href="//cdn.example/new/ie_data.xls?version=8&amp;x=y">`)
	got, err := workbookLink("https://example.org/", raw)
	if err != nil || got != "https://cdn.example/new/ie_data.xls?version=8&x=y" {
		t.Fatalf("resolved link = %q, %v", got, err)
	}
}

func TestSupplement(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var series []point
	for i := range 16 {
		series = append(series, point{start.AddDate(0, i, 0).Format(time.DateOnly), 25})
	}
	get := func(rows []point) func() ([]point, error) {
		return func() ([]point, error) { return rows, nil }
	}
	fail := func() ([]point, error) { return nil, errors.New("unavailable") }
	mirror := series[:12:12]
	for _, tc := range []struct {
		name     string
		multpl   func() ([]point, error)
		original func() ([]point, error)
		wantErr  bool
	}{
		{"multpl unavailable", fail, get(series), false},
		{"multpl stale", get(series[:15]), get(series), false},
		{"multpl current", get(series), func() ([]point, error) { t.Fatal("unnecessary workbook fetch"); return nil, nil }, false},
		{"both unavailable", fail, fail, true},
		{"incomplete tail", fail, get(append(slices.Clone(series[:13]), series[14:]...)), true},
		{"missing overlap", fail, get(series[12:]), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := supplement(mirror, series[15].date, tc.multpl, tc.original)
			if (err != nil) != tc.wantErr || !tc.wantErr && !slices.Equal(got, series) {
				t.Fatalf("got %v, %v", got, err)
			}
		})
	}
	wrong := slices.Clone(series)
	for i := range wrong {
		wrong[i].cape *= 1.08
	}
	if _, _, err := supplement(mirror, series[15].date, fail, get(wrong)); err == nil {
		t.Fatal("accepted supplement 8% above mirror")
	}
	primary := slices.Clone(series[:15])
	primary[14].cape = 26
	got, sources, err := supplement(mirror, series[15].date, get(primary), get(series))
	if err != nil || got[14] != primary[14] || got[15] != series[15] || len(sources) != 2 {
		t.Fatalf("stale multpl and newer workbook: %v, %v, %v", got, sources, err)
	}
}

func TestWritePreservesRecentCoverage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cape.csv")
	const previous = "# provisional observation\ndate,cape\n2026-11-01,39.37\n"
	if err := os.WriteFile(path, []byte(previous), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := write(path, "mirror", []point{{date: "2023-09-01", cape: 29}}, nil); err == nil {
		t.Fatal("accepted a truncated download")
	}
	if err := write(path, "mirror", []point{{date: "2026-10-01", cape: 40}, {date: "2026-12-01", cape: 39}}, nil); err == nil {
		t.Fatal("accepted a download missing a previously shipped intermediate month")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != previous {
		t.Fatalf("previous observations changed: %q, %v", data, err)
	}
	if err := write(path, "mirror", []point{{date: "2026-11-01", cape: 39.38}}, nil); err != nil {
		t.Fatalf("revision of the latest month: %v", err)
	}
}
