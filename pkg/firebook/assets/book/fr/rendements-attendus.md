# Les rendements attendus prospectifs (Morningstar, Vanguard, banques d'investissement)

Toute la mécanique d'un plan de retrait repose sur un nombre que personne ne connaît : le rendement réel que vos actifs délivreront pendant les décennies à venir. Face à cette inconnue, deux postures s'affrontent.

Le rétroviseur : « les actions ont fait 7 % réel depuis 1900, je mets 7 % ». Et l'approche prospective : estimer ce que les prix, les rendements courants et les fondamentaux d'**aujourd'hui** permettent de promettre, c'est le métier des « capital market assumptions » (CMA) que publient chaque année Vanguard, Morningstar, BlackRock, Research Affiliates, GMO, AQR et les grands fonds de pension.

Cette page explique pourquoi le rétroviseur est le pire des deux (surtout pour un rentier), comment les prévisions prospectives se fabriquent (vous saurez refaire le calcul vous-même, il tient sur une enveloppe), ce que disent les fourchettes actuelles, la précision qu'on peut en attendre (médiocre, mais bien meilleure que l'alternative), et comment injecter tout cela dans les curseurs sans empiler les prudences en double. À la fin, vous saurez produire et défendre **votre** μ.

::: cle Le renversement à opérer
Le rendement passé d'un actif n'est pas son espérance ; c'est souvent son **inverse** partiel : une décennie exceptionnelle est en grande partie une expansion de valorisation, c'est-à-dire du rendement futur consommé par avance. À l'inverse, les obligations affichent littéralement leur espérance sur l'étiquette (le rendement à l'échéance), et les actions l'affichent à moitié (l'earnings yield, [[valorisations-et-cape]]). Estimer prospectivement, ce n'est pas prédire. C'est lire ce qui est déjà écrit dans les prix.
:::

## Pourquoi le rétroviseur trompe, et doublement pour un rentier

L'argument anti-rétroviseur tient en trois constats.

**Premier constat : le passé long contient son propre biais.** Les « 7 % réels des actions » sont américains ([[anarkulova-cederburg]]) et incluent un vent arrière non répétable : sur le siècle, le CAPE américain est passé d'environ 15 à plus de 30, et les taux réels ont fini en baisse séculaire. Une partie de la performance historique est une revalorisation ponctuelle, pas un rendement d'exploitation. Le monde entier hors États-Unis a fait ~4,5 % réel sur la même période, et c'est déjà avec des survivants.

**Deuxième constat : le passé court est encore pire.** Calibrer sur les 10-20 dernières années de ses fonds (le réflexe naturel, et ce que ferait naïvement un simulateur ajusté sur votre historique) capture une fenêtre particulière : celle qui vous a précisément amené à la cible, donc probablement favorable. C'est le raisonnement qui a fait entrer des générations d'épargnants en retraite avec des hypothèses de bulle. Le modèle se défend structurellement contre ce biais : les paramètres ajustés sur votre portefeuille sont **mélangés** vers un prior mondial prudent à proportion de ce que l'horizon excède l'historique disponible ([[rendre-monte-carlo-pertinent]], [[la-machine-pofo]]).

**Troisième constat, spécifique au rentier : l'erreur est asymétrique dans le temps.** Pour un épargnant, se tromper d'un point de rendement espéré décale la date d'arrivée. Pour un rentier, le rendement des **dix premières** années domine le sort du plan ([[sequence-des-rendements]]), et c'est exactement l'horizon où les mesures de valorisation ont leur pouvoir prédictif maximal ([[valorisations-et-cape]]). Ignorer l'information prospective, c'est jeter le seul éclairage disponible sur la décennie qui décide.

## Comment se fabrique une prévision : les briques

Toutes les CMA sérieuses, malgré leurs machineries différentes, reviennent à la même décomposition dite « building blocks » (Bogle l'avait popularisée dès 1991) : le rendement d'un actif = ce qu'il distribue + ce qu'il croît + ce que sa valorisation fait. Détaillons par classe, parce que vous pouvez faire ce calcul vous-même.

**Obligations : lisez l'étiquette.** L'espérance de rendement nominal d'une obligation (ou d'un fonds obligataire de duration constante) à son horizon de duration est, à peu de chose près, son **rendement à l'échéance courant** (YTM). C'est la seule classe d'actifs dont l'espérance est quasiment observable : la corrélation historique entre le YTM de départ et le rendement réalisé à 10 ans dépasse 0,9. Pour le rendement **réel** : YTM moins l'inflation anticipée (le point mort d'inflation), ou directement le rendement réel des obligations indexées ([[obligations-indexees]]). Exemple : OAT/Bund 10 ans à ~3 %, point mort ~2 % → ~1 % réel ; TIPS américains à ~2 % réel → 2 % réel, affiché tel quel. Conséquence importante : l'espérance obligataire **bouge** avec les taux, beaucoup ; celle de 2021 (taux réels négatifs, espérance réelle négative garantie) et celle de 2024-2026 (~1,5-2 % réel) ne se ressemblent pas, et un plan calibré sur « les obligations font 5 % » (leur moyenne nominale historique, gonflée par la désinflation 1982-2021) est calibré sur un monde disparu.

**Actions : trois briques.** Rendement réel ≈ rendement de distribution (dividendes + rachats nets, ~2-2,5 % pour le S&P actuel, ~3 % pour l'Europe) + croissance réelle des bénéfices par action (~1,5-2 % de tendance longue, davantage dans les phases d'expansion de marges, mais les marges ne montent pas au ciel) + variation de valorisation (le terme querelleur, 0 si l'on suppose les multiples stables, négatif si l'on suppose un retour partiel vers des CAPE plus normaux). Avec les chiffres de la zone 2024-2026 : actions américaines ~2 + 1,75 + (0 à −1,5) = **2,5 à 4 % réel** ; actions non américaines, moins chères et mieux distribuantes, **4 à 6 % réel**. Vous venez de refaire, à un point près, les tables de Vanguard et Research Affiliates.

**Monétaire et or.** Le cash suit le taux directeur réel : autour de 0 à 1 % réel en zone euro, cycliquement variable, espérance de long terme ~0. L'or n'a ni coupon ni bénéfices : son espérance réelle de très long terme est proche de 0 à 1 %, ce qui n'est **pas** son rôle dans un portefeuille (il se paie en espérance ce qu'il rend en couverture de régimes, [[or-en-retrait]], [[actifs-defensifs]]).

::: exemple Construire le μ d'un portefeuille 70/30, sur une enveloppe
Portefeuille : 70 % actions mondiales (dont ~65 % US), 30 % obligations euro aggregate. Actions : pondérons 0,65 × 3 % (US, hypothèse médiane) + 0,35 × 5 % (reste du monde) ≈ 3,7 % réel. Obligations : YTM ~3,2 %, inflation anticipée ~2 % → 1,2 % réel. Portefeuille : 0,7 × 3,7 + 0,3 × 1,2 ≈ **2,95 % réel géométrique attendu**. Pour le curseur μ, qui demande la moyenne **arithmétique** ([[utiliser-la-page-fire]]) : ajoutez σ²/2, soit ~0,6 point à σ = 11 % → **μ ≈ 3,5 %**. Comparez ce chiffre aux 5 % par défaut et à ce que votre historique de fonds suggère : l'écart est le débat, et il vaut d'être tranché consciemment plutôt que subi.
:::

## Qui publie quoi, et ce que disent les fourchettes

Le paysage des prévisionnistes, avec leurs méthodes et leurs biais connus :

| Maison | Méthode dominante | Style | Où la trouver |
|---|---|---|---|
| Vanguard (VCMM) | Modèle stochastique, fourchettes à 10 ans | Centriste, publie des intervalles honnêtes | Rapport annuel « economic and market outlook » (gratuit) |
| Morningstar (ex-Ibbotson) | Building blocks + valorisation | Centriste ; alimente son étude SWR annuelle | « State of Retirement Income » (gratuit) |
| Research Affiliates | Retour à la moyenne des valorisations | Structurellement prudent sur les actifs chers | Asset Allocation Interactive (gratuit, interactif, par pays) |
| GMO | Retour à la moyenne agressif à 7 ans | Le plus pessimiste sur les actifs chers ; historique de sous-estimation des bulles... et de leur éclatement | Lettres trimestrielles (gratuit) |
| AQR | Primes de risque théoriques + valorisations | Rigueur académique ; « Capital Market Assumptions » annuelles | Papers (gratuit) |
| BlackRock, JP Morgan, banques | CMA institutionnelles à 10-15 ans | Plutôt lisses, usage allocation d'actifs | « Long-Term Capital Market Assumptions » (gratuit) |
| Fonds de pension (ex. néerlandais, canadiens) | Hypothèses réglementaires prudentielles | La borne engageante : ils **paient** s'ils se trompent | Rapports actuariels publics |

Ce tableau appelle trois lectures. D'abord la **convergence des ordres de grandeur** dans la zone 2024-2026 : actions américaines 2 à 4,5 % réel selon le poids donné au retour de valorisation, actions internationales 4 à 6 %, obligations de qualité 1,5 à 2,5 % réel, cash ~0,5 %, soit un 60/40 mondial autour de **2,5 à 3,5 % réel géométrique**. Ensuite la **dispersion résiduelle** : entre GMO et la plus optimiste des banques, il y a couramment 3 points d'écart sur les actions américaines ; personne ne « sait », et un plan qui exige de trancher entre eux au point près est un plan trop tendu. Enfin le **signal des institutions engagées** : les taux d'actualisation réels des grands fonds de pension (2,5-4 % réel pour des portefeuilles diversifiés avec parts d'illiquide) sont une borne haute raisonnable de ce que des professionnels acceptent de **promettre** ; un particulier qui met 6 % réel dans son simulateur promet plus que CalPERS.

**Le cas Morningstar mérite un paragraphe**, parce qu'il boucle directement sur le taux de retrait. Chaque année depuis 2021, « The State of Retirement Income » recalcule le taux de retrait initial recommandé (30 ans, 90 % de succès, portefeuille équilibré) à partir de leurs rendements **prospectifs**, et la série vaut leçon : **3,3 % en 2021** (marchés chers, taux nuls), **3,8 % en 2022** (les valorisations avaient dégonflé), **4,0 % en 2023** (taux obligataires restaurés), **~3,7 % en 2024-2025** (actions redevenues chères). Le chiffre **bouge**, et c'est le message le plus profond de l'exercice : le taux de retrait soutenable n'est pas une constante universelle mais une fonction des conditions d'entrée ([[valorisations-et-cape]]), et une maison sérieuse assume de le republier chaque année. Utilisez leur dernier millésime comme deuxième avis gratuit sur votre propre calibration ([[guardrails-morningstar]] pour leur cadre complet).

::: science Quelle précision en attendre ?
Les études rétrospectives sur les CMA (notamment celles de Morningstar sur ses propres archives et les comparatifs académiques) donnent un verdict nuancé : à 10 ans, les prévisions prospectives ont une erreur moyenne substantielle (± 2-3 points) mais battent nettement le rétroviseur naïf et surtout, elles se trompent **moins souvent** dans le sens dangereux (surestimer après une bulle). Le classement de fiabilité par classe : obligations (excellent, le YTM est un quasi-contrat), cash (bon), actions (utile mais bruité, le terme de valorisation domine à 10 ans et reste incertain), alternatives (médiocre). Traduction pour votre plan : prenez les fourchettes au sérieux, pas les points ; et rappelez-vous que la ruine, elle aussi, se lit en intervalle ([[ruine-et-probabilites]]).
:::

## Injecter tout cela dans pofo, sans double-compter la prudence

La page FIRE vous donne quatre mécanismes qui touchent à l'espérance de rendement, et le piège le plus courant du planificateur consciencieux est de les **empiler** ([[utiliser-la-page-fire]]) :

1. **Le μ ajusté sur votre historique**, mélangé automatiquement vers le prior mondial prudent (μ 4,5 %, σ 13, df 4) à proportion de l'horizon. C'est déjà une correction anti-rétroviseur.
2. **La case « Broad-sample prior »**, qui réécrit les curseurs avec les hypothèses du siècle mondial (~3,5 % réel géométrique) : une deuxième couche de prudence, plus dure.
3. **L'ancre CAPE**, qui remplace la seule moyenne par 1/CAPE : la correction prospective de cette page, appliquée à la brique actions.
4. **La colonne broad-sample du tableau**, qui ignore vos curseurs et rejoue le siècle des 16 pays : la borne empirique, toujours visible quoi que vous fassiez.

La discipline recommandée : choisissez une calibration centrale et assumez-la, puis lisez les autres comme des bornes. Concrètement : soit vous faites confiance au blending automatique (défaut raisonnable), soit vous entrez votre μ building-blocks à la main (l'exemple ci-dessus), soit vous cochez l'ancre CAPE (équivalent automatisé du building-blocks pour la brique actions). Mais cocher l'ancre CAPE **puis** baisser encore μ à la main **puis** ne juger le plan que sur la colonne broad-sample, c'est compter la même prudence trois fois : le plan exigera des années de travail superflues, ce qui est aussi une erreur de planification ([[une-annee-de-plus]]). La prudence se budgète comme le reste.

Notez enfin ce que l'exercice prospectif **ne** remplace pas : la volatilité et les queues (σ, df) ne se lisent pas dans les valorisations ; elles viennent de l'histoire et de la structure du portefeuille, et pofo les ajuste sur vos fonds ([[queues-epaisses]]). L'espérance dit où mène la route en moyenne ; σ et df disent à quel point le trajet secoue, et pour un rentier les deux comptent presque autant ([[rendements-arithmetiques-geometriques]]).

## Les erreurs de manipulation courantes

**Confondre les conventions.** Un « 5 % » peut être arithmétique ou géométrique, réel ou nominal, brut ou net : les CMA publiées mélangent les conventions d'une maison à l'autre (Vanguard publie du nominal géométrique à 10 ans, AQR du réel, Research Affiliates du réel). Convertissez **tout** en réel géométrique avant comparaison, puis en arithmétique (+σ²/2) pour le curseur. Les trois questions réflexes de [[rendements-arithmetiques-geometriques]] s'appliquent d'abord aux prévisionnistes.

**Sur-réagir au millésime.** Les CMA bougent chaque année ; votre plan ne doit pas tanguer avec elles. Le bon rythme : une relecture par revue annuelle ([[revue-annuelle]]), une action seulement si le paysage a changé de régime (comme 2021 → 2023 pour les obligations), pas au gré des +0,3/−0,3.

**Oublier que le futur peut être pire que toutes les prévisions.** Les CMA sont des espérances centrales : aucune ne « contient » 1929, le Japon 1990 ou une guerre. C'est le rôle des autres colonnes (stress de séquence, décennie perdue, broad-sample) et des protections structurelles ([[portefeuilles-tous-temps]], [[flexibilite-realite]]). L'espérance prospective calibre le centre ; elle ne remplace ni les queues ni les marges.

**Et l'inverse : le pessimisme perpétuel.** GMO annonce des rendements américains quasi nuls depuis 2013 ; qui l'a suivi intégralement a raté la meilleure décennie de l'histoire récente. La leçon n'est pas « les pessimistes ont tort » (leur heure finit par arriver, 2000 et 2008 l'ont montré) mais : ne confiez jamais le **portefeuille** aux prévisions ; confiez-leur le **plan** (le taux, les marges, les attentes). La différence entre ces deux usages est exactement celle décrite pour le CAPE ([[valorisations-et-cape]]).

## L'essentiel à retenir

- Le rendement passé n'est pas l'espérance ; pour les obligations l'espérance est affichée (YTM réel), pour les actions elle se construit : distribution + croissance + terme de valorisation.
- Zone 2024-2026, en réel géométrique : actions US 2-4,5 %, internationales 4-6 %, obligations 1,5-2,5 %, 60/40 mondial 2,5-3,5 %. Un μ de simulateur au-dessus de ces fourchettes se justifie, ou se corrige.
- Morningstar recalcule chaque année le taux de retrait recommandé sur ces bases (3,3 → 3,8 → 4,0 → ~3,7 % depuis 2021) : la preuve vivante que le taux soutenable dépend des conditions d'entrée.
- La précision est en fourchettes (± 2-3 points à 10 ans sur les actions), mais bat le rétroviseur, surtout dans le sens qui protège : après les bulles.
- Dans pofo : une seule calibration centrale (blending par défaut, building-blocks manuel, ou ancre CAPE), les autres vues en bornes ; n'empilez pas trois fois la même prudence, elle se paie en années de travail.

---

## Pour aller plus loin

- Vanguard, « Economic and Market Outlook » (annuel) ; Morningstar, « The State of Retirement Income » (annuel) ; JP Morgan et BlackRock, « Long-Term Capital Market Assumptions » : les quatre lectures gratuites qui couvrent le paysage.
- Research Affiliates, « Asset Allocation Interactive » : les espérances par pays et classes, recalculées en continu ; l'outil interactif le plus pédagogique.
- John Bogle, « Investing in the 1990s » (1991) et *Common Sense on Mutual Funds* : la décomposition building-blocks originale (« Occam's razor »).
- Antti Ilmanen (AQR), *Expected Returns* (2011) : la référence livre du sujet, exhaustive et lisible.
- Dans ce livre : [[valorisations-et-cape]] (la brique actions en détail), [[rendements-arithmetiques-geometriques]] (les conversions), [[rendre-monte-carlo-pertinent]] (comment pofo mélange votre historique et les priors), [[primes-de-risque]] (pourquoi ces rendements existent et devraient persister).
