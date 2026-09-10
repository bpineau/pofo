package compare

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// failingClient returns a client whose every outbound base points at a server
// that fails: the fixture must resolve purely from embedded data, so any real
// fetch is a bug we want to see loudly. The field set mirrors marketdata's own
// stubAllBases (client.go).
func failingClient(t *testing.T) *marketdata.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no network in tests", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	c := marketdata.NewClient("")
	c.ChartBase, c.SearchBase, c.StooqBase = srv.URL, srv.URL, srv.URL
	c.FTBase, c.BoursoramaBase, c.MorningstarBase = srv.URL, srv.URL, srv.URL
	c.JustETFBase, c.EurostatBase, c.FredBase = srv.URL, srv.URL, srv.URL
	c.ECBBase, c.CBOEBase = srv.URL, srv.URL
	return c
}

// TestComputeOffline runs the full Compute pipeline on the two source:"index"
// catalog assets (MSCIWORLD, SP500), served from the embedded reconstruction:
// no live symbol, so with every base stubbed to failure the run proves it needs
// no network. USD deflates by the embedded ^CPI-US, so real stats stay offline
// too. The assertions pin shape, not bytes: real quotes drift, embedded
// reconstructions do not, but we keep the checks structural for robustness.
func TestComputeOffline(t *testing.T) {
	c := failingClient(t)

	spec, err := portfolio.Parse("idx", strings.NewReader("60 MSCIWORLD\n40 SP500\n"))
	if err != nil {
		t.Fatal(err)
	}
	cmp, err := Compute(context.Background(), c, []*portfolio.Spec{spec}, Options{
		Currency: "USD", Benchmark: "", NoFees: true, Rebalance: 90,
		Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		t.Fatal(err)
	}
	cols := cmp.Columns()
	if len(cols) != 1 || cols[0].Name != "idx" {
		t.Fatalf("columns = %+v", cols)
	}
	if len(cols[0].WinValues) < 2 || cols[0].WinValues[0] != 100 {
		t.Errorf("window not rebased to 100: %v", cols[0].WinValues[:1])
	}
}

// ExampleCompute shows the library entry point end to end on the embedded
// index assets. It prints a shape invariant (the column count) rather than
// drifting values, so the Output stays stable offline.
func ExampleCompute() {
	client := marketdata.NewClient("") // "" = no disk cache
	spec, _ := portfolio.Parse("idx", strings.NewReader("60 MSCIWORLD\n40 SP500\n"))
	cmp, err := Compute(context.Background(), client, []*portfolio.Spec{spec}, Options{
		Currency: "USD", NoFees: true, Rebalance: 90,
		Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("columns:", len(cmp.Columns()))
	// Output: columns: 1
}

// An optimized file is shown as TWO columns: the weights as written, and the
// weights the optimizer chose, so the report compares the two rather than
// silently replacing one with the other. Only the computed column carries the
// optimizer's account of its choice.
func TestComputeOptimizeAddsTheComputedColumn(t *testing.T) {
	c := failingClient(t)
	spec, err := portfolio.Parse("idx", strings.NewReader(
		"#meta optimize:min-volatility,max-weight:80\n60 MSCIWORLD\n40 SP500\n"))
	if err != nil {
		t.Fatal(err)
	}
	cmp, err := Compute(context.Background(), c, []*portfolio.Spec{spec}, Options{
		Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		t.Fatal(err)
	}
	cols := cmp.Columns()
	if len(cols) != 2 {
		t.Fatalf("columns = %d, want the written one and the optimized one", len(cols))
	}
	if cols[0].Name != "idx (as written)" || cols[1].Name != "idx (min-volatility)" {
		t.Fatalf("names = %q / %q", cols[0].Name, cols[1].Name)
	}
	if cols[0].Note != "" || !strings.Contains(cols[1].Note, "computed by the optimizer") {
		t.Errorf("notes = %q / %q, want only the computed column to explain itself", cols[0].Note, cols[1].Note)
	}
	if cols[0].Assets[0].Weight != 0.6 {
		t.Errorf("the written column moved to %v, want the file's 60 %%", cols[0].Assets[0].Weight)
	}
	if w := cols[1].Assets[0].Weight; w > 0.80+1e-9 {
		t.Errorf("optimized weight %v exceeds the 80 %% cap", w)
	}
	// Both columns share the same window, and the palette gives them distinct
	// identity colors.
	if cols[0].Color == cols[1].Color {
		t.Errorf("both columns colored %q", cols[0].Color)
	}
}

// An unreachable benchmark is a degraded report, not a failed one: the run
// continues without the relative rows rather than aborting a comparison the
// user asked for. Every base is stubbed to failure here, so the benchmark
// cannot resolve; the deadline is what makes that failure quick, since a
// resolution attempt sweeps every source with a backoff between retries.
func TestComputeSurvivesAnUnreachableBenchmark(t *testing.T) {
	c := failingClient(t)
	spec, err := portfolio.Parse("idx", strings.NewReader("100 MSCIWORLD\n"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmp, err := Compute(ctx, c, []*portfolio.Spec{spec}, Options{
		Currency: "USD", Benchmark: "NOSUCHTICKER", NoFees: true, Rebalance: 90,
		Framework: suggest.RegimeFramework(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if cmp.bench != nil {
		t.Error("a benchmark that could not be fetched was kept")
	}
	if cmp.columns[0].hasRel || cmp.columns[0].stats.HasBeta {
		t.Error("relative statistics computed without a benchmark")
	}
	// The rows still exist (the table's shape is fixed), reading "-".
	if got := row(t, cmp.StatRows(), "Beta").Cells[0].Text; got != "-" {
		t.Errorf("beta = %q, want \"-\"", got)
	}
}

// The failure modes of the pipeline, each of which must name what went wrong:
// nothing to compare, and an identifier no source can resolve.
func TestComputeErrors(t *testing.T) {
	c := failingClient(t)
	opt := Options{Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework()}

	if _, err := Compute(context.Background(), c, nil, opt); err == nil {
		t.Error("an empty comparison was computed")
	}

	spec, err := portfolio.Parse("bad", strings.NewReader("100 NOSUCHTICKER\n"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = Compute(ctx, c, []*portfolio.Spec{spec}, opt)
	if err == nil {
		t.Fatal("an unresolvable holding was accepted")
	}
	if !strings.Contains(err.Error(), "bad") || !strings.Contains(err.Error(), "NOSUCHTICKER") {
		t.Errorf("error = %v, want it to name the portfolio and the asset", err)
	}
}
