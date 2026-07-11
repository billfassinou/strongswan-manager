# Variables d'environnement

Toute la configuration du serveur passe par l'environnement. Les valeurs par défaut permettent de démarrer en mode démo sans rien régler.

---

## Tableau complet

| Variable | Défaut | Rôle |
|---|---|---|
| `HTTP_ADDR` | `:7926` | Adresse et port d'écoute du serveur HTTP. |
| `DATABASE_URL` | `postgres://swan:swan@postgres:5432/swan?sslmode=disable` | Chaîne de connexion PostgreSQL. Les migrations sont appliquées automatiquement au démarrage. |
| `JWT_SECRET` | `dev-insecure-change-me` | **À changer.** Secret de signature des jetons d'authentification (HS256). |
| `JWT_TTL` | `1h` | Durée de vie d'un jeton. Accepte une durée Go (`30m`, `2h`) ou un entier (secondes). |
| `SEED_ADMIN_PASSWORD` | `admin1234` | **À changer.** Mot de passe des 4 comptes créés **au premier démarrage** (`admin`, `operator`, `auditor`, `viewer`). |
| `SECRETS_KEY` | `dev-insecure-secrets-key-change-me` | **À changer, et à ne plus jamais modifier.** Passphrase dont dérive la clé AES-256-GCM qui chiffre les secrets et les clés privées de la PKI. |
| `VICI_ENDPOINTS` | *(vide)* | Passerelles à piloter, sous la forme `nom=endpoint`, séparées par des virgules. **Si vide : mode démo** (adaptateur simulé + passerelle `gw-local`). |
| `POLL_INTERVAL` | `3s` | Période d'interrogation VICI (état des SA, version des passerelles). Augmentez-la sur un parc important. |
| `CORS_ORIGINS` | `*` | Origines autorisées pour l'API. À restreindre en production. |
| `CRL_URL` | *(vide)* | URL publique de la CRL, **inscrite dans les certificats émis** (CRL Distribution Point). Ex. `http://mon-serveur:7926/crl.der`. Doit être définie **avant** d'émettre des certificats. |
| `CRL_VALIDITY` | `24h` | Durée de validité (`nextUpdate`) des CRL générées : détermine la fréquence à laquelle les passerelles re-téléchargent la liste. |

---

## Format de `VICI_ENDPOINTS`

```bash
VICI_ENDPOINTS="gw-a=unix:/gw/a/charon.vici,gw-b=tcp:10.0.0.5:4502"
```

| Forme | Signification |
|---|---|
| `unix:/chemin/charon.vici` | Socket UNIX du démon (local ou partagé par volume) |
| `tcp:hôte:port` | VICI exposé en TCP |

Les passerelles sont enregistrées **au démarrage**. Voir [Connecter de vraies passerelles](14-connecter-passerelles-reelles.md).

---

## Configuration recommandée en production

```bash
HTTP_ADDR=":7926"
DATABASE_URL="postgres://user:motdepasse@db:5432/swan?sslmode=require"
JWT_SECRET="<32+ octets aléatoires>"
JWT_TTL="30m"
SECRETS_KEY="<32+ octets aléatoires, sauvegardés à part>"
SEED_ADMIN_PASSWORD="<mot de passe fort, avant le 1er démarrage>"
CORS_ORIGINS="https://vpn.mondomaine.fr"
VICI_ENDPOINTS="gw-paris=unix:/run/charon.vici"
POLL_INTERVAL="5s"
CRL_URL="https://vpn.mondomaine.fr/crl.der"
CRL_VALIDITY="12h"
```

Génération de secrets :

```bash
openssl rand -hex 32
```

---

## Trois pièges à connaître

1. **`SECRETS_KEY` ne se change pas.** Tous les secrets et clés privées déjà enregistrés ont été chiffrés avec sa valeur : la modifier les rend **définitivement illisibles**. Sauvegardez-la avec la base.
2. **`SEED_ADMIN_PASSWORD` n'agit qu'au premier démarrage**, sur une base vide. La changer plus tard ne modifie pas les comptes existants.
3. **`CRL_URL` doit être définie avant d'émettre des certificats.** Un certificat émis sans elle ne contient **pas** de point de distribution : la révocation ne pourra pas lui être appliquée. Réémettez-le si besoin.

---

## Où les positionner

Dans `backend/docker-compose.yml`, section `environment` du service `backend`, ou par l'environnement du shell :

```bash
JWT_SECRET="$(openssl rand -hex 32)" docker compose up -d
```

Les cibles `make run` et `make lab-up` fournissent des valeurs adaptées à la démonstration (`make lab-up` définit notamment `VICI_ENDPOINTS`, `CRL_URL` et un `CRL_VALIDITY` court).
