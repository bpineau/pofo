# Construire en UCITS : le portefeuille de retrait de l'investisseur européen

Toute la littérature du retrait est écrite en dollars, avec des véhicules américains (VTI, BND, les TIPS au guichet) qu'un résident français ne **peut** pas acheter : la réglementation PRIIPs ferme les ETF américains aux particuliers européens depuis 2018. La bonne nouvelle : l'écosystème européen, les fonds UCITS, offre aujourd'hui tout ce qu'il faut pour construire chaque brique de ce livre, souvent avec des avantages propres (les parts capitalisantes, la retenue à la source optimisée des fonds irlandais) : à condition de connaître les règles du jeu, qui ne sont pas celles des blogs américains.

Cet article est le guide d'assemblage : ce que UCITS garantit vraiment (et ce que ça ne garantit pas), les quatre choix techniques qui comptent (capitalisant/distribuant, réplication physique/synthétique, domicile, part couverte ou non), la vraie mesure des coûts (la tracking difference, pas le TER), la table des briques UCITS correspondant à chaque article de cette partie, la répartition entre enveloppes françaises (quoi loger où, la mécanique fiscale détaillée étant dans [[enveloppes-francaises]]), et la pratique de la décumulation en ETF, car vendre proprement chaque mois est un savoir-faire en soi.

::: cle Le renversement capitalisant
L'habitude française cherche du « revenu » : des fonds distribuants, des dividendes. En décumulation, c'est exactement l'inverse qu'il faut : des parts **capitalisantes** (les dividendes sont réinvestis dans le fonds, sans taxation immédiate en CTO), et un revenu fabriqué par des **ventes** programmées du montant exact de votre règle de retrait. Triple avantage : le contrôle (**votre** règle décide du montant, pas les politiques de dividendes, [[panorama-strategies-retrait]]), la fiscalité (une vente n'est taxée que sur sa **part** de plus-value, un dividende l'est à 100 %, à flux égal, la vente est toujours moins taxée, [[flat-tax-et-imposition]]), et l'absence du piège du yield ([[actifs-defensifs]], chasser du dividende déforme le portefeuille). Le « revenu » d'un rentier moderne est un ordre de vente mensuel, pas un coupon.
:::

## UCITS : ce que le label garantit, et les quatre choix techniques

**Le label.** UCITS (OPCVM coordonnés) est le cadre réglementaire européen des fonds grand public : il garantit la **ségrégation** des actifs (le fonds est une entité séparée, la faillite de l'émetteur ou du dépositaire ne touche pas vos parts), la diversification minimale, la liquidité quotidienne, et l'encadrement des risques annexes (prêt de titres collatéralisé, exposition swap limitée à 10 %). C'est une protection structurelle réelle : le risque d'un ETF UCITS large est le risque de son **marché**, pas celui de son fournisseur. Ce que le label ne garantit pas : la qualité de l'indice suivi, les frais, ni la pertinence du produit (il y a des UCITS absurdes).

**Choix 1 : capitalisant (Acc) ou distribuant (Dist).** Réglé dans l'encart : Acc partout où c'est possible, Dist seulement si une contrainte d'enveloppe l'impose.

**Choix 2 : réplication physique ou synthétique.** La physique (le fonds détient les titres) est le défaut naturel. La synthétique (le fonds détient un panier de substitution et échange sa performance contre celle de l'indice via un swap) a un usage français décisif : elle permet à un ETF détenant des actions européennes de livrer la performance du **monde entier** tout en restant éligible au PEA : c'est le mécanisme des « ETF Monde PEA », la clé de la diversification internationale dans l'enveloppe la plus douce ([[diversification-internationale]], [[enveloppes-francaises]]). Le risque swap est réel mais borné (10 % max, collatéralisé, remis à zéro régulièrement) : un compromis raisonnable pour l'avantage fiscal.

**Choix 3 : le domicile : Irlande d'abord pour les actions américaines.** Détail méconnu qui vaut des points de base durables : les dividendes des actions **américaines** subissent une retenue à la source avant d'arriver au fonds : 15 % pour un fonds **irlandais** (traité fiscal Irlande-USA), 30 % pour un luxembourgeois : sur ~1,3 % de rendement de dividende, l'écart vaut ~0,2 %/an, chaque année, gratuitement : à TER égal, préférez les fonds domiciliés en Irlande (suffixe IE dans l'ISIN) pour l'exposition US et monde.

**Choix 4 : couvert ou non couvert.** Déjà tranché ([[diversification-internationale]], [[obligations-en-retrait]]) : actions **non** couvertes, obligations en euro ou couvertes : concrètement, ignorer les parts « EUR Hedged » des ETF actions (un coût annuel pour supprimer un amortisseur) et n'utiliser la couverture que pour d'éventuelles briques obligataires non-euro.

## Les coûts : la chaîne complète, et la vraie mesure

Le TER affiché (0,07-0,45 % pour les briques de ce livre) n'est que le premier maillon. La chaîne complète : TER + écart de suivi interne (prêt de titres, optimisations, parfois **négatif**, certains ETF battent leur indice) + spread à l'achat/vente (0,02-0,15 % sur les gros fonds) + courtage (0 à quelques euros chez les courtiers modernes) + la couche d'**enveloppe** éventuelle (0,5-1 %/an de frais de gestion d'une assurance-vie sur ses UC, le maillon qui domine tout quand il existe, [[enveloppes-francaises]]). La mesure qui agrège les deux premiers maillons est la **tracking difference** (la performance du fonds moins celle de son indice, sur 3-5 ans, lisible sur les sites spécialisés) : c'est elle qu'on compare entre deux fonds candidats, pas le TER. Ordre de grandeur du portefeuille complet bien construit : **0,15-0,30 %/an tout compris** : contre 1,5-2,5 % pour le même portefeuille en unités de compte chargées ou en fonds actifs de réseau : sur 40 ans de décumulation, l'écart est de l'ordre de 0,5 à 1 point de taux de retrait : le choix de la tuyauterie vaut autant que bien des débats de stratégie ([[rendements-attendus]]).

## La table des briques

Chaque article de cette partie a sa brique UCITS ; voici la correspondance (types de produits plutôt que noms commerciaux, l'offre bouge, les critères restent, gros encours, tracking difference propre, domicile IE pour l'US, part Acc) :

| Brique du livre | Type de produit UCITS | Enveloppe naturelle | Points de contrôle |
|---|---|---|---|
| Actions monde ([[diversification-internationale]]) | ETF MSCI World / FTSE All-World, physique, Acc, IE | CTO (et AV en UC propre) | Tracking diff, encours > 1 Md€ |
| Actions monde en PEA | ETF Monde **synthétique** éligible PEA | PEA | Qualité du swap, encours |
| Tilt SCV / value ([[facteurs-fama-french]]) | ETF Small Cap Value Weighted US/Europe ; Enhanced Value | CTO | Exposition réelle au facteur, capacité |
| Cœur obligataire ([[obligations-en-retrait]]) | ETF obligations d'État euro 5-8 ans ou aggregate euro | CTO / AV | Duration affichée, qualité (État/IG) |
| Duration longue (assurance-déflation) | ETF État euro 15-30 ans | CTO | Dose limitée et assumée |
| Linkers ([[obligations-indexees]]) | ETF euro inflation-linked 1-5 ans (courts) | CTO | Duration **réelle** courte, indice euro |
| Tranche courte / buffer ([[cash-buffer]]) | Fonds euros (AV) ; ETF monétaire €STR | AV / CTO | Rendement net de frais d'enveloppe |
| Or ([[or-en-retrait]]) | ETC or **physique** alloué | CTO | Adossement physique, TER ≤ 0,25 % |
| Trend ([[managed-futures]]) | Fonds UCITS trend à frais fixes / réplication | CTO | Contrôle SG Trend, frais sans perf fee |

La règle d'assemblage : **six à huit lignes suffisent** pour tout le programme de cette partie. Chaque ligne supplémentaire au-delà doit déloger une ligne existante ou justifier un rôle que la table ne couvre pas : la collectionnite d'ETF (quinze lignes qui se recouvrent) est l'erreur cosmétique la plus répandue : elle n'ajoute aucune diversification (les indices se contiennent les uns les autres) et complique chaque rééquilibrage et chaque succession ([[couple-et-famille]]).

::: astuce La répartition entre enveloppes, en une règle
La mécanique fiscale complète est dans [[enveloppes-francaises]] ; la règle d'implantation qui en sort tient en trois lignes : le PEA prend les actions (via le Monde synthétique) jusqu'à son plafond : c'est l'enveloppe la plus douce en sortie ; l'**assurance-vie** prend le fonds euros (buffer, tranche courte) et, si les UC sont propres et peu chargées, un complément actions/obligations : ses atouts sont la succession et l'abattement annuel en rachat ; le CTO prend **tout** ce que les deux autres ne peuvent pas loger proprement : or, linkers, trend, SCV, duration longue : sa fiscalité simple (PFU sur la part de gain) est moins pénalisante qu'on ne le croit en décumulation, précisément parce que les ventes ne sont taxées que sur leur fraction de plus-value.
:::

## La décumulation en pratique : vendre proprement

Le moment venu, le portefeuille UCITS se consomme : quelques savoir-faire évitent les frottements.

**La vente programmée.** Le « salaire » mensuel ou trimestriel ([[retrait-fixe-bengen]]) : un ordre au marché sur la ligne désignée par le prélèvement-rééquilibrage (vendre le surpoids), aux heures où le sous-jacent est ouvert (pour un ETF monde, l'après-midi, quand les États-Unis cotent, les spreads y sont au plus serré) ; jamais d'ordre au marché à l'ouverture ou dans les minutes de panique : un ordre limité au milieu de fourchette coûte un clic de plus et des points de base de moins.

**La fiscalité des ventes partielles.** En CTO, la plus-value imposable d'une vente se calcule au **prix moyen pondéré** d'acquisition de la ligne : conséquence pratique : les ventes des premières années de retraite (sur des lignes dont le PMP est proche du cours) réalisent peu de gain taxable : la friction fiscale réelle démarre **bas** et monte avec les années : exactement ce que le modèle de taxe de pofo simule (la charge effective qui dérive vers le haut, [[utiliser-la-page-fire]], [[flat-tax-et-imposition]]).

**L'hygiène de long terme.** Une revue par an ([[revue-annuelle]]) : tracking difference des lignes (un fonds qui dérive se remplace), encours (un fonds qui rétrécit sous ~200 M€ finit fusionné ou fermé, événement gérable mais taxable en CTO), et veille passive de l'offre (les frais baissent tendanciellement, mais on ne change **pas** de fonds pour 3 points de base si la vente réalise dix ans de plus-values, le coût fiscal de la migration se calcule avant, et il gagne souvent).

**Les erreurs de construction à éviter**, en rafale : les ETF de **niche** (thématiques, sectoriels, pays uniques, des paris, pas des briques) ; les parts distribuantes « pour le revenu » (l'encart d'ouverture) ; les parts hedgées actions ; les produits à levier ou inverses (le drag quotidien les disqualifie, [[rendements-arithmetiques-geometriques]]) ; et le fournisseur unique par principe (deux émetteurs pour six lignes, une diversification opérationnelle gratuite).

::: exemple Le portefeuille complet, en sept lignes
La cible de Karim et Léa ([[diversification-internationale]]), 1,6 M€, 65/35, phase à découvert de 14 ans, tout ce livre assemblé : PEA (300 k€) : ETF Monde synthétique Acc : 19 %. CTO (1 050 k€) : All-World physique IE Acc : 33 % ; SCV US + Europe : 8 % ; État euro 5-8 ans Acc : 14 % ; linkers euro 1-5 ans : 8 % ; ETC or physique : 6 % ; fonds trend UCITS frais fixes : 6 %. AV (250 k€) : fonds euros : 16 % (buffer 30 mois + tranche courte). Sept lignes, trois enveloppes, ~0,22 %/an tout compris, chaque ligne un rôle nommé dans la table des défenses, rééquilibrage aux bandes ±5 exécuté par les ventes de retrait. C'est **tout** : la sophistication de ce portefeuille est dans sa conception, pas dans son nombre de lignes : et il tient sur la page écrite du plan ([[construire-son-plan]]).
:::

## L'essentiel à retenir

- L'écosystème UCITS couvre toutes les briques du livre, avec deux avantages propres : les parts capitalisantes (le revenu du rentier = des ventes programmées, moins taxées et mieux contrôlées que tout dividende) et les fonds irlandais (retenue US à 15 %, ~0,2 %/an gagnés sur les actions américaines).
- Quatre choix techniques : Acc partout, synthétique pour le Monde en PEA, domicile IE pour l'US, jamais de hedge sur les actions ; et la vraie mesure des coûts est la tracking difference, pas le TER : cible portefeuille complet : 0,15-0,30 %/an.
- Six à huit lignes suffisent (monde, PEA-monde, tilt optionnel, cœur obligataire, linkers, or, trend, fonds euros) : chaque ligne un rôle de la table des défenses : la collectionnite n'ajoute rien et complique tout.
- Répartition : PEA = actions ; AV = fonds euros (+UC propres) ; CTO = tout le reste : sa fiscalité au prix moyen pondéré rend les premières années de retraits très peu taxées.
- Vendre proprement : ordres limités aux heures US, prélèvement-rééquilibrage, revue annuelle des fonds (tracking, encours), et jamais de migration de fonds sans calcul du coût fiscal.

---

## Pour aller plus loin

- justETF et TrackingDifferences.com : la sélection et le contrôle des fonds (tracking difference historique par part).
- Bogleheads wiki, sections « eu investing » et « Non-US investor » : le corpus de référence de l'investisseur UCITS.
- L'AMF ([amf-france.org](https://www.amf-france.org)) : les documents d'information clé (DIC) et la pédagogie officielle des ETF.
- Dans ce livre : [[enveloppes-francaises]] (la mécanique fiscale des trois enveloppes), [[flat-tax-et-imposition]] (le calcul des ventes), et chaque brique dans son article de la partie V.
