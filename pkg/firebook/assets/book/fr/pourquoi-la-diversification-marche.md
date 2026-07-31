# Pourquoi la diversification fonctionne : la mécanique du seul repas gratuit

« Ne mettez pas tous vos œufs dans le même panier » est le conseil financier le plus ancien du monde, et le plus mal compris. La plupart des épargnants diversifient comme on récite une prière, sans voir que derrière le proverbe se cache un **mécanisme mathématique précis**, le seul de toute la finance qui donne quelque chose sans rien demander en échange (Markowitz l'appelait « le seul repas gratuit »). Ce mécanisme a des conditions, des limites, et surtout une conséquence méconnue : un portefeuille diversifié et rééquilibré peut rendre **plus** que la moyenne de ses composants, un supplément qui porte le joli nom de rebalancing premium.

Cet article démonte le mécanisme pièce par pièce, avec l'arithmétique minimale (trois formules, aucune au-delà du lycée), chiffre le supplément, montre pourquoi l'effet est encore plus fort en décumulation qu'en accumulation, et finit par les limites, car la diversification a aussi son folklore. Après lecture, vous saurez dire exactement ce que chaque brique de votre portefeuille apporte au collectif, et repérer la fausse diversification qui encombre tant de portefeuilles réels.

::: cle L'idée en une phrase
Quand deux actifs ne chutent pas en même temps, leur mélange a une volatilité **inférieure à la moyenne** de leurs volatilités, alors que son rendement moyen reste **exactement la moyenne** de leurs rendements. Le risque se dilue, le rendement non : voilà le repas gratuit. Et comme la volatilité coûte du rendement composé ([[rendements-arithmetiques-geometriques]]), moins de volatilité à rendement moyen égal signifie plus de rendement **final**, pas seulement moins d'émotions.
:::

## L'arithmétique du panier

Prenez deux actifs de même rendement attendu (disons 5 %) et de même volatilité (disons 20 %). Le portefeuille 50/50 a un rendement attendu de 5 %, toujours, quelle que soit leur relation. Sa volatilité, elle, dépend entièrement de la **corrélation** ρ entre les deux → ρ = 1 (ils bougent ensemble) donne 20 %, rien n'a changé ; ρ = 0,5 donne 17,3 % ; ρ = 0 (indépendants) donne 14,1 % ; ρ = −0,5 donne 10 %. La formule générale (la volatilité du panier vaut la racine de w²σ² + w'²σ'² + 2ww'σσ'ρ) dit tout : le seul levier est ρ, et chaque cran de corrélation en moins retire du risque **gratuitement**.

D'où la première conclusion pratique, qui vaut audit : la diversification ne se compte pas en nombre de lignes, elle se compte en **corrélations basses**. Trente fonds actions sont UN actif (corrélations 0,85-0,95 entre eux, la trentaine de noms ne dilue à peu près rien). À l'inverse, quatre briques bien choisies (actions mondiales, obligations longues, or, trend) affichent des corrélations croisées entre −0,2 et +0,3, et font plus de travail que les trente fonds ([[actifs-defensifs]], [[portefeuilles-tous-temps]]). Le bénéfice marginal s'effondre vite : passer de 1 à 4 briques décorrélées transforme le portefeuille, passer de 8 à 20 ne change presque rien et multiplie les frais et les erreurs de suivi.

## Le supplément caché : le rebalancing premium

La dilution du risque n'est que la moitié de l'histoire. L'autre moitié demande de relier deux faits vus séparément ailleurs. Un, le rendement qui compose votre capital est le rendement **géométrique**, et il vaut environ le rendement moyen moins la moitié de la variance (le volatility drag). Deux, la diversification réduit la variance à rendement moyen constant. Donc elle **augmente le rendement géométrique**. Ce supplément a été nommé « diversification return » par Booth et Fama (1992) : un portefeuille rééquilibré rend plus que la moyenne pondérée des rendements géométriques de ses composants, l'écart valant grosso modo la moitié de la variance économisée.

L'ordre de grandeur honnête : 0,2 à 0,5 point par an pour un portefeuille classique, davantage quand on mélange des briques à la fois **volatiles et décorrélées**. L'exemple canonique est l'or : zéro rendement réel propre ([[primes-de-risque]]), 15-20 % de volatilité, corrélation ~0 aux actions → dans un panier rééquilibré, cette ligne « stérile » fabrique du rendement de panier, ce qui résout le paradoxe apparent des portefeuilles tous-temps où une brique sans espérance améliore le total ([[or-en-retrait]]).

Le **rééquilibrage** est la pompe qui récolte ce supplément : vendre ce qui a monté, racheter ce qui a baissé, mécaniquement, aux bandes ou au calendrier ([[revue-annuelle]]). Deux mises en garde d'honnêteté. D'abord la pompe ne crée pas le supplément, elle l'encaisse : l'essentiel du gain vient de la variance évitée, et un portefeuille non rééquilibré finit par la reperdre en laissant une ligne dominer. Ensuite le rééquilibrage a un ennemi, la **tendance** : dans un marché qui monte (ou baisse) en ligne droite pendant des années, vendre le gagnant coûte ; le supplément se récolte sur les allers-retours, pas sur les lignes droites. Sur données réelles, mélange des deux, la version disciplinée gagne modestement mais sûrement, et surtout elle maintient le profil de risque que le plan a promis, ce qui en décumulation est sa vraie mission.

::: exemple Le démon de Shannon, version rentier
L'illustration limite, due à Claude Shannon : un actif qui fait pile ou face, +100 % ou −50 % chaque année. Son rendement géométrique est **nul** (un aller-retour laisse le capital inchangé), détenu seul il ne construit rien. Mélangé à 50/50 avec du cash à 0 % et rééquilibré chaque année, le panier gagne en moyenne géométrique environ **6 % par an**, fabriqués à partir de deux ingrédients dont aucun ne rapporte rien. Aucun actif réel n'est aussi caricatural, mais la leçon est exacte : la volatilité décorrélée, capturée par un rééquilibrage discipliné, est une matière première de rendement. C'est la version formelle de « acheter bas, vendre haut », exécutée par une règle plutôt que par un talent.
:::

## Pourquoi l'effet est doublé en décumulation

En accumulation, la diversification améliore le confort et un peu le rendement final. En décumulation, elle joue sur un levier bien plus puissant : le **risque de séquence** ([[sequence-des-rendements]]). Le taux de retrait soutenable n'est pas fixé par la trajectoire moyenne mais par les **pires** trajectoires, et la diversification agit précisément là : elle raccourcit la queue gauche, réduit la profondeur et la durée des drawdowns réels, donc relève le plancher qui dimensionne tout le plan. C'est visible dans n'importe quel simulateur : en passant d'un 100 % actions à un panier quatre briques, la richesse médiane à trente ans **baisse** souvent un peu, tandis que le SWR à 95 % de succès **monte** ([[lire-un-fan-chart]] pour lire les deux à la fois). Le rentier ne diversifie pas pour la moyenne, il diversifie pour le percentile 5, là où vivent ses nuits blanches et sa probabilité de ruine.

Le même raisonnement éclaire un point de vocabulaire qui trompe beaucoup : diversifier **dans le temps**. Étaler ses retraits sur trente ans expose chaque euro à des marchés différents, et les règles de retrait flexibles ([[choisir-sa-strategie]]) sont, au fond, une manière de diversifier la consommation entre les années fastes et les années maigres. Portefeuille et règle de retrait travaillent le même risque par deux bouts.

## Les limites, sans folklore

**Les corrélations sont des passagères de beau temps.** Dans une panique de liquidité, presque tout tombe ensemble pendant quelques semaines (2008, mars 2020) : la corrélation actions-immobilier-crédit-hedge funds monte vers 1 exactement quand on comptait sur elle. Ce qui survit à ce test est court : la duration d'État de qualité (dans les crises **désinflationnistes** seulement, 2022 l'a rappelé, [[regimes-de-marche]]), le cash, parfois l'or, le trend si la crise dure ([[managed-futures]]). La diversification ne supprime pas les chocs courts, elle différencie les **régimes** : c'est déjà énorme, ce n'est pas l'immunité.

**La fausse diversification, ou diworsification.** Ajouter des lignes corrélées à ce qu'on détient déjà (un fonds monde + un fonds US + un fonds tech + dix titres américains) donne le sentiment du panier avec le risque du monolithe. Le test en une question par ligne : « dans quel régime cette position gagne-t-elle quand le reste perd ? ». Pas de réponse → pas de diversification, juste des frais ([[primes-de-risque]] pour la version complète de l'audit).

**Le coût psychologique, enfin.** Un portefeuille vraiment diversifié contient **en permanence** une ligne qui déçoit, c'est sa signature (si tout monte ensemble, tout baissera ensemble). La tracking error vis-à-vis du voisin 100 % actions culmine dans les grandes hausses, et c'est là que les paniers se font démonter par leurs propriétaires, juste avant d'être utiles. La défense est la même que partout : la thèse de chaque brique écrite au plan, et le jugement porté sur le panier, jamais sur une ligne ([[psychologie-du-retrait]], [[construire-son-plan]]).

## L'essentiel à retenir

- Le mélange d'actifs décorrélés réduit la volatilité sans réduire le rendement moyen : c'est un théorème, pas une opinion, et le seul levier est la corrélation, pas le nombre de lignes (trente fonds actions = un actif).
- Moins de variance à rendement moyen égal = plus de rendement géométrique : le rebalancing premium (0,2-0,5 point/an, davantage avec des briques volatiles et décorrélées comme l'or) est la moitié oubliée du repas gratuit, récoltée par un rééquilibrage discipliné.
- En décumulation l'effet est démultiplié : la diversification travaille la queue gauche et le risque de séquence, donc le SWR, même quand elle abaisse un peu la médiane ; on diversifie pour le percentile 5.
- Limites honnêtes : les corrélations montent vers 1 dans les paniques courtes (seuls duration, cash, or et trend selon le régime y survivent), et la diworsification (des lignes corrélées empilées) imite le panier sans en avoir le mécanisme.
- Le portefeuille diversifié contient toujours une ligne décevante, par construction ; qui ne l'accepte pas par écrit finira par vendre la brique la veille du jour où elle servait.

---

## Pour aller plus loin

- Harry Markowitz, « Portfolio Selection » (1952) : l'article fondateur, étonnamment lisible.
- Booth & Fama, « Diversification Returns and Asset Contributions » (1992) : la formalisation du supplément de diversification.
- William Bernstein, « The Rebalancing Bonus » (Efficient Frontier) : le chiffrage accessible du premium et de ses conditions.
- Portfolio Charts (portfoliocharts.com) : les corrélations et le comportement par régime de dizaines de portefeuilles types, visualisés.
- Dans ce livre : [[primes-de-risque]] (ce que chaque brique rapporte), [[rendements-arithmetiques-geometriques]] (le volatility drag, moteur du premium), [[portefeuilles-tous-temps]] (la diversification poussée au bout), [[sequence-des-rendements]] (le risque que tout cela travaille).
