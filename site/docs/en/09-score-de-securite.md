# Security score

Every tunnel gets a score **out of 100**, computed from its cryptographic configuration. It shows up everywhere: dashboard, connection list, editor (live), Security page.

---

## How it is computed

Start from **100** and subtract:

| Finding | Penalty |
|---|---|
| **IKEv1** (obsolete protocol) | −42 |
| **3DES / DES** (weak encryption) | −28 |
| **MD5** (weak digest) | −18 |
| **modp1024 / modp768** (weak Diffie-Hellman group) | −16 |
| **Perfect Forward Secrecy disabled** | −10 |
| **No ML-KEM** (no post-quantum preparation) | −6 |

The result is clamped between **5** and **100**.

The same computation runs on the server side (on save) and in the browser (live in the editor): you see the score move **as you type**.

---

## Reading the colour

| Score | Colour | Meaning |
|---|---|---|
| **≥ 85** | green | Compliant with best practices |
| **65 – 84** | orange | Acceptable but improvable (often: no ML-KEM) |
| **< 65** | red | **Critical** — fix as a priority |

Real examples:

- `aes256gcm16-sha384-ecp384-mlkem768` + IKEv2 + PFS → **100**
- `aes256-sha256-modp2048` + IKEv2 + PFS → **94** (only post-quantum is missing)
- `3des-md5-modp1024` + IKEv1 → **5** (everything you shouldn't do, all at once)

---

## Seeing the state of your estate

**Advanced → Security & Compliance** (*Avancé → Sécurité & Conformité*):

- a **ring** gives the **average score across the estate**;
- three counters: compliant / to harden / critical;
- a table lists **every weak algorithm detected**, tunnel by tunnel.

---

## Hardening a tunnel

In the weak-algorithms table, click **Fix** (*Corriger*): the **editor opens pre-filled** with the tunnel concerned.

Replace the proposals:

| Before | After |
|---|---|
| `3des-md5-modp1024` | `aes256gcm16-sha384-ecp384` |
| IKEv1 | **IKEv2** |
| PFS unchecked | PFS **checked** |

The score updates live. **Validate & apply** (*Valider & appliquer*): the configuration is hot-reloaded and a new version is created.

> The tunnel will be **renegotiated**: expect a short interruption, and make sure **the peer accepts the new proposals** — otherwise the tunnel will not come back up. You can always [roll back](10-versions-et-rollback.md).

---

## About ML-KEM (post-quantum)

`mlkem768` protects the key exchange against a future quantum computer. It is the only thing that keeps a classic tunnel from reaching 100.

**Careful**: ML-KEM requires **StrongSwan 6.0 or later** on the gateway. Current distributions (Debian 12, for example) ship **5.9.8**, which does not support it. Check the version under **Gateways** (*Passerelles*) before enabling it — otherwise negotiation will fail.

A **−6** penalty is therefore **acceptable** on a 5.9.x gateway: it is not an exploitable weakness today, it is preparation.

---

## What's next?

→ [Versions & rollback](10-versions-et-rollback.md) — fixing things without risk.
