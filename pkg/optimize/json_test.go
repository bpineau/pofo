package optimize

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestSpecJSONOpenBounds(t *testing.T) {
	s, err := ParseSpec("max-return,bounds:A:10-")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Resolve([][]string{{"A"}, {"B"}}); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"Upper":[null,null]`) || !strings.Contains(string(b), `"A":[0.1,null]`) {
		t.Fatalf("open bounds lost: %s", b)
	}
}

func TestResultJSONUndefined(t *testing.T) {
	b, err := json.Marshal(Result{Weights: []float64{1}, Return: math.NaN(), Sharpe: math.Inf(1), Implied: []float64{math.NaN()}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"Return":null`) || !strings.Contains(string(b), `"Sharpe":null`) || !strings.Contains(string(b), `"Implied":[null]`) {
		t.Fatalf("undefined values lost: %s", b)
	}
}
