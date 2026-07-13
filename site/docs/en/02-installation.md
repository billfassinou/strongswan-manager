# Installation

Four paths, depending on what you are doing. The first one is for a real deployment; the last
one is to explore the product in five minutes.

| You want to… | Go to |
|---|---|
| **Install a server** (systemd) | [One-command install](#one-command-install) |
| **Build from the repository** | [Install from source](#install-from-source) |
| Use **apt / dnf** (upgrades included) | [.deb and .rpm packages](#deb-and-rpm-packages) |
| Install on a machine with **no Internet access** | [Offline install (air-gap)](#offline-install-air-gap) |
| **Try** the product, without installing anything for good | [Quick trial](#quick-trial) |

---

## One-command install

On a **Debian/Ubuntu** or **RHEL/AlmaLinux/Rocky** machine (amd64 or arm64), with systemd:

```bash
curl -fsSL https://raw.githubusercontent.com/billfassinou/strongswan-manager/main/deploy/install.sh | sudo bash
```

The script first shows you **what it is going to change**, then asks for confirmation. It:

1. installs **PostgreSQL** and **strongSwan** if they are missing;
2. creates the `swan` database and a `swanmgr` system user (no shell);
3. generates `/etc/strongswan-manager/strongswan-manager.env` with **random secrets**;
4. opens the VICI socket to the `swanmgr` group — this is what lets the console drive charon
   **without running as root**;
5. installs the systemd service, opens ports 7926/7927 in firewalld or ufw, and **checks that
   the console answers** before handing control back to you.

It ends by printing the console URL and the `admin` password.

> The script downloads the release archive, **verifies its SHA-256 digest** (and its `cosign`
> signature when the tool is available) and refuses to go on if it does not match.

Useful options:

| Option | Effect |
|---|---|
| `--no-strongswan` | Do not install strongSwan. For a console that only drives **remote** gateways. |
| `--skip-deps` | Install no packages at all. See [offline](#offline-install-air-gap). |
| `--version vX.Y.Z` | Install a specific version. |
| `--yes` | Ask nothing. |

---

## Install from source

You cloned the repository and want to install **your** build:

```bash
git clone https://github.com/billfassinou/strongswan-manager.git
cd strongswan-manager
sudo ./deploy/install.sh --from-source
```

The installer builds the web interface (it is **embedded** in the binary), then the Go binary,
then installs the result **exactly as the bundle would** — one single code path downstream, so
the same behaviour and the same checks.

Building requires **Go ≥ 1.23** and **Node ≥ 20**. If they are missing or too old — which is the
case for the Go shipped by AlmaLinux 9 — the installer fetches the **official toolchains**
(go.dev, nodejs.org) into a temporary directory and builds with those. **Nothing is installed
permanently on your machine**; the directory is gone when it finishes.

> `--no-strongswan`, `--skip-deps` and `--yes` apply to this mode too.

---

## .deb and .rpm packages

Every release ships native packages. Their point: `apt upgrade` / `dnf upgrade` update the
console like any other piece of software, **without touching your configuration or your
database**.

```bash
# Debian / Ubuntu
sudo apt install ./strongswan-manager_1.0.0_amd64.deb

# RHEL / AlmaLinux / Rocky
sudo dnf install ./strongswan-manager-1.0.0-1.x86_64.rpm
```

The post-install does the same work as the script: system user, database, secrets, service.
**Removing the package deletes neither the database nor the configuration** — you would have
to erase them explicitly (the removal message gives you the commands).

---

## Offline install (air-gap)

The `linux` release archives are not plain binaries: they are **self-contained bundles** with
the binary, the installer, `swanmgrctl` and the systemd unit. Nothing is downloaded during the
installation itself.

On a connected machine:

```bash
curl -LO https://github.com/billfassinou/strongswan-manager/releases/download/v1.0.0/strongswan-manager_v1.0.0_linux_amd64.tar.gz
curl -LO https://github.com/billfassinou/strongswan-manager/releases/download/v1.0.0/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing
```

Carry the archive over, then on the target machine:

```bash
tar xzf strongswan-manager_v1.0.0_linux_amd64.tar.gz
cd strongswan-manager_v1.0.0_linux_amd64
sudo ./install.sh --skip-deps
```

`--skip-deps` means "install no packages": **PostgreSQL must already be there** (from your
system image or a local mirror), otherwise the script stops and tells you so. Without that
option, the installer would try to reach the distribution repositories.

---

## Quick trial

> This is **not** a deployment mode: it is the **development lab**, meant for exploring the
> interface. To put the product into service, use one of the paths above.

You need the repository and Docker:

```bash
git clone <repo> && cd strongswan/backend
make run
# → https://localhost:7926, accounts admin/operator/auditor/viewer, password admin1234
```

This is **demo mode**: tunnels are **simulated**, no traffic is actually encrypted. Everything
else (create a tunnel, bring it up, watch the status change live) works. See
[Connect real gateways](14-connecter-passerelles-reelles.md).

---

## The end-of-install report

Whatever the mode, the installer does not merely note that packages are in place: it **opens the
connections**, exactly as the application will.

| Check | What is actually tested |
|---|---|
| **Service** | The console answers on `https://…:7926/healthz`. |
| **Database** | A connection is **opened** with the configured DSN, and migrations are counted. |
| **strongSwan (VICI)** | `swanctl` is invoked **as the service user** (`swanmgr`), not as root — which is exactly what the console does. A root test would pass where the service fails. |

### strongSwan unreachable ⇒ the install fails

If strongSwan is installed but the console cannot talk to it, **the install stops with an
error**. This is deliberate: without VICI the console would start in **demo mode** and display
**simulated** tunnels — giving the illusion of a VPN while **no traffic is encrypted at all**. An
install that stops is less dangerous than a console that lies.

The service and its configuration stay in place: there is **nothing to reinstall**. Fix the
reported point, then re-run the check:

```bash
swanmgrctl doctor
```

> Driving **remote** gateways only? Install with `--no-strongswan` and set `VICI_ENDPOINTS` in
> the configuration.

---

## First login

### The password must be changed

The four accounts (`admin`, `operator`, `auditor`, `viewer`) are created on first start with
**the same password** — the one the installer drew at random and wrote into the configuration
file. The console **forces you to change it**: until you do, the API answers `403` on
everything else. The lock lives in the server, not just in the screen.

The three other accounts share that password: deal with them too, or disable them.

### The browser warning

This is expected, and it is not a flaw: the application **generated its own certificate**,
signed by its internal CA. Your browser does not know that authority, so it warns you.

> **The connection really is encrypted.** What is not attested is the server's *identity* —
> not the confidentiality of the exchange. This is exactly how Proxmox, pfSense or TrueNAS
> behave on first start.

To go on: **Advanced** → **Proceed to…**

To make the warning go away for good:

| You want… | Do |
|---|---|
| To keep the internal CA | Import it into the trust store of your admin workstations (**PKI & Certificates** screen). **Once and for all.** |
| A **publicly trusted** certificate | `ACME_DOMAIN=vpn.example.com` → **Let's Encrypt**. Requires a **public domain** and **port 80 reachable**, forwarded to port 7927. |
| To use **your own** certificate | `TLS_CERT` and `TLS_KEY` |
| To terminate TLS on a **reverse proxy** | `TLS_ENABLED=false` |

Importing the CA on an admin workstation:

- **macOS**: double-click `ca.crt` → *System* keychain → "Always Trust".
- **Linux**: `sudo cp ca.crt /usr/local/share/ca-certificates/ && sudo update-ca-certificates`
- **Windows**: `certutil -addstore -f Root ca.crt` (as administrator).

⚠️ If you reach the console through a domain name, **declare it** in `TLS_SANS`
(e.g. `TLS_SANS="localhost,127.0.0.1,vpn.internal.example"`) — otherwise the browser will also
report a name mismatch.

---

## The two ports

| Port | Protocol | Serves |
|---|---|---|
| **7926** | **HTTPS** | The interface, the API, the live updates. |
| **7927** | Plain HTTP | Only the **CRL** (`/crl.der`) and `/healthz`. Everything else is redirected to HTTPS. |

Port 7927 is not an oversight: the CRL distribution point **must** stay on plain HTTP, because
it is charon that fetches it, and charon would not trust our internal CA — validating that
certificate would require the very CRL it is fetching (RFC 5280).

---

## Operating it: `swanmgrctl`

The installer also ships a lifecycle tool:

| Command | Effect |
|---|---|
| `swanmgrctl doctor` | Full diagnosis: database, VICI socket, certificate, ports, firewall. **The first thing to run when something looks wrong.** |
| `swanmgrctl backup` | Backs up the database **and** `SECRETS_KEY` into one archive. |
| `swanmgrctl restore ARCHIVE` | Restores. Refuses if the archive's key does not match this installation. |
| `swanmgrctl upgrade` | Backs up, upgrades, verifies, and **rolls back on its own** if the console stops answering. |
| `swanmgrctl status` / `logs` | systemd shortcuts. |

### Backups: the database alone is worthless

`SECRETS_KEY` encrypts your secrets, the **CA's private key** and the **TLS certificate's key**.
A PostgreSQL dump without that key is **permanently unreadable** — and the server would refuse
to even start. There is no recovery procedure: that is why `swanmgrctl backup` archives both
together, and why `restore` checks that they match before touching anything.

> The archive therefore holds `SECRETS_KEY` **in clear**. Treat it as a secret.

---

## Uninstalling

```bash
sudo /usr/share/strongswan-manager/uninstall.sh           # keeps database and configuration
sudo /usr/share/strongswan-manager/uninstall.sh --purge   # ⚠️ erases everything, irreversible
```

Without `--purge`, a later reinstall finds your tunnels, your PKI and your audit trail again.
PostgreSQL and strongSwan are never uninstalled (other services may depend on them).

---

## What next?

→ [Discover the interface](03-decouvrir-linterface.md)
