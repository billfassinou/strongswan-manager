# Troubleshooting

---

## "Weak cryptographic proposal detected" (*Proposition cryptographique faible détectée*) — error 422

**What it means**: validation **rejected** your configuration. It was **not** applied to the gateway — you are not at risk.

**Most common cause**: `modp1024` in the IKE proposals (an obsolete Diffie-Hellman group).

**Fix**: replace it with `modp2048`, or better `ecp384`.

Other possible rejections: missing name, invalid IKE version, malformed CIDR, empty proposals. The message names the **exact field** at fault. See [Creating a tunnel](04-creer-un-tunnel.md).

---

## "VICI apply failed" (*application VICI échouée*) — error 502 `vici_error`

**What it means**: the configuration was valid, but the `charon` daemon **refused** to load it. The message repeats the error returned by the daemon.

| Message from charon | Cause | Fix |
|---|---|---|
| `parsing X509 certificate failed` | The gateway cannot read the certificate — often the **missing `openssl` plugin** (no ECDSA) | Install `libstrongswan-standard-plugins` on the gateway |
| `invalid certificate type 'crl'` | Attempt to push a CRL over VICI — **not possible** | Revocation goes through the CDP, see [PKI](07-pki-certificats.md) |
| an error on a proposal | The daemon does not know an algorithm (e.g. `mlkem768` on StrongSwan 5.9) | Check the **version** under **Gateways** (*Passerelles*) and remove the unsupported algorithm |

When the apply fails at creation time, the record is **rolled back**: the database and the gateway stay consistent.

---

## The tunnel stays `down` after "Bring up"

Work through this list, in order:

1. **Is the peer reachable?** IKE uses **UDP 500** and **UDP 4500**. A firewall blocking them is enough to break it.
2. **Are the proposals symmetric?** Both ends must share at least one common suite (IKE **and** ESP).
3. **Do the secret or the certificate match?**
   - PSK: the **same value** on both sides.
   - Certificate: the **SAN** must match the endpoint's address, and the authority must be known to the peer.
4. **Are the subnets swapped on the peer side?** Your "local network" is its "remote network".
5. **Is the clock** correct on both machines? A large drift makes certificate validation fail.

Then read the daemon logs:

```bash
docker compose logs strongswan-a | tail -40
```

The **diagnostic assistant** (Advanced → AI assistant & anomalies / *Avancé → Assistant & anomalies IA*) walks through this same list for a given tunnel.

---

## A gateway is `unknown` or `down`

The server cannot reach its VICI socket.

| Symptom in the server logs | Cause | Fix |
|---|---|---|
| `permission denied` on `charon.vici` | The socket is owned by `root` (`0770`), the server runs as another user | Run the server as root (this is what the lab does), or adjust the permissions |
| `connect: no such file or directory` | The socket path is wrong, or `charon` is not running | Check `VICI_ENDPOINTS` and the state of the daemon |
| `passerelle injoignable à l'enrôlement` ("gateway unreachable at enrollment") | The gateway was not ready when the server started | It will be retested at every poll; otherwise restart the server |

---

## A revoked certificate is still accepted

This is the **normal behaviour of charon's CRL cache**.

1. The gateway downloads the CRL through the **CDP** and **caches** it until `nextUpdate`.
2. As long as the cache is fresh, it does not download it again — and therefore does not see the new revocation.

**Checks:**

- Does the certificate actually contain a CDP? It only has one **if `CRL_URL` was set at issuance time**. If not: reissue the certificate.
- Can the gateway reach the URL?
  ```bash
  docker compose exec strongswan-a curl -s -o /dev/null -w '%{http_code}\n' http://backend:7927/crl.der
  ```
- Does the CRL actually list the certificate?
  ```bash
  curl -s http://localhost:7927/crl.der | openssl crl -inform DER -noout -text | grep -A2 Revoked
  ```
- Lower **`CRL_VALIDITY`** to speed up cache renewal (the lab uses `30s`).

---

## "Your connection is not private" / `NET::ERR_CERT_AUTHORITY_INVALID`

The certificate is **auto-generated**, signed by the application's internal CA — which your browser does not know. **The connection is encrypted**; it is the server's *identity* that is not attested.

Three ways out, from the quickest to the cleanest:

1. **Click through**: **Advanced** → **Proceed**. Acceptable locally, not in production.
2. **Import the internal CA** onto your admin workstations — the warning goes away for good. Steps in [Installation](02-installation.md).
3. **Get a real certificate**: `ACME_DOMAIN` (Let's Encrypt), or your own via `TLS_CERT`/`TLS_KEY`.

---

## `NET::ERR_CERT_COMMON_NAME_INVALID` — the name does not match

You are reaching the console through a name (`https://vpn.interne.fr:7926`) that is **not** in the certificate. By default, it only covers `localhost`, `127.0.0.1`, `::1` and the machine's hostname.

Declare the access name, then restart — the certificate will be **reissued automatically**:

```bash
TLS_SANS="localhost,127.0.0.1,vpn.interne.fr" docker compose up -d
```

---

## `curl` refuses to connect: `SSL certificate problem: self-signed certificate`

Expected, for the same reason. On the command line:

```bash
curl -sk https://localhost:7926/healthz               # -k : ignore la vérification
curl -s --cacert ca.crt https://localhost:7926/healthz  # mieux : valide contre la CA interne
```

---

## `http://localhost:7926` does not answer / answers "400 Bad Request"

Normal: **7926 is now an HTTPS port**. Use `https://localhost:7926`.

The **7927** listener is the plaintext one — but it only serves `/crl.der` and `/healthz`; everything else there is redirected with a 308 to HTTPS.

---

## Let's Encrypt (ACME) fails

The HTTP-01 challenge requires **the public port 80** to reach the application's plaintext listener.

- Publish **`80:7927`** (and `443:7926`) — the challenge arrives on port 80 on the public side.
- The domain in `ACME_DOMAIN` must **resolve publicly** to this machine.
- Mount `ACME_CACHE` **as a volume**: without it, every restart requests a new certificate and you will quickly hit Let's Encrypt's **rate limits** (5 failures/hour, then a lockout).
- ACME is **unusable on a private or air-gapped network** — use the internal CA instead.

---

## Port 7926 (or 5432) is already in use

`Bind for 0.0.0.0:7926 failed: port is already allocated`

Another service is holding the port. Either stop it, or change the published port in `docker-compose.yml`:

```yaml
ports: ["9090:7926"]
```

(PostgreSQL exposes **no port** on the host: the server reaches it over Docker's internal network.)

---

## The "real time" indicator stays grey

The WebSocket connection is not being established.

- Behind a **reverse proxy**, make sure it **forwards WebSockets** (`Upgrade` / `Connection`).
- Over HTTPS, the application automatically uses `wss://`.

> **Known security limitation**: the server only verifies the WebSocket connection's token **if one is provided**. A connection **without** a `?token=` parameter is currently accepted. Do not expose the application directly on the Internet without protection in front of it until this is fixed.

---

## I forgot the admin password

Accounts are only created **on first startup**, on an empty database. There is no reset screen yet.

Options:

- **In demo / development**: `docker compose down -v` (⚠️ wipes **all** data), then restart with a new `SEED_ADMIN_PASSWORD`.
- **In production**: update the bcrypt hash directly in the accounts table of the database.

---

## I can't find my answer

- [FAQ](16-faq.md)
- The logs: `make logs` (server) and `docker compose logs strongswan-a` (daemon)
