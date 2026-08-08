package simgen

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// quoted marks a canned series as market data, so the audit accepts it as an
// independent reference (a series pofo reconstructs is refused on purpose).
func quoted(s *marketdata.Series) *marketdata.Series {
	s.Source = "yahoo"
	s.Currency = "USD"
	return s
}

// scaled returns the series with every daily return multiplied by k: same
// shape, a different level, which is exactly the "hot engine" failure the
// level verdict hunts.
func scaled(symbol string, s *marketdata.Series, k float64) *marketdata.Series {
	out := &marketdata.Series{Symbol: symbol, Currency: s.Currency, Source: s.Source}
	v := 100.0
	for i, p := range s.Points {
		if i > 0 {
			v *= 1 + k*(p.Close/s.Points[i-1].Close-1)
		}
		out.Points = append(out.Points, marketdata.Point{Date: p.Date, Close: v})
	}
	return out
}

// staticRecipe wraps a canned series as a recipe's engine output.
func staticRecipe(id string, engine *marketdata.Series, donors ...string) Recipe {
	return Recipe{
		ID: id, Name: id + " test", Method: "canned",
		Donors: donors,
		Build: func(Fetcher, time.Time) (*marketdata.Series, error) {
			return engine, nil
		},
	}
}

func TestAuditMeasuresEngineAgainstQuotes(t *testing.T) {
	real := quoted(mkWobbly("X", 900, 4e-4, 0.01))
	engine := scaled("X engine", real, 1.3) // same path, 30 % hotter
	f := fakeFetcher{"X": real}

	a := Audit(f, staticRecipe("X", engine))
	if !a.Measured() {
		t.Fatalf("audit failed: %s", a.Err)
	}
	if a.Reference != "X" {
		t.Errorf("reference = %q, want X", a.Reference)
	}
	if a.MonthlyCorr < 0.95 {
		t.Errorf("monthly correlation = %.3f, want a near-perfect shape match", a.MonthlyCorr)
	}
	if a.Delta <= 0 {
		t.Errorf("a levered engine must read hot, got %+.4f/yr", a.Delta)
	}
	if a.VolSim <= a.VolReal {
		t.Errorf("vol engine %.4f should exceed real %.4f", a.VolSim, a.VolReal)
	}
	// TotalDrift is the gap compounded over the window.
	if want := math.Pow(1+a.Delta, a.Years) - 1; math.Abs(a.TotalDrift-want) > 1e-9 {
		t.Errorf("TotalDrift = %.6f, want %.6f", a.TotalDrift, want)
	}
	if a.Level != VerdictBad {
		t.Errorf("level = %s, want bad for a %+.2f %%/yr gap", a.Level, a.Delta*100)
	}
	if a.RealFrom.IsZero() {
		t.Error("RealFrom should carry the date from which a SIM consumer gets real quotes")
	}
	if len(a.Engine.Points) == 0 || len(a.Real.Points) == 0 {
		t.Error("the audit must return both curves, clipped to the window")
	}
}

func TestAuditRejectsCircularReference(t *testing.T) {
	real := quoted(mkWobbly("Y", 600, 3e-4, 0.008))
	f := fakeFetcher{"Y": real}

	// The engine IS the reference: a perfect fit that proves nothing.
	a := Audit(f, staticRecipe("Y", real))
	if a.Measured() {
		t.Fatalf("a self-referential audit must not report statistics, got %+v", a)
	}
	if !strings.Contains(a.Err, "already inside the engine") {
		t.Errorf("Err = %q, want the circularity spelled out", a.Err)
	}
	if a.Level != VerdictUnknown || a.Path != VerdictUnknown {
		t.Errorf("verdicts = %s/%s, want n/a", a.Level, a.Path)
	}
}

func TestAuditRejectsReconstructionAsReference(t *testing.T) {
	real := mkWobbly("Z", 600, 3e-4, 0.008)
	real.Source = "simdata" // a pofo reconstruction, not a quote
	f := fakeFetcher{"Z": real}

	a := Audit(f, staticRecipe("Z", scaled("Z engine", real, 1.1)))
	if a.Measured() {
		t.Fatal("a reconstruction cannot arbitrate another reconstruction")
	}
	if !strings.Contains(a.Err, "a pofo reconstruction") {
		t.Errorf("Err = %q, want the rejected candidate named", a.Err)
	}
}

func TestAuditFailedEngineHeadsTheList(t *testing.T) {
	r := staticRecipe("W", nil)
	r.Build = func(Fetcher, time.Time) (*marketdata.Series, error) {
		return nil, errUnbuildable
	}
	a := Audit(fakeFetcher{}, r)
	if a.Measured() || !strings.Contains(a.Err, "engine failed") {
		t.Fatalf("Err = %q, want the build error reported", a.Err)
	}
	if a.Score <= 0 {
		t.Error("a broken engine must sort first")
	}
}

var errUnbuildable = &buildError{}

type buildError struct{}

func (*buildError) Error() string { return "missing component" }

func TestAuditGradesDonorChain(t *testing.T) {
	real := quoted(mkWobbly("F", 900, 4e-4, 0.01))
	near := quoted(mkWobbly("NEAR", 1200, 3e-4, 0.01))
	deep := quoted(mkWobbly("DEEP", 1500, 2e-4, 0.01))
	// The chain reaches further back than the fund, junction by junction.
	shift := func(s *marketdata.Series, back int) *marketdata.Series {
		out := &marketdata.Series{Symbol: s.Symbol, Currency: s.Currency, Source: s.Source}
		for _, p := range s.Points {
			out.Points = append(out.Points, marketdata.Point{Date: p.Date.AddDate(0, 0, -back), Close: p.Close})
		}
		return out
	}
	f := fakeFetcher{"F": real, "NEAR": shift(near, 400), "DEEP": shift(deep, 900)}

	a := Audit(f, staticRecipe("F", scaled("F engine", real, 1.3), "NEAR", "DEEP"))
	if len(a.Chain) != 2 {
		t.Fatalf("chain has %d junctions, want one per donor link", len(a.Chain))
	}
	for _, j := range a.Chain {
		if !j.Measured {
			t.Errorf("junction %q unmeasured: %s", j.Pair, j.Note)
			continue
		}
		if j.Months < 12 {
			t.Errorf("junction %q graded on %d months", j.Pair, j.Months)
		}
		if math.IsNaN(j.Corr) {
			t.Errorf("junction %q has no correlation", j.Pair)
		}
	}
	if !strings.Contains(a.Chain[0].Pair, "NEAR") {
		t.Errorf("first junction = %q, want the nearest donor first", a.Chain[0].Pair)
	}
}

func TestAuditChainSkipsMismatchedCurrencies(t *testing.T) {
	real := quoted(mkWobbly("G", 900, 4e-4, 0.01))
	donor := quoted(mkWobbly("EURDONOR", 1200, 3e-4, 0.01))
	donor.Currency = "EUR"
	f := fakeFetcher{"G": real, "EURDONOR": donor}

	a := Audit(f, staticRecipe("G", scaled("G engine", real, 1.3), "EURDONOR"))
	if len(a.Chain) != 1 || a.Chain[0].Measured {
		t.Fatalf("chain = %+v, want the junction left ungraded", a.Chain)
	}
	if !strings.Contains(a.Chain[0].Note, "currencies") {
		t.Errorf("note = %q, want the currency mismatch explained", a.Chain[0].Note)
	}
}

func TestGradeThresholds(t *testing.T) {
	for _, tc := range []struct {
		name        string
		a           AuditResult
		level, path Verdict
	}{
		{"tight", AuditResult{Delta: 0.005, MonthlyCorr: 0.95, TrackingErr: 0.02, VolReal: 0.10}, VerdictOK, VerdictOK},
		{"level warns", AuditResult{Delta: 0.015, MonthlyCorr: 0.95, TrackingErr: 0.02, VolReal: 0.10}, VerdictWarn, VerdictOK},
		{"level fails", AuditResult{Delta: -0.030, MonthlyCorr: 0.95, TrackingErr: 0.02, VolReal: 0.10}, VerdictBad, VerdictOK},
		{"path warns", AuditResult{Delta: 0.005, MonthlyCorr: 0.80, TrackingErr: 0.02, VolReal: 0.10}, VerdictOK, VerdictWarn},
		{"path fails", AuditResult{Delta: 0.005, MonthlyCorr: 0.60, TrackingErr: 0.08, VolReal: 0.10}, VerdictOK, VerdictBad},
	} {
		t.Run(tc.name, func(t *testing.T) {
			level, path, _ := grade(tc.a)
			if level != tc.level || path != tc.path {
				t.Errorf("grade = %s/%s, want %s/%s", level, path, tc.level, tc.path)
			}
		})
	}
	// A short window is discounted so noise does not head the list.
	long := AuditResult{Delta: 0.03, MonthlyCorr: 0.8, TrackingErr: 0.05, VolReal: 0.1}
	short := long
	short.Short = true
	_, _, ls := grade(long)
	_, _, ss := grade(short)
	if ss >= ls {
		t.Errorf("short-window score %.2f should sit below %.2f", ss, ls)
	}
}

func TestAuditGroupsCoverEveryRecipe(t *testing.T) {
	known := map[string]bool{}
	for _, r := range All() {
		known[strings.ToUpper(r.ID)] = true
	}
	for _, g := range auditGroups {
		for _, id := range g.IDs {
			if !known[strings.ToUpper(id)] {
				t.Errorf("group %q lists %s, which is no longer a recipe", g.Title, id)
			}
		}
	}
	// Every recipe lands somewhere: the grouped ones in their family, the rest
	// in Other, so a new recipe is never dropped from the report.
	for _, r := range All() {
		if groupOf(r.ID) == "" {
			t.Errorf("%s has no group", r.ID)
		}
	}
	if groupOf("NOT-A-RECIPE") != otherGroup {
		t.Errorf("an unlisted id must fall into %q", otherGroup)
	}
}

func TestAuditAllOrdersWorstFirst(t *testing.T) {
	real := quoted(mkWobbly("A", 900, 4e-4, 0.01))
	other := quoted(mkWobbly("B", 900, 4e-4, 0.01))
	f := fakeFetcher{"A": real, "B": other}
	groups := AuditAll(f, []Recipe{
		staticRecipe("A", scaled("A engine", real, 1.2)),
		staticRecipe("B", scaled("B engine", other, 1.6)),
	}, nil)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want the single Other family", len(groups))
	}
	if got := groups[0].Results[0].ID; got != "B" {
		t.Errorf("first result = %s, want the worse engine B", got)
	}
}

func TestRebaseAndDrift(t *testing.T) {
	if got := Rebase([]float64{2, 4, 1}); got[0] != 100 || got[1] != 200 || got[2] != 50 {
		t.Errorf("Rebase = %v", got)
	}
	if got := Rebase(nil); len(got) != 0 {
		t.Errorf("Rebase(nil) = %v", got)
	}
	real := quoted(mkWobbly("D", 400, 3e-4, 0.006))
	a := Audit(fakeFetcher{"D": real}, staticRecipe("D", scaled("D engine", real, 1.5)))
	dates, vals := a.Drift()
	if len(dates) != len(vals) || len(vals) < 100 {
		t.Fatalf("drift series: %d dates, %d values", len(dates), len(vals))
	}
	if vals[0] != 100 {
		t.Errorf("drift starts at %v, want 100", vals[0])
	}
	if vals[len(vals)-1] <= 100 {
		t.Errorf("a hot engine must drift up, ended at %.1f", vals[len(vals)-1])
	}
}
