#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Après retrait du paquet.
#
# Comme preremove.sh, ce scriptlet est aussi appelé pendant une MISE À JOUR : retirer le
# drop-in VICI à ce moment-là couperait l'accès de la console à charon jusqu'au prochain
# redémarrage de strongSwan. On ne nettoie donc que sur un vrai retrait.
#
#   dpkg : $1 = remove | purge | upgrade | …
#   rpm  : $1 = 0 (désinstallation) | 1 (mise à jour)
#
# On ne supprime NI la base, NI /etc/strongswan-manager : ce dernier contient SECRETS_KEY,
# sans laquelle la base — et toute sauvegarde de celle-ci — devient définitivement illisible.
# Une désinstallation ne doit pas pouvoir détruire des données par surprise.

set -e

case "${1:-}" in
  remove|purge|0) ;;
  *) exit 0 ;;
esac

# Le drop-in posé sur strongswan.service n'a plus lieu d'être : charon n'a plus à ouvrir son
# socket au groupe swanmgr.
rm -f /etc/systemd/system/strongswan.service.d/10-vici-swanmgr.conf \
      /etc/systemd/system/strongswan-swanctl.service.d/10-vici-swanmgr.conf
rmdir /etc/systemd/system/strongswan.service.d \
      /etc/systemd/system/strongswan-swanctl.service.d 2>/dev/null || true
rm -rf /etc/systemd/system/strongswan-manager.service.d

systemctl daemon-reload >/dev/null 2>&1 || true
systemctl try-restart strongswan >/dev/null 2>&1 || true

cat <<'EOF'

  StrongSwan Manager retiré. CONSERVÉS (une réinstallation les retrouvera tels quels) :
    - la base PostgreSQL « swan » (tunnels, PKI, secrets, audit)
    - /etc/strongswan-manager/ (dont SECRETS_KEY)

  Pour tout effacer — IRRÉVERSIBLE, aucune sauvegarde de la base ne sera plus déchiffrable :
    su - postgres -c 'dropdb --if-exists swan; dropuser --if-exists swan'
    rm -rf /etc/strongswan-manager /var/lib/strongswan-manager
    userdel swanmgr

EOF
