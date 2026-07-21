# Suivre l'inflation : les indices, et la vôtre

Tout le plan vit en euros constants. Les retraits s'indexent, les simulations tournent en réel, le plancher est un pouvoir d'achat ([[la-machine-pofo]]). Mais constants par rapport à **quoi** ?

« L'inflation » n'est pas un fait brut. C'est une mesure construite, avec un panier qui n'est pas le vôtre. L'écart entre l'indice officiel et **votre** inflation personnelle paraît petit chaque année. Composé sur trente ans, il peut valoir autant qu'un point de taux de retrait. Cet article est le manuel de mesure. Il déroule d'abord les indices qui comptent pour un Français, l'IPC et l'IPCH, celui qui sert de déflateur, leur construction et leurs angles morts, dont le plus important, le logement du propriétaire. Il pèse ensuite le débat honnête sur leur fiabilité. Il aborde la question centrale de l'inflation **personnelle**, pourquoi celle d'un retraité dérive au-dessus de l'indice et comment estimer la vôtre en une soirée. Il lit les anticipations de marché, les points morts, la seule « prévision » d'inflation qui engage de l'argent. Il finit par la traduction opérationnelle, sur quoi indexer ses retraits et comment régler la dérive de dépenses dans un simulateur.

::: cle Les trois inflations à distinguer
Il y a trois inflations. L'inflation **officielle** est celle de l'IPC et de l'IPCH, celle des statistiques, des pensions et des linkers. L'inflation de **votre panier** pondère à votre façon la santé, les services, l'énergie et les loisirs, typiquement +0,2 à +0,5 point au-dessus de l'officielle pour un retraité. L'inflation de votre **plan** est celle que vous choisissez d'appliquer à vos retraits, soit l'officielle plus une dérive santé explicite, le réglage honnête. Confondre les trois fait des plans qui semblent tenir en euros constants... mais des mauvais euros.
:::

## Les indices : ce qu'ils sont, ce qu'ils contiennent

**L'IPC de l'INSEE.** C'est l'indice des prix à la consommation français. Il couvre environ 1 100 familles de produits, environ 200 000 relevés mensuels, pondérés par la consommation moyenne des ménages. Sa variante « hors tabac » sert de référence légale (c'est elle qui indexe le SMIC, les pensions... et les OATi, [[obligations-indexees]]).

**L'IPCH (HICP).** C'est l'indice **harmonisé** européen, la version comparable entre pays de la zone. Il sert de cible à la BCE (2 %), il indexe les OAT€i, et c'est **la série qui sert à déflater** (le symbole `^HICP-FR`, par lequel passent toutes les conversions en réel du simulateur, [[la-machine-pofo]]). L'IPC et l'IPCH français divergent peu, car leurs méthodes sont proches. L'IPCH pondère un peu différemment, notamment la santé nette des remboursements, d'où quelques dixièmes d'écart certaines années.

**Ce que le panier contient mal : le logement du propriétaire.** C'est le point technique le plus important. L'IPCH ne compte le logement que par les **loyers** effectifs, soit environ 7 % du panier français, le poids des locataires. Le coût du logement du **propriétaire**, le prix des logements et l'équivalent-loyer, n'y figure pratiquement pas. Quand l'immobilier double, l'IPCH ne bouge pas. Les conséquences pour vous sont doubles. À la baisse, un rentier propriétaire toit payé ([[immobilier-en-retrait]]) a une inflation personnelle structurellement plus douce que l'indice, car son plus gros poste est éteint. À la hausse, l'épargnant en accumulation qui vise un achat voit sa vraie inflation-cible galoper sans que l'indice le dise. Même logique pour tout ce que l'indice moyenne et que vous ne consommez pas.

**Les corrections de qualité (effets hédoniques).** Quand le smartphone de 2026 fait plus que celui de 2020 au même prix, l'indice compte une **baisse** de prix. C'est défendable en méthode, car on mesure le prix du service rendu. C'est débattu en pratique, car vous ne pouvez pas acheter « moins de qualité » et votre dépense, elle, ne baisse pas. Le débat honnête sur la fiabilité tient en deux phrases. Les indices officiels européens sont sérieux, audités, sans manipulation démontrée. Et ils mesurent un panier moyen à qualité corrigée qui peut s'écarter durablement de l'expérience de dépense d'un ménage donné. Les deux propositions sont vraies, d'où la section suivante.

## Votre inflation : pourquoi celle d'un retraité dérive au-dessus

La dispersion des inflations personnelles autour de l'indice est documentée. L'INSEE publie des indices par catégorie de ménage et un simulateur d'inflation personnalisée, qui chiffrent quelques dixièmes de point par an entre ménages types. Et le profil **retraité** cumule les pondérations défavorables :

- **La santé et la dépendance.** Ce poste croît en part du budget avec l'âge. Ses prix courent structurellement au-dessus de l'IPC, à commencer par la mutuelle (+4-6 %/an tendanciels, portés par l'âge et la dérive des coûts médicaux), puis les établissements et l'aide à domicile ([[sante-et-protection-sociale]], [[depenses-en-retraite]]).
- **Les services en général** (aide, entretien, restauration, assurances). Intensifs en main-d'œuvre locale, leurs prix suivent les salaires, historiquement environ 1 point au-dessus des biens. Et un budget de retraité est plus servi que la moyenne.
- **Ce qui baisse ne le concerne plus autant.** La déflation des biens technologiques et manufacturés tire l'indice vers le bas, mais elle pèse peu dans un panier de retraité.

L'ordre de grandeur qui en sort est cohérent avec les études étrangères sur les indices « seniors ». Le CPI-E américain court +0,2-0,3 point/an au-dessus du CPI général sur longue période. L'inflation d'un ménage retraité dépasse ainsi l'indice de ~0,2 à 0,5 point par an, davantage aux grands âges. Composé sur trente ans, +0,3 point vaut ~9-10 % de pouvoir d'achat d'écart, soit une année et demie de dépenses. Ce n'est pas un raffinement, c'est un poste du plan.

**Estimer la vôtre en une soirée.** Reprenez vos 12-24 mois de relevés déjà classés ([[combien-il-vous-faut]]). Pondérez vos catégories réelles. Croisez-les avec les indices INSEE par fonction, santé, services, alimentation, énergie, tous publiés. La moyenne pondérée est votre inflation rétrospective, et le simulateur d'inflation personnalisée de l'INSEE fait le calcul en ligne. L'objectif n'est pas la décimale. C'est de savoir si vous êtes un ménage « indice », « indice +0,3 » ou « indice +0,6 », et de le donner au plan.

::: astuce Les points morts : la seule prévision qui engage de l'argent
Où lire ce que « le marché » anticipe ? Dans les **points morts** d'inflation ([[obligations-indexees]]), l'écart entre taux nominaux et taux réels des obligations d'État. Le breakeven 10 ans français/euro (~2 %) est l'anticipation moyenne d'investisseurs qui misent des milliards dessus. Elle est imparfaite, car elle contient des primes de risque et de liquidité. Mais elle reste infiniment plus disciplinée que les sondages et les éditoriaux. Son usage pour le plan n'est **pas** le market timing, c'est un étalonnage. Si votre simulation suppose implicitement 2 % et que les breakevens montent durablement vers 2,7 %, c'est un fait nouveau qui mérite la revue annuelle. Et si quelqu'un vous vend un produit « parce que l'hyperinflation arrive », les breakevens sont la réponse polie ([[hyperinflation-et-extremes]]).
:::

## Le suivi pratique, et le réglage du plan

**Le tableau de bord minimal** tient en dix minutes à la revue annuelle ([[revue-annuelle]]). Premier chiffre, l'IPCH France des douze derniers mois (publication mensuelle INSEE/Eurostat, c'est lui qui a indexé vos retraits). Deuxième chiffre, votre inflation personnelle rétrospective, le calcul d'une soirée refait d'un coup d'œil. Troisième chiffre, le breakeven 10 ans, l'étalonnage des anticipations. Tout le reste est du bruit pour un plan de 40 ans, le chiffre du mois, le débat core/headline, les gros titres. L'inflation se pilote à l'année, pas au flash mensuel.

**Sur quoi indexer les retraits ?** La règle fixe canonique indexe sur l'IPC ([[retrait-fixe-bengen]]). C'est le bon **défaut**, observable, incontestable, cohérent avec les pensions et les linkers. Il se complète par une dérive explicite. Le réglage propre n'est pas « j'indexe sur IPC + 0,4 » (invérifiable), mais « j'indexe sur l'IPC et je budgète une dérive réelle des dépenses ». C'est ce qu'un simulateur fait littéralement. Le curseur **« Real spending drift /yr »** ajoute une pente réelle aux dépenses (+0,3 à +0,5 %/an est la valeur de planification recommandée pour la dérive santé), et le « retirement smile » module le profil par âge ([[depenses-en-retraite]], [[utiliser-la-page-fire]]). Un plan simulé **sans** dérive sur 45 ans suppose que votre panier suivra exactement l'indice moyen national pendant un demi-siècle. C'est l'hypothèse optimiste déguisée la plus courante du sujet.

**Et rappelez-vous ce qu'un simulateur fait déjà.** Il travaille en **réel** de bout en bout, avec des séries déflatées par l'IPCH et des retraits en pouvoir d'achat constant ([[la-machine-pofo]]). L'inflation **moyenne** est donc dans la machine par construction. Ce que les curseurs ajoutent, c'est **votre** écart à la moyenne, la dérive. Et ce que les modèles de régime testent, c'est le **risque** d'épisode ([[inflation-et-taux-de-retrait]] pour la mécanique complète).

::: exemple L'inflation personnelle de Denise et Paul
Denise (63 ans) et Paul (66 ans) sont propriétaires toit payé, avec 46 000 €/an de dépenses. Voici leur soirée de calcul. Santé et mutuelle pèsent 14 % du budget (indice perso de la catégorie, +4,5 %/an). Services, aide et assurances font 22 % (+3 %). L'alimentation fait 18 % (+2 %). L'énergie et le transport font 12 % (+2,5 % volatil). Les loisirs et voyages font 24 % (+2 %). Le divers fait 10 % (+2 %). L'inflation personnelle pondérée ressort à ~2,7 % quand l'IPCH fait 2,1 %. L'écart est de +0,6, dont l'essentiel vient de la santé, cohérent avec leur âge. Côté réglage du plan, ils indexent les retraits sur l'IPC (contractuel), fixent le spendDrift à +0,4 %, et activent le sourire (la dérive santé nette du ralentissement des voyages après 80 ans). La simulation bouge, la ruine centrale passe de 3,9 à 4,8 %. C'est le vrai prix de leur panier, qu'aucun indice national ne leur aurait facturé. Et il vaut mieux le connaître à 63 ans qu'à 83.
:::

## L'essentiel à retenir

- Trois inflations : l'officielle (IPC/IPCH, celle des pensions, des linkers, et du déflateur `^HICP-FR`), la vôtre (votre panier, typiquement indice +0,2 à +0,5 pour un retraité, santé et services en tête), et celle du plan (officielle + dérive explicite).
- Les indices européens sont sérieux, mais ils moyennent un panier qui n'est pas le vôtre. L'angle mort principal est le logement du propriétaire, quasi absent de l'IPCH. Un toit payé adoucit **votre** inflation, l'indice ne le sait pas.
- L'écart composé compte. +0,3 point/an sur 30 ans ≈ 10 % de pouvoir d'achat, soit un an et demi de dépenses. Estimez votre panier en une soirée (relevés × indices INSEE par fonction, ou le simulateur INSEE).
- Le tableau de bord annuel tient en trois chiffres : IPCH 12 mois, votre inflation rétrospective, le breakeven 10 ans (la seule anticipation qui engage de l'argent, un étalonnage, jamais un signal de timing).
- Le réglage propre du plan tient en une ligne. Indexer sur l'IPC et budgéter la dérive (spendDrift +0,3-0,5 %, sourire activé). Un plan de 45 ans sans dérive suppose un panier moyen éternel, c'est l'optimisme le mieux déguisé du sujet.

---

## Pour aller plus loin

- [insee.fr](https://www.insee.fr) : l'IPC mensuel, les indices par fonction (COICOP) et le simulateur d'inflation personnalisée, les outils de la soirée de calcul.
- Eurostat : l'IPCH par pays et le détail méthodologique (dont le dossier du logement du propriétaire, l'OOH).
- Le Boskin Report (1996) et ses suites : le débat qualité/substitution, version documentée.
- BLS, le CPI-E américain : la référence sur l'inflation des seniors.
- Dans ce livre : [[inflation-et-taux-de-retrait]] (l'effet exact sur le plan), [[depenses-en-retraite]] (la dérive et le sourire), [[obligations-indexees]] (les points morts en pratique).
