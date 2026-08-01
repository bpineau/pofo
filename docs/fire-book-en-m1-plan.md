# FIRE book English edition, M1 (technical preparation) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** All the machinery of the English edition (per-language Edition
value, sync stamps + drift report, figure-translation pass, EN guard tests,
two pilot translations), with ZERO change to what the French edition renders
and NO public English surface yet.

**Architecture:** Refactor `pkg/firebook` around a per-language `Edition`
value; the existing exported API becomes thin French wrappers. English
articles live in `assets/book/en/` with translated slugs paired to their
French source by an `Article.Source` field and a content-hash stamp. Figures
stay single-source in French; the English edition translates the rendered
SVG text nodes through a dictionary. Full rationale:
`docs/fire-book-en-edition-design.md` (the spec; read it first). This file
is transient: delete it once M1 ships.

**Tech Stack:** Go stdlib only (house rule, no exceptions). Repo conventions
in `CLAUDE.md` and `pkg/firebook/doc.go`.

## Global Constraints

- Stdlib only; no third-party dependency.
- Never write an em-dash, anywhere (prose, code, commits); a guard test
  already enforces it for articles.
- `make check` (fmt-check + vet + staticcheck + tests) must pass at every
  commit; commit and push directly to `master`, one commit per task.
- The French rendering must be BYTE-IDENTICAL before and after the refactor
  (Task 1 builds the harness that proves it).
- `/firebook/en/` is NOT mounted in M1: no changes to `cmd/pofo/serve.go`,
  `cmd/pofo/hub.go`, `cmd/pofo/adapt.go` or `pkg/decumul/web/server.go`.
- English prose: US English, point decimals ("6.6%", "1,000,000"), no space
  before "%"; keep euro amounts in euros. Style sheet: spec section
  "English style sheet".
- After the Task 3 refactor, run `cd ../finador && go test ./...` (finador
  imports `firebook.Handler`/`EPUB` via a replace directive); it must pass
  untouched.

---

### Task 1: Frozen-French-rendering harness (temporary)

A characterization test proving the refactor never changes French output.
It hashes every article page body, the index page and the EPUB bytes, and
compares against a recorded baseline. It lives in the repo during M1 and is
DELETED in Task 10.

**Files:**
- Create: `pkg/firebook/frozen_test.go`

**Interfaces:**
- Consumes: `Categories`, `articleHTML`, `indexHTML`, `EPUB` (current
  package-level forms; Task 3 keeps these working as wrappers, which is
  exactly what this test checks).

- [x] **Step 1: Write the test**

```go
package firebook

// TEMPORARY characterization harness for the Edition refactor (M1 of
// docs/fire-book-en-edition-design.md): records a digest of everything the
// French edition renders, then fails if any later change moves it.
// Deleted at the end of M1.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestFrozenFrenchRendering(t *testing.T) {
	path := os.Getenv("FIREBOOK_FROZEN")
	if path == "" {
		t.Skip("set FIREBOOK_FROZEN=<file> to record or compare the French rendering digest")
	}
	h := sha256.New()
	_, _ = io.WriteString(h, indexHTML(0))
	for _, cat := range Categories {
		for _, a := range cat.Articles {
			_, _ = io.WriteString(h, articleHTML(a, cat))
		}
	}
	epubBytes, err := EPUB(time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	h.Write(epubBytes)
	sum := hex.EncodeToString(h.Sum(nil))
	prev, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(path, []byte(sum), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("recorded baseline %s", sum)
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if string(prev) != sum {
		t.Errorf("French rendering changed: digest %s, baseline %s", sum, string(prev))
	}
}
```

- [x] **Step 2: Record the baseline and verify stability**

Run twice:
`FIREBOOK_FROZEN=/tmp/firebook-frozen go test ./pkg/firebook -run TestFrozenFrenchRendering -v`
Expected: first run logs "recorded baseline", second run PASSes (digest
stable). Keep `/tmp/firebook-frozen` for the whole M1 session; if it is
lost, re-record from a clean `git stash` state.

- [x] **Step 3: Full check and commit**

```bash
make check
git add pkg/firebook/frozen_test.go && git commit -m "firebook: temporary frozen-rendering harness for the Edition refactor" && git push
```

---

### Task 2: bookmd callout-label override

**Files:**
- Modify: `pkg/bookmd/render.go` (Options struct, the callout branch around
  line 175), `pkg/bookmd/doc.go`
- Test: `pkg/bookmd/render_test.go`

**Interfaces:**
- Produces: `Options.Callouts map[string]Callout` (nil keeps the built-in
  French `Callouts`); unknown types still degrade to the "encart" entry of
  the ACTIVE map. Task 3 consumes this via `Edition.Callouts`.

- [x] **Step 1: Write the failing test**

```go
func TestCalloutLabelOverride(t *testing.T) {
	en := map[string]Callout{
		"encart": {"❖", "Aside"},
		"cle":    {"🔑", "The key idea"},
	}
	got := ToHTML("::: cle\nBody.\n:::", Options{Callouts: en})
	if !strings.Contains(got, "The key idea") {
		t.Errorf("override ignored: %s", got)
	}
	got = ToHTML("::: mystery\nBody.\n:::", Options{Callouts: en})
	if !strings.Contains(got, "Aside") {
		t.Errorf("unknown type must degrade to the override's encart: %s", got)
	}
	got = ToHTML("::: cle\nBody.\n:::", Options{})
	if !strings.Contains(got, "L'idée clé") {
		t.Errorf("nil Callouts must keep the built-in French labels: %s", got)
	}
}
```

- [x] **Step 2: Run it, verify it fails** (`go test ./pkg/bookmd -run TestCalloutLabelOverride`)

- [x] **Step 3: Implement**

Add to `Options`: `Callouts map[string]Callout // display labels; nil -> the built-in French Callouts`.
In `ToHTML`, resolve once at the top: `callouts := opt.Callouts; if callouts == nil { callouts = Callouts }`,
then use `callouts` where the code reads the package map (including the
"unknown degrades to encart" lookup). Update `doc.go`'s Options paragraph.

- [x] **Step 4: Tests pass, French frozen digest unchanged**

`go test ./pkg/bookmd` then the FIREBOOK_FROZEN comparison run from Task 1.

- [x] **Step 5: Commit** (`bookmd: per-render callout label override (Options.Callouts)`)

---

### Task 3: The Edition refactor (French only, behavior-preserving)

The heart of M1. Pure moves, no behavior change: every French literal in
`handler.go` / `epub.go` migrates into a `French` Edition value; the
exported API becomes wrappers.

**Files:**
- Create: `pkg/firebook/edition.go`
- Modify: `pkg/firebook/manifest.go` (Article gains `Source`),
  `pkg/firebook/handler.go`, `pkg/firebook/epub.go`,
  `pkg/firebook/render.go`, `pkg/firebook/doc.go`
- Test: `pkg/firebook/edition_test.go`

**Interfaces:**
- Produces (later tasks and finador rely on these exact forms):

```go
// edition.go
type UIStrings struct {
	IndexLink          string            // "Sommaire"
	SameCategory       string            // "Dans la même partie"
	LinkCopied         string            // "lien copié" (the bookJS toast)
	SectionAnchorLabel string            // aria-label of the § heading anchor
	PartAnchorLabel    string            // aria-label of the § part anchor
	EPUBLink           string            // "Version epub"
	EPUBUnavailable    string            // 500 body of the epub route
	CatalogUnavailable string            // 500 body of the opds route
	NotFound           string            // "Article introuvable."
	EditionNote        string            // title-page sentence; %s = HomePath
	HumanSize          func(int) string  // "1,4 Mo" vs "1.4 MB"
}

type Edition struct {
	Lang, OGLocale                       string
	SiteName, SiteDescription, SiteLede  string
	HomePath, AssetDir                   string // "/firebook/fr/", "assets/book/fr"
	EPUBFileName, EPUBIdentifier         string
	Categories                           []Category
	UI                                   UIStrings
	Callouts                             map[string]bookmd.Callout
	Figure                               func(id string) string
}

var French = &Edition{ /* every current constant, verbatim */ }

func (e *Edition) Handler(opts ...Option) http.Handler
func (e *Edition) EPUB(modified time.Time) ([]byte, error)
func (e *Edition) Titles() map[string]string
func (e *Edition) ToHTML(src string, titles map[string]string) string
```

- Package-level wrappers keep compiling for finador and cmd/pofo:
  `func Handler(opts ...Option) http.Handler { return French.Handler(opts...) }`,
  same one-liners for `EPUB`, `Titles`, `ToHTML`.
- `Article` gains `Source string` (slug of the French original; empty on
  every French article and on EN-only articles). The French manifest
  literals DO NOT change (the field zero-values).

- [x] **Step 1: Write the regression test**

```go
// edition_test.go
func TestFrenchEditionWiring(t *testing.T) {
	if French.Lang != "fr" || French.AssetDir != "assets/book/fr" {
		t.Fatalf("French edition misconfigured: %+v", French)
	}
	if French.EPUBFileName != "le-fire-tranquille.epub" {
		t.Errorf("epub filename moved: %s", French.EPUBFileName)
	}
	// The wrappers must be the French edition.
	a, b := Titles(), French.Titles()
	if len(a) == 0 || len(a) != len(b) {
		t.Errorf("Titles wrapper diverges: %d vs %d", len(a), len(b))
	}
	// A served article must carry the French chrome.
	srv := httptest.NewServer(French.Handler())
	defer srv.Close()
	res, err := http.Get(srv.URL + "/fire-cest-quoi")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	for _, want := range []string{`lang="fr"`, "Dans la même partie", "fr_FR"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("article page lost %q", want)
		}
	}
}
```

- [x] **Step 2: Run it, verify it fails** (French/Edition undefined)

- [x] **Step 3: Implement the moves**

Mechanics, in order; each is a move, not a rewrite:
1. `edition.go`: the two types, `French` populated from today's constants
   (`siteName`, `siteDescription`, `siteLede`, `epubFileName`,
   `epubIdentifier`, `bookHomePath`, `humanSize`, the aria labels and
   error strings currently inlined). Keep the old constants as the
   initializers' source, then delete the constants and inline their values
   into `French` so there is exactly one home.
2. `handler.go`: `Handler` becomes `(e *Edition) Handler`; `writePage`,
   `indexHTML`, `articleHTML`, `find`, `jsonLD` become methods (or take
   `e *Edition` as first parameter, whichever reads better); every
   hardcoded string, `lang="fr"`, `og:locale`, `inLanguage`,
   `assets/book/fr/` path and `humanSize` call goes through `e`.
3. `epub.go`: `EPUB` becomes `(e *Edition) EPUB`; language, identifier,
   title page, asset path via `e`.
4. `render.go`: `(e *Edition) ToHTML` passes `bookmd.Options{Titles:
   titles, Figure: e.Figure, Callouts: e.Callouts}`; `French.Figure =
   FigureSVG`, `French.Callouts = nil` (built-in French labels).
5. Package wrappers at the bottom of `edition.go`. `doc.go` gains an
   "Editions" paragraph.

- [x] **Step 4: Prove nothing moved**

```bash
make check
FIREBOOK_FROZEN=/tmp/firebook-frozen go test ./pkg/firebook -run TestFrozenFrenchRendering -v   # must PASS against the Task 1 baseline
cd ../finador && go test ./... && cd ../pofo
```

- [x] **Step 5: Commit** (`firebook: refactor around a per-language Edition value (French unchanged)`)

---

### Task 4: The English Edition value and slug plan

**Files:**
- Create: `pkg/firebook/manifest_en.go`
- Modify: `pkg/firebook/edition.go` (English var), `pkg/firebook/manifest_test.go` (plannedEN)

**Interfaces:**
- Produces: `var English = &Edition{...}` with `Lang: "en"`,
  `OGLocale: "en_US"`, `SiteName: "The Quiet FIRE"`,
  `EPUBFileName: "the-quiet-fire.epub"`, `HomePath: "/firebook/en/"`,
  `AssetDir: "assets/book/en"`, a freshly minted `urn:uuid:` (generate once
  with `uuidgen`, lowercase, hardcode; it must differ from the French one
  and never change), `CategoriesEN` (initially: every category empty of
  articles except the two pilots added in Task 8), English `UIStrings`, the
  English callout map from Task 2's labels ("Aside", "The key idea",
  "Pro tip", "Watch out", "Worked example", "What the research says",
  "From the field"; same glyphs), `Figure: FigureSVGEnglish` (Task 6).
- Produces: `plannedEN`, the full translated slug list (test data, like
  `planned`), and with it the de facto FR->EN slug map used by every later
  translation.

- [x] **Step 1: Draft the complete EN slug list**

Translate all ~90 French slugs (short, kebab-case, keyword-bearing US
English; proper nouns stay: `vpw`, `guyton-klinger`,
`anarkulova-cederburg`). Examples setting the tone:
`fire-cest-quoi` -> `what-is-fire`, `combien-il-vous-faut` ->
`how-much-you-need`, `sequence-des-rendements` -> `sequence-of-returns`,
`etude-trinity` -> `the-trinity-study`, `cash-ameliore` -> `enhanced-cash`,
`echelle-obligataire` -> `bond-ladders`, `utiliser-la-page-fire` ->
`using-the-fire-simulator`. The French tax part is NOT mapped; in its slot,
the three US articles from the spec: `us-accounts-and-account-order`,
`us-taxes-in-the-withdrawal-phase`, `us-healthcare-and-social-security`.
Record the list as `plannedEN` (a `[]string` next to `planned` in the test
file) plus a comment table `EN slug <- FR slug` so the translator finds
targets without guessing.

- [x] **Step 2: Write English SiteDescription and SiteLede**

One dense sentence each, mirroring the French intent (see `handler.go`'s
current French values); mark both with a `// draft, review at M4` comment.
`UIStrings`: "Contents", "In the same part", "link copied",
"Link to this section" / "Link to this part", "EPUB edition",
"EPUB unavailable", "Catalog unavailable", "Article not found.",
"The online edition, kept current, is published by pofo at %s.",
and `HumanSize` producing "312 KB" / "1.4 MB" (point decimal).

- [x] **Step 3: Minimal wiring test**

```go
func TestEnglishEditionWiring(t *testing.T) {
	if English.EPUBIdentifier == French.EPUBIdentifier {
		t.Fatal("editions must not share a publication identifier")
	}
	if English.SiteName != "The Quiet FIRE" || English.Lang != "en" {
		t.Errorf("misconfigured: %+v", English)
	}
	if English.UI.HumanSize(1<<20+400<<10) != "1.4 MB" {
		t.Errorf("HumanSize: got %s", English.UI.HumanSize(1<<20+400<<10))
	}
}
```

- [x] **Step 4: `make check`, frozen digest still green, commit**
  (`firebook: English edition value, chrome strings and slug plan`)

---

### Task 5: Source stamps and the drift report

**Files:**
- Create: `pkg/firebook/drift.go`
- Test: `pkg/firebook/drift_test.go`
- Modify: `cmd/pofo/main.go` (flag), create `cmd/pofo/bookdrift.go`, `Makefile`

**Interfaces:**
- Produces:

```go
// drift.go
type DriftItem struct {
	ENSlug string // "" when the FR article has no EN counterpart yet
	FRSlug string
	Reason string // "stale" or "untranslated"
}

// Drift compares each English article's source stamp with the current
// French file and lists what the translation campaign owes. Wraps driftFS
// over the embedded assets.
func Drift() []DriftItem
// driftFS is the testable core: fsys holds "book/fr" and "book/en"
// (Drift passes fs.Sub(assets, "assets") and English.Categories).
func driftFS(fsys fs.FS, english []Category) []DriftItem
// sourceStamp extracts (frSlug, hash12) from an EN article body, or ok=false.
func sourceStamp(body []byte) (frSlug, hash string, ok bool)
```

- Stamp format, line 2 of every paired EN article (right after `# Title`):
  `<!-- source: <fr-slug> @ <hash12> -->` where hash12 = first 12 hex chars
  of sha256 over the FR file's exact bytes. Regexp:
  `^<!-- source: ([a-z0-9-]+) @ ([0-9a-f]{12}) -->$`.
- "untranslated" items: French manifest articles whose slug is neither any
  EN article's `Source` nor in the French tax part (hardcode the seven tax
  slugs in drift.go with a comment; they intentionally have no EN
  counterpart).
- CLI: `-book-drift` in `main.go` next to `-export-epub` (needs no
  portfolio; dispatch early like `-export-epub` does), printing one line
  per item and exiting 0; `runBookDrift` lives in `cmd/pofo/bookdrift.go`.
  Makefile: `book-drift: build` target running `./pofo -book-drift`.

- [ ] **Step 1: Write the failing tests** (use `fstest.MapFS`; no network,
  no disk)

```go
func TestDrift(t *testing.T) {
	fr := []byte("# Titre\n\nCorps.\n")
	sum := sha256.Sum256(fr)
	stamp := hex.EncodeToString(sum[:])[:12]
	fsys := fstest.MapFS{
		"book/fr/un.md":    {Data: fr},
		"book/fr/deux.md":  {Data: []byte("# Deux\n\nCorps.\n")},
		"book/en/one.md":   {Data: []byte("# One\n<!-- source: un @ " + stamp + " -->\n\nBody.\n")},
	}
	english := []Category{{Title: "Start", Articles: []Article{{Slug: "one", Title: "One", Source: "un"}}}}
	got := driftFS(fsys, english) // walks "book/fr", "book/en"
	// "deux" is untranslated; "one" is fresh.
	want := []DriftItem{{FRSlug: "deux", Reason: "untranslated"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
	// Now stale: FR content moves, stamp does not.
	fsys["book/fr/un.md"] = &fstest.MapFile{Data: []byte("# Titre\n\nCorps revu.\n")}
	got = driftFS(fsys, english)
	if len(got) != 2 || got[0].Reason != "stale" || got[0].ENSlug != "one" {
		t.Errorf("expected one stale + one untranslated, got %+v", got)
	}
}

func TestSourceStamp(t *testing.T) {
	fr, h, ok := sourceStamp([]byte("# T\n<!-- source: abc @ 0123456789ab -->\nBody"))
	if !ok || fr != "abc" || h != "0123456789ab" {
		t.Errorf("parse failed: %q %q %v", fr, h, ok)
	}
	if _, _, ok := sourceStamp([]byte("# T\n\nno stamp")); ok {
		t.Error("false positive")
	}
}
```

- [ ] **Step 2: Run, verify failure; implement; run, verify pass.**
  Sort the returned items (stale first, then untranslated, alphabetical
  within each) so the report and the tests are deterministic.

- [ ] **Step 3: Wire the CLI flag + Makefile target; smoke it**

`make build && ./pofo -book-drift` currently prints ~86 "untranslated"
lines (every non-tax French article; the two pilots shrink it in Task 8).

- [ ] **Step 4: `make check`, commit**
  (`firebook: source stamps and pofo -book-drift, the FR->EN sync report`)

---

### Task 6: Figure translation pass

**Files:**
- Create: `pkg/firebook/figures_i18n.go`
- Test: `pkg/firebook/figures_i18n_test.go`

**Interfaces:**
- Produces:

```go
// FigureSVGEnglish renders a plate and translates its text nodes to
// English: dictionary first (per-plate "<id>|<french>" key wins over the
// global "<french>" key), then mechanical French->English number
// reformatting for neutral payloads. English.Figure points here.
func FigureSVGEnglish(id string) string
// anglicizeNumbers rewrites French-formatted numbers in one text payload:
// decimal commas to points, (narrow) no-break-space thousands to commas,
// "6,6 %" to "6.6%".
func anglicizeNumbers(s string) string
// figureTextNodes extracts every <text>/<tspan> payload of a rendered
// plate (leaf payloads only; a <text> wrapping <tspan>s contributes the
// tspan payloads). Exported to the guard test.
func figureTextNodes(svg string) []string
var figureDict = map[string]string{ /* grows with the campaign */ }
// reFrenchDecimal (`\d,\d`) and reFrenchPercent
// (`\d[\x{00a0}\x{202f} ]%`) flag French-formatted residues; shared with
// the Task 7 guards.
```

- Implementation notes: RE2 has no lookahead, so `anglicizeNumbers` works
  in three ordered passes over the payload:
  1. decimal commas: `(\d),(\d)` -> `$1.$2` (in French figure text a comma
     between digits is ALWAYS a decimal; thousands use spaces);
  2. thousands: replace `(\d)[\x{00a0}\x{202f} ](\d\d\d)` -> `$1,$2`,
     looping until stable ("1 000 000" needs two passes);
  3. percent: `(\d)[\x{00a0}\x{202f} ]?%` -> `$1%`.
- Neutral payloads (eligible for pass-through + anglicize, no dictionary
  entry needed): match `^[\d\s\p{Zs}%×+.,:;()'’/\-−–]+$` AND contain at
  least one digit.
- Node extraction: two regexps, `<tspan[^>]*>([^<]*)</tspan>` and
  `<text[^>]*>([^<]*)</text>` (the second only matches childless <text>).
  Replacement uses the same shapes. Payloads are already-escaped SVG text
  (`&amp;` stays as-is in dictionary keys).

- [ ] **Step 1: Write the failing tests**

```go
func TestAnglicizeNumbers(t *testing.T) {
	cases := map[string]string{
		"6,6 %":       "6.6%",
		"1 000 000":   "1,000,000",
		"1 000":  "1,000",
		"1966-1995":   "1966-1995",
		"+4,1 % / an": "+4.1% / an", // words untouched: dictionary's job
	}
	for in, want := range cases {
		if got := anglicizeNumbers(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestFigureSVGEnglish(t *testing.T) {
	// vol-drag is small and stable; pin one dictionary entry end to end.
	fr, en := FigureSVG("vol-drag"), FigureSVGEnglish("vol-drag")
	if fr == en {
		t.Error("translation pass changed nothing")
	}
	for _, node := range figureTextNodes(en) {
		if reFrenchDecimal.MatchString(node) { // `\d,\d` after anglicize
			t.Errorf("French decimal survived: %q", node)
		}
	}
}
```

- [ ] **Step 2: Run, verify failure; implement; seed `figureDict` with the
  labels of `vol-drag` only** (render it, list its payloads, translate
  them). Run, verify pass.

- [ ] **Step 3: `make check`, frozen digest still green (the French path
  never calls the new code), commit**
  (`firebook: English figure pass (dictionary over SVG text nodes)`)

---

### Task 7: English guard tests

**Files:**
- Create: `pkg/firebook/manifest_en_test.go`

**Interfaces:**
- Consumes: `English`, `plannedEN`, `sourceStamp`, `figureTextNodes`,
  `figureDict`, the neutral regexp from Task 6.

- [ ] **Step 1: Write the guards** (all in one file; they mirror
  `manifest_test.go` and run green while the EN tree is small)

```go
// 1. Manifest <-> files, both ways, over assets/book/en (same shape as
//    TestManifestMatchesFiles, English.Titles() and plannedEN).
// 2. Per EN article: opens with "# Title"; no em-dash; every [[slug]]
//    resolves in plannedEN; French number patterns are rejected:
//    reFrenchPercent = `\d[\x{00a0}\x{202f} ]%` and
//    reFrenchDecimal applied to prose lines outside code spans is too
//    noisy; enforce only reFrenchPercent + any U+00A0/U+202F adjacent to a
//    digit. No minimum word count yet (pilots only).
// 3. Per paired EN article (Source != ""): the stamp exists, is
//    well-formed, and names an existing FR file; Source matches the
//    stamp's fr-slug; the SET of "::: figure" ids equals the FR
//    article's set (parity: a translation cannot drop or add a plate).
// 4. Figure dictionary coverage: for every figure id referenced from
//    assets/book/en, every payload of FigureSVG(id) is dictionary-covered
//    (global or id-scoped key) or neutral; and FigureSVGEnglish(id) leaves
//    no payload matching reFrenchDecimal or reFrenchPercent.
// 5. Completeness (env-gated until M4): unless os.Getenv("FIREBOOK_EN_COMPLETE") == "",
//    skip; when set, every French article outside the seven tax slugs must
//    be some EN article's Source.
```

Write them as real Go tests, reusing the loops of `manifest_test.go`
adjusted to `English.Categories` / `assets/book/en`.

- [ ] **Step 2: Run `go test ./pkg/firebook`**: green (EN tree is empty,
  loops are vacuous; guard 4 covers nothing yet).

- [ ] **Step 3: `make check`, commit** (`firebook: English edition guard tests`)

---

### Task 8: Two pilot translations

The end-to-end validation of stamps, guards, dictionary and style sheet.

**Files:**
- Create: `pkg/firebook/assets/book/en/sequence-of-returns.md`,
  `pkg/firebook/assets/book/en/vpw.md`
- Modify: `pkg/firebook/manifest_en.go` (the two Article entries, with
  Source), `pkg/firebook/figures_i18n.go` (dictionary entries for the
  plates those two articles use)

**Interfaces:**
- Consumes: the spec's "English style sheet" section (US English, no
  em-dash, glosses dropped, real idioms, EN link targets from the Task 4
  slug map, euro amounts kept, point decimals) and the Task 5 stamp format.

- [ ] **Step 1: Translate `sequence-des-rendements.md` ->
  `sequence-of-returns.md`** (figure-heavy pilot). Title `# The sequence of
  returns: the retiree's real enemy` (or better; manifest title must match
  the in-file h1). Stamp on line 2 with the current hash:
  `python3 -c "import hashlib;print(hashlib.sha256(open('pkg/firebook/assets/book/fr/sequence-des-rendements.md','rb').read()).hexdigest()[:12])"`.
  Wiki-links -> EN slugs (they may target unwritten plannedEN slugs, the
  guard allows it). Translate the `::: figure` captions; keep the ids.

- [ ] **Step 2: Fill the dictionary for that article's plates**, run
  `go test ./pkg/firebook -run 'English|Figures'`, iterate until the
  coverage guard passes.

- [ ] **Step 3: Same for `vpw.md`** (table-heavy pilot; slug unchanged).
  Tables: translated headers, KEEP THEM SHORT (EPUB rule), numbers in
  English format, euro amounts kept.

- [ ] **Step 4: Verify the whole loop**

```bash
make check
./pofo -book-drift        # the two pilots gone from "untranslated"; nothing "stale"
# then edit one word in the FR sequence article, rebuild, re-run:
./pofo -book-drift        # sequence-of-returns now "stale"; revert the FR edit
```

- [ ] **Step 5: Read both articles line by line** against the style sheet
  (mechanical checks do not catch bad English; this is the review the
  French campaigns proved necessary), fix, then commit one commit per
  article (`firebook: pilot EN translation of <slug>`).

---

### Task 9: figure-audit.sh (label overflow, per edition)

**Files:**
- Create: `scripts/figure-audit.sh`

**Interfaces:**
- Usage: `scripts/figure-audit.sh [fr|en]` (default fr). Needs Chrome;
  NOT part of `make check`.

- [ ] **Step 1: Write the script.** Shape (mirror `scripts/report-shot.sh`
  for the Chrome invocation): a small `go run` helper inlined in the script
  (heredoc under the scratch dir) renders every plate of the edition
  (`FigureSVG` or `FigureSVGEnglish` over `FigureIDs()`) into one HTML
  harness that inlines `webui.FontsCSS` + `webui.CSS` + the book's figure
  CSS; then headless Chrome (`--headless --dump-dom` with a page script)
  walks every `<text>` node comparing `getComputedTextLength() + x` (or
  anchor-adjusted extent) against the plate's viewBox width minus its
  padding, and prints one `PLATE id: "label" overflows by Npx` line per
  offender; exit 1 if any.

- [ ] **Step 2: Run `scripts/figure-audit.sh fr`**: expect zero offenders
  (the French plates were hand-checked; if a marginal one trips, widen the
  tolerance to match reality and say so in the commit message).
  Run `scripts/figure-audit.sh en`: fix any overflow caused by the pilot
  dictionary entries (shorten the English label, not the plate).

- [ ] **Step 3: Commit** (`scripts: figure-audit.sh, per-edition plate label overflow check`)

---

### Task 10: Cleanup and documentation

**Files:**
- Delete: `pkg/firebook/frozen_test.go`
- Modify: `pkg/firebook/doc.go`, `CLAUDE.md` (pkg/firebook row: mention
  editions + `-book-drift`), `docs/fire-book-en-edition-design.md` (status
  line: M1 done), `docs/README.md`

**Steps:**

- [ ] **Step 1: Final frozen comparison, then delete the harness**

`FIREBOOK_FROZEN=/tmp/firebook-frozen go test ./pkg/firebook -run TestFrozenFrenchRendering -v`
must PASS one last time; then `git rm pkg/firebook/frozen_test.go`.

- [ ] **Step 2: Documentation pass** (doc.go Editions section with a
  three-line usage example; CLAUDE.md row; spec status line "M1 done
  <date>, M2 translation campaign next"; drop this plan's line from
  docs/README.md and `git rm docs/fire-book-en-m1-plan.md`).

- [ ] **Step 3: Final gate**

```bash
make check && make golden
cd ../finador && go test ./... && cd ../pofo
./pofo -book-drift | head
```

- [ ] **Step 4: Commit** (`firebook: close M1 of the English edition (docs, harness removal)`)

---

## Self-review notes (already applied)

- The EN handler stays unmounted; `English.Handler()` is still exercised
  in-process by Task 4/7 tests via httptest, so M1 proves it works without
  exposing it.
- `WithAlternate`, `-book-lang` on `-export-epub`, hreflang, nav links:
  M4, deliberately absent here.
- The frozen harness hashes `articleHTML`/`indexHTML`/`EPUB` BEFORE the
  refactor renames them into methods; Task 3 keeps package-level wrappers,
  so the harness compiles unchanged throughout. If a signature must move
  anyway, update the harness in the same commit and re-record from the
  pre-refactor commit to keep the guarantee honest.
