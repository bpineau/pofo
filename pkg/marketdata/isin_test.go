package marketdata

import "testing"

func TestIsISIN(t *testing.T) {
	valid := []string{
		"US0378331005", // Apple
		"IE00B4L5Y983", // iShares Core MSCI World
		"FR0000120271", // TotalEnergies
		"DE0007164600", // SAP
	}
	for _, id := range valid {
		if !IsISIN(id) {
			t.Errorf("%s should be a valid ISIN", id)
		}
	}
	invalid := []string{
		"US0378331004", // wrong check digit
		"AAPL",
		"VOO",
		"^GSPC",
		"IWDA.AS",
		"US03783310050", // 13 characters
		"123456789012",  // does not start with two letters
	}
	for _, id := range invalid {
		if IsISIN(id) {
			t.Errorf("%s should not be a valid ISIN", id)
		}
	}
}
