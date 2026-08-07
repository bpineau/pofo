# Inflation et taux de retrait : le lien exact

Tout le monde sait que « l'inflation est mauvaise pour les retraités ». Presque personne ne sait dire **par où** elle attaque un plan de retrait. Un épisode de cinq ans à 8 % est pourtant incomparablement plus destructeur que trente ans à un point de plus que prévu, alors que la perte de pouvoir d'achat cumulée est la même.

Cet article établit le lien exact, mécanisme par mécanisme. D'abord l'effet ciseaux : les retraits indexés montent **pendant** que les actifs nominaux stagnent, et le plan est attaqué par les deux bouts. Ensuite la compression des rendements réels pendant les épisodes, où presque tout perd en réel **en même temps** ; c'est cette corrélation qui fait des épisodes d'inflation les pires millésimes de l'histoire, 1966 devant 1929. Viennent alors les chiffres conditionnels, c'est-à-dire ce que valent les taux de retrait selon le régime d'inflation de départ, une question que la série d'ERN a mesurée. On en tire enfin un audit : l'**inventaire d'indexation** de votre plan, qui distingue ce qui suit les prix, ce qui ne les suit pas et ce qui les suit en négatif, pour lire la vraie exposition nette. Reste à voir ce que les simulations en réel contiennent déjà, et ce qu'il faut leur demander en plus.

::: cle Le point technique central
Un plan simulé « en réel » contient **déjà** l'inflation moyenne. Raisonner en euros constants est la convention de tout simulateur sérieux. Les rendements sont déflatés, les retraits sont exprimés en pouvoir d'achat constant, et l'inflation régulière se trouve ainsi neutralisée par construction. Ce que le réel ne neutralise **pas**, c'est le **risque d'épisode**. Un épisode d'inflation n'est pas « des prix qui montent » : c'est un régime où les rendements **réels** de presque tout deviennent simultanément négatifs pendant des années ([[regimes-de-marche]]). Vu d'une simulation, cela ressemble à une longue séquence de très mauvaises années réelles corrélées. Seuls les modèles qui ont de la mémoire savent produire cette séquence. Un tirage indépendant année par année, lui, la sous-représente ([[historique-vs-parametrique]]).
:::

## Mécanisme 1 : l'effet ciseaux

Reprenons la mécanique du retrait fixe indexé ([[retrait-fixe-bengen]]) pendant un épisode. L'inflation à 8 % **gonfle** le retrait de 8 % par an : c'est le contrat, le pouvoir d'achat est maintenu. Pendant ce temps, le portefeuille nominal ne suit pas, car les obligations perdent en prix et les multiples des actions se compriment. Le taux de retrait courant (retrait divisé par portefeuille) grimpe alors par ses deux étages, numérateur en hausse mécanique et dénominateur en baisse réelle. C'est le mécanisme des ciseaux, et sa violence est arithmétique. Trois ans à 9 % d'inflation avec un portefeuille nominal stable font passer un taux courant de 4 % à ~5,2 %, l'équivalent d'un krach de 23 %, et cela **sans un seul jour de krach**.

Voilà pourquoi le millésime 1966 est pire que 1929 dans toutes les études ([[etude-trinity]]). Le krach de 1929 est brutal, puis il rend : la déflation fait **baisser** les retraits en nominal avec les prix, et l'indexation joue alors en votre faveur. L'épisode 1966-1981, lui, ne rend jamais : quinze ans de ciseaux.

Le corollaire est immédiat pour les règles de retrait. Les amendements anti-inflation du fixe attaquent **précisément** ce mécanisme : plafonner l'indexation, ou la geler après les années rouges ([[retrait-fixe-bengen]]). Renoncer à 2-3 points d'indexation pendant un épisode, c'est désarmer la moitié des ciseaux pour un coût de pouvoir d'achat modeste et étalé. C'est le meilleur rapport douleur/protection de toute la flexibilité ([[flexibilite-realite]]).

## Mécanisme 2 : la compression réelle simultanée

Le second mécanisme est celui des régimes ([[regimes-de-marche]], [[inflation-histoire]]). Pendant un épisode, les rendements **réels** de presque toutes les classes deviennent négatifs ensemble. Les obligations nominales le font par définition, car leur coupon fixe court après les prix. Le cash et le fonds euros le font par la répression, car les taux servis suivent avec retard. Les actions le font **pendant** l'épisode, par compression des multiples : le CAPE américain passe de 24 en 1966 à 7 en 1982, les bénéfices nominaux courent mais les prix non ([[valorisations-et-cape]]). La diversification classique ne diversifie alors plus rien. C'est la situation où le portefeuille 60/40 n'a aucune poche qui gagne. L'année 2022 en fut le rappel éclair, 1973-1981 la version longue.

Pour le plan, la combinaison des deux mécanismes définit le profil de l'ennemi : des années réelles négatives, corrélées entre actifs, persistantes, avec un passif qui grossit en face. Relisez cette phrase. C'est la définition exacte du pire cas du risque de séquence ([[sequence-des-rendements]]). L'inflation n'est pas un risque du plan parmi d'autres. Elle est la cause première historique du pire cas de séquence, et les deux chapitres décrivent la même bête sous deux angles.

::: science Les chiffres conditionnels : partir dans quel régime ?
ERN a consacré deux volets à la question conditionnelle, le volet 41 sur les environnements de basse inflation et le volet 51 sur la retraite en haute inflation. Le résultat est contre-intuitif. Le **niveau** d'inflation au départ ne dit presque rien du taux soutenable ; c'est l'inflation des années **qui suivent** qui le rabote, de l'ordre de 0,2 point de retrait par point d'inflation sur dix ans. Les pires millésimes américains sont ainsi des départs du milieu des années 1960, à 3,8-4,2 % même sur 30 ans, partis en inflation basse et rattrapés par la grande inflation. Les meilleurs sont les départs en marché bradé du début des années 1980, autour de 6-8 %, quand la désinflation était déjà installée. La grande variable conditionnante lisible au départ reste donc le CAPE ([[valorisations-et-cape]]), et les deux sujets se recoupent, puisque les épisodes compriment le CAPE.

Le point le plus subtil est ailleurs. La **basse** inflation n'est pas un régime sûr en soi. Elle s'accompagne de taux et de rendements attendus bas ([[rendements-attendus]]), si bien que le taux de retrait soutenable est comprimé par l'autre bout. En pratique, le régime d'inflation se **lit** (IPCH, points morts d'inflation ou breakevens, [[suivre-inflation]]) et module la prudence initiale, sans peser autant que le CAPE et sans jamais se prédire à 30 ans.
:::

## L'audit : l'inventaire d'indexation de votre plan

Voici l'outil pratique central de l'article. Puisque l'ennemi attaque par l'écart entre passifs indexés et actifs nominaux, dressez l'inventaire des **deux** colonnes de votre plan :

| Élément du plan | Indexation | Note |
|---|---|---|
| Dépenses / retraits | **Indexés** (le contrat du plan) | Plus la dérive personnelle ([[suivre-inflation]]) |
| Pension légale | **Indexée** (sur l'IPC, par la loi) | L'actif anti-inflation n° 1 du plan français, avec des revalorisations parfois décalées ou gelées politiquement ([[retraite-legale]]) |
| Rentes privées | **Non** (revalorisation discrétionnaire) | La grande faiblesse française du produit ([[rentes-et-annuites]]) |
| Linkers / échelle indexée | **Indexés** (contractuel) | La couverture propre ([[obligations-indexees]]) |
| Loyers perçus | **Quasi** (IRL, plafonnements politiques possibles) | Le linker vivant, avec risque réglementaire ([[immobilier-en-retrait]]) |
| Actions mondiales | **Non** à court terme, **oui** à long terme | Souffrent pendant, se revalorisent après : la protection lente |
| Obligations nominales, fonds euros, cash | **Non** | Les victimes des ciseaux et de la répression |
| Or, actifs réels | Épisodiquement | La couverture de crise, pas de croisière ([[or-en-retrait]]) |
| Crédit à taux fixe restant dû | **Indexation négative** | L'inflation rembourse pour vous : le seul poste qui **aime** l'épisode ([[immobilier-en-retrait]]) |

L'exposition nette du plan se lit dans ce tableau. Un plan français type (pension différée indexée, portefeuille 60/40 nominal, fonds euros) est **long** inflation sur ses vieux jours, grâce à la pension. Il est très **court** pendant la phase à découvert, où tout le reste le tire dans l'autre sens ; c'est précisément la période où la séquence décide ([[horizon-et-esperance-de-vie]]). La conclusion d'allocation tombe toute seule. C'est la phase à découvert qu'il faut indexer, avec des linkers, en poche comme en échelle de plancher, une part d'actifs réels et une dérive budgétée. La phase adossée, elle, l'est déjà par la pension.

## Ce que la simulation teste, et la check-list

Récapitulons l'outillage, car chaque idée de cet article se met à l'épreuve en simulation. Le **réel** de bout en bout neutralise l'inflation moyenne, et il n'y a rien à faire de plus pour elle. La **dérive** personnelle, en revanche, se déclare à la main. Elle demande deux réglages qu'un bon simulateur expose, une pente réelle ajoutée aux dépenses (+0,3 à +0,5 point par an couvre la dérive santé) et un profil de dépenses par âge, le sourire ([[suivre-inflation]], [[depenses-en-retraite]]). Le **risque d'épisode**, lui, ne se lit que dans les modèles qui ont de la mémoire.

Le rééchantillonnage du siècle multi-pays vient d'abord, celui d'Anarkulova-Cederburg, que les simulateurs appellent le broad-sample ([[anarkulova-cederburg]]). Ses blocs contiennent les vraies stagflations, le Royaume-Uni et l'Italie des années 1970, la France d'après-guerre. C'est le seul modèle où l'ennemi de cet article existe à l'état natif ([[historique-vs-parametrique]]). Le stress de séquence vient ensuite, avec ses grappes d'années réelles négatives, agnostique sur leur cause. Le rejeu des **millésimes** ferme la marche : 1966 et 1973 sont vos deux tests d'inflation nommés, et tout outil de rejeu historique américain les déroule cohorte par cohorte (cFIREsim et FICalc, par exemple).

La check-list de l'article tient en quatre questions. Le plan tient-il sous le siècle multi-pays ? Si les échecs sont des blocs inflationnistes, la réponse est dans l'inventaire d'indexation, pas dans plus de capital. Le millésime 1966 est-il traversé ? La dérive personnelle est-elle budgétée ? Et l'indexation du plancher de la phase à découvert est-elle organisée (linkers, échelle) ou seulement espérée ?

::: exemple Le même plan, avant et après l'audit d'indexation
Le plan de départ : 1,6 M€, 52 000 €/an, 45 ans, avec une pension indexée de 21 000 € en année 16. Le portefeuille initial est 65 % actions et 35 % nominal (fonds euros et aggregate). La lecture donne 4,3 % de ruine en central et 11,2 % en broad-sample. Les trajectoires échouées du broad-sample sont aux deux tiers des blocs inflationnistes, et le rejeu du millésime 1973 casse à l'année 24. L'audit est sans appel : la phase à découvert est 100 % courte d'inflation.

On applique l'inventaire. On prend 8 % de linkers courts, une échelle indexée-approchée sur 6 ans de plancher et 5 % d'or, le tout pris sur le nominal. On budgète une dérive de dépenses de +0,4 point par an au-dessus de l'indice. On écrit dans la règle le gel d'indexation des retraits après une année rouge. Nouvelle lecture : 4,1 % en central (rien de plus, l'assurance ne paie pas en croisière), 7,8 % en broad-sample, et 1973 traversé. Aucun capital n'a été ajouté. Le plan a simplement cessé d'être court sur son ennemi principal.
:::

## L'essentiel à retenir

- Deux mécanismes se combinent. Les **ciseaux** font monter les retraits indexés pendant que les actifs nominaux stagnent : trois ans à 9 % équivalent à un krach de 23 % sans krach. La **compression réelle simultanée** fait perdre à presque tout de la valeur réelle en même temps. Leur combinaison est le pire cas du risque de séquence, 1966 devant 1929.
- La déflation, elle, joue pour le rentier indexé : les retraits baissent avec les prix et les obligations montent. Cette asymétrie justifie que la couverture porte sur l'inflation, l'assurance-déflation restant une simple dose ([[obligations-en-retrait]]).
- Les simulations en réel contiennent l'inflation **moyenne**, pas le risque d'**épisode**. Celui-ci se teste dans les modèles qui ont de la mémoire, le broad-sample surtout, seul endroit où les vraies stagflations existent, et dans le rejeu des millésimes 1966 et 1973.
- L'outil pratique est l'**inventaire d'indexation** : pensions et linkers indexés, rentes privées et nominal non indexés, crédit fixe en négatif. Le plan français type est long inflation à l'arrivée, grâce à la pension, et dangereusement court pendant la phase à découvert. C'est elle qu'on indexe, avec des linkers, une échelle, des actifs réels et une dérive budgétée.
- Les amendements d'indexation du retrait (gel après une année rouge, plafond) désarment la moitié des ciseaux pour un coût étalé : c'est le meilleur rapport douleur/protection de la flexibilité.

---

## Pour aller plus loin

- Early Retirement Now, volet 41 (basse inflation) et volet 51 (haute inflation) : les SAFEMAX conditionnels au régime ([[serie-ern]]).
- Les données du millésime 1966-1981 dans toutes les études historiques ([[etude-trinity]]) : l'épisode de référence, à connaître dans le détail.
- Dans ce livre : [[inflation-histoire]] (les régimes), [[suivre-inflation]] (la mesure et le réglage de la dérive), [[se-proteger-de-inflation]] (les défenses une à une), [[obligations-indexees]] (la couverture contractuelle).
- La façon dont pofo implémente ces bancs d'essai, broad-sample et millésimes compris, est décrite dans [[la-machine-pofo]].
