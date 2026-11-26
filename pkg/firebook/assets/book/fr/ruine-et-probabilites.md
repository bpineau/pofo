# La probabilité de ruine : la lire, la choisir, ne pas la subir

Une simulation affiche « 95 % de succès ». Les 5 % restants correspondent-ils à un retrait manqué de quelques euros la dernière année, ou à quinze années de dépenses essentielles sans financement ? Le pourcentage seul ne le dit pas.

Pour l'interpréter, commencez par deux questions. Quel événement l'outil appelle-t-il un échec ? Quelles décisions votre plan prend-il quand la situation se dégrade ?

::: cle La définition, sans raccourci
Dans pofo, une trajectoire échoue la première année où le portefeuille ne peut pas verser le retrait que demande la règle, après prise en compte des revenus du foyer. Le taux d'échec est la part des trajectoires qui connaissent au moins une telle année pendant l'horizon retenu. Il dépend donc de votre **plan** et du **modèle** qui produit les scénarios ([[historique-vs-parametrique]]).
:::

## Ce que le chiffre mesure, et ne mesure pas

Décomposons la définition, terme à terme, parce que chaque terme cache un piège.

**« La fraction des futurs simulés ».** Le chiffre est une fréquence dans une population de scénarios générés par un modèle. Le modèle change tout. S'il tire les années indépendamment, il sous-estime les grappes de mauvaises années ([[sequence-des-rendements]]). S'il rejoue l'histoire américaine, il hérite de son biais optimiste ([[etude-trinity]], [[pieges-des-simulateurs]]). S'il rejoue l'échantillon mondial, il inclut des pays et des époques peut-être plus durs que votre futur plausible ([[anarkulova-cederburg]]). Aucun n'est « le vrai ». C'est pourquoi les simulateurs sérieux montrent un plan sous plusieurs modèles à la fois. Leurs chiffres ne forment pas un intervalle de confiance : chacun pose une question différente, et certains ne regardent même pas votre portefeuille. Lisez-les un par un, jamais un seul.

**« Un retrait ne peut pas être versé ».** Imaginez qu'il reste 15 000 € à retirer la dernière année et que le portefeuille contienne exactement cette somme. Le ménage paie tout et termine à zéro : le plan a réussi. S'il reste une année à vivre avec un portefeuille vide, une pension qui couvre tout le budget peut encore suffire. Si elle laisse un manque de 1 000 €, c'est cette année-là que pofo compte l'échec. Regardez ensuite **quand** il survient et combien de dépenses restent découvertes ([[lire-un-fan-chart]]).

**« Selon la règle prévue ».** Avec des retraits fixes, le calcul maintient le montant indexé même après une baisse. Si vous avez prévu de réduire les voyages avant de toucher aux dépenses indispensables, simulez une règle qui le fasse ([[plancher-plafond]], [[guardrails-morningstar]]). Elle peut diminuer le taux d'échec au prix de plusieurs années de vacances en moins. Le montant des dépenses servies et la durée des baisses vous disent ce que cette souplesse a coûté.

**La durée couverte.** Avec un horizon fixe de quarante ans, chaque ménage doit couvrir les quarante années. Avec une durée de vie simulée, un manque ne compte que si le ménage est encore en vie ([[horizon-et-esperance-de-vie]]). Dans les deux cas, vérifiez quels revenus, dépenses exceptionnelles et recours le calcul inclut.

## Choisir son seuil : pourquoi il n'y a pas de bonne réponse universelle

Quelle ruine accepter : 1 %, 5 %, 10 % ? La question semble technique ; elle est en réalité personnelle, et dépend de trois choses.

**La qualité de vos filets.** Une pension inscrite dans le calcul réduit déjà les retraits nécessaires. Si vous la comptez une deuxième fois parmi les secours possibles, vous surestimez votre marge. Pour un recours encore hors modèle, comme reprendre un emploi, précisez à quel âge il serait accessible et quel revenu il pourrait fournir.

**Le coût d'une année de marge.** Passer de 5 % à 2 % de ruine coûte typiquement 10 à 20 % de capital en plus, soit deux à quatre ans de travail. Passer de 5 % à 10 % les rend. Le seuil est un prix d'arbitrage entre deux risques ([[une-annee-de-plus]]) : chiffrez ce que chaque point de ruine vous coûte ou vous rend en années de vie active, la discussion devient concrète. Le cadre général de ce type d'arbitrage (utilité, équivalent certain, regret) est posé dans [[decider-sous-incertitude]].

**La nature de l'échec dans votre plan.** Regardez quand échouent les scénarios qui échouent. Un 8 % de ruine fait d'échecs tardifs (après 85 ans), adossés à une pension qui couvre le plancher, est plus confortable qu'un 4 % fait d'effondrements à 70 ans.

::: astuce Le réflexe des praticiens
Un solveur qui calcule un capital cible ou une date de départ a besoin d'un seuil d'échec acceptable. Morningstar retient par exemple 90 % de succès sur trente ans dans son scénario de référence ([[guardrails-morningstar]]). Pour choisir votre propre seuil, ouvrez quelques trajectoires qui échouent. Quel budget pourriez-vous y tenir ? Combien d'années ? Quelles dépenses resteriez-vous incapable de financer ?
:::

## D'où vient l'incertitude autour du pourcentage ?

Le simulateur affiche « 4,7 % » et l'esprit enregistre une précision d'orfèvre. Elle n'existe pas, pour trois raisons cumulées.

**Le bruit d'échantillonnage.** Avec 2 000 trajectoires indépendantes et une probabilité de 5 % dans le modèle, environ 95 tirages sur 100 donnent une fréquence entre 4 et 6 %. Obtenir zéro échec sur un tirage fini ne prouve pas que le risque du modèle est nul. Ce calcul d'incertitude ne s'applique pas directement aux fenêtres historiques qui se chevauchent.

**La sensibilité aux paramètres.** Gardez la même méthode de simulation, mais baissez de 0,5 point le rendement réel attendu ([[rendements-attendus]]). Les retraits pèsent alors plus lourd sur le portefeuille. Refaites le même exercice en augmentant la fréquence des pertes extrêmes ([[queues-epaisses]]). Ces écarts vous indiquent quelles hypothèses méritent le plus d'attention dans votre plan.

**Le choix du modèle.** Les fenêtres historiques rejouent des années observées. Un modèle paramétrique tire des rendements selon des hypothèses explicites. Un test sur plusieurs pays peut, lui, observer un portefeuille domestique standard différent du vôtre. Un résultat de 2 % dans le premier et de 14 % dans le dernier décrit deux expériences différentes. L'intervalle de 2 à 14 % n'a donc pas d'interprétation comme intervalle de confiance.

::: figure trois-bruits
Les trois bruits mis à la même échelle, sur un seul plan de référence : le chiffre affiché, les deux points d'incertitude que laisse le tirage, les trois et demi que laissent les paramètres, et les trente-cinq que laisse le choix du modèle. La barre que l'on regarde en croyant lire une probabilité est la plus courte des trois, et de loin.
:::


Comparez d'abord les plans avec **les mêmes scénarios**. Si le classement change lorsque vous modifiez le rendement attendu ou le modèle, votre préférence dépend de cette hypothèse. S'il se maintient, vous avez une raison de le prendre au sérieux. Une différence entre 2 % et 8 % peut être bien réelle dans un modèle donné, même quand on connaît mal le risque dans la vie réelle.

::: exemple Une décision bien posée
Un ménage prévoit 42 000 € de dépenses annuelles et touchera une pension de 12 000 € à 66 ans. Avant la pension, le portefeuille doit fournir 42 000 € par an. Après, 30 000 €. Si le budget peut descendre à 34 000 €, il restera encore 22 000 € par an à financer après la pension. Simulez la règle qui déclenche cette baisse et relevez le nombre d'années passées à 34 000 € dans les mauvais scénarios. Le ménage peut alors juger si le gain sur le risque d'échec justifie la baisse de train de vie.
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
- Pour choisir un seuil, examinez le montant et la durée des dépenses non couvertes, les recours réellement disponibles et le coût d'une marge supplémentaire.
- Distinguez la variation due au tirage de celle qui vient des hypothèses. Augmenter le nombre de simulations réduit la première.
- En cours de retraite, des seuils d'action écrits aident à réagir à une hausse durable du taux de retrait courant ([[quand-s-inquieter]]).

---

## Pour aller plus loin

- Early Retirement Now, SWR Series volet 11 (« Six Criteria to Grade Dynamic Withdrawal Rules ») et volet 46 (« The Need for Precision in an Uncertain World ») : [earlyretirementnow.com](https://earlyretirementnow.com) ([[serie-ern]]).
- Michael Kitces, « Flexible Spending Rules To Avoid FIREing At 4% » et « Does the Probability of Success Drive Your Retirement Plan? » ([kitces.com](https://www.kitces.com)) : la lecture praticienne du taux de succès.
- Derek Tharp et Michael Kitces sur les « guardrails » en probabilité de succès : le pilotage plutôt que le chiffre statique.
- Dans ce livre : [[historique-vs-parametrique]] (pourquoi les modèles divergent sur un même plan) et [[la-machine-pofo]] (comment pofo calcule cette ruine et la croise avec la mortalité).
