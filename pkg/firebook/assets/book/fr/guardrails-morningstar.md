# Les guardrails modernes (Morningstar) : l'état de l'art

La famille des garde-fous ne s'est pas arrêtée à Guyton-Klinger ([[guyton-klinger]]). Deux chantiers l'ont refondée depuis 2020 : les rapports annuels *The State of Retirement Income* de Morningstar, devenus la référence institutionnelle du domaine (c'est le rapport que citent la presse et les conseillers quand ils disent « le 4 % est devenu 3,7 % »), et les guardrails **par risque** de Michael Kitces et Derek Tharp, qui remplacent le capteur de 2006 (le taux de retrait courant) par le bon instrument : la probabilité de succès du plan, recalculée.

Le résultat cumulé est l'état de l'art de tout le retrait flexible : la stratégie que Morningstar classe première année après année sur le couple consommation totale / stabilité, et l'architecture que les praticiens sérieux déploient. Cet article présente les deux chantiers en détail : la méthodologie et les chiffres de Morningstar (dont l'histoire du taux recommandé, 3,3 % → 4,0 % → 3,7 %, qui est une leçon en soi), la mécanique exacte des guardrails par risque et pourquoi leur capteur est supérieur, leurs coûts réels (la dépendance au modèle, principalement), et la transposition pratique pour un lecteur français avec un simulateur comme capteur, car c'est très exactement l'usage pour lequel la page FIRE est taillée.

::: cle Ce qui a changé depuis 2006
Deux déplacements. Le **critère** : Morningstar juge les règles sur la consommation totale vécue et sa variabilité (plus le legs), pas sur le seul taux de succès : le critère qui avait condamné Guyton-Klinger ([[panorama-strategies-retrait]]). Le **capteur** : Kitces-Tharp déclenchent les ajustements sur la probabilité de succès recalculée du plan complet (horizon restant, pensions à venir, allocation), pas sur le ratio brut retrait/portefeuille, aveugle à tout cela. Même architecture qu'en 2006 (un corridor, des ajustements de ±10 %), mais un juge honnête et un thermomètre juste.
:::

## Morningstar : la référence annuelle, et ce qu'elle établit

Depuis 2021, l'équipe retraite de Morningstar (Christine Benz, Jeffrey Ptak, John Rekenthaler, puis Amy Arnott et Jason Kephart) publie chaque décembre un rapport qui refait le calcul du taux de retrait soutenable avec trois choix méthodologiques qui le distinguent de toute la littérature historique : des rendements **prospectifs** (les capital market assumptions de Morningstar Investment Management à 30 ans, [[rendements-attendus]]), pas les moyennes du passé ; un critère de 90 % de succès sur 30 ans (le standard praticien, plus tolérant que le pire-cas d'ERN, [[ruine-et-probabilites]]) ; et surtout la comparaison **systématique** des règles flexibles sur une grille à quatre dimensions : taux initial permis, consommation totale de la vie du plan, volatilité du revenu, legs final.

**La série des taux de base** vaut leçon d'épistémologie : 3,3 % (2021, valorisations extrêmes, taux nuls), 3,8 % (2022, actions dégonflées), 4,0 % (2023, rendements obligataires restaurés), ~3,7 % (2024-2025, actions redevenues chères). Le taux « sûr » n'est pas une constante. Il respire avec les conditions d'entrée ([[valorisations-et-cape]]), et une institution sérieuse assume de le republier. Retenez l'usage : le chiffre Morningstar de l'année est le meilleur « second avis » gratuit pour calibrer un plan de 30 ans ; pour un horizon FIRE, il se décote comme d'habitude ([[horizon-et-esperance-de-vie]]).

**Le classement des règles flexibles**, remarquablement stable d'une édition à l'autre : la règle fixe indexée est la plus coûteuse en consommation sacrifiée ; le simple **gel** d'indexation après année rouge ([[retrait-fixe-bengen]]) est le meilleur rapport bénéfice/simplicité ; le RMD (l'amortissement réglementaire américain, cousin du VPW [[vpw]]) maximise la consommation au prix d'une forte variabilité ; et les **guardrails arrivent premiers au classement combiné** : le taux initial permis le plus élevé (souvent 5 % et plus dans leurs éditions), la consommation totale la plus haute à variabilité tolérable, au prix du legs le plus faible et de la complexité. La version Morningstar des guardrails est un Guyton-Klinger assagi : corridor de ±20 % sur le taux courant, ajustements de 10 %, **mais** coupes plafonnées en fréquence et hausses freinées : les leçons de la pathologie de 2006 sont intégrées ([[guyton-klinger]]).

L'honnêteté oblige à répéter le garde-fou de lecture : ce classement porte sur un retraité de 65 ans, 30 ans d'horizon, avec la pension américaine en toile de fond. Le FIRE à 45 ans transpose l'**architecture**, pas les taux : un guardrail à 5 % initial sur 50 ans en marché cher reste une imprudence, quelle que soit la qualité du garde-fou ([[flexibilite-realite]]).

## Kitces-Tharp : changer le capteur

Le second chantier attaque le défaut résiduel de toute la famille : le taux de retrait courant est un **mauvais** thermomètre. Il ignore l'horizon restant (6 % de taux courant à 85 ans est sain ; à 52 ans, il est grave), les flux futurs (le même 6 % deux ans avant la liquidation des pensions est bénin, [[revenus-complementaires]]), l'allocation et les valorisations. Résultat : le guardrail de 2006 coupe des retraités qui n'en avaient pas besoin et rassure des plans déjà condamnés.

La proposition de Kitces et Tharp (développée sur [kitces.com](https://www.kitces.com) et industrialisée dans des outils comme Income Lab) : **surveiller directement la probabilité de succès du plan complet**, recalculée avec un simulateur, et poser les garde-fous dessus. L'architecture type :

- **La cible** : maintenir le plan autour de ~80-90 % de succès (soit 10-20 % de ruine simulée, la zone de travail praticienne, [[ruine-et-probabilites]]).
- **Le garde-fou bas** : si la probabilité de succès tombe sous ~70-75 % (le portefeuille a souffert, ou les hypothèses se sont dégradées), couper les dépenses de ~10 %, ce qui la remonte au-dessus de la cible.
- **Le garde-fou haut** : si elle dépasse ~99 % (le plan est en train de mourir riche), augmenter de ~10 %.
- **La revue** : annuelle à date fixe, ou déclenchée par franchissement ([[revue-annuelle]], [[quand-s-inquieter]]).

La supériorité du capteur tient en deux exemples. Le retraité de 62 ans dont la pension arrive à 64 : un krach fait bondir son taux courant (le guardrail 2006 coupe), mais sa probabilité de succès bouge à peine (deux ans de pont à financer, le reste est adossé) : le guardrail par risque, lui, ne coupe pas. Il a raison. Inversement, le FIRE de 48 ans en marché très cher : son taux courant est nominal (le guardrail 2006 dort), mais son simulateur, ancré aux valorisations ([[regles-cape]]), voit la probabilité de succès glisser : l'alerte précoce arrive des années avant le ratio brut. Le capteur par risque intègre **toute** l'information que ce livre a construite (séquence, valorisations, horizon, flux) parce qu'il est le simulateur.

::: attention Le prix du bon capteur : la dépendance au modèle
Le thermomètre par risque hérite de tout ce que le simulateur a de fragile ([[pieges-des-simulateurs]]) : une probabilité de succès calculée sur un modèle gaussien-optimiste dira « tout va bien » jusque dans le mur ; une calculée sur le broad-sample coupera peut-être trop tôt. Les faux signaux existent aussi : la probabilité de succès **bouge** avec les marchés et les mises à jour d'hypothèses, et un couple de garde-fous trop serrés (75/95) fera yo-yo. Les parades : un capteur multi-modèles (déclencher sur la ruine du modèle **central**, vérifier sur les colonnes dures, exactement la lecture en faisceau), des bandes larges, une hystérésis (ne réagir qu'à un franchissement confirmé deux revues de suite), et des ajustements doux. Le guardrail par risque est l'état de l'art **pour qui** a un simulateur honnête et la discipline de s'en servir ; sans cela, le vieux ratio brut plus un plancher reste défendable ([[guyton-klinger]]).
:::

## La transposition française, avec un simulateur comme capteur

Assemblons le tout pour un lecteur de ce livre : la version exécutable des guardrails par risque, le simulateur en instrument.

**À la conception.** Fixez le trio : la ruine cible du modèle central (par exemple 5 %, le défaut du solveur), le garde-fou bas (ruine centrale > 12-15 %, à caler pour que la coupe de 10 % suffise à revenir sous la cible, le solveur §09 le vérifie en un clic), le garde-fou haut (ruine < 1 % et taux courant très bas, la condition du cliquet). Écrivez les trois chiffres, l'ampleur des ajustements (±10 % des dépenses de confort, jamais sous le plancher, [[combien-il-vous-faut]]), et l'hystérésis (deux revues consécutives).

**En exploitation.** Une fois par an, à date fixe : mettre à jour capital, dépenses réelles, pensions ; relancer la page ; lire la ruine centrale et les colonnes dures. Trois cas : dans le corridor, ne rien faire (le cas normal, c'est une vertu, pas une déception) ; sous le garde-fou haut deux ans de suite, appliquer la hausse (ou laisser le cliquet natif le faire) ; au-dessus du garde-fou bas deux ans de suite, appliquer la coupe écrite. Entre deux revues : rien, sauf franchissement massif ([[quand-s-inquieter]]).

**Les approximations natives**, pour qui veut la version simulée en continu plutôt que la procédure annuelle : le curseur « Also cut above this WR » (coupe déclenchée par le taux courant, le capteur 2006, utile en simulation pour chiffrer la famille) et la case guardrails avec plancher ([[guyton-klinger]]) donnent dans la §04 et la frontière §06 une excellente image de ce que **votre** version des garde-fous ferait vivre : réglez-les sur vos paramètres écrits et regardez le pire quartile avant de signer.

::: exemple Des guardrails par risque, dimensionnés en séance
Plan : 1,5 M€, confort 54 000 €, plancher 44 000 €, 45 ans, pension 16 000 € en année 18. Séance de conception : au confort, ruine centrale 5 % (cible OK), broad-sample 11 % (accepté, échecs tardifs, pension au plancher). Test du garde-fou bas. On simule le plan après un krach précoce (capital 1,1 M€ en année 3) : ruine centrale 14 %. C'est le seuil ; la coupe écrite (−10 % → 48 600 €) la ramène à 7 % : l'ajustement suffit, le garde-fou est bien placé. Test du haut : à 2,1 M€ (année 8 faste), ruine 0,6 % : la hausse de 10 % la porte à 1,1 % : accordée. Le plan tient sur une demi-page : trois seuils, deux ajustements, une date de revue. Et chaque chiffre a été **vérifié** dans les deux sens plutôt que recopié d'un article américain. C'est toute la différence entre appliquer une règle et posséder la sienne.
:::

## Ce que ça change, et ce que ça ne change pas

Les guardrails modernes sont l'aboutissement de la famille « fixe qui plie », et à ce titre le meilleur choix par défaut pour le ménage qui veut un revenu **stable** la plupart du temps avec une sécurité pilotée. Ils ne changent pas les lois de la physique : la flexibilité achète toujours ~0,3-0,5 point de taux au-dessus du fixe équivalent, pas un point et demi ([[flexibilite-realite]]) ; le plancher reste la vraie frontière de l'admissibilité ; et dans la comparaison finale, la famille actuarielle ([[amortissement-abw]]) garde ses avantages propres (pas de falaise du tout, consommation supérieure) contre une variabilité assumée. L'arbitrage entre les deux familles gagnantes est l'objet de [[choisir-sa-strategie]].

## L'essentiel à retenir

- Deux refondations : Morningstar (le juge honnête, consommation vécue + variabilité + legs, rendements prospectifs, rapport annuel dont la série 3,3 → 4,0 → 3,7 % enseigne que le taux respire) et Kitces-Tharp (le bon thermomètre, la probabilité de succès du plan complet, pas le ratio brut).
- Au classement Morningstar, les guardrails assagis dominent le couple consommation/stabilité ; le simple gel d'indexation reste le meilleur rapport bénéfice/simplicité ; tout se transpose en architecture, jamais en taux, pour un horizon FIRE.
- Le capteur par risque intègre horizon, pensions et valorisations. Il ne coupe pas le retraité dont la pension arrive, et alerte le FIRE en marché cher des années avant le ratio. Mais il hérite des fragilités du simulateur : multi-modèles, bandes larges, hystérésis et ajustements doux obligatoires.
- Version exécutable en simulation : trois seuils de ruine écrits (cible ~5 %, coupe >12-15 %, hausse <1 %), ±10 % jamais sous le plancher, revue annuelle, confirmation sur deux revues. Et chaque seuil **vérifié** au solveur dans les deux sens.
- La famille est l'état de l'art du revenu stable piloté ; son concurrent final est l'amortissement ([[amortissement-abw]]) : rendez-vous dans [[choisir-sa-strategie]].

---

## Pour aller plus loin

- Morningstar, *The State of Retirement Income* (annuel, gratuit sur [morningstar.com](https://www.morningstar.com)) : le rapport, sa méthodologie et le comparatif des règles.
- Kitces & Tharp : « Probability-Of-Success-Driven Guardrails », et la série guardrails de [kitces.com](https://www.kitces.com) ; Income Lab pour l'industrialisation.
- Guyton & Klinger (2006) pour l'ancêtre, et ERN volets 9-11 pour la critique qui a rendu ces refontes nécessaires ([[serie-ern]], [[guyton-klinger]]).
- Dans pofo : le solveur §09 (dimensionner les seuils), la §04 (le pire quartile vécu), la frontière §06 ([[utiliser-la-page-fire]]).
