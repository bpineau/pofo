# VPW, le retrait à pourcentage variable des Bogleheads

Le VPW (« Variable Percentage Withdrawal ») est la réponse de la communauté Bogleheads à une question précise : comment consommer un portefeuille **sans** prévision, **sans** risque de ruine, et **sans** mourir sur un tas d'or ? Sa solution : un pourcentage du portefeuille, comme le pourcentage fixe ([[pourcentage-fixe]]), mais qui **augmente** avec l'âge selon une table calculée une fois pour toutes : ~3,8 % à 40 ans, ~4,8 % à 65 ans, ~7 % à 80 ans, jusqu'à 100 % la dernière année.

Ce pourcentage croissant n'est pas un bricolage : c'est la formule d'amortissement d'un prêt, appliquée à l'envers à votre portefeuille : le VPW est le pont historique entre la famille proportionnelle et la famille actuarielle ([[amortissement-abw]]), et l'une des stratégies les plus utilisées du monde FIRE réel. Cet article détaille sa mécanique exacte (vous saurez recalculer la table), ses choix de conception assumés et discutables (les rendements supposés **fixes**), ses propriétés remarquables, sa pathologie (la même volatilité de train de vie que tout pourcentage, plus une bosse de fin de vie), le garde-fou que ses auteurs eux-mêmes imposent (le test de tolérance à la perte, trop souvent sauté), et sa place face à l'ABW moderne et dans pofo.

::: cle L'idée en une phrase
Chaque année, le VPW retire le pourcentage qui épuiserait exactement le portefeuille sur les années restantes si les rendements futurs étaient égaux à des valeurs de référence fixées d'avance : c'est la mensualité d'un crédit, recalculée chaque année sur le capital courant et l'horizon restant. Jeune, l'horizon est long, le pourcentage est bas ; âgé, l'horizon raccourcit, le pourcentage monte : le portefeuille est consommé délibérément, jamais épuisé prématurément, jamais thésaurisé par accident.
:::

## D'où ça vient, et la philosophie

Le VPW naît sur le forum Bogleheads au début des années 2010, œuvre du contributeur « longinvest », avec une doctrine en trois refus très boglehead : refus de la ruine (donc pas de montant fixe, [[retrait-fixe-bengen]]) ; refus du legs accidentel (le 4 % prudent meurt riche trois fois sur quatre, de la vie non vécue, [[depenses-en-retraite]]) ; et refus de la **prévision** : pas de rendements attendus recalculés chaque année, pas de CAPE, pas de paramètres à débattre : une table unique, publiée, gravée. La stratégie s'accompagne d'un classeur (« VPW worksheet ») maintenu par la communauté, qui gère aussi les ponts de pension (nous y revenons) : c'est l'un des outils gratuits les plus aboutis du monde FIRE.

## La mécanique exacte : la formule du crédit inversée

Le cœur est la formule d'annuité, celle de toutes les mensualités de prêt. Pour un capital C, un horizon de n années restantes et un taux de croissance supposé g, le paiement constant qui épuise exactement C en n ans est :

> retrait = C × g / (1 − (1 + g)^(−n))

Le VPW tabule ce ratio (retrait / capital) pour chaque âge, avec n courant de l'âge à 100 ans, et g fixé **une fois** par classe d'actifs : dans la table actuelle, 5,0 % réel pour les actions mondiales et 1,9 % réel pour les obligations, combinés au prorata de votre allocation (un 60/40 suppose donc ~3,8 % réel). Extrait de la logique de la table pour un 60/40 :

| Âge | Années restantes (jusqu'à 100) | Pourcentage VPW approximatif |
|---|---|---|
| 40 | 60 | ~3,9 % |
| 50 | 50 | ~4,1 % |
| 65 | 35 | ~4,8 % |
| 75 | 25 | ~5,7 % |
| 85 | 15 | ~7,9 % |
| 99 | 1 | 100 % |

Deux propriétés de la formule méritent l'arrêt. D'abord, à horizon long, le pourcentage tend vers g lui-même (à 60 ans d'horizon, on retire à peine plus que la croissance supposée, le capital est quasi préservé) : le VPW d'un FIRE de 40 ans est en pratique un pourcentage fixe amélioré, et sa montée en âge ne devient sensible qu'après 65-70 ans. Ensuite, la montée finale est la **consommation délibérée** du capital : c'est un choix de conception (mourir à zéro à 100 ans), pas un accident, et il appelle nécessairement un traitement du risque de longévité : la doctrine VPW elle-même recommande d'annuitiser une part du portefeuille vers 80 ans pour couvrir les années au-delà de la table ([[rentes-et-annuites]], [[horizon-et-esperance-de-vie]]).

**Le pont de pension**, l'autre innovation pratique du classeur : avant la liquidation de vos pensions ([[retraite-legale]]), le VPW met de côté (virtuellement) le capital nécessaire pour « fabriquer » la pension manquante pendant les années de pont (par exemple 15 ans × 15 000 € pour une pension à 67 ans), l'investit en obligations, et n'applique le pourcentage qu'au reste. C'est exactement la décomposition phase à découvert / phase adossée de [[horizon-et-esperance-de-vie]], rendue opérationnelle : le besoin permanent est amorti, le besoin temporaire est provisionné.

## Ce que le VPW réussit, et ce qu'il assume de rater

**Les réussites.** Toutes les vertus du pourcentage ([[pourcentage-fixe]]) : ruine du capital impossible, contra-cyclicité, auto-correction face aux erreurs de rendement ; **plus** la conscience de l'horizon : là où le pourcentage fixe thésaurise éternellement, le VPW ose consommer, et sa consommation totale moyenne sur la vie du plan est parmi les plus élevées de toutes les règles : c'est la stratégie anti-« mourir riche » par excellence. Et une gouvernance remarquable : une table imprimée, un ratio par an, pas un paramètre à rediscuter : la règle survit à son auteur et aux années de panique ([[psychologie-du-retrait]]).

**Les ratages assumés.** Le premier est celui de toute la famille : le revenu suit le portefeuille. Le VPW ne lisse **rien** par construction (la doctrine refuse le lissage comme une dette déguisée) : −30 % de portefeuille = −30 % de revenu l'année suivante. D'où le garde-fou que la doctrine impose et que tout le monde saute : le **test de tolérance à la perte** : avant d'adopter le VPW, calculez votre revenu dans l'hypothèse « actions −50 % » (le classeur l'affiche en permanence) et vérifiez qu'il couvre encore votre plancher ([[combien-il-vous-faut]]). Si non, le VPW vous dit lui-même de réduire la part actions ou de couvrir le plancher autrement (pension, rente) : c'est une stratégie qui exige un plancher externe ou une vraie élasticité, exactement comme son parent proportionnel.

**Le ratage discutable : les rendements supposés fixes.** Le g de la table (5 % réel actions) est une moyenne historique de très long terme, la même en 2013, en 2021 (CAPE 38 !) et en 2026. C'est un choix philosophique cohérent (pas de prévision) mais coûteux en marché cher : le VPW retire alors davantage que ce que les valorisations promettent ([[valorisations-et-cape]], [[rendements-attendus]]), et l'ajustement se fera ex post, par la baisse du revenu quand la déception arrivera. La famille ABW/TPAW fait le choix inverse (brancher les rendements attendus courants, CAPE compris) : plus juste en espérance, plus dépendant des modèles : c'est la ligne de partage entre les deux cousins ([[amortissement-abw]]).

::: science VPW et ABW : la même formule, deux épistémologies
Mathématiquement, VPW = ABW à rendements supposés constants et sans valeur actualisée fine des flux futurs : la même annuité inversée. La divergence est épistémologique : le VPW parie que le retraité moyen se trompera moins avec une table gravée qu'avec des prévisions annuelles (robustesse comportementale) ; l'ABW parie que l'information des prix courants vaut mieux qu'une moyenne séculaire (justesse conditionnelle). La recherche penche pour l'ABW sur les chiffres... sur des simulations où la règle est appliquée sans faille ; la sagesse des forums penche pour le VPW chez les humains réels. Le choix honnête dépend de qui exécutera la règle dans vingt ans ([[couple-et-famille]], [[choisir-sa-strategie]]).
:::

## Pour qui, et les réglages qui comptent

Le profil idéal du VPW cumule trois traits : un **plancher** couvert hors portefeuille (pension présente ou pontée, rente, le test de tolérance à la perte passe alors naturellement) ; une vraie élasticité de train de vie au-dessus du plancher ; et un goût pour la simplicité auditable (la table, le classeur, rien d'autre). C'est très exactement le retraité Bogleheads type, et c'est aussi la phase **adossée** d'un plan FIRE français ([[horizon-et-esperance-de-vie]]) : après 65-67 ans, pension au plancher, le VPW sur le portefeuille résiduel est difficile à battre.

En phase à découvert d'un FIRE précoce, il demande deux aménagements : le pont de pension du classeur (obligatoire, sans lui, le pourcentage s'applique à un capital qui doit **aussi** fabriquer quinze ans de pension, et le test de perte échoue presque toujours) ; et, en marché cher, une décote manuelle de g (utiliser 4 % au lieu de 5 % pour les actions revient à intégrer grossièrement l'ancre CAPE sans trahir l'esprit de la table).

**Dans pofo** : le curseur « Spend % of portfolio (VPW) » applique un pourcentage **constant** : c'est le VPW d'un horizon encore long (la zone plate de la table), l'approximation est excellente avant 60-65 ans. Pour la dynamique complète (pourcentage croissant, horizon exact, flux futurs actualisés), la case « Amortize over the horizon (ABW/TPAW) » est la généralisation : cochez-la pour voir votre plan sous l'annuité inversée intégrale, rendements centraux (et CAPE si l'ancre est active) : c'est le VPW moderne ([[utiliser-la-page-fire]]). La frontière §06 positionne les deux contre le reste ; la §04 montre, comme toujours, la vie vécue.

::: exemple Le VPW d'un couple FIRE, pont compris
Nora et Malik, 47 ans, 1,6 M€, 60/40, pensions estimées 21 600 €/an à 67 ans, plancher 38 000 €, confort 52 000 €. Pont de pension : 20 ans × 21 600 € ≈ 380 000 € provisionnés en obligations (le classeur affine avec l'actualisation, ~350 000 €). VPW sur le reste (1 250 000 €, 47 ans, table 60/40) : ~4,0 % = 50 000 €, **plus** la tranche de pont annuelle ~19 000 € les premières années... le classeur agrège tout : revenu initial ~54 000 €. Test de tolérance : actions −50 % : portefeuille VPW → ~875 000 €, revenu → ~40 000 € : au-dessus du plancher, le plan passe (grâce au pont, sans lui, il échouait). Vingt ans plus tard : pensions liquidées au plancher, VPW à 5,2 % sur le résiduel pour le confort et les projets. Deux régimes, une seule table, zéro prévision : c'est le VPW bien construit.
:::

## L'essentiel à retenir

- VPW = l'annuité d'un crédit inversée : chaque année, le pourcentage (croissant avec l'âge, ~3,9 % à 40 ans, ~4,8 % à 65, 100 % à 99) qui épuiserait le portefeuille sur les années restantes à rendements supposés **fixes** (5 % réel actions, 1,9 % obligations).
- Il hérite du pourcentage fixe (jamais de ruine du capital, contra-cyclique, auto-correcteur) et y ajoute la conscience de l'horizon : consommation délibérée, legs quasi nul, la consommation moyenne la plus généreuse du panorama.
- Ses exigences : le **test de tolérance** à **la perte** (revenu sous « actions −50 % » ≥ plancher, ne le sautez jamais), le pont de pension en phase à découvert, et l'annuitisation vers 80 ans pour la longévité au-delà de la table.
- Sa ligne de partage avec l'ABW : table gravée (robustesse comportementale) contre rendements courants (justesse conditionnelle) : même formule, deux paris sur l'exécutant.
- Dans pofo : le curseur % (VPW à horizon long) et la case ABW (le VPW dynamique complet) ; jugez sur la §04 et la frontière, et en marché cher, décotez g d'un point.

---

## Pour aller plus loin

- Bogleheads wiki, « Variable percentage withdrawal (VPW) » et le fil « VPW forward test » du forum : la doctrine, la table, le classeur, et dix ans d'exécution documentée en conditions réelles.
- Le VPW worksheet (classeur officiel, gratuit) : ponts de pension et test de perte intégrés.
- Early Retirement Now, volet 11 (le VPW noté contre les autres règles) ([[serie-ern]]).
- Dans ce livre : [[pourcentage-fixe]] (le parent), [[amortissement-abw]] (le cousin moderne), [[rentes-et-annuites]] (le complément de fin de vie), [[choisir-sa-strategie]] (l'arbitrage final).
