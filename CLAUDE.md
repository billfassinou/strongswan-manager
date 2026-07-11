# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Nature of this repository

A **web management interface for StrongSwan** (open-source IPsec VPN), targeting network
admins, MSPs, and DevSecOps teams — the equivalent of a graphical console (à la
Fortinet FortiManager / Palo Alto Panorama) built on StrongSwan's VICI API. The repo
holds the **specification**, an interactive **front-end mock**, and now a **Go backend**
(first vertical slice / walking skeleton) under `backend/`.

Repo layout (what is **published** at github.com/billfassinou/strongswan-manager):
`backend/` (the app) · `site/` (the published website, docs included) · `.github/workflows/` ·
`LICENSE` (AGPL-3.0) · `CLA.md` + `CONTRIBUTING.md`.

## Open-core: the licence boundary is a repo boundary

The core (this repo) is **AGPL-3.0**; the Premium/Enterprise modules (compliance, advanced
alerting, AI, multi-tenant, SSO) are **commercial and live in a separate private repo** — never
add them here. Two consequences to respect:

- **Every source file carries an SPDX header** (`// SPDX-License-Identifier: AGPL-3.0-or-later`).
  New files must too. In Go files with a `//go:build` tag, the header goes **after** the build
  directive (otherwise the tag stops applying).
- **Contributions require a signed-off commit** (the CLA, `CLA.md`): without it the project
  could not relicense a contribution into the commercial edition. Do not merge unsigned work.

## `spec/` — local only, NEVER commit it

`spec/` exists **on disk but not in the repo**: the whole folder is gitignored, and its history
was purged. It is the private working material. **Never `git add` it back**, and never make a
tracked file link to a path inside it (a public reader would hit a dead link) — refer to its
contents by name instead ("the cahier des charges", "the reference mock").

- `spec/description.rtf` — the original intent brief (French), the **source of truth** for scope.
- `spec/cahier_des_charges.md` — the functional & technical specification (16 sections, French).
  **This `.md` is the editable source of truth for the spec.**
- `spec/cahier_des_charges.docx` — Word version, **generated from the `.md`**. Never edit it
  directly; it must always be derived from the `.md`.
- `spec/app.html` — the interactive mock; the **visual reference** for the React front, and the
  origin of the `scoreTunnel` scoring algorithm.

## Published site & documentation (`site/`)

**`site/` IS the web root.** GitHub Actions uploads it verbatim to GitHub Pages
(`.github/workflows/pages.yml`, `path: ./site`) → **https://billfassinou.github.io/strongswan-manager/**.
The local path and the public URL are the same shape: `site/docs/` ⇢ `/docs/`.

- `site/index.html` (FR) + `site/en/index.html` (EN) — the showcase. Static, no build, no CDN,
  **zero external requests** (must stay renderable air-gapped). See `site/README.md`.
- `site/docs/` — the **user documentation**: task-based Markdown pages (`01-…` → `16-…`, plus
  annexes `A1-…` → `A4-…`) rendered by a self-contained viewer in `site/docs/index.html`
  (hash routing, `fetch`es the `.md` — so it must be **served**, not opened via `file://`).
  `site/docs/en/` is the English mirror and **keeps the same French filenames** — the viewer's
  nav and every cross-link depend on that. A new page means: write FR + EN, then add it to
  the `PARTS` array in **both** `site/docs/index.html` and `site/docs/en/index.html`.
- The docs are **user-facing** (installation → everyday tasks, per role); technical
  reference lives in the annexes. Everything cited (screen, button, route, env var, make
  target) must actually exist — what is not implemented is stated as such.
- **`site/.nojekyll` must not be deleted**: Jekyll would compile the `.md` files to HTML and the
  viewer (which fetches raw Markdown) would break.
- `site/assets/styles.css` mirrors the product's design tokens (`backend/web/src/styles.css`)
  and is shared by the docs viewer. Keep the two in sync.
- **SEO**: `canonical` + `hreflang` + Open Graph are hardcoded in the 4 HTML shells, and
  `site/sitemap.xml` lists 4 URLs. If the site address changes, update them together
  (`grep -rl billfassinou.github.io site/`). Known gap: the docs are client-side rendered, so
  the 40 pages are **not individually indexable** (a pre-render pass is the pending fix).
- Root `README.md` is the repo's public front page: online site, docs, and release downloads.

## Working on the specification

The spec was produced with the `sft` skill (Générateur de cahier des charges). When
iterating:

- Make **surgical edits to `spec/cahier_des_charges.md`** (find the exact passage, edit it) —
  do not rewrite the whole document for a targeted change.
- If a change touches a term referenced in several places (e.g. a renamed section or an
  EF-xx requirement id), `grep` for it and update all occurrences for consistency.
- **Regenerate the `.docx` in full** after any `.md` change:
  ```bash
  python3 ~/.claude/skills/sft/scripts/md_to_docx.py \
    --input spec/cahier_des_charges.md --output spec/cahier_des_charges.docx
  ```
  (Add `--logo path/to/logo.png` if a logo is provided.) The script builds the cover page
  from the first `# ` title, the bold subtitle line, and the first metadata table; it
  replaces any "Table des matières" section with a real Word TOC field.
- If the `.docx` is open in Word the script fails to save — close it first.

## Spec conventions (keep these when editing)

- **Traceability markers**: content added beyond the intent brief is marked
  **🆕** (recommandation basée sur les standards du marché). Preserve this distinction —
  never present a 🆕 recommendation as a validated requirement.
- **Points de vigilance** (near the top): open arbitrations are tracked as **V1, V2, …**
  and referenced from the body. Keep them in sync when a decision is made.
- **Functional requirements** are numbered **EF-01 … EF-20**, each priced **P1 (MVP) /
  P2 (Premium) / P3 (IA & multi-tenant)** — this mapping drives the roadmap (§12) and the
  traceability matrix (§15.4). Changing a priority means updating both.
- **Do not name internal artifacts** (source file names, internal mockup names) anywhere
  in the spec — describe features in business language. External references (standards,
  StrongSwan docs, RFCs, competitors) *should* stay cited (§16.4).
- Diagrams are **Mermaid** blocks (flowchart / erDiagram / gantt). Convert to real PNGs
  only if the user explicitly asks for "vrais schémas".

## Key domain decisions already recorded (see §5–§7 of the spec)

These shape any future implementation and should not be silently contradicted:
- **VICI is the primary integration path** with StrongSwan (live state, hot reload via
  `swanctl --load-all`); generating `swanctl.conf` files is only for export/versioning
  (GitOps) and as a fallback for old versions. Never edit config files by hand.
- Target compatibility: StrongSwan **≥ 5.9** (floor), **6.0+ recommended** (post-quantum
  ML-KEM, RFC 9370 multiple key exchanges).
- Recommended stack: **Go** backend + remote agent, **React/TypeScript** SPA,
  **PostgreSQL**, **Prometheus** (+TimescaleDB), **Vault** for secrets, Docker/Helm.
- **Open-core** model: some modules (multi-tenant, IA, compliance) are premium — module
  boundaries must coincide with license boundaries.
- **Air-gapped** deployment must remain possible: the anomaly-detection engine runs
  locally; the LLM-based assistant is optional and disableable.

## Backend (`backend/`) — Go vertical slice

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
  files — V3); the security score is the Go port of `scoreTunnel` from `spec/app.html` (keep both
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
- **Out of scope for this slice** (later increments): OCSP responder, SCEP/EST enrollment
  (rest of EF-24), Vault (replace the app-level cipher), remote mTLS agent,
  pools/RADIUS/policies/daemon-params, IA, multi-tenant/SSO, TimescaleDB, Helm.

### Frontend (`backend/web/`)

- **React + TypeScript (Vite)** SPA, **served by the Go backend at the same origin** as the
  API (so JWT/WebSocket work without CORS). It is built to `web/dist` and **embedded** via
  `web/embed.go` (`//go:embed all:dist`); `web.Handler()` serves `/`, `/assets/*` and
  SPA-falls-back unknown routes to `index.html`. It's mounted as the chi catch-all `/*` after
  the API routes (`API.SPA` field, set in `main`).
- **`web/dist` must exist for `go build`** (embed). `make web` (npm install + build) produces
  it; the Dockerfile builds it in a `node` stage so `make run`/`lab-up` are self-contained.
  `web/node_modules` and `web/dist` are gitignored and `.dockerignore`d.
- The SPA talks to the API via `src/api.ts` (fetch + Bearer JWT in localStorage), `src/ws.ts`
  (WebSocket `/api/v1/ws?token=` for live tunnel status), and hides write actions when
  `/me.can_write` is false. Dev-only: `cd web && npm run dev` proxies `/api` to `:7926`.
- Keep the SPA's design tokens (`web/src/styles.css`) consistent with the mock `spec/app.html`.
- **Generic config modules**: Pools, RADIUS, Policies, Authorities, VPN users, alert rules and
  daemon settings are all backed by one **generic CRUD** over a `config_items` table (`kind` +
  JSONB), exposed at `GET/POST /api/v1/config/{kind}` and `PUT/DELETE /api/v1/config/{kind}/{id}`
  (`internal/httpapi/handlers_config.go`, `store.ConfigRepo`, `configStore` interface). Allowed
  kinds: `pool, radius, policy, authority, vpnuser, alert, daemon`. To add a config module, add
  the kind to `configKinds` (backend) and a schema in `web/src/schemas.ts` (front `Crud` page) —
  no new table/handler needed. Topology and the AI assistant are front-only, computed from real
  `/tunnels` + `/gateways` data (no fake ML).

### Testing

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
