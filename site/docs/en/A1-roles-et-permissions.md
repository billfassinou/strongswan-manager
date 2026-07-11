# Roles & permissions

Four roles. The rule is simple: **`admin` and `operator` can write, `auditor` and `viewer` cannot.**

---

## The roles

| Role | Read | Write | Intended for |
|---|---|---|---|
| `admin` | ✅ | **✅** | Console administrator |
| `operator` | ✅ | **✅** | Operations, NOC |
| `auditor` | ✅ | ❌ | CISO, audit, compliance |
| `viewer` | ✅ | ❌ | Monitoring, level 1 support |

The four corresponding accounts are created on first startup. See [Installation](02-installation.md).

---

## What "read-only" really guarantees

Two barriers, not one:

1. **The interface hides** the action buttons (create, edit, delete, bring up, tear down, generate, revoke…).
2. **The server refuses.** Every modifying route is protected: a direct API call with an `auditor` or `viewer` token gets:

```json
403 Forbidden
{"error":"forbidden","message":"action réservée — rôle en lecture seule"}
```

A read-only user therefore cannot change **anything**, even by bypassing the interface.

---

## Who can do what

| Action | admin | operator | auditor | viewer |
|---|:---:|:---:|:---:|:---:|
| View the dashboard, the topology, the gateways | ✅ | ✅ | ✅ | ✅ |
| Read tunnels, secrets (masked), certificates, audit log | ✅ | ✅ | ✅ | ✅ |
| Create / modify / delete a tunnel | ✅ | ✅ | ❌ | ❌ |
| Bring up / tear down / rekey a tunnel | ✅ | ✅ | ❌ | ❌ |
| Roll back a configuration | ✅ | ✅ | ❌ | ❌ |
| Create / delete a secret | ✅ | ✅ | ❌ | ❌ |
| Issue / revoke a certificate, publish the CRL | ✅ | ✅ | ❌ | ❌ |
| Modify the configuration modules (pools, RADIUS…) | ✅ | ✅ | ❌ | ❌ |

> **Note**: `admin` and `operator` currently have **the same rights**. The distinction exists to prepare for finer granularity (Enterprise edition), and for traceability: the audit log records who acted.

---

## Authentication

- **Password** hashed with **bcrypt**.
- On login, the server issues a **JWT** (HS256) carrying the identity and the role, valid for **1 hour** by default (`JWT_TTL`).
- The token is sent on every call: `Authorization: Bearer <token>`.
- The front end keeps it in the browser's `localStorage`.

Without a valid token, every protected route answers **401**.

---

## Known limitation — WebSocket

The real-time stream (`/api/v1/ws`) accepts a token as a parameter (`?token=`), but **only verifies it when it is present**: a connection **without** a token is currently accepted.

**Consequence**: anyone with access to the server's network can observe tunnel state changes without authenticating.

**Mitigation**: do not expose the application directly on the Internet; put it behind a reverse proxy that filters access. A fix is planned.

---

## Account management

There is **not yet** a screen for creating/disabling accounts, nor SSO (these functions belong to the **Enterprise** edition). The four accounts are seeded on first startup; any change currently goes through the database.

---

## See also

- [Getting to know the interface](03-decouvrir-linterface.md)
- [REST & WebSocket API](A3-api.md)
