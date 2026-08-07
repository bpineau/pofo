# Plancher-plafond et règles Vanguard : la flexibilité bornée

Entre le montant fixe qui ignore les marchés ([[retrait-fixe-bengen]]) et le pourcentage qui les épouse ([[pourcentage-fixe]]), il existe une troisième voie d'une simplicité désarmante. On suit le pourcentage, mais on borne le mouvement. C'est la famille plancher-plafond. Sa version industrielle est la « dynamic spending rule » de Vanguard : chaque année, on vise X % du portefeuille courant, mais on interdit au revenu réel de monter de plus de +5 % ou de descendre de plus de −2,5 % par rapport à l'an dernier.

Deux bornes, rien d'autre. Cette asymétrie douce (on monte deux fois plus vite qu'on ne descend) suffit à transformer la volatilité brutale du pourcentage en glissements vivables, sans rien perdre de l'essentiel de son auto-correction. C'est la règle que le plus grand gestionnaire d'actifs du monde recommande à ses clients retraités. Elle est aussi l'une des plus faciles à exécuter du panorama, et un simulateur la reproduit directement.

Cet article la détaille. Les deux lignées d'abord : le plancher-plafond originel de Bengen lui-même, puis le corridor Vanguard. Ensuite la mécanique année par année, ce que les bornes font exactement au risque (elles recréent une ruine possible, qu'il faut comprendre et accepter), le choix des paramètres, et sa place face aux guardrails et à l'ABW.

::: cle La règle Vanguard en trois lignes
1) Cible de l'année = w × portefeuille courant (w fixé au départ, par exemple 4 %). 2) Plafond : le retrait ne peut dépasser le retrait de l'an dernier (en réel) × 1,05. 3) Plancher : il ne peut descendre sous × 0,975. Le retrait de l'année est la cible, écrêtée par les deux bornes. En croisière, on vit un quasi-pourcentage. En krach, on glisse vers le bas de 2,5 % réel par an au lieu de tomber. En boom, on monte de 5 % par an au lieu de s'emballer.
:::

## Deux lignées pour une même idée

**La lignée Bengen, des bornes en niveau.** Peu le savent, mais l'inventeur du retrait fixe a lui-même proposé une variante « floor-and-ceiling », dès ses travaux des années 1990-2000. Le principe : prélever un pourcentage du portefeuille courant, mais borné en niveau absolu. Jamais sous ~85-90 % du retrait initial réel, c'est le plancher. Jamais au-dessus de ~120-125 %, c'est le plafond. On tient un pourcentage enfermé dans un couloir fixé une fois pour toutes autour du point de départ. Le plancher garantit un revenu minimal prévisible, excellent pour caler le budget incompressible ([[combien-il-vous-faut]]). En contrepartie, dans un très mauvais régime, on prélève « au plancher » sur un portefeuille effondré, et la ruine redevient possible, concentrée dans les scénarios extrêmes. Bengen mesurait que ce couloir rapportait ~0,25-0,5 point de taux initial de mieux que sa règle fixe. La flexibilité bornée, déjà, achetait son demi-point ([[flexibilite-realite]]).

**La lignée Vanguard, un corridor en variation.** La règle publiée par Vanguard (recherche « From assets to income », socle de son conseil retraite depuis les années 2010) déplace les bornes. Elles ne portent plus sur le niveau par rapport à l'an 1, mais sur la variation d'une année à l'autre (+5 %/−2,5 % en réel). La différence est profonde. Le couloir de Bengen est ancré au passé : le retrait initial reste la référence éternelle, la « mémoire morte » des règles fixes. Le corridor Vanguard, lui, n'a pas d'ancre. Après dix ans de glissements, le revenu peut être à 70 % ou à 140 % du niveau initial. Il a suivi la réalité, lentement. C'est un lissage exponentiel asymétrique du pourcentage fixe, cousin direct de la règle de Yale ([[pourcentage-fixe]]), avec des demi-vies différentes à la hausse et à la baisse.

L'asymétrie des bornes n'est pas décorative. Elle encode une préférence humaine documentée : les baisses de train de vie font deux fois plus mal que les hausses ne font du bien. Elle encode aussi une réalité statistique. Les marchés montent plus souvent qu'ils ne baissent, donc la borne haute travaille plus souvent et freine l'euphorie ; la borne basse, plus rare, amortit les chocs. Le couple (+5/−2,5) est le réglage Vanguard, et il se personnalise (voir plus bas).

## La mécanique en action : cinq ans difficiles

Rien ne vaut un déroulé. Plan : 1,4 M€, w = 4 %, retrait initial 56 000 €.

| Année | Portefeuille (réel) | Cible 4 % | Bornes (réel) | Retrait servi |
|---|---|---|---|---|
| 1 | 1 400 000 | 56 000 | (référence) | 56 000 |
| 2 | 1 190 000 (−15 %) | 47 600 | ≥ 54 600 | **54 600** (plancher −2,5 %) |
| 3 | 1 010 000 (−15 %) | 40 400 | ≥ 53 235 | **53 235** (plancher) |
| 4 | 1 090 000 (+8 %) | 43 600 | ≥ 51 904 | **51 904** (plancher, encore) |
| 5 | 1 240 000 (+14 %) | 49 600 | ≤ 54 499 / ≥ 50 606 | **50 606** (plancher) |
| 6 | 1 400 000 (+13 %) | 56 000 | ≤ 53 136 | **53 136** (plafond +5 %) |

::: figure corridor-1966
Le même millésime hostile, trois règles, un seul capital de 1 M€. Le fixe indexé sert 40 k€ sans broncher, jusqu'au mur : le portefeuille est vide en 1994. Le pourcentage pur ne ruine jamais, mais il fait vivre la crise en direct, jusqu'à −55 % de revenu au creux de 1982. Le corridor Vanguard opère la même correction de fond, en glissements de 2,5 % l'an au pire : le revenu descend de 40 à 22,8 k€, mais sur vingt-trois ans, sans jamais imposer de saut. **C'est le produit de la règle : la même adaptation, étalée jusqu'à devenir vivable.**
:::

Lisez bien les années 4-6. Le portefeuille remonte, mais le retrait continue de glisser vers le bas un moment, car la cible 4 % est encore loin en dessous ; il remonte ensuite au rythme plafonné. En six ans de traversée sévère (−28 % au creux), le revenu n'a jamais bougé de plus de 2,5 % par an, pour un sacrifice cumulé maximal de ~10 %. La même séquence sous pourcentage pur aurait servi −28 % en deux ans. Voilà le produit : des glissements à la place des chutes.

Le prix se lit dans la même table. Aux années 3-5, on prélève 5,0-5,3 % d'un portefeuille amputé : la descente plafonnée « emprunte » au capital pendant les crises. Dans les scénarios ordinaires, le portefeuille rembourse à la reprise. Dans un régime vraiment long et hostile ([[regimes-de-marche]]), l'emprunt s'accumule et la ruine redevient possible. C'est le point clé à comprendre. Les bornes revendent une partie de la garantie anti-ruine du pourcentage pour acheter de la stabilité. La règle Vanguard se place donc entre le fixe et le pourcentage sur la frontière ([[panorama-strategies-retrait]]). Sa ruine, faible mais non nulle, est un vrai chiffre, comparable à celui du fixe. C'est là toute sa différence avec le 0 % creux du pourcentage pur, ou avec le chiffre incomparable du Guyton-Klinger sans plancher ([[guyton-klinger]]).

::: science Ce que disent les comparatifs
Dans les études de Vanguard et les comparatifs indépendants (Morningstar teste une variante proche chaque année, [[guardrails-morningstar]]), la règle bornée se comporte en élève régulière : ruine nettement sous le fixe à taux initial égal (typiquement divisée par 2 à 3), consommation totale légèrement supérieure, variabilité du revenu très inférieure au pourcentage pur (écart-type des variations annuelles ~2 % contre ~11 %), legs intermédiaire. Elle ne gagne aucune catégorie. Elle est deuxième partout, sans pathologie connue. C'est précisément son argument. Les règles qui gagnent une catégorie paient ailleurs, l'ABW en consommation, les guardrails en stabilité. Le corridor borné, lui, est le choix de qui refuse de payer cher quoi que ce soit.
:::

## Choisir ses paramètres

Trois décisions, dans l'ordre d'importance.

**Le pourcentage w.** Même logique que le pourcentage fixe ([[pourcentage-fixe]]). La borne géométrique s'applique : w doit rester sous le rendement réel géométrique espéré pour un revenu médian stable. L'auto-correction autorise la même générosité qu'ailleurs, soit 4-4,5 % défendable là où le fixe exigerait 3,25-3,5 %. En marché cher, décotez ([[valorisations-et-cape]]) ou, mieux, laissez l'ancre CAPE juger votre w.

**Le couple de bornes.** (+5/−2,5) est le standard. Deux variantes sont utiles : (+4/−2) pour les budgets plus rigides, avec une descente encore plus douce et une ruine un peu plus haute ; (+6/−4) pour les budgets élastiques qui veulent suivre la cible de plus près. La règle de cohérence est simple : la borne basse doit rester tenable une fois composée sur plusieurs années. Une descente de −2,5 %/an pendant six ans fait −14 %, est-ce au-dessus de votre plancher réel ? C'est le test d'admissibilité de la famille, à vérifier sur la §04.

::: figure corridor-borne
Le test d'admissibilité, en une image. Chaque courbe est votre revenu après n années de baisse consécutive à la borne basse ; la ligne rouge est le plancher du ménage de l'exemple ci-dessous. Le standard Vanguard tient 7,8 ans avant de croiser ce plancher, la variante douce presque dix, la variante élastique moins de cinq. Le cercle rappelle ce que le millésime 1966 a réellement demandé au corridor : vingt-trois années de glisse d'affilée, jusqu'à 57 % du revenu de départ. **Testez la borne sur la durée d'un vrai régime hostile, pas sur six ans de politesse.**
:::

**L'interaction avec les revenus externes.** Comme toute la famille proportionnelle, la règle s'applique au portefeuille seul. En phase à découvert d'un FIRE, le pont de pension se provisionne à part ([[vpw]], [[horizon-et-esperance-de-vie]]). Une fois la pension au plancher, le corridor sur le portefeuille résiduel devient presque sans risque.

::: astuce La règle sur la page FIRE
La case « Bounded % of portfolio (Vanguard-style) » implémente exactement la règle : cible = pourcentage initial du portefeuille courant, variations réelles bornées à +5 %/−2,5 %, prime sur les règles flex, guardrails et ratchet. Deux lectures sont utiles. La §04 donne la distribution du revenu, où vous vérifiez le test de la borne basse composée. La frontière §06 vous la montre assise entre le fixe et le VPW. L'aide au survol rappelle honnêtement que la descente plafonnée peut laisser filer un effondrement : « unlike VPW/ABW this rule can still run out » ([[utiliser-la-page-fire]]).
:::

## Pour qui, face aux deux finalistes

Voici le profil du corridor borné. C'est le ménage qui veut d'abord la simplicité d'exécution : une multiplication, deux comparaisons, pas de seuils de ruine à surveiller, pas d'outil actuariel à faire tourner. Il veut un revenu qui ne surprend jamais, à ±2,5-5 % l'an, connus d'avance. Et il accepte d'être « deuxième partout ». C'est probablement la meilleure règle par défaut pour un conjoint survivant non gestionnaire ([[couple-et-famille]]), et une excellente règle de croisière pour la phase adossée.

Face aux finalistes, deux arbitrages. Contre les guardrails ([[guardrails-morningstar]]), le corridor échange les marches rares de ±10 %, avec leur charge émotionnelle de « décision », contre des glissements continus sans décision : moins optimal, mais plus vivable pour beaucoup. Contre l'ABW ([[amortissement-abw]]), il abandonne l'optimalité de consommation et la conscience de l'horizon, contre une gouvernance de carte postale. La synthèse générale de sélection reste [[choisir-sa-strategie]].

::: exemple Calibrer sa version en vingt minutes
Ménage : plancher réel 41 000 €, confort 50 000 €, portefeuille 1,3 M€ (w cible de 3,85 %), pension à 15 ans. Passons le test de la borne basse. Une traversée de six ans au plancher −2,5 % mène le revenu à 50 000 × 0,975^6 ≈ 43 100 €, au-dessus du plancher, donc (+5/−2,5) est admissible. En simulation, case Vanguard cochée, la ruine centrale ressort à 2,8 % (contre 6,1 % en fixe à 50 000 €). La §04 montre un pire quartile à −9 % pendant 5-7 ans, que le ménage accepte. La variante (+4/−2), testée ensuite, donne une ruine de 3,4 % et un pire quartile de −7 % : le ménage la préfère et le note dans son plan écrit. Vingt minutes, deux simulations, une règle possédée. À comparer aux heures de débat qu'exigent les guardrails par risque, pour un résultat proche dans la plupart des scénarios.
:::

## L'essentiel à retenir

- Deux lignées : le plancher-plafond de Bengen (bornes en niveau autour du retrait initial) et le corridor Vanguard (bornes en variation, +5 %/−2,5 % réels par an). Le second, sans mémoire morte, est devenu le standard.
- Le produit : des glissements au lieu de chutes, jamais plus de 2,5 % de baisse réelle par an. Le prix : la descente plafonnée « emprunte » au capital dans les crises, et la ruine redevient possible, faible mais honnête (comparable à celle du fixe).
- Deuxième partout, première nulle part, aucune pathologie : c'est la règle de qui refuse de payer cher une optimalité quelconque, et probablement le meilleur défaut pour un exécutant non spécialiste.
- Paramètres : w comme un pourcentage (borne géométrique, 4-4,5 % défendable), bornes (+5/−2,5) standard, et le test d'admissibilité. La borne basse composée sur six ans doit rester au-dessus du plancher réel.
- Reproduite telle quelle par un simulateur (« Bounded % of portfolio ») : vérifiez la §04 et la frontière, et comparez en deux clics aux guardrails et à l'ABW avant de choisir ([[choisir-sa-strategie]]).

---

## Pour aller plus loin

- Vanguard, « From assets to income: a goals-based approach to retirement spending » (recherche fondatrice de la règle, gratuite) et ses mises à jour.
- Bengen, *Conserving Client Portfolios During Retirement* (2006) : le floor-and-ceiling originel.
- Morningstar, *The State of Retirement Income* : la variante bornée dans le comparatif annuel ([[guardrails-morningstar]]).
- Dans ce livre : [[pourcentage-fixe]] (la matière première et la règle de Yale), [[guyton-klinger]] et [[guardrails-morningstar]] (l'autre voie du corridor), [[amortissement-abw]] (l'optimalité contre la simplicité).
