# The FIRE book: design

Status: v1 COMPLETE and accepted (2026-07-13): all 79 planned articles
written. EXTENSION 2026-07-17 (after an external review of the TOC):
7 new articles, bringing the plan to 86. A new category "Les actifs
alternatifs" (a must-have of the extension) hosts long-volatility, global-macro and
return-stacking, and managed-futures moves into it from "Le portefeuille
de retrait" (slug and URL unchanged; categories are only index grouping).
Two financial-theory articles (primes-de-risque, pourquoi-la-diversification-
marche) join "Le portefeuille de retrait"; decision theory
(decider-sous-incertitude) and the mathematical anatomy of the 4 % rule
(les-maths-du-4-pourcent) join "La science du retrait". Rejected from the
review: actuarial chapters (already covered by rentes-et-annuites and
horizon-et-esperance-de-vie), GARCH/copulas/ML/Bayesian-updating chapters
(too exotic for the book's audience; at most passing mentions in the
modelling articles), and reproducing the reviewer's table of contents.
The 7 extension articles are illustrated by second-generation figures
("plates", figures_v2.go): left-aligned serif title + letterspaced kicker,
Instrument Sans labels, Spline Sans Mono numbers, hairline grids, rounded
data ends, CVD-validated series trio (amber #b4783c / blue #3a6db4 / red
#c0655b, + green #2f9068 for stacked segments with gaps and direct labels).
New figures should follow this system, not the first-batch style (the v1
look was judged amateur); always screenshot through the real page CSS (headless
Chrome on a harness that inlines fonts.css/theme.css) and check for label
overlaps before committing (a headless-Chrome pass over the plate's `<text>`
nodes, comparing `getComputedTextLength` against the viewBox, catches the
silent right-edge clipping the eye misses). Where a plate carries numbers that
this repository can compute, compute them: `mc-entrees-vs-tirages`
(2026-07-24, monte-carlo-forces-faiblesses) plots ruin probabilities produced
by `decumul.Plan` over a `scenario.ParametricSource` (400 000 paths) rather
than illustrative values, and states the plan and the model in a footnote.
The withdrawal-rule plates of the same day (`corridor-1966`, `corridor-borne`,
`gk-cascade-1966`, `abw-1966`, `vpw-table`) go further and share one ground:
the 1966 US 60/40 real sequence rebuilt from the bundled reference series
(SP500-USD + TREASURY-INT-USD deflated by the `^CPI-US` snapshot), run through
`decumul.Plan.RunPath` one rule at a time with 1 M EUR and no tax, so the
plates compare lives and not model assumptions. The scratch program that
produced the series is not kept; the series are inlined in the plate with a
comment stating exactly how to reproduce them. Decision: the 12 pre-depth-bar articles (the early batches of
parts I-II) need NO deepening pass; they stand as-is. Remaining work: the
later English translation, and continuous upkeep of the dated French
tax/social chapters. The ledger below tracks per-article state.

Illustration campaign 2026-07-30, 29 new v2 plates. Selected from the two
figure backlogs opened at the close of the line-by-line review
(`docs/fire-book-illustrations-*`), priority to ideas whose data was already
bundled and to the ones carrying a thesis the article could only assert. One
plate per commit, each with its guard test. The rule that made the batch worth
its cost: WHERE THE REPOSITORY CAN COMPUTE THE NUMBER, THE PLATE COMPUTES IT
AND THE ARTICLE MUST AGREE WITH THE ENGINE. Ten articles lost a number or a
claim to that rule, and those corrections matter more than the drawings:

  - enveloppes-francaises treated the 4 600 / 9 200 EUR life-insurance
    allowance as a near-exemption. It only erases income tax; social levies
    fall on the whole redeemed gain. The worked example's friction went from
    5 % to 7,8 % and its capitalised gap from 240 k to 180 k EUR.
  - valorisations-et-cape claimed the CAPE explains 40 to 60 % of the variance
    of ten-year real returns. Measured on the earnings yield it is 0,29, and
    0,28 on Shiller's own series from 1881, so it is not a window artefact.
  - horizon-et-esperance-de-vie claimed mortality weighting cuts ruin to a half
    or a third. For a couple retiring at 47 it removes a fifth (17,7 % gross
    against 14,1 % lived); the discount is a property of the reader's age, not
    of the plan.
  - erreurs-classiques-fire announced its ten sections were ordered by cost.
    Measured, the two dearest are sections 6 and 7, together outweighing the
    other four chiffrable ones combined.
  - the book-wide "les traversées durent 2 à 7 ans" anchor was wrong at both
    ends (16 months to more than ten years, median 32), corrected in
    cash-buffer, regimes-de-marche and strategie-buckets.
  - also corrected: the French monetary leg of 1973-1985 (+6,6 % real, not
    "perd peu") and the ranking of inflation's victims in inflation-histoire;
    the all-weather `science` block in portefeuilles-tous-temps; gold's 2022 in
    dollars and its volatility band in or-en-retrait; "des rendements de long
    terme comparables" for 1966 against 1982 in sequence-des-rendements;
    "mourir à trois fois sa mise" in psychologie-du-retrait; and the reload
    doubling claim in succession-et-transmission.

Three plates carry an explicitly editorial object, labelled as such on the
plate itself rather than passed off as measurement: the link thickness of
`risques-briques`, the tolerance zone of `coupe-exigee-tenable`, and the
season attribution of `tous-temps-saisons`. Two traps cost a round trip each
and are worth knowing: a raw ampersand in SVG text breaks the EPUB XHTML guard
(write `S&amp;amp;P`), and a Monte Carlo recipe must freeze its worker count,
since DrawPaths splits per worker and NumCPU makes a plate machine-dependent.

Second wave 2026-07-31: eleven more plates and ten reference tables. The
tables come from the backlog ideas costed C ("a table beats a drawing") and
were held to one rule, a table only earns its place if it REPLACES prose, so
each shipped with the paragraph it makes redundant cut down to its point. Two
of them found contradictions rather than merely tidying: the bond pocket's
worked example sized its indexed sleeve at 19 % against its own article's
25-50 % doctrine, and the lexique's numeric memo is now a book-wide guard, its
test re-reading, for every anchor, the chapter that owns it.

The same engine-first rule kept correcting the prose, and this wave it reached
published sources too:

  - long-volatility described the PPUT index as costing 1,5 to 2,5 points a
    year for a drawdown "à peine amélioré". Both halves were wrong: Cboe and
    Wilshire publish 6,6 % against 9,8 % over 1986-2018 (3,2 points) and the
    drawdown does improve, from −51 % to −38,9 %. Israelov's point is that the
    cut is not worth the return given up, which is what the plate shows.
  - rendements-attendus stopped its Morningstar staircase at 3,7 %. The
    December 2025 edition publishes 3,9 %, and reading the report itself
    showed the rise is a METHOD change, not a market move (it states the
    previous method would have given 3,6 %). The same series was stale in
    guardrails-morningstar and was corrected there too.
  - la-regle-des-4-pourcents claimed "la moyenne des millésimes supporte plus
    de 6 %". Measured over 66 vintages the mean is 6,05 % and the median
    5,89 %, so the sentence now speaks of the median vintage.
  - pourquoi-la-diversification-marche put its equity block at 0,85-0,95; it
    measures 0,85-1,00, with the S&P 500 against the whole US market at
    0,9955, the same asset at rounding.

Three findings were recorded rather than drawn, which is the other half of the
discipline. The glidepath plate could not test the article's CAPE thesis at
all, since no vintage in the available window started above a CAPE of 25, so
the plate marks vintage hardness instead and a test will fail the day a longer
sample makes the valuation highlight possible. The 4 % floor is not one freak
vintage but six consecutive ones, 1964 to 1969. And the put-protection point
does not merely sit inside the equity/bond cloud, it is strictly dominated by
every rung from 50/50 to 75/25.

Two construction habits are worth keeping. When a new plate answers a question
a shipped plate already touches, reuse the shipped construction and change only
what the question requires (puts-domines keeps tous-temps-echange's ladder and
moves only the window and the basis, so the two cannot contradict each other).
And anchor a reconstruction on a published number when one exists: that ladder's
100 % equity rung reproduces Cboe's published S&P 500 return and drawdown to the
decimal over the same 390 months, which is what puts a published point and a
recomputed cloud on the same footing.

A rendering bug surfaced and was fixed in pkg/bookmd: the pipe-table splitter
cut rows on a plain split over "|", so a labelled wiki-link ([[slug|Label]])
tore its cell in two. amortissement-abw's comparison table had been rendering
with a four-cell header over three-cell rows.

Style-finishing pass done 2026-07-16 (full line-by-line read of all 79 FR
articles): rewrote the telegraphic `cas-types` into prose and broke the worst
colon-cascade sentences elsewhere (couple-et-famille, plancher-plafond,
inflation-histoire/-et-taux, hyperinflation, guyton-klinger); fixed generation
artifacts (embedded NUL bytes in guardrails-morningstar, "et." broken clauses,
spurious capital-L elisions, orphan ":" table cells, sentence-start casing,
a few elision/grammar slips); switched `seaux` -> `buckets` in running text
(the strategie-buckets title keeps the French gloss once, as does the lexique
entry). Added two themed SVG figures (figures.go): `allocation-plateau` in
allocation-actions-obligations and `withdrawal-frontier` in
panorama-strategies-retrait. No pofo-advertising phrasing was found needing the
"simulateur FIRE" softening beyond the cas-types rewrite; the many remaining
pofo mentions are legitimate references to the page's actual features.

## Goal

An extremely complete, engaging FIRE/decumulation handbook, written in French
(English translation later), cut into standalone articles so a reader can
enter anywhere. Every page must be rich, long, practical and illustrated:
examples, pro-tips, callouts, cross-links everywhere. The bar for total volume
is the locador embedded doc (~90 articles, ~198k words); the target here is at
least that, ideally more, with individual articles longer than locador's.

The scientific state of the art must be mobilised throughout: the classics
(Trinity, Bengen, Guyton-Klinger, Fama-French...) described and explicitly
labelled as classic/dated where they are, and the current research (Anarkulova
& Cederburg, ERN's SWR series, Morningstar's guardrails and annual SWR
reports, Kitces, Pfau, pension-fund practice) given the leading role.

## Architecture

New package `pkg/firebook`, stdlib only:

- `assets/book/fr/<slug>.md`: the articles, French, `go:embed`ed. The
  language directory leaves room for the planned English translation
  (`assets/book/en/`), and pofo's technical URLs stay English: the book is
  mounted at `/firebook/fr/` (the old `/book/fr/` path 301-redirects there).
- `manifest.go`: the table of contents as data (category -> [slug, title,
  blurb]). Single source of truth: the index page and navigation are generated
  from it. It only lists articles that exist; the full planned TOC lives in
  this design doc's ledger.
- `render.go`: a mini Markdown-to-HTML engine (same dialect as locador's
  docs.js, but in Go so pages are rendered server-side): ## / ### headings,
  bold/italic/inline code, external links, wiki-links `[[slug]]` /
  `[[slug|label]]`, `-` and `1.` lists, pipe tables, `>` quotes, `---` rules,
  and callout blocks `::: type Title` ... `:::`.
- `handler.go`: `func Handler() http.Handler` serving the index (sommaire) at
  `/` and each article at `/<slug>`, as full HTML pages with an embedded
  reading-oriented stylesheet (comfortable measure, generous line-height,
  webui visual identity: petrol accent, same fonts). Article headings get a
  hover-revealed "§" anchor link (web chrome only, injected post-render, so
  the EPUB stays clean); clicking it copies the section URL to the clipboard
  with a "lien copié" confirmation, and degrades to plain hash navigation
  without JavaScript or a secure context.
- Guard test: every file under `assets/book/` appears in the manifest and vice
  versa; every `[[slug]]` in every article resolves to a manifest slug.

Mounted in `pkg/decumul/web.Handler` under `/firebook/fr/`, and linked very
discreetly (small link at the bottom of the "How this machine works"
fold) in the fire page. Because the book is its own package with a self-contained
handler, finador can mount the exact same book by importing `firebook`.

Alternatives considered: copying locador's client-side docs.js engine into the
fire SPA (duplicates a JS engine, couples the book to the fire page's app
shell); pre-rendering at build time (needless build machinery). Server-side
rendering in Go is testable, stdlib-only and reusable.

## Writing conventions

- French; no em-dash ever; numbers in French style in prose (4 %, 1 000 000).
- PROSE STYLE (2026-07-17): clear, simple, engaging, readable, pleasant,
  fluid, good French. Hard rule: **no sentence may contain more than one French
  colon ` : `.** The colon-chain "A : B : C" reads clumsy ("claquee au sol") and
  hard to parse. Keep the announcing colon; make the second a period + new
  sentence, or a comma / "car" / "tandis que", or parentheses for `[[link]]`
  refs, or an arrow "→" for stat readouts. This applies to ALL new text from now
  on (do not introduce new double-colon sentences). A one-off book-wide cleanup
  of the pre-existing ~1100 offenders is in progress and paused; see the
  `firebook-no-double-colon` note. Do NOT add an enforcing guard test until that
  cleanup finishes (it would fail on the backlog).
- NO OUTRANCIER GALLICISMS (2026-07-17): never translate an English idiom
  literally ("sur une serviette" for "on a napkin" is out). Use the English
  expression as-is or a real French idiom ("un calcul de coin de table").
  Occasional English terms are welcome (free lunch, drawdown, trend...);
  "bon francais" means CORRECT French (syntax, vocabulary, spelling), not
  French-only vocabulary. Also: do not abuse bold (and never uppercase for
  emphasis; acronyms and typographic conventions only).
- A REAL ENGLISH WORD BEATS A CALQUE (2026-07-29). This ranks the rule above.
  Given the choice, keep the actual English term, glossed in French at first
  use, rather than mint a French word shaped like the English one. A calque is
  the worst of both: neither French nor English, and it teaches the reader
  nothing. So "playbook", "time buckets", "cap-weighted" and "free lunch" stay;
  "visualisateur de portefeuille", "bull obligataire", "retour de valorisation"
  and "generateur de permission" had to go. Only replace English with French
  when a REAL French word or idiom exists. Mechanical greps do not find calques
  (they surface "cout d'opportunite" and "releve de carriere", which are
  correct): the only method is to read the added prose.
- STANDALONE BOOK (2026-07-17): the book is not "a part of pofo" and must
  read offline without it. Avoid pofo references except when really useful
  (a pro-tip or a usage explanation), and then as an `::: encart` or
  `::: astuce` callout, not in running text.
- When ADDING articles, also AMEND the existing ones: hub articles
  (actifs-defensifs, panorama-strategies-retrait, lexique...) must mention and
  link the new dedicated articles; themes judged too small for an article
  should be evaluated for a paragraph or a mention in an existing article.
- Callout types: `cle` (the one idea to retain), `astuce` (pro-tip),
  `attention` (trap), `exemple` (worked numbers), `encart` (side note),
  `science` (what the research actually says, with references), `terrain`
  (practitioner/FIRE-community experience and testimony).
- DEPTH IS THE BAR (review of the first batches, 2026-07-12): the style
  (clear, airy, illustrated, practical) is right, but the first-batch length
  (~2 000 words) only passes for the introductory pages. Every non-intro
  article must be AT LEAST TWICE as long (target 4 000-5 000 words), deeper
  and more detailed: the book must consign ALL the important FIRE knowledge,
  with special weight on the transition into retirement and the
  decumulation/after phase. Concretely: full mechanisms (not summaries),
  the actual numbers and tables from the research, multiple worked examples,
  parameter choices and edge cases, objections and counter-arguments, and
  the how-to-apply-in-France angle. The 12 articles written before this bar
  was set (batches 1-3) are due a deepening pass, lowest priority for the
  intro ones.
- Every article opens with a plain-language paragraph stating what the reader
  will be able to DO after reading; dense cross-linking `[[slug]]`; worked
  examples with numbers throughout; a "Pour aller plus loin" closing block
  with external references where relevant.
- Classic vs state of the art: whenever a concept is a classic that research
  has since qualified or superseded (Trinity, 4% rule, buckets...), say so
  explicitly in the text, keep it (it is still interesting), and point to the
  modern treatment.
- External references: ERN's SWR series (earlyretirementnow.com) is a primary
  source; also Morningstar's "State of Retirement Income" reports, Kitces,
  Pfau, Bogleheads wiki, Anarkulova/Cederburg papers, AMF/impots.gouv for the
  French tax pages. Cite by name and URL in "Pour aller plus loin".
- French tax/social pages carry a dated-accuracy warning (rules move yearly).

## Table of contents

Every article below is written, embedded and in the manifest (79 v1 +
7 extension). The link guard test keeps its own copy of this slug list;
update both when adding an article.

### I. Demarrer
- fire-cest-quoi: Le FIRE, c'est quoi ? (histoire, variantes Lean/Fat/Barista/Coast, ordres de grandeur)
- la-regle-des-4-pourcents: La regle des 4 % en dix minutes (et pourquoi ce n'est qu'un point de depart)
- combien-il-vous-faut: Combien il vous faut (25x, 28x, 33x : du budget au capital cible)
- les-trois-phases: Accumulation, transition, retrait : les trois vies d'un plan FIRE
- utiliser-la-page-fire: Utiliser la page FIRE de pofo (chaque section, chaque controle)
- erreurs-classiques-fire: Les dix erreurs qui ruinent un plan FIRE

### II. La science du retrait
- etude-trinity: Bengen, l'etude Trinity et la naissance du taux de retrait sur (classique)
- sequence-des-rendements: Le risque de sequence : le vrai ennemi du retraite
- ruine-et-probabilites: La probabilite de ruine : la lire, la choisir, ne pas la subir
- rendements-arithmetiques-geometriques: Moyenne arithmetique, moyenne geometrique et volatility drag
- anarkulova-cederburg: Au-dela des Etats-Unis : Anarkulova, Cederburg et l'echantillon mondial (etat de l'art)
- valorisations-et-cape: Les valorisations (CAPE) et ce qu'elles disent du taux de retrait
- rendements-attendus: Les rendements attendus prospectifs (Morningstar, Vanguard, banques d'investissement)
- horizon-et-esperance-de-vie: Horizon, esperance de vie et retraites de 50 ans
- serie-ern: La serie Safe Withdrawal Rate d'ERN : guide de lecture
- les-maths-du-4-pourcent: Pourquoi 4 % ? L'anatomie mathematique de la regle (rendement reel, vol drag, sequence, horizon)
- decider-sous-incertitude: Decider sous incertitude : utilite, Kelly, equivalent certain, regret

### III. Modeliser : Monte-Carlo et autres machines
- monte-carlo-forces-faiblesses: Monte-Carlo : forces, faiblesses, bon usage
- historique-vs-parametrique: Fenetres historiques, bootstrap, parametrique : trois familles de modeles
- queues-epaisses: Queues epaisses, crises et Student-t
- lire-un-fan-chart: Lire un fan chart et des percentiles sans se tromper
- pieges-des-simulateurs: Les pieges des simulateurs (independance, biais americain, survivant...)
- rendre-monte-carlo-pertinent: Rendre un Monte-Carlo pertinent (blending, regimes, stress)
- regimes-de-marche: Les regimes de marche (croissance x inflation, ours collants) et pourquoi ils comptent

### IV. Les strategies de retrait
- panorama-strategies-retrait: Panorama des strategies de retrait : la carte avant le territoire
- retrait-fixe-bengen: Le retrait fixe indexe (Bengen) : le classique de reference
- pourcentage-fixe: Le pourcentage fixe du portefeuille : increvable mais inconfortable
- guyton-klinger: Guyton-Klinger : les guardrails historiques, grandeur et limites
- vpw: VPW, le retrait a pourcentage variable des Bogleheads
- regles-cape: Les regles CAPE : ajuster le retrait aux valorisations (ERN)
- guardrails-morningstar: Les guardrails modernes (Morningstar) : l'etat de l'art
- amortissement-abw: Le retrait par amortissement (ABW/TPAW) : l'approche actuarielle
- plancher-plafond: Plancher-plafond et regles Vanguard : la flexibilite bornee
- rentes-et-annuites: Rentes, annuites et safety first : acheter un plancher
- sept-facons-de-vivre: Sept facons de vivre du meme portefeuille (les sept regles rejouees sur 1973, 1985 et 2000)
- choisir-sa-strategie: Choisir sa strategie : criteres, comparatif, cas d'usage

### V. Le portefeuille de retrait
- primes-de-risque: D'ou viennent les rendements : les primes de risque (actions, terme, credit, et pourquoi l'or ne rapporte rien)
- pourquoi-la-diversification-marche: Pourquoi la diversification fonctionne : correlation, rebalancing premium, volatility harvesting
- allocation-actions-obligations: L'allocation actions/obligations en retrait
- glidepaths: Les glidepaths : bond tent, rising equity et la fenetre fragile
- portefeuilles-tous-temps: Les portefeuilles tous-temps : Browne, All-Weather, Golden Butterfly, Dragon
- actifs-defensifs: Les actifs defensifs : panorama et roles
- or-en-retrait: L'or dans un portefeuille de retrait
- obligations-en-retrait: Les obligations en retrait : types, duree, role exact
- obligations-indexees: Les obligations indexees sur l'inflation
- facteurs-fama-french: Les facteurs (Fama-French, value, momentum) en phase de retrait
- diversification-internationale: La diversification internationale (et le biais domestique)
- etf-ucits-europeens: Construire en UCITS : le portefeuille de retrait de l'investisseur europeen

### V bis. Les actifs alternatifs
- managed-futures: Managed futures et suivi de tendance : la diversification qui travaille dans les crises (moved here 2026-07-17)
- long-volatility: Long volatility et tail hedging : payer pour les krachs
- global-macro: Global macro et strategies de primes alternatives (dont commodity carry)
- return-stacking: Return stacking, overlays et portable alpha : empiler les primes

### VI. Buffers et protections
- cash-buffer: Le matelas de liquidites : taille, cout, vrai role
- strategie-buckets: Les buckets : la strategie des seaux, promesse et critique
- echelle-obligataire: Les echelles d'obligations (et l'echelle de linkers)
- recharger-ou-pas: Consommer et recharger un buffer : les regles qui marchent
- immobilier-en-retrait: L'immobilier dans un plan FIRE (residence, locatif)
- levier-et-marges: Levier, marge et lombard en retrait (avance)

### VII. L'inflation
- inflation-histoire: L'inflation sur les dernieres decennies : ce que 1970-2025 enseigne
- suivre-inflation: Suivre l'inflation : les indices, et la votre
- inflation-et-taux-de-retrait: Inflation et taux de retrait : le lien exact
- se-proteger-de-inflation: Se proteger de l'inflation : ce qui marche vraiment
- hyperinflation-et-extremes: Hyperinflations et scenarios extremes

### VIII. Fiscalite et cadre francais
- enveloppes-francaises: PEA, assurance-vie, CTO : les enveloppes du rentier francais
- flat-tax-et-imposition: PFU, bareme, abattements : l'imposition des retraits
- taxe-puma: La taxe PUMa : le piege du rentier francais
- retraite-legale: FIRE et retraite legale : trimestres, AGIRC-ARRCO, decote
- sante-et-protection-sociale: Sante et protection sociale du rentier
- succession-et-transmission: Succession et transmission
- expatriation-fiscale: L'expatriation : fiscalite et protection sociale

### IX. Le facteur humain
- psychologie-du-retrait: La psychologie du retrait : pourquoi depenser est si dur
- temoignages-fire: Ce que disent les vrais FIRE : temoignages et conseils
- sens-et-identite: Sens, identite, structure : la vie apres le travail
- couple-et-famille: FIRE en couple et en famille
- flexibilite-realite: La flexibilite : mythe et realite (ce qu'elle peut vraiment absorber)
- une-annee-de-plus: Le syndrome de l'annee de plus
- retour-au-travail: Barista, coast, side income : le travail choisi

### X. En pratique
- construire-son-plan: Construire son plan pas a pas
- revue-annuelle: La revue annuelle : la check-list du rentier
- quand-s-inquieter: Quand s'inquieter, quand laisser courir
- marche-baissier-en-retraite: Traverser un marche baissier en retraite : le playbook
- revenus-complementaires: Pensions et revenus complementaires dans le plan
- depenses-en-retraite: Les depenses reelles en retraite (retirement smile, Die With Zero)
- cas-types: Trois plans complets, chiffres de bout en bout

### XI. References
- lexique: Lexique du FIRE et du retrait
- bibliotheque: La bibliotheque : sites, papiers, livres, outils
- la-machine-pofo: Sous le capot : comment pofo calcule ce livre

87 articles planned (79 v1 + 7 of the 2026-07-17 extension + sept-facons-de-vivre,
added 2026-07-29); at 2 500 words each the book lands around 218k words.

### Data-backed articles

`sept-facons-de-vivre` is the first article whose figures and tables carry
computed numbers rather than illustrative ones. The pattern to follow when
another one needs it: the computation lives in a library package
(`pkg/replay`), the plate keeps FROZEN literal arrays so book figures stay
pure dependency-free functions like every other plate, and a guard test
recomputes both the arrays and the article's markdown tables from the engine
and fails on any drift. That way a refreshed dataset breaks the build instead
of quietly leaving a wrong chart in the book. Its summary tables are also
TRANSPOSED (rules as rows, statistics as columns): a book page and an e-reader
cannot carry a year x rule matrix without horizontal scrolling, and the
year-by-year detail belongs in a figure anyway.

`figures_strategies.go` (2026-07-29) applies the same pattern to the plates
of the withdrawal-strategies part, with a taste rather than a rule: a figure
earns its page by fitting ITS question, which in that batch happened to give
every plate a different form (a warning light climbing over nineteen years in
`bengen-falaise`, a plane of constant incomes crossed two ways in
`cape-contracyclique`, two curves swapping rank at a nameable age in
`credits-mortalite`, a narrowing procedure in `arbre-decision`, four
instrument strips in `deux-thermometres`). Reusing an existing form is fine
whenever it is the right one; what is avoided is defaulting to a past figure
as a template. This is deliberately NOT a "never repeat a form" rule.
Where a plate carries a closed-form model rather than a replayed series (the
CAPE rule, the Gompertz mortality law, the geometric bound), the model lives in
the plate as a small pure function and the guard test checks it against the
numbers the article quotes in prose, so figure and text cannot drift apart. A
plate may also freeze raw inputs and derive the interesting series in code
(`pourcentage-lissages` freezes twelve real returns and runs the three smoothing
rules on them; `cape-depuis-1881` freezes 146 CAPE readings and applies the
formula), which keeps the rule under study readable next to the picture it
produces.

Fourteen of those plates shipped on 2026-07-29, from the candidate list in
`fire-book-illustrations-2026-07.md`, which is now closed: six candidates were
dropped on purpose and each carries its reason there. Two of the reasons are
worth knowing outside that file.

Replaying Guyton-Klinger at four initial rates contradicts a sourced sentence of
`guyton-klinger.md`, and `pkg/replay`'s GK is only the corridor, without the
indexation freeze or the last-fifteen-years suspension, so it is not faithful
enough to overturn the literature. Do not ship a figure that argues with its own
article.

And `rentes-et-annuites.md` claims that annuitising can raise the headline ruin
while improving the worst late-life outcomes. A sweep over ages 60-75 and rates
3.5-5.5 % could not reproduce it: the two readings always move together. The
cause is structural rather than a calibration problem, and it is a limitation of
the engine worth remembering: `decumul` runs a FIXED horizon, so the risk an
annuity insures against (outliving the plan) does not exist inside it, and
`LifeCurve` applying mortality after the fact cannot recreate it. Evaluating
annuities properly would need a stochastic-lifetime kernel.

## Adding articles

Research with live sources where needed (ERN, Morningstar,
service-public/impots.gouv for the French pages), write, add to
`manifest.go`, extend the TOC above and the guard test's slug list,
`make check`, commit. Articles may link `[[slugs]]` that do not exist yet:
the guard test checks links against the planned slug list (kept as data in
the test), so forward links are allowed while typos are still caught.
