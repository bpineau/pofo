# Suivre l'inflation : les indices, et la vôtre

Tout le plan vit en euros constants : les retraits s'indexent, les simulations tournent en réel, le plancher est un pouvoir d'achat ([[la-machine-pofo]]). Mais constants par rapport à **quoi** ?

« L'inflation » n'est pas un fait brut. C'est une mesure, construite, avec un panier qui n'est pas le vôtre. Et l'écart entre l'indice officiel et **votre** inflation personnelle, petit chaque année, composé sur trente ans, peut valoir autant qu'un point de taux de retrait. Cet article est le manuel de mesure : les indices qui comptent pour un Français (IPC, IPCH, celui qui sert de déflateur, leur construction, leurs angles morts, dont le plus important, le logement du propriétaire), le débat honnête sur leur fiabilité, la question centrale de l'inflation **personnelle** (pourquoi celle d'un retraité dérive au-dessus de l'indice, comment estimer la vôtre en une soirée), les anticipations de marché (les points morts, la seule « prévision » d'inflation qui engage de l'argent), et la traduction opérationnelle : sur quoi indexer ses retraits, et comment régler la dérive de dépenses dans pofo.

::: cle Les trois inflations à distinguer
L'inflation **officielle** (l'IPC/IPCH, celle des statistiques, des pensions et des linkers) ; l'inflation de **votre panier** (santé, services, énergie, loisirs pondérés à votre façon, typiquement +0,2 à +0,5 point au-dessus de l'officielle pour un retraité) ; et l'inflation de votre **plan** (celle que vous choisissez d'appliquer à vos retraits, officielle + une dérive santé explicite, le réglage honnête). Confondre les trois fait des plans qui semblent tenir en euros constants... des mauvais euros.
:::

## Les indices : ce qu'ils sont, ce qu'ils contiennent

**L'IPC de l'INSEE** : l'indice des prix à la consommation français : ~1 100 familles de produits, ~200 000 relevés mensuels, pondérés par la consommation moyenne des ménages. Sa variante « hors tabac » sert de référence légale (c'est elle qui indexe le SMIC, les pensions... et les OATi, [[obligations-indexees]]).

**L'IPCH (HICP)** : l'indice **harmonisé** européen, la version comparable entre pays de la zone. C'est la cible de la BCE (2 %), l'indice des OAT€i, et **la série qui sert à déflater** (le symbole `^HICP-FR`, toutes les conversions en réel du simulateur passent par lui, [[la-machine-pofo]]). IPC et IPCH français divergent peu (méthodes proches ; l'IPCH pondère un peu différemment, notamment la santé nette des remboursements) : quelques dixièmes certaines années.

**Ce que le panier contient mal : le logement du propriétaire.** Le point technique le plus important : l'IPCH ne compte le logement que par les **loyers** effectifs (~7 % du panier français, le poids des locataires) : le coût du logement du **propriétaire** (le prix des logements, l'équivalent-loyer) n'y est pratiquement pas : quand l'immobilier double, l'IPCH ne bouge pas. Conséquences pour vous : à la baisse, un rentier **propriétaire** toit payé ([[immobilier-en-retrait]]) a une inflation personnelle **structurellement** plus douce que l'indice (son plus gros poste est éteint) ; à la hausse, l'épargnant en phase d'accumulation qui vise un achat voit sa vraie inflation-cible galoper sans que l'indice le dise. Même logique pour ce que l'indice moyenne et que vous ne consommez pas.

**Les corrections de qualité (effets hédoniques)** : quand le smartphone de 2026 fait plus que celui de 2020 au même prix, l'indice compte une **baisse** de prix : méthodologiquement défendable (on mesure le prix du service rendu), pratiquement débattu (vous ne pouvez pas acheter « moins de qualité », la dépense ne baisse pas). Le débat honnête sur la fiabilité tient en deux phrases : les indices officiels européens sont sérieux, audités, sans manipulation démontrée. Et ils mesurent un panier moyen à qualité corrigée qui peut s'écarter durablement de l'expérience de dépense d'un ménage donné : les deux propositions sont vraies. D'où la section suivante.

## Votre inflation : pourquoi celle d'un retraité dérive au-dessus

La dispersion des inflations personnelles autour de l'indice est documentée (l'INSEE publie des indices par catégorie de ménage et un simulateur d'inflation personnalisée) : quelques dixièmes de point par an entre ménages types. Et le profil **retraité** cumule les pondérations défavorables :

- **La santé et la dépendance** : le poste qui croît en part du budget avec l'âge, et dont les prix (mutuelle surtout, +4-6 %/an tendanciels, portés par l'âge et la dérive des coûts médicaux, puis établissements et aide à domicile) courent structurellement au-dessus de l'IPC ([[sante-et-protection-sociale]], [[depenses-en-retraite]]).
- **Les SERVICES en général** (aide, entretien, restauration, assurances) : intensifs en main-d'œuvre locale, leurs prix suivent les salaires, historiquement ~1 point au-dessus des biens. Et un budget de retraité est plus servi que la moyenne.
- **Ce qui baisse ne le concerne plus autant** : la déflation des biens technologiques et manufacturés, qui tire l'indice vers le bas, pèse peu dans un panier de retraité.

L'ordre de grandeur qui en sort, cohérent avec les études étrangères sur les indices « seniors » (le CPI-E américain, +0,2-0,3 point/an au-dessus du CPI général sur longue période) : **l'inflation d'un ménage retraité dépasse l'indice de ~0,2 à 0,5 point par an**, davantage aux grands âges. Composé sur trente ans, +0,3 point = ~9-10 % de pouvoir d'achat d'écart : l'équivalent d'une année et demie de dépenses : ce n'est pas un raffinement, c'est un poste du plan.

**Estimer la vôtre en une soirée** : reprenez vos 12-24 mois de relevés déjà classés ([[combien-il-vous-faut]]) ; pondérez vos catégories réelles ; croisez avec les indices INSEE par fonction (santé, services, alimentation, énergie, tous publiés) : la moyenne pondérée est votre inflation rétrospective ; et le simulateur d'inflation personnalisée de l'INSEE fait le calcul en ligne. L'objectif n'est pas la décimale. C'est de savoir si vous êtes un ménage « indice », « indice +0,3 » ou « indice +0,6 », et de le donner au plan.

::: astuce Les points morts : la seule prévision qui engage de l'argent
Où lire ce que « le marché » anticipe ? Dans les **points morts** d'inflation ([[obligations-indexees]]) : l'écart entre taux nominaux et taux réels des obligations d'État : le breakeven 10 ans français/euro (~2 %) est l'anticipation moyenne d'investisseurs qui misent des milliards dessus : imparfaite (elle contient des primes de risque et de liquidité) mais infiniment plus disciplinée que les sondages et les éditoriaux. Usage pour le plan : **pas** de market timing : un étalonnage : si votre simulation suppose implicitement 2 % et que les breakevens montent durablement vers 2,7 %, c'est un fait nouveau qui mérite la revue annuelle ; et si quelqu'un vous vend un produit « parce que l'hyperinflation arrive », les breakevens sont la réponse polie ([[hyperinflation-et-extremes]]).
:::

## Le suivi pratique, et le réglage du plan

**Le tableau de bord minimal** (dix minutes à la revue annuelle, [[revue-annuelle]]) : l'IPCH France des douze derniers mois (publication mensuelle INSEE/Eurostat, c'est le chiffre qui a indexé vos retraits) ; votre inflation personnelle rétrospective (le calcul d'une soirée, refait d'un coup d'œil) ; et le breakeven 10 ans (l'étalonnage des anticipations). Tout le reste : le chiffre du mois, le débat core/headline, les gros titres : est du bruit pour un plan de 40 ans : l'inflation se pilote à l'année, pas au flash mensuel.

**Sur quoi indexer les retraits ?** La règle fixe canonique indexe sur l'IPC ([[retrait-fixe-bengen]]). C'est le bon **défaut** (observable, incontestable, cohérent avec pensions et linkers) : complété par la dérive explicite : le réglage propre n'est pas « j'indexe sur IPC + 0,4 » (invérifiable) mais « j'indexe sur l'IPC et je budgète une dérive réelle des dépenses » : ce que pofo fait littéralement : le curseur **« Real spending drift /yr »** ajoute une pente réelle aux dépenses (+0,3 à +0,5 %/an est la valeur de planification recommandée pour la dérive santé), et le « retirement smile » module le profil par âge ([[depenses-en-retraite]], [[utiliser-la-page-fire]]). Un plan simulé **sans** dérive sur 45 ans suppose que votre panier suivra exactement l'indice moyen national pendant un demi-siècle. C'est l'hypothèse optimiste déguisée la plus courante du sujet.

**Et rappelez-vous ce que le simulateur fait déjà** : pofo travaille en **réel** de bout en bout (séries déflatées par l'IPCH, retraits en pouvoir d'achat constant, [[la-machine-pofo]]) : l'inflation **moyenne** est donc dans la machine par construction ; ce que les curseurs ajoutent, c'est **votre** écart à la moyenne (la dérive) ; et ce que les modèles de régime testent, c'est le **risque** d'épisode ([[inflation-et-taux-de-retrait]] pour la mécanique complète).

::: exemple L'inflation personnelle de Denise et Paul
Denise (63 ans) et Paul (66 ans), propriétaires toit payé, 46 000 €/an. Leur soirée de calcul : santé-mutuelle 14 % du budget (indice perso de la catégorie, +4,5 %/an), services-aide-assurances 22 % (+3 %), alimentation 18 % (+2 %), énergie-transport 12 % (+2,5 % volatil), loisirs-voyages 24 % (+2 %), divers 10 % (+2 %) : inflation personnelle pondérée ~2,7 % quand l'IPCH fait 2,1 % : écart +0,6, dont l'essentiel est la santé : cohérent avec leur âge. Réglage du plan : indexation des retraits sur IPC (contractuel), spendDrift à +0,4 %, et le sourire activé (la dérive santé nette du ralentissement des voyages après 80 ans). La simulation bouge : ruine centrale de 3,9 à 4,8 %. C'est le vrai prix de leur panier, qu'aucun indice national ne leur aurait facturé. Et il vaut mieux le connaître à 63 ans qu'à 83.
:::

## L'essentiel à retenir

- Trois inflations : l'officielle (IPC/IPCH, celle des pensions, des linkers, et du déflateur `^HICP-FR`), la vôtre (votre panier, typiquement indice +0,2 à +0,5 pour un retraité, santé et services en tête), et celle du plan (officielle + dérive explicite).
- Les indices européens sont sérieux et moyennent un panier qui n'est pas le vôtre : l'angle mort principal est le logement du propriétaire (quasi absent de l'IPCH) : un toit payé adoucit **votre** inflation, l'indice ne le sait pas.
- L'écart composé compte : +0,3 point/an sur 30 ans ≈ 10 % de pouvoir d'achat ≈ un an et demi de dépenses : estimez votre panier en une soirée (relevés × indices INSEE par fonction, ou le simulateur INSEE).
- Le tableau de bord annuel tient en trois chiffres : IPCH 12 mois, votre inflation rétrospective, le breakeven 10 ans (la seule anticipation qui engage de l'argent, un étalonnage, jamais un signal de timing).
- Le réglage propre du plan : indexer sur l'IPC et budgéter la dérive (spendDrift +0,3-0,5 %, sourire activé) : un plan de 45 ans sans dérive suppose un panier moyen éternel. C'est l'optimisme le mieux déguisé du sujet.

---

## Pour aller plus loin

- [insee.fr](https://www.insee.fr) : l'IPC mensuel, les indices par fonction (COICOP) et le simulateur d'inflation personnalisée : les outils de la soirée de calcul.
- Eurostat : l'IPCH par pays et le détail méthodologique (dont le dossier du logement du propriétaire, l'OOH).
- Le Boskin Report (1996) et ses suites : le débat qualité/substitution, version documentée.
- BLS, le CPI-E américain : la référence sur l'inflation des seniors.
- Dans ce livre : [[inflation-et-taux-de-retrait]] (l'effet exact sur le plan), [[depenses-en-retraite]] (la dérive et le sourire), [[obligations-indexees]] (les points morts en pratique).
