# Découvrir l'interface

Vous êtes connecté. Voici ce que vous avez sous les yeux.

---

## La barre latérale : quatre groupes

| Groupe | Modules |
|---|---|
| **Supervision** | Tableau de bord · Connexions · Topologie · Monitoring & alertes · Journal d'audit |
| **Configuration** | Éditeur de tunnel · PKI & Certificats · Secrets · Utilisateurs VPN |
| **StrongSwan** | Pools & IP virtuelles · RADIUS / AAA · Politiques & routage · Autorités & enrôlement · Paramètres du démon |
| **Avancé** | Sécurité & Conformité · Assistant & anomalies IA · Passerelles & ZTP · Administration |

Un **badge rouge** apparaît sur **Connexions** dès qu'un tunnel est *down* : c'est votre alerte immédiate.

En bas de la barre latérale : votre identité, votre rôle, et le bouton de **déconnexion**.

---

## La barre du haut

- Le **titre** de la page courante et votre identité/rôle.
- **◑** : bascule **thème clair / sombre** (le choix est mémorisé).
- **+ Nouveau tunnel** : raccourci vers l'éditeur. Ce bouton **n'apparaît pas** si votre rôle est en lecture seule.

---

## Le tableau de bord

C'est la page d'accueil. Vous y trouvez :

- **Quatre compteurs** : tunnels actifs, en négociation, *down*, nombre de passerelles.
- **La liste des tunnels** avec leur état et leur **score de sécurité**.
- **La liste des passerelles** avec leur version de StrongSwan et leur état.
- Un indicateur **« temps réel »** : quand il est vert, l'interface reçoit les changements d'état en direct (WebSocket). Vous n'avez pas besoin de rafraîchir la page — quand un tunnel monte, la ligne change toute seule.

---

## Ce que vous pouvez faire selon votre rôle

| Vous êtes… | Vous voyez | Vous pouvez modifier |
|---|---|---|
| **Administrateur** (`admin`) | Tout | **Tout** |
| **Opérateur** (`operator`) | Tout | **Tout** |
| **Auditeur** (`auditor`) | Tout | Rien (boutons masqués) |
| **Lecture seule** (`viewer`) | Tout | Rien |

En lecture seule, les boutons d'action (créer, modifier, supprimer, monter, couper…) sont **masqués**. Et si quelqu'un tentait d'appeler l'API directement, le serveur répondrait **403**. La lecture seule est une vraie garantie, pas un simple habillage.

Détail complet : [Rôles & permissions](A1-roles-et-permissions.md).

---

## Les états d'un tunnel

Vous les verrez partout (pastilles colorées) :

| État | Couleur | Signification |
|---|---|---|
| `up` | vert | La SA est établie, le tunnel fonctionne |
| `negotiating` | orange | Négociation IKE / renégociation en cours |
| `installing` | orange | La configuration vient d'être appliquée, pas encore établie |
| `down` | rouge | Aucune SA établie |
| `unknown` | gris | La passerelle n'a pas pu être interrogée |

---

## Le score de sécurité

Chaque tunnel affiche un score sur 100, coloré :

- **≥ 85** vert — conforme aux bonnes pratiques
- **65 – 84** orange — à durcir
- **< 65** rouge — critique

Cliquer dessus ne fait rien : le détail est dans **Sécurité & Conformité**, avec un bouton **« Corriger »** qui ouvre l'éditeur pré-rempli. Voir [Score de sécurité](09-score-de-securite.md).

---

## Et ensuite ?

→ [Créer votre premier tunnel](04-creer-un-tunnel.md)
