# Les échelles d'obligations (et l'échelle de linkers)

Il existe deux façons de détenir des obligations, et tout ce livre jusqu'ici a surtout parlé de la première : le **fonds** à duration constante, la poche permanente qui amortit et se rééquilibre ([[obligations-en-retrait]]). La seconde est l'**échelle** : une série d'obligations détenues chacune jusqu'à son échéance, un « barreau » par année de dépense à financer : 2027, 2028, 2029...

L'échelle n'est pas une variante technique du fonds. C'est un changement de **paradigme** : on ne gère plus un portefeuille contre un marché, on **adosse** des passifs datés avec des flux contractuels : la retraite traitée en actuaire plutôt qu'en investisseur ([[rentes-et-annuites]], la même école safety-first, en version obligataire et réversible). Bien employée, l'échelle est l'outil le plus puissant du livre pour trois travaux précis : le pont vers la pension, le plancher des premières années, et les grosses dépenses datées ; mal employée, c'est une collection de lignes illiquides qui rejoue en nominal la fausse sécurité que l'inflation dévore.

Cet article donne le principe et sa vraie propriété (pourquoi « détenu à terme » change réellement la nature du risque **ici**, alors que c'était une illusion pour la poche permanente), les trois cas d'usage avec leur construction, la version indexée (l'échelle de linkers, l'objet le plus proche d'une « retraite garantie » qui existe), la pratique française avec ses contournements, et les pièges.

::: cle Quand « tenir à échéance » cesse d'être une illusion
Pour une poche **permanente**, tenir à échéance ne protège de rien : la hausse des taux coûte pareil, en manque à gagner au lieu de moins-value ([[obligations-en-retrait]]). Mais pour un **passif daté**, tout change : si l'obligation qui échoit en 2031 finance exactement les dépenses de 2031, les variations de prix intermédiaires sont **sans objet** : le flux à l'échéance est contractuel, et il tombe le jour où on en a besoin. L'échelle ne supprime pas le risque de taux. Elle l'**annule** par appariement (immunisation) : le risque n'existe que s'il y a mismatch entre l'horizon de l'actif et celui du besoin. C'est la seule structure du livre où le mot « garanti » a un sens littéral.
:::

## Les trois travaux de l'échelle

**Travail 1 : le pont vers la pension.** Le cas d'usage FIRE par excellence, déjà rencontré ([[vpw]], [[horizon-et-esperance-de-vie]]) : entre le départ et la liquidation des pensions s'étend une phase à découvert de 10-20 ans où le portefeuille finance tout ; la fraction **plancher** de ces années est un passif daté, connu, non négociable : le candidat parfait à l'adossement. Construction : un barreau par année (le montant du plancher non couvert), de l'année 1 à l'année de liquidation ; le reste du patrimoine, déchargé du plancher, porte le confort et le long terme avec une liberté retrouvée ([[allocation-actions-obligations]], la couverture du plancher libère vers le haut du plateau).

**Travail 2 : le plancher des premières années (la fenêtre fragile).** Version plus courte du même geste : adosser 5-8 ans de plancher au moment du départ, précisément les années où le risque de séquence est maximal ([[sequence-des-rendements]]). C'est le deuxième bucket de la stratégie des buckets, en version contractuelle et sans ambiguïté ([[strategie-buckets]]). Et contrairement au matelas cash ([[cash-buffer]]), les barreaux à 3-8 ans sont **rémunérés** au taux du marché : l'échelle courte est un buffer qui ne paie presque pas de coût d'opportunité.

**Travail 3 : les dépenses datées.** Les études des enfants (2031-2036), le solde d'un crédit, des travaux programmés : tout passif daté et chiffrable s'adosse à un barreau, sort du portefeuille risqué, et cesse de polluer le dimensionnement du plan ([[combien-il-vous-faut]]).

Ce que l'échelle ne fait **pas** : la poche obligataire permanente du portefeuille (l'amortisseur rééquilancé, c'est le travail du fonds), et le **très** long terme ouvert (on n'échelonne pas 40 ans, les barreaux lointains coûtent cher en inflation incertaine et la longévité est ouverte, au-delà de l'échelle, les bons outils sont le portefeuille et, en fin de vie, la rente, [[rentes-et-annuites]]).

## Nominal ou indexé : la question qui décide de tout

Un barreau **nominal** de 40 000 € en 2038 financera... 40 000 € de 2038 : soit ~30 000 € d'aujourd'hui à 2,5 % d'inflation : sur les barreaux au-delà de 5-7 ans, la « garantie » nominale est une garantie de **pouvoir** d'**achat décroissant** : la fausse sécurité classique ([[inflation-et-taux-de-retrait]]). Trois réponses, par ordre de propreté :

- **L'échelle de LINKERS** : chaque barreau indexé sur les prix : le flux de 2038 sera 40 000 € **réels**, contractuellement. C'est l'objet « retraite garantie » déjà rencontré ([[obligations-indexees]], l'échelle TIPS à 4,6 %, sa cousine euro à ~3,9 %) : la solution canonique dès que l'échelle dépasse 5-7 ans.
- **L'échelle nominale GONFLÉE** : des barreaux nominaux croissants (40 000 × 1,025^n) : couvre l'inflation **anticipée**, reste nue contre les surprises : acceptable pour les barreaux courts, de plus en plus fragile ensuite.
- **L'échelle courte roulée** : nominale sur 3-5 ans seulement, reconstruite chaque année par le haut : l'inflation courte est peu incertaine ; le risque long reste dans le portefeuille, traité par ses briques ([[actifs-defensifs]]).

La doctrine qui en sort : **courte = nominale acceptable ; longue = indexée ou rien.**

## La pratique française : les contournements du guichet absent

Le particulier américain construit son échelle TIPS en ligne en une heure ; le Français doit composer.

**Les obligations d'État en direct** : les OAT s'achètent au détail chez certains courtiers (marché secondaire), avec des tickets raisonnables sur les souches liquides : praticable pour une échelle nominale de qualité ; les OATi/OAT€i en direct restent difficiles d'accès au détail ([[obligations-indexees]]).

**Les fonds à échéance (« iBonds » et équivalents UCITS)** : des ETF qui détiennent un panier d'obligations échéant **toutes** la même année, puis se liquident : le barreau prêt-à-l'emploi, en versions obligations d'État ou entreprises investment grade, années 2026-2034 et au-delà : la brique qui a rendu l'échelle praticable en Europe. Points de contrôle : la qualité du panier (préférer État/IG large), les frais (~0,1-0,2 %), et l'année exacte de liquidation. Pas encore de version **indexée** euro à ce jour : l'échelle de linkers directe reste le chaînon manquant français : les approximations : ETF linkers courts roulés + barreaux nominaux gonflés, en attendant que l'offre suive.

**Le fonds euros en barreau court** : pour les années 1-2, le fonds euros fait un barreau parfait (garanti, liquide, rémunéré) : l'échelle française type commence souvent par lui ([[cash-buffer]]).

**Et dans pofo** : l'échelle se modélise par équivalence : le pont de pension adossé se représente en retirant du capital simulé le coût de l'échelle et en entrant le plancher couvert comme revenu ([[utiliser-la-page-fire]], les curseurs side income/pension font l'affaire pour la structure) : la ruine restante est alors celle du **confort** seul. C'est tout l'intérêt, et l'exemple de [[obligations-indexees]] l'a chiffré.

::: attention Les pièges de construction
**Quatre** pièges récurrents. Le yield-chasing : troquer l'État contre du corporate à 1 point de plus : sur un instrument dont **tout** l'intérêt est la certitude du flux, réintroduire du risque de défaut est un contresens ([[obligations-en-retrait]], le high yield, jamais). La granularité excessive : quinze lignes de 12 000 € avec spreads OTC à chaque fois : les fonds à échéance règlent ça. L'échelle-prison : tout le patrimoine échelonné « pour la sécurité » : l'échelle n'a ni croissance, ni flexibilité à la hausse, ni legs. Elle adosse le **plancher**, jamais le confort ([[rentes-et-annuites]], même règle que la rente). Et le barreau oublié : une échelle vit (l'inflation réalisée, les dépenses qui bougent). Elle se re-toise à la revue annuelle ([[revue-annuelle]]), pas à sa construction seulement.
:::

::: exemple Le pont de Claire et Idris, construit
Reprenons le couple de [[choisir-sa-strategie]] : plancher 45 000 €, pensions dans 13 ans qui le couvriront à ~53 %. Le passif à adosser : 13 années × 45 000 € de plancher, moins ce que le portefeuille de confort peut raisonnablement toujours servir. Ils décident d'adosser 100 % du plancher des années 1-6 (la fenêtre fragile) et 60 % des années 7-13. Construction : années 1-2 : fonds euros (92 k€) ; années 3-8 : fonds à échéance État/IG 2028-2033, montants gonflés à 2,5 % (~180 k€) ; années 9-13 : ETF linkers courts roulés provisionnés (~120 k€). Coût total : ~392 k€ sur 1,7 M€ ; le solde (1,31 M€) porte le confort en guardrails puis VPW, à un taux de retrait effectif de ~1,9 % : la ruine du confort devient anecdotique, celle du plancher contractuellement nulle jusqu'aux pensions. Le prix payé : l'espérance des 392 k€ : chiffré, accepté, écrit.
:::

## L'essentiel à retenir

- L'échelle = un barreau par année de passif, détenu à terme : l'appariement actif-passif **annule** le risque de taux (là où, pour une poche permanente, « tenir à échéance » n'était qu'une illusion comptable) : le seul « garanti » littéral du livre.
- Ses trois travaux : le pont vers la pension, le plancher de la fenêtre fragile (un buffer rémunéré), les dépenses datées : jamais la poche permanente (le fonds) ni le très long terme ouvert (le portefeuille, puis la rente).
- Nominal court = acceptable ; long = **indexé** ou rien : un barreau nominal à 12 ans garantit un pouvoir d'achat décroissant : l'échelle de linkers est la solution canonique, encore imparfaitement accessible en euro (fonds à échéance nominaux + linkers courts roulés en attendant).
- Pratique française : fonds euros pour les barreaux 1-2, fonds à échéance UCITS pour le cœur, OAT en direct possible ; contrôles : qualité État/IG, frais, année exacte ; et l'échelle se re-toise chaque année.
- Les pièges : yield-chasing (contresens absolu), granularité OTC, l'échelle-prison qui adosse le confort, le barreau oublié : l'échelle sert le plancher, le portefeuille sert la vie.

---

## Pour aller plus loin

- Allan Roth et les outils d'échelle TIPS ([tipsladder.com](https://www.tipsladder.com)) : la version aboutie américaine, le modèle à transposer.
- Les gammes de fonds à échéance UCITS (iBonds et équivalents) : les fiches produits, pour la construction concrète.
- Wade Pfau, *Safety-First Retirement Planning* : l'adossement des passifs comme doctrine ([[rentes-et-annuites]]).
- Dans ce livre : [[obligations-indexees]] (le barreau réel et le résultat de l'échelle garantie), [[obligations-en-retrait]] (fonds contre titres, le vrai débat), [[strategie-buckets]] (le deuxième bucket rendu contractuel), [[cash-buffer]] (les barreaux zéro et un).
