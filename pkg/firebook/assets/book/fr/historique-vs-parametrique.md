# Fenêtres historiques, bootstrap, paramétrique : trois familles de modèles

Un simulateur de retraite déroule des plans dans des futurs générés ([[monte-carlo-forces-faiblesses]]). Toute la question est de savoir d'où viennent ces futurs. Il n'existe que trois grandes réponses, trois familles de modèles. On peut rejouer l'histoire telle quelle (les fenêtres historiques). On peut la rééchantillonner en la remélangeant (le bootstrap). On peut enfin générer des rendements synthétiques à partir de quelques paramètres (les modèles paramétriques).

Chaque famille répond en réalité à une question différente. Chacune a ses vertus et ses angles morts. Et quand leurs verdicts divergent sur un même plan, ce n'est pas une anomalie : c'est l'information la plus précieuse que vous obtiendrez. Cette page fait le tour complet des trois familles. Pour chacune, on verra la mécanique exacte, sa traduction dans la page FIRE (les six colonnes du tableau s'y rattachent toutes), ses forces et ses pièges. À la fin vient la grille de lecture combinée, qui dit quel modèle croire pour quelle question.

::: cle Trois familles, trois questions
Fenêtres historiques : « qu'aurait donné mon plan si j'étais parti à chaque date du passé ? » Bootstrap : « et dans des histoires **plausibles** faites des mêmes ingrédients que le passé, mais assemblés autrement ? » Paramétrique : « et dans un monde dont je choisis explicitement la moyenne, la volatilité et les queues ? » Aucune ne répond à « que va-t-il se passer ? » ; ensemble, elles **encadrent** cette question inaccessible. C'est le principe de conception de la page FIRE : les six colonnes du tableau principal sont ces trois familles déclinées ([[utiliser-la-page-fire]]).
:::

## Famille 1 : les fenêtres historiques (rejeu, cohortes)

**La mécanique.** C'est la méthode fondatrice de Bengen ([[etude-trinity]]). On prend la série réelle des rendements de votre portefeuille (ou d'un indice). Puis on déroule le plan à partir de chaque date de départ possible : la fenêtre janvier 1975 → décembre 2019, puis février 1975 → janvier 2020, et ainsi de suite. Chaque fenêtre est un « millésime », une cohorte. Le taux d'échec est la fraction des millésimes ruinés.

**Dans la page FIRE** : la colonne « Historical windows », disponible en mode portefeuille. Le moteur reconstruit d'abord l'historique réel long de vos lignes (extensions `SIM`, déflaté en euros constants). Il échantillonne ensuite les fenêtres au pas mensuel. Chaque mois de départ possible donne une cohorte, ce qui multiplie les fenêtres par douze par rapport à l'échantillonnage annuel et préserve les enchaînements intra-annuels ([[la-machine-pofo]]). Quand l'historique est plus court que l'horizon (20 ans de données pour 45 ans de plan), le modèle refuse honnêtement de répondre plutôt que d'extrapoler. C'est le message « not enough history », et c'est un parti pris, pas une faiblesse. La section §02 (« The retirements that actually happened ») applique la même famille en version dirigée : votre plan rejoué aux dates de départ célèbres (USA 1929, 1966, 2000, Japon 1990) sur les données longues embarquées.

**Les forces.** La fidélité absolue au réel. Les corrélations entre actifs, les grappes de crises, les enchaînements krach-inflation-reprise : tout y est, puisque rien n'est modélisé, tout est cité. C'est le seul modèle dont chaque trajectoire s'est réellement produite. D'où une vertu pédagogique inégalée (voir son plan traverser 1966 parle plus que mille probabilités) et un excellent détecteur de fragilité face aux régimes réels.

**Les pièges.** Ils sont trois, et sérieux. D'abord l'échantillon minuscule. Cent ans de données ne contiennent que trois ou quatre retraites longues indépendantes. Les fenêtres se chevauchent massivement : le krach de 2008 apparaît dans 350 fenêtres mensuelles. Le « taux d'échec » porte donc des barres d'erreur énormes, que son affichage ne montre pas. Ensuite le plafond du réalisé. Le pire du passé y est traité comme le pire absolu, alors que rien ne garantit que le pire soit derrière nous. Enfin le biais de la fenêtre disponible. L'historique de vos fonds couvre le plus souvent les décennies récentes, plutôt favorables. C'est la borne optimiste de la page, et la page FIRE l'assortit d'un avertissement explicite quand l'historique est plus court que l'horizon.

## Famille 2 : le bootstrap (rééchantillonnage par blocs)

**La mécanique.** Le bootstrap répond au problème de l'échantillon minuscule. Plutôt que de rejouer l'histoire dans l'ordre, on la découpe et on tire au sort, avec remise, des morceaux qu'on recolle en histoires synthétiques. Tirer mois par mois détruirait les grappes et les tendances, car on retomberait sur de l'i.i.d. On tire donc des blocs de plusieurs années. La variante de référence est le bootstrap stationnaire (Politis-Romano, 1994) : des blocs de longueur aléatoire (moyenne fixée à 24 mois dans la page FIRE), ce qui évite les artefacts de coupe des blocs de taille fixe. Chaque trajectoire simulée est alors une histoire qui n'a jamais eu lieu, mais dont chaque morceau de deux ans, lui, a réellement eu lieu, avec ses corrélations internes et une bonne partie de sa mémoire.

**Dans la page FIRE** : deux colonnes relèvent de cette famille. « Block bootstrap » (mode portefeuille) rééchantillonne le panel mensuel de vos lignes sous vos poids courants. Et surtout le modèle « Broad sample » applique le bootstrap par blocs à un tout autre panel : le siècle des 16 pays développés (Jorda-Schularick-Taylor, 1870-2020). Le tirage se fait par pays entiers, pour que les grands désastres (1929, les années 1970, le Japon) survivent intacts dans les trajectoires, sur un portefeuille 60/40 domestique conforme à la littérature ([[anarkulova-cederburg]]). Même famille mathématique, deux questions différentes : « mes fonds, remélangés » contre « le siècle développé, remélangé ».

**Les forces.** Le meilleur compromis entre fidélité et diversité. On garde les corrélations et la mémoire courte du réel (via les blocs), et on obtient des milliers de trajectoires distinctes (via le remélange). Le broad-sample y ajoute la profondeur, avec des régimes que l'historique de vos fonds n'a jamais vus. C'est la famille que la recherche moderne préfère pour estimer le risque de long horizon, et c'est celle d'Anarkulova-Cederburg.

**Les pièges.** La mémoire au-delà du bloc est détruite. Un marché baissier de sept ans ne peut pas naître d'un tirage de blocs de deux ans, sauf par malchance de tirages sombres consécutifs. Le retour de valorisation étalé sur plusieurs décennies ([[valorisations-et-cape]]) disparaît lui aussi. Les ingrédients, ensuite, restent ceux du passé disponible : le bootstrap remélange, il n'invente rien. Appliqué au seul historique de vos fonds, il hérite du biais de fenêtre de la famille 1, d'où la mise en garde commune aux deux colonnes du haut. Enfin, le choix de la longueur de bloc est un vrai paramètre. Trop court, on tue les grappes ; trop long, on retombe dans le rejeu et son échantillon pauvre. Les 24 mois moyens suivent la pratique de la littérature : assez long pour contenir une récession type, assez court pour diversifier.

## Famille 3 : le paramétrique (Student-t, et les régimes)

**La mécanique.** On abandonne les données brutes. On décrit le monde par quelques paramètres, et on tire dedans. La version la plus simple est l'i.i.d. gaussien (moyenne, volatilité, tirages annuels indépendants), celle de la plupart des simulateurs commerciaux. Elle a deux défauts corrigibles : des queues trop fines et aucune mémoire. Le modèle central corrige le premier en tirant dans une Student-t à trois paramètres : μ (la moyenne), σ (la volatilité) et df (l'épaisseur des queues). À df 5, l'année à −30 % réel est environ dix fois plus probable qu'en loi normale ([[queues-epaisses]]). En mode portefeuille, les tirages sont mensuels puis composés, et les trois paramètres sont ajustés sur vos fonds avant d'être mélangés vers un prior mondial prudent ([[rendre-monte-carlo-pertinent]]).

Le second défaut, l'absence de mémoire, donne naissance à la sous-famille des modèles à régimes, paramétriques mais séquencés. Une chaîne de Markov alterne des états « normal » et « baissier », avec des probabilités de transition qui rendent ces marchés baissiers persistants : y entrer est rare, y rester est probable, et les mauvaises années arrivent en grappes de trois ans environ. C'est la colonne « Sequence stress ». Elle garde, par construction, la même moyenne de long terme que le central : le stress mesure le risque d'ordre, pas un pessimisme caché sur le niveau. Seule la volatilité est concentrée en épisodes. Sa variante extrême est « Lost decade » : un marché baissier de type Japon 1990, long et profond, délibérément non compensé (la moyenne est tirée vers le bas). C'est un scénario de queue à survivre, pas une espérance ([[utiliser-la-page-fire]]).

**Les forces.** La transparence et le contrôle d'abord : trois curseurs, pas de boîte noire. On peut y brancher les espérances prospectives ([[rendements-attendus]]) ou l'ancre CAPE ([[valorisations-et-cape]]), et tester « et si σ montait de deux points ». La généralité ensuite : le paramétrique explore des mondes que l'histoire n'a pas produits, ce que ni le rejeu ni le bootstrap ne savent faire. L'isolation des causes enfin : la paire central/stress, identique en tout sauf l'ordre des années, mesure le prix de la séquence dans votre plan ([[sequence-des-rendements]]). Aucune autre famille ne permet cette expérience contrôlée.

**Les pièges.** Ils sont le miroir des forces. Tout repose sur trois nombres que personne ne connaît, et la sensibilité aux entrées de [[monte-carlo-forces-faiblesses]] est ici maximale. La structure choisie (i.i.d., ou Markov à deux états) reste une caricature du réel : pas de retour de valorisation, pas de corrélation stochastique entre actifs (le portefeuille est agrégé avant tirage), et une inflation implicite (tout est en réel). Le paramétrique est un instrument de laboratoire : parfait pour les expériences contrôlées, à ne jamais confondre avec le monde.

::: science Le tableau de correspondance complet
Les six colonnes de la page FIRE, rattachées à leur famille : Historical windows = famille 1 sur vos fonds (mensuel) ; Block bootstrap = famille 2 sur vos fonds (blocs ~24 mois) ; Student-t = famille 3 i.i.d. calibrée puis mélangée (le central, celui des sections de détail par défaut) ; Sequence stress = famille 3 à régimes de Markov, moyenne préservée ; Broad sample = famille 2 sur le siècle des 16 pays (blocs par pays, 60/40) ; Lost decade = famille 3 à régime forcé, moyenne dégradée. Les curseurs μ/σ/df ne pilotent que la famille 3 ; les familles 1-2 les ignorent totalement. C'est la première chose à savoir pour comprendre pourquoi une case cochée « ne fait rien » sur certaines colonnes ([[la-machine-pofo]]).
:::

## La grille de lecture combinée

Reste la vraie question : que faire quand les familles sont en désaccord, ce qui est le cas normal ? Voici la grille, cas de désaccord par cas de désaccord.

**Historique et bootstrap optimistes, paramétrique central plus dur.** C'est le cas le plus courant : vos fonds ont vécu une bonne fenêtre, et le mélange vers le prior tire le central vers le bas. La lecture : l'écart mesure le biais de votre fenêtre historique. Croyez plutôt le central, dimensionnez dessus, et gardez les colonnes historiques comme scénario « le monde continue comme je l'ai connu ».

**Central acceptable, stress de séquence nettement pire.** Votre plan est exposé à l'ordre des rendements : retrait initial élevé, peu de flexibilité, pas de revenus précoces. Ce n'est pas un problème de niveau d'espérance, mais de structure. Les parades sont celles de la table anti-séquence ([[sequence-des-rendements]]) : flexibilité écrite, coussin de sécurité (buffer), glidepath, revenus des premières années.

**Tout va bien sauf le broad-sample.** Votre plan tient dans le monde de vos hypothèses, mais pas dans le siècle développé complet. La lecture : regardez où échouent les trajectoires broad-sample. C'est presque toujours dans les blocs d'inflation persistante et les décrochages nationaux. Les réponses sont la diversification internationale et les actifs de régime ([[anarkulova-cederburg]], [[portefeuilles-tous-temps]]), pas forcément plus de capital.

**Même la décennie perdue passe.** Votre plan est surdimensionné. La question n'est plus la ruine, mais le coût d'opportunité : des années de travail en trop, un capital qui mourra intact ([[une-annee-de-plus]], [[depenses-en-retraite]]).

La règle de synthèse, déjà donnée mais qui prend ici tout son sens : **planifiez entre le central et le broad-sample, testez l'ordre avec le stress, éprouvez le plan avec la décennie perdue, et gardez les colonnes historiques comme borne optimiste et comme pédagogie.** Quatre familles de futurs, une seule décision.

::: exemple Un désaccord instructif
Portefeuille réel de 15 ans d'historique (belle fenêtre 2010-2025), plan à 3,9 %. Les verdicts : fenêtres 0 %, bootstrap 2 %, central 6 %, stress 10 %, broad-sample 13 %. Un lecteur naïf choisit la colonne qui lui plaît. La grille de lecture, elle, raisonne autrement. La fenêtre 2010-2025 ne contient ni inflation longue ni décennie perdue : l'écart entre les familles 1-2 et la famille 3 trahit un biais de fenêtre. Le plan est en outre sensible à l'ordre (6 → 10, retrait un peu haut, zéro flexibilité). Et le broad-sample confirme la vulnérabilité à l'inflation. La décision qui en découle : retrait ramené à 3,6 %, règle de flexibilité écrite (coupe de 10 % au-delà de 4,5 % de taux courant), et 10 % du portefeuille basculés vers les linkers et l'or ([[obligations-indexees]], [[or-en-retrait]]). Aucune colonne seule n'aurait produit ce diagnostic en trois points. C'est le désaccord qui l'a produit.
:::

## L'essentiel à retenir

- Trois familles : rejouer (fenêtres/cohortes), remélanger (bootstrap par blocs, dont le broad-sample sur le siècle des 16 pays), générer (paramétrique Student-t, et ses variantes à régimes pour la mémoire).
- Chacune répond à une autre question, et leurs pièges sont complémentaires : échantillon minuscule et plafond du réalisé (1), mémoire tronquée au bloc et ingrédients du passé (2), sensibilité aux entrées et structure caricaturale (3).
- Les curseurs μ/σ/df ne pilotent que la famille paramétrique ; les modèles de données les ignorent. Sachez toujours quelle famille vous regardez.
- Les désaccords entre colonnes sont le vrai résultat : biais de fenêtre, exposition à l'ordre, vulnérabilité de régime, surdimensionnement. Chaque motif de désaccord a son diagnostic et sa parade.
- Synthèse de décision : dimensionner entre central et broad-sample, tester l'ordre au stress, éprouver le plan à la décennie perdue, garder l'historique comme borne optimiste et comme leçon de choses.

---

## Pour aller plus loin

- Politis & Romano, « The Stationary Bootstrap » (1994) : la méthode des blocs de longueur aléatoire que le simulateur implémente.
- Anarkulova, Cederburg & O'Doherty (2023) : le bootstrap par blocs appliqué au siècle développé, la référence de la colonne broad-sample ([[anarkulova-cederburg]]).
- Early Retirement Now, volet 8 : la méthode du rejeu mensuel systématique ([[serie-ern]]).
- Dans ce livre : [[queues-epaisses]] (le choix Student-t en détail), [[rendre-monte-carlo-pertinent]] (mélange et ancres du modèle central), [[la-machine-pofo]] (l'implémentation exacte, colonne par colonne).
