# Panorama des stratégies de retrait : la carte avant le territoire

Vous avez un portefeuille, un horizon, un besoin de revenu ; reste **la** décision opérationnelle de toute la décumulation : selon quelle règle, exactement, allez-vous prélever ? Combien la première année, et surtout comment ce montant réagira-t-il aux marchés, à l'inflation, au temps qui passe ?

Il existe une bonne douzaine de réponses nommées : Bengen, pourcentage fixe, Guyton-Klinger, VPW, règles CAPE, guardrails Morningstar, ABW/TPAW, plancher-plafond Vanguard, rentes... et la littérature les compare depuis vingt-cinq ans. La bonne nouvelle : ce zoo apparent s'organise entièrement autour d'**un** arbitrage fondamental à trois pôles, et une fois la carte en tête, chaque règle devient un point identifiable sur elle, avec ses forces prévisibles et ses pathologies prévisibles.

Cette page est cette carte : le triangle impossible du retrait, les deux stratégies extrêmes qui le bornent, la taxonomie complète des familles avec leurs articles dédiés, les critères d'évaluation qui permettent de les noter honnêtement, et la lecture de la frontière des règles que pofo trace pour **votre** plan (§06). Les dix articles suivants de cette partie détaillent chaque règle ; celui-ci vous évite de vous y perdre.

::: cle Le triangle impossible du retrait
Toute règle de retrait promet trois choses désirables : un **revenu stable** (le même train de vie chaque année), la **sécurité du capital** (ne jamais épuiser le portefeuille), et un **revenu élevé** (consommer une part généreuse de l'espérance de rendement). Aucune règle ne peut garantir les trois : sur un actif risqué, c'est mathématiquement impossible. Chaque stratégie nommée est un choix de sacrifice : Bengen sacrifie la sécurité, le pourcentage fixe sacrifie la stabilité, le retrait minuscule sacrifie le niveau, et toutes les règles « intelligentes » sont des façons de **répartir** le sacrifice plus finement. Demander « quelle est la meilleure règle ? » sans dire quel sacrifice vous préférez n'a pas de réponse.
:::

## Les deux extrêmes purs, et la frontière entre eux

Commençons par les deux stratégies qui bornent tout l'espace, car chacune garantit parfaitement deux sommets du triangle en abandonnant le troisième.

**L'extrême « revenu d'abord » : le montant fixe indexé** (la règle de Bengen, [[retrait-fixe-bengen]]). Retirez X € la première année, puis le même montant ajusté de l'inflation, quoi qu'il arrive. Revenu : parfaitement stable, par construction. Prix payé : la ruine est **possible** : si les marchés déçoivent assez longtemps, le portefeuille s'épuise avant vous ([[la-regle-des-4-pourcents]], [[sequence-des-rendements]]). Tout le risque est concentré sur un seul événement binaire et lointain.

**L'extrême « capital d'abord » : le pourcentage fixe** ([[pourcentage-fixe]]). Retirez chaque année Y % du portefeuille **courant**. Ruine : impossible, par construction (Y % de quelque chose n'est jamais tout). Prix payé : le revenu épouse intégralement les marchés : −35 % de portefeuille = −35 % de train de vie, potentiellement pendant des années. Tout le risque est étalé sur la vie quotidienne.

Entre ces deux pôles s'étend un continuum, et c'est **le** graphique à comprendre de toute cette partie : la **frontière de décumulation**, que la section §06 de la page FIRE trace pour votre plan précis (« policy frontier », [[utiliser-la-page-fire]]). Chaque règle y est un point dans le plan (probabilité de ruine ; variabilité du niveau de vie) : Bengen tout à droite en bas (ruine maximale, variabilité nulle), le pourcentage fixe tout à gauche en haut (ruine nulle, variabilité maximale), et toutes les règles intermédiaires dessinent un arc entre les deux. La frontière rend visible d'un coup ce que vingt-cinq ans de littérature ont établi : **on n'élimine pas le risque d'un plan de retrait, on choisit sa forme** : faillite rare et brutale, ou ajustements fréquents et bornés. Les règles « intelligentes » ne sont pas magiques ; elles sont mieux **placées** sur l'arc (plus proches du coin idéal ruine faible + variabilité faible) parce qu'elles utilisent de l'information (le portefeuille courant, les valorisations, l'horizon restant) que les extrêmes ignorent.

## La taxonomie complète

Voici le zoo, organisé en cinq familles selon l'**information** que chaque règle écoute. Le tableau donne la carte ; chaque règle a son article détaillé.

| Famille | Ce qu'elle écoute | Règles nommées | Article |
|---|---|---|---|
| **Fixe** | Rien (le plan initial) | Bengen 4 %, variantes à indexation partielle | [[retrait-fixe-bengen]] |
| **Proportionnelle** | Le portefeuille courant | Pourcentage fixe, VPW des Bogleheads (pourcentage croissant avec l'âge) | [[pourcentage-fixe]], [[vpw]] |
| **À garde-fous** (le fixe qui plie) | Le portefeuille courant, par seuils | Guyton-Klinger (le classique), plancher-plafond, « dynamic spending » Vanguard, guardrails modernes Morningstar/Kitces-Tharp | [[guyton-klinger]], [[plancher-plafond]], [[guardrails-morningstar]] |
| **Actuarielle / par amortissement** | Le portefeuille, l'horizon restant **et** le rendement attendu | ABW, TPAW, les RMD américains, les règles CAPE dynamiques d'ERN | [[amortissement-abw]], [[regles-cape]] |
| **Par plancher garanti** (safety-first) | Rien : elle externalise | Rentes viagères, échelle de linkers pour l'essentiel + portefeuille pour le reste | [[rentes-et-annuites]], [[echelle-obligataire]] |

Trois observations structurent ce tableau.

**Plus une règle écoute d'information, mieux elle se place sur la frontière, et plus elle exige de gouvernance.** Le fixe n'exige rien (et n'apprend rien) ; les garde-fous exigent d'appliquer des coupes décidées à froid au pire moment émotionnel ; l'amortissement exige un recalcul annuel et l'acceptation d'un revenu officiellement variable. Le gain de placement est réel et mesurable ; le coût est comportemental, et il est réel aussi ([[psychologie-du-retrait]]).

**Les familles 3 et 4 sont les gagnantes de la recherche récente.** Le consensus moderne (Morningstar, Kitces, ERN, les Bogleheads, la littérature actuarielle) s'est déplacé : le fixe pur est pédagogique mais dominé ; le proportionnel pur est increvable mais invivable ; les règles à information (garde-fous bien bornés, amortissement lissé) offrent le meilleur rapport ruine/vie vécue. La colonne ABW/TPAW a la préférence de la littérature récente pour sa cohérence interne (c'est la seule famille qui ne peut par construction ni s'épuiser prématurément ni mourir sur un tas d'or, [[amortissement-abw]]) ; les guardrails gardent la préférence des praticiens pour leur lisibilité client ([[guardrails-morningstar]]).

**La famille 5 change de terrain.** Le safety-first (Pfau en est le théoricien contemporain : « la retraite n'est pas un problème de portefeuille, c'est un problème d'adossement de passif ») ne cherche pas un meilleur point sur la frontière : il **sort** du triangle une partie des dépenses (le plancher vital, couvert par rente ou pension, [[retraite-legale]]) et laisse le portefeuille ne porter que le confort, où la variabilité est tolérable. Les deux écoles (« probability-based » : optimiser la frontière ; « safety-first » : garantir le plancher) structurent tout le débat professionnel américain, et la meilleure réponse pratique est souvent hybride : c'est exactement ce que fait un plan français typique, pension légale en plancher + portefeuille en confort ([[revenus-complementaires]]).

## Noter une règle honnêtement : les six critères

Comparer des règles exige de regarder plus loin que le taux de succès, sinon le pourcentage fixe « gagne » toujours (ruine nulle !) en cachant sa pathologie sous le tapis. La grille d'évaluation, héritée d'ERN (Part 11) et de la pratique ([[serie-ern]]) :

1. **La ruine** (à horizon et modèle donnés) : le critère classique, nécessaire, jamais suffisant ([[ruine-et-probabilites]]).
2. **Le niveau de vie moyen servi** : combien la règle vous laisse réellement consommer, en moyenne, sur la vie du plan. Deux règles à ruine égale peuvent différer de 15 % de consommation totale.
3. **La distribution du niveau de vie dans les MAUVAIS scénarios** : **le** critère discriminant. Que sert la règle dans le pire quartile : combien d'années sous le plancher de confort, à quelle profondeur ? C'est la §04 de pofo (« The spending you actually live »), et c'est elle qui a démasqué Guyton-Klinger (des taux de succès superbes payés par des décennies amputées de 30-45 % dans les millésimes 1966, [[guyton-klinger]], [[flexibilite-realite]]).
4. **Le legs** (richesse finale distribuée) : certaines règles meurent systématiquement riches (le fixe prudent), d'autres consomment tout (l'amortissement à horizon exact) ; ni l'un ni l'autre n'est « bien », mais il faut le savoir ([[depenses-en-retraite]], [[succession-et-transmission]]).
5. **La gouvernance** : la règle est-elle exécutable par un humain sous stress, et par son conjoint survivant ? Une règle à sept paramètres recalculée mensuellement est une promesse de non-application ([[couple-et-famille]], [[revue-annuelle]]).
6. **La robustesse aux erreurs d'hypothèses** : que devient la règle si μ était surestimé d'un point ? Les règles à information se corrigent en route (c'est leur force profonde) ; le fixe encaisse l'erreur en silence jusqu'à la falaise.

::: attention Le piège des comparaisons de taux de succès
Vous verrez partout des tableaux « la règle X réussit à 98 % contre 91 % pour Bengen ». Réflexe : demandez **ce que** la règle a fait aux dépenses pour acheter ces points. Toute règle flexible peut atteindre 100 % de succès en coupant assez fort assez longtemps : le succès du portefeuille se paie alors en échec du train de vie, que le tableau ne montre pas. La seule comparaison honnête est bidimensionnelle (ruine **et** vie vécue : la frontière), et c'est pourquoi pofo trace la frontière plutôt qu'un palmarès ([[utiliser-la-page-fire]]).
:::

## La carte en action : le même plan sous quatre règles

Rien ne vaut une passe concrète. Plan : 1,5 M€, besoin de confort 54 000 €/an (3,6 %), plancher établi à 42 000 € ([[combien-il-vous-faut]]), 45 ans, pension 15 000 €/an en année 17. Les quatre points sur la frontière (chiffres indicatifs du modèle central ; testez les vôtres) :

- **Bengen 54 000 € indexés** : ruine ~9 %. Revenu parfaitement stable... jusqu'à la falaise. Legs médian confortable. Le point de référence.
- **Pourcentage fixe 3,6 %** : ruine 0 %. Mais le pire quartile passe sous le plancher de 42 000 € pendant des années au premier régime hostile, et le revenu oscille de ±25 % d'une décennie à l'autre : intenable pour ce ménage.
- **Guardrails (coupe à −10 % si le taux courant dépasse 4,5 %, plancher à 78 % du confort)** : ruine ~3 %, revenu stable la plupart du temps, et dans le mauvais quartile : deux à quatre coupes, jamais sous le plancher établi. Coût moyen : ~4 % de consommation totale en moins que Bengen.
- **ABW (amortissement sur l'horizon restant, rendement central, lissé)** : ruine structurellement ~0, consommation totale moyenne la plus **élevée** des quatre (la règle ose dépenser ce que les autres thésaurisent), revenu officiellement variable mais borné par la pension et le lissage ; legs faible, assumé.

La leçon n'est pas « ABW gagne » : c'est que le **choix** dépend du ménage. Celui-ci a un plancher haut (78 % du confort) et veut protéger le conjoint gestionnaire : les guardrails simples l'emportent. Un ménage flexible sans souci de legs aurait pris l'ABW ; un ménage angoissé par toute variation, Bengen à 3,2 % avec plus de capital. Le travail de sélection détaillé, critère par critère et profil par profil, est l'objet de [[choisir-sa-strategie]], l'article de synthèse de cette partie.

## Comment lire la suite de cette partie

L'ordre des articles suit la carte. Les extrêmes d'abord, parce qu'ils bornent tout : [[retrait-fixe-bengen]] puis [[pourcentage-fixe]]. Les garde-fous ensuite, du classique historique ([[guyton-klinger]], à lire pour comprendre **et** pour ses pathologies) aux versions modernes ([[plancher-plafond]] pour la famille Vanguard, [[guardrails-morningstar]] pour l'état de l'art praticien). Puis la famille actuarielle : [[vpw]] (le pont entre proportionnel et actuariel), [[amortissement-abw]] (le chouchou de la littérature) et [[regles-cape]] (l'information des valorisations, [[valorisations-et-cape]]). Enfin le changement de terrain : [[rentes-et-annuites]] (le plancher garanti). Et la synthèse décisionnelle : [[choisir-sa-strategie]]. Chaque article donne la mécanique exacte, les paramètres recommandés par la recherche, les pathologies connues, et la mise en œuvre dans pofo (toutes ces règles sont simulables nativement dans le tiroir « Spending policy », [[utiliser-la-page-fire]]).

## L'essentiel à retenir

- Le triangle impossible : revenu stable, capital sûr, revenu élevé : toute règle sacrifie un sommet ; « la meilleure règle » n'existe pas sans dire quel sacrifice vous choisissez.
- Les deux extrêmes bornent l'espace : Bengen (stable, ruine possible) et pourcentage fixe (increvable, invivable) ; toutes les autres règles vivent sur la frontière entre les deux, et pofo trace cette frontière pour **votre** plan (§06).
- Cinq familles par information écoutée : fixe, proportionnelle, garde-fous, actuarielle, plancher garanti ; la recherche récente favorise garde-fous bien bornés et amortissement, les praticiens la lisibilité des guardrails, et le plancher garanti change de terrain plutôt que de point.
- Six critères pour noter : ruine, niveau de vie moyen, vie vécue dans le pire quartile (**le** discriminant : §04), legs, gouvernance, robustesse aux erreurs. Tout tableau qui ne montre que le taux de succès cache l'essentiel.
- Un même plan change de règle optimale selon le plancher, le legs souhaité et la tolérance aux ajustements : la sélection raisonnée est dans [[choisir-sa-strategie]].

---

## Pour aller plus loin

- Early Retirement Now, Part 11 (les critères de notation des règles dynamiques) : la grille fondatrice ([[serie-ern]]).
- Wade Pfau, *Safety-First Retirement Planning* (2019) : l'école de l'adossement, face aux « probability-based ».
- Morningstar, *The State of Retirement Income* : le comparatif annuel des règles avec le critère de vie vécue ([[guardrails-morningstar]]).
- Bogleheads wiki, « Withdrawal methods » : l'inventaire communautaire, bien tenu.
- Dans pofo : le tiroir « Spending policy » (toutes les familles simulables) et la frontière §06 ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
