# Score de sécurité

Chaque tunnel reçoit une note **sur 100**, calculée à partir de sa configuration cryptographique. Elle apparaît partout : tableau de bord, liste des connexions, éditeur (en direct), page Sécurité.

---

## Comment il est calculé

On part de **100** et on retire :

| Constat | Pénalité |
|---|---|
| **IKEv1** (protocole obsolète) | −42 |
| **3DES / DES** (chiffrement faible) | −28 |
| **MD5** (empreinte faible) | −18 |
| **modp1024 / modp768** (groupe Diffie-Hellman faible) | −16 |
| **Perfect Forward Secrecy désactivé** | −10 |
| **Pas de ML-KEM** (aucune préparation post-quantique) | −6 |

Le résultat est borné entre **5** et **100**.

Le même calcul est fait côté serveur (à l'enregistrement) et côté navigateur (en direct dans l'éditeur) : vous voyez le score bouger **pendant que vous saisissez**.

---

## Lire la couleur

| Score | Couleur | Lecture |
|---|---|---|
| **≥ 85** | vert | Conforme aux bonnes pratiques |
| **65 – 84** | orange | Correct mais perfectible (souvent : pas de ML-KEM) |
| **< 65** | rouge | **Critique** — corriger en priorité |

Exemples réels :

- `aes256gcm16-sha384-ecp384-mlkem768` + IKEv2 + PFS → **100**
- `aes256-sha256-modp2048` + IKEv2 + PFS → **94** (il manque juste le post-quantique)
- `3des-md5-modp1024` + IKEv1 → **5** (cumul de tout ce qu'il ne faut pas faire)

---

## Voir l'état du parc

**Avancé → Sécurité & Conformité** :

- un **anneau** donne le score **moyen du parc** ;
- trois compteurs : conformes / à durcir / critiques ;
- un tableau liste **chaque algorithme faible détecté**, tunnel par tunnel.

---

## Durcir un tunnel

Dans le tableau des algorithmes faibles, cliquez sur **Corriger** : l'**éditeur s'ouvre pré-rempli** avec le tunnel concerné.

Remplacez les propositions :

| Avant | Après |
|---|---|
| `3des-md5-modp1024` | `aes256gcm16-sha384-ecp384` |
| IKEv1 | **IKEv2** |
| PFS décoché | PFS **coché** |

Le score se met à jour en direct. **Valider & appliquer** : la configuration est rechargée à chaud, une nouvelle version est créée.

> Le tunnel sera **renégocié** : prévoyez une courte interruption, et assurez-vous que **le pair accepte les nouvelles propositions** — sinon le tunnel ne remontera pas. Vous pourrez toujours [revenir en arrière](10-versions-et-rollback.md).

---

## À propos de ML-KEM (post-quantique)

`mlkem768` protège l'échange de clés contre un futur ordinateur quantique. C'est le seul point qui empêche un tunnel classique d'atteindre 100.

**Attention** : ML-KEM exige **StrongSwan 6.0 ou plus** sur la passerelle. Les distributions actuelles (Debian 12, par exemple) fournissent **5.9.8**, qui ne le supporte pas. Vérifiez la version dans **Passerelles** avant de l'activer — sinon la négociation échouera.

Une pénalité de **−6** est donc **acceptable** sur une passerelle 5.9.x : ce n'est pas une faiblesse exploitable aujourd'hui, c'est une préparation.

---

## Et ensuite ?

→ [Versions & rollback](10-versions-et-rollback.md) — corriger sans risque.
