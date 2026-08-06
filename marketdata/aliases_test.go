package marketdata

import "testing"

func TestCanonicalID(t *testing.T) {
	cases := map[string]string{
		"GOLD":                    "XAUUSD",
		"gold":                    "XAUUSD",
		" wti ":                   "CL=F",
		"WINTON-TREND-EQUITY":     "IE000O1VI174",
		"AMUNDI-VOLATILITY":       "LU0319687124",
		"AMUNDI-VOLATILITY-WORLD": "LU0319687124",
		"BHMG":                    "GG00BQBFY362",
		"VOO":                     "VOO", // pas un alias
		"IE00B4L5Y983":            "IE00B4L5Y983",
	}
	for in, want := range cases {
		if got := CanonicalID(in); got != want {
			t.Errorf("CanonicalID(%q) = %q, attendu %q", in, got, want)
		}
	}
}

// TestCanonicalIndexNoConflicts garantit qu'aucun identifiant (ID, ISIN,
// alias, ticker embarqué) ne pointe vers deux actifs différents.
func TestCanonicalIndexNoConflicts(t *testing.T) {
	r := canonicalIndex()
	for _, c := range r.conflicts {
		t.Errorf("collision d'identifiants: %s", c)
	}
	// Les aliases par entrée du catalogue fonctionnent.
	for alias, want := range map[string]string{
		"GOLD":  "XAUUSD",
		"WTI":   "CL=F",
		"BHMG":  "GG00BQBFY362",
		"DJXXF": "DE0002635307",
		"CRRY":  "XS3022291473",
		"CW8":   "LU1681043599",
		"IWDA":  "IE00B4L5Y983", // ticker de la liste embarquée
		"NTSG":  "IE00077IIPQ8",
	} {
		if got := CanonicalID(alias); got != want {
			t.Errorf("CanonicalID(%q) = %q, attendu %q", alias, got, want)
		}
	}
	// Le symbole de cotation du jumeau US n'est PAS indexé: NTSX reste
	// l'UCITS (convention utilisateur), NTSX-US se désigne par son ID.
	if got := CanonicalID("NTSX"); got != "IE000KF370H3" {
		t.Errorf("NTSX doit rester l'UCITS, obtenu %q", got)
	}
	if got := CanonicalID("NTSX-US"); got != "NTSX-US" {
		t.Errorf("NTSX-US doit rester accessible, obtenu %q", got)
	}
}

func TestSplitSim(t *testing.T) {
	cases := map[string]struct {
		base string
		sim  bool
	}{
		"DBMFSIM":                {"DBMF", true},
		"dbmfsim":                {"DBMF", true},
		"VOOSIM":                 {"VOO", true},
		"WINTON-TREND-EQUITYSIM": {"WINTON-TREND-EQUITY", true},
		"IE000O1VI174SIM":        {"IE000O1VI174", true},
		"NTSG":                   {"NTSG", false},
		"GOLD":                   {"GOLD", false}, // ne finit pas par SIM
		"SIM":                    {"SIM", false},  // base vide interdite
	}
	for in, want := range cases {
		base, sim := SplitSim(in)
		if base != want.base || sim != want.sim {
			t.Errorf("SplitSim(%q) = (%q, %v), attendu (%q, %v)", in, base, sim, want.base, want.sim)
		}
	}
}
