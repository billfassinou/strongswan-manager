#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Avant retrait du paquet.
#
# ATTENTION : ce scriptlet est aussi appelé lors d'une MISE À JOUR (dpkg passe « upgrade »,
# rpm passe « 1 »). Y désactiver le service inconditionnellement le laisserait arrêté ET
# désactivé après chaque « apt upgrade » — la console ne redémarrerait plus au boot.
# On ne touche donc au service que s'il s'agit d'un vrai retrait.
#
#   dpkg : $1 = remove | upgrade | deconfigure
#   rpm  : $1 = 0 (désinstallation) | 1 (mise à jour)

set -e

case "${1:-}" in
  remove|purge|0)
    systemctl disable --now strongswan-manager >/dev/null 2>&1 || true
    ;;
  *)
    # Mise à jour : le service reste activé ; postinstall le redémarrera sur le nouveau binaire.
    ;;
esac
