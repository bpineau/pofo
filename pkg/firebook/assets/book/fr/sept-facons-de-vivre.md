# Sept façons de vivre du même portefeuille

Une règle de retrait porte un nom, pas un visage. On sait que Bengen verse un montant fixe, que VPW verse un pourcentage, que les guardrails coupent quand le taux dérape. On sait beaucoup moins à quoi ressemble une vie menée sous chacune de ces règles. Combien on touche la première année, combien la dixième, si la baisse arrive d'un coup ou s'étire sur quinze ans, et ce qu'il reste sur le compte le jour où l'on s'arrête de compter.

C'est la seule chose que cette page cherche à montrer. Elle fige tout ce qui peut l'être, prend trois retraites qui ont réellement eu lieu, et imprime ce que chaque règle aurait versé, année après année. Aucune colonne ne gagne. Les taux d'échec sont ailleurs, dans les pages qui simulent des milliers d'avenirs ([[monte-carlo-forces-faiblesses]], [[ruine-et-probabilites]]) ; ici il n'y a aucun hasard, seulement trois séquences de rendements qui ont eu lieu et sept manières de les traverser.

## Le protocole

Un même ménage, sept fois. Il part avec **600 000 €**, il vise **2 000 € par mois**, soit 24 000 € par an, ce qui fait un taux de retrait initial de 4,0 %, exactement le chiffre de Bengen ([[la-regle-des-4-pourcents]]). Il ne touche aucune pension, il n'a aucun revenu d'appoint, il ne paie aucun impôt et il ne garde aucune poche de cash. Toutes ces simplifications flattent les sept colonnes de la même façon, donc elles ne changent pas la comparaison. Elles rendent seulement chaque chiffre un peu plus généreux que la vraie vie ([[flat-tax-et-imposition]], [[cash-buffer]]).

Le portefeuille est un 60/40 américain, 60 % de S&P 500 et 40 % de Treasuries à 5 ans, rééquilibré chaque janvier, déflaté par l'indice des prix américain. Ce n'est pas une recommandation. C'est l'étalon sur lequel toute la littérature du retrait a été écrite, de Bengen à l'étude Trinity jusqu'aux guardrails modernes ([[etude-trinity]]), donc le terrain le plus honnête pour comparer des règles qui y ont été calibrées.

::: cle Tout est en euros constants
Chaque montant de cette page, revenu comme capital, est exprimé en pouvoir d'achat d'aujourd'hui. L'inflation a déjà été retirée. Une ligne plate ne signifie donc pas un chèque identique chaque année, mais un niveau de vie identique chaque année : le chèque nominal, lui, a été indexé au passage. C'est la convention de tout le livre et elle est indispensable ici, parce que la décennie 1973-1982 aurait été impossible à lire autrement ([[inflation-et-taux-de-retrait]]).
:::

L'horizon de plan est de **40 ans**, toujours. Ce détail compte plus qu'il n'en a l'air. Deux des sept règles raisonnent sur l'horizon qui reste, l'amortissement en étalant le capital sur les années restantes et les guardrails par risque en surveillant le taux encore tenable à cet horizon. Leur dire que la retraite s'arrête commodément là où s'arrêtent les données les rendrait bien plus généreuses qu'elles ne le sont. Le plan court donc sur quarante ans dans tous les cas, et la page n'affiche que les années effectivement couvertes par l'historique disponible. Ces deux règles ont aussi besoin d'une hypothèse de rendement pour coter leur versement ; elles reçoivent une hypothèse générique de 4,5 % réel arithmétique à 10 % de volatilité, jamais les rendements de l'époque rejouée. Aucun retraité de 1973 ne savait ce qui l'attendait, et une règle nourrie au rétroviseur n'est plus la règle que décrit la littérature.

Les sept règles, de la plus rigide à la plus proportionnelle au capital :

| Règle | Ce qu'elle fait | Page dédiée |
|---|---|---|
| Retrait fixe | Le même montant réel chaque année | [[retrait-fixe-bengen]] |
| Flex −10 % | Le fixe, moins 10 % quand le portefeuille est 20 % sous son sommet | [[flexibilite-realite]] |
| Guardrails (GK) | ±10 % quand le taux courant sort d'un corridor autour du taux initial | [[guyton-klinger]] |
| Guardrails par risque | Même corridor, mais centré sur le taux encore sûr à l'horizon restant | [[guardrails-morningstar]] |
| % borné | Un pourcentage du capital, sans jamais bouger de plus de +5 % ou −2,5 % par an | [[plancher-plafond]] |
| Amortissement (ABW) | Le versement qui épuise exactement le capital sur l'horizon restant | [[amortissement-abw]] |
| % du portefeuille (VPW) | Une part fixe de ce qui reste, chaque année | [[vpw]] |

## Janvier 1973 : la crise d'abord

::: figure replay-marche-1973
Le portefeuille lui-même, sans un seul retrait. Le krach de 1973-1974 emporte 38 % en termes réels, puis l'inflation de la décennie ronge ce qui reste, si bien que le capital vaut encore un tiers de moins neuf ans après le départ. Il faut attendre 1983 pour repasser durablement au-dessus de la valeur initiale. Sur les quarante ans, le portefeuille finit pourtant à 4,7 % réel par an.
:::

Voilà la séquence pour laquelle toutes ces règles ont été inventées ([[sequence-des-rendements]]). Le rendement moyen des quarante ans est correct. C'est l'ordre qui tue. Un retrait pris pendant un krach vend une part bien plus grosse du portefeuille, puisque celui-ci a fondu alors que le montant retiré, lui, n'a pas bougé. Ces parts-là partent au plus bas et ne participent jamais à la reprise. Le même krach quinze ans plus tard trouverait un capital qui a déjà encaissé ses meilleures années, et qui a beaucoup moins d'années à financer derrière lui.

::: figure replay-revenus-1973
Sept vies dans la même crise. Le retrait fixe est la ligne droite : il verse 24 000 € pendant quarante ans sans jamais broncher. Les six autres coupent, et c'est la forme de la coupe qui les distingue. Le % borné descend par petites marches sur une quinzaine d'années et le ménage s'aperçoit à peine du glissement. Les deux guardrails plongent vite et fort, puis remontent très haut. VPW suit le marché sans amortisseur, année après année.
:::

| Règle | Revenu moyen | Écart | Pire année | Années maigres | Capital final |
|---|---|---|---|---|---|
| Retrait fixe | 24,0 | 0 % | 24,0 | 0 | 19 |
| Flex −10 % | 22,0 | 4 % | 21,6 | 33 | 344 |
| Guardrails (GK) | 23,5 | 33 % | 14,2 | 23 | 781 |
| Guardrails par risque | 25,6 | 39 % | 12,8 | 19 | 762 |
| % borné | 21,3 | 16 % | 16,0 | 28 | 575 |
| Amortissement (ABW) | 30,9 | 35 % | 15,4 | 13 | 0 |
| % du portefeuille (VPW) | 24,5 | 34 % | 12,5 | 19 | 734 |

Revenus en k€ par an, capital final en k€, le tout en euros constants. L'écart est le coefficient de variation du revenu annuel, autrement dit sa dispersion rapportée à sa moyenne, et c'est la mesure de ce que la règle fait bouger dans la vie du ménage. Les années maigres comptent les années vécues sous les 24 000 € prévus, sur les quarante.

Le retrait fixe a tenu. Il a versé son dû quarante années de suite et il termine avec 19 000 € sur le compte, soit trois pour cent de ce qu'il avait au départ. Une année de marché de plus dans le mauvais sens et la colonne racontait une ruine. C'est la vraie nature de cette règle, et le mot juste est falaise plutôt que risque. Tout va bien jusqu'à la seconde où plus rien ne va.

Les deux guardrails sont l'inverse exact. Ils ont sauvé le capital, plus de 750 000 € à l'arrivée, mais ils ont fait vivre le ménage à 14 200 € puis 12 800 € pendant l'essentiel des années 1970 et 1980. Ce sont des coupes de 40 à 47 %, tenues sur plus d'une décennie. Le taux de succès de ces colonnes est parfait et le train de vie qu'elles décrivent ne l'est pas du tout ([[guyton-klinger]]).

Entre les deux, deux réponses honnêtes et très différentes. Le % borné descend jusqu'à 16 000 €, mais il met seize ans pour y aller, par marches de 2,5 % par an. Personne ne vit une baisse de 2,5 % comme une crise ; on la vit comme une année un peu serrée, et c'est exactement l'intérêt de la règle ([[plancher-plafond]]). L'amortissement, lui, verse le plus gros revenu moyen des sept, 30 900 € par an, et il termine à zéro. Ce n'est pas un échec, c'est sa définition : il a coté chaque année le versement qui épuise le capital sur l'horizon restant, et l'horizon s'est terminé ([[amortissement-abw]]).

## Janvier 1985 : le problème inverse

::: figure replay-marche-1985
Quarante ans de vent arrière. Le portefeuille encaisse 1987, 2000, 2008 et 2022 depuis une position confortable et finit multiplié par douze en termes réels, à 6,5 % par an. La première décennie fait 8,4 % par an, et cette décennie-là décide de tout ce qui suit.
:::

Personne ne manque d'argent dans cette retraite. La question intéressante s'inverse donc, et elle est au moins aussi importante que la précédente. Qui a dépensé ce que le portefeuille pouvait manifestement payer, et qui a passé quarante ans à vivre petitement sur une fortune qu'il n'a jamais touchée ?

::: figure replay-revenus-1985
Sept vies dans le même vent arrière. Le retrait fixe et sa variante flexible restent collés à leur ligne de 24 000 €, pendant que les cinq autres montent. L'amortissement va chercher 82 000 € au sommet de 2022. Guardrails, % borné et VPW se tiennent autour de 47 000 € en moyenne, soit le double du plan initial.
:::

| Règle | Revenu moyen | Écart | Pire année | Années maigres | Capital final |
|---|---|---|---|---|---|
| Retrait fixe | 24,0 | 0 % | 24,0 | 0 | 3 829 |
| Flex −10 % | 23,9 | 2 % | 21,6 | 2 | 3 843 |
| Guardrails (GK) | 46,6 | 26 % | 24,0 | 0 | 1 656 |
| Guardrails par risque | 32,9 | 16 % | 21,6 | 7 | 3 053 |
| % borné | 47,2 | 25 % | 24,0 | 0 | 1 571 |
| Amortissement (ABW) | 58,3 | 23 % | 29,1 | 0 | 0 |
| % du portefeuille (VPW) | 46,4 | 22 % | 24,0 | 0 | 1 448 |

Le retrait fixe meurt avec 3,83 millions d'euros constants sur le compte, après avoir vécu quarante ans à 2 000 € par mois. Il a multiplié son capital par six et n'en a rien fait. Personne n'appelle cela un échec parce que le mot ruine ne s'applique pas, et pourtant le ménage a renoncé à la moitié de sa retraite. L'amortissement a vécu la même période à 58 300 € par an en moyenne, deux fois et demie mieux, et il finit à zéro parce que c'était le contrat.

On tient ici le prix caché de la rigidité, et il ne se lit sur aucun taux de succès. Une règle qui ne monte jamais est une règle qui ne saura pas dépenser un bon marché. Comme la plupart des retraites tombent sur des marchés plutôt bons ([[decider-sous-incertitude]]), c'est le cas le plus fréquent, pas le cas exotique.

Deux nuances utiles. Les guardrails par risque restent nettement plus bas que leurs cousins, à 32 900 €, parce que leur corridor est plafonné à 150 % du niveau prévu dans cette page ; sans ce plafond la règle grimpe indéfiniment à mesure que l'horizon raccourcit, ce qui est une pathologie connue et non une performance ([[guardrails-morningstar]]). Et le flex à −10 % ne coupe que deux fois en quarante ans, ce qui rappelle qu'une flexibilité conditionnée à une baisse de 20 % ne sert presque jamais dans un marché haussier. C'est une assurance, pas une stratégie de revenu.

## Janvier 2000 : celle qu'on vit

::: figure replay-marche-2000
La décennie perdue, puis le rattrapage. Le départ se fait sur le sommet de la bulle internet, la première décennie fait exactement zéro pour cent réel par an, 2008 arrive sur un portefeuille qui n'avait jamais récupéré 2000, et 2022 tombe sur un coussin plus mince que prévu. Il aura fallu treize ans pour retrouver la valeur réelle du départ.
:::

Cette retraite n'est pas finie. Vingt-six des quarante années prévues sont écoulées, et c'est précisément ce qui la rend instructive : elle se lit comme la vôtre pourrait se lire aujourd'hui, à mi-parcours, sans savoir la fin.

::: figure replay-revenus-2000
Sept vies dans la décennie perdue. Six règles sur sept ont déjà fait vivre le ménage en dessous de son plan, la plupart pendant plus de vingt ans. Seul le retrait fixe a tenu sa ligne, et il lui reste 271 000 € pour financer quatorze années à 24 000 €.
:::

| Règle | Revenu moyen | Écart | Pire année | Années maigres | Capital final |
|---|---|---|---|---|---|
| Retrait fixe | 24,0 | 0 % | 24,0 | 0 | 271 |
| Flex −10 % | 21,9 | 4 % | 21,6 | 23 | 387 |
| Guardrails (GK) | 18,9 | 10 % | 17,5 | 24 | 552 |
| Guardrails par risque | 19,7 | 24 % | 15,7 | 21 | 589 |
| % borné | 19,5 | 11 % | 16,9 | 25 | 511 |
| Amortissement (ABW) | 23,7 | 12 % | 18,2 | 16 | 316 |
| % du portefeuille (VPW) | 19,1 | 12 % | 14,7 | 25 | 568 |

Les années maigres se comptent ici sur vingt-six, pas sur quarante.

Regardez la ligne du retrait fixe et faites le calcul qu'un ménage ferait à sa place. Il lui reste 271 000 € et quatorze ans de plan, soit un taux de retrait courant de 8,9 %. Aucune de ces quatorze années n'est financée par autre chose que de la chance. Les colonnes qui ont coupé, elles, ont environ le double de capital pour la même distance restante, et elles l'ont payé en vivant à 19 000 € pendant vingt ans. Personne ne peut dire aujourd'hui qui avait raison, et c'est le fond du sujet.

::: attention Ce que la dernière colonne ne dit pas
Un capital final élevé n'est ni une victoire ni une défaite tant qu'on ne sait pas ce qu'on voulait en faire. Pour un ménage sans héritier et sans projet de transmission, les 3,83 millions du retrait fixe en 1985 sont une perte sèche de niveau de vie. Pour un ménage qui veut transmettre, ou qui redoute une dépendance coûteuse en fin de vie, c'est exactement le but recherché ([[succession-et-transmission]], [[sante-et-protection-sociale]]). La bonne règle dépend d'une question qui n'est pas financière.
:::

## Trois retraites ne sont pas une distribution

Trois dates de départ, ce sont trois tirages. Une règle qui a traversé 1973 et 2000 n'est pas prouvée sûre ; elle est simplement plausible sur trois parcours d'un seul pays, celui dont le siècle a été le plus favorable de tous ([[anarkulova-cederburg]]). C'est exactement pour cela que les simulateurs tirent des milliers d'avenirs au lieu de trois ([[historique-vs-parametrique]]).

Ces trois-là ont d'ailleurs été choisies, ce qui est une forme de biais qu'il vaut mieux annoncer. L'historique américain disponible contient trente-trois fenêtres de quarante ans, et la fenêtre médiane rend 5,5 % réel par an. 1973 en rend 4,7 % et 1985 en rend 6,5 %, donc les deux périodes retenues encadrent ce cas médian au lieu de le flatter. Elles n'épuisent pas pour autant l'éventail des cas possibles. Il existe des séquences pires que 1973 ailleurs dans le monde, et le Japon d'après 1990 en est l'exemple le plus brutal ([[hyperinflation-et-extremes]]).

Lisez donc cette page pour le comportement des règles, pas pour leur classement. Ce comportement, lui, est stable et se retrouve d'une époque à l'autre : le fixe ne bouge jamais et casse d'un coup ou pas du tout ; les guardrails plongent tôt, fort et longtemps ; le % borné glisse au lieu de tomber ; l'amortissement suit le marché en montant avec l'âge et finit à zéro par construction ; VPW encaisse tout, sans amortisseur et sans jamais s'épuiser.

## Ce qu'on en retire

Trois choses tiennent après ces trois retraites, et elles ne se lisent sur aucune probabilité de ruine.

La première, c'est que la stabilité du revenu et la sécurité du capital sont la même ressource, dépensée à deux endroits ([[panorama-strategies-retrait]]). Le retrait fixe a acheté quarante ans de tranquillité en 1973 et l'a payée avec les 19 000 € qui lui restaient à l'arrivée. Les guardrails ont acheté 780 000 € de capital et les ont payés avec une décennie de vaches maigres. Aucune règle ne crée du confort, elles le déplacent.

La deuxième, c'est que la forme d'une coupe compte autant que sa profondeur. Le % borné et les guardrails descendent tous deux vers 15 à 16 000 € en 1973. L'un met seize ans, par marches insensibles ; l'autre y arrive en sept ans, par sauts de 10 %. Le second est infiniment plus dur à vivre, et il est très mal payé par les indicateurs habituels ([[psychologie-du-retrait]]).

La troisième, c'est qu'une règle qui ne monte jamais est un choix, pas une prudence. Choisir de ne jamais augmenter son train de vie revient à parier sur 1973 dans un monde où la plupart des retraites ressemblent à 1985. Si ce pari est le vôtre, prenez-le en connaissance de cause. Si le vrai objectif est de ne pas manquer, il existe des façons moins coûteuses d'y arriver, à commencer par un plancher explicite sur la baisse ([[choisir-sa-strategie]]).
