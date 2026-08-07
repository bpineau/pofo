# Les valorisations (CAPE) et ce qu'elles disent du taux de retrait

Voici le fait empirique le plus important de tout le sujet après le risque de séquence, et il tient en une phrase : tous les pires millésimes de départ à la retraite de l'histoire, 1929, 1966, 2000, le Japon 1990, ont un point commun, et ce n'est pas la malchance, mais le fait que le marché était historiquement **cher** au moment du départ. Le niveau de valorisation au jour du premier retrait est le meilleur prédicteur connu du taux de retrait qu'un millésime pourra soutenir.

L'instrument de mesure standard de cette cherté est le CAPE de Shiller, et cette page en fait le tour complet : ce que c'est exactement et comment il se calcule, pourquoi et dans quelle mesure il prédit les rendements, la relation chiffrée CAPE-taux de retrait que la recherche a établie, les critiques sérieuses de l'indicateur (il y en a), et surtout les quatre façons concrètes de s'en servir dans un plan, dont l'ancre CAPE intégrée à la page FIRE. À la fin, vous saurez lire le CAPE du jour comme un rentier, pas comme un trader : non pas « faut-il vendre ? », mais « que puis-je promettre à mon plan ? ».

::: cle L'idée en une phrase
Le prix que vous payez pour un flux de bénéfices détermine le rendement que ce flux pourra vous servir : acheter cher, c'est accepter des rendements futurs plus bas. Pour un rentier, dont le sort se joue dans la première décennie ([[sequence-des-rendements]]), partir quand le marché est cher signifie que la décennie décisive a une espérance de rendement comprimée et un risque de correction accru : la double peine. Le CAPE ne prédit pas les krachs ; il mesure la taille de la promesse que le marché peut tenir.
:::

::: figure cape-swr
Le taux de retrait qui aurait survécu à tous les millésimes, en fonction du CAPE au départ (ordres de grandeur, données américaines). Partir cher comprime le taux soutenable ; partir bon marché l'élargit.
:::

## Le CAPE : définition, calcul, origine

CAPE signifie *Cyclically Adjusted Price-to-Earnings ratio* : le ratio cours/bénéfices ajusté du cycle. La recette, proposée par Robert Shiller et John Campbell en 1988 (et inspirée de Graham et Dodd, 1934, qui recommandaient déjà de moyenner les bénéfices « sur cinq à dix ans, de préférence dix ») :

1. Prenez le prix de l'indice (le S&P 500, historiquement).
2. Divisez-le non pas par les bénéfices de l'année (le P/E classique), mais par la **moyenne des bénéfices réels des dix dernières années**, chaque année de bénéfices étant d'abord réévaluée en euros ou dollars constants via l'inflation.

Pourquoi dix ans ? Parce que les bénéfices d'une année sont un très mauvais dénominateur : en récession ils s'effondrent (ce qui gonfle mécaniquement le P/E au pire moment et fait paraître le marché « cher » au fond du trou de 2009, exactement le contresens), en haut de cycle ils sont dopés par des marges insoutenables. Dix ans couvrent un cycle économique complet : le dénominateur devient une estimation de la capacité bénéficiaire **normale** des entreprises, et le ratio mesure alors vraiment ce qu'on paie pour cette capacité.

Quelques repères pour calibrer l'œil (S&P 500, données Shiller depuis 1871) :

| Époque | CAPE | Ce qui a suivi (rendement réel actions, 10-15 ans) |
|---|---|---|
| Moyenne 1871-2025 | ~17 | ~6,5 %/an en moyenne de très long terme |
| 1921 (plancher d'après-guerre) | ~5 | Les années folles : >15 %/an |
| 1929 (avant le krach) | ~33 | Négatif sur 10 ans ; le pire millésime de Bengen avec 1966 |
| 1966 (sommet des « Nifty Fifty » avant l'ère inflationniste) | ~24 | ~0 % réel sur 15 ans : le pire départ américain ([[etude-trinity]]) |
| 1982 (plancher de la stagflation) | ~7 | Le plus grand bull market du siècle : >13 %/an |
| Décembre 1999 (bulle internet) | ~44 | Record absolu ; deux krachs de 50 % dans la décennie suivante |
| 2009 (creux de la crise financière) | ~13 | ~12 %/an réel sur la décennie suivante |
| Zone 2024-2026 | 33-38 | À écrire ; historiquement, cette zone n'a jamais livré mieux que ~4 % réel sur 10 ans |

La page FIRE affiche le CAPE du jour en tête de page (section §00, « Where we are in the cycle »), replacé sur son siècle d'histoire, précisément pour que chaque session de planification commence par ce constat de position ([[utiliser-la-page-fire]]).

## Pourquoi ça prédit, et ce que ça ne prédit pas

Le mécanisme n'a rien de mystique ; c'est de l'arithmétique de flux. Détenir le marché actions, c'est détenir un droit sur les bénéfices futurs des entreprises. Le rendement de long terme de l'actionnaire se décompose en trois morceaux : le rendement des bénéfices au prix d'achat (l'**earnings yield**, environ 1/CAPE), la croissance réelle de ces bénéfices (historiquement 1,5 à 2 %/an aux États-Unis), et la variation de la valorisation elle-même entre l'achat et la vente (l'expansion ou la contraction du multiple). Sur un an, le troisième terme domine tout et le CAPE ne prédit rien. Sur dix à quinze ans, il se moyenne : restent les deux premiers termes, dont le premier est **connu** au moment de l'achat. Un CAPE de 33, c'est un earnings yield de 3 % : la composante « certaine » de votre rendement réel futur est déjà plafonnée bas, quoi qu'il arrive au reste.

Empiriquement, la relation est l'une des plus solides de la finance : sur les données américaines 1871-2025, le CAPE de départ explique de l'ordre de 40 à 60 % de la variance des rendements réels des 10-15 années suivantes (le R² dépend de la fenêtre et de la période). C'est à la fois énorme (rien d'autre ne fait mieux) et très insuffisant pour du timing : à CAPE égal, l'éventail des issues à 10 ans reste large. La formulation honnête : **le CAPE déplace le CENTRE de la distribution des rendements futurs, sans en réduire beaucoup la LARGEUR**. Exactement l'information dont un planificateur a besoin, et exactement celle dont un trader ne peut rien faire.

Trois choses que le CAPE ne fait **pas**, à graver avant d'aller plus loin :

- **Il ne date rien.** Le CAPE a dépassé sa moyenne historique en 1996 ; la bulle a continué **quatre** ans et +100 %. Shiller lui-même publiait *Irrational Exuberance* en mars 2000, un timing de légende, mais son indicateur était « en alerte » depuis des années. Sortir du marché sur signal CAPE est la stratégie qui a ruiné le plus de gens prudents.
- **Il ne prédit pas les krachs.** Un marché cher peut dégonfler par krach (2000) ou par stagnation des prix pendant que les bénéfices rattrapent (au moins partiellement 2013-2019). Le CAPE prédit des rendements moyens faibles, pas leur chorégraphie.
- **Il ne se compare pas naïvement entre pays ni entre époques** : on y revient dans la section critiques.

## Le lien CAPE-taux de retrait : les chiffres

Venons-en au cœur du sujet pour un rentier. Puisque le sort d'une retraite se joue essentiellement dans sa première décennie ([[sequence-des-rendements]]), et que le CAPE prédit justement le rendement moyen de la décennie qui suit, on s'attend à une relation forte entre CAPE au départ et taux de retrait maximal soutenable (SAFEMAX, [[etude-trinity]]) du millésime. C'est massivement le cas, et c'est l'un des résultats les mieux rejoués de la littérature : Bengen l'esquisse en 2006, Kitces le documente en 2008 (« Resolving the Paradox: Is the Safe Withdrawal Rate Sometimes Too Safe? »), Pfau le formalise en régression en 2011, ERN le systématise sur données mensuelles au volet 3 puis au volet 54 de sa série ([[serie-ern]]).

Les ordres de grandeur qui ressortent de l'analyse d'ERN (retraites de 30 ans, portefeuille ~75/25, données mensuelles américaines 1871-2016 ; les seuils exacts varient d'une étude à l'autre, la structure jamais) :

| CAPE au départ | Fréquence historique | SAFEMAX approximatif (30 ans) | Lecture |
|---|---|---|---|
| < 15 (marché bon marché) | ~1/3 du temps | 5,5 à 13 % | La règle des 4 % est très conservatrice |
| 15 à 20 | ~1/4 du temps | ~4,5 à 5,5 % | Le 4 % a de la marge |
| 20 à 30 | ~1/3 du temps | ~3,8 à 4,5 % | Le 4 % passe, sans marge |
| > 30 (cher) | ~1/10 du temps, mais **souvent** ces dernières décennies | ~3,2 à 3,8 % | Le 4 % rigide est en zone d'échec historique |

Et pour les horizons longs du FIRE (50-60 ans), ERN trouve qu'à CAPE > 30, le taux qui aurait survécu à tous les millésimes descend vers **3,0 à 3,25 %**. Autrement dit : la fameuse fourchette moderne « 3,25-3,5 % pour un départ précoce » ([[la-regle-des-4-pourcents]]) n'est **pas** une moyenne tous temps, c'est déjà le chiffre **conditionnel** à un départ en marché cher, c'est-à-dire la situation de la plupart des candidats FIRE actuels. Symétriquement, celui qui part après un grand marché baissier, à CAPE 15, peut légitimement retirer bien davantage : le millésime 1982 supportait plus de 7 %.

Ce résultat a une conséquence conceptuelle profonde : **le « taux de retrait sûr » n'est pas une constante, c'est une fonction du prix d'entrée**. La règle des 4 % moyenne des situations de départ radicalement différentes ; le CAPE permet de dé-moyenner. D'où la génération suivante de règles, dites CAPE-based, dont la forme canonique (ERN volet 54) est :

> taux de retrait = a + b × (1/CAPE)

avec typiquement a ≈ 1,5 à 2 % (la part du retrait que financent la croissance des bénéfices et le reste du portefeuille) et b ≈ 0,4 à 0,5 (la sensibilité à l'earnings yield). Exemple avec a = 1,75 %, b = 0,5 : à CAPE 20, retrait 4,25 % ; à CAPE 33, retrait 3,27 % ; à CAPE 12, retrait 5,9 %. Ces règles, leur comportement dynamique (le taux se recalcule chaque année sur le CAPE et le portefeuille courants, ce qui en fait des cousines disciplinées du pourcentage fixe) et leurs paramètres ont leur article dédié : [[regles-cape]].

::: science Pourquoi la relation est si forte : les trois canaux
Le CAPE de départ agit sur le SAFEMAX par trois canaux qui se cumulent. Canal 1, l'espérance : earnings yield bas = rendement moyen de la décennie décisive plus bas, mécaniquement ([[rendements-arithmetiques-geometriques]]). Canal 2, le risque de retour à la moyenne : partir à CAPE 35, c'est courir le risque supplémentaire que le multiple revienne vers 20, soit -40 % de valorisation à absorber pendant qu'on retire ; partir à CAPE 12 offre le vent inverse. Canal 3, la corrélation avec l'inflation : les grands épisodes de compression de multiple (1966-1982) sont souvent inflationnistes, et l'inflation gonfle les retraits indexés au moment exact où le portefeuille encaisse ([[inflation-et-taux-de-retrait]]). Les trois pires configurations de l'histoire du retrait combinent les trois canaux.
:::

## Les critiques sérieuses du CAPE

Un indicateur aussi utilisé a été attaqué de partout ; la plupart des attaques contiennent une part de vérité, et les connaître évite les deux naïvetés symétriques (l'ignorer, ou le lire au dixième près).

**« Les normes comptables ont changé. »** Vrai. Les bénéfices GAAP d'aujourd'hui ne sont pas ceux de 1950 : dépréciations plus agressives (surtout depuis 2001), traitement des stock-options, part des intangibles passés en charges plutôt qu'immobilisés. Effet net : les bénéfices modernes sont plutôt sous-évalués à méthode constante, donc le CAPE moderne plutôt sur-évalué de quelques points par rapport aux comparaisons centenaires. Jeremy Siegel en a fait sa critique centrale, proposant un CAPE sur bénéfices NIPA (comptabilité nationale) qui ressort structurellement plus bas.

**« Les buybacks faussent la comparaison. »** Partiellement vrai. Les entreprises redistribuent aujourd'hui davantage par rachats d'actions que par dividendes ; à politique de distribution différente, la croissance du bénéfice **par action** est plus rapide qu'avant, ce qui rend le dénominateur moyenné sur dix ans (donc en retard) un peu trop bas, donc le CAPE un peu trop haut. Corrections proposées : le « Total Return CAPE » de Shiller lui-même.

**« Les taux d'intérêt justifient des multiples plus élevés. »** L'argument dominant des années 2010 : à taux réels nuls, l'actualisation des bénéfices futurs justifie des CAPE de 30+. Shiller y a répondu avec l'**Excess CAPE Yield** (ECY) : 1/CAPE moins le taux réel à 10 ans, soit la prime de l'actionnaire sur l'obligataire. L'ECY relativise la cherté des actions **relativement** aux obligations... mais pour un rentier c'est une consolation limitée : un monde où actions et obligations promettent peu (2021, CAPE 38 et taux réels négatifs) est un monde où le taux de retrait soutenable est bas, point ([[rendements-attendus]]). La remontée des taux réels de 2022-2023 a d'ailleurs restauré l'argument : à taux réel 2 %, un CAPE de 35 redevient difficile à justifier.

**« La composition sectorielle a changé. »** Vrai : un indice à 30-40 % de technologie à forte marge et faible intensité capitalistique « mérite » un multiple structurellement plus élevé qu'un indice de conglomérats industriels de 1970. Difficile à quantifier proprement ; ce point justifie surtout de ne pas comparer le niveau absolu d'aujourd'hui aux moyennes d'avant 1990.

**La synthèse pratique de ces critiques** : le CAPE américain moderne est probablement surévalué de 3 à 8 points dans une comparaison centenaire naïve, et sa « moyenne de retour » n'est plus 17 mais plutôt 22-25. **Mais**, et c'est le point décisif pour nous : ces corrections déplacent le niveau, pas la **pente**. Même corrigé, un CAPE à 35 reste dans le quintile cher de sa propre ère, et la relation « plus cher au départ = SAFEMAX plus bas » survit à toutes les corrections publiées. Pour un usage de planification (ordinal, par grandes zones), les critiques commandent l'humilité sur les seuils exacts, pas l'abandon de l'outil.

::: attention Le contresens de la moyenne mobile
Le mésusage le plus répandu : « le CAPE est au-dessus de sa moyenne historique depuis 1991 sauf quelques mois de 2009, donc il est cassé, donc je l'ignore ». Ce raisonnement confond deux usages. Comme signal de **position** (êtes-vous dans le quintile cher de votre époque ?), le CAPE fonctionne toujours : 1999 et 2021 étaient bien des sommets relatifs, 2009 un creux relatif, et les rendements suivants l'ont confirmé. Comme signal de **retour** à une moyenne éternelle de 17, il est effectivement cassé depuis trente ans. Utilisez le rang (percentile dans les 30-40 dernières années), pas l'écart à la moyenne de 1871.
:::

## Les quatre usages dans un plan, du plus sûr au plus risqué

**Usage 1 : calibrer l'espérance de rendement du plan (recommandé, intégré à pofo).** C'est l'usage le plus direct et le moins contestable : puisque 1/CAPE estime la composante centrale du rendement réel actions de la décennie à venir, injectez-le dans le modèle. La case « Anchor return to today's valuation (CAPE) » de la page FIRE fait exactement cela. Elle remplace la seule **moyenne** du modèle central par l'estimation impliquée par le CAPE du jour, en laissant volatilité et queues à leurs valeurs ajustées ([[utiliser-la-page-fire]], [[la-machine-pofo]]). L'effet typique en marché cher : la ruine centrale monte de plusieurs points, ce qui est une information, pas une punition. C'est le prix du point d'entrée rendu visible. Un plan qui ne tient qu'avec l'ancre CAPE décochée est un plan qui parie sur « cette fois c'est différent ».

**Usage 2 : dimensionner le taux initial (recommandé).** Au moment de fixer votre multiple ([[combien-il-vous-faut]]), consultez la zone CAPE : en zone > 30, dimensionnez sur 3-3,5 % rigide ou prévoyez des marges explicites ; en zone < 20 (typiquement, vous partez après un grand marché baissier, félicitations), 4 % et plus se défend historiquement. C'est un usage à grosses mailles, robuste à toutes les critiques ci-dessus.

**Usage 3 : piloter le retrait en continu (les règles CAPE-based, pour les rigoureux).** Le taux de retrait recalculé chaque année en fonction du CAPE courant : plus sophistiqué, avec d'excellentes propriétés (le retrait baisse dans les bulles, remonte dans les creux, la règle est contra-cyclique par construction) et de vraies exigences de discipline. Article dédié ([[regles-cape]]) ; comparaison avec les autres stratégies ([[choisir-sa-strategie]]).

**Usage 4 : moduler la date de départ (à petites doses).** Puisque partir cher est le facteur de risque numéro un, décaler un départ prévu au sommet d'une euphorie de quelques trimestres, ou saisir la fenêtre d'un marché purgé, a une vraie valeur ([[les-trois-phases]], [[une-annee-de-plus]]). La limite est psychologique : le CAPE peut rester cher **dix ans** ; « je partirai quand ce sera moins cher » est une variante du one-more-year sans condition de sortie. Si vous utilisez cet usage, bornez-le : une condition datée (« au plus tard en... ») et un plan B (départ avec revenus partiels, [[retour-au-travail]]).

Et l'usage interdit : **le timing binaire du portefeuille** (tout vendre à CAPE haut, tout racheter à CAPE bas). Toutes les études le confirment : les stratégies de sortie sur signal de valorisation détruisent de la valeur en moyenne, parce qu'elles ratent les fins de bulles (les meilleures années) en échange d'une protection qui arrive trop tôt. Le CAPE règle le **plan** (dépenses, taux, espérances) ; il ne règle pas le portefeuille. La seule exception défendable, douce, est l'allocation glissante autour du départ ([[glidepaths]]), qui peut tenir compte du régime de valorisation.

::: exemple Deux départs, deux mondes
Jumeaux : 1,3 M€, 60/40 mondial, 45 ans d'horizon, mêmes dépenses visées de 45 000 €/an (3,46 %). Amel part en janvier 2000, CAPE 43. Boris part en janvier 2010, CAPE 20. Dans pofo, ancre CAPE active, le modèle central d'Amel tourne avec ~2,3 % de rendement réel actions attendu sur la décennie décisive : ruine centrale au-dessus de 15 %, verdict « pas en l'état ». Les parades chiffrées par le solveur (§09) : baisser à 39 000 € (3 %), ou 1 000 €/mois de revenus d'appoint pendant 8 ans, ou repousser de 30 mois. Boris, à CAPE 20, tourne avec ~5 % attendu : ruine centrale ~4 %, plan validé sans modification. Vingt ans plus tard, l'histoire a tranché exactement dans ce sens : le millésime 2000 américain est passé à un cheveu de la zone rouge et n'a survécu au 4 % que grâce à la décennie 2010 ; le millésime 2009-2010 est l'un des plus opulents du siècle. Deux plans identiques, deux prix d'entrée, deux destins. C'est **toute** l'information que le CAPE apporte, et elle était disponible le jour du départ.
:::

## Le CAPE hors des États-Unis, et le CAPE de votre portefeuille

Presque tout ce qui précède est calibré sur le S&P 500, parce que c'est là que sont les données longues. Trois compléments pour l'investisseur européen en portefeuille mondial ([[etf-ucits-europeens]]).

**Les CAPE nationaux existent** (Barclays-Shiller, Research Affiliates, StarCapital) et la relation valorisation-rendements futurs tient dans tous les marchés étudiés, avec la même pente approximative. Les niveaux, en revanche, ne se comparent pas naïvement d'un pays à l'autre (composition sectorielle, normes comptables, gouvernance) : le Japon a « mérité » des CAPE plus hauts pendant des décennies, l'Europe des CAPE plus bas. Utilisez chaque CAPE contre sa propre histoire.

**Un portefeuille mondial dilue le problème sans le supprimer.** Le marché américain pèse 60-70 % des indices mondiaux : quand il est cher, votre ETF World est cher. La partie non américaine, structurellement moins chère ces dernières années, améliore l'earnings yield agrégé d'un point ou deux : réel, mais pas transformateur. L'ancre CAPE utilise le CAPE américain de Shiller comme proxy prudent de la cherté mondiale, un choix conservateur et assumé ([[la-machine-pofo]]).

**Le CAPE ne dit rien de vos obligations, de votre or, de vos actifs alternatifs.** C'est un indicateur du moteur actions. L'espérance du reste du portefeuille se calibre autrement (taux réels courants pour les obligations, c'est l'un des rares cas où l'espérance est presque littéralement **affichée** sur l'étiquette, [[obligations-en-retrait]], [[rendements-attendus]]).

## L'essentiel à retenir

- Le CAPE = prix / bénéfices réels moyens sur 10 ans : une mesure de cherté qui prédit le **centre** des rendements réels à 10-15 ans (R² ~0,4-0,6 aux États-Unis), pas leur calendrier ni leurs krachs.
- Tous les pires millésimes de retraite partent à CAPE élevé : le « taux de retrait sûr » est une fonction du prix d'entrée ; à CAPE > 30 et horizon long, la zone historique est 3,0-3,25 % rigide, à CAPE < 15 elle dépasse 5 %.
- Les critiques (comptabilité, buybacks, taux, secteurs) déplacent les seuils, pas la pente : usage ordinal et par zones, jamais au dixième ; lisez le rang dans les 30-40 dernières années, pas l'écart à la moyenne de 1871.
- Quatre usages légitimes, par ordre de sûreté : calibrer l'espérance du modèle (l'ancre CAPE), dimensionner le taux initial, piloter le retrait ([[regles-cape]]), moduler (un peu, avec une borne datée) la date de départ. Un usage interdit : le timing binaire du portefeuille.
- Le réflexe de session : regardez le §00 de la page FIRE, cochez l'ancre CAPE, et si le plan ne tient plus, c'est le solveur §09 qui vous dit le prix de votre point d'entrée, en euros, en années ou en flexibilité.

---

## Pour aller plus loin

- Campbell & Shiller, « Stock Prices, Earnings, and Expected Dividends » (1988) ; Shiller, *Irrational Exuberance* (2000, rééditions avec l'ECY) : les sources.
- Les données : le site de Robert Shiller (Yale) publie la série CAPE 1871-aujourd'hui, mise à jour mensuellement (c'est la série utilisée ici, complétée en continu via [multpl.com](https://www.multpl.com)).
- Early Retirement Now, SWR Series volet 3 (CAPE et SAFEMAX), volet 18 et volet 54 (les règles CAPE-based) : la formalisation pour rentiers ([[serie-ern]]).
- Michael Kitces, « Resolving the Paradox: Is the Safe Withdrawal Rate Sometimes Too Safe? » (2008) : le lien valorisation-retrait côté praticien.
- Wade Pfau, « Can We Predict the Sustainable Withdrawal Rate for New Retirees? » (2011) : la régression SAFEMAX ~ CAPE.
- Research Affiliates (Asset Allocation Interactive) et le Global Investment Returns Yearbook : les CAPE et espérances par pays, mis à jour.
- La suite dans ce livre : [[regles-cape]] (l'usage dynamique), [[rendements-attendus]] (les espérances prospectives toutes classes d'actifs) et [[rendre-monte-carlo-pertinent]] (comment une espérance comprimée entre dans un modèle).
