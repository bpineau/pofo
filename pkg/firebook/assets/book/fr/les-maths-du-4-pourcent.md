# Pourquoi 4 % ? L'anatomie mathématique de la règle

Le chapitre [[la-regle-des-4-pourcents]] raconte d'où vient la règle, et [[etude-trinity]] comment elle a été testée. Mais pourquoi 4 %, si les actions américaines ont rapporté davantage ? Imaginons d'abord un portefeuille qui rapporte exactement 4 % après inflation, chaque année. Combien pouvez-vous en retirer si vous voulez le conserver intact ? Et si vous avez trente ans pour le dépenser ?

Les deux réponses se calculent. Pour la troisième étape, où les rendements varient d'une année à l'autre, il faut ouvrir les tables historiques.

::: cle L'idée en une phrase
À rendement réel **constant** de 4 %, on peut retirer 4 % du capital initial chaque fin d'année indéfiniment, ou environ 5,8 % par an pendant trente ans. Dans l'étude américaine de Bengen, un retrait initial voisin de 4 %, ensuite indexé sur l'inflation, a résisté aux mauvais départs sur trente ans. La différence avec 5,8 % vient à la fois des rendements traversés et de leur ordre.
:::

## Étage 1 : la rente perpétuelle, ou le rendement réel comme plafond naturel

Commençons par un portefeuille de 100 000 € qui rapporte exactement 4 % par an après inflation. À la fin de la première année, il contient 104 000 € en pouvoir d'achat de départ. Vous retirez 4 000 €. Il reste 100 000 €, et vous pouvez recommencer l'année suivante ([[inflation-et-taux-de-retrait]]).

Pourquoi avoir choisi 4 % dans cet exemple ? C'est un ordre de grandeur du rendement réel géométrique d'un portefeuille historique composé de 60 % d'actions et 40 % d'obligations. La moyenne **géométrique** donne le taux constant qui aurait produit le même capital final, dividendes réinvestis et sans retrait. La moyenne **arithmétique** des rendements annuels est généralement plus élevée ([[rendements-arithmetiques-geometriques]]). Dès que vous retirez de l'argent en cours de route, deux portefeuilles de même rendement géométrique peuvent finir avec des soldes très différents.

## Étage 2 : dépenser le capital sur trente ans

Avec trente ans devant vous, vous pouvez puiser dans le capital. À 4 % réels constants, un portefeuille de 100 000 € finance trente retraits de fin d'année d'environ 5 783 € en euros constants. Après le dernier, le solde est nul. Cette faculté de consommer le capital ajoute environ 1,8 point au retrait initial. C'est le **bonus d'amortissement** que les méthodes ABW et VPW recalculent en fonction de la durée restante ([[amortissement-abw]], [[vpw]]).

Ce bonus explique au passage la sensibilité à l'horizon ([[horizon-et-esperance-de-vie]]). À 50 ans d'horizon, le même calcul donne 4,7 % et le bonus fond à +0,7. À l'infini, il disparaît. La retraite très précoce ne « perd » pas son taux à cause d'un mystère. Elle perd simplement le droit d'amortir vite.

::: figure amortissement-horizon
Le retrait qu'une pure arithmétique d'amortissement finance à 4 % réel, selon le nombre d'années à couvrir. La courbe tombe vite, puis s'aplatit sur son asymptote, le rendement seul. Les vingt premières années emportent l'essentiel du bonus, tandis que passer de 50 ans d'horizon à l'éternité ne coûte plus que sept dixièmes de point. La retraite très précoce paie donc son horizon, mais beaucoup moins qu'on ne le croit.
:::

## Étage 3 : quand les mauvaises années arrivent en premier

Vous avez 100 € et en retirez 10 à la fin de chaque année. Si le marché perd 20 % puis gagne 25 %, il vous reste 70 € après le premier retrait et 77,50 € après le second. Inversez les rendements : il vous reste 115 € puis 82 €. Sans les retraits, les deux ordres auraient ramené les 100 € de départ. Avec les retraits, la perte précoce réduit la somme qui profite du rebond ([[sequence-des-rendements]]).

Bengen a comparé des départs à des dates différentes aux États-Unis. Les retraités n'ont donc ni les mêmes rendements sur trente ans, ni les mêmes années de forte inflation au début de leur retraite. Les quelque 1,8 point entre notre rente théorique et le retrait historiquement soutenable illustrent cet écart. Pour isoler le coût de l'ordre seul, il faudrait garder exactement les mêmes rendements réels et les permuter, comme dans l'exemple des 100 €.

::: figure cascade-4pct
Les deux premières marches se calculent à rendement constant. La dernière relie ce calcul au retrait observé sur l'histoire américaine. La barre « séquence » inclut aussi le niveau des rendements propres aux périodes difficiles.
:::

::: exemple La cascade complète, sur un coin de table
Dans cet exemple, on passe de 4 % de rendement réel à 5,8 % de retrait sur trente ans à rendement constant, puis à environ 4 % dans le test historique. Le graphique suivant montre des ordres de grandeur quand une hypothèse change à la fois. Horizon de 50 ans → le bonus fond, la pénalité s'allège un peu, ~3,4 % dans l'exemple. Échantillon mondial ([[anarkulova-cederburg]]) → −0,5 à −1 point. Valorisation élevée au départ ([[valorisations-et-cape]]) → étage 1 raboté d'un point dans ce scénario. 0,5 % de frais → −0,5, presque un pour un dans l'illustration. Règle flexible à plancher → +0,3 à +0,5, en contrepartie de possibles baisses de dépenses. Si votre plan change sur plusieurs points, rejouez-le avec toutes ses hypothèses.
:::

::: figure clavier-leviers
Chaque hypothèse déplacée seule, à partir du plan de référence de la cascade : 60/40 historique, 30 ans, retrait fixe indexé, soit 4,0 %. Les fourchettes sont celles de l'article, et le signe de chaque barre dit dans quel sens le levier joue.
:::

## Pourquoi c'est si robuste (et ce qui le casserait)

Un taux choisi pour traverser le départ américain le plus difficile laisse souvent un capital important aux ménages partis à une date plus favorable ([[lire-un-fan-chart]]). Pourquoi garder cette marge ? Au moment de partir, vous ignorez à quel groupe vous appartiendrez. Le capital laissé par un autre ménage ne paiera pas vos dépenses si votre propre retraite commence mal.

Trois vérifications rapprochent ce résultat de votre situation. Utilisez des rendements qui correspondent aux marchés où vous investissez ([[diversification-internationale]]). Déduisez les frais annuels du portefeuille et appliquez la fiscalité aux flux qu'elle concerne ([[etf-ucits-europeens]], [[flat-tax-et-imposition]]). Enfin, vérifiez combien de temps vous devrez financer le retrait et quelles dépenses pourraient réellement baisser si les marchés déçoivent ([[flexibilite-realite]]).

::: science Ce que les reconstructions modernes confirment
Les travaux cités en fin de chapitre explorent des questions différentes. Early Retirement Now compare les retraits selon la valorisation de départ et l'allocation ([[serie-ern]]). Blanchett et ses coauteurs étudient l'ajustement des dépenses sur la durée de la retraite. Anarkulova et Cederburg élargissent l'échantillon de marchés. Pour rapprocher deux taux publiés, relevez d'abord la durée, les marchés retenus et la définition de l'échec.
:::

## L'essentiel à retenir

- À rendement réel constant de 4 %, trente retraits annuels de fin d'année peuvent atteindre environ 5,8 % du capital initial en pouvoir d'achat constant.
- Le retrait historique d'environ 4 % s'obtient en rejouant des retraites. Comparer leur résultat à celui d'une rente constante ne suffit pas à mesurer le risque de séquence seul.
- Pour voir l'effet propre de l'ordre, changez l'ordre des rendements en gardant les montants et les dates de retrait identiques.
- Une dépense indexée pendant cinquante ans, des frais récurrents et des placements sur d'autres marchés appellent chacun un nouveau calcul.
- Devant un autre taux publié, vérifiez l'horizon retenu, les marchés étudiés et la possibilité de réduire les dépenses en cours de route.

---

## Pour aller plus loin

- William Bengen, [« Determining Withdrawal Rates Using Historical Data »](https://www.financialplanningassociation.org/article/journal/MAR04-determining-withdrawal-rates-using-historical-data) (1994). Les graphiques donnent la durée du portefeuille selon l'année de départ et le retrait initial.
- Early Retirement Now, série SWR (notamment les volets sur le CAPE et l'horizon) : la cascade recalculée sur données modernes ([[serie-ern]]).
- Blanchett, Kowara & Chen, « Optimal Withdrawal Strategy for Retirement Income Portfolios » (2012) et les rapports State of Retirement Income de Morningstar : le bonus d'amortissement dans les taux par âge.
- Dans ce livre : [[rendements-arithmetiques-geometriques]] (l'étage 1 en détail), [[amortissement-abw]] (l'étage 2 érigé en stratégie), [[sequence-des-rendements]] (l'étage 3), [[la-regle-des-4-pourcents]] (la règle racontée) et [[etude-trinity]] (la règle mesurée).
