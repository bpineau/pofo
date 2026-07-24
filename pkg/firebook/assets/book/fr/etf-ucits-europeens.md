# Construire en UCITS : le portefeuille de retrait de l'investisseur européen

Toute la littérature du retrait est écrite en dollars. Elle s'appuie sur des véhicules américains (VTI, BND, les TIPS au guichet) qu'un résident français ne **peut** pas acheter. Depuis 2018, la réglementation PRIIPs ferme les ETF américains aux particuliers européens. La bonne nouvelle est que l'écosystème européen, les fonds UCITS, offre aujourd'hui tout ce qu'il faut pour construire chaque brique de ce livre. Il ajoute même des avantages propres : les parts capitalisantes et la retenue à la source optimisée des fonds irlandais. Une seule condition, connaître les règles du jeu, qui ne sont pas celles des blogs américains.

Cet article est le guide d'assemblage. Il couvre ce que UCITS garantit vraiment, et ce que le label ne garantit pas. Il détaille les quatre choix techniques qui comptent : capitalisant ou distribuant, réplication physique ou synthétique, domicile, part couverte ou non. Il pose la vraie mesure des coûts, l'écart de suivi (tracking difference) plutôt que le TER. Il dresse la table des briques UCITS correspondant à chaque article de cette partie. Il précise la répartition entre enveloppes françaises, c'est-à-dire quoi loger où, la mécanique fiscale détaillée restant dans [[enveloppes-francaises]]. Il termine par la pratique de la décumulation en ETF, car vendre proprement chaque mois est un savoir-faire en soi.

::: cle Le renversement capitalisant
L'habitude française cherche du « revenu » : des fonds distribuants, des dividendes. En décumulation, il faut faire exactement l'inverse. On choisit des parts **capitalisantes**, où les dividendes sont réinvestis dans le fonds, sans taxation immédiate en CTO. On fabrique ensuite le revenu par des **ventes** programmées, du montant exact de votre règle de retrait. Le gain est triple. D'abord le contrôle : **votre** règle décide du montant, pas les politiques de dividendes ([[panorama-strategies-retrait]]). Ensuite la fiscalité : une vente n'est taxée que sur sa **part** de plus-value, un dividende l'est à 100 %. À flux égal, la vente est toujours moins taxée ([[flat-tax-et-imposition]]). Enfin, on échappe au piège du rendement (yield), car chasser du dividende déforme le portefeuille ([[actifs-defensifs]]). Le « revenu » d'un rentier moderne est un ordre de vente mensuel, pas un coupon.
:::

## UCITS : ce que le label garantit, et les quatre choix techniques

**Le label.** UCITS (OPCVM coordonnés) est le cadre réglementaire européen des fonds grand public. Il garantit d'abord la **ségrégation** des actifs : le fonds est une entité séparée, et la faillite de l'émetteur ou du dépositaire ne touche pas vos parts. Il impose aussi une diversification minimale, la liquidité quotidienne et l'encadrement des risques annexes (prêt de titres collatéralisé, exposition swap limitée à 10 %). La protection structurelle est réelle : le risque d'un ETF UCITS large est le risque de son **marché**, pas celui de son fournisseur. En revanche, le label ne garantit ni la qualité de l'indice suivi, ni les frais, ni la pertinence du produit. Il existe des UCITS absurdes.

**Choix 1 : capitalisant (Acc) ou distribuant (Dist).** L'encart l'a déjà réglé : Acc partout où c'est possible, Dist seulement si une contrainte d'enveloppe l'impose.

**Choix 2 : réplication physique ou synthétique.** La réplication physique, où le fonds détient les titres, est le choix par défaut. La réplication synthétique a un usage français décisif : le fonds détient un panier de substitution et échange sa performance contre celle de l'indice via un swap. Elle permet à un ETF composé d'actions européennes de restituer la performance du **monde entier** tout en restant éligible au PEA. C'est le mécanisme des « ETF Monde PEA », la clé de la diversification internationale dans l'enveloppe la plus douce ([[diversification-internationale]], [[enveloppes-francaises]]). Le risque de swap est réel, mais borné : 10 % maximum, collatéralisé, remis à zéro régulièrement. C'est un compromis raisonnable pour l'avantage fiscal.

**Choix 3 : le domicile, Irlande d'abord pour les actions américaines.** Voici un détail méconnu qui vaut des points de base durables. Les dividendes des actions **américaines** subissent une retenue à la source avant d'arriver au fonds. Elle est de 15 % pour un fonds **irlandais**, grâce au traité fiscal Irlande-USA, contre 30 % pour un fonds luxembourgeois. Sur environ 1,3 % de rendement de dividende, l'écart vaut à peu près 0,2 %/an, chaque année, gratuitement. À TER égal, préférez donc les fonds domiciliés en Irlande (suffixe IE dans l'ISIN) pour l'exposition américaine et mondiale.

**Choix 4 : couvert ou non couvert.** La question est déjà tranchée ([[diversification-internationale]], [[obligations-en-retrait]]) : actions **non** couvertes, obligations en euro ou couvertes. Concrètement, ignorez les parts « EUR Hedged » des ETF actions, qui coûtent chaque année pour supprimer un amortisseur. Réservez la couverture aux éventuelles briques obligataires hors zone euro.

## Les coûts : la chaîne complète, et la vraie mesure

Le TER affiché (0,07 à 0,45 % pour les briques de ce livre) n'est que le premier maillon. La chaîne complète en compte cinq. Le TER, d'abord. Puis l'écart de suivi interne, lié au prêt de titres et aux optimisations, parfois **négatif** quand un ETF bat son indice. Puis le spread à l'achat et à la vente, soit 0,02 à 0,15 % sur les gros fonds. Puis le courtage, de 0 à quelques euros chez les courtiers modernes. Enfin, la couche d'**enveloppe** éventuelle : 0,5 à 1 %/an de frais de gestion pour une assurance-vie sur ses UC, le maillon qui domine tout quand il existe ([[enveloppes-francaises]]). La mesure qui agrège les deux premiers maillons est l'écart de suivi (tracking difference) : la performance du fonds moins celle de son indice, sur 3 à 5 ans, lisible sur les sites spécialisés. C'est elle qu'on compare entre deux fonds candidats, pas le TER. Un portefeuille complet bien construit revient à **0,15 à 0,30 %/an tout compris**, contre 1,5 à 2,5 % pour le même portefeuille en unités de compte chargées ou en fonds actifs de réseau. Sur 40 ans de décumulation, l'écart atteint de l'ordre de 0,5 à 1 point de taux de retrait. Le choix de la tuyauterie vaut donc autant que bien des débats de stratégie ([[rendements-attendus]]).

## La table des briques

Chaque article de cette partie a sa brique UCITS. Voici la correspondance, en types de produits plutôt qu'en noms commerciaux, car l'offre bouge alors que les critères restent : gros encours, écart de suivi propre, domicile IE pour l'américain, part Acc.

| Brique | Produit UCITS | Enveloppe naturelle | À vérifier |
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

La règle d'assemblage tient en une phrase : **six à huit lignes suffisent** pour tout le programme de cette partie. Au-delà, chaque ligne supplémentaire doit déloger une ligne existante ou remplir un rôle que la table ne couvre pas. La collectionnite d'ETF, quinze lignes qui se recouvrent, est l'erreur cosmétique la plus répandue. Elle n'ajoute aucune diversification, car les indices se contiennent les uns les autres, et elle complique chaque rééquilibrage et chaque succession ([[couple-et-famille]]).

::: astuce La répartition entre enveloppes, en une règle
La mécanique fiscale complète est dans [[enveloppes-francaises]]. La règle d'implantation qui en sort tient en trois lignes. Le PEA prend les actions, via le Monde synthétique, jusqu'à son plafond ; c'est l'enveloppe la plus douce en sortie. L'**assurance-vie** prend le fonds euros (buffer, tranche courte) et, si les UC sont propres et peu chargées, un complément actions ou obligations ; ses atouts sont la succession et l'abattement annuel en rachat. Le CTO prend **tout** ce que les deux autres ne peuvent pas loger proprement : or, linkers, trend, SCV, duration longue. Sa fiscalité simple, le PFU sur la part de gain, est moins pénalisante qu'on ne le croit en décumulation, précisément parce que les ventes ne sont taxées que sur leur fraction de plus-value.
:::

## La décumulation en pratique : vendre proprement

Le moment venu, le portefeuille UCITS se consomme. Quelques savoir-faire évitent les frottements.

**La vente programmée.** Elle sert le « salaire » mensuel ou trimestriel ([[retrait-fixe-bengen]]). On passe un ordre sur la ligne désignée par le prélèvement-rééquilibrage, celle en surpoids, aux heures où le sous-jacent est ouvert. Pour un ETF monde, ce sera l'après-midi, quand les États-Unis cotent : les spreads y sont les plus serrés. Jamais d'ordre au marché à l'ouverture ou dans les minutes de panique. Un ordre limité au milieu de fourchette coûte un clic de plus et des points de base de moins.

**La fiscalité des ventes partielles.** En CTO, la plus-value imposable d'une vente se calcule au **prix moyen pondéré** (PMP) d'acquisition de la ligne. La conséquence pratique est nette. Les ventes des premières années de retraite portent sur des lignes dont le PMP est proche du cours : elles réalisent donc peu de gain taxable. La friction fiscale réelle démarre **bas** et monte avec les années. C'est exactement ce que simule le modèle de taxe, avec une charge effective qui dérive vers le haut ([[utiliser-la-page-fire]], [[flat-tax-et-imposition]]).

**L'hygiène de long terme.** Une revue par an suffit ([[revue-annuelle]]). On vérifie l'écart de suivi (tracking difference) des lignes, car un fonds qui dérive se remplace. On surveille les encours : un fonds qui rétrécit sous environ 200 M€ finit fusionné ou fermé, un événement gérable mais taxable en CTO. On garde enfin une veille passive de l'offre. Les frais baissent tendanciellement, mais on ne change **pas** de fonds pour 3 points de base si la vente réalise dix ans de plus-values. Le coût fiscal de la migration se calcule avant, et il l'emporte souvent.

**Les erreurs de construction à éviter**, en rafale : les ETF de **niche** (thématiques, sectoriels, pays uniques, des paris et non des briques) ; les parts distribuantes « pour le revenu », visées par l'encart d'ouverture ; les parts couvertes sur les actions ; les produits à levier ou inverses, que leur érosion quotidienne (drag) disqualifie ([[rendements-arithmetiques-geometriques]]) ; le fournisseur unique par principe, alors que deux émetteurs pour six lignes offrent une diversification opérationnelle gratuite.

::: exemple Le portefeuille complet, en sept lignes
La cible de Karim et Léa ([[diversification-internationale]]) : 1,6 M€, 65/35, phase à découvert de 14 ans, tout ce livre assemblé. Le PEA (300 k€) loge l'ETF Monde synthétique Acc → 19 %. Le CTO (1 050 k€) loge le reste : All-World physique IE Acc → 33 % ; SCV US et Europe → 8 % ; État euro 5-8 ans Acc → 14 % ; linkers euro 1-5 ans → 8 % ; ETC or physique → 6 % ; fonds trend UCITS à frais fixes → 6 %. L'AV (250 k€) loge le fonds euros → 16 % (buffer de 30 mois et tranche courte). Au total, sept lignes, trois enveloppes, environ 0,22 %/an tout compris. Chaque ligne remplit un rôle nommé dans la table des défenses, et le rééquilibrage aux bandes de ±5 s'exécute par les ventes de retrait. C'est **tout** : la sophistication de ce portefeuille est dans sa conception, pas dans son nombre de lignes. Et il tient sur la page écrite du plan ([[construire-son-plan]]).
:::

## L'essentiel à retenir

- L'écosystème UCITS couvre toutes les briques du livre, avec deux avantages propres : les parts capitalisantes (le revenu du rentier vient de ventes programmées, moins taxées et mieux contrôlées que tout dividende) et les fonds irlandais (retenue américaine à 15 %, environ 0,2 %/an gagnés sur les actions américaines).
- Quatre choix techniques : Acc partout, synthétique pour le Monde en PEA, domicile IE pour l'américain, jamais de couverture sur les actions. La vraie mesure des coûts est l'écart de suivi (tracking difference), pas le TER, pour une cible de 0,15 à 0,30 %/an sur le portefeuille complet.
- Six à huit lignes suffisent (monde, PEA-monde, tilt optionnel, cœur obligataire, linkers, or, trend, fonds euros), chaque ligne remplissant un rôle de la table des défenses. La collectionnite n'ajoute rien et complique tout.
- Répartition : PEA pour les actions, AV pour le fonds euros (et des UC propres), CTO pour tout le reste. Sa fiscalité au prix moyen pondéré rend les premières années de retraits très peu taxées.
- Vendre proprement : ordres limités aux heures américaines, prélèvement-rééquilibrage, revue annuelle des fonds (écart de suivi, encours), et jamais de migration de fonds sans calcul du coût fiscal.

---

## Pour aller plus loin

- justETF et TrackingDifferences.com : la sélection et le contrôle des fonds (tracking difference historique par part).
- Bogleheads wiki, sections « eu investing » et « Non-US investor » : le corpus de référence de l'investisseur UCITS.
- L'AMF ([amf-france.org](https://www.amf-france.org)) : les documents d'information clé (DIC) et la pédagogie officielle des ETF.
- Dans ce livre : [[enveloppes-francaises]] (la mécanique fiscale des trois enveloppes), [[flat-tax-et-imposition]] (le calcul des ventes), et chaque brique dans son article de la partie V.
