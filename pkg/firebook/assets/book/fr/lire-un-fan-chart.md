# Lire un fan chart et des percentiles sans se tromper

Le graphique le plus dense de toute la planification de retraite est le cône de richesse, le « fan chart » : des milliers de trajectoires simulées résumées en bandes de percentiles qui s'évasent avec le temps, quelques chemins d'exemple, une ligne de zéro. C'est aussi le graphique le plus mal lu : on y voit des promesses là où il y a des fréquences, des scénarios là où il y a des quantiles ponctuels, et de l'incertitude du monde là où il y a parfois surtout l'incertitude du modèle.

Or ce graphique est au cœur de la section §01 de la page FIRE de pofo, et les mêmes conventions se retrouvent partout (Banque d'Angleterre, qui a popularisé le format pour l'inflation, projections climatiques, tous les simulateurs sérieux). Cette page apprend à le lire en professionnel : l'anatomie exacte du cône, pourquoi il a la forme qu'il a (chaque trait de sa géométrie est un fait de probabilité), les cinq erreurs de lecture classiques avec leur correction, et la façon dont pofo a résolu quelques problèmes de représentation que la plupart des outils escamotent.

Après cette page, un cône se lira comme une phrase.

::: cle Le renversement de lecture
Un fan chart ne montre **pas** des futurs : il montre, pour chaque date, la **distribution** des états possibles à cette date. La bande à 90 % de l'année 20 dit « à l'année 20, 90 % des trajectoires simulées sont dans cette tranche » ; elle ne dit **rien** sur la façon dont une trajectoire donnée y est arrivée ni où elle ira ensuite. Le cône est une pile de coupes transversales, pas un faisceau de chemins ; les chemins, eux, traversent les bandes en zigzag toute leur vie. Toutes les erreurs de lecture découlent de la confusion entre ces deux objets.
:::

## Anatomie d'un cône

Prenons le cône type de la §01 (un plan à 45 ans d'horizon), élément par élément.

**Les bandes emboîtées.** Du plus foncé au plus clair, elles couvrent des intervalles de percentiles croissants autour de la médiane : typiquement 25-75 (la moitié centrale des futurs), 10-90, 5-95. À chaque date, on a trié les milliers de richesses simulées et tracé les quantiles ; la bande n'est donc pas un « scénario prudent » ou « optimiste », c'est une statistique d'ensemble, recalculée indépendamment à chaque pas de temps.

**La ligne médiane.** Le 50e percentile : à chaque date, la moitié des futurs sont au-dessus, la moitié en dessous. Attention, première subtilité : la ligne médiane n'**est pas** une trajectoire (aucun futur ne longe la médiane ; y rester exigerait un miracle de régularité) et elle n'est pas non plus la moyenne : les distributions de richesse composée sont très asymétriques (bornées à zéro en bas, illimitées en haut), donc la moyenne est tirée au-dessus de la médiane par les scénarios opulents. Quand un vendeur dit « en moyenne, vous finirez avec 4 M€ », il cite souvent la moyenne parce qu'elle flatte ; la médiane est la bonne intuition de « ce qui vous arrivera plausiblement ».

**Les chemins d'exemple.** Huit trajectoires individuelles fines dans pofo, et leur choix est une astuce de conception qui mérite d'être comprise : elles sont prélevées à rangs **régulièrement** espacés de richesse finale (la pire, puis environ les 14e, 29e, 43e... percentiles, jusqu'à la meilleure). Conséquence directe : **le nombre de chemins rouges (échoués) encode le taux d'échec** : un rouge ≈ « le pire septième échoue » (ruine ≤ ~14 %), deux rouges ≈ un futur sur sept, trois ≈ un sur quatre. Un coup d'œil aux rouges donne l'ordre de grandeur de la ruine sans lire un seul chiffre ([[la-machine-pofo]]).

**La ligne de zéro et l'écrêtage du haut.** Le zéro est la seule frontière absolue du graphique : la ruine ([[ruine-et-probabilites]]). Et pofo écrête l'axe vertical à 10 fois le capital initial (« upside clipped ») : sans cela, les scénarios composés à 30-40 ans (qui peuvent atteindre 20-50 fois la mise) écraseraient visuellement la zone du zéro, c'est-à-dire **la** zone qui motive tout l'exercice. C'est un choix de représentation assumé : on sacrifie le spectacle de la queue haute pour garder lisible la queue basse, celle qui décide.

## Pourquoi le cône a cette forme

Chaque trait géométrique du cône est un théorème déguisé ; les connaître, c'est savoir diagnostiquer un plan d'un regard.

**Il s'évase, et de plus en plus lentement.** Pour des rendements sans mémoire, l'incertitude cumulée croît comme la racine du temps (√t) : le cône s'élargit vite au début, puis de moins en moins. Un cône qui s'évase **plus** que √t signale de la mémoire aggravante (les retraits : un début raté creuse un écart que les retraits fixes amplifient ensuite) ; les vrais marchés, avec leur léger retour de valorisation, s'évasent un peu **moins** à très long terme : c'est l'une des raisons pour lesquelles les modèles i.i.d. sur-dispersent légèrement l'horizon lointain ([[monte-carlo-forces-faiblesses]]).

**La médiane d'un plan sain MONTE.** à retrait de 3-3,5 % et rendement réel espéré supérieur ([[rendements-attendus]]), le scénario central croît plus vite qu'on ne ponctionne : la médiane à 30 ans dépasse souvent le double du capital initial en termes réels. Voir simultanément une médiane opulente **et** des chemins rouges n'est pas une contradiction : c'est la définition même d'un plan de retrait, dont l'asymétrie (médiane riche, queue basse ruinée) est irréductible ([[horizon-et-esperance-de-vie]] explique pourquoi survivre aux 30 premières années suffit presque toujours).

**Le bas du cône plonge d'abord, ou pas.** Regardez le 5e percentile des dix premières années : c'est la fenêtre fragile rendue visible ([[sequence-des-rendements]]). Un plan bien défendu (buffer, flexibilité, revenus précoces) montre un bas de cône qui s'enfonce lentement ; un plan à retrait tendu sans marges montre un 5e percentile qui pique vers zéro dès les années 5-10. La **pente** du bas du cône au début est le meilleur résumé visuel de votre exposition à la séquence.

::: science Percentiles ponctuels et trajectoires : la « percentile path fallacy »
L'erreur conceptuelle la plus profonde du fan chart mérite son encart. Le « 5e percentile » tracé n'est **pas** un scénario : c'est la couture de milliers de scénarios différents, chacun n'occupant ce rang qu'un moment. Une vraie trajectoire du pire décile visite typiquement la bande médiane certaines années ; symétriquement, la trajectoire qui **finit** ruinée a souvent passé ses premières années dans la moitié haute (c'est le millésime 2000 : deux belles années avant le plongeon, [[etude-trinity]]). Conséquence pratique : on ne peut **pas** lire sur le cône « si je suis sous le 10e percentile en année 5, je finirai ruiné » : le cône ne contient pas les probabilités **conditionnelles** le long des chemins. Cette question (« où en suis-je, et que dois-je en conclure ? ») est légitime et cruciale, mais elle exige un autre outil : le suivi par seuils de taux de retrait courant ([[quand-s-inquieter]]), pas la contemplation du cône.
:::

## Les cinq erreurs de lecture, et leur correction

**Erreur 1 : lire la médiane comme une promesse.** « Le simulateur dit que j'aurai 2,8 M€ à 75 ans. » Non : il dit que la moitié des futurs simulés dépassent ce montant **si** les hypothèses tiennent. Correction : la médiane sert à comparer des plans entre eux et à calibrer les décisions de type legs/dépenses ([[depenses-en-retraite]]) ; les décisions de sécurité se prennent sur le bas du cône et la ruine.

**Erreur 2 : choisir sa bande comme on choisit un menu.** « Je planifie sur le 25e percentile, c'est prudent. » Le percentile ponctuel n'est pas un scénario vivable (erreur de la percentile path fallacy ci-dessus) ; et la prudence par percentile est incohérente dans le temps. Correction : la prudence se règle par les **modèles** (planifier entre central et broad-sample, [[historique-vs-parametrique]]) et par le seuil de ruine acceptable, pas en se promenant dans les bandes.

**Erreur 3 : oublier que le cône est conditionnel au modèle.** Le cône du modèle central et celui du broad-sample sont deux objets différents pour le même plan ; l'écart **entre** cônes (l'incertitude épistémique : on ne sait pas quel monde est le bon) est souvent plus grand que la largeur d'**un** cône (l'incertitude aléatoire : le hasard dans un monde donné). C'est pour rendre cet écart visible que la §01 de pofo affiche les quatre cônes **côte** à **côte**, même échelle, même plan. Correction : lisez d'abord la différence entre les cônes, ensuite la forme de chacun.

**Erreur 4 : l'illusion d'échelle.** Un axe linéaire écrase visuellement les premières années (là où tout se joue) et dramatise les dernières ; un axe logarithmique ferait l'inverse et rendrait le zéro... impossible à tracer (log 0 = −∞). Le compromis de pofo : axe linéaire écrêté à 10×, qui garde le zéro **et** la lisibilité du début. Correction : quel que soit l'outil, demandez-vous toujours ce que l'échelle choisie amplifie et ce qu'elle cache, et cherchez la ligne de zéro avant tout.

**Erreur 5 : compter les pixels plutôt que les probabilités.** La surface visuelle de la zone « ruine » dépend du nombre de chemins tracés, de l'épaisseur des traits, de l'écrêtage : rien de tout cela n'est une probabilité. Correction : le cône donne la **forme** du risque (quand, comment, avec quelle brutalité) ; les **chiffres** viennent de la jauge de ruine et du tableau ([[utiliser-la-page-fire]]). Les deux se lisent ensemble, jamais l'un pour l'autre.

## Au-delà de la richesse : les autres éventails de la page

Le format « distribution par date » sert trois autres lectures dans pofo, chacune répondant à une question que le cône de richesse ne traite pas.

**L'éventail des dépenses servies (§04).** Dès que votre plan a une règle flexible (guardrails, VPW, ABW, [[panorama-strategies-retrait]]), la richesse ne suffit plus : la question devient « que vais-je **vivre** ? ». La §04 montre la distribution du niveau de vie année par année : son bas de cône à vous, c'est le train de vie du mauvais quart des futurs, pendant combien d'années, financé par quoi (portefeuille, pension, buffer). C'est **la** vue qui départage les règles de retrait : une règle qui « réussit toujours » avec un 10e percentile de dépenses à −35 % pendant douze ans n'a pas éliminé le risque, elle l'a déplacé de la faillite vers la vie ([[flexibilite-realite]]).

**La distribution des héritages (§05, bequest).** La coupe **finale** du cône de richesse, présentée en histogramme : combien reste-t-il au bout, dans tous les futurs ? La lecture type d'un plan sain choque toujours un peu : la masse est loin au-dessus du capital initial (on meurt le plus souvent riche, [[une-annee-de-plus]], [[depenses-en-retraite]]), avec une petite barre à zéro : la ruine, vue de l'autre bout.

**Les causes d'échec (§05, causes).** Parmi les seuls futurs ruinés, la forme de la trajectoire dit le mode de défaillance : krach précoce (richesse divisée par deux dans les dix premières années : la catastrophe de séquence), érosion lente (le portefeuille n'a jamais décollé : la décennie perdue), ou longévité (le plan a prospéré, culminé, et la retraite lui a survécu). Trois modes, trois parades différentes ([[glidepaths]] et buffer pour le premier, actifs de régime pour le deuxième, rentes pour le troisième [[rentes-et-annuites]]) : c'est le graphique qui transforme « 6 % de ruine » en diagnostic actionnable.

::: exemple Une lecture complète, en quatre regards
Plan : 1,6 M€, 55 000 €/an avec guardrails, 48 ans d'horizon. Regard 1, les quatre cônes côte à côte : formes semblables, mais le bas du cône broad-sample s'enfonce nettement plus vite : l'incertitude dominante est épistémique, le monde compte plus que le hasard. Regard 2, le bas du cône central sur 10 ans : pente douce, un seul chemin rouge sur huit : exposition à la séquence contenue (~10 % ou moins). Regard 3, la §04 : le 25e percentile des dépenses passe 6 ans à −12 % du confort : c'est le prix réel des guardrails, jugé acceptable contre le plancher établi ([[combien-il-vous-faut]]). Regard 4, les causes : échecs résiduels aux trois quarts « érosion lente » : la parade prioritaire n'est pas plus de cash mais des briques anti-décennie-perdue ([[portefeuilles-tous-temps]]). Quatre regards, quatre décisions instruites : voilà ce qu'un fan chart bien lu livre en deux minutes.
:::

## L'essentiel à retenir

- Un fan chart est une pile de distributions par date, pas un faisceau de futurs : les bandes sont des quantiles ponctuels, les vraies trajectoires les traversent en zigzag ; la médiane n'est ni une trajectoire ni la moyenne.
- La géométrie parle : évasement ≈ √t (plus = mémoire aggravante des retraits), médiane montante = plan sain, pente du bas de cône des 10 premières années = votre exposition à la séquence.
- Les huit chemins d'exemple de pofo sont à rangs réguliers : compter les rouges donne l'ordre de grandeur de la ruine d'un coup d'œil ; l'axe est écrêté à 10× pour garder le zéro lisible.
- Les cinq erreurs : médiane-promesse, percentile-scénario, oubli du conditionnement au modèle (lisez d'abord l'écart **entre** cônes), illusion d'échelle, pixels pris pour des probabilités.
- Les autres éventails complètent : dépenses servies (le vrai juge des règles flexibles), héritages (la ruine vue de l'autre bout), causes d'échec (le diagnostic qui choisit la parade).

---

## Pour aller plus loin

- Bank of England, *Inflation Report fan charts* (la note méthodologique historique) : l'origine du format et de ses conventions.
- Early Retirement Now, Part 46 (la fausse précision des sorties de simulation) ([[serie-ern]]).
- Dans pofo : §01 (les quatre cônes), §04 (l'éventail des dépenses), §05 (bequest et causes), et les aides au survol de chaque graphique ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
- La suite logique : [[pieges-des-simulateurs]] (les biais en amont du graphique) et [[quand-s-inquieter]] (le bon outil pour la question « où en suis-je sur ma trajectoire ? »).
