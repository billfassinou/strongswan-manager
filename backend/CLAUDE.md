# Backend (`backend/`) — Go vertical slice

The repo-wide rules (open-core, SPDX header, `spec/` is never committed, domain decisions)
are in the root `CLAUDE.md`. The SPA has its own file: `backend/web/CLAUDE.md`.

The first backend increment is a **walking skeleton** implementing the Community edition's
core path end-to-end. Module path `strongswan-manager`; monolithe modulaire (§5) with the
VICI adapter and poller behind interfaces so they can be externalized later.

- **Stack**: Go 1.23 · chi (HTTP) · pgx/PostgreSQL · golang-jwt + RBAC · **govici** (VICI,
  primary path) · coder/websocket · Prometheus (`/metrics`) · OpenAPI (`/api/v1/docs`).
- **Layout**: `cmd/server` (composition root) · `internal/{config,domain,store,auth,vici,ws,
  poller,metrics,httpapi}` · `migrations/` (embedded, applied at startup) · `openapi/openapi.yaml`
  (the **API contract**, source of truth for §10) · `lab/` (strongSwan container config).
- **Go toolchain**: installed locally via Homebrew (`go version` → 1.26.x). The Makefile
  (`make build`, `make test`, `make vet`, `make tidy`) uses the local `go` when present and
  falls back to a Docker `golang:1.23` image otherwise. Both paths are equivalent.
- **Run**: `make run` (postgres + backend, **mock VICI** adapter, fully exercisable API,
  seeds users `admin/operator/auditor/viewer` with `SEED_ADMIN_PASSWORD`, default `admin1234`).
  `make lab-up` adds two real strongSwan 6.x containers (VICI over a shared `/var/run/charon.vici`
  volume); set `VICI_ENDPOINTS` in `docker-compose.yml` to switch the backend from mock to real.
- **Key invariants to preserve**: tunnels are applied via VICI `load-conn` (never by writing
  files — V3); the security score is the Go port of `scoreTunnel` from the reference mock (keep both
  in sync); `audit_log` is append-only (enforced by a DB trigger) and hash-chained; every
  config change creates a `config_versions` snapshot (rollback depends on it); error responses
  follow the §10 shape (`error/message/details/correlation_id`, 422 for validation).
- **Secrets (EF-05)** are implemented: `internal/secrets` (AES-256-GCM at rest, key derived
  from `SECRETS_KEY`), `store.SecretRepo`, `/api/v1/secrets` CRUD (values **never** returned —
  masked in responses, `List` omits the ciphertext). On tunnel apply, a PSK tunnel whose
  `auth.secret_ref` names an existing secret triggers a VICI `load-shared` on the gateway
  (`applyToVICI`). `secret_ref`/`cert_ref` are **TEXT names** (not UUIDs) referencing
  `secrets.name`. Verified against real charon in the lab.
- **Site-to-site both-ends** works: a tunnel with `peer_gateway_id` (nullable `*string` on
  `domain.Tunnel`) makes `applyToVICI` load a **mirror connection** (`mirrorTunnel`, swapped
  addrs/subnets) + PSK on the peer gateway too; `initiate` then establishes a real SA. Verified
  in the lab (`hub: ESTABLISHED`, `ESP:AES_GCM_16-256 INSTALLED`). `secret_ref`/`cert_ref`/
  `peer_gateway_id` are TEXT.
- **PKI (EF-04)** is implemented: `internal/pki` (ECDSA CA + `IssueCert`), `store.CARepo`/
  `CertRepo`, `certificates`/`cert_authorities` tables, `/api/v1/ca` + `/api/v1/certificates`
  (issue/list/revoke — **private keys never returned**, stored encrypted via the secrets
  cipher). The CA is generated once at startup (`ensureCA`). A tunnel with `auth.method=cert`
  and `cert_ref` (+ `peer_cert_ref` for the peer, both TEXT names) makes `applyOne` load
  CA+cert+key on the gateway via VICI. **govici `load-cert`/`load-key` must receive DER** (the
  adapter converts PEM→DER; charon rejects raw PEM). Cert-based S2S establishment verified in
  the lab — needs `libstrongswan-standard-plugins` (openssl) in the strongSwan image for
  ECDSA/ECDH.
- **CRL (revocation)** is implemented: `pki.GenerateCRL` (signed by the CA), `store` CRL
  columns (`crl_number`/`crl_pem`, `revoked_at`), `regenerateCRL` on every revoke, `GET
  /api/v1/crl` and a **public unauthenticated `GET /crl.der`** (the CDP). strongSwan has **no
  VICI command to load a CRL** — revocation works via the **CRL Distribution Point** embedded
  in issued certs (`CRL_URL`) fetched by charon's `curl` plugin. `CRL_VALIDITY` sets the CRL
  nextUpdate window (short in the lab to force re-fetch). Verified in the lab: the gateway
  fetches `/crl.der` and the revoked serial is listed; charon's automatic rejection then
  depends on its CRL cache/`remote.revocation` policy.
- **HTTPS is the default** (`TLS_ENABLED=true`). **Two listeners**, and this split is load-bearing:
  - **`:7926` — HTTPS**: UI, API, WebSocket (the front already derives `wss://` from
    `location.protocol`, so it needed no change).
  - **`:7927` — PLAIN HTTP** (`HTTP_REDIRECT_ADDR`, `httpapi.PlainRouter`): serves **only**
    `/crl.der` and `/healthz`; everything else gets a **308** to HTTPS. **Never move the CDP
    behind TLS**: charon fetches it with its `curl` plugin and would not trust our internal CA —
    and validating that cert would require the very CRL it is fetching (RFC 5280). `CRL_URL`
    must therefore stay `http://…:7927/crl.der`. An ACME HTTP-01 challenge is served here too,
    unredirected.
  - Certificate sources, in priority order (`buildTLS` in `cmd/server/main.go`): **ACME**
    (`ACME_DOMAIN`, needs a public domain + port 80) → **your own** (`TLS_CERT`/`TLS_KEY`) →
    **auto-generated** (`internal/tlsx`, signed by the internal CA, **persisted in DB**,
    migration `0007_server_tls`, key encrypted with the secrets cipher). It is persisted, not
    regenerated, so the fingerprint stays stable across restarts — otherwise the admin sees a
    new browser warning every time and learns to ignore it. Reissued only if absent, expiring
    (<30 days), or if `TLS_SANS` changed.
  - **HSTS is deliberately NOT set** with a self-signed cert: it would lock the admin out of
    their own console.
  - `TLS_ENABLED=false` keeps the old plain-HTTP behaviour, for deployments behind a TLS
    reverse proxy. Do not remove it.
  - **Go 1.23 is the floor**: `golang.org/x/{crypto,net,text,sys,sync}` are **pinned** to the
    last versions that still declare `go 1.23`. Running `go get @latest` bumps the `go`
    directive to 1.25 and **breaks the Docker build** (`golang:1.23-alpine`). If you must
    upgrade them, bump the Dockerfile and `release.yml` together.
- **Insecure defaults are fatal**: `config.Validate()` refuses to start if `JWT_SECRET` or
  `SECRETS_KEY` still hold their dev value (they are in the repo, hence public). The escape
  hatch is `ALLOW_INSECURE_DEFAULTS=true`, set **only** in `backend/docker-compose.yml` (lab).
  It also drives `seedUsers`: outside the lab the 4 seeded accounts get
  `must_change_password`, and `requirePasswordChanged` (in `httpapi`) then answers **403 on
  every route except `/me` and `/me/password`** — the lock is server-side, not just in the SPA.
  Keeping the lab exempt is deliberate: `make run` must stay immediately usable, and the API
  doc examples log in with `admin/admin1234`.
- The binary answers `--version` (ldflags `-X main.version`), which `swanmgrctl` relies on.
- **Out of scope for this slice** (later increments): OCSP responder, SCEP/EST enrollment
  (rest of EF-24), Vault (replace the app-level cipher), remote mTLS agent,
  pools/RADIUS/policies/daemon-params, IA, multi-tenant/SSO, TimescaleDB, Helm.

## Testing

- **Test layout**: unit tests live **next to the code they test** (Go requires `_test.go` files
  to sit in the same package to reach unexported identifiers and to count toward that package's
  coverage — don't try to move them into a shared folder). **Integration tests** (black-box,
  exported API only, real Postgres) live in **`backend/test/`** (`package test`, `//go:build
  integration`).
- `make test` — unit tests, **no external deps**. `make cover` for per-package coverage.
  `make test-integration` — runs `./test/` against a disposable Postgres (needs local `go` +
  docker).
- **The HTTP layer and the poller depend on repo interfaces** (`internal/httpapi/interfaces.go`,
  `internal/poller`), not the concrete `*store.Store` — this is what makes them unit-testable
  with in-memory fakes + the mock VICI adapter. Keep new handlers behind these interfaces; if a
  handler needs a new store method, add it to the relevant interface so fakes stay in sync.
- Coverage is high on pure/logic packages (config/auth/domain/metrics/ws) and the HTTP layer;
  the **real govici adapter** is only exercised by the lab (`make lab-up`), its pure helpers by
  unit tests. Don't try to unit-test the govici methods without a socket.
- Add or update a test alongside every behavior change and run `go test ./...` before finishing.
