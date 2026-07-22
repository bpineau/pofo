# L'allocation actions/obligations en retrait

« Combien d'actions ? » C'est la première question que tout le monde pose sur le portefeuille de retrait. Depuis Bengen, la recherche y répond avec une constance remarquable. La courbe du taux de retrait soutenable en fonction de la part d'actions dessine un **plateau**. Il est étonnamment large et étonnamment plat, entre environ 50 et 80 % d'actions, et il plonge des **deux** côtés. Trop peu d'actions est plus dangereux que trop. Voilà qui contredit frontalement l'intuition « retraite = prudence = obligations » héritée du conseil bancaire.

Cet article établit ce résultat et son mécanisme, c'est-à-dire pourquoi il **faut** de la croissance pour financer 40 ans de retraits réels. Il décrit ce qui se passe aux deux bords du plateau. Il montre comment l'horizon, la couverture du plancher et la règle de retrait déplacent votre position optimale à l'intérieur. Il fait le point sur ce que la recherche récente a ajouté et relativisé, avec Anarkulova-Cederburg et le débat « 100 % actions ». Il explique enfin comment tester tout cela sur votre plan dans un simulateur, où les poids se déplacent en direct. La méthode d'ensemble qui l'encadre, concevoir par les risques plutôt que par les actifs, est dans [[concevoir-un-portefeuille]]. Cet article ouvre la partie portefeuille. Les suivants raffinent chaque brique ([[obligations-en-retrait]], [[or-en-retrait]], [[portefeuilles-tous-temps]]) et la dimension temporelle ([[glidepaths]]).

::: cle Le résultat central
Sur toutes les données et tous les modèles, le taux de retrait soutenable à long horizon suit la même courbe en fonction de la part d'actions. Il reste faible de 0 à 30 % d'actions, car les obligations seules ne battent pas l'inflation augmentée des retraits. Il devient maximal et **plat** de ~50 à ~80 %. Il fléchit légèrement au-delà de 90 %, quand la volatilité pure commence à coûter plus que sa prime ne rapporte. Deux conséquences en découlent. L'erreur grave est d'être sous le plateau, pas au-dessus. Et à l'intérieur du plateau, le choix fin, 60/40 contre 75/25, compte peu pour la ruine. Il se joue sur d'autres critères, la profondeur des creux que vous pouvez vivre et votre règle de retrait.
:::

::: figure allocation-plateau
Le taux de retrait soutenable en fonction de la part d'actions (ordres de grandeur, horizon long). Trop peu d'actions et l'érosion l'emporte ; trop, et la volatilité coûte plus que sa prime. Entre les deux, un plateau large qui pardonne l'imprécision.
:::

## Pourquoi il faut de la croissance : l'arithmétique du plateau

Reprenons le mécanisme à la racine. Un plan de retrait à 3,5 % sur 45 ans doit servir des retraits **réels**, donc indexés, pendant des décennies. Il lui faut un moteur qui produise durablement plus de 3,5 % réels par an, frein de volatilité (le *drag*) et séquence compris ([[rendements-arithmetiques-geometriques]]). Sur longue période, les obligations nominales de qualité rapportent 1 à 2 % réels ([[rendements-attendus]]). Un portefeuille 20/80 a donc une espérance géométrique réelle de ~2 %. Il finance mathématiquement moins de 2,5 % de retrait perpétuel. Sous 3,5 %, il s'épuise **non** par malchance mais par construction, lentement, sûrement, sans krach. C'est le mode de défaillance « érosion » ([[lire-un-fan-chart]]). C'est aussi le bord gauche du plateau. Voilà pourquoi la grille de Trinity montrait déjà le 25/75 échouer une fois sur trois là où le 75/25 tenait ([[etude-trinity]]).

Le bord droit est plus subtil. De 80 à 100 % d'actions, l'espérance monte encore, mais deux coûts croissent plus vite. Le premier est le *drag* : σ²/2, passer de 12 à 16 % de volatilité coûte ~0,6 point de composition. Le second, plus lourd, est la **séquence**. Sans aucun amortisseur, un krach précoce de 50 % frappe le plan de plein fouet dans sa fenêtre fragile ([[sequence-des-rendements]]). Résultat net, le SAFEMAX du 100 % actions est légèrement **inférieur** à celui du 70/30 dans les données historiques américaines, avec des pires cas nettement plus profonds. Le plateau existe parce que ces deux forces, besoin de croissance et coût de la volatilité, s'équilibrent sur une plage large. C'est une chance, car la nature du problème pardonne l'imprécision sur ce paramètre.

Dans ce cadre, les obligations ont un rôle précis et borné. Elles ne sont pas là pour « rapporter » mais pour rendre trois services. Le premier est d'**amortir** les krachs d'actions. En régime désinflationniste, leur corrélation négative fait remonter leur prix quand les actions plongent ; c'est l'amortisseur qui a si bien marché de 2000 à 2021. Le deuxième est de **fournir** la liquidité des retraits pendant les creux. On vend l'obligation qui a monté, pas l'action qui a baissé, c'est le prélèvement-rééquilibrage ([[retrait-fixe-bengen]]). Le troisième est de **réduire** la profondeur des drawdowns à un niveau tenable pour votre règle et vos nerfs. Leur talon d'Achille est l'inflation, qui casse les trois services à la fois. C'est le sujet de [[obligations-en-retrait]] et la raison d'être des briques suivantes de cette partie ([[regimes-de-marche]]).

## Se placer dans le plateau : les trois curseurs qui comptent

Puisque la ruine discrimine peu entre 55/45 et 80/20, qu'est-ce qui doit décider ? Trois choses, dans l'ordre.

**1. La profondeur de creux vivable.** C'est le critère le plus concret. Le drawdown réel à prévoir une ou deux fois par décennie vaut environ −0,55 × (part d'actions) pour un portefeuille diversifié : un 60/40 encaisse ~−33 %, un 80/20 ~−44 %, le 100 % −55 % et plus. Restent deux tests d'admissibilité. Votre **règle** le tient-elle ? C'est le test « actions −50 % » des règles proportionnelles ([[vpw]]) et le plancher des guardrails ([[guyton-klinger]]). Et vous, le tenez-vous ? L'expérience d'accumulation ne préjuge de rien : subir −40 % quand on **vit** du portefeuille est une autre épreuve ([[psychologie-du-retrait]]). La part d'actions maximale admissible sort de ces deux tests, pas d'un optimiseur.

**2. La couverture du plancher.** C'est le grand modulateur, déjà rencontré partout. Un plancher couvert par une pension ou une rente ([[rentes-et-annuites]], [[retraite-legale]]) libère le portefeuille vers le haut du plateau. C'est la logique TPAW : la pension actualisée est une grosse position obligataire implicite ([[amortissement-abw]]), donc le portefeuille visible peut porter davantage d'actions. À l'inverse, un plancher financé à 100 % par le portefeuille pendant vingt ans tire vers le milieu-bas du plateau et vers les protections dédiées ([[cash-buffer]], [[glidepaths]]).

**3. La règle de retrait.** Les règles proportionnelles et actuarielles ([[pourcentage-fixe]], [[amortissement-abw]]) encaissent structurellement mieux la volatilité, car le retrait s'ajuste. Elles tolèrent le haut du plateau. Le fixe indexé, qui encaisse tout dans le capital, préfère le milieu, 60-70 %. Règle et allocation se choisissent **ensemble**. C'est le sens de la frontière §06, qui les croise ([[choisir-sa-strategie]]).

::: science Le débat « 100 % actions » (Anarkulova-Cederburg), remis à sa place
Le papier « Beyond the Status Quo » (2023) de l'équipe Cederburg a fait grand bruit. Dans leur échantillon mondial, un portefeuille 100 % actions **diversifié internationalement** (50 % domestique, 50 % international) domine les mélanges actions-obligations pour le retraité. La raison : sur un siècle mondial, les obligations nominales se font périodiquement détruire par l'inflation, exactement quand les actions locales souffrent ([[anarkulova-cederburg]], [[regimes-de-marche]]). La lecture honnête est nuancée. Le résultat attaque, à raison, **les obligations nominales domestiques** comme actif « sûr » de long terme, pas l'idée d'amortisseur ; la diversification internationale des actions y fait le travail défensif. Les répliques d'ERN et de Kitces soulignent la violence des chemins 100 % actions : des drawdowns réels de −60 à −80 % dans les queues, où aucun test d'admissibilité humain ne passe. Elles rappellent aussi la dépendance au traitement des périodes de guerre. La synthèse utilisable tient en deux points. Le bord droit du plateau est moins pénalisé qu'on ne le disait **si** les actions sont mondialement diversifiées ([[diversification-internationale]]). Et la vraie leçon n'est pas « 100 % actions » mais « la poche défensive ne doit pas être que des obligations nominales ». D'où les linkers, l'or et la duration courte, que détaille la suite de cette partie ([[obligations-indexees]], [[or-en-retrait]], [[actifs-defensifs]]).
:::

## Au-delà du ratio : ce que « actions » et « obligations » doivent contenir

Le ratio ne suffit pas ; le contenu des deux poches déplace les résultats autant que le ratio lui-même.

**La poche actions.** Mondiale, à capitalisation large, la plus simple possible ([[etf-ucits-europeens]]). Le biais domestique est le premier défaut à corriger ([[diversification-internationale]]). Les inclinaisons factorielles (la décote de valorisation *value*, les *small caps*, [[facteurs-fama-french]]) sont un raffinement de second ordre, légitime mais optionnel. Ce qui est exclu, c'est la concentration, qu'il s'agisse de titres vifs, de secteurs ou d'un pays unique : le retraité n'a pas le temps de moyenner un risque idiosyncratique.

**La poche obligataire.** C'est elle qui demande le plus de soin, et elle a son article ([[obligations-en-retrait]]). Trois décisions en avant-première. D'abord la **qualité** : obligations d'État et *investment grade* uniquement. Le high yield est une action déguisée, qui fait défaut exactement dans les krachs et ne rend aucun des trois services. Ensuite la **duration** : intermédiaire, 5-8 ans, le compromis standard entre pouvoir amortisseur et risque de taux. L'année 2022 a rappelé ce que coûte la duration longue en régime inflationniste. Enfin la part **indexée** : les linkers couvrent le service que les nominales ne rendent pas ([[obligations-indexees]]).

**Le rééquilibrage**, enfin, est le troisième acteur silencieux. C'est lui qui matérialise le bénéfice de l'allocation : vendre ce qui a monté pour financer les retraits, racheter ce qui a baissé aux bornes. Le consensus pratique procède par bandes, ±5 points absolus autour de la cible, plutôt qu'au calendrier strict, et utilise les retraits comme premier outil de retour à la cible. La fréquence exacte est du troisième ordre (ERN volet 39). L'ennemi n'est pas le réglage fin, c'est la non-exécution dans les krachs. D'où le rééquilibrage écrit dans le plan, comme la règle de retrait ([[construire-son-plan]]).

## Tester sur votre plan : le mode portefeuille

Tout ce qui précède se vérifie en direct. En mode portefeuille, les poids de chaque ligne se déplacent au curseur et **toute** la page se recalcule : ruine sous les quatre modèles, cônes et décennie décisive compris ([[utiliser-la-page-fire]]). Voici la séance type « allocation ». Partez de votre allocation réelle. Faites glisser la part actions de 40 à 90 % par pas de 10, et notez à chaque pas la ruine centrale et la ruine broad-sample. Vous **verrez** votre plateau, souvent plus plat qu'attendu, et ses deux bords. Fixez ensuite la part admissible par vos deux tests, creux vivable et règle. Vérifiez enfin la §03, la décennie décisive, qui départage les hauts du plateau, et la §02, les millésimes réels : le 80/20 dans 1966 et 2000 est à regarder en face avant de signer. Une nuance de lecture : les colonnes historiques rejouent **votre** fenêtre, favorable aux actions récentes, tandis que le broad-sample juge l'allocation sur le siècle. L'arbitrage entre les deux est le même que partout ([[historique-vs-parametrique]]).

::: astuce Ouvrir la page sur son portefeuille
Dans pofo, `pofo -fire portfolio.txt` ouvre la page FIRE directement sur votre portefeuille réel, chaque ligne pilotable au curseur.
:::

::: exemple Un plateau rendu visible
Le plan : 1,5 M€, 52 000 €/an avec guardrails, 45 ans, pension à l'année 16. Le balayage donne, en ruine centrale puis broad-sample : 30/70 → 11 % / 19 % (bord gauche, l'érosion), 50/50 → 5 % / 11 %, 65/35 → 4 % / 9 %, 80/20 → 4 % / 9 %, 95/5 → 5 % / 10 % (bord droit, doux car la règle est flexible). Le plateau 50-90 est manifeste. Viennent les tests d'admissibilité. Le plancher des guardrails tient jusqu'à ~75 % d'actions ; au-delà, le pire quartile de la §04 passe des années sous le plancher. Le couple, lui, s'est connu à −35 % en 2020 sans vendre, pas au-delà. La décision : 70/30, avec une poche obligataire de 20 en nominales intermédiaires et 10 en linkers, rééquilibrage à bandes ±5, écrit au plan. Le balayage a pris dix minutes. La conversation sur le creux vivable, une soirée. Les bonnes proportions, encore.
:::

## L'essentiel à retenir

- La courbe taux-soutenable/part d'actions est un plateau large (~50-80 %) qui plonge des deux côtés : l'erreur grave est le sous-investissement en actions (l'érosion certaine), pas le sur-investissement (la volatilité coûteuse).
- Les obligations ne sont pas là pour rapporter mais pour trois services : amortir, financer les retraits au creux, borner les drawdowns. Ces services sont conditionnés au régime (l'inflation les casse tous, d'où linkers, or et diversification en complément).
- Dans le plateau, trois curseurs décident : la profondeur de creux vivable (~−0,55 × part actions, à confronter à votre règle et à vos nerfs), la couverture du plancher (pension = obligation implicite, libère vers le haut), la règle de retrait (proportionnelle tolère plus d'actions que fixe).
- Le débat « 100 % actions » enseigne surtout que la poche défensive ne doit pas être 100 % obligations nominales domestiques, et que la diversification internationale des actions est elle-même défensive.
- Balayez **votre** plateau en dix minutes dans un simulateur (poids en direct, quatre modèles), fixez la part par les tests d'admissibilité, écrivez ratio, contenu et règle de rééquilibrage. Puis passez à la dimension temporelle : [[glidepaths]].

---

## Pour aller plus loin

- Bengen (1996) sur l'allocation optimale ; Trinity (1998) pour la grille par allocation ([[etude-trinity]]).
- Cederburg et al., « Beyond the Status Quo » (2023) et les réponses d'ERN et Kitces : le débat 100 % actions, dans le texte ([[anarkulova-cederburg]]).
- Early Retirement Now, volet 19-20 (allocation et glidepaths) et volet 39 (rééquilibrage) ([[serie-ern]]).
- Dans ce livre : toute la suite de la partie V, qui remplit les deux poches brique par brique.
