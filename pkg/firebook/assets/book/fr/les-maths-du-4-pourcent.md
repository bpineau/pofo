# Pourquoi 4 % ? L'anatomie mathématique de la règle

Le chapitre [[la-regle-des-4-pourcents]] raconte d'où vient la règle et [[etude-trinity]] comment elle a été établie. Celui-ci répond à la question qu'on pose rarement : pourquoi ce chiffre-là ? Pourquoi pas 6 %, puisque les actions américaines ont rapporté 6 à 7 % réels en moyenne ? Pourquoi pas 2 %, puisque tout peut arriver ? Le 4 % n'est pas une constante magique tombée d'un backtest. On peut le lire **en cascade**, en comparant le retrait à rendement constant à celui qu'ont permis les pires départs historiques. Une fois la cascade comprise, la règle cesse d'être un dogme à croire ou à démentir. Elle devient un résultat dont vous comprenez les ressorts, et dont vous savez prévoir le sens quand une hypothèse change.

C'est probablement le chapitre le plus pédagogique du livre : trois étages, une addition, deux soustractions, aucun outil au-delà de la calculatrice. Après lecture, vous saurez reconstruire le chiffre de tête, dire précisément ce qui le ferait monter ou descendre, et comprendre pourquoi il s'est montré si étonnamment robuste à travers un siècle qui a connu 1929, 1966 et 2000.

::: cle L'idée en une phrase
À rendement réel **constant** de 4 %, on peut retirer 4 % du capital initial chaque fin d'année indéfiniment, ou environ 5,8 % par an pendant trente ans. Dans l'étude américaine de Bengen, un retrait initial voisin de 4 %, ensuite indexé sur l'inflation, a résisté aux mauvais départs sur trente ans. La différence avec 5,8 % vient à la fois des rendements traversés et de leur ordre.
:::

## Étage 1 : la rente perpétuelle, ou le rendement réel comme plafond naturel

Commençons par le monde le plus simple : un portefeuille au rendement constant et connu, et un rentier immortel qui ne veut jamais entamer son capital. La réponse tient en une ligne : le retrait soutenable est exactement le rendement **réel** du portefeuille. Réel, car le retrait doit suivre l'inflation ([[inflation-et-taux-de-retrait]]). Tout raisonnement en nominal est une illusion d'optique qui coûte 2 à 3 points.

Reste à chiffrer ce rendement. Un 60/40 mondial a historiquement produit autour de 4,6 % réels, en moyenne **arithmétique**. Mais le capital ne compose pas la moyenne arithmétique. Il compose la **géométrique**, amputée du frein de volatilité (volatility drag), soit environ la moitié de la variance ([[rendements-arithmetiques-geometriques]]). Résultat : environ **4 % réels géométriques** pour le 60/40 historique. Premier jalon posé. L'immortel prudent, au portefeuille classique, peut retirer environ 4 %. Ce chiffre reste un rendement, pas encore un taux de retrait. La coïncidence numérique avec la règle finale est un hasard qui entretient la confusion, car les deux étages suivants vont s'annuler presque parfaitement.

## Étage 2 : le bonus d'amortissement, ou le droit de mourir

Le rentier n'est pas immortel. Sur un horizon de 30 ans, il a le droit de consommer le capital lui-même, pas seulement ses fruits. La finance de crédit donne la formule exacte, celle d'une mensualité de prêt inversée : à 4 % réels sur 30 ans, le retrait qui épuise exactement le capital au dernier jour est d'environ 5,8 % par an. Le droit de finir à zéro vaut donc +1,8 point. C'est le **bonus d'amortissement**. C'est aussi toute la logique des méthodes ABW/VPW, qui le recalculent chaque année ([[amortissement-abw]], [[vpw]]).

Ce bonus explique au passage la sensibilité à l'horizon ([[horizon-et-esperance-de-vie]]). À 50 ans d'horizon, le même calcul donne 4,7 % et le bonus fond à +0,7. À l'infini, il disparaît. La retraite très précoce ne « perd » pas son taux à cause d'un mystère. Elle perd simplement le droit d'amortir vite.

::: figure amortissement-horizon
Le retrait qu'une pure arithmétique d'amortissement finance à 4 % réel, selon le nombre d'années à couvrir. La courbe tombe vite, puis s'aplatit sur son asymptote, le rendement seul. Les vingt premières années emportent l'essentiel du bonus, tandis que passer de 50 ans d'horizon à l'éternité ne coûte plus que sept dixièmes de point. La retraite très précoce paie donc son horizon, mais beaucoup moins qu'on ne le croit.
:::

## Étage 3 : la pénalité de séquence, ou le désordre qui coûte

Vous avez 100 € et en retirez 10 à la fin de chaque année. Si le marché perd 20 % puis gagne 25 %, il vous reste 70 € après le premier retrait et 77,50 € après le second. Inversez les rendements : il vous reste 115 € puis 82 €. Sans les retraits, les deux ordres auraient ramené les 100 € de départ. Avec les retraits, la perte précoce réduit la somme qui profite du rebond ([[sequence-des-rendements]]).

Bengen a comparé des départs à des dates différentes aux États-Unis. Les retraités n'ont donc ni les mêmes rendements sur trente ans, ni les mêmes années de forte inflation au début de leur retraite. Les quelque 1,8 point entre notre rente théorique et le retrait historiquement soutenable illustrent cet écart. Pour isoler le coût de l'ordre seul, il faudrait garder exactement les mêmes rendements réels et les permuter, comme dans l'exemple des 100 €.

::: figure cascade-4pct
Les deux premières marches se calculent à rendement constant. La dernière relie ce calcul au retrait observé sur l'histoire américaine. La barre « séquence » inclut aussi le niveau des rendements propres aux périodes difficiles.
:::

::: exemple La cascade complète, sur un coin de table
60/40 historique, 30 ans, retrait fixe indexé. Rendement réel arithmétique ≈ 4,6 % → moins le volatility drag ≈ 4 % géométrique (étage 1) → plus le bonus d'amortissement 30 ans ≈ 5,8 % (étage 2) → moins l'écart avec le pire départ historique ≈ 4,0 % (étage 3). Voilà la règle, reconstruite. Maintenant, faites-la respirer. Horizon de 50 ans → le bonus fond, la pénalité s'allège un peu, ~3,4 % (le −1,8 de l'étage 3 est calibré sur trente ans). Échantillon mondial plutôt que le seul cas américain ([[anarkulova-cederburg]]) → −0,5 à −1 point. CAPE élevé au départ ([[valorisations-et-cape]]) → étage 1 raboté d'un point. 0,5 % de frais → −0,5, presque un pour un. Règle flexible à plancher → +0,3 à +0,5. Chaque débat sur « le vrai chiffre » de la règle porte sur un seul étage, et il se règle étage par étage, pas par slogans.
:::

::: figure clavier-leviers
Chaque hypothèse déplacée seule, à partir du plan de référence de la cascade : 60/40 historique, 30 ans, retrait fixe indexé, soit 4,0 %. Les fourchettes sont celles de l'article, et le signe de chaque barre dit dans quel sens le levier joue.
:::

## Pourquoi c'est si robuste (et ce qui le casserait)

Un taux choisi pour traverser le départ américain le plus difficile laisse souvent un capital important aux ménages partis à une date plus favorable ([[lire-un-fan-chart]]). Pourquoi garder cette marge ? Au moment de partir, vous ignorez à quel groupe vous appartiendrez. Le capital laissé par un autre ménage ne paiera pas vos dépenses si votre propre retraite commence mal.

La liste de ce qui la casserait est donc précise. Un régime **hors échantillon** pour votre pays (le retraité de Tokyo 1990 ou de Buenos Aires a connu pire que le pire américain, [[diversification-internationale]], [[pieges-des-simulateurs]]). Des **frottements** non provisionnés (la cascade est calculée brute, frais et fiscalité se retranchent de l'étage 1 presque un pour un, [[etf-ucits-europeens]], [[flat-tax-et-imposition]]). Une **indexation rigide** sur un horizon très long, la combinaison qui a le moins de marge. Et l'**abandon en cours de route**, statistiquement le tueur numéro un, contre lequel la cascade ne peut rien ([[psychologie-du-retrait]]).

::: science Ce que les reconstructions modernes confirment
Les travaux cités en fin de chapitre explorent des questions différentes. Early Retirement Now compare les retraits selon la valorisation de départ et l'allocation ([[serie-ern]]). Blanchett et ses coauteurs étudient l'ajustement des dépenses sur la durée de la retraite. Anarkulova et Cederburg élargissent l'échantillon de marchés. Pour rapprocher deux taux publiés, relevez d'abord la durée, les marchés retenus et la définition de l'échec.
:::

## L'essentiel à retenir

- Le 4 % se reconstruit en trois étages : ~4 % de rendement réel géométrique du 60/40 (après volatility drag), +1,8 point de bonus d'amortissement sur 30 ans, −1,8 point de différence avec le pire départ historique (niveau et ordre des rendements). La quasi-annulation du bonus et de la pénalité est le « miracle » de la règle.
- Chaque étage a ses leviers. L'horizon joue sur le bonus (50 ans, +0,7 seulement). La volatilité et la flexibilité jouent sur la pénalité. Les valorisations, les frais et l'échantillon jouent sur l'étage rendement, presque un pour un pour les frais.
- Pour voir l'effet propre de l'ordre, changez l'ordre des rendements en gardant les montants et les dates de retrait identiques.
- Ce qui casse la cascade est connu : régime hors échantillon national, frottements non provisionnés, indexation rigide sur très long horizon, abandon en route.
- Savoir refaire ce calcul de tête vaut tous les débats : quand quelqu'un annonce « le 4 % est mort » ou « 5 % passe très bien », demandez-lui quel étage il a modifié, et de combien.

---

## Pour aller plus loin

- William Bengen, [« Determining Withdrawal Rates Using Historical Data »](https://www.financialplanningassociation.org/article/journal/MAR04-determining-withdrawal-rates-using-historical-data) (1994). Les graphiques donnent la durée du portefeuille selon l'année de départ et le retrait initial.
- Early Retirement Now, série SWR (notamment les volets sur le CAPE et l'horizon) : la cascade recalculée sur données modernes ([[serie-ern]]).
- Blanchett, Kowara & Chen, « Optimal Withdrawal Strategy for Retirement Income Portfolios » (2012) et les rapports State of Retirement Income de Morningstar : le bonus d'amortissement dans les taux par âge.
- Dans ce livre : [[rendements-arithmetiques-geometriques]] (l'étage 1 en détail), [[amortissement-abw]] (l'étage 2 érigé en stratégie), [[sequence-des-rendements]] (l'étage 3), [[la-regle-des-4-pourcents]] (la règle racontée) et [[etude-trinity]] (la règle mesurée).
