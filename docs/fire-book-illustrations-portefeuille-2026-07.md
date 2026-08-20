# Firebook: figure backlog for « Le portefeuille de retrait »

Status: RESIDUAL backlog, opened 2026-07-30 during the line-by-line review of
the fourteen articles of the portfolio part (one reviewer per article; the
prose fixes shipped as the "line review of ..." commit series). Thirteen ideas
were built over 2026-07-30 and 2026-07-31, eight more during the 2026-08 plate
campaign, and their entries have been removed from this file (last purge
2026-08-21), so what remains is exactly what has NOT been built.

What is left is a pool of ideas to RE-EVALUATE when a figure is wanted, not a
to-do list: several were written before their article had any illustration, and
some are weak or superseded. The per-article context notes date from the review
and predate the plates built since, so read the article itself before acting on
an idea. Delete this file once the pool stops being useful.

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

---

## concevoir-un-portefeuille

Article de méthode, aujourd'hui sans aucune figure. Il n'a pas besoin d'une
figure de données pour exister, mais deux passages purement verbaux portent la
thèse et gagneraient beaucoup à être vus : la carte risques -> briques
(question 5) et l'arbitrage pires chemins / médiane (question 6). Trois idées,
classées par valeur.

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

## actifs-defensifs

L'article est le hub de la partie portefeuille (il porte la table des défenses et les
verdicts par candidat) et n'a aucune figure aujourd'hui. Ses articles-fils en ont
(`correl-sign`, `trend-smile`, `trend-annees`, les quatre d'`allocation-actions-obligations`).
Une figure ici a donc une vraie valeur de carte.

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

## obligations-indexees

L'article n'a aucune figure aujourd'hui, alors qu'il porte deux passages
purement verbaux très visuels : le résultat de l'échelle (bloc `science`) et la
décomposition de 2022. Trois idées, par ordre de valeur.

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
