# Le retrait fixe indexé (Bengen) : le classique de référence

Le retrait fixe indexé sur l'inflation est la règle fondatrice du domaine ([[etude-trinity]]) et le point zéro de toute comparaison : quand pofo affiche « Safe spend », c'est elle qu'il résout ; quand une étude note une règle dynamique, c'est contre elle. Mais ce statut d'étalon fait oublier qu'elle est aussi une **stratégie** réelle, que des gens exécutent : la plus simple à vivre, la plus dure à défendre intellectuellement, et pourtant la bonne réponse pour certains profils précis.

Cet article la traite en stratégie opérationnelle et non plus en objet historique (l'histoire est dans [[etude-trinity]], la vulgarisation dans [[la-regle-des-4-pourcents]]) : la mécanique exacte et ses détails d'exécution qui changent les résultats (d'où sort l'argent, quand, comment on rééquilibre), son profil d'échec caractéristique, les variantes qui la réparent sans la dénaturer (indexation partielle, gel après les mauvaises années, cliquet à la hausse), les paramètres que la recherche recommande, pour qui elle reste le bon choix, et sa mise en œuvre dans pofo.

::: cle La règle, opérationnellement
Année 0 : fixez le retrait initial R = taux × capital (le taux venant de votre analyse, 3 à 4 % selon horizon, valorisations et marges, [[combien-il-vous-faut]]). Chaque année suivante : R ← R × (1 + inflation constatée), prélevé quoi qu'aient fait les marchés, en vendant ce qui dépasse l'allocation cible (le prélèvement **sert** au rééquilibrage). C'est tout. La règle n'écoute rien : ni le portefeuille, ni les valorisations, ni votre humeur : c'est sa faiblesse statistique et sa force psychologique.
:::

## La mécanique fine : trois détails qui pèsent

La règle tient en deux lignes, mais trois choix d'exécution, rarement explicités, déplacent réellement les résultats.

**D'où sort l'argent : le prélèvement-rééquilibrage.** La version naïve vend « un peu de tout » au prorata ; la bonne pratique vend **en priorité** la classe d'actifs en surpoids par rapport à l'allocation cible : après une bonne année actions on vend des actions, après un krach on vend des obligations (et on laisse les actions récupérer). Ce simple choix fait du retrait un rééquilibrage gratuit et améliore mesurablement la survie des pires millésimes : c'est une des raisons pour lesquelles les études historiques (qui supposent le rééquilibrage annuel) trouvent des SAFEMAX plus élevés que certains portefeuilles réels jamais rééquilibrés. pofo simule exactement cela : portefeuille agrégé aux poids cibles, prélèvement neutre ([[la-machine-pofo]]).

**Quand : la fréquence du prélèvement.** Annuel en début d'année (la convention des études), mensuel (la vraie vie, la case « Monthly withdrawals » de pofo) : l'écart est faible (le mensuel lisse un peu le point d'entrée des ventes, quelques dixièmes de point sur la ruine) mais le mensuel a une vertu de gouvernance : il transforme le retrait en « salaire » et évite la tentation de décaler la grosse vente annuelle en regardant le marché, qui est du market timing déguisé.

**Quelle inflation : l'indice.** La règle canonique indexe sur l'IPC national ; votre inflation personnelle peut en diverger durablement (santé, dépendance, la dérive réglable de pofo, [[suivre-inflation]], [[depenses-en-retraite]]). Indexer sur l'IPC en laissant la dérive santé non financée est l'angle mort le plus courant des plans « à la Bengen » de plus de 30 ans.

## Le profil d'échec : la falaise silencieuse

Ce qui distingue le fixe indexé de toutes les autres règles, c'est la **forme** de son échec, et il faut la voir pour comprendre à la fois son danger et les instruments de surveillance qu'il exige.

La règle n'échoue jamais d'un coup : elle échoue par un mécanisme de ciseaux silencieux. Dans un régime hostile ([[regimes-de-marche]]), le portefeuille baisse pendant que le retrait, lui, **monte** avec l'inflation : le taux de retrait courant (retrait / portefeuille) grimpe de 4 % à 6, 8, 12 %... Passé un seuil (empiriquement autour de 8-10 % de taux courant sans espoir de pension proche), même un marché redevenu généreux ne suffit plus : les prélèvements dépassent toute croissance plausible, et la trajectoire est condamnée **des années** avant le zéro ([[ruine-et-probabilites]]). Le millésime 1966 met quinze ans à devenir irrécupérable, et près de trente à s'épuiser ([[etude-trinity]]).

Deux conséquences pratiques de cette forme d'échec. D'abord, la règle fixe est la **seule** dont l'échec est parfaitement prévisible en cours de route : le taux de retrait courant est un voyant fiable, gradué, lisible par n'importe qui : quiconque exécute du Bengen doit suivre ce ratio une fois par an, avec des seuils écrits ([[quand-s-inquieter]]). Ensuite, personne, en pratique, ne va jusqu'à la falaise : face à un taux courant à 7 %, les humains coupent. Le fixe indexé réel est donc presque toujours un fixe-avec-flexibilité-implicite ; tout l'apport des règles à garde-fous ([[guyton-klinger]], [[guardrails-morningstar]]) est de rendre cette flexibilité **explicite**, décidée à froid, plutôt qu'improvisée dans la peur. C'est le sens profond de la comparaison de la frontière ([[panorama-strategies-retrait]]) : le Bengen pur est un point théorique ; le vrai choix est entre garde-fous écrits et garde-fous improvisés.

::: science Ce que coûte la stabilité, chiffré
Le fixe indexé achète la stabilité parfaite du revenu au prix le plus élevé de toutes les règles en capital immobilisé. Les ordres de grandeur de la littérature : pour la même ruine de 5 % à 45 ans, le fixe exige typiquement 10 à 20 % de capital de **plus** que des guardrails bien bornés, et 20 à 30 % de plus qu'un amortissement lissé ([[amortissement-abw]]) ; symétriquement, à capital égal, il sert la consommation totale moyenne la plus **faible** (il thésaurise dans les bons scénarios, son legs médian est énorme, souvent plusieurs fois la mise initiale, [[depenses-en-retraite]]). La stabilité est un produit de luxe ; le fixe indexé en est le prix affiché. Ce n'est pas un argument contre (certains ménages veulent exactement cela) ; c'est le prix qu'il faut connaître avant d'acheter.
:::

## Les variantes qui réparent sans dénaturer

Trois amendements historiques conservent l'esprit (un revenu cible stable) en limant les pathologies. Ils sont classés du plus doux au plus actif.

**L'indexation partielle, ou plafonnée.** Bengen lui-même l'a proposé : n'indexer qu'à hauteur de l'inflation moins un petit délta (inflation − 0,5 point), ou plafonner l'indexation annuelle (à 6 % par exemple). La logique : les grands désastres du fixe sont les régimes inflationnistes, précisément parce que l'indexation y devient un accélérateur ([[inflation-et-taux-de-retrait]]) ; freiner l'indexation dans ces épisodes est une mini-flexibilité quasi indolore (quelques pour cent de pouvoir d'achat étalés sur des années) qui remonte le SAFEMAX d'environ 0,25 à 0,5 point. La version comportementale, encore plus simple : **geler le montant nominal l'année qui suit une année de baisse du portefeuille** : pas d'indexation après un exercice négatif. Presque invisible dans la vie, mesurablement efficace dans les simulations : c'est la flexibilité minimale viable.

**Le cliquet à la hausse (ratcheting).** L'amendement symétrique, proposé par Kitces : puisque le fixe prudent finit riche dans la majorité des scénarios, autorisez des hausses **irréversibles** du retrait quand le plan est manifestement gagné : par exemple +10 % du retrait (en réel) chaque fois que le portefeuille dépasse 150 % de sa valeur initiale réelle, au plus tous les trois ans. Le cliquet ne dégrade presque pas la sécurité (on ne monte que du haut d'un coussin) et répare la pathologie du luxe non consommé. pofo l'implémente nativement (la case « Ratchet lifestyle up when rich », avec ses bornes, +10 % de la base, au plus tous les 2 ans, plafonné, et seulement quand le taux courant est redescendu très bas, [[utiliser-la-page-fire]]).

**Le taux initial conditionné aux valorisations.** Le troisième amendement ne touche pas la règle mais son **point de départ** : fixer le taux initial selon le CAPE du jour (3-3,25 % en marché cher, 4-4,5 % en marché purgé) plutôt qu'un 4 % universel ([[valorisations-et-cape]]). C'est la moitié du chemin vers les règles CAPE complètes ([[regles-cape]]), sans leur dynamique : une seule décision, prise le jour où l'on est le plus lucide.

Un fixe indexé équipé de ces trois amendements (indexation gelée après les années rouges, cliquet borné à la hausse, taux initial conditionné au CAPE) est une stratégie honorablement placée sur la frontière, à une fraction de la complexité des règles dynamiques. C'est la version que ce livre recommande à qui veut du Bengen.

## Pour qui c'est le bon choix

Le fixe indexé (amendé) reste la bonne réponse dans quatre situations identifiables :

- **Le plancher est presque tout le budget.** Si vos dépenses sont à 90 % incompressibles ([[combien-il-vous-faut]]), les règles flexibles n'ont presque rien à couper : leur avantage s'évanouit, autant prendre la stabilité et dimensionner le capital en conséquence.
- **La gouvernance prime.** Ménage où un seul gère, conjoint survivant peu à l'aise, patrimoine qui devra être exécuté par des tiers : la règle tient sur une carte postale et survit à son auteur ([[couple-et-famille]]).
- **Le retrait est déjà très bas.** Sous ~3 % de taux initial, toutes les règles convergent (la ruine est quasi nulle partout, [[horizon-et-esperance-de-vie]]) : la sophistication ne rapporte plus rien, la simplicité gagne par forfait.
- **Une phase à découvert courte.** Si la pension couvre le plancher dans moins de dix ans ([[revenus-complementaires]]), le fixe n'a qu'une courte fenêtre de vulnérabilité à traverser, et sa simplicité vaut le petit surcoût de capital.

À l'inverse, le fixe pur est le **mauvais** choix pour le profil FIRE typique : horizon de 45 ans, marché cher au départ, plancher nettement sous le confort : c'est exactement le profil où sa falaise est la plus probable et où la flexibilité rapporte le plus ([[flexibilite-realite]], [[choisir-sa-strategie]]).

::: exemple Le Bengen amendé, exécuté sur dix ans
Plan : 1,2 M€, retrait initial 3,4 % (CAPE haut) = 40 800 €, gel après année rouge, cliquet +10 % si portefeuille > 150 % réel. Années 1-3 : marchés moyens, indexation normale (42 900 € en année 3, inflation cumulée 5 %). Année 4 : −22 % : le retrait est **gelé** à 42 900 € nominal (l'inflation de 3 % n'est pas répercutée, −3 % de pouvoir d'achat, imperceptible). Années 5-9 : reprise ; l'indexation reprend ; le taux courant, monté à 4,9 % au creux, redescend à 3,1 %. Année 10 : le portefeuille atteint 158 % de la valeur initiale réelle : cliquet, +10 % : 51 300 € (réels). Bilan : dix ans, deux décisions non triviales (un gel, une hausse), zéro angoisse de falaise : le taux courant a été suivi chaque année et n'a jamais approché les seuils d'alerte. C'est à cela que ressemble la règle fondatrice bien exécutée.
:::

## Dans pofo

Le fixe indexé est le mode par défaut de la page FIRE : « Net spending /yr » avec toutes les règles de flexibilité à zéro. Trois lectures spécifiques à cette règle : la jauge « Safe spend » (le montant fixe qui atteint exactement votre ruine acceptable, la règle résolue à l'envers, [[utiliser-la-page-fire]]) ; la §03 (la décennie décisive, le fixe est la règle la **plus** exposée à la séquence, votre écart central/stress sera maximal ici, [[sequence-des-rendements]]) ; et le cliquet via sa case dédiée. Pour simuler l'indexation gelée, l'approximation pratique est une petite flexibilité (« Cut in downturns » à 5 %) : l'effet est du même ordre.

## L'essentiel à retenir

- Opérationnellement : R initial = taux × capital, puis indexation inflation, prélèvement qui rééquilibre (vendre le surpoids), mensuel de préférence, avec un œil sur l'écart IPC/inflation personnelle.
- Son échec est une falaise silencieuse à ciseaux (retrait qui monte, portefeuille qui baisse) : prévisible des années à l'avance par le taux de retrait courant, qui **doit** être suivi avec des seuils écrits.
- La stabilité parfaite coûte cher : 10-30 % de capital de plus que les règles à information pour la même sécurité, et un legs médian énorme (du luxe non consommé).
- Trois amendements quasi gratuits la réparent : gel de l'indexation après année rouge, cliquet borné à la hausse (natif dans pofo), taux initial conditionné au CAPE. Le « Bengen amendé » est une stratégie légitime de la frontière.
- Bon choix si : plancher ≈ budget, gouvernance prioritaire, taux déjà < 3 %, phase à découvert courte. Mauvais choix pour le FIRE type à 45 ans d'horizon en marché cher : lisez la suite de la partie.

---

## Pour aller plus loin

- Bengen, « Determining Withdrawal Rates Using Historical Data » (1994) et *Conserving Client Portfolios During Retirement* (2006) : la règle et ses propres amendements par son auteur.
- Kitces, « The Ratcheting Safe Withdrawal Rate » (2015) : le cliquet à la hausse.
- Early Retirement Now, volet 5 (les ajustements d'indexation) et volet 24 (la flexibilité minimale) ([[serie-ern]]).
- Dans ce livre : les deux articles amont ([[la-regle-des-4-pourcents]], [[etude-trinity]]), et les héritiers directs : [[guyton-klinger]] et [[plancher-plafond]].
