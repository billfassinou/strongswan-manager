# Piloter un tunnel

> Rôle requis : **administrateur** ou **opérateur**.

Allez dans **Supervision → Connexions**. Chaque ligne porte quatre actions.

---

## Monter un tunnel

Cliquez sur **Monter**.

La console demande à `charon` d'établir la connexion (commande VICI `initiate`). Le tunnel passe alors :

`installing` → `negotiating` → **`up`**

Vous n'avez **rien à rafraîchir** : le serveur sonde l'état des SA en continu et pousse le changement dans votre navigateur (WebSocket). La pastille change toute seule, généralement en quelques secondes.

**Si le tunnel reste `down`**, voir [Dépannage](15-depannage.md) — c'est presque toujours le pair qui ne répond pas, ou une configuration asymétrique.

---

## Couper un tunnel

Cliquez sur **Couper** (`terminate`). La SA est détruite, le tunnel repasse `down`.

La **connexion reste configurée** sur la passerelle : vous pourrez la remonter d'un clic. Couper ≠ supprimer.

---

## Renégocier

L'API expose une action **`rekey`** qui force la renégociation des clés sans couper le trafic. Elle est disponible via l'API (`POST /api/v1/tunnels/{id}/rekey`) — voir [API](A3-api.md).

---

## Éditer un tunnel

Cliquez sur **Éditer** : l'éditeur s'ouvre **pré-rempli**.

Modifiez, puis **Valider & appliquer** :

- la nouvelle configuration est **validée** puis **rechargée à chaud** sur la passerelle ;
- une **nouvelle version** est créée (v2, v3…) ;
- le score est recalculé.

Si vous vous êtes trompé, vous pouvez revenir en arrière : [Versions & rollback](10-versions-et-rollback.md).

---

## Supprimer un tunnel

Cliquez sur **✕**, puis confirmez.

L'application **décharge** la connexion de la passerelle (`unload-conn`) **avant** de supprimer l'enregistrement. Si la passerelle pair était gérée elle aussi, la connexion miroir est également déchargée.

> ⚠️ La suppression est **définitive** : le tunnel disparaît de la base, ses versions avec lui. Pour interrompre temporairement un tunnel, utilisez **Couper**.

---

## Comprendre les états

| État | Ce qui se passe réellement | Que faire |
|---|---|---|
| `installing` | La configuration vient d'être chargée dans `charon`, aucune SA n'est encore montée | Cliquez sur **Monter** |
| `negotiating` | Un échange IKE est en cours (établissement ou renégociation) | Patientez quelques secondes |
| `up` | La SA IKE **et** la SA enfant (ESP) sont établies : le trafic passe | Rien |
| `down` | Aucune SA. Soit vous n'avez pas monté le tunnel, soit la négociation échoue | **Monter**, sinon [Dépannage](15-depannage.md) |
| `unknown` | La passerelle n'a pas répondu à l'interrogation VICI | Vérifiez la passerelle |

L'état affiché **provient du démon** : il est lu périodiquement via VICI (`list-sas`), il ne s'agit pas d'un état déclaratif stocké en base.

---

## Et ensuite ?

- Le tunnel utilise un PSK ? → [Gérer les secrets](06-secrets.md)
- Vous voulez de l'authentification par certificat ? → [PKI & certificats](07-pki-certificats.md)
- Les deux passerelles sont chez vous ? → [Site-à-site des deux côtés](08-site-a-site-deux-cotes.md)
