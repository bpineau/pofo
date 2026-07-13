# Le FIRE, c'est quoi ?

FIRE : *Financial Independence, Retire Early*. Indépendance financière, retraite précoce.

Derrière l'acronyme, une idée simple et ancienne : accumuler un capital suffisant pour que son rendement couvre vos dépenses, puis vivre de ce capital, avec ou sans travail, à l'âge que vous choisissez plutôt qu'à celui que fixe la loi. Cette page pose la carte du sujet : d'où vient le mouvement, ce que recouvrent ses variantes (Lean, Fat, Barista, Coast), les ordres de grandeur qui structurent tout le reste, et ce que ce livre va vous apprendre à faire.

À la fin, vous saurez situer votre propre projet sur cette carte et par quel article continuer.

::: cle Le principe en une phrase
Quand votre patrimoine investi atteint environ 25 à 33 fois vos dépenses annuelles, son rendement réel (après inflation) peut, avec une forte probabilité mais jamais une certitude, financer votre vie indéfiniment. Tout le reste du sujet, et tout ce livre, consiste à comprendre, affiner, sécuriser et vivre cette phrase.
:::

## D'où ça vient

L'idée de vivre de ses rentes est aussi vieille que le capital, mais le mouvement FIRE moderne a une généalogie précise.

**1992 : *Your Money or Your Life***. Le livre de Vicki Robin et Joe Dominguez pose le socle philosophique : l'argent est de l'énergie vitale échangée, chaque dépense s'évalue en heures de vie, et l'indépendance financière est le point où les revenus du capital croisent les dépenses. Pas encore de taux de retrait, mais déjà le graphique fondateur : deux courbes, dépenses et revenus passifs, et le croisement qui libère.

**1994 : William Bengen**. Un conseiller financier californien publie dans le *Journal of Financial Planning* l'article qui invente le concept de taux de retrait sûr : en rejouant toutes les retraites historiques américaines depuis 1926, il montre qu'un retrait initial de 4 % du capital, réindexé sur l'inflation chaque année, a survécu à toutes les périodes de 30 ans, y compris pour ceux qui ont pris leur retraite en 1929 ou en 1966. C'est la naissance de la règle des 4 % ([[etude-trinity]]).

**1998 : l'étude Trinity** confirme et popularise le résultat avec une grille de probabilités de succès par taux de retrait, allocation et horizon.

**2011 : Mr. Money Mustache**. Le blog de Pete Adeney transforme une littérature de planification financière en mouvement culturel : un ingénieur retraité à 30 ans, un ton provocateur, et une arithmétique limpide reliant taux d'épargne et années de travail restantes (voir le tableau plus bas). La communauté FIRE explose, forums, blogs, podcasts, subreddits.

**2016 et après : la maturité scientifique**. Le blog *Early Retirement Now* de Karsten Jeske ([[serie-ern]]) reprend le sujet avec la rigueur d'un économiste et démonte les simplifications de la première vague : la règle des 4 % n'est pas adaptée aux retraites de 50 ans, les valorisations de départ comptent énormément, la flexibilité a des limites. En parallèle, la recherche académique élargit l'échantillon au-delà des États-Unis ([[anarkulova-cederburg]]) et les praticiens institutionnels (Morningstar, Vanguard, fonds de pension) publient leurs propres cadres ([[guardrails-morningstar]], [[rendements-attendus]]).

::: encart Où en est le mouvement aujourd'hui
Le FIRE de 2026 n'est plus celui de 2012. La première vague vendait une certitude (« 4 %, c'est réglé ») ; l'état de l'art actuel vend une méthode : modéliser honnêtement l'incertitude, choisir une stratégie de retrait adaptative, construire un portefeuille qui résiste aux régimes hostiles, et garder des marges. C'est exactement le programme de ce livre.
:::

## Les variantes : Lean, Fat, Barista, Coast

Le mouvement a produit un vocabulaire pour désigner des projets très différents qui partagent la même mécanique.

| Variante | L'idée | Ordre de grandeur (dépenses) | Le profil type |
|---|---|---|---|
| **Lean FIRE** | Frugalité assumée, capital minimal | 15 000 à 25 000 €/an | Célibataire ou couple sobre, logement payé ou pays à bas coût |
| **FIRE « classique »** | Vie normale sans travail | 25 000 à 45 000 €/an | Ménage médian discipliné |
| **Fat FIRE** | Confort sans compromis | 60 000 €/an et bien au-delà | Hauts revenus, entrepreneurs, cadres tardifs |
| **Barista FIRE** | Capital partiel + petit boulot choisi | Le capital couvre 50 à 80 % des dépenses | Quitte le job principal tôt, complète par une activité plaisir |
| **Coast FIRE** | Le capital est déjà suffisant... à terme | On ne retire rien : on laisse composer | A « fini d'épargner » à 35 ans, travaille pour ses dépenses courantes jusqu'à 50 ou 60 ans |

Deux remarques importantes sur ce tableau.

D'abord, **la difficulté n'est pas linéaire**. Passer de Lean à Fat ne multiplie pas seulement le capital cible par trois : cela change la nature du problème. Un Lean FIRE à 20 000 €/an dispose de leviers énormes (géo-arbitrage, retour ponctuel au travail, [[retour-au-travail]]) parce qu'un SMIC couvre ses dépenses. Un Fat FIRE à 100 000 €/an ne peut pas « se rattraper » par un petit boulot : sa sécurité doit être entièrement dans le plan.

Ensuite, **Barista et Coast sont des amortisseurs de risque de séquence**, pas seulement des styles de vie. Quelques années de revenus partiels au début de la retraite réduisent drastiquement les retraits pendant la fenêtre la plus dangereuse ([[sequence-des-rendements]]) : c'est mathématiquement l'un des outils les plus puissants du sujet, on y reviendra dans [[revenus-complementaires]].

## L'arithmétique qui rend le FIRE possible

Le moteur du FIRE n'est pas le rendement, c'est le **taux d'épargne**, pour une raison à double détente : épargner plus accélère l'accumulation **et** abaisse la cible (vous prouvez que vous vivez avec moins). D'où le tableau célèbre popularisé par Mr. Money Mustache (hypothèses : 5 % de rendement réel, retrait à 4 %, départ de zéro) :

| Taux d'épargne | Années de travail avant l'indépendance |
|---|---|
| 10 % | ~51 ans |
| 20 % | ~37 ans |
| 35 % | ~25 ans |
| 50 % | ~17 ans |
| 65 % | ~10,5 ans |
| 80 % | ~5,5 ans |

::: exemple Un couple concret
Léa et Sam, 32 ans, gagnent 5 400 € nets par mois à deux et en dépensent 3 200. Taux d'épargne : 41 %. Cible à 4 % : 3 200 × 12 × 25 = 960 000 €. Avec 90 000 € déjà investis et 2 200 €/mois d'épargne à 5 % réel, ils atteignent la cible vers 53 ans ; à 4 % réel, vers 55 ans. S'ils réduisent leurs dépenses de 400 €/mois, la cible tombe à 840 000 € **et** l'épargne monte à 2 600 €/mois : l'indépendance avance d'environ 4 ans. C'est la double détente. Pour aller plus loin sur le calcul de la cible : [[combien-il-vous-faut]].
:::

::: attention Le rendement ne vous sauvera pas
La tentation classique du débutant : compenser une épargne faible par un portefeuille « agressif ». Sur 15 ans d'accumulation, passer de 5 % à 7 % de rendement espéré gagne 2 à 3 ans... quand tout va bien, et expose à des séquences bien pires. Passer de 20 % à 35 % d'épargne gagne 12 ans, sans aléa. L'ordre des priorités est sans ambiguïté : dépenses, puis revenus, puis seulement l'optimisation du portefeuille.
:::

## Ce que le FIRE n'est pas

Quelques mises au point que la première vague du mouvement a parfois laissées dans le flou.

**Ce n'est pas « ne plus jamais travailler ».** La majorité des FIRE réels ont des revenus après leur départ : projets devenus lucratifs, conseil ponctuel, passion monétisée. Non parce que le plan a échoué, mais parce que des gens capables d'épargner 50 % de leurs revenus pendant 15 ans sont rarement du genre à s'arrêter de produire ([[sens-et-identite]]).

**Ce n'est pas une garantie.** Tout taux de retrait est une probabilité, pas une promesse. « 95 % de succès historique » signifie que dans 1 cas défavorable sur 20 du passé américain, le plan échouait, et le futur n'est pas tenu de ressembler au passé américain ([[pieges-des-simulateurs]]). Le vrai livrable d'un plan FIRE n'est pas un chiffre, c'est un chiffre **plus** une stratégie d'ajustement ([[panorama-strategies-retrait]]) **plus** des marges ([[cash-buffer]]).

**Ce n'est pas réservé aux ingénieurs américains à 200 000 $ par an.** Le cadre s'applique à tout niveau de dépenses ; c'est le délai qui change. Et le lecteur français dispose d'atouts spécifiques (des enveloppes fiscales efficaces, [[enveloppes-francaises]], une retraite légale qui finira par arriver en soutien, [[retraite-legale]]) et de pièges spécifiques ([[taxe-puma]]).

**Ce n'est pas de la privation.** Le mouvement a une aile ascétique, mais la version durable du FIRE optimise le coût de la vie *que vous voulez vivre*, pas la vie la moins chère possible. Un plan bâti sur des dépenses artificiellement écrasées échoue de la façon la plus humaine qui soit : on ne le tient pas ([[psychologie-du-retrait]]).

## Les trois nombres qui gouvernent tout

Tout projet FIRE, quelle que soit sa variante, se résume à trois nombres et à leurs incertitudes.

**1. Vos dépenses annuelles cibles.** Le nombre le plus important et le plus mal estimé. Pas vos dépenses actuelles : celles de la vie que vous visez, santé, impôts et lissage des grosses dépenses irrégulières compris ([[depenses-en-retraite]]).

**2. Votre taux de retrait.** Le pont entre dépenses et capital. 4 % est le point de départ historique ([[la-regle-des-4-pourcents]]) ; pour un départ précoce, un portefeuille mondial et des valorisations élevées, la fourchette de travail moderne est plutôt 3 à 3,5 % en retrait rigide, davantage avec une stratégie flexible ([[choisir-sa-strategie]]).

**3. Votre horizon.** 30 ans (l'hypothèse des études fondatrices) et 50 ans (un départ à 40 ans) sont deux problèmes différents. L'horizon long réduit le taux sûr, mais moins qu'on ne l'imagine : au-delà de 40 ans, la courbe s'aplatit, un portefeuille qui survit 40 ans est presque toujours devenu si gros qu'il survit indéfiniment ([[horizon-et-esperance-de-vie]]).

::: astuce Commencez par l'inventaire, pas par le simulateur
Avant tout calcul savant : douze mois de relevés bancaires, un tableur, et le vrai chiffre de vos dépenses annuelles, grosses dépenses irrégulières incluses (voiture, travaux, santé, cadeaux). La plupart des gens découvrent un écart de 15 à 25 % entre leurs dépenses perçues et réelles. Ce chiffre-là, multiplié par 25 ou 30, vaut tous les simulateurs du monde pour savoir où vous en êtes.
:::

## Comment lire ce livre

Le livre est découpé en articles autonomes, densément liés entre eux : prenez-le par où votre situation l'exige.

- **Vous découvrez le sujet** : enchaînez sur [[la-regle-des-4-pourcents]] puis [[combien-il-vous-faut]], et gardez [[erreurs-classiques-fire]] en garde-fou.
- **Vous voulez comprendre la science** : la partie « La science du retrait » se lit dans l'ordre, de [[etude-trinity]] à [[serie-ern]] ; le pivot conceptuel est [[sequence-des-rendements]].
- **Vous êtes proche du départ** : les stratégies de retrait ([[panorama-strategies-retrait]]), le portefeuille ([[allocation-actions-obligations]], [[portefeuilles-tous-temps]]) et les protections ([[cash-buffer]], [[strategie-buckets]]) sont vos chapitres.
- **Vous êtes français** : la partie fiscale ([[enveloppes-francaises]], [[taxe-puma]], [[retraite-legale]]) est écrite pour vous, ce qui est rare dans une littérature massivement américaine.
- **Vous utilisez pofo** : [[utiliser-la-page-fire]] explique chaque contrôle de la page FIRE, et [[la-machine-pofo]] ce qui se passe sous le capot.

::: terrain Le conseil le plus répété par ceux qui l'ont fait
Interrogez des FIRE effectifs (forums, blogs, meetups) et un conseil revient plus que tout autre : « j'aurais dû passer moins de temps à optimiser mon taux de retrait au dixième de point, et plus de temps à préparer ce que j'allais faire de mes journées ». La mécanique financière est la partie facile ; la partie difficile est humaine ([[temoignages-fire]], [[sens-et-identite]]). Ce livre traite les deux, dans cet ordre, mais ne sautez pas la seconde.
:::

---

## Pour aller plus loin

- Vicki Robin & Joe Dominguez, *Your Money or Your Life* (1992, réédité) : le socle philosophique.
- William Bengen, « Determining Withdrawal Rates Using Historical Data », *Journal of Financial Planning*, 1994 : l'article fondateur ([[etude-trinity]]).
- Mr. Money Mustache, « The Shockingly Simple Math Behind Early Retirement » (mrmoneymustache.com, 2012) : l'arithmétique du taux d'épargne.
- Early Retirement Now, « Safe Withdrawal Rate Series » (earlyretirementnow.com) : la référence moderne, plus de 60 volets ([[serie-ern]]).
- Reddit r/Fire et r/vosfinances (pour la France) : les communautés actives.
