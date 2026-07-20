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
		if i > 0 && strings.ToLower(infos[i-1].Title) > strings.ToLower(in.Title) {
			t.Errorf("List() not sorted by Title at %s (%q after %q)",
				in.Name, in.Title, infos[i-1].Title)
		}
		byName[in.Name] = in
	}
	h, ok := byName["dragon-decumulation-household"]
	if !ok {
		t.Fatal("dragon-decumulation-household missing")
	}
	if !strings.HasPrefix(h.Title, "Dragon decumulation") {
		t.Errorf("Title = %q, want the file's first comment line", h.Title)
	}
	if h.Blurb == "" {
		t.Error("Blurb empty, want the part after the -- separator")
	}
	if _, err := FS.ReadFile(h.Name + ".txt"); err != nil {
		t.Errorf("FS.ReadFile: %v", err)
	}

	// Every bundled file now opens with a real "# Title -- blurb" line, so
	// the UI never has to fall back to the bare file id: the title must be
	// surfaced, and the blurb after the separator kept.
	if cof, ok := byName["coffeehouse-schultheis"]; !ok {
		t.Fatal("coffeehouse-schultheis missing")
	} else {
		if cof.Title != "The Coffeehouse Portfolio" {
			t.Errorf("coffeehouse-schultheis.Title = %q, want %q",
				cof.Title, "The Coffeehouse Portfolio")
		}
		if cof.Blurb == "" {
			t.Error("coffeehouse-schultheis.Blurb empty, want the part after the -- separator")
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

// TestTitleFromLine covers the pure first-line parser directly, including the
// degrade paths (directive lines, holding lines) that no bundled fixture
// currently exercises now that every file carries a real title.
func TestTitleFromLine(t *testing.T) {
	cases := []struct {
		name  string
		line  string
		title string
		blurb string
	}{
		{"title and blurb", "# The Coffeehouse Portfolio -- Bill Schultheis, UCITS",
			"The Coffeehouse Portfolio", "Bill Schultheis, UCITS"},
		{"title only", "# MSCI World", "MSCI World", ""},
		{"title only trailing dot", "# Just a name.", "Just a name", ""},
		{"meta directive", "#meta sim:on", "", ""},
		{"meta bare", "#meta", "", ""},
		{"holding line", "IWDA 60", "", ""},
		{"blank line", "", "", ""},
		// titleOf tolerates the em-dash separator via its escape.
		{"em-dash separator", "# Title \u2014 the blurb", "Title", "the blurb"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			title, blurb := titleFromLine(c.line)
			if title != c.title || blurb != c.blurb {
				t.Errorf("titleFromLine(%q) = (%q, %q), want (%q, %q)",
					c.line, title, blurb, c.title, c.blurb)
			}
		})
	}
}
