package suggest

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestConstantHoldingsAreNotCorrelatedEvidence(t *testing.T) {
	h := []Holding{{ID: "A", Weight: .5}, {ID: "B", Weight: .5}}
	if g := Redundancies(h, [][]float64{{0, 0, 0}, {0, 0, 0}}, .9); len(g) != 0 {
		t.Fatalf("constant histories grouped: %+v", g)
	}
	b, err := json.Marshal(Suggestion{Corr: math.NaN()})
	if err != nil || !strings.Contains(string(b), `"Corr":null`) {
		t.Fatalf("undefined correlation JSON = %s, %v", b, err)
	}
}
