# Contribuer

Merci de l'intérêt porté au projet. Voici ce qu'il faut savoir avant d'ouvrir une *pull
request*.

## Deux choses à accepter d'abord

1. **Licence** — le cœur est sous **[AGPL-3.0](LICENSE)**. Toute contribution y est intégrée
   sous cette licence.
2. **CLA** — le projet suit un modèle **open-core** : signer le [CLA](CLA.md) est
   **obligatoire**. Concrètement, il suffit de signer vos commits :

   ```bash
   git commit -s -m "votre message"
   ```

   Cela ajoute un `Signed-off-by:` qui vaut acceptation. Sans lui, la contribution ne peut pas
   être fusionnée — non par formalisme, mais parce que le projet perdrait le droit de faire
   évoluer son modèle de licence. Le [CLA](CLA.md) explique pourquoi, et ne vous retire aucun
   droit sur votre travail.

## Mettre en route l'environnement

```bash
cd backend
make run                # PostgreSQL + application (mode démo, VICI simulé)
make lab-up             # + 2 vraies passerelles strongSwan (tunnels réels)
```

Prérequis : **Docker**. Pour un build local : **Go 1.23+** et **Node 22+** (sinon le
`Makefile` retombe sur une image Docker `golang`).

## Avant de proposer un changement

```bash
cd backend
make web                # construit le front (requis avant un build Go local)
make build
make vet
make test               # tests unitaires — doivent rester verts
make test-integration   # tests d'intégration (PostgreSQL jetable)
```

**Ajoutez ou mettez à jour un test à chaque changement de comportement.** Les tests unitaires
vivent **à côté du code qu'ils testent** (convention Go) ; les tests d'intégration sont dans
[`backend/test/`](backend/test/).

## Les invariants à ne pas casser

Ce sont des choix structurants, pas des préférences :

- **Les tunnels s'appliquent par VICI** (`load-conn`), **jamais** en écrivant un fichier de
  configuration. C'est la raison d'être du produit.
- **Le journal d'audit est append-only** et chaîné par hachage — la base elle-même refuse toute
  modification.
- **Les secrets et les clés privées ne sortent jamais de l'API**, pour aucun rôle.
- **Toute modification de configuration crée une version** (le rollback en dépend).
- **Les erreurs suivent le contrat** `error` / `message` / `details` / `correlation_id`
  (422 pour une validation).
- **La couche HTTP dépend d'interfaces**, pas du store concret — c'est ce qui la rend testable.

## En-têtes de licence

Tout nouveau fichier source commence par :

```go
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
```

## Signaler un bug ou une faille

- **Bug** : ouvrez une *issue* avec la version, les étapes de reproduction et les logs.
- **Faille de sécurité** : **n'ouvrez pas d'issue publique.** Écrivez en privé au mainteneur.

## Documentation

La documentation utilisateur est dans [`site/docs/`](site/docs/) — Markdown, **FR et EN**.
Une nouvelle page doit être écrite **dans les deux langues** et déclarée dans le tableau `PARTS`
de `site/docs/index.html` **et** `site/docs/en/index.html`.

Règle non négociable : **tout ce qui est écrit doit exister**. Une fonctionnalité non
implémentée est présentée comme telle.
