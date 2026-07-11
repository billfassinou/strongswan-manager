# Environment variables

The whole server configuration comes from the environment. The default values are enough to start in demo mode without setting anything.

---

## Full table

| Variable | Default | Purpose |
|---|---|---|
| `HTTP_ADDR` | `:7926` | Main listen address. **Serves HTTPS** (unless `TLS_ENABLED=false`). |
| `DATABASE_URL` | `postgres://swan:swan@postgres:5432/swan?sslmode=disable` | PostgreSQL connection string. Migrations are applied automatically at startup. |
| `JWT_SECRET` | `dev-insecure-change-me` | **Must be changed.** Signing secret for authentication tokens (HS256). |
| `JWT_TTL` | `1h` | Token lifetime. Accepts a Go duration (`30m`, `2h`) or an integer (seconds). |
| `SEED_ADMIN_PASSWORD` | `admin1234` | **Must be changed.** Password of the 4 accounts created **on first startup** (`admin`, `operator`, `auditor`, `viewer`). |
| `SECRETS_KEY` | `dev-insecure-secrets-key-change-me` | **Must be changed, and then never modified again.** Passphrase from which the AES-256-GCM key encrypting the secrets and the PKI private keys is derived. |
| `VICI_ENDPOINTS` | *(empty)* | Gateways to drive, in the form `name=endpoint`, comma-separated. **If empty: demo mode** (mock adapter + `gw-local` gateway). |
| `POLL_INTERVAL` | `3s` | VICI polling period (SA state, gateway version). Increase it on a large estate. |
| `CORS_ORIGINS` | `*` | Origins allowed for the API. Should be restricted in production. |
| `CRL_URL` | *(empty)* | Public URL of the CRL, **embedded in the issued certificates** (CRL Distribution Point). E.g. `http://my-server:7927/crl.der`. Must be set **before** issuing any certificate. |
| `CRL_VALIDITY` | `24h` | Validity window (`nextUpdate`) of the generated CRLs: it determines how often the gateways re-download the list. |

### TLS

| Variable | Default | Purpose |
|---|---|---|
| `TLS_ENABLED` | `true` | The application **serves HTTPS by default**. Set it to `false` **only** behind a reverse proxy (nginx, Traefik, Caddy) that already terminates TLS. |
| `HTTP_REDIRECT_ADDR` | `:7927` | **Plaintext** listener. It serves exactly two things: the **CRL** (`/crl.der`) and `/healthz`; everything else is redirected with a **308** to HTTPS. |
| `TLS_CERT` / `TLS_KEY` | *(empty)* | Paths to **your** certificate and its key (PEM). When provided, they replace the auto-generated certificate. |
| `TLS_SANS` | `localhost,127.0.0.1,::1,<hostname>` | Names and IPs covered by the auto-generated certificate. **Add the name your users use to reach the console**, otherwise their browser will report a name mismatch. |
| `ACME_DOMAIN` | *(empty)* | If set → **Let's Encrypt**: a publicly trusted certificate, no warning. Requires a **public domain** and **port 80 reachable from the Internet**. |
| `ACME_EMAIL` | *(empty)* | ACME contact (expiry notices). |
| `ACME_CACHE` | `./acme` | Cache for ACME certificates. **Mount it as a volume**: without it, every restart requests a new certificate and you will hit Let's Encrypt's rate limits. |

---

## Why two ports?

| Port | Protocol | Serves |
|---|---|---|
| **7926** | **HTTPS** | The interface, the API, the WebSocket. |
| **7927** | **Plain HTTP** | Only `/crl.der` and `/healthz`. Everything else → **308** to HTTPS. |

This is not an oversight: **the CRL distribution point has to stay on HTTP**. It is charon that
reads it, and charon would reject a certificate signed by your internal CA — yet to validate that
certificate, it would need precisely that CRL. RFC 5280 breaks this circularity by serving CDPs in
the clear. A CRL is **signed** and **public** data: serving it unencrypted exposes nothing.

That is why `CRL_URL` is written as **`http://…:7927/crl.der`**, not `https`.

---

## Where does the certificate come from?

Three sources, in this order of precedence:

1. **`ACME_DOMAIN`** → Let's Encrypt. A certificate trusted by every browser.
2. **`TLS_CERT` + `TLS_KEY`** → your certificate (from your corporate PKI, for example).
3. **Otherwise** → an **auto-generated** certificate, signed by the application's **internal CA**,
   issued on first startup and **persisted in the database**.

Case 3 is what lets the application start in HTTPS **with no configuration at all**.
The certificate is kept in the database — rather than regenerated at every startup — so that its
fingerprint stays stable: otherwise the administrator would see a warning on every restart, and
would quickly get into the habit of ignoring it.

**The browser will warn** as long as the internal CA has not been imported. To get rid of the
warning: fetch the CA (the **PKI & Certificates** (*PKI & Certificats*) screen, or `GET /api/v1/ca`)
and import it into the trust store of your admin workstations. See [Installation](02-installation.md).

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

With ACME, publish ports **80 → 7927** (the HTTP-01 challenge arrives on port 80) and
**443 → 7926**.

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
