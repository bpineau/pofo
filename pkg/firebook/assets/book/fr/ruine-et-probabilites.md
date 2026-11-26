# La probabilité de ruine : la lire, la choisir, ne pas la subir

Tous les simulateurs de retraite résument leur verdict en un chiffre : la probabilité de ruine (ou son complément, le « taux de succès »). C'est le chiffre le plus regardé du sujet, et le plus mal compris. On le lit comme une météo (« 5 % de risque, ça va »), on le compare entre outils qui ne mesurent pas la même chose, on lui demande une précision qu'il n'a pas, et on oublie qu'il ne tient compte que des réactions prévues par la règle.

Cette page apprend à lire ce chiffre en professionnel : ce qu'il mesure exactement, comment choisir son seuil acceptable, pourquoi le même plan peut afficher 2 % ou 8 % selon les hypothèses, et ce que la ruine simulée a de très différent d'une ruine réelle.

::: cle La définition, sans raccourci
Dans pofo, une trajectoire échoue la première année où le portefeuille ne peut pas verser le retrait que demande la règle, après prise en compte des revenus du foyer. Le taux d'échec est la part des trajectoires qui connaissent au moins une telle année pendant l'horizon retenu. Il dépend donc de votre **plan** et du **modèle** qui produit les scénarios ([[historique-vs-parametrique]]).
:::

## Ce que le chiffre mesure, et ne mesure pas

Décomposons la définition, terme à terme, parce que chaque terme cache un piège.

**« La fraction des futurs simulés ».** Le chiffre est une fréquence dans une population de scénarios générés par un modèle. Le modèle change tout. S'il tire les années indépendamment, il sous-estime les grappes de mauvaises années ([[sequence-des-rendements]]). S'il rejoue l'histoire américaine, il hérite de son biais optimiste ([[etude-trinity]], [[pieges-des-simulateurs]]). S'il rejoue l'échantillon mondial, il inclut des pays et des époques peut-être plus durs que votre futur plausible ([[anarkulova-cederburg]]). Aucun n'est « le vrai ». C'est pourquoi les simulateurs sérieux montrent un plan sous plusieurs modèles à la fois. Leurs chiffres ne forment pas un intervalle de confiance : chacun pose une question différente, et certains ne regardent même pas votre portefeuille. Lisez-les un par un, jamais un seul.

**« Un retrait ne peut pas être versé ».** Imaginez qu'il reste 15 000 € à retirer la dernière année et que le portefeuille contienne exactement cette somme. Le ménage paie tout et termine à zéro : le plan a réussi. S'il reste une année à vivre avec un portefeuille vide, une pension qui couvre tout le budget peut encore suffire. Si elle laisse un manque de 1 000 €, c'est cette année-là que pofo compte l'échec. Regardez ensuite **quand** il survient et combien de dépenses restent découvertes ([[lire-un-fan-chart]]).

**« Selon la règle prévue ».** Avec des retraits fixes, le calcul maintient le montant indexé même après une baisse. Si vous avez prévu de réduire les voyages avant de toucher aux dépenses indispensables, simulez une règle qui le fasse ([[plancher-plafond]], [[guardrails-morningstar]]). Elle peut diminuer le taux d'échec au prix de plusieurs années de vacances en moins. Le montant des dépenses servies et la durée des baisses vous disent ce que cette souplesse a coûté.

**Et ce qu'il ignore superbement** : votre mortalité si l'horizon est fixe (une « ruine » à 97 ans concerne peu de monde, d'où l'intérêt de croiser la ruine avec les tables de mortalité, [[horizon-et-esperance-de-vie]]), les à-coups de dépenses réels, la fiscalité fine, et tous les filets hors modèle (famille, patrimoine immobilier, retour au travail).

## Choisir son seuil : pourquoi il n'y a pas de bonne réponse universelle

Quelle ruine accepter : 1 %, 5 %, 10 % ? La question semble technique ; elle est en réalité personnelle, et dépend de trois choses.

**La qualité de vos filets.** Une pension inscrite dans le calcul réduit déjà les retraits nécessaires. Si vous la comptez une deuxième fois parmi les secours possibles, vous surestimez votre marge. Pour un recours encore hors modèle, comme reprendre un emploi, précisez à quel âge il serait accessible et quel revenu il pourrait fournir.

**Le coût d'une année de marge.** Passer de 5 % à 2 % de ruine coûte typiquement 10 à 20 % de capital en plus, soit deux à quatre ans de travail. Passer de 5 % à 10 % les rend. Le seuil est un prix d'arbitrage entre deux risques ([[une-annee-de-plus]]) : chiffrez ce que chaque point de ruine vous coûte ou vous rend en années de vie active, la discussion devient concrète. Le cadre général de ce type d'arbitrage (utilité, équivalent certain, regret) est posé dans [[decider-sous-incertitude]].

**La nature de l'échec dans votre plan.** Regardez quand échouent les scénarios qui échouent. Un 8 % de ruine fait d'échecs tardifs (après 85 ans), adossés à une pension qui couvre le plancher, est plus confortable qu'un 4 % fait d'effondrements à 70 ans.

::: astuce Le réflexe des praticiens
Un solveur qui calcule un capital cible ou une date de départ a besoin d'un seuil d'échec acceptable. Morningstar retient par exemple 90 % de succès sur trente ans dans son scénario de référence ([[guardrails-morningstar]]). Pour choisir votre propre seuil, ouvrez quelques trajectoires qui échouent. Quel budget pourriez-vous y tenir ? Combien d'années ? Quelles dépenses resteriez-vous incapable de financer ?
:::

## La précision illusoire : pourquoi le même plan peut afficher 2 % ou 8 %

Le simulateur affiche « 4,7 % » et l'esprit enregistre une précision d'orfèvre. Elle n'existe pas, pour trois raisons cumulées.

**Le bruit d'échantillonnage.** Avec 2 000 trajectoires indépendantes et une probabilité de 5 % dans le modèle, environ 95 tirages sur 100 donnent une fréquence entre 4 et 6 %. Obtenir zéro échec sur un tirage fini ne prouve pas que le risque du modèle est nul. Ce calcul d'incertitude ne s'applique pas directement aux fenêtres historiques qui se chevauchent.

**La sensibilité aux paramètres** est bien pire. Baisser le rendement réel espéré de 0,5 point (une finesse que personne ne sait estimer, [[rendements-attendus]]) peut doubler la ruine. L'épaisseur des queues (les degrés de liberté de la loi de Student, [[queues-epaisses]]) la déplace encore. Vos paramètres sont incertains, donc votre ruine l'est au moins autant.

**Le choix du modèle domine tout** : le même plan peut afficher 2 % en fenêtres historiques, 5 % en paramétrique central, 9 % en stress de séquence et 14 % en échantillon mondial. Aucun n'est faux ; ils répondent à des questions différentes (« et si le futur ressemble à l'histoire de mes fonds / à un monde i.i.d. calibré prudemment / au même monde avec des sticky bears / et si j'avais été un épargnant en 60/40 domestique dans l'un des 16 pays développés du siècle »). Le dernier ne regarde pas votre portefeuille : c'est un crash-test, pas une mesure de votre risque.

::: figure trois-bruits
Les trois bruits mis à la même échelle, sur un seul plan de référence : le chiffre affiché, les deux points d'incertitude que laisse le tirage, les trois et demi que laissent les paramètres, et les trente-cinq que laisse le choix du modèle. La barre que l'on regarde en croyant lire une probabilité est la plus courte des trois, et de loin.
:::


La conséquence pratique tient en une règle : **lisez la ruine en ordinal, pas en cardinal**. Elle compare admirablement (le plan A est plus robuste que le plan B ; ce levier réduit le risque plus que celui-là ; ce modèle pessimiste reste acceptable) et mesure médiocrement (« mon risque réel est 4,7 % »). Les décimales sont du bruit ; les écarts entre scénarios et entre modèles sont du signal.

::: exemple Une décision bien posée
Plan : 1,2 M€, 42 000 €/an, 45 ans, pension 12 000 €/an à 66 ans. Lecture modèle par modèle : fenêtres historiques 1 %, central 4 %, stress de séquence 7 %, échantillon mondial 11 %. Décision : le central et le stress sont sous 10 %, l'échantillon mondial au-dessus de 10 % mais ses échecs surviennent après 80 ans, pension acquise ; plancher de dépenses à 34 000 € tenable. La pension n'en couvre que 12 000 € : les 22 000 € restants viennent encore du portefeuille, d'où la règle écrite. Verdict : plan acceptable, avec une règle écrite. Si le taux de retrait courant dépasse 5 % (portefeuille sous ~840 k€ avant la pension, ~600 k€ après, quand le besoin du portefeuille tombe à 30 000 €), on baisse au plancher jusqu'à retour sous 4,5 %. La même analyse avec des échecs précoces ou un plancher intenable aurait conclu : un an de plus, ou 10 % de dépenses en moins.
:::

## La ruine réelle ne ressemble pas à la ruine simulée

Le pourcentage d'échec efface la chronologie. Le graphique ci-dessous la rétablit pour un départ en 1966. À mesure que le capital baisse, le retrait prévu représente une part croissante du solde. Ce rapport donne un signal que le ménage peut suivre à sa revue annuelle. Il laisse parfois du temps pour agir. Un choc de marché brutal ou une grosse dépense imprévue peuvent au contraire raccourcir ce délai ([[quand-s-inquieter]], [[revue-annuelle]]).

::: figure preavis-1966
Départ en 1966, portefeuille américain 60/40 de pofo, retrait initial de 4 % puis indexation sur l'inflation. En haut, le patrimoine restant. En bas, le retrait annuel en pourcentage du solde. Le voyant passe au rouge en 1974, alors qu'il reste 673 k€ sur le million initial. Il reste rouge en 1975. Le portefeuille s'épuise en 1994. Ces dates appartiennent à cette trajectoire et aux seuils choisis ([[quand-s-inquieter]]).
:::

C'est la vraie raison pour laquelle la probabilité de ruine, bien lue, est un instrument de **conception** et non d'angoisse. Elle sert à comparer des plans et à dimensionner des marges avant le départ. Après le départ, elle cède la place au pilotage : des indicateurs simples, des seuils écrits, des réponses préparées. Un plan à 8 % de ruine avec un pilote attentif peut être plus sûr qu'un plan à 3 % avec un pilote endormi.

## L'essentiel à retenir

- La ruine est une propriété du couple plan + modèle : lisez plusieurs modèles, chacun pour la question qu'il pose, jamais un seul.
- Le calcul suit les décisions prévues par la règle, y compris les baisses de dépenses que vous lui avez données.
- Choisissez votre seuil selon vos filets réels et le prix de la marge en années de travail.
- Lisez en ordinal : les écarts comparent, les décimales mentent ; comparez deux plans avec les mêmes hypothèses.
- Une dégradation peut être visible des années à l'avance : la parade d'après-départ n'est pas un chiffre plus bas, c'est un pilotage écrit ([[quand-s-inquieter]]).

---

## Pour aller plus loin

- Early Retirement Now, SWR Series volet 11 (« Six Criteria to Grade Dynamic Withdrawal Rules ») et volet 46 (« The Need for Precision in an Uncertain World ») : [earlyretirementnow.com](https://earlyretirementnow.com) ([[serie-ern]]).
- Michael Kitces, « Flexible Spending Rules To Avoid FIREing At 4% » et « Does the Probability of Success Drive Your Retirement Plan? » ([kitces.com](https://www.kitces.com)) : la lecture praticienne du taux de succès.
- Derek Tharp et Michael Kitces sur les « guardrails » en probabilité de succès : le pilotage plutôt que le chiffre statique.
- Dans ce livre : [[historique-vs-parametrique]] (pourquoi les modèles divergent sur un même plan) et [[la-machine-pofo]] (comment pofo calcule cette ruine et la croise avec la mortalité).
