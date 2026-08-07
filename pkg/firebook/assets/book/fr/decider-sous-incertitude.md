# Décider sous incertitude : utilité, Kelly, regret et décisions robustes

Ce livre passe beaucoup de temps à répondre à des questions du type « quelle stratégie maximise le taux de retrait » ou « quelle allocation minimise la ruine ». Cet article prend un cran de recul et pose la question d'avant : **qu'est-ce qu'une bonne décision quand l'avenir est incertain ?** Car le rentier a un problème que les moyennes ignorent : il ne vit qu'une seule trajectoire. Le simulateur en tire dix mille, la vie en tire une, sans nouvelle partie. Maximiser une espérance calculée sur dix mille mondes n'est pas automatiquement le bon objectif pour quelqu'un qui n'en habitera qu'un.

La théorie de la décision réunit un siècle de travaux, entre économie, mathématiques et psychologie. Elle répond précisément à cette question. Vous verrez d'abord pourquoi l'espérance de gain est un mauvais critère, ce que corrige la notion d'utilité. Puis comment comparer deux plans risqués en un seul chiffre honnête, l'équivalent certain. Ensuite pourquoi le critère de Kelly, si séduisant, ne s'applique pas tel quel à une retraite. Enfin ce que valent le regret et la robustesse comme critères d'adulte, et comment tout cela se condense en un protocole de cinq règles. C'est l'article le plus conceptuel du livre, et peut-être le plus rentable. Les erreurs de critère coûtent plus cher que les erreurs de calcul.

::: cle L'idée en une phrase
Entre deux plans, ne choisissez pas celui qui a la meilleure moyenne. Choisissez celui dont vous préférez la distribution complète, pondérée par ce que chaque niveau de richesse vaut réellement pour vous. Un euro de plus quand tout va bien ne vaut presque rien ; un euro de plus dans une trajectoire ruinée vaut tout. Toute la théorie de la décision tient dans cette asymétrie, et presque toutes les recommandations prudentes de ce livre en découlent mécaniquement.
:::

## L'utilité : pourquoi la moyenne est un mauvais juge

Le point de départ classique est un paradoxe de 1738, celui de Saint-Pétersbourg. C'est un jeu de pile ou face à espérance de gain infinie, que personne ne paierait pourtant plus de quelques dizaines d'euros pour jouer. La résolution de Daniel Bernoulli fonde toute la suite. Ce qui compte n'est pas l'argent mais l'**utilité** de l'argent, et celle-ci croît de moins en moins vite. Passer de 500 000 € à 1 M€ change une vie ; passer de 10 M€ à 10,5 M€ ne change rien. Cette **concavité** a une conséquence directe. À espérance égale, moins de dispersion vaut plus d'utilité. L'aversion au risque n'est pas une émotion, mais un théorème.

Pour le rentier, la concavité est extrême et asymétrique. La zone basse de sa distribution détruit le mode de vie, la dignité, les options. Ce sont les trajectoires qui frôlent ou touchent la ruine ([[ruine-et-probabilites]]). La zone haute, elle, mourir sur un tas d'or, n'ajoute presque rien, sinon un héritage plus gros ([[depenses-en-retraite]]). Un critère qui pèse ces deux zones symétriquement, comme la moyenne, est donc structurellement faux pour lui. D'où une habitude constante dans ce livre. On juge les plans sur le percentile 5 et la médiane, jamais sur la moyenne ([[lire-un-fan-chart]]).

L'outil qui rend tout cela opérationnel s'appelle l'**équivalent certain**. C'est le montant garanti que vous accepteriez en échange d'une distribution risquée. Prenez deux plans de même moyenne. L'un donne 40 000 €/an à coup sûr, l'autre entre 25 000 et 60 000 selon les marchés. Pour une personne raisonnablement prudente, le second « vaut » peut-être 36 000 certains. L'équivalent certain convertit ainsi n'importe quelle loterie en un chiffre comparable, et il baisse d'autant plus vite que la queue basse est laide. Nul besoin de le calculer formellement pour l'utiliser. Se demander « à quel revenu garanti je préférerais ce plan risqué ? » est déjà le bon geste mental, et cela dégonfle aussitôt les plans à belle moyenne et queue pourrie.

::: figure utilite-ce
Sur une courbe d'utilité concave, la loterie 50/50 entre 20 et 65 k€ vaut moins que son espérance de 42,5 k€. Son équivalent certain tombe vers 36 k€. L'écart entre les deux est le prix du risque.
:::

## Tolérance et capacité : deux curseurs, pas un

Le vocabulaire courant confond deux choses que la décision doit séparer. La **tolérance** au risque est psychologique. C'est la baisse que vous supportez sans paniquer ni vendre ([[psychologie-du-retrait]]). La **capacité** de risque est objective. C'est la baisse que votre plan encaisse sans casser, quelle que soit votre sérénité. Un ancien trader flegmatique, avec un plan tendu à 4,5 % de retrait, a une tolérance haute et une capacité basse ; un anxieux assis sur 50 fois ses dépenses a l'inverse. La règle de composition est simple et sans exception. C'est le minimum des deux qui commande. La capacité se calcule, un simulateur fait exactement cela. La tolérance se découvre, hélas, surtout dans les vraies baisses. D'où la valeur des tests d'admissibilité par stratégie ([[choisir-sa-strategie]]) et des questions posées à froid, du type « votre retrait peut-il baisser de 20 % une année ? ».

::: science Kelly : le critère génial qu'il ne faut pas suivre
Le critère de Kelly (1956) répond élégamment à une vraie question. Quelle fraction de son capital risquer sur un pari favorable pour maximiser la croissance géométrique à long terme ([[rendements-arithmetiques-geometriques]]) ? Appliqué aux marchés, il recommande typiquement des expositions actions de 120 à 200 %. Il souffre de trois défauts rédhibitoires pour une retraite. D'abord, il suppose un horizon infini, alors que le rentier a 30-50 ans et un besoin de consommation daté. Ensuite, il est indifférent aux baisses maximales (drawdowns). Or le Kelly complet traverse sereinement des chutes de 50 % (probabilité 1/2 d'y passer un jour) et de 90 % (1/10), que le retrait transforme en ruine par risque de séquence ([[sequence-des-rendements]]). Enfin, il exige de connaître les vrais paramètres. Une petite erreur d'estimation en sur-Kelly détruit la croissance au lieu de la maximiser, car l'erreur est asymétrique. La leçon utile est le **Kelly fractionnaire**. Les praticiens sérieux misent un demi-Kelly ou moins, sacrifiant peu de croissance contre beaucoup de tranquillité. La position actions d'un plan de retrait bien construit se retrouve d'ailleurs, sans l'avoir cherché, quelque part vers le quart ou le tiers de Kelly. Si un produit ou un blog vous vend du Kelly complet pour votre retraite, il a lu la formule et pas les hypothèses.
:::

## Le regret, la satisfaction suffisante et la robustesse

L'utilité ne capture pas tout ce qui compte. Trois idées complètent la boîte à outils, toutes trois plus respectables que leur réputation.

**Le regret d'abord.** Harry Markowitz lui-même, inventeur de l'optimisation de portefeuille, racontait avoir placé sa propre épargne en 50/50 « pour minimiser mon regret futur », plutôt qu'à l'optimum de sa théorie. Le critère du minimax regret consiste à choisir l'option dont le pire regret rétrospectif est le plus petit. Il est parfaitement rationnel quand on ne vit qu'une trajectoire. Il explique pourquoi garder une part d'actions dans un plan prudent, par regret de rater trente ans de hausse, et une part d'obligations dans un plan agressif, par regret du krach de l'an 1. Ces choix dominent les solutions en coin. Beaucoup d'options « tièdes » de ce livre en relèvent, la dose d'or, le corridor de retrait, le 60/40 élargi. Ce sont des minimisations de regret assumées, et c'est un compliment.

**La satisfaction suffisante ensuite** (le satisficing de Herbert Simon). L'idée est de chercher une solution suffisante plutôt que la meilleure. Un plan qui atteint 95 % de réussite, avec un revenu qui vous suffit, n'a rien à gagner à optimiser davantage. Il a même beaucoup à y perdre, en complexité, en frais et en fragilité. La question de Simon (« assez bon pour quoi ? ») sert mieux le rentier que celle de l'optimiseur (« maximal selon quel modèle ? »). Cette dernière a toujours pour vraie réponse « selon un modèle faux » ([[monte-carlo-forces-faiblesses]]).

**La robustesse enfin.** Une décision robuste reste bonne même quand le modèle qui l'a produite se trompe. Les optima de portefeuille sont notoirement des crêtes fragiles. Changer le rendement attendu d'un point déplace l'allocation « optimale » de vingt. Mais la nature offre une grâce, car autour de l'optimum la surface est plate. Entre 50 et 80 % d'actions, le SWR bouge à peine ([[allocation-actions-obligations]]) ; entre 5 et 15 % d'or ou de trend, idem. La bonne pratique découle de ce relief. On identifie le plateau, on s'installe en son milieu, pas au bord où une erreur de modèle fait basculer, et on fuit toute recommandation qui n'existe que sur une crête. C'est l'optimisation robuste, dans sa version praticable sans mathématiques. On optimise contre la pire des hypothèses plausibles, pas pour la meilleure des estimations.

## Le protocole : cinq règles pour décider

Tout ce qui précède se condense en un protocole applicable à chaque décision du plan, de l'allocation au choix de la date ([[une-annee-de-plus]]).

Un : jugez les décisions sur le **processus**, jamais sur le résultat d'une trajectoire (une bonne décision peut mal tourner, un pari stupide peut payer ; sur une seule vie, confondre les deux est l'erreur la plus chère). Deux : comparez les plans sur le couple **percentile 5 + médiane** (votre équivalent certain artisanal), pas sur la moyenne ni sur le meilleur cas. Trois : préférez le **milieu des plateaux** aux sommets des crêtes, et méfiez-vous de toute option dont la supériorité disparaît quand on bouge une hypothèse d'un point. Quatre : à choix proche, minimisez le regret anticipé, dans les deux directions (le krach comme la hausse ratée). Cinq : écrivez la décision et ses conditions de révision **à froid** ([[construire-son-plan]]), car la meilleure théorie de la décision du monde ne survit pas à une décision prise en mars 2020 à 23 heures.

::: exemple Deux plans, deux critères, deux gagnants
Capital 1,2 M€, besoin 43 000 €/an. Plan A : 85 % actions, retrait fixe → médiane à 30 ans 4,1 M€, percentile 5 en ruine à l'année 24, moyenne superbe. Plan B : 65 % actions diversifiées, corridor flexible → médiane 2,9 M€, percentile 5 qui finit à 400 000 € avec un retrait descendu à 36 000 € au pire moment. Au critère « moyenne », A gagne largement. À l'équivalent certain d'une personne normalement prudente, B gagne sans discussion, car les 1,2 M€ de médiane supplémentaire de A pèsent moins que sa probabilité de 5 % de misère. Un simulateur départage d'autant mieux qu'on lui demande le bon critère. La question n'était pas « quel plan rapporte le plus ? » mais « quelle distribution préférez-vous habiter ? ».
:::

## L'essentiel à retenir

- Vous ne vivez qu'une trajectoire. L'espérance calculée sur dix mille mondes n'est donc pas votre critère ; l'utilité l'est, concave et brutalement asymétrique autour de la ruine. Elle justifie de juger tout plan sur le percentile 5 et la médiane.
- L'équivalent certain (« quel revenu garanti vaudrait ce plan risqué ? ») est le convertisseur universel entre plans ; il dégonfle mécaniquement les belles moyennes à queue laide.
- La décision obéit au minimum de deux curseurs : la tolérance (psychologique, découverte dans les baisses) et la capacité (objective, calculée par un simulateur).
- Kelly maximise la croissance géométrique sous des hypothèses (horizon infini, zéro consommation, paramètres connus) qui sont toutes fausses en retraite ; retenir l'idée du Kelly fractionnaire, fuir le Kelly complet.
- Regret minimisé, satisficing et robustesse (le milieu du plateau, pas le bord) sont des critères d'adulte, pas des renoncements ; le protocole tient en cinq règles, et la cinquième (décider à froid, par écrit) porte toutes les autres.

---

## Pour aller plus loin

- Daniel Bernoulli (1738), l'exposition du paradoxe de Saint-Pétersbourg ; Von Neumann & Morgenstern pour la théorie de l'utilité espérée.
- Herbert Simon, « A Behavioral Model of Rational Choice » (1955) : le satisficing par son inventeur, prix Nobel pour cela.
- Edward Thorp, « The Kelly Criterion in Blackjack, Sports Betting, and the Stock Market » : le meilleur exposé de Kelly, hypothèses incluses.
- Kahneman & Tversky, « Prospect Theory » (1979) : l'aversion aux pertes mesurée, le pont vers [[psychologie-du-retrait]].
- Dans ce livre : [[ruine-et-probabilites]] (choisir son seuil), [[choisir-sa-strategie]] (le protocole appliqué aux règles de retrait), [[allocation-actions-obligations]] (le plateau), [[monte-carlo-forces-faiblesses]] (pourquoi tout modèle est faux, et comment décider quand même).
