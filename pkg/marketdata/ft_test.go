package marketdata

import "testing"

func TestFTSymbolCurrency(t *testing.T) {
	tests := []struct {
		symbol, want string
	}{
		{"LU0171310443:EUR", "EUR"},
		{"NTSG:GER:EUR", "EUR"},
		{"IWDA:LSE:GBX", "GBX"},
		// The last segment can be an exchange code, not a currency (the
		// real WEBN case: the Munich listing carries no currency segment).
		// Callers must then fall back to the FT chart API's currency.
		{"WEBN:MUN", ""},
		{"XDWD:BER", ""},
		{"LU0171310443", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := ftSymbolCurrency(tc.symbol); got != tc.want {
			t.Errorf("ftSymbolCurrency(%q) = %q, want %q", tc.symbol, got, tc.want)
		}
	}
}
