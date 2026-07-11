# Operate a tunnel

> Required role: **administrator** or **operator**.

Go to **Monitoring → Connections** (*Supervision → Connexions*). Each row carries four actions.

---

## Bringing a tunnel up

Click **Bring up** (*Monter*).

The console asks `charon` to establish the connection (VICI command `initiate`). The tunnel then goes:

`installing` → `negotiating` → **`up`**

You have **nothing to refresh**: the server polls the state of the SAs continuously and pushes the change to your browser (WebSocket). The status dot changes on its own, usually within a few seconds.

**If the tunnel stays `down`**, see [Troubleshooting](15-depannage.md) — it is nearly always the peer not answering, or an asymmetric configuration.

---

## Bringing a tunnel down

Click **Bring down** (*Couper*) (`terminate`). The SA is destroyed and the tunnel goes back to `down`.

The **connection stays configured** on the gateway: you can bring it back up with one click. Bringing down ≠ deleting.

---

## Rekeying

The API exposes a **`rekey`** action that forces a key renegotiation without interrupting traffic. It is available through the API (`POST /api/v1/tunnels/{id}/rekey`) — see [API](A3-api.md).

---

## Editing a tunnel

Click **Edit** (*Éditer*): the editor opens **pre-filled**.

Make your changes, then **Validate & apply** (*Valider & appliquer*):

- the new configuration is **validated** then **hot-reloaded** on the gateway;
- a **new version** is created (v2, v3…);
- the score is recomputed.

If you made a mistake, you can roll back: [Versions & rollback](10-versions-et-rollback.md).

---

## Deleting a tunnel

Click **✕**, then confirm.

The application **unloads** the connection from the gateway (`unload-conn`) **before** deleting the record. If the peer gateway was managed too, the mirror connection is unloaded as well.

> ⚠️ Deletion is **permanent**: the tunnel disappears from the database, and its versions with it. To interrupt a tunnel temporarily, use **Bring down**.

---

## Understanding the states

| State | What is actually happening | What to do |
|---|---|---|
| `installing` | The configuration has just been loaded into `charon`, no SA is up yet | Click **Bring up** |
| `negotiating` | An IKE exchange is in progress (establishment or rekeying) | Wait a few seconds |
| `up` | The IKE SA **and** the child SA (ESP) are established: traffic is flowing | Nothing |
| `down` | No SA. Either you have not brought the tunnel up, or negotiation is failing | **Bring up**, otherwise [Troubleshooting](15-depannage.md) |
| `unknown` | The gateway did not answer the VICI query | Check the gateway |

The state shown **comes from the daemon**: it is read periodically over VICI (`list-sas`); it is not a declarative state stored in the database.

---

## What's next?

- Does the tunnel use a PSK? → [Manage secrets](06-secrets.md)
- Want certificate-based authentication? → [PKI & certificates](07-pki-certificats.md)
- Are both gateways yours? → [Site-to-site on both ends](08-site-a-site-deux-cotes.md)
