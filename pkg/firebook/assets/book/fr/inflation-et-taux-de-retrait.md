# Inflation et taux de retrait : le lien exact

Tout le monde sait que « l'inflation est mauvaise pour les retraités » ; presque personne ne sait dire PAR OÙ elle attaque un plan de retrait, ni pourquoi un épisode de cinq ans à 8 % est incomparablement plus destructeur que trente ans à 2,5 % de plus que prévu : alors que la perte de pouvoir d'achat cumulée peut être la même. Cet article établit le lien exact, mécanisme par mécanisme : l'effet ciseaux (les retraits indexés montent PENDANT que les actifs nominaux stagnent : le plan attaqué par les deux bouts), la compression des rendements réels pendant les épisodes (pourquoi presque tout perd en réel EN MÊME TEMPS : la corrélation qui fait des épisodes d'inflation les pires millésimes de l'histoire, 1966 devant 1929), les chiffres conditionnels (ce que valent les taux de retrait selon le régime d'inflation de départ : la série d'ERN l'a mesuré), puis l'audit qui en découle : l'INVENTAIRE D'INDEXATION de votre plan (ce qui suit les prix, ce qui ne les suit pas, ce qui les suit en négatif : la vraie exposition nette) : et ce que les simulations en réel de pofo contiennent déjà, et ce qu'il faut leur demander en plus.

::: cle Le point technique central
Un plan simulé « en réel » ([[la-machine-pofo]]) contient DÉJÀ l'inflation moyenne : les rendements sont déflatés, les retraits sont en pouvoir d'achat constant : l'inflation régulière est neutralisée par construction. Ce que le réel ne neutralise PAS : le RISQUE D'ÉPISODE : car un épisode d'inflation n'est pas « des prix qui montent », c'est un régime où les rendements RÉELS de presque tout deviennent simultanément négatifs pendant des années ([[regimes-de-marche]]) : c'est-à-dire, vu du simulateur, une séquence de très mauvaises années réelles corrélées : exactement ce que les colonnes stress et broad-sample existent pour tester, et ce que le modèle central i.i.d. sous-représente.
:::

## Mécanisme 1 : l'effet ciseaux

Reprenons la mécanique du retrait fixe indexé ([[retrait-fixe-bengen]]) sous un épisode : l'inflation à 8 % GONFLE le retrait de 8 % par an (c'est le contrat : le pouvoir d'achat est maintenu) pendant que le portefeuille nominal, lui, ne suit pas (les obligations perdent en prix, les actions compressent leurs multiples) : le taux de retrait courant (retrait/portefeuille) grimpe alors par les DEUX étages : numérateur en hausse mécanique, dénominateur en baisse réelle. C'est le mécanisme des ciseaux, et sa violence est arithmétique : trois ans à 9 % d'inflation avec un portefeuille nominal stable font passer un taux courant de 4 % à ~5,2 % : l'équivalent d'un krach de 23 % : SANS UN SEUL JOUR DE KRACH. Voilà pourquoi le millésime 1966 est pire que 1929 dans toutes les études ([[etude-trinity]]) : le krach de 1929 est brutal puis rend (déflation : les retraits BAISSENT en nominal avec les prix : l'indexation joue en votre faveur !) ; l'épisode 1966-1981 ne rend jamais : quinze ans de ciseaux.

Notez le corollaire immédiat pour les règles de retrait : les amendements anti-inflation du fixe (plafonner l'indexation, la geler après les années rouges, [[retrait-fixe-bengen]]) attaquent PRÉCISÉMENT ce mécanisme : renoncer à 2-3 points d'indexation pendant un épisode, c'est désarmer la moitié des ciseaux pour un coût de pouvoir d'achat modeste et étalé : le meilleur rapport douleur/protection de toute la flexibilité ([[flexibilite-realite]]).

## Mécanisme 2 : la compression réelle simultanée

Le second mécanisme est celui des régimes ([[regimes-de-marche]], [[inflation-histoire]]) : pendant un épisode, les rendements RÉELS de presque toutes les classes deviennent négatifs ensemble : les obligations nominales par définition (leur coupon fixe court après les prix), le cash et le fonds euros par la répression (les taux servis suivent avec retard), et les actions PENDANT l'épisode (la compression des multiples : le CAPE de 1966-1982 passe de 24 à 7 : les bénéfices nominaux courent, les prix non, [[valorisations-et-cape]]). La diversification classique ne diversifie alors plus rien : c'est LA situation où le portefeuille 60/40 n'a aucune poche qui gagne : 2022 en fut le rappel éclair, 1973-1981 la version longue.

Pour le plan, la combinaison des deux mécanismes définit le profil de l'ennemi : **des années réelles négatives, corrélées entre actifs, PERSISTANTES, avec un passif qui grossit en face** : relisez cette phrase : c'est la définition exacte du pire cas du risque de séquence ([[sequence-des-rendements]]). L'inflation n'est pas UN risque du plan parmi d'autres : elle est la cause première historique du pire cas de séquence : les deux chapitres décrivent la même bête sous deux angles.

::: science Les chiffres conditionnels : partir dans quel régime ?
ERN a consacré deux volets à la question conditionnelle (Part 41 : les environnements de basse inflation ; Part 51 : la retraite en haute inflation) : les résultats structurants : le SAFEMAX historique conditionné au régime d'inflation DE DÉPART est le plus bas pour les départs en inflation élevée ET MONTANTE (les années 1960-70 : 3,8-4,2 % même sur 30 ans) et le plus haut pour les départs en désinflation installée (1982+ : 6-8 %) ; l'inflation de départ est, avec le CAPE, l'autre grande variable conditionnante ([[valorisations-et-cape]] : et les deux se recoupent : les épisodes compriment le CAPE) ; et le point le plus subtil : la BASSE inflation n'est pas un régime sûr en soi : elle s'accompagne de taux et de rendements attendus bas ([[rendements-attendus]]) : le taux de retrait soutenable est comprimé par l'autre bout. La traduction planificateur : le régime d'inflation au départ se LIT (IPCH, breakevens, [[suivre-inflation]]) et module la prudence initiale, comme le CAPE : sans jamais se prédire à 30 ans.
:::

## L'audit : l'inventaire d'indexation de votre plan

Voici l'outil pratique central de l'article : puisque l'ennemi attaque par l'écart entre passifs indexés et actifs nominaux, dressez l'inventaire des DEUX colonnes de votre plan :

| Élément du plan | Indexation | Note |
|---|---|---|
| Dépenses / retraits | INDEXÉS (le contrat du plan) | Plus la dérive personnelle ([[suivre-inflation]]) |
| Pension légale | INDEXÉE (sur l'IPC, par la loi) | L'actif anti-inflation n° 1 du plan français : mais revalorisations parfois décalées/gelées politiquement ([[retraite-legale]]) |
| Rentes privées | NON (revalorisation discrétionnaire) | La grande faiblesse française du produit ([[rentes-et-annuites]]) |
| Linkers / échelle indexée | INDEXÉS (contractuel) | La couverture propre ([[obligations-indexees]]) |
| Loyers perçus | QUASI (IRL, plafonnements politiques possibles) | Le linker vivant, avec risque réglementaire ([[immobilier-en-retrait]]) |
| Actions mondiales | NON à court terme, OUI à long terme | Souffrent pendant, repricent après : la protection lente |
| Obligations nominales, fonds euros, cash | NON | Les victimes des ciseaux et de la répression |
| Or, actifs réels | Épisodiquement | La couverture de crise, pas de croisière ([[or-en-retrait]]) |
| Crédit à taux fixe restant dû | INDEXATION NÉGATIVE | L'inflation rembourse pour vous : le seul poste qui AIME l'épisode ([[immobilier-en-retrait]]) |

L'exposition nette du plan se lit dans ce tableau : un plan français type (pension différée indexée + portefeuille 60/40 nominal + fonds euros) est LONG inflation sur ses vieux jours (la pension) et très COURT pendant la phase à découvert (tout le reste) : précisément la période où la séquence décide ([[horizon-et-esperance-de-vie]]). La conclusion d'allocation tombe toute seule : c'est la phase à découvert qu'il faut indexer (linkers en poche et en échelle de plancher, part d'actifs réels, dérive budgétée), la phase adossée l'étant déjà par la pension.

## Ce que pofo teste, et la check-list

Récapitulons l'outillage, car les idées de cet article se vérifient toutes en séance ([[utiliser-la-page-fire]]) : le RÉEL de bout en bout neutralise l'inflation moyenne (rien à faire) ; la DÉRIVE personnelle se règle (spendDrift +0,3-0,5, sourire) ; le RISQUE D'ÉPISODE se lit dans les colonnes à régimes : le broad-sample d'abord (ses blocs contiennent les vraies stagflations : le Royaume-Uni et l'Italie des années 1970, la France d'après-guerre : c'est la seule colonne où l'ennemi de cet article existe à l'état natif, [[historique-vs-parametrique]]), le stress ensuite (les grappes d'années réelles négatives, agnostiques sur la cause) ; et les MILLÉSIMES §02 (1966 et 1973 : vos deux tests d'inflation nommés). La check-list de l'article, en quatre questions : le plan tient-il dans le broad-sample (sinon : regardez si les échecs sont des blocs inflationnistes : la réponse est alors dans l'inventaire d'indexation, pas dans plus de capital) ? le millésime 1966 est-il traversé ? la dérive personnelle est-elle budgétée ? et l'indexation du plancher de la phase à découvert est-elle organisée (linkers/échelle) ou espérée ?

::: exemple Le même plan, avant et après l'audit d'indexation
Plan : 1,6 M€, 52 000 €/an, 45 ans, pension indexée 21 000 € en année 16 : portefeuille initial 65 % actions / 35 % nominal (fonds euros + aggregate). Lecture pofo : central 4,3 %, broad-sample 11,2 % : et les trajectoires échouées du broad-sample sont aux deux tiers des blocs inflationnistes ; le millésime 1973 (§02) casse à l'année 24. Audit : la phase à découvert est 100 % courte d'inflation. Correction (l'inventaire appliqué) : 8 % de linkers courts + échelle indexée-approchée sur 6 ans de plancher + 5 % d'or, pris sur le nominal ; spendDrift +0,4 ; gel d'indexation des retraits après année rouge écrit dans la règle. Relecture : central 4,1 % (rien : l'assurance ne paie pas en croisière), broad-sample 7,8 %, 1973 traversé. Aucun capital ajouté : le plan a simplement cessé d'être court sur son ennemi principal.
:::

## L'essentiel à retenir

- Deux mécanismes : les CISEAUX (retraits indexés qui montent, actifs nominaux qui stagnent : trois ans à 9 % = un krach de 23 % sans krach) et la COMPRESSION RÉELLE SIMULTANÉE (pendant l'épisode, presque tout perd en réel ensemble) : leur combinaison EST le pire cas du risque de séquence : 1966 devant 1929.
- La déflation, elle, joue POUR le rentier indexé (les retraits baissent avec les prix, les obligations montent) : l'asymétrie justifie que la couverture porte sur l'inflation, l'assurance-déflation restant une dose ([[obligations-en-retrait]]).
- Les simulations en réel contiennent l'inflation MOYENNE, pas le risque d'ÉPISODE : celui-ci se teste dans les colonnes à régimes (broad-sample surtout : le seul modèle où les vraies stagflations existent) et les millésimes 1966/1973.
- L'outil pratique est l'INVENTAIRE D'INDEXATION : pensions et linkers indexés, rentes privées et nominal non, crédit fixe en négatif : le plan français type est long inflation à l'arrivée (pension) et dangereusement court pendant la phase à découvert : c'est ELLE qu'on indexe (linkers, échelle, actifs réels, dérive budgétée).
- Les amendements d'indexation du retrait (gel après année rouge, plafond) désarment la moitié des ciseaux pour un coût étalé : le meilleur rapport douleur/protection de la flexibilité.

---

## Pour aller plus loin

- Early Retirement Now, Part 41 (basse inflation) et Part 51 (haute inflation) : les SAFEMAX conditionnels au régime ([[serie-ern]]).
- Les données du millésime 1966-1981 dans toutes les études historiques ([[etude-trinity]]) : l'épisode de référence, à connaître dans le détail.
- Dans pofo : le broad-sample et la §02 comme bancs d'essai d'inflation, la dérive et le sourire comme réglages ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
- Dans ce livre : [[inflation-histoire]] (les régimes), [[suivre-inflation]] (la mesure), [[se-proteger-de-inflation]] (les défenses une à une), [[obligations-indexees]] (la couverture contractuelle).
