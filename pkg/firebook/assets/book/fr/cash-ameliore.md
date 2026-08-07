# Le cash amélioré : monétaire, CLO AAA, fonds euros

L'article sur le matelas ([[cash-buffer]]) répond à la question de la taille : combien d'années de dépenses garder en liquidités, et à quel prix de performance. Il ne dit pas **avec quoi** remplir cette poche, et c'est pourtant une décision à part entière. Un rentier qui garde trois ans de dépenses de côté immobilise souvent 60 000 à 120 000 €, pendant trente ans. Un point de rendement de plus sur cette somme, c'est un voyage par an, indéfiniment.

L'ennui, c'est que chaque cran de rendement supplémentaire s'achète en risque, et que ce risque a la mauvaise habitude de se réveiller au moment précis où la poche courte doit servir. Cet article classe les instruments du plus inerte au plus rémunérateur, mesure l'écart réel entre eux, et démonte l'étage le moins connu du public français, les **CLO AAA**, arrivés en ETF européen en 2024.

::: cle Les trois questions d'une poche courte
Une ligne de trésorerie n'est bonne que si elle répond oui aux trois.

- Puis-je la vendre le pire jour de la décennie ?
- Vais-je la vendre à peu près à sa valeur ce jour-là ?
- Ce qu'elle rapporte en plus du monétaire, après frais et après impôt, vaut-il la réponse aux deux premières ?

La troisième question élimine plus de produits que les deux autres réunies.
:::

## Les étages de l'échelle

**Les livrets réglementés** sont le seul instrument sans aucun risque de marché. Le capital est garanti, la liquidité est immédiate, l'intérêt est exonéré. Leur taux est administré, révisé deux fois par an, et leurs plafonds sont vite atteints. C'est l'étage zéro, celui des six premiers mois de dépenses, et il n'a pas de concurrent sérieux pour ce rôle.

**Le fonds euros** de l'assurance-vie offre une garantie en capital contractuelle et un rendement publié chaque année. Il a deux qualités rares pour un rentier : une valeur qui ne baisse jamais, et une fiscalité d'enveloppe plus douce que celle du compte-titres ([[flat-tax-et-imposition]], [[enveloppes-francaises]]). Il a deux défauts, un rendement qui suit les taux avec deux ans de retard, et une liquidité contractuelle plutôt qu'instantanée, l'assureur disposant légalement d'un délai de règlement.

**L'ETF monétaire** réplique le taux au jour le jour de la zone euro, l'ESTR, pour environ 0,10 % de frais. Sur un compte-titres, c'est l'instrument le plus proche du cash pur : 0,25 % de volatilité, une pire baisse qui se compte en points de base. C'est l'étalon auquel tout le reste doit se comparer.

**L'obligataire ultra-court** y ajoute du crédit d'entreprise à très courte échéance. Le supplément mesuré est réel mais modeste, environ 27 points de base par an, pour une pire baisse d'une vingtaine de points de base. Un étage honnête, sans surprise.

**Le CLO AAA** occupe le dernier étage, le plus rémunérateur : environ 130 points de base au-dessus du monétaire, contre un risque de prix les mauvaises années. C'est aussi le moins connu du lot, et le plus mal-aimé à cause de son sigle, si bien qu'il mérite d'être démonté pièce par pièce. C'est l'objet de la section suivante.

**Les fonds à échéance**, ou fonds datés, forment enfin un étage à part, ni tout à fait de la trésorerie ni tout à fait de l'obligataire classique. Ils ont leur section plus bas.

## Le CLO, démonté pièce par pièce

Le sigle fait peur (il sonne comme CDO, ces titrisations de crédits immobiliers au cœur de la crise de 2008), et cette peur est le principal obstacle. Elle se soigne en démontant l'objet, qui est plus simple qu'il n'en a l'air.

**Le carburant : des prêts d'entreprise.** Un CLO (collateralised loan obligation, obligation adossée à des prêts) est un véhicule qui achète un portefeuille de 150 à 250 prêts, dits leveraged loans (prêts à effet de levier). Ces prêts financent des entreprises endettées, notées en catégorie spéculative, souvent à la suite d'un rachat par un fonds. Trois traits les définissent. Ils sont à taux variable, l'Euribor plus une marge de l'ordre de 350 à 400 points de base. Ils sont garantis par les actifs de l'emprunteur. Et ils sont de premier rang : en cas de faillite, ils passent avant toutes les obligations de l'entreprise, d'où des recouvrements historiques de 60 à 70 %, contre 40 % environ pour une obligation classique.

**La cascade.** Pour financer cet achat, le véhicule émet des tranches hiérarchisées, et les flux du portefeuille les servent dans l'ordre, comme une cascade remplit des bassins successifs. Les intérêts des prêts paient d'abord la tranche AAA, puis la AA, et ainsi de suite jusqu'au reliquat, l'**equity**, qui touche ce qui reste. Les pertes remontent en sens inverse : chaque défaut ronge d'abord l'equity, puis les tranches basses, une à une. La tranche AAA est donc protégée par tout ce qui se trouve sous elle, un coussin d'environ 38 % du portefeuille. Les ordres de grandeur d'une structure européenne type, avec les écarts cotés en 2026 :

| Tranche | Part de la structure | Coussin sous la tranche | Écart sur l'Euribor |
|---|---|---|---|
| AAA | ≈ 60 % | ≈ 38 % | ≈ 130 bp |
| AA | ≈ 11 % | ≈ 27 % | ≈ 200 bp |
| A | ≈ 6 % | ≈ 21 % | ≈ 250 bp |
| BBB | ≈ 7 % | ≈ 15 % | ≈ 335 bp |
| BB | ≈ 5 % | ≈ 10 % | 550 à 850 bp |
| Equity | ≈ 9 % | aucun | le solde, non contractuel |

(bp, basis points : les points de base, centièmes de point de pourcentage.)

**L'arithmétique du coussin.** La perte sur un prêt vaut le défaut multiplié par la part non recouvrée. À 65 % de recouvrement, chaque défaut coûte 35 % de sa mise ; pour percer un coussin de 38 %, il faudrait donc que la quasi-totalité du portefeuille fasse défaut pendant la vie du véhicule. La pire année de l'histoire des leveraged loans, 2009, a vu environ 10 % de défauts et des recouvrements dégradés, soit une perte de panier de l'ordre de 4 %. Il faudrait enchaîner une décennie de 2009 pour approcher le coussin, et les garde-fous internes fermeraient les vannes bien avant.

**Les garde-fous internes.** La structure se surveille elle-même, par des tests contractuels vérifiés chaque trimestre. Le test de surdimensionnement (overcollateralisation) contrôle que la valeur des prêts couvre chaque tranche avec la marge prévue ; le test de couverture des intérêts fait de même pour les flux. Dès qu'un test casse, la cascade se referme : les paiements aux tranches basses et à l'equity sont suspendus, et l'argent rembourse la AAA par anticipation. Autrement dit, la structure se dégrade au profit du senior. S'y ajoutent des règles de diversification, au plus 2 % environ par emprunteur et des plafonds par secteur, qui interdisent au gérant de concentrer le portefeuille.

**La vie du véhicule.** Un CLO naît avec un calendrier : deux ans environ sans remboursement anticipé possible (le non-call), puis quatre à cinq ans de réinvestissement, pendant lesquels le gérant recycle les remboursements en nouveaux prêts, puis l'amortissement, où chaque euro remboursé descend la cascade et éteint la AAA en premier. L'ETF efface ce calendrier pour vous : il vend les tranches qui s'amortissent, souscrit aux émissions nouvelles, et livre en continu le même profil. C'est un confort, et un rappel : vous ne détenez pas un titre qui mûrit, vous détenez une exposition permanente au spread.

**Et 2008, alors ?** Les CLO d'avant-crise ont traversé l'épreuve sans qu'une seule tranche AAA perde un centime de nominal, pendant que leurs cousins CDO, adossés à des crédits immobiliers corrélés entre eux, s'effondraient. La différence tenait au collatéral et à la cascade. Les structures actuelles, redessinées après 2010, portent en plus un coussin renforcé : la subordination de la AAA est passée d'environ 25 % à près de 40 %. Le bilan chiffré est l'argument à retenir : sur environ 7 000 tranches AAA notées entre 1993 et 2022, aucun défaut, jamais.

**D'où vient le spread, si le défaut est introuvable ?** De trois choses qui ne sont pas du risque de crédit. La complexité : chaque CLO s'accompagne d'une documentation de plusieurs centaines de pages, qu'une équipe doit savoir lire. La liquidité : les tranches se négocient de gré à gré, par appels d'offres, pas en continu sur une bourse. Et l'étroitesse de la base d'acheteurs : les fonds monétaires n'ont pas le droit d'y toucher, beaucoup de mandats institutionnels non plus. C'est une prime de barrière à l'entrée plus qu'une prime de danger. L'ETF abaisse justement la barrière ; si assez d'épargne s'y engouffre, la prime se comprimera, comme finissent toutes les primes de ce genre.

**Ce que ça rapporte, mesuré et non promis.** Le compartiment est jeune. Les premiers ETF UCITS en euro sont nés entre septembre 2024 et l'été 2025, et facturent 0,25 à 0,35 %. Mesurés chacun sur sa propre fenêtre contre l'ETF monétaire, ils se tiennent dans un mouchoir : 125 à 136 points de base de plus par an, nets de frais, pour une volatilité inférieure à 1 % et une pire baisse comprise entre 0,3 et 0,7 %. Cette uniformité est le vrai enseignement. Sur un actif de spread, tant que rien ne casse, l'écart entre gérants ne se voit pas, et la seule différence durable entre deux véhicules est leur coût. L'équivalent américain, plus ancien, montre un cycle de plus : depuis octobre 2020, 4,37 % par an en dollars contre 3,04 % pour les bons du Trésor à trois mois, soit 133 points de base, avec une pire baisse de 2,60 % à l'été 2022, quand le spread s'est écarté.

::: figure echelle-du-cash
Les quatre étages, mesurés et non promis. À droite, l'écart de rendement annualisé au monétaire. À gauche, la pire baisse subie sur la même fenêtre. Les deux barres n'ont pas la même nature, l'une est un flux annuel, l'autre un accident ponctuel, et c'est exactement l'arbitrage à trancher. La dernière ligne est en dollars : elle sert à voir un cycle de plus, pas à décrire ce qu'achète un Européen.
:::

**Ce qui peut mal se passer.** Ni le taux (coupon variable, duration presque nulle : 2022, l'année qui a ravagé l'obligataire à taux fixe, a été neutre pour ces tranches), ni le défaut, on vient de le voir. Le risque est le prix de marché, et l'histoire en donne la mesure, avec à chaque fois le délai de retour à la normale.

- **2009.** Des tranches AAA de première génération cotent 70 à 80 % du **pair**, la valeur de remboursement du titre (100 % du nominal). Elles seront toutes remboursées à 100, mais les cotations ont mis deux à trois ans, le temps que le marché du crédit se rétablisse, à y remonter.
- **Mars 2020.** Les écarts passent d'environ 130 à plus de 500 points de base en quelques semaines, et les prix tombent de 100 à 85-90. Le retour au pair intervient au quatrième trimestre 2020, soit près de neuf mois plus tard, les défauts constatés (3,2 %) s'étant révélés très inférieurs aux 8 à 12 % que le marché redoutait.
- **Automne 2022.** Les fonds de pension britanniques, pris dans la crise de leurs couvertures de taux, vendent leurs CLO en urgence pour lever du cash, et les tranches en euro décrochent. Sur l'ETF en dollars, le creux est à −2,6 % ; il est effacé en cinq mois, et le sommet précédent retrouvé un peu moins d'un an après avoir été quitté.

Trois épisodes, une même leçon : cette ligne se vend au prix du jour, pas à sa valeur de remboursement, et le jour où l'on a besoin de vendre n'est pas forcément un bon jour. D'où la règle d'emploi plus bas, qui interdit d'y loger les dépenses des douze prochains mois.

## Sous la AAA : la mezzanine et l'equity

Le tableau plus haut donne le tarif des étages inférieurs : quelques dizaines de points de base de plus à chaque cran jusqu'à la BBB, puis un saut vers 550 à 850 pour la BB. Chaque marche descendue enlève du coussin et ajoute du rendement.

Le point à ne pas manquer est que **le palmarès de la AAA ne se transmet pas vers le bas**. L'absence totale de défaut vaut pour le sommet de la structure, précisément parce que les tranches inférieures existent pour absorber les pertes à sa place. Les CLO émis avant 2008, ceux de la génération que le marché appelle 1.0, ont bel et bien fait souffrir leurs tranches basses, et une BB de CLO se comporte, dans un accès de tension sur le crédit, comme du haut rendement avec un levier intégré. Ce n'est pas de la trésorerie améliorée, c'est un actif risqué, à comparer aux actions et non au monétaire ([[faux-actifs-defensifs]]).

L'offre disponible suit cette frontière, et elle protège le lecteur malgré lui. En UCITS et en euro, seule la AAA existe en ETF, avec quatre véhicules concurrents. La mezzanine n'est logée en ETF qu'aux États-Unis, hors de portée du particulier européen ([[etf-ucits-europeens]]). Pour qui veut vraiment descendre l'escalier, l'accès passe par des sociétés d'investissement cotées à Amsterdam et à Londres, spécialisées dans la mezzanine et l'equity de CLO, qui versent des dividendes trimestriels et cotent souvent sous leur valeur d'actif. C'est un placement en actions de crédit structuré, avec la volatilité qui va avec, et ce n'est plus le sujet de cet article.

Une contrainte réglementaire mérite enfin d'être connue, car elle façonne ce que vous achetez. La règle européenne de rétention du risque oblige l'émetteur à conserver une part de ce qu'il vend. Elle exclut de fait la plupart des CLO américains d'un véhicule UCITS, si bien qu'un ETF CLO en euro est pour l'essentiel un portefeuille de prêts européens. Une concentration géographique assumée, en échange d'un alignement d'intérêts que le marché de 2007 n'avait pas.

## Les fonds à échéance, un panier d'obligations avec une date de fin

Un fonds à échéance, ou fonds daté, détient un panier d'obligations qui arrivent toutes à maturité la même année. À la date prévue, il rembourse les porteurs et se dissout. Ce détail de construction change tout : il rend au fonds une propriété que les fonds obligataires classiques n'ont pas, la date de fin. Un fonds obligataire ordinaire remplace ses titres à mesure qu'ils mûrissent, sa duration reste constante et sa valeur monte et descend avec les taux, indéfiniment. Un fonds daté voit sa duration fondre année après année jusqu'à zéro, exactement comme l'obligation qu'il imite.

**Ce que vous achetez à l'entrée.** Le jour de l'achat, le rendement actuariel du panier est connu et publié. Si vous conservez jusqu'à l'échéance, c'est votre rendement attendu, sous deux réserves qui ne sont pas des détails, les frais et les défauts. Entre les deux dates, la valeur liquidative bouge avec les taux et les spreads comme n'importe quel fonds obligataire ; cette volatilité intermédiaire ne vous concerne pas si, et seulement si, vous tenez jusqu'au bout.

**Ce qui existe, et jusqu'où.** Deux familles cohabitent. Les ETF à millésimes d'abord, un fonds par année d'échéance, avec deux gammes concurrentes en euro, chez iShares et chez Invesco, pour 0,10 à 0,12 % de frais. Elles publient un millésime par année civile, de l'année en cours jusqu'à six ou sept ans plus loin, surtout en crédit d'entreprise de bonne qualité. C'est la limite à connaître avant de bâtir quoi que ce soit. **Il n'existe pas de millésime à quinze ou vingt ans.** Un fonds daté sait financer une dépense de la décennie qui vient, pas le plancher d'une retraite de trente ans, pour lequel il faut revenir aux titres vifs et à l'échelle construite à la main ([[echelle-obligataire]]). La seconde famille est celle des fonds datés des maisons de gestion françaises, très vendus en assurance-vie, souvent investis en haut rendement ou en crédit de qualité moyenne, avec des frais dix fois supérieurs, une fenêtre de souscription et parfois des pénalités de sortie anticipée.

**Ce que la durée change au rendement.** Rien qui tienne au produit : tout vient de la courbe des taux. Le rendement d'un millésime est celui des obligations de son échéance, et la pente de la courbe décide seule de ce que rapporte un millésime lointain par rapport à un proche. Fin juillet 2026, la courbe souveraine de la zone euro donnait 2,62 % à un an, 2,86 % à cinq ans et 3,18 % à dix ans, plus un écart de crédit qui s'élargit lui aussi avec l'échéance : allonger de quatre ans achetait environ un quart de point. Et le signe s'inverse, c'est le vrai piège. Fin juillet 2023, la même courbe payait 3,42 % à un an contre 2,54 % à dix ans, si bien que le millésime le plus court était le mieux payé de toute la gamme.

La règle qui en découle est courte. Le millésime se choisit sur la **date de la dépense**, jamais sur le rendement affiché. La courbe se regarde ensuite, pour savoir ce que l'appariement coûte ou rapporte cette année-là. Un rendement supérieur au bout de la gamme n'est pas une prime gratuite, c'est le prix de quatre années d'incertitude en plus.

**Comment s'en servir.** L'usage propre est l'appariement, pas la course au rendement. Vous connaissez une dépense datée, le pont jusqu'à la pension, un projet, la part du plancher à financer telle année ([[echelle-obligataire]]). Vous achetez le millésime correspondant, et vous cessez de vous demander où seront les taux ce jour-là. C'est une brique de **plancher**, pas une brique de matelas.

**Les quatre pièges, dans l'ordre.** Vendre avant l'échéance, d'abord, ce qui ressuscite le risque de taux que le produit prétendait éteindre. Le crédit ensuite : un millésime à haut rendement qui affiche 6 % perdra des émetteurs en route, et le rendement affiché est un plafond, pas une promesse. Les frais, qui pèsent proportionnellement lourd sur trois ou quatre ans. Le réinvestissement enfin : à l'échéance, vous récupérez du cash et la question se repose aux taux du moment, ce que la brochure appelle rarement un risque.

::: attention Le vrai comparatif n'est pas le rendement affiché
Deux pièges classiques. La fiscalité d'abord, car ce supplément se compare après impôt : sur un compte-titres, un ETF capitalisant reporte l'imposition à la vente, alors que le fonds euros supporte des prélèvements sociaux annuels et une fiscalité d'enveloppe différente ([[flat-tax-et-imposition]]). La devise ensuite : un ETF CLO libellé en dollar et non couvert transforme un débat à 130 points de base en un pari de change à 10 % de volatilité. Pour une poche courte, seule la version dans votre devise de dépense a un sens.
:::

::: exemple Trois ans de dépenses, trois façons de les garder
Un ménage dépense 30 000 € par an et garde trois ans de côté, soit 90 000 €. Tout en ETF monétaire, cette poche rapporte le taux court, disons 2 %, soit 1 800 € par an. Répartie en trois tiers, livrets et fonds euros, monétaire, CLO AAA, elle rapporte environ 400 € de plus par an avant impôt, contre un risque de baisse d'environ 300 € sur le seul tiers exposé, dans une tension comparable à 2022. L'arbitrage se défend, et il reste modeste. C'est la leçon la plus utile de cet article : la poche courte se gère pour la sécurité, son rendement est un bonus, jamais un objectif.
:::

## La règle d'assemblage

Le plan qui suit est celui d'un ménage en phase de retrait, qui vit de son capital et dont la poche courte a une mission précise : payer les factures sans jamais forcer la vente d'un actif risqué au mauvais moment ([[les-trois-phases]]). Un épargnant encore en activité, dont le salaire couvre les dépenses courantes, a une contrainte de liquidité bien plus légère et peut compresser les deux premières couches. Pour le rentier, trois couches suffisent, et l'ordre compte.

- **Les six premiers mois** de dépenses sont intouchables et disponibles sur-le-champ. Livrets réglementés, éventuellement fonds euros. Aucun risque de marché, aucune exception.
- **Les douze à dix-huit mois suivants** vont en ETF monétaire, ou en fonds euros si l'enveloppe s'y prête. C'est le cœur du matelas, celui qui finance une traversée de marché baissier sans vendre d'actifs risqués ([[marche-baissier-en-retraite]]).
- **Le solde éventuel**, et lui seul, peut monter d'un étage, vers l'obligataire ultra-court ou les CLO AAA. Cette part n'est jamais celle des dépenses des douze prochains mois.

Et une liste d'exclusions, qui vaut autant que les règles. Le haut rendement obligataire n'est pas de la trésorerie : il perd 20 % quand les actions perdent 30 %. Un fonds « obligataire court terme » à deux ou trois ans de duration n'en est pas non plus, 2022 l'a rappelé. Les produits structurés à capital protégé facturent la protection plus cher qu'elle ne vaut ([[faux-actifs-defensifs]]). Et les stablecoins rémunérés, quelle que soit la plateforme, n'ont ni garantie ni prêteur en dernier ressort.

## L'essentiel à retenir

- La poche courte se remplit par étages, du livret au CLO AAA, et le bon critère n'est pas le rendement affiché mais la capacité à vendre à un prix décent le pire jour de la décennie.
- Un CLO AAA est la tranche senior d'un portefeuille de prêts d'entreprise garantis, protégée par un coussin d'environ 38 % : avec des recouvrements de 60 à 70 %, il faudrait que la quasi-totalité des emprunteurs fassent défaut pour l'entamer. Aucune tranche AAA n'a fait défaut sur environ 7 000 tranches notées depuis 1993, 2008 compris.
- Mesuré et non promis, l'ETF CLO AAA en euro rend 125 à 136 points de base de plus que le monétaire, frais déduits, pour une pire baisse sous 0,7 % ; son équivalent en dollars a rendu 133 points de base de plus que les bons du Trésor depuis 2020, avec une pire baisse de 2,60 % en 2022.
- Le spread rémunère la complexité, l'illiquidité et l'étroitesse de la base d'acheteurs, pas un risque de défaut. Il se paie en prix de marché les mauvais jours : 70 à 80 % du pair en 2009, décrochages en mars 2020 et à l'automne 2022.
- Sous la AAA, la mezzanine ajoute du rendement en retirant du coussin, d'environ 335 points de base pour la BBB à 550-850 pour la BB, et le palmarès sans défaut ne descend pas d'un étage. En UCITS et en euro, seule la AAA existe en ETF ; rien sous elle n'est de la trésorerie.
- Le fonds à échéance est la brique voisine, avec une date de fin et un rendement actuariel connu à l'achat. Il finance une dépense datée, pas le matelas, et ne tient sa promesse que si on le garde jusqu'au bout.
- Les millésimes disponibles en euro s'arrêtent à six ou sept ans : cette brique couvre une décennie, pas une retraite. Ce que la durée ajoute au rendement n'est que la pente de la courbe, un quart de point environ mi-2026, négative en 2023 ; on choisit un millésime sur la date de la dépense, pas sur son rendement.
- Assemblage type : six mois en livrets, douze à dix-huit mois en monétaire ou fonds euros, et seulement le solde un étage plus haut. Exclusions permanentes, le haut rendement, l'obligataire court à deux ans de duration, les structurés protégés et les stablecoins rémunérés.

---

## Pour aller plus loin

- LSTA et S&P Global Ratings : les études de défaut par tranche sur le marché des CLO, la source du palmarès cité ici.
- J.P. Morgan, indices CLOIE : les séries de référence du marché des CLO en euro et en dollar, par tranche de notation.
- Banque de France et fédérations d'assureurs : les taux servis par les fonds euros, publiés chaque début d'année.
- Dans ce livre : [[cash-buffer]] (combien garder), [[echelle-obligataire]] (l'autre usage du court terme, l'appariement), [[primes-d-assurance]] (l'étage d'après, quand la poche courte devient une poche de primes), [[obligations-en-retrait]] (pourquoi la duration est une décision séparée).
