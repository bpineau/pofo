# Le matelas de liquidités : taille, coût, vrai rôle

C'est la protection la plus intuitive de toute la décumulation : garder deux ou trois ans de dépenses en liquidités, et vivre dessus quand les marchés plongent, pour ne jamais vendre d'actions au pire moment. L'intuition est si forte que le « cash buffer » figure dans tous les plans amateurs et tous les discours de conseillers. Et c'est précisément pourquoi il faut regarder les chiffres en face, car la recherche, elle, est étonnamment tiède : simulé mécaniquement, le matelas améliore **peu** la probabilité de ruine (le coût d'opportunité du cash rend à peu près ce que la protection de séquence rapporte), et les simulations l'affichent sans fard : la courbe d'arbitrage du buffer est le plus souvent plate.

Alors, gadget ? Non. Mais son vrai rôle n'est pas là où on le croit.

Cet article démonte le dossier complet : pourquoi l'intuition est juste et l'effet net quand même modeste, ce qui distingue un buffer **utile** d'un buffer décoratif (les règles de consommation et de recharge font tout), la taille et le placement corrects pour un Français (le fonds euros change un peu la donne), la comparaison avec le glidepath, et la vraie valeur du matelas : comportementale et de gouvernance : qui justifie amplement sa place à condition de le payer au juste prix.

::: cle Le paradoxe du buffer, en une phrase
Chaque euro mis au matelas est un euro retiré du moteur : la protection contre les **mauvaises** séquences se paie d'un manque à gagner dans **toutes** les autres. Et comme les bonnes séquences sont majoritaires, le net quantitatif est proche de zéro (±0,5 point de ruine selon les règles). Le buffer ne se justifie donc pas par les statistiques. Il se justifie par ce qu'il fait de vous : un rentier qui dort, n'improvise pas dans les krachs, et exécute sa règle : ce qui, en pratique, vaut plus que bien des points de simulation ([[psychologie-du-retrait]]).
:::

## L'intuition, et pourquoi elle ne suffit pas

**L'intuition est mathématiquement fondée.** Le mécanisme mortel du retrait est de vendre des parts dépréciées pendant les creux : chaque vente au creux transforme une perte temporaire en perte définitive ([[sequence-des-rendements]]). Un matelas qui absorbe les retraits pendant la traversée supprime **exactement** ces ventes : sur le papier, c'est l'arme anti-séquence parfaite.

**Mais le financement du matelas a un coût symétrique.** Les 2-3 ans de dépenses du matelas (7-10 % d'un plan à 3,5 %) rapportent ~0-1 % réel au lieu des 3-5 % du portefeuille : un manque à gagner de ~0,25-0,4 % par an **sur tout le plan**, payé chaque année, y compris les trente années où aucun krach ne mord. La simulation fait les comptes des deux colonnes : ERN (volet 12, « Cash Cushion Concerns ») trouve le net légèrement **négatif** pour les buffers statiques naïfs ; les études plus favorables (avec des règles intelligentes) trouvent un net légèrement positif ; en balayant le buffer de 0 à 10 ans, un simulateur montre selon les plans une courbe plate ou un optimum mou vers 2-4 ans. Le verdict quantitatif honnête : **le buffer bien géré est à peu près gratuit ; il n'est presque jamais transformateur.**

::: figure buffer-flat
La probabilité de ruine en fonction de la taille du matelas (allure typique). La courbe est presque plate : sur toute la plage, moins d'un point de ruine sépare le meilleur du pire réglage. Il existe un optimum mou vers 2-3 ans, mais au-delà la courbe **remonte** : trop de buffer appauvrit le moteur plus qu'il ne protège. Quantitativement, le matelas est donc à peu près neutre ; sa vraie valeur est comportementale.
:::

**Pourquoi si peu, alors que l'intuition est si forte ?** Deux raisons profondes. D'abord, l'arithmétique des durées : les traversées du désert durent 2 à 7 ans, retour au sommet **réel** compris ([[regimes-de-marche]], et l'histogramme « years underwater » de la §07) : un matelas de 2-3 ans ne couvre que la première moitié des vraies traversées. Il déplace les ventes au creux, il ne les supprime pas toutes. Ensuite, le rééquilibrage fait **déjà** la moitié du travail : un portefeuille 70/30 rééquilibré prélève naturellement sur les obligations pendant les krachs d'actions ([[retrait-fixe-bengen]], [[obligations-en-retrait]]) : le buffer explicite ajoute une couche à un mécanisme largement présent. D'où le rendement décroissant.

## Ce qui sépare un buffer utile d'un buffer décoratif

Le même matelas peut être un instrument ou un totem ; quatre choix font la différence.

**1. La règle de consommation, écrite.** Le buffer décoratif n'a pas de règle. On « sent » quand l'utiliser : c'est-à-dire qu'on improvise sous stress, trop tôt (à la première correction de 10 %, gaspillant le matelas avant le vrai creux) ou trop tard. Le buffer utile a un déclencheur quantitatif : par exemple « les retraits basculent sur le matelas quand le portefeuille est en drawdown réel de plus de 15-20 % ; ils y restent jusqu'à retour sous ce seuil ». Simple, mécanique, exécutable par le conjoint ([[quand-s-inquieter]], [[couple-et-famille]]).

**2. La règle de recharge : le sujet est assez riche pour son propre article** ([[recharger-ou-pas]]) : retenons ici le principe : un matelas consommé se reconstitue aux **sommets** (jamais au creux, recharger en vendant des actions déprimées annule tout le bénéfice), et un matelas qui ne se recharge jamais devient une simple tranche de dépenses prépayées : légitime aussi, mais c'est un autre objet (le « pont » des premières années, cousin de l'échelle, [[echelle-obligataire]]).

**3. La taille : 18-36 mois, pas plus.** En dessous de 12 mois, l'effet est cosmétique ; au-delà de 3-4 ans, le coût d'opportunité croît linéairement pendant que la protection marginale s'effondre (les traversées de 5+ ans sont rares et trop longues pour être pré-financées en cash, c'est le travail de la flexibilité et des actifs de régime, [[flexibilite-realite]], [[actifs-defensifs]]). La courbe §07 de **votre** plan arbitre ce chiffre en deux clics ; les optima mous sortent presque toujours dans la fourchette 2-3 ans.

**4. Le placement : rémunéré, liquide, insensible.** Le matelas français idéal n'est pas le compte courant. C'est le **fonds euros** (capital garanti, rendement obligataire lissé, disponible en jours, le véhicule qui réduit réellement le coût d'opportunité du buffer de moitié, [[obligations-en-retrait]], [[enveloppes-francaises]]) et/ou le monétaire €STR en CTO, plus les livrets réglementés pour la première tranche. Ce que le matelas n'est **jamais** : investi en obligations longues (il doit être insensible aux taux, c'est sa définition) ni, évidemment, en actions.

::: attention La convention, à connaître pour lire la §07
Dans pofo, le buffer est **prélevé** sur le capital de départ, pas ajouté par-dessus : « 3 ans de buffer » sur un plan à 1,5 M€ signifie ~150 k€ au matelas et 1,35 M€ au moteur : la richesse affichée est toujours la somme des deux. C'est la convention honnête (celle qui fait payer au buffer son coût d'opportunité), et c'est pourquoi la courbe d'arbitrage de la §07 peut **monter** à droite : trop de buffer appauvrit le moteur plus qu'il ne protège. Les deux autres réglages : le rendement réel du matelas (« Buffer real return », calez-le sur votre fonds euros net, ~0,5-1 % réel) et l'année d'arrêt de la recharge (« refill stops in year », utile pour modéliser un buffer de première décennie seulement, cohérent avec la concentration du risque, [[sequence-des-rendements]]).
:::

## La vraie valeur : ce que les simulations ne comptent pas

Si le net statistique est ~nul, pourquoi ce livre recommande-t-il quand même un matelas ? Parce que trois services réels échappent aux simulations, et qu'ils sont précisément les plus rares.

**L'anti-panique.** La simulation applique la règle sans émotion ; l'humain, non. Le désastre comportemental type (vendre **tout** en mars 2009 ou mars 2020 « pour sauver ce qui reste ») coûte 20-40 % de richesse finale : des ordres de grandeur au-dessus de tous les débats de ce chapitre. Or le mécanisme déclencheur est identifié : la peur de devoir vendre pour **vivre**. Le rentier qui **sait** que ses 30 prochains mois de courses sont garantis en fonds euros regarde le même krach avec un autre système nerveux : le buffer est un anxiolytique structurel, et son efficacité anti-capitulation est la mieux documentée de ses vertus ([[psychologie-du-retrait]], [[marche-baissier-en-retraite]]).

**La permission de dépenser.** Symétrique et sous-estimé : les retraités sous-consomment massivement par peur du lendemain ([[depenses-en-retraite]]) ; le matelas visible désinhibe la consommation planifiée : le voyage se réserve, parce que « l'argent est déjà là ».

**La gouvernance du ménage.** Le matelas est l'instrument le plus simple à léguer opérationnellement : « en cas de gros temps, on vit sur l'assurance-vie X » est une instruction qu'un conjoint non gestionnaire exécute sans aide ([[couple-et-famille]]) : aucune règle de retrait sophistiquée n'a cette propriété.

La conclusion d'assemblage s'écrit alors simplement : le buffer se **dimensionne** au minimum efficace (18-36 mois, placé en fonds euros, règles écrites) pour capter les trois services comportementaux au plus bas coût statistique. Et l'on refuse les matelas pharaoniques (5-10 ans) qui achètent les mêmes services au triple du prix. Face au glidepath, qui couvre le même risque par l'allocation ([[glidepaths]]) : les deux se recouvrent, la combinaison modérée (tente douce + 2 ans de matelas) domine les versions extrêmes de l'un ou l'autre.

::: exemple Un matelas écrit, et sa décennie
Plan : 1,5 M€, 52 000 €/an, corridor Vanguard. Matelas : 130 k€ (30 mois) en fonds euros, règles écrites : « consommation : drawdown réel > 18 % ; recharge : aux nouveaux sommets réels, par les retraits excédentaires, jusqu'à 30 mois ; jamais de recharge en drawdown ». Simulation §07 : ruine 4,1 % sans matelas, 3,8 % avec : l'honnête presque-rien attendu. La décennie vécue, elle : année 3, krach de 26 % : les retraits basculent 19 mois sur le fonds euros, zéro vente d'actions sous −18 %. C'est tout le point : zéro insomnie et zéro tentation de « tout sécuriser » au creux ; années 5-6 : sommets, recharge par excédents ; année 8 : correction de 12 % : rien, le seuil n'est pas franchi, le matelas ne bouge pas. Le 0,3 point de ruine était le prix d'entrée ; la décennie exécutée sans panique était le produit.
:::

## L'essentiel à retenir

- L'intuition (ne jamais vendre au creux) est juste ; l'arithmétique est têtue : le coût d'opportunité du cash rend à peu près ce que la protection rapporte : net quantitatif ±0,5 point : la courbe §07 l'affiche sans fard, buffer **prélevé** sur le capital.
- Un buffer utile a quatre attributs : déclencheur de consommation écrit (drawdown > 15-20 %), règle de recharge aux sommets ([[recharger-ou-pas]]), taille 18-36 mois, placement en fonds euros/monétaire (jamais de duration).
- Sa vraie valeur est hors simulation : l'anti-panique (le désastre comportemental coûte 10 fois tous les débats de taille), la permission de dépenser, la gouvernance du ménage : trois services qui s'achètent au minimum efficace, pas au maximum rassurant.
- Les traversées durent 2-7 ans : le matelas couvre la première moitié, le reste appartient à la flexibilité, au rééquilibrage et aux actifs de régime : un buffer de 5-10 ans achète trois fois trop cher.
- Buffer et glidepath couvrent le même risque : combinaison modérée plutôt que maximum de l'un ; et le buffer se règle finement (rendement réel, année d'arrêt de recharge) pour simuler **votre** version, pas la caricature.

---

## Pour aller plus loin

- Early Retirement Now, volet 12 (« Cash Cushion Concerns ») : le contre-dossier quantitatif fondateur ([[serie-ern]]).
- Michael Kitces, « Managing Sequence Of Return Risk With Bucket Strategies Vs A Total Return Rebalancing Approach » : la réconciliation buffer/rééquilibrage.
- Dans pofo : la §07 (l'arbitrage buffer et l'histogramme des années sous l'eau, le dimensionnement sur **vos** traversées) ([[utiliser-la-page-fire]]).
- Dans ce livre : [[recharger-ou-pas]] (les règles de flux du matelas), [[strategie-buckets]] (la version en étages, et sa critique), [[glidepaths]] (l'alternative par l'allocation).
