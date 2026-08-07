package firebook

import (
	"regexp"
	"strings"
	"testing"
)

// The guard: files, manifest and wiki-links must stay consistent as the book
// grows (the full plan lives in docs/fire-book-design.md, mirrored by
// planned).

func plannedSet(t *testing.T) map[string]bool {
	t.Helper()
	set := make(map[string]bool, len(planned))
	for _, slug := range planned {
		if set[slug] {
			t.Errorf("planned lists %q twice", slug)
		}
		set[slug] = true
	}
	return set
}

func TestManifestMatchesFiles(t *testing.T) {
	files, err := assets.ReadDir("assets/book/fr")
	if err != nil {
		t.Fatal(err)
	}
	onDisk := make(map[string]bool)
	for _, f := range files {
		slug, ok := strings.CutSuffix(f.Name(), ".md")
		if !ok {
			t.Errorf("assets/book/fr/%s: not a .md file", f.Name())
			continue
		}
		onDisk[slug] = true
	}
	inManifest := Titles()
	for slug := range inManifest {
		if !onDisk[slug] {
			t.Errorf("manifest lists %q but assets/book/fr/%s.md does not exist", slug, slug)
		}
	}
	for slug := range onDisk {
		if _, ok := inManifest[slug]; !ok {
			t.Errorf("assets/book/fr/%s.md exists but is not in the manifest", slug)
		}
	}
	set := plannedSet(t)
	for slug := range inManifest {
		if !set[slug] {
			t.Errorf("manifest lists %q but it is missing from planned", slug)
		}
	}
}

var reWikiRef = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)

func TestArticles(t *testing.T) {
	set := plannedSet(t)
	for _, cat := range Categories {
		for _, a := range cat.Articles {
			raw, err := assets.ReadFile("assets/book/fr/" + a.Slug + ".md")
			if err != nil {
				t.Errorf("%s: %v", a.Slug, err)
				continue
			}
			body := string(raw)
			if !strings.HasPrefix(body, "# ") {
				t.Errorf("%s: must open with a '# Title' line", a.Slug)
			}
			if strings.Contains(body, "—") {
				t.Errorf("%s: contains an em-dash", a.Slug)
			}
			if n := len(strings.Fields(body)); n < 1200 {
				t.Errorf("%s: only %d words; book articles are long (>= 1200)", a.Slug, n)
			}
			for _, m := range reWikiRef.FindAllStringSubmatch(body, -1) {
				if slug := strings.TrimSpace(m[1]); !set[slug] {
					t.Errorf("%s: wiki-link [[%s]] targets no planned article (typo?)", a.Slug, slug)
				}
			}
		}
	}
}

func TestFiguresResolve(t *testing.T) {
	known := map[string]bool{}
	for _, id := range FigureIDs() {
		known[id] = true
	}
	used := map[string]bool{}
	re := regexp.MustCompile(`(?m)^::: figure (\S+)`)
	files, _ := assets.ReadDir("assets/book/fr")
	for _, f := range files {
		b, _ := assets.ReadFile("assets/book/fr/" + f.Name())
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			used[m[1]] = true
			if !known[m[1]] {
				t.Errorf("%s: ::: figure %q has no SVG in figures.go", f.Name(), m[1])
			}
		}
	}
	for id := range known {
		if !used[id] {
			t.Logf("figure %q defined but not used in any article", id)
		}
	}
}
