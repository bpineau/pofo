# Horizon, espérance de vie et retraites de 50 ans

Combien de temps votre plan doit-il durer ? La question paraît triviale, elle est piégée trois fois.

Premier piège : presque tout le monde raisonne sur la mauvaise espérance de vie (celle de la naissance, ou une moyenne, alors qu'il faut un quantile prudent, conditionnel à votre âge, et pour le **dernier** survivant du couple). Deuxième piège : l'intuition « horizon deux fois plus long = il faut deux fois plus de prudence » est fausse : la relation entre horizon et taux de retrait s'aplatit remarquablement au-delà de 40 ans, et comprendre pourquoi change la façon de voir une retraite précoce.

Troisième piège : l'horizon d'un plan FIRE n'est pas homogène. Il contient une phase à découvert (avant les pensions) et une phase adossée, et c'est la première qui gouverne le risque. Cette page traite les trois : comment choisir **votre** horizon avec les vraies tables de mortalité françaises, ce que l'horizon fait exactement au taux de retrait (avec les chiffres), et comment lire la vue mortalité, celle qui répond à la seule question qui compte vraiment : quelle est la probabilité d'être un jour vivant et ruiné ?

::: cle Les trois règles à retenir
1) Planifiez sur un quantile de survie prudent (le 85e-90e percentile du dernier survivant), jamais sur l'espérance : la moitié des gens vivent plus longtemps que la moyenne, par définition. 2) Le taux de retrait sûr baisse vite de 30 à 40 ans d'horizon, puis presque plus : un plan qui tient 45 ans tient à peu près perpétuellement. 3) Le vrai risque de longévité ne se gère pas en allongeant l'horizon du simulateur à l'infini, mais avec des actifs qui paient tant que vous vivez : pensions et rentes ([[retraite-legale]], [[rentes-et-annuites]]).
:::

## L'espérance de vie, correctement

**Erreur n° 1 : l'espérance à la naissance.** « L'espérance de vie est de 80 ans (hommes) / 86 ans (femmes) » : ces chiffres INSEE incluent les décès précoces. Or vous avez déjà survécu jusqu'à aujourd'hui : l'espérance **conditionnelle** à votre âge est nettement plus haute, et elle monte à mesure que vous vieillissez. Ordres de grandeur des tables françaises récentes (INSEE, arrondis) :

| À l'âge de... | Homme : vivra en moyenne jusqu'à | Femme : vivra en moyenne jusqu'à |
|---|---|---|
| Naissance | ~80 ans | ~86 ans |
| 45 ans | ~82 ans | ~87 ans |
| 65 ans | ~84 ans | ~88 ans |
| 85 ans | ~91 ans | ~92 ans |

**Erreur n° 2 : la moyenne.** L'espérance est le centre d'une distribution étalée : un homme de 45 ans a environ une chance sur quatre d'atteindre 92 ans, une femme sur quatre d'atteindre 95. Planifier sur la moyenne, c'est accepter ~50 % de risque de survivre à son plan, une définition étrange de la prudence. La convention des actuaires et des bons planificateurs : le **85e-90e percentile de survie**.

**Erreur n° 3 : l'individu au lieu du couple.** Pour un couple, ce qui compte est la durée jusqu'au décès du **dernier** survivant : le portefeuille doit durer tant que l'un des deux respire. Les probabilités se composent : pour un couple de 45 et 45 ans, la probabilité qu'**au moins** l'un des deux atteigne 95 ans dépasse 40 %, et le 90e percentile du dernier survivant se situe vers 98-100 ans. C'est cette mécanique de « dernier survivant » que la section §05 de la page FIRE utilise (mortalité d'un couple français, [[utiliser-la-page-fire]]).

**Deux correctifs de calibration** pour finir le travail proprement. Le gradient socio-économique : les tables nationales moyennent des populations très hétérogènes ; les cadres vivent 5 à 6 ans de plus que les ouvriers (INSEE), et le profil type du FIRE (éducation, revenus, patrimoine, accès aux soins, non-pénibilité) se situe dans le haut du gradient : ajoutez 2 à 4 ans aux tables. Et la dérive générationnelle : les tables prospectives (TGH-05/TGF-05, utilisées par les assureurs précisément pour les rentes) intègrent l'amélioration continue de la mortalité ; un quadragénaire de 2026 doit raisonner sur les longévités de 2070, pas sur celles des décédés de 2020.

::: exemple Le calcul d'horizon d'un couple type
Léa (44 ans) et Sam (46 ans), cadres, départ prévu à 47/49 ans. Dernier survivant : 90e percentile vers 98 ans pour Léa, correction socio-économique + prudence prospective → planifier jusqu'aux **100 ans de Léa**, soit un horizon de plan de **53 ans**. Dans pofo : horizon 53 (ou le plafond du curseur), âge 47, et la vue §05 fera le reste. Elle pondère chaque année de ruine potentielle par la probabilité qu'au moins l'un des deux soit encore là pour la subir. On verra plus bas pourquoi ce 53 n'est pas deux fois plus effrayant que le 30 du retraité classique.
:::

## Ce que l'horizon fait au taux de retrait : la courbe qui s'aplatit

Passons au deuxième piège, le plus contre-intuitif et le plus libérateur. Voici, en ordres de grandeur, le taux de retrait rigide qui aurait survécu à tous les millésimes américains (le SAFEMAX de Bengen, [[etude-trinity]]), en fonction de l'horizon, tel que la série d'ERN l'a systématisé sur données mensuelles ([[serie-ern]]) :

| Horizon | SAFEMAX approximatif (60-75 % actions) | Baisse vs ligne précédente |
|---|---|---|
| 30 ans | ~4,0-4,15 % | (référence) |
| 40 ans | ~3,5-3,6 % | −0,5 point |
| 50 ans | ~3,35-3,45 % | −0,15 point |
| 60 ans | ~3,25-3,4 % | −0,05 point |
| Perpétuité (capital préservé) | ~3,25 % | ≈ 0 |

::: figure horizon-flatten
Le taux de retrait qui aurait survécu à tous les millésimes, selon l'horizon (ordres de grandeur, données américaines). L'essentiel du durcissement se joue entre 30 et 40 ans ; au-delà, la courbe s'aplatit vers une asymptote (~3,25 %, le rendement géométrique de croisière). Un plan qui traverse ses 30-40 premières années est donc presque devenu une perpétuité, et prendre un horizon très prudent ne coûte presque rien.
:::

La lecture saute aux yeux : **l'essentiel du durcissement se produit entre 30 et 40 ans d'horizon ; au-delà, la courbe est presque plate**. Passer d'une retraite de 50 ans à une retraite de 60 ans ne coûte quasiment rien ; c'est passer de 30 à 40 qui coûte. Pourquoi cette asymptote ? Le mécanisme est lumineux une fois vu : un retrait de ~3,25 % est **inférieur** au rendement réel géométrique d'un portefeuille diversifié (~3,5-4,5 %, [[rendements-attendus]]). À ce niveau, dans la grande majorité des scénarios, le portefeuille croît **plus vite** qu'on ne le ponctionne : au bout de 30-35 ans, il a doublé ou triplé en termes réels, et le taux de retrait courant est descendu à 1-2 %. Un plan qui a survécu à ses 30-35 premières années n'est plus en risque. Il est devenu une perpétuité. L'horizon supplémentaire n'ajoute du risque que dans les scénarios déjà mauvais au départ, et ceux-là échouent de toute façon dans les 30 premières années. Le risque d'un plan long n'est donc pas « durer », c'est **traverser le début** : on retombe, encore, sur la fenêtre fragile du risque de séquence ([[sequence-des-rendements]]).

Trois corollaires pratiques de cette asymptote :

**« Je pars à 40 ans » n'est pas deux fois plus dur que « je pars à 60 ans ».** C'est environ 0,6-0,8 point de taux de retrait plus dur (4 % → ~3,3 %), soit un multiple de 30x au lieu de 25x ([[combien-il-vous-faut]]) : substantiel, mais fini, et largement compensable par les marges spécifiques du jeune retraité (employabilité, pension à venir, flexibilité, [[flexibilite-realite]]).

**L'incertitude sur votre longévité est un problème de second ordre pour le portefeuille.** Entre planifier 50 ou 60 ans, la différence de taux est ~0,1 point : vous pouvez prendre le quantile très prudent sans surcoût réel. C'est une excellente nouvelle : l'inconnue la plus angoissante du plan (combien de temps vivrai-je ?) est celle qui coûte le moins cher à couvrir... côté portefeuille. Côté **dépenses** de fin de vie, c'est une autre histoire ([[depenses-en-retraite]]).

**Le « capital préservé » est presque gratuit au-delà de 45 ans d'horizon.** À ~3,25 %, le portefeuille se maintient en termes réels dans la plupart des mondes : viser la préservation (pour transmettre, [[succession-et-transmission]], ou par sécurité pure) plutôt que l'épuisement ne renchérit un plan long que marginalement. ERN appelle ça la convergence entre « capital depletion » et « capital preservation » aux horizons FIRE.

::: attention Le piège inverse : l'horizon trop court
L'erreur symétrique existe et elle est plus grave : le retraité de 65 ans qui planifie « 25 ans, jusqu'à 90 ans » alors que sa femme a 62 ans et une chance sur trois de dépasser 95. Pour lui, l'horizon est dans la zone pentue de la courbe (25 → 35 ans coûte cher), et le tronquer de dix ans surestime le taux soutenable de 0,5 point ou plus. La règle : c'est aux horizons **courts** que le choix de l'horizon est critique ; aux horizons FIRE, il est presque indolore. Planifiez long, ça ne coûte rien ; tronquez, ça peut coûter tout.
:::

## La ruine pondérée par la mortalité : la vue qui remet les chiffres à leur place

La probabilité de ruine standard traite la mortalité comme une insulte. Elle suppose que vous êtes vivant pour constater la faillite à 99 ans ([[ruine-et-probabilites]]). Or la question réelle est conjointe : **quelle est la probabilité d'être un jour vivant ET ruiné ?** C'est ce que calcule la section §05 de la page FIRE (« Alive, broke or gone ») : à chaque année de chaque scénario, trois états possibles (vivant et solvable, vivant et ruiné, décédé), les probabilités de décès venant des tables d'un couple français.

L'effet de cette pondération est systématiquement apaisant, et il est juste qu'il le soit : une ruine « à 40 ans de plan » (à 87-92 ans pour notre couple type) n'est subie que si quelqu'un est encore là, ce qui n'arrive que dans une fraction des cas. En pratique, la ruine mortalité-pondérée ressort typiquement à la **moitié** ou au **tiers** de la ruine brute pour des plans dont les échecs sont tardifs, et proche de la ruine brute pour des plans dont les échecs sont précoces. C'est donc aussi un excellent révélateur du **profil** temporel de votre risque : si la pondération mortalité ne réduit presque rien, c'est que vos échecs arrivent tôt, et c'est un signal sérieux ([[sequence-des-rendements]]).

Faut-il alors planifier directement sur la ruine pondérée ? Non : la bonne pratique est de dimensionner sur la ruine brute à horizon prudent (elle protège aussi le conjoint survivant et les imprévus de longévité), et d'utiliser la vue pondérée pour **arbitrer** les cas limites : un plan à 8 % de ruine brute dont la version pondérée tombe à 2,5 % avec des échecs après 85 ans, pension acquise, est un plan acceptable qu'une lecture brute aurait rejeté au prix d'années de travail en trop ([[une-annee-de-plus]]).

Et pour la queue de longévité extrême (centenaire), le bon outil n'est pas le portefeuille. C'est l'actif qui paie **tant que** vous vivez, exactement calibré sur le risque. Votre pension de retraite légale est déjà une rente viagère indexée ([[retraite-legale]]) ; une annuité complémentaire achetée tard (75-80 ans, quand la mutualisation joue à plein) est le complément technique ([[rentes-et-annuites]]). Le curseur « Annuitise % of capital » permet de tester cet arbitrage : la rente dégrade souvent la ruine **moyenne** (on convertit des actifs de croissance en revenu plancher) mais elle écrase la queue « vivant, ruiné, 95 ans », qui est précisément le scénario qu'on cherche à éliminer.

## L'horizon FIRE n'est pas homogène : la phase à découvert

Dernier piège, spécifique aux retraites précoces. L'horizon de 50 ans d'un plan FIRE se décompose en réalité en deux régimes très différents :

**La phase à découvert** (du départ à la liquidation des pensions, typiquement 15-25 ans) : le portefeuille finance **tout** : c'est ici que se concentrent la fenêtre fragile, le gros des retraits nets, et l'essentiel du risque. Le taux de retrait courant y est à son maximum.

**La phase adossée** (des pensions au décès) : la pension couvre une fraction substantielle du plancher de dépenses ([[revenus-complementaires]]), le retrait net s'effondre, et le portefeuille n'a plus qu'un rôle de complément et de réserve. Pour beaucoup de plans français réalistes, le taux de retrait net de cette phase passe sous 2 % : hors de danger structurel.

Cette décomposition explique un phénomène qu'on observe systématiquement en simulation : **ajouter la pension au modèle transforme le plan** (souvent 2 à 4 fois moins de ruine, [[erreurs-classiques-fire]]), parce qu'elle raccourcit l'horizon **effectif** du risque de 50 ans à la seule phase à découvert. Le plan FIRE de 50 ans bien construit est en fait un pont de 20 ans vers un régime de croisière adossé, plus une réserve. D'où deux règles de conception : dimensionnez la solidité sur la phase à découvert (c'est elle que les millésimes hostiles doivent traverser), et vérifiez la phase adossée surtout sur le risque d'inflation, le seul qui menace une pension et un plancher à 30 ans de distance ([[inflation-et-taux-de-retrait]], [[se-proteger-de-inflation]]).

::: terrain Ce que ça change psychologiquement
Les FIRE en cours de route le disent souvent : reformuler « mon plan doit tenir 50 ans » en « mon plan doit traverser 18 ans à découvert, ensuite les pensions prennent le relais du plancher » change la charge mentale du tout au tout. Le premier énoncé est un gouffre inimaginable ; le second est un projet borné, avec des jalons (chaque année de traversée en réduit la longueur, chaque trimestre validé compte, [[retraite-legale]]). La bonne granularité psychologique du pilotage n'est pas la vie entière, c'est la traversée ([[psychologie-du-retrait]], [[quand-s-inquieter]]).
:::

## L'essentiel à retenir

- La bonne cible d'horizon : le 85e-90e percentile de survie du **dernier** survivant, conditionnel à vos âges, corrigé du gradient socio-économique (+2 à 4 ans pour le profil FIRE type) : pour un couple de quadragénaires, planifier jusqu'à ~100 ans, soit 50-55 ans d'horizon.
- La courbe taux-horizon s'aplatit : ~4 % à 30 ans, ~3,5 % à 40, ~3,3 % à 50-60 et en perpétuité ; un plan qui survit à ses 30-35 premières années est devenu une perpétuité. Le risque d'un plan long, c'est son début, pas sa longueur.
- Corollaires : prendre le quantile très prudent ne coûte presque rien aux horizons FIRE ; tronquer l'horizon est surtout dangereux pour les retraites **courtes** ; la préservation du capital est presque gratuite au-delà de 45 ans.
- La vue « alive, broke or gone » (§05) répond à la vraie question (vivant et ruiné) : dimensionnez sur la ruine brute, arbitrez les cas limites sur la pondérée, et traitez la queue de longévité avec des rentes, pas avec du capital en plus.
- Un plan FIRE de 50 ans = une traversée à découvert de 15-25 ans + une phase adossée aux pensions : dimensionnez la solidité sur la première, l'anti-inflation sur la seconde.

---

## Pour aller plus loin

- INSEE : tables de mortalité françaises ([insee.fr](https://www.insee.fr), « tables de mortalité des années N ») et études sur les écarts d'espérance de vie par catégorie sociale ; les tables prospectives TGH-05/TGF-05 pour la vision assureur.
- Early Retirement Now, SWR Series volets 1-2 (les SAFEMAX par horizon, dépletion vs préservation) et volet 56 (rentes et Sécurité sociale dans le plan) ([[serie-ern]]).
- Moshe Milevsky, *The 7 Most Important Equations for Your Retirement* : la formalisation actuarielle accessible (dont l'équation de Fibonacci de la longévité et la logique de l'annuitisation).
- Dans pofo : le curseur d'horizon, l'âge (qui pilote §05) et « Annuitise % of capital » ([[utiliser-la-page-fire]]) ; la mécanique interne : [[la-machine-pofo]].
