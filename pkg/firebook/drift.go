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

// DriftItem is one entry of the synchronization report: either a translated
// article whose French original moved since, or a French article no
// translation covers yet.
type DriftItem struct {
	ENSlug string // "" when the French article has no counterpart yet
	FRSlug string
	Reason string // "stale" or "untranslated"
}

// reSourceStamp matches a source stamp on a line of its own.
var reSourceStamp = regexp.MustCompile(`(?m)^<!-- source: ([a-z0-9-]+) @ ([0-9a-f]{12}) -->$`)

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

// articleBody prepares one article's markdown for rendering: it drops the
// in-file "# Title" line (the page shell renders the h1) and the source stamp,
// which is metadata for Drift and never content. French articles carry no
// stamp, so for them this is exactly the old title-stripping behavior.
func articleBody(raw []byte) string {
	body := strings.TrimSpace(string(raw))
	if strings.HasPrefix(body, "# ") {
		_, rest, found := strings.Cut(body, "\n")
		if !found {
			return ""
		}
		body = rest
	}
	return strings.TrimSpace(reSourceStamp.ReplaceAllString(body, ""))
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
// French articles no translation covers yet ("untranslated"). The French tax
// part is never owed (see taxOnlyFR).
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
	var stale, untranslated []DriftItem
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

	entries, err := fs.ReadDir(fsys, "book/fr")
	if err != nil {
		return append(stale, untranslated...)
	}
	for _, e := range entries {
		slug, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok || taxOnlyFR[slug] || translated[slug] {
			continue
		}
		untranslated = append(untranslated, DriftItem{FRSlug: slug, Reason: "untranslated"})
	}

	// Stale first (a moved original is the urgent half), alphabetical within
	// each half, so the report and its tests are deterministic.
	sort.Slice(stale, func(i, j int) bool { return stale[i].FRSlug < stale[j].FRSlug })
	sort.Slice(untranslated, func(i, j int) bool { return untranslated[i].FRSlug < untranslated[j].FRSlug })
	return append(stale, untranslated...)
}
