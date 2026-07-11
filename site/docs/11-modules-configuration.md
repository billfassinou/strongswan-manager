# Modules de configuration

> Rôle requis : **administrateur** ou **opérateur** pour modifier. Tout le monde peut consulter.

Ces modules couvrent les fonctionnalités de StrongSwan au-delà des tunnels eux-mêmes. Ils fonctionnent **tous de la même manière** : une liste, un bouton d'ajout, une modale d'édition, une suppression. Les entrées sont **persistées en base** et **auditées**.

---

## Pools & IP virtuelles

**StrongSwan → Pools & IP virtuelles**

Les plages d'adresses distribuées aux clients nomades, et les attributs poussés avec.

| Champ | Exemple |
|---|---|
| **Nom** | `pool-roadwarrior` |
| **Plage** | `10.9.0.0/24` |
| **Source** | `Interne`, `SQL`, `DHCP` ou `RADIUS` |
| **DNS poussé** | `10.1.0.53` |
| **Split-tunnel** | `10.0.0.0/8` (les réseaux à faire passer dans le tunnel) |

---

## RADIUS / AAA

**StrongSwan → RADIUS / AAA**

Les serveurs RADIUS vers lesquels déléguer l'authentification EAP et la comptabilité.

| Champ | Exemple |
|---|---|
| **Nom** | `radius-primaire` |
| **Adresse:port** | `10.1.0.20:1812` |
| **Rôle** | `Primaire` ou `Secondaire` |
| **Accounting** | activé / désactivé |

---

## Politiques & routage

**StrongSwan → Politiques & routage**

| Type | À quoi ça sert |
|---|---|
| **shunt** | Exclure du tunnel un trafic donné (`pass`) ou le bloquer (`drop`). Typique : ne pas chiffrer le LAN d'administration. |
| **trap** | Monter le tunnel **au premier paquet** (à la demande). |
| **route-based** | Router le trafic dans une interface dédiée (XFRM/VTI) au lieu de politiques par sous-réseau. |

| Champ | Exemple |
|---|---|
| **Nom** | `bypass-management` |
| **Type** | `shunt` |
| **Détail** | `pass 10.1.0.0/24` |
| **Interface** | `xfrm1` (pour du route-based) |

---

## Autorités & enrôlement

**StrongSwan → Autorités & enrôlement**

Déclarer des autorités de certification **externes** et leurs points de vérification.

| Champ | Exemple |
|---|---|
| **Nom** | `Partner Sub-CA` |
| **URI CRL** | `http://crl.partner.io/sub.crl` |
| **URI OCSP** | `http://ocsp.partner.io` |
| **Enrôlement** | `—`, `SCEP` ou `EST` |

> Ces entrées sont **enregistrées** mais l'enrôlement automatique SCEP/EST n'est **pas encore exécuté** par le serveur. L'autorité **interne** (celle qui émet vos certificats) se gère dans [PKI & certificats](07-pki-certificats.md).

---

## Utilisateurs VPN

**Configuration → Utilisateurs VPN**

Les accès nomades (road warriors).

| Champ | Exemple |
|---|---|
| **Identité** | `marc.diallo` |
| **Méthode** | `EAP-TLS`, `Certificat` ou `EAP-MSCHAPv2` |
| **Quota** | `50 Go` |
| **Plage horaire** | `08:00–20:00` |
| **Actif** | oui / non |

Désactiver un utilisateur est immédiat. Pour lui couper l'accès de façon certaine quand il utilise un **certificat**, révoquez-le également dans la [PKI](07-pki-certificats.md).

---

## Monitoring & alertes

**Supervision → Monitoring & alertes**

Les règles d'alerte.

| Champ | Exemple |
|---|---|
| **Nom** | `Chute de tunnel` |
| **Condition** | `état = down` |
| **Canaux** | `Email, Slack` |
| **Activée** | oui / non |

> Les règles sont **enregistrées**, mais l'**envoi effectif** des notifications (email, Slack, Telegram, webhook) n'est pas encore implémenté. En attendant, les métriques sont exposées au format **Prometheus** sur `/metrics` : vous pouvez y brancher l'alerting de votre choix (Alertmanager, Grafana). Voir [Superviser](12-superviser.md).

---

## Paramètres du démon

**StrongSwan → Paramètres du démon**

Les réglages globaux de `charon` (l'équivalent de `strongswan.conf`).

| Réglage | Défaut | Effet |
|---|---|---|
| **Threads worker** | 16 | Nombre de fils de traitement |
| **Retransmissions IKE** | 5 | Nombre de tentatives avant abandon |
| **Timeout retransmission** | 4 s | Délai entre deux tentatives |
| **DNS attribué** | `10.1.0.53` | DNS poussé aux clients |
| **install_routes** | activé | Le démon installe-t-il les routes ? |
| **Fragmentation IKEv2** | activé | Fragmentation des messages IKE (utile derrière un MTU réduit) |
| **Niveau de log IKE** | 1 (audit) | Verbosité du sous-système IKE |

Cliquez sur **Valider & recharger** pour enregistrer.

> Ces paramètres sont **persistés** et rejouables ; la propagation vers un `strongswan.conf` réel sur la passerelle n'est pas encore automatisée.

---

## Sous le capot (pour les curieux)

Tous ces modules partagent **un seul mécanisme** côté serveur : une table générique où chaque entrée porte un **type** (`kind`) et ses champs en JSON, exposée sur `/api/v1/config/{kind}`.

Les `kind` disponibles : `pool`, `radius`, `policy`, `authority`, `vpnuser`, `alert`, `daemon`.

C'est ce qui permet d'ajouter un module sans nouvelle table ni nouveau code serveur. Voir [API](A3-api.md).

---

## Et ensuite ?

→ [Superviser](12-superviser.md)
