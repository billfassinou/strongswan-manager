# API REST & WebSocket

Tout ce que fait l'interface est disponible par l'API. C'est la voie pour **automatiser** : pipelines CI/CD, scripts, intégration dans votre outillage.

- Base : `/api/v1`
- Format : JSON
- Documentation interactive : **`/api/v1/docs`** (Swagger UI) · spécification brute : **`/api/v1/openapi.yaml`**

---

## S'authentifier

```bash
TOKEN=$(curl -sk https://localhost:7926/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identity":"admin","password":"admin1234"}' | jq -r .token)
```

Puis, sur chaque appel :

```bash
curl -sk https://localhost:7926/api/v1/me -H "Authorization: Bearer $TOKEN"
```

Le jeton expire au bout de `JWT_TTL` (1 h par défaut).

---

## Droits

| Réponse | Signification |
|---|---|
| **401** | Jeton absent, invalide ou expiré |
| **403** | Rôle en lecture seule (`auditor`, `viewer`) sur une route modifiante |
| **422** | Validation métier échouée (voir plus bas) |

Voir [Rôles & permissions](A1-roles-et-permissions.md).

---

## Toutes les routes

### Publiques (sans authentification)

| Méthode | Chemin | Description |
|---|---|---|
| GET | `/healthz` | Sonde de vie |
| GET | `/metrics` | Métriques Prometheus |
| GET | `/crl.der` | **CRL en DER** — le point de distribution lu par les passerelles |
| POST | `/api/v1/auth/login` | Authentification → jeton JWT |
| GET | `/api/v1/openapi.yaml` | Spécification OpenAPI |
| GET | `/api/v1/docs` | Documentation interactive (Swagger UI) |
| GET | `/api/v1/ws` | Flux temps réel (WebSocket), jeton en `?token=` |

### Lecture (jeton requis — tous les rôles)

| Méthode | Chemin | Description |
|---|---|---|
| GET | `/api/v1/me` | Profil : identité, rôle, `can_write` |
| GET | `/api/v1/gateways` | Passerelles |
| GET | `/api/v1/tunnels` | Tunnels (état live inclus) |
| GET | `/api/v1/tunnels/{id}` | Détail d'un tunnel |
| GET | `/api/v1/tunnels/{id}/versions` | Historique des configurations |
| GET | `/api/v1/secrets` | Secrets (**valeurs masquées**) |
| GET | `/api/v1/certificates` | Certificats (**sans clé privée**) |
| GET | `/api/v1/ca` | Autorité de certification (PEM public) |
| GET | `/api/v1/crl` | CRL au format PEM |
| GET | `/api/v1/config/{kind}` | Items d'un module de configuration |
| GET | `/api/v1/audit` | Journal d'audit (`?limit=`) |

### Écriture (jeton requis — `admin` ou `operator` ; sinon **403**)

| Méthode | Chemin | Description |
|---|---|---|
| POST | `/api/v1/tunnels` | Créer un tunnel (valide, score, applique via VICI) |
| PUT | `/api/v1/tunnels/{id}` | Modifier |
| DELETE | `/api/v1/tunnels/{id}` | Supprimer (décharge la connexion) |
| POST | `/api/v1/tunnels/{id}/initiate` | Monter |
| POST | `/api/v1/tunnels/{id}/terminate` | Couper |
| POST | `/api/v1/tunnels/{id}/rekey` | Renégocier |
| POST | `/api/v1/tunnels/{id}/rollback` | Restaurer une version |
| POST | `/api/v1/secrets` | Créer un secret (chiffré au repos) |
| DELETE | `/api/v1/secrets/{id}` | Supprimer un secret |
| POST | `/api/v1/certificates` | Émettre un certificat |
| POST | `/api/v1/certificates/{id}/revoke` | Révoquer (régénère la CRL) |
| POST | `/api/v1/crl/publish` | Régénérer la CRL |
| POST | `/api/v1/config/{kind}` | Créer un item de configuration |
| PUT | `/api/v1/config/{kind}/{id}` | Modifier |
| DELETE | `/api/v1/config/{kind}/{id}` | Supprimer |

`{kind}` ∈ `pool`, `radius`, `policy`, `authority`, `vpnuser`, `alert`, `daemon`. Un `kind` inconnu renvoie **404** `unknown_kind`.

---

## Créer un tunnel

```bash
curl -sk https://localhost:7926/api/v1/tunnels \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "paris-dakar",
    "gateway_id": "<UUID de la passerelle>",
    "type": "site-to-site",
    "ike_version": 2,
    "local":  { "addr": "203.0.113.10",  "subnets": ["10.1.0.0/16"] },
    "remote": { "addr": "198.51.100.20", "subnets": ["10.2.0.0/16"] },
    "auth":   { "method": "psk", "secret_ref": "psk-dakar" },
    "proposals": { "ike": ["aes256-sha256-modp2048"], "esp": ["aes256gcm16"] },
    "pfs": true
  }'
```

Réponse `201` :

```json
{
  "id": "…",
  "name": "paris-dakar",
  "status": "installing",
  "security_score": 94,
  "warnings": ["Pas de préparation post-quantique (ML-KEM)"],
  "config_version": 1
}
```

Champs optionnels utiles :

| Champ | Effet |
|---|---|
| `peer_gateway_id` | Configure aussi la **passerelle pair** (connexion miroir) |
| `peer_cert_ref` | Certificat du pair (authentification par certificat des deux côtés) |
| `auth.cert_ref` | Certificat local (avec `"method": "cert"`) |

---

## Format d'erreur

Toutes les erreurs suivent la même structure :

```json
{
  "error": "validation_failed",
  "message": "Proposition cryptographique faible détectée",
  "details": [
    { "field": "proposals.ike", "issue": "modp1024 déconseillé (DH groupe 2)" }
  ],
  "correlation_id": "…"
}
```

`correlation_id` permet de retrouver la requête dans les logs du serveur.

---

## Les modules de configuration

Un seul schéma pour tous :

```bash
# créer un pool
curl -sk https://localhost:7926/api/v1/config/pool \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"pool-rw","data":{"range":"10.9.0.0/24","source":"Interne","dns":"10.1.0.53"}}'

# lister
curl -sk https://localhost:7926/api/v1/config/pool -H "Authorization: Bearer $TOKEN"

# modifier / supprimer
curl -sk -X PUT    https://localhost:7926/api/v1/config/pool/<ID> -H "Authorization: Bearer $TOKEN" \
     -H 'Content-Type: application/json' -d '{"name":"pool-rw","data":{"range":"10.9.0.0/22"}}'
curl -sk -X DELETE https://localhost:7926/api/v1/config/pool/<ID> -H "Authorization: Bearer $TOKEN"
```

Le champ `data` est libre (JSON) : chaque module y met ses propres champs.

---

## Le flux temps réel (WebSocket)

```
ws://localhost:7926/api/v1/ws?token=<jeton>
```

Chaque changement d'état émet un message :

```json
{ "type": "tunnel_status", "id": "…", "name": "paris-dakar", "status": "up" }
```

> **Limitation** : le jeton n'est vérifié **que s'il est fourni**. Une connexion sans `?token=` est actuellement acceptée. Voir [Rôles & permissions](A1-roles-et-permissions.md).

---

## Exemple complet : un tunnel de bout en bout

```bash
B=https://localhost:7926
TOKEN=$(curl -sk $B/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"identity":"admin","password":"admin1234"}' | jq -r .token)
GW=$(curl -sk $B/api/v1/gateways -H "Authorization: Bearer $TOKEN" | jq -r '.items[0].id')

# secret
curl -sk $B/api/v1/secrets -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"psk-demo","type":"psk","value":"clef-partagee"}' >/dev/null

# tunnel
TID=$(curl -sk $B/api/v1/tunnels -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{
  "name":"demo","gateway_id":"'"$GW"'","type":"site-to-site","ike_version":2,
  "local":{"addr":"203.0.113.10","subnets":["10.1.0.0/16"]},
  "remote":{"addr":"198.51.100.20","subnets":["10.2.0.0/16"]},
  "auth":{"method":"psk","secret_ref":"psk-demo"},
  "proposals":{"ike":["aes256-sha256-modp2048"],"esp":["aes256gcm16"]},"pfs":true}' | jq -r .id)

# monter, puis lire l'état
curl -sk -X POST $B/api/v1/tunnels/$TID/initiate -H "Authorization: Bearer $TOKEN"
sleep 4
curl -sk $B/api/v1/tunnels/$TID -H "Authorization: Bearer $TOKEN" | jq '{name, status, security_score}'
```
