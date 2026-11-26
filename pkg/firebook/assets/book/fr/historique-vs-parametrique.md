# Fenêtres historiques, bootstrap, paramétrique : trois familles de modèles

Votre plan aurait-il résisté à l'inflation des années 1970 ? Un rejeu historique peut vous montrer le résultat avec les rendements de l'époque. Pour savoir ce qui se passerait si plusieurs mauvaises périodes se succédaient, il faut pouvoir changer leur ordre. Le choix du modèle dépend de la question que vous posez ([[monte-carlo-forces-faiblesses]]).

Nous comparerons trois façons de construire une trajectoire de rendements. Gardez les mêmes dépenses, la même pension et le même horizon pour pouvoir interpréter les écarts.

::: cle Trois familles, trois questions
Le rejeu parcourt les dates dans leur ordre historique. Le bootstrap assemble des périodes prélevées dans les données. Un modèle paramétrique tire des rendements dans une loi dont on choisit les caractéristiques. Des hypothèses différentes peuvent conduire les trois modèles à se tromper dans le même sens.
:::

## Famille 1 : les fenêtres historiques (rejeu, cohortes)

**La mécanique.** C'est la méthode fondatrice de Bengen ([[etude-trinity]]). On prend la série réelle des rendements de votre portefeuille (ou d'un indice). Puis on déroule le plan à partir de chaque date de départ possible : la fenêtre janvier 1975 → décembre 2019, puis février 1975 → janvier 2020, et ainsi de suite. Chaque fenêtre est un « millésime », une cohorte. Le taux d'échec est la fraction des millésimes ruinés.

**Choisir les départs.** Avec 45 ans de données mensuelles et un horizon de 45 ans, il y a une seule fenêtre complète. Un mois d'historique supplémentaire permet de tester un deuxième départ. Les deux ménages ont alors traversé presque exactement les mêmes marchés. Déplacer le départ mois par mois, comme le fait ERN ([[serie-ern]]), permet d'étudier les effets du calendrier. Avec seulement 20 ans de données, un plan sur 45 ans n'a aucun résultat de rejeu complet.

**Ce qu'on apprend.** En rejouant un départ en 1966, on conserve la succession des rendements boursiers et de l'inflation observés. Les prix de marché appartiennent à l'histoire. Les ventes de titres, les dépenses et le solde du ménage restent le résultat de vos hypothèses.

**Ce qui manque.** Cent ans contiennent à peine trois périodes de 30 ans sans chevauchement. Un même krach réapparaît dans de nombreuses fenêtres mensuelles ; il n'apporte chaque fois qu'une nouvelle date de départ. Le taux d'échec décrit donc ces départs sur ce marché. Une série limitée à une période prospère peut donner trop confiance ; une période défavorable peut donner l'impression inverse.

## Famille 2 : le bootstrap (rééchantillonnage par blocs)

**La mécanique.** Choisissez un mois dans l'historique et copiez ensemble les rendements de tous les actifs pour ce mois. Continuez avec les mois suivants, puis repartez d'une autre date tirée au sort. Un même bloc peut revenir plusieurs fois. À l'intérieur d'un bloc on conserve par exemple la chute des actions puis leur reprise, ainsi que le comportement des obligations pendant ces mêmes mois. À la jonction de deux blocs, la continuité historique s'arrête.

Dans le bootstrap stationnaire de Politis et Romano, la longueur des blocs varie. Avec une moyenne de 24 mois, il y a chaque mois une chance sur 24 de repartir d'une date aléatoire. Une crise de trois ans peut donc être reprise en entier, coupée en deux ou absente d'une trajectoire.

::: figure bootstrap-blocs
Le dernier tirage reprend les mois 5 et 6 pour la deuxième fois. À chaque passage, les actions et les obligations gardent leurs rendements respectifs. Chaque bloc dure ici deux mois.
:::

**Un même moteur, deux questions.** Tout dépend du panel que l'on remélange. Appliqué à l'historique de vos propres fonds, sous vos poids courants, le bootstrap fabrique des milliers de variantes de ce que vos lignes ont réellement vécu. Appliqué à un panel long et multi-pays, il change de nature. C'est l'esprit de la méthode d'Anarkulova-Cederburg (38 pays dans leur étude) : rééchantillonner le siècle des pays développés, ici sur le panel académique Jorda-Schularick-Taylor (16 pays, 1870-2020), en tirant chaque bloc dans un même pays et en le faisant long (dix ans en moyenne), pour que les grands désastres (1929, les années 1970, le Japon) survivent en bonne partie dans les trajectoires, sur un portefeuille 60/40 domestique conforme à la littérature ([[anarkulova-cederburg]]). Même famille mathématique, deux questions différentes : « mes fonds, remélangés » contre « le siècle développé, remélangé ».

**Ce qu'on apprend.** En changeant l'ordre des blocs, on peut observer des pertes qui se suivent alors qu'elles étaient séparées de plusieurs décennies. C'est une façon d'éprouver un plan face à d'autres enchaînements de rendements déjà connus.

**Ce qui manque.** Simuler dix mille retraites à partir de vingt années de marchés laisse toujours vingt années d'observations. Des blocs courts brisent plus souvent les tendances longues. Des blocs longs les conservent plus souvent mais offrent moins de combinaisons. Si la série d'origine ne contient pas de forte inflation persistante, multiplier les tirages ne vous apprend pas comment vos fonds auraient réagi dans ce régime ([[valorisations-et-cape]]).

## Famille 3 : le paramétrique (Student-t, et les régimes)

**La mécanique.** On abandonne les données brutes. On décrit le monde par quelques paramètres, et on tire dedans. La version la plus simple est l'i.i.d. gaussien (moyenne, volatilité, tirages annuels indépendants), celle de la plupart des simulateurs commerciaux. Elle a deux défauts corrigibles : des queues trop fines et aucune mémoire. Un paramétrique sérieux corrige le premier en tirant dans une Student-t à trois paramètres : μ (la moyenne), σ (la volatilité) et df (l'épaisseur des queues). À df 5 et à volatilité égale, l'année à −30 % réel est environ trois fois plus probable qu'en loi normale ([[queues-epaisses]]). Les outils les plus soignés ajustent les trois paramètres sur les données de vos fonds, et les mélangent vers le prior broad-sample (actions mondiales, pas vos fonds) quand l'horizon dépasse l'historique ([[rendre-monte-carlo-pertinent]]).

Le second défaut, l'absence de mémoire, donne naissance à la sous-famille des modèles à régimes, paramétriques mais séquencés. Une chaîne de Markov alterne des états « normal » et « baissier », avec des probabilités de transition qui rendent ces marchés baissiers persistants : y entrer est rare, y rester est probable, et les mauvaises années arrivent en grappes de trois ans environ. C'est le principe d'un stress de séquence. Bien construit, il garde la même moyenne de long terme que le modèle central : le stress mesure le risque d'ordre, pas un pessimisme caché sur le niveau. Sa variante extrême est la décennie perdue : un marché baissier de type Japon 1990, long et profond, délibérément non compensé (la moyenne est tirée vers le bas). C'est un scénario de queue à survivre, pas une espérance.

**Ce qu'on peut tester.** Supposons que la pire perte mensuelle du portefeuille observée soit de 12 %. Un bootstrap mensuel peut enchaîner cette perte deux fois. Une loi de rendement peut aussi produire une baisse de 18 % en un seul mois, si cette valeur appartient à sa distribution. Pour savoir ce qui fragilise le plan, changez une seule hypothèse à la fois. Vous pouvez abaisser le rendement attendu ([[rendements-attendus]], [[valorisations-et-cape]]) ou conserver la même moyenne et regrouper les mauvaises années ([[sequence-des-rendements]]).

**Les pièges.** Ils sont le miroir des forces. Tout repose sur trois nombres que personne ne connaît, et la sensibilité aux entrées de [[monte-carlo-forces-faiblesses]] est ici maximale. La structure choisie (i.i.d., ou Markov à deux états) reste une caricature du réel : pas de retour des valorisations vers leur moyenne, pas de corrélation stochastique entre actifs (le portefeuille est agrégé avant tirage), et une inflation implicite (tout est en réel). Le paramétrique est un instrument de laboratoire : parfait pour les expériences contrôlées, à ne jamais confondre avec le monde.

## Les trois familles chez les simulateurs courants

Ce repérage change la façon d'utiliser n'importe quel outil, car un verdict ne se lit qu'en sachant quelle famille l'a produit. L'inventaire qui suit décrit les modèles mis en œuvre par les simulateurs les plus utilisés, tels que leur documentation les présente au moment où ce livre s'écrit (les outils évoluent, vérifiez la leur). C'est un inventaire, pas un classement. Un outil mono-famille excellent rend plus de services qu'un fourre-tout, et une famille absente signale un choix de conception, pas un défaut. Une transparence est due au lecteur : pofo accompagne ce livre et son auteur n'est pas neutre à son sujet, raison de plus pour croiser ses verdicts avec un outil indépendant.

| Outil | Familles | À savoir |
|---|---|---|
| cFIREsim | 1 (rejeu US depuis 1871) | Gratuit. Données Shiller actions/obligations/or, plans de dépense paramétrables. La référence de la simplicité. |
| ERN SWR Toolbox | 1 (rejeu mensuel US depuis 1871, tous les mois de départ) | Gratuit. Un classeur à copier, auditable formule par formule, flux annexes et règles CAPE de la série compris ([[serie-ern]]). Exige de mettre les mains dans le tableur. |
| FI Calc | 1 (rejeu US depuis 1871) | Gratuit. Le plus riche en règles de retrait prêtes à l'emploi (une douzaine, guardrails et VPW compris), pédagogie soignée. |
| pofo | 1, 2 et 3 (fenêtres mensuelles de vos fonds ; bootstrap de vos fonds et broad-sample du siècle des 16 pays ; Student-t calibrée, stress de séquence, décennie perdue) | L'outil développé avec ce livre. Le plus jeune de la liste, fiscalité réduite à un taux mixte global, pas de modèle à volatilité persistante. |
| Portfolio Visualizer | 2 et 3 (tirage d'années historiques, sans blocs ; normale, Student-t, GARCH, espérances prévisionnelles) | Le GARCH capture les grappes de volatilité, chose rare. Le tirage historique est année par année, donc sans mémoire. Une bonne partie des fonctions est devenue payante. |
| Rich, Broke or Dead | 1 (cycles US depuis 1871) | Gratuit. Croise chaque cycle avec les tables de mortalité : la visualisation « riche, ruiné ou mort » la plus parlante du genre. |
| TPAW Planner | 1 et 2 (séquences historiques, et tirages historiques recentrés sur les espérances du jour, 1/CAPE et taux réels) | Gratuit. L'implémentation de référence de l'amortissement (ABW), et le seul de la liste à ancrer d'office ses espérances sur les valorisations. |

Avant de comparer deux outils, relevez le pays étudié, la période et l'allocation. Rejouer un portefeuille américain depuis 1871 et rééchantillonner vos propres fonds depuis leur création change plusieurs hypothèses à la fois. Vérifiez aussi si l'outil utilise directement les rendements historiques ou les recentre sur une espérance choisie. Ce détail explique pourquoi modifier l'espérance peut laisser un rejeu inchangé ou déplacer tout un ensemble de simulations.

## La grille de lecture combinée

Reste la vraie question : que faire quand les familles sont en désaccord, ce qui est le cas normal ? Voici la grille, cas de désaccord par cas de désaccord.

**L'historique est plus favorable que le modèle central.** Comparez d'abord le rendement moyen de la période à celui retenu dans le modèle. L'écart peut venir d'une période particulièrement favorable, d'une hypothèse prudente ou d'une différence de risque. Pour distinguer ces explications, modifiez une seule hypothèse puis relancez le même plan.

**Central acceptable, stress de séquence nettement pire.** Votre plan est exposé à l'ordre des rendements : retrait initial élevé, peu de flexibilité, pas de revenus précoces. Ce n'est pas un problème de niveau d'espérance, mais de structure. Les parades sont celles de la table anti-séquence ([[sequence-des-rendements]]) : flexibilité écrite, coussin de sécurité (buffer), glidepath, revenus des premières années.

**Seul le modèle international donne beaucoup d'échecs.** Il applique vos dépenses et votre pension à un portefeuille domestique 60/40 dans une succession de périodes et de pays. Examinez les périodes difficiles. Pour tester ce que changerait la diversification de vos propres actifs, revenez ensuite à un modèle qui tient compte de leurs poids ([[anarkulova-cederburg]], [[portefeuilles-tous-temps]]).

**Le plan traverse même la décennie perdue.** Vérifiez le niveau de dépenses qu'il a fallu accepter. Si le plancher de vie est préservé, testez une question concrète, comme avancer la date de départ d'un an. Une seule épreuve de résistance ne suffit pas à établir que vous épargnez trop ([[une-annee-de-plus]], [[depenses-en-retraite]]).

Gardez avec chaque résultat la question qu'il permet de trancher. Un taux d'échec calculé sur vos fonds et un taux calculé sur un 60/40 domestique peuvent tous deux être instructifs sans mesurer le même risque.

::: exemple Un désaccord instructif
Vous disposez de 15 ans de cours pour vos fonds et prévoyez une retraite de 40 ans. La case « fenêtres historiques » doit rester vide. Un bootstrap peut assembler 40 ans à partir de ces 15 ans, mais chaque bloc provient de la même courte période. Pour tester la sensibilité du plan, gardez votre portefeuille et baissez le rendement supposé dans le modèle paramétrique. Regardez alors à quel moment les dépenses essentielles cessent d'être couvertes. Comparez ce délai à celui du scénario central pour mesurer l'effet de votre hypothèse.
:::

## L'essentiel à retenir

- Trois familles : rejouer (fenêtres/cohortes), remélanger (bootstrap par blocs, dont le broad-sample sur le siècle des 16 pays), générer (paramétrique Student-t, et ses variantes à régimes pour la mémoire).
- Chacune répond à une autre question, et leurs pièges sont complémentaires : échantillon minuscule et plafond du réalisé (1), mémoire tronquée au bloc et ingrédients du passé (2), sensibilité aux entrées et structure caricaturale (3).
- Les rendements d'un rejeu pur sont imposés par les dates choisies. Les modèles qui recentrent des données historiques sur une espérance choisie combinent rééchantillonnage et hypothèses prospectives.
- Un écart entre deux modèles indique une hypothèse à examiner. Des écarts de rendement, d'allocation ou de période peuvent produire le même symptôme.
- Pour chaque scénario difficile, regardez combien de temps le ménage a pu financer ses dépenses essentielles et ce qui lui a permis de le faire.

---

## Pour aller plus loin

- Politis & Romano, [« The Stationary Bootstrap » (1994)](https://users.ssc.wisc.edu/~bhansen/718/Politis%20Romano.pdf), sur la construction des blocs de longueur aléatoire.
- Anarkulova, Cederburg, O'Doherty & Sias (2023) : le bootstrap par blocs appliqué au siècle développé, la référence du modèle broad-sample ([[anarkulova-cederburg]]).
- Early Retirement Now, volet 8 : la méthode du rejeu mensuel systématique ([[serie-ern]]).
- Dans ce livre : [[queues-epaisses]] (le choix Student-t en détail), [[rendre-monte-carlo-pertinent]] (mélange et ancres du modèle central), [[la-machine-pofo]] (l'implémentation exacte de chacun de ces modèles dans pofo).
