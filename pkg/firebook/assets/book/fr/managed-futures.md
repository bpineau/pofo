# Managed futures et suivi de tendance : la diversification qui travaille dans les crises

Dans la table des défenses ([[actifs-defensifs]]), une ligne restait à pourvoir : les **régimes longs**, ces marchés baissiers ou inflationnistes de plusieurs années qui épuisent les amortisseurs classiques (le buffer se vide, les obligations peuvent perdre avec les actions, l'or peut dormir). Le titulaire de cette ligne est le plus étrange des actifs défensifs : les managed futures, c'est-à-dire les stratégies de **suivi de tendance** (trend following) : des programmes systématiques qui, sur des dizaines de marchés à terme (indices actions, taux, devises, matières premières), achètent ce qui monte et vendent À **découvert** ce qui baisse, mécaniquement.

Le résultat est unique dans la panoplie : le seul défensif à **espérance positive** de long terme, dont les meilleures années sont précisément les pires années des autres (2008 : +20-30 % ; 2022, l'année qui a tout cassé : la meilleure de l'histoire de la catégorie). Le prix est tout aussi unique : complexité, frais réels, dispersion entre gérants, et des « hivers » de plusieurs années où la stratégie déçoit pendant que les actions s'envolent : le test de patience le plus dur du portefeuille.

Cet article donne le mécanisme et les preuves (un siècle de données), la raison structurelle de son profil de crise, le dossier spécifique du rentier, la mise en œuvre UCITS pour un Français (le point délicat), les pièges, et la dose.

::: cle L'idée en une phrase
Les grands mouvements de marché ne sont pas instantanés : ils **durent** (sous-réaction initiale, ralliement progressif, capitulations : des mois, parfois des années, [[regimes-de-marche]]) : une règle qui se contente de suivre la direction des derniers mois capte une part de chaque grand mouvement : à la hausse comme à la baisse : et se retrouve mécaniquement positionnée **dans** le sens de la crise quand la crise est longue. Le trend n'anticipe rien : il est simplement du bon côté de tout ce qui persiste, et les catastrophes persistent.
:::

## Le mécanisme, concrètement

Un programme de trend type opère ainsi : sur 50 à 100 marchés à terme liquides (S&P, Bund, pétrole, or, euro-dollar, blé...), il mesure la tendance de chacun (signaux simples : moyenne mobile, rendement sur 3-12 mois) ; il se positionne **long** sur les marchés en tendance haussière, **short** sur les baissiers ; il dimensionne chaque position en **inverse** de sa volatilité (vol targeting : chaque marché contribue également au risque, le levier des futures rendant cela possible sans emprunt) ; et il révise en continu. Trois propriétés découlent de cette recette : la **symétrie** (gagner dans les baisses durables est aussi naturel que dans les hausses : unique parmi vos actifs) ; la **multi-classe** (en 2022, l'essentiel des gains venait des shorts **obligataires** et des longs matières premières et dollar : pas des actions : la stratégie pêche là où la tendance est, [[obligations-en-retrait]]) ; et le **risque contrôlé** (le vol targeting vise 10-15 % de volatilité constante : c'est un profil de risque d'actions, piloté).

**Les preuves, sur un siècle.** L'anomalie est documentée comme peu d'autres : Moskowitz, Ooi et Pedersen (« Time Series Momentum », 2012) l'établissent sur 58 marchés depuis 1965 ; Hurst, Ooi et Pedersen (« A Century of Evidence on Trend-Following Investing », AQR) la reconstruisent depuis 1880 : rendement positif dans **chaque** décennie, y compris à travers 1929-1939, 1970-1979 et 2000-2009, avec une corrélation aux actions ~0 et aux obligations ~0. Les explications tiennent la route théoriquement : sous-réaction comportementale aux informations (ancrage), flux de couverture et de gestion du risque qui amplifient les mouvements entamés, contraintes institutionnelles : rien qui doive disparaître vite, même si la prime, connue et arbitrée, s'est probablement comprimée (les estimations modernes prudentes : 2-4 % réels bruts au niveau d'un programme diversifié, avant frais, [[rendements-attendus]]).

**Le « crisis alpha », et sa condition.** Pourquoi la stratégie brille-t-elle dans les crises ? Parce que les grandes crises **sont** des tendances : 2008 (dix-huit mois de baisse : le trend est short actions et long obligations dès l'automne), 1973-74, 2000-2002, et 2022 (la tendance la plus large de l'histoire récente : short obligations, long énergie et dollar : +20 à +30 % pour les indices CTA pendant que le 60/40 vivait sa pire année, [[regimes-de-marche]]). La condition, symétrique : la crise doit **durer**. Le krach éclair de 2020 (un mois de chute, reprise en V) est le contre-exemple canonique : le trend n'a pas eu le temps de se retourner : −5 à 0 % : ni défense ni désastre. D'où la place exacte dans la table : titulaire des régimes **longs**, doublure inutile contre les chocs courts (c'est le travail de la duration et du cash, [[actifs-defensifs]]).

::: science Le dossier du rentier
Pour la décumulation, trois faits. **Un** : les simulations d'allocation (ERN Part 63 sur le momentum, les études AQR et Man sur les mélanges, Portfolio Charts sur les variantes accessibles) convergent : 10-15 % de trend ajoutés à un portefeuille actions-obligations améliorent le SWR des pires millésimes de l'ordre de 0,2-0,4 point et raccourcissent nettement les drawdowns réels : l'effet est concentré exactement là où le rentier meurt (les régimes 1966-1981, [[etude-trinity]]). **Deux** : la source de l'amélioration est **propre** : ni levier caché ni beta déguisé : de la décorrélation à espérance positive : le seul « repas gratuit » partiel du menu défensif. **Trois** : l'avertissement d'implémentation domine tout : ces résultats utilisent des indices ou programmes **nets** de frais **raisonnables** : la même stratégie à 2/20 avec un gérant médiocre rend le dossier négatif : ici plus qu'ailleurs, le véhicule **est** la thèse.
:::

## La mise en œuvre française : le point délicat

C'est la brique la plus difficile à acheter proprement pour un particulier européen, et il faut le dire sans détour.

**Ce qui existe.** Des **fonds** UCITS de trend systématique gérés par les grandes maisons quantitatives (les programmes historiques déclinés en format UCITS : trend pur diversifié, liquidité quotidienne), avec des frais fixes de l'ordre de 0,7-1,5 %/an (parfois une commission de performance : à éviter quand une part « frais fixes seuls » existe) ; et, plus récemment, des fonds et ETF de **réplication** (qui copient le positionnement agrégé de l'industrie CTA à frais réduits : l'approche popularisée aux États-Unis par DBMF, en cours d'arrivée en format UCITS). Le point de contrôle universel : comparer le véhicule au **SG Trend Index** (l'indice de référence des dix grands programmes) : un bon véhicule le suit à quelques points près ; un mauvais fait autre chose sous le même nom.

**Ce qu'on évite.** Les « alternatifs multi-stratégies » vendus comme équivalents (souvent du beta déguisé et des frais), les certificats et produits structurés sur indices CTA (risque émetteur + marges opaques), et les CTA à 2/20 accessibles via des enveloppes exotiques. Et l'auto-réplication (faire son trend soi-même sur futures) : théoriquement possible, pratiquement un métier : la discipline d'exécution quotidienne sur 50 marchés N'**est pas** un projet de retraite ([[psychologie-du-retrait]]).

**Logement et fiscalité.** CTO uniquement (fonds non éligibles PEA ; rares UC d'assurance-vie, aux frais de contrat près) : PFU sur les plus-values ([[flat-tax-et-imposition]]). La brique se loge donc en concurrence avec l'or ETC dans le budget CTO : un argument de plus pour partager la ligne des régimes hostiles entre les deux ([[or-en-retrait]]).

**Dans pofo** : la brique existe : le catalogue embarque des historiques de managed futures reconstruits (backcasts TSMOM sur données longues, calibrés en volatilité) : l'A/B se joue comme pour l'or : composez 10 % de trend contre la même part en obligations, et lisez le stress, la décennie perdue, le broad-sample et les millésimes inflationnistes de la §02 : le profil attendu : central quasi inchangé, queues raccourcies, et : regardez-le : le drawdown réel maximal des trajectoires médiocres qui recule ([[utiliser-la-page-fire]], [[lire-un-fan-chart]]).

## Les pièges, par ordre de mortalité

**L'abandon pendant l'hiver.** Le piège n° 1, de très loin. Le trend a des **saisons sèches** pluriannuelles : 2011-2019, la « CTA winter » (marchés sans tendance, retournements incessants : ~0 % cumulé pendant que les actions triplaient), a fait capituler la majorité des détenteurs : juste avant 2022. La tracking-error psychologique est maximale : huit ans à payer une assurance en regardant les voisins s'enrichir. La parade est la même que partout, en plus fort : une dose écrite, une thèse écrite (« cette poche perd ou stagne la plupart des années ; elle existe pour 1973, 2008, 2022 »), et l'interdiction de la juger sur une période sans grand régime ([[construire-son-plan]]).

**Le mauvais véhicule.** La dispersion entre gérants de trend est énorme (des dizaines de points sur une année de crise : le signal, l'univers et la vitesse diffèrent) : d'où la réplication d'indice ou les très grands programmes diversifiés, et le contrôle SG Trend. Ne détenez pas **un** petit CTA : détenez l'industrie ou son cœur.

**Le contresens sur le rôle.** Acheter du trend « pour la performance » (déception garantie : l'espérance nette est modeste) ou le vendre après une année de crise gagnée « pour prendre les profits » (c'est **précisément** le moment du rééquilibrage vers les actions massacrées : l'assurance vient de payer, on encaisse et on re-arme, [[or-en-retrait]] : même logique de salaire par rééquilibrage).

**Le sur-dosage d'enthousiasme.** Après 2022, la tentation des 25-30 % : mais la stratégie reste un programme systématique à queues propres **mais** à hivers longs, avec un risque de modèle réel (la prime peut se comprimer davantage). La dose raisonnée : 5-15 %, comme l'or, et pour les mêmes raisons épistémiques ([[portefeuilles-tous-temps]]).

::: exemple Dix pour cent de trend, à l'épreuve
Plan : 1,5 M€, 51 000 €/an, corridor Vanguard, 45 ans. A : 65 % actions / 25 % obligations / 10 % or. B : 65 / 20 / 7,5 % or / 7,5 % trend (fonds UCITS à frais fixes, contrôlé SG Trend). Verdicts pofo type : central 3,8 → 3,7 % (rien, comme il se doit) ; stress 5,9 → 5,1 % ; décennie perdue : le drawdown réel médian des mauvaises trajectoires passe de −31 % à −26 %, et la §02 voit 1973 **et** 2000 s'adoucir (le trend est le seul actif du plan à avoir « gagné » les deux). Richesse médiane : −2 %. La clause écrite au plan : « poche trend : 7,5 %, jugée **uniquement** sur les années de régime (l'hiver est le fonctionnement normal), rééquilibrée aux bandes comme le reste ». Sans cette clause, ne pas acheter : l'actif sans la patience est une donation différée au marché.
:::

## L'essentiel à retenir

- Le trend suit la direction de 50-100 marchés à terme, long et short, à risque ciblé : il est mécaniquement du bon côté de tout ce qui **dure** : et les grandes crises durent : 2008, 1973, 2022 (sa meilleure année, l'année où tout le reste cassait).
- C'est le seul défensif à espérance positive (2-4 % réels bruts estimés), corrélation ~0, documenté sur un siècle (positif chaque décennie depuis 1880) : le titulaire de la ligne « régimes longs » de la table des défenses ; inutile contre les chocs courts (2020).
- Pour le rentier : 10-15 % améliorent les pires millésimes de 0,2-0,4 point de SWR et raccourcissent les drawdowns : l'effet vit exactement où les plans meurent : à condition d'un véhicule propre.
- Mise en œuvre française : fonds UCITS de trend à frais fixes (0,7-1,5 %) ou réplication d'indice, contrôlés contre le SG Trend Index, en CTO ; éviter multi-stratégies déguisés, 2/20 et auto-réplication.
- Le piège mortel est comportemental : les hivers de 5-8 ans (2011-2019) font capituler juste avant la récolte : dose écrite, thèse écrite, jugement sur les seuls régimes : sans cette discipline, cette brique n'est pas pour vous.

---

## Pour aller plus loin

- Moskowitz, Ooi & Pedersen, « Time Series Momentum » (2012) ; Hurst, Ooi & Pedersen, « A Century of Evidence on Trend-Following Investing » (AQR) : les preuves.
- Man Institute et AQR : les analyses « trend et crises » (crisis alpha) et les mélanges avec un portefeuille classique.
- Le SG Trend Index (Société Générale) : le référentiel public de l'industrie, pour contrôler tout véhicule.
- Early Retirement Now, Part 63 (momentum et retrait) ([[serie-ern]]).
- Dans ce livre : [[actifs-defensifs]] (la ligne servie), [[regimes-de-marche]] (pourquoi les crises sont des tendances), [[portefeuilles-tous-temps]] (le Dragon et la place du trend), [[or-en-retrait]] (le partage de la ligne inflation).
