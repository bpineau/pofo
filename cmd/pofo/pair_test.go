package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
)

// -pair reads a bundled backcast against its bundled reference offline, and
// prints the study as text.
func TestPairText(t *testing.T) {
	out, err := runArgs(t, "-offline", "-data", t.TempDir(),
		"-pair", "SP500,SP500-USD", "-start", "2000-01-01", "-end", "2009-12-31")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, want := range []string{
		"A  SP500 ", "B  SP500-USD ", "window      2000-01-31 to 2009-12-31",
		"\nmonthly  119 ", "\n2008  -37.",
		"no daily figures: A quotes daily and B monthly",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("no %q in:\n%s", want, out)
		}
	}
}

// A path is read as a CSV file, labelled as written: the old-versus-new
// check, here with a revised last month, as JSON.
func TestPairFilesJSON(t *testing.T) {
	dir := t.TempDir()
	d := func(m time.Month) time.Time { return time.Date(2024, m+1, 0, 0, 0, 0, 0, time.UTC) }
	dates := []time.Time{d(1), d(2), d(3), d(4), d(5), d(6)}
	write := func(name string, closes ...float64) string {
		var b strings.Builder
		b.WriteString("date,close\n")
		for i, c := range closes {
			b.WriteString(dates[i].Format(time.DateOnly) + "," + strconv.FormatFloat(c, 'g', -1, 64) + "\n")
		}
		path := filepath.Join(dir, name+".csv")
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	oldPath := write("old", 100, 101, 99, 103, 104, 106)
	newPath := write("new", 100, 101, 99, 103, 104, 107)

	out, err := runArgs(t, "-offline", "-data", t.TempDir(), "-pair", newPath+","+oldPath, "-json")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	var st analyze.PairStudy
	if err := json.Unmarshal([]byte(out), &st); err != nil {
		t.Fatalf("not a PairStudy: %v\n%s", err, out)
	}
	if st.A.ID != newPath || st.B.ID != oldPath || st.Shared != 6 || !st.FirstDivergence.Equal(d(6)) {
		t.Errorf("study: A %s, B %s, %d shared, first divergence %v", st.A.ID, st.B.ID, st.Shared, st.FirstDivergence)
	}
}

func TestPairRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		argv []string
		want string
	}{
		{"one side", []string{"-pair", "SP500"}, "-pair takes two series"},
		{"three sides", []string{"-pair", "SP500,SP500-USD,TBILL-3M"}, "-pair takes two series"},
		{"a missing file", []string{"-pair", "SP500,./nope.csv"}, "nope.csv"},
		{"no overlap", []string{"-pair", "SP500,SP500-USD", "-start", "2026-09-12"}, "SP500-USD: no point between -start and -end"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runArgs(t, append([]string{"-offline", "-data", t.TempDir()}, tc.argv...)...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want one about %q\n%s", err, tc.want, out)
			}
		})
	}
}
