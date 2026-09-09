package main

import (
	"reflect"
	"strings"
	"testing"
)

// The symbol list is typed by hand: blanks, empty fields and lower case all
// have to survive it.
func TestSplitSymbols(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
	}{
		{"^ESTR", []string{"^ESTR"}},
		{" ^estr , ^euribor3m ,, ^ecb-dfr ", []string{"^ESTR", "^EURIBOR3M", "^ECB-DFR"}},
		{"", nil},
		{" , ,", nil},
	} {
		if got := splitSymbols(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitSymbols(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// The catalog is what -rates prints when it is asked what it understands: the
// registry first, then the market yields, then a worked example.
func TestPrintRateCatalog(t *testing.T) {
	out := captureOutput(t, printRateCatalog)
	for _, want := range []string{"^ESTR", "^ECB-DFR", "^IRX", "^TYX", "Example: pofo -rates"} {
		if !strings.Contains(out, want) {
			t.Errorf("the rate catalog does not mention %q:\n%s", want, out)
		}
	}
}

// "list" and an empty list both mean "tell me what you know", and neither
// fetches anything: a nil client would panic if either did.
func TestRunRatesCatalogPaths(t *testing.T) {
	for _, list := range []string{"list", "LIST", " , "} {
		out := captureOutput(t, func() {
			if err := runRates(t.Context(), &options{}, nil, list); err != nil {
				t.Errorf("runRates(%q): %v", list, err)
			}
		})
		if !strings.Contains(out, "Policy and money-market rates") {
			t.Errorf("runRates(%q) printed no catalog:\n%s", list, out)
		}
	}
}
