# Le pourcentage fixe du portefeuille : increvable mais inconfortable

Chaque année, retirez 4 % du portefeuille tel qu'il est, et non d'un montant historique indexé. Voilà le pourcentage fixe. C'est la stratégie la plus simple après celle de Bengen, et son miroir exact sur le triangle impossible ([[panorama-strategies-retrait]]). Elle offre la propriété la plus rassurante de toute la décumulation, l'impossibilité mathématique de la ruine. Mais elle la vend au prix le plus visible : votre train de vie suit directement les marchés.

Cet article la démonte pièce par pièce. Pourquoi elle ne peut pas ruiner, et pourquoi cette garantie vaut moins qu'il n'y paraît. La vraie nature de son risque, la « ruine de train de vie » que les tableaux de succès ne montrent jamais. Le choix du pourcentage. Les techniques de lissage que les fondations universitaires affinent depuis cinquante ans pour la rendre vivable, dont la règle de Yale, transposable telle quelle à un particulier. Pour qui elle est vraiment adaptée, enfin, et quelle place lui donner dans un plan. Elle mérite l'examen. Dominée à l'état pur, elle reste l'ingrédient de base de la moitié des règles modernes : [[vpw]], [[amortissement-abw]] et [[plancher-plafond]] en descendent toutes.

::: cle La règle et sa propriété
Chaque année, le retrait vaut w × portefeuille courant, avec w fixé une fois pour toutes, par exemple 4 %. La ruine est impossible par construction : w % de quelque chose n'est jamais tout. Le portefeuille tend vers zéro sans jamais l'atteindre. Mais relisez la phrase. C'est le **portefeuille** qui ne meurt pas. Votre revenu, lui, suit chaque soubresaut. La règle ne supprime pas le risque, elle le déplace intégralement du capital vers la vie quotidienne. C'est le pôle opposé à Bengen, le point le plus à gauche et le plus haut de la frontière de décumulation.
:::

## Pourquoi elle ne peut pas ruiner, et ce que ça vaut vraiment

La preuve tient en une ligne. Après retrait et rendement r, le portefeuille vaut P × (1 − w) × (1 + r), un produit de facteurs strictement positifs, donc jamais nul. L'intuition économique est plus parlante : la règle vend toujours une **fraction**, jamais un montant. Quand le portefeuille fond de moitié, le prélèvement fond de moitié aussi, et la pression sur le capital reste constante. Le mécanisme de ciseaux qui tue la règle fixe, ce retrait qui monte pendant que le portefeuille baisse ([[retrait-fixe-bengen]]), est ainsi désamorcé à la racine.

Mais que vaut cette garantie ? Regardons ce qu'elle assure exactement : que le portefeuille reste positif. Elle ne garantit pas son niveau. Il peut osciller sous 30 % de sa valeur initiale réelle pendant une décennie. Le millésime 1966, sous pourcentage fixe, voit le revenu réel fondre de moitié et y rester des années. Elle ne garantit surtout pas votre capacité à vivre du prélèvement. Un revenu de w × (presque rien) reste un revenu de presque rien. La ruine formelle laisse alors place à un appauvrissement continu et parfaitement légal. La littérature appelle cela la **ruine de train de vie** (« lifestyle ruin ») : le portefeuille survit, le plan de vie meurt. Tout jugement honnête du pourcentage fixe porte donc sur la distribution du **revenu** servi, jamais sur le taux de succès, égal à 100 % par définition et parfaitement creux. C'est le critère n° 3 de la grille de notation ([[panorama-strategies-retrait]]). La vue de la distribution du revenu est faite pour cela ([[utiliser-la-page-fire]]).

Il faut quand même créditer la règle de deux vertus réelles, derrière sa garantie. D'une part, elle est **contracyclique** côté portefeuille. Elle prélève peu au creux, ce qui protège la reprise, et beaucoup au sommet, ce qui écrème l'euphorie. C'est l'exact inverse du fixe indexé, et la raison profonde pour laquelle elle laisse le capital si robuste. Voilà une vraie qualité anti-séquence ([[sequence-des-rendements]]). D'autre part, elle s'auto-corrige face aux erreurs d'hypothèses. Si les rendements déçoivent durablement ([[rendements-attendus]]), le revenu s'ajuste en continu au monde réel, au lieu d'accumuler une dette silencieuse jusqu'à la falaise. Ces deux vertus forment l'héritage que toutes ses descendantes cherchent à conserver, en domestiquant sa volatilité.

## Le vrai risque, chiffré : votre revenu a la volatilité de votre portefeuille

Le revenu vaut w × portefeuille, donc sa volatilité est celle du portefeuille. Une volatilité de 11 % donne un train de vie qui bouge de ±11 % les années ordinaires, et qui suit les drawdowns dans toute leur profondeur. Voici les ordres de grandeur à intérioriser, pour un 60/40 mondial :

| Épisode | Drawdown réel | Votre revenu |
|---|---|---|
| Correction ordinaire (tous les 2-3 ans) | −10 à −15 % | −10 à −15 % pendant 1-2 ans |
| Krach type 2008 | −30 à −35 % | −30 % pendant 2-4 ans |
| Régime hostile type 1966-1981 | −40 à −50 % au pire, une décennie sous l'eau | revenu réel amputé d'un tiers à moitié pendant ~10 ans |

La dernière ligne pose la vraie question d'admissibilité : **pouvez-vous, structurellement, vivre dix ans à 55-65 % de votre confort ?** Pour un ménage dont le plancher est à 90 % du confort ([[combien-il-vous-faut]]), la réponse est non. Le pourcentage fixe pur est alors inadmissible, quel que soit son taux de succès. Pour un ménage dont la pension couvre déjà le plancher, et dont le portefeuille ne finance que le surplus ([[revenus-complementaires]]), la réponse peut être oui. La règle devient soudain très attractive. C'est le profil type de la phase adossée d'un plan FIRE ([[horizon-et-esperance-de-vie]]).

Notez aussi le renversement psychologique. Sous Bengen, l'angoisse porte sur un événement lointain et binaire, la falaise. Sous pourcentage fixe, elle porte sur le prochain relevé, mon revenu de l'an prochain. Ces deux stress ne conviennent pas aux mêmes personnes ([[psychologie-du-retrait]]), et aucun n'est objectivement moindre.

## Le lissage : cinquante ans d'ingénierie des fondations

Le pourcentage fixe a un utilisateur institutionnel historique : les fondations et les fonds de dotation universitaires. Ils doivent par nature durer perpétuellement, donc ne jamais se ruiner, tout en finançant des budgets stables, donc en lissant. Leur demi-siècle d'expérience a produit des techniques directement transposables. C'est le chaînon manquant entre le pourcentage fixe brut et les règles modernes.

**La moyenne mobile.** C'est la version la plus simple. On prélève w × la moyenne du portefeuille sur les 12 derniers trimestres, au lieu du dernier point. Un krach de 30 % ne se répercute plus qu'en trois ans d'à-coups de ~10 %. La volatilité du revenu est ainsi divisée, grossièrement, par la racine de la fenêtre de lissage. Le prix ? Le prélèvement « en retard » mord un peu plus le capital au creux, puisqu'on prélève sur une moyenne encore haute. Cela laisse une toute petite probabilité de trajectoires très dégradées. Le lissage revend une miette de la garantie contre beaucoup de confort.

**La règle de Yale (Tobin).** C'est le standard des grandes dotations, d'une élégance remarquable. Le retrait vaut 70 % × (retrait de l'an dernier, indexé sur l'inflation) + 30 % × (w × portefeuille courant). C'est un lissage exponentiel : chaque année, le revenu ne fait que 30 % du chemin vers sa cible proportionnelle. Les propriétés sont exactement l'hybride recherché. À court terme, le revenu a l'inertie du fixe indexé, avec ses 70 % de mémoire. À long terme, il retrouve la vérité du pourcentage fixe, car il converge vers w × portefeuille en 4-5 ans. Un krach de 30 % ne coupe le revenu que de ~9 % la première année, et de ~16 % cumulés la deuxième. La pente est vivable, et la direction honnête.

**Le corridor.** Troisième école : on prélève w × portefeuille, mais on borne la variation annuelle du revenu. Pas plus de +5 %, pas moins de −2,5 % en réel d'une année sur l'autre. C'est exactement la règle « dynamic spending » de Vanguard, à laquelle ce livre consacre un article ([[plancher-plafond]]). Le corridor réintroduit une ruine possible, car la descente plafonnée peut ne pas suivre un effondrement. C'est un choix assumé, un point intermédiaire de la frontière.

Ces trois techniques racontent la même leçon : le pourcentage fixe brut n'est pas une règle terminale, mais une **matière première**. Lissée par moyenne, par mémoire ou par corridor, elle donne les règles du milieu de la frontière. Croisée avec l'horizon restant, elle donne la famille actuarielle ([[vpw]], [[amortissement-abw]]).

::: science Choisir w : la borne géométrique
Quel pourcentage choisir ? La théorie donne une borne claire. À long terme, le portefeuille sous pourcentage fixe croît en réel si, et seulement si, w reste inférieur au rendement réel **géométrique** espéré ([[rendements-arithmetiques-geometriques]]). Prenons w = 3-3,5 % face à un géométrique réel de ~3,5-4,5 % pour un portefeuille diversifié ([[rendements-attendus]]) : le revenu réel médian est stable ou croissant. À w = 5-6 %, il s'érode tendanciellement. Chaque année prélève alors plus que la croissance, et le revenu suit le capital vers le bas, sans jamais l'annuler. Il n'y a pas de falaise à franchir, donc w peut légitimement être plus généreux que le taux de Bengen. Un w = 4-4,5 % est défendable là où un fixe indexé exigerait 3,25-3,5 %. C'est le dividende de l'auto-correction. La pratique des dotations, autour de 4,5-5 % lissé pour des portefeuilles plus agressifs, confirme l'ordre de grandeur.
:::

## Pour qui, et comment la piloter

**Les bons profils.** Le pourcentage fixe, une fois lissé, convient à trois situations. La première : un plancher couvert par ailleurs, par une pension ou une rente, le portefeuille ne finançant que le compressible. C'est le profil de la retraite française installée. La deuxième : des budgets naturellement élastiques, à dépenses discrétionnaires dominantes, avec une vraie capacité à voyager moins les mauvaises années ([[flexibilite-realite]]). La troisième : un objectif perpétuel, transmettre un capital réel intact, dans la logique d'une dotation ([[succession-et-transmission]]). La règle est en revanche contre-indiquée en phase à découvert d'un FIRE tendu, quand un plancher élevé est financé à 100 % par le portefeuille. C'est là que sa ruine de train de vie mord exactement comme la vraie.

::: astuce Piloter la règle dans pofo
Le curseur « Spend % of portfolio (VPW) » du tiroir Spending policy applique le pourcentage fixe pur. Il écrase le besoin fixe et les règles flex, guardrails et ratchet. La frontière §06 le positionne contre les autres règles pour **votre** plan, et la vue §04 montre la distribution du revenu servi. C'est elle qui dit si la règle vous est admissible. La ruine affichée sera ~0, mais vous savez maintenant que ce zéro ne se lit pas seul ([[utiliser-la-page-fire]]). Pour approcher un Yale lissé, la règle bornée de Vanguard, la case « Bounded % of portfolio », en est le cousin direct, simulable nativement ([[plancher-plafond]]).
:::

::: exemple Le même régime hostile, brut contre lissé
Portefeuille de 1,4 M€, w = 4 %, soit 56 000 €. Scénario type 1973-1974 : −40 % réel en deux ans, puis une reprise lente. Sous pourcentage fixe brut, le revenu passe de 56 000 à 44 000 puis 34 000 € en deux ans, soit −39 %. Il remonte ensuite au rythme du marché, et reste sept ans sous 45 000 €. Sous la règle de Yale (70/30), il passe de 56 000 à 52 900 puis 48 600 €, soit −13 % en deux ans. Il touche un plancher vers 44 000 € en année 4, et sa remontée s'amorce avant même le retour du marché. Même portefeuille, même « garantie » anti-ruine, mais la version lissée transforme une chute libre en pente douce. Dans un budget avec 25 % de compressible, la première trajectoire est une crise de plan, la seconde une gestion courante. Le lissage n'est pas un raffinement. Il est la condition d'admissibilité de toute la famille proportionnelle.
:::

## L'essentiel à retenir

- Retrait = w × portefeuille courant. La ruine du **capital** est mathématiquement impossible, celle du **train de vie** ne l'est pas. Jugez cette règle sur la distribution du revenu servi, jamais sur son taux de succès de 100 %, parfaitement creux.
- Son revenu a la volatilité du portefeuille. La question d'admissibilité est simple : « puis-je vivre dix ans à 55-65 % du confort dans un régime hostile ? » Oui si le plancher est couvert par ailleurs, non en phase à découvert tendue.
- Ses deux vertus profondes, héritées par toutes les règles modernes : la contracyclicité, qui prélève peu au creux, et l'auto-correction face aux erreurs d'hypothèses.
- Le lissage la rend vivable : moyenne mobile, règle de Yale (70 % de mémoire + 30 % de cible, le standard des dotations), ou corridor borné (la version Vanguard, [[plancher-plafond]]). Et w peut être plus généreux que le taux de Bengen, la borne étant le rendement géométrique réel espéré.
- À l'état pur, c'est une matière première plus qu'une stratégie. Ses descendantes domestiquées ([[vpw]], [[amortissement-abw]], [[plancher-plafond]]) occupent le milieu de la frontière.

---

## Pour aller plus loin

- James Tobin et la règle de dépense des dotations (« spending rule » de Yale) : les rapports annuels du Yale Endowment en décrivent la version en vigueur.
- Early Retirement Now, volet 10 (le pourcentage fixe face à Guyton-Klinger) et volet 11 (les critères) ([[serie-ern]]).
- Bogleheads wiki, « Variable percentage withdrawal » : la généalogie proportionnelle → actuarielle ([[vpw]]).
- Dans ce livre : [[plancher-plafond]] (le corridor industrialisé), [[amortissement-abw]] (le pourcentage rendu conscient de l'horizon), et [[flexibilite-realite]] (ce que « vivre la variabilité » veut dire).
