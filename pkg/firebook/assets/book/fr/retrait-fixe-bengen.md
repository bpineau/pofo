# Le retrait fixe indexé (Bengen) : le classique de référence

Le retrait fixe indexé sur l'inflation est la règle fondatrice du domaine ([[etude-trinity]]). C'est le point zéro de toute comparaison. Quand un simulateur affiche « Safe spend », c'est elle qu'il résout. Quand une étude évalue une règle dynamique, c'est contre elle qu'elle la mesure. Ce statut d'étalon fait pourtant oublier une chose : c'est aussi une stratégie réelle, que des gens appliquent. La plus simple à vivre, la plus difficile à défendre sur le papier, et pourtant la bonne réponse pour certains profils précis.

Cet article la traite en stratégie opérationnelle, et non plus en objet historique (l'histoire est dans [[etude-trinity]], la vulgarisation dans [[la-regle-des-4-pourcents]]). On y verra sa mécanique exacte et les détails d'exécution qui changent les résultats : d'où sort l'argent, quand, comment on rééquilibre. Puis son profil d'échec caractéristique, et les variantes qui la réparent sans la dénaturer (indexation partielle, gel après les mauvaises années, cliquet à la hausse). Enfin, les paramètres que la recherche recommande, les profils pour qui elle reste le bon choix, et sa mise en œuvre sur la page FIRE.

::: cle La règle, opérationnellement
Année 0 : fixez le retrait initial R = taux × capital. Le taux vient de votre analyse, 3 à 4 % selon l'horizon, les valorisations et les marges ([[combien-il-vous-faut]]). Chaque année suivante : R ← R × (1 + inflation constatée). On prélève quoi qu'aient fait les marchés, en vendant ce qui dépasse l'allocation cible, si bien que le prélèvement sert au rééquilibrage. C'est tout. La règle n'écoute rien : ni le portefeuille, ni les valorisations, ni votre humeur. C'est sa faiblesse statistique et sa force psychologique.
:::

## La mécanique fine : trois détails qui pèsent

La règle tient en deux lignes. Mais trois choix d'exécution, rarement explicités, déplacent réellement les résultats.

**D'où sort l'argent : le prélèvement-rééquilibrage.** La version naïve vend « un peu de tout » au prorata. La bonne pratique vend en priorité la classe d'actifs en surpoids par rapport à l'allocation cible. Après une bonne année actions, on vend des actions ; après un krach, on vend des obligations et on laisse les actions récupérer. Ce simple choix fait du retrait un rééquilibrage gratuit et améliore nettement la survie des pires millésimes. C'est une des raisons pour lesquelles les études historiques, qui supposent un rééquilibrage annuel, trouvent des SAFEMAX plus élevés que ceux de portefeuilles réels jamais rééquilibrés. Un simulateur reproduit exactement cela : portefeuille agrégé aux poids cibles, prélèvement neutre ([[la-machine-pofo]]).

**Quand : la fréquence du prélèvement.** Deux conventions coexistent : annuel en début d'année dans les études, mensuel dans la vraie vie (la case « Monthly withdrawals »). L'écart de résultat est faible, quelques dixièmes de point sur la ruine, car le mensuel lisse un peu le point d'entrée des ventes. Mais il a une vertu de gouvernance. Il transforme le retrait en salaire et retire la tentation de décaler la grosse vente annuelle en regardant le marché, ce qui n'est que du market timing déguisé.

**Quelle inflation : l'indice.** La règle canonique indexe sur l'IPC national. Votre inflation personnelle peut en diverger durablement, à cause de la santé, de la dépendance ou de la dérive réglable ([[suivre-inflation]], [[depenses-en-retraite]]). Indexer sur l'IPC en laissant la dérive santé non financée est l'angle mort le plus courant des plans « à la Bengen » de plus de 30 ans.

## Le profil d'échec : la falaise silencieuse

Ce qui distingue le fixe indexé de toutes les autres règles, c'est la forme de son échec. Il faut la voir pour comprendre à la fois son danger et les instruments de surveillance qu'il exige.

La règle n'échoue jamais d'un coup. Elle échoue par un mécanisme de ciseaux silencieux. Dans un régime hostile ([[regimes-de-marche]]), le portefeuille baisse pendant que le retrait, lui, monte avec l'inflation. Le taux de retrait courant (retrait / portefeuille) grimpe alors de 4 % à 6, 8, 12 %. Passé un seuil, situé empiriquement autour de 8 à 10 % de taux courant sans pension proche, même un marché redevenu généreux ne suffit plus. Les prélèvements dépassent toute croissance plausible, et la trajectoire est condamnée des années avant le zéro ([[ruine-et-probabilites]]). Le millésime 1966 met quinze ans à devenir irrécupérable, et près de trente à s'épuiser ([[etude-trinity]]).

Cette forme d'échec a deux conséquences pratiques. D'abord, la règle fixe est la seule dont l'échec est parfaitement prévisible en cours de route. Le taux de retrait courant est un voyant fiable, gradué, lisible par n'importe qui. Quiconque applique du Bengen doit suivre ce ratio une fois par an, avec des seuils écrits ([[quand-s-inquieter]]). Ensuite, personne, en pratique, ne va jusqu'à la falaise : face à un taux courant à 7 %, les humains coupent. Le fixe indexé réel est donc presque toujours un fixe assorti d'une flexibilité implicite. Tout l'apport des règles à garde-fous ([[guyton-klinger]], [[guardrails-morningstar]]) est de rendre cette flexibilité explicite, décidée à froid plutôt qu'improvisée dans la peur. C'est le sens profond de la comparaison sur la frontière ([[panorama-strategies-retrait]]) : le Bengen pur est un point théorique, tandis que le vrai choix se joue entre garde-fous écrits et garde-fous improvisés.

::: science Ce que coûte la stabilité, chiffré
Le fixe indexé achète la stabilité parfaite du revenu au prix fort : c'est la règle qui immobilise le plus de capital. Voici les ordres de grandeur de la littérature. Pour une même ruine de 5 % à 45 ans, le fixe exige typiquement 10 à 20 % de capital de plus que des garde-fous bien bornés, et 20 à 30 % de plus qu'un amortissement lissé ([[amortissement-abw]]). Symétriquement, à capital égal, il sert la consommation totale moyenne la plus faible : il thésaurise dans les bons scénarios, et son legs médian est énorme, souvent plusieurs fois la mise initiale ([[depenses-en-retraite]]). La stabilité est un produit de luxe, et le fixe indexé en est le prix affiché. Ce n'est pas un argument contre lui, car certains ménages veulent exactement cela. C'est le prix qu'il faut connaître avant d'acheter.
:::

## Les variantes qui réparent sans dénaturer

Trois amendements historiques conservent l'esprit (un revenu cible stable) en limant les pathologies. Ils sont classés du plus doux au plus actif.

**L'indexation partielle, ou plafonnée.** Bengen lui-même l'a proposée : n'indexer qu'à hauteur de l'inflation diminuée d'un petit écart (inflation − 0,5 point), ou plafonner l'indexation annuelle (à 6 % par exemple). Le raisonnement est simple. Les grands désastres du fixe sont les régimes inflationnistes, justement parce que l'indexation y devient un accélérateur ([[inflation-et-taux-de-retrait]]). Freiner l'indexation dans ces épisodes est une petite flexibilité quasi indolore, quelques pour cent de pouvoir d'achat étalés sur des années, qui remonte le SAFEMAX d'environ 0,25 à 0,5 point. Une version comportementale est encore plus simple : geler le montant nominal l'année qui suit une baisse du portefeuille, sans indexation après un exercice négatif. Presque invisible dans la vie, elle est nettement efficace dans les simulations. C'est la flexibilité minimale viable.

**Le cliquet à la hausse (ratcheting).** C'est l'amendement symétrique, proposé par Kitces. Puisque le fixe prudent finit riche dans la majorité des scénarios, on s'autorise des hausses irréversibles du retrait quand le plan est manifestement gagné. Par exemple : +10 % du retrait (en réel) chaque fois que le portefeuille dépasse 150 % de sa valeur initiale réelle, au plus tous les trois ans. Le cliquet ne dégrade presque pas la sécurité, puisqu'on ne monte que du haut d'un coussin, et il répare la pathologie du luxe non consommé. La page FIRE l'implémente nativement, via la case « Ratchet lifestyle up when rich » : +10 % de la base, plafonné, au plus tous les 2 ans, et seulement quand le taux courant est redescendu très bas ([[utiliser-la-page-fire]]).

**Le taux initial conditionné aux valorisations.** Le troisième amendement ne touche pas la règle, mais son point de départ : fixer le taux initial selon le CAPE du jour (3 à 3,25 % en marché cher, 4 à 4,5 % en marché purgé) plutôt qu'un 4 % universel ([[valorisations-et-cape]]). C'est la moitié du chemin vers les règles CAPE complètes ([[regles-cape]]), sans leur dynamique. Il s'agit d'une seule décision, prise le jour où l'on est le plus lucide.

Un fixe indexé équipé de ces trois amendements (indexation gelée après les années rouges, cliquet borné à la hausse, taux initial conditionné au CAPE) est une stratégie honorablement placée sur la frontière, à une fraction de la complexité des règles dynamiques. C'est la version que ce livre recommande à qui veut du Bengen.

## Pour qui c'est le bon choix

Le fixe indexé (amendé) reste la bonne réponse dans quatre situations identifiables :

- **Le plancher est presque tout le budget.** Si vos dépenses sont à 90 % incompressibles ([[combien-il-vous-faut]]), les règles flexibles n'ont presque rien à couper : leur avantage s'évanouit, autant prendre la stabilité et dimensionner le capital en conséquence.
- **La gouvernance prime.** Ménage où une seule personne gère, conjoint survivant peu à l'aise, patrimoine qui devra être exécuté par des tiers : la règle tient sur une carte postale et survit à son auteur ([[couple-et-famille]]).
- **Le retrait est déjà très bas.** Sous ~3 % de taux initial, toutes les règles convergent (la ruine est quasi nulle partout, [[horizon-et-esperance-de-vie]]) : la sophistication ne rapporte plus rien, la simplicité gagne par forfait.
- **Une phase à découvert courte.** Si la pension couvre le plancher dans moins de dix ans ([[revenus-complementaires]]), le fixe n'a qu'une courte fenêtre de vulnérabilité à traverser, et sa simplicité vaut le petit surcoût de capital.

À l'inverse, le fixe pur est le mauvais choix pour le profil FIRE typique : horizon de 45 ans, marché cher au départ, plancher nettement sous le confort. C'est exactement le profil où sa falaise est la plus probable et où la flexibilité rapporte le plus ([[flexibilite-realite]], [[choisir-sa-strategie]]).

::: exemple Le Bengen amendé, exécuté sur dix ans
Plan : 1,2 M€, retrait initial 3,4 % (CAPE haut) = 40 800 €, gel après année rouge, cliquet +10 % si le portefeuille dépasse 150 % réel. Années 1-3 : marchés moyens, indexation normale (42 900 € en année 3, inflation cumulée 5 %). Année 4 → −22 %. Le retrait est gelé à 42 900 € nominal ; l'inflation de 3 % n'est pas répercutée, soit −3 % de pouvoir d'achat, imperceptible. Années 5-9 : reprise, l'indexation redémarre, et le taux courant, monté à 4,9 % au creux, redescend à 3,1 %. Année 10 : le portefeuille atteint 158 % de la valeur initiale réelle, d'où le cliquet, +10 %, soit 51 300 € réels. Bilan : dix ans, deux décisions non triviales (un gel, une hausse), zéro angoisse de falaise. Le taux courant a été suivi chaque année et n'a jamais approché les seuils d'alerte. C'est à cela que ressemble la règle fondatrice bien exécutée.
:::

## Dans pofo

Le fixe indexé est le mode par défaut de la page FIRE : « Net spending /yr » avec toutes les règles de flexibilité à zéro. Trois lectures lui sont propres. La jauge « Safe spend » donne le montant fixe qui atteint exactement votre ruine acceptable, c'est-à-dire la règle résolue à l'envers ([[utiliser-la-page-fire]]). La §03 montre la décennie décisive : le fixe est la règle la plus exposée à la séquence, et votre écart central/stress y sera maximal ([[sequence-des-rendements]]). Le cliquet, enfin, passe par sa case dédiée. Pour simuler l'indexation gelée, l'approximation pratique est une petite flexibilité (« Cut in downturns » à 5 %), dont l'effet est du même ordre.

## L'essentiel à retenir

- Opérationnellement : R initial = taux × capital, puis indexation inflation, prélèvement qui rééquilibre (vendre le surpoids), mensuel de préférence, avec un œil sur l'écart IPC/inflation personnelle.
- Son échec est une falaise silencieuse à ciseaux (retrait qui monte, portefeuille qui baisse) : prévisible des années à l'avance par le taux de retrait courant, qui doit être suivi avec des seuils écrits.
- La stabilité parfaite coûte cher : 10-30 % de capital de plus que les règles à information pour la même sécurité, et un legs médian énorme (du luxe non consommé).
- Trois amendements quasi gratuits la réparent : gel de l'indexation après année rouge, cliquet borné à la hausse (natif sur la page FIRE), taux initial conditionné au CAPE. Le « Bengen amendé » est une stratégie légitime de la frontière.
- Bon choix si : plancher ≈ budget, gouvernance prioritaire, taux déjà < 3 %, phase à découvert courte. Mauvais choix pour le FIRE type à 45 ans d'horizon en marché cher : lisez la suite de la partie.

---

## Pour aller plus loin

- Bengen, « Determining Withdrawal Rates Using Historical Data » (1994) et *Conserving Client Portfolios During Retirement* (2006) : la règle et ses propres amendements par son auteur.
- Kitces, « The Ratcheting Safe Withdrawal Rate » (2015) : le cliquet à la hausse.
- Early Retirement Now, volet 5 (les ajustements d'indexation) et volet 24 (la flexibilité minimale) ([[serie-ern]]).
- Dans ce livre : les deux articles amont ([[la-regle-des-4-pourcents]], [[etude-trinity]]), et les héritiers directs [[guyton-klinger]] et [[plancher-plafond]].
