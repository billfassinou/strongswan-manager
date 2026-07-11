# PKI & certificats

> Rôle requis : **administrateur** ou **opérateur** pour émettre/révoquer. Tout le monde peut consulter.

L'authentification par certificat est plus robuste qu'un PSK : chaque passerelle a sa propre identité, on peut **révoquer** une identité compromise sans toucher aux autres.

---

## L'autorité de certification interne

Elle est **créée automatiquement au premier démarrage** — vous n'avez rien à faire.

**Configuration → PKI & Certificats** affiche l'autorité (`StrongSwan Manager Root CA`, clé ECDSA). Sa clé privée est **chiffrée au repos**, comme les secrets.

---

## Émettre un certificat

Un certificat par passerelle.

1. **PKI & Certificats** → **Générer**.
2. Renseignez :
   - **Nom** : l'identifiant interne, ex. `cert-gw-a` ;
   - **Common Name (CN)** : ex. `gw-a` ;
   - **Usage** : `Serveur` (une passerelle) ou `Client` ;
   - **SAN** : ⚠️ **le point important** — mettez **l'adresse IP de la passerelle**, ex. `203.0.113.10`. Plusieurs valeurs séparées par des virgules (IP ou noms DNS).
3. **Générer**.

> **Pourquoi le SAN est critique** : c'est l'identité IKE. Le pair vérifie que le certificat présenté correspond bien à l'adresse avec laquelle il parle. Un SAN qui ne correspond pas à l'extrémité du tunnel ⇒ l'authentification échoue.

La clé privée est générée, **chiffrée** et stockée. Elle **n'est jamais renvoyée** par l'API.

---

## Créer un tunnel authentifié par certificat

1. **Éditeur de tunnel**.
2. **Authentification** → `Certificat`.
3. Choisissez le **Certificat local** (celui de votre passerelle, dont le SAN = votre extrémité locale).
4. Si le pair est **aussi géré** par la console, choisissez son **Certificat du pair**.
5. **Valider & appliquer**.

À l'application, la console charge sur la passerelle :

- l'**autorité de certification** (pour qu'elle sache valider le certificat du pair) ;
- le **certificat** de la passerelle ;
- sa **clé privée**.

Puis elle configure la connexion en authentification par clé publique, avec les **identités IKE** positionnées sur les adresses des extrémités.

Vous pouvez ensuite **Monter** le tunnel normalement.

---

## Révoquer un certificat

1. **PKI & Certificats** → **Révoquer** sur la ligne → confirmer.
2. Le certificat passe en état **Révoqué**.
3. La **CRL est immédiatement régénérée** et signée par l'autorité.

---

## La CRL (liste de révocation)

### Comment les passerelles l'obtiennent

StrongSwan **n'a pas de commande pour « pousser » une CRL**. Le mécanisme standard est le **CRL Distribution Point (CDP)** : une URL inscrite **dans le certificat**, que la passerelle va chercher elle-même (plugin `curl` de charon).

Cette URL se configure avec la variable **`CRL_URL`**, par exemple :

```
CRL_URL=http://mon-serveur:8080/crl.der
```

L'endpoint **`/crl.der`** est **public** (sans authentification) : c'est normal, une CRL est un objet public, et les passerelles doivent pouvoir la lire sans jeton.

> ⚠️ **Les certificats déjà émis ne contiennent pas de CDP** si `CRL_URL` était vide au moment de leur émission. Définissez `CRL_URL` **avant** d'émettre vos certificats.

### Actions disponibles

| Bouton | Effet |
|---|---|
| **Publier la CRL** | Régénère la CRL (numéro incrémenté) et la persiste |
| **Télécharger la CRL (.der)** | Récupère le fichier `/crl.der` |

### Quand la révocation prend-elle effet ?

Quand la passerelle **re-télécharge** la CRL. Elle la met en cache jusqu'à la date `nextUpdate`, réglée par **`CRL_VALIDITY`** (24 h par défaut).

- Pour une prise en compte rapide en test, baissez `CRL_VALIDITY` (ex. `30s` — c'est ce que fait `make lab-up`).
- En production, gardez une durée raisonnable (quelques heures à quelques jours).

Voir [Dépannage](15-depannage.md) si un certificat révoqué reste accepté.

---

## Vérifier soi-même la CRL

```bash
curl -s http://localhost:8080/crl.der | openssl crl -inform DER -noout -text
```

Vous devez y voir le numéro de CRL et les numéros de série révoqués.

---

## Ce qui n'existe pas encore

- **Répondeur OCSP** (vérification en ligne, à la place de la CRL).
- **Enrôlement SCEP/EST** (les passerelles demandent leur certificat automatiquement).
- Import d'une **autorité externe** (aujourd'hui l'autorité est interne).

---

## Et ensuite ?

→ [Site-à-site des deux côtés](08-site-a-site-deux-cotes.md) — établir un vrai tunnel entre deux passerelles que vous gérez.
