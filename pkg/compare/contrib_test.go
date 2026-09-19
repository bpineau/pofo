package compare

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// The macro panel opens in 1960-01 while the bundled backcasts reach 1871
// (S&P 500) and 1953 (long Treasuries), so a /view of a deep portfolio asks
// the regime strip about months nothing measured. Those months used to be
// head-filled with the strip's zero value, "growth": eighty-nine years of a
// public chart painted with a macro state that does not exist, and the same
// months averaged into the growth column of the per-regime matrix. They must
// come back unclassified, appear in no band and be counted in no column.
func TestMonthQuadrantsLeavesPrePanelMonthsUnclassified(t *testing.T) {
	var months []time.Time
	for y := 1940; y <= 1970; y++ {
		for m := 1; m <= 12; m++ {
			months = append(months, time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC))
		}
	}
	quads := monthQuadrants(months)
	if quads == nil {
		t.Fatal("no quadrants: the bundled macro panel did not load")
	}
	before, after := 0, 0
	for i, m := range months {
		switch {
		case m.Year() < 1960:
			if quads[i] != unclassified {
				t.Fatalf("%s classified %q, want no regime before the panel starts", m.Format("2006-01"), quads[i])
			}
			before++
		default:
			if quads[i] != unclassified {
				after++
			}
		}
	}
	if before == 0 || after == 0 {
		t.Fatalf("fixture degenerate: %d pre-panel, %d classified months", before, after)
	}

	// The matrix counts only the classified months and says how many it left out.
	labels := []string{"A"}
	mc := [][]float64{make([]float64, len(months))}
	for i := range mc[0] {
		mc[0][i] = 0.001
	}
	svg := contribMatrix(months, mc, quads, labels, []string{"#0880A8"})
	if svg == "" {
		t.Fatal("no matrix rendered")
	}
	if !strings.Contains(svg, fmt.Sprintf("%d months precede the macro panel", before)) {
		t.Errorf("the matrix does not say how many months it left out (%d expected)", before)
	}
	// Growth must no longer claim the pre-panel months: with 240 of them, a
	// head-filled growth column would read at least 240 months.
	if strings.Contains(svg, fmt.Sprintf("%d months", len(months))) {
		t.Error("a column claims the whole window, head-fill is back")
	}
}
