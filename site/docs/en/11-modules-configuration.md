# Configuration modules

> Role required: **admin** or **operator** to make changes. Everyone can view.

These modules cover StrongSwan features beyond the tunnels themselves. They **all work the same way**: a list, an add button, an edit dialog, a delete action. Entries are **persisted in the database** and **audited**.

---

## Pools & virtual IPs

**StrongSwan → Pools & virtual IPs** (*StrongSwan → Pools & IP virtuelles*)

The address ranges handed out to roaming clients, and the attributes pushed along with them.

| Field | Example |
|---|---|
| **Name** (*Nom*) | `pool-roadwarrior` |
| **Range** (*Plage*) | `10.9.0.0/24` |
| **Source** (*Source*) | `Internal` (*Interne*), `SQL`, `DHCP` or `RADIUS` |
| **Pushed DNS** (*DNS poussé*) | `10.1.0.53` |
| **Split-tunnel** (*Split-tunnel*) | `10.0.0.0/8` (the networks to route through the tunnel) |

---

## RADIUS / AAA

**StrongSwan → RADIUS / AAA**

The RADIUS servers to which EAP authentication and accounting are delegated.

| Field | Example |
|---|---|
| **Name** (*Nom*) | `radius-primaire` |
| **Address:port** (*Adresse:port*) | `10.1.0.20:1812` |
| **Role** (*Rôle*) | `Primary` (*Primaire*) or `Secondary` (*Secondaire*) |
| **Accounting** | enabled / disabled |

---

## Policies & routing

**StrongSwan → Policies & routing** (*StrongSwan → Politiques & routage*)

| Type | What it is for |
|---|---|
| **shunt** | Exclude given traffic from the tunnel (`pass`) or block it (`drop`). Typical use: don't encrypt the management LAN. |
| **trap** | Bring the tunnel up **on the first packet** (on demand). |
| **route-based** | Route traffic into a dedicated interface (XFRM/VTI) instead of per-subnet policies. |

| Field | Example |
|---|---|
| **Name** (*Nom*) | `bypass-management` |
| **Type** (*Type*) | `shunt` |
| **Detail** (*Détail*) | `pass 10.1.0.0/24` |
| **Interface** (*Interface*) | `xfrm1` (for route-based) |

---

## Authorities & enrollment

**StrongSwan → Authorities & enrollment** (*StrongSwan → Autorités & enrôlement*)

Declare **external** certificate authorities and their verification endpoints.

| Field | Example |
|---|---|
| **Name** (*Nom*) | `Partner Sub-CA` |
| **CRL URI** (*URI CRL*) | `http://crl.partner.io/sub.crl` |
| **OCSP URI** (*URI OCSP*) | `http://ocsp.partner.io` |
| **Enrollment** (*Enrôlement*) | `—`, `SCEP` or `EST` |

> These entries are **stored**, but automatic SCEP/EST enrollment is **not yet performed** by the server. The **internal** authority (the one that issues your certificates) is managed in [PKI & certificates](07-pki-certificats.md).

---

## VPN users

**Configuration → VPN users** (*Configuration → Utilisateurs VPN*)

Roaming access (road warriors).

| Field | Example |
|---|---|
| **Identity** (*Identité*) | `marc.diallo` |
| **Method** (*Méthode*) | `EAP-TLS`, `Certificate` (*Certificat*) or `EAP-MSCHAPv2` |
| **Quota** (*Quota*) | `50 GB` |
| **Time window** (*Plage horaire*) | `08:00–20:00` |
| **Active** (*Actif*) | yes / no |

Disabling a user takes effect immediately. To cut off access for certain when the user authenticates with a **certificate**, revoke it as well in the [PKI](07-pki-certificats.md).

---

## Monitoring & alerts

**Monitoring → Monitoring & alerts** (*Supervision → Monitoring & alertes*)

The alert rules.

| Field | Example |
|---|---|
| **Name** (*Nom*) | `Tunnel down` (*Chute de tunnel*) |
| **Condition** (*Condition*) | `state = down` |
| **Channels** (*Canaux*) | `Email, Slack` |
| **Enabled** (*Activée*) | yes / no |

> The rules are **stored**, but the **actual delivery** of notifications (email, Slack, Telegram, webhook) is not implemented yet. In the meantime, metrics are exposed in **Prometheus** format on `/metrics`: you can plug in the alerting stack of your choice (Alertmanager, Grafana). See [Monitoring](12-superviser.md).

---

## Daemon settings

**StrongSwan → Daemon settings** (*StrongSwan → Paramètres du démon*)

The global settings of `charon` (the equivalent of `strongswan.conf`).

| Setting | Default | Effect |
|---|---|---|
| **Worker threads** (*Threads worker*) | 16 | Number of processing threads |
| **IKE retransmits** (*Retransmissions IKE*) | 5 | Number of attempts before giving up |
| **Retransmit timeout** (*Timeout retransmission*) | 4 s | Delay between two attempts |
| **Assigned DNS** (*DNS attribué*) | `10.1.0.53` | DNS pushed to clients |
| **install_routes** | enabled | Does the daemon install the routes? |
| **IKEv2 fragmentation** (*Fragmentation IKEv2*) | enabled | Fragmentation of IKE messages (useful behind a reduced MTU) |
| **IKE log level** (*Niveau de log IKE*) | 1 (audit) | Verbosity of the IKE subsystem |

Click **Validate & reload** (*Valider & recharger*) to save.

> These settings are **persisted** and replayable; propagating them to a real `strongswan.conf` on the gateway is not automated yet.

---

## Under the hood (for the curious)

All of these modules share **a single mechanism** on the server side: a generic table where each entry carries a **type** (`kind`) and its fields as JSON, exposed on `/api/v1/config/{kind}`.

The available `kind` values: `pool`, `radius`, `policy`, `authority`, `vpnuser`, `alert`, `daemon`.

That is what makes it possible to add a module without a new table or new server code. See [API](A3-api.md).

---

## What's next?

→ [Monitoring](12-superviser.md)
