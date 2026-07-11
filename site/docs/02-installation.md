# Installation

## Prérequis

- **Docker** et **Docker Compose** (c'est tout).
- Le port **7926** libre sur votre machine.

Rien d'autre n'est nécessaire : PostgreSQL, les migrations de base et le front sont pris en charge automatiquement.

> Vous voulez développer, pas seulement utiliser ? Il vous faudra aussi **Go** et **Node**. Voir [FAQ](16-faq.md).

---

## Démarrer en trois commandes

```bash
git clone <dépôt> && cd strongswan/backend
make run
# → ouvrez https://localhost:7926
```

`make run` construit l'image, démarre PostgreSQL, applique les migrations, crée les comptes de démonstration, génère l'autorité de certification interne **et le certificat TLS du serveur**, puis lance l'application.

**Ce que vous devez voir** : au bout de quelques secondes, `https://localhost:7926` affiche l'écran de connexion — **précédé d'un avertissement de sécurité du navigateur**.

---

## Le premier accès : l'avertissement du navigateur

C'est normal, et ce n'est pas un défaut : l'application a **généré son propre certificat**,
signé par sa CA interne. Votre navigateur ne connaît pas cette autorité, donc il vous prévient.

> **La connexion est bel et bien chiffrée.** Ce qui n'est pas attesté, c'est l'*identité* du
> serveur — pas la confidentialité de l'échange. C'est exactement le comportement de Proxmox,
> pfSense ou TrueNAS au premier démarrage.

Pour continuer : **Avancé** → **Continuer vers localhost**.

### Faire disparaître l'avertissement (recommandé pour un usage durable)

Importez la CA interne dans le magasin de confiance de vos postes d'administration. À faire
**une seule fois** — elle vaudra ensuite pour toutes vos instances.

```bash
# récupérer la CA (elle est publique : aucun secret ici)
curl -sk https://localhost:7926/api/v1/ca \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["cert_pem"])' > ca.crt
```

Vous la trouverez aussi dans l'écran **PKI & Certificats**.

- **macOS** : double-cliquez sur `ca.crt` → Trousseau *Système* → réglez sur « Toujours approuver ».
- **Linux** : `sudo cp ca.crt /usr/local/share/ca-certificates/ && sudo update-ca-certificates`
- **Windows** : `certutil -addstore -f Root ca.crt` (en administrateur).

### Les autres options

| Vous voulez… | Faites |
|---|---|
| Un certificat **reconnu**, sans avertissement | `ACME_DOMAIN=vpn.mondomaine.fr` → **Let's Encrypt**. Exige un **domaine public** et le **port 80 joignable depuis Internet**. |
| Utiliser **votre** certificat (PKI d'entreprise) | `TLS_CERT` et `TLS_KEY` |
| Terminer le TLS sur un **reverse proxy** existant | `TLS_ENABLED=false` |

⚠️ Si vous accédez à la console par un nom de domaine, **déclarez-le** dans `TLS_SANS`
(ex. `TLS_SANS="localhost,127.0.0.1,vpn.interne.fr"`) — sinon le navigateur signalera, en plus,
une incohérence de nom. Voir [Variables d'environnement](A2-configuration.md).

---

## Les deux ports

| Port | Protocole | Sert |
|---|---|---|
| **7926** | **HTTPS** | L'interface, l'API, le temps réel. |
| **7927** | HTTP en clair | Uniquement la **CRL** (`/crl.der`) et `/healthz`. Le reste est redirigé vers HTTPS. |

Le port 7927 n'est pas un oubli : le point de distribution de CRL **doit** rester en HTTP, car
c'est charon qui le lit et il ne ferait pas confiance à notre CA interne. Détail dans
[Variables d'environnement](A2-configuration.md).

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

> ⚠️ **`SECRETS_KEY` chiffre aussi la clé de votre certificat TLS.** La changer rendrait le
> serveur incapable de le relire — il faudrait en réémettre un.

Autres points pour la production :

- **Le TLS est déjà actif** : l'application sert en HTTPS d'emblée. Donnez-lui un certificat
  reconnu (`ACME_DOMAIN`) ou le vôtre (`TLS_CERT`/`TLS_KEY`), ou importez sa CA interne sur les
  postes d'administration.
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
