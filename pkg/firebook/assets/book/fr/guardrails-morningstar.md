# Les guardrails modernes (Morningstar) : l'état de l'art

La famille des garde-fous ne s'est pas arrêtée à Guyton-Klinger ([[guyton-klinger]]). Deux chantiers l'ont refondée depuis 2020. Le premier : les rapports annuels *The State of Retirement Income* de Morningstar, devenus la référence institutionnelle du domaine. C'est le rapport que citent la presse et les conseillers quand ils disent « le 4 % est devenu 3,7 % ». Le second : les guardrails **par risque** de Michael Kitces et Derek Tharp. Ils remplacent le capteur de 2006, le taux de retrait courant, par le bon instrument : la probabilité de succès du plan, recalculée.

Le résultat cumulé est l'état de l'art du retrait flexible. C'est la stratégie que Morningstar classe première année après année sur le couple consommation totale / stabilité, et l'architecture que déploient les praticiens sérieux. Cet article déroule les deux chantiers en détail. D'abord la méthodologie et les chiffres de Morningstar, dont l'histoire du taux recommandé (3,3 % → 4,0 % → 3,7 %), une leçon en soi. Ensuite la mécanique exacte des guardrails par risque, et pourquoi leur capteur est supérieur. Puis leurs coûts réels, la dépendance au modèle surtout. Puis un déroulé de sept ans, table et graphique à l'appui, pour voir ce que la règle fait vivre, le rythme des décisions et l'ampleur des variations. Enfin la transposition pratique pour un lecteur français, un simulateur en guise de capteur, l'usage précis pour lequel la page FIRE est taillée.

::: cle Ce qui a changé depuis 2006
Deux déplacements. Le **critère** d'abord : Morningstar juge les règles sur la consommation totale vécue et sa variabilité (le legs compris), pas sur le seul taux de succès. C'est le critère qui avait condamné Guyton-Klinger ([[panorama-strategies-retrait]]). Le **capteur** ensuite : Kitces-Tharp déclenchent les ajustements sur la probabilité de succès recalculée du plan complet (horizon restant, pensions à venir, allocation), pas sur le ratio brut retrait/portefeuille, aveugle à tout cela. Même architecture qu'en 2006, un corridor et des ajustements de ±10 %, mais un juge honnête et un thermomètre juste.
:::

## Morningstar : la référence annuelle, et ce qu'elle établit

Depuis 2021, l'équipe retraite de Morningstar (Christine Benz, Jeffrey Ptak, John Rekenthaler, puis Amy Arnott et Jason Kephart) publie chaque décembre un rapport qui refait le calcul du taux de retrait soutenable. Trois choix méthodologiques le distinguent de toute la littérature historique. D'abord des rendements prospectifs, les capital market assumptions de Morningstar Investment Management à 30 ans ([[rendements-attendus]]), et non les moyennes du passé. Ensuite un critère de 90 % de succès sur 30 ans, le standard praticien, plus tolérant que le pire-cas d'ERN ([[ruine-et-probabilites]]). Enfin, et surtout, la comparaison systématique des règles flexibles sur une grille à quatre dimensions : taux initial permis, consommation totale sur la vie du plan, volatilité du revenu et legs final.

**La série des taux de base** vaut une leçon d'épistémologie : 3,3 % (2021, valorisations extrêmes, taux nuls), 3,8 % (2022, actions dégonflées), 4,0 % (2023, rendements obligataires restaurés), ~3,7 % (2024-2025, actions redevenues chères). Le taux « sûr » n'est pas une constante. Il respire avec les conditions d'entrée ([[valorisations-et-cape]]), et une institution sérieuse assume de le republier. D'où l'usage à retenir. Le chiffre Morningstar de l'année est le meilleur « second avis » gratuit pour calibrer un plan de 30 ans. Pour un horizon FIRE, il se décote comme d'habitude ([[horizon-et-esperance-de-vie]]).

**Le classement des règles flexibles** est remarquablement stable d'une édition à l'autre. La règle fixe indexée est la plus coûteuse en consommation sacrifiée. Le simple gel d'indexation après une année rouge ([[retrait-fixe-bengen]]) offre le meilleur rapport bénéfice/simplicité. Le RMD, l'amortissement réglementaire américain, cousin du VPW ([[vpw]]), maximise la consommation au prix d'une forte variabilité. Et les **guardrails arrivent premiers au classement combiné** : le taux initial permis le plus élevé (souvent 5 % et plus dans leurs éditions), la consommation totale la plus haute à variabilité tolérable, au prix du legs le plus faible et de la complexité. La version Morningstar des guardrails est un Guyton-Klinger assagi. Corridor de ±20 % sur le taux courant, ajustements de 10 %, mais coupes plafonnées en fréquence et hausses freinées. Les leçons de la pathologie de 2006 sont intégrées ([[guyton-klinger]]).

L'honnêteté oblige à répéter le garde-fou de lecture : ce classement porte sur un retraité de 65 ans, 30 ans d'horizon, la pension américaine en toile de fond. Le FIRE à 45 ans transpose l'**architecture**, pas les taux. Un guardrail à 5 % initial sur 50 ans en marché cher reste une imprudence, quelle que soit la qualité du garde-fou ([[flexibilite-realite]]).

## Kitces-Tharp : changer le capteur

Le second chantier attaque le défaut résiduel de toute la famille : le taux de retrait courant est un **mauvais** thermomètre. Il ignore l'horizon restant (6 % de taux courant à 85 ans est sain, à 52 ans il est grave). Il ignore les flux futurs (le même 6 % deux ans avant la liquidation des pensions est bénin, [[revenus-complementaires]]). Il ignore l'allocation et les valorisations. Résultat : le guardrail de 2006 coupe des retraités qui n'en avaient pas besoin et rassure des plans déjà condamnés.

La proposition de Kitces et Tharp, développée sur [kitces.com](https://www.kitces.com) et industrialisée dans des outils comme Income Lab, tient en une phrase : **surveiller directement la probabilité de succès du plan complet**, recalculée avec un simulateur, et poser les garde-fous dessus. C'est, au fond, une mise à jour bayésienne du plan. Chaque année de marchés et de dépenses vécus révise la croyance sur sa réussite, et la règle agit sur cette croyance révisée plutôt que sur un chiffre figé à l'an 1. L'architecture type :

- **La cible** : maintenir le plan autour de ~80-90 % de succès (soit 10-20 % de ruine simulée, la zone de travail praticienne, [[ruine-et-probabilites]]).
- **Le garde-fou bas** : si la probabilité de succès tombe sous ~70-75 % (le portefeuille a souffert, ou les hypothèses se sont dégradées), couper les dépenses de ~10 %, ce qui la remonte au-dessus de la cible.
- **Le garde-fou haut** : si elle dépasse ~99 % (le plan est en train de mourir riche), augmenter de ~10 %.
- **La revue** : annuelle à date fixe, ou déclenchée par franchissement ([[revue-annuelle]], [[quand-s-inquieter]]).

La supériorité du capteur tient en deux exemples. Prenez le retraité de 62 ans dont la pension arrive à 64. Un krach fait bondir son taux courant, donc le guardrail 2006 coupe. Mais sa probabilité de succès bouge à peine : deux ans de pont à financer, le reste est adossé. Le guardrail par risque, lui, ne coupe pas, et il a raison. Prenez à l'inverse le FIRE de 48 ans en marché très cher. Son taux courant reste modéré, donc le guardrail 2006 dort. Mais son simulateur, ancré aux valorisations ([[regles-cape]]), voit la probabilité de succès glisser : l'alerte précoce arrive des années avant le ratio brut. Le capteur par risque intègre toute l'information que ce livre a construite (séquence, valorisations, horizon, flux), parce qu'il est le simulateur.

::: attention Le prix du bon capteur : la dépendance au modèle
Le thermomètre par risque hérite de tout ce que le simulateur a de fragile ([[pieges-des-simulateurs]]). Une probabilité de succès calculée sur un modèle gaussien-optimiste dira « tout va bien » jusque dans le mur. Une probabilité calculée sur le broad-sample coupera peut-être trop tôt. Les faux signaux existent aussi. La probabilité de succès bouge avec les marchés et les mises à jour d'hypothèses, et un couple de garde-fous trop serrés (75/95) fera yo-yo. Les parades sont connues. Un capteur multi-modèles, qui déclenche sur la ruine du modèle central et vérifie sur les colonnes dures, exactement la lecture en faisceau. Des bandes larges. Une hystérésis, qui ne réagit qu'à un franchissement confirmé deux revues de suite. Et des ajustements doux. Le guardrail par risque est l'état de l'art pour qui a un simulateur honnête et la discipline de s'en servir. Sans cela, le vieux ratio brut assorti d'un plancher reste défendable ([[guyton-klinger]]).
:::

## Sept ans avec un guardrail par risque

Tout ce qui précède reste abstrait tant qu'on n'a pas vu la règle vivre. Rien ne vaut un déroulé, comme pour le corridor Vanguard ([[plancher-plafond]]). Prenons le plan qui servira d'exemple plus bas : 1,5 M€, retrait de confort 54 000 € (3,6 %), plancher 44 000 €, pension en année 18. Le corridor est écrit ainsi : coupe de 10 % si le succès simulé tombe sous 85 %, hausse de 10 % s'il dépasse 99 %, tout franchissement devant être confirmé par deux revues consécutives (l'hystérésis). La séquence est un marché baissier précoce suivi d'une reprise, et les valeurs du capteur sont illustratives.

| Revue | Portefeuille (réel) | Succès lu | Décision (corridor 85-99, hystérésis) | Retrait servi (réel) |
|---|---|---|---|---|
| 1 | 1 500 000 | 93 % | dans le corridor : rien | 54 000 |
| 2 | 1 150 000 (−20 %) | 82 % | 1re alerte basse : on attend | 54 000 |
| 3 | 990 000 (−10 %) | 76 % | alerte confirmée : coupe −10 % | **48 600** |
| 4 | 1 060 000 (+12 %) | 88 % | retour dans le corridor : rien | 48 600 |
| 5 | 1 120 000 (+10 %) | 91 % | corridor : rien | 48 600 |
| 6 | 1 340 000 (+24 %) | 99,2 % | 1re alerte haute : on attend | 48 600 |
| 7 | 1 430 000 (+11 %) | 99,4 % | confirmée : hausse +10 % | **53 460** |

::: figure guardrails-capteur
Les sept revues de la table, en deux panneaux alignés. En haut, le capteur : la probabilité de succès recalculée traverse le corridor 85-99 %, et seuls les franchissements confirmés deux revues de suite déclenchent un ajustement (les cercles vides sont des premières alertes, mises en attente). En bas, le revenu : deux marches de ±10 % en sept ans, loin du plancher. Le cas normal est l'immobilité.
:::

**Lisez le rythme d'abord.** Sept revues, deux décisions. Le cas normal, cinq années sur sept, est « ne rien faire », et c'est voulu : un guardrail bien dimensionné est silencieux la plupart du temps ([[revue-annuelle]]). L'hystérésis a coûté un an de retard sur la coupe, puisque l'année 2 alertait déjà. Mais c'est son office. Si l'année 3 avait rebondi, l'alerte se serait éteinte sans décision, et le revenu n'aurait jamais bougé. On échange un peu de réactivité contre l'absence de yo-yo.

**Lisez les ampleurs ensuite.** C'est la réponse à « qu'est-ce que ça fait vivre ? ». Dans une traversée sévère (−28 % réel au creux), le revenu a fait une seule marche, de −10 %, tenue quatre ans, avant de remonter. Jamais il ne s'est approché du plancher (48 600 contre 44 000). Le pourcentage pur aurait servi −28 % en deux ans ([[pourcentage-fixe]]) ; le fixe indexé n'aurait rien changé du tout, en laissant monter la ruine silencieuse ([[retrait-fixe-bengen]]). Les marches de ±10 %, rares et datées, sont la signature de la famille : des décisions peu fréquentes mais réelles, là où le corridor Vanguard préfère des glissements continus sans décision ([[plancher-plafond]]).

**Lisez enfin ce que le capteur voit.** Entre les revues 3 et 6, le succès simulé remonte de 76 à 99 %. Les marchés n'expliquent qu'une partie du chemin. S'y ajoutent la coupe elle-même, moins de dépenses donc plus de succès, et un facteur que le ratio brut ne verra jamais : chaque année écoulée raccourcit l'horizon restant et rapproche la pension de l'année 18. Le temps travaille pour le capteur. C'est exactement l'information que le thermomètre de 2006 ignorait, et la raison d'être de la refonte.

## La transposition française, avec un simulateur comme capteur

Assemblons le tout pour un lecteur de ce livre : la version exécutable des guardrails par risque, le simulateur en guise d'instrument.

**À la conception.** Fixez le trio de seuils. La ruine cible du modèle central, par exemple 5 %, le défaut du solveur. Le garde-fou bas, une ruine centrale supérieure à 12-15 %, à caler pour que la coupe de 10 % suffise à repasser sous la cible (le solveur §09 le vérifie en un clic). Le garde-fou haut, une ruine sous 1 % et un taux courant très bas, la condition du cliquet. Écrivez ensuite les trois chiffres, l'ampleur des ajustements (±10 % des dépenses de confort, jamais sous le plancher, [[combien-il-vous-faut]]) et l'hystérésis (deux revues consécutives).

**En exploitation.** Une fois par an, à date fixe, faites trois gestes : mettre à jour le capital, les dépenses réelles et les pensions, relancer la page, lire la ruine centrale et les colonnes dures. Trois cas se présentent. Dans le corridor, ne rien faire : le cas normal est une vertu, pas une déception. Sous le garde-fou haut deux ans de suite, appliquer la hausse, ou laisser le cliquet natif la faire. Au-dessus du garde-fou bas deux ans de suite, appliquer la coupe écrite. Entre deux revues, rien, sauf franchissement massif ([[quand-s-inquieter]]).

**Les approximations natives du simulateur** servent à qui veut la version simulée en continu plutôt que la procédure annuelle. Le curseur « Also cut above this WR » applique une coupe déclenchée par le taux courant, le capteur 2006, utile en simulation pour chiffrer la famille. La case guardrails avec plancher ([[guyton-klinger]]) complète le tableau. Dans la §04 et la frontière §06, les deux donnent une excellente image de ce que votre version des garde-fous ferait vivre. Réglez-les sur vos paramètres écrits, puis regardez le pire quartile avant de signer.

::: exemple Des guardrails par risque, dimensionnés en séance
Plan : 1,5 M€, confort 54 000 €, plancher 44 000 €, 45 ans, pension de 16 000 € en année 18. La séance de conception commence au confort : ruine centrale 5 % (cible OK), broad-sample 11 % (accepté, échecs tardifs, pension au plancher). Vient le test du garde-fou bas. On simule le plan après un krach précoce, capital tombé à 1,1 M€ en année 3 : ruine centrale 14 %. C'est le seuil. La coupe écrite (−10 % → 48 600 €) la ramène à 7 %, donc l'ajustement suffit et le garde-fou est bien placé. Test du haut, à 2,1 M€ (année 8 faste), ruine 0,6 %. La hausse de 10 % la porte à 1,1 %, accordée. Le plan tient sur une demi-page : trois seuils, deux ajustements, une date de revue. Et chaque chiffre a été **vérifié** dans les deux sens, plutôt que recopié d'un article américain. C'est toute la différence entre appliquer une règle et posséder la sienne.
:::

## Ce que ça change, et ce que ça ne change pas

Les guardrails modernes sont l'aboutissement de la famille « fixe qui plie ». À ce titre, ils forment le meilleur choix par défaut pour le ménage qui veut un revenu **stable** la plupart du temps, avec une sécurité pilotée. Ils ne changent pas pour autant les lois de la physique. La flexibilité achète toujours ~0,3-0,5 point de taux au-dessus du fixe équivalent, pas un point et demi ([[flexibilite-realite]]). Le plancher reste la vraie frontière de l'admissibilité. Et dans la comparaison finale, la famille actuarielle ([[amortissement-abw]]) garde ses avantages propres, l'absence totale de falaise et une consommation supérieure, contre une variabilité assumée. L'arbitrage entre les deux familles gagnantes est l'objet de [[choisir-sa-strategie]].

## L'essentiel à retenir

- Deux refondations : Morningstar (le juge honnête, consommation vécue + variabilité + legs, rendements prospectifs, rapport annuel dont la série 3,3 → 4,0 → 3,7 % enseigne que le taux respire) et Kitces-Tharp (le bon thermomètre, la probabilité de succès du plan complet, pas le ratio brut).
- Au classement Morningstar, les guardrails assagis dominent le couple consommation/stabilité ; le simple gel d'indexation reste le meilleur rapport bénéfice/simplicité ; tout se transpose en architecture, jamais en taux, pour un horizon FIRE.
- Le capteur par risque intègre horizon, pensions et valorisations. Il ne coupe pas le retraité dont la pension arrive, et alerte le FIRE en marché cher des années avant le ratio. Mais il hérite des fragilités du simulateur : multi-modèles, bandes larges, hystérésis et ajustements doux obligatoires.
- Concrètement, le rythme est lent : des années de « rien », puis une marche de ±10 % confirmée par deux revues. Dans le déroulé type, une traversée à −28 % se vit comme une seule coupe de 10 % tenue quatre ans, loin du plancher, le capteur remontant ensuite par trois canaux (les marchés, la coupe, l'horizon qui raccourcit).
- Version exécutable en simulation : trois seuils de ruine écrits (cible ~5 %, coupe >12-15 %, hausse <1 %), ±10 % jamais sous le plancher, revue annuelle, confirmation sur deux revues. Et chaque seuil **vérifié** au solveur dans les deux sens.
- La famille est l'état de l'art du revenu stable piloté ; son concurrent final est l'amortissement ([[amortissement-abw]]) : rendez-vous dans [[choisir-sa-strategie]].

---

## Pour aller plus loin

- Morningstar, *The State of Retirement Income* (annuel, gratuit sur [morningstar.com](https://www.morningstar.com)) : le rapport, sa méthodologie et le comparatif des règles.
- Kitces & Tharp : « Probability-Of-Success-Driven Guardrails », et la série guardrails de [kitces.com](https://www.kitces.com) ; Income Lab pour l'industrialisation.
- Guyton & Klinger (2006) pour l'ancêtre, et ERN volets 9-11 pour la critique qui a rendu ces refontes nécessaires ([[serie-ern]], [[guyton-klinger]]).
- Dans la page FIRE : le solveur §09 (dimensionner les seuils), la §04 (le pire quartile vécu), la frontière §06 ([[utiliser-la-page-fire]]).
