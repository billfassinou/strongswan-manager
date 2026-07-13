#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Scriptlet post-installation des paquets .deb / .rpm.
#
# Il est rejoué à CHAQUE mise à jour : tout doit donc être idempotent, et surtout il ne doit
# JAMAIS réécrire le fichier de configuration — SECRETS_KEY y est unique et irremplaçable.

set -e

# shellcheck source=../lib/common.sh
. /usr/share/strongswan-manager/lib/common.sh

FIRST_INSTALL=0
[ -f "$ENV_FILE" ] || FIRST_INSTALL=1

ensure_user
migrate_from_usr_local
install -d -m 0750 -o "$SVC_USER" -g "$SVC_USER" "$STATE_DIR"

# Base de données : uniquement si PostgreSQL tourne ici. Avec une base distante,
# l'administrateur renseigne DATABASE_URL lui-même dans le fichier de configuration.
DB_PASS=""
if pg_local; then
  pg_start_and_wait
  DB_PASS="$(provision_db)"
else
  [ "$FIRST_INSTALL" -eq 1 ] && warn "aucun PostgreSQL local détecté : renseignez DATABASE_URL dans $ENV_FILE."
fi

write_env_file "$DB_PASS" 0

# Socket VICI : seulement si strongSwan est installé sur cette machine.
if command -v swanctl >/dev/null 2>&1; then
  install_vici_dropin || true
  if [ "$FIRST_INSTALL" -eq 1 ]; then
    sed -i 's|^VICI_ENDPOINTS=.*|VICI_ENDPOINTS=local=unix:/var/run/charon.vici|' "$ENV_FILE"
  fi
fi

apply_db_dependency
systemctl daemon-reload

if [ "$FIRST_INSTALL" -eq 1 ]; then
  systemctl enable --now "$SVC_NAME" >/dev/null 2>&1 || true
  open_firewall

  if wait_health 40; then
    IP="$(host_ip)"; IP="${IP:-127.0.0.1}"
    cat <<EOF

  ✔ StrongSwan Manager est installé.

    Console      https://$IP:$HTTPS_PORT
    Identifiant  admin
    Mot de passe $(env_get SEED_ADMIN_PASSWORD)
                 (la console vous demandera de le changer à la première connexion)

    Diagnostic   swanmgrctl doctor
    Sauvegarde   swanmgrctl backup

  ⚠️ Sauvegardez $ENV_FILE : il contient SECRETS_KEY, sans laquelle vos secrets et clés
     privées seraient DÉFINITIVEMENT illisibles, même avec une sauvegarde de la base.

EOF
  else
    warn "le service ne répond pas encore. Diagnostic : swanmgrctl doctor"
  fi
else
  # Mise à jour : on redémarre sur le nouveau binaire. « enable » est réaffirmé (et non un
  # simple try-restart) car dpkg a pu désactiver l'unité en route ; les migrations de schéma
  # sont appliquées par l'application elle-même au démarrage.
  systemctl enable "$SVC_NAME" >/dev/null 2>&1 || true
  systemctl restart "$SVC_NAME" >/dev/null 2>&1 || true
  if wait_health 40; then
    ok "StrongSwan Manager mis à jour (configuration et base conservées)."
  else
    warn "le service ne répond pas après mise à jour. Diagnostic : swanmgrctl doctor"
  fi
fi
