# Moyenne arithmétique, moyenne géométrique et volatility drag

Voici la question piège la plus rentable de toute la finance personnelle : un placement fait +50 % la première année, −50 % la seconde. Rendement moyen : 0 %. Combien avez-vous gagné ?

Réponse : vous avez **perdu** 25 % (100 → 150 → 75). La « moyenne » qu'on vous a annoncée est parfaitement exacte, et parfaitement trompeuse. Cette page démonte le mécanisme, la différence entre moyenne arithmétique et moyenne géométrique et le « volatility drag » qui les sépare, parce qu'il est la clé de voûte silencieuse de tout le sujet FIRE. C'est lui qui explique pourquoi les rendements des plaquettes ne sont pas des rendements vivables, pourquoi la volatilité est un coût même sans krach, pourquoi les fonds à levier déçoivent, et pourquoi le taux de retrait sûr est si loin des « 7 % des actions ».

Après cette page, plus personne ne pourra vous vendre une moyenne.

::: cle Les deux moyennes
La moyenne **arithmétique** additionne les rendements et divise : (+50 − 50) / 2 = 0 %. Elle répond à « combien rapporte une année typique, prise isolément ? ». La moyenne **géométrique** compose : √(1,50 × 0,50) − 1 = −13,4 % par an. Elle répond à « à quel taux régulier mon capital a-t-il réellement crû ? ». C'est le CAGR, le seul chiffre qui décrit ce que vit un capital qui reste investi. La géométrique est **toujours** inférieure ou égale à l'arithmétique, et l'écart grandit avec la volatilité. Votre patrimoine vit en géométrique ; les plaquettes parlent en arithmétique.
:::

## Le volatility drag : la volatilité est un coût

L'écart entre les deux moyennes obéit à une approximation célèbre et remarquablement précise :

> rendement géométrique ≈ rendement arithmétique − volatilité² / 2

Le terme σ²/2 est le **volatility drag** (traînée de volatilité). Il vient d'une asymétrie que tout le monde connaît sans en tirer les conséquences : après −20 %, il faut +25 % pour revenir ; après −50 %, il faut +100 %. Les pertes pèsent mécaniquement plus lourd que les gains de même taille, et plus les oscillations sont amples, plus la composition en souffre.

Quelques ordres de grandeur pour calibrer l'intuition (drag = σ²/2) :

| Actif | Volatilité annuelle | Volatility drag | Arithmétique 7 % devient... |
|---|---|---|---|
| Fonds monétaire | ~1 % | ~0,005 % | 7,0 % |
| Portefeuille 60/40 | ~10 % | ~0,5 % | 6,5 % |
| Actions mondiales | ~15 % | ~1,1 % | 5,9 % |
| Actions émergentes | ~22 % | ~2,4 % | 4,6 % |
| Actions à levier ×2 | ~30 % | ~4,5 % | ~2,5 % (avant frais de levier !) |

La dernière ligne explique un phénomène qui surprend tous les débutants : un ETF à levier quotidien ×2 sur un indice volatil peut faire **moins** bien que l'indice sur longue période, alors qu'il double fidèlement chaque journée. Le levier double l'arithmétique mais quadruple la variance : dès que 2 × drag dépasse le rendement gagné, le levier détruit ([[levier-et-marges]]). Même mécanique pour comprendre pourquoi la diversification est le seul « repas gratuit » : combiner des actifs décorrélés baisse σ sans baisser la moyenne arithmétique, donc **augmente** la géométrique. La diversification ne promet pas de meilleures années moyennes ; elle promet une meilleure composition ([[portefeuilles-tous-temps]], [[actifs-defensifs]]).

::: exemple Vérifiez sur deux lignes
Actif A : +7 % chaque année, sans variation. Actif B : alternance +27 % / −13 %, même moyenne arithmétique de 7 %, volatilité 20 points. Après 30 ans : A × 7,61 ; B × (1,27 × 0,87)^15 = 1,1049^15 ≈ × 4,47. Même « moyenne », 70 % de richesse finale d'écart. Le drag prédit : 7 % − 0,20²/2 = 5 % ; et 1,05^30 ≈ 4,32, très proche du 4,47 exact. L'approximation σ²/2 est un excellent outil de coin de table.
:::

## La cascade : de la plaquette au taux de retrait

Le volatility drag est la première marche d'une cascade qui mène du chiffre marketing au chiffre vivable. Suivons « les actions font 10 % » jusqu'au taux de retrait, étape par étape :

1. **10 % arithmétique nominal** : la moyenne des années, celle des plaquettes et des manuels.
2. **− drag (~1,1 % à 15 % de vol)** → ~8,9 % géométrique nominal : ce que compose un capital investi.
3. **− inflation (~2,5 %)** → **~6,4 % géométrique RÉEL** : la seule monnaie qui compte sur 40 ans ([[inflation-et-taux-de-retrait]]). Historiquement, les actions mondiales ont livré ~5 % géométrique réel ; les portefeuilles diversifiés 60/40, plutôt 3,5 à 4,5 %.
4. **− frais et fiscalité** (0,3 à 1,5 % selon vos enveloppes et véhicules, [[etf-ucits-europeens]], [[flat-tax-et-imposition]]).
5. **− prime de séquence** : même le géométrique réel net n'est retirable que si les rendements arrivent sans désordre ; leur irrégularité face à des retraits fixes coûte encore 1 à 1,5 point ([[sequence-des-rendements]]).

Arrivée : 3 à 3,5 % de taux de retrait rigide soutenable pour un horizon long, exactement la fourchette que la recherche moderne trouve par des chemins indépendants ([[la-regle-des-4-pourcents]], [[serie-ern]]). Ce n'est pas une coïncidence : la règle des 4 % n'**est pas** mystérieuse, elle est l'aboutissement comptable de cette cascade. Mémorisez la hiérarchie : **arithmétique > géométrique > géométrique réel > taux de retrait soutenable**, avec 0,5 à 1,5 point perdu à chaque marche.

::: attention Le test du vendeur
Quand on vous annonce un rendement, posez systématiquement les trois questions de la cascade : arithmétique ou géométrique (composé) ? Nominal ou réel ? Brut ou net de frais ? « 8 % » peut signifier 6,9 % composé, 4,4 % réel, 3,4 % net : moins de la moitié du chiffre annoncé, sans aucun mensonge formel. Les rétro-projections de produits structurés et les moyennes de fonds affichent presque toujours le coin le plus flatteur du tableau. Un professionnel qui ne sait pas répondre à ces trois questions n'a pas compris son propre produit.
:::

## Trois applications directes au FIRE

**1. Calibrer un simulateur.** Les curseurs de la page FIRE de pofo travaillent en **réel** : le μ demandé est un rendement réel espéré, et le moteur applique la volatilité σ pour générer les trajectoires, donc le drag émerge tout seul dans les résultats ([[la-machine-pofo]]). Le piège classique consiste à entrer un μ arithmétique nominal (« 8 % ») : vous venez de fabriquer un monde de rêve. Repère : pour un portefeuille diversifié mondial, un μ réel de 4 à 5 % avec σ 12-15 % est la zone raisonnable ; pofo le pré-remplit depuis l'historique de vos fonds puis le tire vers un prior prudent, précisément pour vous éviter cette erreur ([[rendre-monte-carlo-pertinent]], [[rendements-attendus]]).

**2. Juger un portefeuille de retrait.** Deux portefeuilles de même espérance arithmétique ne se valent pas : le moins volatil compose mieux et résiste mieux à la séquence, double dividende. C'est pourquoi les portefeuilles de retrait sérieux sacrifient de la moyenne pour de la régularité (obligations, or, diversification de régimes, [[allocation-actions-obligations]], [[portefeuilles-tous-temps]]) et pourquoi le « 100 % actions, c'est optimal à long terme » de l'accumulation ne survit pas au premier retrait ([[erreurs-classiques-fire]]).

**3. Lire ses propres performances.** Votre relevé annuel moyen « +9 % sur 5 ans » est probablement arithmétique. Le seul chiffre honnête pour vous-même : (valeur finale / valeur initiale)^(1/n) − 1, corrigé des apports (pofo calcule le TRI et le CAGR proprement sur vos flux réels). Beaucoup d'investisseurs découvrent que leur performance composée réelle est 2 à 3 points sous leur impression, la différence part en drag, frais et mauvais timing des apports.

## Pour les curieux : pourquoi σ²/2 exactement

Sans formalisme : le logarithme d'un rendement, ln(1+r), est ce qui s'additionne vraiment d'une année à l'autre (composer, c'est additionner des logs). Or ln(1+r) est une fonction concave. Elle pénalise les écarts vers le bas plus qu'elle ne récompense les écarts vers le haut. En développant ln(1+r) ≈ r − r²/2 et en prenant l'espérance, le terme −r²/2 fait apparaître −σ²/2 en moyenne. C'est l'inégalité de Jensen rendue comptable. Le drag n'est donc ni un frottement de marché ni un coût caché. C'est une propriété mathématique de la composition, aussi inévitable que les intérêts composés eux-mêmes, dont il est exactement le côté obscur. Les modèles de pofo travaillent nativement en composition période par période, le drag y est donc capté par construction, pas par approximation ([[la-machine-pofo]]).

## L'essentiel à retenir

- Deux moyennes : l'arithmétique (l'année typique) et la géométrique (la croissance vécue) ; votre capital vit en géométrique, le marketing parle en arithmétique.
- L'écart est le volatility drag ≈ σ²/2 : la volatilité est un coût de composition, même sans krach et même en moyenne nulle.
- La cascade plaquette → vivable : − drag, − inflation, − frais, − prime de séquence ; il reste 3-3,5 % de retrait rigide soutenable, la règle des 4 % démystifiée.
- Diversifier augmente la géométrique à arithmétique égale. C'est la justification mathématique du portefeuille de retrait diversifié ; le levier fait l'inverse.
- Trois questions réflexes devant tout chiffre : composé ? réel ? net ? Et pour vos simulateurs : μ **réel** et modeste, jamais la moyenne de la plaquette.

---

## Pour aller plus loin

- Early Retirement Now, SWR Series volet 8 (l'appendice technique) et volet 33 (calculer un taux de retrait sans simulation) : la cascade formalisée ([[serie-ern]]).
- William Bernstein, *The Intelligent Asset Allocator*, chapitre 1 : la meilleure introduction écrite aux deux moyennes.
- Markowitz (1952) et la lecture moderne « geometric mean maximization » (Kelly, Latané) : pourquoi maximiser la géométrique, pas l'arithmétique.
- La suite logique dans ce livre : [[sequence-des-rendements]] (la marche suivante de la cascade) et [[rendements-attendus]] (quelles valeurs de μ sont défendables aujourd'hui).
