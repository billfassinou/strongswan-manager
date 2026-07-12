# Variables d'environnement

Toute la configuration du serveur passe par l'environnement. Les valeurs par défaut permettent de démarrer en mode démo sans rien régler.

---

## Tableau complet

| Variable | Défaut | Rôle |
|---|---|---|
| `HTTP_ADDR` | `:7926` | Adresse d'écoute principale. **Sert en HTTPS** (sauf si `TLS_ENABLED=false`). |
| `DATABASE_URL` | `postgres://swan:swan@postgres:5432/swan?sslmode=disable` | Chaîne de connexion PostgreSQL. Les migrations sont appliquées automatiquement au démarrage. |
| `JWT_SECRET` | `dev-insecure-change-me` | **Le serveur refuse de démarrer** avec cette valeur, qui est publique. Secret de signature des jetons (HS256). `openssl rand -hex 32`. |
| `JWT_TTL` | `1h` | Durée de vie d'un jeton. Accepte une durée Go (`30m`, `2h`) ou un entier (secondes). |
| `SEED_ADMIN_PASSWORD` | `admin1234` | Mot de passe des 4 comptes créés **au premier démarrage** (`admin`, `operator`, `auditor`, `viewer`). La console **impose de le changer** à la première connexion : tant que ce n'est pas fait, l'API répond `403` partout ailleurs. |
| `SECRETS_KEY` | `dev-insecure-secrets-key-change-me` | **Le serveur refuse de démarrer** avec cette valeur. Passphrase dont dérive la clé AES-256-GCM qui chiffre les secrets, les clés privées de la PKI **et celle du certificat TLS**. À fixer avant le premier démarrage, **et à ne plus jamais modifier**. Sauvegardez-la avec la base — `swanmgrctl backup` archive les deux ensemble. |
| `ALLOW_INSECURE_DEFAULTS` | `false` | Autorise le démarrage avec les deux secrets ci-dessus laissés à leur valeur par défaut. **Réservé au lab** (`make run`, `make lab-up`) : ces valeurs sont dans le dépôt, donc publiques. |
| `VICI_ENDPOINTS` | *(vide)* | Passerelles à piloter, sous la forme `nom=endpoint`, séparées par des virgules. **Si vide : mode démo** (adaptateur simulé + passerelle `gw-local`). |
| `POLL_INTERVAL` | `3s` | Période d'interrogation VICI (état des SA, version des passerelles). Augmentez-la sur un parc important. |
| `CORS_ORIGINS` | `*` | Origines autorisées pour l'API. À restreindre en production. |
| `CRL_URL` | *(vide)* | URL publique de la CRL, **inscrite dans les certificats émis** (CRL Distribution Point). Ex. `http://mon-serveur:7927/crl.der`. Doit être définie **avant** d'émettre des certificats. |
| `CRL_VALIDITY` | `24h` | Durée de validité (`nextUpdate`) des CRL générées : détermine la fréquence à laquelle les passerelles re-téléchargent la liste. |

### TLS

| Variable | Défaut | Rôle |
|---|---|---|
| `TLS_ENABLED` | `true` | L'application **sert en HTTPS par défaut**. Ne passez à `false` **que** derrière un reverse proxy (nginx, Traefik, Caddy) qui termine déjà le TLS. |
| `HTTP_REDIRECT_ADDR` | `:7927` | Écouteur **en clair**. Il ne sert que deux choses : la **CRL** (`/crl.der`) et `/healthz` ; tout le reste est redirigé en **308** vers HTTPS. |
| `TLS_CERT` / `TLS_KEY` | *(vide)* | Chemins vers **votre** certificat et sa clé (PEM). Fournis, ils remplacent le certificat auto-généré. |
| `TLS_SANS` | `localhost,127.0.0.1,::1,<hostname>` | Noms et IP que couvre le certificat auto-généré. **Ajoutez-y le nom par lequel vos utilisateurs accèdent à la console**, sinon leur navigateur signalera une incohérence. |
| `ACME_DOMAIN` | *(vide)* | Si défini → **Let's Encrypt** : certificat reconnu, aucun avertissement. Exige un **domaine public** et le **port 80 joignable depuis Internet**. |
| `ACME_EMAIL` | *(vide)* | Contact ACME (avis d'expiration). |
| `ACME_CACHE` | `./acme` | Cache des certificats ACME. **À monter en volume** : sans lui, chaque redémarrage redemande un certificat et vous atteindrez les quotas de Let's Encrypt. |

---

## Pourquoi deux ports ?

| Port | Protocole | Sert |
|---|---|---|
| **7926** | **HTTPS** | L'interface, l'API, le WebSocket. |
| **7927** | **HTTP en clair** | Uniquement `/crl.der` et `/healthz`. Tout le reste → **308** vers HTTPS. |

Ce n'est pas un oubli : **le point de distribution de CRL doit rester en HTTP**. C'est charon
qui va le lire, et il refuserait un certificat signé par votre CA interne — or, pour valider ce
certificat, il lui faudrait justement la CRL. La RFC 5280 tranche cette circularité en servant
les CDP en clair. Une CRL est une donnée **signée** et **publique** : la servir en clair
n'expose rien.

C'est pourquoi `CRL_URL` s'écrit en **`http://…:7927/crl.der`**, et non en `https`.

---

## D'où vient le certificat ?

Trois sources, dans cet ordre de priorité :

1. **`ACME_DOMAIN`** → Let's Encrypt. Certificat reconnu par tous les navigateurs.
2. **`TLS_CERT` + `TLS_KEY`** → votre certificat (celui de votre PKI d'entreprise, par exemple).
3. **Sinon** → un certificat **auto-généré**, signé par la **CA interne** de l'application, émis
   au premier démarrage et **persisté en base**.

Le cas 3 est ce qui permet à l'application de démarrer en HTTPS **sans aucune configuration**.
Le certificat est conservé en base — et non régénéré à chaque démarrage — pour que son empreinte
reste stable : sinon l'administrateur reverrait un avertissement à chaque redémarrage, et
prendrait vite l'habitude de l'ignorer.

**Le navigateur avertira** tant que la CA interne n'est pas importée. Pour supprimer
l'avertissement : récupérez la CA (écran **PKI & Certificats**, ou `GET /api/v1/ca`) et importez-la
dans le magasin de confiance de vos postes d'administration. Voir [Installation](02-installation.md).

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
CRL_URL="http://vpn.mondomaine.fr:7927/crl.der"   # en http : c'est charon qui le lit
CRL_VALIDITY="12h"

# TLS — deux options, au choix :
# (a) certificat reconnu, automatique (exige un domaine public + le port 80 ouvert)
ACME_DOMAIN="vpn.mondomaine.fr"
ACME_EMAIL="admin@mondomaine.fr"
ACME_CACHE="/data/acme"                # À MONTER EN VOLUME (quotas Let's Encrypt)

# (b) votre propre certificat
# TLS_CERT="/etc/ssl/vpn.crt"
# TLS_KEY="/etc/ssl/vpn.key"

# (c) ne rien mettre : certificat auto-généré (avertissement navigateur tant que la CA
#     interne n'est pas importée). Pensez alors à déclarer le nom d'accès :
# TLS_SANS="localhost,127.0.0.1,vpn.mondomaine.fr"
```

Avec ACME, publiez les ports **80 → 7927** (le challenge HTTP-01 arrive sur le port 80) et
**443 → 7926**.

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
