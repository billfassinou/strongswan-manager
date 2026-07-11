// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

export interface Field {
  k: string;
  label: string;
  type?: "text" | "select" | "toggle";
  opts?: string[];
}
export interface Schema {
  kind: string;
  title: string;
  sub: string;
  add: string;
  fields: Field[];
}

// Schémas des modules de configuration (pilotent le CRUD générique du front).
export const SCHEMAS: Record<string, Schema> = {
  pool: {
    kind: "pool",
    title: "Pools & IP virtuelles",
    sub: "IP virtuelles · DNS/NBNS · split-tunnel (section pools de swanctl)",
    add: "+ Pool",
    fields: [
      { k: "name", label: "Nom" },
      { k: "range", label: "Plage" },
      { k: "source", label: "Source", type: "select", opts: ["Interne", "SQL", "DHCP", "RADIUS"] },
      { k: "dns", label: "DNS poussé" },
      { k: "split", label: "Split-tunnel" },
    ],
  },
  radius: {
    kind: "radius",
    title: "RADIUS / AAA",
    sub: "relais EAP · accounting · CoA (plugin eap-radius)",
    add: "+ Serveur",
    fields: [
      { k: "name", label: "Nom" },
      { k: "addr", label: "Adresse:port" },
      { k: "role", label: "Rôle", type: "select", opts: ["Primaire", "Secondaire"] },
      { k: "acct", label: "Accounting", type: "toggle" },
    ],
  },
  policy: {
    kind: "policy",
    title: "Politiques & routage",
    sub: "shunt (pass/drop/bypass) · on-demand (trap) · route-based (XFRM/VTI)",
    add: "+ Politique",
    fields: [
      { k: "name", label: "Nom" },
      { k: "kind", label: "Type", type: "select", opts: ["shunt", "trap", "route-based"] },
      { k: "detail", label: "Détail" },
      { k: "iface", label: "Interface (route-based)" },
    ],
  },
  authority: {
    kind: "authority",
    title: "Autorités & enrôlement",
    sub: "section authorities · URIs CRL/OCSP · SCEP/EST",
    add: "+ Autorité",
    fields: [
      { k: "name", label: "Nom" },
      { k: "crl_uri", label: "URI CRL" },
      { k: "ocsp_uri", label: "URI OCSP" },
      { k: "enroll", label: "Enrôlement", type: "select", opts: ["—", "SCEP", "EST"] },
    ],
  },
  vpnuser: {
    kind: "vpnuser",
    title: "Utilisateurs VPN",
    sub: "road warriors : provisioning, quotas, plages horaires, révocation",
    add: "+ Provisionner",
    fields: [
      { k: "name", label: "Identité" },
      { k: "method", label: "Méthode", type: "select", opts: ["EAP-TLS", "Certificat", "EAP-MSCHAPv2"] },
      { k: "quota", label: "Quota" },
      { k: "window", label: "Plage horaire" },
      { k: "enabled", label: "Actif", type: "toggle" },
    ],
  },
  alert: {
    kind: "alert",
    title: "Monitoring & alertes",
    sub: "règles d'alerte multicanales (métriques exposées sur /metrics)",
    add: "+ Règle",
    fields: [
      { k: "name", label: "Nom" },
      { k: "condition", label: "Condition" },
      { k: "channels", label: "Canaux" },
      { k: "enabled", label: "Activée", type: "toggle" },
    ],
  },
};
