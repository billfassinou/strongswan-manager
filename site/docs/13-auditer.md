# Auditer

**Supervision → Journal d'audit** — accessible à **tous les rôles**, y compris `auditor` et `viewer`.

---

## Ce qui est tracé

Toute action significative écrit une entrée : **qui**, **quoi**, **sur quoi**, **quand**.

| Action | Enregistrée quand… |
|---|---|
| `login` | Quelqu'un se connecte |
| `tunnel.create` / `tunnel.update` / `tunnel.delete` | Un tunnel est créé, modifié, supprimé |
| `tunnel.initiate` / `tunnel.terminate` / `tunnel.rekey` | Un tunnel est monté, coupé, renégocié |
| `tunnel.rollback` | Une configuration antérieure est restaurée |
| `secret.create` / `secret.delete` | Un secret est créé ou supprimé |
| `cert.issue` / `cert.revoke` | Un certificat est émis ou révoqué |
| `crl.publish` | La CRL est régénérée |
| `config.<kind>.create` / `config.update` / `config.delete` | Un module de configuration est modifié |

La page se rafraîchit automatiquement toutes les 5 secondes.

---

## Pourquoi ce journal est fiable

### Il est *append-only*

La base de données **refuse** toute modification ou suppression d'une entrée d'audit — pas par convention, mais par un **déclencheur (trigger) PostgreSQL**. Une tentative d'`UPDATE` ou de `DELETE` lève une erreur, même exécutée directement en SQL par un administrateur de la base.

C'est vérifié par les tests d'intégration du projet.

### Il est chaîné

Chaque entrée porte un **hachage d'intégrité** calculé à partir du hachage de l'entrée précédente. Autrement dit : on ne peut pas retirer ou réécrire une ligne au milieu sans casser la chaîne — la falsification devient **détectable**.

---

## Exploiter le journal

### Dans l'interface

La page affiche les 100 dernières entrées : horodatage, action, cible.

### Par l'API (pour un export, un SIEM…)

```bash
curl -s "http://localhost:8080/api/v1/audit?limit=500" \
  -H "Authorization: Bearer $TOKEN" | jq
```

Chaque entrée renvoie `id`, `actor_id`, `action`, `target`, `timestamp`, `prev_hash`, `integrity_hash`.

Vous pouvez ainsi :

- alimenter un **SIEM** (Splunk, Elastic…) ;
- produire un **export de preuve** pour un audit ISO 27001 / PCI-DSS ;
- vérifier vous-même la chaîne de hachage.

---

## Répondre aux questions d'un auditeur

| Question | Où trouver la réponse |
|---|---|
| « Qui a modifié ce tunnel, et quand ? » | Journal d'audit (`tunnel.update`) + historique des versions du tunnel |
| « Quel est le niveau cryptographique du parc ? » | **Sécurité & Conformité** — score moyen, algorithmes faibles |
| « Ce certificat compromis a-t-il été révoqué ? » | **PKI & Certificats** (état `Révoqué`) + `cert.revoke` dans l'audit |
| « Peut-on effacer une trace ? » | Non — la base le refuse (trigger *append-only*) |
| « Qui a un accès en écriture ? » | [Rôles & permissions](A1-roles-et-permissions.md) |

---

## Ce qui n'existe pas encore

Les **rapports de conformité** clés en main (ANSSI, ISO 27001, PCI-DSS) et les **rapports exécutifs** périodiques relèvent de l'édition **Premium**. Les données brutes, elles, sont déjà là et exportables.

---

## Et ensuite ?

→ [Connecter de vraies passerelles](14-connecter-passerelles-reelles.md)
