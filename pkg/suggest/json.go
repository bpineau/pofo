package suggest

import (
	"encoding/json"
	"math"
)

// MarshalJSON preserves an undefined correlation as null.
func (s Suggestion) MarshalJSON() ([]byte, error) {
	type plain Suggestion
	var corr *float64
	if !math.IsNaN(s.Corr) && !math.IsInf(s.Corr, 0) {
		corr = &s.Corr
	}
	return json.Marshal(struct {
		plain
		Corr *float64
	}{plain(s), corr})
}
