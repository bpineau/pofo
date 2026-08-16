# The FIRE book, English edition: the translator's brief

Opened 2026-08-16. This is the one file a translation session needs to
read before touching an article. It is deliberately procedural; the reasons
behind each rule live in `docs/fire-book-en-edition-design.md`, the
vocabulary in `docs/fire-book-en-glossary.md`. Read all three, in that
order, then work.

The French edition ("Le FIRE tranquille", `pkg/firebook/assets/book/fr/`)
is the SOURCE OF TRUTH. The English edition ("The Quiet FIRE",
`pkg/firebook/assets/book/en/`) is a translation of it, article by article,
one session per article, one commit per article.

## 0. The unit of work

One French article -> one English article. Never batch several articles in
one session, and never touch a French file (the only exception is a French
typo you are certain of: fix it in a SEPARATE commit first, then translate
the fixed version, because the source stamp hashes the French bytes).

## 1. Pick the article

```sh
make book-drift
```

prints three sections:

- `stale  <fr-slug>  -> <en-slug>`: the French original moved since the
  English article was made. Refresh the translation and its stamp.
- `untranslated  <fr-slug>  -> <en-slug>`: no English article yet. Translate.
- `fr-only  <fr-slug>`: French-only, NEVER translated (see 2). Skip.

The `-> <en-slug>` on the right is the English slug you MUST use: it comes
from `plannedEN` in `pkg/firebook/manifest_en.go`, which is the FR -> EN
slug map of the whole edition (English URLs are English words; proper nouns
keep their slug: `vpw`, `guyton-klinger`, `anarkulova-cederburg`).

Stale first, then untranslated in reading order of `plannedEN` (part I,
then II, ...), unless told otherwise.

## 2. French-only articles: the marker

A French article whose second line is

```
<!-- edition: fr-only -->
```

(optionally `<!-- edition: fr-only: <reason> -->`) belongs to the French
edition only. It is never translated, never listed as owed by
`make book-drift`, and no English article may name it as `Source`. Today the
seven articles of the "Fiscalité et cadre français" part carry it; the
English edition writes its own "Taxes and the US framework" part in that
slot (three `us-*` articles, original writing, no `Source`, no stamp).

When the article you translate wiki-links to a fr-only article
(`[[taxe-puma]]`, `[[retraite-legale]]`, `[[enveloppes-francaises]]`,
`[[flat-tax-et-imposition]]`, ...): do NOT translate the link target.
Neutralize the sentence ("before your pensions start", "net of taxes") and,
only where the reader needs the concrete local answer, point at the
matching `us-*` article of `plannedEN`. Record it in the ledger (see 7).

## 3. Read the source, decide generalize vs adapt

Every general article may carry French passages outside the tax part
("La pratique française" sections, envelope names in examples, euro market
practice). Two treatments, and only two:

- GENERALIZE (the default): replace the France-specific passage with
  locale-neutral phrasing, plus at most one pointer sentence to the relevant
  `us-*` article. Cheap, safe, keeps the two articles structurally close.
- ADAPT (the escalation): where local practice IS the point of a section,
  rewrite that section for the US reader (US products, US venues, US
  practice), same slot, same length order of magnitude, facts checked
  against primary sources at writing time. The rest of the article stays a
  translation.

The design doc's ledger says which articles are already decided ADAPT
(`rentes-et-annuites`, `or-en-retrait`, `etf-ucits-europeens`) and lists
the proposals awaiting arbitration. If your article is in neither list,
GENERALIZE. If a passage clearly needs ADAPT and no decision exists, do the
generalize version, and say so in your report so the maintainer can decide.

Never keep an "in France, ..." passage verbatim as a curiosity: it is dead
weight for the target reader.

## 4. Translate

The style sheet, in full, is the "English style sheet" section of the
design doc; the vocabulary is `docs/fire-book-en-glossary.md`. The
non-negotiables:

- US English. NO em-dash, NO en-dash, anywhere (prose, tables, captions,
  commit message). Use a colon, a comma, parentheses or a hyphen.
- The book's voice, not a translation's voice: translate meaning, not
  sentences; rebuild idioms as real English idioms; same register (clear,
  warm, precise, no hype). Bad English is the number one failure mode and
  no test catches it: reread your output line by line before finishing.
- French glosses of English finance terms disappear ("le volatility
  harvesting, la récolte de volatilité" -> "volatility harvesting").
- Numbers: decimal point, comma thousands, `4%` with no space, `6.6%`.
  Euro amounts STAY in euros (the plates and frozen data behind them are
  euro-denominated), written the way the pilots do: `EUR 40,000`,
  `EUR 1M`. A guard test rejects French number patterns in
  `assets/book/en/`.
- Structure is preserved: same headings in the same order, same callouts,
  same figures, same "L'essentiel à retenir" / "Pour aller plus loin"
  closing sections (English titles per the glossary). Callout TOKENS are
  syntax and do not change (`::: cle`, `::: astuce`, `::: attention`,
  `::: exemple`, `::: science`, `::: terrain`, `::: encart`, `::: admin`); only their
  content is translated. The optional title after the token IS translated.
- Wiki-links: `[[fr-slug]]` -> `[[en-slug]]` per `plannedEN`, written or
  not (an unwritten target degrades to plain text until it exists). Links
  to fr-only articles: see 2. Never invent a slug.
- Figures: `::: figure <id>` blocks keep the SAME id (the plate is
  single-source; the English edition translates its rendered SVG through
  `figureDict` in `pkg/firebook/figures_i18n.go`). The caption under the
  block is prose: translate it. A guard test requires the set of figure ids
  to be identical between the French and English article.
- The in-file `# Title` (line 1) must equal the manifest title (see 6).

## 5. Stamp the file

Line 2 of the English file, right after the title, is the source stamp:

```
<!-- source: <fr-slug> @ <hash> -->
```

where `<hash>` is the first 12 hex digits of the sha256 of the FRENCH file
bytes, taken NOW, at translation time:

```sh
shasum -a 256 pkg/firebook/assets/book/fr/<fr-slug>.md | cut -c1-12
```

The stamp is metadata: it never renders (a guard test checks that), and
`make book-drift` compares it with the current French file to detect drift.
A refreshed translation rewrites the stamp.

## 6. Register the article in the manifest

`pkg/firebook/manifest_en.go`, `CategoriesEN`: add a keyed `Article` entry
in the part that mirrors the French one, in French reading order:

```go
{Slug: "how-much-you-need", Title: "How much you need", Blurb: "...", Source: "combien-il-vous-faut"},
```

`Title` = the English `# Title` of the file, `Blurb` = the French blurb
translated (it is the index-page teaser), `Source` = the French slug. The
index page grows with the manifest.

## 7. Figures dictionary and the ledger

- If `make test` fails in `TestFigureDictionaryCoversEN`, the plates your
  article uses carry French text nodes not yet in `figureDict`
  (`pkg/firebook/figures_i18n.go`). Add the missing entries (plain French ->
  English; use the `"<figure-id>|<french>"` key form only when the same
  French string must translate differently in another plate). Numeric
  payloads are reformatted mechanically and need no entry. English labels
  run wider than French: if you can, run `scripts/figure-audit.sh en`
  (needs Chrome) and eyeball a PNG of any plate whose labels moved; if you
  cannot, say so in the report.
- Append one row to the ledger table of
  `docs/fire-book-en-edition-design.md` ("France-specific passages"
  section): FR slug, EN slug, generalize/adapt, one-line note of what you
  neutralized or rewrote and which fr-only links you dropped.

## 8. Check, commit

```sh
make check          # fmt-check + lint + test, must be green
make book-drift     # your article must have left the list
git add -A && git commit -m "firebook-en: translate <fr-slug> as <en-slug>"
git push
```

One commit per article, directly on `master`. The commit message names the
pair and, if any, the adapt decision and the dictionary entries added. Never
leave a finished translation uncommitted.

## 9. Report

End your session with: the pair (FR -> EN slug), generalize/adapt and what
exactly was neutralized or rewritten, fr-only links dropped, figureDict
entries added, whether the figure audit ran, and any French sentence you
were unsure how to render (quote it, give your choice).

## Anti-patterns seen in past campaigns

- Translating sentence by sentence and shipping calques.
- Keeping French number formatting (`1 000 000 €`, `6,6 %`).
- Translating a `[[slug]]` target into words that are not in `plannedEN`.
- Dropping or adding a `::: figure` block "because the plate is French":
  the plate is translated at render time, keep the block.
- Touching the French file in the same commit as the translation.
- Editing the design doc's ledger for another article than yours.
