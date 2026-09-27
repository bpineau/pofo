package pofo_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// TestExampleRuns runs the package example, whose output is not pinned, so a
// bundled series it reads going missing fails a test rather than a reader.
func TestExampleRuns(t *testing.T) { Example() }

// A first program: two bundled series, a monthly 60/40 of them, and the
// statistics of all three, offline. Its figures move with every data
// refresh, so it prints no pinned output; TestExampleRuns runs it.
func Example() {
	// The S&P 500 total return since 1871, bundled with the module.
	sp500, err := marketdata.Bundled("SP500-USD")
	if err != nil {
		log.Fatal(err)
	}
	bonds, err := marketdata.Bundled("TREASURY-INT-USD")
	if err != nil {
		log.Fatal(err)
	}

	// Monthly returns on the months both share, and a 60/40 rebalanced
	// every month (weights are fractions).
	p, err := marketdata.NewPanel(marketdata.Monthly, sp500, bonds)
	if err != nil {
		log.Fatal(err)
	}
	p, err = p.Mix("60/40", map[string]float64{"SP500-USD": 0.6, "TREASURY-INT-USD": 0.4})
	if err != nil {
		log.Fatal(err)
	}
	for _, id := range p.IDs {
		s, err := p.Series(id)
		if err != nil {
			log.Fatal(err)
		}
		st, err := s.Stats()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-16s CAGR %5.2f %%/yr, volatility %4.1f %%/yr, max drawdown %5.1f %%\n",
			id, st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
	}
}
