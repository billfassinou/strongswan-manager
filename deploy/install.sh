#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Installe StrongSwan Manager en service système autonome (systemd).
#
# Deux façons de l'utiliser :
#
#   1. En ligne  — curl -fsSL https://raw.githubusercontent.com/billfassinou/strongswan-manager/main/deploy/install.sh | sudo bash
#      Le script télécharge le bundle de la dernière version, en vérifie l'intégrité,
#      puis se relance depuis celui-ci.
#
#   2. DEPUIS LES SOURCES — après avoir cloné le dépôt :
#      git clone https://github.com/billfassinou/strongswan-manager.git
#      cd strongswan-manager && sudo ./deploy/install.sh --from-source
#      Compile l'interface web puis le binaire, et installe le résultat exactement comme le
#      ferait le bundle. Si Go (≥ 1.23) ou Node manquent, les chaînes OFFICIELLES sont
#      récupérées dans un dossier temporaire — rien n'est installé durablement.
#
#   3. HORS LIGNE — depuis le bundle déjà téléchargé (air-gap) :
#      tar xzf strongswan-manager_vX.Y.Z_linux_amd64.tar.gz
#      cd strongswan-manager_vX.Y.Z_linux_amd64 && sudo ./install.sh
#      Aucun accès réseau n'est alors nécessaire — sauf pour installer PostgreSQL et
#      strongSwan depuis les dépôts de la distribution. Si ces paquets sont déjà présents
#      (ou fournis par un miroir local), ajoutez --skip-deps : le script ne touche plus au
#      réseau du tout.
#
# Ce que ce script MODIFIE sur la machine (il le récapitule et demande confirmation) :
#   - installe PostgreSQL (et strongSwan si absent) via le gestionnaire de paquets ;
#   - crée l'utilisateur système « swanmgr » (sans shell, sans mot de passe) ;
#   - crée la base « swan » et son utilisateur, avec un mot de passe aléatoire ;
#   - pose /usr/bin/strongswan-manager, /usr/bin/swanmgrctl,
#     /etc/strongswan-manager/, l'unité systemd ;
#   - pose un drop-in sur strongswan.service pour ouvrir le socket VICI au groupe swanmgr ;
#   - ouvre les ports 7926/7927 dans firewalld ou ufw, si l'un des deux est actif.
#
# Mise à jour, sauvegarde, diagnostic : swanmgrctl (voir « swanmgrctl --help »).
# Désinstallation : uninstall.sh

set -euo pipefail

REPO="${REPO:-billfassinou/strongswan-manager}"

ASSUME_YES="${ASSUME_YES:-0}"
WITH_STRONGSWAN=1
SKIP_DEPS=0
FROM_SOURCE=0
VERSION="latest"

usage() {
  cat <<'EOF'
Usage : install.sh [options]

  --from-source        Compile depuis le dépôt cloné, puis installe (voir plus haut).
                       Go et Node sont récupérés automatiquement s'ils manquent.
  --version vX.Y.Z     Installe une version précise (défaut : la dernière).
                       Ignoré depuis un bundle ou avec --from-source.
  --no-strongswan      N'installe pas strongSwan (console pilotant des passerelles distantes)
  --skip-deps          N'installe aucun paquet (PostgreSQL/strongSwan supposés présents).
                       C'est l'option des installations hors ligne.
  --yes                Ne pose aucune question
  --help
EOF
  exit 0
}

while [ $# -gt 0 ]; do
  case "$1" in
    --from-source) FROM_SOURCE=1; shift ;;
    --version) VERSION="$2"; shift 2 ;;
    --no-strongswan) WITH_STRONGSWAN=0; shift ;;
    --skip-deps) SKIP_DEPS=1; shift ;;
    --yes|-y) ASSUME_YES=1; shift ;;
    --help|-h) usage ;;
    *) printf 'option inconnue : %s (--help)\n' "$1" >&2; exit 1 ;;
  esac
done

# --- Bundle ou amorçage réseau ? --------------------------------------------
#
# Le bundle est auto-suffisant : binaire, unité, swanmgrctl et lib/ sont côte à côte.
# Si on ne les trouve pas (cas du « curl … | bash », où le script n'a même pas de
# répertoire), on télécharge le bundle et on se relance depuis lui : le chemin
# d'installation réel est ainsi TOUJOURS le chemin hors-ligne, donc toujours testé.

SELF="${BASH_SOURCE[0]:-}"
SELF_DIR=""
if [ -n "$SELF" ] && [ -f "$SELF" ]; then
  SELF_DIR="$(cd "$(dirname "$SELF")" && pwd)"
fi

# --- Mode « depuis les sources » --------------------------------------------
#
# On compile, puis on reconstitue un bundle : la suite est alors STRICTEMENT le même chemin de
# code que l'installation par binaire. Un seul chemin à maintenir, un seul à tester.
if [ "$FROM_SOURCE" -eq 1 ]; then
  [ -n "$SELF_DIR" ] || { printf '✘ --from-source exige de lancer le script depuis le dépôt cloné (./deploy/install.sh), pas via un pipe.\n' >&2; exit 1; }
  REPO_ROOT="$(cd "$SELF_DIR/.." && pwd)"
  [ -f "$REPO_ROOT/deploy/lib/common.sh" ] || { printf '✘ dépôt introuvable depuis %s\n' "$SELF_DIR" >&2; exit 1; }

  # shellcheck source=lib/common.sh
  . "$REPO_ROOT/deploy/lib/common.sh"
  need_root
  command -v git  >/dev/null 2>&1 || warn "git absent : la version du binaire sera « dev »."
  command -v curl >/dev/null 2>&1 || die "curl est requis pour récupérer la chaîne d'outils."
  command -v tar  >/dev/null 2>&1 || die "tar est requis."

  info "compilation depuis les sources ($REPO_ROOT)"
  build_from_source "$REPO_ROOT"
  SELF_DIR="$SOURCE_STAGE"
  SWANMGR_VERSION="$("$SELF_DIR/strongswan-manager" --version 2>/dev/null || echo dev)"
fi

if [ -z "$SELF_DIR" ] || [ ! -f "$SELF_DIR/lib/common.sh" ] || [ ! -x "$SELF_DIR/strongswan-manager" ]; then
  [ "${SWANMGR_BOOTSTRAPPED:-0}" -eq 1 ] && {
    printf '✘ bundle incomplet : lib/common.sh ou le binaire manquent.\n' >&2; exit 1; }

  [ "$(id -u)" -eq 0 ] || { printf '✘ ce script doit être lancé en root (sudo).\n' >&2; exit 1; }
  command -v curl >/dev/null 2>&1 || { printf '✘ curl est requis pour le téléchargement.\n' >&2; exit 1; }
  command -v tar  >/dev/null 2>&1 || { printf '✘ tar est requis.\n' >&2; exit 1; }

  case "$(uname -m)" in
    x86_64|amd64)  arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) printf '✘ architecture non prise en charge : %s\n' "$(uname -m)" >&2; exit 1 ;;
  esac

  if [ "$VERSION" = latest ]; then
    # On récupère la réponse ENTIÈRE avant de l'analyser. Écrire « curl … | grep -m1 … »
    # est un piège : grep s'arrête au premier match et ferme le tube, curl reçoit un EPIPE
    # et sort en 23 ; pipefail propage ce 23 et set -e tue le script SANS UN MOT. C'est une
    # course (elle passe si la réponse tient dans le tampon), donc un échec intermittent.
    api="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest")" \
      || { printf '✘ impossible d'"'"'interroger l'"'"'API GitHub (réseau ?).\n' >&2; exit 1; }
    VERSION="$(printf '%s\n' "$api" | grep -m1 '"tag_name"' | cut -d'"' -f4)"
    [ -n "$VERSION" ] || { printf '✘ impossible de déterminer la dernière version.\n' >&2; exit 1; }
  fi

  printf '▸ téléchargement du bundle %s (%s)\n' "$VERSION" "$arch"
  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
  base="https://github.com/$REPO/releases/download/$VERSION"
  name="strongswan-manager_${VERSION}_linux_${arch}"

  curl -fsSL "$base/$name.tar.gz" -o "$tmp/bundle.tar.gz" \
    || { printf '✘ téléchargement de %s.tar.gz échoué.\n' "$name" >&2; exit 1; }
  curl -fsSL "$base/SHA256SUMS" -o "$tmp/SHA256SUMS" \
    || { printf '✘ téléchargement de SHA256SUMS échoué.\n' >&2; exit 1; }

  # On NE fait PAS confiance à une archive non vérifiée.
  expected="$(awk -v f="$name.tar.gz" '$2 == f || $2 == "*"f {print $1}' "$tmp/SHA256SUMS" | head -1)"
  actual="$(sha256sum "$tmp/bundle.tar.gz" | cut -d' ' -f1)"
  if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
    printf '✘ SOMME DE CONTRÔLE INVALIDE — archive corrompue ou altérée. Installation interrompue.\n' >&2
    exit 1
  fi
  printf '✔ somme de contrôle SHA-256 vérifiée\n'

  # Signature cosign (keyless, OIDC GitHub) : vérifiée si l'outil est disponible. Elle
  # atteste l'origine de SHA256SUMS, ce que le condensat seul ne peut pas faire.
  if command -v cosign >/dev/null 2>&1 \
     && curl -fsSL "$base/SHA256SUMS.sig" -o "$tmp/SHA256SUMS.sig" 2>/dev/null \
     && curl -fsSL "$base/SHA256SUMS.pem" -o "$tmp/SHA256SUMS.pem" 2>/dev/null; then
    if cosign verify-blob "$tmp/SHA256SUMS" \
         --signature "$tmp/SHA256SUMS.sig" --certificate "$tmp/SHA256SUMS.pem" \
         --certificate-identity-regexp "^https://github.com/$REPO/" \
         --certificate-oidc-issuer https://token.actions.githubusercontent.com >/dev/null 2>&1; then
      printf '✔ signature cosign vérifiée (origine attestée)\n'
    else
      printf '✘ SIGNATURE COSIGN INVALIDE. Installation interrompue.\n' >&2
      exit 1
    fi
  else
    printf '! cosign absent : intégrité vérifiée (SHA-256), origine non attestée.\n'
  fi

  tar -xzf "$tmp/bundle.tar.gz" -C "$tmp"
  chmod +x "$tmp/$name/install.sh" "$tmp/$name/strongswan-manager" "$tmp/$name/swanmgrctl" 2>/dev/null || true

  args=()
  [ "$ASSUME_YES" -eq 1 ] && args+=(--yes)
  [ "$WITH_STRONGSWAN" -eq 0 ] && args+=(--no-strongswan)
  [ "$SKIP_DEPS" -eq 1 ] && args+=(--skip-deps)

  # Cas « curl … | sudo bash » : la FIN de ce script est encore dans le tube. Si on ré-exécute
  # sans l'avoir lue, curl meurt en écriture (« curl: (23) Failure writing output »), et le
  # script ré-exécuté lit ces octets résiduels comme réponse à sa confirmation. On draine donc
  # d'abord le tube — uniquement si stdin n'est pas déjà un terminal (lancement depuis un
  # fichier), sinon on bloquerait en attendant une saisie.
  [ -t 0 ] || cat >/dev/null 2>&1 || true

  # La confirmation du script ré-exécuté doit pouvoir être lue. On teste l'OUVERTURE de
  # /dev/tty (le nœud peut exister sans terminal de contrôle : cloud-init, CI) : si un terminal
  # est joignable, on le rebranche sur stdin ; sinon, installation non interactive (--yes).
  if { : < /dev/tty; } 2>/dev/null; then
    SWANMGR_BOOTSTRAPPED=1 SWANMGR_VERSION="$VERSION" \
      exec bash "$tmp/$name/install.sh" ${args[@]+"${args[@]}"} < /dev/tty
  else
    case " ${args[*]} " in *" --yes "*) : ;; *) args+=(--yes) ;; esac
    SWANMGR_BOOTSTRAPPED=1 SWANMGR_VERSION="$VERSION" \
      exec bash "$tmp/$name/install.sh" ${args[@]+"${args[@]}"} < /dev/null
  fi
fi

# --- À partir d'ici : on est DANS le bundle ---------------------------------

# shellcheck source=lib/common.sh
. "$SELF_DIR/lib/common.sh"

need_root
command -v systemctl >/dev/null 2>&1 || die "systemd est requis (systemctl introuvable)."
command -v openssl   >/dev/null 2>&1 || die "openssl est requis."

PKG="$(detect_pkg)"
ARCH="$(detect_arch)"
VERSION="${SWANMGR_VERSION:-${VERSION#latest}}"
[ -n "$VERSION" ] || VERSION="$(basename "$SELF_DIR" | sed -n 's/^strongswan-manager_\(v[^_]*\)_.*/\1/p')"
[ -n "$VERSION" ] || VERSION="(bundle local)"

[ "$SKIP_DEPS" -eq 1 ] || [ "$PKG" != none ] \
  || die "gestionnaire de paquets non reconnu (ni dnf, ni apt-get). Relancez avec --skip-deps."

# --- Récapitulatif et confirmation ------------------------------------------

cat <<EOF

  StrongSwan Manager — installation en service système

  Machine      : $(uname -srm)$([ "$PKG" != none ] && echo ", $PKG")
  Version      : $VERSION ($ARCH)
  Source       : bundle local — $SELF_DIR
  Binaire      : $BIN
  Configuration: $ENV_FILE
  Service      : $SVC_NAME.service (utilisateur « $SVC_USER »)
  Ports        : $HTTPS_PORT (HTTPS) et $HTTP_PORT (HTTP, CRL + redirection)

$(if [ "$SKIP_DEPS" -eq 1 ]; then
    echo "  Aucun paquet ne sera installé (--skip-deps) : PostgreSQL doit déjà être présent."
  else
    echo "  Seront installés si absents : PostgreSQL$([ "$WITH_STRONGSWAN" -eq 1 ] && echo ", strongSwan")"
  fi)

EOF

if [ "$ASSUME_YES" -ne 1 ]; then
  read -rp "  Continuer ? [o/N] " a
  case "$a" in o|O|y|Y|oui|yes) ;; *) die "annulé." ;; esac
fi

# --- Paquets ----------------------------------------------------------------

if [ "$SKIP_DEPS" -eq 1 ]; then
  pg_available || die "psql introuvable et --skip-deps demandé.
  Installez PostgreSQL (par ex. « dnf install postgresql-server » ou « apt install postgresql »),
  ou relancez sans --skip-deps si la machine a accès aux dépôts."
  if [ "$WITH_STRONGSWAN" -eq 1 ] && ! command -v swanctl >/dev/null 2>&1; then
    warn "swanctl introuvable : la console démarrera en mode démo (aucun tunnel réel)."
    WITH_STRONGSWAN=0
  fi
  ok "aucun paquet installé (--skip-deps)"
else
  pkgs=""
  command -v curl >/dev/null 2>&1 || pkgs="$pkgs curl"
  command -v tar  >/dev/null 2>&1 || pkgs="$pkgs tar"

  if [ "$PKG" = dnf ]; then
    # Un dépôt tiers cassé (typiquement un dépôt qui pointe encore sur el/8 depuis un el/9)
    # fait échouer TOUT « dnf install », même quand le paquet demandé n'en dépend en rien :
    # dnf refuse d'agir s'il ne peut pas rafraîchir un dépôt activé. On lui demande donc
    # d'ignorer les dépôts injoignables. Si le dépôt manquant était réellement nécessaire,
    # l'erreur qui suivra sera explicite (paquet introuvable), pas un échec de métadonnées.
    DNF="dnf -y --setopt=*.skip_if_unavailable=1"

    pg_available || pkgs="$pkgs postgresql-server postgresql-contrib"
    if [ "$WITH_STRONGSWAN" -eq 1 ] && ! command -v swanctl >/dev/null 2>&1; then
      # strongSwan n'est PAS dans les dépôts de base de RHEL/AlmaLinux/Rocky : il vient d'EPEL.
      # Sans ce dépôt, « dnf install strongswan » échoue (vérifié sur AlmaLinux 9).
      if ! $DNF -q list strongswan >/dev/null 2>&1; then
        info "activation du dépôt EPEL (strongSwan n'est pas dans les dépôts de base)"
        $DNF install epel-release >/dev/null 2>&1 \
          || $DNF install "https://dl.fedoraproject.org/pub/epel/epel-release-latest-$(rpm -E %rhel).noarch.rpm" >/dev/null 2>&1 \
          || die "impossible d'activer EPEL. Relancez avec --no-strongswan, ou installez strongSwan vous-même."
        $DNF -q makecache >/dev/null 2>&1 || true
      fi
      pkgs="$pkgs strongswan"
    fi
  else
    pg_available || pkgs="$pkgs postgresql postgresql-contrib"
    if [ "$WITH_STRONGSWAN" -eq 1 ] && ! command -v swanctl >/dev/null 2>&1; then
      # Sur Debian/Ubuntu, swanctl et le plugin vici sont dans un paquet séparé.
      pkgs="$pkgs strongswan strongswan-swanctl charon-systemd"
    fi
  fi

  if [ -n "$pkgs" ]; then
    info "installation des paquets :$pkgs"
    if [ "$PKG" = dnf ]; then
      # shellcheck disable=SC2086
      $DNF install $pkgs >/dev/null || die "installation des paquets échouée :$pkgs

  Si l'erreur mentionne un dépôt tiers (métadonnées, repomd.xml, « All mirrors were tried »),
  ce dépôt est cassé sur cette machine et bloque dnf, indépendamment de StrongSwan Manager.
  Désactivez-le, puis relancez :
      dnf repolist                                  # repérer le dépôt fautif
      dnf config-manager --set-disabled <son-nom>

  Ou installez PostgreSQL vous-même et relancez avec --skip-deps."
    else
      export DEBIAN_FRONTEND=noninteractive
      # Une source apt cassée ne doit pas empêcher l'installation si les listes en cache
      # suffisent : on avertit plutôt que d'abandonner.
      apt-get update -qq >/dev/null 2>&1 || warn "« apt-get update » a échoué (source cassée ?) — on tente avec les listes en cache."
      # shellcheck disable=SC2086
      apt-get install -y -qq $pkgs >/dev/null || die "installation des paquets échouée :$pkgs

  Si l'erreur vient d'une source apt tierce cassée, corrigez /etc/apt/sources.list.d/,
  ou installez PostgreSQL vous-même et relancez avec --skip-deps."
    fi
    ok "paquets installés"
  fi
fi

# --- PostgreSQL --------------------------------------------------------------

info "PostgreSQL"
pg_start_and_wait
DB_PASS="$(provision_db)"

# --- Utilisateur, binaires, configuration -----------------------------------

ensure_user
migrate_from_usr_local

install -m 0755 "$SELF_DIR/strongswan-manager" "$BIN"
install -m 0755 "$SELF_DIR/swanmgrctl" "$BIN_DIR/swanmgrctl"
install -d -m 0755 "$SHARE_DIR/lib"
install -m 0644 "$SELF_DIR/lib/common.sh" "$SHARE_DIR/lib/common.sh"
install -m 0755 "$SELF_DIR/uninstall.sh" "$SHARE_DIR/uninstall.sh"
ok "binaires installés : $BIN, $BIN_DIR/swanmgrctl"

write_env_file "$DB_PASS" "$WITH_STRONGSWAN"

# --- Accès au socket VICI ---------------------------------------------------

if [ "$WITH_STRONGSWAN" -eq 1 ] && command -v swanctl >/dev/null 2>&1; then
  install_vici_dropin || true
fi

# --- Service ----------------------------------------------------------------

info "installation du service"
install -m 0644 "$SELF_DIR/strongswan-manager.service" "$UNIT"
install -d -m 0750 -o "$SVC_USER" -g "$SVC_USER" "$STATE_DIR"
apply_db_dependency
systemctl enable --now "$SVC_NAME" >/dev/null

open_firewall

# --- Vérification : on prouve, on ne suppose pas ------------------------------

echo
info "vérification de l'installation"

TROUBLES=0

# 1. Le service répond.
if wait_health 40; then
  ok "service actif et console joignable (HTTPS :$HTTPS_PORT)"
else
  die "le service ne répond pas.
  Diagnostic : swanmgrctl doctor
  Journaux   : journalctl -u $SVC_NAME -n 50 --no-pager"
fi

# 2. La base : on ÉTABLIT la connexion avec le DSN configuré.
if verify_db; then
  ok "base de données joignable ($(psql "$(env_get DATABASE_URL)" -tAc 'SELECT count(*) FROM schema_migrations' 2>/dev/null || echo '?') migrations appliquées)"
else
  rc=$?
  if [ "$rc" -eq 2 ]; then
    warn "connexion à la base non vérifiable ici (client psql absent) — le service, lui, a démarré."
  else
    TROUBLES=1
    warn "BASE DE DONNÉES INJOIGNABLE avec le DSN de $ENV_FILE."
    warn "  Vérifiez DATABASE_URL, puis : systemctl restart $SVC_NAME"
  fi
fi

# 3. strongSwan : on vérifie que l'utilisateur DU SERVICE parle réellement à charon.
#    Un test en root ne prouverait rien — c'est « swanmgr » qui doit y arriver.
if [ "$WITH_STRONGSWAN" -eq 1 ]; then
  if verify_vici; then
    ok "strongSwan piloté par VICI : $(runuser -u "$SVC_USER" -- swanctl --version 2>/dev/null || swanctl --version 2>/dev/null)"
  else
    # Échec dur, et assumé : une console qui affiche des tunnels SIMULÉS en croyant piloter
    # de vrais tunnels est plus dangereuse qu'une installation qui s'arrête.
    SW_UNIT="$(strongswan_unit)"
    die "STRONGSWAN EST INSTALLÉ MAIS INJOIGNABLE PAR LA CONSOLE.

  L'utilisateur du service (« $SVC_USER ») ne parvient pas à parler à charon via VICI.
  Laissée ainsi, la console démarrerait en MODE DÉMO : elle afficherait des tunnels
  SIMULÉS, sans qu'aucun trafic ne soit réellement chiffré. L'installation s'arrête ici.

  À vérifier, dans cet ordre :
    1. charon tourne-t-il ?          systemctl status $SW_UNIT
    2. le socket existe-t-il ?       ls -l /run/charon.vici
    3. est-il ouvert au groupe ?     stat -c '%G %a' /run/charon.vici   → « $SVC_USER 660 »
       (c'est le rôle du drop-in /etc/systemd/system/$SW_UNIT.d/10-vici-swanmgr.conf)
    4. le plugin vici est-il chargé ? $SW_UNIT doit lister « vici » dans ses plugins

  Le service et la configuration SONT en place : corrigez le point ci-dessus, puis
  relancez la vérification avec «  swanmgrctl doctor  ». Rien n'est à réinstaller.

  Vous pilotez des passerelles DISTANTES et pas de charon local ? Relancez avec
  --no-strongswan, et renseignez VICI_ENDPOINTS dans $ENV_FILE."
  fi
else
  TROUBLES=1
  warn "MODE DÉMO : aucune passerelle strongSwan locale n'est pilotée (--no-strongswan)."
  warn "  Les tunnels affichés seront SIMULÉS. Renseignez VICI_ENDPOINTS dans $ENV_FILE"
  warn "  pour piloter de vraies passerelles, puis : systemctl restart $SVC_NAME"
fi

IP="$(host_ip)"; IP="${IP:-127.0.0.1}"
PASS="$(env_get SEED_ADMIN_PASSWORD || true)"

echo
if [ "$TROUBLES" -eq 0 ]; then
  echo "  ✔ Tout est opérationnel."
else
  echo "  ! Installation terminée, MAIS des points demandent votre attention (ci-dessus)."
fi

cat <<EOF

    Console    https://$IP:$HTTPS_PORT
    Identifiant admin
    Mot de passe $PASS   (la console vous demandera de le changer à la première connexion)

  Le navigateur affichera un AVERTISSEMENT au premier accès : le certificat est
  auto-signé (la connexion est chiffrée, c'est l'identité du serveur qui n'est pas
  attestée par un tiers). Pour le supprimer, importez la CA interne, ou renseignez
  ACME_DOMAIN dans $ENV_FILE si la machine a un domaine public.

    Diagnostic    swanmgrctl doctor
    Sauvegarde    swanmgrctl backup
    Mise à jour   swanmgrctl upgrade
    Journaux      journalctl -u $SVC_NAME -f
    Configuration $ENV_FILE
    Désinstaller  $SHARE_DIR/uninstall.sh

  ⚠️ Sauvegardez $ENV_FILE : il contient SECRETS_KEY, sans laquelle vos secrets et vos
     clés privées seraient DÉFINITIVEMENT illisibles, même avec une sauvegarde de la base.
     « swanmgrctl backup » archive la base ET cette clé ensemble.

  Documentation : https://billfassinou.github.io/strongswan-manager/docs/

EOF
