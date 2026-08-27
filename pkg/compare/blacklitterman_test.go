package compare

import (
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/optimize"
)

// blResult is a Black-Litterman answer over three named lines, with the
// implied and posterior returns a report has to explain.
func blResult() ([]string, optimize.Result) {
	names := []string{"IGLN", "DBMFE", "DTLA"}
	res := optimize.Result{
		Weights:   []float64{0.10, 0.55, 0.35},
		Lambda:    7.9,
		Implied:   []float64{0.056, 0.031, 0.024},
		Posterior: []float64{0.020, 0.044, 0.0240001},
	}
	return names, res
}

// The note has to carry the risk aversion and where it came from, what each
// view did to the expected returns, and the views themselves. Weights without
// that are a tilt nobody can attribute to a belief.
func TestBlackLittermanNote(t *testing.T) {
	sp, err := optimize.ParseSpec("black-litterman,view:IGLN:2,view:DBMFE>DTLA:3@70,prior-return:4.6")
	if err != nil {
		t.Fatal(err)
	}
	names, res := blResult()
	note := blackLittermanNote(sp, res, names, []float64{0.14, 0.09, 0.17})

	for _, want := range []string{
		"λ = 7.9",
		"prior-return 4.6 %/yr",
		"IGLN implied 5.6 % → posterior 2.0 %",
		"DBMFE implied 3.1 % → posterior 4.4 %",
		"unchanged: DTLA",
		"IGLN earns 2.0 %/yr, confidence 50 %",
		"DBMFE beats DTLA by 3.0 points/yr, confidence 70 %",
		"The Sharpe above is the EXPECTED one",
		"EXPECTED RETURNS only",
	} {
		if !strings.Contains(note, want) {
			t.Errorf("the note misses %q:\n%s", want, note)
		}
	}
	if strings.Contains(note, "CAUTION") {
		t.Errorf("no viewed line is cash-like here:\n%s", note)
	}
}

// The default risk aversion has to say it is an assumption, not a measurement.
func TestBlackLittermanNoteDefaultLambda(t *testing.T) {
	sp, err := optimize.ParseSpec("black-litterman,view:IGLN:2")
	if err != nil {
		t.Fatal(err)
	}
	names, res := blResult()
	note := blackLittermanNote(sp, res, names, []float64{0.14, 0.09, 0.17})
	if !strings.Contains(note, "Sharpe of 0.4") || strings.Contains(note, "prior-return") {
		t.Errorf("the default lambda must name its assumption:\n%s", note)
	}
}

// With no view the objective returns the file's own weights, and the note has
// to say that is what happened, with the returns those weights imply.
func TestBlackLittermanNoteWithoutViews(t *testing.T) {
	sp, err := optimize.ParseSpec("black-litterman")
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"IGLN", "DBMFE", "DTLA"}
	res := optimize.Result{
		Weights:   []float64{0.10, 0.55, 0.35},
		Lambda:    7.9,
		Implied:   []float64{0.056, 0.031, 0.024},
		Posterior: []float64{0.056, 0.031, 0.024},
	}
	note := blackLittermanNote(sp, res, names, []float64{0.14, 0.09, 0.17})
	for _, want := range []string{"No view moved", "IGLN 5.6 %", "DBMFE 3.1 %", "DTLA 2.4 %"} {
		if !strings.Contains(note, want) {
			t.Errorf("the note misses %q:\n%s", want, note)
		}
	}
	if strings.Contains(note, "Views:") {
		t.Errorf("there are no views to list:\n%s", note)
	}
}

// A view on a cash-like line is the zero-risk-free-rate trap: the note must
// name it, since the utility will size that line by a Sharpe nobody meant.
func TestBlackLittermanNoteFlagsCashLikeViews(t *testing.T) {
	sp, err := optimize.ParseSpec("black-litterman,view:DTLA:3")
	if err != nil {
		t.Fatal(err)
	}
	names, res := blResult()
	note := blackLittermanNote(sp, res, names, []float64{0.14, 0.09, 0.021})
	if !strings.Contains(note, "CAUTION") || !strings.Contains(note, "DTLA at 2.1 % volatility") {
		t.Errorf("a cash-like viewed line must be flagged:\n%s", note)
	}
}
