package firebook

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"regexp"
	"sort"
	"strings"
)

// The synchronization contract between the editions.
//
// The French edition is the source of truth: editing a French article never
// breaks the build, but the staleness it creates in a translation must be
// visible on demand and impossible to lose. Every translated article therefore
// carries a SOURCE STAMP on the line right after its "# Title":
//
//	<!-- source: sequence-des-rendements @ 3f1c8a02be74 -->
//
// naming the French article it was made from and the first 12 hex digits of
// the sha256 of that file's exact bytes at translation time. Drift compares
// each stamp with the French file as it stands now: a differing hash means the
// original moved since, so the translation is stale. Its output is the
// translation worklist, printed by "pofo -book-drift" (make book-drift).
//
// A hash beats tracking the pair in git: it works on the embedded FS, needs no
// repository at test time, and survives history rewrites.
//
// The other half of the contract runs the other way. Some French articles
// belong to the French edition alone (the tax part is French law end to end,
// and the English edition writes its own US-framework part in that slot).
// Such an article says so in its own file, on the line right after its
// "# Title":
//
//	<!-- edition: fr-only: French law end to end -->
//
// The reason after the second colon is optional and purely documentation.
// Drift reports a marked article as "fr-only" and never as "untranslated", so
// a translation agent reading the worklist can tell "not yet done" from "never
// to be done" without consulting a table elsewhere. The marker lives in the
// file for the same reason the stamp does: it travels with the article.

// DriftItem is one entry of the synchronization report: a translated article
// whose French original moved since ("stale"), a French article no
// translation covers yet ("untranslated"), or a French article marked as
// belonging to the French edition alone ("fr-only").
type DriftItem struct {
	// ENSlug is the English article: the existing one for "stale", the
	// planned one for "untranslated" (empty when the French article is in no
	// plan), and always empty for "fr-only".
	ENSlug string
	FRSlug string
	Reason string // "stale", "untranslated" or "fr-only"
}

// reSourceStamp matches a source stamp on a line of its own.
var reSourceStamp = regexp.MustCompile(`(?m)^<!-- source: ([a-z0-9-]+) @ ([0-9a-f]{12}) -->$`)

// reEditionMark matches an edition marker on a line of its own, with or
// without its free-text reason.
var reEditionMark = regexp.MustCompile(`(?m)^<!-- edition: fr-only(?:: [^>]*)? -->$`)

// sourceStamp extracts the French slug and the recorded hash from a
// translated article's body, or reports ok=false when it carries no
// well-formed stamp.
func sourceStamp(body []byte) (frSlug, hash string, ok bool) {
	m := reSourceStamp.FindSubmatch(body)
	if m == nil {
		return "", "", false
	}
	return string(m[1]), string(m[2]), true
}

// frOnly reports whether an article carries the fr-only edition marker, which
// makes it French-edition-only and never owed to a translation.
func frOnly(body []byte) bool {
	return reEditionMark.Match(body)
}

// articleBody prepares one article's markdown for rendering: it drops the
// in-file "# Title" line (the page shell renders the h1), the source stamp and
// the edition marker, which are metadata for Drift and never content. Most
// articles carry neither, so for them this is exactly the old title-stripping
// behavior.
func articleBody(raw []byte) string {
	body := strings.TrimSpace(string(raw))
	if strings.HasPrefix(body, "# ") {
		_, rest, found := strings.Cut(body, "\n")
		if !found {
			return ""
		}
		body = rest
	}
	body = reSourceStamp.ReplaceAllString(body, "")
	return strings.TrimSpace(reEditionMark.ReplaceAllString(body, ""))
}

// SourceHash is the stamp value of a French article's bytes: the first 12 hex
// digits of their sha256. A translator writes it into the stamp; Drift
// recomputes it to detect the French article moving on.
func SourceHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])[:12]
}

// Drift reports what the English translation owes: the translated articles
// whose French original changed since they were made ("stale"), then the
// French articles no translation covers yet ("untranslated"), then the French
// articles marked fr-only, which are owed to nobody and are listed only so the
// worklist accounts for every French file ("fr-only").
//
// The report is advisory and nothing enforces immediacy: the workflow is to
// commit a French edit as usual, then run it and translate what it lists.
func Drift() []DriftItem {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil
	}
	return driftFS(sub, English.Categories)
}

// driftFS is the testable core of Drift: fsys holds the two article trees at
// "book/fr" and "book/en", english is the translated manifest.
func driftFS(fsys fs.FS, english []Category) []DriftItem {
	var stale, untranslated, french []DriftItem
	translated := make(map[string]bool)

	for _, cat := range english {
		for _, a := range cat.Articles {
			if a.Source == "" {
				continue // an article the edition writes for itself
			}
			translated[a.Source] = true
			body, err := fs.ReadFile(fsys, "book/en/"+a.Slug+".md")
			if err != nil {
				stale = append(stale, DriftItem{ENSlug: a.Slug, FRSlug: a.Source, Reason: "stale"})
				continue
			}
			_, hash, ok := sourceStamp(body)
			source, srcErr := fs.ReadFile(fsys, "book/fr/"+a.Source+".md")
			// No readable stamp, or a French original that moved since (or
			// vanished): either way the translation cannot be called current.
			if !ok || srcErr != nil || hash != SourceHash(source) {
				stale = append(stale, DriftItem{ENSlug: a.Slug, FRSlug: a.Source, Reason: "stale"})
			}
		}
	}

	entries, _ := fs.ReadDir(fsys, "book/fr")
	for _, e := range entries {
		slug, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok {
			continue
		}
		// The marker wins over everything: a French-only article is never
		// owed, translated or not.
		if body, err := fs.ReadFile(fsys, "book/fr/"+e.Name()); err == nil && frOnly(body) {
			french = append(french, DriftItem{FRSlug: slug, Reason: "fr-only"})
			continue
		}
		if translated[slug] {
			continue
		}
		untranslated = append(untranslated, DriftItem{ENSlug: plannedENSource[slug], FRSlug: slug, Reason: "untranslated"})
	}

	// Stale first (a moved original is the urgent half), then what is owed,
	// then what never will be; alphabetical within each section, so the report
	// and its tests are deterministic.
	byFRSlug := func(s []DriftItem) { sort.Slice(s, func(i, j int) bool { return s[i].FRSlug < s[j].FRSlug }) }
	byFRSlug(stale)
	byFRSlug(untranslated)
	byFRSlug(french)
	return append(append(stale, untranslated...), french...)
}
