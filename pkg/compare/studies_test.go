package compare

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// parseSpec parses one portfolio file body, failing the test on error.
func parseSpec(t *testing.T, name, body string) *portfolio.Spec {
	t.Helper()
	spec, err := portfolio.Parse(name, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

// Every column is an analyze study, and Studies hands them out in column
// order. The two embedded index assets start years apart, so the common
// window cuts the older column and leaves the younger one whole: the whole
// column must read its study's numbers as they are, and the cut one the same
// calls on the same Sim.Index, over the common window only.
func TestComputeColumnsAreStudies(t *testing.T) {
	c := failingClient(t)
	specs := []*portfolio.Spec{
		parseSpec(t, "us", "100 SP500\n"),
		parseSpec(t, "world", "100 MSCIWORLD\n"),
	}
	cmp, err := Compute(context.Background(), c, specs, Options{
		Currency: "USD", Benchmark: "SP500", Rebalance: 90, Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		t.Fatal(err)
	}
	cols, studies := cmp.Columns(), cmp.Studies()
	if len(studies) != len(cols) {
		t.Fatalf("%d studies for %d columns", len(studies), len(cols))
	}
	for i, st := range studies {
		if st == nil || st.Spec.Name != cols[i].Name || st.Sim.Dates[0] != cols[i].SimDates[0] {
			t.Fatalf("study %d does not describe column %q", i, cols[i].Name)
		}
		if len(st.Holdings) != 1 || len(st.Correlation) != 1 {
			t.Errorf("study %d: %d holdings, %d correlation rows, want 1 and 1", i, len(st.Holdings), len(st.Correlation))
		}
	}

	us, world := cmp.columns[0], cmp.columns[1]
	if !us.sim.Dates[0].Before(world.sim.Dates[0]) {
		t.Fatalf("fixture degenerate: the S&P 500 column starts %s, not before the world's %s",
			us.sim.Dates[0].Format(time.DateOnly), world.sim.Dates[0].Format(time.DateOnly))
	}
	if world.stats != world.study.Stats || !world.hasRel || world.rel != *world.study.Relative {
		t.Error("the column spanning its whole simulation does not read its study's statistics")
	}
	i, j := window(us.sim.Dates, cmp.CommonStart(), cmp.CommonEnd())
	want, err := metrics.Compute(us.sim.Dates[i:j], us.sim.Index[i:j])
	if err != nil {
		t.Fatal(err)
	}
	if us.stats.CAGR != want.CAGR || us.stats.Volatility != want.Volatility || us.stats.MaxDrawdown != want.MaxDrawdown {
		t.Errorf("cut column CAGR/vol/MDD = %v/%v/%v, want %v/%v/%v on the common window of its Sim.Index",
			us.stats.CAGR, us.stats.Volatility, us.stats.MaxDrawdown, want.CAGR, want.Volatility, want.MaxDrawdown)
	}
	// The S&P 500 against itself: a beta of one, measured on the window.
	if !us.stats.HasBeta || us.stats.Beta < 0.999 || us.stats.Beta > 1.001 {
		t.Errorf("cut column beta = %v (set %v), want 1 against itself", us.stats.Beta, us.stats.HasBeta)
	}
	if us.stats == us.study.Stats {
		t.Error("the cut column reads its study's whole-window statistics")
	}
}

// The optimizer's column is a study too, of the weights the optimizer chose:
// its spec carries them exactly, under the column's name, with the directive
// the comparison already acted on cleared.
func TestComputeOptimizedColumnIsStudied(t *testing.T) {
	c := failingClient(t)
	spec := parseSpec(t, "idx", "#meta optimize:min-volatility,max-weight:80\n60 MSCIWORLD\n40 SP500\n")
	cmp, err := Compute(context.Background(), c, []*portfolio.Spec{spec}, Options{
		Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		t.Fatal(err)
	}
	cols, studies := cmp.Columns(), cmp.Studies()
	if len(studies) != 2 {
		t.Fatalf("%d studies, want the written one and the optimized one", len(studies))
	}
	for i, st := range studies {
		if st.Spec.Name != cols[i].Name || st.Spec.Optimize != nil {
			t.Errorf("study %d: spec %q (optimize %v), want the column's name and no directive left", i, st.Spec.Name, st.Spec.Optimize)
		}
		for k, h := range st.Spec.Holdings {
			if h.Weight != cols[i].Assets[k].Weight || st.Portfolio.Assets[k].Weight != h.Weight {
				t.Errorf("study %d holding %s: spec %v, built %v, column %v", i, h.ID, h.Weight, st.Portfolio.Assets[k].Weight, cols[i].Assets[k].Weight)
			}
		}
	}
	if spec.Optimize == nil || spec.Holdings[0].Weight != 0.6 {
		t.Error("Compute rewrote the caller's spec")
	}
}

// A comparison of one levered and one decumulating spec, each failing its
// own way: the financing rate cannot be fetched (every base fails), and the
// withdrawals exhaust the capital. Both still render, and the report's
// warnings name the ruin, as the study's do.
func TestComputeColumnWarnings(t *testing.T) {
	c := failingClient(t)
	specs := []*portfolio.Spec{
		parseSpec(t, "levered", "#meta leverage:on\n90 SP500\n60 MSCIWORLD\n"),
		parseSpec(t, "spent", "#meta capital:1000\n#meta withdraw:300/year\n100 MSCIWORLD\n"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmp, err := Compute(ctx, c, specs, Options{Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework()})
	if err != nil {
		t.Fatal(err)
	}
	levered, spent := cmp.columns[0], cmp.columns[1]
	if levered.p.Cash != nil {
		t.Error("a financing rate was found with every base failing")
	}
	if !spent.sim.Ruined {
		t.Fatal("fixture degenerate: 300 a year out of 1000 did not ruin the plan")
	}
	ruin := func(ws []string) bool {
		for _, w := range ws {
			if strings.HasPrefix(w, "capital wiped out on ") && strings.Contains(w, "withdrawals exhausted the capital") {
				return true
			}
		}
		return false
	}
	if !ruin(spent.warnings) || !ruin(spent.study.Warnings) {
		t.Errorf("ruin not reported: report %q, study %q", spent.warnings, spent.study.Warnings)
	}
	if page := cmp.HTMLPage(Decoration{}); !ruin(page.Portfolios[1].Warnings) {
		t.Errorf("the page lost the ruin: %q", page.Portfolios[1].Warnings)
	}
}

// countingClient is failingClient that counts the requests reaching the
// network, so a test can tell a memoized answer from a fresh fetch.
func countingClient(t *testing.T) (*marketdata.Client, *atomic.Int64) {
	t.Helper()
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "no network in tests", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	c := marketdata.NewClient("")
	c.ChartBase, c.SearchBase, c.StooqBase = srv.URL, srv.URL, srv.URL
	c.FTBase, c.BoursoramaBase, c.MorningstarBase = srv.URL, srv.URL, srv.URL
	c.JustETFBase, c.EurostatBase, c.FredBase = srv.URL, srv.URL, srv.URL
	c.ECBBase, c.CBOEBase = srv.URL, srv.URL
	return c, &hits
}

// The source applies the comparison's policy and answers every request once:
// a series shared by two columns is the same series, a failure is not
// retried, -no-simulate wins over a SIM request, only an identifier outside
// the catalog resolves exactly, and -no-fees leaves every TER unknown.
func TestSourcePolicyAndMemo(t *testing.T) {
	c, hits := countingClient(t)
	src := newSource(c, Options{NoSim: true, NoFees: true, ExactForeign: true})
	ctx := context.Background()

	a, err := src.FetchExtended(ctx, "SP500SIM", src.holding("USD"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := src.FetchExtended(ctx, "SP500SIM", src.holding("USD"))
	if a != b {
		t.Error("the same request fetched twice")
	}

	short, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	_, err1 := src.FetchExtended(short, "NOSUCHTICKER", src.holding("USD"))
	before := hits.Load()
	_, err2 := src.FetchExtended(ctx, "NOSUCHTICKER", src.holding("USD"))
	if err1 == nil || !errors.Is(err2, err1) || hits.Load() != before {
		t.Errorf("a failed request was retried: %v then %v, %d more hits", err1, err2, hits.Load()-before)
	}

	for key := range src.memo {
		if !key.noSim {
			t.Errorf("%s fetched with SIM despite NoSim", key.id)
		}
		if key.exact != (key.id == "NOSUCHTICKER") {
			t.Errorf("%s resolved exact=%v, want exact only outside the catalog", key.id, key.exact)
		}
	}
	if _, ok := src.Fees(ctx, "SP500"); ok {
		t.Error("a TER was looked up under NoFees")
	}
	_, ok1 := newSource(c, Options{}).Fees(ctx, "SP500")
	if _, ok2 := c.Fees(ctx, "SP500"); ok1 != ok2 {
		t.Errorf("TER known = %v, want the client's %v", ok1, ok2)
	}
}
