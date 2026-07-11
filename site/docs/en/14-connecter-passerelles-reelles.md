# Connecting real gateways

By default, the application runs in **demo mode**: a single `gw-local` gateway backed by a **simulated VICI adapter**. Everything works (create a tunnel, bring it up, see its state) but no IPsec packet is actually encrypted.

This page explains how to drive **real StrongSwan daemons**.

---

## How the server talks to the gateways

It uses **VICI**, StrongSwan's official control interface (the `vici` plugin of `charon`). Concretely, the server opens the daemon's **VICI socket** and sends it commands:

| VICI command | Used for |
|---|---|
| `version` | Detecting the StrongSwan version |
| `load-conn` | Loading/updating a connection |
| `unload-conn` | Removing a connection |
| `load-shared` | Loading a PSK |
| `load-cert` / `load-key` | Loading a certificate and its key |
| `list-sas` | Reading the real state of the SAs (used by monitoring) |
| `initiate` / `terminate` / `rekey` | Bringing up, tearing down, renegotiating |

**No configuration file is ever written.**

---

## The lab: two real gateways in one command

```bash
cd backend
make lab-up
```

On top of the application, this starts **two strongSwan containers** (`strongswan-a`, `strongswan-b`) and connects the server to their VICI sockets.

You will then see **`gw-a`** and **`gw-b`** under **Gateways** (*Passerelles*), with the version reported by the daemon itself.

Now build a real tunnel between the two: [Site-to-site on both ends](08-site-a-site-deux-cotes.md).

To stop everything and clean up:

```bash
make lab-down
```

### Checking on the daemon side

```bash
docker compose exec strongswan-a swanctl --list-conns --uri unix:///vicirun/charon.vici
docker compose exec strongswan-a swanctl --list-sas   --uri unix:///vicirun/charon.vici
```

The connection you created in the interface must show up there — proof that it really was loaded through VICI.

---

## Declaring your own gateways

The server reads the **`VICI_ENDPOINTS`** variable, a comma-separated list of `name=endpoint` pairs:

```bash
VICI_ENDPOINTS="gw-paris=unix:/chemin/charon.vici,gw-dakar=tcp:10.0.0.5:4502"
```

| Form | When to use it |
|---|---|
| `unix:/path/to/charon.vici` | The daemon runs on the same machine (or its socket is shared, as in the lab) |
| `tcp:host:port` | The daemon exposes VICI over TCP |

Gateways are registered **at server startup**. If `VICI_ENDPOINTS` is empty, you fall back to demo mode.

### What the gateway needs

1. **StrongSwan ≥ 5.9** with the **`vici` plugin** loaded (which is the default with `strongswan-swanctl`).
2. The server must be able to **reach the VICI socket** — this is the trickiest part (see below).
3. For **certificate** authentication: the **`openssl`** plugin (the `libstrongswan-standard-plugins` package on Debian) is **required** — without it, charon cannot read an ECDSA certificate.

---

## Known limitations (read before deploying)

### Access to the VICI socket

The `charon.vici` socket is owned by `root` with `0770` permissions. In the lab, the server therefore runs **as root** and shares the socket through a Docker volume.

**This is not a production architecture.** The design calls for a **lightweight agent** installed on each gateway, bridging the local VICI socket and the server over **mTLS**. **That agent is not implemented yet.** For now, keep this mode to a trusted management network.

### StrongSwan version

Debian 12 ships **5.9.8**. That is above the minimum floor (5.9), but:

- **no post-quantum** (ML-KEM) — that needs StrongSwan **6.0+**;
- the interface flags `5.x` versions in orange under **Gateways**.

### Certificates and CRL

- Certificates are handed to charon in **DER** (the server converts from PEM): charon rejects raw PEM on these versions.
- StrongSwan **has no VICI command to load a CRL**. Revocation therefore goes through the **CRL Distribution Point** embedded in the certificates (`CRL_URL`), which the gateway downloads itself via its `curl` plugin. See [PKI & certificates](07-pki-certificats.md).

---

## What's next?

→ [Troubleshooting](15-depannage.md)
