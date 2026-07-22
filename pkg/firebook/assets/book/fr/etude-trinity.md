# Bengen, l'étude Trinity et la naissance du taux de retrait sûr

Avant 1994, la question « combien puis-je retirer de mon portefeuille chaque année ? » recevait une réponse fausse avec une belle constance : « le rendement moyen, voyons ». Actions à 10 % historiques, donc retirez 8 %, disaient sans rire des professionnels du conseil.

William Bengen, puis l'étude Trinity, ont démontré pourquoi cette réponse ruine les gens, et inventé la méthode qui structure encore tout le domaine : rejouer l'histoire. Cette page raconte ce que ces travaux fondateurs ont réellement établi, comment leur méthode fonctionne, ce qu'elle a de génial et ce qu'elle a de daté.

C'est un article d'histoire des idées autant que de technique : les concepts introduits ici (millésime, SAFEMAX, taux de succès) servent dans tout le reste du livre.

::: cle Le renversement fondateur
L'apport de Bengen n'est pas le chiffre « 4 % ». C'est d'avoir montré que le taux de retrait soutenable ne dépend **pas** du rendement moyen, mais du pire enchaînement de rendements et d'inflation que le retraité traverse, surtout dans ses dix premières années. La moyenne des retraites américaines historiques supportait plus de 6 % ; le millésime 1966 ne supportait que ~4 %. Planifier, c'est planifier pour la queue de distribution, pas pour la moyenne. Tout le sujet moderne découle de ce renversement ([[sequence-des-rendements]]).
:::

## Le contexte : pourquoi la réponse « rendement moyen » ruine

L'intuition « les actions font 10 %, je peux retirer 8 % » commet deux fautes qui se cumulent.

La première : confondre moyenne arithmétique et croissance réellement composée. La volatilité fait que le portefeuille croît moins vite que sa moyenne annuelle ([[rendements-arithmetiques-geometriques]]), et l'inflation retranche encore 2 à 3 points. Le vrai moteur d'un portefeuille équilibré est un rendement **réel** géométrique de l'ordre de 3 à 5 %, pas 10.

La seconde, plus meurtrière : un retraité qui retire un montant fixe vend davantage de parts quand les cours sont bas. Deux séquences de rendements de **même** moyenne donnent alors des fortunes opposées selon que les mauvaises années arrivent au début ou à la fin ([[sequence-des-rendements]]). La moyenne ne dit presque rien ; l'ordre dit presque tout.

Au début des années 1990, Bengen, ingénieur du MIT reconverti en conseiller financier, voit arriver des clients à qui l'on a servi le « 8 % ». Plutôt que d'opposer une autre opinion, il fait ce que personne n'avait publié. Il teste.

## Bengen 1994 : la méthode des millésimes

L'idée est d'une simplicité lumineuse. Prenez les données annuelles américaines depuis 1926 (actions S&P, obligations d'État à moyen terme, inflation ; Bengen utilise les données Ibbotson). Fabriquez un retraité fictif partant le 1er janvier 1926 avec un portefeuille 50/50 et un retrait initial de, disons, 4 % indexé sur l'inflation. Déroulez année après année : rendement, retrait, rééquilibrage. Notez combien d'années le portefeuille survit. Recommencez pour un départ en 1927, 1928... chaque année de départ est un **millésime** (une cohorte, le « retirement cohort » d'ERN). Puis recommencez tout pour d'autres taux de retrait.

Le résultat tient dans un graphique resté célèbre : la durée de survie du portefeuille par millésime, pour chaque taux. À 3 %, tous les millésimes dépassent 50 ans. À 4 %, le pire millésime tient 33 ans, et **tous** tiennent au moins 30 ans. À 5 %, les départs de la fin des années 1960 s'épuisent en ~20 ans. À 6 %, des dizaines de millésimes échouent avant 20 ans.

Bengen nomme **SAFEMAX** le taux maximal qui survit à tous les millésimes sur l'horizon choisi : environ 4,15 % pour 30 ans en 50/50. Et il identifie les trois pires époques pour partir : 1929 (déflation et krach), 1937, et surtout **1966**, non pas le pire krach, mais la pire **combinaison**, quinze ans de marché réel nul avec une inflation qui gonfle les retraits. Leçon capitale : l'ennemi du rentier n'est pas le krach spectaculaire, c'est l'érosion réelle prolongée ([[inflation-et-taux-de-retrait]]).

Ses articles suivants (1996-2006) complètent le cadre. L'allocation optimale se situe entre 50 et 75 % d'actions, car descendre plus bas **abaisse** le taux sûr : les obligations seules ne résistent pas à l'inflation. Ajouter des petites capitalisations (small caps) remonte le SAFEMAX. Et l'horizon compte, ~4,3 % pour 25 ans, ~4,1 % pour 30 ans, ~3,5 % pour du très long.

::: encart Pourquoi cette méthode était géniale, et ce qu'elle vaut encore
Le rejeu historique (les « fenêtres historiques ») reste, trente ans après, l'un des quatre modèles de référence de la page FIRE ([[la-machine-pofo]]). Sa force : il préserve tout ce que les modèles synthétiques peinent à capturer, les enchaînements réels (krach **puis** inflation **puis** reprise), les corrélations actions-obligations changeantes, les longues mémoires. Sa faiblesse : il ne contient que le passé américain, un échantillon d'un seul pays, béni entre tous, où les fenêtres se chevauchent (il n'y a que 3 ou 4 périodes de 30 ans réellement indépendantes depuis 1926). D'où les correctifs modernes : échantillon mondial ([[anarkulova-cederburg]]), bootstrap et modèles paramétriques ([[historique-vs-parametrique]]).
:::

## Trinity 1998 : du plancher à la probabilité

Quatre ans plus tard, trois professeurs de finance de la Trinity University (Texas), Philip Cooley, Carl Hubbard et Daniel Walz, publient « Retirement Savings: Choosing a Withdrawal Rate That Is Sustainable ». Même méthode de rejeu, mais un déplacement conceptuel : au lieu du taux plancher qui survit à **tout** (le SAFEMAX de Bengen), ils publient une **grille de taux de succès**. Pour chaque combinaison taux de retrait × allocation × horizon, elle donne le pourcentage des fenêtres historiques où le portefeuille finit avec un solde positif.

Extrait de la logique de la grille (chiffres de l'étude actualisée, retraits indexés sur l'inflation, données 1926-2009) :

| Taux initial | 100 % actions, 30 ans | 75/25, 30 ans | 50/50, 30 ans | 25/75, 30 ans |
|---|---|---|---|---|
| 3 % | 100 % | 100 % | 100 % | 100 % |
| 4 % | 98 % | 100 % | 96 % | 71 % |
| 5 % | 80 % | 82 % | 67 % | 27 % |
| 6 % | 62 % | 60 % | 51 % | 20 % |

Trois enseignements durables sortent de cette grille. D'abord la **falaise**. Entre 4 et 5 %, le succès s'effondre : le sujet est non linéaire, et c'est pour cela que « juste un peu plus » de retrait coûte si cher. Ensuite l'effet d'allocation, asymétrique : trop peu d'actions est bien plus dangereux que trop (le 25/75 échoue une fois sur trois là où le 75/25 ne faiblit pas). Enfin, la notion même de « taux de succès ». C'est Trinity qui installe la probabilité de ruine comme langue commune du domaine, celle que parlent tous les simulateurs modernes ([[ruine-et-probabilites]]).

::: attention Ce que « 95 % de succès » veut dire chez Trinity, et ne veut pas dire
Le pourcentage de Trinity compte des **fenêtres historiques chevauchantes** du seul marché américain : « 95 % » signifie « 95 % des départs entre 1926 et 1980 auraient tenu », pas « votre plan a 95 % de chances de réussir ». Les fenêtres partagent leurs années (le krach de 1929 apparaît dans des dizaines de fenêtres), l'échantillon indépendant est minuscule, et le futur n'est pas tiré de cette urne. Les probabilités affichées par les simulateurs modernes ont des limites cousines ([[pieges-des-simulateurs]], [[lire-un-fan-chart]]) ; la parade est toujours la même, croiser plusieurs modèles et garder des marges.
:::

## Ce que les fondateurs n'avaient pas (encore) vu

Lire Bengen et Trinity aujourd'hui, c'est admirer la méthode et mesurer le chemin parcouru depuis. Les angles morts, tous traités ailleurs dans ce livre, dessinent le programme de la recherche moderne :

- **L'horizon FIRE.** 30 ans était l'horizon du retraité de 65 ans. À 45-55 ans d'horizon, le SAFEMAX américain descend vers 3,25-3,5 % ([[serie-ern]]) et la grille de Trinity ne s'applique plus telle quelle ([[horizon-et-esperance-de-vie]]).
- **Le biais du survivant géographique.** Les données Ibbotson commencent en 1926 aux États-Unis, le pays qui a gagné deux guerres mondiales sur le sol des autres avant de finir superpuissance. L'échantillon mondial (16 pays, 1870-2020) raconte une histoire plus dure ([[anarkulova-cederburg]]). Il est repris comme modèle « broad sample ».
- **Les valorisations.** Bengen note dès 2006 le lien entre niveaux de marché au départ et SAFEMAX ; la formalisation (CAPE → taux initial) viendra après ([[valorisations-et-cape]], [[regles-cape]]).
- **La rigidité du retrait.** Le retraité de Bengen exécute sa règle 30 ans sans regarder. Toute la génération suivante de stratégies (Guyton-Klinger [[guyton-klinger]], guardrails modernes [[guardrails-morningstar]], amortissement [[amortissement-abw]]) part de l'idée inverse : réagir à l'information.
- **Frais, impôts, dépenses réelles** : hors champ chez les fondateurs, de premier ordre dans la vraie vie ([[combien-il-vous-faut]]).

Aucun de ces points n'est une réfutation. La méthode des millésimes est toujours debout ; ce sont ses entrées et son cadre qu'on a élargis. Bengen lui-même n'a cessé d'actualiser son chiffre, à la hausse pour le retraité américain classique de 65 ans avec un portefeuille plus diversifié, tout en rappelant que le chiffre dépend du cadre. Quand vous entendez « la règle des 4 % est morte » ou « le 4 % est trop timide », la bonne question est toujours : dans quel cadre, pour quel horizon, avec quelles marges ?

## Refaire Bengen vous-même

C'est l'un des grands mérites pédagogiques de la méthode. Elle se refait. La page FIRE propose un mode « fenêtres historiques » qui rejoue exactement la logique des millésimes sur l'historique de **votre** portefeuille. Une vue « millésimes » (vintages) montre alors, départ par départ, où votre plan aurait tenu ou cassé ([[utiliser-la-page-fire]]). L'exercice vaut la peine : voir son plan traverser 1966 ou 2000 rend le risque de séquence plus concret que n'importe quelle probabilité.

::: exemple Lire un millésime
Plan : 1 M€, 60/40, retrait 4 % indexé. Dans la vue millésimes, le départ « janvier 2000 » montre la trajectoire type d'un mauvais cru : deux krachs dans la première décennie, le portefeuille réel divisé par deux vers 2009, une remontée qui ne rattrape jamais la trajectoire des bons millésimes, et une arrivée à 30 ans essoufflée mais solvable. Le départ « 2009 », lui, plane loin au-dessus. Même règle, même portefeuille, même moyenne de long terme : seule la **date** de départ diffère. C'est le risque de séquence rendu visible, et la meilleure introduction possible à [[sequence-des-rendements]].
:::

## L'essentiel à retenir

- Bengen (1994) invente la méthode du rejeu par millésimes et le SAFEMAX : le taux qui survit au **pire** départ historique, ~4,15 % sur 30 ans aux États-Unis. Le « 4 % » est un plancher de pire cas américain, pas une moyenne.
- Trinity (1998) transforme le plancher en grille de probabilités de succès et installe le langage de la ruine ; sa grille montre la falaise entre 4 et 5 % et le danger des portefeuilles trop peu actions.
- Le pire ennemi identifié n'est pas le krach mais l'érosion réelle prolongée (millésime 1966) : inflation et marché plat.
- Les angles morts (horizon FIRE, biais américain, valorisations, rigidité, frais et impôts) définissent la recherche moderne. C'est l'objet du reste de cette partie.
- La méthode se refait sur votre propre plan, sur la page FIRE : faites-le, un millésime vécu vaut mille probabilités.

---

## Pour aller plus loin

- William Bengen, « Determining Withdrawal Rates Using Historical Data », *Journal of Financial Planning*, octobre 1994 (en libre accès sur le site du FPA) ; et *Conserving Client Portfolios During Retirement* (2006) pour la synthèse.
- Cooley, Hubbard & Walz, « Retirement Savings: Choosing a Withdrawal Rate That Is Sustainable », *AAII Journal*, février 1998, et ses mises à jour (2011).
- Early Retirement Now, SWR Series volet 1 et volet 8 (l'appendice technique de la méthode) : [earlyretirementnow.com](https://earlyretirementnow.com) ([[serie-ern]]).
- Wade Pfau, « An International Perspective on Safe Withdrawal Rates » (2010) : la première grande sortie du cadre américain, prélude à [[anarkulova-cederburg]].
- Dans ce livre : [[les-maths-du-4-pourcent]] (pourquoi le chiffre de Bengen tient mathématiquement, étage par étage).
