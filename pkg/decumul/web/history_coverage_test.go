package web

import (
	"strings"
	"testing"
)

func TestHistoryCoverageDoesNotCertifyForecast(t *testing.T) {
	for _, tc := range []struct {
		months      int
		label, note string
	}{
		{420, "FULL HORIZON", "full-horizon start windows: 1 "},
		{720, "FULL HORIZON", "full-horizon start windows: 301 "},
		{120, "SHORT", "no complete 35-year window"},
		{12, "PARAMETRIC", "insufficient"},
	} {
		panel := monthlyPanel(tc.months)
		label, note := historyCoverage(Params{Years: 35}, &panel)
		if label != tc.label || !strings.Contains(note, tc.note) {
			t.Errorf("%d months: %s / %s", tc.months, label, note)
		}
		if tc.months >= 420 && !strings.Contains(note, "not a confidence level") {
			t.Errorf("forecast qualification missing: %s", note)
		}
	}
}
