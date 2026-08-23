# Les primes d'assurance : cat bonds et arbitrage de fusions

Les diversifiants passés en revue jusqu'ici gagnent, quand ils gagnent, contre un autre acteur des marchés. Le trend suit des tendances que d'autres subissent ([[managed-futures]]), le long volatility achète la convexité que d'autres vendent ([[long-volatility]]), le macro discrétionnaire parie contre le consensus ([[global-macro]]). Il reste une famille dont on parle moins, où la contrepartie n'est pas un spéculateur mais un professionnel qui cherche à se défaire d'un risque trop lourd pour lui. Un assureur de Floride ne peut pas porter seul la facture d'un ouragan majeur. L'actionnaire d'une société en cours de rachat préfère un prix certain aujourd'hui à un chèque dans dix-huit mois, conditionné à la réussite de l'opération. Dans les deux cas, quelqu'un paie une prime pour dormir tranquille, et cette prime est un rendement.

Deux marchés rendent cette famille accessible à l'investisseur européen : les **cat bonds** (obligations catastrophe), qui titrisent le risque de catastrophe naturelle, et l'**arbitrage de fusions** (merger arbitrage), qui achète les sociétés en cours de rachat pour encaisser l'écart entre leur cours et le prix promis. Cet article démonte les deux mécaniques, chiffre ce qu'elles ont réellement rapporté à un investisseur en euros, mesure leur effet sur un portefeuille de retrait, et termine par la seule question qui compte, celle de la dose.

::: cle Vous passez de l'autre côté du guichet
Le rendement de ces stratégies est une prime d'assurance ; leur perte est un sinistre. Or le sinistre a une cause physique ou juridique, un ouragan, le veto d'une autorité de la concurrence, pas une cause économique. La décorrélation avec les actions ne sort donc pas d'une fenêtre statistique qui pourrait se refermer : elle tient à la nature même du risque. Gare au corollaire, qui déçoit toujours. Un actif décorrélé ne protège de rien. Il ne monte pas quand les actions tombent, il vit sa vie ailleurs, ce qui est utile mais n'est pas une couverture ([[actifs-defensifs]]).
:::

## Le cat bond, un contrat de réassurance découpé en titres

Un assureur, ou plus souvent un réassureur, cherche à se couvrir contre un événement rare mais ruineux, l'ouragan qui traverse Miami, le séisme sous Tokyo. Plutôt que d'acheter cette protection à un confrère, il émet un titre. Les investisseurs versent le nominal, aussitôt placé en titres monétaires et bloqué en garantie. En échange, ils touchent chaque trimestre le taux court de la devise du collatéral, augmenté d'un écart, le **spread**, qui est le prix de la protection. Si la catastrophe définie au contrat se produit, la garantie joue : le nominal est amputé, en partie ou en totalité, et sert à payer les sinistres. Sinon, l'investisseur récupère son capital à l'échéance, au bout de trois ans en général. La perte, elle, ne se répare pas : une obligation amputée l'est pour de bon, et ce qui « se refait » après une mauvaise saison, ce ne sont pas les titres brûlés, ce sont les spreads élargis des émissions suivantes.

Le contrat précise un **déclencheur** (trigger). Il peut être indemnitaire (indexé sur les pertes réelles de l'assureur émetteur), paramétrique (indexé sur une grandeur physique, vitesse du vent, magnitude d'un séisme) ou sectoriel (indexé sur les pertes de toute la profession). Le marché s'est concentré sur quelques périls bien modélisés : l'ouragan américain d'abord, puis les séismes californien et japonais, la tempête européenne, l'inondation et l'incendie.

Le marché existe depuis 1997, et il vient de changer d'échelle. L'encours a fini 2025 à un peu plus de 61 milliards de dollars, en hausse de 24 % sur un an. Ce n'est plus une curiosité : c'est une classe d'actifs cotée, avec ses indices, ses gérants spécialisés et, depuis peu, un ETF UCITS.

**L'historique, tel qu'il se publie.** L'indice de référence, le Swiss Re Global Cat Bond, affiche sur vingt-trois ans un rendement moyen d'environ 6,7 % par an, avec une seule année négative, 2022, celle de l'ouragan Ian. Les trois derniers exercices donnent le vertige : 19,7 % en 2023, 17,3 % en 2024, 11,4 % en 2025. Ces chiffres sont exacts, et ils sont trompeurs, pour deux raisons que la suite détaille. Ils sont exprimés en dollars, donc gonflés par un taux monétaire américain remonté à 5 %. Et ils datent d'après Ian, quand les spreads se sont envolés et que les assureurs ont payé leur couverture au prix fort. Rien ne promet ce régime-là au prochain acheteur.

::: figure sinistres-calendrier
Douze années de rendements mensuels réels en euros, cat bonds en haut, actions mondiales en bas, à la même échelle. Le panneau du haut compte deux accidents, septembre 2017 (Irma et Maria) et septembre 2022 (Ian, que l'inflation de 2022 aggrave dans une série en euros constants). Septembre 2017 fut pourtant un bon mois boursier. Sur 146 mois, deux seulement voient les deux panneaux perdre ensemble : mars 2020, marqué du trait rouge, quand les vendeurs forcés font tout tomber en même temps, et septembre 2022, par pure coïncidence, Ian d'un côté, le choc de taux de l'autre.
:::

## Ce que la prime rapporte vraiment, quand on vit en euros

C'est le passage que les plaquettes commerciales passent sous silence, et il change le verdict.

::: attention Le piège du collatéral pour l'investisseur en euros
Le nominal d'un cat bond dort en titres monétaires américains. Le rendement total vaut donc « taux court du dollar plus spread ». Achetez une part couverte en euro, et la couverture de change fait exactement son office : elle troque le taux court du dollar contre celui de l'euro. Vous gardez le spread, vous perdez l'écart de taux. De 2015 à 2021, le taux court de l'euro était négatif, si bien que la famille a rapporté en euros à peu près le spread, moins les frais. Mesuré sur le fonds couvert en euro le plus ancien du marché, une fois l'inflation française retirée, le résultat est de 1,7 % par an de novembre 2013 à décembre 2025, pour 4,3 % de volatilité, un pire mois à −9,4 % et une pire baisse réelle de 18 %. Le même actif, compté en dollars courants, affiche 6,7 % par an. Les deux chiffres sont vrais. Un seul est le vôtre.
:::

Trois autres coûts complètent la facture. Les **frais** d'abord : environ 1,3 % par an sur les véhicules accessibles, soit le quart ou le tiers d'un spread normal. Sur quinze ans, un fonds défensif facturant 1,75 % a servi environ 1,9 % par an : dans cette classe, le choix de la part n'est pas un détail, c'est la moitié de la décision. Le **risque de modèle** ensuite : le prix d'un cat bond repose sur des probabilités d'événement calculées par trois cabinets spécialisés, et le climat évolue plus vite que leurs modèles. La **liquidité** enfin, correcte en temps normal, évaporée en mars 2020, comme celle de tout ce qui se négocie de gré à gré. L'histoire ajoute un avertissement de plomberie : en 2008, des fonds cat bonds ont baissé avec tout le reste, sans le moindre ouragan, parce que le collatéral de certains titres reposait sur des montages signés Lehman Brothers. Les structures l'ont corrigé depuis, le nominal dort désormais en titres d'État ou assimilés, mais l'épisode rappelle que la décorrélation causale ne protège pas de la tuyauterie financière.

Reste que la prime, elle, est solidement fondée. C'est même l'une des rares du menu alternatif dont l'origine s'explique sans économétrie : quelqu'un doit céder ce risque, le capital manque pour le porter, donc il paie. La question n'est jamais « la prime existe-t-elle », mais « que vaut-elle aujourd'hui, et que reste-t-il après frais ».

## L'arbitrage de fusions, ou le rendement d'un calendrier juridique

Le mécanisme tient en trois phrases. Une société annonce le rachat d'une autre à 50 € l'action ; dès le lendemain, la cible cote 48 €, pas 50. L'écart de 2 € rémunère deux incertitudes, le délai d'ici la réalisation et le risque que l'opération échoue. L'arbitragiste achète la cible, vend l'acquéreur ou un indice pour neutraliser le marché, et empoche l'écart quand l'opération se conclut.

Le profil de rendement est, là encore, celui d'un vendeur d'assurance : de petits gains réguliers, rythmés par un calendrier juridique et non par le cycle économique, et de rares pertes brutales quand une opération échoue. Deux propriétés en découlent, et elles intéressent le rentier. L'écart se cote au-dessus du taux court, donc la prime monte avec les taux : c'est de la trésorerie améliorée, pas un substitut d'actions. Et la sensibilité au marché est faible par construction, l'exposition longue étant couverte.

Le revers se découvre quand les opérations échouent. Les échecs n'arrivent pas au hasard, ils arrivent par vagues : quand le crédit se ferme et que les financements sautent, ou quand une autorité de la concurrence durcit sa doctrine, comme les arbitragistes américains l'ont appris à leurs dépens entre 2021 et 2023. La corrélation aux actions, faible en moyenne, remonte donc précisément dans les épisodes qui vous inquiètent.

**En pratique, tout se joue sur les frais.** Un fonds UCITS d'arbitrage de fusions facture couramment 1,8 % de gestion, plus une commission de surperformance de 20 %. Sur une prime brute de 3 à 4 points au-dessus du monétaire, il ne reste pas grand-chose, et ce reste passe encore par le prélèvement forfaitaire ([[flat-tax-et-imposition]]). L'ETF américain qui suit un indice d'opérations annoncées coûte cinq fois moins cher, mais il achète toutes les opérations sans les trier, y compris celles dont l'écart est large parce que le marché doute à juste titre, et il est de toute façon inaccessible au particulier européen ([[etf-ucits-europeens]]).

## Ce que ces briques font vraiment à un portefeuille de retrait

Cette partie manque dans presque tout ce qui se publie sur le sujet, et c'est pourtant elle qui décide. Une brique décorrélée n'est jamais bonne en soi. Elle est bonne, ou non, **à la place de quelque chose**, et le choix de ce quelque chose emporte le verdict.

L'exercice qui suit part d'un portefeuille de retrait déjà diversifié, du type de ceux de [[portefeuilles-tous-temps]] : des actions, de la duration, du trend, de l'or, des obligations indexées. On y ajoute dix points de l'une des deux briques, sur une fenêtre commune à toutes les variantes, et on lit le taux de retrait soutenable à 5 % de ruine sur quarante ans.

| Variante | Rendement réel | Volatilité | Taux soutenable |
|---|---|---|---|
| Portefeuille de départ | 6,4 % | 7,2 % | 4,94 % |
| +10 points de cat bonds, financés au prorata de toutes les lignes | 5,9 % | 6,6 % | 4,70 % |
| +10 points d'arbitrage, financés au prorata de toutes les lignes | 6,0 % | 6,8 % | 4,74 % |
| +10 points de cat bonds, pris sur la poche obligataire | 6,6 % | 6,9 % | 5,11 % |
| +10 points d'arbitrage, pris sur la poche obligataire | 6,7 % | 7,1 % | 5,16 % |

La lecture est nette. Financer la brique en rognant toutes les lignes, actions comprises, **dégrade** le plan : on remplace du rendement par du portage. Financer la même brique sur la poche obligataire l'améliore, d'environ deux dixièmes de point de taux de retrait. La règle générale se lit dans ce contraste : une trésorerie améliorée se compare à de la trésorerie, jamais à des actions.

Trois réserves accompagnent ce tableau. La fenêtre commune s'arrête à douze ans, faute d'historique plus profond côté cat bonds : ni 2008 ni les années 1970 n'y figurent. Et aucun historique public de fonds ne commence avant 2006 : Katrina, la pire saison moderne, reste hors échantillon de tous les backtests de la classe, un biais optimiste structurel. Cette fenêtre est aussi la pire du siècle pour les obligations longues, ce qui flatte mécaniquement toute idée consistant à les remplacer. Enfin, si l'on retire deux points de rendement annuel à la brique, pour simuler un régime de spreads moins généreux, l'avantage retombe à zéro, sans toutefois devenir négatif.

Le résultat le plus utile est ailleurs, dans ce que le ménage vit. Rejouée sous une règle de dépense flexible, la même comparaison déplace le quartile bas des dépenses servies de moins de 0,4 %. Autrement dit, ces briques ne changent pas la vie du rentier. Elles déplacent la nature du risque et liment un peu la volatilité ; c'est un réglage d'ingénieur, pas une transformation ([[flexibilite-realite]] dit la même chose des règles de retrait).

::: exemple Cinq points de cat bonds, pris au bon endroit
Un ménage détient 13 points d'obligations indexées et 4 points de dette souveraine longue. Il déplace 5 points vers un fonds de cat bonds couvert en euro. Ce qu'il achète : un rendement attendu de 3 à 5 points au-dessus du monétaire quand les spreads sont larges, sans risque de taux, avec un risque d'ouragan à la place. Ce qu'il abandonne : la convexité de la duration longue le jour d'un krach déflationniste, car les cat bonds ne montent pas dans les krachs, ils les ignorent. Ce que cela coûte : 1,3 % de frais sur la ligne, et un compte-titres ordinaire, ces fonds n'existant ni en PEA ni en assurance-vie grand public. Et la phrase à écrire avant d'acheter, pour tenir le jour de la mauvaise saison : « cette ligne peut perdre 10 % en un mois sans qu'aucun marché ne baisse, et je ne la vendrai pas pour autant ».
:::

## La dose, et les conditions d'entrée

Quatre conditions, à vérifier dans l'ordre.

- **Votre défense de base est déjà en place.** Ces primes ne remplacent ni la duration, ni le trend, ni l'or, qui répondent chacun à un régime précis ([[actifs-defensifs]]). Elles s'ajoutent après, ou pas du tout.
- **Le financement vient de la poche courte.** Trésorerie, obligataire court, fonds euros, éventuellement une part des obligations indexées. Jamais les actions, jamais le trend, sinon le calcul se retourne.
- **Le prix du risque est bon aujourd'hui.** Pour les cat bonds, cela se lit sur le spread offert et sur le taux court de votre devise, tous deux publics. Le cycle joue pour l'acheteur patient : les spreads s'élargissent après les mauvaises saisons, jamais avant, et les meilleurs millésimes ont suivi 2017 et 2022. Pour l'arbitrage, sur l'écart moyen des opérations en cours. Vendre de l'assurance au rabais reste une mauvaise affaire, même décorrélée.
- **Le véhicule est propre.** Des frais totaux sous 1,5 %, un encours qui éloigne le risque de fermeture, une part couverte en euro si vous ne voulez pas que le change fasse les trois quarts de la variance. L'ETF cat bond européen lancé fin 2025 est une piste à suivre, mais il cote en dollar, sans couverture de change ni historique. Les parts profondes et couvertes en euro s'achètent à la valeur liquidative via des plateformes de fonds, rarement dans la liste d'un courtier ordinaire : dans cette famille, l'accès est une contrainte à vérifier avant le reste.

La dose raisonnable va de 0 à 10 points, pris sur la poche courte. Zéro est une réponse parfaitement défendable pour un patrimoine simple, exactement comme pour le long volatility. Dix points sont un plafond : une mauvaise saison de catastrophes peut coûter 10 à 20 % à la ligne, et il faut pouvoir traverser cela sans toucher au plan ([[psychologie-du-retrait]]). La concentration du marché justifie aussi ce plafond : environ deux tiers du risque total tiennent en deux périls, l'ouragan et le séisme américains, et un seul atterrissage majeur sur Miami, scénario parfaitement plausible, peut emporter de l'ordre du cinquième d'un fonds entier.

## L'essentiel à retenir

- Les cat bonds et l'arbitrage de fusions se font payer pour porter le risque d'un autre, l'ouragan ou la fusion incertaine. Leur décorrélation aux actions est causale, pas statistique ; mais un actif décorrélé ne protège de rien.
- Le marché des cat bonds pèse 61 milliards de dollars et son indice affiche 6,7 % par an sur vingt-trois ans, avec une seule année négative. Ce chiffre est en dollars ; pour un investisseur en euros couvert, le rendement réel mesuré de 2013 à 2025 tombe à 1,7 % par an, l'écart venant du taux court de l'euro et des frais.
- L'arbitrage de fusions est de la trésorerie améliorée : la prime monte avec les taux, et le risque, réglementaire, arrive par vagues. Les frais des véhicules UCITS en absorbent souvent la moitié.
- Dans un portefeuille de retrait déjà diversifié, le financement décide de tout. Pris sur la poche obligataire, dix points ajoutent environ 0,2 point de taux de retrait soutenable ; pris au prorata de toutes les lignes, ils en retirent autant.
- L'effet sur le niveau de vie est marginal, moins de 0,4 % sur le quartile bas des dépenses. Dose de 0 à 10 points, financée par la poche courte, avec un véhicule à moins de 1,5 % de frais et une part couverte en euro.

---

## Pour aller plus loin

- Artemis (artemis.bm) : la chronique quotidienne du marché des cat bonds, avec l'encours, les émissions et les pertes, en accès libre.
- Swiss Re, Global Cat Bond Index : la série de référence depuis 2002, et sa méthodologie.
- Morningstar, « Catastrophe Bonds as Portfolio Diversifiers » : la revue équilibrée, avantages et limites, pour un lecteur qui découvre.
- Dans ce livre : [[actifs-defensifs]] (le cahier des charges d'une brique défensive), [[global-macro]] (le catalogue des autres primes alternatives), [[cash-ameliore]] (l'autre façon de faire travailler la poche courte), [[concevoir-un-portefeuille]] (où ces points se prennent).
