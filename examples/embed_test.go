package examples

import (
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	infos := List()
	if len(infos) < 10 {
		t.Fatalf("List() = %d entries, want at least 10", len(infos))
	}
	byName := map[string]Info{}
	for i, in := range infos {
		if in.Title == "" {
			t.Errorf("%s: empty Title", in.Name)
		}
		if strings.ContainsAny(in.Name, `/\.`) {
			t.Errorf("%s: Name must be a bare base name", in.Name)
		}
		if i > 0 && infos[i-1].Name >= in.Name {
			t.Errorf("List() not sorted at %s", in.Name)
		}
		byName[in.Name] = in
	}
	h, ok := byName["dragon-decumulation-household"]
	if !ok {
		t.Fatal("dragon-decumulation-household missing")
	}
	if !strings.HasPrefix(h.Title, "Dragon-decumulation") {
		t.Errorf("Title = %q, want the file's first comment line", h.Title)
	}
	if h.Blurb == "" {
		t.Error("Blurb empty, want the part after the -- separator")
	}
	if _, err := FS.ReadFile(h.Name + ".txt"); err != nil {
		t.Errorf("FS.ReadFile: %v", err)
	}
}
