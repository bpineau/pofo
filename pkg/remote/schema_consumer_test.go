package remote

import "testing"

// A tool-definition consumer may render oneOf branches as standalone union
// members. Partial branches hide sibling properties and their input types.
// Keep conditional validation separate from the discoverable object shape.
func TestFIREToolInputsKeepDiscoverableShape(t *testing.T) {
	for _, tool := range Tools() {
		if tool.Name != "simulate_fire" && tool.Name != "solve_fire" {
			continue
		}
		if _, ok := tool.InputSchema["oneOf"]; ok {
			t.Errorf("%s root union hides the shared input fields", tool.Name)
		}
		props := tool.InputSchema["properties"].(map[string]any)
		market := props["market"].(map[string]any)
		if _, ok := market["oneOf"]; ok {
			t.Errorf("%s market union hides the numerical input fields", tool.Name)
		}
		for _, key := range []string{"real_mean_return_pct", "annual_volatility_pct", "degrees_of_freedom", "holdings"} {
			p, ok := market["properties"].(map[string]any)[key].(map[string]any)
			if !ok || p["type"] == nil {
				t.Errorf("missing typed market.%s", key)
			}
		}
	}
}
