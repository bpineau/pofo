# The FIRE book: design

Status: infrastructure shipped; writing in progress (see the ledger at the
bottom, which is the authoritative progress tracker across sessions).

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
  mounted at `/book/fr/`.
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
  webui visual identity: petrol accent, same fonts).
- Guard test: every file under `assets/book/` appears in the manifest and vice
  versa; every `[[slug]]` in every article resolves to a manifest slug.

Mounted in `pkg/decumul/web.Handler` under `/book/fr/`, and linked very
discreetly (small `book` link at the bottom of the "How this machine works"
fold) in the fire page. Because the book is its own package with a self-contained
handler, finador can mount the exact same book by importing `firebook`.

Alternatives considered: copying locador's client-side docs.js engine into the
fire SPA (duplicates a JS engine, couples the book to the fire page's app
shell); pre-rendering at build time (needless build machinery). Server-side
rendering in Go is testable, stdlib-only and reusable.

## Writing conventions

- French; no em-dash ever; numbers in French style in prose (4 %, 1 000 000).
- Callout types: `cle` (the one idea to retain), `astuce` (pro-tip),
  `attention` (trap), `exemple` (worked numbers), `encart` (side note),
  `science` (what the research actually says, with references), `terrain`
  (practitioner/FIRE-community experience and testimony).
- DEPTH IS THE BAR (Ben's review of the first batches, 2026-07-12): the style
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

## Planned table of contents and progress ledger

Mark `[x]` when an article is written, embedded and in the manifest.

### I. Demarrer
- [x] fire-cest-quoi: Le FIRE, c'est quoi ? (histoire, variantes Lean/Fat/Barista/Coast, ordres de grandeur)
- [x] la-regle-des-4-pourcents: La regle des 4 % en dix minutes (et pourquoi ce n'est qu'un point de depart)
- [x] combien-il-vous-faut: Combien il vous faut (25x, 28x, 33x : du budget au capital cible)
- [x] les-trois-phases: Accumulation, transition, retrait : les trois vies d'un plan FIRE
- [x] utiliser-la-page-fire: Utiliser la page FIRE de pofo (chaque section, chaque controle)
- [x] erreurs-classiques-fire: Les dix erreurs qui ruinent un plan FIRE

### II. La science du retrait
- [x] etude-trinity: Bengen, l'etude Trinity et la naissance du taux de retrait sur (classique)
- [x] sequence-des-rendements: Le risque de sequence : le vrai ennemi du retraite
- [x] ruine-et-probabilites: La probabilite de ruine : la lire, la choisir, ne pas la subir
- [x] rendements-arithmetiques-geometriques: Moyenne arithmetique, moyenne geometrique et volatility drag
- [x] anarkulova-cederburg: Au-dela des Etats-Unis : Anarkulova, Cederburg et l'echantillon mondial (etat de l'art)
- [x] valorisations-et-cape: Les valorisations (CAPE) et ce qu'elles disent du taux de retrait
- [x] rendements-attendus: Les rendements attendus prospectifs (Morningstar, Vanguard, banques d'investissement)
- [x] horizon-et-esperance-de-vie: Horizon, esperance de vie et retraites de 50 ans
- [x] serie-ern: La serie Safe Withdrawal Rate d'ERN : guide de lecture

### III. Modeliser : Monte-Carlo et autres machines
- [x] monte-carlo-forces-faiblesses: Monte-Carlo : forces, faiblesses, bon usage
- [x] historique-vs-parametrique: Fenetres historiques, bootstrap, parametrique : trois familles de modeles
- [x] queues-epaisses: Queues epaisses, crises et Student-t
- [x] lire-un-fan-chart: Lire un fan chart et des percentiles sans se tromper
- [x] pieges-des-simulateurs: Les pieges des simulateurs (independance, biais americain, survivant...)
- [x] rendre-monte-carlo-pertinent: Rendre un Monte-Carlo pertinent (blending, regimes, stress)
- [x] regimes-de-marche: Les regimes de marche (croissance x inflation, ours collants) et pourquoi ils comptent

### IV. Les strategies de retrait
- [x] panorama-strategies-retrait: Panorama des strategies de retrait : la carte avant le territoire
- [x] retrait-fixe-bengen: Le retrait fixe indexe (Bengen) : le classique de reference
- [x] pourcentage-fixe: Le pourcentage fixe du portefeuille : increvable mais inconfortable
- [x] guyton-klinger: Guyton-Klinger : les guardrails historiques, grandeur et limites
- [x] vpw: VPW, le retrait a pourcentage variable des Bogleheads
- [x] regles-cape: Les regles CAPE : ajuster le retrait aux valorisations (ERN)
- [x] guardrails-morningstar: Les guardrails modernes (Morningstar) : l'etat de l'art
- [x] amortissement-abw: Le retrait par amortissement (ABW/TPAW) : l'approche actuarielle
- [x] plancher-plafond: Plancher-plafond et regles Vanguard : la flexibilite bornee
- [x] rentes-et-annuites: Rentes, annuites et safety first : acheter un plancher
- [x] choisir-sa-strategie: Choisir sa strategie : criteres, comparatif, cas d'usage

### V. Le portefeuille de retrait
- [x] allocation-actions-obligations: L'allocation actions/obligations en retrait
- [x] glidepaths: Les glidepaths : bond tent, rising equity et la fenetre fragile
- [x] portefeuilles-tous-temps: Les portefeuilles tous-temps : Browne, All-Weather, Golden Butterfly, Dragon
- [x] actifs-defensifs: Les actifs defensifs : panorama et roles
- [x] or-en-retrait: L'or dans un portefeuille de retrait
- [x] obligations-en-retrait: Les obligations en retrait : types, duree, role exact
- [x] obligations-indexees: Les obligations indexees sur l'inflation
- [x] managed-futures: Managed futures et suivi de tendance : la diversification qui travaille dans les crises
- [x] facteurs-fama-french: Les facteurs (Fama-French, value, momentum) en phase de retrait
- [x] diversification-internationale: La diversification internationale (et le biais domestique)
- [x] etf-ucits-europeens: Construire en UCITS : le portefeuille de retrait de l'investisseur europeen

### VI. Buffers et protections
- [x] cash-buffer: Le matelas de liquidites : taille, cout, vrai role
- [x] strategie-buckets: Les buckets : la strategie des seaux, promesse et critique
- [x] echelle-obligataire: Les echelles d'obligations (et l'echelle de linkers)
- [x] recharger-ou-pas: Consommer et recharger un buffer : les regles qui marchent
- [x] immobilier-en-retrait: L'immobilier dans un plan FIRE (residence, locatif)
- [x] levier-et-marges: Levier, marge et lombard en retrait (avance)

### VII. L'inflation
- [x] inflation-histoire: L'inflation sur les dernieres decennies : ce que 1970-2025 enseigne
- [x] suivre-inflation: Suivre l'inflation : les indices, et la votre
- [x] inflation-et-taux-de-retrait: Inflation et taux de retrait : le lien exact
- [x] se-proteger-de-inflation: Se proteger de l'inflation : ce qui marche vraiment
- [x] hyperinflation-et-extremes: Hyperinflations et scenarios extremes

### VIII. Fiscalite et cadre francais
- [x] enveloppes-francaises: PEA, assurance-vie, CTO : les enveloppes du rentier francais
- [x] flat-tax-et-imposition: PFU, bareme, abattements : l'imposition des retraits
- [x] taxe-puma: La taxe PUMa : le piege du rentier francais
- [x] retraite-legale: FIRE et retraite legale : trimestres, AGIRC-ARRCO, decote
- [x] sante-et-protection-sociale: Sante et protection sociale du rentier
- [x] succession-et-transmission: Succession et transmission
- [x] expatriation-fiscale: L'expatriation : fiscalite et protection sociale

### IX. Le facteur humain
- [x] psychologie-du-retrait: La psychologie du retrait : pourquoi depenser est si dur
- [x] temoignages-fire: Ce que disent les vrais FIRE : temoignages et conseils
- [x] sens-et-identite: Sens, identite, structure : la vie apres le travail
- [x] couple-et-famille: FIRE en couple et en famille
- [x] flexibilite-realite: La flexibilite : mythe et realite (ce qu'elle peut vraiment absorber)
- [x] une-annee-de-plus: Le syndrome de l'annee de plus
- [ ] retour-au-travail: Barista, coast, side income : le travail choisi

### X. En pratique
- [ ] construire-son-plan: Construire son plan pas a pas
- [ ] revue-annuelle: La revue annuelle : la check-list du rentier
- [ ] quand-s-inquieter: Quand s'inquieter, quand laisser courir
- [ ] marche-baissier-en-retraite: Traverser un marche baissier en retraite : le playbook
- [ ] revenus-complementaires: Pensions et revenus complementaires dans le plan
- [ ] depenses-en-retraite: Les depenses reelles en retraite (retirement smile, Die With Zero)
- [ ] cas-types: Trois plans complets, chiffres de bout en bout

### XI. References
- [ ] lexique: Lexique du FIRE et du retrait
- [ ] bibliotheque: La bibliotheque : sites, papiers, livres, outils
- [ ] la-machine-pofo: Sous le capot : comment pofo calcule ce livre

79 articles planned; at 2 500 words each the book lands around 200k words.

## Writing plan (multi-session)

Each session: pick the next unwritten articles (order of the ledger, but
Demarrer + science first since everything links into them), research with
live sources where needed (ERN, Morningstar, service-public/impots.gouv for
the French pages), write, add to `manifest.go`, tick the ledger, `make check`,
commit. Articles may link `[[slugs]]` that do not exist yet; the guard test
only checks links against the PLANNED slug list above (kept as data in the
test) so forward links are allowed before their targets are written, while
typos are still caught.
