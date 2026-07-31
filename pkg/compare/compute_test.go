package compare

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
