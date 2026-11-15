package simgen

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestConstantReferenceAuditJSON(t *testing.T) {
	v, err := Validate(mkWobbly("A", 400, .001, .01), mkSeries("B", 400, 0))
	if err != nil {
		t.Fatal(err)
	}
	a := AuditResult{DailyCorr: v.Corr, WeeklyCorr: v.WeeklyCorr, Beta: v.Beta, MonthlyCorr: v.Corr,
		Chain: []Junction{{Measured: true, Corr: v.Corr}}}
	for _, value := range []any{v, []AuditGroup{{Title: "constant", Results: []AuditResult{a}}}} {
		b, err := json.Marshal(value)
		if err != nil || !strings.Contains(string(b), `"Corr":null`) {
			t.Fatalf("undefined audit JSON = %s, %v", b, err)
		}
	}
}

func TestUndefinedCorrelationCannotGradePath(t *testing.T) {
	_, path, _ := grade(AuditResult{MonthlyCorr: math.NaN(), Months: 24, VolReal: .1})
	if path != VerdictUnknown {
		t.Fatalf("unmeasured path graded %s", path)
	}
}
