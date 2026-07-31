# Managed futures et suivi de tendance : la diversification qui travaille dans les crises

Dans la table des défenses ([[actifs-defensifs]]), une ligne restait à pourvoir : les **régimes longs**. Ce sont ces marchés baissiers ou inflationnistes qui s'étirent sur plusieurs années et épuisent les amortisseurs classiques. Le matelas de liquidités se vide, les obligations peuvent perdre en même temps que les actions, et l'or peut rester endormi. Le titulaire de cette ligne est le plus étrange des actifs défensifs : les managed futures, c'est-à-dire les stratégies de **suivi de tendance** (trend following). Ce sont des programmes systématiques qui, sur des dizaines de marchés à terme (indices actions, taux, devises, matières premières), achètent mécaniquement ce qui monte et vendent à découvert ce qui baisse.

Le résultat est unique dans la panoplie. C'est le seul défensif à **espérance positive** de long terme, et ses meilleures années sont précisément les pires années des autres (2008, +20-30 % ; 2022, l'année qui a tout cassé, la meilleure de l'histoire de la catégorie). Le prix l'est tout autant : complexité, frais réels, dispersion entre gérants, et des « hivers » de plusieurs années où la stratégie déçoit pendant que les actions s'envolent. C'est le test de patience le plus dur du portefeuille.

Cet article donne le mécanisme et les preuves (un siècle de données), la raison structurelle de son profil de crise, le dossier spécifique du rentier, la mise en œuvre UCITS pour un Français (le point délicat), les pièges, et la dose.

::: cle L'idée en une phrase
Les grands mouvements de marché ne sont pas instantanés. Ils **durent** : sous-réaction initiale, accélération progressive, capitulations, sur des mois et parfois des années ([[regimes-de-marche]]). Une règle qui se contente de suivre la direction des derniers mois capte donc une part de chaque grand mouvement, à la hausse comme à la baisse. Elle se retrouve mécaniquement positionnée dans le sens de la crise quand la crise est longue. Le trend n'anticipe rien. Il est simplement du bon côté de tout ce qui persiste, et les catastrophes persistent.
:::

## Le mécanisme, concrètement

Un programme de trend type opère en quelques étapes simples. Il suit 50 à 100 marchés à terme liquides (S&P, Bund, pétrole, or, euro-dollar, blé...) et mesure la tendance de chacun par des signaux simples, une moyenne mobile ou un rendement sur 3-12 mois. Il se positionne à l'achat (long) sur les marchés en tendance haussière, à la vente (short) sur les baissiers. Il dimensionne chaque position à l'inverse de sa volatilité : c'est le vol targeting, qui fait contribuer chaque marché à parts égales au risque, le levier des futures le rendant possible sans emprunt. Et il révise en continu.

Trois propriétés découlent de cette recette. La première est la **symétrie** : gagner dans les baisses durables est aussi naturel que gagner dans les hausses, un cas unique parmi vos actifs. La deuxième est le caractère multi-classe. En 2022, l'essentiel des gains venait des positions vendeuses sur les obligations et des positions acheteuses sur les matières premières et le dollar, pas des actions ; la stratégie va chercher la tendance là où elle se trouve ([[obligations-en-retrait]]). La troisième est le **risque contrôlé** : le vol targeting vise 10-15 % de volatilité constante, soit un profil de risque proche de celui des actions, mais piloté.

**Les preuves, sur un siècle.** L'anomalie est documentée comme peu d'autres. Moskowitz, Ooi et Pedersen (« Time Series Momentum », 2012) l'établissent sur 58 marchés depuis 1965. Hurst, Ooi et Pedersen (« A Century of Evidence on Trend-Following Investing », AQR) la reconstruisent depuis 1880 : rendement positif dans **chaque** décennie, y compris pendant 1929-1939, 1970-1979 et 2000-2009, avec une corrélation aux actions et aux obligations proche de zéro. Les explications tiennent la route sur le plan théorique. La sous-réaction comportementale à l'information (l'ancrage), les flux de couverture et de gestion du risque qui amplifient les mouvements déjà entamés, les contraintes institutionnelles : rien de tout cela ne doit disparaître vite. La prime reste connue et arbitrée, et elle s'est probablement comprimée ; les estimations modernes prudentes la situent à 2-4 % réels bruts pour un programme diversifié, avant frais ([[rendements-attendus]]).

**Le « crisis alpha » et sa condition.** Pourquoi la stratégie brille-t-elle dans les crises ? Parce que les grandes crises sont des tendances. En 2008, sur dix-huit mois de baisse, le trend est vendeur d'actions et acheteur d'obligations dès l'automne. Même schéma en 1973-74 et en 2000-2002. En 2022, la tendance la plus large de l'histoire récente, il est vendeur d'obligations et acheteur d'énergie et de dollar, pour +20 à +30 % sur les indices CTA pendant que le 60/40 vivait sa pire année ([[regimes-de-marche]]). La condition est symétrique : la crise doit **durer**. Le krach éclair de 2020 en est le contre-exemple canonique. Un mois de chute, une reprise en V, et le trend n'a pas eu le temps de se retourner : −5 à 0 %, ni défense ni désastre. D'où sa place exacte dans la table. Il est titulaire des régimes **longs** et reste une doublure inutile contre les chocs courts, qui sont le travail de la duration et du cash ([[actifs-defensifs]]).

::: science Le dossier du rentier
Pour la décumulation, retenez trois faits. Un. Les simulations d'allocation convergent (ERN volet 63 sur le momentum, les études AQR et Man sur les mélanges, Portfolio Charts sur les variantes accessibles) : ajouter 10-15 % de trend à un portefeuille actions-obligations améliore le SWR des pires millésimes de l'ordre de 0,2-0,4 point et raccourcit nettement les drawdowns réels. L'effet se concentre exactement là où le rentier meurt, dans les régimes de 1966-1981 ([[etude-trinity]]). Deux. La source de cette amélioration est **propre** : ni levier caché ni beta déguisé, mais de la décorrélation à espérance positive, le seul « free lunch » partiel du menu défensif. Trois. L'avertissement d'implémentation domine tout. Ces résultats reposent sur des indices ou des programmes nets de frais raisonnables. La même stratégie à 2/20 avec un gérant médiocre rend le dossier négatif. Ici plus qu'ailleurs, le véhicule est la thèse.
:::

## La mise en œuvre française : le point délicat

C'est la brique la plus difficile à acheter proprement pour un particulier européen, et il faut le dire sans détour.

**Ce qui existe.** D'abord des **fonds** UCITS de trend systématique, gérés par les grandes maisons quantitatives : des programmes historiques déclinés en format UCITS, trend pur diversifié, à liquidité quotidienne. Leurs frais fixes tournent autour de 0,7-1,5 %/an, parfois assortis d'une commission de performance, à éviter quand une part « frais fixes seuls » existe. Plus récemment sont apparus des fonds et ETF de **réplication**, qui copient le positionnement agrégé de l'industrie CTA à frais réduits ; c'est l'approche popularisée aux États-Unis par DBMF, en cours d'arrivée en format UCITS. Le point de contrôle est universel : comparer le véhicule au **SG Trend Index**, l'indice de référence des dix grands programmes. Un bon véhicule le suit à quelques points près ; un mauvais fait autre chose sous le même nom.

**Ce qu'on évite.** Les « alternatifs multi-stratégies » vendus comme équivalents (souvent du beta déguisé et des frais), les certificats et produits structurés sur indices CTA (risque émetteur et marges opaques), et les CTA à 2/20 accessibles via des enveloppes exotiques. On évite aussi l'auto-réplication, c'est-à-dire faire son trend soi-même sur les futures. C'est théoriquement possible, mais c'est un métier en pratique : la discipline d'exécution quotidienne sur 50 marchés n'est pas un projet de retraite ([[psychologie-du-retrait]]).

**Logement et fiscalité.** Le CTO uniquement : les fonds ne sont pas éligibles au PEA, et l'assurance-vie ne les propose que rarement, aux frais de contrat près. Les plus-values relèvent du PFU ([[flat-tax-et-imposition]]). La brique se loge donc dans le budget CTO, en concurrence avec l'or ETC. C'est un argument de plus pour partager la ligne des régimes hostiles entre les deux ([[or-en-retrait]]).

::: astuce Tester la dose avant de l'acheter
La brique se teste en simulation. La page FIRE embarque des historiques de managed futures reconstruits : des backcasts de suivi de tendance sur données longues, calibrés en volatilité. L'A/B se joue comme pour l'or. Composez 10 % de trend contre la même part en obligations, puis lisez le stress, la décennie perdue, le broad-sample et les millésimes inflationnistes. Le profil attendu : un scénario central quasi inchangé, des queues raccourcies, et surtout un drawdown réel maximal qui recule sur les trajectoires médiocres ([[utiliser-la-page-fire]], [[lire-un-fan-chart]]).
:::

## Les pièges, par ordre de mortalité

**L'abandon pendant l'hiver.** C'est le piège n° 1, de très loin. Le trend connaît des **saisons sèches** de plusieurs années. La période 2011-2019, la « CTA winter » (marchés sans tendance, retournements incessants, environ 0 % cumulé pendant que les actions triplaient), a fait capituler la majorité des détenteurs, juste avant 2022. La douleur psychologique de l'écart est alors maximale : huit ans à payer une assurance en regardant les voisins s'enrichir. La parade est la même que partout, mais en plus fort. Une dose écrite, une thèse écrite (« cette poche perd ou stagne la plupart des années ; elle existe pour 1973, 2008, 2022 »), et l'interdiction de la juger sur une période sans grand régime ([[construire-son-plan]]).

**Le mauvais véhicule.** La dispersion entre gérants de trend est énorme : des dizaines de points d'écart sur une année de crise, car le signal, l'univers et la vitesse diffèrent d'un gérant à l'autre. D'où le recours à la réplication d'indice ou aux très grands programmes diversifiés, et au contrôle par le SG Trend. Ne détenez pas un petit CTA : détenez l'industrie ou son cœur.

**Le contresens sur le rôle.** Deux erreurs symétriques. Acheter du trend « pour la performance » mène à une déception garantie, car l'espérance nette est modeste. Le vendre après une année de crise gagnée « pour prendre les profits » est tout aussi fautif : c'est **précisément** le moment de rééquilibrer vers les actions massacrées. L'assurance vient de payer ; on encaisse et on se réarme, selon la même logique de salaire par rééquilibrage que pour l'or ([[or-en-retrait]]).

**Le sur-dosage d'enthousiasme.** Après 2022, la tentation des 25-30 % est réelle. Mais la stratégie reste un programme systématique aux queues propres et aux hivers longs, avec un vrai risque de modèle, car la prime peut se comprimer davantage. La dose raisonnée est de 5-15 %, comme l'or, et pour les mêmes raisons épistémiques ([[portefeuilles-tous-temps]]).

::: exemple Dix pour cent de trend, à l'épreuve
Plan : 1,5 M€, 51 000 €/an, corridor Vanguard, 45 ans. Variante A : 65 % actions / 25 % obligations / 10 % or. Variante B : 65 / 20 / 7,5 % or / 7,5 % trend (fonds UCITS à frais fixes, contrôlé SG Trend). Verdicts types : central 3,8 → 3,7 % (rien, comme il se doit), stress 5,9 → 5,1 %. Pour la décennie perdue, le drawdown réel médian des mauvaises trajectoires passe de −31 % à −26 %, et la §02 voit 1973 et 2000 s'adoucir, le trend étant le seul actif du plan à avoir « gagné » les deux. Richesse médiane : −2 %. La clause écrite au plan tient en une phrase : « poche trend : 7,5 %, jugée **uniquement** sur les années de régime (l'hiver est le fonctionnement normal), rééquilibrée aux bandes comme le reste ». Sans cette clause, ne pas acheter : l'actif sans la patience est une donation différée au marché.
:::

## L'essentiel à retenir

- Le trend suit la direction de 50-100 marchés à terme, à l'achat comme à la vente, à risque ciblé. Il est mécaniquement du bon côté de tout ce qui **dure**, et les grandes crises durent : 2008, 1973, 2022 (sa meilleure année, celle où tout le reste cassait).
- C'est le seul défensif à espérance positive (2-4 % réels bruts estimés), à corrélation proche de zéro, documenté sur un siècle (positif chaque décennie depuis 1880). Il tient la ligne « régimes longs » de la table des défenses, et reste inutile contre les chocs courts (2020).
- Pour le rentier, 10-15 % améliorent les pires millésimes de 0,2-0,4 point de SWR et raccourcissent les drawdowns. L'effet agit exactement là où les plans meurent, à condition d'un véhicule propre.
- Mise en œuvre française : fonds UCITS de trend à frais fixes (0,7-1,5 %) ou réplication d'indice, contrôlés contre le SG Trend Index, en CTO. On évite les multi-stratégies déguisés, le 2/20 et l'auto-réplication.
- Le piège mortel est comportemental : les hivers de 5-8 ans (2011-2019) font capituler juste avant la récolte. Dose écrite, thèse écrite, jugement sur les seuls régimes. Sans cette discipline, cette brique n'est pas pour vous.

---

## Pour aller plus loin

- Moskowitz, Ooi & Pedersen, « Time Series Momentum » (2012) ; Hurst, Ooi & Pedersen, « A Century of Evidence on Trend-Following Investing » (AQR) : les preuves.
- Man Institute et AQR : les analyses « trend et crises » (crisis alpha) et les mélanges avec un portefeuille classique.
- Le SG Trend Index (Société Générale) : le référentiel public de l'industrie, pour contrôler tout véhicule.
- Early Retirement Now, volet 63 (momentum et retrait) ([[serie-ern]]).
- Dans ce livre : [[actifs-defensifs]] (la ligne servie), [[regimes-de-marche]] (pourquoi les crises sont des tendances), [[portefeuilles-tous-temps]] (le Dragon et la place du trend), [[or-en-retrait]] (le partage de la ligne inflation).
