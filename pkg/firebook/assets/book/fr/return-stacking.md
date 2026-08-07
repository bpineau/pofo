# Return stacking, overlays et portable alpha : empiler les primes

Tout diversifiant pose au rentier un problème de financement. Ajouter 10 % de trend ([[managed-futures]]) ou d'or ([[or-en-retrait]]) signifie vendre 10 % d'actions ou d'obligations, donc renoncer à une part de la prime qui fait vivre le plan ([[primes-de-risque]]). C'est le dilemme classique de la table des défenses : la protection se paie en espérance. Le **return stacking** (littéralement, l'empilement de rendements) propose une troisième voie : utiliser un levier modéré, logé **à l'intérieur** de fonds construits pour cela, afin de détenir les diversifiants **en plus** du cœur actions-obligations, et non à sa place.

L'idée n'est pas une mode de forum. C'est la version grand public du **portable alpha**, une technique institutionnelle vieille de quarante ans, remise en lumière par les travaux de Corey Hoffstein (Newfound) et ReSolve Asset Management, et incarnée par une génération de fonds « efficient core » et « stacked ». Cet article vous en donne la mécanique exacte, l'arithmétique (y compris le coût de financement que le marketing oublie), ce que 2022 a appris à ses adeptes, la mise en œuvre européenne, et les règles de dose. Après lecture, vous saurez décider si votre plan mérite un étage de plus, et surtout lequel.

::: cle L'idée en une phrase
Un future sur indice ne demande en garantie qu'une fraction de son exposition. Un fonds peut donc détenir 90 % d'actions au comptant et y superposer 60 % d'obligations via futures → 150 % d'exposition pour 100 € investis, financée implicitement au taux court. Détenir 67 € de ce fonds équivaut à un 60/40 classique, et il vous reste 33 € pour loger du trend ou de l'or **sans rien avoir vendu**. Le levier ne sert pas ici à amplifier un pari, il sert à faire tenir plus de diversification dans le même portefeuille.
:::

## La mécanique, et l'arithmétique complète

Le prototype est le fonds **90/60** : 90 % d'actions en direct, 60 % d'obligations d'État via futures, soit un 60/40 à levier 1,5. Les déclinaisons « stacked » empilent autre chose que des obligations : 100 % actions + 100 % trend, 100 % obligations + 100 % trend, etc. Trois points d'arithmétique commandent tout le reste.

**Le coût de financement d'abord.** L'exposition au-delà de 100 % est financée au taux monétaire (c'est le prix implicite d'un future, [[levier-et-marges]] pour la mécanique). L'étage empilé ne rapporte donc pas son rendement brut, mais son rendement **moins le cash**. Empiler des obligations qui rendent 3 % quand le taux court est à 3,5 % détruit de la valeur ; la même pile devient rentable quand la courbe se pentifie. En termes de primes, un étage empilé capture une prime **au-dessus du cash**, jamais un rendement absolu. C'est le critère d'admission de tout étage : a-t-il une espérance positive nette du taux court et des frais ?

**Le volatility drag ensuite.** Le levier amplifie la volatilité, et la volatilité coûte en rendement composé ([[rendements-arithmetiques-geometriques]]). Un 60/40 à levier 1,5 a un drag supérieur au 60/40 simple ; tant que les deux jambes se compensent (corrélation actions-obligations négative ou nulle), l'effet reste petit devant la prime ajoutée. Ce qui amène directement au troisième point.

**La corrélation enfin, leçon 2022.** L'année où actions et obligations ont chuté ensemble, le 90/60 a encaissé environ une fois et demie la perte du 60/40 (de l'ordre de −25 % contre −17 %). Le levier ne crée pas de nouvelle défense, il **agrandit** ce qu'on lui donne : empiler des obligations sur des actions n'aide que dans les régimes où les obligations défendent ([[regimes-de-marche]]). C'est l'argument le plus sérieux pour préférer, en décumulation, l'empilement d'un actif à corrélation vraiment basse, le trend étant le candidat naturel, plutôt que l'empilement de duration seule.

::: science D'où vient l'idée : le portable alpha institutionnel
Dans les années 1980, PIMCO lance StocksPLUS : l'exposition actions via futures ne mobilisant que peu de capital, le reste est investi en obligations gérées activement ; si celles-ci battent le cash (le coût du future), le fonds bat l'indice actions. Les caisses de retraite ont généralisé le principe sous le nom de portable alpha, transporter une source de rendement au-dessus d'un beta, avant que la crise de 2008 n'en révèle la limite (des piles illiquides financées à court terme, des appels de marge en pleine panique). La version moderne « return stacked » en retient la leçon : n'empiler que des instruments **liquides** (futures sur indices, obligations d'État, programmes de trend), avec un levier plafonné et un rebalancement quotidien dans le fonds, sans appel de marge pour le porteur. Les backtests publiés (Newfound/ReSolve, et les études académiques sur le 60/40 levé) montrent le résultat attendu : à long terme, un 60/40 à levier 1,5 domine le 60/40 en rendement pour un drawdown comparable à un 100 % actions, et l'ajout d'un étage trend améliore nettement les pires trajectoires, exactement la géométrie que cherche un rentier ([[sequence-des-rendements]]).
:::

## Ce que ça change pour un plan de décumulation

Le return stacking répond à deux situations bien précises, et il faut résister à l'envie de lui en faire dire plus.

**Situation un : le budget de diversification est douloureux.** Vous êtes convaincu par le dossier du trend (0,2-0,4 point de SWR sur les pires millésimes), mais amputer la poche actions vous coûte en espérance centrale. Un cœur en fonds efficient core libère la place → le plan détient l'équivalent de son allocation cible ET son diversifiant, l'amélioration des queues ne se paie plus en médiane. C'est l'usage le plus propre, et le plus étayé.

**Situation deux : le capital est un peu court.** À plan constant, un levier global de 1,1 à 1,3 relève l'espérance et donc le taux de retrait soutenable, au prix de queues plus profondes. C'est un usage défendable pour un plan jeune et flexible, dangereux pour un plan tendu (le levier agrandit aussi le risque de séquence, relire [[levier-et-marges]] avant toute décision, ses cinq règles s'appliquent intégralement ici).

Et une non-situation : le stacking ne transforme pas un mauvais étage en bon. Empiler une stratégie à espérance nulle après frais (la plupart des « alternatifs » du chapitre [[global-macro]]) ajoute du risque et du frottement pour rien. L'empilement est un **amplificateur de convictions justifiées**, pas une source de rendement en soi.

## La mise en œuvre européenne, sans se raconter d'histoires

L'offre UCITS est **jeune et étroite**, c'est la vraie limite du chapitre. Les fonds efficient core existent en format européen (y compris des déclinaisons zone euro, actions de la zone + futures Bund empilés) mais l'encours reste modeste, et les fonds « stacked » trend ou or, déjà rares aux États-Unis, n'ont pour la plupart pas (encore) d'équivalent UCITS. D'où trois disciplines d'achat. **La liquidité du véhicule** : un petit ETF à encours faible peut fermer, ce qui pour un rentier signifie une vente forcée et une facture fiscale non choisie ([[etf-ucits-europeens]]). **La transparence de la pile** : exiger de savoir exactement ce qui est empilé, à quel levier, financé comment ; un document qui ne permet pas de reconstituer « X % de ceci + Y % de cela moins le cash » est un refus d'achat. **Le logement** : ces fonds vivent en CTO (pas d'éligibilité PEA, rares UC), leur place se dispute donc au trend et à l'or dans le même budget d'enveloppe ([[enveloppes-francaises]]).

L'alternative artisanale (répliquer soi-même la pile avec des futures sur un compte sur marge) existe et son verdict est le même que pour le trend maison : c'est un métier, avec des appels de marge, du roll trimestriel et une fiscalité pénible, à réserver aux profils qui l'exercent déjà. Le fonds à levier **quotidien** de type ETF x2 n'est pas un substitut : son rebalancement journalier produit une érosion en marché agité (le beta slippage) qui le disqualifie comme brique de long terme.

**Dans pofo**, la brique se teste avant de s'acheter : le catalogue embarque des historiques reconstruits de fonds efficient core et de stratégies empilées actions + trend, calibrés sur les indices de référence. Composez votre plan actuel contre sa version « cœur levé + trend libéré » et lisez les trois mêmes juges que d'habitude → le central (qui doit monter un peu), le stress 1966/1973 (qui doit s'améliorer nettement), et le drawdown réel maximal (qui dit le prix psychologique, [[utiliser-la-page-fire]], [[lire-un-fan-chart]]).

::: exemple Libérer 25 % de diversification sans vendre une action
Plan : 1 M€, 38 000 €/an, 60 % actions / 40 % obligations, 45 ans. Version empilée : 67 % en fonds 90/60 (soit 60 actions + 40 obligations d'exposition), 18 % trend, 10 % or, 5 % cash. Exposition totale ≈ 133 %, dont un étage défensif de 33 % qui n'existait pas. Lecture type au simulateur : central 3,7 → 3,9 % (la prime empilée nette du cash), stress 5,4 → 4,6 % (le trend et l'or travaillent les régimes longs sans avoir désarmé les actions), pire drawdown réel −34 % → −29 %, mais l'année 2022 rejouée fait −21 % contre −17 % (le levier obligataire pique quand tout baisse ensemble). La clause écrite au plan : « levier global plafonné à 1,35 ; l'étage empilé est jugé net du taux court ; si le fonds cœur ferme ou change de politique, retour au 60/40 simple sous un mois ». Sans accepter la ligne 2022, ne pas signer les deux autres.
:::

## Les règles de dose

Elles tiennent en quatre lignes. Le levier **global** du plan (exposition totale / capital) reste sous 1,5, et sous 1,3 pour un plan déjà tendu ou un horizon court. L'étage empilé est réservé aux instruments à dossier solide et corrélation basse (obligations d'État quand leur prime au-dessus du cash est positive, trend, or), jamais aux paris. Le levier vit **dans** les fonds, jamais sur un compte sur marge personnel adossé au portefeuille de retrait. Et le tout se juge comme n'importe quelle brique, sur les pires millésimes du simulateur et non sur le backtest de la plaquette ([[pieges-des-simulateurs]]).

## L'essentiel à retenir

- Le return stacking loge un levier modéré dans des fonds construits pour cela, afin de détenir les diversifiants en plus du cœur actions-obligations au lieu de les financer en vendant ce cœur ; c'est le portable alpha institutionnel en format grand public.
- L'arithmétique a trois lignes : tout étage empilé rapporte son rendement moins le cash, le levier amplifie le volatility drag, et il agrandit ce qu'on lui donne (2022 → un 90/60 perd 1,5 fois un 60/40 quand actions et obligations chutent ensemble).
- Pour le rentier, l'usage propre est de libérer le budget de diversification : cœur efficient core + trend/or dégagés, queues améliorées sans sacrifier la médiane ; l'usage levier-pour-rendement existe mais aggrave le risque de séquence.
- L'offre UCITS est jeune : exiger liquidité, transparence complète de la pile et logement CTO assumé ; fuir les ETF à levier quotidien comme substituts, et l'auto-réplication sur marge sauf compétence établie.
- Dose : levier global ≤ 1,3-1,5, étages réservés aux primes documentées, clause de démontage écrite. L'empilement amplifie les bonnes décisions comme les mauvaises, il n'en prend aucune à votre place.

---

## Pour aller plus loin

- Corey Hoffstein (Newfound Research) et ReSolve AM : les papiers fondateurs « Return Stacking: Strategies for Overcoming a Low Return Environment » (2021) et le site returnstacked.com.
- WisdomTree : la documentation des fonds Efficient Core (le 90/60 originel et ses déclinaisons UCITS).
- PIMCO : l'histoire de StocksPLUS et la littérature portable alpha ; AQR, « Why Not 100% Equities » (Asness, 1996), l'argument académique du 60/40 levé.
- Dans ce livre : [[levier-et-marges]] (les règles non négociables, qui s'appliquent toutes ici), [[managed-futures]] (l'étage le plus souvent empilé), [[rendements-arithmetiques-geometriques]] (le volatility drag), [[portefeuilles-tous-temps]] (le cousin sans levier).
