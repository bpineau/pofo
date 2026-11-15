package decumul

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestEmptyLifeOutcomeUndefined(t *testing.T) {
	o := (Ensemble{}).LifeOutcome()
	v := reflect.ValueOf(o)
	for i := range v.NumField() {
		if !math.IsNaN(v.Field(i).Float()) {
			t.Errorf("empty %s = %v", v.Type().Field(i).Name, v.Field(i).Float())
		}
	}
	b, err := json.Marshal(o)
	if err != nil || !strings.Contains(string(b), `"RuinAlive":null`) {
		t.Fatalf("JSON = %s, %v", b, err)
	}
}

func TestLifeOutcomeWithoutSpendingYears(t *testing.T) {
	o := (Ensemble{Paths: []PathResult{{Wealth: []float64{100}}}}).LifeOutcome()
	if !math.IsNaN(o.IncomeMean) || o.EstateP50 != 100 || o.RuinAlive != 0 {
		t.Fatalf("unlived income versus defined estate: %+v", o)
	}
}
