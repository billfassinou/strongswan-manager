# Glossaire

---

**IPsec** — Suite de protocoles qui chiffre et authentifie le trafic IP. C'est ce que met en œuvre StrongSwan.

**IKE** *(Internet Key Exchange)* — Le protocole qui négocie les clés et les paramètres du tunnel. Deux versions : **IKEv1** (obsolète) et **IKEv2** (à utiliser).

**SA** *(Security Association)* — Le « contrat » de sécurité négocié entre deux pairs. Il y a la **SA IKE** (le canal de négociation) et la **CHILD_SA** (celle qui chiffre réellement vos données).

**CHILD_SA** — La SA enfant : c'est elle qui porte le chiffrement du trafic (ESP). Quand l'interface affiche `up`, c'est qu'elle est installée.

**ESP** *(Encapsulating Security Payload)* — Le protocole qui chiffre les paquets une fois le tunnel monté.

**charon** — Le démon de StrongSwan. C'est lui qui négocie et maintient les tunnels.

**VICI** — L'interface de contrôle officielle de StrongSwan. C'est par elle que la console pilote `charon` (charger une connexion, monter un tunnel, lire l'état). **Aucun fichier de configuration n'est écrit.**

**swanctl** — L'outil en ligne de commande de StrongSwan. Il utilise VICI, comme la console. `swanctl --list-sas` sert à vérifier ce que la console a fait.

---

**PSK** *(Pre-Shared Key)* — Une clé secrète partagée par les deux extrémités. Simple, mais tout le monde partage le même secret.

**Certificat X.509** — Une identité signée par une autorité. Chaque passerelle a la sienne : on peut en révoquer une sans toucher aux autres. Plus robuste qu'un PSK.

**CA** *(Certificate Authority)* — L'autorité qui signe les certificats. La console en embarque une, créée automatiquement.

**SAN** *(Subject Alternative Name)* — Le champ du certificat qui porte l'identité réseau (une IP, un nom DNS). **Il doit correspondre à l'adresse de l'extrémité du tunnel**, sinon l'authentification échoue.

**CRL** *(Certificate Revocation List)* — La liste signée des certificats révoqués.

**CDP** *(CRL Distribution Point)* — L'URL, inscrite **dans le certificat**, où la passerelle va télécharger la CRL. C'est ainsi que la révocation se propage (StrongSwan n'a pas de commande pour « pousser » une CRL).

**OCSP** — Une vérification de révocation en ligne, à la place de la CRL. *Pas encore implémenté.*

---

**PFS** *(Perfect Forward Secrecy)* — Propriété qui garantit qu'un attaquant ayant volé une clé ne pourra pas déchiffrer les échanges **passés**. À laisser activé, toujours.

**Diffie-Hellman / groupe DH** — L'algorithme d'échange de clés. `modp1024` est **faible** ; `modp2048` est acceptable ; `ecp384` est recommandé.

**Proposition** *(proposal)* — Une suite cryptographique complète, écrite avec des tirets : `aes256-sha256-modp2048` = chiffrement AES-256, empreinte SHA-256, échange de clés modp2048. Les deux pairs doivent avoir au moins une proposition **en commun**.

**AES-GCM** — Chiffrement authentifié moderne (`aes256gcm16`), recommandé pour l'ESP.

**ML-KEM** — Échange de clés **post-quantique** (`mlkem768`). Protège contre un futur ordinateur quantique. **Exige StrongSwan 6.0+**.

**3DES / MD5 / modp1024** — Algorithmes **obsolètes**. Leur présence fait chuter le score de sécurité. À remplacer.

---

**Site-à-site** — Un tunnel entre deux réseaux (deux agences, par exemple). Les deux extrémités ont une adresse fixe.

**Host-à-host** — Un tunnel entre deux machines, sans réseau derrière.

**Road warrior** — Un utilisateur nomade, dont l'adresse est dynamique. Se connecte typiquement en EAP ou par certificat.

**EAP / XAuth** — Méthodes d'authentification des utilisateurs (souvent adossées à un serveur RADIUS).

**RADIUS** — Serveur d'authentification et de comptabilité (AAA) vers lequel déléguer les accès nomades.

**Pool d'IP virtuelles** — La plage d'adresses distribuée aux clients nomades une fois connectés.

**Split-tunnel** — Le fait de ne faire passer dans le tunnel qu'une partie du trafic (les réseaux de l'entreprise), le reste sortant directement sur Internet.

---

**Rekey** — Renégociation des clés d'un tunnel, sans coupure du trafic.

**Trap / on-demand** — Une politique qui monte le tunnel automatiquement au premier paquet.

**Shunt** — Une politique d'exception : laisser passer (`pass`) ou bloquer (`drop`) un trafic **sans** le chiffrer.

**XFRM / VTI** — Interfaces réseau permettant un VPN « route-based » (on route dans une interface, au lieu de définir des politiques par sous-réseau).

---

**JWT** — Le jeton d'authentification remis à la connexion, à joindre à chaque appel d'API.

**RBAC** — Le contrôle d'accès par rôle (`admin`, `operator`, `auditor`, `viewer`).

**Append-only** — Se dit du journal d'audit : on ne peut qu'**ajouter**. La base refuse toute modification ou suppression.

**Prometheus** — Le format de métriques exposé sur `/metrics`, à brancher sur votre supervision.
