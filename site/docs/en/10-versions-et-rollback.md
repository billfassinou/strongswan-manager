# Versions & rollback

Every time a tunnel configuration is applied, the console **records a version**: a complete snapshot, timestamped and attributed to its author.

This is your safety net: you can change a tunnel in production knowing that **rolling back takes a second**.

---

## When is a version created?

| Action | Version |
|---|---|
| Create a tunnel | **v1** |
| Modify it | v2, v3, … |
| Perform a rollback | a **new** version (v4) that restores the content of an older one |

A rollback therefore never erases history: it **adds** a version. The timeline stays complete and auditable.

---

## Browsing the history

A tunnel's history is available through the API:

```bash
curl -s http://localhost:7926/api/v1/tunnels/<ID>/versions \
  -H "Authorization: Bearer $TOKEN" | jq
```

Each entry carries its number (`n`), its message (`création`, `mise à jour`, `rollback vers v1`), its author and its date.

> The dedicated history screen is not exposed in the interface yet: versions are browsed and restored through the API. The server-side feature, however, is complete.

---

## Rolling back to an earlier version

```bash
curl -s -X POST http://localhost:7926/api/v1/tunnels/<ID>/rollback \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"version": 1}'
```

What the console does:

1. it reads back the snapshot of the requested version;
2. it **recomputes the score**;
3. it **hot-reloads the configuration** on the gateway (and on the peer gateway, if the tunnel is managed on both ends);
4. it creates a new version (`rollback vers v1`) and an **audit entry**.

If you omit `version`, the **previous version** is the one restored.

---

## Typical use case

You harden an old tunnel (score 5 → 100). The peer, a legacy device, does not accept AES-GCM: the tunnel does not come back up.

1. **Roll back** to the previous version → the tunnel comes back up immediately.
2. You agree on a maintenance window with the team on the other side.
3. You reapply the hardening on both ends.

No time lost, no configuration to rebuild by hand.

---

## What is not versioned

- **Secrets** and **certificates** (they live in the vault and the PKI, with their own lifecycle).
- **Configuration modules** (pools, RADIUS, policies…): they are persisted, but without history.
- **Deleting** a tunnel: it takes its versions with it. For a reversible interruption, use **Bring down** (*Couper*) rather than **Delete** (*Supprimer*).

---

## What's next?

→ [Configuration modules](11-modules-configuration.md)
