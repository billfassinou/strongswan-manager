# Connecter de vraies passerelles

Par défaut, l'application tourne en **mode démo** : une passerelle `gw-local` adossée à un **adaptateur VICI simulé**. Tout fonctionne (créer un tunnel, le monter, voir l'état) mais aucun paquet IPsec n'est réellement chiffré.

Cette page explique comment piloter de **vrais démons StrongSwan**.

---

## Comment le serveur parle aux passerelles

Il utilise **VICI**, l'interface de contrôle officielle de StrongSwan (le plugin `vici` de `charon`). Concrètement, le serveur ouvre le **socket VICI** du démon et lui envoie des commandes :

| Commande VICI | Utilisée pour |
|---|---|
| `version` | Détecter la version de StrongSwan |
| `load-conn` | Charger/mettre à jour une connexion |
| `unload-conn` | Retirer une connexion |
| `load-shared` | Charger un PSK |
| `load-cert` / `load-key` | Charger un certificat et sa clé |
| `list-sas` | Lire l'état réel des SA (utilisé par la supervision) |
| `initiate` / `terminate` / `rekey` | Monter, couper, renégocier |

**Aucun fichier de configuration n'est écrit.**

---

## Le lab : deux vraies passerelles en une commande

```bash
cd backend
make lab-up
```

Cela démarre, en plus de l'application : **deux conteneurs strongSwan** (`strongswan-a`, `strongswan-b`), et connecte le serveur à leurs sockets VICI.

Vous verrez alors **`gw-a`** et **`gw-b`** dans **Passerelles**, avec leur version remontée par le démon lui-même.

Faites ensuite un vrai tunnel entre les deux : [Site-à-site des deux côtés](08-site-a-site-deux-cotes.md).

Pour tout arrêter et nettoyer :

```bash
make lab-down
```

### Vérifier côté démon

```bash
docker compose exec strongswan-a swanctl --list-conns --uri unix:///vicirun/charon.vici
docker compose exec strongswan-a swanctl --list-sas   --uri unix:///vicirun/charon.vici
```

La connexion que vous avez créée dans l'interface doit y apparaître — la preuve qu'elle a bien été chargée par VICI.

---

## Déclarer vos propres passerelles

Le serveur lit la variable **`VICI_ENDPOINTS`**, une liste `nom=endpoint` séparée par des virgules :

```bash
VICI_ENDPOINTS="gw-paris=unix:/chemin/charon.vici,gw-dakar=tcp:10.0.0.5:4502"
```

| Forme | Quand l'utiliser |
|---|---|
| `unix:/chemin/vers/charon.vici` | Le démon tourne sur la même machine (ou son socket est partagé, comme dans le lab) |
| `tcp:hôte:port` | Le démon expose VICI en TCP |

Les passerelles sont enregistrées **au démarrage** du serveur. Si `VICI_ENDPOINTS` est vide, on retombe sur le mode démo.

### Ce qu'il faut sur la passerelle

1. **StrongSwan ≥ 5.9** avec le **plugin `vici`** chargé (c'est le cas par défaut avec `strongswan-swanctl`).
2. Le serveur doit pouvoir **atteindre le socket VICI** — c'est le point le plus délicat (voir ci-dessous).
3. Pour l'authentification par **certificat** : le plugin **`openssl`** (paquet `libstrongswan-standard-plugins` sur Debian) est **requis** — sans lui, charon ne sait pas lire un certificat ECDSA.

---

## Limitations connues (à lire avant de déployer)

### Accès au socket VICI

Le socket `charon.vici` appartient à `root` avec des droits `0770`. Dans le lab, le serveur tourne donc **en root** et partage le socket par un volume Docker.

**Ce n'est pas une architecture de production.** La conception prévoit un **agent léger** installé sur chaque passerelle, faisant le pont entre le socket VICI local et le serveur en **mTLS**. **Cet agent n'est pas encore implémenté.** Aujourd'hui, réservez ce mode à un réseau d'administration de confiance.

### Version de StrongSwan

Debian 12 fournit **5.9.8**. C'est au-dessus du socle minimum (5.9), mais :

- **pas de post-quantique** (ML-KEM) — il faut StrongSwan **6.0+** ;
- l'interface signale les versions `5.x` en orange dans **Passerelles**.

### Certificats et CRL

- Les certificats sont transmis à charon en **DER** (le serveur convertit depuis le PEM) : charon rejette le PEM brut sur ces versions.
- StrongSwan **n'a pas de commande VICI pour charger une CRL**. La révocation passe donc par le **CRL Distribution Point** inscrit dans les certificats (`CRL_URL`), que la passerelle télécharge elle-même via son plugin `curl`. Voir [PKI & certificats](07-pki-certificats.md).

---

## Et ensuite ?

→ [Dépannage](15-depannage.md)
