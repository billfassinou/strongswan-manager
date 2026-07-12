#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Fonctions partagées par install.sh, uninstall.sh, swanmgrctl et les scriptlets
# post-installation des paquets .deb/.rpm. Ce fichier est *sourcé*, jamais exécuté.
#
# Règle d'or de tout ce dossier : NE JAMAIS régénérer SECRETS_KEY si le fichier de
# configuration existe déjà. Elle déchiffre les secrets, la clé de la CA et la clé du
# certificat TLS ; la perdre rend la base définitivement illisible.

# --- Emplacements (surchargeables pour les tests) ----------------------------

REPO="${REPO:-billfassinou/strongswan-manager}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
BIN="${BIN:-$BIN_DIR/strongswan-manager}"
ETC_DIR="${ETC_DIR:-/etc/strongswan-manager}"
ENV_FILE="${ENV_FILE:-$ETC_DIR/strongswan-manager.env}"
STATE_DIR="${STATE_DIR:-/var/lib/strongswan-manager}"
SVC_USER="${SVC_USER:-swanmgr}"
SVC_NAME="${SVC_NAME:-strongswan-manager}"
UNIT="${UNIT:-/etc/systemd/system/$SVC_NAME.service}"
DB_NAME="${DB_NAME:-swan}"
DB_USER="${DB_USER:-swan}"
HTTPS_PORT="${HTTPS_PORT:-7926}"
HTTP_PORT="${HTTP_PORT:-7927}"

# --- Sortie -----------------------------------------------------------------

if [ -t 1 ]; then
  info() { printf '\033[1;34m▸\033[0m %s\n' "$*"; }
  ok()   { printf '\033[1;32m✔\033[0m %s\n' "$*"; }
  warn() { printf '\033[1;33m!\033[0m %s\n' "$*"; }
  die()  { printf '\033[1;31m✘ %s\033[0m\n' "$*" >&2; exit 1; }
else
  info() { printf '▸ %s\n' "$*"; }
  ok()   { printf '✔ %s\n' "$*"; }
  warn() { printf '! %s\n' "$*"; }
  die()  { printf '✘ %s\n' "$*" >&2; exit 1; }
fi

need_root() { [ "$(id -u)" -eq 0 ] || die "ce script doit être lancé en root (sudo)."; }

# --- Détection de la plateforme ---------------------------------------------

# detect_pkg → dnf | apt | none
detect_pkg() {
  if   command -v dnf >/dev/null 2>&1; then echo dnf
  elif command -v apt-get >/dev/null 2>&1; then echo apt
  else echo none
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    *) die "architecture non prise en charge : $(uname -m) (amd64 et arm64 uniquement)." ;;
  esac
}

# --- Fichier de configuration -----------------------------------------------

# env_get CLÉ [FICHIER] — lit une valeur sans dépendre du PCRE de GNU grep
# (busybox et le grep de BSD n'ont pas -P).
env_get() {
  local key="$1" file="${2:-$ENV_FILE}"
  [ -f "$file" ] || return 1
  sed -n "s/^[[:space:]]*${key}=//p" "$file" | head -1
}

# Empreinte de SECRETS_KEY : permet de vérifier qu'une sauvegarde correspond bien à
# l'installation en place, sans jamais écrire la clé elle-même dans un manifeste.
secrets_key_fingerprint() {
  local key="${1:-}"
  [ -n "$key" ] || key="$(env_get SECRETS_KEY || true)"
  [ -n "$key" ] || return 1
  printf '%s' "$key" | sha256sum | cut -c1-16
}

random_hex() { openssl rand -hex "${1:-32}"; }

# --- Répertoires temporaires -------------------------------------------------
#
# Un seul trap, posé ici. Une fonction qui poserait son propre « trap … EXIT » sur une
# variable locale la verrait disparaître avant que le trap ne s'exécute (« unbound
# variable » sous set -u), et écraserait au passage le trap de son appelant.

SWANMGR_TMPDIRS=""

_swanmgr_cleanup() {
  local d
  for d in $SWANMGR_TMPDIRS; do rm -rf "$d"; done
}
trap _swanmgr_cleanup EXIT

# make_tmpdir → crée un répertoire temporaire privé, nettoyé à la sortie du script.
make_tmpdir() {
  local d
  d="$(mktemp -d)"
  chmod 0700 "$d"
  SWANMGR_TMPDIRS="$SWANMGR_TMPDIRS $d"
  printf '%s' "$d"
}

# --- Utilisateur système ----------------------------------------------------

ensure_user() {
  id "$SVC_USER" >/dev/null 2>&1 && return 0
  useradd --system --no-create-home --shell /usr/sbin/nologin "$SVC_USER" 2>/dev/null \
    || useradd --system --no-create-home --shell /sbin/nologin "$SVC_USER" \
    || die "création de l'utilisateur système « $SVC_USER » impossible."
  ok "utilisateur système « $SVC_USER » créé"
}

# --- PostgreSQL --------------------------------------------------------------

pg_available() { command -v psql >/dev/null 2>&1; }

# Un PostgreSQL local est-il présent ET piloté par systemd ?
pg_local() {
  pg_available && systemctl list-unit-files --type=service --no-legend 2>/dev/null \
    | awk '{print $1}' | grep -qx 'postgresql.service'
}

pg_start_and_wait() {
  # RHEL/AlmaLinux : le paquet n'initialise PAS le cluster, contrairement à Debian.
  if [ "$(detect_pkg)" = dnf ] && [ ! -s /var/lib/pgsql/data/PG_VERSION ]; then
    postgresql-setup --initdb >/dev/null 2>&1 || /usr/bin/postgresql-setup --initdb >/dev/null \
      || die "initialisation du cluster PostgreSQL échouée."
    ok "cluster PostgreSQL initialisé"
  fi
  systemctl enable --now postgresql >/dev/null 2>&1 \
    || die "PostgreSQL n'a pas démarré (systemctl status postgresql)."
  local i
  for i in $(seq 30); do
    su - postgres -c "psql -tAc 'SELECT 1'" >/dev/null 2>&1 && return 0
    sleep 1
  done
  die "PostgreSQL ne répond pas."
}

# provision_db → écrit le mot de passe de la base sur la sortie standard.
# Réutilise la base existante si le rôle est déjà là ; ne touche au mot de passe que
# lorsqu'aucune configuration ne le référence plus.
provision_db() {
  local pass
  if su - postgres -c "psql -tAc \"SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'\"" 2>/dev/null | grep -q 1; then
    if [ -f "$ENV_FILE" ]; then
      pass="$(env_get DATABASE_URL | sed -n "s|^postgres://$DB_USER:\([^@]*\)@.*|\1|p")"
      [ -n "$pass" ] || warn "mot de passe de la base illisible dans $ENV_FILE"
    else
      pass="$(random_hex 24)"
      su - postgres -c "psql -c \"ALTER ROLE $DB_USER WITH PASSWORD '$pass'\"" >/dev/null
      warn "mot de passe de la base régénéré (aucune configuration précédente trouvée)" >&2
    fi
  else
    pass="$(random_hex 24)"
    su - postgres -c "psql -c \"CREATE ROLE $DB_USER LOGIN PASSWORD '$pass'\"" >/dev/null
    su - postgres -c "createdb -O $DB_USER $DB_NAME" >/dev/null
    ok "base « $DB_NAME » et utilisateur « $DB_USER » créés" >&2
  fi
  printf '%s' "$pass"
}

# --- Génération de la configuration -----------------------------------------

default_sans() {
  local host_ip sans
  host_ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
  sans="localhost,127.0.0.1,::1,$(hostname -f 2>/dev/null || hostname)"
  [ -n "$host_ip" ] && sans="$sans,$host_ip"
  printf '%s' "$sans"
}

host_ip() { hostname -I 2>/dev/null | awk '{print $1}'; }

# write_env_file DB_PASS WITH_VICI(0|1)
# Ne fait RIEN si le fichier existe déjà : régénérer SECRETS_KEY détruirait les données.
write_env_file() {
  local db_pass="$1" with_vici="${2:-0}" ip sans admin_pass

  install -d -m 0750 -o root -g "$SVC_USER" "$ETC_DIR"

  if [ -f "$ENV_FILE" ]; then
    ok "configuration existante conservée (secrets inchangés)"
    return 0
  fi

  ip="$(host_ip)"; ip="${ip:-127.0.0.1}"
  sans="$(default_sans)"
  admin_pass="$(openssl rand -base64 18 | tr -d '/+=' | cut -c1-16)"

  local old_umask; old_umask="$(umask)"
  umask 077
  cat > "$ENV_FILE" <<EOF
# StrongSwan Manager — configuration du service.
# Générée le $(date -u +%Y-%m-%dT%H:%M:%SZ).
#
# ⚠️ SECRETS_KEY chiffre les secrets, les clés privées de la PKI et le certificat TLS.
#    NE LA CHANGEZ JAMAIS après le premier démarrage : les données deviendraient illisibles.
#    Sauvegardez ce fichier EN MÊME TEMPS que la base — « swanmgrctl backup » fait les deux.

DATABASE_URL=postgres://$DB_USER:$db_pass@127.0.0.1:5432/$DB_NAME?sslmode=disable
JWT_SECRET=$(random_hex 32)
SECRETS_KEY=$(random_hex 32)

# Mot de passe initial des 4 comptes (admin, operator, auditor, viewer). Il n'est utilisé
# qu'au TOUT PREMIER démarrage, sur une base vide, et la console impose de le changer à la
# première connexion.
SEED_ADMIN_PASSWORD=$admin_pass

HTTP_ADDR=:$HTTPS_PORT
HTTP_REDIRECT_ADDR=:$HTTP_PORT

# Noms et IP couverts par le certificat auto-généré. Ajoutez-y le nom par lequel vos
# administrateurs accèdent à la console, sinon leur navigateur signalera une incohérence.
TLS_SANS=$sans

# CRL Distribution Point — EN HTTP, sur l'écouteur clair : c'est charon qui le lit, et il
# ne ferait pas confiance à notre CA interne (RFC 5280). Fixez-le AVANT d'émettre des
# certificats : les certificats déjà émis embarquent cette URL telle quelle.
CRL_URL=http://$ip:$HTTP_PORT/crl.der
CRL_VALIDITY=24h

# Passerelles pilotées. Vide = mode démo (adaptateur simulé, aucun vrai tunnel).
# charon local :   VICI_ENDPOINTS=local=unix:/var/run/charon.vici
# à distance   :   VICI_ENDPOINTS=gw-paris=tcp:10.0.0.5:4502
VICI_ENDPOINTS=$([ "$with_vici" -eq 1 ] && echo "local=unix:/var/run/charon.vici")

POLL_INTERVAL=3s
CORS_ORIGINS=*

# Certificat reconnu par les navigateurs (exige un domaine public et le port 80 joignable,
# à rediriger vers $HTTP_PORT) :
# ACME_DOMAIN=vpn.mondomaine.fr
# ACME_EMAIL=admin@mondomaine.fr
ACME_CACHE=$STATE_DIR/acme
EOF
  umask "$old_umask"
  chown root:"$SVC_USER" "$ENV_FILE"
  chmod 0640 "$ENV_FILE"
  ok "configuration générée : $ENV_FILE (secrets aléatoires)"
}

# --- Socket VICI -------------------------------------------------------------

strongswan_unit() {
  local u
  u="$(systemctl list-unit-files --type=service --no-legend 2>/dev/null \
    | awk '{print $1}' | grep -E '^strongswan(-swanctl)?\.service$' | head -1)"
  printf '%s' "${u:-strongswan.service}"
}

vici_socket() {
  local s
  for s in /run/charon.vici /var/run/charon.vici; do
    [ -S "$s" ] && { printf '%s' "$s"; return 0; }
  done
  return 1
}

# charon RECRÉE le socket (0770, root:root) à chaque démarrage : un chmod ponctuel serait
# perdu au premier redémarrage. D'où ce drop-in, rejoué à chaque fois — c'est lui qui permet
# à la console de piloter charon SANS tourner en root.
install_vici_dropin() {
  local sw_unit; sw_unit="$(strongswan_unit)"
  mkdir -p "/etc/systemd/system/${sw_unit}.d"
  cat > "/etc/systemd/system/${sw_unit}.d/10-vici-swanmgr.conf" <<EOF
# Posé par StrongSwan Manager : ouvre le socket VICI au groupe « $SVC_USER ».
[Service]
ExecStartPost=/bin/sh -c 'for i in \$(seq 100); do [ -S /run/charon.vici ] && break; sleep 0.1; done; \
if [ -S /run/charon.vici ]; then chgrp $SVC_USER /run/charon.vici && chmod 0660 /run/charon.vici; fi'
EOF
  systemctl daemon-reload
  systemctl enable --now "$sw_unit" >/dev/null 2>&1 \
    || { warn "$sw_unit n'a pas démarré — la console démarrera en mode démo."; return 1; }
  systemctl restart "$sw_unit" >/dev/null 2>&1 || true
  sleep 2

  local sock; sock="$(vici_socket || true)"
  if [ -n "$sock" ] && [ "$(stat -c '%G' "$sock")" = "$SVC_USER" ]; then
    ok "socket VICI accessible au groupe « $SVC_USER » (la console ne tourne pas en root)"
    return 0
  fi
  warn "le socket VICI n'a pas pu être ouvert au groupe « $SVC_USER »."
  warn "la console démarrera en MODE DÉMO ; voir « journalctl -u $sw_unit »."
  return 1
}

remove_vici_dropin() {
  rm -f /etc/systemd/system/strongswan.service.d/10-vici-swanmgr.conf \
        /etc/systemd/system/strongswan-swanctl.service.d/10-vici-swanmgr.conf
  rmdir /etc/systemd/system/strongswan.service.d \
        /etc/systemd/system/strongswan-swanctl.service.d 2>/dev/null || true
  systemctl daemon-reload
  systemctl restart strongswan 2>/dev/null || true
}

# --- Pare-feu ----------------------------------------------------------------

open_firewall() {
  if systemctl is-active --quiet firewalld 2>/dev/null; then
    firewall-cmd --permanent --add-port="$HTTPS_PORT"/tcp >/dev/null && \
    firewall-cmd --permanent --add-port="$HTTP_PORT"/tcp  >/dev/null && \
    firewall-cmd --reload >/dev/null && ok "firewalld : ports $HTTPS_PORT et $HTTP_PORT ouverts"
  elif command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -q "^Status: active"; then
    ufw allow "$HTTPS_PORT"/tcp >/dev/null && ufw allow "$HTTP_PORT"/tcp >/dev/null \
      && ok "ufw : ports $HTTPS_PORT et $HTTP_PORT ouverts"
  else
    warn "aucun pare-feu actif détecté — pensez à autoriser les ports $HTTPS_PORT et $HTTP_PORT."
  fi
}

# --- Service -----------------------------------------------------------------

# Le service dépend d'un PostgreSQL LOCAL uniquement si la base est locale : avec une base
# distante, « Requires=postgresql.service » empêcherait purement et simplement le démarrage.
apply_db_dependency() {
  local url dropin="/etc/systemd/system/$SVC_NAME.service.d/10-local-postgres.conf"
  url="$(env_get DATABASE_URL || true)"
  if pg_local && printf '%s' "$url" | grep -qE '@(127\.0\.0\.1|localhost|\[::1\])[:/]'; then
    mkdir -p "$(dirname "$dropin")"
    cat > "$dropin" <<'EOF'
# Base PostgreSQL locale : ne pas démarrer avant elle.
[Unit]
Requires=postgresql.service
After=postgresql.service
EOF
  else
    rm -f "$dropin"
  fi
  systemctl daemon-reload
}

health_url() { printf 'https://127.0.0.1:%s/healthz' "$HTTPS_PORT"; }

# wait_health [SECONDES]
wait_health() {
  local deadline="${1:-40}"
  local i
  for i in $(seq "$deadline"); do
    : "$i"
    if [ "$(curl -sk -o /dev/null -w '%{http_code}' "$(health_url)" 2>/dev/null)" = 200 ]; then
      return 0
    fi
    sleep 1
  done
  return 1
}
