// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useEffect, useMemo, useState } from "react";
import { api, Gateway, Secret, Certificate, Tunnel } from "../api";
import { scoreTunnel, scoreColor } from "../score";
import { useToast } from "../ui";
import { useNav } from "../nav";

const BLANK = {
  id: "",
  name: "nouveau-tunnel",
  gateway_id: "",
  peer_gateway_id: "",
  type: "site-to-site",
  ike_version: 2,
  localAddr: "203.0.113.10",
  localNet: "10.1.0.0/16",
  remoteAddr: "198.51.100.20",
  remoteNet: "10.2.0.0/16",
  auth: "psk",
  secret_ref: "",
  cert_ref: "",
  peer_cert_ref: "",
  ike: "aes256-sha256-modp2048",
  esp: "aes256gcm16",
  pfs: true,
};

export function Editor() {
  const toast = useToast();
  const { go, arg } = useNav();
  const [gws, setGws] = useState<Gateway[]>([]);
  const [secrets, setSecrets] = useState<Secret[]>([]);
  const [certs, setCerts] = useState<Certificate[]>([]);
  const [f, setF] = useState({ ...BLANK });
  const [busy, setBusy] = useState(false);
  const editing = !!f.id;

  useEffect(() => {
    (async () => {
      const [g, s, c] = await Promise.all([api.get("/gateways"), api.get("/secrets"), api.get("/certificates")]);
      setGws(g.items || []);
      setSecrets(s.items || []);
      setCerts(c.items || []);
      const t: Tunnel | null = arg?.tunnel || null;
      if (t) {
        setF({
          id: t.id,
          name: t.name,
          gateway_id: t.gateway_id,
          peer_gateway_id: t.peer_gateway_id || "",
          type: t.type,
          ike_version: t.ike_version,
          localAddr: t.local?.addr || "",
          localNet: (t.local?.subnets || []).join(","),
          remoteAddr: t.remote?.addr || "",
          remoteNet: (t.remote?.subnets || []).join(","),
          auth: t.auth_method,
          secret_ref: "",
          cert_ref: "",
          peer_cert_ref: "",
          ike: (t.proposals_ike || []).join(","),
          esp: (t.proposals_esp || []).join(","),
          pfs: t.pfs,
        });
      } else {
        setF((s0) => ({ ...s0, gateway_id: (g.items || [])[0]?.id || "" }));
      }
    })();
  }, [arg]);

  const set = (k: string, v: any) => setF((s) => ({ ...s, [k]: v }));

  const sc = useMemo(
    () =>
      scoreTunnel({
        ike_version: f.ike_version,
        pfs: f.pfs,
        proposals_ike: f.ike.split(",").map((x) => x.trim()),
        proposals_esp: f.esp.split(",").map((x) => x.trim()),
      }),
    [f.ike, f.esp, f.ike_version, f.pfs]
  );
  const col = scoreColor(sc.score);
  const off = 314 - (314 * sc.score) / 100;

  async function apply() {
    setBusy(true);
    const body: any = {
      name: f.name,
      gateway_id: f.gateway_id,
      peer_gateway_id: f.peer_gateway_id || undefined,
      peer_cert_ref: f.peer_cert_ref || undefined,
      type: f.type,
      ike_version: f.ike_version,
      local: { addr: f.localAddr, subnets: splitList(f.localNet) },
      remote: { addr: f.remoteAddr, subnets: splitList(f.remoteNet) },
      auth: { method: f.auth, secret_ref: f.secret_ref || undefined, cert_ref: f.cert_ref || undefined },
      proposals: { ike: splitList(f.ike), esp: splitList(f.esp) },
      pfs: f.pfs,
    };
    try {
      const r = editing ? await api.put(`/tunnels/${f.id}`, body) : await api.post("/tunnels", body);
      toast(`${editing ? "Modifié" : "Créé"} · score ${r.security_score} · v${r.config_version}`, "ok");
      go("tunnels");
    } catch (e: any) {
      const details = e.body?.details?.map((d: any) => d.issue).join(", ");
      toast(details ? `${e.message} — ${details}` : e.message, "err");
      setBusy(false);
    }
  }

  return (
    <div className="grid" style={{ gridTemplateColumns: "1.5fr 1fr", gap: 16 }}>
      <div className="card">
        <div className="card-head">
          <h2>{editing ? "Éditer : " + f.name : "Nouveau tunnel"}</h2>
        </div>
        <div style={{ padding: 18, display: "flex", flexDirection: "column", gap: 12 }}>
          <F label="Nom" v={f.name} on={(v) => set("name", v)} />
          <div className="row">
            <Sel label="Passerelle" v={f.gateway_id} on={(v) => set("gateway_id", v)} opts={gws.map((g) => [g.id, g.name])} />
            <Sel
              label="Passerelle pair (S2S géré des 2 côtés)"
              v={f.peer_gateway_id}
              on={(v) => set("peer_gateway_id", v)}
              opts={[["", "— aucune —"], ...gws.map((g) => [g.id, g.name] as [string, string])]}
            />
          </div>
          <div className="row">
            <Sel label="Type" v={f.type} on={(v) => set("type", v)} opts={[["site-to-site", "Site-à-site"], ["host-to-host", "Host-à-host"], ["road-warrior", "Road warrior"]]} />
            <Sel label="Version IKE" v={String(f.ike_version)} on={(v) => set("ike_version", Number(v))} opts={[["2", "IKEv2"], ["1", "IKEv1"]]} />
          </div>
          <div className="row">
            <F label="Extrémité locale" v={f.localAddr} on={(v) => set("localAddr", v)} mono />
            <F label="Réseau local" v={f.localNet} on={(v) => set("localNet", v)} mono />
          </div>
          <div className="row">
            <F label="Extrémité distante" v={f.remoteAddr} on={(v) => set("remoteAddr", v)} mono />
            <F label="Réseau distant" v={f.remoteNet} on={(v) => set("remoteNet", v)} mono />
          </div>
          <Sel label="Authentification" v={f.auth} on={(v) => set("auth", v)} opts={[["psk", "PSK"], ["cert", "Certificat"], ["eap", "EAP"]]} />
          {f.auth === "psk" && (
            <Sel label="Secret PSK" v={f.secret_ref} on={(v) => set("secret_ref", v)} opts={[["", "— aucun —"], ...secrets.filter((s) => s.type === "psk").map((s) => [s.name, s.name] as [string, string])]} />
          )}
          {f.auth === "cert" && (
            <div className="row">
              <Sel label="Certificat local" v={f.cert_ref} on={(v) => set("cert_ref", v)} opts={[["", "— aucun —"], ...certs.map((c) => [c.name, c.name] as [string, string])]} />
              <Sel label="Certificat du pair" v={f.peer_cert_ref} on={(v) => set("peer_cert_ref", v)} opts={[["", "— aucun —"], ...certs.map((c) => [c.name, c.name] as [string, string])]} />
            </div>
          )}
          <F label="Propositions IKE" v={f.ike} on={(v) => set("ike", v)} mono />
          <F label="Propositions ESP" v={f.esp} on={(v) => set("esp", v)} mono />
          <label className="row" style={{ gap: 8, fontSize: 13 }}>
            <input type="checkbox" checked={f.pfs} onChange={(e) => set("pfs", e.target.checked)} /> Perfect Forward Secrecy
          </label>
          <div className="row" style={{ paddingTop: 6 }}>
            <button className="btn pri" disabled={busy} onClick={apply}>
              {busy ? "Application…" : "Valider & appliquer"}
            </button>
            <button className="btn ghost" onClick={() => go("tunnels")}>
              Annuler
            </button>
          </div>
        </div>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <div className="card" style={{ padding: 18, display: "flex", alignItems: "center", gap: 18 }}>
          <div style={{ position: "relative", width: 112, height: 112, flex: "none" }}>
            <svg width="112" height="112" style={{ transform: "rotate(-90deg)" }}>
              <circle cx="56" cy="56" r="50" fill="none" stroke="var(--surface-3)" strokeWidth="10" />
              <circle cx="56" cy="56" r="50" fill="none" stroke={`var(--${col})`} strokeWidth="10" strokeLinecap="round" strokeDasharray="314" strokeDashoffset={off} />
            </svg>
            <div style={{ position: "absolute", inset: 0, display: "grid", placeItems: "center", textAlign: "center" }}>
              <b className={col} style={{ fontSize: 30, fontWeight: 700 }}>
                {sc.score}
              </b>
            </div>
          </div>
          <div>
            <h3 style={{ marginBottom: 6 }}>Score de sécurité — {sc.score >= 85 ? "excellent" : sc.score >= 65 ? "bon" : "à durcir"}</h3>
            {sc.notes.length ? (
              <ul style={{ margin: 0, paddingLeft: 18, color: "var(--text-dim)", fontSize: 12.5 }}>
                {sc.notes.map((n) => (
                  <li key={n}>{n}</li>
                ))}
              </ul>
            ) : (
              <p className="muted" style={{ margin: 0 }}>
                Configuration conforme aux bonnes pratiques.
              </p>
            )}
          </div>
        </div>
        <div className="card" style={{ padding: 16, fontSize: 13, color: "var(--text-dim)" }}>
          Application à chaud via <span className="mono">swanctl --load-all</span> (VICI). Pour un site-à-site géré des
          deux côtés, renseignez la passerelle pair : la connexion miroir et le secret/certificat sont chargés
          automatiquement sur le pair.
        </div>
      </div>
    </div>
  );
}

function splitList(s: string): string[] {
  return s
    .split(",")
    .map((x) => x.trim())
    .filter(Boolean);
}
function F({ label, v, on, mono }: { label: string; v: string; on: (v: string) => void; mono?: boolean }) {
  return (
    <div style={{ flex: 1 }}>
      <label className="flabel">{label}</label>
      <input className={"field " + (mono ? "mono" : "")} value={v} onChange={(e) => on(e.target.value)} />
    </div>
  );
}
function Sel({ label, v, on, opts }: { label: string; v: string; on: (v: string) => void; opts: [string, string][] }) {
  return (
    <div style={{ flex: 1 }}>
      <label className="flabel">{label}</label>
      <select className="field" value={v} onChange={(e) => on(e.target.value)}>
        {opts.map(([val, lab]) => (
          <option key={val} value={val}>
            {lab}
          </option>
        ))}
      </select>
    </div>
  );
}
