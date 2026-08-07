package web

import (
	"encoding/json"
	"net/http"

	"github.com/bpineau/pofo/pkg/scenario"
	"github.com/bpineau/pofo/pkg/webui"
)

// Handler returns the decumulation UI: the embedded page at / and the
// simulation endpoint at POST /api/sim. A non-nil panel enables the
// portfolio models (bootstrap/cohorts) and live allocation sliders; labels
// names the holdings for the allocation UI.
func Handler(panel *scenario.Panel, labels []string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(mustSub())))
	// The shared visual identity (webui.CSS) is served here so both HTML
	// surfaces link the same stylesheet; the report inlines the same bytes.
	mux.HandleFunc("/theme.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(webui.CSS))
	})
	mux.HandleFunc("/fonts.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(webui.FontsCSS))
	})
	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		meta := map[string]any{"labels": labels, "hasPanel": panel != nil, "cape": capeSnapshot(), "capeGauge": capeGauge()}
		if panel != nil {
			f := FitParametric(*panel, panel.Weights)
			meta["weights"] = panel.Weights
			meta["mu"] = f.Mu
			meta["sigma"] = f.Sigma
			meta["df"] = f.Df
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(meta)
	})
	mux.HandleFunc("/api/fit", func(w http.ResponseWriter, r *http.Request) {
		if panel == nil {
			http.Error(w, "no portfolio", http.StatusBadRequest)
			return
		}
		var body struct {
			Weights []float64 `json:"weights"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f := FitParametric(*panel, body.Weights)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]float64{"mu": f.Mu, "sigma": f.Sigma, "df": f.Df})
	})
	// Every simulation endpoint shares the same shape: POST a Params, get a
	// JSON result. post factors the boilerplate once.
	post := func(path string, compute func(Params) any) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
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
			_ = json.NewEncoder(w).Encode(compute(pr))
		})
	}
	post("/api/sim", func(pr Params) any { return ComputeWithPanel(pr, panel) })
	post("/api/models", func(pr Params) any { return Models(pr, panel) })
	post("/api/paths", func(pr Params) any { return Paths(pr, panel) })
	post("/api/sensitivity", func(pr Params) any { return Sensitivity(pr, panel) })
	post("/api/frontier", func(pr Params) any { return Frontier(pr, panel) })
	post("/api/solvemenu", func(pr Params) any { return SolveMenu(pr, panel) })
	post("/api/solve", func(pr Params) any { return Solve(pr, panel) })
	post("/api/spending", func(pr Params) any { return Spending(pr, panel) })
	post("/api/lifecycle", func(pr Params) any { return Lifecycle(pr, panel) })
	post("/api/curves", func(pr Params) any { return Curves(pr, panel) })
	return mux
}
