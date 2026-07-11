// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

import { useEffect, useState } from "react";
import { api, Secret } from "../api";
import { Pill, Modal, useToast } from "../ui";
import { useAuth } from "../auth";

export function Secrets() {
  const { me } = useAuth();
  const toast = useToast();
  const canWrite = !!me?.can_write;
  const [items, setItems] = useState<Secret[]>([]);
  const [creating, setCreating] = useState(false);

  async function load() {
    setItems((await api.get("/secrets")).items || []);
  }
  useEffect(() => {
    load();
  }, []);

  async function del(s: Secret) {
    if (!confirm(`Supprimer le secret ${s.name} ?`)) return;
    try {
      await api.del(`/secrets/${s.id}`);
      toast("Secret supprimé", "ok");
      load();
    } catch (e: any) {
      toast(e.message, "err");
    }
  }

  return (
    <div className="card">
      <div className="card-head">
        <h2>Coffre de secrets</h2>
        <div className="act muted" style={{ fontSize: 12, marginRight: "auto", marginLeft: 10 }}>
          chiffrés au repos · valeurs jamais affichées
        </div>
        <div className="act">
          <button className="btn sm pri" disabled={!canWrite} onClick={() => setCreating(true)}>
            + Secret
          </button>
        </div>
      </div>
      {items.length === 0 ? (
        <div className="empty">Aucun secret.</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Nom</th>
              <th>Type</th>
              <th>Valeur</th>
              <th>Utilisé par</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {items.map((s) => (
              <tr key={s.id}>
                <td>
                  <b>{s.name}</b>
                </td>
                <td>
                  <Pill status="a" text={s.type} />
                </td>
                <td className="mono muted">••••••••</td>
                <td className="mono muted">{s.used_by || "—"}</td>
                <td className="rowact">
                  {canWrite && (
                    <button className="btn xs ghost down" onClick={() => del(s)}>
                      ✕
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {creating && (
        <Create
          onClose={() => setCreating(false)}
          onDone={() => {
            setCreating(false);
            load();
          }}
        />
      )}
    </div>
  );
}

function Create({ onClose, onDone }: { onClose: () => void; onDone: () => void }) {
  const toast = useToast();
  const [f, setF] = useState({ name: "psk-dakar", type: "psk", value: "", used_by: "" });
  const [busy, setBusy] = useState(false);
  const set = (k: string, v: string) => setF((s) => ({ ...s, [k]: v }));
  async function submit() {
    setBusy(true);
    try {
      await api.post("/secrets", f);
      toast("Secret enregistré (chiffré)", "ok");
      onDone();
    } catch (e: any) {
      toast(e.message, "err");
      setBusy(false);
    }
  }
  return (
    <Modal
      title="Nouveau secret"
      onClose={onClose}
      footer={
        <>
          <button className="btn ghost" onClick={onClose}>
            Annuler
          </button>
          <button className="btn pri" disabled={busy} onClick={submit}>
            Enregistrer
          </button>
        </>
      }
    >
      <div>
        <label className="flabel">Nom</label>
        <input className="field" value={f.name} onChange={(e) => set("name", e.target.value)} />
      </div>
      <div>
        <label className="flabel">Type</label>
        <select className="field" value={f.type} onChange={(e) => set("type", e.target.value)}>
          <option value="psk">PSK</option>
          <option value="eap">EAP</option>
          <option value="xauth">XAuth</option>
        </select>
      </div>
      <div>
        <label className="flabel">Valeur (chiffrée au repos)</label>
        <input className="field" value={f.value} onChange={(e) => set("value", e.target.value)} />
      </div>
      <div>
        <label className="flabel">Utilisé par</label>
        <input className="field" value={f.used_by} onChange={(e) => set("used_by", e.target.value)} />
      </div>
    </Modal>
  );
}
