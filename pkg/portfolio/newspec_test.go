package portfolio

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

// NewSpec and Parse share one validation: the same lines, written in a file
// or passed in code, give the same Spec.
func TestNewSpecMatchesParse(t *testing.T) {
	cases := []struct {
		name  string
		file  string
		lines []Line
	}{
		{"sixty-forty", "60 IWDA 0.2\n40 AGGH\n",
			[]Line{{ID: "IWDA", Weight: 0.6, Fees: 0.2}, {ID: "AGGH", Weight: 0.4, Fees: -1}}},
		{"three-lines", "50 VTI 0.03\n30 IE00B4L5Y983\n20 IGLN\n", // a Line's zero fee is the file's absent column
			[]Line{{ID: "VTI", Weight: 0.5, Fees: 0.03}, {ID: "IE00B4L5Y983", Weight: 0.3, Fees: -1}, {ID: "IGLN", Weight: 0.2}}},
		{"normalized", "30 VTI\n30 BND\n",
			[]Line{{ID: "VTI", Weight: 0.3, Fees: -1}, {ID: "BND", Weight: 0.3, Fees: -1}}},
		{"duplicate", "50 VTI\n50 VTI\n",
			[]Line{{ID: "VTI", Weight: 0.5, Fees: -1}, {ID: "VTI", Weight: 0.5, Fees: -1}}},
		{"sim-suffix", "100 IWDASIM\n", []Line{{ID: " IWDASIM ", Weight: 1, Fees: -1}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			parsed, err := Parse("file", strings.NewReader(c.file))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			built, err := NewSpec("code", c.lines...)
			if err != nil {
				t.Fatalf("NewSpec: %v", err)
			}
			built.Name = parsed.Name
			if !reflect.DeepEqual(built, parsed) {
				t.Errorf("NewSpec = %+v\nParse   = %+v", built, parsed)
			}
		})
	}
}

func TestNewSpecDefaults(t *testing.T) {
	spec, err := NewSpec("p", Line{ID: "VTI", Weight: 1, Fees: -3})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Name != "p" || spec.RebalanceDays != -1 || spec.Capital != -1 ||
		spec.BorrowSpread != -1 || spec.EnvelopeFees != -1 || spec.Meta != nil ||
		spec.Leverage || spec.Sim || spec.Optimize != nil || spec.Warnings != nil {
		t.Errorf("directives not left unset: %+v", spec)
	}
	h := spec.Holdings[0]
	if h.Weight != 1 || h.RawWeight != 100 || h.Fees != -1 {
		t.Errorf("holding = %+v, want weight 1, raw 100, fees -1 (any negative means unknown)", h)
	}
	// The zero value is unknown too: a Line that says nothing about fees must
	// not declare a 0 % TER behind the caller's back.
	spec, err = NewSpec("p", Line{ID: "VTI", Weight: 1})
	if err != nil {
		t.Fatal(err)
	}
	if h := spec.Holdings[0]; h.Fees != -1 {
		t.Errorf("fees left at zero read %g, want -1 (unknown)", h.Fees)
	}
}

func TestNewSpecNormalizes(t *testing.T) {
	spec, err := NewSpec("p", Line{ID: "VTI", Weight: 0.3, Fees: -1}, Line{ID: "BND", Weight: 0.1, Fees: -1})
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.Holdings[0].Weight + spec.Holdings[1].Weight; math.Abs(got-1) > 1e-12 {
		t.Errorf("weights sum to %v, want 1", got)
	}
	if math.Abs(spec.Holdings[0].Weight-0.75) > 1e-12 {
		t.Errorf("VTI weight %v, want 0.75", spec.Holdings[0].Weight)
	}
	if len(spec.Warnings) != 1 || !strings.Contains(spec.Warnings[0], "normalized") {
		t.Errorf("warnings = %q, want the normalization note", spec.Warnings)
	}
}

func TestNewSpecErrors(t *testing.T) {
	cases := map[string][]Line{
		"no line":         nil,
		"empty id":        {{ID: " ", Weight: 1}},
		"blank inside id": {{ID: "IW DA", Weight: 1}},
		"negative weight": {{ID: "VTI", Weight: -0.5}, {ID: "BND", Weight: 1}},
		"zero weight":     {{ID: "VTI", Weight: 0}},
		"all zero":        {{ID: "VTI"}, {ID: "BND"}},
		"above one":       {{ID: "VTI", Weight: 1.5}},
		"NaN weight":      {{ID: "VTI", Weight: math.NaN()}},
		"fees too high":   {{ID: "VTI", Weight: 1, Fees: 25}},
		"NaN fees":        {{ID: "VTI", Weight: 1, Fees: math.NaN()}},
	}
	for name, lines := range cases {
		if _, err := NewSpec("p", lines...); err == nil {
			t.Errorf("%s: NewSpec accepted %+v", name, lines)
		}
	}
}

// A NaN passes every "out of range" comparison written the naive way; the
// shared bounds refuse it in a file too.
func TestParseRefusesNaN(t *testing.T) {
	for _, file := range []string{"NaN VTI\n", "100 VTI NaN\n"} {
		if _, err := Parse("p", strings.NewReader(file)); err == nil {
			t.Errorf("Parse accepted %q", file)
		}
	}
}
