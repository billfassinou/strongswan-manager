# Rôles & permissions

Quatre rôles. La règle est simple : **`admin` et `operator` écrivent, `auditor` et `viewer` non.**

---

## Les rôles

| Rôle | Lecture | Écriture | Pour qui |
|---|---|---|---|
| `admin` | ✅ | **✅** | Administrateur de la console |
| `operator` | ✅ | **✅** | Exploitation, NOC |
| `auditor` | ✅ | ❌ | RSSI, audit, conformité |
| `viewer` | ✅ | ❌ | Supervision, support niveau 1 |

Les quatre comptes correspondants sont créés au premier démarrage. Voir [Installation](02-installation.md).

---

## Ce que « lecture seule » garantit vraiment

Deux barrières, pas une :

1. **L'interface masque** les boutons d'action (créer, éditer, supprimer, monter, couper, générer, révoquer…).
2. **Le serveur refuse.** Toute route modifiante est protégée : un appel direct à l'API avec un jeton `auditor` ou `viewer` reçoit :

```json
403 Forbidden
{"error":"forbidden","message":"action réservée — rôle en lecture seule"}
```

Un utilisateur en lecture seule ne peut donc **rien** modifier, même en contournant l'interface.

---

## Qui peut faire quoi

| Action | admin | operator | auditor | viewer |
|---|:---:|:---:|:---:|:---:|
| Voir le tableau de bord, la topologie, les passerelles | ✅ | ✅ | ✅ | ✅ |
| Lire les tunnels, secrets (masqués), certificats, audit | ✅ | ✅ | ✅ | ✅ |
| Créer / modifier / supprimer un tunnel | ✅ | ✅ | ❌ | ❌ |
| Monter / couper / renégocier un tunnel | ✅ | ✅ | ❌ | ❌ |
| Rollback d'une configuration | ✅ | ✅ | ❌ | ❌ |
| Créer / supprimer un secret | ✅ | ✅ | ❌ | ❌ |
| Émettre / révoquer un certificat, publier la CRL | ✅ | ✅ | ❌ | ❌ |
| Modifier les modules de configuration (pools, RADIUS…) | ✅ | ✅ | ❌ | ❌ |

> **Note** : `admin` et `operator` ont aujourd'hui **les mêmes droits**. La distinction existe pour préparer une granularité plus fine (édition Enterprise), et pour la traçabilité : l'audit enregistre qui a agi.

---

## Authentification

- **Mot de passe** haché en **bcrypt**.
- À la connexion, le serveur émet un **JWT** (HS256) contenant l'identité et le rôle, valable **1 heure** par défaut (`JWT_TTL`).
- Le jeton est transmis à chaque appel : `Authorization: Bearer <jeton>`.
- Le front le conserve dans le `localStorage` du navigateur.

Sans jeton valide, toute route protégée répond **401**.

---

## Limitation connue — WebSocket

Le flux temps réel (`/api/v1/ws`) accepte un jeton en paramètre (`?token=`), mais **ne le vérifie que s'il est présent** : une connexion **sans** jeton est actuellement acceptée.

**Conséquence** : quelqu'un ayant accès au réseau du serveur peut observer les changements d'état des tunnels sans s'authentifier.

**Mitigation** : n'exposez pas l'application directement sur Internet ; placez-la derrière un reverse proxy qui filtre l'accès. Correction prévue.

---

## Gestion des comptes

Il n'existe **pas encore** d'écran de création/désactivation de comptes ni de SSO (ces fonctions relèvent de l'édition **Enterprise**). Les quatre comptes sont seedés au premier démarrage ; toute modification passe aujourd'hui par la base.

---

## Voir aussi

- [Découvrir l'interface](03-decouvrir-linterface.md)
- [API REST & WebSocket](A3-api.md)
