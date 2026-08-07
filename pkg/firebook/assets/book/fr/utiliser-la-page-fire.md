# Utiliser la page FIRE de pofo

`pofo -fire` ouvre dans votre navigateur un laboratoire de plan de retraite. Vous décrivez votre situation : capital, dépenses, âge, pension, règles de dépense. La page déroule alors le même plan sous plusieurs modèles de marché, du plus fidèle à vos fonds au plus pessimiste du siècle. Cette page du livre en est le mode d'emploi complet. Elle explique dans quel ordre lire les sections, ce que contrôle chaque groupe de paramètres et surtout comment **interpréter** ce que vous voyez. L'outil est conçu pour instruire une décision, pas pour rendre un verdict.

(Vous êtes d'ailleurs peut-être en train de lire ce livre depuis la page elle-même, le lien « book » se cache en bas du volet « How this machine works ».)

::: cle La philosophie de l'outil
Un seul chiffre de ruine est une opinion de modèle ([[ruine-et-probabilites]]). La page affiche donc le même plan sous plusieurs lentilles à la fois. D'un côté, les modèles fondés sur les **données** : l'historique de vos fonds rejoué ou rééchantillonné, et le siècle des 16 pays développés. De l'autre, les modèles fondés sur des **paramètres** : Student-t central, stress de séquence, décennie perdue. La bonne lecture est l'intervalle entre les colonnes. La bonne décision est celle qui reste acceptable dans les colonnes pessimistes.
:::

## Démarrer : les deux façons de lancer la page

**`pofo -fire` seul** ouvre le mode paramétrique : vous décrivez le marché par trois curseurs (rendement réel μ, volatilité σ, épaisseur des queues df). C'est le mode « bac à sable », parfait pour comprendre les mécanismes et dimensionner grossièrement.

**`pofo -fire portfolio.txt`** est le mode portefeuille. pofo reconstruit l'historique réel long de **vos** lignes (via les extensions `SIM`), le déflate en euros constants, et en tire deux choses. D'abord les paramètres du modèle central : μ/σ/df ajustés sur vos fonds, puis prudemment mélangés vers un prior mondial (voir plus bas). Ensuite deux modèles purement historiques, les fenêtres historiques et le bootstrap par blocs ([[historique-vs-parametrique]]). Vous pouvez aussi faire glisser le poids de chaque ligne et voir la ruine se recalculer en direct.

Tous les montants sont **réels** (pouvoir d'achat constant, inflation déjà retirée) et nets : 60 000 €/an, c'est 5 000 €/mois de dépenses d'aujourd'hui, pour toujours ([[rendements-arithmetiques-geometriques]]).

## Le tableau de bord : la bande des modèles

En haut de page, la bande de plan résume chaque modèle par une perle colorée (vert = sûr, rouge = catastrophique à votre niveau de dépenses). Le tableau principal, lui, aligne les colonnes : fenêtres historiques, bootstrap, Student-t central, stress de séquence, échantillon large (broad-sample), décennie perdue. **Ce tableau est un sélecteur.** Cliquez une colonne et **toutes** les sections de détail de la page se recalculent sous cette lentille (le soulignement ambre marque la colonne active). Le réflexe de lecture est le suivant.

1. **Les colonnes historiques de vos fonds** (fenêtres, bootstrap) : la borne optimiste ; vos fonds n'ont vécu qu'une fenêtre, souvent favorable.
2. **Le Student-t central** : le cas de planification par défaut, calibré sur vos fonds puis tiré vers la prudence.
3. **Le stress de séquence** : même rendement moyen, mais les mauvaises années arrivent en grappes ; l'écart avec le central est le prix de la séquence dans votre plan ([[sequence-des-rendements]]).
4. **Le broad-sample** : le siècle entier de 16 pays développés (1870-2020) en 60/40 domestique, catastrophes comprises ([[anarkulova-cederburg]]) ; l'estimation honnête du risque de long horizon.
5. **La décennie perdue** : un scénario de queue à la japonaise, à rendre **surmontable**, pas improbable.

La règle de décision suggérée : planifiez entre le central et le broad-sample. Servez-vous du stress de séquence comme test d'ordre des rendements, et de la décennie perdue comme crash-test.

À côté du tableau, la jauge affiche la ruine du modèle sélectionné, comparée à votre ruine acceptable (le curseur « acceptable ruin », 4 % par défaut). Ce seuil alimente tous les solveurs de la page ([[ruine-et-probabilites]] pour bien le choisir).

## Les sections, dans l'ordre de lecture

**§00 Where we are in the cycle** : le CAPE de Shiller du jour, replacé dans un siècle d'histoire. Pas un signal de timing : un rappel que le point de départ conditionne la première décennie, celle qui décide ([[valorisations-et-cape]]).

**§01 Simulated futures** : les cônes de richesse (fan charts) sous les quatre lentilles côte à côte, avec des trajectoires d'exemple, les échouées en rouge. Comment lire un cône sans se tromper : [[lire-un-fan-chart]].

**§02 The retirements that actually happened** : votre plan rejoué aux pires dates de départ du siècle, USA 1929, 1966, 2000, Japon 1990. Des millésimes réels, pas des tirages ([[etude-trinity]]).

**§03 The decisive decade** : la ruine décomposée selon le rendement des dix premières années de chaque scénario. C'est le risque de séquence rendu visible sur **votre** plan ([[sequence-des-rendements]]).

**§04 The spending you actually live** : ce que vos règles de dépense produisent réellement (le niveau de vie vécu année par année, et qui le finance, portefeuille, pension ou buffer). Cette section est cruciale dès que vous activez une règle flexible. La ruine baisse, mais vous voyez ici le **prix** payé, en années de dépenses réduites ([[flexibilite-realite]]).

**§05 Alive, broke or gone** : la ruine croisée avec la mortalité d'un couple français. Être ruiné à 61 ans ou à 92 ans, ce n'est pas le même événement. C'est ici qu'on relativise (ou non) un chiffre de ruine brut.

**§06 What moves the risk** : la frontière des règles de retrait (chaque règle = un point ruine × variabilité du niveau de vie, [[panorama-strategies-retrait]]), puis les leviers classés par sensibilité. On voit ce qui bouge vraiment votre risque, et ce qui ne le bouge pas.

**§07 Buffer & recovery** : l'arbitrage du matelas de liquidités (ruine et richesse finale selon le nombre d'années de buffer, [[cash-buffer]]), et la distribution des « années sous l'eau ». Elle montre combien de temps durent les traversées du désert que le buffer doit couvrir.

**§08 Plan detail** et **§09 Reaching your target** : le détail chiffré du plan sous le modèle sélectionné, et le menu du solveur. Ce menu liste les mouvements équivalents (capital en plus, dépenses en moins, année de plus, pension...) qui ramèneraient chacun votre ruine sous votre seuil. C'est la section « négociation avec soi-même ». Elle chiffre le prix de chaque marge ([[une-annee-de-plus]]).

## Le tiroir de paramètres, groupe par groupe

Le bouton « parameters » ouvre le tiroir. Chaque contrôle a une aide au survol ; voici la carte et les pièges.

**Your situation** : capital déployé (hors résidence et fonds d'urgence), âge au départ (alimente §05), horizon (planifiez **au-delà** de votre espérance de vie, 40 → 50 ans double presque la ruine, [[horizon-et-esperance-de-vie]]), dépenses nettes annuelles (réelles, nettes d'impôt ; la friction fiscale est modélisée à part).

**Pension & side income** : la pension nette réelle et son année de début, puis les revenus d'appoint temporaires des premières années. Les préréglages couvrent un scénario stressé, le central « droits acquis » et le simulateur officiel ([[retraite-legale]]). Ces revenus d'appoint sont la meilleure assurance anti-séquence qui existe ([[revenus-complementaires]]). C'est la deuxième sensibilité du plan après les dépenses : ne les laissez pas à zéro par fausse prudence ([[erreurs-classiques-fire]]).

**Spending policy** : le cœur stratégique. La coupe réversible en drawdown (flexCut, 15 % de coupe divise environ la ruine par deux), le déclencheur par taux de retrait courant, les guardrails Guyton-Klinger avec plancher ([[guyton-klinger]]), le cliquet à la hausse, la dérive structurelle des dépenses et le « retirement smile » ([[depenses-en-retraite]]), le VPW pur ([[pourcentage-fixe]]), l'amortissement ABW/TPAW ([[amortissement-abw]]), le pourcentage borné à la Vanguard ([[plancher-plafond]]) et la rente viagère indexée ([[rentes-et-annuites]]). Chaque règle écrase les précédentes quand elle est activée ; le volet « How this machine works » de la page les détaille une à une, et la partie IV de ce livre y consacre un article chacune ([[choisir-sa-strategie]]).

**Market model** : μ (moyenne **arithmétique** réelle du moteur de croissance, le géométrique vécu vaut ≈ μ − σ²/2), σ (volatilité de long horizon, plus basse que la vol à un an des brochures), df (queues, à df 5, l'année catastrophique est ~10 fois plus probable qu'en loi normale, [[queues-epaisses]]). En mode portefeuille, ces valeurs sont pré-remplies depuis vos fonds puis **mélangées** vers un prior mondial prudent à proportion de ce que l'horizon excède l'historique ([[rendre-monte-carlo-pertinent]]). Deux ancres utiles : « Broad-sample prior » (réécrit les curseurs avec les hypothèses prudentes du siècle) et « Anchor to CAPE » (remplace la seule moyenne par 1/CAPE, la compression de rendement qu'impliquent les valorisations du jour, [[regles-cape]]). S'y trouvent aussi le glidepath rising-equity ([[glidepaths]]) et les retraits mensuels.

**Cash buffer** : années de dépenses en matelas (défaut 3), son rendement réel, et l'année où le rechargement s'arrête. Attention à la convention : le buffer est **prélevé** sur le capital de départ, pas ajouté ([[cash-buffer]], [[recharger-ou-pas]]).

**Taxes** : le taux sur les gains (31,4 % en préréglage PFU + PUMa approché), appliqué en majorant chaque vente. Retirer 60 k€ nets vend plus de 60 k€ d'actifs, et la charge effective monte avec les plus-values latentes ([[flat-tax-et-imposition]], [[taxe-puma]]). Réglez-le sur votre taux mixte réel, selon vos enveloppes ([[enveloppes-francaises]]).

::: astuce La séance type, en six gestes
1) Entrez situation, pension, dépenses réelles auditées ([[combien-il-vous-faut]]). 2) Lisez la bande : l'intervalle des ruines, colonne par colonne. 3) Cliquez broad-sample et regardez §02 et §03 : votre plan survit-il aux vrais désastres ? 4) Activez votre règle de dépense candidate et lisez §04 : le niveau de vie vécu vous convient-il dans le mauvais quart ? 5) Ouvrez §09 : que coûterait de ramener la ruine sous votre seuil, et quel levier est le moins cher pour vous ? 6) Notez la configuration retenue et les seuils dans votre plan écrit ([[construire-son-plan]]). Durée honnête : une heure la première fois, dix minutes en revue annuelle ([[revue-annuelle]]).
:::

::: attention Les trois mésusages classiques
Pousser μ « parce que mon fonds a fait mieux » : l'historique court d'un fonds est une fenêtre favorable, pas une espérance ([[pieges-des-simulateurs]]) ; le mélange vers le prior existe précisément contre ça. Ne regarder que la colonne verte : si le plan n'est acceptable que dans le modèle optimiste, il n'est pas acceptable. Confondre les décimales avec du signal, enfin : lisez les **écarts** (entre colonnes, entre scénarios, entre leviers), jamais la deuxième décimale ([[ruine-et-probabilites]]).
:::

## Ce que la page ne fait pas

Par honnêteté de conception, plusieurs choses restent hors champ. Pas de prévision, d'abord : aucun modèle ne prédit, ils encadrent. Pas de fiscalité fine par enveloppe : un taux mixte global, à vous de le calibrer. Pas d'actifs illiquides : l'immobilier locatif se modélise en revenu d'appoint ([[immobilier-en-retrait]]). Pas de conseil, enfin : l'outil explore des hypothèses, les décisions restent les vôtres. Le détail des modèles, des données qui les nourrissent et des réserves méthodologiques vit dans les deux volets pliants en bas de page, et dans [[la-machine-pofo]] pour la version approfondie.

## L'essentiel à retenir

- Deux modes : bac à sable paramétrique (`pofo -fire`) ou calibré sur vos fonds (`pofo -fire portfolio.txt`) ; tout est en euros réels, dépenses nettes.
- Le tableau du haut est un sélecteur : cliquez une colonne, toute la page se recalcule sous cette lentille ; planifiez entre le central et le broad-sample.
- L'ordre de lecture : la bande (l'intervalle), §02-03 (les désastres réels et la séquence), §04 (le prix vécu de la flexibilité), §09 (le prix de chaque marge).
- Les leviers les plus puissants, dans l'ordre habituel : les dépenses et leur flexibilité, la pension et les revenus d'appoint précoces, puis seulement le portefeuille.
- Les paramètres de marché sont pré-remplis et volontairement tirés vers la prudence ; les ancres broad-sample et CAPE sont vos garde-fous, pas des options exotiques.

---

## Pour aller plus loin

- Les deux volets pliants de la page elle-même : « How this machine works » (chaque contrôle, chaque modèle) et « Method & honest caveats ».
- La version longue de la tuyauterie ([[la-machine-pofo]]), les familles de modèles ([[historique-vs-parametrique]]) et la lecture des cônes ([[lire-un-fan-chart]]).
- Le **readme** de pofo (section « Decumulation / FIRE analysis ») pour les options de la CLI.
