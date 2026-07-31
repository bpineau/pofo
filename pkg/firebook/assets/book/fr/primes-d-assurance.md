# Les primes d'assurance : cat bonds et arbitrage de fusions

Les diversifiants visités jusqu'ici prennent tous le contrepied d'un marché. Le trend suit les tendances que d'autres subissent ([[managed-futures]]), le long volatility achète la convexité que d'autres vendent ([[long-volatility]]), le macro discrétionnaire parie contre le consensus ([[global-macro]]). Il existe une autre famille, moins racontée, où votre contrepartie n'est pas un spéculateur mais quelqu'un qui a un vrai besoin de céder un risque. Un assureur floridien doit se débarrasser du gros ouragan. L'actionnaire d'une société rachetée veut son argent maintenant plutôt qu'un chèque dans dix-huit mois, si la fusion aboutit. Dans les deux cas, quelqu'un paie une prime pour dormir tranquille, et cette prime est un rendement.

Deux marchés rendent cette famille accessible à un investisseur européen. Les **cat bonds** (obligations catastrophe), qui titrisent le risque de catastrophe naturelle. Et l'**arbitrage de fusions** (merger arbitrage), qui achète les sociétés déjà rachetées et encaisse l'écart de prix. Cet article explique les deux mécaniques, chiffre ce qu'elles ont vraiment rapporté à un investisseur en euros, mesure ce qu'elles font à un portefeuille de retrait, et finit par la seule question qui compte, la dose.

::: cle Vous passez de l'autre côté du guichet
Le rendement de ces stratégies est une prime d'assurance, et leur perte est un sinistre. Le sinistre a une cause physique ou juridique, un ouragan ou un veto de la concurrence, pas une cause économique. La décorrélation avec les actions n'est donc pas une statistique observée sur une fenêtre, elle est causale. Attention au corollaire, qui déçoit toujours. Un actif décorrélé ne protège de rien. Il ne monte pas quand les actions tombent, il fait sa vie ailleurs, ce qui est utile mais n'est pas une couverture ([[actifs-defensifs]]).
:::

## Le cat bond, un contrat de réassurance découpé en titres

Un assureur, ou plus souvent un réassureur, veut se couvrir contre un événement rare et énorme. Plutôt que d'acheter cette protection à un confrère, il émet un titre. Les investisseurs versent le nominal, qui est immédiatement placé en instruments monétaires et bloqué en garantie. Ils reçoivent chaque trimestre le taux court de la devise du collatéral, plus un écart, le **spread**, qui est le prix de la protection. Si l'événement défini au contrat se produit, la protection joue. Le nominal est amputé, en partie ou en totalité, et sert à payer les sinistres. Sinon, l'investisseur récupère son capital à l'échéance, en général trois ans.

Le contrat définit un **déclencheur** (trigger). Il peut être indemnitaire, indexé sur les pertes réelles du sponsor, ou paramétrique, indexé sur une grandeur physique comme la vitesse du vent ou la magnitude d'un séisme, ou encore indexé sur les pertes de la profession entière. Le marché s'est spécialisé sur quelques périls bien modélisés, l'ouragan américain d'abord, puis le séisme californien et japonais, la tempête européenne, l'inondation et l'incendie.

Le marché existe depuis 1997 et il a changé d'échelle récemment. L'encours a terminé 2025 à un peu plus de 61 milliards de dollars, en hausse de 24 % en un an. Ce n'est plus une curiosité, c'est une classe d'actifs cotée, avec des indices, des gérants spécialisés et désormais un ETF UCITS.

**L'historique, tel qu'il est publié.** L'indice de référence, le Swiss Re Global Cat Bond, affiche sur vingt-trois ans un rendement moyen d'environ 6,7 % par an, avec une seule année négative, 2022, celle de l'ouragan Ian. Les trois derniers exercices donnent le vertige, avec 19,7 % en 2023, 17,3 % en 2024 et 11,4 % en 2025. Ces chiffres sont exacts et ils sont trompeurs, pour deux raisons que la suite de l'article détaille. Ils sont en dollars, donc gonflés par un taux monétaire américain remonté à 5 %. Et ils suivent la grande repentification des spreads d'après Ian, quand les assureurs ont payé cher pour se couvrir.

::: figure sinistres-calendrier
Douze années de rendements mensuels réels en euros, cat bonds en haut, actions mondiales en bas, à la même échelle. Le panneau du haut a deux accidents, septembre 2017 (Irma et Maria) et septembre 2022 (Ian, plus le choc d'inflation qui frappe un actif au rendement fixe). Septembre 2017 fut même un bon mois boursier. Sur 146 mois, deux seulement voient les deux panneaux perdre ensemble. Mars 2020, marqué par le trait rouge, où les vendeurs forcés font tomber ce qui n'a rien à voir. Et septembre 2022, par pure coïncidence, Ian d'un côté et le choc de taux de l'autre.
:::

## Ce que la prime coûte vraiment, quand on vit en euros

C'est le passage que les brochures oublient, et il change le verdict.

::: attention Le piège du collatéral pour l'investisseur en euros
Le nominal d'un cat bond dort en titres monétaires américains. Le rendement total est donc « taux court du dollar plus spread ». Achetez une part couverte en euro, et la couverture de change fait exactement son travail, elle échange le taux court du dollar contre celui de l'euro. Vous gardez le spread, vous perdez l'écart de taux. Entre 2015 et 2021, le taux court de l'euro était négatif, et cette famille a donc rapporté, en euros, à peu près le spread moins les frais. Mesuré sur le fonds couvert en euro le plus ancien du marché, déflaté de l'inflation française, le résultat réel est de 1,7 % par an de novembre 2013 à décembre 2025, pour 4,3 % de volatilité, un pire mois à −9,4 % et une pire baisse réelle de 18 %. Le même actif, raconté en dollars nominaux, affiche 6,7 % par an. Les deux chiffres sont vrais. Un seul est le vôtre.
:::

Trois autres coûts méritent d'être posés à côté du premier. Les **frais**, d'abord, qui tournent autour de 1,3 % par an sur les véhicules accessibles, ce qui absorbe un quart à un tiers d'un spread normal. Le **risque de modèle**, ensuite, car le prix d'un cat bond repose sur une probabilité d'événement calculée par trois cabinets spécialisés, avec un climat qui bouge sous le modèle. Et la **liquidité**, enfin, qui est bonne en temps normal et disparaît en mars 2020, exactement comme celle de tout ce qui se négocie de gré à gré.

Il reste que la structure est saine, et qu'elle est même l'une des rares primes du menu alternatif dont l'origine se raconte sans économétrie. Quelqu'un doit céder ce risque, il n'y a pas assez de capital pour le porter, donc il paie. La question n'est jamais « est-ce que la prime existe », mais « à quel prix aujourd'hui, et après quels frais ».

## L'arbitrage de fusions, ou le rendement d'un calendrier juridique

Le mécanisme tient en trois lignes. Une société annonce le rachat d'une autre à 50 € l'action. Le lendemain, la cible ne cote pas 50 €, mais 48 €. Cet écart de 2 € est le prix de deux incertitudes, le délai jusqu'à la réalisation et la possibilité que l'opération échoue. L'arbitragiste achète la cible, vend l'acquéreur ou un indice pour neutraliser le marché, et encaisse l'écart quand l'opération se conclut.

Le profil de rendement est celui d'une vente d'assurance, encore une fois. De petits gains réguliers, calés sur un calendrier juridique et non sur un cycle économique, et de rares pertes brutales quand une opération casse. Deux propriétés en découlent, et elles intéressent un rentier. La prime est cotée **au-dessus du taux court**, donc elle monte avec les taux, ce qui en fait un actif de trésorerie amélioré plutôt qu'un substitut d'actions. Et la sensibilité au marché est faible par construction, puisque l'exposition longue est couverte.

Le prix à payer se lit dans les échecs. Les ruptures d'opérations ne sont pas indépendantes entre elles. Elles arrivent par vagues, quand le crédit se ferme et que les financements sautent, ou quand une autorité de la concurrence change de doctrine, comme l'ont vécu les arbitragistes américains entre 2021 et 2023. La corrélation aux actions, faible en moyenne, remonte donc précisément dans les épisodes qui vous inquiètent.

**Le verdict pratique est un verdict de frais.** Un fonds UCITS d'arbitrage de fusions facture couramment 1,8 % de gestion, plus une commission de surperformance de 20 %. Sur une prime brute de 3 à 4 points au-dessus du monétaire, il ne reste pas grand-chose, et ce qui reste passe encore au prélèvement forfaitaire ([[flat-tax-et-imposition]]). L'ETF américain qui suit un indice d'opérations annoncées coûte cinq fois moins cher, mais il achète l'univers en bloc, y compris les écarts larges qui le sont pour de bonnes raisons, et il n'est de toute façon pas accessible au particulier européen ([[etf-ucits-europeens]]).

## Ce que ces briques font vraiment à un portefeuille de retrait

Voici la partie que ce livre doit à ses lecteurs, parce qu'elle manque partout ailleurs. Une brique décorrélée n'est pas bonne en soi. Elle est bonne, ou non, **à la place de quelque chose**, et c'est le financement qui décide.

L'exercice suivant part d'un portefeuille de retrait déjà diversifié, du type de ceux que décrit [[portefeuilles-tous-temps]], avec des actions, de la duration, du trend, de l'or et des obligations indexées. On y ajoute dix points d'une des deux briques, sur la même fenêtre pour toutes les variantes, et on regarde le taux de retrait soutenable à 5 % de ruine sur quarante ans.

| Variante | Rendement réel | Volatilité | Taux soutenable |
|---|---|---|---|
| Portefeuille de départ | 6,4 % | 7,2 % | 4,94 % |
| +10 points de cat bonds, financés au prorata de tout | 5,9 % | 6,6 % | 4,70 % |
| +10 points d'arbitrage, financés au prorata de tout | 6,0 % | 6,8 % | 4,74 % |
| +10 points de cat bonds, pris sur la poche obligataire | 6,6 % | 6,9 % | 5,11 % |
| +10 points d'arbitrage, pris sur la poche obligataire | 6,7 % | 7,1 % | 5,16 % |

La lecture est nette. Financer la brique en vendant un peu de tout, actions comprises, **dégrade** le plan, parce qu'on remplace du rendement par du portage. Financer la même brique sur la poche obligataire l'améliore, et le gain vaut environ deux dixièmes de point de taux de retrait. Une brique de trésorerie améliorée se compare à de la trésorerie, jamais à des actions.

Trois honnêtetés doivent accompagner ce tableau. La fenêtre commune s'arrête à douze ans, faute d'historique plus profond sur les cat bonds, et elle ne contient donc ni 2008 ni les années 1970. Cette fenêtre est aussi la pire du siècle pour les obligations longues, ce qui flatte mécaniquement toute idée qui consiste à les remplacer. Et en retirant deux points de rendement annuel à la brique, pour simuler un régime de spreads moins généreux, l'avantage retombe à zéro sans devenir négatif.

Le résultat le plus utile est ailleurs, dans ce que le ménage vit. En passant la même comparaison sous une règle de dépense flexible, le quartile bas des dépenses servies bouge de moins de 0,4 %. Autrement dit, ces briques ne changent pas la vie du rentier. Elles changent la texture du portefeuille, réduisent un peu la volatilité et déplacent la nature du risque. C'est un ajustement d'ingénieur, pas une transformation ([[flexibilite-realite]] dit la même chose des règles de retrait).

::: exemple Cinq points de cat bonds, pris là où il faut
Un ménage détient 13 points d'obligations indexées et 4 points de dette souveraine longue. Il déplace 5 points vers un fonds de cat bonds couvert en euro. Ce qu'il achète, c'est un rendement attendu supérieur au monétaire de 3 à 5 points quand les spreads sont larges, sans risque de taux, avec un risque d'ouragan à la place. Ce qu'il perd, c'est la convexité de la duration longue le jour d'un krach déflationniste, puisque les cat bonds ne montent pas dans les krachs, ils les ignorent. Ce que ça coûte, 1,3 % de frais sur la ligne, plus un compte-titres ordinaire, ces fonds n'existant ni en PEA ni en assurance-vie grand public. Ce qu'il faut écrire avant d'acheter, la phrase qui tiendra le jour de la mauvaise saison, « je sais que cette ligne peut perdre 10 % en un mois sans qu'aucun marché ne baisse, et je ne la vendrai pas pour autant ».
:::

## La dose, et les conditions d'entrée

Le cadre de décision se résume à quatre conditions, à vérifier dans l'ordre.

- **Votre défense de base est déjà en place.** Ces primes ne remplacent ni la duration, ni le trend, ni l'or, qui répondent chacun à un régime précis ([[actifs-defensifs]]). Elles s'ajoutent après, ou pas du tout.
- **Le financement vient de la poche courte.** Trésorerie, obligataire court, fonds euros, éventuellement une part des obligations indexées. Jamais les actions, jamais le trend, sinon le calcul se retourne.
- **Le prix du risque est bon aujourd'hui.** Pour les cat bonds, cela se lit sur le spread offert et sur le taux court de votre devise, tous deux publics. Pour l'arbitrage, sur l'écart moyen des opérations en cours. Une prime d'assurance vendue trop bon marché reste une mauvaise affaire, même décorrélée.
- **Le véhicule est propre.** Frais totaux sous 1,5 %, encours suffisant pour ne pas craindre une fermeture, part couverte en euro si vous ne voulez pas que le change fasse les trois quarts de la variance. L'ETF cat bond européen lancé fin 2025 est une nouveauté utile, mais il cote en dollar et pèse encore peu.

La dose raisonnable est de 0 à 10 points au total, pris sur la poche courte. Zéro est une réponse parfaitement défendable pour un patrimoine simple, exactement comme pour le long volatility. Dix points sont un maximum, parce qu'un sinistre de saison peut coûter 10 à 20 % à la ligne, et qu'il faut pouvoir traverser cela sans toucher au plan ([[psychologie-du-retrait]]).

## L'essentiel à retenir

- Les cat bonds et l'arbitrage de fusions vendent une assurance à quelqu'un qui doit céder un risque. Leur décorrélation aux actions est causale, l'ouragan et le veto antitrust n'étant pas des variables économiques, mais un actif décorrélé ne protège de rien.
- Le marché des cat bonds pèse 61 milliards de dollars et son indice affiche 6,7 % par an sur vingt-trois ans, avec une seule année négative. Ce chiffre est en dollars ; pour un investisseur en euros couvert, le rendement réel mesuré de 2013 à 2025 tombe à 1,7 % par an, l'écart venant du taux court de l'euro et des frais.
- L'arbitrage de fusions est un actif de trésorerie amélioré, dont la prime monte avec les taux et dont le risque est réglementaire et groupé. Les frais des véhicules UCITS en consomment souvent la moitié.
- Dans un portefeuille de retrait déjà diversifié, le financement décide de tout. Pris sur la poche obligataire, dix points ajoutent environ 0,2 point de taux de retrait soutenable ; pris au prorata de tout, ils en retirent autant.
- L'effet sur le niveau de vie est marginal, moins de 0,4 % sur le quartile bas des dépenses. Dose de 0 à 10 points, financée par la poche courte, avec un véhicule à moins de 1,5 % de frais et une part couverte en euro.

---

## Pour aller plus loin

- Artemis (artemis.bm) : la chronique quotidienne du marché des cat bonds, avec l'encours, les émissions et les pertes, en accès libre.
- Swiss Re, Global Cat Bond Index : la série de référence depuis 2002, et sa méthodologie.
- Morningstar, « Catastrophe Bonds as Portfolio Diversifiers » : la revue équilibrée, avantages et limites, pour un lecteur qui découvre.
- Dans ce livre : [[actifs-defensifs]] (le cahier des charges qu'une brique doit remplir), [[global-macro]] (le catalogue des autres primes alternatives), [[cash-ameliore]] (l'autre façon de faire travailler la poche courte), [[concevoir-un-portefeuille]] (où ces points se prennent).
