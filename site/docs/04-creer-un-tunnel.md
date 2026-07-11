# Créer un tunnel

> Rôle requis : **administrateur** ou **opérateur**.

---

## Le principe

Vous remplissez un formulaire, l'application :

1. **valide** votre saisie (une configuration dangereuse est refusée, pas appliquée) ;
2. calcule un **score de sécurité** ;
3. **traduit** la configuration en instruction VICI et la **charge à chaud** dans le démon `charon` ;
4. enregistre une **version** et une **entrée d'audit**.

À aucun moment un fichier de configuration n'est écrit à la main.

---

## Pas à pas

### 1. Ouvrir l'éditeur

Cliquez sur **+ Nouveau tunnel** (en haut à droite), ou allez dans **Configuration → Éditeur de tunnel**.

### 2. Remplir les champs

| Champ | Ce qu'on y met | Exemple |
|---|---|---|
| **Nom** | Un nom unique **par passerelle** | `paris-dakar` |
| **Passerelle** | La passerelle StrongSwan qui portera le tunnel | `gw-local` |
| **Passerelle pair** | *(optionnel)* l'autre extrémité, si elle est **aussi gérée** par la console | — |
| **Type** | `Site-à-site`, `Host-à-host` ou `Road warrior` | Site-à-site |
| **Version IKE** | `IKEv2` (recommandé) ou `IKEv1` | IKEv2 |
| **Extrémité locale** | L'adresse IP publique de **votre** passerelle | `203.0.113.10` |
| **Réseau local** | Le(s) sous-réseau(x) protégé(s) de votre côté | `10.1.0.0/16` |
| **Extrémité distante** | L'adresse IP du pair | `198.51.100.20` |
| **Réseau distant** | Le(s) sous-réseau(x) protégé(s) du pair | `10.2.0.0/16` |
| **Authentification** | `PSK`, `Certificat` ou `EAP` | PSK |
| **Propositions IKE** | La suite cryptographique de la phase 1 | `aes256-sha256-modp2048` |
| **Propositions ESP** | La suite cryptographique du trafic | `aes256gcm16` |
| **Perfect Forward Secrecy** | Laissez **coché** | ✔ |

Plusieurs valeurs se séparent par des virgules (réseaux, propositions).

### 3. Regarder le score

À droite, l'**anneau de score** se recalcule **en direct** pendant que vous saisissez. Les constats sont listés en dessous : *« Pas de préparation post-quantique »*, *« Chiffrement 3DES/DES faible »*, etc.

C'est votre garde-fou : vous voyez la qualité de la configuration **avant** de l'appliquer.

### 4. Valider & appliquer

Cliquez sur **Valider & appliquer**.

- ✅ **Succès** : un message vous donne le score et la version (`Créé · score 94 · v1`), et vous êtes ramené à la liste des Connexions.
- ❌ **Erreur** : voir ci-dessous.

---

## Les trois types de tunnel

### Site-à-site
Deux réseaux reliés à travers Internet. Les deux extrémités ont une adresse fixe. C'est le cas le plus courant.
Si **les deux passerelles sont gérées** par la console, voir [Site-à-site des deux côtés](08-site-a-site-deux-cotes.md).

### Host-à-host
Deux machines qui se parlent directement (pas de réseau derrière). Mettez l'adresse de l'hôte comme réseau protégé (ex. `192.0.2.44/32`).

### Road warrior
Des postes nomades qui se connectent depuis n'importe où. L'extrémité distante est **dynamique** : laissez « Extrémité distante » vide, l'authentification se fait typiquement en **EAP** ou par **certificat**. Les utilisateurs se gèrent dans **Utilisateurs VPN** (voir [Modules de configuration](11-modules-configuration.md)).

---

## Quand la validation refuse (erreur 422)

L'application **bloque** une configuration invalide ou manifestement dangereuse. Le message vous dit précisément quoi corriger.

**Essayez pour voir** : mettez `aes256-sha256-modp1024` comme proposition IKE, puis appliquez. Vous obtenez :

> `Proposition cryptographique faible détectée — modp1024 déconseillé (DH groupe 2)`

Autres refus fréquents :

| Message | Cause | Correction |
|---|---|---|
| `nom requis` | Le champ Nom est vide | Nommez le tunnel |
| `ike_version doit valoir 1 ou 2` | Version IKE invalide | Choisissez IKEv2 |
| `au moins une proposition IKE est requise` | Champ propositions vide | Saisissez une suite, ex. `aes256-sha256-ecp384` |
| `adresse locale requise` | Extrémité locale vide | Renseignez l'IP de la passerelle |
| `CIDR/adresse invalide: …` | Un réseau est mal écrit | Utilisez la notation CIDR (`10.1.0.0/16`) |
| `un tunnel de ce nom existe déjà sur cette passerelle` | Doublon | Changez le nom |

**Important** : quand la validation échoue, **rien n'a été appliqué** sur la passerelle. Vous pouvez corriger et réessayer sans risque.

---

## Suites cryptographiques recommandées

| Usage | Proposition IKE | Proposition ESP |
|---|---|---|
| Moderne (recommandé) | `aes256gcm16-sha384-ecp384-mlkem768` | `aes256gcm16` |
| Classique, très compatible | `aes256-sha256-modp2048` | `aes256gcm16` |
| **À éviter** | `3des-md5-modp1024` | `3des-md5` |

`mlkem768` (post-quantique) exige **StrongSwan 6.0 ou plus** sur la passerelle. Voir [Score de sécurité](09-score-de-securite.md).

---

## Et ensuite ?

Le tunnel est créé mais n'est pas forcément établi. → [Piloter un tunnel](05-piloter-un-tunnel.md)
