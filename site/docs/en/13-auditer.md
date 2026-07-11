# Auditing

**Monitoring → Audit log** (*Supervision → Journal d'audit*) — accessible to **all roles**, including `auditor` and `viewer`.

---

## What is recorded

Every significant action writes an entry: **who**, **what**, **on what**, **when**.

| Action | Recorded when… |
|---|---|
| `login` | Someone signs in |
| `tunnel.create` / `tunnel.update` / `tunnel.delete` | A tunnel is created, modified, deleted |
| `tunnel.initiate` / `tunnel.terminate` / `tunnel.rekey` | A tunnel is brought up, torn down, renegotiated |
| `tunnel.rollback` | A previous configuration is restored |
| `secret.create` / `secret.delete` | A secret is created or deleted |
| `cert.issue` / `cert.revoke` | A certificate is issued or revoked |
| `crl.publish` | The CRL is regenerated |
| `config.<kind>.create` / `config.update` / `config.delete` | A configuration module is modified |

The page refreshes automatically every 5 seconds.

---

## Why this log is trustworthy

### It is *append-only*

The database **refuses** any modification or deletion of an audit entry — not by convention, but through a **PostgreSQL trigger**. An `UPDATE` or `DELETE` attempt raises an error, even when run directly in SQL by a database administrator.

This is verified by the project's integration tests.

### It is chained

Each entry carries an **integrity hash** computed from the hash of the previous entry. In other words: you cannot remove or rewrite a row in the middle without breaking the chain — tampering becomes **detectable**.

---

## Making use of the log

### In the interface

The page shows the last 100 entries: timestamp, action, target.

### Through the API (for an export, a SIEM…)

```bash
curl -sk "https://localhost:7926/api/v1/audit?limit=500" \
  -H "Authorization: Bearer $TOKEN" | jq
```

Each entry returns `id`, `actor_id`, `action`, `target`, `timestamp`, `prev_hash`, `integrity_hash`.

This lets you:

- feed a **SIEM** (Splunk, Elastic…);
- produce an **evidence export** for an ISO 27001 / PCI-DSS audit;
- verify the hash chain yourself.

---

## Answering an auditor's questions

| Question | Where to find the answer |
|---|---|
| "Who modified this tunnel, and when?" | Audit log (`tunnel.update`) + the tunnel's version history |
| "What is the cryptographic level of the estate?" | **Security & Compliance** (*Sécurité & Conformité*) — average score, weak algorithms |
| "Was this compromised certificate revoked?" | **PKI & Certificates** (*PKI & Certificats*) (state `Revoked` / *Révoqué*) + `cert.revoke` in the audit log |
| "Can a trace be erased?" | No — the database refuses it (*append-only* trigger) |
| "Who has write access?" | [Roles & permissions](A1-roles-et-permissions.md) |

---

## What does not exist yet

Turnkey **compliance reports** (ANSSI, ISO 27001, PCI-DSS) and periodic **executive reports** belong to the **Premium** edition. The raw data, however, is already there and exportable.

---

## What's next?

→ [Connecting real gateways](14-connecter-passerelles-reelles.md)
