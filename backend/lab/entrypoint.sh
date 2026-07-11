#!/bin/sh
# Démarre charon (mode swanctl/vici) en avant-plan, en localisant le binaire selon
# la distribution. Le socket VICI est configuré dans /etc/strongswan.conf.
set -e

mkdir -p /vicirun

for c in /usr/sbin/charon-systemd /usr/lib/ipsec/charon-systemd /usr/libexec/ipsec/charon-systemd; do
    if [ -x "$c" ]; then
        echo "[entrypoint] démarrage $c"
        exec "$c"
    fi
done

for c in /usr/lib/ipsec/charon /usr/libexec/ipsec/charon; do
    if [ -x "$c" ]; then
        echo "[entrypoint] démarrage $c"
        exec "$c"
    fi
done

echo "[entrypoint] charon introuvable — contenu des répertoires ipsec :"
ls -la /usr/lib/ipsec /usr/libexec/ipsec /usr/sbin 2>/dev/null | grep -i charon || true
exit 1
