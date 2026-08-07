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
		// MSCI World and S&P 500 have friendly index aliases, dashed and
		// case-insensitive; the long history is one SIM suffix away.
		"MSCIWORLD":   "IE00B4L5Y983",
		"MSCI-WORLD":  "IE00B4L5Y983",
		"msciworld":   "IE00B4L5Y983",
		"msci-world":  "IE00B4L5Y983",
		" MSCIWorld ": "IE00B4L5Y983",
		"SP500":       "IE00BFMXXD54",
		"SP-500":      "IE00BFMXXD54",
		"sp500":       "IE00BFMXXD54",
		"sp-500":      "IE00BFMXXD54",
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

// TestIndexAliasesResolve locks the friendly MSCI World and S&P 500 aliases
// (and their long-history SIM forms) so a bare "MSCIWORLD"/"SP500" can never
// again drift onto an unrelated fuzzy match. The base id owns simdata
// (IE00B4L5Y983, IE00BFMXXD54), so the SIM suffix reaches the deep
// reconstruction (MSCI World net TR ~1969, S&P 500 total return ~1871).
func TestIndexAliasesResolve(t *testing.T) {
	cases := map[string]struct {
		base string
		sim  bool
	}{
		"MSCIWORLD":     {"IE00B4L5Y983", false},
		"MSCI-WORLD":    {"IE00B4L5Y983", false},
		"msciworld":     {"IE00B4L5Y983", false},
		"MSCIWORLDSIM":  {"IE00B4L5Y983", true},
		"msci-worldsim": {"IE00B4L5Y983", true},
		"SP500":         {"IE00BFMXXD54", false},
		"SP-500":        {"IE00BFMXXD54", false},
		"sp500sim":      {"IE00BFMXXD54", true},
		"SP-500SIM":     {"IE00BFMXXD54", true},
	}
	for in, want := range cases {
		base, sim := SplitSim(in)
		if got := CanonicalID(base); got != want.base || sim != want.sim {
			t.Errorf("%q -> (%q, sim=%v), want (%q, sim=%v)", in, got, sim, want.base, want.sim)
		}
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
