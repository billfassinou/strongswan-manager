# PKI & certificates

> Role required: **admin** or **operator** to issue/revoke. Everyone can view.

Certificate authentication is more robust than a PSK: every gateway has its own identity, and you can **revoke** a compromised identity without touching the others.

---

## The internal certificate authority

It is **created automatically on first startup** — there is nothing for you to do.

**Configuration → PKI & Certificates** (*Configuration → PKI & Certificats*) shows the authority (`StrongSwan Manager Root CA`, ECDSA key). Its private key is **encrypted at rest**, like the secrets.

---

## Issuing a certificate

One certificate per gateway.

1. **PKI & Certificates** → **Generate** (*Générer*).
2. Fill in:
   - **Name** (*Nom*): the internal identifier, e.g. `cert-gw-a`;
   - **Common Name (CN)**: e.g. `gw-a`;
   - **Usage** (*Usage*): `Server` (*Serveur*) (a gateway) or `Client` (*Client*);
   - **SAN**: ⚠️ **this is the important one** — put **the gateway's IP address**, e.g. `203.0.113.10`. Several values separated by commas (IPs or DNS names).
3. **Generate**.

> **Why the SAN is critical**: it is the IKE identity. The peer checks that the certificate presented really matches the address it is talking to. A SAN that does not match the tunnel endpoint ⇒ authentication fails.

The private key is generated, **encrypted** and stored. It is **never returned** by the API.

---

## Creating a certificate-authenticated tunnel

1. **Tunnel editor** (*Éditeur de tunnel*).
2. **Authentication** (*Authentification*) → `Certificate` (*Certificat*).
3. Choose the **Local certificate** (*Certificat local*) (your gateway's, whose SAN = your local endpoint).
4. If the peer is **also managed** by the console, choose its **Peer certificate** (*Certificat du pair*).
5. **Validate & apply** (*Valider & appliquer*).

On apply, the console loads onto the gateway:

- the **certificate authority** (so it knows how to validate the peer's certificate);
- the gateway's **certificate**;
- its **private key**.

It then configures the connection with public-key authentication, with the **IKE identities** set to the endpoint addresses.

You can then **Bring up** (*Monter*) the tunnel as usual.

---

## Revoking a certificate

1. **PKI & Certificates** → **Revoke** (*Révoquer*) on the row → confirm.
2. The certificate moves to the **Revoked** (*Révoqué*) state.
3. The **CRL is regenerated immediately** and signed by the authority.

---

## The CRL (revocation list)

### How gateways get it

StrongSwan has **no command to "push" a CRL**. The standard mechanism is the **CRL Distribution Point (CDP)**: a URL written **into the certificate**, which the gateway fetches by itself (charon's `curl` plugin).

That URL is configured with the **`CRL_URL`** variable, for example:

```
CRL_URL=http://my-server:7927/crl.der   # plain http, on the cleartext listener
```

The **`/crl.der`** endpoint is **public** (no authentication): this is expected — a CRL is a public object, and gateways must be able to read it without a token.

> ⚠️ **Already-issued certificates contain no CDP** if `CRL_URL` was empty when they were issued. Set `CRL_URL` **before** issuing your certificates.

### Available actions

| Button | Effect |
|---|---|
| **Publish the CRL** (*Publier la CRL*) | Regenerates the CRL (number incremented) and persists it |
| **Download the CRL (.der)** (*Télécharger la CRL (.der)*) | Retrieves the `/crl.der` file |

### When does a revocation take effect?

When the gateway **re-downloads** the CRL. It caches it until the `nextUpdate` date, set by **`CRL_VALIDITY`** (24 h by default).

- For fast turnaround in testing, lower `CRL_VALIDITY` (e.g. `30s` — that's what `make lab-up` does).
- In production, keep a reasonable duration (a few hours to a few days).

See [Troubleshooting](15-depannage.md) if a revoked certificate is still accepted.

---

## Checking the CRL yourself

```bash
curl -s http://localhost:7927/crl.der | openssl crl -inform DER -noout -text
```

You should see the CRL number and the revoked serial numbers.

---

## What does not exist yet

- **OCSP responder** (online checking, instead of the CRL).
- **SCEP/EST enrollment** (gateways request their certificate automatically).
- Importing an **external authority** (today the authority is internal).

---

## What's next?

→ [Site-to-site on both ends](08-site-a-site-deux-cotes.md) — establishing a real tunnel between two gateways you manage.
