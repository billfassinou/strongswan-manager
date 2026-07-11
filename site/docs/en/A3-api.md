# REST & WebSocket API

Everything the interface does is available through the API. This is the way to **automate**: CI/CD pipelines, scripts, integration into your own tooling.

- Base: `/api/v1`
- Format: JSON
- Interactive documentation: **`/api/v1/docs`** (Swagger UI) · raw specification: **`/api/v1/openapi.yaml`**

---

## Authenticating

```bash
TOKEN=$(curl -sk https://localhost:7926/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identity":"admin","password":"admin1234"}' | jq -r .token)
```

Then, on every call:

```bash
curl -sk https://localhost:7926/api/v1/me -H "Authorization: Bearer $TOKEN"
```

The token expires after `JWT_TTL` (1 h by default).

---

## Rights

| Response | Meaning |
|---|---|
| **401** | Token missing, invalid or expired |
| **403** | Read-only role (`auditor`, `viewer`) on a modifying route |
| **422** | Business validation failed (see below) |

See [Roles & permissions](A1-roles-et-permissions.md).

---

## All routes

### Public (no authentication)

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Liveness probe |
| GET | `/metrics` | Prometheus metrics |
| GET | `/crl.der` | **CRL in DER** — the distribution point read by the gateways |
| POST | `/api/v1/auth/login` | Authentication → JWT token |
| GET | `/api/v1/openapi.yaml` | OpenAPI specification |
| GET | `/api/v1/docs` | Interactive documentation (Swagger UI) |
| GET | `/api/v1/ws` | Real-time stream (WebSocket), token in `?token=` |

### Read (token required — all roles)

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/me` | Profile: identity, role, `can_write` |
| GET | `/api/v1/gateways` | Gateways |
| GET | `/api/v1/tunnels` | Tunnels (live state included) |
| GET | `/api/v1/tunnels/{id}` | Details of a tunnel |
| GET | `/api/v1/tunnels/{id}/versions` | Configuration history |
| GET | `/api/v1/secrets` | Secrets (**values masked**) |
| GET | `/api/v1/certificates` | Certificates (**without private key**) |
| GET | `/api/v1/ca` | Certificate authority (public PEM) |
| GET | `/api/v1/crl` | CRL in PEM format |
| GET | `/api/v1/config/{kind}` | Items of a configuration module |
| GET | `/api/v1/audit` | Audit log (`?limit=`) |

### Write (token required — `admin` or `operator`; otherwise **403**)

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/tunnels` | Create a tunnel (validates, scores, applies via VICI) |
| PUT | `/api/v1/tunnels/{id}` | Modify |
| DELETE | `/api/v1/tunnels/{id}` | Delete (unloads the connection) |
| POST | `/api/v1/tunnels/{id}/initiate` | Bring up |
| POST | `/api/v1/tunnels/{id}/terminate` | Tear down |
| POST | `/api/v1/tunnels/{id}/rekey` | Rekey |
| POST | `/api/v1/tunnels/{id}/rollback` | Restore a version |
| POST | `/api/v1/secrets` | Create a secret (encrypted at rest) |
| DELETE | `/api/v1/secrets/{id}` | Delete a secret |
| POST | `/api/v1/certificates` | Issue a certificate |
| POST | `/api/v1/certificates/{id}/revoke` | Revoke (regenerates the CRL) |
| POST | `/api/v1/crl/publish` | Regenerate the CRL |
| POST | `/api/v1/config/{kind}` | Create a configuration item |
| PUT | `/api/v1/config/{kind}/{id}` | Modify |
| DELETE | `/api/v1/config/{kind}/{id}` | Delete |

`{kind}` ∈ `pool`, `radius`, `policy`, `authority`, `vpnuser`, `alert`, `daemon`. An unknown `kind` returns **404** `unknown_kind`.

---

## Creating a tunnel

```bash
curl -sk https://localhost:7926/api/v1/tunnels \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "paris-dakar",
    "gateway_id": "<UUID de la passerelle>",
    "type": "site-to-site",
    "ike_version": 2,
    "local":  { "addr": "203.0.113.10",  "subnets": ["10.1.0.0/16"] },
    "remote": { "addr": "198.51.100.20", "subnets": ["10.2.0.0/16"] },
    "auth":   { "method": "psk", "secret_ref": "psk-dakar" },
    "proposals": { "ike": ["aes256-sha256-modp2048"], "esp": ["aes256gcm16"] },
    "pfs": true
  }'
```

`201` response:

```json
{
  "id": "…",
  "name": "paris-dakar",
  "status": "installing",
  "security_score": 94,
  "warnings": ["Pas de préparation post-quantique (ML-KEM)"],
  "config_version": 1
}
```

Useful optional fields:

| Field | Effect |
|---|---|
| `peer_gateway_id` | Also configures the **peer gateway** (mirror connection) |
| `peer_cert_ref` | Peer certificate (certificate authentication on both sides) |
| `auth.cert_ref` | Local certificate (with `"method": "cert"`) |

---

## Error format

All errors follow the same structure:

```json
{
  "error": "validation_failed",
  "message": "Proposition cryptographique faible détectée",
  "details": [
    { "field": "proposals.ike", "issue": "modp1024 déconseillé (DH groupe 2)" }
  ],
  "correlation_id": "…"
}
```

> The server's `message` and `issue` strings are emitted in **French** — this is the app's
> actual output, shown here verbatim. Match on the machine-readable `error` code
> (`validation_failed`, `forbidden`, `unknown_kind`, `conflict`…), not on the prose.

`correlation_id` lets you find the request again in the server logs.

---

## The configuration modules

A single schema for all of them:

```bash
# créer un pool
curl -sk https://localhost:7926/api/v1/config/pool \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"pool-rw","data":{"range":"10.9.0.0/24","source":"Interne","dns":"10.1.0.53"}}'

# lister
curl -sk https://localhost:7926/api/v1/config/pool -H "Authorization: Bearer $TOKEN"

# modifier / supprimer
curl -sk -X PUT    https://localhost:7926/api/v1/config/pool/<ID> -H "Authorization: Bearer $TOKEN" \
     -H 'Content-Type: application/json' -d '{"name":"pool-rw","data":{"range":"10.9.0.0/22"}}'
curl -sk -X DELETE https://localhost:7926/api/v1/config/pool/<ID> -H "Authorization: Bearer $TOKEN"
```

The `data` field is free-form (JSON): each module puts its own fields in it.

---

## The real-time stream (WebSocket)

```
ws://localhost:7926/api/v1/ws?token=<jeton>
```

Every state change emits a message:

```json
{ "type": "tunnel_status", "id": "…", "name": "paris-dakar", "status": "up" }
```

> **Limitation**: the token is verified **only if it is supplied**. A connection without `?token=` is currently accepted. See [Roles & permissions](A1-roles-et-permissions.md).

---

## Complete example: a tunnel end to end

```bash
B=https://localhost:7926
TOKEN=$(curl -sk $B/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"identity":"admin","password":"admin1234"}' | jq -r .token)
GW=$(curl -sk $B/api/v1/gateways -H "Authorization: Bearer $TOKEN" | jq -r '.items[0].id')

# secret
curl -sk $B/api/v1/secrets -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"psk-demo","type":"psk","value":"clef-partagee"}' >/dev/null

# tunnel
TID=$(curl -sk $B/api/v1/tunnels -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{
  "name":"demo","gateway_id":"'"$GW"'","type":"site-to-site","ike_version":2,
  "local":{"addr":"203.0.113.10","subnets":["10.1.0.0/16"]},
  "remote":{"addr":"198.51.100.20","subnets":["10.2.0.0/16"]},
  "auth":{"method":"psk","secret_ref":"psk-demo"},
  "proposals":{"ike":["aes256-sha256-modp2048"],"esp":["aes256gcm16"]},"pfs":true}' | jq -r .id)

# monter, puis lire l'état
curl -sk -X POST $B/api/v1/tunnels/$TID/initiate -H "Authorization: Bearer $TOKEN"
sleep 4
curl -sk $B/api/v1/tunnels/$TID -H "Authorization: Bearer $TOKEN" | jq '{name, status, security_score}'
```
