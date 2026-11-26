package firebook

import (
	"slices"
	"testing"
)

func TestBootstrapPlateKeepsMonthlyReturnsTogether(t *testing.T) {
	// Each column contains the source month and its stock and bond returns.
	// Both illustrations should copy whole columns from the history at the top.
	for _, render := range []func(string) string{FigureSVG, FigureSVGEnglish} {
		var numbers []string
		for _, match := range reFigText.FindAllStringSubmatch(render("bootstrap-blocs"), -1) {
			if isNeutralPayload(match[2]) {
				numbers = append(numbers, match[2])
			}
		}
		if len(numbers) != (8+6+6)*3 {
			t.Fatalf("got %d numeric labels; want three per month", len(numbers))
		}
		replayed := []int{2, 3, 4, 5, 6, 7}
		resampled := []int{5, 6, 2, 3, 5, 6}
		for i, month := range append(replayed, resampled...) {
			src, dst := 3*(month-1), 3*(8+i)
			if !slices.Equal(numbers[src:src+3], numbers[dst:dst+3]) {
				t.Errorf("month %d copied as %v; source is %v", month, numbers[dst:dst+3], numbers[src:src+3])
			}
		}
	}
}
