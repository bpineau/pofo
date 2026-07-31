package replay_test

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/replay"
)

// Replaying the same household through the 1973 crisis: same capital, same
// target, same forty-year plan, seven withdrawal rules.
func ExampleRun() {
	res, err := replay.Run(replay.Setup{
		Start: 1973, Capital: 600000, Spend: 24000, Years: 40,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d-%d, real CAGR %.1f%%, worst year %d at %.0f%%\n",
		res.Start, res.End, res.CAGR*100, res.WorstAt, res.WorstYear*100)
	for _, r := range res.Rules[:2] {
		fmt.Printf("%-10s mean %.0fk, leanest %.0fk, %d lean years, %.0fk left\n",
			r.Name, r.Mean/1000, r.Low/1000, r.LeanYears, r.Final/1000)
	}
	// Output:
	// 1973-2012, real CAGR 4.7%, worst year 1974 at -23%
	// Fixed      mean 24k, leanest 24k, 0 lean years, 19k left
	// Flex -10%  mean 22k, leanest 22k, 33 lean years, 344k left
}

// The reference series is the bundled US 60/40 in real terms.
func ExampleReference() {
	s, err := replay.Reference()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s, %d complete years from %d\n", replay.Label, len(s.Years), s.Years[0])
	// Output:
	// US 60/40 (S&P 500 + 5-year Treasuries, rebalanced yearly), 72 complete years from 1954
}
