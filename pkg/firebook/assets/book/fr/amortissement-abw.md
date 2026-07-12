# Le retrait par amortissement (ABW/TPAW) : l'approche actuarielle

Demandez à un économiste académique comment consommer un patrimoine sur une vie : il ne citera ni Bengen ni Guyton-Klinger, il écrira le modèle de cycle de vie de Samuelson et Merton et en sortira une prescription : chaque année, consommez la rente actuarielle de votre richesse TOTALE (portefeuille + valeur actualisée de vos revenus futurs) sur votre horizon restant, aux rendements attendus courants. Cette prescription a un nom opérationnel dans la communauté FIRE : l'ABW (« Amortization Based Withdrawal »), et une incarnation aboutie : le TPAW (« Total Portfolio Allocation and Withdrawal ») de Ben Mathew, planificateur gratuit devenu la coqueluche des Bogleheads. C'est la règle que la littérature récente préfère : elle est la seule à ne pouvoir NI s'épuiser prématurément NI mourir sur une montagne d'or, elle s'ajuste par petites touches continues plutôt que par les à-coups de −10 % des guardrails, et elle intègre nativement pensions, legs et valorisations. C'est aussi une règle exigeante : son revenu suit les marchés, et elle ne s'exécute pas sans outil. Cet article la traite à fond : la mécanique (le crédit immobilier à l'envers, re-coté chaque année), ses fondations théoriques, les quatre paramètres qui la personnalisent (rendement, horizon, legs, pente), ses pathologies, la comparaison finale avec les guardrails, et son implémentation native dans pofo.

::: cle La règle en une image
Un crédit immobilier calcule la mensualité qui rembourse exactement un capital sur une durée à un taux donné. L'ABW retourne la formule : votre portefeuille est le « prêt » que la vie vous a consenti, et le retrait de l'année est la mensualité qui l'épuiserait EXACTEMENT sur vos années restantes, au rendement attendu courant : recalculée chaque 1er janvier sur le capital réel, l'horizon raccourci d'un an, et les attentes du moment. Tout ce qui arrive (krach, boom, inflation, année de plus) est absorbé par la prochaine cotation, en douceur : c'est un plan qui se replanifie tout seul, perpétuellement.
:::

## La mécanique, pas à pas

La formule est celle du VPW ([[vpw]]) : retrait = W × g / (1 − (1+g)^(−n)) ; ce qui change, c'est le contenu des trois lettres, et chaque changement est une montée en réalisme.

**W n'est pas le portefeuille : c'est la richesse totale.** L'ABW ajoute au portefeuille la **valeur actualisée de tous les flux futurs** : pensions à venir ([[retraite-legale]]), revenus d'appoint programmés, moins les dépenses exceptionnelles prévues. Un couple de 47 ans avec 1,4 M€ de portefeuille et 20 000 €/an de pensions à partir de 67 ans possède en réalité ~1,4 M€ + ~350 000 € (la pension actualisée) : c'est sur CE total que la rente se calcule. Conséquence immédiate et précieuse : le « pont de pension » bricolé du VPW devient un simple terme de la formule, et le retrait est naturellement plus élevé avant la pension (on consomme par anticipation une richesse certaine) : le lissage de consommation que les économistes prescrivent depuis Modigliani.

**g n'est pas gravé : c'est le rendement attendu courant.** Là où le VPW fige 5 % réel pour l'éternité, l'ABW branche l'estimation du moment : les rendements prospectifs ([[rendements-attendus]]), et dans sa meilleure version l'ancrage aux valorisations (g actions ≈ 1/CAPE, [[regles-cape]], [[valorisations-et-cape]]). En marché cher, la rente se calcule sur 3 % et la règle dépense prudemment d'elle-même ; en marché purgé, sur 5,5 % et elle ose. C'est la ligne de partage épistémologique avec le VPW, déjà discutée : justesse conditionnelle contre robustesse de table gravée.

**n n'est pas « jusqu'à 100 ans par convention » : c'est votre horizon choisi**, au quantile prudent du dernier survivant ([[horizon-et-esperance-de-vie]]), éventuellement prolongé ou raccourci par le paramètre de legs (ci-dessous).

Le TPAW ajoute la touche mertonienne finale : puisque la richesse totale inclut une grosse « obligation implicite » (la pension actualisée est un actif sans risque de marché), l'ALLOCATION elle-même se raisonne sur le total : 1,4 M€ de portefeuille à 70 % d'actions plus 350 000 € de pension actualisée font un total à ~56 % d'actions : le quadragénaire à pension future peut porter plus d'actions dans son portefeuille visible qu'il ne le croit, à risque total égal. Allocation et retrait sortent du même calcul : d'où le nom, « Total Portfolio ».

## Les quatre paramètres qui font VOTRE version

**1. Le rendement g, et sa marge de prudence.** Le débat classique : brancher le rendement attendu CENTRAL (la rente est juste en espérance, les déceptions se paient en baisses de revenu futures) ou un rendement DÉCOTÉ de 0,5-1 point (la rente est trop basse en espérance, les bonnes surprises se paient en hausses : on préfère ce sens). La pratique TPAW et la logique de ce livre penchent pour la décote légère : elle transforme la distribution du revenu de « symétrique autour du plan » en « plancher probable + bonnes surprises », psychologiquement bien meilleure ([[psychologie-du-retrait]]).

**2. L'horizon n et le legs.** Viser zéro à l'horizon est le défaut ; pour transmettre, on soustrait de W la valeur actualisée du legs visé (léguer 300 000 € réels = amortir W − 220 000 € environ à 30 ans de distance) : le legs devient un CHOIX chiffré, pas un résidu accidentel ([[succession-et-transmission]], [[depenses-en-retraite]]). Pour la longévité au-delà de n : la même réponse que le VPW, une rente en fin de parcours ([[rentes-et-annuites]]), ou un n déjà pris au 90e percentile.

**3. La pente de consommation (« spending tilt »).** Le TPAW permet de PENCHER la rente : consommer plus au début (les années go-go, quand la santé permet les projets) contre moins à la fin, ou l'inverse. Une pente de −0,5 %/an reproduit le sourire des dépenses réelles ([[depenses-en-retraite]]) et augmente le retrait initial de ~10-15 % : c'est l'argument « Die With Zero » rendu paramétrable et réversible.

**4. Le lissage d'affichage.** L'ABW brut re-cote chaque année : son revenu bouge chaque année. Les implémentations sérieuses ajoutent un amortisseur (moyenner la richesse sur 12 mois, ou n'appliquer que la moitié de l'écart vers la nouvelle rente) : les mêmes techniques que la famille proportionnelle ([[pourcentage-fixe]]), pour les mêmes raisons.

::: science Pourquoi la littérature la préfère
Trois arguments reviennent. La COHÉRENCE : c'est la seule famille dérivée d'un modèle de décision (le cycle de vie de Merton, 1969) plutôt que d'une heuristique testée ex post : chaque propriété s'explique, chaque paramètre a un sens. La DOMINANCE sur les critères modernes : dans les comparatifs (Morningstar sur les RMD, ses cousins réglementaires ; les travaux Bogleheads/TPAW ; ERN sur les règles actuarielles), l'amortissement sert la consommation totale la plus élevée à ruine quasi nulle, sans falaise ni cascade de coupes : ses ajustements sont continus et petits (typiquement ±3-6 %/an hors crise, contre les marches de ±10 % des guardrails). L'ABSENCE DE MÉMOIRE MORTE : comme les règles CAPE, l'ABW ne dépend que de l'état présent : pas de « retrait de référence » historique qui fossilise une décision de l'an 1 ([[regles-cape]]). Le prix, connu : le revenu suit les marchés, et la règle exige un outil et des hypothèses : la dépendance au modèle est son talon d'Achille assumé ([[pieges-des-simulateurs]]).
:::

## ABW contre guardrails : le match des deux finalistes

Les deux familles gagnantes du panorama ([[panorama-strategies-retrait]]) méritent leur comparaison directe, critère par critère :

| Critère | Guardrails modernes ([[guardrails-morningstar]]) | ABW/TPAW |
|---|---|---|
| Revenu au quotidien | STABLE entre les franchissements (le confort du fixe) | Variable chaque année (amorti, mais variable) |
| Forme des ajustements | Marches de ±10 %, rares, après franchissement | Petites touches continues (±3-6 %) |
| Falaise / cascade | Possibles si mal bornés (plancher obligatoire) | Impossibles par construction |
| Consommation totale | Bonne | La plus élevée du panorama |
| Legs | Résiduel, dispersé | Choisi, paramétré |
| Pensions, valorisations | Via le capteur par risque (bien) | Natives dans la formule (mieux) |
| Gouvernance | Trois seuils écrits, revue annuelle | Un outil à faire tourner, des hypothèses à assumer |
| Profil psychologique servi | « Je veux MON revenu, touchez-y le moins possible » | « Je veux consommer juste, j'accepte que ça respire » |

La dernière ligne est la vraie ligne de partage, et elle est personnelle, pas technique. Les guardrails vendent la stabilité du revenu et facturent en à-coups rares mais brutaux ; l'ABW vend l'optimalité de la consommation et facture en variabilité permanente mais douce. Un ménage au plancher haut et au tempérament anxieux vivra mieux sous guardrails ; un ménage élastique et quantitatif, mieux sous ABW ; et l'hybride est légitime : ABW pour le calcul de référence annuel, corridor de ±10 % autour pour l'affichage du budget : beaucoup de praticiens TPAW font exactement cela.

## Dans pofo, et un exemple

L'ABW est natif : la case « Amortize over the horizon (ABW/TPAW) » du tiroir Spending policy. Son implémentation suit fidèlement la doctrine : chaque année simulée, le retrait est le paiement qui épuiserait la richesse courante APRÈS impôts, AUGMENTÉE de la valeur actualisée des pensions à venir, sur les années restantes exactement, au rendement géométrique central : celui de l'ancre CAPE si elle est cochée (la version règle-CAPE-aboutie, [[regles-cape]]). Elle écrase toutes les autres règles de dépense. Les lectures : la §04 (la distribution du revenu vécu : LE critère d'admissibilité contre votre plancher), la frontière §06 (l'ABW se place typiquement dans le coin ruine quasi nulle / variabilité moyenne), et la comparaison A/B avec vos guardrails écrits en deux clics ([[utiliser-la-page-fire]]).

::: exemple Une année d'ABW, chiffres en main
Solène et Marc, 49 ans, horizon jusqu'à 99 ans (50 ans), portefeuille 1,55 M€, pensions 19 000 €/an à 67 ans (valeur actualisée ~310 000 €), legs visé 200 000 € réels (VA ~90 000 €), g décoté 3,2 % réel. Richesse à amortir : 1 550 + 310 − 90 = 1 770 000 € ; rente sur 50 ans à 3,2 % : ~4,05 % de W : **71 700 €**... dont il faut retrancher l'impôt de retrait : net vécu ~66 000 €. Année suivante, deux mondes : marché −18 % (portefeuille 1,27 M€) : W = 1 490 000 €, n = 49 : rente ~60 400 € brut : le revenu net baisse de ~8 % (pas 18 : la pension actualisée et l'horizon raccourci amortissent). Marché +12 % : rente ~76 800 € : +7 %. Dix ans de ce régime : un revenu qui respire de ±5-8 % par an autour d'une pente choisie, jamais de falaise, jamais de cliquet : et le legs de 200 000 € qui se construit en silence dans la formule.
:::

## L'essentiel à retenir

- ABW = la rente actuarielle re-cotée chaque année : retrait = annuité de (portefeuille + pensions actualisées − legs visé) sur l'horizon restant au rendement attendu courant : le modèle de cycle de vie des économistes rendu exécutable ; TPAW en est l'incarnation outillée (allocation comprise).
- Ses propriétés uniques : ni épuisement prématuré ni mort sur un tas d'or, ajustements continus et doux, pensions/legs/valorisations natifs, aucune mémoire morte : la préférence de la littérature récente.
- Ses quatre paramètres personnels : g (décotez de 0,5-1 point : plancher probable + bonnes surprises), n (90e percentile du dernier survivant), legs (un choix chiffré), pente (le sourire des dépenses, paramétrable).
- Son prix : un revenu qui respire (±3-8 %/an) et une dépendance au modèle assumée : outil obligatoire, hypothèses à auditer, lissage d'affichage recommandé.
- Face aux guardrails : stabilité contre optimalité : un choix de tempérament plus que de technique, hybride légitime : la décision finale se prend dans [[choisir-sa-strategie]], et l'A/B se joue dans pofo (case ABW + ancre CAPE, §04, §06).

---

## Pour aller plus loin

- Ben Mathew, le planificateur TPAW (tpawplanner.com, gratuit) et le fil « Total portfolio allocation and withdrawal » sur Bogleheads : la doctrine complète et l'outil.
- Merton, « Lifetime Portfolio Selection under Uncertainty » (1969) et Samuelson (1969) : les fondations ; Irlam (aacalc) pour les versions numériques modernes.
- Bogleheads wiki, « Amortization based withdrawal formulas » : les formules et variantes.
- Early Retirement Now sur les règles actuarielles et la critique de « Die With Zero » (Part 60) ([[serie-ern]]).
- Dans ce livre : [[vpw]] (le cousin à table gravée), [[regles-cape]] (le même esprit côté valorisations), [[guardrails-morningstar]] (le finaliste adverse), [[choisir-sa-strategie]] (le verdict).
