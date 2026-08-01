# PFU, barème, abattements : l'imposition des retraits

Les études de taux de retrait raisonnent hors impôt ([[etude-trinity]]). Votre banquier raisonne en rendement brut. Votre plan, lui, vit en **net**. Chaque euro de dépense doit être extrait du portefeuille à travers la fiscalité française. Cette extraction a un coût, typiquement 5 à 15 % du flux, et jusqu'à plus de 20 % si l'on s'y prend mal. C'est l'équivalent d'un demi-point de taux de retrait, autant que bien des débats de stratégie ([[choisir-sa-strategie]]).

Cet article est le manuel de calcul. Il couvre le PFU et son alternative, l'option pour le barème, un arbitrage annuel que la plupart des rentiers précoces gagnent à examiner, car leur TMI est souvent bien plus basse qu'ils ne croient. Il détaille la mécanique exacte d'une vente, c'est-à-dire la taxation de la seule part de gain au prix moyen pondéré, le mécanisme qui rend les premières années de retraite si peu taxées. Il présente les techniques de lissage, qui consistent à remplir chaque année les tranches basses, l'actif périssable du rentier. Il répond à la question la plus posée du chapitre, l'ordre dans lequel vider PEA, assurance-vie et CTO, où le tri par taux instantané se trompe de sens. Il aborde la purge des plus-values par donation et par décès, le « step-up » à la française qui change la fin de partie. Il passe en revue les couches annexes (CSG déductible, CEHR, IFI). Enfin, il montre comment calibrer le curseur « Tax on gains » sur **votre** situation plutôt que d'en rester au réglage par défaut.

Même avertissement que pour tout le chapitre. **Les chiffres sont à jour de 2026, à re-vérifier chaque année. Les structures de raisonnement, elles, survivent aux lois de finances.**

::: cle Les deux mécanismes qui dominent tout
**Un**. On n'est jamais taxé sur ce qu'on retire, mais sur la **part de gain** de ce qu'on vend, calculée au prix moyen pondéré d'acquisition. Un retrait de 50 000 € sur une ligne dont la part de gain atteint 30 % ne déclenche l'impôt que sur 15 000 €. La friction réelle d'un flux est donc (taux) × (fraction de gain). Elle démarre basse et monte avec les années. **Deux**. Le choix PFU/barème se refait **chaque** année, et la TMI d'un rentier précoce sans salaire est souvent 0 ou 11 %. Les années de pont sont des années d'or fiscal. Les gaspiller à ne rien réaliser est l'erreur silencieuse la plus chère de la décumulation française.
:::

## Le PFU, et l'option barème : l'arbitrage annuel

**Le régime par défaut est le prélèvement forfaitaire unique** (PFU), 31,4 % tout compris (12,8 % d'IR + 18,6 % de prélèvements sociaux), sur les dividendes, intérêts et plus-values mobilières du CTO ([[enveloppes-francaises]] ; le PEA et l'AV ont leurs régimes propres). Il est simple, prévisible et indépendant de vos autres revenus. C'est le bon choix par défaut des années à revenus élevés.

**L'option pour le barème** est globale et annuelle. Elle s'applique à **tous** les revenus mobiliers de l'année. Ceux-ci rejoignent le barème progressif (0/11/30/41/45 %) auquel s'ajoutent 18,6 % de PS. Trois compensations viennent en contrepartie. L'abattement de 40 % sur les dividendes. La CSG déductible (6,8 points déduits du revenu imposable de l'année suivante). Et, pour les titres acquis avant 2018, les abattements de durée de détention sur les plus-values (50 % au-delà de 2 ans, 65 % au-delà de 8, sur l'IR seulement, jamais sur les prélèvements sociaux).

L'arbitrage se lit par tranche. À **TMI 0 ou 11 %, le barème gagne presque toujours**, car il coûte au pire 11 % d'IR plus 18,6 % de PS, soit 29,6 % contre 31,4 %, et bien moins quand la base est abattue. À TMI 30 %, le PFU reprend l'avantage sur les dividendes comme sur les plus-values récentes ; seuls les abattements de durée des titres acquis avant 2018 peuvent encore faire basculer le calcul, et comme l'option reste globale, elle se simule en bloc. À TMI 41 % et plus, c'est le PFU sans discussion.

Le point clé pour un FIRE tient à la phase de pont. **Sans** salaire ni pension, votre revenu **imposable** peut être minuscule, car les rachats AV sous abattement et le PEA n'en créent presque pas ([[enveloppes-francaises]]). La TMI 0-11 % devient alors votre régime de croisière, et l'option barème s'impose presque toujours.

## La mécanique d'une vente, et le lissage : les deux gestes du rentier

**Le PMP et la fraction de gain.** Chaque ligne de CTO porte un prix moyen pondéré d'acquisition (PMP). Chaque vente réalise une plus-value égale à (cours − PMP) × quantité. La friction d'un flux de retrait vaut donc taux × (1 − PMP/cours). Les conséquences sont concrètes. Les premières années d'un plan, avec un PMP proche du cours, extraient le cash presque gratuitement. La friction **dérive** ensuite vers le haut à mesure que la plus-value latente s'accumule. C'est exactement ce que reproduit le modèle fiscal, en ne taxant chaque vente que sur sa part de gain, croissante au fil du plan ([[la-machine-pofo]]). Enfin, à flux égal, mieux vaut vendre les lignes au PMP le plus haut, celles dont la friction est minimale, sauf objectif inverse de purge (voir ci-dessous).

::: figure friction-derive-pmp
Impôt payé pour 100 € de flux extrait, selon l'âge de la ligne vendue, aux taux du 1er janvier 2026 et pour un cours qui monte de 5 % par an en réel. La courbe en pointillés suppose un versement constant chaque année : le PMP se recharge et la dérive ralentit.
:::

**Le lissage de taux, ou remplir les tranches basses chaque année.** Les tranches à 0 et 11 % et l'abattement AV sont des capacités **annuelles**. Non utilisées, elles sont perdues. Le geste d'optimisation central du rentier consiste donc à **réaliser** chaque année, même sans besoin de cash, assez de gains pour remplir ces capacités basses. Trois outils pour cela. Vendre puis racheter une ligne de CTO relève le PMP et pré-purge les gains futurs au taux d'aujourd'hui ; il n'y a pas de délai de carence sur ce « rafraîchissement » en droit français, contrairement au wash sale américain, mais vérifiez l'état du droit. Racheter de l'AV à hauteur de l'abattement. Opter pour le barème les années creuses. Sur les 10-15 ans d'un pont à TMI basse, le lissage systématique purge une fraction majeure des plus-values latentes à 19-29 % au lieu de 31,4 %. Cela représente des dizaines de milliers d'euros pour trois ordres par an ([[revue-annuelle]], où le lissage est un point fixe).

::: science La purge des plus-values : la fin de partie française
La France n'a pas de « step-up » du vivant, mais **deux** mécanismes de purge structurent la fin de partie patrimoniale. Le premier est **le décès**. Les plus-values latentes du CTO sont alors purgées ; les héritiers reçoivent les titres au cours du jour, l'impôt de PV disparaît, et les droits de succession s'appliquent sur la valeur ([[succession-et-transmission]]). La conséquence est contre-intuitive. Au grand âge, il devient rationnel de **conserver** les lignes aux plus grosses PV latentes et de consommer le reste, soit l'inverse exact de la logique de friction des jeunes années.

Le second est **la donation-cession**. On donne des titres, ce qui purge la PV latente ; les abattements de 100 k€ par parent et par enfant se rechargent tous les 15 ans. Le donataire vend ensuite au cours du jour, pour zéro impôt de PV. C'est l'outil canonique pour financer les enfants ou anticiper la transmission, à exécuter dans les règles (donation **avant** la cession, réelle et non fictive, car le formalisme compte).

Ces deux mécanismes font de la détention CTO longue une stratégie de transmission à part entière. Ils renversent l'ordre de consommation des enveloppes quand le legs entre dans les objectifs ([[enveloppes-francaises]]).
:::

## L'ordre des enveloppes : le taux de demain, pas celui d'aujourd'hui

PEA, assurance-vie et CTO garnis, lequel entamer en premier ? Le réflexe courant trie par taux et conclut « le PEA d'abord, puisqu'il est le moins taxé, le CTO en dernier ». C'est l'inverse de la bonne réponse, pour une raison qui tient en deux phrases. Le taux d'aujourd'hui frappe le gain déjà acquis, qui sera taxé peu ou prou pareil, qu'on vende maintenant ou dans dix ans. Ce que l'ordre de consommation décide vraiment, c'est le taux qui frappera la croissance **future** de l'argent laissé en place.

Un euro conservé sur le CTO fabrique de la plus-value future taxée à 31,4 %. Le même euro conservé sur le PEA fabrique la même plus-value, taxée à 18,6 % pour toujours. On vide donc d'abord l'enveloppe dont la croissance à venir est la plus taxée, et on garde la moins taxée en serre. À parts de gain comparables, le CTO se consomme avant le PEA. Le réflexe inverse n'est pas absurde, il est myope, car il optimise l'impôt de l'année au prix de toutes les années suivantes.

L'assurance-vie s'intercale par ses frais. Au-delà de l'abattement, ses rachats coûtent 24,7 %, entre le PEA et le CTO. Mais elle est la seule enveloppe à prélever des frais de gestion sur tout l'encours, chaque année, rachats ou pas, typiquement 0,5 à 1 % sur les unités de compte. Ce péage est un impôt privé sur la croissance future, et il s'ajoute au taux fiscal dans le tri. Sur quinze ou vingt ans, 0,5 % par an pèse plus lourd que les 6,7 points d'écart de taux avec le CTO : un contrat ordinaire se vide donc en premier, et seul un contrat exceptionnellement peu chargé mérite d'attendre son tour derrière le CTO. D'où la règle complète, en une ligne. **Chaque année, servir d'abord les capacités périssables (abattement, tranches basses), puis puiser dans les enveloppes par taux décroissant sur leur croissance future, frais compris, soit le contrat chargé, puis le CTO, et le PEA en dernier.**

::: exemple Trois enveloppes jumelles, six ordres possibles
100 000 € du même ETF Monde, placés il y a dix ans sur un CTO, un PEA et une assurance-vie à 0,5 % de frais annuels, et multipliés par trois depuis. Le CTO et le PEA valent 300 000 €, l'assurance-vie 285 000 €, les frais ayant rogné le reste. La part de gain atteint 67 % (65 % côté assurance-vie), et le tri par taux instantané se lit dans le coût de 100 € vendus.

| Robinet | Impôt pour 100 € vendus |
|---|---|
| Assurance-vie, sous l'abattement annuel | 11,2 € |
| PEA | 12,4 € |
| Assurance-vie, au-delà de l'abattement | 16,0 € |
| CTO | 20,9 € |

Faisons maintenant vivre les trois poches. Des retraits de 25 000 € nets par an, un rendement réel de 4 %, quinze ans, puis tout est liquidé et l'on compare le patrimoine net final des six ordres possibles. Le meilleur vide l'assurance-vie, puis le CTO, et garde le PEA pour la fin : environ 756 000 € nets. L'ordre naïf du tri par taux, le PEA d'abord et le CTO en dernier, laisse 733 000 €. Mêmes marchés, mêmes retraits, 23 000 € d'écart, et il triple sur vingt-cinq ans (plus de 60 000 €). Les deux gestes qui payent, dans l'ordre de leur poids, sont de vider tôt le contrat qui prélève des frais, et de garder la serre PEA pour la fin.

L'abattement, lui, reste servi en premier chaque année, mais il est petit. À 65 % de part de gain, les 4 600 € d'un célibataire couvrent environ 7 100 € de rachat, les 9 200 € d'un couple 14 200 €. Et les années creuses gardent leur privilège : à TMI 0, le CTO au barème tombe à 18,6 %, le prix du PEA, et chaque euro purgé ces années-là est gagné sur le tarif plein (le lissage, ci-dessus).
:::

Trois réserves ferment le sujet. L'ordre s'entend à exposition égale : quand une brique ne vit que dans une enveloppe, c'est l'allocation qui décide quoi vendre, la fiscalité choisit ensuite où le vendre. L'objectif de legs le renverse, comme vu à l'encart précédent, le CTO et l'assurance-vie d'avant 70 ans passant en queue ([[succession-et-transmission]]). Et trier suppose des taux stables, or ils viennent de bouger : différer la consommation du PEA parie que ses 18,6 % ne flamberont pas, un pari raisonnable mais à revoir à chaque loi de finances ([[revue-annuelle]]).

## Les couches annexes, en bref

**La CEHR** (contribution exceptionnelle sur les hauts revenus) vaut 3 % au-delà de 250 k€ de revenu fiscal de référence (célibataire ; 500 k€ pour un couple), puis 4 % au-delà du double. Elle frappe les **années** à gros revenus réalisés, une raison de plus de lisser, car une grosse vente unique peut la déclencher là où trois ventes étalées y échappent. Sa cousine, **la CDHR** (contribution différentielle sur les hauts revenus, art. 224 CGI, pérennisée par la LFI 2026), impose aux mêmes seuils de RFR un plancher d'imposition de 20 % ; elle mord précisément quand une grosse plus-value au PFU ferait descendre le taux moyen sous ce plancher, et l'étalement est là encore la parade.

**L'IFI** est l'impôt sur la fortune **immobilière** seulement, avec un seuil de 1,3 M€ de patrimoine immobilier net. Le portefeuille financier n'y est pas soumis. C'est un argument de structure pour le rentier financier, à connaître sans en faire un dogme ([[immobilier-en-retrait]]).

**Les prélèvements sociaux** sont la couche incompressible de presque tout. La loi de financement de la Sécurité sociale pour 2026 a porté la CSG du capital de 9,2 à 10,6 points, donc les prélèvements de 17,2 à **18,6 %**. Mais elle a épargné toute une famille de supports, et elle ne mord pas partout à la même date. Les revenus du patrimoine de l'article L. 136-6 du code de la Sécurité sociale sont touchés dès les revenus 2025. Les produits de placement de l'article L. 136-7 ne le sont qu'à partir des encaissements du 1er janvier 2026. Voici la carte, aux taux de la LFSS 2026 (loi du 30 décembre 2025, art. 12).

| Support | PS | Application |
|---|---|---|
| Plus-values mobilières (CTO) | 18,6 % | revenus 2025 |
| Crypto | 18,6 % | revenus 2025 |
| Location meublée | 18,6 % | revenus 2025 |
| Dividendes et intérêts | 18,6 % | encaissements 2026 |
| Gains de PEA à la sortie | 18,6 % | encaissements 2026 |
| Assurance-vie et capitalisation | 17,2 % | hausse écartée |
| Foncier nu et SCPI | 17,2 % | hausse écartée |
| Plus-values immobilières | 17,2 % | hausse écartée |
| PEL, CEL et PEP anciens | 17,2 % | hausse écartée |

Les livrets défiscalisés (A, LDDS, LEP) restent hors de tout cela. Un plan de retrait français croise donc désormais deux taux, et l'assurance-vie y gagne un avantage relatif. La CSG n'est déductible qu'en cas d'option barème, un des termes de l'arbitrage.

**La PUMa**, enfin, est la couche spécifique du rentier sans activité, assez importante pour mériter son chapitre entier ([[taxe-puma]]).

## Calibrer le simulateur : votre taux mixte, en trois lignes

Le curseur « Tax on gains » applique un taux unique à la part de gain de chaque vente ([[utiliser-la-page-fire]] ; le défaut, à 32,8 %, approxime PFU + PUMa). Pour le calibrer sur **votre** plan, procédez en trois temps. Estimez d'abord la répartition de vos flux de retrait entre robinets ([[enveloppes-francaises]]) et le taux effectif de chacun (PEA 18,6 %, AV 17,2 % sous abattement puis 24,7 %, CTO 31,4 %, ou votre barème + PS des années creuses, souvent 19-29 %). Pondérez ensuite par les parts. Le taux mixte d'un plan bien organisé ressort typiquement à **15-25 %** ; ajoutez la PUMa si elle vous concerne ([[taxe-puma]], selon la structure des revenus). Re-testez enfin la ruine. L'écart entre le défaut prudent et votre taux calibré vaut souvent 0,5-1 point de ruine, soit un vrai paramètre du plan et non un détail d'affichage.

::: attention Les erreurs qui coûtent
**Cinq** classiques. La première, laisser dormir les tranches basses du pont (l'actif fiscal périssable, des années à TMI 0-11 % sans aucune réalisation, le gaspillage silencieux n° 1). La deuxième, l'option barème oubliée ou cochée à tort ; elle est **globale**, alors simulez les deux chaque année, comme le font les simulateurs de la déclaration. La troisième, la grosse vente unique (une année à 300 k€ de gains réalisés, PFU + CEHR + PUMa maximale, là où l'étalement sur trois ans passait sous tous les seuils). La quatrième, vider le PEA d'abord « parce que c'est le moins taxé » ; le taux d'aujourd'hui est acquis, c'est celui de la croissance future que l'ordre pilote (section ci-dessus). La cinquième, l'optimisation fiscale qui pilote le portefeuille (garder une ligne pourrie « pour ne pas payer la PV », vendre une bonne ligne « pour l'abattement »). La fiscalité module l'**exécution**, c'est-à-dire quelle ligne et quelle année, jamais l'**allocation** ([[allocation-actions-obligations]] ; le portefeuille d'abord, la fiscalité ensuite).
:::

::: exemple L'année fiscale type d'un couple en phase de pont
Un couple, 55 et 57 ans, a besoin de 52 500 € pour vivre, sans aucun salaire, les pensions arrivant dans 10 ans. Voici leurs flux de l'année. Les rachats AV apportent 24 000 € (part de gains 9 000 €, sous l'abattement de 9 200 €, donc impôt ~0 et PS sur gains ~1 550 €). Les retraits PEA apportent 20 000 € (part de gains 9 500 €, PS 1 770 €). Les ventes CTO apportent 12 000 € (gains 4 200 €), auxquelles s'ajoute un **lissage** : une vente-rachat complémentaire porte les gains CTO réalisés à 22 000 €, juste sous les 23 200 € que vaut la tranche à 0 % d'un couple. Le tout passe à l'option barème, et l'IR ressort donc à zéro. Restent 4 090 € de PS sur les gains du CTO, et une CSG déductible l'an prochain. L'option globale ne coûte rien ici, l'AV restant sous son abattement et le PEA hors IR, alors qu'au PFU les mêmes 22 000 € auraient porté 2 820 € d'IR en plus.

La friction totale sur 56 000 € extraits atteint ~7 410 €, dont ~3 310 € de purge **volontaire** de PV futures à 18,6 % au lieu de 31,4 % et plus. La friction du flux réellement vécu tombe donc à ~7 %, et s'y ajoutent 18 000 € de PMP rehaussé pour les années suivantes. Côté trésorerie, les PS de l'AV et du PEA sont retenus à la source quand l'impôt du CTO n'arrive qu'au printemps suivant. Trois robinets, un arbitrage barème et un lissage, cela fait quarante minutes par an, à la revue ([[revue-annuelle]]).
:::

## L'essentiel à retenir

- On est taxé sur la part de **gain** des ventes (au PMP), pas sur les retraits. La friction démarre basse et dérive vers le haut. Un simulateur reproduit exactement cette dérive ; calibrez son curseur sur votre taux mixte réel (15-25 % pour un plan organisé) plutôt que sur le réglage par défaut.
- L'arbitrage PFU/barème se refait chaque année et il est **global**. Les années de pont à TMI 0-11 % sont des années d'or, où le barème gagne presque toujours et où les tranches basses non remplies sont perdues. Le **lissage** (réaliser des gains chaque année, vente-rachat, abattement AV) est le geste central.
- L'ordre des enveloppes se décide sur le taux qui frappera la croissance **future**, frais de contrat compris, jamais sur le taux instantané. Capacités périssables d'abord, puis contrat chargé, puis CTO, et le PEA en dernier ; l'objectif de legs renverse cet ordre.
- La fin de partie a ses purges. Le décès efface les PV latentes du CTO, d'où l'intérêt de conserver les grosses PV au grand âge. La donation-cession les efface du vivant (100 k€ par parent et par enfant tous les 15 ans). Le legs renverse l'ordre de consommation.
- Les couches annexes se gèrent par l'étalement (CEHR), par la structure (IFI, immobilier seulement) et par le chapitre dédié (PUMa). La fiscalité module l'exécution, **jamais** l'allocation.
- Tout est daté de 2026. L'arbitrage annuel et la veille font partie du plan. Les taux changeront ; les mécanismes (part de gain, tranches, purges) sont les invariants à comprendre.

---

## Pour aller plus loin

- [impots.gouv.fr](https://www.impots.gouv.fr) : le simulateur officiel (l'arbitrage PFU/barème s'y teste en dix minutes) et le BOFiP pour les régimes fins (donation-cession, abattements de durée).
- Les notices 2074/2042 : la mécanique déclarative des PV, aride et instructive.
- Dans ce livre : [[enveloppes-francaises]] (les robinets et leur ordre), [[taxe-puma]] (la couche du rentier), [[succession-et-transmission]] (les purges en stratégie), [[utiliser-la-page-fire]] (le curseur fiscal).
