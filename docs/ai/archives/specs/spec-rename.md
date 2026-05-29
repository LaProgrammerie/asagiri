# Mission — Global rename: AgentFlow / asagiri → Asagiri / asa

> Spec active de rebranding (2026-05-20). Les autres `spec*.md` conservent le vocabulaire legacy à titre d’historique.
Objectif :
Réaliser un renommage global, propre et cohérent du projet afin de remplacer :
- le nom produit actuel :
  - AgentFlow
  - asagiri
par :
- nom produit :
  - Asagiri
- commande CLI :
  - asa
Le renommage doit être :
- complet ;
- cohérent ;
- sans drift ;
- sans régression ;
- compatible release/documentation/Homebrew/CI.
IMPORTANT :
Cette migration doit être traitée comme une opération d’industrialisation critique.
Aucun ancien nom ne doit rester dans les zones publiques ou critiques du produit.
---
# Nouvelle identité
## Produit
```txt
Asagiri

Signification

朝霧 — “brume du matin”

Le nom évoque :

* clarification ;
* réduction du bruit ;
* structuration ;
* émergence ;
* orchestration maîtrisée.

Cela correspond à la philosophie du projet :

* local-first ;
* réduction de contexte ;
* orchestration déterministe ;
* workflows fiables ;
* optimisation coût/tokens ;
* visibilité et contrôle.

⸻

Commande CLI

La commande CLI officielle devient :

asa

Exemples :

asa work "develop billing-v2"
asa estimate billing-v2
asa continue
asa doctor
asa version

IMPORTANT :

* toutes les documentations ;
* exemples ;
* screenshots ;
* tests ;
* workflows ;
* CI ;
* golden files ;
* références CLI

doivent utiliser asa.

⸻

Contraintes absolues

1. Renommage cohérent global

Le renommage doit couvrir :

* code ;
* docs ;
* CI ;
* release ;
* Homebrew ;
* Cloudflare ;
* package names ;
* screenshots ;
* examples ;
* tests ;
* configs ;
* URLs ;
* artefacts ;
* logs ;
* branding ;
* generated files.

⸻

2. Ne PAS casser le produit

Le comportement fonctionnel doit rester identique.

Objectif :

* migration branding ;
* migration naming ;
* migration CLI ;
* pas refactor produit.

⸻

3. Éviter le drift

Après migration :

* ne pas avoir :
    * AgentFlow ;
    * asagiri ;
    * agentflow ;
    * hfb ;
    * anciens noms de package ;
    * anciens noms de projet.

sauf :

* migration notes ;
* compatibilité explicitement documentée ;
* historique Git.

⸻

Travaux attendus

1. Rename produit

Remplacer :

AgentFlow
asagiri

par :

Asagiri

dans :

* README ;
* docs ;
* Fumadocs ;
* pages ;
* headers ;
* titles ;
* descriptions ;
* package metadata ;
* release notes ;
* Homebrew formula ;
* GoReleaser ;
* GitHub Actions ;
* banners ;
* terminal UI ;
* logs ;
* help CLI ;
* version output ;
* screenshots ;
* diagrams ;
* comments publics importants.

⸻

2. Rename CLI

Remplacer :

agentflow

par :

asa

dans :

* Cobra root command ;
* examples ;
* docs ;
* workflows ;
* tests ;
* completions shell ;
* Homebrew install ;
* release archives ;
* screenshots ;
* terminal recordings ;
* generated docs ;
* markdown examples ;
* CI scripts ;
* Makefile ;
* integration tests ;
* snapshots ;
* golden tests.

⸻

3. Binary & release rename

Le binaire final devient :

asa

Releases :

asa_Darwin_arm64.tar.gz
asa_Linux_x86_64.tar.gz
asa_Windows_x86_64.zip

Checksums :

* mis à jour automatiquement.

Homebrew :

* formule renommée ;
* installation :

brew install asa

ou :

brew install asagiri

Décision :

* évaluer quelle convention Brew est la plus propre ;
* documenter le choix.

⸻

4. Repository & module path

Analyser :

* repo GitHub ;
* module Go ;
* paths internes.

Décider proprement :

* ce qui doit être renommé immédiatement ;
* ce qui peut rester temporairement pour éviter casse massive.

IMPORTANT :
Le module Go peut rester temporairement si nécessaire pour éviter migration destructrice immédiate.

Mais :

* tous les éléments publics doivent devenir :
    * Asagiri
    * asa

⸻

5. Documentation

Mettre à jour :

* Fumadocs ;
* README ;
* install docs ;
* workflows ;
* architecture docs ;
* screenshots ;
* examples ;
* release docs ;
* Homebrew docs ;
* CI docs ;
* Cloudflare docs ;
* contribution docs.

IMPORTANT :
Toutes les langues doivent être mises à jour :

* anglais ;
* français ;
* autres langues existantes.

Ne pas laisser :

* des commandes agentflow dans certaines langues ;
* des titres divergents ;
* du branding mixé.

⸻

6. Cloudflare Pages

Mettre à jour :

* project names ;
* variables ;
* documentation ;
* URLs si nécessaire.

Exemples :

CLOUDFLARE_PAGES_PROJECT

doit pointer vers le nouveau projet :

* asagiri-docs ;
* asa-docs ;
* ou convention cohérente.

⸻

7. GitHub Actions

Mettre à jour :

* release workflows ;
* docs workflows ;
* artifact names ;
* cache names ;
* release titles ;
* generated filenames ;
* package names ;
* brew publish ;
* comments PR éventuels.

⸻

8. Homebrew

Mettre à jour :

* tap formula ;
* package name ;
* install instructions ;
* generated formula ;
* tests Homebrew ;
* release URLs.

⸻

9. Terminal UI

Mettre à jour :

* help ;
* banners ;
* headers ;
* logs ;
* progress UI ;
* status messages ;
* reports ;
* generated markdown.

Exemple :

Avant :

AgentFlow — orchestrateur CLI local

Après :

Asagiri — deterministic orchestration for AI coding workflows

⸻

10. Generated docs & references

Regénérer :

* Cobra docs ;
* CLI reference ;
* generated markdown ;
* completions ;
* examples ;
* screenshots si automatisés.

⸻

Vérifications obligatoires

1. Search complète

Faire des recherches globales :

AgentFlow
agentflow
asagiri
hyper fast builder

Identifier :

* occurrences restantes ;
* occurrences acceptables ;
* occurrences critiques oubliées.

⸻

2. Build & release

Vérifier :

* build local ;
* tests ;
* release snapshot ;
* GoReleaser ;
* Homebrew ;
* docs build ;
* Cloudflare deploy.

⸻

3. Install flow

Tester :

brew install ...
asa version
asa doctor

⸻

4. Documentation

Vérifier :

* cohérence multi-langues ;
* cohérence screenshots ;
* cohérence examples ;
* cohérence install docs.

⸻

Critères d’acceptation

La migration est terminée si :

* le produit s’appelle Asagiri partout publiquement ;
* la commande officielle est asa ;
* les releases utilisent asa ;
* les docs utilisent asa ;
* les workflows utilisent asa ;
* Homebrew fonctionne ;
* Cloudflare Pages fonctionne ;
* les GitHub Actions fonctionnent ;
* aucun ancien branding critique ne subsiste ;
* toutes les langues sont cohérentes ;
* la release reste fonctionnelle ;
* aucun drift de configuration n’est introduit.

Le résultat attendu doit donner l’impression que :

* le projet s’est toujours appelé Asagiri ;
* la CLI s’est toujours appelée asa.

