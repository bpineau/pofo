package main

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// -dump prints nothing but long CSV: a bundled reference series and a catalog
// index, offline on an empty cache, cut to the window and resampled to
// month-ends, with the junction moved onto its month's kept close.
func TestDumpWritesLongCSV(t *testing.T) {
	out, err := runArgs(t, "-offline", "-data", t.TempDir(),
		"-dump", "TREASURY-LONG-YIELD, SP500", "-monthly", "-start", "1972-07-01", "-end", "1973-06-30")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	list, err := marketdata.ReadLongCSV(strings.NewReader(out))
	if err != nil {
		t.Fatalf("the output is not long CSV: %v\n%s", err, out)
	}
	if len(list) != 2 || list[0].Symbol != "TREASURY-LONG-YIELD" || list[1].Symbol != "SP500" {
		t.Fatalf("got %d series, want TREASURY-LONG-YIELD then SP500:\n%s", len(list), out)
	}
	for _, s := range list {
		if s.Len() != 12 || s.First().Date.Month() != time.July || s.Last().Date.After(time.Date(1973, 6, 30, 0, 0, 0, 0, time.UTC)) {
			t.Errorf("%s: %d points from %s to %s, want the twelve month-ends of the window",
				s.Symbol, s.Len(), s.First().Date.Format(time.DateOnly), s.Last().Date.Format(time.DateOnly))
		}
	}
	if want := []time.Time{time.Date(1973, 1, 31, 0, 0, 0, 0, time.UTC)}; !reflect.DeepEqual(list[0].Junctions, want) {
		t.Errorf("junctions %v, want the 1973-01-04 definition break on its month-end", list[0].Junctions)
	}
}

func TestDumpRefusals(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		want string
	}{
		{"nothing local", []string{"-dump", "ZZQQXXNOPE"}, "offline, and nothing cached or bundled answers"},
		{"a reference series has no currency", []string{"-dump", "TBILL-3M", "-currency", "EUR"}, "states no currency"},
		{"an empty window", []string{"-dump", "TBILL-3M", "-start", "1900-01-01", "-end", "1901-01-01"}, "no point between -start and -end"},
		{"offline warmup", []string{"-warmup"}, "-offline cannot be combined"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			argv := append([]string{"-offline", "-data", t.TempDir()}, tc.argv...)
			out, err := runArgs(t, argv...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want one about %q\n%s", err, tc.want, out)
			}
		})
	}
}

func TestDumpList(t *testing.T) {
	out, err := runArgs(t, "-dump", "list")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if !strings.HasPrefix(lines[0], "ID ") || len(lines) != len(marketdata.BundledIDs())+1 {
		t.Fatalf("%d lines, want a header and one per bundled series:\n%s", len(lines), out)
	}
	if !strings.Contains(out, "TREASURY-LONG-USD ") || !strings.Contains(out, " refdata ") || !strings.Contains(out, " simdata ") {
		t.Errorf("the listing lacks a reference series or a family:\n%s", out)
	}
}

func TestSplitIDs(t *testing.T) {
	if got := splitIDs(" IWDA, ,DBMFSIM,"); !reflect.DeepEqual(got, []string{"IWDA", "DBMFSIM"}) {
		t.Errorf("splitIDs = %q", got)
	}
	if got := splitIDs(""); got != nil {
		t.Errorf("splitIDs(\"\") = %q, want nil", got)
	}
}
