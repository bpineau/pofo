# Monte-Carlo : forces, faiblesses, bon usage

Derrière chaque probabilité de ruine, chaque cône de richesse, chaque « votre plan réussit dans 94 % des cas », il y a la même machine : la simulation de Monte-Carlo. Elle génère des milliers de futurs possibles, puis compte ce qui s'y passe. C'est l'outil central de la planification moderne, celui de tous les simulateurs sérieux. Il est magnifique, à condition de savoir ce qu'il fait vraiment. Il ne prédit pas l'avenir. Il déroule les conséquences de vos hypothèses, avec une rigueur qu'aucun calcul de coin de table n'atteint.

Cette page démonte la machine complètement. D'où vient la méthode et comment elle marche, pas à pas. Ce qu'elle fait mieux que ses alternatives, le rejeu historique, les formules fermées et l'intuition. Ses quatre faiblesses structurelles et leur gravité réelle. Enfin son mode d'emploi raisonné, celui qui sépare l'utilisateur qui instruit sa décision de celui qui se fait raconter une histoire par un générateur de nombres aléatoires. Deux articles prolongent celui-ci. Les familles de modèles qui alimentent la machine ([[historique-vs-parametrique]]), et les corrections qui la rendent pertinente ([[rendre-monte-carlo-pertinent]]).

::: cle Ce qu'est vraiment une simulation
Un Monte-Carlo ne contient aucune information sur l'avenir. Il contient trois choses : vos hypothèses de marché (une distribution de rendements), votre plan (capital, retraits, règles) et un dé. Sa sortie, la probabilité de ruine, est un **théorème**. Il dit : « si le monde tire ses années dans cette distribution, alors ce plan échoue x % du temps ». Toute la valeur est dans le si. Un Monte-Carlo bien utilisé est un microscope à hypothèses. Mal utilisé, c'est une machine à blanchir des espoirs en leur donnant trois décimales.
:::

## D'où ça vient, et comment ça marche exactement

La méthode naît en 1946 à Los Alamos. Stanislaw Ulam, convalescent, joue aux réussites et se demande quelle fraction des parties est gagnable. Le calcul combinatoire exact est inextricable. L'idée lui vient de simplement jouer un grand nombre de parties, puis de compter. Avec von Neumann et Metropolis, l'idée devient méthode. Elle est baptisée du nom du casino où l'oncle d'Ulam perdait son argent, et elle résout les calculs de diffusion neutronique de la bombe. Le principe est général : quand un système est trop complexe pour une formule, on le fait tourner des milliers de fois et on regarde la distribution des issues.

Le problème du rentier est exactement de cette classe. Un plan de retrait est un système à mémoire. Le retrait de l'année 12 dépend du portefeuille de l'année 12, qui dépend de toute la séquence antérieure, des règles de dépense, du buffer, des impôts et de la pension qui démarre à l'année 15. Aucune formule fermée ne capture cela dès que le plan a un peu de réalisme. La simulation, si. Concrètement, pour une trajectoire, le moteur fait ceci ([[la-machine-pofo]]) :

1. **Tirer une séquence de rendements réels** pour tout l'horizon, dans le modèle choisi : tirages Student-t indépendants pour le modèle central, blocs d'histoire pour le bootstrap, fenêtre réelle pour les cohortes, régimes de Markov pour le stress ([[historique-vs-parametrique]]).
2. **Dérouler l'année 1** : appliquer le rendement au portefeuille, calculer le retrait selon la règle active (fixe indexé, flex, guardrails, VPW, ABW..., avec la majoration fiscale sur chaque vente), consommer ou recharger le buffer selon ses règles, encaisser pension et revenus s'ils ont commencé.
3. **Répéter** année après année jusqu'à la fin de l'horizon, ou jusqu'à l'épuisement (ruine, avec sa date).
4. **Noter tout** : ruine ou pas, date de ruine, richesse finale, dépenses réellement servies chaque année, temps passé sous l'eau.

Puis recommencer 4 000 fois (le curseur nPaths) et compter. La fraction ruinée donne la probabilité de ruine. Les richesses année par année donnent le cône ([[lire-un-fan-chart]]). Les dépenses servies donnent la section §04, et ainsi de suite. Il n'y a rien d'autre dans la boîte, juste de la comptabilité exacte appliquée à des futurs tirés au sort.

## Ce que Monte-Carlo fait mieux que tout le reste

Pour apprécier l'outil, comparons-le à ses trois concurrents.

**Contre le rejeu historique pur** (la méthode Bengen, [[etude-trinity]]). L'histoire américaine ne contient qu'une centaine d'années chevauchantes, soit 3-4 retraites de 30 ans réellement indépendantes. Le rejeu répond à « qu'aurait donné mon plan dans le passé ? ». Il ne peut, par construction, rien dire des futurs qui ne ressemblent à aucun passé. Et il traite le pire millésime historique comme une borne, alors que rien ne garantit que le pire soit déjà arrivé. Monte-Carlo génère, lui, des dizaines de milliers d'années synthétiques. Il explore l'espace des possibles autour des hypothèses, pas seulement le chemin réalisé. Les deux sont complémentaires. C'est pourquoi un simulateur les affiche côte à côte.

**Contre les formules fermées** (l'espérance de ruine analytique, les règles empiriques). Les formules exigent des hypothèses irréalistes pour rester solubles : rendements normaux, retraits proportionnels, aucune règle. Dès qu'on ajoute un plancher de guardrails, une pension différée, un buffer avec sa règle de recharge et la fiscalité sur les ventes, seule la simulation suit. C'est sa force distinctive : la complexité du plan est gratuite. Ajouter une règle réaliste ne coûte que quelques lignes de code, jamais une approximation.

**Contre l'intuition**, enfin, et c'est le plus important. L'esprit humain est catastrophiquement mauvais pour composer de l'aléatoire sur 40 ans. Personne ne « sent » correctement ce que 11 % de volatilité font à un retrait de 3,8 % sur 45 ans, avec une pension qui tombe en année 16. La simulation, si, exactement. Son rôle psychologique est d'ailleurs symétrique. Elle dégrise les optimistes, car le cône contient des trajectoires ruinées même avec de « bonnes » hypothèses. Et elle rassure les anxieux, car la médiane d'un plan raisonnable est étonnamment opulente ([[psychologie-du-retrait]]).

## Les quatre faiblesses structurelles, et leur gravité réelle

**Faiblesse 1 : garbage in, garbage out, avec effet de levier.** La sortie est hypersensible aux entrées. Un écart de ±0,5 point sur μ, indétectable statistiquement ([[rendements-attendus]]), peut faire varier la ruine du simple au double. Le simulateur amplifie même la précision apparente : trois décimales de ruine, calculées sur un μ connu à ±1 point près. Gravité : maximale, mais entièrement gérable. La discipline des entrées (calibration prospective, blending, ancres) et la lecture en intervalle y répondent ([[ruine-et-probabilites]]). C'est la raison d'être de la conception multi-modèles.

**Faiblesse 2 : l'indépendance des tirages.** Le Monte-Carlo naïf tire chaque année indépendamment de la précédente, à pile ou face, sans mémoire. Or les marchés réels ont de la mémoire. Les mauvaises années s'agglutinent, en récessions et en marchés baissiers de plusieurs années. Les valorisations créent des tendances décennales ([[valorisations-et-cape]]). La volatilité fait des grappes. La conséquence est précise. À mêmes moyenne et variance, le modèle i.i.d. sous-estime la probabilité des longues séquences médiocres, celles qui tuent les rentiers ([[sequence-des-rendements]]). Il surestime aussi un peu la dispersion à très long terme, car les vrais marchés ont un retour vers la moyenne qui resserre les cônes à 30 ans. Gravité : réelle mais quantifiable, de l'ordre de quelques points de ruine. Les correctifs existent, et un bon simulateur en implémente trois : blocs, régimes de Markov et cohortes ([[rendre-monte-carlo-pertinent]]).

**Faiblesse 3 : la forme de la distribution.** Beaucoup de simulateurs commerciaux tirent dans une loi normale. Elle assigne aux années catastrophiques (−35 % réel) des probabilités astronomiquement faibles, si bien que 2008 devient un événement « impossible ». Un Monte-Carlo gaussien est structurellement optimiste sur les queues. Gravité : sérieuse chez les autres, corrigée ici par un Student-t à df ajusté sur le kurtosis de vos fonds. Les queues épaisses entrent alors dans le modèle central ([[queues-epaisses]]).

**Faiblesse 4 : ce qui n'est pas dans le modèle n'existe pas.** Aucun tirage ne contient la confiscation, l'hyperinflation qui détruit la monnaie de référence ([[hyperinflation-et-extremes]]), la fraude du courtier, le divorce ou la dépendance à 300 €/jour. Le modèle simule le risque de marché d'un plan aux paramètres constants. Les risques de la vie, eux, restent hors champ. Gravité : structurelle et incompressible. C'est pourquoi la sortie d'un simulateur ne remplace ni les marges ([[erreurs-classiques-fire]]), ni l'assurance, ni le pilotage en route ([[quand-s-inquieter]]).

::: science Combien de trajectoires faut-il ?
L'erreur d'échantillonnage d'une probabilité p estimée sur N trajectoires vaut ≈ √(p(1−p)/N). À 4 000 trajectoires et p = 5 %, elle tombe à ±0,3 point. C'est largement assez fin face aux ±2-3 points d'incertitude des entrées. Monter à 10 000 (le curseur nPaths) lisse les courbes de sensibilité et les fractiles extrêmes des cônes, utile pour comparer finement des règles. Au-delà, on polit un chiffre dont l'incertitude vient d'ailleurs. Le vrai gain de précision d'un Monte-Carlo n'est jamais dans N. Il est dans les entrées et la structure du modèle. Méfiez-vous des outils qui vantent « 100 000 simulations » comme argument de sérieux. C'est confondre la finesse de la photo et la justesse de la mise au point.
:::

## Le bon usage : huit règles de conduite

Voici la synthèse pratique, forgée par la littérature (Kitces a beaucoup écrit sur le « bon usage du Monte-Carlo » côté conseillers) et incarnée dans une conception résolument multi-modèles.

**1. Soignez les entrées dix fois plus que la lecture des sorties.** Une heure sur le budget réel et la calibration de μ ([[combien-il-vous-faut]], [[rendements-attendus]]) vaut plus que dix heures à contempler des cônes.

**2. Ne lisez jamais un seul modèle.** L'intervalle entre les colonnes (central, stress, broad-sample, historique) est l'information. Une colonne seule n'est qu'une opinion ([[ruine-et-probabilites]]).

**3. Lisez en ordinal.** Comparer (plan A contre plan B, levier X contre levier Y) est la force de l'outil. Mesurer (« mon risque est 4,7 % ») est son point faible. Les écarts sont du signal, les décimales du bruit.

**4. Utilisez-le en machine à « et si », pas en oracle.** Sa vraie vocation. Et si je pars deux ans plus tôt ? Et si l'inflation ajoute 0,5 point aux dépenses ? Et si la pension est rabotée de 20 % ? La section §09 (le solveur en « mouvements équivalents ») industrialise cet usage. Elle convertit chaque marge en euros, en années ou en flexibilité ([[utiliser-la-page-fire]]).

**5. Regardez les trajectoires, pas seulement les agrégats.** Une probabilité résume. Les chemins d'exemple du cône (dont les rouges, échoués) montrent comment on échoue : vite par un krach précoce, ou lentement par érosion ([[lire-un-fan-chart]]). Le mode de défaillance dicte la parade, pas le taux d'échec.

**6. Simulez vos règles réelles, pas la caricature rigide.** Regardez ensuite le niveau de vie servi (§04), pas seulement la survie. Une règle flexible « réussit » toujours au sens de la ruine. La vraie question devient ce qu'elle vous fait vivre ([[flexibilite-realite]]).

**7. Re-simulez peu souvent.** Le plan se re-simule à la revue annuelle ([[revue-annuelle]]) ou sur événement, jamais chaque semaine. La sortie bouge avec les marchés, votre plan ne doit pas.

**8. Décidez sur les scénarios pessimistes, vivez sur le central.** Le plan doit être acceptable dans les colonnes dures, broad-sample et stress. Mais une fois la décision prise, c'est le scénario central qui décrit votre vie probable. Confondre les deux registres fabrique soit des imprudents, soit des anxieux perpétuels.

::: exemple Une session de bon usage, en vrai
Question : « puis-je partir en 2027 ou faut-il attendre 2029 ? ». Mauvais usage : lancer le simulateur sur 2027, obtenir 93,8 %, décréter que ça passe. Bon usage : figer les entrées auditées, puis comparer les deux plans sous les quatre modèles. Résultat type : 2027 donne 4 %/7 %/11 %/6 % de ruine (central/stress/broad/historique), 2029 donne 2 %/4 %/6 %/3 %. L'écart entre les deux dates, environ 5 points dans les colonnes dures, est le prix de sécurité de deux ans de vie active. Le solveur montre qu'un revenu d'appoint de 800 €/mois sur les 5 premières années achète à peu près le même écart. La décision (partir en 2027 avec un mi-temps doux, [[retour-au-travail]]) n'est pas dans le simulateur. Elle est dans l'arbitrage que le simulateur a rendu chiffrable, et c'est toute la différence entre instruire et se faire raconter.
:::

## L'essentiel à retenir

- Monte-Carlo = vos hypothèses + votre plan + un dé, déroulés en comptabilité exacte sur des milliers de futurs : un théorème conditionnel, jamais une prédiction.
- Ses forces : explorer au-delà du passé réalisé, digérer gratuitement la complexité des vrais plans (règles, pension, buffer, fiscalité), corriger une intuition humaine incapable de composer 40 ans d'aléatoire.
- Ses faiblesses, par gravité : la sensibilité aux entrées (gérable par discipline et multi-modèles), l'indépendance des tirages (corrigée par blocs/régimes/cohortes), les queues gaussiennes (corrigées par Student-t), le hors-champ (incompressible, marges et pilotage).
- 4 000-10 000 trajectoires suffisent : la précision d'un Monte-Carlo vient de ses entrées, jamais de son N.
- Les huit règles tiennent en une : utilisez-le pour comparer des choix sous plusieurs modèles, à entrées auditées. Jamais pour vous faire dire un chiffre rassurant à trois décimales.

---

## Pour aller plus loin

- Michael Kitces, « The Problem With Monte Carlo Analyses In Retirement Projections » et les articles associés ([kitces.com](https://www.kitces.com)) : le bon usage vu du métier de conseiller.
- Early Retirement Now, volet 8 (l'appendice technique) et volet 46 (la fausse précision) ([[serie-ern]]).
- Metropolis & Ulam, « The Monte Carlo Method » (1949) : l'article fondateur, étonnamment lisible.
- La suite directe dans ce livre : [[historique-vs-parametrique]] (les trois familles de sources de rendements), [[queues-epaisses]] (le choix Student-t), [[rendre-monte-carlo-pertinent]] (blending, régimes, stress) et [[lire-un-fan-chart]] (la lecture des sorties).
