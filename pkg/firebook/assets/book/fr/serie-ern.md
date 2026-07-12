# La série Safe Withdrawal Rate d'ERN : guide de lecture

S'il ne fallait recommander qu'une seule source externe sur tout le sujet de ce livre, ce serait celle-là : la « Safe Withdrawal Rate Series » du blog *Early Retirement Now* (ERN), tenue par Karsten Jeske, docteur en économie passé par la Fed d'Atlanta et la gestion quantitative, retraité précoce en 2018. Plus de soixante volets publiés depuis 2016, tous fondés sur des simulations reproductibles (données mensuelles américaines depuis 1871), qui ont, article après article, démonté les slogans du FIRE de première génération et établi l'essentiel de ce qui se dit de sérieux sur les retraites longues. Cette page est un guide de lecture : les résultats majeurs de la série, où les trouver, et ce qu'il faut savoir de ses partis pris pour la lire intelligemment. Beaucoup d'articles de ce livre dialoguent avec elle ; celui-ci vous donne la carte.

::: cle Pourquoi cette série compte
Avant ERN, le débat FIRE opposait des slogans (« 4 %, c'est réglé » contre « tout peut arriver »). La série a imposé une méthode : des simulations exhaustives sur 150 ans de données mensuelles, TOUS les millésimes, des horizons de 50-60 ans, et une honnêteté systématique sur ce qui marche, ce qui ne marche pas, et combien ça coûte. On peut être en désaccord avec ses conclusions (certains le sont, [[anarkulova-cederburg]]) ; on ne peut plus revenir aux slogans.
:::

## L'auteur et la méthode

Le cadre de travail d'ERN, constant sur toute la série : données mensuelles américaines depuis 1871 (actions S&P composite, obligations d'État 10 ans, via Shiller), retraites simulées à TOUS les mois de départ possibles, horizons de 30 à 60 ans, et le critère regardé de près : non seulement la ruine, mais la richesse finale et le chemin (quand échoue-t-on, après quoi). Les outils sont publics : la « toolbox » Google Sheets du Part 28 permet de refaire chaque calcul chez soi.

Deux partis pris à connaître pour lire correctement. D'abord, données AMÉRICAINES : la série hérite du biais optimiste de l'échantillon ([[etude-trinity]], [[anarkulova-cederburg]]) ; ERN le sait, le dit, et considère que 150 ans mensuels d'un grand marché, avec 1929 et 1966 dedans, disciplinent déjà fortement les conclusions. Ensuite, une posture : ERN écrit pour le retraité précoce qui ne veut pas dépendre de la chance, d'où une exigence de robustesse (souvent : « survivre au pire millésime ») plus dure que les 90-95 % de succès des praticiens ([[guardrails-morningstar]], [[ruine-et-probabilites]]).

## Les résultats majeurs, partie par partie

**Le socle : 4 % n'est pas fait pour vous (Parts 1-3, 26).** Le résultat fondateur de la série : la règle des 4 % tient sur 30 ans, mais un horizon de 50-60 ans exige plutôt 3,25-3,5 % en rigide, et le taux sûr dépend fortement des valorisations de départ (Part 3) : à CAPE élevé (au-dessus de 20), les SAFEMAX historiques tombent, tous horizons confondus. Le Part 26 (« Ten Things the Makers of the 4% Rule Don't Want You to Know ») est le meilleur résumé grand public ([[la-regle-des-4-pourcents]], [[valorisations-et-cape]]).

**La séquence explique tout (Parts 14-15).** Démonstration chiffrée que le rendement moyen sur 30 ans compte moins que le rendement des 5-10 premières années : le cœur de [[sequence-des-rendements]]. Corollaire (Part 53) : l'épargnant et le rentier ont des expositions OPPOSÉES à la séquence.

**La flexibilité est surestimée (Parts 9-11, 23-25, 58).** La série la plus contrariante et la plus utile. Les règles de Guyton-Klinger (Parts 9-10) affichent des taux de « succès » flatteurs mais au prix, dans les mauvais millésimes, de DÉCENNIES de dépenses amputées de 30-45 % : la ruine est remplacée par une pauvreté prolongée qu'aucun tableau de taux de succès ne montre ([[guyton-klinger]]). Les Parts 23-25 et 58 généralisent : toute flexibilité réaliste (bornée, tenable) vaut quelques dixièmes de point de taux de retrait, pas la magie annoncée ([[flexibilite-realite]]). C'est ce résultat qui a poussé tout le domaine (et pofo, section §04) à afficher le niveau de vie VÉCU, pas seulement la survie du portefeuille.

**Le CAPE comme règle, pas comme peur (Parts 18, 54).** La proposition constructive de la série : des règles de retrait CAPE-based, où le taux initial (et courant) s'ajuste aux valorisations, formalisées au Part 54. C'est l'ancêtre direct de l'ancre CAPE de la page FIRE ([[regles-cape]]).

**Les buckets et le cash ne font pas ce qu'on croit (Parts 12, 48, 55).** Analyse à contre-courant du dogme des « seaux » : un matelas de cash consommé-rechargé mécaniquement améliore peu la ruine (le cash coûte en rendement ce qu'il économise en séquence) et les stratégies de buckets populaires sont souvent du market timing déguisé sans règle claire ([[cash-buffer]], [[strategie-buckets]], [[recharger-ou-pas]]). Nuance importante : pofo trouve le même ordre de grandeur (l'arbitrage buffer de la §07 est généralement plat), ce qui n'enlève pas au buffer sa valeur PSYCHOLOGIQUE et de gouvernance ([[psychologie-du-retrait]]).

**Les glidepaths marchent (Parts 19-20, 43).** La contrepartie positive : partir avec 60 % d'actions et remonter vers 100 % sur 10-15 ans améliore matériellement les pires millésimes, pour un coût modeste dans les bons : la protection concentrée sur la fenêtre fragile ([[glidepaths]], [[les-trois-phases]]).

**Rentes, obligations, or, immobilier, levier (Parts 29-31, 34-36, 40, 49, 52, 56, 59).** Les inventaires d'actifs : le « yield illusion » des portefeuilles à dividendes (Parts 29-31 : vivre des dividendes n'est PAS plus sûr, [[erreurs-classiques-fire]]), l'or comme couverture partielle de séquence (Part 34, [[or-en-retrait]]), l'immobilier (Part 36, [[immobilier-en-retrait]]), les rentes et la Sécurité sociale américaine (Parts 56, 59 : à adapter au système français, [[rentes-et-annuites]], [[retraite-legale]]), le levier (Parts 49, 52 : sophistiqué, à petites doses, [[levier-et-marges]]).

**Les à-côtés qui changent la vie (Parts 21-22, 27, 37, 42, 47, 60).** Le crédit immobilier en retraite (Part 21 : le garder est un pari de levier sur la séquence), le « one more year » chiffré (Part 42, [[une-annee-de-plus]]), quand s'inquiéter en cours de route (Part 47, [[quand-s-inquieter]]), traverser un ours (Part 37, [[marche-baissier-en-retraite]]), et la critique de « Die With Zero » (Part 60, [[depenses-en-retraite]]).

::: astuce Par où commencer, selon votre question
Le tour d'horizon : Parts 1 et 26. « Combien puis-je retirer ? » : Parts 2-3 puis 54 (CAPE). « La flexibilité me sauvera-t-elle ? » : Parts 23-25 puis 58. « Mon allocation autour du départ » : Parts 19-20. « Les buckets ? » : Parts 12 et 48. Chaque volet se lit en 20-40 minutes ; la série entière est un livre de fait, plus long que celui-ci, et le complète : ERN pousse les simulations américaines plus loin, ce livre couvre le cadre français, l'échantillon mondial et l'outillage pofo.
:::

## Lire ERN avec les bons filtres

Trois filtres pour un usage optimal depuis la France.

**Le filtre géographique.** Tout est américain : données, fiscalité, Social Security, comptes 401(k). Les MÉCANISMES (séquence, CAPE, glidepaths, flexibilité) se transposent parfaitement ; les CHIFFRES héritent du biais d'échantillon (plutôt optimistes, [[anarkulova-cederburg]]) et le chapitre fiscal ne se transpose pas du tout ([[enveloppes-francaises]], [[taxe-puma]]).

**Le filtre de posture.** L'exigence « survivre au pire millésime historique » est une posture de prudence assumée, plus dure que le 90 % de succès des praticiens. Si vous avez des filets solides (pension, employabilité, flexibilité réelle), les recommandations d'ERN sont pour vous une borne prudente, pas un minimum vital ([[ruine-et-probabilites]], [[combien-il-vous-faut]]).

**Le filtre du perfectionnisme.** La série démontre qu'on PEUT raffiner sans fin (le dixième de point de SWR, l'allocation à 5 % près). Rappel utile en la lisant : les trois premiers leviers d'un plan réel restent les dépenses auditées, la pension comptée et une règle écrite ([[erreurs-classiques-fire]]) ; le raffinement vient après.

::: terrain Ce que la communauté en a fait
La série a durablement changé la conversation FIRE : « quel est ton SWR ? » a remplacé « tu as fait tes 25x ? », les simulateurs sérieux affichent désormais le niveau de vie vécu et pas seulement le taux de succès, et « ERN Part N » est devenu une référence courante sur les forums, signe d'une communauté qui a monté son niveau d'exigence. Le conseil d'expérience : ne lisez pas les 60 volets d'affilée ; prenez ceux qui répondent à votre question du moment, et revenez-y. C'est une encyclopédie, pas un roman.
:::

## L'essentiel à retenir

- La série SWR d'ERN (earlyretirementnow.com, 60+ volets depuis 2016) est la référence moderne du retrait de long horizon : méthodique, reproductible, gratuite.
- Ses résultats structurants : 3,25-3,5 % rigide pour 50-60 ans, le rôle dominant de la séquence et des valorisations, la flexibilité surestimée (elle déplace la douleur), les buckets démystifiés, les glidepaths validés.
- Ses partis pris : données américaines (chiffres plutôt optimistes), exigence de pire cas (recommandations plutôt prudentes) : les deux biais se compensent partiellement, sachez-le en la lisant.
- Depuis la France : gardez les mécanismes, adaptez les chiffres, ignorez le chapitre fiscal américain ; ce livre fait la jonction.
- Beaucoup de concepts de la page FIRE de pofo (ancre CAPE, niveau de vie vécu, stress de séquence, arbitrage buffer) sont en dialogue direct avec cette série ([[utiliser-la-page-fire]], [[la-machine-pofo]]).

---

## Pour aller plus loin

- Le point d'entrée : earlyretirementnow.com/safe-withdrawal-rate-series/ (la table des matières complète et tenue à jour de la série).
- La toolbox (Part 28) : le classeur de simulation public, pour refaire les calculs.
- Karsten Jeske en podcast (ChooseFI, Rational Reminder, Bogleheads) : la version orale, souvent plus accessible, des mêmes résultats.
- Les contrepoints dans ce livre : [[anarkulova-cederburg]] (l'échantillon au-delà des États-Unis) et [[guardrails-morningstar]] (la lecture praticienne, moins pire-cas).
