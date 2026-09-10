package marketdata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWriteReadSimdataRoundTrip(t *testing.T) {
	dir := t.TempDir()
	sf := &SimdataFile{
		ID:         "IE000KF370H3",
		Name:       "WisdomTree US Efficient Core",
		Method:     "0.90 VFINX + 0.60 VFITX",
		Validation: "corr=0.98 vs NTSX",
		Generated:  "2026-06-12",
		Points: []Point{
			{Date: d(2000, 1, 3), Close: 100},
			{Date: d(2000, 1, 4), Close: 101.5},
		},
	}
	if err := WriteSimdata(dir, sf); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "IE000KF370H3.csv"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# pofo simdata v1", "# id: IE000KF370H3", "# name: WisdomTree US Efficient Core",
		"# method: 0.90 VFINX + 0.60 VFITX", "# validation: corr=0.98 vs NTSX",
		"# generated: 2026-06-12", "date,close", "2000-01-03,100.000000",
	} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("the file does not carry %q:\n%s", want, raw)
		}
	}

	s, ok, err := ReadSimdata(dir, "IE000KF370H3")
	if err != nil || !ok {
		t.Fatalf("read back: ok=%v err=%v", ok, err)
	}
	if s.Name != sf.Name || s.Source != "simdata" || len(s.Points) != 2 {
		t.Fatalf("round trip lost metadata: %+v", s)
	}
	if s.Last().Close != 101.5 || !s.First().Date.Equal(d(2000, 1, 3)) {
		t.Errorf("round trip lost data: %+v", s.Points)
	}
	// A directory with no such file is a miss, not an error.
	if _, ok, err := ReadSimdata(dir, "NOSUCH"); ok || err != nil {
		t.Errorf("absent file: ok=%v err=%v", ok, err)
	}
}

func TestWriteSimdataRejectsIncomplete(t *testing.T) {
	dir := t.TempDir()
	cases := []*SimdataFile{
		{ID: "", Points: []Point{{Date: d(2000, 1, 3), Close: 1}}},
		{ID: "X", Points: nil},
	}
	for _, sf := range cases {
		if err := WriteSimdata(dir, sf); err == nil {
			t.Errorf("WriteSimdata(%+v) should have failed", sf)
		}
	}
}

// TestWriteSimdataNamesTheCanonicalFile pins the trap documented in CLAUDE.md:
// the file is named after the canonical id, never after the alias asked for.
func TestWriteSimdataNamesTheCanonicalFile(t *testing.T) {
	dir := t.TempDir()
	alias := "SP500"
	canonical := CanonicalID(alias)
	sf := &SimdataFile{ID: alias, Points: []Point{{Date: d(1990, 1, 2), Close: 10}}}
	if err := WriteSimdata(dir, sf); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, sanitizeFilename(canonical)+".csv")); err != nil {
		t.Fatalf("the file should be named after %q: %v", canonical, err)
	}
	s, ok, err := ReadSimdata(dir, alias)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	// No "# name:" header: the symbol carries the fallback label.
	if !strings.HasSuffix(s.Name, "(simdata)") {
		t.Errorf("default name = %q, want the %q (simdata) fallback", s.Name, s.Symbol)
	}
}

func TestReadSimdataFSMalformed(t *testing.T) {
	const head = "# pofo simdata v1\n# id: X\ndate,close\n"
	cases := []struct {
		name string
		body string
		want string // "" = no error
	}{
		{"sorted output from unsorted rows", head + "2000-01-04,2\n2000-01-03,1\n", ""},
		{"comment without a colon is skipped", "# pofo simdata v1\n" + head + "2000-01-03,1\n", ""},
		{"blank lines are skipped", head + "\n2000-01-03,1\n\n", ""},
		{"no data at all", head, "no data"},
		{"line without a comma", head + "2000-01-03\n", "invalid line"},
		{"unparseable date", head + "03/01/2000,1\n", "invalid date"},
		{"unparseable close", head + "2000-01-03,abc\n", "invalid close"},
		{"non-positive close", head + "2000-01-03,0\n", "invalid close"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{"X.csv": &fstest.MapFile{Data: []byte(tc.body)}}
			s, ok, err := ReadSimdataFS(fsys, "X")
			switch {
			case tc.want == "" && (err != nil || !ok):
				t.Fatalf("ok=%v err=%v", ok, err)
			case tc.want == "":
				if !sortedByDate(s.Points) {
					t.Errorf("points not sorted: %+v", s.Points)
				}
			case err == nil || !strings.Contains(err.Error(), tc.want):
				t.Fatalf("error = %v, want one about %q", err, tc.want)
			case ok:
				t.Error("a malformed file must not report ok")
			}
		})
	}
	// An absent file is a miss, an unreadable one an error.
	if _, ok, err := ReadSimdataFS(fstest.MapFS{}, "X"); ok || err != nil {
		t.Errorf("absent: ok=%v err=%v", ok, err)
	}
}

func sortedByDate(pts []Point) bool {
	for i := 1; i < len(pts); i++ {
		if !pts[i].Date.After(pts[i-1].Date) {
			return false
		}
	}
	return true
}
