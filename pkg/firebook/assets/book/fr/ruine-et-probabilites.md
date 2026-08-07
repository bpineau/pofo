# La probabilité de ruine : la lire, la choisir, ne pas la subir

Tous les simulateurs de retraite, pofo compris, résument leur verdict en un chiffre : la probabilité de ruine (ou son complément, le « taux de succès »). C'est le chiffre le plus regardé du sujet, et le plus mal compris : on le lit comme une météo (« 5 % de risque, ça va »), on le compare entre outils qui ne mesurent pas la même chose, on lui demande une précision qu'il n'a pas, et on oublie qu'il décrit un monde où personne ne réagit jamais.

Cette page apprend à lire ce chiffre en professionnel : ce qu'il mesure exactement, comment choisir son seuil acceptable, pourquoi 2 % et 8 % sont souvent indiscernables, et ce que la ruine simulée a de très différent d'une ruine réelle.

::: cle La définition, sans raccourci
La probabilité de ruine d'un plan est la fraction des futurs simulés (ou des fenêtres historiques rejouées) dans lesquels le portefeuille atteint zéro **avant** la fin de l'horizon, sous une règle de retrait appliquée mécaniquement, sans aucune réaction humaine. C'est une propriété du **couple** plan + modèle, jamais du plan seul : changer de modèle change le chiffre, parfois du simple au triple, sans que votre plan ait bougé ([[historique-vs-parametrique]]).
:::

## Ce que le chiffre mesure, et ne mesure pas

Décomposons la définition, terme à terme, parce que chaque terme cache un piège.

**« La fraction des futurs simulés ».** Le chiffre est une fréquence dans une population de scénarios générés par un modèle. Si le modèle tire les années indépendamment, il sous-estime les grappes de mauvaises années ([[sequence-des-rendements]]) ; s'il rejoue l'histoire américaine, il hérite de son biais optimiste ([[etude-trinity]], [[pieges-des-simulateurs]]) ; s'il rejoue l'échantillon mondial, il inclut des pays et des époques peut-être plus durs que votre futur plausible ([[anarkulova-cederburg]]). Aucun n'est « le vrai » : c'est pourquoi la page FIRE de pofo affiche le même plan sous quatre modèles côte à côte, et pourquoi la bonne lecture est l'**intervalle** qu'ils dessinent, pas une colonne.

**« Le portefeuille atteint zéro ».** La ruine du simulateur est binaire et terminale. Elle ne distingue pas l'échec à 71 ans de l'échec à 94 ans, ni le scénario qui finit à zéro de celui qui finit à 5 000 € (échec) ou 15 000 € (succès !). Deux plans à 5 % de ruine peuvent cacher des réalités très différentes : l'un échoue tôt et brutalement, l'autre s'essouffle en toute fin de parcours avec la pension légale en soutien. D'où l'intérêt des vues complémentaires : **quand** surviennent les échecs, et quelle richesse médiane en fin d'horizon ([[lire-un-fan-chart]]).

**« Sans aucune réaction humaine ».** L'hypothèse la plus irréaliste et la plus utile. Irréaliste : aucun humain ne maintient 40 000 € de retraits indexés pendant que son portefeuille passe de 1 M€ à 150 k€ ; il aurait coupé ses dépenses des années plus tôt ([[quand-s-inquieter]]). Utile : c'est justement parce que la règle simulée est aveugle que le chiffre mesure la robustesse **intrinsèque** du plan, sans se payer de mots sur une flexibilité future hypothétique. Un plan flexible se simule avec sa règle flexible (pofo le fait, [[plancher-plafond]], [[guardrails-morningstar]]) ; mais alors la « ruine » chute et c'est le niveau de vie délivré qu'il faut regarder, car la flexibilité ne supprime pas la douleur, elle la déplace vers des années de dépenses réduites.

**Et ce qu'il ignore superbement** : votre mortalité (une « ruine » à 97 ans concerne peu de monde ; pofo a une vue croisée mortalité × ruine), les à-coups de dépenses réels, la fiscalité fine, et tous les filets hors modèle (famille, patrimoine immobilier, retour au travail).

## Choisir son seuil : pourquoi il n'y a pas de bonne réponse universelle

Quelle ruine accepter : 1 %, 5 %, 10 % ? La question semble technique ; elle est en réalité personnelle, et dépend de trois choses.

**La qualité de vos filets.** La ruine simulée suppose zéro recours. Un quadragénaire employable, propriétaire, avec pension légale à venir et famille solidaire peut rationnellement accepter 10 à 15 % de ruine **simulée**, parce que sa ruine **réelle** (fin de vie dans le dénuement sans aucun recours) est bien plus rare que le chiffre. Une personne de 60 ans sans pension notable, sans immobilier et sans possibilité de retravailler doit lire le chiffre presque littéralement, et viser bas.

**Le coût d'une année de marge.** Passer de 5 % à 2 % de ruine coûte typiquement 10 à 20 % de capital en plus, soit deux à quatre ans de travail. Passer de 5 % à 10 % les rend. Le seuil est un prix d'arbitrage entre deux risques ([[une-annee-de-plus]]) : chiffrez ce que chaque point de ruine vous coûte ou vous rend en années de vie active, la discussion devient concrète.

**La nature de l'échec dans VOTRE plan.** Regardez quand échouent les scénarios qui échouent. Des échecs tardifs (après 85 ans), adossés à une pension qui couvre le plancher : un 8 % de ruine ainsi composé est plus confortable qu'un 4 % fait d'effondrements à 70 ans.

::: astuce Le réflexe des praticiens
Les planificateurs financiers sérieux (Kitces en tête) convergent vers une fourchette de travail de 5 à 20 % de ruine simulée pour des plans avec filets et règle d'ajustement, et rappellent qu'un taux de succès de 100 % n'est pas un objectif sain : il signifie presque toujours que vous aurez travaillé des années de trop et mourrez au maximum de votre richesse. Morningstar calibre ses recommandations à 90 % de succès (10 % de ruine) sur 30 ans ([[guardrails-morningstar]]). pofo utilise 5 % par défaut pour ses solveurs, un choix prudent que le contrôle « acceptable ruin » vous laisse déplacer ([[utiliser-la-page-fire]]).
:::

## La précision illusoire : 2 % et 8 % sont souvent le même chiffre

Le simulateur affiche « 4,7 % » et l'esprit enregistre une précision d'orfèvre. Elle n'existe pas, pour trois raisons cumulées.

**Le bruit d'échantillonnage** est la moindre : avec 5 000 trajectoires, un vrai 5 % s'affiche entre 4,4 et 5,6 % ; gênant, mais borné.

**La sensibilité aux paramètres** est bien pire : baisser le rendement réel espéré de 0,5 point (bien en deçà de ce que quiconque sait estimer, [[rendements-attendus]]) peut doubler la ruine ; l'épaisseur des queues (le df de Student, [[queues-epaisses]]) la déplace encore. Vos paramètres sont incertains, donc votre ruine l'est au moins autant.

**Le choix du modèle domine tout** : le même plan peut afficher 2 % en fenêtres historiques, 5 % en paramétrique central, 9 % en stress de séquence et 14 % en échantillon mondial. Aucun n'est faux ; ils répondent à des questions différentes (« et si le futur ressemble à l'histoire de mes fonds / à un monde i.i.d. calibré prudemment / au même monde avec des sticky bears / au siècle des 16 pays développés »).

La conséquence pratique tient en une règle : **lisez la ruine en ordinal, pas en cardinal**. Elle compare admirablement (le plan A est plus robuste que le plan B ; ce levier réduit le risque plus que celui-là ; ce modèle pessimiste reste acceptable) et mesure médiocrement (« mon risque réel est 4,7 % »). Les décimales sont du bruit ; les écarts entre scénarios et entre colonnes sont du signal.

::: exemple Une décision bien posée
Plan : 1,2 M€, 42 000 €/an, 45 ans, pension 12 000 €/an à 66 ans. Lecture en intervalle : fenêtres historiques 1 %, central 4 %, stress séquence 7 %, échantillon mondial 11 %. Décision : le central et le stress sont sous 10 %, le broad-sample au-dessus de 10 % mais ses échecs surviennent après 80 ans, pension acquise ; plancher de dépenses à 34 000 € tenable. Verdict : plan acceptable, avec une règle écrite : si le taux de retrait courant dépasse 5 % (portefeuille sous ~840 k€), baisse au plancher jusqu'à retour sous 4,5 %. La même analyse avec des échecs précoces ou un plancher intenable aurait conclu : un an de plus, ou 10 % de dépenses en moins.
:::

## La ruine réelle ne ressemble pas à la ruine simulée

Dernier recadrage, le plus important pour dormir. Dans le simulateur, la ruine est une falaise : le solde passe par zéro un mardi et tout s'arrête. Dans la vie, l'échec d'un plan de retraite est un processus lent et **visible** : le portefeuille décroche de la trajectoire prévue, le taux de retrait courant monte année après année, les voyants passent à l'orange une décennie avant le gouffre. Les études sur les trajectoires historiques défaillantes le confirment : entre le moment où un plan « condamné » devient statistiquement identifiable et l'épuisement effectif, il s'écoule typiquement 8 à 15 ans, un préavis énorme pour qui a prévu des seuils d'action ([[quand-s-inquieter]], [[revue-annuelle]]).

C'est la vraie raison pour laquelle la probabilité de ruine, bien lue, est un instrument de **conception** et non d'angoisse : elle sert à comparer des plans et dimensionner des marges avant le départ. Après le départ, elle cède la place au pilotage : des indicateurs simples, des seuils écrits, des réponses préparées. Un plan à 8 % de ruine avec un pilote attentif est plus sûr qu'un plan à 3 % avec un pilote endormi.

## L'essentiel à retenir

- La ruine est une propriété du couple plan + modèle : lisez l'intervalle entre plusieurs modèles, jamais une colonne seule.
- Le chiffre suppose zéro réaction humaine : il mesure la robustesse intrinsèque, pas votre destin.
- Choisissez votre seuil selon vos filets réels et le prix de la marge en années de travail ; 5-10 % simulés est la zone de travail commune avec filets, 100 % de succès est un anti-objectif.
- Lisez en ordinal : les écarts comparent, les décimales mentent ; 2 % et 8 % sont souvent indiscernables une fois l'incertitude des paramètres comptée.
- La ruine réelle est lente et visible des années à l'avance : la parade d'après-départ n'est pas un chiffre plus bas, c'est un pilotage écrit ([[quand-s-inquieter]]).

---

## Pour aller plus loin

- Early Retirement Now, SWR Series volet 11 (comment noter une règle de retrait) et volet 46 (« The Need for Precision in an Uncertain World ») : [earlyretirementnow.com](https://earlyretirementnow.com) ([[serie-ern]]).
- Michael Kitces, « The Problem With FIREing At A 4% Withdrawal Rate » et « Is A Probability Of Success-Driven Retirement Plan Actually Riskier? » ([kitces.com](https://www.kitces.com)) : la lecture praticienne du taux de succès.
- Derek Tharp & Kitces sur les « guardrails » en probabilité de succès : le pilotage plutôt que le chiffre statique.
- Dans pofo : la vue mortalité × ruine et le solveur « acceptable ruin » ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
