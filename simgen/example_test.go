package simgen

import (
	"fmt"
	"time"
)

// Composite assemble un indice base 100 à poids constants à partir de
// composants quelconques — ici un 90/60 actions/obligations avec jambe
// « excess » (futures) financée au taux cash et 0,20 %/an de frais.
func ExampleComposite() {
	fetch := fakeFetcher{
		"ACTIONS": mkSeries("ACTIONS", 300, 0.0008),
		"OBLIG":   mkSeries("OBLIG", 300, 0.0002),
		"^IRX":    mkLevels("^IRX", 300, 3.0), // taux annualisé en %
	}
	fr, err := BuildFrame(fetch, []string{"ACTIONS", "OBLIG", "^IRX"}, day(0))
	if err != nil {
		panic(err)
	}
	values, err := Composite(fr, []Leg{
		{ID: "ACTIONS", Weight: 0.90},
		{ID: "OBLIG", Weight: 0.60, Excess: true},
		{ID: "^IRX", Weight: 0.10},
	}, "^IRX", 0.0020)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d points, base %.0f\n", len(values), values[0])
	// Output:
	// 300 points, base 100
}

// TSMOM rejoue une stratégie time-series momentum paramétrable sur un
// panier de marchés. (Exemple d'API ; non exécuté.)
func Example_tsmom() {
	var fetch Fetcher // p.ex. marketdata.NewClient("data")
	fr, _ := BuildFrame(fetch, []string{"^IRX", "VFINX", "GC=F"}, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	values, start, _ := TSMOM(fr, TSMOMConfig{
		Markets:  []string{"VFINX", "GC=F"},
		CashID:   "^IRX",
		Lookback: 252, VolWindow: 63, Rebalance: 21,
		TargetVol: 0.10, MaxLeverage: 2,
	})
	_ = values[start:]
}
