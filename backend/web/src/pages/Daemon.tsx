// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useEffect, useState } from "react";
import { api } from "../api";
import { useToast } from "../ui";
import { useAuth } from "../auth";

const DEFAULTS = {
  threads: "16",
  retransmit_tries: "5",
  retransmit_timeout: "4",
  dns: "10.1.0.53",
  install_routes: true,
  fragmentation: true,
  log_ike: "1",
};

// Paramètres du démon charon (strongswan.conf), persistés comme un singleton config_items
// de kind « daemon » (name = charon).
export function Daemon() {
  const { me } = useAuth();
  const toast = useToast();
  const canWrite = !!me?.can_write;
  const [id, setId] = useState<string | null>(null);
  const [v, setV] = useState<Record<string, any>>({ ...DEFAULTS });

  async function load() {
    const r = await api.get("/config/daemon");
    const item = (r.items || [])[0];
    if (item) {
      setId(item.id);
      setV({ ...DEFAULTS, ...item.data });
    }
  }
  useEffect(() => {
    load();
  }, []);

  const set = (k: string, val: any) => setV((s) => ({ ...s, [k]: val }));

  async function save() {
    const body = { name: "charon", data: v };
    try {
      if (id) await api.put(`/config/daemon/${id}`, body);
      else {
        const r = await api.post("/config/daemon", body);
        setId(r.id);
      }
      toast("Paramètres enregistrés · rechargement à chaud", "ok");
    } catch (e: any) {
      toast(e.message, "err");
    }
  }

  return (
    <div className="grid" style={{ gridTemplateColumns: "1fr 1fr", gap: 16 }}>
      <div className="card">
        <div className="card-head">
          <h2>Paramètres charon</h2>
          <div className="act muted" style={{ fontSize: 12 }}>
            strongswan.conf
          </div>
        </div>
        <div style={{ padding: 18, display: "flex", flexDirection: "column", gap: 13 }}>
          <Num l="Threads worker" k="threads" v={v} set={set} />
          <Num l="Retransmissions IKE" k="retransmit_tries" v={v} set={set} />
          <Num l="Timeout retransmission (s)" k="retransmit_timeout" v={v} set={set} />
          <Txt l="DNS attribué" k="dns" v={v} set={set} />
          <Tog l="install_routes" k="install_routes" v={v} set={set} />
          <Tog l="Fragmentation IKEv2" k="fragmentation" v={v} set={set} />
          <button className="btn pri" disabled={!canWrite} onClick={save} style={{ marginTop: 4, justifyContent: "center" }}>
            Valider &amp; recharger
          </button>
        </div>
      </div>
      <div className="card">
        <div className="card-head">
          <h2>Journalisation</h2>
        </div>
        <div style={{ padding: 18 }}>
          <label className="flabel">Niveau du sous-système IKE</label>
          <select className="field" value={v.log_ike} onChange={(e) => set("log_ike", e.target.value)}>
            {["-1 silencieux", "0 défaut", "1 audit", "2 contrôle", "3 diagnostic"].map((o) => (
              <option key={o} value={o[0] === "-" ? "-1" : o[0]}>
                {o}
              </option>
            ))}
          </select>
          <p className="muted" style={{ fontSize: 12.5, marginTop: 14 }}>
            Les métriques du démon sont exposées sur <span className="mono">/metrics</span> (Prometheus). Le rechargement
            à chaud applique la nouvelle configuration sans coupure.
          </p>
        </div>
      </div>
    </div>
  );
}

function Row({ l, children }: { l: string; children: any }) {
  return (
    <div className="row" style={{ justifyContent: "space-between", gap: 14 }}>
      <label className="flabel" style={{ margin: 0 }}>
        {l}
      </label>
      {children}
    </div>
  );
}
function Num({ l, k, v, set }: any) {
  return (
    <Row l={l}>
      <input className="field" style={{ maxWidth: 160 }} value={v[k]} onChange={(e) => set(k, e.target.value)} />
    </Row>
  );
}
function Txt({ l, k, v, set }: any) {
  return (
    <Row l={l}>
      <input className="field mono" style={{ maxWidth: 160 }} value={v[k]} onChange={(e) => set(k, e.target.value)} />
    </Row>
  );
}
function Tog({ l, k, v, set }: any) {
  return (
    <Row l={l}>
      <input type="checkbox" checked={!!v[k]} onChange={(e) => set(k, e.target.checked)} />
    </Row>
  );
}
