# Guyton-Klinger : les guardrails historiques, grandeur et limites

En 2006, le planificateur financier Jonathan Guyton et l'informaticien William Klinger publient la règle de retrait dynamique la plus influente jamais écrite : un jeu de « decision rules » qui promettait de porter le taux de retrait initial à 5,2-5,6 % (contre le 4 % de Bengen !) en échange d'ajustements occasionnels du train de vie. Le succès fut immense : « guardrails » est devenu le nom commun de toute une famille de stratégies, et la règle reste aujourd'hui l'une des plus utilisées par les conseillers américains. Puis la recherche moderne, ERN en tête, a rouvert le capot et trouvé le vice caché : dans les mauvais millésimes, les coupes se répètent année après année et le revenu s'enfonce de 30 à 45 % pendant des décennies : le succès du portefeuille était acheté par un échec du train de vie que les tableaux de l'article original ne montraient pas. Cet article raconte les deux moitiés de l'histoire : les règles exactes (souvent citées, rarement énoncées correctement), pourquoi elles semblaient si puissantes, la pathologie précise et sa démonstration, les correctifs modernes (le plancher, les bandes resserrées, et la descendance Kitces-Tharp-Morningstar), les paramètres défendables si vous l'utilisez, et sa mise en œuvre dans pofo, plancher compris.

::: cle Le principe en une phrase
Guyton-Klinger est un retrait fixe indexé qui se surveille lui-même : tant que le taux de retrait courant (retrait / portefeuille) reste dans un corridor de ±20 % autour du taux initial, on vit comme sous Bengen ; s'il en sort par le haut (le portefeuille a trop baissé), on coupe le retrait de 10 % ; s'il en sort par le bas (le portefeuille s'est envolé), on l'augmente de 10 %. La grandeur : transformer la flexibilité vague (« je ferai attention si ça va mal ») en mécanique écrite. La limite : rien ne borne le **nombre** de coupes.
:::

## Les règles exactes, pour une fois

L'article de 2006 (« Decision Rules and Maximum Initial Withdrawal Rates ») définit quatre règles, presque toujours tronquées dans les citations. Les voici complètes, car les détails font la stratégie.

**1. La règle de gestion du portefeuille (Portfolio Management Rule).** L'ordre des ventes : on finance le retrait d'abord par les liquidités et les gains des classes en surperformance, jamais en vendant les actions une année où elles ont baissé (leurs ventes attendront la récupération, les obligations et le cash faisant le pont). C'est un mini-buffer procédural intégré ([[cash-buffer]]), souvent oublié dans les répliques, et une partie du bénéfice mesuré de la stratégie vient de là.

**2. La règle de retenue d'inflation (Withdrawal Rule).** L'indexation inflation est **sautée** les années qui suivent un rendement négatif du portefeuille, si le taux de retrait courant dépasse le taux initial. C'est exactement le « gel après année rouge » ([[retrait-fixe-bengen]]), en version conditionnelle ; l'article plafonne par ailleurs l'indexation à 6 % par an.

**3. La règle de préservation du capital (Capital Preservation Rule), le garde-fou bas.** Si le taux de retrait courant dépasse 120 % du taux initial (exemple : taux initial 5 %, courant > 6 %), le retrait est **coupé** de 10 %. Ne s'applique plus dans les quinze dernières années de l'horizon (couper à 82 ans pour protéger un portefeuille qui n'a plus que dix ans à tenir n'a pas de sens).

**4. La règle de prospérité (Prosperity Rule), le garde-fou haut.** Si le taux courant descend sous 80 % du taux initial (le portefeuille s'est envolé), le retrait est **augmenté** de 10 %. C'est la sœur du cliquet de Kitces, mais réversible : la hausse pourra être reprise par une coupe future.

L'ensemble est cohérent et exécutable : chaque 1er janvier, un ratio à calculer, trois comparaisons, au plus un ajustement de ±10 %. La promesse chiffrée de l'article : avec ces règles, un portefeuille 65 % actions soutenait un taux **initial** de 5,2-5,6 % avec 99 % de « succès » sur 40 ans : un point et demi de mieux que Bengen, soit, en multiple, 18-19x au lieu de 25x ([[combien-il-vous-faut]]). On comprend l'enthousiasme : c'était, en apparence, cinq ans de travail économisés.

## Pourquoi ça semblait si fort, et où est le vice

D'où vient ce point et demi « gratuit » ? De trois sources, très inégalement avouables. La première est légitime : les règles 1 et 2 sont de vraies améliorations quasi indolores (l'ordre des ventes et le gel d'indexation valent ensemble ~0,3-0,5 point, la recherche ultérieure l'a confirmé). La deuxième est un artefact d'époque : les simulations de 2006 portaient sur des données américaines favorables et un horizon de 40 ans au plus ([[pieges-des-simulateurs]]). La troisième est le vice de construction, et il faut le regarder en face : **le « succès » de l'article mesure la survie du portefeuille, pas celle du train de vie, et les coupes de la règle 3 sont ILLIMITÉES en nombre.**

Déroulons la pathologie sur le millésime type ([[etude-trinity]], 1966). Le régime hostile s'installe ([[regimes-de-marche]]) : le portefeuille baisse, le taux courant franchit 120 % du taux initial : coupe de 10 %. L'année suivante, le portefeuille a encore baissé (les ours sont collants) : le taux courant re-franchit le seuil : nouvelle coupe. Et ainsi de suite : dans les simulations d'ERN (Parts 9-10 de la série, [[serie-ern]]), les départs de 1966-1969 sous Guyton-Klinger subissent des cascades de coupes qui amènent le revenu réel à **−35 à −45 % du niveau initial, et l'y maintiennent pendant dix à vingt ans**. Le portefeuille, lui, survit magnifiquement : c'est précisément **parce que** le retraité a été mis à la diète pendant deux décennies. Le taux de succès affiche 99 % ; la vie vécue affiche une génération de vaches maigres. Et l'asymétrie psychologique est cruelle : chaque coupe de 10 % arrive **après** une année de baisse, au moment de moral minimal, et la règle de prospérité ne rend les hausses que des années plus tard.

La conclusion de la recherche moderne n'est pas « Guyton-Klinger ne marche pas » : c'est que **son chiffre de ruine n'est pas comparable à celui d'une règle fixe**. Une ruine de 1 % sous GK signifie « même en coupant sans limite, 1 % des scénarios s'épuisent » : c'est un tout autre événement, bien plus grave, qu'une ruine de 5 % sous Bengen. Comparer les deux colonnes sans lire le revenu servi est **le** contresens de toute cette partie ([[panorama-strategies-retrait]]) : pofo l'écrit d'ailleurs noir sur blanc dans l'aide de sa case guardrails.

::: attention Le taux initial était le vrai coupable
La pathologie est aggravée par un choix marketing de l'article original : le taux **initial** de 5,2-5,6 %. Partir aussi haut, c'est entrer dans le corridor déjà proche du garde-fou bas : la première décennie médiocre déclenche la cascade. Les mêmes règles avec un taux initial de 4-4,5 % coupent rarement, et brièvement. La leçon générale vaut pour toutes les règles flexibles : la flexibilité permet de partir un peu plus haut que Bengen, pas un point et demi plus haut ([[flexibilite-realite]]) ; celui qui consomme d'avance tout le bénéfice de sa flexibilité n'a plus de flexibilité, seulement un programme d'austérité différée.
:::

## Les correctifs modernes : borner la descente

La postérité a réparé Guyton-Klinger de trois façons, du patch au remplacement.

**Le plancher (le patch indispensable).** Interdire aux coupes de pousser le revenu sous un pourcentage du niveau initial : 75-80 % est la valeur de planification courante, à caler sur **votre** plancher réel établi ([[combien-il-vous-faut]]). Effet : la descente est bornée, la « génération de vaches maigres » devient au pire « quelques années à −20-25 % ». Contrepartie honnête et inévitable : borner la descente **recrée** de la ruine (si le plancher lui-même est insoutenable dans un scénario, le portefeuille s'épuise) : le chiffre de ruine redevient un vrai chiffre, comparable, et c'est très bien ainsi. C'est exactement l'implémentation de pofo : la case guardrails plus le curseur « Guardrails floor » ([[utiliser-la-page-fire]]).

**Les bandes et les doses.** Deuxième famille de réglages : resserrer ou desserrer le corridor (±20 % standard) et la taille des ajustements (±10 %). Des bandes plus étroites lissent (ajustements plus fréquents mais plus petits) ; des coupes plus petites (5 %) répétées font moins mal que des coupes de 10 % espacées : la recherche penche pour des ajustements plus fréquents et plus doux, qui rapprochent la règle d'un lissage continu ([[pourcentage-fixe]], règle de Yale).

**La descendance : les guardrails par risque.** Le remplacement conceptuel, dû à Kitces et Derek Tharp puis industrialisé par Morningstar : au lieu de surveiller le taux de retrait courant (un thermomètre grossier qui ignore l'horizon restant et les pensions à venir), surveiller la **probabilité de succès** recalculée du plan, et ajuster quand elle sort de son corridor. Même architecture, meilleur capteur : un taux courant de 6 % à 80 ans avec pension n'a rien d'alarmant, le même à 52 ans est grave : le guardrail par risque le sait, celui de 2006 non. C'est l'état de l'art de la famille, et il a son article : [[guardrails-morningstar]].

## Si vous l'utilisez : les paramètres défendables

Pour un plan FIRE long qui choisit la famille guardrails version 2006 (pour sa simplicité de calcul à la main : le guardrail par risque exige un simulateur), la configuration que la littérature post-ERN soutient :

- **Taux initial : 4 à 4,5 %**, pas 5,5 : la flexibilité achète ~0,5 point au-dessus du fixe équivalent, pas davantage ([[flexibilite-realite]]).
- **Corridor ±20 %, ajustements de 10 %** (le standard), ou ±15 %/5 % pour la version douce.
- **Plancher à 75-80 % du retrait initial**, aligné sur votre plancher réel : non négociable.
- **Gel d'indexation après année rouge et ordre des ventes** (règles 1-2) : gardez-les, c'est la partie gratuite.
- **Suspension des coupes en toute fin d'horizon** (l'article original le prévoyait) et dès que la pension couvre le plancher ([[revenus-complementaires]]).
- **La revue est annuelle, à date fixe** : la règle se calcule le 1er janvier, pas à chaque frayeur ([[revue-annuelle]]).

::: exemple La cascade, avec et sans plancher
Plan : 1,3 M€, taux initial 4,3 % (55 900 €), corridor ±20 % (seuils : 3,44 %/5,16 %), coupes de 10 %. Régime hostile : le portefeuille réel glisse à 950 000 € en trois ans. Année 3 : taux courant 5,9 % > 5,16 % : coupe à 50 300 €. Années 4-5 : l'ours colle, deux nouvelles coupes : 40 700 € (−27 %). **Sans** plancher, le scénario 1966 continuerait : cinq coupes, revenu à ~33 000 € (−41 %) pendant une décennie. **Avec** plancher à 78 % (43 600 €), la troisième coupe s'arrête au plancher : le revenu passe la traversée à −22 %, et la ruine du plan remonte de ~1 % à ~4 % : c'est le prix, honnête et visible, d'avoir refusé la diète illimitée. La §04 de pofo montre exactement cette comparaison pour votre plan : cochez guardrails, bougez le plancher, regardez la vie vécue changer de forme.
:::

## L'essentiel à retenir

- Les quatre règles de 2006 : ordre des ventes, gel d'indexation conditionnel, coupe de 10 % quand le taux courant dépasse 120 % de l'initial, hausse de 10 % sous 80 % : une mécanique exécutable qui a inventé les guardrails.
- La promesse (5,2-5,6 % initial, 99 % de succès) reposait sur un vice : des coupes illimitées dont les mauvais millésimes abusent : revenu réel à −35/−45 % pendant des décennies. Le taux de succès de GK ne se compare **jamais** à celui d'une règle fixe sans lire le revenu servi.
- Les correctifs : un plancher à 75-80 % (qui recrée de la ruine honnête : c'est le but), des ajustements plus doux et plus fréquents, et, en remplacement conceptuel, les guardrails par risque de Kitces-Tharp-Morningstar ([[guardrails-morningstar]]).
- Paramètres défendables aujourd'hui : taux initial 4-4,5 %, corridor ±20 %, coupes 10 %, plancher aligné sur le plancher réel, gel d'indexation conservé, revue annuelle à date fixe.
- Dans pofo : case guardrails (corridor ±20 %, ajustements ±10 %) + curseur de plancher ; jugez **toujours** sur la §04 (la vie vécue) et la frontière §06, jamais sur la seule ruine.

---

## Pour aller plus loin

- Guyton & Klinger, « Decision Rules and Maximum Initial Withdrawal Rates », *Journal of Financial Planning*, 2006 : l'article original (lisible, et instructif à relire en connaissant la suite).
- Early Retirement Now, Parts 9-10 : la démonstration de la pathologie, simulations à l'appui ([[serie-ern]]).
- Kitces & Tharp, « The Extraordinary Upside Potential Of Sequence Of Return Risk In Retirement » et les publications guardrails de kitces.com : la descendance par risque.
- Dans ce livre : [[guardrails-morningstar]] (l'état de l'art de la famille), [[plancher-plafond]] (l'autre voie du corridor), [[flexibilite-realite]] (ce que la flexibilité peut vraiment acheter).
