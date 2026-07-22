# Panorama des stratégies de retrait : la carte avant le territoire

Vous avez un portefeuille, un horizon et un besoin de revenu. Reste la décision qui commande toute la décumulation. Selon quelle règle, exactement, allez-vous prélever ? Combien la première année ? Et surtout, comment ce montant réagira-t-il aux marchés, à l'inflation et au temps qui passe ?

Il existe une bonne douzaine de réponses nommées : Bengen, pourcentage fixe, Guyton-Klinger, VPW, règles CAPE, guardrails Morningstar, ABW/TPAW, plancher-plafond Vanguard, rentes. La littérature les compare depuis vingt-cinq ans. Bonne nouvelle, ce zoo apparent s'organise autour d'un seul arbitrage à trois pôles. Une fois la carte en tête, chaque règle devient un point que l'on situe d'un coup d'œil, avec ses forces prévisibles et ses pathologies prévisibles.

Cette page est cette carte. Elle pose le triangle impossible du retrait, les deux stratégies extrêmes qui le bornent, la taxonomie complète des familles avec leurs articles dédiés, les six critères qui permettent de les noter honnêtement, et la lecture de la frontière des règles que la page FIRE trace pour votre plan (§06). Les dix articles suivants détaillent chaque règle. Celui-ci vous évite de vous y perdre.

::: cle Le triangle impossible du retrait
Toute règle de retrait promet trois choses désirables : un **revenu stable** (le même train de vie chaque année), la **sécurité du capital** (ne jamais épuiser le portefeuille) et un **revenu élevé** (consommer une part généreuse de l'espérance de rendement). Aucune ne peut garantir les trois à la fois. Sur un actif risqué, c'est mathématiquement impossible. Chaque stratégie nommée est donc un choix de sacrifice. Bengen sacrifie la sécurité, le pourcentage fixe sacrifie la stabilité, le retrait minuscule sacrifie le niveau. Les règles « intelligentes » sont des façons de répartir ce sacrifice plus finement. Demander « quelle est la meilleure règle ? » sans dire quel sacrifice vous préférez n'a pas de réponse.
:::

::: figure withdrawal-frontier
La frontière de décumulation : chaque règle est un point entre probabilité de ruine et variabilité du niveau de vie. Les deux extrêmes (Bengen, pourcentage fixe) bornent l'arc ; les règles modernes se rapprochent du coin idéal en écoutant plus d'information.
:::

## Les deux extrêmes purs, et la frontière entre eux

Commençons par les deux stratégies qui bornent tout l'espace, car chacune garantit parfaitement deux sommets du triangle en abandonnant le troisième.

**L'extrême « revenu d'abord », le montant fixe indexé** (la règle de Bengen, [[retrait-fixe-bengen]]). Vous retirez X € la première année, puis le même montant ajusté de l'inflation, quoi qu'il arrive. Le revenu est parfaitement stable, par construction. Le prix payé, c'est que la ruine devient possible : si les marchés déçoivent assez longtemps, le portefeuille s'épuise avant vous ([[la-regle-des-4-pourcents]], [[sequence-des-rendements]]). Tout le risque se concentre sur un seul événement, binaire et lointain.

**L'extrême « capital d'abord », le pourcentage fixe** ([[pourcentage-fixe]]). Vous retirez chaque année Y % du portefeuille courant. La ruine devient impossible, par construction : Y % de quelque chose n'est jamais tout. En échange, le revenu épouse intégralement les marchés. Un portefeuille à −35 % donne un train de vie à −35 %, potentiellement pendant des années. Tout le risque se répartit sur la vie quotidienne.

Entre ces deux pôles s'étend un continuum. C'est le graphique clé de toute cette partie : la **frontière de décumulation**, que la section §06 de la page FIRE trace pour votre plan précis (« policy frontier », [[utiliser-la-page-fire]]). Chaque règle y est un point, repéré par sa probabilité de ruine et par la variabilité du niveau de vie. Bengen se place tout en bas à droite, ruine maximale et variabilité nulle. Le pourcentage fixe se place tout en haut à gauche, ruine nulle et variabilité maximale. Les règles intermédiaires dessinent un arc entre les deux. Ce dessin rend visible d'un coup ce que vingt-cinq ans de littérature ont établi : on n'élimine pas le risque d'un plan de retrait, on en choisit la forme, entre faillite rare et brutale et ajustements fréquents et bornés. Les règles « intelligentes » ne sont pas magiques. Elles sont simplement mieux placées sur l'arc, plus proches du coin idéal (ruine faible et variabilité faible), parce qu'elles écoutent une information que les extrêmes ignorent : le portefeuille courant, les valorisations, l'horizon restant.

## La taxonomie complète

Voici le zoo, rangé en cinq familles selon l'information que chaque règle écoute. Le tableau donne la carte, chaque règle a son article détaillé.

| Famille | Ce qu'elle écoute | Règles nommées | Article |
|---|---|---|---|
| **Fixe** | Rien (le plan initial) | Bengen 4 %, variantes à indexation partielle | [[retrait-fixe-bengen]] |
| **Proportionnelle** | Le portefeuille courant | Pourcentage fixe, VPW des Bogleheads (pourcentage croissant avec l'âge) | [[pourcentage-fixe]], [[vpw]] |
| **À garde-fous** (le fixe qui plie) | Le portefeuille courant, par seuils | Guyton-Klinger (le classique), plancher-plafond, « dynamic spending » Vanguard, guardrails modernes Morningstar/Kitces-Tharp | [[guyton-klinger]], [[plancher-plafond]], [[guardrails-morningstar]] |
| **Actuarielle / par amortissement** | Le portefeuille, l'horizon restant et le rendement attendu | ABW, TPAW, les RMD américains, les règles CAPE dynamiques d'ERN | [[amortissement-abw]], [[regles-cape]] |
| **Par plancher garanti** (safety-first) | Rien : elle externalise | Rentes viagères, échelle de linkers pour l'essentiel + portefeuille pour le reste | [[rentes-et-annuites]], [[echelle-obligataire]] |

Trois observations structurent ce tableau.

**Plus une règle écoute d'information, mieux elle se place sur la frontière, et plus elle exige de gouvernance.** Le fixe n'exige rien (et n'apprend rien) ; les garde-fous exigent d'appliquer des coupes décidées à froid au pire moment émotionnel ; l'amortissement exige un recalcul annuel et l'acceptation d'un revenu officiellement variable. Le gain de placement est réel et mesurable ; le coût est comportemental, et il est réel aussi ([[psychologie-du-retrait]]).

**Les familles 3 et 4 sont les gagnantes de la recherche récente.** Le consensus moderne (Morningstar, Kitces, ERN, les Bogleheads, la littérature actuarielle) s'est déplacé. Le fixe pur est pédagogique mais dominé. Le proportionnel pur est increvable mais invivable. Les règles à information (garde-fous bien bornés, amortissement lissé) offrent le meilleur rapport entre ruine et vie vécue. La famille ABW/TPAW a la préférence de la littérature pour sa cohérence interne : c'est la seule qui ne peut, par construction, ni s'épuiser prématurément ni mourir sur un tas d'or ([[amortissement-abw]]). Les guardrails gardent la préférence des praticiens pour leur lisibilité côté client ([[guardrails-morningstar]]).

**La famille 5 change de terrain.** Le safety-first (la sécurité d'abord ; Pfau en est le théoricien contemporain : « la retraite n'est pas un problème de portefeuille, c'est un problème d'adossement de passif ») ne cherche pas un meilleur point sur la frontière. Il sort du triangle une partie des dépenses, le plancher vital, couvert par rente ou pension ([[retraite-legale]]). Le portefeuille ne porte alors plus que le confort, là où la variabilité est tolérable. Deux écoles structurent tout le débat professionnel américain : le camp « probability-based », qui optimise la frontière, et le camp « safety-first », qui garantit le plancher. La meilleure réponse pratique est souvent hybride. C'est justement ce que fait un plan français typique : pension légale en plancher, portefeuille en confort ([[revenus-complementaires]]).

## Noter une règle honnêtement : les six critères

Comparer des règles exige de regarder plus loin que le taux de succès. Sinon le pourcentage fixe « gagne » toujours (ruine nulle !) en cachant sa pathologie sous le tapis. Voici la grille d'évaluation, héritée d'ERN (volet 11) et de la pratique ([[serie-ern]]) :

1. **La ruine** (à horizon et modèle donnés) : le critère classique, nécessaire, jamais suffisant ([[ruine-et-probabilites]]).
2. **Le niveau de vie moyen servi** : combien la règle vous laisse réellement consommer, en moyenne, sur la vie du plan. Deux règles à ruine égale peuvent différer de 15 % de consommation totale.
3. **La distribution du niveau de vie dans les mauvais scénarios** : le critère qui discrimine vraiment. Que sert la règle dans le pire quartile, combien d'années sous le plancher de confort, et à quelle profondeur ? C'est la §04 (« The spending you actually live »). C'est elle qui a démasqué Guyton-Klinger, avec des taux de succès superbes payés par des décennies amputées de 30-45 % dans les millésimes 1966 ([[guyton-klinger]], [[flexibilite-realite]]).
4. **Le legs** (richesse finale distribuée) : certaines règles meurent systématiquement riches (le fixe prudent), d'autres consomment tout (l'amortissement à horizon exact) ; ni l'un ni l'autre n'est « bien », mais il faut le savoir ([[depenses-en-retraite]], [[succession-et-transmission]]).
5. **La gouvernance** : la règle est-elle exécutable par un humain sous stress, et par son conjoint survivant ? Une règle à sept paramètres recalculée mensuellement est une promesse de non-application ([[couple-et-famille]], [[revue-annuelle]]).
6. **La robustesse aux erreurs d'hypothèses** : que devient la règle si μ était surestimé d'un point ? Les règles à information se corrigent en route (c'est leur force profonde) ; le fixe encaisse l'erreur en silence jusqu'à la falaise.

::: attention Le piège des comparaisons de taux de succès
Vous verrez partout des tableaux du genre « la règle X réussit à 98 % contre 91 % pour Bengen ». Ayez le bon réflexe : demandez ce que la règle a fait aux dépenses pour acheter ces points. Toute règle flexible peut atteindre 100 % de succès en coupant assez fort et assez longtemps. Le succès du portefeuille se paie alors en échec du train de vie, que le tableau ne montre pas. La seule comparaison honnête est bidimensionnelle, la ruine et la vie vécue, c'est-à-dire la frontière. C'est pourquoi la page FIRE trace la frontière plutôt qu'un palmarès ([[utiliser-la-page-fire]]).
:::

## La carte en action : le même plan sous quatre règles

Rien ne vaut un cas concret. Le plan : 1,5 M€, un besoin de confort de 54 000 €/an (3,6 %), un plancher établi à 42 000 € ([[combien-il-vous-faut]]), 45 ans d'horizon, une pension de 15 000 €/an à partir de l'année 17. Voici les quatre points sur la frontière (chiffres indicatifs du modèle central, testez les vôtres) :

- **Bengen 54 000 € indexés** : ruine ~9 %. Revenu parfaitement stable, jusqu'à la falaise. Legs médian confortable. Le point de référence.
- **Pourcentage fixe 3,6 %** : ruine 0 %. Mais le pire quartile passe sous le plancher de 42 000 € pendant des années dès le premier régime hostile, et le revenu oscille de ±25 % d'une décennie à l'autre. Intenable pour ce ménage.
- **Guardrails (coupe à −10 % si le taux courant dépasse 4,5 %, plancher à 78 % du confort)** : ruine ~3 %, revenu stable la plupart du temps, et dans le mauvais quartile, deux à quatre coupes, jamais sous le plancher établi. Coût moyen d'environ 4 % de consommation totale en moins que Bengen.
- **ABW (amortissement sur l'horizon restant, rendement central, lissé)** : ruine structurellement ~0, consommation totale moyenne la plus élevée des quatre (la règle ose dépenser ce que les autres thésaurisent), revenu officiellement variable mais borné par la pension et le lissage ; legs faible, assumé.

La leçon n'est pas « ABW gagne ». C'est que le choix dépend du ménage. Celui-ci a un plancher haut (78 % du confort) et veut protéger le conjoint gestionnaire : les guardrails simples l'emportent. Un ménage flexible et sans souci de legs aurait pris l'ABW. Un ménage angoissé par la moindre variation aurait choisi Bengen à 3,2 %, avec plus de capital. Le travail de sélection détaillé, critère par critère et profil par profil, fait l'objet de [[choisir-sa-strategie]], l'article de synthèse de cette partie.

## Comment lire la suite de cette partie

L'ordre des articles suit la carte. Les extrêmes d'abord, parce qu'ils bornent tout : [[retrait-fixe-bengen]] puis [[pourcentage-fixe]]. Les garde-fous ensuite, du classique historique ([[guyton-klinger]], à lire pour comprendre et pour ses pathologies) aux versions modernes ([[plancher-plafond]] pour la famille Vanguard, [[guardrails-morningstar]] pour l'état de l'art praticien). Puis la famille actuarielle : [[vpw]] (le pont entre proportionnel et actuariel), [[amortissement-abw]] (le chouchou de la littérature) et [[regles-cape]] (l'information des valorisations, [[valorisations-et-cape]]). Enfin le changement de terrain : [[rentes-et-annuites]] (le plancher garanti). Et la synthèse décisionnelle : [[choisir-sa-strategie]]. Chaque article donne la mécanique exacte, les paramètres recommandés par la recherche, les pathologies connues, et la mise en œuvre pratique (toutes ces règles sont simulables nativement dans le tiroir « Spending policy » de la page FIRE, [[utiliser-la-page-fire]]).

## L'essentiel à retenir

- Le triangle impossible, revenu stable, capital sûr, revenu élevé : toute règle sacrifie un sommet, et « la meilleure règle » n'existe pas sans dire quel sacrifice vous choisissez.
- Les deux extrêmes bornent l'espace : Bengen (stable, ruine possible) et pourcentage fixe (increvable, invivable). Toutes les autres règles vivent sur la frontière entre les deux, et la page FIRE trace cette frontière pour votre plan (§06).
- Cinq familles par information écoutée : fixe, proportionnelle, garde-fous, actuarielle, plancher garanti ; la recherche récente favorise garde-fous bien bornés et amortissement, les praticiens la lisibilité des guardrails, et le plancher garanti change de terrain plutôt que de point.
- Six critères pour noter : ruine, niveau de vie moyen, vie vécue dans le pire quartile (le discriminant, §04), legs, gouvernance, robustesse aux erreurs. Tout tableau qui ne montre que le taux de succès cache l'essentiel.
- Un même plan change de règle optimale selon le plancher, le legs souhaité et la tolérance aux ajustements : la sélection raisonnée est dans [[choisir-sa-strategie]].

---

## Pour aller plus loin

- Early Retirement Now, volet 11 (les critères de notation des règles dynamiques) : la grille fondatrice ([[serie-ern]]).
- Wade Pfau, *Safety-First Retirement Planning* (2019) : l'école de l'adossement, face aux « probability-based ».
- Morningstar, *The State of Retirement Income* : le comparatif annuel des règles avec le critère de vie vécue ([[guardrails-morningstar]]).
- Bogleheads wiki, « Withdrawal methods » : l'inventaire communautaire, bien tenu.
- Dans pofo : le tiroir « Spending policy » (toutes les familles simulables) et la frontière §06 ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
