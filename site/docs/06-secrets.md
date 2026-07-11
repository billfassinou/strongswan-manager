# Gérer les secrets

> Rôle requis : **administrateur** ou **opérateur** pour créer/supprimer. Tout le monde peut consulter la liste (sans les valeurs).

Le coffre stocke les secrets partagés utilisés par IPsec : **PSK** (clé pré-partagée), **EAP** et **XAuth**.

---

## Deux règles à connaître

1. **Les valeurs sont chiffrées au repos** (AES-256-GCM), avec une clé dérivée de la variable `SECRETS_KEY`.
2. **Une valeur n'est jamais réaffichée.** Ni dans l'interface (`••••••••`), ni par l'API. Même un administrateur ne peut pas relire un PSK enregistré. Si vous l'avez perdu, créez-en un nouveau.

---

## Créer un secret PSK

1. **Configuration → Secrets** → **+ Secret**.
2. Renseignez :
   - **Nom** : l'identifiant que vous utiliserez dans le tunnel, ex. `psk-dakar` ;
   - **Type** : `PSK` ;
   - **Valeur** : la clé partagée (la même que celle configurée en face) ;
   - **Utilisé par** : un texte libre pour vous y retrouver, ex. `paris-dakar`.
3. **Enregistrer**.

Le secret apparaît dans la liste, valeur masquée.

---

## Rattacher le secret à un tunnel

1. Ouvrez l'**Éditeur de tunnel** (nouveau tunnel ou **Éditer** un existant).
2. **Authentification** → `PSK`.
3. Un menu **Secret PSK** apparaît : choisissez `psk-dakar`.
4. **Valider & appliquer**.

Au moment d'appliquer, la console :

- charge la **connexion** sur la passerelle (`load-conn`) ;
- déchiffre le secret et le charge sur la passerelle (`load-shared`), en l'associant aux identités IKE (les adresses des deux extrémités).

Le PSK ne transite donc que du serveur vers le démon `charon`, jamais vers votre navigateur.

> Si le tunnel est un **site-à-site géré des deux côtés**, le même PSK est chargé **sur les deux passerelles** automatiquement. Voir [Site-à-site des deux côtés](08-site-a-site-deux-cotes.md).

---

## Supprimer un secret

**Secrets** → **✕** sur la ligne → confirmer.

> Attention : si un tunnel référence encore ce secret, il ne pourra plus s'authentifier à la prochaine négociation. Vérifiez la colonne **Utilisé par** avant de supprimer.

---

## EAP et XAuth

Les types `EAP` et `XAuth` se créent de la même manière. Ils servent aux accès nomades (road warriors), en complément des **Utilisateurs VPN** — voir [Modules de configuration](11-modules-configuration.md).

---

## Que se passe-t-il si je change `SECRETS_KEY` ?

Tous les secrets **déjà enregistrés deviennent illisibles** : ils ont été chiffrés avec l'ancienne clé. L'application ne pourra plus les charger sur les passerelles.

Ne changez cette variable qu'à l'installation, **avant** de créer le moindre secret. Voir [Installation](02-installation.md).

---

## Et ensuite ?

→ [PKI & certificats](07-pki-certificats.md) — l'alternative au PSK, plus robuste.
