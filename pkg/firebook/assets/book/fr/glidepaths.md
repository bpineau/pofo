# Les glidepaths : bond tent, rising equity et la fenêtre fragile

L'article précédent a établi **où** se placer sur l'axe actions/obligations ([[allocation-actions-obligations]]) ; celui-ci ajoute la dimension que l'allocation statique ignore : le **temps**. Le risque d'un plan de retrait n'est pas uniformément réparti : il se concentre massivement sur les cinq à dix premières années, la fenêtre fragile du risque de séquence ([[sequence-des-rendements]]). Or une allocation constante paie la protection (le renoncement au rendement des obligations) au même tarif **toute** la vie, alors que le danger, lui, passe.

D'où l'idée du glidepath (« sentier de descente », le terme vient des fonds à date cible) : faire **varier** l'allocation dans le temps, prudente quand le danger est maximal, croissante en actions à mesure que la fenêtre se referme. La forme complète s'appelle le bond tent (la « tente obligataire » de Michael Kitces) : monter la poche obligataire à L'**approche** du départ, la redescendre pendant la première décennie de retraite : un sommet de prudence exactement sur la zone rouge. C'est l'une des rares idées du domaine validée par tous les camps (Kitces-Pfau 2014, ERN volets 19-20 et 43, la pratique institutionnelle), à condition d'en connaître les vraies proportions : le bénéfice est **ciblé** (les pires cas), pas gratuit, et l'exécution demande d'acheter des actions pendant des années où plus personne n'en veut.

Cet article donne les résultats chiffrés, la mécanique d'exécution des deux versants, les critiques honnêtes, la comparaison avec l'alternative (le buffer), et le test dans pofo, qui implémente le versant montant nativement.

::: cle L'idée en une phrase
Puisque le sort du plan se joue dans la première décennie, la prudence doit être une **dépense temporaire** concentrée sur cette décennie, pas une taxe perpétuelle : partir prudent (50-60 % d'actions au jour J) puis **remonter** méthodiquement vers l'allocation de croisière (70-100 %) sur 10-15 ans capte l'essentiel de la protection d'une allocation prudente permanente, pour une fraction de son coût en rendement.
:::

## Les résultats fondateurs, chiffrés

**Kitces-Pfau (2014).** L'article « Reducing Retirement Risk with a Rising Equity Glide Path » teste, sur simulations et données historiques, des retraites de 30 ans où l'allocation **monte** linéairement (par exemple de 30 % à 70 % d'actions) contre les allocations statiques équivalentes. Résultats : dans les scénarios **médians**, le glidepath fait à peine mieux ou pareil que la statique de même exposition moyenne ; dans les **pires** scénarios (les millésimes à krach précoce), il améliore nettement le taux soutenable et la richesse finale : jusqu'à +0,2-0,4 point de SWR dans les configurations défavorables. Le mécanisme est limpide une fois vu : si le krach frappe **tôt** (le cas dangereux), le retraité montant l'a traversé avec sa prudence maximale, **puis** a racheté des actions aux prix de soldes pendant sa remontée : il a mécaniquement acheté bas ; si le krach frappe tard, il est riche depuis longtemps et rien ne compte plus ([[horizon-et-esperance-de-vie]]). Le glidepath montant est un achat programmé d'actions concentré sur la décennie où acheter bas rapporte le plus.

**ERN (volets 19-20), la version FIRE.** Jeske refait l'exercice sur 60 ans d'horizon et des données mensuelles : mêmes conclusions, calibrées pour la retraite précoce : le chemin type gagnant part de ~60 % d'actions et remonte vers 100 % en une dizaine d'années ; le bénéfice sur le SAFEMAX des pires millésimes est de l'ordre de +0,1 à +0,3 point, ce qui, rappelons l'échelle ([[choisir-sa-strategie]]), est substantiel pour un levier gratuit en espérance. Et une nuance importante : le bénéfice est **maximal** quand les valorisations de départ sont **élevées** ([[valorisations-et-cape]]) : précisément la situation où le krach précoce est le plus probable : le glidepath est une assurance dont la prime baisse quand le risque monte, une rareté.

**Le versant amont : ERN volet 43 et la tente complète.** La même logique vaut **avant** le départ : les cinq dernières années d'accumulation appartiennent déjà à la fenêtre fragile (un krach à cible atteinte moins un an détruit la date de départ, [[les-trois-phases]]). D'où la tente complète de Kitces : réduire les actions de ~80 % vers 50-60 % sur les 5-10 ans qui précèdent le départ (le versant descendant, financé par les cotisations nouvelles dirigées vers les obligations, sans vendre), sommet de prudence au jour J, puis le versant remontant. Le profil dessine une tente dont le sommet coïncide avec le maximum de risque : c'est l'anti-fonds-à-date-cible, qui lui descend pour toujours et laisse le retraité de 80 ans à 25 % d'actions, sous le plateau, exposé à l'érosion ([[allocation-actions-obligations]]).

## L'exécution, versant par versant

**Le versant montant (la retraite), trois décisions.** La **pente** : linéaire sur 10-15 ans est le standard (de 60 à 90 % d'actions, +2 à +3 points par an) ; plus court concentre le bénéfice mais brusque l'exécution. Le **véhicule** : la remontée s'exécute sans frais par le flux naturel des retraits : on prélève **tout** sur la poche obligataire pendant la remontée (le portefeuille dérive mécaniquement vers les actions) : aucune vente d'actions pendant dix ans, ce qui est exactement le service anti-séquence attendu ; le rééquilibrage ne reprend qu'à l'allocation de croisière atteinte. Et la **règle de fin** : écrire l'allocation cible et s'y arrêter : le glidepath n'est pas une dérive perpétuelle.

**Le versant descendant (la transition), une discipline différente.** Il s'exécute par les **flux** (épargne nouvelle vers les obligations et le monétaire, constitution du buffer au passage, [[cash-buffer]]) plutôt que par des ventes massives (fiscalité, et risque de refaire le market timing qu'on prétend éviter). Sur cinq ans à fort taux d'épargne, rediriger les cotisations suffit généralement à passer de 85/15 à 60/40 sans vendre une action : c'est aussi fiscalement optimal ([[enveloppes-francaises]]).

::: attention La vraie difficulté est comportementale
Relisez ce que le versant montant demande : **augmenter** sa part d'actions, chaque année, pendant la première décennie de sa retraite : y compris, et surtout, si cette décennie contient un krach : c'est-à-dire acheter des actions à 62 ans, sans salaire, en plein 2008, parce que le plan l'a écrit. C'est l'exécution la plus contre-intuitive de tout ce livre, et l'échec type du glidepath n'est pas mathématique mais humain : la remontée « suspendue » au premier gros creux, transformant la tente en prudence perpétuelle, qui coûte alors son plein tarif en érosion. Les parades : l'automatisation par les retraits (décrite ci-dessus, la remontée se fait **toute seule** si l'on prélève sur les obligations, aucun ordre d'achat à passer), et l'inscription de la pente dans le plan écrit avec le même statut que la règle de retrait ([[construire-son-plan]], [[psychologie-du-retrait]]).
:::

## Glidepath ou buffer ? Les deux outils de la fenêtre fragile, comparés

Le glidepath et le matelas de liquidités ([[cash-buffer]]) visent le même risque (vendre des actions au creux pendant la fenêtre fragile) par deux mécaniques différentes ; les confondre ou les empiler sans le voir est courant. La comparaison :

| Critère | Glidepath montant | Buffer de liquidités |
|---|---|---|
| Mécanique | L'**allocation** varie dans le temps | Une **poche** dédiée absorbe les retraits au creux |
| Coût en espérance | Faible (prudence temporaire) | Modéré (2-3 ans de dépenses à ~0 % réel en permanence, si maintenu) |
| Bénéfice | Concentré sur les pires millésimes précoces | Concentré sur les creux de 1-3 ans ; insuffisant seul pour un régime de 7 ans ([[regimes-de-marche]]) |
| Gouvernance | Automatique via les retraits ; demande de **tenir** la pente | Très intuitive (« je vis sur le cash ») ; demande des règles de consommation/recharge ([[recharger-ou-pas]]) |
| Valeur psychologique | Faible (invisible) | Très élevée (dormir) |

La lecture honnête, confirmée par les arbitrages de la §07 de pofo : quantitativement, les deux se valent à peu près et leurs bénéfices se **recouvrent** largement (un gros buffer **plus** une tente profonde, c'est payer deux fois la même assurance) ; qualitativement, le glidepath gagne sur l'exécution automatique, le buffer sur la psychologie. La combinaison raisonnable : une tente modérée (60 → 85 % sur 10 ans) plus un buffer modeste (18-24 mois), plutôt que le maximum des deux ([[choisir-sa-strategie]], la prudence est un budget).

## Les nuances de l'état de l'art

**Le bénéfice moyen est modeste : c'est une assurance, pas un alpha.** Répétons l'échelle : +0,1-0,3 point de SWR dans les mauvais cas, ~0 en médiane. Le glidepath ne remplace ni un taux initial correct ([[valorisations-et-cape]]) ni une règle flexible ([[choisir-sa-strategie]]) : il complète. Quiconque le vend comme transformateur a mal lu les papiers.

**Conditionner à l'entrée est légitime.** Puisque le bénéfice croît avec les valorisations de départ, la version affinée module la profondeur de la tente sur le CAPE au jour J : à CAPE < 20, une tente légère ou pas de tente (le risque de krach précoce est faible, le coût d'opportunité réel) ; à CAPE > 30, la tente complète. C'est cohérent avec tout le cadre de ce livre, et ça évite la tente « rituelle » dans un marché déjà purgé.

**L'interaction avec la pension.** La remontée d'actions du versant montant est encore plus défendable quand une pension arrive en cours de route ([[revenus-complementaires]]) : la pension actualisée est une position obligataire qui **grossit** en s'approchant ([[amortissement-abw]]) : remonter les actions du portefeuille visible ne fait alors que maintenir le risque **total** constant. Le retraité français à pension dans 15 ans a une double raison de suivre la pente.

**Dans pofo** : la case « Rising-equity glidepath (bond tent) » du groupe Market model applique le versant montant au portefeuille simulé ; testez-la en A/B sur les colonnes stress et broad-sample (c'est là que son bénéfice vit, [[historique-vs-parametrique]]), et regardez la §03 : la sensibilité de la ruine à la première décennie doit visiblement s'adoucir. En mode portefeuille, l'exécution réelle (prélever sur les obligations) se répète ensuite dans la vraie vie, à la revue annuelle ([[revue-annuelle]]).

::: exemple Une tente dimensionnée, de bout en bout
Iris, 44 ans, cible à 49 ans, CAPE au-dessus de 30 : la tente complète se justifie. Versant descendant (44-49 ans) : allocation actuelle 85/15 ; toute l'épargne nouvelle (2 800 €/mois) va aux obligations intermédiaires et aux linkers, plus 24 mois de dépenses en monétaire la dernière année : au départ, 58/34/8 (actions/obligations/cash), zéro vente, zéro frottement fiscal. Versant montant (49-61 ans) : plan écrit : « tous les retraits sur la poche obligataire jusqu'à 85 % d'actions, puis rééquilibrage à bandes ; pente suspendue par **aucun** événement de marché ». Simulation pofo : au taux de 3,6 % avec guardrails, la tente réduit la ruine stress de 8 % à 6 % et la broad-sample de 12 % à 10 % ; la §03 montre le pire décile de première décennie nettement moins létal. Coût : ~0,15 point d'espérance pendant douze ans. Iris signe : c'est exactement le prix d'une assurance dont elle est le sinistre type.
:::

## L'essentiel à retenir

- Le risque étant concentré sur la première décennie, la prudence doit l'être aussi : partir à 50-60 % d'actions et remonter vers la croisière sur 10-15 ans (le versant montant), après avoir réduit par les flux sur les 5 ans d'avant (le versant descendant) : la tente de Kitces, sommet au jour J.
- Le bénéfice est ciblé : +0,1-0,3 point de SWR dans les pires millésimes, ~0 en médiane, **maximal** quand le CAPE de départ est haut : c'est une assurance à prime négative dans les marchés chers, pas un alpha.
- L'exécution du versant montant est automatique si l'on prélève tout sur les obligations pendant la remontée : aucun ordre contre-intuitif à passer : et c'est la parade à sa vraie faiblesse, comportementale (la pente suspendue au premier krach).
- Glidepath et buffer couvrent le même risque : combinaison modérée des deux plutôt que maximum de l'un et l'autre ; la §07 et le mode A/B de pofo chiffrent l'arbitrage sur votre plan.
- Anti-modèle à éviter : le fonds à date cible qui descend pour toujours et laisse un octogénaire sous le plateau d'allocation, exposé à l'érosion ([[allocation-actions-obligations]]).

---

## Pour aller plus loin

- Kitces & Pfau, « Reducing Retirement Risk with a Rising Equity Glide Path », *Journal of Financial Planning* (2014), et Kitces, « The Bond Tent » ([kitces.com](https://www.kitces.com)) : les fondateurs.
- Early Retirement Now, volets 19-20 (glidepaths en retraite, version 60 ans d'horizon) et volet 43 (pré-retraite) ([[serie-ern]]).
- Pfau, *Retirement Planning Guidebook*, chapitre allocation dynamique.
- Dans ce livre : [[sequence-des-rendements]] (le risque visé), [[cash-buffer]] (l'alternative), [[les-trois-phases]] (le calendrier), [[allocation-actions-obligations]] (le point d'arrivée de la pente).
