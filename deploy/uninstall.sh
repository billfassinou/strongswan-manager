#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Désinstalle StrongSwan Manager.
#
# Par défaut, ce script CONSERVE la base de données et la configuration : elles contiennent
# vos tunnels, votre PKI, votre audit — et SECRETS_KEY, sans laquelle la base est illisible.
# Utilisez --purge pour tout supprimer (irréversible).

set -euo pipefail

for c in "$(dirname "${BASH_SOURCE[0]}")/lib/common.sh" \
         /usr/share/strongswan-manager/lib/common.sh \
         /usr/local/share/strongswan-manager/lib/common.sh; do
  # shellcheck source=lib/common.sh
  [ -f "$c" ] && { . "$c"; break; }
done
command -v die >/dev/null 2>&1 || { echo "lib/common.sh introuvable" >&2; exit 1; }

PURGE=0
[ "${1:-}" = "--purge" ] && PURGE=1
need_root

if [ "$PURGE" -eq 1 ]; then
  cat <<EOF

  ⚠️  --purge SUPPRIME DÉFINITIVEMENT :
      - la base « $DB_NAME » (tunnels, PKI, secrets chiffrés, journal d'audit) ;
      - $ETC_DIR (dont SECRETS_KEY : aucune sauvegarde de la base ne sera plus déchiffrable).

  Pour en garder une copie avant de tout effacer : swanmgrctl backup

EOF
  read -rp "  Taper « supprimer » pour confirmer : " a
  [ "$a" = "supprimer" ] || die "annulé."
fi

info "arrêt du service"
systemctl disable --now "$SVC_NAME" >/dev/null 2>&1 || true
rm -f "$UNIT"
rm -rf "/etc/systemd/system/$SVC_NAME.service.d"

# Le drop-in posé sur strongswan.service n'a plus lieu d'être.
remove_vici_dropin

rm -f "$BIN" "$BIN_DIR/swanmgrctl"
ok "service et binaires supprimés"

if [ "$PURGE" -eq 1 ]; then
  su - postgres -c "dropdb --if-exists $DB_NAME" >/dev/null 2>&1 || true
  su - postgres -c "dropuser --if-exists $DB_USER" >/dev/null 2>&1 || true
  rm -rf "$ETC_DIR" "$STATE_DIR" "$SHARE_DIR" /usr/local/share/strongswan-manager
  userdel "$SVC_USER" 2>/dev/null || true
  ok "base, configuration et utilisateur supprimés"
else
  # $SHARE_DIR (ce script + lib/) est CONSERVÉ : c'est le seul
  # moyen de lancer « --purge » plus tard. Le purger ici reviendrait à retirer l'escabeau.
  cat <<EOF

  Conservés (réutilisables lors d'une réinstallation) :
    - la base PostgreSQL « $DB_NAME »
    - $ETC_DIR (dont SECRETS_KEY)
    - l'utilisateur système « $SVC_USER »

  Pour tout supprimer : $SHARE_DIR/uninstall.sh --purge
  PostgreSQL et strongSwan ne sont PAS désinstallés (d'autres services peuvent en dépendre).

EOF
fi
