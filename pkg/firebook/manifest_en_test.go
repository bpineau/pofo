package firebook

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The English edition's guards, mirroring manifest_test.go: files, manifest,
// wiki-links, figures and source stamps must stay consistent as the
// translation campaign fills assets/book/en. They run green while the tree is
// small, and every loop is vacuous on an empty tree.

func plannedENSet(t *testing.T) map[string]bool {
	t.Helper()
	set := make(map[string]bool, len(plannedEN))
	for _, slug := range plannedEN {
		set[slug] = true
	}
	return set
}

// enArticleFiles lists the English article files, tolerating an absent
// directory (the tree does not exist until the first translation lands).
func enArticleFiles(t *testing.T) map[string]bool {
	t.Helper()
	onDisk := make(map[string]bool)
	files, err := assets.ReadDir(English.AssetDir)
	if err != nil {
		return onDisk
	}
	for _, f := range files {
		slug, ok := strings.CutSuffix(f.Name(), ".md")
		if !ok {
			t.Errorf("%s/%s: not a .md file", English.AssetDir, f.Name())
			continue
		}
		onDisk[slug] = true
	}
	return onDisk
}

func TestManifestENMatchesFiles(t *testing.T) {
	onDisk := enArticleFiles(t)
	inManifest := English.Titles()
	for slug := range inManifest {
		if !onDisk[slug] {
			t.Errorf("EN manifest lists %q but %s/%s.md does not exist", slug, English.AssetDir, slug)
		}
	}
	for slug := range onDisk {
		if _, ok := inManifest[slug]; !ok {
			t.Errorf("%s/%s.md exists but is not in the EN manifest", English.AssetDir, slug)
		}
	}
	set := plannedENSet(t)
	for slug := range inManifest {
		if !set[slug] {
			t.Errorf("EN manifest lists %q but it is missing from plannedEN", slug)
		}
	}
}

// reNBSPDigit catches a French thin/no-break space next to a digit, the
// signature of French number formatting surviving a translation. A bare
// reFrenchDecimal sweep over prose would be too noisy (English prose uses
// "1,000"), so the guard is the percent spacing plus these spaces.
var reNBSPDigit = regexp.MustCompile(`[\x{00a0}\x{202f}]\d|\d[\x{00a0}\x{202f}]`)

func TestArticlesEN(t *testing.T) {
	set := plannedENSet(t)
	for _, cat := range English.Categories {
		for _, a := range cat.Articles {
			raw, err := assets.ReadFile(English.AssetDir + "/" + a.Slug + ".md")
			if err != nil {
				t.Errorf("%s: %v", a.Slug, err)
				continue
			}
			body := string(raw)
			if !strings.HasPrefix(body, "# ") {
				t.Errorf("%s: must open with a '# Title' line", a.Slug)
			}
			if title, _, _ := strings.Cut(strings.TrimPrefix(body, "# "), "\n"); strings.TrimSpace(title) != a.Title {
				t.Errorf("%s: in-file title %q does not match the manifest title %q",
					a.Slug, strings.TrimSpace(title), a.Title)
			}
			if strings.Contains(body, "—") {
				t.Errorf("%s: contains an em-dash", a.Slug)
			}
			if m := reFrenchPercent.FindString(body); m != "" {
				t.Errorf("%s: French percent spacing %q; English writes \"4%%\"", a.Slug, m)
			}
			if m := reNBSPDigit.FindString(body); m != "" {
				t.Errorf("%s: French thin/no-break space around a digit (%q)", a.Slug, m)
			}
			for _, m := range reWikiRef.FindAllStringSubmatch(body, -1) {
				if slug := strings.TrimSpace(m[1]); !set[slug] {
					t.Errorf("%s: wiki-link [[%s]] targets no planned EN article (typo?)", a.Slug, slug)
				}
			}
		}
	}
}

var reFigureBlock = regexp.MustCompile(`(?m)^::: figure (\S+)`)

func figureIDsOf(t *testing.T, path string) map[string]bool {
	t.Helper()
	ids := map[string]bool{}
	raw, err := assets.ReadFile(path)
	if err != nil {
		t.Errorf("%s: %v", path, err)
		return ids
	}
	for _, m := range reFigureBlock.FindAllStringSubmatch(string(raw), -1) {
		ids[m[1]] = true
	}
	return ids
}

// A translated article must carry a well-formed source stamp naming an
// existing French file, and must use exactly the plates its original uses: a
// translation never silently drops or adds a figure.
func TestPairedArticlesEN(t *testing.T) {
	sources := map[string]string{}
	for _, cat := range English.Categories {
		for _, a := range cat.Articles {
			if a.Source == "" {
				continue // an article the English edition writes for itself
			}
			if prev, dup := sources[a.Source]; dup {
				t.Errorf("%s and %s both claim %q as their French source", prev, a.Slug, a.Source)
			}
			sources[a.Source] = a.Slug
			raw, err := assets.ReadFile(English.AssetDir + "/" + a.Slug + ".md")
			if err != nil {
				continue // TestManifestENMatchesFiles already reported it
			}
			frSlug, _, ok := sourceStamp(raw)
			if !ok {
				t.Errorf("%s: no well-formed source stamp on line 2", a.Slug)
				continue
			}
			if frSlug != a.Source {
				t.Errorf("%s: stamp names %q but the manifest says Source %q", a.Slug, frSlug, a.Source)
			}
			frPath := French.AssetDir + "/" + frSlug + ".md"
			if _, err := assets.ReadFile(frPath); err != nil {
				t.Errorf("%s: stamp names %q, which is not a French article", a.Slug, frSlug)
				continue
			}
			en := figureIDsOf(t, English.AssetDir+"/"+a.Slug+".md")
			fr := figureIDsOf(t, frPath)
			for id := range fr {
				if !en[id] {
					t.Errorf("%s: drops figure %q, which %s uses", a.Slug, id, frSlug)
				}
			}
			for id := range en {
				if !fr[id] {
					t.Errorf("%s: adds figure %q, which %s does not use", a.Slug, id, frSlug)
				}
			}
		}
	}
}

// Every plate an English article shows must be fully covered: each of its
// French payloads is either in the dictionary or neutral, and nothing French
// survives the pass.
func TestFigureDictionaryCoversEN(t *testing.T) {
	used := map[string]bool{}
	for slug := range enArticleFiles(t) {
		for id := range figureIDsOf(t, English.AssetDir+"/"+slug+".md") {
			used[id] = true
		}
	}
	for id := range used {
		for _, payload := range figureTextNodes(FigureSVG(id)) {
			if payload == "" || isNeutralPayload(payload) {
				continue
			}
			if _, ok := figureDict[id+"|"+payload]; ok {
				continue
			}
			if _, ok := figureDict[payload]; !ok {
				t.Errorf("figure %q: payload %q is neither in figureDict nor neutral", id, payload)
			}
		}
		for _, payload := range figureTextNodes(FigureSVGEnglish(id)) {
			if hasFrenchDecimal(payload) {
				t.Errorf("figure %q: French decimal survives translation: %q", id, payload)
			}
			if reFrenchPercent.MatchString(payload) {
				t.Errorf("figure %q: French percent spacing survives translation: %q", id, payload)
			}
		}
	}
}

// Dictionary values land verbatim inside SVG text nodes, so they must be
// XML-safe: a bare "&" or a "<" from a careless entry would corrupt every
// page and EPUB showing the plate, without failing any other test.
func TestFigureDictValuesAreXMLSafe(t *testing.T) {
	reEntity := regexp.MustCompile(`&(amp|lt|gt|quot|apos|#\d+);`)
	for fr, en := range figureDict {
		if strings.Contains(en, "<") {
			t.Errorf("figureDict[%q]: value %q contains a raw '<'", fr, en)
		}
		if strings.ContainsRune(reEntity.ReplaceAllString(en, ""), '&') {
			t.Errorf("figureDict[%q]: value %q carries a bare '&'; write &amp;", fr, en)
		}
	}
}

// Completeness. Env-gated while the campaign runs; it becomes unconditional
// when the edition ships (M4 of docs/fire-book-en-edition-design.md).
func TestEnglishEditionIsComplete(t *testing.T) {
	if os.Getenv("FIREBOOK_EN_COMPLETE") == "" {
		t.Skip("set FIREBOOK_EN_COMPLETE=1 to require every French article to have an English counterpart")
	}
	covered := map[string]bool{}
	for _, cat := range English.Categories {
		for _, a := range cat.Articles {
			if a.Source != "" {
				covered[a.Source] = true
			}
		}
	}
	for slug := range French.Titles() {
		if !taxOnlyFR[slug] && !covered[slug] {
			t.Errorf("%s has no English counterpart", slug)
		}
	}
}

// The source stamp is metadata, never content: it must not reach the page.
func TestStampNeverRenders(t *testing.T) {
	srv := httptest.NewServer(English.Handler())
	defer srv.Close()
	for _, cat := range English.Categories {
		for _, a := range cat.Articles {
			res, err := http.Get(srv.URL + "/" + a.Slug)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if strings.Contains(string(body), "source:") {
				t.Errorf("%s: the source stamp leaks into the page", a.Slug)
			}
		}
	}
}
