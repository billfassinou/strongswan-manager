# Managing secrets

> Role required: **admin** or **operator** to create/delete. Everyone can view the list (without the values).

The vault stores the shared secrets used by IPsec: **PSK** (pre-shared key), **EAP** and **XAuth**.

---

## Two rules to know

1. **Values are encrypted at rest** (AES-256-GCM), with a key derived from the `SECRETS_KEY` variable.
2. **A value is never shown again.** Not in the interface (`••••••••`), not through the API. Even an admin cannot read back a stored PSK. If you lost it, create a new one.

---

## Creating a PSK secret

1. **Configuration → Secrets** (*Configuration → Secrets*) → **+ Secret**.
2. Fill in:
   - **Name** (*Nom*): the identifier you will use in the tunnel, e.g. `psk-dakar`;
   - **Type** (*Type*): `PSK`;
   - **Value** (*Valeur*): the shared key (the same one configured on the far end);
   - **Used by** (*Utilisé par*): free text to help you keep track, e.g. `paris-dakar`.
3. **Save** (*Enregistrer*).

The secret appears in the list, with its value masked.

---

## Attaching the secret to a tunnel

1. Open the **Tunnel editor** (*Éditeur de tunnel*) (new tunnel, or **Edit** (*Éditer*) an existing one).
2. **Authentication** (*Authentification*) → `PSK`.
3. A **PSK secret** (*Secret PSK*) menu appears: pick `psk-dakar`.
4. **Validate & apply** (*Valider & appliquer*).

When applying, the console:

- loads the **connection** on the gateway (`load-conn`);
- decrypts the secret and loads it on the gateway (`load-shared`), binding it to the IKE identities (the addresses of both endpoints).

The PSK therefore only travels from the server to the `charon` daemon, never to your browser.

> If the tunnel is a **site-to-site managed on both ends**, the same PSK is loaded **on both gateways** automatically. See [Site-to-site on both ends](08-site-a-site-deux-cotes.md).

---

## Deleting a secret

**Secrets** → **✕** on the row → confirm.

> Warning: if a tunnel still references this secret, it will no longer be able to authenticate at the next negotiation. Check the **Used by** column before deleting.

---

## EAP and XAuth

The `EAP` and `XAuth` types are created in the same way. They are used for roaming access (road warriors), alongside **VPN users** (*Utilisateurs VPN*) — see [Configuration modules](11-modules-configuration.md).

---

## What happens if I change `SECRETS_KEY`?

All **already stored secrets become unreadable**: they were encrypted with the old key. The application will no longer be able to load them onto the gateways.

Only change this variable at install time, **before** creating a single secret. See [Installation](02-installation.md).

---

## What's next?

→ [PKI & certificates](07-pki-certificats.md) — the more robust alternative to PSK.
