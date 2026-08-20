# La probabilité de ruine : la lire, la choisir, ne pas la subir

Tous les simulateurs de retraite résument leur verdict en un chiffre : la probabilité de ruine (ou son complément, le « taux de succès »). C'est le chiffre le plus regardé du sujet, et le plus mal compris. On le lit comme une météo (« 5 % de risque, ça va »), on le compare entre outils qui ne mesurent pas la même chose, on lui demande une précision qu'il n'a pas, et on oublie qu'il décrit un monde où personne ne réagit jamais.

Cette page apprend à lire ce chiffre en professionnel : ce qu'il mesure exactement, comment choisir son seuil acceptable, pourquoi 2 % et 8 % sont souvent indiscernables, et ce que la ruine simulée a de très différent d'une ruine réelle.

::: cle La définition, sans raccourci
La probabilité de ruine d'un plan est la fraction des futurs simulés (ou des fenêtres historiques rejouées) dans lesquels le portefeuille atteint zéro **avant** la fin de l'horizon, sous une règle de retrait appliquée mécaniquement, sans aucune réaction humaine. C'est une propriété du **couple** plan + modèle, jamais du plan seul : changer de modèle change le chiffre, parfois du simple au triple, sans que votre plan ait bougé ([[historique-vs-parametrique]]).
:::

## Ce que le chiffre mesure, et ne mesure pas

Décomposons la définition, terme à terme, parce que chaque terme cache un piège.

**« La fraction des futurs simulés ».** Le chiffre est une fréquence dans une population de scénarios générés par un modèle. Le modèle change tout. S'il tire les années indépendamment, il sous-estime les grappes de mauvaises années ([[sequence-des-rendements]]). S'il rejoue l'histoire américaine, il hérite de son biais optimiste ([[etude-trinity]], [[pieges-des-simulateurs]]). S'il rejoue l'échantillon mondial, il inclut des pays et des époques peut-être plus durs que votre futur plausible ([[anarkulova-cederburg]]). Aucun n'est « le vrai ». C'est pourquoi les simulateurs sérieux montrent le même plan sous plusieurs modèles à la fois, et pourquoi la bonne lecture est l'**intervalle** qu'ils dessinent, jamais un modèle isolé.

**« Le portefeuille atteint zéro ».** La ruine du simulateur est binaire et terminale. Elle ne distingue pas l'échec à 71 ans de l'échec à 94 ans, ni le plan qui touche zéro dans son dernier mois (échec) de celui qui finit avec 15 000 € en poche (succès !). Deux plans à 5 % de ruine peuvent cacher des réalités très différentes : l'un échoue tôt et brutalement, l'autre s'essouffle en toute fin de parcours avec la pension légale en soutien. D'où l'intérêt de regarder deux choses de plus. **Quand** surviennent les échecs, et quelle richesse médiane reste en fin d'horizon ([[lire-un-fan-chart]]).

**« Sans aucune réaction humaine ».** L'hypothèse la plus irréaliste et la plus utile. Irréaliste : aucun humain ne maintient 40 000 € de retraits indexés pendant que son portefeuille passe de 1 M€ à 150 k€ ; il aurait coupé ses dépenses des années plus tôt ([[quand-s-inquieter]]). Utile : c'est justement parce que la règle simulée est aveugle que le chiffre mesure la robustesse **intrinsèque** du plan, sans se payer de mots sur une flexibilité future hypothétique. Un plan flexible se simule avec sa propre règle flexible ([[plancher-plafond]], [[guardrails-morningstar]]). Mais alors la « ruine » chute, et c'est le niveau de vie délivré qu'il faut regarder. La flexibilité ne supprime pas la douleur, elle la déplace vers des années de dépenses réduites.

**Et ce qu'il ignore superbement** : votre mortalité (une « ruine » à 97 ans concerne peu de monde, d'où l'intérêt de croiser la ruine avec les tables de mortalité, [[horizon-et-esperance-de-vie]]), les à-coups de dépenses réels, la fiscalité fine, et tous les filets hors modèle (famille, patrimoine immobilier, retour au travail).

## Choisir son seuil : pourquoi il n'y a pas de bonne réponse universelle

Quelle ruine accepter : 1 %, 5 %, 10 % ? La question semble technique ; elle est en réalité personnelle, et dépend de trois choses.

**La qualité de vos filets.** La ruine simulée suppose zéro recours. Un quadragénaire employable, propriétaire, avec pension légale à venir et famille solidaire peut rationnellement accepter 10 à 15 % de ruine **simulée**, parce que sa ruine **réelle** (fin de vie dans le dénuement sans aucun recours) est bien plus rare que le chiffre. Une personne de 60 ans sans pension notable, sans immobilier et sans possibilité de retravailler doit lire le chiffre presque littéralement, et viser bas.

**Le coût d'une année de marge.** Passer de 5 % à 2 % de ruine coûte typiquement 10 à 20 % de capital en plus, soit deux à quatre ans de travail. Passer de 5 % à 10 % les rend. Le seuil est un prix d'arbitrage entre deux risques ([[une-annee-de-plus]]) : chiffrez ce que chaque point de ruine vous coûte ou vous rend en années de vie active, la discussion devient concrète. Le cadre général de ce type d'arbitrage (utilité, équivalent certain, regret) est posé dans [[decider-sous-incertitude]].

**La nature de l'échec dans votre plan.** Regardez quand échouent les scénarios qui échouent. Un 8 % de ruine fait d'échecs tardifs (après 85 ans), adossés à une pension qui couvre le plancher, est plus confortable qu'un 4 % fait d'effondrements à 70 ans.

::: astuce Le réflexe des praticiens
Les planificateurs financiers sérieux (Kitces en tête) convergent vers une fourchette de travail de 10 à 20 % de ruine simulée pour des plans avec filets et règle d'ajustement, et rappellent qu'un taux de succès de 100 % n'est pas un objectif sain. Il signifie presque toujours que vous aurez travaillé des années de trop et mourrez au maximum de votre richesse. Morningstar calibre ses recommandations à 90 % de succès (10 % de ruine) sur 30 ans ([[guardrails-morningstar]]). Un solveur qui dimensionne un plan (capital cible, date de départ, retrait tenable) a besoin d'une ruine acceptable en entrée. Fixez-la vous-même. 10 % est un point de départ raisonnable, à descendre quand vos filets manquent, à remonter quand ils sont solides et que votre règle de retrait sait s'ajuster.
:::

## La précision illusoire : 2 % et 8 % sont souvent le même chiffre

Le simulateur affiche « 4,7 % » et l'esprit enregistre une précision d'orfèvre. Elle n'existe pas, pour trois raisons cumulées.

**Le bruit d'échantillonnage** est la moindre. Avec les 2 000 trajectoires que tirent couramment les simulateurs, un vrai 5 % s'affiche entre 4 et 6 % selon le tirage. Gênant, mais borné.

**La sensibilité aux paramètres** est bien pire. Baisser le rendement réel espéré de 0,5 point (une finesse que personne ne sait estimer, [[rendements-attendus]]) peut doubler la ruine. L'épaisseur des queues (les degrés de liberté de la loi de Student, [[queues-epaisses]]) la déplace encore. Vos paramètres sont incertains, donc votre ruine l'est au moins autant.

**Le choix du modèle domine tout** : le même plan peut afficher 2 % en fenêtres historiques, 5 % en paramétrique central, 9 % en stress de séquence et 14 % en échantillon mondial. Aucun n'est faux ; ils répondent à des questions différentes (« et si le futur ressemble à l'histoire de mes fonds / à un monde i.i.d. calibré prudemment / au même monde avec des sticky bears / au siècle des 16 pays développés »).

La conséquence pratique tient en une règle : **lisez la ruine en ordinal, pas en cardinal**. Elle compare admirablement (le plan A est plus robuste que le plan B ; ce levier réduit le risque plus que celui-là ; ce modèle pessimiste reste acceptable) et mesure médiocrement (« mon risque réel est 4,7 % »). Les décimales sont du bruit ; les écarts entre scénarios et entre modèles sont du signal.

::: exemple Une décision bien posée
Plan : 1,2 M€, 42 000 €/an, 45 ans, pension 12 000 €/an à 66 ans. Lecture en intervalle : fenêtres historiques 1 %, central 4 %, stress de séquence 7 %, échantillon mondial 11 %. Décision : le central et le stress sont sous 10 %, l'échantillon mondial au-dessus de 10 % mais ses échecs surviennent après 80 ans, pension acquise ; plancher de dépenses à 34 000 € tenable. Verdict : plan acceptable, avec une règle écrite. Si le taux de retrait courant dépasse 5 % (portefeuille sous ~840 k€), on baisse au plancher jusqu'à retour sous 4,5 %. La même analyse avec des échecs précoces ou un plancher intenable aurait conclu : un an de plus, ou 10 % de dépenses en moins.
:::

## La ruine réelle ne ressemble pas à la ruine simulée

Dernier recadrage, le plus important pour dormir. Dans le simulateur, la ruine est une falaise : le solde passe par zéro un mardi et tout s'arrête. Dans la vie, l'échec d'un plan de retraite est un processus lent et **visible** : le portefeuille décroche de la trajectoire prévue, le taux de retrait courant monte année après année, les voyants passent à l'orange longtemps avant le gouffre. Les trajectoires historiques défaillantes le confirment : entre le moment où un plan « condamné » devient statistiquement identifiable et l'épuisement effectif, il s'écoule une dizaine d'années au moins, et vingt ans sur le pire millésime connu. Le rouge du millésime 1966 se confirme en 1974-1975 ; le capital ne s'épuise qu'en 1994. Un préavis énorme pour qui a prévu des seuils d'action ([[quand-s-inquieter]], [[revue-annuelle]]).

::: figure preavis-1966
Le millésime 1966, le pire que l'histoire américaine ait produit, sous un retrait fixe de 4 % indexé. En haut, le capital qui décroche puis s'épuise. En bas, le même plan vu par les voyants du tableau de bord ([[quand-s-inquieter]]) : l'orange dès 1967, plus jamais de vert à partir de 1970, le rouge en 1974 et sa confirmation en 1975. Le compte n'est vide qu'en 1994. Le jour où le voyant passe au rouge, il reste encore 670 k€ en caisse et vingt revues annuelles pour agir : voilà à quoi ressemble une ruine réelle.
:::

C'est la vraie raison pour laquelle la probabilité de ruine, bien lue, est un instrument de **conception** et non d'angoisse. Elle sert à comparer des plans et à dimensionner des marges avant le départ. Après le départ, elle cède la place au pilotage : des indicateurs simples, des seuils écrits, des réponses préparées. Un plan à 8 % de ruine avec un pilote attentif est plus sûr qu'un plan à 3 % avec un pilote endormi.

## L'essentiel à retenir

- La ruine est une propriété du couple plan + modèle : lisez l'intervalle entre plusieurs modèles, jamais un modèle seul.
- Le chiffre suppose zéro réaction humaine. Il mesure la robustesse intrinsèque, pas votre destin.
- Choisissez votre seuil selon vos filets réels et le prix de la marge en années de travail ; 10-20 % simulés est la zone de travail des praticiens quand les filets et une règle d'ajustement existent, 100 % de succès est un anti-objectif.
- Lisez en ordinal : les écarts comparent, les décimales mentent ; 2 % et 8 % sont souvent indiscernables une fois l'incertitude des paramètres comptée.
- La ruine réelle est lente et visible des années à l'avance : la parade d'après-départ n'est pas un chiffre plus bas, c'est un pilotage écrit ([[quand-s-inquieter]]).

---

## Pour aller plus loin

- Early Retirement Now, SWR Series volet 11 (« Six Criteria to Grade Withdrawal Rules ») et volet 46 (« The Need for Precision in an Uncertain World ») : [earlyretirementnow.com](https://earlyretirementnow.com) ([[serie-ern]]).
- Michael Kitces, « Flexible Spending Rules To Avoid FIREing At 4% » et « Is A Probability Of Success-Driven Retirement Plan Actually Riskier? » ([kitces.com](https://www.kitces.com)) : la lecture praticienne du taux de succès.
- Derek Tharp et Michael Kitces sur les « guardrails » en probabilité de succès : le pilotage plutôt que le chiffre statique.
- Dans ce livre : [[historique-vs-parametrique]] (pourquoi les modèles divergent sur un même plan) et [[la-machine-pofo]] (comment pofo calcule cette ruine et la croise avec la mortalité).
