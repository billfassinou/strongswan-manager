# Overview

**StrongSwan Manager** is a web administration console for **StrongSwan**, the open-source IPsec VPN implementation.

It replaces what you do by hand today:

| Without the tool | With the tool |
|---|---|
| Editing `swanctl.conf` in a terminal, on every gateway | A guided form, validated before it is applied |
| `swanctl --load-all`, `--list-sas`, `--initiate`… on the command line | Buttons: **Bring up** (*Monter*), **Bring down** (*Couper*), **Rekey** (*Renégocier*) |
| No idea of the overall state of your estate | Real-time dashboard, topology, alerts |
| An IKEv1/3DES tunnel that survives for years | A **security score** that flags it, and a "Fix" (*Corriger*) button |
| No record of who changed what | A **tamper-proof audit log** |

> Key point: configuration is applied **through StrongSwan's VICI API** (its official control interface), never by writing configuration files by hand.

---

## Who it is for

### The network administrator
They create and operate tunnels: site-to-site interconnections, remote access, links to the cloud. They need speed and safety (not breaking a production tunnel).
→ Start with [Create a tunnel](04-creer-un-tunnel.md), then [Operate a tunnel](05-piloter-un-tunnel.md).

### The operator (operations, NOC)
They monitor, bring tunnels up and down, and troubleshoot incidents. They have the same write permissions as the administrator.
→ See [Monitoring](12-superviser.md) and [Troubleshooting](15-depannage.md).

### The auditor / CISO
They change nothing, but must **prove**: who changed what, what the cryptographic level of the estate is, which certificates are expiring.
→ See [Security score](09-score-de-securite.md) and [Auditing](13-auditer.md).

### The "read-only" profile
Monitoring only: they can view, they cannot change anything (the interface hides the actions, the API refuses them).

### The DevSecOps team
They want to automate: create tunnels from a pipeline, export metrics, integrate with their tooling.
→ See [REST & WebSocket API](A3-api.md).

The full breakdown of permissions is in [Roles & permissions](A1-roles-et-permissions.md).

---

## What the tool actually does

- **IPsec tunnels**: site-to-site, host-to-host, road warrior. IKEv1 and IKEv2. Authentication by **PSK**, **certificate** or **EAP**.
- **Hot apply**: the configuration is loaded into the `charon` daemon without downtime, after validation.
- **Operation**: bring up (`initiate`), bring down (`terminate`), rekey (`rekey`) a connection from the interface.
- **Built-in PKI**: internal certificate authority, X.509 certificate issuance, **revocation** and **CRL**.
- **Secret vault**: PSK/EAP/XAuth encrypted at rest; values are never shown again.
- **Security score**: every tunnel is graded; weak algorithms are flagged.
- **Versions & rollback**: every change creates a version; you can roll back.
- **Real-time monitoring**: tunnel state is pushed live (WebSocket); Prometheus metrics.
- **Tamper-proof audit**: *append-only* log, hash-chained.
- **StrongSwan modules**: virtual IP pools, RADIUS/AAA, policies & routing, authorities, VPN users, alert rules, daemon settings.

---

## What is not there (yet)

To be straight with you, here is what is **planned but not implemented** at this stage:

- **OCSP** responder and **SCEP/EST** enrollment (revocation works via CRL).
- **Remote mTLS agent** (today the server talks directly to the gateway's VICI socket).
- **Multi-tenant / SSO** SAML-OIDC (Enterprise edition).
- Replacing the application-level encryption of secrets with **Vault**.

---

## What's next?

1. [Installation](02-installation.md) — the application runs in three commands.
2. [Discover the interface](03-decouvrir-linterface.md) — the guided tour.
3. [Create a tunnel](04-creer-un-tunnel.md) — your first tunnel.
