# Le pourcentage fixe du portefeuille : increvable mais inconfortable

Retirez chaque année 4 % non pas d'un montant historique indexé, mais du portefeuille **tel qu'il est** : voilà le pourcentage fixe, la stratégie la plus simple après celle de Bengen et son exact miroir sur le triangle impossible ([[panorama-strategies-retrait]]). Elle possède la propriété la plus rassurante de toute la décumulation, l'impossibilité mathématique de la ruine, et la vend au prix le plus visible : votre train de vie devient une fonction affine des marchés.

Cet article la démonte complètement : pourquoi elle ne peut pas ruiner (et pourquoi cette garantie vaut moins qu'il n'y paraît), la vraie nature de son risque (la « ruine de train de vie », que les tableaux de succès ne montrent jamais), le choix du pourcentage, les techniques de lissage que les fondations universitaires ont mises au point depuis cinquante ans pour la rendre vivable (la règle de Yale, transposable telle quelle à un particulier), pour qui elle est réellement adaptée, et sa place dans pofo. Elle mérite l'examen : dominée à l'état pur, elle est l'ingrédient de base de la moitié des règles modernes ([[vpw]], [[amortissement-abw]], [[plancher-plafond]] en descendent toutes).

::: cle La règle et sa propriété
Chaque année : retrait = w × portefeuille courant (w fixé une fois, par exemple 4 %). La ruine est impossible par construction : w % de quelque chose n'est jamais tout, le portefeuille tend vers zéro sans jamais l'atteindre. Mais relisez la phrase. C'est le **portefeuille** qui ne meurt pas ; votre revenu, lui, suit chaque soubresaut. La règle ne supprime pas le risque, elle le déplace intégralement du capital vers la vie quotidienne. C'est le pôle opposé à Bengen, et le point le plus à gauche (et le plus haut) de la frontière de décumulation.
:::

## Pourquoi elle ne peut pas ruiner, et ce que ça vaut vraiment

La preuve tient en une ligne : après retrait et rendement r, le portefeuille vaut P × (1 − w) × (1 + r) ; c'est un produit de facteurs strictement positifs, donc jamais zéro. L'intuition économique est plus parlante : la règle vend toujours une **fraction**, jamais un montant. Quand le portefeuille fond de moitié, le prélèvement fond de moitié aussi, et la pression sur le capital reste constante. Le mécanisme des ciseaux qui tue la règle fixe (retrait qui monte pendant que le portefeuille baisse, [[retrait-fixe-bengen]]) est désamorcé à la racine.

Mais que vaut cette garantie ? Regardons ce qu'elle garantit **exactement** : que le portefeuille reste positif. Elle ne garantit ni son niveau (il peut osciller sous 30 % de sa valeur initiale réelle pendant une décennie, le millésime 1966 sous pourcentage fixe voit le revenu réel fondre de moitié et y rester des années), ni surtout votre capacité à vivre du prélèvement. Un revenu de w × (presque rien) est un revenu de presque rien : la ruine formelle est remplacée par un appauvrissement continu et parfaitement légal. La littérature appelle cela la **ruine de train de vie** (« lifestyle ruin ») : le portefeuille survit, le plan de vie meurt. Tout jugement honnête du pourcentage fixe se fait donc sur la distribution du **revenu** servi, jamais sur le taux de succès (100 % par définition, et parfaitement creux). C'est le critère n° 3 de la grille de notation ([[panorama-strategies-retrait]]), et la §04 est la vue faite pour ça ([[utiliser-la-page-fire]]).

Il faut quand même créditer la règle de deux vertus réelles derrière sa garantie. D'une part, elle est **contra**-cyclique côté portefeuille. Elle prélève peu au creux (protégeant la reprise) et beaucoup au sommet (écrémant l'euphorie) : exactement l'inverse du fixe indexé, et la raison profonde pour laquelle elle laisse le capital si robuste. C'est une qualité anti-séquence authentique ([[sequence-des-rendements]]). D'autre part, elle s'auto-corrige face aux erreurs d'hypothèses : si les rendements déçoivent structurellement ([[rendements-attendus]]), le revenu s'ajuste en continu au monde réel au lieu d'accumuler une dette silencieuse jusqu'à la falaise. Ces deux vertus sont l'héritage que toutes ses descendantes cherchent à conserver en domestiquant sa volatilité.

## Le vrai risque, chiffré : votre revenu a la volatilité de votre portefeuille

Le revenu étant w × portefeuille, sa volatilité est celle du portefeuille : 11 % de volatilité de portefeuille = un train de vie qui bouge de ±11 % les années ordinaires, et qui suit les drawdowns dans toute leur profondeur. Les ordres de grandeur à intérioriser, pour un 60/40 mondial :

| Épisode | Drawdown réel du portefeuille | Votre revenu sous pourcentage fixe |
|---|---|---|
| Correction ordinaire (tous les 2-3 ans) | −10 à −15 % | −10 à −15 % pendant 1-2 ans |
| Krach type 2008 | −30 à −35 % | −30 % pendant 2-4 ans |
| Régime hostile type 1966-1981 | −40 à −50 % au pire, une décennie sous l'eau | revenu réel amputé d'un tiers à moitié pendant ~10 ans |

La dernière ligne est la vraie question d'admissibilité : **pouvez-vous, structurellement, vivre dix ans à 55-65 % de votre confort ?** Pour un ménage dont le plancher est à 90 % du confort ([[combien-il-vous-faut]]), la réponse est non, et le pourcentage fixe pur est inadmissible quel que soit son taux de succès. Pour un ménage dont la pension couvre déjà le plancher et dont le portefeuille ne finance que le surplus ([[revenus-complementaires]]), la réponse peut être oui, et la règle devient soudain très attractive. C'est le profil type de la phase adossée d'un plan FIRE ([[horizon-et-esperance-de-vie]]).

Notez aussi le renversement psychologique : sous Bengen, l'angoisse porte sur un événement lointain et binaire (la falaise) ; sous pourcentage fixe, elle porte sur le prochain relevé (mon revenu de l'an prochain). Les deux stress ne conviennent pas aux mêmes personnes ([[psychologie-du-retrait]]) ; aucun n'est objectivement moindre.

## Le lissage : cinquante ans d'ingénierie des fondations

Le pourcentage fixe a un utilisateur institutionnel historique : les fondations et fonds de dotation universitaires, qui doivent par nature durer perpétuellement (donc, jamais de ruine) tout en finançant des budgets stables (donc, lisser). Leur demi-siècle d'expérience a produit des techniques directement transposables, et c'est le chaînon manquant entre le pourcentage fixe brut et les règles modernes.

**La moyenne mobile.** La version la plus simple : prélever w × (moyenne du portefeuille sur les 12 derniers trimestres) au lieu du dernier point. Un krach de 30 % ne se répercute plus qu'en trois ans d'à-coups de ~10 % : la volatilité du revenu est grossièrement divisée par la racine de la fenêtre de lissage. Prix : le prélèvement « en retard » mord un peu plus le capital au creux (on prélève sur une moyenne encore haute), ce qui laisse une toute petite probabilité de trajectoires très dégradées. Le lissage revend une miette de la garantie contre beaucoup de confort.

**La règle de Yale (Tobin).** Le standard des grandes dotations, d'une élégance remarquable : retrait = 70 % × (retrait de l'an dernier, indexé inflation) + 30 % × (w × portefeuille courant). C'est un lissage exponentiel : chaque année, le revenu ne fait que 30 % du chemin vers sa cible proportionnelle. Les propriétés sont exactement l'hybride recherché : le revenu a l'inertie du fixe indexé à court terme (70 % de mémoire) et la vérité du pourcentage fixe à long terme (il converge vers w × portefeuille en 4-5 ans). Un krach de 30 % ne coupe le revenu que de ~9 % la première année, ~16 % cumulés la deuxième : la pente est vivable, et la direction honnête.

**Le corridor.** Troisième école : prélever w × portefeuille, mais borner la variation annuelle du revenu (pas plus de +5 %, pas moins de −2,5 % en réel d'une année sur l'autre). C'est exactement la règle « dynamic spending » de Vanguard, à qui ce livre consacre son article ([[plancher-plafond]]) : le corridor réintroduit une ruine possible (la descente plafonnée peut ne pas suivre un effondrement), et c'est un choix assumé, un point intermédiaire de la frontière.

Ces trois techniques racontent la même leçon : le pourcentage fixe brut n'est pas une règle terminale, c'est une **matière première**. Lissée par moyenne, par mémoire ou par corridor, elle donne les règles du milieu de la frontière ; croisée avec l'horizon restant, elle donne la famille actuarielle ([[vpw]], [[amortissement-abw]]).

::: science Choisir w : la borne géométrique
Quel pourcentage ? La théorie donne une borne claire : à long terme, le portefeuille sous pourcentage fixe croît en réel si et seulement si w < rendement réel **géométrique** espéré ([[rendements-arithmetiques-geometriques]]). À w = 3-3,5 % contre un géométrique réel de ~3,5-4,5 % pour un portefeuille diversifié ([[rendements-attendus]]), le revenu réel **médian** est stable ou croissant ; à w = 5-6 %, il s'érode tendanciellement (chaque année prélève plus que la croissance, le revenu suit le capital vers le bas, sans jamais l'annuler). Il n'y a pas de falaise à franchir, donc w peut légitimement être plus généreux que le taux de Bengen : w = 4-4,5 % est défendable là où un fixe indexé exigerait 3,25-3,5 %. C'est le dividende de l'auto-correction. La pratique des dotations (4,5-5 % lissé, pour des portefeuilles plus agressifs) confirme l'ordre de grandeur.
:::

## Pour qui, et dans pofo

**Les bons profils.** Le pourcentage fixe (lissé !) convient à trois situations : le plancher couvert par ailleurs (pension, rente, le portefeuille ne finance que le compressible, le profil de la retraite française installée) ; les budgets naturellement élastiques (dépenses discrétionnaires dominantes, capacité réelle à voyager moins les mauvaises années, [[flexibilite-realite]]) ; et l'objectif perpétuel (transmettre un capital réel intact, la logique de dotation, [[succession-et-transmission]]). Il est contre-indiqué en phase à découvert d'un FIRE tendu (plancher élevé financé à 100 % par le portefeuille). C'est là que sa ruine de train de vie mord exactement comme la vraie.

**Dans pofo** : le curseur « Spend % of portfolio (VPW) » du tiroir Spending policy applique le pourcentage fixe pur (il écrase le besoin fixe et les règles flex/guardrails/ratchet) ; la frontière §06 le positionne contre les autres règles pour **votre** plan, et la §04 montre la distribution du revenu servi. C'est elle qui dit si la règle est admissible pour vous. La ruine affichée sera ~0 : vous savez maintenant que ce zéro ne se lit pas seul ([[utiliser-la-page-fire]]). Pour approcher un Yale lissé, la règle bornée de Vanguard (case « Bounded % of portfolio ») en est le cousin direct simulable nativement ([[plancher-plafond]]).

::: exemple Le même régime hostile, brut contre lissé
Portefeuille 1,4 M€, w = 4 % (56 000 €). Scénario type 1973-1974 : −40 % réel en deux ans, reprise lente ensuite. Pourcentage fixe brut : revenu 56 000 → 44 000 → 34 000 € en deux ans (−39 %), remontée au rythme du marché, et sept ans passés sous 45 000 €. Règle de Yale (70/30) : 56 000 → 52 900 → 48 600 € (−13 % en deux ans), plancher vers 44 000 € en année 4, remontée entamée avant même le retour du marché. Même portefeuille, même « garantie » anti-ruine, mais la version lissée transforme une chute libre en pente douce : dans un budget avec 25 % de compressible, la première est une crise de plan, la seconde une gestion courante. Le lissage n'est pas un raffinement. Il est la condition d'admissibilité de toute la famille proportionnelle.
:::

## L'essentiel à retenir

- Retrait = w × portefeuille courant : la ruine du **capital** est mathématiquement impossible ; celle du **train de vie** ne l'est pas. Jugez cette règle sur la distribution du revenu servi (§04), jamais sur son taux de succès de 100 %, parfaitement creux.
- Son revenu a la volatilité du portefeuille : la question d'admissibilité est « puis-je vivre dix ans à 55-65 % du confort dans un régime hostile ? » : oui si le plancher est couvert par ailleurs, non en phase à découvert tendue.
- Ses deux vertus profondes, héritées par toutes les règles modernes : contra-cyclicité (prélève peu au creux) et auto-correction face aux erreurs d'hypothèses.
- Le lissage la rend vivable : moyenne mobile, règle de Yale (70 % mémoire + 30 % cible, le standard des dotations), ou corridor borné (la version Vanguard, [[plancher-plafond]]) ; w peut être plus généreux que le taux de Bengen (borne, le rendement géométrique réel espéré).
- À l'état pur, c'est une matière première plus qu'une stratégie ; ses descendantes domestiquées ([[vpw]], [[amortissement-abw]], [[plancher-plafond]]) occupent le milieu de la frontière.

---

## Pour aller plus loin

- James Tobin et la règle de dépense des dotations (« spending rule » de Yale) : les rapports annuels du Yale Endowment en décrivent la version en vigueur.
- Early Retirement Now, volet 10 (le pourcentage fixe face à Guyton-Klinger) et volet 11 (les critères) ([[serie-ern]]).
- Bogleheads wiki, « Variable percentage withdrawal » : la généalogie proportionnelle → actuarielle ([[vpw]]).
- Dans ce livre : [[plancher-plafond]] (le corridor industrialisé), [[amortissement-abw]] (le pourcentage rendu conscient de l'horizon), et [[flexibilite-realite]] (ce que « vivre la variabilité » veut dire).
