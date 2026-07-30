package marketdata_test

import (
	"fmt"
	"time"

	"portfodor/pkg/marketdata"
)

// CanonicalID suit les chaînes d'aliases vers l'identifiant canonique.
func ExampleCanonicalID() {
	fmt.Println(marketdata.CanonicalID("gold"))
	fmt.Println(marketdata.CanonicalID("AMUNDI-VOLATILITY"))
	fmt.Println(marketdata.CanonicalID("VOO"))
	// Output:
	// XAUUSD
	// LU0319687124
	// VOO
}

// FundISIN traduit les tickers d'ETF/OPCVM européens en ISIN grâce à la
// liste de correspondance embarquée.
func ExampleFundISIN() {
	isin, ok := marketdata.FundISIN("IWDA")
	fmt.Println(isin, ok)
	// Output:
	// IE00B4L5Y983 true
}

// Example_fetch montre l'utilisation type du client: résolution
// (alias → ISIN → source), téléchargement et cache disque transparents.
// (Non exécuté: nécessite le réseau.)
func Example_fetch() {
	client := marketdata.NewClient("data")
	client.Logf = func(format string, args ...any) { /* journalisation optionnelle */ }

	series, err := client.Fetch("IWDA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s — %d cotations depuis %s\n",
		series.Name, len(series.Points), series.First().Date.Format("2006-01-02"))
}
