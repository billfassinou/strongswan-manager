# Dépannage

---

## « Proposition cryptographique faible détectée » (erreur 422)

**Ce que ça veut dire** : la validation a **refusé** votre configuration. Elle n'a **pas** été appliquée sur la passerelle — vous ne risquez rien.

**Cause la plus fréquente** : `modp1024` dans les propositions IKE (groupe Diffie-Hellman obsolète).

**Correction** : remplacez par `modp2048`, ou mieux `ecp384`.

Autres refus possibles : nom manquant, version IKE invalide, CIDR mal écrit, propositions vides. Le message indique le **champ exact** en cause. Voir [Créer un tunnel](04-creer-un-tunnel.md).

---

## « application VICI échouée » (erreur 502 `vici_error`)

**Ce que ça veut dire** : la configuration était valide, mais le démon `charon` a **refusé** de la charger. Le message reprend l'erreur renvoyée par le démon.

| Message de charon | Cause | Correction |
|---|---|---|
| `parsing X509 certificate failed` | La passerelle ne sait pas lire le certificat — souvent le **plugin `openssl` manquant** (pas d'ECDSA) | Installez `libstrongswan-standard-plugins` sur la passerelle |
| `invalid certificate type 'crl'` | Tentative de pousser une CRL par VICI — **impossible** | La révocation passe par le CDP, voir [PKI](07-pki-certificats.md) |
| erreur sur une proposition | Le démon ne connaît pas un algorithme (ex. `mlkem768` sur StrongSwan 5.9) | Vérifiez la **version** dans **Passerelles** et retirez l'algorithme non supporté |

Quand l'application échoue à la création, l'enregistrement est **annulé** : la base et la passerelle restent cohérentes.

---

## Le tunnel reste `down` après « Monter »

Passez en revue, dans l'ordre :

1. **Le pair est-il joignable ?** IKE utilise **UDP 500** et **UDP 4500**. Un pare-feu qui les bloque suffit.
2. **Les propositions sont-elles symétriques ?** Les deux extrémités doivent partager au moins une suite commune (IKE **et** ESP).
3. **Le secret ou le certificat correspondent-ils ?**
   - PSK : la **même valeur** des deux côtés.
   - Certificat : le **SAN** doit correspondre à l'adresse de l'extrémité, et l'autorité doit être connue du pair.
4. **Les sous-réseaux sont-ils inversés côté pair ?** Votre « réseau local » est son « réseau distant ».
5. **L'horloge** des deux machines est-elle correcte ? Un décalage important fait échouer la validation des certificats.

Ensuite, lisez les logs du démon :

```bash
docker compose logs strongswan-a | tail -40
```

L'**assistant de diagnostic** (Avancé → Assistant & anomalies IA) reprend cette liste pour un tunnel donné.

---

## Une passerelle est `unknown` ou `down`

Le serveur n'arrive pas à joindre son socket VICI.

| Symptôme dans les logs du serveur | Cause | Correction |
|---|---|---|
| `permission denied` sur `charon.vici` | Le socket appartient à `root` (`0770`), le serveur tourne sous un autre utilisateur | Faire tourner le serveur en root (c'est ce que fait le lab), ou ajuster les droits |
| `connect: no such file or directory` | Le chemin du socket est faux, ou `charon` n'est pas démarré | Vérifiez `VICI_ENDPOINTS` et l'état du démon |
| `passerelle injoignable à l'enrôlement` | La passerelle n'était pas prête au démarrage du serveur | Elle sera re-testée à chaque sondage ; sinon redémarrez le serveur |

---

## Un certificat révoqué est toujours accepté

C'est le comportement **normal du cache CRL** de charon.

1. La passerelle télécharge la CRL via le **CDP** et la **met en cache** jusqu'à `nextUpdate`.
2. Tant que le cache est frais, elle ne re-télécharge pas — et ne voit donc pas la nouvelle révocation.

**Vérifications :**

- Le certificat contient-il bien un CDP ? Il n'en a un **que si `CRL_URL` était défini au moment de l'émission**. Sinon : réémettez le certificat.
- La passerelle atteint-elle l'URL ?
  ```bash
  docker compose exec strongswan-a curl -s -o /dev/null -w '%{http_code}\n' http://backend:7927/crl.der
  ```
- La CRL contient-elle bien le certificat ?
  ```bash
  curl -s http://localhost:7927/crl.der | openssl crl -inform DER -noout -text | grep -A2 Revoked
  ```
- Réduisez **`CRL_VALIDITY`** pour accélérer le renouvellement du cache (le lab utilise `30s`).

---

## « Votre connexion n'est pas privée » / `NET::ERR_CERT_AUTHORITY_INVALID`

Le certificat est **auto-généré**, signé par la CA interne de l'application — que votre navigateur ne connaît pas. **La connexion est chiffrée** ; c'est l'*identité* du serveur qui n'est pas attestée.

Trois issues, de la plus rapide à la plus propre :

1. **Passer outre** : **Avancé** → **Continuer**. Acceptable en local, pas en production.
2. **Importer la CA interne** sur vos postes d'administration — l'avertissement disparaît définitivement. Marche à suivre dans [Installation](02-installation.md).
3. **Obtenir un vrai certificat** : `ACME_DOMAIN` (Let's Encrypt), ou le vôtre via `TLS_CERT`/`TLS_KEY`.

---

## `NET::ERR_CERT_COMMON_NAME_INVALID` — le nom ne correspond pas

Vous accédez à la console par un nom (`https://vpn.interne.fr:7926`) qui n'est **pas** dans le certificat. Par défaut, celui-ci ne couvre que `localhost`, `127.0.0.1`, `::1` et le nom d'hôte de la machine.

Déclarez le nom d'accès, puis redémarrez — le certificat sera **réémis automatiquement** :

```bash
TLS_SANS="localhost,127.0.0.1,vpn.interne.fr" docker compose up -d
```

---

## `curl` refuse de se connecter : `SSL certificate problem: self-signed certificate`

Attendu, pour la même raison. En ligne de commande :

```bash
curl -sk https://localhost:7926/healthz               # -k : ignore la vérification
curl -s --cacert ca.crt https://localhost:7926/healthz  # mieux : valide contre la CA interne
```

---

## `http://localhost:7926` ne répond pas / répond « 400 Bad Request »

Normal : **7926 est désormais un port HTTPS**. Utilisez `https://localhost:7926`.

L'écouteur **7927**, lui, est en clair — mais il ne sert que `/crl.der` et `/healthz` ; tout le reste y est redirigé en 308 vers HTTPS.

---

## Let's Encrypt (ACME) échoue

Le challenge HTTP-01 exige que **le port 80 public** atteigne l'écouteur clair de l'application.

- Publiez **`80:7927`** (et `443:7926`) — le challenge arrive sur le port 80 côté public.
- Le domaine de `ACME_DOMAIN` doit **résoudre publiquement** vers cette machine.
- Montez `ACME_CACHE` **en volume** : sans lui, chaque redémarrage redemande un certificat et vous atteindrez vite les **quotas** de Let's Encrypt (5 échecs/heure, puis blocage).
- ACME est **inutilisable en réseau privé ou isolé** — utilisez alors la CA interne.

---

## Le port 7926 (ou 5432) est déjà utilisé

`Bind for 0.0.0.0:7926 failed: port is already allocated`

Un autre service occupe le port. Soit vous l'arrêtez, soit vous changez le port publié dans `docker-compose.yml` :

```yaml
ports: ["9090:7926"]
```

(PostgreSQL n'expose **aucun port** sur l'hôte : le serveur y accède par le réseau interne de Docker.)

---

## L'indicateur « temps réel » reste gris

La connexion WebSocket ne s'établit pas.

- Derrière un **reverse proxy**, assurez-vous qu'il **relaie les WebSockets** (`Upgrade` / `Connection`).
- En HTTPS, l'application utilise automatiquement `wss://`.

> **Limitation de sécurité connue** : le serveur ne vérifie le jeton de la connexion WebSocket **que s'il est fourni**. Une connexion **sans** paramètre `?token=` est actuellement acceptée. N'exposez pas l'application directement sur Internet sans protection en amont tant que ce point n'est pas corrigé.

---

## J'ai oublié le mot de passe administrateur

Les comptes ne sont créés qu'**au premier démarrage**, sur une base vide. Il n'y a pas encore d'écran de réinitialisation.

Solutions :

- **En démo / développement** : `docker compose down -v` (⚠️ efface **toutes** les données) puis redémarrer avec un nouveau `SEED_ADMIN_PASSWORD`.
- **En production** : mettez à jour le hachage bcrypt directement dans la table des comptes de la base.

---

## Je ne trouve pas ma réponse

- [FAQ](16-faq.md)
- Les logs : `make logs` (serveur) et `docker compose logs strongswan-a` (démon)
