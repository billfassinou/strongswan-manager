<div align="center">

# StrongSwan Manager

**La console web d'administration pour StrongSwan.**
Créez, pilotez et supervisez vos tunnels IPsec depuis une interface — plus jamais depuis un
fichier de configuration.

[![Licence](https://img.shields.io/badge/licence-AGPL--3.0-blue.svg)](LICENSE)
[![Site](https://img.shields.io/badge/site-en%20ligne-brightgreen.svg)](https://billfassinou.github.io/strongswan-manager/)
[![Documentation](https://img.shields.io/badge/documentation-FR%20%7C%20EN-blueviolet.svg)](https://billfassinou.github.io/strongswan-manager/docs/)

**[🌐 Site](https://billfassinou.github.io/strongswan-manager/)** ·
**[📘 Documentation](https://billfassinou.github.io/strongswan-manager/docs/)** ·
**[⬇️ Télécharger](https://github.com/billfassinou/strongswan-manager/releases/latest)** ·
[English](#english)

</div>

---

## Installer

Sur un serveur Debian/Ubuntu ou RHEL/AlmaLinux/Rocky (systemd, amd64 ou arm64) :

```bash
curl -fsSL https://raw.githubusercontent.com/billfassinou/strongswan-manager/main/deploy/install.sh | sudo bash
```

Le script récapitule ce qu'il va modifier avant d'agir. Il installe PostgreSQL et strongSwan
si besoin, génère ses secrets, ouvre le socket VICI au service (qui ne tourne donc **pas** en
root), pose l'unité systemd, et vérifie que la console répond avant de rendre la main. Il vous
donne alors l'URL et le mot de passe — que la console vous fera changer à la connexion.

Ensuite : `swanmgrctl doctor`, `backup`, `restore`, `upgrade` (avec retour arrière automatique).

**Autres chemins** — paquets `.deb` / `.rpm` (mises à jour par `apt`/`dnf`), image
`ghcr.io/billfassinou/strongswan-manager`, ou **installation hors ligne** : les archives Linux
des [releases](https://github.com/billfassinou/strongswan-manager/releases/latest) sont des
bundles autonomes, `sudo ./install.sh --skip-deps` n'accède à aucun réseau. Voir la
**[documentation d'installation](https://billfassinou.github.io/strongswan-manager/docs/#/02-installation)**.

## Essayer en trois commandes

```bash
git clone https://github.com/billfassinou/strongswan-manager.git
cd strongswan-manager/backend
make run
# → https://localhost:7926   —   connexion : admin / admin1234
#   (certificat auto-signé : le navigateur avertit au 1er accès — c'est normal)
```

Docker suffit : la base, les migrations, la PKI et les comptes de démonstration sont créés au
premier démarrage. Sans passerelle réelle configurée, l'application démarre en **mode démo**
(passerelle simulée) : toute l'interface est explorable immédiatement, mais **aucun trafic
n'est réellement chiffré**.

Pour de **vrais tunnels**, entre deux démons strongSwan dockerisés :

```bash
make lab-up      # + 2 passerelles strongSwan réelles
```

## Télécharger

Chaque [release](https://github.com/billfassinou/strongswan-manager/releases/latest) publie un
**bundle autonome** — le binaire contient l'API **et** l'interface web, et l'archive y ajoute
l'installeur et `swanmgrctl` — pour linux/amd64, linux/arm64, darwin/amd64 et darwin/arm64,
plus des paquets `.deb`/`.rpm`, un SBOM et une signature `cosign`.

```bash
tar xzf strongswan-manager_v0.1.1_linux_amd64.tar.gz
cd strongswan-manager_v0.1.1_linux_amd64
sha256sum -c ../SHA256SUMS          # vérifiez l'archive
sudo ./install.sh                   # …ou lancez le binaire vous-même :
DATABASE_URL='postgres://…' JWT_SECRET="$(openssl rand -hex 32)" \
  SECRETS_KEY="$(openssl rand -hex 32)" ./strongswan-manager
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

## Licence & modèle open-core

Ce dépôt contient l'**édition Community**, sous **[AGPL-3.0](LICENSE)**. Elle est complète et
suffit à administrer StrongSwan au quotidien : rien n'y est bridé, ni limité dans le temps, ni
plafonné en nombre de tunnels.

Concrètement, l'AGPL vous garantit le droit d'utiliser, de modifier et de redistribuer ce code.
En contrepartie, **si vous exploitez une version modifiée comme service en réseau, vous devez en
publier les sources**. C'est ce qui empêche que le cœur soit refermé par un tiers.

Les modules **Premium** et **Enterprise** (conformité, alerting avancé, IA, multi-tenant, SSO)
sont distribués **séparément, sous licence commerciale** — ils ne sont pas dans ce dépôt.

**Contribuer** : lisez [CONTRIBUTING.md](CONTRIBUTING.md). La signature du [CLA](CLA.md) est
requise (`git commit -s`) — c'est ce qui permet au projet de maintenir ce double modèle.

---

<a name="english"></a>

## English

**StrongSwan Manager** is a web admin console for the StrongSwan IPsec VPN: guided tunnel
creation, built-in PKI, real-time monitoring, an automatic security score and a tamper-proof
audit log — all driven through StrongSwan's official **VICI** API, never by writing
configuration files.

Install it on a systemd server (Debian/Ubuntu, RHEL/AlmaLinux/Rocky):

```bash
curl -fsSL https://raw.githubusercontent.com/billfassinou/strongswan-manager/main/deploy/install.sh | sudo bash
```

Or just try it, with Docker:

```bash
git clone https://github.com/billfassinou/strongswan-manager.git
cd strongswan-manager/backend && make run
# → https://localhost:7926   —   sign in: admin / admin1234
#   (self-signed certificate: the browser warns on first access — this is expected)
```

Every [release](https://github.com/billfassinou/strongswan-manager/releases/latest) ships
**self-contained bundles** (binary + installer + `swanmgrctl`), `.deb`/`.rpm` packages and a
container image. The Linux bundles install **with no network access at all**
(`sudo ./install.sh --skip-deps`) — air-gapped sites are a first-class path, not an
afterthought.

→ **[Website](https://billfassinou.github.io/strongswan-manager/en/)** ·
**[Documentation](https://billfassinou.github.io/strongswan-manager/docs/en/)**
