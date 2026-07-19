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

	// ntsg.txt starts with the "#meta sim:on" directive, not a title line:
	// it must degrade to Title = Name rather than surface the directive.
	if ntsg, ok := byName["ntsg"]; !ok {
		t.Fatal("ntsg missing")
	} else {
		if ntsg.Title != "ntsg" {
			t.Errorf("ntsg.Title = %q, want %q (degraded from directive)", ntsg.Title, "ntsg")
		}
		if ntsg.Blurb != "" {
			t.Errorf("ntsg.Blurb = %q, want empty", ntsg.Blurb)
		}
	}

	// predictis.txt starts with a raw holdings line (no leading "#" at
	// all): it must degrade to Title = Name rather than surface the line.
	if predictis, ok := byName["predictis"]; !ok {
		t.Fatal("predictis missing")
	} else {
		if predictis.Title != "predictis" {
			t.Errorf("predictis.Title = %q, want %q (degraded, no comment line)", predictis.Title, "predictis")
		}
		if predictis.Blurb != "" {
			t.Errorf("predictis.Blurb = %q, want empty", predictis.Blurb)
		}
	}
}
