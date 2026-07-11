# Create a tunnel

> Required role: **administrator** or **operator**.

---

## How it works

You fill in a form, and the application:

1. **validates** what you typed (a dangerous configuration is rejected, not applied);
2. computes a **security score**;
3. **translates** the configuration into a VICI instruction and **hot-loads** it into the `charon` daemon;
4. records a **version** and an **audit entry**.

At no point is a configuration file written by hand.

---

## Step by step

### 1. Open the editor

Click **+ New tunnel** (*+ Nouveau tunnel*, top right), or go to **Configuration → Tunnel editor** (*Éditeur de tunnel*).

### 2. Fill in the fields

| Field | What goes in it | Example |
|---|---|---|
| **Name** (*Nom*) | A name that is unique **per gateway** | `paris-dakar` |
| **Gateway** (*Passerelle*) | The StrongSwan gateway that will host the tunnel | `gw-local` |
| **Peer gateway** (*Passerelle pair*) | *(optional)* the other end, if it is **also managed** by the console | — |
| **Type** | `Site-to-site`, `Host-to-host` or `Road warrior` | Site-to-site |
| **IKE version** (*Version IKE*) | `IKEv2` (recommended) or `IKEv1` | IKEv2 |
| **Local endpoint** (*Extrémité locale*) | The public IP address of **your** gateway | `203.0.113.10` |
| **Local network** (*Réseau local*) | The protected subnet(s) on your side | `10.1.0.0/16` |
| **Remote endpoint** (*Extrémité distante*) | The peer's IP address | `198.51.100.20` |
| **Remote network** (*Réseau distant*) | The peer's protected subnet(s) | `10.2.0.0/16` |
| **Authentication** (*Authentification*) | `PSK`, `Certificate` or `EAP` | PSK |
| **IKE proposals** (*Propositions IKE*) | The phase 1 cipher suite | `aes256-sha256-modp2048` |
| **ESP proposals** (*Propositions ESP*) | The cipher suite for the traffic | `aes256gcm16` |
| **Perfect Forward Secrecy** | Leave it **ticked** | ✔ |

Separate multiple values with commas (networks, proposals).

### 3. Watch the score

On the right, the **score ring** recomputes **live** as you type. The findings are listed below it: *"No post-quantum readiness"*, *"Weak 3DES/DES encryption"*, and so on.

That is your safety net: you see the quality of the configuration **before** applying it.

### 4. Validate & apply

Click **Validate & apply** (*Valider & appliquer*).

- ✅ **Success**: a message gives you the score and the version (`Créé · score 94 · v1`), and you are taken back to the Connections list.
- ❌ **Error**: see below.

---

## The three tunnel types

### Site-to-site
Two networks linked across the Internet. Both ends have a fixed address. This is the most common case.
If **both gateways are managed** by the console, see [Site-to-site on both ends](08-site-a-site-deux-cotes.md).

### Host-to-host
Two machines talking directly to each other (no network behind them). Use the host's address as the protected network (e.g. `192.0.2.44/32`).

### Road warrior
Roaming clients connecting from anywhere. The remote endpoint is **dynamic**: leave "Remote endpoint" empty; authentication is typically done with **EAP** or with a **certificate**. Users are managed under **VPN users** (*Utilisateurs VPN*) (see [Configuration modules](11-modules-configuration.md)).

---

## When validation rejects your input (error 422)

The application **blocks** a configuration that is invalid or plainly dangerous. The message tells you exactly what to fix.

**Try it out**: set `aes256-sha256-modp1024` as the IKE proposal, then apply. You get:

> `Proposition cryptographique faible détectée — modp1024 déconseillé (DH groupe 2)`

Other common rejections:

| Message | Cause | Fix |
|---|---|---|
| `nom requis` | The Name field is empty | Give the tunnel a name |
| `ike_version doit valoir 1 ou 2` | Invalid IKE version | Choose IKEv2 |
| `au moins une proposition IKE est requise` | The proposals field is empty | Enter a suite, e.g. `aes256-sha256-ecp384` |
| `adresse locale requise` | Local endpoint is empty | Enter the gateway's IP |
| `CIDR/adresse invalide: …` | A network is malformed | Use CIDR notation (`10.1.0.0/16`) |
| `un tunnel de ce nom existe déjà sur cette passerelle` | Duplicate | Change the name |

**Important**: when validation fails, **nothing has been applied** on the gateway. You can fix it and try again with no risk.

---

## Recommended cipher suites

| Use case | IKE proposal | ESP proposal |
|---|---|---|
| Modern (recommended) | `aes256gcm16-sha384-ecp384-mlkem768` | `aes256gcm16` |
| Classic, highly compatible | `aes256-sha256-modp2048` | `aes256gcm16` |
| **Avoid** | `3des-md5-modp1024` | `3des-md5` |

`mlkem768` (post-quantum) requires **StrongSwan 6.0 or later** on the gateway. See [Security score](09-score-de-securite.md).

---

## What's next?

The tunnel is created but not necessarily established. → [Operate a tunnel](05-piloter-un-tunnel.md)
