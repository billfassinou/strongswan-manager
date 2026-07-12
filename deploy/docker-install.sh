#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Bill Fassinou
#
# Installe StrongSwan Manager avec Docker Compose, en production :
# génère un .env aux secrets aléatoires, démarre la pile, attend que la console réponde.
#
#   ./docker-install.sh [--tag vX.Y.Z]
#
# À la différence de backend/docker-compose.yml (le lab), rien n'est laissé à une valeur
# par défaut publique : le serveur refuserait d'ailleurs de démarrer.

set -euo pipefail

cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

COMPOSE="docker-compose.prod.yml"
ENV=".env"
TAG="${TAG:-latest}"

info() { printf '\033[1;34m▸\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m✔\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m!\033[0m %s\n' "$*"; }
die()  { printf '\033[1;31m✘ %s\033[0m\n' "$*" >&2; exit 1; }

while [ $# -gt 0 ]; do
  case "$1" in
    --tag) TAG="$2"; shift 2 ;;
    --help|-h) sed -n '5,12p' "$0"; exit 0 ;;
    *) die "option inconnue : $1" ;;
  esac
done

command -v docker >/dev/null 2>&1 || die "Docker est requis."
docker compose version >/dev/null 2>&1 || die "Docker Compose v2 est requis (« docker compose »)."
[ -f "$COMPOSE" ] || die "$COMPOSE introuvable (lancez ce script depuis le dossier qui le contient)."

if [ -f "$ENV" ]; then
  # SECRETS_KEY chiffre les secrets, la CA et la clé TLS déjà en base : la régénérer les
  # rendrait illisibles et le serveur refuserait de démarrer. On n'y touche jamais.
  ok "$ENV existant conservé (secrets inchangés)"
else
  rand() { openssl rand -hex 32 2>/dev/null || head -c32 /dev/urandom | od -An -tx1 | tr -d ' \n'; }
  ADMIN_PASS="$(openssl rand -base64 18 | tr -d '/+=' | cut -c1-16)"
  HOST="$(hostname -f 2>/dev/null || hostname)"
  IP="$(hostname -I 2>/dev/null | awk '{print $1}')"; IP="${IP:-127.0.0.1}"

  umask 077
  cat > "$ENV" <<EOF
# StrongSwan Manager — déploiement Docker. Généré le $(date -u +%Y-%m-%dT%H:%M:%SZ).
#
# ⚠️ SECRETS_KEY chiffre les secrets, les clés privées de la PKI et le certificat TLS.
#    NE LA CHANGEZ JAMAIS après le premier démarrage. Sauvegardez ce fichier avec le
#    volume PostgreSQL : l'un sans l'autre ne vaut rien.

TAG=$TAG
POSTGRES_PASSWORD=$(rand)
JWT_SECRET=$(rand)
SECRETS_KEY=$(rand)

# Mot de passe initial des 4 comptes. La console impose de le changer à la 1re connexion.
SEED_ADMIN_PASSWORD=$ADMIN_PASS

TLS_SANS=localhost,127.0.0.1,::1,$HOST,$IP
CRL_URL=http://$IP:7927/crl.der
CRL_VALIDITY=24h

# Passerelles pilotées. Vide = MODE DÉMO (tunnels simulés, aucun trafic réellement chiffré).
#   charon sur l'hôte : VICI_ENDPOINTS=local=unix:/var/run/charon.vici
#                       (+ décommentez le montage du socket dans $COMPOSE)
#   à distance        : VICI_ENDPOINTS=gw-paris=tcp:10.0.0.5:4502
VICI_ENDPOINTS=

POLL_INTERVAL=3s
CORS_ORIGINS=*

# Certificat reconnu par les navigateurs (domaine public + port 80 redirigé vers 7927) :
# ACME_DOMAIN=vpn.mondomaine.fr
# ACME_EMAIL=admin@mondomaine.fr
ACME_CACHE=/var/lib/strongswan-manager/acme
EOF
  umask 022
  chmod 0600 "$ENV"
  ok "$ENV généré (secrets aléatoires)"
fi

info "démarrage de la pile"
docker compose -f "$COMPOSE" up -d

info "attente de la console"
for _ in $(seq 60); do
  if [ "$(curl -sk -o /dev/null -w '%{http_code}' https://127.0.0.1:7926/healthz 2>/dev/null)" = 200 ]; then
    break
  fi
  sleep 1
done
[ "$(curl -sk -o /dev/null -w '%{http_code}' https://127.0.0.1:7926/healthz 2>/dev/null)" = 200 ] \
  || die "la console ne répond pas. Journaux : docker compose -f $COMPOSE logs backend"

IP="$(hostname -I 2>/dev/null | awk '{print $1}')"; IP="${IP:-127.0.0.1}"
PASS="$(sed -n 's/^SEED_ADMIN_PASSWORD=//p' "$ENV" | head -1)"

cat <<EOF

  ✔ StrongSwan Manager tourne.

    Console      https://$IP:7926
    Identifiant  admin
    Mot de passe $PASS   (à changer à la première connexion)

  Le navigateur affichera un avertissement : le certificat est auto-signé. La connexion est
  bien chiffrée ; c'est l'identité du serveur qui n'est pas attestée par un tiers.

    Journaux     docker compose -f $COMPOSE logs -f backend
    Arrêter      docker compose -f $COMPOSE stop
    Sauvegarder  docker compose -f $COMPOSE exec -T postgres pg_dump -U swan -Fc swan > swan.dump
                 …ET conservez $ENV avec : sans SECRETS_KEY, le dump est illisible.

EOF
