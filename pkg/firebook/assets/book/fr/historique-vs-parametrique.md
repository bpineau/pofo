# Fenêtres historiques, bootstrap, paramétrique : trois familles de modèles

Un simulateur de retraite déroule des plans dans des futurs générés ([[monte-carlo-forces-faiblesses]]). Toute la question est de savoir d'où viennent ces futurs. Il n'existe que trois grandes réponses, trois familles de modèles. On peut rejouer l'histoire telle quelle (les fenêtres historiques). On peut la rééchantillonner en la remélangeant (le bootstrap). On peut enfin générer des rendements synthétiques à partir de quelques paramètres (les modèles paramétriques).

Chaque famille répond en réalité à une question différente. Chacune a ses vertus et ses angles morts. Et quand leurs verdicts divergent sur un même plan, ce n'est pas une anomalie : c'est l'information la plus précieuse que vous obtiendrez. Cette page fait le tour complet des trois familles. Pour chacune, on verra la mécanique exacte, les choix d'implémentation qui comptent, ses forces et ses pièges. À la fin vient la grille de lecture combinée, qui dit quel modèle croire pour quelle question.

::: cle Trois familles, trois questions
Fenêtres historiques : « qu'aurait donné mon plan si j'étais parti à chaque date du passé ? » Bootstrap : « et dans des histoires **plausibles** faites des mêmes ingrédients que le passé, mais assemblés autrement ? » Paramétrique : « et dans un monde dont je choisis explicitement la moyenne, la volatilité et les queues ? » Aucune ne répond à « que va-t-il se passer ? » ; ensemble, elles **encadrent** cette question inaccessible. C'est pourquoi les simulateurs sérieux affichent plusieurs familles côte à côte, au lieu d'un verdict unique.
:::

## Famille 1 : les fenêtres historiques (rejeu, cohortes)

**La mécanique.** C'est la méthode fondatrice de Bengen ([[etude-trinity]]). On prend la série réelle des rendements de votre portefeuille (ou d'un indice). Puis on déroule le plan à partir de chaque date de départ possible : la fenêtre janvier 1975 → décembre 2019, puis février 1975 → janvier 2020, et ainsi de suite. Chaque fenêtre est un « millésime », une cohorte. Le taux d'échec est la fraction des millésimes ruinés.

**Les choix d'implémentation qui comptent.** Le pas d'échantillonnage, d'abord. Prendre un départ par mois plutôt que par année permet de comparer des départs voisins, même si leurs fenêtres se chevauchent ; c'est la pratique de référence, celle d'ERN notamment ([[serie-ern]]). L'honnêteté face à l'horizon, ensuite. Quand l'historique est plus court que le plan (20 ans de données pour 45 ans de retraite), aucune fenêtre complète n'existe et le taux d'échec n'est tout simplement pas calculable. Un outil sérieux le dit et refuse de répondre, plutôt que d'extrapoler en silence. La variante dirigée, enfin. Au lieu de dérouler toutes les dates de départ, on rejoue les pires : USA 1929, 1966, 2000, Japon 1990. Des millésimes réels choisis pour l'épreuve, et la version la plus parlante de la famille ([[etude-trinity]]).

**Les forces.** La fidélité aux séquences réellement observées. Les corrélations entre actifs, les grappes de crises, les enchaînements krach-inflation-reprise : tout y est, puisque rien n'est modélisé, tout est cité. C'est le seul modèle dont chaque trajectoire s'est réellement produite. D'où une vertu pédagogique inégalée (voir son plan traverser 1966 parle plus que mille probabilités) et un excellent détecteur de fragilité face aux régimes réels.

**Les pièges.** Ils sont trois, et sérieux. D'abord l'échantillon minuscule. Cent ans de données ne contiennent que trois ou quatre retraites longues indépendantes. Les fenêtres se chevauchent massivement : un même krach peut apparaître dans des centaines de fenêtres mensuelles. Le « taux d'échec » porte donc des barres d'erreur énormes, que son affichage ne montre pas. Ensuite le plafond du réalisé. Le pire du passé y est traité comme le pire absolu, alors que rien ne garantit que le pire soit derrière nous. Enfin le biais de la fenêtre disponible. L'historique de vos fonds couvre le plus souvent les décennies récentes, plutôt favorables. Si votre historique couvre une période favorable, le rejeu en portera la trace.

## Famille 2 : le bootstrap (rééchantillonnage par blocs)

**La mécanique.** Choisissez un mois dans l'historique et copiez ensemble les rendements de tous les actifs pour ce mois. Continuez avec les mois suivants, puis repartez d'une autre date tirée au sort. Un même bloc peut revenir plusieurs fois. À l'intérieur d'un bloc on conserve par exemple la chute des actions puis leur reprise, ainsi que le comportement des obligations pendant ces mêmes mois. À la jonction de deux blocs, la continuité historique s'arrête.

Dans le bootstrap stationnaire de Politis et Romano, la longueur des blocs varie. Avec une moyenne de 24 mois, il y a chaque mois une chance sur 24 de repartir d'une date aléatoire. Une crise de trois ans peut donc être reprise en entier, coupée en deux ou absente d'une trajectoire.

::: figure bootstrap-blocs
Le dernier tirage reprend les mois 5 et 6 pour la deuxième fois. À chaque passage, les actions et les obligations gardent leurs rendements respectifs. Chaque bloc dure ici deux mois.
:::

**Un même moteur, deux questions.** Tout dépend du panel que l'on remélange. Appliqué à l'historique de vos propres fonds, sous vos poids courants, le bootstrap fabrique des milliers de variantes de ce que vos lignes ont réellement vécu. Appliqué à un panel long et multi-pays, il change de nature. C'est l'esprit de la méthode d'Anarkulova-Cederburg (38 pays dans leur étude) : rééchantillonner le siècle des pays développés, ici sur le panel académique Jorda-Schularick-Taylor (16 pays, 1870-2020), en tirant chaque bloc dans un même pays et en le faisant long (dix ans en moyenne), pour que les grands désastres (1929, les années 1970, le Japon) survivent en bonne partie dans les trajectoires, sur un portefeuille 60/40 domestique conforme à la littérature ([[anarkulova-cederburg]]). Même famille mathématique, deux questions différentes : « mes fonds, remélangés » contre « le siècle développé, remélangé ».

**Les forces.** Le meilleur compromis entre fidélité et diversité. On garde les corrélations et la mémoire courte du réel (via les blocs), et on obtient des milliers de trajectoires distinctes (via le remélange). Le broad-sample y ajoute la profondeur, avec des régimes que l'historique de vos fonds n'a jamais vus. C'est la famille que la recherche moderne préfère pour estimer le risque de long horizon, et c'est celle d'Anarkulova-Cederburg.

**Les pièges.** La continuité se perd à la jonction des blocs. Une baisse de sept ans peut être reprise dans un long bloc ou résulter de plusieurs mauvais blocs successifs. Le retour des valorisations vers leur moyenne, étalé sur plusieurs décennies ([[valorisations-et-cape]]) disparaît lui aussi. Les ingrédients, ensuite, restent ceux du passé disponible : le bootstrap remélange, il n'invente rien. Appliqué au seul historique de vos fonds, il hérite du biais de fenêtre de la famille 1 : la diversité des tirages ne corrige pas la pauvreté des ingrédients. Enfin, le choix de la longueur de bloc est un vrai paramètre. Trop court, on tue les grappes ; trop long, on retombe dans le rejeu et son échantillon pauvre. Les 24 mois moyens appliqués à vos fonds sont un choix courant, assez long pour contenir une récession type, assez court pour diversifier ; le broad-sample, qui vise les désastres longs, tire des blocs de dix ans en moyenne.

## Famille 3 : le paramétrique (Student-t, et les régimes)

**La mécanique.** On abandonne les données brutes. On décrit le monde par quelques paramètres, et on tire dedans. La version la plus simple est l'i.i.d. gaussien (moyenne, volatilité, tirages annuels indépendants), celle de la plupart des simulateurs commerciaux. Elle a deux défauts corrigibles : des queues trop fines et aucune mémoire. Un paramétrique sérieux corrige le premier en tirant dans une Student-t à trois paramètres : μ (la moyenne), σ (la volatilité) et df (l'épaisseur des queues). À df 5 et à volatilité égale, l'année à −30 % réel est environ trois fois plus probable qu'en loi normale ([[queues-epaisses]]). Les outils les plus soignés ajustent les trois paramètres sur les données de vos fonds, et les mélangent vers le prior broad-sample (actions mondiales, pas vos fonds) quand l'horizon dépasse l'historique ([[rendre-monte-carlo-pertinent]]).

Le second défaut, l'absence de mémoire, donne naissance à la sous-famille des modèles à régimes, paramétriques mais séquencés. Une chaîne de Markov alterne des états « normal » et « baissier », avec des probabilités de transition qui rendent ces marchés baissiers persistants : y entrer est rare, y rester est probable, et les mauvaises années arrivent en grappes de trois ans environ. C'est le principe d'un stress de séquence. Bien construit, il garde la même moyenne de long terme que le modèle central : le stress mesure le risque d'ordre, pas un pessimisme caché sur le niveau. Sa variante extrême est la décennie perdue : un marché baissier de type Japon 1990, long et profond, délibérément non compensé (la moyenne est tirée vers le bas). C'est un scénario de queue à survivre, pas une espérance.

**Les forces.** La transparence et le contrôle d'abord : trois nombres explicites, pas de boîte noire. On peut y brancher les espérances prospectives ([[rendements-attendus]]) ou l'ancre CAPE ([[valorisations-et-cape]]), et tester « et si σ montait de deux points ». La généralité ensuite : le paramétrique explore des rendements mensuels jamais observés, là où le bootstrap crée de nouveaux enchaînements de mois déjà observés. L'isolation des causes enfin : la paire central/stress, de même moyenne mais aux mauvaises années regroupées (et un peu plus fortes), montre ce que coûte une séquence défavorable dans votre plan ([[sequence-des-rendements]]). Aucune autre famille ne permet si facilement cette expérience.

**Les pièges.** Ils sont le miroir des forces. Tout repose sur trois nombres que personne ne connaît, et la sensibilité aux entrées de [[monte-carlo-forces-faiblesses]] est ici maximale. La structure choisie (i.i.d., ou Markov à deux états) reste une caricature du réel : pas de retour des valorisations vers leur moyenne, pas de corrélation stochastique entre actifs (le portefeuille est agrégé avant tirage), et une inflation implicite (tout est en réel). Le paramétrique est un instrument de laboratoire : parfait pour les expériences contrôlées, à ne jamais confondre avec le monde.

## Les trois familles chez les simulateurs courants

Ce repérage change la façon d'utiliser n'importe quel outil, car un verdict ne se lit qu'en sachant quelle famille l'a produit. L'inventaire qui suit décrit les modèles mis en œuvre par les simulateurs les plus utilisés, tels que leur documentation les présente au moment où ce livre s'écrit (les outils évoluent, vérifiez la leur). C'est un inventaire, pas un classement. Un outil mono-famille excellent rend plus de services qu'un fourre-tout, et une famille absente signale un choix de conception, pas un défaut. Une transparence est due au lecteur : pofo accompagne ce livre et son auteur n'est pas neutre à son sujet, raison de plus pour croiser ses verdicts avec un outil indépendant.

| Outil | Familles | À savoir |
|---|---|---|
| cFIREsim | 1 (rejeu US depuis 1871) | Gratuit. Données Shiller actions/obligations/or, plans de dépense paramétrables. La référence de la simplicité. |
| ERN SWR Toolbox | 1 (rejeu mensuel US depuis 1871, tous les mois de départ) | Gratuit. Un classeur à copier, auditable formule par formule, flux annexes et règles CAPE de la série compris ([[serie-ern]]). Exige de mettre les mains dans le tableur. |
| FI Calc | 1 (rejeu US depuis 1871) | Gratuit. Le plus riche en règles de retrait prêtes à l'emploi (une douzaine, guardrails et VPW compris), pédagogie soignée. |
| pofo | 1, 2 et 3 (fenêtres mensuelles de vos fonds ; bootstrap de vos fonds et broad-sample du siècle des 16 pays ; Student-t calibrée, stress de séquence, décennie perdue) | L'outil développé avec ce livre. Le plus jeune de la liste, fiscalité réduite à un taux mixte global, pas de modèle à volatilité persistante. |
| Portfolio Visualizer | 2 et 3 (tirage d'années historiques, sans blocs ; normale, Student-t, GARCH, espérances prévisionnelles) | Le GARCH capture les grappes de volatilité, chose rare. Le tirage historique est année par année, donc sans mémoire. Une bonne partie des fonctions est devenue payante. |
| Rich, Broke or Dead | 1 (cycles US depuis 1871) | Gratuit. Croise chaque cycle avec les tables de mortalité : la visualisation « riche, ruiné ou mort » la plus parlante du genre. |
| TPAW Planner | 1 et 2 (séquences historiques, et tirages historiques recentrés sur les espérances du jour, 1/CAPE et taux réels) | Gratuit. L'implémentation de référence de l'amortissement (ABW), et le seul de la liste à ancrer d'office ses espérances sur les valorisations. |

Trois remarques pour se servir de ce tableau. D'abord, identifiez la famille avant de lire le verdict. Un outil purement famille 1 sur données américaines montre comment le plan aurait traversé l'histoire américaine. Ensuite, rappelez-vous que dans pofo les paramètres μ, σ et df ne pilotent que la famille 3 ; les modèles de données les ignorent, et un réglage qui « ne fait rien » n'est pas un bug. Enfin, le faisceau multi-modèles se reconstitue très bien à la main : le même plan saisi dans deux outils de familles différentes vaut un tableau de bord. Notez au passage que tous les rejoueurs cités travaillent sur l'histoire américaine ; le siècle multi-pays reste, à ce jour, la denrée rare de l'inventaire.

## La grille de lecture combinée

Reste la vraie question : que faire quand les familles sont en désaccord, ce qui est le cas normal ? Voici la grille, cas de désaccord par cas de désaccord.

**Historique et bootstrap optimistes, paramétrique central plus dur.** C'est le cas le plus courant : vos fonds ont vécu une bonne fenêtre, et le mélange vers le prior tire le central vers le bas. La lecture : vérifiez si le rendement moyen de cette période dépasse celui du modèle central. Croyez plutôt le central, dimensionnez dessus, et gardez les modèles historiques comme scénario « le monde continue comme je l'ai connu ».

**Central acceptable, stress de séquence nettement pire.** Votre plan est exposé à l'ordre des rendements : retrait initial élevé, peu de flexibilité, pas de revenus précoces. Ce n'est pas un problème de niveau d'espérance, mais de structure. Les parades sont celles de la table anti-séquence ([[sequence-des-rendements]]) : flexibilité écrite, coussin de sécurité (buffer), glidepath, revenus des premières années.

**Tout va bien sauf le broad-sample.** Votre plan tient dans le monde de vos hypothèses, mais vos flux, appliqués au 60/40 domestique du siècle développé, ne passent pas. Le broad-sample ne regarde pas votre portefeuille : il ne dit donc pas ce qui lui manque, il dit quels régimes menacent le plan. Regardez où échouent ses trajectoires. C'est presque toujours dans les blocs d'inflation persistante et les décrochages nationaux. Testez ensuite, dans les modèles qui utilisent vos poids, ce qu'y changent la diversification internationale et les actifs de régime ([[anarkulova-cederburg]], [[portefeuilles-tous-temps]]), avant d'ajouter du capital.

**Même la décennie perdue passe.** Votre plan est probablement surdimensionné. Vérifiez tout de même que ce crash-test était assez sévère, car sa profondeur croît avec la volatilité du modèle. La question n'est plus la ruine, mais le coût d'opportunité : des années de travail en trop, un capital qui mourra intact ([[une-annee-de-plus]], [[depenses-en-retraite]]).

La règle de synthèse, déjà donnée mais qui prend ici tout son sens : **planifiez sur le central, testez l'ordre avec le stress, passez le broad-sample et la décennie perdue comme des crash-tests, et gardez les modèles historiques comme repère sur les périodes observées.**

::: exemple Un désaccord instructif
Vous disposez de 15 ans de cours pour vos fonds et prévoyez une retraite de 40 ans. La case « fenêtres historiques » doit rester vide. Un bootstrap peut assembler 40 ans à partir de ces 15 ans, mais chaque bloc provient de la même courte période. Pour tester la sensibilité du plan, gardez votre portefeuille et baissez le rendement supposé dans le modèle paramétrique. Regardez alors à quel moment les dépenses essentielles cessent d'être couvertes. Comparez ce délai à celui du scénario central pour mesurer l'effet de votre hypothèse.
:::

## L'essentiel à retenir

- Trois familles : rejouer (fenêtres/cohortes), remélanger (bootstrap par blocs, dont le broad-sample sur le siècle des 16 pays), générer (paramétrique Student-t, et ses variantes à régimes pour la mémoire).
- Chacune répond à une autre question, et leurs pièges sont complémentaires : échantillon minuscule et plafond du réalisé (1), mémoire tronquée au bloc et ingrédients du passé (2), sensibilité aux entrées et structure caricaturale (3).
- Dans pofo, les paramètres μ/σ/df ne pilotent que la famille paramétrique ; les modèles de données les ignorent. Sachez toujours quelle famille vous regardez.
- Les désaccords entre familles sont le vrai résultat : biais de fenêtre, exposition à l'ordre, vulnérabilité de régime, surdimensionnement. Chaque motif de désaccord a son diagnostic et sa parade.
- Synthèse de décision : dimensionner sur le central, tester l'ordre au stress, passer le broad-sample et la décennie perdue comme des crash-tests, garder l'historique comme repère sur les périodes observées.

---

## Pour aller plus loin

- Politis & Romano, [« The Stationary Bootstrap » (1994)](https://users.ssc.wisc.edu/~bhansen/718/Politis%20Romano.pdf), sur la construction des blocs de longueur aléatoire.
- Anarkulova, Cederburg, O'Doherty & Sias (2023) : le bootstrap par blocs appliqué au siècle développé, la référence du modèle broad-sample ([[anarkulova-cederburg]]).
- Early Retirement Now, volet 8 : la méthode du rejeu mensuel systématique ([[serie-ern]]).
- Dans ce livre : [[queues-epaisses]] (le choix Student-t en détail), [[rendre-monte-carlo-pertinent]] (mélange et ancres du modèle central), [[la-machine-pofo]] (l'implémentation exacte de chacun de ces modèles dans pofo).
