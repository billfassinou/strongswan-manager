# Frontend (`backend/web/`) — the embedded React SPA

Backend rules are in `backend/CLAUDE.md`; repo-wide rules in the root `CLAUDE.md`.

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
- **First login is blocking**: when `/me.must_change_password` is true, `App.tsx` renders
  `pages/ChangePassword.tsx` and nothing else. This mirrors the server, which answers 403 on
  every route but `/me` and `/me/password` — so don't "fix" the front by routing around it.
  The login screen must **not** pre-fill any credentials.
- Keep the SPA's design tokens (`web/src/styles.css`) consistent with the reference mock, and
  with `site/assets/styles.css` (the website mirrors the same tokens).
- **Generic config modules**: Pools, RADIUS, Policies, Authorities, VPN users, alert rules and
  daemon settings are all backed by one **generic CRUD** over a `config_items` table (`kind` +
  JSONB), exposed at `GET/POST /api/v1/config/{kind}` and `PUT/DELETE /api/v1/config/{kind}/{id}`
  (`internal/httpapi/handlers_config.go`, `store.ConfigRepo`, `configStore` interface). Allowed
  kinds: `pool, radius, policy, authority, vpnuser, alert, daemon`. To add a config module, add
  the kind to `configKinds` (backend) and a schema in `web/src/schemas.ts` (front `Crud` page) —
  no new table/handler needed. Topology and the AI assistant are front-only, computed from real
  `/tunnels` + `/gateways` data (no fake ML).
