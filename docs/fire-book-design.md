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

86 articles planned (79 v1 + 7 of the 2026-07-17 extension); at 2 500 words
each the book lands around 215k words.

## Adding articles

Research with live sources where needed (ERN, Morningstar,
service-public/impots.gouv for the French pages), write, add to
`manifest.go`, extend the TOC above and the guard test's slug list,
`make check`, commit. Articles may link `[[slugs]]` that do not exist yet:
the guard test checks links against the planned slug list (kept as data in
the test), so forward links are allowed while typos are still caught.
