package marketdata

import (
	"bytes"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestReadCSV(t *testing.T) {
	const file = `# pofo simdata v1
# id: TESTID
# name: A test series: with a colon
# method: ignored, as every unknown header
# currency: USD
# junctions: 2020-01-03
date,close

2020-01-03,-0.25
2020-01-02,1.5e2
   2020-01-06 , 101.25
`
	s, err := ReadCSV(strings.NewReader(file), "")
	if err != nil {
		t.Fatal(err)
	}
	want := &Series{
		Symbol:    "TESTID",
		Name:      "A test series: with a colon",
		Currency:  "USD",
		Junctions: []time.Time{d(2020, 1, 3)},
		Points: []Point{
			{Date: d(2020, 1, 2), Close: 150},
			{Date: d(2020, 1, 3), Close: -0.25}, // a rate may be negative
			{Date: d(2020, 1, 6), Close: 101.25},
		},
	}
	if !reflect.DeepEqual(s, want) {
		t.Errorf("got  %+v\nwant %+v", s, want)
	}
	// The id argument wins over the file's own stamp.
	if s, err := ReadCSV(strings.NewReader(file), "MINE"); err != nil || s.Symbol != "MINE" {
		t.Errorf("Symbol = %v (%v), want MINE", s, err)
	}
	// No header at all is fine.
	if s, err := ReadCSV(strings.NewReader("2020-01-02,1\n2020-01-03,2\n"), "X"); err != nil || s.Len() != 2 {
		t.Errorf("headerless: %v, %v", s, err)
	}
}

func TestReadCSVErrorsNameTheLine(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"empty", "# only a comment\n", "no data"},
		{"header only", "date,value\n", "no data"},
		{"a typo on the first row is not a header", "03/01/2020,1\n", `line 1: invalid date "03/01/2020"`},
		{"second header", "date,value\n2020-01-02,1\ndate,value\n", `line 3: invalid date "date"`},
		{"three fields", "2020-01-02,1,2\n", `line 1: "2020-01-02,1,2" is not a date,value row`},
		{"not a number", "2020-01-02,1\n2020-01-03,x\n", `line 2: invalid value "x"`},
		{"not finite", "2020-01-02,NaN\n", `line 1: invalid value "NaN"`},
		{"infinite", "2020-01-02,+Inf\n", `line 1: invalid value "+Inf"`},
		{"duplicate", "2020-01-02,1\n\n2020-01-02,2\n", "lines 1 and 3: 2020-01-02 given twice"},
		{"bad junction", "# junctions: 2020-01-02,soon\n2020-01-02,1\n", `line 1: invalid junction date "soon"`},
		{"bad frontier", "# simulated-before: later\n2020-01-02,1\n", `line 1: invalid simulated-before date "later"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadCSV(strings.NewReader(tc.body), "X")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want one about %q", err, tc.want)
			}
			if !strings.HasPrefix(err.Error(), "marketdata: ReadCSV X: ") {
				t.Errorf("error %q does not name the call and the id", err)
			}
		})
	}
}

// The long layout round-trips exactly: every value WriteCSV writes parses back
// to the same float64, and the metadata a CSV can carry comes back with it.
func TestWriteCSVRoundTrip(t *testing.T) {
	in := []*Series{
		{
			Symbol:          "TREASURY-LONG-YIELD",
			Name:            "US long Treasury yield: percent",
			Currency:        "USD",
			SimulatedBefore: d(1973, 1, 3),
			Junctions:       []time.Time{d(1973, 1, 4), d(1977, 2, 15)},
			Points: []Point{
				{Date: d(1973, 1, 3), Close: 6.177},
				{Date: d(1973, 1, 4), Close: 6.89},
				{Date: d(1977, 2, 15), Close: 1.0 / 3},
			},
		},
		{
			Symbol: "my fund", // spaces are fine
			Points: []Point{
				{Date: d(2020, 1, 2), Close: math.Nextafter(100, 101)},
				{Date: d(2020, 1, 3), Close: -1e-300},
				{Date: d(2020, 1, 6), Close: 0},
				{Date: d(2020, 1, 7), Close: 123456789012345678},
			},
		},
	}
	var buf bytes.Buffer
	if err := WriteCSV(&buf, in...); err != nil {
		t.Fatal(err)
	}
	out, err := ReadLongCSV(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("%v\n%s", err, buf.String())
	}
	if !reflect.DeepEqual(out, in) {
		t.Errorf("round trip changed the series:\n%s\ngot  %+v\nwant %+v", buf.String(), out, in)
	}
	for _, line := range []string{
		"# TREASURY-LONG-YIELD name: US long Treasury yield: percent",
		"# TREASURY-LONG-YIELD junctions: 1973-01-04,1977-02-15",
		"id,date,value",
		"TREASURY-LONG-YIELD,1973-01-04,6.89",
		"my fund,2020-01-06,0",
	} {
		if !strings.Contains(buf.String(), line+"\n") {
			t.Errorf("output lacks the line %q:\n%s", line, buf.String())
		}
	}
}

func TestWriteCSVRefusesWhatWouldNotReadBack(t *testing.T) {
	ok := func(sym string) *Series {
		return &Series{Symbol: sym, Points: []Point{{Date: d(2020, 1, 2), Close: 1}}}
	}
	cases := []struct {
		name string
		list []*Series
		want string
	}{
		{"nil", []*Series{nil}, "nil series"},
		{"no symbol", []*Series{ok("")}, "no symbol"},
		{"comma", []*Series{ok("A,B")}, "comma"},
		{"leading hash", []*Series{ok("#A")}, "leading #"},
		{"surrounding blank", []*Series{ok("A ")}, "surrounding blanks"},
		{"twice", []*Series{ok("A"), ok("A")}, "A given twice"},
		{"empty", []*Series{{Symbol: "E"}}, "E has no point"},
		{"not finite", []*Series{{Symbol: "N", Points: []Point{{Date: d(2020, 1, 2), Close: math.NaN()}}}}, "not finite"},
		{"unsorted", []*Series{{Symbol: "U", Points: []Point{{Date: d(2020, 1, 3), Close: 1}, {Date: d(2020, 1, 2), Close: 1}}}}, "2020-01-02 does not follow 2020-01-03"},
		{"multi-line name", []*Series{{Symbol: "M", Name: "a\nb", Points: ok("M").Points}}, "line break"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteCSV(&buf, append([]*Series{ok("FIRST")}, tc.list...)...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want one about %q", err, tc.want)
			}
			if buf.Len() != 0 {
				t.Errorf("a refused list wrote %q", buf.String())
			}
		})
	}
}

func TestReadLongCSV(t *testing.T) {
	const file = `# a free comment, ignored
# B currency: EUR
id,date,value
B,2020-01-03,2
A,2020-01-02,10
B,2020-01-02,1
`
	got, err := ReadLongCSV(strings.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}
	// Order of first appearance, a metadata line included; dates sorted.
	if len(got) != 2 || got[0].Symbol != "B" || got[1].Symbol != "A" {
		t.Fatalf("got %+v, want B then A", got)
	}
	if got[0].Currency != "EUR" || got[0].Len() != 2 || !got[0].First().Date.Equal(d(2020, 1, 2)) {
		t.Errorf("B = %+v", got[0])
	}

	cases := []struct{ name, body, want string }{
		{"duplicate", "A,2020-01-02,1\nB,2020-01-02,1\nA,2020-01-02,2\n", "ReadLongCSV A: lines 1 and 3: 2020-01-02 given twice"},
		{"metadata without rows", "# Z name: ghost\nA,2020-01-02,1\n", "ReadLongCSV Z: no data"},
		{"two fields", "id,date,value\n2020-01-02,1\n", "line 2: 2 fields, want id,date,value"},
		{"empty id", ",2020-01-02,1\n", "line 1: empty id"},
		{"bad value", "A,2020-01-02,one\n", `line 1: invalid value "one"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadLongCSV(strings.NewReader(tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want one about %q", err, tc.want)
			}
		})
	}
}

func TestLongMetadata(t *testing.T) {
	cases := []struct{ text, id, key, val string }{
		{"A name: x", "A", "name", "x"},
		{"my fund currency: EUR", "my fund", "currency", "EUR"},
		{"EPA:CW8 name: Amundi: MSCI World", "EPA:CW8", "name", "Amundi: MSCI World"},
		{"A name: a currency: b", "A", "name", "a currency: b"},
		{"A simulated-before: 2000-01-03", "A", "simulated-before", "2000-01-03"},
	}
	for _, tc := range cases {
		id, key, val, ok := longMetadata(tc.text)
		if !ok || id != tc.id || key != tc.key || val != tc.val {
			t.Errorf("longMetadata(%q) = %q, %q, %q, %v", tc.text, id, key, val, ok)
		}
	}
	for _, text := range []string{"pofo series", "name: x", "A method: y"} {
		if _, _, _, ok := longMetadata(text); ok {
			t.Errorf("longMetadata(%q) took a comment for metadata", text)
		}
	}
}
