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
			t.Errorf("%s devrait être un ISIN valide", id)
		}
	}
	invalid := []string{
		"US0378331004", // mauvaise clé de contrôle
		"AAPL",
		"VOO",
		"^GSPC",
		"IWDA.AS",
		"US03783310050", // 13 caractères
		"123456789012",  // ne commence pas par deux lettres
	}
	for _, id := range invalid {
		if IsISIN(id) {
			t.Errorf("%s ne devrait pas être un ISIN valide", id)
		}
	}
}
