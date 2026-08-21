# Les échelles d'obligations (et l'échelle de linkers)

Il existe deux façons de détenir des obligations. Tout ce livre, jusqu'ici, a surtout parlé de la première, le **fonds** à duration constante ([[obligations-en-retrait]]). C'est la poche permanente qui amortit et se rééquilibre. La seconde est l'**échelle**. On détient une série d'obligations, chacune jusqu'à son échéance, un « barreau » par année de dépense à financer, 2027, 2028, 2029...

L'échelle n'est pas une variante technique du fonds. C'est un changement de **paradigme**. On ne gère plus un portefeuille contre un marché. On **adosse** des passifs datés à des flux contractuels. La retraite se traite alors en actuaire plutôt qu'en investisseur ([[rentes-et-annuites]], la même école safety-first, en version obligataire et réversible). Bien employée, l'échelle est l'outil le plus puissant du livre pour trois travaux précis, le pont vers la pension, le plancher des premières années et les grosses dépenses datées. Mal employée, c'est une collection de lignes illiquides qui rejoue en nominal la fausse sécurité que l'inflation dévore.

Cet article donne le principe et la propriété qui le distingue (pourquoi « détenu à terme » change réellement la nature du risque **ici**, alors que c'était une illusion pour la poche permanente). Il déroule ensuite les trois cas d'usage avec leur construction, la version indexée (l'échelle de linkers, l'objet le plus proche d'une « retraite garantie » qui existe), la pratique française avec ses contournements, et les pièges.

::: cle Quand « tenir à échéance » cesse d'être une illusion
Pour une poche **permanente**, tenir à échéance ne protège de rien. La hausse des taux coûte pareil, en manque à gagner au lieu d'une moins-value ([[obligations-en-retrait]]). Mais pour un **passif daté**, tout change. Si l'obligation qui échoit en 2031 finance exactement les dépenses de 2031, les variations de prix intermédiaires sont **sans objet**. Le flux à l'échéance est contractuel, et il tombe le jour où on en a besoin. L'échelle ne supprime pas le risque de taux. Elle l'**annule** par appariement (immunisation). Le risque n'existe que s'il y a un décalage entre l'horizon de l'actif et celui du besoin. C'est la seule structure du livre où le mot « garanti » a un sens littéral.
:::

::: science D'où vient l'échelle : trois générations d'actuaires
L'idée n'est pas née sur un forum FIRE. Elle vient de l'assurance-vie et des fonds de pension, qui affrontent le même problème depuis toujours, des engagements datés à honorer avec des actifs qui bougent. L'actuaire britannique Frank Redington lui donne son nom en 1952, dans « Review of the Principles of Life-Office Valuations » (*Journal of the Institute of Actuaries*) : un fonds est **immunisé** quand sa valeur ne bouge plus, au premier ordre, lors d'un déplacement général des taux. Sa recette est déjà la nôtre, égaliser le terme moyen des actifs et celui des engagements. Lawrence Fisher et Roman Weil la mettent à l'épreuve des données en 1971, dans « Coping with the Risk of Interest-Rate Fluctuations » (*Journal of Business*) : une stratégie d'appariement élimine presque entièrement le risque de taux d'un portefeuille obligataire de bonne qualité détenu jusqu'à un horizon donné. Martin Leibowitz en fait enfin un métier de gérant, dans les deux volets du *Financial Analysts Journal* de 1986, « The Dedicated Bond Portfolio in Pension Funds ». Ils séparent nettement l'appariement des flux (le cash matching, exactement l'échelle de cet article) de l'immunisation par la duration, qui gère un horizon sans coller à chaque échéance.

L'adossement n'est donc ni une astuce de forum ni une trouvaille récente. C'est le métier des caisses de retraite, réduit à la taille d'un ménage. Et cette filiation dit aussi ce qu'une échelle ne fera jamais, puisque aucune caisse ne finance ainsi ses engagements les plus lointains ni ses hausses de prestations.
:::

## Les trois travaux de l'échelle

**Travail 1 : le pont vers la pension.** Le cas d'usage FIRE par excellence, déjà rencontré ([[vpw]], [[horizon-et-esperance-de-vie]]). Entre le départ et la liquidation des pensions s'étend une phase à découvert de 10-20 ans où le portefeuille finance tout. La fraction **plancher** de ces années est un passif daté, connu, non négociable, le candidat parfait à l'adossement. La construction est simple, un barreau par année (le montant du plancher non couvert), de l'année 1 à l'année de liquidation. Le reste du patrimoine, déchargé du plancher, porte le confort et le long terme avec une liberté retrouvée ([[allocation-actions-obligations]], la couverture du plancher libère vers le haut du plateau).

**Travail 2 : le plancher des premières années (la fenêtre fragile).** C'est une version plus courte du même geste. On adosse 5-8 ans de plancher au moment du départ, précisément les années où le risque de séquence est maximal ([[sequence-des-rendements]]). C'est le deuxième bucket de la stratégie des buckets, en version contractuelle et sans ambiguïté ([[strategie-buckets]]). Et contrairement au matelas cash ([[cash-buffer]]), les barreaux à 3-8 ans sont **rémunérés** au taux du marché. L'échelle courte est un buffer qui ne paie presque pas de coût d'opportunité.

**Travail 3 : les dépenses datées.** Les études des enfants (2031-2036), le solde d'un crédit, des travaux programmés. Tout passif daté et chiffrable s'adosse à un barreau, sort du portefeuille risqué, et cesse de polluer le dimensionnement du plan ([[combien-il-vous-faut]]).

Ce que l'échelle ne fait **pas**. Elle ne remplace pas la poche obligataire permanente du portefeuille (l'amortisseur rééquilibré, c'est le travail du fonds). Elle ne couvre pas non plus le **très** long terme ouvert. On n'échelonne pas 40 ans, car les barreaux lointains coûtent cher face à une inflation incertaine, et la longévité reste ouverte. Au-delà de l'échelle, les bons outils sont le portefeuille et, en fin de vie, la rente ([[rentes-et-annuites]]).

## Nominal ou indexé : la question qui décide de tout

Un barreau **nominal** de 40 000 € en 2038 financera 40 000 € de 2038, soit ~30 000 € d'aujourd'hui à 2,5 % d'inflation. Sur les barreaux au-delà de 5-7 ans, la « garantie » nominale est une garantie de **pouvoir d'achat décroissant**, la fausse sécurité classique ([[inflation-et-taux-de-retrait]]). Trois réponses, par ordre de propreté.

- **L'échelle de linkers.** Chaque barreau est indexé sur les prix. Le flux de 2038 sera 40 000 € **réels**, contractuellement. C'est l'objet « retraite garantie » déjà rencontré ([[obligations-indexees]], l'échelle TIPS à 4,5 %, sa cousine euro à ~3,9 %). C'est la solution canonique dès que l'échelle dépasse 5-7 ans. La fenêtre de taux réels dans laquelle une telle échelle bat la règle des 4 %, et l'écart que fait l'horizon retenu, se lisent sur la figure de [[obligations-indexees]].
- **L'échelle nominale gonflée.** Des barreaux nominaux croissants (40 000 × 1,025^n). Elle couvre l'inflation **anticipée** mais reste nue contre les surprises. Acceptable pour les barreaux courts, elle devient de plus en plus fragile ensuite.
- **L'échelle courte roulée.** Nominale sur 3-5 ans seulement, reconstruite chaque année par le haut. L'inflation courte est peu incertaine. Le risque long reste dans le portefeuille, traité par ses briques ([[actifs-defensifs]]).

La doctrine qui en sort est simple. Courte, la nominale est acceptable. Longue, c'est indexé ou rien. Elle n'est pas neuve. Zvi Bodie et Michael Clowes en avaient fait le cœur de *Worry-Free Investing* (2003), un livre entier consacré à bâtir le financement de la retraite sur des obligations indexées plutôt que sur des actions.

## La pratique française : les contournements du guichet absent

Le particulier américain construit son échelle TIPS en ligne en une heure. Le Français, lui, compose avec la boîte à outils ci-dessous, dans son état de 2026. L'OAT nominale en direct y est praticable, l'OAT€i beaucoup moins ([[obligations-indexees]]), et la ligne vide dit le reste.

| Véhicule | Horizon utile | Indexé | Frais | Accès |
|---|---|---|---|---|
| Fonds euros | Années 1-2 | Non | Pris avant le taux servi | Toute assurance-vie |
| Fonds à échéance État | Jusqu'à 2030 | Non | 0,1 à 0,2 % | CTO, tout courtier |
| Fonds à échéance IG | Jusqu'à 2036 | Non | 0,1 à 0,2 % | CTO, tout courtier |
| OAT en direct | Toutes maturités | Non | Courtage et spread OTC | Quelques courtiers |
| ETF linkers courts roulés | 5 ans et plus, approché | Oui | ~0,1 % | CTO, tout courtier |
| OAT€i en direct | Toutes maturités | Oui | Courtage et spread large | Très difficile |
| Fonds à échéance indexé euro |  | Oui |  | N'existe pas |

**Les fonds à échéance, la brique qui a tout changé.** Ce sont des ETF qui détiennent un panier d'obligations échéant **toutes** la même année, puis se liquident. C'est le barreau prêt-à-l'emploi, et c'est lui qui a rendu l'échelle praticable en Europe. Deux contrôles complètent la ligne du tableau, la qualité du panier (préférer État ou IG large) et l'année exacte de liquidation. Faute de version indexée en euro, on approche le résultat avec des ETF linkers courts roulés et des barreaux nominaux gonflés. C'est le chaînon manquant de l'échelle française, et il faut suivre l'offre.

**Le fonds euros en barreau court.** Pour les années 1-2, le fonds euros fait un barreau parfait (garanti, liquide, rémunéré). L'échelle française type commence souvent par lui ([[cash-buffer]]).

**Dans un simulateur.** L'échelle se modélise par équivalence. Le pont de pension adossé se représente en retirant du capital simulé le coût de l'échelle, puis en entrant le plancher couvert comme revenu ([[utiliser-la-page-fire]], les curseurs side income/pension font l'affaire pour la structure). La ruine restante est alors celle du **confort** seul. C'est tout l'intérêt, et l'exemple de [[obligations-indexees]] l'a chiffré.

::: attention Les pièges de construction
**Quatre pièges récurrents.** Le yield-chasing, d'abord, troquer l'État contre du corporate à 1 point de plus. Sur un instrument dont **tout** l'intérêt est la certitude du flux, réintroduire du risque de défaut est un contresens ([[obligations-en-retrait]], le high yield, jamais). La granularité excessive ensuite, quinze lignes de 12 000 € avec spreads OTC à chaque fois. Les fonds à échéance règlent ça. L'échelle-prison, aussi, tout le patrimoine échelonné « pour la sécurité ». L'échelle n'a ni croissance, ni flexibilité à la hausse, ni legs. Elle adosse le **plancher**, jamais le confort ([[rentes-et-annuites]], même règle que la rente). Et le barreau oublié, enfin. Une échelle vit, avec l'inflation réalisée et les dépenses qui bougent. Elle se révise à la revue annuelle ([[revue-annuelle]]), pas seulement à sa construction.
:::

::: exemple Le pont de Claire et Idris, construit
Reprenons le couple de [[choisir-sa-strategie]], plancher 45 000 €, pensions dans 13 ans qui le couvriront à ~53 %. Le passif à adosser, c'est 13 années × 45 000 € de plancher, moins ce que le portefeuille de confort peut servir en toutes circonstances. Ils décident d'adosser 100 % du plancher des années 1-6 (la fenêtre fragile) et 60 % des années 7-13. La construction se lit poste par poste. Années 1-2, le fonds euros (92 k€). Années 3-8, des fonds à échéance État/IG 2028-2033, montants gonflés à 2,5 % (~220 k€). Années 9-13, des ETF linkers courts roulés provisionnés (~120 k€). Coût total, ~430 k€ sur 1,7 M€. Le solde (1,27 M€) porte le confort en guardrails puis VPW, à un taux de retrait effectif de ~1,8 %. La ruine du confort devient anecdotique, celle du plancher contractuellement nulle jusqu'aux pensions. Le prix payé, c'est l'espérance des 430 k€, chiffré, accepté, écrit.
:::

::: figure echelle-passif
Le passif de Claire et Idris, année par année. Le besoin annuel ne bouge pas en euros d'aujourd'hui (45 k€ de plancher, 58 k€ de confort) ; ce qui change, c'est le payeur. Les barreaux couvrent 100 % du plancher les six premières années, 60 % ensuite, et les pensions prennent le relais à l'année 14. Le portefeuille de confort ne porte plus que le complément, 13 k€ par an d'abord, 31 k€ ensuite, soit 22,7 k€ en moyenne sur les treize ans : les ~1,8 % des 1,27 M€ restants.
:::

## L'essentiel à retenir

- L'échelle, c'est un barreau par année de passif, détenu à terme. L'appariement actif-passif **annule** le risque de taux, là où « tenir à échéance » n'était qu'une illusion comptable pour une poche permanente. C'est le seul « garanti » littéral du livre.
- Ses trois travaux sont le pont vers la pension, le plancher de la fenêtre fragile (un buffer rémunéré) et les dépenses datées. Jamais la poche permanente (le fonds), ni le très long terme ouvert (le portefeuille, puis la rente).
- Nominal court, acceptable. Long, **indexé** ou rien. Un barreau nominal à 12 ans garantit un pouvoir d'achat décroissant. L'échelle de linkers est la solution canonique, encore imparfaitement accessible en euro (fonds à échéance nominaux et linkers courts roulés en attendant).
- Pratique française, la boîte à outils tient dans un tableau, et son trou est l'absence de fonds à échéance indexé euro. Les contrôles portent sur la qualité État/IG, les frais et l'année exacte. Et l'échelle se révise chaque année.
- Les pièges, ce sont le yield-chasing (contresens absolu), la granularité OTC, l'échelle-prison qui adosse le confort et le barreau oublié. L'échelle sert le plancher, le portefeuille sert la vie.

---

## Pour aller plus loin

- Allan Roth et les outils d'échelle TIPS ([tipsladder.com](https://www.tipsladder.com)) : la version aboutie américaine, le modèle à transposer.
- Les gammes de fonds à échéance UCITS (iBonds et équivalents) : les fiches produits, pour la construction concrète.
- Wade Pfau, *Safety-First Retirement Planning* (2019) : l'adossement des passifs comme doctrine ([[rentes-et-annuites]]).
- Zvi Bodie & Michael Clowes, *Worry-Free Investing* (2003) : la thèse du financement indexé, en version grand public. Les trois papiers fondateurs cités plus haut sont dans [[bibliotheque]].
- Dans ce livre : [[obligations-indexees]] (le barreau réel et le résultat de l'échelle garantie), [[obligations-en-retrait]] (fonds contre titres, le vrai débat), [[strategie-buckets]] (le deuxième bucket rendu contractuel), [[cash-buffer]] (les barreaux zéro et un).
