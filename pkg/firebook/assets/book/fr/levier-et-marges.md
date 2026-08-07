# Levier, marge et lombard en retrait (avancé)

Ce chapitre est le plus contre-intuitif de la partie protections. Il parle d'**emprunter**, dans un livre consacré à ne pas manquer d'argent. Disons d'emblée la position : pour l'immense majorité des plans de retrait, le levier n'a aucune place. Il amplifie exactement le risque, la séquence, que tout le reste du livre s'emploie à réduire.

Mais « aucun levier » mérite d'être une **conclusion**, pas un tabou. Il existe trois usages dont la logique, en décumulation, est réelle et documentée. Le premier est le crédit de liquidité : le lombard sert de pont, il évite de vendre au creux, une alternative au matelas de cash. Le deuxième est le levier d'efficacité de capital, le return stacking ([[return-stacking]]). On emprunte non pour détenir plus d'actions, mais pour ajouter des actifs défensifs sans diluer le moteur. C'est l'idée derrière les fonds 90/60. Le troisième est le levier contracyclique d'ERN : emprunter dans les creux plutôt que vendre (volets 49 et 52 de sa série).

Cet article traite les trois honnêtement. Il montre d'abord pourquoi le levier « naïf » est toxique précisément en retrait. Il passe ensuite les instruments en revue, du crédit immobilier au lombard, des ETF à levier (non) aux fonds capital-efficient (peut-être). Il finit par les règles de sécurité, non négociables, pour les rares profils concernés.

::: cle Pourquoi le levier change de signe à la retraite
En accumulation, un jeune épargnant à flux entrants peut rationnellement s'endetter pour s'exposer. C'est la thèse du lifecycle investing d'Ayres et Nalebuff : lisser l'exposition aux actions sur toute la vie. En **décumulation**, les trois mécanismes du levier jouent tous contre vous. Il amplifie la volatilité, donc le frein de volatilité (volatility drag) : un levier ×2 fait ×4 sur la variance ([[rendements-arithmetiques-geometriques]]). Il amplifie la séquence, car les retraits, les intérêts et d'éventuels appels de marge sortent tous pendant les creux ([[sequence-des-rendements]]). Et il introduit le seul risque capable de tuer un plan en une semaine : la **liquidation forcée** au pire prix. Le levier permanent sur le moteur d'un rentier n'est pas « agressif ». Il est incohérent avec la structure du problème.
:::

## La revue des instruments, du légitime au toxique

Six instruments, du plus doux au plus douteux. Un seul critère de tri décide vraiment, la possibilité d'un appel de marge, alors lisez la deuxième colonne avant les autres.

| Instrument | Appel de marge | Coût | Durée | Verdict |
|---|---|---|---|---|
| Crédit immobilier conservé | Aucun | Le taux fixe d'origine | Échéancier connu | Acceptable sous conditions |
| Avance d'assurance-vie | Aucun, quotité 60 % en unités de compte | Emprunts d'État + 1 point environ | 3 ans, renouvelable | Le pont le plus praticable |
| Lombard bancaire | Oui, quotité de 50 à 70 % | Monétaire + 0,5 à 1,5 point | Ligne révisable par le prêteur | Un pont, jamais une ligne permanente |
| Fonds 90/60 | Aucun pour le porteur | Le taux court, payé dans le fonds | Permanent, revendu quand on veut | Le seul levier structurel défendable |
| ETF à levier quotidien | Aucun, perte bornée à la mise | Le frein de volatilité, tous les jours | Quelques jours, pas plus | Non, définitivement |
| Marge de courtage | Oui, du jour au lendemain | Variable, au gré du courtier | À vue | Interdiction absolue |

**Le crédit immobilier conservé** est le levier le plus doux, avec son taux fixe et son assurance décès ([[immobilier-en-retrait]]). Ses deux conditions tiennent en une ligne, une mensualité logée dans un plancher détendu et un capital dû couvert plusieurs fois.

**Le lombard et l'avance d'assurance-vie** sont le même geste, emprunter contre son portefeuille sans le vendre. La version française est la plus praticable et la moins connue, puisqu'on emprunte à l'assureur contre son propre contrat, sans le racheter ni déclencher l'impôt. Le lombard bancaire est plus souple et remboursable par anticipation, mais le prêteur y reste maître de la quotité. C'est l'instrument des usages légitimes décrits plus bas, et c'est aussi celui des dérives, quand la ligne « temporaire » devient permanente.

**Les ETF à levier quotidien (×2, ×3)** sont le cas d'école du frein de volatilité, où 7 % de moyenne arithmétique ne laissent que 2,5 % composé ([[rendements-arithmetiques-geometriques]]). Des instruments de trading, pas des briques de plan.

**Les fonds « capital efficient » (90/60 et cousins).** Voilà la catégorie intéressante. Un tel fonds détient 90 % d'actions et 60 % d'obligations via des futures, soit 150 % d'exposition. Le levier de 1,5 est modéré et **permanent**, et il vit dans le fonds, pas sur votre compte. L'idée s'appelle le return stacking ([[return-stacking]]). On utilise le levier non pour sur-exposer au même risque, mais pour empiler des diversifiants sans réduire le moteur ([[actifs-defensifs]]). Un exemple : sur 100 €, 67 € en 90/60 plus 33 € en or ou trend donnent à peu près l'exposition d'un 60/40 classique, augmentée d'une vraie poche de régimes. La diversification par euro investi progresse. L'offre UCITS existe, avec les « Efficient Core » globaux, et elle s'étoffe. Trois vigilances restent de mise. Le coût du financement implicite d'abord : les futures paient le taux court, et quand ce taux est élevé, le levier coûte cher. La duration embarquée ensuite : le volet obligataire a encaissé 2022 comme les autres ([[obligations-en-retrait]]). La traçabilité enfin, avec les mêmes contrôles que pour tout fonds ([[etf-ucits-europeens]]).

Reste la seule interdiction absolue du chapitre, la marge de courtage brute et les options, le levier à appel de marge posé sur le cœur du plan. Un plan de retrait ne contient rien qui puisse être liquidé de force un mardi de panique.

## Les trois usages défendables, avec leurs chiffres

**Usage 1 : le pont de liquidité (le lombard comme matelas).** Le besoin est simple : financer 6 à 24 mois de dépenses pendant un creux sans vendre d'actifs dépréciés. C'est exactement le service du matelas de cash ([[cash-buffer]]). La version lombard supprime le matelas permanent : le capital reste investi et le coût d'opportunité disparaît. Quand le seuil de creux est franchi, le même que pour le rechargement ([[recharger-ou-pas]]), on tire sur la ligne pour vivre. On rembourse au retour du calme, par des ventes redevenues sereines. L'arithmétique est parlante. Emprunter 18 mois de dépenses, soit environ 6 % du portefeuille, à 4 % pendant 2 ans coûte au total un demi-point de portefeuille. Le matelas équivalent, lui, coûte 0,3 à 0,4 % par an, toutes les années. Sur le seul coût, le pont l'emporte donc tant que les creux ne s'enchaînent pas. Mais il perd la sécurité psychologique du cash ([[cash-buffer]]) : la vraie valeur du matelas est d'empêcher la panique, or une dette qui grossit pendant un krach fait l'inverse. Verdict : techniquement supérieur, mais comportementalement exigeant. Il vise les profils quantitatifs à plancher couvert, et l'avance d'assurance-vie en est la porte d'entrée.

**Usage 2 : l'efficacité de capital (la dose de 90/60).** Le besoin vient d'ailleurs : les protections de régime, or, trend et linkers, coûtent de la place. Une poche de régimes de 30 à 40 % réduit d'autant le moteur ([[portefeuilles-tous-temps]]). La version stackée loge le cœur actions-obligations dans une brique capital-efficient et consacre la place libérée aux diversifiants. À exposition moteur égale, le portefeuille gagne ses défenses. Le levier finance de la diversification, le seul achat qui améliore la composition sans gonfler la variance ([[rendements-arithmetiques-geometriques]]). C'est l'usage intellectuellement le plus propre. Bien dosé, avec un levier total du plan plafonné vers 1,35 et sans appel de marge, il tient plus de l'ingénierie d'allocation que de la spéculation. La vigilance principale : ne pas empiler l'illusion, car ajouter plus d'actions au lieu de diversifiants ramène au levier naïf.

**Usage 3 : le levier contracyclique d'ERN (volets 49 et 52).** L'idée : plutôt que vendre pendant les drawdowns, on emprunte temporairement au lombard et on rembourse à la reprise. Dans la version du volet 52, on n'ouvre la ligne que dans les creux profonds : au-delà de 20 à 30 % de drawdown, les retraits se financent par l'emprunt plutôt que par des ventes, et la ligne se rembourse au retour vers les sommets. Les simulations d'ERN créditent la version disciplinée d'une amélioration réelle mais modeste des pires millésimes, de l'ordre de 0,3 point de taux de retrait. Elles soulignent surtout l'évidence : tout dépend de l'exécution mécanique. C'est du rééquilibrage amplifié, avec le risque d'un instrument que l'on peut techniquement vous couper. Les quotités lombard fondent dans les krachs, et le prêteur réduit la ligne juste au moment où vous en avez besoin. Verdict : réservé aux plans surdimensionnés qui cherchent l'optimisation, jamais aux plans tendus qui cherchent le salut.

::: attention Les règles de sécurité non négociables
Pour les rares profils qui retiennent un usage, voici cinq règles écrites.
1) LTV total inférieur à 20-25 % du portefeuille au tirage maximal prévu. Les quotités des prêteurs fondent en crise, donc votre plafond doit rester loin du leur.
2) Jamais d'instrument à liquidation forcée sur le cœur du plan. Avance d'assurance-vie et lombard à quotité confortable, oui ; marge de courtage, non.
3) Le coût est un spread connu sur taux court, à recalculer si les taux courts changent de régime. Le pont à 1 % de 2021 et le pont à 4,5 % de 2023 ne sont pas le même produit.
4) Un plan de sortie daté pour chaque usage. Le pont se rembourse au désarmement du seuil, le stack se dénoue si l'offre de fonds se dégrade.
5) Le test du conjoint : si la personne qui héritera de l'exécution ne peut pas expliquer la ligne en deux phrases, la ligne n'existe pas ([[couple-et-famille]], [[construire-son-plan]]).
:::

## Modéliser le levier, et un exemple

**La modélisation.** Chiffrer un plan à levier demande deux outils distincts, et la première erreur est de leur poser la même question. L'étage du portefeuille d'abord. Un simulateur d'allocation rejoue votre portefeuille ligne par ligne sur l'historique, et les outils complets y acceptent des poids dont la somme dépasse 100 %, avec un coût d'emprunt déduit sur la part financée. C'est là, et seulement là, que se fait la comparaison honnête entre un cœur classique et un cœur stacké, frais de financement compris. L'étage du plan ensuite. Un moteur de décumulation ne voit plus les lignes, il déroule les retraits sur le portefeuille agrégé ([[la-machine-pofo]] en décrit un en détail). Le levier y entre donc en amont, par les rendements du portefeuille une fois stacké, soit l'historique du fonds capital-efficient lui-même, soit une moyenne et une volatilité relevées à la main. Le pont lombard, lui, s'y approxime par un matelas de sécurité aux mêmes flux, mais dont le rendement est rendu négatif à hauteur du coût de la ligne. Les deux étages dans cet ordre, jamais le second sans le premier.

::: exemple Deux usages, deux verdicts
Cas A, le pont. Mireille, 58 ans, 1,8 M€ dont 600 k€ d'assurance-vie, plancher couvert aux deux tiers par une pension proche. Elle remplace son projet de matelas de 36 mois par 12 mois de fonds euros et une avance d'assurance-vie documentée, à taux fixé et quotité 60 %, tirée seulement si le drawdown dépasse 20 %. Résultat de la simulation : ruine inchangée et environ 0,25 % par an de coût d'opportunité rendu au moteur. Adopté, avec la règle de remboursement écrite.

Cas B, le levier de rendement. Bruno, 52 ans, plan **tendu** à 4,2 % de retrait, veut « compenser » par un cœur ×1,5. La simulation est sans appel. Sous un stress de séquence, la ruine double, car le levier amplifie précisément les chemins qui tuent le plan. Refusé. Son problème est le taux de retrait, pas l'outillage ([[combien-il-vous-faut]]). Le levier aide parfois les plans riches, il achève les plans pauvres. C'est sa seule loi fiable.
:::

## L'essentiel à retenir

- Le levier change de signe à la retraite. Il amplifie le frein de volatilité et la séquence, et il introduit la liquidation forcée. « Aucun levier » est la bonne réponse par défaut, mais en conclusion, pas en tabou.
- Trois usages défendables, et trois seulement. Le **pont** de liquidité : lombard ou avance d'assurance-vie en creux déclenché, le matelas sans coût d'opportunité, pour profils quantitatifs. L'**efficacité** de capital : le 90/60 qui finance des diversifiants, levier total plafonné vers 1,35, sans appel de marge. Le **contracyclique** d'ERN : emprunter au creux plutôt que vendre, pour optimiser des plans déjà sûrs.
- Les interdits : ETF à levier quotidien où le frein broie, marge de courtage sur le cœur avec sa liquidation forcée, et tout levier au service d'un plan **tendu**. Le levier aide les riches et achève les pauvres.
- Cinq règles en cas d'usage : LTV inférieur à 20-25 % au tirage maximal prévu, zéro liquidation forcée possible, spread connu et retesté, sortie datée, test du conjoint.
- Le levier se chiffre à deux étages. Côté portefeuille, avec des poids qui dépassent 100 % et un coût d'emprunt explicite, pour comparer un cœur stacké à un cœur classique. Côté plan, sur le portefeuille agrégé, où le pont se représente par un matelas au rendement négatif. Chiffrez avant, comme tout le reste. C'est la différence entre un instrument et un pari.

---

## Pour aller plus loin

- Early Retirement Now, volets 49 et 52 (« Leverage in Retirement », « Timing Leverage ») : le dossier contracyclique complet ([[serie-ern]]).
- Ayres & Nalebuff, *Lifecycle Investing* (2010) : la théorie du levier par âge. Et pourquoi elle s'inverse en décumulation.
- [returnstacked.com](https://www.returnstacked.com) et la littérature sur la capital efficiency (Newfound, ReSolve) : le stacking expliqué par ses promoteurs, à lire avec l'esprit critique de rigueur.
- Dans ce livre : [[immobilier-en-retrait]] (le crédit conservé), [[cash-buffer]] et [[recharger-ou-pas]] (ce que le pont remplace), [[rendements-arithmetiques-geometriques]] (le frein qui gouverne tout).
