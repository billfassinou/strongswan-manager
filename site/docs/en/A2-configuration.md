# Environment variables

The whole server configuration comes from the environment. The default values are enough to start in demo mode without setting anything.

---

## Full table

| Variable | Default | Purpose |
|---|---|---|
| `HTTP_ADDR` | `:7926` | Listen address and port of the HTTP server. |
| `DATABASE_URL` | `postgres://swan:swan@postgres:5432/swan?sslmode=disable` | PostgreSQL connection string. Migrations are applied automatically at startup. |
| `JWT_SECRET` | `dev-insecure-change-me` | **Must be changed.** Signing secret for authentication tokens (HS256). |
| `JWT_TTL` | `1h` | Token lifetime. Accepts a Go duration (`30m`, `2h`) or an integer (seconds). |
| `SEED_ADMIN_PASSWORD` | `admin1234` | **Must be changed.** Password of the 4 accounts created **on first startup** (`admin`, `operator`, `auditor`, `viewer`). |
| `SECRETS_KEY` | `dev-insecure-secrets-key-change-me` | **Must be changed, and then never modified again.** Passphrase from which the AES-256-GCM key encrypting the secrets and the PKI private keys is derived. |
| `VICI_ENDPOINTS` | *(empty)* | Gateways to drive, in the form `name=endpoint`, comma-separated. **If empty: demo mode** (mock adapter + `gw-local` gateway). |
| `POLL_INTERVAL` | `3s` | VICI polling period (SA state, gateway version). Increase it on a large estate. |
| `CORS_ORIGINS` | `*` | Origins allowed for the API. Should be restricted in production. |
| `CRL_URL` | *(empty)* | Public URL of the CRL, **embedded in the issued certificates** (CRL Distribution Point). E.g. `http://my-server:7926/crl.der`. Must be set **before** issuing any certificate. |
| `CRL_VALIDITY` | `24h` | Validity window (`nextUpdate`) of the generated CRLs: it determines how often the gateways re-download the list. |

---

## Format of `VICI_ENDPOINTS`

```bash
VICI_ENDPOINTS="gw-a=unix:/gw/a/charon.vici,gw-b=tcp:10.0.0.5:4502"
```

| Form | Meaning |
|---|---|
| `unix:/path/charon.vici` | UNIX socket of the daemon (local or shared through a volume) |
| `tcp:host:port` | VICI exposed over TCP |

Gateways are registered **at startup**. See [Connecting real gateways](14-connecter-passerelles-reelles.md).

---

## Recommended production configuration

```bash
HTTP_ADDR=":7926"
DATABASE_URL="postgres://user:motdepasse@db:5432/swan?sslmode=require"
JWT_SECRET="<32+ random bytes>"
JWT_TTL="30m"
SECRETS_KEY="<32+ random bytes, backed up separately>"
SEED_ADMIN_PASSWORD="<strong password, before the 1st startup>"
CORS_ORIGINS="https://vpn.mondomaine.fr"
VICI_ENDPOINTS="gw-paris=unix:/run/charon.vici"
POLL_INTERVAL="5s"
CRL_URL="https://vpn.mondomaine.fr/crl.der"
CRL_VALIDITY="12h"
```

Generating secrets:

```bash
openssl rand -hex 32
```

---

## Three pitfalls to know about

1. **`SECRETS_KEY` cannot be changed.** All the secrets and private keys already stored were encrypted with its value: changing it makes them **permanently unreadable**. Back it up along with the database.
2. **`SEED_ADMIN_PASSWORD` only takes effect on the first startup**, on an empty database. Changing it later does not modify the existing accounts.
3. **`CRL_URL` must be set before issuing certificates.** A certificate issued without it contains **no** distribution point: revocation cannot be applied to it. Reissue it if needed.

---

## Where to set them

In `backend/docker-compose.yml`, in the `environment` section of the `backend` service, or through the shell environment:

```bash
JWT_SECRET="$(openssl rand -hex 32)" docker compose up -d
```

The `make run` and `make lab-up` targets provide values suited to a demo (`make lab-up` notably sets `VICI_ENDPOINTS`, `CRL_URL` and a short `CRL_VALIDITY`).
