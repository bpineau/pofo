# Firebook: figure backlog for « Le portefeuille de retrait »

Status: ACTIVE backlog, opened 2026-07-30 during the line-by-line review of the
fourteen articles of the portfolio part (one reviewer per article; the prose
fixes shipped as the "line review of ..." commit series ending at this one).
Nothing here is implemented; the ideas await selection. Delete this file once
the retained figures are shipped or the list is dropped (docs/ stays curated).

Figures, when built, follow the v2 plate system and the frozen-array +
guard-test pattern recorded in `fire-book-design.md`. Costs: **A** = data
already in the repo or the article, **B** = a computation to write first,
**C** = a table or callout does the job better. Written in French because the
book is.

---

## primes-de-risque

L'article porte déjà `::: figure primes-echelle` (barres d'intervalle classées,
l'échelle des primes au-dessus du cash). Toute idée ci-dessous doit donc éviter
cette forme et servir un autre passage.

## 1. Le nuage « douleur au mauvais moment » contre « prime » (coût B)

Un nuage de points, un actif par point. En abscisse, le rendement réel moyen de
l'actif dans les dix pires années des actions depuis 1954 (l'axe de la douleur,
donc). En ordonnée, sa prime réelle annualisée au-dessus du cash sur la période
complète. Actions, obligations longues, 60/40, crédit, cash s'alignent sur une
droite décroissante, et l'or se pose franchement hors de la droite, en haut à
gauche de l'axe de douleur avec une prime nulle.

C'est la thèse centrale de l'article, aujourd'hui entièrement verbale (« le
rendement est le salaire du risque porté », « ce qui compte est la covariance
avec les mauvais états du monde ») et la seule preuve visuelle possible que
l'or n'est pas un actif raté mais un actif d'une autre nature. La figure
existante donne les niveaux, celle-ci donnerait la raison des niveaux.

Coût B : le calcul est à écrire, mais les séries sont là (`pkg/replay` pour le
réel US 60/40 depuis 1954, S&P 500 et Treasuries 5 ans déflatés par le CPI ;
`pkg/datasets` pour l'or et les taux courts). Limiter le nuage aux actifs
réellement bundlés et ne pas inventer de point crédit ou trend faute de série.

## 2. La grille d'audit, en tableau (coût C)

Pas de figure : un tableau à quatre colonnes (la ligne du portefeuille, quelle
prime, qui la paie, pourquoi elle tient encore dans vingt ans), sur huit lignes
typiques d'un portefeuille de particulier (ETF monde, obligations d'État
longues, linkers, or, trend, fonds thématique, produit structuré, SCPI ou
private equity grand public).

La section « L'audit de votre portefeuille » est le passage le plus utile de
l'article et le plus dense à lire, car ses trois exemples sont enchaînés dans
un seul paragraphe. Un tableau les rend consultables ligne par ligne, ce qui
est exactement l'usage visé (auditer son propre portefeuille, feuille en main),
et il accueille sans effort les deux ou trois cas que le paragraphe n'a pas la
place de traiter.

## 3. La décote des primes après publication, en pente (coût B)

Un graphe en pente à trois points sur l'axe du temps de vie d'une anomalie
(période étudiée, hors échantillon, après publication) avec le rendement
ramené à 100 au départ, soit 100, 74 et 42 d'après McLean et Pontiff. Trois
familles annotées en bout de course selon leur destin (risque macro non
assurable, prime comportementale à limite d'arbitrage, anomalie statistique
fine).

Le bloc `::: science` raconte cette hiérarchie de robustesse sans jamais la
montrer, et c'est pourtant l'argument qui autorise le lecteur à faire confiance
au moteur actions tout en dosant le trend et en ignorant le reste.

Réserve honnête : seuls les trois points agrégés (100 / 74 / 42) sont
documentés par l'article original. Les trois familles doivent rester des
annotations qualitatives, sans courbes chiffrées propres, sous peine d'inventer
des données. Si cette contrainte rend la figure trop maigre, l'idée retombe en
coût C sous forme d'encadré à trois étages.

---

## pourquoi-la-diversification-marche

L'article a déjà `::: figure correlation-vol` (courbe volatilité du panier 50/50
selon ρ). Toute nouvelle idée doit donc éviter la forme « courbe unique d'une
grandeur en fonction d'un paramètre », déjà prise dans le même article.

## 1. Le carnet de comptes du démon de Shannon (coût C)

Le bloc `::: exemple` affirme qu'un actif à rendement géométrique nul, mélangé à
50/50 avec du cash à 0 % et rééquilibré, produit environ 6 % par an. Le lecteur
doit croire sur parole. Un petit tableau à trois colonnes (année, tirage de
l'actif, valeur du panier après rééquilibrage) sur quatre ou six ans rend le
mécanisme vérifiable au crayon : +50 % après un pile, −25 % après un face, et
√(1,50 × 0,75) − 1 = 6,07 % par an. C'est le seul endroit du livre où le free
lunch se voit naître ligne à ligne.

Pourquoi un tableau et pas une figure : la démonstration est arithmétique, pas
visuelle, et deux trajectoires qui divergent seraient une forme déjà prise
(courbes qui se croisent, petites-multiples de séries). Les chiffres se
recalculent à la main dans l'article, aucune donnée externe n'est nécessaire.

## 2. Le triangle de corrélations : trente fonds actions contre quatre briques (coût B)

Le passage le plus fort de l'article est aussi le plus verbal : « trente fonds
actions ne font qu'un seul actif » (ρ de 0,85 à 0,95) alors que quatre briques
bien choisies tiennent entre −0,2 et +0,3. Une demi-matrice de corrélations, en
triangle, sur sept lignes (trois variantes actions monde/US/tech, puis actions
mondiales, obligations longues, or, trend) montre d'un seul coup d'œil un bloc
saturé en haut à gauche et une zone pâle ailleurs. La thèse « la diversification
se compte en corrélations, pas en lignes » devient une image.

Forme neuve pour le livre (aucune matrice de corrélation n'existe encore ; la
« matrice de barres » déjà prise est un objet différent). Attention à la règle
maison : labels horizontaux, pas de diagonale décorative, échelle de couleur
divergente validée. Coût B : les séries longues sont dans le repo (simdata
actions/obligations longues/or/trend) et `suggest.Correlation` existe déjà, mais
il faut écrire le petit calcul et figer les valeurs dans la figure.

## 3. La médiane baisse, le percentile 5 monte (coût B)

La section décumulation porte la thèse centrale de l'article et n'a aucune
image : en passant de 100 % actions à un panier de quatre briques, la richesse
médiane à trente ans baisse un peu tandis que le SWR à 95 % de succès monte.
Deux mesures qui bougent en sens contraire, c'est exactement ce qu'une figure
sait dire mieux qu'une phrase. Forme proposée : deux axes verticaux côte à côte
(richesse médiane à gauche, SWR à 95 % à droite), chaque portefeuille tracé par
un segment reliant sa position sur les deux axes, les deux segments se croisant
en X. Le lecteur voit le prix payé (un peu de médiane) et ce qu'il achète (du
plancher).

Ce n'est ni un fan chart, ni des barres appariées, ni un plan à iso-courbes :
la forme (deux échelles reliées par des segments) est libre. Coût B : les
briques du panier et le moteur de décumulation existent, mais les deux couples
de chiffres doivent être calculés et figés, et le choix des briques documenté
dans la légende pour ne pas laisser croire à un résultat universel.

---

## concevoir-un-portefeuille

Article de méthode, aujourd'hui sans aucune figure. Il n'a pas besoin d'une
figure de données pour exister, mais deux passages purement verbaux portent la
thèse et gagneraient beaucoup à être vus : la carte risques -> briques
(question 5) et l'arbitrage pires chemins / médiane (question 6). Trois idées,
classées par valeur.

## 1. La carte bipartite « un risque, une réponse » (coût A)

Deux colonnes reliées par des liens : à gauche les risques nommés à la question
1 (séquence, marché baissier prolongé, inflation persistante, déflation, crise
monétaire, plus longévité et santé en gris, hors portefeuille), à droite les
briques de la question 5 (actions monde, duration d'État, matelas, linkers, or,
trend, flexibilité des dépenses). Chaque lien porte l'épaisseur de l'intensité
de défense ; les risques acceptés restent à gauche sans lien sortant.

Elle porte exactement la thèse du chapitre, que l'allocation est la somme des
réponses et non un point de départ, et elle rend visibles les deux pathologies
que le texte décrit sans les montrer : la brique sans lien entrant (« pas de
réponse, pas de diversification, seulement des frais ») et le risque défendu
deux fois pendant qu'un autre reste nu. Diagnostic supplémentaire, elle exhibe
le trou de l'article lui-même, le marché baissier prolongé n'ayant aucune ligne
dans la carte du texte.

Coût A : toute la matière est dans l'article. Forme neuve dans le livre (aucun
graphe bipartite / de liens existant). Piège : garder les libellés horizontaux
des deux côtés, pas de texte le long des liens.

## 2. La fiche de thèse d'une brique (coût C, encadré ou tableau)

La question 7 fait reposer toute la discipline sur « la thèse écrite de chaque
brique », mais le livre ne montre jamais à quoi ressemble une thèse écrite. Un
encadré `exemple` d'une dizaine de lignes, rempli pour une seule brique (l'or
du ménage de l'exemple, disons), avec six champs : risque défendu, régime où
elle gagne quand le reste perd, coût annuel assumé, preuve retenue, taille et
son motif, condition de révision.

C'est le seul livrable réutilisable du chapitre, et il transforme une injonction
(« écrivez la thèse ») en gabarit copiable. Une figure ne ferait rien de plus
ici, d'où le coût C : ni graphique ni dataviz, un encadré.

## 3. Le nuage des arbitrages : ce qu'une brique fait au pire chemin et à la médiane (coût B)

Un point par variante de conception, l'origine étant le portefeuille nu. En
abscisse l'effet sur la médiane du revenu (ou du capital final réel), en
ordonnée l'effet sur le percentile 5. Chaque brique défensive apparaît comme un
déplacement depuis l'origine, et son dosage comme une trace de points alignés
(5 %, 10 %, 15 % d'or, etc.). Les quadrants se lisent seuls : haut-gauche
« défend en coûtant un peu » (le bon cas décrit par le texte), bas-quelconque
« ne défend rien », droite basse « trop grosse ».

C'est la traduction visuelle du paragraphe le plus opérationnel de l'article,
celui du verdict à lire, qui reste aujourd'hui une consigne abstraite. On y voit
aussi que le dosage est un chemin qui se retourne, ce que le texte affirme sans
le prouver.

Coût B : demande un balayage (une conception de base plus chaque brique à
plusieurs doses, percentile 5 et médiane relevés à chaque fois) avant de
dessiner. Forme neuve à condition de ne pas la dériver vers la grille 2x2 de
régimes ni vers les barres d'intervalle classées, déjà prises.

---

## allocation-actions-obligations

Contexte : l'article porte déjà quatre plates (`allocation-plateau`, `duration-choc`,
`obligations-rendements`, `obligations-regimes`). Le plateau, la sensibilité de prix,
le rendement par espèce et le comportement par régime sont donc couverts. Il ne reste
qu'un seul passage à thèse forte encore entièrement verbal, la section « le véhicule
change la donne ». Le reste tient très bien en prose et une figure de plus alourdirait.

## 1. Le profil de duration des trois véhicules (coût A, recommandée)

Une plate à trois tracés, abscisse = années de détention (0 à 12), ordonnée = duration
restante du véhicule. L'ETF roulant « 7-10 ans » trace une horizontale plate à 8 : il
n'arrive jamais à échéance. Le fonds à maturité datée descend de 8 à 0 et s'arrête net
l'année de liquidation. Le zéro-coupon 12 ans descend en droite de 12 à 0. Trois
étiquettes en bout de tracé, aucune légende.

Ce qu'elle porte : la thèse de la section, à savoir que le véhicule décide du service
rendu, alors qu'aujourd'hui elle est énoncée trois fois en prose (« duration à peu près
constante », « duration décroissante, valeur finale à peu près connue », « sa duration
égale sa maturité »). Elle rend visible d'un coup l'argument le plus contre-intuitif du
chapitre, le fonds obligataire qui n'échoit jamais, donc l'illusion du « je tiens à
échéance ». Elle explique aussi, sans un mot de plus, pourquoi l'échelle se fait en
titres ou en fonds datés et pourquoi la poche rééquilibrée se fait en ETF.

Coût A : arithmétique pure, aucune donnée à chercher. Forme neuve dans le livre (aucune
plate n'a la duration en ordonnée ni la durée de détention en abscisse) et sans parenté
avec `duration-choc`, qui est un diagramme de sensibilité en barres.

## 2. La grille de tri de la poche défensive (coût C, tableau)

Un tableau compact, une ligne par espèce du bestiaire (souveraine du cœur, souveraine à
spread, indexée, IG, high yield, émergente, supranationale), trois colonnes : ce qu'elle
paie en plus du Bund, son comportement en krach d'actions, verdict (cœur défensif /
dose modérée / côté risque). C'est la mise en grille du test énoncé à la fin du
bestiaire, « chaque ligne doit pouvoir monter quand les actions plongent ».

Pourquoi C plutôt qu'une figure : les deux dimensions chiffrables sont déjà tracées
(`obligations-rendements` pour le prix, `obligations-regimes` pour le comportement de
crise). Ce qui manque n'est pas une mesure de plus, c'est le verdict par ligne, qui est
du texte. Un tableau de sept lignes le fait mieux qu'une plate et se relit en trente
secondes au moment d'ouvrir une poche obligataire.

## 3. Écartées, et pourquoi

- **La courbe des taux comme grille tarifaire.** Passage aujourd'hui verbal et joli
  (« le prix exact auquel vous vendez la tranquillité de la duration courte »), mais la
  plate serait un rendement en fonction de la maturité, c'est-à-dire la forme exacte de
  `carry-courbes` (deux courbes lisses sur un axe d'échéances) dans le chapitre global
  macro. Répétition de forme, donc non.
- **La décomposition du rendement en quatre briques** (taux réel, inflation anticipée,
  prime de terme, spread) par espèce. Pédagogiquement tentante, mais toute mise en image
  honnête est une barre empilée par espèce, forme déjà prise, et le total est déjà tracé
  par `obligations-rendements`. Non.
- **Le balayage d'allocation de l'exemple** (ruine centrale contre broad-sample, cran par
  cran). Ce serait `allocation-plateau` en version chiffrée, donc la même forme deux fois
  dans le même article. Les chiffres en prose de l'encart suffisent.

---

## glidepaths

L'article a déjà `::: figure bond-tent` (la forme de la tente : part d'actions de
85 % à -8 ans, creux 57 % au jour J, remontée à 90 % à +12 ans). Elle dit *ce
qu'est* la tente. Aucune illustration ne dit aujourd'hui *ce qu'elle rapporte*,
ni *comment elle s'exécute* : les deux thèses les plus contre-intuitives de
l'article restent purement verbales.

## 1. La courbe des millésimes triés : la pente transfère, elle ne crée pas

Un seul trait monotone. En abscisse, les millésimes de départ classés du plus
défavorisé au plus favorisé par la tente ; en ordonnée, le gain (ou la perte) en
points de taux soutenable, tente 58 → 85 % contre statique équivalente. La
courbe part vers +0,3/+0,4, traverse zéro autour de la médiane, et finit en
négatif : le lecteur voit d'un coup que la tente ne fabrique pas de rendement,
elle déplace du résultat des bons millésimes vers les mauvais. On surligne sur
la même courbe les cohortes parties à CAPE > 25, qui se massent du côté
gagnant : la seconde thèse de l'article (« assurance dont la prime baisse quand
le risque monte ») tombe sans un mot de plus. C'est la figure qui remplace les
trois formulations verbales de « +0,1-0,3 point dans les mauvais cas, ~0 en
médiane ».

Forme neuve : une courbe de gains triés qui coupe l'axe zéro, ni série
chronologique, ni éventail, ni barres. À ne pas confondre avec
`bengen-millesimes`, qui est un nuage chronologique de résultats bruts.

**Coût : B.** Le 60/40 réel américain de `pkg/replay` (depuis 1954) et le CAPE
de `pkg/datasets/cape` existent, mais il faut écrire le calcul du taux
soutenable maximal sous allocation variable, par cohorte. Attention à la
maigreur de l'échantillon : 1954 + 30 ans de retraite ne laisse qu'une
quarantaine de cohortes, il faudra sans doute remonter aux séries longues de
`pkg/datasets/refdata` pour que la courbe ne soit pas un accident.

## 2. La carte des sauvetages : ce que la tente et le buffer se disputent

Une bande de cases, une par millésime de départ, chacune coloriée par ce qui
sauve ce millésime : la tente seule, le buffer seul, les deux, aucun des deux
(et la vaste majorité, « n'avait pas besoin d'être sauvé »). L'article affirme
que les deux protections « se recouvrent largement » et qu'un gros buffer plus
une tente profonde, « c'est payer deux fois la même assurance ». C'est une
affirmation forte, actuellement livrée sans preuve, et le tableau comparatif qui
la précède ne la montre pas : il compare des propriétés, pas des sauvetages. La
carte, elle, rend l'argument irréfutable, parce que la catégorie « les deux »
occupe visiblement l'essentiel des cases sauvées.

Forme neuve : une bande catégorielle discrète (qui sauve quoi), pas une série,
pas des barres empilées.

**Coût : B.** Même moteur de calcul que l'idée 1, plus une variante avec buffer
consommé/rechargé. À ne lancer que si l'idée 1 est faite, le calcul est
mutualisé.

## 3. L'encadré de la pente automatique : trois lignes d'arithmétique

Le cœur opérationnel de l'article est que la remontée ne demande aucun ordre
d'achat. Le lecteur doit aujourd'hui le croire. Trois lignes suffisent à le lui
faire vérifier : portefeuille 500 k€, 58 % d'actions (290 k€) et 42 % de
défensif (210 k€) ; on prélève 18 k€ de l'année sur le seul défensif ; il reste
290 k€ d'actions sur 482 k€, soit 60,2 %. Aucune transaction, +2,2 points en un
an. Et la petite loi qui en sort mérite d'être écrite, car elle explique
pourquoi le standard de la littérature est justement de 2 à 3 points par an : la
dérive annuelle vaut à peu près la part d'actions multipliée par le taux de
retrait. À 3,6 % et 58 % d'actions, la pente se fait toute seule, au bon rythme,
sans qu'on l'ait décidé.

**Coût : C.** Pas de figure, un encadré `::: exemple` ou `::: astuce`. Les
chiffres sont ceux d'Iris (500 k€ implicites, 3,6 %, 58/34/8), l'arithmétique
est vérifiée ci-dessus.

---

## portefeuilles-tous-temps

L'article n'a aujourd'hui aucune figure, alors qu'il enchaîne quatre
compositions chiffrées, un bloc `science` qui n'est qu'un tableau déguisé en
phrase, et deux thèses (l'échange espérance/queues, l'écart à l'indice) qui
sont purement verbales. C'est l'article de la partie portefeuille qui a le
plus à gagner d'être illustré.

Toutes les données citées ci-dessous ont été recalculées pendant la review à
partir du repo seul (`pkg/datasets/simdata/{SP500,TLT,IEF,SHY,XAUUSD}.csv`,
`pkg/datasets/refdata/{USSCV-USD,TBILL-3M,WTI-USD}.csv`, déflateur
`pkg/marketdata/data/cpi-us.csv`), 1972-2024, rééquilibrage annuel, en réel.
Valeurs mesurées, utiles comme garde-fou d'implémentation :

| Portefeuille | CAGR réel | vol | pire année réelle | pire drawdown réel |
|---|---|---|---|---|
| Browne 4 × 25 | 4,4 % | 7,2 % | 2022, −17 % | −22 % (2020-07 → 2022-10) ; −21 % (1980-01 → 1982-06) |
| Golden Butterfly | 5,9 % | 8,2 % | 2022, −16 % | −21 % (2021-12 → 2022-09) |
| All-Weather (Robbins) | 5,0 % | 7,7 % | 2022, −25 % | −29 % (2021-11 → 2023-10) |
| 60/40 | 5,4 % | 9,4 % | 1974, −23 % | −37 % (1972-12 → 1974-09) |
| 80/20 | 6,2 % | 12,2 % | 1974, −29 % | −45 % (1972-12 → 1974-09) |
| 100 % actions | 6,8 % | 15,3 % | 2008, −37 % | −54 % (2000-08 → 2009-02) |

## 1. La grille des régimes, recette par recette (coût A)

Barres horizontales empilées, une ligne par portefeuille (60/40 en ligne de
référence, puis Browne, All-Weather, Golden Butterfly, Dragon), segments
coloriés non pas par actif mais par **saison couverte** (prospérité,
inflation, déflation, récession/liquidité). Chaque composition de l'article
devient une ligne, et la thèse « ce sont des cousins qui appliquent la même
grille » se lit d'un coup d'œil, tout comme le trou du 60/40 sur la ligne
inflation.

Elle vaut sa place parce que l'article égrène aujourd'hui quatre listes de
pourcentages dans quatre paragraphes, que personne ne peut comparer de tête.
C'est aussi la figure qui rend visible le lien avec `regimes-de-marche`.
Coût A : tous les poids sont dans l'article, le seul travail est la table de
correspondance actif → saison (elle existe déjà en prose dans le paragraphe
Browne). Attention à la poche long volatility du Dragon, à marquer comme non
implémentable (hachure ou libellé), sinon la figure vend un portefeuille
inaccessible.

## 2. L'échange, vu comme un nuage rendement / pire chemin (coût A)

Nuage à deux axes, espérance réelle en x, pire drawdown réel en y, un point
par portefeuille (les six du tableau ci-dessus), famille tous-temps dans une
couleur, échelle actions/obligations dans une autre, étiquettes horizontales
aux points. On voit alors que la famille occupe un coin que l'échelle
actions/obligations n'atteint jamais, à mi-chemin de rendement mais à
drawdown deux fois plus court.

C'est littéralement le bloc `science` de l'article, aujourd'hui écrit en une
phrase de six nombres séparés par des points-virgules, que le lecteur ne
retiendra pas. Le nuage porte en plus la phrase clé « à rendement égal ou
inférieur, les pires chemins sont deux fois moins profonds », qui devient
vérifiable à l'œil. Coût A (chiffres du tableau ci-dessus). Deux précautions,
annoter le point All-Weather comme dépendant de la jambe matières premières
retenue, et rappeler « US, 1972-2024 » dans le sous-titre puisque tout
l'article insiste sur la décote à appliquer.

## 3. L'écart à l'indice, année après année (coût A)

Barres annuelles de l'écart de rendement réel entre un tous-temps (Golden
Butterfly ou Browne) et 100 % actions, 1972-2024, avec les longues séries
négatives grisées en fond (1995-1999, 2013-2019, 2023-2024) et le nombre
d'années consécutives de retard écrit dans chaque bande.

L'article appelle l'écart à l'indice « la critique la plus sérieuse » et la
traite en cinq lignes sans un seul chiffre. Or c'est la raison n° 1 des
abandons, et une figure qui montre cinq ou sept barres rouges d'affilée fait
exactement le travail qu'un paragraphe ne fait pas, préparer le lecteur à la
durée de l'épreuve. Coût A, les deux séries sont déjà calculées ; la seule
décision de forme est de ne pas empiler une courbe de cumul par-dessus, qui
brouillerait le message (le sujet est la longueur des séquences, pas le
total).

## 4. Le curseur, du 100 % croissance au tous-temps complet (coût B)

Balayage de la dose de « poche de régimes » (0 %, 10 %, 20 %, 30 %, 40 %),
deux courbes qui se croisent, espérance réelle qui descend doucement et pire
drawdown réel qui remonte vite quand la dose baisse, avec la zone 30-40 %
marquée comme le plateau recommandé. La figure justifie chiffre en main la
phrase « 0,3-0,6 point d'espérance » de la section usage recommandé, et
montre que le rendement marginal de la dose décroît, ce qui est l'argument
central contre le tout-ou-rien.

Coût B, il faut écrire le balayage (composer le cœur croissance plus la poche
or/linkers/duration longue et itérer sur la dose) ; les briques de données
existent toutes dans le repo. Si l'on veut y mettre le SWR plutôt que le
drawdown, le coût monte encore d'un cran (moteur de retrait), et le drawdown
suffit au propos.

---

## actifs-defensifs

L'article est le hub de la partie portefeuille (il porte la table des défenses et les
verdicts par candidat) et n'a aucune figure aujourd'hui. Ses articles-fils en ont
(`correl-sign`, `trend-smile`, `trend-annees`, les quatre d'`allocation-actions-obligations`).
Une figure ici a donc une vraie valeur de carte.

## 1. Le bulletin de crise, candidat par épisode (coût B, données presque toutes dans le repo)

Une grille signée : en lignes les candidats (actions monde, cash, obligations d'État
courtes, longues, linkers, or, trend, et deux faux défensifs, REIT et high yield) ; en
colonnes les quatre épisodes de test que l'article invoque en prose (1973-1974
stagflation, 2008 panique déflationniste, mars 2020 krach éclair, 2022 choc de taux).
Dans chaque case le rendement réel de l'épisode, coloré en divergent signé.

Pourquoi ça vaut la place : c'est la thèse centrale de l'article (« aucun actif ne
défend contre tout, chacun a son régime et son régime tueur ») et son test infaillible
(« qu'a fait ce produit en 2008 et en 2022 ? »), aujourd'hui dispersés dans dix
paragraphes de revue. La lecture en diagonale rend visible d'un coup que chaque colonne
a un gagnant différent, et que les deux dernières lignes n'ont de bulletin nulle part.
C'est aussi la seule figure qui justifie visuellement le tableau ennemi/titulaire/doublure
qui suit.

Données : `pkg/datasets/refdata` et `simdata` couvrent déjà l'essentiel. Valeurs
calculées en séance sur les séries du repo (années civiles, nominal) :
US long 20 ans +25,8 % / +13,2 % / −23,6 % en 2008/2020/2022 ; TLT +34,0 / +18,2 / −31,2 ;
euro govt 25 ans+ +12,2 / +10,8 / −37,8 ; euro govt agrégé +8,6 / +3,9 / −20,7 ;
US 1-3 ans +6,6 / +3,0 / −3,9 ; or (USD) +5,8 / +24,6 / −0,4 ; backcast trend (CTA)
+17,1 / −29,0 / +18,2 ; S&P 500 −37,0 / +18,4 / −18,1. À déflater par le CPI et à
convertir en euros pour un lecteur français (l'or 2022 devient positif en euros, ce que
l'article dit déjà). Manquent : linkers (reconstruction taux réel + IPC), REIT et high
yield (valeurs d'indices publiées, hors repo), et 1973-1974 pour le trend (backcast).
D'où le B.

## 2. La carte prime contre paiement (coût B, moitié stylisée)

Un nuage à deux axes : en abscisse le coût de portage annuel assumé (de « quasi nul »
pour la duration courte à « −2 à −5 %/an » pour les puts), en ordonnée le paiement en
régime hostile (rendement moyen dans le quartile des pires années actions). Les vrais
défenseurs se rangent en haut ; les faux défensifs (dividendes, high yield, min-vol,
structurés) tombent dans le quadrant « coûte le drawdown des actions et ne paie rien ».

Pourquoi ça vaut la place : c'est l'exigence n° 4 du cahier des charges (« toute vraie
assurance a une prime ») et la galerie des faux défensifs réunies dans un seul plan de
lecture. C'est la figure qui répond à « pourquoi pas 100 % de défense ? » sans un mot.

Coût : l'ordonnée se calcule sur les séries du repo pour cash, duration, or, trend et
actions ; l'abscisse vient des chiffres déjà écrits dans le livre. Les quatre faux
défensifs n'ont pas de série ici et seraient placés en ordres de grandeur, ce qu'il faut
assumer dans la légende (profil stylisé, comme `trend-smile`).

## 3. Sept plus sept contre quinze (coût B)

L'article affirme, sans le montrer, que « 7 % d'or plus 7 % de linkers couvrent mieux
que 15 % d'or seul ». Trois variantes du même portefeuille (défense mono-titulaire or,
défense mono-titulaire linkers, défense partagée) mesurées sur quatre régimes hostiles
(1966-1982 stagflation, 1980-2000 désinflation à taux réels positifs, 2008, 2022), en
barres groupées ou en haltères par régime.

Pourquoi ça vaut la place : c'est le deuxième principe d'assemblage, le moins intuitif
des trois (chaque défense a son mode d'échec, donc la défense se diversifie aussi), et
le seul passage de l'article où une affirmation quantitative est avancée nue. La figure
montre au passage le purgatoire 1980-2000 de l'or, ce qui évite d'avoir à le concéder
en note.

Coût B : il faut une série de linkers longue, donc reconstruite (IPC plus taux réel), et
la reconstruction doit être annoncée comme telle dans la légende, cohérence oblige avec
la mise en garde du livre sur les backcasts.

## Ce qui ne mérite PAS de figure

Le tableau ennemi / titulaire / doublure est déjà la bonne forme pour la section
d'assemblage : le transformer en visuel ferait perdre les libellés. Et la section
« Éprouver la poche défensive en simulation » se lit comme une procédure, pas comme un
résultat ; elle n'a rien à illustrer.

---

## faux-actifs-defensifs

L'article n'a aucune figure et son épine dorsale est un bulletin de crise
candidat par candidat. Tout ce bulletin est aujourd'hui verbal et dispersé dans
sept paragraphes, alors que c'est exactement le genre de matière qui se lit d'un
coup d'œil. Quatre idées, par ordre de valeur.

## 1. Le bulletin de crise en grille, avec groupe de contrôle (coût C, sinon B)

Une grille candidats × épreuves. En lignes, les sept faux défensifs de
l'article, puis un trait de séparation et les vrais défenseurs en groupe de
contrôle (cash, duration d'État longue, or, trend). En colonnes, les trois
épreuves qui structurent le chapitre : le krach 2007-2009, la panique de
février-mars 2020, l'année 2022. Une quatrième colonne pour le rebond
(participation à la reprise), car c'est là que meurent les covered calls et pas
dans la baisse.

Ce que ça porte : la thèse centrale du chapitre, « aucun ne passe le test »,
devient vérifiable ligne par ligne au lieu d'être affirmée. Le groupe de
contrôle est indispensable, sinon le lecteur ne voit pas la différence entre
« tombe un peu moins » et « défend ».

Coût : **C** dans sa meilleure version, un tableau chiffré sourcé (les séries
REIT, high yield, private equity et covered calls ne sont pas dans le repo, et
un tableau assume mieux qu'une figure le mélange de sources). **B** si on veut
une grille dessinée et calculée pour les seules lignes disponibles.

## 2. Le rebond, pas le krach : où meurent les covered calls (coût B)

Deux courbes cumulées de fin 2007 à fin 2012, actions en total return contre
stratégie buy-write, avec deux annotations seulement : le matelas des primes
pendant la chute, puis la reprise 2009 captée seulement aux deux tiers environ.
Le passage « on garde toute la baisse et on plafonne la hausse » est
aujourd'hui un raisonnement ; la courbe qui ne rattrape jamais est la preuve.
C'est aussi le passage le plus utile commercialement, vu la vague QYLD/JEPI.

Coût : **B**. Le repo a le S&P 500 mensuel long, pas d'indice buy-write ; il
faut soit récupérer BXM, soit modéliser un buy-write sur les rendements
mensuels du S&P 500 (plafond de prime constant), en assumant l'approximation
dans la légende.

## 3. Le revenu qui maigrit exactement quand on en a besoin (coût B)

Deux escaliers de revenu annuel, base 100 en 2007, pour un même capital : le
coupon d'un sleeve d'emprunts d'État, et le dividende d'un sleeve actions à
haut rendement. Le coupon reste plat ou monte, le dividende recule d'un
cinquième en 2009 pendant que le capital sous-jacent a été divisé par deux. Le
chapitre affirme « un dividende n'est pas un coupon » ; cette figure le montre
en unités de revenu du rentier, la seule unité qui compte pour le lecteur.

Coût : **B**. Le repo a les Treasuries (TREASURY-INT/LONG) mais pas la série
des dividendes du S&P 500 ; elle est à une extraction près du jeu Shiller déjà
utilisé pour la CAPE.

## 4. Encadré d'honnêteté : « 2022 a été un bon millésime pour trois d'entre eux » (coût C)

Un encart court, pas une figure. En 2022, le min-vol américain a reculé
d'environ 9 % contre 20 % pour l'indice, les fonds à haut dividende ont fini
quasiment à l'équilibre, et les covered calls ont perdu moins que leur indice.
2022 fut un marché baissier lent, avec une volatilité chère et un vent
favorable au style value, donc le millésime le plus flatteur pour ces produits.
Leur vrai régime d'échec est le krach rapide suivi d'un rebond en V, 2020 et
2009. Reconnaître ce point renforce beaucoup le chapitre au lieu de l'affaiblir,
et il désamorce l'objection immédiate du lecteur qui a regardé les
performances 2022.

Coût : **C**, un encadré `attention` ou `science`, aucune donnée à produire.

---

## or-en-retrait

L'article n'a aucune figure aujourd'hui, et deux de ses passages centraux sont
purement verbaux (l'alternance de régimes, et l'A/B or contre pas-d'or). Quatre
idées, par ordre de priorité.

## 1. Les décennies de l'or, en réel (barres signées, une par décennie glissante)

Une barre par décennie de détention (1970-79, 1980-89, ... ou décennies
glissantes tous les cinq ans), hauteur = rendement réel total de l'or sur la
période, plus un trait horizontal à la moyenne (~0 %). Le lecteur voit d'un
coup les deux moitiés de la thèse de l'article : l'espérance est nulle, et
pourtant aucune décennie ne ressemble à zéro, de +250 % réel à −70 %. C'est
l'argument « dormant puis explosif, par régimes » et l'argument « qui le juge
en solo le rejettera toujours » dans une seule image, là où le paragraphe
actuel enchaîne cinq épisodes en liste.

Coût : **A**. `pkg/datasets/simdata/XAUUSD.csv` (quotidien, depuis 1968) et le
déflateur `^CPI-US` déjà embarqué (voir `pkg/replay/reference.go`,
`referenceCPI`). Une simple déflation et des agrégats par décennie.

## 2. L'A/B or contre pas-d'or, modèle par modèle (barres appariées ou dumbbell)

Quatre lignes, une par modèle de marché (central, stress de séquence, inflation
longue, millésime 1966 rejoué), et pour chacune deux points reliés : variante A
(70/30) et variante B (70/20/10 or). Le trait est quasi nul au central et
s'allonge vers la gauche à mesure que le modèle devient hostile. C'est
exactement la thèse la plus importante et la moins visible de l'article
(« l'A/B ne se voit pas dans le scénario central, il apparaît sur les modèles
qui contiennent les longues inflations »), aujourd'hui noyée dans le bloc
`exemple`. Prévoir une note sur le coût, la richesse médiane à −5 %, pour ne
pas laisser croire à un repas gratuit.

Coût : **A**. Les huit chiffres sont déjà dans le bloc `exemple` de l'article
(4,1/3,9 ; 7,2/6,1 ; 10,8/8,9 ; 1966 épuisé/traversé). Aucun calcul nouveau si
l'on illustre l'exemple tel qu'il est écrit.

## 3. L'or dans les pires années des actions (barres classées par année d'actions)

Les dix ou quinze pires années des actions mondiales, classées, et pour chacune
le rendement de l'or et celui des obligations d'État longues. La lecture porte
le critère n° 2 du cahier des charges défensif : la décorrélation de l'or
survit au stress (2008, 2020, 2022), et elle échoue là où l'on l'attend
(1981), alors que la ligne obligataire, elle, se retourne franchement en 2022.
Plus honnête qu'un chiffre de corrélation moyenne, qui ne dit rien des queues.

Coût : **B**. Les séries existent (`XAUUSD`, `SP500-USD` ou `MSCIWORLD-USD`,
`TREASURY-LONG-USD` dans `refdata`, plus `^CPI-US`), mais l'alignement annuel
et le choix nominal/réel/devise sont un vrai calcul à écrire, avec le piège
habituel : en euros et en dollars le verdict de 2022 n'est pas le même.

## 4. Les deux régimes fiscaux de l'or physique : où est la bascule (tableau)

Un petit tableau à trois colonnes (multiple de plus-value, impôt sous la taxe
forfaitaire sur le prix de vente, impôt sous le régime réel avec abattement
pour durée de détention) sur trois ou quatre lignes de multiples et deux durées
de détention. Le paragraphe actuel énonce les deux régimes et conclut « le
régime réel est presque toujours meilleur pour une détention longue » sans
jamais montrer la bascule, qui est précisément ce que le lecteur veut savoir
avant de choisir son option de sortie.

Coût : **C** (un tableau vaut mieux qu'une figure ici). Attention : à ne
construire qu'après avoir revérifié les deux taux dans les sources fiscales à
jour, le taux du régime réel cité dans l'article étant douteux (voir le rapport
de review).

---

## obligations-en-retrait

Contexte : l'article porte déjà `::: figure correl-sign` (corrélation glissante
actions-obligations). Deux formes voisines sont DÉJÀ prises par
allocation-actions-obligations et ne doivent pas être redemandées ici :
`duration-choc` (duration × choc de taux, les cinq douleurs) et
`obligations-regimes` (cinq espèces × deux chocs, 2008/2022). La mécanique
duration et la grille de régimes sont donc couvertes ailleurs ; les idées
ci-dessous portent ce que cet article a en propre.

## 1. « Tenir à échéance ne protège de rien » : deux comptabilités, une richesse

Deux trajectoires de richesse après une hausse de taux de +2 points au mois 12,
sur un horizon égal à la duration (7 ans). Le fonds à duration constante
plonge tout de suite en prix puis rattrape par le réinvestissement des coupons ;
le titre détenu à échéance ne baisse jamais à l'écran mais reste bloqué sur son
ancien coupon. Les deux courbes se rejoignent à l'échéance, et l'écart intermédiaire
est étiqueté d'un côté « baisse de prix affichée », de l'autre « coût
d'opportunité invisible ».

Pourquoi elle vaut la place : c'est l'illusion la plus tenace de tout l'article
(« je ne perds jamais »), et le passage qui la démonte est aujourd'hui à 100 %
verbal, sans aucun appui visuel dans le livre. Une figure où les deux courbes
convergent règle le débat en une seconde, là où le paragraphe demande un effort
de foi. Forme : deux courbes qui convergent, avec la zone d'écart hachurée et
nommée dans les deux sens.

Coût : **B** (arithmétique obligataire déterministe à écrire, aucune donnée de
marché nécessaire ; un fonds de duration constante et un titre à coupon fixe).

## 2. Le YTM est une prévision : rendement affiché contre rendement encaissé

Nuage de points : rendement à l'échéance du 10 ans américain à la date t (axe
x) contre rendement annualisé réellement encaissé sur les 7 à 10 années
suivantes (axe y), mois par mois depuis les années 1950, avec la diagonale 45°
tracée. Les points s'y collent, avec un biais visible seulement autour des
grands retournements de régime.

Pourquoi elle vaut la place : elle démontre la deuxième « idée qui suffit », le
cœur de l'article, à savoir que l'obligation est le seul actif dont l'espérance
est lisible sur l'étiquette. Aucun autre passage du livre ne prouve cette
affirmation, qui est pourtant réutilisée dans [[rendements-attendus]] et dans
tous les paramétrages de plan. C'est aussi la seule figure du livre qui puisse
montrer une prévision qui marche, ce qui contraste utilement avec les figures
d'incertitude (fans, Monte Carlo).

Coût : **B** (données présentes : `pkg/datasets/macropanel/oecd-monthly.csv`
donne `longrate` mensuel par pays, `refdata/TREASURY-INT-USD.csv` donne le
total return ; le calcul des paires t → t+7 ans reste à écrire). Variante euro
possible avec `longrate` FRA/DEU et `refdata/EUROGOV-EUR.csv`.

## 3. Le retard du fonds euros, dans les deux sens

Deux lignes de 1995 à aujourd'hui : le taux servi moyen des fonds euros et le
taux de marché de référence (OAT 10 ans, ou rendement de l'aggregate euro). Le
fonds euros passe vingt ans AU-DESSUS du marché pendant la baisse des taux,
puis nettement EN DESSOUS à partir de 2022. La zone entre les deux courbes est
coloriée selon son signe, « le lissage vous paie » puis « le lissage vous
coûte ».

Pourquoi elle vaut la place : la section fonds euros affirme un mécanisme de
retard (« le temps que le stock obligataire de l'assureur tourne ») et en tire
la limite d'emploi, la moitié de la poche au maximum. C'est le passage le plus
spécifiquement français du livre, aujourd'hui purement verbal, et le retard est
exactement le genre de fait qu'une image prouve mieux qu'un chiffre isolé. Elle
sert aussi [[cash-buffer]] et [[enveloppes-francaises]], qui s'appuient sur le
même mécanisme sans le montrer.

Coût : **B** (le taux de marché est dans le repo via `macropanel` FRA
`longrate` ; la série des taux servis moyens n'y est pas et demande une
quinzaine de valeurs annuelles saisies à la main depuis les rapports de la
profession, à sourcer explicitement).

## 4. Le cahier des charges de la poche, en tableau

Un tableau de récapitulation à la fin de la section des quatre décisions, une
ligne par brique et cinq colonnes : brique, service rendu (1, 2, 3), duration,
régime qui la sert, régime qui la casse, part indicative de la poche. Il fixe
en une grille ce que l'article déroule en quatre paragraphes plus un exemple
chiffré.

Pourquoi il vaut la place : la promesse de l'article est « votre poche
obligataire aura un cahier des charges écrit », et ce cahier des charges
n'existe nulle part sous forme consultable. Le lecteur qui revient six mois
plus tard cherche cette grille, pas les paragraphes. Attention au recouvrement
avec la table des rôles de [[actifs-defensifs]] : ici on ne liste que
l'intérieur de la poche obligataire, avec la colonne duration qui est propre à
cet article.

Coût : **C** (pas de figure ; un tableau, entièrement dérivé du texte existant
et de l'exemple chiffré).

---

## obligations-indexees

L'article n'a aucune figure aujourd'hui, alors qu'il porte deux passages
purement verbaux très visuels : le résultat de l'échelle (bloc `science`) et la
décomposition de 2022. Trois idées, par ordre de valeur.

## 1. Ce que finance une échelle de linkers, selon le taux réel et l'horizon

**Ce qu'elle montre.** Trois courbes du taux de retrait garanti par une échelle
qui se consomme entièrement, `r / (1 − (1+r)^−N)`, pour N = 20, 30 et 40 ans, en
fonction du taux réel de marché de −1 % à +3 %. Une horizontale à 4 % (la règle)
et quatre repères pointés sur la courbe 30 ans : 0 % réel → 3,3 %, 1 % (euro) →
3,9 %, 2 % → 4,5 %, plus le niveau TIPS du moment (~2,3 % → ~4,7 %). C'est la
figure qui porte toute la thèse du bloc `science` : l'échelle bat la règle des
4 % dans une fenêtre de taux réels, et seulement là. La courbe 40 ans dit
visuellement pourquoi un candidat FIRE ne peut pas s'en contenter (à 1 % réel,
elle ne finance que ~3,0 %), ce que le texte affirme sans le chiffrer. Elle
remplace aussi les trois chiffres épars de l'article (3,3 / 3,9 / 4,5) par un
seul objet lisible.

**Coût : A/B.** Formule fermée, aucune donnée externe ; quelques lignes de
calcul dans le générateur de figures, plus le taux réel courant à citer en
annotation datée.

## 2. 2022, les deux moteurs d'un linker, par duration

**Ce qu'elle montre.** Une cascade par bucket de duration (2, 5, 8, 12) avec
trois barres chacune : l'indexation créditée (+8-10 %, identique partout), l'effet
prix des taux réels (−duration × Δtaux réel), et le total. Même année, même
indexation, résultats de +6 % à −30 % selon la seule duration. C'est le mode
d'emploi que l'article énonce en prose (« la protection joue à l'échéance »), et
la figure rend l'argument non discutable : elle montre que le coupable est la
duration, pas le contrat d'indexation. Elle sert aussi de garde-fou à la
doctrine de dose (durations courtes en poche, détention à terme pour le
plancher).

**Coût : B.** Il faut la variation des taux réels euro 2022 par segment de
maturité (source externe, ou reconstruction depuis les ETF linkers du catalogue
`assetmeta` et la série IPCH pour la jambe indexation). Le calcul est simple,
mais la donnée de taux réels n'est pas dans le repo aujourd'hui.

## 3. Le plancher adossé du bloc `exemple`, en bilan plus escalier de revenu

**Ce qu'elle montre.** Deux panneaux côte à côte. À gauche, le patrimoine de
1,6 M€ coupé en deux, 520 k€ d'échelle indexée (14 barreaux de 40 k€ réels) et
1,08 M€ de portefeuille de confort. À droite, le revenu année par année de 52 à
75 ans : les barreaux consommés jusqu'à 66 ans, le relais des pensions ensuite,
et par-dessus le filet de confort de 15 k€ servi par le portefeuille. Le passage
est le plus concret de l'article et le plus difficile à tenir en tête en lecture
linéaire, car il mêle un stock, un flux et une date de relais. La figure fait
voir d'un coup pourquoi le taux de retrait du solde tombe à 1,4 %.

**Coût : A.** Tous les chiffres sont dans le bloc `exemple` et ils tombent juste
(annuité 14 ans à 1 % réel = 520 k€, 1,6 − 0,52 = 1,08 M€, 15 / 1 080 = 1,4 %).

## Écarté

Le point mort contre l'inflation réalisée ensuite (nuage ou double courbe) serait
la bonne figure pour la section « mécanique », mais la série de taux réels
manque dans le repo et l'histoire utile est courte (2003 pour les TIPS, moins en
euro). Un encadré chiffré sur deux ou trois dates de départ ferait le même
travail pour moins cher (coût C).

---

## facteurs-fama-french

L'article n'a aucune figure et porte deux thèses très visuelles : « le tilt ne
donne pas plus, il donne d'autres décennies perdues » et « le prix d'entrée,
c'est dix ans d'écart à l'indice ». Trois idées, par ordre de priorité.

## 1. L'écart glissant sur dix ans : SCV moins marché large (coût A)

Une seule courbe, l'écart de CAGR réel sur dix ans glissants entre le
small-cap value américain et le S&P 500, de 1973 à aujourd'hui (fenêtres de
dix ans à partir de 1963), en aire divergente autour de zéro. Au-dessus de
zéro, les décennies où le tilt a payé (la fenêtre qui se termine vers 2010,
sommet du dossier de l'an 2000) ; en dessous, les décennies de honte, dont
2010-2020, la plus longue et la plus profonde. Trois annotations horizontales
suffisent : « décennie perdue du marché large », « purgatoire de la value »,
« 1999 ».

C'est la figure qui porte le bloc-clé (« pas les mêmes décennies perdues »),
l'encadré du prix d'entrée et le verdict, aujourd'hui tous les trois purement
verbaux. Elle montre aussi, sans le dire, que l'aire au-dessus et l'aire en
dessous s'équilibrent à peu près : exactement la thèse « déplacement, pas
supplément ».

Coût A : tout est embarqué. `pkg/datasets/refdata/USSCV-USD.csv` est le
portefeuille small × high book-to-market de Ken French en quotidien depuis
1963-07, `refdata/SP500-USD.csv` le S&P 500 total return en fin de mois, et le
déflateur CPI est celui que `pkg/replay/reference.go` utilise déjà. Repli en
mensuel pour les deux, puis fenêtres de 120 mois.

Variante si l'aire divergente paraît trop abstraite : les deux CAGR réels
glissants en courbes qui se croisent, avec la zone entre elles remplie selon le
signe. Plus lisible pour un lecteur non technique, mais deux fois plus d'encre
pour la même information.

## 2. Le verdict d'ERN en quatre nombres (coût C, un tableau)

Pièce 2 repose maintenant sur quatre nombres du volet 62 (taux sûr moyen
3,39 % sans tilt, 3,27 % avec du SCV à alpha nul, 3,51 % avec des primes
futures généreuses, −0,22 point sur les départs des années 1930). En prose,
ils passent vite et le lecteur retient « le SCV aide ». En tableau de trois
lignes (hypothèse de prime / taux sûr moyen / effet sur les pires millésimes),
ils deviennent un garde-fou.

Pas de figure ici : quatre nombres et deux hypothèses ne remplissent pas un
graphique, et un histogramme de trois barres presque identiques suggérerait à
tort de la précision. Un tableau ou un encadré « ce que dit vraiment le volet
62 » fait mieux.

## 3. Value moins growth, par régime d'inflation (coût B)

Cinq ou six barres horizontales, l'écart de rendement annuel value moins
growth dans chaque régime (inflation basse et stable, inflation en
accélération, choc inflationniste, désinflation, récession déflationniste),
avec les époques nommées sous chaque barre (1970-1981, 2021-2022, 2010-2020,
1929-1932). La barre du choc inflationniste est celle qui justifie la
« demi-brique défensive » de la pièce 3, et celle de la récession déflationniste
rappelle le revers.

Cette pièce 3 est le seul argument pro-tilt qui survit au volet 62, et elle
est aujourd'hui appuyée sur une seule année citée de mémoire (2022). Une figure
de régimes la rend vérifiable, ou la dégonfle, ce qui serait aussi utile.

Coût B : il faut une série growth en face de la value (les portefeuilles
2x3 de Ken French sont déjà téléchargés pour `USSCV-USD`, le growth vient du
même fichier) et une définition de régime, celle du `pkg/datasets/macropanel`
ou, plus simplement, des seuils sur le CPI américain. Le classement des
périodes doit être écrit avant de regarder les rendements, sinon la figure
n'est qu'un ajustement rétrospectif.

---

## diversification-internationale

L'article n'a aucune figure et il est pourtant le plus « chiffré » de la partie
portefeuille. Quatre idées, par ordre de valeur décroissante.

## 1. Le SAFEMAX du destin national, pays par pays, contre celui du monde

Barres horizontales classées : pour chacun des 18 pays du panel JST, le taux de
retrait maximal soutenable sur 30 ans d'un 50/50 domestique (le pire millésime
du pays), et une règle verticale unique pour le même calcul sur un panier
mondial équipondéré. Le lecteur voit d'un coup que la moitié des pays est sous
2 %, que la France y est, et que la barre mondiale se pose vers 3-3,5 %.

C'est LA figure de l'article, parce que le paragraphe « L'effet sur les taux de
retrait » est aujourd'hui la thèse centrale entièrement en prose, et parce
qu'elle transforme un argument moral (« diversifiez ») en écart chiffré et
comparable. Elle sert aussi de charnière vers [[anarkulova-cederburg]].

Coût : **B**. Les données sont dans le repo (`pkg/datasets/broadsample/country-real.csv`,
rendements réels actions/obligations annuels par pays) et la mécanique de
retrait existe (`pkg/decumul`, `pkg/replay`) ; il reste à écrire le panier
mondial équipondéré et la boucle SAFEMAX par pays. Pièges : dire dans la
légende que les SAFEMAX de Pfau cités dans le texte viennent d'une autre base,
donc que les valeurs ne coïncideront pas au dixième ; et le panel JST donne un
rendement réel français de −0,1 %/an sur 1900-2020, bien pire que les 3,3 % de
Dimson-Marsh-Staunton cités plus haut dans l'article, écart à assumer
explicitement sous la figure.

## 2. Trente ans sous l'eau, quatre destins

Petites multiples de courbes « sous l'eau » en termes réels (perte depuis le
dernier sommet réel), 1900-2020, mêmes axes : France, Japon, Italie, panier
mondial équipondéré. Aires remplies. On lit directement le trou français
d'après-guerre, les trente ans japonais post-1990, et le fait que le panier
mondial remonte toujours en une décennie et des poussières.

Elle porte le paragraphe « Les écarts de siècle », dont l'idée est justement
que la moyenne annualisée cache les traversées. Un tableau de CAGR ne peut pas
montrer une durée d'immersion ; c'est le seul objet graphique qui le fait.

Coût : **B**. Mêmes données que l'idée 1, plus un cumul et un max courant.
Attention à ne pas doublonner avec la longue série annotée déjà utilisée
ailleurs dans le livre : ici la variable est la profondeur ET la durée, pas le
niveau.

## 3. Ce que le change a vraiment fait à l'investisseur euro

Barres divergentes classées : pour une dizaine d'années marquantes, l'écart en
points entre le MSCI World en euros et le même indice en dollars. Les barres
positives (2000 +6,9 ; 2008 +3,4 ; 2022 +5,1) et les négatives (2002 −12,8 ;
2003 −22,6 ; 2020 −9,8) dans la même image.

Elle sauve la section change d'un procès en angélisme : l'amortisseur existe et
se voit, mais il n'est pas systématique, et la baisse de 2000-2003 a coûté
7 points de plus à l'investisseur euro. C'est exactement la nuance que le texte
vient d'acquérir, et une figure la rend intuitive là où une phrase reste
abstraite.

Coût : **B**. `pkg/datasets/refdata/MSCIWORLD-USD.csv` est dans le repo (mensuel
fin de mois), l'EUR/USD non : il faut le tirer de DBnomics comme le fait déjà
`cmd/gen-euro-refdata` (série ECB/EXR quotidienne), ou l'ajouter en refdata.
Les chiffres ci-dessus sont déjà calculés et vérifiés.

## 4. L'atlas de Karim contre l'atlas du monde

Trois barres empilées côte à côte, mêmes catégories géographiques : le
portefeuille de Karim avant, après, et la capitalisation mondiale comme
référence. Le « 75 % Europe plus France pour environ 17 % du marché » devient visible
sans une seule ligne de calcul mental.

Coût : **A**. Tous les chiffres sont dans le bloc `::: exemple` de l'article.
Réserve honnête : c'est la moins indispensable des quatre, parce que le bloc
exemple se lit déjà bien. À ne faire que si la partie portefeuille manque de
respiration visuelle.

---

## etf-ucits-europeens

Article aujourd'hui sans aucune figure. Quatre idées, par ordre de valeur.

## 1. La chaîne des cinq maillons de coût, deux portefeuilles sur le même axe (coût A)

Deux barres horizontales segmentées sur une échelle commune de 0 à 2,5 %/an. En
haut le portefeuille UCITS propre (TER, écart de suivi interne, spread,
courtage, enveloppe ≈ 0), en bas le même portefeuille en unités de compte
chargées ou en fonds actifs de réseau. Chaque segment porte son nom et sa
fourchette. La figure montre ce que le paragraphe met dix lignes à dire : les
quatre premiers maillons sont des poussières comparées au cinquième, la couche
d'enveloppe, et c'est donc la seule décision de tuyauterie qui pèse vraiment.
Elle vaut sa place parce que le lecteur français arrive avec l'intuition
inverse (il compare des TER et signe un contrat à 1 %). Les chiffres sont tous
dans l'article. Une annotation à droite convertit l'écart total en points de
taux de retrait (0,5 à 1 point sur 40 ans) ; l'affirmer d'après l'article est
coût A, la recalculer sur le 60/40 réel de `pkg/replay` en ajoutant un drag de
frais constant est un petit coût B, et ce serait mieux.

## 2. La grille d'implantation : sept lignes, trois enveloppes, sept rôles (coût A)

Une matrice. En lignes les sept briques dans l'ordre des rôles (moteur monde,
tilt SCV, cœur amortisseur, linkers, or, trend, buffer), en colonnes les trois
enveloppes PEA / AV / CTO. Chaque cellule est un carré ou une barre
proportionnelle au poids (19, 38, 8, 11, 6, 5, 5, 8 %), et la marge de droite
nomme le rôle servi dans la table des défenses plus le critère « à vérifier ».
Le paragraphe d'exemple est aujourd'hui un mur de huit pourcentages et trois
montants que personne ne reconstitue de tête ; la grille rend d'un coup d'œil
les deux thèses de l'article, à savoir que chaque ligne a un rôle nommé et
qu'une enveloppe donnée n'accueille que ce qu'elle loge proprement. C'est aussi
le récapitulatif que le lecteur photographiera. Données entièrement dans
l'article (vérifier que la somme fait 100 % dans la figure comme dans le
texte).

## 3. Le trajet d'un dividende américain selon le domicile (coût C, encadré)

Quatre nombres : 100 € de dividende américain deviennent 85 € dans un fonds
irlandais et 70 € dans un fonds luxembourgeois, soit, sur 1,3 % de rendement de
dividende, environ 0,2 %/an d'écart permanent. Une figure ne ferait que
décorer une soustraction. Un petit encadré chiffré à deux colonnes (Irlande /
Luxembourg) posé à côté du « Choix 3 » suffit, et il a l'avantage de rendre le
mécanisme mémorisable au moment du choix de l'ISIN.

## 4. La base taxable d'une vente contre celle d'un dividende, au fil du plan (coût B)

Deux courbes sur 30 ans, en base imposable et non en impôt, pour un même flux
annuel encaissé. Le dividende est une horizontale à 100 % du flux. La vente
suit la fraction de gain (1 − PMP/cours), qui part bas et dérive vers le haut
sans jamais rejoindre l'horizontale. C'est la thèse de l'encart d'ouverture
(« à flux égal, la vente est toujours moins taxée ») et celle du paragraphe sur
le PMP, aujourd'hui purement verbales. Tracer la base et non la charge évite
d'inscrire un taux d'imposition français dans une figure, qui vieillirait mal.
Coût B : il faut une petite simulation de PMP sous un chemin de rendement (les
séries du repo suffisent). Attention, l'article de fiscalité couvre le même
mécanisme et n'a pas de figure ; à arbitrer entre les deux, une seule fois.

## Écartée

Un nuage TER contre tracking difference sur une quinzaine d'ETF monde, qui
montrerait que les deux classements diffèrent. La thèse est juste et centrale,
mais les données ne sont pas dans le repo, elles se périment en un an et une
figure datée dans un livre imprimé se retourne contre l'argument.
