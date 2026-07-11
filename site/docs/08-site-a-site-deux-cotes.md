# Site-à-site géré des deux côtés

> Rôle requis : **administrateur** ou **opérateur**.

Quand **les deux passerelles** d'un tunnel site-à-site sont gérées par la console (cas typique : vos propres sites, ou un MSP qui administre les deux extrémités), vous pouvez configurer **les deux bouts en une seule opération**.

---

## Ce que fait la console

Vous renseignez une **Passerelle pair**. À l'application, la console :

1. charge la connexion sur **votre** passerelle ;
2. charge la **connexion miroir** sur la passerelle **pair** — c'est-à-dire la même connexion avec les **extrémités et les réseaux inversés** ;
3. charge le **secret PSK sur les deux**, ou **le certificat de chacun sur sa propre passerelle**.

Résultat : les deux démons sont configurés de façon cohérente. Il ne reste plus qu'à établir la SA.

---

## Pas à pas — avec un PSK

1. **Secrets** → créez un PSK, ex. `psk-hub` (voir [Gérer les secrets](06-secrets.md)).
2. **Éditeur de tunnel** :
   - **Nom** : `hub`
   - **Passerelle** : `gw-a`
   - **Passerelle pair** : `gw-b` ← *c'est la clé*
   - **Extrémité locale** : l'IP de `gw-a`
   - **Réseau local** : le réseau derrière `gw-a`, ex. `192.168.10.0/24`
   - **Extrémité distante** : l'IP de `gw-b`
   - **Réseau distant** : le réseau derrière `gw-b`, ex. `192.168.20.0/24`
   - **Authentification** : `PSK` → **Secret PSK** : `psk-hub`
   - **Propositions** : `aes256-sha256-modp2048` / `aes256gcm16`
3. **Valider & appliquer**.
4. **Connexions** → **Monter**.

**Ce que vous devez voir** : le tunnel passe en `negotiating` puis **`up`** en quelques secondes.

---

## Pas à pas — avec des certificats

1. **PKI & Certificats** → émettez **deux** certificats :
   - `cert-a`, CN `gw-a`, **SAN = l'IP de gw-a**
   - `cert-b`, CN `gw-b`, **SAN = l'IP de gw-b**
2. **Éditeur de tunnel** : mêmes champs que ci-dessus, mais
   - **Authentification** : `Certificat`
   - **Certificat local** : `cert-a`
   - **Certificat du pair** : `cert-b`
3. **Valider & appliquer**, puis **Monter**.

La console charge l'autorité + `cert-a` + sa clé sur `gw-a`, et l'autorité + `cert-b` + sa clé sur `gw-b`.

> ⚠️ Les **SAN doivent correspondre aux adresses des extrémités** du tunnel. C'est l'erreur numéro un.

---

## Vérifier côté démon

Si vous utilisez le lab (voir [Connecter de vraies passerelles](14-connecter-passerelles-reelles.md)) :

```bash
docker compose exec strongswan-a swanctl --list-sas --uri unix:///vicirun/charon.vici
```

Vous devez lire quelque chose comme :

```
hub: #1, ESTABLISHED, IKEv2
  hub-net: #1, INSTALLED, TUNNEL, ESP:AES_GCM_16-256
```

`ESTABLISHED` = la SA IKE est montée. `INSTALLED` = la SA enfant (le chiffrement du trafic) est installée dans le noyau. Le tunnel fonctionne réellement.

---

## Supprimer

Supprimer le tunnel **décharge les deux connexions** (la vôtre et le miroir). Rien à nettoyer à la main.

---

## Et si le pair n'est pas géré par la console ?

Laissez **Passerelle pair** vide. Vous configurez seulement **votre** côté ; l'autre extrémité (un FortiGate, un Cisco, un StrongSwan non géré…) doit être configurée par ailleurs, **avec des paramètres symétriques** : mêmes propositions, même PSK ou certificats de confiance, sous-réseaux inversés.

---

## Et ensuite ?

→ [Score de sécurité](09-score-de-securite.md)
