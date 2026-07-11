# FAQ

---

### L'outil écrit-il dans mes fichiers `swanctl.conf` ?

**Non.** Toute la configuration passe par l'**API VICI** du démon. Vos fichiers ne sont ni lus ni modifiés. Une connexion chargée par la console existe **dans le démon** (visible avec `swanctl --list-conns`), pas sur le disque.

---

### Puis-je l'utiliser en environnement isolé (air-gap) ?

Oui, avec des réserves :

- Aucune connexion Internet n'est requise au fonctionnement : les images, la base, la PKI et le front sont autonomes.
- L'autorité de certification est **interne** — pas besoin d'ACME / Let's Encrypt.
- La page de documentation d'API (`/api/v1/docs`) charge Swagger UI depuis un CDN : **elle ne s'affichera pas** hors ligne. La spécification brute (`/api/v1/openapi.yaml`) reste, elle, servie localement.

---

### Comment sauvegarder ?

Tout ce qui compte est dans **PostgreSQL** : tunnels, versions, secrets chiffrés, PKI, audit, comptes.

```bash
docker compose exec postgres pg_dump -U swan swan > sauvegarde.sql
```

⚠️ Conservez aussi **`SECRETS_KEY`** en lieu sûr : sans elle, les secrets et clés privées de la sauvegarde sont **indéchiffrables**.

Restauration :

```bash
cat sauvegarde.sql | docker compose exec -T postgres psql -U swan swan
```

---

### L'application est-elle servie en HTTPS ?

**Oui, par défaut, et sans rien configurer.** Au premier démarrage, elle émet son propre certificat (signé par sa CA interne) et sert en **HTTPS sur le port 7926**.

Le port **7927** reste en HTTP en clair, mais il ne sert **que** la CRL (`/crl.der`) et `/healthz` — tout le reste y est redirigé en 308 vers HTTPS.

Trois façons d'obtenir un certificat :

| Vous voulez… | Faites |
|---|---|
| Démarrer tout de suite | Rien. Certificat auto-généré. **Le navigateur avertira** tant que la CA interne n'est pas importée. |
| Aucun avertissement, domaine public | `ACME_DOMAIN=vpn.mondomaine.fr` → **Let's Encrypt**. Exige le **port 80 joignable**. |
| Utiliser votre PKI d'entreprise | `TLS_CERT` + `TLS_KEY` |

Voir [Variables d'environnement](A2-configuration.md).

---

### Pourquoi mon navigateur affiche un avertissement de sécurité ?

Parce que le certificat est **auto-généré**, signé par la CA interne de l'application — que votre navigateur ne connaît pas. **La connexion est bien chiffrée** ; c'est l'identité du serveur qui n'est pas attestée par un tiers.

C'est le comportement de tout équipement auto-hébergé (Proxmox, pfSense, TrueNAS). Deux façons de le supprimer :

- **Importer la CA interne** dans le magasin de confiance de vos postes d'administration (écran **PKI & Certificats**, ou `GET /api/v1/ca`). À faire une fois, vaut pour toutes vos instances.
- **Utiliser Let's Encrypt** (`ACME_DOMAIN`) si la console a un domaine public.

⚠️ N'ignorez l'avertissement que si vous **savez** que le certificat est le vôtre. Sur un réseau non maîtrisé, un avertissement peut aussi signaler une véritable interception.

---

### Je suis déjà derrière un reverse proxy TLS. Comment éviter le double TLS ?

`TLS_ENABLED=false`. L'application sert alors en HTTP en clair sur 7926, et c'est votre proxy (Nginx, Traefik, Caddy) qui termine le TLS. Pensez à **relayer les WebSockets**.

---

### Le front et l'API sont-ils séparés ?

Non : le front React est **embarqué dans le binaire** et servi **à la même origine** que l'API. C'est ce qui permet au jeton JWT et au WebSocket de fonctionner sans configuration CORS.

---

### Puis-je utiliser l'outil sans installer StrongSwan ?

Oui — c'est le **mode démo** (par défaut). Une passerelle simulée permet d'explorer toute l'interface. Voir [Installation](02-installation.md).

---

### Combien de passerelles / tunnels peut-il gérer ?

L'objectif de conception est **1 000 tunnels et 10 passerelles** avec une interface fluide. Le serveur interroge chaque passerelle toutes les `POLL_INTERVAL` (3 s par défaut) ; augmentez cette valeur si vous gérez un parc important.

---

### Où sont les mots de passe des comptes ?

Hachés en **bcrypt** dans la base. Ils ne sont jamais stockés en clair et ne peuvent pas être relus.

---

### Pourquoi je ne peux pas revoir la valeur d'un PSK ?

C'est volontaire. Un secret entré n'est **jamais** réaffiché — ni dans l'interface, ni par l'API, pour aucun rôle. Si vous l'avez perdu, créez-en un nouveau et mettez à jour les deux extrémités. Voir [Gérer les secrets](06-secrets.md).

---

### Comment automatiser (Terraform, CI/CD, scripts) ?

Tout ce que fait l'interface est disponible par l'**API REST**. Créez un jeton avec un compte dédié et appelez l'API. Voir [API REST & WebSocket](A3-api.md).

---

### Je veux contribuer / modifier le code. Par où commencer ?

```bash
cd backend
make web     # construit le front (obligatoire avant un build Go local)
make build   # compile
make test    # tests unitaires
make cover   # + couverture
make test-integration   # tests d'intégration (démarre un PostgreSQL jetable)
```

Les **tests unitaires** vivent à côté du code qu'ils testent ; les **tests d'intégration** sont dans `backend/test/`.

Prérequis : **Go** et **Node** (sinon, le `Makefile` retombe sur une image Docker `golang`).

---

### Qu'est-ce qui est prévu mais pas encore là ?

- Répondeur **OCSP**, enrôlement **SCEP/EST**
- **Agent distant mTLS** (remplacerait l'accès direct au socket VICI)
- **Multi-tenant** et **SSO** SAML/OIDC
- Envoi effectif des **notifications** (email, Slack, webhook)
- **Vault** à la place du chiffrement applicatif des secrets
- Écran d'**historique des versions** dans l'interface (l'API, elle, est complète)

---

### Une question qui n'est pas là ?

→ [Dépannage](15-depannage.md)
