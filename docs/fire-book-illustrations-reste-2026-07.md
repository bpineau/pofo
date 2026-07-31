# Firebook: figure backlog for the remaining parts

Status: ACTIVE backlog, opened 2026-07-30 at the close of the book-wide
line-by-line review (one reviewer per article; the prose fixes shipped as the
"line review of ..." commit series). It gathers the illustration ideas for
every part reviewed after the portfolio one: alternatives, buffers, inflation,
French tax frame, human factor, references, withdrawal science and the starter
part. Thirty-seven ideas were built over 2026-07-30 and 2026-07-31, figures and
reference tables both, and their entries have been removed from this file, so
what remains is exactly what has NOT been built.

What is left is a pool of ideas to RE-EVALUATE when a figure is wanted, not a
to-do list: several were written before their article had any illustration, and
some are weak or superseded. Delete this file once the pool stops being useful.
The portfolio part has its own file,
`fire-book-illustrations-portefeuille-2026-07.md`.

Figures, when built, follow the v2 plate system and the frozen-array +
guard-test pattern recorded in `fire-book-design.md`. Costs: **A** = data
already in the repo or the article, **B** = a computation to write first,
**C** = a table or callout does the job better. Written in French because the
book is. `la-machine-pofo` has no entry: its reviewer was lost to a session
limit before writing one.


---

# Les actifs alternatifs


## managed-futures

L'article porte déjà deux figures (`trend-smile`, le profil stylisé en sourire ;
`trend-annees`, le quart de siècle de SG Trend en barres annuelles avec le lavis
de l'hiver). Elles couvrent la thèse « les deux queues paient » et la thèse
« l'hiver fait partie du contrat ». Trois passages restent purement verbaux
alors qu'ils portent une thèse contre-intuitive.

## 2. Le même actif vu par trois indices, petit tableau (coût C)

Un tableau de quatre lignes (2008, 2014, 2020, 2022) et trois colonnes
(SG Trend, SG CTA, BTOP50). Il montre noir sur blanc que l'écart entre
référentiels atteint une dizaine de points sur la même année de crise (2022 :
+27 % contre environ +14 % pour l'industrie large), ce qui est la
démonstration du paragraphe « le bon usage est affaire d'appariement » et de la
règle de contrôle répétée trois fois dans l'article. Un lecteur qui compare son
fonds au mauvais indice conclut faux, et c'est le seul endroit du livre où ce
piège se voit chiffré.

Coût C : pas une figure, un tableau de douze cellules, plus lisible qu'un
graphique. Mais les chiffres viennent de l'extérieur du repo (Société Générale
et BarclayHedge), donc il faut les relever et les dater proprement.

## 3. La décomposition « cash + prime − frais », deux régimes de taux (coût A)

Deux colonnes empilées côte à côte, régime de taux courts nuls (2015-2021) et
régime de taux courts à 4 %, chacune décomposée en collatéral rémunéré, prime
de trend brute et frais retirés, pour aboutir au chiffre affiché (~2 % contre
~6 %). Le passage « le cash travaille aussi » est le plus décisif de l'article
pour lire une plaquette, et il est aujourd'hui purement arithmétique dans le
texte. La figure fait comprendre en un regard qu'une grande part de la
déception de l'hiver et de l'embellie récente ne vient pas de la stratégie.

Coût A : les trois nombres sont déjà dans l'article, la série de taux courts du
repo ne sert qu'à dater les deux régimes.


## long-volatility

L'article a déjà `longvol-profil` (valeur d'une poche de puts dans le temps :
saignée linéaire, deux pics de convexité 2008 et 2020). Elle couvre bien le
« comment ça se comporte ». Ce qui reste purement verbal, ce sont les deux
arguments qui décident vraiment de la réponse : la comparaison avec « détenir
moins d'actions » (Israelov) et la division du travail avec le trend.

## 2. Krach rapide contre marché baissier lent : qui paie quoi (coût B)

Petite matrice de barres, quatre crises en lignes (1987, 2000-2002, mars 2020,
2022) et quatre défenses en colonnes (obligations longues, or, trend, poche de
puts), chaque case portant la performance de l'épisode. On voit d'un bloc que
les colonnes s'allument en damier : les puts sur les deux krachs rapides, le
trend sur les deux crises lentes, jamais les deux ensemble. C'est la thèse du
bloc `science`, aujourd'hui livrée en prose, et c'est aussi l'argument qui
justifie l'ordre de priorité du livre (trend d'abord, long vol en appoint).
Coût **B** : trois colonnes sur quatre sortent du repo
(`TREASURY-LONG-USD`, l'or de `simdata`, `simdata/DBMF.csv` pour le trend) ;
la colonne puts n'a pas de série et doit rester un ordre de grandeur assumé,
signalé dans la légende. Si cette approximation gêne, la même information
passe très bien en tableau, et l'article en a déjà un.

## 3. Monétisé contre gardé « au cas où » (coût B, en extension de la figure existante)

Plutôt qu'une figure de plus, ajouter à `longvol-profil` une seconde courbe
après le pic de 2020 : celle de la poche qu'on n'a pas vendue, qui redescend
au niveau de la saignée en quelques mois. La discipline finale de l'article
(« ne détenez jamais un actif d'assurance sans avoir écrit quand vous
vendrez ») est la seule chose qu'un lecteur doit retenir s'il n'en retient
qu'une, et le graphique actuel la mentionne en légende sans la montrer.
Coût **B** : profil stylisé, cohérent avec la courbe déjà tracée ; deux
étiquettes de bout de courbe, pas de nouvelle donnée.

## Écartée

Le saignement des ETP VIX (−99 % en échelle log) serait spectaculaire, mais la
donnée n'est pas dans le repo et la phrase de l'article suffit déjà à couper
l'envie. Une figure de plus sur un produit qu'on interdit serait de la place
donnée au mauvais sujet.


## global-macro

L'article a déjà `carry-courbes`, qui explique bien la **mécanique** du roll sur
une échéance. Ce qui manque, ce sont les deux conséquences qu'il énonce en prose
seulement : le coût **cumulé** du contango, et l'impossibilité d'acheter la
catégorie macro. Les trois idées ci-dessous visent ces passages.

## 1. Le prix cumulé du contango : spot en hausse, tracker à plat (coût B)

Deux courbes cumulées depuis les années 1990 sur une matière première large ou
un panier : le prix **spot** d'un côté, la valeur d'un tracker long-only qui
roule ses positions de l'autre. La divergence est la thèse du passage « des
décennies de performance décevante malgré des prix spot en hausse », aujourd'hui
purement affirmée. C'est aussi le seul endroit du livre où le lecteur peut
mesurer combien coûte un roll subi, alors que `carry-courbes` ne montre qu'une
seule échéance.

Coût **B** : le repo a des séries de prix (WTI, or), pas de courbes de contrats
à terme, donc pas de rendement roulé reconstructible directement. Il faut soit
un indice spot contre un indice excess-return d'une même famille (source
externe à vérifier et à citer), soit une reconstruction front-month documentée.
À ne faire que si les deux jambes viennent de la même famille d'indices, sinon
la divergence mesure autre chose.

## 2. La dispersion qui interdit d'acheter la catégorie (coût B)

Un axe unique de rendement annualisé, deux bandes empilées : les fonds macro,
et les programmes trend en dessous. Sur chaque bande, le premier quartile, la
médiane, le dernier quartile, plus un repère pour l'indice de catégorie
(HFRI Macro, SG Trend). La lecture est immédiate. La bande trend est étroite et
son indice est au milieu, la bande macro est deux fois plus large et son indice
ne représente aucun fonds achetable. C'est exactement l'argument « pas d'indice
réplicable » que reprend [[actifs-defensifs]], et il est bien plus convaincant
vu que lu.

Coût **B** : les quartiles de dispersion ne sont pas dans le repo et demandent
une source citable (rapports HFR, études SG). Sans source solide, ne pas
inventer les bandes ; l'idée 3 dit la même chose sans chiffres fabriqués.

## 3. Ce qu'il reste d'une prime alternative, par étage (coût C, tableau)

Le passage décisif est arithmétique : 4 % brut, moins 1,5 % de frais, moins
31,4 % de fiscalité, soit environ 1,7 % net. Trois lignes en tableau (trend
UCITS, multi-primes ARP, actions mondiales en repère), avec brut, frais,
fiscalité et net, disent tout et se relisent d'un coup d'œil. Une figure n'ajoute
rien ici, et le livre a déjà `cascade-4pct` pour la forme cascade et
`primes-echelle` pour les primes brutes en barres d'intervalle. Un tableau évite
la troisième variation sur le même geste.

Coût **C** : tous les chiffres sont déjà dans l'article et dans
[[primes-de-risque]].


## return-stacking

L'article a déjà `stacking-expo` (trois façons d'investir 100 €), qui porte bien la
mécanique d'exposition. Ce qui reste purement verbal, c'est l'arithmétique du
financement, l'anatomie de 2022 et le drag. Trois idées, plus une option sans figure.

## 1. La porte d'entrée de tout étage obligataire : la prime au-dessus du cash (coût A/B)

Une seule série longue, 1930-2025 : le rendement excédentaire annualisé glissant sur
5 ans des Treasuries intermédiaires moins celui des T-bills, remplie en positif d'un
côté du zéro et en négatif de l'autre, avec 1981-2021 et 2020-2022 annotés. Elle porte
la thèse la plus importante et la plus abstraite de l'article, aujourd'hui tenue par
une seule phrase (« empiler des obligations qui rendent 3 % quand le taux court est à
3,5 % détruit de la valeur ») : un étage empilé ne touche jamais un rendement absolu,
seulement l'écart au cash, et cet écart a été négatif sur des fenêtres entières. C'est
aussi la figure que [[primes-de-risque]] appelle implicitement quand il dit que
l'étage obligataire ne mérite sa place que quand la prime de terme existe.
Données : séries treasury TR + `^IRX`/bills déjà dans `pkg/datasets` (A pour les
niveaux, B si on veut le rendement excédentaire réalisé glissant, ~30 lignes de calcul).

## 2. Anatomie de 2022, en cascade (coût A)

Une cascade qui part du 60/40 nu (−17 %) et arrive au plan empilé de l'encadré
d'exemple : coût du levier obligataire (environ −4 points), apport du trend (+4 à
+5 points), apport de l'or en euros (+0,5 point), arrivée vers −10 %. Elle règle
visuellement le point que le texte vient de corriger, et qui se lit mal en prose : le
levier pique bien, mais c'est le diversifiant libéré qui paie la facture, donc le plan
empilé n'est pas « celui qui perd plus en 2022 », il est celui qui dépend du trend
cette année-là. Toutes les valeurs existent déjà (NTSX 2022 = −25,8 % en USD, ≈ −21 %
pour un porteur en euros, trend +20 à 30 % selon [[obligations-en-retrait]]).
Forme déjà présente dans le livre, mais c'est ici la bonne : une décomposition
additive d'un résultat annuel.

## 3. Grille levier × corrélation : où le drag mange la prime (coût B)

Petite grille 5 × 5, levier de 1,0 à 2,0 en lignes, corrélation actions-obligations de
−0,4 à +0,6 en colonnes, cellules colorées par le rendement géométrique réel d'un
60/40 levé (μ et σ des jambes fixés et affichés en légende). On y voit d'un coup d'œil
que la zone verte s'arrête net quand la corrélation devient positive, et que le levier
1,5 reste inoffensif tant qu'elle est négative. C'est la démonstration du paragraphe
« frein de la volatilité » et de la règle de dose (≤ 1,5, ≤ 1,3 si le plan est tendu),
qui sont aujourd'hui affirmées sans support. Calcul fermé, quelques dizaines de lignes,
mais il faut assumer et afficher les hypothèses de μ/σ.

## 4. Ce qui existe vraiment en UCITS (coût C, pas de figure)

Le passage européen gagnerait un petit tableau plutôt qu'un graphique : trois lignes
« cœur levé actions-obligations », « actions + trend empilés », « obligations + trend
empilés », deux colonnes « existe en UCITS » et « existe aux États-Unis », avec l'encours
en ordre de grandeur. Le contraste (le cœur levé existe, y compris en version zone
euro ; les piles trend n'existent qu'outre-Atlantique) est un fait d'inventaire, pas une
courbe. Un tableau le dit mieux et vieillit plus honnêtement.


---

# Buffers et protections


## cash-buffer

Existant : une figure, `buffer-flat` (ruine ~plate en fonction de la taille du
matelas). Elle porte bien la thèse quantitative centrale. Ce qui reste
purement verbal, et qui vaudrait une image, tient en trois points.

## 2. Le balayage à deux sorties : ruine plate, richesse qui fond (coût B)

Ce qu'elle montre : le même balayage que `buffer-flat` (matelas de 0 à 10 ans,
prélevé sur le capital de départ), mais avec les deux sorties côte à côte, la
probabilité de ruine et la richesse finale médiane, chacune indexée à son
niveau sans matelas. La ruine trace une ligne quasi plate, la richesse médiane
une pente franchement descendante.

Pourquoi elle vaut la place : l'article affirme que « le vrai arbitrage du
matelas est une assurance payée en patrimoine terminal plutôt qu'en probabilité
de survie », et c'est la seule phrase du texte que la figure existante ne
soutient pas, puisqu'elle ne montre que la ruine. Deux courbes qui divergent
disent cette asymétrie mieux qu'un paragraphe. Elle rend aussi visible que la
neutralité du matelas est une neutralité *de ruine*, pas de patrimoine.

Coût : un balayage à écrire (moteur de décumulation, matelas prélevé, rendement
réel du support non nul), puis une plaque à deux séries. Variante moins chère si
le calcul coince : ajouter la seconde courbe à la figure `buffer-flat` existante
plutôt que d'en créer une.

## 3. Le coût de portage par véhicule (coût C, un tableau suffit)

Le passage « 4. Le placement » avance que le fonds euros réduit de moitié le
coût d'opportunité du buffer. Un petit tableau à quatre lignes (compte courant,
livret réglementé, monétaire €STR, fonds euros) et deux colonnes (rendement
réel attendu, écart au rendement du portefeuille en points réels) tranche la
question en six lignes de texte, et il chiffre au passage l'ancre du livre, 2 à
4 points de portage. Une figure serait de la décoration ici, les quatre nombres
se lisent mieux alignés.

## Pas retenu

Une frise de la décennie de l'exemple (niveau du matelas en mois + drawdown,
année par année) serait jolie, mais le bloc `exemple` la raconte déjà très
bien, et elle illustrerait un cas fabriqué plutôt qu'un fait. Deux figures plus
un tableau suffisent pour un article de ce format.


## strategie-buckets

Article aujourd'hui sans aucune figure. Trois thèses purement verbales méritent
une image : « les buckets sont une allocation » (pièce 1), « le waterfall est du
rééquilibrage » (pièce 2), « la discrétion dérive et vide les buckets »
(pièce 3).

## 2. Le réservoir des buckets sur la cohorte 1966 (coût B)

Une seule courbe, les mois de dépenses encore couverts par les buckets 1 et 2
(96 mois au départ, soit 8 ans), année par année à partir de 1966, en variante
discrétionnaire (« on remplira quand le marché ira mieux »). La courbe descend
en escalier et touche zéro avant la fin de la traversée, et le point où elle
touche zéro est la date à laquelle il faut vendre des actions au pire moment,
c'est-à-dire l'échec exact que la promesse fondatrice prétendait éliminer.
L'article affirme « le cas 1966-1981 les vide » sans le montrer. Données dans le
repo (`pkg/replay`, 60/40 réel US depuis 1954 déflaté du CPI), calcul de
consommation/recharge à écrire.

## 3. La dérive d'allocation, buckets discrétionnaires contre bandes (coût B)

Deux courbes de part actions, 1990-2010, même patrimoine et mêmes dépenses. Le
plan à bandes ±5 reste plat autour de sa cible, le plan à remplissage
discrétionnaire monte vers 85-90 % pendant le marché haussier des années 1990,
et arrive au sommet de 2007 avec la pire exposition de sa vie. Une annotation
sur 2008 suffit à conclure. C'est la seule affirmation chiffrée forte de la
pièce 3 (« plus exposé que son plan ne le disait ») et elle est aujourd'hui
présentée comme une évidence à croire. Mêmes séries que l'idée 2, plus une règle
de remplissage discrétionnaire à modéliser (par exemple, remplir seulement après
une année positive).

## 4. Les mêmes ordres, deux comptabilités (coût C, tableau)

Pas une figure, un petit tableau à trois années contrastées (actions +18 %,
actions −22 %, actions +6 %) et deux colonnes de comptabilité. Colonne
« waterfall avec suspension en baisse » et colonne « prélèvement-rééquilibrage
aux bandes », chacune donnant le montant vendu en actions et en obligations. Les
deux colonnes affichent les mêmes montants, ligne par ligne. La pièce 2 dit
« écrivez la comptabilité » puis ne l'écrit pas ; ce tableau est cette
comptabilité, et il rend la démonstration irréfutable en dix lignes.


## echelle-obligataire

Article aujourd'hui sans aucune figure. Trois idées, plus un tableau.

## 1. Le passif année par année, et qui le paie (barres empilées par année)

Une barre par année de 2026 à 2039 pour Claire et Idris. Hauteur de barre = le
besoin de l'année (plancher 45 000 € plus confort jusqu'à 58 000 €). Chaque
barre est découpée en segments selon la source qui paie : fonds euros (années
1-2), fonds à échéance nominal gonflé (3-8), linkers courts roulés (9-13),
portefeuille de confort (le complément, présent chaque année), pensions
(24 000 € à partir de l'année 14). On lit d'un coup la règle de couverture
(100 % du plancher les six premières années, 60 % ensuite), le relais des
pensions, et le fait que le portefeuille ne porte plus que la tranche haute.

Pourquoi elle vaut sa place : c'est la thèse centrale de l'article,
l'appariement actif-passif, et elle est aujourd'hui uniquement racontée en
prose dans le bloc exemple, poste par poste, ce qui est exactement le genre
d'énumération qu'une figure remplace mieux. C'est aussi la seule image qui
montre que l'échelle se consomme au lieu de composer.

Coût : **A** (tous les chiffres sont dans l'article et dans
`choisir-sa-strategie`).

## 2. Ce que devient un barreau nominal de 40 000 €, par année d'échéance

Courbe unique du pouvoir d'achat d'un barreau nominal de 40 000 € en fonction
de son échéance (1 à 25 ans), déflaté à 2,5 %, avec une bande grise pour
l'incertitude d'inflation (1,5 % en haut, 4 % en bas) et la ligne plate du
barreau indexé à 40 000 € réels. L'annotation de l'article, 40 000 € en 2038
valent ~30 000 € d'aujourd'hui, devient un point sur la courbe, et le seuil de
5-7 ans au-delà duquel la garantie nominale décroche se lit sur l'axe.

Pourquoi elle vaut sa place : elle porte la doctrine de la section « Nominal ou
indexé », c'est-à-dire « court, la nominale est acceptable ; long, indexé ou
rien », et surtout elle montre ce que le texte ne peut pas dire, que le vrai
problème n'est pas l'inflation anticipée mais sa dispersion, la bande.

Coût : **A/B** (deux lignes de calcul, aucune série externe).

## recharger-ou-pas

Article aujourd'hui sans aucune figure, alors qu'il est le seul du livre à décrire
une mécanique temporelle (armement, hystérésis, recharge plafonnée, extinction).
Trois idées, par ordre de valeur.

## 1. La décennie d'un matelas, en deux étages (coût B)

Deux panneaux alignés sur le même axe de temps, sur une décennie historique réelle
(1969-1979 ou 2000-2010, séries réelles US 60/40 de `pkg/replay`, déjà déflatées).
En haut, le drawdown réel du portefeuille avec la bande du seuil d'armement
(15 %) et la ligne de désarmement (10 %). En bas, le niveau du buffer en mois,
en escalier, qui descend quand le déclencheur est armé, plat pendant la baisse
(aucune recharge), puis remonte par petites marches au retour du calme, avant de
s'éteindre après l'année 10.

Pourquoi elle vaut la place : c'est la thèse centrale de l'article (« toute la
valeur est dans les flux ») et elle est aujourd'hui entièrement verbale. Une seule
figure montre en même temps le déclencheur, l'hystérésis, l'interdiction de
recharger en baisse et le buffer fondant, c'est-à-dire les quatre règles que le
texte énumère l'une après l'autre. C'est aussi le pendant visuel de l'exemple de
Salomé.

Coût : la série réelle existe, la plomberie du matelas aussi
(`pkg/decumul`, `BufferSleeve`) ; il faut écrire le rejeu déterministe sur
l'historique et en sortir les deux séries à tracer.

## 2. Le balayage du seuil d'armement (coût B)

Pour X = 5, 10, 15, 20, 30 %, trois quantités mesurées sur le même jeu de
trajectoires : le nombre d'armements par décennie, la part des mois de retrait
effectivement payés par le buffer, et la part des trajectoires où le buffer est
vide au pire moment (au plus bas du drawdown le plus profond). Barres groupées
ou barres d'intervalle classées, labels horizontaux.

Pourquoi elle vaut la place : le paragraphe « le réglage de X demande du doigté »
affirme les deux pathologies (trop bas, on gaspille ; trop haut, on regarde
passer les baisses moyennes) sans jamais les chiffrer, et c'est le seul réglage
du matelas sur lequel le lecteur a une vraie décision à prendre. La figure
trancherait aussi la petite divergence de doctrine du livre (10 % ici, 15-20 %
dans `cash-buffer` et le lexique).

Coût : un balayage à écrire, mais sur des briques existantes.

## 3. « Faites les comptes » : la double peine de la recharge en baisse (coût C)

Pas une figure, un encadré chiffré de six lignes. Même ménage, même krach de
−30 %, même retour au sommet. Ménage A recharge 12 mois de buffer au plus bas,
ménage B attend le retour du calme. On affiche les parts vendues dans chaque cas,
la valeur de ces parts au retour au sommet, et l'écart final en mois de dépenses.

Pourquoi elle vaut la place : l'article dit « Faites les comptes » et ne les fait
pas. Un ordre de grandeur explicite (l'écart se compte en mois de dépenses perdus,
pas en points de base) est ce qui transforme l'interdiction en réflexe. Un tableau
de trois lignes bat ici n'importe quel graphique, et les nombres se posent à la
main, sans nouvelle simulation.

## Non retenu

Une courbe de ruine en fonction de la taille du matelas : déjà portée par la
figure `buffer-flat` de `cash-buffer`, et ce n'est pas le sujet de cet article.


## immobilier-en-retrait

Article aujourd'hui sans aucune figure, alors que c'est l'un des plus chiffrés
du livre. Deux passages sont porteurs et purement verbaux (la cascade du
rendement, le classement flux/capital/réserve). Ordre de priorité ci-dessous,
et deux figures suffiraient (idées 1 et 2).

## 2. Où va chaque brique du patrimoine dans le plan (priorité haute, coût C)

Pas un graphique, un schéma à trois colonnes : capital du plan (× 25), flux
(revenu complémentaire décoté), réserve (listée, non comptée). On y range
résidence principale, locatif, SCPI, nue-propriété, portefeuille, et on montre
la flèche interdite (la résidence qui saute dans la colonne capital alors que
son service est déjà dans le budget). C'est le squelette de l'encart « Les
deux règles de comptage » et de l'article entier, et c'est exactement le genre
de chose qu'un lecteur photographie pour la refaire sur son propre patrimoine.
Un tableau à trois colonnes ferait déjà 80 % du travail, d'où le coût C.

## 3. L'escalier fiscal de la plus-value (priorité moyenne, coût A)

Le taux effectif total (IR 19 % + PS 17,2 %, abattements de durée appliqués)
en fonction des années de détention, de 0 à 30 ans, en aires empilées IR et
PS. On voit la première marche à 6 ans, la chute de l'IR jusqu'à zéro à 22
ans, puis la pente raide des PS de 23 à 30 ans. Cela porte la phrase « le
calendrier de la vente devient une variable fiscale majeure », aujourd'hui
affirmée sans preuve, et cela justifie visuellement le choix de Yann de vendre
en année 22. Grilles disponibles (BOI-RFPI-PVI), aucun calcul de série à
écrire. Réserve : c'est du droit vivant, donc une figure à dater explicitement
dans sa légende, sous peine de vieillir mal.

## 4. Loyers indexés contre inflation, 2006-2025 (priorité basse, coût B)

Deux courbes rebasées à 100, l'IRL et l'indice des prix, avec la zone
2022-2024 grisée où l'IRL a été plafonné à +3,5 % par la loi. La figure dit ce
que le texte affirme désormais avec nuance : l'indexation locative est réelle
mais retardée et politiquement révisable, donc un linker imparfait. C'est
l'argument le plus distinctif de l'article, et le seul dont la contrepartie
(le plafonnement) reste invisible en prose. Coût B, et les séries ne sont pas
dans le repo (IRL trimestriel INSEE à récupérer), ce qui la relègue
derrière les trois autres.


## levier-et-marges

Article aujourd'hui sans aucune figure, alors qu'il porte deux mécaniques
franchement visuelles (le coût comparé pont/matelas, et l'appel de marge).
Attention à ne pas empiéter sur la figure `stacking-expo` de
[[return-stacking]], qui montre déjà les 100 € et l'exposition à 150 %. Rien
ici ne doit refaire ce dessin.

## 1. L'escalier du pont contre la droite du matelas (coût A)

Deux courbes de coût cumulé sur une longue histoire réelle : le matelas de
cash coûte 0,3 à 0,4 % par an, tous les ans, donc une droite qui monte sans
répit ; le pont lombard ne coûte un demi-point que pendant les épisodes de
creux, donc un escalier plat la plupart du temps. Les marches se placent aux
vrais drawdowns > 20 % du 60/40 réel américain (`pkg/replay`, série depuis
1954), ce qui rend la comparaison historique et non hypothétique.

C'est la thèse chiffrée de l'usage 1, aujourd'hui entièrement verbale, et le
dessin dit d'un coup ce que le texte doit énoncer en quatre phrases : le pont
gagne sur le coût, et il perd ailleurs (le confort). Bonus honnête : annoter
la dernière marche « 2008 : la dette monte pendant que le portefeuille
baisse » pour ne pas laisser croire que la figure tranche le débat.

Coût **A** : les deux taux sont dans l'article, la série de drawdowns est déjà
dans le repo.

## 2. La quotité qui fond, ou l'anatomie d'un appel de marge (coût B)

Un seul épisode (2007-2009 sur le 60/40 réel) et trois traits : la valeur du
portefeuille qui baisse, le LTV d'une dette constante qui monte
mécaniquement, et la quotité que le prêteur accepte, en escalier
descendant parce qu'elle se resserre en crise. L'aire entre les deux derniers
traits est la marge de sécurité restante ; le point où ils se croisent est
l'appel de marge.

C'est la règle n° 1 (« votre plafond doit rester loin du leur ») et le risque
central du chapitre, la liquidation forcée, qui n'existent aujourd'hui que
sous forme d'affirmation. Une figure les rend inévitables. Elle sert aussi
d'argument visuel pour le plafond LTV de 20-25 % : on voit qu'un tirage de
50 % touche le mur, et qu'un tirage de 20 % passe l'épisode sans être appelé.

Coût **B** : le chemin du portefeuille existe, mais le durcissement des
quotités en crise est une hypothèse à poser et à documenter en légende
(60 % → 40 % par exemple), pas une donnée du repo.

## 4. Le plateau des seuils d'ERN (coût B)

Le volet 52 teste plusieurs seuils de déclenchement et le résultat a une forme
intéressante : le taux de retrait soutenable monte de 3,58 % sans levier à
3,84 % à 20 % de drawdown, plafonne vers 3,92-3,93 % à 25 et 30 %, puis se
dégrade nettement à 35 %. Un petit graphique en barres avec le plateau et la
chute finale dit à la fois « le gain est réel » et « il est modeste, et le
réglage n'est pas libre ».

Utile parce que la section promet des chiffres et que l'usage 3 n'en donne
qu'un seul. À réserver si l'on accepte une figure entièrement sourcée chez ERN
(cohorte de novembre 1965), valeurs à revérifier dans le volet 52 avant
publication.

Coût **B** : rien de tout cela n'est dans le repo, les valeurs viennent d'ERN.


---

# L'inflation


## inflation-histoire

L'article a déjà `franc-decay` (l'érosion du franc 1914-2025), qui couvre bien
la section « grandes destructions ». Trois passages restent purement verbaux et
portent chacun une thèse chiffrable.

## 2. Le même capital dans les trois régimes, en trajectoires

Coût **A** : toute l'arithmétique est dans le bloc `exemple` de l'article.

Ce qu'elle montre : les trois chemins sur dix ans du même million placé en
obligations d'État et fonds euros, en valeur réelle. Stabilité en légère pente
montante (+5 %), répression en dérive régulière (−14 %), épisode en falaise de
deux ans puis plateau (~−15 %). Et surtout, en surimpression, la ligne du
capital **nominal**, qui ne baisse jamais dans les trois cas.

Pourquoi elle vaut sa place : la thèse du bloc est « aucun relevé n'affiche de
perte ». C'est une thèse visuelle, pas numérique. Le contraste entre une ligne
nominale plate et trois lignes réelles qui divergent la démontre en une
seconde ; le tableau de chiffres actuel demande au lecteur de la reconstituer.
La différence de **forme** entre la dérive de la répression et la falaise de
l'épisode est le second enseignement, et elle n'existe que sur un graphique.

## 3. Le taux réel servi à l'épargne française depuis 1945

Coût **B**. Le rendement réel des bills français est dans JST jusqu'en 2020 ; il
faut y raccorder 2021-2025 (monétaire euro contre IPCH, ou `refdata/EURCASH-EUR`
déflaté) et fixer la convention de raccord.

Ce qu'elle montre : année par année, l'écart entre ce que rapporte l'épargne
sans risque et l'inflation, autour d'une ligne zéro. Trente-cinq ans presque
continûment sous zéro de 1945 à 1980, un plateau positif de 1985 à 2008, puis
un retour sous zéro après 2010 et un trou en 2021-2023.

Pourquoi elle vaut sa place : le bloc `science` sur la répression financière est
le passage le plus prospectif de l'article, et le seul qui n'apporte aucune
preuve. Une seule figure établit que la répression n'est pas une théorie mais
l'état par défaut de l'épargne française sur la moitié du siècle, ce qui rend le
scénario « les trente prochaines années » beaucoup plus difficile à balayer.

Forme : aires divergentes autour de zéro (négatif rempli, positif rempli plus
clair), régimes de l'article annotés en bandeau au-dessus.

Recoupement à arbitrer : la ligne « monétaire » de l'idée 1 dit une partie de
l'idée 3. Si une seule des deux passe, garder l'idée 1, qui porte la thèse
principale de l'article ; l'idée 3 est la meilleure candidate si le bloc
répression doit être renforcé pour lui-même.


## suivre-inflation

Article aujourd'hui sans aucune figure. Trois passages portent une thèse
quantitative en prose pure et gagneraient à être vus.

## 3. L'indice contre votre panier, poste par poste (coût B)

Graphique à pentes (dumbbell) : à gauche les pondérations de l'IPCH France par
fonction, à droite celles d'un retraité propriétaire toit payé, une ligne par
poste. Les deux lignes qui comptent visuellement sont le logement (loyers
effectifs ~7 % à gauche, presque zéro à droite pour un toit payé) et la santé
avec les services (faibles à gauche, gonflés à droite).

Pourquoi ça vaut la place : la section « Ce que le panier contient mal » et
l'encart « Les trois inflations » disent tous deux « le panier n'est pas le
vôtre » sans jamais le montrer. C'est la figure qui rend l'angle mort du
logement du propriétaire palpable, et elle sert deux fois dans l'article.

Données : pondérations COICOP de l'IPCH France (Eurostat, à récupérer, pas dans
le repo) et un panier retraité de référence à poser explicitement comme
hypothèse illustrative. Donc un calcul et une source à écrire d'abord.

## Écarté

- Une longue série « prix des logements contre IPCH » pour illustrer « quand
  l'immobilier double, l'IPCH ne bouge pas » : très parlante, mais l'indice
  Notaires-INSEE des prix des logements n'est pas dans le repo, et la figure 3
  porte déjà l'angle mort du logement à moindre coût.
- Les points morts d'inflation : le bloc `astuce` suffit, une figure de série
  de breakevens inviterait à la lecture au mois près, exactement ce que
  l'article déconseille (coût C, l'encart fait mieux).


## inflation-et-taux-de-retrait

Article pivot de la partie « L'inflation », aujourd'hui sans aucune figure, alors
qu'il porte deux mécanismes purement verbaux (les ciseaux, la compression réelle
simultanée) et une thèse forte (« 1966 devant 1929 »). Quatre idées, par ordre de
priorité. Si une seule est construite, prendre la n° 1 ; si deux, les n° 1 et 2.

## 1. « 1966 devant 1929 » : deux voyants, un qui redescend, un qui ne redescend jamais

Le taux de retrait courant (retrait / portefeuille, en réel) du millésime 1929 et
du millésime 1966, superposés sur 30 ans depuis le départ, avec la bande
d'irrécupérabilité 8-10 % marquée. 1929 monte violemment puis **redescend** (la
déflation baisse les retraits en nominal, les actifs nominaux reprennent de la
valeur réelle) ; 1966 monte lentement, franchit 8 % vers 1975 et ne redescend
plus. C'est la démonstration visuelle de la phrase centrale de l'article, « le
krach rend, l'épisode ne rend jamais », que le texte affirme sans la montrer.
Différenciation : `bengen-falaise` (dans [[retrait-fixe-bengen]]) montre déjà le
voyant de 1966 seul ; ici tout l'intérêt est le **contraste** avec 1929, qui
n'existe nulle part dans le livre.

Coût **B** : les séries sont dans le repo (SP500-USD, reconstruction Treasury,
CPI-US, tous longs), mais `pkg/replay` démarre en 1954, donc le millésime 1929
demande un petit calcul de voyant à écrire sur les séries mensuelles.

## 3. 1973-1981 en réel : la diversification qui ne diversifie plus

Barres d'intervalle classées, ou petites-multiples de cumulés réels, pour
S&P 500, obligations d'État longues et intermédiaires, cash, or et 60/40 sur
1973-1981, en réel. Tout le monde est du mauvais côté de zéro sauf l'or. La
phrase « le 60/40 n'a aucune poche qui gagne » est aujourd'hui une assertion ;
là, elle devient un fait lisible en trois secondes, et elle prépare l'inventaire
d'indexation qui suit. Différenciation : `regime-grid` ([[regimes-de-marche]])
est une grille qualitative de saisons, celle-ci est le chiffre d'un épisode
nommé.

Coût **A/B** : toutes les séries existent (`pkg/datasets`), le calcul est un
cumul déflaté sur une fenêtre fixe.

## 4. L'exposition nette à l'inflation, phase par phase

Barres divergentes en euros pour le plan de référence de l'article : à gauche ce
qui suit les prix (pension indexée, linkers, loyers, crédit fixe en négatif), à
droite ce qui ne les suit pas (fonds euros, nominal, rentes privées), une barre
pour la phase à découvert et une pour la phase adossée. La phase adossée penche
franchement du bon côté grâce à la pension, la phase à découvert est un mur du
mauvais côté. C'est la conclusion d'allocation de l'article rendue arithmétique,
au lieu d'être déduite d'un tableau de qualificatifs.

Coût **A** : tous les chiffres sont dans l'article et dans son bloc d'exemple, le
tableau d'indexation fournit les signes. Réserve honnête : ce n'est pas loin
d'être une mise en forme du tableau existant, donc à ne construire que si les
n° 1 à 3 sont faites et que la place reste.


## se-proteger-de-inflation

Article d'assemblage, aujourd'hui sans aucune figure. Sa thèse est un
classement (nature de preuve, dose, coût en croisière) et une liste noire.
Trois idées, dans l'ordre de valeur.

## 1. La carte des protections : gain pendant l'épisode contre coût en croisière

Un plan en deux axes où chaque brique est un point posé nommément. En abscisse,
le rendement réel moyen pendant les épisodes d'inflation documentés ; en
ordonnée, le rendement réel moyen hors épisode (le coût de détention en
croisière). Les quatre quadrants racontent exactement l'encart « la règle
d'achat » : en haut à droite les protections payées pour être détenues
(actions, value, pension rachetée, linkers au point mort à 2 %), en bas à
droite les épisodiques qui coûtent en croisière (or, matières premières,
trend), en bas à gauche la liste noire entière (fonds euros, monétaire,
dividendes, taux variables, crypto), qui ne gagne ni pendant ni après.

C'est la figure qui remplace le mieux du texte : elle porte à elle seule les
étages 1 à 3, la liste noire et la doctrine de dose, aujourd'hui étalés sur
quatre sections purement verbales. Une taille de marqueur proportionnelle à la
dose écrite au plan ferait passer un troisième message sans un mot de plus.

Coût : **B**. L'or, les actions, les Treasuries, le CPI et le trend sont dans
le repo (`pkg/datasets/simdata`, `refdata`, backcast DBMF) ; il faut écrire la
définition des fenêtres d'épisode (celles de Neville et al., déjà citées en
fin d'article) et accepter des points documentés à la main pour les briques
non série (pension rachetée, crypto sur son unique épisode).

## 3. La cascade du programme de Sonia

Une cascade de quatre barres : ruine broad-sample de départ 12,4 %, puis l'apport
de l'étage 4 (gratuit), celui de l'étage 1 (linkers, échelle, rachat de
trimestres), celui de l'étage 3 (or et trend en dose), jusqu'aux 8,1 % finaux.
Une seconde étiquette par barre donne le coût en point d'espérance, pour montrer
que la plus grosse marche est aussi la moins chère.

L'exemple chiffré donne aujourd'hui le point de départ et le point d'arrivée,
mais pas la contribution de chaque étage, qui est précisément la thèse de
l'article (l'ordre d'achat). La cascade est la forme qui montre un ordre.

Coût : **B**. Il faut faire tourner les quatre variantes du plan de Sonia pour
obtenir les marches intermédiaires ; les bornes 12,4 et 8,1 % sont dans le
texte, les marches ne le sont pas.


## hyperinflation-et-extremes

Article aujourd'hui sans aucune figure. Trois passages portent la thèse et sont
purement verbaux : le bilan « ce qui survit, ce qui meurt », la démonstration
que les créances domestiques meurent alors que les actifs réels sont seulement
amochés, et la calibration de la peur (le piège du prepper).

## 1. La grille des extrêmes : qui meurt, qui survit, dans quel scénario

Une grille à quatre colonnes (hyperinflation, contrôle des capitaux et
ponction, guerre sur le sol, répression réglementaire douce) et six lignes
(cash et livrets, obligations et fonds euros, rentes et assurance-vie, actions
domestiques, actifs étrangers en monnaie étrangère, or physique et pierre).
Chaque case porte un verdict à trois niveaux plus une ancre historique en trois
mots (« Weimar 1923 », « Chypre 2013 », « loi de 1948 »). Le paragraphe actuel
énumère ce bilan en huit phrases denses ; la grille le rend consultable et
montre d'un coup la colonne invariante, celle des actifs hors juridiction et
hors monnaie, ce qui est exactement l'argument du chapitre.

Coût : **C**. Tout le contenu est déjà dans l'article ; un tableau bien réglé
fait mieux qu'un graphique ici, puisque les cases sont des verdicts, pas des
nombres. Attention à ne pas doubler la « grille de régimes » du livre : ce
n'est pas un tableau de rendements, c'est une matrice de survie.

## 3. Fréquence contre fascination : où va le budget d'assurance

Un graphique à deux volets appariés, sur la même liste de causes d'échec d'un
plan français (séquence ordinaire mal gérée, inflation persistante à 4 %,
dépenses sous-estimées, contrôle des capitaux ou ponction, redénomination,
hyperinflation domestique). À gauche, l'ordre de grandeur de la fréquence, en
échelle logarithmique. À droite, la part du plan que le chapitre leur accorde,
cinq lignes et 2 à 8 % du patrimoine. La lecture est immédiate : les deux
classements sont inversés chez le prepper, alignés dans le plan du livre. C'est
l'argument « on meurt de ce qui arrive, pas de ce qui fascine », aujourd'hui
tenu par la seule phrase « dix à cent fois plus ».

Coût : **B**. Il faut d'abord assumer une table de fréquences : les taux de
ruine du livre pour les causes ordinaires, un comptage d'épisodes par
pays-décennie de type Reinhart et Rogoff pour les contrôles et ponctions, et un
majorant pour l'hyperinflation en zone euro. Figure honnête seulement si elle
affiche des ordres de grandeur (des barres par facteur 10, sans décimale) et
non des probabilités précises, sinon elle fabrique la fausse précision que le
chapitre reproche à l'industrie de la peur.


---

# Fiscalite et cadre francais


## enveloppes-francaises

L'article n'a aujourd'hui aucune figure, et c'est celui du livre où le lecteur
doit comparer quatre taux effectifs de tête. Quatre idées, classées par valeur.

## 2. Les mêmes 55 000 €, deux organisations : un tableau, pas une figure (coût C)

Le bloc `::: exemple` porte maintenant huit nombres cohérents en prose continue
(20 000 / 28 000 / 10 000 vendus, 0 / 2 300 / 600 d'impôt, 58 000 contre
64 000). C'est de la matière de tableau, pas de graphique : trois lignes
(assurance-vie, PEA, CTO), deux blocs de colonnes (organisation A, organisation
B), colonnes « vendu » et « impôt », plus une ligne de total. Le lecteur veut
vérifier l'addition, ce qu'aucune barre empilée ne permet.

À défaut, deux barres horizontales empilées (une par organisation, segments par
enveloppe, la fraction d'impôt en surimpression) diraient le volume vendu d'un
coup d'oeil, au prix de la vérifiabilité.

## 4. L'abattement annuel, un actif périssable qui se chiffre (coût A)

Deux escaliers cumulés sur 30 ans de retraite. Le premier, celui du couple qui
consomme chaque année son abattement de 9 200 € de gains : 276 000 € de gains
sortis sans impôt, soit environ 68 000 € d'impôt évité à 24,7 %. Le second,
celui qui ne rachète qu'une année sur trois : la moitié de la courbe disparaît,
définitivement, car l'abattement ne se reporte pas.

Le bloc `::: science` affirme que l'abattement est « perdu s'il n'est pas
utilisé » sans jamais dire combien. Ces 68 000 € sont le chiffre qui manque, et
la figure le construit année par année, ce qui montre en même temps que
l'avantage est un flux à entretenir, pas un stock qu'on récupère plus tard.


## flat-tax-et-imposition

Article sans aucune figure aujourd'hui, alors que sa thèse centrale (« la friction
n'est pas un taux, c'est un taux multiplié par une fraction de gain qui dérive »)
est purement verbale et contre-intuitive. Deux figures et deux tableaux la
porteraient mieux que n'importe quel paragraphe de plus.

## 2. PFU contre barème, la lecture par tranche (coût B)

Barres appariées par TMI (0, 11, 30, 41, 45 %) et ligne horizontale de
référence à 31,4 %, en deux panneaux, plus-values récentes d'un côté,
dividendes de l'autre. Chaque barre est le taux total effectif au barème
(TMI + 18,6 % de PS, moins l'effet de la CSG déductible, avec l'abattement de
40 % côté dividendes). On voit la bascule là où elle est vraiment : le barème
domine jusqu'à 11 % dans les deux panneaux, et le point d'indifférence des
dividendes tombe vers TMI 24 %, donc le PFU l'emporte déjà à 30 %.

Cette figure remplace le paragraphe le plus difficile de l'article (celui qui
énumère les cas par tranche) et rend visible le seul chiffre qu'on ne peut pas
deviner, le point de croisement. Elle protège aussi contre l'erreur courante
du « barème toujours bon pour les dividendes ».

Calcul court à écrire, entièrement analytique, aucune série requise.

## 3. L'année fiscale du couple, en tableau (coût C)

Le bloc `::: exemple` aligne aujourd'hui huit montants en prose. Un tableau de
quatre colonnes (robinet, flux extrait, part de gain, impôt) plus une ligne de
total, avec le lissage sur sa propre ligne clairement marquée « volontaire »,
serait lisible en trois secondes et rendrait le recoupement possible pour le
lecteur. Pas une figure, un tableau, et il gagnerait la place qu'il prend.

## taxe-puma

Article aujourd'hui sans aucune figure. Il porte une formule à deux entrées, donc
une carte vaut mieux qu'un paragraphe. Trois idées, par ordre de valeur.

## 2. L'étalement, en barres décroissantes (coût A)

Un même total de 120 000 € de revenus du capital à réaliser, découpé en 1, 2, 3, 4
puis 5 années égales. La CSM totale payée s'effondre à mesure que le découpage
s'affine, parce que la franchise annuelle se rejoue chaque année : environ 6 240 €
en une seule fois, 4 680 € en deux ans, 3 110 € en trois, 1 550 € en quatre, zéro
en cinq. La figure porte à elle seule le conseil « étaler les grosses plus-values »
de la parade n° 2, et le zéro final est plus convaincant que n'importe quelle phrase.

Coût A, même formule que ci-dessus. À garder très petit, cinq barres et un chiffre
au bout de chacune. Complémentaire de l'idée 1 et non redondante : l'une montre la
règle, l'autre montre un geste.

## 3. Le zigzag 2016 / 2019, en tableau (coût C)

Pas une figure, un tableau de quatre lignes et trois colonnes (paramètre, régime
2016, régime depuis 2019) : taux 8 % contre 6,5 %, franchise 25 % contre 50 % du
PASS, seuil d'activité 10 % contre 20 % du PASS, décote absente contre linéaire.
Le paragraphe du zigzag résume déjà ces chiffres, mais le tableau montre que les
quatre paramètres ont bougé ensemble, ce qui justifie l'avertissement de veille
annuelle bien mieux que l'affirmation seule.

Coût C : tout est dans l'article, aucune donnée à charger.

## Rien d'autre

Les trois versions du pont (A, B, C) sont déjà lisibles en prose, trois constantes
multipliées par douze ans. Une figure ne leur ajouterait rien.


## retraite-legale

Article sans figure aujourd'hui. Trois idées valent la place, une quatrième est
défendable en tableau. Classées par valeur décroissante.

## 1. Le point mort 64 contre 67, deux cumuls qui se croisent (coût B)

Deux courbes de pension cumulée perçue en fonction de l'âge atteint, l'une pour
une liquidation à 64 ans (86 % de la pension pleine, trois années d'avance, plus
la CSM évitée en marche d'escalier au départ), l'autre pour 67 ans (100 %, trois
années de retard). Elles se croisent vers 85-87 ans, et l'espérance de vie du
lecteur se pose en repère vertical sur le même axe.

C'est la thèse la plus contre-intuitive de l'article, et elle est aujourd'hui
purement verbale. La figure montre d'un coup que le plafond de décote à 64 ans
(12 trimestres, −15 % sur la base et −12 % sur la complémentaire) rend
l'arbitrage serré au lieu d'évident, et que le verdict se lit en longévité, pas
en pourcentage de décote. Coût B, le calcul est de l'arithmétique simple sur les
coefficients officiels déjà cités dans l'article, mais il reste à écrire.

## 2. La cascade de la carrière écourtée (coût B)

Cascade partant de la pension d'une carrière complète pour un cadre, puis les
marches successives, effet SAM des 25 meilleures années sur une carrière de 22
ans, proratisation 88/172, taux plein automatique conservé à 67 ans (marche
nulle, et c'est le message), complémentaire acquise au prorata des cotisations.
Arrivée sur la fourchette 1 200-1 900 €/mois de l'article.

L'article affirme une « double peine » puis explique qu'à 67 ans une seule des
deux joue. Une cascade rend cette asymétrie visible en une seconde et ancre
l'ordre de grandeur, qui est le chiffre que le lecteur emportera. Le livre
utilise déjà des cascades, et c'est ici la bonne forme. Coût B, il faut poser
des paramètres de cadre explicites et assumés dans la légende.

## 3. Les quatre rendements d'un même petit revenu (coût C, tableau)

Tableau à une colonne d'entrée (7 200 € puis 9 500 € de revenus cotisés dans
l'année) et quatre colonnes de sortie chiffrées, trimestres validés,
proratisation gagnée en euros de pension annuelle à 67 ans, CSM éteinte en
euros, retraits évités l'année du choc. Une ligne de total en équivalent capital.

Le passage « quatre bénéfices d'un seul geste » est l'argument le plus
actionnable du chapitre et il n'est jamais chiffré ligne par ligne. Un tableau
fait mieux qu'un graphique ici, car les quatre effets n'ont pas la même unité.
Il rend aussi visible le décalage entre les deux seuils, 7 200 € pour les
trimestres et ~9 500 € pour la PUMa, qui est exactement le piège corrigé dans
cette passe.

## 4. La grille des générations, après la suspension (coût A, tableau, réserve)

Petit tableau âge légal et durée requise par génération, de 1963 à 1969 et
au-delà, état après le gel voté pour 2026-2028. Données vérifiées, coût A.

Réserve sérieuse, c'est le contenu qui périmera le plus vite de tout le livre,
et l'article a délibérément choisi de renvoyer à info-retraite.fr plutôt que de
recopier des barèmes. À ne retenir que si le livre accepte un encadré
explicitement daté, avec sa date d'arrêté visible dans le titre.


## sante-et-protection-sociale

Article aujourd'hui sans aucune figure. Trois passages portent des chiffres que
la prose énumère sans les faire voir. Ordre de valeur décroissante.

## 1. La cascade du reste à charge en EHPAD, deux ménages côte à côte (coût B)

Deux cascades verticales alignées sur la même échelle en €/mois. À gauche un
ménage modeste, à droite un ménage aisé (le lecteur du livre). Chacune part du
prix affiché (2 200 € en chambre habilitée à l'aide sociale, 3 100 € sinon),
retire l'APA, l'aide au logement puis la réduction d'impôt, et atterrit sur le
reste à charge. La thèse du passage est exactement là et reste invisible en
prose. Les aides publiques ne retirent que 400 à 450 €/mois en moyenne, et le
ménage aisé perd les deux plus grosses barres, donc son reste à charge est
proche du prix affiché. C'est ce qui justifie la provision et non une ligne de
budget. Coût B, un petit calcul d'aides à écrire (barème APA en établissement,
plafond de la réduction d'impôt), rien à chercher côté séries.

## 2. La règle du séjour, en années et en euros de provision (coût B)

Une seule règle horizontale, l'axe du bas en années de séjour, l'axe du haut en
k€ de reste à charge cumulé (à 30 k€/an). On y pose les repères de la DREES,
le premier quartile (moins de deux mois), la médiane, la moyenne à deux ans et
trois mois, le troisième quartile (trois ans et deux mois), et la queue au-delà
de huit ans, où la lecture du haut donne les 250 k€ que l'article cite deux
fois. Une seule image répond à la question « combien provisionner », et elle
montre pourquoi la moyenne est le mauvais chiffre pour dimensionner. Coût B,
les quartiles ne sont pas dans le repo, à figer en constantes avec la source.

## 3. Le budget santé d'un rentier de 47 à 95 ans (coût B)

Un escalier de dépenses annuelles en euros constants, empilé en trois bandes,
la mutuelle (qui court à +4-6 %/an, donc quadruple sur la période), le reste à
charge courant, et la temporaire décès qui s'arrête net à la fin du pont. À
droite de la zone empilée, hors de l'empilement, un bloc grisé isolé pour la
provision dépendance. La séparation graphique porte la doctrine du chapitre et
celle de [[depenses-en-retraite]], la dérive va au budget, le choc va dans une
provision, jamais les deux mélangés. C'est aussi la meilleure façon de montrer
que la mutuelle, et non le prix des soins remboursés, est la ligne qui dérive.
Coût B, arithmétique simple à partir des fourchettes de l'article.

## 4. Ce qui disparaît le jour du départ (coût C, un tableau)

Pas une figure. Un tableau de quatre lignes, une par garantie perdue
(complémentaire santé, capital décès, rente d'invalidité, prévoyance
conjoint-enfants), et trois colonnes, ce que le salarié avait, ce que la
portabilité prolonge (douze mois au plus), ce qui remplace ensuite (contrat
individuel, temporaire décès, marges du plan). L'encadré « trou invisible » dit
tout cela en un paragraphe dense, et un tableau se relit à la veille du départ,
ce que le paragraphe ne permet pas. Aucun calcul.


## succession-et-transmission

L'article n'a aucune figure aujourd'hui, alors qu'il repose sur trois chiffres
purement assertés (la friction organisée contre improvisée, le plateau de
donation sans coût pour le plan, le nombre de recharges selon l'âge de départ).
Quatre idées, la 2 étant la plus précieuse parce qu'elle est la seule à relier
la transmission au reste du livre.

## 1. Le genou de la donation : ce que donner coûte au plan

Courbe de la probabilité de ruine (axe y) en fonction du montant donné
aujourd'hui (axe x, de 0 à ~700 k€) pour le ménage de l'exemple. On y voit un
long plateau plat, puis un genou net où la ruine décolle. Le plateau EST
l'« excédent manifeste » que l'article invoque sans le montrer, et le genou est
la frontière opérationnelle de la règle d'or. Une seule figure remplace tout le
paragraphe « simulez, et si la ruine bouge, c'est non », et elle rattache la
fiscalité successorale à l'objet central du livre, le plan.

Forme : courbe simple avec la zone plateau ombrée et le genou annoté, éventuellement
deux tracés (55 ans et 70 ans) pour montrer que le plateau s'élargit avec l'âge.

Coût : **B**. Le moteur de décumulation du repo produit déjà la ruine pour un
capital donné ; il suffit de balayer le capital initial diminué de la donation
(et, pour la variante démembrée, de retirer le capital tout en gardant le
revenu). Une centaine de lignes, pas de données nouvelles.

## 2. Friction effective : organisée contre improvisée

Deux courbes de taux effectif de droits (axe y, 0-35 %) contre patrimoine total
du couple (axe x, 0,5 à 4 M€), deux enfants. Ligne du haut, la succession non
préparée. Ligne du bas, la même après le socle de l'article (quatre étages d'AV
avant 70 ans, deux vagues de recharges à 15 ans, démembrement à 60 ans). L'écart
entre les deux courbes est la thèse entière de l'article, aujourd'hui résumée en
un « 5-10 % contre 10-30 % » que le lecteur doit croire sur parole. On voit en
plus ce que les fourchettes cachent, à savoir que la friction improvisée grimpe
avec la taille du patrimoine alors que la friction organisée reste presque plate
jusqu'à ~2 M€.

Utile aussi pour désamorcer le fantasme du « 45 % » de l'introduction, car le
taux marginal du barème et le taux effectif divergent énormément.

Coût : **B**. Barème de l'art. 777 CGI plus les abattements, une trentaine de
lignes ; aucune série de marché. Attention à afficher les deux routes du cas non
préparé (conjoint en usufruit contre tout au second décès), elles diffèrent d'un
facteur proche de 2 (~13 % contre ~21 % à 2 M€) et c'est précisément ce que
recouvre la fourchette 10-20 %.

## expatriation-fiscale

L'article est entièrement verbal aujourd'hui (aucune figure). Deux passages
portent la thèse et souffrent de rester en prose dense : « les régimes de faveur
sont des fenêtres, pas des droits », et le solde gains-coûts proche de zéro.

## 1. La frise des fenêtres qui se referment (coût A)

Une frise 2009-2026, une ligne par régime, barre pleine quand le régime est
ouvert aux nouveaux entrants, barre creuse quand il est fermé, avec un marqueur
daté sur chaque changement de règle. Portugal NHR (ouvert 2009, fermé aux
nouveaux entrants au 1er janvier 2024, remplacé par l'IFICI), Italie 7 %
pensionnés (2019, seuil de commune élargi en 2026), Italie forfait néo-résidents
(100 k€ en 2017, 200 k€ au 10 août 2024), Grèce 7 % pensionnés (2020) et forfait
100 k€ (2019), Belgique plus-values (exonérées jusqu'au 31 décembre 2025, 10 %
ensuite), exit tax française (15 ans avant 2019, 2 ou 5 ans depuis, tentative de
retour à 15 ans votée puis abandonnée en 2026).

C'est la meilleure candidate. Elle transforme la phrase-clé de l'article en
preuve visuelle, et elle montre que le mouvement va dans les deux sens (l'Italie
s'ouvre côté communes, se ferme côté forfait), ce que la prose ne dit pas. Toutes
les dates sont dans l'article ou vérifiables sans calcul, donc aucune donnée du
repo n'est nécessaire. Forme nouvelle pour le livre, qui n'a pas encore de frise
d'états ouvert/fermé.

## 2. Le solde net d'expatriation selon la taille du patrimoine (coût B)

Trois petites cascades côte à côte, à 1 M€, 2 M€ et 5 M€ de patrimoine, même
échelle en euros par an. Chaque cascade empile les gains (PFU/PS évité, CSM
éteinte, sortie de l'IFI immobilier) puis retranche les coûts (santé privée ou
CFE, double vie, conseil) pour atterrir sur le solde. On voit d'un coup que les
coûts sont quasi fixes tandis que les gains montent avec le patrimoine, donc que
le solde ne devient franchement positif qu'en haut de l'échelle.

C'est exactement la thèse du paragraphe « Le calcul complet », aujourd'hui livrée
sous forme de liste de fourchettes que le lecteur doit additionner de tête. Le
seuil de bascule est l'information que tout le monde cherche dans ce chapitre.
Coût B, car il faut écrire le petit modèle qui relie taille du patrimoine, part
de gain réalisée et fourchettes de coûts. À faire avec une seule hypothèse de
train de vie, affichée sur la figure.

## 3. L'exit tax en quatre questions (coût C)

Pas une figure, un encadré-décision de quatre lignes. Résident six des dix
dernières années ? Plus de 800 000 € de titres, ou plus de 50 % d'une société ?
Départ vers l'UE, l'EEE ou un pays lié par convention de recouvrement ? Titres
conservés deux ans, ou cinq au-delà de 2,57 M€ ? Un tableau à deux colonnes
(question, conséquence) rend mieux qu'un graphique, et il corrige une croyance
répandue chez les lecteurs FIRE, celle du compte-titres d'ETF hors champ.

## 4. La dérive du coût santé avec l'âge (coût B, sous réserve de sources)

Trois courbes de 45 à 80 ans, cotisation annuelle par personne : PUMa française
(plate, puisque la CSM s'éteint à la pension), CFE, assurance privée
internationale. Le passage « le poste oublié qui inverse les calculs » repose
entièrement sur cette pente, et la voir explique pourquoi la décision se prend
avant les ennuis de santé plutôt qu'après.

Réserve sérieuse sur le coût de production. Les barèmes CFE sont publics mais
datés, et les tarifs privés varient trop d'un assureur à l'autre pour qu'une
courbe unique soit honnête. À ne faire qu'avec des barèmes sourcés et une année
de référence affichée, sinon s'en tenir aux fourchettes du texte.


---

# Le facteur humain


## psychologie-du-retrait

Article sans aucune figure aujourd'hui. Deux passages portent une thèse
chiffrable mais restent purement verbaux : « on meurt riche dans la grande
majorité des futurs » et « espacez les relevés ». Ce sont les deux meilleures
cibles.

## 2. La probabilité de voir du rouge, selon la fréquence de consultation

Coût : **B** (calcul court sur les séries longues de `pkg/datasets`, un
comptage de fenêtres glissantes ; aucune donnée nouvelle).

Une courbe décroissante, ou cinq barres, donnant la part des fenêtres de
consultation qui affichent une perte, pour une fenêtre d'un jour, d'un mois,
d'un trimestre, d'un an, de cinq ans, sur le 60/40 réel. La chute est brutale
et elle transforme le conseil « espacez les relevés » en une décision chiffrée
sur laquelle le lecteur peut caler son propre calendrier. C'est aussi la
démonstration visuelle de l'aversion aux pertes myope citée dans le bloc
science : à information identique, la fréquence seule fabrique la douleur.

## 3. Les vingt ans de vie non vécue d'Hélène

Coût : **A** (les chiffres sont dans l'exemple de l'article).

Deux barres empilées côte à côte, cumulées sur vingt ans : ce que le plan
autorise, et ce qu'elle dépense si l'année 1 se répète. L'écart, hachuré et
nommé « vie non vécue », vaut plus d'un demi-million d'euros, qui finit en
legs involontaire. La vignette raconte trois ans, la figure montre où mène
la trajectoire si personne n'équipe la psychologie. C'est le seul endroit du
livre où la sous-consommation se lit comme une somme, pas comme un pourcentage.

## 4. Le prix de la capitulation, en deux chemins

Coût : **A/B** (séries bundle, calcul à écrire ; attention au doublon).

Deux courbes de patrimoine réel partant de mars 2009 et de mars 2020 : rester
investi selon la règle, contre vendre au creux et revenir douze mois plus tard.
L'écart terminal donne enfin un visage aux « 20-40 % de richesse finale » que
l'article cite de mémoire. Réserve : cette figure appartient peut-être plutôt
à [[cash-buffer]], où l'ancre 20-40 % est établie. À arbitrer avec le
relecteur de ce chapitre pour ne pas illustrer deux fois le même chiffre.

## Écarté

La comparaison « à richesse égale, les détenteurs de revenus garantis dépensent
plus » (Blanchett & Finke) serait une belle figure, mais aucune donnée du repo
ne la porte et deux barres recopiées d'un papier américain n'apprennent rien de
plus que la phrase. **C** : un encadré ou une ligne de tableau suffit.


## temoignages-fire

Article aujourd'hui sans aucune figure, comme presque toute la partie « Le facteur
humain » (seul `sequence-des-rendements` porte une figure dans le voisinage).
Trois idées, dont une seule vraiment graphique.

## 1. La fenêtre que le corpus a vécue (et les deux qu'il n'a pas vécues)

Trois trajectoires de patrimoine réel d'un 60/40 US à 4 % de retrait initial,
alignées sur l'axe « années depuis le départ » : départ 1966, départ 2000,
départ 2012. Sur la trajectoire 2012, une bande met en évidence les années
effectivement racontées par les blogs (les seules dont le corpus témoigne). Le
lecteur voit d'un coup pourquoi les bilans publiés sont rassurants : ils
couvrent le millésime le plus favorable de l'après-guerre, à un stade où 1966
et 2000 n'avaient pas encore montré leur visage.

Porte la mise en garde du constat 1 (« une seule décennie, favorable, et les
millésimes 2000 et 1966 n'ont pas de blogueurs »), qui est aujourd'hui une
phrase entre parenthèses alors qu'elle conditionne tout le chapitre, et le biais
du survivant du bloc méthode.

Coût **A** : `pkg/replay` fournit le 60/40 US réel depuis 1954 (S&P 500 +
Treasuries 5 ans, déflaté CPI), donc les trois chemins et le retrait fixe se
calculent sans nouvelle donnée. Attention au recouvrement : `retrait-fixe-bengen`
a déjà `bengen-millesimes` (millésimes en éventail) et `sept-facons-de-vivre`
a les replays 1973/1985/2000. La forme doit donc être différente et le
message aussi : ici l'axe est « années depuis le départ », la comparaison est
à âge de plan égal, et l'objet mis en scène est la fenêtre d'observation du
corpus, pas la règle de retrait.

## 2. La frise du motif temporel : lune de miel, mur, reconstruction

Une frise horizontale de 0 à 60 mois découpée en trois zones (euphorie 0-12,
creux 12-24, remontée ensuite), avec sous la frise les prescriptions des
vétérans posées au bon moment : prototyper avant le mois 0, convalescence en
année 1, construction en année 2, pratiques structurantes et lien social
entretenus en continu. La courbe en U du bien-être est citée dans trois
chapitres (`sens-et-identite`, `psychologie-du-retrait`, ici) et n'est nulle
part montrée.

Impératif s'il est retenu : pas d'axe vertical chiffré et pas de courbe de
« bien-être » inventée. Aucune série ne mesure ça, et une belle courbe fausse
serait pire que le texte actuel. La valeur de la figure est le calendrier des
gestes, pas une mesure.

Coût **B** : rien à calculer, mais tout à dessiner (plate v2, libellés
horizontaux). Si le dessin ne tient pas sans axe crédible, l'idée 3 fait
mieux pour moins cher.

## 3. Ce que le modèle voit, ce que le corpus raconte

Un tableau à deux colonnes en face à face. À gauche, les risques que les
soixante chapitres précédents savent chiffrer (séquence des rendements,
inflation, longévité, fiscalité, ruine). À droite, les crises que les récits
rapportent réellement (perte de structure, identité, lien social, anxiété des
premières années, désynchronisation du couple), avec pour chacune le chapitre
qui la traite. La thèse centrale du chapitre (« les crises du FIRE ne sont
presque jamais financières ») devient visible en un coup d'œil, et le tableau
sert de table d'orientation vers le reste de la partie humaine.

Coût **C** : un tableau, pas une figure. C'est le meilleur rapport
place/valeur des trois idées et il ne demande aucune donnée.


## sens-et-identite

État actuel : aucune figure. L'article est peu chiffré, donc la question n'est pas
« quelle série tracer » mais « quel passage verbal gagnerait à devenir une image
qu'on retient ». Trois candidats sérieux, un quatrième en tableau.

## 1. « Cinq choses, une remplacée » (la facture invisible du travail)

Ce qu'elle montre : les cinq apports du travail en cinq lignes (revenu,
structure, identité, lien, utilité), une seule marquée « remplacée par le plan
financier », les quatre autres marquées « à construire soi-même ». Une plaque
minimale, pas un graphique.

Pourquoi elle vaut la place : c'est la thèse de tout l'article, énoncée au
premier paragraphe et jamais visualisée. Le lecteur du livre arrive avec
soixante chapitres de finance derrière lui, et l'image lui dit en une seconde
que le plan qu'il vient d'apprendre à construire couvre un cinquième du sujet.
C'est aussi la figure la plus citable du chapitre.

Coût : **A** (les données sont dans l'article lui-même).

## 2. Marc et Sarah, deux voies parallèles décalées de deux ans

Ce qu'elle montre : deux pistes horizontales (une par personnage du bloc
exemple) sur un axe qui va de deux ans avant le départ à cinq ans après, avec
les mêmes jalons repérés sur chacune : prototypage, départ, creux, premier
engagement qui prend, identité stabilisée. Sarah atteint chaque jalon avant
Marc, et Marc paie un creux là où Sarah n'en a pas.

Pourquoi elle vaut la place : le bloc exemple se termine sur « même plan
financier, deux années de préparation d'écart », et cet écart est exactement ce
qu'un décalage visuel montre mieux qu'un paragraphe. C'est la figure la plus
argumentative de l'article, celle qui convertit le conseil n° 1 des vétérans
(prototyper) en évidence graphique.

Coût : **A** (tout est dans le bloc exemple ; les dates sont des repères
narratifs, à assumer comme tels dans la légende).

## 3. La frise du calendrier (2-3 ans avant, année 1, années 2-3, ensuite)

Ce qu'elle montre : une frise en quatre bandes reprenant la section calendrier,
chaque bande portant ses deux ou trois gestes (prototyper, récupérer, fixer les
rituels, entretenir), avec le creux de 6-24 mois placé à sa vraie position sur
l'axe.

Pourquoi elle vaut la place : la section calendrier est une liste de quatre
paragraphes gras qui décrit un objet spatial, un axe de temps. La frise rend le
chapitre actionnable et donne au lecteur le seul repère qui manque, où se situe
le mur par rapport à la date de départ.

Coût : **A**, mais attention au doublon. Cette idée et l'idée 2 occupent la même
case (le temps qui passe autour du départ). Si une seule doit être retenue, la
2 porte davantage la thèse ; la 3 est plus utile comme aide-mémoire. Les
fusionner est possible (la frise en fond, les deux trajectoires par-dessus), au
risque d'une plaque chargée.

## 4. Les quatre chantiers en tableau de revue (pas une figure)

Ce qu'il montre : quatre lignes (structure, identité, lien, utilité), trois
colonnes (le symptôme quand ça va mal, les outils, la question à se poser à la
revue annuelle). Le contenu existe déjà, éparpillé dans les quatre sections.

Pourquoi il vaut la place : la ligne non financière de la revue annuelle
([[revue-annuelle]]) demande « les quatre chantiers, ça va ? » et renvoie ici.
Un tableau donne au lecteur la grille de réponse, ce que la prose ne fait pas.

Coût : **C** (un tableau ou un encadré fait mieux qu'une figure ; aucune donnée
à produire).


## couple-et-famille

L'article est entièrement verbal (aucune figure aujourd'hui). Trois passages
portent une thèse chiffrable et gagneraient à être montrés.

## 1. Le divorce comparé aux pires krachs, en taux de retrait

Barres d'intervalle classées, une ligne par choc, mesurant une seule grandeur :
ce que le choc fait au **taux de retrait courant** du même plan, à l'instant du
choc puis dix ans après. Trois krachs de référence (1929-1932, 1966-1981,
2007-2009, rejoués sur le 60/40 réel de `pkg/replay`) et une ligne « divorce »
construite à la main dans l'article, capital divisé par deux et dépenses
individuelles multipliées par ~1,8, soit un taux courant multiplié par ~3,6.
La figure porte la thèse centrale du chapitre, que le texte affirme aujourd'hui
sans preuve, et surtout elle montre le point qui compte : la barre « dix ans
après » redescend pour les krachs et pas pour le divorce.

Coût : **B**. Les séries et les rejeux existent (`pkg/replay`, S&P 500 +
Treasuries réels depuis 1954, plus les millésimes anciens de `pkg/datasets`) ;
il faut écrire le petit calcul de taux courant à t et t+10 et poser les deux
hypothèses de la ligne divorce, qui doivent rester visibles dans la légende.

## 2. Décalage du second départ : combien de mois de salaire achètent quoi

Courbe unique, ruine centrale en ordonnée, durée du décalage en abscisse
(0, 6, 12, 18, 24, 36, 48 mois), sur un ménage type du livre. La pente dit
l'essentiel : le gros de la protection anti-séquence est acheté par les
premiers 12 à 24 mois, et la courbe s'aplatit ensuite. C'est exactement
l'argument « court et daté » de la section, aujourd'hui asséné sans chiffres,
et l'aplatissement est le seul contre-argument valable au décalage indéfini,
donc au coût humain que le texte décrit juste après.

Coût : **B**. Le moteur (`pkg/decumul`, flux externes annuels pendant la
fenêtre) fait le travail, mais le balayage et le choix du ménage de référence
restent à écrire. Une variante en deux courbes (avec et sans mutuelle plus
trimestres) surchargerait pour rien.

## 3. La frise des passifs familiaux datés

Frise horizontale sur trente ans, une bande par engagement, alignée sur les
barreaux de l'échelle obligataire : études des enfants (3 à 5 ans par enfant,
10 à 25 k€/an), fenêtre de donation d'installation (25-35 ans des enfants),
plage probable d'aide aux parents, provision dépendance de la fin. Le chapitre
répète que ces coûts sont « datables et provisionnables » ; une frise le prouve
d'un coup d'œil et montre le vrai problème, la superposition des bandes autour
de la soixantaine, quand études, parents et dépendance se chevauchent.

Coût : **A**. Toutes les données sont dans l'article et dans
[[echelle-obligataire]] ; c'est de la mise en forme, pas du calcul. Une
alternative honnête serait un simple tableau à trois colonnes (engagement,
fenêtre, montant annuel), mais la superposition, qui est le message, ne se voit
qu'en frise.


## flexibilite-realite

L'article est aujourd'hui sans aucune figure, et son thèse centrale (« la durée bat
la profondeur ») est un fait chiffré, pas une opinion. C'est le meilleur candidat de
la partie à recevoir une figure porteuse.

## 2. Trois escaliers de revenu servi, sur le plancher testé

Ce qu'elle montre : les trois versions du bloc « Le même plan, trois honnêtetés »
côte à côte, en petites-multiples. Pour chacune, le niveau de vie réellement servi
année après année dans le pire quartile, en escalier, avec deux lignes horizontales,
le confort visé et le plancher testé à 73 %. On voit d'un coup ce que le texte doit
énumérer : la version « phrase magique » vit neuf ans sous son plancher, la version
« flexibilité écrite » fait deux épisodes de trois à quatre ans qui ne touchent
jamais le plancher, et la version rigide ne coupe jamais jusqu'à la falaise.

Pourquoi elle vaut la place : l'article demande explicitement de « sortir de la
seule probabilité de ruine » pour lire le niveau de vie servi. Une figure qui montre
ce qu'il faut lire vaut mieux qu'un paragraphe qui l'ordonne. Elle referme aussi
l'article sur son critère, la profondeur ET la durée dans la même image.

Coût : **B** (une passe de décumulation sur les trois paramétrages, avec extraction
du percentile de dépense servie année par année). Une version honnête en **A** est
possible en dessinant les chiffres déjà écrits dans le bloc exemple, mais elle
serait un schéma, à étiqueter comme tel.

## 3. L'arithmétique du plancher (une barre décomposée)

Ce qu'elle montre : le budget de confort en une barre horizontale segmentée en
incompressible, compressible douloureux et compressible indolore, avec le trait du
plancher posé à « incompressible + la moitié du douloureux », puis la flèche du
recalage de 10 à 15 % vers le haut que produit le trimestre d'essai. Deux ou trois
profils en petites-multiples (le sobre à 81 % de ratio, le confortable à 72 %, le
faux flexible à 90 %) suffisent à montrer que le ratio se lit dans la composition
du budget, pas dans le tempérament.

Pourquoi elle vaut la place : le bloc « science » énonce une méthode en cinq gestes
que le lecteur doit reconstituer mentalement. La barre décomposée est la méthode
elle-même, et elle rend le seuil « 90 % = vous n'êtes pas flexible » évident au lieu
d'assené.

Coût : **A**. Les trois profils sortent de l'article et de `cas-types` (Jules,
plancher testé à 81 % du confort).

## 4. Un tableau récapitulatif des six formes

À la place d'une figure. La liste des six formes mélange trois grandeurs (la valeur
en points de taux, la douleur vécue, la durée de validité) qui ne se comparent pas
en lecture linéaire. Quatre colonnes, forme / valeur / douleur / péremption, avec
« expire avec l'âge » pour les revenus d'appoint, « une seule fois » pour le report,
« jamais comptée » pour la coupe de survie, rendraient le classement scannable et
révéleraient ce que le texte laisse implicite, à savoir que les deux formes les
mieux notées ne sont pas des coupes.

Coût : **C** (tableau, aucune donnée nouvelle).


## une-annee-de-plus

Article aujourd'hui sans aucune figure, alors que sa thèse est arithmétique
(« ce que l'année achète décroît, ce qu'elle coûte ne décroît pas »). Trois
idées, par ordre de valeur.

## 1. Les deux colonnes, sur le même axe (coût A/B)

Abscisse : le nombre d'années de travail supplémentaires, 0 à 5. Deux courbes,
même graphique. La sécurité achetée en cumulé (points de taux de retrait, ou
points de ruine effacés) monte puis s'aplatit dès la deuxième année. Le coût
cumulé en années pleines monte en ligne droite, une marche par année, et ne
s'aplatit jamais. L'écart entre les deux se retourne quelque part entre la
première et la deuxième année, et c'est exactement l'endroit où le syndrome
commence. C'est la figure centrale de l'article : elle rend visible en un
coup d'œil ce que quatre paragraphes doivent aujourd'hui affirmer.

Coût : **B**. La courbe de coût est triviale (une droite). La courbe de
sécurité demande un petit balayage, un plan cible relancé avec un an, deux
ans, ... de capital et d'horizon en plus, sur le moteur de décumulation du
repo, puis lecture du taux de retrait sûr ou de la ruine. À faire une fois,
avec les hypothèses du ménage de référence du livre.

## 2. Le stock d'années pleines, en pictogramme (coût A)

L'article dit « voici la statistique qui devrait s'afficher dans tous les
open-spaces du mouvement », puis la laisse en prose. Une bande de 20 carrés,
un par année pleine entre 45 ans et l'espérance de vie en bonne santé
(~64-65 ans), suivie des années restantes de l'espérance de vie totale dans
un ton éteint, avec limitation d'activité. Par-dessus, les deux ou trois
premiers carrés barrés, légendés « une année de plus », « deux », « trois ».
La forme dit ce qu'aucune phrase ne dit aussi vite : le stock est petit, il
est en tête de vie, et l'OMY se sert dedans en premier.

Coût : **A**. Les seules données sont dans l'article (Eurostat/DREES,
espérance de vie en bonne santé et totale). Attention, gabarit v2 obligatoire
et libellés horizontaux.

## 3. Le menu des mouvements équivalents, en tableau (coût C)

Le passage le plus actionnable de l'article est une liste de leviers de même
valeur, aujourd'hui noyée dans un paragraphe. Un tableau à trois colonnes
ferait mieux qu'une figure : le levier, la dose qui ramène la ruine sous le
seuil, et le prix payé (euros, contrainte, ou années pleines pour l'année de
travail). La ligne « une année de plus » y arrive dernière au prix, ce qui est
la démonstration voulue.

Coût : **C**, un tableau, pas une figure. Les doses restent à calculer si on
veut des chiffres propres, mais quatre ou cinq lignes suffisent et le message
tient sans précision décimale.

## Écarté

Une figure sur le legs involontaire (« les plans prudents meurent riches »)
serait redondante avec `bengen-millesimes`, désormais citée en toutes lettres
dans le paragraphe « ce qu'une année achète ».


## retour-au-travail

Article aujourd'hui sans aucune figure. Trois thèses chiffrées y sont posées
verbalement et gagneraient à être vues.

## 1. Le taux de change « revenu d'appoint contre capital », année par année (coût B)

Une seule courbe, ou une série de barres décroissantes : pour un plan de
référence (1 M€, 40 k€/an, l'exemple fondateur de `sequence-des-rendements`),
combien de capital supplémentaire faudrait-il pour égaler la baisse de ruine
qu'apportent 12 k€/an pendant 8 ans, selon que ces 8 ans démarrent en année 1,
5, 10, 15 ou 20 ? L'axe vertical est en euros de capital équivalent, ce qui
donne directement le « prix » de l'heure travaillée à chaque moment du plan.

Elle porte la thèse centrale de la section chiffrage (« l'heure travaillée
pendant la fenêtre fragile est la mieux payée de votre vie financière ») et,
du même coup, la thèse du bloc `attention` (l'option est une marge de première
décennie). Aujourd'hui les deux reposent sur une seule phrase et sur la
fourchette 200-300 k€, non montrée.

Coût B : le moteur (`pkg/decumul`, sweeps + solveur en mouvements équivalents)
fait déjà exactement ce calcul, mais il faut écrire le balayage sur l'année de
début et convertir en capital équivalent.

## 2. Le quadruplé français en cascade (coût A, une jambe en B)

Cascade à quatre marches partant de 10 k€ de revenu d'activité et arrivant au
total annuel revendiqué : le revenu lui-même, l'extinction de la CSM (~1,8 k€),
la valeur annualisée des 4 trimestres validés (proratisation plus points de
complémentaire), la couverture santé/prévoyance selon le statut. Une marche
« séquence amortie » peut fermer la cascade en gris, hors total, puisque
l'article la compte à part.

C'est la seule affirmation du chapitre qui donne un multiple (« 10 k€ valent
15-18 k€ de retraits évités ») sans montrer sa décomposition, et c'est aussi
l'argument que le lecteur français doit pouvoir vérifier poste par poste. La
cascade rend visible ce qui est solide (revenu, CSM) et ce qui est une
valorisation actuarielle discutable (les trimestres).

Coût A pour trois marches (chiffres déjà dans l'article, `taxe-puma` et
`retraite-legale`) ; la marche « trimestres » demande une petite actualisation
de pension à écrire, donc B pour elle seule. Utiliser la même forme de cascade
que le livre emploie ailleurs est ici le bon choix, la question étant
justement « d'où vient le total ».

## lexique

Remarque de cadrage : un glossaire alphabétique n'a pas de « passage verbal
qui porte une thèse » à illustrer par un graphique. Aucune figure au sens du
livre (série, fan, cascade) ne s'impose ici, et il ne faut surtout pas en
plaquer une. En revanche l'article promet deux choses dans son chapeau (« le
lexique se lit aussi », « la table d'orientation du livre ») qu'il ne tient
qu'à moitié, faute d'un accès autre qu'alphabétique. Les trois idées ci-dessous
visent exactement ce manque.

## 2. La table d'orientation thématique (tableau ou encadré)

Un second index, non alphabétique, en cinq ou six familles : mesure du risque et
des rendements, portefeuille et actifs, règles de retrait, cadre français,
modélisation, psychologie et pilotage. Chaque famille liste ses six à dix
termes dans l'ordre où on a intérêt à les lire, pas dans l'ordre de l'alphabet.

Pourquoi ça vaut la place : le chapeau annonce un lexique qui « se lit », et
c'est aujourd'hui contredit par la seule entrée alphabétique disponible, qui
range Bengen loin de Trinity et le buffer loin du glidepath. Cette table donne
au lecteur qui découvre le sujet un parcours, et au lecteur avancé une carte de
ce qu'il a lu. Coût : **C** (pas de figure, un tableau fait mieux, et les
regroupements se déduisent des wiki-liens déjà présents dans chaque entrée).

## 3. La plaque des cinq carrefours du vocabulaire (figure, si une seule)

Si une seule vraie figure devait exister ici, ce serait un schéma sobre à cinq
nœuds, les cinq concepts dont tous les autres dépendent (risque de séquence,
moyenne géométrique, régimes de marché, probabilité de ruine, plancher de
dépenses), chacun relié aux termes qu'il commande. Labels horizontaux, pas de
flèches décoratives, le graphe reste lisible sans couleur.

Pourquoi ça vaut la place : elle dit visuellement ce que le livre répète en
prose, à savoir que le vocabulaire du retrait n'est pas une liste mais une
hiérarchie, et que cinq idées suffisent à tenir les quatre-vingts autres.
Coût : **B** (rien à calculer, mais le placement des nœuds et le tri des arêtes
sont un travail de mise en page à part entière ; à ne tenter qu'après les idées
1 et 2, qui rendent le même service pour beaucoup moins cher).


## bibliotheque

Article-annuaire. Une figure « graphique » y est presque toujours décorative, car
le texte est déjà une liste structurée. Les deux meilleures idées sont donc un
tableau et une mise en page ; une seule vraie figure mérite d'être discutée.

## 2. La descente du taux publié, 1994-2025 (figure, coût B)

Un axe temporel horizontal, un point par étude fondatrice, en ordonnée le taux
de retrait que l'étude publie : Bengen 1994, Trinity 1998, Guyton-Klinger 2006
(taux initial admissible), Morningstar 2021 à 2025 (la série 3,3 → 3,8 → 4,0 →
3,7 %), Anarkulova-Cederburg-O'Doherty-Sias 2023. Chaque point porte en étiquette
son horizon et son critère de succès, sans quoi la figure mentirait, car ces
taux ne mesurent pas la même chose.

Ce que ça porte : la section « papiers » affirme que la recherche a déplacé les
bornes vers le bas, et le lecteur doit la croire sur parole. La figure montre à
la fois la baisse et sa cause visible, à savoir le changement d'échantillon et
d'horizon plutôt qu'un changement d'avis. Coût B, car les valeurs existent dans
le livre mais éparpillées sur cinq articles, et il faut une table de correspondance
propre horizon/critère/portefeuille. Réserve honnête : la figure servirait peut-être
mieux [[la-regle-des-4-pourcents]] ou [[anarkulova-cederburg]] qu'ici, où elle
risque de doubler leur propos.

## 3. Les trois parcours en trois colonnes (mise en page, coût C)

Le bloc « astuce » final empile trois itinéraires dans un seul paragraphe, avec
des flèches et du gras qui se disputent la ligne. Trois colonnes parallèles, un
titre de profil par colonne, quatre à cinq étapes numérotées dessous, rendent la
comparaison instantanée et donnent envie de suivre l'un des trois. Aucun chiffre,
aucune donnée nouvelle, juste la même matière rendue scannable. C'est la
transformation la plus rentable de l'article.


---

# La science du retrait


## etude-trinity

Article sans aucune figure aujourd'hui, alors qu'il porte deux objets visuels
canoniques du domaine (le graphique de Bengen, la grille de Trinity). Trois
idées, par ordre de valeur.

## 1. La falaise de Trinity (coût B)

Le taux de succès en ordonnée, le taux de retrait en abscisse (de 3 à 8 %), une
courbe par allocation (100/0, 75/25, 50/50, 25/75). Le tableau de l'article
donne quatre lignes ; la courbe donne la **forme**, c'est-à-dire exactement ce
que le texte affirme sans pouvoir le montrer, l'effondrement entre 4 et 5 % et
le croisement des allocations (la courbe 25/75 décroche bien avant les autres,
la courbe 100 % actions passe sous la 75/25 en haut de gamme). C'est le passage
« trois enseignements durables » rendu lisible d'un coup d'oeil.

Coût B : il faut recalculer la grille (fenêtres glissantes de 30 ans sur le
60/40 réel de `pkg/replay`, ou sur S&P 500 + Treasuries + CPI de
`pkg/datasets`, par pas de 0,25 point de taux et par allocation). Rien de neuf
côté données, seulement une boucle à écrire. Ne pas reprendre les chiffres
publiés de Trinity dans une figure maison, sous peine de mélanger deux sources.

## 2. Le graphique fondateur : années de survie par millésime (coût A/B)

L'article dit « le résultat tient dans un graphique resté célèbre » puis le
décrit en prose sur quatre phrases. Autant le montrer : une barre par millésime
de départ (1926 à 1995), la hauteur = nombre d'années tenues, une bande par
taux (3, 4, 5, 6 %), et la ligne horizontale des 30 ans. Le SAFEMAX devient
visuellement ce qu'il est, le taux le plus haut dont toutes les barres restent
au-dessus de la ligne, et 1966 se voit immédiatement comme le point bas.

Coût A/B : les séries longues sont dans le repo (S&P 500 mensuel, Treasuries,
CPI), le calcul est le même moteur que l'idée 1. Attention au recouvrement avec
`bengen-millesimes` (retrait-fixe-bengen.md), qui montre déjà des issues par
millésime, mais en capital final et à un seul taux. L'angle distinctif ici est
la **définition du SAFEMAX**, donc la comparaison entre plusieurs taux ; si la
planche ne peut pas porter les quatre taux, elle ne vaut pas la place.

## 3. Trois millésimes, trois vies (coût A)

Le capital réel année par année de trois départs sur le 60/40 américain à 4 %
indexé : 1966 (ruine à l'année 29), 2000 (encore sous la moitié de sa valeur
initiale après un quart de siècle) et 2009 (au-dessus du départ, presque le
double). Trois courbes, même règle, même portefeuille, seule la date change.
C'est l'encadré « Lire un millésime » de l'article, qui est aujourd'hui une
description de graphique sans le graphique.

Coût A : vérifié dans `pkg/replay` (Reference + Annual), trois boucles, aucune
donnée à produire. Caveat : `sequence-risk` (sequence-des-rendements.md) montre
déjà deux retraités, mais synthétiques et symétriques ; ici les trois
trajectoires sont réelles et datées, ce qui est le propos de cet article-ci. Si
une seule figure doit passer, préférer l'idée 1 ou 2 ; celle-ci est la plus
facile, pas la plus nécessaire.


## sequence-des-rendements

L'article porte déjà une figure (`sequence-risk`, deux retraités, krach tôt ou
tard). Elle couvre le mécanisme abstrait. Ce qui reste purement verbal et
porteur, c'est (a) l'ancrage historique 1966 contre 1982, (b) la concentration
du risque sur la première décennie, (c) la hiérarchie des parades.

## 2. Le profil d'importance année par année (coût B)

Une seule courbe (ou une rangée de barres), abscisse = rang de l'année de
retraite, 1 à 40, ordonnée = part de la variance de l'issue finale expliquée par
le rendement de cette année-là seule. Le profil s'effondre : très haut sur les
années 1 à 10, quasi plat ensuite. Une bande grisée sur les dix premières années
matérialise la « fenêtre fragile », et l'aire sous cette bande porte le chiffre
des trois quarts cité dans le texte.

Pourquoi elle vaut sa place : la section « La fenêtre fragile » énonce un
chiffre (« aux trois quarts dans sa première décennie ») sans jamais le montrer,
et ce chiffre commande ensuite les trois conséquences pratiques, le glidepath,
la date de départ et les seuils datés. Une décroissance vue une fois se retient
pour toujours, et cette forme, un profil d'importance par rang temporel, n'existe
pas encore dans le livre.

Coût B : il faut écrire la décomposition (régression de l'issue finale sur le
rendement de chaque année, ou variance expliquée par permutation) sur un jeu de
trajectoires simulées ou sur les cohortes historiques. Les données sont là, le
calcul est à écrire, et il fixerait au passage le chiffre exact des trois quarts.

## 3. La ruine décomposée par quartile de première décennie (coût B)

Quatre barres empilées, une par quartile de rendement réel des dix premières
années, chacune découpée en « ruiné / abîmé mais vivant / confortable / opulent ».
Le premier quartile est presque tout rouge, le quatrième presque tout vert.

Pourquoi elle vaut sa place : c'est exactement la lecture que l'encadré
« Mesurer le risque de séquence chez vous » décrit en trois phrases de prose
abstraite. La montrer une fois donne au lecteur l'image mentale qu'il devra
retrouver sur son propre plan. À arbitrer contre l'idée 2 : les deux servent la
même section et une seule suffit sans doute. L'idée 2 est plus élégante, la 3
est plus directement actionnable.

Coût B : même famille de calcul que l'idée 2, tri des trajectoires par rendement
de la première décennie puis classement des issues.

## 4. Les variantes A/B/C de l'encadré « exemple » : ne pas illustrer (coût C)

Quatre nombres de ruine (18, 12, 8, 7, 3 %) se lisent mieux dans la phrase que
dans un graphique de barres, et la hiérarchie annoncée tient en une ligne. Si
l'on veut la rendre visible, un petit tableau à trois colonnes (variante, coût
consenti, ruine) ferait mieux qu'une figure, en ajoutant la colonne qui manque
aujourd'hui, le prix payé pour chaque parade.


## ruine-et-probabilites

Article aujourd'hui sans aucune figure. Trois passages portent la thèse et sont
100 % verbaux : la précision illusoire, la nature de l'échec (quand il frappe),
et la ruine réelle lente. Ordre de priorité ci-dessous.

## 1. L'échelle des trois bruits (coût B, la meilleure)

Une seule règle horizontale graduée en points de ruine (0 à 20 %), et trois
segments empilés verticalement qui montrent l'amplitude de chaque source
d'imprécision pour UN même plan : le bruit d'échantillonnage à 2 000
trajectoires (un segment étroit, environ 4 à 6 % autour d'un vrai 5 %), la
sensibilité des paramètres (±0,5 point de rendement réel, environ 4 à 9 %), le
choix du modèle (2 à 14 %). Un marqueur unique « le chiffre affiché : 4,7 % »
posé dessus.

C'est la thèse centrale de la section « précision illusoire », qui ne s'appuie
aujourd'hui que sur des chiffres jetés en prose. Le lecteur voit d'un coup que
la barre du haut, celle qu'il regarde, est écrasée par celles du bas, et que
2 % et 8 % tombent dans le même segment. Coût B : le bruit d'échantillonnage
est en forme fermée, les deux autres segments demandent quelques exécutions du
moteur de décumulation sur un plan de référence.

## 2. Quand la ruine frappe, et qui est encore là pour la voir (coût B)

Deux plans de MÊME probabilité de ruine (8 %), côte à côte : pour chacun, la
distribution des âges d'épuisement des trajectoires ruinées (barres), et par
dessus, sur le même axe des âges, la courbe de survie Gompertz. L'un échoue
autour de 70-75 ans, l'autre après 88 ans, où la moitié des porteurs sont déjà
morts.

Elle porte deux passages d'un coup, « la ruine est binaire et terminale, deux
plans à même taux sont des vies différentes » et « le chiffre ignore
superbement votre mortalité ». C'est aussi le seul endroit du livre où la
correction de mortalité devient visuelle plutôt qu'argumentée. Coût B : dates
de ruine et table de mortalité existent dans le moteur, la mise en forme est à
écrire.

## 3. La falaise contre la pente (coût A)

Une trajectoire historique réellement défaillante (départ 1966, 60/40 réel
américain, données déjà embarquées dans `pkg/replay`), tracée en deux
registres superposés : en haut le capital réel qui décroche puis s'épuise, en
bas le taux de retrait courant qui monte marche par marche, avec les bandes
verte, orange, rouge de [[quand-s-inquieter]] et deux repères datés, l'année où
le voyant passe à l'orange et l'année de l'épuisement. L'écart entre les deux
repères, annoté en clair, matérialise le préavis de 8 à 15 ans.

C'est l'argument le plus important de la fin de l'article, celui qui doit faire
dormir le lecteur, et il n'existe aujourd'hui que sous forme d'affirmation.
Coût A : la série et les millésimes sont déjà dans le repo.

## 4. Le seuil selon vos filets (coût C, encadré ou tableau)

Pas une figure. Un petit tableau de trois ou quatre profils (quadragénaire
employable et propriétaire, couple avec deux pensions, sexagénaire sans
recours, indépendant à revenus irréguliers), avec en colonnes les filets
présents, la ruine simulée acceptable, et ce qui déclenche la révision. La
section « choisir son seuil » est une liste de trois critères que le lecteur
doit croiser lui-même ; le tableau fait ce croisement sous ses yeux et se lit
en dix secondes. Un graphique n'apporterait rien ici.


## rendements-arithmetiques-geometriques

Figure déjà présente : `vol-drag` (deux trajectoires de même moyenne arithmétique,
×7,6 contre ×4,5 sur 30 ans). Elle porte bien l'ouverture. Les idées ci-dessous
visent les passages encore purement verbaux ou tabulaires.

## 1. La courbe du drag : pourquoi doubler la volatilité quadruple le coût (coût A)

Une seule courbe, géométrique obtenue en ordonnée, volatilité en abscisse, à
moyenne arithmétique constante (7 %), soit y = 7 − σ²/2. On y pose les points du
tableau existant comme des repères nommés horizontalement (monétaire ~1 %,
60/40 ~10 %, actions mondiales ~15 %, 90/60 ~15 %, émergentes ~22 %, ×2
quotidien ~30 %). Deux flèches annotées suffisent ensuite à raconter les deux
passages qui suivent le tableau : vers la gauche à hauteur constante pour la
diversification (moins de σ, même moyenne, donc plus de géométrique), et vers la
droite pour le levier quotidien, qui déplace deux fois plus loin sur l'axe des
abscisses que sur celui des ordonnées.

Elle vaut sa place parce que le tableau, lui, ne montre pas la seule chose qui
compte vraiment ici : la non-linéarité. Une colonne de nombres se lit comme une
suite de cas ; la courbe montre d'un coup que le coût s'effondre à gauche et
explose à droite, et c'est exactement ce que la prose demande au lecteur de
croire sur parole. Données déjà dans l'article, formule fermée, aucun calcul
externe.

## 2. Ce qu'il reste du « 8 % » annoncé (coût A)

Une barre horizontale unique, de 0 à 8 %, découpée en quatre segments lisibles :
6,9 % après drag, 4,4 % après inflation, 3,4 % net de frais, le reste étant la
part évaporée. Un seul repère vertical à la moitié de la barre montre que le
chiffre vivable passe sous la moitié du chiffre annoncé, sans qu'aucun mensonge
n'ait été commis.

Elle sert l'encadré « Le test du vendeur », qui est le passage le plus
opérationnel de la page et aujourd'hui 100 % verbal. Attention toutefois à ne
pas la dessiner en cascade à étages : le livre a déjà `cascade-4pct` dans
[[les-maths-du-4-pourcent]], et deux cascades voisines se concurrenceraient. La
barre unique, plus modeste, tient le rôle sans empiéter. Tous les nombres sont
dans l'encadré.

## 3. L'approximation σ²/2 confrontée à l'histoire (coût B)

Nuage de points sur les fenêtres glissantes de 10 ans d'une longue série réelle
(le 60/40 américain réel de `pkg/replay` depuis 1954, ou le S&P 500 réel des
datasets) : en abscisse le drag prédit σ²/2 de la fenêtre, en ordonnée l'écart
réellement observé entre moyenne arithmétique et CAGR de la même fenêtre, avec
la première bissectrice tracée. Les points doivent se coller à la diagonale,
avec une légère dérive au-dessus pour les fenêtres les plus agitées, où les
termes d'ordre supérieur mordent.

C'est la seule figure qui prouve au lieu d'affirmer. L'article qualifie
l'approximation de « remarquablement précise » et ne l'établit que sur un
exemple jouet à deux années alternées. Sa place naturelle est la section « Pour
les curieux », qui justifie déjà un objet plus technique que le reste de la
page. Coût B : les séries existent, mais le calcul des fenêtres glissantes reste
à écrire.


## anarkulova-cederburg

Article aujourd'hui sans aucune figure, alors que sa thèse est intégralement
quantitative (« le taux sûr dépend de l'échantillon »). Trois idées, classées
par valeur.

## 2. Les deux courbes d'échec qui se croisent (coût B)

Probabilité de ruine (axe y) en fonction du taux de retrait rigide (axe x, de
2 à 5 %), deux courbes : histoire américaine seule, et rééchantillonnage
broad-sample. Une ligne horizontale à 5 % d'échec, et deux repères verbaux là
où chaque courbe la croise. La figure met en scène exactement les deux chiffres
que l'article martèle (17 % d'échec à 4 % contre ~2 % aux États-Unis, taux à
5 % d'échec vers 2,26 %), et surtout elle montre ce qu'aucune phrase ne montre :
l'écart n'est pas un décalage constant, il explose dans la zone où tout le
monde se dimensionne. C'est aussi la meilleure illustration possible de la
règle de conduite finale (« gardez les deux bornes affichées côte à côte »),
puisque la figure *est* les deux bornes côte à côte. Coût B, un balayage de
taux sur les deux sources.

## valorisations-et-cape

Figure déjà en place : `cape-swr` (taux de retrait soutenable en fonction du CAPE
de départ). Elle couvre la thèse « le SAFEMAX est une fonction du prix d'entrée ».
Les idées ci-dessous couvrent des passages qui restent aujourd'hui purement verbaux.

## 2. La règle du rang : où se situe le CAPE du jour, dans son siècle et dans son époque

Deux règles horizontales superposées, l'une graduée sur la distribution complète
1881-2026, l'autre sur les quarante dernières années seulement. Sur chacune, un
curseur pour le CAPE du jour, plus quelques repères datés (1982, 2000, 2009, 2021).
Le curseur est en haut de la première règle et nettement moins extrême sur la
seconde.

Ce serait la traduction visuelle de l'encadré « Le contresens de la moyenne mobile »,
qui est aujourd'hui le passage le plus utile de l'article et n'a aucun support :
lisez le rang dans votre propre époque, pas l'écart à une moyenne éternelle de 17.
Une figure qui apprend un geste, pas un fait, et que le lecteur peut refaire chaque
année.

Coût : **A**. Deux quantiles calculés sur le même CSV. Attention à ne pas dater la
figure au point de la périmer : les repères historiques portent la lecture, le
curseur du jour n'est qu'un exemple.

## 3. Le prix du signal : ce que coûte la sortie sur CAPE haut

Deux courbes de patrimoine réel depuis 1990, base 100 : rester investi, contre
sortir des actions dès que le CAPE dépasse sa moyenne longue et revenir quand il
repasse dessous. L'écart terminal se lit d'un coup d'œil, et les zones grisées
marquent les périodes « hors du marché », dont les dix ans d'affilée à partir de
1992.

Le texte affirme que « sortir du marché sur signal CAPE est la stratégie qui a ruiné
le plus de gens prudents » et que « toutes les études le confirment ». C'est
l'affirmation la moins étayée de l'article, et c'est celle qui protège le lecteur du
mésusage le plus coûteux. Une seule courbe la démontre mieux qu'un paragraphe.

Coût : **B**. Les données existent (CAPE + S&P 500 réel + un proxy monétaire réel
dans les séries bundlées), mais la règle de sortie et de rentrée doit être écrite et
justifiée, et il faut choisir un seuil défendable, en signalant que le résultat
dépend un peu du seuil retenu.


## rendements-attendus

Article aujourd'hui sans aucune figure, très dense en chiffres. Quatre idées,
par ordre de priorité.

## 3. Prévu contre réalisé, à dix ans (priorité 3, coût B)

Nuage de points : en abscisse la prévision naïve building-blocks du départ
(earnings yield + 1,75 point de croissance), en ordonnée le rendement réel
effectivement réalisé sur les dix années suivantes, un point par mois de départ
depuis 1881, avec la diagonale parfaite et une marginale triée des erreurs sur
le côté. Le nuage est large mais franchement orienté, et les millésimes chers
(CAPE > 30) surlignés se rangent tous dans le bas du graphique.

Pourquoi elle vaut sa place : c'est l'illustration exacte du bloc « Quelle
précision en attendre ? », qui affirme aujourd'hui sans preuve « erreur moyenne
de ± 2 à 3 points, mais bien mieux que le rétroviseur ». Le nuage montre les
deux à la fois : la dispersion, et la pente qui existe malgré elle. On peut
ajouter en gris le même nuage pour le rétroviseur (rendement des dix années
précédentes en abscisse), dont l'orientation est nulle voire inverse. C'est
l'argument central de la page rendu vérifiable.

Coût B : le calcul est à écrire, mais les données sont dans le repo (série
Shiller CAPE et prix/CPI, ou la série S&P réelle de `pkg/replay`). Attention à
ne pas déborder sur [[valorisations-et-cape]], qui traite déjà le R² : ici le
sujet est l'erreur de prévision, pas la relation CAPE-SWR.

## 4. La prudence comptée trois fois (priorité 4, coût B, ou C en encadré)

Cascade descendante du μ retenu : prior mélangé 4,5 % arithmétique, puis ancre
CAPE, puis abattement manuel, puis lecture sur le seul modèle du siècle mondial,
chaque marche annotée du coût en années de travail supplémentaires qu'elle
impose au plan. La dernière colonne donne le total, qui est le prix payé pour
avoir compté la même prudence trois fois.

Pourquoi elle vaut sa place : la section « sans double-compter la prudence » est
un avertissement fort mais abstrait ; chiffrer l'empilement en années le rend
concret et le relie à [[une-annee-de-plus]].

Coût B (il faut faire tourner le solveur d'horizon sur quatre calibrations). Si
le calcul paraît trop lourd ou trop dépendant d'un profil particulier, un simple
encadré-tableau à quatre lignes (calibration, μ réel, taux soutenable) fait
presque le même travail, coût C.


## horizon-et-esperance-de-vie

Figure déjà en place : `horizon-flatten` (courbe taux-horizon qui s'aplatit). Elle
couvre entièrement la section 2. Les trois idées ci-dessous portent les sections
1, 3 et 4, aujourd'hui purement verbales.

## 1. Les trois courbes de survie d'un couple, avec les quantiles marqués (coût B)

Trois courbes de survie sur un même axe des âges, de 45 à 105 ans : l'homme
seul, la femme seule, et le dernier survivant du couple. Trois repères verticaux
en travers : la médiane, le 85e et le 90e percentile de la courbe « dernier
survivant ». Le lecteur voit d'un coup les deux erreurs de la section 1, la
courbe du couple est très à droite de chacune des deux courbes individuelles, et
le 90e percentile est très à droite de la médiane. C'est l'argument central de
l'article, celui qui justifie les 53 ans d'horizon de Léa et Sam, et il n'existe
aujourd'hui que sous forme de trois paragraphes et d'un tableau d'espérances
moyennes qui dit exactement le contraire de ce qu'il faut regarder.

Coût B : les primitives existent (`decumul.Gompertz`, `FrenchMortality`,
`Survival`, `CoupleSurvival` dans `pkg/decumul/mortality.go`), mais la loi
embarquée est unisexe (mode 88, dispersion 10). Il faut caler deux jeux de
paramètres homme/femme sur les tables INSEE avant de tracer, sinon les deux
courbes individuelles se superposent et la figure perd son propos.

## 3. L'horizon coupé en deux : le pont et le régime de croisière (coût B)

Une frise du plan de 47 à 100 ans, avec en fond deux zones contrastées, la phase
à découvert et la phase adossée, et par-dessus la courbe du taux de retrait
**net** du portefeuille. Elle part vers 4 %, tient son plateau pendant la
traversée, puis décroche sous 2 % à la liquidation des pensions et n'en ressort
plus. La section 4 repose entièrement sur cette image mentale et sur le
recadrage psychologique de l'encadré terrain (« 18 ans à traverser », pas
« 50 ans à tenir ») ; une frise la donne en une seconde là où le texte demande
trois paragraphes.

Coût B : besoin d'un profil de pension type et d'un plan de dépenses, mais aucun
tirage aléatoire, un seul scénario déterministe suffit et vaut mieux ici.

## Ce qu'il ne faut pas illustrer

Le tableau des espérances conditionnelles par âge (section 1) est déjà la bonne
forme : quatre lignes, deux colonnes, lecture immédiate. En faire une courbe le
rendrait moins lisible, pas plus (coût C).


## serie-ern

Article sans figure aujourd'hui. C'est une carte de lecture d'un corpus de
60 volets, et tout y est verbal, y compris la hiérarchie des résultats. Deux
choses gagneraient vraiment à être vues : où sont les volets, et combien vaut
chaque résultat.

## 1. La carte des volets (coût B)

Une grille des ~60 volets, un petit carré numéroté par volet, regroupés en
blocs thématiques (socle, séquence, flexibilité, CAPE, buckets, glidepaths,
actifs, à-côtés) et colorés par verdict : démonté, validé, nuancé. Les volets
cités comme portes d'entrée (1, 26, 2-3, 54, 23-25, 58, 19-20, 12, 48) sont
marqués d'un point.

Pourquoi elle vaut la place : l'article promet « la carte » et la livre en
paragraphes. Une grille rend d'un coup d'œil ce que six paragraphes déroulent,
et le lecteur y revient (c'est la fonction d'index de la page). Elle rend
visible aussi le fait central du corpus, à savoir que la majorité des volets
sont des démolitions et une minorité des validations.

Données : entièrement dans l'article et dans la table des matières publique de
la série. Le coût est celui d'un tracé nouveau (grille + légende de verdict),
pas d'un calcul.

## 2. Le barème des leviers, en points de taux de retrait (coût A/B)

Barres d'intervalle classées, un levier par ligne, axe horizontal en points de
SWR (de −1 à +0,5). Passer d'un horizon de 30 à 60 ans : −0,5 à −0,75. Partir à
CAPE > 30 plutôt qu'à CAPE 15 : −0,5 et plus. Chaque point d'inflation moyenne
sur les dix premières années : −0,2. Glidepath 60 → 100 % : +0,1 à +0,3.
Flexibilité réaliste et bornée : +0,1 à +0,3. Matelas de cash rechargé : ~0.

Pourquoi elle vaut la place : c'est la thèse implicite de tout l'article, et
elle n'est écrite nulle part d'un bloc. Elle porte à elle seule le « filtre du
perfectionnisme » de la fin, car on voit que les deux premiers leviers pèsent
un ordre de grandeur de plus que les raffinements d'allocation. C'est aussi la
figure qui rend le passage sur la flexibilité et celui sur les glidepaths
comparables entre eux, alors que le texte les présente l'un comme une déception
et l'autre comme un succès pour des gains identiques.

Coût : les six valeurs existent, dispersées dans le livre
([[horizon-et-esperance-de-vie]], [[valorisations-et-cape]],
[[inflation-et-taux-de-retrait]], [[glidepaths]], [[flexibilite-realite]],
[[cash-buffer]]). Le travail est de les figer en un barème cohérent (même
horizon, même définition du SAFEMAX) avant de tracer.

## 3. Les deux biais qui se compensent (coût A)

Une règle graduée horizontale en % de taux de retrait, de 2 à 4,5 %, portant
les trois bornes du livre (mondial 2,3-2,7 ; ERN US 3,25-3,5 ; prospectif
Morningstar ~3,7) et, au-dessus, deux flèches opposées : l'échantillon
américain qui pousse le chiffre vers le haut, le critère du pire millésime qui
le tire vers le bas.

Pourquoi elle vaut la place : l'avant-dernière puce de « l'essentiel » affirme
que les deux biais se compensent partiellement, ce qui est l'énoncé le plus
utile de la page et le plus difficile à saisir en une lecture. Une règle
graduée le montre en une image et situe ERN entre les deux autres écoles.

Coût A, les trois bornes sont déjà chiffrées dans le livre. Attention à ne pas
recouper la figure `cape-swr` de [[valorisations-et-cape]], qui répond à une
autre question (le taux conditionnel au prix d'entrée).


## les-maths-du-4-pourcent

L'article porte déjà la figure `cascade-4pct`, qui couvre la décomposition
(4,0 → +1,8 → −1,8 → 4,0). Les idées ci-dessous complètent, elles ne
redoublent pas : elles portent des passages aujourd'hui purement verbaux.

## 2. La fonte du bonus d'amortissement avec l'horizon (coût A)

L'étage 2 est celui qui parle le moins aux lecteurs FIRE, alors que c'est celui
qui les concerne le plus : ils partent à 45 ans, pas à 65. Une seule courbe,
taux d'amortissement à 4 % réel en fonction de l'horizon (10 à 60 ans), avec
son asymptote horizontale à 4 % et trois points annotés (10 ans ≈ 12,3 %,
30 ans ≈ 5,8 %, 50 ans ≈ 4,7 %), montre d'un regard que le bonus s'écrase très
vite puis ne bouge presque plus. La conséquence est contre-intuitive et vaut la
place : passer de 30 à 50 ans coûte cher, mais passer de 50 ans à « pour
toujours » ne coûte presque plus rien.

Pure formule de mensualité, rien à calculer d'autre qu'un `r/(1-(1+r)^-n)`.
La forme (courbe à asymptote) n'existe pas encore dans le livre.

## 3. La marge cachée : capital final par millésime (coût B)

Le passage « la médiane subventionne la queue » affirme que 90 % des millésimes
finissent au-dessus du capital initial, et l'affirmation reste abstraite tant
qu'on ne voit pas la distribution. Des barres classées par capital final réel
(4 % rigide, 30 ans, chaque millésime de départ), en échelle logarithmique,
donnent l'image exacte : une longue pente qui culmine à cinq ou six fois la
mise, et une poignée de millésimes qui frôlent le zéro à droite. C'est aussi la
figure qui justifie l'asymétrie assumée du plan, et elle rend concret le mot
« calibré sur le pire cas ».

Coût B : le 60/40 réel américain de `pkg/replay` fournit la série, mais il
démarre en 1954, donc pas de millésime 1929 ni de fenêtre complète après 1995.
À décider avant de la produire : soit on assume l'échantillon court en le
disant, soit on la range en attendant une série plus longue. C'est aussi
l'occasion de vérifier le « 90 % » du texte, qui n'est aujourd'hui adossé à
aucune source du livre.


## decider-sous-incertitude

Figure déjà en place : `utilite-ce` (courbe d'utilité concave, loterie 20/65 k€,
espérance 42,5 k€, équivalent certain 36 k€). Elle porte bien la section utilité.
Les passages ci-dessous restent, eux, entièrement verbaux.

## 1. La cloche de Kelly, et ce que coûte d'en dépasser le sommet (coût B)

Courbe du taux de croissance géométrique en fonction de la fraction misée
(g(f) = f·μ − f²σ²/2, μ et σ de marché actions), avec quatre repères horizontaux
sur l'axe des abscisses : tiers de Kelly, demi-Kelly, Kelly complet, double
Kelly (croissance retombée à zéro). Elle montre d'un coup les deux thèses du
bloc science : le sommet est plat à gauche (le demi-Kelly garde environ les
trois quarts de la croissance) et la pente est brutale à droite (une erreur
d'estimation qui vous pousse en sur-Kelly détruit ce qu'elle prétendait
maximiser). Le bloc science est aujourd'hui le passage le plus dense et le seul
sans support visuel. Une seconde échelle discrète sous la courbe, la baisse
maximale médiane associée à chaque fraction, achève de disqualifier le Kelly
complet pour un rentier. Coût B : formule fermée, aucune série à charger, mais
un petit calcul à écrire et des paramètres à assumer.

## 2. Le plateau et la crête (coût B)

Deux courbes superposées sur le même axe « part d'actions, 0 à 100 % ». En trait
plein, le taux de retrait soutenable : plateau quasi horizontal entre 50 et 80 %,
chutes nettes de part et d'autre. En trait fin, la même courbe calculée avec un
rendement actions relevé d'un point : le plateau se déplace à peine, mais son
maximum ponctuel saute d'une vingtaine de points d'actions. Le lecteur voit alors
pourquoi la phrase « l'allocation optimale est à 72 % » est du bruit, et pourquoi
le milieu du plateau est la seule position robuste. C'est la thèse de toute la
section robustesse, aujourd'hui purement affirmée. Données dans le repo
(`pkg/replay` : 60/40 réel américain depuis 1954, actions et obligations
séparées ; ou `pkg/datasets` pour une fenêtre plus longue) ; le balayage par
part d'actions reste à écrire, d'où B.

## 3. Tolérance × capacité, le minimum commande (coût A)

Petite grille 2 × 2, tolérance en abscisse, capacité en ordonnée, avec la
diagonale de contrainte et deux points nommés du texte : l'ancien trader
flegmatique au plan tendu à 4,5 % (tolérance haute, capacité basse) et l'anxieux
assis sur 50 fois ses dépenses (l'inverse). Une teinte unique marque, pour chaque
case, le curseur qui commande. C'est le passage le plus contre-intuitif de
l'article, et le plus facile à retenir sous forme de plan ; il tient en une
figure sans aucune donnée. Coût A : tout est dans l'article.

## 4. Deux plans, deux distributions (coût B, à ne faire que si la place existe)

Le bloc exemple compare Plan A et Plan B par des chiffres en ligne. Deux bandes
verticales de distribution côte à côte (p5, médiane, p95, moyenne marquée d'un
signe distinct, équivalent certain d'une utilité logarithmique en trait épais)
rendraient visible l'inversion du classement selon le critère : A gagne à la
moyenne, B gagne à l'équivalent certain. Utile, mais redondant avec la figure
`utilite-ce`, qui fait déjà le même point sur une loterie plus simple. À garder
en réserve, et seulement si le bloc exemple devient une section à part entière.
Coût B : une simulation à lancer pour produire les deux distributions.


---

# Demarrer


## fire-cest-quoi

Article d'entrée du livre, aujourd'hui sans aucune figure. C'est le premier
contact visuel du lecteur avec l'ouvrage, donc une planche au moins se
justifie, mais elle doit être immédiatement lisible sans culture préalable.
Trois idées, par ordre de priorité.

## 1. L'épargne écrase le rendement (courbes années-avant-l'indépendance)

**Ce que ça montre.** Quatre courbes du nombre d'années de travail avant
l'indépendance en fonction du taux d'épargne (de 5 % à 80 % en abscisse), une
courbe par hypothèse de rendement réel (3 %, 4 %, 5 %, 6 %). Les quatre
courbes sont serrées les unes contre les autres alors que chacune plonge de
50 ans à 5 ans le long de l'axe du taux d'épargne. Deux repères annotés
horizontalement : le passage de 20 % à 35 % d'épargne (12 ans gagnés) et le
passage de 5 % à 7 % de rendement à taux d'épargne constant (2 à 3 ans).

**Pourquoi elle vaut sa place.** C'est la thèse centrale de la section
« L'arithmétique qui rend le FIRE possible » et de l'encadré « Le rendement ne
vous sauvera pas », aujourd'hui portée par un tableau à une seule colonne de
rendement et par deux chiffres en prose. L'écart d'échelle entre les deux
leviers ne se voit pas dans un tableau ; il saute aux yeux dans un faisceau de
courbes quasi confondues. Elle remplacerait avantageusement le tableau MMM, ou
le prolongerait.

**Coût : B.** Pas de données, une formule fermée (annuité + capitalisation,
cible = 25 fois les dépenses). Quelques dizaines de lignes.

## 2. La grille de la cible : dépenses annuelles × taux de retrait

**Ce que ça montre.** Une petite grille de capital cible, lignes = dépenses
annuelles (20 000 à 80 000 € par pas de 10 000), colonnes = taux de retrait
(4 %, 3,5 %, 3 %), cellules = capital nécessaire, avec un dégradé de fond
monotone. La ligne de lecture du livre (3 à 3,5 % pour un horizon long) est
encadrée, la colonne 4 % restant visible comme point de départ historique.

**Pourquoi elle vaut sa place.** La section « Les trois nombres qui gouvernent
tout » énonce le pont entre dépenses et capital sans jamais le rendre tangible,
et le lecteur d'un premier chapitre veut immédiatement situer son propre
chiffre. La grille montre aussi, gratuitement, ce que coûte le passage de 4 % à
3 % (un tiers de capital en plus), qui est un des messages de tout le livre.

**Coût : C** plutôt que figure. Une division suffit ; un tableau markdown
soigné, ou un encadré, fait le travail aussi bien qu'une planche SVG et se lit
mieux sur liseuse. À trancher selon la place déjà prise par les deux tableaux
existants de l'article.

## 3. Frise 1992-2026 : le taux « sûr » qui recule

**Ce que ça montre.** Une frise horizontale des cinq jalons de la section
« D'où ça vient » (Robin & Dominguez 1992, Bengen 1994, Trinity 1998, Mr. Money
Mustache 2011, Early Retirement Now 2016), et sous la frise, en escalier, le
taux de retrait de référence associé à chaque époque : rien avant 1994, 4 %
de 1994 à la fin des années 2000, puis la descente vers la fourchette 3 à
3,5 % des travaux récents pour un horizon FIRE.

**Pourquoi elle vaut sa place.** La généalogie est aujourd'hui une liste de
paragraphes en gras, et l'encadré « Où en est le mouvement aujourd'hui » affirme
que le sujet a changé de nature sans le montrer. L'escalier donne à voir le
seul chiffre qui intéresse vraiment le lecteur au moment où il découvre le
domaine, et il installe dès la première page la position du livre par rapport
à la règle des 4 %.

**Coût : B, avec une réserve.** Les dates et les deux taux extrêmes sont dans
l'article et vérifiés (Bengen 1994, SAFEMAX 4,15 % ; Trinity 1998). En
revanche l'escalier intermédiaire est un jugement éditorial, pas une série
mesurée : il faut ou bien l'assumer comme tel dans la légende, ou bien s'en
tenir à deux marches seulement (4 % puis 3 à 3,5 %) pour ne rien inventer.

## Écartées

- Une carte 2D des variantes (Lean/classique/Fat/Barista/Coast) : le tableau
  existant est déjà la bonne forme, une figure n'ajouterait rien.
- La courbe taux sûr contre horizon (« au-delà de 40 ans, la courbe
  s'aplatit ») : déjà traitée par la figure `horizon-flatten` de
  [[horizon-et-esperance-de-vie]], à ne pas dupliquer ici.
- L'effet amortisseur de Barista/Coast sur le risque de séquence : thèse forte
  et aujourd'hui purement verbale, mais la planche (ruine en fonction du nombre
  d'années de revenu partiel initial) appartient à
  [[revenus-complementaires]], pas à l'article d'ouverture.


## la-regle-des-4-pourcents

Article aujourd'hui sans aucune figure, alors que c'est la porte d'entrée du
livre sur le sujet. Trois idées, par ordre de valeur.

## 2. Ce que le retrait rigide fait vraiment, sur le millésime 1966

Deux tracés superposés sur le même axe temporel, un millésime unique (1966, ou
2000 pour un cas moderne) : en bas, le retrait annuel en euros constants, une
droite parfaitement plate ; en haut, ce même retrait exprimé en pourcentage du
capital restant, qui grimpe de 4 % à 8 ou 10 % au creux avant de redescendre.
La zone entre les deux courbes est le risque, et il est invisible pour le
retraité qui ne regarde que son virement mensuel.

Pourquoi elle vaut sa place : elle illustre les deux premières « propriétés »
de la section « La mécanique », qui sont le cœur du chapitre et n'ont
aujourd'hui qu'un exemple arithmétique (700 000 € et 5,9 %). Elle montre en une
image pourquoi la règle est confortable à vivre et dangereuse à tenir, et elle
prépare [[sequence-des-rendements]].

Coût : **A**. `pkg/replay` embarque déjà le 60/40 américain réel depuis 1954,
donc le millésime 1966 est directement rejouable ; les deux séries sortent du
même rejeu.

## 3. Le multiple de 25, en table de conversion

Pas une figure, un petit tableau dans l'encart « La règle inversée » : quatre
niveaux de dépenses mensuelles (500, 1 000, 2 000, 3 000 €) en lignes, trois
taux (4 %, 3,5 %, 3 %) en colonnes, et le capital nécessaire dans les cellules.
Le lecteur y trouve directement son ordre de grandeur, et voit que passer de
4 à 3 % ajoute un tiers de capital.

Pourquoi il vaut sa place : l'article affirme deux fois que le multiple de 25
est « le meilleur réflexe mental du sujet » sans jamais le donner à manipuler.
Une table de conversion est ce qu'on relit et ce qu'on retient.

Coût : **C**. Aucune donnée, c'est une division ; le tableau fait mieux qu'un
graphique ici.

## Écartées

- Une cascade des étages du 4 % : elle appartient à
  [[les-maths-du-4-pourcent]], la dupliquer ici affaiblirait les deux pages.
- Une grille de taux de succès à la Trinity : c'est la figure naturelle de
  [[etude-trinity]], pas de ce chapitre d'introduction.


## combien-il-vous-faut

Article aujourd'hui sans aucune figure. Sa thèse tient en une phrase : l'erreur
n'est pas dans la multiplication, elle est dans ses deux termes. Trois passages
purement verbaux méritent une image ; un quatrième mérite un tableau.

## 1. Le levier réel : ce que chaque paramètre coûte en années de travail

Barres horizontales classées par amplitude, une par paramètre, chacune montrant
l'écart de capital cible entre sa borne basse et sa borne haute, sur DEUX axes
lus en parallèle (euros en bas, années de travail équivalentes en haut, à taux
d'épargne fixé). Paramètres : ± 10 % sur les dépenses, taux 3 % / 3,5 % / 4 %,
friction fiscale 8 % / 20 %, pension comptée ou non, pension décalée de deux
ans. Le classement est le message : sur le cas Nadia et Marc, l'oubli de la
pension et l'erreur de dépenses pèsent plus lourd que le demi-point de taux sur
lequel se concentrent toutes les discussions. C'est exactement la conclusion de
l'étape 5 (« stressez-le »), qui n'est aujourd'hui qu'une liste de trois
secousses sans hiérarchie. Conversion en années de travail à annoncer comme une
hypothèse affichée sur la planche (taux d'épargne, rendement réel).

Coût : **B**. Arithmétique simple à écrire (cible = dépenses/taux, puis temps
d'accumulation), aucune donnée externe. La variante ruine (5 % contre 12 %)
demanderait `pkg/decumul` et n'est pas nécessaire ici.

## 4. Net vers brut, par enveloppe (tableau, pas figure)

Un petit tableau à quatre lignes (CTO, assurance-vie de plus de 8 ans, PEA de
plus de 5 ans, PUMa) avec, pour chacune, le taux applicable, l'assiette réelle
(la seule part de gains, pas le retrait) et la friction effective sur un retrait
type contenant 40 % de plus-values latentes. Le paragraphe de l'étape 2 empile
aujourd'hui quatre régimes en prose ; le lecteur ne peut pas voir d'où sort la
fourchette 8-20 %. Un graphique n'apporterait rien, les quatre régimes ne se
comparent pas sur un axe commun.

Coût : **C**. Tableau ; taux à relire dans les données fiscales à chaque mise à
jour, la ligne PUMa reste un ordre de grandeur.


## les-trois-phases

Article aujourd'hui sans aucune figure, très dense en listes à puces. Sa thèse
centrale (le bloc `cle`) est purement verbale et mérite une planche.

## 2. Le même krach de −40 %, à trois moments du plan (coût A)

Trois petites multiples partageant l'axe des ordonnées, un même plan unique
(mêmes versements, même cible, même retrait), avec un krach identique
déclenché à l'année 10 de l'accumulation, à l'année du départ, et à l'année 25
du retrait. Sous chaque panneau, un seul chiffre de verdict : date
d'indépendance avancée de X mois (le krach précoce est une aubaine), retardée
de Y années, effet quasi nul.

Pourquoi elle vaut la place : c'est la thèse « les réflexes d'une phase sont
toxiques dans une autre » rendue en une image, et l'ancre du livre selon
laquelle un krach à l'année 25 d'une retraite de 30 ans est presque indolore.
Elle ne double pas la figure `sequence-risk` de [[sequence-des-rendements]],
qui compare deux retraités entre eux : ici on compare les trois phases d'une
même vie, accumulation comprise, et le panneau de gauche (krach = bonne
nouvelle) n'existe nulle part ailleurs dans le livre.

Coût A : `pkg/replay` fournit le réel US 60/40 depuis 1954 et `pkg/datasets`
les séries longues ; un vrai krach historique (1973-74 ou 2000-2002) rejoué à
trois positions du plan est plus crédible qu'un −40 % synthétique.

## 3. Le reste : rien à ajouter (coût C)

Le bloc `astuce` (les deux compteurs), le bloc `attention` (le piège du
sommet) et l'exemple Inès sont déjà des encadrés efficaces et courts. Un
tableau « trois phases × ce qu'on surveille » ne ferait que redire la section
« L'essentiel à retenir », qui joue déjà ce rôle. Deux planches suffisent pour
un article de colonne vertébrale.


## utiliser-la-page-fire

Article aujourd'hui sans aucune figure, et intégralement verbal. C'est un mode
d'emploi, donc la moitié des passages n'a pas vocation à être illustrée. Trois
passages portent en revanche une thèse qui gagnerait beaucoup à être vue.

## 1. L'intervalle des six colonnes, sur un plan unique (coût B)

Une seule planche horizontale : un plan de référence (le cas d'école du
simulateur, 600 000 € / 24 000 € / 42 ans) évalué par les six modèles, un
segment ou une pastille par modèle, rangés de l'optimiste au catastrophique,
avec la ruine acceptable en trait vertical. On lit d'un coup ce que l'article
répète en mots, à savoir que le même plan vaut 2 % dans une colonne et 15 %
dans une autre, et que la décision se prend sur l'intervalle, pas sur un
chiffre. C'est la thèse centrale de l'encart d'ouverture et de la règle
« planifiez entre le central et le broad-sample », aujourd'hui purement
affirmée. Coût B : les six sources existent (`pkg/decumul` + le panel
broad-sample embarqué), il faut écrire le petit calcul qui les fait tourner
sur un plan figé et geler les chiffres.

## 2. Le bruit d'échantillonnage, ou pourquoi la deuxième décimale ment (coût B)

Le même plan recalculé une vingtaine de fois à 1 000, 2 000 et 8 000
trajectoires, en trois colonnes de points dispersés autour de la vraie valeur,
avec l'amplitude annotée (±0,7 point à 2 000, moitié moins à 8 000). La forme
est délibérément pauvre, un nuage vertical par réglage, parce que la
démonstration est justement que le nuage a une largeur. Elle sert le mésusage
n° 3 de l'article (« confondre les décimales avec du signal ») et le nouveau
paragraphe du groupe Simulation, tous deux verbaux. Coût B : boucle triviale
sur `Simulate`, aucune donnée nouvelle.

## 3. Le mélange vers le prior, en fonction de l'écart horizon / historique (coût B)

Une courbe unique : en abscisse le rapport horizon sur historique disponible,
en ordonnée le poids donné au prior mondial, montant de 0 vers le plafond de
50 %, avec deux ou trois cas repérés (fonds de 20 ans pour un plan de 45 ans,
etc.) et le μ résultant en second axe ou en étiquettes de bout de courbe. Le
paragraphe « pré-remplies puis mélangées à proportion de ce que l'horizon
excède l'historique » est le plus abstrait de l'article et le plus contesté par
les lecteurs qui veulent pousser leur μ. Voir la règle rend le garde-fou
compréhensible plutôt que subi. Coût B : la règle est une formule, le tracé ne
demande aucune donnée.

## 4. La carte de lecture de la page : pas une figure (coût C)

L'idée d'un plan annoté de la page avec le parcours numéroté est tentante, mais
elle vieillirait à la première évolution de l'interface et doublonnerait
l'encart « la séance type, en six gestes », qui fait déjà le travail. Si l'on
veut renforcer ce passage, un petit tableau à trois colonnes (geste, section,
question à laquelle on répond) est plus robuste et plus consultable qu'une
capture ou un schéma.


## erreurs-classiques-fire

L'article est entièrement verbal (dix sections + une check-list) et ne porte
aucune figure. Sa thèse la plus forte est aussi la moins démontrée sur la page :
« dans l'ordre approximatif du coût ». Une seule figure peut la prouver.

## 2. Où la pension déplace le risque (deux profils d'échec dans le temps)

Pour la seule erreur n° 4, le taux de défaillance annuel du même plan avec et
sans pension, année 1 à année 50. Sans pension, la courbe monte et ne redescend
jamais ; avec pension, elle s'écrase après l'année 20 et le risque se concentre
tout entier sur le pont. La figure porte deux affirmations du texte à la fois :
le facteur 2 à 4 sur la ruine, et le fait que la pension arrive exactement là où
le portefeuille fatigue. Elle complète `horizon-flatten` (qui traite l'horizon)
sans la répéter, car ici l'axe est le calendrier du risque, pas sa durée.

Coût : **B** (même balayage que l'idée 1, sortie « hazard par année » à
extraire ; à ne réaliser que si l'idée 1 n'est pas retenue, sinon redondance
partielle sur l'erreur n° 4).

## 3. Le même plan vu par quatre modèles (pas une figure : un tableau court)

Pour l'erreur n° 8, quatre lignes (fenêtres historiques US, broad sample
mondial, Monte Carlo paramétrique, bootstrap par blocs) et une colonne de ruine
pour un plan identique. L'écart entre la première et la deuxième ligne fait tout
le travail rhétorique. Quatre nombres n'ont pas besoin d'un graphique, et un
tableau se lit mieux en EPUB.

Coût : **C** (tableau ; les quatre sources existent dans `pkg/scenario` et
`pkg/datasets/broadsample`, mais la place d'une vraie figure serait mal
employée ici, et le sujet appartient d'abord à `pieges-des-simulateurs`).
