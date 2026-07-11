# Site-to-site managed on both ends

> Role required: **admin** or **operator**.

When **both gateways** of a site-to-site tunnel are managed by the console (typical case: your own sites, or an MSP administering both endpoints), you can configure **both ends in a single operation**.

---

## What the console does

You fill in a **Peer gateway** (*Passerelle pair*). On apply, the console:

1. loads the connection on **your** gateway;
2. loads the **mirror connection** on the **peer** gateway — that is, the same connection with the **endpoints and networks swapped**;
3. loads the **PSK secret on both**, or **each certificate on its own gateway**.

Result: both daemons are configured consistently. All that's left is to establish the SA.

---

## Step by step — with a PSK

1. **Secrets** → create a PSK, e.g. `psk-hub` (see [Managing secrets](06-secrets.md)).
2. **Tunnel editor** (*Éditeur de tunnel*):
   - **Name** (*Nom*): `hub`
   - **Gateway** (*Passerelle*): `gw-a`
   - **Peer gateway** (*Passerelle pair*): `gw-b` ← *this is the key part*
   - **Local endpoint** (*Extrémité locale*): the IP of `gw-a`
   - **Local network** (*Réseau local*): the network behind `gw-a`, e.g. `192.168.10.0/24`
   - **Remote endpoint** (*Extrémité distante*): the IP of `gw-b`
   - **Remote network** (*Réseau distant*): the network behind `gw-b`, e.g. `192.168.20.0/24`
   - **Authentication** (*Authentification*): `PSK` → **PSK secret** (*Secret PSK*): `psk-hub`
   - **Proposals** (*Propositions*): `aes256-sha256-modp2048` / `aes256gcm16`
3. **Validate & apply** (*Valider & appliquer*).
4. **Connections** (*Connexions*) → **Bring up** (*Monter*).

**What you should see**: the tunnel goes to `negotiating` and then **`up`** within a few seconds.

---

## Step by step — with certificates

1. **PKI & Certificates** (*PKI & Certificats*) → issue **two** certificates:
   - `cert-a`, CN `gw-a`, **SAN = the IP of gw-a**
   - `cert-b`, CN `gw-b`, **SAN = the IP of gw-b**
2. **Tunnel editor**: same fields as above, but
   - **Authentication**: `Certificate` (*Certificat*)
   - **Local certificate** (*Certificat local*): `cert-a`
   - **Peer certificate** (*Certificat du pair*): `cert-b`
3. **Validate & apply**, then **Bring up**.

The console loads the authority + `cert-a` + its key on `gw-a`, and the authority + `cert-b` + its key on `gw-b`.

> ⚠️ The **SANs must match the tunnel endpoint addresses**. This is the number-one mistake.

---

## Checking on the daemon side

If you are using the lab (see [Connecting real gateways](14-connecter-passerelles-reelles.md)):

```bash
docker compose exec strongswan-a swanctl --list-sas --uri unix:///vicirun/charon.vici
```

You should read something like:

```
hub: #1, ESTABLISHED, IKEv2
  hub-net: #1, INSTALLED, TUNNEL, ESP:AES_GCM_16-256
```

`ESTABLISHED` = the IKE SA is up. `INSTALLED` = the child SA (the actual traffic encryption) is installed in the kernel. The tunnel really works.

---

## Deleting

Deleting the tunnel **unloads both connections** (yours and the mirror). Nothing to clean up by hand.

---

## What if the peer is not managed by the console?

Leave **Peer gateway** empty. You only configure **your** side; the other endpoint (a FortiGate, a Cisco, an unmanaged StrongSwan…) must be configured elsewhere, **with symmetric parameters**: same proposals, same PSK or trusted certificates, swapped subnets.

---

## What's next?

→ [Security score](09-score-de-securite.md)
