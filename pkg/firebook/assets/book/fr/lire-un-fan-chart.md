# Lire un fan chart et des percentiles sans se tromper

Vous voulez savoir ce qu'il vous restera à 75 ans. Sur le graphique en éventail, repérez cet âge sur l'axe horizontal, puis lisez les montants à la verticale. La bande sombre contient la moitié centrale des soldes simulés à cet âge. Le trait du milieu indique le solde médian.

Pour savoir comment un ménage en est arrivé là, changez de lecture et suivez une seule des trajectoires fines. Elle peut être au-dessus de la médiane à 65 ans et sous la bande sombre à 75 ans.

La première lecture compare des ménages à une date. La seconde suit une retraite dans le temps.

::: cle Une bande à 90 %
Les bornes de la bande sont recalculées à chaque date. Un ménage peut la quitter une année et y rentrer l'année suivante. Pour savoir combien de ménages y restent pendant toute leur retraite, il faudrait suivre chaque trajectoire du début à la fin.
:::

## Anatomie d'un cône

::: figure fan-anatomy
Un plan à 45 ans d'horizon. Les bandes emboîtées sont des **quantiles par date** (25-75 %, 5-95 %), pas des scénarios. La médiane monte. Et les deux chemins d'exemple traversent les bandes : celui qui finit ruiné est parti par le haut, preuve qu'une bande n'est pas une trajectoire.
:::

Prenons le cône ci-dessus, élément par élément.

**Les bandes emboîtées.** Du plus foncé au plus clair, elles couvrent des intervalles de percentiles croissants autour de la médiane : ici la moitié centrale des futurs (25-75) et la quasi-totalité (5-95). À chaque date, on a trié les milliers de richesses simulées et tracé les quantiles ; la bande n'est donc pas un « scénario prudent » ou « optimiste », c'est une statistique d'ensemble, recalculée indépendamment à chaque pas de temps.

**La ligne médiane.** À chaque date, la médiane partage les soldes en deux moitiés. Les ménages changent de rang d'une année sur l'autre ; rien n'oblige une retraite individuelle à suivre la ligne. La moyenne répond à une autre question. Quelques soldes très élevés peuvent la tirer bien au-dessus de ce que détient le ménage du milieu.

**Les chemins d'exemple.** Les deux trajectoires fines de la figure (verte et rouge) sont des futurs individuels, tracés pour une seule raison : montrer qu'un chemin réel traverse les bandes au lieu d'en longer une. La verte prospère en zigzaguant ; la rouge part dans le peloton (elle passe même au-dessus de la médiane les premières années) puis s'effondre jusqu'à la ruine.

Certains graphiques montrent huit trajectoires choisies en classant tous les ménages par patrimoine final. On prend le premier, le dernier et six rangs régulièrement espacés entre eux. Même si un seul ménage sur mille finit à zéro, il peut donc y avoir une ligne rouge parmi les huit. Pour connaître la fréquence des échecs, divisez leur nombre par le nombre total de simulations. Les lignes d'exemple servent à voir *comment* un échec s'est produit.

**La ligne de zéro et l'échelle.** À zéro, le portefeuille est vide. Pour savoir si le ménage peut encore payer ses dépenses, regardez aussi sa pension et le montant du prochain retrait ([[ruine-et-probabilites]]). L'axe peut être plafonné, par exemple à dix fois le capital initial. Quelques ménages très riches sortent alors du cadre, mais on distingue mieux ceux qui manquent de réserves.

## Pourquoi le cône a cette forme

Trois éléments méritent une lecture séparée : la largeur des bandes, le niveau du ménage médian et la distance des soldes les plus faibles au zéro.

**Pourquoi il s'évase.** Sans retraits, si les rendements logarithmiques annuels sont indépendants et de même variance, l'écart-type du rendement logarithmique cumulé croît comme la racine carrée du temps (√t). Avec des retraits, la courbe représente aussi l'effet des ventes de titres sur le capital restant. Sa largeur seule ne permet donc pas de déterminer si les mauvaises années ont tendance à se suivre ([[monte-carlo-forces-faiblesses]]).

**Quand la médiane monte.** Dans la figure, le solde du ménage médian augmente alors que certains ménages épuisent leur portefeuille. Les deux résultats sont compatibles. À l'inverse, un plan peut prévoir de consommer une partie du capital avant le début d'une pension. Sa médiane peut alors baisser pendant que les dépenses prévues restent couvertes ([[horizon-et-esperance-de-vie]]).

**Le bas du cône plonge d'abord, ou pas.** C'est le trait le plus utile à lire, et il vaut d'être vu sur deux plans côte à côte.

::: figure fan-two-plans
Les deux plans couvrent le même horizon. À gauche, le ménage dispose d'une réserve de liquidités, de dépenses ajustables et de revenus précoces. À droite, il dispose de moins de marge. Comparez leur 5e percentile au bout de dix ans.
:::

Dans les premières années, le 5e percentile se rapproche plus vite de zéro à droite. Pour en comprendre la raison, ouvrez une trajectoire qui a mal commencé et regardez quels retraits ont suivi la baisse. Si vous voulez mesurer l'effet du matelas de liquidités, refaites la comparaison en ne modifiant que ce matelas. Les deux plans de la figure diffèrent aussi par les revenus et la flexibilité ([[sequence-des-rendements]]).

::: science Percentiles ponctuels et trajectoires : la « percentile path fallacy »
Le ménage au 5e percentile en année 10 peut être différent de celui qui occupait ce rang en année 5. Pour répondre à « que sont devenus les ménages du bas de la distribution à cinq ans ? », prenez tous ceux qui se trouvaient sous le 10e percentile à cet âge du plan. Suivez ces mêmes trajectoires jusqu'au bout et comptez combien ont épuisé leur portefeuille. Le cône seul ne fournit pas cette probabilité conditionnelle ([[quand-s-inquieter]]).
:::

## Les cinq erreurs de lecture, et leur correction

**Erreur 1 : lire la médiane comme une promesse.** « Le simulateur dit que j'aurai 2,8 M€ à 75 ans. » Non. Il dit que la moitié des futurs simulés dépassent ce montant si les hypothèses tiennent. Correction : la médiane sert à comparer des plans entre eux et à calibrer les décisions de type legs/dépenses ([[depenses-en-retraite]]) ; les décisions de sécurité se prennent sur le bas du cône et la ruine.

**Erreur 2 : choisir sa bande comme on choisit un menu.** « Je planifie sur le 25e percentile, c'est prudent. » Le percentile ponctuel n'est pas un scénario vivable (la percentile path fallacy ci-dessus) ; et la prudence par percentile est incohérente dans le temps. Correction : la prudence se règle par les **modèles** (planifier entre un central honnête et un modèle pessimiste, [[historique-vs-parametrique]]) et par le seuil de ruine acceptable, pas en se promenant dans les bandes.

**Erreur 3 : oublier que le cône est conditionnel au modèle.** Le cône d'un modèle paramétrique central et celui d'un rejeu historique dur sont deux objets différents pour le même plan. L'écart entre ces cônes montre votre sensibilité au choix du modèle, une part de l'incertitude épistémique, celle du monde inconnu : on ne sait pas lequel est le bon. Il dépasse souvent la largeur d'un seul cône, qui ne mesure, elle, que l'incertitude aléatoire, le hasard dans un monde donné. Les meilleurs simulateurs affichent donc plusieurs cônes, un par modèle de marché, à la même échelle. Quand ces cônes se ressemblent, le plan est robuste au choix de modèle. Quand ils divergent, cet écart mérite plus d'attention que la largeur d'un cône, et la décision doit rester acceptable sous les modèles durs. Correction : lisez d'abord la différence entre les cônes, ensuite la forme de chacun.

**Erreur 4 : l'illusion d'échelle.** Un axe linéaire écrase visuellement les premières années (là où tout se joue) et dramatise les dernières ; un axe logarithmique ferait l'inverse et rendrait le zéro... impossible à tracer (log 0 = −∞). Le compromis courant est un axe linéaire écrêté (voir l'anatomie), qui garde le zéro et la lisibilité du début. Correction : quel que soit l'outil, demandez-vous toujours ce que l'échelle choisie amplifie et ce qu'elle cache, et cherchez la ligne de zéro avant tout.

**Erreur 5 : compter les pixels plutôt que les probabilités.** La surface visuelle de la zone « ruine » dépend du nombre de chemins tracés, de l'épaisseur des traits, de l'écrêtage : rien de tout cela n'est une probabilité. Correction : le cône donne la **forme** du risque (quand, comment, avec quelle brutalité) ; les **chiffres** viennent de la probabilité de ruine et du tableau de résultats. Les deux se lisent ensemble, jamais l'un pour l'autre.

::: astuce Mise en pratique : manipuler un vrai cône
Rien ne vaut un cône vivant, calculé sur votre propre plan. Un quart d'heure suffit. Dans le simulateur FIRE de pofo, la section des futurs simulés affiche quatre cônes du même plan, à la même échelle, un par modèle de marché.

- Commencez par le cône central, seul. Repérez la médiane, la bande 25-75, la bande 5-95, la ligne de zéro.
- Passez au cône le plus dur, celui de la décennie perdue. Comparez les deux bas de cône avant de comparer les deux médianes.
- Suivez une trajectoire qui finit à zéro. Repérez la première année de baisse et les retraits qui la suivent. Cherchez ensuite le taux d'échec dans le tableau de l'ensemble des simulations.
- Basculez enfin la règle de retrait, du fixe indexé vers une règle flexible. Le bas du cône remonte, et l'éventail des dépenses servies montre ce que cette sécurité coûte à vivre.

Le même parcours se fait ailleurs, dans tout outil qui trace une distribution de trajectoires (TPAW Planner et Portfolio Visualizer en bandes de percentiles, cFIREsim et FICalc en faisceau de millésimes historiques). Il faut seulement relancer le calcul modèle par modèle, puis garder les cônes côte à côte sous les yeux.
:::

## Quelles dépenses, quel héritage et quels échecs ?

L'éventail de patrimoine laisse trois questions ouvertes. Chacune demande une lecture différente des trajectoires.

**L'éventail des dépenses servies.** Avec une règle flexible (guardrails, VPW, ABW, [[panorama-strategies-retrait]]), tracez aussi ce que le ménage peut effectivement dépenser chaque année. Si le 10e percentile est inférieur de 35 % au budget pendant douze dates successives, vous savez que le bas de la distribution reste très bas. Pour connaître la durée des coupes subies par *un même ménage*, comptez les années consécutives sous son plancher dans chaque trajectoire ([[flexibilite-realite]]).

**Le patrimoine final.** Un histogramme des soldes à l'horizon choisi permet de lire la fréquence des patrimoines faibles ou élevés. Si l'horizon est fixé à 90 ans, il s'agit du capital à 90 ans. Pour estimer un héritage au décès, il faut également tenir compte de l'âge auquel le décès survient ([[horizon-et-esperance-de-vie]], [[depenses-en-retraite]]).

**Les formes d'échec.** Parmi les ménages qui ont épuisé leur portefeuille, certains ont perdu la moitié de leur capital dès les premières années. D'autres ont vu leur solde s'éroder lentement ou ont épuisé un capital qui avait d'abord augmenté. Ces profils suggèrent où regarder. Pour comprendre un cas, confrontez les rendements, les dépenses et les revenus extérieurs année par année ([[glidepaths]], [[rentes-et-annuites]]).

::: exemple Combien de temps faudra-t-il vivre avec moins ?
Un ménage prévoit 55 000 € par an pendant 48 ans et accepte de réduire cette dépense si les marchés baissent. Son éventail de dépenses montre un 25e percentile inférieur de 12 % au budget de confort pendant six années successives. Pour savoir si la règle lui convient, il reste deux questions concrètes. Les dépenses au niveau réduit couvrent-elles son plancher ([[combien-il-vous-faut]]) ? Et dans les trajectoires concernées, combien d'années d'affilée dure la coupe ? La première réponse se lit sur le budget du ménage. La seconde exige de suivre les mêmes trajectoires d'une année sur l'autre.
:::

## L'essentiel à retenir

- Un fan chart est une pile de distributions par date, pas un faisceau de futurs : les bandes sont des quantiles ponctuels, les vraies trajectoires les traversent en zigzag ; la médiane n'est ni une trajectoire ni la moyenne.
- Une médiane qui monte peut coexister avec des échecs. Comparez les dépenses financées et le capital restant aux objectifs du plan.
- Des chemins choisis par rang final permettent d'examiner des réussites et des échecs. La fréquence de la ruine se calcule sur l'ensemble des simulations.
- Les cinq erreurs : médiane-promesse, percentile-scénario, oubli du conditionnement au modèle (lisez d'abord l'écart entre cônes), illusion d'échelle, pixels pris pour des probabilités.
- Ouvrez les autres graphiques selon la question qui vous préoccupe. Quelles dépenses ont été servies ? Quel capital reste-t-il au décès ? Qu'est-il arrivé aux ménages qui ont manqué d'argent ?

---

## Pour aller plus loin

- Bank of England, *Inflation Report fan charts* (la note méthodologique historique) : l'origine du format et de ses conventions.
- Early Retirement Now, volet 46 (la fausse précision des sorties de simulation) ([[serie-ern]]).
- Les cônes multi-modèles, les éventails de dépenses, d'héritages et de causes : [[utiliser-la-page-fire]] et [[la-machine-pofo]].
- La suite logique : [[pieges-des-simulateurs]] (les biais en amont du graphique) et [[quand-s-inquieter]] (le bon outil pour la question « où en suis-je sur ma trajectoire ? »).
