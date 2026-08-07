# Consommer et recharger un buffer : les règles qui marchent

Le débat sur les matelas de liquidités se concentre toujours sur la taille. Deux ans, trois ans ? La taille est pourtant la variable la moins importante. Toute la valeur d'un buffer se joue dans ses flux. Quand y puise-t-on, quand et comment le remplit-on, et quand cesse-t-on de l'entretenir ? Deux ménages peuvent partager le même matelas de 30 mois. L'un en tire une vraie protection de séquence, l'autre un simple compte d'attente coûteux. La différence tient entièrement aux règles.

Ce court traité de plomberie complète [[cash-buffer]], qui traite du pourquoi et du dimensionnement. Il passe en revue les déclencheurs de consommation possibles et leurs pathologies, les trois bonnes sources de recharge et la seule interdiction absolue (recharger en baisse), l'option souvent optimale de ne pas recharger (le buffer fondant de première décennie), et l'orchestration avec la règle de retrait et le rééquilibrage. Il décrit enfin la mécanique exacte du simulateur. Ses réglages par défaut (consommation sous drawdown de 10 %, recharge progressive plafonnée, année d'arrêt) sont précisément les règles que la littérature recommande. Les connaître, c'est pouvoir les régler en connaissance de cause.

::: cle Les trois règles d'or des flux
Un. On consomme le buffer sur un déclencheur écrit, un seuil de drawdown, jamais à l'humeur. Trop tôt, on le gaspille avant le vrai creux. Trop tard, il ne protège rien. Deux. On recharge en terrain calme, progressivement, et jamais pendant une baisse. Recharger au creux, c'est vendre des actions déprimées pour remplir du cash, l'exact péché que le buffer existe pour empêcher. Trois. Un buffer peut légitimement ne pas être rechargé. Il fond sur la première décennie puis disparaît, apportant la protection au moment où elle sert et éteignant ensuite son coût d'opportunité ([[sequence-des-rendements]]).
:::

## Consommer : le choix du déclencheur

Le déclencheur décide de tout. Passons les candidats en revue.

**Le drawdown du portefeuille (recommandé).** « Les retraits basculent sur le buffer tant que le portefeuille est en baisse de plus de X % par rapport à son sommet réel. » C'est le déclencheur propre. Il est observable sans ambiguïté, aligné sur le mécanisme à combattre (vendre déprécié), et auto-terminant, car le retour sous le seuil rebascule les retraits sur le portefeuille. Le réglage de X demande du doigté. Trop bas (5 %), le buffer s'active à chaque respiration du marché et se vide avant les vrais creux. Trop haut (30 %), il regarde passer les baisses moyennes, qui font pourtant l'essentiel des dégâts cumulés. La zone raisonnable tient dans 10-20 %, le défaut étant 10 % (réglable). On vise plutôt 15-20 % pour les petits buffers, afin d'économiser les munitions pour les vraies traversées, et plutôt 10 % pour les gros.

**L'année négative (acceptable, plus fruste).** « Après une année de rendement négatif, l'année suivante se finance sur le buffer. » C'est simple et exécutable de tête, mais binaire. Une année à −2 % déclenche, alors qu'un krach de −18 % en cours d'année ne déclenche qu'en janvier. Le drawdown continu fait mieux pour le même effort.

**Le taux de retrait courant (à réserver au pilotage global).** « Buffer si le taux courant dépasse 4,5 %. » Cet indicateur-là appartient à la règle de retrait et aux guardrails ([[quand-s-inquieter]], [[guardrails-morningstar]]). L'utiliser aussi pour le buffer mélange deux étages de décision. Gardez les instruments séparés.

**La discrétion (à proscrire).** « Je verrai bien. » C'est la variante qui transforme le matelas en totem ([[strategie-buckets]]). Sous stress, l'humain consomme trop tôt, car la correction de 8 % « fait peur ». Ou bien il refuse de consommer, par ce réflexe paradoxal et très documenté qui pousse à protéger le buffer (« c'est ma sécurité, n'y touchons pas ! ») et à vendre des actions au creux à la place. L'instrument se retrouve exactement retourné contre son usage.

Deux raffinements sont utiles au passage. Le premier est l'hystérésis, qui consiste à rebasculer sur le portefeuille seulement quand le drawdown repasse nettement sous le seuil. Déclencher à 15 % et désarmer à 10 %, par exemple, évite les allers-retours au voisinage du seuil. Le second est la consommation partielle, qui bascule d'abord la moitié des retraits, puis tout au-delà d'un second seuil. Elle étire les munitions dans les traversées longues ([[regimes-de-marche]]).

## Recharger : trois bonnes sources, une interdiction

**Source 1 : le terrain calme, progressivement.** C'est la recharge standard. Hors drawdown, une fraction des ventes mensuelles est dirigée vers le buffer jusqu'au retour à la cible. C'est exactement la mécanique du simulateur, active seulement quand le déclencheur est désarmé, plafonnée par mois pour rester progressive, et jamais une grosse vente de reconstitution d'un coup. La progressivité compte. Reconstituer 30 mois de buffer en un trimestre concentre le risque de timing que l'étalement dilue.

**Source 2 : les nouveaux sommets (la version stricte).** « Recharge uniquement quand le portefeuille marque un nouveau plus-haut réel. » Cette règle est plus conservatrice que « hors drawdown ». Elle garantit qu'on ne vend, pour recharger, que des actifs à leur meilleur historique. Elle a un coût. Après une traversée longue, le buffer peut rester dégarni des années en attendant le sommet. La version « hors drawdown » est le compromis praticable, la version « sommets » l'idéal-type pour les gros buffers.

**Source 3 : les excédents, sans vendre.** C'est la plus élégante, car elle n'exige aucune vente. Les retraits budgétés non dépensés, les revenus d'appoint ([[retour-au-travail]]), les dividendes non capitalisés d'un vieux compte, les rentrées exceptionnelles, tout cela va d'office vers le buffer tant qu'il est sous sa cible. Le coût de timing est nul, et la gouvernance y gagne. Le buffer devient la destination par défaut des bonnes surprises.

**L'interdiction : jamais en baisse.** Elle mérite son paragraphe, car la tentation est réelle et son habillage rationnel. « Le buffer est bas, les marchés peuvent rebaisser, sécurisons pendant qu'il est temps. » Faites la comptabilité. Vendre à −20 % pour remplir du cash, c'est cristalliser la perte sur la fraction vendue et rater le rebond dessus, la double peine exacte que le dispositif combat. Un buffer vide en fin de traversée n'est pas un échec. C'est un buffer qui a servi, et il se reconstituera au calme. L'écrire dans le plan en toutes lettres (« aucune recharge tant que le drawdown excède le seuil ») coûte une ligne et sauve la stratégie.

::: science Ne pas recharger du tout : le buffer fondant
L'option la plus sous-employée est aussi la plus logique. Puisque le risque de séquence se concentre sur la première décennie ([[sequence-des-rendements]]), un buffer qui fond aligne le coût de la protection sur la période où elle sert. Un tel buffer est consommé si besoin, jamais rechargé au-delà d'une date, puis absorbé par le portefeuille. C'est l'analogue en cash du glidepath ([[glidepaths]]). Les deux se combinent d'ailleurs naturellement, car le buffer fondant est la première marche de la remontée en actions. Un simulateur l'implémente d'un curseur, « Buffer refill stops in year ». La recharge reste normale jusqu'à l'année N (typiquement 8-12), puis s'arrête. Le manteau devient cape, puis disparaît. Dans les simulations, le buffer fondant domine généralement le buffer perpétuel de même taille. Il offre la même protection quand ça compte et éteint le coût d'opportunité quand ça ne compte plus. C'est le réglage que ce livre recommande par défaut pour la phase à découvert.
:::

## L'orchestration : buffer, règle de retrait, rééquilibrage

Les trois mécanismes coexistent dans un plan. Leur ordre de préséance mérite d'être écrit, pour éviter les doubles emplois.

**L'ordre des sources en baisse.** Quand le déclencheur est armé, les retraits viennent 1) du buffer, 2) puis, s'il est vide, de la poche obligataire en surpoids, où le prélèvement-rééquilibrage continue de fonctionner ([[obligations-en-retrait]]), 3) et des actions en dernier. Le buffer est la première ligne, pas la seule. Derrière lui, le rééquilibrage fait le même travail en seconde ligne. C'est pourquoi un buffer modeste suffit ([[cash-buffer]]).

**L'articulation avec la flexibilité.** Si votre règle coupe les dépenses en drawdown ([[guyton-klinger]], [[plancher-plafond]]), buffer et coupe se déclenchent souvent ensemble. C'est cohérent, car le buffer finance le plancher réduit et dure d'autant plus longtemps. Mais additionnez les effets dès la conception. Un plan à guardrails plus buffer fondant a besoin d'un buffer plus petit que le plan rigide (18-24 mois contre 30-36). La flexibilité est déjà une partie du matelas ([[flexibilite-realite]]).

**Les réglages qu'un simulateur expose.** La taille (« Buffer years »), le rendement réel du matelas (calez votre fonds euros), le seuil de consommation (10 % par défaut, que l'API permet de durcir), la recharge progressive plafonnée hors drawdown, et l'année d'arrêt (« refill stops in year »). La §07 arbitre la taille. L'A/B avec et sans arrêt de recharge teste le buffer fondant. Dix minutes suffisent pour posséder ses règles au lieu de les subir ([[utiliser-la-page-fire]], [[la-machine-pofo]]).

::: exemple Les règles complètes, sur une carte postale
Le plan de Salomé, 49 ans, phase à découvert de 15 ans. « Buffer : 24 mois de plancher, en fonds euros. Consommer : retraits basculés sur le buffer quand le drawdown réel dépasse 15 %, retour au portefeuille sous 10 %. Recharger : hors drawdown seulement, par un cinquième des ventes mensuelles, plus tous les excédents ; cible 24 mois ; jamais de recharge en baisse, même si le buffer est à zéro. Arrêt : plus aucune recharge après l'année 10, le buffer fond et le portefeuille l'absorbe. » Six phrases, et tout y est, le déclencheur, l'hystérésis, la source, l'interdiction, l'extinction. La simulation §07 le confirme, avec une ruine équivalente au buffer perpétuel de 36 mois pour un tiers de coût d'opportunité en moins. La carte postale va dans le plan écrit, à côté de la règle de retrait ([[construire-son-plan]]).
:::

## L'essentiel à retenir

- La valeur d'un buffer est dans ses flux, pas dans sa taille. Déclencheur écrit, recharge disciplinée, extinction programmée. Sans règles, c'est un compte d'attente coûteux, ou un totem qu'on protège en vendant des actions au creux.
- **Consommer** : sur drawdown réel de 10-20 % (avec hystérésis, éventuellement en deux paliers). Jamais à l'humeur, et jamais sur un indicateur qui appartient à un autre étage, car le taux courant pilote la règle de retrait, pas le buffer.
- **Recharger** : hors drawdown, progressivement (une fraction des ventes mensuelles), aux sommets pour les stricts, par les excédents toujours. Et jamais en baisse. Un buffer vide après une traversée a servi, et il se reconstituera au calme.
- Le buffer fondant (recharge stoppée après l'année 8-12) aligne le coût sur la fenêtre fragile et domine généralement le perpétuel. C'est un curseur natif, et le réglage par défaut recommandé en phase à découvert.
- Orchestration : buffer en première ligne, prélèvement-rééquilibrage en seconde, actions en dernier. Avec une règle flexible, réduisez la taille (18-24 mois), car la flexibilité est déjà du matelas.

---

## Pour aller plus loin

- Early Retirement Now, volet 12 (le contre-dossier du cash, qui vise précisément les buffers sans règles) ([[serie-ern]]).
- Kitces sur les stratégies de « bucket maintenance » : la plomberie comparée des recharges.
- Dans un simulateur : les quatre réglages du buffer et la §07 ([[utiliser-la-page-fire]]) ; la mécanique interne est décrite dans [[la-machine-pofo]].
- Dans ce livre : [[cash-buffer]] (le pourquoi et la taille), [[glidepaths]] (le jumeau en allocation du buffer fondant), [[strategie-buckets]] (ce que deviennent les buckets sans règles de flux).
