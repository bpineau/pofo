# Pensions et revenus complémentaires dans le plan

Presque aucun plan réel n'est un portefeuille seul contre le monde. Il y a la pension légale, qui arrivera un jour ([[retraite-legale]]). Parfois des loyers ([[immobilier-en-retrait]]). Une activité dosée ([[retour-au-travail]]). Une rente, et peut-être un héritage lointain. Or la façon de **compter** ces flux dans le plan pèse plus lourd que la plupart des choix de portefeuille. Les analyses de sensibilité classent d'ailleurs la pension au deuxième rang des variables du plan, juste derrière les dépenses.

Ce chapitre est le manuel de comptabilité de ces revenus. Il pose d'abord la taxonomie : garantis, quasi sûrs, espérés, trois traitements distincts que l'on confond à ses dépens. Il détaille ensuite les mécanismes par lesquels un flux transforme un plan. Un flux ne fait pas que réduire les retraits ; il raccourcit l'horizon à risque, écrase la queue de longévité et vaut une allocation obligataire implicite. Vient la modélisation concrète sur la page FIRE (pension, revenus d'appoint, curseurs, presets et pièges de saisie), puis les cas particuliers (loyers, rentes, héritages délicats, revenus du conjoint). Enfin la contre-vérification, car un plan peut aussi être **trop** adossé. Quand les flux promis dominent tout, c'est leur risque propre (politique, locatif, santé de l'activité) qui devient le risque du plan.

::: cle La taxonomie qui décide de tout
**Trois** catégories, trois traitements.

Les **garantis** : pension légale liquidable, rentes en cours. Ils entrent au plan, décotés de leur risque propre, soit 10-20 % de décote politique pour la pension.

Les **quasi sûrs** : loyers d'un bien détenu, pension du conjoint. Ils entrent aussi, décotés de 15-25 %.

Les **espérés** : activité future, revenus d'appoint (side income), héritages. Ceux-là n'**entrent pas**. Le plan doit tenir sans eux. Ce sont des **marges**, nommées dans le plan et cultivées, mais jamais comptées ([[retour-au-travail]]).

La discipline paie double. Le plan reste honnête, et chaque bonne surprise le renforce au lieu de le sauver.
:::

## Ce qu'un flux fait à un plan : les quatre mécanismes

Un revenu de 15 k€/an dans un plan à 50 k€ de dépenses ne « réduit pas les retraits de 30 % ». Il fait quatre choses, dont trois échappent à l'intuition.

**1. Il réduit les retraits nets, mais seulement quand il tombe.** C'est l'évidence, avec sa subtilité temporelle. Une pension qui démarre à l'année 17 ne réduit **rien** pendant la phase de pont. De là vient la structure en deux régimes de tout plan français : une phase à découvert, puis une phase adossée ([[horizon-et-esperance-de-vie]]). Et de là vient le mécanisme suivant.

**2. Il raccourcit l'horizon à risque.** C'est le vrai service de la pension. Le « plan de 50 ans » devient un pont de 17 ans, complété ensuite. Le risque de séquence et la falaise du fixe se concentrent alors sur le seul pont ([[sequence-des-rendements]]). C'est pourquoi ajouter la pension au modèle divise souvent la ruine par 2 à 4, l'oubli le plus cher ([[erreurs-classiques-fire]]). La pension arrive **précisément** dans les scénarios longs, ceux où le portefeuille fatigue.

**3. Il écrase la queue de longévité.** Un flux **viager** (pension, rente) couvre le seul risque que le capital couvre mal : vivre très longtemps ([[rentes-et-annuites]], [[horizon-et-esperance-de-vie]]). Regardez la ruine pondérée par la mortalité. Les scénarios « vivant, ruiné, 95 ans » deviennent « vivant, au plancher pension, 95 ans ». C'est un tout autre événement.

**4. Il constitue une allocation obligataire implicite.** C'est la lecture TPAW ([[amortissement-abw]]). Une pension de 20 k€/an, actualisée, « pèse » 300-450 k€ d'obligation d'État indexée dans le patrimoine total. À risque global égal, le portefeuille **visible** peut donc porter plus d'actions ([[allocation-actions-obligations]], le curseur « couverture du plancher »). Le quinquagénaire à pension future qui garde 40 % d'obligations par prudence est souvent, en réalité, sous-investi. La prudence était déjà dans sa pension.

## La modélisation sur la page FIRE : les curseurs et leurs pièges

La page FIRE a deux entrées de flux, et leur bon usage fait la qualité du plan simulé ([[utiliser-la-page-fire]]).

**La pension** (« Pension /yr » plus « starts in year ») modélise un flux viager, réel et net. Réel, car indexé : la convention colle à la pension française. Y entrent la pension légale décotée, les rentes viagères et la réversion pondérée. Les presets stress, central et officiel encodent la fourchette M@REL ([[retraite-legale]]). Trois pièges de saisie guettent. D'abord le **montant brut** recopié du relevé : entrez plutôt le net décoté. Ensuite l'**année** trop optimiste : pour une carrière courte, comptez 67 ans plutôt que 64, sauf calcul contraire. Enfin l'oubli de la **deuxième** pension du couple : sommez les flux du ménage, chacun à sa date, quitte à pondérer l'écart d'années par une entrée moyenne.

**Le side income** (« Side income /yr » plus « until year ») modélise un flux temporaire, réel lui aussi. Y entrent les loyers nets décotés d'un bien qu'on vendra, l'activité **structurelle** d'un semi-FIRE et l'allocation chômage de la transition. L'activité ne se compte que dans ce cas, lorsqu'elle est un paramètre assumé du plan ([[retour-au-travail]]). Le piège est unique mais grave : y glisser les revenus **espérés** de la catégorie 3. Le simulateur vous dira alors ce que vous voulez entendre. C'est exactement le p-hacking de scénarios ([[pieges-des-simulateurs]]).

Reste la lecture après saisie. La sensibilité se **teste** : faites varier la pension de ±20 % et de ±2 ans. Si la ruine bascule, le plan est un pari sur la pension (voir la contre-vérification plus bas). Le solveur §09 affiche par ailleurs le prix des flux en équivalents ([[une-annee-de-plus]]). On y voit « 500 €/mois d'appoint pendant 5 ans » côtoyer « 80 k€ de capital », un menu qui remet les grandeurs en face.

::: science Les cas délicats : loyers, héritages, conjoint
**Les loyers** sont quasi sûrs, mais ni viagers ni sans travail. La règle vient de [[immobilier-en-retrait]] : un net-net réaliste, décoté de 15-25 %. Placez-les en side income jusqu'à l'année de vente prévue, ou en pension si la détention est à vie, et listez la valeur du bien en réserve. Nommez au passage leur risque propre : vacance, réglementation, gestion à 75 ans.

**Les héritages** forment le sujet inconfortable. Statistiquement probables pour beaucoup de quadragénaires FIRE, ils ne se comptent **jamais**. Le montant, la date et l'affectation sont triplement incertains, ne serait-ce qu'à cause de la dépendance possible des parents ([[sante-et-protection-sociale]]). Et faire dépendre son plan d'un décès pour boucler est malsain par construction. Catégorie 3, marge nommée, point final. Le plan qui a **besoin** de l'héritage n'est pas un plan.

**Le conjoint qui travaille encore** relève d'un autre cas ([[couple-et-famille]]). Son salaire est un side income daté, jusqu'à sa date de départ, avec sa volonté pour risque propre, à décoter si le décalage est subi. Sa pension future, elle, s'ajoute au flux pension du ménage.

**Les allocations de transition** (le chômage après une rupture conventionnelle) sont un side income court et sûr. Ces 18-24 mois amortissent exactement le début de la fenêtre fragile. À compter sans état d'âme.
:::

## La contre-vérification : le plan trop adossé

La discipline de ce chapitre pousse à adosser. La lucidité impose la borne inverse. Un plan dont 70-80 % du plancher repose sur des flux promis a **concentré** son risque sur ces promesses. La pension légale porte un risque politique : les décotes de 10-20 % existent pour cela, et un plan qui casse si la décote passe à 30 % mérite d'être regardé de près ([[retraite-legale]]). Les loyers portent un risque réglementaire et de gestion. L'activité structurelle porte un risque de santé et d'envie : le Barista de 48 ans sera-t-il encore le Barista de 61 ans ?

Le test de robustesse s'applique à la conception comme aux revues ([[revue-annuelle]]). Reprenez le plan **sans** chaque flux, un par un, le curseur à zéro, deux minutes par test. La ruine qui en sort n'a pas à rester sous le seuil, c'est tout l'intérêt des flux. Mais elle doit rester dans la zone du **rattrapable**, celle pour laquelle le playbook a des paliers ([[quand-s-inquieter]]). Si la perte d'un flux rend le plan irrécupérable, ce flux n'est pas une marge du plan. Il est le plan. Il mérite alors le traitement d'un actif central : diversification, entretien, plan B écrits.

::: exemple La comptabilité des flux d'un plan réel
Prenons le ménage type de ce livre, à la conception.

La pension légale du couple : M@REL donne 31 k€/an à 67 ans, décote 15 %, soit **26 k€ en pension à l'année 19** (garanti décoté). Les loyers du T2 conservé : 8,4 k€ net-net, décote 20 %, soit **6,7 k€ en side income jusqu'à l'année 12** (vente prévue à 60 ans, quasi sûr et daté). Le chômage de transition ajoute **18 k€ les années 1-2** (sûr et court). Les missions probables de madame (~10 k€/an), l'héritage plausible et la réversion sont **nommés** dans le bloc marges, mais comptés nulle part.

Les résultats parlent. La ruine centrale tombe à 3,9 %, contre 10,8 % en ignorant tous ces flux, et la moitié de l'écart vient de la seule pension. La contre-vérification tient. Sans les loyers, la ruine monte à 4,6 % (rattrapable). Avec la pension décotée à 30 % au lieu de 15, elle monte à 5,4 % (rattrapable). **Tout** à zéro, elle revient à 10,8 %, dur mais non irrécupérable avec les paliers. Le plan est adossé, pas suspendu. Vingt minutes de saisie rigoureuse : le facteur 2,8 sur la ruine était là, pas dans un choix d'ETF.
:::

## L'essentiel à retenir

- Trois catégories, trois traitements : garantis (comptés, décotés 10-20 %), quasi sûrs (comptés, décotés 15-25 %), espérés (**jamais** comptés, marges nommées et cultivées). La confusion entre les trois est l'erreur structurante du dimensionnement.
- Un flux fait quatre choses : réduire les retraits à sa date, raccourcir l'horizon à risque (le plan devient un pont), écraser la queue de longévité (s'il est viager), et constituer de l'allocation obligataire implicite (le portefeuille visible peut oser davantage).
- Sur la page FIRE : la pension est un flux viager, réel, net et décoté (sommé pour le couple, calé sur la bonne année) ; le side income est un flux temporaire et réel (loyers datés, activité **structurelle**, chômage de transition). Et jamais les espérés : un simulateur complaisant est un simulateur mort.
- Les héritages ne se comptent jamais ; le salaire du conjoint est un side income daté à risque propre ; les loyers gardent leur décote et leur date de vente.
- La contre-vérification ferme le chapitre : le plan privé de chaque flux, un par un, doit rester **rattrapable**. Le flux dont la perte est irrécupérable n'est pas une marge. Il est le plan, et se gère comme tel.

---

## Pour aller plus loin

- Early Retirement Now, volet 4 (la Sécurité sociale dans le SWR) et volet 32 (« You are a Pension Fund of One ») : la comptabilité des flux, version US ([[serie-ern]]).
- Dans pofo : les curseurs pension/side income et leurs presets, le solveur §09 comme menu des équivalences ([[utiliser-la-page-fire]]).
- Dans ce livre : [[retraite-legale]] (le flux n° 1 et sa décote), [[retour-au-travail]] (quand l'activité se compte), [[immobilier-en-retrait]] (les loyers), [[amortissement-abw]] (la richesse totale actualisée), [[horizon-et-esperance-de-vie]] (le pont).
