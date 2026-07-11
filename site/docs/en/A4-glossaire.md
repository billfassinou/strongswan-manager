# Glossary

---

**IPsec** — The protocol suite that encrypts and authenticates IP traffic. It is what StrongSwan implements.

**IKE** *(Internet Key Exchange)* — The protocol that negotiates the tunnel's keys and parameters. Two versions: **IKEv1** (obsolete) and **IKEv2** (the one to use).

**SA** *(Security Association)* — The security "contract" negotiated between two peers. There is the **IKE SA** (the negotiation channel) and the **CHILD_SA** (the one that actually encrypts your data).

**CHILD_SA** — The child SA: it is the one that carries the traffic encryption (ESP). When the interface shows `up`, it means it is installed.

**ESP** *(Encapsulating Security Payload)* — The protocol that encrypts the packets once the tunnel is up.

**charon** — StrongSwan's daemon. It negotiates and maintains the tunnels.

**VICI** — StrongSwan's official control interface. It is what the console uses to drive `charon` (load a connection, bring up a tunnel, read the state). **No configuration file is ever written.**

**swanctl** — StrongSwan's command-line tool. It uses VICI, just like the console. `swanctl --list-sas` is used to check what the console did.

---

**PSK** *(Pre-Shared Key)* — A secret key shared by both ends. Simple, but everyone shares the same secret.

**X.509 certificate** — An identity signed by an authority. Each gateway has its own: you can revoke one without touching the others. More robust than a PSK.

**CA** *(Certificate Authority)* — The authority that signs the certificates. The console ships with one, created automatically.

**SAN** *(Subject Alternative Name)* — The certificate field that carries the network identity (an IP, a DNS name). **It must match the address of the tunnel endpoint**, otherwise authentication fails.

**CRL** *(Certificate Revocation List)* — The signed list of revoked certificates.

**CDP** *(CRL Distribution Point)* — The URL, embedded **in the certificate**, from which the gateway downloads the CRL. This is how revocation propagates (StrongSwan has no command to "push" a CRL).

**OCSP** — An online revocation check, instead of the CRL. *Not implemented yet.*

---

**PFS** *(Perfect Forward Secrecy)* — The property guaranteeing that an attacker who has stolen a key will not be able to decrypt **past** exchanges. Always leave it enabled.

**Diffie-Hellman / DH group** — The key exchange algorithm. `modp1024` is **weak**; `modp2048` is acceptable; `ecp384` is recommended.

**Proposal** — A complete cryptographic suite, written with dashes: `aes256-sha256-modp2048` = AES-256 encryption, SHA-256 hash, modp2048 key exchange. Both peers must have at least one proposal **in common**.

**AES-GCM** — Modern authenticated encryption (`aes256gcm16`), recommended for ESP.

**ML-KEM** — **Post-quantum** key exchange (`mlkem768`). Protects against a future quantum computer. **Requires StrongSwan 6.0+**.

**3DES / MD5 / modp1024** — **Obsolete** algorithms. Their presence drags the security score down. Replace them.

---

**Site-to-site** — A tunnel between two networks (two branch offices, for example). Both ends have a fixed address.

**Host-to-host** — A tunnel between two machines, with no network behind them.

**Road warrior** — A roaming user, whose address is dynamic. Typically connects with EAP or a certificate.

**EAP / XAuth** — User authentication methods (often backed by a RADIUS server).

**RADIUS** — Authentication and accounting server (AAA) to which roaming access can be delegated.

**Virtual IP pool** — The address range handed out to roaming clients once connected.

**Split tunnel** — Sending only part of the traffic through the tunnel (the corporate networks), the rest going straight out to the Internet.

---

**Rekey** — Renegotiation of a tunnel's keys, without interrupting traffic.

**Trap / on-demand** — A policy that brings the tunnel up automatically on the first packet.

**Shunt** — An exception policy: let traffic through (`pass`) or block it (`drop`) **without** encrypting it.

**XFRM / VTI** — Network interfaces enabling a "route-based" VPN (you route into an interface, instead of defining policies per subnet).

---

**JWT** — The authentication token issued at login, to be attached to every API call.

**RBAC** — Role-based access control (`admin`, `operator`, `auditor`, `viewer`).

**Append-only** — Said of the audit log: you can only **add** to it. The database refuses any modification or deletion.

**Prometheus** — The metrics format exposed on `/metrics`, to plug into your monitoring.
