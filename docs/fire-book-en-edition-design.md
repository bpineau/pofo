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
| fire-cest-quoi | what-is-fire | generalize | The "atouts spécifiques du lecteur français" passage (efficient wrappers, a state pension arriving later, PUMa) became one locale-neutral sentence on the plumbing that changes from country to country, with a single pointer to `us-accounts-and-account-order`; the reading-guide bullet "Vous êtes français" became "You want the local framework" and points at the three US-part articles. Links dropped: `enveloppes-francaises`, `retraite-legale`, `taxe-puma`. The product name in the simulator bullet dropped per the glossary; r/vosfinances replaced by r/financialindependence. |
| la-regle-des-4-pourcents | the-4-percent-rule | generalize | The fee row of the assumptions table lost "un contrat d'assurance-vie" for "an insurance wrapper", the tax row lost "PFU, prélèvements sociaux, PUMa" for "tax on a withdrawal is a spending item" with one pointer to `us-taxes-in-the-withdrawal-phase`, and the three "retraite légale" mentions became a plain pension (income table row, the science callout's list of protections, the worked example's EUR 800 a month from 67). Links dropped: `flat-tax-et-imposition`, `taxe-puma`, `retraite-legale`. Dictionary: the sixteen labels of the `millesimes-soutenables` plate. |
| combien-il-vous-faut | how-much-you-need | generalize + pointer | Step 2 kept its logic (tax on withdrawals is a spending item you gross up) and lost every French rate: the PFU / prélèvements sociaux / assurance-vie / PUMa paragraph became the locale-neutral shape of the correction (only the gain is taxed, the embedded gain fraction climbs, a sheltered account defers without erasing) plus one pointer to `us-taxes-in-the-withdrawal-phase`, and the "+ 10 to 15% for a French retiree" ballpark became a friction the reader estimates and states as an assumption. The worked example keeps its 12%, now labeled an assumption. Step 1's mutuelle line lost its EUR 150 to 300 a month and points at `us-healthcare-and-social-security`; step 3's "retraite légale" became a plain state pension, same pointer; the info-retraite.fr resource became the reader's own pension statement, with the SSA statement named for the US. Links dropped: `sante-et-protection-sociale`, `enveloppes-francaises`, `flat-tax-et-imposition`, `taxe-puma`, `retraite-legale`. Dictionary: the two `cible-*` plates, whose only French-tax label ("impôts et PUMa") translates as "tax on withdrawals". |
| les-trois-phases | the-three-phases | generalize | Accumulation point 4 was the French wrapper clocks (PEA and assurance-vie dated early, the 5 and 8 year counters, the PEA ceiling filled over years); it became the locale-neutral shape of the same advice (tax shelters pay you in time, clocks that start when an account is opened and funded, contribution room that accrues and expires) with one pointer to `us-accounts-and-account-order`. The transition's "ajoute des trimestres" became "adds to the pension you will eventually claim". The "Pour aller plus loin" simulator bullet dropped the product name per the glossary. Links dropped: `enveloppes-francaises`, `retraite-legale`. Dictionary: the twenty labels of the `flux-relatif-phases` plate; audit green, plate eyeballed. |
| utiliser-la-page-fire | using-the-fire-simulator | generalize | Translated as-is per decision 3: the simulator's UI is already English, so the control names, the six model columns and the ten section titles are quoted verbatim from `pkg/decumul/web/assets/index.html` and `app.js` rather than back-translated. Three France-specific passages neutralized. The Taxes group lost "32,8 %, soit le PFU 2026 à 31,4 % plus une approximation de la PUMa": the preset is now "one country's blended rate and nothing more", with one pointer to `us-taxes-in-the-withdrawal-phase`. The pension paragraph's "le résultat du simulateur officiel" became "whatever your official statement says", pointing at `us-healthcare-and-social-security` for reading a US one. Section 05's "la mortalité d'un couple français" became "a couple's odds of still being alive", with the French life tables named in parentheses since the page really does use them. The article names the tool in its opening line and in the readme bullet, per glossary 4.2. Links dropped: `flat-tax-et-imposition`, `taxe-puma`, `enveloppes-francaises`, `retraite-legale`. No figure block, so no dictionary entry. |
| erreurs-classiques-fire | ten-plan-wrecking-mistakes | generalize | Mistake 3 was "Ignorer la fiscalité et la PUMa": the PFU / prélèvements sociaux / cotisation subsidiaire maladie mechanics and the 8 to 20% friction band became the locale-neutral shape (which account you sell from, how much of the sale is gain, the local rules, health coverage once no employer pays), with one pointer to `us-taxes-in-the-withdrawal-phase`; the default 10 to 15% gross-up became "an explicit friction assumption, the way the chart above uses 12%", which keeps the plate anchored, and the "organiser les enveloppes PEA/AV/CTO" sentence became the order you fill accounts in. Mistake 4 lost "la retraite légale" and the 64-67 age band for a plain state pension arriving "in your sixties", and info-retraite.fr became your own official statement, with the Social Security estimate named for the US. Mistake 6's "60 % de fonds euros" became "60% parked in cash and stable-value holdings". Links dropped: `flat-tax-et-imposition`, `taxe-puma`, `enveloppes-francaises`, `retraite-legale`; `revenus-complementaires` picked up the pension pointer instead. Dictionary: the 22 labels of the `cout-des-erreurs` plate (the tax bar reads "Tax on withdrawals ignored"); audit green, plate eyeballed. |
| etude-trinity | the-trinity-study | generalize | Almost nothing France-specific to start with: an intellectual-history article on Bengen 1994 and Trinity 1998. The "Refaire Bengen vous-même" section named the product ("les outils qui savent lire un portefeuille réel, comme pofo"); the name dropped per glossary 4.2 for "a simulator that reads a real portfolio", FICalc and cFIREsim kept since the French names them. The worked example keeps its euro figures (EUR 1M, and the "one euro left" definition of success). No fr-only link in the article, so none dropped. No figure block, so no dictionary entry; audit green.
| ruine-et-probabilites | failure-probability | generalize | Barely any France-specific material: the article is about how to read the failure-probability number. Two "pension legale" mentions were neutralized, one into "Social Security underneath it" (the retiree who runs out of breath late), the other into a plain "a pension coming" among the safety nets; the closing pointer "comment pofo calcule cette ruine" became "how the simulator computes this number", per glossary 4.2. No French rate appears in the source. No fr-only link, so none dropped. No figure block, so no dictionary entry; audit green.
| rendements-arithmetiques-geometriques | arithmetic-vs-geometric-returns | generalize | Almost nothing France-specific: the article is the two means, the volatility drag and the cascade down to the withdrawal rate. Step 4 of the cascade lost its French wrappers ("selon vos enveloppes et véhicules") for "depending on the accounts and the funds you use", with the fr-only `flat-tax-et-imposition` link dropped and one pointer to `us-taxes-in-the-withdrawal-phase` alongside `building-it-with-us-etfs`. Two product mentions dropped per glossary 4.2: "un outil de visualisation de portefeuille comme pofo" became "a portfolio visualizer", and the leveraged-stack table row "Empilement 90/60 (type NTSG)" became "Stacked 90/60 (efficient core)". No French rate in the source. The `vol-drag` figure block is unchanged and needed no dictionary entry; audit green.
| rendements-attendus | expected-returns | generalize | Nothing France-specific in the argument, only euro-investor examples in the bond block. "une OAT ou un Bund 10 ans à ~3 %" became "a 10-year government bond at about 3%", with TIPS kept as the instrument that prints its real yield; "le cash suit le taux directeur réel, autour de 0 à 1 % réel en zone euro" lost "en zone euro"; the worked example's "30 % obligations euro aggregate" became "30% investment-grade bonds", every number left untouched so the prose still matches the `briques-esperance` plate. The plate itself was generalized in the dictionary: "Obligations euro" -> "Government bonds", "OAT ou Bund 10 ans, nominal" -> "10-year government bond, nominal", "Monétaire euro" -> "Cash". "pofo mélange les paramètres vers un prior mondial" became "the simulator behind this book blends parameters toward a world prior", per glossary 4.2. No French rate in the source, no fr-only link, so none dropped. Dictionary: 41 entries for the two plates (`briques-esperance`, `escalier-morningstar`), two of them keyed because "distribution" and "actions" are ambiguous across plates; audit green.
| les-maths-du-4-pourcent | the-math-of-4-percent | generalize | Almost nothing France-specific: the article is the three-tier cascade behind the 4% figure. Only the "ce qui la casserait" list carried French material, in the frictions item; "frais et fiscalité se retranchent de l'étage 1 presque un pour un" keeps its wording, the fr-only `flat-tax-et-imposition` link is dropped, and one pointer to `us-taxes-in-the-withdrawal-phase` sits alongside `building-it-with-us-etfs`. No French rate in the source, no product name in the prose. Dictionary: 25 entries for the two plates (`cascade-4pct`, `clavier-leviers`), "étage N" rendered as "tier N" per the glossary's three-tier budget; audit green. |
| decider-sous-incertitude | deciding-under-uncertainty | generalize | Nothing France-specific in the source: the article is decision theory (utility, certainty equivalent, tolerance against capacity, Kelly, regret, satisficing, robustness) and its five-rule protocol. No French rate, no envelope, no fr-only link, so none dropped. The two simulator mentions were already generic in the French and stay unnamed per glossary 4.2 ("a simulator draws ten thousand", "a simulator settles the question well only when you ask it the right one"). Euro amounts kept as euros throughout, French decimals converted. Dictionary: 12 entries for the `utilite-ce` plate, including "ÉC" -> "CE" and the axis label "revenu annuel (k€)" -> "annual income (EUR k)"; audit green. |
| monte-carlo-forces-faiblesses | monte-carlo-strengths-and-limits | generalize | Nothing France-specific in the argument: the article is the simulation machine itself, its four weaknesses and eight usage rules. No French rate, no wrapper, no fr-only link, so none dropped. Two product mentions removed per glossary 4.2: the list of complete simulators ("pofo, TPAW Planner ou Portfolio Visualizer") lost the product name, and "le détail de cette boucle telle que pofo l'implémente" became "how that loop is really implemented" pointing at `under-the-hood`. Euro amounts kept as euros (long-term care at EUR 300 a day, EUR 800 a month of side income). Dictionary: 16 entries for the `mc-entrees-vs-tirages` plate, including "MONTE-CARLO" -> "MONTE CARLO" and the caption line; audit green. |
| historique-vs-parametrique | historical-vs-parametric | generalize | Nothing France-specific in the source: the article is the three families of return sources (historical windows, block bootstrap, parametric) and what their disagreements mean. No French rate, no wrapper, no fr-only link, so none dropped. One product mention removed per glossary 4.2: the tool-inventory table row "pofo" became "This book's simulator", and the transparency sentence "pofo accompagne ce livre et son auteur n'est pas neutre a son sujet" became "the simulator that comes with this book is not one its author can judge from the outside". The six model names are quoted with the UI strings (Historical windows, Block bootstrap, Student-t, Sequence stress, Broad-sample, Lost decade). No figure block, so no dictionary entry; audit green. |
| queues-epaisses | fat-tails | generalize | Nothing France-specific in the source: the article is the fat-tail phenomenon, the Student-t and its df parameter, and what tails change in a withdrawal plan. No French rate, no wrapper, no fr-only link, so none dropped. One product mention removed per glossary 4.2: the closing bullet "L'estimation du df et son usage exact dans pofo" became "How df is estimated, and exactly how it is used", still pointing at `under-the-hood` and `using-the-fire-simulator`. Euro amounts kept as euros in the sensitivity test (EUR 1.5M, EUR 52,000 a year). Dictionary: 9 entries for the `fat-tails` plate, including the title "À volatilité égale, deux mondes : la cloche et ses queues" -> "Same volatility, two worlds: the bell and its tails" and "~10× plus d'années à −30 %" -> "~10× more years at −30%"; audit green. |
| lire-un-fan-chart | reading-a-fan-chart | generalize | Nothing France-specific in the source: the article is how to read a wealth fan. Two neutralizations only, both product naming per glossary 4.2: the list of simulators "pofo, FICalc, cFIREsim, Portfolio Visualizer" drops the product name and keeps the three public tools, and "Dans le simulateur FIRE de pofo" becomes "In the FIRE simulator", with the section it points at quoted as the real UI string (Simulated futures, Lost decade, the wealth fans). The euro figures of the worked example are kept (EUR 1.6M, EUR 55,000 a year). No fr-only link in the article, so none dropped. Dictionary: 11 entries for the two plates (`fan-anatomy`, `fan-two-plans`); audit green. |
| pieges-des-simulateurs | simulator-traps | generalize | The only France-specific passage is the tax pointer of trap 7: "au taux mixte que vous reglez ... ([[flat-tax-et-imposition]])" becomes "at the blended rate you actually pay, with an embedded gain fraction that rises over the plan ([[us-taxes-in-the-withdrawal-phase]])", the single local pointer of the article, and "un contrat charge" (a loaded French life policy) becomes "an expensive account". One fr-only link dropped, `flat-tax-et-imposition`; no other fr-only link in the source. Product naming per glossary 4.2: the list of tools with nothing to sell "cFIREsim, FICalc, la toolbox d'ERN, pofo" becomes "cFIREsim, FICalc, ERN's toolbox, the simulator behind this book", and "comment pofo repond, piege par piege" becomes "how the simulator answers, trap by trap". Euro figures kept (EUR 1.2M, EUR 45,000 a year, EUR 3,000). No figure block in the article: no dictionary entry, audit green by construction. |
| rendre-monte-carlo-pertinent | making-monte-carlo-relevant | generalize | Fix 6 lists what a realistic plan carries: "la pension a sa date ([[retraite-legale]])" becomes "your pension on the date it starts" (fr-only link dropped, no replacement), and "la fiscalite qui majore chaque vente" becomes "the taxes that gross up every sale under the account rules where you live ([[us-taxes-in-the-withdrawal-phase]])", the single local pointer of the article. No French rate anywhere in the source, so nothing else to neutralize. Product naming per glossary 4.2: "ces six corrections telles que pofo les implementes" becomes "these six fixes as the simulator implements them"; the controls the article describes keep their real UI wording (the broad-sample prior, the CAPE anchor, df, the Sequence stress and Broad-sample readings). Euro figures kept (EUR 1.4M, EUR 50,000 a year, EUR 14k). No figure block in the article: no dictionary entry, audit green.
| regimes-de-marche | market-regimes | generalize | Only two France-specific spots. The worked audit example held "30 % obligations d'Etat euro nominales 7-10 ans", generalized to "30% nominal government bonds of 7 to 10 years" (the euro leg was incidental, the argument is about nominal duration). The forecasting callout credited "le Permanent Portfolio tactique de Darcet", a French-language practitioner an English reader cannot follow up: generalized to "tactical versions of the Permanent Portfolio shift their four sleeves with the measured macro regime". No fr-only wiki-link in the source, so none dropped; the UK 1970s stagflation, the 1930s deflation and Japan stay, they are evidence in a multi-country list. No French rate anywhere. Product naming per glossary 4.2: the closing bullet's "l'implementation ... dans pofo" becomes "for how the sequence stress, the broad sample and the lost decade are implemented". 13 figureDict entries added for the regime-grid plate, audit green.
| panorama-strategies-retrait | withdrawal-strategies-overview | generalize | Two France-specific spots only. Family 5's "désormais couvert par une rente ou une pension ([[retraite-legale]])" became a plain annuity or pension, and the closing line "C'est justement ce que fait un plan français typique : pension légale en plancher, portefeuille en confort" became "most real plans already are one: Social Security covers the floor, the portfolio funds comfort", with the single pointer to `us-healthcare-and-social-security` alongside `pensions-and-other-income`. The bequest criterion dropped its `succession-et-transmission` link with no replacement. "les RMD américains" became "the RMD tables", now the reader's own rule. The frontier-drawing paragraph and the closing "comment pofo simule ces règles" leave the tool unnamed, per glossary 4.2. No French rate in the source; the euro figures of the four-rule case are unchanged. Links dropped: `retraite-legale`, `succession-et-transmission`. Dictionary: 33 entries for the two plates (`withdrawal-frontier`, `familles-information`), seven of them keyed because the family-map column headers wrap into bare articles ("le", "les", "les flux"); the plate's "rachats de trimestres" became "delayed claiming". Audit green. |
| retrait-fixe-bengen | fixed-inflation-adjusted-withdrawal | generalize | Almost no France in the source. The indexation index ("sur l'IPC", "la règle canonique indexe sur l'IPC national") became the generic consumer price index, and the blind-spot sentence "Indexer sur l'IPC en laissant la dérive santé non financée" became "indexing on the index while leaving the health-care drift unfunded". The closing pointer "Les réglages correspondants dans pofo sont détaillés dans [[utiliser-la-page-fire]]" leaves the tool unnamed and now quotes the real control string, "Cut in downturns (0 = fixed rule)", per glossary 4.2. No French rate in the source, no fr-only link to drop, and the euro figures of the plates and of the worked example are unchanged. Dictionary: 22 entries for the two plates (`bengen-falaise`, `bengen-millesimes`), none keyed. Audit green. |
| pourcentage-fixe | fixed-percentage | generalize | Two France-specific touches only. "C'est le profil de la retraite française installée" became "That is the profile of a retiree whose pensions are already in payment", and the perpetual-endowment profile lost its `[[succession-et-transmission]]` link (fr-only), keeping "passing on a real capital intact, endowment-style". No French rate anywhere in the source. The simulator stays unnamed and the astuce now quotes the real controls, "Spend % of portfolio (VPW)" and "Bounded % of portfolio (Vanguard-style)", per glossary 4.2. Settled vocabulary applied: "effet ciseaux" -> the squeeze (glossed once), "ruine de train de vie" -> lifestyle ruin. Euro figures of the worked example unchanged (EUR 1.4M, EUR 56,000). Dictionary: 25 entries for the two plates (`pourcentage-lissages`, `borne-geometrique`), none keyed. Audit green after shortening the corridor footnote. |
| guyton-klinger | guyton-klinger | generalize | Nothing France-specific in the source: a rules-and-research article on the 2006 decision rules and their pathology. The only locale-bound word was "l'IPC" in the indexation bullet, now plain "inflation"; the "conseillers américains" of the opening already faced the target reader and stayed. No fr-only wiki-link appears in the article, so none was dropped, and no French rate survives (the euro figures of the worked example are unchanged: EUR 1.3M, EUR 55,900, EUR 43,600). The simulator stays unnamed: "un outil qui accepte les quatre paramètres" became "a tool that takes all four parameters". Settled vocabulary applied: "traversée" -> the crossing (glossed once in the figure caption), "cliquet" -> ratchet, "revenu servi" -> the income delivered, "contresens" -> mistake. Dictionary: 8 entries for the `gk-cascade-1966` plate, none keyed. Audit green after shortening the plate's closing note. |
| regles-cape | cape-based-rules | generalize | Almost nothing France-specific in the source: the article is the a + b/CAPE formula, its parameters and its implementation. The only locale-bound link was `etf-ucits-europeens` in the "CAPE of what?" paragraph, kept as the EN counterpart `building-it-with-us-etfs` since that article is translated, with the sentence rewritten around a global equity sleeve rather than a European one. "pofo combine les deux de la même façon" became "the simulator behind this book combines the two the same way", per glossary 4.2. No French rate, no wrapper, no fr-only link, so none dropped. Euro amounts kept as euros (Sofia's EUR 1.5M example). Dictionary: 30 entries for the two plates (`cape-contracyclique`, `cape-depuis-1881`), including the six k EUR income labels; audit green. |
| guardrails-morningstar | morningstar-guardrails | generalize | The last section, "La transposition française, le simulateur pour indicateur", becomes "The executable version, with the simulator as the instrument": nothing in its body was French, only the framing, so the heading and its opening sentence were neutralized. "la pension américaine en toile de fond" -> "with Social Security in the background" (a US reader is the target, so the aside stops being foreign); the RMD gloss "le retrait minimal que le fisc américain impose" -> "the minimum withdrawal the IRS forces on retirees"; "recopié d'un article américain" -> "copied out of somebody else's article". "dans pofo" in the last bullet of Going further became "step by step" with the simulator unnamed (glossary 4.2), and the continuous-version paragraph names the real control, Risk-based guardrails (Kitces / Morningstar). No fr-only wiki-link in the source, so none dropped; every one of the 21 links maps through plannedEN. No French rate anywhere. Euro amounts kept (EUR 1.5M plan, EUR 54,000 comfort, EUR 44,000 floor). Dictionary: 29 entries for the two plates (deux-thermometres, guardrails-indicateur); audit green. |
| amortissement-abw | amortization-based-withdrawal | generalize | The actuarial rule is locale-neutral by construction; three passages carried France. The total-wealth definition and the mechanics section pointed at `retraite-legale` for "les pensions a venir": neutralized to "the pensions ahead of you, Social Security included" with one pointer to `us-healthcare-and-social-security`, the article's only local link. The four-move implementation recipe valued the portfolio "apres l'impot qu'un retrait declenche" with a `flat-tax-et-imposition` link: it became "after the tax a withdrawal triggers, under whatever account rules apply where you live", no pointer added and no rate anywhere. The bequest parameter dropped its `succession-et-transmission` link and keeps `spending-in-retirement` alone. The tool name went per glossary 4.2: "pofo en propose une politique equivalente" became "a general-purpose simulator carries the equivalent policy", quoting the real control string "Amortize over the horizon (ABW / TPAW)". Links dropped: `retraite-legale`, `flat-tax-et-imposition`, `succession-et-transmission`. Dictionary: 23 labels across the `richesse-totale` and `abw-1966` plates; audit green. |
| plancher-plafond | floor-and-ceiling | generalize | Almost nothing to neutralize: the rule, its two lineages and its arithmetic are locale-neutral, and no French link, rate or product appears in the source. Euro amounts stay in euros per glossary 5. Two touch-ups only: the "Deux bornes" opening called Vanguard "l'un des deux plus grands gestionnaires d'actifs du monde", kept as is since it is a fact and not a local reference, and the simulator went unnamed per glossary 4.2, with "cherchez une politique de depense de type pourcentage borne" quoting the real control string "Bounded % of portfolio (Vanguard-style)". Links dropped: none, every wiki-link had a `plannedEN` target. Dictionary: 18 labels across the `corridor-1966` and `corridor-borne` plates; audit green. |
| rentes-et-annuites | annuities-and-safety-first | adapt | The section "Le cadre francais : trois produits, dont un imbattable" (life-annuity payout from an assurance-vie contract and its rente viagere a titre onereux fractions, the PER, and the rachat de trimestres sold as the best annuity on the market) was rewritten in the same slot as "The US market: what you can buy, and the one annuity nobody sells you": SPIA / deferred income annuity / QLAC in one paragraph, the disappearance of the CPI-linked SPIA and the fixed graded step that replaced it, taxation in one sentence (exclusion ratio on taxable money, fully taxable on pre-tax money, IRS Publication 939), counterparty in one sentence (state guaranty association, $250,000 of present value in most states), and the US analogue of the pension buy-back, delaying Social Security to 70 (8% a year of delayed retirement credits, 42 U.S.C. 402(w), CPI-adjusted by statute, survivor benefit, pointing once at [[us-healthcare-and-social-security]]), closed by a dated-accuracy caveat sentence. Generalized elsewhere: the payout-rate ballpark (French TGH/TGF-05 table and ~4-4.5% at 65 -> a conservative annuitant table and recent US quotes near 8% at 65, 10% at 75, above 12% at 80), the 8-15% load figure dropped for a qualitative phrase, the 70,000 EUR fonds de garantie replaced by a pointer to the state guaranty ceiling, the inflation callout retitled and rewritten around US products, the deferred-annuity paragraph flipped (rare in France -> sold in the US, QLAC inside an IRA), the tontine paragraph flipped (Le Conservateur -> nothing at retail scale in the US yet), and the "how much" paragraph inverted: the French says the state pension already covers the floor so massive annuitization has no purpose, the English says an early retiree's short earnings record makes the residual gap real. Links dropped: [[flat-tax-et-imposition]], [[enveloppes-francaises]], [[retraite-legale]] (twice, once in prose and once in the "Going further" list, both replaced by [[us-healthcare-and-social-security]]); the info-retraite.fr / service-public.fr bullet became ssa.gov plus IRS Publication 939. Euro amounts and the worked example kept in euros per glossary 5, since the `etages-du-plancher` plate freezes them; the 5.3% joint quote got a half-clause explaining why it sits under the single-life rates. Dictionary: 23 labels across `credits-mortalite` and `etages-du-plancher`, product label "rente viagere, 12" -> "life annuity, 12"; audit green. |
| sept-facons-de-vivre | seven-ways-to-live-on-one-portfolio | generalize | Native to the US reader already (three real US retirements replayed on an S&P 500 / 5-year Treasury 60/40 deflated by CPI), so the pass was light. "un 60/40 americain ... deflate par l'indice des prix americain" -> "the standard 60/40 ... deflated by CPI" (the nationality of the yardstick is implicit in English); "L'historique americain disponible" -> "The available US history". Links dropped: [[flat-tax-et-imposition]] in the no-tax simplification list (the sentence keeps [[cash-buffer]] and needs no local pointer, since the whole point is that the simplification flatters all seven rules equally), and [[succession-et-transmission]] plus [[sante-et-protection-sociale]] in the "what the last column does not say" callout, whose sentence now reads "a household that wants to leave something behind, or that fears an expensive spell of care at the end of life" with no pointer. Euro amounts kept in euros per glossary 5 (the replay plates freeze them). The seven rule labels were fixed against the EN slugs: Fixed (inflation-adjusted), Flex -10%, Guardrails (GK), Risk-based guardrails, Bounded %, Amortization (ABW), % of portfolio (VPW), same wording in prose, tables and plates. Dictionary: 49 entries across the six replay plates (30 labels and captions plus 19 "moy. X . pire Y" panel stats); audit green. |
| choisir-sa-strategie | choosing-your-strategy | generalize | Decision procedure, structurally neutral, so the pass was mostly local plumbing. Step 1's external coverage "(pension actualisee a sa date, rentes, revenus quasi surs, [[retraite-legale]], [[rentes-et-annuites]])" -> "a pension discounted to its start date, annuities and other near-certain income ([[annuities-and-safety-first]]; for what the US system actually promises, [[us-healthcare-and-social-security]])", the one pointer allowed. Matrix row "rente/rachats de trimestres au plancher" -> "an annuity, or delaying Social Security, up to the floor" (buying pension quarters has no US counterpart). Matrix row "Le legs choisi plutot que residuel ([[succession-et-transmission]])" -> "A bequest you chose, instead of one left over", link dropped. The phase hybrid keeps "your pension" as the generic trigger of the covered phase. "pas ceux d'un article americain" -> "not the ones from somebody else's blog post" (the French aside about US sources is meaningless to a US reader). Closing "Pour aller plus loin" bullet "deroulee pas a pas dans pofo" -> "walked through one control at a time" per glossary 4.2. Links dropped: [[retraite-legale]], [[succession-et-transmission]]. Euro amounts kept in euros in the Claire and Idris worked case, first names unchanged. Dictionary: 39 entries across the two plates (arbre-decision, hierarchie-attention); audit green. |
| managed-futures | managed-futures | generalize (light) | Only substantial France passage is the implementation section "La mise en oeuvre francaise : le point delicat" -> "Buying it cleanly: the delicate part", same slot, same lessons (fee load, replication vs single manager, match the vehicle to the right index, what to avoid, no DIY). European share classes and UCITS wording stripped: "des fonds UCITS de trend systematique" -> "systematic trend funds run by the big quant shops", "les declinaisons UCITS des deux ecoles se deploient" -> "both schools keep showing up in new fund ranges", plus one sentence saying the same exposure has been listed in the US since 2019 with DBMF and KMLM as the best-known examples, pointing once at [[building-it-with-us-etfs]] (no performance claim). The "Logement et fiscalite" paragraph (CTO only, PEA ineligible, assurance-vie, PFU) -> "Where it sits", one neutral sentence: rarely tax-efficient, put it where your own rules tax it most lightly, and it competes with gold for the same shelf ([[gold-in-retirement]]). Intro list item "la mise en oeuvre UCITS pour un Francais" -> "how to buy it cleanly", essentials bullet 5 rewritten the same way. "pofo, par exemple, embarque de tels backcasts" -> "the reconstruction behind this book's data carries trend backcasts built exactly that way" (glossary 4.2). Links dropped: [[flat-tax-et-imposition]]. Euro amounts kept in the worked example. Dictionary: 24 entries across the three plates (trend-smile, trend-annees, trend-correlation); audit green. |
| long-volatility | long-volatility | generalize (light) | Access and vehicle passages neutralized: "ces fonds ne sont tout simplement pas ouverts au particulier europeen, et rien d'equivalent n'existe en UCITS" -> "these funds are simply not open to individual investors, and no fund range you can buy from carries anything equivalent"; "Les fonds UCITS etiquetes volatilite" -> "Most funds labeled volatility"; "produits structures vendus en assurance-vie" -> "sold through insurance and bank channels"; "Un CTO avec acces aux options" -> "A brokerage account with options approval"; "la fiscalite ajoute son frottement" kept generic, no rate; table verdicts "interdit de sejour" -> "banned from a retirement plan", "fermes au particulier europeen" -> "closed to individual investors"; Dragon sleeve "difficile a repliquer honnetement en Europe" -> "hardest part of it to replicate honestly", no geography. One pointer kept at [[building-it-with-us-etfs]] for what a clean vehicle looks like. US-listed names the French already cites (CBOE PPUT, CBOE Eurekahedge Tail Risk, XIV, VIX ETPs, Universa) stay, no performance claim added. No fr-only link to drop (none in the source). Euro amounts kept in the worked example. Universa's +4,144% written as "roughly 42 times its money" because the guard test reads a thousands comma before a percent sign as a French decimal. Dictionary: 26 entries across the two plates (longvol-profil, puts-domines); audit green. |
| global-macro | global-macro | generalize (light) | Locale wording neutralized throughout: title "la diversification des grandes maisons" -> "how the big institutions diversify"; "la conclusion pratique du rentier europeen" -> "the practical conclusion for a retiree"; heading "Le verdict pratique du rentier europeen" -> "The practical verdict for a retiree"; "ses vehicules UCITS controlables contre un indice public" -> "vehicles you can check against a public index", with the one allowed pointer at [[building-it-with-us-etfs]]; "Les fonds ARP au format UCITS" -> "Retail ARP funds" (same in the essentials bullet); "ne sont pas accessibles au particulier europeen" -> "do not take money from individual investors"; "10 % de trend UCITS controle contre le SG Trend" -> "10% trend, checked against the SG Trend". The one French rate is gone: "moins 1,5 % de frais et 31,4 % de fiscalite" -> "minus 1.5% of fees and then tax on what is left". "un simulateur" stays unnamed per glossary 4.2. "La grille de lecture d'une plaquette" -> "Screening an \"alternative\" brochure comes down to five questions", never "reading grid". No fr-only link in the source, so none dropped. Euro amounts kept in the worked example (EUR 1.2M, EUR 45,000 a year). Dictionary: 11 entries for the carry-courbes plate; audit green. |
| primes-d-assurance | insurance-premia | generalize (light) | The euro-investor framing is gone. The callout "Le piege du collateral pour l'investisseur en euros" becomes "The collateral currency trap" and is rewritten as the general lesson: the collateral sits in one currency, total return is that currency's short rate plus the spread, a hedged share class swaps one short rate for the other, so check the collateral currency and the hedge cost. The euro-hedged measurement (1,7 % real from 2013 to 2025, 4,3 % vol, worst month, worst drawdown) is dropped with it; the same essentials bullet now says the headline is mostly the cash leg and the premium is the spread net of about 1.3% of fees. Vehicle wording neutralized: "accessible a l'investisseur europeen" -> "within reach of a fund buyer"; "un fonds UCITS d'arbitrage de fusions" -> "a merger arbitrage fund"; "L'ETF americain ... inaccessible au particulier europeen" -> an index ETF that buys every announced deal, with the one allowed pointer at [[building-it-with-us-etfs]]; "l'ETF cat bond europeen lance fin 2025" -> "the first cat bond ETF, launched at the end of 2025"; "une part couverte en euro" -> "a share class in your own currency" (twice); "fonds euros" dropped from the short-sleeve list. Envelope asides generalized: "un compte-titres ordinaire, ces fonds n'existant ni en PEA ni en assurance-vie grand public" -> "a taxable brokerage account, since these funds are not on the menu of ordinary retirement wrappers". Fr-only link dropped: [[flat-tax-et-imposition]], sentence neutralized to "what is left is still taxed"; no French rate survives. Euro amounts kept where the plate is euro-denominated (the figure caption) and in the merger-spread illustration (EUR 50 / EUR 48 / EUR 2). Dictionary: 11 entries for the sinistres-calendrier plate; audit green. |
| return-stacking | return-stacking | generalize (light) | The section "La mise en oeuvre europeenne, sans se raconter d'histoires" becomes "Buying it without kidding yourself": the UCITS-availability asides go ("L'offre UCITS est jeune et peu fournie", "des declinaisons zone euro (actions de la zone plus futures Bund empiles)", "n'ont pour la plupart pas encore d'equivalent UCITS") and are replaced by a product-age statement, efficient core first and the stacked trend or gold versions newer and still few, with the one allowed pointer at [[building-it-with-us-etfs]]. The third buying discipline swaps wrapper for currency: "Le logement : ces fonds vivent en CTO, sans eligibilite PEA et dans de rares unites de compte" -> "The currency: check which currency the stack is built in and whether the exposure is hedged, because that is what decides the cash leg"; the euro-envelope competition sentence and its fr-only link [[enveloppes-francaises]] are dropped. "une fiscalite penible" -> "a tax treatment that will not thank you", no rate anywhere. The [[etf-ucits-europeens]] link inside the liquidity discipline is not doubled: the single US-ETF pointer sits in the opening of the section. In the astuce the tool is unnamed ("pofo en embarque" -> "this book's data includes them"). WisdomTree stays by name in Going further with "ses declinaisons UCITS" -> "its variants". Euro amounts kept (EUR 100 / EUR 67 / EUR 33, EUR 1M, EUR 38,000 a year). Dictionary: 12 entries for the stacking-expo plate; audit green. |
| primes-de-risque | risk-premia | generalize | Very little France-specific material: the article is where returns come from. The illiquidity row of the audit table dropped the French retail wrapper ("SCPI ou private equity grand public") for "Non-traded real estate fund or retail private equity", and gold's "son prix en euros monte" became "whose price rises" with no currency named. The cash paragraph kept financial repression but lost its fr-only link `inflation-histoire`, the only fr-only link in the source; no pointer to a `us-*` article was needed anywhere, and no French rate appears in the article. Euro kept once, in the covariance image ("one more euro would be worth the most"). Dictionary: 14 entries for the `primes-echelle` plate, including the ranges "4 a 6" -> "4 to 6" and "0,5 a 1" -> "0.5 to 1" (the French decimal comma would otherwise survive translation); audit green. |
| pourquoi-la-diversification-marche | why-diversification-works | generalize | Nothing France-specific in the source: the article is correlation arithmetic, the rebalancing premium and the limits. No French wrapper, no French rate, no fr-only link, and no pointer to a `us-*` article was needed. Two French glosses of English terms disappeared per glossary 2 ("le repas gratuit (free lunch)" and "la recolte de volatilite (volatility harvesting)"), and "la fausse diversification" became diworsification in the section title. Euro kept once, in the time-diversification image ("exposes every euro to a different market"). Dictionary: 25 entries for the `correlation-vol` and `triangle-correlations` plates, including the French decimals "+0,85 a +1,00" and "-0,12 a +0,23" and the percent spacing of "volatilite de chaque actif (20 %)"; audit green. |
| concevoir-un-portefeuille | designing-a-portfolio | generalize | Almost nothing France-specific: the article is a design method (seven questions) and its example. Question 4 dropped the fr-only `[[retraite-legale]]` link and the French "rachat de trimestres" (buying back pension quarters); the sentence now reads "A future pension, Social Security, side income, rent, an annuity", pointing at `[[pensions-and-other-income]]`, so no `us-*` pointer was needed. The monetary-crisis line lost its euro-investor framing ("l'exposition dollar deja presente dans un ETF monde non couvert") and became "reserve-currency exposure (an unhedged global equity fund already carries some)". The worked case keeps its French first names (Lea and Marc) and its euro amount (EUR 1.3M). Dictionary: 28 entries for the `risques-briques` plate; audit green. |
| allocation-actions-obligations | stock-bond-allocation | generalize | The article is mostly market mechanics, so the France-specific content was thin. Dial 2 ("la couverture du plancher") dropped the fr-only `retraite-legale` link for "a pension, Social Security or an annuity". The bond bestiary was reframed around a US reader without losing the euro examples: the core-sovereign entry now leads with the US Treasury (Bund and gilt as the euro and sterling equivalents), the spread-sovereign entry says "inside a currency union" before naming the BTP, and the linkers entry is "TIPS in the US, and their equivalents elsewhere" instead of the French OATi and OAT-i. The ETF paragraph asks for "a share class hedged into your own currency" instead of a euro-hedged one. The zero-coupon paragraph lost the UCITS availability sentence and the French stripped-OAT aside: long STRIPS stay as the example, and very long duration is bought through a 25-year-and-up government bond ETF. The simulator name dropped per glossary 4.2 ("a simulator that reads the portfolio holding by holding"); cFIREsim and FICalc kept, since the French names them. Euro amounts kept per glossary 5 (EUR 1.5M, EUR 52,000 a year, EUR 50,000 twelve years out). Dictionary: 57 entries for the five plates (allocation-plateau, duration-choc, obligations-rendements, obligations-regimes, duration-vehicules); audit green. |
| glidepaths | glidepaths | generalize | Nearly locale-free to start with. The declining-leg paragraph dropped the fr-only `enveloppes-francaises` link: "C'est aussi fiscalement optimal" became "without triggering a tax bill on the way", folded into the sentence about redirecting contributions. "Le retraité français dont la pension tombe dans 15 ans" became "anyone whose pension or Social Security starts in fifteen years", the article's only local pointer. The simulator stays unnamed per glossary 4.2 ("not every simulator can move the allocation over the life of a plan"), with the real English control name quoted once: "Rising-equity glidepath (bond tent)". Euro amounts kept per glossary 5 (EUR 2,800 a month). Vocabulary: versant montant/descendant -> the rising/declining leg, allocation de croisière -> your long-run allocation, la pente -> the path ("stay on the path", "the path suspended at the first crash"). Dictionary: 22 entries for the two plates (bond-tent, tente-transfert), one keyed (`bond-tent|départ`, since "départ" alone is too generic to translate globally); audit green. |
| portefeuilles-tous-temps | all-weather-portfolios | generalize | Nearly locale-free: the family, its numbers and its criticisms. Two euro-market asides generalized. The Dragon's "la poche long volatility n'existe pas en format UCITS accessible" became "there is no cheap, accessible fund for the long volatility sleeve" (same in the `tous-temps-saisons` plate legend), and the A/B recipe's "les briques disponibles" now points once at `building-it-with-us-etfs` for the shopping list. No French rate, no wrapper, no fr-only link, so none dropped. The example keeps its euros (EUR 1.5M, EUR 51,000 a year, EUR 3.1M against EUR 2.6M) and its Vanguard corridor. The simulator stays unnamed per glossary 4.2 ("ruine centrale" -> central-case failure probability, "stress" -> sequence stress, "décennie perdue" -> lost decade). Vocabulary: tous-temps -> all-weather (All Weather capitalized for Bridgewater's own), curseur -> dial, écart à l'indice -> the gap to the index, poche de régimes -> regime sleeve, demi-tous-temps -> half all-weather. Dictionary: 54 entries for the three plates (`tous-temps-saisons`, `tous-temps-echange`, `tous-temps-ecart`), three shared bare keys across them; audit green. |
| actifs-defensifs | defensive-assets | generalize | The one France-specific candidate, the "fonds euros", became the US analogue in category terms: the cash paragraph now reads "Cash (money market funds, T-bills, stable value)" and its second half describes a stable value fund inside a workplace plan (smoothed bond yield, principal protected by an insurance wrapper) with no product name and no yield figure. Same swap in the table row ("Cash, money market") and in the worked example, where the 35% starting sleeve is a stable value fund whose crediting rate follows market rates years late, which is what makes the inflation and confidence rows wide open. "OATi/TIPS" -> "linkers, TIPS"; "la brique la plus sous-detenue des portefeuilles francais" -> "the most under-owned block in retirement portfolios"; "le portefeuille francais type, actions plus fonds euros" -> "the typical retail portfolio, stocks plus a savings account". The euro-investor currency aside was neutralized both ways: an unhedged world equity fund is a free dollar position for anyone whose home currency is not the dollar, and the cushion is already there for a dollar-based investor (same treatment in the essentials bullet and in the "USD non couvert" table cell, now "safe-haven currency"). One fr-only link dropped: [[enveloppes-francaises]], at the end of the cash paragraph. One pointer added: [[building-it-with-us-etfs]] in the A/B recipe, for the shopping list. The simulator stays unnamed (glossary 4.2). Vocabulary: cahier des charges -> the spec, titulaire/doublure -> starter/backup, regime tueur -> killer regime, cout de portage -> carry cost, faux defensifs -> false defensives, socle -> bedrock, purgatoire -> purgatory / the wilderness, traverses -> ridden out. No figure block in the article, so no dictionary entry; audit green. |
| faux-actifs-defensifs | false-defensive-assets | generalize | The French products became US category analogues: "les SIIC en France" dropped from the REIT paragraph (glossary: listed property (REITs)); "les fonds evergreen et les ELTIF" -> "evergreen and interval funds", the retail wrappers a US reader meets, in both the private equity section and the "two traps of the moment" callout; "une volatilite de fonds euros" -> "the volatility of a stable value fund". No French rate, no tax aside and no envelope appears in the article, so nothing else needed neutralizing. No fr-only link dropped: all twelve wiki-links map through plannedEN. The two US fund names the French already cites (QYLD, JEPI) were kept as the category illustration they are, with "leurs clones europeens" generalized to "the clones that followed". The simulator is not named. Vocabulary: bulletin -> report card, le test infaillible -> the infallible test, poche de rendement -> the return-seeking sleeve, poche revenu -> the "income" sleeve, contre-galerie -> the gallery of the fakes, volatilite blanchie -> laundered volatility, risk-on deguise -> risk-on in disguise. No figure block in the article, so no dictionary entry; audit green. |
| or-en-retrait | gold-in-retirement | adapt | The section "La pratique francaise : vehicules et fiscalite" was rewritten in its slot as "Owning gold in the US: funds, coins, and the tax bill": physically backed gold ETFs (GLD and IAU named as the best known, no performance claim, expense ratios compared not quoted) replace the physical ETC paragraph; the French PFU/CTO/PEA/assurance-vie block became the IRS collectibles mechanism (gold and physically backed grantor trusts taxed as collectibles, long-term rate capped at 28% against 20% at the top for ordinary long-term gains, IRS Topic 409; short-term at ordinary rates; the 3.8% NIIT on top), plus the grantor trust selling metal to pay its fee and lowering basis; a new paragraph on the tax-advantaged account (no collectibles rate, IRA gold ETF trivial, metal only under 26 U.S.C. 408(m)(3) with a qualified trustee, home storage is a distribution); the Napoleon/Souverain/11.5% forfait/22-year exemption paragraph became US coins and bars from a dealer, premium over spot and dealer bid, storage and homeowner-policy insurance caps, keep invoices for basis; a new paperwork paragraph (broker 1099-B, dealer 1099-B only for certain metals and quantities, Form 8300 over $10,000 cash) closed by the dated-accuracy caveat. Sources checked live: irs.gov Topic 409, irs.gov Schedule D instructions (28% Rate Gain Worksheet), irs.gov Net Investment Income Tax, irs.gov Form 8300, 26 U.S.C. 408(m) on law.cornell.edu, SPDR Gold Trust prospectus on sec.gov for the grantor trust reading. Generalized elsewhere: "le portefeuille francais type" -> "a typical retail portfolio", "la hausse du dollar a joue pour l'investisseur en euros" -> "for anyone whose spending currency was not the dollar", "honorable en euros" (2022) -> "a draw in dollars", the worked example's "ETC physique, CTO" -> "a physically backed fund, taxable account" with the euro amounts kept. Fr-only links dropped: [[enveloppes-francaises]] and [[flat-tax-et-imposition]], both inside the rewritten section, replaced by one pointer to [[us-taxes-in-the-withdrawal-phase]]; [[building-it-with-us-etfs]] added once on the fund choice. figureDict: nine entries for the or-decennies plate, including the euro caveat label neutralized to "Outside the dollar, the experience differs: the currency move adds to the price of gold."; audit green. |
| obligations-en-retrait | bonds-in-retirement | generalize | The euro market details became locale-neutral: "l'aggregate euro a perdu 17 %" -> "broad investment grade indexes lost double digits", the ECB path "taux passes de -0,5 % a +3 %" -> "policy rates off the floor and up several points", the grid rows "Etat euro intermediaire / long" and "Linkers euro courts" -> Intermediate Treasuries, Long Treasuries, Short TIPS, and decision 4 "euro, ou couverte" -> "home, or hedged" (an unhedged foreign bond, a hedged share class, no product name). The section "Le fonds euros : l'exception francaise" was generalized in its slot as "The stable value fund, in its right place": same lesson, category terms only (a managed bond portfolio wrapped by an insurer, principal guarantee, yield smoothed out of reserves, credited rate lagging market rates by years, no crisis convexity, guarantee resting on the insurer and on contract terms that can restrict transfers, up to about half the sleeve, money market funds and T-bills where no such fund is offered); the loi Sapin 2 sentence became the contract-terms clause, and the French 2022-2024 served-yield figures became "lagged by a point or two". "Avec la fiscalite en moins" (high yield) -> "taxed as ordinary income at that". Fr-only link dropped: [[enveloppes-francaises]] in the stable value paragraph, neutralized to "in a tax-deferred account" with no pointer added. figureDict: eight entries for the correl-sign plate; audit green. |
| obligations-indexees | inflation-linked-bonds | adapt | The section "La pratique francaise" was rewritten in its slot as "TIPS and I bonds in practice", same length order of magnitude: TIPS bought at auction through TreasuryDirect or a broker (5, 10, 30-year terms, $100 minimum) and on the secondary market, broad and short-maturity TIPS ETFs (TIP, SCHP, VTIP named as the best-known, no performance claim); a new paragraph on the ladder a US individual CAN build rung by rung, which inverts the French point of the missing retail counter, with the maturity hole of the 2030s and the double-up fix; one paragraph on Series I savings bonds (composite rate = fixed + inflation reset every six months, $10,000 a year per Social Security number, twelve-month lockup, three months of interest given up before five years, 30-year term, state and local exemption, federal tax deferrable) carrying a dated-accuracy caveat; and the tax paragraph rebuilt around phantom income (yearly taxable accretion, no US wrapper defers it, so the sleeve belongs in a tax-advantaged account) plus the state-tax exemption. Sources checked live at writing time: treasurydirect.gov I bonds and TIPS pages (both reachable, terms as stated). Generalized elsewhere: "OATi et OAT€i francaises, TIPS americains" -> TIPS with linkers as the trade word, the reference index list -> CPI-U, the retail-distribution reason for low ownership -> "nobody's default, and a real yield reads like a typo", the euro-vs-US ladder comparison -> a real-yield sensitivity ("at 1% real the same ladder funds about 3.9%"), the worked example's "faute d'echelle directe euro" dropped, and the Agence France Tresor reference replaced by treasurydirect.gov. Euro amounts kept per the glossary. Fr-only links dropped: [[flat-tax-et-imposition]] and [[enveloppes-francaises]], both in the tax paragraph, replaced by the single pointer [[us-accounts-and-account-order]]; [[etf-ucits-europeens]] became [[building-it-with-us-etfs]]. figureDict: twelve entries for the linkers-echelle plate (the "zone euro recente" marker neutralized to "1% real"); audit green. One test tightened: TestHandlerEnglishIndexSkipsEmptyParts matched the bare part title as a substring, and the title "Inflation-linked bonds" made the empty part "Inflation" look rendered; it now matches the rendered `<h2 id=...>` heading. |
| facteurs-fama-french | factors-in-retirement | generalize | The UCITS implementation section became "Buying it: a short menu, and three checks": the named European share classes (SPDR Small Value Weighted, MSCI Enhanced Value, World Momentum, World multifactor) became fund categories, with the Avantis and Dimensional ranges named once as the US reference and no performance claim, and "le paysage europeen" became "how deep the menu goes depends on where you buy your funds". The account sentence ("c'est le CTO pour la plupart, certains fonds eligibles au PEA") became "decided by the account and not by the factor" with a single pointer to [[building-it-with-us-etfs]]. "SCV Europe" became international small-cap value in both the sizing paragraph and the worked example. Fr-only link dropped: [[enveloppes-francaises]]. Euro amounts kept (EUR 1.5M, EUR 51,000 a year). figureDict: fifteen entries for the scv-ecart-10ans plate (title, annotations, the four point labels, the two footnotes); audit green. |
| diversification-internationale | international-diversification | adapt | Two sections rewritten for a US reader in the same slots. Home bias: the French/euro home bias (CAC, "actions France", the PEA nudge toward Europe) became the US home bias, US stocks around 60% of world market capitalization against roughly 80% held at home, the country that won the last century as the sharpest form of the world sample's warning, with the French, Japanese and German fates kept as evidence; the wrapper cause became the workplace-menu cause. Currency: the euro-investor crisis cushion inverts, so the section now says plainly that the dollar is the haven, that unhedged foreign stocks fall a little harder in dollars in a panic, and that the case for leaving equities unhedged rests on cost, mean reversion and the small volatility contribution; the 2008 MSCI World euro-versus-dollar fact and the 2000 to 2003 fact are kept and reread from the dollar side, the bond-side rule unchanged. Target in one fund: VT named once as the best-known US example, no performance claim, one pointer to [[building-it-with-us-etfs]]. Karim's worked example moved to a 401(k) plus IRA plus taxable account with a 75% US sleeve, EUR amounts kept. Fr-only links dropped: [[retraite-legale]], [[enveloppes-francaises]], [[flat-tax-et-imposition]] (tax sentence neutralized, no French rate survives). figureDict: no entries, the article carries no plate; audit green. |
| etf-ucits-europeens | building-it-with-us-etfs | adapt (heavy) | European practice end to end, so the whole article was rewritten for a US buyer in the same four sections, same callout rhythm, same job (the shopping list and how to sell cleanly). Section 1: the UCITS label and its four choices became the 1940 Act fund plus creation and redemption in kind, then ETF versus mutual fund, the distribution you cannot switch off (no accumulating class here), physical replication with the three blocks that are not 1940 Act funds (gold as a grantor trust taxed at the 28% collectibles rate, K-1 partnerships, the managed-futures Cayman subsidiary), and sponsor / fee / securities lending; the Irish-domicile arbitrage became "domicile is not your question", with the foreign tax credit as its one echo and the account as the real choice. Section 2: the cost chain rebuilt on US links (expense ratio, internal drift, spread, zero commissions, the 1% advisory layer), tracking difference kept as the true measure, 0.10 to 0.25% a year all in against 1.5 to 2% advised. Section 3: the UCITS table became a US table with a category, best-known tickers, a typical fee range, the natural account and what to check (VT, VTI + VXUS, the 401(k) menu row replacing the PEA row, AVUV/VBR, VGIT/BND, VGLT/EDV, VTIP/SCHP/TIP, SGOV, IAU/GLD, DBMF/KMLM), every ticker verified live against fund data at writing time; the envelope callout became account placement pointing at [[us-accounts-and-account-order]]. Section 4: selling cleanly rewritten as specific identification versus the FIFO default, long-term versus short-term and the wide 0% band, the 30-day wash-sale window (and the loss lost for good inside an IRA), T+1 settlement, limit orders in thin funds, and RMDs at 73 framing the account order, with one pointer to [[us-taxes-in-the-withdrawal-phase]]. Worked example moved to taxable / Roth / pre-tax with the same seven holdings and EUR amounts kept. Sources checked live: investor.gov, IRS Publication 550 and Topic 409, the IRS RMD FAQ, SEC T+1 bulletin, sponsor and fund data pages. Fr-only links dropped: [[enveloppes-francaises]], [[flat-tax-et-imposition]] (both appeared three times; no French rate or envelope survives). figureDict: no entries, the article carries no plate; audit green. |
| cash-buffer | cash-buffer | generalize | No French product survives. The placement bullet "Le matelas français idéal n'est pas le compte courant. C'est le fonds euros (...) ou le monétaire €STR en CTO, plus les livrets réglementés pour la première tranche" becomes "A checking account is not a buffer. A Treasury money market fund or a short T-bill ladder is, with a high-yield savings account for the first few months of spending and a stable value fund if a workplace plan offers one", keeping the halved-opportunity-cost claim and the no-duration rule. Elsewhere fonds euros becomes "a money market fund" (three times: the intro roadmap, the anti-panic paragraph, the worked example) and "~0,5-1 %" becomes "about 0.5 to 1%" as the buffer's real return assumption. The governance instruction "on vit sur l'assurance-vie X" becomes "we live off the account at X". "corridor Vanguard" becomes "Vanguard-style dynamic spending". Fr-only link dropped: [[enveloppes-francaises]] in the placement bullet, replaced by the single allowed pointer [[us-accounts-and-account-order]] on which account holds the buffer. The tool stays unnamed: "Dans pofo :" becomes "In the simulator:" and names the three real controls ("Buffer (years of spending)", "Buffer real return", "Buffer refill stops in year"). Euro amounts kept (EUR 1.5M, EUR 150k, EUR 1.35M, EUR 130k, EUR 52,000 a year). figureDict: 23 entries for the two plates (`buffer-flat`, `traversees-matelas`), the four wrapped lines of the traversees annotation re-split to keep the line lengths; audit green. |
| cash-ameliore | enhanced-cash | adapt | The ladder of cash tiers is rewritten as the US one, slot for slot: livrets become the FDIC-insured savings account, fonds euros becomes the stable value fund inside a workplace plan (smoothed credited rate, insurance wrap, equity wash rule, does not survive a rollover), the ETF monétaire becomes the Treasury money market fund plus T-bills and short Treasury ETFs (with the state-tax exemption), the obligataire ultra-court becomes the ultra-short bond fund, the CLO AAA section translates as is (generic mechanics) with JAAA named once as the best-known US-listed example and the fee stated at the category level. Euro yield figures specific to fonds euros or the ESTR become the mechanism or an order of magnitude anchored to the T-bill rate; the worked example keeps its euro amounts but runs at a 4% short rate with the CLO third as the only source of the extra EUR 400. The CLO measured paragraph now leads with the dollar record (133 bp over three-month T-bills since 2020-10, worst drawdown 2.60% in 2022) and keeps the euro-listed cohort only as the uniformity cross-check. The EU risk-retention paragraph becomes the US one: open-market CLO managers were exempted by LSTA v. SEC (D.C. Circuit, 2018), so no skin in the game by law and the quarterly tests do the aligning. The 'ETF a millesimes vs fonds date francais' table becomes 'defined-maturity ETF vs the ladder you build yourself'; the French ranges become iBonds and BulletShares with their US families (Treasury, TIPS, corporate, high yield, muni), vintages out to about 2035 and per-segment fees, so the 'no high-yield vintage' point is replaced by the TIPS-vintage one; the assurance-vie-versus-CTA arbitrage becomes the US placement lesson (interest taxed as ordinary income yearly, Treasury interest exempt from state and local tax, corporate and high yield in a tax-advantaged account, munis at high brackets) with the single allowed pointer [[us-taxes-in-the-withdrawal-phase]]; the capitalisant/distribuant execution detail becomes the monthly-income one. Euro curve quotes replaced by the Treasury par curve on 2026-07-31 (4.08 / 4.45 / 4.75) and 2023-07-31 (5.37 / 3.97), from treasury.gov. Dated-accuracy caveat added after the measured CLO paragraph. Fr-only links dropped: [[flat-tax-et-imposition]] (three times) and [[enveloppes-francaises]] (twice), each sentence rewritten around the US placement rule. Sources checked: treasury.gov daily par yield curve, fdic.gov (250,000 limit), invesco.com BulletShares page, ishares.com iBonds range, stablevalue.org (equity wash), LSTA v. SEC (D.C. Circuit 2018). figureDict: 23 entries for `echelle-du-cash` (7 of them the identical bp labels); audit green. |
| strategie-buckets | the-bucket-strategy | generalize | Little France-specific material: the article is the case against buckets. The worked example's "matelas de 100 k€ en fonds euros" became "a EUR 100k buffer in a money market fund", and the bond sleeve kept its linker ladder and its 5 to 8 year fund, both locale-neutral. "Une bonne partie du conseil américain" became "much of the advice industry", the reader now being inside it. No French rate in the source, no fr-only link, so none dropped; the simulator is unnamed in the closing bullets per glossary 4.2. Dictionary: the 12 labels of the `buckets-allocation` plate; audit green. |
| echelle-obligataire | bond-ladders | adapt | The section "La pratique francaise : les contournements du guichet absent" INVERTS for a US reader and was rewritten in its slot as "Building it, rung by rung": a US individual can buy the real thing at retail, so the French workaround table (fonds euros, fonds a echeance Etat/IG, OAT and OAT€i in direct, rolled short-linker ETFs, and the missing indexed target-maturity fund) became a US vehicle table (T-bills and money market for years 1-2, Treasury notes and bonds, TIPS, defined-maturity Treasury/corporate ETFs, defined-maturity TIPS ETFs, brokered CDs) plus five paragraphs: which rung for which job, buying at auction against the secondary market (noncompetitive bid, $100 increments, TreasuryDirect holdings must be transferred to a broker before sale, dated-accuracy caveat), the TIPS maturity gap in the 2030s and the shoulder-year fix, defined-maturity ETFs at category level (iShares iBonds, Invesco BulletShares, including the indexed line that France lacks), brokered CDs as the FDIC-insured cousin in one sentence, and placement (Treasury interest exempt from state and local tax, TIPS accretion taxed as it accrues, one pointer to `us-accounts-and-account-order`). Facts checked on treasurydirect.gov in August 2026 (terms, increments, auction calendar, sale requires transfer) and on the iShares/Invesco lineups. The "fonds euros en barreau court" paragraph folded into the cash-buffer sentence of the same section; the worked example keeps every euro figure and swaps its vehicles for the US ones (T-bills, Treasury rungs and defined-maturity funds, TIPS for years 9 to 13). Elsewhere: the pension bridge became "your pension or Social Security" with one pointer to `us-healthcare-and-social-security`, the linker-ladder bullet dropped the euro-cousin aside for the pilot's TIPS numbers, and the "spreads OTC" trap became bid-ask spreads. No fr-only link in the source, so none dropped; the simulator is unnamed per glossary 4.2. No figure block, so no dictionary entry; audit green by construction. |
| recharger-ou-pas | refilling-the-buffer | generalize | Only two France-specific spots, both vehicle names: "en fonds euros" in the model-parameters paragraph and in the postcard example became "your money market fund" / "in a money market fund", matching the vehicle language of `cash-buffer`. No `## La pratique francaise` section in the source and no fr-only wiki-link, so none dropped; the seventeen links all map through `plannedEN`. The simulator stays unnamed and the three buffer controls are quoted with their real UI strings ("Buffer (years of spending)", "Buffer real return", "Buffer refill stops in year"). No figure block, so no dictionary entry; `figure-audit.sh en` green. |
| immobilier-en-retrait | real-estate-in-retirement | adapt | French property taxation rewritten into US mechanics in the same slots: rental income as ordinary income after expenses, depreciation over 27.5 years as the shelter, unrecaptured section 1250 gain at up to 25% and the 3.8% surtax at the sale (IRS Topic 409), the primary-residence exclusion at USD 250,000 / USD 500,000 with the two-of-five test (IRS Topic 701), property tax deductible only inside the SALT cap (USD 40,000, IRS Topic 503). The numeric cascade keeps its euro amounts and its structure; its tax step became an explicit assumption ("call the friction on the flow 45%", the same 45.2% the plate freezes). The three last-resort taps became downsizing, the HECM reverse mortgage (FHA-insured, 62+, HUD counseling, verified on hud.gov) and a home equity line of credit; viager and demembrement dropped to one clause on life estates and remainder interests. SCPI became non-traded property funds and interval funds, with the 2022-2023 non-traded REIT gating in place of the French redemption queues, plus one sentence pointing listed REITs at [[false-defensive-assets]]. The mortgage argument keeps ERN Part 21 and is reframed on the 30-year fixed (no repricing, refinance down but never up). Dropped fr-only links: [[inflation-histoire]], [[flat-tax-et-imposition]] (twice) and [[enveloppes-francaises]], with the sentences neutralized; one [[us-taxes-in-the-withdrawal-phase]] pointer added. Dictionary: 30 entries for `immobilier-net-net`, the tax labels neutralized ("foncier nu" -> "taxed on the flow", "LMNP au reel" -> "sheltered by depreciation", the two French-rate footnotes rewritten as the 45.2% friction and as depreciation erasing the base); `immobilier-net-net|brut` keyed because "brut" already means "raw" elsewhere. `figure-audit.sh en` green. |
| levier-et-marges | leverage-and-margin | generalize | The lombard/assurance-vie pair generalized into US categories: "credit lombard" became a securities-backed line of credit throughout (advance rate kept, no bank names, no French rate), "avance d'assurance-vie" became a loan against a cash-value policy (no taxable sale while the policy stays in force). The instrument table lost its French pricing ("Emprunts d'Etat + 1 point" -> "the rate written into the contract"; "Monetaire + 0,5 a 1,5 point" -> "the short rate plus a spread"; brokerage margin -> "the broker's margin rate, a spread over the policy rate") and its "assurance deces" clause on the kept mortgage. One US mechanism added in a clause: Reg T lets you borrow half a position to open it, but the maintenance level is what sells you out. "L'offre UCITS existe, avec les Efficient Core globaux" became "a handful of these funds are listed in the US, and the shelf keeps growing", with the sole [[building-it-with-us-etfs]] pointer; the second traceability link to the same article was dropped. "Fonds euros" in the worked example became cash. Simulator and visualizer stay unnamed (a portfolio visualizer, a decumulation engine, [[under-the-hood]]). No fr-only links in the source, so none dropped. No figures in this article: zero dictionary entries; `figure-audit.sh en` green. |
| anarkulova-cederburg | anarkulova-cederburg | generalize | Pfau's worst-country aside stops singling France out (Japan, Italy and France "among others"); no fr-only link in the article; the simulator stays unnamed in the closing bullet, which now points at `under-the-hood`. |
| valorisations-et-cape | valuations-and-cape | generalize | The European-investor framing of the last section becomes "anyone holding a global portfolio" and its `etf-ucits-europeens` link is dropped; "votre ETF World" becomes "your world equity fund"; the simulator stays unnamed and its "Anchor return to today's valuation (CAPE)" control is named instead of the tool, pointing at `under-the-hood`. |
| horizon-et-esperance-de-vie | horizon-and-life-expectancy | generalize | The INSEE/INED life tables and their numbers are kept as computed (the plates are frozen on them), with one aside that US Social Security period tables run a little lower and nothing in the reasoning moves; the SSA/SOA tables are named in "Going further". Pension-age references become "when your pensions start", and `retraite-legale` is dropped, with the one pointer to `us-healthcare-and-social-security` on the state pension as a life annuity; `succession-et-transmission` dropped ("to leave something behind"); "plans francais realistes" becomes "plenty of realistic plans"; "chaque trimestre valide compte" dropped; the simulator stays unnamed in the closing bullet, which points at `under-the-hood`. |
| serie-ern | the-ern-series | generalize | The "Lire ERN avec les bons filtres" section is rewritten: the geographic filter now says that a US reader takes ERN straight (data, tax chapters, Social Security, 401(k)/IRA all native, one pointer to `us-taxes-in-the-withdrawal-phase`) and that only the sample optimism survives for them, while a non-US reader still filters currency, home market, and the tax and pension chapters; the posture and perfectionism filters translate unchanged. `enveloppes-francaises` and `taxe-puma` dropped from that section, `retraite-legale` dropped from the asset-review block ("Annuities and Social Security", no "to be adapted to the French system"); the closing bullet and the "where to start" tip lose "le cadre francais" for "the tax and account frame"; the simulator stays unnamed (the CAPE anchor is named instead). No figures in this article. |

#### Triage of the general articles (settled 2026-08-16)

A sweep of the non-tax articles for France-specific density (term counts,
then a reading of the sections) yields the list below, settled the same day.
Three outcomes per article: fr-only (marker on the French file, no English
counterpart, dropped from `plannedEN`), adapt (one or more sections rewritten
for the US reader; the rest translates), or generalize (the default, see
`docs/fire-book-en-translation-brief.md`, section 3). An article absent from
this table is generalized.

| FR slug | EN slug | Decision | Why |
|---|---|---|---|
| cas-types | (none) | fr-only | The three households are built on PEA / assurance-vie / CTO / PUMa / pension quarters; a US version would be three new plans, not a translation. The English "In practice" part has six articles. |
| inflation-histoire | (none) | fr-only | The spine is French monetary history 1914-2025 (ruined rentiers, post-war financial repression, 1974-81); a US reader would expect the dollar's 1913-2025, which is another article. English wiki-links to it are neutralized. |
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
- Restated 2026-08-16 as the campaign's first axis: idiomatic US English
  that is clear, fluid, engaging and pleasant to read, in short sentences
  with a good rhythm, and free of gallicisms (calqued word order, French
  connectives and scaffolding, nouns where English wants verbs). The
  translator's brief spells out the banned patterns and the reread-aloud
  discipline; no test catches bad English, the line review does.
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
