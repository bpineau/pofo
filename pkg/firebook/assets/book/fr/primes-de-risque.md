# D'où viennent les rendements : les primes de risque

Tout ce livre repose sur une hypothèse si fondamentale qu'on oublie de la regarder : un portefeuille d'actions et d'obligations **rapporte quelque chose**, durablement, au-dessus de l'inflation. Sans elle, pas de règle des 4 %, pas de FIRE, rien. Cet article la sort de l'ombre et pose les trois questions qui fâchent. Pourquoi les actions rapportent-elles plus que les obligations, et les obligations plus que le cash ? Qui paie ces écarts, et pourquoi continuerait-il de payer ? Et pourquoi certains actifs (l'or en tête) ne rapportent-ils **rien** en termes réels sans que ce soit un défaut ?

La réponse tient en un mot, la **prime de risque**, et en une discipline : chaque ligne de votre portefeuille doit pouvoir nommer la prime qu'elle récolte et la raison pour laquelle cette prime survivra à sa propre célébrité. Après lecture, vous saurez auditer un portefeuille ligne par ligne avec cette grille, ce qui est la meilleure défense connue contre les produits qui promettent du rendement sans dire qui le paie.

::: cle L'idée en une phrase
Un actif ne rapporte pas parce qu'il est « bon », il rapporte parce qu'il fait **mal au mauvais moment** et que quelqu'un doit être payé pour accepter cette douleur. Les actions rapportent 4 à 6 points de plus que le cash précisément parce qu'elles peuvent perdre la moitié de leur valeur au moment où vous perdez votre emploi. Le rendement est le salaire du risque porté, pas une propriété magique de l'actif ; c'est aussi pourquoi les primes ne disparaissent pas quand tout le monde les connaît, car les connaître n'enlève pas la douleur.
:::

## La prime actions : la plus grande, la mieux payée, la plus documentée

L'écart de rendement entre les actions et le cash (la **prime de risque actions**) est le fait empirique central de la finance. Sur un siècle et demi et sur tous les pays développés, il ressort entre 4 et 6 points par an en moyenne géométrique ([[anarkulova-cederburg]] pour l'échantillon mondial, moins flatteur que le seul cas américain). Composé sur trente ans, cet écart transforme 1 € en 4 à 6 € de plus que le placement monétaire, c'est le moteur de tout plan FIRE.

Pourquoi existe-t-elle ? La théorie donne une réponse précise : ce qui compte n'est pas la volatilité en soi, mais la **covariance avec les mauvais états du monde**. Les actions s'effondrent pendant les récessions, quand les revenus baissent, quand le chômage monte, quand votre voisin vend sa maison, exactement quand un euro de plus aurait le plus de valeur. Un actif qui trahit dans ces moments doit offrir une compensation élevée pour trouver preneur. S'y ajoute le **risque de désastre** : les marchés actions nationaux peuvent perdre 80 à 100 % (Allemagne 1948, Japon 1945-1949, Russie 1917), et une part de la prime rémunère ces queues rarement observées mais jamais abolies ([[queues-epaisses]], [[diversification-internationale]]).

Fait remarquable, la prime observée est même **trop grosse** pour la théorie standard : c'est le célèbre « equity premium puzzle » (Mehra et Prescott, 1985), toujours débattu quarante ans après. Pour le rentier, le puzzle est plutôt une bonne nouvelle épistémique : une prime que la théorie peine à justifier entièrement par le risque a peu de chances d'être arbitrée à zéro, car personne ne peut « acheter » la disparition des récessions. Les meilleures raisons de la voir persister sont structurelles → l'aversion au risque humaine ne se met pas à jour, les horizons courts des institutions les empêchent de porter le risque long, et le stock mondial d'épargne cherche massivement la sécurité. La prudence honnête consiste à la projeter plus basse que l'histoire (3 à 4 points au-dessus du cash aux valorisations actuelles, [[rendements-attendus]], [[valorisations-et-cape]]), pas à zéro.

## La prime de terme, la prime de crédit, et les petites monnaies

**La prime de terme** rémunère la détention d'obligations longues plutôt que de cash : le porteur accepte le risque de taux et, surtout, le risque d'**inflation** sur des flux nominaux fixes ([[obligations-en-retrait]]). Historiquement elle vaut 1 à 2 points, mais c'est la prime la plus instable du catalogue : elle a été somptueuse pendant les quarante ans de désinflation 1981-2021, négative au creux de 2020 (des obligations rendant moins que le cash attendu, on payait pour le privilège), redevenue positive depuis. La règle pratique : elle se lit en direct dans la pente de la courbe des taux, et un étage obligataire ne mérite sa place que quand elle existe ([[return-stacking]] applique ce critère à la lettre).

**La prime de crédit** rémunère le risque de défaut des obligations d'entreprises. Son secret honteux est sa petitesse une fois les défauts réalisés déduits : 0,5 à 1 point net pour l'investment grade, guère plus pour le high yield, avec une corrélation aux actions qui monte en flèche dans les crises (le défaut arrive en récession). C'est pourquoi ce livre traite le high yield en faux défensif ([[actifs-defensifs]]) : il empile la prime actions et la prime de crédit dans le même mauvais état du monde, au prix obligataire.

**La prime d'illiquidité** (private equity, dette privée, immobilier non coté) rémunère l'impossibilité de vendre. Elle est réelle en théorie et souvent illusoire en pratique pour le particulier : les frais des véhicules accessibles la consomment, et le lissage comptable des valorisations la maquille en basse volatilité ([[immobilier-en-retrait]] pour le cas immobilier). Pour un rentier, dont le métier est précisément de **vendre régulièrement**, l'illiquidité est en outre un coût de premier ordre, pas un détail.

**Les primes alternatives** (trend, carry, value) ont leur chapitre ([[global-macro]], [[managed-futures]], [[facteurs-fama-french]]). Leur particularité est d'être partiellement **comportementales** : elles rémunèrent moins un risque macro qu'une erreur persistante des autres participants (sous-réaction, recherche de loterie, contraintes institutionnelles). C'est ce qui les rend à la fois précieuses (décorrélées des mauvais états du monde) et plus fragiles (une erreur peut s'apprendre, une contrainte peut sauter).

::: science La décomposition après publication : McLean et Pontiff
Que devient une prime une fois publiée dans une revue académique ? McLean et Pontiff (« Does Academic Research Destroy Stock Return Predictability? », 2016) ont mesuré la réponse sur 97 anomalies : leur rendement baisse d'environ **un tiers à la moitié** après publication, mais ne tombe pas à zéro. La lecture correcte est une hiérarchie de robustesse. Les primes adossées à un risque macro non assurable (actions, terme) sont les plus solides, personne ne pouvant arbitrer les récessions. Les primes comportementales à limites d'arbitrage documentées (trend, value) se compriment mais survivent, des décennies après leur publication. Les anomalies statistiques fines sans histoire de risque ni de comportement meurent, et c'était souvent du data mining ([[pieges-des-simulateurs]]). Moralité pour le plan : fonder le gros du rendement sur les primes du premier étage, doser celles du deuxième, ignorer le troisième.
:::

## Pourquoi l'or ne rapporte rien (et pourquoi on en veut quand même)

L'or est le contre-exemple pédagogique parfait. Pas de flux, pas de dividende, pas de coupon, personne à qui transférer un risque contre salaire : il n'y a **pas de prime de risque de l'or**, et son rendement réel séculaire est proche de zéro ([[or-en-retrait]]). Ce n'est pas un défaut caché, c'est sa définition. L'or n'est pas un actif de rendement, c'est une **monnaie alternative** dont le prix en euros monte quand la confiance dans les monnaies officielles baisse. On ne le détient pas pour sa prime (il n'en a pas) mais pour sa **corrélation** : il paie dans des états du monde (crises monétaires, répression financière, stagflation) où presque tout le reste trahit. Le même raisonnement, sans le millénaire d'historique, s'applique aux cryptomonnaies : zéro flux, zéro prime théorique, un pur pari monétaire et comportemental, ce qui interdit de les dimensionner comme un actif de rendement dans un plan de retrait.

Le cash ferme la marche : il est l'étalon (les primes se mesurent au-dessus de lui) et rapporte, en réel, à peu près zéro sur longue période, avec de longs épisodes négatifs quand la répression financière s'en mêle ([[inflation-histoire]]). Le détenir a un rôle (la manœuvre, [[cash-buffer]]) mais aucun rendement à attendre.

## L'audit de votre portefeuille, prime par prime

La grille tient en trois questions par ligne : quelle prime cette position récolte-t-elle, qui la paie, et pourquoi le paiera-t-il encore dans vingt ans ? Trois réponses types montrent la méthode. « ETF monde » → prime actions, payée par l'économie réelle via les profits, persistante car adossée au risque macro : la réponse parfaite, c'est le cœur du plan. « Fonds thématique intelligence artificielle » → aucune prime identifiable au-delà de la prime actions déjà détenue, un pari sectoriel payé plus cher en frais : la ligne double un risque existant sans salaire supplémentaire. « Produit structuré à capital protégé » → le porteur **vend** de la convexité et **achète** du risque de crédit bancaire, primes négatives des deux côtés une fois les marges prélevées : la grille vient de vous économiser des années de rendement.

::: exemple Le portefeuille du chapitre, audité
Le portefeuille type de ce livre (60 % actions monde, 25 % obligations dont linkers, 7,5 % or, 7,5 % trend) se lit ainsi. Actions → prime actions, 3-4 points réels attendus au-dessus du cash, le moteur. Obligations → prime de terme (positive à nouveau) + le contrat réel des linkers ([[obligations-indexees]]), l'amortisseur des récessions désinflationnistes. Or → aucune prime, achat de corrélation pour les régimes monétaires hostiles. Trend → prime comportementale documentée un siècle, comprimée mais vivante, l'assurance à espérance positive des régimes longs. Chaque ligne nomme sa prime ou son rôle, aucune ne double une autre au même moment de douleur → c'est exactement ce que « diversifié » veut dire, et la suite logique se lit dans [[pourquoi-la-diversification-marche]].
:::

## L'essentiel à retenir

- Le rendement est le salaire d'un risque porté : les actifs qui font mal dans les mauvais états du monde (actions, crédit) doivent payer une prime pour trouver preneur ; ceux qui n'en font porter aucun (cash) ou servent de monnaie (or) ne rapportent rien en réel, par construction.
- La prime actions (4-6 points historiques au-dessus du cash, 3-4 prudents en prospectif) est la mieux documentée et la plus robuste, car personne ne peut arbitrer les récessions ; la prime de terme se lit dans la pente de la courbe, la prime de crédit nette est petite et mal placée.
- Les primes publiées se compriment (McLean-Pontiff, environ −30 à −50 %) mais celles adossées à un risque macro ou à une limite d'arbitrage durable survivent ; les anomalies sans histoire meurent.
- L'or et les cryptomonnaies n'ont pas de prime : on peut détenir le premier pour sa corrélation aux régimes monétaires hostiles, mais aucun des deux ne se dimensionne comme un actif de rendement.
- La grille d'audit : pour chaque ligne, quelle prime, qui la paie, pourquoi encore dans vingt ans. Une ligne sans réponse est un pari ou une commission déguisée, et le portefeuille de retrait n'a de place ni pour l'un ni pour l'autre.

---

## Pour aller plus loin

- Antti Ilmanen, Expected Returns (2011) et Investing Amid Low Expected Returns (2022) : le traité de référence des primes, classe par classe.
- Mehra & Prescott, « The Equity Premium: A Puzzle » (1985) ; Dimson, Marsh & Staunton, le Global Investment Returns Yearbook (les primes sur 120 ans et 20 pays).
- McLean & Pontiff, « Does Academic Research Destroy Stock Return Predictability? » (2016) : ce que deviennent les primes publiées.
- Dans ce livre : [[rendements-attendus]] (chiffrer les primes prospectives), [[pourquoi-la-diversification-marche]] (les assembler), [[actifs-defensifs]] (les rôles sans prime), [[anarkulova-cederburg]] (la prime actions hors du seul cas américain).
