# Discover the interface

You are logged in. Here is what you are looking at.

---

## The sidebar: four groups

| Group | Modules |
|---|---|
| **Monitoring** (*Supervision*) | Dashboard · Connections · Topology · Monitoring & alerts · Audit log |
| **Configuration** | Tunnel editor · PKI & Certificates · Secrets · VPN users |
| **StrongSwan** | Pools & virtual IPs · RADIUS / AAA · Policies & routing · Authorities & enrollment · Daemon settings |
| **Advanced** (*Avancé*) | Security & Compliance · AI assistant & anomalies · Gateways & ZTP · Administration |

A **red badge** appears on **Connections** (*Connexions*) as soon as a tunnel goes *down*: that is your immediate alert.

At the bottom of the sidebar: your identity, your role, and the **logout** button.

---

## The top bar

- The **title** of the current page and your identity/role.
- **◑**: toggles **light / dark theme** (your choice is remembered).
- **+ New tunnel** (*+ Nouveau tunnel*): a shortcut to the editor. This button **does not appear** if your role is read-only.

---

## The dashboard

This is the home page. On it you will find:

- **Four counters**: active tunnels, negotiating, *down*, number of gateways.
- **The list of tunnels** with their state and their **security score**.
- **The list of gateways** with their StrongSwan version and their state.
- A **"real time"** indicator: when it is green, the interface is receiving state changes live (WebSocket). You do not need to refresh the page — when a tunnel comes up, the row updates on its own.

---

## What you can do depending on your role

| You are… | You see | You can change |
|---|---|---|
| **Administrator** (`admin`) | Everything | **Everything** |
| **Operator** (`operator`) | Everything | **Everything** |
| **Auditor** (`auditor`) | Everything | Nothing (buttons hidden) |
| **Read-only** (`viewer`) | Everything | Nothing |

In read-only mode, the action buttons (create, edit, delete, bring up, bring down…) are **hidden**. And if someone were to call the API directly, the server would answer **403**. Read-only is a real guarantee, not just window dressing.

Full details: [Roles & permissions](A1-roles-et-permissions.md).

---

## Tunnel states

You will see them everywhere (coloured dots):

| State | Colour | Meaning |
|---|---|---|
| `up` | green | The SA is established, the tunnel is working |
| `negotiating` | orange | IKE negotiation / rekeying in progress |
| `installing` | orange | The configuration has just been applied, not established yet |
| `down` | red | No SA established |
| `unknown` | grey | The gateway could not be queried |

---

## The security score

Every tunnel shows a score out of 100, colour-coded:

- **≥ 85** green — in line with best practice
- **65 – 84** orange — needs hardening
- **< 65** red — critical

Clicking it does nothing: the details are in **Security & Compliance** (*Sécurité & Conformité*), with a **"Fix"** (*Corriger*) button that opens the editor pre-filled. See [Security score](09-score-de-securite.md).

---

## What's next?

→ [Create your first tunnel](04-creer-un-tunnel.md)
