# Les règles CAPE : ajuster le retrait aux valorisations (ERN)

Si le niveau de valorisation au départ prédit le sort d'un millésime ([[valorisations-et-cape]]), pourquoi ne l'utiliser qu'une fois, le jour du départ ? Les règles CAPE poussent la logique au bout : le taux de retrait devient une **fonction** du CAPE courant, recalculée chaque année : on retire davantage quand les marchés sont bon marché (rendements futurs élevés), moins quand ils sont chers (rendements futurs comprimés).

C'est la stratégie que Karsten Jeske (ERN) a formalisée pour lui-même au Part 54 de sa série, celle qu'il exécute réellement depuis 2018, et à bien des égards l'aboutissement intellectuel de la famille proportionnelle : un pourcentage du portefeuille dont le w n'est plus une constante arbitraire mais une estimation du rendement soutenable de l'instant. Cet article donne la formule exacte et ses paramètres (avec les valeurs qu'ERN recommande et pourquoi), la double contra-cyclicité qui rend son revenu étonnamment plus **stable** que celui du pourcentage fixe, ses vraies difficultés (le niveau du CAPE se discute, le revenu reste variable, l'exécution exige de la foi dans les mauvaises années), et sa mise en œuvre : y compris la voie moderne dans pofo, où l'amortissement ABW ancré au CAPE en est l'incarnation propre.

::: cle La règle en une formule
Chaque année : taux de retrait = a + b × (1/CAPE), appliqué au portefeuille **courant**. 1/CAPE est l'earnings yield, l'estimation du rendement réel des actions que les prix du moment permettent ([[valorisations-et-cape]]) ; b est la part de ce rendement que vous consommez (~0,5) ; a est le socle indépendant des valorisations (~1,5-2 %). À CAPE 20 : 1,75 + 0,5 × 5 = 4,25 %. À CAPE 33 : 1,75 + 0,5 × 3 = 3,27 %. À CAPE 12 : 5,9 %. Le taux respire avec les prix : c'est un pourcentage fixe devenu conscient du monde.
:::

## La logique : consommer le rendement estimé, pas un chiffre gravé

Relisons la formule comme un raisonnement économique. Un portefeuille diversifié a, à tout instant, un rendement réel soutenable approximatif : l'earnings yield pour la poche actions, le taux réel courant pour les obligations, plus la croissance des bénéfices ([[rendements-attendus]]). La règle CAPE dit simplement : consommez une fraction prudente de **cette** estimation, réévaluée chaque année. Le terme a agrège ce qui ne dépend pas du CAPE (la croissance réelle des bénéfices, ~1,5-2 points, et la contribution des autres poches) ; le terme b × (1/CAPE) fait respirer la part actions avec les prix. Les paramètres d'ERN (a = 1,75, b = 0,5 pour un horizon de 60 ans avec préservation partielle du capital ; a = 2, b = 0,5 pour 30 ans) sortent de régressions sur les données mensuelles 1871-2016 : ce sont les valeurs qui auraient maintenu le pouvoir d'achat du capital à travers tous les millésimes, queues comprises.

Ce simple déplacement (consommer une estimation plutôt qu'une constante) achète trois propriétés remarquables.

**L'héritage proportionnel intégral.** La règle est un pourcentage du portefeuille courant : ruine du capital impossible, auto-correction face aux erreurs, contra-cyclicité des prélèvements ([[pourcentage-fixe]]).

**La double contra-cyclicité, la propriété magique.** Voici le point le plus contre-intuitif et le plus important. Sous pourcentage fixe, un krach de 30 % coupe le revenu de 30 %. Sous règle CAPE, le krach fait **aussi** baisser le CAPE, donc monter 1/CAPE, donc monter le taux w : le revenu = w × portefeuille voit ses deux facteurs bouger en sens **inverses**. Chiffrons : portefeuille −30 % pendant que le CAPE passe de 30 à 21 : w passe de 3,42 % à 4,13 % (+21 %) : le revenu ne baisse que de ~16 %, pas 30. La règle CAPE est un pourcentage **auto-lissé** : elle amortit précisément les baisses dues à la compression des multiples (les plus fréquentes), et c'est un lissage économiquement fondé, pas un artifice de moyenne mobile ([[pourcentage-fixe]]) : on consomme plus du capital quand il promet plus. Symétriquement, dans l'euphorie (portefeuille +40 %, CAPE 25 → 35), w baisse et le revenu ne monte que modérément : la règle refuse de dépenser des plus-values de bulle, exactement le comportement qu'on voudrait avoir et qu'on n'a jamais spontanément.

**La cohérence temporelle.** Le paradoxe des deux voisins ([[la-regle-des-4-pourcents]] : partis à un an d'écart avec le même portefeuille, des retraits différents à jamais) disparaît : la règle CAPE donne le même retrait à tout détenteur du même portefeuille au même moment, quelle que soit son histoire. C'est la marque des règles « sans mémoire morte » : seul l'état présent compte, comme pour l'ABW ([[amortissement-abw]]).

## Les difficultés honnêtes

**Le niveau du CAPE se discute.** Tout le débat de [[valorisations-et-cape]] (dérive comptable, buybacks, taux : le CAPE moderne « vaut » 3-8 points de moins qu'en comparaison naïve) frappe une règle qui consomme le **niveau**, pas seulement le rang. Avec les paramètres historiques d'ERN, un CAPE structurellement plus haut qu'autrefois donne des taux structurellement plus bas : prudent, mais peut-être trop. Les parades : utiliser un CAPE ajusté (total-return, ou l'excess CAPE yield qui intègre les taux), ou recalibrer a en conséquence (ERN lui-même a publié des variantes) : et accepter qu'une règle fondée sur une estimation hérite des incertitudes de l'estimation.

**Le CAPE de quoi ?** La règle canonique utilise le CAPE américain ; votre portefeuille est mondial ([[etf-ucits-europeens]]). Le CAPE monde (pondéré) est moins disponible mais publié (Barclays, Research Affiliates) ; l'approximation par le CAPE US est conservatrice (il est le plus cher) et cohérente avec son poids de 60-70 % dans les indices mondiaux : acceptable, à savoir.

**Le revenu reste variable, et l'exécution exige du cran.** L'auto-lissage amortit, il ne supprime pas : dans un vrai régime hostile où prix **et** bénéfices baissent (le CAPE peut ne pas baisser autant que les cours), le revenu descend. Et il faut exécuter la règle dans les **deux** sens : retirer 5,5 % d'un portefeuille amputé en plein 2009 demande une vraie confiance dans la formule (c'est pourtant là qu'elle a raison : c'est le moment où les rendements futurs sont les plus élevés) ; beaucoup d'utilisateurs trichent à la baisse au creux, détruisant précisément la propriété qui justifiait la règle. Comme toujours : le test d'admissibilité est le plancher ([[combien-il-vous-faut]]) : la règle CAPE convient si plancher × prudence < revenu du pire scénario, et la §04 de pofo répond à cette question.

::: attention La confusion à ne pas commettre
« Règle CAPE » désigne **ajuster le retrait** aux valorisations : pas timer le portefeuille. La règle ne vend pas d'actions à CAPE haut, ne sort pas du marché, ne fait aucun arbitrage d'allocation : elle règle uniquement le robinet de consommation. La confusion est fréquente et fatale : le timing d'allocation sur CAPE détruit de la valeur ([[valorisations-et-cape]]), le réglage du retrait sur CAPE en crée. Même indicateur, deux usages, deux verdicts opposés.
:::

## La mise en œuvre, du tableur à pofo

**À la main** : la règle tient dans une cellule de tableur. Chaque 1er janvier : lire le CAPE (site de Shiller, multpl.com), calculer a + b/CAPE, multiplier par le portefeuille, diviser par douze : le « salaire » de l'année. La gouvernance est celle d'un pourcentage : simple, auditable, exécutable par le conjoint ([[revue-annuelle]]).

**Dans pofo, la voie propre : l'ABW ancré au CAPE.** pofo n'implémente pas la formule a + b/CAPE comme règle de dépense nommée, et c'est un choix réfléchi : il implémente mieux. Cochez « Amortize over the horizon (ABW/TPAW) » **et** l'ancre CAPE (« Anchor return to today's valuation ») : l'amortissement calcule alors chaque année le paiement qui épuise le portefeuille sur l'horizon restant **au rendement impliqué par le** CAPE ([[amortissement-abw]], [[utiliser-la-page-fire]]). C'est exactement l'esprit de la règle CAPE (consommer ce que les valorisations promettent), avec deux raffinements que la formule linéaire n'a pas : l'horizon restant exact (a + b/CAPE suppose implicitement un horizon long constant) et l'actualisation des pensions futures. La règle d'ERN est la version « dos d'enveloppe » géniale ; l'ABW-CAPE en est la forme actuarielle complète : même famille, même information, mécanique aboutie.

**En hybride, la voie douce** : garder un retrait fixe amendé ([[retrait-fixe-bengen]]) mais conditionner au CAPE le taux **initial** (le jour du départ) et les **révisions** lors des grands franchissements (recalculer le retrait de référence quand le CAPE change de zone : sous 18, 18-28, au-dessus de 28). C'est la moitié du bénéfice pour le dixième de la variabilité, et une excellente stratégie de transition pour qui vient du monde Bengen.

::: exemple Dix ans de règle CAPE, années difficiles comprises
Sofia, 1,5 M€, a = 1,75, b = 0,5. Année 1 (CAPE 32) : w = 3,31 %, revenu 49 700 €. Années 2-3 : krach, portefeuille 1,15 M€, CAPE 22 : w = 4,02 % : revenu 46 200 € (−7 %, quand le pourcentage fixe aurait servi −23 %). Années 4-7 : reprise molle, CAPE 24-26, revenu 47-51 000 €. Année 8 : euphorie, portefeuille 1,9 M€, CAPE 36 : w = 3,14 % : revenu 59 700 € : la règle laisse monter, mais deux fois moins vite que le portefeuille : le reste est thésaurisé contre la suite. Bilan de la décennie : revenu entre 46 200 et 59 700 € (±13 % autour de la moyenne) pour un portefeuille qui a oscillé de ±27 % : l'auto-lissage a fait son travail, sans un seul paramètre retouché. Le même chemin est visible dans pofo via ABW + ancre CAPE, §04.
:::

## Pour qui

Le profil règle CAPE recoupe celui du VPW ([[vpw]]) : plancher couvert ou budget élastique, goût de la simplicité exécutable ; avec deux traits propres : une adhésion intellectuelle à la logique des valorisations (celui qui doute du CAPE trichera au premier creux) et l'acceptation d'un revenu officiellement variable contre la consommation la plus intelligemment datée du panorama : la règle CAPE est celle qui dépense le plus quand dépenser coûte le moins cher en rendements futurs sacrifiés. Pour le FIRE en phase à découvert, elle demande le même pont de pension que le VPW ; en marché cher au départ, elle a l'élégance de **démarrer** prudente d'elle-même (c'est sa nature), là où toutes les autres règles exigent qu'on pense à les décoter.

## L'essentiel à retenir

- Règle : taux = a + b × (1/CAPE) sur le portefeuille courant (ERN : a ≈ 1,75-2, b ≈ 0,5) : un pourcentage fixe dont le w devient l'estimation du rendement soutenable du moment.
- Sa propriété distinctive : la double contra-cyclicité : dans un krach de compression de multiples, w monte pendant que le portefeuille baisse : le revenu ne prend qu'une fraction du choc : c'est un lissage économiquement fondé, pas cosmétique.
- Ses difficultés : le niveau du CAPE se discute (utiliser une version ajustée ou recalibrer a), le CAPE US pour un portefeuille mondial est une approximation prudente, et l'exécution au creux exige du cran dans les deux sens.
- Ne **jamais** confondre avec le timing d'allocation : la règle règle le robinet de consommation, pas le portefeuille.
- Mise en œuvre : une cellule de tableur ; dans pofo, la forme aboutie est ABW + ancre CAPE (amortissement au rendement impliqué par les valorisations) ; en douceur, conditionner au CAPE le taux initial et les grandes révisions d'un retrait fixe amendé.

---

## Pour aller plus loin

- Early Retirement Now, Part 18 (les règles flexibles CAPE) et surtout Part 54 (« Dynamic Withdrawal Rates Based on the Shiller CAPE ») : la formalisation, les régressions et les paramètres ([[serie-ern]]).
- Les données CAPE : le site de Robert Shiller ; multpl.com pour la lecture du jour ; Barclays et Research Affiliates pour les CAPE par pays.
- Kitces, « Should Equity Valuation Impact Safe Withdrawal Rates? » : la version praticien du débat.
- Dans ce livre : [[valorisations-et-cape]] (l'indicateur et ses critiques), [[amortissement-abw]] (la forme actuarielle complète), [[choisir-sa-strategie]] (l'arbitrage final).
