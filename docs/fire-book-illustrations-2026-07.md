# Firebook: figure backlog for « Les stratégies de retrait »

Status: ACTIVE backlog, opened 2026-07-29 after a line-by-line review of the
twelve articles of the withdrawal-strategies part. The shortlist of five
(2.1, 6.1, 10.1, 12.1, 7.1) SHIPPED on 2026-07-29 as `figures_strategies.go`;
everything else is still an idea. Delete this file once the remaining retained
figures are shipped, or once they are dropped (docs/ stays curated).

The review itself is done and shipped; this is only the illustration half of it.
The figures follow the v2 plate system (`figures_v2.go`) and the data-backed
pattern recorded in `fire-book-design.md` (frozen literal arrays in the plate, a
guard test recomputing them from the engine).

Written in French because the book is; the surrounding docs are English.

Pour chaque idée : ce qu'elle montre, pourquoi elle vaut la place qu'elle prend,
et son coût.

- **A** = les données existent déjà (`pkg/replay`, `pkg/datasets`, ou l'article
  lui-même donne les chiffres). Reste la plate à dessiner.
- **B** = un calcul à écrire d'abord (backtest, balayage, nouvelle règle dans
  `pkg/replay`), puis la plate.
- **C** = pas de figure. Un tableau ou un encadré fait mieux le travail.

Un principe que je me suis appliqué en les cherchant : ne pas refaire six fois
la même courbe. Chaque article a une thèse différente, donc mérite une forme
différente. J'ai écarté plusieurs idées qui n'étaient que « la même courbe de
revenu, encore une fois ».

---

## 1. panorama-strategies-retrait
*Existant : `withdrawal-frontier`.*

**1.1 La carte des cinq familles, par information écoutée (A)**
La frontière montre les résultats ; il manque les entrées. Cinq bandes
empilées, de « n'écoute rien » à « externalise », chaque bande portant ses
règles en pastilles. À droite, deux jauges qui montent ensemble : place sur la
frontière, gouvernance exigée. C'est la thèse du paragraphe « plus une règle
écoute d'information, mieux elle se place, plus elle exige de gouvernance »,
aujourd'hui affirmée sans être montrée.

**1.2 Les quatre règles du cas travaillé (A)**
L'exemple chiffré est en liste à puces. En figure : quatre lignes, deux barres
opposées (ruine à gauche, consommation totale à droite), plus un point pour le
pire quartile. On voit d'un coup que ruine et consommation ne se classent pas
dans le même ordre, ce qui est toute la leçon.

**1.3 La grille des six critères (C)**
Six critères en lignes, cinq familles en colonnes, ●/◐/○. Remplace une prose
qui énumère. Se lit en dix secondes.

---

## 2. retrait-fixe-bengen
*Aucune figure aujourd'hui. C'est l'article le plus sous-illustré de la partie.*

**2.1 La falaise silencieuse (A, prioritaire). LIVRÉE : `bengen-falaise`**
LA figure manquante du livre. Sur 1973 (ou 1966), deux courbes sur le même
axe temps : le portefeuille réel qui descend, et le taux de retrait courant qui
monte, 4 → 6 → 8 → 12 %. Une bande marque la zone 8-10 %, et une verticale le
« point de non-retour », des années avant le zéro. L'article dit « prévisible
des années à l'avance » ; ici on le voit. Et le lecteur repart avec le voyant
qu'il doit surveiller chaque janvier.

**2.2 Le luxe non consommé (A)**
Distribution du patrimoine terminal du fixe prudent, avec le capital de départ
en repère. Le message « legs médian énorme » est aujourd'hui une affirmation.
Sur les trois millésimes de `pkg/replay`, 1985 donne 3,83 M€ laissés pour un
départ à 600 k€ : le chiffre est déjà là.

**2.3 Les trois amendements, en cascade (B)**
Waterfall : SAFEMAX de base, puis +0,25-0,5 pt (gel d'indexation), puis l'effet
du cliquet, puis l'effet du taux conditionné au CAPE. Demande un balayage.
Même forme que `cascade-4pct`, donc cohérent avec le livre.

---

## 3. pourcentage-fixe

**3.1 Le revenu EST le portefeuille (A)**
Deux courbes superposées, portefeuille et revenu, littéralement la même forme à
un facteur près. Trente secondes à dessiner, et ça rend inutile un paragraphe
d'explication. À mettre juste avant le tableau des épisodes.

**3.2 Brut contre Yale contre corridor, même krach (B)**
L'encadré `::: exemple` donne déjà les trois trajectoires en chiffres
(56 → 34 k€ contre 56 → 48,6 k€). En figure, la chute libre contre la pente
douce se lit d'un coup. Demande d'implémenter le lissage de Yale (70/30), qui
n'existe pas dans `pkg/replay` : une quinzaine de lignes.

**3.3 La borne géométrique (B)**
Revenu réel médian sur 30 ans pour w = 3, 4, 5, 6 %, avec le rendement
géométrique en repère. On voit le point de bascule entre revenu stable et
revenu qui s'érode. C'est l'encadré `::: science` rendu visible.

---

## 4. guyton-klinger
*Existant : `gk-cascade-1966`, qui couvre déjà bien la pathologie.*

**4.1 Le corridor vu du pilote (A)**
Ce que le ménage regarde vraiment chaque janvier : le taux de retrait courant
dans son couloir (3,44 % / 5,16 %), avec les franchissements marqués et la
coupe qui en découle. La figure existante montre les conséquences ; celle-ci
montre l'instrument. Complémentaire, pas redondante.

**4.2 Le taux initial était le coupable (B)**
Le vrai apport possible. Même millésime, quatre taux initiaux (4,0 / 4,3 / 5,0
/ 5,5 %), et pour chacun : nombre de coupes, revenu au creux, années sous le
plan. Un petit tableau-graphique 4 lignes. L'encadré `::: attention` l'affirme,
personne ne l'a jamais montré au lecteur français.

---

## 5. vpw
*Existant : `vpw-table` et `vpw-pont`. Bien servi, je ne vois rien d'urgent.*

**5.1 Le test de tolérance à la perte (B, léger)**
La doctrine l'impose et « tout le monde le saute », dit l'article. Une figure
le rendrait incontournable : revenu VPW normal contre revenu sous « actions
−50 % », pour trois allocations (40/60, 60/40, 80/20), avec la ligne du
plancher. On voit immédiatement quelle allocation est admissible. Arithmétique
pure, aucun backtest.

---

## 6. regles-cape

**6.1 La double contracyclicité (A, prioritaire). LIVRÉE : `cape-contracyclique`**
La thèse de l'article, aujourd'hui purement verbale malgré des chiffres tout
prêts. Trois barres alignées sur le même krach : portefeuille −30 %, taux w
+21 %, revenu −16 %. Le lecteur voit les deux facteurs qui se compensent. C'est
la figure la plus rentable de la partie après la falaise de Bengen.

**6.2 Le taux que la règle aurait servi depuis 1881 (B)**
La courbe a + b/CAPE appliquée à l'historique CAPE (déjà embarqué dans
`pkg/datasets/cape`), avec les zones repères. Le lecteur voit que la règle
aurait dit 6 % en 1982 et 3,2 % en 2000, et que ce n'est pas un réglage
arbitraire.

---

## 7. guardrails-morningstar
*Existant : `guardrails-capteur` (les sept revues).*

**7.1 Deux thermomètres sur le même plan (B, prioritaire). LIVRÉE : `deux-thermometres`**
Le cœur de l'article. Le retraité de 62 ans dont la pension arrive à 64 : le
taux courant s'affole et déclenche une coupe, la probabilité de succès ne
bouge pas. Deux courbes, deux corridors, une décision opposée. L'article
raconte cet exemple en prose ; c'est exactement ce qui se dessine.

**7.2 La série Morningstar 3,3 → 3,8 → 4,0 → 3,7 (C)**
Quatre valeurs et leur contexte d'année. Un mini-tableau suffit, et il porte la
leçon d'épistémologie mieux qu'une courbe à quatre points.

---

## 8. amortissement-abw
*Existant : `abw-1966`.*

**8.1 La richesse totale, en une pile (A)**
L'idée neuve de l'article (W = portefeuille + pensions actualisées − legs) est
un empilement, donc se dessine comme tel : une barre en trois segments, à côté
de la barre « portefeuille visible seul ». Le lecteur comprend d'un coup
pourquoi le retrait est plus élevé avant la pension et pourquoi l'allocation se
raisonne sur le total. Sert aussi `revenus-complementaires`.

**8.2 Le legs comme paramètre (A)**
Trois trajectoires de capital pour trois legs visés (0, 200 k€, 500 k€), même
règle. Montre que le legs est un choix chiffré et pas un résidu, ce que
l'article affirme deux fois.

---

## 9. plancher-plafond
*Existant : `corridor-1966` et `corridor-borne`. Le mieux servi de la partie.*
Rien à ajouter. La table des six années fait déjà le travail au fil du texte.

---

## 10. rentes-et-annuites
*Aucune figure.*

**10.1 Les crédits de mortalité par âge (A, prioritaire). LIVRÉE : `credits-mortalite`**
Le concept fondateur de l'article, et il est purement verbal aujourd'hui.
Une courbe : crédit de mortalité en % par an, de 60 à 95 ans, avec la zone
75-82 marquée comme fenêtre d'achat. On voit pourquoi la rente est « un mauvais
produit de sexagénaire et un excellent produit d'octogénaire ».

**10.2 Ce que couvre chaque étage (A)**
Un empilement plancher / confort avec, en face, qui le finance (pension, rente,
portefeuille). C'est la doctrine safety-first en une image, et elle est
réutilisable telle quelle dans `choisir-sa-strategie`.

**10.3 La ruine qui monte alors que tout va mieux (B)**
Le piège que l'article annonce. Deux panneaux : à gauche la ruine du besoin
total qui augmente après annuitisation, à droite la ruine pondérée par la
mortalité qui s'effondre. La démonstration que le chiffre unique ment.

---

## 11. sept-facons-de-vivre
*Existant : six plates.* Rien à ajouter, c'est le plus illustré du livre.

---

## 12. choisir-sa-strategie

**12.1 L'arbre de décision en cinq étapes (A). LIVRÉE : `arbre-decision`**
L'article EST une procédure, et une procédure se dessine. Plancher couvert ?
→ tests d'admissibilité → tempérament → phases. Les feuilles pointent vers les
articles. C'est la figure de synthèse de toute la partie, et probablement la
plus utile du livre entier pour un lecteur qui arrive à la décision.

**12.2 Où porte vraiment votre attention (A)**
L'encadré `::: science` donne déjà tous les chiffres : erreur de dépenses 3-6
pts, pension oubliée 3-8 pts, taux initial 5-10 pts, choix de la règle 1-3 pts.
Quatre barres classées, et la hiérarchie de l'attention devient évidente. Un
antidote visuel à l'optimisation prématurée.

---

## Les cinq retenues, livrées le 2026-07-29

1. **2.1** la falaise silencieuse (Bengen n'a aucune figure, et c'est l'article
   le plus lu de la partie)
2. **6.1** la double contracyclicité (thèse verbale, chiffres déjà écrits)
3. **10.1** les crédits de mortalité par âge (concept fondateur, zéro figure)
4. **12.1** l'arbre de décision (la synthèse que le lecteur vient chercher)
5. **7.1** les deux thermomètres (le cœur de l'état de l'art)

Livrées dans `pkg/firebook/figures_strategies.go`, avec `figures_strategies_test.go`
en garde (le millésime 1966 est recalculé depuis `pkg/replay`, la formule CAPE et
la loi de mortalité sont vérifiées contre les chiffres du texte).

La forme retenue diffère à chaque fois, ce qui était le but : un voyant qui monte
pendant dix-neuf ans, un plan de revenus constants traversé de deux façons, deux
courbes qui échangent leur rang à un âge nommable, une procédure qui se referme,
quatre bandes d'instrument lues côte à côte.
