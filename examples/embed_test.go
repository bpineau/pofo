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

	// Every bundled file now opens with a real "# Title -- blurb" line, so
	// the UI never has to fall back to the bare file id. ntsg.txt keeps its
	// "#meta sim:on" directive below a prose title line: the title must be
	// surfaced, not the directive.
	if ntsg, ok := byName["ntsg"]; !ok {
		t.Fatal("ntsg missing")
	} else {
		if ntsg.Title != "NTSG" {
			t.Errorf("ntsg.Title = %q, want %q", ntsg.Title, "NTSG")
		}
		if ntsg.Blurb == "" {
			t.Error("ntsg.Blurb empty, want the part after the -- separator")
		}
	}

	// predictis.txt used to open on a raw holdings line; it now carries a
	// prose title line above the holdings, which must be surfaced.
	if predictis, ok := byName["predictis"]; !ok {
		t.Fatal("predictis missing")
	} else {
		if predictis.Title != "Predictis" {
			t.Errorf("predictis.Title = %q, want %q", predictis.Title, "Predictis")
		}
		if predictis.Blurb == "" {
			t.Error("predictis.Blurb empty, want the part after the -- separator")
		}
	}
}
