# VPW, le retrait à pourcentage variable des Bogleheads

Le VPW (« Variable Percentage Withdrawal ») est la réponse de la communauté Bogleheads à une question précise : comment consommer un portefeuille sans prévision, sans risque de ruine, et sans mourir sur un tas d'or ? Sa solution tient en une idée. On retire un pourcentage du portefeuille, comme le pourcentage fixe ([[pourcentage-fixe]]), mais ce pourcentage augmente avec l'âge selon une table calculée une fois pour toutes. Il vaut environ 3,9 % à 40 ans, 4,8 % à 65 ans, 7 % à 80 ans, jusqu'à 100 % la dernière année.

Ce pourcentage croissant n'est pas un bricolage. C'est la formule d'amortissement d'un prêt, appliquée à l'envers à votre portefeuille. Le VPW est le pont historique entre la famille proportionnelle et la famille actuarielle ([[amortissement-abw]]), et l'une des stratégies les plus utilisées du monde FIRE réel. Cet article en détaille la mécanique exacte, de quoi recalculer la table vous-même. Il expose ses choix de conception assumés et discutables, à commencer par des rendements supposés fixes. Il décrit ses propriétés remarquables et sa pathologie, la même volatilité de train de vie que tout pourcentage, plus une bosse de fin de vie. Il présente enfin le garde-fou que ses auteurs imposent, le test de tolérance à la perte, trop souvent sauté, puis sa place face à l'ABW moderne et la façon de l'éprouver dans un simulateur.

::: cle L'idée en une phrase
Chaque année, le VPW retire le pourcentage qui épuiserait exactement le portefeuille sur les années restantes, en supposant les rendements futurs égaux à des valeurs de référence fixées d'avance. C'est la mensualité d'un crédit, recalculée chaque année sur le capital courant et l'horizon restant. Jeune, l'horizon est long et le pourcentage est bas. Âgé, l'horizon raccourcit et le pourcentage monte. Le portefeuille est consommé délibérément, jamais épuisé prématurément, jamais thésaurisé par accident.
:::

::: admin Mode d'emploi
- **Le taux s'applique au portefeuille courant diminué du pont de pension.** Le pont, c'est-à-dire les annuités de pension manquantes actualisées et placées en obligations, se met de côté **avant** d'appliquer le pourcentage. C'est l'erreur d'implémentation la plus fréquente du VPW : appliquer la table au portefeuille entier, alors que celui-ci doit aussi fabriquer quinze ou vingt ans de pension. Le reste est un pourcentage ordinaire, sans mémoire du capital de départ.
- **Indexation.** Aucune ligne à écrire, comme pour tout pourcentage du portefeuille courant.
- **Fréquence.** Annuelle : une lecture de table, une multiplication.
- **Paramètres.** Ce ne sont pas des seuils mais deux nombres. Le rendement supposé g (5,0 % réel pour les actions, 1,9 % pour les obligations, combinés au prorata de votre allocation) et l'horizon n (100 moins votre âge). Ils sont gravés par doctrine, ce qui est un choix assumé et non une négligence : la table ne se renégocie pas. La seule modulation défendable est de décoter g d'un point en marché cher.
- **Plancher.** Externe et obligatoire, vérifié par le test de tolérance à la perte : le revenu servi sous l'hypothèse « actions −50 % » doit rester au-dessus du plancher.
- **De tête.** w = g / (1 − (1+g) puissance −n), c'est-à-dire la mensualité d'un crédit. Une précision d'implémentation vaut d'être connue, car elle explique la plupart des écarts entre deux tables : selon que le retrait est pris en début ou en fin d'année, on divise ou non le résultat par (1 + g), ce qui déplace le taux d'environ 4 % en relatif. En pratique, lisez la table officielle plutôt que de la recalculer, mais savoir d'où elle sort permet de vérifier un chiffre douteux.
:::

## D'où ça vient, et la philosophie

Le VPW naît sur le forum Bogleheads au début des années 2010, œuvre du contributeur « longinvest ». Sa doctrine tient en trois refus, très dans l'esprit boglehead. Refus de la ruine, d'abord, donc pas de montant fixe ([[retrait-fixe-bengen]]). Refus du legs accidentel, ensuite, car le 4 % prudent meurt riche trois fois sur quatre, faute d'avoir vécu ([[depenses-en-retraite]]). Refus de la prévision, enfin : pas de rendements attendus recalculés chaque année, pas de CAPE, pas de paramètres à débattre, seulement une table unique, publiée, gravée. La stratégie s'accompagne d'un classeur (« VPW worksheet ») maintenu par la communauté, qui gère aussi les ponts de pension, nous y reviendrons. C'est l'un des outils gratuits les plus aboutis du monde FIRE.

## La mécanique exacte : la formule du crédit inversée

Le cœur est la formule d'annuité, celle de toutes les mensualités de prêt. Pour un capital C, un horizon de n années restantes et un taux de croissance supposé g, le paiement constant qui épuise exactement C en n ans est :

> retrait = C × g / (1 − (1 + g)^(−n))

Le VPW tabule ce ratio (retrait / capital) pour chaque âge, avec n qui court de l'âge courant à 100 ans. Le taux g, lui, est fixé une fois pour toutes par classe d'actifs. La table actuelle retient 5,0 % réel pour les actions mondiales et 1,9 % réel pour les obligations, combinés au prorata de votre allocation. Un 60/40 suppose donc environ 3,8 % réel. Voici un extrait de la logique de la table pour un 60/40 :

| Âge | Années restantes | % VPW |
|---|---|---|
| 40 | 60 | ~3,9 % |
| 50 | 50 | ~4,1 % |
| 65 | 35 | ~4,8 % |
| 75 | 25 | ~5,7 % |
| 85 | 15 | ~7,9 % |
| 99 | 1 | 100 % |

::: figure vpw-table
En haut, la table : le taux part juste au-dessus du rendement supposé, reste presque plat pendant vingt-cinq ans, puis s'envole quand l'horizon se raccourcit. En bas, la vie que cette table produit pour un ménage parti à 40 ans avec 1 M€, si le marché sert exactement les rendements supposés. Le revenu est parfaitement plat, et le capital fond jusqu'à zéro à 100 ans : c'est la propriété de l'annuité, pas un hasard. La ligne pointillée montre l'autre visage de la règle. Un krach de 30 % à 70 ans fait descendre le revenu de 38,5 à 27,0 k€ **et il n'y revient pas**, car le VPW ne lisse rien : il recalcule.
:::

Deux propriétés de la formule méritent qu'on s'y arrête. D'abord, à horizon long, le pourcentage tend vers g lui-même. À 60 ans d'horizon, on retire à peine plus que la croissance supposée et le capital est quasi préservé. Le VPW d'un FIRE de 40 ans est donc, en pratique, un pourcentage fixe amélioré, et sa montée en âge ne devient sensible qu'après 65-70 ans. Ensuite, la montée finale est la consommation délibérée du capital. C'est un choix de conception, mourir à zéro à 100 ans, et non un accident. Il appelle donc un traitement du risque de longévité : la doctrine VPW recommande d'annuitiser une part du portefeuille vers 80 ans, pour couvrir les années au-delà de la table ([[rentes-et-annuites]], [[horizon-et-esperance-de-vie]]).

**Le pont de pension** est l'autre innovation pratique du classeur. Avant la liquidation de vos pensions ([[retraite-legale]]), le VPW met de côté, virtuellement, le capital nécessaire pour « fabriquer » la pension manquante pendant les années de pont, par exemple 15 ans × 15 000 € pour une pension à 67 ans. Il l'investit en obligations et n'applique le pourcentage qu'au reste. C'est la décomposition phase à découvert / phase adossée de [[horizon-et-esperance-de-vie]], rendue opérationnelle : le besoin permanent est amorti, le besoin temporaire est provisionné.

::: figure vpw-pont
Le pont, sur le ménage de l'exemple ci-dessous. La bande basse est le revenu qui n'existe pas encore : vingt tranches de 21,6 k€ prélevées sur les 356 k€ d'obligations mises de côté, puis la pension elle-même, qui prend le relais sans que le ménage voie la différence. La bande haute est le VPW appliqué au reste. **Le bord supérieur est plat : à 67 ans, le revenu ne fait pas de marche, il change simplement de main.** Sans le pont, le même ménage vivrait vingt ans à 64,3 k€ avant de sauter à 85,9 k€, c'est-à-dire trop peu quand il est jeune et trop tard quand il ne l'est plus.
:::

## Ce que le VPW réussit, et ce qu'il assume de rater

**Les réussites.** Le VPW hérite de toutes les vertus du pourcentage ([[pourcentage-fixe]]) : ruine du capital impossible, contracyclicité, auto-correction face aux erreurs de rendement. Il y ajoute la conscience de l'horizon. Là où le pourcentage fixe thésaurise éternellement, le VPW ose consommer, et sa consommation totale moyenne sur la vie du plan compte parmi les plus élevées de toutes les règles. C'est la stratégie anti-« mourir riche » par excellence. Sa gouvernance, enfin, est remarquable : une table imprimée, un ratio par an, aucun paramètre à rediscuter. La règle survit ainsi à son auteur et aux années de panique ([[psychologie-du-retrait]]).

**Les ratages assumés.** Le premier est celui de toute la famille : le revenu suit le portefeuille. Le VPW ne lisse rien par construction, car la doctrine refuse le lissage comme une dette déguisée. Une baisse de 30 % du portefeuille entraîne donc une baisse de 30 % du revenu l'année suivante. D'où le garde-fou que la doctrine impose et que tout le monde saute, le **test de tolérance à la perte**. Avant d'adopter le VPW, calculez votre revenu dans l'hypothèse « actions −50 % », que le classeur affiche en permanence, et vérifiez qu'il couvre encore votre plancher ([[combien-il-vous-faut]]). Sinon, le VPW vous dit lui-même de réduire la part actions ou de couvrir le plancher autrement, par une pension ou une rente. C'est une stratégie qui exige un plancher externe ou une vraie élasticité, exactement comme son parent proportionnel.

::: figure vpw-test-de-perte
Le test que la doctrine impose, sur le cas du couple traité plus bas. À gauche, le pont de pension dort en obligations : le krach ne mord que sur la poche VPW et le ménage reste au-dessus de son confort. À droite, sans pont, la même règle sert davantage pendant vingt ans, puis le même krach passe sous le confort. C'est la démonstration en deux barres que le pont n'est pas un raffinement du classeur mais la condition d'admissibilité de la règle.
:::

**Le ratage discutable : les rendements supposés fixes.** Le g de la table, soit 5 % réel pour les actions, est une moyenne historique de très long terme. Il vaut la même chose en 2013, en 2021 (CAPE 38 !) et en 2026. C'est un choix philosophique cohérent, celui de ne pas prévoir, mais coûteux quand le marché est cher. Le VPW retire alors davantage que ce que les valorisations promettent ([[valorisations-et-cape]], [[rendements-attendus]]), et l'ajustement se fera ex post, par la baisse du revenu quand la déception arrivera. La famille ABW/TPAW fait le choix inverse, celui de brancher les rendements attendus courants, CAPE compris : plus juste en espérance, mais plus dépendant des modèles. C'est la ligne de partage entre les deux cousins ([[amortissement-abw]]).

::: science VPW et ABW : la même formule, deux épistémologies
Mathématiquement, le VPW est un ABW à rendements supposés constants, sans valeur actualisée fine des flux futurs : la même annuité inversée. La divergence est épistémologique. Le VPW parie que le retraité moyen se trompera moins avec une table gravée qu'avec des prévisions annuelles, c'est la robustesse comportementale. L'ABW parie que l'information des prix courants vaut mieux qu'une moyenne séculaire, c'est la justesse conditionnelle. La recherche penche pour l'ABW sur les chiffres, mais sur des simulations où la règle est appliquée sans faille. La sagesse des forums penche pour le VPW chez les humains réels. Le choix honnête dépend de qui exécutera la règle dans vingt ans ([[couple-et-famille]], [[choisir-sa-strategie]]).
:::

## Pour qui, et les réglages qui comptent

Le profil idéal du VPW cumule trois traits : un plancher couvert hors portefeuille, une vraie élasticité de train de vie au-dessus de ce plancher, et un goût pour la simplicité auditable. Le plancher peut venir d'une pension présente ou pontée, ou d'une rente ; le test de tolérance à la perte passe alors naturellement. La simplicité, elle, tient à la table et au classeur, rien d'autre. C'est très exactement le retraité Bogleheads type. C'est aussi la phase adossée d'un plan FIRE français ([[horizon-et-esperance-de-vie]]) : après 65-67 ans, pension au plancher, le VPW sur le portefeuille résiduel est difficile à battre.

En phase à découvert d'un FIRE précoce, le VPW demande deux aménagements. Le premier est le pont de pension du classeur, et il est obligatoire. Sans lui, le pourcentage s'applique à un capital qui doit aussi fabriquer quinze ans de pension, et le test de perte échoue presque toujours. Le second, en marché cher, est une décote manuelle de g : utiliser 4 % au lieu de 5 % pour les actions revient à intégrer grossièrement l'ancre CAPE, sans trahir l'esprit de la table.

::: astuce Éprouver le VPW dans un simulateur
La table officielle vit dans le classeur Bogleheads. Dans un simulateur généraliste, deux règles l'approchent, à essayer l'une après l'autre, puisqu'une seule politique de dépense pilote un plan à la fois.

- **Le retrait en pourcentage constant du portefeuille.** C'est le VPW d'un horizon encore long, la zone plate de la table. L'approximation est excellente avant 60-65 ans.
- **Le retrait amorti sur l'horizon restant** (ABW/TPAW). C'est l'annuité inversée intégrale : pourcentage croissant, horizon exact, flux futurs actualisés et rendements attendus courants, ancre de valorisation comprise si l'outil en propose une. C'est le VPW moderne, celui de TPAW Planner ([[amortissement-abw]]).

Dans les deux cas, jugez sur le niveau de vie servi année après année, jamais sur la probabilité de ruine, que toute règle proportionnelle annule par construction ([[flexibilite-realite]]).
:::

::: exemple Le VPW d'un couple FIRE, pont compris
Nora et Malik, 47 ans, 1,6 M€, 60/40, pensions estimées 21 600 €/an à 67 ans, plancher 38 000 €, confort 52 000 €. Le pont provisionne les vingt annuités de pension manquantes, actualisées à 1,9 % réel, soit 356 000 € placés en obligations. Le VPW s'applique au reste, 1 243 000 € à 47 ans, table 60/40 à 4,0 %, c'est-à-dire 50 000 €. S'y ajoute la tranche de pont, 21 600 € par an, pour un revenu initial de 71 600 €. Vient le test de tolérance, actions −50 %. La poche VPW tombe à 870 000 € et sa part du revenu à 35 000 €, mais le pont ne bouge pas, puisqu'il dort en obligations. Le ménage vit donc encore sur 56 600 €, au-dessus de son confort et loin au-dessus de son plancher. Sans le pont, la même règle aurait servi 64 300 € pendant vingt ans, puis 85 900 € d'un coup à 67 ans, et le krach l'aurait ramené à 45 000 €, sous le confort. Vingt ans plus tard, pensions liquidées, le VPW retire 5,0 % du résiduel pour le confort et les projets. Deux régimes, une seule table, zéro prévision. C'est le VPW bien construit.
:::

## L'essentiel à retenir

- Le VPW est l'annuité d'un crédit inversée. Chaque année, il applique le pourcentage qui épuiserait le portefeuille sur les années restantes, à rendements supposés fixes (5 % réel actions, 1,9 % obligations). Ce pourcentage croît avec l'âge : environ 3,9 % à 40 ans, 4,8 % à 65 ans, 100 % à 99 ans.
- Il hérite du pourcentage fixe (jamais de ruine du capital, contracyclique, auto-correcteur) et y ajoute la conscience de l'horizon : consommation délibérée, legs quasi nul, et la consommation moyenne la plus généreuse du panorama.
- Ses exigences : le test de tolérance à la perte (revenu sous « actions −50 % » ≥ plancher, à ne jamais sauter), le pont de pension en phase à découvert, et l'annuitisation vers 80 ans pour la longévité au-delà de la table.
- Sa ligne de partage avec l'ABW : table gravée (robustesse comportementale) contre rendements courants (justesse conditionnelle). Même formule, deux paris sur celui qui l'exécute.
- Deux règles l'approchent dans un simulateur, à tester séparément : le pourcentage constant (le VPW à horizon long) et le retrait amorti sur l'horizon (le VPW dynamique complet). Jugez-les sur le niveau de vie servi, et en marché cher, décotez g d'un point.

---

## Pour aller plus loin

- Bogleheads wiki, « Variable percentage withdrawal (VPW) » et le fil « VPW forward test » du forum : la doctrine, la table, le classeur, et dix ans d'exécution documentée en conditions réelles.
- Le VPW worksheet (classeur officiel, gratuit) : ponts de pension et test de perte intégrés.
- Early Retirement Now, volet 11 (le VPW noté contre les autres règles) ([[serie-ern]]).
- Dans ce livre : [[pourcentage-fixe]] (le parent), [[amortissement-abw]] (le cousin moderne), [[rentes-et-annuites]] (le complément de fin de vie), [[choisir-sa-strategie]] (l'arbitrage final).
