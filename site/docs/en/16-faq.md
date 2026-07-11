# FAQ

---

### Does the tool write to my `swanctl.conf` files?

**No.** All configuration goes through the daemon's **VICI API**. Your files are neither read nor modified. A connection loaded by the console exists **inside the daemon** (visible with `swanctl --list-conns`), not on disk.

---

### Can I use it in an isolated (air-gapped) environment?

Yes, with some caveats:

- No Internet connection is required to run it: the images, the database, the PKI and the front end are self-contained.
- The certificate authority is **internal** — no need for ACME / Let's Encrypt.
- The API documentation page (`/api/v1/docs`) loads Swagger UI from a CDN: **it will not render** offline. The raw specification (`/api/v1/openapi.yaml`) is still served locally.

---

### How do I back up?

Everything that matters is in **PostgreSQL**: tunnels, versions, encrypted secrets, PKI, audit, accounts.

```bash
docker compose exec postgres pg_dump -U swan swan > sauvegarde.sql
```

⚠️ Also keep **`SECRETS_KEY`** somewhere safe: without it, the secrets and private keys in the backup are **undecryptable**.

Restore:

```bash
cat sauvegarde.sql | docker compose exec -T postgres psql -U swan swan
```

---

### Can the application be served over HTTPS?

The application serves **plain HTTP** on port 7926. Put it behind a TLS reverse proxy (Nginx, Traefik, Caddy). Remember to forward **WebSockets**.

---

### Are the front end and the API separate?

No: the React front end is **embedded in the binary** and served from the **same origin** as the API. That is what lets the JWT token and the WebSocket work without any CORS configuration.

---

### Can I use the tool without installing StrongSwan?

Yes — that is **demo mode** (the default). A simulated gateway lets you explore the whole interface. See [Installation](02-installation.md).

---

### How many gateways / tunnels can it handle?

The design target is **1,000 tunnels and 10 gateways** with a responsive interface. The server polls each gateway every `POLL_INTERVAL` (3 s by default); raise that value if you manage a large estate.

---

### Where are the account passwords stored?

Hashed with **bcrypt** in the database. They are never stored in clear text and cannot be read back.

---

### Why can't I see a PSK's value again?

That is intentional. Once entered, a secret is **never** shown again — not in the interface, not through the API, for any role. If you lost it, create a new one and update both endpoints. See [Managing secrets](06-secrets.md).

---

### How do I automate this (Terraform, CI/CD, scripts)?

Everything the interface does is available through the **REST API**. Create a token with a dedicated account and call the API. See [REST API & WebSocket](A3-api.md).

---

### I want to contribute / change the code. Where do I start?

```bash
cd backend
make web     # builds the front end (required before a local Go build)
make build   # compile
make test    # unit tests
make cover   # + coverage
make test-integration   # integration tests (starts a disposable PostgreSQL)
```

**Unit tests** live next to the code they test; **integration tests** are in `backend/test/`.

Prerequisites: **Go** and **Node** (otherwise the `Makefile` falls back to a `golang` Docker image).

---

### What is planned but not there yet?

- **OCSP** responder, **SCEP/EST** enrollment
- **Remote mTLS agent** (would replace direct access to the VICI socket)
- **Multi-tenancy** and SAML/OIDC **SSO**
- Actual delivery of **notifications** (email, Slack, webhook)
- **Vault** instead of the application-level encryption of secrets
- A **version history** screen in the interface (the API is already complete)

---

### A question that isn't here?

→ [Troubleshooting](15-depannage.md)
