# Les régimes de marché (croissance × inflation, ours collants) et pourquoi ils comptent

Les rendements des marchés ne tombent pas d'une urne homogène : ils sont produits par une économie qui traverse des **saisons**, longues de plusieurs années, pendant lesquelles presque tout se comporte différemment : la croissance des années 1990, la stagflation des années 1970, la déflation des années 1930, la décennie perdue japonaise. Ces saisons, les régimes de marché, sont la raison profonde de presque tout ce que les articles précédents ont constaté en surface : les grappes de mauvaises années ([[sequence-des-rendements]]), les queues épaisses ([[queues-epaisses]]), l'insuffisance des tirages indépendants ([[pieges-des-simulateurs]]).

Et elles sont la clé de la question qui domine la partie portefeuille de ce livre : pourquoi un 60/40 classique, si robuste dans certaines décennies, se fait-il détruire dans d'autres, et que faut-il détenir pour survivre à **toutes** les saisons ? Cette page pose le cadre : la preuve que les régimes existent et persistent, la grille croissance × inflation qui les classe (le langage commun de Browne, Dalio et de la recherche macro), ce que chaque régime fait à chaque classe d'actifs, et les deux conséquences pratiques : comment on modélise des régimes (les ours collants de pofo) et comment on s'y prépare sans prétendre les prévoir.

::: cle L'idée en une phrase
Un portefeuille n'est jamais neutre : il est un pari implicite sur un régime (le 60/40 parie sur la croissance désinflationniste), et un plan de retrait est une promesse de 40 ans qui traversera **statistiquement** deux ou trois régimes hostiles. La question n'est pas d'éviter les mauvaises saisons (personne ne les prédit de façon fiable) mais de détenir de quoi survivre à chacune, et de les avoir mises dans le simulateur avant qu'elles n'arrivent dans la vie.
:::

::: figure regime-grid
La grille croissance × inflation : chaque saison a ses gagnants. Le 60/40 ne couvre que la colonne de gauche ; la stagflation (bas-droite) est le trou de la défense classique et le cauchemar du rentier.
:::

## Les régimes existent : la preuve en trois faits

**Fait 1 : les marchés baissiers sont des épisodes, pas des accidents.** Sur un siècle américain, les grands marchés baissiers d'actions (−20 % réel et plus) durent en médiane un à deux ans, mais leur queue est longue : 1929-1932 (34 mois, −80 % réel), 1973-1974 (21 mois, −55 % réel en comptant l'inflation), 2000-2002 (30 mois), 2007-2009 (17 mois). Et le retour au sommet **réel** précédent prend bien plus longtemps : 7 ans après 1974, 13 ans après 2000, plus de 30 ans au Japon après 1990. Un monde i.i.d. produirait des baisses de cette ampleur, mais pas ces **durées** : l'enchaînement persistant d'années médiocres est la signature des régimes.

**Fait 2 : la volatilité et les corrélations changent d'état.** La volatilité fait des grappes (des années entières à 25 % d'agitation succèdent à des années à 10 %), et la corrélation actions-obligations elle-même change de **signe** selon le régime : négative de 2000 à 2021 (les obligations amortissaient les krachs d'actions : l'âge d'or du 60/40), **positive** dans les décennies inflationnistes (1970s, et brutalement 2022 : actions −20 %, obligations longues −30 %, la pire année du 60/40 depuis un siècle). Ce basculement de corrélation est le fait de marché le plus important pour un rentier moderne, parce que toute la protection « classique » de son portefeuille en dépend ([[obligations-en-retrait]]).

**Fait 3 : l'économétrie le confirme.** Depuis Hamilton (1989), les modèles à changement de régime (chaînes de Markov cachées) battent les modèles homogènes pour décrire les rendements : les données **préfèrent** une description en états persistants (expansion/récession, calme/crise) avec des probabilités de transition faibles : on ne sort pas d'un régime facilement, on y colle. C'est exactement la structure que la colonne « Sequence stress » de pofo emprunte ([[rendre-monte-carlo-pertinent]]).

D'où viennent ces régimes ? De ce que l'économie elle-même a des états persistants : les cycles de crédit et d'endettement se construisent sur des années, les politiques monétaires corrigent avec retard puis surcorrigent, les récessions s'auto-entretiennent (licenciements → moins de demande → licenciements), et la psychologie collective amplifie tout (l'euphorie fabrique les valorisations chères, [[valorisations-et-cape]], la peur fabrique les planchers). Les marchés héritent de la persistance du monde réel qu'ils actualisent.

## La grille croissance × inflation : les quatre saisons

Pour classer les régimes, le cadre le plus fécond croise deux surprises macroéconomiques : la **croissance** (au-dessus ou en dessous des attentes) et l'**inflation** (idem). Quatre quadrants, quatre saisons, chacune avec ses gagnants et ses victimes. C'est le langage commun d'Harry Browne (le Permanent Portfolio, 1981), de Ray Dalio (l'All-Weather, années 1990) et de la recherche macro moderne ([[portefeuilles-tous-temps]]).

| Régime | Croissance | Inflation | Ce qui gagne | Ce qui souffre | Épisodes types |
|---|---|---|---|---|---|
| **Prospérité** (désinflation boom) | + | − | Actions, obligations, tout ce qui est long | Or, cash (coût d'opportunité) | 1982-1999, 2009-2021 |
| **Surchauffe / inflation** | + | + | Matières premières, or, actifs réels, linkers | Obligations nominales, actions chères | 1965-1969, 2021-2022 |
| **Stagflation** | − | + | Or, linkers, cash rémunéré, TSMOM | **Tout** le reste : actions et obligations | 1973-1981, le cauchemar du rentier |
| **Déflation / bust** | − | − | Obligations d'État longues, cash | Actions, crédit, immobilier, matières premières | 1929-1938, Japon 1990s, 2008 |

Trois lectures de cette grille valent la peine d'être détaillées.

**Le 60/40 est un pari sur la colonne de gauche.** Actions et obligations nominales gagnent ensemble en prospérité et se couvrent mutuellement en déflation (les obligations montent quand les actions krachent : 2008). Mais la **ligne** inflationniste les frappe **ensemble** : l'inflation détruit la valeur réelle des coupons obligataires et comprime les multiples d'actions. Un portefeuille classique n'est donc pas « diversifié » au sens des régimes : il est diversifié à l'intérieur de deux quadrants sur quatre. Tant que l'inflation dort (1982-2021 !), personne ne s'en aperçoit ; 2022 a été la piqûre de rappel générationnelle.

**Le pire quadrant du rentier est la stagflation, et ce n'est pas un hasard si le pire millésime est 1966.** Relisez [[etude-trinity]] avec la grille en main : le retraité de 1966 n'a pas subi un krach, il a subi **quinze ans** de quadrant hostile : actions réelles nulles, obligations laminées par l'inflation, et des retraits indexés qui gonflaient pendant que tout le reste rétrécissait ([[inflation-et-taux-de-retrait]]). La stagflation cumule les trois plaies du rentier : rendements réels négatifs sur les deux actifs classiques, retraits qui montent, et durée. C'est le régime contre lequel un portefeuille de retrait doit être **explicitement** armé : or, obligations indexées, actifs réels, stratégies de tendance ([[or-en-retrait]], [[obligations-indexees]], [[managed-futures]]).

**Aucun actif ne gagne partout, mais chaque régime a ses gagnants.** C'est le fondement logique des portefeuilles tous-temps : plutôt que maximiser le rendement du quadrant probable (le réflexe de l'accumulation), détenir en permanence au moins un actif gagnant par quadrant, et laisser le rééquilibrage vendanger les rotations ([[portefeuilles-tous-temps]], [[actifs-defensifs]]). Le coût est un rendement espéré central plus faible ; le gain est une distribution **resserrée**, exactement l'échange que la mathématique du retrait récompense ([[rendements-arithmetiques-geometriques]], [[sequence-des-rendements]]).

::: attention Peut-on prévoir le régime ? L'honnêteté d'abord
La tentation naturelle : identifier le régime courant et surpondérer ses gagnants. La recherche est nuancée : la détection du régime **en cours** est partiellement faisable (les indicateurs macro type production industrielle, inflation et pente des taux classent raisonnablement le présent), mais la **prévision** des bascules est restée largement hors de portée, et les stratégies tactiques grand public détruisent en moyenne de la valeur par leurs retards et leurs faux signaux. Des approches quantitatives disciplinées existent (le Permanent Portfolio tactique de Darcet, que pofo a étudié et outillé par ailleurs, module les quatre poches selon le régime macro mesuré ; les managed futures sont, de fait, une détection de régime par les prix, [[managed-futures]]), mais elles exigent une exécution mécanique sans états d'âme. Pour l'immense majorité des plans : la préparation structurelle (tous-temps) bat la prédiction, et la grille sert à **auditer** le portefeuille, pas à le faire tourner.
:::

## Modéliser des régimes : ce que fait pofo, et pourquoi

La conséquence directe des régimes pour la simulation a déjà été posée ([[rendre-monte-carlo-pertinent]]) ; résumons-la du point de vue de cette page, car c'est ici qu'elle se justifie.

**Le stress de séquence est une machine à régimes minimaliste.** Deux états (normal/ours), des transitions collantes calibrées pour produire ~19 % d'années d'ours en épisodes d'environ trois ans, une volatilité amplifiée dans l'ours, une moyenne de long terme préservée : c'est Hamilton réduit à l'os, juste assez de structure pour restituer la propriété des régimes qui tue les rentiers (la persistance des mauvaises passes) sans prétendre modéliser la macro. L'écart central/stress mesure votre exposition à cette persistance.

**Le broad-sample contient les régimes en vrai.** Le bootstrap par blocs pays entiers du modèle broad-sample ([[historique-vs-parametrique]]) fait traverser à vos trajectoires de **vraies** stagflations (le Royaume-Uni 1970s), de **vraies** déflations (les années 1930), de **vrais** décrochages nationaux (le Japon), avec leurs corrélations actions-obligations d'époque. C'est le seul modèle de la page où la ligne inflationniste de la grille existe à l'état natif : d'où son rôle de juge de paix pour les portefeuilles trop « colonne de gauche » ([[anarkulova-cederburg]]).

**La décennie perdue est un quadrant devenu scénario.** Le régime « bust long » isolé et poussé à sa durée japonaise : le crash-test du quadrant sud, à rendre survivable ([[marche-baissier-en-retraite]]).

Ce que la page ne modélise volontairement **pas** : la rotation fine des classes d'actifs par quadrant (le portefeuille est agrégé avant simulation). La grille des régimes sert en **amont**, au moment de composer le portefeuille ; le simulateur teste ensuite la robustesse de la composition retenue. Les deux étages se complètent : composer avec la grille, tester avec les colonnes.

::: exemple Auditer un portefeuille avec la grille
Portefeuille : 70 % actions mondiales, 30 % obligations d'État euro nominales 7-10 ans. Audit par quadrant : prospérité, excellent (les deux gagnent) ; déflation, bon (la duration amortit) ; surchauffe, médiocre (les deux souffrent, modérément) ; stagflation, catastrophique (les deux perdent en réel, durablement : **aucune** poche gagnante). Verdict : trois quadrants sur quatre couverts en apparence, mais le pire quadrant du rentier est à découvert total. Correction minimale type : 10 % d'or et 10 % d'obligations indexées prélevés sur les deux poches ([[or-en-retrait]], [[obligations-indexees]]) : le rendement central espéré baisse de ~0,2 point, et la colonne broad-sample de pofo (celle qui contient les stagflations) rend typiquement plusieurs points de ruine. C'est l'échange régime contre espérance, rendu visible et chiffrable : la grille a servi d'audit, le simulateur d'arbitre.
:::

## Ce que les régimes changent à la conduite du plan

Au-delà du portefeuille et du modèle, trois conséquences de pilotage.

**Les traversées se comptent en années, budgétez-les ainsi.** Un buffer de liquidités dimensionné pour « un krach » (18 mois) est dimensionné pour le mauvais objet : les épisodes hostiles durent 2 à 7 ans, retour au sommet réel compris. C'est l'ordre de grandeur qui doit calibrer le buffer et les règles de flexibilité ([[cash-buffer]], [[recharger-ou-pas]], [[flexibilite-realite]]) : tenir un régime, pas amortir une secousse.

**Le régime d'entrée en retraite mérite un regard, pas une obsession.** Partir en fin de prospérité euphorique (valorisations chères, [[valorisations-et-cape]]) ou au creux d'un bust purgé ne présente pas le même risque de séquence : le §00 de la page FIRE existe pour ce constat. Mais on ne **choisit** pas son régime de départ ; on choisit ses marges ([[les-trois-phases]]).

**Ne confondez pas régime et bruit.** Le corollaire psychologique : une année de −15 % n'est pas un changement de régime, et le pilotage écrit ([[quand-s-inquieter]]) doit résister à la tentation de « voir » un 1973 dans chaque correction. Les régimes se reconnaissent à leur **persistance** macro (inflation installée, récession déclarée), pas aux gros titres d'un trimestre. Le plan qui réagit à chaque pseudo-régime fait pire que le plan rigide ; c'est tout l'intérêt d'avoir des seuils quantitatifs décidés à froid.

## L'essentiel à retenir

- Les marchés ont des saisons persistantes de plusieurs années ; leurs signatures statistiques (ours longs, grappes de volatilité, corrélations qui basculent) sont exactement ce que les modèles i.i.d. ratent.
- La grille croissance × inflation classe les saisons en quatre quadrants ; le 60/40 n'en couvre que deux : toute la ligne inflationniste (surchauffe, stagflation) frappe actions et obligations ensemble : 1966-1981 hier, 2022 en rappel.
- Le pire quadrant du rentier est la stagflation (rendements réels négatifs des deux actifs classiques + retraits gonflés + durée) : un portefeuille de retrait s'arme **explicitement** contre elle (or, linkers, actifs réels, tendance).
- La préparation structurelle bat la prédiction : la grille sert à auditer la composition (un gagnant par quadrant), le simulateur à tester la robustesse (stress = persistance, broad-sample = régimes réels, décennie perdue = crash-test).
- Pilotage : budgétez les traversées en années (2-7 ans), regardez le régime d'entrée sans en faire une obsession, et ne prenez pas une correction pour un changement d'ère : la persistance macro, pas les gros titres.

---

## Pour aller plus loin

- Hamilton, « A New Approach to the Economic Analysis of Nonstationary Time Series and the Business Cycle » (1989) : l'article fondateur des régimes de Markov.
- Harry Browne, *Fail-Safe Investing* (1999) et Ray Dalio (Bridgewater), « The All Weather Story » : les deux formulations classiques de la grille en portefeuilles ([[portefeuilles-tous-temps]]).
- Ilmanen, *Investing Amid Low Expected Returns* (2022), chapitres régimes et inflation : l'état de l'art praticien.
- Dans pofo : les colonnes stress/broad-sample/lost-decade ([[la-machine-pofo]]) ; et pour la version tactique quantitative du Permanent Portfolio (Darcet), le document de conception dédié dans le dépôt.
- La suite naturelle : [[portefeuilles-tous-temps]] et [[actifs-defensifs]], où la grille devient allocation.
