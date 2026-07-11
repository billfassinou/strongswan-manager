// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Client HTTP minimal vers l'API backend (même origine). Le JWT est conservé en mémoire
// + localStorage pour survivre à un rechargement.

const TOKEN_KEY = "ssm_token";

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || "";
}
export function setToken(t: string) {
  if (t) localStorage.setItem(TOKEN_KEY, t);
  else localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  body: any;
  constructor(status: number, body: any) {
    super(body?.message || body?.error || `HTTP ${status}`);
    this.status = status;
    this.body = body;
  }
}

async function req(method: string, path: string, body?: any): Promise<any> {
  const headers: Record<string, string> = {};
  const tok = getToken();
  if (tok) headers["Authorization"] = "Bearer " + tok;
  if (body !== undefined) headers["Content-Type"] = "application/json";
  const res = await fetch("/api/v1" + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;
  if (!res.ok) throw new ApiError(res.status, data);
  return data;
}

export const api = {
  get: (p: string) => req("GET", p),
  post: (p: string, b?: any) => req("POST", p, b),
  put: (p: string, b?: any) => req("PUT", p, b),
  del: (p: string) => req("DELETE", p),
};

export interface Me {
  id: string;
  identity: string;
  role: string;
  can_write: boolean;
}

export async function login(identity: string, password: string): Promise<{ role: string }> {
  const r = await req("POST", "/auth/login", { identity, password });
  setToken(r.token);
  return r;
}

export interface Endpoint {
  addr: string;
  subnets: string[];
}
export interface Tunnel {
  id: string;
  name: string;
  gateway_id: string;
  peer_gateway_id?: string;
  type: string;
  ike_version: number;
  local: Endpoint;
  remote: Endpoint;
  auth_method: string;
  proposals_ike: string[];
  proposals_esp: string[];
  pfs: boolean;
  status: string;
  security_score: number;
  warnings: string[];
  config_version: number;
}

export interface Gateway {
  id: string;
  name: string;
  endpoint: string;
  version: string;
  status: string;
}

export interface Certificate {
  id: string;
  name: string;
  cn: string;
  kind: string;
  serial: string;
  status: string;
  not_after: string;
}

export interface Secret {
  id: string;
  name: string;
  type: string;
  used_by: string;
}

export interface AuditEntry {
  id: string;
  actor_id: string;
  action: string;
  target: string;
  timestamp: string;
}

export interface Version {
  id: string;
  n: number;
  message: string;
  created_at: string;
}
