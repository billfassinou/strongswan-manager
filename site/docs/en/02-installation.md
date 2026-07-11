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
# → ouvrez https://localhost:7926
```

`make run` builds the image, starts PostgreSQL, applies the migrations, creates the demo accounts, generates the internal certificate authority **and the server's TLS certificate**, then launches the application.

**What you should see**: after a few seconds, `https://localhost:7926` shows the login screen — **preceded by a security warning from your browser**.

---

## First access: the browser warning

This is normal, and it is not a defect: the application **generated its own certificate**, signed
by its internal CA. Your browser does not know that authority, so it warns you.

> **The connection really is encrypted.** What is not attested is the server's *identity* — not the
> confidentiality of the exchange. This is exactly how Proxmox, pfSense or TrueNAS behave on first
> startup.

To continue: **Advanced** → **Proceed to localhost**.

### Making the warning go away (recommended for long-term use)

Import the internal CA into the trust store of your admin workstations. This is a **one-off** —
it will then cover all your instances.

```bash
# récupérer la CA (elle est publique : aucun secret ici)
curl -sk https://localhost:7926/api/v1/ca \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["cert_pem"])' > ca.crt
```

You will also find it in the **PKI & Certificates** (*PKI & Certificats*) screen.

- **macOS**: double-click `ca.crt` → *System* keychain → set it to "Always Trust".
- **Linux**: `sudo cp ca.crt /usr/local/share/ca-certificates/ && sudo update-ca-certificates`
- **Windows**: `certutil -addstore -f Root ca.crt` (as administrator).

### The other options

| You want… | Do this |
|---|---|
| A **publicly trusted** certificate, with no warning | `ACME_DOMAIN=vpn.mondomaine.fr` → **Let's Encrypt**. Requires a **public domain** and **port 80 reachable from the Internet**. |
| To use **your** certificate (corporate PKI) | `TLS_CERT` and `TLS_KEY` |
| To terminate TLS on an existing **reverse proxy** | `TLS_ENABLED=false` |

⚠️ If you reach the console through a domain name, **declare it** in `TLS_SANS`
(e.g. `TLS_SANS="localhost,127.0.0.1,vpn.interne.fr"`) — otherwise the browser will report, on top
of everything else, a name mismatch. See [Environment variables](A2-configuration.md).

---

## The two ports

| Port | Protocol | Serves |
|---|---|---|
| **7926** | **HTTPS** | The interface, the API, real time. |
| **7927** | Plain HTTP | Only the **CRL** (`/crl.der`) and `/healthz`. The rest is redirected to HTTPS. |

Port 7927 is not an oversight: the CRL distribution point **must** stay on HTTP, because it is
charon that reads it and charon would not trust our internal CA. Details in
[Environment variables](A2-configuration.md).

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

> ⚠️ **`SECRETS_KEY` also encrypts the key of your TLS certificate.** Changing it would leave the
> server unable to read it back — you would have to reissue one.

Other production considerations:

- **TLS is already on**: the application serves HTTPS out of the box. Give it a publicly trusted
  certificate (`ACME_DOMAIN`) or your own (`TLS_CERT`/`TLS_KEY`), or import its internal CA onto
  your admin workstations.
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
