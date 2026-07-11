<div align="center">

# StrongSwan Manager

**La console web d'administration pour StrongSwan.**
Créez, pilotez et supervisez vos tunnels IPsec depuis une interface — plus jamais depuis un
fichier de configuration.

[![Licence](https://img.shields.io/badge/licence-Apache--2.0-blue.svg)](LICENSE)
[![Site](https://img.shields.io/badge/site-en%20ligne-brightgreen.svg)](https://billfassinou.github.io/strongswan-manager/)
[![Documentation](https://img.shields.io/badge/documentation-FR%20%7C%20EN-blueviolet.svg)](https://billfassinou.github.io/strongswan-manager/docs/)

**[🌐 Site](https://billfassinou.github.io/strongswan-manager/)** ·
**[📘 Documentation](https://billfassinou.github.io/strongswan-manager/docs/)** ·
**[⬇️ Télécharger](https://github.com/billfassinou/strongswan-manager/releases/latest)** ·
[English](#english)

</div>

---

## Démarrer en trois commandes

```bash
git clone https://github.com/billfassinou/strongswan-manager.git
cd strongswan-manager/backend
make run
# → http://localhost:8080   —   connexion : admin / admin1234
```

Docker suffit : la base, les migrations, la PKI et les comptes de démonstration sont créés au
premier démarrage. Sans passerelle réelle configurée, l'application démarre en **mode démo**
(passerelle simulée) : toute l'interface est explorable immédiatement.

Pour de **vrais tunnels**, entre deux démons strongSwan dockerisés :

```bash
make lab-up      # + 2 passerelles strongSwan réelles
```

## Télécharger

Chaque [release](https://github.com/billfassinou/strongswan-manager/releases/latest) publie un
**binaire autonome** — il contient l'API **et** l'interface web — pour linux/amd64,
linux/arm64, darwin/amd64 et darwin/arm64.

```bash
tar xzf strongswan-manager_v0.1.0_linux_amd64.tar.gz
cd strongswan-manager_v0.1.0_linux_amd64
sha256sum -c ../SHA256SUMS          # vérifiez l'archive
DATABASE_URL='postgres://…' JWT_SECRET="$(openssl rand -hex 32)" ./strongswan-manager
```

Seul prérequis : un **PostgreSQL** joignable. Voir les
[variables d'environnement](https://billfassinou.github.io/strongswan-manager/docs/#/A2-configuration).

## Ce que fait le produit

- **Tunnels en quelques clics** — site-à-site, host-à-host, road warrior. Formulaire guidé,
  validation avant application, chargement à chaud **via l'API VICI** (aucun fichier écrit).
- **Score de sécurité automatique** — IKEv1, 3DES, MD5, modp1024, PFS désactivé, absence de
  ML-KEM : tout est noté et signalé, avec un bouton « Corriger ».
- **PKI intégrée** — CA interne, émission de certificats X.509, authentification par
  certificat, révocation et **CRL** publiée sur un point de distribution.
- **Coffre de secrets** — PSK/EAP/XAuth chiffrés en AES-256-GCM, jamais réaffichés.
- **Supervision temps réel** — état des SA poussé en WebSocket, métriques Prometheus.
- **Audit inaltérable** — journal *append-only* chaîné par hachage, garanti en base.
- **Versions & rollback**, **rôles & permissions** (4 rôles, l'API refuse ce que l'interface
  masque), pools d'IP, RADIUS, politiques, autorités, utilisateurs VPN, règles d'alerte,
  paramètres du démon.

## Le dépôt

| Dossier | Contenu |
|---|---|
| [`backend/`](backend/) | L'application : API Go (chi, PostgreSQL, VICI, JWT/RBAC) **et** le front React embarqué dans le binaire. Voir [`backend/README.md`](backend/README.md). |
| [`site/`](site/) | Le site publié : vitrine (FR/EN) **et** documentation (`site/docs/`). Déployé sur GitHub Pages. Voir [`site/README.md`](site/README.md). |

## Architecture

```
Front React (embarqué)  →  API Go (REST + WebSocket, JWT/RBAC, audit)  →  VICI  →  charon
                                        ↓
                                   PostgreSQL
```

Le front est compilé dans le binaire Go et servi **à la même origine** que l'API. Le serveur
pilote chaque passerelle par **VICI**, l'interface de contrôle officielle de StrongSwan — jamais
en écrivant des fichiers de configuration.

## Développement

```bash
cd backend
make web                # construit le front (requis avant un build Go local)
make build              # compile
make test               # tests unitaires
make cover              # + couverture
make test-integration   # tests d'intégration (PostgreSQL jetable)
make lab-up / lab-down  # lab avec 2 vraies passerelles strongSwan
```

Les **tests unitaires** vivent à côté du code qu'ils testent (convention Go) ; les **tests
d'intégration** sont dans [`backend/test/`](backend/test/).

## Licence

[Apache-2.0](LICENSE).

---

<a name="english"></a>

## English

**StrongSwan Manager** is a web admin console for the StrongSwan IPsec VPN: guided tunnel
creation, built-in PKI, real-time monitoring, an automatic security score and a tamper-proof
audit log — all driven through StrongSwan's official **VICI** API, never by writing
configuration files.

```bash
git clone https://github.com/billfassinou/strongswan-manager.git
cd strongswan-manager/backend && make run
# → http://localhost:8080   —   sign in: admin / admin1234
```

Prebuilt standalone binaries (API + web UI in one file) are attached to every
[release](https://github.com/billfassinou/strongswan-manager/releases/latest).

→ **[Website](https://billfassinou.github.io/strongswan-manager/en/)** ·
**[Documentation](https://billfassinou.github.io/strongswan-manager/docs/en/)**
