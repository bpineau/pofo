# Le retrait par amortissement (ABW/TPAW) : l'approche actuarielle

Demandez à un économiste académique comment consommer un patrimoine sur une vie. Il ne citera ni Bengen ni Guyton-Klinger. Il écrira le modèle de cycle de vie de Samuelson et Merton, et en sortira une prescription. Chaque année, consommez la rente actuarielle de votre richesse **totale** (le portefeuille plus la valeur actualisée de vos revenus futurs) sur votre horizon restant, aux rendements attendus du moment. Cette prescription a un nom dans la communauté FIRE, l'ABW (« Amortization Based Withdrawal »). Elle a aussi une incarnation aboutie, le TPAW (« Total Portfolio Allocation and Withdrawal ») de Ben Mathew, un planificateur gratuit devenu la coqueluche des Bogleheads.

C'est la règle que la littérature récente préfère. Elle est la seule à ne pouvoir ni s'épuiser trop tôt ni laisser mourir son détenteur sur une montagne d'or. Elle s'ajuste par petites touches continues, là où les guardrails (les garde-fous) procèdent par à-coups de −10 %. Et elle intègre d'emblée les pensions, les legs et les valorisations. C'est aussi une règle exigeante. Son revenu suit les marchés, et elle ne s'exécute pas sans outil.

Cet article la traite à fond. D'abord la mécanique, le crédit immobilier à l'envers, recalculé chaque année. Puis ses fondations théoriques, les quatre paramètres qui la personnalisent (rendement, horizon, legs, pente) et ses pathologies. Enfin la comparaison avec les guardrails, et sa mise en œuvre dans la page FIRE.

::: cle La règle en une image
Un crédit immobilier calcule la mensualité qui rembourse exactement un capital sur une durée, à un taux donné. L'ABW retourne la formule. Votre portefeuille est le « prêt » que la vie vous a consenti, et le retrait de l'année est la mensualité qui l'épuiserait **exactement** sur vos années restantes, au rendement attendu du moment. Cette mensualité se recalcule chaque 1er janvier, sur le capital réel, avec l'horizon raccourci d'un an et les attentes du moment. Tout ce qui arrive (krach, boom, inflation, une année de plus) est absorbé en douceur par le calcul suivant. C'est un plan qui se replanifie tout seul, indéfiniment.
:::

## La mécanique, pas à pas

La formule est celle du VPW ([[vpw]]) : retrait = W × g / (1 − (1+g)^(−n)). Ce qui change, c'est le contenu des trois lettres. Et chaque changement est un gain de réalisme.

**W n'est pas le portefeuille. C'est la richesse totale.** L'ABW ajoute au portefeuille la **valeur actualisée de tous les flux futurs** : les pensions à venir ([[retraite-legale]]) et les revenus d'appoint programmés, moins les dépenses exceptionnelles prévues. Prenons un couple de 47 ans, avec 1,4 M€ de portefeuille et 20 000 €/an de pensions à partir de 67 ans. Il possède en réalité ~1,4 M€ plus ~350 000 €, la pension une fois actualisée. C'est sur ce total que la rente se calcule. La conséquence est précieuse. Le « pont de pension » bricolé du VPW devient un simple terme de la formule, et le retrait est plus élevé avant la pension, car on consomme par anticipation une richesse certaine. C'est le lissage de consommation que les économistes prescrivent depuis Modigliani.

**g n'est pas gravé. C'est le rendement attendu du moment.** Là où le VPW fige 5 % réel pour l'éternité, l'ABW branche l'estimation courante, les rendements prospectifs ([[rendements-attendus]]). Dans sa meilleure version, il l'ancre aux valorisations (g actions ≈ 1/CAPE, [[regles-cape]], [[valorisations-et-cape]]). En marché cher, la rente se calcule sur 3 % et la règle dépense prudemment d'elle-même. En marché purgé, elle se calcule sur 5,5 % et elle ose davantage. C'est la ligne de partage avec le VPW, déjà discutée : la justesse conditionnelle contre la robustesse d'une table gravée.

**n n'est pas « jusqu'à 100 ans par convention ». C'est votre horizon choisi**, au quantile prudent du dernier survivant ([[horizon-et-esperance-de-vie]]). Le paramètre de legs, décrit plus bas, peut ensuite le prolonger ou le raccourcir.

Le TPAW ajoute la touche mertonienne finale. La richesse totale inclut une grosse « obligation implicite », car la pension actualisée est un actif sans risque de marché. L'**allocation** se raisonne donc sur le total. Reprenons le couple : 1,4 M€ de portefeuille à 70 % d'actions, plus 350 000 € de pension actualisée, font un ensemble à ~56 % d'actions. Autrement dit, le quadragénaire à pension future peut porter plus d'actions dans son portefeuille visible qu'il ne le croit, à risque total égal. Allocation et retrait sortent du même calcul. D'où le nom, « Total Portfolio ».

## Les quatre paramètres qui font VOTRE version

**1. Le rendement g, et sa marge de prudence.** Le débat est classique. On peut brancher le rendement attendu **central** : la rente est alors juste en espérance, et les déceptions se paient en baisses de revenu futures. On peut aussi le **décoter** de 0,5 à 1 point : la rente est un peu basse en espérance, mais les bonnes surprises se paient en hausses, un sens que l'on préfère. La pratique TPAW et la logique de ce livre penchent pour la décote légère. Elle transforme la distribution du revenu, de « symétrique autour du plan » en « plancher probable, plus des bonnes surprises ». C'est psychologiquement bien meilleur ([[psychologie-du-retrait]]).

**2. L'horizon n et le legs.** Viser zéro à l'horizon est le réglage par défaut. Pour transmettre, on soustrait de W la valeur actualisée du legs visé. Léguer 300 000 € réels revient à amortir W moins 220 000 € environ, à 30 ans de distance. Le legs devient ainsi un **choix** chiffré, pas un résidu accidentel ([[succession-et-transmission]], [[depenses-en-retraite]]). Pour la longévité au-delà de n, la réponse est la même que pour le VPW : une rente en fin de parcours ([[rentes-et-annuites]]), ou un n déjà pris au 90e percentile.

**3. La pente de consommation (« spending tilt »).** Le TPAW permet d'**incliner** la rente. On consomme plus au début, dans les années go-go, quand la santé permet encore les projets, contre moins à la fin. Ou l'inverse. Une pente de −0,5 %/an reproduit le sourire des dépenses réelles ([[depenses-en-retraite]]) et relève le retrait initial de ~10-15 %. C'est l'argument « Die With Zero », rendu paramétrable et réversible.

**4. Le lissage d'affichage.** L'ABW brut se recalcule chaque année, donc son revenu bouge chaque année. Les implémentations sérieuses ajoutent un amortisseur : moyenner la richesse sur 12 mois, ou n'appliquer que la moitié de l'écart vers la nouvelle rente. Ce sont les techniques de la famille proportionnelle ([[pourcentage-fixe]]), employées pour les mêmes raisons.

::: science Pourquoi la littérature la préfère
Trois arguments reviennent. Le premier est la **cohérence**. C'est la seule famille dérivée d'un modèle de décision, le cycle de vie de Merton (1969), qui applique à la consommation le cadre d'utilité posé dans [[decider-sous-incertitude]]. Les autres règles ne sont que des heuristiques testées après coup. Ici, chaque propriété s'explique et chaque paramètre a un sens. Le deuxième est la **dominance** sur les critères modernes. Dans les comparatifs (Morningstar sur les RMD et leurs cousins réglementaires, les travaux Bogleheads et TPAW, ERN sur les règles actuarielles), l'amortissement sert la consommation totale la plus élevée pour une ruine quasi nulle, sans falaise ni cascade de coupes. Ses ajustements restent continus et petits, typiquement ±3-6 %/an hors crise, contre les marches de ±10 % des guardrails. Le troisième est l'**absence de mémoire morte**. Comme les règles CAPE, l'ABW ne dépend que de l'état présent : aucun « retrait de référence » historique ne vient fossiliser une décision de l'an 1 ([[regles-cape]]). Le prix est connu. Le revenu suit les marchés, et la règle exige un outil et des hypothèses. Cette dépendance au modèle est son talon d'Achille assumé ([[pieges-des-simulateurs]]).
:::

## ABW contre guardrails : le match des deux finalistes

Les deux familles gagnantes du panorama ([[panorama-strategies-retrait]]) méritent une comparaison directe, critère par critère :

| Critère | [[guardrails-morningstar|Guardrails]] | ABW/TPAW |
|---|---|---|
| Revenu au quotidien | **Stable** entre les franchissements (le confort du fixe) | Variable chaque année (amorti, mais variable) |
| Forme des ajustements | Marches de ±10 %, rares, après franchissement | Petites touches continues (±3-6 %) |
| Falaise / cascade | Possibles si mal bornés (plancher obligatoire) | Impossibles par construction |
| Consommation totale | Bonne | La plus élevée du panorama |
| Legs | Résiduel, dispersé | Choisi, paramétré |
| Pensions, valorisations | Via l'indicateur par risque (bien) | Natives dans la formule (mieux) |
| Gouvernance | Trois seuils écrits, revue annuelle | Un outil à faire tourner, des hypothèses à assumer |
| Profil psychologique servi | « Je veux mon revenu, touchez-y le moins possible » | « Je veux consommer juste, j'accepte que ça respire » |

La dernière ligne est la vraie ligne de partage. Elle est personnelle, pas technique. Les guardrails vendent la stabilité du revenu et la facturent en à-coups rares mais brutaux. L'ABW vend l'optimalité de la consommation et la facture en variabilité permanente mais douce. Un ménage au plancher de dépenses élevé et au tempérament anxieux vivra mieux sous guardrails. Un ménage élastique et à l'aise avec les chiffres vivra mieux sous ABW. L'hybride est légitime aussi : l'ABW pour le calcul de référence annuel, avec un corridor de ±10 % autour pour l'affichage du budget. Beaucoup de praticiens TPAW font exactement cela.

## Dans la page FIRE, et un exemple

Dans la page FIRE, l'ABW est natif : c'est la case « Amortize over the horizon (ABW/TPAW) » du tiroir Spending policy. Son implémentation suit fidèlement la doctrine. Chaque année simulée, le retrait est le paiement qui épuiserait la richesse courante **après** impôts, **augmentée** de la valeur actualisée des pensions à venir, sur les années restantes exactement, au rendement géométrique central. C'est celui de l'ancre CAPE si elle est cochée, la version règle-CAPE aboutie ([[regles-cape]]). Ce réglage prend le pas sur toutes les autres règles de dépense. Trois lectures aident à le juger. La §04 montre la distribution du revenu vécu et le teste contre votre plancher. La frontière §06 situe l'ABW, en général dans le coin ruine quasi nulle et variabilité moyenne. Enfin, la comparaison A/B avec vos guardrails s'obtient en deux clics ([[utiliser-la-page-fire]]).

::: exemple Une année d'ABW, chiffres en main
Solène et Marc ont 49 ans. Leur horizon va jusqu'à 99 ans, soit 50 ans. Le portefeuille vaut 1,55 M€, les pensions 19 000 €/an à partir de 67 ans (valeur actualisée ~310 000 €), et le legs visé 200 000 € réels (valeur actualisée ~90 000 €). Le rendement décoté g est de 3,2 % réel. La richesse à amortir vaut donc 1 550 + 310 − 90 = 1 770 000 €. La rente sur 50 ans à 3,2 % ressort à ~4,05 % de W, soit **71 700 €**. Il faut en retrancher l'impôt de retrait, pour un net vécu d'environ 66 000 €. L'année suivante ouvre deux mondes. Si le marché fait −18 % (portefeuille à 1,27 M€), alors W = 1 490 000 € et n = 49 → rente ~60 400 € brut, et le revenu net baisse de ~8 % seulement. Pas de −18 %, car la pension actualisée et l'horizon raccourci amortissent le choc. Si le marché fait +12 %, la rente monte à ~76 800 €, soit +7 %. Dix ans de ce régime donnent un revenu qui respire de ±5-8 % par an autour d'une pente choisie, jamais de falaise, jamais de cliquet. Et le legs de 200 000 € se construit en silence dans la formule.
:::

## L'essentiel à retenir

- ABW = la rente actuarielle recalculée chaque année. Le retrait est l'annuité de (portefeuille + pensions actualisées − legs visé) sur l'horizon restant, au rendement attendu du moment. C'est le modèle de cycle de vie des économistes rendu exécutable. Le TPAW en est l'incarnation outillée, allocation comprise.
- Ses propriétés sont uniques : ni épuisement trop tôt ni mort sur un tas d'or, des ajustements continus et doux, des pensions, legs et valorisations intégrés d'emblée, aucune mémoire morte. D'où la préférence de la littérature récente.
- Ses quatre paramètres personnels : g (décotez de 0,5 à 1 point, pour un plancher probable et des bonnes surprises), n (90e percentile du dernier survivant), le legs (un choix chiffré) et la pente (le sourire des dépenses, rendu paramétrable).
- Son prix : un revenu qui respire (±3-8 %/an) et une dépendance au modèle assumée. L'outil est obligatoire, les hypothèses sont à auditer, et le lissage d'affichage est recommandé.
- Face aux guardrails, c'est la stabilité contre l'optimalité, un choix de tempérament plus que de technique, et l'hybride reste légitime. La décision finale se prend dans [[choisir-sa-strategie]], et l'A/B se joue dans la page FIRE (case ABW plus ancre CAPE, §04, §06).

---

## Pour aller plus loin

- Ben Mathew, le planificateur TPAW ([tpawplanner.com](https://tpawplanner.com), gratuit) et le fil « Total portfolio allocation and withdrawal » sur Bogleheads : la doctrine complète et l'outil.
- Merton, « Lifetime Portfolio Selection under Uncertainty » (1969) et Samuelson (1969) : les fondations ; Irlam (aacalc) pour les versions numériques modernes.
- Bogleheads wiki, « Amortization based withdrawal formulas » : les formules et variantes.
- Early Retirement Now sur les règles actuarielles et la critique de « Die With Zero » (volet 60) ([[serie-ern]]).
- Dans ce livre : [[vpw]] (le cousin à table gravée), [[regles-cape]] (le même esprit côté valorisations), [[guardrails-morningstar]] (le finaliste adverse), [[choisir-sa-strategie]] (le verdict).
