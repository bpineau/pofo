# The FIRE book, English edition: design

Status: M1 (wiring) DONE 2026-08-01; campaign signals (fr-only marker,
FR -> EN pairing as data, number guard, translator's brief, glossary) DONE
2026-08-16; M2, the translation campaign, is next.
This document is the implementation brief for the
English edition of the embedded FIRE book ("Le FIRE tranquille",
`pkg/firebook`). Read `docs/fire-book-design.md` first: everything there
(depth bar, callout types, figure system, guard-test discipline) applies to
the English edition too unless this document says otherwise.

## Goal

A complete English edition of the book, served at `/firebook/en/` next to the
French one, with its own EPUB and OPDS catalog. Aimed at a US reader: the
"Fiscalité et cadre français" part does not appear; a shorter, simpler
US-framework part replaces it.

The French edition is and stays the SOURCE OF TRUTH. All future writing,
corrections and upkeep happen in French first, then propagate to English.
The design below is shaped by one requirement above all others: keeping the
two editions in sync must be cheap, mechanical and impossible to forget
silently.

Non-goals: no third language provision beyond what the layout already gives
for free (a language directory and an Edition value); no bilingual
single-source files; no per-reader language negotiation (the two editions are
two separate books at two separate URLs).

## Current state (what is French-hardcoded, and where)

The 2026-07 architecture anticipated this work: articles already live under
`assets/book/fr/`, the mount is already `/firebook/fr/`, and
`pkg/decumul/web/server.go` carries a comment reserving `/firebook/en/`.
The French-specific surface is concentrated in five places:

1. `manifest.go`: `Categories` (French titles and blurbs), plus the
   `siteName` / `siteDescription` / `siteLede` constants in `handler.go`.
2. `handler.go`: page chrome ("Sommaire", "Dans la même partie",
   "lien copié", "Version epub", aria labels, error strings), `lang="fr"`,
   `og:locale`, JSON-LD `inLanguage`, French file sizes ("Ko"/"Mo"), and the
   hardcoded `assets/book/fr/` read path (also in `epub.go`).
3. `epub.go`: `epubFileName` ("le-fire-tranquille.epub"), `epubIdentifier`
   (the publication UUID), `Language: "fr"`, the title-page edition note,
   `bookHomePath`.
4. `pkg/bookmd/render.go`: the `Callouts` map carries French display labels
   ("L'idée clé", "Astuce", "Point de vigilance", ...).
5. `figures*.go`: every plate's kicker, title, axis and series labels are
   French literals inside the plate functions, and plate numbers use French
   formatting (decimal comma, narrow-space thousands, "6,6 %").

Everything else (bookmd syntax, epub writer, opds builder, webui identity,
the CSS) is language-neutral already.

## Architecture: the Edition value

Introduce one type that carries everything a language needs, and make the
existing package-level API thin French wrappers over it:

```go
// Edition is one language of the book: its manifest, its chrome strings,
// its EPUB identity and its figure-translation pass.
type Edition struct {
    Lang            string       // "fr", "en": html lang, epub dc:language, JSON-LD inLanguage
    OGLocale        string       // "fr_FR", "en_US"
    SiteName        string       // the book title
    SiteDescription string       // SEO meta description
    SiteLede        string       // hero sentence (index page + EPUB title page)
    HomePath        string       // "/firebook/fr/", "/firebook/en/"
    AssetDir        string       // "assets/book/fr", "assets/book/en"
    EPUBFileName    string
    EPUBIdentifier  string       // fixed urn:uuid, one per edition, never changes
    Categories     []Category
    UI              UIStrings    // all chrome strings, see below
    Callouts        map[string]bookmd.Callout // per-language callout labels
    Figure          func(id string) string    // FigureSVG for fr, translated pass for en
}

var French  = &Edition{...}  // current values, verbatim
var English = &Edition{...}
```

`UIStrings` collects every user-visible chrome string currently inlined in
`handler.go` and `epub.go`: index link label, "same category" footer title,
"link copied" toast, anchor aria labels, EPUB download label and error
strings, the title-page edition note, and a `HumanSize func(int) string`
(French "1,4 Mo" vs English "1.4 MB").

Methods move from package functions to the Edition: `(*Edition).Handler`,
`(*Edition).EPUB`, `(*Edition).Titles`, `(*Edition).ToHTML`, plus the
internal `find`, `indexHTML`, `articleHTML`. The existing exported API keeps
working as one-line wrappers over `French` (finador imports
`firebook.Handler()` and `firebook.EPUB()`; do not break it, and run
finador's tests after the refactor).

The EN publication UUID is generated once at implementation time and
hardcoded, exactly like the French one; the two editions are two distinct
publications to e-readers (annotations, refresh identity), so they MUST NOT
share an identifier. The EN title is "The Quiet FIRE" and the EPUB file is
`the-quiet-fire.epub` (settled 2026-08-01).

### bookmd change

Add `Callouts map[string]Callout` to `bookmd.Options` (nil keeps the
built-in French map, so locador and existing callers are untouched). The
callout TYPE tokens in markdown (`::: cle`, `::: astuce`, ...) are syntax,
not content: English articles keep the same tokens, which keeps FR and EN
sources diffable line-for-line. Only the rendered display labels change
("L'idée clé" -> "The key idea", "Astuce" -> "Pro tip",
"Point de vigilance" -> "Watch out", "Exemple chiffré" -> "Worked example",
"Ce que dit la recherche" -> "What the research says",
"Retour de terrain" -> "From the field", "En passant" -> "Aside").

## Articles: tree, slugs, manifest, pairing

- Files: `assets/book/en/<slug>.md`, embedded by the existing
  `//go:embed assets/book` directive (no build change needed).
- Slugs are TRANSLATED: the English book gets English URLs
  (`how-much-you-need`, not `combien-il-vous-faut`). Reader-facing URLs and
  SEO beat the convenience of shared slugs. Alternative considered and
  rejected: sharing French slugs across both editions would make pairing
  free but would put French URLs, French anchors and French cross-link
  targets in front of an American reader for the life of the book.
- Pairing is explicit data: `Article` gains a `Source string` field, the
  slug of the French original. Empty `Source` marks an edition-specific
  article (the US-framework part in EN; the whole French tax part simply has
  no EN counterpart). The French manifest leaves it empty everywhere.
- The EN manifest is its own file (`manifest_en.go`), same `Categories`
  shape, English titles and blurbs, categories in the same order (minus
  "Fiscalité et cadre français", plus "Taxes and the US framework" in the
  same position).
- Wiki-links in EN articles use EN slugs. The translator translates link
  targets along with the text; the guard tests catch any slug that does not
  resolve.

### Guard tests (mirror the French ones, per edition)

For the EN edition: manifest matches files both ways; every `[[slug]]`
resolves against the EN slug space (with the same planned-slug allowance);
every `::: figure` id resolves; every `Source` names an existing French
article, and no French article outside the tax part is left without an EN
counterpart once the edition ships (this last check is env-gated during the
translation campaign, hard afterwards). One new cross-edition check: for
each paired article, the SET of figure ids used must be identical between FR
and EN (a translated article cannot silently drop or add a plate).

The French numeric-claim guard tests (`tables_test.go`, plate-vs-prose
anchors) stay French-only: French is the source of truth and drift is caught
there. Extending claim-checks to EN prose is deliberately out of scope for
v1; the sync stamp below is what keeps EN numbers honest.

## Synchronization: stamp + drift report

The maintenance contract: editing a French article never breaks the build,
but the staleness it creates must be visible on demand and impossible to
lose.

- Every EN article carries, on the line right after its `# Title`, a stamp:
  `<!-- source: <fr-slug> @ <first 12 hex of sha256 of the FR file bytes> -->`
  written when the translation is made or refreshed.
- A guard test checks the stamp EXISTS and is well-formed on every paired EN
  article, and that the named FR file exists. It does NOT check freshness.
- `firebook.Drift()` (new, in the library, unit-tested) compares each
  stamp's hash with the current FR file and returns the stale list plus the
  FR articles with no EN counterpart. `./pofo -book-drift` prints it: that
  output IS the translation worklist. Add a `make book-drift` alias.
- Workflow after any French edit: commit the French change as usual, run
  `make book-drift`, translate the stale articles (each translation updates
  its stamp), commit. Nothing enforces immediacy; the report enforces
  memory.

Hash stamps beat git-based tracking (works on the embedded FS, no git at
test time, survives history rewrites) and beat single-source bilingual files
with conditional markers (which would make every French edit wade through
English text and were rejected outright).

### Campaign signals (2026-08-16)

The M2 worklist is read by translation agents, so it says everything a fresh
session needs and says it once, in the report itself.

- FR-ONLY MARKER. A French article that belongs to the French edition alone
  declares it in its own file, on the line right after its `# Title`:
  `<!-- edition: fr-only: French law end to end -->`. The text after the
  second colon is optional documentation. Drift reports such an article as
  "fr-only" and never as "untranslated", and `articleBody` strips the marker
  before rendering, exactly as it strips the source stamp. The marker carries
  the same argument as the stamp: it travels with the article, so moving,
  renaming or copying a file cannot separate the two. The seven articles of
  the tax part carry it. The old hardcoded skip list left the library, since
  the files now answer the question; the list survives as a test-only
  expectation, checked both ways against the markers, so neither a lost nor a
  stray marker passes unnoticed.
- FR -> EN PAIRING IS DATA. `plannedEN` is a table of `{EN, FR}` pairs, not a
  list of English slugs with the pairing in trailing comments, and Drift fills
  `DriftItem.ENSlug` with the planned English slug of every untranslated
  article. A worklist line is therefore a complete instruction: the French
  file to read and the English file to write. Guard tests hold the table to
  covering every non-marked French article exactly once, and to agreeing with
  the `Source` of every translation already written.
- NUMBER GUARD. A guard test greps every file under `assets/book/en/` for
  French number formatting (a decimal comma before a percent sign, a
  (narrow) no-break space between digits, a space before a percent sign) and
  reports file:line. It covers the whole tree rather than the manifest, so a
  file lands guarded from the moment it is written. The figures already had
  the equivalent check on their translated payloads.

`pofo -book-drift` prints one line per unsettled article, `<reason> <fr-slug>
-> <en-slug>`, and a three-way count on stderr.

## Figures: one generator, a translation pass

The ~100 plates stay SINGLE-SOURCE: every `func() string` keeps its French
literals and is never duplicated. The English edition translates the
RENDERED SVG:

- `translateFigure(id, svg string) string` in a new `figures_i18n.go`
  post-processes the text nodes (`<text>` and `<tspan>` payloads; the
  generator's output shape is our own, so extraction is a small regex, not
  an XML parser). Each payload goes through, in order:
  1. the dictionary: `map[string]string` French -> English, with an optional
     per-plate override key (`"<figure-id>|<french>"`) for the rare string
     that translates differently in different plates;
  2. if not in the dictionary and the payload matches the NEUTRAL pattern
     (digits, %, ×, ±, +, −, /, dates, isolated capitals), a mechanical
     French-to-English number reformat: decimal comma to point, narrow-space
     thousands to comma, "6,6 %" to "6.6%";
  3. otherwise it is left untouched, and the guard test fails.
- The guard test renders every plate referenced from `assets/book/en/`
  (scanning `::: figure` blocks), extracts every text payload and asserts it
  is dictionary-covered or neutral. The covered set grows with the
  translation campaign and equals all plates once the edition is complete;
  the dictionary never has to be filled ahead of the articles that need it.
  Consequence: changing a label in a plate used by a translated article
  breaks `make test` until the dictionary entry is updated IN THE SAME
  COMMIT. Figures get the hard guard (they change rarely, and the
  dictionary lives next to the code); prose gets the soft drift report.
- `English.Figure` = FigureSVG then translateFigure; `French.Figure` =
  FigureSVG unchanged.
- Captions are not figure text: they live in the article markdown and are
  translated with the prose.
- English labels run wider or narrower than French ones, and the plates were
  tuned to the French strings. Formalize the label-overflow audit that the
  illustration campaigns ran by hand into `scripts/figure-audit.sh
  [fr|en]`: render every plate of the edition through the real page CSS in
  headless Chrome and compare each text node's `getComputedTextLength`
  against its room (the silent right-edge clipping the eye misses). Not part
  of `make check` (needs Chrome); run it over the EN edition whenever the
  dictionary changes, and eyeball a PNG render of any plate whose labels
  moved, per the book's print-quality bar.

Alternatives rejected: threading a language argument through every plate
function (touches ~40 files and every future plate for zero added power);
duplicating plate functions per language (the exact maintenance trap this
design exists to avoid).

## Serving, SEO, EPUB, OPDS, CLI

- Mounts: `serve.go` and `pkg/decumul/web/server.go` add
  `/firebook/en/ -> http.StripPrefix("/firebook/en", firebook.English.Handler(...))`.
  Site nav gains "Book-en" next to "Book-fr" (hub page too). No redirects
  involved; `/book/` redirect handling is untouched.
- Per-article language switch: a `WithAlternate(base string, ed *Edition)`
  handler option adds, when the counterpart article exists (via `Source`
  pairing), a `<link rel="alternate" hreflang="...">` pair in the head and a
  discreet "English version" / "Version française" link in the top bar.
  hreflang needs the sibling mount's base path, which only the caller knows;
  hence the option, mirroring how `WithNav` already works.
- `og:locale`, `lang`, JSON-LD `inLanguage` come from the Edition.
- EPUB: `English.EPUB(modified)` with the EN identifier, `Language: "en"`,
  the EN title page and edition note. Validate with epubcheck and on-device
  KOReader like the French one (see `docs/epub-export-design.md`).
- OPDS: each edition's handler serves its own `opds.xml` listing its own
  book, exactly as today. A combined two-entry catalog was considered and
  dropped: a reader adds the edition it reads.
- CLI: `-export-epub` gains `-book-lang fr|en` (default `fr`).

## Editorial scope (what gets written, what gets adapted)

### The US-framework part

"Fiscalité et cadre français" (7 articles) does not exist in EN. In its
place, a deliberately SHORTER and SIMPLER part, "Taxes and the US
framework", three articles:

1. Accounts and account order: 401(k)/403(b), Traditional and Roth IRA,
   HSA, taxable brokerage; early-access mechanics that matter to FIRE (Roth
   conversion ladder, 72(t)/SEPP, the age-55 rule).
2. Taxes on the withdrawal phase: long-term capital gains brackets and the
   0% bracket, qualified dividends, state taxes in one paragraph, why the
   effective rate of a modest FIRE withdrawal is often near zero.
3. Healthcare before Medicare and Social Security in a FIRE plan: ACA
   subsidies as an income cliff, Medicare at 65, how a decades-long FIRE
   career shrinks Social Security and how to read an SSA estimate.

Same conventions as the French tax pages: dated-accuracy warning, research
against live primary sources (IRS, SSA, healthcare.gov) at writing time, no
personal-advice tone. These three articles are ORIGINAL writing, not
translations, and are the only EN-only content; they carry empty `Source`.

### France-specific passages inside general articles

Many general articles carry French material outside the tax part: "la
pratique française" sections (or-en-retrait, obligations-indexees),
French envelope mentions woven into examples, the French annuity frame in
rentes-et-annuites, and one article that is European practice end to end
(etf-ucits-europeens). The author has not settled this question; the
implementation must apply the following default and record every deviation,
so the choice stays revisable:

- DEFAULT (generalize): the translation replaces the France-specific passage
  with locale-neutral phrasing and, where the reader needs the concrete
  local answer, one pointer sentence to the relevant article of the US
  part. Cheap, safe, and keeps paired articles structurally close, which
  the sync stamp rewards.
- ESCALATION (adapt): where local practice IS the article's point, the EN
  article rewrites the passage for the US reader instead of neutralizing
  it. Known candidates: etf-ucits-europeens becomes "building it with
  US-listed ETFs" (paired, same slot, but flagged as heavily adapted);
  rentes-et-annuites' French annuity mechanics become the US SPIA/DIA
  market; or-en-retrait's French buying practice becomes US practice.
- A ledger table in this document (added during the campaign: FR slug, EN
  slug, generalize/adapt, one-line note) records the per-article decision.
  Adapted passages still live under the same stamp: a French edit flags the
  EN article stale, and the translator re-derives the adapted passage.

The ledger, opened with the M1 pilots:

| FR slug | EN slug | Decision | Note |
|---|---|---|---|
| sequence-des-rendements | sequence-of-returns | generalize | Nothing France-specific; a straight translation. |
| vpw | vpw | generalize | The pension-bridge passage pointed at `retraite-legale`, a tax-part article with no EN counterpart. Neutralized to "before your pensions start" with NO replacement pointer: a cross-link there earns little, and a reader who wants the local rules will find the US part on their own. The "phase adossée d'un plan FIRE français" aside became locale-neutral. |

#### Triage of the general articles (2026-08-16, awaiting arbitration)

A sweep of the non-tax articles for France-specific density (term counts,
then a reading of the sections) yields the list below. "Proposed" rows are
NOT decisions: until a row is settled here, a translator applies the
GENERALIZE default and reports the passages that pushed for more (see
`docs/fire-book-en-translation-brief.md`, section 3). Three outcomes are
possible per article: fr-only (marker on the French file, no English
counterpart), adapt (one or more sections rewritten for the US reader), or
generalize.

| FR slug | EN slug | Proposed | Why |
|---|---|---|---|
| cas-types | three-worked-plans | fr-only OR original rewrite | The three households are built on PEA / assurance-vie / CTO / PUMa / pension quarters; a US version is three new plans, not a translation. |
| inflation-histoire | inflation-history | adapt (heavy) OR fr-only | The spine is French monetary history 1914-2025 (ruined rentiers, post-war financial repression, 1974-81); a US reader expects the dollar's 1913-2025. |
| etf-ucits-europeens | building-it-with-us-etfs | adapt (decided 2026-08-01) | European practice end to end; effectively an original article on US-listed ETFs. |
| cash-ameliore | enhanced-cash | adapt | The "fonds euros" third has no US equivalent (stable value, MMF, T-bills); CLO AAA and money-market parts translate. |
| immobilier-en-retrait | real-estate-in-retirement | adapt | French property taxation, SCPI, viager, dismemberment, fixed-rate mortgage frame; US counterparts are REITs, reverse mortgage, HELOC. |
| retour-au-travail | going-back-to-work | adapt | "La boîte à outils française du travail dosé" and the "quadruplé" (PUMa, quarters); the US pendant is ACA subsidies and Social Security credits. |
| echelle-obligataire | bond-ladders | adapt | "La pratique française : les contournements du guichet absent" inverts in the US (TreasuryDirect exists). |
| obligations-indexees | inflation-linked-bonds | adapt | "La pratique française" section becomes TIPS / I-bonds. |
| diversification-internationale | international-diversification | adapt | Home bias and the currency section are written from the euro side; for a US reader home bias IS the S&P and the FX asymmetry flips. |
| combien-il-vous-faut | how-much-you-need | generalize + pointer | Step 2 "la correction fiscale" is PFU/PUMa; replace with a pointer to `us-taxes-in-the-withdrawal-phase`. |
| revue-annuelle | the-annual-review | generalize | Bloc 4 (fiscal/administratif) neutralized, pointer to the US part. |
| construire-son-plan | building-your-plan | generalize | Step 7 (day-1 execution) names French envelopes and brokers. |
| suivre-inflation | tracking-inflation | adapt (light) | IPCH/INSEE indices become CPI-U/PCE. |
| bibliotheque | the-library | adapt (light) | "Les sources officielles françaises" section becomes IRS/SSA/BLS. |
| lexique | glossary | generalize | Drop the entries of the tax part (PUMa, fonds euros, ...); add the US-part terms once M3 lands. |
| rentes-et-annuites | annuities-and-safety-first | adapt (decided 2026-08-01) | French annuity products become the US SPIA/DIA market. |
| or-en-retrait | gold-in-retirement | adapt (decided 2026-08-01) | French buying practice becomes US practice. |

Light generalization suffices (UCITS/PEA mentions in passing) for
managed-futures, facteurs-fama-french, allocation-actions-obligations,
levier-et-marges, hyperinflation-et-extremes, temoignages-fire (the quoted
corpus is already English-language), la-machine-pofo, utiliser-la-page-fire.

Rejected: keeping "in France, ..." passages verbatim as curiosities (dead
weight for the target reader), and conditional markers in shared sources
(see Synchronization).

### Currency and numbers

Worked examples KEEP their euro amounts (formatted English: "EUR 1,000,000"
or the article's existing style anglicized), because the plates and the
frozen data arrays behind them are euro-denominated and single-source; a
currency swap would fork the data the guard tests pin. The US-framework
articles are natively in dollars. Prose numbers follow English convention
(decimal point, comma thousands, "4%" without the French space); a small
guard grep over `assets/book/en/` rejects French number patterns
(digit-comma-digit before %, narrow-space thousands) to catch translation
slips mechanically.

### English style sheet (the translator's brief)

- US English. No em-dash, same as the French rule.
- Same register as the French: clear, warm, precise, no hype; the book's
  voice, not a translation's voice. Translate meaning, not sentences:
  rebuild idioms as real English idioms.
- The French text glosses English finance terms at first use; in English
  those glosses simply disappear (the term stands alone). Book-specific
  French coinages map once in a glossary file kept next to the dictionary
  (`millésime` -> vintage, `matelas` -> cash buffer, `retrait` ->
  withdrawal, `rente` -> annuity, `traversée` -> crossing/bear stretch,
  ...); the glossary seeds consistency across ~90 files translated by many
  agent sessions.
- Callout tokens, wiki-link syntax, figure blocks: unchanged syntax,
  translated content, EN link targets.
- Titles and blurbs are translated in the manifest, and the in-file
  `# Title` must match the manifest title (same rule as French).

## Rollout

- M1, wiring (no public surface): Edition refactor + bookmd Callouts option
  + EN manifest scaffolding + stamp/drift machinery + figure translation
  pass with its guard + `scripts/figure-audit.sh`. Validated with two pilot
  articles (one figure-heavy, one table-heavy) living in `assets/book/en/`
  but NOT mounted. finador re-tested against the refactor.
- M2, the translation campaign: one agent session per article against the
  style sheet and glossary, stamp written per file, figure dictionary grown
  as plates surface, `figure-audit.sh en` run per dictionary change, one
  commit per article (the proven per-file cadence of the French review
  campaigns). The per-article line review discipline from the French
  campaigns applies: mechanical checks do not catch bad English any more
  than they caught bad French.
- M3, the US-framework part: three original articles, researched live,
  woven into EN hub articles (lexique equivalent, panorama, revue-annuelle)
  the same way French hubs link the French tax part.
- M4, publication: mount `/firebook/en/`, nav + hub links, `WithAlternate`
  cross-links, `-export-epub -book-lang`, epubcheck + KOReader validation,
  the completeness guard flips from env-gated to hard. Update
  `docs/webapp-design.md` (route map), `CLAUDE.md` (map row), `README.md`,
  and this document's status line.
  Partially pulled forward (2026-08-01, the landing-page reorganization):
  `/firebook/en/` is mounted under `-serve` with cross-navigation both ways,
  and the index skips parts with no translated article yet, so the page grows
  with M2. Still owed to M4: `WithAlternate` hreflang cross-links,
  `-export-epub -book-lang`, epubcheck/KOReader validation, the hard
  completeness guard.

M1 is a self-contained refactor PR-sized task; M2 is the bulk of the cost
(~91 articles, ~220k French words) and is embarrassingly parallel after M1.

## Decisions settled 2026-08-01

1. The English title is "The Quiet FIRE" (epub file `the-quiet-fire.epub`).
2. France-specific passages default to GENERALIZE, with the three named
   adapt escalations (etf-ucits-europeens, rentes-et-annuites,
   or-en-retrait).
3. `utiliser-la-page-fire` is translated as-is: it documents the FIRE
   simulator page, whose UI is already in English, so only the prose needs
   translating (links and control names come out naturally).

## M1, as shipped (2026-08-01)

All of the machinery, none of the public surface: `/firebook/en/` is still
unmounted. What landed, and the three places the implementation had to
deviate from this document:

- `bookmd.Options.Callouts` (nil keeps the French labels), the `Edition`
  refactor with `French` and `English`, `manifest_en.go` (`CategoriesEN`,
  `plannedEN` as the FR -> EN slug map, `taxOnlyFR`, `usFrameworkEN`),
  `drift.go` (`Drift`, source stamps) behind `pofo -book-drift` and
  `make book-drift`, `figures_i18n.go` (`FigureSVGEnglish` + `figureDict`),
  the EN guard tests, `scripts/figure-audit.sh`, and two pilot translations
  (`sequence-of-returns`, `vpw`). The French rendering was proven
  byte-identical throughout by a temporary digest harness, since deleted.
- DEVIATION: the French manifest used positional `Article` literals, so
  adding `Source` forced all 91 of them to keyed form.
- DEVIATION: `bookCSS` carried one edition-dependent literal (the "link
  copied" toast); the const is split in two around it.
- DEVIATION: the source stamp rendered as a visible paragraph. Body
  preparation is now one `articleBody` helper shared by the web and EPUB
  renderers, which drops the stamp; a guard test fails if it reaches a page.
- `plannedEN` and the two slug lists live in `manifest_en.go`, not in the
  test file, because the French `planned` they mirror is production code too.

The M2 worklist is whatever `make book-drift` prints (82 untranslated
today).
