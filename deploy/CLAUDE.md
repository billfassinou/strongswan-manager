# Installer & lifecycle (`deploy/`) — published, and load-bearing

`deploy/` is the supported way to install the product. Repo-wide rules (SPDX header on every
file, including these scripts) are in the root `CLAUDE.md`.

- **`lib/common.sh` is the single source of truth** for install logic (distro/arch detection,
  `swanmgr` user, PostgreSQL provisioning, env-file generation, VICI drop-in, firewall,
  healthcheck). `install.sh`, `uninstall.sh`, `swanmgrctl` and the `.deb`/`.rpm` scriptlets all
  source it. Add logic there, not in a copy.
- **The release tarball is a self-contained bundle** (binary + `install.sh` + `swanmgrctl` +
  unit + `lib/`). `install.sh` detects whether it runs *from* a bundle: if not (the
  `curl | sudo bash` one-liner), it downloads the bundle, verifies SHA-256 (+ cosign when
  available) and **re-execs itself from inside it** — so the offline path is the only real code
  path, and is therefore always exercised. `--skip-deps` makes it touch no network at all.
- **Never regenerate `SECRETS_KEY` if the env file exists.** It decrypts the secrets, the CA key
  and the TLS key; losing it makes the database — and every backup of it — permanently
  unreadable. `swanmgrctl backup` therefore archives the DB **and** the env file together, and
  `restore` refuses an archive whose key fingerprint doesn't match (`--adopt-key` to override).
- **Package scriptlets must distinguish upgrade from removal** (dpkg passes `upgrade`, rpm
  passes `1`). A `prerm` that disables the service unconditionally leaves it stopped **and
  disabled** after every `apt upgrade`; a `postrm` that removes the VICI drop-in on upgrade cuts
  the console off from charon. Both are guarded — don't undo that.
- **`systemd` unit**: no `Requires=postgresql.service` in the unit itself (it would break a
  remote database); `install.sh` drops in `10-local-postgres.conf` only when the DB is local.
- **nfpm does not expand env vars in `contents.src`** (it does in `arch`/`version`): the
  workflow copies the built binary to a fixed path (`backend/dist/pkg/`) instead. nfpm also
  resolves `contents` paths against the **current directory**, so the packaging step must run
  from the repo root.
- **Shell style**: one cleanup `trap` only, in `lib/common.sh` (`make_tmpdir`). A function that
  sets its own `trap … EXIT` on a *local* variable sees it vanish before the trap runs
  (`unbound variable` under `set -u`) and clobbers its caller's trap.
- `ci.yml` really installs the bundle on the Ubuntu runner (systemd + apt) and asserts the
  403-until-password-changed behaviour, backup/restore, idempotent reinstall and uninstall. If
  you change `deploy/`, that job is what tells you it still works. Locally:
  `shellcheck -x --source-path=deploy -e SC1091 deploy/*.sh deploy/swanmgrctl deploy/lib/common.sh deploy/scripts/*.sh`
