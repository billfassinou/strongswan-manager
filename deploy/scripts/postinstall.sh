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

  # --- Vérifications réelles : on prouve, on ne suppose pas ---
  TROUBLES=0
  VICI_BROKEN=0

  if wait_health 40; then
    ok "service actif et console joignable (HTTPS :$HTTPS_PORT)"
  else
    TROUBLES=1
    warn "le service ne répond pas. Diagnostic : swanmgrctl doctor"
  fi

  if verify_db; then
    ok "base de données joignable"
  elif [ $? -eq 2 ]; then
    warn "connexion à la base non vérifiable ici (client psql absent)."
  else
    TROUBLES=1
    warn "BASE DE DONNÉES INJOIGNABLE avec le DSN de $ENV_FILE."
  fi

  # strongSwan installé sur cette machine ? Alors il DOIT être pilotable par le service.
  if command -v swanctl >/dev/null 2>&1; then
    if verify_vici; then
      ok "strongSwan piloté par VICI (testé sous l'identité « $SVC_USER »)"
    else
      TROUBLES=1; VICI_BROKEN=1
      warn "STRONGSWAN EST INSTALLÉ MAIS INJOIGNABLE PAR LA CONSOLE (utilisateur « $SVC_USER »)."
      warn "  Laissée ainsi, la console tournerait en MODE DÉMO : tunnels SIMULÉS, aucun trafic"
      warn "  réellement chiffré. Vérifiez : systemctl status $(strongswan_unit) ;"
      warn "  stat -c '%G %a' /run/charon.vici  (attendu : « $SVC_USER 660 »), puis swanmgrctl doctor."
    fi
  else
    TROUBLES=1
    warn "strongSwan n'est PAS installé : la console démarre en MODE DÉMO (tunnels simulés)."
    warn "  Installez-le, ou renseignez VICI_ENDPOINTS dans $ENV_FILE pour des passerelles distantes."
  fi

  IP="$(host_ip)"; IP="${IP:-127.0.0.1}"
  echo
  if [ "$TROUBLES" -eq 0 ]; then
    echo "  ✔ StrongSwan Manager est installé — tout est opérationnel."
  else
    echo "  ! StrongSwan Manager est installé, MAIS des points demandent votre attention (ci-dessus)."
  fi
  cat <<EOF

    Console      https://$IP:$HTTPS_PORT
    Identifiant  admin
    Mot de passe $(env_get SEED_ADMIN_PASSWORD)
                 (la console vous demandera de le changer à la première connexion)

    Diagnostic   swanmgrctl doctor
    Sauvegarde   swanmgrctl backup

  ⚠️ Sauvegardez $ENV_FILE : il contient SECRETS_KEY, sans laquelle vos secrets et clés
     privées seraient DÉFINITIVEMENT illisibles, même avec une sauvegarde de la base.

EOF

  # strongSwan présent mais injoignable : on SORT EN ERREUR. Le paquet reste posé et le
  # service configuré (rien à réinstaller), mais l'installation est signalée comme ayant
  # échoué — plutôt que de laisser croire à une console qui pilote de vrais tunnels alors
  # qu'elle les simule. Corrigez VICI, puis « swanmgrctl doctor ».
  if [ "$VICI_BROKEN" -eq 1 ]; then
    exit 1
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
