# Installation

## Prérequis

- **Docker** et **Docker Compose** (c'est tout).
- Le port **8080** libre sur votre machine.

Rien d'autre n'est nécessaire : PostgreSQL, les migrations de base et le front sont pris en charge automatiquement.

> Vous voulez développer, pas seulement utiliser ? Il vous faudra aussi **Go** et **Node**. Voir [FAQ](16-faq.md).

---

## Démarrer en trois commandes

```bash
git clone <dépôt> && cd strongswan/backend
make run
# → ouvrez http://localhost:8080
```

`make run` construit l'image, démarre PostgreSQL, applique les migrations, crée les comptes de démonstration, génère l'autorité de certification interne, puis lance l'application.

**Ce que vous devez voir** : au bout de quelques secondes, `http://localhost:8080` affiche l'écran de connexion.

---

## Se connecter

Quatre comptes sont créés au premier démarrage. Le mot de passe est le même pour tous : **`admin1234`** (valeur par défaut de `SEED_ADMIN_PASSWORD`).

| Identifiant | Rôle | Peut modifier ? |
|---|---|---|
| `admin` | Administrateur | **Oui** |
| `operator` | Opérateur | **Oui** |
| `auditor` | Auditeur | Non (lecture seule) |
| `viewer` | Lecture seule | Non |

Connectez-vous avec **`admin` / `admin1234`** pour tout voir.

> Ces comptes ne sont créés **qu'au premier démarrage**, si la base est vide. Ils ne sont pas recréés ensuite.

---

## Mode démo vs vraies passerelles

Par défaut, l'application démarre en **mode démo** : elle enregistre une passerelle nommée `gw-local` reliée à un **adaptateur VICI simulé**. Toute l'interface et l'API sont pleinement utilisables (créer un tunnel, le monter, voir l'état changer en temps réel) **sans installer StrongSwan**.

Pour piloter de **vraies passerelles** StrongSwan :

```bash
make lab-up      # démarre en plus 2 conteneurs strongSwan et s'y connecte via VICI
```

Voir [Connecter de vraies passerelles](14-connecter-passerelles-reelles.md).

---

## Sécuriser une installation réelle

Les valeurs par défaut sont volontairement lisibles pour la démonstration. **Avant toute mise en service**, changez au minimum :

| Variable | Défaut (non sûr) | À faire |
|---|---|---|
| `JWT_SECRET` | `dev-insecure-change-me` | Une chaîne aléatoire longue (≥ 32 caractères) |
| `SECRETS_KEY` | `dev-insecure-secrets-key-change-me` | Une passphrase forte — **elle chiffre tous vos secrets et clés privées** |
| `SEED_ADMIN_PASSWORD` | `admin1234` | Un mot de passe fort, **avant le premier démarrage** |

Exemple :

```bash
JWT_SECRET="$(openssl rand -hex 32)" \
SECRETS_KEY="$(openssl rand -hex 32)" \
SEED_ADMIN_PASSWORD='UnMotDePasseFort!' \
docker compose up --build -d
```

> ⚠️ **Ne changez pas `SECRETS_KEY` après coup** : les secrets et clés privées déjà stockés ont été chiffrés avec l'ancienne valeur et deviendraient illisibles.

Autres points pour la production :

- Placez l'application **derrière un reverse proxy HTTPS** (Nginx, Traefik, Caddy). L'application sert du HTTP en clair.
- Restreignez `CORS_ORIGINS` à votre domaine plutôt que `*`.
- Sauvegardez la base PostgreSQL (elle contient la configuration, la PKI et l'audit).

La liste complète est dans [Variables d'environnement](A2-configuration.md).

---

## Arrêter, redémarrer, nettoyer

```bash
docker compose stop            # arrêter (les données sont conservées)
docker compose up -d           # redémarrer
docker compose logs -f backend # suivre les logs   (ou : make logs)

docker compose down            # arrêter et supprimer les conteneurs
docker compose down -v         # ⚠️ + EFFACER la base (tunnels, PKI, audit, comptes)
```

`make lab-down` fait la même chose pour le lab (conteneurs strongSwan compris).

---

## Et ensuite ?

→ [Découvrir l'interface](03-decouvrir-linterface.md)
