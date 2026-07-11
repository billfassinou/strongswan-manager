# Monitoring

Every page in this chapter is accessible to **all roles**, including read-only ones.

---

## The dashboard

**Monitoring → Dashboard** (*Supervision → Tableau de bord*)

- **Counters**: tunnels up / negotiating / down, number of gateways.
- **List of tunnels** with state and score.
- **List of gateways** with their StrongSwan version.
- A **"real time"** (*temps réel*) indicator.

### Real time, concretely

The server polls each gateway **every 3 seconds** (tunable via `POLL_INTERVAL`) to read the actual state of the SAs. As soon as a state changes, it pushes it to your browser over **WebSocket**.

Practical consequence: **do not refresh the page**. Bring a tunnel up, watch the row: it goes `installing` → `negotiating` → `up` all by itself.

If the indicator is **grey**, the real-time connection has dropped (the browser retries automatically every 3 seconds).

---

## The topology

**Monitoring → Topology** (*Supervision → Topologie*)

A graphical view of your estate:

- each **gateway** is a node, surrounded by its tunnels;
- each **tunnel** is an edge, **coloured by its state** (green = up, dashed orange = negotiating, dashed red = down);
- a tunnel whose peer is **not managed** by the console appears as a small satellite.

This is computed from real data (`/gateways` + `/tunnels`): the map is always up to date, there is nothing to fill in.

Below it: counters for gateways / healthy links / negotiating / down.

---

## The gateways

**Advanced → Gateways & ZTP** (*Avancé → Passerelles & ZTP*)

The list of managed StrongSwan daemons:

| Column | What it tells you |
|---|---|
| **Name** (*Nom*) | `gw-a`, `gw-local`… |
| **VICI endpoint** (*Endpoint VICI*) | How the server reaches it (`unix:/gw/a/charon.vici`, or `mock`) |
| **Version** (*Version*) | The StrongSwan version **reported by the daemon itself**. Shown in orange for `5.x` (no post-quantum). |
| **State** (*État*) | `up` if the gateway answers VICI queries |

A gateway that is `down` or `unknown` means the server cannot reach it. See [Troubleshooting](15-depannage.md).

---

## Anomalies & diagnostic assistant

**Advanced → AI assistant & anomalies** (*Avancé → Assistant & anomalies IA*)

### The anomalies

They are **derived from the real state** of your estate, with no guesswork:

- a tunnel that is **down** → *critical*;
- a tunnel in **prolonged negotiation** → *to watch*;
- a tunnel with a **score < 65** → *to watch*.

### The assistant

Ask a question in French; the assistant answers **from your real tunnels**:

- « **pourquoi paris-dakar est down ?** » ("why is paris-dakar down?") → it identifies the tunnel, recalls its state and score, and lists the likely causes (peer reachability on UDP 500/4500, NTP clock, asymmetric proposals/PSK).
- « **quels tunnels sont à durcir ?** » ("which tunnels need hardening?") → it lists the ones whose score is critical and tells you the fix.
- « **quels tunnels sont down ?** » ("which tunnels are down?") → it enumerates them.

> **Honesty**: this is **not** a language model. It is a **deterministic rules engine** applied to your data. It cannot make things up, but it will not understand a question outside its scope. A true AI assistant belongs to the Premium edition.

---

## Prometheus metrics

The **`/metrics`** endpoint (public, no authentication) exposes:

| Metric | Type | Meaning |
|---|---|---|
| `strongswan_tunnel_status` | gauge | Per tunnel: `1` = up, `0.5` = negotiating, `0` = down |
| `strongswan_vici_errors_total` | counter | Number of VICI communication errors |

… plus the standard Go process metrics.

Point Prometheus at it, then Grafana or Alertmanager:

```yaml
scrape_configs:
  - job_name: strongswan-manager
    static_configs:
      - targets: ['mon-serveur:7926']
```

Today this is **the way to get alerted** (the built-in notification channels are not active yet — see [Configuration modules](11-modules-configuration.md)).

---

## What's next?

→ [Auditing](13-auditer.md)
