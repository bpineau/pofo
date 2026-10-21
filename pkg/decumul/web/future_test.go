package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Every sample a fan hands out opens as the very path the chart drew: the
// replay draws that one path from its own stream and the ledger closes each
// year on the drawn wealth.
func TestOpenFutureReplaysTheDrawnSample(t *testing.T) {
	pr := Params{Capital: 900_000, NeedAnnual: 42_000, Years: 35, BufferYears: 2,
		Mu: 0.045, Sigma: 0.13, Df: 5, TaxRate: 0.314, GainFrac: 0.5, Inflation: 0.02,
		PensionYear: 12, PensionAnnual: 14_000, PensionDrift: -0.005, NPaths: 600, Age: 55}
	base := pr.plan()
	sources := modelSources(pr, nil)
	paths := Paths(pr, nil)
	sawRuin := false
	for fi, name := range fanModels {
		p, ok := fanPlan(base, sources, name)
		if !ok {
			t.Fatalf("no %s fan", name)
		}
		e := p.Simulate(pr.NPaths, simWorkers, fanSeed)
		for k, s := range e.Fan(fanPercentiles, 8).Samples {
			if ref := paths.Fans[fi].Samples[k]; ref.Index != s.Index || ref.Ruined != s.Ruined {
				t.Fatalf("%s sample %d: the page was handed %+v, the fan drew index %d", name, k, ref, s.Index)
			}
			q := pr
			q.PathModel, q.PathIndex = name, s.Index
			got := OpenFuture(q, nil)
			if got.Note != "" || len(got.Years) != pr.Years {
				t.Fatalf("%s sample %d: %q, %d rows", name, k, got.Note, len(got.Years))
			}
			for y, row := range got.Years {
				if !row.Broke && row.Close != s.Wealth[y+1] {
					t.Fatalf("%s sample %d year %d: closes at %v, the fan drew %v", name, k, y, row.Close, s.Wealth[y+1])
				}
				if row.Age != 55+y {
					t.Fatalf("age %d in year %d", row.Age, y)
				}
			}
			if got.Ruined != s.Ruined {
				t.Fatalf("%s sample %d: ruined %v, the fan drew %v", name, k, got.Ruined, s.Ruined)
			}
			sawRuin = sawRuin || got.Ruined
		}
	}
	if !sawRuin {
		t.Error("no sample future ran out: the test plan does not exercise a failed ledger")
	}
}

// A request that names no fan, or a path the run never drew, is answered with
// a note rather than a future nobody saw; through the handler, an index past
// the cap is clamped before it reaches the model.
func TestOpenFutureRefusesWhatWasNotDrawn(t *testing.T) {
	pr := Params{Capital: 900_000, NeedAnnual: 42_000, Years: 30, Mu: 0.045, Sigma: 0.13, Df: 5, NPaths: 500}
	for _, q := range []Params{
		{PathModel: "Crystal ball", PathIndex: 1},
		{PathModel: "Student-t", PathIndex: 500},
		{PathModel: "Student-t", PathIndex: -1},
	} {
		in := pr
		in.PathModel, in.PathIndex = q.PathModel, q.PathIndex
		if got := OpenFuture(in, nil); got.Note == "" || len(got.Years) != 0 {
			t.Errorf("%s #%d opened: %+v", q.PathModel, q.PathIndex, got)
		}
	}
	body := `{"capital":900000,"needAnnual":42000,"years":30,"mu":0.045,"sigma":0.13,"df":5,` +
		`"nPaths":500,"pathModel":"Student-t","pathIndex":99999999}`
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/path", strings.NewReader(body)))
	var got FutureResult
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("status %d, %v", rec.Code, err)
	}
	if got.Note == "" || got.Index != maxPaths {
		t.Errorf("an index past the cap was not clamped and refused: %+v", got)
	}
}
