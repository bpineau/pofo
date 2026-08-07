# Plancher-plafond et règles Vanguard : la flexibilité bornée

Entre le montant fixe qui ignore les marchés ([[retrait-fixe-bengen]]) et le pourcentage qui les épouse ([[pourcentage-fixe]]), il existe une troisième voie d'une simplicité désarmante : suivre le pourcentage, mais **borner** le mouvement. C'est la famille plancher-plafond, dont la version industrielle est la « dynamic spending rule » de Vanguard : chaque année, visez X % du portefeuille courant, mais interdisez au revenu réel de monter de plus de +5 % ou de descendre de plus de −2,5 % par rapport à l'an dernier.

Deux bornes, rien d'autre : et cette asymétrie douce (on monte deux fois plus vite qu'on ne descend) suffit à transformer la volatilité brutale du pourcentage en glissements vivables, tout en gardant l'essentiel de son auto-correction. C'est la règle que le plus grand gestionnaire d'actifs du monde recommande à ses clients retraités, l'une des plus faciles à exécuter du panorama, et elle est simulable nativement dans pofo.

Cet article la détaille : les deux lignées (le plancher-plafond originel de Bengen lui-même, puis le corridor Vanguard), la mécanique année par année, ce que les bornes font exactement au risque (elles **recréent** une ruine possible : il faut le comprendre et l'accepter), le choix des paramètres, et sa place face aux guardrails et à l'ABW.

::: cle La règle Vanguard en trois lignes
1) Cible de l'année = w × portefeuille courant (w fixé au départ, par exemple 4 %). 2) Plafond : le retrait ne peut dépasser (retrait de l'an dernier, en réel) × 1,05. 3) Plancher : il ne peut descendre sous × 0,975. Le retrait de l'année est la cible, écrêtée par les deux bornes. En croisière, on vit un quasi-pourcentage ; en krach, on **glisse** vers le bas à 2,5 % réel par an au lieu de tomber ; en boom, on monte à 5 % par an au lieu de s'emballer.
:::

## Deux lignées pour une même idée

**La lignée Bengen : le plancher-plafond sur NIVEAU.** Peu le savent : l'inventeur du retrait fixe a lui-même proposé, dès ses travaux des années 1990-2000, une variante « floor-and-ceiling » : prélever un pourcentage du portefeuille courant, mais borné en **niveau** absolu : jamais sous ~85-90 % du retrait initial réel (le plancher), jamais au-dessus de ~120-125 % (le plafond). C'est un pourcentage enfermé dans un couloir fixé une fois pour toutes autour du point de départ. Propriétés : le plancher garantit un revenu minimal prévisible (excellent pour caler le budget incompressible, [[combien-il-vous-faut]]) ; en contrepartie, dans un très mauvais régime, on prélève « au plancher » un portefeuille effondré : la ruine redevient possible, concentrée dans les scénarios extrêmes. Bengen mesurait que ce couloir permettait ~0,25-0,5 point de taux initial de mieux que sa règle fixe : la flexibilité bornée, déjà, achetait son demi-point ([[flexibilite-realite]]).

**La lignée Vanguard : le corridor sur VARIATION.** La règle publiée par Vanguard (recherche « From assets to income », base de son conseil retraite depuis les années 2010) déplace les bornes : elles ne portent plus sur le niveau par rapport à l'an 1, mais sur la **variation** d'une année à l'autre (+5 %/−2,5 % en réel). La différence est profonde : le couloir de Bengen est ancré au passé (le retrait initial reste la référence éternelle : la « mémoire morte » des règles fixes) ; le corridor Vanguard est **sans** ancre : après dix ans de glissements, le revenu peut être à 70 % ou à 140 % du niveau initial : il a suivi la réalité, lentement. C'est un lissage exponentiel asymétrique du pourcentage fixe, cousin direct de la règle de Yale ([[pourcentage-fixe]]) avec des demi-vies différentes à la hausse et à la baisse.

L'asymétrie des bornes n'est pas décorative : elle encode une préférence humaine documentée (les baisses de train de vie font deux fois plus mal que les hausses ne font du bien) et une réalité statistique (les marchés montent plus souvent qu'ils ne baissent : la borne haute travaille plus souvent, freinant l'euphorie ; la borne basse, plus rare, amortit les chocs). Le couple (+5/−2,5) est le réglage Vanguard ; il se personnalise (voir plus bas).

## La mécanique en action : cinq ans difficiles

Rien ne vaut un déroulé. Plan : 1,4 M€, w = 4 %, retrait initial 56 000 €.

| Année | Portefeuille (réel) | Cible 4 % | Bornes (vs an dernier, réel) | Retrait servi |
|---|---|---|---|---|
| 1 | 1 400 000 | 56 000 | : | 56 000 |
| 2 | 1 190 000 (−15 %) | 47 600 | ≥ 54 600 | **54 600** (plancher −2,5 %) |
| 3 | 1 010 000 (−15 %) | 40 400 | ≥ 53 235 | **53 235** (plancher) |
| 4 | 1 090 000 (+8 %) | 43 600 | ≥ 51 904 | **51 904** (plancher, encore) |
| 5 | 1 240 000 (+14 %) | 49 600 | ≤ 54 499 / ≥ 50 606 | **50 606** (plancher) |
| 6 | 1 400 000 (+13 %) | 56 000 | ≤ 53 136 | **53 136** (plafond +5 %) |

Lisez bien les années 4-6 : le portefeuille remonte, mais le retrait **continue** de glisser vers le bas un moment (la cible 4 % est encore loin dessous), puis remonte au rythme plafonné. En six ans de traversée sévère (−28 % au creux), le revenu n'a jamais bougé de plus de 2,5 % par an, pour un sacrifice cumulé maximal de ~10 % : la même séquence sous pourcentage pur aurait servi −28 % en deux ans. Voilà le produit : des glissements à la place des chutes.

Et le prix, visible dans la même table : aux années 3-5, on prélève 5,0-5,3 % d'un portefeuille amputé : la descente plafonnée « emprunte » au capital pendant les crises. Dans les scénarios ordinaires, le portefeuille rembourse à la reprise ; dans un régime vraiment long et hostile ([[regimes-de-marche]]), l'emprunt s'accumule et **la ruine redevient possible**. C'est le point à comprendre : les bornes revendent une partie de la garantie anti-ruine du pourcentage pour acheter la stabilité : la règle Vanguard se place entre le fixe et le pourcentage sur la frontière ([[panorama-strategies-retrait]]), et sa ruine, faible mais non nulle, est un **vrai** chiffre, comparable à celui du fixe : contrairement au 0 % creux du pourcentage pur, et contrairement au chiffre incomparable du Guyton-Klinger sans plancher ([[guyton-klinger]]).

::: science Ce que disent les comparatifs
Dans les études de Vanguard et les comparatifs indépendants (Morningstar teste une variante proche chaque année, [[guardrails-morningstar]]), la règle bornée se comporte en élève régulière : ruine nettement sous le fixe à taux initial égal (typiquement divisée par 2 à 3), consommation totale légèrement supérieure, variabilité du revenu très inférieure au pourcentage pur (écart-type des variations annuelles ~2 % contre ~11 %), legs intermédiaire. Elle ne gagne **aucune** catégorie : elle est deuxième partout, sans pathologie connue : c'est précisément son argument. Les règles qui gagnent une catégorie (l'ABW en consommation, les guardrails en stabilité) paient ailleurs ; le corridor borné est le choix de qui refuse de payer cher quoi que ce soit.
:::

## Choisir ses paramètres

Trois décisions, dans l'ordre d'importance.

**Le pourcentage w.** Même logique que le pourcentage fixe ([[pourcentage-fixe]]) : la borne géométrique s'applique (w < rendement réel géométrique espéré pour un revenu médian stable), avec la même générosité permise par l'auto-correction : 4-4,5 % défendable là où le fixe exigerait 3,25-3,5 %. En marché cher, décotez ([[valorisations-et-cape]]) ou, mieux, laissez l'ancre CAPE de pofo juger votre w.

**Le couple de bornes.** (+5/−2,5) est le standard ; les variantes utiles : (+4/−2) pour les budgets plus rigides (descente encore plus douce, ruine un peu plus haute), (+6/−4) pour les budgets élastiques qui veulent suivre la cible de plus près. La règle de cohérence : la borne basse doit rester **tenable** composée plusieurs années (−2,5 %/an pendant six ans = −14 % : est-ce au-dessus de votre plancher réel ?) : c'est le test d'admissibilité de la famille, à vérifier sur la §04 de pofo.

**L'interaction avec les revenus externes.** Comme toute la famille proportionnelle, la règle s'applique au portefeuille **seul** : en phase à découvert d'un FIRE, le pont de pension se provisionne à part ([[vpw]], [[horizon-et-esperance-de-vie]]) ; une fois la pension au plancher, le corridor sur le portefeuille résiduel devient presque sans risque.

**Dans pofo** : la case « Bounded % of portfolio (Vanguard-style) » implémente exactement la règle (cible = pourcentage initial du portefeuille courant, variations réelles bornées à +5 %/−2,5 %, prime sur les règles flex/guardrails/ratchet). Les lectures utiles : la §04 pour la distribution du revenu (vérifiez le test de la borne basse composée), la frontière §06 où vous la verrez s'asseoir entre le fixe et le VPW, et l'aide au survol qui rappelle honnêtement que la descente plafonnée **peut** laisser filer un effondrement : « unlike VPW/ABW this rule **can** still run out » ([[utiliser-la-page-fire]]).

## Pour qui, face aux deux finalistes

Le profil du corridor borné : le ménage qui veut de la **simplicité** d'**exécution** avant tout (une multiplication, deux comparaisons : pas de seuils de ruine à surveiller, pas d'outil actuariel à faire tourner), un revenu qui ne surprend jamais (±2,5-5 % l'an, connus d'avance), et qui accepte d'être « deuxième partout ». C'est probablement la meilleure règle par défaut pour un conjoint survivant non gestionnaire ([[couple-et-famille]]) et une excellente règle de croisière pour la phase adossée.

Face aux finalistes : contre les guardrails ([[guardrails-morningstar]]), le corridor échange les marches rares de ±10 % (avec leur charge émotionnelle de « décision ») contre des glissements continus sans décision : moins optimal, plus vivable pour beaucoup ; contre l'ABW ([[amortissement-abw]]), il abandonne l'optimalité de consommation et la conscience de l'horizon contre une gouvernance de carte postale. La synthèse générale de sélection reste [[choisir-sa-strategie]].

::: exemple Calibrer sa version en vingt minutes
Ménage : plancher réel 41 000 €, confort 50 000 €, portefeuille 1,3 M€ (w cible : 3,85 %), pension à 15 ans. Test de la borne basse : une traversée de six ans au plancher −2,5 % mène le revenu à 50 000 × 0,975^6 ≈ 43 100 € : au-dessus du plancher : (+5/−2,5) est admissible. Simulation pofo, case Vanguard cochée : ruine centrale 2,8 % (contre 6,1 % en fixe à 50 000 €), §04 : pire quartile à −9 % pendant 5-7 ans : accepté. Variante (+4/−2) testée : ruine 3,4 %, pire quartile −7 % : le ménage préfère et le note dans son plan écrit. Vingt minutes, deux simulations, une règle possédée : à comparer aux heures de débat qu'exigent les guardrails par risque, pour un résultat proche dans la plupart des scénarios.
:::

## L'essentiel à retenir

- Deux lignées : le plancher-plafond de Bengen (bornes en **niveau** autour du retrait initial) et le corridor Vanguard (bornes en **variation** : +5 %/−2,5 % réels par an) : le second, sans mémoire morte, est devenu le standard.
- Le produit : des glissements au lieu de chutes (jamais plus de 2,5 % de baisse réelle par an) ; le prix : la descente plafonnée « emprunte » au capital dans les crises : la ruine redevient possible, faible et **honnête** (comparable à celle du fixe).
- Deuxième partout, première nulle part, aucune pathologie : la règle de qui refuse de payer cher une optimalité quelconque ; probablement le meilleur défaut pour un exécutant non spécialiste.
- Paramètres : w comme un pourcentage (borne géométrique, 4-4,5 % défendable), bornes (+5/−2,5) standard, et le test d'admissibilité : la borne basse composée sur six ans doit rester au-dessus du plancher réel.
- Native dans pofo (« Bounded % of portfolio ») : vérifiez la §04 et la frontière, et comparez en deux clics aux guardrails et à l'ABW avant de choisir ([[choisir-sa-strategie]]).

---

## Pour aller plus loin

- Vanguard, « From assets to income: a goals-based approach to retirement spending » (recherche fondatrice de la règle, gratuite) et ses mises à jour.
- Bengen, *Conserving Client Portfolios During Retirement* (2006) : le floor-and-ceiling originel.
- Morningstar, *The State of Retirement Income* : la variante bornée dans le comparatif annuel ([[guardrails-morningstar]]).
- Dans ce livre : [[pourcentage-fixe]] (la matière première et la règle de Yale), [[guyton-klinger]] et [[guardrails-morningstar]] (l'autre voie du corridor), [[amortissement-abw]] (l'optimalité contre la simplicité).
