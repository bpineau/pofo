# PFU, barème, abattements : l'imposition des retraits

Les études de taux de retrait raisonnent hors impôt ([[etude-trinity]]). Votre banquier raisonne en rendement brut. Votre plan, lui, vit en **net**. Chaque euro de dépense doit être extrait du portefeuille à travers la fiscalité française. Cette extraction a un coût, typiquement 5 à 15 % du flux, et jusqu'à plus de 20 % si l'on s'y prend mal. C'est l'équivalent d'un demi-point de taux de retrait, autant que bien des débats de stratégie ([[choisir-sa-strategie]]).

Cet article est le manuel de calcul. Il couvre le PFU et son alternative au barème, un arbitrage annuel que la plupart des rentiers précoces gagnent à examiner, car leur TMI est souvent bien plus basse qu'ils ne croient. Il détaille la mécanique exacte d'une vente, c'est-à-dire la taxation de la seule part de gain au prix moyen pondéré, le mécanisme qui rend les premières années de retraite si peu taxées. Il présente les techniques de lissage, qui consistent à remplir chaque année les tranches basses, un actif périssable du rentier. Il aborde la purge des plus-values par donation et par décès, le « step-up » à la française qui change la fin de partie. Il passe en revue les couches annexes (CSG déductible, CEHR, IFI). Enfin, il montre comment calibrer le curseur « Tax on gains » sur **votre** situation plutôt que de laisser le défaut.

Même avertissement que pour tout le chapitre. **Les chiffres sont à jour de 2026, à re-vérifier chaque année. Les structures de raisonnement, elles, survivent aux lois de finances.**

::: cle Les deux mécanismes qui dominent tout
**Un**. On n'est jamais taxé sur ce qu'on retire, mais sur la **part de gain** de ce qu'on vend, calculée au prix moyen pondéré d'acquisition. Un retrait de 50 000 € sur une ligne dont la part de gain atteint 30 % ne déclenche l'impôt que sur 15 000 €. La friction réelle d'un flux est donc (taux) × (fraction de gain). Elle démarre basse et monte avec les années. **Deux**. Le choix PFU/barème se refait **chaque** année, et la TMI d'un rentier précoce sans salaire est souvent 0 ou 11 %. Les années de pont sont des années d'or fiscal. Les gaspiller à ne rien réaliser est l'erreur silencieuse la plus chère de la décumulation française.
:::

## Le PFU, et l'option barème : l'arbitrage annuel

**Le régime par défaut est le prélèvement forfaitaire unique** (PFU), 30 % tout compris (12,8 % d'IR + 17,2 % de prélèvements sociaux), sur les dividendes, intérêts et plus-values mobilières du CTO ([[enveloppes-francaises]] ; le PEA et l'AV ont leurs régimes propres). Il est simple, prévisible et indépendant de vos autres revenus. C'est le bon défaut des années à revenus élevés.

**L'option pour le barème** est globale et annuelle. Elle s'applique alors à **tous** les revenus mobiliers de l'année. Ceux-ci rejoignent le barème progressif (0/11/30/41/45 %) auquel s'ajoutent 17,2 % de PS. Trois compensations viennent en face. L'abattement de 40 % sur les dividendes. La CSG déductible (6,8 points déduits du revenu imposable de l'année suivante). Et, pour les titres acquis avant 2018, les abattements de durée de détention sur les plus-values (50 % au-delà de 2 ans, 65 % au-delà de 8).

L'arbitrage se lit par tranche. À **TMI 0 ou 11 %, le barème gagne presque toujours** (11 % + 17,2 % de PS sur une fraction abattue inférieure à 30 %). À TMI 30 %, le match est serré et se calcule cas par cas, car les dividendes sont abattus au barème quand les PV se taxent mieux au PFU, mais l'option reste globale. À TMI 41 % et plus, c'est le PFU sans discussion.

Le point clé pour un FIRE tient à la phase de pont. **Sans** salaire ni pension, votre revenu **imposable** peut être minuscule, car les rachats AV sous abattement et le PEA n'en créent presque pas ([[enveloppes-francaises]]). La TMI 0-11 % devient alors votre régime de croisière, et l'option barème s'impose presque toujours.

## La mécanique d'une vente, et le lissage : les deux gestes du rentier

**Le PMP et la fraction de gain.** Chaque ligne de CTO porte un prix moyen pondéré d'acquisition (PMP). Chaque vente réalise une plus-value égale à (cours − PMP) × quantité. La friction d'un flux de retrait vaut donc taux × (1 − PMP/cours). Les conséquences sont concrètes. Les premières années d'un plan, avec un PMP proche du cours, extraient le cash presque gratuitement. La friction **dérive** ensuite vers le haut à mesure que la plus-value latente s'accumule. C'est exactement ce que reproduit le modèle fiscal, en majorant chaque vente sur la seule part de gain, croissante au fil du plan ([[la-machine-pofo]]). Enfin, à flux égal, mieux vaut vendre les lignes au PMP le plus haut, celles dont la friction est minimale, sauf objectif inverse de purge (voir ci-dessous).

**Le lissage de taux, ou remplir les tranches basses chaque année.** Les tranches à 0 et 11 % et l'abattement AV sont des capacités **annuelles**. Non utilisées, elles sont perdues. Le geste d'optimisation central du rentier consiste donc à **réaliser** chaque année, même sans besoin de cash, assez de gains pour remplir ces capacités basses. Trois outils pour cela. Vendre puis racheter une ligne de CTO relève le PMP et pré-purge les gains futurs au taux d'aujourd'hui ; il n'y a pas de délai de carence sur ce « rafraîchissement » en droit français, contrairement au wash sale américain, mais vérifiez l'état du droit. Racheter de l'AV à hauteur de l'abattement. Arbitrer au barème les années creuses. Sur les 10-15 ans d'un pont à TMI basse, le lissage systématique purge une fraction majeure des plus-values latentes à 17-28 % au lieu de 30. Cela représente des dizaines de milliers d'euros pour trois ordres par an ([[revue-annuelle]], où le lissage est un point fixe).

::: science La purge des plus-values : la fin de partie française
La France n'a pas de « step-up » du vivant, mais **deux** mécanismes de purge structurent la fin de partie patrimoniale. Le premier est **le décès**. Les plus-values latentes du CTO sont alors purgées ; les héritiers reçoivent les titres au cours du jour, l'impôt de PV disparaît, et les droits de succession s'appliquent sur la valeur ([[succession-et-transmission]]). La conséquence est contre-intuitive. Au grand âge, il devient rationnel de **conserver** les lignes aux plus grosses PV latentes et de consommer le reste, soit l'inverse exact de la logique de friction des jeunes années.

Le second est **la donation-cession**. On donne des titres, ce qui purge la PV latente ; les abattements de 100 k€ par parent et par enfant se rechargent tous les 15 ans. Le donataire vend ensuite au cours du jour, pour zéro impôt de PV. C'est l'outil canonique pour financer les enfants ou anticiper la transmission, à exécuter dans les règles (donation **avant** la cession, réelle et non fictive, car le formalisme compte).

Ces deux mécanismes font de la détention CTO longue une stratégie de transmission à part entière. Ils renversent l'ordre de consommation des enveloppes quand le legs entre dans les objectifs ([[enveloppes-francaises]]).
:::

## Les couches annexes, en bref

**La CEHR** (contribution exceptionnelle sur les hauts revenus) vaut 3 % au-delà de 250 k€ de revenu fiscal de référence (célibataire ; 500 k€ pour un couple), puis 4 % au-delà du double. Elle frappe les **années** à gros revenus réalisés, une raison de plus de lisser, car une grosse vente unique peut la déclencher là où trois ventes étalées y échappent.

**L'IFI** est l'impôt sur la fortune **immobilière** seulement, avec un seuil de 1,3 M€ de patrimoine immobilier net. Le portefeuille financier n'y est pas soumis. C'est un argument de structure pour le rentier financier, à connaître sans en faire un dogme ([[immobilier-en-retrait]]).

**Les prélèvements sociaux (17,2 %)** sont la couche incompressible de presque tout, sauf les livrets réglementés. La CSG n'est déductible qu'en cas d'option barème, un des termes de l'arbitrage.

**La PUMa**, enfin, est la couche spécifique du rentier sans activité, assez importante pour mériter son chapitre entier ([[taxe-puma]]).

## Calibrer le simulateur : votre taux mixte, en trois lignes

Le curseur « Tax on gains » applique un taux unique à la part de gain de chaque vente ([[utiliser-la-page-fire]] ; le défaut, à 31,4 %, approxime PFU + PUMa). Pour le calibrer sur **votre** plan, procédez en trois temps. Estimez d'abord la répartition de vos flux de retrait entre robinets ([[enveloppes-francaises]]) et le taux effectif de chacun (PEA 17,2 %, AV sous abattement ~0 % puis 24,7 %, CTO 30 %, ou votre barème + PS des années creuses, souvent 17-28 %). Pondérez ensuite par les parts. Le taux mixte d'un plan bien organisé ressort typiquement à **15-25 %** ; ajoutez la PUMa si elle vous concerne ([[taxe-puma]], selon la structure des revenus). Re-testez enfin la ruine. L'écart entre le défaut prudent et votre taux calibré vaut souvent 0,5-1 point de ruine, soit un vrai paramètre du plan et non un détail d'affichage.

::: attention Les erreurs qui coûtent
**Quatre** classiques. La première, laisser dormir les tranches basses du pont (l'actif fiscal périssable, des années à TMI 0-11 % sans aucune réalisation, le gaspillage silencieux n° 1). La deuxième, l'option barème oubliée ou cochée à tort ; elle est **globale**, alors simulez les deux chaque année, comme le font les simulateurs de la déclaration. La troisième, la grosse vente unique (une année à 200 k€ de gains réalisés, PFU + CEHR + PUMa maximale, là où l'étalement sur trois ans passait sous tous les seuils). La quatrième, l'optimisation fiscale qui pilote le portefeuille (garder une ligne pourrie « pour ne pas payer la PV », vendre une bonne ligne « pour l'abattement »). La fiscalité module l'**exécution**, c'est-à-dire quelle ligne et quelle année, jamais l'**allocation** ([[allocation-actions-obligations]] ; le portefeuille d'abord, la fiscalité ensuite).
:::

::: exemple L'année fiscale type d'un couple en phase de pont
Un couple, 55 et 57 ans, a besoin de 54 000 €, sans aucun salaire, les pensions arrivant dans 10 ans. Voici leurs flux de l'année. Les rachats AV apportent 24 000 € (part de gains 9 000 €, sous l'abattement de 9 200 €, donc impôt ~0 et PS sur gains ~1 550 €). Les retraits PEA apportent 20 000 € (part de gains 9 500 €, PS 1 630 €). Les ventes CTO apportent 12 000 € (gains 4 200 €), auxquelles s'ajoute un **lissage** : une vente-rachat complémentaire porte les gains CTO réalisés à 22 000 €. Le tout passe à l'option barème (TMI 11 % après abattements), soit IR ~1 900 €, PS ~3 780 €, et une CSG déductible l'an prochain.

La friction totale sur 56 000 € extraits atteint ~8 850 €, dont ~2 400 € de purge **volontaire** de PV futures à 28 % au lieu de 30 et plus. La friction nette du flux vécu tombe ainsi à ~11 %. S'y ajoutent 18 000 € de PMP rehaussé pour les années suivantes. Trois robinets, un arbitrage barème et un lissage, cela fait quarante minutes par an, à la revue ([[revue-annuelle]]).
:::

## L'essentiel à retenir

- On est taxé sur la part de **gain** des ventes (au PMP), pas sur les retraits. La friction démarre basse et dérive vers le haut. Un simulateur reproduit exactement cette dérive ; calibrez son curseur sur votre taux mixte réel (15-25 % pour un plan organisé) plutôt que sur le défaut.
- L'arbitrage PFU/barème se refait chaque année et il est **global**. Les années de pont à TMI 0-11 % sont des années d'or, où le barème gagne presque toujours et où les tranches basses non remplies sont perdues. Le **lissage** (réaliser des gains chaque année, vente-rachat, abattement AV) est le geste central.
- La fin de partie a ses purges. Le décès efface les PV latentes du CTO, d'où l'intérêt de conserver les grosses PV au grand âge. La donation-cession les efface du vivant (100 k€ par parent et par enfant tous les 15 ans). Le legs renverse l'ordre de consommation.
- Les couches annexes se gèrent par l'étalement (CEHR), par la structure (IFI, immobilier seulement) et par le chapitre dédié (PUMa). La fiscalité module l'exécution, **jamais** l'allocation.
- Tout est daté de 2026. L'arbitrage annuel et la veille font partie du plan. Les taux changeront ; les mécanismes (part de gain, tranches, purges) sont les invariants à comprendre.

---

## Pour aller plus loin

- [impots.gouv.fr](https://www.impots.gouv.fr) : le simulateur officiel (l'arbitrage PFU/barème s'y teste en dix minutes) et le BOFiP pour les régimes fins (donation-cession, abattements de durée).
- Les notices 2074/2042 : la mécanique déclarative des PV, aride et instructive.
- Dans ce livre : [[enveloppes-francaises]] (les robinets et leur ordre), [[taxe-puma]] (la couche du rentier), [[succession-et-transmission]] (les purges en stratégie), [[utiliser-la-page-fire]] (le curseur fiscal).
