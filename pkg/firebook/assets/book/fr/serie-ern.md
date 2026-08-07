# La série Safe Withdrawal Rate d'ERN : guide de lecture

S'il ne fallait recommander qu'une seule source externe sur tout le sujet de ce livre, ce serait celle-là : la « Safe Withdrawal Rate Series » du blog *Early Retirement Now* (ERN). Elle est tenue par Karsten Jeske, docteur en économie passé par la Fed d'Atlanta et la gestion quantitative, retraité précoce en 2018. Plus de soixante volets depuis 2016, tous fondés sur des simulations reproductibles à partir de données mensuelles américaines depuis 1871. Article après article, ils ont démonté les slogans du FIRE de première génération et posé l'essentiel de ce qui se dit de sérieux sur les retraites longues.

Cette page est un guide de lecture : les résultats majeurs de la série, où les trouver, et ce qu'il faut savoir de ses partis pris pour la lire intelligemment. Beaucoup d'articles de ce livre dialoguent avec elle ; celui-ci vous donne la carte.

::: cle Pourquoi cette série compte
Avant ERN, le débat FIRE opposait des slogans (« 4 %, c'est réglé » contre « tout peut arriver »). La série a imposé une méthode. Des simulations exhaustives sur 150 ans de données mensuelles, tous les millésimes, des horizons de 50-60 ans. Et une honnêteté systématique sur ce qui marche, ce qui ne marche pas, et combien ça coûte. On peut être en désaccord avec ses conclusions (certains le sont, [[anarkulova-cederburg]]) ; on ne peut plus revenir aux slogans.
:::

## L'auteur et la méthode

Le cadre de travail d'ERN est constant sur toute la série. Données mensuelles américaines depuis 1871 (actions S&P composite, obligations d'État à 10 ans, via Shiller). Des retraites simulées à tous les mois de départ possibles, sur des horizons de 30 à 60 ans. Et un critère regardé de près, qui n'est pas seulement la ruine, mais aussi la richesse finale et le chemin, c'est-à-dire quand on échoue et après quoi. Les outils sont publics. La « toolbox » Google Sheets du volet 28 permet de refaire chaque calcul chez soi.

Deux partis pris sont à connaître pour la lire correctement. Le premier tient aux données, qui sont américaines. La série hérite donc du biais optimiste de l'échantillon ([[etude-trinity]], [[anarkulova-cederburg]]). ERN le sait, le dit, et considère que 150 ans de données mensuelles d'un grand marché, 1929 et 1966 compris, disciplinent déjà fortement les conclusions. Le second tient à une posture. ERN écrit pour le retraité précoce qui ne veut pas dépendre de la chance. De là vient une exigence de robustesse, souvent « survivre au pire millésime », plus dure que les 90-95 % de succès des praticiens ([[guardrails-morningstar]], [[ruine-et-probabilites]]).

## Les résultats majeurs, partie par partie

**Le socle : 4 % n'est pas fait pour vous (volets 1-3, 26).** C'est le résultat fondateur de la série. La règle des 4 % tient sur 30 ans. Mais un horizon de 50-60 ans exige plutôt 3,25-3,5 % en rigide, et le taux sûr dépend fortement des valorisations de départ (volet 3). À CAPE élevé, au-dessus de 20, les SAFEMAX historiques tombent, tous horizons confondus. Le volet 26 (« Ten Things the Makers of the 4% Rule Don't Want You to Know ») est le meilleur résumé grand public ([[la-regle-des-4-pourcents]], [[valorisations-et-cape]]).

**La séquence explique tout (volets 14-15).** La série y démontre, chiffres à l'appui, que le rendement moyen sur 30 ans compte moins que le rendement des 5-10 premières années. C'est le cœur de [[sequence-des-rendements]]. Corollaire du volet 53, l'épargnant et le rentier ont des expositions opposées à la séquence.

**La flexibilité est surestimée (volets 9-11, 23-25, 58).** C'est la série la plus contrariante, et la plus utile. Les règles de Guyton-Klinger (volets 9-10) affichent des taux de « succès » flatteurs. Mais dans les mauvais millésimes, elles les paient par des décennies de dépenses amputées de 30-45 %. La ruine est simplement remplacée par une pauvreté prolongée qu'aucun tableau de taux de succès ne montre ([[guyton-klinger]]). Les volets 23-25 et 58 généralisent le constat. Toute flexibilité réaliste, bornée et tenable, vaut quelques dixièmes de point de taux de retrait, pas la magie annoncée ([[flexibilite-realite]]). C'est ce résultat qui a poussé tout le domaine à afficher le niveau de vie vécu, pas seulement la survie du portefeuille (c'est l'objet de la §04).

**Le CAPE comme règle, pas comme peur (volets 18, 54).** C'est la proposition constructive de la série. Des règles de retrait fondées sur le CAPE, où le taux initial et courant s'ajuste aux valorisations, formalisées au volet 54. C'est l'ancêtre direct de l'ancre CAPE de la page FIRE ([[regles-cape]]).

**Les buckets et le cash ne font pas ce qu'on croit (volets 12, 48, 55).** C'est une analyse à contre-courant du dogme des buckets. Un matelas de cash consommé puis rechargé mécaniquement améliore peu la ruine, car le cash coûte en rendement ce qu'il économise en séquence. Et les stratégies de buckets populaires sont souvent du market timing déguisé, sans règle claire ([[cash-buffer]], [[strategie-buckets]], [[recharger-ou-pas]]). La nuance est importante. Une simulation trouve le même ordre de grandeur, l'arbitrage buffer de la §07 étant généralement plat. Cela n'enlève rien à la valeur psychologique et de gouvernance du buffer ([[psychologie-du-retrait]]).

**Les glidepaths marchent (volets 19-20, 43).** C'est la contrepartie positive. Partir avec 60 % d'actions et remonter vers 100 % sur 10-15 ans améliore nettement les pires millésimes, pour un coût modeste dans les bons. La protection se concentre sur la fenêtre fragile ([[glidepaths]], [[les-trois-phases]]).

**Rentes, obligations, or, immobilier, levier (volets 29-31, 34-36, 40, 49, 52, 56, 59).** Ce sont les inventaires d'actifs. L'illusion du rendement (yield illusion) des portefeuilles à dividendes, où vivre des dividendes n'est pas plus sûr (volets 29-31, [[erreurs-classiques-fire]]). L'or comme couverture partielle de séquence (volet 34, [[or-en-retrait]]). L'immobilier (volet 36, [[immobilier-en-retrait]]). Les rentes et la Sécurité sociale américaine, à adapter au système français (volets 56, 59, [[rentes-et-annuites]], [[retraite-legale]]). Le levier, sophistiqué et à petites doses (volets 49, 52, [[levier-et-marges]]).

**Les à-côtés qui changent la vie (volets 21-22, 27, 37, 42, 47, 60).** Le crédit immobilier en retraite, dont le garder est un pari de levier sur la séquence (volet 21). Le « one more year » chiffré (volet 42, [[une-annee-de-plus]]). Quand s'inquiéter en cours de route (volet 47, [[quand-s-inquieter]]). Traverser un marché baissier (volet 37, [[marche-baissier-en-retraite]]). Et la critique de « Die With Zero » (volet 60, [[depenses-en-retraite]]).

::: astuce Par où commencer, selon votre question
Le tour d'horizon → volets 1 et 26. « Combien puis-je retirer ? » → volets 2-3 puis 54 (CAPE). « La flexibilité me sauvera-t-elle ? » → volets 23-25 puis 58. « Mon allocation autour du départ » → volets 19-20. « Les buckets ? » → volets 12 et 48. Chaque volet se lit en 20-40 minutes. La série entière est de fait un livre, plus long que celui-ci, et elle le complète. ERN pousse les simulations américaines plus loin, tandis que ce livre couvre le cadre français, l'échantillon mondial et l'outillage du simulateur.
:::

## Lire ERN avec les bons filtres

Trois filtres pour un usage optimal depuis la France.

**Le filtre géographique.** Tout est américain : les données, la fiscalité, la Social Security, les comptes 401(k). Les mécanismes se transposent parfaitement, qu'il s'agisse de la séquence, du CAPE, des glidepaths ou de la flexibilité. Les chiffres, eux, héritent du biais d'échantillon et restent plutôt optimistes ([[anarkulova-cederburg]]). Quant au chapitre fiscal, il ne se transpose pas du tout ([[enveloppes-francaises]], [[taxe-puma]]).

**Le filtre de posture.** L'exigence « survivre au pire millésime historique » est une posture de prudence assumée, plus dure que le 90 % de succès des praticiens. Si vous avez des filets solides (pension, employabilité, flexibilité réelle), les recommandations d'ERN sont pour vous une borne prudente, pas un minimum vital ([[ruine-et-probabilites]], [[combien-il-vous-faut]]).

**Le filtre du perfectionnisme.** La série démontre qu'on peut raffiner sans fin, jusqu'au dixième de point de SWR ou à l'allocation à 5 % près. Gardez un rappel utile en la lisant. Les trois premiers leviers d'un plan réel restent les dépenses auditées, la pension comptée et une règle écrite ([[erreurs-classiques-fire]]). Le raffinement vient après.

::: terrain Ce que la communauté en a fait
La série a durablement changé la conversation FIRE. « Quel est ton SWR ? » a remplacé « tu as fait tes 25x ? ». Les simulateurs sérieux affichent désormais le niveau de vie vécu, et pas seulement le taux de succès. Et « ERN Part N » est devenu une référence courante sur les forums, signe d'une communauté qui a monté son niveau d'exigence. Un conseil d'expérience pour finir. Ne lisez pas les 60 volets d'affilée. Prenez ceux qui répondent à votre question du moment, puis revenez-y. C'est une encyclopédie, pas un roman.
:::

## L'essentiel à retenir

- La série SWR d'ERN ([earlyretirementnow.com](https://earlyretirementnow.com), 60+ volets depuis 2016) est la référence moderne du retrait de long horizon : méthodique, reproductible, gratuite.
- Ses résultats structurants : 3,25-3,5 % rigide pour 50-60 ans, le rôle dominant de la séquence et des valorisations, la flexibilité surestimée (elle déplace la douleur), les buckets démystifiés, les glidepaths validés.
- Ses partis pris : données américaines (chiffres plutôt optimistes), exigence de pire cas (recommandations plutôt prudentes). Les deux biais se compensent partiellement, sachez-le en la lisant.
- Depuis la France : gardez les mécanismes, adaptez les chiffres, ignorez le chapitre fiscal américain. Ce livre fait la jonction.
- Beaucoup de concepts de la page FIRE (ancre CAPE, niveau de vie vécu, stress de séquence, arbitrage buffer) sont en dialogue direct avec cette série ([[utiliser-la-page-fire]], [[la-machine-pofo]]).

---

## Pour aller plus loin

- Le point d'entrée : earlyretirementnow.com/safe-withdrawal-rate-series/ (la table des matières complète et tenue à jour de la série).
- La toolbox (volet 28) : le classeur de simulation public, pour refaire les calculs.
- Karsten Jeske en podcast (ChooseFI, Rational Reminder, Bogleheads) : la version orale, souvent plus accessible, des mêmes résultats.
- Les contrepoints dans ce livre : [[anarkulova-cederburg]] (l'échantillon au-delà des États-Unis) et [[guardrails-morningstar]] (la lecture praticienne, moins pire-cas).
