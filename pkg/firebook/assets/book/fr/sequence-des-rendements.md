# Le risque de séquence : le vrai ennemi du retraité

Deux retraités partent avec le même million, le même portefeuille, le même retrait de 40 000 € indexé. Sur trente ans, leurs portefeuilles réalisent exactement le même rendement moyen.

L'un finit avec deux millions ; l'autre est ruiné à 78 ans. La seule différence : l'**ordre** dans lequel les mêmes rendements sont arrivés.

C'est le risque de séquence des rendements (sequence of returns risk), et c'est le concept central de tout ce livre. Il explique pourquoi le taux de retrait sûr est si bas au regard des rendements moyens. Il explique pourquoi les dix premières années de retraite dominent tout. Il explique enfin pourquoi la plupart des protections intelligentes, trajectoires d'allocation glissantes (glidepaths), matelas de liquidités (buffers), flexibilité et revenus partiels, sont en réalité des armes anti-séquence. À la fin de cette page, vous saurez reconnaître ce risque, le mesurer sur votre propre plan, et vous connaîtrez la carte des parades.

::: cle L'asymétrie fondamentale
Sans retraits, l'ordre des rendements est indifférent : +30 % puis −20 % ou −20 % puis +30 % donnent le même capital (l'addition des logarithmes commute). Avec des retraits, l'ordre devient déterminant : chaque euro retiré pendant un creux est un euro vendu au pire prix, définitivement soustrait au rebond. Le retraité transforme des pertes temporaires en pertes permanentes, à hauteur exacte de ses retraits. Voilà pourquoi le même portefeuille est bien plus risqué en phase de retrait qu'en phase d'accumulation.
:::

::: figure sequence-risk
Deux retraités, même rendement moyen sur trente ans et retraits identiques : celui qui subit le krach tôt s'épuise, celui qui le subit tard survit. Seul l'ordre des rendements diffère.
:::

## Le mécanisme, sur un exemple qu'on n'oublie pas

Prenons trois années de rendements réels : +20 %, +10 %, −25 %. Moyenne géométrique : peu ou prou nulle (1,20 × 1,10 × 0,75 = 0,99). Capital de départ : 1 000 000 €, retrait de 40 000 € en début de chaque année.

**Séquence favorable (le krach à la fin)** : +20 %, +10 %, −25 %.

| Année | Capital avant retrait | Après retrait | Rendement | Capital final |
|---|---|---|---|---|
| 1 | 1 000 000 | 960 000 | +20 % | 1 152 000 |
| 2 | 1 152 000 | 1 112 000 | +10 % | 1 223 200 |
| 3 | 1 223 200 | 1 183 200 | −25 % | 887 400 |

**Séquence défavorable (le krach d'abord)** : −25 %, +10 %, +20 %.

| Année | Capital avant retrait | Après retrait | Rendement | Capital final |
|---|---|---|---|---|
| 1 | 1 000 000 | 960 000 | −25 % | 720 000 |
| 2 | 720 000 | 680 000 | +10 % | 748 000 |
| 3 | 748 000 | 708 000 | +20 % | 849 600 |

Mêmes rendements, même moyenne, mêmes retraits : 887 400 € contre 849 600 €. Presque 38 000 € d'écart, un retrait annuel entier, en trois ans seulement. Étirez ce mécanisme sur une décennie entière de marché baissier au départ. Ajoutez des retraits qui pèsent chaque année un pourcentage croissant d'un capital qui fond. Vous obtenez alors la différence entre les millésimes 1966 et 1982 de l'histoire américaine. Le premier frôle la ruine, le second finit opulent, pour des rendements de long terme comparables ([[etude-trinity]]).

Voici l'intuition à retenir. En accumulation, un krach précoce est une aubaine, car vous achetez bas pendant des années. En retrait, c'est une hémorragie, car vous vendez bas pendant des années. Le même événement change de signe selon le sens des flux. C'est pourquoi votre glorieux historique d'épargnant, « j'ai traversé 2008 et 2020 sans broncher », ne prouve rien sur votre exposition de rentier. Vous étiez simplement du bon côté des flux.

## La fenêtre fragile : les dix premières années

Le risque de séquence n'est pas uniformément réparti dans le temps. Il se concentre massivement sur le début de la retraite, pour une raison mécanique. C'est là que le capital est le plus gros en proportion des retraits restants, et que la trajectoire a le plus d'années devant elle pour diverger. Un krach à l'année 25 d'une retraite de 30 ans est presque indolore : l'essentiel des retraits est derrière, et le capital requis pour finir est faible. Le même krach à l'année 2 gouverne tout le reste de la trajectoire.

La recherche (ERN volet 15 notamment, [[serie-ern]]) quantifie cette intuition : la corrélation entre le succès final d'un plan et les rendements réalisés est écrasante pour les 5 à 10 premières années, faible ensuite. En pratique, le sort d'une retraite de 40 ans se joue **aux trois quarts dans sa première décennie**. La page FIRE consacre une section entière à cette « décennie décisive » ([[utiliser-la-page-fire]]). Elle montre la dispersion des issues finales, conditionnée au rendement des dix premières années de chaque scénario.

Trois conséquences pratiques découlent de cette concentration temporelle.

**1. La protection peut être temporaire.** Puisque le danger est concentré, les défenses coûteuses (allocations prudentes, buffers, revenus d'appoint) n'ont pas besoin d'être éternelles. Les concentrer sur la fenêtre fragile capte l'essentiel du bénéfice pour une fraction du coût. C'est le fondement des glidepaths « rising equity » de Kitces et Pfau ([[glidepaths]]) : partir prudent, puis remonter l'exposition actions à mesure que la fenêtre se referme.

**2. La date de départ est un paramètre de risque.** Partir au sommet d'un marché cher augmente la probabilité que la fenêtre fragile contienne le krach ([[valorisations-et-cape]]). C'est aussi ce qui rend le « one more year » partiellement rationnel dans un marché euphorique, et coûteux dans un marché déjà purgé ([[une-annee-de-plus]]).

**3. Les premières années d'un plan se surveillent différemment.** Un plan qui traverse sa première décennie sans accroc majeur a, statistiquement, gagné la partie ; la vigilance peut décroître. Les seuils d'alerte utiles sont donc datés, pas uniformes ([[quand-s-inquieter]]).

::: science Mesurer le risque de séquence chez vous
Deux lectures dans la page FIRE rendent votre exposition visible. La première, la section « décennie décisive ». Si les scénarios dont la première décennie tombe dans le pire quartile finissent presque tous ruinés, votre plan est un pari sur la séquence ; s'ils survivent abîmés, il est robuste. La seconde, le modèle « sequence stress ». Il garde la même moyenne de long terme que le modèle central, mais les mauvaises années y arrivent en grappes, à la façon des marchés baissiers persistants (chaînes de Markov), au lieu d'être saupoudrées indépendamment. L'écart de ruine entre les deux colonnes est, précisément, le prix de la séquence dans votre plan ([[la-machine-pofo]]). Un écart faible signale un plan naturellement bien défendu, par un retrait bas, de la flexibilité ou des revenus. Un écart de plusieurs points signale que les parades ci-dessous méritent votre attention.
:::

## Pourquoi les moyennes vous mentent

Le risque de séquence explique un paradoxe qui déroute tous les débutants : comment un portefeuille « à 5 % réel de moyenne » ne peut-il soutenir que 3,5 % de retrait ? Ne devrait-on pas pouvoir retirer la moyenne, perpétuellement ?

Non, pour deux raisons qui s'empilent. La première est le frein de la volatilité (volatility drag). La croissance composée d'un portefeuille volatil est inférieure à sa moyenne arithmétique, d'environ la moitié de la variance ([[rendements-arithmetiques-geometriques]]). Un « 5 % de moyenne » assorti de 15 % de volatilité compose en réalité à ~3,9 %. La seconde raison est la séquence. Même le rendement géométrique n'est « retirable » que si les rendements arrivent régulièrement. Leur irrégularité, combinée à des retraits fixes, consomme une prime supplémentaire. Le taux de retrait sûr est donc structurellement inférieur au rendement géométrique espéré, lui-même inférieur à la moyenne arithmétique qu'affichent les brochures. Retenez la hiérarchie : **moyenne arithmétique > moyenne géométrique > taux de retrait soutenable**. Chaque marche coûte typiquement 0,5 à 1,5 point.

::: attention Le simulateur trop lisse
Tout modèle qui tire les années indépendamment (Monte-Carlo naïf, y compris le modèle Student-t central) sous-estime légèrement le risque de séquence : les vrais marchés font des grappes, des tendances, des décennies perdues, pas des tirages de loterie ([[pieges-des-simulateurs]]). C'est pourquoi la page FIRE affiche, à côté du modèle central, un « sequence stress » et le rejeu de l'échantillon mondial ([[anarkulova-cederburg]]). La règle est nette : si votre plan n'est acceptable que dans la colonne centrale, il ne l'est pas ([[rendre-monte-carlo-pertinent]]).
:::

## La carte des parades

Chaque grande famille de protections du sujet est, au fond, une arme anti-séquence. Les voici en une table, avec leur mécanisme et leur chapitre.

| Parade | Mécanisme anti-séquence | Où c'est traité |
|---|---|---|
| Retrait initial plus bas | Réduit la ponction pendant un éventuel creux précoce | [[combien-il-vous-faut]] |
| Dépenses flexibles, garde-fous (guardrails) | Retire moins, précisément quand les prix sont bas | [[flexibilite-realite]], [[guardrails-morningstar]], [[guyton-klinger]] |
| Matelas de liquidités, compartiments (buckets) | Vend du cash plutôt que des actions au creux | [[cash-buffer]], [[strategie-buckets]], [[recharger-ou-pas]] |
| Glidepath (bond tent) | Réduit l'exposition actions pendant la fenêtre fragile seulement | [[glidepaths]] |
| Revenus partiels au début (Barista) | Réduit les retraits nets pendant la fenêtre fragile | [[retour-au-travail]], [[revenus-complementaires]] |
| Actifs défensifs décorrélés | Amortit la profondeur du creux lui-même | [[actifs-defensifs]], [[or-en-retrait]], [[managed-futures]] |
| Rente ou plancher garanti | Sort une partie des dépenses du jeu de la séquence | [[rentes-et-annuites]] |
| Départ conditionné aux valorisations | Évite d'ouvrir la fenêtre fragile au sommet | [[valorisations-et-cape]] |

Aucune parade n'est gratuite. Le retrait bas coûte des années de travail. Le cash et les rentes coûtent du rendement. La flexibilité coûte du confort, les revenus partiels coûtent de la liberté. Concevoir un plan ([[construire-son-plan]], [[choisir-sa-strategie]]), c'est acheter cette protection anti-séquence au meilleur prix pour votre situation. Un ménage au plancher de dépenses déjà haut achètera plutôt du matelas de liquidités et des revenus différés. Un ménage flexible achètera surtout une règle de retrait adaptative, la protection au meilleur rapport qualité-prix pour presque tout le monde.

::: exemple La même retraite, avec et sans parade
Plan de base : 1 M€, 60/40, 40 000 €/an rigides, 45 ans d'horizon ; ruine « sequence stress » ~18 %. Variante A, garde-fous (baisse à 36 000 € plafonnée quand le taux de retrait courant dépasse 5 %) → ruine ~7 %. Variante B, 3 ans de dépenses en matelas de liquidités, consommé dans les creux et rechargé aux sommets → ruine ~12 %. Variante C, 12 000 €/an de revenus d'appoint les 8 premières années → ruine ~8 %. Variante A+C → ~3 %. Les chiffres exacts dépendent du modèle (testez les vôtres, [[utiliser-la-page-fire]]) ; la hiérarchie, elle, est robuste : la flexibilité règle d'abord, les revenus précoces ensuite, le matelas en appoint.
:::

## Trois malentendus à liquider

**« Le risque de séquence, c'est le risque de krach. »** Non. C'est le risque d'un krach ou d'une érosion **mal placés**. Le pire millésime américain n'est pas 1929 mais 1966 : pas de krach spectaculaire, quinze ans d'étouffement réel par l'inflation ([[etude-trinity]], [[inflation-et-taux-de-retrait]]). Les parades purement « anti-krach » (options, stop-loss) ratent ce mode de défaillance ; les parades anti-séquence (flexibilité, revenus, actifs réels) le couvrent.

**« Je suis passif et long terme, donc immunisé. »** L'immunité du passif vaut en accumulation, où l'ordre est indifférent. Dès le premier retrait, vous êtes dans le jeu de la séquence, aussi passif soit le portefeuille. C'est le point aveugle classique de l'épargnant chevronné qui aborde la retraite avec les réflexes de l'accumulation ([[erreurs-classiques-fire]]).

**« La séquence est imprévisible, donc rien à faire. »** La séquence est imprévisible ; l'**exposition** à la séquence, elle, se mesure et se réduit, c'est tout l'objet de la table ci-dessus. On ne prévoit pas la pluie, on construit un toit.

## L'essentiel à retenir

- Avec des retraits, l'ordre des rendements compte autant que leur moyenne : les pertes précoces deviennent permanentes à hauteur des retraits effectués dedans.
- Le danger se concentre sur les 5-10 premières années : la « fenêtre fragile » ; un plan qui la traverse bien a statistiquement gagné.
- Hiérarchie à mémoriser : moyenne arithmétique > géométrique > taux soutenable ; les brochures vous vendent la première, vous vivez du troisième.
- Toutes les grandes protections du sujet sont des armes anti-séquence, chacune avec son coût ; la flexibilité encadrée offre le meilleur rapport protection/coût pour la plupart des plans.
- Mesurez votre exposition : sections « décennie décisive » et modèle « sequence stress » de la page FIRE, et lisez ensuite [[ruine-et-probabilites]] pour interpréter les chiffres.

---

## Pour aller plus loin

- Early Retirement Now, SWR Series volets 14-15 (« Sequence of Return Risk ») : la démonstration que la séquence explique l'essentiel du résultat ; volet 53 sur les couvertures ([[serie-ern]]).
- Kitces & Pfau, « Reducing Retirement Risk with a Rising Equity Glide Path », *Journal of Financial Planning*, 2014 : la parade par l'allocation ([[glidepaths]]).
- Moshe Milevsky, « Retirement Ruin and the Sequencing of Returns » : la formalisation actuarielle.
- Dans ce livre : [[les-maths-du-4-pourcent]] (la pénalité de séquence chiffrée, ~1,8 point dans la cascade du 4 %) et [[pourquoi-la-diversification-marche]] (la diversification comme remède au même risque).
- Dans la page FIRE : la section « décennie décisive » et le modèle « sequence stress » ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
