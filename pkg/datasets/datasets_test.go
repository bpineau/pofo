package datasets

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedDatasetsPresent(t *testing.T) {
	sim, err := fs.ReadDir(Simdata(), ".")
	if err != nil || len(sim) < 8 {
		t.Fatalf("embedded simdata incomplete: %d files, %v", len(sim), err)
	}
}

func TestMacroPanelWellFormed(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(string(MacroPanel())), "\n")
	var rows, header int
	isos := map[string]bool{}
	for _, l := range lines {
		if strings.HasPrefix(l, "#") {
			continue
		}
		if strings.HasPrefix(l, "iso,") {
			header++
			continue
		}
		f := strings.Split(l, ",")
		if len(f) != 7 {
			t.Fatalf("row has %d fields, want 7: %q", len(f), l)
		}
		isos[f[0]] = true
		rows++
	}
	if header != 1 {
		t.Fatalf("expected exactly one header row, got %d", header)
	}
	if rows < 10000 {
		t.Fatalf("macro panel looks truncated: %d rows", rows)
	}
	if len(isos) < 20 {
		t.Fatalf("macro panel covers only %d countries", len(isos))
	}
}

func TestAssetMetaIsValidJSON(t *testing.T) {
	var raw []map[string]any
	if err := json.Unmarshal(AssetMeta(), &raw); err != nil {
		t.Fatalf("AssetMeta is not a JSON array: %v", err)
	}
	if len(raw) < 100 {
		t.Fatalf("AssetMeta looks truncated: %d entries", len(raw))
	}
}

func TestCatalogParsesAndIsTyped(t *testing.T) {
	assets := Catalog()
	if len(assets) < 100 {
		t.Fatalf("Catalog looks truncated: %d assets", len(assets))
	}
	// Spot-check a well-known entry decodes into the full typed record.
	var iwda Asset
	for _, a := range assets {
		if a.ID == "IE00B4L5Y983" {
			iwda = a
			break
		}
	}
	if iwda.Name == "" || !iwda.UCITS || iwda.Fees == 0 || iwda.Geography["US"] == 0 || iwda.AssetClass != "equity" {
		t.Fatalf("IWDA did not decode into a full Asset: %+v", iwda)
	}
}

// TestCatalogEURetailConsistent enforces the invariants of the eu_retail
// flag: every UCITS fund is EU-retail buyable by definition, a US-ISIN
// instrument without a PRIIPs KID never is, and fee-free index series
// (source "index") are not tradable so the flag must be absent (false).
func TestCatalogEURetailConsistent(t *testing.T) {
	for _, a := range Catalog() {
		if a.UCITS && !a.EURetail {
			t.Errorf("%s: ucits implies eu_retail", a.ID)
		}
		if strings.HasPrefix(a.ISIN, "US") && a.EURetail {
			t.Errorf("%s: US-listed instrument flagged eu_retail", a.ID)
		}
		if a.Source == "index" && a.EURetail {
			t.Errorf("%s: index series cannot be eu_retail", a.ID)
		}
	}
}
