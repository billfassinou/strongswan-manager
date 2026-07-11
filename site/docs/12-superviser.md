# Superviser

Toutes les pages de ce chapitre sont accessibles **à tous les rôles**, y compris en lecture seule.

---

## Le tableau de bord

**Supervision → Tableau de bord**

- **Compteurs** : tunnels actifs / en négociation / down, nombre de passerelles.
- **Liste des tunnels** avec état et score.
- **Liste des passerelles** avec leur version de StrongSwan.
- Indicateur **« temps réel »**.

### Le temps réel, concrètement

Le serveur interroge chaque passerelle **toutes les 3 secondes** (réglable via `POLL_INTERVAL`) pour lire l'état réel des SA. Dès qu'un état change, il le pousse à votre navigateur par **WebSocket**.

Conséquence pratique : **ne rafraîchissez pas la page**. Montez un tunnel, regardez la ligne : elle passe `installing` → `negotiating` → `up` toute seule.

Si l'indicateur est **gris**, la connexion temps réel est tombée (le navigateur retente automatiquement toutes les 3 secondes).

---

## La topologie

**Supervision → Topologie**

Une vue graphique du parc :

- chaque **passerelle** est un nœud, entouré de ses tunnels ;
- chaque **tunnel** est une arête, **colorée selon son état** (vert = up, orange pointillé = négociation, rouge pointillé = down) ;
- un tunnel dont le pair **n'est pas géré** par la console apparaît comme un petit satellite.

C'est calculé à partir des données réelles (`/gateways` + `/tunnels`) : la carte est toujours à jour, il n'y a rien à saisir.

En dessous : compteurs passerelles / liens sains / en négociation / down.

---

## Les passerelles

**Avancé → Passerelles & ZTP**

La liste des démons StrongSwan gérés :

| Colonne | Ce que ça dit |
|---|---|
| **Nom** | `gw-a`, `gw-local`… |
| **Endpoint VICI** | Comment le serveur y accède (`unix:/gw/a/charon.vici`, ou `mock`) |
| **Version** | La version de StrongSwan **remontée par le démon lui-même**. Affichée en orange si `5.x` (pas de post-quantique). |
| **État** | `up` si la passerelle répond aux interrogations VICI |

Une passerelle `down` ou `unknown` = le serveur n'arrive pas à la joindre. Voir [Dépannage](15-depannage.md).

---

## Anomalies & assistant de diagnostic

**Avancé → Assistant & anomalies IA**

### Les anomalies

Elles sont **dérivées de l'état réel** du parc, sans devinette :

- tunnel **down** → *critique* ;
- tunnel en **négociation prolongée** → *à surveiller* ;
- tunnel avec un **score < 65** → *à surveiller*.

### L'assistant

Posez une question en français ; l'assistant répond **à partir de vos vrais tunnels** :

- « **pourquoi paris-dakar est down ?** » → il identifie le tunnel, rappelle son état et son score, et liste les causes probables (joignabilité du pair sur UDP 500/4500, horloge NTP, propositions/PSK asymétriques).
- « **quels tunnels sont à durcir ?** » → il liste ceux dont le score est critique et indique la correction.
- « **quels tunnels sont down ?** » → il les énumère.

> **Honnêteté** : ce n'est **pas** un modèle de langage. C'est un moteur de **règles déterministes** appliquées à vos données. Il ne peut pas inventer, mais il ne comprendra pas une question hors de son périmètre. Un véritable assistant IA relève de l'édition Premium.

---

## Les métriques Prometheus

L'endpoint **`/metrics`** (public, sans authentification) expose :

| Métrique | Type | Signification |
|---|---|---|
| `strongswan_tunnel_status` | jauge | Par tunnel : `1` = up, `0.5` = négociation, `0` = down |
| `strongswan_vici_errors_total` | compteur | Nombre d'erreurs de communication VICI |

… plus les métriques standard du processus Go.

Branchez-y Prometheus, puis Grafana ou Alertmanager :

```yaml
scrape_configs:
  - job_name: strongswan-manager
    static_configs:
      - targets: ['mon-serveur:8080']
```

C'est aujourd'hui **le moyen d'être alerté** (les canaux de notification internes ne sont pas encore actifs — voir [Modules de configuration](11-modules-configuration.md)).

---

## Et ensuite ?

→ [Auditer](13-auditer.md)
