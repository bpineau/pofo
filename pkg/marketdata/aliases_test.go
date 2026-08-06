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
		"VOO":                     "VOO", // not an alias
		"IE00B4L5Y983":            "IE00B4L5Y983",
	}
	for in, want := range cases {
		if got := CanonicalID(in); got != want {
			t.Errorf("CanonicalID(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestCanonicalIndexNoConflicts guarantees that no identifier (ID, ISIN,
// alias, embedded ticker) points to two different assets.
func TestCanonicalIndexNoConflicts(t *testing.T) {
	r := canonicalIndex()
	for _, c := range r.conflicts {
		t.Errorf("identifier collision: %s", c)
	}
	// Per-entry catalog aliases work.
	for alias, want := range map[string]string{
		"GOLD":  "XAUUSD",
		"WTI":   "CL=F",
		"BHMG":  "GG00BQBFY362",
		"DJXXF": "DE0002635307",
		"CRRY":  "XS3022291473",
		"CW8":   "LU1681043599",
		"IWDA":  "IE00B4L5Y983", // ticker from the embedded list
		"NTSG":  "IE00077IIPQ8",
	} {
		if got := CanonicalID(alias); got != want {
			t.Errorf("CanonicalID(%q) = %q, want %q", alias, got, want)
		}
	}
	// The US twin's quote symbol is NOT indexed: NTSX stays the UCITS
	// (the user-facing convention), NTSX-US is addressed by its ID.
	if got := CanonicalID("NTSX"); got != "IE000KF370H3" {
		t.Errorf("NTSX must stay the UCITS, got %q", got)
	}
	if got := CanonicalID("NTSX-US"); got != "NTSX-US" {
		t.Errorf("NTSX-US must stay reachable, got %q", got)
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
		"GOLD":                   {"GOLD", false}, // does not end in SIM
		"SIM":                    {"SIM", false},  // empty base forbidden
	}
	for in, want := range cases {
		base, sim := SplitSim(in)
		if base != want.base || sim != want.sim {
			t.Errorf("SplitSim(%q) = (%q, %v), want (%q, %v)", in, base, sim, want.base, want.sim)
		}
	}
}
