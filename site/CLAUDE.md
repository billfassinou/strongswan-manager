# Published site & documentation (`site/`)

Repo-wide rules are in the root `CLAUDE.md`. Adding or changing a documentation page takes
**four** steps and forgetting one fails silently (see "A new page means…" below); if the local
`doc-page` skill is available, use it — it does all four and verifies them.

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

To read the docs as a user would (the viewer `fetch`es the Markdown, so `file://` fails):

```bash
python3 -m http.server -d site 8000   # → http://localhost:8000/docs/
```
