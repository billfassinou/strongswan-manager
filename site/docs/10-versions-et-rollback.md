# Versions & rollback

Chaque fois qu'une configuration de tunnel est appliquée, la console **enregistre une version** : un instantané complet, horodaté, attribué à son auteur.

C'est votre filet de sécurité : vous pouvez modifier un tunnel en production en sachant que **le retour arrière prend une seconde**.

---

## Quand une version est-elle créée ?

| Action | Version |
|---|---|
| Créer un tunnel | **v1** |
| Le modifier | v2, v3, … |
| Faire un rollback | une **nouvelle** version (v4) qui restaure le contenu d'une ancienne |

Un rollback n'efface donc jamais l'historique : il **ajoute** une version. La chronologie reste complète et auditable.

---

## Consulter l'historique

L'historique d'un tunnel est disponible via l'API :

```bash
curl -sk https://localhost:7926/api/v1/tunnels/<ID>/versions \
  -H "Authorization: Bearer $TOKEN" | jq
```

Chaque entrée porte son numéro (`n`), son message (`création`, `mise à jour`, `rollback vers v1`), son auteur et sa date.

> L'écran dédié à l'historique n'est pas encore exposé dans l'interface : les versions se consultent et se restaurent par l'API. La fonctionnalité serveur, elle, est complète.

---

## Revenir à une version antérieure

```bash
curl -sk -X POST https://localhost:7926/api/v1/tunnels/<ID>/rollback \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"version": 1}'
```

Ce que fait la console :

1. elle relit l'instantané de la version demandée ;
2. elle **recalcule le score** ;
3. elle **recharge la configuration à chaud** sur la passerelle (et sur la passerelle pair, si le tunnel est géré des deux côtés) ;
4. elle crée une nouvelle version (`rollback vers v1`) et une **entrée d'audit**.

Si vous omettez `version`, c'est la **version précédente** qui est restaurée.

---

## Cas d'usage typique

Vous durcissez un vieux tunnel (score 5 → 100). Le pair, un équipement ancien, n'accepte pas AES-GCM : le tunnel ne remonte pas.

1. **Rollback** vers la version d'avant → le tunnel remonte immédiatement.
2. Vous négociez avec l'équipe d'en face une fenêtre de maintenance.
3. Vous réappliquez le durcissement des deux côtés.

Aucune perte de temps, aucune reconstitution de configuration à la main.

---

## Ce qui n'est pas versionné

- Les **secrets** et les **certificats** (ils vivent dans le coffre et la PKI, avec leur propre cycle de vie).
- Les **modules de configuration** (pools, RADIUS, politiques…) : ils sont persistés mais sans historique.
- La **suppression** d'un tunnel : elle emporte ses versions. Pour une interruption réversible, utilisez **Couper** plutôt que **Supprimer**.

---

## Et ensuite ?

→ [Modules de configuration](11-modules-configuration.md)
