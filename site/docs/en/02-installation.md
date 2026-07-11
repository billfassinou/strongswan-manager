# Installation

## Requirements

- **Docker** and **Docker Compose** (that's all).
- Port **7926** free on your machine.

Nothing else is needed: PostgreSQL, the database migrations and the front end are all taken care of automatically.

> Want to develop, not just use? You will also need **Go** and **Node**. See the [FAQ](16-faq.md).

---

## Get started in three commands

```bash
git clone <dépôt> && cd strongswan/backend
make run
# → ouvrez http://localhost:7926
```

`make run` builds the image, starts PostgreSQL, applies the migrations, creates the demo accounts, generates the internal certificate authority, then launches the application.

**What you should see**: after a few seconds, `http://localhost:7926` shows the login screen.

---

## Logging in

Four accounts are created on first startup. The password is the same for all of them: **`admin1234`** (the default value of `SEED_ADMIN_PASSWORD`).

| Username | Role | Can change things? |
|---|---|---|
| `admin` | Administrator | **Yes** |
| `operator` | Operator | **Yes** |
| `auditor` | Auditor | No (read-only) |
| `viewer` | Read-only | No |

Log in with **`admin` / `admin1234`** to see everything.

> These accounts are only created **on first startup**, if the database is empty. They are not recreated afterwards.

---

## Demo mode vs real gateways

By default, the application starts in **demo mode**: it registers a gateway named `gw-local` wired to a **mock VICI adapter**. The whole interface and API are fully usable (create a tunnel, bring it up, watch the state change in real time) **without installing StrongSwan**.

To drive **real** StrongSwan gateways:

```bash
make lab-up      # démarre en plus 2 conteneurs strongSwan et s'y connecte via VICI
```

See [Connect real gateways](14-connecter-passerelles-reelles.md).

---

## Securing a real installation

The default values are deliberately readable for demo purposes. **Before going live**, change at least:

| Variable | Default (not safe) | What to do |
|---|---|---|
| `JWT_SECRET` | `dev-insecure-change-me` | A long random string (≥ 32 characters) |
| `SECRETS_KEY` | `dev-insecure-secrets-key-change-me` | A strong passphrase — **it encrypts all your secrets and private keys** |
| `SEED_ADMIN_PASSWORD` | `admin1234` | A strong password, **before the first startup** |

Example:

```bash
JWT_SECRET="$(openssl rand -hex 32)" \
SECRETS_KEY="$(openssl rand -hex 32)" \
SEED_ADMIN_PASSWORD='UnMotDePasseFort!' \
docker compose up --build -d
```

> ⚠️ **Do not change `SECRETS_KEY` after the fact**: the secrets and private keys already stored were encrypted with the old value and would become unreadable.

Other production considerations:

- Put the application **behind an HTTPS reverse proxy** (Nginx, Traefik, Caddy). The application serves plain HTTP.
- Restrict `CORS_ORIGINS` to your domain rather than `*`.
- Back up the PostgreSQL database (it holds the configuration, the PKI and the audit log).

The complete list is in [Environment variables](A2-configuration.md).

---

## Stopping, restarting, cleaning up

```bash
docker compose stop            # arrêter (les données sont conservées)
docker compose up -d           # redémarrer
docker compose logs -f backend # suivre les logs   (ou : make logs)

docker compose down            # arrêter et supprimer les conteneurs
docker compose down -v         # ⚠️ + EFFACER la base (tunnels, PKI, audit, comptes)
```

`make lab-down` does the same thing for the lab (including the strongSwan containers).

---

## What's next?

→ [Discover the interface](03-decouvrir-linterface.md)
