# Présentation

**StrongSwan Manager** est une console web d'administration pour **StrongSwan**, l'implémentation open source du VPN IPsec.

Il remplace ce que l'on fait aujourd'hui à la main :

| Sans l'outil | Avec l'outil |
|---|---|
| Éditer `swanctl.conf` dans un terminal, sur chaque passerelle | Un formulaire guidé, validé avant application |
| `swanctl --load-all`, `--list-sas`, `--initiate`… en ligne de commande | Des boutons : **Monter**, **Couper**, **Renégocier** |
| Aucune idée de l'état global du parc | Tableau de bord temps réel, topologie, alertes |
| Un tunnel en IKEv1/3DES qui survit des années | Un **score de sécurité** qui le signale, et un bouton « Corriger » |
| Aucune trace de qui a changé quoi | Un **journal d'audit inaltérable** |

> Point clé : la configuration est appliquée **via l'API VICI** de StrongSwan (l'interface de contrôle officielle), jamais en écrivant des fichiers de configuration à la main.

---

## À qui ça s'adresse

### L'administrateur réseau
Il crée et exploite les tunnels : interconnexions entre sites, accès nomades, liaisons vers le cloud. Il a besoin de rapidité et de sûreté (ne pas casser un tunnel en production).
→ Commencez par [Créer un tunnel](04-creer-un-tunnel.md) puis [Piloter un tunnel](05-piloter-un-tunnel.md).

### L'opérateur (exploitation, NOC)
Il surveille, monte et coupe des tunnels, diagnostique les incidents. Il a les mêmes droits d'écriture que l'administrateur.
→ Voir [Superviser](12-superviser.md) et [Dépannage](15-depannage.md).

### L'auditeur / le RSSI
Il ne modifie rien, mais doit **prouver** : qui a changé quoi, quel est le niveau cryptographique du parc, quels certificats expirent.
→ Voir [Score de sécurité](09-score-de-securite.md) et [Auditer](13-auditer.md).

### Le profil « lecture seule »
Supervision uniquement : il consulte, il ne peut rien modifier (l'interface masque les actions, l'API les refuse).

### L'équipe DevSecOps
Elle veut automatiser : créer des tunnels depuis un pipeline, exporter des métriques, intégrer dans son outillage.
→ Voir [API REST & WebSocket](A3-api.md).

Le détail des droits est dans [Rôles & permissions](A1-roles-et-permissions.md).

---

## Ce que fait l'outil, concrètement

- **Tunnels IPsec** : site-à-site, host-à-host, road warrior. IKEv1 et IKEv2. Authentification par **PSK**, **certificat** ou **EAP**.
- **Application à chaud** : la configuration est chargée dans le démon `charon` sans coupure, avec validation préalable.
- **Pilotage** : monter (`initiate`), couper (`terminate`), renégocier (`rekey`) une connexion depuis l'interface.
- **PKI intégrée** : autorité de certification interne, émission de certificats X.509, **révocation** et **CRL**.
- **Coffre de secrets** : PSK/EAP/XAuth chiffrés au repos ; les valeurs ne sont jamais réaffichées.
- **Score de sécurité** : chaque tunnel est noté ; les algorithmes faibles sont signalés.
- **Versions & rollback** : chaque changement crée une version ; on peut revenir en arrière.
- **Supervision temps réel** : l'état des tunnels est poussé en direct (WebSocket) ; métriques Prometheus.
- **Audit inaltérable** : journal *append-only*, chaîné par hachage.
- **Modules StrongSwan** : pools d'IP virtuelles, RADIUS/AAA, politiques & routage, autorités, utilisateurs VPN, règles d'alerte, paramètres du démon.

---

## Ce qui n'est pas (encore) là

Pour rester honnête, voici ce qui est **prévu mais pas implémenté** à ce stade :

- Répondeur **OCSP** et enrôlement **SCEP/EST** (la révocation fonctionne via CRL).
- **Agent distant mTLS** (aujourd'hui le serveur accède directement au socket VICI de la passerelle).
- **Multi-tenant / SSO** SAML-OIDC (édition Enterprise).
- Remplacement du chiffrement applicatif des secrets par **Vault**.

---

## Et ensuite ?

1. [Installation](02-installation.md) — l'application tourne en trois commandes.
2. [Découvrir l'interface](03-decouvrir-linterface.md) — la visite guidée.
3. [Créer un tunnel](04-creer-un-tunnel.md) — votre premier tunnel.
