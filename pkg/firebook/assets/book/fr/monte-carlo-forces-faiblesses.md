# Monte-Carlo : forces, faiblesses, bon usage

Derrière chaque probabilité de ruine, chaque cône de richesse, chaque « votre plan réussit dans 94 % des cas », il y a la même machine : la simulation de Monte-Carlo, qui consiste à générer des milliers de futurs possibles et à compter ce qui s'y passe. C'est l'outil central de toute la planification moderne, celui de pofo comme de tous les simulateurs sérieux, et c'est un outil magnifique À **condition** de savoir ce qu'il fait vraiment : il ne prédit pas l'avenir, il déroule les conséquences de **vos** hypothèses avec une rigueur qu'aucun raisonnement de coin de table n'atteint.

Cette page démonte la machine complètement : d'où vient la méthode et comment elle marche pas à pas, ce qu'elle fait mieux que toutes les alternatives (le rejeu historique, les formules fermées, l'intuition), ses quatre faiblesses structurelles et leur gravité réelle, et le mode d'emploi raisonné, celui qui distingue l'utilisateur qui **instruit** sa décision de celui qui se fait raconter une histoire par un générateur de nombres aléatoires. Les deux articles suivants prolongent : les familles de modèles qui alimentent la machine ([[historique-vs-parametrique]]) et les corrections qui la rendent pertinente ([[rendre-monte-carlo-pertinent]]).

::: cle Ce qu'est vraiment une simulation
Un Monte-Carlo ne contient **aucune** information sur l'avenir. Il contient trois choses : vos hypothèses de marché (une distribution de rendements), votre plan (capital, retraits, règles), et un dé. Sa sortie, la probabilité de ruine, est un **théorème** : « **si** le monde tire ses années dans cette distribution, **alors** ce plan échoue x % du temps ». Toute la valeur est dans le **si**. Un Monte-Carlo bien utilisé est un microscope à hypothèses ; mal utilisé, c'est une machine à blanchir des espoirs en leur donnant trois décimales.
:::

## D'où ça vient, et comment ça marche exactement

La méthode naît en 1946 à Los Alamos : Stanislaw Ulam, convalescent, joue aux réussites et se demande quelle fraction des parties est gagnable. Le calcul combinatoire exact est inextricable ; l'idée lui vient de simplement **jouer** un grand nombre de parties et de compter. Avec von Neumann et Metropolis, l'idée devient méthode (baptisée du nom du casino où l'oncle d'Ulam perdait son argent) et résout les calculs de diffusion neutronique de la bombe : quand un système est trop complexe pour être résolu par une formule, on le fait tourner des milliers de fois et on regarde la distribution des issues.

Le problème du rentier est **exactement** de cette classe. Un plan de retrait est un système à mémoire : le retrait de l'année 12 dépend du portefeuille de l'année 12, qui dépend de toute la séquence antérieure, des règles de dépense, du buffer, des impôts, de la pension qui démarre en l'année 15... Aucune formule fermée ne capture ça dès que le plan a un peu de réalisme. La simulation, si. Concrètement, pour **une** trajectoire, le moteur de pofo fait ceci ([[la-machine-pofo]]) :

1. **Tirer une séquence de rendements réels** pour tout l'horizon, dans le modèle choisi : tirages Student-t indépendants pour le modèle central, blocs d'histoire pour le bootstrap, fenêtre réelle pour les cohortes, régimes de Markov pour le stress ([[historique-vs-parametrique]]).
2. **Dérouler l'année 1** : appliquer le rendement au portefeuille, calculer le retrait selon la règle active (fixe indexé, flex, guardrails, VPW, ABW..., avec la majoration fiscale sur chaque vente), consommer ou recharger le buffer selon ses règles, encaisser pension et revenus s'ils ont commencé.
3. **Répéter** année après année jusqu'à la fin de l'horizon, ou jusqu'à l'épuisement (ruine, avec sa date).
4. **Noter tout** : ruine ou pas, date de ruine, richesse finale, dépenses réellement servies chaque année, temps passé sous l'eau.

Puis recommencer 4 000 fois (le curseur nPaths), et compter : la fraction ruinée donne la probabilité de ruine, les richesses année par année donnent le cône ([[lire-un-fan-chart]]), les dépenses servies donnent la section §04, etc. Il n'y a rien d'autre dans la boîte : de la comptabilité exacte appliquée à des futurs tirés au sort.

## Ce que Monte-Carlo fait mieux que tout le reste

Pour apprécier l'outil, comparons-le à ses trois concurrents.

**Contre le rejeu historique pur** (la méthode Bengen, [[etude-trinity]]) : l'histoire américaine ne contient qu'une centaine d'années chevauchantes, soit 3-4 retraites de 30 ans réellement indépendantes. Le rejeu répond à « qu'aurait donné mon plan dans le passé ? » ; il ne peut par construction rien dire des futurs qui ne ressemblent à aucun passé, et il traite le pire millésime historique comme une borne alors que rien ne garantit que le pire soit déjà arrivé. Monte-Carlo génère des dizaines de milliers d'années synthétiques : il explore l'**espace** des possibles autour des hypothèses, pas seulement le chemin réalisé. Les deux sont complémentaires, et c'est pourquoi pofo affiche les deux côte à côte.

**Contre les formules fermées** (l'espérance de ruine analytique, les règles de pouce) : les formules exigent des hypothèses irréalistes (rendements normaux, retraits proportionnels, pas de règles) pour rester solubles. Dès qu'on ajoute un plancher de guardrails, une pension différée, un buffer avec règle de recharge et la fiscalité sur les ventes, seule la simulation suit. C'est **sa** force distinctive : la complexité du plan est gratuite. Ajouter une règle réaliste ne coûte que quelques lignes de code, jamais une approximation.

**Contre l'intuition**, enfin, et c'est le plus important : l'esprit humain est catastrophiquement mauvais pour composer de l'aléatoire sur 40 ans. Personne ne « sent » correctement ce que 11 % de volatilité font à un retrait de 3,8 % sur 45 ans avec une pension en année 16. La simulation, si, exactement. Son rôle psychologique est d'ailleurs symétrique : elle dégrise les optimistes (le cône contient des trajectoires ruinées même avec de « bonnes » hypothèses) et rassure les anxieux (la médiane d'un plan raisonnable est étonnamment opulente, [[psychologie-du-retrait]]).

## Les quatre faiblesses structurelles, et leur gravité réelle

**Faiblesse 1 : garbage in, garbage out, avec effet de levier.** La sortie est hypersensible aux entrées : ±0,5 point sur μ, indétectable statistiquement ([[rendements-attendus]]), peut faire varier la ruine du simple au double. Le simulateur **amplifie** la précision apparente : trois décimales de ruine calculées sur un μ connu à ±1 point. Gravité : maximale, mais entièrement gérable par la discipline des entrées (calibration prospective, blending, ancres) et la lecture en intervalle ([[ruine-et-probabilites]]). C'est **la** raison d'être de la conception multi-modèles de pofo.

**Faiblesse 2 : l'indépendance des tirages.** Le Monte-Carlo naïf tire chaque année indépendamment de la précédente : pile ou face, sans mémoire. Or les marchés réels ont de la mémoire : les mauvaises années s'agglutinent (récessions, ours de plusieurs années), les valorisations créent des tendances décennales ([[valorisations-et-cape]]), la volatilité fait des grappes. Conséquence précise : à mêmes moyenne et variance, le modèle i.i.d. sous-estime la probabilité des **longues** séquences médiocres, celles qui tuent les rentiers ([[sequence-des-rendements]]), et surestime légèrement la dispersion à très long terme (les vrais marchés ont un retour vers la moyenne qui resserre les cônes à 30 ans). Gravité : réelle mais quantifiable, de l'ordre de quelques points de ruine ; les correctifs existent et pofo en implémente trois (blocs, régimes de Markov, cohortes, [[rendre-monte-carlo-pertinent]]).

**Faiblesse 3 : la forme de la distribution.** Beaucoup de simulateurs commerciaux tirent dans une loi **normale**, qui assigne aux années catastrophiques (−35 % réel) des probabilités astronomiquement faibles : 2008 devient un événement « impossible ». Un Monte-Carlo gaussien est structurellement optimiste sur les queues. Gravité : sérieuse chez les autres, traitée dans pofo (Student-t à df ajusté sur le kurtosis de vos fonds : les queues épaisses sont dans le modèle central, [[queues-epaisses]]).

**Faiblesse 4 : ce qui n'est pas dans le modèle n'existe pas.** Aucun tirage ne contient la confiscation, l'hyperinflation qui détruit la monnaie de référence ([[hyperinflation-et-extremes]]), la fraude du courtier, le divorce, la dépendance à 300 €/jour. Le modèle simule le risque de **marché** d'un plan aux paramètres constants ; les risques de la vie restent hors champ. Gravité : structurelle et incompressible ; c'est pourquoi la sortie d'un simulateur ne remplace ni les marges ([[erreurs-classiques-fire]]), ni l'assurance, ni le pilotage en route ([[quand-s-inquieter]]).

::: science Combien de trajectoires faut-il ?
L'erreur d'échantillonnage d'une probabilité p estimée sur N trajectoires vaut ≈ √(p(1−p)/N). À 4 000 trajectoires et p = 5 % : ±0,3 point, largement assez fin comparé aux ±2-3 points d'incertitude des **entrées**. Monter à 10 000 (le curseur nPaths) lisse les courbes de sensibilité et les fractiles extrêmes des cônes, utile pour les comparaisons fines de règles ; au-delà, on polit un chiffre dont l'incertitude vient d'ailleurs. Le vrai gain de précision d'un Monte-Carlo n'est **jamais** dans N : il est dans les entrées et la structure du modèle. Méfiez-vous des outils qui vantent « 100 000 simulations » comme argument de sérieux : c'est confondre la finesse de la photo et la justesse de la mise au point.
:::

## Le bon usage : huit règles de conduite

Voici la synthèse pratique, forgée par la littérature (Kitces a beaucoup écrit sur le « bon usage du Monte-Carlo » côté conseillers) et incarnée dans la conception de pofo.

**1. Soignez les entrées dix fois plus que la lecture des sorties.** Une heure sur le budget réel et la calibration de μ ([[combien-il-vous-faut]], [[rendements-attendus]]) vaut plus que dix heures à contempler des cônes.

**2. Ne lisez jamais un seul modèle.** L'intervalle entre les colonnes de pofo (central, stress, broad-sample, historique) est l'information ; une colonne seule est une opinion ([[ruine-et-probabilites]]).

**3. Lisez en ordinal.** Comparer (plan A vs plan B, levier X vs levier Y) est la force de l'outil ; mesurer (« mon risque est 4,7 % ») est son point faible. Les écarts sont du signal, les décimales du bruit.

**4. Utilisez-le en machine à « et si », pas en oracle.** Sa vraie vocation : et si je pars deux ans plus tôt ? et si l'inflation ajoute 0,5 point aux dépenses ? et si la pension est rabotée de 20 % ? La section §09 de pofo (le solveur en « mouvements équivalents ») industrialise cet usage : elle convertit chaque marge en euros, en années ou en flexibilité ([[utiliser-la-page-fire]]).

**5. Regardez les trajectoires, pas seulement les agrégats.** Une probabilité résume ; les chemins d'exemple du cône (dont les rouges, échoués) montrent **comment** on échoue : vite par krach précoce, ou lentement par érosion ([[lire-un-fan-chart]]). Le mode de défaillance dicte la parade, pas le taux d'échec.

**6. Simulez vos règles réelles, pas la caricature rigide,** puis regardez le niveau de vie servi (§04), pas seulement la survie : une règle flexible « réussit » toujours au sens de la ruine, la question devient ce qu'elle vous fait vivre ([[flexibilite-realite]]).

**7. Re-simulez peu souvent.** Le plan se re-simule à la revue annuelle ([[revue-annuelle]]) ou sur événement, pas chaque semaine : la sortie bouge avec les marchés, votre plan ne doit pas.

**8. Décidez sur les scénarios pessimistes, vivez sur le central.** Le plan doit être acceptable dans les colonnes dures (broad-sample, stress) ; mais une fois la décision prise, c'est le scénario central qui décrit votre vie probable. Confondre les deux registres fabrique soit des imprudents, soit des anxieux perpétuels.

::: exemple Une session de bon usage, en vrai
Question : « puis-je partir en 2027 ou faut-il attendre 2029 ? » Mauvais usage : lancer le simulateur sur 2027, obtenir 93,8 %, décréter que ça passe. Bon usage : figer les entrées auditées, puis comparer les **deux** plans sous les **quatre** modèles. Résultat type : 2027 donne 4 %/7 %/11 %/6 % de ruine (central/stress/broad/historique), 2029 donne 2 %/4 %/6 %/3 %. Lecture : l'écart entre les deux dates (~5 points dans les colonnes dures) est le prix de sécurité de deux ans de vie active ; le solveur montre qu'un revenu d'appoint de 800 €/mois sur les 5 premières années achète à peu près le même écart. La décision (partir en 2027 avec un mi-temps doux, [[retour-au-travail]]) n'est pas dans le simulateur : elle est dans l'arbitrage que le simulateur a rendu **chiffrable**. C'est toute la différence entre instruire et se faire raconter.
:::

## L'essentiel à retenir

- Monte-Carlo = vos hypothèses + votre plan + un dé, déroulés en comptabilité exacte sur des milliers de futurs : un théorème conditionnel, jamais une prédiction.
- Ses forces : explorer au-delà du passé réalisé, digérer gratuitement la complexité des vrais plans (règles, pension, buffer, fiscalité), corriger une intuition humaine incapable de composer 40 ans d'aléatoire.
- Ses faiblesses, par gravité : la sensibilité aux entrées (gérable par discipline et multi-modèles), l'indépendance des tirages (corrigée par blocs/régimes/cohortes), les queues gaussiennes (corrigée par Student-t), le hors-champ (incompressible : marges et pilotage).
- 4 000-10 000 trajectoires suffisent : la précision d'un Monte-Carlo vient de ses entrées, jamais de son N.
- Les huit règles tiennent en une : utilisez-le pour **comparer** des choix sous plusieurs modèles, à entrées auditées ; jamais pour vous faire dire un chiffre rassurant à trois décimales.

---

## Pour aller plus loin

- Michael Kitces, « The Problem With Monte Carlo Analyses In Retirement Projections » et les articles associés (kitces.com) : le bon usage vu du métier de conseiller.
- Early Retirement Now, Part 8 (l'appendice technique) et Part 46 (la fausse précision) ([[serie-ern]]).
- Metropolis & Ulam, « The Monte Carlo Method » (1949) : l'article fondateur, étonnamment lisible.
- La suite directe dans ce livre : [[historique-vs-parametrique]] (les trois familles de sources de rendements), [[queues-epaisses]] (le choix Student-t), [[rendre-monte-carlo-pertinent]] (blending, régimes, stress) et [[lire-un-fan-chart]] (la lecture des sorties).
