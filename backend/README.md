# StrongSwan Manager — Backend (tranche verticale)

> 📘 **Vous cherchez à *utiliser* le produit ?** La documentation utilisateur est en ligne :
> **https://billfassinou.github.io/strongswan-manager/docs/** — installation, prise en main et
> tous les usages courants, pas à pas. Sa source est dans [`../site/docs/`](../site/docs/).
> Ce fichier-ci s'adresse aux **développeurs du backend**.

Backend Go du produit *StrongSwan Manager*. Cet incrément est un **walking skeleton** :
un chemin bout-en-bout étroit mais réel (Auth → REST/WebSocket → cœur Go → VICI → charon)
qui valide toute l'architecture avant élargissement. Stack conforme au cahier des charges
du projet (§5–§7, §10, §14).

## Stack

Go · chi (HTTP) · pgx/PostgreSQL · JWT + RBAC · **govici** (VICI, voie primaire) ·
coder/websocket (temps réel) · Prometheus (`/metrics`) · OpenAPI (`/api/v1/docs`) ·
**front React + TypeScript (Vite)** embarqué dans le binaire et servi par le backend.

## Front React

La SPA (`web/`, React + TS + Vite) est **buildée puis embarquée** dans le binaire
(`web/embed.go`, `//go:embed all:dist`) et servie par le backend à la **même origine** que
l'API — pas de CORS, le JWT et le WebSocket fonctionnent directement.

```bash
make web        # npm install + build (produit web/dist, requis avant un build go local)
```

L'image Docker builde le front dans un stage `node` — `make run` / `make lab-up` sont donc
autonomes. En développement du front seul : `cd web && npm run dev` (proxifie `/api` et le
WebSocket vers `localhost:7926`). Le backend sert `/` (l'app), `/assets/*` et fait un **repli SPA** sur toute route client
inconnue. Navigation groupée conforme à la maquette de référence, thème clair/sombre, RBAC
(masquage des actions selon `/me.can_write`).

**Écrans branchés à l'API réelle** : Login (JWT) · Tableau de bord (tunnels + passerelles,
**temps réel WebSocket**) · Connexions (liste, initiate/terminate, suppression) · **Éditeur
de tunnel** (formulaire guidé PSK/certificat + pair S2S, **score de sécurité recalculé en
direct**, création/édition) · **PKI & Certificats** (CA, génération, révocation, CRL/CDP) ·
**Secrets** (CRUD, valeurs masquées) · **Sécurité & Conformité** (score du parc, algorithmes
faibles → « Corriger ») · **Passerelles** · **Journal d'audit** (rafraîchi) · **Administration**
(session, rôles, liens OpenAPI/metrics).

**Tous les modules de la maquette sont désormais fonctionnels** : Topologie (graphe calculé
depuis `/gateways`+`/tunnels`), Monitoring (règles d'alerte), Utilisateurs VPN, **Pools,
RADIUS, Politiques, Autorités** (CRUD persistés), Paramètres du démon (singleton), Assistant
IA (diagnostic à base de règles sur les tunnels réels + anomalies dérivées de l'état).

Ces modules de configuration reposent sur un **CRUD générique** : côté backend une table
`config_items` (`kind` + JSONB) exposée sur `GET/POST /api/v1/config/{kind}` et
`PUT/DELETE /api/v1/config/{kind}/{id}` (RBAC + audit) ; côté front un composant `Crud`
piloté par des schémas (`web/src/schemas.ts`). `kind ∈ {pool, radius, policy, authority,
vpnuser, alert, daemon}`.

## Démarrage rapide (sans lab : adaptateur VICI mock)

```bash
make run        # postgres + backend ; API sur http://localhost:7926
```

Le backend applique les migrations, seede 4 comptes (`admin`/`operator`/`auditor`/`viewer`,
mot de passe `admin1234`) et enregistre une passerelle `gw-local` adossée à un adaptateur
VICI **mock** — l'API est donc pleinement exerçable sans strongSwan.

## Démarrage avec lab VICI réel

```bash
make lab-up     # postgres + backend + 2 conteneurs strongSwan (plugin vici)
```

`make lab-up` exporte automatiquement `VICI_ENDPOINTS=gw-a=unix:/gw/a/charon.vici,gw-b=…`
pour brancher le backend sur les sockets VICI partagés via volumes. Vérifié de bout en bout :
le backend détecte la version (`version`), charge une connexion (`load-conn`) et la retire
(`unload-conn`) sur un **vrai charon** — visible avec :

```bash
docker compose exec strongswan-a swanctl --list-conns --uri unix:///vicirun/charon.vici
```

**Établissement bout-en-bout gw-a↔gw-b** : un tunnel site-à-site avec `peer_gateway_id`
configure les **deux** passerelles depuis la console (connexion + PSK sur chaque côté,
extrémités inversées côté pair). En passant les IP réelles des conteneurs comme adresses
IKE, `POST /api/v1/tunnels/{id}/initiate` établit une vraie SA — vérifié :
`swanctl --list-sas` affiche `hub: ESTABLISHED, IKEv2` et `hub-net: INSTALLED, TUNNEL,
ESP:AES_GCM_16-256`, et l'API remonte `status: up`.

**Établissement par certificats (PKI interne)** : une CA ECDSA est générée au démarrage.
On émet un certificat par passerelle (`POST /api/v1/certificates`, SAN = IP), puis un
tunnel `auth.method=cert` avec `cert_ref` (+ `peer_cert_ref` pour le pair) : le backend
charge la CA, le certificat et la clé sur chaque passerelle (`load-cert`/`load-key`) et
établit la SA. Vérifié : `certhub: ESTABLISHED, IKEv2` (ECDH ecp256), `ESP:AES_GCM_16-256`.

Points d'attention du lab :
- **Version strongSwan** : Debian 12 fournit **5.9.8** (≥ socle 5.9, mais pas 6.0). Pour
  disposer du post-quantique (ML-KEM, 6.0+), construire strongSwan depuis les sources ou
  partir d'une image 6.x — étape ultérieure.
- **Plugins crypto** : l'image installe `libstrongswan-standard-plugins` (openssl) pour
  l'ECDSA/ECDH, indispensable à l'auth par certificats et aux propositions `ecp*`.
- **Certificats via VICI** : `load-cert`/`load-key` reçoivent le **DER** (converti depuis
  le PEM dans l'adaptateur govici) — charon rejette le PEM brut sur cette version.
- **Révocation (CRL)** : strongSwan n'expose **pas** de commande VICI pour charger une CRL ;
  la révocation passe par le **CRL Distribution Point** inscrit dans les certificats
  (`CRL_URL`, ex. `http://backend:7926/crl.der`), récupéré par le plugin `curl` de charon.
  Vérifié : la passerelle fetch `/crl.der` (HTTP 200) et le certificat révoqué y figure.
  L'application automatique par charon dépend ensuite de sa politique de cache/revalidation
  (`remote.revocation`, durée de vie de la CRL via `CRL_VALIDITY`).
- **Backend en root dans le lab** (`user: "0:0"` dans le compose) pour accéder au socket
  VICI (`root:root` 0770). En production, cet accès passe par l'**agent distant mTLS**, pas
  par un socket partagé.
- **Établissement de SA** entre gw-a et gw-b (miroir de connexion sur le pair + `load-shared`
  du PSK des deux côtés + `initiate`) est l'incrément suivant ; à ce stade, la preuve porte
  sur `load-conn`/`unload-conn`/`version`/`list-sas` contre le vrai démon.

## Endpoints principaux (contrat §10)

| Méthode | Chemin | Rôle |
|---|---|---|
| POST | `/api/v1/auth/login` | Auth → JWT |
| GET | `/api/v1/me` | Profil |
| GET | `/api/v1/gateways` | Passerelles |
| GET/POST | `/api/v1/secrets` · DELETE `/{id}` | Coffre de secrets (chiffrés au repos, valeurs jamais renvoyées) |
| GET | `/api/v1/ca` · GET/POST `/api/v1/certificates` · POST `/{id}/revoke` | PKI interne : CA, émission/révocation de certificats X.509 (clés jamais renvoyées) |
| GET | `/api/v1/crl` · POST `/api/v1/crl/publish` · GET `/crl.der` (public) | CRL : liste de révocation signée ; `/crl.der` est le point de distribution (CDP) |
| GET/POST | `/api/v1/tunnels` | Liste / création |
| GET/PUT/DELETE | `/api/v1/tunnels/{id}` | Détail / màj / suppression |
| POST | `/api/v1/tunnels/{id}/{initiate\|terminate\|rekey}` | Actions VICI |
| GET | `/api/v1/tunnels/{id}/versions` · POST `/rollback` | Versioning |
| GET | `/api/v1/audit` | Journal immuable |
| GET | `/api/v1/ws?token=…` | Flux temps réel |
| GET | `/metrics` · `/api/v1/docs` | Prometheus · Swagger UI |

## Exemple

```bash
TOKEN=$(curl -s localhost:7926/api/v1/auth/login -d '{"identity":"admin","password":"admin1234"}' | jq -r .token)
GW=$(curl -s localhost:7926/api/v1/gateways -H "Authorization: Bearer $TOKEN" | jq -r .items[0].id)
curl -s localhost:7926/api/v1/tunnels -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{
  "name":"paris-dakar","gateway_id":"'"$GW"'","type":"site-to-site","ike_version":2,
  "local":{"addr":"203.0.113.10","subnets":["10.1.0.0/16"]},
  "remote":{"addr":"198.51.100.20","subnets":["10.2.0.0/16"]},
  "auth":{"method":"psk"},
  "proposals":{"ike":["aes256-sha256-modp2048"],"esp":["aes256gcm16"]},"pfs":true
}' | jq
```

## Tests

```bash
make test              # tests unitaires (aucune dépendance externe)
make cover             # + rapport de couverture par paquet
make test-integration  # repositories contre un postgres jetable (go local + docker)
```

Couverture unitaire par paquet (repère) : `config` 100 %, `auth` ~98 %, `ws` ~94 %,
`domain` ~93 %, `metrics` ~92 %, `httpapi` ~72 %, `poller` ~67 %, `vici` ~42 %.

- La couche HTTP est testée de bout en bout (routeur chi + `httptest`) avec des **fakes de
  repositories en mémoire** et l'**adaptateur VICI mock** — pas de PostgreSQL requis :
  auth (200/401/403), RBAC (viewer/auditeur → 403), création `201`, validation `422`
  (format §10), versions, rollback, `initiate`, suppression + `unload` VICI, audit,
  `/metrics`, OpenAPI.
- Le paquet `store` (SQL réel, contraintes d'unicité, **immuabilité du journal d'audit**
  via trigger, chaînage d'intégrité) est couvert par des **tests d'intégration** taggés
  `integration`.
- L'**adaptateur govici réel** (`Version`/`LoadConn`/`ListSAs`/…) nécessite un socket
  charon : ses helpers purs sont testés unitairement, et le chemin complet est validé par
  le **lab** (`make lab-up`, cf. ci-dessus). Le reste (mapping `load-conn`, mock, parsing
  d'endpoint) est couvert unitairement.

## Périmètre & suites

Hors périmètre de cet incrément (increments suivants) : PKI/step-ca, Vault, agent distant
mTLS (le lab partage le socket VICI), pools/RADIUS/politiques/paramètres démon, IA,
multi-tenant/SSO, TimescaleDB, Helm. Voir la roadmap §12 du cahier des charges.
