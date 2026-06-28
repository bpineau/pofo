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

func TestLookup(t *testing.T) {
	all := Catalog()
	if len(all) == 0 {
		t.Fatal("empty catalog")
	}
	want := all[0]

	got, ok := Lookup(want.ID)
	if !ok || got.ID != want.ID {
		t.Fatalf("Lookup by id: got %q (%v), want %q", got.ID, ok, want.ID)
	}
	if _, ok := Lookup("__definitely_not_an_asset__"); ok {
		t.Error("Lookup of an unknown id returned ok")
	}
	if want.ISIN != "" {
		if g, ok := Lookup(want.ISIN); !ok || g.ID != want.ID {
			t.Errorf("Lookup by ISIN failed for %q", want.ISIN)
		}
		// Case-insensitive.
		if g, ok := Lookup(strings.ToLower(want.ISIN)); !ok || g.ID != want.ID {
			t.Errorf("Lookup is not case-insensitive for %q", want.ISIN)
		}
	}
}
