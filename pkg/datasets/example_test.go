package datasets_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Catalog returns the bundled assets as typed datasets.Asset records (use
// AssetMeta for the same data as raw JSON). Resolve a ticker/alias/ISIN to a
// single record with marketdata.Lookup.
func ExampleCatalog() {
	for _, a := range datasets.Catalog() {
		if a.ID == "IE00B4L5Y983" { // iShares Core MSCI World (IWDA)
			fmt.Printf("%s: TER %.2f%%, UCITS=%v, US=%g%%\n",
				a.Name, a.Fees, a.UCITS, a.Geography["US"])
		}
	}
	// Output: iShares Core MSCI World UCITS ETF USD (Acc): TER 0.20%, UCITS=true, US=68%
}

// Simdata exposes the embedded simulated histories as a read-only file system:
// one "<canonical id>.csv" per asset, comment stamps then "date,close" rows.
// marketdata.ReadSimdataFS parses one into a Series; a bare fs.ReadFile is
// enough to inspect the stamps.
func ExampleSimdata() {
	b, err := fs.ReadFile(datasets.Simdata(), "DBMF.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	stamp, _, _ := strings.Cut(string(b), "\n")
	fmt.Println(stamp)
	// Output: # pofo simdata v1
}

// Refdata holds the long reference series the simgen recipes are built on
// (indices, government-bond and cash legs, the FCPE NAV snapshots), in the
// same format as Simdata. Regeneration therefore needs no external directory.
func ExampleRefdata() {
	names, err := fs.Glob(datasets.Refdata(), "SP500-USD.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(names)
	// Output: [SP500-USD.csv]
}

// CAPE is the Shiller PE10, one "date,cape" row per month since January 1881,
// dated on the first of the month.
func ExampleCAPE() {
	for _, line := range strings.Split(string(datasets.CAPE()), "\n") {
		if line == "" || strings.HasPrefix(line, "#") || line == "date,cape" {
			continue
		}
		date, _, _ := strings.Cut(line, ",")
		fmt.Println("first month:", date)
		break
	}
	// Output: first month: 1881-01-01
}

// BroadSample is the Jorda-Schularick-Taylor panel of REAL annual total
// returns as fractions ("iso,year,equity,bond,bill"), with empty cells where
// the source has a gap.
func ExampleBroadSample() {
	countries := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(datasets.BroadSample())), "\n") {
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "iso,") {
			continue
		}
		iso, _, _ := strings.Cut(line, ",")
		countries[iso] = true
	}
	fmt.Println("economies:", len(countries))
	// Output: economies: 16
}

// MacroPanel is the OECD monthly panel behind the macro-regime work. Parse it
// with permanent.ParsePanel rather than by hand; the columns are:
func ExampleMacroPanel() {
	for _, line := range strings.Split(string(datasets.MacroPanel()), "\n") {
		if strings.HasPrefix(line, "iso,") {
			fmt.Println(line)
			break
		}
	}
	// Output: iso,date,ip,cpi,shortrate,longrate,shareprice
}

// AssetMeta returns the catalog as raw JSON, for a caller decoding it into its
// own type; Catalog returns the same data as typed datasets.Asset records.
func ExampleAssetMeta() {
	var catalog []struct {
		ID   string  `json:"id"`
		Fees float64 `json:"fees"`
	}
	if err := json.Unmarshal(datasets.AssetMeta(), &catalog); err != nil {
		fmt.Println(err)
		return
	}
	for _, a := range catalog {
		if a.ID == "IE00B4L5Y983" {
			fmt.Printf("IWDA TER: %.2f%%\n", a.Fees)
		}
	}
	// Output: IWDA TER: 0.20%
}
