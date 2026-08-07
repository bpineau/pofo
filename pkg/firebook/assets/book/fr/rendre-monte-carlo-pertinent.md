# Rendre un Monte-Carlo pertinent (blending, régimes, stress)

Un Monte-Carlo naïf (tirages gaussiens indépendants, paramètres copiés de l'historique récent, plan caricatural) est pire qu'inutile. Il est convaincant. Il produit de beaux cônes et des probabilités à trois décimales à partir d'hypothèses qui sous-estiment systématiquement le risque d'un rentier ([[pieges-des-simulateurs]]).

La bonne nouvelle : chacun de ses défauts a une correction connue, documentée par la recherche, et implémentable. Cette page présente les six corrections qui transforment le générateur de nombres en instrument de planification, dans l'ordre où elles s'appliquent : calibrer les entrées sans hériter du biais de sa propre fenêtre (le « blending », l'idée la plus importante et la moins connue), épaissir les queues, réintroduire la mémoire des marchés (les régimes), ancrer aux valorisations, confronter aux données brutes, et simuler le vrai plan plutôt que sa caricature.

C'est très exactement la liste de construction du modèle central, et cette page sert donc aussi de justification de conception : pourquoi la page FIRE calcule ce qu'elle calcule ([[la-machine-pofo]] donne la plomberie ; ici, le pourquoi).

::: cle Le principe directeur
Il n'existe pas de « meilleur modèle » ; il existe un modèle **central** honnête (calibré, corrigé, tiré vers la prudence là où l'information manque) et des **bornes** qui l'encadrent (l'optimiste, vos données rejouées ; les pessimistes, stress, siècle mondial, décennie perdue). Rendre un Monte-Carlo pertinent, ce n'est pas trouver la vérité. C'est construire ce faisceau, et décider dedans. Toute la suite détaille la fabrication du faisceau.
:::

## Correction 1 : le blending, ou comment ne pas croire sa propre fenêtre

Le problème fondamental de la calibration : les paramètres de votre portefeuille (μ, σ, df) ne peuvent s'estimer que sur son historique, disons vingt ans. Or vingt ans, statistiquement, c'est presque rien pour une moyenne. L'erreur type sur μ vaut environ σ/√n, soit ±2,5 points avec σ = 11 %. Autrement dit, l'estimation « 6,8 % réel » signifie « quelque part entre 4 et 9 % ». Et ce n'est pas qu'imprécis, c'est **biaisé** : votre fenêtre est celle qui vous a mené à la cible, donc probablement favorable ; elle ne contient qu'un ou deux régimes de marché sur les quatre possibles ([[regimes-de-marche]]).

La réponse statistique classique s'appelle le rétrécissement (shrinkage), l'intuition de James-Stein. Quand une estimation individuelle est bruitée, on améliore toujours la prévision en la tirant vers une moyenne de référence plus large. Ici, cela revient à mélanger les paramètres ajustés sur vos fonds avec un **prior** issu d'un échantillon immensément plus profond, l'expérience mondiale de long terme (μ ≈ 4,5 % arithmétique réel, σ ≈ 13 %, df ≈ 4, les valeurs prudentes que suggère le siècle développé, [[anarkulova-cederburg]]).

Reste à choisir le poids du mélange, et la page FIRE applique ici une règle simple : **le poids du prior croît avec ce que l'horizon dépasse l'historique**, plafonné à 50/50. La logique est directe. Si vous avez 20 ans de données pour un plan de 20 ans, vos données parlent d'expérience. Pour un plan de 45 ans, elles n'ont rien observé des 25 années au-delà de leur fenêtre, et c'est le prior qui doit parler pour l'inconnu. Concrètement, avec 20 ans d'historique et 45 ans d'horizon, le mélange est au plafond : moitié vos fonds, moitié le siècle. Votre portefeuille atteint alors le modèle central par ses statistiques, jamais par sa séquence particulière. Ses vertus mesurables (diversification, volatilité contenue) sont créditées, mais pas sa chance de fenêtre.

Deux remarques d'usage. D'abord, le blending s'applique à μ, σ et df : la prudence porte aussi sur les queues. Ensuite, il est automatique mais pas tyrannique. Les curseurs restent les maîtres, et vous pouvez saisir à la main votre μ par briques (building-blocks, [[rendements-attendus]]). Le blending n'est que le défaut raisonnable pour qui ne veut pas trancher.

## Corrections 2 et 3 : les queues, puis la mémoire

**Les queues** : remplacer la gaussienne par une Student-t dont le df est ajusté sur le kurtosis mensuel de vos fonds ([[queues-epaisses]]). La correction est locale, à un seul paramètre, et son effet est direct. Les années catastrophiques retrouvent leur vraie fréquence dans le modèle central, pas seulement dans un scénario de stress annexe. C'est un choix de conception important : beaucoup d'outils gardent un central gaussien « pour la lisibilité » et relèguent les queues dans un mode expert que personne n'ouvre.

**La mémoire** : le central corrigé reste i.i.d. Ses mauvaises années tombent au hasard, jamais en grappes, alors que les vraies traversées difficiles sont des épisodes (2000-2002, 2007-2009, 1973-1974, et leurs versions longues). D'où la colonne « Sequence stress » : une chaîne de Markov à deux états, normal et « baissier », construite avec trois propriétés précises. Un, les marchés baissiers persistent : y entrer est rare, y rester probable. On obtient ainsi des épisodes d'environ trois ans et ~19 % d'années de marché baissier au total, cohérent avec l'histoire. Deux, la volatilité est amplifiée dans le marché baissier (× 1,5) : les crises sont agitées, pas seulement baissières, et c'est là que loge l'asymétrie négative (skew) observée dans les faits ([[queues-epaisses]]). Trois, et c'est le point méthodologique décisif : le modèle est à **moyenne préservée**, sa moyenne de long terme étant exactement celle du central. Le stress ne cache donc aucun pessimisme sur le niveau. Il isole chirurgicalement le risque d'**ordre**. L'écart de ruine entre les colonnes central et stress est, par construction, le prix de la séquence dans votre plan, et rien d'autre ([[sequence-des-rendements]]). C'est une expérience contrôlée, impossible à mener avec des données réelles.

La variante extrême complète l'arsenal : « Lost decade », un régime de marché baissier long et profond à la manière du Japon des années 1990, délibérément non compensé (la moyenne est dégradée). Ce n'est plus une estimation, c'est un crash-test. La question qu'on lui pose n'est pas « quelle est ma ruine ? » mais « peut-on y survivre ? », et la réponse attendue n'est pas un pourcentage confortable mais un plan de traversée ([[marche-baissier-en-retraite]]).

## Correction 4 : les ancres, ou l'information du présent

Le blending corrige le passé, c'est-à-dire votre fenêtre. Il reste aveugle au présent : où en sont les valorisations aujourd'hui ? Un même portefeuille n'a pas la même espérance à CAPE 20 et à CAPE 38 ([[valorisations-et-cape]]), et aucune moyenne historique, si bien mélangée soit-elle, ne porte cette information. D'où les deux ancres de la page FIRE, deux boutons qui réécrivent les paramètres avec une information extérieure :

**L'ancre CAPE** remplace la seule moyenne du central par l'estimation qu'implique le CAPE du jour (~1/CAPE pour la brique actions), en laissant σ et df à leurs valeurs ajustées. C'est la correction prospective : le modèle central cesse de supposer que la décennie décisive ressemblera à la moyenne des décennies, et suppose qu'elle ressemblera à ce que les prix actuels permettent. En marché cher, elle est la plus dure des corrections. C'est normal, c'est elle qui porte la mauvaise nouvelle.

**Le prior broad-sample** (« Broad-sample prior ») réécrit les trois curseurs d'un coup avec les valeurs prudentes du siècle mondial : l'équivalent d'un blending poussé à 100 %. Utile comme borne, ou comme position par défaut pour qui se méfie de tout ajustement sur données récentes.

Reste le garde-fou déjà posé ailleurs, mais qui prend ici toute sa portée : ces mécanismes **ne s'empilent pas** ([[rendements-attendus]]). Blending automatique, μ manuel, ancre CAPE et prior broad-sample sont quatre façons de calibrer le même modèle central. On en choisit une, les autres servent de contre-lectures. La prudence est un budget, pas une vertu cumulable.

## Corrections 5 et 6 : les données brutes, et le vrai plan

**Correction 5 : garder les données en juges.** Toutes les corrections précédentes raffinent un modèle paramétrique, et le risque ultime serait de ne plus regarder que lui. D'où la règle de conception : les modèles de données (fenêtres historiques, bootstrap de vos fonds, broad-sample du siècle) restent affichés en permanence, insensibles aux curseurs, dans le même tableau ([[historique-vs-parametrique]]). Ils jouent le rôle de la réalité dans la pièce. Si le central corrigé s'écarte beaucoup des fenêtres historiques, l'écart rend visible le biais de votre fenêtre. S'il s'écarte du broad-sample, c'est votre pari sur « le futur sera plus doux que le siècle » qui devient explicite. Le désaccord entre colonnes n'est pas un défaut à résoudre, c'est le produit fini ([[ruine-et-probabilites]]).

**Correction 6 : simuler le plan réel.** La dernière correction ne porte pas sur le marché, mais sur vous. Un moteur parfait appliqué à un plan caricatural (retrait rigide éternel, sans pension, sans impôts, sans buffer) produit des chiffres sans objet. Le réalisme du plan compte autant que celui du marché : la pension à sa date ([[retraite-legale]]), les revenus d'appoint des premières années, la fiscalité qui majore chaque vente, le buffer avec ses règles de consommation et de recharge ([[recharger-ou-pas]]), et surtout votre règle de dépense (flex, guardrails avec plancher, VPW, ABW, [[panorama-strategies-retrait]]). Chacun de ces éléments déplace la ruine de plusieurs points, soit davantage que bien des débats de modélisation. C'est la force distinctive de la simulation ([[monte-carlo-forces-faiblesses]]) : cette complexité-là est gratuite, et ne pas s'en servir est un gâchis.

::: science Ce que dit la littérature de chaque correction
Chaque correction a sa généalogie académique : le shrinkage des estimations de rendement remonte à James-Stein (1961) et à son application en finance par Jorion (1986, « Bayes-Stein estimation ») ; les queues Student-t à Praetz (1972) et Blattberg-Gonedes (1974) ; les régimes de Markov à Hamilton (1989), devenus l'outil standard des marchés baissiers persistants ; l'ancrage aux valorisations à Campbell-Shiller (1988) et sa version taux-de-retrait à Kitces (2008) et ERN (volet 54) ; le bootstrap par blocs à Politis-Romano (1994), appliqué au problème du rentier par Anarkulova-Cederburg (2023). Rien dans cette liste n'est exotique. C'est la boîte à outils normale de l'économétrie financière, simplement appliquée avec constance à un problème que les outils grand public traitent avec les moyens de 1998.
:::

## Le tout assemblé : une montée en pertinence, chiffrée

Suivons un même plan à travers les corrections, pour voir ce que chacune coûte ou révèle. Plan : 1,4 M€, 50 000 €/an, 45 ans, portefeuille mondial avec 18 ans d'historique favorable (μ ajusté brut, 6,5 % réel, σ 11 %, kurtosis mensuel 8).

| Étape | Modèle | Ruine | Ce que l'étape apporte |
|---|---|---|---|
| 0 | Gaussien i.i.d., μ historique brut, plan rigide sans pension | ~1 % | Le chiffre de brochure |
| 1 | + blending vers le prior mondial (μ 5,5 %, df 4 au plafond 50/50) | ~4 % | La fenêtre favorable cesse de faire loi |
| 2 | + Student-t (df ajusté ≈ 5) | ~5,5 % | Les catastrophes retrouvent leur fréquence |
| 3 | + ancre CAPE (marché cher : μ actions ramené vers ~3 %) | ~9 % | Le présent entre dans le modèle |
| 4 | Lecture stress (moyenne préservée, sticky bears) | ~12 % | Le prix de l'ordre : +3 points |
| 5 | Lecture broad-sample | ~13 % | Le siècle confirme la zone |
| 6 | + le vrai plan : pension 14 k€ en année 16, flexibilité écrite −10 % | central ~3,5 %, stress ~6 % | Le réalisme du plan rend ce que la rigueur du modèle avait retiré |

La trajectoire 1 % → 9 % → 3,5 % raconte toute la philosophie : le naïf flattait (1 %), la rigueur dégrise (9-13 %), et le réalisme d'un plan complet redonne des marges honnêtes (3,5-6 %). Le chiffre final ressemble au chiffre naïf, mais il n'a plus rien à voir. Il a été gagné contre les pièges, pas obtenu par eux, et l'on sait exactement quelles hypothèses le portent et quelles marges le défendent.

## L'essentiel à retenir

- Un Monte-Carlo devient pertinent par six corrections ordonnées : blending des paramètres vers un prior mondial (l'antidote au biais de sa propre fenêtre, dosé par le déficit d'historique face à l'horizon), queues Student-t ajustées, régimes à moyenne préservée pour la mémoire (le stress mesure le prix de l'ordre, rien d'autre), ancres au présent (CAPE), données brutes gardées en juges, et simulation du plan réel.
- Le produit fini n'est pas un chiffre mais un faisceau : un central honnête encadré de bornes ; on décide dans le faisceau, sur les colonnes dures.
- Les calibrations ne s'empilent pas : blending ou μ manuel ou ancre CAPE comme central, les autres en contre-lectures ; la prudence cumulée en triple couche se paie en années de travail.
- Le réalisme du plan (pension, revenus, fiscalité, règles de dépense) pèse souvent plus que les raffinements de marché. C'est la complexité gratuite de la simulation, servez-vous-en.
- Chaque correction a trente ans de littérature derrière elle ; la seule originalité est de les appliquer **toutes**, par défaut, dans un outil grand public.

---

## Pour aller plus loin

- Jorion, « Bayes-Stein Estimation for Portfolio Analysis » (1986) : le shrinkage appliqué aux rendements.
- Hamilton, « A New Approach to the Economic Analysis of Nonstationary Time Series » (1989) : les régimes de Markov.
- Early Retirement Now, volet 54 (les règles CAPE) et volet 15 (la mesure du risque de séquence) ([[serie-ern]]).
- Le volet « How this machine works » de la page FIRE : chaque correction, curseur par curseur ([[utiliser-la-page-fire]]) ; et [[la-machine-pofo]] pour l'implémentation.
- La suite : [[regimes-de-marche]] (le fondement économique des sticky bears et des quatre saisons macro).
