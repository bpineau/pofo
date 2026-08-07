# Consommer et recharger un buffer : les règles qui marchent

Le débat sur les matelas de liquidités se concentre toujours sur la **taille** (deux ans ? trois ?), alors que la taille est la variable la moins importante : toute la valeur d'un buffer se joue dans ses **flux** : quand y puise-t-on, quand et comment le remplit-on, et quand cesse-t-on de l'entretenir. Deux ménages avec le même matelas de 30 mois peuvent en tirer l'un une vraie protection de séquence, l'autre un simple compte d'attente coûteux : la différence est entièrement dans les règles.

Ce court traité de plomberie complète [[cash-buffer]] (le pourquoi et le dimensionnement). Il passe en revue les déclencheurs de consommation candidats et leurs pathologies, les trois bonnes sources de recharge et la seule interdiction absolue (recharger en baisse), l'option souvent optimale de **ne pas** recharger (le buffer fondant de première décennie), l'orchestration avec la règle de retrait et le rééquilibrage. Et la mécanique exacte, dont les défauts (consommation sous drawdown de 10 %, recharge progressive plafonnée, année d'arrêt) sont précisément les règles que la littérature recommande : les connaître, c'est pouvoir les régler en connaissance de cause.

::: cle Les trois règles d'or des flux
Un. On consomme le buffer sur un **déclencheur écrit** (un seuil de drawdown), jamais à l'humeur : trop tôt le gaspille avant le vrai creux, trop tard ne protège rien. **Deux** : on recharge en **terrain calme**, progressivement, et **jamais** pendant une baisse : recharger au creux, c'est vendre des actions déprimées pour remplir du cash : l'exact péché que le buffer existe pour empêcher. **Trois** : un buffer peut légitimement ne **pas** être rechargé : fondre sur la première décennie et disparaître : la protection au moment où elle sert, le coût d'opportunité qui s'éteint ensuite ([[sequence-des-rendements]]).
:::

## Consommer : le choix du déclencheur

Le déclencheur décide de tout ; passons les candidats en revue.

**Le drawdown du portefeuille (recommandé).** « Les retraits basculent sur le buffer tant que le portefeuille est en baisse de plus de X % par rapport à son sommet réel. » C'est le déclencheur propre : observable sans ambiguïté, aligné sur le mécanisme à combattre (vendre déprécié), auto-terminant (le retour sous le seuil rebascule les retraits sur le portefeuille). Le réglage de X : trop bas (5 %), le buffer s'active à chaque respiration du marché et se vide avant les vrais creux ; trop haut (30 %), il regarde passer les baisses moyennes qui font pourtant l'essentiel des dégâts cumulés. La zone raisonnable : **10-20 %**, le défaut étant 10 % (réglable) : plutôt 15-20 % pour les petits buffers (économiser les munitions pour les vraies traversées), plutôt 10 % pour les gros.

**L'année négative (acceptable, plus fruste).** « Après une année de rendement négatif, l'année suivante se finance sur le buffer. » Simple, exécutable de tête, mais binaire (une année à −2 % déclenche, un krach de −18 % en cours d'année ne déclenche qu'en janvier) : le drawdown continu fait mieux pour le même effort.

**Le taux de retrait courant (à réserver au pilotage global).** « Buffer si le taux courant dépasse 4,5 %. » Cet indicateur-là appartient à la règle de retrait et aux guardrails ([[quand-s-inquieter]], [[guardrails-morningstar]]) : l'utiliser **aussi** pour le buffer mélange deux étages de décision : gardez les instruments séparés.

**La discrétion (à proscrire).** « Je verrai bien ». C'est la variante qui transforme le matelas en totem ([[strategie-buckets]]) : sous stress, l'humain consomme trop tôt (la correction de 8 % « fait peur ») ou refuse de consommer du tout : le réflexe paradoxal, très documenté, de **protéger** le buffer (« c'est ma sécurité, n'y touchons pas ! ») et de vendre des actions au creux à la place : l'instrument exact retourné contre son usage.

Deux raffinements utiles au passage : l'**hystérésis** (rebasculer sur le portefeuille seulement quand le drawdown repasse **nettement** sous le seuil, par exemple déclencher à 15 % et désarmer à 10 %, évite les allers-retours au voisinage du seuil) ; et la **consommation partielle** (basculer la moitié des retraits d'abord, tout au-delà d'un second seuil, étire les munitions dans les traversées longues, [[regimes-de-marche]]).

## Recharger : trois bonnes sources, une interdiction

**Source 1 : le terrain calme, progressivement.** La recharge standard : hors drawdown, une fraction des ventes mensuelles est dirigée vers le buffer jusqu'au retour à la cible. C'est exactement la mécanique (recharge active seulement quand le déclencheur est désarmé, plafonnée par mois pour rester progressive, jamais une grosse vente de reconstitution d'un coup). La progressivité compte : reconstituer 30 mois de buffer en un trimestre concentre le risque de timing que l'étalement dilue.

**Source 2 : les nouveaux sommets (la version stricte).** « Recharge uniquement quand le portefeuille marque un nouveau plus-haut **réel** » : plus conservatrice que « hors drawdown ». Elle garantit qu'on ne vend pour recharger que des actifs à leur meilleur historique. Coût : après une traversée longue, le buffer peut rester dégarni des années en attendant le sommet : la version « hors drawdown » est le compromis praticable, la version « sommets » l'idéal-type pour les gros buffers.

**Source 3 : les excédents, sans vendre.** La plus élégante : tout ce qui n'exige aucune vente : les retraits budgétés non dépensés, les revenus d'appoint ([[retour-au-travail]]), les dividendes non capitalisés d'un vieux compte, les rentrées exceptionnelles : dirigés d'office vers le buffer tant qu'il est sous sa cible. Zéro coût de timing, et une vertu de gouvernance : le buffer devient la destination par défaut des bonnes surprises.

**L'interdiction : jamais en baisse.** Elle mérite son paragraphe car la tentation est réelle et son habillage rationnel : « le buffer est bas, les marchés peuvent rebaisser, sécurisons pendant qu'il est temps ». Faites la comptabilité : vendre à −20 % pour remplir du cash, c'est cristalliser la perte sur la fraction vendue et rater le rebond dessus : la double peine exacte que le dispositif combat. Un buffer vide en fin de traversée n'est pas un échec. C'est un buffer qui a **servi** : il se reconstituera au calme. L'écrire dans le plan en toutes lettres (« aucune recharge tant que le drawdown excède le seuil ») coûte une ligne et sauve la stratégie.

::: science Ne pas recharger du tout : le buffer fondant
L'option la plus sous-employée est structurellement la plus logique : puisque le risque de séquence se concentre sur la première décennie ([[sequence-des-rendements]]), un buffer qui **fond** (consommé si besoin, jamais rechargé au-delà d'une date, absorbé par le portefeuille ensuite) aligne le coût de la protection sur la période où elle sert. C'est l'analogue en cash du glidepath ([[glidepaths]], et les deux se combinent naturellement, le buffer fondant est la première marche de la remontée en actions). pofo l'implémente d'un curseur : « Buffer refill stops in year » : recharge normale jusqu'à l'année N (typiquement 8-12), plus aucune ensuite : le manteau devient cape, puis disparaît. Dans les simulations, le buffer fondant domine généralement le buffer perpétuel de même taille : même protection quand ça compte, coût d'opportunité éteint quand ça ne compte plus. C'est le réglage que ce livre recommande par défaut pour la phase à découvert.
:::

## L'orchestration : buffer, règle de retrait, rééquilibrage

Les trois mécanismes coexistent dans un plan ; leur ordre de préséance mérite d'être écrit pour éviter les doubles emplois.

**L'ordre des sources en baisse** : quand le déclencheur est armé, les retraits viennent 1) du buffer, 2) puis, s'il est vide, de la poche obligataire en surpoids (le prélèvement-rééquilibrage continue de fonctionner, [[obligations-en-retrait]]), 3) les actions en dernier. Le buffer est la **première** ligne, pas la seule : derrière lui, le rééquilibrage fait le même travail en seconde ligne. C'est pourquoi un buffer modeste suffit ([[cash-buffer]]).

**L'articulation avec la flexibilité** : si votre règle coupe les dépenses en drawdown ([[guyton-klinger]], [[plancher-plafond]]), buffer et coupe se déclenchent souvent ensemble. C'est cohérent (le buffer finance le plancher réduit, il dure d'autant plus longtemps). Mais additionnez les effets à la conception : un plan à guardrails + buffer fondant a besoin d'un buffer **plus petit** que le plan rigide (18-24 mois contre 30-36) : la flexibilité est déjà une partie du matelas ([[flexibilite-realite]]).

**Ce que pofo simule, en résumé de réglages** : la taille (« Buffer years »), le rendement réel du matelas (calez votre fonds euros), le seuil de consommation (10 % par défaut, l'API permet de le durcir), la recharge progressive plafonnée hors drawdown, et l'année d'arrêt (« refill stops in year ») : la §07 arbitre la taille, l'A/B avec et sans arrêt de recharge teste le buffer fondant : dix minutes pour posséder ses règles au lieu de les subir ([[utiliser-la-page-fire]], [[la-machine-pofo]]).

::: exemple Les règles complètes, sur une carte postale
Le plan de Salomé, 49 ans, phase à découvert de 15 ans : « Buffer : 24 mois de plancher, en fonds euros. **Consommer** : retraits basculés sur le buffer quand le drawdown réel dépasse 15 % ; retour au portefeuille sous 10 %. **Recharger** : hors drawdown seulement, par un cinquième des ventes mensuelles, plus tous les excédents ; cible 24 mois ; **jamais** de recharge en baisse, même si le buffer est à zéro. **Arrêt** : plus aucune recharge après l'année 10 : le buffer fond, le portefeuille l'absorbe. » Six phrases, tout y est : le déclencheur, l'hystérésis, la source, l'interdiction, l'extinction. Et la simulation §07 confirme : ruine équivalente au buffer perpétuel de 36 mois, pour un tiers de coût d'opportunité en moins. La carte postale va dans le plan écrit, à côté de la règle de retrait ([[construire-son-plan]]).
:::

## L'essentiel à retenir

- La valeur d'un buffer est dans ses flux, pas dans sa taille : déclencheur écrit, recharge disciplinée, extinction programmée : sans règles, c'est un compte d'attente coûteux ou un totem qu'on protège en vendant des actions au creux.
- **Consommer** : sur drawdown réel de 10-20 % (avec hystérésis, éventuellement en deux paliers) : jamais à l'humeur, jamais sur un indicateur qui appartient à un autre étage (le taux courant pilote la règle de retrait, pas le buffer).
- **Recharger** : hors drawdown, progressivement (une fraction des ventes mensuelles), aux sommets pour les stricts, par les excédents toujours. Et **jamais** en baisse : un buffer vide après une traversée a servi, il se reconstituera au calme.
- Le buffer **fondant** (recharge stoppée après l'année 8-12) aligne le coût sur la fenêtre fragile et domine généralement le perpétuel. C'est un curseur natif, et le défaut recommandé en phase à découvert.
- Orchestration : buffer en première ligne, prélèvement-rééquilibrage en seconde, actions en dernier ; avec une règle flexible, réduisez la taille (18-24 mois) : la flexibilité est déjà du matelas.

---

## Pour aller plus loin

- Early Retirement Now, volet 12 (le contre-dossier du cash, qui vise précisément les buffers **sans** règles) ([[serie-ern]]).
- Kitces sur les stratégies de « bucket maintenance » : la plomberie comparée des recharges.
- Dans pofo : les quatre réglages du buffer et la §07 ([[utiliser-la-page-fire]]) ; la mécanique interne : [[la-machine-pofo]].
- Dans ce livre : [[cash-buffer]] (le pourquoi et la taille), [[glidepaths]] (le jumeau en allocation du buffer fondant), [[strategie-buckets]] (ce que deviennent les buckets sans règles de flux).
