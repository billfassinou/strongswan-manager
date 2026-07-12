# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

It holds only what must be true **everywhere**. The rest lives next to the code it describes,
and is loaded when you work there:

| File | Covers |
|---|---|
| `backend/CLAUDE.md` | The Go backend: stack, invariants, VICI/PKI/TLS, testing. |
| `backend/web/CLAUDE.md` | The React SPA embedded in the binary. |
| `deploy/CLAUDE.md` | The installer, the packages, `swanmgrctl`. |
| `site/CLAUDE.md` | The published website and the user documentation. |

Local tooling (kept out of the repo, under a gitignored `.claude/`): hooks enforce the
invariants below, and skills carry the repeated workflows (`doc-page`, `release`). The rules
here hold regardless — the hooks just catch slips early; CI is the backstop.

## Nature of this repository

A **web management interface for StrongSwan** (open-source IPsec VPN), targeting network
admins, MSPs, and DevSecOps teams — the equivalent of a graphical console (à la
Fortinet FortiManager / Palo Alto Panorama) built on StrongSwan's VICI API. The repo
holds the **specification**, an interactive **front-end mock**, and now a **Go backend**
(first vertical slice / walking skeleton) under `backend/`.

Repo layout (what is **published** at github.com/billfassinou/strongswan-manager):
`backend/` (the app) · `deploy/` (the installer & lifecycle tooling) · `site/` (the published
website, docs included) · `.github/workflows/` · `LICENSE` (AGPL-3.0) · `CLA.md` +
`CONTRIBUTING.md`. Root `README.md` is the repo's public front page: online site, docs, and
release downloads.

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

This is the one irreversible accident in this project (a public history cannot be un-published):
never stage it, including with `git add -f` (which would bypass the gitignore). A local hook
blocks it, but do not rely on the hook being present — treat it as an absolute rule.

- `spec/description.rtf` — the original intent brief (French), the **source of truth** for scope.
- `spec/cahier_des_charges.md` — the functional & technical specification (16 sections, French).
  **This `.md` is the editable source of truth for the spec.**
- `spec/cahier_des_charges.docx` — Word version, **generated from the `.md`**. Never edit it
  directly; it must always be derived from the `.md`.
- `spec/app.html` — the interactive mock; the **visual reference** for the React front, and the
  origin of the `scoreTunnel` scoring algorithm.

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
