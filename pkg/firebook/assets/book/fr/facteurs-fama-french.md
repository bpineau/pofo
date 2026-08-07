# Les facteurs (Fama-French, value, momentum) en phase de retrait

Depuis trente ans, la recherche académique découpe le rendement des actions en « facteurs » : des caractéristiques systématiques (la décote de valorisation, la petite taille, la qualité, la tendance) qui ont historiquement payé une prime au-dessus du marché. C'est l'un des corpus les plus solides de la finance **et** l'un des plus surexploités par le marketing : entre les deux, le rentier doit trancher une question précise, qui n'est pas celle de l'accumulateur : les inclinaisons factorielles améliorent-elles un plan de **décumulation**, où ce qui compte est le pire chemin et non la moyenne ([[sequence-des-rendements]]) ?

La réponse honnête est nuancée et cet article la déroule : ce que les facteurs sont vraiment (et pourquoi ils existeraient encore), les chiffres avec leur décote post-publication, le dossier spécifique du retrait (où le small-cap value a des états de service remarquables : c'est le choix du Golden Butterfly et l'objet du Part 62 d'ERN : et où la value a une affinité de régime précieuse), les coûts réels (la décennie 2010 a été un purgatoire value : la tracking-error psychologique est le prix d'entrée), la mise en œuvre UCITS depuis la France (praticable mais étroite), et la dose : car ici plus qu'ailleurs, c'est un raffinement **optionnel**, à ne prendre que si l'on comprend ce qu'on achète.

::: cle L'idée en une phrase
Un facteur est une source de rendement **systématique** distincte du marché : détenir « les actions décotées » ou « les petites rentables » plutôt que « les actions », c'est diversifier les **raisons** pour lesquelles le portefeuille peut gagner : et pour un rentier, diversifier les raisons de gagner, c'est diversifier les époques où l'on souffre : le vrai argument factoriel en décumulation n'est pas « plus de rendement » mais « pas les mêmes décennies perdues » ([[regimes-de-marche]]).
:::

## De quoi parle-t-on : le socle académique, en bref

L'histoire tient en quatre dates. 1964 : le CAPM (Sharpe) : un seul facteur, le marché ; tout le reste serait du bruit. 1992-1993 : Fama et French montrent que deux caractéristiques prédisent les rendements en coupe : la **taille** (les petites capitalisations battent les grandes) et la **value** (les décotées sur leurs fondamentaux battent les chères) : le modèle à trois facteurs. 1997 : Carhart ajoute le **momentum** (les gagnantes récentes persistent : le cousin en coupe du trend, [[managed-futures]]). 2015 : Fama-French passent à cinq avec la **rentabilité** (profitability : les entreprises rentables battent les autres à prix égal) et l'**investissement** (les sobres battent les dépensières) : la « qualité » des praticiens. Au-delà commence le « zoo des facteurs » (des centaines publiés, la plupart morts en réplication) : les cinq ci-dessus, plus le momentum, sont le noyau répliqué sur longues périodes et hors échantillon (autres pays, autres classes d'actifs).

Pourquoi ces primes existeraient-elles **encore**, une fois publiées ? Deux familles d'explications, probablement complémentaires : le **risque** (les value et les petites sont plus fragiles en récession : la prime paie un vrai risque : auquel cas elle survit par définition mais se **paie** en mauvais moments) ; le **comportement** (extrapolation des gloires récentes, négligence des ennuyeuses : auquel cas la prime survit tant que l'arbitrage est limité). La décote post-publication est mesurée (McLean-Pontiff : −30 à −50 % de la prime après publication) : les chiffres historiques bruts (value : 3-5 %/an en long-short académique ; en pratique **indicielle** long-only, un tilt value ou small-value capte plutôt 1-2 %/an d'excès espéré) se lisent avec cette décote ([[rendements-attendus]]).

## Le dossier du rentier : trois pièces

**Pièce 1 : Bengen et les petites capitalisations.** Dès ses travaux de suivi, Bengen a montré qu'ajouter des small caps US au portefeuille de référence remontait le SAFEMAX d'environ 0,3-0,5 point ([[etude-trinity]]) : premier indice que la diversification **interne** de la poche actions compte pour les pires chemins.

**Pièce 2 : ERN Part 62 et le small-cap value.** Jeske a consacré un volet à la question exacte : le SCV (petites décotées : le croisement le plus étudié) en décumulation. Constat sur les données longues : les millésimes où le marché large meurt (1966, 2000 : des départs à valorisations extrêmes du marché **large**) sont souvent des millésimes où le SCV, parti de valorisations normales, traverse : le retraité 2000 en SCV a vécu une décennie honorable pendant la décennie perdue du S&P. Mécanisme : les grandes catastrophes de valorisation sont **concentrées** (la bulle 2000 était une bulle de grandes valeurs de croissance ; 1966, une bulle des Nifty Fifty) : le tilt écarte du point de concentration ([[valorisations-et-cape]]). Contre-exemple assumé : 1929-1932 et 2008, où les petites décotées ont souffert **plus** (le facteur paie son risque récessif). Bilan ERN : un tilt SCV partiel améliore la plupart des pires millésimes historiques, sans garantie de structure.

**Pièce 3 : l'affinité de régime de la value.** La pièce moderne : la value est historiquement le style qui résiste le **mieux** aux régimes inflationnistes (flux proches, actifs tangibles, dette qui s'érode) quand la croissance-longue-duration souffre (ses flux lointains sont actualisés plus durement : 2022 : value mondiale ~0 %, croissance −25 %). Pour un portefeuille dont l'ennemi n° 1 est le régime inflationniste ([[regimes-de-marche]]), un tilt value est une demi-brique défensive logée **dans** la poche actions : gratuite en espérance (voire payée), là où l'or et les linkers coûtent leur portage ([[actifs-defensifs]]).

::: attention Le prix d'entrée : la décennie 2010, et toutes les suivantes
Toute prime factorielle se paie en **tracking error** : des années, parfois une décennie, à faire moins bien que l'indice que tout le monde détient. La décennie 2010-2020 est l'exemple à encadrer : la value mondiale a sous-performé la croissance de ~5 %/**an** pendant dix ans (le pire épisode de son histoire, pire que 1999) : les capitulations ont été massives : juste avant le retournement de 2020-2024. Symétrie parfaite avec l'hiver des CTA ([[managed-futures]]) et le purgatoire de l'or 1980-2000 ([[or-en-retrait]]) : **toute** diversification réelle a ses années de honte : c'est même à cela qu'on la reconnaît (ce qui ne diverge jamais de l'indice n'apporte rien à l'indice). La conséquence pratique est la même partout : un tilt ne se prend que documenté, écrit, dimensionné pour être tenu dix ans sans preuve : sinon, l'indice large et la paix ([[psychologie-du-retrait]]).
:::

## La mise en œuvre UCITS : praticable, mais étroite

Le paysage européen s'est amélioré mais reste mince comparé aux États-Unis.

**Ce qui existe et fonctionne** : le **small-cap value** via les ETF « Small Value Weighted » (SPDR MSCI USA et Europe Small Cap Value Weighted : les deux briques standard de la communauté, frais ~0,30 %, méthodologie propre) ; la **value** large via les « Enhanced Value » (MSCI World/USA Value factor) : attention aux indices « value » dilués des gammes bas de gamme ; le **momentum** via les « World Momentum » (utile mais rotation et frais internes) ; la **qualité** via « Quality » ou, indirectement, les « World Quality/profitability screened ». Le **multifacteur** en un fonds existe (World multifactor) : commodité contre opacité de construction : à n'acheter qu'après lecture de la méthodologie.

**Les points de contrôle** : la **capacité** du fonds (le SCV est étroit : préférer les grands véhicules), le **taux de rotation** (le momentum surtout : des frais internes invisibles), et l'exposition **réelle** au facteur (beaucoup d'ETF « value » ont un tilt homéopathique : regarder le P/B et P/E relatifs à l'indice parent). Logement : CTO pour la plupart (certains éligibles PEA via les versions européennes : à vérifier ligne par ligne, [[enveloppes-francaises]], [[etf-ucits-europeens]]).

**Dans pofo** : les briques factorielles se testent comme les autres : l'A/B type : remplacer 20-30 % de la poche actions monde par du SCV US/Europe, et lire **non** le central (l'espérance du tilt, décotée, y est presque invisible) mais les millésimes §02 (2000 : le terrain de gloire du SCV ; 1929 : son terrain de peine) et la dispersion des trajectoires : le tilt **élargit** légèrement le cône (plus de risque spécifique) en déplaçant les pires chemins : c'est exactement le compromis à juger ([[lire-un-fan-chart]]).

## La dose, et le verdict honnête

**La dose praticienne** : un tilt de 10 à 30 % de la **poche actions** (soit 7-20 % du portefeuille total), en une ou deux briques simples (SCV US + Europe est le standard communautaire ; ou un tilt value large pour la version douce). En dessous de 10 %, cosmétique ; au-delà de 30 %, le portefeuille devient un pari de style dont la tracking error dominera la vie du plan.

**Le verdict, sans esquive.** Les facteurs sont du **second ordre** en décumulation : loin derrière les dépenses auditées, le taux initial, la règle de retrait et l'allocation ([[choisir-sa-strategie]]) ; du même ordre que le choix fin dans le plateau d'allocation ([[allocation-actions-obligations]]). L'argument **pour**, réel : diversifier les régimes de souffrance de la poche actions (la pièce 2000/SCV et l'affinité inflation/value sont de vraies améliorations des pires chemins historiques). Les arguments **contre**, réels aussi : primes décotées et incertaines, tracking error décennale, complexité et frais, offre UCITS étroite. La position par défaut de ce livre : **le tilt est LÉGITIME et OPTIONNEL** : un plan sans facteurs, bien construit par ailleurs, ne manque de rien d'essentiel ; un plan avec 20 % de SCV documenté et tenu gagne un peu de robustesse de régime : et un plan avec des facteurs abandonnés en 2019 a payé la prime des autres. Choisissez votre camp **une** fois, par écrit ([[construire-son-plan]]).

::: exemple Un tilt jugé sur pièces
Plan : 1,5 M€, 51 000 €/an, 70 % actions monde / 30 % défensif. Variante : la poche actions devient 50 % monde + 10 % SCV US + 10 % SCV Europe. Verdicts pofo type : central 3,9 → 3,8 % (l'espérance du tilt, décotée, à peine visible) ; millésime 2000 (§02) : nettement adouci (le SCV a traversé la décennie perdue du large) ; millésime 1929 : légèrement aggravé ; stress et broad-sample : −0,3 à −0,5 point (l'affinité inflation de la value joue dans les blocs 1970) ; dispersion du cône : +5 % de largeur. Décision type d'un couple quantitatif : adopté, avec la clause : « tilt jugé sur dix ans glissants contre sa **thèse** (les régimes), jamais contre le Nasdaq de l'année ». Décision type d'un couple qui veut la paix : refusé, et c'est tout aussi défendable : la robustesse gagnée vaut ~0,3 point, la certitude de tenir l'indice large en vaut autant.
:::

## L'essentiel à retenir

- Le noyau répliqué : marché, taille, value, momentum, rentabilité/investissement : des primes réelles mais **décotées** post-publication (~1-2 %/an en tilt indiciel long-only), expliquées par un mélange de risque (payé aux mauvais moments) et de comportement.
- Le dossier du rentier tient en trois pièces : les small caps de Bengen (+0,3-0,5 point de SAFEMAX), le SCV d'ERN Part 62 (traverse les millésimes de bulle du marché large : 2000 : mais souffre plus en récession pure : 1929), l'affinité value-inflation (une demi-brique défensive gratuite dans la poche actions).
- Le prix d'entrée est la tracking error décennale (value 2010-2020 : −5 %/an pendant dix ans) : toute diversification réelle a ses années de honte : tilt écrit et tenu, ou pas de tilt.
- Mise en œuvre UCITS : SCV via les Small Value Weighted (US/Europe), value large via Enhanced Value, contrôles de capacité/rotation/exposition réelle ; 10-30 % de la poche actions, en CTO surtout.
- Verdict : légitime et **optionnel** : du second ordre derrière dépenses, taux, règle et allocation : à juger dans pofo sur les millésimes et les queues (§02, stress), jamais sur le central.

---

## Pour aller plus loin

- Fama & French, « The Cross-Section of Expected Stock Returns » (1992) et « A Five-Factor Asset Pricing Model » (2015) : les sources.
- Larry Swedroe, *Your Complete Guide to Factor-Based Investing* : la synthèse praticienne honnête (critères : persistant, pervasif, robuste, investissable, intuitif).
- Early Retirement Now, Part 62 (small-cap value en retrait) ([[serie-ern]]) ; McLean & Pontiff (2016) sur la décote post-publication.
- Dans ce livre : [[allocation-actions-obligations]] (le contenu de la poche actions), [[regimes-de-marche]] (l'affinité de régime), [[etf-ucits-europeens]] (les véhicules), [[managed-futures]] (le momentum en série temporelle, son cousin plus puissant).
