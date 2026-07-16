# Pourquoi 4 % ? L'anatomie mathématique de la règle

Le chapitre [[la-regle-des-4-pourcents]] raconte d'où vient la règle et [[etude-trinity]] comment elle a été établie. Celui-ci répond à la question qu'on pose rarement : pourquoi ce chiffre-là ? Pourquoi pas 6 %, puisque les actions rapportent 6 ou 7 % réels en moyenne ? Pourquoi pas 2 %, puisque tout peut arriver ? Le 4 % n'est pas une constante magique tombée d'un backtest, c'est le **résidu d'un calcul en cascade** dont chaque étage a un nom, un signe et un ordre de grandeur. Une fois la cascade comprise, la règle cesse d'être un dogme à croire ou à démentir, elle devient un résultat que vous savez recalculer quand une hypothèse change.

C'est probablement le chapitre le plus pédagogique du livre : quatre étages, une soustraction à chaque fois, aucun outil au-delà de la calculatrice. Après lecture, vous saurez reconstruire le chiffre de tête, dire précisément ce qui le ferait monter ou descendre, et comprendre pourquoi il s'est montré si étonnamment robuste à travers un siècle qui contenait 1929, 1966 et 2000.

::: cle L'idée en une phrase
Le taux de retrait sûr, c'est le rendement réel géométrique du portefeuille, **plus** un bonus parce qu'on a le droit de finir le capital (l'horizon est fini), **moins** une pénalité parce que les rendements arrivent dans le désordre et qu'on retire pendant les baisses (le risque de séquence). Sur un 60/40 historique et 30 ans, cela donne environ 4 + 1,8 − 1,8 → autour de 4 %. Tout le mystère de la règle tient dans le fait que le bonus et la pénalité, deux forces énormes et opposées, se compensent presque exactement.
:::

## Étage 1 : la rente perpétuelle, ou le rendement réel comme plafond naturel

Commençons par le monde le plus simple : un portefeuille au rendement constant et connu, et un rentier immortel qui veut ne jamais entamer son capital. La réponse tient en une ligne → le retrait soutenable est exactement le rendement **réel** du portefeuille. Réel, car le retrait doit suivre l'inflation ([[inflation-et-taux-de-retrait]]) ; tout raisonnement en nominal est une illusion d'optique qui coûte 2 à 3 points.

Reste à chiffrer ce rendement. Un 60/40 mondial a historiquement produit autour de 5 % réels en moyenne **arithmétique**, mais le capital ne compose pas la moyenne arithmétique, il compose la **géométrique**, amputée du volatility drag (environ la moitié de la variance, [[rendements-arithmetiques-geometriques]]). Résultat : environ **4 % réels géométriques** pour le 60/40 historique. Premier jalon posé : l'immortel prudent au portefeuille classique peut retirer environ 4 %, et ce chiffre est un rendement, pas encore un taux de retrait. La coïncidence numérique avec la règle finale est un hasard qui entretient la confusion, car les deux étages suivants vont s'annuler presque parfaitement.

## Étage 2 : le bonus d'amortissement, ou le droit de mourir

Le rentier n'est pas immortel. Sur un horizon de 30 ans, il a le droit de consommer le capital lui-même, pas seulement ses fruits. La finance de crédit donne la formule exacte (celle d'une mensualité de prêt, inversée) : à 4 % réels sur 30 ans, le retrait qui épuise exactement le capital au dernier jour est d'environ 5,8 % par an. Le droit de finir à zéro vaut donc +1,8 point, c'est le **bonus d'amortissement**, et c'est toute la logique des méthodes ABW/VPW qui le recalculent chaque année ([[amortissement-abw]], [[vpw]]).

Ce bonus explique au passage la sensibilité à l'horizon ([[horizon-et-esperance-de-vie]]) : à 50 ans d'horizon, le même calcul donne 4,7 % (le bonus fond à +0,7) ; à l'infini, il disparaît. La retraite très précoce ne « perd » pas son taux à cause d'un mystère, elle perd simplement le droit d'amortir vite.

## Étage 3 : la pénalité de séquence, ou le désordre qui coûte

Tout l'étage 2 supposait un rendement constant. Les vrais rendements arrivent en désordre, et le retrait fixe transforme ce désordre en risque asymétrique : vendre pendant les baisses détruit du capital qui ne verra pas le rebond ([[sequence-des-rendements]]). Mathématiquement, les retraits font que le résultat final ne dépend plus seulement de la moyenne géométrique des rendements mais aussi de leur ordre, les premières années pesant plusieurs fois plus que les dernières. Et comme la règle veut survivre aux pires ordres observés (1929, 1966), et non à l'ordre moyen, il faut provisionner le désordre maximal.

Combien coûte-t-il ? C'est la découverte centrale de Bengen : sur l'histoire américaine, le pire millésime ne supportait qu'environ 4 %, alors que le calcul d'amortissement au rendement moyen en promettait 5,8 ; la pénalité de séquence historique vaut donc environ −1,8 point. Elle n'est pas une constante universelle : elle grandit avec la volatilité du portefeuille (un 100 % actions a à la fois un meilleur rendement moyen et une pénalité plus lourde, ce qui explique le plateau d'allocation, [[allocation-actions-obligations]]), et elle se réduit avec tout ce qui amortit les baisses → diversification et rééquilibrage (qui force à acheter bas pendant le krach, [[pourquoi-la-diversification-marche]]), et flexibilité du retrait (couper 10 % en année rouge rachète une partie de la pénalité, [[flexibilite-realite]]).

::: exemple La cascade complète, sur un coin de table
60/40 historique, 30 ans, retrait fixe indexé. Rendement réel arithmétique ≈ 5 % → moins le volatility drag ≈ 4 % géométrique (étage 1) → plus le bonus d'amortissement 30 ans ≈ 5,8 % (étage 2) → moins la pénalité du pire ordre historique ≈ 4,0 % (étage 3). Voilà la règle, reconstruite. Maintenant, faites-la respirer : horizon 50 ans → le bonus fond, ~3,4 % ; échantillon mondial au lieu du seul cas américain ([[anarkulova-cederburg]]) → retirer 0,3-0,7 ; CAPE élevé au départ ([[valorisations-et-cape]]) → étage 1 raboté d'un point ; 0,5 % de frais → −0,5, presque un pour un ; règle flexible à plancher → +0,3-0,5. Chaque débat sur « le vrai chiffre » de la règle est un débat sur un seul étage de la cascade, et il se règle étage par étage, pas par slogans.
:::

## Pourquoi c'est si robuste (et ce qui le casserait)

La vraie surprise de la règle n'est pas sa valeur, c'est sa stabilité : elle a tenu à travers deux guerres mondiales, une dépression, une stagflation et deux krachs boursiers du XXIe siècle. Trois mécanismes l'expliquent, tous déjà rencontrés. La **compensation des étages** d'abord : les époques à rendements faibles sont souvent des époques à valorisations basses ensuite, donc à rendements futurs meilleurs, la cascade se ré-alimente en cours de route (c'est la version mécanique du retour à la moyenne). Le **rééquilibrage** ensuite, qui transforme chaque krach en achat forcé au rabais, rognant la pénalité de séquence exactement quand elle mord. La **marge cachée** enfin : la règle est calibrée sur le pire cas historique, si bien que dans 90 % des millésimes elle laisse un capital final supérieur au capital initial ([[lire-un-fan-chart]]), la médiane subventionne la queue.

La liste de ce qui la casserait est, du coup, précise. Un régime **hors échantillon** pour votre pays (le retraité de Tokyo 1990 ou de Buenos Aires a connu pire que le pire américain, [[diversification-internationale]], [[pieges-des-simulateurs]]). Des **frottements** non provisionnés, car la cascade est calculée brute → frais et fiscalité se soustraient de l'étage 1 presque un pour un ([[etf-ucits-europeens]], [[flat-tax-et-imposition]]). Une **indexation rigide** sur un horizon très long, la combinaison qui a le moins de marge. Et l'**abandon en cours de route**, statistiquement le tueur numéro un, contre lequel la cascade ne peut rien ([[psychologie-du-retrait]]).

::: science Ce que les reconstructions modernes confirment
La cascade n'est pas qu'une pédagogie, elle se vérifie terme à terme. Les reconstructions d'ERN (série SWR, [[serie-ern]]) retrouvent la pénalité de séquence dans la dépendance du taux au CAPE initial et à l'allocation ; Blanchett, Pfau et Morningstar retrouvent le bonus d'amortissement dans leurs taux croissants avec l'âge (les taux « sûrs » publiés montent d'environ 0,3-0,5 point par tranche de cinq ans d'horizon en moins) ; et Anarkulova-Cederburg chiffrent le terme « échantillon » en remplaçant l'histoire américaine par le panier mondial (−0,5 à −1 point selon le critère). Quand ce livre recommande un point de départ prudent vers 3,3-3,7 % pour un FIRE long plutôt que le 4 % canonique, ce n'est pas une humeur, c'est la même cascade avec les étages remis à leurs valeurs prospectives ([[rendements-attendus]]).
:::

## L'essentiel à retenir

- Le 4 % se reconstruit en trois étages : ~4 % de rendement réel géométrique du 60/40 (après volatility drag), +1,8 point de bonus d'amortissement sur 30 ans, −1,8 point de pénalité pour le pire ordre de rendements observé. La quasi-annulation du bonus et de la pénalité est le « miracle » de la règle.
- Chaque étage a ses leviers → l'horizon joue sur le bonus (50 ans : +0,7 seulement), la volatilité et la flexibilité jouent sur la pénalité, les valorisations, les frais et l'échantillon jouent sur l'étage rendement, presque un pour un pour les frais.
- La robustesse historique vient du retour à la moyenne (les mauvaises décennies préparent les bonnes), du rééquilibrage (achat forcé au creux) et de la marge du pire cas (la médiane finit plus riche qu'au départ).
- Ce qui casse la cascade est connu : régime hors échantillon national, frottements non provisionnés, indexation rigide sur très long horizon, abandon en route.
- Savoir refaire ce calcul de tête vaut tous les débats : quand quelqu'un annonce « le 4 % est mort » ou « 5 % passe très bien », demandez-lui quel étage il a modifié, et de combien.

---

## Pour aller plus loin

- William Bengen, « Determining Withdrawal Rates Using Historical Data » (1994) : l'article original, où la cascade est implicite dans les tables.
- Early Retirement Now, série SWR (notamment les volets sur le CAPE et l'horizon) : la cascade recalculée sur données modernes ([[serie-ern]]).
- Blanchett, « Exploring the Retirement Consumption Puzzle » et les rapports State of Retirement Income de Morningstar : le bonus d'amortissement dans les taux par âge.
- Dans ce livre : [[rendements-arithmetiques-geometriques]] (l'étage 1 en détail), [[amortissement-abw]] (l'étage 2 érigé en stratégie), [[sequence-des-rendements]] (l'étage 3), [[la-regle-des-4-pourcents]] (la règle racontée) et [[etude-trinity]] (la règle mesurée).
