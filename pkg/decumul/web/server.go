package web

import (
	"encoding/json"
	"net/http"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Handler returns the decumulation UI: the embedded page at / and the
// simulation endpoint at POST /api/sim. A non-nil panel enables the
// portfolio models (bootstrap/cohorts) and live allocation sliders; labels
// names the holdings for the allocation UI.
func Handler(panel *scenario.Panel, labels []string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(mustSub())))
	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		meta := map[string]any{"labels": labels, "hasPanel": panel != nil}
		if panel != nil {
			mu, sigma := FitParametric(*panel, panel.Weights)
			meta["weights"] = panel.Weights
			meta["mu"] = mu
			meta["sigma"] = sigma
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(meta)
	})
	mux.HandleFunc("/api/sim", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var pr Params
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ComputeWithPanel(pr, panel))
	})
	return mux
}
