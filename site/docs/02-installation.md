# Installation

Quatre chemins, selon ce que vous faites. Le premier est celui d'une mise en service réelle ;
le dernier sert à découvrir le produit en cinq minutes.

| Vous voulez… | Allez à |
|---|---|
| **Installer un serveur** (systemd, sans Docker) | [Installation en une commande](#installation-en-une-commande) |
| **Compiler depuis le dépôt** | [Installation depuis les sources](#installation-depuis-les-sources) |
| Passer par **apt / dnf** (mises à jour intégrées) | [Paquets .deb et .rpm](#paquets-deb-et-rpm) |
| Installer sur une machine **sans accès Internet** | [Installation hors ligne (air-gap)](#installation-hors-ligne-air-gap) |
| Déployer avec **Docker** | [Docker](#docker) |
| **Essayer** le produit, sans rien installer durablement | [Essai rapide](#essai-rapide) |

---

## Installation en une commande

Sur une machine **Debian/Ubuntu** ou **RHEL/AlmaLinux/Rocky** (amd64 ou arm64), avec systemd :

```bash
curl -fsSL https://raw.githubusercontent.com/billfassinou/strongswan-manager/main/deploy/install.sh | sudo bash
```

Le script vous montre d'abord **ce qu'il va modifier**, puis demande confirmation. Il :

1. installe **PostgreSQL** et **strongSwan** s'ils sont absents ;
2. crée la base `swan` et un utilisateur système `swanmgr` (sans shell) ;
3. génère `/etc/strongswan-manager/strongswan-manager.env` avec des **secrets aléatoires** ;
4. ouvre le socket VICI au groupe `swanmgr` — c'est ce qui permet à la console de piloter
   charon **sans tourner en root** ;
5. pose le service systemd, ouvre les ports 7926/7927 dans firewalld ou ufw, et **vérifie
   que la console répond** avant de vous rendre la main.

À la fin, il affiche l'URL de la console et le mot de passe du compte `admin`.

> Le script télécharge l'archive de la release, **vérifie son empreinte SHA-256** (et sa
> signature `cosign` si l'outil est présent) et refuse de continuer si elle ne correspond pas.

Options utiles :

| Option | Effet |
|---|---|
| `--no-strongswan` | N'installe pas strongSwan. Pour une console qui ne pilote que des passerelles **distantes**. |
| `--skip-deps` | N'installe aucun paquet. Voir [hors ligne](#installation-hors-ligne-air-gap). |
| `--version vX.Y.Z` | Installe une version précise. |
| `--yes` | Ne pose aucune question. |

---

## Installation depuis les sources

Vous avez cloné le dépôt et vous voulez installer **votre** version compilée :

```bash
git clone https://github.com/billfassinou/strongswan-manager.git
cd strongswan-manager
sudo ./deploy/install.sh --from-source
```

L'installeur compile l'interface web (elle est **embarquée** dans le binaire), puis le binaire
Go, puis installe le résultat **exactement comme le ferait le bundle** — un seul chemin de code
en aval, donc le même comportement et les mêmes vérifications.

La compilation exige **Go ≥ 1.23** et **Node ≥ 20**. S'ils manquent ou sont trop anciens — c'est
le cas du Go livré par AlmaLinux 9 —, l'installeur récupère les **chaînes officielles**
(go.dev, nodejs.org) dans un dossier temporaire et compile avec. **Rien n'est installé
durablement sur votre machine** ; le dossier disparaît à la fin.

> Les options `--no-strongswan`, `--skip-deps` et `--yes` s'appliquent aussi à ce mode.

---

## Paquets .deb et .rpm

Chaque release publie des paquets natifs. Leur intérêt : `apt upgrade` / `dnf upgrade`
mettent la console à jour comme n'importe quel autre logiciel, **sans toucher à votre
configuration ni à votre base**.

```bash
# Debian / Ubuntu
sudo apt install ./strongswan-manager_1.0.0_amd64.deb

# RHEL / AlmaLinux / Rocky
sudo dnf install ./strongswan-manager-1.0.0-1.x86_64.rpm
```

Le post-installation fait le même travail que le script : utilisateur système, base, secrets,
service. **Retirer le paquet ne supprime ni la base, ni la configuration** — il faudrait les
effacer explicitement (le message de désinstallation vous donne les commandes).

---

## Installation hors ligne (air-gap)

Les archives `linux` des releases ne sont pas de simples binaires : ce sont des **bundles
autonomes** contenant le binaire, l'installeur, `swanmgrctl` et l'unité systemd. Rien n'est
téléchargé pendant l'installation.

Sur une machine connectée :

```bash
curl -LO https://github.com/billfassinou/strongswan-manager/releases/download/v1.0.0/strongswan-manager_v1.0.0_linux_amd64.tar.gz
curl -LO https://github.com/billfassinou/strongswan-manager/releases/download/v1.0.0/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing
```

Transportez l'archive, puis sur la machine cible :

```bash
tar xzf strongswan-manager_v1.0.0_linux_amd64.tar.gz
cd strongswan-manager_v1.0.0_linux_amd64
sudo ./install.sh --skip-deps
```

`--skip-deps` signifie « n'installe aucun paquet » : **PostgreSQL doit déjà être présent**
(fourni par votre image système ou un miroir local), sinon le script s'arrête en vous le
disant. Sans cette option, l'installeur essaierait d'atteindre les dépôts de la distribution.

---

## Docker

Une image est publiée à chaque release. Le fichier `docker-compose.prod.yml` et le script
`docker-install.sh` sont dans le bundle (et dans `deploy/` du dépôt).

```bash
./docker-install.sh
```

Il génère un `.env` aux secrets aléatoires, démarre PostgreSQL et la console, attend qu'elle
réponde, puis affiche l'URL et le mot de passe.

Pour un site sans Internet, transportez les images :

```bash
docker save ghcr.io/billfassinou/strongswan-manager:v1.0.0 postgres:16-alpine | gzip > images.tar.gz
# sur la machine cible :
docker load < images.tar.gz
```

> ⚠️ Ne confondez pas avec `backend/docker-compose.yml`, qui est le **lab de développement** :
> ses secrets sont publics et ses tunnels sont simulés.

---

## Essai rapide

Pour découvrir l'interface sans rien installer durablement — il vous faut Docker et le dépôt :

```bash
git clone <dépôt> && cd strongswan/backend
make run
# → https://localhost:7926, comptes admin/operator/auditor/viewer, mot de passe admin1234
```

C'est le **mode démo** : les tunnels sont **simulés**, aucun trafic n'est réellement chiffré.
Tout le reste (créer un tunnel, le monter, voir l'état changer en temps réel) fonctionne.
Voir [Connecter de vraies passerelles](14-connecter-passerelles-reelles.md).

---

## Le bilan de fin d'installation

Quel que soit le mode, l'installeur ne se contente pas de constater que les paquets sont posés :
il **établit les connexions**, comme le fera l'application.

| Vérification | Ce qui est réellement testé |
|---|---|
| **Service** | La console répond sur `https://…:7926/healthz`. |
| **Base de données** | Une connexion est **ouverte** avec le DSN configuré, et les migrations sont comptées. |
| **strongSwan (VICI)** | `swanctl` est appelé **sous l'identité du service** (`swanmgr`), pas en root — c'est exactement ce que fait la console. Un test en root passerait là où le service échouerait. |

### strongSwan injoignable ⇒ l'installation échoue

Si strongSwan est installé mais que la console ne parvient pas à lui parler, **l'installation
s'arrête en erreur**. C'est volontaire : sans VICI, la console démarrerait en **mode démo** et
afficherait des tunnels **simulés** — donnant l'illusion d'un VPN alors qu'**aucun trafic n'est
chiffré**. Une installation qui s'arrête est moins dangereuse qu'une console qui ment.

Le service et la configuration restent en place : il n'y a **rien à réinstaller**. Corrigez le
point signalé, puis relancez la vérification :

```bash
swanmgrctl doctor
```

> Vous pilotez uniquement des passerelles **distantes** ? Installez avec `--no-strongswan` et
> renseignez `VICI_ENDPOINTS` dans la configuration.

---

## La première connexion

### Le mot de passe doit être changé

Les quatre comptes (`admin`, `operator`, `auditor`, `viewer`) sont créés au premier démarrage
avec **le même mot de passe**, celui que l'installeur a tiré au hasard et écrit dans le fichier
de configuration. La console **impose de le changer** : tant que ce n'est pas fait, l'API
répond `403` sur tout le reste — le blocage n'est pas qu'un écran, il est dans le serveur.

Les trois autres comptes partagent ce mot de passe : traitez-les aussi, ou désactivez-les.

### L'avertissement du navigateur

C'est normal, et ce n'est pas un défaut : l'application a **généré son propre certificat**,
signé par sa CA interne. Votre navigateur ne connaît pas cette autorité, donc il vous prévient.

> **La connexion est bel et bien chiffrée.** Ce qui n'est pas attesté, c'est l'*identité* du
> serveur — pas la confidentialité de l'échange. C'est exactement le comportement de Proxmox,
> pfSense ou TrueNAS au premier démarrage.

Pour continuer : **Avancé** → **Continuer vers…**

Pour faire disparaître l'avertissement durablement :

| Vous voulez… | Faites |
|---|---|
| Garder la CA interne | Importez-la dans le magasin de confiance de vos postes d'administration (écran **PKI & Certificats**). À faire **une seule fois**. |
| Un certificat **reconnu** | `ACME_DOMAIN=vpn.mondomaine.fr` → **Let's Encrypt**. Exige un **domaine public** et le **port 80 joignable**, redirigé vers le port 7927. |
| Utiliser **votre** certificat | `TLS_CERT` et `TLS_KEY` |
| Terminer le TLS sur un **reverse proxy** | `TLS_ENABLED=false` |

Import de la CA sur un poste d'administration :

- **macOS** : double-cliquez sur `ca.crt` → Trousseau *Système* → « Toujours approuver ».
- **Linux** : `sudo cp ca.crt /usr/local/share/ca-certificates/ && sudo update-ca-certificates`
- **Windows** : `certutil -addstore -f Root ca.crt` (en administrateur).

⚠️ Si vous accédez à la console par un nom de domaine, **déclarez-le** dans `TLS_SANS`
(ex. `TLS_SANS="localhost,127.0.0.1,vpn.interne.fr"`) — sinon le navigateur signalera, en plus,
une incohérence de nom.

---

## Les deux ports

| Port | Protocole | Sert |
|---|---|---|
| **7926** | **HTTPS** | L'interface, l'API, le temps réel. |
| **7927** | HTTP en clair | Uniquement la **CRL** (`/crl.der`) et `/healthz`. Le reste est redirigé vers HTTPS. |

Le port 7927 n'est pas un oubli : le point de distribution de CRL **doit** rester en HTTP, car
c'est charon qui le lit et il ne ferait pas confiance à notre CA interne — il lui faudrait,
pour la valider, la CRL qu'il est justement en train de récupérer (RFC 5280).

---

## Exploiter : `swanmgrctl`

L'installeur pose un outil de cycle de vie :

| Commande | Effet |
|---|---|
| `swanmgrctl doctor` | Diagnostic complet : base, socket VICI, certificat, ports, pare-feu. **La première chose à lancer quand quelque chose cloche.** |
| `swanmgrctl backup` | Sauvegarde la base **et** `SECRETS_KEY` dans une archive. |
| `swanmgrctl restore ARCHIVE` | Restaure. Refuse si la clé de l'archive ne correspond pas à l'installation. |
| `swanmgrctl upgrade` | Sauvegarde, met à jour, vérifie, et **revient en arrière tout seul** si la console ne répond plus. |
| `swanmgrctl status` / `logs` | Raccourcis systemd. |

### La sauvegarde : la base seule ne suffit pas

`SECRETS_KEY` chiffre vos secrets, la **clé privée de la CA** et celle du **certificat TLS**.
Un dump PostgreSQL sans cette clé est **définitivement illisible** — et le serveur refuserait
même de démarrer. Il n'existe aucune procédure de récupération : c'est pourquoi
`swanmgrctl backup` archive les deux ensemble, et pourquoi `restore` vérifie qu'elles
correspondent avant de toucher à quoi que ce soit.

> L'archive contient donc `SECRETS_KEY` **en clair**. Traitez-la comme un secret.

---

## Désinstaller

```bash
sudo /usr/share/strongswan-manager/uninstall.sh           # conserve base et configuration
sudo /usr/share/strongswan-manager/uninstall.sh --purge   # ⚠️ efface tout, irréversible
```

Sans `--purge`, une réinstallation ultérieure retrouve vos tunnels, votre PKI et votre audit.
PostgreSQL et strongSwan ne sont jamais désinstallés (d'autres services peuvent en dépendre).

---

## Et ensuite ?

→ [Découvrir l'interface](03-decouvrir-linterface.md)
