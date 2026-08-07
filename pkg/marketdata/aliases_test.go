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
		// The fee-free index benchmarks resolve to their own catalog ids,
		// dashed and case-insensitive; the investable ETFs keep separate ids.
		"MSCIWORLD":   "MSCIWORLD",
		"MSCI-WORLD":  "MSCIWORLD",
		"msciworld":   "MSCIWORLD",
		"msci-world":  "MSCIWORLD",
		" MSCIWorld ": "MSCIWORLD",
		"SP500":       "SP500",
		"SP-500":      "SP500",
		"sp500":       "SP500",
		"sp-500":      "SP500",
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

// TestIndexAliasesResolve locks the fee-free index benchmarks and every
// spelling (plain, dashed, lower-case, SIM) so a bare "MSCIWORLD"/"SP500" can
// never again drift onto an unrelated fuzzy match. They are their own catalog
// ids, served long by default from the embedded reconstruction, so the SIM
// suffix is a harmless no-op that still resolves to the same id.
func TestIndexAliasesResolve(t *testing.T) {
	cases := map[string]struct {
		base string
		sim  bool
	}{
		"MSCIWORLD":     {"MSCIWORLD", false},
		"MSCI-WORLD":    {"MSCIWORLD", false},
		"msciworld":     {"MSCIWORLD", false},
		"MSCIWORLDSIM":  {"MSCIWORLD", true},
		"msci-worldsim": {"MSCIWORLD", true},
		"SP500":         {"SP500", false},
		"SP-500":        {"SP500", false},
		"sp500sim":      {"SP500", true},
		"SP-500SIM":     {"SP500", true},
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

func TestKnownLocal(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"NTSG", true},         // catalog id
		{" ntsg ", true},       // normalization
		{"IE00B0M62X26", true}, // catalog ISIN (IBCI)
		{"GOLDSIM", true},      // SIM suffix on a catalog alias
		{"IWDA", true},         // embedded European ticker
		{"ZZZNOTANID", false},  // unknown
		{"^GSPC", false},       // quote symbols are deliberately not local
		{"", false},
	}
	for _, c := range cases {
		if got := KnownLocal(c.id); got != c.want {
			t.Errorf("KnownLocal(%q) = %v, want %v", c.id, got, c.want)
		}
	}
}
